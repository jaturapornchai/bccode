# Procurement Flow — PR → RFQ → PO

## Document Flow (มาตรฐาน)

```
PR (ใบขอซื้อ) → [อนุมัติ PR] → RFQ (สืบราคา) → [อนุมัติราคา] → PO (ใบสั่งซื้อ)
     ↑                              ↑                              ↑
  แผนกส่งคำขอ              ขอราคาจาก vendor 3+         เลือก vendor ที่ดีที่สุด
  ระบุเหตุผล+งบ            เปรียบเทียบ side-by-side     อ้างอิง PR + RFQ
```

- PR สามารถข้าม RFQ ไปสร้าง PO ตรงได้ (กรณีไม่ต้องสืบราคา)
- PO สามารถเลือกอ้าง PR หรือ RFQ ก็ได้
- ทั้ง PR และ RFQ มีระบบอนุมัติแบบเดียวกับ PO (multi-level, generic engine)

---

## TransFlag

| transflag | ประเภท | Prefix DocNo |
|-----------|--------|-------------|
| 20 | ใบสั่งซื้อ (Purchase Order) | PO |
| 21 | ใบขอซื้อ (Purchase Requisition) | PR |
| 22 | สืบราคา (Request for Quotation) | RFQ |

---

## PR — ใบขอซื้อ (Purchase Requisition)

### MongoDB Collection
`transactionPurchaseRequisition`

### PostgreSQL Tables
- `purchase_requisition_transaction`
- `purchase_requisition_transaction_detail`

### Kafka Topics
- `when-purchaserequisition-created/updated/deleted`
- `when-purchaserequisition-bulk-created/updated/deleted`

### PR-specific Fields (เพิ่มจาก Transaction base)
| Field | Type | คำอธิบาย |
|-------|------|----------|
| `requestercode` | string | รหัสผู้ขอซื้อ (รหัสพนักงาน) |
| `requestername` | string | ชื่อผู้ขอซื้อ |
| `departmentcode` | string | รหัสแผนก |
| `departmentnames` | []NameX | ชื่อแผนก (multi-lang) |
| `purpose` | string | วัตถุประสงค์/เหตุผล |
| `budgetcode` | string | รหัสงบประมาณ |
| `budgetamount` | float64 | วงเงินงบ |
| `urgency` | int8 | 1=ปกติ 2=เร่งด่วน 3=เร่งด่วนมาก |
| `requesteddeliverydate` | string | วันที่ต้องการรับ |
| `refpodocno` | string | เลขที่ PO ที่สร้างจาก PR |
| `refrfqdocno` | string | เลขที่ RFQ ที่สร้างจาก PR |
| `conversionstatus` | string | none/converted_to_rfq/converted_to_po |

### API Endpoints
```
POST   /transaction/purchase-requisition           → Create
GET    /transaction/purchase-requisition/:id        → Info
PUT    /transaction/purchase-requisition/:id        → Update
DELETE /transaction/purchase-requisition/:id        → Delete
GET    /transaction/purchase-requisition            → SearchPage
GET    /transaction/purchase-requisition/list       → SearchStep
POST   /transaction/purchase-requisition/bulk       → SaveBulk
```

### Approval Endpoints (goapi)
```
POST /api/approval/pr-settings       → Get settings
POST /api/approval/pr-setting/save   → Save setting
POST /api/approval/pr-setting/delete → Delete setting
POST /api/approval/pr-status/get     → Get status
POST /api/approval/pr-status/batch   → Batch status
POST /api/approval/pr-status/submit  → Submit for approval
POST /api/approval/pr-status/approve → Approve
POST /api/approval/pr-status/reject  → Reject
POST /api/approval/pr-status/withdraw → Withdraw
POST /api/approval/pr-status/pending  → List pending
POST /api/approval/pr-status/rejected → List rejected
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

## RFQ — สืบราคา (Request for Quotation)

### MongoDB Collection
`transactionRequestForQuotation`

### PostgreSQL Tables
- `rfq_transaction`
- `rfq_transaction_detail`

### Kafka Topics
- `when-rfq-created/updated/deleted`
- `when-rfq-bulk-created/updated/deleted`

### RFQ-specific Fields
| Field | Type | คำอธิบาย |
|-------|------|----------|
| `refprdocno` | string | อ้างอิง PR |
| `refprguidfixed` | string | GUID ของ PR |
| `selectedvendor` | string | vendor ที่เลือก (custcode) |
| `selectionreason` | string | เหตุผลเลือก vendor |
| `refpodocno` | string | เลขที่ PO ที่สร้างจาก RFQ |
| `refpoguidfixed` | string | GUID ของ PO |
| `minvendors` | int8 | จำนวน vendor ขั้นต่ำ |
| `conversionstatus` | string | none/converted_to_po |
| `submissiondeadline` | string | กำหนดส่งใบเสนอราคา |

### VendorEntry (MongoDB only — เก็บใน vendorentries array)
| Field | Type | คำอธิบาย |
|-------|------|----------|
| `vendorcode` | string | รหัส vendor |
| `vendornames` | []NameX | ชื่อ vendor |
| `quotationdocno` | string | เลขที่ใบเสนอราคา |
| `quotationdate` | string | วันที่ใบเสนอราคา |
| `creditdays` | int | เครดิตเทอม (วัน) |
| `deliverydays` | int | ส่งมอบภายใน (วัน) |
| `deliveryterms` | string | เงื่อนไขการส่ง |
| `qualitynotes` | string | หมายเหตุคุณภาพ |
| `totalamount` | float64 | ยอดรวม |
| `isselected` | bool | เป็น vendor ที่เลือก |
| `items` | []VendorItem | รายการสินค้า+ราคา |

### API Endpoints
```
POST   /transaction/rfq           → Create
GET    /transaction/rfq/:id       → Info
PUT    /transaction/rfq/:id       → Update
DELETE /transaction/rfq/:id       → Delete
GET    /transaction/rfq           → SearchPage
GET    /transaction/rfq/list      → SearchStep
POST   /transaction/rfq/bulk      → SaveBulk
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

ทั้ง PR, RFQ, PO ใช้ generic approval handlers เดียวกัน:
- Handler อยู่ที่ `internal/goapi/handlers/approval/`
- แยกประเภทด้วย `purchase_type_code` ที่ frontend ส่งมา
- Routes แยกด้วย prefix: `/api/approval/po-*`, `/api/approval/pr-*`, `/api/approval/rfq-*`
- รองรับ multi-level approval + LINE OA notification + LIFF token

---

## Status Flow (ใช้ร่วมกัน PR/RFQ/PO)

```
draft → pending_approval → approved → converted (to RFQ/PO)
                         → rejected → draft (แก้ไขส่งใหม่)
```
