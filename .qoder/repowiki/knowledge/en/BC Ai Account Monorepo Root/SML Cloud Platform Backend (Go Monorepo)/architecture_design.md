The repository is a single Go module (`smlcloudplatform`) with one `main.go` entry point that bootstraps an Echo HTTP server plus a Kafka consumer pipeline, plus many sibling `cmd/<binary>/main.go` programs (e.g. `authenticationservice`, `masterservice`, `shopservice`, `imageuploadservice`, `migrationapi`, `ws`, `goapi`).

Internal layering per feature package under `internal/<domain>/`:
- `models/` — GORM structs / MongoDB documents.
- `repositories/` — persistence implementations (PostgreSQL via GORM, MongoDB, ClickHouse, OpenSearch, file storage).
- `services/` — business logic; thin HTTP handlers live in `*_http.go` at the package root which wire to services.
- Optional `config/` sub-package for per-feature queue/config wiring.

Cross-cutting infrastructure lives in `pkg/microservice/` (lifecycle, Echo wrapper, Redis cacher, PostgreSQL/MongoDB/ClickHouse/OpenSearch/ELK persister factories, Kafka producer/consumer, WebSocket pool, JWT auth middleware) and `internal/config/` (bootstrap.json-driven config interface). Feature packages depend on these two layers but never on each other directly — they communicate through Kafka topics consumed by `internal/transaction/transactionconsumer/*` and `internal/stockprocess`, `internal/warehouse`, etc.

Build-time mode selection in `main.go` via `DEV_API_MODE`: empty/default runs HTTP + consumers, `3` runs only migrations, `2` runs HTTP-only. A separate `/goapi` group hosts BI/analytics routes implemented in `internal/goapi/`. Dockerfiles at the repo root (`Dockerfile`, `Dockerfile-consumer`, `Dockerfile-member`, `Dockerfile-migration`, `Dockerfile.goapi`) build different slices of the same module into distinct containers.