---
source: process-model.go
tags: [datamodel, general-type]
---

# ProcessStockCostDetailStruct

รายละเอียดรายการสต๊อกสำหรับคำนวณต้นทุน (cost) ต่อบรรทัดเอกสาร ไม่มี bson/json tag — ใช้ภายในกระบวนการ process เท่านั้น

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | - | รหัสกิจการ (tenant) |
| DocDateTime | time.Time | - | - | วันที่/เวลาเอกสาร |
| DocNo | string | - | - | เลขที่เอกสาร |
| LineNumber | int | - | - | ลำดับบรรทัดในเอกสาร |
| TransFlag | int | - | - | ประเภทรายการ (trans flag) |
| ItemCode | string | - | - | รหัสสินค้า |
| Barcode | string | - | - | บาร์โค้ด |
| BarcodeMain | string | - | - | บาร์โค้ดหลัก |
| UnitCode | string | - | - | รหัสหน่วยนับ |
| WhCode | string | - | - | รหัสคลังสินค้า |
| LocationCode | string | - | - | รหัสที่เก็บ |
| Qty | float64 | - | - | จำนวน |
| Price | float64 | - | - | ราคา |
| PriceExcludeVat | float64 | - | - | ราคาไม่รวม VAT |
| UnitStand | float64 | - | - | ตัวคูณหน่วยมาตรฐาน |
| UnitDivide | float64 | - | - | ตัวหารหน่วย |
| UnitCost | float64 | - | - | ต้นทุนต่อหน่วย |
| AverageCost | float64 | - | - | ต้นทุนเฉลี่ย |
| SumAmount | float64 | - | - | มูลค่ารวม |
| CalcAmount | float64 | - | - | มูลค่าที่คำนวณได้ |
| BalanceAmount | float64 | - | - | มูลค่าคงเหลือ |
| BalanceQty | float64 | - | - | จำนวนคงเหลือ |
| Guid | string | - | - | GUID ของรายการ |
| DocRef | string | - | - | เลขที่เอกสารอ้างอิง |

## ความสัมพันธ์
- `HoldingCode` อ้างอิงกิจการ (tenant), `ItemCode`/`Barcode`/`BarcodeMain` อ้างอิงสินค้า/บาร์โค้ด, `WhCode` อ้างอิงคลัง, `LocationCode` อ้างอิงที่เก็บ, `DocNo`/`DocRef` อ้างอิงเอกสาร
