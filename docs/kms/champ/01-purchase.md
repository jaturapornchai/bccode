# Champ -> BC : ระบบซื้อ (PO) — พิมพ์เขียวหน้าจอ เอกสาร กระบวนการ และรายงาน

> ที่มา: ลุงจืดสั่ง 2026-10-07 ให้ออกแบบ BC ใหม่โดยยึด Champ เป็นหลัก ให้เหมือนกันแทบทุกอย่าง (BRIEF: `scratch/champ-atlas/BRIEF.md`) ; ไฟล์นี้ครอบคลุมโมดูล PO ของ Champ ทั้ง 96 รายการเมนู (menu id 5797-5814, 5048, 5049 = จอ/กระบวนการ 19 รายการ + รายงาน 77 รายการ) ตามแอตลาส `scratch/champ-atlas/atlas-01-PO.md` และรายงานจัดซื้อ 33 รายการในศูนย์รายงาน (Report Center) จาก `atlas-10-รายงาน.md`
> สถานะ BC (ตรวจ 2026-10-07): เมนูซื้อใน `frontend/src/lib/menu-data.ts:176-235` ; จอเอกสารซื้อเป็นเส้นทาง ERP ที่ถูกนับเป็น "รอพัฒนา" (`frontend/src/lib/menu-screen-status.ts:46-50`) ; ฝั่ง Go ไม่มีตาราง/endpoint ใบสั่งซื้อ รับสินค้า ตั้งหนี้ หรือซื้อสินค้าที่บันทึกเอกสารได้. ขอบเขตปัจจุบันของลุงจืดคือ GL อย่างเดียว — พิมพ์เขียวนี้คือสเปกสำหรับตอนที่ระบบซื้อถูกสร้างจริง
> ข้อควรระวังการอ้างอิง: ข้อความไทยของ Champ ตอนรันมาจากตาราง SQL `myresource` ที่ไม่อยู่ใน repo — ป้ายในเอกสารนี้จึงมาจาก 3 ทาง (ดู §0.2) ; ห้ามถือว่าตรงกับหน้าจอจริงทุกตัวอักษรโดยไม่ตรวจกับ Champ ที่รันอยู่
> ชื่อไฟล์ย่อ: `champ\champ\` = `D:\project-champ\champ\champ\` ; `rep-po\` = `D:\project-champ\champ-report\BC5REP_PO\` ; `lib\` = `D:\project-champ\champ-lib\` (BC5PROCESS = `bcprocess.cpp`, BC5FRMWRK = พื้นฟอร์ม) ; เอกสารเจ้าหนี้ที่ต่อจากการซื้อ (ใบรับวางบิล จ่ายชำระ ตัดหนี้สูญ) อยู่ใน `03-ap.md`

## สารบัญ

- §0 กรอบร่วม: เมนู 96 รายการ · วิธีอ่าน · โมเดลเอกสารซื้อ 3 แบบ (ซื้อตรง / รับสินค้า GR / ตั้งหนี้ IR) · การลงบัญชี · การส่งต่อสถานะระหว่างเอกสาร
- §1 เอกสารและกระบวนการ 19 รายการ
  - เอกสารหลักซื้อ-รับ: 1.1 [5810] ใบสั่งซื้อ · 1.2 [5808] ยกเลิกใบสั่งซื้อ · 1.3 [5798] รับสินค้า (GR) · 1.4 [5799] ตั้งหนี้ (IR) · 1.5 [5812] ซื้อสินค้า/บริการ
  - ส่งคืน ลดหนี้ เพิ่มหนี้: 1.6 [5800] ส่งคืนสินค้า (GR) · 1.7 [5813] ส่งคืนสินค้า/ลดหนี้ · 1.8 [5801] ลดหนี้ · 1.9 [5814] เพิ่มหนี้/เพิ่มสินค้า
  - ก่อนสั่งซื้อ: 1.10 [5797] เสนอซื้อ · 1.11 [5805] อนุมัติเสนอซื้อ · 1.12 [5802] สืบราคา · 1.13 [5809] ตารางราคารวม · 1.14 [5803] ประมวลใบสั่งซื้ออัตโนมัติ
  - เงินจ่ายล่วงหน้า: 1.15 [5811] จ่ายเงินมัดจำ · 1.16 [5806] จ่ายเงินล่วงหน้า · 1.17 [5807] รับคืนเงินจ่ายล่วงหน้า
  - เครื่องมือ: 1.18 [5048] ต้นทุนแฝง (Weight cost) · 1.19 [5049] ปรับปรุงใบรับสินค้า
- §2 รายงานจัดซื้อ — 2.1 รายงานในเมนูซื้อ 77 รายการ · 2.2 รายงานในศูนย์รายงาน 33 รายการ · 2.3 หมายเหตุเฉพาะรายงาน
- §3 BC เทียบ Champ และสิ่งที่ต้องปรับ

## 0. กรอบร่วม

### 0.1 โครงเมนู PO ของ Champ (96 รายการ)

ที่มา `champ\menuconfigxml\menuconfig.xml:2-147` (ไฟล์ UTF-16LE ; ที่ใช้ตรวจเป็นสำเนาถอดรหัสแล้ว). ลำดับด้านล่างคือลำดับบนแถบเมนู :

| ส่วน | menu id | ชื่อเมนู (ตาม `menuconfig.xml`) | ประเภท | หัวข้อ |
|---|---|---|---|---|
| ใบสั่งซื้อ | 5810 | บันทึกใบสั่งซื้อสินค้า | จอบันทึกเอกสาร | §1.1 |
| | 5808 | ยกเลิกใบสั่งซื้อสินค้า | จอดึงรายการค้าง-ติ๊กเลือก | §1.2 |
| รับ/ตั้งหนี้/คืน (คั่นด้วยเส้นแบ่ง) | 5798 | บันทึกรับสินค้าจากการซื้อ | จอบันทึกเอกสาร (GR) | §1.3 |
| | 5799 | บันทึกตั้งหนี้จากการซื้อ | จอบันทึกเอกสาร (IR) | §1.4 |
| | 5800 | บันทึกส่งคืนสินค้า | จอบันทึกเอกสาร | §1.6 |
| | 5801 | บันทึกลดหนี้ | จอบันทึกเอกสาร | §1.8 |
| ซื้อตรง/คืน/เพิ่มหนี้ (คั่นด้วยเส้นแบ่ง) | 5812 | บันทึกซื้อสินค้า,บริการ | จอบันทึกเอกสาร | §1.5 |
| | 5813 | บันทึกส่งคืนสินค้า/ลดหนี้ | จอบันทึกเอกสาร | §1.7 |
| | 5814 | บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้) | จอบันทึกเอกสาร | §1.9 |
| ก่อนสั่งซื้อ (คั่นด้วยเส้นแบ่ง) | 5797 | บันทึกใบเสนอซื้อสินค้า | จอบันทึกเอกสาร | §1.10 |
| | 5805 | บันทึกอนุมัติใบเสนอซื้อสินค้า | จอบันทึกเอกสาร | §1.11 |
| ระบบสืบราคาซื้อสินค้า (เมนูย่อย) | 5802 | บันทึกใบสืบราคาสินค้ารวม | จอบันทึกเอกสาร | §1.12 |
| | 5809 | บันทึกใบตารางราคารวม | จอบันทึกเอกสาร | §1.13 |
| | 5803 | ประมวลใบผลสั่งซื้อสินค้าอัตโนมัติ | กระบวนการ (process) | §1.14 |
| เงินจ่ายล่วงหน้า/เครื่องมือ | 5811 | บันทึกใบจ่ายเงินมัดจำ | จอบันทึกเอกสาร | §1.15 |
| | 5806 | บันทึกใบจ่ายเงินล่วงหน้า | จอบันทึกเอกสาร | §1.16 |
| | 5807 | บันทึกรับคืนเงินจ่ายล่วงหน้า | จอบันทึกเอกสาร | §1.17 |
| | 5048 | บันทึกต้นทุนแฝง (Weight cost) | จอบันทึกเอกสาร | §1.18 |
| | 5049 | ปรับปรุงใบรับสินค้า | เครื่องมือซ่อมสถานะ (ไม่ใช่เอกสาร) | §1.19 |
| รายงานรายวัน (popup ซ่อน `disable="yes"` `menuconfig.xml:27-28`) | 5815, 5837 | รายงานการบันทึกใบขออนุมัติซื้อพร้อมรายการย่อย ; รายงานใบขออนุมัติซื้อที่ทำสั่งซื้อหมดแล้ว | รายงาน | §2.1 |
| | 5817 | รายงานใบจ่ายเงินมัดจำที่คงค้าง | รายงาน | §2.1 |
| | 5841, 5828, 5838 | รายงานการซื้อสินค้าพร้อมค่าใช้จ่ายอื่นๆ(ต้นทุนแฝง) ; รายงานใบซื้อสินค้าที่ถึงกำหนดจ่ายเงิน ; รายงานเปรียบเทียบราคาซื้อ | รายงาน | §2.1 |
| | 5832, 5833 | รายงานการซื้อสินค้าแยกตามประเภทการซื้อ ; รายงานการซื้อสินค้าแจกแจงตามประเภทการซื้อ | รายงาน | §2.1 |
| | 5823, 5824 | รายงานใบส่งคืนที่ตัดยอดทันที ; รายงานใบส่งคืนที่รับคืนเงิน | รายงาน | §2.1 |
| | 5825, 5826, 5733, 5785 | รายงานการบันทึกใบเพิ่มเจ้าหนี้พร้อมรายการย่อย ; Vender Evalution Report (**ไม่มี handler** — ดู §2.3) ; รายงานการตัดใบซื้อสินค้า,บริการ ; รายงานบันทึกใบเสนอซื้อสินค้า | รายงาน | §2.1 |
| รายงานเชิงวิเคราะห์ (popup ซ่อน `menuconfig.xml:48`) 14 กลุ่ม แต่ละกลุ่มแตกตามวันที่/เจ้าหนี้/แผนก/กลุ่มสินค้า/โครงการ ฯลฯ | 5698, 5729, 5737, 5740 · 5742-5745 | ขออนุมัติซื้อ แจกแจง 4 · สรุป 4 | รายงาน | §2.1 |
| | 5746, 5751, 5752, 5738 · 5747-5750, 5736 | สั่งซื้อ แจกแจง 4 · สรุป 5 | รายงาน | §2.1 |
| | 5701-5703 · 5705-5708 | ใบรับสินค้า แจกแจง 3 · สรุป 4 | รายงาน | §2.1 |
| | 5711-5714 · 5716-5718 | ใบตั้งหนี้จากการซื้อ แจกแจง 4 · สรุป 3 | รายงาน | §2.1 |
| | 5699, 5720-5725 · 5726-5728, 5730, 5731, 5734, 5735 | การซื้อสินค้า แจกแจง 7 · สรุป 7 | รายงาน | §2.1 |
| | 5756-5760 · 5761, 5762, 5768 | การส่งคืนสินค้า/ลดหนี้ แจกแจง 5 · สรุป 3 | รายงาน | §2.1 |
| | 5772-5776, 5778, 5779 · 5782, 5783, 5788 | ใบเพิ่มสินค้า(เจ้าหนี้) แจกแจง 7 · สรุป 3 | รายงาน | §2.1 |

รวม: 19 (จอ/กระบวนการ) + 14 (รายงานรายวัน) + 63 (รายงานเชิงวิเคราะห์) = 96. ทั้งสอง popup รายงานเป็น `disable="yes"` (ซ่อนจากแถบเมนู ดู `XMLMenuConfig.cpp:153,179`) เข้าถึงผ่านศูนย์รายงาน (Report Center `menuconfig.xml:617-619` รายการ 5007) ซึ่งมี popup "จัดซื้อ" ของตัวเองอีก 33 รายการ (รายงานรายวัน 25 + รายงานวิเคราะห์ 8, `menuconfig.xml:620-670`) — รายการเหล่านี้ใช้ menu id คนละชุดกับด้านบน (เช่น 5831 ใบสั่งซื้อ, 5741 สั่งซื้อวิเคราะห์) และนับแยกไว้ใน §2.2 ไม่รวมในตัวเลข 96. ไฟล์ตั้งค่าศูนย์รายงาน `reportcenterconfig.xml` ไม่อยู่ใน repo — **ยังไม่verify** ว่าผู้ใช้เห็นรายงานกลุ่มไหนบ้างตอนรันจริง.

### 0.2 วิธีอ่านและข้อจำกัดของหลักฐาน

- **หลักฐาน**: ใช้รูป `ชื่อไฟล์:บรรทัด` ของไฟล์ต้นฉบับ Champ (เลขบรรทัดของสำเนาถอดรหัสตรงกับต้นฉบับ) ; ข้อความที่ไม่มีหลักฐานเขียนว่า "ยังไม่ verify".
- **ป้ายไทยบนจอ** มี 3 ที่มา : (ก) ข้อความที่พิมพ์ใน `BC5Account.rc` / `BCREPPO.rc` (บางป้ายตกค้างจากแม่แบบฟอร์มลูกหนี้/ขาย — ระบุกำกับเมื่อพบ เช่นช่องรหัสเจ้าหนี้พิมพ์ว่า "รหัสลูกค้า"), (ข) key `GetMyRes("KEY")` ที่แปลจากชื่อ key — ขึ้นต้น `~` = อนุมานจากชื่อ key ไม่ใช่ข้อความจริง, (ค) key ที่แปลไม่ได้ เขียน ⟨KEY⟩ ; `«TOKEN»` = ป้ายที่ประกอบจากตัวแปรตอนรัน (เช่น «SALES_NO» = เลขที่เอกสารตามประเภทเอกสาร).
- **ป้าย rc/atlas ของหน้าต่างเงื่อนไขรายงานจัดซื้อหลายตัวเป็นข้อความเก่าของแม่แบบอื่น** ; ตารางรายงานใน §2 จึงสร้างตัวกรองจาก key ที่ใช้จริงในไฟล์ `*Dlg.cpp` แล้วอนุมานป้ายไทยจากชื่อ key — ทุกป้ายในคอลัมน์ "ตัวกรอง" ที่ขึ้นต้น `~` ยังไม่ verify กับ `myresource`.
- **ตัวแปรของลูกค้าเฉพาะราย** (`TOHOME_PROJECT`, `_AFOSS_PROJ`, `EXPRESS_PROJ`, `CHM_PROJ`, `GL_PROJ`, ช่อง SAP `oomdts`) เป็นโค้ดทำเพิ่มให้โครงการเฉพาะ ไม่อยู่ในพิมพ์เขียวนี้ ยกเว้นระบุไว้ชัดเจนว่าข้ามได้. ไฟล์ฟอร์ม `Chm*` (เช่น `ChmFrmPurchaseOrder`, `ChmFrmGoodsReceive`, `ChmFrmGoodsPurchase`) คือฟอร์มที่เมนูเรียกจริงตาม `MainFrm.cpp` ; ไฟล์ `PO*` รุ่นเก่าที่มีชื่อซ้ำ (เช่น `POFrmGoodsReceive.cpp`, `POFrmPurchase.cpp`) ไม่ถูกเรียกจากเมนูนี้.
- **รูปแบบมาตรฐานของฟอร์มเอกสาร** (แถบเครื่องมือ New/Save/Del/Find/Prev/Next/Close/Info/More, แท็บ «MST_HEADER» = ข้อมูลหลัก และ «MST_OTHER» = ข้อมูลอื่น, กริดกรอกรายการ `CBInputListWnd`, ปุ่มลัด F2/F3/F5/F6/F7/F8/F11, การลบ/ยกเลิก/Confirm, วงจรบันทึกและป้ายสถานะ) อธิบายครั้งเดียวใน `00-common-ui.md` (§3 แถบปุ่ม, §4 ปุ่มลัด, §6 กริดรายการ, §10 วงจรบันทึก/สถานะ/ยืนยัน/ยกเลิก, §12 พิมพ์และยอดรวมท้ายหน้ารายงาน) และกลไกร่วมของเอกสารที่ลงสต็อกใน `07-inventory.md` §0.1 (สถานะ `TransState` ของรายการเก่า, ปิดงวด `IsPeriodClose`, `OnStockReprocess`) — จอซื้อไม่อธิบายซ้ำ ระบุเฉพาะส่วนต่าง.
- **ตัวเลข/เงิน (Rule 20)** : ใน DDL ของ Champ คอลัมน์เงินเป็น `MONEY` หรือ `DECIMAL` แต่โค้ดคำนวณด้วย `double` และเขียน SQL ด้วย `%lf` (เช่น `mytools.cpp:2164`, `bcprocess.cpp:3070`) ; BC ต้องใช้ `numeric(p,s)` + decimal ห้ามลอก `double`. หน่วยนับแบบ `UnitType=1` (หน่วยแพ็ก) แปลงเป็นหน่วยฐานด้วย `OnGetQtyPacking` ก่อนรวมจำนวนเสมอ (`mytools.cpp:1711-1712`).
- **แก้ไขเอกสารเดิม = ถอนผลเก่าแล้วลงผลใหม่** : เกือบทุกจอเขียนแบบ "ย้อน (nMark = -1) ผลของเลขที่เดิม → เขียนหัว/รายการใหม่ → ลงผลใหม่ (nMark = +1)" ไม่ใช่การแก้ทับแถวเดียว ; **ไม่มี transaction ครอบตลอดทาง** ในหลายจอ (ระบุที่ข้อสังเกตของแต่ละหัวข้อ) — BC ต้องทำทั้งหมดใน `sql.Tx` เดียว.

### 0.3 โมเดลเอกสารซื้อของ Champ (3 เอกสาร + ห่วงโซ่ก่อนสั่งซื้อ)

**ห่วงโซ่ก่อนสั่งซื้อ** (ทุกขั้นเป็นทางเลือก ข้ามได้ ; ไม่ลงบัญชี ไม่กระทบสต็อก ไม่กระทบเจ้าหนี้) :

`ใบเสนอซื้อ 5797` → `ใบอนุมัติเสนอซื้อ 5805` → `ใบสืบราคา 5802` → `ใบตารางราคารวม 5809` → `ประมวลใบสั่งซื้อ 5803` (สร้างใบสั่งซื้อทีละหลายใบ จากใบอนุมัติ หรือจากตารางราคา) → `ใบสั่งซื้อ 5810` (ออกเองโดยตรงได้ หรือจากใบสั่งขาย — ดู §0.5 ชนิด 23)

| ขั้น | ตาราง (หัว / รายการ) | เลขที่เอกสารต้นทางที่อ้าง | สถานะที่ใช้ |
|---|---|---|---|
| ใบเสนอซื้อ 5797 | `BCStkRequest` / `BCStkRequestSub` | — | `BillStatus` (1 = อนุมัติหมดแล้ว), `IsConfirm`, `RemainQty` ต่อบรรทัด |
| ใบอนุมัติเสนอซื้อ 5805 | `BCReqConfirm` / `BCReqConfirmSub` | `StkRequestNo` + `RefLineNumber` | `BillStatus` (1 = ทำใบสั่งซื้อหมดแล้ว), `RemainQty` ต่อบรรทัด |
| ใบสืบราคา 5802 | `BCEnquirePrice` / `BCEnquirePriceSub` | `ReqConfirmNo` | `BillStatus` |
| ใบตารางราคารวม 5809 | `BCBuyPriceCondit` / `BCByPriceCondSub` | `RefNo` (= เลขที่ใบสืบราคา) | `BillStatus` |
| ใบสั่งซื้อ 5810 | `BCPurchaseOrder` / `BCPurchaseOrderSub` | `ConfirmNo` + `RefLineNumber` (ใบอนุมัติ) ; `SORefNo` + `SOLineNumber` (ใบสั่งขาย) | `BillStatus`, `PoStatus` (0 ยังไม่รับ · 2 รับครบ · 3 รับบางส่วน · 4 ยกเลิก), `IsConfirm`, `IsCancel` |
| ยกเลิกใบสั่งซื้อ 5808 | `BCPOVoid` / `BCPOVoidSub` (`PONo`, `VoidQty`) | เลขที่ใบสั่งซื้อ | ปิดยอดค้างส่งของรายบรรทัด |

**เอกสารซื้อ 3 แบบใช้ตารางเดียวกัน แยกด้วย `GRIRBillStatus`** (ตารางหัว `BCAPInvoice` ; ตัวเลขนี้คือหัวใจของโมเดล BC ต้องตัดสินใจให้ชัดว่าจะแยกเป็น 3 ตารางหรือ 1 ตาราง + ชนิดเอกสาร) :

| เอกสาร | menu | `GRIRBillStatus` | ตารางรายการ | สต็อกเข้า | ตั้งหนี้เจ้าหนี้ (`BuyAPProcess`) | ภาษีซื้อ (`BCInputTax`) | ลง GL | กลุ่มรูปแบบ GL |
|---|---|---|---|---|---|---|---|---|
| ซื้อตรง (ในใบเดียวจบ: รับ+ตั้งหนี้+ภาษี+จ่ายสด) | 5812 | 0 | `BCAPInvoiceSub` | ใช่ | ใช่ (`BuyAPProcess` ; บิลสดบันทึกจ่ายทันทีด้วย F7 `OnSavePayment`) | ใช่ (Source 2) | ใช่ | 1, 2, 3, 4 |
| รับสินค้า GR (ยังไม่มีใบกำกับ) | 5798 | 1 | `BCAPInvoiceSub` (+ `GRRemainQty` ต่อบรรทัด) | ใช่ | **ไม่** | ไม่ | ใช่ | 29 |
| ตั้งหนี้ IR (อ้าง GR หรือซื้อก่อนรับของ) | 5799 | 2 | **`BCIRSUB`** (คอลัมน์ `GRRefNo`, `GRRefLinenum`, `CNQty`) | **ไม่** (สต็อกเข้าตอน GR) | ใช่ | ใช่ | ใช่ | 30 (รวม VAT) / 39 (ภาษีแยกนอก) |

ความหมายสถานะของ GR : `GRBillStatus` = 0 ยังไม่ถูกตั้งหนี้ครบ / 1 ถูกตั้งหนี้ครบแล้ว (`mytools.cpp:2319-2327`) — คำนวณจาก `Σ GRRemainQty` ของ GR (`GRRemainQty = Qty − Σ BCIRSUB.Qty` ที่ไม่ยกเลิกและอ้าง GR/บรรทัดนั้น `mytools.cpp:2313,2332`) ; เครื่องมือ 5049 (§1.19) แก้สถานะนี้เมื่อค้างผิด.

**เอกสารลดหนี้/คืน/เพิ่มหนี้** (ตารางหัวร่วมกัน แยกด้วย `BillType`) :

| เอกสาร | menu | ตารางหัว / รายการ / เอกสารต้นทาง | `BillType` ของหัว | กระทบสต็อก | กระทบยอดเจ้าหนี้ |
|---|---|---|---|---|---|
| ส่งคืนสินค้า (อ้าง GR เท่านั้น ยังไม่ตั้งหนี้) | 5800 | `BCStkRefund` / `BCStkRefundSub` / `BCInvStkRefund` | 2 ("GR Only" `POFrmStkReturn.cpp:1109`) | ใช่ (ตัดออก) | ไม่ (เอนจินกรอง `billtype<>2`) |
| ส่งคืนสินค้า/ลดหนี้ (อ้างใบซื้อ/IR/ยอดยกมา/เจ้าหนี้อื่น/มัดจำ) | 5813 | เดียวกัน | ไม่เขียนค่า (ใช้ค่าเริ่มต้นของคอลัมน์ — ยังไม่ verify DDL) | ใช่ (ประเภทคืนสินค้า) | ใช่ (`StkRefundAPProcess`) |
| ลดหนี้แบบกรอกยอด (อ้าง IR ใบเดียว ไม่มีรายการสินค้า) | 5801 | `BCStkRefund` / — (ไม่มี `BCStkRefundSub`) / `BCInvStkRefund` | 1 ("IR Only" `POFrmRefund.cpp:1108`) | ไม่ | ใช่ |
| เพิ่มหนี้ / เพิ่มสินค้า (เจ้าหนี้) | 5814 | `BCDebitNote2` / `BCDebitNoteSub2` / `BCInvDebitNote2` | — | ใช่ (ประเภท 0 เพิ่มจำนวนสินค้า) หรือปรับต้นทุน (ประเภท 1 ราคาผิด) | ใช่ (`DebitNote2APProcess`) |

เลข `BillType` ของ `BCInvStkRefund` (ตารางตัดใบลดหนี้เข้ากับเอกสารต้นทาง) คือ 0 = ใบซื้อ, 3 = ใบตั้งหนี้ IR, 1 = ยอดยกมา, 2 = เจ้าหนี้อื่นๆ, 8 = เงินมัดจำ (§1.7) — **คนละชุดกับ `BCPaymentSub`/`BCStatementSub` ของระบบเจ้าหนี้** (`03-ap.md` §0.3) ห้ามนำเลขข้ามชุด. ชื่อคลาส/ไฟล์ของใบเพิ่มหนี้ยังเป็น "CreditNote" (`pofrminvcreditnote.cpp`) แต่ทำงานเป็นใบเพิ่มหนี้ — BC ตั้งชื่อ route/ตารางตามความหมาย.

**เงินจ่ายล่วงหน้า 3 แบบ** (ไม่มีตารางรายการ ตารางเดียวต่อเอกสาร ; ทั้งหมดยังไม่ใช่เอกสารตั้งหนี้) : มัดจำ 5811 `BCAPDeposit` (มีใบกำกับ/ภาษีซื้อ/หัก ณ ที่จ่าย) · จ่ายล่วงหน้า 5806 `BCAPDepositSpecial` (ไม่มีภาษี) · รับคืนเงินจ่ายล่วงหน้า 5807 `BCAPReturnDepSpecial` (อ้างใบ 5806 ผ่าน `EarnestNo`) — ใช้หักตอนจ่ายชำระหนี้ (BillType 5 / 8 ใน `03-ap.md` §0.3) ; ต้นทุนแฝง 5048 `BCWeightCost`/`BCWeightCostSub` เพิ่มค่า `BCAPInvoiceSub.SumOfExpCost` ของบิลซื้อที่เลือก.

### 0.4 การลงบัญชี (GL) ของเอกสารซื้อ

การลงบัญชีทำที่ `CGLTLSProcess` (`gltlsprocess.cpp`) โดยเลือกรูปแบบจาก `BCGLFormat.FormatGroup` (`GetDefaultGLFormat` `mytools.cpp:32-50` : `SELECT CODE FROM BCGLFormat WHERE FormatGroup=%i ORDER BY IsDefault DESC`) แล้วเรียก `ProcessConditionGroupN` ตามค่า `m_nGLGroup` (`myespeciallyformview.cpp:6423-6460`). **เลข FormatGroup = ลำดับ (นับจาก 0) ของ key `GLFOR_*` ในรายการเลือกของจอ "รูปแบบการลงบัญชี"** (`glfrmpostformat.cpp:238-299`) ; ค่า `m_unaConditionGroup` ที่ผูกท้ายแต่ละ key คือกลุ่มเงื่อนไข ไม่ใช่เลข FormatGroup. บรรทัดบัญชีเดบิต/เครดิตมาจากรูปแบบ `BCGLFormat` ที่ผู้ใช้ตั้งเอง ไม่ฝังในโค้ด :

| เอกสาร | FormatGroup (key `GLFOR_`) | `m_nSource` / `m_nGLGroup` → เอนจิน | หลักฐาน |
|---|---|---|---|
| 5810 ใบสั่งซื้อ · 5808 ยกเลิก PO · 5797 · 5805 · 5802 · 5809 · 5803 · 5048 · 5049 | — **ไม่ลง GL** (5048 ไม่มีลงบัญชีอัตโนมัติ — ผลต่อต้นทุนสินค้าผ่าน `OnStockReprocess`) | — | §1.1, §1.2, §1.10-§1.14, §1.18, §1.19 |
| 5798 รับสินค้า GR | 29 `CRPURCHASE_GR` | Source 2 / group 1 `ProcessConditionGroup1` (`m_nDocStatus=1`) | §1.3 (`ChmFrmGoodsReceive.cpp` `GetGLFormat` `:2468-2471`) |
| 5799 ตั้งหนี้ IR | 30 `CRPURCHASE_IR` ; **39 `IR_EXCVAT`** เมื่อภาษีแยกนอก (`TaxType=0`) | Source 2 / group 1 (บริการ BillType 2/3 → group 2 ; ภาษีแยกนอก → group 39 → `ProcessConditionGroup1`, `myespeciallyformview.cpp:6453`) `m_nDocStatus=2` | §1.4 (`POFrmInvoiceReceive.cpp:3824-3833`) |
| 5812 ซื้อสินค้า/บริการ | 1 `CASH_PURCHASE` (สินค้าสด) · 2 `CREDIT_PURCHASE` (สินค้าเชื่อ) · 3 `CASHSERVICE_PURCHASE` · 4 `CREDITSERVICE_PURCHASE` | Source 2 / group 1 (สินค้า) หรือ 2 (บริการ ประเภท 2/3) `m_nDocStatus=0` | §1.5 (`ChmFrmGoodsPurchase.cpp:3265-3272`) |
| 5800 ส่งคืนสินค้า (GR) | 31 `CRPURCHASE_GOODSRET` | Source 3 / group **42** → `ProcessConditionGroup38` (`myespeciallyformview.cpp:6456`) | §1.6 (`POFrmStkReturn.cpp:2968`) |
| 5813 ส่งคืนสินค้า/ลดหนี้ | 5 `PURCHASERET_AP` (ประเภท 0) · 50 `PURCHASERET_WRONGP_AP` (ราคาผิด) · 52 `CNDEPOSIT_AP` (ลดหนี้เงินมัดจำ) | Source 3 / group 3 → `ProcessConditionGroup3` | §1.7 (`pofrmstkrefund.cpp:4595-4617`) |
| 5801 ลดหนี้ (กรอกยอด) | 32 `CRPURCHASE_CN` | Source 3 / group **43** → `ProcessConditionGroup39` (`myespeciallyformview.cpp:6457`) | §1.8 (`POFrmRefund.cpp:1860-1863`) |
| 5814 เพิ่มหนี้ | 6 `DN_AP` (ประเภท 0) · 51 `DNWRONGP_AP` (ราคาผิด) | Source 4 / group 4 → `ProcessConditionGroup4` | §1.9 (`pofrminvcreditnote.cpp:3888-3896`) |
| 5811 จ่ายเงินมัดจำ | 0 `DEPSIT_AP` | Source 1 → `ProcessConditionGroup0` | §1.15 |
| 5806 จ่ายเงินล่วงหน้า | 42 `ADVANCEPAYMENT_AP` | Source 42 / group 45 → `ProcessConditionGroup42` (`myespeciallyformview.cpp:6458`) | §1.16 |
| 5807 รับคืนเงินจ่ายล่วงหน้า | 43 `ADVANCERET_AP` | Source 43 / group 46 → `ProcessConditionGroup43` (`myespeciallyformview.cpp:6459`) | §1.17 (`GetGLFormat :1361-1365`) |

ถ้าผู้ใช้เลือกลงบัญชีมือ (`m_nManualKey ≠ 0`) แทนอัตโนมัติ จะเรียก `OnSaveAccount(...)` เปิดหน้าจอบันทึกรายการบัญชีให้แก้เอง (F2 = ดูรายการบัญชี) — พฤติกรรมเดียวกับจอเอกสารสต็อกใน `07-inventory.md` §0.1. เอกสารที่ผ่าน GL แล้วแก้/ลบไม่ได้ (`CANNOT_SAVE_POSTGL`) ต้องย้อนผ่านการยกเลิก/กลับรายการก่อน — ตรงกับกฎของ BC (เอกสารที่ผ่านบัญชีแล้วแก้ด้วยรายการกลับ/ปรับปรุง ห้ามแก้หรือลบเงียบๆ). **BC ห้ามกำหนดบัญชีตามรหัสบัญชี** (กฎ "ห้ามเงื่อนไขตามรหัสบัญชี") — รูปแบบ GL ต้องเป็นข้อมูลที่ผู้ใช้ตั้ง ไม่ใช่เงื่อนไขในโค้ด.

### 0.5 การส่งต่อสถานะระหว่างเอกสาร `UpdateDocumentFlow(nMark, nMyType, RefNo)`

ฟังก์ชันกลาง `UpdateDocumentFlow` (`mytools.cpp:1674-2425`) คำนวณสถานะ/ยอดคงเหลือของ **เอกสารต้นทาง** ใหม่จากเอกสารลูกที่อ้างถึง ; `nMark` = 1 ตอนบันทึก, -1 ตอนลบ/ยกเลิก/แก้ไข ; `RefNo` = เลขที่เอกสารต้นทางหลายเลขคั่นด้วยเครื่องหมายจุลภาค (ว่าง = ไม่ทำอะไร `:1680-1683`). `nMyType` คือ "ชนิดการคำนวณ" ไม่ใช่ชนิดเอกสารที่บันทึกตรง ๆ — ตารางด้านล่างเฉพาะชนิดที่ฝั่งซื้อเรียก :

| nMyType | เรียกจาก | RefNo คือ | ผลที่คำนวณใหม่ | หลักฐาน |
|---|---|---|---|---|
| 2 | 5798 GR, 5812 ซื้อ, 5808 ยกเลิก PO (บันทึก/ลบ/แก้) | เลขที่ใบสั่งซื้อต้นทาง | `ΣPO = Σ BCPurchaseOrderSub.Qty` (ไม่ยกเลิก) เทียบ `Σรับ = Σ BCAPInvoiceSub.Qty ที่ PORefNo=PO + Σ BCPOVoidSub.VoidQty` : `ΣPO ≤ Σรับ` → `BillStatus=1, PoStatus=2` ; `0 < Σรับ < ΣPO` → `BillStatus=0, PoStatus=3` ; ไม่มีเลย → `0/0` (รวมตามหน่วยฐานด้วย `OnGetQtyPacking`) — **ไม่นับรายการใน `BCIRSUB`** (IR ไม่ตัดยอด PO เอง) | `mytools.cpp:1696-1756` |
| 7 | 5800, 5813 (บันทึก/ลบ/แก้) | เลขที่ GR (5800) หรือเลขที่เอกสารต้นทางทั้งหมดที่ลดหนี้ (5813) | `IsCreditNote = 1` เมื่อ `Σ BCAPInvoiceSub.Qty = Σ BCStkRefundSub.TempQty` (คืนครบ) ; `IsConfirm = 1` เมื่อมีการคืนบางส่วนขึ้นไป ; เขียนเฉพาะเมื่อค่าเปลี่ยน | `mytools.cpp:1899-1960` |
| 15 | 5810 ใบสั่งซื้อ (บันทึก/ลบ/ยกเลิก) — **5803 ไม่เรียก** (ตั้งสถานะใบอนุมัติเองทั้งใบ §1.14) | เลขที่ใบอนุมัติเสนอซื้อ | รีเซ็ต `BCReqConfirmSub.RemainQty = ConfirmQty` แล้วลบด้วย `Σ BCPurchaseOrderSub.Qty` (`ConfirmNo`, `RefLineNumber`, ไม่ยกเลิก) ; `BCReqConfirm.BillStatus = 1` เมื่อ `Σ RemainQty ≤ 0` | `mytools.cpp:2128-2178` |
| 16 | 5805 อนุมัติเสนอซื้อ | เลขที่ใบเสนอซื้อ | รีเซ็ต `BCStkRequestSub.RemainQty = Qty` แล้วลบด้วย `Σ BCReqConfirmSub.ConfirmQty` ; `BCStkRequest.BillStatus = 1` เมื่อ `Σ RemainQty ≤ 0` ; `IsConfirm = 1` เมื่อ `Σ RemainQty ≠ Σ Qty` | `mytools.cpp:2179-2232` |
| 21 | 5799 IR (บันทึก/ลบ/แก้) | **ตอนบันทึก (nMark=1) = เลขที่ IR** ; **ตอนลบ/ยกเลิก (nMark≠1) = เลขที่ GR** (ฟังก์ชันตีความ RefNo ต่างกันตาม nMark) | `BCAPInvoiceSub.GRRemainQty = Qty − Σ BCIRSUB.Qty` ของ GR ที่ IR อ้าง ; `BCAPInvoice.GRBillStatus = 0` ถ้า `Σ GRRemainQty > 0` มิฉะนั้น `1` (เฉพาะ `GRIRBillStatus=1`) | `mytools.cpp:2300-2347` |
| 23 | 5810 ใบสั่งซื้อที่อ้างใบสั่งขาย | เลขที่ใบสั่งขาย | รีเซ็ต `BCSaleOrderSub.POQty = Qty` แล้วลบด้วย `Σ BCPurchaseOrderSub.Qty` (`SORefNo`, `SOLineNumber` — **ไม่กรอง `IsCancel`**) | `mytools.cpp:2390-2423` |

ชนิด 1, 3 มีเนื้อว่าง (`:1692-1695,1757-1760`) ; ชนิด 4/22 (ขาย), 5, 6 (ฝั่งลูกหนี้), 8-14, 17-20 (สต็อก/วางบิล/จ่ายชำระ) ไม่ใช่ของฝั่งซื้อ — ดู `02-sales.md`, `04-ar.md`, `07-inventory.md`, `03-ap.md`.

**ข้อบกพร่องของ Champ ที่พิสูจน์ได้จากตารางนี้ (BC อย่าลอก)** :
1. **แก้ไข IR แล้วสถานะ GR ค้าง** : `POFrmInvoiceReceive.cpp:1166-1167` เรียก `UpdateDocumentFlow(-1,21,เลขที่ IR เดิม)` แต่ชนิด 21 ตอน nMark=-1 ตีความ RefNo เป็น **เลข GR** (`mytools.cpp:2330-2347`) — ทำให้ไม่มีผลกับ GR จริงจนกว่าจะบันทึกครั้งต่อไป (nMark=+1) ; ยังไม่ verify ที่ runtime ว่ายอดคลาดเคลื่อนจริงหรือไม่ (§1.4 ข้อสังเกต).
2. **ชนิด 23 ไม่กรองใบสั่งซื้อที่ยกเลิก** (`:2402`) ต่างจากชนิด 2/15 ที่กรอง `IsCancel = 0` — ใบสั่งซื้อที่ยกเลิกแล้วยังหักยอดจองจากใบสั่งขาย.
3. **ชนิด 2 นับเฉพาะ `BCAPInvoiceSub` (GR / ซื้อตรง) และ `BCPOVoidSub`** ไม่นับ `BCIRSUB` — รายการคอลัมน์ที่ IR บันทึกใน §1.4 ไม่มี `PORefNo` จึงสรุปว่า IR ไม่ตัดยอด PO เอง (ยอด PO ลดตอน GR/ซื้อตรง) ; กรณี "ตั้งหนี้ก่อน รับของทีหลัง" (GR เลือก IR มาด้วยปุ่ม `SELECT_IR` §1.3) ยังไม่ verify ที่ runtime ว่าสถานะ PO ถูกต้อง.
4. **ไม่มี transaction** : ทุกชนิดรัน `UPDATE` ทีละประโยคด้วย `Execute` โดยไม่ครอบด้วย `BEGIN/COMMIT` ของตัวเอง — ล้มกลางทางได้สถานะค้างครึ่งหนึ่ง ; BC ต้องคำนวณทั้งหมดใน `sql.Tx` เดียวกับการบันทึกเอกสาร (ไม่ใช่เรียกซ้ำตอนหลัง).

### 0.6 สิทธิ์และตัวเลือกร่วม

- สิทธิ์ผู้ใช้ระดับเมนู/ปุ่มใช้กลไกกลางของ Champ (`00-common-ui.md` §11) ; สิทธิ์ย่อยเฉพาะเอกสารที่พบในจอซื้อ : `bCanChangeBillNo` (แก้เลขที่เอกสาร — §1.1, §1.3, §1.6, §1.12), `bCanChangeBillAmount` (§1.1, §1.4), `bCanChangeBillDiscount` (§1.1), `bCanChangeGRBuyPrice` (§1.3), `IsCanEditDocDate` (§1.2, §1.12) ; เมนู 5049 ผูกกับ command `ID_APSYS_TRANS_7` (ฝั่งระบบเจ้าหนี้) ไม่ใช่ `ID_POSYS_*` (§1.19).
- ปิดงวดบัญชี (`IsPeriodClose`) ตรวจก่อนบันทึก/ลบในจอเอกสารซื้อส่วนใหญ่ (ระบุไว้ที่กติกาก่อนบันทึกของแต่ละหัวข้อ) — **5048 ต้นทุนแฝงไม่ตรวจงวดและไม่มีธุรกรรมครอบ** (§1.18 ข้อสังเกต 1 ; ต้นทุนแฝงเปลี่ยนต้นทุนสินค้าย้อนหลัง จึงเสี่ยงที่สุดใน BC).
- จำนวนทศนิยมมาจากตั้งค่าบริษัท `QtyPoint/PricePoint/AmountPoint` ; หลายสกุลเงินเปิดด้วย `MultiCurrency` (ถ้า 0 อัตราแลกเปลี่ยนถูกบังคับเป็น 1).
- รหัสผู้ใช้ผู้สร้าง/ผู้แก้ (`CreatorCode`, `LastEditorCode`) และเวลา (`CreateDateTime`, `LastEditDateT`) บันทึกทุกเอกสาร ; การลบ PO/เอกสารเป็น **ลบจริง** (`DELETE`) ไม่ใช่ soft-delete (`ChmFrmPurchaseOrder.cpp` ตาม §1.1) และไม่มี audit trail แบบ append-only — BC ต้องออกแบบใหม่ตาม Rule 20 (แก้ด้วยรายการกลับ/ปรับปรุง).

## 1. เอกสารและกระบวนการ (19 รายการ)

### 1.1 รายการ 5810 "บันทึกใบสั่งซื้อสินค้า" (ใบสั่งซื้อ — PO)

- **ฟอร์ม**: `CChmFrmPurchaseOrder` (`ChmFrmPurchaseOrder.cpp`, 2,453 บรรทัด) = หน้าต่างหลัก + แท็บ 2 หน้า `CChmFrmPurchaseOrderPage1` (หัวเอกสาร+กริด+ยอดรวม) และ `CChmFrmPurchaseOrderPage2` (ข้อมูลอื่น). dialog `IDD_CHM_PO_FRM_PO` 444x290, Page1 610x280, Page2 540x280 (`BC5Account.rc:17642,17656,17717`). handler เมนู `OnPosysTrans4` (atlas-01-PO). แท็บชื่อ "ข้อมูลหลัก" / "รายการอื่นๆ" (MST_HEADER/MST_OTHER ตาม §0.2).
- **เป้าหมาย**: ออกใบสั่งซื้อถึงเจ้าหนี้ — **ไม่ลงบัญชี GL, ไม่ตัดสต็อก, ไม่ตั้งหนี้** (ไม่มี `GLTrans`/`GLFormat` ในขั้นบันทึก `OnSaveData` :800-1089); ใช้เป็นเอกสารอ้างอิงให้ 5798 รับสินค้า / 5799 ตั้งหนี้ / 5812 ซื้อ / 5800 ส่งคืน เรียกกลับมาตั้งยอดค้างรับ (`RemainQty`).

**A. หัวเอกสาร แท็บ "ข้อมูลหลัก" (Page1) — ลำดับตามตำแหน่งบนจอ (x,y หน่วย dialog)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / แหล่งข้อมูล / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบสั่งซื้อ: | EDIT + ปุ่มค้นหา (`m_edDocNo`, กล่องรันเลข) (88,1) | ใช่ | รันเลขอัตโนมัติเมื่อกดเอกสารใหม่ (`RunningAgain()` ถ้าไม่มีกลุ่มเอกสาร, ไม่งั้นรันตามกลุ่มของผู้ใช้ `GetUserDocGroup(0)`); ถ้าผู้ใช้ **ไม่มีสิทธิ์แก้เลขที่บิล** (`bCanChangeBillNo`=FALSE) ช่องถูกปิด (disable) และโฟกัสไปที่รหัสเจ้าหนี้; ตรวจเลขซ้ำก่อนบันทึก | `ChmFrmPurchaseOrder.cpp:1485-1514`, `ChmFrmPurchaseOrderPage1.cpp:182-183,211-214`, `ChmFrmPurchaseOrder.cpp:734-778` |
| 2 | วันที่ออกเอกสาร: | วันที่ พ.ศ. อัตโนมัติ `BME_DATE_TH_AUTO` + ปุ่มปฏิทิน (269,1) | ใช่ | ค่าเริ่มต้น = "วันทำงาน" (`theApp.m_strWorkingDate`) ถ้ามี; ห้ามบันทึกถ้าอยู่ในงวดที่ปิด (`IsPeriodClose(...,TRUE)`) | `Page1.cpp:185-186,1859-1862`; `ChmFrmPurchaseOrder.cpp:517-520` |
| 3 | ประเภทภาษี: | Combo 3 ค่า (88,16) | ใช่ (มีค่าเริ่มต้น) | ค่าเริ่มต้น = `CompanyInfo.DefaultTaxType`; ค่า 0 = **แยกนอก**, 1 = **รวมใน**, 2 = **ภาษีอัตราศูนย์** (คำบนจอจาก key `CB_TAXTYPE_1`, `CB_TAXTYPE_2`, `TAX_ZERO` «ไม่พบไทยใน repo»); เปลี่ยนแล้วคำนวณทุกแถวใหม่ (`OnSelchangeTaxType`) | `Page1.cpp:166-174`; `BCStools.cpp:165,195,223` |
| 4 | แผนก: | Combo-table จาก `BCDepartment` (269,16) | ไม่ | ค่าเริ่มต้นของแผนกต่อแถว (แถวไม่มีแผนก → ใช้แผนกหัวเอกสาร) | `Page1.cpp:192`; `ChmFrmPurchaseOrder.cpp:971-972` |
| 5 | รหัสเจ้าหนี้: | ช่อง Link ตรวจกับ `BCAP.Code` + ปุ่มเลือก (88,30) + ช่องชื่อ (อ่านอย่างเดียว) (188,30) | ใช่ | ปุ่ม/Link เปิดรายการเจ้าหนี้ (`Browse2("ApTransConfig")`, filter `Code LIKE 'xxx%'`); ชื่อที่แสดง = `Name1` หรือ `Name2` ตาม `CompanyInfo.DefaultApName`. เมื่อเลือกเจ้าหนี้ ระบบดึง: ภาษีเริ่มต้นของเจ้าหนี้ (`DefaultTaxRate`; ถ้าเจ้าหนี้ไม่ตั้ง (<0) ใช้ `CompanyInfo.DefaultTaxRate`), กลุ่มรายละเอียดสินค้า (`ItemDescGroup`), ส่วนลดท้ายบิล (`Discount`; ล้างทิ้งถ้า `CompanyInfo.BillDiscount=0`), ส่วนลดต่อรายการ (`DetailDiscount` → ใส่ทุกแถวที่ยังไม่อ้างอิงใบเสนอซื้อ), ผู้ติดต่อ (`DefContactCode`), เครดิต(วัน), ระยะเวลาขนส่ง, แล้วคำนวณวันครบกำหนด/วันที่รับสินค้า = วันที่เอกสาร + เครดิต/ขนส่ง | `Page1.cpp:188-190,542-622` |
| 6 | (ปุ่ม) เลือกใบอนุมัติเสนอซื้อ (key `STK_REQUEST`) | ปุ่ม (413,23) | — | ดึงรายการค้างจากใบอนุมัติเสนอซื้อ (5805) เข้ากริด — ต้องมีเจ้าหนี้ก่อน (`PLEASE_INSERT_AP`) | `Page1.cpp:1519-1740` (`OnBttnReqConfirm`) |
| 7 | (ปุ่ม) เลือกใบสั่งขาย/สั่งจอง (key `BROWSE_SOSR`) | ปุ่ม (413,0) | — | ดึงรายการจากใบสั่งขาย (ลูกค้า→สั่งซื้อให้ลูกค้า); อ่าน `BCSALEORDERSUB` | `Page1.cpp:715-760` (`OnBttnSaleOrder`) |

**B. หัวเอกสาร แท็บ "รายการอื่นๆ" (Page2)**

| # | ป้ายบนจอ | ชนิด | ค่าเริ่มต้น / กติกา | บันทึกที่คอลัมน์ | หลักฐาน |
|--|--|--|--|--|--|
| 1 | โครงการ: | Combo-table `BCProject` (85,18) | ค่าเริ่มต้นของโครงการต่อแถว (แถวว่าง → ใช้ค่าหัวเอกสาร) | `ProjectCode` | `Page2.cpp:92`; `ChmFrmPurchaseOrder.cpp:963-965` |
| 2 | รหัสสกุลเงิน: | Combo-table `BCCurrency` (301,18) | ปิดช่องถ้า `CompanyInfo.MultiCurrency=0` | `CurrencyCode` | `Page2.cpp:95,101-104` |
| 3 | รหัสผู้ติดต่อ: | Combo-table `BCContactList WHERE MyType = 1` (85,30) | กรองตามเจ้าหนี้ (`OnContractFilter`) | `ContactCode` | `Page2.cpp:75`; `ChmFrmPurchaseOrder.cpp:1206` |
| 4 | อัตราแลกเปลี่ยน: | ตัวเลข 9 ทศนิยม (301,30) | ต้องกรอก (`MUST`), ปิดถ้าไม่ใช้หลายสกุลเงิน; <=0 → บันทึกเป็น 1.0 | `ExChangeRate` | `Page2.cpp:96,99,102`; `ChmFrmPurchaseOrder.cpp:833-836` |
| 5 | การจัดสรร: | Combo-table `BCAllocate` (85,42) | ค่าเริ่มต้นการจัดสรรต่อแถว | `AllocateCode` | `Page2.cpp:91` |
| 6 | อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข (301,42) | ดึงจากเจ้าหนี้; แก้แล้วคำนวณทุกแถวใหม่ (`OnKillfocusEdTaxrate`) | `TaxRate` (หัว+แถว) | `Page2.cpp:94,315-340` |
| 7 | เครดิต(วัน): | ตัวเลข 0 ทศนิยม (85,54) | แก้แล้ว วันครบกำหนด = วันที่เอกสาร + เครดิต; ถ้าเครดิตที่แก้ **น้อยกว่า** เครดิตเดิมของเจ้าหนี้ → ดึง "ส่วนลดเงินสด" `BCAP.CashDiscount` มาเป็นส่วนลดท้ายบิล | `CreditDay` | `Page2.cpp:114-133` |
| 8 | เงื่อนไขการรับสินค้า: | Combo 2 ค่า (key `OWN_RECEIVE`, `SEND_TO`) (301,54) | ค่าเริ่มต้นตัวเลือก 0 | `IsConditionSend` | `Page2.cpp:77-83`; `ChmFrmPurchaseOrder.cpp:824-825` |
| 9 | เอกสารหมดอายุภายใน: (วัน) | ตัวเลข (85,66) | วัน>0 → วันหมดอายุ = วันที่เอกสาร + วัน; กรอกวันหมดอายุ → คำนวณกลับเป็นจำนวนวัน | `ExpireCredit` | `Page2.cpp:184-215` |
| 10 | วันหมดอายุ: | วันที่ (301,66) | (ดูข้อ 9) | `ExpireDate` | เดียวกัน |
| 11 | ระยะเวลาขนส่ง: (วัน) | ตัวเลข (85,78) | แก้แล้ว กำหนดส่ง = วันที่เอกสาร + วัน; แก้กำหนดส่ง → คำนวณวันกลับ | `LeadTime` | `Page2.cpp:153-182` |
| 12 | กำหนดส่ง: | วันที่ (301,78) | (ดูข้อ 11) | `LeadDate` | เดียวกัน |
| 13 | สั่งซื้อให้กับ: | Link ลูกค้า `BCAR` + ชื่อ (85,121) | ใช้กรณีสั่งซื้อแทนลูกค้า (คู่กับใบสั่งขาย) | `OrderToArCode` | `Page2.cpp:106,217-250`; `ChmFrmPurchaseOrder.cpp:820` |
| 14 | ใบสั่งขาย/สั่งจองสินค้า | Link `BCSaleOrder` (85,133) | **ต้องระบุลูกค้าก่อน** (`PLEASE_INSERT_AR`); ตรวจว่ามีเอกสาร `BCSaleOrder` ของลูกค้านั้นจริง ไม่งั้นแจ้งและล้างช่อง | `SORefNo` | `Page2.cpp:252-313` |
| 15 | กลุ่มเอกสาร: | ข้อความ อ่านอย่างเดียว (301,133) | จาก `BillGroup` | `BillGroup` | `ChmFrmPurchaseOrder.cpp:819,1221` |
| 16 | หมายเหตุ: | ข้อความ (85,145) | — | `MyDescription` | `ChmFrmPurchaseOrder.cpp:818` |
| 17 | (วันที่ครบกำหนด) | วันที่ (300,194) | คำนวณจากเครดิต แสดงอย่างเดียว — **ไม่บันทึก** (`AddFieldForSave("DueDate"…)` ถูก comment ทิ้ง) | — | `ChmFrmPurchaseOrder.cpp:812` |

**C. กริดรายการสินค้า (9 คอลัมน์ที่เห็น) — `Page1.cpp:101-109`**

| # | ป้าย (key) | ชนิด/flag | ความหมาย / การคำนวณ |
|--|--|--|--|
| 0 | รหัสสินค้า (`ITEMCODE`) กว้าง 70 | แก้ได้ + Link | พิมพ์รหัส / `.` = รายการข้อความล้วน (ไม่ใช่สินค้า) / รูปแบบ `จำนวน*บาร์โค้ด` เติมจำนวน+หน่วยจากบาร์โค้ด; แปลงเป็นตัวพิมพ์ใหญ่; หาไม่พบใน `BCItem` ที่ `ActiveStatus=1` → แจ้ง `NOT_SEEITEMCODE` และล้างแถว (`UpdateAndCheckStockOption` :363-525) |
| 1 | ชื่อสินค้า (`ITEMNAME`) 150 | อ่านอย่างเดียว (แก้ได้ถ้า `Option.IsEditItemName=1`) | ชื่อตาม `Name1/Name2` (`CompanyInfo.DefaultItemName`); รายการ `.` ใช้ข้อความอธิบาย |
| 2 | คลังสินค้า (`WAREHOUSE_ITEM`) 60 | Combo-table (`BCItemWarehouse` join `BCWarehouse`, Top 50) | ค่าเริ่มต้น = คลังที่ตั้งไว้ท้ายฟอร์ม (`m_strWHCode`) หรือ `BCItem.DefBuyWHCode` |
| 3 | ที่เก็บ (`SHELF`) 60 | Combo-table (`BCItemWarehouse` join `BCShelf`) | `m_strShelfCode` หรือ `BCItem.DefBuyShelf` |
| 4 | หน่วยนับ (`UINTCODE`) 60 | Combo-table `BCItemUnit` | ค่าเริ่มต้น `BCItem.DefBuyUnitCode` (หน่วยชนิด 2 "หน่วยขาน" → `OppositeUnit`); ต้องเป็นหน่วยของสินค้านั้น (`IsUnitOfItem` ไม่งั้น `NOUNIT`); ถ้าแถวมาจากใบอนุมัติเสนอซื้อ **เปลี่ยนหน่วยไม่ได้** (`CAN_NOT_CHANG_UNIT`) |
| 5 | จำนวน (`ITEMQTY`) 60 | ตัวเลข ชิดขวา (ทศนิยม `QtyPoint`) | ถ้า <=0 ใส่ค่าเริ่มต้น `CompanyInfo.BeginQty`; ถ้าแถวมาจากใบอนุมัติเสนอซื้อ จำนวนเกินที่อนุมัติไม่ได้ (`MORE_QTY_APPROVE` ปรับกลับเป็นจำนวนที่อนุมัติ); ตรวจสั่งซ้ำ (`OrderPODuplicate`) และของแถมจากโปรโมชั่นผู้ขาย (`OnSupplierPromotion`/`BCSUPPLIERPROMO`) |
| 6 | ราคาซื้อ (`BUYPRICE`) 60 | ตัวเลข (`PricePoint`) | ถ้าว่าง/0 ดึงราคาซื้อล่าสุดตามเจ้าหนี้+หน่วย+วันที่ (`GetItemBuyPrice`); ถ้าเปิด `Option.BuyPriceWarning=1` และราคาซื้อ(หลังส่วนลด × อัตราแลกเปลี่ยน) สูงกว่าราคาขาย → แจ้ง `BUYPRICE_MORE_SALEPRICE` และล้างราคา |
| 7 | ส่วนลด%,บาท (`DISCOUNT`) 70 | ข้อความ (เช่น `5%`, `10+3%`, `50`) | ค่าเริ่มต้น (เฉพาะเอกสารใหม่): ถ้า `Option.IsGetLastDiscount=1` ใช้ส่วนลดของการซื้อล่าสุด (`BCAPINVOICESUB.DiscountWord`) ไม่งั้นใช้ `BCAP.DetailDiscount` |
| 8 | จำนวนเงิน (`AMOUNT`) 70 | ตัวเลข (`AmountPoint`) | คำนวณจากจำนวน × ราคา แล้วหักส่วนลด; **แก้จำนวนเงินตรง ๆ ได้**: ถ้าราคาว่างจะหารกลับเป็นราคา = จำนวนเงิน/จำนวน |

สูตรแถว (`DetailCalc` `Page1.cpp:624-687`): `DiscountMethod=0` → จำนวนเงิน = ปัดเศษ(`XDiscount(จำนวน×ราคา)`, `AmountPoint`); `DiscountMethod≠0` → หักส่วนลดจาก **ราคาต่อหน่วย** ก่อน (ปัด `PricePoint`) แล้วคูณจำนวน (ปัด `AmountPoint`) และส่วนลดแถว = ส่วนลดต่อหน่วย × จำนวน. จากนั้น `NetAmount` (shadow 7) = จำนวนเงิน; ถ้า **ภาษีรวมใน** (TaxType=1) และสินค้าไม่ยกเว้นภาษี → `NetAmount = จำนวนเงิน − จำนวนเงิน×อัตราภาษี/(100+อัตราภาษี)`; `HomeAmount` (shadow 8) = NetAmount × อัตราแลกเปลี่ยน.

**Shadow columns (ไม่แสดง แต่บันทึกลงตาราง) — `Page1.cpp:116-149`**: 0 RowOrder · 1 ItemType (ประเภทสินค้า) · 2 MyDescription (คำอธิบายรายการ) · 3 ProjectCode · 4 AllocateCode · 5 OppositeQty · 6 OppositeUnit · 7 NetAmount · 8 HomeAmount · 9 ExceptTax (0 ปกติ/1 ยกเว้นภาษี/2 ภาษีอัตราศูนย์) · 10 BarCode · 11 CategoryCode (หมวด) · 12 GroupCode (กลุ่ม) · 13 BrandCode (ยี่ห้อ) · 14 TypeCode (ประเภท) · 15 FormatCode (รูปแบบ) · 16 WHCode · 17 RefDocNo (เลขที่ใบเสนอซื้อ → คอลัมน์ `StkReqNo`) · 18 ItemCodePormo · 19 PromotionItem ("999" = ของแถม ไม่ดึงราคา/ไม่คำนวณ) · 20 DepartCode · 21 CurrencyCode · 22 ExchangeRate · 23 ConfirmNo (เลขที่ใบอนุมัติเสนอซื้อ) · 24 DiscountAmount (ส่วนลดแถวเป็นเงิน) · 25 ComfirmQty (จำนวนที่อนุมัติ) · 26 ComfirmUnit (หน่วยที่อนุมัติ) · 27 ColorCode · 28 SizeCode · 29 StyleCode · 30 BehindIndex · 31 OldItemCode · 32 RefLineNumber · 33 SORefNo · 34 SOLineNumber.

**D. ยอดรวมท้ายเอกสาร (คอลัมน์ขวา) และสูตร**

| ป้ายบนจอ | ช่อง | แก้ได้ | สูตร / บันทึกที่คอลัมน์ |
|--|--|--|--|
| (ข้อความรายละเอียดสินค้าแถวปัจจุบัน) | `IDC_ED_ITEMDETAIL` | ไม่ | แสดงข้อความ `GetItemNameDetail` ของแถวที่เลือก (`Page1.cpp:1160`) |
| รวมมูลค่าสินค้า: | `m_edSumDetail` | เฉพาะผู้มีสิทธิ์แก้ยอดบิล (`bCanChangeBillAmount`) | ผลรวมจำนวนเงินของแถว **ปกติ** (ไม่รวมแถวยกเว้นภาษี และแถวอัตราศูนย์) → `SumOfItemAmount` (`BillCalc` `Page1.cpp:1810-1838`) |
| มูลค่าสินค้ายกเว้นภาษี: | `m_edExceptTax` | เหมือนกัน | ผลรวมแถว ExceptTax=1 **เมื่อ** `CompanyInfo.ExceptTax=1` → `ExceptTaxAmount`; แถว ExceptTax=2 เมื่อ `CompanyInfo.ZeroTaxRate=1` → `ZeroTaxAmount` |
| ส่วนลด%,บาท: | ปุ่ม `%` (เปิดกล่องใส่ส่วนลดเป็น %/ข้อความ) + `m_edDiscountAmount` | ปิดถ้า `CompanyInfo.BillDiscount=0` หรือผู้ใช้ไม่มีสิทธิ์ `bCanChangeBillDiscount` | `DiscountWord` (ข้อความส่วนลด), `DiscountAmount` |
| ภาษีมูลค่าเพิ่ม: | `m_edTaxAmount` | เฉพาะผู้มีสิทธิ์แก้ยอดบิล | `TaxAmount` |
| มูลค่ารวมภาษี: | `m_edTotalAmount` | เหมือนกัน | `TotalAmount` |
| มูลค่าเงินสุทธิ: | `m_edNetAmount` | เหมือนกัน | `NetAmount` = TotalAmount + ExceptTax + ZeroTax (`CalcTaxAmount` `Page1.cpp:1803-1807`) |

สูตรภาษีท้ายบิล (`BillTaxProcess` `lib/BC5FRMWRK/BCStools.cpp:155-233`, เรียกจาก `CalcTaxAmount` `Page1.cpp:1763` ด้วย `nProcessMode=0`): หลังหักส่วนลด `AfterDiscount = ผลรวมสินค้า − ส่วนลด`; **แยกนอก** (0): `BeforeTax = AfterDiscount` ; `Tax = ปัด(BeforeTax×อัตรา/100)`; `Total = BeforeTax + Tax`; **รวมใน** (1): `Tax = ปัด(AfterDiscount×อัตรา/(100+อัตรา))`; `BeforeTax = AfterDiscount − Tax`; `Total = AfterDiscount`; **อัตราศูนย์** (2): `Tax = 0`; `Total = AfterDiscount`. ปัดเศษด้วย `AmountPoint` (ทศนิยมบริษัท). ถ้าผลรวมสินค้า <= 0 → ล้างส่วนลด/ภาษี/รวม.

**E. ปุ่ม / คีย์ลัด / พฤติกรรมพิเศษ**

- ปุ่มแถบเครื่องมือมาตรฐาน (§0.2): ใหม่ / บันทึก / ลบ / ค้นหา / ก่อนหน้า-ถัดไป / ปิด / ข้อมูล (`ID_REC_INFO`) / ทำซ้ำ (`IDC_RECUR` = เอกสาร Recurring) / "เพิ่มเติม" (`IDC_MORE`). `ChmFrmPurchaseOrder.cpp:74-95`.
- คีย์ลัดที่ผูกจริงในฟอร์มนี้ (`ON_COMMAND`): บันทึก (F11 `ID_HK_SAVE11`), ใหม่ (`ID_HK_NEW`), ค้นหา (`ID_HK_FIND`), พิมพ์ (`ID_HK_PRINT`), **ยืนยันเอกสาร** (`ID_HK_CONFIRMDOC` = Shift+F5), **ยกเลิกเอกสาร** (`ID_HK_CANCELDOC` = Ctrl+F5), รายละเอียดเอกสาร (F5 `ID_HK_DOCDESC`), ลำดับเอกสาร/Document Flow (F12 `ID_HK_DOCFLOW`), ดึงใบสั่งขาย (`ID_HK_GETSO`), ขยายแถว (Shift+F2 → กล่อง `CPODlgPurchaseOrder`), ตั้งคลัง/ที่เก็บเริ่มต้น (Ctrl+F7 / Ctrl+F8), ข้อมูลสินค้า (Ctrl+F12), PgUp/PgDn (เอกสารก่อน/ถัดไป). หมายเหตุ: ไม่มี F2 (ลง GL), F3 (ภาษี), F6 (หัก ณ ที่จ่าย) เพราะ PO ไม่มีการลงบัญชี — `OnHKUpdateUI` ปิดปุ่มเหล่านี้ (`ON_UPDATE_COMMAND_UI_RANGE(ID_HK_GL,ID_HK_WTAX)` :95). ตารางรหัสปุ่มลัดเต็มดู `BC5Account.rc:977-1006` และ 07-inventory.md §0.1.
- ขยายแถว (`OnExpand` `Page1.cpp:1890`): เปิดกล่องแก้ชื่อ/รายละเอียดสินค้าต่อแถว (เก็บใน shadow 2 = `MyDescription`).
- โหมด "ค้นหาด้วยกริด": กดปุ่มค้นหาในช่องรหัสสินค้า (`OnListFindData` `Page1.cpp:1238`) เปิดกล่องค้นหาสินค้า (`CMyFilterItemDlg` = กลุ่ม/ยี่ห้อ/รูปแบบ/ชนิด/หมวด/สี...); ถ้าบริษัทเปิด `BuyItemSupplier=1` ต้องเลือกเจ้าหนี้ก่อน (`PLEASE_INSERT_AP`) และค้นเฉพาะสินค้าของเจ้าหนี้ (`BCITEMSUPPLIER` :1424).

**F. สถานะเอกสารและ workflow**

| คอลัมน์ | ค่า | ความหมาย |
|--|--|--|
| `IsConfirm` | 0/1 | 1 = ยืนยันแล้ว → **แก้/ลบ/ยกเลิกไม่ได้** (`ConfirmSaveData` :550, `ConfirmDeleteData` :403, `OnConfirmCancel` :2252) |
| `IsCancel` | 0/1 | 1 = ยกเลิกแล้ว → แก้/ลบ/ยกเลิกซ้ำไม่ได้ |
| `PoStatus` | 0 / 1 / 2 / 3 / 4 | 0 = ยังไม่ได้รับสินค้า (key `QUOTATION_NOTUSE`), 1 = Sent, 2 = รับเต็ม (`RECEIVE_FULL`), 3 = รับบางส่วน (`PART_RECEIVE`), 4 = ยกเลิก (`CANCEL`) — แสดงในช่องสถานะบนฟอร์ม (`m_edPOStatus`, เฉพาะโหมด `m_bChmMode`) `:1231-1242`; PoStatus 2/3 → **แก้/ลบ/ยกเลิกไม่ได้** (`NOT_EDIT_BILLBALANCE`/`NOT_DELETE_PO`/`CANNOT_CANCEL`) |
| `BillStatus` | 0/1 | ตั้ง 0 ตอนสร้างใหม่ (`:868`); ใช้วาดป้ายสถานะ `DrawBillStatus(IsConfirm, IsCancel, BillStatus==1?0:1)` (`:1274`) |
| `IsCompleteSave` | 0/1 | ตั้ง 0 ก่อนเขียนหัว แล้วตั้ง 1 หลังเขียนรายการย่อยครบ (กันเอกสารค้างครึ่งทาง `:853,1062`); ถ้าโหลดเอกสารที่ `IsCompleteSave≠1` ถามว่าจะโหลด "รายการเก่า" (`TransState=1`) หรือไม่ (`CANNOT_SAVEDETAIL_WANT_OLDITEM` :1283-1286) |

- **อ้างอิงข้ามเอกสาร**: เอกสารที่อ้าง PO แล้ว (`BCAPINVOICESUB.PORefNo`, `BCIRSUB.PORefNo`, `BCPOVOIDSUB.PONo` ที่ `TransState=0 AND IsCancel=0`) → **แก้/ลบ/ยกเลิก PO ไม่ได้** (`CANNOT_EDIT_REFERENCE` / `NOT_DELETE_PO` / `CANNOT_CANCEL`) — `ChmFrmPurchaseOrder.cpp:562-574,416-428,2265-2276`.
- **ยกเลิกเอกสาร** (Ctrl+F5 → `BillCancel("BCPurchaseOrder")` → `BillCancelUpdateData` `:1436`): `UPDATE BCPurchaseOrder SET IsCancel=1, PoStatus=4`; `UPDATE BCPurchaseOrderSub SET IsCancel=1, IsProcess=0`; คำนวณสต็อกใหม่ของสินค้าทุกตัว (`OnStockReprocess`), บันทึก log "ยกเลิกใบสั่งซื้อ", คืนสถานะเอกสารต้นทาง (`UpdateDocumentFlow(-1,15,…)` ใบอนุมัติเสนอซื้อ, `(-1,23,…)` ใบสั่งขาย).
- **ยืนยันเอกสาร** (Shift+F5 → `BillConfirm("BCPurchaseOrder")` `:2289`) ตั้ง `IsConfirm=1` (รายละเอียดภายใน `BillConfirm` อยู่ใน framework — ยังไม่ verify ในรอบนี้).
- **ลบ** (`ConfirmDeleteData` `:384-434` → `OnDeleteData` `:436`): ตรวจเลขที่ไม่ว่าง → งวดปิดไหม → IsCancel/IsConfirm/PoStatus 2-3/มีเอกสารอ้างอิง → ถามยืนยัน `CONFIRM_DELETEDATA` (Yes/No) → `DELETE FROM BCPurchaseOrder` + `DELETE FROM BCPurchaseOrderSub` + คำนวณสต็อกใหม่ + log "ลบใบสั่งซื้อ" + `UpdateDocumentFlow(-1,15)`/`(-1,23)`. (ลบจริง ไม่ใช่ soft-delete.)
- **แก้ไขเอกสารเดิม** (`OnSaveData` โหมด Update `:873-901`): คืนสถานะใบอนุมัติเสนอซื้อเดิมก่อน (`UpdateDocumentFlow(-1,15,…)`), `UpdateDB(BCPurchaseOrder)` (ตั้ง `PoStatus`, `LastEditorCode`, `LastEditDateT`), **ลบรายการย่อยทั้งหมดแล้ว INSERT ใหม่** (`DELETE FROM BCPurchaseOrderSub` :914) — ทั้งหมดอยู่ใน transaction (`StartTransaction/Commit/Abort` :863,907,1064,1083).
- **ส่งต่อไปเอกสารถัดไป**: เอกสาร 5798/5812 (และ 5799 ผ่าน GR) ดึง PO เข้าไปรับ/ตั้งหนี้ (ดู §1.3-§1.7); ทุกครั้งที่เอกสารเหล่านั้นบันทึก/ลบ/ยกเลิก จะเรียก `UpdateDocumentFlow(…, 2, เลขที่ PO)` ซึ่งรวมจำนวนสั่ง (`BCPurchaseOrderSub.Qty` ที่ไม่ยกเลิก) เทียบกับจำนวนที่ถูกอ้าง (`BCAPInvoiceSub.Qty` ที่ `PORefNo` = PO นี้ + `BCPOVoidSub.VoidQty`) แล้วตั้ง **สั่งครบ/รับครบ** `BillStatus=1, PoStatus=2` เมื่อ `ผลรวมสั่ง ≤ ผลรวมที่ถูกอ้าง`; **รับบางส่วน** `BillStatus=0, PoStatus=3` เมื่อมีที่ถูกอ้าง > 0 แต่ยังไม่ครบ; ไม่มีใครอ้าง `BillStatus=0, PoStatus=0` (`mytools.cpp:1696-1755`). (ยอด `RemainQty` รายบรรทัดของ PO คำนวณใหม่ใน `ProcessStock` ผ่านคิว `OnStockReprocess` — กลไกใน Process engine ยังไม่ verify).

**G. ผลข้างเคียง**: ไม่ลง GL · ไม่ตัดสต็อกจริง แต่เรียก `OnStockReprocess(itemCode, docDate)` ทุกสินค้าในเอกสารหลังบันทึก/ลบ/ยกเลิก (คำนวณใหม่ของสินค้านั้น — ยังไม่ verify ว่าเขียนตารางใด) · คืน/ตัดสถานะเอกสารต้นทางผ่าน `UpdateDocumentFlow(1|-1, ชนิดเอกสารที่กำลังบันทึก, เลขที่เอกสารต้นทาง)` — ชนิด **15 = ใบสั่งซื้อ** → คำนวณ `BCReqConfirmSub.RemainQty` และ `BCReqConfirm.BillStatus` ของใบอนุมัติเสนอซื้อต้นทางใหม่; ชนิด **23 = ใบสั่งซื้อที่อ้างใบสั่งขาย** → คำนวณ `BCSaleOrderSub.POQty` ใหม่ (`mytools.cpp:2128-2179, 2390-2420`; ตาราง §0.5) · `OnUpdateReqComfirm` (:351-382 ลด `BCReqConfirmSub.RemainQty`, ตั้ง `BCReqConfirm.BillStatus=1` เมื่อหมด) **ไม่ถูกเรียกในขั้นบันทึก** (การเรียกถูก comment ที่ :1065) · log book (`AddUserLogBook*`; key `SAVE_PURCHASE_ORDER`/`DELETE_PURCHASE_ORDER`/`CANCEL_PURCHASE_ORDER`).

**H. ตาราง/คอลัมน์ที่เขียน** (`ChmFrmPurchaseOrder.cpp:806-1056`; DDL `champ/Script/SQLSERVER_Script.sql`)

- `BCPurchaseOrder` (หัว): DocNo, DocDate, TaxType, ApCode, DepartCode, CreditDay, LeadTime, LeadDate, ExpireDate, ExpireCredit, TaxRate, MyDescription, BillGroup, OrderToArCode, ContactCode, AllocateCode, ProjectCode, IsConditionSend, SORefNo, CurrencyCode, ExChangeRate, RecurName, SumOfItemAmount, DiscountWord, DiscountAmount, AfterDiscount, BeforeTaxAmount, TaxAmount, TotalAmount, ExceptTaxAmount, NetAmount, ZeroTaxAmount, IsCompleteSave; ใหม่เพิ่ม BillStatus=0, CreatorCode (10 ตัวอักษรแรกของรหัสผู้ใช้), CreateDateTime; แก้ไขเพิ่ม PoStatus, LastEditorCode, LastEditDateT.
- `BCPurchaseOrderSub` (รายการ): TaxType, DocNo, DocDate, ApCode, ItemCode, ItemName, WHCode, ShelfCode, UnitCode, Qty, Price, DiscountWord, Amount, **RemainQty (=Qty ตอนบันทึก)**, DiscountAmount, ItemType, MyDescription, ProjectCode, AllocateCode, OppositeQty, OppositeUnit, NetAmount, HomeAmount, ExceptTax, BarCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, StkReqNo, DepartCode, CurrencyCode, ExChangeRate, ConfirmNo, ColorCode, SizeCode, StyleCode, TaxRate, PackingRate1, PackingRate2 (อัตราบรรจุของหน่วยนั้น; ไม่พบ → 1.0), behindindex, RefLineNumber, SORefNo, SOLineNumber, LineNumber (=ลำดับแถว 0..n), IsProcess=0; (ถ้า `Option.m_nProductCentralize=1` เพิ่ม SCGProductId, SCGBarcode, SCGProductname, DealerBuyUnit).
- ข้ามแถวที่รหัสสินค้าว่าง (`!XIsStringEmpty(ItemCode)` :931).

**I. ตรวจสอบก่อนบันทึก (`ConfirmSaveData` :491-798) ตามลำดับ** — ข้อความบนจอมาจาก key (ไทยเต็มไม่อยู่ใน repo; ไทยประมาณ (~) จากชื่อ key)

| ลำดับ | เงื่อนไขที่ไม่ผ่าน | key / ข้อความ | หลักฐาน |
|--|--|--|--|
| 1 | เลขที่เอกสารว่าง | `CANNOT_SAVE_NOTDOCNO` (~ไม่สามารถบันทึกได้ เนื่องจากไม่ได้ระบุเลขที่เอกสาร) | :503-506 |
| 2 | (แก้ไขเอกสาร) แก้เลขที่ไปซ้ำกับเอกสารอื่น | `DOCNO_DUPLICATE` | :507-514 |
| 3 | วันที่อยู่ในงวดปิด | (ข้อความจาก `IsPeriodClose`) | :517 |
| 4 | รหัสเจ้าหนี้ว่าง / ไม่พบใน `BCAP` | `CANNOT_SAVE_NOTAPCODE` / `NOT_APCODE` = "ไม่พบเจ้าหนี้ตามรหัสที่ระบุ" (comment :538) | :522-541 |
| 5 | (แก้ไข) เอกสารยกเลิก / ยืนยัน / PoStatus 2-3 / มีเอกสารอ้างอิง | `NOT_SAVE_ISCANCEL` / `NOT_SAVE_ISCONFIRM` / `NOT_EDIT_BILLBALANCE` / `CANNOT_EDIT_REFERENCE` | :546-574 |
| 6 | ไม่มีรายการในกริด | `CANNOT_SAVE_NOTITEM` | :579-582 |
| 7 | แถวสินค้า (ประเภทสินค้า ≠ 1 และ ≠ 7) ไม่มี คลัง / ที่เก็บ / หน่วยนับ / จำนวน=0 | `NOT_SET_WHCODE` / `NOT_SET_SHELFCODE` / `NOT_SET_UNITCODE` / `QTY_ZERO_LIST` (มีสินค้าบางรายการที่กรอกจำนวนสินค้าเป็นศูนย์ — comment :625) | :587-632 |
| 8 | จำนวนเงินแถว ≠ จำนวน×ราคา−ส่วนลด (เกินเกณฑ์) | `LIST_AT n CHECK_WRONG_AMOUNT` (~"รายการที่ n จำนวนเงินไม่ถูกต้อง"); โหมด `m_bChmMode` ยอมต่างได้ไม่เกิน `Option.dDiffAmount` (ค่า <=0 ใช้ 1.0) ส่วนโหมดอื่นห้ามจำนวนเงิน=0 | :657-691 |
| 9 | ผลรวมจำนวนเงินแถวปกติ ≠ ช่อง "รวมมูลค่าสินค้า" | `CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT` ("ไม่สามารถบันทึกได้ เนื่องจากยอดรวมมูลค่าสินค้า ไม่เท่ากับยอดรายการย่อย" — comment :699) | :695-702 |
| 10 | (เอกสารใหม่) เลขที่ซ้ำ | ถาม "Document No is duplicated / Do you want to try again?" Yes → รันเลขใหม่ (`RunningAgain`) / No → ล้างเลข | :734-778 |

> ข้อสังเกต: การตรวจ "สั่งซื้อสินค้าซ้ำ" ถูก comment ทิ้งทั้งบล็อก (:704-730, "jum 14-02-2003 at BKK comment") — แต่ตอนพิมพ์ **จำนวน** ในกริดยังเรียก `OrderPODuplicate` (`Page1.cpp:1147`) เพื่อเตือนซ้ำรายเจ้าหนี้+สินค้า+หน่วย.

### 1.2 รายการ 5808 "ยกเลิกใบสั่งซื้อสินค้า" (ปิดยอดค้างส่งของ PO แบบรายบรรทัด)

- **ฟอร์ม**: `CPOFrmCancelPoSoView` (`POFrmCancelPoSoView.cpp`, 1,281 บรรทัด) dialog `IDD_DIALOG_PO_SO_CANCEL` 460x277 (`BC5Account.rc:390`). **ใช้ร่วมกับเมนูยกเลิกใบสั่งขาย** — ตัวแปร `m_nMode`: เมนู `ID_POSYS_TRANS_25` (5808) → `m_nMode=0` (PO) ไม่งั้น `=1` (SO) (`:26-33`); ตั้งชื่อหน้าต่างจาก key `20025` (PO) / `30020` (SO), ไฟล์ตั้งค่า `CancelPoConfig`, สิทธิ์ผู้ใช้ผูกกับ `ID_POSYS_TRANS_25` (`:34-38`).
- **ชนิดหน้าจอ**: เอกสารที่ "ดึงรายการค้างจากเงื่อนไข → ติ๊กเลือก → บันทึก" (ไม่พิมพ์รายการสินค้าเอง). **ไม่ลง GL**, ไม่ตัดสต็อก; ผลคือบันทึกเอกสารยกเลิก `BCPOVoid`/`BCPOVoidSub` แล้วคำนวณสต็อกใหม่ (`OnStockReprocess`) เพื่อให้ยอด "ค้างรับ" ของ PO ลดลง.

**A. หัวเอกสาร (ลำดับตามตำแหน่งบนจอ)**

| # | ป้ายบนจอ | ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่เอกสาร: | กล่องรันเลข (เลขของ `BCPOVoid`) (94,4) | ใช่ | รันเลขตามรูปแบบ+กลุ่มเอกสาร; ซ้ำ → `DOCNO_DUPLICATE` | `:121-128,1096-1107` |
| 2 | วันที่วางบิล: (ป้ายเดิมใน rc — ที่จริงคือ "วันที่เอกสาร") | วันที่ พ.ศ. (94,17) | ใช่ | ปิดช่องถ้าผู้ใช้ไม่มีสิทธิ์แก้วันที่เอกสาร (`IsCanEditDocDate=0` และไม่ใช่ `sa`); งวดปิดห้ามบันทึก | `:130-131,207-209,1111` |
| 3 | หมายเหตุ: | ข้อความ (259,4) | ไม่ | — | `:823` |
| 4 | SUM_OF_ITMEAMOUNT (ยอดรวม) | ตัวเลข อ่านอย่างเดียว (94,32) | — | ผลรวม `VoidAmount` ของแถวที่ติ๊ก | `:825,1152` |
| 5 | REASON (เหตุผลการยกเลิก) | Combo-table `BCCancelReason` (top 50) (259,32) | (ไม่บังคับในโค้ด) | บันทึกที่ `VoidReasonCode` | `:145,555` |

**B. เงื่อนไขดึงรายการ (กรอบกลางจอ) + ปุ่ม "Process" (F5_PROCESS_CONDIT)**

| ป้าย | ชนิด | ความหมาย |
|--|--|--|
| เลขที่เอกสาร: ... ถึง ... | Link ค้นหาเอกสาร PO (หรือ SO) | ช่วงเลขที่ใบสั่งซื้อ (กรอกอันเดียว = เท่ากับ) |
| FROM_CODE ... ถึง ... | Link เจ้าหนี้ (โหมด PO) / ลูกค้า (โหมด SO) | ช่วงรหัสเจ้าหนี้ `ApCode` |
| จากวันที่: ... ถึง ... | วันที่ | ช่วงวันที่เอกสาร; ค่าเริ่มต้นตั้งตามงวดบัญชีปัจจุบัน (`OnSetPeriod` อ่าน `BCPERIOD` :215-230) |
| FROM_SALE_CODE ... ถึง ... | Combo `BCSale` (พนักงานขาย) | ใช้เฉพาะโหมด SO (โหมด PO ปิดช่อง :152-153) |
| !SELECT_ALL (เลือกทั้งหมด) | checkbox | ติ๊ก/ถอนทุกแถว (`OnCheckSelectAll`) |

คำสั่งค้น (`OnBttnProcess` :404-537): `SELECT * FROM BCPurchaseOrderSub WHERE RemainQty > 0 AND DocNo NOT IN (SELECT DocNo FROM BCPurchaseOrder WHERE IsCancel = 1)` + เงื่อนไขช่วงเลขที่/วันที่/รหัสเจ้าหนี้ + `ORDER BY DocNo,ItemCode` → เติมกริดทีละแถว (วันครบกำหนด/ชื่อสินค้า/ชื่อหน่วยดึงจาก `BCPurchaseOrder.DueDate`, `BCItem.Name1`, `BCItemUnit.Name`).

**C. กริดรายการ (12 คอลัมน์ — `:169-183`)**

| # | ป้ายบนจอ (ไทยตรงในโค้ด) | ชนิด | ความหมาย |
|--|--|--|--|
| 0 | `C` | checkbox | เลือกบรรทัดที่จะยกเลิก (ค่า "TRUE"/"FALSE") |
| 1 | เลขที่เอกสาร | อ่านอย่างเดียว | เลขที่ PO |
| 2 | วันที่เอกสาร | อ่านอย่างเดียว | |
| 3 | วันที่ครบกำหนด | อ่านอย่างเดียว | `BCPurchaseOrder.DueDate` |
| 4 | รหัสสินค้า | อ่านอย่างเดียว | |
| 5 | ชื่อสินค้า | อ่านอย่างเดียว | |
| 6 | หน่วยนับ | อ่านอย่างเดียว | แสดง `รหัส/ชื่อหน่วย` |
| 7 | มูลค่า | ตัวเลข ชิดขวา | `Amount` ของบรรทัด PO |
| 8 | จำนวนที่สั่งซื้อ (โหมด SO: จำนวนที่สั่งขาย) | ตัวเลข | `Qty` |
| 9 | จำนวนที่ยกเลิก | **แก้ได้** | ค่าตั้งต้น = `RemainQty` (ยอดค้าง) ผู้ใช้ลดได้ → ยกเลิกบางส่วน; `VoidAmount = ราคา × จำนวนที่ยกเลิก` |
| 10 | `A` | checkbox | "ยกเลิกทั้งหมด" → บันทึก `CancelAll = 1` |

Shadow (บันทึกลงตาราง): ApCode (หรือ ArCode), DepartCode, AllocateCode, ProjectCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, Price, WHCode, ShelfCode, RefLineNumber (=`LineNumber` ของแถว PO ต้นทาง), UnitCode, BehindIndex, QtyBalance, VoidAmount (`:188-205`). คีย์: ลบแถวได้ (`SetDeleteSelectRow(0)`).

**D. การบันทึก (`OnSaveData` :539-810)**

1. เอกสารใหม่ต้องติ๊กอย่างน้อย 1 แถว (`OnCheckDetail`) ไม่งั้น "Please select transaction for cancel" (`:545-548`, ข้อความอังกฤษตรงในโค้ด).
2. หัว: `InsertDB`/`UpdateDB` `BCPOVoid` (หรือ `BCSOVoid`): DocNo, DocDate, TotalAmount, Remark, VoidReasonCode, IsCompleteSave(=0 ก่อน), CreatorCode, CreateDateTime (+LastEditorCode, LastEditDateT เมื่อแก้ไข). `:551-585`.
3. แก้ไข: `DELETE FROM BCPOVoidSub WHERE DocNo=…` แล้วเขียนใหม่; ก่อนหน้านั้นคืนสถานะ PO เดิมด้วย `UpdateDocumentFlow(-1, 2, เลขที่ PO)` (`:596-621`).
4. รายการ (เฉพาะแถวที่ติ๊ก): `BCPOVoidSub` = DocNo, DocDate, Depart/Allocate/Project/Category/Group/Brand/Type/Format, Price, WHCode, ShelfCode, RefLineNumber, UnitCode, ApCode, **PONo, PODate, PODueDate**, ItemCode, MyDescription(=ชื่อสินค้า), Amount, Qty, **VoidQty, VoidAmount**, CancelAll, TaxRate, PackingRate1/2, behindindex, LineNumber, IsProcess (`(Option.IsBehindIndex==0 || แก้ไข) ? 1 : 0`) — แต่ละแถวที่ insert สำเร็จเรียก `OnStockReprocess(รหัสสินค้า, วันที่)` ทันที (`:660-762`).
5. ครบ → `UPDATE BCPOVoid SET IsCompleteSave = 1`; ไม่สำเร็จ → "บันทึกรายการย่อยไม่สมบูรณ์" (`SAVEDETAIL_NOTCOMPLETE`); log book `AddUserLogBook(..., IDD_DIALOG_PO_SO_CANCEL, ...)`. (การเรียก `m_ProcJob.PurchaseOrderVoidProcess` ถูก comment ทิ้ง `:777-781`.)
6. ถ้าเอกสารถูกยืนยันแล้ว: "ไม่สามารถแก้ไขได้ เพราะข้อมูลรายการนี้ได้ Confirm แล้ว" (`CANNOT_UPDATE_ISCONFIRM` comment `:573`).

**E. ลบเอกสารยกเลิก** (`ConfirmDeleteData` :1120-1137 / `OnDeleteData` :992-1060): ต้องไม่ว่างเลขที่ + งวดไม่ปิด + ยืนยัน Yes/No → `DELETE FROM BCPOVoid` + `DELETE FROM BCPOVoidSub` → `OnStockReprocess` ทุกสินค้า → `UpdateDocumentFlow(-1,2,PO)` → log "ลบ". ผลคือยอดค้างของ PO กลับมา.

**F. ผลข้างเคียง**: ไม่ลงบัญชี · PO ต้นทางไม่ถูก `UPDATE` ตรงในไฟล์นี้ แต่ **`UpdateDocumentFlow(1|-1, 2, เลขที่ PO)`** (เรียกตอนบันทึก/ลบ) คำนวณ `BCPurchaseOrder.BillStatus/PoStatus` ใหม่โดยนับ `BCPOVoidSub.VoidQty` เป็น "จำนวนที่ถูกใช้ไปแล้ว" รวมกับจำนวนที่ถูกอ้างโดยใบซื้อ/GR/IR (`mytools.cpp:1721-1753`): ผลรวมสั่ง ≤ ผลรวมใช้+ยกเลิก → ปิดเป็น "รับครบ" (`PoStatus=2, BillStatus=1`) · มีบางส่วน → `PoStatus=3` · ไม่มี → `PoStatus=0`. ส่วน `RemainQty` รายบรรทัดลดผ่านคิว `OnStockReprocess` (กลไกใน Process engine ยังไม่ verify). ฟอร์มรับสินค้าดึงเฉพาะบรรทัด `RemainQty>0` จึงไม่เห็นบรรทัดที่ยกเลิกแล้ว.

### 1.3 รายการ 5798 "บันทึกรับสินค้าจากการซื้อ" (ใบรับสินค้า — GR)

- **ฟอร์ม**: `CChmFrmGoodsReceive` (`ChmFrmGoodsReceive.cpp`, 2,793 บรรทัด; ใช้งานจริงตาม `MainFrm.cpp:3127`) + แท็บ `CChmFrmGoodsReceivePage1` (`IDD_CHM_PO_FRM_GR1` 610x300: หัว+กริด) และ `CChmFrmGoodsReceivePage2` (`IDD_CHM_PO_FRM_GR2` 538x257: ช่องรอง). (มีไฟล์ `POFrmGoodsReceive.cpp` รุ่นเก่า — ไม่ถูกเรียกจากเมนูนี้.)
- **ตารางเก็บ**: ใช้ตารางเดียวกับใบซื้อ `BCAPInvoice` / `BCAPInvoiceSub` โดยตั้ง `GRIRBillStatus = 1` (GR) — (0 = ซื้อปกติ 5812, 2 = ใบตั้งหนี้ IR 5799) (`ChmFrmGoodsReceive.cpp:1868`, key เขียนตั้ง `GRIRBillStatus` ใน `OnSaveData`).
- **เป้าหมาย**: รับสินค้าเข้าคลัง **ก่อนมีใบกำกับภาษี** — ตัดยอด PO (รับตามใบสั่งซื้อ) → เพิ่มสต็อก → ลง GL (เดบิตสินค้า/เครดิตเจ้าหนี้รอตั้งหนี้ตามรูปแบบ GL กลุ่ม 29) แล้วค่อยออก "ใบตั้งหนี้" (5799) อ้าง GR นี้ภายหลัง. อีกทางหนึ่งคือ "ตั้งหนี้ก่อน รับของทีหลัง": เลือกใบตั้งหนี้เข้ามาที่ GR (ปุ่ม `เลือกใบตั้งหนี้`).

**A. หัวเอกสาร แท็บ "ข้อมูลหลัก" (Page1) — ลำดับตามตำแหน่งบนจอ**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบรับสินค้า: | EDIT + ปุ่มรันเลข `m_edDocNo` (128,8) | ใช่ | รันเลขตามรูปแบบรายวัน; แก้ได้เฉพาะผู้มีสิทธิ์ `bCanChangeBillNo`; ตรวจซ้ำก่อนบันทึก; เกินความยาว → `DOCNO_MORE_LENG` | `Page1.cpp:188-199,721-740`; `GoodsReceive.cpp:1689-1692` |
| 2 | วันที่ออกเอกสาร: | วันที่ พ.ศ. `BME_DATE_TH_AUTO` + ปุ่มปฏิทิน (332,8) | ใช่ | = วันทำงาน `m_strWorkingDate`; ปิดช่องถ้าไม่มีสิทธิ์แก้วันที่ (`GoodsReceive.cpp:194`) | `Page1.cpp:191-192,1436` |
| 3 | (ปุ่ม) เลือกใบสั่งซื้อสินค้า (key `PO_POTRANS`) | ปุ่ม (501,9) | — | ดึงรายการค้างรับจาก PO — ต้องกรอกเจ้าหนี้ก่อน (`PLEASEINSERTDATAINAPCODE` "กรุณากรอกข้อมูลที่รหัสเจ้าหนี้ก่อน") | `Page1.cpp:274-291` |
| 4 | รหัสเจ้าหนี้: | Link ตรวจกับ `BCAP.Code` + ปุ่มเลือก (128,23) + ช่องชื่อ อ่านอย่างเดียว (222,23) | ใช่ | ออกจากช่อง → ดึงชื่อ (`Name1` หรือ `Name2` ตาม `CompanyInfo.DefaultApName`), อัตราภาษีตั้งต้น `BCAP.DefaultTaxRate`, ส่วนลดเจ้าหนี้ `Discount`, ส่วนลดต่อรายการ `DetailDiscount`, ผู้ติดต่อ `DefContactCode`, เครดิต `CreditDay` → วันถึงกำหนด; ช่องล็อกหลังดึงจาก PO/IR | `Page1.cpp:194-196,656-717` |
| 5 | (ปุ่ม) เลือกใบตั้งหนี้ (key `SELECT_IR`) | ปุ่ม (501,33) | — | ดึงใบตั้งหนี้ที่ยังไม่ถูกรับของ (กรณีตั้งหนี้ก่อน) | `Page1.cpp:479-656` |
| 6 | ประเภทภาษี: | Combo 3 ค่า (128,38) | ใช่ | `CompanyInfo.DefaultTaxType` (0 แยกนอก / 1 รวมใน / 2 ภาษีอัตราศูนย์); ตั้งใหม่ตามเอกสารที่ดึงมา (ล็อกเมื่อ `m_bChmMode`) | `Page1.cpp:183-184,342,517` |

**B. แท็บ "รายการอื่นๆ" (Page2) — ลำดับตามตำแหน่งบนจอ (x,y)**

| # | ป้ายบนจอ | ชนิด | กติกา | บันทึกที่คอลัมน์ | หลักฐาน |
|--|--|--|--|--|--|
| 1 | รหัสแผนก: | Combo-table `BCDepartment` (101,23) | — | `DepartCode` | `Page2.cpp:68` |
| 2 | กลุ่มเอกสาร: | ข้อความ อ่านอย่างเดียว (382,23) | รับจากกลุ่มเอกสารที่ตั้ง/ดึง | `BillGroup` | rc / `GoodsReceive.cpp` |
| 3 | โครงการ: | Combo-table `BCProject` (101,38) | — | `ProjectCode` | `Page2.cpp:69` |
| 4 | รหัสสกุลเงิน: | Combo-table `BCCurrency` (382,38) | ปิดช่องเมื่อไม่ใช้หลายสกุลเงิน (`MultiCurrency=0`); เลือกแล้วดึงอัตราจาก `BCCurrency.ExchangeRate` แล้วคำนวณ HomeAmount ทุกแถวใหม่ | `CurrencyCode` | `Page2.cpp:63-66,100-124,182-200` |
| 5 | รหัสผู้ติดต่อ: | Combo-table `BCContactList` (101,53) | ค่าตั้งต้นจาก `BCAP.DefContactCode` | `ContactCode` | `Page2.cpp:71`; `Page1.cpp:688-689` |
| 6 | อัตราแลกเปลี่ยน: | ตัวเลข 9 ทศนิยม (382,53) | ปิดช่องเมื่อไม่ใช้หลายสกุลเงิน; แก้แล้วคำนวณ HomeAmount (shadow 23/41) ทุกแถว; <=0 → 1.0 | `ExchangeRate` | `Page2.cpp:62,158-180` |
| 7 | การจัดสรร: | Combo-table `BCAllocate` (101,68) | — | `AllocateCode` | `Page2.cpp:70` |
| 8 | อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข (ทศนิยม `AmountPoint`) (382,68) | ตั้งต้นจากเจ้าหนี้/PO/IR | `TaxRate` | `Page2.cpp:61` |
| 9 | เครดิต(วัน): | ตัวเลข (101,83) | แก้ → วันถึงกำหนด = วันที่เอกสาร + เครดิต | `CreditDay` | `Page2.cpp:126-138` |
| 10 | วันที่ถึงกำหนด: | วันที่ + ปุ่มปฏิทิน (382,83) | แก้ → เครดิต = (วันถึงกำหนด − วันที่เอกสาร) − 1 (พฤติกรรมจริงในโค้ด; ยังไม่ verify ว่าตั้งใจ) | `DueDate` | `Page2.cpp:140-155` |
| 11 | ผู้ขนส่ง/บริษัทขนส่ง: | Link `BCAP.Code` + ปุ่ม (101,98) | เลือกจากทะเบียนเจ้าหนี้ | `OwnerTransport` | `Page2.cpp:59-60` |
| 12 | หมายเหตุ: | ข้อความ (101,140) | — | `MyDescription` | rc / `GoodsReceive.cpp` |

**C. กริดรายการสินค้า (10 คอลัมน์ที่เห็น) — `Page1.cpp:77-94`** (กริด `ID_GR_PURCHASE`: แถบซ้ายเลขแถวอัตโนมัติ ลากปรับความกว้างได้; ปุ่มท้ายแถว (`SetOtherCommandIcon IDI_RIGHT`) = ขยายแถว; ลบแถวด้วย `SetDeleteSelectRow(0)`)

| # | ป้าย (key) | กว้าง/flag | ความหมาย / พฤติกรรม |
|--|--|--|--|
| 0 | รหัสสินค้า (`ITEMCODE`) | 90 แก้ได้ + Link | พิมพ์รหัส (ตัวพิมพ์ใหญ่อัตโนมัติ `Page1.cpp:1137`); ไม่พบ → `NOT_SEEITEMCODE`; `.` = รายการข้อความ; ถ้า `CompanyInfo.GRNoAddnew=1` ห้ามเพิ่มแถวเอง (ล็อกคอลัมน์ 0 และปิดเพิ่มแถว — ต้องดึงจาก PO/IR เท่านั้น) `Page1.cpp:322-325` |
| 1 | ชื่อสินค้า (`ITEMNAME`) | 120 อ่านอย่างเดียว | ชื่อสินค้า (รายการ `.` ใช้ `MyDescription`) |
| 2 | คลังสินค้า (`WAREHOUSE_ITEM`) | 60 Combo-table | `BCItemWarehouse` join `BCWarehouse` (Top 50) กรองตามสินค้า (`Page1.cpp:155-165,1253-1260`) |
| 3 | ที่เก็บ (`SHELF`) | 60 Combo-table | `BCItemWarehouse` join `BCShelf`; ไม่กำหนดที่เก็บ → `NOT_SET_SHELFCODE` (`Page1.cpp:1156-1166`) |
| 4 | หน่วยนับ (`UINTCODE`) | 60 Combo-table | `BCItemUnit` (สินค้าหน่วยแพ็ก ใช้ `BCStkPacking`); ไม่ใช่หน่วยของสินค้า → `NOUNIT` (`Page1.cpp:1168-1184,1281-1292`) |
| 5 | จำนวน (`ITEMQTY`) | 60 ตัวเลข `QtyPoint` ชิดขวา | **ถ้าแถวมาจาก PO: ห้ามเกินจำนวนค้างรับ** — เกิน → `QTY_RECEIVE_MUCH_ORDER` แล้วคืนค่าเดิม และตั้งธง `bQtyOverPo` (ห้ามบันทึก); ถ้าดึงจาก IR คอลัมน์นี้ล็อกแก้ไม่ได้ (`Page1.cpp:505-506,1085-1110`) |
| 6 | ราคาซื้อ/หน่วย (`PRICE_UNIT`) | 60 ตัวเลข `PricePoint` | **แก้ได้เฉพาะผู้ที่ `bCanSeeItemCost`** (ไม่มีสิทธิ์ → คอลัมน์ 6-8 ว่าง/อ่านอย่างเดียว และค่าจริงเก็บใน shadow 27,28,40,41); ล็อกถ้า `CompanyInfo.GRNoEditPrice=1` หรือผู้ใช้ไม่มี `bCanChangeGRBuyPrice` เมื่อดึงจาก PO/IR; ถ้าว่าง/0 ดึงราคาซื้อล่าสุด `GetItemBuyPrice(เจ้าหนี้,สินค้า,หน่วย,วันที่)`; เปิด `Option.BuyPriceWarning=1` → ราคาซื้อ (หลังหักส่วนลด) สูงกว่าราคาขาย → `BUYPRICE_MORE_SALEPRICE` แล้วล้างช่อง (`Page1.cpp:83-93,1186-1233`) |
| 7 | ส่วนลด%,บาท (`DISCOUNT`) | 60 ข้อความ | รูปแบบเดียวกับ PO (`5%`, `10+3%`, `50`); ล็อกเมื่อดึงจาก PO/IR (`Page1.cpp:321,510`) |
| 8 | จำนวนเงิน (`AMOUNT`) | 70 ตัวเลข `AmountPoint` | คำนวณจำนวน × ราคา หักส่วนลด (`DetailCalc`) |
| 9 | เลขที่เอกสาร PO/IR (`DOCNO_PO_IR`) | 90 อ่านอย่างเดียว | แสดงเลขที่ PO หรือ IR ต้นทางของแถว (`Page1.cpp:94,416`) |

**Shadow columns (44 คอลัมน์ `Page1.cpp:100-150`)**: 0 RowOrder · 1 ItemType · 2 MyDescription · 3 ProjectCode · 4 AllocateCode · 5 OppositeQty · 6 OppositeUnit · 7 ExceptTax · 8 BarCode · 9 CategoryCode · 10 GroupCode · 11 BrandCode · 12 TypeCode · 13 FormatCode · 14 WHCode · 15 DepartCode · 16 CurrencyCode · 17 ExchangeRate · **18 PORefNo** · **19 OldQty** (ยอดค้างรับเดิม = เพดานจำนวนรับ) · 20 Discount · 21 Amount · 22 NetAmount · 23 HomeAmount · 24 LotScript · 25 DiscountAmount · 26 LineNumber · 27 MyPrice · 28 MyDiscount · **29 POLinenumber** (หมายเลขบรรทัดของ PO/IR ต้นทาง) · **30 IRRefNo** · 31 StatusReceive · 32 LotNumber · 33 LotExpireDate · 34 LotMyDesc · 35 OldWHCode · 36 OldShelfCode (ตรวจ Serial) · 37 ColorCode · 38 SizeCode · 39 StyleCode · 40 MyNetAmount · 41 MyHomeAmount · 42 BehindIndex · 43 OldItemCode.

**D. ยอดรวม**: **ใบรับสินค้าไม่มีช่อง "รวมมูลค่า/ภาษี/สุทธิ" บนจอ** (Page1 ไม่มีช่องยอดรวม — rc `IDD_CHM_PO_FRM_GR1` มีเฉพาะกริด + ข้อความรายละเอียดสินค้า `IDC_LB_ITEMDESC`) — `BillCalc` (`GoodsReceive.cpp:309-327`) คำนวณภายในเท่านั้น: `m_dNetAmount` = ผลรวมจำนวนเงินหลังหักส่วนลดทุกแถว; `m_dSumOfZeroTax` = ผลรวมแถว ExceptTax=2 เมื่อ `CompanyInfo.ZeroTaxRate=1`. ตอนบันทึกจึงคำนวณหัวเอกสารใหม่ทั้งหมดใน `OnSaveData` (`GoodsReceive.cpp:1780-1845`):
  - รวมแถวปกติ → `SumOfItemAmount`; แถวยกเว้นภาษี → `ExceptTaxAmount`; แถว ExceptTax=2 → `ZeroTaxAmount`
  - ภาษีรวมใน (`TaxType=1`): `BeforeTaxAmount = ฐานก่อนภาษีหลังถอด VAT`, `TaxAmount = SumOfItem − BeforeTax`; ภาษีแยกนอก (0/2): `TaxAmount = ปัด(BeforeTax × TaxRate / 100, AmountPoint)`
  - `TotalAmount = BeforeTax + TaxAmount`; `NetDebtAmount = BeforeTax + ExceptTax + TaxAmount`; `BillBalance = NetDebtAmount`; `HomeAmount = ปัด(NetDebtAmount × ExchangeRate)` (`:1805-1845`)

**E. ปุ่ม / คีย์ลัด** — ปุ่มแถบเครื่องมือมาตรฐาน (§0.2) + ผูกจริง (`GoodsReceive.cpp:70-100`): F2 `ID_HK_GL` ดูรายการบัญชี · F12 `ID_HK_DOCFLOW` ตรวจเส้นทางเอกสาร (`OnHkDocumentFlow` ดูผู้อ้างอิง `BCIRSUB`/`BCSTKREFUNDSUB` `:770,2486`) · **Ctrl+F3 `ID_HK_GETSO` = เลือกใบสั่งซื้อ (`OnHkGetPo`)** · **Shift+F3 `ID_HK_GETIR` = เลือกใบตั้งหนี้ (`OnHkGetIr`)** · Ctrl+F5 ยกเลิกเอกสาร · Shift+F5 ยืนยันเอกสาร · Ctrl+F7 `ID_HK_SETDEFWH` ตั้งคลังเริ่มต้น · Ctrl+F8 `ID_HK_SETDEFSH` ตั้งที่เก็บเริ่มต้น (ไม่ได้ตั้ง → `DEFAULT_WH`) · Ctrl+F9 `ID_HK_ITEMSERIAL` Serial (สินค้าไม่มี Serial → `ITEM_NOT_HAVE_SERAIL`; ถ้า `SerialItem=0` ไม่เปิด) · Shift+F9 Lot (`CANNOT_SET_LOT`) · Ctrl+F12 `ID_HK_ITEMINFO` ประวัติซื้อสินค้า (ต้องเลือกแถวสินค้า `SET_CURSER_ITEM` "กรุณาเลือกสินค้าที่ต้องการดูรายละเอียด") · Shift+F7 `ID_HK_ITEMBYPO` (สินค้าตาม PO; `OnSelectItemByPO` ต้องเลือกเจ้าหนี้+PO ก่อน `BEFORE_ARCODE_PO`) · Shift+F2 ขยายแถว · F5 รายละเอียดเพิ่มเติม · F11 บันทึก · PgUp/PgDn เอกสารก่อน/ถัดไป · Ctrl+N ใหม่ · Ctrl+F ค้นหา · Ctrl+P พิมพ์.

**ปุ่ม "เลือกใบสั่งซื้อสินค้า"** (`OnBttnPurchaseOrder` `Page1.cpp:274-477`): เปิดหน้าต่างเลือกหลายใบ (`CMyMultiSelect`, config `POConfig`) กรอง PO ของเจ้าหนี้นี้ที่: ไม่ยกเลิก (`IsCancel=0`), `BillStatus=0`, ไม่หมดอายุ (`ExpireDate >= วันที่เอกสาร หรือว่าง`), `TaxType` ตรงกับเอกสารนี้, ยังมีบรรทัดค้างรับ (`sum(RemainQty)>0`), และไม่ใช่ PO ที่ดึงมาแล้วในกริดนี้ → ล็อกช่องเจ้าหนี้; PO ใบแรกที่เลือกกำหนดค่าหัว: ประเภทภาษี, อัตราภาษี, เครดิต/วันถึงกำหนด, ผู้ติดต่อ, สกุลเงิน/อัตรา, แผนก, โครงการ, การจัดสรร, กลุ่มเอกสาร, หมายเหตุ, ส่วนลดท้ายบิล; แต่ละบรรทัดค้างรับ (`BCPurchaseOrderSub` ที่ `TransState=0 AND IsCancel=0 AND (RemainQty>0 OR ItemCode='.')` เรียงตาม `LineNumber,ItemCode,UnitCode`) ลงกริด: **จำนวน = RemainQty**; ถ้ายอดค้างน้อยกว่าที่สั่ง จำนวนเงิน = ยอดค้าง × ราคา หักส่วนลดใหม่; ภาษีรวมใน → ถอด VAT ออกก่อนเก็บ `NetAmount`; เก็บ `PORefNo`, `POLinenumber`, `OldQty` (=ยอดค้างเดิม). 

**ปุ่ม "เลือกใบตั้งหนี้"** (`OnBttnInvoiceReceive` `Page1.cpp:479-653`): เลือกใบตั้งหนี้ IR ของเจ้าหนี้ (`BCAPInvoice.GRIRBillStatus=2 AND GRBillStatus=0 AND IsCompleteSave=1 AND IsCancel=0 AND TaxType` ตรงกัน; config `InvoiceReceiveConfig`) → **ล้างกริดแล้วโหลดทับ**; ล็อกจำนวน/หน่วย/รหัสสินค้า (รับได้เต็มตามที่ตั้งหนี้) และราคา/ส่วนลดตามสิทธิ์; บรรทัดจาก `BCIRSUB WHERE TransState=0 AND IsCancel=0 AND CNQty>0` เรียงตาม `GRRefNo,LineNumber,ItemCode,UnitCode`; ตั้ง `m_bGetIRData` (บันทึกแล้ว `GRBillStatus=2`).

**F. กติกา/การตรวจก่อนบันทึก (`ConfirmSaveData` `GoodsReceive.cpp:1410-1698` ตามลำดับ)**

| # | เงื่อนไขที่ไม่ผ่าน | key / ข้อความ | หลักฐาน |
|--|--|--|--|
| 1 | มีแถวที่จำนวนรับเกินยอดค้าง PO (`bQtyOverPo`) | ไม่ให้ผ่านทันที | :1413 |
| 2 | วันที่อยู่ในงวดปิด | จาก `IsPeriodClose` | :1417 |
| 3 | (แก้ไข) เอกสารยกเลิกแล้ว / ยืนยันแล้ว | `NOT_DELETE_ISCANCEL` / `NOT_DELETE_ISCONFIRM` | :1423-1430 |
| 4 | เลขที่เอกสารว่าง / (แก้เลขที่) ซ้ำเอกสารอื่น | `CANNOT_SAVE_NOTDOCNO` ("ไม่สามารถบันทึกได้ เนื่องจากไม่ใส่เลขที่เอกสาร") / `DOCNO_DUPLICATE` | :1438-1448 |
| 5 | วันที่ว่าง | `CANNOT_SAVE_NOTDOCDATE` | :1450 |
| 6 | รหัสเจ้าหนี้ว่าง / ไม่พบใน `BCAP` | `CANNOT_SAVE_NOTAPCODE` ("…ไม่ใส่รหัสเจ้าหนี้") / `CANNOT_SAVE_NOTAPCODE_INDB` ("ไม่มีข้อมูลเจ้าหนี้ในฐานข้อมูล") | :1455-1466 |
| 7 | (แก้ไข) GR ถูกนำไปตั้งหนี้หมดแล้ว (`GRBillStatus=1`) หรือบางบรรทัดมี `GRRemainQty` ≠ `Qty` ผสมกับบรรทัดที่ยังไม่ใช้ | `NOT_EDIT_BILLBALANCE` ("ไม่สามารถแก้ไขรายการได้ เนื่องจากเอกสารได้นำไปตัดยอดแล้ว") | :1469-1491 |
| 8 | เอกสารลงบัญชี GL แล้ว (`IsPostGL`) | `CANNOT_SAVE_POSTGL` | :1494 |
| 9 | ไม่มีรายการในกริด | `CANNOT_SAVE_NOTITEM` | :1499 |
| 10 | แถวสินค้า (ประเภท ≠ 1 บริการ, ≠ 7) ไม่มีคลัง / ที่เก็บ / หน่วย | `NOT_SET_WHCODE` / `NOT_SET_SHELFCODE` / `NOT_SET_UNITCODE` (+รหัสสินค้า) | :1521-1537 |
| 11 | จำนวน = 0 | `QTY_ZERO_LIST` ("มีสินค้าบางรายการที่กรอกจำนวนสินค้าเป็นศูนย์") | :1540 |
| 12 | สินค้า Serial (`SerialItem=1`, ประเภท 2): แก้เอกสารที่ Serial เปลี่ยนสถานะแล้ว / จำนวน Serial ไม่เท่าจำนวนรับ / คลังที่เก็บ Serial เปลี่ยน | `NOTEDIT_SERIALSTATUS` / (ข้อความ "รหัสสินค้า … รายการที่ n …" จาก `OnCheckAddSerial`, `OnEqualQtySerial`) / `CANNOT_DELETE_SERIALWH` | :1547-1586 |
| 13 | สินค้าต้นทุน FIFO + ใช้ Lot (`nUseLotStatus=1 และ nLotStatus=1`) ไม่ระบุหมายเลข Lot ครบ (ตัวเลือก Lot สูงสุด 3 ชุด) | ข้อความ "รหัสสินค้า … รายการที่ n …" | :1617-1648 |
| 14 | (เอกสารใหม่) เลขที่ซ้ำ | ถาม "Document No is duplicated / Do you want to try again?" Yes → รันเลขใหม่ | :1654-1697 |

> บล็อก "จำนวนรับเกิน PO" ใน `ConfirmSaveData` (:1591-1612) ถูก comment ทิ้ง — ใช้การตรวจที่ช่องกริดแทน (ข้อ 1).

**G. ลบ / ยกเลิก / ยืนยัน**

| การกระทำ | กติกา | หลักฐาน |
|--|--|--|
| ลบ (`ConfirmDeleteData`) | ห้ามถ้ายกเลิก (`NOT_SAVE_ISCANCEL`) / ยืนยัน (`NOT_SAVE_ISCONFIRM`) แล้ว; เลขที่ว่าง (`CANNOT_DELETE_DOCNO`); งวดปิด; ลง GL แล้ว (`CANNOT_DELETE_POSGL`); GR ถูกตั้งหนี้หมดแล้ว (`GRBillStatus=1`) หรือถูกตั้งหนี้บางส่วน (`GRRemainQty` ≠ `Qty` บางบรรทัด) → `NOT_DELETE_BILLBALANCE` ("ไม่สามารถลบรายการนี้ได้ เนื่องจากเอกสารได้นำไปตัดยอดแล้ว"); สินค้า Serial ที่เปลี่ยนสถานะ/คลังแล้ว (`CANNOT_DELETE_SERIALSTATUS`/`CANNOT_DELETE_SERIALWH`); ถามยืนยัน `DELETE_MSG` Yes/No | `GoodsReceive.cpp:1332-1408` |
| ลบจริง (`OnDeleteData`) | คืน `BCPurchaseOrderSub.IsProcess=0` → `DELETE BCAPInvoice (GRIRBillStatus=1)` → `UpdateGRStatus(-1)` คืน `BCIRSUB.CNQty` + สถานะ IR → `DELETE BCAPInvoiceSub` → `OnStockReprocess` ทุกสินค้า → `DeleteAccount` (GL source 2) → ลบ Serial/Lot → logbook "SUBTRACT_RECEIPT" → `UpdateDocumentFlow(-1,2,PO...)` | `:1701-1749` |
| ยกเลิกเอกสาร (Ctrl+F5 → `OnConfirmCancel` ตรวจก่อน) | ห้ามถ้ายกเลิก/ยืนยันแล้ว, ลง GL แล้ว, ถูกตั้งหนี้หมด; ถ้ามีเอกสารอ้างอิง (`BCInvStkRefund.InvoiceNo` ที่ `BillType=0`, หรือ `BCIRSUB.GRRefNo`) ที่ไม่ยกเลิก → `CANNOT_CANCEL_USE` ("ไม่สามารถลบได้เนื่องจากได้ผ่านรายการย่อยแล้ว") | `:1221-1254` |
| ผลของการยกเลิก (`BillCancelUpdateData`) | `BCPurchaseOrderSub.IsProcess=0`, `UpdateGRStatus(-1)`, `UPDATE BCAPInvoice SET BillBalance=0, IsCancel=1 WHERE GRIRBillStatus=1`, `BCAPInvoiceSub.IsCancel=1`, `OnStockReprocess`, `ReverseAccount` (กลับรายการ GL source 2), ลบ Serial/Lot, logbook "CANCEL_RECEIPT", `UpdateDocumentFlow(-1,2,…)` | `:1262-1309` |
| ยืนยัน (Shift+F5) | `OnHkConfirmDoc` — ไม่มีสิทธิ์ `bCanConfirmBill` → `NOT_CONFIRM_NO` (กลไกกลาง `00-common-ui.md` §10.3) | `:2555` |

**H. สถานะเอกสาร**

| คอลัมน์ | ค่า | ความหมาย |
|--|--|--|
| `BCAPInvoice.GRIRBillStatus` | 1 | = ใบรับสินค้า (GR) |
| `BCAPInvoice.GRBillStatus` (ของ GR) | 0 / 1 / 2 | 0 = ยังไม่ถูกนำไปตั้งหนี้ (ตั้งตอนบันทึก `:2060`); 1 = ถูกนำไปตั้งหนี้หมดแล้ว (ตั้งโดย 5799); **2 = GR นี้ถูกสร้างโดยดึงใบตั้งหนี้มาก่อน** (`m_bGetIRData` `:2061-2062`) |
| `BCAPInvoiceSub.GRRemainQty` | ตัวเลข | จำนวนที่ยังไม่ถูกตั้งหนี้ (ตั้งต้น = จำนวนรับ) — ใช้เป็นตัวกันแก้/ลบ |
| `BCIRSUB.CNQty` (ของ IR) | ตัวเลข | จำนวนของใบตั้งหนี้ที่ **ยังไม่ได้รับของ**; GR ที่ดึง IR มาจะลด (`UpdateGRStatus +1`) และ IR header `GRBillStatus` = 1 เมื่อผลรวม `CNQty ≤ 0` ไม่งั้น 0 (`:463-518`) |
| `IsConfirm` / `IsCancel` / `IsCompleteSave` | 0/1 | เหมือนเอกสารอื่น (`IsCompleteSave` ตั้ง 1 หลังรายการย่อยครบ `:2060`) |
| `GLTransData` / `GLFormat` / `GLStartPosting` | | `m_nManualKey`; รูปแบบ GL; ตั้ง 0 ตอนบันทึก (`:1845`) |

**I. ผลข้างเคียงตอนบันทึก (`OnSaveData` `GoodsReceive.cpp:1752-2128`)**

ลำดับ: `UpdateData` → `ClearFieldDataTemp` → คำนวณหัวเอกสาร (ข้อ D) → `StartTransaction` → (เอกสารใหม่) `InsertDB(BCAPInvoice)` / (แก้ไข) `UpdateGRStatus(เลขเดิม,-1)` คืนสถานะ IR + `BCPurchaseOrderSub.IsProcess=0` ของ PO เดิม + `UpdateDB` → ลบรายการย่อยเดิม (`DELETE BCAPInvoiceSub WHERE DocNo`) + `UpdateDocumentFlow(-1,2,PO เดิม)` → เขียนรายการย่อยใหม่ทุกแถวที่รหัสสินค้าไม่ว่าง (+ `UPDATE BCIRSUB SET GRRefNo = เลขที่ GR` ของบรรทัด IR ที่ดึงมา `:2048`) → ครบ: `UPDATE BCAPInvoice SET IsCompleteSave=1, GRBillStatus=0|2`, `BCPurchaseOrderSub.IsProcess=0` → `SetDTSModifyFlag` → `Commit` → `OnStockReprocess(สินค้า, วันที่)` ทุกตัว → **`INSERT INTO processdoc (DocNo, DocType=2)`** (คิวประมวลผล) → บันทึก Serial (`OnSaveSerial(2,…)`) และ Lot (`OnSaveLotItem(2,…)`) → GL: `m_nManualKey == 0` → `m_GLTLSProcess` (`m_nSource=2`, `m_nGLGroup=1` เชื่อ, `m_nDocStatus=1`) ประมวลผลอัตโนมัติ มิฉะนั้น `OnSaveAccount(…,2,…)` เปิดหน้าจอลงบัญชีมือ → `UpdateDocumentFlow(1,2,PO ที่อ้าง)` (คำนวณ `RemainQty/PoStatus` ของ PO ใหม่) → `UpdateGRStatus(เลขที่,+1)` (ตัด `CNQty` ของ IR) → logbook "SAVE_RECEIPT". ถ้ารายการย่อยไม่ครบ → `Abort()` + `SAVEDETAIL_NOTCOMPLETE` ("บันทึกรายการย่อยไม่สมบูรณ์").
  - **รูปแบบ GL ตั้งต้น = กลุ่ม 29** (`GetGLFormat` `:2468-2471` → `GetDefaultGLFormat(29)` เลือก `BCGLFormat.Code` ที่ `FormatGroup=29` เรียง `IsDefault DESC` `mytools.cpp:32-48`) ตรงกับ key `GLFOR_CRPURCHASE_GR` (§0.4).

**J. คอลัมน์ที่เขียน**

- `BCAPInvoice` (หัว, `:1868`): DocNo, DocDate, BillType, **GRIRBillStatus=1**, TaxType, ApCode, MyDescription, CreditDay, DueDate, ContactCode, DepartCode, ProjectCode, AllocateCode, BillGroup, RecurName, DiscountWord, ZeroTaxAmount, OwnerTransport, SumOfItemAmount, AfterDiscount, NetDebtAmount, **BillBalance**, BeforeTaxAmount, TaxAmount, TotalAmount, TaxRate, ExceptTaxAmount, CurrencyCode, ExchangeRate, HomeAmount, GLTransData, GLFormat, GLStartPosting, IsCompleteSave, CreatorCode, CreateDateTime (+LastEditorCode, LastEditDateT เมื่อแก้).
- `BCAPInvoiceSub` (รายการ, `:2051`): DocNo, TaxType, DocDate, ApCode, ItemCode, ItemName, WHCode, ShelfCode, UnitCode, Qty, Price, DiscountWord, Amount, CNQty, **GRRemainQty**, BalanceAmount, DiscountAmount, StockType, MyDescription, ProjectCode, AllocateCode, OppositeQty, OppositeUnit, ExceptTax, BarCode, Category/Group/Brand/Type/FormatCode, DepartCode, CurrencyCode, ExchangeRate, **PORefNo, PORefLineNum**, NetAmount, HomeAmount, TaxRate, PackingRate1/2, behindindex, LineNumber, StatusReceive, **IRRefNo**, LotNumber, LotExpireDate, LotMyDesc, Color/Size/StyleCode, IsProcess (`(Option.IsBehindIndex==0 || แก้ไข) ? 1 : 0`) + (ถ้า `Option.m_nProductCentralize=1`) SCGProductId, SCGBarcode, SCGProductname, DealerBuyUnit.
- ผลข้างเคียงตาราง: `BCItemWarehouse` (`UpdateWHShelf` `:602-605` — เพิ่ม/ปรับคู่ สินค้า-คลัง-ที่เก็บ), `BCIRSUB` (GRRefNo, CNQty), `BCPurchaseOrderSub` (IsProcess), `processdoc` (DocType 2).

**K. ข้อสังเกต / ยังไม่ verify**
- ข้อความภาษาไทยของ key ที่ขึ้นต้นด้วย `~` อนุมานจากชื่อ key/คอมเมนต์ในโค้ด (ตารางภาษา `myresource` ไม่อยู่ใน repo).
- BC ต้องมีเวอร์ชันของ "ใบรับสินค้า" ที่ **ไม่โชว์ยอดรวม** (ตามต้นฉบับ) แต่ยังคำนวณหัวเอกสารให้ครบเมื่อบันทึก.
- ยังไม่ verify: เนื้อใน `OnSelectItemByPO` (Shift+F7 `GoodsReceive.cpp:615-770`), `OnMore` เมนูป๊อปอัป, `UpdateWHShelf` เงื่อนไขเรียก.

### 1.4 รายการ 5799 "บันทึกตั้งหนี้จากการซื้อ" (ใบตั้งหนี้ — IR / Invoice Receive)

- **ฟอร์ม**: `CPOFrmInvoiceReceive` (`POFrmInvoiceReceive.cpp`, 4,737 บรรทัด) หน้าต่างเดียว `IDD_PO_FRM_INVOICE_RECEIVE` 348x241 (`BC5Account.rc:1749`) — **ไม่มีแท็บ**: หัวเอกสารด้านบน, กริดกลางจอ, ช่องรอง+ยอดรวมด้านล่าง (`MainFrm.cpp:1736`).
- **ตาราง**: หัว `BCAPInvoice` ตั้ง `GRIRBillStatus = 2`; รายการ **`BCIRSUB`** (ไม่ใช่ `BCAPInvoiceSub`) — มีคอลัมน์อ้างใบรับสินค้า `GRRefNo/GRRefLinenum` และ `CNQty` (`POFrmInvoiceReceive.cpp:1284`).
- **เป้าหมาย**: บันทึก **ใบกำกับภาษีซื้อ/ตั้งเจ้าหนี้** โดยอ้างใบรับสินค้า (GR 5798) — ตั้งยอดเจ้าหนี้ (`BuyAPProcess`), บันทึกภาษีซื้อ (`BCInputTax`), หัก ณ ที่จ่าย, หักมัดจำ, ลง GL; **ไม่เพิ่มสต็อกเอง** เพราะสต็อกเข้าตอน GR แล้ว. กรณี "ตั้งหนี้ก่อน รับของทีหลัง" (IR ไม่อ้าง GR) ใบนี้รออยู่ให้ GR ดึงไปรับ (ปุ่ม "เลือกใบตั้งหนี้" ใน 5798).

**A. หัวเอกสารและช่องรอง (ลำดับตามตำแหน่งบนจอ y แล้ว x; หน่วย dialog)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบตั้งหนี้: | รันเลข `m_edDocNo` + ปุ่ม (90,4) | ใช่ | ตามรูปแบบรายวัน `BCAPInvoice`; ยาว > 20 ตัว → `DOCNO_MORE_LENG` | `:314-317,4511` |
| 2 | วันที่ออกเอกสาร: | วันที่ พ.ศ. (246,4) + ปฏิทิน | ใช่ | = วันทำงาน; ปิดช่องถ้าไม่มีสิทธิ์แก้วันที่ | `:329-331,370` |
| 3 | รหัสเจ้าหนี้: | Link `BCAP.Code` + ปุ่ม + ชื่อ (90,16) | ใช่ | ออกจากช่อง → ดึงชื่อ, อัตราภาษีตั้งต้น, ส่วนลด `Discount`/`DetailDiscount`, ผู้ติดต่อ, เครดิต→วันถึงกำหนด (`OnKillfocusEdApcode` :4178); ถ้า `BuyItemSupplier=1` ต้องมีเจ้าหนี้ก่อนค้นสินค้า (`PLEASE_INSERT_AP` :2628) | `:309-311,4178` |
| 4 | เลขที่ใบกำกับ: | รันเลข/ข้อความ `m_edTaxNo` + ปุ่ม (90,28) | ไม่ (แต่ถ้ากรอกต้องมีวันที่) | ยาว > 20 ตัว → `TAXNO_MORE_LENG` ตัดเหลือ 20; กรอกแล้วตั้ง "วันที่ใบกำกับ" = วันที่เอกสารอัตโนมัติ; **ซ้ำกับใบกำกับของเจ้าหนี้รายเดียวกัน** (`BCInputTax WHERE Source=2 AND TaxNo AND ApCode`) → `DUPLICATE_TAXNO` แล้วล้างช่อง; เลขที่ใบกำกับว่าง = ไม่ถือเป็นภาษีซื้อ (ภาษีรวมใน ไม่ถอด VAT จนกว่าจะกรอก) | `:319-324,3604-3660` |
| 5 | วันที่ใบกำกับ: | วันที่ พ.ศ. (246,28) | ใช่ถ้ามีเลขที่ใบกำกับ | ค่าเริ่มต้น = วันทำงาน; มีเลขที่แต่ไม่มีวันที่ → `CHECK_TAXDATE` ("มีการใส่ข้อมูลเลขที่ใบกำกับภาษี แต่ไม่ได้ใส่วันที่ใบกำกับภาษี") | `:333-334,770,1592` |
| 6 | ประเภทการซื้อ: | Combo 4 ค่า key `CP_DOCTYPE_1..4` (90,40) | — | **ล็อก (ปิดช่อง) = ค่า 1** (ซื้อเชื่อสินค้า; ค่า 0/1/2/3 = สินค้าสด/สินค้าเชื่อ/บริการสด/บริการเชื่อ ตามการแมป GL ใน 5812 `GoodsPurchase.cpp:3267-3270`) ; เปลี่ยนไม่ได้ในฟอร์มนี้ | `:298-307` |
| 7 | ประเภทภาษี: | Combo 3 ค่า (key `CB_TAXTYPE_1`, `CB_TAXTYPE_2`, `TAX_ZERO`) (246,40) | ใช่ | `DefaultTaxType`; ตามเอกสาร GR ที่ดึงมา (`OnSelchangeTaxType` คำนวณใหม่) | `:290-296,1572` |
| 8 | SAP | checkbox (90,52) | ไม่ | ธง `IsSAP` บันทึกลงหัวเอกสาร (ใช้กรองตอนเลือก GR: `##SAP##`) — ไม่ใช่คุณสมบัติแกนกลางของ BC ยังไม่ verify เงื่อนไขการแสดง | `:3891`; `m_bIsSAP` |
| — | (ปุ่ม) เลือกใบรับสินค้า (`SELECT_GR`) | ปุ่ม (53,151) | — | ดึงรายการจาก GR (Ctrl+F3) — ต้องมีเจ้าหนี้ก่อน (`PURCHORDER_APCODE`) | `:3857-3864,3573` |
| 9 | รหัสผู้ติดต่อ: | Combo-table `BCContactList` (164,151) | — | กรองตามเจ้าหนี้ (`OnContractFilter`) | `:356` |
| 10 | รหัสแผนก: | Combo-table `BCDepartment` (53,163) | — | | `:353` |
| 11 | โครงการ: | Combo-table `BCProject` (164,163) | — | | `:354` |
| 12 | การจัดสรร: | Combo-table `BCAllocate` (53,175) | — | | `:355` |
| 13 | วันที่วางบิล: | วันที่ (164,175) | — | บันทึกที่ `StatementDate`; ออกจากช่อง → `OnKillfocusEdPayBillDate` | `:336,1704` |
| 14 | เครดิต(วัน): | ตัวเลข (53,187) | — | แก้ → วันถึงกำหนด = วันที่เอกสาร + เครดิต; (ล็อก/ปลดล็อกตามสถานะ ดู :1632-1671) | `:338,1632` |
| 15 | วันที่ถึงกำหนด: | วันที่ (164,187) | — | แก้ → เครดิตคำนวณกลับ | `:335` |
| 16 | กลุ่มเอกสาร: | ข้อความ อ่านอย่างเดียว (53,199) | — | จากกลุ่มเอกสารของรูปแบบเลขที่ | `:1700` |
| 17 | อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข `AmountPoint` (164,199) | — | ตั้งต้น `DefaultTaxRate`/เจ้าหนี้/GR; แก้แล้วคำนวณใหม่ | `:313,3956` |
| 18 | รหัสสกุลเงิน: | Combo-table `BCCurrency` (53,211) | ปิดช่องเมื่อไม่ใช้หลายสกุลเงิน | | `:347-350` |
| 19 | อัตราแลกเปลี่ยน: | ตัวเลข 9 ทศนิยม (164,211) | ปิดช่องเมื่อไม่ใช้หลายสกุลเงิน; <=0 → 1.0 | | `:346-349` |

**ยอดรวม (คอลัมน์ขวาของส่วนล่าง x=278; ป้ายอยู่ที่ x=220)**

| ป้ายบนจอ | ช่อง | แก้ได้ | สูตร |
|--|--|--|--|
| รวมมูลค่าสินค้า: | `m_edSumDetail` (151) | ผู้มีสิทธิ์แก้ยอดบิล (`bCanChangeBillAmount`) | ผลรวมจำนวนเงินแถวปกติ → `SumOfItemAmount`; ต้องเท่ากับผลรวมรายการย่อย ไม่งั้น `CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT` (:931) |
| มูลค่าสินค้ายกเว้นภาษี: | `m_edExceptTax` (163) | เหมือนกัน | → `ExceptTaxAmount` |
| ส่วนลด%,บาท: | ปุ่ม `%` + `m_edDiscountAmount` (175) | ปิดถ้า `CompanyInfo.BillDiscount=0` | ส่วนลดท้ายบิล → `DiscountWord`, `DiscountAmount` |
| ภาษีมูลค่าเพิ่ม: | `m_edTaxAmount` (187) | ผู้มีสิทธิ์ | แก้ได้ (`OnKillfocusEdTaxamount`) → `TaxAmount` |
| ยอดรวมภาษี: | `m_edSumTaxAmount` (199) | ผู้มีสิทธิ์ | → `TotalAmount` |
| ยอดรวมเงินสุทธิ | `m_edNetAmount` (211) | ผู้มีสิทธิ์ | → `NetDebtAmount`; `HomeAmount = ปัด(สุทธิ × อัตราแลกเปลี่ยน)` (`:1095-1099`) |

สูตรภาษีท้ายบิลเหมือน PO (`BillTaxProcess` `lib/BC5FRMWRK/BCStools.cpp:155-233` — ดู §1.1 หัวข้อ D). ผลต่างที่ IR: ภาษีรวมใน ใช้ได้ก็ต่อเมื่อ **มีเลขที่ใบกำกับ** (แถวที่ ExceptTax ≠ 1 ถอด VAT ตอนมี TaxNo `:3656`).

**B. กริดรายการ (11 คอลัมน์ — `:207-219`)**

| # | ป้าย (key) | กว้าง/flag | ความหมาย |
|--|--|--|--|
| 0 | รหัสสินค้า | 60 แก้ได้+Link | ล็อกเมื่อดึงจาก GR; ถ้า `CompanyInfo.BUYNoAddnew=1` ห้ามเพิ่มแถวเอง (`:3930-3933`) |
| 1 | ชื่อสินค้า | 120 อ่านอย่างเดียว | |
| 2 | คลังสินค้า | 60 Combo-table | ไม่อยู่ในคลัง-ที่เก็บของสินค้า → `NO_GOODS`/`NOT_GOODSMASTER` (:2339,2452) |
| 3 | ที่เก็บ | 60 Combo-table | ไม่กำหนด → `NOT_SET_SHELFCODE` (:2335) |
| 4 | หน่วยนับ | 60 Combo-table | ไม่ใช่หน่วยของสินค้า → `NOUNIT` (:2354-2358) |
| 5 | จำนวน | 60 ตัวเลข `QtyPoint` | **เมื่อแถวมีเลขที่ใบสั่งซื้ออ้างอิง (shadow 17 `PORefNo`, เฉพาะตอนสร้างเอกสารใหม่) ห้ามเกินจำนวนเดิม (shadow 18 `OldQty` = ยอดที่ดึงมาจาก GR)** → `IRQTY_MORE_GRQTY` ("จำนวนสินค้าที่กรอกตั้งหนี้ มากกว่าจำนวนที่รับไว้") แล้วคืนค่าเป็น `OldQty` (:2461-2472) |
| 6 | ราคาซื้อ (`BUYPRICE`) | 60 ตัวเลข `PricePoint` | ล็อกถ้า `CompanyInfo.BUYNoEditPrice=1` (คอลัมน์ 6 และ 8) ตอนดึงจาก GR (`:3922-3925`) |
| 7 | ส่วนลด%,บาท | 70 ข้อความ | ล็อกถ้า `BUYNoEditDiscount=1` (`:3927`) |
| 8 | จำนวนเงิน | 70 ตัวเลข | |
| 9 | เลขที่ใบรับสินค้า (`GR_DOCNO`) | 60 แก้ได้ | เลขที่ GR ต้นทางของแถว (ใช้ตัดออกจากรายการเลือกซ้ำ `:3874`) |
| 10 | DP Number | 60 อ่านอย่างเดียว | `BCIRSUB.DPNumber` = เลขที่ใบกำกับของ GR ต้นทาง (`taxno` ของ GR `:3949`) — ข้อความป้ายเป็นอังกฤษตรงในโค้ด |

Shadow columns สำคัญ (`:241-261`): 7 ExceptTax · 15 NetAmount · 16 HomeAmount · **17 PORefNo** · **18 OldQty** (จำนวนเดิมจาก GR) · 19 SetCode (รหัสสินค้าชุด) · 20 DepartCode · 21 DiscountAmount · **22 PORefLinenum** · **23 GRRefLinenum** · 24-26 Color/Size/StyleCode · 27 StatusReceive · 28 OldItemCode · 29 OldPrice — เขียนลง `BCIRSUB` (`:1284`); เลขที่ GR ต้นทางอยู่ในคอลัมน์ที่เห็น #9 (`GRRefNo`).

**C. ปุ่มเลือกใบรับสินค้า (`OnBttnSelectGR` `:3850-4170`)**: หน้าต่างเลือกหลายใบ (config `GoodsReceiveConfig`) แสดง GR ที่: เจ้าหนี้ตรงกัน, `GRIRBillStatus=1`, **`GRBillStatus=0` (ยังไม่ถูกตั้งหนี้หมด)**, ไม่ยกเลิก, `IsCompleteSave=1`, ประเภทภาษีตรงกับ IR, ไม่ใช่ GR ที่เลือกมาแล้วในกริดนี้ (+กรอง SAP) → ล็อกคอลัมน์ 0; ล้างค่าหัวจาก GR ใบแรก (ภาษี, เครดิต/วันถึงกำหนด, ผู้ติดต่อ, สกุลเงิน/อัตรา, แผนก, โครงการ, การจัดสรร, กลุ่มเอกสาร, ส่วนลดท้ายบิล); รายการ = `BCAPInvoiceSub WHERE TransState=0 AND IsCancel=0 AND GRRemainQty>0` เรียง `LineNumber,ItemCode,UnitCode` → จำนวนที่ดึงมา = **`GRRemainQty`** (ยอด GR ที่ยังไม่ตั้งหนี้).

**D. คีย์ลัด (จับใน `PreTranslateMessage` `:3190-3300` + เมนู `OnMore` `:3540-3600`)**: F2 บัญชีแยกประเภท · F3 ภาษีซื้อ (รายการใบกำกับหลายใบ `ClearMultiTaxList`/`GetFirstTaxRow`) · F5 ข้อความอื่นๆ (`CLIBDlgMyDescription` เก็บ `MyDescription`) · Ctrl+F5 ยกเลิก (`OnConfirmCancel` → `BillCancel("BCAPInvoice")`) · Shift+F5 ยืนยัน (`bCanConfirmBill`) · **F8 หักเงินมัดจำ** (`OnApDeposit(เจ้าหนี้, ประเภทภาษี, สกุลเงิน, อัตรา)` แล้ว `CalcTaxAmount`) · Shift+F8 หน่วยคู่ขนาน (`OnOppositQty`) · Ctrl+F8 ตั้งคลังทุกแถว = คลังแถวแรก (ถาม `DEFAULT_WH` Yes/No) · Ctrl+F7 ตั้งคลัง/ที่เก็บเริ่มต้น · Ctrl+F3 เลือกใบรับสินค้า · Ctrl+F12 ประวัติซื้อสินค้า (ต้องเลือกสินค้า `SET_CURSER_ITEM`) · ฟอร์มนี้ผูก `ID_HK_NEW/FIND/PRINT` เท่านั้น (ไม่ผูก F11/PgUp ผ่าน ON_COMMAND — ปุ่มบันทึกใช้ toolbar) (`:152-156`). F6 หัก ณ ที่จ่าย และ Shift+F2 ยังไม่ verify ในฟอร์มนี้ (`OnSaveAPWTax` มีตอนบันทึก — ข้อมูลหัก ณ ที่จ่ายกรอกที่จอ WTax).

**E. กติกาก่อนบันทึก (`ConfirmSaveData` `:720-1020`)**

| # | เงื่อนไข | key | หลักฐาน |
|--|--|--|--|
| 1 | งวดปิด | `IsPeriodClose` | :724 |
| 2 | ลง GL แล้ว | `CANNOT_SAVE_POSTGL` | :729 |
| 3 | เลขที่ว่าง / (แก้เลขที่) ซ้ำ / วันที่ว่าง | `CANNOT_SAVE_NOTDOCNO` / `DOCNO_DUPLICATE` / `CANNOT_SAVE_NOTDOCDATE` | :741-757 |
| 4 | มีเลขที่ใบกำกับ แต่ไม่มีวันที่ใบกำกับ | `CHECK_TAXDATE` | :769 |
| 5 | เลขที่ใบกำกับซ้ำ (เจ้าหนี้เดียวกัน, `BCInputTax Source=2`) | `DUPLICATE_TAXNO` | :780-791 |
| 6 | รหัสเจ้าหนี้ว่าง / ไม่พบ | `CANNOT_SAVE_NOTAPCODE` / `CANNOT_SAVE_NOTAPCODE_INDB` | :798-811 |
| 7 | (แก้ไข) เอกสารลง GL / ยกเลิก / ยืนยันแล้ว | `CANNOT_SAVE_POSTGL` / `NOT_SAVE_ISCANCEL` / `NOT_SAVE_ISCONFIRM` | :823-832 |
| 8 | (แก้ไข) ยอดคงเหลือ ≠ ยอดสุทธิ เฉพาะ `BillType` 1/3 (เชื่อ) = มีการจ่ายชำระ/ลดหนี้/เพิ่มหนี้ตัดยอดแล้ว | `NOT_EDIT_BILLBALANCE` | :835 |
| 9 | (แก้ไข) ยอดวางบิล (`PayBillAmount`) ≠ ยอดสุทธิเดิม | `NOT_EDIT_BILL` | :843 |
| 10 | (แก้ไข) มีเอกสารอ้างอิงต่อ: `BCInvStkRefund`, `BCInvDebitNote2`, `BCStatementSub`, `BCPaymentSub`, `BCPaymentSub2` | `CANNOT_SAVE_BILLSTATUS_REF` / `CANNOT_EDIT_BILL` | :851-868 |
| 11 | จำนวน = 0 (แถวที่มีรหัสสินค้า) | `QTY_ZERO_LIST` | :901 |
| 12 | ผลรวมมูลค่าสินค้า ≠ ผลรวมรายการย่อย | `CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT` | :930 |

**F. ลบ / ยกเลิก / ยืนยัน**

| การกระทำ | กติกา | หลักฐาน |
|--|--|--|
| ลบ (`ConfirmDeleteData` :472-532) | เลขที่ว่าง; งวดปิด; ลง GL แล้ว (`CANNOT_DELETE_POSGL`); ยกเลิก/ยืนยันแล้ว (`NOT_DELETE_ISCANCEL`/`ISCONFIRM`); ยอดคงเหลือ ≠ ยอดสุทธิ (`NOT_DELETE_BILLBALANCE`); มีเอกสารอ้างอิงต่อ (`CANNOT_DELETE_BILLSTATUS_REF`: `BCInvStkRefund`/`BCInvDebitNote2`/`BCStatementSub`/`BCPaymentSub`/`BCPaymentSub2`); ถาม `DELETE_MSG` | ที่อ้าง |
| ลบจริง (`OnDeleteData` :601-666) | `DELETE BCAPInvoice (GRIRBillStatus=2)` → `DELETE BCIRSUB` → `UpdateDocumentFlow(-1,21,เลขที่ IR)` คืนสถานะ GR → `DeleteAccount` (GL source 2) | :626-652 |
| ยกเลิก (`OnConfirmCancel` :4669-4709 → `BillCancelUpdateData` :534-599) | ห้ามถ้า: ลง GL / ยกเลิกแล้ว / ยืนยันแล้ว / ยอดคงเหลือ ≠ ยอดสุทธิ / มีเอกสารอ้างอิง (`CANNOT_CANCEL_USE`); ผล: `BillBalance=0, IsCancel=1` (GRIRBillStatus=2), `BCIRSUB.IsCancel=1`, `UpdateDocumentFlow(-1,21,…)`, `ReverseAccount`, ลบ Serial/Lot | ที่อ้าง |

**G. สถานะ / workflow สามเอกสาร GR-IR-ใบซื้อ**

| คอลัมน์ | ค่า | ความหมาย |
|--|--|--|
| `BCAPInvoice.GRIRBillStatus` | 2 | ใบตั้งหนี้ (IR) |
| `BCAPInvoice.GRBillStatus` (ของ **IR**) | 0 / 1 | 0 = ยังไม่มี GR มารับของ (กรณีตั้งหนี้ก่อน); 1 = ถูก GR ดึงรับของแล้ว (`UpdateGRStatus` ใน 5798) |
| `BCAPInvoice.GRBillStatus` (ของ **GR**) | 0 / 1 | 0 = ยังตั้งหนี้ไม่หมด; **1 = ตั้งหนี้หมดแล้ว** — `UpdateDocumentFlow(…,21,…)` คำนวณ `BCAPInvoiceSub.GRRemainQty = Qty − Σ(BCIRSUB.Qty ที่ไม่ยกเลิกและอ้าง GR/บรรทัดนั้น)` แล้วตั้ง GR header เป็น 1 เมื่อ `Σ GRRemainQty = 0` ไม่งั้น 0 (`mytools.cpp:2300-2345`; `UpdateSaveIRStatus` :1388-1440 ทำงานเดียวกัน) |
| `BCIRSUB.CNQty` | ตัวเลข | จำนวนของ IR ที่ GR ยังไม่ได้รับ (ใช้เมื่อตั้งหนี้ก่อนรับของ) |
| `BillBalance` | ตัวเลข | ตั้งตอนบันทึก; ยอดคงเหลือเจ้าหนี้ — ลดโดยจ่ายชำระ/ลดหนี้ (`BuyAPProcess`) |

**H. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:1023-1386`)**: แก้ไข → `BuyAPProcess(เลขเดิม,-1)` ถอนยอดเจ้าหนี้เดิมก่อน → `GetGLFormat` (ดู ข้อ J) → `StartTransaction` → `InsertDB`/`UpdateDB(BCAPInvoice)` (คอลัมน์ในข้อ I) → (แก้ไข) `UpdateDocumentFlow(-1,21,เลขเดิม)` → `DELETE BCIRSUB` ทั้งเอกสาร แล้ว `InsertDB(BCIRSUB)` ทุกแถวที่มีรหัสสินค้า → `UPDATE BCAPInvoice SET IsCompleteSave=1` → `SetDTSModifyFlag` → `Commit` → `INSERT INTO processdoc (DocNo, DocType=2)` → **`OnSaveInputTax(1, … ,2,…)`** บันทึกภาษีซื้อลง `BCInputTax` (Source=2; ธง "ต่างสกุลเงิน" เมื่ออัตรา ≠ 1) → `OnSaveAPWTax` หัก ณ ที่จ่าย → `OnSaveAPDeposit` ตัดเงินมัดจำที่เลือกด้วย F8 → GL (`m_nManualKey == 0`: `m_GLTLSProcess.m_nSource=2`, `m_nGLGroup` = 2 ถ้าบริการ (`BillType` 2/3) ไม่งั้น 1, **39 ถ้าภาษีแยกนอก (`TaxType=0`)**, `m_nDocStatus=2`; มิฉะนั้น `OnSaveAccount`) → **`BuyAPProcess(เลขที่,1)`** ตั้งยอดเจ้าหนี้ → `UpdateDocumentFlow(1,21,เลขที่ IR)` (ตัดสถานะ GR) → logbook `SAVE_IR` (`SaveLog`). ถ้ารายการย่อยไม่ครบ → `Abort()` + `SAVEDETAIL_NOTCOMPLETE`.

**I. คอลัมน์ที่เขียน**
- `BCAPInvoice` (`:1131`): DocNo, DocDate, **TaxNo**, BillType, **GRIRBillStatus=2**, TaxType, ApCode, DepartCode, CreditDay, DueDate, **StatementDate**, MyDescription, ProjectCode, AllocateCode, BillGroup, ContactCode, RecurName, TaxRate, CurrencyCode, ExChangeRate, ZeroTaxAmount, SumOfItemAmount, DiscountWord, DiscountAmount, AfterDiscount, BeforeTaxAmount, TaxAmount, TotalAmount, ExceptTaxAmount, SumCashAmount, SumChqAmount, SumBankAmount, SumOfWTax, SumOfDeposit1, SumOfDeposit2, HomeAmount, NetDebtAmount, BillBalance, PayBillAmount, ExchangeProfit, GLTransData, GLFormat, GLStartPosting, IsCompleteSave, **IsSAP**, CreatorCode, CreateDateTime (+LastEditorCode, LastEditDateT).
- `BCIRSUB` (`:1284`): DocNo, TaxNo, TaxType, DocDate, ApCode, ItemCode, ItemName, WHCode, ShelfCode, UnitCode, Qty, Price, DiscountWord, Amount, **GRRefNo, DPNumber**, CNQty, DiscountAmount, StockType, MyDescription, Project/AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, Category/Group/Brand/Type/FormatCode, NetAmount, HomeAmount, **PORefNo**, DepartCode, CurrencyCode, ExchangeRate, StatusReceive, BalanceAmount, LineNumber, **PORefLinenum, GRRefLinenum**, Color/Size/StyleCode.
- ตารางข้างเคียง: `BCInputTax`, ตารางหัก ณ ที่จ่ายเจ้าหนี้, ตารางมัดจำเจ้าหนี้, `processdoc`.

**J. รูปแบบ GL ตั้งต้น**: `GetGLFormat` (`:3824-3833`) → กลุ่ม **30** (IR รวม VAT, key `GLFOR_CRPURCHASE_IR`) หรือ **39** ถ้าภาษีแยกนอก (`TaxType=0`; key `GLFOR_IR_EXCVAT`) ผ่าน `GetDefaultGLFormat` (`mytools.cpp:32`).

**K. ข้อสังเกต / ยังไม่verify**: ไทยของ key ที่มี `~` อนุมาน; เนื้อ `OnApDeposit`/`OnSaveInputTax`/`OnSaveAPWTax`/`BuyAPProcess` อยู่ในไลบรารีร่วม (ยังไม่ verify สูตรภายใน); `CalcTaxAmount`/`BillCalc` ของฟอร์มนี้ยังไม่ได้เปรียบเทียบบรรทัดต่อบรรทัดกับ PO (สมมติใช้ `BillTaxProcess` เดียวกัน — ยังไม่ verify).

- **สงสัยข้อบกพร่องตอนแก้ไข IR (ยังไม่รันพิสูจน์)**: ตอนแก้ใบตั้งหนี้ ฟอร์มรวบรวมเลข GR เดิมไว้ใน `strGRRefNo` (`POFrmInvoiceReceive.cpp:1140-1148`) แต่เรียก `UpdateDocumentFlow(-1,21,OldDocNo)` ด้วย **เลขที่ IR เดิม** (`:1166-1167`) ขณะที่ฟังก์ชันตีความค่านี้เป็นเลข GR ในสาขา "ลบ/ยกเลิก" (`mytools.cpp:2330-2347` ใช้ `bcapinvoicesub WHERE docno='<ค่าที่ส่ง>'`) และ `BCIRSUB` เดิมยังไม่ถูกลบจนกว่าจะถึง `:1171-1172` ; ผล: GR ที่ถูกเอาออกจาก IR ระหว่างแก้ไม่ถูกคำนวณ `GRRemainQty/GRBillStatus` ใหม่ (การเรียก `+1` ท้ายบันทึก `:1374` คำนวณเฉพาะ GR ที่ยังถูกอ้างโดยแถวใหม่ `mytools.cpp:2311-2327`). ต่างจากเส้นทางลบ/ยกเลิก (`:564,:636`) ที่ส่ง `strGRRefNo` ถูกต้อง. BC ต้องคำนวณสถานะ GR จากผลรวมแถว IR ที่ใช้งานอยู่ทุกครั้งที่ IR เปลี่ยน ไม่ใช่ปรับเฉพาะ GR ที่ยังอ้างอยู่.

### 1.5 รายการ 5812 "บันทึกซื้อสินค้า,บริการ" (ใบซื้อสินค้า/บริการ — ซื้อตรง สด/เชื่อ)

- **ฟอร์ม**: `CChmFrmGoodsPurchase` (`ChmFrmGoodsPurchase.cpp`, 3,866 บรรทัด; `MainFrm.cpp:1773`) + แท็บ `CChmFrmGoodsPurchasePage1` (`IDD_CHM_PO_FRM_PURCHASE1` 610x280: หัว+กริด+ยอดรวม) และ `CChmFrmGoodsPurchasePage2` (`IDD_CHM_PO_FRM_PURCHASE2` 538x270). (ไฟล์ `POFrmPurchase.cpp` รุ่นเก่าไม่ถูกเรียกจากเมนู.)
- **ตาราง**: `BCAPInvoice` (`GRIRBillStatus = 0` ซื้อปกติ) + `BCAPInvoiceSub`.
- **เป้าหมาย**: บันทึกการซื้อ **ในใบเดียวจบ** (รับสินค้า+ตั้งหนี้+ภาษีซื้อ) — เพิ่มสต็อก, บันทึกภาษีซื้อ (`BCInputTax`), หัก ณ ที่จ่าย (F6), หักเงินมัดจำ (F8), **จ่ายเงินทันทีสำหรับบิลเงินสด (F7)**, ตั้งเจ้าหนี้สำหรับบิลเชื่อ, ลง GL. ดึงใบสั่งซื้อ (PO) ได้เหมือน GR.

**A. หัวเอกสาร แท็บ "ข้อมูลหลัก" (Page1) — ลำดับตามตำแหน่งบนจอ**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | ประเภทการซื้อ: | Combo 4 ค่า `m_edBillType` (97,5) | ใช่ | `CompanyInfo.DefBuyBillType`; **0 = ซื้อสินค้าเงินสด, 1 = ซื้อสินค้าเชื่อ, 2 = ซื้อบริการเงินสด, 3 = ซื้อบริการเชื่อ** (ตามการแมป GL `GoodsPurchase.cpp:3265-3272` และการตรวจเงินสดที่ :2049); เปลี่ยนแล้วดึงข้อมูลเจ้าหนี้ใหม่ (`OnSelchangeDocType` Page1 :921) | `Page1.cpp:177-186` |
| 2 | ประเภทภาษี: | Combo 3 ค่า (295,5) | ใช่ | `DefaultTaxType` (0 แยกนอก / 1 รวมใน / 2 ภาษีอัตราศูนย์) | `Page1.cpp:177-178` |
| 3 | เลขที่ใบซื้อ: | รันเลข + ปุ่ม (97,20) | ใช่ | ตามรูปแบบรายวัน; ยาว > 20 ตัว → `DOCNO_MORE_LENG` | `Page1.cpp:192-215,875` |
| 4 | วันที่ออกเอกสาร: | วันที่ พ.ศ. (295,20) | ใช่ | = วันทำงาน | `Page1.cpp:194-195,325` |
| 5 | รหัสเจ้าหนี้: | Link `BCAP.Code` + ปุ่ม + ชื่อ (97,35) | ใช่ | ออกจากช่อง → ชื่อ/ภาษี/ส่วนลด/ผู้ติดต่อ/เครดิต (`OnKillfocusEdApcode` :791); ล็อกหลังดึงจาก PO | `Page1.cpp:188-190,791` |
| 6 | (ปุ่ม) เลือกใบสั่งซื้อสินค้า (`PO_POTRANS`) | ปุ่ม (402,43) | — | ต้องกรอกเจ้าหนี้ก่อน (`PURCHORDER_APCODE`); PO ที่แสดง: เจ้าหนี้ตรง, `BillStatus=0`, ไม่ยกเลิก, ไม่หมดอายุ, ประเภทภาษีตรง, ยังมี `RemainQty>0`, ไม่ซ้ำกับที่เลือกแล้ว (config `POConfig`) | `Page1.cpp:344-395` |
| 7 | เลขที่ใบกำกับ: | ข้อความ + ปุ่ม (97,49) | ไม่ (ถ้ากรอกต้องมีวันที่) | ยาว > 20 → `TAXNO_MORE_LENG` ตัดเหลือ 20; กรอกแล้วตั้งวันที่ใบกำกับ = วันที่เอกสาร; ซ้ำ (เจ้าหนี้เดียวกัน `BCInputTax Source=2`) → `DUPLICATE_TAXNO` ล้างช่อง | `Page1.cpp:670-710` |
| 8 | วันที่ใบกำกับ: | วันที่ (295,49) | ใช่ถ้ามีเลขที่ใบกำกับ | = วันทำงาน; ขาด → `CHECK_TAXDATE` ตอนบันทึก | `Page1.cpp:196,326` |

**B. แท็บ "รายการอื่นๆ" (Page2) — ลำดับตามตำแหน่ง (x,y)**

| # | ป้ายบนจอ | ชนิด | กติกา | คอลัมน์ | หลักฐาน |
|--|--|--|--|--|--|
| 1 | รหัสแผนก: | Combo-table `BCDepartment` (100,17) | | `DepartCode` | `Page2.cpp:69` |
| 2 | โครงการ: | Combo-table `BCProject` (100,32) | | `ProjectCode` | `:70` |
| 3 | กลุ่มเอกสาร: | ข้อความ อ่านอย่างเดียว (384,32) | | `BillGroup` | rc |
| 4 | รหัสผู้ติดต่อ: | Combo-table `BCContactList` (100,47) | กรองตามเจ้าหนี้ | `ContactCode` | `:72,117` |
| 5 | รหัสสกุลเงิน: | Combo-table `BCCurrency` (384,47) | ปิดช่องเมื่อ `MultiCurrency=0` | `CurrencyCode` | `:64-67` |
| 6 | การจัดสรร: | Combo-table `BCAllocate` (100,62) | | `AllocateCode` | `:71` |
| 7 | อัตราแลกเปลี่ยน: | ตัวเลข 9 ทศนิยม (384,62) | ปิดช่องเมื่อ `MultiCurrency=0` | `ExchangeRate` | `:63-66` |
| 8 | วันที่วางบิล: | วันที่ (100,77) | | `StatementDate` | `:61,164` |
| 9 | อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข (384,77) | ตั้งต้น `DefaultTaxRate`/เจ้าหนี้ | `TaxRate` | `:59,109,176` |
| 10 | เครดิต(วัน): | ตัวเลข (100,92) | แก้ → วันถึงกำหนด = วันที่เอกสาร + เครดิต; ล็อกตามประเภทการซื้อ (เงินสดไม่ใช้เครดิต `:113-114`) | `CreditDay` | `:126-141` |
| 11 | วันที่ถึงกำหนด: | วันที่ (384,92) | แก้ → คำนวณเครดิตกลับ | `DueDate` | `:60,141` |
| 12 | หมายเหตุ: | ข้อความ (100,128) | | `MyDescription` | rc |

**C. กริดรายการสินค้า (9 คอลัมน์ — `Page1.cpp:96-104`)** — เหมือนกริด PO (§1.1) ทุกคอลัมน์: รหัสสินค้า (100 แก้ได้+Link) · ชื่อสินค้า (150) · คลังสินค้า · ที่เก็บ · หน่วยนับ · จำนวน · ราคาซื้อ · ส่วนลด%,บาท · จำนวนเงิน. ผลต่างจาก PO: (1) แถวจาก PO เปลี่ยนหน่วยไม่ได้ → `CANNOT_CHANG_UNIT_PO` (Page1 :1560); (2) ค้นสินค้าเมื่อ `BuyItemSupplier=1`: สินค้าที่ไม่ใช่ของเจ้าหนี้รายนี้ → `NOT_ITEM_SUPPLIER` (:1127), ต้องมีเจ้าหนี้ก่อน (`PLEASE_INSERT_AP` :1042); (3) ราคาซื้อสูงกว่าราคาขาย (`Option.BuyPriceWarning=1`) → `BUYPRICE_MORE_SALEPRICE` (:1649); (4) **โปรโมชันเจ้าหนี้** `OnSupplierPromotion` (:1350-1410) อ่าน `BCSupplierPromo` แจ้งข้อความ "สินค้า […]" ตามเงื่อนไขโปรโมชัน (ยังไม่ verify เนื้อหาเต็ม); (5) ส่วนลดตั้งต้นจากการซื้อครั้งก่อน `GetDiscountFromLastBuy` (:1412, อ่าน `BCAPInvoiceSub`) เมื่อเปิด `Option.IsGetLastDiscount`. Shadow columns 31 ช่อง (`LotNumber` = shadow 30; ดู `Page1.cpp:106-150`) ชุดเดียวกับ GR.

**D. ยอดรวมท้ายเอกสาร (คอลัมน์ขวา x=506/532, ปุ่มกลม x=578)**: รวมมูลค่าสินค้า (178) · มูลค่าสินค้ายกเว้นภาษี (193) · ส่วนลด%,บาท (208, ปุ่ม `%`) · ภาษีมูลค่าเพิ่ม (223) · **มูลค่ารวมภาษี** (238) · **มูลค่าเงินสุทธิ** (253) — สูตรและสิทธิ์แก้เหมือน PO §1.1 D (`BillTaxProcess`); ปรับยอดสุทธิด้วยตัวเลือกท้ายบิล: `มูลค่ารวมภาษี + ยกเว้นภาษี − หัก ณ ที่จ่าย + ภาษีอัตราศูนย์ …` (`GoodsPurchase.cpp:597`) เมื่อเรียกดูข้อมูลสรุป (`OnMore`). การตรวจ: ผลรวมแถว ≠ ช่องรวมมูลค่า → `CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT` (:2042).

**E. ปุ่ม/คีย์ลัด (ผูกจริง `ChmFrmGoodsPurchase.cpp:60-120`)**: F2 บัญชีแยกประเภท (`OnHkGL`) · **F3 ภาษีซื้อ (`OnHkTax` :3518 — หน้าต่างรายการใบกำกับหลายใบ)** · F5 รายละเอียดเพิ่มเติม · **F6 หัก ณ ที่จ่าย (`OnHkWTax` :3577 → ตาราง `BCAPWTaxList SaveFrom=1`)** · **F7 บันทึกจ่ายเงิน (`OnHkPay` :3592 → `OnPay`: เงินสด/เช็ค/โอนธนาคาร/เงินสดย่อย/คูปอง/กระเป๋าเงิน, ส่วนเกิน, รายได้/ค่าใช้จ่ายอื่น)** · **F8 หักเงินมัดจำ (`OnHkDeposit` :3626 → `OnApDeposit` แล้ว `CalcTaxAmount`; ถ้า `Option.IsDepositCash=1` และประเภทเป็นเชื่อ (1/3) → `EARNEST_CASH` ห้ามหัก — ตีความ: เปิดตัวเลือกนี้ = หักมัดจำได้เฉพาะบิลเงินสด ยังไม่ verify ข้อความ)** · F12 ตรวจเส้นทางเอกสาร · Ctrl+F3 `ID_HK_GETSO` เลือกใบสั่งซื้อ · Shift+F7 สินค้าตาม PO · Ctrl+F7/F8 ตั้งคลัง/ที่เก็บเริ่มต้น (`DEFAULT_WH`) · Ctrl+F9 Serial · Shift+F9 Lot (`CANNOT_SET_LOT` ถ้าไม่ใช่ต้นทุน FIFO) · Ctrl+F12 ประวัติซื้อ (`SET_CURSER_APITEM`) · Shift+F2 ขยายแถว · Shift+F5 ยืนยัน · Ctrl+F5 ยกเลิก · F11 บันทึก · PgUp/PgDn · Ctrl+N/F/P. ปุ่ม "ข้อมูล" (`OnInfo` :435-493) แสดงยอดเงินมัดจำ · ยอดภาษีหัก ณ ที่จ่าย · ยอดคงเหลือ.

**F. กติกาก่อนบันทึก (`ConfirmSaveData` `:1696-2070`)**

| # | เงื่อนไข | key | หลักฐาน |
|--|--|--|--|
| 1 | เลขที่ว่าง / (แก้เลขที่) ซ้ำ / วันที่ว่าง | `CANNOT_SAVE_NOTDOCNO` / `DOCNO_DUPLICATE` / `CANNOT_SAVE_NOTDOCDATE` | :1704-1720 |
| 2 | มีเลขที่ใบกำกับ ไม่มีวันที่ | `CHECK_TAXDATE` | :1729 |
| 3 | เลขที่ใบกำกับซ้ำ (เจ้าหนี้เดียวกัน) | `DUPLICATE_TAXNO` | :1748 |
| 4 | งวดปิด | `IsPeriodClose` | :1758 |
| 5 | เจ้าหนี้ว่าง / ไม่พบ | `CANNOT_SAVE_NOTAPCODE` / `CANNOT_SAVE_NOTAPCODE_INDB` | :1764-1777 |
| 6 | (มีเงินมัดจำ) ตัวเลือก `Option.IsWarnDeposit=1` และไม่มีเงินมัดจำ | เตือน | :1784 |
| 7 | (แก้ไข) ลง GL / ยกเลิก / ยืนยัน / **รับวางบิลแล้ว** (`StatementState=1`) | `CANNOT_SAVE_POSTGL` / `NOT_SAVE_ISCANCEL` / `NOT_SAVE_ISCONFIRM` / `CANNOT_EDIT_BILL` | :1797-1810 |
| 8 | (แก้ไข) ยอดวางบิล ≠ ยอดสุทธิเดิม; ยอดคงเหลือ ≠ ยอดสุทธิ (มีการจ่าย/ลดหนี้/เพิ่มหนี้ตัดยอดแล้ว — เฉพาะบิลเชื่อ) | `NOT_EDIT_BILL` / `NOT_EDIT_BILLBALANCE` | :1816-1826 |
| 9 | (แก้ไข) มีเอกสารอ้างอิงต่อ | `CANNOT_SAVE_BILLSTATUS_REF` | :1845 |
| 10 | ไม่มีรายการ / จำนวน = 0 | `CANNOT_SAVE_NOTITEM` / `QTY_ZERO_LIST` | :1854,1896 |
| 11 | Serial: คลังที่เก็บ Serial เปลี่ยนแล้ว (แก้ไข) | `CANNOT_DELETE_SERIALWH` | :1932 |
| 12 | ผลรวมแถว ≠ ช่องรวมมูลค่า | `CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT` | :2041 |
| 13 | **บิลเงินสด (ประเภท 0/2): ยอดที่จ่าย (F7) ไม่เท่ายอดสุทธิ** | `AMOUNT_NOT_EQUAL_PAYMENT` (~ยอดเงินที่กำหนดไม่เท่ากับยอดที่จะจ่ายเงิน กรุณาบันทึกจ่ายเงินก่อน) แล้วเปิดหน้าต่างจ่ายเงินให้อัตโนมัติ | :2049-2058 |

`IsMoneyOk(1, สุทธิ)` (`myespeciallyformview.cpp:4266-4317`): ผลรวม เงินสด + เช็ค + บัตรเครดิต + โอนธนาคาร + เงินสดย่อย + คูปอง + กระเป๋าเงิน (+ ส่วนเกิน) ± รายได้/ค่าใช้จ่ายอื่น ต้องเท่ายอดสุทธิ (ทศนิยม `AmountPoint`).

**G. ลบ / ยกเลิก / ยืนยัน**

| การกระทำ | กติกา | หลักฐาน |
|--|--|--|
| ลบ (`ConfirmDeleteData` :1579-1694) | เลขที่ว่าง; งวดปิด; ลง GL (`CANNOT_DELETE_POSGL`); รับวางบิลแล้ว (`CANNOT_DELETE_APBILL`); มีเอกสารตัดยอดอ้างอิง (`NOT_DELETE_BILLBALANCE`); เช็คจ่ายถูกนำไปตัดหลายใบ (`CHQNOT_EDIT_DELETE_CANCEL` ผ่าน `IsChqRefence`); ยกเลิก/ยืนยันแล้ว; ยอดคงเหลือ ≠ ยอดสุทธิ; มีใบเพิ่มหนี้/ลดหนี้/ส่งคืน/จ่ายชำระอ้างอิง (`CANNOT_DELETE_BILLSTATUS_REF`); Serial เปลี่ยนสถานะ/คลัง; ถาม `DELETE_MSG` | ที่อ้าง |
| ลบจริง (`OnDeleteData` :2137-2195) | คืน `BCPurchaseOrderSub.IsProcess=0` → `DELETE BCAPInvoice` + `BCAPInvoiceSub` → `DeleteAPWTax(1,…)`, `DeleteAPDeposit` → `OnStockReprocess` → `DeleteAccount` (GL source 2) → ลบ Serial/Lot → logbook `SUBTRACT_PURCHASE` → `UpdateDocumentFlow(-1,2,PO)` | :2142-2193 |
| ยกเลิก (`OnConfirmCancel` :3445-3498 → `BillCancelUpdateData` :886-941) | ห้ามถ้า: เช็คถูกตัดหลายใบ / ลง GL / รับวางบิลแล้ว / ยกเลิก-ยืนยันแล้ว / ยอดคงเหลือ ≠ ยอดสุทธิ / มีเอกสารอ้างอิง (`CANNOT_CANCEL_USE`); ผล: `BillBalance=0, IsCancel=1`, `BCAPInvoiceSub.IsCancel=1`, `DeleteAPWTax`, `DeleteAPDeposit`, `OnStockReprocess`, `ReverseAccount`, ลบ Serial/Lot, logbook `CANCEL_PURCHASE`, `UpdateDocumentFlow(-1,2,PO)` | ที่อ้าง |

**H. สถานะเอกสาร**: `GRIRBillStatus=0`; `IsConfirm`/`IsCancel`/`IsCompleteSave` เหมือนเอกสารอื่น; `BillBalance` = **0 สำหรับบิลเงินสด (0/2) / = ยอดสุทธิ สำหรับบิลเชื่อ (1/3)** (`:2267`) — ลดโดยจ่ายชำระ/ลดหนี้; `StatementState` = 1 เมื่อถูกนำไปรับวางบิลแล้ว (ล็อกแก้/ลบ/ยกเลิก); ป้ายสถานะมุมซ้าย (§0.2 → `00-common-ui.md` §10.2).

**I. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:2197-2550`)**: แก้ไข → `BuyAPProcess(เลขเดิม,-1)` → คำนวณหัว/ภาษี → `StartTransaction` → `InsertDB/UpdateDB(BCAPInvoice)` → (แก้ไข) `UpdateDocumentFlow(-1,2,PO เดิม)` + ลบรายการย่อยเดิม → `InsertDB(BCAPInvoiceSub)` ทุกแถว → `IsCompleteSave=1` → `Commit` → `OnStockReprocess` ทุกสินค้า → `INSERT INTO processdoc (DocNo,DocType=2)` → บันทึก Serial / Lot → **`OnSaveInputTax`** (ไม่มีเลขใบกำกับ → `UpdateClearTaxRow` ล้างแถวภาษี) → **`OnSaveAPWTax`** → **`OnSavePayment(1,…)`** (บันทึกเงินสด/เช็ค/ธนาคารที่จ่าย; แก้ไขบิลที่เปลี่ยนจากสดเป็นเชื่อ → `DeletePayment` ก่อน) → **`OnSaveAPDeposit`** → GL (`m_nManualKey==0`: `m_nSource=2`, `m_nGLGroup` = 2 บริการ (ประเภท 2/3) / 1 สินค้า, `m_nDocStatus=0`; ไม่งั้น `OnSaveAccount`) → **`BuyAPProcess(เลขที่,1)`** → `UpdateDocumentFlow(1,2,PO ที่อ้าง)`. ไม่ครบ → `Abort()` + `SAVEDETAIL_NOTCOMPLETE`; logbook `SAVE_PURCHASE`.
  - **รูปแบบ GL ตั้งต้น** `GetGLFormat` (:3265-3272): ประเภท 0→กลุ่ม 1 (`GLFOR_CASH_PURCHASE`), 1→กลุ่ม 2 (`GLFOR_CREDIT_PURCHASE`), 2→กลุ่ม 3 (`GLFOR_CASHSERVICE_PURCHASE`), 3→กลุ่ม 4 (`GLFOR_CREDITSERVICE_PURCHASE`).

**J. คอลัมน์ที่เขียน**
- `BCAPInvoice` (`:2295`): DocNo, DocDate, TaxNo, BillType, GRIRBillStatus(0), TaxType, ApCode, DepartCode, CreditDay, DueDate, StatementDate, **StatementState**, MyDescription, ProjectCode, AllocateCode, BillGroup, ContactCode, RecurName, TaxRate, CurrencyCode, ExChangeRate, ZeroTaxAmount, SumOfItemAmount, DiscountWord, DiscountAmount, AfterDiscount, BeforeTaxAmount, TaxAmount, TotalAmount, ExceptTaxAmount, **SumCashAmount, SumChqAmount, SumBankAmount, PettyCashAmount, SumOfWTax, SumOfDeposit1, SumOfDeposit2, DepositIncTax**, NetDebtAmount, BillBalance, PayBillAmount, ExchangeProfit, HomeAmount, GLTransData, GLFormat, GLStartPosting, **ExcessAmount1/2, OtherIncome, OtherExpense**, IsCompleteSave, CreatorCode, CreateDateTime (+LastEditorCode, LastEditDateT).
- `BCAPInvoiceSub` (`:2449`): DocNo, TaxNo, TaxType, DocDate, ApCode, ItemCode, ItemName, WHCode, ShelfCode, UnitCode, Qty, CNQty, Price, DiscountWord, Amount, NetAmount, HomeAmount, BalanceAmount, DiscountAmount, StockType, MyDescription, Project/AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, Category/Group/Brand/Type/FormatCode, **PORefNo**, DepartCode, CurrencyCode, ExChangeRate, **SumOfExpCost** (ต้นทุนแฝงที่ปันส่วน — ดู 5048), **PORefLinenum**, LotNumber, LotExpireDate, LotMyDesc, Color/Size/StyleCode, StatusReceive, TaxRate, PackingRate1/2, behindindex, LineNumber, IsProcess (+SCG* เมื่อ `m_nProductCentralize=1`).
- ตารางข้างเคียง: `BCInputTax` (Source 2), `BCAPWTaxList`, ตารางจ่ายเงิน (`OnSavePayment`) และมัดจำ (`OnSaveAPDeposit`), `processdoc`, `BCItemWarehouse` (`UpdateWHShelf` :3278-3281).

**K. ข้อสังเกต / ยังไม่ verify**: เนื้อ `OnPay`/`OnSavePayment`/`OnApDeposit`/`OnSaveInputTax` อยู่ในไลบรารีร่วม (ช่องในหน้าต่างจ่ายเงินดู `06-cash-bank.md`); ข้อความ `WANT_SAVE_BILLING` (:3846) = ถามต่อไปบันทึกใบวางบิล? (ยังไม่ verify เงื่อนไข); การตัด `SumOfExpCost` ยังไม่ verify.

### 1.6 รายการ 5800 "บันทึกส่งคืนสินค้า" (ส่งคืนสินค้าที่ยังไม่ตั้งหนี้ — อ้างใบรับสินค้า GR เท่านั้น)

- **ฟอร์ม**: `CPOFrmStkReturn` (`POFrmStkReturn.cpp`, 4,123 บรรทัด; `MainFrm.cpp:3165`; dialog เดียวไม่มีแท็บ `IDD_PO_FRM_INV_STKRETURN` 347x240 `BC5Account.rc:1870`).
- **ตาราง**: หัว `BCStkRefund` (**`BillType = 2`** = "GR Only" `POFrmStkReturn.cpp:1109`) + รายการ `BCStkRefundSub` + รายการอ้างอิงเอกสารต้นทาง `BCInvStkRefund`. (ตารางชุดเดียวกับ 5801/5813 — แยกด้วย `BillType`.)
- **เป้าหมาย**: คืนสินค้าที่รับเข้าคลังด้วย GR ไปแล้วแต่ **ยังไม่ได้ตั้งหนี้** — ตัดสต็อกออก ลดยอด GR ที่ค้างตั้งหนี้ และลง GL; **ไม่แตะยอดเจ้าหนี้** เพราะ GR ไม่เคยเข้ายอดเจ้าหนี้ (`StkRefundAPProcess` กรอง `billtype<>2` ทิ้ง — `lib/BC5PROCESS/bcprocess.cpp:2497-2512`).

**A. หัวเอกสาร (ลำดับตามตำแหน่ง y,x ในไดอะล็อก)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบส่งคืน: | รันเลข `BCStkRefund.DocNo` + ปุ่ม (83,6) | ใช่ | ปิดช่องถ้า `bCanChangeBillNo=0` (ได้เลขอัตโนมัติ); ยาวเกิน → `DOCNO_MORE_LENG` | `:251-252,290-293,3035` |
| 2 | วันที่ออกเอกสาร: | วันที่ พ.ศ. (239,6) | ใช่ | = วันทำงาน; ปิดช่องถ้าผู้ใช้ไม่มีสิทธิ์แก้วันที่ (`IsCanEditDocDate=0` ยกเว้น user `sa`); ใส่วันที่ย้อนหลัง → ตรวจสต็อกย้อนหลังอัตโนมัติ (ข้อความ "ระบบกำลังทำการเช็คสต็อกย้อนหลัง…") | `:255,311-313,3015` |
| 3 | รหัสเจ้าหนี้: | Link `BCAP.Code` + ปุ่ม + ชื่อ (83,18) | ใช่ | ออกช่อง → แสดงชื่อ (`DefaultApName` 0 = Name1 / อื่น = Name2) | `:258-260,2379-2387,2354` |
| 4 | ใบรับสินค้า: | Link (83,30) + ปุ่มเลือก | — (แต่ต้องเลือกก่อนมีรายการ) | ต้องมีรหัสเจ้าหนี้ก่อน → `BEFORE_ARCODE_GR` (~กรุณากรอกรหัสเจ้าหนี้ก่อนเลือกใบรับสินค้า); รายการให้เลือก = GR ของเจ้าหนี้นี้ `GRIRBillStatus=1 AND IsCancel=0 AND IsCreditNote=0` (config `GoodsReceiveConfig`); เลือกแล้ว **ล้างกริดแล้วดึงรายการ GR ทั้งใบ** + คัดลอก แผนก/โครงการ/การจัดสรร/กลุ่มบิล/สกุลเงิน/อัตราแลกเปลี่ยน/ประเภทภาษี/อัตราภาษี จาก GR; พิมพ์เลขที่ตรง ๆ ต้องเป็น GR ที่ผ่านเงื่อนไขเดียวกัน ไม่ผ่าน → ล้างช่อง | `:2251-2342` |
| 5 | รหัสแผนก: | Combo-table `BCDepartment` (239,30) | — | ล็อกตาม GR | `:264,1464` |
| 6 | ประเภทการส่งคืน: | Combo 2 ค่า (83,42) | ใช่ | **0 = ส่งคืนสินค้า (`DEBT_TYPE4`), 1 = ราคาผิด (`DEBT_TYPE2`)** (ค่า 2 "สินค้าแตกเสียหาย" ถูกคอมเมนต์ทิ้ง `:244`); ค่า 0 → ตัดสต็อก, ค่า 1 → ไม่ตัดสต็อก (shadow 25 `StockProcess`); โหมด `oomdts=1` บังคับ = 1 และปิดช่อง; **แก้ไขเอกสารแล้วเปลี่ยนประเภทไม่ได้** → `NOTEDIT_REFUND` + `CANCEL_REFUND` | `:242-248,304-305,3722-3770` |
| 7 | สาเหตุการส่งคืน: | Combo-table `BCCNMaster` (239,42) | — | ค่าเริ่มต้น `01`; โหมด Chm: เปลี่ยนตามประเภท (`ประเภท+1` เป็นเลข 2 หลัก) | `:265,281,3735` |
| 8 | หมายเหตุ: | ข้อความ (83,54) | — | Ctrl+Enter → โฟกัสไปกริด | `:417-419` |

ท้ายฟอร์ม (อ่านอย่างเดียวในโหมด Chm; โหมดทั่วไปแก้ได้): **กลุ่มของบิล** (165,182, อ่านจาก GR) · **มูลค่าที่คิดภาษี** (275,182) · **โครงการ** (58,194) · **การจัดสรร** (165,194) · **ยอดยกเว้นภาษี** (275,194; แก้แล้ว `ยอดสุทธิ = มูลค่าที่คิดภาษี + ยกเว้น + ภาษีอัตราศูนย์` `:3255-3266`) · **สกุลเงิน** (58,206; ปิดช่องเมื่อ `MultiCurrency=0`) · **อัตราแลกเปลี่ยน** (165,206; 9 ทศนิยม) · **ยอดสุทธิ** (275,206). ปุ่ม "ตรวจสอบสต๊อกย้อนหลัง" (46,178; คีย์ Home) = `OnBttnCheckStockReword`.

**B. กริดรายการ (10 คอลัมน์ `:153-162`)**

| # | หัวคอลัมน์ | กว้าง | ชนิด/สิทธิ์ | หมายเหตุ |
|--|--|--|--|--|
| 0 | รหัสสินค้า | 60 | แก้ได้ + Link | ดึงจาก GR อัตโนมัติ; ไม่พบสินค้า → `NOT_SEEITEMCODE`; สินค้าบริการ (`StockType=2`) ห้ามแทรกแถว |
| 1 | ชื่อสินค้า | 120 | แก้ได้เมื่อ `Option.IsEditItemName=1` | |
| 2 | เอกสารอ้างอิง | 80 | แก้ได้ | เลขที่ GR ต้นทาง (shadow 19 `RefDocNo`) |
| 3 | คลังสินค้า | 60 | combo-table | จาก GR |
| 4 | ที่เก็บ | 60 | combo-table | จาก GR |
| 5 | หน่วยนับ | 60 | combo-table | |
| 6 | จำนวนส่งคืน | 60 | แก้ได้ ตัวเลข ชิดขวา | ตั้งต้น = `GRRemainQty`-ฝั่งส่งคืน (`BCAPInvoiceSub.CNQty` คงเหลือที่ยังคืนได้); **คีย์มากกว่า จำนวนเดิม (แปลงหน่วยแล้ว) → `QTY_RETURN_MUCH_RECEIVE` (~จำนวนที่กรอกส่งคืนมากกว่ายอดที่สามารถส่งคืนได้) และคืนค่าเดิม** `:2046-2062`; ถ้า `StockStatus=2` (ห้ามติดลบ) และสต็อกไม่พอ → `STOCK_NOT_ENOUGH` ถาม "แสดงรายการอื่นหรือไม่" |
| 7 | ราคาซื้อลดลง (`BUYPRICE_DECREASE`, ~ราคา) | 60 | ล็อก (แก้ได้เฉพาะเมื่อ `bCanChangePrice` และประเภท = ส่งคืน) | ราคาต่อหน่วยจาก GR |
| 8 | ส่วนลด%,บาท | 70 | ล็อกเมื่อประเภท = ส่งคืน | |
| 9 | จำนวนเงิน | 70 | ล็อก | `จำนวน × ราคา` ปัดตาม `AmountPoint` แล้วหักส่วนลด (`DiscountMethod=0`) หรือหักส่วนลดจากราคาก่อน (`≠0`) `:2858-2870`; ภาษีรวมใน (`TaxType=1`) → ยอดสุทธิบรรทัด = จำนวนเงิน − จำนวนเงิน×อัตรา/(100+อัตรา) `:2876-2880` |
Shadow columns 36 ช่อง (`:2478-2507`): 1 `ItemType`(StockType) · 2 `MyDescription` · 3 `ProjectCode` · 4 `AllocateCode` · 5/6 `OppositeQty/Unit` · 7 `ExceptTax` · 8 `BarCode` · 9-13 หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ · 14 หน่วยเดิม · 16 `NetAmount` · 17 `HomeAmount` · 18 `BillQty`(จำนวนเดิมของ GR) · 19 `RefDocNo`(เลขที่ GR) · 20 `OldQty`(จำนวนคงเหลือเดิม) · 22 `DepartCode` · 23 `CurrencyCode` · 24 `ExchangeRate` · **25 `StockProcess` = ประเภทการส่งคืน (0 ตัดสต็อก / 1 ไม่ตัด)** · 27 `DiscountAmount` · 29 `LineNumber` · **30 `RefLineNumber`** = บรรทัดของ GR (ใช้ตัดยอดคืน) · 31-33 Color/Size/Style · 35 `OldItemCode`.

**C. คีย์ลัด/ปุ่ม** (`PreTranslateMessage` `:409-503`; เมนูปุ่ม "เพิ่มเติม" `OnMore` `:3049-3125`): F2 บัญชีแยกประเภท (`OnAccount`; ซ่อนเมื่อผลิตภัณฑ์ตัดโมดูล GL) · Shift+F2 ขยายรายละเอียดแถว · Ctrl+F2 ดูยอดสต็อก (`GetStock`) · F5 → **Ctrl+F5 ยกเลิกเอกสาร** (ถ้ายืนยันแล้ว → `CANNOT_CANCEL`) / **Shift+F5 ยืนยันเอกสาร** (ต้องมีสิทธิ์ `bCanConfirmBill` ไม่งั้น `NOT_CONFIRM_NO`) · Ctrl+F9 Serial · Shift+F9 Lot · Ctrl+F12 ประวัติการซื้อสินค้า (`OnHistoryPurchase`) · Home ตรวจสต็อกย้อนหลัง · PgDn โฟกัสกริด · **F7 ถูกกลืนไว้ (ไม่มีหน้าต่างจ่ายเงิน)**; F11 บันทึก / Ctrl+N ใหม่ / Ctrl+F ค้น / Ctrl+P พิมพ์ จากคลาสฐาน (§0.2 → `00-common-ui.md` §4).

**D. กติกาก่อนบันทึก (`ConfirmSaveData` `:564-860`)**: เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · (แก้เลขที่) ซ้ำ `DOCNO_DUPLICATE` · งวดปิด · รหัสเจ้าหนี้ว่าง `CANNOT_SAVE_NOTAPCODE` · ลง GL แล้ว `CANNOT_SAVE_POSTGL` · ยกเลิกแล้ว `NOT_SAVE_ISCANCEL` · ยืนยันแล้ว `NOT_SAVE_ISCONFIRM` · ไม่มีรายการ `CANNOT_SAVE_NOTITEM` · (Serial/Lot ต่อสินค้า) ตรวจสต็อก Lot ไม่พอ `STOCK_NOT_ENOUGHLIST` (`:668-760`).

**E. ลบ / ยกเลิก**

| การกระทำ | กติกา | ผล |
|--|--|--|
| ลบ (`ConfirmDeleteData` `:862-901`) | เลขที่ว่าง · งวดปิด · ลง GL (`CANNOT_DELETE_POSGL`) · ยกเลิก/ยืนยันแล้ว · ยอดคงเหลือ ≠ ยอดสุทธิ (`NOT_DELETE_BILLBALANCE`) · ถามยืนยัน `CONFIRM_DELETEDATA` | `DELETE BCStkRefund / BCInvStkRefund / BCStkRefundSub` → `OnStockReprocess` ทุกสินค้า → `DeleteAccount` → ลบ Serial/Lot (`DeleteSerial(7)`, `DeleteLotItem(7)`) → `UpdateDocumentFlow(-1,7,เลขที่ GR)` (`:908-993`) |
| ยกเลิก (`OnConfirmCancel` `:4076-4095` → `BillCancelUpdateData` `:2586-2665`) | ลง GL แล้ว/ยกเลิกแล้ว/ยืนยันแล้ว ห้าม | `BCStkRefund.BillBalance=0, IsCancel=1`, `BCStkRefundSub.IsCancel=1`, `BCInvStkRefund.IsCancel=1`, `OnStockReprocess`, `ReverseAccount`, ลบ Serial/Lot, `UpdateDocumentFlow(-1,7,…)` |

**F. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:1049-1437`)**: `UpdateData` → คำนวณหัว (`BillCalc` `:2921-2964`) → `StartTransaction` → `InsertDB`/`UpdateDB(BCStkRefund)` → `DELETE BCInvStkRefund` เดิม + `INSERT BCInvStkRefund` (1 แถวต่อเอกสาร GR) → `DELETE BCStkRefundSub` + `INSERT` ทุกแถว (รวม Serial/Lot ของสินค้า) → `UPDATE BCStkRefund SET IsCompleteSave=1` → **ลดยอดคงเหลือของ GR**: ถ้ายอดที่คืน = ผลรวมทั้งใบ `UPDATE BCApInvoice SET IsCreditNote=1, IsConfirm=1, BillBalance = BillBalance − ยอดคืน` (ล็อก GR ไม่ให้แก้) · ถ้าคืนไม่หมด `SET IsConfirm=1, BillBalance = BillBalance − ยอดคืน` (+ `IsCreditNote=1` เมื่อยอดคืน = `BillBalance` เดิม) (`:1352-1370`) → `Commit` → `OnStockReprocess` → `INSERT INTO processdoc (DocNo,DocType=7)` → บันทึก Serial / Lot → GL (`m_nManualKey==0`: `m_nSource=3`, **`m_nGLGroup=42` → เอนจิน `ProcessConditionGroup38`** `myespeciallyformview.cpp:6456`; รูปแบบ GL ตั้งต้น = **FormatGroup 31 `GLFOR_CRPURCHASE_GOODSRET`** `:2968`) → `UpdateDocumentFlow(1,7,เลขที่ GR ที่อ้าง)` → logbook `SAVE_APRETURN`. ไม่ครบ → `Abort()` + `SAVEDETAIL_NOTCOMPLETE`.
  - **ข้อสังเกตข้อบกพร่องของต้นฉบับ**: `OnUpdateGRStatus` (`:995-1047`) ประกอบคำสั่ง `UPDATE BCAPInvoiceSub SET CNQty = CNQty − …` แต่เรียก `Execute(strSQL)` (ตัวแปร SELECT) แทน `strupdate` จึงไม่มีผลจริง — BC ต้องตัด `CNQty` รายบรรทัดของ GR ให้ถูกเอง (ยังไม่ verify ว่ามีจุดอื่นลด `CNQty`; ใน `ProcessStock` เป็นไปได้).

**G. คอลัมน์ที่เขียน**
- `BCStkRefund` (`:1056-1118`): CauseType, DocNo, DocDate, DueDate(=DocDate), ApCode, DepartCode, RecurName, CauseCode, MyDescription, AllocateCode, ProjectCode, BillGroup, CurrencyCode, ExchangeRate, SumOfBeforeTax, SumOfTaxAmount, SumOfTotalTax, SumOfExceptTax, NetDebtAmount, SumOfZeroTax, TaxType, TaxRate, **ReturnMoney=0, BillBalance=0, BillType=2**, GLFormat, GLTransData, GLStartPosting=0, IsCompleteSave, CreatorCode/CreateDateTime (หรือ LastEditorCode/LastEditDateT).
- `BCInvStkRefund` (`:1170-1198`): StkRefundNo, DocDate, ApCode, ReturnMoney=0, ReturnStatus=0, InvoiceNo(=เลข GR), DiffAmount(=ยอดสุทธิ), HomeAmount2, ZeroTaxAmount, ExceptTaxAmount, BillType=0.
- `BCStkRefundSub` (`:1328`): DocNo, TaxNo, DocDate, ApCode, CauseType, ItemCode, ItemName, InvoiceNo(เลข GR), WHCode, ShelfCode, UnitCode, Price, DiscountWord, Amount, **DiscQty**(จำนวนคืน), **TempQty**, StockProcess, DiscountAmount, ItemType, MyDescription, ProjectCode, AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, Category/Group/Brand/Type/FormatCode, NetAmount, HomeAmount, BillQty, DepartCode, CurrencyCode, ExchangeRate, **RefLineNumber**, LotNumber, TaxType, TaxRate, PackingRate1/2, behindindex, LineNumber, IsProcess.

**H. ข้อสังเกต / ยังไม่ verify**: (1) `NetDebtAmount`/ช่อง "ยอดสุทธิ" = ผลรวมแถวที่คิดภาษี (ไม่บวกภาษีแยกนอกและไม่รวมยอดยกเว้นในค่าที่เซฟ — `BillCalc` `:2962-2963` คำนวณ `dSum + ยกเว้น` แล้วไม่ใช้) ส่วนภาษีเก็บใน `SumOfTaxAmount`; (2) การตัดสต็อกจริงของ `ProcessStock` (ค่า `StockProcess`) อยู่ในเอนจินคิว ไม่อ่านในรอบนี้; (3) ค่า `m_bChmMode` ("โหมดปูนซีเมนต์ไทย") มาจากการตั้งค่าติดตั้ง ยังไม่ verify — BC ไม่ต้องรองรับโหมดนี้ ใช้พฤติกรรมโหมดทั่วไป.

### 1.7 รายการ 5813 "บันทึกส่งคืนสินค้า/ลดหนี้" (ใบส่งคืน/ลดหนี้เจ้าหนี้ — อ้างใบซื้อ/ใบตั้งหนี้/ยอดยกมา/เจ้าหนี้อื่น/เงินมัดจำ)

- **ฟอร์ม**: `CPOFrmStkRefund` (`pofrmstkrefund.cpp`, 6,427 บรรทัด; `MainFrm.cpp:2256`; dialog เดียว `IDD_PO_FRM_INV_STKREFUND` 461x291 `BC5Account.rc:1951`) + กล่อง "รายละเอียดส่วนอื่นๆ" `CPODlgStkRefund` (`IDD_PO_DLG_STKREFUND` 310x163, เปิดด้วย F5).
- **ตาราง**: หัว `BCStkRefund` + รายการสินค้า `BCStkRefundSub` + **รายการเอกสารต้นทางที่ลดหนี้ `BCInvStkRefund`** (1 แถวต่อ 1 เอกสารที่อ้าง) + มัดจำที่ตัด `BCAPDepositUse`. (ตารางเดียวกับ 5800/5801 ต่างที่ `BillType` — ฟอร์มนี้ **ไม่ใส่ค่า `BillType` ใน `BCStkRefund`** → ใช้ค่าเริ่มต้นของคอลัมน์ ยังไม่ verify DDL; เอนจินกรอง `billtype<>2` จึงนับเอกสารนี้ — `bcprocess.cpp:2509`.)
- **เป้าหมาย**: ลดหนี้เจ้าหนี้จากการคืนสินค้า/ซื้อราคาผิด **ที่ตั้งหนี้แล้ว** — ลดยอดเจ้าหนี้ (`BCAP.DebtAmount`), บันทึกภาษีซื้อลดลง (`BCInputTax` Source 5), ตัดสต็อก (ประเภทคืนสินค้า), ลดยอดค้างของใบซื้อเดิม (เมื่อติ๊ก "ลดยอดใบกำกับภาษีเดิม"), ลง GL.

**A. หัวเอกสาร (y,x ตาม rc)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบส่งคืน/ลดหนี้: | รันเลข `BCStkRefund.DocNo` + ปุ่ม (103,6) | ใช่ | ปิดช่องถ้า `bCanChangeBillNo=0`; ยาวเกิน → `DOCNO_MORE_LENG` | `:332-335,350-351,5537` |
| 2 | วันที่ออกเอกสาร: | วันที่ (326,6) | ใช่ | = วันทำงาน; ล็อกตามสิทธิ์แก้วันที่; ย้อนหลัง → ตรวจสต็อกย้อนหลัง | `:342-344,381-382,5517` |
| 3 | เลขที่ใบกำกับ: | รันเลข `TaxNo` + ปุ่ม (103,19) | ไม่ (ถ้ากรอกต้องมีวันที่) | ยาวเกิน → `TAXNO_MORE_LENG`; **ซ้ำในเจ้าหนี้เดียวกัน → `DUPLICATE_TAXNO` ล้างช่อง**; F3 บันทึกภาษีซื้อก็ดึงเลขที่/วันที่ใบกำกับแถวแรกขึ้นมาตั้งให้อัตโนมัติ | `:337-340,4619-4662,744-763` |
| 4 | วันที่ใบกำกับ: | วันที่ (326,19) | ใช่ถ้ามีเลขที่ใบกำกับ | `CHECK_TAXDATE` | `:340,744` |
| 5 | ประเภทภาษี: | Combo 3 ค่า (103,32) | ใช่ | `DefaultTaxType` (แยกนอก / รวมใน / ภาษีอัตราศูนย์ `TAX_ZERO`); ตั้งตามใบแรกที่เลือกในกริดเอกสาร | `:313-319,4834-4836` |
| 6 | รหัสแผนก: | Combo-table `BCDepartment` (326,32) | — | | `:348` |
| 7 | รหัสเจ้าหนี้: | Link `BCAP.Code` + ปุ่ม + ชื่อ (103,45) | ใช่ | ต้องกรอกก่อนเลือกเอกสาร (`PleaseInsertDataInApCode`) | `:345-347,4001-4005` |
| 8 | ประเภทลดหนี้ (`IDC_ED_CREDITTYPE`, ป้ายในไดอะล็อกซ้ำ "รหัสเจ้าหนี้:" ~ประเภทการส่งคืน/ลดหนี้): | Combo 2 ค่า (103,58) | ใช่ | **0 = ส่งคืนสินค้า/ลดหนี้ (`DEBT_TYPE5`), 1 = ราคาผิด (`DEBT_TYPE2`)**; โหมด `oomdts=1` บังคับ = 1; เปลี่ยนแล้ว reload รูปแบบ GL (`GetGLFormat`) และเขียน `StockProcess` ทุกแถว | `:321-329,5543-5607` |
| 9 | ☐ SAP (`IS_SAP`) | checkbox (257,60) | — | **ซ่อนยกเว้น `oomdts=1`** (การเชื่อม SAP ของลูกค้ารายหนึ่ง — BC ไม่ต้องทำ); กรองเอกสารต้นทางเฉพาะ `CreatorCode='OOMDTS'` | `:362-377,4751` |
| 10 | ☐ ไม่กระทบต่อยอดเจ้าหนี้ (`CASH_APDEBITNOTE_CHM`) | checkbox (103,72) | — | ติ๊ก → `ReturnMoney=1` (ได้รับเงินคืน/เงินสดแทนการลดหนี้ — เอกสารไม่ลดยอดเจ้าหนี้ใน `StkRefundAPProcess`); **ติ๊กแล้วปิดช่อง "ลดยอดใบกำกับภาษีเดิม" และติ๊กออก** | `:1116,5715-5728` |
| 11 | ☐ ลดยอดใบกำกับภาษีเดิม (`REDOCE_OLD_TAX`) | checkbox (257,72) | — | ติ๊ก → `ReturnStatus=1`, **`BillBalance=0` และ `PayBillAmount=0`** ของใบลดหนี้ (ยอดลดไปหักยอดค้างของใบเดิมทันที) ต้องมีเอกสารอ้างอิงอย่างน้อย 1 ใบ (`CANNOT_SAVE_REDUCETAXMUSTREF`); ไม่ติ๊ก → `ReturnStatus=0` และใบลดหนี้เก็บ `BillBalance = PayBillAmount = ยอดลดหนี้` ไว้ให้นำไปตัดตอนจ่ายชำระ; ติ๊กแล้วปิดช่อง "ไม่กระทบยอดเจ้าหนี้" | `:1119-1128,5730-5743,700-707` |
| 12 | ☐ ลดหนี้เงินมัดจำ (`DEPOSIT_CREDITNOTE`) | checkbox (374,72) | — | ติ๊ก → `IsCNDeposit=1`; ล้างติ๊กและปิดช่อง "ลดยอดใบกำกับภาษีเดิม"; กริดบนเลือกจากใบรับเงินมัดจำ (`BCAPDeposit`) แทนใบซื้อ; **เลิกติ๊ก → ลบแถวใบตั้งหนี้ (IR) ออกจากกริดบน**; เปลี่ยนรูปแบบ GL เป็น `GLFOR_CNDEPOSIT_AP` | `:1131,5745-5782,4601-4604` |

**B. กริดบน "เอกสารต้นทาง" (5 คอลัมน์ `:186-190`)** — แถวละ 1 เอกสารที่ลดหนี้

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | ความหมาย |
|--|--|--|--|--|
| 0 | เลขที่เอกสาร | 130 | แก้ได้+Link | คลิกลิงก์ → เมนูเลือกชนิดเอกสาร (ป๊อปอัป): **ใบซื้อสินค้า / ใบตั้งหนี้ (IR) / ยอดยกมาต้นปี / เจ้าหนี้อื่นๆ** (ถ้าติ๊ก "ลดหนี้เงินมัดจำ" ข้ามเมนูไปเลือกใบมัดจำ); ต้องมีรหัสเจ้าหนี้ก่อน; พิมพ์เลขที่ตรง ๆ ตรวจ `OnCheckApInvoice` (`NOT_FOUND_INVOICENO_AP`) |
| 1 | ส่วนลดเดิมของ invoice | 110 | ล็อก | `DiscountWord` ของเอกสาร |
| 2 | ยอดเดิม | 110 | แก้ได้ ตัวเลข | ใบซื้อ/ตั้งหนี้: `NetDebtAmount + มัดจำที่หัก`; อื่น ๆ: `BillBalance + มัดจำ` (`:4850-4871`) |
| 3 | ยอดผลต่าง | 110 | แก้ได้ | **ยอดที่ลดหนี้ (รวมภาษี)** = ยอดสินค้าที่คืนของเอกสารนี้ หักส่วนลดเดิม และบวกภาษีเมื่อ `TaxType=0` และมีเลขที่ใบกำกับ (`BillCalc` `:4463-4521`) |
| 4 | ยอดที่ถูกต้อง | 110 | แก้ได้ | `ยอดเดิม − ผลต่าง − (ภาษีอัตราศูนย์+ยกเว้น)` |
Shadow (13 ช่อง): 1 `InvExchangeRate` · 2 `NewExchangeRate` · 3 `HomeAmount1` · 4 `HomeAmount2` · 5 `ExchangeProfit` · **6 `BillType`** (ใบซื้อ = 0 · **ใบตั้งหนี้ IR = 3** · ยอดยกมา = 1 · เจ้าหนี้อื่น = 2 · **เงินมัดจำ = 8**) · 7 `SumOfItemAmount` · 8 `DiscountAmount` · 9 `ZeroTaxAmount` · 10 `ExceptTaxAmount` · 11 `GRIRBillStatus` · 12 `BillBalance`.
- **ตัวกรองเอกสารที่เลือกได้** (`AnInvoice` `:4747-4793`): `ApCode` ตรง · `IsCancel=0` · `IsCreditNote=0` · **`BillBalance>0`** · `GRIRBillStatus` = 0 (ใบซื้อ) หรือ 2 (IR) — **ไม่เลือก GR (=1) ได้ (ใช้ 5800)** · `TaxType` ตรงกับประเภทภาษีของใบลดหนี้ · **`StatementState=0` (ยังไม่ถูกรับวางบิล)** · เงื่อนไขเลขที่ใบกำกับ: ถ้าใบลดหนี้มีเลขที่ใบกำกับ → เลือกได้เฉพาะเอกสารที่มี `TaxNo`; ไม่มี → เฉพาะที่ไม่มี `TaxNo` · ไม่ซ้ำกับที่เลือกไปแล้ว. ยอดยกมา (`BCAPInvBalance`): ถ้าประเภทภาษี = แยกนอก (ค่า 0) และกริดบนยังว่าง → `INV_BAL_PLS_INC_VAT` (~ทำรายการได้เฉพาะรายการรวมใน) ต้องเปลี่ยนเป็นรวมใน ก่อน (`:4914-4917`); เจ้าหนี้อื่นๆ (`BCAPOtherDebt`) มีเงื่อนไขเดียวกัน (`:5015`).
- ใบมัดจำ (`OnListAPDeposit` `:6159-6282`): `BCAPDeposit` ของเจ้าหนี้, `IsCancel=0`, `IsReturnMoney` ไม่ใช่ 1, `TaxType` ตรง, **`BillBalance>0`**, เงื่อนไข `TaxNo` ตามเลขที่ใบกำกับของใบลดหนี้ (config `APDepositConfig`).

**C. กริดล่าง "รายการสินค้า" (10 คอลัมน์ `:229-238`)** — เหมือน 5800 §1.6 B แต่คอลัมน์ 2 หัว = **ใบซื้อสินค้า/รับสินค้า** (`PURCHASES_RECEIVE_NO`). กติกาต่างจาก 5800:
- ดึงรายการ: คลิกลิงก์รหัสสินค้า → **ต้องเลือกเอกสารในกริดบนก่อน** (`INPUT_DOCNO_REFUND` ~กรุณาระบุเลขที่ใบซื้อสินค้าที่จะทำการส่งคืนสินค้า/ลดหนี้) → หน้าต่างเลือกหลายแถวจาก `BCAPInvoiceSub` (ใบซื้อ/GR) หรือ `BCIRSub` (IR) ที่ `ApCode` ตรง, `IsCancel=0`, `TransState=0`, **`CNQty>0`** (แสดง รหัส/ชื่อ/หน่วย/ราคา/จำนวนส่งคืน(=`CNQty`)/ส่วนลด/จำนวนเงิน/เลขที่ใบซื้อ) · Ctrl+คลิก = ค้นสินค้าปกติ · Shift+คลิก = เลือกสินค้าหลายตัวตรง ๆ (`:3246-3470`).
- **จำนวนคีย์เกินจำนวนที่คืนได้ → `NUM_ITEM_NOTMORE`** แล้วคืนค่าเดิม (`:3028-3036`).
- **ราคาคีย์สูงกว่าราคาเดิม (shadow 28 `OldPrice`) → `INCORRECT_CREDIT_PRICE`** แล้วคืนราคาเดิม — ลดหนี้ได้เฉพาะลดราคา (`:3062-3072`).
- ประเภท "ส่งคืนสินค้า/ลดหนี้" (0): ราคา/ส่วนลดล็อกตามเดิม; ประเภท "ราคาผิด" (1): แก้ราคา (ถ้า `bCanChangePrice`) หรือส่วนลด, จำนวนล็อก — ข้อกำหนดนี้ในโค้ดอยู่ใต้ `m_bChmMode` (โหมดปูนไทย) เท่านั้น; โหมดทั่วไปยังไม่ verify ว่าล็อกหรือไม่ (`:5543-5601`).
- Shadow columns 41 ช่อง (`BC5Account`/`:2419`): ชุดเดียวกับ 5800 + 28 `OldPrice` (ราคาเดิม) · 30 `RefLineNumber` · 31 `LotNumber` · 32 `LotExpireDate` · 33 `LotMyDesc` · 34-35 `OldWHCode/OldShelfCode`(ตรวจ Serial) · 36 `OldItemCode` · 37-39 Color/Size/Style · 40 `BehindIndex`.

**D. ยอดรวมของเอกสาร (คำนวณ `BillCalcInv` `:4526-4593` เก็บในโครงสร้างภายใน แสดงในกล่อง F5)**: `ยอดเดิม`=Σ ยอดเดิม · `ยอดผลต่าง/ลดหนี้ (NetDebtAmount)` = Σ ผลต่าง + ยกเว้นภาษี + ภาษีอัตราศูนย์ · `ยอดที่ถูกต้อง` = ยอดเดิม − NetDebtAmount · `ยอดภาษี` = ผลต่าง × อัตรา ÷ (100+อัตรา) ปัด `AmountPoint` (**อัตรา = `DefaultTaxRate` ของบริษัท ไม่ใช่อัตราของใบเดิม** `:4530,4582`) · `ยอดก่อนภาษี` = ผลต่าง − ภาษี (ถ้ามีเลขที่ใบกำกับ; ไม่มี → = ผลต่าง) · แล้วเรียก `CalcTaxAmount` (`BillTaxProcess`) และอัปเดตแถวภาษีแรก (`MyUpdateFirstTaxRow`).
- **กล่อง "รายละเอียดส่วนอื่นๆ" (F5, `CPODlgStkRefund`)** (ลำดับ y,x): สาเหตุการลดหนี้/ราคาผิด (Combo `BCCNMaster`) · มูลค่าส่วนลด · โครงการ · ยอดก่อนภาษี · การจัดสรร · ยอดภาษี · กลุ่มของบิล · ยอดรวมภาษี · อัตราภาษีมูลค่าเพิ่ม (แก้ได้) · ยอดยกเว้นภาษี · ยอดใบกำกับเดิม · ยอดผลต่าง/ลดหนี้ · รวมมูลค่าสินค้า · ยอดที่ถูกต้อง · หมายเหตุ · สกุลเงิน · อัตราแลกเปลี่ยน (สองช่องสุดท้ายล็อก). ปุ่ม `F11=ตกลง` / `Esc=ยกเลิก`. (**หมายเหตุ/โครงการ/การจัดสรร/สกุลเงิน/กลุ่มบิล/สาเหตุของเอกสารนี้เก็บในกล่องนี้ ไม่มีช่องบนหน้าหลัก**.) กดตกลง → คำนวณ `DetailCalc` ทุกแถวใหม่.

**E. คีย์ลัด** (`PreTranslateMessage` `:3790-3981`; เมนู `OnMore` `:4283-4440`): F2 บัญชีแยกประเภท (Ctrl+F2 ดูสต็อก, Shift+F2 ขยายแถว) · **F3 บันทึกภาษีซื้อ** (`OnInputTax("POCreditNoteConfig",…)` แล้วดึงเลขที่/วันที่ใบกำกับแถวแรกขึ้นหัวเอกสาร) · **F5 รายละเอียดอื่นๆ** (`BilLAdder` เปิดกล่อง §D) · Ctrl+F5 ยกเลิก · Shift+F5 ยืนยัน · Shift+F8 หน่วยคู่ขนาน (ซ่อนเมื่อตัดโมดูล) · Ctrl+F9 Serial · Shift+F9 Lot · Ctrl+F12 ข้อมูลสินค้า/ประวัติซื้อ (`OnHKItemInfo`) · **Shift+F12 ตรวจเส้นทางเอกสาร** (`OnDocumentFlow` อ่าน `BCStatementSub`/`BCPaymentSub`) · PgUp/PgDn สลับกริดบน/ล่าง · Home ตรวจสต็อกย้อนหลัง · Ctrl+Shift+F6 / Ctrl+Shift+F7 = เครื่องมือซ่อนของนักพัฒนา (เปิดไฟล์ batch / คำนวณยอดใหม่ `OnProcessNewvalueForBCCreditNote` — BC ไม่ต้องทำ).

**F. กติกาก่อนบันทึก (`ConfirmSaveData` `:672-1064`)**

| # | เงื่อนไข | key | หลักฐาน |
|--|--|--|--|
| 1 | งวดปิด | `IsPeriodClose` | :692 |
| 2 | ติ๊ก "ลดยอดใบกำกับเดิม" แต่ไม่มีเอกสารอ้างอิง | `CANNOT_SAVE_REDUCETAXMUSTREF` | :700-707 |
| 3 | เลขที่ว่าง / (แก้เลขที่) ซ้ำ | `CANNOT_SAVE_NOTDOCNO` / `DOCNO_DUPLICATE` | :716-724 |
| 4 | รหัสเจ้าหนี้ว่าง | `CANNOT_SAVE_NOTAPCODE` | :732-734 |
| 5 | มีเลขที่ใบกำกับ ไม่มีวันที่ / เลขที่ใบกำกับซ้ำ (เจ้าหนี้เดียวกัน `BCInputTax`) | `CHECK_TAXDATE` / `DUPLICATE_TAXNO` | :743-763 |
| 6 | (แก้ไข) ลง GL / ยกเลิก / ยืนยัน / **ถูกนำไปรับวางบิลแล้ว** / ยอดคงเหลือ ≠ ยอดสุทธิ | `CANNOT_SAVE_POSTGL` / `NOT_SAVE_ISCANCEL` / `NOT_SAVE_ISCONFIRM` / `CANNOT_EDIT_BILL` / `NOT_EDIT_BILLBALANCE` | :777-795 |
| 7 | ไม่มีรายการ / จำนวน = 0 / Serial / Lot ไม่ครบ | `CANNOT_SAVE_NOTITEM` ฯลฯ | :800-1060 |

**G. ลบ / ยกเลิก**

| การกระทำ | กติกา | ผล |
|--|--|--|
| ลบ (`ConfirmDeleteData` `:469-514`) | เลขที่ว่าง · งวดปิด · ลง GL · ยกเลิก/ยืนยัน · **ถูกรับวางบิลแล้ว `CANNOT_DELETE_APBILL`** · ยอดคงเหลือ ≠ ยอดสุทธิ `NOT_DELETE_BILLBALANCE` · ถามยืนยัน | `DELETE BCStkRefund / BCInvStkRefund / BCStkRefundSub / BCAPDepositUse` → `OnStockReprocess` → `DeleteAccount` → Serial/Lot → `UpdateDocumentFlow(-1,7,เลขที่ต้นทาง)` `:590-670` |
| ยกเลิก (`OnConfirmCancel` `:6283-6331` → `BillCancelUpdateData` `:516-588`) | ลง GL / ยกเลิก / ยืนยันแล้ว / รับวางบิลแล้ว / ยอดคงเหลือ ≠ ยอดสุทธิ ห้าม; **ถูกอ้างโดย `BCStatementSub`/`BCPaymentSub` → `CANNOT_CANCEL_USE`** | `BCStkRefund.BillBalance=0, IsCancel=1`; `BCStkRefundSub.IsCancel=1`; `BCInvStkRefund.IsCancel=1`; `OnStockReprocess`; `ReverseAccount`; ลบ Serial/Lot; `UpdateDocumentFlow(-1,7,…)` |

**H. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:1066-1470`)**: (แก้ไข) `StkRefundAPProcess(เลขเดิม,-1)` ถอนผลต่อยอดเจ้าหนี้เดิม → เตรียมหัว → `StartTransaction` → `InsertDB/UpdateDB(BCStkRefund)` → `DELETE BCInvStkRefund` เดิม + `INSERT` แถวต่อเอกสารต้นทาง (ยอดเงินบาทใหม่ `HomeAmount2 = ผลต่าง × อัตราแลกเปลี่ยนใหม่`, กำไรขาดทุนอัตราแลกเปลี่ยน `ExchangeProfit = HomeAmount1 − HomeAmount2`) → `DELETE BCStkRefundSub` + `INSERT` ทุกแถว → `IsCompleteSave=1` → `Commit` → `OnStockReprocess` → `INSERT INTO processdoc (DocNo,DocType=7)` → Serial/Lot → **`OnSaveInputTax(5,…)`** (ภาษีซื้อลดลง; แปลงเงินบาทเมื่อต่างสกุล) → **`OnSaveAPWTax(5,…)`** → (ติ๊กเงินมัดจำ) `OnSaveDeposit` (INSERT `BCAPDepositUse`: DocNo, DocDate, DepositNo, MyDescription, Amount, Balance, DeposTaxType, NetAmount, DPExchangeRate, NewExchangeRate, HomeAmount1/2, ExchangeProfit, LineNumber) → GL (`m_nManualKey==0`: `m_nSource=3`, **`m_nGLGroup=3` → `ProcessConditionGroup3`**) → **`StkRefundAPProcess(เลขที่)`** → `UpdateDocumentFlow(1,7,เลขที่ต้นทางทั้งหมด)` → logbook `SAVE_RETURN_CREDIT`.
- **รูปแบบ GL ตั้งต้น** (`GetGLFormat` `:4595-4617`): ประเภท 0 → FormatGroup **5** `GLFOR_PURCHASERET_AP` · ประเภท 1 (ราคาผิด) → **50** `GLFOR_PURCHASERET_WRONGP_AP` · ติ๊กลดหนี้เงินมัดจำ → **52** `GLFOR_CNDEPOSIT_AP`.
- **ผลของเอนจินยอดเจ้าหนี้** `StkRefundAPProcess(doc, ±1)` (`bcprocess.cpp:2497-2640`): เลือกเอกสาร `ReturnMoney=0 AND IsCancel=0 AND billtype<>2 AND IsCompleteSave=1` → `BCAP.DebtAmount −= NetDebtAmount` (และ `BCAPDebtTable` รายงวด/สกุลเงิน) · `BCAPPeriodState.CNAmount += NetDebtAmount` (EndAmount = Begin + Debt − CN + Debit − Payment) · แล้ว `StkRefundProcessBill`: ต่อแต่ละแถว `BCInvStkRefund` ที่ `ReturnStatus=1` → **`BillBalance` และ `PayBillAmount` ของเอกสารต้นทางลดลงเท่า `DiffAmount`** (BillType 0/3 → `BCAPInvoice`, 1 → `BCAPInvBalance`, 2 → `BCAPOtherDebt`) แล้วถ้า `BillBalance = NetDebtAmount` ตั้ง `IsCreditNote=0` คืน; `ReturnStatus=0` → ไม่ลดยอดค้างแต่ตั้ง `IsCreditNote=1` เมื่อ `TrueAmount=0` · ลดหนี้เงินมัดจำ → `BCAPDeposit.BillBalance` ลดเท่ายอด (แยก VAT ออกเมื่อ `TaxType=0`).

**I. คอลัมน์ที่เขียน**
- `BCStkRefund` (`:1079-1141`): TaxType, CauseType, DocNo, DocDate, TaxNo, DueDate(=DocDate), ApCode, DepartCode, RecurName, TaxRate, MyDescription, AllocateCode, ProjectCode, BillGroup, CurrencyCode, ExchangeRate, CauseCode, SumOfItemAmount, SumOfZeroTax, SumOldAmount, SumTrueAmount, DiscountWord, DiscountAmount, SumOfBeforeTax, SumOfTaxAmount, SumOfTotalTax, SumOfExceptTax, SumOfWTax, **NetDebtAmount**, **ReturnMoney**, **ReturnStatus**, **BillBalance**, **PayBillAmount**, **IsCNDeposit**, GLFormat, GLTransData, GLStartPosting, IsCompleteSave, IsSAP, CreatorCode/CreateDateTime.
- `BCInvStkRefund` (`:1187-1231`): StkRefundNo, DocDate, ApCode, ReturnMoney, ReturnStatus, InvoiceNo, InvoiceDisc, OldAmount, DiffAmount, TrueAmount, InvExchangeRate, NewExchangeRate, HomeAmount1, HomeAmount2, ExchangeProfit, BillType, SumOfItemAmount, DiscountAmount, ZeroTaxAmount, ExceptTaxAmount (+`TransState`, `IsCancel`).
- `BCStkRefundSub` (`:1366`): เหมือน 5800 + TaxNo, TaxType, LotExpireDate, LotMyDesc, ColorCode, SizeCode, StyleCode.

**J. ข้อสังเกต / ยังไม่ verify**: (1) ป้ายคอลัมน์ "ราคาซื้อลดลง" (`BUYPRICE_DECREASE`) ไม่มีข้อความไทยใน rc — ใช้ "~ราคา/หน่วย"; (2) `IsConfirm`/`IsCreditNote` ของใบต้นทางหลังคืน ไม่ได้ถูกตั้งในฟอร์มนี้ (ต่างจาก 5800) — อัปเดตผ่านเอนจินด้านบน; (3) บล็อก SAP, `oomdts`, `m_bChmMode` เป็นของลูกค้าเฉพาะราย — BC ข้าม.

### 1.8 รายการ 5801 "บันทึกลดหนี้" (ใบลดหนี้เจ้าหนี้แบบกรอกยอด — ไม่มีรายการสินค้า, อ้างใบตั้งหนี้ IR ใบเดียว)

- **ฟอร์ม**: `CPOFrmRefund` (`POFrmRefund.cpp`, 2,261 บรรทัด; `MainFrm.cpp:3179`; dialog เดียว `IDD_PO_FRM_INV_REFUND` 400x270 `BC5Account.rc:10328`). เมื่อค่าตั้งค่า `oomdts=1` เปิด `CPOFrmRefund2` แทน (`MainFrm.cpp:3173-3175`; ของลูกค้ารายเฉพาะ — BC ข้าม).
- **ตาราง**: หัว `BCStkRefund` (**`BillType = 1` = "IR Only"** `:1108`, `CauseType=0`) + `BCInvStkRefund` (1 แถว: เอกสารต้นทาง = IR, `BillType=3`; ลดหนี้เงินมัดจำ = `8`). **ไม่มี `BCStkRefundSub`** (ลบอย่างเดียว `:1014`) และ **ไม่กระทบสต็อก**.
- **เป้าหมาย**: ลดหนี้เจ้าหนี้เป็นจำนวนเงิน (เช่น ส่วนลดท้ายบิล/ปรับราคา) โดยอ้างใบตั้งหนี้ (IR) ที่ยังค้างชำระ — ลดยอดเจ้าหนี้ (`BCAP.DebtAmount`), บันทึกภาษีซื้อลดลง (F3), หัก ณ ที่จ่าย (F6), ลง GL.

**A. ฟิลด์ตามตำแหน่งบนจอ (y,x)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | ☐ ลดหนี้เงินมัดจำ (`DEPOSIT_CREDITNOTE`) | checkbox (82,13) | — | ติ๊ก → ป้ายช่อง IR เปลี่ยนเป็น "เลขที่ใบจ่ายเงินมัดจำ" (`DOCNO_PAYMENT`), ล้างช่องแผนก/โครงการ/การจัดสรร/สกุลเงิน/อัตราแลกเปลี่ยน/IR, ติ๊ก "ลดยอดใบกำกับเดิม" ออกและปิดช่อง; รายการที่เลือกมาจาก `BCAPDeposit` แทน IR; `BCInvStkRefund.BillType=8` | `:2058-2075,1210-1212` |
| 2 | เลขที่ใบลดหนี้: | รันเลข `BCStkRefund.DocNo` + ปุ่ม (82,23) | ใช่ | ปิดช่องถ้า `bCanChangeBillNo=0`; ยาวเกิน → `DOCNO_MORE_LENG` | `:166-167,223,1843` |
| 3 | วันที่ออกเอกสาร: | วันที่ (234,23) | ใช่ | = วันทำงาน; ล็อกตามสิทธิ์แก้วันที่ | `:169,224,249` |
| 4 | เลขที่ใบกำกับ: | รันเลข `TaxNo` + ปุ่ม (82,35) | ไม่ | ซ้ำในเจ้าหนี้เดียวกัน → `DUPLICATE_TAXNO` (ล้างช่อง) | `:172-175,1798-1830` |
| 5 | วันที่ใบกำกับ: | วันที่ (234,35) | ใช่ถ้ามีเลขที่ใบกำกับ | `CHECK_TAXDATE`; แก้แล้วคำนวณภาษีใหม่ | `:175,794,1729` |
| 6 | รหัสเจ้าหนี้: | Link `BCAP.Code` + ปุ่ม + ชื่อ (82,47) | ใช่ | | `:180-182,498-526` |
| 7 | เลขที่ใบตั้งหนี้: | Link + ปุ่ม (82,59) | **ใช่ (`CANNOT_SAVE_NOTIR_DOCNO` ~ไม่ได้กำหนดเลขที่ใบตั้งหนี้)** | ต้องมีเจ้าหนี้ก่อน (`IR_BEFORE_AP`); รายการให้เลือก (config `InvoiceReceiveConfig`): `ApCode` ตรง, **`GRIRBillStatus=2` (IR เท่านั้น)**, `IsCancel=0`, **`IsCreditNote=0`**, **`BillBalance>0`** (ถ้าติ๊กมัดจำ → config `APDepositConfig`); พิมพ์เลขที่ตรง ๆ → ดึงค่าเดียวกัน (ไม่พบไม่แจ้งเตือน) | `:184-185,586-700,528-584` |
| 8 | รหัสแผนก: | Combo-table (234,59) | — | คัดลอกจาก IR | `:188,555` |
| 9 | โครงการ: / การจัดสรร: | Combo-table (82,71) / (234,71) | — | คัดลอกจาก IR | `:189-190,553-554` |
| 10 | สาเหตุการลดหนี้: | Combo-table `BCCNMaster` (82,83) | — | ล้างทุกครั้งที่เลือก IR | `:187,574` |
| 11 | กลุ่มของบิล: | ข้อความ (234,83) | — | จาก IR | `:557` |
| 12 | หมายเหตุ: | ข้อความ (82,95) | — | `MyDescription` | `:1396` |
| 13 | สกุลเงิน: / อัตราแลกเปลี่ยน: | Combo-table / ตัวเลข 9 ทศนิยม (82,112)/(234,112) | — | จาก IR; **ปิดช่องเมื่อ `MultiCurrency=0`** | `:215-219,552,556` |
| 14 | ☐ ลดยอดใบกำกับภาษีเดิม (`REDOCE_OLD_TAX`) | checkbox (82,141) | — | ติ๊ก → `ReturnStatus=1`, **`BillBalance=0, PayBillAmount=0`** ของใบลดหนี้, ช่อง "ยอดใบกำกับเดิม" = ยอดค้างของ IR (`m_dBalance`); ไม่ติ๊ก → `ReturnStatus=0`, `BillBalance=PayBillAmount=ยอดลดหนี้`, "ยอดใบกำกับเดิม" = ยอดสุทธิของ IR | `:1093-1101,2048-2056` |
| 15 | ยอดก่อนภาษี: | ตัวเลข (234,142) | — | **ล็อก** (คำนวณ) — เว้นแต่แก้ที่ช่องนี้: `ยอดลด = ก่อนภาษี + ก่อนภาษี×อัตรา/100` (เฉพาะอัตรา>1) แล้วคำนวณช่องอื่นย้อน | `:196,205,1686-1727` |
| 16 | อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข (82,154) | — | จาก IR; แก้ → คำนวณใหม่ | `:192,558,1736` |
| 17 | ยอดภาษี: | ตัวเลข (234,154) | — | ล็อก; แก้ที่ช่องนี้ → คำนวณจากยอดภาษี (`BillTaxProcess` โหมด 3) | `:197,206,1743-1773` |
| 18 | ยอดใบกำกับเดิม: | ตัวเลข (82,166) | — | แก้ได้ | `:193,570` |
| 19 | ยอดรวมภาษี: | ตัวเลข (234,166) | — | ล็อก | `:198,207` |
| 20 | มูลค่าสินค้า: | ตัวเลข (82,178) | — | **ช่องคีย์หลัก**: แก้ → `CalcTaxAmount` (`BillTaxProcess` ตามประเภทภาษีของ IR, ส่วนลด, เลขที่ใบกำกับ) | `:194,572,1641` |
| 21 | ยอดยกเว้นภาษี: | ตัวเลข (234,178) | — | ล็อก (จาก IR) | `:199,208,560` |
| 22 | ส่วนลดรับคืน%,บาท: | ข้อความ (82,190) | — | ล็อก (จาก IR) | `:203,563` |
| 23 | มูลค่าลดหนี้: | ตัวเลข (234,190) | ใช่ (ผล) | = `ยอดรวมภาษี + ยอดยกเว้น`; **แก้ตรง ๆ ได้**: ยอดรวมภาษี = ยอดที่แก้, ก่อนภาษี = ยอด − ยอด×อัตรา/(100+อัตรา) (ค่าเริ่มต้นอัตรา 7 ถ้าช่องอัตราว่าง), ภาษี = ยอด − ก่อนภาษี, ยอดที่ถูกต้อง = ยอดเดิม − ยอดลด | `:200,562,1587-1590,1775-1796` |
| 24 | มูลค่าส่วนลดก่อนภาษี: | ตัวเลข (82,202) | — | ล็อก (คำนวณ); แก้ → คำนวณใหม่ | `:195,204,1648-1684` |
| 25 | ยอดที่ถูกต้อง: | ตัวเลข (234,202) | — | = ยอดเดิม − มูลค่าลดหนี้ | `:201,1588` |

> ไม่มีกริดรายการ ไม่มีแท็บ; `m_nTaxType` มาจาก IR (ไม่มีช่องเลือก).

**B. คีย์ลัด** (`PreTranslateMessage` `:1956-2036`; เมนู `OnMore` `:1866-1955`): F2 บัญชีแยกประเภท · **F3 บันทึกภาษีซื้อ** (`OnInputTax("PORefundConfig",ประเภทภาษี,เจ้าหนี้)`) · **F6 ภาษีหัก ณ ที่จ่าย** (`:2010-2023`; บันทึก `BCAPWTaxList` ผ่าน `OnSaveAPWTax(5,…)`) · F7 กลืนไว้ (ไม่มีการจ่ายเงิน) · F11 บันทึก · **F12 ตรวจเอกสารต้นทาง/ปลายทาง** (`OnDocumentFlow` `:2209`) · Shift+F5 ยืนยัน · Ctrl+F5 ยกเลิก.

**C. กติกาก่อนบันทึก (`ConfirmSaveData` `:742-943`)**: เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · **เลขที่ IR ว่าง `CANNOT_SAVE_NOTIR_DOCNO`** · (แก้เลขที่) ซ้ำ `DOCNO_DUPLICATE` · งวดปิด · เจ้าหนี้ว่าง `CANNOT_SAVE_NOTAPCODE` · เลขที่ใบกำกับมี/ไม่มีวันที่ `CHECK_TAXDATE` · เลขที่ใบกำกับซ้ำ `DUPLICATE_TAXNO` · (แก้ไข) ลง GL `CANNOT_SAVE_POSTGL` / ยกเลิก `NOT_SAVE_ISCANCEL` / ยืนยัน `NOT_SAVE_ISCONFIRM` · **(แก้ไข) ใบลดหนี้ถูกอ้างโดยใบวางบิล/ใบจ่ายชำระ (`BCStatementSub` หรือ `BCPaymentSub` ที่ `InvoiceNo=เลขที่นี้ AND BillType=2 AND IsCancel=0`) → `CANNOT_EDIT_REFERENCE`** (`:836-852`).

**D. ลบ / ยกเลิก**

| การกระทำ | กติกา | ผล |
|--|--|--|
| ลบ (`ConfirmDeleteData` `:945-995`) | เลขที่ว่าง · งวดปิด · ลง GL · ยกเลิก/ยืนยันแล้ว · **ถูกอ้างโดย `BCStatementSub`/`BCPaymentSub` → `NOT_DELETE_BILLBALANCE`** · ถามยืนยัน | `DELETE BCStkRefund` + `DELETE BCInvStkRefund` (+ `BCStkRefundSub`) → **`UPDATE BCAPInvoice SET IsCreditNote=0` ของ IR เดิม** → `DeleteAccount` → logbook `SUBTRACT_CREDITNOTE` (`:997-1038`) |
| ยกเลิก (`OnConfirmCancel` `:2168-2207` → `BillCancelUpdateData` `:1607-1639`) | ลง GL / ยกเลิก / ยืนยันแล้วห้าม; ถูกอ้างโดยใบวางบิล/จ่ายชำระ → `CANNOT_CANCEL_USE` | `BCStkRefund.BillBalance=0, IsCancel=1`; `BCInvStkRefund.IsCancel=1`; `ReverseAccount`; `BCAPInvoice.IsCreditNote=0` ของ IR; logbook `CANCEL_CREDITNOTE` |

**E. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:1040-1272`)**: `InsertDB`/`UpdateDB(BCStkRefund)` (คอลัมน์: TaxType, TaxNo, DocNo, DocDate, DueDate, ApCode, DepartCode, RecurName, TaxRate, MyDescription, AllocateCode, ProjectCode, BillGroup, CurrencyCode, ExchangeRate, CauseCode, SumOfItemAmount, SumOldAmount, SumTrueAmount, DiscountWord, DiscountAmount, SumOfBeforeTax, SumOfTaxAmount, SumOfTotalTax, SumOfExceptTax, SumOfWTax, NetDebtAmount, **ReturnMoney=0, ReturnStatus, BillBalance, PayBillAmount, IsCNDeposit, BillType=1, CauseType=0**, GLFormat, GLTransData, GLStartPosting, IsCompleteSave, CreatorCode/CreateDateTime | LastEditorCode/LastEditDateT) → **(แก้ไขจาก IR คนละใบ)** `BCAPInvoice.IsCreditNote=1` ของ IR ใหม่, `=0` ของ IR เดิม; **(แก้ไขและเดิม `ReturnStatus=1`) คืน `BillBalance` ของ IR เดิมก่อน (`+ ยอดลดหนี้เดิม`)** (`:1143-1156`) → `DELETE BCInvStkRefund` + `INSERT` 1 แถว (InvoiceNo = IR, InvoiceDisc, OldAmount, DiffAmount, TrueAmount, InvExchangeRate, NewExchangeRate=1.0, HomeAmount1/2, ExceptTaxAmount, ExchangeProfit=0, BillType=3 / 8) → `IsCompleteSave=1` → `INSERT INTO processdoc (DocNo,DocType=7)` → **`OnSaveInputTax(5,…)`** (ภาษีซื้อลดลง) → **`OnSaveAPWTax(5,…)`** → (ลดหนี้เงินมัดจำ) `OnSaveDeposit` (`BCAPDepositUse`, ตัดยอด `BCAPDeposit`) → GL: `m_nManualKey==0` → `m_nSource=3`, **`m_nGLGroup=43` → เอนจิน `ProcessConditionGroup39`** (`myespeciallyformview.cpp:6457` "ใบลดหนี้(เจ้าหนี้)-ซื้อเชื่อ"), รูปแบบ GL ตั้งต้น = **FormatGroup 32 `GLFOR_CRPURCHASE_CN`** (`:1860-1863`) → **`StkRefundAPProcess`** (ผลต่อยอดเจ้าหนี้/ยอดค้าง IR เหมือน §1.7 H: ถ้า `ReturnStatus=1` ลด `BCAPInvoice.BillBalance` ของ IR เท่ายอดลดหนี้).

**F. ข้อสังเกต / ยังไม่verify**: (1) ป้ายช่อง IR ในไดอะล็อก `IDC_LB_IR` ไม่อยู่ใน atlas — ป้ายที่เห็น "เลขที่ใบตั้งหนี้:" มาจากช่อง `IDC_ED_IRDOCNO`; (2) ปุ่ม `IDC_CHECK_PUSHCASH` ถูกอ้างในโค้ด (`:2065`) แต่ไม่มีในไดอะล็อก rc ที่สกัดได้ — ไม่ใช้งาน; (3) ช่อง "ยอดใบกำกับเดิม"/"มูลค่าสินค้า"/"มูลค่าลดหนี้"/"ยอดที่ถูกต้อง" ไม่ถูกปิด (`EnableWindow(FALSE)`) จึงแก้ได้ — ยอดลดหนี้เกินยอดค้างของ IR ไม่มีการตรวจในฟอร์ม (ยังไม่ verify ว่ามี trigger ฝั่ง DB); (4) ฟอร์มนี้ไม่ตรวจ `StatementState` ของ IR ต้นทาง (ต่างจาก 5813 ที่กรอง `StatementState=0`).

### 1.9 รายการ 5814 "บันทึกเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)" (ใบเพิ่มหนี้เจ้าหนี้ — เพิ่มยอดหนี้ของใบซื้อ/ใบตั้งหนี้เดิม เพราะรับสินค้าเพิ่มหรือราคาผิด)

- **ฟอร์ม**: `CPOFrmInvCreditNote` (ชื่อคลาสเป็นของฟอร์มลดหนี้เดิมที่ถูกดัดแปลง แต่ทำงานเป็น **ใบเพิ่มหนี้**; `pofrminvcreditnote.cpp` 4,885 บรรทัด; `MainFrm.cpp` ผ่าน `ID_POSYS_TRANS_8` → `OnPosysTrans8`; dialog เดียว `IDD_PO_FRM_INV_CREDITNOTE` 460x291 `BC5Account.rc:2006`) + กล่อง F5 "รายละเอียดส่วนอื่นๆ" `CPODlgDebitNote` (`IDD_PO_DLG_DEBITNOTE` 298x163 `rc:2741`).
- **ตาราง**: หัว `BCDebitNote2` + รายการสินค้า `BCDebitNoteSub2` + รายการเอกสารต้นทาง `BCInvDebitNote2` (1 แถวต่อ 1 ใบซื้อ/ใบตั้งหนี้ที่เพิ่มหนี้; คอลัมน์ `DebitNoteNo, DocDate, ApCode, ReturnStatus(=0 เสมอ), InvoiceNo, InvoiceDisc, OldAmount, DiffAmount, TrueAmount, InvExchangeRate, NewExchangeRate, HomeAmount1, HomeAmount2, ExchangeProfit, SumOfItemAmount, DiscountAmount, ZeroTaxAmount, ExceptTaxAmount` `:1149-1180`).
- **เป้าหมาย**: ตั้งหนี้เพิ่มจากใบซื้อสินค้า (GRIRBillStatus 0) หรือใบตั้งหนี้ IR (GRIRBillStatus 2) ที่เคยบันทึกแล้ว ด้วยเหตุ 2 ประเภท: **(0) เพิ่มหนี้จำนวนสินค้า และเจ้าหนี้** (รับสินค้าเพิ่ม → สต็อกเข้า + เจ้าหนี้เพิ่ม) **(1) ราคาผิด (ปรับต้นทุนสินค้า และเจ้าหนี้)** (ปรับมูลค่า ไม่เพิ่มจำนวน) — ยอดเจ้าหนี้เพิ่ม (`BCAP.DebtAmount`), บันทึกภาษีซื้อ (F3, `BCInputTax.Source=4`), หัก ณ ที่จ่าย (`OnSaveAPWTax(4,…)`), ลง GL. ต่างจาก 5813: **ไม่ลด `BillBalance` ของเอกสารต้นทาง** และ **ไม่มี "ลดยอดใบกำกับเดิม"** (`ReturnStatus` ที่ `BCInvDebitNote2` = 0 เสมอ). ใบเพิ่มหนี้เองมียอดคงค้างจ่าย (`BillBalance = PayBillAmount = NetDebtAmount` `:1048-1051`) จึงถูกเลือกไปวางบิล/จ่ายชำระได้ (`BCStatementSub`/`BCPaymentSub` `BillType=3` `:4612-4613`).

**A. หัวเอกสาร (y,x ตาม rc)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบเพิ่มหนี้: | รันเลข `BCDebitNote2.DocNo` + ปุ่ม (110,7) | ใช่ | ปิดช่องถ้า `bCanChangeBillNo=0`; ยาวเกิน → `DOCNO_MORE_LENG`; ซ้ำตอนบันทึกใหม่ → ถามคีย์ใหม่ (`RunningAgain`) | `:311-314,352,4462` |
| 2 | วันที่ออกเอกสาร: | วันที่ (319,7) | ใช่ | = วันทำงาน; ล็อกตามสิทธิ์แก้วันที่ (`:324,353`); งวดปิด → ห้ามบันทึก | `:323-325,665` |
| 3 | เลขที่ใบกำกับ: | รันเลข `TaxNo` + ปุ่ม (110,20) | ไม่ (ถ้ากรอกต้องมีวันที่) | ยาวเกิน → `TAXNO_MORE_LENG`; ซ้ำในเจ้าหนี้เดียวกัน (`BCInputTax Source=4 AND ApCode`) → `DUPLICATE_TAXNO` ล้างช่อง+ล้างรายการภาษี | `:316-318,3898-3955,737-756` |
| 4 | วันที่ใบกำกับ: | วันที่ (319,20) | ใช่ถ้ามีเลขที่ใบกำกับ | `CHECK_TAXDATE` | `:326,728` |
| 5 | ประเภทภาษี: | Combo 3 ค่า (110,33) | ใช่ | `DefaultTaxType` (แยกนอก `CB_TAXTYPE_1` / ภาษีรวมใน `CB_TAXTYPE_2` / ภาษีอัตราศูนย์ `TAX_ZERO`); **ถูกตั้งตามใบต้นทางใบแรก** (แถว 0) ที่เลือก และตัวกรองรายการต้นทางบังคับ `TaxType` เท่ากับช่องนี้ | `:293-298,3490-3492,3420-3430` |
| 6 | รหัสแผนก: | Combo-table `BCDepartment` (319,33) | — | ล็อกถ้าเอกสารสร้างโดย `OOMDTS` (SAP) | `:334,1491-1503` |
| 7 | รหัสเจ้าหนี้: | Link `BCAP.Code` + ปุ่ม + ชื่อ (110,46) | ใช่ | ตรวจมีใน `BCAP` (`CANNOT_SAVE_NOTAPCODE` ว่าง / `CANNOT_SAVE_NOTAPCODE_INDB` ไม่พบ); เลือกแล้วดึง ชื่อ (`DefaultApName` 0=Name1/1=Name2), `DefaultTaxRate` (ถ้าเจ้าหนี้ไม่มี → ของบริษัท), `CreditDay` → คำนวณ "ลงวันที่" ครบกำหนด; **ต้องกรอกก่อนเลือกเอกสารต้นทาง** (`PleaseInsertDataInApCode`) | `:330-332,4056-4107,3335-3345,700-725` |
| 8 | เครดิต(วัน): | ตัวเลข (110,59) | — | แก้ → ครบกำหนด = วันที่เอกสาร + เครดิต | `:329,4110-4116` |
| 9 | ลงวันที่ (= วันครบกำหนด): | วันที่ (319,59) | — | แก้ → เครดิต(วัน) = (ครบกำหนด − วันที่เอกสาร) − 1 | `:327,4118-4131` |
| 10 | ประเภทเพิ่มหนี้(เจ้าหนี้): | Combo 2 ค่า (110,72) | ใช่ | **0 = เพิ่มหนี้จำนวนสินค้า และเจ้าหนี้ (`DEBT_TYPE6`)** (ค่าเริ่มต้น) / **1 = ราคาผิด (ปรับต้นทุนสินค้า และเจ้าหนี้) (`DEBT_TYPE10`)** — เปลี่ยนแล้ว ตั้ง shadow 25 (`StockProcess`) ทุกแถวสินค้า = ค่าที่เลือก + โหลดรูปแบบ GL ตั้งต้นใหม่ (0→FormatGroup 6, 1→51); ล็อกที่ 1 เมื่อ `oomdts=1` (SAP) | `:303-308,4468-4506,3888-3896,1423-1437` |
| 11 | ☐ SAP (`IS_SAP`) | checkbox (319,75) | — | **ลูกค้ารายเฉพาะ (`oomdts=1`)** — กรองเอกสารต้นทางเฉพาะที่ `CreatorCode='OOMDTS'` `:3396-3398` · BC ไม่ต้องทำ | `:340-348,1054-1058` |

> ซ่อนอยู่ในกล่อง F5 (ไม่มีบนฟอร์มหลัก): สาเหตุการเพิ่มหนี้, โครงการ, การจัดสรร, กลุ่มของบิล, อัตราภาษี, หมายเหตุ, สกุลเงิน/อัตราแลกเปลี่ยน — ดูข้อ D.

**B. กริดบน "เอกสารต้นทาง" (5 คอลัมน์ `:175-182`)** — แถวละ 1 ใบซื้อ/ใบตั้งหนี้ที่เพิ่มหนี้ (`m_BInputListInv`)

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | ความหมาย |
|--|--|--|--|--|
| 0 | เลขที่ใบซื้อ/ตั้งหนี้ (`PO_DOCNO_IR`) | 130 | แก้ได้+Link | คลิกลิงก์ → เมนูเลือกชนิดเอกสาร (ป๊อปอัป): **ใบซื้อสินค้า (`PURCHASE_GOODS`, config `PurchaseConfig`, GRIRBillStatus=0)** / **ใบตั้งหนี้ (`INVOICE_IR`, config `InvoiceReceiveConfig`, GRIRBillStatus=2)** (`:3350-3380`); เลือกได้หลายใบพร้อมกัน; พิมพ์เลขที่ตรง ๆ → ตรวจ `OnCheckApInvoice` (ต้องเป็นของเจ้าหนี้นี้, ไม่ยกเลิก, `IsCompleteSave=1`, **`GRIRBillStatus<>1` = ห้ามใบรับสินค้า GR**, ไม่พบ → `NOT_FOUND_INVOICENO_AP`); ซ้ำในกริด → `DOCNO_DUPLICATE` |
| 1 | ส่วนลดเดิมของ invoice | 110 | ล็อก | `DiscountWord` ของเอกสารต้นทาง |
| 2 | ยอดเดิม | 110 | แก้ได้ ตัวเลข | ใบซื้อสด/ใบตั้งหนี้ที่ `BillType` 0 หรือ 2 → `NetDebtAmount`; อื่น ๆ → `BillBalance` (ยอดค้าง) `:3494-3515` |
| 3 | มูลค่าเพิ่มหนี้ | 110 | แก้ได้ ตัวเลข | ค่าเริ่มต้นเมื่อเลือกจากรายการ = ยอดเดิม; **คำนวณใหม่ทุกครั้งที่แก้รายการสินค้าในกริดล่าง** (`BillCalc` `:3015-3066`: รวมยอดสินค้าที่ `InvoiceNo` ตรงใบนี้, หักส่วนลดใบ, ถ้า `TaxType=0` และมีเลขที่ใบกำกับ → บวกภาษี `× อัตรา/100`; สินค้ายกเว้นภาษี/อัตราศูนย์แยกยอดไปเก็บ shadow 11/10) |
| 4 | ยอดที่ถูกต้อง | 110 | แก้ได้ ตัวเลข | = ยอดเดิม + มูลค่าเพิ่มหนี้ − (อัตราศูนย์ + ยกเว้น) (`DetailCalcInv` `:2990-3003`) |

> shadow กริดบน 12 ช่อง: 0 RowOrder · 1 InvExchangeRate · 2 NewExchangeRate · 3 HomeAmount1 · 4 HomeAmount2 · 5 ExchangeProfit · 6 BillType(0) · **7 `GRIRBillStatus` ของต้นทาง (2=IR → ดึงรายการจาก `BCIRSub`, อื่น ๆ → `BCAPInvoiceSub`)** · 8 SumOfItemAmount · 9 DiscountAmount · 10 ZeroTaxAmount · 11 ExceptTaxAmount (`:179-182`, ตั้งค่า `:3516-3540`). **ตัวกรองรายการต้นทาง** (`AnInvoice` `:3405-3445`): `ApCode` ตรง · `IsCancel=0` · **`IsCreditNote=0`** · `GRIRBillStatus` ตามชนิดที่เลือก · `(BillBalance>0 OR BillType=0 OR BillType=2)` · **`TaxType` เท่ากับช่องประเภทภาษี** · ถ้ากรอกเลขที่ใบกำกับ → ต้นทางต้อง `TaxNo is not null` มิฉะนั้นต้อง `is null` · **ต้องมีบรรทัดที่ `sum(CNQty)>0`** ใน `BCAPInvoiceSub` (ใบตั้งหนี้ใช้ `BCIRSub`) · ไม่ซ้ำกับที่เลือกไว้แล้ว. ใบแรกที่เลือก → ตั้ง สกุลเงิน/อัตราแลกเปลี่ยน/อัตราภาษี ของเอกสาร และประเภทภาษีของฟอร์ม.

**C. กริดล่าง "รายการสินค้า" (10 คอลัมน์ `:219-228`, `m_BInputListWnd`, ดึงด้วย F9)**

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | กติกา |
|--|--|--|--|--|
| 0 | รหัสสินค้า | 100 | แก้ได้+Link | ตรวจ `UpdateAndCheckStockOption` ไม่พบ → `NOT_SEEITEMCODE` ล้างแถว; `.` = แถวหมายเหตุ; สินค้าบริการ (StockType 2) ห้ามแทรกแถว |
| 1 | ชื่อสินค้า | 120 | ล็อก (แก้ได้เมื่อ `IsEditItemName=1`) | จากสินค้า / ใบต้นทาง |
| 2 | เลขที่ใบซื้อ (ใบอ้างอิง) | 90 | แก้ได้ | = เลขที่ใบต้นทางที่แถวนี้ผูก (**ต้องตรงเลขที่ในกริดบน** จึงถูกรวมยอดเข้าแถวนั้นใน `BillCalc`; shadow 19 เก็บเลขเดียวกัน) |
| 3 | คลังสินค้า | 60 | Combo-table | กรองตามคลังของสินค้านั้น (`BCItemWarehouse`) เมื่อ `StockStatus<>0`; ไม่ใช่คลังของสินค้า → `NOT_GOODSMASTER` |
| 4 | ที่เก็บ | 60 | Combo-table | เช่นเดียวกัน |
| 5 | หน่วยนับ | 60 | Combo-table | ต้องเป็นหน่วยของสินค้า (`NOUNIT`); สินค้าหลายหน่วยกรองจาก `BCStkPacking` |
| 6 | จำนวน | 60 | แก้ได้ AUTOMONEY2 ขวา | จำนวนที่เพิ่ม; F9 ดึงมาเป็น `CNQty` คงเหลือของบรรทัดต้นทาง |
| 7 | ราคาซื้อ | 60 | แก้ได้ AUTOMONEY2 ขวา | ถ้าว่าง/0 ตอนเลือกสินค้า → ดึงราคาจากใบต้นทาง (`GetPriceInvoice`) → ไม่มี → ราคาซื้อล่าสุดของเจ้าหนี้ (`GetItemBuyPrice`) |
| 8 | ส่วนลด%,บาท | 70 | แก้ได้ | รูปแบบเดียวกับใบซื้อ |
| 9 | จำนวนเงิน | 70 | แก้ได้ AUTOMONEY2 ขวา | `DiscountMethod=0`: จำนวน×ราคา แล้วหักส่วนลด; `=1`: หักส่วนลดจากราคาก่อน แล้วคูณจำนวน (`DetailCalc` `:2930-2975`); **แก้จำนวนเงินเอง → ถ้าราคาว่างคำนวณราคา = จำนวนเงิน/จำนวน**; `ภาษีรวมใน` + มีเลขที่ใบกำกับ + ไม่ใช่สินค้ายกเว้น → ยอดสุทธิ (shadow 16) = ยอด − ยอด×อัตรา/(100+อัตรา) |

> shadow กริดล่าง 40 ช่อง (ที่ใช้บันทึก): 1 StockType · 2 MyDescription · 3 ProjectCode · 4 AllocateCode · 5/6 OppositeQty/Unit · 7 ExceptTax(1=ยกเว้น,2=อัตราศูนย์) · 8 BarCode · 9–13 หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ · 16 NetAmount · 17 HomeAmount · 18 BillQty(จำนวนเดิม) · **19 เลขที่ใบต้นทาง** · 22 DepartCode · 23/24 สกุลเงิน/อัตรา · **25 StockProcess = ประเภทเพิ่มหนี้ (0/1)** · 26 Lot script · 27 DiscountAmount · **29 RefLineNumber (บรรทัดใบต้นทาง)** · 30–32 Lot · 33–34 คลัง/ที่เก็บเดิม (Serial) · 35–37 สี/ขนาด/แบบ · 38 behindindex · 39 ItemCode. **F9 "ดึงรายการสินค้า"** (`LoadInvoiceSub` `:4140-4262`): ต้องวางเคอร์เซอร์ที่แถวของกริดบนก่อน → โหลดบรรทัดของใบนั้นที่ `CNQty>0 AND IsCancel=0 AND TransState=0 AND ApCode=…` (ใบตั้งหนี้ → `BCIRSub`, อื่น ๆ → `BCAPInvoiceSub`) ข้ามสินค้าที่มีในกริดแล้วจากใบเดียวกัน; จำนวน = `CNQty` คงเหลือ, ถ้าไม่เท่าจำนวนเดิม → จำนวนเงิน = คงเหลือ × ราคา; แล้ว `BillCalc` + `BillCalcInv`. **ไม่มีการตรวจจำนวนที่เพิ่มเกินจำนวนต้นทาง** ในฟอร์ม (ยังไม่ verify ว่ามี trigger ฝั่ง DB).

**D. กล่อง F5 "รายละเอียดส่วนอื่นๆ" (`CPODlgDebitNote` — F11=ตกลง, Esc=ยกเลิก)**

| ป้าย | ช่อง | แก้ได้ | ที่มา/ผล |
|--|--|--|--|
| สาเหตุการเพิ่มหนี้: | Combo-table `BCCNMaster` | ใช่ | บันทึกที่ `BCDebitNote2.CauseCode` (ปูนไทย `m_bChmMode` ตั้งอัตโนมัติ — ข้าม) |
| โครงการ: / การจัดสรร: | Combo-table `BCProject` / `BCAllocate` | ใช่ | → หัวเอกสาร |
| กลุ่มของบิล: | ข้อความ | ใช่ | → `BillGroup` และเป็น "กลุ่ม" ของรายการภาษีซื้อ |
| อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข | ใช่ | แก้ → คำนวณภาษีใหม่ (`BillTaxProcess` โหมด 0 จากรวมมูลค่าสินค้า) |
| หมายเหตุ: | ข้อความ | ใช่ | → `MyDescription` |
| สกุลเงิน: / อัตราแลกเปลี่ยน: | Combo-table `BCCurrency` / 9 ทศนิยม | ปิดเมื่อ `MultiCurrency=0` | เปลี่ยนสกุลเงิน → ดึงอัตราจาก `BCCurrency.ExchangeRate` |
| มูลค่าส่วนลด / ยอดก่อนภาษี / ยอดภาษี / ยอดรวมภาษี / ยอดยกเว้นภาษี / ยอดใบกำกับเดิม / มูลค่าเพิ่มหนี้ / รวมมูลค่าสินค้า / ยอดที่ถูกต้อง | ตัวเลข | **ล็อกทั้งหมด** (คำนวณจากกริด) | `BillTaxProcess`; ถ้ากลับมากดตกลง → คำนวณยอดแถวสินค้าซ้ำ (`DetailCalc(i,9)`) `:2903-2917,142-168` |

**E. คีย์ลัด** (`PreTranslateMessage` `:3115-3320`; เมนู `OnMore` `:3700-3880`): F2 บัญชีแยกประเภท (Shift+F2 = รายละเอียดสินค้าแถว `OnExpand` — หมายเหตุ/โครงการ/การจัดสรร/สต็อก) · **F3 บันทึกภาษีซื้อ** (`OnInputTax("InvDebitNote2Config",1,เจ้าหนี้)` แล้วอัปเดตเลขที่/วันที่ใบกำกับบนจอจากแถวภาษีแรก) · F5 กล่อง "รายละเอียดอื่นๆ" (Ctrl+F5 ยกเลิกเอกสาร; Shift+F5 ยืนยัน) · F8+Shift หน่วยขนาน (`OnOppositQty`) · **F9 ดึงรายการสินค้าจากใบต้นทาง** (Shift+F9 Lot สินค้า; Ctrl+F9 Serial Number เมื่อ `SerialItem=1`) · F11 บันทึก · **F12 ตรวจเอกสารปลายทาง** (`OnDocumentFlow` `:4604` — แสดง `BCStatementSub`/`BCPaymentSub` ที่อ้าง `InvoiceNo=เลขที่นี้ AND BillType=3`; หัวข้อ "ใบเพิ่มหนี้(เจ้าหนี้) เลขที่ …") · Ctrl+F12 รายละเอียดสินค้า (`OnHKCtrlF12`; ไม่เลือกสินค้า → `SET_CURSER_ITEM`) · PageUp/PageDown สลับกริดบน/ล่าง · Enter ที่ช่อง "ลงวันที่" → โฟกัสกริดบน · **Ctrl+Shift+F6/F7 = เครื่องมือแก้ข้อมูลของผู้พัฒนา (`OnProcessNewvalueForBCCreditNote`) — ไม่ต้องทำ**.

**F. กติกาก่อนบันทึก (`ConfirmSaveData` `:643-960`)** — ตามลำดับ: งวดปิด · เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · (แก้เลขที่) ซ้ำ `DOCNO_DUPLICATE` · เจ้าหนี้ว่าง/ไม่มีใน `BCAP` · เลขที่ใบกำกับมี/ไม่มีวันที่ `CHECK_TAXDATE` · เลขที่ใบกำกับซ้ำ `DUPLICATE_TAXNO` · (แก้ไข) ลง GL แล้ว `CANNOT_SAVE_POSTGL` / ยกเลิกแล้ว `NOT_SAVE_ISCANCEL` / ยืนยันแล้ว `NOT_SAVE_ISCONFIRM` / **นำไปรับวางบิลแล้ว (`StatementState=1`) `CANNOT_EDIT_BILL`** · ทุกแถวสินค้า (ไม่ใช่บริการ StockType 1/7 และไม่ใช่ `.`): ต้องมี คลัง/ที่เก็บ/หน่วย (`NOT_SET_WHCODE`/`NOT_SET_SHELFCODE`/`NOT_SET_UNITCODE` ต่อท้าย `NOT_SAVE_FORMITEM`) · สินค้า Serial (`SerialItem=1`, StockType 2): ต้องระบุ Serial และจำนวน = จำนวน Serial; (แก้ไข) คลังเปลี่ยนจากเดิม `CANNOT_EDIT_SERIALWH` · Lot แบบกำหนดเอง (`nUseLotStatus=1 AND nLotStatus=1`, สินค้าต้นทุน FIFO): ต้องระบุ Lot (`NOT_SET_LOT`) · ตั้งรูปแบบ GL ตั้งต้นถ้าว่าง · คำนวณยอดเงินบาท (`HomeAmount1/2`) ของกริดบนใหม่ · **ไม่บังคับให้มีรายการสินค้าหรือเอกสารต้นทางอย่างน้อย 1 แถว** (ยังไม่ verify ว่าเป็นเจตนา).

**G. ลบ / ยกเลิก**

| การกระทำ | กติกา | ผล |
|--|--|--|
| ลบ (`ConfirmDeleteData` `:448-511`) | เลขที่ว่าง · งวดปิด · ลง GL · ยกเลิก/ยืนยันแล้ว · **ถูกรับวางบิลแล้ว `CANNOT_DELETE_APBILL`** · **`BillBalance ≠ NetDebtAmount` (มีการจ่ายแล้ว) `NOT_DELETE_BILLBALANCE`** · Serial ถูกใช้/เปลี่ยนคลัง · ถามยืนยัน | `DebitNote2APProcess(doc,-1)` (ถอนยอดเจ้าหนี้) → `DELETE BCDebitNote2 / BCInvDebitNote2 / BCDebitNoteSub2` → `OnStockReprocess` ทุกสินค้า → `DeleteAccount` → ลบ Serial/Lot `:576-641` |
| ยกเลิก (`OnConfirmCancel` `:4769-4819` → `BillCancelUpdateData` `:513-574`) | ลง GL / ยกเลิกแล้ว ห้าม; รับวางบิลแล้วห้าม; `BillBalance≠NetDebtAmount` ห้าม; **ถูกอ้างโดย `BCStatementSub`/`BCPaymentSub` → `CANNOT_CANCEL_USE`**; ยืนยันแล้ว → `CANNOT_CANCEL` (ยกเว้น `oomdts=1`) | `DebitNote2APProcess(doc,-2)` → `BCDebitNote2.BillBalance=0, IsCancel=1`; `BCDebitNoteSub2.IsCancel=1`; `BCInvDebitNote2.IsCancel=1` → `OnStockReprocess` → `ReverseAccount` → ลบ Serial/Lot |

**H. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:987-1360`)**: (แก้ไข) ถอนยอดเจ้าหนี้เดิมก่อน `DebitNote2APProcess(เลขที่เดิม,-1)` → `InsertDB`/`UpdateDB(BCDebitNote2)` (TaxType, CauseType, DocNo, DocDate, TaxNo, ApCode, DepartCode, RecurName, CreditDay, DueDate, TaxRate, MyDescription, AllocateCode, ProjectCode, BillGroup, CurrencyCode, ExchangeRate, CauseCode, DiscountWord, DiscountAmount, SumOldAmount, SumTrueAmount, SumOfBeforeTax, SumOfTaxAmount, SumOfTotalTax, SumOfExceptTax, **NetDebtAmount**, SumOfZeroTax, SumOfItemAmount, IsSAP, **BillBalance = PayBillAmount = NetDebtAmount**, GLFormat, GLTransData, GLStartPosting=0, IsCompleteSave=0, สร้าง: CreatorCode/CreateDateTime/IsConfirm=0 | แก้: LastEditorCode/LastEditDateT) → `DELETE`+`INSERT BCInvDebitNote2` ต่อแถวกริดบน (HomeAmount2 = มูลค่าเพิ่มหนี้ × อัตราใหม่; `ExchangeProfit = HomeAmount1 − HomeAmount2`) → `DELETE`+`INSERT BCDebitNoteSub2` ต่อแถวสินค้า (ItemCode, ItemName, InvoiceNo, WHCode, ShelfCode, UnitCode, Price, DiscountWord, Amount, **`DiscQty` = จำนวน ถ้าประเภท 0 มิฉะนั้น 0**, TempQty = จำนวน, **StockProcess = ประเภทเพิ่มหนี้**, DiscountAmount, ItemType, ProjectCode, AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ, NetAmount, HomeAmount, BillQty, DepartCode, CurrencyCode, ExchangeRate, RefLineNumber, Lot*, สี/ขนาด/แบบ, IsProcess, TaxRate, PackingRate1/2, behindindex, LineNumber) → ถ้าทุกแถวบันทึกได้: `IsCompleteSave=1` → `Commit` → `OnStockReprocess` ทุกสินค้า → `INSERT INTO processdoc (DocNo,DocType=3)` → Serial/Lot → **`OnSaveInputTax(4,…,1,4,กลุ่มบิล,ต่างสกุล)`** → **`OnSaveAPWTax(4,…)`** → GL: `m_nManualKey==0` → `m_nSource=4`, `m_nGLGroup=4` → เอนจิน `ProcessConditionGroup4` (`myespeciallyformview.cpp:6423-6459`) · รูปแบบ GL ตั้งต้น = **FormatGroup 6 `DN_AP` (ประเภท 0) / FormatGroup 51 `DNWRONGP_AP` (ประเภท 1)** (`GetGLFormat` `:3888-3896`; `GetDefaultGLFormat` → `BCGLFormat` `IsDefault DESC`) → **`DebitNote2APProcess(เลขที่)`** (`bcprocess.cpp:2644-2700`: `BCAP.DebtAmount += NetDebtAmount` และรายงวด/สกุลเงินใน `BCAPDebtTable`; `bcapperiodstate.DebitAmount += NetDebtAmount × อัตรา` — ตำแหน่งที่ 3 ของ `UpdateAPPeriodState`; ท้ายสุด `DELETE BCInvDebitNote2 WHERE TransState=1`) → logbook `SAVE_APINCREASED`. ไม่สำเร็จ → `Abort` + `SAVEDETAIL_NOTCOMPLETE`.
**ผลต่อสต็อก**: เอนจินสต็อกชนิด 7 `DebitNote2StockProcess` (`bcdumpprocjob.cpp:4306-4360`) อ่าน `BCDebitNoteSub2` ที่ `StockProcess<2` และใช้ **`DiscQty` เป็นจำนวน, `HomeAmount` เป็นมูลค่า** → ประเภท 0: สต็อกเข้าเพิ่มตามจำนวน + มูลค่า; ประเภท 1: `DiscQty=0` → ปรับเฉพาะมูลค่า/ต้นทุน (ไม่เพิ่มจำนวน). (`:1326-1327`, ชนิดเอกสารสต็อก 10)

**I. ข้อสังเกต / ยังไม่ verify**: (1) ชื่อคลาส/ไฟล์/IDD มีคำว่า "CreditNote" แต่ใช้ตาราง `BCDebitNote2` และข้อความ "ใบเพิ่มหนี้" — ใน BC ให้ตั้งชื่อ route/ตารางตามความหมาย (เพิ่มหนี้) ไม่ใช่ลดหนี้; (2) ค่าตั้งต้นของ "ยอดที่ถูกต้อง" ตอนเลือกจากรายการ (`:3498-3513`) = ยอดเดิม+ยอดเดิม (คำนวณเกินชั่วคราว) แล้วถูกแก้ทับโดย `BillCalcInv`/`DetailCalcInv`; (3) ปุ่มวิเคราะห์ GL แยกตามรายได้ (`ProcessConditionGroup4` ใช้ `bcdebitnotesub2 JOIN bcitemaccgroup` `gltlsprocess.cpp:13151-13165`) = ต้องมีกลุ่มบัญชีสินค้า (`BCItem.AccGroupCode`) ทุกสินค้า; (4) `HomeAmount1` ใน `BCInvDebitNote2` สร้างจาก shadow 3 ของกริดบน (ตั้งใหม่ตอน `ConfirmSaveData` `:955-963` = (มูลค่าเพิ่มหนี้+อัตราศูนย์+ยกเว้น) × อัตราแลกเปลี่ยน เมื่อ `MultiCurrency=1`) — ตัวคูณซ้ำกับ `HomeAmount2` เมื่อใช้สกุลเงินเดียว (ยังไม่ verify พฤติกรรมหลายสกุลเงิน).

### 1.10 รายการ 5797 "บันทึกใบเสนอซื้อสินค้า" (ใบขอซื้อภายใน — ไม่มีราคา ไม่มีเจ้าหนี้ ไม่ลงบัญชี ไม่กระทบสต็อก)

- **ฟอร์ม**: `CPOFrmStkRequset` (`pofrmstkrequset.cpp` 2,015 บรรทัด; `MainFrm.cpp` `OnPosysTrans1` → `CreateOrActivateFrame(...CPOFrmStkRequset)`; dialog เดียว `IDD_PO_FRM_STKREQUEST` 348x244 `BC5Account.rc:1362`).
- **ตาราง**: หัว `BCStkRequest` + รายการ `BCStkRequestSub` (ลบแล้วเขียนใหม่ทุกครั้งที่บันทึก). ไม่มี `BCInvStkRequest`/ภาษี/GL/`processdoc`. เอกสารปลายทาง = ใบอนุมัติเสนอซื้อ `BCReqConfirm`/`BCReqConfirmSub` (§1.11) — สถานะคำนวณโดย `UpdateDocumentFlow(±1,16,เลขที่ใบเสนอซื้อ)` (`mytools.cpp:2179-2232`): `BCStkRequestSub.RemainQty = Qty − Σ ConfirmQty` ของ `BCReqConfirmSub` ที่ `StkRequestNo=เลขที่นี้ AND ItemCode AND RefLineNumber=LineNumber`; **`BCStkRequest.BillStatus=1` เมื่อ Σ RemainQty ≤ 0 (อนุมัติครบทุกรายการ)** และ **`IsConfirm=1` เมื่อ Σ RemainQty ≠ Σ Qty (อนุมัติแล้วอย่างน้อยบางส่วน)** — ยกเลิก/ลบใบอนุมัติ → คำนวณใหม่ (เริ่มจาก `BillStatus=0, IsConfirm=0, RemainQty=Qty`).

**A. หัวเอกสาร (y,x ตาม rc)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบเสนอซื้อ: | รันเลข `BCStkRequest.DocNo` + ปุ่ม (76,4) | ใช่ | ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่ | `:199-202,211` |
| 2 | วันที่ออกเอกสาร: | วันที่ (231,4) | ใช่ | = วันทำงาน; ล็อกตามสิทธิ์แก้วันที่; งวดปิด → ห้ามบันทึก/ลบ | `:203-205,212,394` |
| 3 | ปุ่มซองจดหมาย (`IDC_BTTN_SENDMAIL`, ไอคอนส่งเมล) | ปุ่ม (312,4) | — | ส่งอีเมลถึงผู้จัดการ สรุปใบเสนอซื้อทุกใบที่ `BillStatus=0 AND IsCancel=0` (เลขที่ + ผู้ขออนุมัติ + แผนก); ไม่มีรายการ → `NO_DOCMENT_APPROVE` · ไม่ต้องทำก่อน (กลไกส่งเมล `SentEmailApprove` ยังไม่ verify) | `:1901-1945` |
| 4 | แผนก: | Combo-table `BCDepartment` (76,16) | — | ค่าตั้งต้นของแผนกในแต่ละแถวสินค้า | `:207,662` |
| 5 | ผู้ขออนุมัติ: | Combo-table `BCSale` (231,16) | — | บันทึกที่ `WorkMan` | `:208,725` |
| 6 | กลุ่มเอกสาร: | ข้อความ (76,28) | — | บันทึกตัดที่ 10 ตัวอักษร (`BillGroup`) | `:1692,566` |
| 7 | หมายเหตุ: | ข้อความ (76,40) | — | `MyDescription` | `:722` |

**B. กริดรายการสินค้า (8 คอลัมน์ `:152-159`, ไม่มีช่องราคา/ส่วนลด/ยอดเงิน)**

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | กติกา |
|--|--|--|--|--|
| 0 | รหัสสินค้า | 80 | แก้ได้+Link | ตรวจ `UpdateAndCheckStockOption` ไม่พบ → `NOT_SEEITEMCODE`; `.` = แถวหมายเหตุ |
| 1 | ชื่อสินค้า | 120 | ล็อก (แก้ได้เมื่อ `IsEditItemName=1`) | จากสินค้า |
| 2 | หน่วยนับ | 70 | Combo-table | สินค้าหลายหน่วยกรองจาก `BCStkPacking`; ว่างตอนบันทึก → `NOT_SET_UNITCODE` |
| 3 | จำนวน | 80 | แก้ได้ AUTOMONEY2 ขวา | **0 ไม่ได้** → `QTY_ZERO_LIST` (เลื่อนเคอร์เซอร์ไปช่องนั้น) ยกเว้นแถว `.` |
| 4 | ประเภทการขอ (`STATUS_REQUEST`) | 100 | Combo (ตามลำดับ) | 0 = **ขอแบบปกติ** (`STATUS_REQUEST2`), 1 = **ขอแบบด่วน** (`STATUS_REQUEST1`) → `BCStkRequestSub.Priority`; ว่างถ้าไม่มีรหัสสินค้า |
| 5 | จำนวนวันที่ต้องการ (`WANT_DATE`; ป้าย ~"ต้องการใน(วัน)") | 80 | แก้ได้ ขวา | แก้ → วันที่ต้องการ (คอลัมน์ 6) = วันที่เอกสาร + จำนวนวัน (`CalcDate`) → `WantDay` |
| 6 | วันที่ต้องการ (`MEET_DATE`) | 80 | วันที่ (ไทย) | แก้ → จำนวนวัน = (วันที่ต้องการ − วันที่เอกสาร) − 1 → `WantDate` |

> shadow 21 ช่อง (`:164-184`): 0 RowOrder · 1 ItemType(StockType) · 2 MyDescription · 3 ProjectCode · 4 AllocateCode · 5/6 OppositeQty/Unit · 7 ExceptTax · 8 BarCode · 9–13 หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ · 14 WHCode · 15 ShelfCode · 16 DepartCode · 17–19 สี/ขนาด/แบบ · 20 BehindIndex. **ไม่มีคอลัมน์ "ราคา" "เจ้าหนี้" "คลัง"** บนจอ — แค่ shadow.

**C. คีย์ลัด** (`PreTranslateMessage` `:1610-1700`): Shift+F2 รายละเอียดสินค้าแถว (`OnExpand`: หมายเหตุ/โครงการ/การจัดสรร) · F5 ไม่มีผล (ต้องมีเลขที่) — **Ctrl+F5 ยกเลิก** (ถ้า `IsConfirm=1` → `CANNOT_CANCEL`), **Shift+F5 ยืนยัน** (สิทธิ์ `bCanConfirmBill`) · F7 กลืนไว้ · Shift+F8 หน่วยขนาน (`OnOppositQty`) · **F12 ตรวจเอกสารปลายทาง** (`OnDocumentFlow` `:1880`: รายการ `BCReqConfirmSub` ที่ `StkRequestNo = เลขที่นี้` — คอลัมน์ วันที่/เลขที่ใบอนุมัติ/รหัสสินค้า/หน่วย/จำนวนที่อนุมัติ) · PageUp → ช่องกลุ่มเอกสาร · PageDown → กริด · เมนู `OnMore` มี F12 และ F5 อื่น ๆ · Recurring (`OnRecur`, `IDC_RECUR`) ใช้ได้.

**D. กติกาก่อนบันทึก (`ConfirmSaveData` `:385-531`)**: เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · งวดปิด · (แก้ไข) ยกเลิกแล้ว `NOT_SAVE_ISCANCEL` / ยืนยันแล้ว `NOT_SAVE_ISCONFIRM` / **`BillStatus=1` (ถูกอนุมัติแล้ว) `CANNOT_SAVE_BILLSTATUS_REF`** · (ใบที่ถูกอนุมัติแม้เพียงบางรายการจะมี `IsConfirm=1` จึงติด `NOT_SAVE_ISCONFIRM`/`NOT_DELETE_ISCONFIRM` ตั้งแต่อนุมัติครั้งแรก — ไม่ต้องรอ `BillStatus=1`) · **ไม่มีรายการสินค้าเลย `CANNOT_SAVE_NOTITEM`** · จำนวน = 0 `QTY_ZERO_LIST` · หน่วยว่าง `NOT_SET_UNITCODE` (ต่อท้าย `NOT_SAVE_FORMITEM`) · ตรวจเลขที่ซ้ำ (`IsDuplicate` → ถาม "ลองรันเลขใหม่ไหม"). (มีโค้ดตรวจ "ถูกนำไปอนุมัติบางส่วนแล้ว → `CANNOT_EDIT_APPROVE`" `:410-440` แต่อยู่ใน `else if(nIsBillStatus==1)` ที่ไม่มีทางถึง — เป็นโค้ดตาย.)

**E. ลบ / ยกเลิก**

| การกระทำ | กติกา | ผล |
|--|--|--|
| ลบ (`ConfirmDeleteData` `:290-347`) | เลขที่ว่าง `CANNOT_DELETE` · งวดปิด · ยกเลิกแล้ว/ยืนยันแล้ว · **`BillStatus=1` ห้าม `CANNOT_DELETE_BILLSTATUS_REF`** · ถามยืนยัน | `DELETE BCStkRequest` + `DELETE BCStkRequestSub` + logbook `DELETE_REQUISITION_PO` `:366-383` |
| ยกเลิก (`OnConfirmCancel` `:1947-1985` → `BillCancelUpdateData` `:349-364`) | ยกเลิกแล้ว `MSG_CANNOT_CANCEL`; `BillStatus=1` ห้าม `CANNOT_CANCEL` | `UPDATE BCStkRequest SET IsCancel=1` + `BCStkRequestSub.IsCancel=1` + logbook `CANCEL_REQUISITION_PO` |

**F. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:533-678`)**: `StartTransaction` → `InsertDB`/`UpdateDB(BCStkRequest)` (DocNo, DocDate, MyDescription, DepartCode, BillGroup, WorkMan, RecurName | สร้าง: CreatorCode/CreateDateTime | แก้: LastEditorCode/LastEditDateT) → `DELETE BCStkRequestSub` → `INSERT` ต่อแถว (DocNo, DocDate, ItemCode, ItemName, UnitCode, **Qty = RemainQty = จำนวน**, Priority, WantDay, WantDate, DepartCode(แถว หรือ แผนกหัว), ItemType, MyDescription (`.` ใช้ชื่อ), ProjectCode, AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ, สี/ขนาด/แบบ, TransState=0, IsCancel=0, PackingRate1/2, behindindex, LineNumber) → ถ้าครบ `Commit` ไม่ครบ `Abort` → logbook `SAVE_REQUISITION_PO`. **ไม่มี `processdoc`/GL/สต็อก/ภาษี**. คอลัมน์ `BillStatus` และ `RemainQty` ถูกปรับโดยใบอนุมัติ (§1.11) ผ่าน `UpdateDocumentFlow`.

### 1.11 รายการ 5805 "บันทึกอนุมัติใบเสนอซื้อสินค้า" (ใบอนุมัติเสนอซื้อ — เลือกรายการจากใบเสนอซื้อที่ยังค้าง แล้วระบุจำนวนที่อนุมัติ)

- **ฟอร์ม**: `CPOFrmApprovePurchase` (`pofrmapprovepurchase.cpp` 1,291 บรรทัด; `MainFrm.cpp` `OnPosysTrans2`; dialog เดียว `IDD_PO_FRM_APPROVE_PURCHASE` 349x239 `BC5Account.rc:2173`).
- **ตาราง**: หัว `BCReqConfirm` (`DocNo, DocDate, ConfirmName(ผู้อนุมัติ), MyDescription, BillStatus`) + รายการ `BCReqConfirmSub` (`DocNo, DocDate, ItemCode, UnitCode, DepartCode, Qty(ที่ขอ), ConfirmQty(ที่อนุมัติ), RemainQty(= ConfirmQty ตอนบันทึก), StkRequestNo(เลขที่ใบเสนอซื้อ), ItemType, MyDescription, OppositeUnit/Qty, ExceptTax, BarCode, RefLineNumber, LineNumber, TransState=0`). ไม่มียกเลิก (`IsCancel`) — **ลบอย่างเดียว**; ไม่มี GL/สต็อก/ภาษี/`processdoc`.
- **ธรรมชาติของเอกสาร**: เป็นเอกสาร "เลือกแล้วบันทึกครั้งเดียว" — เมื่อโหลดเอกสารที่บันทึกแล้วขึ้นมา (`m_bLoadData=TRUE`) **บันทึกซ้ำไม่ได้** (`LIST_APPROVE` ~ "รายการนี้โหลดมาเพื่อดู ไม่สามารถอนุมัติซ้ำ") ต้องลบแล้วทำใหม่.

**A. หัวเอกสาร (y,x ตาม rc; จัดเป็นกล่อง 2 กลุ่ม — ซ้ายเป็นปุ่ม, ขวาเป็นช่องกรอก)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | ปุ่ม **รายการขออนุมัติซื้อ** (`ITEM_APPROVE_PHSE`) (14,20) | ปุ่ม | — | = F3 ที่ถือ Ctrl → เปิดตัวเลือกใบเสนอซื้อ (`POStkRequsetConfig`, หลายใบ) กรอง `IsCancel=0 AND BillStatus=0` และไม่ซ้ำกับใบที่เลือกไว้แล้ว | `:606-700,1157-1161` |
| 2 | ปุ่ม **เลือกรายการทั้งหมด** (`SELECT_ITEM_ALL`) (14,43) | ปุ่ม | — | = Shift+F11 → ติ๊กทุกแถวที่มีรหัสสินค้า | `:760-770,1164-1170` |
| 3 | ปุ่ม **อนุมัติรายการ** (`ITEM_APPROVE`) (14,67) | ปุ่ม | — | = F11 → `OnRecSave` (บันทึก) | `:773-776` |
| 4 | เลขที่ใบอนุมัติเสนอซื้อ: | รันเลข `BCReqConfirm.DocNo` + ปุ่ม (164,14) | ใช่ | ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่ (ระบบรันเลขให้); ยาวเกิน → `DOCNO_MORE_LENG` | `:145-148,1231` |
| 5 | วันที่ออกเอกสาร: | วันที่ (164,26) | ใช่ | = วันทำงาน; ล็อกตามสิทธิ์แก้วันที่; งวดปิด → ห้ามบันทึก/ลบ | `:149-151,282` |
| 6 | ผู้อนุมัติ: | Combo-table `BCSale` (164,38) | ใช่ (`CANNOT_APPROVE_NAME`) | → `ConfirmName` | `:153-154,307` |
| 7 | หมายเหตุ: | ข้อความ (164,50) | — | `MyDescription` | `:789` |
| 8 | ผู้บันทึกรายการ: | ข้อความ ล็อก (164,73) | — | = ผู้ใช้ปัจจุบัน (แสดงอย่างเดียว) | `:517` |

**B. กริดรายการ (9 คอลัมน์ `:113-121`, ID `ID_PO_APPROVELIST`)**

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | ความหมาย |
|--|--|--|--|--|
| 0 | `A` (ช่องติ๊ก) | 50 | ติ๊กได้ | อนุมัติแถวนี้ — **วางเคอร์เซอร์บนแถวที่มีรหัสสินค้า = ติ๊กให้อัตโนมัติ** (`OnListChangeSelect`); เลือกครบด้วย Shift+F11 |
| 1 | รหัสสินค้า | 90 | ล็อก | จาก `BCStkRequestSub` (แถว `.` = หมายเหตุ ถูกติ๊กอัตโนมัติและบันทึกเสมอ) |
| 2 | ชื่อสินค้า | 130 | ล็อก | |
| 3 | หน่วยนับ | 70 | ล็อก | แสดง `รหัส/ชื่อหน่วย` |
| 4 | แผนก (`DEPART_APPROVE`) | 100 | ล็อก | แสดง `รหัส/ชื่อแผนก` |
| 5 | เลขที่ใบเสนอซื้อ (หัวคอลัมน์ในระบบ "เลขที่อนุมัติเสนอซื้อ" `DOCNO_APPROVE` แต่ข้อมูลคือเลขที่ **ใบเสนอซื้อ**) | 90 | ล็อก | → `StkRequestNo` |
| 6 | จำนวนขออนุมัติ (`QTY_APPROVE`) | 100 | ล็อก ขวา | = `RemainQty` คงเหลือของใบเสนอซื้อ → `Qty` |
| 7 | จำนวนที่อนุมัติ (`QTY_REQUEST`) | 90 | **แก้ได้** AUTOMONEY2 ขวา | ค่าตั้งต้น = จำนวนคงเหลือ; **ไม่มีการจำกัดไม่ให้เกินจำนวนขอ**; ติ๊กแต่ = 0 → ไม่บันทึกแถวนั้น → `ConfirmQty`/`RemainQty` |
| 8 | `S` (สถานะ) | 30 | ล็อก กลาง | `**` = รายการที่ขอแบบด่วน (`Priority=1`) |

> shadow 11 ช่อง (`:122-135`): 0 RowOrder · 1 RefDocNo(เลขที่ใบเสนอซื้อ) · 2 ItemType · 3 MyDescription · 4 OppositeUnit · 5 OppositeQty · 6 ExceptTax · 7 BarCode · 8 UnitCode · 9 DepartCode · 10 LineNumber(ของใบเสนอซื้อ → `RefLineNumber`). ดึงรายการต่อใบเสนอซื้อ: `BCStkRequestSub WHERE TransState=0 AND IsCancel=0 AND (RemainQty>0 OR ItemCode='.') ORDER BY DocNo,LineNumber`.

**C. คีย์ลัด** (`PreTranslateMessage` `:1150-1180`; เมนู `OnMore` `:1180-1230`): **Ctrl+F3 รายการขออนุมัติซื้อ** · **F11 อนุมัติรายการ (บันทึก)** · **Shift+F11 เลือกรายการทั้งหมด** · F7 กลืนไว้ · **F12 ตรวจเอกสารปลายทาง** (`OnDocumentFlow` `:1236`: ใบสืบราคาสินค้ารวม `BCEnquirePrice.ReqConfirmNo=เลขที่นี้` + ใบสั่งซื้อ `BCPurchaseOrderSub.ConfirmNo=เลขที่นี้ AND IsCancel=0` — คอลัมน์ วันที่/เลขที่/ประเภทเอกสาร/รหัสสินค้า/หน่วย/จำนวน) · PageDown → กริด · ปุ่มพิมพ์/ใหม่/ค้นหา (`ID_HK_PRINT/NEW/FIND`).

**D. กติกาก่อนบันทึก (`ConfirmSaveData` `:269-378`)**: งวดปิด · โหลดมาดู `LIST_APPROVE` · ไม่ได้ติ๊กแถวใด `CANNOT_APPROVE` · เลขที่ว่าง `CANNOT_APPROVE_NO` · วันที่ว่าง `CANNOT_APPROVE_DATE` · ผู้อนุมัติว่าง `CANNOT_APPROVE_NAME` · เลขที่ซ้ำ → ถามรันเลขใหม่.
**ลบ** (`ConfirmDeleteData` `:215-227`): เฉพาะ งวดปิด + ถามยืนยัน — **ไม่ตรวจว่ามีใบสืบราคา/ใบสั่งซื้ออ้างอิงอยู่หรือไม่** → `OnDeleteData` (`:229-267`): เก็บเลขที่ใบเสนอซื้อที่อ้างทั้งหมด → `DELETE BCReqConfirm` + `DELETE BCReqConfirmSub` → logbook `DELETE_APPROVE_PO` → **`UpdateDocumentFlow(-1,16,ใบเสนอซื้อ)`** (คืนจำนวนค้าง/สถานะของใบเสนอซื้อ).

**E. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:380-503`)**: `InsertDB`/`UpdateDB(BCReqConfirm)` (DocNo, DocDate, ConfirmName, MyDescription, **BillStatus=0**, CreatorCode/CreateDateTime | LastEditorCode/LastEditDateT) → ต่อแถวที่ติ๊กและ `ConfirmQty>0` (หรือแถว `.`): `INSERT BCReqConfirmSub` → logbook (`20003`) → **`UpdateDocumentFlow(+1,16,รายการเลขที่ใบเสนอซื้อคั่น ,)`** (`mytools.cpp:2179-2232`: คำนวณ `BCStkRequestSub.RemainQty = Qty − Σ ConfirmQty`, ตั้ง `BCStkRequest.BillStatus=1` เมื่อ Σ RemainQty ≤ 0, `IsConfirm=1` เมื่อมีการอนุมัติแล้วบางส่วน) → ล้างจอ. (ฟังก์ชัน `OnUpdateRequest` `:555-604` ที่เคยทำเรื่องเดียวกันถูก comment ทิ้ง `:489`.) โหมดปูนไทย (`m_bChmMode`) แถวที่ไม่ติ๊ก → ตัดรายการนั้นออกจากใบเสนอซื้อ (`RemainQty=0, IsCancel=1`) — ข้าม.
**สถานะของใบอนุมัติเอง**: `BCReqConfirm.BillStatus` ถูก `UpdateDocumentFlow(±1,15,เลขที่ใบอนุมัติ)` ตั้งเป็น 1 เมื่อทุกรายการถูกนำไปออกใบสั่งซื้อครบ (`RemainQty = ConfirmQty − Σ POSub.Qty` โดย `BCPurchaseOrderSub.ConfirmNo, ItemCode, RefLineNumber` `mytools.cpp:2128-2176`) → ใบอนุมัติที่ใช้ครบแล้วจะไม่โผล่ในรายการให้เลือกของใบสั่งซื้อ (§1.1).

### 1.12 รายการ 5802 "บันทึกใบสืบราคาสินค้ารวม" (ใบสืบราคา — เลือกใบอนุมัติเสนอซื้อ แล้วติ๊กเลือกเจ้าหนี้ที่จะขอราคาต่อสินค้า)

- **ฟอร์ม**: `CPOFrmInquirePrice` (`POFrmInquirePrice.cpp` 1,474 บรรทัด; `MainFrm.cpp` `OnPosysTrans15`; dialog เดียว `IDD_PO_FRM_INQNIRE_PRICE` 350x242 `BC5Account.rc:2060`; กริด ID `ID_PO_INQUIRE`).
- **ตาราง**: หัว `BCEnquirePrice` (`DocNo, DocDate, ReqConfirmNo(เลขที่ใบอนุมัติเสนอซื้อ), MyDescription, BillStatus, CreatorCode/CreateDateTime | LastEditorCode/LastEditDateT`) + รายการ `BCEnquirePriceSub` (`DocNo, DocDate, ItemCode, UnitCode, Qty, DepartCode, StkRequestNo, SupplierCode, ItemType, MyDescription, OppositeUnit/Qty, ExceptTax, BarCode, TransState=0`). **ไม่มี GL / สต็อก / ภาษี / `processdoc` / ราคา** — เป็นเอกสารเตรียมข้อมูลก่อนทำตารางเปรียบเทียบราคา (§1.13).
- **ห่วงโซ่เอกสาร** (ผูกด้วยสถานะ `BillStatus` เพื่อซ่อนใบที่ถูกใช้แล้วออกจากตัวเลือก): ใบอนุมัติ `BCReqConfirm` (§1.11) **→ บันทึกใบสืบราคา: ตั้ง `BCReqConfirm.BillStatus=1` ทันที** (`OnSaveData` `:1216`) → ใบตารางราคารวม `BCBuyPriceCondit` (§1.13) **→ บันทึก: ตั้ง `BCEnquirePrice.BillStatus=1`** → ประมวลใบสั่งซื้อ (§1.14). ลบใบสืบราคา = คืน `BCReqConfirm.BillStatus=0` (`OnDeleteData` `:1032`).

**A. หัวเอกสาร (y,x ตาม rc)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบสืบราคา: | รันเลข `BCEnquirePrice.DocNo` + ปุ่ม (14,81) | ใช่ | รันเลขอัตโนมัติ; ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่ (`bCanChangeBillNo`) แล้วโฟกัสไปช่องที่ 3; ยาวเกิน → `DOCNO_MORE_LENG`; ซ้ำ → ถาม "ลองรันเลขใหม่ไหม" | `:140-143,160-185,1396` |
| 2 | วันที่ออกเอกสาร: | วันที่ไทย (14,245) | ใช่ | = วันทำงาน; **ล็อก** ถ้าผู้ใช้ไม่มีสิทธิ์แก้วันที่ (`IsCanEditDocDate=0` และไม่ใช่ `sa`) | `:144-156` |
| 3 | เลขที่ใบอนุมัติเสนอซื้อ: | Link + ปุ่ม (26,81) | ใช่ | เปิดตัวเลือก `ApprovePurchaseConfig` กรอง `BillStatus = 0` (ใบอนุมัติที่ยังไม่ถูกใช้) → ดึงรายการ (ดู B) | `:147-148,795-971` |
| 4 | วันที่อนุมัติ | วันที่ ล็อก (26,245) | — | จากใบอนุมัติ (`BCReqConfirm.DocDate`) | `:817,1242` |
| 5 | ผู้อนุมัติ: | ข้อความ ล็อก (38,81) | — | จากใบอนุมัติ (`ConfirmName`) | `:818,1244` |
| 6 | หมายเหตุ: | ข้อความ (50,81) | — | `MyDescription`; แก้ได้หลังบันทึก (`AddCanEditField`) | `:150,1246` |

> ไม่มีช่องผู้ขอ/แผนก/กลุ่มเอกสารที่หัวเอกสาร — แผนกอยู่ระดับรายการสินค้า.

**B. กริด (11 คอลัมน์ `:107-117`; คอลัมน์ 0-4 ล็อกทั้งหมด, แก้ได้เฉพาะช่องติ๊ก "S")**

| # | หัวคอลัมน์ | กว้าง | ชนิด | ความหมาย |
|--|--|--|--|--|
| 0 | รหัสสินค้า | 80 | ล็อก | จาก `BCReqConfirmSub` (`TransState=0` เรียงตาม `LineNumber`) |
| 1 | ชื่อสินค้า | 100 | ล็อก | `BCItem.Name1` |
| 2 | หน่วยนับ | 70 | ล็อก | แสดง `รหัส/ชื่อหน่วย` |
| 3 | แผนก | 70 | ล็อก | แสดง `รหัส/ชื่อแผนก` |
| 4 | จำนวน | 60 | ล็อก ขวา | = `ConfirmQty` ที่อนุมัติ (ไม่ใช่ `RemainQty`) |
| 5 | `S` | 50 | **ช่องติ๊ก** (`FCS_CHECK`) | ติ๊ก = ขอราคาจากเจ้าหนี้แถวนี้ — ค่าเริ่มต้นไม่ติ๊ก (`FALSE`) ตอนดึงจากใบอนุมัติ, `TRUE` ตอนเปิดใบที่บันทึกแล้ว |
| 6 | รหัสเจ้าหนี้ | 80 | ล็อก | `BCItemSupplier.ApCode` |
| 7 | ชื่อเจ้าหนี้ | 110 | ล็อก | ตามตั้งค่า `DefaultApName` (ชื่อ 1/2) |
| 8 | วันที่ซื้อล่าสุด | 80 | ล็อก | ตอนดึงจากใบอนุมัติ: `BCApInvoiceSub` แถวล่าสุดของเจ้าหนี้+สินค้านั้น (`TransState=0 AND IsCancel=0 ORDER BY DocDate DESC, LineNumber DESC`); ตอนเปิดใบเดิม: `BCItemSupplier.LastBuyDate` |
| 9 | ราคาซื้อล่าสุด | 80 | ล็อก | ตอนเปิดใบเดิม: ราคาโปรโมชันถ้ามี (`GetItemBuyPrice(ApCode,Item,Unit,วันที่)`) ไม่งั้น `LastBuyPrice` ไม่งั้น `BuyPrice` |
| 10 | หน่วยซื้อล่าสุด | 80 | ล็อก | แสดง `รหัส/ชื่อหน่วย` |

**วิธีจัดแถว (ต้องทำเหมือน):** สินค้า 1 ตัว (ต่อแผนก) → **1 แถวต่อเจ้าหนี้ที่ขายสินค้านั้น** (`BCItemSupplier WHERE ItemCode=… ORDER BY ApCode, PriorityIndex`); คอลัมน์ 0-4 แสดงเฉพาะแถวแรกของสินค้า/แผนกนั้น แถวต่อ ๆ ไปเว้นว่าง (กริดแบบจัดกลุ่ม); ถ้าสินค้าไม่มีเจ้าหนี้ในทะเบียน → แถวเดียวคอลัมน์ 5-10 ว่าง (ติ๊กไม่ได้ผล เพราะไม่มีรหัสเจ้าหนี้).

> shadow 12 ช่อง (`:123-134`): 0 RowOrder · 1 ItemType · 2 MyDescription · 3 OppositeUnit (จาก `BCItem`) · 4 OppositeQty · 5 ExceptTax (จาก `BCItem`) · 6 BarCode · 7 ItemCode · 8 UnitCode · 9 Qty(= ConfirmQty) · 10 DepartCode · **11 = เลขที่ใบอนุมัติ** (`BCReqConfirmSub.DocNo` — ถูกบันทึกลง `BCEnquirePriceSub.StkRequestNo`; ชื่อคอลัมน์ในตารางชวนเข้าใจผิดว่าเป็นเลขที่ใบเสนอซื้อ แต่ค่าจริงคือเลขที่ใบอนุมัติ).

**C. คีย์ลัด** (`PreTranslateMessage` `:973-1005`; เมนู `OnMore` `:1402`): **F12 ตรวจเอกสารปลายทาง** (`OnDocumentFlow` `:1424`: ใบตารางราคารวม `BCBuyPriceCondit WHERE RefNo = เลขที่นี้` — คอลัมน์ วันที่/เลขที่/หมายเหตุ) · PageUp → ช่องหมายเหตุ · PageDown → กริดแถว 0 · F7 กลืนไว้ · พิมพ์ (`ID_HK_PRINT` → `OnCheckBeforePrint` + `OnPrintForm` จาก `BCEnquirePrice`/`BCEnquirePriceSub`) · ใหม่/ค้นหา (`ID_HK_NEW/FIND`). ไม่มี F5/Ctrl+F5 (ไม่มียกเลิก/ยืนยัน) และไม่มีกล่องรายละเอียดรายการ.

**D. กติกาก่อนบันทึก (`ConfirmSaveData` `:1052-1136`)**: เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · (เพิ่มใหม่) เลขที่ซ้ำ → ถามรันเลขใหม่ · ตรวจ "ติ๊กแต่ไม่มีเจ้าหนี้" → `SELECT_AP_BYITEM` (~"กรุณาเลือกเจ้าหนี้ให้ครบตามรายการสินค้า").
**ไม่มีการบังคับให้ติ๊กอย่างน้อย 1 แถว** — บันทึกใบที่ไม่มีรายการเลยได้ (แต่จะตั้งใบอนุมัติเป็น "ใช้แล้ว" ไปด้วย).
**ลบ** (`ConfirmDeleteData` `:1006-1023`): เลขที่ว่าง `CANNOT_DELETE_DOCNO` · **ถูกนำไปทำใบตารางราคารวมแล้ว (`BCEnquirePrice.BillStatus=1`) ห้ามลบ `CANNOT_ENQNO`** · ถามยืนยัน → `OnDeleteData` (`:1024-1050`): `UPDATE BCReqConfirm SET BillStatus=0` (คืนสถานะใบอนุมัติ) + `DELETE BCEnquirePrice` + `DELETE BCEnquirePriceSub` + logbook `CANCEL_INQUIRED`. ไม่มี `IsCancel` — ลบอย่างเดียว.

**E. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:1138-1224`)**: `InsertDB`/`UpdateDB(BCEnquirePrice)` (`BillStatus=0` ทุกครั้งที่บันทึก และ `ConfirmSaveData` **ไม่ตรวจ** `m_nBillStatus` — จึงแก้ใบที่ถูกใช้ทำตารางราคาแล้วได้และรีเซ็ตสถานะกลับเป็นยังไม่ใช้ ดูข้อสังเกต 3) → แก้ไข: `DELETE BCEnquirePriceSub` เดิม → **เขียนเฉพาะแถวที่ติ๊กและมีรหัสเจ้าหนี้** (`INSERT BCEnquirePriceSub`: ฟิลด์ตามตารางด้านบน, `Qty` = จำนวนที่อนุมัติ) → `UPDATE BCReqConfirm SET BillStatus=1 WHERE DocNo=ใบอนุมัติ` → logbook `SAVE_INQUIRED`. **ไม่เรียก `UpdateDocumentFlow`** และไม่ใช้ transaction ครอบ (เป็นชุดคำสั่งแยก — ในเว็บควรครอบ transaction เดียว).
**เปิดใบเดิม** (`OnLoadData` `:1225-1375`): แสดงเฉพาะแถวที่บันทึกไว้ (ติ๊กครบทุกแถว) เรียง `ItemCode, UnitCode, SupplierCode` — **ไม่แสดงเจ้าหนี้อื่นที่ไม่ได้ติ๊ก** จึงแก้ใบเดิมให้ "เพิ่มเจ้าหนี้" ไม่ได้ ต้องลบแล้วทำใหม่.

**F. ข้อสังเกตจากโค้ดเดิม (ไม่ใช่ข้อกำหนด)**
1. ตัวตรวจ `SELECT_AP_BYITEM` (`:1060-1068`) อ่านคอลัมน์ผิด — ใช้ `GetData(4)` (จำนวน) เทียบ "TRUE" และ `GetData(5)` (ช่องติ๊ก) เป็นรหัสเจ้าหนี้ → เงื่อนไขไม่มีวันเป็นจริง = **ตัวป้องกันไม่ทำงานจริง**; เจตนาคือ "แถวที่ติ๊กต้องมีรหัสเจ้าหนี้" (ถ้าทำใหม่ให้ตรวจที่คอลัมน์ 5/6).
2. การเลือก "ทั้งหมด" ไม่มีปุ่ม — ต้องติ๊กทีละแถว.
3. บันทึกซ้ำ (แก้ไข) ตั้ง `BillStatus=0` ทุกครั้งโดยไม่ตรวจว่าใบสืบราคาถูกนำไปทำตารางราคารวมแล้วหรือยัง (ต่างจากการลบที่ห้าม) — เป็นช่องโหว่ของเดิม; BC ควรห้ามแก้เมื่อ `BillStatus=1` เหมือนกติกาลบ.
4. `OnListChangeSelect` (`:758-792`) ที่จัดการช่องติ๊กอัตโนมัติ **ไม่ได้ผูกเข้า message map** (`:59-75`) จึงไม่ทำงาน — ไม่ต้องทำตาม.
5. BC ควรให้ผลเท่ากัน: บันทึกใบสืบราคา → ใบอนุมัติหายจากรายการให้เลือกของใบสืบราคา/ใบสั่งซื้อ; ลบใบสืบราคา → คืนสถานะ.

### 1.13 รายการ 5809 "บันทึกใบตารางราคารวม" (ตารางเปรียบเทียบราคาเจ้าหนี้ — เลือกใบสืบราคา แล้วกรอกจำนวน/ราคา/ส่วนลดที่จะซื้อจากแต่ละเจ้าหนี้)

- **ฟอร์ม**: `CPOFrmPriceCondition` (`POFrmPriceCondition.cpp` 1,448 บรรทัด; `MainFrm.cpp` `OnPosysTrans3`; dialog เดียว `IDD_PO_FRM_PRICECONDIT` 350x241 `BC5Account.rc:1598`; กริด ID `ID_POPRICECONDITGRID`; กล่องรายละเอียดแถว `IDD_PO_DLG_EXPAND_PRICE` 215x114 `BC5Account.rc:2594`).
- **ตาราง**: หัว `BCBuyPriceCondit` (`DocNo, RefNo(เลขที่ใบสืบราคา), DocDate, MyDescription, BillStatus, CreatorCode/CreateDateTime | LastEditorCode/LastEditDateT`) + รายการ `BCByPriceCondSub` (`DocNo, DocDate, ItemCode, UnitCode, Qty(จำนวนที่สืบราคามา), ApCode, BuyQty, BuyPrice, DiscountWord, MyDescription, CreditDay, TrasportDay, TransportCondit, DepartCode, StkRequestNo(= เลขที่ใบอนุมัติ), RowOrder, CreatePoFlag, IsCancel`). ไม่มี GL / สต็อก / ภาษี / `processdoc` — เป็นเอกสาร "ตัดสินใจซื้อ" ที่ป้อนให้ §1.14.
- **สถานะ**: เมื่อ**บันทึก** ตั้ง `BCEnquirePrice.BillStatus=1` (ใบสืบราคาถูกใช้แล้ว หายจากตัวเลือก; `OnSaveData` `:479`); เมื่อ**ลบ** คืน `BillStatus=0` (`:273`). สถานะของใบตารางราคาเอง `BCBuyPriceCondit.BillStatus` ถูกตั้ง `1` โดยจอ §1.14 เมื่อทุกแถวถูกนำไปออกใบสั่งซื้อครบ (`CreatePoFlag=1` ทุกแถว) — ใบที่ `BillStatus=1` **ลบไม่ได้** (`CANNOT_DELETE`).

**A. หัวเอกสาร (y,x ตาม rc)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบราคารวม: | รันเลข `BCBuyPriceCondit.DocNo` + ปุ่ม (7,84) | ใช่ | รันเลขอัตโนมัติ; ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่ แล้วโฟกัสช่องใบสืบราคา; ยาวเกิน → `DOCNO_MORE_LENG` | `:146-149,168-193,1369` |
| 2 | วันที่ออกเอกสาร: | วันที่ไทย (7,248) | ใช่ | = วันทำงาน; ล็อกถ้าไม่มีสิทธิ์แก้วันที่ | `:150-162` |
| 3 | เลขที่ใบสืบราคา: | Link + ปุ่ม (19,84) | ใช่ | เปิดตัวเลือก `POFrmInquirePriceConfig` กรอง `BillStatus = 0` → ดึงรายการ (ดู B) | `:154-155,1218-1348` |
| 4 | วันที่ออกใบสืบราคา: | วันที่ ล็อก (19,248) | — | จากใบสืบราคา | `:524,1240` |
| 5 | หมายเหตุ: | ข้อความ (31,84) | — | `MyDescription` | `:525` |

**B. กริด (10 คอลัมน์ `:115-124`)** — ดึงจาก `BCEnquirePriceSub WHERE TransState=0 ORDER BY ItemCode, SupplierCode` (1 แถวต่อ สินค้า × เจ้าหนี้ที่เลือกไว้ในใบสืบราคา):

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | กติกา |
|--|--|--|--|--|
| 0 | รหัสสินค้า | 70 | ล็อก | แสดงเฉพาะแถวแรกของสินค้า/แผนกนั้น (แถวถัดไปเว้นว่าง) |
| 1 | ชื่อสินค้า | 100 | ล็อก | |
| 2 | หน่วยนับ | 70 | ล็อก | `รหัส/ชื่อ` |
| 3 | แผนก | 70 | ล็อก | `รหัส/ชื่อ` |
| 4 | จำนวน | 60 | ล็อก ขวา | = จำนวนที่สืบราคามา (`Qty`) — เป็น "เพดาน" ของจำนวนซื้อรวมต่อสินค้า |
| 5 | รหัสเจ้าหนี้ | 70 | ล็อก | `SupplierCode` |
| 6 | ชื่อเจ้าหนี้ | 110 | ล็อก | |
| 7 | จำนวนซื้อ | 70 | **แก้ได้** AUTOMONEY2 ขวา | ค่าตั้งต้นว่าง; ผู้ใช้แบ่งจำนวนซื้อให้เจ้าหนี้แต่ละราย → `BuyQty` |
| 8 | ราคาซื้อ/หน่วย | 80 | **แก้ได้** AUTOMONEY2 ขวา | ค่าตั้งต้น: ราคาโปรโมชัน (`GetItemBuyPrice(เจ้าหนี้,สินค้า,หน่วย,วันที่เอกสาร)`) ถ้ามี ไม่งั้น `BCItemSupplier.LastBuyPrice` ไม่งั้น `BuyPrice` (ตรงหน่วยนับ) → `BuyPrice` |
| 9 | ส่วนลด%,บาท | 80 | **แก้ได้** (ข้อความ เช่น `5%`, `100`) | → `DiscountWord` |

> shadow 11 ช่อง (`:130-140`): 0 RowOrder · 1 MyDescription · 2 DiscountCondit · 3 CreditDay · 4 TrasportDay · 5 TransportCondit · 6 ItemCode · 7 UnitCode · 8 EnqQty (จำนวนที่สืบราคา) · 9 Depart · 10 StkRequestNo (ค่าจริง = เลขที่ใบอนุมัติ).
> **กล่องรายละเอียดแถว** (ปุ่มลูกศรท้ายแถว `ILDN_OTHER_BNCKICKED` → `OnExpand` `:1198`; เปิดได้เฉพาะแถวที่มีรหัสเจ้าหนี้): **จ่ายเงินภายใน(วัน)** (`CreditDay`, ตัวเลข) · **ส่งของภายใน(วัน)** (`TrasportDay`, ตัวเลข) · **เงื่อนไขส่วนลด** (`DiscountCondit`) · **เงื่อนไขการขนส่ง** (`TransportCondit`) · **หมายเหตุ** (`MyDescription`) + ปุ่ม OK.

**C. กติกาตอนกรอก (`OnListChangeColumn` `:1016-1071`)**: จำนวนซื้อ < 0 → `PUT_NOT_QTY` ("ใส่จำนวนสินค้ามากกว่า 0") ล้างช่อง · **ผลรวมจำนวนซื้อของสินค้าตัวเดียวกันทุกเจ้าหนี้ต้องไม่เกินจำนวนที่สืบราคา** ไม่งั้น `ITEMCODE <รหัส> SUM_ITEMQTY_MUCH` (~"รวมจำนวนซื้อเกินจำนวนที่สืบราคา") ล้างช่อง · ราคา < 0 → `PUT_PRICE_MUCHZERO` ล้างช่อง. (หมายเหตุ: อนุญาต 0 — แถวจำนวน 0 จะไม่ถูกบันทึก ดู E.)

**D. คีย์ลัด** (`PreTranslateMessage` `:1151-1196`): **Ctrl+F5 ยกเลิกเอกสาร** (`BillCancel("BCBuyPriceCondit")`) · **Shift+F5 ยืนยันเอกสาร** (ต้องมีสิทธิ์ `bCanConfirmBill` ไม่งั้น `NOT_CONFIRM_NO`; `BillConfirm`) — ทั้งคู่ทำได้เมื่อมีเลขที่เอกสาร · Ctrl+Enter ที่ช่องหมายเหตุ → กริด · F7 กลืนไว้ · PageDown → กริด · พิมพ์ (`ID_HK_PRINT` → `OnCheckBeforePrint` + พิมพ์จาก `BCBuyPriceCondit`) · ใหม่/ค้นหา (`ID_HK_NEW/FIND`). **ไม่มี F12 ตรวจเอกสารปลายทาง** (`OnMore`/`OnDocumentFlow` `:1376-1416` ถูก comment ออกจาก message map `:76-78` และใช้ query ของใบสืบราคาที่ copy มา).

**E. กติกาก่อนบันทึก / ลบ**
- `ConfirmSaveData` (`:292-385`): เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · แถวมีรหัสสินค้าแต่ "จำนวน" ว่าง → ข้อความอังกฤษ "Please Enter Qty..." (**ตรวจผิดคอลัมน์** — อ่านคอลัมน์ 6 = ชื่อเจ้าหนี้ ไม่ใช่จำนวนซื้อคอลัมน์ 7 → ไม่ทำงานจริง) · (เพิ่มใหม่) เลขที่ซ้ำ → ถามรันเลขใหม่. **ไม่ตรวจ `m_nBillStatus` ตอนบันทึกแก้ไข**, ไม่บังคับให้มีจำนวนซื้อ > 0 อย่างน้อย 1 แถว.
- ลบ (`ConfirmDeleteData` `:246-264` → `OnDeleteData` `:266-290`): เลขที่ว่าง / `BillStatus=1` → `CANNOT_DELETE`; ถามยืนยัน → `UPDATE BCEnquirePrice SET BillStatus=0` + `DELETE BCBuyPriceCondit` + `DELETE BCByPriceCondSub` + logbook `SUBTRACT_PRICE_TABLE`. (ลบทั้งที่แถวบางส่วนถูกนำไปออกใบสั่งซื้อแล้ว `CreatePoFlag=1` ก็ยังลบได้ตราบที่หัวเอกสาร `BillStatus=0` — ไม่มีการตรวจ.)

**F. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:387-487`)**: `InsertDB`/`UpdateDB(BCBuyPriceCondit)` → ต่อแถวกริด: แถวที่เคยบันทึกแล้ว (มี `RowOrder`) แต่รหัสสินค้าถูกล้าง → `DELETE BCByPriceCondSub WHERE RowOrder`; นอกนั้น **บันทึกเฉพาะแถวที่ `BuyQty > 0`** — แถวใหม่ `INSERT`, แถวเดิม `UPDATE ... WHERE DocNo AND RowOrder` (ฟิลด์ตามตารางด้านบน; `Qty` = จำนวนที่สืบราคา, `StkRequestNo` = เลขที่ใบอนุมัติ) → `UPDATE BCEnquirePrice SET BillStatus=1` → logbook `SAVE_PRICE_TABLE`. ไม่ใช้ transaction ครอบ.
**เปิดใบเดิม** (`OnLoadData` `:510-614`): โหลด `BCByPriceCondSub WHERE DocNo` ทั้งหมด (ไม่กรอง `CreatePoFlag`), แสดงคอลัมน์ตามเดิมพร้อม shadow ของรายละเอียดแถว.

**G. ข้อสังเกตจากโค้ดเดิม (ไม่ใช่ข้อกำหนด)**
1. `DiscountCondit` (เงื่อนไขส่วนลด) กรอกได้ในกล่องรายละเอียดและถูก**โหลด**กลับ (`:588`) แต่คำสั่งบันทึกถูก comment ทิ้ง (`:460`) → **ค่านี้ไม่เคยถูกเก็บ** ถ้าทำใหม่ควรเก็บให้ครบ (ฟิลด์มีอยู่ในตาราง).
2. แถวเดิมที่แก้ `BuyQty` เป็น 0 จะ**ไม่ถูกลบหรืออัปเดต** (เงื่อนไข `BuyQty>0` อยู่หน้าคำสั่งทั้งหมด) → ค่าเก่าค้างในฐานข้อมูล; BC ควรลบแถวที่จำนวนซื้อเป็น 0.
3. ตรวจ "รวมจำนวนซื้อ ≤ จำนวนสืบราคา" ทำเฉพาะตอนกรอก (ไม่ตรวจซ้ำตอนบันทึก) และ**ไม่บังคับ** ว่าต้องซื้อครบ.
4. ราคาในกริดเป็นราคา "ต่อหน่วยตามหน่วยนับของแถว" ไม่มีสกุลเงิน/ภาษี/ยอดเงิน — การคำนวณยอดเงินทั้งหมดเกิดตอนประมวลใบสั่งซื้อ (§1.14).

### 1.14 รายการ 5803 "ประมวลใบผลสั่งซื้อสินค้าอัตโนมัติ" (สร้างใบสั่งซื้อ `BCPurchaseOrder` ทีละหลายใบ จากใบอนุมัติเสนอซื้อ หรือจากใบตารางราคารวม)

- **ฟอร์ม**: `CPOFrmPOProcess` (`POFrmPOProcess.cpp` 2,197 บรรทัด; `MainFrm.cpp` `OnPosysTrans16`; dialog หลัก `IDD_PO_FRM_POPROCESS` 350x241 `BC5Account.rc:610`; กล่องกำหนดรายละเอียดใบสั่งซื้อรายเจ้าหนี้ `IDD_PO_DLG_POPROCESS` 310x229 `BC5Account.rc:2347`; กล่องรายละเอียดรายการ `IDD_PO_DLG_EXPAND_POPROCESS` 189x121 `BC5Account.rc:2615`; กริด ID `ID_POPROCESS_LIST`).
- **ลักษณะเอกสาร**: เป็น "ขั้นตอนประมวลผล" ไม่ใช่เอกสารของตัวเอง — ผลลัพธ์คือ **ใบสั่งซื้อ 1 ใบต่อเจ้าหนี้ 1 ราย** ลงตารางเดียวกับจอบันทึกใบสั่งซื้อ (§1.1): หัว `BCPurchaseOrder` + รายการ `BCPurchaseOrderSub`. **ไม่มีลบ/ยกเลิก/ค้นหาในจอนี้** (ลบหรือแก้ใบสั่งซื้อที่ได้ไปทำที่ §1.1) — ปุ่มใหม่ (`ID_HK_NEW`) ล้างจอ, ปุ่มค้นหา/พิมพ์ผูกไว้แต่ฟังก์ชันพิมพ์ถูก comment ทิ้ง (`OnHKPrint` `:2169`).
- **สองโหมดต้นทาง (ปุ่มเลือกวิทยุ 2 ตัว)**: **(ก) เลือกจากใบอนุมัติเสนอซื้อ** (ค่าเริ่มต้น `m_bApprove=TRUE`, `:610-640`) = ป้ายวิทยุ `«SELECT_BILL_PO»` · **(ข) เลือกจากใบตารางราคารวม** = ป้ายวิทยุ `«SELECT_BILL»` (ข้อความไทยของสองป้ายนี้ยังไม่ verify — คีย์ไม่ถูก resolve). สลับโหมด (`OnCheckApprove` `:1834` / `OnCheckPricecondit` `:1850`): โหมด ก เปิดช่องรหัสเจ้าหนี้+เลขที่ใบอนุมัติ ปิดช่องเลขที่ใบราคารวม (และล้างช่องใบราคารวม); โหมด ข ทำกลับกัน (ปิดช่องรหัสเจ้าหนี้และใบอนุมัติ ล้างเลขที่ใบอนุมัติ) — **ไม่ล้างกริดตอนสลับ**.

**A. ส่วนหัวของจอ (2 กล่อง; y,x ตาม rc)**

| # | ป้ายบนจอ | ช่อง/ชนิด | โหมด | กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | รหัสเจ้าหนี้: | Link + ปุ่ม (18,99) + ชื่อเจ้าหนี้ ล็อก (18,177) | ก | ต้องกรอกก่อนเลือกใบอนุมัติ (ว่าง → `PURCHORDER_APCODE` (~"กรุณากรอกรหัสเจ้าหนี้") แล้วโฟกัสกลับ `:1327-1331`); ปุ่มค้นหา = ตัวเลือก `ApTransConfig` (กรอง `Code LIKE <ที่พิมพ์>%`); ออกจากช่อง → แสดงชื่อ (ตาม `DefaultApName`); **ถูกปิดช่องทันทีเมื่อดึงรายการจากใบอนุมัติสำเร็จ** | `:1140-1167,1819-1832` |
| 2 | เลขที่ใบอนุมัติเสนอซื้อ: (`STK_REQUEST_NO`) | Link + ปุ่ม (30,99) | ก | ตัวเลือก `ApprovePurchaseConfig` กรอง `BillStatus = 0` (ไม่เลือกหลายใบ); หรือพิมพ์เลขแล้วออกจากช่อง → ดึงถ้ามีใบ `BillStatus = 0`; มีเครื่องหมายบังคับกรอก | `:1325-1360,1519-1535` |
| 3 | วันที่ออกเอกสาร (ของใบอนุมัติ) | วันที่ ล็อก (30,251) | ก | = `BCReqConfirm.DocDate` | `:1371,1540` |
| 4 | เลขที่ใบราคารวม: (`PRICE_CONDIT_NO`) | Link + ปุ่ม (63,99) | ข | ตัวเลือก `POBuyPriceConditionConfig` กรอง `BillStatus = 0` — **ต้องเลือกจากตัวเลือกเท่านั้น** (ตัวจัดการ "พิมพ์เลขแล้วออกจากช่อง" `OnKillfocusEdPricecondit` `:1668` มีโค้ดแต่**ไม่ผูก message map** `:74-91`) | `:1168-1180` |
| 5 | วันที่ออกเอกสาร (ของใบราคารวม) | วันที่ ล็อก (63,251) | ข | = `DocDate` ของใบราคารวม | `:1192` |

**B. กริด (9 คอลัมน์ `:122-134`; ซ้ายมีเลขแถวอัตโนมัติ; ปุ่มลูกศรท้ายแถว = เปิดกล่องรายละเอียด)**

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | ความหมาย |
|--|--|--|--|--|
| 0 | รหัส-ชื่อเจ้าหนี้ (`AP_CODE_NAME`) | 180 | ล็อก | **แถวหัวเจ้าหนี้** แสดง `รหัส - ชื่อ` (แถวสินค้าเว้นว่าง) — แต่ละแถวหัว = ใบสั่งซื้อ 1 ใบ |
| 1 | รหัสสินค้า | 70 | ล็อก | |
| 2 | ชื่อสินค้า | 120 | ล็อก | |
| 3 | หน่วยนับ | 70 | ล็อก | `รหัส/ชื่อ` |
| 4 | แผนก | 70 | ล็อก | `รหัส/ชื่อ` |
| 5 | จำนวนซื้อ (`ITEM_BUYQTY`) | 60 | **แก้ได้** AUTOMONEY2 ขวา | ค่าตั้งต้น = `ConfirmQty` ของใบอนุมัติ (**ไม่ใช่ `RemainQty`**) หรือ `BuyQty` ของใบราคารวม |
| 6 | ราคา/หน่วย (`PRICE_UNIT`) | 70 | **แก้ได้** AUTOMONEY2 ขวา | โหมด ก: `BCItemSupplier.LastBuyPrice` (ตรงสินค้า+`LastBuyUnitCode`=หน่วย) ไม่มีก็ `BuyPrice` (ตรงหน่วย); โหมด ข: `BuyPrice` ของใบราคารวม ถ้า ≤ 0 ใช้ราคาจาก `BCItemSupplier` เหมือนโหมด ก |
| 7 | ส่วนลด (`DISCOUNT`) | 60 | **แก้ได้** (ข้อความ `5%`, `100`) | โหมด ก: `BCAp.DetailDiscount` (ส่วนลดสินค้าของเจ้าหนี้); โหมด ข: `DiscountWord` ของแถวใบราคารวม ถ้าว่างใช้ `BCAp.DetailDiscount` |
| 8 | จำนวนเงิน (`AMOUNT`) | 60 | ล็อก ขวา | คำนวณ (ดู D) |

> shadow 49 ช่อง (`:141-191`): **หัวใบสั่งซื้อ** 1 DocNo · 2 DocDate · 3 TaxType · 4 ApCode · 5 ExpireCredit · 6 ExpireDate · 7 LeadTime · 8 LeadDate · 9 CurrencyCode · 10 ExchangRate · 11 BillGroup · 12 TaxRate · 13 MyDescription · 14 SumOfItemAmount · 15 BeforeTaxAmount · 16 DiscountWord(ส่วนลดท้ายบิล = `BCAp.Discount`) · 17 DiscountAmount · 18 AfterDiscount · 19 TaxAmount · 20 TotalAmount · 21 ExceptTaxAmount · 22 NetAmount · **รายการ** 23 ItemCode · 24 UnitCode · 25 DepartCode · 26 ItemType · 27 Description · 28 ProjectCode · 29 AllocateCode · 30 OppositeQty · 31 OppositeUnit · 32 HomeAmount · 33 ExceptTax · 34 BarCode · 35-39 หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ · 40 WHCode · 41 ShelfCode · **42 StkReqNo** (เลขที่ใบเสนอซื้อ) · 43 ItemNetAmount · **44 ReqQty** (จำนวนที่อนุมัติ/จำนวนในใบราคารวม — เพดานของคอลัมน์ 5) · **45 StkPODocNo** (เลขที่ใบอนุมัติ หรือเลขที่ใบราคารวม) · 46 CreditDay · 47 DueDate · 48 ContactCode. ข้อมูลหัวใบสั่งซื้อเก็บที่ shadow ของ "แถวหัวเจ้าหนี้" (และสำเนาบางช่องที่แถวสินค้าของเจ้าหนี้นั้น).

**C. การดึงรายการ**
- **โหมด ก** (`:1325-1517`, `:1519-1667`): `BCReqConfirmSub ⋈ BCItemSupplier ON ItemCode WHERE RemainQty>0 AND ApCode=<เจ้าหนี้ที่กรอก> AND DocNo=<ใบอนุมัติ> ORDER BY ItemCode, StkRequestNo` → ได้**เฉพาะสินค้าที่เจ้าหนี้รายนี้ขายตามทะเบียน `BCItemSupplier`** (สินค้าอื่นในใบอนุมัติไม่ถูกโหลด). ไม่พบรายการ (เลือกจากตัวเลือก) → ข้อความ `APPROVED_NO <เลขที่>` + `ITEM_NOT_SUPPLIER` (~"ใบอนุมัตินี้ไม่มีสินค้าที่เจ้าหนี้รายนี้ขาย"); (พิมพ์เลขเองแล้วไม่พบ → ไม่มีข้อความ กริดว่าง `:1519-1667`). แถวหัวเจ้าหนี้: ภาษีเริ่มต้นตาม `DefaultTaxType` ของบริษัท, ส่วนลดท้ายบิล = `BCAp.Discount`, แผนก = แผนกของแถวแรก. โหลดรายละเอียดเพิ่มต่อแถว: โครงการ/การจัดสรร/หน่วยขนาน จาก `BCStkRequestSub` (`TransState=0 AND IsCancel=0`) ตามเลขที่ใบเสนอซื้อ+สินค้า (`GetStkRequestSub` `:2132`), หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ/คลัง/ที่เก็บ จาก `BCItem` (`GetItemDetail` `:2113`). เสร็จแล้วคำนวณยอดทุกแถว (`DetailCalc`) และย้ายเคอร์เซอร์เข้ากริด (`Select(0,5)` `:1515`).
- **โหมด ข** (`:1168-1323`, `:1668-1817`): `BCByPriceCondSub WHERE DocNo=<ใบราคารวม> AND IsCancel=0 AND CreatePoFlag=0 ORDER BY ApCode, ItemCode, UnitCode, DepartCode` → **เจ้าหนี้หลายรายในใบเดียว = ได้แถวหัวหลายแถว (ใบสั่งซื้อหลายใบ)**; เฉพาะแถวที่ยังไม่เคยสร้างใบสั่งซื้อ (`CreatePoFlag=0`).
- ทั้งสองโหมด: **เลือกซ้ำ/เปลี่ยนใบต้นทาง = ล้างกริดทั้งหมดแล้วโหลดใหม่** (`DeleteContents`).

**D. การคำนวณ (`DetailCalc` `:1898-1968`, `BillCalc` `:1970-2007`, `CalcTaxAmount` `:1866-1896`)**: ยอดแถว = `จำนวน × ราคา` แล้วหักส่วนลดแถว — วิธีหักตามตั้งค่าบริษัท `DiscountMethod` (0 = หักจากยอดรวมแถวแล้วปัดตาม `AmountPoint`; อื่น ๆ = หักจากราคาต่อหน่วย (ปัด `PricePoint`) แล้วคูณจำนวน) → แสดงที่คอลัมน์ 8 · มูลค่าสุทธิของแถว (shadow 43): ถ้าภาษีรวมใน (`TaxType=1`) และสินค้าไม่ยกเว้น และอัตราภาษี > 1 → `ยอดแถว − ยอดแถว×อัตรา/(100+อัตรา)` · มูลค่าบาท (shadow 32) = สุทธิ × อัตราแลกเปลี่ยน (ถ้ามี) · แล้วรวมยอดของเจ้าหนี้รายนั้นเข้าแถวหัว: ยอดรวมสินค้าที่คิดภาษี (shadow 14) / ยอดยกเว้นภาษี (21: สินค้า `ExceptTax=1` และบริษัทเปิด `ExceptTax`) / ยอดอัตราศูนย์ (ตัวแปรรวม; `ExceptTax=2` และบริษัทเปิด `ZeroTaxRate`) → `BillTaxProcess` (`lib/BC5FRMWRK/BCStools.cpp:155-233`) คำนวณ ส่วนลดท้ายบิล/ก่อนภาษี/ภาษี/รวม (shadow 15,17-20) → ยอดสุทธิใบ (22) = รวม + ยกเว้น + อัตราศูนย์.
**กติกาตอนกรอก (`OnListChangeColumn` `:908-941`)**: จำนวนซื้อ < 0 → `PUT_NOT_QTY` คืนค่าเดิม (= จำนวนที่อนุมัติ) · **จำนวนซื้อ > จำนวนที่อนุมัติ/จำนวนในใบราคารวม (shadow 44) → `QTYOVER_SUPPLIER` คืนค่าเดิม** (แก้ลดได้ แก้เพิ่มเกินไม่ได้) · แก้จำนวน/ราคา/ส่วนลด → คำนวณแถวและยอดเจ้าหนี้รายนั้นใหม่ (`OnListChangeData` `:866-889`).

**E. กล่องกำหนดรายละเอียดใบสั่งซื้อรายเจ้าหนี้ (กดปุ่มลูกศรที่แถวหัวเจ้าหนี้ `OnExpand` `:1013-1109`; หัวข้อหน้าต่าง = `PO_FROM_AP : <รหัสเจ้าหนี้>(<ชื่อ>)`)** — **จำเป็นต้องเปิดและกด OK อย่างน้อยหนึ่งครั้งต่อเจ้าหนี้** เพื่อกำหนดเลขที่และวันที่ใบสั่งซื้อ (จอหลักไม่รันเลขให้เอง — ถ้าไม่กด จะติด `CANNOT_SAVE_NOTDOCNO` ตอนบันทึก):

| # | ป้าย | ชนิด | กติกา (หลักฐาน `PODlgPOProcess.cpp`) |
|--|--|--|--|
| 1 | เลขที่ใบสั่งซื้อ (`DOCNO_PO`) (13,71) | รันเลข `BCPurchaseOrder.DocNo` | ใบแรก: รันตามรูปแบบ `POConfig` (`BCSysTRFormat`/`BCSysTRDocG` `:426-460`); ใบถัดไปในรอบเดียวกัน: **รันต่อจากเลขที่ใบก่อนหน้า** (`XHighRunning(m_strRunDocNo)`) `:112-125` |
| 2 | วันที่ออกเอกสาร (13,221) | วันที่ไทย บังคับ | ค่าที่ส่งเข้ามา (ว่างได้) |
| 3 | ประเภทภาษี (25,71) | Combo 2 ค่า: แยกนอก / รวมใน (`CB_TAXTYPE_1/2`) | ค่าตั้งต้น = `DefaultTaxType` ของบริษัท; เปลี่ยน → คำนวณภาษีใหม่ |
| 4 | แผนก (25,221) | Combo-table `BCDepartment` | |
| 5 | โครงการ (49,71) / การจัดสรร (49,221) | Combo-table `BCProject` / `BCAllocate` | |
| 6 | เครดิต(วัน) (61,71) / รหัสผู้ติดต่อ (61,221) | ตัวเลข / Combo-table `BCContactList WHERE MyType=1` | ค่าตั้งต้นจากเจ้าหนี้ (`BCAP.CreditDay`, `DefContactCode`); แก้เครดิต → วันครบกำหนด (ช่องซ่อน/ล็อก) = วันที่เอกสาร + วัน |
| 7 | เอกสารหมดอายุภายใน(วัน) (73,71) / วันหมดอายุ (73,221) | ตัวเลข / วันที่ | แก้จำนวนวัน → วันหมดอายุ = วันที่เอกสาร + วัน |
| 8 | ระยะเวลาขนส่ง(วัน) (85,71) / ลงวันที่ (85,221) | ตัวเลข / วันที่ | ค่าตั้งต้นจาก `BCAP.LeadTime`; แก้วัน → วันที่รับสินค้า = วันที่เอกสาร + วัน |
| 9 | กลุ่มเอกสาร (97,71) / อัตราภาษีมูลค่าเพิ่ม (97,221) | ข้อความ / ตัวเลข | อัตราภาษีตั้งต้น: ถ้า `BCAP.PersonType=0` ใช้ `BCAP.DefaultTaxRate` ไม่งั้น `DefaultTaxRate` ของบริษัท; แก้อัตรา → คำนวณภาษีใหม่ |
| 10 | หมายเหตุ (109,71) | ข้อความหลายบรรทัด | |
| 11 | รหัสสกุลเงิน (129,71) / อัตราแลกเปลี่ยน (129,221) | Combo-table `BCCurrency` / ตัวเลข 9 ตำแหน่ง | **ปิดทั้งคู่ถ้าบริษัทไม่เปิด `MultiCurrency`**; เลือกสกุล → ดึง `BCCurrency.ExchangeRate` |
| 12 | ยอดรวมมูลค่าสินค้า (161,71) / ยอดรวมภาษี (`SUMTAXAMOUNT`; ข้อความไทยใน rc มีหลายแบบ) (161,221) | ตัวเลข | แก้ยอดรวมสินค้า → คำนวณภาษี/รวมใหม่ |
| 13 | ส่วนลด%,บาท (173,71: ปุ่ม `%` เปิด `CLIBDlgDiscount` + ช่องยอดส่วนลด) / ยอดเงินยกเว้นภาษี (173,221) | ปุ่ม+ตัวเลข / ตัวเลข | ตั้งส่วนลดท้ายบิลแล้วคำนวณใหม่; แก้ยอดยกเว้น → ยอดสุทธิ = รวม + ยกเว้น |
| 14 | ภาษีมูลค่าเพิ่ม (185,71) / ยอดรวมเงินสุทธิ (185,221) | ตัวเลข | **แก้ภาษีเองได้** → ย้อนคำนวณยอดรวม (`nProcessMode=3`) |
| 15 | รหัสเจ้าหนี้ + ชื่อ | ล็อก (37,71) | |
| ปุ่ม | **F11 = ตกลง** (`ID_OK`) / **Esc = ยกเลิก** | | ตกลง → ส่งค่ากลับเข้า shadow ของแถวหัวและ**ทุกแถวสินค้าของเจ้าหนี้นั้น** แล้วคำนวณใหม่ |

**กล่องรายละเอียดรายการ** (ปุ่มลูกศรที่แถวสินค้า `OnItemExpand` `:1111-1138`; ชื่อหน้าต่าง "Other Detail"): โครงการ · การจัดสรร · ที่เก็บซื้อ (`BUYSHELF`, `BCShelf`) · คลังซื้อ (`BUYWAREHOUSE`, `BCWarehouse`) · ยอดเงินสุทธิ (`NET_AMOUNT`) · ยอดเงินบาทสุทธิ (`TOTAL_BATHNET`) · คำอธิบายรายการ (`COMMENT`) + OK.

**F. คีย์ลัด** (`PreTranslateMessage` `:976-1011`): Ctrl+Enter ที่ช่องใบอนุมัติ/ใบราคารวม → กริดคอลัมน์จำนวนซื้อ แถว 0 · PageUp → ช่องรหัสเจ้าหนี้ (โหมด ก) หรือช่องใบราคารวม (โหมด ข) · PageDown → กริดคอลัมน์จำนวนซื้อ แถว 0 · F7 กลืนไว้ · ใหม่/ค้นหา (`ID_HK_NEW/FIND`). **ไม่มีปุ่มบันทึกเฉพาะ** — ใช้ปุ่ม/คีย์บันทึกมาตรฐานของเฟรม (F11 → `ConfirmSaveData` → `OnSaveData`). ปุ่มดูประวัติซื้อ `OnHistoryPurchase` (`:2148`) มีโค้ดแต่ไม่ผูกกับ message map/คีย์ใด ๆ — ไม่ต้องทำตาม.

**G. กติกาก่อนบันทึก (`ConfirmSaveData` `:259-323`)** — วนทุกแถว: แถวหัวเจ้าหนี้ต้องมี **เลขที่เอกสาร** (`CANNOT_SAVE_NOTDOCNO`) และ **วันที่เอกสาร** (`CANNOT_SAVE_NOTDOCDATE`) มิฉะนั้นห้ามบันทึกทั้งชุด · แถวสินค้า: **จำนวน = 0** (`CANNOT_SAVE_BUYQTY`), **ราคา = 0** (`CANNOT_SAVE_BUYPRICE`), **ยอดเงิน < 0** (`CANNOT_SAVE_BUYAMOUNT`) → ห้ามบันทึก (**ข้อความสามข้อนี้ถูกประกอบแต่ไม่เคยถูกแสดง** — ฟังก์ชันคืน FALSE เงียบ ๆ ผู้ใช้ไม่รู้สาเหตุ; BC ควรแสดงพร้อมเลขที่แถว/รหัสสินค้า) · ตรวจจำนวนธุรกรรมตามรุ่นโปรแกรม (`CheckTransactoin` — ข้ามได้).

**H. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:407-604`)** — ต่อแถวหัวเจ้าหนี้แต่ละแถว (วนเรียงจากบนลงล่าง):
1. ถ้าเลขที่ซ้ำกับ `BCPurchaseOrder` ที่มีอยู่ → รันเลขใหม่ให้อัตโนมัติ (`HighRunningDocNo` `:386`: ใช้ `MAX(DocNo)`+1 พร้อมแจ้ง `RUN_DOCNO_AUTO` / `PURCHASE_NO_SET`).
2. `InsertDB(BCPurchaseOrder)`: `DocNo, DocDate, TaxType, ApCode, DepartCode, ProjectCode, AllocateCode, ExpireCredit, ExpireDate, LeadTime, LeadDate, CreditDay, DueDate, ContactCode, CurrencyCode, ExchangeRate(≤0 → 1), BillGroup, TaxRate, MyDescription, SumOfItemAmount, BeforeTaxAmount, DiscountWord, DiscountAmount, AfterDiscount, TaxAmount, TotalAmount, ExceptTaxAmount, NetAmount, ZeroTaxAmount, IsCompleteSave=0, BillStatus=0, CreatorCode, CreateDateTime`.
3. ต่อแถวสินค้าของเจ้าหนี้นั้น `InsertDB(BCPurchaseOrderSub)`: `DocNo, DocDate, TaxType, ApCode, ItemCode, Qty, Price, DiscountWord, Amount, RemainQty(=Qty), DiscountAmount(= Qty×Price − Amount), CurrencyCode, ExchangeRate, UnitCode, DepartCode, ItemType, MyDescription, ProjectCode, AllocateCode, OppositeQty/Unit, HomeAmount, ExceptTax, BarCode, หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ, WHCode, ShelfCode, StkReqNo, NetAmount, ConfirmNo(= เลขที่ใบอนุมัติ หรือ เลขที่ใบราคารวม), LineNumber` → แล้วคิว `OnStockReprocess(รหัสสินค้า, วันที่)` (ประมวลผลสต็อกของสินค้านั้น — ดู §1.1 / `07-inventory.md` §0.1).
4. ถ้าแถวย่อยลงครบ → `UPDATE BCPurchaseOrder SET IsCompleteSave=1` ไม่ครบ → แจ้ง `SAVEDETAIL_NOTCOMPLETE` ("บันทึกรายการย่อยไม่สมบูรณ์") แต่ใบอื่นยังเดินต่อ; logbook `SAVE_PROCESS_PO`.
5. **หลังวนครบทุกใบ — ปิดสถานะต้นทาง**: โหมด ก → `UPDATE BCReqConfirm SET BillStatus=1 WHERE DocNo=<ใบอนุมัติ>` (ปิด**ทั้งใบ**ทันที) · โหมด ข → ต่อแถว `UPDATE BCByPriceCondSub SET CreatePoFlag=1 WHERE DocNo, ItemCode, ApCode` แล้วนับแถวที่ `CreatePoFlag=0` ของใบราคารวมนั้น — ถ้า 0 แถว `UPDATE BCBuyPriceCondit SET BillStatus=1` (`:586-603`).
- **ไม่ใช้ transaction ครอบ** (แต่ละคำสั่งเป็นอิสระ — ล้มกลางทางได้ใบสั่งซื้อที่ลงบางส่วน) · **ไม่เรียก `UpdateDocumentFlow`** · ไม่มี GL/`processdoc`.

**I. ข้อสังเกตจากโค้ดเดิม (ไม่ใช่ข้อกำหนด — ถ้าทำใหม่ควรแก้)**
1. โหมด ก: ใบอนุมัติถูกตั้ง `BillStatus=1` ทั้งใบทันทีที่บันทึก ทั้งที่โหลดมาเฉพาะสินค้าของเจ้าหนี้รายเดียวและแก้จำนวนลดได้ → สินค้า/จำนวนที่เหลือของใบอนุมัติ**ไม่สามารถประมวลต่อ**ได้อีก (ขัดกับ `UpdateDocumentFlow(…,15,…)` ที่คำนวณ `RemainQty` ตามจริงและตั้ง `BillStatus=1` เฉพาะเมื่อครบ; `mytools.cpp:2128-2176`). และ `BCPurchaseOrderSub` ที่สร้างจากจอนี้**ไม่บันทึก `RefLineNumber`** (จอใบสั่งซื้อ §1.1 ใช้ฟิลด์นี้จับคู่) ทำให้ `RemainQty` ของใบอนุมัติ (ถ้าถูกคำนวณใหม่ภายหลัง) อาจไม่นับยอดจากจอนี้ (ยังไม่ verify ค่าเริ่มต้นคอลัมน์).
2. โหมด ข: `GetStkRequestSub` ค้น `BCStkRequestSub` ด้วย "เลขที่ใบเสนอซื้อ" แต่ค่าที่ได้จากใบราคารวมคือ**เลขที่ใบอนุมัติ** (ดู §1.12 หมายเหตุ shadow 11) → โครงการ/การจัดสรร/หน่วยขนานในโหมดนี้โหลดไม่ได้ (ว่าง).
3. `ZeroTaxAmount` ใช้ตัวแปรรวมตัวเดียวของทั้งฟอร์ม (`m_dSumOfZeroTax`) บันทึกให้ทุกใบ — ถ้ามีหลายเจ้าหนี้ ยอดอัตราศูนย์ของใบหลังอาจปนกัน.
4. ในโหมด ก ตัวเลือกกรองแค่ `BillStatus = 0` ของใบอนุมัติ — ใบอนุมัติที่ผ่านการสืบราคา (§1.12) ถูกตั้ง `BillStatus=1` แล้วจึงไม่ปรากฏ ⇒ **ใบอนุมัติหนึ่งใบไปได้ทางเดียว: สืบราคา→ตารางราคา→ประมวล (โหมด ข) หรือประมวลตรง (โหมด ก)**.
5. ปุ่มวิทยุ/ป้ายภาษาไทยจริงมาจากตาราง `myresource` (ไม่อยู่ใน repo) — ที่ระบุในเอกสารนี้คือข้อความจาก `.rc`/คอมเมนต์ที่ resolve ได้เท่านั้น.

### 1.15 รายการ 5811 "บันทึกใบจ่ายเงินมัดจำ" (จ่ายเงินมัดจำให้เจ้าหนี้ล่วงหน้า — มีใบกำกับภาษี/ภาษีซื้อ/หัก ณ ที่จ่าย/จ่ายเงินในใบเดียว)

- **ฟอร์ม**: `CPOFrmReceiveEarnest` (`pofrmreceiveearnest.cpp` 2,144 บรรทัด — ชื่อคลาสว่า "Receive" แต่ทำหน้าที่ **จ่าย**; `MainFrm.cpp` `OnPosysTrans5`; dialog เดียว `IDD_PO_FRM_RECEIVE_EARNEST` 341x244 `BC5Account.rc:1397`; ไฟล์ตั้งค่ารายการเลือก `APDepositConfig`). ฝั่งลูกหนี้มีฟอร์มคู่ขนาน `BILLFrmReceiveEarnest` (`BCARDeposit`) — ดู `04-ar.md`.
- **ตาราง**: `BCAPDeposit` ตารางเดียว (ไม่มีตารางรายการ) — `DocNo, DocDate, TaxNo, TaxDate*, TaxType, ApCode, DepartCode, CreditDay, DueDate, TaxRate, MyDescription, ExcessAmount1/2, OtherIncome, OtherExpense, CurrencyCode, ExChangeRate, BeforeTaxAmount, TaxAmount, TotalAmount, NetAmount, BillBalance, SumOfWTax, SumCashAmount, SumChqAmount, PettyCashAmount, SumBankAmount, GLTransData, GLFormat, AllocateCode, ProjectCode, BillGroup, RecurName, IsCancel, IsConfirm, CreatorCode/CreateDateTime | LastEditorCode/LastEditDateT` (`OnSaveData` `:578-702`; `TaxDate` อ่านจากแถวภาษีซื้อ). ตารางลูกที่ฟอร์มเขียนผ่านฟังก์ชันกลาง: ภาษีซื้อ `BCInputTax` (`Source=1`), ภาษีหัก ณ ที่จ่าย `BCAPWTaxList` (`SaveFrom=0`), การจ่ายเงิน `BCPaymentSub`/ตารางเงินสด-เช็ค-ธนาคาร (ผ่าน `OnSavePayment`), สมุดรายวัน GL.
- **ยอดคงเหลือ (`BillBalance`) = วงจรชีวิตของมัดจำ**: เริ่ม = ยอดก่อนภาษี (แยกนอก, `TaxType=0`) หรือยอดรวมภาษี (รวมใน) → **ลดลงเมื่อถูกนำไปหักในใบซื้อ/ใบจ่ายชำระหนี้** (`BCAPDepositUse(DepositNo, Amount)` → `ProcessAPDeposit`: `BCAPDeposit.BillBalance += Amount × nMark` โดยเลือกแถวที่ `TaxNo IS NULL` ก่อน ไม่พบจึงลองแถวที่มี `TaxNo`; `bcprocess.cpp:3151-3164`) และ**ลดเมื่อทำใบลดหนี้เงินมัดจำ** (`BCInvStkRefund` `BillType=8` — §1.7/§1.8). **แก้ไข/ลบ/ยกเลิกได้เฉพาะเมื่อยังไม่ถูกใช้ = `NetAmount == BillBalance`** (เทียบทศนิยม `AmountPoint`).

**A. หัวเอกสาร (y,x ตาม rc; 2 กล่อง — กล่องบนข้อมูลทั่วไป, กล่องล่างยอดเงิน)**

| # | ป้ายบนจอ | ช่อง/ชนิด | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบจ่ายเงินมัดจำ: | รันเลข `DocNo` (15,83) | ใช่ | รันตามรูปแบบเอกสาร `BCAPDeposit`; ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่; ยาวเกิน `DOCNO_MORE_LENG`; แก้เลขตอนแก้ไขแล้วซ้ำ `DOCNO_DUPLICATE` | `:147-150,2019,377-390` |
| 2 | วันที่ออกเอกสาร: | วันที่ไทย (15,236) | ใช่ | = วันทำงาน; ล็อกตามสิทธิ์แก้วันที่; งวดปิด → ห้ามบันทึก/ลบ | `:151-152,205,364` |
| 3 | เลขที่ใบกำกับ: | รันเลขใบกำกับ `TaxNo` (27,83) | — | รันตามรูปแบบ `m_strTaxFormat`; ใส่แล้วต้องมีวันที่ใบกำกับ; **ถ้าเว้นว่างไม่คำนวณภาษี** (ยอดภาษี = ว่าง, รวม = ก่อนภาษี) | `:154-156,396-419,1753-1794` |
| 4 | วันที่ใบกำกับ: | วันที่ไทย (27,236) | เมื่อมีเลขที่ใบกำกับ | = วันทำงาน; ขาด → `CHECK_TAXDATE` ("มีเลขที่ใบกำกับภาษีแต่ไม่ใส่วันที่") | `:157,397,1801` |
| 5 | ประเภทภาษี: | Combo 2 ค่า: แยกนอก / รวมใน (39,83) | ใช่ | = `DefaultTaxType` ของบริษัท; เปลี่ยน → คำนวณใหม่ | `:166-167,1898` |
| 6 | แผนก: | Combo-table `BCDepartment` (39,236) | — | | `:187` |
| 7 | รหัสเจ้าหนี้: | Link ตรวจกับ `BCAP.Code` (51,83) + ชื่อ ล็อก (51,162) | ใช่ (`CANNOT_SAVE_NOTAPCODE`, ไม่พบ `NOT_FOUND_AP`) | โหมดเพิ่มใหม่เท่านั้น: ออกจากช่อง → ดึงชื่อ (ตาม `DefaultApName`), **อัตราภาษี = `BCAP.DefaultTaxRate` (ถ้า ≥ 0) ไม่งั้นอัตราของบริษัท**, **เครดิต(วัน) = `BCAP.CreditDay`** และวันครบกำหนด = วันที่เอกสาร + เครดิต | `:175-177,1218-1259,428-446` |
| 8 | การจัดสรร: / โครงการ: | Combo-table `BCAllocate` (63,83) / `BCProject` (63,236) | — | | `:188-189` |
| 9 | เครดิต(วัน): / วันที่ครบกำหนด: | ตัวเลข (75,83) / วันที่ (75,236) | — | แก้เครดิต → วันครบกำหนด = วันที่เอกสาร + วัน; แก้วันครบกำหนด → คำนวณเครดิตกลับ | `:158,174,1725,1736` |
| 10 | กลุ่มเอกสาร: | ข้อความ (87,83) | — | → `BillGroup` (กลุ่มของบิล/ภาษี) | `:756` |
| 11 | หมายเหตุ: | ข้อความ (99,83) | — | → `MyDescription` | `:632` |
| 12 | รหัสสกุลเงิน: / อัตราแลกเปลี่ยน: | Combo-table `BCCurrency` (138,83) / ตัวเลข 9 ตำแหน่ง (138,236) | — | **ปิดทั้งคู่ถ้าบริษัทไม่เปิด `MultiCurrency`**; เลือกสกุล → ดึงอัตราจาก `BCCurrency.ExchangeRate`; อัตรา ≤ 0 → 1 | `:179-184,1709-1724,1885` |
| 13 | อัตราภาษีมูลค่าเพิ่ม: | ตัวเลข + ปุ่ม (176,83) | — | แก้ → คำนวณใหม่ | `:169,1820` |
| 14 | จำนวนเงินก่อนภาษี: | ตัวเลข (188,83) | ใช่ (≠ 0) | ผู้ใช้กรอกยอดนี้เป็นหลัก | `:170,1261` |
| 15 | ภาษีมูลค่าเพิ่ม: | ตัวเลข (200,83) | — | คำนวณให้ แต่**แก้ภาษีเองได้** (ย้อนคำนวณยอดรวม) | `:171,1303` |
| 16 | รวมเงินทั้งสิ้น: | ตัวเลข (176,236) | ใช่ (≠ 0) | แก้ยอดนี้ → ย้อนคำนวณก่อนภาษี/ภาษี (กรณีรวมใน) | `:172,1348` |
| 17 | ยอดเงินสุทธิ: | ตัวเลข ล็อก (188,236) | — | = รวมเงินทั้งสิ้น − ภาษีหัก ณ ที่จ่าย (`SumOfWTax`) — **ยอดนี้คือยอดที่ต้องจ่ายจริงและต้องเท่ากับยอดจ่ายเงิน (F7)** | `:622,1286-1294` |
| 18 | ยอดเงินคงเหลือ: | ตัวเลข ล็อก (200,236) | — | ตั้งค่าเมื่อสร้าง: แยกนอก → ก่อนภาษี; รวมใน → รวมเงินทั้งสิ้น (ปรับตามเมื่อแก้ยอดตราบที่ยังไม่ถูกใช้); หลังถูกใช้ลดลงตามยอดที่หัก; เป็นฟิลด์ล็อกแสดงอย่างเดียว | `:774-776,1287,1332` |

**กติกาคำนวณภาษี** (`OnKillfocus*` `:1261-1422`): **ไม่มีเลขที่ใบกำกับ → ไม่คิดภาษี** (ภาษี = ว่าง, รวมเงิน = สุทธิ = ก่อนภาษี); มีเลขที่ใบกำกับ → แยกนอก: ภาษี = ก่อนภาษี × อัตรา/100 ปัดตาม `AmountPoint`, รวม = ก่อนภาษี + ภาษี; รวมใน: ภาษี = รวม × อัตรา/(100 + อัตรา), ก่อนภาษี = รวม − ภาษี (ใช้ `BillTaxProcess` `lib/BC5FRMWRK/BCStools.cpp:155-233` เมื่อแก้ภาษีเอง `nProcessMode=3`). ยอดต่ำกว่า 1 บาทไม่คำนวณ (เงื่อนไข `> 1`). หลังคำนวณทุกครั้งอัปเดตแถวแรกของตารางภาษีซื้อ (`MyUpdateFirstTaxRow` `:2034`) ให้สอดคล้อง.

**B. ไม่มีกริดรายการ** — จอนี้มีแต่ยอดเงินและกล่องซ้อน 4 กล่อง (เรียกด้วยคีย์ลัด):

**C. คีย์ลัด** (`PreTranslateMessage` `:1616-1708`; เมนู `OnMore` `:1423`):
| คีย์ | ทำอะไร |
|--|--|
| F2 | ดู/แก้ **บัญชีแยกประเภท** ของเอกสาร (`OnAccount`; ปิดถ้าสินค้ารุ่นไม่มี GL) |
| F3 | **บันทึกภาษีซื้อ** (`OnInputTax("APDepositConfig", ประเภทภาษี, เจ้าหนี้)` — กล่องรายการใบกำกับภาษีซื้อได้หลายแถว; กลับมาแล้วดึงแถวแรกไปแสดงที่ช่องเลขที่/วันที่ใบกำกับ) |
| F5 | **Ctrl+F5 ยกเลิก** (ผ่านกติกา `OnConfirmCancel` แล้ว `BillCancel("BCAPDeposit")`), **Shift+F5 ยืนยัน** (สิทธิ์ `bCanConfirmBill`; `BillConfirm`) |
| F6 | **ภาษีหัก ณ ที่จ่าย** (`OnWAPTax(ยอดก่อนภาษี)` — ฐานภาษีหัก = ยอดก่อนภาษี; กลับมาแล้วคำนวณยอดเงินสุทธิ = รวมเงิน − ภาษีหัก และตั้งยอดคงเหลือใหม่) |
| F7 | **บันทึกจ่ายเงิน** (`OnPayment(ยอดเงินสุทธิ, …)` — กล่องจ่ายเงินสด/เช็ค/เงินสดย่อย/โอนธนาคาร; กลไกเดียวกับ `04-ar.md:987`) |
| F11 | บันทึก |
| F12 | **ตรวจเอกสารที่อ้างอิง** (`OnDocumentFlow` `:2063`): รวม 3 แหล่ง — ใบซื้อสินค้าที่หักมัดจำนี้ (`BCAPDepositUse WHERE DepositNo`) · ใบจ่ายชำระหนี้ที่หักมัดจำ (`BCPaymentSub WHERE InvoiceNo AND BillType=5`) · ใบลดหนี้เงินมัดจำ (`BCInvStkRefund WHERE InvoiceNo`) — คอลัมน์ วันที่/เลขที่/ประเภท/จำนวนเงิน |
| อื่น ๆ | พิมพ์ (`ID_HK_PRINT`), ใหม่, ค้นหา, Recurring (`OnRecur`), `OnInfo` (`:1533` เมนูสรุปยอด: เงินสดจ่าย/เช็คจ่าย/เงินโอน/เงินสดย่อย/เงินเกิน/ค่าใช้จ่ายอื่น/รายได้อื่น/ภาษีหัก ณ ที่จ่าย/ก่อนภาษี/ภาษี/รวมภาษี/สุทธิ/คงเหลือ) |

**D. กติกาก่อนบันทึก (`ConfirmSaveData` `:360-577`)** ตามลำดับ: งวดปิด · เลขที่ว่าง · (แก้เลข) เลขซ้ำ · มีเลขที่ใบกำกับแต่ไม่มีวันที่ใบกำกับ · **เลขที่ใบกำกับซ้ำกับของเจ้าหนี้รายเดียวกัน** (`BCInputTax` `Source=1`, ไม่นับเอกสารตัวเองตอนแก้ไข) `DUPLICATE_TAXNO` · ไม่มีรหัสเจ้าหนี้ / ไม่พบเจ้าหนี้ · ก่อนภาษี = 0 หรือรวม = 0 `CANNOT_SAVE_NOT_SUMDETAIL`; ติดลบ `CANNOT_SAVE_WORNG` · (**แก้ไข**) ลงบัญชีแล้ว `CANNOT_SAVE_POSTGL` · ยกเลิกแล้ว `NOT_SAVE_ISCANCEL` · ยืนยันแล้ว `NOT_SAVE_ISCONFIRM` · **ถูกนำไปใช้แล้ว (`NetAmount ≠ BillBalance`) `NOT_EDIT_BILLBALANCE`** · **ยอดจ่ายเงิน (F7) ไม่เท่ากับยอดเงินสุทธิ `AMOUNT_NOT_EQUAL_PAYMENT`** → เปิดกล่องจ่ายเงินให้แก้ทันที (`IsMoneyOk(1, net)` — `myespeciallyformview.cpp:4266-4318`) · กำหนดรูปแบบ GL อัตโนมัติถ้ายังว่าง (`GetGLFormat` → FormatGroup 0 `DEPSIT_AP`, ดู §0.4).
**ลบ** (`ConfirmDeleteData` `:251-294`): เลขที่ว่าง · งวดปิด · **ลงบัญชีแล้ว `CANNOT_DELETE_POSGL`** · **เช็คที่จ่ายถูกอ้างอิงหลายใบ `CHQNOT_EDIT_DELETE_CANCEL`** (`IsChqRefence`) · ยกเลิกแล้ว/ยืนยันแล้ว · **ถูกนำไปใช้แล้ว `NOT_DELETE_BILLBALANCE`** · ถามยืนยัน → `OnDeleteData` `:326-358` (ไม่ใช้ transaction ครอบ): `DELETE BCAPDeposit` → คืนยอดเช็คจ่าย `BCAP.ChqOnHand -= SumChqAmount` → `DeleteInputTax(0,เลขที่,1)` (ลบภาษีซื้อ) → `DeleteAPWTax(0,เลขที่)` → `DeletePayment(0,เลขที่)` → `DeleteAccount(เลขที่, source 1)` → logbook `DELETE_PAYDEPOSIT`.
**ยกเลิก** (`OnConfirmCancel` `:2093`, `BillCancelUpdateData` `:296-324`): ลงบัญชีแล้ว/ยกเลิกแล้ว/ยืนยันแล้ว/ถูกใช้แล้ว → ปฏิเสธ (`CANNOT_CANCEL`, `MSG_CANNOT_CANCEL`); ผ่าน → `UPDATE BCAPDeposit SET BillBalance=0, IsCancel=1` → `CancelInputTax(0,เลขที่,1)` (ทำเครื่องหมายยกเลิกภาษีซื้อ — ไม่ลบ) → `DeleteAPWTax` → `DeletePayment` (โหมดยกเลิกเงิน `m_bCancelMoney`) → **`ReverseAccount(เลขที่, source 1)` (ลงรายการกลับบัญชี ไม่ลบ)** → logbook `CANCEL_PAYDEPOSIT`. (ฟังก์ชันนี้ไม่ได้คืน `BCAP.ChqOnHand` เอง — ยังไม่ verify ว่า `DeletePayment` ทำให้.)

**E. ผลข้างเคียงตอนบันทึก (`OnSaveData` `:578-702`)**: (แก้ไข: ถอนยอดเช็คเดิมออกจาก `BCAP.ChqOnHand` ก่อน) → `StartTransaction` → `InsertDB`/`UpdateDB(BCAPDeposit)` (ฟิลด์ตามบน; `ExChangeRate` ≤ 0 → 1; `GLTransData = m_nManualKey`) → ล้มเหลว `Abort` + "Can not save" → สำเร็จ `Commit` แล้วตามด้วยฟังก์ชันกลางตามลำดับ: **`OnDepositUpdateChqAmount(1, เจ้าหนี้, ยอดเช็คจ่าย, +1)`** (`BCAP.ChqOnHand += ยอดเช็คจ่าย` — `myespeciallyformview.cpp:6545`; ทำเฉพาะเมื่อ > 0) → **`OnSaveInputTax(0,…, nProcess=1, GLSource=1, กลุ่มภาษี, หลายสกุลเงิน)`** (เขียน `BCInputTax`) → **`OnSaveAPWTax(0,…)`** (เขียน `BCAPWTaxList` `SaveFrom=0`) → **`OnSavePayment(0,…, สกุลเงิน, อัตรา, แผนก, โครงการ)`** (เขียนรายการจ่ายเงิน) → **ลงบัญชี**: `m_nManualKey==0` → `GLTLSProcess{ m_nSource=1 }.ProcessConditionGroup0()` ("จ่ายเงินมัดจำ"; ใช้รูปแบบ GL `GLFormat` ดู §0.4), มิฉะนั้นบันทึกรายการบัญชีที่ผู้ใช้แก้เอง `OnSaveAccount(…, source 1, …)` → logbook `SAVE_PAYDEPOSIT`. **ไม่เกี่ยวกับสต็อก; ไม่แตะยอดหนี้ `BCAP.DebtAmount`** (มัดจำเป็นสินทรัพย์ของกิจการ ไม่ใช่หนี้ — ยอดคงเหลือมัดจำอยู่ที่ `BCAPDeposit.BillBalance`).
**เปิดเอกสาร** (`OnLoadData` `:703-840`): โหลดยอด/สถานะ (`IsConfirm`,`IsCancel`) แล้ว `LoadInputTax(0)`, `LoadAPWTax(0)`, `LoadPayment(0)`, `LoadAccount(เลขที่, source 1)`; ถ้า `IsCancel`/`IsConfirm` แสดงตรา "ยกเลิก/ยืนยัน" (`DrawBillStatus`); เลขที่ใบกำกับ/วันที่ใบกำกับ = แถวแรกของภาษีซื้อ.

**F. ข้อสังเกตจากโค้ดเดิม (ไม่ใช่ข้อกำหนด)**
1. ข้อความคีย์ลัด F3 ในโค้ดเขียนว่า "ภาษีขาย" (`:1634`) แต่ฟังก์ชันจริงคือ**ภาษีซื้อ** (`OnInputTax`, `BCInputTax Source=1`) — ใช้ "ภาษีซื้อ".
2. ยอดคงเหลือที่โหลดจาก Recurring (เอกสารซ้ำ) ถูกตั้งใหม่ตามประเภทภาษี (ไม่ใช้ค่าเก่า) `:766-773`.
3. หัวข้อ/การจัดวางละเอียดของกล่อง F3/F6/F7 อยู่ในฟังก์ชันกลาง (ดู `00-common-ui.md` และ `04-ar.md` ส่วน C) — ไม่ซ้ำที่นี่.

### 1.16 รายการ 5806 "บันทึกใบจ่ายเงินล่วงหน้า" (จ่ายเงินล่วงหน้าให้เจ้าหนี้แบบ "เงินล่วงหน้าทั่วไป" — ไม่มีใบกำกับภาษี/ภาษีซื้อ ใช้หักตอนจ่ายชำระหนี้ภายหลัง)

- **ฟอร์ม**: `CPOFrmEarnestSpecail` (`POFrmEarnestSpecail.cpp` 1,468 บรรทัด; `MainFrm.cpp` `OnPosysTrans21` ← `ID_POSYS_TRANS_21`; dialog เดียว `IDD_PO_FRM_EARNEST_SPECIAL` 351x241 `BC5Account.rc:14412`). ชื่อหน้าต่าง/บันทึก log = `20021` "บันทึกใบจ่ายเงินล่วงหน้า" (`:32,1377`); log ใช้ id ของจอ 5811 (`IDD_PO_FRM_RECEIVE_EARNEST`) ร่วมกัน (`:296,339,645`) — **ต่างจาก 5811 ตรงที่ไม่มีภาษี**: 5811 = "มัดจำ" ที่มีใบกำกับภาษี/ภาษีซื้อ/หัก ณ ที่จ่าย (`BCAPDeposit`) ส่วน 5806 = เงินล่วงหน้าไม่มีใบกำกับ (`BCAPDepositSpecial`) คู่กับ 5807 (รับคืน) และ AR ฝั่งขาย (`BCARDepositSpecial`, ดู `04-ar.md`).
- **ตาราง**: `BCAPDepositSpecial` ตารางเดียว ไม่มีตารางรายการ — `DocNo, DocDate, ApCode, DepartCode, SaleCode, CreditDay, DueDate, MyDescription, ExcessAmount1/2, OtherIncome, OtherExpense, GLTransData, GLFormat, CurrencyCode, ExChangeRate, TotalAmount, NetAmount, BillBalance, SumOfWTax, SumCashAmount, SumChqAmount, PettyCashAmount, SumBankAmount, AllocateCode, ProjectCode, BillGroup, RecurName, IsConfirm, IsCancel, CreatorCode, CreateDateTime, LastEditorCode, LastEditDateT` (`OnSaveData :515-649`, `OnLoadData :650-745`). ตารางพ่วง: การจ่ายเงินจริง `BCPayMoney` (`SaveFrom=8`, `myespeciallyformview.cpp:2008-2016`), สมุดรายวัน GL (source 42), ล็อก `BCAPWTaxList` (ลบอย่างเดียว).
- **หลักการสำคัญ**: เอกสารนี้ **"จ่ายเงินสด/เช็ค/โอน ออกไปจริง ณ ขณะบันทึก"** — ผู้ใช้กรอกหัว + ยอด แล้วต้องระบุวิธีจ่ายในหน้าจอจ่ายเงิน (F7) ให้รวมเท่ากับยอดสุทธิ ระบบจึงยอมบันทึก; ตอนบันทึก **ยอดสุทธิ = ยอดรวม = ยอดคงเหลือ** (`:554-556`) แล้วยอดคงเหลือลดลงเมื่อถูกนำไปหักตอน "จ่ายชำระหนี้" (§ ใน `03-ap.md`) หรือถูกรับคืน (5807).

**A. หัวเอกสาร (x,y ตาม rc = `IDC_*`; จัดเป็นกล่องเดียว 2 คอลัมน์ — ซ้าย x=78, ขวา x=236; ไม่มีกริด/ท้ายใบ)**

| # | ป้ายบนจอ (key) | ช่อง/ชนิด (y,x) | จำเป็น | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่ใบจ่ายเงินล่วงหน้า: (`DOCNO_PAY`) | รันเลข `DocNo` (22,78) | ใช่ | รันตามรูปแบบเอกสาร `BCAPDepositSpecial` + กลุ่มเอกสาร; **ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่** (ระบบรันให้ แล้วโฟกัสไปรหัสเจ้าหนี้); **ยาวเกิน 20 ตัว** → `DOCNO_MORE_LENG` ตัดเหลือ 20 | `:132-135,1315-1351` |
| 2 | วันที่ออกเอกสาร: (`DOCNO_DATE`) | วันที่ไทย (22,236) | ใช่ | = วันทำงาน; **ล็อกถ้าไม่มีสิทธิ์แก้วันที่** (ยกเว้นผู้ใช้ `sa`); งวดปิด → เตือนทันทีที่เปิดจอและตอนบันทึก/ลบ/ยกเลิก | `:136-137,188-192` |
| 3 | รหัสเจ้าหนี้: (`APCODE`) | Link ตรวจกับ `BCAP.Code` (34,78) + ชื่อ ล็อก (34,157) | ใช่ | **โหมดเพิ่มใหม่เท่านั้น**: ออกจากช่อง → ดึงชื่อ (`Name1` หรือ `Name2` ตาม `DefaultApName`); โหมดแก้ไขไม่ดึงชื่อใหม่ | `:140-142,1037-1058` |
| 4 | รหัสพนักงาน: (`Sale_Code`) | Combo-table `BCSale` (46,78) | — | → `SaleCode` (ป้ายใช้คำว่ารหัสพนักงาน) | `:160,534` |
| 5 | แผนก: (`DEPART`) | Combo-table `BCDepartment` (46,236) | — | โหลดเอกสารที่สร้างโดยผู้ใช้ `OOMDTS` (SAP) → ล็อกช่อง — ข้าม | `:156,665-670` |
| 6 | โครงการ: / การจัดสรร: | Combo-table `BCProject` (58,78) / `BCAllocate` (58,236) | — | → `ProjectCode` / `AllocateCode` | `:158-159` |
| 7 | จ่ายส่วนที่เหลือภายใน: (`CREDITDAY_PAY`) … วัน | ตัวเลขจำนวนเต็ม (70,78) | — | แก้เครดิตแล้วออกจากช่อง → **วันที่ครบกำหนด = วันที่เอกสาร + วัน** | `:144,1253-1262` |
| 8 | **วันที่ส่งของ:** (`DELIVERY_DATE`) → ข้อมูลคือ `DueDate` | วันที่ไทย (70,236) | — | แก้แล้วออกจากช่อง → คำนวณเครดิตกลับ = `(วันครบกำหนด − วันที่เอกสาร) − 1` วัน. **ป้ายเป็น "วันที่ส่งของ" แต่เก็บเป็นวันครบกำหนด (`DueDate`)** — ตามต้นฉบับ (ป้ายเดียวกับ ใบเสนอซื้อ ฯลฯ ที่คัดลอกมา) | `:145,1264-1279` |
| 9 | หมายเหตุ: (`REMARK`) | ข้อความหลายบรรทัด 226x43 (82,78) | — | → `MyDescription`; **ส่งต่อไปเป็นคำอธิบายรายการในหน้าจอจ่ายเงิน (F7) และสมุดรายวัน** | `:535,1092,1237` |
| 10 | รหัสสกุลเงิน: / อัตราแลกเปลี่ยน: | Combo-table `BCCurrency` (125,78) / ตัวเลข 9 ตำแหน่ง (125,236) | — | **ปิดทั้งคู่ถ้าบริษัทไม่เปิด `MultiCurrency`**; เลือกสกุลเงิน (เปลี่ยนค่า/ออกจากช่อง) → ดึง `BCCurrency.ExchangeRate`; บันทึก: อัตรา ≤ 0 → 1.0 | `:147-156,1281-1308,548-552` |
| 11 | กลุ่มเอกสาร: (`DOCGROUP`) | ข้อความ (137,78) | — | → `BillGroup` | `:566` |
| 12 | **ยอดจ่ายชำระ** (`PAYMENT_AMOUNT`) | ตัวเลข (137,236) ตามทศนิยมเงินของบริษัท | ใช่ (> 0) | ช่องเดียวที่ผู้ใช้กรอกจำนวนเงิน → บันทึกเป็น `TotalAmount` = `NetAmount` = `BillBalance` (ค่าเดียวกันทั้งสามช่อง) | `:150,554-556` |
| 13 | ยอดเงินคงเหลือ: (`DEPOSIT_BALANCE`) | ตัวเลข ล็อก (149,236) | — | แสดงยอดที่ยังไม่ถูกนำไปหัก/รับคืน (= `BillBalance`) | `:83,692-703` |

> ขอบเขตของฟอร์มเล็ก: **ไม่มีเลขที่ใบกำกับ/วันที่ใบกำกับ/ประเภทภาษี/อัตราภาษี/ยอดก่อนภาษี/ภาษี** (ต่างจาก 5811 ที่มีทั้งหมด) · **ไม่มีผู้บันทึกบนจอ** (เก็บ `CreatorCode` ตอนบันทึกครั้งแรกเท่านั้น `:579-580`).
> `IsConfirm` ของเอกสารนี้ **ไม่ได้ตั้งตอนบันทึก** — ถูกตั้งโดย (ก) ปุ่ม Confirm เอกสาร (Shift+F5) และ (ข) ฟอร์มจ่ายชำระหนี้เมื่อมีการนำเงินล่วงหน้าไปหัก (`apfrmappayment1view.cpp:3633-3642`: หักแล้ว `BillBalance ≠ NetAmount` → `IsConfirm=1`; คืนครบ → `IsConfirm=0`).

**B. หน้าจอจ่ายเงิน (F7 / ปุ่มบันทึกจ่ายเงิน)** — ใช้กล่องโต้ตอบ "รายการจ่ายเงิน" กลางเดียวกับเอกสารฝั่งเจ้าหนี้ (`OnPayment(ยอดสุทธิ, FALSE, หมายเหตุ, เลขที่เอกสาร, รหัสเจ้าหนี้)` `myespeciallyformview.cpp:4256`; รายละเอียดช่องเงินสด/เช็คจ่าย/เงินโอน/เงินสดย่อย/หักภาษี ณ ที่จ่าย อยู่ใน `04-ar.md:987` และ `06-cash-bank.md`) — **ปุ่มนี้คือขั้นตอนที่ต้องทำก่อนบันทึกทุกครั้ง** เพราะ `ConfirmSaveData` บังคับให้ยอดรวมของวิธีจ่ายเท่ากับยอดสุทธิ (`IsMoneyOk(1,net)` ตามกติกา `myespeciallyformview.cpp:4266-4318`); ถ้าไม่ตรงระบบเปิดหน้าจอนี้ให้เองหลังข้อความเตือน (`:436-442`).

**C. คีย์ลัด** (`PreTranslateMessage :1191-1250`; เมนู `OnMore :1061-1122`)

| คีย์ | การทำงาน | หมายเหตุ |
|--|--|--|
| **F2** | บัญชีแยกประเภท (`OnAccount` — เปิดหน้าจอปรับสมุดรายวันของเอกสาร) | ถูกบล็อกถ้าผลิตภัณฑ์ตัดโมดูล GL (`LMD_REMOVE_OPPOSIT`); มีผลเมื่อ `GLTransData` = มือ |
| **F5** + Ctrl | **ยกเลิกเอกสาร** (`OnConfirmCancel` → `BillCancel`) | ถ้า `IsConfirm=1` → `CANNOT_CANCEL` (ห้าม) |
| **F5** + Shift | **Confirm เอกสาร** (`BillConfirm`) | ต้องมีสิทธิ์ `bCanConfirmBill` (ไม่งั้น `NOT_CONFIRM_NO`) |
| **F7** | บันทึกจ่ายเงิน (ยอดสุทธิ) | เฉพาะเมื่อมีเลขที่เอกสาร |
| **F11** | บันทึก (`OnRecSave`) | จากเมนูพ่วง `OnMore` |
| **F12** | ตรวจสอบเอกสาร (`OnDocumentFlow`) | ดู D |
| ปุ่มพิมพ์/ใหม่/ค้นหา/ข้อมูล/Recurring | `ID_HK_PRINT/NEW/FIND`, `ID_REC_INFO` (โชว์ "ยอดคงเหลือ ^…" `:1178`), `IDC_RECUR` (`BaseRecur`) | Recurring โหลดมาแล้ว **ยอดคงเหลือ = ยอดสุทธิ** และไม่คัดลอกยอดจ่ายจริง `:690-715` |

เมนูพ่วง `OnMore` (6 รายการตามลำดับ): F2=บัญชีแยกประเภท · F7=บันทึกจ่ายเงิน · F11=บันทึก · F12=ตรวจสอบเอกสาร · Shift+F5=Confirm · Ctrl+F5=ยกเลิก (ซ่อนข้อแรกถ้าตัด GL, ดัชนีเลื่อน `:1081-1083`).

**D. F12 ตรวจสอบเอกสาร** (`OnDocumentFlow :1363-1383`): หัวข้อ "ใบจ่ายเงินมัดจำ + เลขที่" แสดงเอกสารที่อ้างถึงใบนี้ — (1) **บันทึกรับคืนเงินจ่ายล่วงหน้า** = `BCAPReturnDepSpecial WHERE EarnestNo = เลขที่นี้ AND IsCancel=0` (ยอด = `NetAmount`) และ (2) **ใบจ่ายชำระหนี้** = `BCPaymentSub WHERE InvoiceNo = เลขที่นี้ AND TransState=0 AND IsCancel=0 AND BillType=8` (ยอด = `PayAmount`); คอลัมน์ วันที่ / เลขที่เอกสาร / ประเภทรายการ / จำนวนเงิน เรียง `DocDate, DocNo`.

**E. กติกาก่อนบันทึก (`ConfirmSaveData :345-514` — ลำดับจริง)**: (1) งวดปิด (`IsPeriodClose`) · (2) โหมดแก้ไข + เอกสารผ่าน GL แล้ว → `CANNOT_SAVE_POSTGL` · (3) เลขที่ว่าง → `CANNOT_SAVE_NOTDOCNO` · (4) เลขที่ซ้ำ (เพิ่มใหม่ หรือแก้เลขที่) → `DOCNO_DUPLICATE` (ในโหมดเพิ่มใหม่มีรันเลขใหม่/ลองซ้ำ) · (5) รหัสเจ้าหนี้ว่าง → `CANNOT_SAVE_NOTAPCODE`; ไม่พบใน `BCAP` → `NOT_FOUND_AP` · (6) **โหมดแก้ไข**: ถูกยกเลิกแล้ว `NOT_SAVE_ISCANCEL` · ถูก Confirm แล้ว `NOT_SAVE_ISCONFIRM` · **ยอดคงเหลือ ≠ ยอดสุทธิ (เคยถูกนำไปหัก/คืนบางส่วน) → `NOT_EDIT_BILLBALANCE` — แก้ไม่ได้** · (7) ยอดสุทธิ ≤ 0 → `CANNOT_SAVE_NO_AMOUNT` · (8) วิธีจ่ายเงินรวม ≠ ยอดสุทธิ → `AMOUNT_NOT_EQUAL_PAYMENT` แล้วเปิดหน้าจอจ่ายเงิน (ตัวเรียกกลับมาบันทึกอีกครั้งเอง) · (9) `GLFormat` ว่าง → ใช้ค่าเริ่มต้นของ FormatGroup 42 (`ADVANCEPAYMENT_AP`, `GetGLFormat :1357-1361`).
**ลบ (`ConfirmDeleteData :216-260`)**: เลขที่ว่าง `CANNOT_DELETE` · งวดปิด · ผ่าน GL แล้ว `CANNOT_DELETE_POSGL` · **เช็คที่จ่ายในใบนี้ถูกนำไปตัดจ่ายหลายใบ** (`IsChqRefence(เลขที่,8,1)`) → `CHQNOT_EDIT_DELETE_CANCEL` · ยกเลิกไปแล้ว `NOT_DELETE_ISCANCEL` · Confirm แล้ว `NOT_DELETE_ISCONFIRM` · ยอดคงเหลือ ≠ ยอดสุทธิ `NOT_DELETE_BILLBALANCE` · ถามยืนยัน `CONFIRM_DELETEDATA` (Yes/No).
**ยกเลิก (`OnConfirmCancel :1394-1440`)**: ต้องเป็นเอกสารที่บันทึกแล้ว · งวดปิด · **ผ่าน GL แล้ว → `CANNOT_CANCEL`** · Confirm แล้ว `MSG_CANNOT_CANCEL` · **มีใบรับคืน (5807, `IsCancel=0`) หรือใบจ่ายชำระหนี้ (`BCPaymentSub BillType=8`) อ้างถึง → `CANNOT_CANCEL_USE`** ("ไม่สามารถลบรายการนี้ได้ เนื่องจากมีการนำข้อมูลไปใช้แล้ว"). (หมายเหตุ: ข้อตรวจ `IsConfirm` ซ้ำสองชั้นมีข้อความต่างกัน `:1412-1419` — ชั้นที่สองไม่มีทางถึง.)

**F. ผลข้างเคียงตอนบันทึก (`OnSaveData :515-649`)**
1. **ถ้าแก้ไข**: ก่อนทำอะไร คืนยอดเช็คจ่ายเดิมของเจ้าหนี้เดิม `OnDepositUpdateChqAmount(1, รหัสเจ้าหนี้เดิม, SumChqAmount เดิม, -1)` (`BCAP.ChqOnHand -= เช็ค` ตาม `myespeciallyformview.cpp:6545`).
2. `StartTransaction` → `InsertDB` (เพิ่มใหม่: `CreatorCode`, `CreateDateTime`) หรือ `UpdateDB WHERE DocNo=เลขเดิม` (แก้ไข: `LastEditorCode`, `LastEditDateT`; ก่อนหน้านั้นรัน `apRecalculate(เจ้าหนี้เดิม, "##CALC_EARNEST_BALANCE##", -100)`) → ล้มเหลว = `Abort` + ข้อความ "Can not save" (ภาษาอังกฤษตายตัว `:610`) → สำเร็จ = `Commit`.
3. หลัง commit: `OnDepositUpdateChqAmount(1, รหัสเจ้าหนี้ใหม่, ยอดเช็คจ่ายใหม่, +1)` → **`OnSavePayment(8, เลขเดิม, เลขใหม่, วันที่, รหัสเจ้าหนี้, สกุลเงิน, อัตรา, แผนก, โครงการ)`** (ลบรายการจ่ายเดิมแล้วเขียนใหม่ลง `BCPayMoney` SaveFrom=8; มี cheque/เงินโอน/เงินสดย่อยตามที่กรอกในหน้าจอ F7) → **`apRecalculate(เจ้าหนี้, "##CALC_EARNEST_BALANCE##", -100)`**.
4. **ลงบัญชี**: ถ้า `GLTransData` (`m_nManualKey`) = 0 (อัตโนมัติ) → `m_GLTLSProcess{ m_nSource=42; m_nGLGroup=45 }.ProcessConditionGroup42()` (engine group 45 → `ProcessConditionGroup42` ตามตาราง `myespeciallyformview.cpp:6423-6459`; FormatGroup ของรูปแบบ = 42 `ADVANCEPAYMENT_AP`); ถ้ามือ → `OnSaveAccount(เลขเดิม, เลขใหม่, วันที่, 42, รหัสเจ้าหนี้, หมายเหตุ)` (เขียนรายการบัญชีที่ผู้ใช้แก้ใน F2). **ไม่มีภาษีซื้อ / ไม่มีหัก ณ ที่จ่ายถูกบันทึกจากจอนี้** (ไม่เรียก `OnSaveInputTax`/`OnSaveAPWTax`; ดู ข้อสังเกต 1) · ไม่มีสต็อก · ไม่มี `processdoc`.
5. **ยอดเงินล่วงหน้าของเจ้าหนี้ `BCAP.EarnestAmount`** คำนวณโดย `calcEarnetAmount` (`BCAPProcess.cpp:725-790`) = **Σ `BCAPDepositSpecial.NetAmount×อัตรา`** (ยังไม่ยกเลิก) **− ( Σ `BCAPReturnDepSpecial.NetAmount×อัตรา` + Σ `BCPaymentSub.PayAmount×InvExchangeRate` ที่ `BillType=8` ยังไม่ยกเลิก )** ปัดเศษตามทศนิยมเงินของบริษัท.
6. log "บันทึกใบจ่ายเงินล่วงหน้า" (`SAVE_PAYDEPOSIT`) ชนิดเพิ่ม/แก้.

**G. ลบ/ยกเลิก — ลำดับงาน**
- **ลบ (`OnDeleteData :302-344`)**: `DELETE BCAPDepositSpecial` → คืนยอดเช็คจ่าย (`OnDepositUpdateChqAmount(1,…,-1)`) → `apRecalculate(เจ้าหนี้,"##CALC_EARNEST_BALANCE##",-100)` → `DeleteInputTax(0,เลขที่,1)` → `DeleteAPWTax(0,เลขที่)` → `DeletePayment(8,เลขที่)` → `DeleteAccount(เลขที่,42)` → log `DELETE_PAYDEPOSIT` → ล้างจอ.
- **ยกเลิก (`BillCancelUpdateData :261-301`)**: `UPDATE BCAPDepositSpecial SET BillBalance=0, IsCancel=1` → `apRecalculate(...)` → `CancelInputTax(0,เลขที่,1)` → `DeleteAPWTax(0,เลขที่)` → `DeletePayment(8,เลขที่)` ภายใต้ธง `m_bCancelMoney` → **`ReverseAccount(เลขที่,42)`** (ทำรายการบัญชีกลับรายการ แทนการลบ) → log `CANCEL_PAYDEPOSIT` → ล้างจอ. (ไม่เรียก `OnDepositUpdateChqAmount` ตอนยกเลิก — ดู ข้อสังเกต 2.)

**H. ผู้ใช้เอกสารนี้ต่อ (การนำเงินล่วงหน้าไปใช้)**: ใบที่ `BillBalance > 0 AND IsCancel=0` ของเจ้าหนี้รายนั้น จะถูกเสนอในหน้าจ่ายชำระหนี้เป็นรายการ "จ่ายเงินล่วงหน้า" (`MyType=8, BillType=8`; ตัวกรอง `apfrmappayment1view.cpp:3548-3562`, ดึงยอด `:3893-3896`) และในหน้าเลือกเอกสารหลายใบ (`MyDlgMultiSelectDocument.cpp:467,485`); เมื่อบันทึกใบจ่ายชำระหนี้ → `BCAPDepositSpecial.BillBalance −= ยอดที่นำไปหัก` และ `BCAP.EarnestAmount −= ยอดเดียวกัน` (`OnProcessEarnest :3618-3650`; เรียกซ้ำตอนแก้/ลบ/ยกเลิกใบจ่ายชำระหนี้ด้วย `nMak = -1/+1`). หน้ารับคืน (5807) ใช้ `BillBalance` ของใบนี้ตรวจยอดสูงสุดที่คืนได้ (`POFrmEarnestSpecailReturn.cpp:659`). เครื่องมือคำนวณยอดคงเหลือใหม่: `CMyBillReCalc::OnProcessAPEarnest` (`MyDebtReCalc.cpp:2536-2575`) = `NetAmount − Σ PaymentSub(BillType=8) − Σ ReturnDepSpecial` (ไม่ต่ำกว่า 0).

**I. ข้อสังเกตจากต้นฉบับ (อย่าคัดลอกเป็นข้อกำหนด)**
1. จอนี้ไม่มีภาษี แต่โค้ดยังเรียก `CancelInputTax/DeleteInputTax/DeleteAPWTax` ตอนยกเลิก/ลบ และตัวพิมพ์มีตัวแปร `BASE_WTAX`/`WTAX_AMOUNT_WORD` อ่านจาก `BCAPWTaxList SaveFrom=0` (`:839-860`) — เป็นของที่คัดลอกมาจาก 5811; เฉพาะ `SumOfWTax` (ที่ได้จากหน้าจอจ่ายเงิน F7 `myespeciallyformview.cpp:4590-4594`) เท่านั้นที่ถูกเก็บในหัว; ไม่มีการเขียน `BCAPWTaxList` ตอนบันทึก → **ยังไม่ verify** ว่าหน้าจอจ่ายเงินของ 5806 อนุญาตกรอกหัก ณ ที่จ่ายได้หรือไม่.
2. ตอนยกเลิกไม่คืนยอดเช็คจ่าย (`BCAP.ChqOnHand`) แต่ตอนลบ/แก้ไขคืน — ไม่สอดคล้อง (อาจแก้ที่ `DeletePayment` — **ยังไม่ verify**).
3. ป้าย "วันที่ส่งของ" ผิดความหมาย (เก็บวันครบกำหนด) และป้าย "รหัสพนักงาน" ใช้ `Sale_Code` — BC ควรใช้ป้ายตามหน้าที่ แต่ **ใช้ถ้อยคำ Champ ตามกฎ champ-wording** ในรอบแรก.
4. `OnProcessAPEarnest` ในเครื่องมือคำนวณใหม่เขียน `UPDATE BCArDepositSpecial` (ตารางฝั่งลูกหนี้) แทน `BCAPDepositSpecial` (`MyDebtReCalc.cpp:2546`) — บั๊กต้นฉบับ (ไม่ทำอะไรกับฝั่งเจ้าหนี้ แต่ทำให้ยอดคงเหลือใบลูกหนี้เลขที่เดียวกันถูกรีเซ็ตได้ถ้าเลขที่ซ้ำข้ามระบบ); **อย่าคัดลอก**.
5. ข้อความ "Can not save" เป็นอังกฤษตายตัว (`:607`) ไม่ผ่านตารางภาษา — BC ต้องใช้ key ข้อความผิดพลาดภาษาไทย.
6. ข้อตรวจ `IsConfirm` ในยกเลิกซ้ำกันสองชั้น (`:1412-1419`).

### 1.17 รายการ 5807 "บันทึกรับคืนเงินจ่ายล่วงหน้า" (เจ้าหนี้คืนเงินที่เราจ่ายล่วงหน้าไว้ — ผูกกับใบจ่ายเงินล่วงหน้า 5806)

- **ฟอร์ม**: `CPOFrmEarnestSpecailReturn` (`POFrmEarnestSpecailReturn.cpp` 1,421 บรรทัด; `MainFrm.cpp` `OnPosysTrans22` ← `ID_POSYS_TRANS_22`; dialog เดียว `IDD_PO_FRM_EARNEST_RETURN` 360x241 `BC5Account.rc:14470`; รายการเลือกใบจ่ายล่วงหน้า `APEarnestSpecailConfig`). ชื่อหน้าต่าง key `20022` "บันทึกรับคืนเงินรับล่วงหน้า / จ่ายล่วงหน้า" (`:32`; ข้อความรันไทม์มาจากตารางภาษา — ไม่อยู่ใน repo → ป้ายเมนูยึดตาม atlas "บันทึกรับคืนเงินจ่ายล่วงหน้า").
- **ตาราง**: `BCAPReturnDepSpecial` ตารางเดียว ไม่มีตารางรายการ — `DocNo, DocDate, ApCode, DepartCode, SaleCode, MyDescription, ExcessAmount1/2, OtherIncome, OtherExpense, CurrencyCode, ExChangeRate, EarnestNo(เลขที่ใบจ่ายล่วงหน้าที่อ้าง), EarnestAmount(ยอดใบเดิม), EarnestBalance(= ยอดใบเดิม − ยอดคืนครั้งนี้), NetAmount(ยอดที่รับคืน), TotalAmount, BillBalance, SumOfWTax, SumCashAmount, SumChqAmount, PettyCashAmount, SumBankAmount, AllocateCode, ProjectCode, BillGroup, RecurName, GLFormat, IsConfirm, IsCancel, CreatorCode, CreateDateTime, LastEditorCode, LastEditDateT` (`OnSaveData :467-604`). **ไม่เก็บ `GLTransData`** (ดู ข้อสังเกต 3). รายการรับเงินจริง → `BCRecMoney` (`SaveFrom=9`; `OnSaveReceive(9,…)`/`LoadReceive(9,…)`/`DeleteReceive(9,…)`).
- **หลักการ**: เป็นเอกสารฝั่ง **"รับเงินเข้า"** (ใช้หน้าจอรับเงิน `OnReceive`, `IsMoneyOk(0,…)`) — เจ้าหนี้คืนเงินจ่ายล่วงหน้า (เงินสด/เช็ครับ/โอนเข้า) → **ตัดยอดคงเหลือของใบจ่ายล่วงหน้า (5806) ลง** และลงบัญชีกลับด้านของ 5806.

**A. หัวเอกสาร (x,y ตาม rc: ซ้าย x=77, ขวา x=244; ไม่มีกริด)**

| # | ป้ายบนจอ | ช่อง/ชนิด (y,x) | จำเป็น / สิทธิ์ | ค่าเริ่มต้น / กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่เอกสาร: (`DOCNO`) | รันเลข `DocNo` (27,77) | ใช่ | รันตามรูปแบบ `BCAPReturnDepSpecial` + กลุ่มเอกสาร; ปิดช่องถ้าไม่มีสิทธิ์แก้เลขที่; **ยาวเกิน 20 ตัว** → `DOCNO_MORE_LENG` | `:131-134,1252-1293` |
| 2 | วันที่เอกสาร: | วันที่ไทย (27,244) | ใช่ | = วันทำงาน; ล็อกถ้าไม่มีสิทธิ์แก้วันที่ (ยกเว้น `sa`); งวดปิดเตือนตอนเปิดจอ/บันทึก/ลบ/ยกเลิก | `:135-137,184-187,735` |
| 3 | **เลขที่ใบจ่ายเงินล่วงหน้า:** (`DOCNO_PAY`) | Link `EarnestNo` (40,77) — คลิกลิงก์/กด → รายการใบจ่ายล่วงหน้า | ใช่โดยอ้อม — รหัสเจ้าหนี้ล็อกและได้ค่าจากใบที่เลือกเท่านั้น จึงบันทึกไม่ได้ถ้ายังไม่เลือกใบ (`CANNOT_SAVE_NOTAPCODE`) | รายการเลือก = ใบจ่ายล่วงหน้า (5806) ที่ `BillBalance > 0` (กรองรหัสเจ้าหนี้ที่กรอกอยู่); เลือก/พิมพ์เลขที่แล้วออกจากช่อง → **ดึงจากใบเดิม**: รหัสเจ้าหนี้+ชื่อ, พนักงาน, แผนก, โครงการ, การจัดสรร, สกุลเงิน, อัตรา, **ยอดใบเดิม (`NetAmount`)**, **ยอดคงเหลือ (`BillBalance`)**, ยอดเจ้าหนี้คงเหลือ (`BCAP.DebtAmount`) | `:139,976-1041` |
| 4 | กลุ่มเอกสาร: | ข้อความ (40,244) | — | → `BillGroup` | `:646` |
| 5 | **คืนเงินให้กับ:** | รหัสเจ้าหนี้ (53,77) + ชื่อ (53,145) **ล็อกทั้งคู่ (`WS_DISABLED`)** | ใช่ (`CANNOT_SAVE_NOTAPCODE`, ไม่พบ `NOT_FOUND_AP`) | ได้ค่าจากใบจ่ายล่วงหน้าที่เลือกเท่านั้น (พิมพ์เองไม่ได้) | `rc:14478-14479; :313-366` |
| 6 | รหัสพนักงาน: (`Sale_Code`) | Combo-table `BCSale` (66,77) | — | แก้ได้ | `:144` |
| 7 | แผนก: / โครงการ: / การจัดสรร: | Combo-table `BCDepartment` (66,244) / `BCProject` (78,77) / `BCAllocate` (78,244) | **อ่านอย่างเดียว (`ES_READONLY`, ข้ามด้วย Tab)** | มาจากใบจ่ายล่วงหน้าที่เลือก; แต่ยังแก้ภายหลังได้ผ่านรายการ "แก้ไขเฉพาะช่อง" (แผนก/การจัดสรร/โครงการ/กลุ่มเอกสาร/หมายเหตุ — `AddCanEditField :177-181`) | `:153-155,177-181` |
| 8 | รหัสสกุลเงิน: / อัตราแลกเปลี่ยน: | Combo (90,77) / ตัวเลข 9 ตำแหน่ง (90,244) **อ่านอย่างเดียว** | — | มาจากใบเดิม; ปิดทั้งคู่ถ้าบริษัทไม่เปิด `MultiCurrency`; บันทึก: อัตรา ≤ 0 → 1.0 | `:141-150,1218-1246,492-499` |
| 9 | คำอธิบาย: | ข้อความ 235x45 (102,77) | — | → `MyDescription` | `:489` |
| 10 | เงินล่วงหน้าสุทธิ: (`EarnestAmount`) | ตัวเลข ล็อก (147,77) | — | ยอดใบเดิม | `:658,1007` |
| 11 | เงินล่วงหน้าคงเหลือ: (`EarnestBalance`) | ตัวเลข ล็อก (147,244) | — | `BCAPDepositSpecial.BillBalance` ของใบเดิม (**ตอนเปิดเอกสารที่บันทึกแล้วจะแสดงยอดปัจจุบัน ไม่ใช่ค่าที่บันทึกไว้**) | `:659` |
| 12 | **จำนวนเงิน:** (`NetAmount`) | ตัวเลข (160,77) | ใช่ (> 0) | ช่องเดียวที่ผู้ใช้กรอกยอดรับคืน → `NetAmount` = `TotalAmount` = `BillBalance` | `:146,506-513` |
| 13 | ยอดเจ้าหนี้คงเหลือ: | ตัวเลข ล็อก (160,244) | — | `BCAP.DebtAmount` (เพื่ออ้างอิงอย่างเดียว; ไม่ถูกแก้โดยเอกสารนี้) | `:660` |

**B. คีย์ลัด** (`PreTranslateMessage :1160-1217`; เมนูพ่วง `OnMore :1042-1094`)

| คีย์ | การทำงาน |
|--|--|
| **F2** | บัญชีแยกประเภท (`OnAccount`) — ถูกบล็อกถ้าผลิตภัณฑ์ตัดโมดูล GL |
| **F5 + Ctrl** | ยกเลิกเอกสาร (`OnConfirmCancel` → `BillCancel`); `IsConfirm=1` → `CANNOT_CANCEL` |
| **F5 + Shift** | Confirm เอกสาร (ต้องมีสิทธิ์ `bCanConfirmBill` ไม่งั้น `NOT_CONFIRM_NO`) |
| **F7** | **บันทึกรับเงิน** (`OnRec` → `OnReceive(ยอดสุทธิ, TRUE, คำอธิบาย, เลขที่, รหัสเจ้าหนี้)`) — ระบุวิธีรับเงินคืน (เงินสด/เช็ครับ/โอนเข้าธนาคาร/เงินสดย่อย) ให้ครบเท่ายอด; รายละเอียดหน้าจอรับเงิน → `04-ar.md:987` |
| **F12** | **ถูกกลืนไว้ ไม่ทำอะไร** (แม้มี `OnDocumentFlow` ผูกกับปุ่มแถบเครื่องมือ `ID_ANALYSE_FLOW` — ข้อสังเกต 4) |
| เมนูพ่วง `OnMore` | F2=บัญชีแยกประเภท · F7=บันทึกรับเงิน · F11=บันทึก · Shift+F5=Confirm · Ctrl+F5=ยกเลิก (5 รายการ; ซ่อนข้อแรกถ้าตัด GL) |
| ปุ่มพิมพ์/ใหม่/ค้นหา/ข้อมูล/Recurring | `ID_HK_PRINT/NEW/FIND`, `ID_REC_INFO` ("ยอดคงเหลือ ^…" `:1147`), `IDC_RECUR` |

**C. กติกาก่อนบันทึก (`ConfirmSaveData :313-466` — ลำดับจริง)**: (1) งวดปิด · (2) `GLFormat` ว่าง → ค่าเริ่มต้นของ FormatGroup 43 (`ADVANCERET_AP`, `GetGLFormat :1361-1365`) · (3) เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · (4) แก้เลขที่แล้วซ้ำกับเอกสารอื่น `DOCNO_DUPLICATE` · (5) รหัสเจ้าหนี้ว่าง `CANNOT_SAVE_NOTAPCODE` / ไม่พบ `NOT_FOUND_AP` · (6) แก้ไขเอกสาร: ยกเลิกแล้ว `NOT_SAVE_ISCANCEL`, Confirm แล้ว `NOT_SAVE_ISCONFIRM` · (7) **วิธีรับเงินรวม ≠ ยอดสุทธิ → `AMOUNT_NOT_EQUAL_RECEIVE` แล้วเปิดหน้าจอรับเงินให้ทันที** (`IsMoneyOk(0,…)`) · (8) ยอดสุทธิ ≤ 0 → `CANNOT_SAVE_WORNG` · (9) เพิ่มใหม่: เลขที่ซ้ำ → ถามรันเลขใหม่/ล้างเลขที่ (ข้อความอังกฤษตายตัว) · (10) ตรวจจำนวนรายการเอกสารของรุ่นผลิตภัณฑ์ (`CheckTransactoin` ตามรุ่น Pro/Express — ข้ามได้).
> **ไม่มีการตรวจ**: (ก) ยอดรับคืน ≤ ยอดคงเหลือของใบจ่ายล่วงหน้า และ `EarnestNo` ต้องมีอยู่จริง/ตรงเจ้าหนี้, (ข) เอกสารผ่าน GL แล้วห้ามแก้ (ไม่มี `IsPostGL` ใน `ConfirmSaveData` ต่างจาก 5806) — ข้อสังเกต 1-2. BC ต้องใส่ guard เหล่านี้ (ตามหลักบัญชี) แม้ Champ ไม่มี.
**ลบ (`ConfirmDeleteData :209-244`)**: เลขที่ว่าง `CANNOT_DELETE` · งวดปิด · ยกเลิกแล้ว `NOT_DELETE_ISCANCEL` · Confirm แล้ว `NOT_DELETE_ISCONFIRM` · เช็ครับที่ใช้ในใบนี้ถูกนำไปตัดหลายใบ (`IsChqRefence(เลขที่,9,0)`) → `CHQNOT_EDIT_DELETE_CANCEL` · ถามยืนยัน `CONFIRM_DELETEDATA`. **ไม่ตรวจ `IsPostGL`** (ต่างจาก 5806).
**ยกเลิก (`OnConfirmCancel :1367-1393`)**: เอกสารที่บันทึกแล้ว · งวดปิด · **ผ่าน GL แล้ว → `CANNOT_CANCEL`** · Confirm แล้ว `MSG_CANNOT_CANCEL`.

**D. ผลข้างเคียงตอนบันทึก (`OnSaveData :467-604`)**
1. แก้ไข: คืนยอดเช็ครับเดิม `OnDepositUpdateChqAmount(1, เจ้าหนี้เดิม, SumChqAmount เดิม, -1)`.
2. `StartTransaction` → `InsertDB` / `UpdateDB`. ในโหมดแก้ไข ก่อนอัปเดต: **`OnProcessEarnest(เลขเดิม, -1)`** (คืนยอดคงเหลือของใบจ่ายล่วงหน้าเท่ากับยอดรับคืนเดิม) + `apRecalculate(เจ้าหนี้เดิม, "##CALC_EARNEST_BALANCE##", -100)`. ล้มเหลว → `Abort` + "Can not save".
3. `EarnestBalance` = ยอดใบเดิมที่แสดงบนจอ − ยอดรับคืนครั้งนี้ (`dEarnestAmount − dNetReturn`, `:506-513`) — **ใช้ยอดทั้งใบ ไม่ใช่ยอดคงเหลือ** ถ้าเคยหัก/คืนบางส่วนมาก่อนค่านี้จึงไม่ใช่ยอดคงเหลือจริง (ข้อสังเกต 5).
4. หลัง commit: `OnDepositUpdateChqAmount(1, เจ้าหนี้, ยอดเช็ครับ, +1)` → **`OnSaveReceive(9, เลขเดิม, เลขใหม่, วันที่, รหัสเจ้าหนี้, สกุลเงิน, อัตรา, แผนก, พนักงาน, โครงการ)`** (เขียนรายการรับเงินลง `BCRecMoney` SaveFrom=9) → **`OnProcessEarnest(เลขใหม่, +1)`** → `apRecalculate(เจ้าหนี้, "##CALC_EARNEST_BALANCE##", -100)`.
5. **`OnProcessEarnest(เลขที่, nMak)`** (`:1319-1344`): อ่านใบรับคืนนี้ (`IsCancel=0`) แล้ว `UPDATE BCAPDepositSpecial SET IsConfirm=0, BillBalance = BillBalance − (nMak × NetAmount) WHERE DocNo = EarnestNo` → ตั้ง `IsConfirm=1` เมื่อ `BillBalance ≠ NetAmount` (มีการใช้/คืนแล้ว) มิฉะนั้น `IsConfirm=0`. (**บันทึก = +1 ลดยอดคงเหลือ; ลบ/แก้ = −1 คืนยอด**).
6. **ลงบัญชี**: `GLTransData` ไม่ถูกเก็บ → `m_nManualKey` เป็น 0 เสมอเมื่อเปิดจอใหม่ → `m_GLTLSProcess{ m_nSource=43; m_nGLGroup=46 }.ProcessConditionGroup43()` (engine group 46 → `ProcessConditionGroup43`; FormatGroup 43 `ADVANCERET_AP`); ถ้าแก้บัญชีมือ (F2 ในเซสชันเดียวกัน → `m_nManualKey=1`) → `OnSaveAccount(เลขเดิม, เลขใหม่, วันที่, 43, เจ้าหนี้, คำอธิบาย)`. ไม่มีภาษี/สต็อก/`processdoc`.
7. `BCAP.EarnestAmount` คำนวณใหม่ด้วย `calcEarnetAmount` (ลบ Σ ใบรับคืนนี้ออกจากยอดเงินล่วงหน้า — สูตรใน §1.16 ข้อ F5).

**E. ลบ/ยกเลิก — ลำดับงาน**
- **ลบ (`OnDeleteData :277-312`)**: `OnProcessEarnest(เลขที่,-1)` (คืนยอดคงเหลือให้ใบจ่ายล่วงหน้า) → `DELETE BCAPReturnDepSpecial` → คืนยอดเช็ครับ (`OnDepositUpdateChqAmount(1,…,-1)`) → `apRecalculate(...)` → `DeleteReceive(9,เลขที่)` → `DeleteAccount(เลขที่,43)` → log `DELETE_PAYDEPOSIT` → ล้างจอ.
- **ยกเลิก (`BillCancelUpdateData :245-276`)**: `UPDATE ... SET BillBalance=0, IsCancel=1` → (**คำสั่งลบรายการรับเงินถูกคอมเมนต์ทิ้ง** `:253-256`) → `OnProcessEarnest(เลขที่,-1)` → `apRecalculate(...)` → `ReverseAccount(เลขที่,43)` → log `CANCEL_PAYDEPOSIT` → ล้างจอ. ข้อสังเกต 6.

**F. F12 / รายการที่เกี่ยวข้อง**: `OnDocumentFlow :1294-1318` ใช้คำสั่งเดียวกับ 5806 (หา "ใบรับคืน" และ "ใบจ่ายชำระหนี้" ที่อ้างถึง `เลขที่นี้`) — สำหรับใบรับคืนจะไม่พบอะไร และถูกกลืนที่ F12 (ข้อสังเกต 4). ใบนี้ถูกอ้างโดย: ใบจ่ายล่วงหน้า 5806 (ตัดยอดคงเหลือ), ยอด `BCAP.EarnestAmount`, F12 ของ 5806, และ `ยกเลิก 5806` ถูกห้ามถ้ามีใบรับคืนที่ยังไม่ยกเลิก (`CANNOT_CANCEL_USE`).

**G. ข้อสังเกตจากต้นฉบับ (อย่าคัดลอกเป็นข้อกำหนด — BC ควรทำให้ถูกต้อง)**
1. **ไม่มี guard ยอดรับคืน ≤ ยอดคงเหลือ** → ผู้ใช้กรอกเกินได้ ทำให้ `BillBalance` ของใบจ่ายล่วงหน้าติดลบ (SQL `BillBalance − NetAmount` ไม่มี `MAX(0)`) ส่วนเครื่องมือคำนวณใหม่ปัดเป็น 0 (`MyDebtReCalc.cpp:2569-2570`) — **BC ต้องห้าม** (ยอดรับคืนต้อง ≤ `BillBalance` ณ เวลาบันทึก + ตอนแก้ไขให้ถือ "คงเหลือ + ยอดรับคืนเดิม").
2. ผู้ใช้เปลี่ยน/ล้างเลขที่ใบจ่ายล่วงหน้า (`EarnestNo`) หลังเลือกแล้ว ช่องรหัสเจ้าหนี้ที่ล็อกยังค้างค่าเดิม → บันทึกได้โดย `EarnestNo` ว่างหรือผิด (ไม่มีข้อตรวจว่า `EarnestNo` มีอยู่จริง/เป็นของเจ้าหนี้รายนี้/ยังไม่ยกเลิก) — BC ต้องตรวจตอนบันทึกเสมอ.
3. `GLTransData` ไม่ถูกบันทึก/โหลด (`m_nManualKey` มีใช้เฉพาะ `:584,729`) → บัญชีมือที่แก้ใน F2 หายเมื่อเปิดแก้ไขแล้วบันทึกใหม่ (ระบบสร้างอัตโนมัติทับ).
4. `OnDocumentFlow` คัดลอกมาจาก 5806 โดยไม่ปรับ และ `F12` ถูกกลืน.
5. `EarnestBalance` ที่บันทึกใช้ "ยอดทั้งใบ − ยอดคืน" ไม่ใช่ "ยอดคงเหลือ − ยอดคืน".
6. **ยกเลิกเอกสาร**: ตั้ง `IsCancel=1` ก่อนแล้วจึงเรียก `OnProcessEarnest(เลขที่,-1)` ซึ่งอ่านเฉพาะแถว `IsCancel=0` → **ไม่พบแถว จึงไม่คืนยอดคงเหลือให้ใบจ่ายล่วงหน้า**; ไม่ลบรายการรับเงิน (`BCRecMoney SaveFrom=9`) และไม่คืนยอดเช็ครับ → ตัวเลขคลาดเคลื่อน (มีทางแก้ด้วยเครื่องมือคำนวณใหม่ `OnProcessAPEarnest`). **ห้ามคัดลอก** — BC ต้องคืนยอดคงเหลือ + ลบ/กลับรายการเงินรับตอนยกเลิก.
7. ตัวกรองรายการเลือกใบจ่ายล่วงหน้าต่อสตริง ` BillBalance>0.0 ` กับ `Code LIKE '…%'` โดยไม่มี `AND` และ `BCAPDepositSpecial` ไม่มีคอลัมน์ `Code` (มีแต่ `ApCode`) (`:976-987`) — เมื่อกรอกรหัสเจ้าหนี้ไว้ก่อน อาจเป็น SQL ผิด; **ยังไม่ verify** ว่า `Browse2`/`SetExFilter` ปรับให้เองหรือไม่.
8. ข้อความ "Can not save"/"Document No is duplicated" เป็นอังกฤษตายตัว.

### 1.18 รายการ 5048 "บันทึกต้นทุนแฝง" (Weight cost — กระจายค่าใช้จ่ายซื้อ เช่น ค่าขนส่ง/ค่านำเข้า เข้าเป็นต้นทุนของสินค้าในบิลซื้อหลายใบ)

- **ฟอร์ม**: `CApFrmBCWeightCostView` (`ApFrmBCWeightCostView.cpp` 1,285 บรรทัด; `MainFrm.cpp` `OnApBCWeightCost` ← `ID_APSYS_TRANS_6` — id ชุดเมนูเจ้าหนี้ แต่อยู่ในกลุ่ม PO ตาม atlas-01; dialog `IDD_AP_FRM_BCWEIGHTCOST` 347x231 `BC5Account.rc:9713`).
- **ตาราง**: หัว `BCWeightCost` (`DocNo, DocDate, RefDocNo, ExpDescription, WeightBy, WeightAmount, FROMPO, TOPO, WeightCostType`) + รายการ `BCWeightCostSub` (`ExpCostNo, DocDate, RefDocNo, BuyDocNo, ExpDescription, ItemCode, ItemQty, ItemAmount, Amount, LineNumber, WeightCostType, BillLineNumber, BillRowOrder`) (`OnSaveData :651-773`); ผลลัพธ์ไปเก็บที่คอลัมน์ **`BCAPInvoiceSub.SumOfExpCost`** (ต้นทุนแฝงสะสมต่อบรรทัดของบิลซื้อ — มีใน `POFrmPurchase.cpp:269,1541,2051` เป็น shadow col 27).
- **เงื่อนไขใช้งาน**: ต้นทุนแฝงเข้าต้นทุนสินค้า **เฉพาะเมื่อตัวเลือกบริษัท "ใช้ระบบต้นทุนแฝง" (`UseExpCost = 1`) เปิดอยู่** — ตั้งที่หน้าตั้งค่าระบบ (`myfrmsysconfig.cpp:912,1165,1230`); ตัวประมวลผลสต็อกเอา `HomeAmount + SumOfExpCost` เป็นมูลค่ารับเข้า (`BCStools.cpp:1170-1175`, `bcdumpprocjob.cpp:2722-2733`, เฉพาะ `nTransType == 2` = ซื้อ). ปิดตัวเลือกนี้ → บันทึกได้แต่ต้นทุนไม่เปลี่ยน.
- **ไม่มี**: GL, ภาษี, การตรวจงวดปิด, `IsCancel`/`IsConfirm`, ธุรกรรม (`StartTransaction`) — ดู ข้อสังเกต.

**A. หัวเอกสาร (x,y ตาม rc: ซ้าย x=67, ขวา x=208)**

| # | ป้ายบนจอ (key) | ช่อง/ชนิด (y,x) | จำเป็น | กติกา | หลักฐาน |
|--|--|--|--|--|--|
| 1 | เลขที่เอกสาร: (`DOCNO`) | รันเลข `DocNo` (7,67) | ใช่ (`CANNOT_SAVE_NOTDOCNO`) | รูปแบบเอกสาร `BCWeightCost`; เพิ่มใหม่แล้วซ้ำ → ถามรันเลขใหม่ ("Document No is duplicated", อังกฤษตายตัว) | `:165-166,55-100` |
| 2 | วันที่เอกสาร: | วันที่ไทย (7,208) | — | = วันทำงาน; ล็อกถ้าไม่มีสิทธิ์แก้วันที่ (ยกเว้น `sa`); **ใช้เป็นวันที่ประมวลผลสต็อกใหม่** | `:169,234-236` |
| 3 | ~อ้างถึงเอกสารค่าใช้จ่าย (`REF_DOCNO_PAY`) | ข้อความ (19,67) | — | เลขที่ใบค่าใช้จ่าย/ใบแจ้งหนี้ค่าขนส่งที่มาของต้นทุน (ข้อความอิสระ ไม่ตรวจกับเอกสารใด) → `RefDocNo` | `:579` |
| 4 | ~คิดตาม (`WEIGHT_BY`) | Combo 2 ค่า (19,208) | ใช่ (ค่าเริ่มต้น 0) | **ตามปริมาณ (0)** / **ตามมูลค่าสินค้า (1)** (`BY_QUANTITY`, `BY_AMOUNTITEM`) → `WeightBy` | `:176-182,579` |
| 5 | ~จากบิลซื้อ (`FROM_BILL_BUY`) / ~ถึงบิลซื้อ (`TO_BILL_BUY`) | Link `FROMPO`/`TOPO` (31,67)/(31,208) | — | ลิงก์เปิดรายการเลือกบิลซื้อ (`PurchaseConfig`, กรอง `GRIRBillStatus <> 2` = ไม่รวมใบตั้งหนี้ IR) เพื่อกำหนดช่วงเลขที่ให้ F9 | `:171-172,1066-1100` |
| 6 | ~มูลค่าต้นทุนเงินบาท (`SET_VALUE`) | ตัวเลขเงิน (43,67) | — | ยอดต้นทุนแฝงรวมที่จะกระจาย (บาท) → `WeightAmount`; หลัง F5 ถ้าผู้ใช้แก้ยอดต่อบรรทัดเอง ยอดรวมนี้ไม่ถูกคำนวณกลับอัตโนมัติ (ฟังก์ชัน `ProcessDetail` ไม่มีผู้เรียก) | `:173,774-786` |
| 7 | วิธีคิดต้นทุนแฝง: (`METHOD_WEIGHT`) | Combo 2 ค่า (43,208) | ใช่ (ค่าเริ่มต้น 0) | **ต้นทุนแฝงด้านบริหาร (0)** / **ต้นทุนแฝงด้านบัญชี (1)** (`METHOD_WEIGHT1/2`) → `WeightCostType`. **เก็บไว้เฉย ๆ — ไม่มีโค้ดอื่นอ่านค่านี้** (ค้นทั้งโครงการแล้ว) | `:183-189,586` |
| 8 | ~หมายเหตุ (`WEIGHT_REMARK`) | ข้อความ (55,67) | — | → `ExpDescription` ของหัว; **F5 คัดลอกลงช่องหมายเหตุของทุกแถวที่ยังว่าง** | `:580,890-1003` |
| ปุ่ม | **F9 = แสดงรายการ** (289,15) · **F5 = ประมวลผล** (289,33) · **F7 = Clear Data** (289,51) | ปุ่มข้างหัวเอกสาร | | ดู B-C | `rc:9713+` |

**B. กริดรายการ (7 คอลัมน์ `:199-212`, ปุ่มซ้ายเลขบรรทัดอัตโนมัติ, เส้นสีเหลืองอ่อน)**

| # | หัวคอลัมน์ | กว้าง | สิทธิ์ | ความหมาย |
|--|--|--|--|--|
| 0 | บิลซื้อ (`BUY_BILL`) | 90 | แก้ได้ + ลิงก์ (ยาว ≤ 25) | เลขที่บิลซื้อ/ใบรับสินค้า: **พิมพ์เลขที่แล้วออกจากช่อง → ระบบเติมทุกบรรทัดสินค้าของบิลนั้น** (ยกเว้นรหัสสินค้าที่มีในกริดแล้ว) `OnListAddData :360-431`; หรือกด F-link เลือกหลายบิล (`OnListFindData :432-524`, กรอง `IsCancel=0 AND GRIRBillStatus <> 2`, ตัดบิลที่อยู่ในกริดแล้ว) |
| 1 | วันที่เอกสาร | 80 | ล็อก | `BCAPInvoice.DocDate` |
| 2 | หมายเหตุ (`REMARK`) | 140 | แก้ได้ | → `ExpDescription` ของแถว |
| 3 | รหัสสินค้า | 100 | ล็อก | `BCAPInvoiceSub.ItemCode` |
| 4 | ปริมาณสินค้า (`QUANTITY_ITEM`) | 80 | ล็อก ขวา | `BCAPInvoiceSub.Qty` → `ItemQty` |
| 5 | มูลค่าสินค้า (`SUM_TOTAL`) | 100 | ล็อก ขวา | `HomeAmount` (ถ้า > 0) มิฉะนั้น `Amount` → `ItemAmount` |
| 6 | ยอดที่คิดได้ (`WEIGHT_AMOUNT`) | 100 | แก้ได้ ขวา | **ส่วนแบ่งต้นทุนแฝงของบรรทัดนี้** (F5 คำนวณให้ แก้เองได้) → `Amount` |

> shadow 3 ช่อง (`:213-215`): 0 RowOrder · 1 `BillLineNumber` (`BCAPInvoiceSub.LineNumber`) · 2 `BillRowOrder` (`RowOrder`). ดึงบรรทัดจาก `BCAPInvoiceSub WHERE TransState=0 AND IsCancel=0 ORDER BY DocNo,DocDate,ItemCode,UnitCode`. (แถวสินค้าบริการ `StockType = 1/7` และรหัส `.` ไม่ถูกนับเป็นสินค้าสำหรับประมวลผลสต็อกใหม่ `:67-81`.)

**C. คีย์ลัดและปุ่ม** (`PreTranslateMessage :787-848`; เมนู `OnMore :1223-1250`)

| คีย์ | การทำงาน |
|--|--|
| **F9** | **แสดงรายการ** (`OnItemProcess`): ต้องกรอก "จากบิลซื้อ" (ว่าง = ไม่ทำอะไร); มี "ถึงบิลซื้อ" → `DocNo BETWEEN จาก AND ถึง` มิฉะนั้น `DocNo = จาก`; โหลดทุกบรรทัดสินค้า (`TransState=0 AND IsCancel=0`) **ต่อท้ายกริด** (ไม่ล้างของเดิม; ไม่ตัดแถวซ้ำ) `:1004-1065` |
| **F5** | **ประมวลผล** (`OnWeight`): กระจาย "มูลค่าต้นทุนเงินบาท" ลงทุกแถวที่มีเลขที่บิล — ตามปริมาณ: `ส่วนแบ่ง = ปริมาณแถว ÷ ปริมาณรวม × ยอดรวม`; ตามมูลค่า: `มูลค่าแถว ÷ มูลค่ารวม × ยอดรวม`; ปัดตามทศนิยมเงินของบริษัท; **แถวสุดท้ายรับเศษ** (= ยอดรวม − ผลรวมแถวก่อนหน้า เพื่อให้รวมเท่ายอดทุกสตางค์); แถวที่หมายเหตุว่างได้ข้อความจากช่องหมายเหตุหัว; แถบสถานะล่าง "ยอดรวมปริมาณ … ยอดรวมมูลค่า … รวมยอดที่คิดได้ …" `:890-1003` |
| **F7** | **Clear Data** (`OnClearData`): ล้างทั้งหน้าจอ **โดยไม่ถามยืนยัน** |
| **F12** | **Update weight cost to all bill purchase** (`UpdateWeightApInvoice`): ถามยืนยัน (ไทย: "การประมวลผลจะ Update รายการต้นทุนแฝงให้กับบิลซื้อทั้งหมดที่มีต้นทุนแฝง เมื่อเสร็จขั้นตอนนี้ให้ทำการคำนวณยอดคงเหลือรวมของสินค้าใหม่ด้วย") → รวม `Σ BCWeightCostSub.Amount` แยกตาม `BuyDocNo, ItemCode, BillLineNumber` ทุกเอกสาร แล้ว **ตั้ง** `BCAPInvoiceSub.SumOfExpCost = ผลรวม` (ซ่อมยอดสะสมที่เพี้ยน); **ไม่เรียกประมวลผลสต็อกใหม่** — ผู้ใช้ต้องคำนวณยอดคงเหลือสินค้าเอง `:1174-1222` |
| Ctrl+Enter ที่ช่องหมายเหตุ | ข้ามไปกริด (0,0) |
| PageUp / PageDown | โฟกัสเลขที่เอกสาร (หรือวันที่ถ้าเลขที่ถูกล็อก) / กริด (0,0) |
| ปุ่มพิมพ์/ใหม่/ค้นหา | `OnHKPrint/New/Find`; เมนูพ่วง `OnMore` = F5, F7, F9, F11 บันทึก, F12 (5 รายการ) |

**D. กติกาก่อนบันทึก (`ConfirmSaveData :55-100`)**: เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` · **ไม่มีรายการ** (`GetRecordCount() < 1`) → `CANNOT_SAVE_NO_DETAIL` + โฟกัสกริด · รวบรวมรหัสสินค้าที่เป็นสินค้าสต็อก (ไม่ใช่ `StockType 1/7`, ไม่ใช่ `.`) ไว้ประมวลผลสต็อกใหม่ · เลขที่ซ้ำตอนเพิ่มใหม่ → ถามรันเลขใหม่/ล้าง. **ไม่ตรวจ**: งวดบัญชีปิด, ยอดรวมแถว = ยอดมูลค่าต้นทุน (F5 ไม่ได้บังคับก่อนบันทึก), แถวที่ยอด 0, บิลถูกยกเลิก/เป็นใบตั้งหนี้.
**ลบ (`ConfirmDeleteData :38-53`)**: เลขที่ว่าง `CANNOT_DELETE` → ถามยืนยัน `WANT_DELETE_ITEM` (Yes/No) เท่านั้น — **ไม่ตรวจงวดปิด**.

**E. ผลข้างเคียงตอนบันทึก (`OnSaveData :651-773`)** — **ไม่มีธุรกรรมครอบ** (ถ้าล้มกลางทางจะเหลือข้อมูลค้าง)
1. `InsertDB`/`UpdateDB(BCWeightCost WHERE DocNo=เลขเดิม)`; ล้มเหลว → "Can not save".
2. **ย้อนผลเก่า**: อ่าน `BCWeightCostSub WHERE ExpCostNo = เลขเดิม` แล้วทุกแถว `UPDATE BCAPInvoiceSub SET SumOfExpCost = SumOfExpCost − Amount WHERE DocNo=BuyDocNo AND LineNumber=BillLineNumber AND ItemCode` → `DELETE BCWeightCostSub WHERE ExpCostNo = เลขเดิม`.
3. **เขียนใหม่ต่อแถวที่มีเลขที่บิล**: `INSERT BCWeightCostSub` (`LineNumber = ลำดับแถว`) → สำเร็จจึง `UPDATE BCAPInvoiceSub SET SumOfExpCost = SumOfExpCost + Amount` ตาม (บิล, บรรทัด, รหัสสินค้า).
4. ถ้าบันทึกรายการครบทุกแถว → **`OnStockReprocess(รหัสสินค้า, วันที่เอกสาร)` ต่อสินค้า** (คำนวณต้นทุน/ยอดคงเหลือสินค้าย้อนหลังใหม่ตั้งแต่วันที่เอกสาร) — ไม่ต้องทำเอง (การปรับต้นทุนในคลังที่เคยเขียนไว้ใน `UpdateStock :1102-1173` ถูกคอมเมนต์ทิ้งทุกจุดเรียก `AutoServerProc` ⇒ ใช้ประมวลผลย้อนหลังแทน). log "บันทึกการตัดต้นทุนแฝง".
5. **ลบ (`OnDeleteData :278-332`)**: ย้อน `SumOfExpCost` ของทุกแถว → `DELETE BCWeightCostSub`, `DELETE BCWeightCost` → `OnStockReprocess` ต่อสินค้าที่อยู่ในกริดตอนโหลด → log (ใช้ชื่อ log "ลบจ่ายชำระหนี้" ผิดที่ — ข้อสังเกต).

**F. ข้อสังเกตจากต้นฉบับ (อย่าคัดลอกเป็นข้อกำหนด)**
1. ไม่มีธุรกรรม/ตรวจงวดปิดเลย — **BC ต้องครอบด้วยธุรกรรมเดียวและห้ามแก้ต้นทุนในงวดที่ปิดแล้ว** (ต้นทุนแฝงเปลี่ยนต้นทุนขาย/สินค้าคงเหลือย้อนหลัง → กระทบ GL/งบการเงิน; Champ ไม่มีลงบัญชีอัตโนมัติจากเอกสารนี้ — **ยังไม่ verify** ว่าการประมวลผลสต็อกใหม่สร้างรายการ GL ปรับต้นทุนหรือไม่).
2. `WeightCostType` (บริหาร/บัญชี) เก็บอย่างเดียว ไม่มีผลต่อการคำนวณ — ถ้า BC ต้องการ 2 มุมมองต้นทุน ต้องถามลุงจืด (ห้ามเดา).
3. บิลที่ถูกเลือก **ไม่ตรวจว่าถูกยกเลิก/ผ่านไปแล้ว** นอกจากคำสั่งเติมแถว (`IsCancel=0` ที่ระดับรายการ) และไม่กันการกระจายเงินซ้ำบนบรรทัดเดียวกันในหลายเอกสาร (สะสมใน `SumOfExpCost` ตามจริง).
4. F5 ปัดเศษด้วยแถว "สุดท้ายของกริด" — ถ้าแถวสุดท้ายไม่มีเลขที่บิล เศษจะไม่ถูกจัดสรร (`:890-1003`).
5. ชื่อ log การลบเป็น "ลบจ่ายชำระหนี้" และรหัสหน้าจอเป็นของหน้าจ่ายชำระหนี้ (`IDD_AP_FRM_BCAPPAYMENT_1`) ใน `OnDeleteData`/`UpdatePrintLog` — คัดลอกมาผิด.
6. `ProcessDetail` (รวมยอดแถวกลับเป็นยอดรวม) ไม่มีผู้เรียก; หลังแก้ยอดแถวเอง ผลรวมแถวอาจไม่เท่า `WeightAmount` และบันทึกได้.

### 1.19 รายการ 5049 "ปรับปรุงใบรับสินค้า" (ปรับสถานะใบรับสินค้า GR ที่ถูกอ้างอิงจนครบแล้วให้เป็น "ใช้หมดแล้ว" — เครื่องมือซ่อมสถานะ ไม่ใช่เอกสาร)

- **ฟอร์ม**: `CPOFrmAdjGRStatus` (`POFrmAdjGRStatus.cpp` 688 บรรทัด; `MainFrm.cpp` `OnAPAdjGR` ← `ID_APSYS_TRANS_7`; dialog `IDD_PO_FRM_ADJUST_GR` 345x238 `BC5Account.rc:11548`; ฐาน `CMyBaseFormView` — **ไม่ใช่ฟอร์มเอกสาร** ไม่มีเลขที่/บันทึก/ลบ/ค้นหา: เมนูแถบเครื่องมือบันทึก/ลบ/ถัดไป/ก่อนหน้า/ค้นหา/ใหม่ ถูกปิดทั้งหมดด้วย `OnUpdateRec*` `:94-122`).
- **ตาราง**: อ่าน `BCApInvoice` (หัวใบรับสินค้า GR — `GRIRBillStatus=1`), `BCApInvoiceSub` (`IRRefNo`, `Qty`), `BCIRSUB` (รายการใบตั้งหนี้ IR ที่อ้าง GR ด้วย `GRRefNo`); **เขียน** `BCAPINVOICE.GRBillStatus` เท่านั้น. ความหมายสถานะดู §1.3 (5798): `GRBillStatus` ของ GR = 0 ยังไม่ถูกตั้งหนี้ครบ / 1 ถูกตั้งหนี้ (IR) ครบแล้ว.
- **วัตถุประสงค์**: ซ่อม GR ที่ปริมาณถูกอ้างอิงครบแล้วแต่สถานะยังค้าง 0 (ทำให้ GR นั้นยังโผล่ในหน้าเลือกใบรับสินค้าของ 5799 `GRBillStatus=0`) — ผู้ใช้เลือก GR ที่ต้องการแล้วกด "ปรับปรุง".

**A. ตัวกรอง (x,y ตาม rc: ช่องซ้าย x=80, ช่องขวา x=167; ปุ่มกระทำ x=260)**

| แถว (y) | ป้ายบนจอ | ช่อง | หมายเหตุ |
|--|--|--|--|
| 11 | จากวันที่: / ถึง | วันที่ไทย + ปุ่มปฏิทิน (80) / (167) | `DocDate BETWEEN` (มีแต่ "จาก" = วันเดียว); ปุ่ม **ประมวลผล** (`PROCESS`, 260,11) |
| 23 | จากใบรับสินค้า: / ถึง | Link เลขที่ใบรับสินค้า (80)/(167) (เลือกจากรายการ `BCApInvoice`) | `DocNo BETWEEN` |
| 31 | — | — | ปุ่ม **ปรับปรุง** (`ADJUST`, 260,31) |
| 35 | จากรหัสเจ้าหนี้: / ถึง | Link `BCAP` (80)/(167) | `ApCode BETWEEN` |
| 47 | จากแผนก: / ถึง | Combo-table `BCDepartment` (80)/(167) | `DepartCode BETWEEN` |
| 50 | ตัวเลือก **เลือกทั้งหมด / ไม่เลือกทั้งหมด** (radio, 241/294) + ปุ่มไอคอน OK 2 ปุ่ม (บนขวา 324/335,7) | | ติ๊ก/ถอนติ๊กทุกแถวในตาราง |
| 68 | ตารางผลลัพธ์ (`IDC_PRICE_LIST`) | ListView มีช่องติ๊ก เส้นตาราง เลือกทั้งแถว | ดูข้อ B |

**B. ตารางผลลัพธ์ (7 คอลัมน์ `BuildColumns :153-211`)**: `S` (ช่องติ๊ก 25) · เลขที่เอกสาร (100) · วันที่ (80) · รหัสเจ้าหนี้ (100) · ชื่อเจ้าหนี้ (200) · **เลขที่อ้างอิง** (`REFERANCE_NO` 100 = `BCApInvoiceSub.IRRefNo`) · **วันที่อ้างอิง** (`REFERANCE_DATE` 80 = `BCApInvoice.DocDate` ของเลขที่อ้างอิง). GR หนึ่งใบที่อ้างอิงหลายเลขที่แสดงหลายบรรทัด — **บรรทัดแรกมีเลขที่/วันที่/เจ้าหนี้/ชื่อ บรรทัดถัดไปเว้นว่างช่องเหล่านั้น**; GR ที่ไม่มีรายการอ้างอิงแสดง 1 บรรทัดโดยเลขที่/วันที่อ้างอิงว่าง. ดับเบิลคลิกแถว → **เปิดหน้าจอใบรับสินค้า 5798 (`ID_POSYS_TRANS_6`) ด้วยตัวกรอง `DOCNO='เลขที่'`** (`OnDblclkList :606-624`). ทุกแถวเริ่มต้นไม่ติ๊ก.

**C. การทำงาน**
1. **ประมวลผล** (`OnBnProcess :369-524`): ล้างตาราง → ค้น `BCApInvoice WHERE IsCancel=0 AND IsConfirm=0 AND GRBillStatus=0 AND GRIRBillStatus=1 AND IsCompleteSave=1` + ตัวกรองด้านบน, เรียง `DocDate, DocNo, ApCode` → ใส่ตาราง (ต่อ GR: `SELECT DISTINCT IRRefNo, DocDate FROM BCApInvoiceSub WHERE DocNo=GR AND TransState=0 AND IsCancel=0`). **ไม่บังคับให้ระบุตัวกรอง** (ว่างหมด = ทุก GR ค้าง).
2. **ปรับปรุง** (`OnBnOk :525-605`): ถ้าตารางว่าง → `NO_LIST` ; ถามยืนยัน "Confirm you adjust bill (GR) YES/NO ?" (ข้อความอังกฤษตายตัว) → ต่อทุกแถวที่ติ๊ก: `ΣQty` ของ GR (`BCApInvoiceSub WHERE DocNo=GR AND IsCancel=0 AND TransState=0`) เทียบ `ΣQty` ของ `BCIRSUB WHERE GRRefNo=GR AND IsCancel=0 AND TransState=0`; **ถ้า `ΣIR ≥ ΣGR` → `UPDATE BCAPINVOICE SET GRBillStatus = 1 WHERE DocNo=GR AND GRIRBillStatus=1 AND IsCompleteSave=1`**, ไม่เช่นนั้นตั้ง 0 (ไม่เปลี่ยนอะไร) → แจ้ง "ปรับปรุงทั้งหมด N รายการ" (`TOTAL_ADJUST`/`LIST`; N = จำนวนที่ถูกตั้งเป็น 1) แล้วล้างตัวกรอง/ตาราง; ถ้าไม่มีแถวถูกติ๊ก `NOT_ADJUST`. ไม่ต้องมีรหัสผ่าน/สิทธิ์เฉพาะ ไม่ตรวจงวดปิด ไม่ลง GL ไม่เขียน log.

**D. ข้อสังเกตจากต้นฉบับ (อย่าคัดลอกเป็นข้อกำหนด)**
1. **บั๊ก `OnBnOk`**: ส่วนหา `ΣQty` ของ IR เปิดด้วยตัวแปร SQL ของฝั่ง GR (`RowsetSub.Open(strSQL,…)` แทน `strSQLSub` `:562`) → `ΣIR` เท่ากับ `ΣGR` เสมอ → **ทุกแถวที่ติ๊กถูกตั้ง `GRBillStatus = 1` โดยไม่ได้ดูการอ้างอิงจริง** (ผู้ใช้ติ๊กผิดใบ = GR หายจากรายการเลือกของ 5799 ทันที และไม่มีเครื่องมือย้อนกลับในจอนี้; ต้องย้อนสถานะด้วยการแก้เอกสาร 5799/5798). **BC ต้องเทียบจำนวนจริง** (`ΣIR(GRRefNo) ≥ ΣGR`) และควรแสดงผลต่างให้ผู้ใช้เห็นก่อนยืนยัน.
2. **บั๊กตัวกรอง `OnBnProcess`**: เงื่อนไขรหัสเจ้าหนี้และแผนกกำหนดด้วย `strCondition = …` (ทับ) ไม่ใช่ `+=` → เมื่อใส่เจ้าหนี้/แผนก **เงื่อนไขวันที่/เลขที่ก่อนหน้าหายไป**; และเงื่อนไขแผนกทับเงื่อนไขเจ้าหนี้ (`:425,437`). ตัวกรองตามแผนกไม่กรองเมื่อใช้ร่วมกับตัวอื่น. **BC ต้องรวมเงื่อนไขแบบ AND ทั้งหมด**.
3. ฟังก์ชัน `ConfirmSave` (`:296-368`, สร้าง SQL อีกชุดที่ต่างจากที่ใช้จริง และบังคับให้มีเจ้าหนี้/เลขที่/วันที่ อย่างน้อย 1 ค่า) **ไม่มีผู้เรียก** — เป็นโค้ดตาย.
4. ผลของเครื่องมือกระทบเฉพาะสถานะ (`GRBillStatus`) ที่ใช้กรองรายการเลือกใน 5799 (`GRBillStatus=0`) — **ไม่คืนสต็อก/ต้นทุน/GL**.

## 2. รายงานจัดซื้อ (77 รายการในเมนูซื้อ + 33 รายการในศูนย์รายงาน)

ทุกรายงานในหัวข้อนี้ซ่อนอยู่ใน popup `disable="yes"` (§0.1) ; handler ที่ใช้จริงคือ `CMainFrame::OnXxx` ที่มี `RUNTIME_CLASS` ไม่ถูกคอมเมนต์ (atlas จับบรรทัดที่ถูกคอมเมนต์ด้วย จึงไม่ใช้ atlas ตัดสินคลาส) และตรวจสิทธิ์ด้วย `IsMenuEnable(ID_POSYS_ANALYREP_n)` ก่อนเปิด (ตัวอย่าง `MainFrm.cpp:3957,4445`). แต่ละรายการมี 2 ส่วน : หน้าต่างเงื่อนไข `*Dlg` (ช่องกรอก) และมุมมองพิมพ์ `*View` (คอลัมน์/รวมยอด) ; กดพิมพ์แล้วเรียกหน้าต่างเงื่อนไขก่อนเสมอ. กลไกกลางของรายงาน (กล่องเงื่อนไข ตัวกรอง F9 ตัวเลือก F6 คลิกทะลุ ส่งออก) อธิบายครั้งเดียวใน `10-reports.md` §3 — ไม่ซ้ำที่นี่.

**วิธีอ่านตาราง**

| คอลัมน์ | ความหมาย |
|---|---|
| เมนู / ชื่อ | menu id ตาม `menuconfig.xml` และชื่อตามเมนู (ชื่อบางตัวตกค้างจากแม่แบบ — ดู §2.3) |
| คลาส / ไฟล์ | คลาส View (`nameKey` = key ชื่อรายงานในวงเล็บ) , ไฟล์ `.cpp` ใน `rep-po\` , และ dialog เงื่อนไขใน `BCREPPO.rc` (`ไฟล์:บรรทัด`) |
| ทิศ | ค่า `m_dmOrientation` ที่ View ตั้ง : ขวาง (landscape) / ตั้ง (portrait) / — = ไม่พบการตั้งค่า |
| ตัวกรอง | ช่องของหน้าต่างเงื่อนไข ตามลำดับ ; "(จาก-ถึง)" = ช่วง 2 ช่อง ; `◉` = ตัวเลือกวิทยุ ; `☐` = ช่องติ๊ก ; "ตัวเลือก combo" = ค่าใน combo เรียงรายงาน/เลือกสถานะ ; `~` = ป้ายอนุมานจากชื่อ key (ยังไม่ verify) ; ⟨KEY⟩ = แปลไม่ได้ ; «TOKEN» = ป้ายประกอบตอนรัน |
| คอลัมน์ | หัวคอลัมน์ตามลำดับที่ View สร้าง ; ถ้ารายงานมี 2 โหมด ชุดคอลัมน์ของทุกโหมดต่อกันเรียงเป็นลำดับเดียว ; `(ตามเงื่อนไขในโค้ด)` = มีคอลัมน์ที่เพิ่ม/ซ่อนตามเงื่อนไข ต้องตามอ่านโค้ด |
| ตาราง | ตารางที่พบหลัง `FROM` / `JOIN` / `AddTableForUserDefine` ใน `*View.cpp` / `*Dlg.cpp` (ค้นจากข้อความ ไม่ใช่การไล่ SQL ทุกสาขา และ **อาจรวมตารางที่อยู่ในบรรทัดที่ถูกคอมเมนต์**) |
| รวมยอด | `มีแถวรวม` = พบตัวแปรผลรวม (`dSumOfAmount`, `dTotal`, `m_dSum` ฯลฯ) ใน View — เป็นการค้นชื่อตัวแปร ยังไม่ verify ว่าพิมพ์เป็นแถวรวมท้ายรายงานจริง ; `+รวมต่อหน้า` = View มี `OnAddPageSummery` (สรุปท้ายแต่ละหน้า ตามรูปแบบ `00-common-ui.md` §12.2) ; `—` = ไม่พบ |

**ข้อควรระวัง** : (1) ป้ายไทยในคอลัมน์ "ตัวกรอง" ที่ไม่มี `~` มาจากข้อความ rc (อาจเป็นข้อความเก่าของแม่แบบอื่นตกค้าง — เช่นป้าย "รหัสเจ้าหนี้" ซ้ำสองช่อง) — ถ้าสงสัยให้ยึด key ใน `*Dlg.cpp` ; (2) ตารางนี้สร้างด้วยการอ่านโค้ดอัตโนมัติ — คอลัมน์ที่ปรากฏซ้ำหมายถึงหลายโหมดการพิมพ์ ; (3) สูตรของแต่ละคอลัมน์และเงื่อนไข SQL ไม่ได้ถอดในตารางนี้ (ยังไม่ verify ระดับสูตร ยกเว้นระบุไว้ในหมายเหตุ §2.3).

### 2.1 รายงานในเมนูซื้อ (77 รายการ — `menuconfig.xml:27-146`)

รายงานรายวัน 14 รายการ ตามด้วยรายงานเชิงวิเคราะห์ 63 รายการ ใน 14 กลุ่ม. หัวข้อตัวหนาคือชื่อ popup ในเมนู.

**รายงานรายวัน**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5815 | รายงานการบันทึกใบขออนุมัติซื้อพร้อมรายการย่อย | `CPORepStkRequestDetailView` (20401)<br>`porepstkrequestdetailview.cpp`<br>rc `BCREPPO.rc:180` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~ที่ยืนยันแล้ว · ~ที่ยังไม่ยืนยัน · ตัวเลือก combo: ~ทั้งหมด / ⟨STK_REQUST_1⟩ / ⟨STK_REQUST_2⟩ | เลขที่เอกสาร · วันที่เอกสาร · แผนก · ผู้ขออนุมัติ · หมายเหตุ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · คำอธิบาย · จำนวนขออนุมัติ · จำนวนคงเหลือที่ยังไม่อนุมัติ | BCStkRequest, BCStkRequestSub | มีแถวรวม+รวมต่อหน้า |
| 5837 | รายงานใบขออนุมัติซื้อที่ทำสั่งซื้อหมดแล้ว | `CPORepStkRequestView` (20404)<br>`porepstkrequestview.cpp`<br>rc `BCREPPO.rc:554` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ~เฉพาะที่อนุมัติแล้ว / ~เฉพาะที่ยังไม่อนุมัติ / ⟨MAKE_APPROVED⟩ | เลขที่เอกสาร · วันที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · คำอธิบาย · หน่วยนับ · ⟨ORDER_BUY_QTY⟩ · จำนวนคงเหลือ · รหัสสินค้า · ชื่อสินค้า · คำอธิบาย · หน่วยนับ · ⟨ORDER_BUY_QTY⟩ | BCStkRequestSub, BCSTKREQUEST | มีแถวรวม+รวมต่อหน้า |
| 5817 | รายงานใบจ่ายเงินมัดจำที่คงค้าง | `CPORepApDeposit2View` (20411)<br>`porepapdeposit2view.cpp`<br>rc `BCREPPO.rc:391` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_CUTZERO⟩ | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ⟨SUM_OF_TAX⟩ · ยอดสุทธิ · ⟨BILL_INVBALANCE⟩ | BCApDeposit | มีแถวรวม+รวมต่อหน้า |
| 5841 | รายงานการซื้อสินค้าพร้อมค่าใช้จ่ายอื่นๆ(ต้นทุนแฝง) | `CPORepWeightCostView` (20420)<br>`porepWeightCostview.cpp`<br>rc `BCREPPO.rc:488` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · วันที่ครบกำหนด · รหัสเจ้าหนี้ · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ⟨SUM_OF_TAX⟩ · ยอดสุทธิ · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ⟨CHECK_PAY⟩ · ⟨EARNEST⟩ · ⟨PAY_TOTAL⟩ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวน · ต้นทุนแฝง/หน่วย · ราคาซื้อรวมต้นทุนแฝง · ~ส่วนลด · จำนวนเงิน · ยอดเงินสุทธิ · ⟨EXPAMOUNT⟩ · ยอดสุทธิ · วันที่เอกสาร · เลขที่เอกสาร · เลขที่อ้างอิง · คำอธิบาย · ⟨GOODS_QTY⟩ · จำนวนเงิน | BCApInvoiceSub, BCWeightCostSub, BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5828 | รายงานใบซื้อสินค้าที่ถึงกำหนดจ่ายเงิน | `CPORepPurchaseDueDateView` (20421)<br>`PORepPurchaseDueDateView.cpp`<br>rc `BCREPPO.rc:517` | ขวาง | ~วันที่ถึงกำหนด (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~เจ้าหนี้ (จาก-ถึง) · ~สินค้า (จาก-ถึง) | วันที่ถึงกำหนดส่ง · เลขที่เอกสาร · วันที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · รวมมูลค่าสินค้า · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · ยอดสุทธิทั้งบิล · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCAPInvoice | มีแถวรวม+รวมต่อหน้า |
| 5838 | รายงานเปรียบเทียบราคาซื้อ | `CPORepComparePriceView` (20406)<br>`PORepComparePriceView.cpp`<br>rc `BCREPPO.rc:207` | ขวาง | ~สินค้า (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · แผนก · ~จำนวนสินค้า · เลขที่ใบราคารวม · วันที่ออกเอกสาร · เลขที่ใบสืบราคา · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · จำนวนซื้อ · ~ราคา/หน่วย · ส่วนลด%,บาท · จ่ายเงินภายใน(วัน) · เงื่อนไขการขนส่ง · ส่งของภายใน(วัน) | BCByPriceCondSub | —+รวมต่อหน้า |
| 5832 | รายงานการซื้อสินค้าแยกตามประเภทการซื้อ | `CPORepApInvoiceTypeView` (20426)<br>`PORepApInvoiceTypeView.cpp`<br>rc `BCREPPO.rc:3143` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~«SALES_NO» (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~«PURCHASE_TYPE» (จาก-ถึง) · ตัวเลือก combo: ⟨CP_DOCTYPE_1⟩ / ⟨CP_DOCTYPE_2⟩ / ⟨CP_DOCTYPE_3⟩ / ⟨CP_DOCTYPE_4⟩ | ~วันที่ · เลขที่ใบซื้อ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ประเภทภาษี · มูลค่าสินค้าคิดภาษี · ~ส่วนลดท้ายบิลรวม · ~รวมรายการ(ไม่รวมภาษี) · ภาษีมูลค่าเพิ่ม · รวมเงินทั้งสิ้น · มูลค่าสุทธิทั้งบิล · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · ~เลขที่ใบสั่งซื้ออ้างอิง | BCApInvoiceSub, BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5833 | รายงานการซื้อสินค้าแจกแจงตามประเภทการซื้อ | `CPORepApInvoiceBillTypeView` (20427)<br>`PORepApInvoiceBillTypeView.cpp`<br>rc `BCREPPO.rc:3178` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่ใบซื้อ (จาก-ถึง) · ~เจ้าหนี้ (จาก-ถึง) | วันที่ออกเอกสาร · เลขที่ใบซื้อ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ซื้อสด(จ่ายเป็นเงินสด · ซื้อสด(รับอื่นๆ · ยอดเงินสดย่อย · ซื้อเชื่อ · รวม | BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5823 | รายงานใบส่งคืนที่ตัดยอดทันที | `CPORepStkRefund2View` (20417)<br>`porepstkrefund2view.cpp`<br>rc `BCREPPO.rc:2874` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ยอดใบกำกับเดิม · มูลค่าสินค้า · มูลค่าส่วนลดก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ยอดยกเว้นภาษี · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ส่วนลดก่อนภาษี · ~ยอดเดิม · ยอดรวมสินคัา · ยอดยกเว้นภาษี · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ⟨GOODS_QTY_REFUND⟩ · ~ส่วนลด · จำนวนเงิน | BCInvStkRefund, BcStkRefundSub, BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5824 | รายงานใบส่งคืนที่รับคืนเงิน | `CPORepStkRefund3View` (20418)<br>`porepstkrefund3view.cpp` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ⟨OLD_DISCOUNT⟩ · ~ยอดเดิม · ~รวมต่อรายการ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ⟨GOODS_QTY_REFUND⟩ · ~ส่วนลด · จำนวนเงิน | BCInvStkRefund, BcStkRefundSub, BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5825 | รายงานการบันทึกใบเพิ่มเจ้าหนี้พร้อมรายการย่อย | `CPORepDebitNote2DetailView` (20419)<br>`porepdebitnote2detailview.cpp`<br>rc `BCREPPO.rc:875` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · มูลค่าเพิ่มหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ⟨OLD_DISCOUNT⟩ · ~ยอดเดิม · ~รวมต่อรายการ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ⟨QTY_INCREASE⟩ · ~ส่วนลด · จำนวนเงิน | BCInvDebitNote2, BCDebitNoteSub2, BCDebitNote2 | มีแถวรวม+รวมต่อหน้า |
| 5826 | Vender Evalution Report | (เมนูไม่ผูก handler/คลาส — ยังไม่ verify) | — | — | — | — | — |
| 5733 | รายงานการตัดใบซื้อสินค้า,บริการ | `CPORepItemMasterAllocateSummaryView` (20523)<br>`porepitemmasterallocatesummaryview.cpp`<br>rc `BCREPPO.rc:2395` | ขวาง | ~การจัดสรร (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) | รหัสการจัดสรร · ⟨DETAILITEMALLOCATE⟩ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCAllocate, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5785 | รายงานบันทึกใบเสนอซื้อสินค้า | `CPORepDebitNoteSDepartView` (20925)<br>`PORepDebitNoteSDepartView.cpp`<br>rc `BCREPPO.rc:1366` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2, BCDepartment | —+รวมต่อหน้า |

**รายงานการขออนุมัติซื้อสินค้าแบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5698 | รายงานการขออนุมัติซื้อสินค้า-ตามวันที่ | `CPORepStkRequestSubDateView` (20501)<br>`porepstkrequestsubdateview.cpp`<br>rc `BCREPPO.rc:826` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่ม (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨DEPART⟩ / ⟨ITEMGROUP⟩ / ~โครงการ | ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ · รหัสแผนก · ⟨DETAILITEMDEPART⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ · รหัสกลุ่ม · ⟨DETAILITEMGROUP⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ · รหัสโครงการ · ⟨DETAILITEMPR0JECT⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ | BCStkRequestSub, BCDepartment, BCItemGroup, BCProject | มีแถวรวม+รวมต่อหน้า |
| 5729 | รายงานการขออนุมัติซื้อสินค้า-ตามแผนก | `CPORepStkRequestSubDepartmentView` (20502)<br>`porepstkrequestsubdepartmentview.cpp`<br>rc `BCREPPO.rc:790` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) | รหัสแผนก · ⟨DETAILITEMDEPART⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ | BCStkRequestSub, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5737 | รายงานการขออนุมัติซื้อสินค้า-ตามกลุ่ม | `CPORepStkRequestSubGroupView` (20503)<br>`porepstkrequestsubgroupview.cpp`<br>rc `BCREPPO.rc:748` | ขวาง | ~กลุ่ม (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) | รหัสกลุ่ม · ⟨DETAILITEMGROUP⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ | BCStkRequestSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5740 | รายงานการขออนุมัติซื้อสินค้า-ตามโครงการ | `CPORepStkRequestSubProjectView` (20504)<br>`porepstkrequestsubprojectview.cpp`<br>rc `BCREPPO.rc:711` | ขวาง | ~โครงการ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) | รหัสโครงการ · ⟨DETAILITEMPR0JECT⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ | BCStkRequestSub, BCProject | มีแถวรวม+รวมต่อหน้า |

**รายงานการขออนุมัติซื้อสินค้าแบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5742 | รายงานการขออนุมัติซื้อสินค้า-ตามวันที่ | `CPORepStkRequestDateView` (20531)<br>`PORepStkRequestDateView.cpp`<br>rc `BCREPPO.rc:3064` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่ม (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨DEPART⟩ / ⟨ITEMGROUP⟩ / ~โครงการ | ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ · รหัสแผนก · ~ชื่อแผนก · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ · รหัสกลุ่ม · ชื่อกลุ่ม · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ · รหัสโครงการ · ~ชื่อโครงการ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนคงเหลือ | BCStkRequestSub, BCDepartment, BCItemGroup, BCProject | มีแถวรวม+รวมต่อหน้า |
| 5743 | รายงานการขออนุมัติซื้อสินค้า-ตามแผนก | `CPORepStkRequestDepartView` (20532)<br>`PORepStkRequestDepartView.cpp`<br>rc `BCREPPO.rc:3031` | ตั้ง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า | BCStkRequestSub, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5744 | รายงานการขออนุมัติซื้อสินค้า-ตามกลุ่ม | `CPORepStkRequestGroupView` (20533)<br>`PORepStkRequestGroupView.cpp`<br>rc `BCREPPO.rc:2998` | ตั้ง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) | รหัสกลุ่ม · ชื่อกลุ่ม · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า | BCStkRequestSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5745 | รายงานการขออนุมัติซื้อสินค้า-ตามโครงการ | `CPORepStkRequestProjectView` (20534)<br>`PORepStkRequestProjectView.cpp`<br>rc `BCREPPO.rc:2964` | ตั้ง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) | รหัสโครงการ · ~ชื่อโครงการ · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า | BCStkRequestSub, BCProject | มีแถวรวม+รวมต่อหน้า |

**รายงานการสั่งซื้อสินค้าแบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5746 | รายงานการสั่งซื้อสินค้าแบบแจกแจง-ตามแผนก | `CPORepPurchaseOrderSubDepartmentView` (20506)<br>`poreppurchaseordersubdepartmentview.cpp`<br>rc `BCREPPO.rc:2566` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสแผนก · ⟨DETAILITEMDEPART⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCPurchaseOrderSub, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5751 | รายงานการสั่งซื้อสินค้าแบบแจกแจง-ตามกลุ่ม | `CPORepPurchaseOrderSubGroupView` (20507)<br>`poreppurchaseordersubgroupview.cpp`<br>rc `BCREPPO.rc:2517` | ขวาง | ~กลุ่ม (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสกลุ่ม · ⟨DETAILITEMGROUP⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCPurchaseOrderSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5752 | รายงานการสั่งซื้อสินค้าแบบแจกแจง-ตามโครงการ | `CPORepPurchaseOrderSubProjectView` (20508)<br>`poreppurchaseordersubprojectview.cpp`<br>rc `BCREPPO.rc:2473` | ขวาง | ~โครงการ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสโครงการ · ⟨DETAILITEMPR0JECT⟩ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCPurchaseOrderSub, BCProject | มีแถวรวม+รวมต่อหน้า |
| 5738 | รายงานการสั่งซื้อสินค้าแบบแจกแจง-ตามเจ้าหนี้ | `CPORepPurchaseOrderApCodeView` (21509)<br>`PORepPurchaseOrderApCodeView.cpp`<br>rc `BCREPPO.rc:667` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เจ้าหนี้ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCPurchaseOrderSub | มีแถวรวม+รวมต่อหน้า |

**รายงานการสั่งซื้อสินค้าแบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5747 | รายงานการสั่งซื้อสินค้าแบบสรุป-ตามวันที่ | `CPORepPurchaseOrderDateView` (20621)<br>`PORepPurchaseOrderDateView.cpp`<br>rc `BCREPPO.rc:2804` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | ~วันที่ · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCPurchaseOrderSub | —+รวมต่อหน้า |
| 5748 | รายงานการสั่งซื้อสินค้าแบบสรุป-ตามแผนก | `CPORepPurchaseOrderDepartView` (20622)<br>`PORepPurchaseOrderDepartView.cpp`<br>rc `BCREPPO.rc:2761` | ตั้ง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCPurchaseOrderSub, BCDepartment | —+รวมต่อหน้า |
| 5749 | รายงานการสั่งซื้อสินค้าแบบสรุป-ตามกลุ่ม | `CPORepPurchaseOrderGroupView` (20623)<br>`PORepPurchaseOrderGroupView.cpp`<br>rc `BCREPPO.rc:2718` | ตั้ง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCPurchaseOrderSub, BCItemGroup | —+รวมต่อหน้า |
| 5750 | รายงานการสั่งซื้อสินค้าแบบสรุป-ตามโครงการ | `CPORepPurchaseOrderProjectView` (20624)<br>`PORepPurchaseOrderProjectView.cpp`<br>rc `BCREPPO.rc:2674` | ตั้ง | ~โครงการ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสโครงการ · ~ชื่อโครงการ · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCPurchaseOrderSub, BCProject | —+รวมต่อหน้า |
| 5736 | รายงานการสั่งซื้อสินค้าแบบสรุป-ตามเจ้าหนี้ | `CPORepPurchaseOrderSApCodeView` (22509)<br>`PORepPurchaseOrderSApCodeView.cpp`<br>rc `BCREPPO.rc:623` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เจ้าหนี้ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · (ตามเงื่อนไขในโค้ด) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCPurchaseOrderSub | มีแถวรวม+รวมต่อหน้า |

**รายงานใบรับสินค้าแบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5701 | รายงานใบรับสินค้า-ตามเจ้าหนี้ | `CPORepGoodRecieveByApView` (21002)<br>`PORepGoodRecieveByApView.cpp`<br>rc `BCREPPO.rc:1864` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น | BCApInvoiceSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5702 | รายงานใบรับสินค้า-ตามแผนก | `CPORepGoodRecieveByDepartView` (21003)<br>`PORepGoodRecieveByDepartView.cpp`<br>rc `BCREPPO.rc:1759` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น | BCApInvioceSub, BCApInvoiceSub, BCItemGroup, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5703 | รายงานใบรับสินค้า-ตามกลุ่มสินค้า | `CPORepGoodRecieveByitemgroupView` (21004)<br>`PORepGoodRecieveByitemgroupView1.cpp`<br>rc `BCREPPO.rc:3474` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น | BCApInvioceSub, BCApInvoiceSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |

**รายงานใบรับสินค้าแบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5705 | รายงานสรุปใบรับสินค้า-ตามวันที่ | `CPORepGoodRecieveByDateSumView` (21011)<br>`PORepGoodRecieveByDateSumView.cpp`<br>rc `BCREPPO.rc:1831` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | ~วันที่ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · หน่วยนับ · ~จำนวนสินค้า · ~รวมทั้งสิ้น | BCApInvioceSub, BCApInvoiceSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5706 | รายงานสรุปใบรับสินค้า-ตามเจ้าหนี้ | `CPORepGoodRecieveByApSumView` (21012)<br>`PORepGoodRecieveByApSumView.cpp`<br>rc `BCREPPO.rc:1864` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · ⟨ITEMGROUP⟩ · หน่วยนับ · ~จำนวนสินค้า · ~รวมทั้งสิ้น | BCApInvioceSub, BCApInvoiceSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5707 | รายงานสรุปใบรับสินค้า-ตามแผนก | `CPORepGoodRecieveByDepartSumView` (21013)<br>`PORepGoodRecieveByDepartSumView.cpp`<br>rc `BCREPPO.rc:1759` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · หน่วยนับ · ~จำนวนสินค้า · ~รวมทั้งสิ้น | BCApInvioceSub, BCApInvoiceSub, BCItemGroup, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5708 | รายงานสรุปใบรับสินค้า-ตามกลุ่มสินค้า | `CPORepGoodRecieveByItemGroupSumView` (21014)<br>`PORepGoodRecieveByitemgroupSumView.cpp`<br>rc `BCREPPO.rc:1690` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) | รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · ~รวมทั้งสิ้น | BCApInvioceSub, BCApInvoiceSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |

**รายงานใบตั้งหนี้จากการซื้อแบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5711 | รายงานใบตั้งหนี้จากการซื้อ-ตามเจ้าหนี้ | `CPORepMakeDebtByApView` (21022)<br>`PORepMakeDebtByApView.cpp`<br>rc `BCREPPO.rc:3507` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ☐ แสดงเป็นหน่วยนับมาตรฐาน | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ราคาซื้อ · ~จำนวนสินค้า · ~ส่วนลด · จำนวนเงิน | BCStkPacking, BCApInvioce, BCIRSUB, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5712 | รายงานใบตั้งหนี้จากการซื้อ-ตามแผนก | `CPORepMakeDebtByDepartView` (21023)<br>`PORepMakeDebtByDepartView.cpp`<br>rc `BCREPPO.rc:1721` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ☐ แสดงเป็นหน่วยนับมาตรฐาน | รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ราคาซื้อ · ~จำนวนสินค้า · ~ส่วนลด · จำนวนเงิน | BCStkPacking, BCApInvioce, BCIRSUB, BCItemGroup, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5713 | รายงานใบตั้งหนี้จากการซื้อ-ตามกลุ่มสินค้า | `CPORepMakeDebtByItemGroupView` (21024)<br>`PORepMakeDebtByItemGroupView.cpp`<br>rc `BCREPPO.rc:1652` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ☐ แสดงเป็นหน่วยนับมาตรฐาน | รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ราคาซื้อ · ~จำนวนสินค้า · ~ส่วนลด · จำนวนเงิน | BCStkPacking, BCApInvioce, BCIRSUB, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5714 | รายงานใบตั้งหนี้จากการซื้อ-ตามโครงการ | `CPORepMakeDebtByProjectView` (21025)<br>`PORepMakeDebtByProjectView.cpp`<br>rc `BCREPPO.rc:1578` | ขวาง | ~โครงการ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ☐ แสดงเป็นหน่วยนับมาตรฐาน | รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ราคาซื้อ · ~จำนวนสินค้า · ~ส่วนลด · จำนวนเงิน | BCStkPacking, BCApInvioce, BCIRSUB, BCProject, BCItemGroup | มีแถวรวม+รวมต่อหน้า |

**รายงานใบตั้งหนี้จากการซื้อแบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5718 | รายงานใบตั้งหนี้จากการซื้อแบบสรุป-ตามใบตั้งหนึ้ | `CPORepMakeDebtByItemGroupSumView` (21034)<br>`PORepMakeDebtByItemGroupSumView.cpp`<br>rc `BCREPPO.rc:3445` | ขวาง | ~เลขที่ใบตั้งหนี้(IR) (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) | เลขที่ใบตั้งหนี้ · วันที่ออกเอกสาร · วันที่ครบกำหนด · เลขที่ใบกำกับ · วันที่ใบกำกับ · ยอดเงินก่อนภาษี · ภาษีมูลค่าเพิ่ม · ยอดรวมทั้งสิ้น | BCApInvioce, BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5716 | รายงานใบตั้งหนี้จากการซื้อแบบสรุป-ตามเจ้าหนี้ | `CPORepMakeDebtByApSumView` (21032)<br>`PORepMakeDebtByApSumView.cpp`<br>rc `BCREPPO.rc:3544` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่ใบตั้งหนี้(IR) (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · เลขที่ใบตั้งหนี้ · วันที่ออกเอกสาร · วันที่ครบกำหนด · เลขที่ใบกำกับ · วันที่ใบกำกับ · ยอดเงินก่อนภาษี · ภาษีมูลค่าเพิ่ม · ยอดรวมทั้งสิ้น | BCApInvoice, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5717 | รายงานใบตั้งหนี้จากการซื้อแบบสรุป-ตามแผนก | `CPORepMakeDebtByDepartSumView` (21033)<br>`PORepMakeDebtByDepartSumView.cpp`<br>rc `BCREPPO.rc:3381` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่ใบตั้งหนี้(IR) (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · เลขที่ใบตั้งหนี้ · วันที่ออกเอกสาร · วันที่ครบกำหนด · เลขที่ใบกำกับ · วันที่ใบกำกับ · ยอดเงินก่อนภาษี · ภาษีมูลค่าเพิ่ม · ยอดรวมทั้งสิ้น | BCApInvoice, BCDepartment | มีแถวรวม+รวมต่อหน้า |

**รายงานการซื้อสินค้าแบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5699 | รายงานการซื้อสินค้าแบบแจกแจง-ตามเจ้าหนี้ | `CPORepItemMasterApView` (20510)<br>`porepitemmasterapview.cpp`<br>rc `BCREPPO.rc:2347` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub | มีแถวรวม+รวมต่อหน้า |
| 5720 | รายงานการซื้อสินค้าแบบแจกแจง-ตามกลุ่มสินค้า | `CPORepItemMasterGroupView` (20511)<br>`porepitemmastergroupview.cpp`<br>rc `BCREPPO.rc:2045` | ขวาง | ~กลุ่ม (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสกลุ่ม · ⟨DETAILITEMGROUP⟩ · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCItemGroup | มีแถวรวม+รวมต่อหน้า |
| 5721 | รายงานการซื้อสินค้าแบบแจกแจง-ตามยี่ห้อสินค้า | `CPORepItemMasterBrandView` (20512)<br>`porepitemmasterbrandview.cpp`<br>rc `BCREPPO.rc:2268` | ขวาง | ~ยี่ห้อ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสยี่ห้อ · ⟨DETAILITEMBRAND⟩ · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5722 | รายงานการซื้อสินค้าแบบแจกแจง-ตามแผนก | `CPORepItemMasterDepartmentView` (20513)<br>`porepitemmasterdepartmentview.cpp`<br>rc `BCREPPO.rc:2127` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสแผนก · ⟨DETAILITEMDEPART⟩ · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCDepartment | มีแถวรวม+รวมต่อหน้า |
| 5723 | รายงานการซื้อสินค้าแบบแจกแจง-ตามโครงการ | `CPORepItemMasterProjectView` (20514)<br>`porepitemmasterprojectview.cpp`<br>rc `BCREPPO.rc:1897` | ขวาง | ~กลุ่ม (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสโครงการ · ⟨DETAILITEMPR0JECT⟩ · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCProject | มีแถวรวม+รวมต่อหน้า |
| 5724 | รายงานการซื้อสินค้าแบบแจกแจง-ตามการจัดสรร | `CPORepItemMasterAllocateView` (20515)<br>`porepitemmasterallocateview.cpp`<br>rc `BCREPPO.rc:2429` | ขวาง | ~การจัดสรร (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสการจัดสรร · ⟨DETAILITEMALLOCATE⟩ · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub, BCAllocate | มีแถวรวม+รวมต่อหน้า |
| 5725 | รายงานการซื้อสินค้าแบบแจกแจง-ตามสินค้า | `CPORepItemMasterProductItemView` (20516)<br>`porepitemmasterproductitemview.cpp`<br>rc `BCREPPO.rc:1969` | ขวาง | ~สินค้า (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า · เลขที่เอกสาร · เลขที่ใบกำกับ · ~วันที่ · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCAPInvoiceSub | มีแถวรวม+รวมต่อหน้า |

**รายงานการซื้อสินค้าแบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5726 | รายงานการซื้อสินค้าแบบสรุป-ตามวันที่ | `CPORepItemMasterDateSummaryView` (20517)<br>`PORepItemMasterDateSummaryView.cpp`<br>rc `BCREPPO.rc:2170` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) | ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5727 | รายงานการซื้อสินค้าแบบสรุป-ตามเจ้าหนี้ | `CPORepItemMasterApSummaryView` (20518)<br>`PORepItemMasterApSummaryView.cpp`<br>rc `BCREPPO.rc:2311` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) | รหัสเจ้าหนี้ · ⟨DETAILITEMAP⟩ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5728 | รายงานการซื้อสินค้าแบบสรุป-ตามกลุ่มสินค้า | `CPORepItemMasterGroupSummaryView` (20519)<br>`porepitemmastergroupsummaryview.cpp`<br>rc `BCREPPO.rc:2006` | ขวาง | ~กลุ่ม (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) | รหัสกลุ่ม · ⟨DETAILITEMGROUP⟩ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCItemGroup, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5730 | รายงานการซื้อสินค้าแบบสรุป-ตามยี่ห้อสินค้า | `CPORepItemMasterBrandSummaryView` (20520)<br>`porepitemmasterbrandsummaryview.cpp`<br>rc `BCREPPO.rc:2235` | ขวาง | ~ยี่ห้อ (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) | รหัสยี่ห้อ · ⟨DETAILITEMBRAND⟩ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCItemBrand, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5731 | รายงานการซื้อสินค้าแบบสรุป-ตามแผนก | `CPORepItemMasterDepartmentSummaryView` (20521)<br>`porepitemmasterdepartmentsummaryview.cpp`<br>rc `BCREPPO.rc:2094` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) | รหัสแผนก · ⟨DETAILITEMDEPART⟩ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCDepartment, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5734 | รายงานการซื้อสินค้าแบบสรุป-ตามสินค้า | `CPORepItemMasterItemSummaryView` (20524)<br>`porepitemmasteritemsummaryview.cpp`<br>rc `BCREPPO.rc:1941` | ขวาง | ~สินค้า (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~หน่วยนับ (จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~จำนวนสินค้า · จำนวนเงิน · ~ส่วนลด · ยอดสุทธิ | BCAPInvoiceSub, BCItemUnit | มีแถวรวม+รวมต่อหน้า |
| 5735 | รายงานการซื้อสินค้าแบบสรุป-ตามเลขที่เอกสาร | `CPORepPrecisDocNoView` (20525)<br>`PORepPrecisDocNoView.cpp`<br>rc `BCREPPO.rc:3573` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่ใบซื้อ (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) | เลขที่ใบซื้อ · พิมพ์จากวันที่ · รหัสเจ้าหนี้ · ชื่อที่1 · ยอดก่อนภาษี · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ภาษีหัก ณ ที่จ่าย · รวมเงินทั้งสิ้น · ยอดยกเว้นภาษี · มูลค่าสุทธิทั้งบิล | BCArInvoiceSub, BCApInvoice | มีแถวรวม+รวมต่อหน้า |

**รายงานการส่งคืนสินค้า/ลดหนี้แบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5756 | รายงานการส่งคืนสินค้า/ลดหนี้แบบแจกแจง-ตามยี่ห้อสินค้า | `CPORepStkRefundBrandView` (20814)<br>`PORepStkRefundBrandView.cpp`<br>rc `BCREPPO.rc:942` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสยี่ห้อ · ~ชื่อยี่ห้อ · เลขที่ใบส่งคืน/ลดหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ราคาซื้อ / หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub, BCItemBrand | —+รวมต่อหน้า |
| 5757 | รายงานการส่งคืนสินค้า/ลดหนี้แบบแจกแจง-ตามแผนก | `CPORepStkRefundDepartView` (20815)<br>`PORepStkRefundDepartView.cpp`<br>rc `BCREPPO.rc:1198` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · เลขที่ใบส่งคืน/ลดหนี้ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ราคาซื้อ / หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub, BCDepartment | —+รวมต่อหน้า |
| 5758 | รายงานการส่งคืนสินค้า/ลดหนี้แบบแจกแจง-ตามโครงการ | `CPORepStkRefundProjectView` (20816)<br>`PORepStkRefundProjectView.cpp`<br>rc `BCREPPO.rc:1073` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสโครงการ · ~ชื่อโครงการ · เลขที่ใบส่งคืน/ลดหนี้ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ราคาซื้อ / หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub, BCProject | —+รวมต่อหน้า |
| 5759 | รายงานการส่งคืนสินค้า/ลดหนี้แบบแจกแจง-ตามการจัดสรร | `CPORepStkRefundAllocateView` (20817)<br>`PORepStkRefundAllocateView.cpp`<br>rc `BCREPPO.rc:1029` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~การจัดสรร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสการจัดสรร · ชื่อการจัดสรร · เลขที่ใบส่งคืน/ลดหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ราคาซื้อ / หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub, BCAllocate | —+รวมต่อหน้า |
| 5760 | รายงานการส่งคืนสินค้า/ลดหนี้แบบแจกแจง-ตามสินค้า | `CPORepStkRefundItemView` (20818)<br>`PORepStkRefundItemView.cpp`<br>rc `BCREPPO.rc:1117` | ขวาง | ~สินค้า (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · เลขที่ใบส่งคืน/ลดหนี้ · เลขที่ใบกำกับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ราคาซื้อ / หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub | —+รวมต่อหน้า |

**รายงานการส่งคืนสินค้า/ลดหนี้แบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5761 | รายงานการส่งคืนสินค้า/ลดหนี้แบบสรุป-ตามวันที่ | `CPORepStkRefundSDateView` (20821)<br>`PORepStkRefundSDateView.cpp`<br>rc `BCREPPO.rc:1409` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | ~วันที่ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub | —+รวมต่อหน้า |
| 5762 | รายงานการส่งคืนสินค้า/ลดหนี้แบบสรุป-ตามเจ้าหนี้ | `CPORepStkRefundSApCodeView` (20822)<br>`PORepStkRefundSApCodeView.cpp`<br>rc `BCREPPO.rc:1490` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~การจัดสรร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub | —+รวมต่อหน้า |
| 5768 | รายงานการส่งคืนสินค้า/ลดหนี้แบบสรุป-ตามสินค้า | `CPORepStkRefundSItemView` (20828)<br>`PORepStkRefundSItemView.cpp`<br>rc `BCREPPO.rc:1285` | ตั้ง | ~สินค้า (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefundSub | —+รวมต่อหน้า |

**รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5772 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามวันที่ | `CPORepDebitNoteDateView` (20911)<br>`PORepDebitNoteDateView.cpp`<br>rc `BCREPPO.rc:904` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | ~วันที่ · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2 | —+รวมต่อหน้า |
| 5773 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามเจ้าหนี้ | `CPORepDebitNoteApCodeView` (20912)<br>`PORepDebitNoteApCodeView.cpp`<br>rc `BCREPPO.rc:985` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~การจัดสรร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2 | —+รวมต่อหน้า |
| 5774 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามกลุ่มสินค้า | `CPORepDebitNoteGroupView` (20913)<br>`PORepDebitNoteGroupView.cpp`<br>rc `BCREPPO.rc:1155` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2, BCItemGroup | —+รวมต่อหน้า |
| 5775 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามยี่ห้อสินค้า | `CPORepDebitNoteBrandView` (20914)<br>`PORepDebitNoteBrandView.cpp`<br>rc `BCREPPO.rc:942` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสยี่ห้อ · ~ชื่อยี่ห้อ · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2, BCItemBrand | —+รวมต่อหน้า |
| 5776 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามแผนก | `CPORepDebitNoteDepartView` (20915)<br>`PORepDebitNoteDepartView.cpp`<br>rc `BCREPPO.rc:1198` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสแผนก · ~ชื่อแผนก · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2, BCDepartment | —+รวมต่อหน้า |
| 5778 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามการจัดสรร | `CPORepDebitNoteAllocateView` (20917)<br>`PORepDebitNoteAllocateView.cpp`<br>rc `BCREPPO.rc:1029` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~การจัดสรร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสการจัดสรร · ชื่อการจัดสรร · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2, BCAllocate | —+รวมต่อหน้า |
| 5779 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบแจกแจง-ตามสินค้า | `CPORepDebitNoteItemView` (20918)<br>`PORepDebitNoteItemView.cpp`<br>rc `BCREPPO.rc:1117` | ขวาง | ~สินค้า (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสสินค้า · เลขที่ใบเพิ่มหนี้ · เลขที่ใบกำกับ · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ราคาซื้อ/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2 | —+รวมต่อหน้า |

**รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบสรุป**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5782 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบสรุป-ตามเจ้าหนี้ | `CPORepDebitNoteSApCodeView` (20922)<br>`PORepDebitNoteSApCodeView.cpp`<br>rc `BCREPPO.rc:1490` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~การจัดสรร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2 | —+รวมต่อหน้า |
| 5783 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบสรุป-ตามกลุ่มสินค้า | `CPORepDebitNoteSGroupView` (20923)<br>`PORepDebitNoteSGroupView.cpp`<br>rc `BCREPPO.rc:1323` | ขวาง | ~แผนก (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2, BCItemGroup | —+รวมต่อหน้า |
| 5788 | รายงานใบเพิ่มสินค้า(เจ้าหนี้)แบบสรุป-ตามสินค้า | `CPORepDebitNoteSItemView` (20928)<br>`PORepDebitNoteSItemView.cpp`<br>rc `BCREPPO.rc:1285` | ขวาง | ~สินค้า (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · จำนวนเพิ่มหนี้ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCDebitNoteSub2 | —+รวมต่อหน้า |

### 2.2 รายงานจัดซื้อในศูนย์รายงาน (Report Center) — 33 รายการ (`menuconfig.xml:620-670`)

popup "จัดซื้อ" ใต้เมนู "รายงาน" มีรายงานของตัวเอง 2 กลุ่ม : "รายงานรายวัน" 25 รายการ (ใบสั่งซื้อ · รับสินค้า · ตั้งหนี้ · ซื้อ · ส่งคืน/ลดหนี้ · เพิ่มหนี้ · ใบเสนอซื้อ · เงินมัดจำ · เงินจ่ายล่วงหน้า) และ "รายงานวิเคราะห์" 8 รายการ (`menuconfig.xml:660-669`). ใช้ menu id และคลาส View คนละชุดกับ §2.1 (ตรวจแล้วไม่มีคลาสซ้ำกัน) ; ทิศกระดาษเป็นขวางทั้งหมด ยกเว้น 5818 (ตั้ง) ; หลายตัวเป็นคลาสรุ่น `Chm*`. ตารางใช้รูปแบบเดียวกับ §2.1.

**รายงานรายวัน**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5831 | รายงานใบสั่งซื้อสินค้า | `CChmRepPOPurchaseOrderView` (25431)<br>`ChmRepPOPurchaseOrderView.cpp`<br>rc `BCREPPO.rc:4741` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · วันหมดอายุ · ⟨SUM_OF_ITMEAMOUNT⟩ · ยอดส่วนลด · ~มูลค่าก่อนภาษี · ~ภาษี · ยอดสุทธิ | BCPurchaseOrder | มีแถวรวม+รวมต่อหน้า |
| 5839 | รายงานใบสั่งซื้อสินค้าพร้อมรายการย่อย | `CChmRepPOPurchaseOrderDetailView` (20407)<br>`ChmRepPOPurchaseOrderDetailView.cpp`<br>rc `BCREPPO.rc:4770` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~เลขที่ใบเสนอซื้อ (จาก-ถึง) · ~วันครบกำหนดใบเสนอราคา (จาก-ถึง) · ~วันที่ถึงกำหนด (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ⟨BY_PO_EXPIRE⟩ / ⟨BY_PO_DUE⟩ · ☐ แสดงเลขที่ใบอนุมัติเสนอซื้อ | วันที่เอกสาร · เลขที่เอกสาร · ชื่อเจ้าหนี้ · หมายเหตุ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QUANTITY_ITEM⟩ · ~ส่วนลด · ⟨SUM_ITEMAMOUNT⟩ · ยอดรวมภาษี · วันที่กำนหนดรับ · เลขที่เอกสาร · วันที่เอกสาร · วันที่ครบกำหนด · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · รวมเงินทั้งสิ้น · รหัสสินค้า · ชื่อสินค้า · เลขที่ใบอนุมัติเสนอซื้อ · เลขที่ใบเสนอซื้อ · หน่วยนับ · ⟨QUANTITY_ITEM⟩ · ~ราคา/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · เลขที่เอกสาร · วันที่เอกสาร · วันที่ครบกำหนด · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · รวมเงินทั้งสิ้น · รหัสสินค้า · ชื่อสินค้า · คำอธิบาย · เลขที่ใบเสนอซื้อ · หน่วยนับ · ⟨QUANTITY_ITEM⟩ · ~ราคา/หน่วย · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · วันหมดอายุ · เลขที่ใบสั่งซื้อ · วันที่ออกเอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · แผนก · ~ส่วนลดท้ายบิลรวม · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · ⟨LAST_DUEDATE⟩ · เลขที่เอกสาร · วันที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดรวมสินค้า · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · ยอดสุทธิ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~รวม · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCPurchaseOrder, BCPurchaseOrderSub | มีแถวรวม+รวมต่อหน้า |
| 5830 | รายงานการตัดใบสั่งซื้อสินค้า | `CPORepCutPurchaseOrderView` (20423)<br>`PORepCutPurchaseOrderView.cpp`<br>rc `BCREPPO.rc:3247` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | เลขที่ใบสั่งซื้อ · พิมพ์จากวันที่ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดสุทธิ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · เลขที่ใบซื้อ/รับสินค้า · พิมพ์จากวันที่ · ประเภทรายวัน · จำนวนที่ทำซื้อ · จำนวนคงเหลือ · จำนวนก่อนทำซื้อ · จำนวนที่ทำซื้อ · จำนวนคงเหลือ | BCPurchaseOrderSub, BCApInvoiceSub, BCPOVoidSub, BCPurchaseOrder | มีแถวรวม+รวมต่อหน้า |
| 5840 | รายงานสถานะใบสั่งซื้อสินค้า | `CPORepPurchaseOrderView` (20408)<br>`poreppurchaseorderview.cpp`<br>rc `BCREPPO.rc:285` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ⟨ONLY_BUY⟩ / ⟨SPECIAL_UNSHIPMENT_ITEM⟩ | วันที่เอกสาร · เลขที่เอกสาร · เจ้าหนี้ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ⟨ORDER_BUY_QTY⟩ · ~ส่วนลด · จำนวนเงิน · จำนวนคงเหลือ | BCPurchaseOrderSub | มีแถวรวม+รวมต่อหน้า |
| 5829 | รายงานการยกเลิกใบสั่งซื้อสินค้าพร้อมรายการย่อย | `CChmRepPOCancelPoView` (20432)<br>`ChmRepPOCancelPoView.cpp`<br>rc `BCREPPO.rc:4824` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่ใบสั่งซื้อ (จาก-ถึง) · ~เลขที่ใบสั่งซื้อ (จาก-ถึง) · ~สินค้า (จาก-ถึง) | ⟨CANCEL_DOC_DATE⟩ · ⟨CANCEL_DOC_NO⟩ · ⟨PO_DOC_DATE⟩ · ⟨PURCHASE_DOCNO⟩ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ยอดรวมเงินสุทธิ · ⟨QTY_PURCHASE⟩ · ⟨QTY_CANCEL⟩ | BCPurchaseOrder, BCPoVoidSub | —+รวมต่อหน้า |
| 5709 | รายงานรับสินค้าจากการซื้อ | `CChmRepGoodsReceiveView` (25433)<br>`ChmRepGoodsReceiveView.cpp`<br>rc `BCREPPO.rc:4860` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · วันที่ถึงกำหนด · ⟨SUM_OF_ITMEAMOUNT⟩ · ยอดส่วนลด · ~มูลค่าก่อนภาษี · ~ภาษี · ยอดสุทธิ | BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5704 | รายงานรับสินค้าจากการซื้อพร้อมรายการย่อย | `CChmRepGoodsReceiveDetailView` (20434)<br>`ChmRepGoodsReceiveDetailView.cpp`<br>rc `BCREPPO.rc:4889` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~เลขที่ใบสั่งซื้อ (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · วันที่ถึงกำหนด · ⟨SUM_OF_ITMEAMOUNT⟩ · ยอดส่วนลด · ~มูลค่าก่อนภาษี · ~ภาษี · ยอดสุทธิ · ⟨PRODUCTCODE⟩ · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · มูลค่าส่วนลด · จำนวนเงิน · ⟨DOCNO_PO_IR⟩ | bcshelf, BCApInvoiceSub, bcapinvoice | มีแถวรวม+รวมต่อหน้า |
| 5820 | รายงานการตัดใบซื้อสินค้า | `CChmRepGrBalanceView` (20414)<br>`ChmRepGrBalanceView.cpp`<br>rc `BCREPPO.rc:4924` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ⟨BILL_INVBALANCE⟩ / ⟨BILLBALANCE_CUTZERO⟩ / ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_PAYZERO⟩ | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · หมายเหตุ · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ⟨SUM_OF_TAX⟩ · ยอดสุทธิ · ⟨BILL_INVBALANCE⟩ | bcapinvoicesub, BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5715 | รายงานตั้งหนี้จากการซื้อ | `CPORepMakeDebtByDateSumView` (25435)<br>`PORepMakeDebtByDateSumView.cpp`<br>rc `BCREPPO.rc:3352` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่ใบตั้งหนี้(IR) (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) | วันที่ออกเอกสาร · เลขที่ใบตั้งหนี้ · วันที่ครบกำหนด · เลขที่ใบกำกับ · วันที่ใบกำกับ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดเงินก่อนภาษี · ภาษีมูลค่าเพิ่ม · ยอดรวมทั้งสิ้น | BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5719 | รายงานตั้งหนี้จากการซื้อพร้อมรายการย่อย | `CChmRepIrDetailView` (20435)<br>`ChmRepIrDetailView.cpp`<br>rc `BCREPPO.rc:4957` | ขวาง | ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่ใบตั้งหนี้(IR) (จาก-ถึง) · ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~เลขที่ใบสั่งซื้อ (จาก-ถึง) | วันที่ออกเอกสาร · เลขที่ใบตั้งหนี้ · วันที่ครบกำหนด · เลขที่ใบกำกับ · วันที่ใบกำกับ · ⟨SUM_MONEY_TAX⟩ · ภาษีมูลค่าเพิ่ม · ยอดรวมเงินสุทธิ · ⟨PRODUCTCODE⟩ · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · มูลค่าส่วนลด · จำนวนเงิน · เลขที่ใบรับสินค้า | bcshelf, BCIRSub, BCAPINVOICE | มีแถวรวม+รวมต่อหน้า |
| 5766 | รายงานการส่งคืนสินค้า | `CChmRepPoStkRefundSumView` (25436)<br>`ChmRepPoStkRefundSumView.cpp`<br>rc `BCREPPO.rc:4992` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง | BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5821 | รายงานการส่งคืนสินค้าพร้อมรายการย่อย | `CPORepStkRefundDetailView` (20415)<br>`porepstkrefunddetailview.cpp`<br>rc `BCREPPO.rc:3111` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ☐ แสดง Serial | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ~ยอดเดิม · ~รวมต่อรายการ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ⟨GOODS_QTY_REFUND⟩ · ~ส่วนลด · จำนวนเงิน · รหัสควบคุมทะเบียนสินค้า · หมายเลขทะเบียน · คลัง · ที่เก็บ · หมายเหตุ · สถานะ · รหัสควบคุมทะเบียนสินค้า · หมายเลขทะเบียน · หมายเลขตัวถัง · คลัง · ที่เก็บ · หมายเลขเครื่อง · หมายเหตุ · สถานะ · วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ~ยอดเดิม · ~รวมต่อรายการ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ราคาเดิม · จำนวนเดิม · ราคาที่ผิด WRONG_PRICE เปลี่ยนเป็น ราคาที่ลดหนี้ · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ส่วนลด · จำนวนเงิน | BCInvStkRefund, BCStkRefundSub, BCSerialMaster, BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5822 | รายงานการลดหนี้ | `CChmRepStkRefundView` (25432)<br>`ChmRepStkRefundView.cpp`<br>rc `BCREPPO.rc:2842` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_CUTZERO⟩ | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง | BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5732 | รายงานซื้อสินค้า,บริการ | `CChmRepPOPurchaseSumView` (25437)<br>`ChmRepPOPurchaseSumView.cpp`<br>rc `BCREPPO.rc:5022` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · รวมเงินทั้งสิ้น · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · ยอดเงินมัดจำ · รวมจ่าย | BCPayment, BCApInvoice | มีแถวรวม+รวมต่อหน้า |
| 5819 | รายงานซื้อสินค้า,บริการพร้อมรายการย่อย | `CPORepApInvoiceDetailView` (20413)<br>`porepapinvoicedetailview.cpp`<br>rc `BCREPPO.rc:424` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ☐ แสดง Serial | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · รวมเงินทั้งสิ้น · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · ยอดเงินมัดจำ · รวมจ่าย · รหัสสินค้า · ชื่อสินค้า · คำอธิบาย · หน่วยนับ · ~ราคา/หน่วย · ~จำนวน · ⟨SUM_ITEMAMOUNT⟩ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · ~เลขที่ใบสั่งซื้ออ้างอิง · ⟨CTRL_NO⟩ · หมายเลขโทรศัพท์ · คลัง · ที่เก็บ · หมายเหตุ · สถานะ · รหัสควบคุมทะเบียนสินค้า · หมายเลขทะเบียน · หมายเลขตัวถัง · คลัง · ที่เก็บ · หมายเลขเครื่อง · หมายเหตุ · สถานะ · วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ~ส่วนลดท้ายบิลรวม · ยอดภาษี · ยอดยกเว้นภาษี · รวมเงินทั้งสิ้น · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · ยอดเงินมัดจำ · รวมจ่าย · รหัสสินค้า · ชื่อสินค้า · คำอธิบาย · หน่วยนับ · ~ราคา/หน่วย · ~จำนวน · ⟨SUM_ITEMAMOUNT⟩ · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · ~เลขที่ใบสั่งซื้ออ้างอิง | BCPayment, BCApInvoiceSub, BCSerialMaster, BCAPInvoice | มีแถวรวม+รวมต่อหน้า |
| 5767 | รายงานบันทึกส่งคืนสินค้า/ลดหนี้ | `CChmRepCreditNoteSumView` (20443)<br>`ChmRepCreditNoteSumView.cpp`<br>rc `BCREPPO.rc:5083` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_CUTZERO⟩ | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง | BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5765 | รายงานบันทึกส่งคืนสินค้า/ลดหนี้พร้อมรายการย่อย | `CChmRepCreditNoteDetailView` (20444)<br>`ChmRepCreditNoteDetailView.cpp`<br>rc `BCREPPO.rc:5116` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_CUTZERO⟩ · ☐ แสดง Serial | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ~ยอดเดิม · ~รวมต่อรายการ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ⟨GOODS_QTY_REFUND⟩ · ~ส่วนลด · จำนวนเงิน · รหัสควบคุมทะเบียนสินค้า · หมายเลขทะเบียน · คลัง · ที่เก็บ · หมายเหตุ · สถานะ · รหัสควบคุมทะเบียนสินค้า · หมายเลขทะเบียน · หมายเลขตัวถัง · คลัง · ที่เก็บ · หมายเลขเครื่อง · หมายเหตุ · สถานะ · วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · ~เลขประจำตัวผู้เสียภาษี · ~ยอดเดิม · ~รวมต่อรายการ · ยอดที่ถูกต้อง · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ราคาเดิม · จำนวนเดิม · ราคาที่ผิด WRONG_PRICE เปลี่ยนเป็น ราคาที่ลดหนี้ · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ส่วนลด · จำนวนเงิน | BCInvStkRefund, BCStkRefundSub, BCSerialMaster, BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5786 | รายงานบันทึกใบเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้) | `CChmRepDebitNoteSumView` (20445)<br>`ChmRepDebitNoteSumView.cpp`<br>rc `BCREPPO.rc:5151` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · มูลค่าเพิ่มหนี้ · ยอดภาษี · รวมเงินทั้งสิ้น · ยอดที่ถูกต้อง | BCDebitNote2 | มีแถวรวม+รวมต่อหน้า |
| 5777 | รายงานบันทึกใบเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)พร้อมรายการย่อย | `CChmRepDebitNoteDetailView` (20446)<br>`ChmRepDebitNoteDetailView.cpp`<br>rc `BCREPPO.rc:5180` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · มูลค่าเพิ่มหนี้ · ยอดภาษี · รวมเงินทั้งสิ้น · ยอดที่ถูกต้อง · เลขที่ใบส่งของ · รหัสสินค้า · ⟨ITEMNAME1⟩ · คลัง · ที่เก็บ · หน่วยนับ · ยอดภาษี · ~ราคา(2) · ~รวมทั้งสิ้น | BCDebitNote2 | มีแถวรวม+รวมต่อหน้า |
| 5787 | รายงานบันทึกใบเสนอซื้อสินค้าพร้อมรายการย่อย | `CChmRepStkRequestDetailView` (20448)<br>`ChmRepStkRequestDetailView.cpp`<br>rc `BCREPPO.rc:5215` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~ที่ยืนยันแล้ว · ~ที่ยังไม่ยืนยัน · ตัวเลือก combo: ~ทั้งหมด / ⟨STK_REQUST_1⟩ / ⟨STK_REQUST_2⟩ | เลขที่เอกสาร · วันที่เอกสาร · แผนก · ผู้ขออนุมัติ · หมายเหตุ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · คำอธิบาย · จำนวนขออนุมัติ · จำนวนคงเหลือที่ยังไม่อนุมัติ | BCStkRequest, BCStkRequestSub | มีแถวรวม+รวมต่อหน้า |
| 5816 | รายงานการบันทึกใบจ่ายเงินมัดจำ | `CPORepApDepositView` (20410)<br>`porepapdepositview.cpp`<br>rc `BCREPPO.rc:360` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ⟨BILLBALANCE_ZERO⟩ | วันที่เอกสาร · เลขที่เอกสาร · ~เลขประจำตัวผู้เสียภาษี · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~มูลค่าก่อนภาษี · ยอดภาษี · ยอดรวมภาษี · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · รวมจ่าย | BCApDeposit | มีแถวรวม+รวมต่อหน้า |
| 5818 | รายงานการตัดใบจ่ายเงินมัดจำ | `CPORepCutApDepositView` (20412)<br>`PORepCutApDepositView.cpp`<br>rc `BCREPPO.rc:3210` | ตั้ง | ~วันที่เอกสาร (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ~เลขที่ใบซื้อ (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) | เลขที่ใบจ่ายเงินมัดจำ · วันที่จ่าย · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดสุทธิ · มูลค่าก่อนตัด · เลขที่ใบซื้อ · พิมพ์จากวันที่ · คำอธิบาย · ยอดตัดจ่าย · ยอดคงเหลือ | BCAPDepositUse, BCInvStkRefund, BCPaymentSub, BCAPDeposit | มีแถวรวม+รวมต่อหน้า |
| 5763 | รายงานบันทึกใบจ่ายเงินล่วงหน้า | `CChmRepApDepositSpecialView` (20451)<br>`ChmRepApDepositSpecialView.cpp`<br>rc `BCREPPO.rc:5242` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) · ตัวเลือก combo: ~ทั้งหมด / ⟨BILLBALANCE_ZERO⟩ | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · รวมจ่าย | BCAPDepositSpecial | มีแถวรวม+รวมต่อหน้า |
| 5764 | รายงานการตัดใบจ่ายเงินล่วงหน้า | `CChmRepApDepositSpecialCutView` (20452)<br>`ChmRepApDepositSpecialCutView.cpp`<br>rc `BCREPPO.rc:5273` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · รวมจ่าย · ⟨BEFORE_AMOUNT_CUT⟩ · ⟨OFFSET_DATE⟩ · เลขที่เอกสาร · ⟨CUT_PAY_AMOUNT⟩ | BCAPDepositSpecial | มีแถวรวม+รวมต่อหน้า |
| 5784 | รายงานบันทึกรับคืนเงินจ่ายล่วงหน้า | `CChmRepApDepositSpecialRetView` (20453)<br>`ChmRepApDepositSpecialRetView.cpp`<br>rc `BCREPPO.rc:5302` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~เลขที่เอกสาร (จาก-ถึง) | วันที่เอกสาร · เลขที่เอกสาร · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ยอดสุทธิ · ภาษีหัก ณ ที่จ่าย · ยอดเงินสด · ยอดเงินสดย่อย · ยอดเงินโอน · ยอดเช็คจ่าย · รวมจ่าย | BCAPReturnDepSpecial | มีแถวรวม+รวมต่อหน้า |

**รายงานวิเคราะห์**

| เมนู | ชื่อ (Champ) | คลาส / ไฟล์ | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง | รวมยอด |
|---|---|---|---|---|---|---|---|
| 5741 | รายงานใบสั่งซื้อสินค้า | `CChmRepAnaPurchaseOrderView` (20431)<br>`ChmRepAnaPurchaseOrderView.cpp`<br>rc `BCREPPO.rc:5331` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · รหัสสินค้า · ชื่อสินค้า · หมายเหตุ · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · ยอดรวม · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCPurchaseOrderSub, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5700 | รายงานรับสินค้าจากการซื้อ | `CChmRepAnaGoodsReceiveView` (20433)<br>`ChmRepAnaGoodsReceiveView.cpp`<br>rc `BCREPPO.rc:5397` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ~จำนวน (จาก-ถึง) · ~จำนวนเงิน (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · ~ชื่อกลุ่มสินค้า · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~รวมทั้งสิ้น | BCApInvoiceSub, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5710 | รายงานตั้งหนี้จากการซื้อ | `CChmRepAnaInvoiceReceiveView` (28435)<br>`ChmRepAnaInvoiceReceiveView.cpp`<br>rc `BCREPPO.rc:5463` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ · ☐ แสดงเป็นหน่วยนับมาตรฐาน | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~ส่วนลด · จำนวนเงิน · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~ส่วนลด · ~รวมทั้งสิ้น · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~ส่วนลด · ~รวมทั้งสิ้น · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~ส่วนลด · ~รวมทั้งสิ้น · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~ส่วนลด · ~รวมทั้งสิ้น · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · ~จำนวนสินค้า · ~ราคา/หน่วย · ~ส่วนลด · ~รวมทั้งสิ้น | BCApInvoice, BCIRSub, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5754 | รายงานการส่งคืนสินค้า | `CChmRepAnaStkReturnView` (20436)<br>`ChmRepAnaStkReturnView.cpp`<br>rc `BCREPPO.rc:5519` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ราคา(2) · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ราคา(2) · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ราคา(2) · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ราคา(2) · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ราคา(2) · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · ~เลขที่ใบส่งคืน/ลดหนี้ · ~ราคา(2) · ~ส่วนลด · จำนวนเงิน · ยอดสุทธิ | BCStkRefund, BCStkRefundSub, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5755 | รายงานการลดหนี้ | `CChmRepAnaCreditNoteView` (21416)<br>`ChmRepAnaCreditNoteView.cpp` | ขวาง | (ไม่พบ) | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~ยอดเดิม · รวมมูลค่าสินค้า · มูลค่าส่วนลด · ~มูลค่าก่อนภาษี · ยอดภาษี · รวมเงินทั้งสิ้น · ~รวมรายการ(ไม่รวมภาษี) · ยอดผลต่าง/ลดหนี้ · ยอดที่ถูกต้อง | BCStkRefund | มีแถวรวม+รวมต่อหน้า |
| 5769 | รายงานซื้อสินค้า,บริการ | `CChmRepAnaPurchaseView` (20437)<br>`ChmRepAnaPurchaseView.cpp`<br>rc `BCREPPO.rc:5625` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · เลขที่ใบสั่งซื้อ · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~จำนวนสินค้า · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ | BCApInvoiceSub, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5753 | รายงานส่งคืนสินค้า/ลดหนี้ | `CChmRepAnaStkRefundView` (20439)<br>`ChmRepAnaStkRefundView.cpp`<br>rc `BCREPPO.rc:5678` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~เลขที่ใบส่งคืน/ลดหนี้ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~เลขที่ใบส่งคืน/ลดหนี้ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~เลขที่ใบส่งคืน/ลดหนี้ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~เลขที่ใบส่งคืน/ลดหนี้ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~เลขที่ใบส่งคืน/ลดหนี้ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ~เลขที่ใบส่งคืน/ลดหนี้ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ | BCStkRefundSub, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |
| 5781 | รายงานใบเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้) | `CChmRepAnaDebitNoteView` (20441)<br>`hmRepAnaDebitNoteView.cpp`<br>rc `BCREPPO.rc:5731` | ขวาง | ~วันที่เอกสาร (จาก-ถึง) · ~สินค้า (จาก-ถึง) · ~รหัสเจ้าหนี้ (จาก-ถึง) · ~แผนก (จาก-ถึง) · ~กลุ่มสินค้า (จาก-ถึง) · ~โครงการ (จาก-ถึง) · ~ยี่ห้อ (จาก-ถึง) · ตัวเลือก combo: ⟨STR_DATE⟩ / ⟨APCODE⟩ / ⟨DEPART_CODE⟩ / ⟨GROUPCODE_ITEM⟩ / ⟨PROJECT_CODE⟩ / ⟨ITEMBRANCODE⟩ | ~วันที่ · (ตามเงื่อนไขในโค้ด) · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QTY_INCDEBT⟩ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสเจ้าหนี้ · ชื่อเจ้าหนี้ · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QTY_INCDEBT⟩ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสแผนก · ~ชื่อแผนก · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QTY_INCDEBT⟩ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสกลุ่ม · ~ชื่อกลุ่มสินค้า · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QTY_INCDEBT⟩ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสโครงการ · ~ชื่อโครงการ · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QTY_INCDEBT⟩ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ · รหัสยี่ห้อ · ~ชื่อยี่ห้อ · ~วันที่ · เลขที่เอกสาร · ~เลขที่ใบสั่งซื้อ(อ้างจาก IR) · รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ~ราคา/หน่วย · ⟨QTY_INCDEBT⟩ · มูลค่าสินค้า · ~ส่วนลด · ยอดสุทธิ | BCDebitNoteSub2, BCItemGroup, BCDepartment, BCProject, BCItemBrand | มีแถวรวม+รวมต่อหน้า |

### 2.3 หมายเหตุเฉพาะรายงานจัดซื้อที่ BC ต้องรู้

- **5826 "Vender Evalution Report"** (เมนู `ID_POSYS_TRANSREP_2_0`) : เมนูมีอยู่ใน `menuconfig.xml` แต่ **ไม่พบ handler `OnXxx` และไม่พบคลาสรายงานที่ผูก** (ตัวสแกน handler ได้ "?") — สะกดชื่อผิดทั้งสองคำ ("Vender", "Evalution") ; สรุปว่าเป็นเมนูที่ยังไม่เคยทำ. BC ไม่ต้องทำ ถ้าลุงจืดไม่สั่ง (ยังไม่ verify ว่า Champ ที่รันจริงซ่อน/แสดงเมนูนี้).
- **ชื่อเมนูไม่ตรงกับรายงานที่เปิดจริง (แม่แบบคัดลอก)** : (ก) **5785 "รายงานบันทึกใบเสนอซื้อสินค้า"** เปิด `CPORepDebitNoteSDepartView` = สรุปใบเพิ่มหนี้(เจ้าหนี้) ตามแผนก (`MainFrm.cpp:4443-4448` ความเห็นในโค้ดเขียนว่า "รายงานสรุปใบเพิ่มหนี้(เจ้าหนี้) - ตามแผนก") คอลัมน์ "จำนวนเพิ่มหนี้" และ dialog `IDD_PO_REP_STKREFUND_SDEPART` (`BCREPPO.rc:1366`) ; รายงานใบเสนอซื้อจริงคือ **5815** (รายวัน §2.1) และ **5787** (ศูนย์รายงาน §2.2). (ข) **5733 "รายงานการตัดใบซื้อสินค้า,บริการ"** เปิด `CPORepItemMasterAllocateSummaryView` (`MainFrm.cpp:3957-3961`, สิทธิ์ `ID_POSYS_ANALYREP_23`) ซึ่งชื่อคลาสบอกว่าเป็นรายงานสรุปตามการจัดสรร — ยังไม่verify หน้าจอจริงว่าตรงกับชื่อเมนูหรือไม่. BC ต้องตั้งชื่อรายงานตามเนื้อหาที่แสดงจริง ไม่คัดลอกชื่อเมนูเดิม.
- **ป้ายตัวกรองซ้ำ** : เช่น 5829 (ศูนย์รายงาน) มีช่อง "เลขที่ใบสั่งซื้อ (จาก-ถึง)" ปรากฏสองครั้งในหน้าต่างเงื่อนไข — ยึด key ใน `*Dlg.cpp` เป็นหลักเมื่อสร้างจอ BC ไม่ลอกป้ายซ้ำ.
- **ทิศกระดาษ** : แนวตั้งเฉพาะ 5837, 5742-5745 (ขออนุมัติซื้อแบบสรุป), 5747-5750 (สั่งซื้อแบบสรุป ตามวันที่/แผนก/กลุ่มสินค้า/โครงการ), 5705-5708 (สรุปใบรับสินค้า), 5768 และ 5818 ; ที่เหลือแนวนอน.
- **ตัวเลือก combo/ช่องติ๊กที่เปลี่ยนสิ่งที่แสดง** (ป้ายจริงอยู่ใน `myresource` — ยังไม่ verify) : 5817 ใบจ่ายเงินมัดจำที่คงค้าง "ทั้งหมด / ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_CUTZERO⟩" ; 5820 รายงานการตัดใบซื้อสินค้า (อ่าน `BCApInvoice`, `bcapinvoicesub`) "ทั้งหมด / ⟨BILL_INVBALANCE⟩ / ⟨BILLBALANCE_CUTZERO⟩ / ⟨BILLBALANCE_ZERO⟩ / ⟨BILLBALANCE_PAYZERO⟩" (กรองตามยอดคงเหลือของบิล) ; 5837 "ทั้งหมด / เฉพาะที่อนุมัติแล้ว / เฉพาะที่ยังไม่อนุมัติ / ⟨MAKE_APPROVED⟩" ; 5840 สถานะใบสั่งซื้อ "ทั้งหมด / ⟨ONLY_BUY⟩ / ⟨SPECIAL_UNSHIPMENT_ITEM⟩" ; 5815 มีช่องติ๊ก "ที่ยืนยันแล้ว / ที่ยังไม่ยืนยัน" + combo ⟨STK_REQUST_1/2⟩ ; 5832 เลือก ⟨CP_DOCTYPE_1..4⟩ (4 ค่า — ความหมายยังไม่ verify).
- **รายงานที่แสดง "จำนวนคงเหลือ" พึ่งสถานะ/ยอดคงเหลือที่เอกสารคำนวณไว้** : 5837 อ่าน `BCStkRequestSub`/`BCSTKREQUEST` (ยอดคงเหลือ = `RemainQty` ที่ `UpdateDocumentFlow` ชนิด 16 เขียน §0.5) ; 5840 อ่าน `BCPurchaseOrderSub` (จำนวนคงเหลือของ PO) ; 5820 และ 5817 อ่านยอดคงเหลือของบิล/มัดจำ (`BillBalance`) — ถ้าสถานะ/ยอดเพี้ยน (เช่นกรณีแก้ IR §0.5 ข้อ 1) รายงานเพี้ยนตาม ; BC ต้องให้รายงานอ่านค่าที่ระบบคำนวณใน transaction เดียวกับเอกสาร ; เงื่อนไข SQL รายตัวยังไม่ได้ถอด (ยังไม่ verify).
- **รวมยอด** : กลุ่ม "แบบสรุป" ของการสั่งซื้อ (5747-5750), ส่งคืน/ลดหนี้ (5756-5762, 5768) และเพิ่มหนี้ (5772-5788) ไม่พบตัวแปรผลรวมในโค้ด พบเฉพาะรวมต่อหน้า — ยังไม่ verify ว่ารายงานเหล่านี้พิมพ์แถวรวมท้ายรายงานจริงหรือไม่ (ตัวสแกนค้นจากชื่อตัวแปร). BC ควรมีแถวรวมท้ายรายงานทุกตัวและตรวจกับรายงานของ Champ ที่รันจริง.
- **คลิกทะลุจากรายงานไปเอกสารต้นทาง** : รายงานลงทะเบียนปลายทางด้วย `REPComm_InsertCommand(ช่อง, command)` (กลไกอธิบายใน `10-reports.md` §3.9). ปลายทางที่พบในรายงานจัดซื้อ (นับจากทั้ง 110 รายงานใน §2.1+§2.2 ; รายงานหนึ่งตัวมีได้หลายปลายทาง) :

| command ปลายทาง | เปิดจอ | จำนวนรายงาน |
|---|---|---|
| `ID_ICSYS_ITEMMASTER` | 5602 รายละเอียดสินค้า (`07-inventory.md`) | 83 |
| `ID_APSYS_APMASTER` | 5033 รายละเอียดเจ้าหนี้ (`03-ap.md` §1.1) | 48 |
| `ID_POSYS_TRANS_6` | 5812 ซื้อสินค้า/บริการ (§1.5) | 20 |
| `ID_POSYS_TRANS_4` | 5810 ใบสั่งซื้อ (§1.1) | 16 |
| `ID_POSYS_TRANS_11` | 5799 ตั้งหนี้ IR (§1.4) | 12 |
| `ID_POSYS_TRANS_8` | 5814 เพิ่มหนี้ (§1.9) | 11 |
| `ID_POSYS_TRANS_7` | 5813 ส่งคืนสินค้า/ลดหนี้ (§1.7) | 10 |
| `ID_POSYS_TRANS_10` | 5798 รับสินค้า GR (§1.3) | 8 |
| `ID_POSYS_TRANS_1` | 5797 ใบเสนอซื้อ (§1.10) | 7 |
| `ID_POSYS_TRANS_5` | 5811 จ่ายเงินมัดจำ (§1.15) | 3 |
| `ID_POSYS_TRANS_12` | 5800 ส่งคืนสินค้า (§1.6) | 3 |
| `ID_POSYS_TRANS_13` | 5801 ลดหนี้ (§1.8) | 2 |
| `ID_POSYS_TRANS_21` | 5806 จ่ายเงินล่วงหน้า (§1.16) | 2 |
| `ID_ICSYS_SERIAL_NUM` | 5618 บันทึก Serial Number (`07-inventory.md`) | 2 |
| `ID_POSYS_TRANS_3`, `_15`, `_22`, `_25` | 5809 · 5802 · 5807 · 5808 (§1.13, §1.12, §1.17, §1.2) | 1 ต่อคำสั่ง |

  (ตัวแปลง command → menu id มาจาก `atlas-01-PO.md` ; ตัวเลขนับจากการลงทะเบียนใน handler ไม่ใช่ผลรันจริง — ยังไม่ verify ว่าทุกคอลัมน์ที่คลิกได้เปิดจอถูกตัว.)
- **ตัวกรองที่ผู้ใช้เพิ่มเอง (F9) และการเรียงลำดับ** : ใช้กลไกกลาง (`10-reports.md` §3.4) ; คอลัมน์ "ตัวเลือก combo: เรียงรายงาน" ในตารางด้านบนคือค่า `sort` ของแต่ละรายงาน.

## 3. BC เทียบ Champ และสิ่งที่ต้องปรับ

> หมายเหตุขอบเขต (กฎลุงจืด): ตอนนี้ BC ทำเฉพาะบัญชีแยกประเภท (GL) — ระบบซื้อทั้งชุดเป็นงานอนาคต ไม่ได้อยู่ในคิวปัจจุบัน ; ส่วนนี้จึงเป็นรายการช่องว่างเมื่อถึงเวลาสร้าง ไม่ใช่งานที่ต้องเริ่มตอนนี้. ข้อที่ชนกับกฎเจ้าของระบุใน §3.5 โดยไม่เสนอให้เปลี่ยน. สถานะ BC ตรวจจริง 2026-10-07 จาก `frontend/src/lib/menu-data.ts:176-235`, `erp-transaction.ts`, `erp-operations.ts`, `erp-reports.ts`, `menu-screen-status.ts` และ `backend/` (ไม่ได้รันระบบ).

### 3.1 ภาพรวมสถานะ BC ของเมนู PO ทั้ง 96 รายการ

กฎสถานะของ BC : เส้นทางเอกสาร ERP (`isErpTransactionRoute`) และหน้า "operations" (`isOperationsRoute`) ถูกนับเป็น "รอพัฒนา" ทั้งหมด (`frontend/src/lib/menu-screen-status.ts:46-50`) ; รายงานที่มี config แต่ไม่อยู่ในชุด API พร้อม จะขึ้น "ข้อมูลยังไม่พร้อม" (`menu-screen-status.ts:79-83` + `erp-reports.ts:530-546`).

| กลุ่มใน Champ | menu id ของ Champ | BC menu id / route (`menu-data.ts` ; config) | สถานะ BC |
|---|---|---|---|
| ใบสั่งซื้อ | 5810 ใบสั่งซื้อ | `purchase-order` `/transaction/purchaseorder` (:185 ; `erp-transaction.ts:195`, apiPath `purchase-order`) | รอพัฒนา ; โมเดลเอกสารทั่วไป ไม่ตรงช่องของ Champ (§3.6) |
| | 5808 ยกเลิกใบสั่งซื้อ | `purchase-order-cancel` `/procurement/order-cancellation` (:186 ; `erp-operations.ts:37`) | รอพัฒนา ; หน้า operations ทั่วไป ไม่อยู่ใน `APPROVAL_KINDS` (`:208-210`) |
| รับ/ตั้งหนี้/คืน | 5798 รับสินค้า (GR) | `purchase-partial` `/transaction/purchasepartial` (:187 ; `erp-transaction.ts:434`, apiPath `purchase`) — ชื่ออังกฤษ "Gradual Receipt" | รอพัฒนา ; ไม่แยก GR/IR/ซื้อตรง (ข้อ 2) |
| | 5799 ตั้งหนี้ (IR) | `accrual-receive` `/transaction/accrualreceive` (:188 ; `:445`, apiPath `purchase`) | รอพัฒนา |
| | 5800 ส่งคืนสินค้า (GR) | `purchase-return` `/transaction/purchasereturn` (:189 ; `:228`, apiPath `purchase-return`) | รอพัฒนา |
| | 5801 ลดหนี้ (กรอกยอด) | `purchase-reduce-debt` `/transaction/purchasereducedebt` (:190) | **ไม่มี config** (`menu-screen-status.test.ts:18` ยืนยันว่ารอ) |
| ซื้อตรง/คืน/เพิ่มหนี้ | 5812 ซื้อสินค้า/บริการ | `purchase` `/transaction/purchase` (:191 ; `:206`, apiPath `purchase`) | รอพัฒนา |
| | 5813 ส่งคืนสินค้า/ลดหนี้ | `purchase-debit-note` `/transaction/purchasedebitnote` (:192 ; `:250`, apiPath `bank/purchasedebitnote`) | รอพัฒนา ; apiPath ชี้โมดูล `bank` (ข้อ 5) |
| | 5814 เพิ่มหนี้/เพิ่มสินค้า | `purchase-credit-note` `/transaction/purchasecreditnote` (:193 ; `:239`, apiPath `purchase-return`) | รอพัฒนา ; ใช้โมดูลเดียวกับใบส่งคืน (ข้อ 5) |
| ก่อนสั่งซื้อ | 5797 ใบเสนอซื้อ | `purchase-requisition` `/transaction/purchaserequisition` (:200 ; `:184`, apiPath `purchase-requisition`) | รอพัฒนา |
| | 5805 อนุมัติใบเสนอซื้อ | `purchase-requisition-approve` `/procurement/requisition-approval` (:201 ; `erp-operations.ts:28`) | หน้า operations ; `pr_approval` ถูกนับว่า API พร้อม (`:208-210`) และ client เรียก `/api/goapi/api/approval/pr-status/pending` (`:250`) แต่ **ไม่พบ endpoint นี้ใน `backend/`** (ค้น "approval" ใน `internal/goapi` ไม่พบ route) — ยังไม่ verify ที่ runtime ; จะแสดงข้อผิดพลาดตอนโหลด |
| | 5802 ใบสืบราคารวม | `rfq` `/transaction/rfq` (:202 ; `:423`, apiPath **`purchase-requisition`**) | รอพัฒนา ; ใช้ตารางเดียวกับใบเสนอซื้อ ต่างจาก Champ (`BCEnquirePrice` แยก) |
| | 5809 ใบตารางราคารวม | `rfq-price-table` `/transaction/rfqpricetable` (:203) | **ไม่มี config** |
| | 5803 ประมวลใบสั่งซื้ออัตโนมัติ | `purchase-order-generate` `/procurement/generate-orders` (:204 ; `erp-operations.ts:118`) | รอพัฒนา ; หน้า operations ทั่วไป |
| เงินจ่ายล่วงหน้า | 5811 จ่ายเงินมัดจำ | `deposit` `/transaction/deposit` (:211 ; `:500`, apiPath `paidadvance`) | รอพัฒนา |
| | 5806 จ่ายเงินล่วงหน้า | `advance-payment` `/transaction/advancepayment` (:212 ; `:478`, apiPath `paidadvance`) | รอพัฒนา |
| | 5807 รับคืนเงินจ่ายล่วงหน้า | `advance-payment-refund` `/transaction/advancepaymentrefund` (:213 ; `:489`, apiPath `paidadvance`) | รอพัฒนา |
| เครื่องมือ | 5048 ต้นทุนแฝง | `purchase-landed-cost` `/transaction/landedcost` (:220 ; `:456`, apiPath `purchase`) | รอพัฒนา |
| | 5049 ปรับปรุงใบรับสินค้า | `purchase-receipt-adjustment` `/transaction/purchasereceiptadjustment` (:221 ; `:467`, apiPath `purchase`) | รอพัฒนา ; BC ทำเป็น "เอกสารมีรายการ" แต่ Champ เป็นเครื่องมือซ่อนสถานะ (ข้อ 8) |
| รายงานรายวัน (14) | 5838 เปรียบเทียบราคาซื้อ | `purchase-price-comparison` `/procurement/price-comparison` (:231 ; `erp-operations.ts:110`) | รอพัฒนา ; มีปุ่ม "สร้างใบสั่งซื้อจากราคาที่เลือก" (`:115`) ซึ่ง Champ ไม่มี (§3.5) |
| | 5815 5837 5817 5841 5828 5832 5833 5823 5824 5825 5826 5733 5785 | ไม่มีเมนู | ยังไม่มี (5826 ไม่มี handler ใน Champ — §2.3) |
| รายงานเชิงวิเคราะห์ — การซื้อ (14) | 5699 5720-5725 (แจกแจง 7) · 5726-5728 5730 5731 5734 5735 (สรุป 7) | `purchase-report` `/report/reportdedebipurchase` (:228 ; `erp-reports.ts:245-258`) ≈ สรุปตามวันที่/เจ้าหนี้ ; `purchase-by-product` `/report/purchasebyproduct` (:229 ; `:261-273`) ≈ 5734 | ต่าง ; หน้ารายงานทั่วไปคอลัมน์ตายตัว 5-6 ช่อง ไม่อยู่ใน API พร้อม (`erp-reports.ts:530-543`) ; รายงานที่เหลือไม่มีเมนู |
| รายงานเชิงวิเคราะห์ — สั่งซื้อ (9) | 5746 5751 5752 5738 · 5747-5750 5736 | `gradual-receipt` `/report/reportdedebipurchasepartial` (:230 ; `:276-288`) ≈ 5840 สถานะใบสั่งซื้อ (ศูนย์รายงาน) | ต่าง (5 คอลัมน์) ; ไม่ใช่ API พร้อม ; รายการอื่นไม่มีเมนู |
| รายงานเชิงวิเคราะห์ — ขออนุมัติซื้อ (8) · ใบรับสินค้า (7) · ตั้งหนี้ (7) · ส่งคืน/ลดหนี้ (8) · เพิ่มหนี้ (10) | 5698 5729 5737 5740 5742-5745 · 5701-5703 5705-5708 · 5711-5714 5716-5718 · 5756-5762 5768 · 5772-5776 5778 5779 5782 5783 5788 | ไม่มีเมนู | ยังไม่มี |
| ศูนย์รายงาน หมวดจัดซื้อ (33) | 5831 5839 5830 5840 5829 5709 5704 5820 5715 5719 5766 5821 5822 5732 5819 5767 5765 5786 5777 5787 5816 5818 5763 5764 5784 · 5741 5700 5710 5754 5755 5769 5753 5781 | ไม่มีศูนย์รายงาน (กลุ่ม "รายงานจัดซื้อ" มี 4 หน้า ด้านบน) | ยังไม่มี (กรอบศูนย์รายงานดู `10-reports.md` §4) |

สรุปจำนวน : เอกสาร/กระบวนการ 19 รายการ — 17 มี config แต่เป็น "รอพัฒนา" ทั้งหมด, 2 ไม่มี config (5801, 5809) ; รายงาน 77 + 33 รายการ — BC มี 4 หน้าที่คล้ายบางรายการ (ไม่ API พร้อมทั้ง 4), ที่เหลือไม่มี.

### 3.2 ความต่างสำคัญที่ BC ต้องปรับ (เรียงตามผลกระทบ)

1. **ทุกจอเอกสารซื้อเป็น "รอพัฒนา" และไม่มี backend รองรับ** — `frontend/src/lib/menu-screen-status.ts:46-50` นับเส้นทาง ERP/operations เป็น retired ; BFF `frontend/src/app/api/erp-transaction/[...erpPath]/route.ts:7-33` อนุญาตโมดูล `purchase`, `purchase-order`, `purchase-requisition`, `purchase-return`, `paidadvance` ฯลฯ แต่ `backend/internal/goapi/bootstrap.go:348-351` ลงทะเบียนเพียงเครื่องคิดเลขธุรกรรมและประวัติซื้อ ไม่มี route เอกสารซื้อเลย ; `backend/internal/goapi/process/process_status.go:9-23` คำนวณสถานะ PO เป็นตัวเปล่า (`TODO: implement logic จริง`). ต้องสร้างตาราง PostgreSQL + endpoint + กลไกสถานะ §0.5 ทั้งชุด.
2. **โมเดลเอกสารเดียวทั่วไป ต่างจาก 3 เอกสาร GR / IR / ซื้อตรง + ห่วงโซ่ 5 ขั้น** — `erp-transaction.ts:3-42` มี `ErpTransactionDoc` ใบเดียว (สถานะ 0/1/2 เท่านั้น `:36`, บรรทัดรายการไม่มีคลัง/ที่เก็บ/เลขที่เอกสารต้นทางต่อบรรทัด) ; ซื้อตรง, GR, IR, ต้นทุนแฝง, ปรับปรุงใบรับสินค้าใช้ apiPath `purchase` ตัวเดียวกัน (`:206,434,445,456,467`) โดยไม่มีตัวแยกชนิด (`transflag` อยู่ในชนิดข้อมูลที่ `:23` แต่ไม่พบการตั้งตามเมนู — ยังไม่ verify). ต้องตัดสินใจโมเดล (§0.3) ก่อนสร้างตาราง.
3. **สถานะและยอดคงเหลือข้ามเอกสารไม่มีเลย** — Champ มี `PoStatus` 0/2/3/4, `BillStatus`, `GRBillStatus`, `GRRemainQty`, `RemainQty`, `IsCreditNote` ที่คำนวณซ้ำโดย `UpdateDocumentFlow` (§0.5) ; BC ต้องคำนวณทั้งหมดใน `sql.Tx` เดียวกับการบันทึกเอกสาร (ไม่ใช่ขั้นแยกตามต้นฉบับ).
4. **เงินและจำนวนเป็น JS `number` / Go `float64`** — `erp-transaction.ts:11-15,28` (`qty`, `price`, `sumamount`, `totalamount`), `backend/internal/goapi/handlers/transaction_calculator.go:24-39` (`float64`), `backend/internal/authentication/models/user.go:233` (`MaxApprovalAmount float64`) ขัด Rule 20 ; Champ เองใช้ `double` + `%lf` ใน SQL (`mytools.cpp:2164`) ห้ามลอก — ใช้ `numeric(p,s)`/decimal + นโยบายปัดเศษที่ระบุ (Champ ปัดตาม `QtyPoint/PricePoint/AmountPoint` ของบริษัท).
5. **ใบส่งคืน/ลดหนี้ และใบเพิ่มหนี้ ปนกันและชื่อสลับ** — Champ แยก `BCStkRefund` (5800/5801/5813 แยกด้วย `BillType` 2/1/ไม่ใส่) และ `BCDebitNote2` (5814) ; BC ใส่ใบเพิ่มหนี้ (`purchasecreditnote`) ลง apiPath `purchase-return` (`erp-transaction.ts:239-249`) และใบส่งคืน/ลดหนี้ (`purchasedebitnote`) ลง `bank/purchasedebitnote` (`:250-259`) ; ชื่อ route/ชื่ออังกฤษของ BC ("Vendor Credit Note" = เพิ่มหนี้) ตรงกับชื่อไฟล์เดิมของ Champ ที่สลับความหมาย (`pofrminvcreditnote.cpp` ทำใบเพิ่มหนี้) แต่ key ลงบัญชีของ Champ ใช้ `DN_AP` = เพิ่มหนี้ และ `CRPURCHASE_CN` = ลดหนี้ (`glfrmpostformat.cpp:244,275`) — ตอนสร้างจริงตั้งชื่อ route/ตารางตามความหมายไทย (ต้องตรวจนิยามใบลดหนี้/ใบเพิ่มหนี้กับแหล่งทางการของกรมสรรพากรก่อน ตาม `docs/kms/21-thai-tax-form-references.md` — ยังไม่ verify ในรอบนี้).
6. **ห่วงโซ่ก่อนสั่งซื้อ 5 ขั้น BC ทำไม่ครบและผูกผิดตาราง** — 5809 ไม่มี config (`menu-data.ts:203`), 5805/5803 เป็นหน้า operations ทั่วไปที่ไม่ผูกตาราง (`erp-operations.ts:28-35,118-123`), 5802 ใช้ apiPath เดียวกับใบเสนอซื้อ (`erp-transaction.ts:423-428`) ในขณะที่ Champ แยก `BCEnquirePrice`/`BCBuyPriceCondit` ; ไม่มีกลไกตัดยอดตามจำนวน (ConfirmQty → RemainQty → PO) ของ §0.5 ชนิด 15/16.
7. **เงินมัดจำ / จ่ายล่วงหน้า / รับคืน ใช้ apiPath เดียว `paidadvance`** (`erp-transaction.ts:478,489,500`) ในขณะที่ Champ แยก 3 ตารางโครงต่างกัน (มัดจำมีใบกำกับ/ภาษีซื้อ/หัก ณ ที่จ่าย ; ล่วงหน้าไม่มีภาษี ; รับคืนอ้างใบล่วงหน้าผ่าน `EarnestNo`) และ `employeeadvance` (`:410-412`) ใช้ `paidadvance` เหมือนกัน — ต้องแยกชนิดเอกสารให้ชัด (§0.3) และห้ามคำนวณภาษีของเงินล่วงหน้าโดยเดา (ต้องตรวจแหล่งทางการ).
8. **5049 และ 5048 ถูกทำเป็นเอกสารปกติ** — Champ 5049 เป็นเครื่องมือซ่อนสถานะ GR (เขียน `GRBillStatus` เท่านั้น ไม่มีเอกสาร §1.19) แต่ BC ตั้งเป็น `hasLineItems: true` prefix `GR-ADJ` (`erp-transaction.ts:467-476`) ; 5048 ต้นทุนแฝงของ Champ เขียน `SumOfExpCost` ของบิลซื้อ (§1.18) แต่ BC ตั้ง `LC` เอกสารตัวเลือกคู่ค้า "ใบรับสินค้า / เจ้าหนี้" (`:456-465`) — ถ้า BC คำนวณสถานะ GR จาก IR เองตาม §0.5 เครื่องมือ 5049 อาจไม่จำเป็น (ถามลุงจืด §3.4 ข้อ 4).
9. **รายงาน** — BC มี 4 หน้า (`menu-data.ts:228-231`) คอลัมน์ตายตัวและป้าย "ภาษีซื้อ 7%" ฝังในหัวคอลัมน์ (`erp-reports.ts:256`) ในขณะที่ Champ อ่านอัตราต่อเอกสาร (`TaxRate`) ; ไม่ API พร้อมทั้งสามหน้ารายงานซื้อ (`erp-reports.ts:530-543`) ; Champ มี 77 + 33 รายงานแต่ละตัวมีตัวกรอง/ทิศกระดาษ/คลิกทะลุของตัวเอง (§2) ; BC ส่งตัวกรองแค่ holding + วันที่ (`10-reports.md` §0 ข้อ 3, `erp-reports.ts:823-829`) ไม่ส่งช่วงรหัสเจ้าหนี้/สินค้า/แผนก.
10. **ประเภทภาษีของเอกสาร ความหมายค่า 2 ต่างกัน** — Champ `TaxType` 0 = แยกนอก, 1 = รวมใน, 2 = **ภาษีอัตราศูนย์** (§1.1 ช่อง 3) ; BC `VatType` 2 = **ไม่กระทบภาษี** และมี `VatCal` 1 = ยกเว้นภาษี แยกต่างหาก (`transaction_calculator.go:27,29`) — อัตราศูนย์กับยกเว้นภาษีมีผลต่อ ภ.พ.30 ต่างกัน ต้องตรวจนิยามกับแหล่งทางการก่อนผูกค่า (ห้ามเดา).
11. **ไม่มีโครงรูปแบบการลงบัญชี** — Champ เลือก `BCGLFormat` ตาม FormatGroup ต่อเอกสาร (§0.4 รวม 14 กลุ่มของฝั่งซื้อ) และผู้ใช้ตั้งบรรทัดเดบิต/เครดิตเอง ; ค้น `backend/internal/generalledger` ไม่พบแนวคิดนี้ (grep `glformat`/`FormatGroup`/`posting format` ไม่พบ) — ต้องออกแบบตารางรูปแบบการลงบัญชีโดยไม่ผูกเงื่อนไขกับรหัสบัญชี (กฎ "ห้ามเงื่อนไขตามรหัสบัญชี").
12. **ลบเอกสารจริง ต่างจากกฎ BC** — Champ ลบใบสั่งซื้อและเอกสารอื่นด้วย `DELETE` (§1.1) ; BC ต้องยกเลิก/กลับรายการ (เอกสารที่ผ่านบัญชีแล้วห้ามแก้หรือลบเงียบ) และบันทึกเหตุผล/ผู้กระทำ (audit log) ; สถานะ `IsCancel`/`IsConfirm` ของ Champ ใช้ต่อได้.
13. **ข้อความผิดพลาดและป้ายต้องผ่านตารางภาษา** — Champ มีข้อความอังกฤษตายตัวหลายจุด ("Can not save" §1.16, "Confirm you adjust bill (GR) YES/NO ?" §1.19, "Please Enter Qty..." §1.13) ; BC ต้องใช้ key ใน `backend/assets/language/languages.tsv` (กฎ i18n) ไม่ลอกข้อความตายตัว.

### 3.3 จุดผิดปกติของ Champ ที่ห้ามลอก (รวมจากทุกหัวข้อ)

| จุด | หลักฐาน | สิ่งที่ BC ต้องทำ |
|---|---|---|
| แก้ใบตั้งหนี้ (IR) แล้วถอนสถานะ GR ผิดตัว (ส่งเลข IR เดิมให้ฟังก์ชันที่ตีความเป็นเลข GR ตอน nMark=-1) | `POFrmInvoiceReceive.cpp:1166-1167`, `mytools.cpp:2330-2347` (§0.5 ข้อ 1) | คำนวณ `GRRemainQty`/สถานะ GR จากเอกสารจริงของ GR ทุกใบที่เกี่ยวข้องใน transaction เดียว |
| `OnUpdateGRStatus` ของใบส่งคืน (5800) ประกอบ UPDATE `CNQty` แต่เรียก `Execute` ด้วยตัวแปร SELECT จึงไม่มีผล | `POFrmStkReturn.cpp:995-1047` (§1.6) | ตัดจำนวนคืนรายบรรทัดของ GR จริงในธุรกรรมเดียว + ทดสอบ |
| เครื่องมือ 5049: ตัวหา `ΣQty` ของ IR เปิดด้วยตัวแปร SQL ฝั่ง GR ทำให้ทุกแถวที่ติ๊กถูกตั้งเป็น "ใช้หมดแล้ว" ; ตัวกรองเจ้าหนี้/แผนกใช้ `=` ทับเงื่อนไขเดิม ; มีฟังก์ชันตายไม่มีผู้เรียก | `POFrmAdjGRStatus.cpp:562,425,437,296-368` (§1.19) | ไม่ลอก ; คำนวณสถานะ GR จาก IR เอง (ถ้าต้องมีหน้านี้ ให้เขียนใหม่และมี guard) |
| ต้นทุนแฝง 5048 ไม่มี transaction ไม่ตรวจงวดปิด ลบไม่ตรวจงวด และใช้ชื่อ log "ลบจ่ายชำระหนี้" ผิด | §1.18 E, F (`OnSaveData :651-773`, `ConfirmDeleteData :38-53`) | ใส่ `sql.Tx`, ตรวจงวด, ห้ามแก้ต้นทุนในงวดที่ปิดแล้ว |
| ประมวลใบสั่งซื้อ 5803 โหมด ก ตั้งใบอนุมัติ `BillStatus=1` ทั้งใบทันที ทำให้จำนวนที่เหลือประมวลต่อไม่ได้ ; โหมด ข ค้น `BCStkRequestSub` ด้วยเลขที่ผิดชนิด ; ไม่มี transaction ไม่เรียก `UpdateDocumentFlow` | §1.14 หัวข้อ H, I (ข้อ 1-4) | ตัดยอดตามจำนวนจริงด้วยกลไกเดียวกับ §0.5 ชนิด 15 ใน transaction เดียว |
| ลบใบอนุมัติเสนอซื้อ (5805) ไม่ตรวจว่ามีใบสืบราคา/ใบสั่งซื้ออ้างอิง ; แก้ใบสืบราคา (5802) รีเซ็ต `BillStatus=0` แม้ถูกใช้แล้ว ; 5809 ตรวจจำนวนผิดคอลัมน์ (ไม่ทำงาน) | §1.11 (`ConfirmDeleteData :215-227`), §1.12 F3, §1.13 (`ConfirmSaveData :292-385`) | ห้ามแก้/ลบเมื่อมีเอกสารอ้างอิง (ตามกติกาลบของ 5802) ; ตรวจที่ backend |
| ใบรับคืนเงินล่วงหน้า (5807) ไม่ตรวจยอดคืน ≤ คงเหลือ / `EarnestNo` มีจริงและตรงเจ้าหนี้ / เอกสารผ่าน GL แล้ว (`IsPostGL`) ; ยกเลิก 5806 ไม่คืนยอดเช็คจ่าย (ยังไม่ verify) | §1.17 หัวข้อกติกา + ข้อสังเกต 1-2, §1.16 I.2 | ใส่ guard ทั้งสามข้อ ; ตรวจยอดคืนเทียบยอดคงเหลือ |
| เครื่องมือคำนวณยอดเงินล่วงหน้าเขียนผิดตาราง (`BCArDepositSpecial` ฝั่งลูกหนี้ แทน `BCAPDepositSpecial`) | `MyDebtReCalc.cpp:2546` (§1.16 I.4) | ไม่ลอก ; คำนวณยอดคงเหลือเงินล่วงหน้าจากเอกสารจริงใน transaction |
| `UpdateDocumentFlow` ชนิด 23 ไม่กรองใบสั่งซื้อที่ยกเลิก (ชนิด 2/15 กรอง) | `mytools.cpp:2402` vs `:1704,2155` | กรอง `IsCancel=0` ให้เหมือนกันทุกชนิด |
| ชื่อ/ป้ายตกค้างจากแม่แบบ : F3 ของ 5811 เขียน "ภาษีขาย" ทั้งที่เป็นภาษีซื้อ ; ป้าย "วันที่ส่งของ" ของ 5806 เก็บวันครบกำหนด ; 5785 ชื่อ "บันทึกใบเสนอซื้อ" เปิดรายงานใบเพิ่มหนี้ ; ชื่อคลาส "CreditNote" ทำใบเพิ่มหนี้ | §1.15 F1, §1.16 I.3, §2.3, §1.9 I.1 | ตั้งป้าย/ชื่อตามหน้าที่จริง (ใช้ถ้อยคำ Champ ตามกฎ champ-wording เมื่อถูกความหมาย) |
| ตรวจ "สั่งซื้อซ้ำ" ถูกคอมเมนต์ทิ้งทั้งบล็อกแต่ยังเตือนตอนพิมพ์จำนวน | `ChmFrmPurchaseOrder.cpp:704-730`, `Page1.cpp:1147` (§1.1) | ตัดสินว่าจะเตือนหรือห้ามซ้ำ (ถามลุงจืด) ไม่ลอกแบบครึ่งๆ กลางๆ |
| ใช้ `double` + `%lf` ต่อ SQL และไม่ครอบ transaction ตลอดทางในหลายจอ (บันทึกหัว → ปรับยอด → GL แยกขั้น) | `mytools.cpp:2164,2215`; หลายจอใน §1 | decimal + `sql.Tx` เดียว ; ทดสอบ `0.1+0.2`, ขอบเขตการปัด, เอกสารซ้ำ, debit = credit |

### 3.4 คำถามที่ต้องให้ลุงจืดตัดสินก่อนเริ่มสร้างโมดูลซื้อ

1. จะยึดโมเดล Champ เต็มรูป (ซื้อตรง / รับสินค้า GR / ตั้งหนี้ IR แยกตาม `GRIRBillStatus` + ห่วงโซ่ 5 ขั้นก่อนสั่งซื้อ) หรือยุบให้เล็กลงตามกฎ "ถ้า Champ ซ้ำซ้อนให้ยุบรวม" — ผลต่อจำนวนตารางและหน้าจอ (ข้อ 2, 6 ของ §3.2)
2. ใช้ทะเบียนคู่ค้าที่มีใน GL (`SubledgerPartner` `is_supplier`, `backend/internal/generalledger/subledger_types.go:18,31`) เป็นเจ้าหนี้ของเอกสารซื้อ หรือสร้างตารางเจ้าหนี้แยก (ดู `03-ap.md` §6)
3. รูปแบบการลงบัญชีของเอกสารซื้อ (§0.4 14 กลุ่ม) จะให้ผู้ใช้ตั้งรูปแบบเองแบบ Champ หรือกำหนดชุดเริ่มต้นให้ — ต้องไม่ผูกตามรหัสบัญชี
4. เครื่องมือ 5049 ปรับปรุงใบรับสินค้า และ 5048 ต้นทุนแฝง จะทำตามต้นฉบับหรือไม่ทำ 5049 (คำนวณสถานะ GR จาก IR เอง) ; `WeightCostType` (บริหาร/บัญชี) ของ 5048 เก็บอย่างเดียวไม่มีผลต่อการคำนวณ — ต้องการ 2 มุมมองต้นทุนหรือไม่ (§1.18 F2)
5. การลบเอกสารซื้อ : ยืนยันว่า BC ใช้ยกเลิก/กลับรายการ + audit แทน `DELETE` ตามกฎ GL (ทุกเอกสารรวมใบสั่งซื้อที่ยังไม่ผ่านบัญชี?)
6. รายงาน 77 + 33 รายการ : ทำตาม Champ ทีละรายการ หรือยุบเป็นรายงานชุดละ 1 หน้าพร้อมตัวเลือกมุมมอง (วันที่/เจ้าหนี้/แผนก/กลุ่มสินค้า/โครงการ) ตามกฎยุบหน้าจอซ้ำซ้อน — และ 5826 "Vender Evalution Report" ที่ไม่มี handler ให้ข้ามหรือไม่
7. ภาษีของเงินมัดจำ/เงินจ่ายล่วงหน้า และค่า `TaxType` 2 (อัตราศูนย์ หรือไม่กระทบภาษี) ต้องกำหนดจากแหล่งทางการของกรมสรรพากร — ให้ผมค้นและทำทะเบียนอ้างอิงใน `docs/kms/21-thai-tax-form-references.md` ก่อนสร้างจอหรือไม่

### 3.5 ที่ขัดกับกฎเจ้าของ — ไม่เสนอให้ทำ ขอลุงจืดตัดสิน

| เรื่อง | ที่มา | ความขัด |
|---|---|---|
| ขอบเขต GL อย่างเดียว | `AGENTS.md` (ขอบเขต GL ล่าสุด) | ระบบซื้อทั้งชุดอยู่นอกขอบเขตปัจจุบัน — พิมพ์เขียวนี้เป็นสเปกอนาคต ไม่ใช่งานที่ต้องเริ่ม |
| ไม่เพิ่มฟังก์ชัน/เมนูที่ Champ ไม่มี (Strict Parity) | `AGENTS.md` กฎ "ยึด D:\project-champ…" | BC ใส่ปุ่ม "สร้างใบสั่งซื้อจากราคาที่เลือก" ในหน้าเปรียบเทียบราคา (`erp-operations.ts:110-116`, ปุ่มที่ `:115`) ที่ Champ 5838 เป็นรายงานล้วน ; BC มีรายงาน "รายงานซื้อ / ซื้อตามสินค้า / ทยอยรับ" ชื่อไม่ตรงรายงานใดของ Champ (`menu-data.ts:228-230`) ; BC เก็บวงเงินอนุมัติใบสั่งซื้อต่อผู้ใช้ (`user.go:230-234,383`) ที่ Champ 5805 ไม่มี (Champ อนุมัติตามจำนวนรายการ ไม่ใช่วงเงิน) |
| Rule 20 ห้าม float | `~/.agents/AGENTS.md` กฎ 20 | `erp-transaction.ts:11-15,28`, `transaction_calculator.go:24-39`, `user.go:233` ใช้ number/float64 ; Champ ใช้ `double` ในตัวคำนวณ — BC ต้องใช้ decimal ไม่ลอก |
| PostgreSQL เท่านั้น / ชื่อ lowercase | `AGENTS.md` | ชนิด `MONEY` และตารางตัวพิมพ์ใหญ่ผสม (`BCAPInvoice`, `BCIRSUB`) ของ Champ ต้องตั้งชื่อใหม่เป็น snake_case เช่น `ap_invoices`, `ir_lines` และใช้ `numeric` |
| ห้ามเงื่อนไขตามรหัสบัญชี | memory กฎ 2026-09-24 | การลงบัญชีของ Champ ผูกกับ `BCGLFormat` ที่ผู้ใช้ตั้ง (ไม่ผูกรหัสบัญชีในโค้ด) — ตรงกฎ ; ห้าม BC เพิ่มเงื่อนไขตามรหัสตอนสร้าง |
| ห้ามระบบเงินเดือน | `AGENTS.md` | ไม่เกี่ยวกับโมดูลซื้อ — ไม่มีรายการที่ชน |

### 3.6 ช่อง/ความสามารถของเอกสารซื้อที่ Champ มีแต่ BC ยังไม่มี

เทียบ `ErpTransactionDoc`/`ErpDetailItem` (`frontend/src/lib/erp-transaction.ts:3-42`) กับหัวข้อ A ของ §1.1, §1.3-§1.5 :

| ความสามารถ | Champ | BC ตอนนี้ | ต้องปรับ |
|---|---|---|---|
| ช่องหัวเอกสาร | เลขที่เอกสาร · วันที่ · ประเภทภาษี · รหัสเจ้าหนี้ · แผนก · โครงการ · สกุลเงิน · อัตราแลกเปลี่ยน · ผู้ติดต่อ · การจัดสรร · อัตราภาษี · เครดิต(วัน) · วันที่ครบกำหนด · กลุ่มเอกสาร · หมายเหตุ | มี `docno`, `docdatetime`, `custcode`, `vattype`, `vatrate`, `description`/`remark`, `branchcode` | เพิ่มแผนก โครงการ สกุลเงิน+อัตรา ผู้ติดต่อ การจัดสรร เครดิตวัน วันครบกำหนด กลุ่มเอกสาร (ลำดับและป้ายตาม §1.x) ; เปลี่ยนชื่อ `custcode` ฝั่งซื้อให้เป็นเจ้าหนี้ |
| ข้อมูลใบกำกับภาษีซื้อ (IR/ซื้อตรง) | เลขที่ใบกำกับ · วันที่ใบกำกับ · ประเภทการซื้อ · วันที่วางบิล · ตรวจเลขซ้ำของเจ้าหนี้เดียวกัน | ไม่มี | เพิ่มช่อง + กติกาเลขซ้ำ (§1.4, §1.5 กติกาก่อนบันทึก) — ภาษีซื้อบันทึกลง `BCInputTax` (ฝั่ง BC ใช้ `details.vats` ของ GL ได้) |
| รายการสินค้า | คลัง + ที่เก็บ + หน่วย + จำนวน + ราคา + ส่วนลด (เปอร์เซ็นต์หรือบาท) + จำนวนเงิน + เลขที่เอกสารต้นทางต่อบรรทัด (`PORefNo`, `GRRefNo`) ; Serial/Lot | ไม่มีคลัง/ที่เก็บ/เลขที่ต้นทาง/Serial (`itemcode`, `unitcode`, `qty`, `price`, `discount`, `sumamount`) | เพิ่มคลัง/ที่เก็บ (บังคับเฉพาะสินค้าสต็อก) + อ้างเอกสารต้นทางรายบรรทัด (หัวใจของ §0.5) |
| ภาษีหัก ณ ที่จ่าย · หักเงินมัดจำ · จ่ายเงินทันที (F6 / F8 / F7) | อยู่ในฟอร์มซื้อ/ตั้งหนี้ (§1.4, §1.5) | ไม่มีในโมเดลเอกสาร | ออกแบบร่วมกับ GL (`SubledgerWithholding` มีแล้วฝั่ง GL) ; ห้ามเดาสูตรภาษี |
| สถานะเอกสาร | ยืนยัน / ยกเลิก / ผ่าน GL / ถูกอ้างอิง (ล็อกแก้-ลบ) | `status` 0/1/2 | ใช้ป้ายสถานะของ `00-common-ui.md` §10.2 + ล็อกตามเอกสารที่อ้าง |
| พิมพ์ | ฟอร์มพิมพ์ต่อเอกสาร | มี `thai-document-print-modal.tsx` (ใช้ `ErpTransactionDoc`) | ตรวจรูปแบบพิมพ์ใบสั่งซื้อ/ใบรับสินค้าตาม Champ — ยังไม่ verify |

### 3.7 สิ่งที่ยังไม่ verify (อย่าถือเป็นข้อเท็จจริง)

- ข้อความไทยของ key ใน `myresource` (ตารางไม่อยู่ใน repo) และป้ายที่ขึ้นต้น `~` ทั้งหมดในไฟล์นี้ ; `reportcenterconfig.xml` (ผู้ใช้เห็นรายงานกลุ่มไหนตอนรันจริง).
- DDL ของตารางฝั่งซื้อ (`BCAPInvoice`, `BCIRSUB`, `BCStkRefund` ฯลฯ — ไม่มีไฟล์ DDL ใน repo ที่ตรวจ) : ชนิดคอลัมน์ ค่าเริ่มต้นของ `BillType` ในใบ 5813 และความยาวช่อง ต้องขอ DDL/ตัวอย่างข้อมูลจริงจากลุงจืดก่อนออกแบบตาราง.
- ผลรันจริงของข้อบกพร่องใน §3.3 (ทั้งหมดอ่านจากโค้ด ไม่ได้รัน Champ) ; ตัวคูณภายใน `BuyAPProcess`/`StkRefundAPProcess`/`DebitNote2APProcess`/`OnSaveInputTax`/`OnSaveAPWTax` (อยู่ในไลบรารีร่วม — ดูสรุปไว้ที่ §1.4-§1.9 เฉพาะผลต่อยอด) ; การทำงานภายในของปุ่มลัดบางตัว (`OnSelectItemByPO` Shift+F7 §1.3).
- ว่าไฟล์ `reppo` ทั้ง 110 รายงานพิมพ์แถวรวมท้ายรายงานจริงหรือไม่ (ตัวสแกนค้นจากชื่อตัวแปร) และเงื่อนไข SQL รายตัว.
- ค่า `PROCESS_CHAMP` / ตัวแปรลูกค้าเฉพาะราย (§0.2) ว่าถูกกำหนดในบิลด์จริงหรือไม่.
- ยังไม่ได้ตรวจกับแหล่งทางการของกรมสรรพากร : การบันทึกภาษีซื้อของเงินมัดจำ/เงินจ่ายล่วงหน้า, นิยามอัตราศูนย์/ยกเลิกภาษี, เกณฑ์ใบลดหนี้/ใบเพิ่มหนี้ฝั่งซื้อ — ห้ามนำสูตรของ Champ ไปใช้เป็นข้อกำหนดทางภาษีโดยไม่ตรวจ (Rule 17, `docs/kms/21-thai-tax-form-references.md`).
