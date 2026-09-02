# Upgrade Go dependencies in phases, verifying build + short tests after each phase.
# Run from D:\bccode\backend with Go 1.26 installed:  .\scripts\upgrade-deps.ps1
# Each phase stops on the first failure so the breaking change is easy to locate.
# Phases 1-2 are drop-in (same major version). Phase 3 is opt-in: import-path changes.
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

function Verify($label) {
  Write-Host "`n=== verify: $label ===" -ForegroundColor Cyan
  go build ./...
  if ($LASTEXITCODE) { throw "build failed after $label" }
  go vet ./cmd/... ./pkg/... ./internal/...
  if ($LASTEXITCODE) { throw "vet failed after $label" }
  go test -short -count=1 ./...
  if ($LASTEXITCODE) { throw "tests failed after $label" }
  git add go.mod go.sum
  git commit -q -m "chore(backend): upgrade deps - $label" 2>$null
}

# Phase 1: stdlib-adjacent + low-risk patch/minor bumps
go get `
  golang.org/x/crypto@latest golang.org/x/net@latest golang.org/x/text@latest `
  go.mongodb.org/mongo-driver@latest go.uber.org/zap@latest `
  github.com/stretchr/testify@latest github.com/google/uuid@latest github.com/rs/xid@latest `
  github.com/gorilla/websocket@latest github.com/jellydator/ttlcache/v3@latest `
  github.com/golang-jwt/jwt/v4@latest github.com/lib/pq@latest github.com/samber/lo@latest `
  github.com/segmentio/kafka-go@latest github.com/shopspring/decimal@latest
go mod tidy
Verify "phase 1 (x/*, mongo, zap, small libs)"

# Phase 2: frameworks and drivers, same major version
go get `
  github.com/labstack/echo/v4@latest github.com/swaggo/echo-swagger@latest github.com/swaggo/swag@latest `
  gorm.io/gorm@latest gorm.io/driver/postgres@latest `
  github.com/ClickHouse/clickhouse-go/v2@latest `
  github.com/go-playground/validator/v10@latest github.com/xuri/excelize/v2@latest `
  github.com/aws/aws-sdk-go-v2@latest github.com/aws/aws-sdk-go-v2/config@latest `
  github.com/aws/aws-sdk-go-v2/credentials@latest github.com/aws/aws-sdk-go-v2/service/s3@latest `
  github.com/elastic/go-elasticsearch/v8@latest firebase.google.com/go/v4@latest google.golang.org/api@latest
go mod tidy
Verify "phase 2 (echo, gorm, clickhouse, aws, google)"

# Phase 3 (manual, opt-in): breaking import-path changes. Do these one at a time.
Write-Host @"

Phase 3 - not automated (import paths change; edit code then re-run Verify):
  1. github.com/go-redis/redis/v8  -> github.com/redis/go-redis/v9   (5 files)
  2. github.com/golang-jwt/jwt (v3) -> github.com/golang-jwt/jwt/v5   (1 file; v3 is unmaintained)
  3. github.com/confluentinc/confluent-kafka-go v1.9.2 -> confluent-kafka-go/v2 (4 files, cgo/librdkafka)
     or drop it and keep only segmentio/kafka-go (3 files) to remove the cgo build dependency.
  4. github.com/labstack/echo-contrib v0.14 -> v0.50 (1 file; check middleware API)
"@ -ForegroundColor Yellow
