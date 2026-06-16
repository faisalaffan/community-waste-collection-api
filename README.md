# Community Waste Collection API

REST API for managing community waste collection — households, pickup requests, payments, and reporting. Built as a take-home test for **PT Inosoft Trans Sistem**.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| HTTP Framework | [Fiber v3](https://docs.gofiber.io) |
| ORM | [GORM](https://gorm.io) |
| Database | PostgreSQL 16 |
| Migration | [Atlas](https://atlasgo.io) (HCL) |
| Config | [Viper](https://github.com/spf13/viper) |
| File Storage | MinIO (S3-compatible) |
| Container | Docker + docker compose |

## Architecture

```
cmd/server/          Entry point, DI wiring, graceful shutdown
internal/
  config/            Viper-based env/.env loader
  domain/            GORM models, DTOs, type constants
  handler/           HTTP handlers (Fiber)
  service/           Business logic layer
  repository/        Data access layer (GORM)
  middleware/         Rate limiter
  router/            Route registration
  worker/            Background goroutines
pkg/
  database/          GORM connection helper
  storage/           MinIO/S3 client
  response/          Consistent JSON response helpers
migrations/          Atlas HCL schema definitions
```

### Key Design Decisions

**Manual constructor injection** — no DI framework. Dependencies are wired explicitly in `cmd/server/main.go:run()`. Chosen over Wire/Fx to keep reasoning transparent and avoid code generation.

**Repository → Service → Handler layering** — each layer depends only on the one below via Go interfaces. Services are tested with mock repositories; handlers with mock services. This keeps unit tests fast (no database required) while enabling full integration testing via Docker.

**Atlas over GORM AutoMigrate** — schema is defined declaratively in `migrations/*.hcl` and applied via `atlas schema apply`. Avoids runtime schema drift and gives explicit versioning.

**Interface extraction for testability** — `FileStorage` interface decouples payment confirmation from MinIO. Package-level function variables (`newMinioClient`, `databaseOpener`) let tests inject mocks without changing production signatures.

**Fiber v3** — chosen for performance and idiomatic Go error handling (`return handler(c)` vs middleware chains).

## Quick Start

### Prerequisites

- Go 1.26+
- Docker & docker compose
- [Atlas CLI](https://atlasgo.io/getting-started) (for schema migrations)

### Option 1: Docker (recommended)

```bash
# Clone and enter project
git clone <repo-url> && cd community-waste-collection-api

# Copy env template
cp .env.example .env

# Start all services (app + postgres + minio)
make docker-up

# Apply database migrations
atlas schema apply \
  --config file://atlas.hcl \
  --env local \
  --url "postgres://postgres:postgres@localhost:5432/waste_collection?sslmode=disable"
```

The API is now running at `http://localhost:8080`.

### Option 2: Local Development

```bash
# Start dependencies only
docker compose up -d postgres minio

# Copy and configure env
cp .env.example .env
# Edit .env: set DB_HOST=localhost, S3_ENDPOINT=localhost:9000

# Apply migrations
atlas schema apply --config file://atlas.hcl --env local

# Run with hot reload (or use `make dev`)
go run ./cmd/server/
```

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
| `S3_BUCKET` | `payments` | S3 bucket for payment proofs |
| `S3_USE_SSL` | `false` | Use HTTPS for S3 |

## Database Migrations

Schema is managed by [Atlas](https://atlasgo.io) using HCL definitions.

```bash
# Apply migrations
atlas schema apply --config file://atlas.hcl --env local

# Check for drift
atlas schema diff --config file://atlas.hcl --env local

# (Optional) Generate migration from existing DB
atlas schema inspect --env local > migrations/latest.hcl
```

Migration file: `migrations/20250617000000_init.hcl` — creates `households`, `waste_pickups`, `payments` tables with foreign keys and indexes.

**Note:** GORM does NOT perform auto-migration (`AutoMigrate` is disabled). All schema changes go through Atlas.

## API Reference

Base URL: `http://localhost:8080/api`

### Response Format

Every response follows a consistent envelope:

```json
// Success
{ "status": "success", "data": { ... } }

// List with pagination
{ "status": "success", "data": [...], "pagination": { "page": 1, "per_page": 10, "total": 42, "total_pages": 5 } }

// Error
{ "status": "error", "error": { "code": "NOT_FOUND", "message": "household tidak ditemukan" } }

// Validation
{ "status": "fail", "error": { "code": "VALIDATION_ERROR", "message": "...", "details": [{ "field": "...", "message": "..." }] } }
```

### Households

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/households` | Create household |
| `GET` | `/api/households` | List households (`?page=1&per_page=10`) |
| `GET` | `/api/households/:id` | Get household by ID |
| `DELETE` | `/api/households/:id` | Delete household |

**Create request:**
```json
{ "owner_name": "Budi Santoso", "address": "Jl. Merdeka No. 1, Jakarta" }
```

### Waste Pickups

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/pickups` | Create pickup request (rate-limited: 30/min) |
| `GET` | `/api/pickups` | List pickups (`?status=pending&household_id=<uuid>`) |
| `PUT` | `/api/pickups/:id/schedule` | Schedule pickup |
| `PUT` | `/api/pickups/:id/complete` | Mark as completed |
| `PUT` | `/api/pickups/:id/cancel` | Cancel pickup |

**Create request:**
```json
{ "household_id": "<uuid>", "type": "organic", "safety_check": false }
```

Valid types: `organic`, `plastic`, `paper`, `electronic`

### Payments

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/payments` | Create payment |
| `GET` | `/api/payments` | List payments (`?status=paid&household_id=<uuid>&date_from=2026-01-01&date_to=2026-06-30`) |
| `PUT` | `/api/payments/:id/confirm` | Confirm payment (multipart: `proof_file`) |

### Reports

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/reports/waste-summary` | Pickup aggregation by type + status |
| `GET` | `/api/reports/payment-summary` | Payment totals by status + total revenue |
| `GET` | `/api/reports/households/:id/history` | Full pickup + payment history for a household |

## Business Rules

| Rule | Description |
|------|-------------|
| BR-01 | Household with pending payment cannot create new pickup (409) |
| BR-02 | Only pickups with status `pending` can be scheduled (409) |
| BR-03 | Electronic waste requires `safety_check: true` before scheduling (422) |
| BR-04 | Organic pickups pending >3 days are auto-canceled (background goroutine) |
| BR-05 | Completing a pickup auto-generates a payment (Rp 50.000 for organic/plastic/paper, Rp 100.000 for electronic) |
| BR-06 | Payment confirmation requires proof file upload to S3/MinIO |

## Testing

```bash
# Run all tests
make test

# With coverage
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Vet
make lint
```

**190 tests | 99.4% coverage** — all testable code at 100%. Only `main()` (Go entry point) is excluded.

Coverage breakdown:
- Repositories: 100% (SQLite in-memory)
- Services: 100% (mock repos)
- Handlers: 100% (mock services + Fiber test utilities)
- Middleware: 100% (Fiber test utilities)
- Router: 100%
- Worker: 100% (controlled ticker timing)
- Storage: 100% (mock minio client)
- Database: 100% (SQLite via GORM dialector)

## Makefile

```bash
make build        # Build binary to bin/server
make run          # Build + run
make dev          # Run with go run
make test         # Run all tests
make lint         # go vet
make clean        # Remove binary
make docker-up    # Start all services
make docker-down  # Stop all services
```
