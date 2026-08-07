---
kind: build_system
name: Go + Next.js Monorepo Build & Container Pipeline
category: build_system
scope:
    - '**'
source_files:
    - backend/Makefile
    - backend/Dockerfile
    - backend/Dockerfile.goapi
    - backend/Dockerfile-consumer
    - backend/Dockerfile.local
    - frontend/Dockerfile
    - .github/workflows/ci.yml
    - frontend/package.json
---

## What system/approach is used
- **Backend (Go)**: Multi-target Go monorepo built via `make` targets and Dockerfiles. Each microservice binary lives under `backend/cmd/<name>/main.go`; a shared `backend/Dockerfile` builds the main API, while per-service Dockerfiles (`Dockerfile.goapi`, `Dockerfile-consumer`, `Dockerfile-member`, `Dockerfile-migration`) build individual binaries. Builds use multi-stage Alpine images with CGO enabled for librdkafka consumers (`-tags musl`) and CGO disabled for pure Go services.
- **Frontend (Next.js 16)**: Standard npm-based pipeline — `npm ci` → `next build` (standalone output) → `node server.js`. Tests run via Vitest; E2E via Playwright.
- **CI**: GitHub Actions on push/PR to `main`, `master`, `dev`. Backend compile gate runs inside `golang:1.26` container; frontend installs Node 20, lints, typechecks, runs Vitest, then builds.
- **Local orchestration**: `docker-compose.yml` / `docker-compose.dev.yml` / `docker-compose.local.yml` in `backend/` spin up Kafka, Zookeeper, Redis, MongoDB, and backend services. A legacy `Dockerfile.local` mirrors the production Dockerfile without BuildKit cache mounts for hosts lacking `buildx`.
- **Swagger generation**: `swag init` is invoked before every dev/run target; generated docs live under `backend/docs/` and `backend/api/swagger/`.

## Key files and packages
- `backend/Makefile` — all local dev, test, staging, and docker-build/push targets (one per service).
- `backend/Dockerfile` — default multi-stage image for the main API (`go-app`).
- `backend/Dockerfile.goapi` — CGO-disabled static build of `./cmd/goapi/`.
- `backend/Dockerfile-consumer` — CGO-enabled consumer image (older golang:21/alpine3.17 base).
- `backend/Dockerfile.local` — BuildKit-free variant for legacy Docker.
- `frontend/Dockerfile` — Next.js standalone production image (Node 20-alpine).
- `.github/workflows/ci.yml` — backend compile gate + frontend lint/typecheck/test/build.
- `backend/docker-compose*.yml` — compose stacks for dev/staging/local.
- `frontend/package.json` — scripts: `dev`, `build`, `lint`, `typecheck`, `test`, `test:e2e`, `start`.
- `backend/swag` integration: `runswagger`, `docker_build_swagger_and_ship`, `swago-install` targets.

## Architecture and conventions
- **One binary per service**: each entrypoint is a separate `cmd/<svc>/main.go`; Makefile targets mirror this layout (`docker_build_<svc>_and_ship`).
- **Image tagging convention**: `smlsoft/smlcloudplatform:<tag>` where `<tag>` is the service name or role (`authen`, `shop`, `inventory`, `member`, `apidev`, `appdev`, `swagger`).
- **Build flags**:
  - Consumers: `CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags musl` (static librdkafka).
  - Pure Go services: `CGO_ENABLED=0` (see `Dockerfile.goapi`).
  - Local dev on Apple Silicon uses `--tags dynamic` plus `PKG_CONFIG_PATH=/opt/homebrew/opt/openssl@3/lib/pkgconfig`.
- **Environment-driven runtime mode**:
  - `DEV_API_MODE`: `1` = consumer, `2` = API, `3` = migration runner.
  - `MODE`: `staging` | `prd` overrides config loading.
  - `SERVICE_PORT` / `PORT` control HTTP listen port.
- **Health checks**: `/healthz` (API) and `/version` (goapi) endpoints wired into Docker `HEALTHCHECK`.
- **Asset packaging**: `tdict-std.txt`, `assets/fonts/`, `assets/language/`, `assets/address/` are copied into the runtime image alongside the binary.
- **Frontend output**: Next.js `standalone` output is copied into a minimal `node:20-alpine` runner image.

## Rules developers should follow
- **Adding a new Go service**: create `cmd/<name>/main.go`, add a matching `docker_build_<name>` / `docker_build_<name>_and_ship` Makefile target, and a dedicated `Dockerfile.<name>` if it needs different build flags.
- **Always regenerate Swagger** before running or building: `make swag init` (or any `run*` target already invokes it).
- **Use the right Dockerfile**: prefer `backend/Dockerfile` (BuildKit caches) for CI/reproducible builds; fall back to `Dockerfile.local` only when `buildx` is unavailable.
- **Keep Go toolchain aligned**: CI pins `golang:1.26`; ensure local `go.mod` matches.
- **Consumer builds must keep CGO+musl tags**: changing `Dockerfile-consumer` requires verifying librdkafka version compatibility across bases.
- **Frontend changes**: run `npm run lint && npm run typecheck && npm test -- --run` locally; CI will also execute `npm run build`.
- **Compose-only services** (Kafka, Mongo, Redis, etc.) stay in `backend/docker-compose*.yml`; do not bake them into application images.