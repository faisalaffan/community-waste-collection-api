<p align="center">
  <img src="assets/02_BANNERS.png" alt="Community Waste Collection API" width="100%" />
</p>

<p align="center">
  <img src="assets/01_LOGO.png" alt="Logo" width="120" />
</p>

# Community Waste Collection API

REST API for community waste collection — households, pickups, payments, and reports.

## Architecture

```
cmd/server/          Entry point, DI wiring, graceful shutdown
internal/
  config/            Viper .env loader
  domain/            GORM models, DTOs, type constants
  handler/           HTTP handlers (Fiber v3)
  service/           Business logic
  repository/        Data access (GORM)
  middleware/         Rate limiter
  router/            Routes + Swagger UI
  worker/            Background goroutines
pkg/
  database/          GORM connection helper
  storage/           MinIO/S3 client
  response/          JSON response envelope
migrations/          Atlas HCL schema + config
docs/                Swagger spec
assets/              Logo + banner
```

**Stack**: Go 1.26 · Fiber v3 · GORM · PostgreSQL 16 · Atlas · Viper · MinIO · Docker

## Quick Start

Prerequisites: Go 1.26+, Docker, [Atlas CLI](https://atlasgo.io/getting-started).

```bash
git clone https://github.com/faisalaffan/community-waste-collection-api.git
cd community-waste-collection-api
cp .env.example .env
make docker-up
make schema-apply
```

API → `http://localhost:8080` · Swagger → `http://localhost:8080/swagger`

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | HTTP listen port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `waste_collection` | Database name |
| `S3_ENDPOINT` | `localhost:9000` | MinIO/S3 endpoint |
| `S3_ACCESS_KEY` | `minioadmin` | S3 access key |
| `S3_SECRET_KEY` | `minioadmin` | S3 secret key |
| `S3_BUCKET` | `payments` | S3 bucket |
| `S3_USE_SSL` | `false` | Use HTTPS for S3 |

## API Reference

Base: `/api`. Response envelope:

- **Success**: `{"status":"success","data":{}}`
- **List**: `{"status":"success","data":[],"pagination":{"page":1,"per_page":10,"total":42,"total_pages":5}}`
- **Error**: `{"status":"error","error":{"code":"NOT_FOUND","message":"..."}}`
- **Validation**: `{"status":"fail","error":{"code":"VALIDATION_ERROR","message":"...","details":[]}}`

### Households

```
POST   /api/households       Create  {owner_name, address}
GET    /api/households       List    ?page=1&per_page=10
GET    /api/households/:id   Get
DELETE /api/households/:id   Delete
```

### Waste Pickups

```
POST /api/pickups              Create  {household_id, type, safety_check}
GET  /api/pickups              List    ?status=pending&household_id=<uuid>
PUT  /api/pickups/:id/schedule  Schedule
PUT  /api/pickups/:id/complete  Complete
PUT  /api/pickups/:id/cancel    Cancel
```

Types: `organic`, `plastic`, `paper`, `electronic`. Rate limit: 30 req/min.

### Payments

```
POST /api/payments             Create
GET  /api/payments             List    ?status=paid&household_id=<uuid>
PUT  /api/payments/:id/confirm  Confirm (multipart: proof_file)
```

### Reports

```
GET /api/reports/waste-summary              Pickup aggregation
GET /api/reports/payment-summary            Payment totals + revenue
GET /api/reports/households/:id/history    Household history
```

## Business Rules

- Pending payment → blocked from new pickup (409)
- Only `pending` pickups can be scheduled (409)
- Electronic waste requires `safety_check: true` (422)
- Organic pickups pending >3 days auto-canceled
- Completing pickup auto-generates payment (Rp 50k organic/plastic/paper, Rp 100k electronic)
- Payment confirmation requires S3/MinIO proof upload

## Database

Schema managed by [Atlas](https://atlasgo.io) (`migrations/schema.pg.hcl`).

```bash
make schema-apply      # Apply
make schema-diff       # Preview (dry-run)
make schema-inspect    # Pull from DB
```

Tables: `households`, `waste_pickups`, `payments` — UUID PKs, FK with CASCADE, B-tree indexes.

## Testing

190 tests · 99.4% coverage.

```bash
make test              # go test ./... -v
make coverage          # profile + summary
make coverage-html     # browser report
```

## Makefile

```bash
make build             # Build binary
make dev               # go run
make test              # Run tests
make lint              # go vet
make docker-up         # Start all services
make docker-down       # Stop all services
make swagger-clean     # Regenerate swagger
make schema-apply      # Apply Atlas schema
make all               # lint + test + build
```
