# Changelog

## [1.0.0] — 2026-06-17

### Added
- REST API for households, waste pickups, payments, and reports
- Business rules: pickup scheduling, electronic waste safety check, auto-payment generation, organic pickup auto-cancel
- Swagger UI with OpenAPI annotations
- Atlas declarative schema management (PostgreSQL)
- Docker Compose orchestration (app + PostgreSQL + MinIO)
- Background worker for auto-canceling stale organic pickups
- Rate limiting middleware on pickup creation
- Git-crypt encryption for sensitive files
- VS Code debug configuration
- 190 tests at 99.4% coverage
