---
name: procurement
description: >
  BC Account Procurement Module — PR (ใบขอซื้อ), RFQ (สืบราคา), PO (ใบสั่งซื้อ) feature tracker
  and competitive analysis. Use when: (1) adding/modifying PR/RFQ/PO features, (2) checking
  feature status or competitive gaps, (3) planning procurement roadmap, (4) after ANY code
  change to PR/RFQ/PO screens or backend. Triggers on: "procurement", "purchase requisition",
  "PR feature", "PO feature", "RFQ", "ใบขอซื้อ", "สืบราคา", "ใบสั่งซื้อ", "ฝั่งซื้อ",
  "competitive analysis", "feature comparison", "คู่แข่ง".
---

# Procurement Module — Feature Tracker & Competitive Analysis

## Mandatory Update Rule

**ทุกครั้งที่แก้ code PR/RFQ/PO → ต้อง update skill นี้ทันที:**
1. แก้ feature ไหน → update status ใน `references/feature-matrix.md`
2. เพิ่ม feature ใหม่ → เพิ่มแถวใน feature matrix
3. แก้ backend model → update `references/feature-matrix.md` → Backend section
4. แก้ frontend UI → update status จาก pending → done
5. **ห้ามจบ session โดยไม่ update** — ถ้ามีการแก้ procurement code

## Quick Reference

### Document Flow
```
PR (ใบขอซื้อ) → [อนุมัติ] → RFQ (สืบราคา) → [อนุมัติ] → PO (ใบสั่งซื้อ)
     ↑                              ↑                         ↑
  แผนกขอซื้อ              vendor 3+ ราย            เลือก vendor ที่ดีที่สุด
  transflag=21            transflag=22              transflag=6
```

### Key Files

| Area | Path |
|------|------|
| **PR Backend Model** | `backend/internal/transaction/purchaserequisition/models/purchaserequisition.go` |
| **PR Frontend Edit** | `frontend/bcaiaccount/lib/screens/purchaserequisition/purchaserequisition_edit_screen.dart` |
| **PR Header Widget** | `frontend/bcaiaccount/lib/screens/purchaserequisition/components/pr_header_widget.dart` |
| **PR Form Controller** | `frontend/bcaiaccount/lib/screens/purchaserequisition/utils/pr_form_controller.dart` |
| **PR Repository** | `frontend/bcaiaccount/lib/repositories/trans_repository.dart` → `savePurchaseRequisition()` |
| **PR List Screen** | `frontend/bcaiaccount/lib/screens/purchaserequisition/purchaserequisition_list_screen.dart` |
| **PR BLoC Events** | `frontend/bcaiaccount/lib/bloc/trans/trans_event.dart` → `TransSave.extraFields` |
| **PR BLoC** | `frontend/bcaiaccount/lib/bloc/trans/trans_bloc.dart` |
| **PO Frontend Edit** | `frontend/bcaiaccount/lib/screens/purchaseorder/purchaseorder_edit_screen.dart` |
| **RFQ Backend Model** | `backend/internal/transaction/rfq/models/rfq.go` |
| **Kafka Consumer PR** | `backend/internal/goapi/handlers/kafka/purchase_requisition.go` |
| **Kafka Consumer RFQ** | `backend/internal/goapi/handlers/kafka/rfq.go` |

### Architecture Patterns

#### ExtraFields Pattern (สำคัญ)
PR/RFQ มี fields ที่ไม่อยู่ใน `TransactionModel` (เช่น departmentcode, jobcode, shiptoaddress)
ใช้ pattern: **extraFields Map ผ่าน BLoC Event → Repository**

```
PRFormController → _buildPRExtraFields() → TransSave(extraFields: {...})
  → TransBloc → savePurchaseRequisition(trans, extraFields: extraFields)
    → Repository: data.addAll(extraFields) ก่อน POST
```

Fields ที่มีใน TransactionModel (creditdays, docrefno, description) → set ตรงบน screenData ได้เลย
Fields ที่ไม่มี (departmentcode, jobcode, shiptoaddress) → ต้องใช้ extraFields pattern

#### Transient Fields Pattern (โหลดข้อมูลกลับ)
PR-specific fields ที่ GoAPI ส่งกลับมา ต้อง map เข้า transient fields บน TransactionModel:
```
TransactionModel.prDepartmentCode  ← data['departmentcode']
TransactionModel.prJobCode         ← data['jobcode']
TransactionModel.prShipToAddress   ← data['shiptoaddress']
TransactionModel.prUrgency         ← data['urgency']
TransactionModel.description       ← data['purpose']
TransactionModel.creditdays        ← data['creditdays']
TransactionModel.docrefno          ← data['docsrefno']
TransactionModel.prEstimatedUnitCost ← data['estimatedunitcost']
TransactionModel.prEstimatedFreight  ← data['estimatedfreight']
TransactionModel.prEstimatedDuty     ← data['estimatedduty']
TransactionModel.prPreferredVendor   ← data['preferredvendor']  // JSON array
TransactionModel.prApprovalDeadline  ← data['approvaldeadline']
TransactionModel.prInternalNote      ← data['internalnote']
TransactionModel.prPurchasingGroup   ← data['purchasinggroup']
TransactionModel.prOverDeliveryTolerance  ← data['overdeliverytolerance']
TransactionModel.prUnderDeliveryTolerance ← data['underdeliverytolerance']
// ExpiryDate → detail item extrajson['expiry_date'] (ไม่ใช่ header field)
TransactionModel.prConversionStatus  ← data['conversionstatus']
TransactionModel.prRefRfqDocNo       ← data['refrfqdocno']
TransactionModel.prRefPoDocNo        ← data['refpodocno']
```
Transient fields ไม่ถูก serialize (ไม่อยู่ใน .g.dart) → ปลอดภัยสำหรับ transaction types อื่น

#### Vendor Preferences Pattern (Dynamic List)
ร้านที่ต้องการ เก็บเป็น JSON array ใน `prPreferredVendor`:
```json
[{"vendor":"ร้านABC","reason":"ราคาดี"}, {"vendor":"ร้านXYZ","reason":"ส่งเร็ว"}]
```
- UI: dynamic list (เพิ่ม/ลบได้) ใน PRHeaderWidget `_vendorPreferences` state
- Auto-sync ไป `screenData.prPreferredVendor` ทุกครั้งที่พิมพ์ผ่าน `_autoSyncVendors()`
- Save: ผ่าน `_buildPRExtraFields()` → `extra['preferredvendor'] = vpJson`

#### Multi-Currency Pattern
PR รองรับหลายสกุลเงิน เหมือน PO:
- State: `_currencies`, `_selectedDocCurrencyCode`, `_selectedExchangeRate`, `_baseCurrency`
- Load: `_loadCurrencies()` ใน initState → CurrencyApiService
- UI: Currency section ใน PRHeaderWidget (dropdown + exchange rate)
- Save: `_saveCurrencyData()` ก่อน save → screenData.docCurrency, exchangerate
- Calculation: priceDoc/sumAmountDoc ใน detail items

#### Department Data Source (สำคัญ — Pitfall)
**Department ดึงจาก `global.companyBranchSelectData.departments` ไม่ใช่จาก API แยก**
- `/organization/department/list` API จะ return empty list (count=0)
- Department เป็น sub-array ของ CompanyBranchModel → เหมือน department_screen.dart
- เช่นเดียวกัน: Job Project, Cost Center ก็ดึงจาก branch data

#### Dialog Display Format
ทุก dialog เลือกข้อมูล (department, job, costCenter) → แสดง "รหัส / ชื่อ":
```dart
fc.departmentNameController.text = name.isNotEmpty ? '${item.code} / $name' : item.code;
```

#### Vendor Price Auto-fill Pattern
เมื่อเพิ่มสินค้าลง PR → ดึง lastPrice จาก PurchaseHistoryRepository → auto-fill ราคา:
- Hook: `DocumentProductListWidget.onDetailAdded` callback → `_autoFillLastPrice(detail)`
- เงื่อนไข: barcode ไม่ว่าง + detail.price == 0 (ยังไม่มีราคา)
- ใช้ประวัติ 12 เดือนย้อนหลัง
- ไม่ block ถ้าไม่มี history

#### Budget Check Pattern
Budget fields (budgetcode/budgetamount) อยู่ใน PR header:
- UI: field งบประมาณ + warning banner สีแดง เมื่อ totalAmount > budgetAmount
- Warning แสดง % ที่เกิน + ตัวเลขเปรียบเทียบ
- **ไม่ block save** — แค่เตือน (approver ตัดสินใจ)
- Save: ผ่าน extraFields pattern (budgetcode + budgetamount)

#### Summary Widget (Shared with PO)
PR ใช้ `DocumentTotalWidget(screenData: screenData)` เหมือน PO
- แสดงยอดครบ: totalvalue, discount, VAT, after VAT, grand total
- รองรับ multi-currency (doc currency vs base currency)
- อยู่ใน `SingleChildScrollView` เลื่อนไปด้วยกันกับ header (ไม่ fixed)

#### PR Layout (2 Tabs)
```
Tab 1: เอกสาร — header + summary (scroll ด้วยกัน)
  1. ข้อมูลเอกสาร — เลขที่ + วันที่ + ผู้ขอ + แผนก + ความเร่งด่วน + วัตถุประสงค์
  2. ผู้จำหน่าย — รหัส + ชื่อเจ้าหนี้
  3. สกุลเงิน — base/doc currency + exchange rate (แสดงเมื่อมีหลายสกุล)
  4. เงื่อนไข — เครดิต + เอกสารอ้างอิง + ประเภทภาษี + อัตราภาษี
  5. โครงการ — โครงการ + งาน + ศูนย์ต้นทุน + ที่อยู่จัดส่ง
  6. ประมาณการต้นทุน — ราคาต่อหน่วย + ค่าขนส่ง + อากร
  7. ร้านที่ต้องการ — dynamic list (เพิ่มได้เรื่อยๆ)
  8. อนุมัติ — กำหนดวันอนุมัติ + tracking
  9. ส่วนลดท้ายบิล — discount word
  10. ยอดรวม — DocumentTotalWidget (shared with PO)
  11. ประมาณการต้นทุนรวม — estimated unit + freight + duty
  12. Approval Status Badge
Tab 2: รายการสินค้า — product list + detail items
```

### Competitive Position

BC Account อยู่ **อันดับ 2/10** (43/47 = 91%) ด้านฟีเจอร์ PR — แซง Dynamics 365 + ใกล้ Prosoft
- ราคา 790 บาท/เดือน vs Prosoft WINSpeed หลายหมื่น/เดือน

### Killer Features (ไม่มีใครมี)
1. **AI Chatbot สั่ง PR ผ่าน LINE** — ใช้ MCP ที่มีอยู่แล้ว (SAP ล้านก็ยังไม่มี)
2. **Approval ผ่าน LINE OA** — คนไทยใช้ LINE ทุกวัน (คู่แข่งทำได้แค่ email)

## Feature Status

ดู `references/feature-matrix.md` สำหรับ:
- Feature list ครบทุกข้อ
- Status (done/in-progress/pending/planned)
- Priority (ระดับ 1-4)
- คู่แข่งที่มี feature นั้น

## Workflow: เพิ่ม Feature ใหม่

1. ตรวจ `references/feature-matrix.md` → feature อยู่ระดับไหน
2. ตรวจ backend model → มี field รองรับหรือยัง
3. ตรวจ frontend TransactionModel → map field ยังไง
4. เขียน code (backend + frontend)
5. **update `references/feature-matrix.md`** → เปลี่ยน status เป็น done
6. update `references/feature-matrix.md` → เพิ่ม notes ถ้ามี pitfall
