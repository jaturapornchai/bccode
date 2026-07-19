---
source: process-doc-purchase-model.go
tags: [datamodel, general-type]
---

# PurchaseStatusDocDetailReferStruct

รายการอ้างอิงการรับสินค้าของบรรทัดเอกสารซื้อ ระบุเอกสารที่มารับสินค้าและจำนวนที่รับ

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocNo | string | docno | docno | เลขที่เอกสารที่อ้างอิง (เอกสารรับสินค้า) |
| TransFlag | int | transflag | transflag | ประเภทเอกสาร (trans flag) |
| ReceivedQty | float64 | - | receivedqty | จำนวนที่รับ |

## ความสัมพันธ์

- `DocNo` → อ้างอิงเลขที่เอกสารรับสินค้า
- ถูก embed อยู่ใน [[PurchaseStatusDocDetailStruct]] (field `DocRefer`)
