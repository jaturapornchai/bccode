---
kind: configuration_system
name: JSON Bootstrap + Env Var Overlay Configuration System
category: configuration_system
scope:
    - '**'
source_files:
    - backend/internal/setupconfig/loader.go
    - backend/internal/goapi/setupconfig/loader.go
    - backend/internal/config/config.go
    - backend/bootstrap.json
    - backend/custom_config.json
    - backend/bootstrap.local.json
    - backend/custom_config.local.json
    - backend/main.go
---

The BC Ai Account backend uses a two-stage configuration system that loads JSON files into environment variables, then reads those env vars through typed config accessors. There is no .env file loading — the bootstrap loader explicitly disables dotenv and writes all values into os.Setenv.

How it works:
1. Bootstrap loader (backend/internal/setupconfig/loader.go) runs at process start (called from main() and from internal/goapi/bootstrap.go). It searches for bootstrap.json in this priority: /app/bootstrap/bootstrap.json, /app/bootstrap.json, bootstrap.json, config/bootstrap.json. If found, it also looks alongside it for custom_config.json and deep-merges matching sections on top of the base.
2. The loader maps each JSON section (mongodb, postgresql, clickhouse, service, integrations, storage, kafka) to one or more environment variable names via a configMapping table. For example, service.jwtsecretkey -> JWT_SECRET_KEY, kafka.serverurl -> KAFKA_SERVER_URL. Values are written with os.Setenv, so downstream code only ever sees env vars.
3. Secret masking: keys whose name contains password, secret, or apikey are logged as ***; URIs are truncated after 30 chars.
4. Defaults & composition: if Redis/Kafka env vars are absent, sensible Docker-network defaults are set (redis:6379, kafka:9092). ClickHouse CH_SERVER_ADDRESS is composed from separate host+port fields when needed.
5. Runtime reload: an HTTP endpoint POST /reload-config (protected by X-Reload-Secret header) calls setupconfig.ReloadConfig(), which re-reads bootstrap.json/custom_config.json and overwrites env vars again. This lets the GoAPI setup UI push new DB/integration credentials without restarting the main API.
6. Typed accessors (backend/internal/config/*.go) read from os.Getenv with fallbacks. NewConfig() returns an IConfig interface used throughout the app; per-feature configs (Mongo, PG, ClickHouse, MQ, logger, cacher, HTTP, etc.) live in sibling files under backend/internal/config/.

File layout:
- backend/internal/setupconfig/loader.go - bootstrap loader, mapping, merge, reload
- backend/internal/goapi/setupconfig/loader.go - parallel loader for the embedded GoAPI server (same pattern)
- backend/internal/config/config.go - root IConfig interface and shared helpers
- backend/internal/config/config_*.go - per-subsystem accessor packages (mongo, postgresql, clickhouse, mq, logger, http, cacher, ...)
- backend/bootstrap.json - canonical runtime config (committed with dev values)
- backend/custom_config.json - overlay for secrets/local overrides (gitignored in practice)
- backend/bootstrap.local.json / backend/custom_config.local.json - local-dev variants mounted via Docker volumes

Conventions developers must follow:
- Never import dotenv; add new settings by extending configMapping in loader.go and adding a field to the relevant config_*.go accessor.
- Keep secrets out of bootstrap.json; put them in custom_config.json (or override via actual env vars).
- Use the RELOAD_CONFIG_SECRET env var to gate the /reload-config endpoint in production.
- When adding a new service flag, mirror its env var name across both internal/setupconfig and internal/goapi/setupconfig loaders if both processes need it.
- Default values should be provided in the accessor's getEnv(key, fallback) call so the service can start even if the bootstrap file omits a field.