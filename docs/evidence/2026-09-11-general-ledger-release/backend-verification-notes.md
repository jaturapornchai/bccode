# GL release verification — 2026-09-11

Final status: PASS backend, PASS outbox, PASS projection after the approved test-setup repair, PASS GL Linux core + HTTP.

## Scope
Verification plus a subsequently authorized test-setup repair only. No runtime source changes, deployment, production connection, or production data mutation performed by this verification task.

## Verified GL Linux test lane
Command: `docker run ... golang:1.26 bash -c 'go test -json -tags=integration ./internal/generalledger/... -count=1 -timeout=180s'`
Source bind mounted read-only. Dedicated Mongo replica set and PostgreSQL test containers on host loopback 15444 and 15443. Tests create unique temporary Mongo databases and PostgreSQL schemas, then remove their exact test targets.
Result: PASS 17 top-level tests + 38 subtests, 0 skipped. Core 12.247s; HTTP 0.016s.
Evidence: `gl-linux-integration.jsonl`, `gl-linux-integration.exit`.

## Mandatory verification runner
Command: `sh tools/verify.sh backend outbox projection`.
Evidence: `backend-outbox-projection.log`, `backend-outbox-projection.exit` when complete.
The backend target compiles all selected packages, runs short unit tests outside quarantine, and compiles integration-tagged tests. It does not execute all integration tests. Outbox and projection targets execute their specified integration tests against isolated Docker services.
The project intentionally quarantines 15 legacy packages in `backend/.ci/test-quarantine.txt`; those are compiled but their unit/business-contract assertions are not run. A green release check does not certify these contracts.

## Production requirements for deploy owner to verify
- GL connects directly to the existing database named lowercase(holdingcode); it does not create a holding database. Source: `backend/internal/generalledger/httpapi/http.go:46` and `:55`.
- The PostgreSQL connection role must connect to each required holding database and create GL tables/indexes, the audit trigger function and trigger in the active schema. Schema initialization is transactional and serialized. Sources: `backend/internal/generalledger/postgres.go:48`, `backend/internal/generalledger/schema.sql:1`, `:33`, `:41`.
- Connection inputs are POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USERNAME, POSTGRES_PASSWORD, POSTGRES_SSL_MODE. GL uses the holding database, not POSTGRES_DB_NAME. Source: `backend/internal/config/config_postgresql.go:23`, GL resolver above.
- MongoDB must support transactions (replica set or sharded deployment), snapshot read concern and majority write concern. The application account must create collections and GL unique indexes. Sources: `backend/internal/generalledger/store.go:55`, `:149`, `:182`.
- Production environment must resolve its intended Mongo URI/database: BC_ENV/APP_ENV/RUN_ENV/ENVIRONMENT/MODE selects pro; Mongo uses MONGODB_PRO_URI or MONGODB_PRODUCTION_URI and MONGODB_PRO_DB aliases. Source: `backend/internal/config/data_environment.go:14`, `:41`, `:52`.
- Main registers GL HTTP and background projection worker. Worker starts with Microservice.Start and retries pending events; no additional Kafka GL service is needed. Sources: `backend/main.go:391`, `backend/internal/generalledger/httpapi/http.go:69`, `backend/pkg/microservice/microservice.go:233`.
- The existence of production holding databases, actual DDL permissions, production Mongo transaction topology, and backup readiness are not certified by local tests; the deploy owner must confirm them in its authorized production verification.
## Historical runner results before the test-setup repair
`backend`: PASS. `outbox`: PASS (2 top-level + 6 subtests; barcode batch 1 top-level; no selected tests skipped).
`projection`: FAIL twice on unchanged source and the exact mandated runner. Both runs passed 4 top-level tests + 4 subtests, skipped 0, and failed `TestProjectionRebalanceIntegration`.
Failure: `backend/internal/product/projection/consumer_kafka_integration_test.go:83`, Kafka `[6] Not Leader For Partition` during initial `leader.WriteMessages` immediately after topic creation. First failure after 0.08s; retry after 0.07s. The consumer rebalance assertions had not started. Source inspection identifies the failing setup sequence at lines 59–83; this does not establish a failure in GL posting or balances.
These two pre-fix runs were RED. The subsequently authorized setup-only repair and passing projection gate are documented below; the failure history is retained for traceability.
Evidence: `projection-first-failed.jsonl`, `projection-retry.log`, `projection-retry.jsonl`, both runner exit files equal 1, and `test-metrics.json`.
All isolated services created by verify.sh were cleaned up by the runner. The pre-existing dedicated GL test containers remain available for the parent task.
## Final result after authorized test-setup repair
Current local release verification status: PASS backend, PASS outbox, PASS projection, PASS GL Linux core + HTTP.
After the deploy owner authorized a setup-only repair, `consumer_kafka_integration_test.go:70` now bounds initial fixture publication to 10 seconds, refreshes leader metadata with a new DialLeader connection, and retries only explicit NotLeaderForPartition / LeaderNotAvailable responses at 100ms intervals. Transport errors, timeouts, and every other error still fail immediately. No runtime production source changed.
The implementation uses the locked kafka-go v0.4.49 API verified from local module source: DialLeader refreshes partition metadata (`dialer.go:132`) and Conn.WriteMessages submits one atomic partition batch (`conn.go:1107`).
Exact mandated rerun: `sh tools/verify.sh projection` — exit 0, 5 top-level tests + 4 subtests, skip 0. Rebalance passed in 3.90s; both live members, both partitions, failed handler, member departure and committed replay were verified.
`git diff --check` passed. Comparing the source from newReader through the end of TestProjectionRebalanceIntegration against HEAD confirmed the consumer assertions are unchanged.
Evidence: `projection-after-fix.log`, `projection-after-fix.jsonl`, `projection-after-fix.exit`, `projection-test-setup.patch`. The two pre-fix failures remain preserved above and in their original logs.
Quarantine and production provisioning limitations documented above still apply. The parent task reviewed and approved the test-only diff. Production prerequisites still require verification by the deploy owner during deployment.