# 10,000 Concurrent Screens Capacity Gate

## Objective

Prove whether BC Ai Account can serve 10,000 concurrent browser screens. K3s gives horizontal scaling, but capacity must be measured end to end.

## Definition Of "10,000 Screens"

Use one measurable target per test run:

- 10,000 authenticated browser sessions idle on menu/workspace screens.
- 10,000 sessions refreshing normal ERP reads.
- Mixed workload with reads, writes, uploads, and report creation.

Do not mix these definitions in one number.

## Required Test Scenarios

| Scenario | What to measure |
| --- | --- |
| Login and workspace | auth latency, token/session errors, MongoDB latency |
| Menu and common lookup | p95/p99 API latency, Redis/cache hit rate |
| ERP list/table reads | PostgreSQL pool wait, slow queries, row limits |
| Document writes | transaction success rate, deadlocks, Kafka lag |
| Kafka consumers | consumer lag per group/topic/partition, duplicate processing rate, DLQ count |
| Reports | ClickHouse query time, report worker queue, PDF generation time |
| Image upload/download | object storage latency, network egress, error rate |
| Language/manual API | cache behavior and response size |

## Pass Criteria

Set exact thresholds per customer, but every run must report:

- concurrent sessions
- requests per second
- p50/p95/p99 latency
- HTTP error rate
- pod CPU/memory
- HPA replica count
- PostgreSQL open/idle/in-use/wait count
- MongoDB slow operations
- ClickHouse query latency
- Kafka consumer lag
- object storage latency

## K3s Scale Checklist

0. Smoke test the manifests in Docker Desktop/k3d with `backend/cluster/k3s/overlays/k3d`.
1. Start with `mainapi` min 6 replicas and max 40 replicas.
2. Start `transaction-consumer` min 2 replicas and scale only after idempotency and Kafka lag are measured.
3. Confirm metrics-server works: `kubectl top pods -n bc-ai-account`.
4. Confirm HPA reads metrics: `kubectl -n bc-ai-account describe hpa mainapi` and `kubectl -n bc-ai-account describe hpa transaction-consumer`.
5. Run baseline load at 500, 1,000, 2,500, 5,000, then 10,000 sessions.
6. Stop at the first bottleneck and fix that layer before increasing load.
7. Keep report generation and long jobs out of interactive pods if they affect menu/API latency.

## Known Risks In Current Codebase

- `bootstrap.json` is mutable local config. Multiple replicas can drift if Settings writes config inside one pod.
- Some code paths allow large page sizes such as 10,000 rows. Those need pagination/index review before high concurrency.
- Report generation shares backend resources with interactive APIs. Heavy reports can steal capacity from normal screens.
- Single-node Kafka/ClickHouse/SeaweedFS/Redis from docker-compose is not production HA.
- Consumer scale can increase duplicate-event risk unless every handler is idempotent by `event_id` or `shopid + docno + version`.
