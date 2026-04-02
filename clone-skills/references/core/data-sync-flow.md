# Data Sync Flow — MongoDB -> Kafka -> PostgreSQL + ClickHouse

## Core Rule (mandatory)

Every entity (master data + transaction) must flow through all stores:
```
MongoDB (primary store)
    | Kafka publish (goroutine, async)
    +-> PostgreSQL (consumer upsert) — structured queries, reporting
    +-> ClickHouse (consumer upsert) — analytics, time-series
```

## Pre-merge Checklist

- [ ] Entity has Kafka MQ config (`config/*_messagequeue_config.go`)
- [ ] Entity has MQ repository (`repositories/*_messagequeue_repository.go`)
- [ ] HTTP handler passes `prod := ms.Producer(cfg.MQConfig())` to service
- [ ] Service publishes in goroutine after MongoDB save
- [ ] PostgreSQL model exists (`models/*_pg.go`) + correct table name
- [ ] ClickHouse repository has upsert + delete
- [ ] Consumer registered in `cmd/transaction_consumer/main.go`
- [ ] Fields match across MongoDB model <-> PG model <-> CH table

## Pattern: Master Data (Organization / Product)

### File Structure
```
internal/organization/costcenter/
├── config/costcenter_messagequeue_config.go     <- Kafka topic names
├── repositories/
│   ├── costcenter_mongo_repository.go           <- MongoDB CRUD
│   ├── costcenter_messagequeue_repository.go    <- Kafka publish
│   └── costcenter_pg_repository.go              <- PostgreSQL CRUD (for HTTP service)
├── models/
│   ├── costcenter.go                            <- MongoDB model (Doc, Info, etc.)
│   └── costcenter_pg.go                         <- PostgreSQL model (GORM)
├── services/costcenter_http_service.go          <- Business logic + publish Kafka
└── costcenter_http.go                           <- Route handler + wire dependencies

internal/transaction/transactionconsumer/costcenter/
├── costcenter_consumer.go                       <- Kafka listener, register topics
├── costcenter_consumer_service.go               <- Upsert/Delete (PG + CH)
├── costcenter_pg_repository.go                  <- Consumer-side PG repo
└── costcenter_ch_repository.go                  <- ClickHouse INSERT/DELETE
```

### Kafka Topics
```
when-organization-costcenter-created
when-organization-costcenter-updated
when-organization-costcenter-deleted
when-organization-costcenter-bulk-created
when-organization-costcenter-bulk-updated
when-organization-costcenter-bulk-deleted
```

### Service Publish Pattern
```go
// In CreateXxx, UpdateXxx, DeleteXxx — call after MongoDB save
go func() {
    svc.repoMessageQueue.Create(docData)
    svc.saveMasterSync(shopID)
}()
```

### HTTP Handler Wire Pattern
```go
func NewCostCenterHttp(ms *microservice.Microservice, cfg config.IConfig) CostCenterHttp {
    pst   := ms.MongoPersister(cfg.MongoPersisterConfig())
    cache := ms.Cacher(cfg.CacherConfig())
    prod  := ms.Producer(cfg.MQConfig())           // <- add this line

    repo             := repositories.NewCostCenterRepository(pst)
    repoMessageQueue := repositories.NewCostCenterMessageQueueRepository(prod) // <- add
    masterSyncRepo   := mastersync.NewMasterSyncCacheRepository(cache)
    svc              := services.NewCostCenterHttpService(repo, repoMessageQueue, masterSyncRepo)
    ...
}
```

## Pattern: Transaction Data (PO, PR, RFQ)

See `internal/transaction/purchaseorder/` + `internal/transaction/transactionconsumer/purchaseorder/`

Transaction consumer uses `TransactionPG` base struct (has DocNo, DocDate, TransFlag, etc.)

## PostgreSQL Model Example

```go
// models/costcenter_pg.go
type CostCenterPg struct {
    ShopID    string          `json:"shopid"    gorm:"column:shopid;primaryKey"`
    GuidFixed string          `json:"guidfixed" gorm:"column:guidfixed;uniqueIndex"`
    Code      string          `json:"code"      gorm:"column:code"`
    Names     pkgModels.JSONB `json:"names"     gorm:"column:names;type:jsonb"`
}

func (CostCenterPg) TableName() string { return "organization_cost_center" }
```

## ClickHouse Table DDL

```sql
-- organization_cost_center
CREATE TABLE IF NOT EXISTS organization_cost_center (
    shopid    String,
    guidfixed String,
    code      String,
    names     String,   -- JSON string
    is_delete UInt8 DEFAULT 0,
    updated_at DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (shopid, guidfixed);
```

## ClickHouse Write Pattern

```go
// ch_repository.go
func (repo *CostCenterCHRepository) Upsert(doc models.CostCenterPg) error {
    namesJSON, _ := json.Marshal(doc.Names)
    return repo.pst.Conn().Exec(context.Background(),
        `INSERT INTO organization_cost_center (shopid, guidfixed, code, names, is_delete)
         VALUES (?, ?, ?, ?, 0)`,
        doc.ShopID, doc.GuidFixed, doc.Code, string(namesJSON),
    )
}

func (repo *CostCenterCHRepository) Delete(shopID, guidFixed string) error {
    return repo.pst.Conn().Exec(context.Background(),
        `ALTER TABLE organization_cost_center DELETE WHERE shopid=? AND guidfixed=?`,
        shopID, guidFixed,
    )
}
```

## Consumer Registration

```go
// cmd/transaction_consumer/main.go
import (
    costcenter_consumer "smlcloudplatform/internal/transaction/transactionconsumer/costcenter"
    jobproject_consumer "smlcloudplatform/internal/transaction/transactionconsumer/jobproject"
)

ms.RegisterConsumer(costcenter_consumer.InitCostCenterConsumer(ms, cfg))
ms.RegisterConsumer(jobproject_consumer.InitJobProjectConsumer(ms, cfg))
```

## Existing Tables (update when creating new ones)

| Entity | MongoDB Collection | PG Table | CH Table | Status |
|--------|-------------------|----------|----------|--------|
| CostCenter | organizationCostCenters | organization_cost_center | organization_cost_center | Done (2026-03-15) |
| JobProject | organizationJobProjects | organization_job_project | organization_job_project | Done (2026-03-15) |
| Department | organizationDepartments | -- | -- | Not yet synced |
| Unit | productUnit | productbarcode (PG) | productbarcode (CH) | Has Kafka publish |
| PurchaseOrder | transactionPurchaseOrder | purchase_order_transaction | -- | Has PG consumer |

## When Creating a New Entity

1. Check if PG/CH tables already exist (see table above)
2. If not -> create immediately following patterns above
3. Update this table
