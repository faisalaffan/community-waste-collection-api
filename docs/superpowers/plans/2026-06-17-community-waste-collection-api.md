# Community Waste Collection API — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Build REST API for community waste collection management with Household, WastePickup, Payment CRUD + 6 business rules.

**Architecture:** Standard Go layout (cmd/internal/pkg), manual constructor DI, Fiber router, Gorm ORM, Atlas migrations, PostgreSQL + MinIO via Docker.

**Tech Stack:** Go 1.26.1, Fiber v3, Gorm v1.28, Viper v1.20, Atlas, PostgreSQL 16, MinIO, Docker

---

## Task 1: Initialize Go Module + Dependencies

**Files:**
- Create: `go.mod`

- [ ] **Step 1: Init module and install dependencies**

```bash
go mod init github.com/faisalaffan/community-waste-collection-api

go get github.com/gofiber/fiber/v3
go get github.com/gofiber/fiber/v3/middleware/limiter
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/spf13/viper
go get github.com/google/uuid
go get github.com/minio/minio-go/v7
```

- [ ] **Step 2: Verify go.mod and go.sum exist**

Run: `cat go.mod | head -1`
Expected: `module github.com/faisalaffan/community-waste-collection-api`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: init go module with dependencies"
```

---

## Task 2: Config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing test**

```go
package config

import (
    "os"
    "testing"
)

func TestLoad(t *testing.T) {
    os.Setenv("APP_PORT", "9090")
    os.Setenv("DB_HOST", "db-test")
    os.Setenv("DB_PORT", "5433")
    os.Setenv("DB_USER", "user-test")
    os.Setenv("DB_PASSWORD", "pass-test")
    os.Setenv("DB_NAME", "db-test")
    os.Setenv("S3_ENDPOINT", "s3-test:9000")
    os.Setenv("S3_ACCESS_KEY", "ak-test")
    os.Setenv("S3_SECRET_KEY", "sk-test")
    os.Setenv("S3_BUCKET", "bucket-test")
    os.Setenv("S3_USE_SSL", "true")
    defer os.Clearenv()

    cfg, err := Load()
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if cfg.AppPort != "9090" {
        t.Errorf("AppPort = %s, want 9090", cfg.AppPort)
    }
    if cfg.DBHost != "db-test" {
        t.Errorf("DBHost = %s, want db-test", cfg.DBHost)
    }
    if cfg.DBPort != "5433" {
        t.Errorf("DBPort = %s, want 5433", cfg.DBPort)
    }
    if cfg.DBUser != "user-test" {
        t.Errorf("DBUser = %s, want user-test", cfg.DBUser)
    }
    if cfg.S3Endpoint != "s3-test:9000" {
        t.Errorf("S3Endpoint = %s, want s3-test:9000", cfg.S3Endpoint)
    }
    if cfg.S3AccessKey != "ak-test" {
        t.Errorf("S3AccessKey = %s, want ak-test", cfg.S3AccessKey)
    }
    if !cfg.S3UseSSL {
        t.Errorf("S3UseSSL = false, want true")
    }
    if cfg.S3Bucket != "bucket-test" {
        t.Errorf("S3Bucket = %s, want bucket-test", cfg.S3Bucket)
    }
}

func TestLoadDefaults(t *testing.T) {
    cfg, err := Load()
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if cfg.AppPort != "8080" {
        t.Errorf("default AppPort = %s, want 8080", cfg.AppPort)
    }
    if cfg.DBHost != "localhost" {
        t.Errorf("default DBHost = %s, want localhost", cfg.DBHost)
    }
    if cfg.DBPort != "5432" {
        t.Errorf("default DBPort = %s, want 5432", cfg.DBPort)
    }
    if cfg.S3Endpoint != "localhost:9000" {
        t.Errorf("default S3Endpoint = %s, want localhost:9000", cfg.S3Endpoint)
    }
    if cfg.S3UseSSL {
        t.Errorf("default S3UseSSL = true, want false")
    }
}

func TestDSN(t *testing.T) {
    cfg := &Config{
        DBHost:     "pg",
        DBPort:     "5432",
        DBUser:     "admin",
        DBPassword: "secret",
        DBName:     "waste",
    }
    dsn := cfg.DSN()
    expected := "host=pg port=5432 user=admin password=secret dbname=waste sslmode=disable TimeZone=Asia/Jakarta"
    if dsn != expected {
        t.Errorf("DSN = %s, want %s", dsn, expected)
    }
}
```

- [ ] **Step 2: Run test — verify fail**

Run: `go test ./internal/config/... -v`
Expected: FAIL — `undefined: Load`

- [ ] **Step 3: Implement config.go**

```go
package config

import (
    "fmt"

    "github.com/spf13/viper"
)

type Config struct {
    AppPort     string
    DBHost      string
    DBPort      string
    DBUser      string
    DBPassword  string
    DBName      string
    S3Endpoint  string
    S3AccessKey string
    S3SecretKey string
    S3Bucket    string
    S3UseSSL    bool
}

func Load() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    _ = viper.ReadInConfig()

    viper.SetDefault("APP_PORT", "8080")
    viper.SetDefault("DB_HOST", "localhost")
    viper.SetDefault("DB_PORT", "5432")
    viper.SetDefault("DB_USER", "postgres")
    viper.SetDefault("DB_PASSWORD", "postgres")
    viper.SetDefault("DB_NAME", "waste_collection")
    viper.SetDefault("S3_ENDPOINT", "localhost:9000")
    viper.SetDefault("S3_ACCESS_KEY", "minioadmin")
    viper.SetDefault("S3_SECRET_KEY", "minioadmin")
    viper.SetDefault("S3_BUCKET", "payments")
    viper.SetDefault("S3_USE_SSL", "false")

    return &Config{
        AppPort:     viper.GetString("APP_PORT"),
        DBHost:      viper.GetString("DB_HOST"),
        DBPort:      viper.GetString("DB_PORT"),
        DBUser:      viper.GetString("DB_USER"),
        DBPassword:  viper.GetString("DB_PASSWORD"),
        DBName:      viper.GetString("DB_NAME"),
        S3Endpoint:  viper.GetString("S3_ENDPOINT"),
        S3AccessKey: viper.GetString("S3_ACCESS_KEY"),
        S3SecretKey: viper.GetString("S3_SECRET_KEY"),
        S3Bucket:    viper.GetString("S3_BUCKET"),
        S3UseSSL:    viper.GetString("S3_USE_SSL") == "true",
    }, nil
}

func (c *Config) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
        c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
    )
}
```

- [ ] **Step 4: Run test — verify pass**

Run: `go test ./internal/config/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with viper"
```

---

## Task 3: Response Helper

**Files:**
- Create: `pkg/response/response.go`
- Create: `pkg/response/response_test.go`

- [ ] **Step 1: Write failing test**

```go
package response

import (
    "encoding/json"
    "net/http"
    "testing"

    "github.com/gofiber/fiber/v3"
)

func TestSuccess(t *testing.T) {
    app := fiber.New()
    app.Get("/test", func(c fiber.Ctx) error {
        return Success(c, http.StatusOK, map[string]string{"foo": "bar"})
    })
    resp, _ := app.Test(fiber.NewRequest("GET", "/test"))
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
    var body map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&body)
    if body["status"] != "success" {
        t.Errorf("status = %v, want success", body["status"])
    }
}

func TestError(t *testing.T) {
    app := fiber.New()
    app.Get("/test", func(c fiber.Ctx) error {
        return Error(c, http.StatusNotFound, "NOT_FOUND", "resource not found")
    })
    resp, _ := app.Test(fiber.NewRequest("GET", "/test"))
    if resp.StatusCode != http.StatusNotFound {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
    }
}

func TestValidationError(t *testing.T) {
    app := fiber.New()
    app.Get("/test", func(c fiber.Ctx) error {
        return ValidationError(c, "invalid input", []ValidationDetail{
            {Field: "owner_name", Message: "wajib diisi"},
        })
    })
    resp, _ := app.Test(fiber.NewRequest("GET", "/test"))
    if resp.StatusCode != http.StatusUnprocessableEntity {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
    }
}

func TestSuccessPaginated(t *testing.T) {
    app := fiber.New()
    app.Get("/test", func(c fiber.Ctx) error {
        return SuccessPaginated(c, []string{"a", "b"}, 1, 10, 2)
    })
    resp, _ := app.Test(fiber.NewRequest("GET", "/test"))
    var body map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&body)
    if body["pagination"] == nil {
        t.Error("pagination should not be nil")
    }
}
```

- [ ] **Step 2: Run test — verify fail**

Run: `go test ./pkg/response/... -v`
Expected: FAIL

- [ ] **Step 3: Implement response.go**

```go
package response

import "github.com/gofiber/fiber/v3"

type envelope struct {
    Status     string      `json:"status"`
    Data       interface{} `json:"data,omitempty"`
    Error_     *apiError   `json:"error,omitempty"`
    Pagination *pagination `json:"pagination,omitempty"`
}

type apiError struct {
    Code    string             `json:"code"`
    Message string             `json:"message"`
    Details []ValidationDetail `json:"details,omitempty"`
}

type ValidationDetail struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type pagination struct {
    Page       int   `json:"page"`
    PerPage    int   `json:"per_page"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}

func Success(c fiber.Ctx, status int, data interface{}) error {
    return c.Status(status).JSON(envelope{Status: "success", Data: data})
}

func SuccessCreated(c fiber.Ctx, data interface{}) error {
    return Success(c, 201, data)
}

func SuccessOK(c fiber.Ctx, data interface{}) error {
    return Success(c, 200, data)
}

func SuccessPaginated(c fiber.Ctx, data interface{}, page, perPage int, total int64) error {
    totalPages := int(total) / perPage
    if int(total)%perPage > 0 {
        totalPages++
    }
    if totalPages == 0 {
        totalPages = 1
    }
    return c.Status(200).JSON(envelope{
        Status: "success",
        Data:   data,
        Pagination: &pagination{
            Page:       page,
            PerPage:    perPage,
            Total:      total,
            TotalPages: totalPages,
        },
    })
}

func Error(c fiber.Ctx, status int, code, message string) error {
    return c.Status(status).JSON(envelope{
        Status: "error",
        Error_: &apiError{Code: code, Message: message},
    })
}

func ValidationError(c fiber.Ctx, message string, details []ValidationDetail) error {
    return c.Status(422).JSON(envelope{
        Status: "fail",
        Error_: &apiError{Code: "VALIDATION_ERROR", Message: message, Details: details},
    })
}
```

- [ ] **Step 4: Run test — verify pass**

Run: `go test ./pkg/response/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/response/
git commit -m "feat: add consistent JSON response helpers"
```

---

## Task 4: Domain Models

**Files:**
- Create: `internal/domain/household.go`
- Create: `internal/domain/pickup.go`
- Create: `internal/domain/payment.go`
- Create: `internal/domain/domain_test.go`

- [ ] **Step 1: Create household.go**

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type Household struct {
    ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    OwnerName string        `gorm:"not null" json:"owner_name"`
    Address   string        `gorm:"not null" json:"address"`
    CreatedAt time.Time     `json:"created_at"`
    UpdatedAt time.Time     `json:"updated_at"`
    Pickups   []WastePickup `gorm:"foreignKey:HouseholdID" json:"pickups,omitempty"`
    Payments  []Payment     `gorm:"foreignKey:HouseholdID" json:"payments,omitempty"`
}

type CreateHouseholdRequest struct {
    OwnerName string `json:"owner_name" validate:"required"`
    Address   string `json:"address" validate:"required"`
}
```

- [ ] **Step 2: Create pickup.go**

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

const (
    PickupTypeOrganic    = "organic"
    PickupTypePlastic    = "plastic"
    PickupTypePaper      = "paper"
    PickupTypeElectronic = "electronic"

    PickupStatusPending   = "pending"
    PickupStatusScheduled = "scheduled"
    PickupStatusCompleted = "completed"
    PickupStatusCanceled  = "canceled"
)

type WastePickup struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    HouseholdID uuid.UUID `gorm:"type:uuid;not null;index" json:"household_id"`
    Type        string    `gorm:"type:varchar(20);not null" json:"type"`
    Status      string    `gorm:"type:varchar(20);default:pending" json:"status"`
    PickupDate  *time.Time `json:"pickup_date"`
    SafetyCheck bool      `gorm:"default:false" json:"safety_check"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePickupRequest struct {
    HouseholdID uuid.UUID `json:"household_id" validate:"required"`
    Type        string    `json:"type" validate:"required,oneof=organic plastic paper electronic"`
    SafetyCheck *bool     `json:"safety_check"`
}

type SchedulePickupRequest struct {
    PickupDate time.Time `json:"pickup_date" validate:"required"`
}
```

- [ ] **Step 3: Create payment.go**

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

const (
    PaymentStatusPending = "pending"
    PaymentStatusPaid    = "paid"
    PaymentStatusFailed  = "failed"
)

var PickupAmounts = map[string]float64{
    PickupTypeOrganic:    50000,
    PickupTypePlastic:    50000,
    PickupTypePaper:      50000,
    PickupTypeElectronic: 100000,
}

type Payment struct {
    ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    HouseholdID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"household_id"`
    WasteID      uuid.UUID  `gorm:"type:uuid;not null" json:"waste_id"`
    Amount       float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
    PaymentDate  *time.Time `json:"payment_date"`
    Status       string     `gorm:"type:varchar(20);default:pending" json:"status"`
    ProofFileURL *string    `json:"proof_file_url"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type CreatePaymentRequest struct {
    HouseholdID uuid.UUID `json:"household_id" validate:"required"`
}
```

- [ ] **Step 4: Create domain_test.go**

```go
package domain

import (
    "testing"

    "github.com/google/uuid"
)

func TestNewHousehold(t *testing.T) {
    req := CreateHouseholdRequest{
        OwnerName: "Budi",
        Address:   "Jl. Merdeka No.1",
    }
    if req.OwnerName != "Budi" {
        t.Errorf("OwnerName = %s, want Budi", req.OwnerName)
    }
}

func TestPickupAmounts(t *testing.T) {
    if PickupAmounts[PickupTypeOrganic] != 50000 {
        t.Errorf("organic = %f, want 50000", PickupAmounts[PickupTypeOrganic])
    }
    if PickupAmounts[PickupTypePlastic] != 50000 {
        t.Errorf("plastic = %f, want 50000", PickupAmounts[PickupTypePlastic])
    }
    if PickupAmounts[PickupTypePaper] != 50000 {
        t.Errorf("paper = %f, want 50000", PickupAmounts[PickupTypePaper])
    }
    if PickupAmounts[PickupTypeElectronic] != 100000 {
        t.Errorf("electronic = %f, want 100000", PickupAmounts[PickupTypeElectronic])
    }
}

func TestValidPickupTypes(t *testing.T) {
    valid := map[string]bool{
        "organic":    true,
        "plastic":    true,
        "paper":      true,
        "electronic": true,
        "metal":      false,
        "glass":      false,
    }
    for tpe, expected := range valid {
        actual := tpe == PickupTypeOrganic || tpe == PickupTypePlastic || tpe == PickupTypePaper || tpe == PickupTypeElectronic
        if actual != expected {
            t.Errorf("type %s: valid=%v, want %v", tpe, actual, expected)
        }
    }
}

func TestUUIDGeneration(t *testing.T) {
    id1 := uuid.New()
    id2 := uuid.New()
    if id1 == id2 {
        t.Error("UUIDs should be unique")
    }
}
```

- [ ] **Step 5: Run test — verify pass**

Run: `go test ./internal/domain/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/domain/
git commit -m "feat: add domain models (Household, WastePickup, Payment)"
```

---

## Task 5: Database Connection

**Files:**
- Create: `pkg/database/postgres.go`

- [ ] **Step 1: Implement postgres.go**

```go
package database

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }
    return db, nil
}
```

- [ ] **Step 2: Verify compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add pkg/database/
git commit -m "feat: add postgres connection helper"
```

---

## Task 6: Atlas Migration

**Files:**
- Create: `migrations/20250617000000_init.hcl`
- Create: `atlas.hcl`

- [ ] **Step 1: Create migration HCL**

```hcl
schema "public" {}

table "households" {
  schema = schema.public
  column "id"         { type = uuid; default = sql("gen_random_uuid()") }
  column "owner_name" { type = varchar(255); null = false }
  column "address"    { type = text; null = false }
  column "created_at" { type = timestamptz; default = sql("now()") }
  column "updated_at" { type = timestamptz; default = sql("now()") }
  primary_key { columns = [column.id] }
}

table "waste_pickups" {
  schema = schema.public
  column "id"           { type = uuid; default = sql("gen_random_uuid()") }
  column "household_id" { type = uuid; null = false }
  column "type"         { type = varchar(20); null = false }
  column "status"       { type = varchar(20); default = "pending" }
  column "pickup_date"  { type = timestamptz; null = true }
  column "safety_check" { type = boolean; default = false }
  column "created_at"   { type = timestamptz; default = sql("now()") }
  column "updated_at"   { type = timestamptz; default = sql("now()") }
  primary_key { columns = [column.id] }
  foreign_key "fk_pickup_household" {
    columns     = [column.household_id]
    ref_columns = [table.households.column.id]
    on_delete   = CASCADE
  }
  index "idx_pickup_household_id" { columns = [column.household_id] }
  index "idx_pickup_status"       { columns = [column.status] }
  index "idx_pickup_type"         { columns = [column.type] }
}

table "payments" {
  schema = schema.public
  column "id"             { type = uuid; default = sql("gen_random_uuid()") }
  column "household_id"   { type = uuid; null = false }
  column "waste_id"       { type = uuid; null = false }
  column "amount"         { type = decimal(10,2); null = false }
  column "payment_date"   { type = timestamptz; null = true }
  column "status"         { type = varchar(20); default = "pending" }
  column "proof_file_url" { type = text; null = true }
  column "created_at"     { type = timestamptz; default = sql("now()") }
  column "updated_at"     { type = timestamptz; default = sql("now()") }
  primary_key { columns = [column.id] }
  foreign_key "fk_payment_household" {
    columns     = [column.household_id]
    ref_columns = [table.households.column.id]
    on_delete   = CASCADE
  }
  foreign_key "fk_payment_pickup" {
    columns     = [column.waste_id]
    ref_columns = [table.waste_pickups.column.id]
    on_delete   = CASCADE
  }
  index "idx_payment_household_id" { columns = [column.household_id] }
  index "idx_payment_status"       { columns = [column.status] }
}
```

- [ ] **Step 2: Create atlas.hcl**

```hcl
env "local" {
  src = "migrations/20250617000000_init.hcl"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/16/dev"
}
```

- [ ] **Step 3: Commit**

```bash
git add migrations/ atlas.hcl
git commit -m "feat: add atlas migration for initial schema"
```

---

## Task 7: Repository Layer

**Files:**
- Create: `internal/repository/household.go`
- Create: `internal/repository/household_test.go`
- Create: `internal/repository/pickup.go`
- Create: `internal/repository/pickup_test.go`
- Create: `internal/repository/payment.go`
- Create: `internal/repository/payment_test.go`

- [ ] **Step 1: Create household repository with tests**

`internal/repository/household.go`:
```go
package repository

import (
    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type HouseholdRepository interface {
    Create(h *domain.Household) error
    FindByID(id uuid.UUID) (*domain.Household, error)
    FindAll(page, perPage int) ([]domain.Household, int64, error)
    Delete(id uuid.UUID) error
}

type householdRepo struct {
    db *gorm.DB
}

func NewHouseholdRepository(db *gorm.DB) HouseholdRepository {
    return &householdRepo{db: db}
}

func (r *householdRepo) Create(h *domain.Household) error {
    return r.db.Create(h).Error
}

func (r *householdRepo) FindByID(id uuid.UUID) (*domain.Household, error) {
    var h domain.Household
    err := r.db.First(&h, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &h, nil
}

func (r *householdRepo) FindAll(page, perPage int) ([]domain.Household, int64, error) {
    var households []domain.Household
    var total int64
    r.db.Model(&domain.Household{}).Count(&total)
    offset := (page - 1) * perPage
    err := r.db.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&households).Error
    return households, total, err
}

func (r *householdRepo) Delete(id uuid.UUID) error {
    return r.db.Delete(&domain.Household{}, "id = ?", id).Error
}
```

`internal/repository/household_test.go`:
```go
package repository

import (
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open db: %v", err)
    }
    db.Exec("PRAGMA foreign_keys = ON")
    db.AutoMigrate(&domain.Household{}, &domain.WastePickup{}, &domain.Payment{})
    return db
}

func TestHouseholdRepo_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewHouseholdRepository(db)
    h := &domain.Household{ID: uuid.New(), OwnerName: "Budi", Address: "Jl. A"}

    err := repo.Create(h)
    assert.NoError(t, err)
}

func TestHouseholdRepo_FindByID(t *testing.T) {
    db := setupTestDB(t)
    repo := NewHouseholdRepository(db)
    h := &domain.Household{ID: uuid.New(), OwnerName: "Ani", Address: "Jl. B"}
    repo.Create(h)

    found, err := repo.FindByID(h.ID)
    assert.NoError(t, err)
    assert.Equal(t, "Ani", found.OwnerName)
}

func TestHouseholdRepo_FindAll(t *testing.T) {
    db := setupTestDB(t)
    repo := NewHouseholdRepository(db)
    repo.Create(&domain.Household{ID: uuid.New(), OwnerName: "A", Address: "X"})
    repo.Create(&domain.Household{ID: uuid.New(), OwnerName: "B", Address: "Y"})

    list, total, err := repo.FindAll(1, 10)
    assert.NoError(t, err)
    assert.Equal(t, int64(2), total)
    assert.Len(t, list, 2)
}

func TestHouseholdRepo_Delete(t *testing.T) {
    db := setupTestDB(t)
    repo := NewHouseholdRepository(db)
    h := &domain.Household{ID: uuid.New(), OwnerName: "C", Address: "Z"}
    repo.Create(h)

    err := repo.Delete(h.ID)
    assert.NoError(t, err)

    _, err = repo.FindByID(h.ID)
    assert.Error(t, err)
}
```

- [ ] **Step 2: Run household tests**

Run: `go test ./internal/repository/... -run Household -v`
Expected: PASS

- [ ] **Step 3: Create pickup repository**

`internal/repository/pickup.go`:
```go
package repository

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type PickupRepository interface {
    Create(p *domain.WastePickup) error
    FindByID(id uuid.UUID) (*domain.WastePickup, error)
    FindAll(filter PickupFilter) ([]domain.WastePickup, int64, error)
    Update(p *domain.WastePickup) error
    CancelOrganicPending(olderThan time.Duration) (int64, error)
}

type PickupFilter struct {
    HouseholdID uuid.UUID
    Status      string
    Page        int
    PerPage     int
}

type pickupRepo struct {
    db *gorm.DB
}

func NewPickupRepository(db *gorm.DB) PickupRepository {
    return &pickupRepo{db: db}
}

func (r *pickupRepo) Create(p *domain.WastePickup) error {
    return r.db.Create(p).Error
}

func (r *pickupRepo) FindByID(id uuid.UUID) (*domain.WastePickup, error) {
    var p domain.WastePickup
    err := r.db.First(&p, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &p, nil
}

func (r *pickupRepo) FindAll(filter PickupFilter) ([]domain.WastePickup, int64, error) {
    var pickups []domain.WastePickup
    var total int64
    query := r.db.Model(&domain.WastePickup{})

    if filter.HouseholdID != uuid.Nil {
        query = query.Where("household_id = ?", filter.HouseholdID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }

    query.Count(&total)
    offset := (filter.Page - 1) * filter.PerPage
    err := query.Offset(offset).Limit(filter.PerPage).Order("created_at DESC").Find(&pickups).Error
    return pickups, total, err
}

func (r *pickupRepo) Update(p *domain.WastePickup) error {
    return r.db.Save(p).Error
}

func (r *pickupRepo) CancelOrganicPending(olderThan time.Duration) (int64, error) {
    cutoff := time.Now().Add(-olderThan)
    result := r.db.Model(&domain.WastePickup{}).
        Where("type = ? AND status = ? AND created_at < ?", domain.PickupTypeOrganic, domain.PickupStatusPending, cutoff).
        Update("status", domain.PickupStatusCanceled)
    return result.RowsAffected, result.Error
}
```

`internal/repository/pickup_test.go`:
```go
package repository

import (
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func TestPickupRepo_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewPickupRepository(db)
    p := &domain.WastePickup{ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic}

    err := repo.Create(p)
    assert.NoError(t, err)
    assert.Equal(t, domain.PickupStatusPending, p.Status)
}

func TestPickupRepo_FindByID(t *testing.T) {
    db := setupTestDB(t)
    repo := NewPickupRepository(db)
    p := &domain.WastePickup{ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypePlastic}
    repo.Create(p)

    found, err := repo.FindByID(p.ID)
    assert.NoError(t, err)
    assert.Equal(t, domain.PickupTypePlastic, found.Type)
}

func TestPickupRepo_FindAll_WithFilters(t *testing.T) {
    db := setupTestDB(t)
    repo := NewPickupRepository(db)
    hID := uuid.New()
    repo.Create(&domain.WastePickup{ID: uuid.New(), HouseholdID: hID, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusPending})
    repo.Create(&domain.WastePickup{ID: uuid.New(), HouseholdID: hID, Type: domain.PickupTypePaper, Status: domain.PickupStatusScheduled})

    list, total, err := repo.FindAll(PickupFilter{HouseholdID: hID, Page: 1, PerPage: 10})
    assert.NoError(t, err)
    assert.Equal(t, int64(2), total)
    assert.Len(t, list, 2)
}

func TestPickupRepo_CancelOrganicPending(t *testing.T) {
    db := setupTestDB(t)
    repo := NewPickupRepository(db)
    oldTime := time.Now().Add(-4 * 24 * time.Hour)
    p := &domain.WastePickup{
        ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic,
        Status: domain.PickupStatusPending, CreatedAt: oldTime, UpdatedAt: oldTime,
    }
    db.Create(p)

    affected, err := repo.CancelOrganicPending(3 * 24 * time.Hour)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), affected)
}
```

- [ ] **Step 4: Run pickup tests**

Run: `go test ./internal/repository/... -run Pickup -v`
Expected: PASS

- [ ] **Step 5: Create payment repository**

`internal/repository/payment.go`:
```go
package repository

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type PaymentRepository interface {
    Create(p *domain.Payment) error
    FindByID(id uuid.UUID) (*domain.Payment, error)
    FindAll(filter PaymentFilter) ([]domain.Payment, int64, error)
    Update(p *domain.Payment) error
    HasPendingByHousehold(householdID uuid.UUID) (bool, error)
}

type PaymentFilter struct {
    HouseholdID uuid.UUID
    Status      string
    DateFrom    *time.Time
    DateTo      *time.Time
    Page        int
    PerPage     int
}

type paymentRepo struct {
    db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
    return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(p *domain.Payment) error {
    return r.db.Create(p).Error
}

func (r *paymentRepo) FindByID(id uuid.UUID) (*domain.Payment, error) {
    var p domain.Payment
    err := r.db.First(&p, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &p, nil
}

func (r *paymentRepo) FindAll(filter PaymentFilter) ([]domain.Payment, int64, error) {
    var payments []domain.Payment
    var total int64
    query := r.db.Model(&domain.Payment{})

    if filter.HouseholdID != uuid.Nil {
        query = query.Where("household_id = ?", filter.HouseholdID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.DateFrom != nil {
        query = query.Where("created_at >= ?", filter.DateFrom)
    }
    if filter.DateTo != nil {
        query = query.Where("created_at <= ?", filter.DateTo)
    }

    query.Count(&total)
    offset := (filter.Page - 1) * filter.PerPage
    err := query.Offset(offset).Limit(filter.PerPage).Order("created_at DESC").Find(&payments).Error
    return payments, total, err
}

func (r *paymentRepo) Update(p *domain.Payment) error {
    return r.db.Save(p).Error
}

func (r *paymentRepo) HasPendingByHousehold(householdID uuid.UUID) (bool, error) {
    var count int64
    err := r.db.Model(&domain.Payment{}).
        Where("household_id = ? AND status = ?", householdID, domain.PaymentStatusPending).
        Count(&count).Error
    return count > 0, err
}
```

- [ ] **Step 6: Run payment tests and all repo tests**

Run: `go test ./internal/repository/... -v`
Expected: All PASS

- [ ] **Step 7: Install testify dependency and verify compilation**

Run: `go get github.com/stretchr/testify && go build ./...`
Expected: No errors

- [ ] **Step 8: Commit**

```bash
git add go.mod go.sum internal/repository/
git commit -m "feat: add repository layer with SQLite-backed tests"
```

---

## Task 8: Household Service

**Files:**
- Create: `internal/service/household.go`
- Create: `internal/service/household_test.go`

- [ ] **Step 1: Write failing service test**

```go
package service

import (
    "errors"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

// mock household repo
type mockHouseholdRepo struct {
    createFn   func(h *domain.Household) error
    findByIDFn func(id uuid.UUID) (*domain.Household, error)
    findAllFn  func(page, perPage int) ([]domain.Household, int64, error)
    deleteFn   func(id uuid.UUID) error
}

func (m *mockHouseholdRepo) Create(h *domain.Household) error               { return m.createFn(h) }
func (m *mockHouseholdRepo) FindByID(id uuid.UUID) (*domain.Household, error) { return m.findByIDFn(id) }
func (m *mockHouseholdRepo) FindAll(page, perPage int) ([]domain.Household, int64, error) {
    return m.findAllFn(page, perPage)
}
func (m *mockHouseholdRepo) Delete(id uuid.UUID) error { return m.deleteFn(id) }

func TestHouseholdService_Create(t *testing.T) {
    repo := &mockHouseholdRepo{
        createFn: func(h *domain.Household) error {
            return nil
        },
    }
    svc := NewHouseholdService(repo)

    h, err := svc.Create(&domain.CreateHouseholdRequest{OwnerName: "Budi", Address: "Jl. A"})
    assert.NoError(t, err)
    assert.Equal(t, "Budi", h.OwnerName)
    assert.Equal(t, "Jl. A", h.Address)
    assert.NotEqual(t, uuid.Nil, h.ID)
}

func TestHouseholdService_GetByID_NotFound(t *testing.T) {
    repo := &mockHouseholdRepo{
        findByIDFn: func(id uuid.UUID) (*domain.Household, error) {
            return nil, gorm.ErrRecordNotFound
        },
    }
    svc := NewHouseholdService(repo)
    _, err := svc.GetByID(uuid.New())
    assert.Error(t, err)
    assert.True(t, errors.Is(err, ErrNotFound))
}

func TestHouseholdService_List(t *testing.T) {
    repo := &mockHouseholdRepo{
        findAllFn: func(page, perPage int) ([]domain.Household, int64, error) {
            return []domain.Household{}, int64(0), nil
        },
    }
    svc := NewHouseholdService(repo)
    list, total, err := svc.List(1, 10)
    assert.NoError(t, err)
    assert.Equal(t, int64(0), total)
    assert.Empty(t, list)
}

func TestHouseholdService_Delete(t *testing.T) {
    id := uuid.New()
    repo := &mockHouseholdRepo{
        findByIDFn: func(uid uuid.UUID) (*domain.Household, error) {
            return &domain.Household{ID: uid}, nil
        },
        deleteFn: func(uid uuid.UUID) error {
            return nil
        },
    }
    svc := NewHouseholdService(repo)
    err := svc.Delete(id)
    assert.NoError(t, err)
}
```

- [ ] **Step 2: Run test — verify fail**

Run: `go test ./internal/service/... -run Household -v`
Expected: FAIL

- [ ] **Step 3: Implement household service**

```go
package service

import (
    "errors"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

var (
    ErrNotFound = errors.New("resource not found")
)

type HouseholdService interface {
    Create(req *domain.CreateHouseholdRequest) (*domain.Household, error)
    GetByID(id uuid.UUID) (*domain.Household, error)
    List(page, perPage int) ([]domain.Household, int64, error)
    Delete(id uuid.UUID) error
}

type householdService struct {
    repo repository.HouseholdRepository
}

func NewHouseholdService(repo repository.HouseholdRepository) HouseholdService {
    return &householdService{repo: repo}
}

func (s *householdService) Create(req *domain.CreateHouseholdRequest) (*domain.Household, error) {
    h := &domain.Household{
        ID:        uuid.New(),
        OwnerName: req.OwnerName,
        Address:   req.Address,
    }
    if err := s.repo.Create(h); err != nil {
        return nil, err
    }
    return h, nil
}

func (s *householdService) GetByID(id uuid.UUID) (*domain.Household, error) {
    h, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return h, nil
}

func (s *householdService) List(page, perPage int) ([]domain.Household, int64, error) {
    if page < 1 {
        page = 1
    }
    if perPage < 1 || perPage > 100 {
        perPage = 10
    }
    return s.repo.FindAll(page, perPage)
}

func (s *householdService) Delete(id uuid.UUID) error {
    if _, err := s.repo.FindByID(id); err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrNotFound
        }
        return err
    }
    return s.repo.Delete(id)
}
```

- [ ] **Step 4: Run test — verify pass**

Run: `go test ./internal/service/... -run Household -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/
git commit -m "feat: add household service"
```

---

## Task 9: Pickup Service (BR-01, BR-02, BR-03, BR-05)

**Files:**
- Create: `internal/service/pickup.go`
- Create: `internal/service/pickup_test.go`

- [ ] **Step 1: Create pickup service**

`internal/service/pickup.go`:
```go
package service

import (
    "errors"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

var (
    ErrPendingPayment     = errors.New("household masih memiliki pending payment")
    ErrInvalidStatus      = errors.New("pickup hanya dapat dijadwalkan saat status pending")
    ErrSafetyCheckRequired = errors.New("pickup electronic memerlukan safety_check = true")
    ErrPickupNotPending   = errors.New("hanya pickup dengan status pending yang dapat dibatalkan")
)

type PickupService interface {
    Create(req *domain.CreatePickupRequest) (*domain.WastePickup, error)
    GetByID(id uuid.UUID) (*domain.WastePickup, error)
    List(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
    Schedule(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error)
    Complete(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error)
    Cancel(id uuid.UUID) (*domain.WastePickup, error)
}

type pickupService struct {
    pickupRepo  repository.PickupRepository
    paymentRepo repository.PaymentRepository
}

func NewPickupService(pr repository.PickupRepository, pmr repository.PaymentRepository) PickupService {
    return &pickupService{pickupRepo: pr, paymentRepo: pmr}
}

// BR-01: Blokir pickup jika household punya pending payment
func (s *pickupService) Create(req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
    hasPending, err := s.paymentRepo.HasPendingByHousehold(req.HouseholdID)
    if err != nil {
        return nil, err
    }
    if hasPending {
        return nil, ErrPendingPayment
    }

    safetyCheck := false
    if req.SafetyCheck != nil {
        safetyCheck = *req.SafetyCheck
    }

    p := &domain.WastePickup{
        ID:          uuid.New(),
        HouseholdID: req.HouseholdID,
        Type:        req.Type,
        Status:      domain.PickupStatusPending,
        SafetyCheck: safetyCheck,
    }
    if err := s.pickupRepo.Create(p); err != nil {
        return nil, err
    }
    return p, nil
}

func (s *pickupService) GetByID(id uuid.UUID) (*domain.WastePickup, error) {
    p, err := s.pickupRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return p, nil
}

func (s *pickupService) List(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
    if filter.Page < 1 {
        filter.Page = 1
    }
    if filter.PerPage < 1 || filter.PerPage > 100 {
        filter.PerPage = 10
    }
    return s.pickupRepo.FindAll(filter)
}

// BR-02: Only pending → scheduled
// BR-03: Safety check untuk electronic
func (s *pickupService) Schedule(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
    p, err := s.pickupRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }

    if p.Status != domain.PickupStatusPending {
        return nil, ErrInvalidStatus
    }

    if p.Type == domain.PickupTypeElectronic && !p.SafetyCheck {
        return nil, ErrSafetyCheckRequired
    }

    p.Status = domain.PickupStatusScheduled
    p.PickupDate = &req.PickupDate
    if err := s.pickupRepo.Update(p); err != nil {
        return nil, err
    }
    return p, nil
}

// BR-05: Auto-generate payment saat completed
func (s *pickupService) Complete(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error) {
    p, err := s.pickupRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil, ErrNotFound
        }
        return nil, nil, err
    }

    p.Status = domain.PickupStatusCompleted
    now := time.Now()
    p.PickupDate = &now
    if err := s.pickupRepo.Update(p); err != nil {
        return nil, nil, err
    }

    amount := domain.PickupAmounts[p.Type]
    payment := &domain.Payment{
        ID:          uuid.New(),
        HouseholdID: p.HouseholdID,
        WasteID:     p.ID,
        Amount:      amount,
        Status:      domain.PaymentStatusPending,
    }
    if err := s.paymentRepo.Create(payment); err != nil {
        return nil, nil, err
    }

    return p, payment, nil
}

func (s *pickupService) Cancel(id uuid.UUID) (*domain.WastePickup, error) {
    p, err := s.pickupRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    if p.Status != domain.PickupStatusPending {
        return nil, ErrPickupNotPending
    }
    p.Status = domain.PickupStatusCanceled
    if err := s.pickupRepo.Update(p); err != nil {
        return nil, err
    }
    return p, nil
}
```

- [ ] **Step 2: Create pickup service test**

`internal/service/pickup_test.go`:
```go
package service

import (
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type mockPickupRepo struct {
    createFn               func(p *domain.WastePickup) error
    findByIDFn             func(id uuid.UUID) (*domain.WastePickup, error)
    findAllFn              func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
    updateFn               func(p *domain.WastePickup) error
    cancelOrganicPendingFn func(olderThan time.Duration) (int64, error)
}

func (m *mockPickupRepo) Create(p *domain.WastePickup) error { return m.createFn(p) }
func (m *mockPickupRepo) FindByID(id uuid.UUID) (*domain.WastePickup, error) { return m.findByIDFn(id) }
func (m *mockPickupRepo) FindAll(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
    return m.findAllFn(filter)
}
func (m *mockPickupRepo) Update(p *domain.WastePickup) error { return m.updateFn(p) }
func (m *mockPickupRepo) CancelOrganicPending(olderThan time.Duration) (int64, error) {
    return m.cancelOrganicPendingFn(olderThan)
}

type mockPaymentRepo struct {
    createFn                   func(p *domain.Payment) error
    findByIDFn                 func(id uuid.UUID) (*domain.Payment, error)
    findAllFn                  func(filter repository.PaymentFilter) ([]domain.Payment, int64, error)
    updateFn                   func(p *domain.Payment) error
    hasPendingByHouseholdFn    func(householdID uuid.UUID) (bool, error)
}

func (m *mockPaymentRepo) Create(p *domain.Payment) error { return m.createFn(p) }
func (m *mockPaymentRepo) FindByID(id uuid.UUID) (*domain.Payment, error) { return m.findByIDFn(id) }
func (m *mockPaymentRepo) FindAll(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
    return m.findAllFn(filter)
}
func (m *mockPaymentRepo) Update(p *domain.Payment) error { return m.updateFn(p) }
func (m *mockPaymentRepo) HasPendingByHousehold(householdID uuid.UUID) (bool, error) {
    return m.hasPendingByHouseholdFn(householdID)
}

func TestPickupService_Create_BR01_Blocked(t *testing.T) {
    pmr := &mockPaymentRepo{
        hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) {
            return true, nil
        },
    }
    pr := &mockPickupRepo{}
    svc := NewPickupService(pr, pmr)

    _, err := svc.Create(&domain.CreatePickupRequest{HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic})
    assert.Equal(t, ErrPendingPayment, err)
}

func TestPickupService_Create_Success(t *testing.T) {
    pmr := &mockPaymentRepo{
        hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) {
            return false, nil
        },
    }
    pr := &mockPickupRepo{
        createFn: func(p *domain.WastePickup) error { return nil },
    }
    svc := NewPickupService(pr, pmr)

    p, err := svc.Create(&domain.CreatePickupRequest{HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic})
    assert.NoError(t, err)
    assert.Equal(t, domain.PickupStatusPending, p.Status)
}

func TestPickupService_Schedule_BR02_InvalidStatus(t *testing.T) {
    pr := &mockPickupRepo{
        findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
            return &domain.WastePickup{ID: id, Status: domain.PickupStatusCompleted}, nil
        },
    }
    svc := NewPickupService(pr, nil)

    _, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
    assert.Equal(t, ErrInvalidStatus, err)
}

func TestPickupService_Schedule_BR03_SafetyCheck(t *testing.T) {
    pr := &mockPickupRepo{
        findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
            return &domain.WastePickup{ID: id, Type: domain.PickupTypeElectronic, Status: domain.PickupStatusPending, SafetyCheck: false}, nil
        },
    }
    svc := NewPickupService(pr, nil)

    _, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
    assert.Equal(t, ErrSafetyCheckRequired, err)
}

func TestPickupService_Complete_BR05_GeneratesPayment(t *testing.T) {
    pickupID := uuid.New()
    householdID := uuid.New()
    pickup := &domain.WastePickup{
        ID: pickupID, HouseholdID: householdID, Type: domain.PickupTypeElectronic, Status: domain.PickupStatusPending,
    }

    pr := &mockPickupRepo{
        findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) { return pickup, nil },
        updateFn:   func(p *domain.WastePickup) error { return nil },
    }
    pmr := &mockPaymentRepo{
        createFn: func(p *domain.Payment) error {
            assert.Equal(t, 100000.0, p.Amount)
            return nil
        },
    }
    svc := NewPickupService(pr, pmr)

    _, payment, err := svc.Complete(pickupID)
    assert.NoError(t, err)
    assert.Equal(t, householdID, payment.HouseholdID)
    assert.Equal(t, pickupID, payment.WasteID)
    assert.Equal(t, 100000.0, payment.Amount)
}
```

- [ ] **Step 3: Run tests — verify pass**

Run: `go test ./internal/service/... -v`
Expected: All PASS

- [ ] **Step 4: Commit**

```bash
git add internal/service/
git commit -m "feat: add pickup service with business rules BR-01 to BR-05"
```

---

## Task 10: Payment Service (BR-06)

**Files:**
- Create: `internal/service/payment.go`
- Create: `internal/service/payment_test.go`

- [ ] **Step 1: Create payment service**

`internal/service/payment.go`:
```go
package service

import (
    "fmt"
    "io"
    "mime/multipart"
    "path/filepath"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
    "github.com/faisalaffan/community-waste-collection-api/pkg/storage"
)

var (
    ErrPaymentNotFound = errors.New("payment tidak ditemukan")
    ErrPaymentNotPending = errors.New("hanya payment pending yang dapat dikonfirmasi")
    ErrProofFileRequired = errors.New("file bukti pembayaran wajib diupload")
)

type PaymentService interface {
    Create(req *domain.CreatePaymentRequest) (*domain.Payment, error)
    GetByID(id uuid.UUID) (*domain.Payment, error)
    List(filter repository.PaymentFilter) ([]domain.Payment, int64, error)
    Confirm(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error)
}

type paymentService struct {
    paymentRepo repository.PaymentRepository
    storage     *storage.S3Client
}

func NewPaymentService(pr repository.PaymentRepository, s3 *storage.S3Client) PaymentService {
    return &paymentService{paymentRepo: pr, storage: s3}
}

func (s *paymentService) Create(req *domain.CreatePaymentRequest) (*domain.Payment, error) {
    p := &domain.Payment{
        ID:          uuid.New(),
        HouseholdID: req.HouseholdID,
        Amount:      0,
        Status:      domain.PaymentStatusPending,
    }
    if err := s.paymentRepo.Create(p); err != nil {
        return nil, err
    }
    return p, nil
}

func (s *paymentService) GetByID(id uuid.UUID) (*domain.Payment, error) {
    p, err := s.paymentRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return p, nil
}

func (s *paymentService) List(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
    if filter.Page < 1 {
        filter.Page = 1
    }
    if filter.PerPage < 1 || filter.PerPage > 100 {
        filter.PerPage = 10
    }
    return s.paymentRepo.FindAll(filter)
}

// BR-06: Upload bukti pembayaran ke S3
func (s *paymentService) Confirm(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
    p, err := s.paymentRepo.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrPaymentNotFound
        }
        return nil, err
    }

    if p.Status != domain.PaymentStatusPending {
        return nil, ErrPaymentNotPending
    }

    if file == nil {
        return nil, ErrProofFileRequired
    }

    src, err := file.Open()
    if err != nil {
        return nil, fmt.Errorf("gagal membuka file: %w", err)
    }
    defer src.Close()

    ext := filepath.Ext(file.Filename)
    objectKey := fmt.Sprintf("proofs/%s/%s%s", p.ID.String(), uuid.New().String(), ext)

    url, err := s.storage.Upload(src, objectKey, file.Size, file.Header.Get("Content-Type"))
    if err != nil {
        return nil, fmt.Errorf("gagal upload ke storage: %w", err)
    }

    now := time.Now()
    p.Status = domain.PaymentStatusPaid
    p.PaymentDate = &now
    p.ProofFileURL = &url

    if err := s.paymentRepo.Update(p); err != nil {
        return nil, err
    }
    return p, nil
}
```

Note: Add `"errors"` and `"path/filepath"` to imports. `"mime/multipart"` should stay.

- [ ] **Step 2: Create test**

`internal/service/payment_test.go`:
```go
package service

import (
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

func TestPaymentService_GetByID_NotFound(t *testing.T) {
    pmr := &mockPaymentRepo{
        findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
            return nil, gorm.ErrRecordNotFound
        },
    }
    svc := NewPaymentService(pmr, nil)
    _, err := svc.GetByID(uuid.New())
    assert.Error(t, err)
}

func TestPaymentService_List(t *testing.T) {
    pmr := &mockPaymentRepo{
        findAllFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
            return []domain.Payment{}, int64(0), nil
        },
    }
    svc := NewPaymentService(pmr, nil)
    list, total, err := svc.List(repository.PaymentFilter{Page: 1, PerPage: 10})
    assert.NoError(t, err)
    assert.Equal(t, int64(0), total)
    assert.Len(t, list, 0)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/service/... -run Payment -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/service/
git commit -m "feat: add payment service with BR-06 S3 upload"
```

---

## Task 11: S3 Storage Client

**Files:**
- Create: `pkg/storage/s3.go`

- [ ] **Step 1: Implement S3 client**

```go
package storage

import (
    "context"
    "io"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
    client     *minio.Client
    bucketName string
}

func NewS3(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Client, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, err
    }

    ctx := context.Background()
    exists, err := client.BucketExists(ctx, bucket)
    if err != nil {
        return nil, err
    }
    if !exists {
        err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
        if err != nil {
            return nil, err
        }
    }

    return &S3Client{client: client, bucketName: bucket}, nil
}

func (s *S3Client) Upload(r io.Reader, objectName string, size int64, contentType string) (string, error) {
    ctx := context.Background()
    opts := minio.PutObjectOptions{ContentType: contentType}
    _, err := s.client.PutObject(ctx, s.bucketName, objectName, r, size, opts)
    if err != nil {
        return "", err
    }
    return objectName, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add pkg/storage/
git commit -m "feat: add S3/MinIO storage client"
```

---

## Task 12: Report Service

**Files:**
- Create: `internal/service/report.go`
- Create: `internal/service/report_test.go`

- [ ] **Step 1: Create report service**

`internal/service/report.go`:
```go
package service

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type WasteSummary struct {
    Type   string `json:"type"`
    Status string `json:"status"`
    Count  int64  `json:"count"`
}

type PaymentSummary struct {
    Status        string  `json:"status"`
    Count         int64   `json:"count"`
    TotalAmount   float64 `json:"total_amount"`
}

type HouseholdHistory struct {
    Household domain.Household    `json:"household"`
    Pickups   []domain.WastePickup `json:"pickups"`
    Payments  []domain.Payment     `json:"payments"`
}

type ReportService interface {
    WasteSummary() ([]WasteSummary, error)
    PaymentSummary() ([]PaymentSummary, float64, error)
    HouseholdHistory(householdID uuid.UUID) (*HouseholdHistory, error)
}

type reportService struct {
    db *gorm.DB
}

func NewReportService(db *gorm.DB) ReportService {
    return &reportService{db: db}
}

func (s *reportService) WasteSummary() ([]WasteSummary, error) {
    var results []WasteSummary
    err := s.db.Model(&domain.WastePickup{}).
        Select("type, status, count(*) as count").
        Group("type, status").Order("type, status").Scan(&results).Error
    return results, err
}

func (s *reportService) PaymentSummary() ([]PaymentSummary, float64, error) {
    var results []PaymentSummary
    err := s.db.Model(&domain.Payment{}).
        Select("status, count(*) as count, COALESCE(sum(amount), 0) as total_amount").
        Group("status").Order("status").Scan(&results).Error
    if err != nil {
        return nil, 0, err
    }
    var totalRevenue float64
    s.db.Model(&domain.Payment{}).Where("status = ?", domain.PaymentStatusPaid).
        Select("COALESCE(sum(amount), 0)").Scan(&totalRevenue)
    return results, totalRevenue, nil
}

func (s *reportService) HouseholdHistory(householdID uuid.UUID) (*HouseholdHistory, error) {
    var h domain.Household
    if err := s.db.First(&h, "id = ?", householdID).Error; err != nil {
        return nil, err
    }
    var pickups []domain.WastePickup
    s.db.Where("household_id = ?", householdID).Order("created_at DESC").Find(&pickups)
    var payments []domain.Payment
    s.db.Where("household_id = ?", householdID).Order("created_at DESC").Find(&payments)

    return &HouseholdHistory{
        Household: h,
        Pickups:   pickups,
        Payments:  payments,
    }, nil
}
```

- [ ] **Step 2: Write minimal test**

`internal/service/report_test.go`:
```go
package service

import (
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func TestReportService_HouseholdHistory(t *testing.T) {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&domain.Household{}, &domain.WastePickup{}, &domain.Payment{})

    hID := uuid.New()
    db.Create(&domain.Household{ID: hID, OwnerName: "Test", Address: "Addr", CreatedAt: time.Now(), UpdatedAt: time.Now()})

    svc := NewReportService(db)
    history, err := svc.HouseholdHistory(hID)
    assert.NoError(t, err)
    assert.Equal(t, "Test", history.Household.OwnerName)
    assert.Len(t, history.Pickups, 0)
    assert.Len(t, history.Payments, 0)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/service/... -run Report -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/service/
git commit -m "feat: add report service"
```

---

## Task 13: Household Handler

**Files:**
- Create: `internal/handler/household.go`
- Create: `internal/handler/household_test.go`

- [ ] **Step 1: Create handler**

`internal/handler/household.go`:
```go
package handler

import (
    "errors"
    "strconv"

    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
    "github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type HouseholdHandler struct {
    svc service.HouseholdService
}

func NewHouseholdHandler(svc service.HouseholdService) *HouseholdHandler {
    return &HouseholdHandler{svc: svc}
}

func (h *HouseholdHandler) Create(c fiber.Ctx) error {
    var req domain.CreateHouseholdRequest
    if err := c.Bind().JSON(&req); err != nil {
        return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
            {Field: "body", Message: err.Error()},
        })
    }
    if req.OwnerName == "" {
        return response.ValidationError(c, "owner_name wajib diisi", []response.ValidationDetail{
            {Field: "owner_name", Message: "wajib diisi"},
        })
    }
    if req.Address == "" {
        return response.ValidationError(c, "address wajib diisi", []response.ValidationDetail{
            {Field: "address", Message: "wajib diisi"},
        })
    }

    household, err := h.svc.Create(&req)
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessCreated(c, household)
}

func (h *HouseholdHandler) Get(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    household, err := h.svc.GetByID(id)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            return response.Error(c, 404, "NOT_FOUND", "household tidak ditemukan")
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, household)
}

func (h *HouseholdHandler) List(c fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    perPage, _ := strconv.Atoi(c.Query("per_page", "10"))

    households, total, err := h.svc.List(page, perPage)
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessPaginated(c, households, page, perPage, total)
}

func (h *HouseholdHandler) Delete(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    if err := h.svc.Delete(id); err != nil {
        if errors.Is(err, service.ErrNotFound) {
            return response.Error(c, 404, "NOT_FOUND", "household tidak ditemukan")
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, fiber.Map{"message": "household berhasil dihapus"})
}
```

- [ ] **Step 2: Write handler test**

`internal/handler/household_test.go`:
```go
package handler

import (
    "encoding/json"
    "net/http"
    "strings"
    "testing"

    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
)

type mockHouseholdSvc struct {
    createFn  func(req *domain.CreateHouseholdRequest) (*domain.Household, error)
    getByIDFn func(id uuid.UUID) (*domain.Household, error)
    listFn    func(page, perPage int) ([]domain.Household, int64, error)
    deleteFn  func(id uuid.UUID) error
}

func (m *mockHouseholdSvc) Create(req *domain.CreateHouseholdRequest) (*domain.Household, error) {
    return m.createFn(req)
}
func (m *mockHouseholdSvc) GetByID(id uuid.UUID) (*domain.Household, error) {
    return m.getByIDFn(id)
}
func (m *mockHouseholdSvc) List(page, perPage int) ([]domain.Household, int64, error) {
    return m.listFn(page, perPage)
}
func (m *mockHouseholdSvc) Delete(id uuid.UUID) error { return m.deleteFn(id) }

func TestHouseholdHandler_Create_ValidationError(t *testing.T) {
    app := fiber.New()
    h := NewHouseholdHandler(nil)
    app.Post("/households", h.Create)

    req, _ := http.NewRequest("POST", "/households", strings.NewReader(`{"owner_name":""}`))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    assert.Equal(t, 422, resp.StatusCode)
}

func TestHouseholdHandler_Create_Success(t *testing.T) {
    app := fiber.New()
    svc := &mockHouseholdSvc{
        createFn: func(req *domain.CreateHouseholdRequest) (*domain.Household, error) {
            return &domain.Household{ID: uuid.New(), OwnerName: req.OwnerName, Address: req.Address}, nil
        },
    }
    h := NewHouseholdHandler(svc)
    app.Post("/households", h.Create)

    body := `{"owner_name":"Budi","address":"Jl. A"}`
    req, _ := http.NewRequest("POST", "/households", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    assert.Equal(t, 201, resp.StatusCode)
}

func TestHouseholdHandler_Get_InvalidUUID(t *testing.T) {
    app := fiber.New()
    h := NewHouseholdHandler(nil)
    app.Get("/households/:id", h.Get)

    req, _ := http.NewRequest("GET", "/households/not-a-uuid", nil)
    resp, _ := app.Test(req)
    assert.Equal(t, 400, resp.StatusCode)
}

func TestHouseholdHandler_Get_NotFound(t *testing.T) {
    app := fiber.New()
    svc := &mockHouseholdSvc{
        getByIDFn: func(id uuid.UUID) (*domain.Household, error) {
            return nil, service.ErrNotFound
        },
    }
    h := NewHouseholdHandler(svc)
    app.Get("/households/:id", h.Get)

    req, _ := http.NewRequest("GET", "/households/"+uuid.New().String(), nil)
    resp, _ := app.Test(req)
    assert.Equal(t, 404, resp.StatusCode)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/handler/... -run Household -v`
Expected: All PASS

- [ ] **Step 4: Commit**

```bash
git add internal/handler/
git commit -m "feat: add household handler"
```

---

## Task 14: Pickup Handler

**Files:**
- Create: `internal/handler/pickup.go`
- Create: `internal/handler/pickup_test.go`

- [ ] **Step 1: Create pickup handler**

`internal/handler/pickup.go`:
```go
package handler

import (
    "errors"
    "strconv"

    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
    "github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type PickupHandler struct {
    svc service.PickupService
}

func NewPickupHandler(svc service.PickupService) *PickupHandler {
    return &PickupHandler{svc: svc}
}

func (h *PickupHandler) Create(c fiber.Ctx) error {
    var req domain.CreatePickupRequest
    if err := c.Bind().JSON(&req); err != nil {
        return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
            {Field: "body", Message: err.Error()},
        })
    }

    validTypes := map[string]bool{"organic": true, "plastic": true, "paper": true, "electronic": true}
    if !validTypes[req.Type] {
        return response.ValidationError(c, "type tidak valid", []response.ValidationDetail{
            {Field: "type", Message: "harus salah satu: organic, plastic, paper, electronic"},
        })
    }
    if req.HouseholdID == uuid.Nil {
        return response.ValidationError(c, "household_id wajib diisi", []response.ValidationDetail{
            {Field: "household_id", Message: "wajib diisi"},
        })
    }

    pickup, err := h.svc.Create(&req)
    if err != nil {
        if errors.Is(err, service.ErrPendingPayment) {
            return response.Error(c, 409, "PICKUP_BLOCKED", err.Error())
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessCreated(c, pickup)
}

func (h *PickupHandler) List(c fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
    householdID, _ := uuid.Parse(c.Query("household_id", ""))
    status := c.Query("status", "")

    pickups, total, err := h.svc.List(repository.PickupFilter{
        HouseholdID: householdID,
        Status:      status,
        Page:        page,
        PerPage:     perPage,
    })
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessPaginated(c, pickups, page, perPage, total)
}

func (h *PickupHandler) Schedule(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    var req domain.SchedulePickupRequest
    if err := c.Bind().JSON(&req); err != nil {
        return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
            {Field: "body", Message: err.Error()},
        })
    }

    pickup, err := h.svc.Schedule(id, &req)
    if err != nil {
        switch {
        case errors.Is(err, service.ErrNotFound):
            return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
        case errors.Is(err, service.ErrInvalidStatus):
            return response.Error(c, 409, "INVALID_STATUS", err.Error())
        case errors.Is(err, service.ErrSafetyCheckRequired):
            return response.ValidationError(c, err.Error(), []response.ValidationDetail{
                {Field: "safety_check", Message: "safety_check harus true untuk electronic waste"},
            })
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, pickup)
}

func (h *PickupHandler) Complete(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    pickup, payment, err := h.svc.Complete(id)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, fiber.Map{
        "pickup":  pickup,
        "payment": payment,
    })
}

func (h *PickupHandler) Cancel(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    pickup, err := h.svc.Cancel(id)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
        }
        if errors.Is(err, service.ErrPickupNotPending) {
            return response.Error(c, 409, "INVALID_STATUS", err.Error())
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, pickup)
}
```

- [ ] **Step 2: Write test (BR-01 coverage)**

`internal/handler/pickup_test.go`:
```go
package handler

import (
    "net/http"
    "strings"
    "testing"

    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
)

func TestPickupHandler_Create_ValidationError_EmptyType(t *testing.T) {
    app := fiber.New()
    h := NewPickupHandler(nil)
    app.Post("/pickups", h.Create)

    body := `{"household_id":"` + uuid.New().String() + `","type":""}`
    req, _ := http.NewRequest("POST", "/pickups", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Create_InvalidID(t *testing.T) {
    app := fiber.New()
    h := NewPickupHandler(nil)
    app.Post("/pickups", h.Create)

    body := `{"household_id":"not-uuid","type":"organic"}`
    req, _ := http.NewRequest("POST", "/pickups", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    assert.Equal(t, 422, resp.StatusCode)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/handler/... -run Pickup -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/handler/
git commit -m "feat: add pickup handler"
```

---

## Task 15: Payment Handler

**Files:**
- Create: `internal/handler/payment.go`
- Create: `internal/handler/payment_test.go`

- [ ] **Step 1: Create payment handler**

`internal/handler/payment.go`:
```go
package handler

import (
    "errors"
    "strconv"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
    "github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type PaymentHandler struct {
    svc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
    return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) Create(c fiber.Ctx) error {
    var req domain.CreatePaymentRequest
    if err := c.Bind().JSON(&req); err != nil {
        return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
            {Field: "body", Message: err.Error()},
        })
    }
    if req.HouseholdID == uuid.Nil {
        return response.ValidationError(c, "household_id wajib diisi", []response.ValidationDetail{
            {Field: "household_id", Message: "wajib diisi"},
        })
    }

    payment, err := h.svc.Create(&req)
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessCreated(c, payment)
}

func (h *PaymentHandler) List(c fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
    householdID, _ := uuid.Parse(c.Query("household_id", ""))
    status := c.Query("status", "")

    filter := repository.PaymentFilter{
        HouseholdID: householdID,
        Status:      status,
        Page:        page,
        PerPage:     perPage,
    }

    if dateFrom := c.Query("date_from", ""); dateFrom != "" {
        t, err := time.Parse("2006-01-02", dateFrom)
        if err != nil {
            return response.Error(c, 400, "INVALID_DATE", "format date_from harus YYYY-MM-DD")
        }
        filter.DateFrom = &t
    }
    if dateTo := c.Query("date_to", ""); dateTo != "" {
        t, err := time.Parse("2006-01-02", dateTo)
        if err != nil {
            return response.Error(c, 400, "INVALID_DATE", "format date_to harus YYYY-MM-DD")
        }
        filter.DateTo = &t
    }

    payments, total, err := h.svc.List(filter)
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessPaginated(c, payments, page, perPage, total)
}

func (h *PaymentHandler) Confirm(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    file, err := c.FormFile("proof_file")
    if err != nil {
        return response.ValidationError(c, "file bukti pembayaran wajib diupload", []response.ValidationDetail{
            {Field: "proof_file", Message: "wajib diupload"},
        })
    }

    payment, err := h.svc.Confirm(id, file)
    if err != nil {
        switch {
        case errors.Is(err, service.ErrPaymentNotFound):
            return response.Error(c, 404, "NOT_FOUND", "payment tidak ditemukan")
        case errors.Is(err, service.ErrPaymentNotPending):
            return response.Error(c, 409, "INVALID_STATUS", err.Error())
        case errors.Is(err, service.ErrProofFileRequired):
            return response.ValidationError(c, err.Error(), []response.ValidationDetail{
                {Field: "proof_file", Message: "wajib diupload"},
            })
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, payment)
}
```

- [ ] **Step 2: Write test**

`internal/handler/payment_test.go`:
```go
package handler

import (
    "net/http"
    "strings"
    "testing"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/utils"
    "github.com/google/uuid"

    "github.com/faisalaffan/community-waste-collection-api/internal/domain"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
)

func TestPaymentHandler_Create_ValidationError(t *testing.T) {
    app := fiber.New()
    h := NewPaymentHandler(nil)
    app.Post("/payments", h.Create)

    body := `{"household_id":""}`
    req, _ := http.NewRequest("POST", "/payments", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)
    utils.AssertEqual(t, 422, resp.StatusCode)
}

func TestPaymentHandler_Confirm_InvalidUUID(t *testing.T) {
    app := fiber.New()
    h := NewPaymentHandler(nil)
    app.Put("/payments/:id/confirm", h.Confirm)

    req, _ := http.NewRequest("PUT", "/payments/not-uuid/confirm", nil)
    resp, _ := app.Test(req)
    utils.AssertEqual(t, 400, resp.StatusCode)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/handler/... -run Payment -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/handler/
git commit -m "feat: add payment handler with file upload"
```

---

## Task 16: Report Handler

**Files:**
- Create: `internal/handler/report.go`

- [ ] **Step 1: Create report handler**

```go
package handler

import (
    "github.com/gofiber/fiber/v3"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/faisalaffan/community-waste-collection-api/internal/service"
    "github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type ReportHandler struct {
    svc service.ReportService
}

func NewReportHandler(svc service.ReportService) *ReportHandler {
    return &ReportHandler{svc: svc}
}

func (h *ReportHandler) WasteSummary(c fiber.Ctx) error {
    result, err := h.svc.WasteSummary()
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, result)
}

func (h *ReportHandler) PaymentSummary(c fiber.Ctx) error {
    result, totalRevenue, err := h.svc.PaymentSummary()
    if err != nil {
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, fiber.Map{
        "summary":       result,
        "total_revenue": totalRevenue,
    })
}

func (h *ReportHandler) HouseholdHistory(c fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
    }

    history, err := h.svc.HouseholdHistory(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return response.Error(c, 404, "NOT_FOUND", "household tidak ditemukan")
        }
        return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
    }
    return response.SuccessOK(c, history)
}
```

Add `"errors"` to imports manually since we already use it in other handler files.

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/handler/report.go
git commit -m "feat: add report handler"
```

---

## Task 17: Rate Limit Middleware

**Files:**
- Create: `internal/middleware/ratelimit.go`

- [ ] **Step 1: Create rate limit middleware**

```go
package middleware

import (
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/limiter"
)

func RateLimitPickup() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:               30,
        Expiration:         1 * time.Minute,
        LimiterMiddleware: limiter.SlidingWindow{},
        KeyGenerator: func(c fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c fiber.Ctx) error {
            return c.Status(429).JSON(fiber.Map{
                "status": "error",
                "error": fiber.Map{
                    "code":    "RATE_LIMIT_EXCEEDED",
                    "message": "terlalu banyak request. coba lagi nanti.",
                },
            })
        },
    })
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/middleware/
git commit -m "feat: add rate limiting middleware for pickup creation"
```

---

## Task 18: Router

**Files:**
- Create: `internal/router/router.go`

- [ ] **Step 1: Create router**

```go
package router

import (
    "github.com/gofiber/fiber/v3"

    "github.com/faisalaffan/community-waste-collection-api/internal/handler"
    "github.com/faisalaffan/community-waste-collection-api/internal/middleware"
)

func Setup(
    hh *handler.HouseholdHandler,
    ph *handler.PickupHandler,
    pmh *handler.PaymentHandler,
    rh *handler.ReportHandler,
) *fiber.App {
    app := fiber.New(fiber.Config{
        AppName: "Community Waste Collection API",
    })

    api := app.Group("/api")

    // Households
    households := api.Group("/households")
    households.Post("/", hh.Create)
    households.Get("/", hh.List)
    households.Get("/:id", hh.Get)
    households.Delete("/:id", hh.Delete)

    // Pickups
    pickups := api.Group("/pickups")
    pickups.Post("/", middleware.RateLimitPickup(), ph.Create)
    pickups.Get("/", ph.List)
    pickups.Put("/:id/schedule", ph.Schedule)
    pickups.Put("/:id/complete", ph.Complete)
    pickups.Put("/:id/cancel", ph.Cancel)

    // Payments
    payments := api.Group("/payments")
    payments.Post("/", pmh.Create)
    payments.Get("/", pmh.List)
    payments.Put("/:id/confirm", pmh.Confirm)

    // Reports
    reports := api.Group("/reports")
    reports.Get("/waste-summary", rh.WasteSummary)
    reports.Get("/payment-summary", rh.PaymentSummary)
    reports.Get("/households/:id/history", rh.HouseholdHistory)

    return app
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/router/
git commit -m "feat: add router with all API routes"
```

---

## Task 19: Worker — Organic Auto-Cancel (BR-04)

**Files:**
- Create: `internal/worker/organic_cancel.go`

- [ ] **Step 1: Create worker**

```go
package worker

import (
    "context"
    "log"
    "time"

    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type OrganicCancelWorker struct {
    pickupRepo repository.PickupRepository
    interval   time.Duration
    olderThan  time.Duration
}

func NewOrganicCancelWorker(repo repository.PickupRepository) *OrganicCancelWorker {
    return &OrganicCancelWorker{
        pickupRepo: repo,
        interval:   1 * time.Hour,
        olderThan:  3 * 24 * time.Hour,
    }
}

func (w *OrganicCancelWorker) Start(ctx context.Context) {
    log.Println("[worker] organic cancel worker started")

    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("[worker] organic cancel worker stopped")
            return
        case <-ticker.C:
            affected, err := w.pickupRepo.CancelOrganicPending(w.olderThan)
            if err != nil {
                log.Printf("[worker] error canceling organic pickups: %v", err)
            } else if affected > 0 {
                log.Printf("[worker] canceled %d organic pickups", affected)
            }
        }
    }
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/worker/
git commit -m "feat: add organic auto-cancel worker (BR-04)"
```

---

## Task 20: Main Entry Point + DI Wiring + Graceful Shutdown

**Files:**
- Create: `cmd/server/main.go`

- [ ] **Step 1: Create main.go**

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/faisalaffan/community-waste-collection-api/internal/config"
    "github.com/faisalaffan/community-waste-collection-api/internal/handler"
    "github.com/faisalaffan/community-waste-collection-api/internal/repository"
    "github.com/faisalaffan/community-waste-collection-api/internal/router"
    "github.com/faisalaffan/community-waste-collection-api/internal/service"
    "github.com/faisalaffan/community-waste-collection-api/internal/worker"
    "github.com/faisalaffan/community-waste-collection-api/pkg/database"
    "github.com/faisalaffan/community-waste-collection-api/pkg/storage"
)

func main() {
    // Config
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Database
    db, err := database.NewPostgres(cfg.DSN())
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }

    sqlDB, _ := db.DB()
    defer sqlDB.Close()

    // S3 Storage
    s3Client, err := storage.NewS3(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
    if err != nil {
        log.Fatalf("failed to connect to S3: %v", err)
    }

    // Repositories
    householdRepo := repository.NewHouseholdRepository(db)
    pickupRepo := repository.NewPickupRepository(db)
    paymentRepo := repository.NewPaymentRepository(db)

    // Services
    householdSvc := service.NewHouseholdService(householdRepo)
    pickupSvc := service.NewPickupService(pickupRepo, paymentRepo)
    paymentSvc := service.NewPaymentService(paymentRepo, s3Client)
    reportSvc := service.NewReportService(db)

    // Handlers
    hh := handler.NewHouseholdHandler(householdSvc)
    ph := handler.NewPickupHandler(pickupSvc)
    pmh := handler.NewPaymentHandler(paymentSvc)
    rh := handler.NewReportHandler(reportSvc)

    // Router
    app := router.Setup(hh, ph, pmh, rh)

    // Worker
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    organicWorker := worker.NewOrganicCancelWorker(pickupRepo)
    go organicWorker.Start(ctx)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-quit
        log.Println("shutting down...")
        cancel()
        app.Shutdown()
    }()

    // Start server
    log.Printf("server starting on port %s", cfg.AppPort)
    if err := app.Listen(":" + cfg.AppPort); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
```

- [ ] **Step 2: Verify full build**

Run: `go build ./... && go build -o bin/server ./cmd/server/`
Expected: No errors, binary created

- [ ] **Step 3: Commit**

```bash
git add cmd/ bin/ -f
# Add bin/ to .gitignore after
git commit -m "feat: add main entry point with DI, worker, and graceful shutdown"
```

---

## Task 21: Docker Setup

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.env.example`

- [ ] **Step 1: Create Dockerfile**

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 2: Create docker-compose.yml**

```yaml
services:
  app:
    build: .
    ports:
      - "${APP_PORT:-8080}:8080"
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_started
    env_file:
      - .env
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
      POSTGRES_DB: ${DB_NAME:-waste_collection}
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres} -d ${DB_NAME:-waste_collection}"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${S3_ACCESS_KEY:-minioadmin}
      MINIO_ROOT_PASSWORD: ${S3_SECRET_KEY:-minioadmin}
    ports:
      - "${S3_ENDPOINT_PORT:-9000}:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data
    restart: unless-stopped

volumes:
  pgdata:
  miniodata:
```

- [ ] **Step 3: Create .env.example**

```
APP_PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=waste_collection
S3_ENDPOINT=minio:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=payments
S3_USE_SSL=false
```

- [ ] **Step 4: Commit**

```bash
git add Dockerfile docker-compose.yml .env.example
git commit -m "feat: add Docker setup with PostgreSQL and MinIO"
```

---

## Task 22: Makefile + Final Polish

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create Makefile**

```makefile
.PHONY: build run test dev clean

build:
	go build -o bin/server ./cmd/server/

run: build
	./bin/server

dev:
	go run ./cmd/server/

test:
	go test ./... -v

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

lint:
	go vet ./...
```

- [ ] **Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All PASS

- [ ] **Step 3: Run lint**

Run: `go vet ./...`
Expected: No issues

- [ ] **Step 4: Final build check**

Run: `go build -o /dev/null ./cmd/server/`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add Makefile
git commit -m "chore: add Makefile with common commands"
```
