# Technical Decisions

Dokumen ini menjelaskan **why** di balik setiap keputusan teknis pada project ini — bukan hanya **what** yang dipakai.

---

## 1. Go 1.26 + Fiber v3

**Keputusan**: Menggunakan Go 1.26 dengan [Fiber v3](https://docs.gofiber.io/) sebagai HTTP framework.

**Mengapa Go?**
- Garbage-collected, statically typed, compile to single binary — ideal untuk deployment
- Concurrency model (goroutine + channel) cocok untuk background worker (organic auto-cancel)
- Rich standard library mengurangi dependency eksternal
- Performa mendekati Rust/C dengan DX lebih sederhana

**Mengapa Fiber v3, bukan Gin/Chi/Echo/net/http?**
- **Fasthttp-based**: Fiber dibangun di atas [fasthttp](https://github.com/valyala/fasthttp), HTTP engine tercepat di Go ecosystem — benchmark menunjukkan 2-10x throughput dibandingkan Gin
- **Express-like API**: Developer experience familiar — method chaining, middleware mounting, grouped routes
- **Zero memory allocation router**: URL parameter parsing tanpa heap allocation
- **v3 dipilih spesifik**: v3 adalah rilis terbaru dengan breaking changes yang membersihkan banyak legacy API dari v2 — forward-looking choice
- **Built-in**: graceful shutdown, request ID, error recovery, CORS, static file serving — tidak butuh middleware eksternal

**Trade-off**: Fiber tidak fully compatible dengan `net/http` handler (berbeda dengan Gin/Chi). Ini trade-off yang diambil untuk performa. Adapter `fasthttpadaptor` tersedia kalau butuh integrasi dengan library `net/http`-based.

---

## 2. Clean Architecture: Handler → Service → Repository

**Keputusan**: Tiga layer terpisah — HTTP handler, business logic service, data access repository.

```
HTTP request → Handler (validasi + binding) → Service (bisnis logic) → Repository (query DB)
                  ↑                               ↑                        ↑
            DTO/response              Domain model + rules          GORM model
```

**Mengapa tiga layer?**
- **Testability**: Setiap layer bisa di-test dengan mock dependency-nya. Handler di-test dengan mock service, service di-test dengan mock repository. Ini alasan utama mengapa project mencapai 243 test dengan 100% coverage.
- **Separation of concerns**: Handler tidak tahu struktur database. Repository tidak tahu aturan bisnis. Service tidak tahu HTTP protocol.
- **Replaceability**: Setiap layer bisa di-swap tanpa mengubah layer lain. Contoh: ganti repository dari GORM ke raw SQL — hanya repository yang berubah.
- **Dependency rule**: Dependency selalu mengarah ke dalam (handler → service → repository). Tidak ada circular reference.

**Alternatif yang dipertimbangkan**:
- **Fat handler (MVC-style)**: Cepat untuk prototype tapi tidak testable dan sulit di-maintain saat aturan bisnis bertambah
- **Domain-Driven Design penuh**: Over-engineered untuk skala project ini (3 entity), menambah boilerplate tanpa manfaat proporsional

---

## 3. GORM sebagai ORM

**Keputusan**: [GORM v2](https://gorm.io/) untuk semua data access.

**Mengapa GORM, bukan sqlx / pgx / raw SQL?**
- **Productivity**: Auto-migration, preloading (eager loading relasi), hook (before/after save) — mengurangi boilerplate signifikan
- **Type safety**: Struct-based query building dengan compile-time check untuk field name
- **Transactions**: GORM menyediakan `db.Transaction()` yang lebih clean dibanding manual `BEGIN/COMMIT/ROLLBACK`
- **Testing dengan SQLite**: GORM mendukung multiple driver — test menggunakan SQLite in-memory (cepat, no external dependency), production menggunakan PostgreSQL. Cukup ganti driver, tidak ada perubahan kode repository.
- **Ekosistem**: Lebih banyak community support, plugin, dan dokumentasi dibanding ORM Go lain (ent, bun)

**Trade-off**:
- GORM tidak se-performant raw SQL untuk query kompleks — tapi untuk use case CRUD+aggregation project ini, overhead tidak signifikan
- Magic behavior kadang membingungkan (soft delete, default scope) — diatasi dengan eksplisit `.Unscoped()` dan `.Select()` saat perlu

---

## 4. PostgreSQL 16 + UUID Primary Key

**Keputusan**: PostgreSQL 16 sebagai database, UUID versi 4 sebagai primary key untuk semua tabel.

**Mengapa PostgreSQL, bukan MySQL/SQLite?**
- **UUID native**: `gen_random_uuid()` built-in tanpa extension — MySQL perlu instalasi plugin
- **JSONB**: Berguna untuk menyimpan data semi-structured di masa depan (log, metadata)
- **Transactional DDL**: Schema migration bisa dalam satu transaksi — rollback aman
- **`text` dan `varchar` tidak ada perbedaan performa**: Tidak perlu debat tipe data string

**Mengapa UUID, bukan auto-increment integer?**
- **Distributed-friendly**: UUID bisa di-generate client-side tanpa collision — penting kalau aplikasi di-scale horizontal
- **Security**: ID tidak dapat ditebak (tidak sequential) — mencegah enumeration attack
- **Frontend independence**: Frontend bisa generate UUID untuk optimistic update tanpa menunggu server
- **Multi-service ready**: Kalau nanti ada service terpisah (notification, reporting), tidak ada konflik primary key

**Trade-off**:
- UUID lebih besar dari BIGINT (16 bytes vs 8 bytes) — tapi untuk skala data project ini (puluhan ribu record), tidak signifikan
- B-tree index pada UUID kolom sedikit lebih lambat — dimitigasi dengan composite index `idx_pickup_household_id`, `idx_pickup_status` yang lebih sering di-query

---

## 5. Atlas Declarative Schema Migration

**Keputusan**: [Atlas](https://atlasgo.io/) dengan declarative HCL untuk schema management.

**Mengapa Atlas, bukan golang-migrate / goose / Flyway?**
- **Declarative, bukan versioned**: Menulis **state akhir yang diinginkan** dalam HCL (`schema.pg.hcl`), bukan urutan DDL incremental. Atlas menghitung diff dan meng-generate SQL migration otomatis.
- **Safety-first**: `--dry-run` menampilkan SQL yang akan dijalankan sebelum eksekusi. Atlas otomatis mendeteksi perubahan destruktif (drop column, drop table) dan minta approval eksplisit.
- **Infrastructure as Code**: File HCL bisa di-review di PR seperti Terraform — reviewer tidak perlu jadi DBA untuk memahami perubahan schema
- **No down migration headache**: Declarative schema tidak butuh file pair `up.sql` + `down.sql` untuk setiap perubahan — Atlas tahu cara rollback dari state sebelumnya

**Konfigurasi**:
```bash
atlas schema apply --env local --to file://migrations/schema.pg.hcl       # Apply
atlas schema apply --env local --to file://migrations/schema.pg.hcl --dry-run  # Preview
atlas schema inspect --env local --format '{{ sql . }}' > schema.sql      # Dump DDL
```

---

## 6. MinIO / S3 untuk File Storage

**Keputusan**: MinIO (S3-compatible) untuk menyimpan bukti pembayaran (file upload).

**Mengapa object storage, bukan filesystem / database blob?**
- **Stateless app**: Aplikasi bisa di-restart / di-scale tanpa kehilangan file — file disimpan di MinIO, bukan di container filesystem
- **S3 API standard**: Kode yang sama bekerja di MinIO (local/dev) dan AWS S3 (production). Ganti endpoint dan credential, tidak ada perubahan kode.
- **Separation of concerns**: Database tidak terbebani menyimpan binary data — PostgreSQL fokus pada structured data
- **CDN-ready**: File di S3 bisa langsung di-serve via CDN tanpa melalui aplikasi Go — `GET /api/files/proof/:paymentID` mem-proxy dari S3, menambahkan akses kontrol

**Mengapa MinIO untuk development?**
- Self-hosted, S3-compatible, tidak butuh koneksi internet
- Satu Docker container — `docker-compose up` langsung siap
- Tidak ada biaya untuk development/testing

---

## 7. Docker Multi-Service

**Keputusan**: Tiga container dalam `docker-compose.yml` — `app` (Go), `postgres` (database), `minio` (object storage).

**Mengapa tiga service terpisah?**
- **Production parity**: Struktur yang sama persis antara development dan production — menghindari "works on my machine"
- **Health check**: `depends_on` dengan `condition: service_healthy` memastikan PostgreSQL siap sebelum aplikasi start — tidak butuh sleep/retry di kode aplikasi
- **Zero-config networking**: Docker Compose otomatis membuat network internal — container bisa saling akses via service name (`postgres:5432`, `minio:9000`)
- **Volume persistence**: `pgdata` dan `miniodata` volumes menyimpan data antar restart — development data tidak hilang saat `docker-compose down`

**Dockerfile (multi-stage build)**:
- Stage 1 (`golang:1.26-alpine`): compile Go binary
- Stage 2 (`alpine:3.21`): copy binary saja — final image hanya ~30MB
- `CGO_ENABLED=0`: pure Go binary, tidak butuh C runtime, aman untuk scratch image

---

## 8. Vue 3 + Tailwind CSS CDN (No Build Step)

**Keputusan**: Admin dashboard dibangun dengan Vue 3 dan Tailwind CSS via CDN — tanpa build tool (Webpack/Vite).

**Mengapa CDN approach, bukan SPA dengan build step?**
- **Zero build complexity**: Tidak ada `package.json`, `node_modules`, atau build pipeline — cukup serve HTML/JS/CSS statis
- **Docker image tetap kecil**: Tidak perlu Node.js di Dockerfile — cukup copy folder `web/`
- **Admin dashboard, bukan consumer app**: SEO tidak relevan, initial load time bukan prioritas — yang penting fungsionalitas operasional
- **Inline dengan Go philosophy**: Satu binary untuk backend + serve frontend statis — tidak ada micro-frontend complexity

**Komponen yang relevan**:
- Vue 3 Options API untuk komponen admin (dashboard, households, pickups, payments, reports)
- Tailwind CDN untuk utility-first styling
- Axios CDN untuk HTTP request ke API backend

---

## 9. Background Worker: Organic Pickup Auto-Cancel

**Keputusan**: Goroutine-based worker dengan ticker 1 jam untuk auto-cancel organic pickup yang pending >3 hari.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
organicWorker := worker.NewOrganicCancelWorker(pickupRepo)
go organicWorker.Start(ctx)
```

**Mengapa goroutine worker, bukan cron job / scheduler eksternal?**
- **Self-contained**: Tidak butuh service tambahan (Celery, RabbitMQ, Redis) — cukup goroutine
- **Graceful shutdown**: Worker dihentikan via `ctx.Done()` saat aplikasi shutdown — tidak ada orphan jobs
- **Ticker pattern**: Lebih sederhana untuk interval tetap (1 jam) dibanding cron expression parsing
- **Testable**: Worker di-test via `Start(ctx)` dengan context cancellable — bisa inject mock repository

**Kenapa tidak scheduler library (gocron/cron)?**
- Project hanya punya 1 background job — tidak butuh framework scheduling penuh
- Menghindari dependency eksternal untuk use case sederhana

---

## 10. Manual Dependency Injection (No Wire/Fx)

**Keputusan**: DI dilakukan manual di `cmd/server/main.go` — tanpa framework injection (Google Wire, Uber Fx).

```go
householdRepo := repository.NewHouseholdRepository(db)
householdSvc := service.NewHouseholdService(householdRepo)
hh := handler.NewHouseholdHandler(householdSvc)
```

**Mengapa manual DI?**
- **Explicit over magic**: Dependency graph terlihat jelas — tidak ada code generation atau reflection yang disembunyikan
- **Compile-time safety**: Tipe mismatch langsung ketahuan saat kompilasi, bukan runtime (berbeda dengan Fx yang injection gagal saat runtime)
- **No code generation**: Wire butuh `wire.go` + `wire_gen.go` — menambah langkah build. Manual DI tidak butuh generate step.
- **Skala project**: Dengan 3 entity dan 4 service, dependency graph masih sederhana — DI framework overkill

**Kapan Wire/Fx lebih cocok?**
- 10+ service dengan dependency graph kompleks
- Multi-modul dengan lifecycle management
- Dynamic injection berdasarkan config/tag

---

## 11. Viper untuk Config Management

**Keputusan**: [Viper](https://github.com/spf13/viper) untuk loading konfigurasi dari `.env` file.

**Mengapa Viper?**
- **Multiple format**: Mendukung `.env`, `.yaml`, `.toml` — fleksibel untuk environment yang berbeda
- **Environment variable override**: `.env` value bisa di-override oleh env var OS — cocok untuk Docker/K8s
- **Default values**: Setiap config key bisa punya fallback default
- **Idiomatic Go**: Standard di Go ecosystem, dokumentasi lengkap

---

## 12. SQLite untuk Testing, PostgreSQL untuk Production

**Keputusan**: Test menggunakan SQLite in-memory, production menggunakan PostgreSQL.

**Mengapa dual driver?**
- **Kecepatan**: SQLite in-memory 10-50x lebih cepat dari PostgreSQL untuk test — 243 test selesai dalam <5 detik
- **Zero setup**: Tidak butuh PostgreSQL berjalan untuk menjalankan test — CI/CD pipeline cukup `go test ./...`
- **Isolasi**: Setiap test case bisa punya database independent — tidak ada data leak antar test
- **GORM abstraction**: Driver swap transparan — repository code sama persis untuk SQLite dan PostgreSQL

---

## 13. JSON Response Envelope

**Keputusan**: Semua API response mengikuti format envelope yang konsisten.

```json
// Success (single)
{"status":"success","data":{}}

// Success (list)
{"status":"success","data":[],"pagination":{"page":1,"per_page":10,"total":42,"total_pages":5}}

// Error
{"status":"error","error":{"code":"NOT_FOUND","message":"..."}}

// Validation error
{"status":"fail","error":{"code":"VALIDATION_ERROR","message":"...","details":[]}}
```

**Mengapa envelope pattern?**
- **Client-side consistency**: Frontend cukup cek `status` untuk menentukan flow (render data / show error / show validation)
- **Extensibility**: Metadata seperti `pagination` atau `debug_info` bisa ditambah tanpa break existing client
- **Error codes**: Machine-readable error codes memungkinkan client menghandle error spesifik (retry, redirect to payment, etc.)
- **Standard**: Pola ini mengikuti [JSend](https://github.com/omniti-labs/jsend) — spec ringan yang dikenal luas

---

## 14. CI/CD dengan GitHub Actions + GHCR

**Keputusan**: GitHub Actions untuk CI, GitHub Container Registry (GHCR) untuk Docker image.

**Mengapa GHCR, bukan Docker Hub?**
- **No rate limiting**: GHCR tidak punya pull rate limit untuk public images (Docker Hub: 100 pulls/6 jam untuk anonymous)
- **Single platform**: Kode, issues, registry, CI/CD — semua di GitHub
- **OIDC-based auth**: GitHub Actions bisa push ke GHCR tanpa menyimpan token — lebih aman

---

## 15. Rate Limiter Middleware

**Keputusan**: In-memory rate limiter 30 req/min per IP.

**Mengapa in-memory, bukan Redis-based?**
- **Simplicity**: Tidak butuh Redis sebagai dependency — cukup counter + timestamp di memory
- **Skala**: Untuk API dengan traffic rendah-menengah, in-memory cukup
- **Stateless backup**: Kalau aplikasi di-scale, ganti ke Redis-based limiter dengan implementasi interface yang sama — perubahan hanya di constructor

---

## 16. Kubernetes Kustomize untuk Production

**Keputusan**: Kustomize (bukan Helm) untuk deployment ke VPS via Kubernetes.

**Mengapa Kustomize?**
- **Built-in di kubectl**: `kubectl apply -k` — tidak butuh instalasi tool tambahan
- **Declarative overlay**: Base + overlay (production) pattern tanpa templating engine
- **secretGenerator**: Generate K8s Secret dari `.env` file — tidak ada secret di commit history
- **Tanpa Go template**: Lebih sederhana dari Helm untuk deployment single-service

---

## Ringkasan Filosofi

| Prinsip | Manifestasi |
|---|---|
| **Explicit over magic** | Manual DI, declarative schema, no code generation |
| **Testability first** | Clean Architecture layering, SQLite test driver, mock-ready interfaces |
| **Minimal dependencies** | Goroutine worker, in-memory rate limiter, Vue CDN |
| **Production-dev parity** | Docker Compose mirror production topology, S3 API compatible |
| **Forward-looking** | UUID PK, Fiber v3, declarative migrations — siap untuk scaling horizontal |
