# High-Scale Multi-Tenant BI Architecture

## Objective

Design BC Ai Account for many concurrent users while preserving the current data flow:

```text
MongoDB -> Kafka -> PostgreSQL -> Kafka -> ClickHouse -> Dashboard/BI/Reports
```

The design must support one owner seeing the overview of many companies/businesses and many branches without weakening tenant isolation.

## Store Roles

- MongoDB is the authoritative operational source for all CRUD, documents, master data, and user-entered business data.
- Cloudflare R2/S3 stores images, attachments, and binary files. MongoDB stores only file metadata and private paths.
- PostgreSQL is the relational processing/projection layer for postings, balances, tax/VAT, AR/AP, GL, stock costing, and strict relational calculations.
- ClickHouse is the BI/analytics/reporting layer fed from processed facts. It is read-only for BI/report consumers and is never the transactional source of truth.

## Tenant Model

Use four scopes:

```text
user_id           = login identity: owner, staff, accountant, support user
company_group_id  = group of companies owned/managed together
tenant_id         = one company/business/workspace
branch_id         = one branch under tenant_id
```

Rules:

- `tenant_id` is the canonical tenant boundary.
- For existing production data, `tenant_id` must use the same value as the existing core `holdingcode`.
- Existing core storage uses `holdingcode` as the physical storage field. Some newer GoAPI/MCP modules expose `holdingcode` at the API/DTO layer, so treat `tenant_id` as the logical/API name and map to the module's real field instead of renaming old data.
- `tenant_id` must represent the selected company/business, not the owner user.
- One `user_id` can access many `tenant_id` values through membership and roles.
- One `company_group_id` can contain many `tenant_id` values for owner-level overview.
- One `tenant_id` can contain many `branch_id` values.
- Backend must derive allowed tenants from authenticated membership. Frontend-provided tenant values are context only, not proof of permission.

Recommended access tables:

```text
users(id, ...)
company_groups(id, owner_user_id, ...)
tenants(tenant_id, company_group_id, ...)
branches(id, tenant_id, code, ...)
tenant_members(user_id, tenant_id, role, status, ...)
```

Detailed admin, owner, tenant, and branch access rules are defined in `admin-access-control.md`.

## Write Workflow

### 1. Source Of Truth

MongoDB remains the document source of truth for transactional documents.

Every write must include:

```text
company_group_id
tenant_id          # same value as holdingcode for existing tenants
branch_id
doc_type
doc_no
doc_date
source_version
updated_by_user_id
```

### 2. Outbox

Do not publish directly to Kafka inside business code after partially writing MongoDB. Use an outbox record written atomically with the document write.

Outbox key fields:

```text
event_id
event_type
event_version
company_group_id
tenant_id          # same value as holdingcode for existing tenants
branch_id
doc_type
doc_no
doc_date
payload_hash
occurred_at
```

### 3. Kafka Document Event

Publish document events from the outbox to Kafka.

Default topic pattern:

```text
erp.document.changed.v1
erp.document.deleted.v1
erp.document.dlq.v1
```

Partition key:

```text
tenant_id|doc_type|doc_no
```

This keeps ordering for the same document. If one large tenant becomes a hot partition, add a documented partition bucket while keeping per-document ordering.

### 4. PostgreSQL Processor

PostgreSQL consumers process business totals:

- stock movement and balance
- AR/AP balance
- GL posting
- payment/deposit totals
- customer/supplier balances

Consumers must be idempotent:

```text
processed_events(event_id, consumer_name, processed_at, payload_hash)
```

Use `doc_no + tenant_id + event_version` or `event_id` to avoid double posting when Kafka redelivers.

### 5. Kafka Posted Fact Event

PostgreSQL emits posted fact events after successful processing. These are the only events that official BI/report dashboards should use.

Default topic pattern:

```text
erp.stock.posted.v1
erp.sales.posted.v1
erp.purchase.posted.v1
erp.arap.posted.v1
erp.gl.posted.v1
erp.posted.dlq.v1
```

Why posted facts are required:

- Raw MongoDB events show document changes.
- PostgreSQL knows business posting rules and final balances.
- ClickHouse reports must match processed accounting/stock results, not raw draft data.

## ClickHouse Design

ClickHouse is read-only analytics/reporting infrastructure. It must not be the transactional source of truth.

### Required Columns

Every ClickHouse fact table must include:

```text
event_id
event_version
company_group_id
tenant_id
branch_id
doc_type
doc_no
doc_date
posted_at
item_id
customer_id
supplier_id
account_id
qty
amount
cost
```

Unused dimension columns may be empty, but the tenant keys must never be empty.

### Table Strategy

Use shared tables, not one table/database per tenant.

Recommended ordering for most business reports:

```sql
PARTITION BY toYYYYMM(doc_date)
ORDER BY (company_group_id, tenant_id, branch_id, doc_date, doc_type, doc_no)
```

Use replicated engines in production clusters. Use append-only facts plus reversal/correction events for audited financial and stock data. Avoid silent in-place mutation for official reports.

### Aggregates

Create aggregate tables/materialized views for common screens:

```text
summary_daily_by_company_group
summary_daily_by_tenant
summary_daily_by_branch
summary_stock_balance_by_tenant
summary_arap_balance_by_tenant
summary_gl_by_tenant
```

Owner overview uses:

```sql
WHERE company_group_id = ?
  AND tenant_id IN (...)
```

Tenant dashboard uses:

```sql
WHERE tenant_id = ?
```

Branch dashboard uses:

```sql
WHERE tenant_id = ?
  AND branch_id = ?
```

The allowed `tenant_id` list must come from backend authorization, not from the browser.

## Read Workflow

### Normal Business Screens

Use MongoDB/PostgreSQL APIs for transactional screens that need current editable documents or strict balances.

### Dashboard And BI

Use ClickHouse through backend APIs only:

```text
Frontend -> Backend report API -> Authorization -> ClickHouse query -> response
```

Frontend must never connect to ClickHouse directly.

### Owner Overview

Flow:

1. User logs in.
2. Backend loads `tenant_members`.
3. Backend derives allowed `tenant_id` list and `company_group_id` values.
4. User selects all companies, one company, or one branch.
5. Backend queries ClickHouse with authorized filters only.

## Scale And Deployment

High-concurrency production deployment target must be explicitly approved and backed by load-test evidence before it is treated as supported.

Recommended separation:

```text
interactive-api pods     = login, menu, workspace, business reads/writes
worker-pg pods           = Kafka -> PostgreSQL processing
worker-clickhouse pods   = Kafka -> ClickHouse ingestion
report-api pods          = ClickHouse report APIs
report-renderer pods     = PDF/export jobs
```

Do not let heavy report/export jobs run inside the same pods that serve login/menu APIs at high load.

## Backpressure And Recovery

Every consumer must expose metrics:

```text
kafka_lag
processed_per_second
failed_events
dlq_events
retry_count
processing_latency_ms
```

Required failure handling:

- Retry transient database/network errors.
- Move poison messages to DLQ with tenant metadata.
- Keep original payload and error reason in DLQ.
- Provide replay tools that can replay by `event_id`, `tenant_id`, date range, and topic.
- Alert when lag grows for PG or ClickHouse consumers.

## Security

- Every API must check authenticated membership before accepting `tenant_id`.
- Every query must filter by `tenant_id` or an authorized list of `tenant_id` values.
- Cross-company owner overview must be explicit and audited.
- Support/admin cross-tenant access must be read-only by default and audit logged.
- Do not log PII, passwords, tokens, or full customer documents in Kafka/error logs.
- Kafka events may contain business-sensitive data; secure broker auth, network access, and retention.

## Config

Minimum production config:

```text
KAFKA_BROKERS
KAFKA_SECURITY_PROTOCOL
KAFKA_CONSUMER_GROUP
MONGODB_URI
POSTGRES_DSN
CLICKHOUSE_DSN
REDIS_URL
OBJECT_STORAGE_ENDPOINT
```

Production must load secrets through an approved secret manager or environment-specific secure configuration. Do not put real credentials in repository files.

## Dependencies

- MongoDB replica set or cluster for reliable source writes and change tracking.
- Kafka cluster with enough partitions and retention for replay.
- PostgreSQL sized for posting workloads and indexed tenant queries.
- ClickHouse cluster for BI/report query volume.
- Redis/cache for hot menu/session/read metadata when needed.
- An approved production orchestration path for horizontal scaling.

## Usage Example

Owner opens dashboard for all companies:

```text
GET /api/v1/company-groups/:companyGroupId/overview?from=2026-05-01&to=2026-05-31
Authorization: Bearer ...
```

Backend behavior:

1. Verify user access to `companyGroupId`.
2. Load allowed `tenant_id` list.
3. Query ClickHouse summary tables by `company_group_id` and allowed `tenant_id`.
4. Return totals grouped by company and branch.

## Limitations

- This design does not prove capacity by itself. Use measured load-test gates before claiming large concurrent screen capacity.
- ClickHouse data is eventually consistent with PostgreSQL because it is event-driven.
- Kafka can redeliver messages; all consumers must be idempotent.
- Historical data needs backfill into `company_group_id`, `tenant_id`, and `branch_id` before owner overview can be trusted.
- Large tenants can create Kafka partition skew; monitor lag per partition and adjust partition strategy with evidence.
