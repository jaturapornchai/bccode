# Procurement Feature Matrix

อัปเดตล่าสุด: 2026-03-17 (session 6 — Multi-currency + DocumentTotalWidget + Department fix + Internal Note + Purchasing Group + Tolerance + Expiry + Vendor Auto-fill + Budget Check)

## สถานะรวม: BC Account PR = 43 done / 47 → เป้าหมาย 43/47 (91%)

## Legend
- **done** = เสร็จแล้ว ใช้งานได้ (verified from code)
- **partial** = ทำบางส่วนแล้ว
- **in-progress** = กำลังทำ
- **pending** = ยังไม่ได้เริ่ม แต่มี plan
- **planned** = วางแผนไว้ ยังไม่มี code
- **future** = Enterprise feature สำหรับอนาคต

---

## ระดับ 1: ต้องมี (คู่แข่ง 7+ เจ้ามี)

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| 1 | เลขที่เอกสาร auto | **done** | - | docno format "PR" + auto-generate |
| 2 | วันที่เอกสาร | **done** | - | CustomDatePicker (พ.ศ./ค.ศ.) |
| 3 | ชื่อผู้ขอซื้อ | **done** | - | auto-fill จาก profileData |
| 4 | วันที่ต้องการรับ | **done** | - | CustomDatePicker + validation |
| 5 | ความเร่งด่วน 3 ระดับ | **done** | - | 1=ปกติ 2=เร่ง 3=วิกฤต (unique feature!) |
| 6 | เจ้าหนี้/Vendor | **done** | - | supplier search (optional) |
| 7 | วัตถุประสงค์/หมายเหตุ | **done** | - | description field (required) |
| 8 | คลัง/สาขา | **done** | - | warehouse selection |
| 9 | บาร์โค้ด/รหัสสินค้า | **done** | - | barcode search + scan |
| 10 | ชื่อสินค้า | **done** | - | product name (multi-lang) |
| 11 | หน่วยนับ + จำนวน | **done** | - | unit + qty fields |
| 12 | ราคาต่อหน่วย | **done** | - | price column + _getDataText wired |
| 13 | มูลค่ารวม | **done** | - | sum_amount column + calculation |
| 14 | Approval Workflow | **done** | - | reuse PO approval: submit/approve/reject/withdraw |
| 15 | สถานะเอกสาร + Dashboard | **done** | - | List card badge + status dashboard (draft/pending/approved/rejected) |
| 16 | เงื่อนไขชำระเงิน | **done** | - | creditdays UI + save + load ครบ |

## ระดับ 2: ควรมี (คู่แข่ง 4-6 เจ้ามี)

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| 17 | แผนก/ฝ่าย (Department) | **done** | - | extraFields save + transient load ครบ |
| 18 | อนุมัติผ่านมือถือ | **done** | - | LINE OA LIFF approval + push notification + in-app approve/reject |
| 19 | เอกสารอ้างอิง (Ref Doc) | **done** | - | docrefno save + load ครบ |
| 20 | ที่อยู่จัดส่ง (Ship-to) | **done** | - | extraFields save + transient load ครบ |
| 21 | แนบไฟล์ (Attachments) | **done** | - | AttachmentCountIconButton ใน AppBar ครบ |
| 22 | ส่วนลดรายการ | **done** | - | discount column ใน headerTableDetail ครบ |
| 23 | ภาษี VAT | **done** | - | VAT type selector (exclude/include/zero/none) + rate field ครบ |
| 24 | ย้ายลำดับรายการ (Reorder) | **done** | - | reorder column + up/down arrows ครบ |

## ระดับ 3: น่ามี (คู่แข่ง 2-3 เจ้ามี)

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| 25 | Job/โครงการ (Project) | **done** | - | 2-step dialog: เลือกโครงการ → เลือกงาน, cache + display |
| 26 | Cost Center | **done** | - | searchable datalist dialog, cache + display |
| 27 | Comment ตีกลับ | **done** | - | reject dialog + comment field + API ครบ (PO+PR) |
| 28 | แจ้งเตือน email/LINE | **done** | - | SendLinePushApprovalV2 + resend reminder + email notification |
| 29 | ประวัติซื้อสินค้า (Purchase History) | **done** | - | PurchaseHistoryRepository + dialog แสดง stats (avg/min/max/last price) + รายการย้อนหลัง 6 เดือน, ปุ่มต่อรายการสินค้า |
| 30 | คัดลอกเอกสาร (Copy Document) | **done** | - | AppBar copy button + confirm dialog, copy header+details เป็น PR ใหม่ |
| 31 | Serial/Lot Number | **done** | - | optional columns ใน detail table, เก็บใน extrajson, click-to-edit dialog |
| 32 | ประมาณการต้นทุน (Estimated Landed Cost) | **done** | - | 3 fields: estimated_unit_cost, estimated_freight, estimated_duty, save via extraFields |
| 33 | ร้านที่ต้องการ/สำรอง (Vendor Preferences) | **done** | - | preferred_vendor + alternative_vendor + reason_preferred fields |
| 34 | กำหนดวันอนุมัติ (Approval Deadline) | **done** | - | date picker, save via extraFields |
| 35 | Tracking สถานะแปลง | **done** | - | read-only: conversion_status + ref_rfq_docno + ref_po_docno (parse จาก backend) |
| 36a | หมายเหตุภายใน (Internal Note) | **done** | - | โน้ตไม่แสดงใน print — สำหรับทีมจัดซื้อ |
| 36b | กลุ่มจัดซื้อ (Purchasing Group) | **done** | - | แบ่งทีมจัดซื้อ เช่น วัตถุดิบ/IT/ก่อสร้าง |
| 36c | Over/Under Delivery Tolerance | **done** | - | % ยอมรับรับเกิน/ขาด — สำหรับวัตถุดิบเกษตร |
| 36d | Expiry Tracking (วันหมดอายุ) | **done** | - | เก็บใน extrajson ของ detail item — ธุรกิจอาหาร/ยา |
| 36e | Vendor Price Auto-fill | **done** | - | ดึง lastPrice จาก purchase history เมื่อเพิ่มสินค้า — onDetailAdded callback |
| 36f | Budget Check Warning | **done** | - | field budgetcode/budgetamount + warning banner เมื่อเกินงบ (ไม่ block save) |

## ระดับ 4: Enterprise (อนาคต)

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| 36 | งบประมาณ (Budget Check) | **done** | P4 | budgetcode/budgetamount in model แล้ว |
| 37 | สืบราคา/เปรียบเทียบ (RFQ) | **done** | P4 | RFQ edit + list + vendor widget + routes wired |
| 38 | Auto PR จาก MRP/Reorder | **future** | P4 | ต้อง MRP module |
| 39 | สกุลเงินต่างประเทศ | **done** | P4 | multi-currency: DocumentTotalWidget + CurrencyApiService + exchange rate |
| 40 | Vendor Price Catalog | **future** | P4 | auto-fill ราคาเมื่อเลือก vendor + สินค้า |
| 41 | Purchase Agreement/Contract | **future** | P4 | blanket PO + contract reference |

---

## PO Integration (Document Conversion Flow)

| Flow | Status | Notes |
|------|--------|-------|
| PR → RFQ (ปุ่มสร้างใบสืบราคา) | **done** | AppBar button + pre-fill supplier/details/description |
| PR → PO (ปุ่มสร้างใบสั่งซื้อ) | **done** | AppBar button + pre-fill supplier/details/description |
| RFQ → PO (ปุ่มสร้างใบสั่งซื้อ) | **done** | AppBar button + pre-fill selected vendor/details |
| ปุ่มแสดงเฉพาะเมื่ออนุมัติแล้ว | **done** | `_poApprovalStatus?.isCompleted == true` |
| Pre-fill docrefno (อ้างอิงเลขที่ต้นทาง) | **done** | sourceDocNo → docrefno field |
| Copy Document (คัดลอก PR) | **done** | AppBar copy icon + confirm + deep-copy details |

---

## Backend Data Pipeline Status

| Stage | PR (transflag=21) | RFQ (transflag=22) | PO (transflag=6) |
|-------|-------------------|---------------------|-------------------|
| MongoDB Save | **done** | **done** | **done** |
| Kafka Publish | **done** | **done** | **done** |
| PostgreSQL Sync | **done** | **done** | **done** |
| GoAPI Kafka Consumer | **done** | **done** | **done** |
| ClickHouse Sync | **done** | **done** | **done** |
| Approval System | **done** (backend) | **done** (backend) | **done** |

## Backend Model Fields (PR)

```
PurchaseRequisition struct:
  PartitionIdentity     (inline)
  Transaction           (inline)
  RequesterCode         string          // ผู้ขอซื้อ
  RequesterName         string          // ชื่อผู้ขอ
  DepartmentCode        string          // แผนก
  DepartmentNames       *[]NameX        // ชื่อแผนก (multi-lang)
  Purpose               string          // วัตถุประสงค์
  BudgetCode            string          // รหัสงบ
  BudgetAmount          float64         // วงเงินงบ
  Urgency               int8            // 1=ปกติ 2=เร่ง 3=วิกฤต
  RequestedDeliveryDate string          // วันที่ต้องการรับ
  RefPODocNo            string          // อ้างอิง PO
  RefPOGuidFixed        string
  RefRFQDocNo           string          // อ้างอิง RFQ
  RefRFQGuidFixed       string
  ConversionStatus      string          // none/partial/full
  CreditDays            int             // เครดิตเทอม (วัน)
  PaymentCondition      string          // เงื่อนไขชำระเงิน
  JobCode               string          // รหัสงาน/โครงการ
  JobNames              *[]NameX        // ชื่องาน (multi-lang)
  ShipToAddress         string          // ที่อยู่จัดส่ง
  ShipToName            string          // ชื่อผู้รับ
  DocRefNo              string          // เอกสารอ้างอิงภายนอก
  Attachments           []string        // ไฟล์แนบ (URIs)
  // New fields (extraFields — save via frontend)
  EstimatedUnitCost     string          // ประมาณการราคาต่อหน่วย
  EstimatedFreight      string          // ประมาณการค่าขนส่ง
  EstimatedDuty         string          // ประมาณการอากร
  EstimatedLandedCost   string          // ประมาณการรวม
  PreferredVendor       string          // ร้านที่ต้องการ
  AlternativeVendor     string          // ร้านสำรอง
  ReasonPreferred       string          // เหตุผลที่เลือก
  ApprovalDeadline      string          // กำหนดวันอนุมัติ
```

## Competitive Landscape

| Software | Score | ราคา/เดือน | จุดเด่น |
|----------|-------|-----------|---------|
| SAP Business One | 38/41 | 100,000+ | ครบที่สุด + custom fields |
| Prosoft WINSpeed | 37/41 | 20,000+ | ครบวงจรจัดซื้อไทย + purchase history |
| Dynamics 365 | 36/41 | 50,000+ | approval mobile + spending limit |
| **BC Account (ปัจจุบัน)** | **43/47** | **790** | **AI LINE + approval LINE + multi-currency + vendor auto-fill + budget check + internal note + tolerance + expiry** |
| AccCloud | 34/41 | 15,000+ | workflow ตามไซต์ + auto PR |
| Formula | 33/41 | 10,000+ | budget check + cost center |
| Nested | 30/41 | 8,000+ | BOQ อ้างอิง + approval |
| Treesoft | 26/41 | 5,000+ | - |
| PEAK Account | 20/41 | 1,500+ | PR/IR/GR ในภาพรวมผู้ติดต่อ |
| Express/SMEMOVE | 8/41 | 500+ | ไม่มี PR โดยเฉพาะ |

## PR vs PO Feature Comparison

| Feature | PO | PR | หมายเหตุ |
|---------|----|----|---------|
| Approval workflow | ✅ | ✅ | Shared code |
| Discount (doc-level) | ✅ | ✅ | Same controllers |
| VAT selection | ✅ | ✅ | VAT type chips + rate field |
| Credit terms | ✅ | ✅ | Same field |
| Department/Job | ❌ | ✅ | PR-specific |
| Multi-currency | ✅ | ✅ | Shared — DocumentTotalWidget + CurrencyApiService |
| Warehouse/Location columns | ✅ | ❌ | PR ไม่ต้องการ (internal request) |
| Reorder items | ❌ | ✅ | PR-specific |
| Purchase History | ❌ | ✅ | PR-specific — ดูประวัติซื้อต่อรายการ |
| Copy Document | ❌ | ✅ | PR-specific — คัดลอกเอกสาร |
| Serial/Lot | ❌ | ✅ | PR-specific — optional columns |
| Estimated Landed Cost | ❌ | ✅ | PR-specific — ประมาณการต้นทุน |
| Vendor Preferences | ❌ | ✅ | PR-specific — ร้านที่ต้องการ/สำรอง |
| Approval Deadline | ❌ | ✅ | PR-specific — กำหนดวันอนุมัติ |
| Tracking (conversion) | ❌ | ✅ | PR-specific — สถานะแปลง + ref docs |
