# Product projection outbox

Updated 2026-09-05. Technical implementation evidence, not a new business contract.

## Objective and workflow

Product Create, Update, Delete and service Resync persist delivery intent in the same MongoDB transaction as the applicable Product/linked Barcode mutation. Company Barcode Create/Update/Delete now use this store too; Create includes automatically created Unit and minimum Product documents in the same transaction. API success means MongoDB committed; PostgreSQL can lag while delivery is pending. A broker failure no longer turns these committed writes into an API delivery error. The existing Unit normalization and name mapping are shared through `unit/services/NewUnitDoc`; the company transaction does not invoke the legacy Unit service write/MQ goroutine.

1. Validate using the existing Product rules.
2. Open a MongoDB transaction, read the aggregate's last version, perform the mutation and insert one `outboxevents` document.
3. A Microservice background worker claims the oldest pending version per Product with a conditional lease.
4. Send each intent through a dedicated producer. Product create/update/delete payloads receive `_projection` metadata at dispatch; Barcode arrays retain their original shape. On complete delivery mark PUBLISHED; otherwise retry with backoff.
5. Nine GoAPI topics (three Product plus six single/bulk Barcode topics) use the shared acknowledged loop in `internal/product/projection`. The legacy Microservice Barcode registrations use that loop too; its selector supports the six Barcode topic names. The existing legacy service registers single create/update/delete and bulk create.
6. Treat incoming payloads as change signals. Acquire PostgreSQL company/row advisory locks, then read current active MongoDB source from the PRIMARY and reconcile metadata. Both GoAPI and the legacy Barcode service force primary reads even when the client prefers a secondary. Source failure returns an error without committing the Kafka offset.
7. Product reconciliation records its per-aggregate version and deletion state in `product_projection_fences` in the same PostgreSQL transaction as the metadata write/delete. Older or duplicate versions cannot overwrite that state. Unversioned signals and delete/recreate with a new GUID refresh current source instead of replaying old snapshots.
8. Product company rebuild acquires an exclusive company lock before its PRIMARY source read. Participating event writers use a shared company lock and exclusive row locks; Barcode batches lock sorted keys. Metadata upserts preserve existing stock/cost columns. The separate legacy DELETE/COPY helper remains atomic and covered by regression tests. Exported GoAPI Barcode insert/update/bulk/delete helpers and `handlers.ProductBarcodeBuild` now forward to canonical reconciliation too; their unused private snapshot writers and empty ClickHouse placeholders were removed.

The worker only handles `aggregatetype=productprojection`; organization outboxes and other contracts are excluded. Aggregate identity is `product:` plus SHA-256 of JSON `[holdingcode,businesscode,guidfixed]`. Company Barcode intents use `barcode:` with the same hash inputs, in a separate version sequence. They can contain Unit-created, Product-created and Barcode messages, in that send order. Only `product:` aggregates receive Product `_projection` metadata; Product messages inside a Barcode intent are unversioned current-source signals, so a Barcode sequence cannot raise the Product fence. An intent may contain both Product and linked Barcode messages. The stored payload is JSON text to preserve the existing JSON representation; this does not convert legacy accounting floats to exact decimals.

The HTTP Resync route still performs its existing synchronous PostgreSQL rebuild first, then queues active Product snapshots through this outbox. Example response data: `{"rebuilt":2,"queued":2,"published":0}`. The legacy `published` field remains, but this request itself performs zero synchronous Kafka sends. Rebuild and all queued snapshots are not one cross-database transaction.

## Configuration and dependencies

- Requires a transaction-capable MongoDB replica set and index-creation permissions. Uses the existing Microservice MongoDB configuration; see `backend/README.md`. Consumer reconciliation now requires primary MongoDB access; a source outage deliberately leaves offsets unacknowledged.
- PostgreSQL requires the existing company-scoped unique keys plus permission to create `product_projection_fences`. Its primary key is `(holding_code,businesscode,itemcode,aggregateuid)`; version is a positive bigint. First use creates the fence table; setup failure is returned and retried. Cache scope is the SQL connection pool.
- Uses existing Kafka URI/security settings from `cfg.MQConfig()`, with a dedicated producer and 30-second message delivery timeout. No new credentials or environment variables.
- Worker runs with the Microservice lifecycle. Shutdown cancels the worker, waits for it, then closes shared resources. A synchronous send may wait for its delivery timeout; the producer's existing close flush may add five seconds. An unfinished intent remains recoverable.
- Poll interval: 1 second. Each pass selects at most 100 eligible aggregate heads. Delivery is sequential. Grouping heads may examine the pending backlog; the output limit does not make the database query constant-cost.
- Lease: 60 seconds, renewed every 20 seconds with a 5-second database timeout. Failure backoff: 1–60 seconds. Attempts count claims.
- Fields: leaseowner/leaseuntil are optional; lasterror stores only a safe category, never the original payload or connection error.
- Indexes: `uniq_product_outbox_version` unique on aggregateuid/version with a productprojection partial filter; `product_outbox_pending` on aggregatetype/status/aggregateuid/version. Creation failure prevents a new mutation; it is retried on subsequent access.

## Read-only operation checks

Select the intended environment's database through the established secure connection. Do not put connection secrets in shell history, docs or logs. These mongosh examples return metadata only:

```javascript
db.outboxevents.aggregate([
  {$match: {aggregatetype: "productprojection"}},
  {$group: {_id: "$status", count: {$sum: 1}, oldest: {$min: "$occurredat"}}}
]);
db.outboxevents.find(
  {aggregatetype: "productprojection", status: "PENDING"},
  {_id: 1, eventuid: 1, aggregateuid: 1, version: 1, attempts: 1,
   occurredat: 1, leaseuntil: 1, lasterror: 1}
).sort({occurredat: 1}).limit(20);
```

If pending age grows: inspect broker connectivity, worker lifecycle, indexes and consumer failures. Fix the cause and let the worker retry. Do not mark events published manually, skip a failed head, reset live Kafka offsets or purge history to make a dashboard green. The current code derives the next version from retained history; safe retention requires a durable version floor first.

## Verification

Run tests only against isolated infrastructure. Tests create unique MongoDB databases, PostgreSQL schemas and Kafka topics, then remove only those targets. `tools/verify.sh` (targets `outbox` and `projection`) runs those isolated suites; there is no GitHub CI any more — it was deleted on 2026-09-09 when GitHub became code storage only. `backend/.ci/projection.compose.yml` runs MongoDB 7, PostgreSQL 18 and Kafka 4.3.1 without host ports or production volumes; the script writes the JSON results to `backend/projection-test-results.json`.

From `backend` on a glibc Go environment with dependencies available:

```sh
SERVERLESS=serverless go test -short -count=1 ./pkg/microservice ./internal/product/product/... ./internal/goapi/mykafkaconsumer ./internal/goapi/handlers/kafka
# Set BC_OUTBOX_TEST_MONGODB_URI to the isolated replica set:
SERVERLESS=serverless go test -tags=integration -count=1 -run '^TestProduct(Outbox|ServiceOutbox)Integration$' ./internal/product/product/outbox ./internal/product/product/services
# Set BC_BARCODE_TEST_POSTGRES_DSN to isolated PostgreSQL:
SERVERLESS=serverless go test -tags=integration -count=1 -run '^TestBarcodeBatchIntegration$' ./internal/goapi/handlers/kafka
```

Alpine `mainapi-builder:latest` requires `-tags musl`, or `-tags musl,integration` for database tests, because of the existing librdkafka static library.

Covered: real MongoDB rollback, duplicate Create rejection, same-document Update, linked Barcode company isolation, soft-delete audit, active-only Resync queueing, retry/order within an aggregate, restart before acknowledgement, competing worker claims, partial Product/Barcode delivery, other outbox exclusion, PostgreSQL DELETE/COPY rollback, replay without extra rows, exact NUMERIC reconciliation, consumer failure/panic/commit failure and real librdkafka timeout to an unavailable loopback port.

Latest local verification passed on MongoDB 7, PostgreSQL 18 and real Kafka 4.3.1. Actual Confluent publishing and the kafka-go acknowledged loop were exercised with delete consumed before create/update, duplicate delivery, a new GUID reusing the same code, stale legacy Barcode snapshots, tenant isolation, and PostgreSQL rejection followed by reopening the same consumer group and successful replay. A failed write left both the version fence and offset unchanged. The legacy GORM writer passed stock/cost preservation, rollback and shared-lock coordination tests. A separate real MongoDB test configured SECONDARY-only reads on a single-primary replica set and proved the consumer override still reads PRIMARY.

The real Kafka transport also passed two live consumer members/two partitions, member departure and replay of a failed uncommitted offset, with all 12 distinct offsets eventually committed. Company Barcode CRUD tests query MongoDB after each step, preserve the holding-wide stock uniqueness guard and another company, retain intent after broker failure, and prove outbox rejection rolls back Unit, Product and Barcode together.

These are service/integration tests, not browser UAT or a full deployed multi-process legacy-service test. Multi-broker failure, ClickHouse and actual restore remain unverified.

From the repo root, a fresh disposable Compose project can run the entire isolated suite:

```sh
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml up -d mongo mongo-init postgres kafka
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml up -d --wait mongo postgres kafka
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml run --rm --no-deps tests
# Verify that bc-projection-check identifies only this disposable stack first:
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml down -v --remove-orphans
```

The default test image is glibc `golang:1.26`. For an installed Alpine builder set `BC_PROJECTION_GO_IMAGE=mainapi-builder:latest` and `BC_PROJECTION_GO_TAGS=musl,integration`. Replica-set init is a one-shot service; only the long-running services belong in the `--wait` command.

## Limits and release/rollback

- At-least-once delivery. A crash after broker acknowledgement but before MongoDB acknowledgement replays the intent. If its second message fails, the first can replay too.
- Product fences now span create/update/delete topics for the same GUID aggregate. Version metadata is a decimal JSON string (including values above 2^53), not an accounting-number migration. The fence is a watermark for a current-state projection, not a global ledger ordering mechanism. Versions from different GUID incarnations are not compared; current source resolves code reuse.
- This convergence guarantee requires all competing writers to use the lock/source-read protocol. GoAPI Product/Barcode event handlers, legacy Barcode UpSert/Delete and Product company rebuild participate. Other manual Barcode/admin/stock rebuilds have not all been converted; do not run them concurrently and assume the same guarantee.
- Legacy Barcode writers without businesscode, standalone Unit writes and raw import/batch paths still have DB-before-MQ gaps. Company Barcode CRUD and its automatically created Unit are covered now. Reconciliation prevents stale snapshots from winning but cannot recreate a signal that was never published. Other legacy Microservice topics retain their old auto-commit/error behavior pending separate review. No ledger/event idempotency retrofit was performed.
- Product metadata upsert preserves stock columns; ClickHouse behavior was not validated by these tests.
- A structurally unusable message (unparsable JSON, missing holding/business/barcode identity, invalid fence identity) is logged by topic/partition/offset and acknowledged so it cannot block the partition; this restores the pre-outbox behaviour for holding-level bulk Barcode payloads. Infrastructure or source failures still hold the offset: that head remains pending until the cause is fixed. No dead-letter topic or production alert routing was added.
- Historical events already lost before this change are not recreated automatically. Organization fallback delivery and unrelated topics are not covered.
- Review pending age and database/consumer metrics during rollout. Broker PUBLISHED proves broker delivery, not downstream database completion.
- Roll out compatible consumers and source-read/locking paths together. The legacy Barcode readers use `<CONSUMER_GROUP_ID>-projection` (kafka-go cannot share a group with the librdkafka members that still serve the other legacy topics); the first start replays the Barcode topics from the earliest offset through idempotent current-source reconciliation. GoAPI reconciles Barcode topics in `biapi-inventory-consumer` only. Existing consumer connections remain plaintext as in the original implementation; the new producer retains its existing TLS configuration support.
- For rollback, stop writes while resolving/draining Product intents with a compatible worker. Preserve intent history, `product_projection_fences` and consumer offsets as a coordinated recovery set. Rolling back only a writer/consumer removes its fence/lock guarantee; reconcile before resuming. No live replay, accounting migration or deployment was performed in this task.

## MongoModel alignment

Revision 1408 → 1412 added lease fields and subtype descriptions; 1412 → 1416 added the queue index `(aggregatetype,status,aggregateuid,version)` and draft technical workflow `product_projection_delivery` (18 steps/22 transitions). The current revision is 1423, including company Barcode/Unit transactions, helper reconciliation and Kafka rebalance evidence. Its data-access references use resolved collection/field IDs. All 12 workflows now lint with zero issues; description and model error-level lints passed.

The six `system_setup_steps` reference errors were field names used instead of existing field IDs. Those references were corrected while preserving approved status, business text and transitions. Unit implementation still uses holding-scoped `units/unitcode`, whereas the design uses company-scoped `unit_of_measure/code`; the technical workflow records this difference without inventing a schema migration. MCP's collection-index schema still cannot represent `partialFilterExpression`/index names. The exact Product partial-unique specification is documented in the collection description and source; the existing generic unique index was preserved. Structural index alignment is therefore incomplete. Model updates do not migrate a live database.
