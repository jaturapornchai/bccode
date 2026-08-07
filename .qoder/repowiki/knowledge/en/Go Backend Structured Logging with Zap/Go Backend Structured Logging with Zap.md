---
kind: logging_system
name: Go Backend Structured Logging with Zap
category: logging_system
scope:
    - '**'
source_files:
    - backend/internal/logger/logger.go
    - backend/internal/config/config_logger.go
    - backend/pkg/microservice/microservice.go
    - backend/internal/middlewares/request_logger_middleware.go
---

The backend uses a centralized, structured logging system built on `go.uber.org/zap`. It is initialized once per process and exposed through a singleton accessed via `logger.GetLogger()`.

**Framework and initialization**
- Core: `zap.Logger` + `zap.SugaredLogger`, configured in `backend/internal/logger/logger.go`.
- Singleton bootstrap: `NewAppLogger(config.NewLoggerConfig())` called from `pkg/microservice/microservice.go`; the resulting `ILogger` instance is stored on the microservice struct and also retrievable globally via `GetLogger()`.
- Configuration comes from environment variables read by `config.ILoggerConfig`: `LOG_LEVEL` (debug/info/warn/error/dpanic/panic/fatal), `LOG_ENCODER` (`console` or JSON), and `MODE` (development enables dev encoder + file sink).

**Sinks and output format**
- Development (`MODE=development`): dual sink to stdout plus `logs/mainapi_debug.log` (cleared on each start); console encoder with color levels and full caller info.
- Production: single stdout sink; JSON encoder with fields `service`, `timestamp`, `level`, `line`, `message`, `caller`.
- Time encoded as ISO8601; durations as strings; callers use short form in prod, full path in dev.

**Structured fields and conventions**
- HTTP access logs go through `HttpMiddlewareAccessLogger(method, uri, status, size, time)` producing fields `method`, `uri`, `status`, `size`, `time`.
- Kafka consumer logs use `KafkaProcessMessage` / `KafkaProcessMessageWithHeaders` with fields `topic`, `partition`, `MessageSize`, `workerID`, `offset`, `timestamp`, `kafkaHeaders`.
- Error enrichment helpers: `Err(msg, err)` wraps errors via `zap.Error(err)`, `WarnErrMsg(msg, err)` wraps via `zap.String("error", ...)`.
- Constants for field names are declared at package level (`HTTP`, `METHOD`, `URI`, `STATUS`, `SIZE`, `TIME`, `Topic`, `Partition`, `MessageSize`, `WorkerID`, `Offset`, `TimeStamp`, `KafkaHeaders`, `Headers`, `Message`).

**Integration points**
- HTTP middleware: `internal/middlewares/request_logger_middleware.go` calls `HttpMiddlewareAccessLogger` after each request, skipping URIs listed in `IgnoreLogUrls` from HTTP config.
- Microservice bootstrapping: `pkg/microservice/microservice.go` creates the logger, attaches it to the service, and uses it for health-check messages (MongoDB/Kafka/Redis/PostgreSQL connection tests).
- Domain code accesses the global logger via `logger.GetLogger().Info/Error/Warn(...)` throughout services (authentication, creditor/debtor, etc.).

**Developer rules**
- Use `logger.GetLogger()` to obtain the shared logger; do not create local zap instances.
- Prefer `Err(msg, err)` over stringifying errors so the error object is preserved as a structured field.
- For request tracing, rely on the existing middleware; avoid ad-hoc access logging.
- Keep log messages concise — attach context via additional key/value pairs rather than long formatted strings.