# Backup and isolated recovery readiness

Updated 2026-09-05. This is a preparation runbook, not evidence that production backups exist or that a restore has passed.

## Objective

Prove recovery into a separate environment without overwriting active data. Agree the acceptable data-loss window (RPO), recovery time (RTO), retention and backup destination before enabling a schedule.

## Source-backed inventory

`compose.yml` declares persistent MongoDB, PostgreSQL, ClickHouse, Redis, MinIO and Kafka volumes. `provision-server.sh` writes runtime configuration under `/etc/bcai-account/`; that configuration includes secrets and must not be copied into the repo or incident notes. The MongoDB rollout rule in `backend/README.md` requires fresh DEV/UAT/PRO databases unless a separate migration/import task is approved.

| Component | Evidence and recovery concern |
| --- | --- |
| MongoDB | Replica set rs0; source documents and pending/published outbox history must survive together. |
| PostgreSQL | Compose specifies version 18; preserve schema, constraints, exact stored totals and `product_projection_fences`. Local projection integration now passes on PostgreSQL 18; this is not a restore drill. |
| MinIO | Bucket versioning is enabled by minio-init. Retain original objects and thumbnail objects referenced by MongoDB; versioning on the same disk is not an independent backup. |
| Kafka | Preserve topic/partition configuration and consumer offsets, and reconcile them with restored outbox history before replay. |
| ClickHouse | Inventory actual datasets and identify which are reproducible before selecting backup scope. |
| Redis | AOF volume exists; determine whether each key family is cache, session or durable workflow state before excluding it. |

## Workflow

1. Inspect existing host/cloud backup jobs and latest successful restore evidence. The repository search did not establish their external status.
2. Record approved source/target identifiers, component versions, backup timestamps, retention, encrypted destination and restore owner. Record secret references only.
3. Define a consistent recovery point across source records, objects, pending events and projections. Do not independently restore arbitrary timestamps and assume they agree.
4. Provision an isolated recovery environment with separate volumes, databases, object bucket and network identity. Disable outbound notifications and live event delivery before starting application workers.
5. Restore the approved snapshot using version-compatible tools. Validate target identifiers before every restore command. No executable restore commands are provided here until a real target and backup format are confirmed.
6. Verify counts and exact accounting totals by currency/scale from persisted values; preserve tenant scopes, unique indexes, soft-delete/audit history, user access and original/thumbnail object pairs.
7. Reconcile pending/published outbox history, per-aggregate Product version fences, consumer offsets and projection rows as one recovery set. A restored fence ahead of a restored row can cause valid replay to be skipped; an offset ahead of source/outbox history can lose a signal. Preserve the selected checkpoint and explicitly reconcile before enabling workers. Test controlled duplicate delivery, delete/recreate and missing-delivery recovery inside the isolated environment.
8. Record start/end time, recovered timestamp, missing data, exceptions and measured RTO/RPO. A passing dump command alone is insufficient.
9. Clean only the explicitly named drill resources after evidence is recorded.

## Dependencies, example evidence and limitation

Requires the actual backup platform, destination, compatible database/object tools, credentials supplied through the approved secret mechanism, isolated resources and an approved reconciliation baseline.

Example drill record (no credentials or customer data): environment identifier; snapshot timestamp; component versions; record/object counts; exact-total comparison result; thumbnail-link check; tenant-isolation check; event-recovery check; elapsed restore time; failures; owner; artifact reference.

Rollback for a failed drill: stop the isolated workers, preserve diagnostic metadata and discard only the explicitly identified drill targets. Production is not a restore-test target. No backup schedule, external alert or actual restore was created or executed in this task.
