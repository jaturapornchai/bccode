# BC AI Backend — Copilot Instructions

## Project
Go 1.26 module `smlcloudplatform` — unified backend (mainapi + goapi) using Echo v4 framework with MongoDB, PostgreSQL, ClickHouse, Kafka, Redis.

## Key Paths
- `cmd/goapi/` — goapi entry (port 8888)
- `cmd/app/` — mainapi entry (port 8080)
- `internal/goapi/` — goapi packages
- goapi imports: `smlcloudplatform/internal/goapi/...`

## CRITICAL: Config
- ALL config from `bootstrap.json` ONLY
- Loader: `internal/goapi/setupconfig/loader.go`
- NEVER use `.env`, `godotenv`, or `env_file`
- Docker mount: `bootstrap.json:/app/bootstrap.json:rw`

## Docker Services
MainAPI gateway(:8888, includes /goapi/*), MongoDB(:27017), PostgreSQL(:5432), ClickHouse(:8123/:9000), Redis(:6379), Kafka(:9092)

## Code Style
- Thai OK in logs/comments
- Logger: `logger.Info()`, `logger.Error()`, `logger.Success()`
- Database: `mydb.DatabaseManager` (circuit breaker + retry)
- goapi build: `go build ./cmd/goapi/` (no CGO)
- mainapi build: `CGO_ENABLED=1 go build -tags musl main.go`
