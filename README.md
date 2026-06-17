# Community Waste Collection API

REST API for managing community waste collection — households, pickup requests, payments, and reporting. Built as a take-home test for **PT Inosoft Trans Sistem**.

- [Repo](#community-waste-collection-api)
- [README](#readme)
- [Environment Variables](#environment-variables)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Business Rules](#business-rules)
- [Database](#database)
- [Testing](#testing)
- [Makefile](#makefile)
- [CHANGELOG](#changelog)
- [Acknowledgment](#acknowledgment)

## README

**`./README.md`**

Intro + architecture + quickstart + API reference + business rules. This file.

## Environment Variables

**`./.env` `./.env.example`**

Loaded by [Viper](https://github.com/spf13/viper) at startup via `config.Load()`. Copy `.env.example` to `.env` and adjust.

- `APP_PORT` — HTTP listen port
  - Default: `8080`
- `DB_HOST` — PostgreSQL host
  - Default: `localhost`
- `DB_PORT` — PostgreSQL port
  - Default: `5432`
- `DB_USER` — Database user
  - Default: `postgres`
- `DB_PASSWORD` — Database password
  - Default: `postgres`
- `DB_NAME` — Database name
  - Default: `waste_collection`
- `S3_ENDPOINT` — MinIO/S3 endpoint
  - Default: `localhost:9000`
- `S3_ACCESS_KEY` — S3 access key
  - Default: `minioadmin`
- `S3_SECRET_KEY` — S3 secret key
  - Default: `minioadmin`
- `S3_BUCKET` — S3 bucket for payment proofs
  - Default: `payments`
- `S3_USE_SSL` — Use HTTPS for S3
  - Allowed: `true`, `false`
  - Default: `false`

DSN built at runtime: `postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable`.

## Architecture

```
cmd/server/          Entry point, DI wiring, graceful shutdown
internal/
  config/            Viper-based .env loader
  domain/            GORM models, DTOs, type constants
  handler/           HTTP handlers (Fiber v3)
  service/           Business logic
  repository/        Data access (GORM)
  middleware/         Rate limiter
  router/            Route registration + Swagger UI
  worker/            Background goroutines
pkg/
  database/          GORM connection helper
  storage/           MinIO/S3 client
  response/          JSON response envelope
migrations/          Atlas HCL schema + config
docs/                Swagger spec + internal docs
```

**Stack**: Go 1.26 · Fiber v3 · GORM · PostgreSQL 16 · Atlas (HCL) · Viper · MinIO · Docker

### Design decisions

- **Manual constructor injection** — no DI framework. Dependencies wired explicitly in `cmd/server/main.go:run()`. Reasoning: transparent, no code generation.
- **Repository → Service → Handler layering** — each layer depends only on the one below via Go interfaces. Services tested with mock repos, handlers with mock services.
- **Atlas schema (declarative)** — schema defined in `migrations/schema.pg.hcl`, applied via `atlas schema apply`. No GORM AutoMigrate.
- **Interface extraction** — `FileStorage` interface decouples payments from MinIO. Package-level function variables (`newMinioClient`, `databaseOpener`) allow test injection.
- **Fiber v3** — performance + idiomatic Go error handling (`return handler(c)`).

## Quick Start

### Prerequisites

- Go 1.26+
- Docker + docker compose
- [Atlas CLI](https://atlasgo.io/getting-started)

### Docker (recommended)

```bash
git clone <repo-url> && cd community-waste-collection-api
cp .env.example .env
make docker-up
make schema-apply          # apply Atlas schema to DB
```

API at `http://localhost:8080`. Swagger at `http://localhost:8080/swagger`.

### Local development

```bash
docker compose up -d postgres minio     # dependencies only
cp .env.example .env                     # edit: DB_HOST=localhost S3_ENDPOINT=localhost:9000
make schema-apply
make dev
```

## Swagger

**`./swagger` `./swagger/doc.json`**

OpenAPI spec served at `/swagger/doc.json` (dynamically generated from `docs.SwaggerInfo`). Swagger UI at `/swagger`. Host follows request `Host` header — works on any port/domain without reconfiguration.

Regenerate after annotation changes:

```bash
make swagger-clean
```

## API Reference

Base path: `/api`. Every response uses a consistent envelope:

- **Success**: `{"status":"success","data":{}}`
- **List**: `{"status":"success","data":[],"pagination":{"page":1,"per_page":10,"total":42,"total_pages":5}}`
- **Error**: `{"status":"error","error":{"code":"NOT_FOUND","message":"..."}}`
- **Validation**: `{"status":"fail","error":{"code":"VALIDATION_ERROR","message":"...","details":[]}}`

### Households

- `POST   /api/households`       — Create (`{"owner_name":"...","address":"..."}`)
- `GET    /api/households`       — List (`?page=1&per_page=10`)
- `GET    /api/households/:id`   — Get by ID
- `DELETE /api/households/:id`   — Delete

### Waste Pickups

- `POST /api/pickups`             — Create (`{"household_id":"<uuid>","type":"organic","safety_check":false}`)
  - Rate-limited: 30 req/min. Types: `organic`, `plastic`, `paper`, `electronic`
- `GET  /api/pickups`             — List (`?status=pending&household_id=<uuid>`)
- `PUT  /api/pickups/:id/schedule` — Schedule
- `PUT  /api/pickups/:id/complete` — Mark completed
- `PUT  /api/pickups/:id/cancel`   — Cancel

### Payments

- `POST /api/payments`            — Create
- `GET  /api/payments`            — List (`?status=paid&household_id=<uuid>&date_from=2026-01-01&date_to=2026-06-30`)
- `PUT  /api/payments/:id/confirm` — Confirm (multipart: `proof_file`)

### Reports

- `GET /api/reports/waste-summary`       — Pickup aggregation by type + status
- `GET /api/reports/payment-summary`     — Payment totals by status + revenue
- `GET /api/reports/households/:id/history` — Full pickup + payment history

## Business Rules

- BR-01 — Household with pending payment blocked from new pickup (409)
- BR-02 — Only `pending` pickups can be scheduled (409)
- BR-03 — Electronic waste requires `safety_check: true` (422)
- BR-04 — Organic pickups pending >3 days auto-canceled (background goroutine)
- BR-05 — Completing pickup auto-generates payment:
  - Organic/Plastic/Paper: Rp 50.000
  - Electronic: Rp 100.000
- BR-06 — Payment confirmation requires proof upload to S3/MinIO

## Database

**`./migrations/`**

Schema managed by [Atlas](https://atlasgo.io) (declarative HCL). File: `migrations/schema.pg.hcl`.

Tables: `households`, `waste_pickups`, `payments` — with UUID primary keys, foreign keys with `CASCADE`, and B-tree indexes on status/type/household_id.

```bash
make schema-apply      # Apply schema to DB
make schema-diff       # Preview changes (dry-run)
make schema-inspect    # Pull current DB schema
make db-url            # Print DSN
```

`migrations/atlas.hcl` — environment config. Reads `DATABASE_URL` from env, uses `docker://postgres/16/dev` for diff calculations.

## Testing

**190 tests · 99.4% coverage** — all testable code at 100%. Only `main()` excluded.

```bash
make test              # go test ./... -v
make test-short        # skip slow tests
make lint              # go vet ./...
make coverage          # coverage profile + summary
make coverage-html     # coverage.html (open in browser)
```

Strategy: repositories on SQLite in-memory, services with mock repos, handlers with mock services, storage with mock MinIO client, worker with controlled ticker timing.

## Makefile

**`./Makefile`**

```bash
make build             # go build -o bin/server ./cmd/server/
make run               # build + run binary
make dev               # go run ./cmd/server/
make test              # go test ./... -v
make lint              # go vet ./...
make clean             # rm -rf bin/
make docker-up         # docker compose up --build
make docker-down       # docker compose down
make swagger           # swag init
make swagger-clean     # regenerate swagger from scratch
make coverage          # coverage profile
make coverage-html     # coverage report in browser
make schema-apply      # atlas schema apply
make schema-diff       # atlas schema apply --dry-run
make schema-inspect    # atlas schema inspect
make all               # lint + test + build
```

## CHANGELOG

See [commit history](https://github.com/faisalaffan/community-waste-collection-api/commits/dev).

## Acknowledgment

Built as technical assessment for **PT Inosoft Trans Sistem**. Thanks to the team for the opportunity.

---

**Tracking** — `community-waste-collection-api` v1.0 — Created 2026-06 — Go 1.26 · Fiber v3 · PostgreSQL 16 · MinIO · Docker — Contact: Muhammad Faisal Affan
