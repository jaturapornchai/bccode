# Procurement Flow — PR -> RFQ -> PO

## Document Flow (Standard)

```
PR (Purchase Requisition) -> [Approve PR] -> RFQ (Request for Quotation) -> [Approve Price] -> PO (Purchase Order)
     ^                              ^                              ^
  Department request        Get quotes from 3+ vendors     Select best vendor
  Specify reason + budget   Compare side-by-side           Reference PR + RFQ
```

- PR can skip RFQ and create PO directly (when price comparison not needed)
- PO can reference either PR or RFQ
- PR and RFQ share the same approval system as PO (multi-level, generic engine)

---

## TransFlag

| transflag | Type | DocNo Prefix |
|-----------|------|-------------|
| 20 | Purchase Order (PO) | PO |
| 21 | Purchase Requisition (PR) | PR |
| 22 | Request for Quotation (RFQ) | RFQ |

---

## PR — Purchase Requisition

### MongoDB Collection
`transactionPurchaseRequisition`

### PostgreSQL Tables
- `purchase_requisition_transaction`
- `purchase_requisition_transaction_detail`

### Kafka Topics
- `when-purchaserequisition-created/updated/deleted`
- `when-purchaserequisition-bulk-created/updated/deleted`

### PR-specific Fields (additional to Transaction base)
| Field | Type | Description |
|-------|------|-------------|
| `requestercode` | string | Requester code (employee ID) |
| `requestername` | string | Requester name |
| `departmentcode` | string | Department code |
| `departmentnames` | []NameX | Department names (multi-lang) |
| `purpose` | string | Purpose/reason |
| `budgetcode` | string | Budget code |
| `budgetamount` | float64 | Budget amount |
| `urgency` | int8 | 1=Normal, 2=Urgent, 3=Critical |
| `requesteddeliverydate` | string | Required delivery date |
| `refpodocno` | string | PO doc number created from PR |
| `refrfqdocno` | string | RFQ doc number created from PR |
| `conversionstatus` | string | none/converted_to_rfq/converted_to_po |

### API Endpoints
```
POST   /transaction/purchase-requisition           -> Create
GET    /transaction/purchase-requisition/:id        -> Info
PUT    /transaction/purchase-requisition/:id        -> Update
DELETE /transaction/purchase-requisition/:id        -> Delete
GET    /transaction/purchase-requisition            -> SearchPage
GET    /transaction/purchase-requisition/list       -> SearchStep
POST   /transaction/purchase-requisition/bulk       -> SaveBulk
```

### Approval Endpoints (goapi)
```
POST /api/approval/pr-settings       -> Get settings
POST /api/approval/pr-setting/save   -> Save setting
POST /api/approval/pr-setting/delete -> Delete setting
POST /api/approval/pr-status/get     -> Get status
POST /api/approval/pr-status/batch   -> Batch status
POST /api/approval/pr-status/submit  -> Submit for approval
POST /api/approval/pr-status/approve -> Approve
POST /api/approval/pr-status/reject  -> Reject
POST /api/approval/pr-status/withdraw -> Withdraw
POST /api/approval/pr-status/pending  -> List pending
POST /api/approval/pr-status/rejected -> List rejected
```

### Backend Files
```
internal/transaction/purchaserequisition/
├── purchaserequisition_http.go
├── models/purchaserequisition.go
├── services/purchaserequisition_http_service.go
├── repositories/purchaserequisition_mongo_repository.go
├── repositories/purchaserequisition_messagequeue_repository.go
├── config/purchaserequisition_messagequeue_config.go
└── validators/
    ├── pr_validator.go
    └── validation_result.go

internal/transaction/models/
└── transaction_purchase_requisition_postgres.go

internal/transaction/transactionconsumer/purchaserequisition/
├── purchase_requisition_consumer_service.go
├── purchase_requisition_transaction_consumer.go
├── purchase_requisition_transaction_phaser.go
└── purchase_requisition_transaction_postgres_repository.go
```

---

## RFQ — Request for Quotation

### MongoDB Collection
`transactionRequestForQuotation`

### PostgreSQL Tables
- `rfq_transaction`
- `rfq_transaction_detail`

### Kafka Topics
- `when-rfq-created/updated/deleted`
- `when-rfq-bulk-created/updated/deleted`

### RFQ-specific Fields
| Field | Type | Description |
|-------|------|-------------|
| `refprdocno` | string | Reference PR doc number |
| `refprguidfixed` | string | PR GUID |
| `selectedvendor` | string | Selected vendor (custcode) |
| `selectionreason` | string | Vendor selection reason |
| `refpodocno` | string | PO doc number created from RFQ |
| `refpoguidfixed` | string | PO GUID |
| `minvendors` | int8 | Minimum vendor count |
| `conversionstatus` | string | none/converted_to_po |
| `submissiondeadline` | string | Quotation submission deadline |

### VendorEntry (MongoDB only — stored in vendorentries array)
| Field | Type | Description |
|-------|------|-------------|
| `vendorcode` | string | Vendor code |
| `vendornames` | []NameX | Vendor names |
| `quotationdocno` | string | Quotation document number |
| `quotationdate` | string | Quotation date |
| `creditdays` | int | Credit term (days) |
| `deliverydays` | int | Delivery within (days) |
| `deliveryterms` | string | Delivery terms |
| `qualitynotes` | string | Quality notes |
| `totalamount` | float64 | Total amount |
| `isselected` | bool | Is selected vendor |
| `items` | []VendorItem | Line items with prices |

### API Endpoints
```
POST   /transaction/rfq           -> Create
GET    /transaction/rfq/:id       -> Info
PUT    /transaction/rfq/:id       -> Update
DELETE /transaction/rfq/:id       -> Delete
GET    /transaction/rfq           -> SearchPage
GET    /transaction/rfq/list      -> SearchStep
POST   /transaction/rfq/bulk      -> SaveBulk
```

### Approval Endpoints (goapi)
Same pattern as PR — prefix `rfq-` (e.g., `/api/approval/rfq-settings`)

### Backend Files
```
internal/transaction/rfq/
├── rfq_http.go
├── models/rfq.go
├── services/rfq_http_service.go
├── repositories/rfq_mongo_repository.go
├── repositories/rfq_messagequeue_repository.go
├── config/rfq_messagequeue_config.go
└── validators/
    ├── rfq_validator.go
    └── validation_result.go

internal/transaction/models/
└── transaction_rfq_postgres.go

internal/transaction/transactionconsumer/rfq/
├── rfq_consumer_service.go
├── rfq_transaction_consumer.go
├── rfq_transaction_phaser.go
└── rfq_transaction_postgres_repository.go
```

---

## Approval System (Generic Engine)

PR, RFQ, and PO all share the same generic approval handlers:
- Handlers at `internal/goapi/handlers/approval/`
- Differentiated by `purchase_type_code` sent from frontend
- Routes separated by prefix: `/api/approval/po-*`, `/api/approval/pr-*`, `/api/approval/rfq-*`
- Supports multi-level approval + LINE OA notification + LIFF token

---

## Status Flow (shared across PR/RFQ/PO)

```
draft -> pending_approval -> approved -> converted (to RFQ/PO)
                          -> rejected -> draft (edit and resubmit)
```
