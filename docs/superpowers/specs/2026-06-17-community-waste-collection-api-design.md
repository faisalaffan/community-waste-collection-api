# Community Waste Collection API — Design Spec

**Date:** 2026-06-17
**Stack:** Go 1.26.1, Fiber, Gorm, Viper, Atlas, PostgreSQL, MinIO, Docker

---

## 1. Project Structure

```
community-waste-collection-api/
├── cmd/
│   └── server/
│       └── main.go              # entry point, DI wiring, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go            # Viper config loading
│   ├── domain/
│   │   ├── household.go         # Gorm model + request/response DTO
│   │   ├── pickup.go
│   │   └── payment.go
│   ├── handler/
│   │   ├── household.go
│   │   ├── pickup.go
│   │   ├── payment.go
│   │   └── report.go
│   ├── service/
│   │   ├── household.go
│   │   ├── pickup.go
│   │   ├── payment.go
│   │   └── report.go
│   ├── repository/
│   │   ├── household.go
│   │   ├── pickup.go
│   │   └── payment.go
│   ├── middleware/
│   │   └── ratelimit.go
│   ├── router/
│   │   └── router.go            # route registration
│   └── worker/
│       └── organic_cancel.go    # BR-04 background goroutine
├── pkg/
│   ├── database/
│   │   └── postgres.go          # Gorm connection
│   ├── storage/
│   │   └── s3.go                # S3-compatible client (MinIO)
│   └── response/
│       └── response.go          # consistent JSON response helper
├── migrations/
│   └── 20250617000000_init.hcl   # Atlas schema definition
├── atlas.hcl                     # Atlas config
├── Dockerfile
├── docker-compose.yml            # app + postgres + minio
├── .env.example
├── go.mod
└── Makefile
```

---

## 2. DI + Data Flow

Constructor injection manual, tanpa library DI.

```
main()
 ├── config.Load()           → *Config
 ├── database.NewPostgres()  → *gorm.DB
 ├── storage.NewS3()         → *S3Client
 │
 ├── repositories            (dep: *gorm.DB)
 ├── services                (dep: repos + *S3Client)
 ├── handlers                (dep: services)
 ├── router.Setup()          (dep: handlers + middlewares) → *fiber.App
 ├── worker.Start()          (dep: pickupRepo) → context.Background
 └── server.Start()          → graceful shutdown (SIGINT/SIGTERM)
```

**Flow:** `HTTP → Fiber Router → Middleware → Handler → Service → Repository → gorm.DB → Postgres` (atau S3Client)

**Graceful shutdown:** SIGINT/SIGTERM → fiber.Shutdown() → worker.Stop() via context.Cancel → gorm db pool close

---

## 3. API Design

### Response Envelope

Success single: `{"status":"success","data":{...}}`
Success list: `{"status":"success","data":[...],"pagination":{"page":1,"per_page":10,"total":42,"total_pages":5}}`
Error: `{"status":"error","error":{"code":"PICKUP_BLOCKED","message":"..."}}`
Validation: `{"status":"fail","error":{"code":"VALIDATION_ERROR","message":"...","details":[{"field":"...","message":"..."}]}}`

### HTTP Status Codes

| Case | Status |
|---|---|
| Create success | 201 |
| Read/List/Update/Delete success | 200 |
| Validation error | 422 |
| Not found | 404 |
| Business rule block | 409 |
| Rate limit | 429 |
| Internal error | 500 |

### Routes

```go
api := app.Group("/api")

// Households
api.Post("/households", hh.Create)
api.Get("/households", hh.List)         // ?page=&per_page=
api.Get("/households/:id", hh.Get)
api.Delete("/households/:id", hh.Delete)

// Pickups
api.Post("/pickups", ph.Create)         // BR-01, rate-limited 30/min/IP
api.Get("/pickups", ph.List)            // ?status=&household_id=
api.Put("/pickups/:id/schedule", ph.Schedule)   // BR-02, BR-03
api.Put("/pickups/:id/complete", ph.Complete)   // BR-05
api.Put("/pickups/:id/cancel", ph.Cancel)

// Payments
api.Post("/payments", pm.Create)
api.Get("/payments", pm.List)           // ?status=&household_id=&date_from=&date_to=
api.Put("/payments/:id/confirm", pm.Confirm)    // BR-06, multipart form

// Reports
api.Get("/reports/waste-summary", rh.WasteSummary)
api.Get("/reports/payment-summary", rh.PaymentSummary)
api.Get("/reports/households/:id/history", rh.HouseholdHistory)
```

### Rate Limiting

Fiber middleware, 30 requests/menit per IP untuk `POST /api/pickups`. Mengembalikan 429 dengan Retry-After header.

---

## 4. Domain Models

### Household
```go
type Household struct {
    ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    OwnerName string        `gorm:"not null"`
    Address   string        `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    Pickups   []WastePickup `gorm:"foreignKey:HouseholdID"`
    Payments  []Payment     `gorm:"foreignKey:HouseholdID"`
}
```

### WastePickup
```go
type WastePickup struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    HouseholdID uuid.UUID  `gorm:"type:uuid;not null;index"`
    Type        string     `gorm:"type:varchar(20);not null"`       // organic|plastic|paper|electronic
    Status      string     `gorm:"type:varchar(20);default:pending"` // pending|scheduled|completed|canceled
    PickupDate  *time.Time
    SafetyCheck bool       `gorm:"default:false"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Payment
```go
type Payment struct {
    ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    HouseholdID  uuid.UUID  `gorm:"type:uuid;not null;index"`
    WasteID      uuid.UUID  `gorm:"type:uuid;not null"`
    Amount       float64    `gorm:"type:decimal(10,2);not null"`
    PaymentDate  *time.Time
    Status       string     `gorm:"type:varchar(20);default:pending"` // pending|paid|failed
    ProofFileURL *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

---

## 5. Business Rules

| Rule | Layer | Logic |
|---|---|---|
| BR-01: Blokir pickup jika pending payment | `pickupService.Create()` | Query payment pending by household_id, return 409 jika ada |
| BR-02: Only pending → scheduled | `pickupService.Schedule()` | Guard: status == "pending" |
| BR-03: Safety check electronic | `pickupService.Schedule()` | Guard: jika type == "electronic" && !safety_check → 422 |
| BR-04: Auto-cancel organic 3 hari | `worker/organic_cancel.go` | Tick tiap 1 jam, cancel organic pending >3 hari. Graceful shutdown via context |
| BR-05: Auto-generate payment | `pickupService.Complete()` | DB transaction: update status + create payment. Amount: organic/plastic/paper=50000, electronic=100000 |
| BR-06: Upload bukti pembayaran | `paymentService.Confirm()` | Multipart form, upload ke MinIO, simpan URL ke proof_file_url |

---

## 6. Atlas Migration

Migration tunggal di `migrations/20250617000000_init.hcl`. Gorm tidak pakai AutoMigrate — schema dikelola sepenuhnya oleh Atlas.

### Configuration (`atlas.hcl`)
```hcl
env "local" {
  src = "migrations/20250617000000_init.hcl"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/16/dev"
}
```

- `atlas schema apply` — apply migration
- `atlas schema diff` — check drift

---

## 7. Docker Setup

`docker-compose.yml`:
- **app**: Go binary, depends on postgres + minio, env dari .env
- **postgres**: PostgreSQL 16, port 5432, volume persist
- **minio**: S3-compatible, ports 9000 (API) + 9001 (console)

---

## 8. Config (Viper)

Variables via `.env` file + env override:

| Key | Default | Usage |
|---|---|---|
| APP_PORT | 8080 | Fiber listen port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | |
| DB_PASSWORD | postgres | |
| DB_NAME | waste_collection | |
| S3_ENDPOINT | localhost:9000 | MinIO endpoint |
| S3_ACCESS_KEY | minioadmin | |
| S3_SECRET_KEY | minioadmin | |
| S3_BUCKET | payments | Bukti pembayaran bucket |
| S3_USE_SSL | false | |
