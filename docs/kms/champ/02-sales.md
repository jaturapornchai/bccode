# Champ -> BC : ระบบขาย (Bill) — พิมพ์เขียวหน้าจอ เอกสาร และรายงาน

> ต้นทุนสต็อก: BC ใช้ถัวเฉลี่ยวิธีเดียว ([ADR](../decisions/2026-10-11-stock-cost-average-only.md)) — ส่วนที่กล่าวถึง FIFO/Lot/ต้นทุนมาตรฐานในเอกสารนี้อธิบายพฤติกรรมของ Champ เท่านั้น BC ไม่ทำตาม

> ที่มา: ลุงจืดสั่ง 2026-10-07 ให้ออกแบบ BC ใหม่โดยยึด Champ เป็นหลัก ให้เหมือนกันแทบทุกอย่าง (BRIEF: `scratch/champ-atlas/BRIEF.md`) ; เอกสารนี้ศึกษาโมดูล Bill (ระบบขาย) 15 รายการเมนู + รายงานขายในกลุ่ม "รายงาน" ของ Champ (`D:\project-champ`, C++/MFC + SQL Server)
> สถานะ BC (ตรวจ 2026-10-07): จอ/รายงานขายทั้งหมดอยู่ใน `frontend/src/lib/menu-data.ts` กลุ่ม Bill (`bill-orders`, `bill-payment`, `bill-transactions`, `bill-reports`) แต่เส้นทาง ERP transaction/report ส่วนใหญ่ถูกนับเป็น "รอพัฒนา" (`frontend/src/lib/menu-screen-status.ts:46-49`) และขอบเขตงานตอนนี้ของลุงจืดคือ GL อย่างเดียว — เอกสารนี้เป็นพิมพ์เขียวไว้สร้างระบบขายภายหลัง ไม่ใช่งานทำทันที
> ข้อควรระวังการอ้างอิง: ข้อความไทยของ Champ ที่มาจาก `GetMyRes("KEY")` เก็บในตาราง SQL `myresource` ซึ่งไม่อยู่ใน repo — ข้อความที่ระบุว่า "(อนุมาน)" ไม่ใช่ถ้อยคำจริงของ Champ ; ข้อความที่ไม่มี "(อนุมาน)" มาจากกล่องโต้ตอบ (.rc), คอมเมนต์ในโค้ด หรือ `menuconfig.xml` ; ถ้าเอกสารระบุ "ยังไม่ verify" หมายถึงยังไม่ได้ตรวจกับโค้ด/ฐานข้อมูลจริง
> อ้างอิงไฟล์: ใช้ชื่อสั้นของ `D:\project-champ\champ\champ\` (เช่น `billfrmsubmitprice.cpp:2437`) ; ไลบรารี `champ-lib` ; รายงาน `champ-report\BC5REP_BILL`/`BC5REP_AR` (ในเอกสารเรียก `rep-bill\`, `rep-ar\`)

## 0. โครงเมนู Bill ของ Champ (15 รายการ ตามลำดับบนจอ)

ที่มา `champ\menuconfigxml\menuconfig.xml:148-168` (ไฟล์ UTF-16LE; เส้นคั่น `separate` แสดงเป็นเส้นแบ่งกลุ่ม) :

| กลุ่ม | menu id | ชื่อเมนู (Champ) | ประเภท | หัวข้อในเอกสารนี้ |
|---|---|---|---|---|
| ใบเสนอราคา | 5303 | บันทึกใบเสนอราคาสินค้า | จอบันทึกเอกสาร | ส่วน 1 [5303] |
| | 5307 | อนุมัติใบเสนอราคาสินค้า | จออนุมัติ | ส่วน 1 [5307] |
| | 5000 | ยกเลิกใบเสนอราคาสินค้า | จอยกเลิก | ส่วน 1 [5000] |
| ใบสั่งขาย/สั่งจอง | 5326 | บันทึกใบสั่งขาย/สั่งจองสินค้า | จอบันทึกเอกสาร | ส่วน 2 [5326] |
| | 5305 | อนุมัติใบสั่งขาย/สั่งจองสินค้า | จออนุมัติ | ส่วน 2 [5305] |
| | 5308 | ยกเลิกใบสั่งขาย/สั่งจองสินค้า | จอยกเลิก | ส่วน 2 [5308] |
| | 5001 | Flow ใบสั่งจองสินค้า | จอติดตามสถานะ | ส่วน 2 [5001] |
| | 591 | ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า | จอตรวจสอบ | ส่วน 2 [591] |
| เงินก่อนส่งของ | 5327 | บันทึกใบรับเงินมัดจำ | จอบันทึกเอกสาร | ส่วน 3 [5327] |
| | 5329 | บันทึกใบรับเงินล่วงหน้า | จอบันทึกเอกสาร | ส่วน 3 [5329] |
| | 5306 | บันทึกคืนเงินรับล่วงหน้า | จอบันทึกเอกสาร | ส่วน 3 [5306] |
| ขาย/รับคืน/เพิ่มหนี้ | 5328 | บันทึกขายสินค้า,บริการ | จอบันทึกเอกสาร (ใบส่งของ/ใบกำกับ) | ส่วน 4 [5328] |
| | 5330 | บันทึกใบรับคืนสินค้า,ลดหนี้ | จอบันทึกเอกสาร | ส่วน 5 [5330] |
| | 5331 | บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้) | จอบันทึกเอกสาร | ส่วน 5 [5331] |
| ปรับปรุง | 5304 | ปรับปรุงวันที่ส่งของให้ลูกค้า | จอปรับปรุง | ส่วน 2 [5304] |

รายงานขาย (ตระกูล "รายงาน" > จัดขาย) เป็น popup `disable="yes"` ที่เปิดจาก Report Center — ดูส่วน 6-7.

## 00. กติการ่วมของจอเอกสารใน Champ (อ่านก่อน — ทุกจอด้านล่างอ้างถึง)

จอบันทึกเอกสารทั้งหมดสืบจาก `CMyEspeciallyFormView` (หัวเอกสาร + ตารางรายการ `m_BInputListWnd`) :

| หัวข้อ | พฤติกรรม | หลักฐาน |
|---|---|---|
| แถบเครื่องมือมาตรฐาน | ปุ่ม 10 ปุ่มตามลำดับ : New, Save, Del, Find, Prev, Next, Close, Info, Key (`IDC_MORE`), Recur (มีเมนูย่อยแบบดรอปดาวน์) ข้อความปุ่มเป็นอังกฤษฝังในโค้ด | `lib-frmwrk\mybasedbformview.cpp:2237-2262` |
| คีย์ลัดระดับโปรแกรม (ตาราง `IDR_MAINFRAME_CHM`) | Ctrl+N เอกสารใหม่ ; F11 บันทึก ; Ctrl+F ค้นหา ; Shift+F4 ค้นหา (F4) ; Ctrl+P พิมพ์ ; F12 เส้นทางเอกสาร (Document Flow) ; Ctrl+F5 ยกเลิกเอกสาร ; Shift+F5 อนุมัติเอกสาร ; F5 คำอธิบาย/รายละเอียดอื่น ; F3 ภาษี ; F6 ภาษีหัก ณ ที่จ่าย ; F7 รับ/จ่ายเงิน ; F8 มัดจำ ; F2 บัญชี (GL) ; Shift+F2 ขยายแถว ; Ctrl+F2 ตรวจสต๊อก ; Shift+F12 หน่วยสินค้า ; Ctrl+F12 ข้อมูลสินค้า ; Shift+F3 / Ctrl+F3 ดึงเอกสารอ้างอิง (`ID_HK_GETIR`/`ID_HK_GETSO`) ; Shift+F9 ล็อต ; Ctrl+F9 ซีเรียล ; Shift+F7 สินค้าตามใบสั่งซื้อ ; Ctrl+F7/Ctrl+F8 ตั้งคลัง/ที่เก็บเริ่มต้น ; Shift+F8 หน่วยคู่ ; Page Up/Page Down เอกสารก่อนหน้า/ถัดไป ; Ctrl+B บาร์โค้ด ; Ctrl+Home ยกเลิกการแจ้งเตือน ; F9 ยกเลิกการพัก (ความหมายของคำสั่งอนุมานจากชื่อ `ID_HK_*` ในตาราง ยกเว้นที่ยืนยันด้วยโค้ดของแต่ละจอ) | `champ\BC5Account.rc:977-1012` |
| คีย์ในฐานคลาสพื้นฐาน | F1 ช่วยเหลือ ; Shift+F11 แก้คำอธิบาย (`OnEditDesc`) | `lib-frmwrk\mybasedbformview.cpp:2264-2275` |
| แต่ละจอ override คีย์เพิ่ม | `PreTranslateMessage` ของแต่ละฟอร์มกำหนดคีย์ที่ใช้ได้จริง (ดูในแต่ละหัวข้อ) | ต่อจอ |
| สถานะเอกสาร (ธง) | `IsCancel` (ยกเลิก), `IsConfirm` (อนุมัติ), `IsCompleteSave` (บันทึกสมบูรณ์), `IsPostGL` (ผ่านบัญชีแยกประเภทแล้ว — ห้ามแก้/ลบ), `PayBillStatus` (สถานะวางบิล) ; ป้ายสถานะมุมจอ "CC" = ยกเลิก, "CF" = อนุมัติ, "0.00" = ยอดคงเหลือหมด (ดูรายละเอียดในส่วน 3) | `myespeciallyformview.cpp:5590-5740` (ดู ส่วน 3) ; `mybasedbformview.cpp:4126` |
| ตรวจก่อนบันทึก | ตรวจงวดบัญชีปิดด้วย `IsPeriodClose(date)` ; ตรวจวงเงินเครดิตลูกค้า (`BillDebtWarning`, `CrAuthorizeMan` — ใช้จอ `CMyConfirmCreditLimit` `IDD_MY_CONFIRM_CREDITLIMIT`) ; ตรวจเช็คอ้างอิง (`IsChqRefence`) | ดูในหัวข้อของแต่ละจอ |
| การลงบัญชี | เอกสารขายไม่ลง GL ทันทีตอนบันทึก — เก็บธง `IsPostGL` แล้ว **โมดูลบัญชีแยกประเภท (ประมวลผลลงบัญชี)** อ่านเอกสารไปสร้างรายวัน ; จอเอกสารมี F2 ดูรายการบัญชีที่จะเกิด | `IsPostGL` ใน `mybasedbformview.cpp:4126` ; รายละเอียดผังบัญชีของแต่ละเอกสารอยู่นอกขอบเขตเอกสารนี้ (ยังไม่ verify) |

---

---

## ส่วน 1 — ใบเสนอราคา: [5303] บันทึก / [5307] อนุมัติ / [5000] ยกเลิก

> ที่มา: `D:\project-champ` ผ่าน mirror ที่ถอดรหัสแล้ว `scratch\champ-atlas\dec-sa\` (เลขบรรทัดตรงกับไฟล์ต้นฉบับ) — อ้างหลักฐานแบบสั้น `ไฟล์:บรรทัด`
> ป้ายไทยที่ระบุ "(rc)" = ข้อความไทยตรงตัวจาก `KEY$ไทย` ใน `BC5Account.rc` ; ป้ายที่ระบุ "(อนุมาน)" = `(ไทยจริงอยู่ใน DB myresource ไม่อยู่ใน repo — ความหมายอนุมาน)` จากคอมเมนต์โค้ดหรือการใช้งาน
> ชื่อตาราง/คอลัมน์ Champ (SQL Server) ใช้เป็นหลักฐานเท่านั้น — BC ใช้ PostgreSQL ตัวพิมพ์เล็ก snake_case ; จำนวนเงิน = `numeric(p,s)` ห้าม float/MONEY (Rule 20)

### 0. ข้อควรรู้ก่อนอ่าน

| หัวข้อ | สรุป | หลักฐาน |
|---|---|---|
| ฟอร์มที่เมนู 5303 เปิดจริง (default build) | `CChmFrmQuotation` ฟอร์ม 2 แท็บ (ไม่ใช่ `CBILLFrmSubmitPrice` ตามที่เดาไว้ตอนแรก) ; `m_bChmMode=TRUE` ตายตัว | `BC5Account.cpp:256`, `BC5Account.cpp:1564-1570`, `MainFrm.cpp:2087-2095` |
| แท็บ 1 `MST_HEADER` (คอมเมนต์โค้ด "ข้อมูลหลัก") | หัวเอกสาร + grid + ยอดรวม = `CChmFrmQuotationPage1` (dialog `IDD_CHM_BL_FRM_QUOTATION_PAGE1`, 560x280) | `ChmFrmQuotationPage1.cpp:37-52` |
| แท็บ 2 `MST_OTHER` (คอมเมนต์ "รายการอื่นๆ") | ข้อมูลอื่น = `CChmFrmQuotationPage2` (`IDD_CHM_BL_FRM_QUOTATION_PAGE2`, 540x280) | `ChmFrmQuotationPage2.cpp` |
| ตัวแปรสถานะร่วม | `Quotation_InfoEx` (`CChmFrmQuotation_InfoEx`) เก็บยอด/ภาษี/วันที่ ที่ 2 แท็บใช้ร่วม | `ChmFrmQuotation_InfoEx.cpp:306-446` |
| ฟอร์มเดิม (legacy) | `CBILLFrmSubmitPrice` (`IDD_BILL_FRM_SUBMIT_PRICE`, rc:7273) ถูกคอมเมนต์ทิ้งใน template — คีย์เฉพาะของมัน (Shift+F3 ประวัติขาย, Ctrl+P ราคาตามระดับ, Shift+F12 ตรวจราคา, F5 รายละเอียดอื่น, ปุ่มส่งอีเมล) **ไม่มีในบิลด์ default** | `BC5Account.cpp:1564-1570` |
| ค่าคงที่ของฟอร์ม | BillCode `"QUO"` ; TransConfig `"QuotationTransConfig"` ; help `QUOTATION_HELP` ; สิทธิ์ `SetUserAuthorization(..., ID_BILLSYS_TRANS_1)` ; `m_bCanUseDocumentFlow=FALSE` แต่ยังผูก F12 → `OnDocumentFlow` | `ChmFrmQuotation.cpp:71-92` |
| ภาษาในจอ | `theApp.GetMyRes(KEY)` → ข้อความไทยอยู่ในตาราง `myresource` (ไม่อยู่ใน repo) | `BCSWinApp.cpp ~1340` |
| ระบบเงินเดือน | ไม่เกี่ยว (ไม่มี ภ.ง.ด.1/สปส. ในกลุ่มนี้) | — |

---

### 1. [5303] บันทึกใบเสนอราคาสินค้า

- **ตำแหน่งเมนู**: Bill › popup "Bill" (resourceid 30000) › รายการที่ 1 ; `menuconfig.xml:149` (id 5303, resourceid 30002 = ชื่อเมนู) ; คำสั่ง `ID_BILLSYS_TRANS_1` → `OnBillsysTrans1` (`MainFrm.cpp:362` message map, `:2087-2095` handler, `IsMenuEnable` ที่ `:2092`) ; ฟอร์ม `CChmFrmQuotation` (`ChmFrmQuotation.cpp`) + `Page1` + `Page2` ; รูปแบบอื่น: legacy ตาม §0
- **วัตถุประสงค์**: ออกใบเสนอราคาให้ลูกค้า (ไม่ตัดสต๊อก ไม่ลง GL) คำนวณภาษี/ส่วนลดระดับบรรทัดและระดับบิล ติดตามคำตอบลูกค้า/วันหมดอายุ ; ใช้เป็นเอกสารต้นทางของใบสั่งขาย/ใบกำกับ (ผ่าน `RemainQty` และ `BillStatus`)

#### 1.1 ช่องหัวเอกสาร — แท็บ 1 (Page1) ตามลำดับบนจอ

| # | ป้ายไทย | control / คอลัมน์ `BCQuotation` | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup / ตาราง | การตรวจ/คำนวณเมื่อออกจากช่อง | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | เลขที่ใบเสนอราคา: (rc `QUOTATION_NO`) | `IDC_ED_DOCNO` + ปุ่ม `IDC_BTTN_DOCNO` / `DocNo` | text `CBRunningEdit` ตัวพิมพ์ใหญ่ ยาวสุด 20 | บังคับ (`CANNOT_SAVE_NOTDOCNO`) | รันอัตโนมัติเมื่อสร้างใหม่ (ดู §1.5) | รูปแบบเลขจาก `BCSysTRFormat.DocFormat` | เกิน 20 ตัว → `DOCNO_MORE_LENG` ตัดเหลือ 20 ; ผู้ใช้ไม่มีสิทธิ์ `bCanChangeBillNo` → ช่องถูก disable + รันอัตโนมัติ | `ChmFrmQuotationPage1.cpp:1875`, `ChmFrmQuotation.cpp:1932` |
| 2 | วันที่ออกสาร: (rc `DOCNO_DATE`; คำว่า "ออกสาร" สะกดแบบนี้ใน rc) | `IDC_ED_DOCDATE` / `DocDate` | date (`BME_DATE_TH_AUTO`) | บังคับ | `theApp.m_strWorkingDate` | — | disable ถ้า `m_NewUserAutosrize.IsCanEditDocDate==0` และผู้ใช้ไม่ใช่ `sa` ; ตรวจงวดปิดตอนบันทึก | `ChmFrmQuotation.cpp:181` |
| 3 | ประเภทภาษี: (rc `TAX_TYPE`) | `IDC_ED_TYPETAX` / `TaxType` | combo 3 ค่า: 0 แยกนอก, 1 รวมใน, 2 ภาษีอัตราศูนย์ (ข้อความ combo: `CB_TAXTYPE_1`, `CB_TAXTYPE_2`, `TAX_ZERO` — ไทย **อนุมาน**) | — | `m_CompanyInfo.DefaultTaxType` ; เมื่อเลือกลูกค้า → `BCAR.TaxType` | — | `OnSelchangeTaxType` คำนวณทุกบรรทัดและยอดรวมใหม่ ; ประเภท 2 บังคับอัตราภาษี = 0 | `ChmFrmQuotationPage1.cpp:368` |
| 4 | ประเภทการเสนอราคา: (rc `QUO_TYPE`) | `IDC_ED_BILLTYPE` / `BillType` | combo 4 ค่า: 0 ขายสด, 1 ขายเชื่อ, 2 ขายบริการสด, 3 ขายบริการเชื่อ (`CB_DOCTYPE_1..4` — ไทย **อนุมาน** จากคอมเมนต์ใน Champ) | — | `DefItemBillType` ; ถ้าผู้ใช้ทำได้เฉพาะสด/เชื่อ → บังคับค่านั้น | — | `OnSelchangeDocType`: ถ้า `AutoBillDocGroup==1` รันกลุ่มเอกสาร/เลขที่ใหม่ ; เตือนถ้า `BCAR.IsMakeTrans==2` และเลือกประเภท 1 หรือ 3 ; ประเภท 0/2 ปิดช่อง "เครดิต(วัน)" ในแท็บ 2 | `ChmFrmQuotationPage1.cpp:340-366`, `:2100-2111`, `ChmFrmQuotationPage2.cpp:538-549` |
| 5 | รหัสลูกค้า: (rc `ARCODE`) | `IDC_ED_ARCODE` + ลิงก์ / `ArCode` ; ชื่อ `IDC_ED_ARNAME` (อ่านอย่างเดียว) | lookup | บังคับ (`CANNOT_SAVE_NOTARCODE`) | ว่าง | ลิงก์เปิด `CExQList` config `"ArTransConfig"` → `BCAR` | ออกจากช่อง: โหลดชื่อ (`Name1/Name2` ตาม `DefaultArName`), `TaxType`, `DefaultTaxRate` (ถ้า >=0 ไม่งั้นค่าบริษัท), `SaleCode`, `DefContactCode`, `Discount`, `DetailDiscount`, `BillCredit` ฯลฯ ; เมื่อ grid มีแถวแล้วช่องนี้ถูก disable (`NotChangeSelectAR`) | `ChmFrmQuotationPage1.cpp:1168`, `:1665` |
| 6 | ปุ่ม ปรับปรุงใบเสนอราคา (rc `UPDATE_QUOTATION`) | `IDC_BTTN_UPDATE_BILL` | ปุ่ม | — | — | เปิด `CBILLDlgUpdateBill` (ดู §2) | หลังตกลงฟอร์มปิดตัวเอง (`OnClose()`) | `ChmFrmQuotationPage1.cpp` (`OnBttnUpdateQuotation`) |
| — | (ปุ่มส่งอีเมล) | `IDC_BTTN_SENDMAIL` | ซ่อน | — | — | — | ถูกซ่อนในโค้ด | `ChmFrmQuotationPage1.cpp:216` |
| — | รายละเอียดสินค้า | `IDC_ED_ITEMDETAIL` (10,194) | text อ่านอย่างเดียว | — | — | `GetItemNameDetail` | แสดงเมื่อเลื่อนแถวใน grid (`showItemDetail`) | `ChmFrmQuotationPage1.cpp` |

#### 1.2 ช่องหัวเอกสาร — แท็บ 2 (Page2 ; `F5` สลับแท็บ)

ผูก control→คอลัมน์ `BCQuotation` ด้านล่างอนุมานจากชื่อคอลัมน์ที่ `OnSaveData` เขียน (ยังไม่ verify การผูกรายตัวใน DDX) ; ป้ายไทยทั้งหมดมาจาก rc `KEY$ไทย`

**คอลัมน์ซ้าย (x=85)**

| # | ป้ายไทย (rc) | key | คอลัมน์ | ชนิด / lookup | กฎ/ผล | หลักฐาน |
|---|---|---|---|---|---|---|
| 1 | รหัสพนักงาน | `SALECODE` | `SaleCode` (และ `UserGroup` จาก `BCSale.UserGroup`) | combo `SELECT * FROM BCSale WHERE ActiveStatus <> 0` | ค่าเริ่มต้นจาก `BCAR.SaleCode` | `ChmFrmQuotationPage2.cpp:297` |
| 2 | แผนก | `DEPART` | `DepartCode` | combo `BCDepartment` | ค่าเริ่มต้นของบรรทัด (ถ้าบรรทัดไม่ระบุ) | `ChmFrmQuotationPage2.cpp:298` |
| 3 | โครงการ | `PROJECT` | `ProjectCode` | combo `BCProject` | ใช้เป็นค่าเริ่มต้นของบรรทัด | `ChmFrmQuotationPage2.cpp:334` |
| 4 | รหัสผู้ติดต่อ | `CODE_CONTACT` | `ContactCode` | combo `BCContactList WHERE MyType=0` กรอง `parentcode`=ลูกค้า (`OnContractFilter`) | — | `ChmFrmQuotationPage2.cpp:332` |
| 5 | การจัดสรร | `ALLOCATECODE` | `AllocateCode` | combo `BCAllocate` | ค่าเริ่มต้นของบรรทัด | `ChmFrmQuotationPage2.cpp:333` |
| 6 | ยืนราคา (+ "วัน") | `VALIDITY` | `Validity` | number | — | `ChmFrmQuotation.cpp:1617-1883` |
| 7 | ลูกค้าต้องตอบรับภายใน | `AR_CREDIT` / `IDC_ED_CUSTOMERASSERT` | `CustomerAssert` (วัน) → `CallCheckDate` | number | `calcDueDate` = วันที่เอกสาร + วัน | `ChmFrmQuotation_InfoEx.cpp:429-446` |
| 8 | "เอกสารหมดอา:" (rc ตัดข้อความ — ไทยเต็มไม่ทราบ) | `EXPCREDIT` | `ExpireCredit` ↔ `ExpireDate` | number | ซ่อนวันหมดอายุเมื่อ `ExpireCredit==0` | `ChmFrmQuotationPage2.cpp:391-450` |
| 9 | ส่งมอบภายใน | `DELIVERE_CREDIT` / `IDC_ED_DAILYDAY` | `DeliveryDay` ↔ `DeliveryDate` | number | — | `ChmFrmQuotationPage2.cpp:209-291` |
| 10 | เครดิต(วัน) | `CREDITDAY` | `CreditDay` ↔ `DueDate` | number | disable เมื่อประเภทบิล 0 หรือ 2 | `ChmFrmQuotationPage2.cpp:538-549` |
| 11 | วันที่เริ่มตามลูกค้า | `TEL_DATE` | `CallCheckDate` | date | คำนวณจาก `CustomerAssert` | `ChmFrmQuotation_InfoEx.cpp:429-446` |
| 12 | Job ID | `JOBID` | `JobID` | text | ส่งออกไประบบภายนอกถ้า `m_nIsHomeBuddy` (ดู §1.8) | `ChmFrmQuotation.cpp:1617-1883` |
| 13 | BOQ Number : | — | `BOQNumber` | text | — | เดียวกัน |
| 14 | BOQ Date : | — | `BOQDate` | date | — | เดียวกัน |

**คอลัมน์ขวา (x=311)**

| # | ป้ายไทย (rc) | key | คอลัมน์ | ชนิด / lookup | กฎ/ผล | หลักฐาน |
|---|---|---|---|---|---|---|
| 1 | กลุ่มเอกสาร | `DOCGROUP` / `IDC_ED_BILLGROUP` | `BillGroup` | lookup กลุ่มเลขที่ (`ShowDocGroup` ด้วย Shift+F4) | กำหนดรูปแบบเลขที่ | `BRunningBttn.cpp:115-479` |
| 2 | รหัสสกุลเงิน | `CURRENCY_CODE` | `CurrencyCode` | combo `BCCurrency` | disable เมื่อ `MultiCurrency==0` ; เลือกแล้วโหลด `BCCurrency.ExchangeRate` ; ถ้า `MultiCurrency==1` และลูกค้ามี `HomeCurCode` → ตั้งสกุล/อัตราจากลูกค้า | `ChmFrmQuotationPage2.cpp:439-449` |
| 3 | อัตราแลกเปลี่ยน | `EXCHANGE_RATE` | `ExchangeRate` | number ทศนิยม 9 ตำแหน่ง | บังคับ | `ChmFrmQuotationPage2.cpp` |
| 4 | อัตราภาษี | `TAXRATE` | `TaxRate` | number | ค่าเริ่มต้น `DefaultTaxRate` ; ประเภทภาษี 2 = 0 | `ChmFrmQuotationPage1.cpp` |
| 5 | ยอดเงินบาทสุทธิ | `TOTAL_BATHNET` | (ไม่เก็บ header — คำนวณ `NetAmount*ExchangeRate`) | number อ่านอย่างเดียว | — | `ChmFrmQuotation_InfoEx.cpp:306-338` |
| 6 | ค่าขนส่ง | `DELLI_CHARGE` | `TotalTransport` | number | ออกจากช่อง → `BillCalc(0,TRUE,ค่าขนส่ง)` ; disable เมื่อ `!bCanChangeBillAmount` | `ChmFrmQuotationPage2.cpp` |
| 7 | วันที่หมดอายุ | `EXPDATE` | `ExpireDate` | date | ผูกกับ `ExpireCredit` สองทาง (ย้อนกลับใช้ `GetDateDiff - 1`) | `ChmFrmQuotationPage2.cpp:209-291` |
| 8 | ลงวันที่ (ส่งมอบ) | `DATED` / `IDC_ED_DAILYDATE` | `DeliveryDate` | date | ผูกกับ `DeliveryDay` | เดียวกัน |
| 9 | ลงวันที่ (ครบกำหนดชำระ) | `SETTLE_DATE` / `IDC_ED_DUEDATE` (rc ใช้ข้อความไทยเดียวกับ `DATED`) | `DueDate` | date | ผูกกับ `CreditDay` | เดียวกัน |
| 10 | คำตอบจากลูกค้า | `CUSTOMER_ASSERT` | `AssertStatus` | combo 3 ค่า `STATUS_CUSTOMER1/2/3` = รอตอบกลับ / ตอบกลับแล้ว / ไม่รับในราคา (ไทย **อนุมาน** จากคอมเมนต์โค้ด) | ค่าเริ่มต้น `SetComboData(0)` ตอน init แต่ `OnClearData` ตั้ง `SetComboData(1)` — ไม่ตรงกัน (ความหมายของพารามิเตอร์ยังไม่ verify) | `ChmFrmQuotationPage2.cpp:91`, `:458` |
| 11 | รูปแบบการสั่งซื้อสินค้า | `ORDER_FORMS` | `OrderChannel` | combo `BCOrderchannel` | — | `ChmFrmQuotationPage2.cpp:342` |
| 12 | ช่องทางการจัดจำหน่าย | `DISTRIBUTION_CHANNEL` | `DistChannel` | combo `BCDistchannel` | — | `ChmFrmQuotationPage2.cpp:341` |

**กลุ่มล่างของแท็บ 2**: เรื่อง (`SUBJECT`, `IDC_ED_SUBJECT` → `Subject`) · สิ่งที่ส่งมาด้วย (`ENCLOSURES` → `Enclosures`) · หมายเหตุ1 (`ONE_DESCRIPTION`, หลายบรรทัด → `MyDescription1`) · หมายเหตุ2 (`TWO_DESCRIPTION` → `MyDescription2`).

**พฤติกรรมแท็บ 2**

| เรื่อง | กฎ | หลักฐาน |
|---|---|---|
| วัน ↔ วันที่ | วันที่ = วันที่เอกสาร + จำนวนวัน ; ย้อนกลับ = ผลต่างวัน − 1 ; ทำเมื่อมีการแก้ช่อง (change flag) เท่านั้น ; 3 คู่: `ExpireCredit↔ExpireDate`, `CreditDay↔DueDate`, `DailyDay↔DeliveryDate` | `ChmFrmQuotationPage2.cpp:209-291` |
| ล้างข้อมูลใหม่ | `OnClearData`: อัตราภาษี = `DefaultTaxRate`, อัตราแลก = 1, `AssertStatus.SetComboData(1)` | `ChmFrmQuotationPage2.cpp:458` |
| คีย์บอร์ด | `PreTranslateMessage`: ESC ถูกกลืน ; `WM_CHAR` ตั้ง `m_bChangeData` ของฟอร์มแม่ (กันปิดโดยไม่เตือน) | `ChmFrmQuotationPage2.cpp` |
| ค่าขนส่งซ้อน | ค่าขนส่งในแท็บ 2 (ถ้า >0) ทับค่าขนส่งรวมจากบรรทัด (shadow 23) | `ChmFrmQuotationPage1.cpp:465-578` |

#### 1.3 ตารางรายการ (grid `m_BInputListWnd`) — คอลัมน์ที่เห็น

| # | ป้าย (key) | กว้าง | แก้ได้? | ชนิด | ความหมาย / สูตร | หลักฐาน |
|---|---|---|---|---|---|---|
| 0 | รหัสสินค้า (`ITEMCODE`) | 70 | แก้+ลิงก์ | EDIT+LINK | ลิงก์ธรรมดา = `CExQList "ItemTransConfig"` ; Shift+ลิงก์ = `CMyFilterItemDlg` เลือกหลายรายการ ; Ctrl+ลิงก์ = `CStockMasterFidingDlg` ; รับ `จำนวน*บาร์โค้ด` หรือ `จำนวน_บาร์โค้ด` (`IsItemBarCode`) ; รหัส `.` = แถวข้อความอิสระ ; ไม่พบ → `NOT_SEEITEMCODE` | `ChmFrmQuotationPage1.cpp:101-179`, `:1148-1156` |
| 1 | ชื่อสินค้า (`ITEMNAME`) | 170 | แก้ได้เมื่อ `m_Option.IsEditItemName==1` | text | `Name1/Name2` ตาม `DefaultItemName` | เดียวกัน |
| 2 | หน่วยนับ (`UINTCODE`) | 60 | แก้ | COMBO_TABLE | `BCItemUnit` ; สินค้าหลายบรรจุภัณฑ์ใช้ `BCStkPacking` กรองตามสินค้า ; ตรวจ `NOT_SET_UNITCODE` / `NOUNIT` (`IsUnitOfItem`) | เดียวกัน |
| 3 | จำนวน (`ITEMQTY`) | 60 | แก้ | EDIT (`QtyPoint`) | ค่าเริ่มต้น `BeginQty` เมื่อ <=0 และ `StockType != 7` | เดียวกัน |
| 4 | ราคา (`PERPRICE`) | 70 | แก้ได้เมื่อ `bCanChangePrice` | EDIT (`PricePoint`) | ราคาจาก `GetItemPrice` | `mytools.cpp:118` |
| 5 | ส่วนลด (`DISCOUNT`) | 80 | แก้ได้เมื่อ `bCanChangeItemDiscount && DetailDiscount==1` | text คำส่วนลด | ค่าเริ่มต้นจาก `BCAR.DetailDiscount` | `ChmFrmQuotationPage1.cpp:101-179` |
| 6 | จำนวนเงิน (`AMOUNT`) | 80 | แก้ได้เมื่อ `bCanChangePrice` | EDIT (`AmountPoint`) | ดูสูตร §1.4 | เดียวกัน |
| 7 | เงื่อนไข (`CONDIT_TRANS`) | 50 | แก้ | COMBO_INDEX | `OWN_RECEIVE` รับเอง / `SEND_TO` ส่งให้ (ไทย **อนุมาน** จากคอมเมนต์) → เก็บ `IsConditionSend` | เดียวกัน |
| 8 | "S" | 15 | — | — | ตัวบอกสถานะแถว | เดียวกัน |

ปุ่มสั่งท้ายแถว = ขยายแถว (`OnListExpand` → `CBILLDlgSubmitPrice`, Shift+F2).

**Shadow columns (ซ่อน) 0-35 — ไม่มี index 24** (`ChmFrmQuotationPage1.cpp:131-168`)

| idx | ความหมาย | idx | ความหมาย | idx | ความหมาย |
|---|---|---|---|---|---|
| 0 | RowOrder | 13 | BrandCode | 26 | SizeCode |
| 1 | ItemType | 14 | TypeCode | 27 | StyleCode |
| 2 | MyDescription | 15 | FormatCode | 28 | BehindIndex |
| 3 | ProjectCode | 16 | SetCode | 29 | ItemSetCode |
| 4 | AllocateCode | 17 | SetQty | 30 | OldItemCode |
| 5 | OppositeQty | 18 | WHCode | 31 | OldUnitCode |
| 6 | OppositeUnit | 19 | ShelfCode | 32 | OldQty |
| 7 | NetAmount | 20 | DiscountAmount | 33 | prompt_price_flag |
| 8 | HomeAmount | 21 | DepartCode | 34 | price_per_unit |
| 9 | ExceptTax | 22 | QuoStatusCode | 35 | prompt_price_per_unit |
| 10 | BarCode | 23 | TransportAmount | | |
| 11 | CategoryCode | 25 | ColorCode | | |
| 12 | GroupCode | | | | |

**พฤติกรรมระดับแถว**

| เรื่อง | กฎ | หลักฐาน |
|---|---|---|
| โหลดสินค้า | อ่าน `bcitem` (name1, name2, defsaleunitcode, defsalewhcode, defsaleshelf, discount, mydescription, stocktype, oppositeunit2, unittype, oppositeunit, excepttax, categorycode, groupcode, brandcode, typecode, formatcode, mysize, colorcode, myclass) เฉพาะ `activestatus=1` | `ChmFrmQuotationPage1.cpp:662-855` |
| ราคา | `unittype=2` ใช้ `OppositeUnit` ; ราคา = `GetItemPrice(item, unit, docdate, BeginQty, AR, discount, taxtype, billtype, transport)` ลำดับค้น: `BCARPricelist` → ราคาหน่วยขนาน → `BCPromoPrice` → `BCPricelist` ตามระดับราคาลูกค้า | `mytools.cpp:118` |
| ชุดสินค้า (`StockType=7`) | ดึงลูกจาก `BCITEMComponents WHERE ParentCode AND TaxType AND SaleType` ; จำนวนลูก = จำนวนแม่ × จำนวนในชุด | `ChmFrmQuotationPage1.cpp:1433-1531` |
| สินค้า+หน่วยซ้ำ | เตือน `ITEM_DUPLICATE` | `ChmFrmQuotationPage1.cpp:1402-1431` |
| ราคาต่ำกว่าทุน | ถ้า `PriceWarning` → `SET_SALE_LOW_COST` | `ChmFrmQuotationPage1.cpp:1646-1652` |
| ราคาต่ำกว่า prompt price | `PromptPriceFlag==1` และราคา/หน่วย < prompt price → `CANNOT_SAVE`+`LIST_AT`+`GetPromptPlusMessage(0)` ; รีเซ็ตราคา | `ChmFrmQuotationPage1.cpp:1319-1337` |
| ลบแถว | `ILDN_DELETESELROW` → `BillCalc` | `ChmFrmQuotationPage1.cpp` |
| รหัส `.` | แถวข้อความอิสระ ไม่ตรวจกับ `BCItem` ตอนบันทึก | `ChmFrmQuotation.cpp:1362-1615` |

#### 1.4 ยอดรวมท้ายเอกสาร และสูตรคำนวณ

**ช่องยอดรวม (rc x=449, y=196-261)**

| ป้ายไทย (rc) | key | control | ตัวแปร / คอลัมน์ | หมายเหตุ |
|---|---|---|---|---|
| รวมมูลค่าสินค้า: | `SUMITEMAMOUNT` | `IDC_ED_SUMDETAIL` | `m_dSumOfNormalTaxProduct` → `SumOfItemAmount` | รวมค่าขนส่งเข้าไปด้วย |
| มูลค่าสินค้ายกเว้นภาษี: | `TOTAL_NOTTAX` | `IDC_ED_EXCEPTTAX` | `ExceptTaxAmount` | นับเมื่อบรรทัด `ExceptTax==1` และ `m_CompanyInfo.ExceptTax==1` |
| ส่วนลด%,บาท: | `DISCOUNT` | ปุ่ม `IDC_BN_DISCOUNT_WORD` ("%") + `IDC_ED_DISC_AMOUNT` | `DiscountWord` / `DiscountAmount` | ส่วนลดระดับบิล |
| ภาษีมูลค่าเพิ่ม: | `TAXAMOUNT` | `IDC_ED_TAXAMOUNT` | `TaxAmount` | disable ตาม `UpdateTaxControl` |
| มูลค่ารวมภาษี: | `SUMTAXAMOUNT` | `IDC_ED_SUMTAXAMOUNT` | `TotalAmount` | — |
| มูลค่าเงินสุทธิ: | `TOTAL_NET` | `IDC_ED_SUMAMOUNT` | `NetAmount` | — |

สิทธิ์ที่ปิดช่องยอดรวม: `!bCanChangeBillDiscount` หรือ `BillDiscount==0` (ส่วนลดบิล) ; `!bCanChangeBillAmount` (ทุกยอด) ; `UpdateTaxControl` (ภาษี).

**ห่วงโซ่คำนวณ**: `DetailCalc` (แถว) → `BillCalc` (รวมจาก grid) → `InfoEx.billCalc` → `calcDiscount` → `calcTax` → `BillTaxProcess`

| ขั้น | สูตร / กฎ | หลักฐาน |
|---|---|---|
| คำนวณแถว (`DetailCalc`) | `DiscountMethod==0`: `amount = qty*price` แล้ว `XDiscount` กับ amount ; `DiscountMethod!=0`: `XDiscount` กับราคาก่อน แล้ว `amount = qty*price`, ส่วนลด = `discount*qty` | `ChmFrmQuotationPage1.cpp:389-455` |
| มูลค่าสุทธิแถว | ภาษีรวมใน (`TaxType==1`) และไม่ยกเว้นภาษี: `NetAmount = amount - amount*rate/(100+rate)` ; `HomeAmount = NetAmount*ExchangeRate` | `ChmFrmQuotationPage1.cpp:389-455` |
| รวมจาก grid (`BillCalc`) | รวม amount/qty/ค่าขนส่ง (shadow 23 × qty) แยกกลุ่มภาษีปกติ/ยกเว้น ; ค่าขนส่งจากบรรทัด (ถ้า >0) ทับ ; ค่าขนส่งแท็บ 2 (ถ้า >0) ทับอีกชั้น ; ไม่มีบรรทัดมียอด → `emptyDocAmt()` ไม่งั้น `billCalc()` ; แล้ว `calcDueDate` | `ChmFrmQuotationPage1.cpp:465-578` |
| `billCalc` | `m_dSumOfNormalTaxProduct += m_dTransportAmount` ; `SumOfItem = ยกเว้น + ปกติ` ; `calcDiscount` ; `calcTax` ; `NetAmount`: ประเภท 1 = `TotalAmount + (ยกเว้น − ส่วนลดยกเว้น)`, ประเภทอื่น = `(ปกติ − ส่วนลดปกติ) + (ยกเว้น − ส่วนลดยกเว้น) + ภาษี` แล้ว `XRound` ; `HomeAmount = NetAmount*ExchangeRate` เมื่อมีสกุลเงิน | `ChmFrmQuotation_InfoEx.cpp:306-338` |
| `calcDiscount` | ข้ามถ้า `SumOfItem==0` ; `XDiscount(SumOfItem, คำส่วนลด)`: คำคั่นด้วย `,` ; มี `%` = ร้อยละของยอดคงเหลือสะสม, ไม่มี `%` = จำนวนเงิน ; `XRound` ; ส่วนลดของส่วนยกเว้น = `ยกเว้น*ส่วนลด/SumOfItem` | `ChmFrmQuotation_InfoEx.cpp:390-427` |
| `calcTax` | เรียก `BillTaxProcess` เฉพาะเมื่อ `TaxNo` ไม่ว่าง (โค้ดตั้ง `m_strTaxNo="TAXNO"`) และยอดปกติ >0 | `ChmFrmQuotation_InfoEx.cpp:340-388` |
| `BillTaxProcess` โหมด 0-3 | โหมด 3 = คำนวณจากยอดภาษี ; โหมด 2 = จากยอดส่วนลด ; โหมด 0/1 = จากผลรวมรายการ+คำส่วนลด. ประเภท 0 (แยกนอก): `ก่อนภาษี = หลังลด − มัดจำ`, `ภาษี = ก่อนภาษี*rate/100`, `รวม = ก่อนภาษี + ภาษี` ; ประเภท 1 (รวมใน): `ภาษี = หลังลด*rate/(100+rate)`, `ก่อนภาษี = หลังลด − ภาษี`, `รวม = หลังลด` ; ประเภท 2: `ภาษี = 0` | `BCStools.cpp:155` |
| แก้ยอดส่วนลดเอง | ออกจาก `IDC_ED_DISC_AMOUNT` → โหมดภาษี 2 ; คำส่วนลดถูกแทนด้วยตัวเลขจำนวนเงิน | `ChmFrmQuotationPage1.cpp:1720` |
| กดปุ่ม "%" | เปิด `CLIBDlgDiscount` เฉพาะเมื่อ `SumOfItem>0` | `ChmFrmQuotationPage1.cpp:1677` |
| อัตราภาษี | ใบใหม่ = `m_CompanyInfo.DefaultTaxRate` ; จากลูกค้า = `DefaultTaxRate` ถ้า >=0 ไม่งั้นค่าบริษัท ; ประเภท 2 = 0 | `ChmFrmQuotationPage1.cpp` |
| วันที่คำนวณ (`calcDueDate`) | `ExpireDate`, `DueDate`, `DeliveryDate`, `CallCheckDate` = วันที่เอกสาร + (`ExpireCredit`, `CreditDay`, `DeliveryDay`, `CustomerAssert`) วัน | `ChmFrmQuotation_InfoEx.cpp:429-446` |

#### 1.5 เลขที่เอกสาร

| เรื่อง | กฎ | หลักฐาน |
|---|---|---|
| ตัวควบคุม | `CBRunningEdit` + `CBRunningBttn` ; รูปแบบอ่านจาก `BCSysTRFormat.DocFormat` ตาม TransKey | `BaseDBFormView.cpp:1022` |
| สัญลักษณ์ในรูปแบบ | `@` = กลุ่มเอกสาร ; `#` = เลขรัน (จำนวน `#` = ความยาวเลขศูนย์นำหน้า) ; `YY`/`MM`/`DD` แทนด้วยวันที่เอกสาร ; ปี พ.ศ. เมื่อ `EXPORT_nYearType==0` | `BRunningBttn.cpp:115-479` |
| `OnRun` ไม่มีรูปแบบ | `MAX(docno) + 1` (HighRunning) | `BRunningBttn.cpp:115-479` |
| `OnRun` มีรูปแบบ | `MAX(docno)` ที่ `LIKE คำนำหน้า%` แล้ว +1 เติมศูนย์ให้ครบจำนวน `#` | เดียวกัน |
| คีย์ลัด | F4 = รันเลขที่ ; Shift+F4 = เลือกกลุ่มเอกสาร (`ShowDocGroup`) | เดียวกัน |
| รันอัตโนมัติ | สร้างใบใหม่ → `OnBeginNewRecord` ; เปลี่ยนประเภทบิลเมื่อ `AutoBillDocGroup==1` ; ไม่มีสิทธิ์ `bCanChangeBillNo` → ช่อง disable + รันเอง | `ChmFrmQuotation.cpp:1932`, `ChmFrmQuotationPage1.cpp:340-366` |
| เลขซ้ำ (ใบใหม่) | ก่อนบันทึก `m_edDocNo.IsDuplicate()` → ถาม "Document No is duplicated Do you want to try again?" ; YES → `RunningAgain` ; NO → ล้างช่อง | `ChmFrmQuotation.cpp:1362-1615` |
| ความยาว | สูงสุด 20 ตัวอักษร (`DOCNO_MORE_LENG`) | `ChmFrmQuotationPage1.cpp:1875` |

#### 1.6 ปุ่ม แถบเครื่องมือ คีย์ลัด

ตารางเร่งปุ่ม `IDR_MAINFRAME_CHM` (`BC5Account.rc:977-1010`) ; ตัวจัดการจริงอยู่ใน message map ของฟอร์ม (`ChmFrmQuotation.cpp:71-92`)

| คีย์ | คำสั่ง | ผล | หลักฐาน |
|---|---|---|---|
| F11 | `ID_HK_SAVE11` | `OnRecSave` | `ChmFrmQuotation.cpp:71-92` |
| Ctrl+N | `ID_HK_NEW` | `OnHKNew` ใบใหม่ | เดียวกัน |
| Ctrl+F | `ID_HK_FIND` | `OnRecFind` → `OnRecFind2(3)` config `QuotationTransConfig` | เดียวกัน |
| Ctrl+P | `ID_HK_PRINT` | `OnCheckBeforePrint` + `OnTBPrintForm` (ตัวแปรพิมพ์ใน `GetFormData`, `UpdatePrintLog`) | `ChmFrmQuotation.cpp:641-864` |
| F12 | `ID_HK_DOCFLOW` | `OnDocumentFlow` — แสดงเส้นทางเอกสาร (ดู §1.7) | `ChmFrmQuotation.cpp:360-390` |
| F5 | `ID_HK_DOCDESC` | `OnHKF5` สลับแท็บ 1/2 | `ChmFrmQuotation.cpp:71-92` |
| Shift+F2 | `ID_HK_ROWEXPAND` | `OnHKShiftF2` → ขยายแถว (`CBILLDlgSubmitPrice`) | เดียวกัน |
| Shift+F5 | `ID_HK_CONFIRMDOC` | ยืนยันเอกสาร (§1.7) | `ChmFrmQuotation.cpp:2062` |
| Ctrl+F5 | `ID_HK_CANCELDOC` | ยกเลิกเอกสาร (§1.7) | `ChmFrmQuotation.cpp:2077` |
| Ctrl+F12 | `ID_HK_ITEMINFO` | `OnHKCtrlF12` เปิด `CChmItemDisplayDlg` (ต้องมีสินค้าหรือลูกค้า ไม่งั้น `SET_CURSER_ARITEM`) ; เมื่อตกลงและมีสิทธิ์ `bCanChangePrice` ใส่ราคา (ภาษีนอก/ใน ตามประเภทภาษี) ลงคอลัมน์ 4 แล้ว `DetailCalc(Row,5)` + `BillCalc` | `ChmFrmQuotation.cpp:2092-2156` |
| PageUp | `ID_HK_PAGE_UP` | โฟกัสเลขที่เอกสาร | `ChmFrmQuotation.cpp:71-92` |
| PageDown | `ID_HK_PAGE_DOWN` | โฟกัสเซลล์แรกของ grid | เดียวกัน |
| F3, F6, F7, F8, F2, Ctrl+F2, Shift+F12 | อยู่ในตารางเร่งปุ่ม | ฟอร์มนี้ **ไม่จัดการ** | `BC5Account.rc:977-1010` |

**แถบเครื่องมือมาตรฐาน**: ใหม่ / บันทึก / ลบ / ค้นหา / ก่อนหน้า / ถัดไป / ปิด / `ID_REC_INFO` (`OnInfo` แสดงเส้นทางเอกสาร + หน้าต่างยอดตัวเลขต่างๆ) / "Key" (`IDC_MORE`, `OnMore`) / `IDC_RECUR` (`OnRecur` → `BaseRecur`, เก็บ `RecurName`) — `mybasedbformview.cpp ~2237`, `ChmFrmQuotation.cpp:392-488`, `:506`

**หน้าต่างป๊อปอัพ "Key" (`IDC_MORE`)**: `F5_EDESCRIPTION`, `F11_SAVE`, `F12_CHECK_DOCUMENT`, `SHIFT_F2_ITEM`, `SHIFT_F5`, `CTRL_F5`, `HISTORY_PURCHASE` (Ctrl+F12) — ไทยอนุมาน (`ChmFrmQuotation.cpp:506`)

#### 1.7 สถานะเอกสารและ workflow

| flag (`BCQuotation`) | ความหมาย | ใครตั้ง / เมื่อไร | ผลต่อการแก้ไข | หลักฐาน |
|---|---|---|---|---|
| `IsConfirm` | ยืนยัน/อนุมัติแล้ว | Shift+F5 (ฟอร์มนี้) หรือเมนู [5307] ; ตั้ง `ConfirmCode`, `ConfirmDateTime` | ห้ามแก้ (`NOT_SAVE_ISCONFIRM`) ห้ามลบ (`NOT_DELETE_ISCONFIRM`) ห้ามยกเลิกเอกสาร (`CANNOT_CANCEL`) | `myespeciallyformview.cpp:5646`, `ChmFrmQuotation.cpp:2062` |
| `IsCancel` | ยกเลิกทั้งใบ | Ctrl+F5 ; ตั้ง `CancelCode`, `CancelDateTime` ; ตั้ง `IsCancel=1` ทั้ง `BCQuotation` และ `BCQuotationSub` | ห้ามแก้ (`NOT_SAVE_ISCANCEL`) ห้ามลบ (`NOT_DELETE_ISCANCEL`) | `myespeciallyformview.cpp:5590`, `ChmFrmQuotation.cpp:586` |
| `BillStatus` | 1 = ถูกอ้างอิง/ตัดยอดหมดแล้ว (คงเหลือ 0) | ตั้งโดยกระบวนการ [5000] (§4) | ห้ามแก้ (`CANNOT_SAVE_BILLSTATUS_REF`) | `BILLFrmCancelQuotation.cpp:593-621` |
| `AssertStatus` | คำตอบจากลูกค้า (รอตอบกลับ/ตอบกลับแล้ว/ไม่รับในราคา — อนุมาน) | Page2 หรือกล่อง "ปรับปรุงใบเสนอราคา" (หลังยืนยันแล้วแก้ได้ทางนี้ทางเดียว) | ไม่ล็อกการบันทึก | §2 |
| งวดบัญชีปิด | `IsPeriodClose(docdate, TRUE)` | ตรวจตอนบันทึก/ลบ | บันทึก/ลบไม่ได้ | `ChmFrmQuotation.cpp:1362-1615` |
| `IsPostGL` | **ไม่มี** ในใบเสนอราคา (ฟังก์ชันฐาน `BillCancel` ตรวจ `IsPostGL()` แต่ตารางนี้ไม่ลง GL) | — | — | `myespeciallyformview.cpp:5590` |

**ป้ายสถานะบนจอ (`DrawBillStatus`)**: `CC` = ยกเลิก (ขาวบนแดง) ; `0.00` เมื่อ `BillStatus==1` (ส่งยอดคงเหลือ 0, สีน้ำตาล) ; `CF` = ยืนยันแล้ว (เขียวน้ำเงิน) ; `N` = ปกติ — `myespeciallyformview.cpp:5685`

**ยืนยันเอกสาร (Shift+F5)** — `OnHKConDoc` (`ChmFrmQuotation.cpp:2062`)
- ถ้า `!bCanConfirmBill` แสดง `NOT_CONFIRM_NO` **แต่ไม่ return** (ข้อบกพร่อง) แล้วเรียก `BillConfirm` ฐาน (`ChmFrmQuotation.cpp:2069-2072`)
- `BillConfirm` ฐาน: `!bCanConfirmBill` → `NOT_CONFIRM_NO` + return ; ตรวจ `IsConfirm` ซ้ำ (`MSG_CANNOT_CONFIRM`) ; เปิด `CMyDlgBillConfirm` (ดู §2) ; ตกลง → `UPDATE ConfirmCode, ConfirmDateTime, IsConfirm=1 WHERE DocNo=` — `myespeciallyformview.cpp:5646`

**ยกเลิกเอกสาร (Ctrl+F5)** — `OnHKCancelDoc` (`ChmFrmQuotation.cpp:2077`)
- ถ้า `IsConfirm==1` → `CANNOT_CANCEL` ; ไม่งั้น `OnConfirmCancel` (`IsCancel==1` → `MSG_CANNOT_CANCEL`)
- `BillCancel` ฐาน: `IsPostGL()` → return ; `!bCanCancelBill` → `NOT_CANCEL_NO` ; ยกเลิกแล้ว → `MSG_CANNOT_CANCEL` ; เปิด `CMyDlgBillCancel` (§2) — `myespeciallyformview.cpp:5590`
- เมื่อตอบ YES ตั้ง `CancelCode`, `CancelDateTime`, `IsCancel=1` แล้ว `BillCancelUpdateData`: `UPDATE BCQuotation SET IsCancel=1` ; `UPDATE BCQuotationSub SET IsCancel=1` ; `AddUserLogBookCancel` ; ส่งออก "cancel" ถ้าเปิดระบบภายนอก ; `OnClearData` — `ChmFrmQuotation.cpp:586`

**เส้นทางเอกสาร (F12 / `OnInfo`)** — `CMyDocuemtFlow` (`IDD_DOCFLOW`) แสดง UNION ของ (`ChmFrmQuotation.cpp:360-390`):

| ตาราง | เงื่อนไขอ้างถึงใบเสนอราคา |
|---|---|
| `BCARDeposit` | `QARefno = DocNo` และ `IsCancel=0` |
| `BCSaleOrderSub` | `StkReserveNo = DocNo`, `IsCancel=0`, `TransState=0` |
| `BCARInvoiceSub` | `SORefNo = DocNo`, `IsCancel=0`, `TransState=0`, `RefType=2` |
| `BCCancelBillSub` | `RefNo = DocNo`, `IsCancel=0`, `TransState=0` |

คอลัมน์แสดง: `DOCNO_DATE`, `NO_REFERANCE`, `TRANSECTION`, `PRODUCTCODE`, `ITEMNAME1`, `QUANTITY_ITEM`, `AMOUNT` (ไทยอนุมาน)

#### 1.8 ผลข้างเคียง — ตารางและคอลัมน์ที่เขียน

**บันทึก** (`OnSaveData`, `ChmFrmQuotation.cpp:1617-1883`) — รูปแบบ: header `InsertDB`/`UpdateDB(BCQuotation, "DocNo='เลขเดิม'")` → `DELETE FROM BCQuotationSub WHERE DocNo=...` → ใส่ทุกแถวใหม่ ; **ไม่มี transaction ชัดเจน** และเมื่อใส่แถวล้มเหลวมีแค่ตั้ง `bDetailCompleteSave=FALSE` ซึ่งไม่มีโค้ดใดใช้ต่อ

| ตาราง | คอลัมน์ที่เขียน |
|---|---|
| `BCQuotation` (header) | DocNo, DocDate, TaxType, BillType, ArCode, SumOfItemAmount, DiscountWord, DiscountAmount, AfterDiscount, BeforeTaxAmount, TaxAmount, TotalAmount, ExceptTaxAmount, NetAmount, ZeroTaxAmount, SaleCode, DepartCode, Validity, AssertStatus, ExpireCredit, ExpireDate, CreditDay, DueDate, DeliveryDay, DeliveryDate, CustomerAssert, CallCheckDate, TaxRate, TotalTransport, BillGroup, ContactCode, CurrencyCode, ExchangeRate, AllocateCode, ProjectCode, Subject, Enclosures, MyDescription1, MyDescription2, JobID, BOQNumber, BOQDate, SCGID, DistChannel, OrderChannel, SCGBPNumber (จาก `BCAR`), UserGroup (`BCSale.UserGroup`), RecurName, IsCancel=0 ; เพิ่ม `CreatorCode` (Left 10) + `CreateDateTime` เมื่อเพิ่มใหม่ ; เพิ่ม `LastEditorCode` + `LastEditDateT` เมื่อแก้ |
| `BCQuotationSub` (ต่อแถวที่ไม่ว่าง) | DocNo, DocDate, TaxType, ArCode, SaleCode, UserGroup, CurrencyCode, ExchangeRate, ItemCode, ItemName, UnitCode, Qty, Price, DiscountWord, Amount, **RemainQty=Qty**, IsConditionSend, DiscountAmount (shadow 20), DepartCode (แถวหรือ header), ItemType, MyDescription, ProjectCode และ AllocateCode (แถวหรือ header), OppositeQty, OppositeUnit, NetAmount, HomeAmount, ExceptTax, BarCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, WHCode, ShelfCode, QuoStatusCode, TransportAmount, ColorCode, SizeCode, StyleCode, ItemSetCode, TransState=0, IsCancel=0, PackingRate1/2 (`GetPackingRateForTrans` ไม่ได้ = 1.0), behindindex (`GetBehindIndex` ตอนเพิ่ม), LineNumber=i, price_per_unit และ prompt_price_flag=0 ; คอลัมน์เฉพาะลูกค้ารายหนึ่ง (`SCG*`) เมื่อ `m_nProductCentralize==1` |
| บันทึกผู้ใช้ | `AddUserLogBook(SAVE_QUOTATION ...)` |
| ระบบเชื่อมภายนอก | เมื่อ `m_nIsHomeBuddy` และมี `JobID` → ส่ง flag add/update/delete/cancel (ฟีเจอร์เสริม ไม่ใช่แกน) |

**ลบ** (`ChmFrmQuotation.cpp:1298-1360`): ลบ `BCQuotation` และ `BCQuotationSub` แล้ว `AddUserLogBookDelete`. **ไม่เขียน**: `HomeAmount` ลง header (เขียนเฉพาะที่บรรทัด), `BillStatus`, `IsConfirm`, `IsCompleteSave` (Page1 `OnLoadData` อ่าน `IsCompleteSave` แต่ไม่มีโค้ดใดตั้งค่า)

**ไม่มีผลต่อ**: สต๊อก, ยอดลูกหนี้, GL, เช็ค, ธนาคาร (ใบเสนอราคาไม่ลงบัญชี)

#### 1.9 Guard / ข้อความเตือน

**บันทึก** (`ChmFrmQuotation.cpp:1362-1615`) ตามลำดับที่ตรวจ:

| # | เงื่อนไข | ข้อความ (key) |
|---|---|---|
| 1 | เลขที่เอกสารว่าง | `CANNOT_SAVE_NOTDOCNO` |
| 2 | งวดบัญชีปิด `IsPeriodClose(docdate, TRUE)` | (ข้อความของฟังก์ชัน) |
| 3 | รหัสลูกค้าว่าง | `CANNOT_SAVE_NOTARCODE` |
| 4 | โหมดแก้: `IsCancel` / `IsConfirm` / `BillStatus==1` | `NOT_SAVE_ISCANCEL` / `NOT_SAVE_ISCONFIRM` / `CANNOT_SAVE_BILLSTATUS_REF` |
| 5 | ไม่มีรายการสินค้า | `CANNOT_SAVE_NOTITEM` |
| 6 | แถวที่ไม่ใช่ "." ไม่พบใน `BCItem` | `PRODUCTCODE <รหัส> NOT_DATA` |
| 7 | จำนวน = 0 (ประเภทบิล != 1) | `QTY_ZERO_LIST` |
| 8 | ราคา > 0 แต่จำนวนเงิน <= 0 | `CHECK_AMOUNT_ZERO` |
| 9 | หน่วยนับว่าง | `NOT_SET_UNITCODE` (นำหน้าด้วย `NOT_SAVE_FORMITEM`) |
| 10 | ยอดรวม header ไม่ตรงกับคำนวณซ้ำจากรายการ `BillCalc(0,FALSE)` | `CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT` |
| 11 | โหมดแก้: ตรวจ prompt price | `CANNOT_SAVE` + `LIST_AT` + `GetPromptPlusMessage(0)` |
| 12 | โหมดเพิ่มใหม่: เลขซ้ำ (ผ่าน transaction ที่ rollback ถ้า `IsBoxDuplicate==1` ไม่งั้น `IsDuplicate()`) | "Document No is duplicated ..." |

หมายเหตุ: ลูปรวมที่ `:1480-1502` คำนวณ `dSum`, `dExceptTax`, `dZeroTax` แต่ไม่ได้ใช้ (การเทียบเดิมถูกคอมเมนต์) 

**ลบ**: ไม่มีเลขที่ → `CANNOT_DELETE_DOCNO` ; งวดปิด ; `IsCancel` → `NOT_DELETE_ISCANCEL` ; `IsConfirm` → `NOT_DELETE_ISCONFIRM` ; ถามยืนยัน `CONFIRM_DELETEDATA` — `ChmFrmQuotation.cpp:1298-1360`

#### 1.10 ข้อบกพร่อง/กับดักของ Champ (อย่าลอกไป BC)

| # | จุด | หลักฐาน |
|---|---|---|
| 1 | `InfoEx` โหลด `ExchangeRate` จากคอลัมน์ `ExceptTaxAmount` (บรรทัด 87 ทับบรรทัด 82) | `ChmFrmQuotation_InfoEx.cpp:87` |
| 2 | ฮาร์ดโค้ดภาษี 7% และ `*1.07` (ไม่มีการกันหาร `qty==0` ที่ :383) ใช้กับ shadow `price_per_unit` | `ChmFrmQuotationPage1.cpp:383`, `:394`, `:452` |
| 3 | `CalcTaxAmount` return ตั้งแต่บรรทัดแรก = โค้ดตาย | `ChmFrmQuotationPage1.cpp:857-946` |
| 4 | `billCalc` บวกค่าขนส่งซ้ำทุกครั้งที่เรียก ; `BillCalc` โหมด amount (1) ไม่รีเซ็ตยอดสะสม จึงอาจบวกซ้ำ | `ChmFrmQuotation_InfoEx.cpp:306-338`, `ChmFrmQuotationPage1.cpp:465-578` |
| 5 | Shift+F5 แจ้งไม่มีสิทธิ์แล้วไม่ return | `ChmFrmQuotation.cpp:2069-2072` |
| 6 | `AssertStatus` ตั้งค่าเริ่มต้น 0 ตอน init แต่ 1 ตอนล้างข้อมูล | `ChmFrmQuotationPage2.cpp:91`, `:458` |
| 7 | ตัวแปร `dExchangeRate` ไม่ได้ใช้ใน Page2 | `ChmFrmQuotationPage2.cpp:182-186` |
| 8 | `SelectExpireDateQuotation` และ `GetStock` ไม่มีผู้เรียก (โค้ดตาย) | `ChmFrmQuotation.cpp:295-358`, `ChmFrmQuotationPage1.cpp:2136` |
| 9 | ไม่มี transaction ครอบ header+บรรทัด ; แถวใส่ไม่สำเร็จไม่มีใครตรวจ | `ChmFrmQuotation.cpp:1617-1883` |
| 10 | `CChmItemDisplayDlg` ถูกบังคับ `m_nDisplayMode = 0` ทั้งที่ส่ง `DisplayMode=2` | `ChmItemDisplayDlg.cpp:96`, `:344` |

---

### 2. กล่องโต้ตอบที่ใช้ร่วม (อธิบายครั้งเดียว — [5307]/[5000] อ้างกลับมาที่นี่)

| กล่อง (class / IDD / rc) | ช่อง / การทำงาน | ผลลัพธ์ | หลักฐาน |
|---|---|---|---|
| `CExQList` (`Browse2` ค้นหาทั่วไป) config `"ArTransConfig"`, `"ItemTransConfig"`, `"QuotationTransConfig"` | รายการค้นหาตามคอนฟิก (คอลัมน์/ตาราง/ตัวกรองมาจากคอนฟิก) | คืนรหัสที่เลือกใส่ช่อง | `ChmFrmQuotation.cpp:71-92` |
| `CMyFilterItemDlg` (`IDD_MY_FILTER_FIND`, rc 2845, "Find Goods ID") | ตัวกรอง: กลุ่ม (`GROUPCODE_ITEM2`), ยี่ห้อ (`ITEM_BRAN`), รูปแบบ (`FORMATCODE2`), ประเภท (`ITEMTYPE2`), หมวด (`CATEGORY_CODE2`), สี (`COLOR_ITEM`), รุ่น (`ITEM_SUB`), น้ำหนัก (`ITEMWEIGHT`), เกรด (`ITEMGRADE`), ขนาด (`ITEM_SIZE`) ; ปุ่ม Auto / Process / Select All / Deselect All / OK / CANCEL | เลือกหลายรายการ → เพิ่มเป็นแถวใน grid | rc 2845 |
| `CStockMasterFidingDlg` (`IDD_LIB_STOCKMASTERFIDING`, rc 12004) | รายการสินค้า + ช่องค้น "ใส่ชื่อย่อ:" ; ENTER ค้น ; F7 = คำค้นก่อนหน้า ; F6 = ย้ายเข้ารายการ (ข้อความบางส่วนถูกตัดใน rc) | คืนสินค้าที่เลือก | rc 12004 |
| `CLIBDlgDiscount` (`IDD_LIB_DLG_DISCOUNT`, rc 3453) | จำนวนเงิน (อ่านอย่างเดียว) ; ส่วนลด % (คำส่วนลด) ; ยอดที่ลด ; ยอดหลังลด ; F11 = ตกลง ; คำนวณ `XDiscount` ใหม่เมื่อออกจากช่องคำส่วนลด | คืนคำส่วนลด + ยอดส่วนลด + ยอดหลังลด | `LIBDlgDiscount.cpp:60-113` |
| `CBILLDlgSubmitPrice` "Other" (`IDD_BILL_DLG_SUBMIT_PRICE`, rc 8087) กลุ่ม `BILLOTHER_PROPERTIES` / `ITEM_PROPERTIES` | สถานะการเสนอราคา `QUOTATION_STATUS` (`BCQuotationStatus`) ; แผนก ; โครงการ ; การจัดสรร ; คำอธิบายรายการ ; ยอดเงินสุทธิ ; ยอดบาทสุทธิ ; ค่าขนส่ง ; combo คุณสมบัติสินค้า (กลุ่ม ยี่ห้อ หมวด ประเภท รูปแบบ สี ขนาด สไตล์) ; ยอดสุทธิ/ยอดบาทแก้ไม่ได้เมื่อ `!m_bChmMode` (ไม่เกิดใน build นี้) | เขียนกลับค่าระดับแถว (`QuoStatusCode`, `DepartCode`, ...) | rc 8087 |
| `CBILLDlgUpdateBill` (`IDD_BILL_DLG_UPDATEBILL`, rc 8074, หัว `UPDATE_BILL` "ปรับปรุงบิลใบเสนอราคา") | เลือกหลายใบ (`CMyMultiSelect`) เฉพาะ `IsConfirm=1 AND IsCancel=0` ที่ยังไม่อยู่ในรายการ ; grid: `QUOTATION_NO`, `DOCNO_DATE`, `ARNAME1`, `EXPDATE` (วันที่), `CUSTOMER_ASSERT` (combo) ; ว่าง → `NOT_QUOTATION_UPDATE` | ตกลง → `UPDATE BCQuotation SET AssertStatus, ExpireDate WHERE DocNo` **ไม่ตรวจยืนยัน/ยกเลิก/งวดปิด** แล้วฟอร์มปิด | rc 8074 |
| `CChmItemDisplayDlg` (`IDD_CHM_ITEM_DISPLAY` + แท็บลูก _1.._4, rc 17992-18060+) หัว `ITEM_PROPERTIES` + รหัส/ชื่อ | แท็บ: จำนวนคงเหลือตามคลัง (`QTY` + จุดสั่ง `POINTMIN`/`POINT_MAX`/`REORDERPOINT` + สินค้าทดแทน) ; ราคาขาย (`SELLING_PRICE` : ราคาขาย 8 ระดับ `PRICE_8_LEVEL` + ราคาโปรโมชั่น `PROMOTION_SALE_PRICE`, เลือกแถวคืนราคาภาษีนอกคอลัมน์ 4 / ภาษีในคอลัมน์ 5) ; ประวัติซื้อ `BUY_HIS` ; ประวัติขาย `SALE_HIS` (สิทธิ์ `IsViewHistoryGP` / `IsViewHistoryBL`) | คืนราคาที่เลือก | `ChmItemDisplayDlg.cpp:96`, `:344` |
| `CMyDocuemtFlow` (`IDD_DOCFLOW`) | ตารางเส้นทางเอกสาร (§1.7) | อ่านอย่างเดียว | `ChmFrmQuotation.cpp:360-390` |
| `CMyDlgBillConfirm` (`IDD_LIB_BILLCONFIRM_DIALOG`, rc 3490) | คำเตือน `WANNING2` (หลังยืนยันแก้ไขเอกสารไม่ได้) + เลขที่เอกสาร (อ่านอย่างเดียว) + combo Yes/No `IDC_CBO_CONFIRM` | ผ่านต่อเมื่อเลือก "YES" เท่านั้น | `myespeciallyformview.cpp:5646` |
| `CMyDlgBillCancel` (`IDD_LIB_BILLCANCEL_DIALOG`, rc 3511) | combo Yes/No เช่นเดียวกัน | YES → ตั้ง `CancelCode`, `CancelDateTime`, `IsCancel=1` (มีกรณีพิเศษอัปเดตเช็คสำหรับ `BCRECEIPT1`/`BCPAYMENT` ซึ่งไม่เกี่ยวกับใบเสนอราคา) | `myespeciallyformview.cpp:5590` |
| `CMyDlgStockDisplay` | ใช้เฉพาะ `GetStock` ที่เป็นโค้ดตาย | — | `ChmFrmQuotationPage1.cpp:2136` |
| `CBComboPopup` | ป๊อปอัพ `OnInfo` (ยอดตัวเลขต่างๆ) และป๊อปอัพ "Key" | — | `ChmFrmQuotation.cpp:392-488`, `:506` |
| `CLIBDlgMultiTax` | **ใบเสนอราคาไม่ใช้** (ไม่พบการเรียกในไฟล์ฟอร์ม Chm ของใบเสนอราคา) | — | ยังไม่ verify นอกจากการ grep `DoModal`/`CLIBDlg` ที่พบเฉพาะ `CLIBDlgDiscount` |

---

### 3. [5307] อนุมัติใบเสนอราคาสินค้า

- **ตำแหน่งเมนู**: Bill › popup "Bill" รายการที่ 2 ; `menuconfig.xml:150` (id 5307, resourceid 30003) ; `ID_BILLSYS_TRANS_2` → `OnBillsysTrans2` (`MainFrm.cpp:413` message map, `:2494-2501`) ; ฟอร์ม `CBILLFrmApproveSubmit : CMyBaseFormView` (`billfrmapprovesubmit.cpp`, 599 บรรทัด) ; `IDD_BILL_FRM_APPROVE_SUBMIT` 343x238 (rc 8032) ; หัวจอ `GetMyRes("30003")` = อนุมัติใบเสนอราคาสินค้า (คอมเมนต์ `billfrmapprovesubmit.cpp:98` "อนุมัติใบเสนอราคา")
- **วัตถุประสงค์**: ยืนยันใบเสนอราคาที่ยังไม่ยืนยันแบบเป็นชุด (ติ๊กเลือกหลายใบ) แทนการเปิดทีละใบแล้วกด Shift+F5

##### เงื่อนไขกรอง (ตามลำดับบนจอ)

| # | ป้ายไทย (rc) | ชนิด | lookup | หมายเหตุ | หลักฐาน |
|---|---|---|---|---|---|
| 1 | เลขที่ใบเสนอราคา (`QUOTATION_NO`) จาก–ถึง ("ถึง:" `TO`) | text + ลิงก์ | `CExQList "QuotationTransConfig"` กรอง `IsConfirm=0 AND IsCancel=0` | ใส่ค่าเดียว = `=` | `billfrmapprovesubmit.cpp:451-497` |
| 2 | วันที่ออกเอกสาร (`DOCNO_DATE`) จาก–ถึง | date | — | — | `billfrmapprovesubmit.cpp:245-288` |
| 3 | รหัสลูกค้า (`ARCODE`) จาก–ถึง | text + ลิงก์ | `ArTransConfig` | — | เดียวกัน |
| 4 | รหัสแผนก (`DEPART_CODE`) จาก–ถึง | combo | `BCDepartment` | SQL ต่อสตริง (เสี่ยง injection) | เดียวกัน |
| 5 | รหัสพนักงาน (`EMPCODE`) จาก–ถึง | combo | `BCSale` | SQL ต่อสตริง | เดียวกัน |
| 6 | เลือกทั้งหมด (check box "Select All") | check | — | ติ๊กทุกแถวใน grid | — |

ปุ่ม: F5 = ประมวลผลเงื่อนไข (`F5_PROCESS_CONDIT`) · F11 = อนุมัติรายการ (`F11_ITEM_APPROVE`)

##### รายการ (ListView มี checkbox)

| # | ป้าย (key) | กว้าง | แสดงค่า | หมายเหตุ |
|---|---|---|---|---|
| 0 | "S" | 35 | checkbox | ค่าเริ่มต้นตาม `IsConfirm` ของแถวที่โหลด (= 0 เสมอ) |
| 1 | `BILL_QUOTATION` | 100 | `DocNo` | ไทยอนุมาน "ใบเสนอราคา" |
| 2 | `QUO_DATE` | 90 | `DocDate` | ไทยอนุมาน "วันที่เสนอราคา" |
| 3 | `ARCODE` | 90 | `ArCode` | รหัสลูกค้า |
| 4 | `SUBJECT` | 150 | `Subject` | เรื่อง |
| 5 | `TAXRATE` | 100 ชิดขวา | **`TaxAmount`** (ไม่ใช่อัตรา) | หัวคอลัมน์ไม่ตรงกับค่า ; อนุมาน "ภาษีมูลค่าเพิ่ม" |
| 6 | `NET_AMOUNT` | 100 ชิดขวา | **`TotalAmount`** | หัวคอลัมน์ไม่ตรงกับค่า ; อนุมาน "ยอดเงินสุทธิ" |

(`billfrmapprovesubmit.cpp:245-288`) — แถวคี่แรเงา (242,242,230)

**คำสั่งค้น**: `SELECT DocNo,DocDate,ArCode,Subject,TaxAmount,TotalAmount FROM BCQuotation WHERE IsConfirm=0 AND IsCancel=0` + เงื่อนไขช่วง, `ORDER BY DocNo, ArCode`

##### คีย์ / การทำงาน

| คีย์/การกระทำ | ผล | หลักฐาน |
|---|---|---|
| F5 | ประมวลผลเงื่อนไขแสดงรายการ | `billfrmapprovesubmit.cpp:245-288` |
| F11 | อนุมัติแถวที่ติ๊ก (ดูด้านล่าง) | `billfrmapprovesubmit.cpp:373-419` |
| PageUp | โฟกัสเลขที่จาก | — |
| F7 / PageDown | ถูกกลืน | — |
| ดับเบิลคลิกแถว | เปิดใบนั้นในจอ [5303]: `theApp.m_strUpdateData = DOCNO='..'` → ส่ง `ID_BILLSYS_TRANS_1` → `REPComm_SetCurFilter` + `REPComm_RELOAD_THIS` (มีตัวแปรท้องถิ่น `strDocNo` บังตัวแปรนอก) | `billfrmapprovesubmit.cpp:499-516` |

##### ผลข้างเคียงเมื่ออนุมัติ (F11)

| ขั้น | ผล |
|---|---|
| ต่อแถวที่ติ๊ก | `UpdateDB("BCQuotation","DocNo='..'")`: `ConfirmCode` = ผู้ใช้ปัจจุบัน, `ConfirmDateTime` = เวลาปัจจุบัน, `IsConfirm=1` |
| สำเร็จ | ข้อความ `APPROVE_QUOTATION_OK` (อนุมาน "ได้ทำการอนุมัติใบเสนอราคาเรียบร้อยแล้ว") ; `AddUserLogBook(30003, ID_BILLSYS_TRANS_2)` ; `OnClearData` |
| ไม่มีแถวติ๊ก | `NOT_PARTICULAR` (อนุมาน "ไม่มีรายการที่ต้องการอนุมัติ") |
| ตารางที่เขียน | `BCQuotation` เท่านั้น (`ConfirmCode`, `ConfirmDateTime`, `IsConfirm`) |

##### Guard (และสิ่งที่ **ไม่มี**)

- ไม่ตรวจ `IsPeriodClose` ; ไม่ตรวจสิทธิ์ `bCanConfirmBill` ในฟอร์ม (ผ่านแค่ `IsMenuEnable` ของเมนู — ยังไม่ verify ว่ามีการตั้งสิทธิ์ซ่อนที่อื่น) ; ไม่ตรวจสถานะซ้ำตอนอัปเดต (ไม่มี `WHERE IsConfirm=0` ในคำสั่งอัปเดต) — `billfrmapprovesubmit.cpp:373-419`
- ตัวกรองแผนก/พนักงานต่อสตริง SQL (`billfrmapprovesubmit.cpp:245-288`) — BC ต้องใช้ parameter
- ความหมายของ "อนุมัติ" = `IsConfirm=1` เหมือน Shift+F5 ใน [5303] (§1.7) ; ผลคือใบนั้นแก้/ลบ/ยกเลิกเอกสารไม่ได้อีก

---

### 4. [5000] ยกเลิกใบเสนอราคาสินค้า

- **ตำแหน่งเมนู**: Bill › popup "Bill" รายการที่ 3 ; `menuconfig.xml:151` (id 5000, resourceid 30022) ; `ID_30000_30022` → `OnBillSysCancelBill` (`MainFrm.cpp:1077` message map, `:2209-2216`, `IsMenuEnable(ID_30000_30022)`) ; ฟอร์ม `CBILLFrmCancelQuotation : CMyEspeciallyFormView` (`BILLFrmCancelQuotation.cpp`, 1229 บรรทัด) ; `IDD_BILL_FRM_QUOCANCELBILL` 460x277 (rc 429) ; BillCode `"CQO"` ; TransConfig `"CancelQuotaionConfig"` (สะกดผิดในโค้ด) ; help `CANCEL_QUO_HELP` ; `MainFilter = " BCCancelBill.CancelType = 2 "`
- **วัตถุประสงค์**: เป็น "ตัดยอดคงเหลือ" (void จำนวนที่ยังไม่ถูกอ้างอิง) ของบรรทัดใบเสนอราคา แล้วออกเป็นเอกสารของตัวเอง (เลขที่ `CQO`) — **ไม่ใช่** การยกเลิกทั้งใบด้วย Ctrl+F5 ใน [5303] ; เมื่อทุกบรรทัดคงเหลือ 0 ใบเสนอราคาจะได้ `BillStatus=1`

##### ช่องหัวเอกสาร (ป้ายไทยใน rc เป็นแค่ KEY — ไทยทั้งหมดต่อไปนี้ **อนุมาน** (ไทยจริงอยู่ใน DB myresource ไม่อยู่ใน repo — ความหมายอนุมาน) ; atlas ให้คำใบ้ยืมจาก dialog อื่น)

| # | ป้าย (key) | control / คอลัมน์ `BCCancelBill` | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup | หลักฐาน |
|---|---|---|---|---|---|---|---|
| 1 | `DOCNO` เลขที่เอกสาร | `CBRunningEdit` → `DocNo` | text | บังคับ | รันอัตโนมัติ | `BCSysTRFormat` | `BILLFrmCancelQuotation.cpp:808-841` |
| 2 | `DOCDATE` วันที่เอกสาร (คำใบ้ "วันที่วางบิล" น่าสงสัย) | `DocDate` | date `BME_DATE_TH_AUTO` | บังคับ | วันนี้ ; แก้ได้เมื่อ `IsCanEditDocDate` | — | เดียวกัน |
| 3 | `REMARK` หมายเหตุ | `Remark` | text | — | — | — | `:450-591` |
| 4 | `REASON` เหตุผล | `IDC_ED_REASON` → `VoidReasonCode` | combo | — | — | `SELECT top 50 * FROM BCCancelReason` | `:450-591` |
| 5 | `SUM_OF_ITMEAMOUNT` รวมมูลค่า | `IDC_ED_TOTAL` → `TotalAmount` | number อ่านอย่างเดียว | — | ผลรวม `VoidQty*Price` | — | `:450-591` |

**ตัวกรอง (กลุ่มค้นหารายการ)**

| # | ป้าย (key) | ชนิด | lookup |
|---|---|---|---|
| 1 | `FROM_DATE1` / `TO` ช่วงวันที่ | date จาก–ถึง | ค่าเริ่มต้น = งวด `BCPeriod` ที่ครอบวันนี้ (`OnSetPeriod`) |
| 2 | `GL_FROM_DOCNO` / `TO` เลขที่ใบเสนอราคา | text + ลิงก์ | `QuotationTransConfig` กรอง `BillStatus = 0 AND IsCancel = 0 AND DocNo IN (Select Distinct DocNo From BCQuotationSub where RemainQty >0.0)` |
| 3 | `FROM_AR_CODE` / `TO` | text + ลิงก์ | `ArTransConfig` |
| 4 | `FROM_SALE_CODE` / `TO` | combo | `BCSale` |
| 5 | `FROM_GOODS` / `TO` | text + ลิงก์ | `ItemTransConfig` |
| 6 | `!SELECT_ALL` | check box | ติ๊กคอลัมน์ 0 และ 10 ทุกแถว |
| 7 | `!F5_PROCESS_CONDIT` | ปุ่ม (F5) | ประมวลผลค้นรายการ |

##### ตารางรายการ — 11 คอลัมน์ (`BILLFrmCancelQuotation.cpp:150-160`)

| # | ป้าย (key) | กว้าง | แก้ได้? | ความหมาย |
|---|---|---|---|---|
| 0 | "S" | 25 | check | เลือกแถว |
| 1 | `DOCNO` | 90 | ไม่ | เลขที่ใบเสนอราคา |
| 2 | `DOCDATE` | 80 | ไม่ | วันที่ใบเสนอราคา |
| 3 | `PRODUCTCODE` | 90 | ไม่ | รหัสสินค้า |
| 4 | `ITEMNAME1` | 150 | ไม่ | ชื่อสินค้า |
| 5 | `UNITCODE` | 70 | ไม่ | หน่วยนับ |
| 6 | `SUMITEMAMOUNT` | 80 | ไม่ | มูลค่าสินค้า (อนุมานจากคอมเมนต์) |
| 7 | `QTY_QUOTATION` | 80 | ไม่ | จำนวนเสนอราคา |
| 8 | `BALANCE_NUMBER` | 80 | ไม่ | จำนวนคงเหลือ |
| 9 | `QTY_CANCEL` | 80 | **แก้ได้** | จำนวนที่ยกเลิก (ค่าเริ่มต้น = `RemainQty`) |
| 10 | "A" | 25 | check | อนุมัติ (approved) |

**Shadow 0-16**: RowOrder, ApCode, DepartCode, AllocateCode, ProjectCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, Price, RefLineNumber, UnitCode, QtyBalance, VoidAmount, BehindIndex, RemainAmount.

**ประมวลผล (F5)** — `BILLFrmCancelQuotation.cpp:325-448`: `SELECT * FROM BCQuotationSub WHERE ItemCode <> '.' AND RemainQty >0.0 AND DocNo Not In (Select DocNo From BCQuotation WHERE IsCancel = 1)` + ตัวกรอง ; `ORDER BY DocDate,DocNo,ItemCode` ; ค่าเริ่มต้นคอลัมน์ 8 และ 9 = `RemainQty` ; `RemainAmount = RemainQty*Price`. **ไม่กรองตาม `IsConfirm` ของใบเสนอราคา**. ข้อบกพร่อง: กรณีรหัสพนักงานค่าเดียวใช้ตัวแปรของลูกค้า `strStartCode` แทน `strStartSale` (`:384`).

**ตรวจจำนวนที่ยกเลิก** (`OnListChangeData` คอลัมน์ 9, `:886-907`): `VoidQty > RemainQty (+ จำนวนที่ยกเลิกเดิมตอนแก้)` → `QTY_CANCEL_NOT_OVER` + จำนวนคงเหลือ + `SET_AGAIN` แล้วรีเซ็ต ; เลือกคอลัมน์ 0 หรือ 10 กระโดดไปคอลัมน์ 9

**คีย์**: `ID_HK_NEW`, `ID_HK_FIND`, `ID_HK_PRINT`, `ID_HK_DOCDESC` (F5 = ประมวลผล) ; พิมพ์ใช้ตัวแปร DOCNO, DOCDATE, TOTALAMOUNT, DESCRIPTION/REMARK, REASONCODE, REASONNAME + ตาราง `MAIN_DETAIL` (VOIDQTY, VOIDAMOUNT ...) — SQL พิมพ์ที่ `:1029` ขาดช่องว่าง (`BCCancelBillSubWHERE`) = บั๊ก

##### สถานะ/workflow และผลข้างเคียง

| เหตุการณ์ | ผล | หลักฐาน |
|---|---|---|
| ตรวจก่อนบันทึก `OnCheckDetail` | ต้องมีอย่างน้อย 1 แถวที่ติ๊ก + approved + มีจำนวน ; ยอดรวม = Σ `VoidQty*Price` ; ไม่ผ่าน → "Please select transaction for cancel" | `:450-591` |
| บันทึก header | `BCCancelBill`: DocNo, DocDate, TotalAmount, Remark, VoidReasonCode, **CancelType=2** ; เพิ่มใหม่ + `CreatorCode`/`CreateDateTime` ; แก้ + `LastEditorCode`/`LastEditDateT` ; แก้ไม่ได้ถ้า `IsConfirm==1` (`CANNOT_UPDATE_ISCONFIRM`) ; ล้มเหลวพิมพ์ "Can not save" | `:450-591` |
| บันทึกรายการ | `BCCancelBillSub` (เฉพาะแถวที่ติ๊กและ approved): DocNo, DocDate, CancelType=2, RefNo (เลขที่ใบเสนอราคา), RefDate, ItemCode, MyDescription, UnitCode, Amount, Qty, RemainQty, VoidQty, VoidAmount (=Price*VoidQty), RemainAmount, ParentCode, DepartCode, AllocateCode, ProjectCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, Price, RefLineNumber, Approved, PackingRate1/2, behindindex, LineNumber | `:450-591` |
| ลบก่อนเพิ่ม | `DELETE FROM BCCancelBillSub WHERE DocNo AND CancelType=2` — รอบแรก (เพิ่มใหม่) ส่ง `OldDocNo=""` = ลบด้วยเลขที่ว่าง (พฤติกรรมแปลก) | `:450-591` |
| สำเร็จ | `UPDATE BCCancelBill SET IsCompleteSave=1` ; `OnUpdateQuotation(docno, +1)` ; log ; ไม่สำเร็จ → `SAVEDETAIL_NOTCOMPLETE` | `:450-591` |
| `OnUpdateQuotation` | ต่อแถว (`TransState=0`, `IsCancel=0`, `CancelType=2`): `UPDATE BCQuotationSub SET RemainQty = RemainQty - (VoidQty*mark) WHERE DocNo=RefNo AND LineNumber=RefLineNumber` (mark = +1 ตอนบันทึก, −1 ตอนแก้/ลบเพื่อย้อนกลับ) ; แล้วตั้ง `BCQuotation.BillStatus = 1` เมื่อ Σ`RemainQty <= 0` ไม่งั้น 0 | `:593-621` |
| โหลดเอกสารเดิม | อ่าน `BCCancelBillSub` (`TransState=0` หรือ `=1` ถ้าบันทึกครั้งก่อนไม่สมบูรณ์และผู้ใช้ตอบ YES `CANNOT_SAVE_DETAIL_WANT_OLDITEM`) ; จำนวนคงเหลือมาจากบรรทัดใบเสนอราคาปัจจุบัน | `:623-735` |
| ลบ | `OnUpdateQuotation(-1)` แล้ว `DELETE BCCancelBill` + `BCCancelBillSub (CancelType=2)` ; `AddUserLogBookDelete` | `:737-752` |
| ตารางที่แตะ | `BCCancelBill`, `BCCancelBillSub`, `BCQuotationSub` (RemainQty), `BCQuotation` (BillStatus), `BCCancelReason` (lookup), `BCPeriod` (ช่วงเริ่มต้น) | — |

**Guard** (`ConfirmSaveData`/`ConfirmDeleteData`, `:808-841`): เลขที่ว่าง `CANNOT_SAVE_NOTDOCNO` ; โหมดแก้และเลขที่เปลี่ยน ตรวจซ้ำใน `BCCancelBill` (`DOCNO_DUPLICATE`) ; วันที่ว่าง `CANNOT_SAVE_NOTDOCDATE` ; ลบ: `CANNOT_DELETE_DOCNO`, `CONFIRM_DELETEDATA`. **ไม่มี `IsPeriodClose`** ในฟอร์มนี้.

**ข้อบกพร่อง/กับดัก**: (1) ตัวแปรพนักงานผิด `:384` (2) SQL พิมพ์ขาดช่องว่าง `:1029` (3) `OldDocNo` ว่างตอนเพิ่มใหม่ (4) TransConfig สะกด `CancelQuotaionConfig` (5) ไม่กรอง `IsConfirm` ของใบเสนอราคา และไม่ตรวจงวดปิด

---

---

## ส่วน 2 — ใบสั่งขาย/สั่งจอง: [5326] [5305] [5308] [5001] [591] [5304]

ขอบเขตไฟล์นี้: 6 รายการเมนู — [5326] บันทึกใบสั่งขาย/สั่งจองสินค้า, [5305] อนุมัติใบสั่งขาย/สั่งจองสินค้า, [5308] ยกเลิกใบสั่งขาย/สั่งจองสินค้า, [5001] Flow ใบสั่งจองสินค้า, [591] ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า, [5304] ปรับปรุงวันที่ส่งของให้ลูกค้า
(หน้าต่างย่อยที่ใช้ร่วมกับใบเสนอราคา — ค้นหาสินค้า, ราคาขาย, ประวัติการขาย, ส่วนลด, รายละเอียดอื่นๆ, หลายภาษี — ไม่เขียนซ้ำ ระบุเฉพาะส่วนที่ต่าง)

หลักฐานอ้างเป็น `ไฟล์:บรรทัด` ของ mirror ที่ถอดรหัสแล้ว (`D:\bccode\scratch\champ-atlas\dec-sa\champ\`) เลขบรรทัดตรงกับไฟล์ต้นฉบับ
ข้อความไทย: ป้ายบนหน้าจอที่มาจาก .rc dialog = ตัวจริง (ผ่าน atlas-02-Bill.md); ป้ายที่มาจาก `theApp.GetMyRes("KEY")` = ไทยจริงอยู่ใน DB `myresource` ไม่อยู่ใน repo — ผมใส่ไทยจาก comment ในโค้ดบรรทัดเดียวกัน (ระบุ "comment") หรือเป็นความหมายอนุมาน (ระบุ "อนุมาน") ห้ามถือเป็นถ้อยคำจริงของ Champ

---

### 0. ภาพรวม 6 รายการ

| รหัสเมนู | ชื่อเมนู (ตรง menuconfig.xml) | menuconfig.xml | resourceid | handler (MainFrm.cpp) | หน้าจอที่เปิดจริง | หมายเหตุ |
|---|---|---|---|---|---|---|
| 5326 | บันทึกใบสั่งขาย/สั่งจองสินค้า | :153 | 30005 | `OnBillsysTrans4` 2441-2448 (ID_BILLSYS_TRANS_4) | `CChmBLFrmSaleOrder` (ChmBLFrmSaleOrder.cpp 4532 บรรทัด; IDD_CHM_BL_FRM_SALEORDER, BC5Account.rc:17176) เป็น "เปลือกแท็บ" ครอบ `CChmBLFrmSaleOrderPage1` (IDD_CHM_BL_FRM_SO_PAGE1 560x280, rc:17190) + `CChmBLFrmSaleOrderPage2` (IDD_CHM_BL_FRM_SO_PAGE2 540x296, rc:17258) + ออบเจกต์คำนวณ `SaleOrder_Info` (ChmBLFrmSaleOrder_InfoEx.cpp) | เส้นทางเริ่มต้น = ชุด `Chm*` (`m_bChmMode = TRUE`, BC5Account.cpp:256); `CBILLFrmSaleOrder`/`CBILLFrmSaleOrder3` รุ่นเก่าไม่ได้ลงทะเบียน (BC5Account.cpp:1587-1599) |
| 5305 | อนุมัติใบสั่งขาย/สั่งจองสินค้า | :154 | 30012 | `OnBillsysTrans12` 2891-2899 | `CBILLFrmApproveSalesOrder` (CMyBaseFormView; IDD_BILL_FRM_APPROVE_SALESORDER 343x236, rc:11361) | อนุมัติเฉพาะใบที่ "พักไว้รออนุมัติ" (HoldingStatus=1) |
| 5308 | ยกเลิกใบสั่งขาย/สั่งจองสินค้า | :155 | 30020 | `OnPOSysCancelPoSo2` 10177-10185 (ID_BILLSYS_TRANS_20) | `CPOFrmCancelPoSoView` (IDD_DIALOG_PO_SO_CANCEL 460x277, rc:390) โหมด m_nMode=1 (SO) | ไม่ใช่การพลิก IsCancel ทั้งใบ — สร้างเอกสาร "ยกเลิกยอดค้าง" BCSOVoid แยกต่างหาก (คลาสเดียวกับยกเลิกใบสั่งซื้อ ID_POSYS_TRANS_25) |
| 5001 | Flow ใบสั่งจองสินค้า | :156 | 30024 | `OnChmSOSRFlow` 2199-2207 (ID_30000_30024) | `CChmFrmSOSRFlow` (ChmFrmSOSRFlow.cpp; IDD_CHM_BILL_FRM_SOSR_FLOW 379x317, rc:17105) | หน้าจอติดตามสายเอกสาร + ปุ่ม "สร้างใบสั่งซื้อ" |
| 591 | ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า | :157 | 30025 | `OnCheckSOSRData` 10739-10748 (ID_30000_30025) | `RunChampMagaDlg("cf_sosrdate_dlg")` → `CConfirmSOSRDateDlg` ใน champ-lib\CHAMPMEGA (IDD_CONFIRM_SOSRDATE 564x250; caption "30025$Confirm SO/SR date") | อยู่นอก mirror dec-sa — อ่านต้นฉบับด้วย lib.read |
| 5304 | ปรับปรุงวันที่ส่งของให้ลูกค้า | :167 | 30011 | `OnBillsysTrans11` 2882-2889 | `CBILLDlgOverhaulTransport` (modal; IDD_BILL_DLG_OVERHAUL_TRANSPORT 329x191, rc:8169) | ชื่อทำให้เข้าใจผิด: ปรับ "ใบขายเชื่อ (BCARInvoice)" ไม่ใช่ใบสั่งขาย |

กลุ่มเมนู (ชื่อหัวกลุ่ม Bill ที่ครอบ 6 รายการนี้) ยังไม่ verify — ดูลำดับจาก menuconfig.xml:153-157, 167 (5304 อยู่ห่างจากอีก 5 รายการ)

ตารางหลักของ Champ ที่เกี่ยวข้อง (ชื่อ SQL Server เป็นหลักฐานเท่านั้น): `BCSaleOrder` (หัว), `BCSaleOrderSub` (รายการ), `BCSOVoid`/`BCSOVoidSub` (ยกเลิกยอดค้าง), `BCQuotation`/`BCQuotationSub` (ต้นทาง), `BCPurchaseOrderSub` (ถูกอ้างอิงด้วย SORefNo), `BCARInvoice`/`BCARInvoiceSub` (ถูกอ้างอิงด้วย SORefNo), `BCARDeposit` + `BCOutputTax` (มัดจำ), `BCRecurTemp`/`BCRecurTempSub` (พักบิล)

---

### 1. [5326] บันทึกใบสั่งขาย/สั่งจองสินค้า

- **ตำแหน่งเมนู**: Bill > (กลุ่มยังไม่ verify) ; menuconfig.xml:153 ; `OnBillsysTrans4` (MainFrm.cpp:2441-2448) ; `CChmBLFrmSaleOrder` + Page1 + Page2 + `SaleOrder_Info` ; config key `SaleOrderConfig`, help key SALEORDER_HELP, authorization ID_BILLSYS_TRANS_4, ตารางเลขที่เอกสาร `BCSaleOrder` (ChmBLFrmSaleOrder.cpp:207) ; รุ่น Legacy ไม่ใช้
- **วัตถุประสงค์**: บันทึกใบสั่งขาย (SOStatus=0) หรือใบสั่งจองสินค้า (SOStatus=1) ให้ลูกค้า — ดึงจากใบเสนอราคาได้ ตรวจวงเงินเครดิต/สต๊อก ทำให้ยอดค้างส่ง (RemainQty) ถูกอ้างอิงต่อไปยังใบสั่งซื้อ/ใบขาย/ยกเลิกยอดค้าง ไม่ลง GL (ไม่มี journal)

#### 1.1 โครงหน้าจอ

| ส่วน | เนื้อหา | หลักฐาน |
|---|---|---|
| เปลือก `CChmBLFrmSaleOrder` | แท็บ 2 หน้า สลับด้วย F5 (`OnHkDocdesc`) ; แถบเครื่องมือมาตรฐาน (New/Save/Del/Find/Prev/Next/Close/Info/Key/Recur) + เมนู "Key" (`OnMore`) | ChmBLFrmSaleOrder.cpp:337, hotkey ~4310-4460 |
| Page1 | หัวเอกสารบรรทัดบน + grid รายการ + ยอดรวมท้ายเอกสาร + กล่องรายละเอียดสินค้า (IDC_ED_ITEMDETAIL แสดง MyDescription ของแถวปัจจุบัน) | ChmBLFrmSaleOrderPage1.cpp:49-64 (DDX), 100-196 (grid) |
| Page2 | ข้อมูลอื่นๆ (พนักงานขาย แผนก โครงการ สกุลเงิน เครดิต วันส่งของ ขนส่ง ที่อยู่ออกบิล Ship to) | ChmBLFrmSaleOrderPage2.cpp:28-62 |

#### 1.2 หัวเอกสาร หน้า 1 (ตามลำดับบนจอ บน→ล่าง ซ้าย→ขวา ตาม atlas-02-Bill.md:342-395)

| # | ป้ายไทย (ตัวจริงจาก .rc) | control / คอลัมน์ DB | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup | การตรวจ/คำนวณเมื่อออกจากช่อง | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | ประเภทการสั่งขาย: | IDC_ED_BILLTYPE (`m_edBillType`) → `BCSaleOrder.BillType` | combo 4 ค่า: 0 สินค้าสด / 1 สินค้าเชื่อ / 2 บริการสด / 3 บริการเชื่อ (ป้ายมาจาก key CB_DOCTYPE_1..4 — ไทยอนุมาน) | ใช่ | จากสิทธิ์ผู้ใช้ (IsBillCash/IsBillCredit) ถ้าไม่มีใช้ `CompanyInfo.DefItemBillType` | — | กระทบการตรวจวงเงิน (BillType 0/2 = เงินสด ไม่ตรวจเครดิต) ; กลุ่มรูปแบบเลขที่ GL format group 8-11 ตาม BillType | Page1.cpp:4002 (OnClearData), 218-262 ; cpp:2050-2062 |
| 2 | ประเภทภาษี: | IDC_ED_TAXTYPE → `TaxType` | combo: 0 แยกนอก (CB_TAXTYPE_1, comment) / 1 รวมใน (comment) / 2 ภาษีอัตราศูนย์ (TAX_ZERO, comment) | ใช่ | `CompanyInfo.DefaultTaxType` | — | เปลี่ยนแล้วคำนวณทุกแถวใหม่ (`DetailCalc`/`BillCalc`) ; ใบเสนอราคาที่ดึงมาต้องมี TaxType ตรงกัน | Page1.cpp:4002 ; cpp:2978-3003 (comment แยกนอก/รวมใน) ; ARFrmARMasterViewPage2.cpp:128-130 |
| 3 | เลขที่ใบสั่งขาย/สั่งจอง: | IDC_ED_DOCNO → `DocNo` | text + เลขรัน | ใช่ | เลขรันอัตโนมัติจากตั้งค่าเลขที่เอกสาร (ตาราง `BCSaleOrder`) | — | ว่าง → CANNOT_SAVE_NOTDOCNO ; ซ้ำ → DOCNO_DUPLICATE (เมื่อ option IsBoxDuplicate) | cpp:433, 439, 1020, 207 |
| 4 | วันที่ออกเอกสาร: | IDC_ED_DOCDATE → `DocDate` | date | ใช่ | วันที่ทำงาน (working date) | — | แก้แล้วคำนวณวันครบกำหนด/วันส่งของใหม่ ; ถ้าย้อนหลังและสต๊อกเข้มงวด (option IsNoStockReword ≠ 1 และ StockStatus ≠ 0) ตรวจสต๊อกย้อนหลัง (FunctionStock/FunctionStockResult) ; งวดปิด → ห้ามบันทึก (IsPeriodClose) | Page1.cpp:3231 ; cpp:421 |
| 5 | รหัสลูกค้า: (+ ชื่อ read-only) | IDC_ED_ARCODE / IDC_ED_ARNAME → `ArCode` | lookup (บาร์โค้ด `*`) | ใช่ | ว่าง | `bcar where activestatus=1` | ออกจากช่อง: `InfoEx.loadAr` เติมอัตราภาษี/ประเภทภาษี, DebtLimit1/DebtLimitBal, SaleCode, DefContactCode, BillAddress (ชื่อ/ที่อยู่ออกบิล), DefDeliveryAddr, เครดิตวัน (billcredit), วันส่งของ (transportcredit), ส่วนลดท้ายบิล, ชื่อ/เบอร์ Ship to ; ไม่พบ → MSG_ARCODE_ERROR | Page1.cpp:3046 ; cpp:447, 452 |
| 6 | ปุ่ม "Ctrl+F3=เลือกใบเสนอราคา" | IDC_BTTN_QUATATION | ปุ่ม (ข้อความตัวจริงจาก .rc) | — | — | CMyMultiSelect config "QuotationTransConfig" | ดู 1.7 | Page1.cpp:2533 |
| 7 | ประเภทเอกสาร: | IDC_ED_DOCTYPE (`m_edDocType_SOorSR`) → `SOStatus` | combo: 0 สั่งขาย (SO) / 1 สั่งจอง (SR) (key PO_SALE/PO_ORDER — ไทยอนุมาน) | ใช่ | `m_Option.nSOSRDefType` | — | SOStatus=1 (สั่งจอง) ข้ามการเตือนสต๊อก ; SOStatus=0 + StockStatus=2 ห้ามขายเกินสต๊อก | Page1.cpp:4002 ; cpp:899 |
| 8 | เงื่อนไขการส่งมอบสินค้า: | IDC_ED_CONDITION (`m_edSendCondition`) → `IsConditionSend` | combo: รับเอง (OWN_RECEIVE, comment) / ส่งให้ (SEND_TO — อนุมาน) | ใช่ | 0 (รับเอง) | — | เป็นค่าตั้งต้นของคอลัมน์ 9 ใน grid | Page1.cpp:4002 ; cpp:1457 |

ยอดรวมท้ายเอกสาร (Page1 ส่วนล่าง) — รายละเอียดสูตรดู 1.5

| # | ป้ายไทย (ตัวจริงจาก .rc) | control → คอลัมน์ DB | ชนิด | แก้ได้ | หลักฐาน |
|---|---|---|---|---|---|
| 9 | รวมมูลค่าสินค้า: | IDC_ED_SUMDETAIL (`m_edSumOfNormalTax`) → `SumOfItemAmount` | ตัวเลข AmountPoint | ได้ (ออกจากช่อง → `BillCalc(1)`) | Page1.cpp:3294 |
| 10 | มูลค่าสินค้ายกเว้นภาษี: | IDC_ED_EXCEPTTAX → `ExceptTaxAmount` | ตัวเลข | อ่านอย่างเดียวตามสูตร | cpp:1300-1330 |
| 11 | ส่วนลด%,บาท: (+ ปุ่ม "%") | IDC_ED_DISC_AMOUNT → `DiscountAmount` ; คำส่วนลด → `DiscountWord` | ตัวเลข + ปุ่มเปิดช่องพิมพ์ % | เมื่อ `bCanChangeBillDiscount && BillDiscount==1` | Page1.cpp:3003 (`OnBnDiscountWord`), 3306 (`OnKillfocusEdDiscAmount` ตั้ง m_nTaxProcessMode=2) |
| 12 | ภาษีมูลค่าเพิ่ม: | IDC_ED_TAXAMOUNT → `TaxAmount` | ตัวเลข | ตามสิทธิ์ (UpdateTaxControl) | Page1.cpp (DDX 49-64) |
| 13 | มูลค่ารวมภาษี: | IDC_ED_SUMTAXAMOUNT → `TotalAmount` | ตัวเลข | อ่านอย่างเดียว | เช่นเดียวกัน |
| 14 | มูลค่าเงินสุทธิ: | IDC_ED_SUMAMOUNT → `NetAmount` | ตัวเลข | อ่านอย่างเดียว | เช่นเดียวกัน |

ค่าที่บันทึกพ่วง: `BeforeTaxAmount` (ยอดถอดภาษีแล้วกรณีภาษีรวมใน), `AfterDiscount`, `ZeroTaxAmount` (ยอดสินค้าภาษีอัตราศูนย์), `TotalTransport` (ค่าขนส่งจาก Page2) — ChmBLFrmSaleOrder.cpp:1283-1300

#### 1.3 หัวเอกสาร หน้า 2 "รายละเอียดอื่นๆ" (F5 สลับมาหน้านี้)

การจับคู่ "ป้ายไทย ↔ ตัวแปร ↔ คอลัมน์" อิงชื่อตัวแปร DDX (Page2.cpp:28-62) + `AddFieldForSave` (cpp:1236-1330) — พิกัดป้ายต่อช่องแต่ละตัวยังไม่ verify 1:1 (control `IDC_ED_REFPONO3..9` ใช้ชื่อ ID ซ้ำกับช่องอื่น จึงจับคู่ด้วยตัวแปรเท่านั้น)

| # | ป้ายไทย (ตัวจริงจาก .rc) | ตัวแปร / control → คอลัมน์ `BCSaleOrder` | ชนิด | lookup | การตรวจ/ผลต่อช่องอื่น |
|---|---|---|---|---|---|
| 1 | รหัสพนักงานขาย: | `m_edSaleCode` (IDC_ED_SALECODE) → `SaleCode` (+ `UserGroup` ดึงจาก BCSale.UserGroup) | lookup | `BCSale` (ActiveStatus<>0) | — |
| 2 | อ้างใบสั่งซื้อเลขที่: | `m_edPORefNo` (IDC_ED_REFPONO) → `RefDocNo` และ `porefno` (เขียนค่าเดียวกันทั้งสองคอลัมน์) | text | — | เลขที่ PO ของลูกค้า |
| 3 | รหัสแผนก: | `m_edDepart` → `DepartCode` | lookup | `BCDepartment` | — |
| 4 | รหัสสกุลเงิน: | `m_edCurrency` → `CurrencyCode` | lookup | `BCCurrency` | `MultiCurrency==0` → ช่องสกุลเงิน/อัตราแลกเปลี่ยนปิด |
| 5 | โครงการ: | `m_edProject` → `ProjectCode` | lookup | `BCProject` | — |
| 6 | อัตราแลกเปลี่ยน: | `m_edExchangeRate` → `ExChangeRate` | ตัวเลข | — | ≤ 0 → บันทึกเป็น 1.0 (cpp:1266-1270) ; มีสกุลเงิน → Home = net × rate |
| 7 | รหัสผู้ติดต่อ: | `m_edContactCode` → `ContactCode` | lookup | `BCContactList` (MyType=0, parentcode = รหัสลูกค้า) | — |
| 8 | อัตราภาษีมูลค่าเพิ่ม: | `m_edTaxRate` → `TaxRate` | ตัวเลข | — | เปิด/ปิดตาม `bCanChangeBillAmount`/`UpdateTaxControl` |
| 9 | การจัดสรร: | `m_edAllocate` → `AllocateCode` | lookup | `BCAllocate` | — |
| 10 | ยอดเงินบาทสุทธิ: | `m_edHomeAmount` (read-only; ปุ่มปิด) | ตัวเลข | — | = NetAmount × อัตราแลกเปลี่ยน (ไม่มีคอลัมน์หัวเอกสารเขียนค่านี้ — เก็บที่ระดับแถว shadow 15) |
| 11 | ส่งมอบภายใน: (วัน) | `m_edDeliveryCredit` (IDC_ED_DELYCREDIT) → `deliveryday` | จำนวนวัน | — | คำนวณ 2 ทาง: วันที่ส่งของ = วันที่เอกสาร + วัน (InfoEx.cpp:510-535) |
| 12 | วันที่ส่งของ: | `m_edDeliveryDate` → `deliverydate` | date | — | แก้วันที่ → คำนวณวันกลับ |
| 13 | เครดิต(วัน): | `m_edCreditDay` → `CreditDay` | จำนวนวัน | — | วันครบกำหนด = วันที่เอกสาร + เครดิต |
| 14 | วันที่ถึงกำหนด: | `m_edDueDate` → `DueDate` | date | — | แก้วันที่ → คำนวณเครดิตกลับ |
| 15 | ค่าขนส่ง: | `m_edTransportAmount` → `TotalTransport` | ตัวเลข | — | ตามสูตร `BillCalc` (Σ shadow28 × qty) |
| 16 | รูปแบบการสั่งซื้อสินค้า: | `m_edOrderChannelCode` → `OrderChannel` | lookup | `BCOrderchannel` | — |
| 17 | Job ID : | `m_edJobID` → `JobID` | text | — | — |
| 18 | ช่องทางการจัดจำหน่าย: | `m_edDistChannelCode` → `DistChannel` | lookup | `BCDistchannel` | — |
| 19 | ชื่อในการออกบิล: | `m_edArBillName` → `ArName` | text | — | เติมจากลูกค้า (BillAddress) |
| 20 | ที่อยู่ในการออกบิล: | `m_edArBillAddress` → `ArAddress` | text หลายบรรทัด | — | เติมจากลูกค้า |
| 21 | ผู้รับสินค้า: | `m_edPersonReceive` (IDC_ED_PERSON) → `OwnReceive` (comment: ชื่อผู้รับสินค้า) | text | — | — |
| 22 | ทะเบียนรถ: (ตัวแรก) | `m_edCarLicense` → `CarLicense` (comment: ทะเบียนรถ) | text | — | — |
| 23 | สถานที่ขนส่ง: | `m_edDelivery` (IDC_ED_TRANSPORT_PLACE) → `DeliveryAddr` | lookup | `BCDeliveryAddr` (กรอง ArCode) | เติมค่าเริ่มต้นจาก DefDeliveryAddr ของลูกค้า |
| 24 | เวลาที่รับ/ส่งสินค้า: | `m_edTransportTime` → `TimeTransport` | text | — | — |
| 25 | กลุ่มเอกสาร: | `m_edDocGroup` (IDC_ED_BILLGROUP) → `BillGroup` | text | — | comment: กลุ่มของบิล/ภาษี |
| 26 | หมายเหตุ: | `m_edMyDescription` (IDC_ED_ITEMDESE) → `MyDescription` | text หลายบรรทัด | — | — |
| 27 | Ship to: | `m_edShipto` (IDC_ED_ITEMDESE2) → `Shipto` | text | — | เติมจากลูกค้า |
| 28 | ที่อยู่(เลขที่ , ตึก , ชื่อถนน): | `m_edTransAddress` → `Transportaddress` | text | — | — |
| 29 | ตำบล/แขวง : | `m_edTSubDistinct` → `SubDistrict` | text | — | — |
| 30 | อำเภอ/เขต : | `m_edTDistinct` → `District` | text | — | — |
| 31 | จังหวัด: | `m_edTProvince` → `Province` | text | — | — |
| 32 | รหัสไปรษณีย์: | `m_edTPostCode` → `PostCode` | text | — | — |
| 33 | ทะเบียนรถ: (ตัวที่สอง — ชุด Ship to) | `m_edTruckLicense` → `TruckLicense` | combo | `BCCarMaster` | — |
| 34 | Ship to Tel : | `m_edShiptoTel` → `ShiptoTel` | text | — | เติมจากลูกค้า |

เขียนเพิ่มโดยไม่มีช่องบนจอ: `ArCode`, `SOStatus`, `IsConditionSend`, `RecurName` (ชื่อพักบิล), `HoldingStatus`, `CrAuthorizeMan`, `SCGBPNumber` (อ่านจาก BCAR.SCGBPNumber), `IsUseRobotSale` (พิมพ์ใบจัดด้วย Robot) — cpp:1236-1330

#### 1.4 ตารางรายการ (grid `m_BInputListWnd`) — คอลัมน์ที่เห็น 10 คอลัมน์ (Page1.cpp:100-148)

ป้ายคอลัมน์มาจาก `GetMyRes(KEY)` → ไทยจริงอยู่ใน `myresource` (ไทยเป็นความหมายอนุมาน)

| # | key (ป้ายอนุมาน) | กว้าง | แก้ได้? | ชนิด | ความหมาย/สูตร | หลักฐาน |
|---|---|---|---|---|---|---|
| 0 | ITEMCODE (รหัสสินค้า) | 60 | ได้ (edit + link, ยาว 25) | lookup | พิมพ์/ค้นสินค้า ; ดูตัวล็อกเมื่อมาจากใบเสนอราคา (option SONoAddnew) | Page1.cpp:113 |
| 1 | ITEMNAME (ชื่อสินค้า) | 120 | เฉพาะ `m_Option.IsEditItemName==1` | text | ชื่อสินค้าตอนขาย | :114 |
| 2 | WAREHOUSE_ITEM (คลัง) | 60 | ได้ | combo ตาราง `BCItemWarehouse` | ไม่ตั้ง → NOT_SET_WHCODE (เมื่อ IsUseWarehouse≠1) ; สิทธิ์คลัง → NO_SELL_FROM_WH | :115 ; cpp:716, 725 |
| 3 | SHELF (ที่เก็บ) | 60 | ได้ | combo ตาราง | ไม่ตั้ง → NOT_SET_SHELFCODE | :116 |
| 4 | UINTCODE (หน่วยนับ) | 60 | ได้ | combo `BCItemUnit` (หรือ `BCStkPacking` เมื่อ unit type 1) | ไม่ตั้ง → NOT_SET_UNITCODE | :117 |
| 5 | ITEMQTY (จำนวน) | 60 | ได้ | ตัวเลข AUTOMONEY2 ชิดขวา ทศนิยม `QtyPoint` | = 0 → QTY_ZERO_LIST | :118 ; cpp:682 |
| 6 | PERPRICE (ราคา/หน่วย) | 60 | เฉพาะ `bCanChangePrice` (และ option SONoEditPrice ล็อกเมื่อมาจากใบเสนอราคา) | ตัวเลข ทศนิยม `PricePoint` | `UpdatePrice` ดึงราคาขายอัตโนมัติ (ข้ามเมื่อแถวมาจากใบเสนอราคา) ; ต่ำกว่าราคามาตรฐาน → เตือน | :119-125 ; Page1.cpp:3527, 4258 |
| 7 | DISCOUNT (ส่วนลด) | 60 | เฉพาะ `DetailDiscount==1` และ `bCanChangeItemDiscount` | text (คำส่วนลด เช่น 5%+3) | ส่วนลดต่อแถว (ดู 1.5) ; option SONoEditDiscount ล็อก | :126-136 |
| 8 | AMOUNT (จำนวนเงิน) | 70 | เฉพาะ `bCanChangePrice` | ตัวเลข ทศนิยม `AmountPoint` | = qty × price − ส่วนลด ; ไม่ตรงสูตร → CHECK_WRONG_AMOUNT ; = 0 → CHECK_AMOUNT_ZERO | :137-142 ; cpp:753, 809, 816 |
| 9 | CONDIT_TRANS (เงื่อนไขขนส่ง) | 40 | ได้ | combo ตามลำดับ (รับเอง/ส่งให้) กว้างข้อความ 15 | → `IsConditionSend` ระดับแถว ; ค่าตั้งต้นจากช่อง "เงื่อนไขการส่งมอบสินค้า" | :143-148 |
| — | ไอคอน "คำสั่งอื่น" ท้ายแถว | — | — | — | เรียก `OnExpand` = Shift+F2 เปิด `CBILLDlgSheetOrder` (กำหนดรายละเอียดสินค้าเพิ่มเติม) | Page1.cpp:3404 |

Shadow columns (ซ่อน 44 ตัว ดัชนี 0-43 — Page1.cpp:149-196) :
`0 RowOrder · 1 ItemType · 2 MyDescription · 3 Project · 4 Allocate · 5 OppositeQty · 6 OppositeUnit · 7 ExceptTax · 8 BarCode · 9 CategoryCode · 10 GroupCode · 11 BrandCode · 12 TypeCode · 13 FormatCode · 14 NetAmount · 15 HomeAmount · 16 OldUnitCode · 17 RefDocNo (= เลขที่ใบเสนอราคาต้นทาง → StkReserveNo) · 18 OldQty · 19 SetCode · 20 DepartCode · 21 CurrencyCode · 22 ExchangeRate · 23 SaleCode · 24 DiscountAmount · 25 OppositePrice2 · 26 StockProcess · 27 QtyBalance · 28 TransportAmount · 29 ColorCode · 30 SizeCode · 31 StyleCode · 32 ItemSetCode · 33 OldItemCode · 34 PORefDocNo · 35 BehindIndex · 36 OldQtyPrice · 37 OldPrice · 38 OldDiscount · 39 SalepriceColumn · 40 prompt_price_flag · 41 price_per_unit · 42 prompt_price_per_unit · 43 OldUnitPrompt`

หมายเหตุ shadow ที่ใช้จริงในสูตร: 7 ExceptTax (ยกเว้นภาษี) ; 14 NetAmount (ถอดภาษีแถว) ; 15 HomeAmount ; 17 เลขที่ใบเสนอราคา ; 28 TransportAmount ; 41 price_per_unit

#### 1.5 สูตรคำนวณ

| รายการ | สูตร (ตามโค้ด) | หลักฐาน |
|---|---|---|
| ต่อแถว `DetailCalc` | DiscountMethod 0: amount = qty × price แล้วหักส่วนลดจากยอด ; DiscountMethod 1: หักส่วนลดจากราคาต่อหน่วยก่อน แล้ว amount = qty × price (ส่วนลดรวม = ส่วนลด × qty) ; ปัดด้วย QtyPoint/PricePoint/AmountPoint | Page1.cpp:435 |
| ภาษีรวมใน (TaxType 1, สินค้าไม่ยกเว้น, rate > 1) | NetAmount(แถว) = amount − amount × rate / (100 + rate) ; HomeAmount = net × อัตราแลกเปลี่ยน | Page1.cpp:435+ |
| shadow 41 | = amount / qty ; Champ ใส่ `× 1.07` แบบ hard-code เมื่อ TaxType 0 (ข้อบกพร่องของ Champ — ห้ามลอก ต้องใช้อัตราจากเอกสาร) | Page1.cpp:429,519 |
| ท้ายบิล `BillCalc` | แยกยอดสินค้าปกติ / ยกเว้นภาษี (shadow 7 หรือ option บริษัท ExceptTax==1) ; ค่าขนส่ง = Σ(shadow28 × qty) ; ส่วนลดท้ายบิลแบ่งตามสัดส่วนยกเว้น/ปกติ ; ภาษีคำนวณผ่าน `BillTaxProcess` (m_strTaxNo="TAXNO") | Page1.cpp:532 ; InfoEx.cpp:442-508 |
| มูลค่าเงินสุทธิ | TaxType 1 (รวมใน): TotalAmount + (ยกเว้น − ส่วนลดยกเว้น) − WTax ; อื่น: (ปกติ − ส่วนลดปกติ) + (ยกเว้น − ส่วนลดยกเว้น) + ภาษี − WTax ; ปัดด้วย AmountPoint ; มีสกุลเงิน → Home = net × rate | InfoEx.cpp (billCalc → calcDiscount → calcTax) |
| วันครบกำหนด / วันส่งของ | due = DocDate + เครดิต ; delivery = DocDate + วันส่งมอบ ; แก้วันที่ → คำนวณวันกลับ | InfoEx.cpp:510-535 |
| มัดจำ (โหมด Chm) | `CalcDeposit` = Σ XDiscount(จำนวนเงินแถว, `BCItem.DepositCondit`) | ChmBLFrmSaleOrder.cpp:3939 |

#### 1.6 ปุ่ม แถบเครื่องมือ คีย์ลัด

| คีย์/ปุ่ม | ฟังก์ชัน | ผล | หลักฐาน |
|---|---|---|---|
| F11 | บันทึก | `ConfirmSaveData` → `OnSaveData` | accelerator IDR_MAINFRAME_CHM BC5Account.rc:977-1008 |
| Ctrl+N / Ctrl+F / Ctrl+P / PgUp / PgDn | ใหม่ / ค้นหา / พิมพ์ / ก่อนหน้า-ถัดไป | มาตรฐาน | เช่นเดียวกัน |
| F5 | รายละเอียดอื่นๆ (comment) | สลับแท็บ Page1 ↔ Page2 | `OnHkDocdesc` |
| F8 | พักบิล (comment) | `OnSaleOrderHoldBill` เก็บใน `BCRecurTemp`/`BCRecurTempSub` MyType 0 | cpp:4083 |
| F9 | เลิกพักบิล (comment) | `OnSelectLoadHoldBill` | cpp:4049 |
| Ctrl+F9 | ยกเลิกรายการพักบิล | `CBILLDlgSaleCancelHold` | ~4140 |
| F12 | ตรวจสอบเอกสาร (comment) | `OnDocumentFlow` เปิด `CMyDocuemtFlow` แสดงเอกสารที่อ้างถึงใบสั่งขายนี้ (union `bcarinvoicesub` / `bcsovoidsub` / `bcpurchaseordersub`) คอลัมน์ DOCNO_DATE, NO_REFERANCE, PRODUCTCODE, ITEMNAME1, REMARK, QUANTITY_ITEM, AMOUNT | cpp:3450 |
| Shift+F2 | กำหนดรายละเอียดสินค้าเพิ่มเติม (comment) | `CBILLDlgSheetOrder` | Page1.cpp:3404 |
| Shift+F5 | Confirm เอกสาร (comment) | `BillConfirm` → IsConfirm | `OnHkConfirmdoc` |
| Ctrl+F5 | ยกเลิกเอกสาร (comment) | `OnHkCanceldoc` ; ถูกบล็อกเมื่อ IsConfirm=1 | cpp:3832 |
| Ctrl+F3 | เลือกใบเสนอราคา | ดู 1.7 | Page1.cpp:2533 |
| Ctrl+F2 | ยอดคงเหลือสินค้าทดแทน (comment) | `OnStockNotEnough` / ItemCheckStock | Page1.cpp:3466 |
| Ctrl+F7 / Ctrl+F8 | Default คลัง,ที่เก็บเริ่มต้น / Default คลังสินค้าอัตโนมัติ (comment) | ใส่คลังทั้งตาราง | `OnMore` cpp:337 |
| Ctrl+F12 | ตรวจสอบการเข้ามาของสินค้า (comment) | ItemInfo | `OnMore` |
| Home | CHECKSTOCK | ตรวจสต๊อก | `OnMore` |

เมนู "Key" (`OnMore` cpp:337) มี 15 รายการ: F5_EDESCRIPTION, F8_HOLDBILL, F9_SELECTHOLDBILL, CTRL_F9_CANCELHOLD, F11_SAVE, F12_CHECK_DOCUMENT, SHIFT_F2_ITEM, SHIFT_F5, HOME_CHECKSTOCK, CTRL_F3_QUOTATION, CTRL_F5, CTRL_7, CTRL_8, CTRL_F12 — ID_HK_HOLD_BILL (5961) และ ID_HK_SHOW_HOLDBILL (5962) ไม่มี accelerator ใช้ได้จากเมนู Key เท่านั้น (resource.h:4726, 4728)

#### 1.7 ดึงจากใบเสนอราคา (Ctrl+F3) — ต่างจากใบเสนอราคาเอง

| ประเด็น | พฤติกรรม | หลักฐาน |
|---|---|---|
| เงื่อนไขก่อนกด | ต้องมีรหัสลูกค้า ; ถ้าไม่มีลูกค้าแต่มีรหัสพนักงานขาย → `OnGetQuotationBySale` ; ไม่มีทั้งสองอย่าง → ข้อความ SALEORDER_ARCODE ("โปรดระบุรหัสลูกค้า" — comment) | Page1.cpp:2533, 2820 |
| ตัวกรองรายการใบเสนอราคา | `BCQuotation.ArCode=…, BillStatus=0, AssertStatus=1, IsCancel=0, IsConfirm=1, TaxType=…, billtype=…, DocNo Not in (ที่ถูกอ้างแล้ว)` (+ SaleCode ถ้ามี) | Page1.cpp:2533+ |
| หลังดึง | ช่องรหัสลูกค้าล็อก ; ส่วนหัวคัดลอกจากใบเสนอราคาใบแรก (SaleCode, DepartCode, ContactCode, CreditDay, DueDate, DeliveryDay, DeliveryDate, ExchangeRate, CurrencyCode, Project, Allocate, DiscountWord, DiscountAmount, TotalTransport, DistChannel, OrderChannel) | เช่นเดียวกัน |
| รายการ | `bcquotationsub where transstate=0 and iscancel=0 and (remainqty>0 or itemcode='.')` ; จำนวน = RemainQty (จำนวนเงินคำนวณใหม่เมื่อ qty ≠ remain) ; shadow 17 = เลขที่ใบเสนอราคา → บันทึกเป็น `StkReserveNo` | เช่นเดียวกัน ; cpp:1457+ |
| ตัวล็อกตาม option บริษัท | SONoEditPrice ล็อกคอลัมน์ 6 ; SONoEditDiscount ล็อกคอลัมน์ 7 ; SONoAddnew ล็อกคอลัมน์ 0 และปิดเพิ่มแถวใหม่ ; SONoEditQty==1 → จำนวนสั่งขายห้ามเกิน RemainQty ของใบเสนอราคา (เฉพาะ SOStatus 0, StockType≠1) | Page1.cpp:4151 (`SaleOrderMuchQuotations` ข้อความ SALEORDER_MUCH_QUOTATION, QTY_LIST, BILL_QUOTATION_NO, QUOTATION_BALANCE) |

#### 1.8 สถานะเอกสารและ workflow

| คอลัมน์ | ค่า | ความหมาย |
|---|---|---|
| `IsCompleteSave` | 0 ระหว่างบันทึก → 1 เมื่อบันทึกรายการครบ | ถ้าค้างเป็น 0 (บันทึกขาด) เปิดโหลดแล้วถามเอาข้อมูลเก่า |
| `HoldingStatus` | 1 = พักรออนุมัติ (เกินวงเงิน) ; 0 = ปกติ | แถวรายการ `BCSaleOrderSub.HoldingStatus` ตามหัว |
| `IsConfirm` | 0 ยังไม่ยืนยัน ; 1 ยืนยัน/อนุมัติ (Shift+F5 หรือหน้าอนุมัติ 5305) ; 2 = อนุมัติเกินวงเงินตอนบันทึก (เขียนเฉพาะค่า 2) | IsConfirm ≠ 0 → แก้ไม่ได้ (NOT_SAVE_ISCONFIRM) ลบไม่ได้ (NOT_DELETE_ISCONFIRM) |
| `IsCancel` | 1 ยกเลิกทั้งใบ (Ctrl+F5) | แก้ไม่ได้ (NOT_SAVE_ISCANCEL) |
| `BillStatus` | 0 ยังไม่ทำจองหมด ; 1 ทำจองหมดแล้ว (comment) | ถูกอ้างอิงครบ → CANNOT_CANCEL |
| `RemainQty` / `POQty` (แถว) | ตั้ง = Qty ตอนบันทึก | RemainQty = ยอดค้างส่ง (ลดเมื่อขาย/ยกเลิกยอดค้าง) ; POQty = ยอดที่ทำใบสั่งซื้อได้ ([5001]) |
| `IsProcess` (แถว) | 0 ตอนบันทึก | ใช้กับ stock reprocess |
| `IsPostGL` | ไม่ได้ตั้งโดยแบบฟอร์มนี้ | เป็น guard เท่านั้น: โพสต์แล้วห้ามลบ/ยกเลิก (CANNOT_DELETE_POSGL / CANNOT_CANCEL) |

ลำดับ: บันทึก (IsCompleteSave 0→1) → [ถ้าเกินวงเงิน: พัก HoldingStatus=1 รอหน้าอนุมัติ 5305 หรืออนุมัติทันที IsConfirm=2] → Confirm (Shift+F5) → ถูกอ้างอิงโดย ใบสั่งซื้อ (SORefNo) / ใบขาย (BCARInvoiceSub.SORefNo, RefType=1) / ยกเลิกยอดค้าง (BCSOVoidSub.SONo)

#### 1.9 ผลข้างเคียงของการบันทึก (`OnSaveData` ChmBLFrmSaleOrder.cpp:1219-1675 ทั้งหมดอยู่ใน `StartTransaction`)

| ลำดับ | การกระทำ | รายละเอียด |
|---|---|---|
| 1 | เขียนหัว `BCSaleOrder` ด้วย `IsCompleteSave=0` | คอลัมน์: TaxType, BillType, DocNo, DocDate, ArCode, SOStatus, IsConditionSend, DepartCode, DueDate, CreditDay, deliverydate, deliveryday, UserGroup, SaleCode, TaxRate, RefDocNo, ContactCode, DeliveryAddr, TimeTransport, AllocateCode, ProjectCode, MyDescription, OwnReceive, CarLicense, porefno, BillGroup, CurrencyCode, ZeroTaxAmount, CrAuthorizeMan, IsConfirm (เฉพาะ =2), ExChangeRate, RecurName, HoldingStatus, ArName, ArAddress, SumOfItemAmount, DiscountWord, DiscountAmount, BeforeTaxAmount, AfterDiscount, TaxAmount, TotalAmount, ExceptTaxAmount, NetAmount, TotalTransport, JobID, DistChannel, OrderChannel, SCGBPNumber, IsUseRobotSale, Transportaddress, SubDistrict, District, Province, PostCode, TruckLicense, Shipto, ShiptoTel |
| 1b | เพิ่มเอกสารใหม่ | + BillStatus=0, CreatorCode (10 ตัวอักษรแรกของผู้ใช้), CreateDateTime, PrintCount=0 |
| 1c | แก้เอกสารเดิม | + LastEditorCode, LastEditDateT ; ลบรายการเดิมแล้วเขียนใหม่ |
| 2 | เขียนรายการ `BCSaleOrderSub` ทีละแถว | DocNo, TaxType, SOStatus, DocDate, ArCode, ItemCode, ItemName, WHCode, ShelfCode, UnitCode, Qty, Price, DiscountWord, Amount, RemainQty=Qty, POQty=Qty, HoldingStatus, IsConditionSend (คอลัมน์ 9), DiscountAmount, ItemType, MyDescription, ProjectCode, AllocateCode, OppositeQty, OppositeUnit, ExceptTax, BarCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, NetAmount, HomeAmount, StkReserveNo (= shadow 17 เลขที่ใบเสนอราคา), DepartCode, CurrencyCode, TransportAmount, ColorCode, SizeCode, StyleCode, ItemSetCode, PORefNo, TaxRate, PackingRate1/2, behindindex, IsUseRobotSale, OppositePrice2, ExChangeRate, UserGroup, SaleCode, RefType=1 (เขียนทุกแถว ; comment บอก 0=ใบสั่งจอง 1=ใบเสนอราคา), IsProcess=0, LineNumber, price_per_unit, prompt_price_flag, prompt_price_per_unit |
| 3 | `UPDATE BCSaleOrder SET IsCompleteSave=1` | หลังรายการครบ ; แถวใดล้ม → ข้อความบันทึกไม่สมบูรณ์ |
| 4 | `UpdateDocumentFlow(+1, 14, รายการเลขที่ใบเสนอราคา)` | บันทึกสายเอกสาร (type 14 = ใบสั่งขาย) |
| 5 | `OnStockReprocess` ทุกสินค้าในเอกสาร | คำนวณยอดจอง/ค้างส่งใหม่ — สิ่งที่ reprocess เขียนลงตารางสต๊อกยังไม่ verify |
| 6 | มัดจำ (เมื่อมีการรับมัดจำ) | เขียน `BCARDeposit` (RefNo = เลขที่ใบสั่งขาย) + `BCOutputTax` (ChmBLFrmSaleOrder.cpp ~1598-1650) ผ่าน `CBILLDlgDeposit` ; เงื่อนไขเรียกรับมัดจำตรวจ `m_nBillType == 1` + โหมดเพิ่ม (comment บอก "สั่งจองสินค้า" แต่โค้ดไม่ได้ดู SOStatus) |
| 7 | `AddUserLogBook` | ลงบันทึกการใช้งาน |

ลบ (`OnDeleteData` cpp:1154): ลบ `BCSaleOrder` + `BCSaleOrderSub` → `OnStockReprocess` ทุกสินค้า → `AddUserLogBookDelete` (SUBTRACT_SALES_ORDER) → `UpdateDocumentFlow(-1, 14, …)`
ยกเลิก (`BillCancelUpdateData` cpp:3513): `UpdateDocumentFlow(-1,14)` → `UPDATE BCSaleOrder SET IsCancel=1` → `BCSaleOrderSub SET IsCancel=1, IsProcess=0` → stock reprocess → `AddUserLogBookCancel` (CANCEL_SALES_ORDER)

#### 1.10 Guard / ข้อความเตือน (key → เงื่อนไข) — `ConfirmSaveData` cpp:401-~1095

| ลำดับ | key ข้อความ | เงื่อนไข | บรรทัด |
|---|---|---|---|
| 1 | NOT_SAVE_FORMITEM | แบบฟอร์มสินค้าไม่ถูกต้อง | 411 |
| 2 | (IsPeriodClose) | วันที่เอกสารอยู่ในงวดที่ปิดแล้ว | 421 |
| 3 | CANNOT_SAVE_NOTDOCNO / DOCNO_DUPLICATE | ไม่มีเลขที่ / เลขที่ซ้ำ | 433, 439, 1020 |
| 4 | CANNOT_SAVE_NOTARCODE / MSG_ARCODE_ERROR | ไม่ใส่/ไม่พบรหัสลูกค้า | 447, 452 |
| 5 | NOT_SAVE_ISCANCEL / NOT_SAVE_ISCONFIRM | โหมดแก้ไข: ถูกยกเลิก/ยืนยันแล้ว | 465, 470 |
| 6 | CANNOT_SAVE_BILLSTATUS_REF | โหมดแก้ไข: ถูกอ้างอิงโดย BCARInvoiceSub / BCDeliveryOrderSub / BCSOVoidSub / BCPurchaseOrderSub | 487, 506, 574, 597 (บรรทัด 581 ใช้สตริงดิบไม่ผ่าน GetMyRes) |
| 7 | CANNOT_SAVE + CREDIT_ARCODE_OVER + BILL_BALANCE + CANCEL_ARCREDIT | BillType ไม่ใช่ 0/2, net > 0, `DebtOverWarning==3`, DebtLimit1 > 0 และ net > DebtLimitBal → หยุดบันทึกเด็ดขาด | 517-526 |
| 8 | `ArCreditExceed(...)` | `SODebtWarning==1 && DebtOverWarning>1` และ (HoldingStatus≠1 หรือเปลี่ยนลูกค้าหรือหยุดบิล) → ถาม 3 ทาง (ดู 1.11) | 528-565 |
| 9 | CANNOT_SAVE_NOTITEM | ไม่มีรายการสินค้า | 605 |
| 10 | ตรวจต่อแถว: สต๊อก (StockStatus≠0 หรือ StockProcess==1), QTY_ZERO_LIST, ไม่พบสินค้า (NOT_DATA), NOT_SET_WHCODE, NO_SELL_FROM_WH, NOT_SET_SHELFCODE, NOT_SET_UNITCODE, CHECK_AMOUNT_ZERO, CHECK_WRONG_AMOUNT (amount ≠ qty × price − ส่วนลด), CANNOT_SALE_LOWER_PRICE | ~639-844 |
| 11 | CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT | ผลรวมหัว ≠ ผลรวมรายการ | 880 |
| 12 | ตรวจมัดจำ | BillType==1 + เพิ่มใหม่ | 884-895 |
| 13 | ห้ามขายเกินสต๊อก | `StockStatus==2 && SOStatus==0` (SR ข้าม) | 899 |
| 14 | `GetPromptPlusMessage(0)` | ราคา prompt | 1006 |
| 15 | WARNING_TEXT_SALELOWER_SALEPRICE | ราคาต่ำกว่าราคามาตรฐาน เมื่อ option `WarningLowerStdPrice` (ตาราง bcconfigurations) | Page1.cpp:4258 |

ลบ (`ConfirmDeleteData` cpp:1097): CANNOT_DELETE_DOCNO, ปิดงวด, IsPostGL → CANNOT_DELETE_POSGL, NOT_DELETE_ISCANCEL, NOT_DELETE_ISCONFIRM, CANNOT_DELETE_BILLSTATUS_REF, DELETE_MSG (ใช่/ไม่)
ยกเลิก (`OnConfirmCancel` cpp:1800): IsPostGL → CANNOT_CANCEL ; IsCancel อยู่แล้ว → MSG_CANNOT_CANCEL ; IsConfirm=1 หรือ BillStatus=1 → CANNOT_CANCEL ; ถูกอ้างโดยใบขาย (RefType=1)/SOVoid/ใบสั่งซื้อ → CANNOT_CANCEL_USE

#### 1.11 ตรวจวงเงินเครดิต (ArCreditExceed)

| ทางเลือก | `nConfirmStatus` | ผล |
|---|---|---|
| ไม่ (No) | 0 | โหมดแก้ไข: ตั้ง `IsCancel=1` ทั้งหัวและรายการ แล้วเคลียร์/ยกเลิกการบันทึก ; โหมดเพิ่ม: ยกเลิกการบันทึก |
| อนุมัติเดี๋ยวนี้ | 1 | `IsConfirm=2` ผ่านการบันทึกต่อ |
| พัก (Holding) | 2 | `HoldingStatus=1` + เก็บ `CrAuthorizeMan` (ผู้อนุมัติ) → ไปรออนุมัติที่ [5305] |

หมายเหตุข้อบกพร่อง: การเตือนใน `BillCalc` (Page1.cpp ~725) เทียบกับ `dNetAmount` ซึ่งเป็น 0.0 ตลอด → ไม่เคยเตือนตอนคีย์ ; ตัวตรวจจริงอยู่ที่ `ConfirmSaveData` เท่านั้น

#### 1.12 ข้อสังเกตความผิดปกติของ Champ (ไม่ต้องลอกตามไปใน BC)

| # | ข้อสังเกต | หลักฐาน |
|---|---|---|
| 1 | `price_per_unit` (shadow 41) คูณ 1.07 แบบ hard-code เมื่อ TaxType 0 | Page1.cpp:429,519 |
| 2 | `OnInitDialog` ใช้ `nSOSRDefType` เป็นประเภทบิล แต่ `OnClearData` ใช้ `DefItemBillType` | Page1.cpp:100+, 4002 |
| 3 | เตือนวงเงินใน `BillCalc` ใช้ตัวแปรที่เป็น 0 เสมอ | Page1.cpp ~725 |
| 4 | เงื่อนไขมัดจำดู BillType==1 ไม่ใช่ SOStatus (comment ขัดกับโค้ด) | cpp:884-895 |
| 5 | `RefType=1` เขียนทุกแถวแม้ไม่ได้มาจากใบเสนอราคา | cpp ~1457+ |
| 6 | อนุมัติผ่านหน้า 5305 ตั้ง `IsConfirm=1` แต่อนุมัติเกินวงเงินตอนบันทึกตั้ง `IsConfirm=2` → ค่าคนละแบบ และ IsConfirm=1 ทำให้แก้ไม่ได้ | BILLFrmApproveSalesOrder.cpp:404-454 ; cpp:1264 |
| 7 | `OnKillfocusEdTaxamount`, `OnGetItemMatch`, `OnGetItemMatchSO` ถูก comment ทิ้ง | Page1.cpp |

---

### 2. [5305] อนุมัติใบสั่งขาย/สั่งจองสินค้า

- **ตำแหน่งเมนู**: Bill > (กลุ่มยังไม่ verify) ; menuconfig.xml:154 ; `OnBillsysTrans12` (MainFrm.cpp:2891-2899) ; `CBILLFrmApproveSalesOrder` (BILLFrmApproveSalesOrder.cpp ; view CMyBaseFormView ; IDD_BILL_FRM_APPROVE_SALESORDER 343x236, rc:11361) ; title key 30012
- **วัตถุประสงค์**: ผู้มีสิทธิ์อนุมัติใบสั่งขายที่ถูก "พัก" เพราะเกินวงเงินลูกค้า (HoldingStatus=1) เป็นชุดหลายใบ

##### หน้าจอและช่องเงื่อนไข

| # | ป้ายไทย (ตัวจริงจาก .rc) | control / คอลัมน์ | ชนิด | บังคับ? | lookup | หมายเหตุ |
|---|---|---|---|---|---|---|
| 1 | เลขที่ใบสั่งขาย: | จาก–ถึง | lookup (link) | ไม่ | ใบสั่งขายที่ `IsCompleteSave=1 AND IsCancel=0 AND HoldingStatus<>0` | ใส่แต่ช่องเริ่ม = เท่ากับ (ไม่ใช่ช่วง) |
| 2 | วันที่ใบสั่งขาย: | จาก–ถึง | date | ไม่ | — | — |
| 3 | รหัสลูกค้า: | — | lookup (link) | ไม่ | `bcar` | — |
| 4 | รหัสแผนก: | — | combo | ไม่ | `BCDepartment` | — |
| 5 | รหัสพนักงาน: | — | combo | ไม่ | `BCSale` | — |
| 6 | Select All / เลือกทั้งหมด | checkbox | — | — | — | ติ๊กทุกแถว |

ปุ่ม: "ประมวลผลเงื่อนไข" (F5) และ "อนุมัติ" (F11)

##### รายการ (ListView มี checkbox) — คอลัมน์ตามลำดับ

| # | key | ป้าย | ความหมาย |
|---|---|---|---|
| 0 | S | (ช่องเลือก) | ค่าเริ่มต้น "ไม่ติ๊ก" |
| 1 | DOCNO_SALE | เลขที่ใบสั่งขาย: | `DocNo` |
| 2 | DOCDATE | วันที่ (อนุมาน) | `DocDate` |
| 3 | ARCODE | รหัสลูกค้า (อนุมาน) | `ArCode` |
| 4 | ARNAME1 | ชื่อลูกค้า (อนุมาน) | ชื่อจาก bcar |
| 5 | TOTAL_NET | ยอดรวมเงินสุทธิ | `NetAmount` |

ดับเบิลคลิก = เปิดหน้าใบสั่งขาย (ID_BILLSYS_TRANS_4, filter DOCNO) (cpp 305-350)

##### SQL ค้นหา (BILLFrmApproveSalesOrder.cpp:305-350)
`SELECT DocNo,ArCode,DocDate,HoldingStatus,NetAmount FROM BCSaleOrder WHERE IsCancel=0 AND IsCompleteSave=1 AND HoldingStatus=1 [+ ตัวกรอง] ORDER BY DocNo, ArCode`

##### ผลข้างเคียง (อนุมัติ cpp:404-454)
ต่อแถวที่ติ๊ก: `UpdateDB BCSaleOrder SET HoldingStatus=0, ApproveCode=ผู้ใช้ปัจจุบัน, ApproveDateTime=ตอนนี้, IsConfirm=1` แล้ว `UPDATE BCSaleOrderSub SET HoldingStatus=0` ; ข้อความ APPROVE_SALESORDER_OK (comment: ได้ทำการอนุมัติใบสั่งขายเรียบร้อยแล้ว) / NOT_PARTICULAR (comment: ไม่มีรายการที่ต้องการอนุมัติ) ; `AddUserLogBook(30012, ID_BILLSYS_TRANS_12)`

##### Guard
ไม่มีการตรวจวงเงินซ้ำ ไม่ตรวจงวดปิด และไม่ตรวจ IsPostGL ในหน้านี้ (ไม่พบในโค้ด) — การตรวจวงเงินทำที่ตอนบันทึก (1.11) ; อนุมัติแล้ว IsConfirm=1 ทำให้ใบนั้นแก้ไขไม่ได้ต่อ (NOT_SAVE_ISCONFIRM)

---

### 3. [5308] ยกเลิกใบสั่งขาย/สั่งจองสินค้า

- **ตำแหน่งเมนู**: Bill > (กลุ่มยังไม่ verify) ; menuconfig.xml:155 ; `OnPOSysCancelPoSo2` (MainFrm.cpp:10177-10185) ; `CPOFrmCancelPoSoView` (POFrmCancelPoSoView.cpp ; CMyEspeciallyFormView ; IDD_DIALOG_PO_SO_CANCEL 460x277, rc:390) ; โหมด SO (m_nMode=1): title key 30020, config `CancelSoConfig`, help CANCEL_SO_HELP, authorization ID_BILLSYS_TRANS_20, ตารางเอกสาร `BCSOVoid`/`BCSOVoidSub` (โหมด PO ใช้เมนู ID_POSYS_TRANS_25 กับคลาสเดียวกัน)
- **วัตถุประสงค์**: **ยกเลิกยอดค้างส่งรายแถว** โดยสร้างเอกสารยกเลิก (BCSOVoid) แยกต่างหากอ้างใบสั่งขาย — ไม่ใช่การตั้ง IsCancel ทั้งใบ (นั่นคือ Ctrl+F5 ในหน้า 5326)

##### หน้าจอ — หัวเอกสาร

| # | ป้ายไทย | control → คอลัมน์ | ชนิด | บังคับ? | ค่าเริ่มต้น/กติกา |
|---|---|---|---|---|---|
| 1 | เลขที่เอกสาร: | DOCNO$ → `BCSOVoid.DocNo` | text + เลขรัน | ใช่ | ล็อกถ้า `!bCanChangeBillNo` |
| 2 | วันที่เอกสาร (ป้ายใน .rc ยืมจาก "วันที่วางบิล" ไม่ใช่ป้ายจริง) | DOCDATE$ → `DocDate` | date | ใช่ | ปิดแก้ถ้า `IsCanEditDocDate==0` และผู้ใช้ไม่ใช่ sa |
| 3 | หมายเหตุ: | REMARK$ → `Remark` | text หลายบรรทัด | ไม่ | — |
| 4 | (เหตุผลการยกเลิก) | REASON$ → `VoidReasonCode` | combo `BCCancelReason` (50 แถวบน) | ไม่ | — |
| 5 | (รวมมูลค่าที่ยกเลิก) | SUM_OF_ITMEAMOUNT$ → `TotalAmount` | ตัวเลข read-only | — | = Σ voidqty × price |

##### หน้าจอ — กลุ่มเงื่อนไขค้นหา

| # | ป้าย | ชนิด |
|---|---|---|
| 1 | เลขที่เอกสาร: จาก–ถึง (GL_FROM_DOCNO$) | lookup |
| 2 | รหัสลูกค้า จาก–ถึง (FROM_CODE$) | lookup |
| 3 | จากวันที่: จาก–ถึง (FROM_DATE1$) | date |
| 4 | รหัสพนักงานขาย จาก–ถึง (FROM_SALE_CODE$) | combo |
| 5 | ปุ่ม F5 ประมวลผล (F5_PROCESS_CONDIT$) ; checkbox เลือกทั้งหมด (SELECT_ALL$) | — |

ป้ายภาษาไทยของ DOCNO$/REMARK$/REASON$ ยังไม่ verify ครบ — ดู atlas-02-Bill.md:801-814

SQL ประมวลผล (cpp:404-537): `SELECT * FROM BCSaleOrderSub WHERE RemainQty>0 AND DocNo Not In (Select DocNo From BCSaleOrder WHERE IsCancel=1)` + เงื่อนไขเลขที่/วันที่/ArCode + `ORDER BY DocNo,ItemCode` — ข้อบกพร่อง Champ: ตัวกรองรหัสพนักงานขายถูกใช้กับ ArCode และกรณีค่าเดียวใช้ `strStartCode` ; หน้านี้ไม่กรอง IsConfirm/IsCompleteSave

##### grid (cpp:169-186)

| # | ป้าย (hard-code ไทยในโค้ด) | ความหมาย | แก้ได้ |
|---|---|---|---|
| 0 | C | เลือก | ได้ |
| 1 | เลขที่เอกสาร | `DocNo` ใบสั่งขาย | ไม่ |
| 2 | วันที่เอกสาร | `DocDate` | ไม่ |
| 3 | วันที่ครบกำหนด | SODueDate | ไม่ |
| 4 | รหัสสินค้า | `ItemCode` | ไม่ |
| 5 | ชื่อสินค้า | `ItemName` | ไม่ |
| 6 | หน่วยนับ (รหัส/ชื่อ) | `UnitCode` | ไม่ |
| 7 | มูลค่า | `Amount` | ไม่ |
| 8 | จำนวนที่สั่งขาย | `Qty` | ไม่ |
| 9 | จำนวนที่ยกเลิก | `VoidQty` (แสดงค่าตั้งต้น = ยอดค้าง) | **ได้** |
| 10 | A | ยกเลิกทั้งแถว (`CancelAll`) | ได้ |

Shadow 0-17: RowOrder, ApCode(ArCode), DepartCode, AllocateCode, ProjectCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, Price, WHCode, ShelfCode, RefLineNumber, UnitCode, BehindIndex, QtyBalance, VoidAmount

##### สถานะ/workflow
เลือกแถว → กำหนดจำนวนยกเลิก (≤ ยอดค้าง) → บันทึก → เกิดเอกสาร BCSOVoid ; ใบสั่งขายต้นทางไม่เปลี่ยนสถานะที่หัว แต่ RemainQty ที่แถวลดลงผ่าน stock reprocess ; แก้ไขเอกสารได้จนกว่าถูกยืนยัน (CANNOT_UPDATE_ISCONFIRM) ; ลบ = ลบ BCSOVoid/BCSOVoidSub + คืนยอด

##### ผลข้างเคียงของการบันทึก (cpp:539-808)
1. ต้องเลือกอย่างน้อย 1 แถว ไม่เช่นนั้นข้อความ "Please select transaction for cancel" (ข้อความตัวอักษรในโค้ด)
2. หัว `BCSOVoid` (IsCompleteSave=0 ก่อน): DocNo, DocDate, TotalAmount, Remark, VoidReasonCode, CreatorCode/CreateDateTime หรือ LastEditorCode/LastEditDateT
3. ลบ `BCSOVoidSub` เก่า แล้วเขียนใหม่: DocNo, DocDate, DepartCode, AllocateCode, ProjectCode, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode, Price, WHCode, ShelfCode, RefLineNumber, UnitCode, ArCode, SONo, SODate, SODueDate, ItemCode, MyDescription, Amount, Qty, VoidQty, VoidAmount (= price × voidqty), CancelAll, TaxRate, PackingRate1/2, behindindex, LineNumber, IsProcess
4. `OnStockReprocess` ทุกสินค้า → `IsCompleteSave=1` → `UpdateDocumentFlow(+1, 4, เลขที่ใบสั่งขาย)` (type 4 = ยกเลิกใบสั่งขาย) ; รายการแถวล้ม → SAVEDETAIL_NOTCOMPLETE ; `AddUserLogBook("บันทึกการยกเลิกใบสั่งขาย", IDD_DIALOG_PO_SO_CANCEL)`
5. ลบ (cpp:982-1060): `UpdateRefDoc` + ลบ BCSOVoid/BCSOVoidSub → stock reprocess → `UpdateDocumentFlow(-1, 4)`

##### Guard
| key | เงื่อนไข | บรรทัด |
|---|---|---|
| QTY_CANCEL_NOT_OVER + จำนวน + SET_AGAIN | จำนวนยกเลิกเกินยอดค้าง → รีเซ็ตเป็นยอดค้าง | grid edit |
| CANNOT_SAVE_NOTDOCNO / DOCNO_DUPLICATE (ตรวจ BCSOVoidSub) / IsPeriodClose | บันทึก | 1083 |
| CANNOT_DELETE_DOCNO / IsPeriodClose / CONFIRM_DELETEDATA (ใช่/ไม่) | ลบ | 1120 |
| CANNOT_UPDATE_ISCONFIRM | แก้ไขเอกสารที่ยืนยันแล้ว | 539+ |
| CANNOT_SAVE_DETAIL_WANT_OLDITEM | โหลดเอกสารที่ IsCompleteSave≠1 | 811-931 |

คีย์ลัด: Ctrl+N ใหม่, Ctrl+F ค้นหา, Ctrl+P พิมพ์, F5 = ประมวลผลเงื่อนไข (ID_HK_DOCDESC → `OnBttnProcess`)

---

### 4. [5001] Flow ใบสั่งจองสินค้า

- **ตำแหน่งเมนู**: Bill > (กลุ่มยังไม่ verify) ; menuconfig.xml:156 ; `OnChmSOSRFlow` (MainFrm.cpp:2199-2207) ; `CChmFrmSOSRFlow` (ChmFrmSOSRFlow.cpp ; CMyBaseFormView ; IDD_CHM_BILL_FRM_SOSR_FLOW 379x317, rc:17105) ; title key 30024
- **วัตถุประสงค์**: ติดตามสายเอกสาร ใบสั่งขาย/สั่งจอง → ใบสั่งซื้อ → ใบรับสินค้า (GP/GR) → ใบขาย → ใบยกเลิกยอดค้าง และสร้างใบสั่งซื้อจากใบสั่งขายได้

##### หน้าจอ

| # | ป้ายไทย | ชนิด | หมายเหตุ |
|---|---|---|---|
| 1 | วันที่ใบสั่งขาย: จาก–ถึง | date | — |
| 2 | เลขที่ใบสั่งขาย: จาก–ถึง | lookup (link) | ช่องเริ่ม: `IsCompleteSave=1 AND IsCancel=0 AND HoldingStatus<>1` ; ช่องสิ้นสุดไม่กรอง HoldingStatus |
| 3 | ปุ่ม "ค้นหา" (F5) | — | `OnBttnSearchData` cpp:522 |
| 4 | ปุ่ม "สร้างใบสั่งซื้อ" (F11) | — | hotkey ใน `PreTranslateMessage` cpp:1035 |
| 5 | ปุ่ม "คำนวณยอดสั่งจองที่สามารถทำ PO ได้" | — | ข้อความตัวจริงในโค้ด |
| 6 | "Current SO No. :" | label | แสดงใบสั่งขายที่เลือก |

รายการซ้าย (`BuildColumns` cpp:114): No., PO_ORDER (เลขที่ใบสั่งขาย), POORDER (เลขที่ใบสั่งซื้อ), GPGR_DOCNO (เลขที่ GP/GR), GOODSSALES (เลขที่ใบขาย), NO_DOC + IDCANCEL (เลขที่เอกสารยกเลิก), ซ่อน: BillStatus, GRIRBillStatus, POStatus, SORefNo
รายการขวา (`BuildColumns2` cpp:131): No., PRODUCTCODE, ITEMNAME1, UNITCODE, ITEM_SALEORDER (จำนวนสั่งขาย), ORDER_BUY_QTY (จำนวนสั่งซื้อ), ITEM_GPGRQTY (ยอด GP), ITEM_SALESQTY (ยอดขาย), QTY_CANCEL (ยกเลิกสั่งขาย)
(ป้ายคอลัมน์เป็น key GetMyRes — ไทยอนุมานจากชื่อ key)

##### ตรรกะค้นหา (cpp:522)
`FROM BCSaleOrderSub A left join BCPurchaseOrderSub B on A.DocNo=B.SORefNo left join BCARInvoiceSub C on A.DocNo=C.SORefNo left join BCSOVoidSub D on A.DocNo=D.SONo WHERE A.IsCancel=0 AND A.TransState=0 AND A.ItemType<>1` (+ตัวกรอง) → จัดกลุ่มต่อใบสั่งขาย ขยายแถวตาม PO/ใบขาย/ยกเลิก เรียง SO, PO, Bill, Cancel ; ตัวอักษรสีน้ำเงินเมื่อ RemainQty==0 หรือ POQty==0 (ถูกอ้างอิงหมดแล้ว) ; คลิกแถว → โหลดรายการขวา พร้อม BLQty (Σ BCArInvoiceSub.Qty ตาม SORefNo + RefLineNumber), SOCancelQty (Σ BCSOVoidSub.VoidQty), GetPOGPQty (Σ BCPurchaseOrderSub.Qty ; GP = BCApInvoiceSub.Qty ตาม PORefNo)

##### เมนูคลิกขวา (ข้อความตัวจริงในโค้ด)

| รายการ | ผล |
|---|---|
| สร้างใบสั่งซื้อสินค้า | เปิด ID_POSYS_TRANS_4 พร้อม `LOADFROMSO=<docno>` หลังผ่าน `OnReferencePO` (ต้อง Σ RemainQty≠0 และ Σ POQty≠0 ; ไม่ผ่าน → "เอกสาร ถูกอ้างอิงไปทำสั่งซื้อ หรือขาย หมดแล้ว !!!") |
| เปิดหน้าจอใบสั่งขาย/สั่งจองสินค้า | ID_BILLSYS_TRANS_4 |
| เปิดหน้าจอใบสั่งซื้อสินค้า | ID_POSYS_TRANS_4 |
| เปิดหน้าจอซื้อ/รับสินค้า | ID_POSYS_TRANS_10 ถ้า GRIRBillStatus=1 มิฉะนั้น ID_POSYS_TRANS_6 |
| เปิดหน้าจอขายสินค้า,บริการ | ID_BILLSYS_TRANS_6 |

ดับเบิลคลิกแถวขวา → เปิดทะเบียนสินค้า (ID_ICSYS_ITEMMASTER)

##### ผลข้างเคียง
- ปุ่ม "คำนวณยอดสั่งจองที่สามารถทำ PO ได้": `UPDATE BCSaleOrderSub SET POQty = Qty − IsNull(sum(BCPurchaseOrderSub.Qty by SORefNo & SOLINENUMBER),0) WHERE IsCancel=0 AND POQty Is Null …` ; ข้อความ "Process Complete!!!" / "ใบสั่งจองสินค้าในตารางถูกต้องแล้ว !!!"
- อื่นๆ อ่านอย่างเดียว ไม่เขียนตาราง

---

### 5. [591] ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า

- **ตำแหน่งเมนู**: Bill > (กลุ่มยังไม่ verify) ; menuconfig.xml:157 ; `OnCheckSOSRData` (MainFrm.cpp:10739-10748) → `RunChampMagaDlg("cf_sosrdate_dlg")` → `CConfirmSOSRDateDlg` (champ-lib\CHAMPMEGA\ConfirmSOSRDateDlg.cpp/.h, champmega.rc ; IDD_CONFIRM_SOSRDATE 564x250 ; caption "30025$Confirm SO/SR date" ; อยู่นอก mirror dec-sa)
- **วัตถุประสงค์**: บันทึก "วันที่ยืนยัน" และ "วันที่ส่งจริง" รายแถวของใบสั่งขาย เพื่อเทียบกับวันที่กำหนดเดิม

##### หน้าจอ

| # | ป้าย/key | control → คอลัมน์ | ชนิด | กติกา |
|---|---|---|---|---|
| 1 | SALE_DOCNO$ (เลขที่ใบสั่งขาย) | — | lookup (config `SaleorderConfig`) | เลือกแล้วเติมรหัส/ชื่อลูกค้า |
| 2 | ARCODE$ / ARNAME | — | read-only | เติมจากใบสั่งขาย |
| 3 | BILL_DOCNO$ (เลขที่ใบขาย) | — | lookup (config `BillingTransConfig`) | read-only จนกว่าเลือกใบสั่งขาย ; กรองเฉพาะใบขายที่อ้างใบสั่งขายนี้ ; ออกจากช่องตรวจ `bcarinvoicesub` ไม่ตรง → SALEINV_SO_NOTMATCH |
| 4 | ปุ่ม F5_SHOWDATA$ (OK) / CLEAN_SCREEN$ (New) / F11_SAVE$ (Save, ปิดจนกว่าโหลด) / ESC_CLOSE$ (Close) | — | — | — |
| 5 | label สถานะ "Ready" | — | — | ตัวจับเวลา 30 วินาที |

grid: col0 `*` = แถวนี้มีใบขายแล้ว ; ITEMCODE, ITEMNAME, UINTCODE, ITEMQTY, ORIGINAL_DATE (`BCSaleOrderSub.originaldate` read-only), CONFIRM_DATE_ENG (`confirmdate` แก้ได้ วันที่ไทย), ACTUAL_DATE (`actualdate` แก้ได้เมื่อเลือกเลขที่ใบขาย)

##### ผลข้างเคียง
- โหลด: `select itemcode,itemname,qty,unitcode,originaldate,confirmdate,actualdate,roworder,linenumber,(count of bcarinvoicesub non-cancelled ตาม docno/sorefno/reflinenumber) from bcsaleordersub where docno=… order by linenumber`
- บันทึก: กติกา `originaldate ≤ confirmdate ≤ actualdate` (ผิด → SOSRCONFIRMDATE_ERR พร้อมเลขแถว) ; ไม่มีใบขาย → `update bcsaleordersub set confirmdate=… where roworder=…` ; มีใบขายและแถวมี `*` → อัปเดต actualdate ด้วย ; แต่ละแถวที่แก้เขียน `AddUserLogBook` "ChampMega update line=[n] confirm [old]->[new] actual [old]->[new]" ; สำเร็จ → SAVE_COMPLETED [n] TRANSAC
- หมายเหตุ: ฟอร์มใบสั่งขายปกติ (5326) ไม่เขียน originaldate/confirmdate/actualdate (DDL มีคอลัมน์ใน BCSaleOrderSub) — จุดที่ตั้งค่า originaldate ครั้งแรก ยังไม่ verify (ไม่พบในฟอร์ม 5326)

##### Guard
ไม่พบตรวจงวดปิด/IsPostGL ในหน้านี้ ; ตรวจเฉพาะลำดับวันที่และความตรงกันของใบขายกับใบสั่งขาย

---

### 6. [5304] ปรับปรุงวันที่ส่งของให้ลูกค้า

- **ตำแหน่งเมนู**: Bill > (กลุ่มยังไม่ verify) ; menuconfig.xml:167 ; `OnBillsysTrans11` (MainFrm.cpp:2882-2889) ; `CBILLDlgOverhaulTransport` (modal CMyBaseDialog ; IDD_BILL_DLG_OVERHAUL_TRANSPORT 329x191, rc:8169) ; caption key 30011
- **วัตถุประสงค์**: ปรับ "วันที่ส่งของ" และ "วันที่ครบกำหนด" ของ **ใบขายเชื่อ (BCARInvoice)** โดยนับเครดิตจากวันที่ส่งของ — ไม่เกี่ยวกับใบสั่งขายแม้ชื่อเมนูจะใกล้กัน

##### หน้าจอ

| # | ป้ายไทย | ชนิด | หมายเหตุ |
|---|---|---|---|
| 1 | รหัสลูกค้า: | lookup (config ArTransConfig) + ชื่อ read-only | — |
| 2 | ปุ่ม "เลือกบิล" | — | `CMyMultiSelect "BillingTransConfig"` กรอง `BCARInvoice.ARCode=… AND BillType<>0 AND BillType<>2 AND BillBalance<>0 AND IsCancel=0 AND IsCompleteSave=1` (+ ไม่ซ้ำที่เลือกแล้ว) |
| 3 | ปุ่ม "ปรับปรุงเครดิตนับจากวันที่ส่งของ" | — | คำนวณวันครบกำหนดใหม่ทุกแถว |
| 4 | ปุ่ม "F11=ตกลง" / "Esc=ยกเลิก" | — | — |

grid: DOCNO (เลขที่), DOCDATE, CREDIT_OF_BILL (เครดิตตามออกบิล), OLD_DUEDATE (วันที่ครบกำหนดเดิม), DELIVERY_DATE (วันที่ส่งของ — แก้ได้), CREDIT_DELIVERY_DATE (เครดิตนับจากวันที่ส่งของ — แก้ได้ ค่าตั้งต้น `BCAR.TransportCredit`), NEW_DUEDATE (วันที่ครบกำหนดใหม่ — แก้ได้)

##### คำนวณและผลข้างเคียง
- วันครบกำหนดใหม่ = `CalcDate(วันที่ส่งของ, เครดิตนับจากวันที่ส่งของ)` (เมื่อแก้คอลัมน์ 4/5 และเมื่อกดปุ่มปรับปรุง)
- ตกลง: `Update BCARINVOICE SET DeliveryDate=…, DueDate=… WHERE ArCode=… AND DocNo=…` ต่อแถว แล้ว `EndDialog(1)` ; ไม่มี AddUserLogBook ; ไม่พบตรวจงวดปิด/IsPostGL
- คีย์: PageUp/PageDown/Enter เลื่อนแถว (`PreTranslateMessage`)

---

### 7. สิ่งที่ยังไม่ verify / ไทยที่อนุมาน

| # | หัวข้อ | สถานะ |
|---|---|---|
| 1 | ไทยจริงของ GetMyRes keys (หัวคอลัมน์ ITEMCODE…CONDIT_TRANS, ข้อความ guard ทุก key, CB_DOCTYPE_1..4, PO_SALE/PO_ORDER, SEND_TO) | อยู่ใน DB `myresource` ไม่อยู่ใน repo — ไทยที่เขียนไว้เป็น comment ในโค้ดหรือการอนุมาน |
| 2 | กลุ่มหัวข้อ Bill ที่ครอบ 6 เมนู | ยังไม่ verify |
| 3 | ส่วนพิมพ์ฟอร์ม `OnPrintForm/GetFormData` (ChmBLFrmSaleOrder.cpp ~2145-3450), `GetWeightItem`, `ReceiveTableData` | ยังไม่อ่าน |
| 4 | ภายใน `CMyDlgStockDisplay`, `CBILLDlgDeposit`, `CBILLDlgSaleCancelHold`, `CBILLDlgSheetOrder` | ยังไม่ verify |
| 5 | การเขียนตารางสต๊อกของ `OnStockReprocess` | ยังไม่ verify |
| 6 | จุดที่ตั้งค่า `originaldate` ใน BCSaleOrderSub | ยังไม่พบ |
| 7 | พิกัด/ลำดับป้ายแต่ละช่อง Page2 ต่อ control ID | จับคู่ด้วยตัวแปร DDX ไม่ใช่พิกัด |

---

---

## ส่วน 3 — เงินก่อนส่งของ: [5327] มัดจำ / [5329] รับล่วงหน้า / [5306] คืนเงินรับล่วงหน้า

ขอบเขต: เมนู [5327] [5329] [5306] ใน atlas-02-Bill.md:828-930 (เมนู ROOT > Bill ตาม atlas.json, เรียงตามลำดับ 5327 → 5329 → 5306)
แหล่ง: decoded mirror `dec-sa\champ\` (เลขบรรทัดเท่าต้นฉบับ) + `champ-lib\BC5PROCESS\*` อ่านผ่าน `lib.read()` (gltlsprocess.cpp, bcprocess.cpp, BCARProcess.cpp ไม่อยู่ใน mirror)

> **แก้ข้อมูลใน brief:** brief ระบุไฟล์ผิด — หน้าจอที่เปิดจริงตาม `CMainFrame` คือ
> - 5327 `OnBillsysTrans5` → `CBILLFrmReceiveEarnest` (**billfrmreceiveearnest.cpp**) — MainFrm.cpp:2271-2279
> - 5329 `OnBillSys6_01` → `CBILLFrmEarnestSpecail` (**BILLFrmEarnestSpecail.cpp**) — MainFrm.cpp:10522-10529
> - 5306 `OnBillsysTrans14` → `CBILLFrmEarnestSpecailReturn` (**BILLFrmEarnestSpecailReturn.cpp**) — MainFrm.cpp:2901-2908
> - `billfrmardeposit.cpp` (`CBILLFrmArDeposit`) เปิดจาก `OnBillsysTrans8` (MainFrm.cpp:2451-2458) — **ไม่ใช่เมนูทั้ง 3 นี้**; `BILLFrmReturnEarnest.cpp` เปิดจาก `OnBillSys6_02` ซึ่งอยู่ใน `#ifdef TOHOME_PROJECT` (MainFrm.cpp:10531, 10541-10547) — variant ลูกค้าเฉพาะ ไม่ใช่ default
> - template ลงทะเบียน BC5Account.cpp:1572-1577, 5766-5778; menu id ใน resource.h:4075 (5306) / 4096 (5327) / 4098 (5329)
> - **ไม่มีตาราง `BCARDepositSub`** ใน DDL — รายละเอียดการรับเงินอยู่ที่ `BCRecMoney`/`BCPayMoney` + `BCChqIn`/`BCChqOut` + `BCCreditCard` + ตารางภาษีขาย/ภาษีหัก ณ ที่จ่าย + `BCTrans`/`BCTransSub`

Variants ที่เห็น (ไม่ใช่ default): `TOHOME_PROJECT` (5329 ใช้ `CGLSpecial::ProcessConditionGroup71` แทน group 5 และเมนู 6_02 คืนเงิน), `CHM_PROJ` (ปิดช่องภาษีให้แก้มือ billfrmreceiveearnest.cpp:210; DueDate ไม่ย้อนคำนวณ CreditDay :1927), SCG fields (SCGID/JobID/Get Promotion Msg — โค้ดลูกค้าเฉพาะราย การซ่อน/แสดงตาม option **ยังไม่ verify**)

---

### 0. ส่วนกลางที่ทั้ง 3 จอใช้ร่วมกัน (อ้างจากแต่ละเมนูด้านล่าง)

#### 0.1 หน้าต่างรับเงิน/จ่ายเงิน (F7) — `CLIBMoneyRec`
- dialog `IDD_LIB_MONEY_RECEIVE` caption **"บันทึกรับเงิน"** (BC5Account.rc:3027); เปิดจาก `OnRec` (myespeciallyformview.cpp:4320-4431) ส่งยอดสุทธิของเอกสารเข้าไป
- Tab สร้างใน LIBMoneyRec.cpp:294-349 (ป้ายเป็น language key — ไทยจริงอยู่ใน DB myresource ไม่อยู่ใน repo — ความหมายอนุมาน):

| tab | key | แสดงเมื่อ | ช่องในหน้า (จาก .rc) | หลักฐาน |
|---|---|---|---|---|
| 0 | `CASH` หรือ `CASH_BANK` | เสมอ (`CASH` เมื่อ product ตัดโมดูลธนาคาร/เช็ค `LMD_REMOVE_BANKCHQ`) | เงินสด, เงินเกิน (ExcessAmount), โอนเข้าธนาคาร + สมุดบัญชีธนาคาร + วันที่โอน, รายได้อื่น/ค่าใช้จ่ายอื่น, เงินสดย่อย | LIBMoneyRec.cpp:292-302; rc:3126 |
| 1 | `CREDITCARD` | โหมดรับ (mode 0/2) | เลขที่บัตร, เลขอนุมัติ, ธนาคาร, สาขา, วันครบกำหนด, จำนวนเงิน, ประเภทบัตร, % ค่าธรรมเนียม, ค่าธรรมเนียม, รวม | LIBMoneyRec.cpp:307-312; rc:3043 |
| 2 (โหมดรับ) / 1 (โหมดจ่าย) | `CHEQ` | ทั้งรับและจ่าย (mode ≠ 4) | radio เช็คฉบับใหม่ / ตัดยอดจากฉบับเดิม; ธนาคาร, สาขา, มูลค่าเช็ค, เลขที่เช็ค, ยอดจ่าย, วันครบกำหนด, หมายเหตุ; checkbox ตัดจ่ายหลายใบเสร็จ; ปุ่ม F9 Update / F6 Delete / F5 Clear | LIBMoneyRec.cpp:315-317, 344-349; rc:3082 |
| 3 | `COUPON_RECMONEY` | m_nMyType==1 และมี coupon type | คูปอง | LIBMoneyRec.cpp:320-324 |
| 3/4 | `OTHER` | option HomeBuddy เท่านั้น | อื่น ๆ | LIBMoneyRec.cpp:326-338 |

- โหมดจ่าย (ใช้ใน 5306) มีแค่ tab 0 + เช็ค (ไม่มีบัตรเครดิต) — LIBMoneyRec.cpp:344-349
- ทอนเงิน: ถ้า `IsSmallChange=1` เปิด `CLIBMoneyCashChange` เพิ่ม (ยังไม่ verify รายละเอียดช่อง)

#### 0.2 สูตรตรวจยอดรับเงินครบ — `IsMoneyOk` (myespeciallyformview.cpp:4266-4318)
```
totalRec = cash + chq + credit + bank + petty + coupon + wallet + ExcessAmount2 − ExcessAmount1
โหมดรับ : totalRec += (OtherExpense − OtherIncome)      โหมดจ่าย : totalRec += (OtherIncome − OtherExpense)
ต้อง totalRec == ยอดสุทธิของเอกสาร  ไม่เท่า → บันทึกไม่ได้ + เปิดหน้าต่าง F7 ให้อัตโนมัติ
```
(ความหมายแยกของ ExcessAmount1/2 = เงินเกิน 2 ทิศ — แยกชื่อไทยไม่ได้จากโค้ด ยังไม่ verify)

#### 0.3 การบันทึกรายการรับ/จ่ายเงิน — `OnSaveRec` (myespeciallyformview.cpp:1998-2330)
- ตาราง: โหมดรับ → `BCRecMoney`, โหมดจ่าย → `BCPayMoney`; หนึ่งแถวต่อหนึ่งช่องทางชำระ
- `SaveFrom`: **0 = รับเงินมัดจำ (5327)**, **8 = รับเงินล่วงหน้า (5329)**, **9 = คืนเงินรับล่วงหน้า (5306, ฝั่ง BCPayMoney)**
- `PaymentType`: 0 เงินสด, 1 บัตรเครดิต, 2 เช็ค, 3 โอนธนาคาร, 4 เงินสดย่อย, 10 คูปอง, 12 wallet
- คอลัมน์ที่เขียน: DocNo, DocDate, SaveFrom, CurrencyCode, ExchangeRate, ArCode(หรือ ApCode), LineNumber, PayAmount, PaymentType, MyDescription, DepartCode, ProjectCode, SaleCode, CashierCode, ChangeAmount, RefNo, RefDate, ChqTotalAmount, PayChqState, BankCode, BankBranchCode, TransBankDate, CreditType, ConfirmNo, ChargeAmount, ChargeWord, BookNo
- ผลข้างเคียงตามช่องทาง:

| ช่องทาง | ผล |
|---|---|
| โอนธนาคาร | `UPDATE BCBankBook SET Balance = Balance + amt` (จ่าย = ลบ) |
| เงินสดย่อย | `UPDATE BCPettyCash SET Balance = Balance + amt` |
| เช็คฉบับใหม่ (รับ) | `INSERT BCChqIn` (Amount, Balance, ChqUseStatus, ChqNumber, BankCode, BankBranchCode, DocNo, DueDate, SaveFrom, ArCode, ReceiveDate, IsSAP …) |
| ตัดยอดเช็คเดิม | `UPDATE BCChqIn SET Balance = Balance − amt` |
| เช็คจ่าย (5306) | เขียน `BCChqOut` (ยังไม่ verify รายชื่อคอลัมน์) |
| บัตรเครดิต | `INSERT BCCreditCard` |

- ก่อนบันทึกซ้ำ (แก้เอกสาร) โค้ดย้อนยอดเดิมของธนาคาร/เงินสดย่อย/เช็คก่อนเสมอ
- `IsChqRefence(DocNo, nSaveFrom, nChqType)` (mytools.cpp:911) = จริงเมื่อเช็คที่รับในเอกสารนี้ถูกเอกสารอื่นตัดไปแล้ว (PayChqState=1) → ห้ามแก้/ลบ/ยกเลิก

#### 0.4 ยกเลิก / อนุมัติ / ป้ายสถานะ
| การกระทำ | ฟังก์ชัน | เงื่อนไข | หน้าต่าง | เขียน DB |
|---|---|---|---|---|
| ยกเลิก (Ctrl+F5) | `BillCancel` myespeciallyformview.cpp:5590 | ไม่ผ่าน GL (IsPostGL), สิทธิ์ `bCanCancelBill`, ยังไม่ยกเลิก (`MSG_CANNOT_CANCEL`) | `CMyDlgBillCancel` IDD_LIB_BILLCANCEL_DIALOG caption "Cancel Document" (rc:3511) — ต้องเลือก "YES" ใน combo; ข้อความเตือน "เมื่อท่านยกเลิกเอกสารนี้แล้วท่านจะไม่สามารถนำเอกสารนี้ไปใช้ในระบบได้อีก" | CancelCode, CancelDateTime, IsCancel=1 (MyDlgBillCancel.cpp:64-66) แล้วเรียก `BillCancelUpdateData` ของจอ |
| อนุมัติ (Shift+F5) | `BillConfirm` myespeciallyformview.cpp:5646 | สิทธิ์ `bCanConfirmBill` | `CMyDlgBillConfirm` (rc:3490) ต้องเลือก "YES"; เตือนว่าอนุมัติแล้วแก้ไม่ได้ | ConfirmCode, ConfirmDateTime, IsConfirm=1 (MyDlgBillConfirm.cpp:64-66) |

- ป้ายสถานะ `DrawBillStatus` (myespeciallyformview.cpp:5685): **"CC"** แดง = ยกเลิก; **"0.00"** น้ำตาล = ยอดคงเหลือ 0 (ถูกตัดใช้หมด); **"CF"** เขียวอมฟ้า = อนุมัติ; **"N"** = ปกติ
- `IsPostGL` (lib-frmwrk mybasedbformview.cpp:4126) อ่านคอลัมน์ IsPostGL ของแถว + แสดง BatchNo/`BCPostTrans`
- `OnDepositUpdateChqAmount(mode, code, amt, mark)` (myespeciallyformview.cpp:6545): mode 0 → `UPDATE BCAR SET ChqOnHand = ChqOnHand ± chq`, mode 1 → `BCAP`

#### 0.5 การลงบัญชี GL อัตโนมัติ (ไม่มีรหัสบัญชีฝังในโค้ด)
- บัญชีทุกบรรทัดมาจาก **สูตรลงบัญชี (GL format) ที่ผู้ใช้กำหนด** (`pGLFormat.pFormatList[].strAccount` + `nCondition`) — ตารางเก็บสูตร ยังไม่ verify
- สูตรเริ่มต้นต่อจอ: 5327 = กลุ่ม 7 (billfrmreceiveearnest.cpp:1896), 5329 = กลุ่ม 40 คอมเมนต์ "รับเงินล่วงหน้า(ไม่คิดภาษี)" (BILLFrmEarnestSpecail.cpp:1320), 5306 = กลุ่ม 41 (BILLFrmEarnestSpecailReturn.cpp:1469)
- source ของ GL: 5 = BCARDeposit, 38 = BCARDepositSpecial, 39 = BCReturnDepSpecial (GLDlgCancelPost.cpp:450, 552, 555)
- `ProcessConditionGroup5` (gltlsprocess.cpp:3211-3705) — ใช้กับ 5327 (source 5) และ 5329 (source 38, ตาราง BCARDepositSpecial :3227). ยอดตาม condition:

| cond | ยอดที่ใช้ | cond | ยอดที่ใช้ |
|---|---|---|---|
| 1 | BeforeTaxAmount | 8 | ChargeAmount |
| 2 | TaxAmount | 9 | ExcessAmount2 |
| 3 | SumOfWTax | 10 | ExcessAmount1 |
| 4 | SumCashAmount | 14 | OtherIncome |
| 5 | SumCreditAmount + ChargeAmount | 15 | OtherExpense |
| 6 | SumChqAmount | 83 | sumcurrentacc |
| 7 | SumBankAmount (แยกรายสมุดธนาคารด้วย `PostBankTransferSub`) | 90/91 | ปัดเศษ Dr/Cr |

  - เขียน `BCTransSub` (Participant = ArCode) + `BCTrans` + `AddGLChq(BCRecMoney)` → `bcglchq`
  - เลือกเฉพาะเอกสาร `gltransdata = 0 AND iscancel = 0`; ประมวลผลซ้ำ = ลบ bctrans/bctranssub/bcglchq ของ source นั้นก่อน
  - ถ้าบริษัทตั้ง `AutoPostGL = 1` → isconfirm=1 ใน BCTrans และ `Update BCARDeposit set ispostgl=1` (gltlsprocess.cpp:3676)
- `ProcessConditionGroup44` (gltlsprocess.cpp:15991-16420) — ใช้กับ 5306 ตาราง `BCReturnDepSpecial` (:15994): cond 1 NETAMOUNT, 2 SumCashAmount, 3 SumChqAmount, 4 SumBankAmount (`PostBankTransferSub` โหมด bcpaymoney SaveFrom 9 :16188-16194), 5 PettyCashAmount, 83 sumcurrentacc (`bcpaymoney paymenttype = 9` :16220), 90/91 ปัดเศษ; AutoPostGL → `AddGLChq("BCPayMoney")` (:16403-16415) + update ispostgl (:16420)
- ถ้าไม่ตั้ง auto: จอบันทึกบัญชีเองผ่าน `OnSaveAccount(...,source,...)` (F2 เปิดหน้าต่างบัญชี)

---

### [5327] บันทึกใบรับเงินมัดจำ
- **ตำแหน่งเมนู**: Bill > (ลำดับหลัง 591 ตรวจสอบวันที่ใบสั่งขาย, atlas-02-Bill.md:828); handler `OnBillsysTrans5` (MainFrm.cpp:2271-2279); form `CBILLFrmReceiveEarnest` (billfrmreceiveearnest.cpp, 2,888 บรรทัด); `IDD_BILL_FRM_RECEIVE_EARNEST` 495x288 (BC5Account.rc:7638); view key "30006", bill code "DEP" (:36), config `ARDepositConfig`, GL slip source 5, สิทธิ์ `ID_BILLSYS_TRANS_5` (:29-66)
- **วัตถุประสงค์**: รับเงินมัดจำจากลูกค้าก่อนส่งของ/ออกใบกำกับขาย ออกใบกำกับภาษีของเงินมัดจำได้ (ถ้ากรอกเลขที่ใบกำกับ) ยอดคงเหลือ (BillBalance) จะถูกตัดใช้ภายหลังในใบขาย/ใบกำกับ (F8) หรือใบลดหนี้เงินมัดจำ

#### หน้าจอและช่องหัวเอกสาร (ตามลำดับบนจอ บน→ล่าง ซ้าย→ขวา; ป้ายไทย verbatim จาก .rc)
| # | ป้ายไทย | control / คอลัมน์ DB (BCARDeposit) | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup / ตาราง | การตรวจ/คำนวณ | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | เลขที่ใบรับเงินมัดจำ: | IDC_ED_DOCNO / DocNo | text ≤20 | ใช่ | เลขรันตามกลุ่มเอกสาร (running บน BCARDeposit.DocNo) | — | ว่าง → CANNOT_SAVE_NOTDOCNO; ซ้ำ → DOCNO_DUPLICATE; ถ้าผู้ใช้ไม่มีสิทธิ์เปลี่ยนเลข ช่องถูกปิด โฟกัสไปรหัสลูกค้า | :166-262, :2404, :2431, :451-667 |
| 2 | วันที่ออกเอกสาร: | IDC_ED_DOCDATE / DocDate | date (พ.ศ.) | ใช่ | วันนี้ | — | ว่าง → INSERT_DOCDATE; ตรวจงวดปิด; ปิดช่องถ้าไม่ตั้ง `IsCanEditDocDate=1` และไม่ใช่ sa | :166-262 |
| 3 | SCG Family ID/เบอร์โทร : | IDC_ED_SCGID / SCGID | text | ไม่ | ดึงจากลูกค้า | BCAR | ฟิลด์ลูกค้าเฉพาะราย | :2125-2199 |
| 4 | เลขที่ใบกำกับ: | IDC_ED_TAXNO / TaxNo | text ≤20 | ไม่ | **= DocNo** (คัดลอกตอนเริ่มเอกสารใหม่) | running บน BCARDeposit.TaxNo | >20 → TAXNO_MORE_LENG; เปลี่ยนแล้วคำนวณภาษีใหม่; **ว่าง = ไม่คิด VAT** | :2404, :1944 |
| 5 | วันที่ใบกำกับ: | IDC_ED_TAXDATE / TaxDate | date | ถ้ามี TaxNo | วันที่เอกสาร | — | มี TaxNo แต่ไม่มีวันที่ → CHECK_TAXDATE | :451-667 |
| 6 | (ปุ่ม) GET_SCGPROMOTIONMSG2 "Get Promotion Msg" | IDC_BTTN_GETPROMOTION | button | — | — | — | ลูกค้าเฉพาะราย | atlas :836 |
| 7 | ประเภทภาษี: | IDC_ED_TAXTYPE / TaxType | combo | ใช่ | ค่าบริษัท `DefaultTaxType` แล้วทับด้วยของลูกค้า | ตัวเลือก key `CB_TAXTYPE_1` / `CB_TAXTYPE_2` / `TAX_ZERO` = 0 แยกนอก / 1 รวมใน / 2 อัตราศูนย์ (ไทยจริงอยู่ใน DB myresource ไม่อยู่ใน repo — ความหมายอนุมาน) | เปลี่ยนแล้วคำนวณใหม่ (`OnSelchangeTaxType`); 2 → บังคับอัตรา 0 | :166-262, :2214-2355 |
| 8 | กลุ่มเอกสาร: | IDC_ED_DOCGROUP / BillGroup | lookup | ไม่ | — | กลุ่มเอกสาร (รูปแบบเลขรัน) | เลือกแล้วรันเลขตามรูปแบบกลุ่ม; แก้ได้หลังผ่าน GL | :166-262 |
| 9 | รหัสลูกค้า: + ชื่อ | IDC_ED_ARCODE / ArCode (+IDC_ED_ARNAME แสดงอย่างเดียว) | lookup | ใช่ | — | BCAR.Code (`ArTransConfig`) | ว่าง → CANNOT_SAVE_NOTARCODE; ไม่พบ → NOT_FOUND_AR; ออกจากช่องดึง: ชื่อ (Name1/Name2 ตาม `DefaultArName`), TaxType, DefaultTaxRate (ไม่มีใช้ของบริษัท), SaleCode (ถ้าว่าง), BillCredit → จ่ายส่วนที่เหลือภายใน/วันที่ส่งของ, SCGID; เตือน WARN_AR_QUOTATION ถ้าเปิด ReminderQuotaion | :2125-2199, :1979 |
| 10 | (ปุ่ม) เลือกใบเสนอราคา | IDC_BTTN_QUATATION → QARefno | browse | ไม่ | — | `QuotationTransConfig` กรอง `ArCode=… AND BCQuotation.isconfirm=1 AND DocNo NOT IN (QARefNo ของใบมัดจำที่ไม่ยกเลิก)` | ต้องมีลูกค้าก่อน (SALEORDER_ARCODE); เติม QARefno, BeforeTax, Total, JobID (แล้วล็อก JobID) แล้วคำนวณภาษี | :2662-2691 |
| 11 | ชื่อในการออกบิล: | IDC_ED_ARNAME2 / ARName (200) | text | ไม่ | จาก lookup ลูกค้า | BCAR | — | :1979 |
| 12 | ที่อยู่ในการออกบิล: | IDC_ED_ARADDRESS / ARAddress (255) | text | ไม่ | จาก lookup ลูกค้า | BCAR | — | :1979 |
| 13 | เลขที่ใบจอง: + หมายเหตุใบจอง | IDC_ED_REFNO / RefNo + IDC_ED_REFNO_REMARK (แสดง) | lookup | ไม่ | — | `SaleOrderConfig` กรอง `BCSaleOrder.ArCode=… [AND SaleCode] AND BillStatus=0 AND IsCancel=0 AND SOStatus=1 AND TaxType=…` | ออกจากช่องตรวจ: ยกเลิก → RESERVE_ISCANCEL; ออกบิลแล้ว → RESERVE_BILLSTATUS; ไม่พบ → DATABASE_NOTHAVE_RESERVE; เลือกแล้วคัดลอก หมายเหตุ/แผนก/การจัดสรร/โครงการ/ชื่อ-ที่อยู่/พนักงาน + ยอดเงิน | :2052-2123, :2362-2402 |
| 14 | รหัสพนักงาน: (key Sale_Code) | IDC_ED_SALECODE / SaleCode | combo | ไม่ | จากลูกค้า | BCSale | แก้ได้หลังผ่าน GL | :166-262 |
| 15 | แผนก: | IDC_ED_DEPART / DepartCode | combo | ไม่ | — | BCDepartment | แก้ได้หลังผ่าน GL | :166-262 |
| 16 | โครงการ: | IDC_ED_PROJECT / ProjectCode | combo | ไม่ | — | BCProject | แก้ได้หลังผ่าน GL | :166-262 |
| 17 | การจัดสรร: | IDC_ED_ALLOCATE / AllocateCode | combo | ไม่ | — | BCAllocate | แก้ได้หลังผ่าน GL | :166-262 |
| 18 | จ่ายส่วนที่เหลือภายใน: … วัน | IDC_ED_CREDITDAY / CreditDay | number | ไม่ | BillCredit ของลูกค้า | — | วันที่ส่งของ = วันที่เอกสาร + วัน | :1916 |
| 19 | วันที่ส่งของ: | IDC_ED_DUEDATE / DueDate | date | ไม่ | คำนวณ | — | ย้อนคำนวณ CreditDay = ส่วนต่าง − 1 (ไม่ทำใน Chm mode) | :1927 |
| 20 | หมายเหตุ: | IDC_ED_REMARK / MyDescription | text | ไม่ | — | — | แก้ได้หลังผ่าน GL | :166-262 |
| 21 | รหัสสกุลเงิน: | IDC_ED_CURRENCY / CurrencyCode | combo | ไม่ | สกุลหลัก | BCCurrency | ปิดเมื่อ MultiCurrency=0; เลือกแล้วดึง ExchangeRate | :166-262 |
| 22 | อัตราแลกเปลี่ยน: | IDC_ED_EXCHANGE_RATE / ExChangeRate DECIMAL(19,10) | number | ไม่ | BCCurrency.ExchangeRate | — | — | DDL |
| 23 | Job ID : | IDC_ED_JOBID / JobID | text | ไม่ | จากใบเสนอราคา | — | ล็อกหลังเลือกใบเสนอราคา | :2662-2691 |
| 24 | " :" (ป้ายว่าง) | IDC_ED_JOBID2 / QARefno | text | ไม่ | จากปุ่มเลือกใบเสนอราคา | BCQuotation | — | :2662-2691 |
| **กลุ่มยอดเงินท้ายจอ** |
| 25 | อัตราภาษีมูลค่าเพิ่ม: | IDC_ED_TAXRATE / TaxRate | number | ไม่ | ของลูกค้า/บริษัท | — | แก้ได้ตามสิทธิ์ `IsUserEditTax` (`UpdateTaxControl`) | :166-262 |
| 26 | รวมเงินทั้งสิ้น: | IDC_ED_TOTALAMOUNT / TotalAmount | number | ใช่ | — | — | ภาษีรวมใน: กรอกช่องนี้แล้วแยก VAT; =0 → CANNOT_SAVE_NOT_SUMDETAIL; ติดลบ → CANNOT_SAVE_WORNG | :1445-1475 |
| 27 | จำนวนเงินก่อนภาษี: | IDC_ED_BEFORETAX / BeforeTaxAmount | number | ใช่ | — | — | ภาษีแยกนอก: กรอกช่องนี้แล้วบวก VAT | :1360-1390 |
| 28 | ยอดเงินสุทธิ: | IDC_ED_SUMNETAMOUNT / NetAmount | number (คำนวณ) | ใช่ | — | — | = รวมเงินทั้งสิ้น − ภาษีหัก ณ ที่จ่าย | :1373, :1455 |
| 29 | ภาษีมูลค่าเพิ่ม: | IDC_ED_TAXAMOUNT / TaxAmount | number | — | คำนวณ | — | แก้มือได้ตามสิทธิ์ (ไม่ได้ใน Chm mode) → `CalcTaxAmount` | :210, :1542-1621 |
| 30 | ยอดเงินคงเหลือ: | IDC_ED_BILLBALANCE / BillBalance | number (แสดง) | — | ดูสูตร | — | ปรับเฉพาะตอนยังไม่ถูกตัดใช้ (Net == Balance) | :1378, :1460 |

คอลัมน์ที่บันทึกแต่ไม่อยู่บนจอ: SumOfWTax (จาก F6), SumCashAmount/SumChqAmount/SumCreditAmount/ChargeAmount/SumBankAmount/ExcessAmount1/2/OtherIncome/OtherExpense (จาก F7), GLFormat, GLTransData, RecurName, IsConfirm/IsCancel/IsPostGL/IsReturnMoney, creator/editor (OnSaveData :669-890)

#### ตารางรายการ (grid)
N/A — จอนี้เป็นหัวเอกสารล้วน ไม่มี `m_BInputListWnd`

#### ยอดรวมท้ายเอกสาร / สูตรคำนวณ (billfrmreceiveearnest.cpp:1354-1540, `OnSelchangeTaxType` :2214-2355)
| กรณี | ช่องที่ผู้ใช้กรอก | สูตร | BillBalance (ยอดให้ตัดใช้) |
|---|---|---|---|
| ภาษีแยกนอก (0) + มีเลขที่ใบกำกับ | จำนวนเงินก่อนภาษี | VAT = round(BeforeTax × rate / 100); Total = BeforeTax + VAT; Net = Total − WTax (:1370-1373) | **= BeforeTax** (คอมเมนต์ "ภาษีแยกนอก ยอดเงินคงเหลือ = ยอดเงินก่อนภาษี" :1360) |
| ภาษีรวมใน (1) + มีเลขที่ใบกำกับ | รวมเงินทั้งสิ้น | VAT = round(Total × rate / (100 + rate)); BeforeTax = Total − VAT; Net = Total − WTax (:1452-1455) | **= Total** (:1460) |
| ไม่มีเลขที่ใบกำกับ (ทุกประเภท) | ช่องใดก็ได้ | ไม่คำนวณ VAT (ล้างช่อง) ; Total = BeforeTax ; Net = ยอด − WTax (:1381-1386, :1463-1468 คอมเมนต์ "กรณีไม่ใส่เลขที่ใบกำกับภาษี ไม่คำนวณยอดภาษี") | = ยอดที่กรอก |
| อัตราศูนย์ (2) | — | อัตราบังคับ 0 | — |
- คำนวณเฉพาะเมื่อยอด **> 1** (`if(dBeforeTax >1)` :1367, :1449) — ยอด ≤ 1 จะไม่คำนวณใหม่ (edge case ห้ามลอก)
- ปัดเศษ `XRound(…, AmountPoint)` บน double — BC ต้องใช้ decimal (Rule 20)
- F6 ภาษีหัก ณ ที่จ่าย: ฐาน = BeforeTax; Net = Total − WTax; Balance = BeforeTax (แยกนอก) หรือ Total (:1623-1721)
- แถวภาษีขายแถวแรก `MyUpdateFirstTaxRow` แปลงเป็นสกุลหลัก คำอธิบาย key `DEP_CONDIT` (:2446)

#### ปุ่ม แถบเครื่องมือ คีย์ลัด (PreTranslateMessage :1623-1721; OnMore :1723-1817)
| ปุ่ม/คีย์ | ผล | เงื่อนไข |
|---|---|---|
| แถบมาตรฐาน | New, Save, Del, Find, Prev, Next, Close, Info, Key, Recur (เปิด recur ใน ctor :29-66) | — |
| F2 | หน้าต่างบัญชี GL (`OnAccount`) | — |
| F3 | หน้าต่างภาษีขาย `OnOutputTax("ARDepositConfig",…)` แล้วคัดลอก TaxNo/TaxDate จากแถวแรกกลับหัวเอกสาร | — |
| F6 | หน้าต่างภาษีหัก ณ ที่จ่าย `OnWARTax` (ฐาน = ก่อนภาษี) | — |
| F7 | หน้าต่างรับเงิน (§0.1) ด้วยยอดสุทธิ | ต้องมีเลขที่เอกสาร |
| F11 | บันทึก | — |
| F12 | เส้นทางเอกสาร (Document flow) | — |
| Shift+F5 | อนุมัติ (§0.4) | สิทธิ์ `bCanConfirmBill` ไม่มี → NOT_CONFIRM_NO |
| Ctrl+F5 | ยกเลิก → `OnConfirmCancel` → `BillCancel("BCARDeposit")` | อนุมัติแล้ว → CANNOT_CANCEL |
| Key (IDC_MORE) | เมนู: F2GL, F3BILLTAX, F6_TAX, F7_BL_RECEIVE, F11_SAVE, F12_CHECK_DOCUMENT, SHIFT_F5, CTRL_F5 | — |
| Info | ยอด: เงินสด, เช็ค, บัตร, โอน, ก่อนภาษี, VAT, รวม, ภาษีหัก ณ ที่จ่าย, สุทธิ, คงเหลือ | :1819 |
| Print (Ctrl+P) | ตรวจ `OnCheckBeforePrint` + นับครั้งพิมพ์ใน `bcprintlog` | :1249 |

#### สถานะเอกสารและ workflow
1. **บันทึก** (F11) → ต้องรับเงินครบ (`IsMoneyOk`) มิฉะนั้น AMOUNT_NOT_EQUAL_RECEIVE + เปิด F7 อัตโนมัติ
2. **อนุมัติ** (Shift+F5) → IsConfirm=1 → แก้/ลบ/ยกเลิกไม่ได้ (NOT_SAVE_ISCONFIRM / NOT_DELETE_ISCONFIRM / CANNOT_CANCEL)
3. **ถูกตัดใช้** → BillBalance ลดลง; เมื่อ Net ≠ BillBalance → แก้/ลบ/ยกเลิกไม่ได้ (NOT_EDIT_BILLBALANCE / NOT_DELETE_BILLBALANCE); เหลือ 0 → ป้าย "0.00"
4. **คืนเงินแล้ว** (IsReturnMoney=1) → แก้/ลบ/ยกเลิกไม่ได้ (เอกสารที่ตั้ง flag นี้ไม่อยู่ใน 3 เมนูนี้ — ยังไม่ verify ว่าจอใด)
5. **ผ่าน GL** (IsPostGL=1) → แก้ได้เฉพาะ แผนก/การจัดสรร/โครงการ/จ่ายส่วนที่เหลือภายใน/วันที่ส่งของ/หมายเหตุ/กลุ่มเอกสาร/พนักงาน (`AddCanEditField` :166-262); ลบ/ยกเลิกไม่ได้
6. **ยกเลิก** (Ctrl+F5) → ตรวจ IsPostGL, IsCancel, IsConfirm, ยอดถูกใช้, IsReturnMoney (`OnConfirmCancel` :2510-2541)

**การตัดใช้ใบรับเงินมัดจำ (ตัดใบรับเงินมัดจำ) จากใบขาย/ใบกำกับ — F8**
- `OnDeposit` (myespeciallyformview.cpp:4771-4854) เปิด `CLIBDlgDeposit` (libdlgdeposit.cpp)
- กรอง (libdlgdeposit.cpp:341-358): `BCARDeposit.ARCode = ลูกค้าในใบขาย AND BillBalance > 0 AND TaxType = TaxType ของใบขาย [AND DocNo NOT IN แถวที่เลือกแล้ว]` + สกุลเงินตรงกันเมื่อ multi-currency; checkbox `IDC_CHK_INCLUDETAX` = 1 → `TaxNo IS NOT NULL` (มัดจำที่ออกใบกำกับแล้ว) ไม่ติ๊ก → `TaxNo IS NULL`
- คอลัมน์ (libdlgdeposit.cpp:187-196, key ไทยจริงอยู่ใน DB myresource — ความหมายอนุมาน): `DOCNO_RECEIVE` เลขที่ใบมัดจำ (link), `REMARK`, `BALANCE` ยอดคงเหลือ, `RECEIPTAMOUNT` ยอดตัด (แก้ได้), `NETAMOUNT` สุทธิ, `SHEETORDER_DOCNO` เลขที่ใบจอง, `NEW_EXCHANGE_RATE` (multi-currency); shadow: RowOrder, DeposTaxType, DPExchangeRate, INV_HOMEAMOUNT, NEW_HOMEAMOUNT, EXCHANGE_PROFIT
- มัดจำภาษีรวมในที่มี TaxNo: สุทธิ = bal − bal × rate/(100+rate) (เอา VAT ออก) (:397-401)
- ตัดเกินยอดคงเหลือ → PAYMANT_MORE_BALANCE (:632); เลือกมัดจำผิดลูกค้า → SELET_WRONG_DEPOSIT (:545)
- บันทึก `OnSaveDeposit` (myespeciallyformview.cpp:4917-5003) → **`BCARDepositUse`**: DocNo (เลขใบขาย), DocDate, DepositNo, MyDescription, Amount, Balance, DeposTaxType, NetAmount, DPExchangeRate, NewExchangeRate, HomeAmount1, HomeAmount2, ExchangeProfit, LineNumber
- ปรับยอด `CBCProcess::ProcessARDeposit` (bcprocess.cpp:3129-3149): `UPDATE BCARDeposit SET BillBalance = BillBalance + (Amount × nMark)` — mark −1 ตอนบันทึก, +1 ตอนย้อน (`DeleteARDeposit` myespeciallyformview.cpp:5005)
- หัวใบขายเก็บ DepositIncTax, SumOfDeposit1, SumOfDeposit2 (ChmFrmBilling.cpp:3331, 3363-3364; ความหมายแยก 1/2 ยังไม่ verify); บันทึก `OnSaveARDeposit` :3843; ลบ :3195/:5883
- ตอนบันทึกใบขาย เตือนให้ตัดมัดจำเมื่อ `IsWarnDeposit=1` (key AR_DEPOSIT_OFFSET + F8EARNEST_RECEIVE, ChmFrmBilling.cpp:1327); บิลเงินสด (BillType 1/3) ห้ามใช้มัดจำเมื่อ `IsDepositCash` (EARNEST_CASH :6662)
- **ใบลดหนี้เงินมัดจำ** (BCInvCreditNote BillType=3 / IsCNDeposit) ลด BillBalance ด้วย (bcprocess.cpp:2213-2238; `Decrease_ARDeposit` myespeciallyformview.cpp:7276) — มัดจำแยกนอก หารยอดด้วย (100+rate)/100
- Document flow ของใบมัดจำ (:2475-2503) = UNION ของ BCARDepositUse (DepositNo = เอกสารนี้ ไม่นับ CN), BCInvCreditNote (InvoiceNo = เอกสารนี้ AND BillType=3 "ลดหนี้เงินมัดจำ"), BCReceiptSub1/Sub2 (BillType=2)

#### ผลข้างเคียง (OnSaveData :669-890)
| สิ่งที่เกิด | รายละเอียด | หลักฐาน |
|---|---|---|
| หัวเอกสาร | INSERT/UPDATE `BCARDeposit`: TaxType, TaxRate, DocNo, DocDate, TaxNo, TaxDate, ArCode, RefNo, DepartCode, CreditDay, DueDate, SaleCode, AllocateCode, ProjectCode, BillGroup, MyDescription, GLTransData, GLFormat, RecurName, CurrencyCode, ExChangeRate, ExcessAmount1/2, OtherIncome, OtherExpense, SumCashAmount, SumChqAmount, SumCreditAmount, ChargeAmount, SumBankAmount, SumOfWTax, TaxAmount, TotalAmount, BeforeTaxAmount, NetAmount, BillBalance, ARName, ARAddress, SCGID, JobID, QARefno, CreatorCode/CreateDateTime หรือ LastEditorCode/LastEditDateT | :669-760 |
| เช็คในมือของลูกค้า | `BCAR.ChqOnHand += ยอดเช็ค` (ย้อนยอดเก่าก่อน) | :770 |
| ภาษีขาย | `OnSaveOutputTax(0, old, new, date, ar, 1, 5, BillGroup, isMultiCurrency)` | :774 |
| ภาษีหัก ณ ที่จ่าย (ลูกค้าหักเรา) | `OnSaveARWTax` | :776 |
| รายการรับเงิน | `OnSaveReceive(0, …)` → BCRecMoney SaveFrom 0 + ธนาคาร/เช็ค/บัตร/เงินสดย่อย (§0.3) | :781 |
| GL | AutoPost: `GLTLSProcess.m_nSource = 5; ProcessConditionGroup5()` (§0.5) ไม่งั้น `OnSaveAccount(…,5,…)` | :784-791 |
| log | SAVE_RECEIVE_DEPOSIT / CANCEL_RECEIVE_DEPOSIT | :784-791, :334-388 |
| ยกเลิก | `UPDATE BCARDeposit SET BillBalance = 0.0, IsCancel = 1`; `CancelOutputTax(0, doc, 5)`; `DeleteARWTax`; `DeleteReceive(0, doc)` (ตาม m_bCancelMoney); `ReverseAccount(doc, 5)` (กลับรายการ GL) | :334-388 |
| ลบ | DELETE BCARDeposit; ย้อน BCAR.ChqOnHand; `DeleteOutputTax(0, doc, 5)`; `DeleteARWTax`; `DeleteReceive`; `DeleteAccount(doc, 5)` | :390-449 |
- **ไม่แตะยอดลูกหนี้ BCAR.DebtAmount** ตอนรับมัดจำ (เห็นเฉพาะ ChqOnHand) — ยอดมัดจำลดหนี้ผ่านใบขายที่ตัดใช้

#### Guard / ข้อความเตือนสำคัญ (key — ไทยจริงอยู่ใน DB myresource ไม่อยู่ใน repo — ความหมายอนุมาน)
| key | เงื่อนไข | หลักฐาน |
|---|---|---|
| CANNOT_DELETE | ลบเอกสารที่ไม่มีเลขที่ | :282-332 |
| (ตรวจงวดปิด `IsPeriodClose`) | วันที่อยู่ในงวดที่ปิดแล้ว (key ยังไม่ verify) | :282, :451 |
| CANNOT_DELETE_POSGL / CANNOT_SAVE_POSTGL | ผ่าน GL แล้ว | :282-332, :451-667 |
| CHQNOT_EDIT_DELETE_CANCEL | เช็คในเอกสารถูกเอกสารอื่นตัดแล้ว (`IsChqRefence(doc,0,0)`) | :282-332 |
| NOT_DELETE_ISCANCEL / NOT_SAVE_ISCANCEL | ยกเลิกแล้ว | 〃 |
| NOT_DELETE_ISCONFIRM / NOT_SAVE_ISCONFIRM | อนุมัติแล้ว | 〃 |
| NOT_DELETE_BILLBALANCE / NOT_EDIT_BILLBALANCE | ถูกตัดใช้แล้ว (Net ≠ BillBalance) หรือ IsReturnMoney=1 | 〃 |
| CONFIRM_DELETEDATA | ยืนยันลบ yes/no | :282-332 |
| CANNOT_SAVE_NOTDOCNO / DOCNO_DUPLICATE / INSERT_DOCDATE / CHECK_TAXDATE | หัวเอกสารไม่ครบ | :451-667 |
| CANNOT_SAVE_NOTARCODE / NOT_FOUND_AR | ลูกค้า | 〃 |
| CANNOT_SAVE_NO_WALLET | wallet | 〃 |
| CANNOT_SAVE_NOT_SUMDETAIL / CANNOT_SAVE_WORNG | ยอด = 0 / ติดลบ | 〃 |
| AMOUNT_NOT_EQUAL_RECEIVE | ยอดรับเงินไม่เท่ายอดสุทธิ → เปิด F7 | 〃 |
| TAXNO_MORE_LENG | เลขที่ใบกำกับ > 20 | :1944 |
| RESERVE_ISCANCEL / RESERVE_BILLSTATUS / DATABASE_NOTHAVE_RESERVE | ใบจองยกเลิก / ออกบิลแล้ว / ไม่พบ | :2362-2402 |
| SALEORDER_ARCODE | กดเลือกใบเสนอราคาก่อนเลือกลูกค้า | :2662-2691 |
| WARN_AR_QUOTATION | ลูกค้ามีใบเสนอราคาค้าง (option ReminderQuotaion) | :2125-2199 |
| DEPOSIT_HAS_REFERENCED_TO_PO | (โหลดเอกสาร SCGProject) มัดจำถูกอ้างถึง SO/PO | :892-1056 |
| CANNOT_CANCEL / MSG_CANNOT_CANCEL / NOT_CONFIRM_NO | ยกเลิก/อนุมัติไม่ได้ | :2510-2541, :1623-1721 |

---

### [5329] บันทึกใบรับเงินล่วงหน้า
- **ตำแหน่งเมนู**: Bill > ต่อจาก 5327 (atlas-02-Bill.md:868); handler `OnBillSys6_01` (MainFrm.cpp:10522-10529); form `CBILLFrmEarnestSpecail` (BILLFrmEarnestSpecail.cpp, 1,661 บรรทัด); `IDD_BILL_FRM_EARNEST_SPECIAL` 349x227 (BC5Account.rc:13355); view key "300601", config `ARDepositSpecialConfig`, GL slip source 38, DataID `AR_CCDEPOSPEC`, ไม่พิมพ์ WTax, สิทธิ์ `ID_BILLSYS_TRANS_6_01` (:26-62); bill code "DEP" ถูกคอมเมนต์ทิ้ง (:33)
- **วัตถุประสงค์**: รับเงินล่วงหน้าจากลูกค้า **ไม่คิดภาษี** (ไม่มีเลขที่ใบกำกับ/VAT บนจอ) เก็บเป็นยอดเงินล่วงหน้าของลูกค้า (BCAR.EarnestAmount) ไว้หักในใบเสร็จรับเงิน หรือคืนเงินด้วยเมนู 5306

#### หน้าจอและช่องหัวเอกสาร (BCARDepositSpecial)
| # | ป้ายไทย | control / คอลัมน์ DB | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup / ตาราง | การตรวจ/คำนวณ | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | เลขที่ใบรับเงินล่วงหน้า: | IDC_ED_DOCNO / DocNo | text ≤35 | ใช่ | running บน BCARDepositSpecial.DocNo | — | ว่าง/ซ้ำ เหมือน 5327 | :155-199, :1548 |
| 2 | วันที่ออกเอกสาร: | IDC_ED_DOCDATE / DocDate | date | ใช่ | วันนี้ | — | งวดปิด | :388-597 |
| 3 | รหัสลูกค้า: + ชื่อ | IDC_ED_ARCODE / ArCode | lookup | ใช่ | — | BCAR | ดึงชื่อ + SaleCode เท่านั้น (ไม่ดึงภาษี/เครดิต) | :1368, :1439 |
| 4 | เลขที่ใบจอง: | IDC_ED_REFNO / RefNo | lookup | ไม่ | — | `SaleOrderConfig` กรอง `ArCode [AND SaleCode] AND IsCancel=0 AND SOStatus=1` | ต้องเลือกลูกค้าก่อน ("กรุณาเลือกรหัสลูกค้าก่อน"); เติม สกุลเงิน, อัตรา, แผนก, พนักงาน, NetAmount → รับชำระ และยอดคงเหลือ; ออกจากช่อง: ยกเลิก → RESERVE_ISCANCEL, ไม่พบ → DATABASE_NOTHAVE_RESERVE | :1392-1437, :1482 |
| 5 | วิธีการชำระเงิน: | IDC_ED_PAYMENT_TYPE / PaymentTypeCode | combo | ไม่ | — | `SELECT Code,Name FROM BCPaymentType` | — | :155-199 |
| 6 | รหัสพนักงาน: | IDC_ED_SALECODE / SaleCode | combo | ไม่ | จากลูกค้า | BCSale `WHERE ActiveStatus<>0` | — | :155-199 |
| 7 | แผนก: | IDC_ED_DEPART / DepartCode | combo | ไม่ | — | BCDepartment | — | 〃 |
| 8 | โครงการ: | IDC_ED_PROJECT / ProjectCode | combo | ไม่ | — | BCProject | — | 〃 |
| 9 | การจัดสรร: | IDC_ED_ALLOCATE / AllocateCode | combo | ไม่ | — | BCAllocate | — | 〃 |
| 10 | จ่ายส่วนที่เหลือภายใน: … วัน | IDC_ED_CREDITDAY / CreditDay | number | ไม่ | — | — | ↔ วันที่ส่งของ (เหมือน 5327 — ยังไม่ verify สูตรในไฟล์นี้) | :599-761 |
| 11 | วันที่ส่งของ: | IDC_ED_DUEDATE / DueDate | date | ไม่ | — | — | — | 〃 |
| 12 | หมายเหตุ: | IDC_ED_REMARK / MyDescription (500) | text | ไม่ | — | — | — | DDL |
| 13 | รหัสสกุลเงิน: | IDC_ED_CURRENCY / CurrencyCode | combo | ไม่ | สกุลหลัก | BCCurrency | — | :155-199 |
| 14 | อัตราแลกเปลี่ยน: | IDC_ED_EXCHANGE_RATE / ExChangeRate | number | ไม่ | — | — | — | 〃 |
| 15 | กลุ่มเอกสาร: | IDC_ED_DOCGROUP / BillGroup (35) | lookup | ไม่ | — | — | — | 〃 |
| 16 | รับชำระ: | IDC_ED_TOTALAMOUNT (m_edNetAmount) → TotalAmount = BeforeTaxAmount = NetAmount | number | ใช่ | — | — | ≤ 0 → CANNOT_SAVE_NO_AMOUNT; ตอนสร้างใหม่ ยอดคงเหลือ = ยอดนี้; ตอนแก้ ถ้าอนุมัติแล้ว → NOT_SAVE_ISCONFIRM | :1105-1122 |
| 17 | ยอดเงินคงเหลือ: | IDC_ED_BILLBALANCE / BillBalance | number (แสดง) | — | = รับชำระ | — | ลดเมื่อถูกใช้ | :1105-1122 |

ไม่มี: เลขที่/วันที่ใบกำกับ, ประเภท/อัตรา/ยอดภาษี, ชื่อ-ที่อยู่ออกบิล, F3, F6 (DDL BCARDepositSpecial ไม่มี TaxNo/TaxDate/TaxType/TaxRate/TaxAmount แต่มี PettyCashAmount, PaymentTypeCode)

#### ตารางรายการ (grid)
N/A

#### ยอดรวมท้ายเอกสาร / สูตร
- TotalAmount = BeforeTaxAmount = NetAmount = ยอดรับชำระ; BillBalance = ยอดรับชำระ (ตอนสร้าง) (:599-761, :1105-1122)
- ต้องรับเงินครบตาม §0.2; หน้าต่าง F7 เปิดโดยอนุญาตค่าใช้จ่ายอื่น (TRUE) (:388-597)

#### ปุ่ม แถบเครื่องมือ คีย์ลัด (PreTranslateMessage :1127-1187; OnMore :1189)
F2 GL, F7 รับเงิน, F11 บันทึก, F12 เส้นทางเอกสาร, Shift+F5 อนุมัติ, Ctrl+F5 ยกเลิก; เมนู Key: F2GL, F7_BL_RECEIVE, F11_SAVE, F12_CHECK_DOCUMENT, SHIFT_F5, CTRL_F5 — **ไม่มี F3/F6** (ปุ่ม Recur ในแถบมาตรฐาน: ยังไม่ verify flag; คอลัมน์ RecurName ถูกบันทึก)

#### สถานะเอกสารและ workflow
1. บันทึก → รับเงินครบ → `arRecalculate` ยอด EarnestAmount ของลูกค้า (ทั้งรหัสเก่าและใหม่)
2. **IsConfirm ของใบนี้ = "ถูกใช้แล้ว"** — ถูกตั้งอัตโนมัติโดย `OnProcessEarnest` เมื่อ BillBalance ≠ NetAmount (ใช้บางส่วน/หมด) และกลับเป็น 0 เมื่อเท่ากัน (arfrmbcreceipt1view.cpp:4724-4766; BILLFrmEarnestSpecailReturn.cpp:1424-1451) — นอกจากนี้กด Shift+F5 อนุมัติเองได้
3. **การใช้เงินล่วงหน้า** มี 2 ทาง:
   - **ใบเสร็จรับเงิน AR** (`CARFrmBCReceipt1View`, view key "50009") เมนู Key "ใบรับเงินล่วงหน้า" (arfrmbcreceipt1view.cpp:1884/1908, `ReceiveEarnest()` :4636) กรอง `BCARDepositSpecial.ArCode = … AND IsCancel = 0 AND BillBalance > 0` (:4662-4664) → แถว BillType = 8 หักจากยอดรับ (:1119), หัวใบเสร็จเก็บ SumOfEarnest (:1061); `OnProcessEarnest` (:4724-4766): `UPDATE BCARDepositSpecial SET IsConfirm = 0, BillBalance = BillBalance − (mark × PayAmount)` แล้วตั้ง IsConfirm ตามข้างต้น + `arRecalculate CALC_EARNEST_BALANCE`
   - **คืนเงิน** เมนู 5306
4. ยกเลิก (`OnConfirmCancel` :1588-1633): งวดปิด, IsPostGL, IsConfirm, ถูกใช้ในใบคืนหรือในใบเสร็จ (billtype 8/66) → CANNOT_CANCEL_USE
5. Document flow (:1560-1585): BCReturnDepSpecial (EarnestNo = เอกสารนี้, IsCancel=0) UNION BCReceiptSub1 (InvoiceNo = เอกสารนี้, BillType=8)

**ยอดเงินล่วงหน้าคงเหลือของลูกค้า** `calcEarnetAmount` (BCARProcess.cpp:873-933):
```
BCAR.EarnestAmount = Σ round(BCARDepositSpecial.netamount × exchangerate)
                   − ( Σ BCReturnDepSpecial.netamount × rate + Σ bcreceiptsub1.payamount × invexchangerate [billtype 8 หรือ 66] )   — ทุกตัว iscancel = 0
```

#### ผลข้างเคียง (OnSaveData :599-761)
| สิ่งที่เกิด | รายละเอียด | หลักฐาน |
|---|---|---|
| หัวเอกสาร | `BCARDepositSpecial`: DocNo, DocDate, ArCode, RefNo, DepartCode, CreditDay, DueDate, SaleCode, AllocateCode, ProjectCode, BillGroup, MyDescription, GLTransData, GLFormat, RecurName, PaymentTypeCode, CurrencyCode, ExChangeRate, ExcessAmount1/2, SumCashAmount, SumChqAmount, SumCreditAmount, ChargeAmount, SumBankAmount, TotalAmount = BeforeTaxAmount = NetAmount, BillBalance, OtherIncome, OtherExpense, creator/editor | :599-761 |
| ยอดลูกค้า | `arRecalculate` → BCAR.EarnestAmount (รหัสเก่า+ใหม่); BCAR.ChqOnHand | :599-761 |
| รับเงิน | `OnSaveReceive(8, …)` → BCRecMoney SaveFrom 8 (§0.3) | 〃 |
| GL | `m_nSource = 38, m_nGLGroup = 40, ProcessConditionGroup5` (TOHOME: `CGLSpecial::ProcessConditionGroup71`) หรือบันทึกมือ source 38 | 〃 |
| ยกเลิก | `UPDATE BCARDepositSpecial SET BillBalance = 0.0, IsCancel = 1`; recalc EarnestAmount; `CancelOutputTax(0, doc, 5)`; `DeleteARWTax`; `DeleteReceive(8, doc)`; `ReverseAccount(doc, 38)` | :297-339 |
| ลบ | DELETE; recalc; ChqOnHand −; `DeleteOutputTax(0, doc, 5)`; `DeleteARWTax`; `DeleteReceive(8)`; `DeleteAccount(doc, 38)` | :341-386 |
- ข้อสังเกต (ห้ามลอก): ยกเลิก/ลบเรียกลบภาษีขายด้วย source **5** (ของใบมัดจำ) ทั้งที่จอนี้ไม่มีภาษี (:297-339, :341-386)

#### Guard / ข้อความเตือนสำคัญ (key — ความหมายอนุมาน)
| key | เงื่อนไข | หลักฐาน |
|---|---|---|
| CANNOT_DELETE_POSGL / CANNOT_SAVE_POSTGL | ผ่าน GL | :220-295, :388-597 |
| CHQNOT_EDIT_DELETE_CANCEL | `IsChqRefence(doc, 8, 0)` หรือ `(doc, 2, 0)` | :220-295 |
| DELETE_UPDATE_CHQ / CANCEL_UPDATE_CHQ | เช็คที่รับถูกใช้ต่อ (ลบ/แก้) | :220-295, :388-597 |
| DELETE_UPDATE_CREDITCARD / CANCEL_UPDATE_CREDITCARD | รายการบัตรเครดิตถูกอ้างอิง | 〃 |
| NOT_*_ISCANCEL / NOT_*_ISCONFIRM / NOT_*_BILLBALANCE | ยกเลิก / ใช้แล้ว / ยอดถูกตัด / IsReturnMoney | 〃 |
| CANNOT_SAVE_NO_AMOUNT | ยอดรับชำระ ≤ 0 | :388-597 |
| AMOUNT_NOT_EQUAL_RECEIVE | รับเงินไม่ครบ → เปิด F7 | 〃 |
| CANNOT_CANCEL_USE | ถูกใช้ในใบคืน/ใบเสร็จแล้ว | :1588-1633 |
| "กรุณาเลือกรหัสลูกค้าก่อน" (ไทยในโค้ด) | เลือกใบจองก่อนเลือกลูกค้า | :1392-1437 |

---

### [5306] บันทึกคืนเงินรับล่วงหน้า
- **ตำแหน่งเมนู**: Bill > ต่อจาก 5329 (atlas-02-Bill.md:903); handler `OnBillsysTrans14` (MainFrm.cpp:2901-2908); form `CBILLFrmEarnestSpecailReturn` (BILLFrmEarnestSpecailReturn.cpp, 1,528 บรรทัด); `IDD_BILL_FRM_EARNEST_RETURN` 359x240 (BC5Account.rc:14531); view key "300602", config `ARReturnDepositSpecialConfig`, GL slip source 39, DataID `AR_RETDEPOSPEC`, bill code "DEP" (:32), สิทธิ์ `ID_BILLSYS_TRANS_14` (:24-61)
- **วัตถุประสงค์**: จ่ายคืนเงินรับล่วงหน้า (จาก 5329) ให้ลูกค้า ทั้งก้อนหรือบางส่วน ลดยอดคงเหลือของใบรับเงินล่วงหน้า

#### หน้าจอและช่องหัวเอกสาร (BCReturnDepSpecial)
| # | ป้ายไทย | control / คอลัมน์ DB | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup / ตาราง | การตรวจ/คำนวณ | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | เลขที่เอกสาร: | IDC_ED_DOCNO / DocNo | text ≤20 | ใช่ | running บน BCReturnDepSpecial | — | ว่าง/ซ้ำ | :132, :334-483 |
| 2 | วันที่เอกสาร: | IDC_ED_DOCDATE / DocDate | date | ใช่ | วันนี้ | — | งวดปิด | :334-483 |
| 3 | เลขที่ใบรับเงินล่วงหน้า: | IDC_ED_EARNESTNO / EarnestNo | lookup | ใช่ (โดยปริยาย) | — | `ARDepositSpecialConfig` กรอง `BillBalance > 0.0 [AND CreatorCode='OOMDTS' ถ้าติ๊ก SAP] [AND Code LIKE ลูกค้า%]` | เติม: ลูกค้า, ชื่อ, พนักงาน, แผนก, โครงการ, การจัดสรร, สกุลเงิน, อัตรา, เงินล่วงหน้าสุทธิ = NetAmount, เงินล่วงหน้าคงเหลือ = BillBalance, ยอดลูกหนี้คงเหลือ = BCAR.DebtAmount; **พิมพ์เลขเอง (`OnKillfocusEdEarnest` :1076) ไม่ตรวจ BillBalance > 0 หรือ IsCancel** | :1033-1074, :1076 |
| 4 | กลุ่มเอกสาร: | IDC_ED_BILLGROUP / BillGroup | lookup | ไม่ | — | — | — | :485-624 |
| 5 | คืนเงินให้กับ: + ชื่อ | IDC_ED_ARCODE / ArCode | text (มาจากใบรับเงินล่วงหน้า; link BCAR ถูกคอมเมนต์ทิ้ง) | ใช่ | จากข้อ 3 | BCAR | ว่าง → CANNOT_SAVE_NOTArCode; ไม่พบ → NOT_FOUND_AR | :132, :334-483 |
| 6 | รหัสพนักงาน: | IDC_ED_SALECODE / SaleCode | combo | ไม่ | จากข้อ 3 | BCSale | — | :1033-1074 |
| 7 | แผนก: | IDC_ED_DEPART / DepartCode | combo | ไม่ | 〃 | BCDepartment | — | 〃 |
| 8 | โครงการ: | IDC_ED_PROJECT / ProjectCode | combo | ไม่ | 〃 | BCProject | — | 〃 |
| 9 | การจัดสรร: | IDC_ED_ALLOCATE / AllocateCode | combo | ไม่ | 〃 | BCAllocate | — | 〃 |
| 10 | รหัสสกุลเงิน: | IDC_ED_CURRENCY / CurrencyCode | combo | ไม่ | 〃 | BCCurrency | — | 〃 |
| 11 | อัตราแลกเปลี่ยน: | IDC_ED_EXCHANGE_RATE / ExChangeRate | number | ไม่ | 〃 | — | — | 〃 |
| 12 | คำอธิบาย: | IDC_ED_REMARK / MyDescription | text | ไม่ | — | — | — | :485-624 |
| 13 | เงินล่วงหน้าสุทธิ: | IDC_ED_EARNEST_AMOUNT / EarnestAmount | number (แสดง) | — | NetAmount ของใบรับเงินล่วงหน้า | — | — | :1033-1074 |
| 14 | เงินล่วงหน้าคงเหลือ: | IDC_ED_EARNEST_BALANCE / EarnestBalance | number (แสดง) | — | BillBalance ของใบรับเงินล่วงหน้า | — | บันทึกเป็น EarnestAmount − ยอดคืน | :485-624 |
| 15 | จำนวนเงิน: | IDC_ED_AMOUNT / NetAmount = TotalAmount = BillBalance | number | ใช่ | — | — | > เงินล่วงหน้าคงเหลือ → **รีเซ็ตเป็นยอดคงเหลือเงียบ ๆ** (key CANNOT_EDIT_AMOUNT_OVER สร้างข้อความแต่ไม่แสดง MessageBox); ≤ 0 → CANNOT_SAVE_WORNG | :1103-1124, :334-483 |
| 16 | :ยอดลูกหนี้คงเหลือ (ป้ายใน .rc มี ":" นำหน้า) | IDC_ED_ARBALANCE (แสดงอย่างเดียว) | number | — | BCAR.DebtAmount | BCAR | — | :1033-1074 |
| 17 | (checkbox) IS_SAP "SAP" | IDC_CHECK_SAP / IsSAP | checkbox | — | — | — | แสดงเฉพาะ config `oomdts = 1` | :197-208 |

DDL มี CreditDay/DueDate แต่จอนี้ไม่เขียน (:485-624)

#### ตารางรายการ (grid)
N/A

#### ยอดรวมท้ายเอกสาร / สูตร
- NetAmount = TotalAmount = BillBalance = จำนวนเงินคืน; EarnestBalance = EarnestAmount − จำนวนเงินคืน (:485-624)
- ต้องจ่ายเงินครบตาม §0.2 โหมดจ่าย (`IsMoneyOk(1, …)`) มิฉะนั้น AMOUNT_NOT_EQUAL_PAYMENT + เปิดหน้าต่างจ่ายเงิน

#### ปุ่ม แถบเครื่องมือ คีย์ลัด (PreTranslateMessage :1262-1321; OnMore :1126)
F2 GL, F7 **จ่ายเงิน** (`OnPay`), F11 บันทึก, F12 เส้นทางเอกสาร, Shift+F5 อนุมัติ (ข้ามถ้ายกเลิกแล้ว), Ctrl+F5 ยกเลิก; เมนู Key: F2GL, F7PAYMENT, F11_SAVE, F12_CHECK_DOCUMENT, SHIFT_F5, CTRL_F5; Info (:1197) แสดงยอดเงินสดจ่าย, เช็คจ่าย, โอน, เงินสดย่อย, เงินเกิน 2 ช่อง, สุทธิ, คงเหลือ (ป้ายไทย Info ยังไม่ verify ว่า verbatim)

#### สถานะเอกสารและ workflow
1. เลือกใบรับเงินล่วงหน้า → กรอกจำนวนเงินคืน → F7 จ่ายเงิน (เงินสด/โอน/เช็คจ่าย/เงินสดย่อย) → บันทึก
2. บันทึกแล้ว `OnProcessEarnest(doc, +1)` (:1424-1451): `UPDATE BCARDepositSpecial SET IsConfirm = 0, BillBalance = BillBalance − (mark × NetAmount) WHERE DocNo = EarnestNo` แล้ว IsConfirm = 1 ถ้า BillBalance ≠ NetAmount ไม่งั้น 0 → recalc BCAR.EarnestAmount
3. อนุมัติ (Shift+F5) → แก้/ลบ/ยกเลิกไม่ได้
4. ยกเลิก (`OnConfirmCancel` :1474-1500: งวดปิด, IsPostGL, IsConfirm) → `BillCancelUpdateData` (:265-296)
5. Document flow (:1399-1422) — **บั๊ก**: query `BCAPReturnDepSpecial`/`BCPaymentSub` (ตารางฝั่งเจ้าหนี้) (:1408)

#### ผลข้างเคียง (OnSaveData :485-624 — อยู่ใน `StartTransaction`/`Commit`)
| สิ่งที่เกิด | รายละเอียด | หลักฐาน |
|---|---|---|
| หัวเอกสาร | `BCReturnDepSpecial`: DocNo, DocDate, ArCode, DepartCode, GLFormat, MyDescription, IsSAP, CurrencyCode, ExChangeRate, EarnestNo, EarnestAmount, EarnestBalance, NetAmount = TotalAmount = BillBalance, ExcessAmount1/2, OtherIncome, OtherExpense, SumOfWTax, SumCashAmount, SumChqAmount, PettyCashAmount, SumBankAmount, AllocateCode, ProjectCode, BillGroup, RecurName, SaleCode, creator/editor | :485-624 |
| เช็ค | ย้อนยอดเก่าด้วย `OnDepositUpdateChqAmount(1, …)` = **BCAP**.ChqOnHand (:492) แต่บวกยอดใหม่ด้วย mode 0 = **BCAR**.ChqOnHand (:588) — **ไม่สอดคล้อง (บั๊ก ห้ามลอก)** | :492, :588 |
| จ่ายเงิน | `OnSavePayment(9, …)` → BCPayMoney SaveFrom 9; เช็คจ่าย → BCChqOut; ธนาคาร −; เงินสดย่อย − | :485-624 |
| ใบรับเงินล่วงหน้า | `OnProcessEarnest(doc, +1)` ลด BillBalance + recalc EarnestAmount | :1424-1451 |
| GL | AutoPost: source 39, กลุ่ม 41, `ProcessConditionGroup44` (§0.5) หรือบันทึกมือ source 39 | :485-624; gltlsprocess.cpp:15991 |
| ยกเลิก | `UPDATE BCReturnDepSpecial SET BillBalance = 0.0, IsCancel = 1`; `OnProcessEarnest(doc, −1)` คืนยอดให้ใบรับเงินล่วงหน้า; recalc; `ReverseAccount(doc, 39)`; **ไม่ลบรายการจ่ายเงิน** (`DeletePayment` ถูกคอมเมนต์) | :265-296 |
| ลบ | `OnProcessEarnest(−1)`; recalc; DELETE; `OnDepositUpdateChqAmount(1, …, −1)` (BCAP); `DeletePayment(9, doc)`; `DeleteAccount(doc, 39)` | :298-333 |

#### Guard / ข้อความเตือนสำคัญ (key — ความหมายอนุมาน)
| key | เงื่อนไข | หลักฐาน |
|---|---|---|
| (งวดปิด) | วันที่อยู่ในงวดปิด | :230-263, :334-483 |
| CHQNOT_EDIT_DELETE_CANCEL | `IsChqRefence(doc, 9, 1)` เช็คจ่ายถูกอ้างอิง | :230-263 |
| NOT_DELETE_ISCANCEL / NOT_DELETE_ISCONFIRM | ลบเอกสารยกเลิก/อนุมัติแล้ว — **ไม่มีตรวจ IsPostGL ตอนลบ** (ช่องโหว่) | :230-263 |
| NOT_SAVE_ISCANCEL / NOT_SAVE_ISCONFIRM | แก้เอกสารยกเลิก/อนุมัติแล้ว | :334-483 |
| CANNOT_SAVE_NOTArCode / NOT_FOUND_AR | ลูกค้า | 〃 |
| AMOUNT_NOT_EQUAL_PAYMENT | จ่ายเงินไม่ครบ → เปิดหน้าต่างจ่ายเงิน | 〃 |
| CANNOT_SAVE_WORNG | จำนวนเงิน ≤ 0 | 〃 |
| CANNOT_EDIT_AMOUNT_OVER | คืนเกินยอดคงเหลือ (รีเซ็ตโดยไม่แสดงข้อความ) | :1103-1124 |
| (พิมพ์) | `OnHKPrint` ตรวจกับตาราง BCARDepositSpecial แทน BCReturnDepSpecial — บั๊กคัดลอก | :1507 |

---

---

## ส่วน 4 — [5328] บันทึกขายสินค้า,บริการ

> แหล่งอ้างอิงทั้งหมดอยู่ในสำเนาที่ถอดรหัสแล้ว `scratch/champ-atlas/dec-sa/champ/` (อ้างแบบย่อ `ไฟล์:บรรทัด`) ไฟล์ไลบรารีใช้ `lib-frmwrk/`, `lib-comm/` และ `bcprocess.cpp` (= `D:\project-champ\champ-lib\BC5PROCESS\bcprocess.cpp`)
> คำว่า "(อนุมาน)" หมายถึง ข้อความไทยจริงอยู่ในตาราง `myresource` ของ DB ซึ่งไม่มีใน repo ความหมายจึงอนุมานจากโค้ด ไม่ใช่ถ้อยคำจริงของ Champ

### [5328] บันทึกขายสินค้า,บริการ
- **ตำแหน่งเมนู**:
  - เมนู Bill (ระบบขาย) › บันทึกขายสินค้า,บริการ เป็นรายการเดี่ยว ไม่อยู่ในกลุ่มย่อย ลำดับถัดจาก 5306 บันทึกคืนเงินรับล่วงหน้า (atlas-02-Bill.md:931)
  - ID `ID_BILLSYS_TRANS_6` เรียก `OnBillsysTrans6` (MainFrm.cpp:345 ON_COMMAND, MainFrm.cpp:1228-1238) ซึ่งตรวจ `IsMenuEnable` ก่อนเปิด
- **Form ที่เปิดจริง (build ปกติ)**:
  - `CChmFrmBilling` : `CMyEspeciallyFormView` (ChmFrmBilling.cpp) IDD_BILL_FRM_BILL 444x290 (BC5Account.rc:16490)
  - build ปกติเปิดตัวนี้เพราะ `CHM_PROJ` ถูก define ใน config CHMDebug/CHMRelease (BC5Account.vcproj:217, :1623 ; BC5Account.dsp:124, :150)
  - variant ที่ไม่ใช่ CHM: `CBILLFrmBilling` (billfrmbilling.cpp ~10,000 บรรทัด; atlas-02-Bill.md:1399) — ไม่ได้ลงรายละเอียดในเอกสารนี้
- **แท็บของจอ** (ChmFrmBilling.cpp:172-308):
  - แท็บ 0 `MST_HEADER` (ข้อมูลหลัก — ไทยจาก comment ในโค้ด) = `CChmFrmBillingPage1` IDD_BILL_FRM_BILL_PAGE1 626x277 (BC5Account.rc:16503)
  - แท็บ 1 `MST_OTHER` (รายการอื่นๆ — ไทยจาก comment ในโค้ด) = `CChmFrmBillingPage2` IDD_BILL_FRM_BILL_PAGE2 540x309 (BC5Account.rc:16611)
  - ข้อมูลที่ทั้งสองแท็บใช้ร่วมกันอยู่ใน `CChmBilling_InfoEx billing_Info` (ChmBilling_InfoEx.cpp)
- **ชื่อจอ**: `GetMenuRes("30007")` มี comment ว่า "บันทึกขายสินค้า, บริการ" (ChmFrmBilling.cpp:39-95)
- **ค่าตั้งของจอ**:
  - TransConfig = `BillingTransConfig`, help = `BILLING_HELP`
  - ตั้ง `m_nPrintSlipGLSource=6`, `m_strSubFindTable="BCARInvoiceSub"`, `m_bCanUseDocumentFlow=FALSE`
  - สิทธิ์ผูกกับ `SetUserAuthorization(..., ID_BILLSYS_TRANS_6)` (ChmFrmBilling.cpp:39-95)
- **วัตถุประสงค์**:
  - เอกสารขายประจำวันฉบับเดียวที่ทำครบทุกอย่าง ได้แก่ ใบขาย/ใบส่งของ, ใบกำกับภาษีขาย, ตัดสต๊อก, ตั้งหนี้ลูกหนี้ (กรณีขายเชื่อ) หรือรับเงิน (กรณีขายสด) และลงบัญชี GL
  - รายการค้นหาไม่รวมบิลจาก POS (filter `BCARInvoice.PosStatus = 0`, ChmFrmBilling.cpp:172-308)

#### ประเภทเอกสาร (BillType) และประเภทภาษี (TaxType)
| ค่า | BillType: key บนจอ | ความหมาย | สด/เชื่อ | กลุ่ม GL format | หลักฐาน |
|---|---|---|---|---|---|
| 0 | `CB_DOCTYPE_1` | ขายสด (อนุมาน) | สด | 8 สินค้าสด | lib-frmwrk/companyinfo.cpp:111 ; ChmFrmBilling.cpp:4240-4252 |
| 1 | `CB_DOCTYPE_2` | ขายเชื่อ (อนุมาน) | เชื่อ | 9 สินค้าเชื่อ | เดียวกัน |
| 2 | `CB_DOCTYPE_3` | ขายบริการสด (อนุมาน) | สด | 10 บริการสด | เดียวกัน |
| 3 | `CB_DOCTYPE_4` | ขายบริการเชื่อ (อนุมาน) | เชื่อ | 11 บริการเชื่อ | เดียวกัน |

- ความหมาย 0–3 มาจาก comment ใน `companyinfo.cpp:111` ส่วนป้าย combo `CB_DOCTYPE_1..4` ไม่พบใน keyThai
- ค่าเริ่มต้นคือ `DefItemBillType` (ChmBilling_InfoEx.cpp:202-356)
- พิมพ์ฟอร์มใช้ key ชุดเดียวกัน (`BILLTYPE`, ChmFrmBilling.cpp:4388-4396)
- BillType 3 ล้างเลขที่ใบกำกับทิ้งตอนขึ้นเอกสารใหม่ (ChmFrmBilling.cpp:5984-6068)

| ค่า | TaxType: key | ความหมาย | สูตรภาษี | หลักฐาน |
|---|---|---|---|---|
| 0 | `CB_TAXTYPE_1` | แยกนอก | ภาษี = ฐาน×อัตรา/100 แล้วบวกเพิ่ม | ChmFrmBillingPage1.cpp:304-311 ; ไทยจาก comment ARFrmARMasterViewPage2.cpp:128-130 |
| 1 | `CB_TAXTYPE_2` | ภาษีรวมใน | ภาษี = ยอด×อัตรา/(100+อัตรา) | เดียวกัน |
| 2 | `TAX_ZERO` | ภาษีอัตราศูนย์ | ภาษี 0 และเก็บยอดลง `ZeroTaxAmount` | เดียวกัน ; ChmFrmBilling.cpp:3275-3992 |

- ค่าเริ่มต้นคือ `theApp.m_CompanyInfo.DefaultTaxType`
- อัตราภาษีใช้ `DefaultTaxRate` ของลูกค้าก่อน ถ้าไม่มีจึงใช้ของบริษัท ประเภท 2/3 อัตราเป็น 0 (ChmFrmBillingPage1.cpp:495-545)

#### หน้าจอและช่องหัวเอกสาร — แท็บ 1 "ข้อมูลหลัก" (Page1 ตามตำแหน่ง y บนจอ บน→ล่าง ซ้าย→ขวา; atlas-02-Bill.md:938-966)
| # | ป้ายไทย (verbatim) | control → คอลัมน์ DB | ชนิด | บังคับ? | ค่าเริ่มต้น | lookup/ตาราง | การตรวจ/คำนวณเมื่อออกจากช่อง | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | ปุ่ม "ตรวจสอบสต๊อกย้อนหลัง" | IDC_BTTN_STOCKREWORD | ปุ่ม | – | – | stock ณ วันที่เอกสาร | ตรวจยอดคงเหลือย้อนหลังตามวันที่เอกสาร | atlas:941 ; Page1 OnKillfocusEdDocDate :4732-4794 |
| 2 | SCG Family ID/เบอร์โทร : | IDC_ED_SCGID → `SCGID` | text | ไม่ | จากลูกค้า | – | **เฉพาะลูกค้ารายหนึ่ง (variant)** — ตัดออกได้ | atlas:942 ; Page1:6483 |
| 3 | ประเภทการขาย: | IDC_ED_DOCTYPE → `BillType` | combo 4 ค่า | ใช่ | `DefItemBillType` | – | OnSelchangeDocType: เปลี่ยนกลุ่ม GL format; ถ้าเปิด AutoBillDocGroup จะรันเลขใหม่; ขายสด → เครดิต 0 วัน + วันครบกำหนด = วันที่เอกสาร; ขายเชื่อ → เปิดช่องเครดิต และเตือนวงเงินตามสถานะลูกค้า; ถ้าเปิด `m_nSaleBaseOnSO` และเป็นขายเชื่อ จะปิดตาราง (ต้องดึงจากใบสั่งขาย) | Page1:427-493, :432 |
| 4 | ประเภทภาษี: | IDC_ED_TAXTYPE → `TaxType` | combo 3 ค่า | ใช่ | `DefaultTaxType` | – | OnSelchangeTaxType: ตั้งอัตราภาษีใหม่ และคำนวณ NetAmount/HomeAmount ของทุกแถวใหม่ | Page1:495-545 |
| 5 | (จำนวนครั้งที่พิมพ์) | IDC_LB_PRINTCOUNT ← `PrintCount` | label | – | 0 ตอนสร้าง | – | แสดงอย่างเดียว | atlas:943 ; ChmFrmBilling.cpp:3275-3992 |
| 6 | ปุ่ม "เลือกใบเสนอราคา" | IDC_BTTN_QUATATION | ปุ่ม | – | – | BCQuotation | ดึงใบเสนอราคามาเป็นรายการ | Page1:3749 (รายละเอียด ยังไม่ verify) |
| 7 | ปุ่ม "Get Promotion Msg" | IDC_BTTN_GETPROMOTION | ปุ่ม | – | – | – | **variant ลูกค้ารายหนึ่ง** | Page1:6364 |
| 8 | เลขที่ใบขายสินค้า: | IDC_ED_DOCNO → `DocNo` | text + running | ใช่ | รันเลขจากรูปแบบเอกสาร (`SetOption(db,"BCARInvoice",m_strDocFormat,&m_saDocGroup,"docno")`); ถ้าเปิด `AutoBillDocGroup==1` จะใช้กลุ่มเอกสารตาม BillType | BCARInvoice | ห้ามว่าง และห้ามซ้ำ; ล็อกไม่ให้แก้ถ้าไม่มีสิทธิ์ `bCanChangeBillNo` | ChmFrmBilling.cpp:172-308, :5984-6068 (:6028), :4092 |
| 9 | วันที่ออกเอกสาร: | IDC_ED_DOCDATE → `DocDate` | date (BME_DATE_TH_AUTO) | ใช่ | วันทำงาน | – | คำนวณวันครบกำหนด/วันส่งของ; ถ้าลงวันย้อนหลังจะตรวจสต๊อกใหม่ (session "30007"+docno); ปิดช่องถ้า `IsCanEditDocDate==0` และผู้ใช้ไม่ใช่ sa | ChmFrmBilling.cpp:172-308 ; Page1:4732-4794 |
| 10 | เลขที่ใบกำกับภาษี: | IDC_ED_TAXNO → `TaxNo` | text + running | มีเงื่อนไข | = DocNo (BillType 3 เว้นว่าง) | BCOutputTax | ยาวไม่เกิน 20 (`TAXNO_MORE_LENG` ถ้าเกินจะถูกตัด); ห้ามซ้ำใน BCOutputTax (`DUPLICATE_TAXNO`); ถ้าตั้ง `WarnTaxNoEmpty` ห้ามเว้นว่างเมื่อยอดไม่เป็น 0; ทุกแถวคิด NetAmount ใหม่ | Page1:4796-4853 ; ChmFrmBilling.cpp:1244, :5984-6068 |
| 11 | วันที่ใบกำกับ: | IDC_ED_TAXDATE → เก็บใน `BCOutputTax.TaxDate` (ไม่มีคอลัมน์นี้ใน BCARInvoice) | date | ใช่ถ้ามีเลขที่ใบกำกับ | วันที่เอกสาร | – | `CHECK_TAXDATE` | ChmFrmBilling.cpp:1594, :4400-4416 |
| 12 | ปุ่ม "Ctrl+F3=เลือกใบสั่งขาย/สั่งจองสินค้า" | IDC_BTTN_SALE_ORDER | ปุ่ม | – | – | BCSaleOrder | ดูหัวข้อ "อ้างอิงเอกสาร" | Page1:2701 |
| 13 | ปุ่ม "Get Coupon" | IDC_BTTN_GETPROMOTION2 | ปุ่ม | – | – | – | **variant** | Page1:6431 |
| 14 | รหัสลูกค้า: + ชื่อ | IDC_ED_ARCODE → `ArCode` ; IDC_ED_ARNAME (แสดงชื่อ) | lookup (`*`+barcode ได้) | ใช่ (must-entry) | – | BCAR (เฉพาะ active) ผ่านหน้าค้นหา `ArTransConfig` | loadAr เติม SaleCode, ผู้ติดต่อ, เขตขาย, ที่อยู่บิล, เลขผู้เสียภาษี, ขนส่ง, ระดับราคา, ส่วนลด; ถ้ามีใบเสนอราคาค้าง → `WARN_AR_QUOTATION` (yes/no); ถ้ามีใบสั่งขายค้าง → ถาม "ต้องการที่จะใช้งานด้วยหรือไม่?"; ถ้าตั้ง `NotChangeSelectAR` จะล็อกลูกค้าเมื่อมีรายการแล้ว | Page1:786-873, :4522-4703, :1903-2037 ; InfoEx:145-200 |
| 15 | (บรรทัดรายละเอียดสินค้า) | IDC_ED_ITEMDETAIL | text แสดงอย่างเดียว | – | – | – | แสดงข้อมูลสินค้าของแถวที่เลือก | atlas:950 |
| 16 | ตารางรายการ | m_BInputListWnd | grid | ต้องมีอย่างน้อย 1 แถว | – | – | ดูหัวข้อถัดไป | Page1:123-351 |
| 17 | รายการของแถม | IDC_REP_LIST (m_ListCtrl) | list (checkbox) | – | – | โปรโมชัน | คอลัมน์ `PREMIUM_ITEM_CODE`, `LIST_OF_PREMIUM` (กว้าง 480), `UNITTYPE` + คอลัมน์ซ่อน WHCode/SHCode/UnitCode/RefRow | Page1:5907-5976, :5985 |

ยอดท้ายเอกสาร (คอลัมน์ขวา x=368) และช่องประกอบ (คอลัมน์ซ้าย) อยู่ที่ระดับเดียวกันบนจอ:

| # | ป้ายไทย (verbatim) | control → คอลัมน์ DB | ชนิด | ค่าเริ่มต้น / แหล่ง | การตรวจ/คำนวณ | หลักฐาน |
|---|---|---|---|---|---|---|
| 18 | รวมมูลค่าสินค้า: | IDC_ED_SUMDETAIL (m_edSumOfNormalTax) → `SumOfItemAmount` | number (AmountPoint) | Σ แถวที่ไม่ยกเว้นภาษี + ค่าขนส่ง | แก้ได้เมื่อมีสิทธิ์ `bCanChangeBillAmount`; ออกจากช่องแล้วเรียก `BillCalc(1)` | Page1:4883 ; InfoEx:403-461 |
| 19 | มูลค่าสินค้ายกเว้นภาษี: | IDC_ED_EXCEPTTAX → `ExceptTaxAmount` | number | Σ แถว ExceptTax=1 (เมื่อบริษัทเปิด ExceptTax) | – | Page1:652-775 |
| 20 | ส่วนลด%,บาท: + ปุ่ม "%" | IDC_BN_DISCOUNT_WORD (เปิด `CLIBDlgDiscount`) + IDC_ED_DISC_AMOUNT → `DiscountWord`, `DiscountAmount` | text สูตรส่วนลด | ส่วนลดท้ายบิลจาก BCAR.discount | ปิดช่องถ้าไม่มีสิทธิ์ `bCanChangeBillDiscount` หรือ `BillDiscount==0`; ออกจากช่องแล้ว discountword = จำนวนเงิน (mode 2) | Page1:4484-4504, :4896-4914 |
| 21 | ภาษีมูลค่าเพิ่ม: | IDC_ED_TAXAMOUNT → `TaxAmount` | number | คำนวณ | – | BCStools.cpp:155 |
| 22 | มูลค่ารวมภาษี: | IDC_ED_SUMTAXAMOUNT (m_edTotalAmount) → `TotalAmount` | number | คำนวณ | – | – |
| 23 | มูลค่าเงินสุทธิ: | IDC_ED_SUMAMOUNT (m_edNetAmount) → `NetDebtAmount` | number | คำนวณ (หักภาษี ณ ที่จ่ายและมัดจำแล้ว) | – | InfoEx:403-461 |
| 24 | รหัสพนักงานขาย: | IDC_ED_SALECODE → `SaleCode` (+`UserGroup` จาก BCSale) | combo | จาก BCAR.SaleCode | BCSale ที่ ActiveStatus<>0 | Page1:123-351 |
| 25 | เงื่อนไขการส่งมอบสินค้า: | IDC_ED_CONDPAY → `IsConditionSend` | combo: รับเอง(0) / ส่งให้(1) — ไทยจาก comment | 0 (variant ขนส่งตั้งเป็น 1) | – | Page1:123-351 |
| 26 | รหัสแผนก: | IDC_ED_DEPART → `DepartCode` | combo | – | BCDepartment | เดียวกัน |
| 27 | โครงการ: | IDC_ED_PROJECT → `ProjectCode` | combo | – | BCProject | เดียวกัน |
| 28 | " :" (ป้ายหายใน .rc) | IDC_ED_JOBID2 (m_edQARefno) → `QARefno` | text | – | **variant** (OnCheckQARefNo) | Page1:6505-6530 |
| 29 | น้ำหนักสุทธิ … KG. | IDC_ED_WEIGHT → `Netweight` | number 2 ตำแหน่ง | Σ น้ำหนักแถว | แสดงอย่างเดียว | Page1:652-775 |

#### หน้าจอแท็บ 2 "รายการอื่นๆ" (Page2; atlas-02-Bill.md:967-992; ChmFrmBillingPage2.cpp)
| # | ป้ายไทย (verbatim) | control → คอลัมน์ DB | ชนิด | ค่าเริ่มต้น / lookup | การตรวจ/คำนวณ | หลักฐาน |
|---|---|---|---|---|---|---|
| 1 | รหัสผู้ติดต่อ: | IDC_ED_CONTACT → `ContactCode` | combo | BCAR.DefContactCode ; BCContactList MyType=0 parentcode=ลูกค้า | – | Page2:91-162 |
| 2 | รหัสสกุลเงิน: | IDC_ED_CURRENCY → `CurrencyCode` | combo | BCCurrency | ปิดช่องถ้า `MultiCurrency==0` | เดียวกัน |
| 3 | การจัดสรร: | IDC_ED_ALLOCATE → `AllocateCode` | combo | BCAllocate | – | เดียวกัน |
| 4 | อัตราแลกเปลี่ยน: | IDC_ED_EXCHANGE_RATE → `ExchangeRate` | number 9 ตำแหน่ง (must-entry) | 1.0 | ปิดช่องถ้า `MultiCurrency==0` | Page2:91-162 ; InfoEx:202-356 |
| 5 | ส่งมอบภายใน: … Days | IDC_ED_DELYCREDIT → `DeliveryDay` | number วัน | – | วันส่ง = วันที่เอกสาร + จำนวนวัน (คำนวณได้ทั้งสองทาง) | Page2:204-233 |
| 6 | อัตราภาษีมูลค่าเพิ่ม: | IDC_ED_TAXRATE → `TaxRate` | number | อัตราของลูกค้า ถ้าไม่มีใช้ของบริษัท | – | Page1:495-545 |
| 7 | เครดิต(วัน): … Days | IDC_ED_CREDITDAY → `CreditDay` | number | BCAR.billcredit | ปิดช่องเมื่อขายสด (`enbleCreditDayStatus`); วันครบกำหนด ↔ เครดิตคำนวณได้ทั้งสองทาง | Page2:174-201, :682 |
| 8 | ยอดเงินบาทสุทธิ: | IDC_ED_HOMEAMOUNT → `HomeAmount` | number (ปิดช่อง) | สุทธิ × อัตราแลกเปลี่ยน | – | InfoEx:403-461 |
| 9 | ค่าขนส่ง: | IDC_ED_TRANSPORT2 → `TotalTransport` | number | Σ ค่าขนส่งรายแถว (shadow 31 × จำนวน) | ปิดช่องถ้าไม่มีสิทธิ์ `bCanChangeBillAmount`; ถ้ากรอกจะทับยอดจากแถว | Page1:652-775 |
| 10 | วันที่ส่งของ: | IDC_ED_DELIVERYDATE → `DeliveryDate` | date | วันที่เอกสาร + ส่งมอบภายใน | – | InfoEx:588 |
| 11 | ทะเบียนรถ: (บน) | IDC_ED_CARLICENSE ← `CarLicense` | text | – | **โหลดจาก DB แต่ไม่พบใน OnSaveData — ยังไม่ verify ว่าบันทึกที่ไหน** | Page2:422, :495 |
| 12 | Job ID : | IDC_ED_JOBID → `JobID` | text | – | – | Page2:462 |
| 13 | วันที่ถึงกำหนด: | IDC_ED_DUEDATE → `DueDate` | date | ขายสด = วันที่เอกสาร; ขายเชื่อ = วันที่เอกสาร + เครดิต (ถ้าว่างใช้วันที่เอกสาร) | – | InfoEx:588 ; ChmFrmBilling.cpp:3275-3992 |
| 14 | รูปแบบการสั่งซื้อสินค้า: | IDC_ED_ORDERFORMSCODE → `OrderChannel` | combo | BCOrderchannel | – | Page2:466 |
| 15 | SCG Coupon NO | IDC_ED_GIFTVN → `Couponno` | text | – | **variant** | Page2:472 |
| 16 | ช่องทางการจัดจำหน่าย: | IDC_ED_DISTCHANNELCODE → `DistChannel` | combo | BCDistchannel | – | Page2:465 |
| 17 | ชื่อในการออกบิล: | IDC_ED_ARNAME → `ArName` | text | ถ้าว่าง: name2 ถ้า `DefaultArName==1` ไม่เช่นนั้นใช้ name1 | เติมตอนบันทึก | ChmFrmBilling.cpp:1405-1450 |
| 18 | ที่อยู่ในการออกบิล: | IDC_ED_ARADDRESS → `ArAddress` | text หลายบรรทัด | BCAR.BillAddress | เติมตอนบันทึก | เดียวกัน |
| 19 | หมายเลขประจำตัวผู้เสียภาษี: | IDC_ED_ARTAXNO → `ARTaxNo` | text | BCAR.TaxNo | ใช้ในรายการภาษีขาย | InfoEx:145-200 |
| 20 | หมายเหตุอื่นๆ: | IDC_ED_BILLDESCRIPTION → `MyDescription` | text | จากใบสั่งขายที่อ้างอิง | – | Page1:2701-2830 |
| 21 | เขตการขาย: | IDC_ED_SALEAREA → `SaleAreaCode` | combo | BCAR.SaleAreaCode ; BCSaleArea | – | Page2:91-162 |
| 22 | รหัสขนส่ง: | IDC_ED_TRANSPORT → `TransportCode` | combo | BCAR.DefTransport ; BCTranSport | – | เดียวกัน |
| 23 | สถานที่ขนส่ง: | IDC_ED_DELIVERY → `DeliveryAddr` | combo | BCAR.DefDeliveryAddr ; BCDeliveryAddr | เลือกแล้วเติมที่อยู่ส่ง (ข้อ 28–33) | Page2:717-751 ; InfoEx:663 |
| 24 | เวลาที่รับ/ส่งสินค้า: | IDC_ED_TIMETRANSPORT → `TimeTransport` | text | – | – | – |
| 25 | วันที่วางบิล: | IDC_ED_PAYBILLDATE → `PayBillDate` | date | – | – | – |
| 26 | Pass Bill To: | IDC_ED_PASSBILL → `PassBillTo` | lookup ลูกค้า | – | – | – |
| 27 | อ้างใบสั่งซื้อเลขที่: | IDC_ED_REFPONO → `RefDocNo` | text | refdocno ของใบสั่งขาย | – | Page1:2701-2830 |
| 27b | กลุ่มเอกสาร: | IDC_ED_BILLGROUP → `BillGroup` | text | – | ส่งต่อไปที่ OnSaveOutputTax | ChmFrmBilling.cpp:3826 |
| 28 | Ship to : | IDC_ED_BILLDESCRIPTION2 → `Shipto` | text | ชื่อลูกค้า | – | Page2:503 ; InfoEx:145-200 |
| 29 | ที่อยู่(เลขที่ , ตึก , ชื่อถนน): | IDC_ED_REFPONO2 → `Transportaddress` | text | จากสถานที่ขนส่ง | – | Page2:717-751 |
| 30 | ตำบล/แขวง : / อำเภอ/เขต : | IDC_ED_REFPONO3 / 5 → `SubDistrict` / `District` | text | เดียวกัน | – | เดียวกัน |
| 31 | จังหวัด: / รหัสไปรษณีย์: | IDC_ED_REFPONO4 / 6 → `Province` / `PostCode` | text | เดียวกัน | – | เดียวกัน |
| 32 | ทะเบียนรถ: (ล่าง) | IDC_ED_REFPONO7 (m_edTruckLicense) → `TruckLicense` | combo | BCCarMaster | – | Page2:66, :135 ; ChmFrmBilling.cpp:3409 |
| 33 | Ship to Tel : | IDC_ED_REFPONO8 → `ShiptoTel` | text | BCAR.Telephone | – | Page2:68, :504 |

- ทั้งสองแท็บกด ESC ไม่ได้ และถ้าพิมพ์อักขระใด ๆ จะตั้งสถานะว่ามีการแก้ไข (`m_bChangeData`) (Page1:5803-5822 ; Page2:696)
- คอลัมน์ที่บันทึกแต่ไม่มีช่องบนจอ:
  - `CustGroupCode`, `CustTypeCode` (จากลูกค้า)
  - `HoldingStatus`, `RecurName`, `CrAuthorizeMan` (ผู้อนุมัติเกินวงเงิน)
  - `GLTransData` (0=Auto / 1=Manual), `GLFormat`, `IsReceiveMoney`
  - `CashierCode`, `RedeemNo`/`RedeemAmount`, `SORefNo` (สูงสุด 100 ตัวอักษร)

#### ตารางรายการ (grid) — คอลัมน์ตามลำดับ (ChmFrmBillingPage1.cpp:138-166)
ตารางเป็นแบบ ILDS_LEFTBAR มีเลขแถวอัตโนมัติ ลบแถวได้ (`SetDeleteSelectRow(0)`, :169) และยืนยันก่อนลบด้วย `Confirm_DeleteLine`

| # | ป้าย | key | กว้าง | แก้ได้? | ชนิด | ความหมาย/สูตร | หลักฐาน |
|---|---|---|---|---|---|---|---|
| 0 | รหัสสินค้า | ITEMCODE | 60 | ใช่ (EDIT\|LINK, ยาวสุด 25) | lookup BCItem (active) | รับ "จำนวน*บาร์โค้ด" และ "จำนวน_บาร์โค้ด"; "." = บรรทัดข้อความ (ล้างคอลัมน์ 2–9); รหัสซ้ำคู่หน่วยเดียวกันเตือน `ITEM_DUPLICATE`; สินค้าชุด (stocktype 7) แตกจาก BCITEMComponents; serial (stocktype 2) ไม่แทรกแถว; แปลงเป็นตัวพิมพ์ใหญ่ | :138 ; :1242-1752, :1754-1887, :1903-2037 |
| 1 | ชื่อสินค้า | ITEMNAME | 120 | เฉพาะเมื่อ `IsEditItemName==1` | text | ชื่อจาก bcitem | :139 |
| 2 | คลังสินค้า | WAREHOUSE_ITEM | 60 | ใช่ | COMBO_TABLE | ค่าเริ่ม DefSaleWHCode (หรือค่าตั้ง Ctrl+F7) ; รายการจาก BCItemWarehouse ตามเงื่อนไขคลังที่ขายได้ | :140 ; :2368-2657 |
| 3 | ที่เก็บ | SHELF | 60 | ใช่ | COMBO_TABLE | DefSaleShelf ; BCItemWarehouse+BCShelf | :141 |
| 4 | หน่วยนับ | UINTCODE | 60 | ใช่ | COMBO_TABLE | DefSaleUnitCode ; BCItemUnit (หรือ BCStkPacking ถ้าหลายขนาดบรรจุ); ตรวจ `NOT_SET_UNITCODE`/`NOUNIT` | :142 ; :1242-1752 |
| 5 | จำนวน (อนุมาน) | ITEMQTY | 60 | ใช่ | number (QtyPoint) | ค่าเริ่ม `BeginQty` ของบริษัท หรือจำนวนจากบาร์โค้ด; ตรวจเกินใบสั่งขาย / สต๊อก (ดู Guard) | :143 |
| 6 | ราคา/หน่วย (อนุมาน) | PERPRICE | 60 | เฉพาะเมื่อมีสิทธิ์ `bCanChangePrice` | number (PricePoint) | `GetItemPrice(item,unit,date,qty,ar,disc,taxtype,billtype,transport)`; ถ้า `PriceWarning!=0` และต่ำกว่าทุน เตือน `SET_SALE_LOW_COST` | :146-148 ; :2368-2657 |
| 7 | ส่วนลด%,บาท | DISCOUNT | 60 | อ่านอย่างเดียวถ้า `DetailDiscount==0`, ไม่เช่นนั้นแก้ได้เมื่อมีสิทธิ์ `bCanChangeItemDiscount` | text สูตรส่วนลด | จากตารางราคา ถ้าไม่มีใช้ BCAR.DetailDiscount | :151-157 |
| 8 | จำนวนเงิน | AMOUNT | 70 | เฉพาะเมื่อมีสิทธิ์ `bCanChangePrice` | number (AmountPoint) | = จำนวน×ราคา − ส่วนลด; ถ้ากรอกเงินแต่ราคาว่าง ระบบคิดราคา = เงิน/จำนวน | :161-163 ; :547-650 |
| 9 | เงื่อนไขขนส่ง (ไทยจาก comment) | CONDIT_TRANS | 40 | ใช่ | COMBO_INDEX รับเอง/ส่งให้ | เงื่อนไขส่งรายแถว → `IsConditionSend` ของแถว | :165 |
| 10 | น้ำหนักสินค้า(Kg.) | ITEMWEIGHT+"(Kg.)" | 40 | ไม่ | number | น้ำหนักต่อหน่วย × จำนวน | :166 ; ChmFrmBilling.cpp:5776 |

**คอลัมน์ซ่อน (shadow 0–56, ChmFrmBillingPage1.cpp:173-232)**

| index | ชื่อ | ความหมาย / ใช้บันทึก |
|---|---|---|
| 0 | RowOrder | ลำดับแถว |
| 1 | ItemType | StockType ของสินค้า (1=บริการ, 2=serial, 7=ชุด) |
| 2 | MyDescription | คำอธิบายรายแถว (Shift+F2) |
| 3–4 | Project, Allocate | โครงการ / การจัดสรร รายแถว |
| 5–6 | OppositeQty, OppositeUnit | หน่วยคู่ |
| 7 | ExceptTax | 1 = สินค้ายกเว้นภาษี |
| 8 | BarCode | – |
| 9–13 | Category/Group/Brand/Type/FormatCode | หมวด/กลุ่ม/ยี่ห้อ/ประเภท/รูปแบบ (บันทึกลง sub) |
| 14 | OldUnitCode | – |
| 15 | NetAmount | ยอดสุทธิของแถว (ภาษีรวมใน = เงิน − เงิน×อัตรา/(100+อัตรา)) |
| 16 | HomeAmount | ยอดเป็นเงินบาท |
| 17 | SORefNo | เลขที่ใบสั่งขายที่อ้างอิง |
| 18 | OldQty | ยอดค้างของใบสั่งขาย ณ ตอนดึง |
| 19 | SetCode | – |
| 20–23 | DepartCode, CurrencyCode, ExchangeRate, SaleCode | – |
| 24 | LotScript | ผลจาก `CLIBDlgFifoOut` |
| 25 | OppositePrice2 | – |
| 26 | DiscountAmount | ส่วนลดเป็นเงินของแถว |
| 27 | LineNumber | – |
| 28–29 | StockProcess, QtyBalance | – |
| 30 | SOLineNumber | บรรทัดของใบสั่งขาย |
| 31 | TransportAmount | ค่าขนส่งต่อหน่วย |
| 32–33 | ContainerWeight, ContainerCapacity | – |
| 34 | OldItemCode | – |
| 35 | LotNumber | – |
| 36–38 | Color/Size/StyleCode | – |
| 39 | ItemSet | – |
| 40 | HoldingStatus | – |
| 41–43 | OldPrice, OldDiscountWord, OldQtyPrice | ใช้คืนค่าเดิมเมื่อราคา prompt |
| 44 | DORefNo | – |
| 45 | PORefNo | – |
| 46 | BehindIndex | – |
| 47 | RefType | – |
| 48 | ItemNameDesc | – |
| 49 | UnitType | – |
| 50 | masteritemcode | – |
| 51 | PremiumStatus | 1 = ของแถม |
| 52 | DocRunning | = `"<docno>:%03i"` ตั้งตอนบันทึก |
| 53–55 | prompt_price_flag, price_per_unit, prompt_price_per_unit | – |
| 56 | OldUnitPrompt | – |

#### ยอดรวมท้ายเอกสาร / สูตรคำนวณ
**1. คำนวณรายแถว — DetailCalc** (ChmFrmBillingPage1.cpp:547-650)
- **ตัวเลือก DiscountMethod 0** (:594): เงิน = จำนวน×ราคา → หักส่วนลด → ปัดตาม AmountPoint
- **ตัวเลือก DiscountMethod 1**: หักส่วนลดที่ราคาก่อน → ปัดตาม PricePoint → คูณจำนวน; ส่วนลดเป็นเงิน = ส่วนลดต่อหน่วย × จำนวน
- **ส่วนลดของแถว**: ตั้งค่า shadow 26 = ส่วนลดเป็นเงิน และคอลัมน์ 10 = น้ำหนัก
- **NetAmount ของแถว**: ถ้ามีเลขที่ใบกำกับ ไม่ใช่สินค้ายกเว้นภาษี และเป็นภาษีรวมใน จะถอดภาษีออก ไม่เช่นนั้นใช้ยอดเต็ม
- **ราคารวมภาษีต่อหน่วย**: shadow 54 `price_per_unit` ของภาษีแยกนอกคิดด้วย `*1.07` แบบฝังค่าตายตัว (:540, :632) — **ข้อควรระวัง: BC ต้องใช้อัตราจริง**
- จบด้วยการเรียก `BillCalc`

**2. ส่วนลดท้ายบิล — XDiscount** (lib-comm/BEXTools.cpp:263)
- สูตรเป็นข้อความคั่นด้วยจุลภาค เช่น `"10%,5"`
- แต่ละขั้นที่มี "%" คิดจากยอดคงเหลือหลังขั้นก่อน ขั้นที่ไม่มี "%" เป็นจำนวนเงินคงที่
- ผลรวมปัดตาม nPoint (ค่าเริ่ม 4)
- calcDiscount (ChmBilling_InfoEx.cpp:514-586) คิดส่วนลดบิลจากยอดรวมสินค้า แล้วแบ่งตามสัดส่วนให้ส่วนยกเว้นภาษี ส่วนที่เหลือเป็นส่วนลดของส่วนที่มีภาษี

**3. ยอดท้ายบิล — BillCalc** (Page1:652-775 → InfoEx `billCalc` :403-461)
- แยกยอดเป็น ยอดยกเว้นภาษี (ExceptTax==1 และบริษัทเปิด ExceptTax) กับ ยอดปกติ
- ยอดปกติบวกค่าขนส่ง (Σ shadow31×จำนวน หรือค่าที่กรอกในแท็บ 2)
- ลำดับคำนวณ: calcDiscount → calcTax → calcDueDate → MyUpdateFirstTaxRow (แถวแรกของรายการภาษีขาย) → แสดงผล → ถ้าเป็นเอกสารใหม่ตรวจโปรโมชัน
- **สูตรยอดสุทธิ**:
  - แยกนอก: สุทธิ = (ปกติ−ส่วนลดปกติ) + (ยกเว้น−ส่วนลดยกเว้น) + ภาษี − (ภาษีหัก ณ ที่จ่าย + มัดจำ1)
  - รวมใน / อัตราศูนย์: สุทธิ = (ปกติ−ส่วนลดปกติ) + (ยกเว้น−ส่วนลดยกเว้น) − (ภาษีหัก ณ ที่จ่าย + มัดจำ1)
- จากนั้นปัดเศษ และถ้ามีสกุลเงิน HomeAmount = สุทธิ × อัตราแลกเปลี่ยน

**4. ภาษี — calcTax** (InfoEx:463-512) **→ BillTaxProcess** (lib-frmwrk/BCStools.cpp:155-~235)
- คิดภาษีเมื่อ (มีเลขที่ใบกำกับ หรือ BillType เป็น 2/3) และยอดปกติ > 0
- ถ้าไม่เข้าเงื่อนไข: ไม่มีภาษี, ยอดก่อนภาษี = ยอดหลังส่วนลด, และถ้าไม่มีส่วนลด ยอดรวม = ปกติ − มัดจำ1
- **BillTaxProcess**:
  - เลขที่ใบกำกับว่าง → อัตราเป็น 0
  - หลังส่วนลด = ยอดรวม − ส่วนลด
  - แยกนอก: ก่อนภาษี = หลังส่วนลด − มัดจำ ; ภาษี = ก่อนภาษี×อัตรา/100 ; รวม = ก่อนภาษี + ภาษี
  - รวมใน: หลังส่วนลด −= มัดจำ ; ภาษี = หลังส่วนลด×อัตรา/(100+อัตรา) ; ก่อนภาษี = หลังส่วนลด − ภาษี ; รวม = หลังส่วนลด
  - อัตราศูนย์: ภาษี 0
  - mode 3 (คำนวณจากภาษีที่กรอก): แยกนอก รวม = ก่อนภาษี + ภาษี ; แบบอื่น ก่อนภาษี = รวม − ภาษี
- คอลัมน์ผลลัพธ์: `BeforeTaxAmount`, `AfterDiscount`, `TaxAmount`, `TotalAmount`, `ZeroTaxAmount` (= สุทธิ เมื่อ TaxType 2)

**5. การปัดเศษ**
- `XRound` เรียก `RoundNumber` ซึ่ง **ไม่อยู่ใน repo → โหมดปัดเศษยังไม่ verify**
- จำนวนตำแหน่งทศนิยมแยกตาม `QtyPoint` / `PricePoint` / `AmountPoint` ของบริษัท
- **ตรวจตอนบันทึก**: ยอดเงินของแถวต้องต่างจากผลคำนวณใหม่ไม่เกิน `m_Option.dDiffAmount` (ค่าเริ่ม 1.0) (ChmFrmBilling.cpp:1784)
- **ตรวจตอนบันทึก**: ยอดหัวบิลต้องเท่ากับผลรวมแถว (:2508)

**6. หลายอัตราภาษี / รายการภาษีขาย (F3)** — `OnHkTax` (ChmFrmBilling.cpp:6776)
- เปิด `OnOutputTax("BillingTransConfig",taxtype,ar,artaxno)` = dialog `CLIBDlgMultiTax` "กำหนดรายการภาษีขาย" (ช่อง: รหัสเครื่อง, "Press F9=Calc Include Tax") (atlas-02-Bill.md:1151)
- กลับมาแล้วอ่าน TaxNo/TaxDate จากแถวแรกของรายการภาษี แล้วคำนวณภาษีใหม่
- แถวแรกของรายการภาษีสร้างอัตโนมัติ (`MyUpdateFirstTaxRow`/`UpdateTaxRow`, ChmFrmBilling.cpp:6070-6193):
  - คำอธิบาย `SALE_ITEM` (ขายสินค้า — อนุมาน) + "#ชื่อในการออกบิล" ถ้าต่างจากชื่อลูกค้า
  - ใส่ IsBranch/BranchVatNo จาก BCAR
  - ภาษีอัตราศูนย์ลง ZeroTaxAmount
  - ไม่ปรับถ้าไม่มีเลขที่หรือวันที่ใบกำกับ
- **F9 ในจอขาย** ไม่ใช่ dialog ภาษี: accelerator F9 = cancel hold (BC5Account.rc:975-1009); ปุ่ม F9 อยู่ภายใน `CLIBDlgMultiTax`

#### การรับชำระ (ขายสด) / มัดจำ / ภาษีหัก ณ ที่จ่าย
| ปุ่ม | Dialog | ช่อง/ผล | ใช้กับ | หลักฐาน |
|---|---|---|---|---|
| F7 "F7=บันทึกรับเงิน" | `CLIBMoneyRec` (atlas:1138); แท็บ เงินสด/เงินโอน, บัตรเครดิต, เช็ค, คูปอง, อื่นๆ (LIBMoneyRec.cpp:285-345) | **เงินสด** (rc:3126): ยอดเงินสด, เงินโอนเข้าบัญชี (ยอดเงินโอน, เลขที่สมุดบัญชี), เงินเกิน (ยอดเงินเกินในใบเสร็จ, ยอดเงินเกินนำมาหัก), รายได้/ค่าใช้จ่ายอื่นๆ, Cashier, วงเงินสดย่อย, หมายเหตุในรายงานรับเงินประจำวัน · **บัตรเครดิต** (rc:3043): เลขที่บัตรเครดิต, ยอดบัตรเครดิต, เลขที่อนุมัติ, ประเภทบัตร, รหัสธนาคาร, % Charge, สาขาธนาคาร, ยอด Charge, วันที่ถึงกำหนด, รวมเป็นเงิน (F9 Update/F6 Delete/F5 Clear) · **เช็ค** (rc:3082): เช็คฉบับใหม่/ตัดยอดจากฉบับเดิม, รหัสธนาคาร, สาขาธนาคาร, ตัดจ่ายหลายใบเสร็จ, มูลค่าของเช็ค, เลขที่เช็ค, ยอดรับชำระ, วันที่ถึงกำหนด, หมายเหตุ · ปิดแล้วแสดง `CLIBMoneyCashChange` (ยอดรับ/เงินทอน) | ขายสด (BillType 0/2) เท่านั้น; มีเตือนมัดจำก่อน | `OnRec` ChmFrmBilling.cpp:6525-6585 ; `OnHkMoney` :6753 |
| F8 "F8=หักเงินมัดจำรับ" | `CLIBDlgDeposit` "บันทึกตัดใบรับเงินมัดจำ" (checkbox เฉพาะรายการที่มีภาษี) (atlas:1164) | ได้ `SumOfDeposit1/2`, `DepositIncTax` แล้วคำนวณใหม่ | ทุกแบบ ยกเว้นตั้งค่า `IsDepositCash==1` ขณะขายเชื่อ → `EARNEST_CASH` (ห้าม) | `OnHkDeposit` :6651 ; myespeciallyformview.cpp:4761 |
| F6 "F6=ภาษีหัก ณ ที่จ่าย" | `CLIBDlgARWTax` "กำหนดภาษีหัก ณ ที่จ่าย" (Auto Wtax) (atlas:1158) | ฐาน = BeforeTaxAmount → `SumOfWTax` แล้วหักจากยอดสุทธิ | ทุกแบบ | `OnHkWtax` :6806 |
| F3 "F3=บันทึกภาษีขาย" | `CLIBDlgMultiTax` | ดูสูตรข้อ 6 | ทุกแบบ | :6776 |

- **คอลัมน์ยอดรับที่บันทึก** (เมื่อไม่ใช่เอกสารตั้งซ้ำ): `SumCashAmount`, `SumChqAmount`, `SumCreditAmount`, `ChargeAmount`, `SumBankAmount`, `SumOfWTax`, `SumOfDeposit1`, `SumOfDeposit2`, `ExchangeProfit`, `CoupongAmount` (ChmFrmBilling.cpp:3275-3992)
- **ขายสดต้องรับเงินครบ** (`IsMoneyOk(0,net)`) เมื่อ (`IsReceiveCash==1` และ `IsReceiveMoney==1`) หรือ `IsReceiveMoney==0`
  - ถ้าไม่ครบ แจ้ง `AMOUNT_NOT_EQUAL_RECEIVE` แล้วเปิดจอรับเงิน (:2524-2526)

#### อ้างอิงเอกสาร / ดึงจากเอกสารอื่น
| แหล่ง | ทางเข้า | เงื่อนไขค้นหา | ค่าที่คัดลอก | ผลข้างเคียง | หลักฐาน |
|---|---|---|---|---|---|
| ใบสั่งขาย/สั่งจอง | Ctrl+F3 หรือปุ่ม หรือคำถามตอนเลือกลูกค้า | holdingstatus=0, billstatus=0, iscancel=0, ลูกค้าเดียวกัน, taxtype เดียวกัน, billtype เดียวกันหรือ null, ถ้าระบุ salecode ต้องตรงกัน, มีบรรทัด RemainQty>0, ไม่รวมใบที่ดึงแล้ว (ถ้าไม่ได้ส่ง salecode จะเพิ่ม iscompletesave=1); เลือกได้หลายใบ (`CMyMultiSelect`, `SaleOrderConfig`) | **หัว**: สกุลเงิน, อัตรา, refdocno, หมายเหตุ, สถานที่ขนส่ง (+รหัสขนส่ง), เวลา · **บรรทัด** (`transstate=0, iscancel=0, remainqty>0 หรือ '.'` เรียงตาม linenumber): สินค้า, คลัง/ที่เก็บ, หน่วย, **จำนวน = RemainQty**, ราคา, ส่วนลด, เงิน, เงื่อนไขส่ง, shadow 17 = เลข SO, 30 = บรรทัด SO | ล็อกช่องลูกค้า; ถ้าตั้ง `BILLNoEditPrice` ห้ามแก้ราคา/เงิน, `BILLNoEditDiscount` ห้ามแก้ส่วนลด, `BILLNoAddnew` ห้ามเพิ่มแถว; เมื่อบันทึกจะตั้ง `bcsaleordersub.isprocess=0` ให้คำนวณยอดค้างใหม่ (ตัวคำนวณ RemainQty ยังไม่ verify) + document flow 4 | Page1:2701-3137 ; ChmFrmBilling.cpp:3460, :3797, :3864 |
| ไม่มีลูกค้าแต่มีพนักงานขาย | Ctrl+F3 | ค้นตามพนักงานขาย | – | – | Page1:2711-2716, :3138 (ยังไม่ verify รายละเอียด) |
| ไม่มีทั้งคู่ | Ctrl+F3 | – | แจ้ง `SALEORDER_ARCODE` ("โปรดระบุรหัสลูกค้า" — comment) | – | Page1:2720 |
| ใบเสนอราคา | ปุ่ม "เลือกใบเสนอราคา" / คำถาม `WARN_AR_QUOTATION` | (ยังไม่ verify) | (ยังไม่ verify) | document flow 22 | Page1:3749, :4150 ; ChmFrmBilling.cpp:3868 |
| ส่งตรงจากใบซื้อ | – | – | โหลดใบซื้อ | – | `OnLoadPurchase` ChmFrmBilling.cpp:7452 |

- **Serial (Ctrl+F9)**: ใช้ได้เมื่อ `SerialItem!=0` และ stocktype==2
  - เรียก `OnSerialNumber(4, qty, ...)`; ถ้าไม่เข้าเงื่อนไขแจ้ง `ITEM_NOT_HAVE_SERAIL`
  - ตอนบันทึกตรวจ `NOT_SPECIFY_SERIAL` / `SPECIFY_NOTEQUAL_SERIAL`
  - หลักฐาน: :646-678, :6872, :1939-1950
- **Lot (Shift+F9)**: ใช้ได้เมื่อ `nLotStatus!=0` และ costtype==1 (FIFO)
  - เปิด `CLIBDlgFifoOut` แล้วเก็บผลลง shadow 24
  - ตอนบันทึกตรวจ `NOT_SET_LOT` / `STOCK_NOT_ENOUGHLIST`
  - หลักฐาน: :602-645, :6855, :2000-2010
- **ตัวเลือกสต๊อก**:

| ตัวเลือก | ค่า / ผล |
|---|---|
| `StockStatus` | 1 = เตือน `STOCK_NOT_MOVEMENT` ; 2 = ห้ามติดลบ (เตือนรายแถวแบบถาม แล้วรวมยอดตรวจตอนบันทึก) |
| `StockRemain` | 2 = `IsSalesStockEnough` |
| `IsItemRemainQty` | รวมยอดตรวจถึงระดับที่เก็บ |
| `IsUseWarehouse` | 1 = ไม่บังคับคลัง/ที่เก็บ |
| `nEnableAssignSaleByWH` | จำกัดคลังที่ผู้ใช้ขายได้ (`OnCheckSellWarehouse`; ผู้ใช้ sa ผ่านเสมอ, mytools.cpp:3206-) |
| ตัดสต๊อก | บันทึกเสร็จแล้วสั่ง `OnStockReprocess` รายสินค้า (:3806) |

#### ปุ่ม แถบเครื่องมือ คีย์ลัด
- **แถบเครื่องมือมาตรฐาน**: New, Save, Del, Find, Prev, Next, Close, Info, Key(IDC_MORE), Recur
  - Recur ใช้ได้เพราะจอนี้เปิดให้ตั้งเอกสารซ้ำ (ChmFrmBilling.cpp:679)
- **ปุ่ม "Key"** (`OnMore` :431-470) แสดง popup รายการคีย์ ไทยจาก keyThai: F2=บัญชีแยกประเภท · F3=บันทึกภาษีขาย · F5=รายละเอียดอื่นๆ · F6=ภาษีหัก ณ ที่จ่าย · F7=บันทึกรับเงิน · F8=หักเงินมัดจำรับ · F11 บันทึก · F12=ตรวจสอบเอกสาร · Shift+F2=กำหนดรายเอียดสินค้าเพิ่มเติม · Shift+F5=Confirm เอกสาร · Home ตรวจสต๊อก · Ctrl+F3=เลือกสั่งขายสินค้า · Ctrl+F5=ยกเลิกเอกสาร · Ctrl+F7/Ctrl+F8 (คลังเริ่มต้น) · Ctrl+F9 = Serial Number · ประวัติการซื้อ
- `PreTranslateMessage` ของ form ถูก comment ทั้งก้อน (:684-1032) คีย์ทั้งหมดจึงมาจากตาราง accelerator `IDR_MAINFRAME_CHM` (BC5Account.rc:975-1009)
- **คีย์ที่ form ผูก handler จริง** (message map :113-152):

| คีย์ | ผล | handler |
|---|---|---|
| F2 | บัญชีแยกประเภท (`OnAccount`) — ใช้ตอนลงบัญชีแบบ Manual | :6589 |
| F3 | รายการภาษีขาย (`CLIBDlgMultiTax`) | :6776 |
| F5 | สลับแท็บ ข้อมูลหลัก ↔ รายการอื่นๆ | :6689 |
| F6 | ภาษีหัก ณ ที่จ่าย | :6806 |
| F7 | รับเงิน (เฉพาะขายสด) | :6753 |
| F8 | ตัดมัดจำ | :6651 |
| F11 | บันทึก (`OnRecSave`) | message map |
| F12 | Document flow | :6714 |
| Shift+F2 | รายละเอียดแถว `CBILLDlgBillExpand` (เอกสารอ้างอิง, แผนก, โครงการ, การจัดสรร, คำอธิบายรายการ, คุณสมบัติสินค้า) | :6829 ; atlas:1042 |
| Shift+F5 | Confirm เอกสาร | :6632 |
| Shift+F9 / Ctrl+F9 | Lot / Serial | :6855 / :6872 |
| Ctrl+F3 | ดึงใบสั่งขาย | :6732 |
| Ctrl+F5 | ยกเลิกเอกสาร | :6610 |
| Ctrl+F2 | ตรวจสต๊อกย้อนหลัง + GetStock | :7058 |
| Ctrl+F7 | ตั้งคลัง/ที่เก็บเริ่มต้น (`OnSetShelfWarehouse`) | :7072 |
| Ctrl+F8 | ถาม `DEFAULT_WH` แล้วใช้คลังของแถวแรกกับทุกแถว | :7081 |
| Ctrl+F12 | `CChmItemDisplayDlg` เลือกคลัง/ราคา (ตั้งราคาได้เฉพาะเมื่อมีสิทธิ์ `bCanChangePrice`) | :6979 |
| Ctrl+P | พิมพ์ฟอร์ม | :6950 |
| Ctrl+F | ค้นหา | :6961 |
| Ctrl+N | ใหม่ | :6970 |
| PageDown | ไปตาราง (0,0) | :6926 |
| PageUp | ไปช่องเลขที่เอกสาร | :6938 |

- **คีย์ใน accelerator ที่ form นี้ไม่ได้ผูก handler** (ID_HK_GETIR, ITEMBYPO, PARALUNIT, BARCODE และอื่น ๆ) จึงไม่มีผลในจอนี้
- **การไล่โฟกัส Enter/Tab** อยู่ใน grid control กลาง (BInputList) — ไม่ได้ verify ในงานนี้
- **พิมพ์ฟอร์ม**:
  - เงื่อนไข: ต้องมีสิทธิ์ `bCanPrintBill` และเอกสารบันทึกแล้ว (`OnCheckBeforePrint`, mytools.cpp:3170-3183)
  - ฟอร์มเลือกอัตโนมัติตามลำดับ = BillType (`m_nDefFormIndex`, ChmFrmBilling.cpp:3295, :4006)
  - ตัวแปรในฟอร์มมาจาก `GetFormData` (:4336-) เช่น DIFF_AMOUNT, TRANSPORTCODE/NAME, SALEAREACODE/NAME, AR_TAXNO, BILLGROUP, BILLTYPE, TAXDATE (อ่านจาก bcoutputtax), RECEIVETYPE
  - พิมพ์ serial ใช้ `BCSerialPrintForm` SaveFrom=4 (:4316)
  - นับจำนวนครั้งที่พิมพ์ด้วย `PrintCount` และบันทึก log ผ่าน `UpdatePrintLog` (:6367)
  - พิมพ์ slip GL ใช้ source 6

#### สถานะเอกสารและ workflow
**ป้ายสถานะ** `DrawBillStatus` (myespeciallyformview.cpp:5690):

| ป้าย | ความหมาย |
|---|---|
| **N** | ปกติ |
| **CF** (สีเขียวน้ำเงิน) | ยืนยันแล้ว |
| **CC** (สีแดง) | ยกเลิก |
| **0.00** (สีน้ำตาล) | ยอดคงเหลือ 0 / รับชำระครบ |

**การกระทำแต่ละขั้น**:

| ขั้น | ใครทำได้ / คีย์ | คอลัมน์ที่เปลี่ยน | ผล | หลักฐาน |
|---|---|---|---|---|
| บันทึก | สิทธิ์ ID_BILLSYS_TRANS_6; ผู้ใช้ที่ตั้ง `IsBillCash==1` ห้ามขายสด, `IsBillCredit==1` ห้ามขายเชื่อ | สร้างหัว + บรรทัด, `IsCompleteSave` 0→1 | `BillBalance` = `PayBillAmount` = `BillTemporary` = 0 (ขายสด) หรือ = สุทธิ (ขายเชื่อ) | ChmFrmBilling.cpp:1314-1318, :3275-3992 |
| Confirm | Shift+F5, สิทธิ์ `bCanConfirmBill` (ไม่มีสิทธิ์ → `NOT_CONFIRM_NO`), dialog `CMyDlgBillConfirm` พิมพ์ "YES" | `IsConfirm=1`, `ConfirmCode`, `ConfirmDateTime` | หลังยืนยันแล้ว แก้/ลบ/ยกเลิกไม่ได้ | myespeciallyformview.cpp:5646 ; MyDlgBillConfirm.cpp:51-69 |
| ยกเลิก | Ctrl+F5, สิทธิ์ `bCanCancelBill` (`NOT_CANCEL_NO`), dialog `CMyDlgBillCancel` พิมพ์ "YES" | `IsCancel=1`, `CancelCode`, `CancelDateTime`, `BillBalance=0`, sub `IsCancel=1` | กลับรายการทั้งหมด **เก็บเอกสารไว้** (ดูผลข้างเคียง) | :6610-6631, :5813-5956 ; myespeciallyformview.cpp:5590 |
| ลบ | ปุ่ม Del + ยืนยัน `DELETE_MSG` | ลบจริงจาก bcarinvoice/sub | ลบผลข้างเคียงทั้งหมด (ภาษี, รับเงิน, มัดจำ, GL `DeleteAccount`) | :3055-3273 |
| วางบิล | จากจอใบวางบิล | `PayBillStatus=1` | ห้ามแก้/ลบ/ยกเลิก | :1274 |
| รับชำระบางส่วน (ขายเชื่อ) | จากจอรับชำระ | `BillBalance` < สุทธิ | ห้ามแก้/ลบ/ยกเลิก | :1283 |
| ถูกอ้างอิง | ใบลดหนี้ (bcinvcreditnote billtype 0), ใบเพิ่มหนี้, ใบวางบิล, ใบเสร็จ 1/2/3 | – | ห้ามแก้/ลบ/ยกเลิก | :1304, :1703 ; :4100-4153 |
| ผ่าน GL | โมดูล GL | `IsPostGL` | ห้ามแก้/ลบ/ยกเลิก | :1258 |
| ปิดงวด | – | – | บันทึกหรือลบไม่ได้ (`IsPeriodClose`) | :1189 |
| บิล POS | – | `PosStatus=1` | ไม่แสดงในจอนี้ และบันทึกไม่ได้ | :1212 |

- การลงบัญชีมี 2 โหมด:
  - `GLTransData`=0 (Auto): `m_GLTLSProcess` ใช้ source 6, group 6 และรูปแบบตาม BillType 8–11
  - =1 (Manual): ผู้ใช้กด F2 แล้ว `OnSaveAccount(...,6,...)`
- ผังบัญชีที่ใช้จริงมาจาก master รูปแบบ GL — ไม่อยู่ในจอนี้ (ChmFrmBilling.cpp:3849-3858, :4240-4252)

#### ผลข้างเคียง (ลำดับใน `OnSaveData`, ChmFrmBilling.cpp:3275-3992)

**ขั้นที่ 1–6 อยู่ใน transaction**

| ลำดับ | การทำงาน | หลักฐาน |
|---|---|---|
| 1 | ถ้าแก้เอกสาร: กลับยอดลูกหนี้เดิม `BillingARProcess(old,-1)` | :3285 |
| 2 | เริ่ม transaction | :3434 |
| 3 | เขียนหัว `BCARInvoice` (คอลัมน์ทั้งหมดตามตาราง "หน้าจอ" ด้านบน + DocumentRunning, CreatorCode, CreateDateTime, PrintCount=0 ตอนสร้าง / LastEditorCode, LastEditDateT ตอนแก้) | :3275-3434 |
| 4 | ถ้าแก้เอกสาร: ตั้ง SO เดิม `isprocess=0`, ลบ sub เดิม, ถอน document flow 4/22 | :3460-3485 |
| 5 | เขียน `BCARInvoiceSub` ทีละแถว | ดูรายการคอลัมน์ด้านล่าง |
| 6 | `update bcarinvoice set iscompletesave=1, sorefno=…` → ตั้ง SO `isprocess=0` → **Commit** | :3794-3800 |

**คอลัมน์ `BCARInvoiceSub` (ขั้นที่ 5)**
- **รหัส/อ้างอิง**: DocNo, DocDate, ArCode, TaxNo, TaxType, SaleAreaCode, CustGroupCode, ItemCode, ItemName, WHCode, ShelfCode, UnitCode, BarCode, RefLineNumber, RefType, SORefNo, PORefNo, LineNumber, DocumentRunning
- **จำนวน/เงิน**: Qty, Price, DiscountWord, Amount, DiscountAmount, NetAmount, HomeAmount, BalanceAmount, CNQty, TransportAmount, BTDisc (= NetAmount ของแถว × ส่วนลดบิล ÷ ยอดรวมสินค้า), price_per_unit, prompt_price_*
- **มิติ/จัดกลุ่ม**: ProjectCode / AllocateCode / DepartCode (ถ้าแถวว่างใช้ค่าหัว), Category/Group/Brand/Type/FormatCode, Color/Size/Style, ItemSetCode, LotNumber, CurrencyCode, ExChangeRate, UserGroup, SaleCode
- **อื่น ๆ**: IsConditionSend, Netweight, StockType, MyDescription, OppositeQty/Unit/Price2, ExceptTax, ItemNameDesc, behindindex, IsProcess, PackingRate1/2
- **ของแถม**: แถวหัว "." ใช้ชื่อ `ADDITION` ตามด้วยแถวของแถมที่ Price=0, PremiumStatus=1, masteritemcode

**ขั้นที่ 7–17 อยู่หลัง commit (ไม่อยู่ใน transaction เดียวกัน)**

| ลำดับ | การทำงาน | หลักฐาน |
|---|---|---|
| 7 | ตัดสต๊อก `OnStockReprocess` รายสินค้า + `insert into processdoc (docno,doctype) values(...,4)` | :3806, :3812 |
| 8 | serial `OnSaveSerial(4)`, lot `OnSaveLotItem(4)` | :3818-3820 |
| 9 | ภาษีขาย `OnSaveOutputTax(1,…,1,6,billgroup,…)` → **BCOutputTax** (ถ้าไม่มีเลขที่ใบกำกับจะล้างแถวภาษี) | :3823-3826 ; myespeciallyformview.cpp:1005 |
| 10 | ภาษีหัก ณ ที่จ่ายลูกหนี้ `OnSaveARWTax` | :3828 ; myespeciallyformview.cpp:1149 |
| 11 | log คูปอง/วอลเล็ต (**variant**) | :3831-3832 |
| 12 | รับเงิน: ขายสด → `OnSaveReceive(1,…)` (รายละเอียดด้านล่าง); แก้จากสดเป็นเชื่อ → `DeleteReceive(1)` | :3836-3840 ; myespeciallyformview.cpp:1985 |
| 13 | ตัดมัดจำ `OnSaveARDeposit` | :3843 ; myespeciallyformview.cpp:4907 |
| 14 | GL (Auto/Manual ตามหัวข้อก่อนหน้า) | :3847-3858 |
| 15 | ลูกหนี้ `BillingARProcess(docno,1)` (รายละเอียดด้านล่าง) | :3861 ; bcprocess.cpp:1863-2000+ |
| 16 | document flow (1,4,SO) และ (1,22,ใบเสนอราคา) | :3864-3868 |
| 17 | integration ภายนอก (**variant**) → `AddUserLogBook` | :3988 |

- **รายละเอียดการรับเงิน (ขั้นที่ 12)**: เขียน **BCRecMoney** (SaveFrom 1 = ขาย)
  - PayType: 0=เงินสด 1=บัตร 2=เช็ค 3=เงินโอน 4=เงินสดย่อย 5=เงินเกินพัก 6=เงินเกินหัก 7=รายได้อื่นๆ 8=ค่าใช้จ่ายอื่นๆ (myespeciallyformview.cpp:2014)
  - เขียนเพิ่ม: BCChqIn, BCCreditCard, ปรับ BCBankBook, BCPettyCash
- **รายละเอียดลูกหนี้ (ขั้นที่ 15)**:
  - ขายสด: ยอดหนี้ = 0 แต่นับยอดเช็ค/บัตรค้าง
  - ขายเชื่อ: ยอดหนี้ = NetDebtAmount
  - `ProcessAR` (bcprocess.cpp:2974-3056) ปรับ BCARDebtTable (ต่อลูกค้า+สกุลเงิน+งวด) และ BCAR.DebtAmount/ChqOnHand/CreditOnHand/ChqReturn
  - ลดวงเงินคงเหลือ `BCAR.DebtLimitBal -= หนี้` (เมื่อ DebtLimit1>0)
  - `UpdateARPeriodState` (:3171) ปรับ bcarperiodstate
  - ปรับ `bcar.lastcontactdate` และขายสดปรับ `lastpaydate` ด้วย
- **ถ้าบันทึกรายละเอียดล้มเหลว**: Abort แล้วแจ้ง `SAVEDETAIL_NOTCOMPLETE` (:3981)
- **ยกเลิก** (`BillCancelUpdateData` :5813-5956):
  - SO isprocess=0
  - `BillingARProcess(-2)`
  - หัว: BillBalance=0, IsCancel=1 ; sub: IsCancel=1 ; คูปอง IsCancel=1
  - คำนวณสต๊อกใหม่
  - `CancelOutputTax(1,doc,6)` (ยกเลิกแถวภาษี ไม่ใช่ลบ)
  - ลบ WTax / รับเงิน / มัดจำ
  - **`ReverseAccount(doc,6)`** (กลับรายการ GL)
  - ลบ serial/lot, ถอน doc flow, log `CANCEL_BILL`
- **ลบ** (`OnDeleteData` :3130-3273): เหมือนยกเลิก แต่ลบหัว/sub จริง, ใช้ `DeleteOutputTax`, ใช้ **`DeleteAccount(doc,6)`** และ log `SUBTRACT_BILL`

#### Guard / ข้อความเตือนสำคัญ (`ConfirmSaveData`, ChmFrmBilling.cpp:1159-2640 ตามลำดับ)
| บรรทัด | key ข้อความ | เงื่อนไข | ผล |
|---|---|---|---|
| 1189 | (ปิดงวด) | `IsPeriodClose(docdate)` | หยุด |
| 1212 | DOC+POS_SYSTEM+CANNOT_SAVE | แก้บิล POS | หยุด |
| 1220 / 1229 | CANNOT_SAVE_NOTDOCNO / DOCNO_DUPLICATE | เลขที่ว่าง / เปลี่ยนเลขแล้วซ้ำ | หยุด |
| 1236 | CANNOT_SAVE_NOTDOCDATE | วันที่ว่าง | หยุด |
| 1244 | CANNOT_SAVE_EMPTY_TAXDOCNO | `WarnTaxNoEmpty` (bcconfigurations) และเลขที่ใบกำกับว่างและยอด≠0 | หยุด |
| 1258–1304 | CANNOT_SAVE_POSTGL · NOT_SAVE_ISCANCEL · NOT_SAVE_ISCONFIRM · CANNOT_EDIT_BILL1 · NOT_EDIT_BILLBALANCE · CANNOT_SAVE_BILLSTATUS_REF | แก้เอกสารที่ ผ่าน GL / ยกเลิก / ยืนยัน / วางบิล / รับชำระบางส่วน / ถูกอ้างอิง | หยุด |
| 1314–1357 | CODE_USER + BILL_CASH / BILL_CREDIT | ผู้ใช้ถูกห้ามขายสด/เชื่อ | หยุด |
| 1330 | AR_DEPOSIT_OFFSET + F8EARNEST_RECEIVE | ขายเชื่อ ไม่ได้ตัดมัดจำ, `IsWarnDeposit==1` และลูกค้ามีมัดจำ | ถาม; Yes → เปิดตัดมัดจำและหยุดบันทึก |
| 1387 / 1393 | CANNOT_SAVE_NOTITEM / CANNOT_SAVE_NOTARCODE | ไม่มีแถว / ไม่มีลูกค้า | หยุด |
| 1438 / 1446 / 1453 | CANNOT_SAVE_NOTARCODE_INDB / AR_DISCONTACT / CANNOT_ISSUE_BILL | ไม่พบลูกค้า / ลูกค้าเลิกติดต่อ / `IsMakeTrans==1` ("สถานะลูกค้า ไม่สามารถออกบิลขายได้" — comment) | หยุด |
| 1504–1567 | **วงเงิน**: BILL_CREDIT+CREDIT_ARCODE_OVER · CREDIT_ARCODE_OVER+BILL_BALANCE · PRIVILEGE_NOT_APPROVED | ดูตารางวงเงินด้านล่าง | หยุด / เตือน / ขออนุมัติ |
| 1594 | CHECK_TAXDATE | มีเลขที่ใบกำกับแต่ไม่มีวันที่ | หยุด |
| 1604 | CANNOT_SAVE_NOTFOUND_CHAMPMACHINECODE | เปิด `nUseMachineCode` แต่ไม่พบรหัสเครื่อง | หยุด |
| 1628 | DUPLICATE_TAXNO | เปลี่ยนเลขที่ใบกำกับแล้วซ้ำใน bcoutputtax | หยุด |
| 1703 | CANNOT_EDIT_REFERENCE | ถูกอ้างโดยใบลดหนี้/เพิ่มหนี้/วางบิล/ใบเสร็จ (transstate=0) | หยุด |
| 1738 / 1784 | PRODUCTCODE+NOT_DATA / LIST_AT n CHECK_WRONG_AMOUNT | ไม่พบสินค้า / เงิน≠จำนวน×ราคา−ส่วนลด เกินค่าคลาดเคลื่อน | หยุด |
| 1811 | CANNOT_SALE_LOWER_PRICE | `PriceWarning==2` และราคาต่ำกว่าทุน (ไม่รวมของแถม) | หยุด |
| 1885–1964 | CANNOT_SUBTRACT · NO_SELL_FROM_WH · NOT_SET_WHCODE · NOT_SET_SHELFCODE · NOT_SET_UNITCODE · QTY_ZERO_LIST · NOT_SPECIFY_SERIAL · SPECIFY_NOTEQUAL_SERIAL · CHECK_AMOUNT_ZERO | ตรวจรายแถว (ไม่รวมบริการ=1 และชุด=7) | หยุด |
| 2000–2131 | NOT_SET_LOT · STOCK_NOT_ENOUGHLIST · SUMAMOUNT…BALANCE | lot / `StockStatus==2` ยอดรวมต่อสินค้า+คลัง(+ที่เก็บ) เกินคงเหลือ (รวมของแถม) / `StockRemain==2` | หยุด |
| 2462–2481 | CANNOT_SAVE_NO_CHECKSCGID / _NO_REDEEMCOUPON / _NO_WALLET | **variant** | หยุด |
| 2489 | WARNING_TEXT_SALELOWER_SALEPRICE | `WarningLowerStdPrice` (bcconfigurations) และราคาต่ำกว่าราคามาตรฐาน | เตือน (Page1:6151-6212) |
| 2508 | CANNOT_SAVE_SUMGOOD_NOTEQUAL_SUMDETAILIT | ยอดหัว ≠ ผลรวมแถว | หยุด |
| 2526 | AMOUNT_NOT_EQUAL_RECEIVE | ขายสดรับเงินไม่ครบ | หยุดและเปิดจอรับเงิน |
| 2568-2598 | "Document No is duplicated\nDo you want to try again?" / "Tax No. is duplicated…" | ตอนสร้าง เลขที่ (หรือเลขที่ใบกำกับ) ถูกใช้ไปแล้ว | Yes = `RunningAgain` รันเลขใหม่ (กรณีเลขเอกสารซ้ำ รันใหม่ทั้งเลขเอกสารและเลขใบกำกับ), No = ล้างช่อง; ทั้งสองกรณี **หยุดบันทึก** ต้องกดบันทึกอีกครั้ง |

**วงเงินลูกหนี้** (ChmFrmBilling.cpp:1504-1590 ; :7106-7190 ; myespeciallyformview.cpp:5750):
- **เงื่อนไขเข้าตรวจ** (ต้องครบทุกข้อ):
  - บริษัทตั้ง `DebtOverWarning>0` (เตือนลูกหนี้เกินวงเงิน)
  - ลูกค้ามี `DebtLimit1>0`
  - บริษัทตั้ง `BillDebtWarning>0` (เตือนที่บันทึกขาย)
  - เป็นขายเชื่อ (BillType 1/3)
  - เป็นเอกสารใหม่ หรือยอดเพิ่มขึ้น (ถ้าลูกค้าเดิม ระบบหักยอด NetDebtAmount เดิมออกก่อน)
- **ยอดที่นำมาเทียบ**: สุทธิ − แถวที่อ้างใบสั่งขายซึ่งอนุมัติวงเงินแล้ว
  - ใบสั่งขายนับว่าอนุมัติแล้วเมื่อ (isconfirm=2 และมี CrAuthorizeMan) หรือ (isconfirm=1 และมี Approvecode) (`OnCheckArCredit` :7165)
- **เทียบกับ** `DebtLimitBal` ของลูกค้า (วงเงินคงเหลือ)

| สถานะลูกค้า/ตัวเลือก | ผล |
|---|---|
| BCAR `IsMakeTrans==2` และเกินวงเงิน | ห้ามออกบิลเชื่อ (หยุด) |
| `DebtOverWarning==1` | เตือนอย่างเดียว พร้อมยอดคงเหลือ |
| `DebtOverWarning==2` | เตือน (ผู้ใช้ไม่มีสิทธิ์ `bCanAuthorizeCreditLimit` จะเห็นข้อความ PRIVILEGE_NOT_APPROVED เพิ่ม) → `checkAndPopupCreditLimit` → `ArCreditExceed` → dialog `CMyConfirmCreditLimit` "อนุมัติขายเกินวงเงิน" |
| แก้เอกสารจากขายสดเป็นเชื่อ | เรียก `checkAndPopupCreditLimit` ด้วย |
| ตอนเลือกประเภทขายเชื่อ (Page1:466, :473) | `IsMakeTrans==2` แสดงยอดคงเหลือ ; `==0` แสดง CREDIT_ARCODE_OVER+CANCEL_ARCREDIT (ความหมายค่า 0 อนุมาน) |

- **Dialog อนุมัติ** มีช่อง: รหัสลูกค้า, จ่ายเงินล่าสุด, ชื่อลูกค้า, ยอดวงเงิน, ยอดลูกหนี้, เช็คคงค้าง, ยอดเช็คคืน, วงเงินคงเหลือ, ยอดเงินที่จะอนุมัติ, คำถาม "ต้องการอนุมัติขายเกินวงเงินหรือไม่", รหัสผู้อนุมัติ, รหัสผ่าน (atlas:1185)
- **อนุมัติผ่าน**: เก็บรหัสผู้อนุมัติลง `CrAuthorizeMan`
- **ไม่อนุมัติ**: ถ้า nConfirmStatus==0 จะล้างจอ (`OnClearData`)

**ข้อความเตือนตอนคีย์** (ไม่ได้หยุดการบันทึก):

| key | เงื่อนไข / ผล | หลักฐาน |
|---|---|---|
| ITEM_DUPLICATE | สินค้า+หน่วยซ้ำในตาราง | Page1:1242-1752 |
| STOCK_NOT_ENOUGH | ถามแบบ yes/no แล้วใช้ GetStock | เดียวกัน |
| STOCK_NOT_MOVEMENT | `StockStatus==1` | เดียวกัน |
| SET_SALE_LOW_COST | ราคาต่ำกว่าทุน | เดียวกัน |
| (เกินใบสั่งขาย) | ตัวเลือก `BILLNoEditQty`/`StockRemain`; ตรวจด้วย `SaleMuchSaleOrder` | Page1:6222 |
| TAXNO_MORE_LENG | เลขที่ใบกำกับยาวเกิน 20 | – |
| WARN_AR_QUOTATION | ลูกค้ามีใบเสนอราคาค้าง | – |
| ITEM_NOT_HAVE_SERAIL | สินค้าไม่ใช่ serial | – |
| EARNEST_CASH | ห้ามตัดมัดจำกรณีขายเชื่อเมื่อตั้ง `IsDepositCash==1` | – |

---

## ส่วน 5 — [5330] ใบรับคืนสินค้า,ลดหนี้ และ [5331] ใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)

> ขอบเขต: เมนู Bill (ระบบขาย) [5330] และ [5331] (atlas-02-Bill.md:1457-1856)
> หลักฐานอ้างจาก decoded mirror `scratch/champ-atlas/dec-sa/champ/` (เลขบรรทัดเท่าต้นฉบับ) ส่วน process อยู่ที่ `D:\project-champ\champ-lib\BC5PROCESS\` (`bcprocess.cpp`, `bcdumpprocjob.cpp`)
> ป้ายไทยที่มาจาก `.rc` = ตัวจริง ส่วนป้ายที่เขียนว่า "(อนุมาน)" ไม่ใช่ข้อความจริงของ Champ เพราะข้อความจริงอยู่ใน DB `myresource` ซึ่งไม่อยู่ใน repo — ความหมายเดาจากชื่อ key และ comment ในโค้ด

### พื้นฐานร่วมของทั้ง 2 จอ

- **ทางที่ build จริง**: โปรแกรมตั้งค่า `theApp.m_bChmMode = TRUE` ไว้ตายตัว (BC5Account.cpp:256) และใช้ `m_nBCModel = BC_MODEL_PRO` (BC5Account.cpp:235) — ทุกกิ่ง `if(theApp.m_bChmMode)` ในเอกสารนี้คือพฤติกรรมจริง ส่วนกิ่ง else ("BC Professional") เป็นทางสำรองที่ไม่ได้ใช้
- **การคำนวณภาษี** ทั้งสองจอใช้ฟังก์ชันกลาง `BillTaxProcess` (lib-frmwrk/BCStools.cpp:155-235)
  - ถ้า **ไม่มีเลขที่ใบกำกับ** → อัตราภาษีถูกบังคับเป็น 0 (:157-158)
  - nTaxType 0 = แยกนอก: ภาษี = ก่อนภาษี×อัตรา/100
  - nTaxType 1 = รวมใน: ภาษี = ยอด×อัตรา/(100+อัตรา)
  - nTaxType 2 = อัตราศูนย์: ภาษี = 0
  - ปัดเศษด้วย `XRound` ตาม `CompanyInfo.AmountPoint` ทุกขั้น
- **กลไกผลข้างเคียง**
  - เมื่อหัวและรายการบันทึกครบ จะ `IsCompleteSave=1` → `Commit` → INSERT `processdoc` → บันทึกภาษีขาย, เลข Serial และ Lot → ลงบัญชี GL (อัตโนมัติผ่าน `m_GLTLSProcess` หรือแบบคีย์เอง) → ประมวลผลลูกหนี้/สต๊อก
  - สต๊อกตัดผ่าน proc job แบบ async ส่วนลูกหนี้ใช้ `CBCProcess::*ARProcess`
- **ตารางภาษีขาย** `BCOutputTax` (DDL ใน atlas.json) เก็บแยกตามแหล่งที่มา:
  - คอลัมน์ `SaveFrom` (myespeciallyformview.cpp:1036): 4 = ใบเพิ่มหนี้, 5 = รับคืนสินค้า
  - คอลัมน์ `Process` (:1035): 1 = บวกภาษี, 2 = ลดภาษี
  - `TaxDate` อยู่ในตารางนี้ ไม่ได้อยู่บนหัวเอกสาร
  - ตารางนี้ **ไม่มีคอลัมน์อ้างใบกำกับเดิม** — การโยงไปใบขายเดิมเก็บใน `BCInvCreditNote` / `BCInvDebitNote1`
- **ยืนยันเอกสาร / ยกเลิกเอกสาร** ใช้ dialog กลาง
  - `CMyDlgBillConfirm` เขียน `ConfirmCode`, `ConfirmDateTime`, `IsConfirm=1` (MyDlgBillConfirm.cpp:64-66)
  - `CMyDlgBillCancel` เขียน `CancelCode`, `CancelDateTime`, `IsCancel=1` (MyDlgBillCancel.cpp:64-66) **ก่อน** เรียก `BillCancelUpdateData` (myespeciallyformview.cpp:5626-5627)
  - ยกเลิกได้ต้องมีสิทธิ์ `bCanCancelBill` และเอกสารต้องยังไม่ผ่าน GL (myespeciallyformview.cpp:5592-5598)
  - ยืนยันได้ต้องมีสิทธิ์ `bCanConfirmBill` (:5648-5651)

---

### [5330] บันทึกใบรับคืนสินค้า,ลดหนี้

- **ตำแหน่งเมนู**
  - Bill (ระบบขาย) › ต่อจาก 5328 บันทึกขายสินค้า,บริการ
  - handler `CMainFrame::OnBillsysTrans7` (MainFrm.cpp:2405-2413; map :398) → `CreateOrActivateFrame(theApp.m_pBILLFrmInvCreditNote, RUNTIME_CLASS(CBILLFrmInvCreditNote))`
  - template ที่ BC5Account.cpp:1579-1584
  - form class `CBILLFrmInvCreditNote` (billfrminvcreditnote.cpp, 7,454 บรรทัด), IDD_BILL_FRM_INV_CREDITNOTE 455×291 (BC5Account.rc:7788)
  - คีย์ชื่อจอ "30008" (billfrminvcreditnote.cpp:51), `m_strBillCode="CN"` (:53), config "InvCreditNoteConfig" (:55), GL source สำหรับพิมพ์ slip = 7 (:75), สิทธิ์ `ID_BILLSYS_TRANS_7` (:90)
- **Variants**
  - `CBILLFrmInvCreditNote4` (BILLFrmInvCreditNote4.cpp) ถูก compile แต่ไม่มีที่ใดสร้าง instance จึงเป็นโค้ดตาย
  - `JJ_PROJECT` คัดลอก DocNo→TaxNo และ DocDate→TaxDate (:6324, :6335)
  - `EXPRESS_PROJ` มีเมนู Key / F8 / F9 แบบย่อ
  - `TOHOME_PROJECT` ถูก comment ออกแล้ว
- **วัตถุประสงค์**: ออกใบรับคืนสินค้า/ใบลดหนี้ให้ลูกค้า โดยอ้างใบขายเดิม (หรือเอกสารยกมา / ลูกหนี้อื่น / เงินมัดจำ) แล้วลดยอดลูกหนี้หรือคืนเงินสด ลดภาษีขาย รับสินค้ากลับเข้าคลัง และลงบัญชี GL

#### ประเภทเอกสาร (CauseType) และผลของแต่ละประเภท

| ค่า | key ป้าย (ไทยจาก comment โค้ด) | สต๊อก | DiscQty (ใช้ลด CNQty ของใบขาย) | รูปแบบ GL (GetGLFormat) | นับใน "ยอดที่ลดหนี้ไปแล้ว" | คอลัมน์ที่แก้ได้ในตาราง (ChmMode) | หลักฐาน |
|---|---|---|---|---|---|---|---|
| 0 | `DEBT_TYPE7` "รับคืน/ลดหนี้" | `StockProcess=0` → **รับเข้าคลัง** (+) ตามคลัง/ที่เก็บบนบรรทัด | = จำนวน | กลุ่ม 12 | นับ | จำนวน (ราคาแก้ได้ถ้ามีสิทธิ์ `bCanChangePrice`) | billfrminvcreditnote.cpp:346, :1466-1478, :5222-5250, :6351-6360 |
| 1 | `DEBT_TYPE2` "ราคาผิด" | `StockProcess=2` → ไม่กระทบสต๊อก | 0 | กลุ่ม 47 | นับ | ราคา + ส่วนลด (จำนวนล็อก) | :347, :1478 (`(n==1)?2:n`), :6362-6372 |
| 2 | `DEBT_TYPE3` "สินค้าแตกเสียหาย" | `StockProcess=2` → **ไม่รับเข้าคลัง** | = จำนวน (ลด CNQty ของใบขาย) | กลุ่ม 48 | **ไม่นับ** (`getAmtHasRefToCN` ข้าม causetype 2) | จำนวน | :348, :1473-1478, :7416-7432 |

- ค่าเริ่มต้น CauseType = 0
- ถ้า AppConfig `oomdts` เปิดอยู่ ระบบบังคับ CauseType = 1 และปิดไม่ให้เปลี่ยน (:349-354)
- การรับเข้าคลังทำใน `bcdumpprocjob.cpp:995-1121`
  - เลือกเฉพาะแถว `stockprocess=0`
  - trans type 5 = `BCCreditNoteSub`, nMark +1 (รับเข้า), ใช้คอลัมน์ `DiscQty` (:2600-2640)
  - ต้นทุนรับคืนใส่ `0 AS MyAmount` (:1330-1333) — วิธีหาต้นทุนจริงของสินค้ารับคืน **ยังไม่ verify**
  - หลังรับเข้า ระบบคำนวณยอดคืนได้คงเหลือของใบขายใหม่ `bcarinvoicesub.cnqty = qty − Σ discqty ของบรรทัดลดหนี้ที่ยังไม่ยกเลิก`

**ตัวเลือกเพิ่มเติม (checkbox)** — บันทึกลงหัวเอกสาร:

| checkbox (ป้ายจริงจาก .rc) | คอลัมน์ | ผล |
|---|---|---|
| ลดยอดใบกำกับภาษีเดิม | `ReturnStatus=1` (:1230) | ใบลดหนี้ `BillBalance=0` (:1267-1270) และ **ตัดยอดคงค้างของใบขายเดิมทันที** (`CreditNoteProcessBill`, bcprocess.cpp:2011-2123) ต้องมีแถวอ้างใบขาย (CANNOT_SAVE_REDUCETAXMUSTREF :857) |
| (ไม่ติ๊ก) | `ReturnStatus=0` | ใบลดหนี้มี `BillBalance=PayBillAmount=BillTemporary=NetDebtAmount` (:1273-1277) เป็นเอกสารเครดิตรอตัดในใบวางบิล/ใบเสร็จ (BillType 2 — bcprocess.cpp:1009-1011) |
| ไม่กระทบต่อยอดลูกหนี้ | `ReturnMoney=1`, `ReturnCash=1` (:1238-1241) | ไม่ประมวลผลลูกหนี้ (`CreditNoteARProcess` เลือกเฉพาะ `ReturnMoney=0` — bcprocess.cpp:2142); ใบขายเงินสดที่อ้างถูกตั้ง `iscreditnote=1`; ต้องคืนเงินผ่าน F7 (`OnSavePayment(10,…)` :1612) ยอดต้องเท่า NetDebt (AMOUNT_NOT_EQUAL_PAYMENT :864); GL กลุ่ม 54 |
| คืนวงเงิน | `DebtLimitState=1` (:1234) ค่าเริ่มต้น **ติ๊ก** (:1877) | ProcessAR คืนวงเงินเครดิตลูกค้า (bDebtLimitProc) |
| ลดหนี้เงินมัดจำ | `IsCNDeposit=1` (:1263) | ตารางอ้างอิงเปลี่ยนไปเลือก `BCARDeposit` (BillType 3) เขียน `BCARDepositUse` (`OnSaveDeposit` :1711-1785) ตั้ง `BCARDeposit.IsReturnMoney=1`; GL กลุ่ม 53 |
| รับคืนเงินสด(รับเป็นเงินสด) — **ซ่อน** (`NOT WS_VISIBLE`, rc:7788-7840) | ไม่มีคอลัมน์ของตัวเอง → บังคับ PayCash | ตั้งอัตโนมัติจากประเภทใบขายแรก: BillType 0/2 (ขายเงินสด) = ติ๊ก, 1/3 (ขายเชื่อ) = ไม่ติ๊ก (:5551-5565; ChmFrmBilling.cpp:1127,1313,527) |

checkbox บางตัวใช้ร่วมกันไม่ได้ (handlers :6759-6885):

| เมื่อติ๊ก | ผลกับ checkbox อื่น |
|---|---|
| ไม่กระทบต่อยอดลูกหนี้ | ล้างและปิด "ลดยอดใบกำกับภาษีเดิม" กับ "คืนวงเงิน" และปิด "ลดหนี้เงินมัดจำ" |
| ลดยอดใบกำกับภาษีเดิม | ล้าง "ไม่กระทบต่อยอดลูกหนี้", "ลดหนี้เงินมัดจำ" และ "รับคืนเงินสด" |
| ลดหนี้เงินมัดจำ | ล้าง "ลดยอดใบกำกับภาษีเดิม", "คืนวงเงิน" และ "ไม่กระทบต่อยอดลูกหนี้"; ถ้าเอาติ๊กออก จะลบแถวเงินมัดจำออกด้วย |

#### หน้าจอและช่องหัวเอกสาร (บน→ล่าง ซ้าย→ขวา; rc:7788-7840, atlas-02-Bill.md:1461-1468)

| # | ป้ายไทย (verbatim) | control → คอลัมน์ `BCCreditNote` | ชนิด | บังคับ? | ค่าเริ่มต้น | แหล่ง lookup | การตรวจ/คำนวณเมื่อออกจากช่อง | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | เลขที่ใบรับคืน/ลดหนี้: (`CREDITNOTE_NO`) | IDC_ED_DOCNO → `DocNo` | text ตัวใหญ่ + ปุ่มรันเลข | ใช่ (CANNOT_SAVE_NOTDOCNO) | รันอัตโนมัติ `SetOption(BCCreditNote, DocNo)` ตามวันที่ทำงาน | รูปแบบเลขเอกสาร | >20 ตัว → DOCNO_MORE_LENG แล้วตัดเหลือ 20; เปลี่ยนเลขตอนแก้ → ตรวจซ้ำ DOCNO_DUPLICATE; ไม่มีสิทธิ์ `bCanChangeBillNo` = ปิดช่อง | :6278-6316, ConfirmSave :803-1152 |
| 2 | วันที่ออกเอกสาร: (`DOCNO_DATE`) | IDC_ED_DOCDATE → `DocDate`; `DueDate` = DocDate (:1173) | date (พ.ศ. auto) | ใช่ | WorkingDate | — | `IsPeriodClose` (ปิดงวดแล้วห้ามบันทึก/ลบ); ไม่มีสิทธิ์ `IsCanEditDocDate` และไม่ใช่ sa = ปิดช่อง | :409-411, :1787-1934 |
| 3 | เลขที่ใบกำกับ: (`TAXNO`) | IDC_ED_TAXNO → `TaxNo` | text + ปุ่มรันเลข | ไม่ (ถ้าว่าง = ไม่มี VAT) | ว่าง (โค้ดคัดลอกจาก DocNo ถูก comment ออก :6284) | รูปแบบเลขใบกำกับ | >20 → TAXNO_MORE_LENG แล้วตัด; คำนวณภาษีใหม่ (`CalcTaxAmount`); ถ้ามี TaxNo ตอนเลือกใบขายจะกรองเฉพาะใบที่มี TaxNo (และกลับกัน) | :5252, :5331-5614 |
| 4 | วันที่ใบกำกับ: (`TAXDATE`) | IDC_ED_TAXDATE → **ไม่ได้เขียนลงหัวเอกสาร** แต่ไปอยู่ที่ `BCOutputTax.TaxDate` ผ่าน `MyUpdateFirstTaxRow` | date | ต้องมีถ้ามี TaxNo (CHECK_TAXDATE) | WorkingDate | — | อัปเดตแถวภาษีแรก | :875, :7086-7121 |
| 5 | ประเภทภาษี: (`TAX_TYPE`) | IDC_ED_TAXTYPE → `TaxType` | combo `CB_TAXTYPE_1` / `CB_TAXTYPE_2` / `TAX_ZERO` = 0 แยกนอก / 1 รวมใน / 2 อัตราศูนย์ (ป้าย combo อนุมาน) | ใช่ | `CompanyInfo.DefaultTaxType` → ทับด้วย `BCAR.TaxType` และ TaxType ของใบขายแรก | — | `OnSelchangeTaxType`: ประเภท 2 → อัตรา 0; อย่างอื่นใช้อัตราของลูกค้า แล้วค่อยอัตราตั้งต้นของบริษัท → `BillCalcInv` | :335-341, :7231 |
| 6 | รหัสแผนก: (`DEPART_CODE`) | IDC_ED_DEPART → `DepartCode` | combo-table | ไม่ | ว่าง | `BCDepartment` | — | OnInitialUpdate :187-417 |
| 7 | รหัสลูกค้า: (`ARCODE`) | IDC_ED_ARCODE → `ArCode` (+ ปุ่มค้นหา) | lookup | ใช่ (`m_bMustEntry`, CANNOT_SAVE_NOTARCODE) | ว่าง | `BCAR.Code` (ArTransConfig) | ดึงชื่อ (Name1 หรือ Name2 ตาม `DefaultArName`), TaxType, TaxRate (`BCAR.DefaultTaxRate` หรือของบริษัท), SaleCode, ชื่อ/ที่อยู่ออกบิล; Enter → ไปตารางใบขาย | :4187-4225, :4479-4535, :4244 |
| 8 | (ชื่อลูกค้า) | IDC_ED_ARNAME | read-only | — | — | BCAR | — | rc:7788-7840 |
| 9 | ประเภทรับคืน/ลดหนี้: (`AR_CREDITNOTE_TYPE`) | IDC_ED_CREDITTYPE → `CauseType` (ลงทั้งหัวและทุกบรรทัด) | combo 0/1/2 (ตารางด้านบน) | ใช่ | 0 | — | ตั้ง shadow25 ของทุกบรรทัด เปลี่ยนคอลัมน์ที่แก้ได้ และโหลดรูปแบบ GL ใหม่ | :344-354, :6339-6412 |
| 10 | ลดยอดใบกำกับภาษีเดิม (`REDOCE_OLD_TAX`) | IDC_CHECK_REDUCETAX → `ReturnStatus` | checkbox | — | ไม่ติ๊ก | — | ดูตารางตัวเลือก | :1230 |
| 11 | ลดหนี้เงินมัดจำ (`DEPOSIT_CREDITNOTE`) | IDC_CHECK_DEPOSIT → `IsCNDeposit` | checkbox | — | ไม่ติ๊ก | — | — | :6804 |
| 12 | ปุ่ม ลดหนี้ไม่อ้างบิลเดิม (`DN_NOREF_DOCNO`) | IDC_BTTN_CONSTRUCT | ปุ่ม | — | **ซ่อนเมื่อ ChmMode** | dialog `CBillDlgConstructionDlg` "ข้อแนะนำ" (rc:14945) | — | :405-408 |
| 13 | ไม่กระทบต่อยอดลูกหนี้ (`CASH_CREDITNOTE_CHM`) | IDC_CHECK_PAYCASH → `ReturnMoney` / `ReturnCash` | checkbox | — | ไม่ติ๊ก | — | — | :1238-1245 |
| 14 | คืนวงเงิน (`PUSHCASH`) | IDC_CHECK_PUSHCASH → `DebtLimitState` | checkbox | — | **ติ๊ก** | — | — | :1234, :1877 |
| 15 | รับคืนเงินสด(รับเป็นเงินสด) (`CREDITNOTE_CASH`) | IDC_CHECK_RETURNCASH | checkbox ซ่อน | — | อัตโนมัติ | ประเภทใบขาย | — | :5551-5565 |

**F5 "รายละเอียดส่วนอื่นๆ"** — `CBILLDlgCreditNote` (BC5Account.rc:8386; BILLDlgCreditNote.cpp) — ตามลำดับบนจอ:

| ซ้าย (ป้ายจริง) | คอลัมน์ | ขวา (ป้ายจริง) | คอลัมน์ |
|---|---|---|---|
| ชื่อในการออกบิล: | `ARName` | — | — |
| ที่อยู่ในการออกบิล: | `ARAddress` | — | — |
| สาเหตุการรับคืน/ลดหนี้: | `CauseCode` (combo `BCCNMaster`; ChmMode ค่าเริ่มต้น "01"/"02"/"03" = CauseType+1, BILLDlgCreditNote.cpp:153) | (Job ID — ซ่อน) | — |
| โครงการ: | `ProjectCode` (`BCProject`) | มูลค่าส่วนลด: | `DiscountAmount` (ปิดเสมอ) |
| การจัดสรร: | `AllocateCode` (`BCAllocate`) | ยอดก่อนภาษี: | `SumOfBeforeTax` |
| รหัสพนักงานขาย: | `SaleCode` (`BCSale`; `UserGroup` ← `BCSale.UserGroup`) | ยอดภาษี: | `SumOfTaxAmount` |
| กลุ่มของบิล: | `BillGroup` (= TaxGroup ของภาษีขาย) | ยอดรวมภาษี: | `SumOfTotalTax` |
| อัตราภาษีมูลค่าเพิ่ม: | `TaxRate` | ยอดยกเว้นภาษี: | `SumOfExceptTax` |
| ยอดใบกำกับเดิม: | `SumOldAmount` | ยอดผลต่าง/ลดหนี้: | `NetDebtAmount` |
| รวมมูลค่าสินค้า: | `SumOfItemAmount` | ยอดที่ถูกต้อง: | `SumTrueAmount` |
| หมายเหตุ: | `MyDescription` | | |
| สกุลเงิน: | `CurrencyCode` (`BCCurrency`; ปิดถ้าไม่ใช้หลายสกุล) | อัตราแลกเปลี่ยน: | `ExchangeRate` (≤0 → 1) |

- ปุ่ม F11=ตกลง / Esc=ยกเลิก
- ChmMode ปิดช่องยอดสรุปทั้งหมด (ยอดเดิม, ยอดที่ถูกต้อง, รวมสินค้า, ก่อนภาษี, ภาษี, ยกเว้น, สุทธิ, รวมภาษี) — แสดงอย่างเดียว (BILLDlgCreditNote.cpp:81-183)
- อัตราภาษีและยอดภาษีแก้ได้ตามสิทธิ์ `IsUserEditTax` (:177-178)
- เมื่อกด OK ระบบเรียก `DetailCalc(i,9)` ใหม่ทุกบรรทัด (billfrminvcreditnote.cpp:5016-5031)

#### ตารางรายการ (grid) — ตารางบน "ใบขายที่อ้างอิง" (`ID_INVOICE_LIST`, rect IDC_LB_INVOICE; :207-211)

| # | ป้าย (key; ไทยจาก comment โค้ด) | กว้าง | แก้ได้? | ชนิด | ความหมาย/สูตร | หลักฐาน |
|---|---|---|---|---|---|---|
| 0 | `DOCNO` "เลขที่ใบ invoice" | 140 | แก้ได้ + ค้นหา | text | เลขที่ใบขาย (หรือเอกสารยกมา / ลูกหนี้อื่น / มัดจำ) | :207, :4537-4776 |
| 1 | `ODL_DISCOUNT` "ส่วนลดเดิมของ invoice" | 120 | แก้ได้ | text ส่วนลด | ส่วนลดท้ายบิลของใบเดิม ใช้ปันส่วนลดจากยอดสินค้าที่คืน | :208, BillCalc :5033-5150 |
| 2 | `OLD_AMOUNT` "ยอดเดิม" | 120 | แก้ได้ | เงิน | ChmMode: = NetDebt ของใบขาย + มัดจำที่หัก (ถ้าแยกนอก เงินมัดจำบวก VAT ก่อน) | :209, :5331-5614 |
| 3 | `DIFF_AMOUNT_SUMVAT` "ยอดผลต่าง" | 120 | แก้ได้ | เงิน | ยอดลดหนี้ **รวม VAT** ของใบนี้ = Σ ยอดสินค้าที่อ้างใบนี้ − ส่วนลด (แยกนอก+มีใบกำกับ → ×(1+อัตรา/100)) | :210, :5033-5150 |
| 4 | `TRUE_AMOUNT` "ยอดที่ถูกต้อง" | 120 | ไม่ | เงิน | = ยอดเดิม − ผลต่าง − (ศูนย์+ยกเว้น) (ถ้าผลต่าง > ยอดเดิม = 0) | :211, :4148-4185 |

**Shadow columns ของตารางบน**

| # | ชื่อ | # | ชื่อ |
|---|---|---|---|
| 0 | RowOrder | 9 | ZeroTaxAmount |
| 1 | InvExchangeRate | 10 | ExceptTaxAmount |
| 2 | NewExchangeRate | 11 | BillBalance (ยอดคงค้างใบเดิม) |
| 3 | HomeAmount1 | 12 | CNAmount (ลดหนี้ไปแล้ว) |
| 4 | HomeAmount2 | 13 | OldDiffAmount |
| 5 | ExchangeProfit | 14 | CNItemAmount (เพดานจากสินค้า) |
| 6 | BillType — 0 ใบส่งของ / 1 เอกสารยกมา / 2 ตั้งลูกหนี้อื่นๆ / 3 มัดจำ | 15 | QARefno |
| 7 | SumOfItemAmount | 16 | JobID |
| 8 | DiscountAmount | | |

**การเลือกใบขาย** (คลิกค้นหาที่คอลัมน์ 0 — `OnListFindDataInv` :4537-4600)
- ต้องกรอกรหัสลูกค้าก่อน (PleaseInsertDataInArCode :4558)
- เปิด popup menu ให้เลือกแหล่ง 3 แบบ: `INVOICE` (ใบส่งของ) / `INVOICE_YEAR` / `OTHER_AR`
  - ถ้าติ๊ก "รับคืนเงินสด" → ข้ามเมนู ไปเลือกใบขายทันที
  - ถ้าติ๊ก "ลดหนี้เงินมัดจำ" → ไปเลือก `BCARDeposit` (:6915-7065)

| แหล่ง | ฟังก์ชัน / config | เงื่อนไขที่ใช้กรอง | ค่าที่ตั้งให้ |
|---|---|---|---|
| ใบขาย | `BrowseInvoice` :5331-5614, config `BillingTransConfig` (เลือกได้หลายใบ) | `bcarinvoice` ของลูกค้านี้ที่ยังไม่อยู่ในตาราง, iscancel=0, iscreditnote=0, posstatus=0, taxtype ตรงกับจอ, มี/ไม่มี TaxNo ตรงกับจอ; โหมดไม่กระทบลูกหนี้ → เฉพาะ billtype 0/2; ไม่ใช่ → paybillstatus=0 และ billbalance>0 | ใบแรกกำหนดสกุลเงิน อัตราแลกเปลี่ยน อัตราภาษี และชื่อ/ที่อยู่ออกบิล; ถ้าใบนั้นเคยออกใบลดหนี้ประเภท 1 แล้วจะเตือน INVOICE_CREDITNOTE (:5597) |
| เอกสารยกมา | `Balance_Year` :5632-5791, `ArINVBalanceConfig` | BillBalance>0, PayBillStatus=0; ภาษีต้องเป็นรวมใน (INV_BAL_PLS_INC_VAT :5636) | BillType=1; ยอดเดิม=NetDebt, ผลต่าง=BillBalance |
| ลูกหนี้อื่น | `OtherDebt` :5794-5954, `ArOtherDebtConfig` | — | BillType=2 (ตาราง `BCAROtherDebt`) |

- **กันลดหนี้เกิน (ตอนเลือก)**: ถ้า `getAmtHasRefToCN` ≥ ยอดสุทธิ + มัดจำ → REF_CN_EXHAUST แล้วไม่ให้เลือก (:5463, :7416-7432)
- **พิมพ์เลขที่เอง** (:4602-4776, :6528-6741)
  - เลขต้องมีอยู่ใน bcarinvoice ของลูกค้านี้ → ไม่พบ = NOT_FOUND_INVOICENO (:6444); "." = ใช้ได้
  - เลขซ้ำในตาราง → DOCNO_DUPLICATE

#### ตารางรายการ (grid) — ตารางล่าง "สินค้ารับคืน" (`ID_BILLING_LIST`, rect IDC_LB_RECT; :256-268)

| # | ป้าย (key) | กว้าง | แก้ได้? (ChmMode) | ชนิด | ความหมาย/สูตร | หลักฐาน |
|---|---|---|---|---|---|---|
| 0 | `ITEMCODE` รหัสสินค้า | 60 | แก้ได้ + ค้นหา | lookup | ไม่พบ → NOT_SEEITEMCODE | :256, :3461-3543 |
| 1 | `ITEMNAME` ชื่อสินค้า | 120 | แก้ได้เฉพาะ `Option.IsEditItemName==1` | text | — | :257 |
| 2 | `INVOICE_NO` "เอกสารอ้างอิง" (comment) | 70 | แก้ได้ | text | เลขใบขายที่บรรทัดนี้อ้าง (ใช้จับคู่กับตารางบนใน BillCalc) | :258 |
| 3 | `WAREHOUSE_ITEM` คลังสินค้า | 60 | combo | `BCItemWarehouse` ตาม `GetSaleWHCondition` | คลังที่รับคืน (ค่าเริ่มต้น = คลังของบรรทัดใบขาย หรือ `BCItem.DefSaleWHCode`) | :259, :3941 |
| 4 | `SHELF` ที่เก็บ | 60 | combo | `BCItemWarehouse`/`BCShelf` | ที่เก็บ | :260 |
| 5 | `UNITCODE` หน่วยนับ | 60 | combo | `BCItemUnit` / `BCStkPacking` | ตรวจ `IsUnitOfItem`; หน่วยว่าง → ข้อความไทยฝังโค้ด "ไม่ได้กำหนดหน่วยนับเดิม" (:3369) | :261, :3347-3417 |
| 6 | `ISSUERET_NUMBER` (อนุมาน: จำนวนรับคืน) | 60 | แก้ได้เมื่อประเภท 0/2; ล็อกเมื่อประเภท 1 | qty (`QtyPoint`) | ถ้ามีใบอ้าง จำนวนแปลงเป็นหน่วยฐานต้อง ≤ BillQty (shadow18) ไม่งั้น NUM_ITEM_NOTMORE แล้วคืนค่าเดิม | :262, :3507, :6353 |
| 7 | `PERPRICE` (อนุมาน: ราคา/หน่วย) | 60 | ประเภท 0/2 แก้ได้ถ้ามีสิทธิ์ `bCanChangePrice`; ประเภท 1 แก้ได้ | price (`PricePoint`) | ประเภท 1: ราคาใหม่ต้อง ≤ ราคาเดิม (shadow28) ไม่งั้น INCORRECT_CREDIT_PRICE | :263, :3523, :6359-6372 |
| 8 | `DISCOUNT` ส่วนลด | 70 | แก้ได้เฉพาะ `CompanyInfo.DetailDiscount≠0` และประเภท 1 | text ส่วนลด | — | :264-267, :6366 |
| 9 | `AMOUNT` จำนวนเงิน | 70 | ไม่ (ChmMode) | เงิน | = จำนวน×ราคา − ส่วนลด | :268, :1820-1828 |

**Shadow columns ของตารางล่าง** (:273-315)

| # | ชื่อ | # | ชื่อ | # | ชื่อ |
|---|---|---|---|---|---|
| 0 | RowOrder | 14 | WHCode | 28 | OldPrice (ราคาเดิม) |
| 1 | ItemType | 15 | ShelfCode | 29 | LineNumber |
| 2 | MyDescription | 16 | NetAmount | 30 | RefLineNumber (บรรทัดของใบขาย) |
| 3 | ProjectCode | 17 | HomeAmount | 31 | LotNumber |
| 4 | AllocateCode | 18 | BillQty (จำนวนคืนได้คงเหลือ) | 32 | LotExpireDate |
| 5 | OppositeQty | 19 | RefDocNo | 33 | LotMyDesc |
| 6 | OppositeUnit | 20 | OldQty | 34 | OldWHCode |
| 7 | ExceptTax | 21 | SetCode | 35 | OldShelfCode |
| 8 | BarCode | 22 | DepartCode | 36 | OldUnitCode |
| 9 | CategoryCode | 23 | CurrencyCode | 37 | ColorCode |
| 10 | GroupCode | 24 | ExchangeRate | 38 | SizeCode |
| 11 | BrandCode | 25 | StockProcess | 39 | StyleCode |
| 12 | TypeCode | 26 | LotScript | 40 | BehindIndex |
| 13 | FormatCode | 27 | DiscountAmount | 41 | OldItemCode |

**ค้นหาสินค้าที่คอลัมน์ 0** (`OnListFindData` :3661-3940)

| วิธีกด | ผล |
|---|---|
| Shift+คลิก | เปิด `CMyFilterItemDlg` |
| Ctrl+คลิก | เปิดรายการสินค้าหลัก (`ItemTransConfig`) |
| คลิกปกติ | ต้องมีใบขายในตารางบนก่อน (INPUT_DOCNO :3736) แล้วเปิดรายการบรรทัด `bcarinvoicesub` ที่ยังคืนได้ (`cnqty>0`) ของใบที่เลือก |

เมื่อเลือกบรรทัดจากใบขาย ระบบเติมค่าให้ดังนี้:

| ช่อง | ค่าที่เติม |
|---|---|
| จำนวน | CNQty ที่เหลือ |
| ราคา | ราคาในใบขาย |
| คลัง / ที่เก็บ / หน่วย | ตามบรรทัดใบขาย |
| shadow18 | จำนวนในใบขาย |
| shadow28 | ราคาในใบขาย |
| shadow30 | LineNumber ของบรรทัดใบขาย |

**F9 ดึงรายการ** (`LoadInvoiceSub` :5955-6130)
- ใช้ได้เฉพาะ BillType 0
- ดึงบรรทัด `cnqty>0` ที่ยังไม่อยู่ในตาราง
- ข้ามสินค้าที่ `BCItem.ReturnStatus==1` (ห้ามรับคืน; `CheckReturnStatus` :6743)
- จำนวน = CNQty ที่เหลือ และคำนวณยอดเงินใหม่

#### ยอดรวมท้ายเอกสาร / สูตรคำนวณ

1. **ระดับบรรทัด** `DetailCalc` (:4082-4146)
   - ถ้าแก้จำนวนเงินโดยไม่มีราคา → ราคา = เงิน/จำนวน
   - ถ้า `DiscountMethod==0` → เงิน = จำนวน×ราคา − ส่วนลด; ถ้าไม่ใช่ → หักส่วนลดต่อหน่วยก่อนคูณจำนวน
   - ปัดเศษตาม AmountPoint / PricePoint
   - `NetAmount` = เงิน (กรณีรวมใน + มีใบกำกับ + ไม่ยกเว้นภาษี → เงิน − เงิน×อัตรา/(100+อัตรา))
   - `HomeAmount` = NetAmount × อัตราแลกเปลี่ยน
2. **ระดับใบขายที่อ้าง** `BillCalc` (:5033-5150)
   - รวมยอดสินค้าที่ `INVOICE_NO` ตรงกับใบนั้น (ถ้าเลขใบเป็น "." = รวมทุกบรรทัด)
   - แยกยอดเป็น 3 กอง: ยกเว้นภาษี (ExceptTax=1 และบริษัทเปิดใช้), อัตราศูนย์ (=2 และบริษัทเปิดใช้), มีภาษี
   - หักส่วนลดเดิม (คอลัมน์ 1)
   - ถ้าแยกนอก + มีใบกำกับ → ผลต่าง = ยอดมีภาษี×(1+อัตรา/100) ปัดเศษ
   - ยอดที่ถูกต้อง = ยอดเดิม − ผลต่าง − (ศูนย์+ยกเว้น)
3. **ท้ายเอกสาร** `BillCalcInv` (:5152-5220)

   | ยอด | สูตร |
   |---|---|
   | SumTaxAmount | Σ ผลต่าง |
   | NetDebtAmount | SumTax + ยกเว้น + ศูนย์ − ภาษีหัก ณ ที่จ่าย |
   | ยอดที่ถูกต้อง | ยอดเดิม − NetDebt |
   | ยอดภาษี | SumTax×อัตรา/(100+อัตรา) ปัดเศษ |
   | ยอดก่อนภาษี | SumTax − ภาษี (ถ้าไม่มี TaxNo = SumTax) |

   แล้วเรียก `CalcTaxAmount` (:5285-5329) ต่อ:
   - ส่งเข้า `BillTaxProcess` แบบ **รวมใน (nTaxType=1) เสมอ** ถ้า TaxType≠2 → ยอดผลต่างถือเป็นยอดรวม VAT
   - ฐานคำนวณ = NetDebt ถ้าเป็นลดหนี้มัดจำหรือไม่มีรายการสินค้า; ไม่งั้น = SumTaxAmount
4. **แถวภาษีขาย** `MyUpdateFirstTaxRow` (:7086-7121)
   - `UpdateFirstTaxRow(TaxNo, TaxDate, อัตรา, ก่อนภาษี, ยกเว้น, ภาษี, คำอธิบาย key "RECEIPT_GOODS", "", ศูนย์)`
   - ถ้า TaxType 2 → ย้ายยอดก่อนภาษีไปเป็นยอดอัตราศูนย์
   - หลายสกุลเงิน → คูณอัตราแลกเปลี่ยนแล้วปัดเศษ
   - กด F3 แก้แถวภาษีได้ แล้วระบบคัดลอก TaxNo/TaxDate แถวแรกกลับมาที่จอ

#### ปุ่ม แถบเครื่องมือ คีย์ลัด (`PreTranslateMessage` :4244-4477; เมนู Key `OnMore` :4847-5014)

| คีย์ | ผล |
|---|---|
| Enter ที่รหัสลูกค้า | ไปตารางใบขาย |
| PageUp / PageDown | ไปตารางใบขาย / ตารางสินค้า |
| F2 / Shift+F2 / Ctrl+F2 | ลงบัญชี GL (`OnAccount`) / รายละเอียดบรรทัด (`CBILLDlgExpandArCreditNote` "รายละเอียดสินค้าอื่นๆ" rc:8295) / ไม่มีผล |
| F3 | ภาษีขาย (`OnOutputTax("InvCreditNoteConfig")`) |
| F5 / Shift+F5 / Ctrl+F5 | รายละเอียดส่วนอื่นๆ / ยืนยันเอกสาร (ต้องมี `bCanConfirmBill` ไม่งั้น NOT_CONFIRM_NO) / ยกเลิกเอกสาร (`OnConfirmCancel` → `BillCancel("BCCreditNote")`) |
| F6 | ภาษีหัก ณ ที่จ่าย (`OnWARTax`) — ใช้ได้เฉพาะเมื่อติ๊ก "ไม่กระทบต่อยอดลูกหนี้" หรือ "รับคืนเงินสด" ระบบคำนวณ NetDebt = ยอดรวม − WHT |
| F7 | รับ/คืนเงิน (`OnReceipt` → `OnRec(4, NetDebt)`) — เฉพาะโหมดไม่กระทบต่อยอดลูกหนี้ |
| Ctrl+Shift+F7 | ยูทิลิตี้คำนวณยอดหัวเอกสารใหม่ (`OnProcessNewvalueForBCCreditNote` :7169) |
| F9 / Ctrl+F9 / Shift+F9 | ดึงรายการจากใบขาย / Serial / Lot (FIFO เท่านั้น) |
| F11 | บันทึก |
| F12 / Ctrl+F12 | เส้นทางเอกสาร / ข้อมูลสินค้า `CChmItemDisplayDlg` |
| ปุ่มขวาสุดของแถวใบขาย | `CBILLDlgCreditExpand` "Other" (rc:8135) — อัตราแลกเปลี่ยนเดิม/ใหม่, ยอดลดตามสกุลเงินเดิม/ใหม่, กำไร(ขาดทุน), ประเภทเอกสาร (:6192-6220) |

dialog รายละเอียดบรรทัด (`CBILLDlgExpandArCreditNote`) มีช่อง:
- ผลต่อสต๊อก (combo `UPDATE_QTY_COST` / `NOT_UPDATE`; ChmMode ปิดไม่ให้แก้ — BILLDlgExpandArCreditNote.cpp:82-87, :136-137)
- แผนก, โครงการ, การจัดสรร, คำอธิบายรายการ
- เอกสารอ้างอิง, จำนวนเดิม, ยอดเงินสุทธิ, ยอดเงินบาทสุทธิ
- คุณสมบัติสินค้า: รหัสกลุ่ม, หมวดสินค้า, ประเภทสินค้า, รหัสยี่ห้อ, color, size, สไตล์

#### สถานะเอกสารและ workflow

| ขั้น | เงื่อนไข / ผล | หลักฐาน |
|---|---|---|
| บันทึกใหม่ | หัว `IsCompleteSave=0` → บรรทัดครบ → `IsCompleteSave=1` + Commit; บันทึกไม่ครบ → Abort + SAVEDETAIL_NOTCOMPLETE | :1154-1709 |
| แก้ไข | ห้ามเมื่อ: ผ่าน GL (CANNOT_SAVE_POSTGL), ยกเลิกแล้ว (NOT_SAVE_ISCANCEL), ยืนยันแล้ว (NOT_SAVE_ISCONFIRM), วางบิลแล้ว (CANNOT_EDIT_BILL1), หรือถูกตัดชำระบางส่วน — กรณีไม่ได้ติ๊กลดยอดใบเดิมและไม่ได้ติ๊กไม่กระทบลูกหนี้ ต้อง NetDebt = BillBalance (NOT_EDIT_BILLBALANCE) — การแก้จะกลับรายการลูกหนี้/ใบขายเดิมก่อน (`CreditNoteARProcess(old,-1)`, `CreditNoteProcessBill(old,-1)` :1160-1161) | :880-905 |
| ยืนยัน (Shift+F5) | `IsConfirm=1` + ผู้ยืนยัน/เวลา | MyDlgBillConfirm.cpp:64-66 |
| ยกเลิก (Ctrl+F5) | ห้ามเมื่อ: ผ่าน GL / ยกเลิกแล้ว (MSG_CANNOT_CANCEL) / ยืนยันแล้ว / วางบิลแล้ว (CANNOT_CANCEL) / ถูกใช้ในใบวางบิลหรือใบเสร็จ BillType 2 (CANNOT_CANCEL_USE) → ทำ `BillCancelUpdateData` | :7275-7317, :566-660 |
| ลบ | ห้ามเมื่อ: เลขที่ว่าง / ปิดงวดแล้ว / ผ่าน GL (CANNOT_DELETE_POSGL) / ยกเลิกแล้ว / ยืนยันแล้ว / วางบิลแล้ว (CANNOT_DELETE_BILL1) / ถูกตัดชำระแล้ว (NOT_DELETE_BILLBALANCE) / Serial (CANNOT_DELETE_SERIALSTATUS, …SERIALWH) → ถาม CONFIRM_DELETEDATA | :498-564 |
| ถูกอ้างต่อ | ใบลดหนี้ที่มียอดคงเหลือถูกตัดในใบวางบิล/ใบเสร็จเป็น BillType 2 (`BCPaybillSub`, `BCReceiptSub1/2`) | :6887, bcprocess.cpp:1009-1011 |
| ผ่าน GL | flag `IsPostGL` (ผ่านจากโมดูล GL) → ห้ามแก้ ลบ ยกเลิก | `CMyBaseDBFormView::IsPostGL` mybasedbformview.cpp:4126 |

จอแสดงป้ายสถานะ (ยืนยัน/ยกเลิก/ยอดคงเหลือ) ด้วย `DrawBillStatus` (OnLoadData :1936-2420)

#### ผลข้างเคียง (ตาราง/คอลัมน์ที่เขียน)

**`BCCreditNote` — หัวเอกสาร** (:1167-1302)

| กลุ่ม | คอลัมน์ |
|---|---|
| ข้อมูลเอกสาร | TaxType, CauseType, DocNo, DocDate, ArCode, DueDate, TaxNo, DepartCode, UserGroup, SaleCode, TaxRate, MyDescription, CurrencyCode, ARName, ARAddress, SCGBPNumber, ExchangeRate, AllocateCode, ProjectCode, BillGroup, RecurName, CauseCode |
| ยอดเงิน | DiscountWord, DiscountAmount, SumOldAmount, SumTrueAmount, SumOfBeforeTax, SumOfTaxAmount, SumOfTotalTax, SumOfExceptTax, NetDebtAmount, SumOfZeroTax, SumOfItemAmount |
| flag | IsPos=0, ReturnStatus, DebtLimitState, ReturnMoney, ReturnCash, IsCNDeposit |
| ยอดรับ/จ่ายเงิน | ExcessAmount1/2, OtherIncome, OtherExpense, SumCashAmount, SumChqAmount, SumCreditAmount, ChargeAmount, SumBankAmount, PettyCashAmount |
| ยอดคงค้าง | BillBalance, PayBillAmount, BillTemporary |
| การลงบัญชี | GLFormat, GLTransData (0 = อัตโนมัติ / 1 = คีย์เอง), GLStartPosting=0, SCGID, IsCompleteSave |
| ผู้บันทึก | CreatorCode/CreateDateTime หรือ LastEditorCode/LastEditDateT |

**ตารางอื่นที่เขียน**

| ตาราง | สิ่งที่เขียน / ผล | หลักฐาน |
|---|---|---|
| `BCInvCreditNote` (ลบแล้วเขียนใหม่ต่อแถวใบขาย) | CreditNoteNo, DocDate, ArCode, ReturnMoney, ReturnStatus, InvoiceNo, InvoiceDisc, OldAmount, DiffAmount, TrueAmount, InvExchangeRate, NewExchangeRate, HomeAmount1, HomeAmount2 = (ผลต่าง+ศูนย์+ยกเว้น)×อัตรา, ExchangeProfit, BillType, SumOfItemAmount, DiscountAmount, ZeroTaxAmount, ExceptTaxAmount, QARefno, JobID | :1320-1397 |
| `BCCreditNoteSub` (ลบแล้วเขียนใหม่ต่อบรรทัด) | DocNo, TaxNo, TaxType, CauseType, DocDate, ArCode, SaleCode, UserGroup, ItemCode, ItemName, InvoiceNo, WHCode, ShelfCode, UnitCode, Price, DiscountWord, Amount, DiscQty, TempQty=จำนวน, StockProcess, DiscountAmount, ItemType, MyDescription, ProjectCode, AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, Category/Group/Brand/Type/FormatCode, NetAmount, HomeAmount, BillQty, DepartCode, CurrencyCode, ExChangeRate, RefLineNumber, Lot*, Color/Size/StyleCode, IsPos=0, TaxRate, PackingRate1/2, behindindex, IsProcess, LineNumber | :1420-1549 |
| `processdoc` | (DocNo, DocType=5) — คิวให้ GL/proc | :1590 |
| ภาษีขาย `BCOutputTax` | `OnSaveOutputTax(5, old, doc, date, ar, Process=2 ลดภาษี, Source=7, BillGroup, multicurrency)` | :1602 |
| ภาษีหัก ณ ที่จ่าย | `OnSaveAPWTax(10, …)` | :1605 |
| รับ/จ่ายเงิน | `OnSavePayment(10, …)` เมื่อไม่กระทบลูกหนี้หรือรับคืนเงินสด | :1612 |
| เงินมัดจำ | `OnSaveDeposit` → `BCARDepositUse`: Amount = ผลต่าง; NetAmount = ตัด VAT ถ้ามีใบกำกับ | :1617, :1711-1785 |
| GL | อัตโนมัติ (`m_GLTLSProcess` Source 7, GLGroup 7) หรือคีย์เอง (`OnSaveAccount(…,7,…)`) — **ผังบัญชีมาจากรูปแบบ GL** (`GetDefaultGLFormat(กลุ่ม 12/47/48/53/54)`) ฟอร์มไม่มีรหัสบัญชีเอง | :1623-1630, :5222-5250 |
| สต๊อก | `CreditNoteStockProcess(doc)` (เมื่อ `AutoServerProc≠1`) → รับเข้าคลังเฉพาะ StockProcess 0 + ปรับ `bcarinvoicesub.cnqty` | :1634, bcdumpprocjob.cpp:995-1121 |
| ลูกหนี้ | `CreditNoteARProcess(doc)`: ใช้เฉพาะ ReturnMoney=0 → `ProcessAR(ar, งวด, NetDebt, …, mark×−1, คืนวงเงินถ้า DebtLimitState=1)` + `UpdateARPeriodState(… คอลัมน์ CN …)`; ถ้า IsCNDeposit → `ProcessARDeposit` | :1636, bcprocess.cpp:2129-2336 |
| ใบขายเดิม | `CreditNoteProcessBill(doc)` ทำต่อแถว `BCInvCreditNote`: | :1637, bcprocess.cpp:2011-2123 |

`CreditNoteProcessBill` แยกตามกรณี:

| กรณี | ผลกับใบขายเดิม |
|---|---|
| ReturnMoney=1 และ BillType 0 | `bcarinvoice.iscreditnote=1` |
| BillType 0 และ ReturnStatus=1 | ลด billbalance / paybillamount / billtemporary ของใบขายด้วยยอดผลต่าง |
| BillType 0 และ ReturnStatus=0 | iscreditnote = (TrueAmount==0) |
| BillType 1 และ ReturnStatus=1 | ลดยอดใน `bcarinvbalance` |
| BillType 2 และ ReturnStatus=1 | ลดยอดใน `bcarotherdebt` |

#### Guard / ข้อความเตือนสำคัญ

| key | เงื่อนไข | หลักฐาน |
|---|---|---|
| CANNOT_SAVE_NOTDOCNO / DOCNO_DUPLICATE / "Document No is duplicated\nDo you want to try again?" | ไม่มีเลขที่ / เลขซ้ำ (ระบบเสนอรันเลขใหม่) | ConfirmSave :803-1152 |
| CANNOT_SAVE_NOTARCODE | ไม่มีรหัสลูกค้า | ConfirmSave |
| CANNOT_SAVE_REDUCETAXMUSTREF | ติ๊กลดยอดใบกำกับเดิมแต่ไม่มีแถวใบขาย | :857 |
| AMOUNT_NOT_EQUAL_PAYMENT | ไม่กระทบลูกหนี้ แต่ยอดรับ/จ่ายเงิน ≠ NetDebt (แล้วเปิด F7) | :864 |
| CHECK_TAXDATE | มีเลขที่ใบกำกับแต่ไม่มีวันที่ | :875 |
| NOT_SET_WHCODE / NOT_SET_SHELFCODE / NOT_SET_UNITCODE | สินค้าที่ไม่ใช่บริการ (StockType≠1,7) ขาดคลัง/ที่เก็บ/หน่วย | :932 |
| NOT_SPECIFY_SERIAL / SPECIFY_NOTEQUAL_SERIAL / CANNOT_EDIT_SERIALWH | สินค้า Serial (ยกเว้นประเภท 1) | ConfirmSave |
| NOT_SET_LOT | สินค้าต้นทุน FIFO ไม่ได้ระบุ Lot | ConfirmSave |
| CANNOT_SAVE + ประเภทเอกสาร + REF_CN_EXHAUST | **ลดหนี้เกินใบเดิม** `IsInvoiceOverCN`: ผลต่างใหม่ + ลดหนี้เดิม − ยอดเดิม > ยอดใบขาย (NetDebt + มัดจำ [+VAT ถ้าแยกนอก]) / BillBalance ยกมา / ลูกหนี้อื่น / มัดจำ | :1110-1116, :6452-6525 |
| INCORRECT_CREDITNOTE2 / INCORRECT_CREDITNOTE | พิมพ์ผลต่าง > BillBalance (shadow11) / > ยอดเดิม − ลดหนี้ไปแล้ว / > เพดานจากสินค้า (shadow14) | :4727-4759 |
| NUM_ITEM_NOTMORE | จำนวนคืน > จำนวนคืนได้คงเหลือ | :3507 |
| INCORRECT_CREDIT_PRICE | ประเภทราคาผิด แต่ราคาใหม่ > ราคาเดิม | :3523 |
| BILL_DOCNO / OLD_AMOUNT / DIFF_AMOUNT_EVEN | เตือนยอดที่ถูกต้อง (ข้ามสำหรับประเภท 2) | :4778-4819 |
| INV_BAL_PLS_INC_VAT | เอกสารยกมาใช้ได้เฉพาะภาษีรวมใน | :5636, :5798 |
| SAVEDETAIL_NOTCOMPLETE | บันทึกบรรทัดไม่ครบ → Abort | OnSaveData |

#### ข้อสังเกต / โค้ดที่น่าสงสัยใน Champ (อย่าลอกตาม)
- `OnGetCNAmount` (:5616-5630) ตั้งชื่อคอลัมน์ใน SQL เป็น `sumcreditnote` แต่อ่านค่าจาก `CNAmount` → น่าจะได้ 0 เสมอ (**ยังไม่ verify** runtime)
- **ตอนยกเลิกเอกสาร**
  - `BillCancelUpdateData` ตั้ง `BCInvCreditNote.IsCancel=1` ก่อน (:585-586) แล้วค่อยเรียก `CreditNoteProcessBill(doc,-1)` (:598)
  - แต่ `CreditNoteProcessBill` เลือกเฉพาะแถว `iscancel = 0` (bcprocess.cpp:2016)
  - ผลคือยอดใบขายเดิมที่ถูกลดไปในกรณี ReturnStatus=1 **น่าจะไม่ถูกคืน** ยกเว้นมีการ reprocess ภายหลัง (**ยังไม่ verify**)
  - ส่วนลูกหนี้ใช้ -2 ซึ่งจัดการถูกต้อง (bcprocess.cpp:2135-2142)
- `BillCalcInv` คำนวณยอดภาษีแม้ไม่มีเลขที่ใบกำกับ — ยอดก่อนภาษีจะเท่ากับ SumTax จนกว่า `CalcTaxAmount` จะคำนวณใหม่ด้วยอัตรา 0 (**ยังไม่ verify** ค่าที่บันทึกจริง)

---

### [5331] บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)

- **ตำแหน่งเมนู**
  - Bill (ระบบขาย) › ต่อจาก 5330
  - handler `CMainFrame::OnBillsysTrans8` (MainFrm.cpp:2451-2459; map :404; comment ในโค้ดเขียนว่า "บันทึกเพิ่มลูกค้า") → `CreateOrActivateFrame(theApp.m_pBILLFrmArDeposit, RUNTIME_CLASS(CBILLFrmArDeposit))`
  - template ที่ BC5Account.cpp:1608
  - form class `CBILLFrmArDeposit` (billfrmardeposit.cpp, 5,364 บรรทัด — ชื่อคลาสมีคำว่า "Deposit" แต่จริง ๆ คือใบเพิ่มหนี้), IDD_BILL_FRM_AR_DEPOSIT 461×291 (BC5Account.rc:7979)
  - คีย์ชื่อจอ "30009" (comment: "บันทึกใบเพิ่มลูกหนี้" :39), `m_strBillCode="DBN"` (:42), config "InvDebitNote1Config" (:44), GL source พิมพ์ slip = 8 (:59), สิทธิ์ `ID_BILLSYS_TRANS_8` (:74)
- **Variants**: `EXPRESS_PROJ` มีเมนู Key / F8 / F9 แบบย่อ (:3453-3504, OnMore :3939-4118) — นอกนั้นไม่มี
- **วัตถุประสงค์**: ออกใบเพิ่มหนี้ให้ลูกค้า (ส่งสินค้าเพิ่มหรือคิดราคาเพิ่ม) อ้างใบขายเดิม เพิ่มยอดลูกหนี้ เพิ่มภาษีขาย ตัดสต๊อกออกเมื่อเป็นการส่งสินค้าเพิ่ม และลงบัญชี GL

#### ประเภทเอกสาร (CauseType)

| ค่า | key ป้าย (comment โค้ด) | สต๊อก | DiscQty | รูปแบบ GL | คอลัมน์ที่แก้ได้ (ChmMode) | หลักฐาน |
|---|---|---|---|---|---|---|
| 0 | `DEBT_TYPE8` "เพิ่มหนี้(ลูกหนี้)" | `StockProcess=0` → **ตัดออกจากคลัง** (trans type 6, nMark −1, ใช้ DiscQty) | = จำนวน | กลุ่ม 13 | จำนวน — แก้ได้เฉพาะผู้มีสิทธิ์ `bCanChangePrice`; ราคา ส่วนลด จำนวนเงิน ล็อก | :299, :1158-1170, :4118-4128, :4692-4699; bcdumpprocjob.cpp:2638, :4252-4305 |
| 1 | `DEBT_TYPE2` "ราคาผิด" | `StockProcess=1` ส่งเข้า proc (`StockProcess<2`) แต่ DiscQty=0 และ costing เลือกเฉพาะ StockProcess=0 → ไม่มีผลต่อสต๊อก | 0 | กลุ่ม 49 | ราคา (ถ้ามีสิทธิ์) + ส่วนลด; จำนวนล็อก | :300, :1165-1166, :4701-4708; bcdumpprocjob.cpp:1333 |
| (2) | `DEBT_TYPE3` "สินค้าแตกเสียหาย" | ถูก comment ออก (:301) แต่โค้ดคุมคอลัมน์ยังมีกิ่ง 0 หรือ 2 | — | — | — | :301, :4692 |

- ค่าเริ่มต้น = 0; config `oomdts`="1" บังคับ = 1 และปิดไม่ให้เปลี่ยน (:305-312, :1367-1378)
- comment ในโค้ดของ `GetGLFormat` สลับกับความหมายจริง (ประเภท 1 เขียนว่า "debit note" ส่วนประเภท 0 เขียนว่า "wrong price") — **ให้ยึดตามค่าจริง**: ประเภท 0 → กลุ่ม 13, ประเภท 1 → กลุ่ม 49 (:4120-4127)
- ไม่มี checkbox แบบใบลดหนี้ เพราะใบเพิ่มหนี้ **เพิ่มยอดลูกหนี้เสมอ** และมียอดคงค้างของตัวเอง (`BillBalance=NetDebt` :1008-1010) ไว้ตัดในใบวางบิล/ใบเสร็จเป็น BillType 3 (bcprocess.cpp:1013-1014, :1042) — ไม่แตะยอดคงค้างของใบขายเดิม (ไม่มี ProcessBill สำหรับใบเพิ่มหนี้)

#### หน้าจอและช่องหัวเอกสาร (rc:7979-8030)

| # | ป้ายไทย (verbatim) | control → คอลัมน์ `BCDebitNote1` | ชนิด | บังคับ? | ค่าเริ่มต้น | แหล่ง lookup | การตรวจ/คำนวณ | หลักฐาน |
|---|---|---|---|---|---|---|---|---|
| 1 | เลขที่ใบเพิ่มหนี้: (`CREDITNOTE_DOCNO`) | IDC_ED_DOCNO → `DocNo` | text ตัวใหญ่ + ปุ่มรันเลข | ใช่ | รันอัตโนมัติ `SetOption(BCDebitNote1, DocNo)` | — | >20 → DOCNO_MORE_LENG แล้วตัด; ตรวจซ้ำ; ไม่มีสิทธิ์ `bCanChangeBillNo` = ปิดช่อง | :315-317, :4582-4608, :4664-4677 |
| 2 | วันที่ออกเอกสาร: (`DOCNO_DATE`) | IDC_ED_DOCDATE → `DocDate` | date พ.ศ. auto | ใช่ | WorkingDate | — | `IsPeriodClose`; ถ้าย้อนหลังกว่าวันนี้และนับสต๊อก → "ระบบกำลังทำการเช็คสต็อกย้อนหลัง เนื่องจากมีการใส่วันที่ย้อนหลัง กรุณารอสักครู่" แล้ว `OnCheckStockReword` (ข้อความไทยฝังในโค้ด) | :340-344, :4610-4662 |
| 3 | เลขที่ใบกำกับ: (`TAXNO`) | IDC_ED_TAXNO → `TaxNo` | text + ปุ่มรันเลข | ไม่ | **= เลขที่เอกสาร** (`OnBeginNewRecord` เมื่อไม่มีกลุ่มเลขเอกสาร) | — | >20 → TAXNO_MORE_LENG แล้วตัด; ถ้ามีค่า → วันที่ใบกำกับ = วันที่เอกสาร; คำนวณภาษีใหม่ | :4586-4589, :4199-4222 |
| 4 | วันที่ใบกำกับ: (`TAXDATE`) | IDC_ED_TAXDATE → **`BCDebitNote1` ไม่มีคอลัมน์ TaxDate** (DDL) จึงเก็บใน `BCOutputTax.TaxDate` | date | ต้องมีถ้ามี TaxNo (CHECK_TAXDATE) | WorkingDate | — | อัปเดตแถวภาษีแรก | :650-658, :4224-4236 |
| 5 | ประเภทภาษี: (`TAX_TYPE`) | IDC_ED_TAXTYPE → `TaxType` | combo 0/1/2 (comment :4780 "0 = แยกนอก, 1 = รวมใน"; :291 "ภาษีอัตราศูนย์") | ใช่ | `DefaultTaxType` → ทับด้วย BCAR / ใบขายแถวแรก | — | `OnSelchangeTaxType` (:4993-5030) | :287-295 |
| 6 | รหัสแผนก: (`DEPART_CODE`) | IDC_ED_DEPART → `DepartCode` | combo-table | ไม่ | ว่าง; ถ้าเอกสารสร้างโดย "OOMDTS" → ล็อก | `BCDepartment` | — | :331, :1462-1467 |
| 7 | รหัสลูกค้า: (`ARCODE`) + ชื่อ (read-only) | IDC_ED_ARCODE → `ArCode` | lookup | ใช่ | — | `BCAR` (ArTransConfig) | ดึงชื่อ, TaxType, อัตราภาษี, SaleCode, **BillCredit → เครดิต(วัน) และวันครบกำหนด** (โหมดแก้: ทับเฉพาะช่องที่ว่าง) | :3306-3348, :3546-3624 |
| 8 | เครดิต(วัน): (`CREDITDAY`) + "วัน" | IDC_ED_CREDITDAY → `CreditDay` | number | ไม่ | `BCAR.BillCredit` | — | ออกจากช่อง → วันครบกำหนด = วันที่เอกสาร + เครดิต | :4277-4286 |
| 9 | ลงวันที่: (`DATED`) | IDC_ED_DUEDATE → `DueDate` | date | ไม่ | วันที่เอกสาร + เครดิต | — | ออกจากช่อง → เครดิต = ผลต่างวัน − 1; Enter → ไปตารางใบขาย | :4288-4303, :3373-3377 |
| 10 | ประเภทเพิ่มหนี้(ลูกหนี้): (`AR_DEDITNOTE_TYPE`) | IDC_ED_DEBITTYPE → `CauseType` | combo 0/1 | ใช่ | 0 | — | ตั้ง shadow25 ของทุกบรรทัด เปลี่ยนคอลัมน์ที่แก้ได้ และโหลดรูปแบบ GL ใหม่ | :297-312, :4679-4737 |

**F5 "รายละเอียดส่วนอื่นๆ"** — `CBILLDlgArDeposit` (rc:8599; billdlgardeposit.cpp) — ลำดับบนจอ:

| ซ้าย | คอลัมน์ | ขวา | คอลัมน์ |
|---|---|---|---|
| ชื่อในการออกบิล: | `ARNAME` | — | — |
| ที่อยู่ในการออกบิล: | `ARADDRESS` | — | — |
| สาเหตุการเพิ่มหนี้: | `CauseCode` (`BCCNMaster`; ChmMode ค่าเริ่มต้น "04"/"05" = CauseType+4, billdlgardeposit.cpp:143) | — | — |
| โครงการ: | `ProjectCode` | มูลค่าส่วนลด: | `DiscountAmount` (ปิด) |
| การจัดสรร: | `AllocateCode` | ยอดก่อนภาษี: | `SumOfBeforeTax` |
| รหัสพนักงานขาย: | `SaleCode` (+`UserGroup`) | ยอดภาษี: | `SumOfTaxAmount` |
| กลุ่มของบิล: | `BillGroup` | มูลค่ารวมภาษี: | `SumOfTotalTax` |
| อัตราภาษีมูลค่าเพิ่ม: | `TaxRate` | ยอดยกเว้นภาษี: | `SumOfExceptTax` |
| ยอดใบกำกับเดิม: | `SumOldAmount` | มูลค่าเพิ่มหนี้: | `NetDebtAmount` |
| รวมมูลค่าสินค้า: | `SumOfItemAmount` | ยอดที่ถูกต้อง: | `SumTrueAmount` |
| หมายเหตุ: | `MyDescription` | | |
| สกุลเงิน: | `CurrencyCode` | อัตราแลกเปลี่ยน: | `ExchangeRate` |

- ปุ่ม F11=ตกลง / Esc=ยกเลิก
- ChmMode ปิดช่องยอดสรุป (billdlgardeposit.cpp:107-125); อัตราภาษีและยอดภาษีแก้ได้ตามสิทธิ์ (:166-167)

#### ตารางรายการ (grid) — ตารางบน "ใบขายที่อ้างอิง" (:173-196)

| # | ป้าย (key; comment) | กว้าง | แก้ได้? | ความหมาย/สูตร | หลักฐาน |
|---|---|---|---|---|---|
| 0 | `INVOICE_NO` "เลขที่ใบ invoice" | 140 | แก้ได้ + ค้นหา | เลขที่ใบขาย (เฉพาะ bcarinvoice; BillType=0 เสมอ) | :173, :4817 |
| 1 | `ODL_DISCOUNT` "ส่วนลดเดิมของ invoice" | 120 | **ไม่** | — | :174 |
| 2 | `OLD_AMOUNT` "ยอดเดิม" | 120 | แก้ได้ | ChmMode = NetDebt ของใบขาย | :175, :4792-4811 |
| 3 | `UPDATE_AMOUNT` "ยอดผลต่าง" | 120 | แก้ได้ | ยอดเพิ่มหนี้รวม VAT (คำนวณจากสินค้าใน BillCalc) | :176, :3244-3304 |
| 4 | `TRUE_AMOUNT` "ยอดที่ถูกต้อง" | 120 | ไม่ | = ยอดเดิม **+** ผลต่าง (− ศูนย์/ยกเว้นใน BillCalc) | :177, :3223-3242, :3292 |

- Shadow ของตารางบน (:183-195):

  | # | ชื่อ | # | ชื่อ |
  |---|---|---|---|
  | 0 | RowOrder | 6 | BillType |
  | 1 | InvExchangeRate | 7 | SumOfItemAmount |
  | 2 | NewExchangeRate | 8 | DiscountAmount |
  | 3 | HomeAmount1 | 9 | ZeroTaxAmount |
  | 4 | HomeAmount2 | 10 | ExceptTaxAmount |
  | 5 | ExchangeProfit | | |

  ไม่มีคอลัมน์ BillBalance / CNAmount / เพดาน เหมือนใบลดหนี้
- **การเลือกใบขาย** (`OnListFindDataInv` :3626-3769, config `BillingTransConfig` เลือกได้หลายใบ)
  - ต้องกรอกรหัสลูกค้าก่อน
  - เงื่อนไขกรอง: ลูกค้านี้, ยังไม่อยู่ในตาราง, IsCancel=0, IsCreditNote=0, PosStatus=0, `(BillBalance>0 OR BillType 0 OR BillType 2)`, TaxType ตรงกับจอ, มี/ไม่มี TaxNo ตรงกับจอ
  - ใบแรกกำหนดสกุลเงิน อัตราแลกเปลี่ยน อัตราภาษี และชื่อ/ที่อยู่ออกบิล
- **พิมพ์เลขที่เอง** (:3771-3887)
  - `OnCheckArInvoice` ต้องพบใบใน BCArInvoice (NOT_FOUND_INVOICENO :4746)
  - เลขซ้ำในตาราง → DOCNO_DUPLICATE
  - `OnGetArInvoice` (:4755-4827) เติม ยอดเดิม = NetDebt, ผลต่าง = ว่าง, ยอดที่ถูกต้อง = NetDebt (ขายเชื่อใช้ BillBalance)
- **ไม่มี guard ยอดเพิ่มหนี้สูงสุด และไม่มีเพดานจำนวน** (ต่างจากใบลดหนี้)

#### ตารางรายการ (grid) — ตารางล่าง "สินค้า" (:213-225)

| # | ป้าย (key) | กว้าง | แก้ได้? (ChmMode ประเภท 0 / 1) | ความหมาย | หลักฐาน |
|---|---|---|---|---|---|
| 0 | `ITEMCODE` รหัสสินค้า | 60 | แก้ได้ + ค้นหา | ค้นจาก **สินค้าหลัก** (คลิก = `ItemTransConfig`, Shift = `CMyFilterItemDlg`) ไม่ได้ค้นจากบรรทัดใบขาย; ไม่พบ → NOT_SEEITEMCODE | :2897-2995, :2781-2830 |
| 1 | `ITEMNAME` ชื่อสินค้า | 120 | ตาม `IsEditItemName` | — | :214 |
| 2 | `INVOICE_NO` "เอกสารอ้างอิงใบส่งของ" | 70 | แก้ได้ | ใบขายที่บรรทัดนี้อ้าง | :215 |
| 3 | `WAREHOUSE_ITEM` คลังสินค้า | 60 | combo | คลังที่ตัดออก (ค่าเริ่มต้น `BCItem.DefSaleWHCode`) | :216, :2997-3143 |
| 4 | `SHELF` ที่เก็บ | 60 | combo | `DefSaleShelf` | :217 |
| 5 | `UNITCODE` หน่วยนับ | 60 | combo | `DefSaleUnitCode`; ตรวจ `IsUnitOfItem` (NOUNIT) | :218, :2556-2573 |
| 6 | `ITEMQTY` (อนุมาน: จำนวน) | 60 | ประเภท 0: แก้ได้เฉพาะ `bCanChangePrice` / ประเภท 1: ล็อก | ถ้าห้ามสต๊อกติดลบ (`StockStatus==2`) → `IsStockQtyEnough` ไม่พอ → STOCK_NOT_ENOUGH ถามว่าจะดูคลังอื่นไหม (`CMyDlgStockDisplay`) | :219, :2737-2762 |
| 7 | `PERPRICE` (อนุมาน: ราคา) | 60 | ประเภท 0: ล็อก / ประเภท 1: แก้ได้ถ้ามีสิทธิ์ | ราคาว่าง → `GetItemPrice` (ราคาขายตามลูกค้า วันที่ จำนวน) | :220, :2584-2608 |
| 8 | `DISCOUNT` ส่วนลด | 70 | ประเภท 0: ล็อก / ประเภท 1: แก้ได้ | — | :221-224 |
| 9 | `AMOUNT` จำนวนเงิน | 70 | ล็อก | จำนวน×ราคา − ส่วนลด | :225, :1347-1355 |

**Shadow ของตารางล่าง** (:231-267)

| # | ชื่อ | # | ชื่อ | # | ชื่อ |
|---|---|---|---|---|---|
| 0 | RowOrder | 12 | TypeCode | 24 | ExchangeRate |
| 1 | ItemType | 13 | FormatCode | 25 | StockProcess |
| 2 | MyDescription | 14 | OldUnitCode | 26 | Lot Script |
| 3 | ProjectCode | 15 | ShelfCode | 27 | DiscountAmount |
| 4 | AllocateCode | 16 | NetAmount | 28 | LineNumber |
| 5 | OppositeQty | 17 | HomeAmount | 29 | RefLineNumber |
| 6 | OppositeUnit | 18 | BillQty | 30 | LotNumber |
| 7 | ExceptTax | 19 | RefDocNo | 31 | ColorCode |
| 8 | BarCode | 20 | OldQty | 32 | SizeCode |
| 9 | CategoryCode | 21 | SetCode | 33 | StyleCode |
| 10 | GroupCode | 22 | DepartCode | 34 | BehindIndex |
| 11 | BrandCode | 23 | CurrencyCode | 35 | OldItemCode |

- **F9 ดึงรายการ** (`LoadInvoiceSub` :4305-4438)
  - ดึงบรรทัดของใบขายที่เลือกในตารางบน ที่ `CNQty>0`, IsCancel=0, TransState=0, เป็นของลูกค้านี้ และยังไม่อยู่ในตาราง
  - จำนวน = CNQty ที่เหลือ; ยอดเงินคำนวณใหม่เมื่อจำนวนไม่เท่ากับจำนวนเดิม
  - คัดลอกคลัง/ที่เก็บ/หน่วย/ราคา/ส่วนลด และ shadow (โครงการ, การจัดสรร, ยกเว้นภาษี, หมวด…, RefLineNumber)

#### ยอดรวมท้ายเอกสาร / สูตรคำนวณ
1. `DetailCalc` (:3158-3221): สูตรเดียวกับใบลดหนี้ (DiscountMethod, ปัดเศษ, NetAmount ตัด VAT กรณีรวมใน+มีใบกำกับ, HomeAmount)
2. `BillCalc` (:3244-3304)
   - ต่อใบขาย: รวมยอดสินค้าที่อ้างใบนั้น แยกกองยกเว้น / ศูนย์ / มีภาษี แล้วหักส่วนลดเดิม
   - แยกนอก + มีใบกำกับ → ผลต่าง ×(1+อัตรา/100)
   - **ยอดที่ถูกต้อง = ยอดเดิม + ผลต่าง − (ศูนย์+ยกเว้น)**
3. `BillCalcInv` (:4130-4197)

   | ยอด | สูตร |
   |---|---|
   | SumTaxAmount | Σ ผลต่าง |
   | NetDebtAmount | SumTax + ยกเว้น + ศูนย์ |
   | SumTrueAmount | ยอดเดิม + NetDebt |
   | ยอดภาษี | SumTax×อัตรา/(100+อัตรา) ปัดเศษ |
   | ยอดก่อนภาษี | SumTax − ภาษี (ไม่มี TaxNo = SumTax) |

4. `CalcTaxAmount` (:4238-4275)
   - ส่งเข้า `BillTaxProcess` แบบรวมใน (ถ้า TaxType≠2)
   - ฐาน = NetDebt ถ้าไม่มีรายการสินค้า; ไม่งั้น = ยอดสินค้า − ส่วนลด
   - NetDebt = ยอดรวม + ยกเว้น − WHT + ศูนย์
5. แถวภาษี `MyUpdateFirstTaxRow` (:4856-4892): คำอธิบาย key `DEBT_TYPE8`; ภาษีขายบันทึกแบบ **บวกภาษี** (Process 1)

#### ปุ่ม แถบเครื่องมือ คีย์ลัด (`PreTranslateMessage` :3366-3544; `OnMore` :3939-4118)

| คีย์ | ผล |
|---|---|
| Enter ที่ "ลงวันที่" | ไปตารางใบขาย |
| PageUp / PageDown | ตารางใบขาย / ตารางสินค้า |
| F2 / Shift+F2 / Ctrl+F2 | ลงบัญชี GL / รายละเอียดบรรทัด (`CBILLDlgExpandArDebitNote` "Other Goods Detail" rc:8245) / ตรวจสต๊อกย้อนหลัง (`OnBttnCheckStockReword`) |
| F3 | ภาษีขาย (`OnOutputTax("InvDebitNote1Config")`) แล้วคัดลอก TaxNo/TaxDate กลับมาที่จอ |
| F5 / Shift+F5 / Ctrl+F5 | รายละเอียดส่วนอื่นๆ / ยืนยัน (`bCanConfirmBill`) / ยกเลิก (`OnConfirmCancel` → `BillCancel("BCDebitNote1")`) |
| F6 | ไม่มีหน้าจอภาษีหัก ณ ที่จ่าย (Ctrl+Shift+F6 เปิดไฟล์ debug `c:\BCDebitNote1_BatchNo.txt`) — แต่ตอนบันทึกยังเรียก `OnSaveARWTax(4,…)` |
| F7 | ไม่มี (Ctrl+Shift+F7 = ยูทิลิตี้คำนวณยอดหัวเอกสารใหม่) |
| Shift+F8 | หน่วยนับขนาน (`OnOppositQty`) |
| F9 / Ctrl+F9 | ดึงรายการจากใบขาย / Serial |
| F11 | บันทึก |
| F12 / Ctrl+F12 | เส้นทางเอกสาร (ใบวางบิล/ใบเสร็จ/ใบเสร็จชั่วคราว ที่อ้าง BillType 3 :4829-4854) / ข้อมูลสินค้า |

**dialog รายละเอียดบรรทัด** (`CBILLDlgExpandArDebitNote`) มีช่อง:
- ผลต่อสต๊อก: 3 ค่า `UPDATE_QTY_COST` "ปรับปรุงจำนวนสินค้า และต้นทุน" / `UPDATE_COST` "ปรับปรุงต้นทุนสินค้า" / `NOT_UPDATE` "ไม่ปรับปรุง" (ป้ายจาก comment) — ChmMode ปิดไม่ให้แก้ (BILLDlgExpandArDebitNote.cpp:83-85, :130-131)
- แผนก, โครงการ, การจัดสรร, คำอธิบายรายการ, เอกสารอ้างอิง, จำนวนเดิม, ยอดเงินสุทธิ, ยอดเงินบาทสุทธิ
- คุณสมบัติสินค้า (rc:8245)

#### สถานะเอกสารและ workflow

| ขั้น | เงื่อนไข / ผล | หลักฐาน |
|---|---|---|
| บันทึก | `IsCompleteSave` 0→1, Commit, `processdoc` DocType=6; แก้ไข → กลับรายการลูกหนี้เดิมก่อน (`DebitNote1ARProcess(old,-1)` :950) | :944-1317 |
| แก้ไข | ห้ามเมื่อ ผ่าน GL / ยกเลิกแล้ว / ยืนยันแล้ว / วางบิลแล้ว (CANNOT_EDIT_BILL1) / ถูกรับชำระแล้ว (NetDebt≠BillBalance → NOT_EDIT_BILLBALANCE) | :662-690 |
| ยืนยัน | Shift+F5 | :3418-3424 |
| ยกเลิก | ห้ามเมื่อ ผ่าน GL (CANNOT_CANCEL) / ยกเลิกแล้ว (MSG_CANNOT_CANCEL) / ยืนยันแล้ว / วางบิลแล้ว / ถูกใช้ใน `BCPaybillSub`, `BCReceiptSub1`, `BCReceiptSub2` BillType 3 (CANNOT_CANCEL_USE) | :5240-5283 |
| ลบ | ห้ามเมื่อ ไม่มีเลขที่ / ปิดงวดแล้ว / ผ่าน GL / ยกเลิกแล้ว / ยืนยันแล้ว / วางบิลแล้ว / ถูกรับชำระแล้ว → CONFIRM_DELETEDATA | :431-479 |
| ถูกอ้างต่อ | ตัดในใบวางบิล/ใบเสร็จเป็น BillType 3 → ลด `BCDebitNote1.BillBalance` / `PayBillAmount` | bcprocess.cpp:666-707, :1013-1042 |

#### ผลข้างเคียง

**`BCDebitNote1` — หัวเอกสาร** (:955-1040)

| กลุ่ม | คอลัมน์ |
|---|---|
| ข้อมูลเอกสาร | TaxType, CauseType, DocNo, DocDate, TaxNo, ArCode, CreditDay, DueDate, DepartCode, RecurName, UserGroup, SaleCode, TaxRate, MyDescription, CurrencyCode, ExchangeRate, AllocateCode, ProjectCode, BillGroup, CauseCode, ARName, ARAddress, SCGBPNumber |
| ยอดเงิน | DiscountWord, DiscountAmount, SumOldAmount, SumTrueAmount, SumOfBeforeTax, SumOfTaxAmount, SumOfTotalTax, SumOfExceptTax, NetDebtAmount, SumOfZeroTax, SumOfItemAmount |
| ยอดคงค้าง | PayBillAmount = BillTemporary = BillBalance = NetDebt |
| การลงบัญชี | GLFormat, GLTransData, GLStartPosting=0, IsCompleteSave |
| ผู้บันทึก | Creator/LastEditor |

ตามรายการคอลัมน์ใน DDL ตารางนี้ **ไม่มี TaxDate**

**ตารางอื่นที่เขียน**

| ตาราง | สิ่งที่เขียน / ผล | หลักฐาน |
|---|---|---|
| `BCInvDebitNote1` (ลบแล้วเขียนใหม่) | DebitNoteNo, DocDate, ArCode, **ReturnStatus=0 คงที่**, InvoiceNo, InvoiceDisc, OldAmount, DiffAmount, TrueAmount, InvExchangeRate, NewExchangeRate, HomeAmount1, HomeAmount2, ExchangeProfit, SumOfItemAmount, DiscountAmount, ZeroTaxAmount, ExceptTaxAmount | :1049-1103 |
| `BCDebitNoteSub1` (ลบแล้วเขียนใหม่) | DocNo, TaxNo, TaxType, CauseType, DocDate, ArCode, UserGroup, SaleCode, ItemCode, ItemName, InvoiceNo, WHCode, ShelfCode, UnitCode, Price, DiscountWord, Amount, DiscQty (เฉพาะประเภท 0), TempQty, StockProcess (=CauseType), DiscountAmount, ItemType, MyDescription, ProjectCode, AllocateCode, OppositeQty/Unit, ExceptTax, BarCode, หมวด…, NetAmount, HomeAmount, BillQty, DepartCode, CurrencyCode, ExChangeRate, RefLineNumber, LotNumber, Color/Size/Style, TaxRate, PackingRate1/2, behindindex, IsProcess, LineNumber (+SCG fields ถ้าเปิดใช้) | :1108-1256 |
| Lot | ถ้าระบุ Lot เอง → `InsertListLotOut` (ตัดออก) | :1246-1253 |
| `processdoc` | DocType=6 | :1274 |
| ภาษีขาย `BCOutputTax` | `OnSaveOutputTax(4, old, doc, date, ar, Process=1 บวกภาษี, Source=8, BillGroup, multicurrency)` | :1286 |
| WHT | `OnSaveARWTax(4,…)` | :1288 |
| GL | อัตโนมัติ `m_GLTLSProcess` Source 8 / GLGroup 8 หรือคีย์เอง `OnSaveAccount(…,8,…)`; รูปแบบ GL กลุ่ม 13 / 49 — ไม่มีรหัสบัญชีในฟอร์ม | :1291-1301, :4118-4128 |
| ลูกหนี้ | `DebitNote1ARProcess(doc)` → `ProcessAR(ar, งวด, NetDebt, …, mark=+1)` **เพิ่มหนี้** + `UpdateARPeriodState(… คอลัมน์ DN …)` | :1302; bcprocess.cpp:2349-2400 |
| สต๊อก | `OnStockReprocess` ต่อสินค้า + proc `DebitNote1StockProcess` (ตัดออก DiscQty ที่ StockProcess<2); costing นับ HomeAmount เฉพาะ StockProcess=0 และ DiscQty>0 | :1266-1270; bcdumpprocjob.cpp:1333, :4252-4305 |

- ยกเลิก (:481-525) ทำสิ่งต่อไปนี้:
  - `BCDebitNote1` BillBalance=0, IsCancel=1; Sub และ Inv IsCancel=1
  - reprocess สต๊อก
  - `DebitNote1ARProcess(doc,-1)`
  - `CancelOutputTax(4,doc,8)`
  - `ReverseAccount(doc,8)`
  - ลบ Serial และ Lot
  - log CANCEL_DEBITNOTE
- ลบ (:527-592) ทำสิ่งต่อไปนี้:
  - `DebitNote1ARProcess(doc,-1)`
  - DELETE ตารางหัว, Inv และ Sub
  - reprocess สต๊อก
  - `DeleteOutputTax(4,doc,8)`
  - `DeleteAccount(doc,8)`
  - ลบ Serial และ Lot
  - log SUBTRACT_DEBITNOTE

#### Guard / ข้อความเตือนสำคัญ

| key | เงื่อนไข | หลักฐาน |
|---|---|---|
| CANNOT_SAVE_NOTDOCNO / DOCNO_DUPLICATE / "Document No is duplicated…" | ไม่มีเลขที่ / เลขซ้ำ | :627-640, :862-909 |
| CANNOT_SAVE_NOTARCODE | ไม่มีรหัสลูกค้า | :642-647 |
| CHECK_TAXDATE | มีเลขที่ใบกำกับแต่ไม่มีวันที่ | :650-658 |
| CANNOT_SAVE_POSTGL / NOT_SAVE_ISCANCEL / NOT_SAVE_ISCONFIRM / CANNOT_EDIT_BILL1 / NOT_EDIT_BILLBALANCE | ใช้ตอนแก้ไข | :662-690 |
| NOT_SET_WHCODE / NOT_SET_SHELFCODE / NOT_SET_UNITCODE | สินค้าที่ไม่ใช่บริการขาดคลัง/ที่เก็บ/หน่วย | :696-730 |
| NOT_SPECIFY_SERIAL / SPECIFY_NOTEQUAL_SERIAL | สินค้า Serial | :731-753 |
| NOT_SET_LOT / STOCK_NOT_ENOUGHLIST (Lot ไม่พอ) | ระบุ Lot เอง + ต้นทุน FIFO | :758-802 |
| SUMAMOUNT…STOCK_NOT_ENOUGHLIST | ห้ามสต๊อกติดลบ + ประเภท 0: ยอดรวมต่อสินค้า×คลัง (หักจำนวนเดิม) > คงเหลือ ณ วันที่เอกสาร | :805-853 |
| BUTTON_HOME | เปิดเอกสารที่ไม่ใช่วันนี้และประเภท≠1 (ความหมายอนุมาน: เตือนสต๊อกย้อนหลัง) | :1800-1802 |

#### ข้อสังเกต / โค้ดที่น่าสงสัยใน Champ (อย่าลอกตาม)
- **ยกเลิกแล้วลูกหนี้น่าจะไม่ถูกกลับรายการ**
  - `CMyDlgBillCancel` ตั้ง IsCancel=1 ก่อน (MyDlgBillCancel.cpp:66) และ `BillCancelUpdateData` ตั้งซ้ำ (:488)
  - จากนั้นเรียก `DebitNote1ARProcess(doc,-1)` (:505) ซึ่งเลือกเฉพาะ `IsCancel=0` เมื่อ mark≠-2 (bcprocess.cpp:2357-2365)
  - ส่วนใบลดหนี้ส่ง -2 ซึ่งถูกต้อง — **ยังไม่ verify** runtime
- เลือกใบขายผ่าน browse แล้วได้ ผลต่าง = ยอดเต็ม และยอดที่ถูกต้อง = ยอดเดิม + ยอดเดิม (:3729-3740) ซึ่งไม่ตรงกับตอนพิมพ์เลขที่เอง (:4792-4811) — ตัวเลขจะถูกต้องหลังดึงสินค้าแล้ว BillCalc คำนวณใหม่
- เพิ่มหนี้ได้ไม่จำกัด: ไม่มีเพดานยอดหรือจำนวนเทียบใบเดิม

---

---

## ส่วน 6 — รายงานขายชุดรายวัน (Report Center)

> ขอบเขต: 24 รายการเมนูใน popup `จัดขาย` > `รายงานข้อมูลรายวัน` (menu id 5344, 5333, 5345, 5354, 5363, 5364, 5346, 5334, 5335, 5355, 5341, 5337, 5336, 5347, 5349, 5340, 5338, 5339, 5350, 5342, 5343, 5302, 5686, 5362) — ที่ตั้ง `D:\project-champ\champ\champ\menuconfigxml\menuconfig.xml:671-704` (ไฟล์ UTF-16LE ; ตรวจโดยถอดรหัสแล้วนับบรรทัด).
> ไฟล์ต้นทางย่อ: `MainFrm.cpp` / `BC5Account.cpp` = `D:\project-champ\champ\champ\` ; `BLRep*.cpp`, `BILLRep*.cpp`, `BCRepBilling.rc` = `D:\project-champ\champ-report\BC5REP_BILL\` ; `BaseReportView.cpp`, `BaseReportDialog.cpp` = ชุด lib-report ; `companyinfo.cpp` = ชุด lib-frmwrk. เลขบรรทัดทุกตัวเลขมาจาก `grep -n` / `sed -n` ของไฟล์ที่ถอดรหัสแล้ว (ไฟล์ต้นทางเป็นรหัสไทยเดิม).
> หลักการเขียน: คอลัมน์/ป้ายที่ key ภาษาอยู่ในฐานข้อมูล `myresource` (ไม่อยู่ในรีโป) ใช้ไทยจาก comment ท้ายบรรทัดโค้ด ป้ายใน `.rc` หรือชื่อเมนู ; ถ้าไม่มีที่มา เขียนว่า "(อนุมาน)" ; สิ่งที่ตรวจไม่ได้เขียนว่า "ยังไม่ verify".

### รายการ 24 รายงาน (ตามลำดับในเมนู)

| ลำดับ | menu id | ชื่อ (ตาม menuconfig.xml) | บรรทัดเมนู | resourceid | หัวข้อในไฟล์นี้ |
|---|---|---|---|---|---|
| 1 | 5344 | รายงานใบเสนอราคา | :673 | 34002 | 1 |
| 2 | 5333 | รายงานการบันทึกใบเสนอราคาพร้อมรายการย่อย | :674 | 34001 | 1 |
| 3 | 5345 | รายงานสถานะใบเสนอราคา | :675 | 34020 | 1 |
| 4 | 5354 | รายงานการยกเลิกใบเสนอราคาสินค้าพร้อมรายการย่อย | :676 | 34003 | 1 |
| 5 | 5363 | รายงานการใบสั่งขาย/จองสินค้าพร้อมรายการย่อย | :678 | 34007 | 2 |
| 6 | 5364 | รายงานการยกเลิกใบสั่งขาย/สั่งจองสินค้าพร้อมรายการย่อย | :679 | 34008 | 2 |
| 7 | 5346 | รายงานการบันทึกใบรับเงินมัดจำ | :681 | 34021 | 3 |
| 8 | 5334 | รายงานการตัดใบรับเงินมัดจำ | :682 | 34010 | 3 |
| 9 | 5335 | รายงานใบรับเงินล่วงหน้า | :684 | 38100 | 3 |
| 10 | 5355 | รายงานการตัดใบรับเงินล่วงหน้า | :685 | 34030 | 3 |
| 11 | 5341 | รายงานบันทึกคืนเงินรับล่วงหน้า | :686 | 300603 | 3 |
| 12 | 5337 | รายงานรายวันขายสินค้าและบริการ | :688 | 34013 | 4 |
| 13 | 5336 | รายงานบันทึกขายสินค้า,บริการพร้อมรายการย่อย | :689 | 34012 | 4 |
| 14 | 5347 | รายงานการตัดใบขายสินค้า,บริการ | :690 | 34022 | 4 |
| 15 | 5349 | รายงานบันทึกขายสินค้า,บริการตามประเภทการขาย | :691 | 34025 | 4 |
| 16 | 5340 | รายงานบันทึกใบรับคืนสินค้า,ลดหนี้ | :693 | 34016 | 5 |
| 17 | 5338 | รายงานใบรับคืนสินค้า/ลดหนี้พร้อมรายการย่อย | :694 | 34014 | 5 |
| 18 | 5339 | รายงานการตัดใบรับคืนสินค้า,ลดหนี้ | :695 | 34015 | 5 |
| 19 | 5350 | รายงานสถานะใบรับคืนสินค้า/ลดหนี้ | :696 | 34026 | 5 |
| 20 | 5342 | รายงานใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)พร้อมรายการย่อย | :698 | 34018 | 5 |
| 21 | 5343 | รายงานยอดขายสุทธิของพนักงานขาย | :700 | 34019 | 6 |
| 22 | 5302 | รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย | :701 | 35808 | 6 |
| 23 | 5686 | รายงานการเก็บเงินของพนักงานขาย | :702 | 34040 | 6 |
| 24 | 5362 | รายงานการนำส่งเงินประจำวัน | :703 | 34043 | 6 |

เส้นคั่นกลุ่มในเมนู (`id="0"`) อยู่ที่ :677, :680, :683, :687, :692, :697, :699 — ตรงกับการแบ่งหัวข้อ 1-6 ของไฟล์นี้ (5686 และ 5362 ไม่มีเส้นคั่นจาก 5343/5302 จึงรวมกลุ่ม "พนักงานขาย/เก็บเงิน/นำส่งเงิน").

### R0. กติการ่วมของรายงานกลุ่มนี้ (ใช้กับทุกรายงานด้านล่าง)

| หัวข้อ | พฤติกรรมใน Champ | หลักฐาน |
|---|---|---|
| ตำแหน่งเมนู | popup แม่ `จัดขาย` (resourceid 30000) ตั้ง `disable="yes"` > กลุ่มย่อย `รายงานข้อมูลรายวัน` (34000) > 24 รายการ ; popup `disable="yes"` ไม่ขึ้นบนแถบเมนูหลัก — ส่วน E ของเอกสารลูกหนี้ (04-ar.md) บันทึกว่าเปิดจาก Report Center ; หน้าจอ Report Center เอง ยังไม่ได้เปิดอ่านในไฟล์นี้ (ยังไม่ verify) | `menuconfig.xml:671-672,704` |
| โครงโปรแกรม | 1 รายงาน = `*View.cpp` (สืบจาก `CMyBaseReportView` ; กำหนดคอลัมน์ `DefaultData`, ประมวลผล `RunProcess`, รวมยอดหน้า `OnAddPageSummery`, ท้ายรายงาน `OnAddReportCondition`) + `*Dlg.cpp` (สืบจาก `CMyBaseReportDlg` ; กล่องเงื่อนไขและตัวสร้าง SQL) ; ลำดับ: handler ใน MainFrm เปิดกล่อง + ตั้งช่องคลิกทะลุ -> ปุ่มยืนยัน -> `OnFilterReport` -> `LoadUserDefine()` ถ้าไม่มีโครงที่ผู้ใช้บันทึกไว้ใช้ `DefaultData()` -> `RunProcess(pDlg)` -> `OnAddPageSummery` ปิดท้ายหน้า -> `OnAddReportCondition` | `BLRepBillQuotationView.cpp:11,43,114,139,158,173,272,287,292,318` ; `BLRepBillQuotationDlg.h:6` ; `MainFrm.cpp:4806-4815` |
| ตัวสร้าง SQL | แต่ละ Dlg ต่อสตริง SQL เองจากค่าที่ผู้ใช้กรอก แล้วเรียก `GetSQLScript()` ซึ่งสั่ง `MakeUpper()` กับ **ทั้งก้อน SQL** (รวมค่าที่ผู้ใช้พิมพ์) ก่อนใช้งาน ; ค่า "เงื่อนไขที่เลือก" เก็บใน `m_strCondition` คั่นด้วย `@` (ข้อความป้ายมาจาก `GetMyRes("FROM_DATE")`, `GetMyRes("FROM_DOCNO")` ฯลฯ) แล้วพิมพ์ท้ายรายงานเมื่อ `m_RepOption.bDisplayCondition` | `BaseReportDialog.cpp:254-257` ; `BLRepBillQuotationDlg.cpp:105,124-125,136,155-156` ; `BLRepBillQuotationView.cpp:318-348` |
| ช่วง "จาก–ถึง" | ถ้าช่อง "ถึง" ว่าง ใช้เงื่อนไข `= จาก` (ไม่ใช่ "จากไปจนสุด") ; ถ้าใส่ทั้งสองใช้ `BETWEEN` ; ตัวอย่างวันที่ใน `m_strCondition` มี 2 กิ่ง (มีทั้งช่วง / มีแค่ต้น) | `BLRepBillQuotationDlg.cpp:124-125` (ช่วง) กับ `:136` (ค่าเดียว) ; ซ้ำแบบเดียวกันในทุก Dlg ที่อ่านแล้ว (รายละเอียดแต่ละช่องดูตารางเงื่อนไขของแต่ละหัวข้อ) |
| ตัวเลือกจากข้อมูลบริษัท | `IsPrintCancel` (0 = ไม่พิมพ์ใบที่ยกเลิก, 1 = พิมพ์) · `IsPrintSumPage` (0 = พิมพ์รวมรายการต่อหน้า, 1 = ไม่พิมพ์) · `PosRepSeparate` (1 = รวม POS กับเอกสารปกติ ไม่กรอง `PosStatus` ; ค่าอื่น = กรองเฉพาะ `PosStatus=0` / `IsPos=0`) · `MultiCurrency` (0/1) · `AmountPoint` / `QtyPoint` / `PricePoint` (ทศนิยมของยอดเงิน/จำนวน/ราคา ค่าเริ่มต้น 2) · `DiscountMethod` · `YearType` (0 = พ.ศ.) | `companyinfo.cpp:13,22,41-42,70,81-84,90` (ค่าเริ่มต้น) ; `:173-174,199,209-213` (โหลดจากฐานข้อมูล) ; `BLRepCreditNoteStatusDlg.cpp:25-28` (comment "รายงานไม่แยกหลังร้าน" เมื่อ `PosRepSeparate == 1`) |
| สกุลเงิน | ถ้า `MultiCurrency == 1` และตัวเลือกรายงาน `bHomeCurrencyReport == 1` แปลงยอดเป็นเงินบาท (คูณ `ExchangeRate`) แล้วต่อสัญลักษณ์ "฿" ; ถ้าไม่แปลง แสดงตามสกุลของเอกสาร และบางรายงานไม่พิมพ์แถวรวม (ดูท้ายแต่ละหัวข้อ) | ตาราง "ยอดรวม" ของหัวข้อ 2 (5363/5364), 3 (5334/5335/5355/5341) |
| เอกสารที่ยกเลิก | รายงานที่พิมพ์เอกสารยกเลิกได้ (เมื่อ `IsPrintCancel == 1`) ต่อท้ายเลขที่เอกสารด้วย `(` + ข้อความของ key `ASSERT3` + `)` — เอกสารนี้เขียนย่อว่า `(ASSERT3)` ; ข้อความจริงของ key อยู่ใน `myresource` ยังไม่ verify (ชื่อ key ชี้ว่าหมายถึง "ยกเลิก" — อนุมาน) ; **การนับรวมยอดของใบยกเลิกไม่เหมือนกันทุกรายงาน** (5335 ไม่รวม, 5341 รวม — ดูหัวข้อ 3.4) | `BLRepBillQuotationView.cpp:248` (`GetMyRes("ASSERT3",FALSE)`) ; ตารางหัวข้อ 3.4 |
| โครงสร้างคอลัมน์ | `InsertColumnData(nCol, ข้อความ, ความกว้าง, ชิดซ้าย/กลาง/ขวา, ความกว้างแนวตั้ง)` เขียนที่ดัชนี `nCol` ด้วย `SetAt` — ถ้าสั่งซ้ำดัชนีเดิม ตัวหลังทับตัวก่อน ; `AddColumn(nSpace,…)` ปิดชุดคอลัมน์ 1 ระดับ ; ข้อมูลแถวใช้ `InsertRowData(ช่อง, ข้อความ)` แล้ว `AddRow(รายชื่อช่องที่คลิกได้, ระดับ(ชุดคอลัมน์), bSummary, น้ำหนักตัวอักษร, ตัวเอียง, ขีดเส้นใต้, ขีดฆ่า)` — ถ้ารายงานมีหลายระดับต้อง `AddColumn` หลายครั้ง | `BaseReportView.cpp:321-359,376,421,615` |
| คลิกทะลุ (drill-down) | `OnCallTrans(nCol,nColRow,strData)` ในคลาสฐานว่างเปล่า ; แต่ละรายงานกำหนดเอง ; MainFrm ลงทะเบียนปลายทางเป็นช่อง 0,1,2,(3) ผ่าน `REPComm_InsertCommand(ช่อง, ID)` — ปลายทางที่พบ: 5303 ใบเสนอราคา (`ID_BILLSYS_TRANS_1`) · 5326 ใบสั่งขาย/สั่งจอง · 5327 ใบรับเงินมัดจำ · 5328 ใบขาย · 5329 ใบรับเงินล่วงหน้า · 5330 ใบรับคืน/ลดหนี้ · 5331 ใบเพิ่มหนี้ · 5306 รับคืนเงินล่วงหน้า · 5127 ใบรับชำระ · 5104 ข้อมูลลูกหนี้ (`ID_ARSYS_ARMASTER`) · 5602 รายละเอียดสินค้า (`ID_ICSYS_ITEMMASTER`) · 5618 Serial Number | `BaseReportView.cpp:1993` ; `MainFrm.cpp:4806-4840,4850-4876,4890-5135,5006-5017` ; แผนที่ช่อง-ปลายทางของแต่ละรายงานอยู่ใต้หัวของหัวข้อ 1-6 |
| ตารางเงื่อนไขผู้ใช้ | ท้ายกล่องเงื่อนไขทุกใบมีกรอบ "เงื่อนไขการพิมพ์รายงานมาตรฐาน" = ตาราง 3 คอลัมน์ `FIELDS` (เลือกฟิลด์) · `OPERATE` (Between, =, >=, <=, >, <, UserDefine) · `VALUES` ; กับภาพตัวอย่างหน้ารายงาน (`InitRepExamPicture`) ; ผู้ใช้เพิ่มเงื่อนไขนอกเหนือจากช่องบนจอได้ | `BaseReportDialog.cpp:321-358` (`InitExCondition`), `:360` (`InitRepExamPicture`) |
| เก็บ/ส่งออก | `LoadUserDefine()` โหลดโครงคอลัมน์ที่ผู้ใช้ปรับไว้ ; `OnExportTo()` ส่งออก ; `InsertOwnerRowData(rowset, ระดับ)` เติมแถวสำหรับคอลัมน์ที่ผู้ใช้นิยามเอง — คอลัมน์ชื่อฟิลด์ดึงค่าจาก rowset ตรง ๆ, คอลัมน์สูตร `DecodeFormula` (ถ้าผิด แสดง `#ERR`) ; รูปแบบไฟล์ที่ส่งออกและวิธีตัดหน้าของเอนจินพิมพ์ ยังไม่ได้อ่านละเอียด (ยังไม่ verify) | `BaseReportView.cpp:2804,2380,3234` |
| ค่าเริ่มต้นโหมดวัสดุก่อสร้าง | `m_bChmMode = TRUE` ตั้งไว้ในโปรแกรมหลัก (มี `FALSE` ถูก comment ไว้ที่บรรทัดก่อนหน้า) — รายงานที่มีกิ่ง "ชุด CHM/ปูน" (เช่น 5350) จึงใช้กิ่ง CHM เสมอ | `BC5Account.cpp:253,256` |
| ค่ารหัสที่พบซ้ำ | `BillType` (ชนิดใบขาย): 0 = ขายสด, 1 = ขายเชื่อ, 2 = ขายบริการสด, 3 = ขายบริการเชื่อ (ไทยตาม comment ท้ายบรรทัด) ; รายงานที่แยกสด/เชื่อใช้ `BillType IN (0,2)` = สด, `IN (1,3)` = เชื่อ · `TaxType`: 0 = ภาษีแยกนอก, 1 = ภาษีรวมใน, 2 = อัตราศูนย์ (ไทยของ 0/1 จาก comment `ARFrmARMasterViewPage2.cpp:128-129`, ของ 2 จาก key `TAX_ZERO` — ดูหัวข้อ 4) · `PaymentType` (ชนิดการชำระในใบรับชำระ): 0 = "CA", 1 = "CR", 2 = "CH", 3 = "CO" (รหัสอังกฤษฝังในโค้ด ; ความหมายไทยของ CR/CO เป็นการอนุมาน — ดูหัวข้อ 6) · `CauseType` (ใบรับคืน/ลดหนี้): 0 = รับคืนสินค้า/ลดหนี้, 1 = ลดหนี้ราคาผิด, 2 = สินค้าแตกเสียหาย (ไทยตาม comment ท้ายบรรทัด) | `companyinfo.cpp:111` ; `BLRepSaleOrderByPersonSumView.cpp:644-648` ; `BillRepReceiveMoneyBySaleView.cpp:296-305` ; `BLRepCreditNoteStatusView.cpp:223-231` |

ข้อสังเกตที่ BC **ห้ามลอก** (พบใน Champ แต่ขัดกฎของโปรเจกต์ — ใช้กับทั้ง 24 รายงาน): (1) คำนวณและเก็บยอดเงินด้วย `double` แล้วจัดรูปด้วย `%f` (ดูหัวข้อย่อย "ข้อสังเกต" ของแต่ละหัวข้อ) — BC ใช้ `NUMERIC(p,s)` และส่งยอดเป็นสตริงทศนิยมใน JSON ; (2) สร้าง SQL ด้วยการต่อสตริงจากค่าที่ผู้ใช้กรอก + `MakeUpper()` ทั้งก้อน (`BaseReportDialog.cpp:257`) — BC ใช้พารามิเตอร์ (`$1,$2`) เท่านั้น และห้ามแปลงตัวพิมพ์ของค่าที่ผู้ใช้ให้มา ; (3) ข้อความไทยฝังในโค้ด/SQL — BC ใช้ key ภาษาใน `languages.tsv` ; (4) ถาม N+1 (ค้นชื่อลูกค้า/สินค้าต่อแถว) — BC ใช้ JOIN ; (5) กติกา "ใบที่ยกเลิก" ไม่เป็นมาตรฐานเดียวกันข้ามรายงาน — BC ต้องกำหนดกติกาเดียว (ดูหัวข้อ 3.4/4.4/5.4) ; (6) ตัวอย่างเงื่อนไขอ้างตารางที่ไม่มีคอลัมน์นั้น/ตัวแปรตาย (ดูข้อสังเกตแต่ละหัวข้อ) — อย่าลอกเป็นสเปก ; (7) ตั้งชื่อตาราง/คอลัมน์แบบ `BCArInvoice`, `DocNo` ของ Champ ใช้เป็นหลักฐานเท่านั้น — BC ใช้ `snake_case` ตัวพิมพ์เล็กทั้งหมด.

---

### 1. กลุ่มใบเสนอราคา (4 รายงาน)

| เมนู | ชื่อไทยตาม menuconfig.xml | handler | view class (ไฟล์) | dialog class | ชื่อรายงาน (key) |
|---|---|---|---|---|---|
| 5344 | รายงานใบเสนอราคา | `OnBillsysTransrep2` MainFrm.cpp:4820 | `CBLRepBillQuotationView` BLRepBillQuotationView.cpp (350 บรรทัด) | `CBLRepBillQuotationDlg` (IDD_BILL_REP_QUOTATION, BCRepBilling.rc:8796) | 34002 |
| 5333 | รายงานการบันทึกใบเสนอราคาพร้อมรายการย่อย | `OnBillsysTransrep1` MainFrm.cpp:4806 | `CBLRepBillQuotationSubView` BLRepBillQuotationSubView.cpp (456 บรรทัด) | dialog เดียวกับ 5344 | 34001 (แต่ dialog ตั้ง title 34002 ที่ BLRepBillQuotationDlg.cpp:16) |
| 5345 | รายงานสถานะใบเสนอราคา | `OnBillsysAnalyrep20` MainFrm.cpp:5017 | `CBLRepBillQuotationStatusView` BLRepBillQuotationStatusView.cpp | `CBLRepBillQuotationStatusDlg` (IDD_BL_REP_QUOTATION_STATUS, rc:5340, caption "รายงานแสดงสถานะใบเสนอราคา") | 34020 |
| 5354 | รายงานการยกเลิกใบเสนอราคาสินค้าพร้อมรายการย่อย | `OnBillsysTransrep3` MainFrm.cpp:4835 | `CBLRepBillQuotationCancelView` BLRepBillQuotationCancelView.cpp | `CBLRepBillQuotationCancelDlg` (IDD_BILL_REP_QUOTATION_CANCEL, rc:8836) | 34003 |

- drill-down ทั้ง 4 รายงานตั้งคำสั่ง 3 ช่อง (MainFrm.cpp:4806-4840, :5017): ช่อง 0 = 5303 บันทึกใบเสนอราคาสินค้า (`ID_BILLSYS_TRANS_1`), ช่อง 1 = 5104 รายละเอียดลูกหนี้ (`ID_ARSYS_ARMASTER`), ช่อง 2 = 5602 รายละเอียดสินค้า (`ID_ICSYS_ITEMMASTER`). การคลิกเรียก `OnCallTrans` ตั้ง filter `BCQuotation.DOCNO='…'` / `BCAR.Code='…'` / `BCITEM.Code='…'` แล้วเปิดจอปลายทาง (BLRepBillQuotationView.cpp:66-100).
- ฟอร์มต้นทางของเอกสาร: 5303 (รายละเอียดฟิลด์อยู่ใน part ใบเสนอราคา ไม่ซ้ำที่นี่).

#### 1.1 ช่องเงื่อนไข (เรียงบน-ล่างตามตำแหน่งใน .rc)

| # | ป้ายไทย (จาก .rc) | 5344 / 5333 (dialog ร่วม) | 5345 | 5354 | ชนิด / การทำงาน |
|---|---|---|---|---|---|
| 1 | พิมพ์จากวันที่ : / พิมพ์ถึงวันที่ : | มี | มี (`BCQuotation.DocDate`) | มี | ช่องวันที่ (`InitialDate` Dlg.cpp:73-74); "ถึง" ว่าง = `DocDate = จาก` (BLRepBillQuotationDlg.cpp:129,139 ; StatusDlg.cpp:113,119 ; CancelDlg.cpp:123,133) |
| 2 | พิมพ์จากเลขที่เอกสาร : / พิมพ์ถึงเลขที่เอกสาร : (5345 ใช้ป้าย "พิมพ์จากใบเสนอราคา : / พิมพ์ถึงใบเสนอราคา :") | มี | มี | มี | ช่องลิงก์เลขที่เอกสาร (`InitialLink` Dlg.cpp:76-77); ถึงว่าง = `DocNo = จาก` (Dlg.cpp:158,166 ; StatusDlg.cpp:131,135 ; CancelDlg.cpp:152,160) |
| 3 | พิมพ์จากลูกค้า : / พิมพ์ถึงลูกค้า : | มี | **ไม่มี** | มี | ช่องลิงก์รหัสลูกค้า; `ArCode BETWEEN` / `ArCode =` (Dlg.cpp:183,191 ; CancelDlg.cpp:177,185) |
| 4 | พิมพ์จากสินค้า : / พิมพ์ถึงสินค้า : | มี — กรองระดับ "เอกสาร" `DocNo IN (select Distinct DocNo from BCQuotationSub where ItemCode …)` (Dlg.cpp:208,216) | มี — กรองระดับ "บรรทัด" `BCQuotationSub.ItemCode BETWEEN` (StatusDlg.cpp:147) | มี — กรองบรรทัด `ItemCode BETWEEN` (CancelDlg.cpp:202) | ช่องลิงก์รหัสสินค้า |
| 5 | คำตอบจากลูกค้า : (combo) | มี | ใช้ป้าย "สถานะใบเสนอราคา :" แทน | มีบนจอแต่ **ใช้ไม่ได้** (สร้างรายการ+กรองถูกคอมเมนต์ CancelDlg.cpp:74-84, :217-237) | ดูตารางตัวเลือกด้านล่าง |
| 6 | [ ] แสดงจำนวนใบเสนอราคาตามวันที่ที่กำหนด | มี (`IDC_CHECK_SHOWCOUNT`, Dlg.cpp:43) | ไม่มี | ไม่มี | เมื่อติ๊ก และพิมพ์เงื่อนไขท้ายรายงาน (`bDisplayCondition`) จะมีบรรทัด "จำนวนใบเสนอราคา N รายการ" (key `NUMBER_QUOTATION`) จาก `SELECT COUNT(DocNo) FROM BCQuotation WHERE IsCancel = 0 <เงื่อนไขวันที่/เลขที่/ลูกค้า/สินค้า>` (Dlg.cpp:252-254; View.cpp:336-348) — ไม่รวมเงื่อนไขคำตอบลูกค้า |
| ท้ายจอ | กรอบ "เงื่อนไขการพิมพ์รายงานมาตรฐาน", ภาพตัวอย่าง `BLRepBillQuotationDlg.jpg`, ตารางเงื่อนไขผู้ใช้ | ใช่ | ใช่ | ใช่ | ดู R0 ข้อ 3 |

**combo "คำตอบจากลูกค้า" (5344/5333)** — กรองคอลัมน์ `AssertStatus` ของ `BCQuotation` (BLRepBillQuotationDlg.cpp:84-90, :227); ค่าเริ่มต้น = "ทั้งหมด" (`SetComboData(3)` :90)

| ลำดับ | ข้อความ (key → ไทยจาก comment ท้ายบรรทัดโค้ด) | SQL |
|---|---|---|
| 0 | `STATUS_CUSTOMER1` = รอตอบกลับ | `AND AssertStatus = 0` |
| 1 | `STATUS_CUSTOMER2` = ตอบกลับแล้ว | `AND AssertStatus = 1` |
| 2 | `STATUS_CUSTOMER3` = ไม่รับในราคา | `AND AssertStatus = 2` |
| 3 | `ALL` = ทั้งหมด | ไม่กรอง |

**combo "สถานะใบเสนอราคา" (5345)** — ค่าเริ่มต้นข้อ 0 (BLRepBillQuotationStatusDlg.cpp:58-67); เงื่อนไขต่อท้ายคำสั่งหลัก (:162-208)

| ลำดับ | ข้อความ (key → ไทยจาก comment) | SQL |
|---|---|---|
| 0 | `QUOTATION_NOTUSE` = ยังไม่ได้นำไปทำรายการ | `BCQuotation.IsCancel = 0 AND BCQuotationSub.Qty = BCQuotationSub.RemainQty` (:166) |
| 1 | `QUOTATION_PARTUSE` = นำไปทำรายการบางส่วน | `IsCancel = 0 AND Qty <> RemainQty AND RemainQty > 0` (:173) |
| 2 | `QUOTATION_ENDUSE` = นำไปทำรายการหมดแล้ว | `IsCancel = 0 AND RemainQty = 0` (:180) |
| 3 | `QUOTATION_CANCEL` = รายการที่ยกเลิกไปแล้ว | `BCQuotation.IsCancel = 1` |
| 4 | `LIST_AT_APPROVE` = รายการที่อนุมัติไปแล้ว | `BCQuotation.IsConfirm = 1 AND IsCancel = 0` |
| 5 | `EXPIRED` = หมดอายุแล้ว | `BCQuotation.ExpireDate <= <วันนี้> AND IsCancel = 0` — วันนี้ทำเป็น พ.ศ. เมื่อ `YearType == 0` (:197-208) |

#### 1.2 SQL ตั้งต้นและการเรียง

| เมนู | คำสั่งตั้งต้น (`m_strDefaultSQL`) | เรียง | หมายเหตุ |
|---|---|---|---|
| 5344 / 5333 | `SELECT * FROM BCQuotation WHERE IsCancel = 0`; ถ้า `IsPrintCancel == 1` เป็น `WHERE DocNo IS NOT NULL` (พิมพ์ใบที่ยกเลิกด้วย) (BLRepBillQuotationDlg.cpp:19-21) | `ORDER BY DocDate, DocNo, ArCode` (:249) | ใบที่ยกเลิกแสดงเป็น `เลขที่(ข้อความ ASSERT3)` (View.cpp:247-248) |
| 5345 | `SELECT BCQuotationSub.* FROM BCQuotationSub, BCQuotation WHERE BCQuotationSub.TransState = 0 AND BCQuotationSub.DocNo = BCQuotation.DocNo` (StatusDlg.cpp:19) | `ORDER BY BCQuotation.DocNo, BCQuotationSub.LineNumber, ItemCode, UnitCode` (:215) | ผลเป็นรายบรรทัดสินค้า; หัวเอกสารสร้างในโค้ดแสดงผลเมื่อเลขที่เปลี่ยน |
| 5354 | `SELECT *, (SELECT DueDate FROM BCQuotation WHERE DocNo = BCQuotationSub.DocNo) AS DueDate FROM BCQuotationSub WHERE TransState = 0 AND IsCancel = 1` (CancelDlg.cpp:19) | `ORDER BY DocNo, DocDate, ArCode` (:243) | เงื่อนไขวันที่/ลูกค้าอ้างคอลัมน์ `DocDate`/`ArCode` ของ BCQuotationSub — ยังไม่ verify ว่าตารางนี้มีคอลัมน์ดังกล่าวใน DDL |

#### 1.3 คอลัมน์ตามลำดับ

**5344 — 1 เอกสาร = 1 แถว** (BLRepBillQuotationView.cpp:161-170, :245-260)

| # | key | ความหมาย (ไทยจาก comment บรรทัดเดียวกัน หรืออนุมาน) | จัดชิด/กว้าง | ที่มา |
|---|---|---|---|---|
| 0 | `DOCDATE` | วันที่ (อนุมาน) | ซ้าย 200 | `DocDate` |
| 1 | `DOCNO` | เลขที่เอกสาร (อนุมาน) — คลิกเปิดจอ 5303 | ซ้าย 250 | `DocNo` (+ `(ASSERT3)` ถ้าใบยกเลิก) |
| 2 | `DUE_DATE` | วันครบกำหนด (อนุมาน) | ซ้าย 200 | `DueDate` |
| 3 | `ARCODE` | รหัสลูกค้า (อนุมาน) — คลิกเปิด 5104 | ซ้าย 250 | `ArCode` |
| 4 | `ARNAME` | ชื่อลูกค้า (อนุมาน) | ซ้าย 600 | `GetArName(ArCode)` |
| 5 | `SUM_TAX` | ยอดภาษี | ขวา 250 | `TaxAmount × ExchangeRate` |
| 6 | `SUMTAXAMOUNT` | ยอดรวมภาษี | ขวา 270 | `TotalAmount × ExchangeRate` |
| 7 | `SUM_EXCEPT_TAX` | ยอดยกเว้นภาษี | ขวา 270 | `ExceptTaxAmount × ExchangeRate` |
| 8 | `BILL_NET_TOTAL` | ยอดเงินสุทธิ | ขวา 270 | `NetAmount × ExchangeRate` |

**5333 — หัวเอกสาร 9 คอลัมน์เหมือน 5344 (ระดับ 0 ตัวหนา) + รายการสินค้า (ระดับ 1)** (BLRepBillQuotationSubView.cpp:168-187, :279-340)

| # | key | ความหมาย | ที่มา |
|---|---|---|---|
| 0 | `ITEMCODE` | รหัสสินค้า — คลิกเปิด 5602 | `ItemCode` |
| 1 | `ITEMNAME` | ชื่อสินค้า | `GetItemName`; ถ้า `ItemCode == "."` ใช้ `MyDescription` |
| 2 | `UINTCODE` | หน่วยนับ | `รหัส/ชื่อ` จาก `BCItemUnit` |
| 3 | `PERPRICE` | ราคาขาย/หน่วย | `Price × rate` (ว่างถ้า ≤ 0) |
| 4 | `QUANTITY_ITEM` | ปริมาณ | `Qty` |
| 5 | `SUM_ITEMAMOUNT` | ยอดรวมสินค้า | `Qty × Price × rate` (คำนวณใหม่ ไม่ใช่คอลัมน์ DB) |
| 6 | `DISCOUNTM` | ส่วนลด | `DiscountAmount × rate` |
| 7 | `AMOUNT` | จำนวนเงิน | `Amount × rate` |
| 8 | `NETAMOUNT` | ยอดสุทธิ | `NetAmount × rate` |

อ่านบรรทัดจาก `SELECT * FROM BCQuotationSub WHERE DocNo = '<เลขที่>' AND TransState = 0 AND IsCancel = 0 <m_strQuery> ORDER BY LineNumber` (SubView.cpp:282-283). `m_strQuery` ได้ค่าจาก dialog เฉพาะตัวกรองสินค้า (Dlg.cpp:220).

**5345 — หัวเอกสาร 7 คอลัมน์ + บรรทัด 9 คอลัมน์** (BLRepBillQuotationStatusView.cpp:163-180)

| ระดับ | # | key | ความหมาย |
|---|---|---|---|
| 0 ตัวหนา (คลิกคอลัมน์ 0 และ 2, :301) | 0 | `QUOTATION_NO` | เลขที่ใบเสนอราคา — คลิกเปิด 5303 |
| 0 | 1 | `DOCNO_DATE` | วันที่ออกเอกสาร |
| 0 | 2 | `ARCODE` | รหัสลูกค้า — คลิกเปิด 5104 |
| 0 | 3 | `AR_NAME` | ชื่อลูกค้า |
| 0 | 4 | `QUOTATION_DEPART` | แผนกที่เสนอราคา (`รหัส/ชื่อ` จาก `BCDepartment`) |
| 0 | 5 | `Sale_Code` | รหัสพนักงานขาย |
| 0 | 6 | `SALE_NAME` | ชื่อพนักงานขาย (`BCSale`) |
| 1 | 0 | `ITEMCODE` | รหัสสินค้า (คลิก) |
| 1 | 1 | `ITEMNAME` | ชื่อสินค้า |
| 1 | 2 | `UINTCODE` | หน่วยนับ |
| 1 | 3 | `UNITPRICE` | ราคา/หน่วย (ประกาศ `DT_LEFT` ทั้งที่เป็นตัวเลข :174) |
| 1 | 4 | `QUANTITY` | จำนวนสินค้า |
| 1 | 5 | `DISCOUNTM` | ส่วนลด |
| 1 | 6 | `AMOUNT` | จำนวนเงิน |
| 1 | 7 | `BILL_INVBALANCE` | ยอดคงค้าง = `RemainQty` (จำนวนที่ยังไม่ถูกนำไปใช้; comment "เพิ่ม ohm 22-07-2546 CHM" :178) |
| 1 | 8 | `NETAMOUNT` | ยอดสุทธิ |

**5354 — หัวเอกสาร 9 + บรรทัด 9 คอลัมน์ เหมือน 5333 ทุกคอลัมน์** (BLRepBillQuotationCancelView.cpp:164-183) ข้อแตกต่าง:
- `DOCNO` ต่อท้ายวงเล็บ: ถ้ามีแถวใน `BCCancelBillSub WHERE RefNo = <เลขที่>` = "(ยกเลิกรายการ)" (key `IDCANCEL`+`TRANSAC`) มิฉะนั้น = "(ยกเลิกเอกสาร)" (key `IDCANCEL`+`DOCUMENT`) (CancelView.cpp:273-276, :488-494; ไทยจาก comment)
- ยอดหัวเอกสาร (ภาษี/รวม/ยกเว้น/สุทธิ/สกุลเงิน/อัตราแลก) ดึงใหม่จาก `BCQuotation WHERE DocNo = …` ทีละเอกสาร (CancelView.cpp:474)

#### 1.4 ยอดรวม การจัดกลุ่ม การขึ้นหน้า

| เมนู | โครงแถว | รวมย่อยต่อเอกสาร | รวมหน้า (ถ้า `IsPrintSumPage == 0`) | รวมทั้งสิ้น |
|---|---|---|---|---|
| 5344 | ระดับ 0 ล้วน (`AddRow("1,3",0,FALSE)` :260) | ไม่มี | `SUM_LISTSUB_PAGE N TRANS_TYPE` + คอลัมน์ 5-8 (:292-308) | `SUMTOTALDETAIL_WORD N TRANS_TYPE` + คอลัมน์ 5-8 ตัวหนา (:274-283) |
| 5333 | หัวเอกสาร (ระดับ 0 ตัวหนา คลิก 1,3 :279) → รายการ (ระดับ 1 :340) | `SUMDETAIL_WORD N TRANS_TYPE` + จำนวน/ยอดรวมสินค้า/ส่วนลด/จำนวนเงิน/ยอดสุทธิ คอลัมน์ 4-8 (:368) | ระดับ 1 (:398-414) | คอลัมน์ 4-8 (:389) |
| 5345 | หัวเอกสาร (:301) → บรรทัด (:326) | หลังแต่ละเอกสาร คอลัมน์ 4-8 (:277, :356) | (:380-397) | (:371) |
| 5354 | เหมือน 5333 | เหมือน 5333 | เหมือน 5333 | เหมือน 5333 |

- ไม่พบโค้ดตัดหน้าตามกลุ่ม (page-break-per-group) ใน 4 รายงานนี้; `m_nPageCount` ใช้นับรายการเพื่อพิมพ์รวมหน้าเท่านั้น. การตัดหน้าจริงเป็นของ engine พิมพ์ใน CBaseReportView — ยังไม่ verify.
- ท้ายรายงานถ้า `bDisplayCondition`: หัว `DISPLAYCONDITION` + บรรทัดเงื่อนไข (แยกด้วย `@`) (BLRepBillQuotationView.cpp:318-348).

#### 1.5 ตารางที่อ่าน

| เมนู | ตาราง (SQL Server) |
|---|---|
| 5344 | `BCQuotation`, `BCCurrency` (สัญลักษณ์สกุลเงิน), `BCAR` (ชื่อลูกค้าผ่าน `GetArName`) |
| 5333 | + `BCQuotationSub`, `BCItemUnit`, `BCItem` (ชื่อสินค้า) |
| 5345 | `BCQuotationSub` JOIN `BCQuotation`, `BCDepartment`, `BCSale`, `BCItemUnit` |
| 5354 | `BCQuotationSub`, `BCQuotation`, `BCCancelBillSub` |

#### 1.6 ข้อสังเกตเฉพาะกลุ่ม (BC ห้ามลอก)
1. 5345: แถวรวมย่อยของ **เอกสารสุดท้าย** ใช้ `SumNetAmount` (ผลรวมทั้งรายงาน) ที่คอลัมน์ 8 แทน `SumSubNetAmount` (BLRepBillQuotationStatusView.cpp:355) — ยอดสุทธิย่อยของใบสุดท้ายผิดเมื่อมีหลายใบ. BC ต้องรวมย่อยแยกต่อเอกสาร.
2. 5354: ช่อง "คำตอบจากลูกค้า" อยู่บนจอแต่ไม่มีรายการและไม่มีผลต่อข้อมูล (CancelDlg.cpp:74-84, :217-237 คอมเมนต์ไว้) — BC ไม่ควรแสดงตัวกรองที่ไม่ทำงาน.
3. 5333: ตัวกรองสินค้าเลือก "เอกสาร" ที่มีสินค้านั้นแล้วแสดง **ทุกบรรทัด** ของเอกสาร (BLRepBillQuotationDlg.cpp:147 + SubView.cpp:282) ขณะที่ 5345/5354 กรองระดับบรรทัด — BC ต้องเลือกพฤติกรรมเดียวและระบุไว้ในหัวรายงาน.
4. 5354: อ่าน `ExchangeRate` ด้วยทศนิยม `AmountPoint` (BLRepBillQuotationCancelView.cpp:483) ทั้งที่ที่อื่นใช้ 9 ตำแหน่ง — ยอดสกุลเงินต่างประเทศคลาดเคลื่อน.
5. 5344 และ 5333 ใช้ `m_strTitle` เดียวกัน (34002) และ caption ของ dialog ("รายงานการใบเสนอราคาพร้อมรายการย่อย") ไม่ตรงกับเมนู 5344 "รายงานใบเสนอราคา" — BC ใช้ชื่อเมนูเป็นหัวรายงานเดียว.
6. ทุกเงื่อนไขสร้างด้วยการต่อสตริง `Format("… '%s' …")` จากค่าที่ผู้ใช้พิมพ์ (เช่น BLRepBillQuotationDlg.cpp:129,158,183,208) และ SQL ทั้งก้อนถูก `MakeUpper` (BaseReportDialog.cpp:257) — BC ใช้พารามิเตอร์ (`$1`) และไม่แปลงตัวพิมพ์ค่าที่ผู้ใช้กรอก.

---

### 2. กลุ่มใบสั่งขาย/สั่งจอง (2 รายงาน)

| เมนู | ชื่อไทยตาม menuconfig.xml | handler | view class | dialog class | ชื่อรายงาน (key) |
|---|---|---|---|---|---|
| 5363 | รายงานการใบสั่งขาย/จองสินค้าพร้อมรายการย่อย | `OnBillsysTransrep7` MainFrm.cpp:4850 | `CBLRepSaleOrderDetailView` BLRepSaleOrderDetailView.cpp | `CBLRepSaleOrderDetailDlg` (IDD_BL_REP_SALEORDER_DETAIL 405x332) | 34007 |
| 5364 | รายงานการยกเลิกใบสั่งขาย/สั่งจองสินค้าพร้อมรายการย่อย | `OnBillsysTransrep8` MainFrm.cpp:4863 | `CBLRepSaleOrderCancelView` BLRepSaleOrderCancelView.cpp | `CBLRepSaleOrderCancelDlg` (IDD_BL_REP_SALEORDER_CANCEL 405x319) | 34008 |

- drill-down (MainFrm.cpp:4850-4876): ช่อง 0 = 5326 บันทึกใบสั่งขาย/สั่งจองสินค้า (`ID_BILLSYS_TRANS_4`), ช่อง 1 = 5104 รายละเอียดลูกหนี้, ช่อง 2 = 5602 รายละเอียดสินค้า. หัวเอกสารคลิกได้คอลัมน์ 1 (เลขที่) และ 2 (รหัสลูกค้า) (`AddRow("1,2",0,FALSE,FW_BOLD)` View.cpp:303 / Cancel :300); บรรทัดสินค้าคลิกคอลัมน์ 0 (`AddRow("0",1)` :329 / :325).

#### 2.1 ช่องเงื่อนไข (ตามลำดับบนจอ)

| # | ป้ายไทย (.rc) | 5363 | 5364 | ชนิด / SQL |
|---|---|---|---|---|
| 1 | พิมพ์จากวันที่ : / พิมพ์ถึงวันที่ : | มี | มี | ช่องวันที่; `BCSaleOrderSub.DocDate BETWEEN` หรือ `=` (BLRepSaleOrderDetailDlg.cpp:141,147) |
| 2 | พิมพ์จากเลขที่เอกสาร : / ถึง | มี | มี | `BCSaleOrderSub.DocNo BETWEEN` / `=` (:161,166) |
| 3 | พิมพ์จากลูกค้า : / ถึง | มี | มี | `BCSaleOrderSub.ArCode BETWEEN` / `=` (:180,183) |
| 4 | พิมพ์จากวันที่ครบกำหนด : / พิมพ์ถึงวันที่ครบกำหนด : | มี | มี | `BCSaleOrder.DueDate BETWEEN` / `=` (:199,205; Cancel :187,193) |
| 5 | พิมพ์จากรหัสพนักงานขาย : / ถึง | มี (combo จาก `Select code,name from BCSale` :81-82) | มี | `BCSaleOrderSub.SaleCode BETWEEN` / `=` (:219,222) |
| 6 | พิมพ์จากสินค้า : / พิมพ์ถึงสินค้า : | มี | มี | `BCSaleOrderSub.ItemCode BETWEEN` / `=` — กรองระดับบรรทัด (:236,242) |
| 7 | ประเภทเอกสาร : (combo `IDC_ED_SOSTATUS`) | มี | มี | ดูตาราง |
| 8 | เลือกเอกสาร : (combo `IDC_ED_DOCTYPE`) | มี | **ไม่มี** | ดูตาราง |
| ท้ายจอ | กรอบเงื่อนไขมาตรฐาน + ภาพตัวอย่าง + ตารางเงื่อนไขผู้ใช้ | ใช่ | ใช่ | R0 ข้อ 3 |

| combo | ลำดับ → ข้อความ (key; ไทยจาก comment) | SQL | ค่าเริ่มต้น |
|---|---|---|---|
| ประเภทเอกสาร | 0 `SALES_ORDER` เฉพาะใบสั่งขายสินค้า / 1 `SO_RESERVATION` เฉพาะใบสั่งจองสินค้า / 2 `SUM_DOCNO` รวมเอกสาร (Detail :86-91, Cancel :84-89) | 0 → `BCSaleOrder.SOStatus = 0`; 1 → `SOStatus = 1`; 2 → ไม่กรอง (Detail :255-263) | 2 (รวมเอกสาร) |
| เลือกเอกสาร (เฉพาะ 5363) | 0 `NO_SALES` ยังไม่ทำขาย / 1 `REF_TO_SALES` ทำขายหมดแล้ว / 2 `ALL` ทั้งหมด (รายการ "ทำขายบางส่วน" `REF_TO_SALES2` ถูกคอมเมนต์ทิ้ง :95) (:93-99) | 0 → `BCSaleOrderSub.RemainQty > 0.0`; 1 → `RemainQty <= 0.0` (:283,289) | 2 (ทั้งหมด) |

#### 2.2 SQL ตั้งต้นและการเรียง

| เมนู | `m_strDefaultSQL` | เรียง |
|---|---|---|
| 5363 | `SELECT BCSaleOrderSub.* FROM BCSaleOrderSub, BCSaleOrder WHERE BCSaleOrderSub.DocNo = BCSaleOrder.DocNo AND BCSaleOrderSub.IsCancel = 0 AND BCSaleOrderSub.TransState = 0` (Detail Dlg.cpp:18) | `ORDER BY BCSaleOrderSub.DocDate, DocNo, ArCode, LineNumber` (:302) |
| 5364 | `… AND (BCSaleOrderSub.IsCancel = 1 OR BCSaleOrderSub.DocNo IN (select Distinct SONo from bcsovoidsub)) AND BCSaleOrderSub.TransState = 0` (Cancel Dlg.cpp:18) — รวมใบที่ยกเลิกทั้งใบ **และ** ใบที่มีรายการถูกยกเลิก/void ในตาราง `BCSoVoidSub` | `ORDER BY BCSaleOrderSub.DocDate, DocNo, ArCode, LineNumber` (:264) |

#### 2.3 คอลัมน์ตามลำดับ (ไทยตาม comment บรรทัดเดียวกัน)

**หัวเอกสาร (ระดับ 0 ตัวหนา — เหมือนกันทั้งสองรายงาน)** (5363 View.cpp:164-170, 5364 :163-169; ค่าที่ใส่ :295-301)

| # | key | ความหมาย | ที่มา |
|---|---|---|---|
| 0 | `DOCDATE` | วันที่ออกเอกสาร | `DocDate` |
| 1 | `SALE_DOCNO` | เลขที่ใบสั่งขาย/สั่งจอง — คลิก → 5326 | `DocNo` |
| 2 | `ARCODE` | รหัสลูกค้า — คลิก → 5104 | `ArCode` |
| 3 | `AR_NAME` | ชื่อลูกค้า | `GetArName` |
| 4 | `BILLDOCTYPE` | ประเภทเอกสาร | `SOStatus` 0 = ใบสั่งขายสินค้า (`PO_SALE`) / 1 = ใบสั่งจองสินค้า (`PO_ORDER`) (View.cpp:199-200, :299-300) |
| 5 | `SALE_ID` | พนักงานขาย (หัวคอลัมน์ไม่มี comment; ค่า = ชื่อพนักงานจาก `BCSale`) | `OLEDBGetMasterName("BCSale", SaleCode)` |

**บรรทัดสินค้า 5363 (ระดับ 1, 9 คอลัมน์)** (View.cpp:172-181, :319-329)

| # | key | ความหมาย | สูตร |
|---|---|---|---|
| 0 | `ITEMCODE` | รหัสสินค้า — คลิก → 5602 | `ItemCode` |
| 1 | `ITEMNAME` | ชื่อสินค้า | `GetItemName` |
| 2 | `UINTCODE` | หน่วยนับ | ชื่อหน่วยจาก `BCItemUnit` |
| 3 | `PERPRICE` | ราคาขาย/หน่วย | `Price × rate` (ว่างถ้า = 0) |
| 4 | `ITEMQTY` | จำนวน | `Qty` |
| 5 | `BALANCE_NUMBER` | ยอดคงเหลือ (จำนวน) | `RemainQty` |
| 6 | `DISCOUNTM` | ส่วนลด | `DiscountAmount × rate` |
| 7 | `AMOUNT` | จำนวนเงิน | `Amount × rate` |
| 8 | `ITEMAMOUNT` | จำนวนเงิน (comment) — เป็นมูลค่าคงเหลือ | `RemainQty × Price` แล้วหักส่วนลดด้วย `XDiscount(…, DiscountWord)` เมื่อ `DiscountMethod == 0` (View.cpp:238-243) แล้ว `× rate` (:263) |

**บรรทัดสินค้า 5364 (ระดับ 1, 8 คอลัมน์)** (Cancel View.cpp:171-179, :316-325): เหมือน 5363 แต่ **ไม่มี** `BALANCE_NUMBER` — คอลัมน์ 5 = `DISCOUNTM`, 6 = `AMOUNT`, 7 = `ITEMAMOUNT` (มูลค่าคงเหลือ สูตรเดียวกัน)

#### 2.4 ยอดรวมและการจัดกลุ่ม

| ส่วน | 5363 | 5364 |
|---|---|---|
| จัดกลุ่ม | เมื่อ `DocNo` เปลี่ยน พิมพ์หัวเอกสาร (:295-303) | เหมือนกัน (:296-300) |
| รวมย่อยต่อเอกสาร | `SUMDETAIL_WORD` + คอลัมน์ 4-8 (จำนวน, คงเหลือ, ส่วนลด, จำนวนเงิน, มูลค่าคงเหลือ) (:275-280, :355-360) | คอลัมน์ 4-7 (จำนวน, ส่วนลด, จำนวนเงิน, มูลค่าคงเหลือ) (:277, :355) |
| รวมหน้า (ถ้า `IsPrintSumPage == 0 && m_nPageCount > 0`) | `SUM_LISTSUB_PAGE` + คอลัมน์ 4-8 (:383-400) | คอลัมน์ 4-7 (:377-393) |
| รวมทั้งสิ้น | `SUMTOTALDETAIL_WORD` + คอลัมน์ 4-8 (:368-373) | คอลัมน์ 4-7 (:367) |
| สกุลเงิน | ถ้า `MultiCurrency == 1 && bHomeCurrencyReport == 1` คูณ `ExchangeRate` และต่อสัญลักษณ์ "฿"; มิฉะนั้นแสดงตามสกุลเอกสาร (`BCCurrency.Symbol`); ซ่อนสัญลักษณ์ถ้า `CurrencySymbol == 0` (:245-258) | เหมือนกัน |

#### 2.5 ตารางที่อ่าน
5363: `BCSaleOrderSub`, `BCSaleOrder`, `BCSale`, `BCItemUnit`, `BCCurrency`. 5364: เหมือนกัน + `BCSoVoidSub` (ผ่าน subquery `SONo`).

#### 2.6 ข้อสังเกต (BC ห้ามลอก)
1. ไม่มีเงื่อนไข `IsPrintCancel` ใน 5363 — ใบที่ยกเลิกทั้งใบไม่แสดงเลย (ถูกตัดที่ `BCSaleOrderSub.IsCancel = 0` Dlg.cpp:18) แต่ไปแสดงใน 5364 แทน; BC ควรทำสถานะเอกสารเดียวและให้กรองด้วย "สถานะ" ไม่ต้องแยกรายงาน 2 ตัว (แต่ยึดชื่อเมนูตาม Champ).
2. ตัวกรองผสมตารางหัว (`BCSaleOrder.DueDate`, `SOStatus`) กับตารางบรรทัด (`BCSaleOrderSub.*`) ด้วยการ JOIN แบบ implicit (`FROM A, B WHERE …`) ในสตริง SQL — BC ใช้ `JOIN` ชัดเจนและพารามิเตอร์.
3. หัวคอลัมน์ `ITEMAMOUNT` ถูก comment ว่า "จำนวนเงิน" ซ้ำกับ `AMOUNT` แต่ค่าเป็น "มูลค่าคงเหลือ" — BC ต้องตั้งป้ายใหม่ให้ตรงความหมาย (เช่น "มูลค่าคงเหลือ") ขณะที่ลำดับคอลัมน์เหมือน Champ.
4. ไม่ได้ verify: ข้อความไทยจริงของ `BILLDOCTYPE`, `SALE_ID`, `PO_SALE`, `PO_ORDER`, `BALANCE_NUMBER` (อยู่ใน DB `myresource` ไม่อยู่ใน repo).

---

### 3. กลุ่มเงินมัดจำและเงินรับล่วงหน้า (5 รายงาน)

| เมนู | ชื่อไทยตาม menuconfig.xml | handler | view class (ไฟล์) | dialog class | ชื่อรายงาน (key) |
|---|---|---|---|---|---|
| 5346 | รายงานการบันทึกใบรับเงินมัดจำ | `OnBillsysTransrep21` MainFrm.cpp:4890 | `CBILLRepArDepositView` BILLRepArDepositView.cpp (408 บรรทัด) | `CBILLRepArDepositDlg` (IDD_BILL_REP_ARDEPOSIT, BCRepBilling.rc:957; caption ใน rc = "รายงานการตัดใบรับเงินมัดจำ" ไม่ตรงชื่อเมนู) | 34021 |
| 5334 | รายงานการตัดใบรับเงินมัดจำ | `OnBillsysTransrep10` MainFrm.cpp:4904 | `CBILLRepArCutDepositView` BILLRepArCutDepositView.cpp (376 บรรทัด) | `CBILLRepArCutDepositDlg` (IDD_BL_REP_ARCUTDEPOSIT, rc:6343; caption "รายงานใบรับเงินมัดจำ" สลับกับ 5346) | 34010 |
| 5335 | รายงานใบรับเงินล่วงหน้า | `OnBillsysTransrep11` MainFrm.cpp:5108 | `CBLRepReceiveEarnestView` BLRepReceiveEarnestView.cpp (344 บรรทัด) | `CBLRepReceiveEarnestDlg` (IDD_BILL_REP_EARNEST, rc:8977, caption "รายงานใบรับเงินล่วงหน้า") | 38100 |
| 5355 | รายงานการตัดใบรับเงินล่วงหน้า | `OnBillsysTransrep30` MainFrm.cpp:5121 | `CBLRepReceiveEarnestUseView` BLRepReceiveEarnestUseView.cpp (362 บรรทัด) | `CBLRepReceiveEarnestUseDlg` (IDD_BILL_REP_EARNEST_USE, rc:9008) | 34030 |
| 5341 | รายงานบันทึกคืนเงินรับล่วงหน้า | `OnBillsysTransrep17` MainFrm.cpp:5135 | `CBLRepRetrunEarnestView` BLRepRetrunEarnestView.cpp (344 บรรทัด) | `CBLRepRetrunEarnestDlg` (IDD_BILL_REP_EARNEST_RET, rc:9037; caption "รายงานการตัดใบรับเงินมัดจำ" ไม่ตรงชื่อเมนู) | 300603 |

- ความหมายของ 2 ตระกูล: **มัดจำ** = ตาราง `BCARDeposit` (เอกสาร 5327); **เงินรับล่วงหน้า** = ตาราง `BCARDepositSpecial` (เอกสาร 5329); คืนเงินรับล่วงหน้า = `BCReturnDepSpecial` (เอกสาร 5306)
- ชื่อเมนู 5346 / 5334 ในโปรแกรมต้นแบบสลับกับ caption ของ dialog ใน .rc (ดู quirk 3.6) — ใน BC ให้ใช้ **ชื่อเมนู** เป็นหัวรายงาน
- drill-down (ตาม MainFrm handler): 5346 ช่อง 0 = 5327 ใบรับเงินมัดจำ, 1 = 5104 ลูกหนี้, 2 = 5326 ใบสั่งขาย; 5334 ช่อง 0 = 5327, 1 = 5104, 2 = 5328 ใบขาย; 5335 ช่อง 0 = 5329 ใบรับเงินล่วงหน้า, 1 = 5104; 5355 ช่อง 0 = 5329, 1 = 5104, 2 = 5127 ใบเสร็จรับเงิน; 5341 ช่อง 0 = 5306 คืนเงินรับล่วงหน้า, 1 = 5329, 2 = 5104

#### 3.1 ช่องเงื่อนไข (เรียงบน-ล่างตาม .rc)

| # | ป้ายไทย (จาก .rc) | 5346 | 5334 | 5335 | 5355 | 5341 | การทำงาน |
|---|---|---|---|---|---|---|---|
| 1 | พิมพ์จากวันที่ / ถึงวันที่ | มี | มี | มี (rc:8977) | มี | มี | `DocDate BETWEEN` เมื่อมีทั้งคู่ ; "ถึง" ว่าง = `DocDate = จาก` (เช่น BLRepReceiveEarnestDlg.cpp:167-178) |
| 2 | พิมพ์จากเลขที่เอกสาร / ถึง | มี | มี | มี | มี | มี | `DocNo BETWEEN` / `DocNo =` (:204-212) ; ช่องลิงก์เลือกจาก browse (5341 ใช้ browse `ARReturnDepositSpecialConfig` BLRepRetrunEarnestDlg.cpp:108,123) |
| 3 | พิมพ์จากลูกหนี้ / ถึง | มี | มี | มี | มี | มี | `ArCode BETWEEN` / `ArCode =` (:187-195) ; ช่องลิงก์ลูกหนี้ |
| 4 | เลือกเอกสาร : (combo) | มี | ไม่มี | มี | ไม่มี | ไม่มี | ดูตารางด้านล่าง |
| ท้ายจอ | กรอบ "เงื่อนไขการพิมพ์รายงานมาตรฐาน", ภาพตัวอย่าง, ตารางเงื่อนไขผู้ใช้ | ใช่ | ใช่ | ใช่ | ใช่ | ใช่ | ดู R0 |

**combo "เลือกเอกสาร :" (5346 และ 5335 เหมือนกัน)** — 2 ตัวเลือก ค่าเริ่มต้นข้อ 0 (BLRepReceiveEarnestDlg.cpp:66-72; 5346 Dlg:76-80)

| ลำดับ | ข้อความ (key → ไทยจาก comment) | SQL ที่ต่อท้าย |
|---|---|---|
| 0 | `BILL_INVBALANCE` = ยอดคงค้าง | `AND BillBalance > 0.0` (5335 Dlg:218-224 ; 5346 Dlg:229-237) |
| 1 | `ALL` = ทั้งหมด | ไม่กรอง |

ตัวเลือกที่เลือกถูกพิมพ์ท้ายรายงานในบรรทัด "เลือกเอกสาร : …" (5335 Dlg:221,229).

#### 3.2 SQL ตั้งต้นและการเรียง

| เมนู | คำสั่งตั้งต้น (`m_strDefaultSQL`) | เรียง | หมายเหตุ |
|---|---|---|---|
| 5346 | `SELECT * FROM BCARDeposit WHERE IsCancel = 0` ; `IsPrintCancel==1` → `WHERE DocNo IS NOT NULL` (Dlg:28-30) | `ORDER BY DocDate, DocNo, ArCode` | 1 เอกสาร = 1 แถว |
| 5334 | `SELECT * FROM BCARDeposit WHERE IsCancel = 0` / `DocNo IS NOT NULL` (Dlg:27-29) | `ORDER BY DocDate, DocNo, ArCode` (Dlg:215) | แถวหัว = ใบรับเงินมัดจำ ; แถวย่อย = การตัดมัดจำ (อ่านตาราง `BCARDepositUse`) |
| 5335 | `SELECT * FROM BCARDepositSpecial WHERE IsCancel = 0` / `DocNo IS NOT NULL` (Dlg:19-21) | `ORDER BY DocDate, DocNo, ArCode` (Dlg:235) | 1 เอกสาร = 1 แถว |
| 5355 | `SELECT * FROM BCARDepositSpecial WHERE IsCancel = 0` / `DocNo IS NOT NULL` (Dlg:19-21 ; บรรทัด 18 เป็น SQL เดิมของ `BCARDeposit` ถูกคอมเมนต์) | `ORDER BY DocDate, DocNo, ArCode` (Dlg:205) | แถวย่อย = ใบเสร็จที่ตัดเงินรับล่วงหน้า (View:197) |
| 5341 | `SELECT * FROM BCReturnDepSpecial WHERE IsCancel = 0` / `DocNo IS NOT NULL` (Dlg:19-21) | `ORDER BY DocDate, DocNo, ArCode` (Dlg:206) | 1 เอกสาร = 1 แถว |

#### 3.3 คอลัมน์ตามลำดับ

**5346 — 1 เอกสารมัดจำ = 1 แถว, 16 คอลัมน์** (BILLRepArDepositView.cpp:163-178 ; ไทยของ key ส่วนใหญ่อยู่ใน DB `myresource` — ไทยในวงเล็บอนุมานจากชื่อ key และ comment)

| # | key | ความหมาย | ชิด |
|---|---|---|---|
| 0 | `DOCDATE` | วันที่ (อนุมาน) | ซ้าย |
| 1 | `DOCNO` | เลขที่เอกสาร ; คลิก → 5327 ; เอกสารยกเลิกพิมพ์ `เลขที่(ASSERT3)` เมื่อ `IsPrintCancel==1` | ซ้าย |
| 2 | `TAX_NO` | เลขที่ใบกำกับภาษี (อนุมาน) | ซ้าย |
| 3 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | ซ้าย |
| 4 | `ARNAME` | ชื่อลูกค้า | ซ้าย |
| 5 | `BEFORE_TAX` | ยอดก่อนภาษี | ขวา |
| 6 | `SUM_TAX` | ยอดภาษี | ขวา |
| 7 | `TOTAL_SUMTAX` | ยอดรวมภาษี | ขวา |
| 8 | `SUM_OF_TAX` | ภาษีหัก ณ ที่จ่าย | ขวา |
| 9 | `NETAMOUNT` | ยอดสุทธิ | ขวา |
| 10 | `TOTAL_CASH` | ยอดเงินสด | ขวา |
| 11 | `SUMBANKAMOUNT` | ยอดเงินโอน | ขวา |
| 12 | `SUMCREDITAMOUNT` | ยอดบัตรเครดิต | ขวา |
| 13 | `MST_CHQOUT` | ยอดเช็คจ่าย | ขวา |
| 14 | `SUMTOTAL` | ยอดเงินรวม | ขวา |
| 15 | `BALANCE` | ยอดคงเหลือ | ขวา |

- ใบมัดจำที่ถูกตัดด้วยใบลดหนี้ (`BCInvCreditNote` ชนิด `BillType=3`) จนผลรวม `DiffAmount` >= `NetAmount` จะ **ไม่แสดง** เว้นแต่เลือก "ทั้งหมด" (`dSumDiposit < dNetAmount || m_nDocType == 1` View:227-235 ; ถาม `BCInvCreditNote` ต่อทีละเอกสาร = N+1)
- เอกสารที่ยกเลิก (เมื่อ `IsPrintCancel==1`) แสดงแต่ **ไม่รวม** ในยอดรวม/ยอดรวมต่อหน้า

**5334 — โครงสร้าง 2 ระดับ** (BILLRepArCutDepositView.cpp:157-173 DefaultData ; OnCallTrans :72)

| ระดับ | # | key | ความหมาย | ชิด |
|---|---|---|---|---|
| หัว (ใบรับเงินมัดจำ) | 0 | `DOCNO_RECEIVE` | เลขที่ใบรับเงินมัดจำ ; คลิก → 5327 | ซ้าย |
| | 1 | `RECEIVEDATEVAT` | วันที่รับ | ซ้าย |
| | 2 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | ซ้าย |
| | 3 | `AR_NAME` | ชื่อลูกค้า | ซ้าย |
| | 4 | `NETAMOUNT` | ยอดสุทธิ — ถ้า `TaxType==0` (แยกนอก) ใช้ `BeforeTaxAmount` แทน `NetAmount` (View:214-215) | ขวา |
| ย่อย (การตัดมัดจำ) | 0 | `BEFORE_AMOUNT_CUT` | มูลค่าก่อนตัด = ยอดคงเหลือก่อนบรรทัดนั้น | ขวา |
| | 1 | `BILL_DOCNO` | เลขที่ใบขายสินค้า (`BCARDepositUse.DocNo`) ; คลิก → 5328 | ซ้าย |
| | 2 | `FA_SALEDATE` | วันที่ขาย | ซ้าย |
| | 3 | `DESCRIPTION` | คำอธิบาย (`MyDescription`) | ซ้าย |
| | 4 | `CUT_PAY_AMOUNT` | ยอดตัดจ่าย (`TaxType==0` ใช้ `NetAmount` ของบรรทัด, อื่น ๆ ใช้ `Amount` View:262-265) | ขวา |
| | 5 | `BALANCE` | ยอดคงเหลือ = max(ก่อนตัด − ตัดจ่าย, 0) | ขวา |

อ่านแถวย่อยด้วย `SELECT * FROM BCARDepositUse WHERE DepositNo='<เลขที่>' ORDER BY Balance Desc` (View:206) ; **เอกสารมัดจำที่ไม่มีแถวตัดเลยจะไม่ถูกพิมพ์** (เงื่อนไข `RowsetSub.GetRecordCount() > 0` View:207).

**5335 — 1 ใบรับเงินล่วงหน้า = 1 แถว, 12 คอลัมน์** (BLRepReceiveEarnestView.cpp:147-159)

| # | key | ความหมาย | ชิด / กว้าง |
|---|---|---|---|
| 0 | `DOCDATE` | วันที่ | ซ้าย 200 |
| 1 | `DOCNO_EARNEST` | เลขที่ใบรับเงินล่วงหน้า ; คลิก → 5329 ; ยกเลิกพิมพ์ `เลขที่(ASSERT3)` | ซ้าย 300 |
| 2 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | ซ้าย 220 |
| 3 | `ARNAME` | ชื่อลูกค้า | ซ้าย 400 |
| 4 | `REMARK` | หมายเหตุ (`MyDescription`) | ซ้าย 400 |
| 5 | `RECEIPT_TOTAL_AMOUNT` | ยอดรับสุทธิ (`NetAmount`) | ขวา 220 |
| 6 | `TOTAL_CASH` | ยอดเงินสด (`SumCashAmount`) | ขวา |
| 7 | `SUMBANKAMOUNT` | ยอดเงินโอน (`SumBankAmount`) | ขวา |
| 8 | `SUMCREDITAMOUNT` | ยอดบัตรเครดิต (`SumCreditAmount`) | ขวา |
| 9 | `MST_CHQOUT` | ยอดเช็คจ่าย (`SumChqAmount`) | ขวา |
| 10 | `SUMTOTAL` | ยอดเงินรวม = เงินสด + โอน + เช็ค + บัตร | ขวา |
| 11 | `BALANCE` | ยอดคงเหลือ (`BillBalance`) | ขวา |

**5355 — 2 ระดับ** (BLRepReceiveEarnestUseView.cpp:148-163, :166-)

| ระดับ | # | key | ความหมาย | ชิด |
|---|---|---|---|---|
| หัว (ใบรับเงินล่วงหน้า) | 0 | `DOCNO_EARNEST` | เลขที่ใบรับเงินล่วงหน้า ; คลิก → 5329 | ซ้าย 350 |
| | 1 | `RECEIVEDATEVAT` | วันที่รับ | ซ้าย 200 |
| | 2 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | ซ้าย 300 |
| | 3 | `AR_NAME` | ชื่อลูกค้า | ซ้าย 700 |
| | 4 | `NETAMOUNT` | ยอดสุทธิ (`TaxType==0` ใช้ `BeforeTaxAmount`) | ขวา 350 |
| ย่อย (ใบเสร็จที่ตัด) | 0 | `BEFORE_AMOUNT_CUT` | มูลค่าก่อนตัด | ขวา 350 |
| | 1 | `NO_INVOICE` | เลขที่ใบเสร็จ (`BCReceiptSub1.DocNo`) ; คลิก → 5127 | ซ้าย 300 |
| | 2 | `FA_SALEDATE` | วันที่ขาย (คอลัมน์วันที่ของใบเสร็จ) | ซ้าย 200 |
| | 3 | `DESCRIPTION` | คำอธิบาย (`MyDescription`) | ซ้าย 500 |
| | 4 | `CUT_PAY_AMOUNT` | ยอดตัดจ่าย (`PayAmount`) | ขวา 300 |
| | 5 | `BALANCE` | ยอดคงเหลือ = max(ก่อนตัด − ตัดจ่าย, 0) | ขวา 300 |

อ่านแถวย่อย: `SELECT * FROM BCReceiptSub1 WHERE InvoiceNo = '<เลขที่ใบรับเงินล่วงหน้า>' AND BillType = 8 ORDER BY DocDate,DocNo` (View:197) ; ใบรับเงินล่วงหน้าที่ยังไม่มีใบเสร็จตัด **ไม่ถูกพิมพ์**.

**5341 — 1 ใบคืนเงินรับล่วงหน้า = 1 แถว, 12 คอลัมน์** (BLRepRetrunEarnestView.cpp:152-164) — โครงเหมือน 5335 แต่เพิ่มคอลัมน์เลขที่เอกสารอ้างอิงและไม่มียอดคงเหลือ

| # | key | ความหมาย |
|---|---|---|
| 0 | `DOCDATE` | วันที่ (180) |
| 1 | `DOCNO` | เลขที่ใบคืนเงิน ; คลิก → 5306 |
| 2 | `DOCNO_EARNEST` | เลขที่ใบรับเงินล่วงหน้าที่คืน (`EarnestNo`) ; คลิก → 5329 |
| 3 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 |
| 4 | `ARNAME` | ชื่อลูกค้า |
| 5 | `REMARK` | หมายเหตุ |
| 6-11 | `RECEIPT_TOTAL_AMOUNT`, `TOTAL_CASH`, `SUMBANKAMOUNT`, `SUMCREDITAMOUNT`, `MST_CHQOUT`, `SUMTOTAL` | ยอดสุทธิ / เงินสด / โอน / บัตร / เช็ค / รวม (ตรงกับ 5335 คอลัมน์ 5-10) |

> 5341 **ไม่มีคอลัมน์ `BALANCE`** (InsertColumnData 0-11 จบที่ `SUMTOTAL` View:152-163) แต่ RunProcess ยังอ่าน `BillBalance` คูณอัตราแลกเปลี่ยนและสะสมยอดหน้า/ยอดรวมไว้ (View:211, :235, :251, :281) โดยไม่ได้พิมพ์ลงแถวใด ๆ (ดู 3.6 ข้อ 6).

#### 3.4 ยอดรวม การจัดกลุ่ม และการขึ้นหน้าใหม่

| เมนู | จัดกลุ่ม | ยอดรวมย่อย | ยอดรวมท้ายรายงาน | ยอดรวมต่อหน้า (เมื่อ `IsPrintSumPage==0`) | พิมพ์เงื่อนไข |
|---|---|---|---|---|---|
| 5346 | ไม่มี (แถวละ 1 เอกสาร) ; `AddRow("1,3",0)` View:326 | — | คอลัมน์ 5-15 พร้อมข้อความ "รวมทั้งสิ้น N รายการ" (`TOTAL_AMOUNT` / `LIST`) View:352-367 | OnAddPageSummery View:374-392 | `TextoutReport(m_saTextOut)` พิมพ์เสมอ (:369) |
| 5334 | ระดับ 0 = ใบมัดจำ (ตัวหนา, `AddRow("0,2",0,…,FW_BOLD)` View:249) ; ระดับ 1 = แถวตัดมัดจำ (`AddRow("1",1)` :290) | ท้ายแต่ละใบมัดจำ : "ยอดรวม N รายการ" (`SUMDETAIL_WORD` … `TRANS_TYPE`) เฉพาะคอลัมน์ ยอดตัดจ่าย (:309) | "ยอดรวมทั้งสิ้น N รายการ" (`SUMTOTALDETAIL_WORD`) คอลัมน์ ยอดตัดจ่าย (:328) | `SUM_LISTSUB_PAGE` N `LIST` + ยอดตัดจ่ายต่อหน้า (:336-350) ; ยอดคงเหลือไม่รวม | `bDisplayCondition` → `OnAddReportCondition` (:331) |
| 5335 | ไม่มี ; `AddRow("1,2",0)` :275 | — | คอลัมน์ 5-11 + ข้อความ "รวมทั้งสิ้น N รายการ" (`TOTAL_AMOUNT`, `LIST`) :294-306 | `SUM_LIST_PAGE` N `LIST` + ยอด 7 คอลัมน์ต่อหน้า :315-329 | `TextoutReport(m_saTextOut)` :310 |
| 5355 | ระดับ 0 = ใบรับเงินล่วงหน้า (ตัวหนา `AddRow("0,2",0)` :240) ; ระดับ 1 = ใบเสร็จที่ตัด (`AddRow("1",1)` :277) | ท้ายแต่ละใบ : "ยอดรวม N รายการ" + ยอดตัดจ่าย (:290-296) | "ยอดรวมทั้งสิ้น N รายการ" + ยอดตัดจ่าย (:309-315) | `SUM_LISTSUB_PAGE` N `LIST` + ยอดตัดจ่ายต่อหน้า (:324-340) | `bDisplayCondition` → `OnAddReportCondition` :318 |
| 5341 | ไม่มี ; `AddRow("1,2,3",0)` :278 | — | คอลัมน์ 6-11 + ข้อความ (:295-308) | `SUM_LIST_PAGE` N `LIST` + ยอด 6 คอลัมน์ :316-329 | `TextoutReport(m_saTextOut)` :311 |

กติกายอดรวมที่ต่างกัน:
- 5335: เอกสารยกเลิกไม่รวมในยอดรวมต่อหน้าและยอดรวมท้ายรายงาน (`nIsCancel!=1` View:247-283)
- 5341: **รวมเอกสารยกเลิกในยอดทุกตัว** (View:250-251, :280-281 บวกตรง ๆ ไม่มีเงื่อนไข `IsCancel` ต่างจาก 5335 :247-283) — ต่างจาก 5335 ทั้งที่โครงเกือบเหมือน
- 5334 / 5355: ไม่มีเงื่อนไข `IsCancel` ในยอดรวม (แถวย่อยมาจากตารางตัด ซึ่งไม่ผูกสถานะยกเลิกของใบแม่)
- 5335 และ 5341 ข้ามเอกสารที่ `NetAmount <= 0` (comment "ถ้ามีการลดหนี้ แล้ว ไม่ต้องแสดง" 5335 View:203, 5341 View:208) ; 5346 ใช้เงื่อนไขต่างกัน คือพิมพ์เมื่อ `dSumDiposit < dNetAmount || m_nDocType == 1` (ยอดมัดจำที่ถูกนำไปใช้ยังน้อยกว่ายอดสุทธิ — comment เดียวกัน BILLRepArDepositView.cpp:234) — 3 รายงานใช้กติกา "ซ่อนเมื่อถูกลดหนี้" คนละสูตร

#### 3.5 ตารางที่อ่าน

| เมนู | ตารางที่อ่าน (SQL Server) |
|---|---|
| 5346 | `BCARDeposit`, `BCInvCreditNote` (ตรวจเอกสารที่ถูกตัดจนหมด), `BCCurrency` (สัญลักษณ์สกุลเงิน) |
| 5334 | `BCARDeposit`, `BCARDepositUse`, `BCCurrency` |
| 5335 | `BCARDepositSpecial`, `BCCurrency` |
| 5355 | `BCARDepositSpecial`, `BCReceiptSub1`, `BCCurrency` |
| 5341 | `BCReturnDepSpecial`, `BCCurrency` |

ชื่อลูกหนี้ได้จาก `GetArName(ArCode)` (lookup แยกต่อแถว — N+1) ; สัญลักษณ์สกุลเงิน: ถ้า `MultiCurrency==1` และรายงานแบบ Home Currency ใช้ "฿" และคูณด้วย `ExchangeRate` (5334/5355 ใช้ `NewExchangeRate`), อื่น ๆ ใช้ `BCCurrency.Symbol` ; `CurrencySymbol==0` = ไม่พิมพ์สัญลักษณ์.

#### 3.6 ข้อสังเกตที่ BC ห้ามลอก (กลุ่มเงินมัดจำ/เงินรับล่วงหน้า)

1. ชื่อเมนูกับ caption ของ dialog ใน .rc ไม่ตรงกัน (5346 ↔ 5334 สลับกัน ; 5341 ใช้ caption "รายงานการตัดใบรับเงินมัดจำ") — BC ใช้ชื่อเมนูเป็นหัวรายงานทุกจุด
2. สร้างคำสั่ง SQL ด้วยการต่อสตริง `strSQL.Format("... '%s'")` ทุกจุด (View:197, :206 ฯลฯ) — BC ต้องใช้ parameterized query ; ไม่เขียนภาษาไทยในคำสั่ง SQL
3. เงินทุกช่องเป็น `double` + `XRound` ภายหลัง (`dNetAmount * dExchangeRate`) — BC ต้องใช้ decimal/minor unit และกำหนดจุดปัดเศษครั้งเดียว (Iron Rule 20)
4. ใบมัดจำ/ใบรับเงินล่วงหน้าที่ยังไม่มีรายการตัด **หายไปจากรายงาน** (5334, 5355 พิมพ์เฉพาะที่มีแถวตัด) — เจ้าของกิจการอยากเห็นยอดคงค้างด้วย; BC ควรให้ 5346/5335 เป็นรายงานคงค้าง และ 5334/5355 เป็นรายงานตัด แต่ต้องบอกชัดว่าไม่รวมใบที่ยังไม่ตัด
5. 5341 `AddTableForUserDefine("BCARDepositSpecial")` (View:137) ทั้งที่ข้อมูลมาจาก `BCReturnDepSpecial` — ตารางเงื่อนไขผู้ใช้จะเสนอคอลัมน์ผิดตาราง ; BC ต้องผูกฟิลเตอร์เพิ่มเติมกับตารางจริงของรายงาน
6. 5341 คำนวณ `dBillBalance` / `dTotalBillBalance` (View:211, :281) แต่ไม่พิมพ์ลงแถวหรือยอดรวม ; 5355 สะสม `m_dSumPageBalance` (View:267) แต่ไม่พิมพ์ — โค้ดตาย ไม่ต้องลอก
7. ยอดรวมของ 5341 รวมเอกสารยกเลิกแต่ 5335 ไม่รวม — BC กำหนดกติกาเดียว: เอกสารยกเลิกแสดงได้ (มี tag ยกเลิก) แต่ไม่รวมยอด
8. กรองยอดคงค้างด้วย `BillBalance > 0.0` (float compare) ; BC ใช้ decimal และเปรียบเทียบ `> 0` ที่ scale ของสกุลเงิน
9. 5334/5355 เลือกฐานยอดด้วย `TaxType==0` (แยกนอก) ที่หัวใช้ `BeforeTaxAmount` (5334 View:214-215) แต่แถวตัดของ 5334 ใช้ `NetAmount` ของบรรทัดตัด (:262-265) — ฐานของหัวกับแถวย่อยอ้างคนละคอลัมน์ ยังไม่ verify ว่ายอดตรงกันเมื่อ `TaxType==0` ; BC ควรเก็บ "ฐานภาษีที่บันทึกจริง" ของเอกสารแล้วใช้ฐานเดียวทั้งหัวและรายการตัด

---

### 4. กลุ่มรายงานขายสินค้าและบริการ (4 รายงาน)

| เมนู | ชื่อไทยตาม menuconfig.xml | handler | view class (ไฟล์) | dialog class | ชื่อรายงาน (key) | ทิศกระดาษ |
|---|---|---|---|---|---|---|
| 5337 | รายงานรายวันขายสินค้าและบริการ | `OnBillsysTransrep13` MainFrm.cpp:4934 | `CBILLRepArInvoiceView` billreparinvoiceview.cpp (418 บรรทัด) | `CBILLRepArInvoiceDlg` (IDD_BILL_REP_ARINVOICE, BCRepBilling.rc:912 ; caption ใน rc "รายงานใบจ่ายเงินมัดจำ" ไม่ตรงชื่อเมนู) | 34013 | แนวนอน A4 |
| 5336 | รายงานบันทึกขายสินค้า,บริการพร้อมรายการย่อย | `OnBillsysTransrep12` MainFrm.cpp:4919 | `CBLRepArInvoiceTaxDetailView` BLRepArInvoiceTaxDetailView.cpp (538 บรรทัด) | `CBLRepArInvoiceTaxDetailDlg` (IDD_BL_REP_ARINVOICE_DETAIL, rc:5610 ; caption "รายงานใบส่ง/ใบกำกับภาษีพร้อมรายการย่อย") | 34012 | แนวนอน A4 |
| 5347 | รายงานการตัดใบขายสินค้า,บริการ | `OnBillsysAnalyrep22` MainFrm.cpp:5802 | `CBLRepCutArInvoiceTaxView` BLRepCutArInvoiceTaxView.cpp (336 บรรทัด) | `CBLRepCutArInvoiceTaxDlg` (IDD_BL_REP_CUT_ARINVOICE, rc:5581 ; caption "รายงานการตัดใบส่งของ/ใบกำกับภาษี") | 34022 | แนวตั้ง A4 |
| 5349 | รายงานบันทึกขายสินค้า,บริการตามประเภทการขาย | `OnBillsysTransrep25` MainFrm.cpp:5045 | `CBLRepArInvoiceBillTypeView` BLRepArInvoiceBillTypeView.cpp (334 บรรทัด) | `CBLRepArInvoiceBillTypeDlg` (IDD_BL_REP_ARINVOICE_BILLTYPE, rc:5486 ; caption "รายงานสรุปการขายแจกแจงประเภทการขาย") | 34025 | แนวนอน A4 |

- ทุกรายงานอ่านใบขาย `BCArInvoice` (เอกสาร 5328) ; drill-down ตาม MainFrm: ช่อง 0 = 5328 ใบขาย, 1 = 5104 ลูกหนี้, 2 = 5602 รายละเอียดสินค้า (เฉพาะ 5336), 3 = 5618 Serial Number (เฉพาะ 5336) — 5337 / 5347 / 5349 มีแค่ช่อง 0-1
- `BillType` ของใบขาย (ตั้งต้นจาก `DefItemBillType` companyinfo.cpp:111 ; ชื่อใน combo เป็น key `CB_DOCTYPE_1..4` ที่ไทยจริงอยู่ใน DB — ไทยด้านล่างเป็นการอนุมานตามลำดับ): 0 = ขายสด, 1 = ขายเชื่อ, 2 = ขายบริการสด, 3 = ขายบริการเชื่อ (BLRepSaleOrderByPersonSumView.cpp:646-648 จัด (0,2) = สด, (1,3) = เชื่อ)
- `TaxType`: 0 = แยกนอก (`CB_TAXTYPE_1` ไทยจาก comment ARFrmARMasterViewPage2.cpp:128), 1 = ภาษีรวมใน (`CB_TAXTYPE_2` comment :129), 2 = อัตราศูนย์ (key `TAX_ZERO`, ไทยจาก comment BLRepArInvoiceTaxDetailView.cpp ท้ายบรรทัด)

#### 4.1 ช่องเงื่อนไข (เรียงบน-ล่างตาม .rc)

| # | ป้ายไทย (จาก .rc) | 5337 | 5336 | 5347 | 5349 | การทำงาน |
|---|---|---|---|---|---|---|
| 1 | พิมพ์จากวันที่ / ถึงวันที่ | มี | มี | มี | มี | `DocDate BETWEEN` ; "ถึง" ว่าง = `DocDate = จาก` (5337 Dlg:147-158) |
| 2 | พิมพ์จากเลขที่เอกสาร / ถึง (5336 และ 5349 ใช้ป้าย "…เลขที่ใบขาย") | มี | มี | มี | มี | `DocNo BETWEEN` / `=` (5337 Dlg:186-194) |
| 3 | พิมพ์จากลูกหนี้ / ถึง (5349 "พิมพ์จากรหัสลูกค้า / ถึงรหัสลูกค้า") | มี | มี | มี | มี | `ArCode BETWEEN` / `=` (5337 Dlg:168-176) |
| 4 | พิมพ์จากแผนก / ถึง | มี (combo `BCDepartment`) | — | — | — | `DepartCode BETWEEN` / `=` (5337 Dlg:204-212, combo :103-104) ; ป้ายช่อง "ถึง" ใน .rc พิมพ์ผิดเป็น "พิมพ์ถึงเลขที่เอกสาร:" (`IDC_LB_STOP_DEPART`, rc:912) |
| 5 | พิมพ์จากพนักงาน / ถึงพนักงาน | มี (combo `BCSale`) | — | — | — | `SaleCode BETWEEN` / `=` (5337 Dlg:222-230, combo :106-107) |
| 6 | พิมพ์จากสินค้า / ถึงสินค้า | — | มี | — | — | 5336: ต่อ `and docno in (select docno from bcarinvoicesub where itemcode between … )` เพื่อกรอง "เอกสาร" (Dlg:198-208 ; ตัวพิมพ์เล็กทั้งชุด) **และ** กรอง "บรรทัดสินค้า" ต่อท้าย SQL แถวย่อยด้วย `m_strItemCondit` (Dlg:211, View:342) |
| 7 | ประเภทการขาย : (combo) | มี | มี | — | มี | ดูตารางด้านล่าง |
| 8 | [ ] แสดง Serial | — | มี | — | — | 5336 เท่านั้น ; ติ๊ก = เพิ่มระดับที่ 3 แสดง Serial/ทะเบียน/เลขตัวถัง ; ไม่ติ๊ก = 2 ระดับ (View:151-152, :188) |
| ท้ายจอ | กรอบ "เงื่อนไขการพิมพ์รายงานมาตรฐาน", ภาพตัวอย่าง, ตารางเงื่อนไขผู้ใช้ | ใช่ | ใช่ | ใช่ | ใช่ | ดู R0 |

**combo "ประเภทการขาย :" (5337 / 5336 / 5349)** — 5 ตัวเลือก ค่าเริ่มต้นข้อ 4 (5337 Dlg:84-92 ; 5336 Dlg:79-87 ; 5349 Dlg:71-79)

| ลำดับ | key | ไทย | SQL |
|---|---|---|---|
| 0 | `CB_DOCTYPE_1` | ขายสด (อนุมาน) | `AND BillType = 0` |
| 1 | `CB_DOCTYPE_2` | ขายเชื่อ (อนุมาน) | `AND BillType = 1` |
| 2 | `CB_DOCTYPE_3` | ขายบริการสด (อนุมาน) | `AND BillType = 2` |
| 3 | `CB_DOCTYPE_4` | ขายบริการเชื่อ (อนุมาน) | `AND BillType = 3` |
| 4 (ค่าเริ่มต้น) | 5337 และ 5349: `FROM_LIST_TOTAL` (comment ใน dialog อื่น = "พิมพ์จากรายการทั้งหมด") ; 5336: `SELECT_ALL` = เลือกทั้งหมด | ทั้งหมด | ไม่กรอง |

ตัวเลือกที่เลือกถูกพิมพ์ท้ายรายงานเป็นบรรทัด "ประเภทการขาย <ตัวเลือก>" (5337 Dlg:241).

#### 4.2 SQL ตั้งต้นและการเรียง

| เมนู | คำสั่งตั้งต้น | เรียง | หมายเหตุ |
|---|---|---|---|
| 5337 | `SELECT * FROM BCARInvoice WHERE IsCompleteSave = 1` + (`IsPrintCancel==0` → `AND IsCancel = 0`) + (`PosRepSeparate!=1` → `AND PosStatus = 0`) (Dlg:27-41) | `ORDER BY DocDate, DocNo, ArCode` (:279) | **พิมพ์ใบที่ยกเลิกได้** เมื่อ `IsPrintCancel==1` |
| 5336 | `SELECT * FROM BCArInvoice WHERE IsCancel = 0 AND IsCompleteSave = 1` + (`PosRepSeparate!=1` → `AND PosStatus = 0`) (Dlg:26-31) | `ORDER BY DocDate, DocNo, ArCode` (:250) | ไม่พิมพ์ใบยกเลิกเลย ; แถวย่อย `SELECT * FROM BCArInvoiceSub WHERE DocNo='…' AND IsCancel=0 AND TransState=0 AND StockType<>7 AND ItemCode<>'.' [เงื่อนไขสินค้า] ORDER BY ItemCode,UnitCode` (View:341-342) |
| 5347 | `SELECT * FROM BCArInvoice WHERE IsCancel = 0 AND IsCompleteSave = 1` + (`PosRepSeparate!=1` → `AND PosStatus = 0`) (Dlg:24-31) | `ORDER BY DocDate, DocNo, ArCode` (:155) | ไม่พิมพ์ใบยกเลิก |
| 5349 | `SELECT * FROM BCArInvoice WHERE IsCancel = 0 AND IsCompleteSave = 1` + (`PosRepSeparate!=1` → `AND PosStatus = 0`) (Dlg:26-30) | `ORDER BY DocDate, DocNo, ArCode` (:205) | ไม่พิมพ์ใบยกเลิก |

#### 4.3 คอลัมน์ตามลำดับ

**5337 — 1 ใบขาย = 1 แถว, 18 คอลัมน์** (billreparinvoiceview.cpp:164-182 ; ไทยด้านขวาจาก comment บรรทัดเดียวกัน ถ้าไม่มีให้อนุมาน)

| # | key | ความหมาย | ที่มา |
|---|---|---|---|
| 0 | `DOCDATE` | วันที่ | `DocDate` |
| 1 | `DOCNO` | เลขที่เอกสาร ; คลิก → 5328 ; ใบยกเลิกพิมพ์ `เลขที่(ASSERT3)` (:318) | `DocNo` |
| 2 | `TAX_NO` | เลขที่ใบกำกับภาษี | `TaxNo` |
| 3 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | `ArCode` |
| 4 | `AR_NAME` | ชื่อลูกค้า | `GetArName` |
| 5 | `STKPERIODBAL_AMOUNT` | มูลค่าสินค้า | `SumOfItemAmount` |
| 6 | `TOTAL_DISCOUNT_BILL` | ส่วนลดท้ายบิล | `DiscountAmount` |
| 7 | `EARNEST` | ยอดเงินมัดจำ | `SumOfDeposit1` |
| 8 | `BEFORE_AMOUNT` | ยอดก่อนภาษี | `BeforeTaxAmount` |
| 9 | `SUM_TAX` | ยอดภาษี | `TaxAmount` |
| 10 | `SUMTAXAMOUNT` | มูลค่ารวมภาษี | `TotalAmount` |
| 11 | `TAX_ZERO` | ภาษีอัตราศูนย์ | `ZeroTaxAmount` |
| 12 | `TOTAL_ITEM_NOTTAX` | มูลค่ายกเว้นภาษี | `ExceptTaxAmount` |
| 13 | `SUM_OF_TAX` | ภาษีหัก ณ ที่จ่าย | `SumOfWTax` |
| 14 | `NETAMOUNT` | ยอดสุทธิ | `NetDebtAmount` |
| 15 | `BILL_INVBALANCE` | ยอดคงค้าง — แสดงเฉพาะ `BillType` = 1 หรือ 3 และ `PosStatus < 1` (:336) | `BillBalance` |
| 16 | `EMPCODE` | รหัสพนักงาน | `SaleCode` |
| 17 | `EMPNAME` | ชื่อพนักงาน | `OLEDBGetMasterName("BCSale", …)` |

กติกาคำนวณของ 5337: ใบ POS (`PosStatus != 0`) หักส่วนลดท้ายบิลออกจาก `TotalAmount` (View:242-243) ; ใบยกเลิก (ที่พิมพ์เมื่อ `IsPrintCancel==1`) ตั้งยอดทั้งหมดเป็น 0 (:247-257) แต่ **ไม่ตั้ง `DiscountAmount` เป็น 0** ; ทุกยอดคูณ `ExchangeRate` ของเอกสาร เมื่อ `MultiCurrency==1` + รายงานแบบ Home Currency (:261-276) ; ยอดคงค้างไม่รวมใบ POS (:313, :353).

**5336 — 3 ระดับ** (BLRepArInvoiceTaxDetailView.cpp:188-224 เมื่อติ๊ก Serial ; :226-253 เมื่อไม่ติ๊ก — ชุดหัว+รายการเหมือนกัน)

| ระดับ | # | key | ความหมาย | ที่มา / หมายเหตุ |
|---|---|---|---|---|
| หัวใบขาย | 0 | `BILL_DOCNO` | เลขที่ใบขายสินค้า ; คลิก → 5328 | `DocNo` |
| | 1 | `FA_SALEDATE` | วันที่ขาย | `DocDate` |
| | 2 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | |
| | 3 | `AR_NAME` | ชื่อลูกค้า | |
| | 4 | `TAX_TYPE` | ประเภทภาษี — แสดงเฉพาะเมื่อ `TaxNo` ไม่ว่าง ; ค่า = แยกนอก / ภาษีรวมใน / `TAX_ZERO` | `TaxType` (View:323-327) |
| | 5 | `SALETYPE` | ประเภทการขาย (`CB_DOCTYPE_1..4`) | `BillType` |
| | 6 | `SUMITEMAMOUNT` | รวมมูลค่าสินค้า | `SumOfItemAmount` |
| | 7 | `TOTAL_DISCOUNT_BILL` | ส่วนลดท้ายบิล | `DiscountAmount` |
| | 8 | `EARNEST` | ยอดเงินมัดจำ | `SumOfDeposit1` |
| | 9 | `TAXAMOUNT` | ภาษีมูลค่าเพิ่ม | `TaxAmount` |
| | 10 | `SUMTAXAMOUNT` | มูลค่ารวมภาษี | `TotalAmount` |
| | 11 | `TOTAL_ITEM_NOTTAX` | มูลค่ายกเว้นภาษี | `ExceptTaxAmount` |
| | 12 | `TOTAL_NET` | มูลค่าสุทธิ | `NetDebtAmount` |
| รายการสินค้า | 0 | `ITEMCODE` | รหัสสินค้า ; คลิก → 5602 | `ItemCode` |
| | 1 | `ITEMNAME` | ชื่อสินค้า | `GetItemName` |
| | 2 | `UINTCODE` | หน่วยนับ | `UnitCode` |
| | 3 | `SHELF` | ที่เก็บ | `ShelfCode` |
| | 4 | `WAREHOUSE` | คลัง | `WHCode` |
| | 5 | `PERPRICE` | ราคาขาย/หน่วย | `Price` (PricePoint ทศนิยม) |
| | 6 | `ITEMQTY` | จำนวน | `Qty` (QtyPoint) |
| | 7 | `SUM_ITEMAMOUNT` | ยอดรวมสินค้า | `Qty × Price` **คำนวณใหม่ในรายงาน** (:360) |
| | 8 | `DISCOUNTM` | ส่วนลด | `DiscountAmount` |
| | 9 | `AMOUNT` | จำนวนเงิน | `Amount` |
| | 10 | `NETAMOUNT` | ยอดสุทธิ | `NetAmount` |
| Serial (เมื่อติ๊ก) | 0 | `CONTROL_NO` | รหัสควบคุมทะเบียนสินค้า ; คลิก → 5618 | `BCSerialMaster.CtrlNo` |
| | 1 | `REGISTERNO` | หมายเลขทะเบียน | |
| | 2 | `CASENO` | หมายเลขตัวถัง | |
| | 3-4 | `WAREHOUSE`, `SHELF` | คลัง, ที่เก็บ | |
| | 5 | `SERIAL_NO` | หมายเลขเครื่อง | |
| | 6 | `REMARK` | หมายเหตุ | `ItemRemark` |
| | 7 | `STATUS_SERIAL` | สถานะ = `INSTOCK` (0) / (1) / `SOLD` (2) — ไทยของ key อยู่ใน DB | `StockStatus` (:486-538) |

Serial อ่านจาก `BCSerialMaster` join `BCSerialTrans` ที่ `SaveFrom = 4` + เลขที่ใบขาย + รหัสสินค้า + ลำดับบรรทัด (View:499-502).

**5347 — 2 ระดับ** (BLRepCutArInvoiceTaxView.cpp:169-181)

| ระดับ | # | key | ความหมาย |
|---|---|---|---|
| หัวใบขาย | 0 | `BILL_DOCNO` | เลขที่ใบขายสินค้า ; คลิก → 5328 |
| | 1 | `FA_SALEDATE` | วันที่ขาย |
| | 2 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 |
| | 3 | `AR_NAME` | ชื่อลูกค้า |
| | 4 | `TOTAL_NET` | มูลค่าสุทธิ = `NetDebtAmount` (ไม่มีสัญลักษณ์สกุลเงิน/ไม่คูณอัตราแลกเปลี่ยน) |
| แถวตัด (ใบเสร็จ) | 0 | `BEFORE_AMOUNT_CUT` | มูลค่าก่อนตัด = `InvBalance` |
| | 1 | `CUT_NO` | เลขที่ใบตัด (= เลขที่ใบเสร็จ `BCReceiptSub1.DocNo`) |
| | 2 | `CUT_DATE` | วันที่ตัด |
| | 3 | `DESCRIPTION` | คำอธิบาย (`MyDescription`) |
| | 4 | `CUT_PAY_AMOUNT` | ยอดตัดจ่าย = `PayAmount` |
| | 5 | `BALANCE` | ยอดคงเหลือ = `InvBalance − PayAmount` |

แถวตัด: `SELECT * FROM BCReceiptSub1 WHERE InvoiceNo='<ใบขาย>' AND ArCode='<ลูกค้า>' AND IsCancel=0 AND TransState=0 ORDER BY DocNo` (View:221-222) ; **ใบขายที่ยังไม่มีใบเสร็จตัดไม่ถูกพิมพ์** ; แถวย่อยไม่คลิกได้ (`AddRow("", 1)` :259 — โค้ด drill ระดับ 1 ถูกคอมเมนต์ :94-103).

**5349 — 1 ใบขาย = 1 แถว, 9 คอลัมน์** (BLRepArInvoiceBillTypeView.cpp:162-170)

| # | key | ความหมาย | ที่มา |
|---|---|---|---|
| 0 | `DOCNO_DATE` | วันที่ออกเอกสาร | `DocDate` |
| 1 | `BILL_DOCNO` | เลขที่ใบขายสินค้า ; คลิก → 5328 | `DocNo` |
| 2 | `ARCODE` | รหัสลูกค้า ; คลิก → 5104 | |
| 3 | `AR_NAME` | ชื่อลูกค้า | |
| 4 | `REMARK` | หมายเหตุ | `MyDescription` |
| 5 | `RECEIVE_CASH_SALES` | ขายสด (รับเป็นเงินสด) — เฉพาะ `BillType` 0, 2 | `SumCashAmount` |
| 6 | `RECEIVE_OTHER_SALES` | ขายสด (รับอื่น ๆ) — เฉพาะ `BillType` 0, 2 | `SumChqAmount + SumCreditAmount + SumBankAmount + CoupongAmount(เฉพาะใบ POS) + OtherIncome + ExcessAmount1 − (OtherExpense + ExcessAmount2)` (:228) |
| 7 | `SELL` | ขายเชื่อ — เฉพาะ `BillType` 1, 3 | `NetDebtAmount` |
| 8 | `SUMTAXAMOUNT` | รวม = (5)+(6) สำหรับขายสด หรือ (7) สำหรับขายเชื่อ (:229, :233) | |

#### 4.4 ยอดรวม การจัดกลุ่ม และการขึ้นหน้าใหม่

| เมนู | จัดกลุ่ม | ยอดรวมย่อย | ยอดรวมท้ายรายงาน | ยอดรวมต่อหน้า (`IsPrintSumPage==0`) | พิมพ์เงื่อนไข |
|---|---|---|---|---|---|
| 5337 | ไม่มี ; `AddRow("1,3",0)` :340 | — | "ทั้งหมด N รายการ" (`ALL` … `LIST`) คอลัมน์ 5-15 (:363-377) | `OnAddPageSummery` :385-402 คอลัมน์ 5-15 | `TextoutReport(m_saTextOut)` :380 |
| 5336 | ระดับ 0 = ใบขาย `AddRow("0,2",0)` :336 ; ระดับ 1 = สินค้า `AddRow("0",1)` :387 ; ระดับ 2 = Serial `AddRow("0",2)` :531 | ท้ายแต่ละใบ : "ยอดรวม N รายการ" คอลัมน์ 6-10 ของรายการ (:421) | "ยอดรวมทั้งสิ้น N รายการ" คอลัมน์ 6-10 (:441) | `SUM_LIST_PAGE` N `TRANS_TYPE` :450-461 | `bDisplayCondition` → `OnAddReportCondition` (เรียก :444-445 ; นิยาม :472) |
| 5347 | ระดับ 0 = ใบขาย (ตัวหนา `AddRow("0,2",0)` :231) ; ระดับ 1 = แถวตัด | ท้ายแต่ละใบ : ยอดตัดจ่าย + ยอดคงเหลือ (:278) | ยอดตัดจ่าย + ยอดคงเหลือ (:295) | :304-312 | `bDisplayCondition` :298 |
| 5349 | ไม่มี ; `AddRow("1,2",0)` :266 | — | "ยอดรวมทั้งสิ้น N รายการ" คอลัมน์ 5-8 (:278-287) | :296-312 (แสดงเฉพาะคอลัมน์ที่ยอด ≠ 0 :302-305) | `bDisplayCondition` :290 |

#### 4.5 ตารางที่อ่าน

| เมนู | ตาราง (SQL Server) |
|---|---|
| 5337 | `BCArInvoice`, `BCSale` (ชื่อพนักงาน), `BCCurrency` |
| 5336 | `BCArInvoice`, `BCArInvoiceSub`, `BCSerialMaster`, `BCSerialTrans` |
| 5347 | `BCArInvoice`, `BCReceiptSub1` |
| 5349 | `BCArInvoice` |

#### 4.6 ข้อสังเกตที่ BC ห้ามลอก (กลุ่มรายงานขาย)

1. ชื่อเมนูกับ caption ของ dialog ใน .rc ไม่ตรงกันทั้ง 4 รายงาน (เช่น 5337 caption "รายงานใบจ่ายเงินมัดจำ") และ rc ของ 5337 พิมพ์ป้ายช่อง "ถึงแผนก" ผิดเป็น "พิมพ์ถึงเลขที่เอกสาร:" — BC ใช้ชื่อเมนูและป้ายที่ถูกต้อง
2. 5337 ใบขายที่ยกเลิกถูกตั้งยอดเป็น 0 ยกเว้น `DiscountAmount` ซึ่งยังถูกรวมในคอลัมน์ส่วนลดและยอดรวม (View:247-257, :304, :344) — BC ต้องให้เอกสารยกเลิกเป็น 0 ทุกยอด
3. 5336 กรองสินค้าด้วย SQL ตัวพิมพ์เล็ก (`and docno in (select docno from bcarinvoicesub …)`, Dlg:201-207) เพราะ `GetSQLScript` ทำ `MakeUpper` ทั้งคำสั่ง (BaseReportDialog.cpp:257) — ค่ารหัสสินค้าที่ผู้ใช้พิมพ์จะถูกแปลงเป็นตัวพิมพ์ใหญ่ด้วย ; BC ใช้ parameterized query ไม่แปลงค่า
4. 5336 คำนวณ "ยอดรวมสินค้า" ใหม่เป็น `Qty × Price` แทนค่าที่บันทึก (View:360) — ผลต่างจากเอกสารเมื่อมีการปัดราคา; BC ต้องใช้ค่าที่บันทึกในเอกสารและอธิบายสูตรปัดเศษ
5. 5349 ยอด "ขายสด" คือ **ยอดที่รับชำระ** (เงินสด+โอน+เช็ค+บัตร ฯลฯ) ไม่ใช่ยอดขายของใบ — ใบขายสดที่รับไม่ครบ/เกินจะไม่เท่ากับมูลค่าใบ ; BC ควรแสดงมูลค่าใบกับยอดรับแยกกัน
6. 5347 ยอด "คงเหลือ" รวมต่อใบ = ผลรวมของยอดคงเหลือแต่ละใบเสร็จ (View:244, :263, :267) ซึ่งไม่มีความหมายทางบัญชี (ควรเป็นคงเหลือหลังใบเสร็จล่าสุด) ; ใบขายที่ไม่มีใบเสร็จตัดหายจากรายงาน
7. ทุกรายงานต่อ SQL ด้วย `Format("…'%s'…")` และคำนวณเงินด้วย `double` ; 5336/5347/5349 ไม่รองรับสกุลเงินต่างประเทศ (ไม่คูณ `ExchangeRate`) ต่างจาก 5337 — BC ต้องกำหนดสกุลเงินและอัตราแลกเปลี่ยนในรูปแบบเดียวกันทุกรายงาน
8. เงื่อนไข `PosRepSeparate`: 5337 มีเมื่อ `IsPrintCancel` และ POS รวมกันถึง 4 แบบของ SQL (Dlg:26-42) ซ้อนกันเป็นกิ่ง — BC แยกเป็น 2 พารามิเตอร์ตรง ๆ (รวมเอกสารยกเลิก / รวมใบ POS) แทนการสร้าง SQL หลายแบบ

---

### 5. กลุ่มใบรับคืนสินค้า/ลดหนี้ (4 รายงาน) และใบเพิ่มหนี้ (1 รายงาน)

| เมนู | ชื่อไทยตาม menuconfig.xml | handler | view class (ไฟล์) | dialog class | ชื่อรายงาน (key) |
|---|---|---|---|---|---|
| 5340 | รายงานบันทึกใบรับคืนสินค้า,ลดหนี้ | `OnBillsysTransrep16` MainFrm.cpp:4978 | `CBLRepBillCreditNoteView` BLRepBillCreditNoteView.cpp (475 บรรทัด) | `CBLRepBillCreditNoteDlg` (IDD_BL_REP_CREDITNOTE, BCRepBilling.rc:9066 ; caption ใน rc "รายงานใบจ่ายเงินมัดจำ") | 34016 |
| 5338 | รายงานใบรับคืนสินค้า/ลดหนี้พร้อมรายการย่อย | `OnBillsysTransrep14` MainFrm.cpp:4947 | `CBILLRepCreditNoteDetailView` BILLRepCreditNoteDetailView.cpp (551 บรรทัด) | `CBILLRepCreditNoteDetailDlg` (IDD_BILL_REP_CREDITNOTE_DETAIL, rc:1157 ; caption "รายงานใบจ่ายเงินมัดจำ") | 34014 |
| 5339 | รายงานการตัดใบรับคืนสินค้า,ลดหนี้ | `OnBillsysTransrep15` MainFrm.cpp:4963 | `CBILLRepCreditnoteCutOffView` BILLRepCreditnoteCutOffView.cpp (374 บรรทัด) | `CBILLRepCreditnoteCutOffDlg` (IDD_BILL_REP_CREDITNOTE_CUTOFF, rc:1123 ; caption "รายงานใบส่งคืน/ลดหนี้ที่ตัดหมดแล้ว") | 34015 |
| 5350 | รายงานสถานะใบรับคืนสินค้า/ลดหนี้ | `OnBillsysTransrep26` MainFrm.cpp:5057 | `CBLRepCreditNoteStatusView` BLRepCreditNoteStatusView.cpp (366 บรรทัด) | `CBLRepCreditNoteStatusDlg` (IDD_BL_REP_CREDITNOT_STATUS, rc:5518 ; caption "รายงานใบรับคืนสินค้า/ลดหนี้แจกแจกตามประเภท") | 34026 |
| 5342 | รายงานใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)พร้อมรายการย่อย | `OnBillsysTransrep18` MainFrm.cpp:4992 | `CBILLRepDebitNote1DetailView` BILLRepDebitNote1DetailView.cpp (467 บรรทัด) | `CBILLRepDebitNote1DetailDlg` (IDD_BILL_REP_DEBITNOTE1_DETAIL, rc:1188 ; caption "รายงานใบสั่งจองพร้อมรายการย่อย") | 34018 |

- เอกสารต้นทาง: ใบรับคืนสินค้า/ลดหนี้ = ตาราง `BCCreditNote` (เมนู 5330) ; ใบเพิ่มหนี้ = `BCDebitNote1` (เมนู 5331) ; ตารางเชื่อมใบลดหนี้กับใบกำกับเดิม = `BCInvCreditNote` (คอลัมน์ `CreditNoteNo`, `InvoiceNo`, `OldAmount`, `DiffAmount`, `TrueAmount`), ใบเพิ่มหนี้ = `BCInvDebitNote1`
- drill-down ตาม MainFrm: 5340 และ 5350 ตั้งแค่ช่อง 0 = 5330 ; 5338 ช่อง 0 = 5330, 1 = 5328 ใบขาย, 2 = 5602 สินค้า, 3 = 5618 Serial ; 5339 ช่อง 0 = 5330, 1 = 5104 ลูกหนี้, 2 = 5602 ; 5342 ช่อง 0 = 5331, 1 = 5328, 2 = 5602
- ค่า `CauseType` ของใบลดหนี้ (BLRepCreditNoteStatusView.cpp:223-231): 0 = รับคืนสินค้า/ลดหนี้, 1 = ลดหนี้ราคาผิด, 2 = สินค้าแตกเสียหาย (ไทยจาก comment ท้ายบรรทัด)

#### 5.1 ช่องเงื่อนไข (เรียงบน-ล่างตาม .rc)

| # | ป้ายไทย (จาก .rc) | 5340 | 5338 | 5339 | 5350 | 5342 | การทำงาน |
|---|---|---|---|---|---|---|---|
| 1 | พิมพ์จากวันที่ / ถึงวันที่ | มี | มี | มี | มี | มี | `DocDate BETWEEN` ; "ถึง" ว่าง = `DocDate = จาก` (5340 Dlg:172-183) |
| 2 | พิมพ์จากเลขที่เอกสาร / ถึง (5350 ป้าย "พิมพ์จากเลขที่ใบรับคืนสินค้า/ลดหนี้") | มี | มี | มี | มี | มี | `DocNo BETWEEN` / `=` |
| 3 | พิมพ์จากลูกหนี้ / ถึง (5350 "พิมพ์จากรหัสลูกค้า") (5342 "พิมพ์จากลูกค้า") | มี | มี | มี | มี | มี | `ArCode BETWEEN` / `=` |
| 4 | พิมพ์จากรหัสพนักงาน / ถึง | มี (combo `BCSale`, Dlg:79-80) | — | — | — | — | `SaleCode BETWEEN` / `=` (Dlg:230-238) |
| 5 | พิมพ์เฉพาะ : (combo) | — | — | มี | — | — | ดูตารางด้านล่าง |
| 6 | [ ] แสดง Serial | — (ไม่มีใน dialog) | มี | — | — | — | ติ๊ก = เพิ่มระดับ Serial (5338 View:178-216 / ไม่ติ๊ก :220-) |
| ท้ายจอ | กรอบ "เงื่อนไขการพิมพ์รายงานมาตรฐาน", ภาพตัวอย่าง, ตารางเงื่อนไขผู้ใช้ | ใช่ | ใช่ | ใช่ | ใช่ | ใช่ | ดู R0 |

**combo "พิมพ์เฉพาะ :" ของ 5339** — 3 ตัวเลือก ค่าเริ่มต้นข้อ 0 (BILLRepCreditnoteCutOffDlg.cpp:78-85 ; ใช้ :175, :236-252)

| ลำดับ | key → ไทยจาก comment | SQL |
|---|---|---|
| 0 | `FROM_LIST_TOTAL` = จากรายการทั้งหมด | ไม่กรองเพิ่ม |
| 1 | `BILLBALANCE_ZERO` = เฉพาะที่คงค้าง | `AND BillBalance <> 0.0` |
| 2 | `BILLBALANCE_CUTZERO` = เฉพาะที่ตัดหมดแล้ว | `AND BillBalance = 0.0` |

ตัวเลือกถูกพิมพ์ท้ายรายงาน (Dlg:236-250) ; **ชื่อ key กับความหมายสลับกัน** (`…_ZERO` หมายถึงคงค้าง ≠ 0).

#### 5.2 SQL ตั้งต้นและการเรียง

| เมนู | คำสั่งตั้งต้น | เรียง |
|---|---|---|
| 5340 | `SELECT * FROM BCCreditNote WHERE IsCompleteSave = 1` + (`IsPrintCancel==0` → `AND IsCancel = 0`) + (`PosRepSeparate!=1` → `AND IsPos = 0`) (Dlg:16-34 ; comment :16 "รายงานบันทึกใบรับคืนสินค้า,ลดหนี้") | `ORDER BY DocDate, DocNo, ArCode` (:244) |
| 5338 | แบบเดียวกับ 5340 (Dlg:27-42) | `ORDER BY DocDate, DocNo, ArCode` (:233) |
| 5339 | `SELECT * FROM BCCreditNote WHERE IsCompleteSave = 1 AND IsCancel = 0 [AND IsPos = 0] AND ReturnStatus = 0 AND ReturnMoney = 0` (Dlg:28-30) — ความหมาย `ReturnStatus`/`ReturnMoney` ยังไม่ verify (ReturnMoney = รับคืนเป็นเงินสด ตาม comment BLRepCreditNoteStatusView.cpp:220) | `ORDER BY DocDate, DocNo, ArCode` (:256) |
| 5350 | `SELECT * FROM BCCreditNote WHERE IsCompleteSave = 1 AND IsCancel = 0 [AND IsPos = 0]` (Dlg:26-28) | `ORDER BY DocDate,DocNo,ArCode` (:154) |
| 5342 | `SELECT * FROM BCDebitNote1 WHERE IsCompleteSave = 1 [AND IsCancel = 0 เมื่อ IsPrintCancel==0]` (Dlg:29-30) | `ORDER BY DocDate, DocNo, ArCode` (:215) |

#### 5.3 คอลัมน์ตามลำดับ

**5340 — 4 ระดับ (หัวใบรับคืน/ลดหนี้ > ใบกำกับเดิม > สินค้า > Serial ตามข้อมูลที่พิมพ์)** แต่ `DefaultData` ประกาศชุดคอลัมน์ **ระดับเดียว** (BLRepBillCreditNoteView.cpp:157-172) — ระดับ 1-3 ไม่มีหัวคอลัมน์ของตัวเอง (ยังไม่ verify ว่าเฟรมเวิร์กจัดหัวอย่างไร ; สันนิษฐานว่าเก็บไว้ใน layout ผู้ใช้ `BLRepBillCreditNoteView` ซึ่งไม่อยู่ใน repo)

| ระดับ | # | key | ความหมาย | ที่มา |
|---|---|---|---|---|
| หัวใบรับคืน/ลดหนี้ | 0 | `DOCDATE` | วันที่ | `DocDate` |
| | 1 | `DOCNO` | เลขที่เอกสาร ; คลิก → 5330 ; ยกเลิกพิมพ์ `เลขที่(ASSERT3)` (:244) | `DocNo` |
| | 2 | `TAX_NO` | เลขที่ใบกำกับภาษี | `TaxNo` |
| | 3 | `ARNAME` | ชื่อลูกค้า (คอลัมน์ไม่มีรหัสลูกค้า) | `GetArName` |
| | 4 | `REMARK` | หมายเหตุ | `MyDescription` |
| | 5 | `OLD_AMOUNT` | ยอดเดิม | `SumOldAmount` |
| | 6 | `TRUE_AMOUNT` | ยอดที่ถูกต้อง (อนุมานจากชื่อ key) | `SumTrueAmount` |
| | 7 | `EACH_TOTAL` | ยอดผลต่าง | `SumofDiffAmount` |
| | 8 | `SUM_EXCEPT_TAX` | ยอดยกเว้นภาษี | `SumOfExceptTax` |
| | 9 | `NETAMOUNT` | ยอดสุทธิ | `NetDebtAmount` |
| | 10 | `EMPCODE` | รหัสพนักงาน | `SaleCode` |
| | 11 | `EMPNAME` | ชื่อพนักงาน | `BCSale.Name` |
| ใบกำกับเดิมที่ถูกลดหนี้ | — | (ไม่มีหัวใน DefaultData) | `InvoiceNo` เลขที่ใบกำกับเดิม ; `TaxNo` ของใบกำกับ (จาก `BCArInvoice`) ; ยอดเดิม `OldAmount` ; ผลต่าง `DiffAmount` ; ยอดที่ถูกต้อง `TrueAmount` (:272-301) | `BCInvCreditNote` ที่ `CreditNoteNo` = เลขที่ใบลดหนี้ ; คลิก → 5328 |
| รายการสินค้า | — | (ไม่มีหัวใน DefaultData) | รหัสสินค้า, ชื่อสินค้า, คลัง, ที่เก็บ, หน่วยนับ, จำนวนเดิม (`BillQty`), จำนวนที่ลด (`DiscQty`), ราคา, ส่วนลด, จำนวนเงิน (:306-353) | `BCCreditNoteSub WHERE DocNo AND TransState=0 AND IsCancel=0 ORDER BY LineNumber` |
| Serial | — | — | CtrlNo, RegisterNo, CaseNo, WHCode, ShelfCode, SerialNo, ItemRemark, สถานะ (:437-470) ; แสดงเมื่อ `m_bSerialNo` (:360) | |

- `DefaultData` ซ้ำ index 8 และ 9: บรรทัด 166-167 (`BEFORE_AMOUNT` ยอดก่อนภาษี, `SUM_TAX` ยอดภาษี) ถูกบรรทัด 168-169 (`SUM_EXCEPT_TAX`, `NETAMOUNT`) เขียนทับ เพราะ `InsertColumnData` ใช้ `SetAt(nCol)` (BaseReportView.cpp:350-356) — ผลคือ **ไม่มีคอลัมน์ "ยอดก่อนภาษี" และ "ยอดภาษี" บนจอ** และข้อมูลของ `SumOfExceptTax`/`NetDebtAmount` ลงช่อง 8-9 (View:262-266)
- ทุกยอดคูณ `ExchangeRate` เมื่อ `MultiCurrency==1` + Home Currency (:218-232)

**5338 — 4 ระดับ** (BILLRepCreditNoteDetailView.cpp:178-216 เมื่อติ๊ก Serial ; :220-253 เมื่อไม่ติ๊ก)

| ระดับ | คอลัมน์ตามลำดับ (key = ความหมาย) |
|---|---|
| หัวใบ (10 คอลัมน์ — ไม่มีรหัสพนักงาน) | `DOCDATE` วันที่ ; `DOCNO` เลขที่ (คลิก → 5330) ; `TAX_NO` ; `ARNAME` ชื่อลูกค้า ; `REMARK` ; `OLD_AMOUNT` ยอดเดิม ; `TRUE_AMOUNT` ; `EACH_TOTAL` ยอดผลต่าง ; `SUM_EXCEPT_TAX` ยอดยกเว้นภาษี ; `NETAMOUNT` ยอดสุทธิ (index 8-9 ถูกเขียนทับแบบเดียวกับ 5340 :186-189) |
| ใบส่งของ/ใบกำกับเดิม (5 คอลัมน์ :191-196) | `INVOICE_NO` เลขที่ใบส่งของ (คลิก → 5328) ; `TAX_NO` ; `OLD_AMOUNT` ยอดเดิม ; `EACH_TOTAL` ผลต่าง ; `TRUE_AMOUNT` ยอดที่ถูกต้อง |
| รายการสินค้า (10 คอลัมน์ :197-207) | `ITEMCODE` (คลิก → 5602) ; `ITEMNAME` ; `WAREHOUSE` คลัง ; `SHELF` ที่เก็บ ; `UNITCODE` หน่วยนับ ; `ITEMQTY` จำนวน ; `DISC_QTY` จำนวนที่ลด ; `PRICE` ราคาขาย ; `DISCOUNTM` ส่วนลด ; `AMOUNT` จำนวนเงิน |
| Serial (8 คอลัมน์ :208-216) | เหมือน 5336 ระดับ Serial (`CONTROL_NO` … `STATUS_SERIAL`) ; คลิก → 5618 |

แถวหัวพิมพ์ข้อมูลชุดเดียวกับ 5340 แต่ไม่มีช่อง 10-11 (BILLRepCreditNoteDetailView.cpp:322-336).

**5339 — 2 ระดับ** (BILLRepCreditnoteCutOffView.cpp:164-184)

| ระดับ | # | key | ความหมาย | ที่มา |
|---|---|---|---|---|
| หัวใบ | 0 | `DOCDATE` | วันที่ | `DocDate` |
| | 1 | `DOCNO` | เลขที่เอกสาร ; คลิก → 5330 | `DocNo` |
| | 2 | `TAX_NO` | เลขที่ใบกำกับภาษี | `TaxNo` |
| | 3 | `ARECEIVECODE` | รหัสลูกค้า ; คลิก → 5104 | `ArCode` |
| | 4 | `ARNAME` | ชื่อลูกค้า | |
| | 5 | `REMARK` | หมายเหตุ | `MyDescription` |
| | 6 | `SUMTAXAMOUNT` | มูลค่ารวมภาษี | `SumOfTotalTax` |
| | 7 | `SUM_EXCEPT_TAX` | ยอดยกเว้นภาษี | `SumOfExceptTax` |
| | 8 | `NETDEBT_AMOUNT` | ยอดที่ลดหนี้สุทธิ | `NetDebtAmount` |
| | 9 | `BILL_INVBALANCE` | ยอดคงค้าง | `BillBalance` |
| รายการสินค้า | 0 | `REFERANCE_NO` | เลขที่เอกสารอ้างอิง (= เลขที่ใบกำกับเดิม) | `BCCreditNoteSub.InvoiceNo` |
| | 1 | `ITEMCODE` | รหัสสินค้า ; คลิก → 5602 | |
| | 2 | `ITEMNAME` | ชื่อสินค้า | |
| | 3 | `UINTCODE` | หน่วยนับ (`รหัส/ชื่อหน่วย`) | |
| | 4 | `ISSUERET_NUMBER` | จำนวนรับคืน | `DiscQty` |
| | 5 | `DISCOUNTM` | ส่วนลด | `DiscountAmount` |
| | 6 | `AMOUNT` | จำนวนเงิน | `Amount` |
| | 7 | `NETAMOUNT` | ยอดสุทธิ | `NetAmount` |

แถวรายการ: `SELECT InvoiceNo,ItemCode,UnitCode,DiscQty,DiscountAmount,Amount,NetAmount FROM BCCreditNoteSub WHERE DocNo='…' AND TransState=0 [AND IsCancel=0 เมื่อ IsPrintCancel==0]` (View:277-282).

**5350 — 1 ใบ = 1 แถว** (BLRepCreditNoteStatusView.cpp:151-176 ; ใช้ชุด "ปูน/CHM" เมื่อ `m_bChmMode` = TRUE ซึ่งเป็นค่าเริ่มต้นของ build ตาม BC5Account.cpp:253-257)

| # | key (ชุด CHM — ค่าเริ่มต้น :156-164) | ความหมาย | วิธีได้ค่า |
|---|---|---|---|
| 0 | `CREDITNOTE_DATE` | วันที่รับคืน/ลดหนี้ | `DocDate` |
| 1 | `CREDITNOTE_NO` | เลขที่ใบรับคืน/ลดหนี้ ; คลิก → 5330 | `DocNo` |
| 2 | `ARCODE` | รหัสลูกค้า | |
| 3 | `AR_NAME` | ชื่อลูกค้า | |
| 4 | `CREDITNOTE_CASH` | รับคืนเงินสด (รับเป็นเงินสด) — พิมพ์ "YES"/"NO" | `ReturnMoney==1 \|\| ReturnCash==1` |
| 5 | `CREDITNOTE_CREDIT` | รับคืนเครดิต — "YES"/"NO" | `DebtLimitState==1 && IsCNDeposit==0` (คืนวงเงินเครดิต ไม่ใช่ลดหนี้เงินมัดจำ) |
| 6 | `SUMAMOUNT` | ยอดรวม | `NetDebtAmount` + สัญลักษณ์สกุลเงิน |
| 7 | `CREDITNOTE_DEBITNOTE` | รับคืน(Y)/ลดหนี้(N) — "Y/N" เมื่อ `CauseType` 0 หรือ 2 ; "N" เมื่อ `CauseType==1` หรือ `IsCNDeposit==1` ; ข้อความ `CANCEL_LIST` ถ้ายกเลิก | `CauseType`, `IsCNDeposit` (:223-231, :259-260) |

ชุดคอลัมน์ที่ไม่ใช่ CHM (:167-176) มี 9 คอลัมน์ : เพิ่ม `CREDITNOTE_OTHER` รับคืนเงินสด (รับอื่น ๆ) = `DebtLimitState==0 && ReturnMoney==0` ที่ช่อง 5 และเลื่อนที่เหลือไปทางขวา (:264-277).

**5342 — 3 ระดับ** (BILLRepDebitNote1DetailView.cpp:160-192)

| ระดับ | # | key | ความหมาย | ที่มา |
|---|---|---|---|---|
| หัวใบเพิ่มหนี้ | 0 | `DOCDATE` | วันที่ | `DocDate` |
| | 1 | `DOCNO` | เลขที่เอกสาร ; คลิก → 5331 ; ยกเลิกพิมพ์ `เลขที่(ASSERT3)` | `DocNo` |
| | 2 | `TAX_NO` | เลขที่ใบกำกับภาษี | `TaxNo` |
| | 3 | `ARNAME` | ชื่อลูกค้า | |
| | 4 | `ODL_TOTAL` | ยอดใบกำกับเดิม | `SumOldAmount` |
| | 5 | `STKPERIODBAL_AMOUNT` | มูลค่าสินค้า | `SumOfItemAmount` |
| | 6 | `DISCOUNT_BEFORETAX` | มูลค่าส่วนลดก่อนภาษี | `DiscountAmount` |
| | 7 | `SUM_TAX` | ยอดภาษี | `SumOfTaxAmount` |
| | 8 | `SUMTAXAMOUNT` | มูลค่ารวมภาษี | `SumOfTotalTax` |
| | 9 | `SUM_EXCEPT_TAX` | ยอดยกเว้นภาษี | `SumOfExceptTax` |
| | 10 | `ADD_AMOUNT` | ยอดเพิ่มหนี้ (อนุมาน) | `NetDebtAmount` |
| | 11 | `TRUE_AMOUNT` | ยอดที่ถูกต้อง | `SumTrueAmount` |
| ใบกำกับเดิม | 0 | `INVOICE_NO` | เลขที่ใบกำกับเดิม ; คลิก → 5328 | `BCInvDebitNote1.InvoiceNo` |
| | 1 | `TAX_NO` | เลขที่ใบกำกับภาษีของใบกำกับเดิม — ค้นจาก `BCApInvoice` (ตารางเจ้าหนี้ — ดู quirk) | View:310 |
| | 2 | `OLD_DISCOUNT` | ส่วนลดก่อนภาษี (เดิม) | `InvoiceDisc` |
| | 3 | `OLD_AMOUNT` | ยอดเดิม | `OldAmount` |
| | 4 | `SUM_ITEMAMOUNT` | ยอดรวมสินค้า | `SumOfItemAmount` |
| | 5 | `SUM_EXCEPT_TAX` | ยอดยกเว้นภาษี | `ExceptTaxAmount` |
| | 6 | `ADD_AMOUNT` | ยอดเพิ่ม | `DiffAmount` |
| | 7 | `TRUE_AMOUNT` | ยอดที่ถูกต้อง | `TrueAmount` |
| รายการสินค้า | 0-4 | `ITEMCODE` (คลิก → 5602), `ITEMNAME`, `WAREHOUSE`, `SHELF`, `UNITCODE` | | `BCDebitNoteSub1` (View:349) |
| | 5 | `ITEMQTY` | จำนวนเดิม | `BillQty` |
| | 6 | `QTY_INCREASE` | จำนวนที่เพิ่ม | `DiscQty` |
| | 7 | `PRICE` | ราคา | `Price` |
| | 8 | `DISCOUNTM` | ส่วนลด | `DiscountAmount` |
| | 9 | `AMOUNT` | จำนวนเงิน | `Amount` |

#### 5.4 ยอดรวม การจัดกลุ่ม และการขึ้นหน้าใหม่

| เมนู | แถวหัว/ย่อย | ยอดรวมย่อย | ยอดรวมท้ายรายงาน | ยอดรวมต่อหน้า (`IsPrintSumPage==0`) | พิมพ์เงื่อนไข |
|---|---|---|---|---|---|
| 5340 | หัว `AddRow("1",0)` :260 ; ใบกำกับ `("0",1)` :299 ; สินค้า `("0",2)` :353 ; Serial `("0",3)` :469 | ท้ายรายการสินค้าของแต่ละใบ : "ยอดรวม N รายการ" + ส่วนลด + จำนวนเงิน ที่ช่อง 8-9 (:374) | "ยอดรวม N รายการ" ช่อง 5-9 (:387-399) | `SUM_LISTSUB_PAGE` N `LIST` + **ส่วนลดและจำนวนเงินของรายการสินค้า** ที่ช่อง 8-9 (:406-422) — ไม่ใช่ยอดหัวใบ | `TextoutReport` :401 |
| 5338 | หัว `("1",0,…,FW_BOLD)` :336 ; ใบกำกับ :374 ; สินค้า :428 ; Serial :543 | ท้ายสินค้าแต่ละใบ :449 | "ยอดรวม" ช่อง 5-9 (:462-473) | :480-494 (ส่วนลด+จำนวนเงินรายการสินค้า ช่อง 8-9) | `TextoutReport` :475 |
| 5339 | หัว `("1,3",0,…,FW_BOLD)` :268 ; สินค้า `("0",1)` :305 | ท้ายสินค้าแต่ละใบ : "รวม N รายการ" (`SUMDLIST_ITEM`) + จำนวน/ส่วนลด/จำนวนเงิน/สุทธิ (:318-325) | "รวมทั้งสิ้น N รายการ" ช่อง 6-9 (:337-348) | :355-365 (ช่อง 6-9) | `TextoutReport` :350 |
| 5350 | `AddRow("1",0)` :280 | — | "ยอดรวมทั้งสิ้น N รายการ" (`TRANS_TYPE`) ช่อง 7 เท่านั้น (:298) | :308-320 ช่อง 7 เท่านั้น (:319) | `bDisplayCondition` :302 |
| 5342 | หัว `("1",0,…,FW_BOLD)` :288 ; ใบกำกับ `("0",1)` :342 ; สินค้า `("0",2)` :396 | ท้ายสินค้าแต่ละใบ :411 | ช่อง 3-11 (:423-438) | :445-460 | `TextoutReport` :440 |

#### 5.5 ตารางที่อ่าน

| เมนู | ตาราง (SQL Server) |
|---|---|
| 5340 | `BCCreditNote`, `BCInvCreditNote`, `BCCreditNoteSub`, `BCArInvoice` (เลขกำกับภาษี), `BCSale`, `BCWarehouse`, `BCShelf`, `BCItemUnit`, `BCSerialMaster`, `BCSerialTrans` |
| 5338 | `BCCreditNote`, `BCInvCreditNote`, `BCCreditNoteSub`, `BCSerialMaster`, `BCSerialTrans` |
| 5339 | `BCCreditNote`, `BCCreditNoteSub`, `BCItemUnit` |
| 5350 | `BCCreditNote`, `BCInvCreditNote` และ `BCCreditNoteSub` (ใช้ใน `OnReturnCredit` :345-361 ซึ่งผลไม่ถูกนำไปพิมพ์ — ดู quirk) |
| 5342 | `BCDebitNote1`, `BCInvDebitNote1`, `BCDebitNoteSub1`, `BCApInvoice`, `BCWarehouse`, `BCShelf`, `BCItemUnit` |

#### 5.6 ข้อสังเกตที่ BC ห้ามลอก (กลุ่มใบรับคืน/ลดหนี้/เพิ่มหนี้)

1. **คอลัมน์ซ้ำ index แล้วเขียนทับกัน** (5340/5338 บรรทัด 166-169 / 186-189) ทำให้ "ยอดก่อนภาษี" และ "ยอดภาษี" หายจากรายงานทั้งที่เจ้าของบัญชีต้องใช้ — BC แสดงฐานภาษี, ภาษี, ยอดสุทธิ ครบทุกใบ
2. 5340 ประกาศคอลัมน์ระดับเดียวแต่พิมพ์ข้อมูล 4 ระดับ + MainFrm ตั้ง drill-down แค่ช่อง 0 (MainFrm.cpp:4978) ในขณะที่ `OnCallTrans` เรียกช่อง 1-3 (View:77-101) — คลิกระดับใบกำกับ/สินค้า/Serial จึงไม่มีปลายทางที่ถูกตั้ง (ยังไม่ verify พฤติกรรมเมื่อ `GetAt` เกินขนาด) ; BC ต้องผูกปลายทางของทุกระดับให้ครบ
3. ยอดรวมต่อหน้าของ 5340/5338 สะสมเฉพาะส่วนลด/จำนวนเงินของ **รายการสินค้า** แล้วพิมพ์ใต้คอลัมน์ "ยอดยกเว้นภาษี/ยอดสุทธิ" ของหัวใบ (BLRepBillCreditNoteView.cpp:406-422) — ตัวเลขไม่ตรงหัวคอลัมน์ ; BC ใช้ผลรวมของหัวใบทุกครั้ง
4. ยอดรวมท้ายรายงานของ 5340 บวกใบยกเลิกด้วยเมื่อ `IsPrintCancel==1` (เงื่อนไขพิมพ์ใบยกเลิก :244 ; ยอดสะสม :262-266 ไม่มีเงื่อนไข `IsCancel`) ; 5350 ไม่พิมพ์ใบยกเลิกเลยเพราะ SQL ตั้งต้น `IsCancel = 0` แต่โค้ดยังมีกิ่ง `nIsCancel==1` ที่เข้าไม่ถึง (:259, Dlg:27-29)
5. 5350 เรียก `OnReturnCredit` ต่อใบ (2 query ต่อใบ — N+1) แต่ไม่นำค่า `nReturn` มาใช้พิมพ์ (View:206-207, :345-361) ; ยอดรวม/ยอดต่อหน้าถูกใส่ที่ช่อง 7 (:298, :319) ทั้งที่ในชุด CHM ยอดเงินอยู่ช่อง 6 (ช่อง 7 คือ Y/N) — ยอดรวมจึงพิมพ์ใต้คอลัมน์ผิด
6. 5342 ค้นเลขกำกับภาษีของใบกำกับเดิมจาก `BCApInvoice` (ตารางเจ้าหนี้) ทั้งที่ใบกำกับเดิมของใบเพิ่มหนี้ลูกหนี้คือ `BCArInvoice` (BILLRepDebitNote1DetailView.cpp:310 เทียบกับ 5340 BLRepBillCreditNoteView.cpp:279 ที่ใช้ `BCArInvoice`) — คอลัมน์เลขกำกับภาษีจะว่างหรือผิด ; BC ต้องอ่านจากใบกำกับขายเดิมจริง
7. 5339 ตั้งชื่อ key `BILLBALANCE_ZERO` แต่ความหมาย "เฉพาะที่คงค้าง" (`BillBalance <> 0`) และเงื่อนไขตั้งต้น `ReturnStatus = 0 AND ReturnMoney = 0` ซ่อนใบที่คืนเงินสดไว้ตลอด โดยผู้ใช้ไม่เห็นเงื่อนไขนี้บนจอ — BC ต้องแสดงเงื่อนไขกรองทุกข้อให้เห็นและตั้งชื่อตรงความหมาย
8. 5350 เก็บสถานะเป็นข้อความ "YES/NO" และ "Y/N" ฝังในโค้ด (:251-277) แทนที่จะใช้ key ภาษา — ไม่เปลี่ยนตามภาษาที่เลือก ; BC ใช้ key ภาษาตามกฎ i18n
9. caption ของ dialog ใน .rc ไม่ตรงชื่อเมนู 4 จาก 5 รายการ (5340/5338 "รายงานใบจ่ายเงินมัดจำ", 5342 "รายงานใบสั่งจองพร้อมรายการย่อย")

---

### 6. กลุ่มพนักงานขาย การเก็บเงิน และการนำส่งเงิน (4 รายงาน)

| เมนู | ชื่อไทยตาม menuconfig.xml | handler | view class (ไฟล์) | dialog class | ชื่อรายงาน (key) |
|---|---|---|---|---|---|
| 5343 | รายงานยอดขายสุทธิของพนักงานขาย | `OnBillsysTransrep19` MainFrm.cpp:5006 | `CBILLRepSaletotalView` BILLRepSaletotalView.cpp (757 บรรทัด) | `CBILLRepSaletotalDlg` (IDD_BILL_REP_SALETOTAL, BCRepBilling.rc:1217 ; caption "รายงานยอดขายสุทธิของพนักงานขาย") | 34019 (View:27) |
| 5302 | รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย | `OnBILLSysRepNetSaleBySale` MainFrm.cpp:9330 (เปิด `CBLRepSaleBySaleView` ; `CRepSaleBySaleView` ถูกคอมเมนต์ :9335) | `CBLRepSaleBySaleView` BLRepSaleBySaleView.cpp (796 บรรทัด) | `CBLRepSaleBySaleDlg` (IDD_BL_REP_SALE_BY_SALE, rc:233) | 35808 |
| 5686 | รายงานการเก็บเงินของพนักงานขาย | `OnMnuRepBillReceivemoneyBySale` MainFrm.cpp:9932 | `CBillRepReceiveMoneyBySaleView` BillRepReceiveMoneyBySaleView.cpp (401 บรรทัด) | `CBillRepReceiveMoneyBySaleDlg` (IDD_BILL_REP_RECEIVEMONEY_BY_SALE, rc:7587) | 34040 |
| 5362 | รายงานการนำส่งเงินประจำวัน | `OnBillsysTransrep43` MainFrm.cpp:5198 | `CBLRepProfitRecMoneyView` BLRepProfitRecMoneyView.cpp (576 บรรทัด) | `CBLRepProfitRecMoneyDlg` (IDD_BL_REP_PROFIT_DETAIL, rc:8120 ; caption "รายงานการนำส่งเงิน" ต่างจากชื่อเมนู "…ประจำวัน") | 34043 |

- drill-down ตาม MainFrm: 5343 และ 5302 ตั้งช่อง 0 = 5602 รายละเอียดสินค้า (MainFrm.cpp:5006-5017, :9330-9340) ; **5686 ไม่ตั้งปลายทางเลย** (handler MainFrm.cpp:9932 มีแต่ `CreateOrActivateFrame`, view ไม่มี `OnCallTrans`) ; 5362 ช่อง 0 = 5328 ใบขาย, 1 = 5326 ใบสั่งขาย
- รายงาน 5343 / 5302 อ่านตาราง `BCSale` (ทะเบียนพนักงานขาย) เป็นแถวหลัก แล้วไปรวมยอดจาก `BCArInvoice` / `BCArInvoiceSub` ของพนักงานแต่ละคน ; ตัวกรองวันที่/ลูกค้า/เลขที่เอกสาร/สินค้า ไม่ใช่เงื่อนไขของแถวพนักงาน แต่เป็นเงื่อนไขย่อยที่ส่งต่อให้ query รายเอกสารของพนักงานแต่ละคน (`m_strSQLSub1`, Dlg:129, :192, :209)

#### 6.1 ช่องเงื่อนไข (เรียงบน-ล่างตาม .rc)

| # | ป้ายไทย (จาก .rc) | 5343 | 5302 | 5686 | 5362 | การทำงาน |
|---|---|---|---|---|---|---|
| 1 | พิมพ์จากวันที่ / ถึงวันที่ | มี (`BCArInvoice.DocDate`) | มี (`BCArInvoiceSub.DocDate`) | มี (`DocDate`) | มี (`DocDate` ทั้ง 3 ตาราง) | `BETWEEN` ; ถึงว่าง = `=` |
| 2 | พิมพ์จากกลุ่มพนักงาน / ถึง | มี แต่ combo ถูกคอมเมนต์ (Dlg:85-86) | มี (combo `BCUserGroup` Dlg:66-67) | — | — | `UserGroup BETWEEN` / `=` (5343 Dlg:173-178 ; 5302 :180-185) |
| 3 | พิมพ์จากพนักงานขาย / ถึง | มี (combo `BCSale` Dlg:76-77) | มี (Dlg:63-64) | มี (combo `BCSALE` Dlg:80-81) | — | 5343/5302: `WHERE Code BETWEEN` บนตาราง `BCSale` (5343 Dlg:150 ; 5302 :157) ; 5686: `AND SaleCode BETWEEN` (Dlg:160) |
| 4 | พิมพ์จากรหัสลูกค้า / ถึง | มี (`BCArInvoice.ArCode`) | — | มี (ช่องลิงก์ แต่ browse ตาราง `ArTransConfig`) | มี (`ArCode`) | `ArCode BETWEEN` / `=` (5686 Dlg:179) |
| 5 | พิมพ์จากเลขที่เอกสาร / ถึง | มี (`BCArInvoice.DocNo`) | — | — | มี (`DocNo`, condition key `FROM_SALES_NO`) | `DocNo BETWEEN` (5343 Dlg:209) |
| 6 | พิมพ์จากสินค้า / ถึง | — | มี (`BCArInvoiceSub.ItemCode`) | — | — | 5302 Dlg:251-258 |
| 7 | พิมพ์จากจำนวนเงิน / ถึงจำนวนเงิน | — | — | — | มี (ทศนิยมตาม `QtyPoint`) | ใบขาย: `NetDebtAmount + SumOfDeposit1` ; ใบมัดจำ/เงินรับล่วงหน้า: `NetAmount` (Dlg:214-216) |
| 8 | [ ] แสดงตามกลุ่มพนักงาน | มี (`IDC_CK_SORTAR`) | มี | — | — | ติ๊ก = จัดกลุ่มตาม `UserGroup` แล้วตามด้วย `Code` : `ORDER BY UserGroup,Code` (5343 Dlg:223 ; 5302 :269) ; ไม่ติ๊ก = `ORDER BY Code` (5343 :220 ; 5302 :266) |
| ท้ายจอ | กรอบ "เงื่อนไขการพิมพ์รายงานมาตรฐาน", ภาพตัวอย่าง, ตารางเงื่อนไขผู้ใช้ | ใช่ | ใช่ | ใช่ | ใช่ | ดู R0 |

ข้อความป้ายของ 5686 หายจากไฟล์ .rc (ไม่มี text label — มีแต่ control ตามลำดับ: พนักงาน จาก-ถึง, วันที่ จาก-ถึง, ลูกค้า จาก-ถึง) ป้ายที่แสดงจริงจึงอยู่ใน DB `myresource` ยังไม่ verify.

#### 6.2 SQL ตั้งต้นและการเรียง

| เมนู | คำสั่งตั้งต้น | เรียง |
|---|---|---|
| 5343 | `SELECT * FROM BCSale` (Dlg:27) | `ORDER BY Code` หรือ `UserGroup,Code` |
| 5302 | `SELECT * FROM BCSale` (Dlg:18) | เหมือน 5343 |
| 5686 | UNION ของ 4 ตาราง (Dlg:208-212 ; หัว `m_strDefaultSQL = strSQL1 UNION strSQL2` Dlg:40) : `BCArInvoice` (`SaveFrom=1`, `(BillType=0 OR BillType=3) AND IsCompleteSave=1 AND IsCancel=0`), `BCArDeposit` (`SaveFrom=0`), `BCARDepositSpecial` (`SaveFrom=8`), `BCReceipt1` (`SaveFrom=2`, `IsCancel=0 AND IsCompleteSave=1`) | `ORDER BY SaleCode, ArCode, DocDate, DocNo` |
| 5362 | UNION ของ `BCArInvoice` (`IsCancel=0 AND IsCompleteSave=1 AND (BillType=0 OR BillType=2)` + `PosStatus=0` เมื่อ `PosRepSeparate!=1`, Dlg:28-35) + `BCARDeposit` + `BCARDepositSpecial` (`IsCancel=0`) (Dlg:236-264) | `ORDER BY DocDate, DocNo, ArCode` (Dlg:251, :264) |

#### 6.3 คอลัมน์ตามลำดับ

**5343 — 2 ระดับ** (BILLRepSaletotalView.cpp:161-200 ; ไทยของ key จาก comment ท้ายบรรทัด)

| ระดับ | # | key | ความหมาย |
|---|---|---|---|
| หัวพนักงาน (ปกติ) | 0-1 | `Sale_Code`, `SALE_NAME` | รหัส, ชื่อพนักงานขาย |
| หัวกลุ่ม (เมื่อติ๊ก "แสดงตามกลุ่มพนักงาน") | 0-1 | `GROUPCODE`, `ITEM_USERGROUP` | รหัสกลุ่ม, ชื่อกลุ่ม |
| ลูกค้าของพนักงาน | 0 | `ARCODE` | รหัสลูกค้า |
| | 1 | `ARNAME1` | ชื่อลูกค้า |
| | 2 | `DOC_SALE` | เลขที่เอกสารขาย (อนุมาน) |
| | 3 | `SALE_AMOUNT` | มูลค่าขาย |
| | 4 | `NET_SALE_AMOUNT` | มูลค่าขายสุทธิ |
| | 5 | `DIFF_AMOUNT` | มูลค่าลดหนี้ |
| | 6 | `CN_NETAMOUNT` | มูลค่าลดหนี้สุทธิ |
| | 7 | `ADD_AMOUNT` | มูลค่าเพิ่มหนี้ |
| | 8 | `DN_NETAMOUNT` | มูลค่าเพิ่มหนี้สุทธิ |
| | 9 | `ITEMAMOUNT` | มูลค่าคงเหลือ |
| | 10 | `BAL_NETAMOUNT` | มูลค่าคงเหลือสุทธิ |

สูตรของพนักงานแต่ละคน (RunProcess :201-454): อ่าน `SELECT DocNo,DocDate,ArCode,TaxNo,TaxType,TaxRate,ExchangeRate,BeforeTaxAmount,NetDebtAmount, CNAmount=(Sum DiffAmount FROM BCInvCreditNote WHERE InvoiceNo=DocNo AND IsCancel=0 AND BillType=0), DNAmount=(Sum DiffAmount FROM BCInvDebitNote1 … IsCancel=0) FROM BCARInvoice WHERE SaleCode='…' AND IsCompleteSave=1 AND IsCancel=0 [AND PosStatus=0 เมื่อ PosRepSeparate!=1] ORDER BY ArCode,DocNo` ; ยอดสุทธิของลดหนี้/เพิ่มหนี้ = ยอด − ยอด×อัตราภาษี/(100+อัตรา) **เฉพาะเมื่อมี `TaxNo` และ `TaxType==1`** ; ทุกยอด × `ExchangeRate` แล้ว `XRound` ; คงเหลือ = `NetDebtAmount − CN + DN` ; คงเหลือสุทธิ = `BeforeTaxAmount − CNnet + DNnet`.

**5302 — 2 ระดับ** (BLRepSaleBySaleView.cpp:147-184)

| ระดับ | # | key | ความหมาย |
|---|---|---|---|
| หัวพนักงาน | 0-1 | `Sale_Code`, `SALE_NAME` | รหัส/ชื่อพนักงานขาย |
| สินค้าที่พนักงานขาย (ปกติ :152-164) | 0 | `ITEMCODE` | รหัสสินค้า |
| | 1 | `ITEMNAME` | ชื่อสินค้า |
| | 2 | `UNITCODE` | หน่วยนับ |
| | 3 | `SELLQTY` | จำนวนขาย |
| | 4 | `SALE_AMOUNT` | มูลค่าขาย |
| | 5 | `ISSUERETQTY` | จำนวนรับคืน |
| | 6 | `DIFF_AMOUNT` | มูลค่าลดหนี้ |
| | 7 | `BALANCEQTY` | จำนวนคงเหลือ |
| | 8 | `ITEMAMOUNT` | มูลค่าคงเหลือ |
| แบบแสดงตามกลุ่ม (RunProcess2 :497-720 ; คอลัมน์ :167-180) | 0-1 | `Sale_Code`, `SALE_NAME` | พนักงาน |
| | 2 | `SUMITEMAMOUNT` | รวมมูลค่าสินค้า |
| | 3 | `VALUE_DISCOUNT` | ส่วนลด |
| | 4 | `BEFORE_TAX` | ยอดก่อนภาษี |
| | 5 | `SUM_EXCEPT_TAX` | ยอดยกเว้นภาษี |
| | 6 | `TOTAL_NET` | มูลค่าสุทธิ |
| | 7 | `RETURN_VALUE` | มูลค่ารับคืน/ลดหนี้ |
| | 8 | `ADD_AMOUNT` | มูลค่าเพิ่มหนี้ |
| | 9 | `SUMTOTAL_AMOUNT` | ยอดรวม |

แหล่งข้อมูลโหมดปกติ (:195-215): `SELECT … FROM BCArInvoiceSub WHERE StockType<>7 AND TransState=0 AND IsCancel=0 AND ItemCode<>'.' AND SaleCode='…' [AND PosStatus=0 เมื่อ PosRepSeparate!=1] GROUP BY ItemCode,UnitCode,DocNo,PosStatus,ExchangeRate ORDER BY ItemCode,UnitCode` ; ใบ POS (`PosStatus != 0`) ตัดภาษีออกด้วย `ยอด − ยอด×อัตรา/(100+อัตรา)` โดยอ่านอัตราจาก `BCArInvoice.TaxRate` (:296) ; รับคืนอ่านจาก `BCCreditNoteSub (DiscQty, Amount) WHERE TransState=0 AND IsCancel=0 AND ItemCode, UnitCode, InvoiceNo` (:369) ; คงเหลือ = ขาย − รับคืน ; ยอดรวมต่อพนักงานและท้ายรายงาน + ยอดรวมต่อหน้า (:722). โหมดแสดงตามกลุ่มอ่านจาก `BCArInvoice … GROUP BY SaleCode` (`Sum(… × ExchangeRate)` ; `IsCompleteSave=1 AND IsCancel=0`).

**5686 — 4 ระดับ** (BillRepReceiveMoneyBySaleView.cpp:135-161)

| ระดับ | # | key | ความหมาย |
|---|---|---|---|
| พนักงาน | 0-1 | `EMPCODE`, `EMPNAME` | รหัส, ชื่อพนักงาน (ถ้าไม่ระบุ = `NO_SPEC_SALE`) |
| ลูกค้า | 0-1 | `ARECEIVECODE`, `ARNAME` | รหัส, ชื่อลูกค้า (ถ้าไม่ระบุ = `NOT_SPEC_AR`) |
| หัวกลุ่ม 3 ช่อง | — | `INVOICE_DOCLIST` (ใบส่งสินค้า), `RECEIVE_LIST` (รับชำระเงิน), `CRCQ_DOCLIST` (เช็ค/บัตรเครดิต/เงินโอน) | ไทยจาก comment ท้ายบรรทัด |
| รายการ (11) | 0 | `DOCNO_2` | เลขที่ (เอกสารขาย/มัดจำ) |
| | 1 | `AS_DATE` | วันที่ |
| | 2 | `LIMIT_DATE` | วันครบกำหนด |
| | 3 | `AMOUNT` | จำนวนเงิน |
| | 4 | `COUPONG_AMONT` | ยอดคูปอง |
| | 5 | `DATE_PAY` | วันที่รับชำระ |
| | 6 | `"CA/CH/CO/CR"` | ชนิดการชำระ (ข้อความอังกฤษฝังในโค้ด : โค้ดแปลง `PaymentType` 0→CA, 1→CR, 2→CH, 3→CO (BillRepReceiveMoneyBySaleView.cpp:296-305) ; ไทย = เงินสด / บัตรเครดิต / เช็ค / เงินโอน เป็นการอนุมานจากตัวย่อ (ยังไม่ verify)) |
| | 7 | `RECEIPT_TOTAL` | ยอดรับชำระ |
| | 8 | `DOCNO_2` | เลขที่ (ใบเสร็จ) |
| | 9 | `DATE_SHCS` | วันที่ (เช็ค/บัตร) |
| | 10 | `DATE_COUNT` | จำนวนวัน = ค่าสัมบูรณ์ของ (วันที่อ้างอิงหรือวันครบกำหนด − วันที่เอกสาร − 1) |

การชำระอ่านจาก `BCRecMoney WHERE SaveFrom=<แหล่ง> AND DocNo='…'` (ใบ POS `PosStatus` 1/2 ใช้ `SaveFrom=-1`).

**5362 — 1 เอกสาร = 1 แถว, 16 คอลัมน์** (BLRepProfitRecMoneyView.cpp:181-196)

| # | key | ความหมาย |
|---|---|---|
| 0 | `DATEP` | วันที่ |
| 1 | `DOCNO` | เลขที่เอกสาร ; คลิก → 5328 |
| 2 | `NO_REFERANCE` | เลขที่อ้างอิง (`SORefNo` / `RefNo` ; คลิก → 5326) |
| 3 | `AR_NAME` | ชื่อลูกค้า |
| 4 | `STKPERIODBAL_AMOUNT` | มูลค่าสินค้า = `SumOfItemAmount + ZeroTax + ExceptTax` (0 ถ้า `TaxType==2` และ `ZeroTax>0`) |
| 5 | `DISCOUNTM` | ส่วนลด |
| 6 | `DEPOSIT` | หักมัดจำ |
| 7 | `TOTALBILL_AMOUNT` | จำนวนเงินท้ายบิล |
| 8 | `RECEIPT_TOTAL` | ยอดรับชำระ |
| 9 | `BILL_INVBALANCE` | ยอดคงค้าง |
| 10 | `CASH` | เงินสด |
| 11 | `CHEQ` | เช็ค |
| 12 | `CREDIT_CARD` | บัตรเครดิต |
| 13 | `"Charge"` | (ข้อความอังกฤษฝังในโค้ด — ความหมายอนุมาน: ลูกหนี้/ค้างชำระ) |
| 14 | `MONEY_TRANSFER` | เงินโอน |
| 15 | `COUPONG_AMONT` | ยอดคูปอง |

ยอดติดลบถูกตัดเป็น 0 `(x>0)?x*rate:0` ; แยกวิธีชำระด้วย `GetInvoicePayment` (:521-) : ปันยอดตามสัดส่วน (วิธีชำระ × ยอดรับ / ยอดรวมวิธีชำระ) แล้วโยนเศษปัดไปยังวิธีชำระแรกที่เป็นบวก ; กิ่ง `BillType` 1/3 ที่เรียก `GetInvoicePayment` ทำงานไม่ถึงเพราะ SQL เลือกเฉพาะ `BillType` 0 และ 2.

#### 6.4 ยอดรวม การจัดกลุ่ม และการขึ้นหน้าใหม่

| เมนู | จัดกลุ่ม | ยอดรวมย่อย | ยอดรวมท้ายรายงาน | ยอดรวมต่อหน้า | พิมพ์เงื่อนไข |
|---|---|---|---|---|---|
| 5343 | พนักงาน (หรือกลุ่ม → พนักงาน เมื่อติ๊ก) → ลูกค้า | ท้ายแต่ละพนักงาน "SUMDETAIL_WORD N LIST" | "SUMTOTALDETAIL_WORD" | `IsPrintSumPage==0` → :706-740 (ช่องที่ 6-8 ใช้ตัวแปรสลับกัน — ดู quirk) | `OnAddReportCondition` :740 |
| 5302 | พนักงาน → สินค้า | ท้ายแต่ละพนักงาน + ยอดรวม | ยอดรวมท้ายรายงาน | :722 | `OnAddReportCondition` :781 |
| 5686 | พนักงาน (ระดับ 0 :229) → ลูกค้า (ระดับ 1 :233, :268) → รายการ/ยอดรวม (ระดับ 3 :219-396) ; ไม่มีแถวคลิกได้ | `SUM_BY_AR` ("ยอดรวม(ลูกค้า)"), `SUM_BY_SALE` ("ยอดรวม(พนักงาน)") เฉพาะคอลัมน์ 3 และ 7 (:216-224, :381-396) | `TOTAL_AMOUNT` ("รวมทั้งสิ้น") คอลัมน์ 3 และ 7 | ไม่มี (`OnAddPageSummery` ว่าง) | `TextoutReport(m_saRepConditionText)` พิมพ์เสมอ |
| 5362 | ไม่มี ; `AddRow("2",0)` :372 — **คลิกได้เฉพาะคอลัมน์ 2 (เลขที่อ้างอิง → 5326)** ทั้งที่ `OnCallTrans` เขียนรองรับคอลัมน์ 1 (เลขที่เอกสาร → 5328) ด้วย (View:95-104) จึงกดเลขที่ใบขายไม่ได้ | — | แถว "SUMTOTALDETAIL_WORD N TRANS_TYPE" คอลัมน์ 4-15 (:423) | `SUM_LIST_PAGE` N `TRANS_TYPE` (:432-452) | `bDisplayCondition` :426-427 (นิยาม `OnAddReportCondition` :470) |

#### 6.5 ตารางที่อ่าน

| เมนู | ตาราง (SQL Server) |
|---|---|
| 5343 | `BCSale`, `BCArInvoice`, `BCInvCreditNote`, `BCInvDebitNote1` |
| 5302 | `BCSale`, `BCUserGroup`, `BCArInvoice`, `BCArInvoiceSub`, `BCCreditNoteSub`, `BCInvCreditNote`, `BCInvDebitNote1` |
| 5686 | `BCArInvoice`, `BCArDeposit`, `BCARDepositSpecial`, `BCReceipt1`, `BCRecMoney`, `BCSale` |
| 5362 | `BCArInvoice`, `BCARDeposit`, `BCARDepositSpecial`, `BCRecMoney` (ผ่าน `GetInvoicePayment`) |

#### 6.6 ข้อสังเกตที่ BC ห้ามลอก (กลุ่มพนักงานขาย/เก็บเงิน/นำส่งเงิน)

1. **5343/5302 โหมด "แสดงตามกลุ่มพนักงาน"**: `RunProcess` เรียก `RunProcess2` แล้ว **ไม่ return** (BILLRepSaletotalView.cpp:203-205 ; BLRepSaleBySaleView.cpp:188-189) จึงวนต่อในลูปปกติด้วย — พฤติกรรมรวมหลังจากนั้นยังไม่ verify (อาจพิมพ์ซ้ำ/ผิดโครง) ; `RunProcess2` ของ 5343 ประกาศ `bFirst = TRUE` ในทุกรอบ ทำให้หัวกลุ่มพิมพ์ซ้ำทุกพนักงาน
2. 5302 ใช้ชื่อ layout ผู้ใช้ "BILLRepSaletotalByAr" ซ้ำกับ 5343 (BLRepSaleBySaleView.cpp:115) — การปรับหัวคอลัมน์ของรายงานหนึ่งไปทับอีกรายงาน
3. 5302 โหมดกลุ่ม: SQL `FROM BCArInvoiceWHERE` ไม่มีช่องว่างในกิ่ง `PosRepSeparate!=1` (BLRepSaleBySaleView.cpp:593 → query ล้ม), ใช้ตัวกรอง `BCArInvoiceSub.DocDate/ItemCode` กับตาราง header `BCArInvoice` (m_strSQLSub1) และ subquery ลดหนี้/เพิ่มหนี้ไม่มี aggregate และไม่กรองวันที่ (:609, :617) ; ยอดรวม `dTotalAmount = net + CN + DN` (บวกลดหนี้ :624 — เครื่องหมายยังไม่ verify)
4. 5343 ยอดรวมต่อหน้ามีตัวแปรสลับช่อง (ช่อง 6 ใช้ `m_dPageSumCNAmount`, ช่อง 7 ใช้ CNNet, ช่อง 8 ใช้ DNNet — View:706-735) ทำให้ตัวเลขไม่ตรงหัวคอลัมน์
5. 5686 ยอดใบส่งสินค้าถูกบวกซ้ำต่อแถว **การชำระ** (`dGroupBillAmount/dSubGroupBillAmount/dTotalBillAmount` สะสมที่ BillRepReceiveMoneyBySaleView.cpp:343-345) — ใบเดียวที่ชำระหลายวิธีจะนับยอดหลายรอบ ; ยอดรวมกลุ่มบวก `dNetDebtAmount` (:347) แต่ยอดย่อย/รวมบวก `dPayAmount` (:348-349) — ไม่สอดคล้องกัน
6. 5686 รวม `BillType 0 OR 3` (ขายสด + ขายบริการเชื่อ) ซึ่งไม่ตรงข้อตกลง (0,2) = สด, (1,3) = เชื่อ ที่ใช้ในรายงานอื่น ; ผูกตาราง user-define ผิดเป็น `BCQuotation/BCQuotationSub` (:90-91) ; ช่องลูกค้าเปิด browse ตาราง `ArTransConfig` (ผิดตาราง)
7. 5362 ช่องกรอง "จำนวนเงิน" ส่งค่าด้วย `%lf` (float) และกิ่ง "เท่ากับ" อ้างตารางผิด `BCARDeposit.NetDebtAmount + BCArInvoice.SumOfDeposit1` (BLRepProfitRecMoneyDlg.cpp:222 ; กิ่งช่วงใช้ `BCArInvoice.NetDebtAmount` :214) ; ข้อความ "Charge" ฝังภาษาอังกฤษ ; รายงานเรียกชื่อ "นำส่งเงินประจำวัน" แต่ SQL ไม่กรองเฉพาะวันนั้น (กรองตามช่วงวันที่ที่ผู้ใช้ระบุ)
8. ทุกรายงานในกลุ่มต่อ SQL ด้วยสตริง (`Format("…'%s'…")`) และคำนวณเงินด้วย `double` ; 5343/5302 ตัดภาษีออกด้วยสูตร `ยอด − ยอด×อัตรา/(100+อัตรา)` ซึ่งปัดเศษทีละบรรทัด — BC ต้องคำนวณฐานภาษีจากค่าที่บันทึกในเอกสาร ไม่คำนวณย้อนกลับ
9. ยอดขายสุทธิของพนักงานเป็นข้อมูลที่ผูกกับตัวบุคคล (ฐานคอมมิชชัน/ประเมินผล) — BC ต้องกำหนดสิทธิ์ดูรายงานรายพนักงาน (ผู้ใช้ที่ไม่ใช่เจ้าของเห็นเฉพาะของตนเอง) ; โค้ดต้นแบบมีส่วนจำกัดตามพนักงานที่ล็อกอินแต่ถูกคอมเมนต์ทิ้ง (BILLRepSaletotalDlg.cpp:227-285)

---

### 7. สรุปสิ่งที่ "ยังไม่ verify" และ "อนุมาน" ของทั้งไฟล์

| # | เรื่อง | สถานะ | อยู่ที่ |
|---|---|---|---|
| 1 | ไทยของ key ภาษาที่อยู่ใน DB `myresource` (ไม่อยู่ในรีโป) เช่น `BILLDOCTYPE`, `SALE_ID`, `PO_SALE`, `PO_ORDER`, `BALANCE_NUMBER`, `CB_DOCTYPE_1..4`, `ADD_AMOUNT`, `ASSERT3` | ใช้ไทยจาก comment ท้ายบรรทัดโค้ด / ป้ายใน `.rc` ; ที่ไม่มีที่มาเขียนว่า "(อนุมาน)" ในตารางคอลัมน์ ; ข้อความจริงในฐานข้อมูล ยังไม่ verify | ตารางคอลัมน์หัวข้อ 1-6, R0 |
| 2 | ป้ายของช่องเงื่อนไข 5686 | ไฟล์ `.rc` ไม่มีข้อความป้าย (มีแต่ลำดับ control) ; ป้ายจริงอยู่ใน `myresource` — ยังไม่ verify | 6.1 |
| 3 | วิธีตัดหน้า/จัดหัวคอลัมน์ของเอนจินพิมพ์ใน `CBaseReportView` และรูปแบบไฟล์ส่งออก (`OnExportTo`) | ไม่พบโค้ดตัดหน้าตามกลุ่มในตัวรายงาน ; ส่วนของเอนจินยังไม่ได้อ่านละเอียด | R0, 1.4 |
| 4 | หน้าจอ Report Center ที่เปิดรายงานเหล่านี้ | ไม่ได้เปิดอ่านในไฟล์นี้ | R0 |
| 5 | 5354: คอลัมน์ `DocDate`/`ArCode` มีอยู่จริงในตารางรายการย่อยของใบเสนอราคาหรือไม่ | ไม่มี DDL ใน `D:\project-champ` ที่ตรวจในรอบนี้ | 1.2 |
| 6 | 5334/5355: ฐานยอดหัวเอกสาร (`BeforeTaxAmount`) กับแถวตัด (`NetAmount`) ตรงกันเมื่อ `TaxType == 0` หรือไม่ | ยังไม่ verify | 3.6 ข้อ 9 |
| 7 | 5339: ความหมายของ `ReturnStatus` / `ReturnMoney` (comment ระบุแค่ว่า ReturnMoney = รับคืนเป็นเงินสด) | ยังไม่ verify เต็ม | 5.2 |
| 8 | 5340: เฟรมเวิร์กจัดหัวคอลัมน์ระดับ 1-3 อย่างไร เมื่อ `DefaultData` ประกาศระดับเดียว และพฤติกรรม `GetAt` เกินขนาดเมื่อคลิกระดับที่ไม่ได้ลงทะเบียนปลายทาง | ยังไม่ verify | 5.3, 5.6 ข้อ 2 |
| 9 | 5343/5302 โหมด "แสดงตามกลุ่มพนักงาน": `RunProcess` เรียก `RunProcess2` แล้วไม่ return — ผลลัพธ์รวมหลังจากนั้น (อาจพิมพ์ซ้ำ) | ยังไม่ verify (ไม่ได้รันโปรแกรมจริง) | 6.6 ข้อ 1 |
| 10 | ความหมายไทยของ "CR / CO" ใน `PaymentType` (บัตรเครดิต / เงินโอน) | อนุมานจากตัวย่อ | R0, 6.3 |
| 11 | ไฟล์ `.cpp` ที่ใช้ตัวอย่างผ่าน mirror ถอดรหัสแล้ว (`dec-sa`) — เลขบรรทัดนับบนไฟล์ถอดรหัส ซึ่งถอดจากรหัสไทยเดิมแบบ 1 บรรทัดต่อ 1 บรรทัด | ถ้าเปิดไฟล์ต้นฉบับแล้วเลขเลื่อน ให้ค้นด้วยชื่อฟังก์ชัน | ทั้งไฟล์ |

---

---

## ส่วน 7 — รายงานขายเชิงวิเคราะห์ (Report Center)

> ขอบเขต: 21 รายการเมนู ตามรายการใน `atlas-10-รายงาน.md:220-282` และ `menuconfig.xml:705-733` (`D:\project-champ\champ\champ\menuconfigxml\menuconfig.xml`) ; ทุกรายการเป็น popup `disable="yes"` (ไม่ขึ้นแถบเมนู เปิดจาก Report Center — ดูหมายเหตุภาพรวม ไม่ซ้ำที่นี่) ; ชื่อไทยของรายงานคัดจากข้อความ resourceid ใน `menuconfig.xml` ตามแอตลาส ; ป้ายช่องกรอกที่เป็นไทยแท้ยกจาก `BCRepBilling.rc` (รูปแบบ `"KEY$ไทย"`) หรือคอมเมนต์ไทยในโค้ด ; ป้ายที่ไทยหายไปใน rc (เหลือแต่ `KEY$`) หรือเป็น key ที่อยู่ใน DB `myresource` (ไม่อยู่ใน repo) ระบุ key และความหมายแบบ "(อนุมาน)".
> ไฟล์ต้นทางย่อ: `rep-bill\*` = `D:\project-champ\champ-report\BC5REP_BILL\` ; `MainFrm.cpp` = `D:\project-champ\champ\champ\MainFrm.cpp` ; `BCRepBilling.rc` = `rep-bill\BCRepBilling.rc` ; `BaseReportView.cpp` = `lib-report\BaseReportView.cpp` (เลขบรรทัดตรงกับต้นฉบับ).

### R0. กติการ่วมของรายงานกลุ่มนี้ (ใช้กับทุกรายงานด้านล่าง)

| หัวข้อ | พฤติกรรมใน Champ | หลักฐาน |
|---|---|---|
| โครงสร้างโปรแกรม | 1 รายงาน = `*View.cpp` (`DefaultData` กำหนดคอลัมน์ด้วย `InsertColumnData`/`AddColumn`, `RunProcess` ประมวลผลและพิมพ์แถว, `OnAddPageSummery` รวมท้ายหน้า, `OnCallTrans` คลิกทะลุ, `OnFilterReport` เปิดกล่องเงื่อนไข) + `*Dlg.cpp` (`OnInitDialog` โหลด combo, `GetReportCondition` ต่อ SQL และข้อความเงื่อนไข) ; `m_ReportName` เป็นเลข key หา Thai จาก `myresource` | `BLRepBCArInvoiceView.cpp:177-203,234-258`, `BLRepBCArInvoiceDlg.cpp:117` |
| `AddColumn(nSpace)` / `AddRow` | `AddColumn()` ปิดชุดคอลัมน์ 1 ชุด (ชุดแรก = แถวหัวกลุ่ม `nColumnRow=0`, ชุดที่สอง = แถวรายการ `nColumnRow=1`) ; `AddRow(strColCanClick, nColumnRow, bSummary, weight)` เพิ่มแถว โดย `strColCanClick` = คอลัมน์ที่คลิกทะลุได้ (เช่น `"0,1,3"`), `bSummary=TRUE` = แถวสรุป (ตัวหนา) ; การขึ้นหน้าใหม่ด้วยมือคือ `bEndPageOnThisRow` — **ไม่พบการสั่งขึ้นหน้าใหม่ในไฟล์ view ใดของ 21 รายงานนี้** (grep `EndPage/NewPage/PageBreak` = 0 ไฟล์) | `BaseReportView.cpp:376,615-625` |
| เงื่อนไขช่วง "จาก–ถึง" | ใส่เฉพาะ "จาก" = เท่ากับ (`= '...'`) ; ใส่ทั้งสอง = `BETWEEN` ; ว่างทั้งคู่ = ไม่กรอง ; ข้อความเงื่อนไขที่พิมพ์ท้ายรายงานต่อสตริงด้วยตัวคั่น `@` | `BLRepBCArInvoiceDlg.cpp:158-166,193,223-253` |
| ตัวกรองมาตรฐานของเอกสารขาย | `IsCancel = 0 AND TransState = 0 AND ItemCode <> '.' AND StockType <> 7` (ตัดแถวยกเลิก, แถวที่ถูกตัดต่อ, แถวข้อความ `.`, แถวสินค้าประเภท 7) ; `StockType` ตามคอมโบสินค้า: 0 ทั่วไป, 1 บริการ, 2 มี Serial, 3 เช่า, 4 รับฝากขาย, 5 สินค้าชุด, 6 พิเศษ(บุหรี่), 7 สินค้าชุดแบบที่ 2 (ลำดับนี้ถ้าบริษัทไม่ถอด Serial ออก — ยังไม่ verify ว่าลำดับเลื่อนเมื่อถอด) | `BLRepBCArInvoiceDlg.cpp:27-33`, `icdlgitemmasterpage1.cpp:203-211` |
| "รายงานไม่แยกหลังร้าน" | ถ้า `m_CompanyInfo.PosRepSeparate == 1` ไม่กรอง `PosStatus` ; ถ้าไม่ใช่ เติม `AND PosStatus = 0` (ตัดบิลหลังร้าน/POS) | `BLRepBCArInvoiceDlg.cpp:27-33`, `RepDailySaleDepartment_MonthYear_View.cpp:448-450` |
| ทศนิยม | จำนวน = `QtyPoint`, จำนวนเงิน = `AmountPoint`, ต้นทุน = `CostPoint`, ราคา/หน่วย = `PricePoint` (ตัวเลือกบริษัท) ; แสดงด้วย `MoneyType(...)` หรือ `XMoneyFormat(v,"-","",จุดทศนิยม)` | `BLRepBCArInvoiceView.cpp:518-522`, `blrepqoutationsubdateview.cpp:711-716` |
| หน่วยนับหลายหน่วย | สินค้าที่ `BCItem.UnitType == 1` แปลงจำนวนเป็นหน่วยเก็บ (`DefStkUnitCode`) ด้วย `BCStkPacking.Rate1/Rate2` ; ตัวคูณขึ้นกับรายงาน (5184: `Qty*PackingRate1/PackingRate2`, 5271: `qty*(Rate2/Rate1)`) ; คอลัมน์หน่วยนับแสดง `รหัส/ชื่อหน่วย` | `BLRepBCArInvoiceView.cpp:337-340`, `BL_Rep_Compare12MonthView.cpp:531-553` |
| หลายสกุลเงิน | ถ้า `MultiCurrency == 1` และตัวเลือกรายงาน `bHomeCurrencyReport == 1` ใช้ `ExchangeRate` (ถ้า <= 0 ใช้ 1.0) ; ใช้เฉพาะในรายงานที่อ่านอัตราแลกเปลี่ยนรายแถว (5271) | `BL_Rep_Compare12MonthView.cpp:400-409` |
| รวมท้ายหน้า/ท้ายรายงาน | แถว "SUM_LIST_PAGE n LIST" รวมต่อหน้า พิมพ์เมื่อ `IsPrintSumPage == 0` ; แถวรวมย่อยต่อกลุ่ม "SUMDETAIL_WORD n ..." ; แถวรวมทั้งสิ้น "SUMTOTALDETAIL_WORD n ..." ; ท้ายสุดพิมพ์ข้อความเงื่อนไขที่เลือกเมื่อ `m_RepOption.bDisplayCondition` | `blrepqoutationsubdateview.cpp:762,1286-1292`, `BLRepBCArInvoiceView.cpp:523-524` |
| ข้อความเตือนก่อนรัน | รายงานรายวัน/เดือน (5290-5296) บังคับกรอกวันที่ผ่าน `IsStartReport()` : key `INSERT_DATE` (คอมเมนต์ไทย "โปรดกรอกวันที่ด้วย") ; รายงานอื่นไม่บังคับ | `RepDailySaleDepartment_Day_Dlg.cpp:160-168` |
| ผู้ใช้ออกแบบคอลัมน์เอง | `LoadUserDefine()` / `AddTableForUserDefine("...")` — เก็บตาม `m_strUserDefineKeyName` ; ชื่อตารางที่ลงทะเบียนในบางรายงานเป็นของ copy มาผิด (รุ่นแสดงรายการย่อยของ 5291/5296/5294 ลง `BCAPInvoice`/`BCAPInvoiceSub` ทั้งที่ SQL อ่าน `BCARInvoiceSub`) | `BLRepBCArInvoiceView.cpp:210-214`, `RepDailySaleGroupCode_Day_View.cpp:116-117`, `RepDailySaleGroupCode_DayYearDetail_View.cpp:123-124`, `RepDailySaleGroupCode_MonthYearDetail_View.cpp:132-133` |

ข้อสังเกตที่ BC **ห้ามลอก** (พบใน Champ แต่ขัดกฎของโปรเจกต์): (1) คำนวณยอดด้วย `double` และฝังตัวเลขใน SQL ด้วย `%lf` (เช่น `blrepqoutationsubdatedlg.cpp:441,459`) — BC ต้องใช้ `NUMERIC(p,s)`/decimal และพารามิเตอร์ ; (2) สร้าง SQL ด้วย `Format()` ต่อสตริงจากค่าที่ผู้ใช้กรอก (`BLRepBCArInvoiceDlg.cpp:140-150`) — BC ต้องใช้ parameterized SQL ; (3) มีคำสั่ง `UPDATE` แก้ข้อมูลเอกสารอยู่ในกล่องเงื่อนไขของรายงาน (`BLRepCreditNoteDateDlg.cpp:141-143`, `BL_Rep_Compare12MonthDlg.cpp:139-146`) — รายงานของ BC ต้องอ่านอย่างเดียว ; (4) ยอดรวม/สัดส่วนคำนวณในหน่วยความจำฝั่งโปรแกรมแล้วค่อยแสดง — BC ต้องคำนวณที่ backend (กฎ backend-first) ; (5) ข้อความไทย/วันที่ฝังในโค้ด (`"01/01/%s"`, วันสิ้นเดือนเขียนมือ) — BC ใช้ key ภาษา + วันที่ ISO ; (6) หัวคอลัมน์อังกฤษที่ไม่ผ่าน key (`Margin Percent(%)`) — BC ต้องมี key ภาษา.

---

### กลุ่ม A. รายงานวิเคราะห์ใบเสนอราคา / ใบสั่งขาย แบบรายการย่อย (5180, 5185)

| menu id | ชื่อ (Champ) | handler (`MainFrm.cpp`) | view class (ไฟล์) | dialog (IDD : rc) | report name key |
|---|---|---|---|---|---|
| 5180 | รายงานวิเคราะห์ใบเสนอราคาสินค้า | `OnBillsysAnalyrep11` :5748 (`ID_BILLSYS_ANALYREP_1_1`) | `CBLRepQoutationsubDateView` (`blrepqoutationsubdateview.cpp`, 1344 บรรทัด, key `BLRepQoutationsubDateView`, A4 แนวนอน) | `IDD_BL_REP_QOUTATIONSUB_DATE` 518x348 : `BCRepBilling.rc:1543` caption "รายงานใบเสนอราคาสินค้า-ตามวันที่" | 35100 |
| 5185 | รายงานวิเคราะห์ใบสั่งขาย/จองสินค้าพร้อมรายการย่อย | `OnBillsysAnalyrep31` :5815 (`ID_BILLSYS_ANALYREP_3_1`) | `CBLRepSaleOrderSubDateView` (`blrepsaleordersubdateview.cpp`, 920 บรรทัด, key `BLRepSaleOrderSubDateView`, A4 แนวนอน) | `IDD_BL_REP_SALEORDERSUB_DATE` 505x348 : `BCRepBilling.rc:1936` caption "รายงานใบสั่งขายสินค้า-ตามวันที่" | 35300 |

โครงสร้างเหมือนกัน (อ่านรายการย่อย `BCQuotationSub` / `BCSaleOrderSub` ทีละบรรทัด) ต่างกันที่ตาราง, ช่อง "เลือกเอกสาร" (เฉพาะ 5185), ลำดับ ORDER BY ของบางโหมด และคอลัมน์รายการ.

#### A1. ช่องเงื่อนไข (ลำดับบน -> ล่างตาม rc)

| # | ป้าย (ไทยแท้จาก rc) | control | ชนิด | เงื่อนไข SQL / หมายเหตุ | หลักฐาน (5180) |
|---|---|---|---|---|---|
| 1 | เรียงตาม: | `IDC_ED_START_DATE2` | combo 11 ค่า (ค่าเริ่มต้นข้อแรก) | วันที่เอกสาร / รหัสลูกค้า / รหัสกลุ่มลูกค้า / รหัสแผนก / รหัสโครงการ / `Category` (กลุ่มสินค้า) / `Sub - Category` (ชนิดสินค้า) / `Class` (หมวดสินค้า) / `Sub - Class` (รูปแบบสินค้า) / รหัสยี่ห้อ / รหัสสินค้า ; ค่าที่เลือก = `m_nSortBy` กำหนดหัวกลุ่ม คอลัมน์ และ ORDER BY (ดู A2) ; ป้าย 4 ตัวกลางเป็นอังกฤษในโค้ด (คอมเมนต์ไทยกำกับ) | `blrepqoutationsubdatedlg.cpp:98-110` |
| 2 | พิมพ์จากวันที่: / ถึง: | `IDC_ED_START_DATE`/`STOP_DATE` | วันที่ | `DocDate BETWEEN` (วันที่ผ่าน `XPrepareDate`) | `...dlg.cpp:225` |
| 3 | พิมพ์จากลูกหนี้: / ถึง: | `IDC_ED_START_ARCODE`/`STOP_ARCODE` | lookup ลูกค้า | `ArCode BETWEEN` | `...dlg.cpp:243` |
| 4 | พิมพ์จากกลุ่มลูกค้า: / ถึง: | `IDC_ED_START_ARGROUP`/`STOP_ARGROUP` | combo `BCARGroup` | `(Select GroupCode from BCAR where Code = BCQuotationSub.ArCode) BETWEEN` | `...dlg.cpp:113-114,263` |
| 5 | พิมพ์จากแผนก: / ถึง: | `IDC_ED_START_DEPART`/`STOP_DEPART` | combo `BCDepartment` | `DepartCode BETWEEN` | `...dlg.cpp:116-117,283` |
| 6 | พิมพ์จากโครงการ: / ถึง: | `IDC_ED_START_DEPART2`/`STOP_PROJECT` | combo `BCProject` | `ProjectCode BETWEEN` | `...dlg.cpp:119-120,303` |
| 7 | [ ] แสดงจำนวนใบเสนอราคาตามวันที่ที่กำหนด | `IDC_CHECK_SHOWCOUNT` | checkbox | dialog สร้าง `SELECT COUNT(DocNo) ... FROM BCQuotation` เก็บใน `m_strSQLQuo` แต่ view รับค่าแล้ว**ไม่ได้นำไปพิมพ์** (ติ๊กแล้วไม่มีผลต่อรายงาน) | `...dlg.cpp:490-494`, `...view.cpp:241-242` |
| 8 | Category : / ถึง: | `IDC_ED_START_ITEMGROUP`/`STOP_ITEMGROUP` | combo `BCItemGroup` | `GroupCode BETWEEN` | `...dlg.cpp:122-123,324` |
| 9 | Sub - Category : / ถึง: | `IDC_ED_START_TYPECODE`/`STOP_TYPECODE` | combo `BCItemType` | `TypeCode BETWEEN '<จาก>' AND '<จาก>'` — **บั๊กใน Champ**: ค่า "ถึง" ถูกแทนด้วยค่า "จาก" ช่วงเหลือค่าเดียว | `...dlg.cpp:125-126,346` |
| 10 | Class : / ถึง: | `IDC_ED_START_ITEMCATAGORY`/`STOP_ITEMCATAGORY` | combo `BCItemCategory` | `CategoryCode BETWEEN` | `...dlg.cpp:128-129,363` |
| 11 | Sub - Class : / ถึง: | `IDC_ED_START_FORMAT2`/`STOP_FORMAT2` | combo `BCItemFormat` | `FormatCode BETWEEN` | `...dlg.cpp:134-135,404` |
| 12 | พิมพ์จากยี่ห้อ: / ถึง: | `IDC_ED_START_ITEMBRAND`/`STOP_ITEMBRAND` | combo `BCItemBrand` | `BrandCode BETWEEN` | `...dlg.cpp:131-132,383` |
| 13 | พิมพ์จากสินค้า: / ถึง: | `IDC_ED_START_ITEM`/`STOP_ITEM` | lookup สินค้า | `ItemCode BETWEEN` | `...dlg.cpp:420` |
| 14 | พิมพ์จากจำนวน: / พิมพ์ถึงจำนวน: / พิมพ์จากเงิน: / พิมพ์ถึงเงิน: | `IDC_ED_START_QTY`,`STOP_QTY`,`START_AMOUNT`,`STOP_AMOUNT` | ตัวเลข | มีช่องใน rc แต่โค้ดเงื่อนไข `Qty/Amount BETWEEN %lf` อยู่ในคอมเมนต์ `/* ... */` — **กรอกแล้วไม่มีผล** | `...dlg.cpp:435-467` |
| 15 (เฉพาะ 5185) | เลือกเอกสาร | `IDC_ED_DOCTYPE` | combo 3 ค่า (ค่าเริ่มต้นข้อแรก) | 0 `SALES_ORDER` (คอมเมนต์ไทย "เฉพาะใบสั่งขายสินค้า") -> `SOStatus = 0` ; 1 `SO_RESERVATION` ("เฉพาะใบสั่งจองสินค้า") -> `SOStatus = 1` ; 2 `SUM_DOCNO` ("รวมเอกสาร") -> ไม่กรอง | `blrepsaleordersubdatedlg.cpp:115-117,470-486` |

SQL ตั้งต้น: 5180 `SELECT BCQuotationSub.*, (DueDate), (TaxRate), (ArGroup จาก BCAR) FROM BCQuotationSub WHERE IsCancel = 0 AND TransState = 0` (`blrepqoutationsubdatedlg.cpp:27-28`) ; 5185 เหมือนกันแต่ตาราง `BCSaleOrderSub` (`blrepsaleordersubdatedlg.cpp:27-28`) ; 5185 ช่องจำนวน/เงินไม่มีผลเช่นกัน (`blrepsaleordersubdatedlg.cpp:439,455` อยู่ในคอมเมนต์) และมีบั๊ก `TypeCode` เดียวกัน (`blrepsaleordersubdatedlg.cpp:347`).

#### A2. ตัวเลือก "เรียงตาม" เปลี่ยนอะไร

ORDER BY: 5180 `blrepqoutationsubdatedlg.cpp:476-486` ; 5185 ใส่ `DocNo` ก่อน `DocDate` ในโหมด 1-5 (`blrepsaleordersubdatedlg.cpp:497-501`) ; ท้ายทุกโหมดตามด้วย `LineNumber,ItemCode` (โหมดสินค้า: `ItemCode,DocDate,DocNo,LineNumber`). `RunProcess` แยก 4 ฟังก์ชัน: โหมด 0 = `RunProcess`, 1 = `RunProcess2`, 2-9 = `RunProcess3`, 10 = `RunProcess4` (`blrepqoutationsubdateview.cpp:588-600`).

| โหมด | จัดกลุ่มตาม | หัวกลุ่ม (แถวระดับ 0) | แถวรายการ 5180 (ระดับ 1) | คลิกทะลุ (คอลัมน์) | หลักฐาน (5180 view) |
|---|---|---|---|---|---|
| 0 | วันที่เอกสาร `DocDate` | `DOCDATE` | `DOCNO, ARCODE, ARNAME, ITEMCODE, ITEMNAME, UNITCODE, PERPRICE, QUANTITY_ITEM, SUM_ITEMAMOUNT, DISCOUNTM, AMOUNT, NETAMOUNT` (12 คอลัมน์) | 0 เลขที่เอกสาร, 1 รหัสลูกค้า, 3 รหัสสินค้า | :349-366, :718 |
| 1 | รหัสลูกค้า `ArCode` | `ARCODE, ARNAME` | `DOCDATE, DOCNO, ITEMCODE, ITEMNAME, UNITCODE, PERPRICE, QUANTITY_ITEM, SUM_ITEMAMOUNT, DISCOUNTM, AMOUNT, NETAMOUNT` (11) | 1, 2 | :369-386, :884 |
| 2 | กลุ่มลูกค้า `ArGroup` | `CUSTGROUPCODE, CUSTGROUP_NAME` | `DOCDATE, DOCNO, ARCODE, ARNAME, ITEMCODE, ITEMNAME, UNITCODE, PERPRICE, QUANTITY_ITEM, SUM_ITEMAMOUNT, DISCOUNTM, AMOUNT, NETAMOUNT` (13) | 1, 2, 4 | :389-408, :1072 |
| 3 | แผนก `DepartCode` | `DEPART_CODE, DEPART_NAME` | เหมือนโหมด 2 | 1, 2, 4 | :411 |
| 4 | โครงการ `ProjectCode` | `PROJECT_CODE, PROJECT_NAME` | เหมือนโหมด 2 | 1, 2, 4 | :433 |
| 5 | กลุ่มสินค้า `GroupCode` | `GROUPCODE_ITEM, ITEMGROUP_NAME` | เหมือนโหมด 2 | 1, 2, 4 | :455 |
| 6 | ชนิดสินค้า `TypeCode` | `ITEMTYPE_CODE, ITEMTYPE_NAME2` | เหมือนโหมด 2 | 1, 2, 4 | :477 |
| 7 | หมวดสินค้า `CategoryCode` | `CATEGORYCODE, CATEGORY_NAME` | เหมือนโหมด 2 | 1, 2, 4 | :499 |
| 8 | รูปแบบสินค้า `FormatCode` | `FORMAT_CODE, FORMATNAME` | เหมือนโหมด 2 | 1, 2, 4 | :521 |
| 9 | ยี่ห้อ `BrandCode` | `BRANDCODE, ITEMBRANDNAME` | เหมือนโหมด 2 | 1, 2, 4 | :543 |
| 10 | สินค้า `ItemCode` | `ITEMCODE, ITEMNAME` | `DOCDATE, DOCNO, ARCODE, ARNAME, UNITCODE, PERPRICE, QUANTITY_ITEM, SUM_ITEMAMOUNT, DISCOUNTM, AMOUNT, NETAMOUNT` (11 — ไม่ซ้ำรหัส/ชื่อสินค้า) | 1, 2, 4 | :565, :1236 |

5185 ใช้หัวกลุ่มชุดเดียวกันครบ 11 โหมด แต่แถวรายการต่างจาก 5180: โหมด 0-9 = `DOCNO, ITEMCODE, ITEMNAME, Remark, UINTCODE, PERPRICE, SUM, SUMAMOUNT, DISCOUNTM, AMOUNT, NETAMOUNT` (11 คอลัมน์ มีหมายเหตุ `Remark` แทนรหัส/ชื่อลูกค้า) ; โหมด 10 (สินค้า) = `DOCNO, DOCDATE, ARNAME, Remark, UINTCODE, PERPRICE, SUM, SUMAMOUNT, DISCOUNTM, AMOUNT, NETAMOUNT` (`blrepsaleordersubdateview.cpp:271-470`) ; ป้ายคอลัมน์ `Remark` = key `Remark` (ไทยจริงอยู่ใน DB — อนุมาน "หมายเหตุ").

#### A3. คอลัมน์ ความหมาย และยอดรวม (5180 โหมด 0)

| # | key | ความหมาย (ไทยจากคอมเมนต์โค้ด) | ที่มา/สูตร | หลักฐาน |
|---|---|---|---|---|
| 0 | DOCNO | เลขที่เอกสาร (ต่อท้าย `(ASSERT3)` เมื่อเปิด `IsPrintCancel` และเอกสารยกเลิก — แต่ SQL ตัด `IsCancel=0` ออกแล้ว จึงไม่เคยเกิด) | `BCQuotationSub.DocNo` | `...view.cpp:705` |
| 1-2 | ARCODE, ARNAME | รหัส/ชื่อลูกค้า | `ArCode` + `GetArName` | :706-707 |
| 3-4 | ITEMCODE, ITEMNAME | รหัส/ชื่อสินค้า | `ItemCode` + `GetItemName` | :708-709 |
| 5 | UNITCODE | หน่วยนับ | `OLEDBGetMasterName("BCItemUnit", UnitCode)` | :710 |
| 6 | PERPRICE | ราคาขาย/หน่วย | `Price` ทศนิยม `PricePoint` | :711 |
| 7 | QUANTITY_ITEM | ปริมาณ | `Qty` ทศนิยม `QtyPoint` | :712 |
| 8 | SUM_ITEMAMOUNT | ยอดรวมสินค้า | `Qty * Price` ปัดตาม `AmountPoint` (คำนวณใน view) | :673-674,713 |
| 9 | DISCOUNTM | ส่วนลด | ข้อความส่วนลดตามที่กรอก (`DiscountWord`) ไม่ใช่ตัวเลข ไม่มียอดรวมคอลัมน์นี้ | :714 |
| 10 | AMOUNT | จำนวนเงิน | `Amount` (ว่างเมื่อ 0) | :715 |
| 11 | NETAMOUNT | ยอดสุทธิ | `NetAmount` (ว่างเมื่อ 0) | :716 |

ยอดรวม: รวมย่อยต่อหัวกลุ่มพิมพ์ข้อความ "SUMDETAIL_WORD n TRANSAC" ที่คอลัมน์ชื่อสินค้า รวมเฉพาะ ปริมาณ(7), ยอดรวมสินค้า(8), จำนวนเงิน(10), ยอดสุทธิ(11) (`:641-649`) ; รวมทั้งสิ้นท้ายรายงาน "SUMTOTALDETAIL_WORD n TRANS_TYPE" (`:751-759`) — **ปริมาณรวมทั้งสิ้นพิมพ์ด้วยทศนิยม `AmountPoint` แทน `QtyPoint`** (`:755`) ; รวมท้ายหน้า `OnAddPageSummery` (`:1286-1342`).
ตารางที่อ่าน: 5180 `BCQuotationSub` (+ subselect `BCQuotation.DueDate/TaxRate`, `BCAR.GroupCode`) ; 5185 `BCSaleOrderSub` (+ `BCSaleOrder`, `BCAR`) ; ชื่อหัวกลุ่มอ่านจาก `BCARGroup`, `BCDepartment`, `BCProject`, `BCItemGroup`, `BCItemType`, `BCItemCategory`, `BCItemFormat`, `BCItemBrand`. คลิกทะลุ: เลขที่เอกสาร -> หน้าจอใบเสนอราคา (`BCQuotation.DOCNO='..'`), รหัสลูกค้า -> ทะเบียนลูกค้า (`BCAR.Code='..'`), รหัสสินค้า -> ทะเบียนสินค้า (`BCITEM.Code='..'`) (`blrepqoutationsubdateview.cpp:88-130`).

---

### กลุ่ม B. รายงานวิเคราะห์การขายแบบ "เรียงตาม n โหมด" (5204, 5277, 5310)

3 รายงานนี้ใช้กลไกเดียวกับกลุ่ม A คือ combo "เรียงตาม" เลือกหัวกลุ่ม (วันที่ / ลูกหนี้ / กลุ่มลูกค้า / สินค้า / Category ... / ยี่ห้อ / โครงการ / แผนก / ...) แล้วชุดคอลัมน์หัวกลุ่ม + `GROUP BY` + `ORDER BY` เปลี่ยนตามโหมด ส่วนชุดคอลัมน์รายการ (แถวระดับ 1) เหมือนกันทุกโหมด.

| menu id | ชื่อ (Champ) | handler (`MainFrm.cpp`) | view class (ไฟล์) | dialog (IDD : rc) | key | แหล่งข้อมูล |
|---|---|---|---|---|---|---|
| 5204 | รายงานวิเคราะห์การขายแบบสรุป | `OnBillsysAnalyrep51` :6007 | `CBLRepItemMasterDateSummaryView` (`BLRepItemMasterDateSummaryView.cpp`, 956 บรรทัด) | `IDD_BL_REP_ITEMMASTER_DATE_SUMMARY` 504x277 : `BCRepBilling.rc:3237` caption "35501$รายงานสรุปการขายสินค้า-ตามวันที่" | 35501 | `BCARInvoiceSub` อย่างเดียว (ไม่รวมลดหนี้/เพิ่มหนี้) |
| 5277 | รายงานวิเคราะห์การขายสุทธิ | `OnBillsysAnalyrep91` :6412 | `CBLRepNetDebtDateView` (`BLRepNetDebtDateView.cpp`, 1569 บรรทัด) | `IDD_BL_REP_NETDEBT_DATE` 499x301 : `BCRepBilling.rc:3809` caption "รายงานสรุปการขายสุทธิ - ตามวันที่" | 35901 | UNION `BCARInvoiceSub` + `BCCreditNoteSub` + `BCDebitNoteSub1` |
| 5310 | รายงานสรุปยอดขายสินค้า | `OnBillsysTrans37001` :9769 (`ID_BILLSYS_TRANS_37001`) | `CRepTotalSumByDateView` (`RepTotalSumByDateView.cpp`, 855 บรรทัด, A4 แนวตั้ง) | `IDD_REP_SUMBY_DATE` 502x310 : `BCRepBilling.rc:7214` caption "Dialog" | 37001 | `BCARInvoiceSub` อย่างเดียว (ไม่กรอง `StockType`) |

หมายเหตุชื่อเมนู: ชื่อเมนู 5310 "รายงานสรุปยอดขายสินค้า" ไม่ตรงเนื้อหา — รายงานนี้แยกยอดตามประเภทภาษี (ไม่แสดงสินค้าต่อบรรทัดยกเว้นโหมดสินค้า) ; ชื่อ 5204 "แบบสรุป" แต่คอลัมน์รายการยังเป็นสินค้าต่อบรรทัด (สรุปเป็นยอดรวมต่อหน่วยนับ ไม่ใช่ต่อเอกสาร).

#### B1. ตัวเลือก "เรียงตาม" (โหมด -> หัวกลุ่ม) และตัวเลือกพิเศษ

| โหมด | 5204 (12 โหมด) `dlg:106-117` | 5277 (14 โหมด) `dlg:95-108` | 5310 (15 โหมด) `dlg:109-123` | คอลัมน์หัวกลุ่ม (key) ใน view |
|---|---|---|---|---|
| 0 | วันที่เอกสาร | เหมือน | เหมือน | `DATEP` (5204 :269) / `DOCDATE` (5277 :279, 5310 :80) |
| 1 | รหัสลูกค้า | เหมือน | เหมือน | `ARCODE`+`ARNAME` (5204) / `CUST_TYPE`+`AR_NAME` (5277 :297) — key `CUST_TYPE` ในโค้ดชื่อ "ประเภท" แต่ข้อมูลที่ใส่คือรหัสลูกค้า (อนุมานจาก `OnCallTrans` โหมด 1 คลิกรหัสลูกค้า `BCAR.Code`) |
| 2 | รหัสกลุ่มลูกค้า | เหมือน | เหมือน | `CUSTGROUPCODE`+`CUSTGROUP_NAME` |
| 3 | รหัสสินค้า | เหมือน | เหมือน | ไม่มีหัวกลุ่ม (5204/5277 พิมพ์เฉพาะแถวสินค้า ผ่าน `RunProcess2` 5204 :735) ; 5310 หัวคอลัมน์ `PRODUCTCODE`+`ITEMNAME` |
| 4 | Category (กลุ่มสินค้า) | เหมือน | เหมือน | `GROUPCODE_ITEM`+`ITEMGROUP_NAME` |
| 5 | Sub - Category (ชนิดสินค้า) | เหมือน | เหมือน | `ITEMTYPE_*` |
| 6 | Class (หมวดสินค้า) | เหมือน | เหมือน | `CATEGORYCODE`+`CATEGORY_NAME` |
| 7 | Sub - Class (รูปแบบสินค้า) | เหมือน | เหมือน | `FORMAT_CODE`+`FORMAT_NAME` |
| 8 | รหัสยี่ห้อ | เหมือน | เหมือน | `BRANDCODE`+`ITEMBRANDNAME` |
| 9 | รหัสโครงการ | เหมือน | เหมือน | `PROJECT_CODE`+`PROJECT_NAME` |
| 10 | รหัสแผนก | เหมือน | เหมือน | `DEPART_CODE`+`DEPART_NAME` |
| 11 | รหัสเขตการขาย (`SALEAREA_CODE`) | รหัสการจัดสรร (`ALLOCATE_CODE`) | รหัสการจัดสรร | 5204 ไม่มีโหมดจัดสรร/พนักงานขาย/ขนส่ง |
| 12 | — | รหัสพนักงานขาย (`SALECODE`+`SALE_NAME`) | รหัสพนักงานขาย | |
| 13 | — | รหัสเขตการขาย (`SALEAREA_*`) | รหัสเขตการขาย | |
| 14 | — | — | รหัสการขนส่ง (`TRANSPORT_CODE`+`TRANSPORT_NAME`) | |

หลักฐานชุดหัวกลุ่ม: 5204 `BLRepItemMasterDateSummaryView.cpp:266-427` ; 5277 `BLRepNetDebtDateView.cpp:279-522` (14 case) ; 5310 `RepTotalSumByDateView.cpp:80-247` (15 case). ป้ายโหมด 4-7 เป็นอังกฤษในคอมโบ (Category / Sub - Category / Class / Sub - Class) เหมือนกลุ่ม A. ป้ายไทยของ key หัวกลุ่มอยู่ใน DB `myresource` — ใช้ความหมายตามชื่อ key (อนุมาน).

ตัวเลือกพิเศษเฉพาะรายงาน:
- 5277 มีช่องติ๊ก "Display summary" (`IDC_CHK_DISPLAYSUMMARY`) — เมื่อติ๊กพิมพ์เฉพาะแถวหัวกลุ่ม+ยอดรวม ไม่พิมพ์รายการสินค้า (`RunProcess_bySummary`, `BLRepNetDebtDateView.cpp:557-560,973`) ; ช่องนี้ถูก disable เมื่อเลือกโหมด 13 (เขตการขาย) (`BLRepNetDebtDateDlg.cpp:1145-1157`).
- 5277 มีช่องติ๊ก "แสดงเฉพาะรายการสินค้าที่มีโปรโมชั่น" — เพิ่มเงื่อนไข `ItemCode IN (SELECT ItemCode FROM BCPromoPrice WHERE DocDate BETWEEN <จากวันที่> AND <ถึงวันที่>)` ให้ทั้ง 3 ตาราง (ใบขาย/ลดหนี้/เพิ่มหนี้) (`BLRepNetDebtDateDlg.cpp:608-615`).
- 5277 จำกัดสิทธิ์ระดับแถว: ถ้าพารามิเตอร์ระบบ `BCS_GET_SHOW_AR_BY_SALE` = 1 และผู้ใช้ปัจจุบันเป็นพนักงานขาย/กลุ่มพนักงานขาย เติม `salecode = <ผู้ใช้>` (หรือสมาชิกกลุ่ม) ให้ทั้ง 3 sub-query (`BLRepNetDebtDateDlg.cpp:626-672`).
- 5310 ไม่มีช่องพิเศษ ; มีตัวกรองเพิ่มคือ "การขนส่ง" (`BCTransport`).

#### B2. ช่องเงื่อนไข (ลำดับตาม rc) — baseline ของ 3 รายงาน

| # | ป้าย (ไทยแท้จาก rc ถ้ามี) | ชนิด | SQL | 5204 | 5277 | 5310 | หลักฐาน (5204 / 5277 / 5310) |
|---|---|---|---|---|---|---|---|
| 1 | เรียงตาม: | combo | กำหนดโหมด | มี | มี | มี | `dlg:106-117` / `:95-108` / `:109-123` |
| 2 | วันที่ จาก–ถึง | วันที่ | `DocDate BETWEEN` (ใส่จากอย่างเดียว = `=`) | มี | มี | มี | `5204dlg:223` |
| 3 | ลูกหนี้ จาก–ถึง | lookup (รหัส) | `ArCode` | มี | มี | มี | `5204dlg:261` |
| 4 | กลุ่มลูกค้า จาก–ถึง | lookup | `CustGroupCode` | มี | มี | มี | `5204dlg:280` |
| 5 | แผนก จาก–ถึง | lookup | `DepartCode` | มี | มี | มี | `5204dlg:299` |
| 6 | โครงการ จาก–ถึง | lookup | `ProjectCode` | มี | มี | มี | `5204dlg:318` |
| 7 | การจัดสรร จาก–ถึง | lookup | `AllocateCode` | — | มี | มี | (5277/5310 dlg) |
| 8 | พนักงานขาย จาก–ถึง | lookup | `SaleCode` | — | มี | มี | |
| 9 | เขตการขาย จาก–ถึง | lookup | `SaleAreaCode` | มี | มี | มี | `5204dlg:338` |
| 10 | การขนส่ง จาก–ถึง | lookup | `TransportCode` | — | — | มี | |
| 11 | หน่วยนับ จาก–ถึง | lookup | `UnitCode` | มี | — | — | `5204dlg:244` |
| 12 | Category จาก–ถึง | lookup | `GroupCode` | มี | มี | มี | `5204dlg:359` |
| 13 | Sub-Category จาก–ถึง | lookup | `TypeCode BETWEEN` | **บั๊ก** `:377` ส่ง `Start,Start` | ปกติ `:511-513` | ปกติ `:626-628` | ดู R0/quirk |
| 14 | Class จาก–ถึง | lookup | `CategoryCode` | **บั๊ก** `:391-395` ตรวจ `strStopTypeCode` แทนตัวแปรของ Class | ปกติ | ปกติ | |
| 15 | Sub-Class จาก–ถึง | lookup | `FormatCode` | มี `:431` | มี | มี | |
| 16 | ยี่ห้อ จาก–ถึง | lookup | `BrandCode` | มี `:413` | มี | มี | |
| 17 | สินค้า จาก–ถึง | lookup | `ItemCode` | มี `:451` | มี | มี | `5204dlg:451` |

ไฟล์ dialog: 5204 `BLRepItemMasterDateSummaryDlg.cpp` ; 5277 `BLRepNetDebtDateDlg.cpp` ; 5310 `RepTotalSumByDateDlg.cpp`. ลำดับบนจอตาม rc (ยังไม่ได้ไล่เทียบพิกัด x,y ของ 5277/5310 ทีละช่อง — "ยังไม่ verify" ลำดับบนจอ ; ลำดับในตารางนี้อิงลำดับใน `GetReportCondition`).

SQL จริงต่อโหมด (5204 `dlg:469-`): `SELECT <คีย์กลุ่ม>, ItemCode, UnitCode, CurrencyCode, ExchangeRate, TaxType, [PosStatus], SUM(Qty) Sum_Qty, SUM(DiscountAmount) Sum_DisCount, SUM(Amount) Sum_Amount, SUM(NetAmount) Sum_NetAmount FROM BCARInvoiceSub ... GROUP BY ... ORDER BY <คีย์กลุ่ม>, ItemCode, UnitCode` ; `m_strDefaultSQL` ที่ `dlg:28-37` ผิดรูป (ขาด/เกินเครื่องหมายจุลภาค) แต่ถูกเขียนทับก่อนใช้. 5277 รวมสามแหล่ง (`OnGetSQLScript` `BLRepNetDebtDateDlg.cpp:690-`): ใบขาย `MyType=4` + `StockType<>7` + กฎ `PosStatus` ; ลดหนี้ `BCCreditNoteSub` (`ItemType<>7`, `IsPos AS PosStatus`) ; เพิ่มหนี้ `BCDebitNoteSub1` (`ItemType<>7`). 5310: `SELECT <คีย์>, ExchangeRate, ExceptTax, TaxType, SUM(Qty), SUM(Amount), SUM(DiscountAmount), SUM(NetAmount)` จาก `BCARInvoiceSub` `TransState=0 AND IsCancel=0 [AND PosStatus=0]` — ไม่กรอง `StockType`/`ItemCode<>'.'` (ต่างจากรายงานอื่นในกลุ่ม).

#### B3. คอลัมน์รายการ (แถวระดับ 1) และสูตร

| รายงาน | คอลัมน์ตามลำดับ (key = ความหมาย) | หลักฐาน |
|---|---|---|
| 5204 (7 คอลัมน์) | `ITEMCODE` รหัสสินค้า ; `ITEMNAME` ชื่อสินค้า ; `UINTCODE` หน่วยนับ ; `ITEMQTY` จำนวน ; `DISCOUNTAMOUNT` ยอดส่วนลด ; `AMOUNT` จำนวนเงิน ; `NETAMOUNT` ยอดสุทธิ (ไทยจากคอมเมนต์โค้ด) | `BLRepItemMasterDateSummaryView.cpp:271-278` |
| 5277 (11 คอลัมน์) | `ITEMCODE` รหัสสินค้า ; `ITEMNAME` ชื่อสินค้า ; `UINTCODE` หน่วยนับ ; `SUMQTY` จำนวนขาย ; `QTY_RETURN` จำนวนคืน ; `QTY_ADD` จำนวนเพิ่ม ; `SALE_AMOUNT` มูลค่าขาย ; `RETURN_AMOUNT` มูลค่าคืน ; `ADD_AMOUNT` มูลค่าเพิ่มหนี้ ; `NETAMOUNT` Net amount ; `EXCHANGE_PROFIT` Profit(Loss) กำไร(ขาดทุน) — คอลัมน์สุดท้ายแสดงเฉพาะเมื่อ `IsShowProfitLoss` | `BLRepNetDebtDateView.cpp:284-294,684` |
| 5310 (ต่อแถวหัวกลุ่ม 1 แถว ไม่มีรายการย่อย) | คีย์+ชื่อกลุ่ม ; `TOTAL_ITEM_CALCTAX` ยอดสินค้าคิดภาษี (`ExceptTax=0`) ; `TOTAL_NOTTAX` มูลค่าสินค้ายกเว้นภาษี (`=1`) ; `GOODS_VALUE_VATZERO` ยอดสินค้าภาษีอัตราศูนย์ (`=2`) ; `DISCOUNTAMOUNT` ยอดส่วนลด ; `TOTAL_SUM` ยอดขายรวม (ไทยจากคอมเมนต์) | `RepTotalSumByDateView.cpp:80-250` |

สูตร/ข้อกำหนดรายตัว:

| หัวข้อ | 5204 | 5277 | 5310 |
|---|---|---|---|
| การแปลงหน่วยนับ | `OnGetQtyPacking(...,nMak)` แปลงเป็นหน่วยเก็บด้วย `Rate1/Rate2` (nMak=1: `q*Rate1/Rate2`, อื่นๆ: `q*Rate2/Rate1`) แล้วแปลงกลับเป็นหน่วยของรายการ (`-1`) `:929-956` | แปลงเป็นหน่วยเริ่มต้น `qty*dRate2/dRate1` | ไม่แปลง (รวม `SUM(Qty)` ดิบ — ไม่แสดงคอลัมน์จำนวนในชุดคอลัมน์ที่อ่านได้) |
| สกุลเงิน | ถ้ารายงานเป็นสกุลบริษัท คูณ `ExchangeRate` ; ไม่เช่นนั้นต่อท้ายสัญลักษณ์ (`BCCurrency.Symbol`, `฿` สกุลบริษัท, ว่างถ้า `CurrencySymbol=0`) `:512-532` | — | คูณ `ExchangeRate` เมื่อ `MultiCurrency=1` และ `bHomeCurrencyReport=1` ; `XRound(dNetAmount)` เรียกซ้ำสองครั้ง ส่วน `dDiscount` ไม่ถูกปัด `:310-325` (โหมด >0 `RunProcess2` `:486-501` ; โหมด 0 = `RunProcess` `:262-`) |
| ข้ามแถว | `if(dItemQty>0.0)` **ข้ามสินค้าที่รวมจำนวนเป็น 0 หรือติดลบ** `:540-577` | — | — |
| `TaxType` | อ่านแต่ไม่ใช้ | — | ใช้เฉพาะ `ExceptTax=0`: `TaxType==0` ใช้ `Amount` (ก่อนภาษี) ; อื่นๆ ใช้ `NetAmount` `:365` (โหมด >0 `:560`) |
| กำไร | — | `MyType` 4: `+= amount − cost` ; 5 (ลดหนี้): `−= (amount − cost)` ; 6 (เพิ่มหนี้): `+= (amount − cost)` `:827-845` | — |
| มูลค่าสุทธิ | `NetAmount` | `NETAMOUNT = ขาย − ลดหนี้ + เพิ่มหนี้` `:683` (คอมเมนต์ในโค้ดตั้งชื่อผิดว่า "มูลค่าเพิ่ม") | `TOTAL_SUM = คิดภาษี + ยกเว้น + อัตราศูนย์` `:329,377` (ส่วนลดแยกคอลัมน์ ไม่นำมาลบซ้ำ) |
| ยอดรวมย่อย/รวมทั้งสิ้น | แถว `SUMDETAIL_WORD n TRANS_TYPE` ต่อหัวกลุ่ม `:889`, รวมทั้งสิ้น `SUMTOTALDETAIL_WORD n <คำ>` `:715-719`, รวมต่อหน้า `OnAddPageSummery` `:905` | รวมย่อย `:573`, รวมทั้งสิ้น `:574`, รวมต่อหน้า `:1512` (มีบรรทัด `OnAddPageSummery` ถูกคอมเมนต์ที่ `:1237`) | รวมทั้งสิ้น `:405-409,618-622`, รวมต่อหน้า `:641` |
| คลิกทะลุ (`OnCallTrans`) | โหมด 1: หัวกลุ่ม col 0 -> ทะเบียนลูกหนี้ (`BCAR.Code`), รายการ col 0 -> สินค้า (`BCITEM.Code`) ; โหมด 2: แถวหัวกลุ่ม col 0 -> สินค้า ; อื่นๆ: แถวรายการ col 0 -> สินค้า `:79-` | เหมือน 5204 (`:79-133`) | โหมด 1 -> ลูกหนี้, โหมด 3 -> สินค้า เท่านั้น `:824-855` |
| ตารางที่อ่าน | `BCARInvoiceSub`, `BCItem`, `BCStkPacking`, `BCCurrency` | `BCARInvoiceSub`, `BCCreditNoteSub`, `BCDebitNoteSub1`, `BCPromoPrice`, `BCItem`, `BCStkPacking` | `BCARInvoiceSub` |

ข้อสังเกตเฉพาะกลุ่ม B ที่ BC ต้องตัดสินใจ: 5204 ไม่รวมลดหนี้/เพิ่มหนี้ (ต่างจาก 5277) — ถ้า BC จะทำ "การขายสุทธิ" ต้องยึด 5277 ; 5310 ไม่กรอง `StockType <> 7`/`ItemCode<>'.'` จึงยอดอาจต่างจาก 5204 เมื่อมีแถวสินค้าชุดแบบที่ 2 (ยังไม่ได้ทดสอบกับข้อมูลจริง).

---

### กลุ่ม C. รายงานวิเคราะห์ใบรับคืนสินค้า/ลดหนี้ และใบเพิ่มหนี้ (ลูกหนี้) (5232, 5236, 5240, 5244)

โครงสร้างเหมือนกันทั้ง 4 รายการ: กล่องเงื่อนไขเดียวกัน (วันที่ / แผนก / กลุ่ม / โครงการ / สินค้า / จำนวน / เงิน / เรียงตาม) ต่างกันที่ตารางต้นทาง (ลดหนี้ `BCCreditNoteSub` เทียบเพิ่มหนี้ `BCDebitNoteSub1`) และ "แบบแจกแจง" (อ่านรายการย่อยตรง ๆ) เทียบ "แบบสรุป" (`GROUP BY` ให้ยอดรวมต่อกลุ่ม).

| menu id | ชื่อ (Champ) | handler (`MainFrm.cpp`) | view class (ไฟล์, บรรทัด) | dialog (IDD : rc, caption) | key รายงาน / key ชื่อ dialog | ตาราง |
|---|---|---|---|---|---|---|
| 5232 | รายงานวิเคราะห์ใบรับคืนสินค้า/ลดหนี้แบบแจกแจง | `OnBillsysAnalyrep461` :5556 | `CBLRepCreditNoteDateView` (`BLRepCreditNoteDateView.cpp`, 963) | `IDD_BL_REP_CREDITNOTE_DATE` 414x339 : `BCRepBilling.rc:4342` "รายงานใบรับคืนสินค้า - ตามวันที่" | 35461 / 35460 | `BCCreditNoteSub` |
| 5236 | รายงานวิเคราะห์ใบรับคืนสินค้า/ลดหนี้แบบสรุป | `OnBillsysAnalyrep471` :5603 | `CBLRepCreditNoteSDateView` (`BLRepCreditNoteSDateView.cpp`, 989) | `IDD_BL_REP_CREDITNOTE_SDATE` 398x326 : `BCRepBilling.rc:4150` "รายงานใบรับคืนสินค้า - ตามวันที่" | 35471 / 35470 | `BCCreditNoteSub` (GROUP BY) |
| 5240 | รายงานวิเคราะห์ใบเพิ่มสินค้า(ลูกหนี้)แบบแจกแจง | `OnBillsysAnalyrep481` :5651 | `CBLRepDebitNoteDateView` (`BLRepDebitNoteDateView.cpp`, 931) | `IDD_BL_REP_DEBITNOTE_DATE` 411x343 : `BCRepBilling.rc:4726` "รายงานใบเพิ่มหนี้(ลูกหนี้) - ตามวันที่" | 35481 / 35480 | `BCDebitNoteSub1` |
| 5244 | รายงานวิเคราะห์ใบเพิ่มสินค้า(ลูกหนี้)แบบสรุป | `OnBillsysAnalyrep491` :5698 | `CBLRepDebitNoteSDateView` (`BLRepDebitNoteSDateView.cpp`, 956) | `IDD_BL_REP_DEBITNOTE_SDATE` 431x343 : `BCRepBilling.rc:4534` "รายงานใบเพิ่มหนี้(ลูกหนี้) - ตามวันที่" | 35491 / 35490 | `BCDebitNoteSub1` (GROUP BY) |

หลักฐาน key: `BLRepCreditNoteDateDlg.cpp:24`, `BLRepCreditNoteSDateDlg.cpp:24`, `BLRepDebitNoteDateDlg.cpp:24`, `BLRepDebitNoteSDateDlg.cpp:24`. ทุกรายงาน A4 แนวนอน. ชื่อเมนู "ใบเพิ่มสินค้า(ลูกหนี้)" ตรงกับตาราง `BCDebitNoteSub1` (ใบเพิ่มหนี้ลูกหนี้) — ในรายงานเรียก "ใบเพิ่มหนี้(ลูกหนี้)".

#### C1. ช่องเงื่อนไข (ตรงกันทั้ง 4 รายงาน ; ป้ายไทยแท้จาก `BCRepBilling.rc:4342` และ dialog อื่นในกลุ่ม)

| # | ป้าย | ชนิด | SQL | หมายเหตุ / หลักฐาน (5232) |
|---|---|---|---|---|
| 1 | พิมพ์จากวันที่: / พิมพ์ถึงวันที่: | วันที่ | `DocDate BETWEEN` (ใส่จากอย่างเดียว = `=`) | `BLRepCreditNoteDateDlg.cpp:147-160` |
| 2 | พิมพ์จากแผนก: / ถึงแผนก: | lookup `BCDepartment` | `DepartCode BETWEEN` | `:79-80,172-176` |
| 3 | พิมพ์จากกลุ่ม: / ถึงกลุ่ม: | lookup `BCItemGroup` | `GroupCode BETWEEN` | `:81-82,191-195` ; ข้อความเงื่อนไขท้ายรายงานใช้ key `FROM_GROUP`/`GROUP_TO` |
| 4 | พิมพ์จากโครงการ: / ถึงโครงการ: | lookup `BCProject` | `ProjectCode BETWEEN` | `:83-84,210-214` ; ข้อความเงื่อนไขใช้ key `FROM_DEPART`/`TO_DEPART_CODE` (ของแผนก) ผิด key — ป้ายเงื่อนไขท้ายรายงานจึงเขียนว่า "แผนก" ทั้งที่กรองโครงการ |
| 5 | พิมพ์จากสินค้า: / ถึงสินค้า: | ช่องรหัส + ปุ่มค้นหา (`OnFindStartItem` เปิด `Browse2 "ItemTransConfig"`) | `ItemCode BETWEEN` | `:229-237` |
| 6 | พิมพ์จากจำนวน: / ถึงจำนวน: (`FROM_SUM`/`TO_SUM`) | ตัวเลข (`QtyPoint`) | `DiscQty BETWEEN` (ลดหนี้) / `TempQty BETWEEN ... dStartQty AND dStartQty` (เพิ่มหนี้ — **บั๊ก** ใช้ค่าจากเป็นค่าถึง) ; ทำงานเมื่อ "จาก" > 0 เท่านั้น | `:243-255` (5232) ; `BLRepDebitNoteDateDlg.cpp:243` ; `BLRepDebitNoteSDateDlg.cpp:281` |
| 7 | พิมพ์จากเงิน: / พิมพ์ถึงเงิน: (`FROM_AMOUNT`/`TO_AMOUNT`) | ตัวเลข (`AmountPoint`) | `Amount BETWEEN` ; ทำงานเมื่อ "จาก" > 0 | `:266-275` |
| 8 | เรียงตาม: | combo 4 ค่า | 0 วันที่ (`DOCDATE`) / 1 แผนก (`DEPART`) / 2 กลุ่มสินค้า (`ITEMGROUP`) / 3 โครงการ (`PROJECTCODE`) — ไทยแท้ยังไม่พบ (key ใน `myresource` ; คอมเมนต์ไทยในโค้ด `// วันที่ / แผนก / กลุ่มสินค้า / โครงการ`) | `:92-99` |
| — | พื้นที่เงื่อนไขเพิ่มเติมของผู้ใช้ (`IDC_EXCON_RECT`) และรูปตัวอย่างรายงาน (`IDC_REP_PIC_RECT`) | — | `InitExCondition` / `InitRepExamPicture` | `:102-104` |

เงื่อนไขหลักที่ฝังไว้เสมอ: `TransState = 0 AND IsCancel = 0` ; ลดหนี้เพิ่ม `AND IsPos = 0` เมื่อ `PosRepSeparate <> 1` (`BLRepCreditNoteDateDlg.cpp:26-29`) ; เพิ่มหนี้ไม่มีเงื่อนไข POS (`BLRepDebitNoteDateDlg.cpp:26`). ไม่กรอง `StockType`/`ItemCode<>'.'`. ข้อความ "พิมพ์จากรายการทั้งหมด" (`FROM_LIST_TOTAL`) ขึ้นเมื่อไม่ได้ใส่เงื่อนไขใด (`:280-282`).

#### C2. ORDER BY / GROUP BY ตามโหมด "เรียงตาม"

| โหมด | 5232 / 5240 (แจกแจง) | 5236 / 5244 (สรุป) |
|---|---|---|
| 0 วันที่ | `ORDER BY DocDate,DocNo,LineNumber,ItemCode,UnitCode` (5232 `dlg:288`) ; 5240 ไม่มี `ItemCode` (`dlg:281`) | `GROUP BY DocDate,ItemCode,UnitCode,CurrencyCode,ExchangeRate` : หนึ่งแถวต่อวัน × สินค้า × หน่วยนับ × สกุลเงิน (5236 `dlg:359`, 5244 `dlg:321`) |
| 1 แผนก | `ORDER BY DepartCode,DocDate,DocNo,LineNumber,ItemCode,UnitCode` (`dlg:294`) | `GROUP BY DepartCode,DocNo,DocDate,ItemCode,UnitCode,...` — ยังแตกต่อเลขที่เอกสาร ไม่ใช่สรุปต่อแผนกจริง (`5236 dlg:365`, `5244 dlg:327`) |
| 2 กลุ่มสินค้า | `ORDER BY GroupCode,DocDate,DocNo,...` (`dlg:300`) | `GROUP BY GroupCode,DocNo,DocDate,...` (`5236 dlg:371`, `5244 dlg:333`) |
| 3 โครงการ | `ORDER BY ProjectCode,DocDate,DocNo,...` (`dlg:306`) | `GROUP BY ProjectCode,DocNo,DocDate,...` (`5236 dlg:377`, `5244 dlg:339`) |

ข้อสรุป: "แบบสรุป" แตกต่างจาก "แบบแจกแจง" จริงเฉพาะโหมดวันที่ (รวมสินค้าเดียวกันในวันเดียวกันเป็นแถวเดียว ไม่มีเลขที่เอกสาร) ; โหมดแผนก/กลุ่ม/โครงการยังเห็นเลขที่เอกสารต่อบรรทัดเหมือนแบบแจกแจง (รวมเฉพาะรายการที่ซ้ำสินค้า/หน่วยนับ/สกุลเงินในเอกสารเดียวกัน).

5236 เลือกฟิลด์สรุป `SUM(DiscQty) Sum_Qty, SUM(Amount) Sum_Amount, SUM(DiscountAmount) Sum_DisCount, SUM(NetAmount) Sum_NetAmount` (`BLRepCreditNoteSDateDlg.cpp:28-36`) ; เงื่อนไข "จำนวน/เงิน" ของ 5236 ถูกเปลี่ยนเป็น `HAVING SUM(DISCQTY) ...` / `HAVING SUM(AMOUNT) ...` (`:324,346,348`) แต่ 5244 ใช้ `WHERE` ระดับแถวตามเดิม (`BLRepDebitNoteSDateDlg.cpp:281,302`).

#### C3. คอลัมน์ และสูตร

| รายงาน / โหมด | คอลัมน์หัวกลุ่ม (ชุดที่ 1) | คอลัมน์รายการ (ชุดที่ 2) | หลักฐาน |
|---|---|---|---|
| 5232 โหมด 0 (ไม่มีหัวกลุ่ม แถวเดียวต่อบรรทัดรายการ) | — | `DATEP` วันที่ ; `CREDITNOTENO` เลขที่ใบรับคืน ; `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `ISSUERET_NUMBER` จำนวนรับคืน ; `DISCOUNTM` ส่วนลด ; `AMOUNT` จำนวนเงิน ; `NETAMOUNT` ยอดสุทธิ | `BLRepCreditNoteDateView.cpp:205-216` |
| 5232 โหมด 1/2/3 | แผนก `DEPART_CODE`+`DEPART_NAME` / กลุ่มสินค้า `GROUPCODE_ITEM`+`ITEMGROUP_NAME` / โครงการ `PROJECT_CODE`+`PROJECT_NAME` | `CREDITNOTENO` ; `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `ISSUERET_NUMBER` ; `DISCOUNTM` ; `AMOUNT` ; `NETAMOUNT` (ไม่มีคอลัมน์วันที่) | `:219-265` (ชุดหัวกลุ่ม+รายการ ต่อโหมด) |
| 5236 โหมด 0 | `DATEP` | `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `ISSUERET_NUMBER` ; `DISCOUNTM` ; `AMOUNT` ; `NETAMOUNT` | `BLRepCreditNoteSDateView.cpp:198-212` |
| 5236 โหมด 1 (แผนก) | `DEPART_CODE`+`DEPART_NAME` | `CREDITNOTENO` ; `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `ISSUERET_NUMBER` ; **`AMOUNT` (คอลัมน์ 5) ; `DISCOUNTM` (คอลัมน์ 6)** ; `NETAMOUNT` | `:213-226` (หัวคอลัมน์ 5,6 สลับกับโหมดอื่น) |
| 5236 โหมด 2/3 (กลุ่ม/โครงการ) | `GROUPCODE_ITEM`+`ITEMGROUP_NAME` / `PROJECT_CODE`+`PROJECT_NAME` | `CREDITNOTENO` ; `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `ISSUERET_NUMBER` ; `DISCOUNTM` ; `AMOUNT` ; `NETAMOUNT` | `:229-258` |
| 5240 โหมด 0 | — | `DATEP` ; `CREDITNOTE_DOCNO` ; `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `QTY_INCDEBT` ; `DISCOUNTM` ; `AMOUNT` ; `NETAMOUNT` | `BLRepDebitNoteDateView.cpp:194-206` |
| 5240 โหมด 1/2/3 | เหมือน 5232 | `CREDITNOTE_DOCNO` ; `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `QTY_INCDEBT` (โหมด 2 ใช้ `ITEMQTY`) ; `DISCOUNTM` ; `AMOUNT` ; `NETAMOUNT` | `:208-252` |
| 5244 โหมด 0 | `DATEP` | `ITEMCODE` ; `ITEMNAME` ; `UINTCODE` ; `QTY_INCDEBT` ; `DISCOUNTM` ; `AMOUNT` ; `NETAMOUNT` | `BLRepDebitNoteSDateView.cpp:184-196` |
| 5244 โหมด 1/2/3 | เหมือน 5232 | โหมด 1 ใช้ `CREDITNOTENO` เป็นเลขที่ ส่วนโหมด 2/3 ใช้ `CREDITNOTE_DOCNO` ; **โหมด 1 คอลัมน์ 5 = `AMOUNT`, 6 = `DISCOUNTM`** (สลับ) | `:198-244` |

ไทยของ key: `DATEP` วันที่, `ITEMCODE` รหัสสินค้า, `ITEMNAME` ชื่อสินค้า, `UINTCODE` หน่วยนับ, `ISSUERET_NUMBER` จำนวนรับคืน (คอมเมนต์โค้ด `// จำนวนรับคืน` ที่ `BLRepCreditNoteDateView.cpp:212`), `DISCOUNTM` ส่วนลด, `AMOUNT` จำนวนเงิน, `NETAMOUNT` ยอดสุทธิ, `CREDITNOTENO` เลขที่ใบรับคืน/ใบส่งคืน (คอมเมนต์ `:208`), `QTY_INCDEBT` จำนวนที่เพิ่มหนี้ (อนุมาน — ไทยจริงอยู่ใน `myresource`), `CREDITNOTE_DOCNO` เลขที่เอกสารเพิ่มหนี้ (อนุมาน), `DEPART_*`/`GROUPCODE_ITEM`/`ITEMGROUP_NAME`/`PROJECT_*` รหัส/ชื่อ แผนก/กลุ่มสินค้า/โครงการ.

| หัวข้อ | พฤติกรรม | หลักฐาน |
|---|---|---|
| จำนวน (5232) | `qty = (CauseType != 1) ? DiscQty : 0` — ประเภท "ราคาผิด" (CauseType 1) ไม่นับจำนวน แต่ยังนับมูลค่า | `BLRepCreditNoteDateView.cpp:313-316,438-441,607-610,774-777` |
| จำนวน (5236) | รวม `SUM(DiscQty)` ดิบ — ไม่เช็ค `CauseType` (โค้ดตรวจ `CauseType` ถูกคอมเมนต์ทิ้ง) ; จึงจำนวนของ 5236 อาจต่างจาก 5232 เมื่อมีใบ "ราคาผิด" | `BLRepCreditNoteSDateView.cpp:310,474` |
| จำนวน (5240/5244) | `DiscQty` ตรง ๆ ไม่มีกฎ CauseType ; แต่ตัวกรองจำนวนใช้ `TempQty` | `BLRepDebitNoteDateView.cpp:301` , `...Dlg.cpp:243` |
| ประเภทรับคืนของใบลดหนี้ | ค่า `CauseType` 0 = รับคืน/ลดหนี้, 1 = ราคาผิด, 2 = สินค้าแตกเสียหาย ; บรรทัดลดหนี้เก็บ `TempQty` (จำนวนที่คีย์) และ `DiscQty` (จำนวนที่กระทบสต๊อก ; ถ้า StockProcess = 0 ใช้เท่ากับ `TempQty`) | `BILLFrmInvCreditNote4.cpp:362-364,1390-1397` |
| UPDATE ข้อมูลตอนเปิดกล่องเงื่อนไข | `UPDATE BCCreditNoteSub SET DiscQty = TempQty WHERE DiscQty <> TempQty AND CauseType = 2` — ซ่อมจำนวนของใบ "สินค้าแตกเสียหาย" ให้ตรงจำนวนที่คีย์ทุกครั้งที่กดพิมพ์ ; **BC ต้องไม่ทำ** (รายงานต้องอ่านอย่างเดียว) | `BLRepCreditNoteDateDlg.cpp:141-143`, `BLRepCreditNoteSDateDlg.cpp:146-148` |
| หลายสกุลเงิน | คูณ `ExchangeRate` ถ้า `MultiCurrency=1` และ `bHomeCurrencyReport=1` (สัญลักษณ์ `฿`) ; ไม่เช่นนั้นต่อท้ายสัญลักษณ์ `BCCurrency.Symbol` ; ปัดตาม `AmountPoint` ; ค่าศูนย์ของส่วนลด/จำนวนเงิน/สุทธิแสดงเป็นช่องว่าง | `BLRepCreditNoteDateView.cpp:321-362` |
| ยอดรวม | รวมต่อหน้า (`OnAddPageSummery` `:902`) ; โหมด 0 (5232): แถวรวมทั้งสิ้น `SUMTOTALDETAIL_WORD n <คำ>` แถวเดียวท้ายรายงาน `:376-389` ; โหมด 1-3: หัวกลุ่ม -> รายการ -> แถวรวมกลุ่ม `SUMDETAIL_WORD n ...` -> รวมทั้งสิ้น (`:481,542,557`) ; 5236 โหมด 0 มีแถวรวมต่อวัน (`:344,403`) + รวมทั้งสิ้น (`:417`) | ตามที่ระบุ |
| ไม่ติดลบ | ใบลดหนี้ (5232/5236) แสดงยอดเป็นบวก ไม่คูณ −1 (ต่างจาก 5277 ที่ลบออกจากยอดขาย) — ยังไม่ได้ verify ว่าค่าใน `BCCreditNoteSub.Amount` เก็บเป็นบวกเสมอ | `BLRepCreditNoteDateView.cpp:316-319` |
| คลิกทะลุ | 5232 โหมด 0: คอลัมน์ 1 (เลขที่) -> ใบรับคืน (`BCCreditNote.DOCNO`), คอลัมน์ 2 (รหัสสินค้า) -> ข้อมูลหลักสินค้า (`BCITEM.Code`) ; โหมด 1-3: ระดับรายการ คอลัมน์ 0 (เลขที่) / 1 (รหัสสินค้า) ; 5240 -> `BCDebitNote1.DOCNO` / `BCITEM.Code` ; 5236/5244 -> สินค้า/เลขที่ในโหมดแผนก-กลุ่ม-โครงการ | `BLRepCreditNoteDateView.cpp:77-132`, `BLRepDebitNoteDateView.cpp:89,96`, `BLRepCreditNoteSDateView.cpp:90,101,108`, `BLRepDebitNoteSDateView.cpp:87` |

#### C4. ข้อผิดพลาดใน Champ ของกลุ่ม C ที่ BC ห้ามคัดลอก

| # | ข้อผิดพลาด | หลักฐาน |
|---|---|---|
| 1 | ตัวกรองจำนวนเพิ่มหนี้ใช้ `TempQty BETWEEN จาก AND จาก` (ค่า "ถึง" ถูกแทนด้วยค่า "จาก") | `BLRepDebitNoteDateDlg.cpp:243`, `BLRepDebitNoteSDateDlg.cpp:281` |
| 2 | ข้อความเงื่อนไขท้ายรายงานแสดงจำนวน "จาก" ซ้ำสองครั้ง (ไม่แสดงค่า "ถึง") | `BLRepCreditNoteDateDlg.cpp:249`, `BLRepCreditNoteSDateDlg.cpp:315` |
| 3 | 5236: เงื่อนไขเงินต่อ SQL สองที่ — เติม `BETWEEN ...` ลงท้าย `WHERE` โดยไม่มีชื่อคอลัมน์ (`strSQL += temp`) และสร้าง `HAVING SUM(AMOUNT)` พร้อมกัน ; ถ้าใส่ทั้งจำนวนและเงินจะได้คำว่า `HAVING` ซ้ำสองครั้ง ; ผลคือ SQL ผิดไวยากรณ์เมื่อกรองด้วยเงิน (ยังไม่ได้รันทดสอบกับฐานข้อมูล — อนุมานจากการอ่านโค้ด) | `BLRepCreditNoteSDateDlg.cpp:343-348` |
| 4 | 5236 และ 5244 โหมดแผนก: หัวคอลัมน์ "จำนวนเงิน"/"ส่วนลด" สลับตำแหน่งกับข้อมูล — ช่อง "จำนวนเงิน" แสดงค่าส่วนลด และช่อง "ส่วนลด" แสดงค่าจำนวนเงิน (ข้อมูลใส่คอลัมน์ 5 = ส่วนลด, 6 = จำนวนเงิน แต่หัวคอลัมน์ 5 = `AMOUNT`, 6 = `DISCOUNTM`) | `BLRepCreditNoteSDateView.cpp:223-224,546-547` ; `BLRepDebitNoteSDateView.cpp:208-209,525-526` |
| 5 | ใบลดหนี้ "แบบสรุป" ไม่เช็ค `CauseType` แต่ "แบบแจกแจง" เช็ค — สองรายงานให้จำนวนรวมไม่ตรงกัน | ดู C3 |
| 6 | ป้ายเงื่อนไขโครงการพิมพ์ด้วย key ของแผนก | `BLRepCreditNoteDateDlg.cpp:211-212` |
| 7 | `UPDATE` แก้ข้อมูลใบลดหนี้ในกล่องเงื่อนไข | ดู C3 |
| 8 | เงื่อนไข "จาก > 0" ทำให้กรองจำนวน/เงินที่เป็น 0 หรือติดลบไม่ได้ | `BLRepCreditNoteDateDlg.cpp:244,262` |
| 9 | ชื่อรายงานในกล่องเงื่อนไข (caption) ของ 5236 ตรงกับ 5232 ("รายงานใบรับคืนสินค้า - ตามวันที่") และของ 5244 ตรงกับ 5240 — ทั้งที่เป็นรายงานคนละแบบ | `BCRepBilling.rc:4150,4342,4534,4726` |

---

### กลุ่ม D. รายงานจัดอันดับ / กำไรขั้นต้น / วิเคราะห์ยอดขายรายสินค้า (5248, 5259, 5184)

| menu id | ชื่อ (Champ) | handler (`MainFrm.cpp`) | view class (ไฟล์, บรรทัด) | dialog (IDD : rc, caption) | key รายงาน | ตาราง |
|---|---|---|---|---|---|---|
| 5248 | รายงานจัดอันดับยอด | `OnBillsysAnalyrep61` :5343 | `CBL_Rep_RankByItemView` (`bl_rep_rankbyitemview.cpp`, 562) | `IDD_BL_REP_QUEUELEVEL_ITEM` 377x158 : `BCRepBilling.rc:4957` "รายงานจัดอันดับยอดขาย-ตามสินค้า" | 35601 (ชื่อ dialog ใช้ key เดียวกัน `bl_rep_rankbyitemdlg.cpp:24`) | `BCARInvoiceSub` + `BCCreditNoteSub` + `BCDebitNoteSub1` |
| 5259 | รายงานกำไรขั้นต้น(เก่า) | `OnBillSysAnalyRep71` :6276 | `CBLRepProfitDocDateView` (`blrepprofitdocdateview.cpp`, 811) | `IDD_BL_REP_PROFIT_DOCDATE` 348x211 : `BCRepBilling.rc:382` "รายงานกำไรขั้นต้น - ตามวันที่" | 35701 | `BCARInvoiceSub` (+ `BCCreditNoteSub`, `BCDebitNoteSub1`, `bcitemstandardcost`) |
| 5184 | รายงาน Product Movement Analysis | `OnBillsysAnalyrep10` :5247 (คอมเมนต์ที่ handler คัดมาผิดเป็น "รายงานเปรียบเทียบยอดขายสินค้า12เดือน") | `CBLRepBCArInvoiceView` (`BLRepBCArInvoiceView.cpp`, 695) | `IDD_BL_REP_BCARINVOICE` 415x314 : `BCRepBilling.rc:8274` caption **"รายงานใบจ่ายเงินมัดจำ"** (ไม่ตรงเนื้อหา) | 35120 | `BCARInvoiceSub` + `BCARInvoice` |

ชื่อเมนู 5184 "Product Movement Analysis" ไม่ใช่รายงานความเคลื่อนไหวสต๊อก — เนื้อหาคือการจัดอันดับสินค้าตามยอดขาย/กำไร (เหมือน 5248 แต่มีสัดส่วนสะสม) ; BC มีเมนู "รายงานเคลื่อนไหวสินค้า" (`product-movement` -> `/report/reportstockmovement`, `menu-data.ts:539`) ซึ่งเป็นรายงานสต๊อกคนละเรื่อง — ห้ามจับคู่ชื่อ.

#### D1. 5248 รายงานจัดอันดับยอด (จัดอันดับสินค้าตามจำนวน/มูลค่า)

ช่องเงื่อนไข (ลำดับตาม rc `BCRepBilling.rc:4957-`):

| # | ป้าย (ไทยแท้จาก rc) | ชนิด | SQL / พฤติกรรม | หลักฐาน |
|---|---|---|---|---|
| 1 | พิมพ์จากวันที่: / พิมพ์ถึงวันที่: | วันที่ | `docdate BETWEEN` (จากอย่างเดียว = `=`) ; แทนที่ placeholder `##DATE_CONDITION##` ใน SQL ทั้ง 3 ส่วน | `bl_rep_rankbyitemdlg.cpp:226-246` |
| 2 | พิมพ์จากรหัสสินค้า: / พิมพ์ถึงสินค้า: | รหัสสินค้า | `itemcode BETWEEN` (`##ITEM_CONDITON##`) | `:249-266` |
| 3 | พิมพ์จากหน่วยนับ: / พิมพ์ถึงหน่วยนับ: (ป้ายช่อง "จาก" ใน rc เขียนว่า "พิมพ์ถึงหน่วยนับ:" ซ้ำ — คัดลอกผิด) | combo `BCItemUnit` | `unitcode BETWEEN` (`##UNIT_CONDITON##`) | `:177-178,268-284` ; rc ป้าย `TO_UNITCODE` ซ้ำ 2 ช่อง |
| 4 | พิมพ์เรียงตาม: (`SORT_BY`) | combo 2 ค่า | 0 = จำนวน (`SORT_QTY`), 1 = มูลค่า (`SORT_AMOUNT`) — ไทยจริงใน `myresource` | `:184-186` |
| 5 | พิมพ์เรียงจาก: (`ORDER_BY`) | combo 2 ค่า | 0 = `SORT_ASC` (น้อยไปมาก), 1 = `SORT_DESC` (มากไปน้อย) | `:180-182` |

SQL: UNION 3 ส่วน (ตัวพิมพ์เล็กทั้งหมด) ต่อ `itemcode, unitcode` : (4) ใบขาย `SUM(qty), SUM(netamount), SUM(sumofcost)` จาก `bcarinvoicesub` (`transstate=0 AND iscancel=0 AND itemcode<>'.' AND stocktype<>7` + `posstatus=0` เมื่อ `PosRepSeparate<>1`) ; (5) ลดหนี้ `SUM(discqty*-1), SUM(netamount*-1), SUM(sumofcost*-1)` จาก `bccreditnotesub` (+ `ispos=0`) ; (6) เพิ่มหนี้ `SUM(discqty), SUM(netamount), SUM(sumofcost)` จาก `bcdebitnotesub1` — ผลคือ มูลค่า/ต้นทุน = ขาย − ลดหนี้ + เพิ่มหนี้ (`bl_rep_rankbyitemdlg.cpp:41-135`). อีก SQL หนึ่งนับจำนวนสินค้าไม่ซ้ำ (`select distinct itemcode ...`) เพื่อกำหนดขนาดอาร์เรย์ (`:335-372`).

คอลัมน์ (ชุดเดียว ไม่มีหัวกลุ่ม ; `bl_rep_rankbyitemview.cpp:227-234`): 0 `ITEMCODE` รหัสสินค้า ; 1 `ITEMNAME` ชื่อสินค้า ; 2 `UNITCODE` หน่วยนับ ; 3 `ITEMQTY` จำนวน ; 4 `SUM_TOTAL` มูลค่า ; 5 `COST_ITEM` ต้นทุน ; 6 `EXCHANGE_PROFIT` กำไร(ขาดทุน) ; 7 `"%"` (ตัวอักษร `%` ฝังในโค้ด ไม่ผ่าน key). ไทยจากคอมเมนต์โค้ดตรงบรรทัดเดียวกัน.

| หัวข้อ | พฤติกรรม | หลักฐาน |
|---|---|---|
| รวมต่อสินค้า | อ่านแถวสรุปต่อ (สินค้า, หน่วยนับ) แล้วแปลงจำนวนเป็นหน่วยเก็บ `BCItem.DefStkUnitCode` ด้วย `OnGetPackingRate` (`UnitType=1`: `q*Rate1/Rate2` ของหน่วยที่ขาย แล้วหารด้วยอัตราของหน่วยเก็บ) และรวมเป็นแถวเดียวต่อสินค้า ; หน่วยนับที่แสดงคือหน่วยเก็บ | `bl_rep_rankbyitemview.cpp:281,359` , `:510-` |
| **ข้อควรระวัง จำนวนคืนหาย** | `OnGetPackingRate` คืนค่า 0 เมื่อ `qty <= 0` — ส่วนลดหนี้ที่ใช้ `discqty*-1` (ติดลบ) จึงถูกตัดเป็น 0 ทั้งจำนวน ; ส่วน "มูลค่า/ต้นทุน" ยังถูกหักตามปกติ -> **จำนวนไม่หักลบตามใบรับคืน แต่มูลค่าหัก** (ผลการอ่านโค้ด ยังไม่ได้รันกับข้อมูลจริง) | `bl_rep_rankbyitemview.cpp:510-517` + `bl_rep_rankbyitemdlg.cpp:57` |
| กำไร | `กำไร = มูลค่า − ต้นทุน` ; `% = กำไร*100/ต้นทุน` (**เทียบกับต้นทุน ไม่ใช่ยอดขาย**) ปัด 2 ตำแหน่ง ; ถ้าต้นทุน = 0 แสดงข้อความ "100.00" | `:351,419-427` |
| จัดอันดับ | เรียงในหน่วยความจำด้วย `qsort`: `m_nSortList` 0 = จำนวนน้อย->มาก, 1 = จำนวนมาก->น้อย, 2 = มูลค่าน้อย->มาก, 3 = มูลค่ามาก->น้อย ; **ไม่มีการจำกัด Top-N** และไม่มีคอลัมน์ "ลำดับที่" | `:382-387` |
| ป้ายเงื่อนไขท้ายรายงาน | ผิดป้าย: เรียงตามมูลค่าน้อย->มาก พิมพ์ข้อความ `SORT_QTY`, มาก->น้อย พิมพ์ `SORT_AMOUNT` (ควรเป็น "มูลค่า...") ; เรียงตามจำนวนพิมพ์ `SORT_ASC`/`SORT_DESC` | `bl_rep_rankbyitemdlg.cpp:296-326` |
| สิทธิ์ | คอลัมน์ 4-6 และ `%` แสดงเมื่อผู้ใช้มีสิทธิ์ `bCanSeeItemCost` ; คอลัมน์กำไรต้องมี `IsShowProfitLoss` ด้วย ; (ยอดรวมท้ายรายงานก็เช่นกัน) | `:406-427,455-460` |
| ยอดรวม | รวมต่อหน้า "SUM_LIST_PAGE n TRANS_TYPE" (`OnAddPageSummery` `:470-482`) ; รวมทั้งสิ้น `SUMTOTALDETAIL_WORD n <คำ>` (จำนวน, มูลค่า, ต้นทุน, กำไร) `:446-461` ; ไม่แสดงผลรวมของ % | ตามระบุ |
| คลิกทะลุ | คอลัมน์ 0 (รหัสสินค้า) -> ข้อมูลหลักสินค้า (`BCITEM.Code`) | `:134-150` |
| โค้ดค้าง | มีบล็อก `nDecreaseWithCN` (หักลดหนี้/เพิ่มหนี้ต่อสินค้าแบบ SQL ย่อยต่อแถว) ตั้งเป็น 0 คงที่ (ปิดใช้ ตั้งแต่ปี 2549) — SQL UNION ใหม่ทำหน้าที่แทน | `:289-350` |

#### D2. 5259 รายงานกำไรขั้นต้น(เก่า) — รายการขายต่อบรรทัดพร้อมต้นทุนและกำไร

ช่องเงื่อนไข (ลำดับตาม rc `BCRepBilling.rc:382`):

| # | ป้าย | ชนิด | SQL | หลักฐาน |
|---|---|---|---|---|
| 1 | พิมพ์จากวันที่: / พิมพ์ถึงวันที่: | วันที่ | `BCArInvoiceSub.DocDate BETWEEN` | `blrepprofitdocdatedlg.cpp:132` |
| 2 | พิมพ์จากสินค้า: / พิมพ์ถึงสินค้า: | รหัส + ปุ่มค้นหา (`Browse2 "ItemTransConfig"`) | `ItemCode BETWEEN` | `:149`, `OnFindStartItem` |
| 3 | พิมพ์จากเลขที่เอกสาร: / พิมพ์ถึงเลขที่เอกสาร: | ข้อความ | `DocNo BETWEEN` | `:165` |
| 4 | พิมพ์จากจำนวน: / พิมพ์ถึงจำนวน: | ตัวเลข | `Qty BETWEEN` (จากอย่างเดียว = `=`) | `:183-190` |
| 5 | ช่องติ๊ก `USE_STANDARD_COST_FOR_SHOW` (ไทยหายใน rc — อนุมาน "ใช้ต้นทุนมาตรฐานในการแสดง") | checkbox | ใช้ต้นทุน = ต้นทุนมาตรฐานล่าสุด × จำนวน แทน `SumOfCost` | `blrepprofitdocdateview.cpp:301-304`, `GetStandardCost` `:799-809` |
| 6 | "Use Bottom Line Discount" (`USE_BTDISCOUNT_FORCALC`) | checkbox | หักส่วนลดท้ายบิล (`btdisc`) ออกจาก `NetAmount` | `:430-433` |
| 7 | "Display sale after cn" (`DISPLAYSALE_WITHCN`) | checkbox | หักจำนวน/มูลค่า/ต้นทุน/ส่วนลดตามใบลดหนี้ที่อ้างบรรทัดนี้ (`invoiceno`, `reflinenumber`) พร้อมหักใบลดหนี้แบบ "ราคาผิด" (ไม่มีต้นทุน/จำนวน 0) และบวกใบเพิ่มหนี้ | `:312-419` |

ไม่มีช่อง "เรียงตาม" — ลำดับตายตัว `ORDER BY BCArInvoiceSub.DocDate, DocNo, ItemCode` (`:198`). พื้นฐาน SQL: `SELECT * FROM BCARInvoiceSub WHERE TransState=0 AND IsCancel=0 [AND PosStatus=0] AND ItemCode<>'.' AND StockType<>7` (`blrepprofitdocdatedlg.cpp:29-31`). ช่องกรองพนักงานขายอยู่ใน dialog อื่น (`IDD_BL_REP_PROFIT_SALECODE`, `BCRepBilling.rc:429`) ไม่ใช่ของรายงานนี้.

คอลัมน์ (`blrepprofitdocdateview.cpp:178-197`): หัวกลุ่ม 0 `DATEP` วันที่ ; รายการ: 0 `DOCNO` เลขที่เอกสาร ; 1 `ITEMCODE` ; 2 `ITEMNAME` ; 3 `UINTCODE` ; 4 `PERPRICE` ราคาขาย/หน่วย ; 5 `ITEMQTY` จำนวน ; 6 `SUMAMOUNT` ยอดรวม (= จำนวน × ราคา) ; 7 `DISCOUNTM` ส่วนลด ; 8 `AMOUNT` จำนวนเงิน ; 9 `CUSTBTLINE_PROMOTION_TYPE` Bottom line discount ; 10 `NETAMOUNT` ยอดสุทธิ ; 11 `PRICE_COST` ราคาต้นทุน ; 12 `EXCHANGE_PROFIT` กำไร(ขาดทุน) ; 13 `PERCENT` เปอร์เซ็น ; 14 `PERCENT1` เปอร์เซ็น (ไทยจากคอมเมนต์โค้ด ; ไทยแท้ของ `PERCENT`/`PERCENT1` ไม่แยกความต่าง ต้องอ่านสูตรด้านล่าง).

| หัวข้อ | พฤติกรรม | หลักฐาน |
|---|---|---|
| กำไร | `กำไร = HomeAmount − ต้นทุน` (ใช้ `HomeAmount` คือมูลค่าสกุลบริษัท ไม่ใช่ `NetAmount`) | `:481` |
| `PERCENT` (คอลัมน์ 13) | `กำไร/ต้นทุน*100` ; ต้นทุน = 0 -> 0 | `:486-488,498` |
| `PERCENT1` (คอลัมน์ 14) | `กำไร/HomeAmount*100` (เทียบยอดขาย) ; HomeAmount = 0 -> 100 | `:490-492` |
| รวมต่อวัน | หัวกลุ่มเป็นวันที่เอกสาร ; ท้ายวันมีแถว `SUMDETAIL_WORD n <คำ>` รวมทุกคอลัมน์ (คำนวณ % ใหม่จากยอดรวม) | `:501-558` |
| สินค้าชุด (`StockType=7`) | โค้ดมีสาขารวมแถวลูก (`itemsetcode`) ของสินค้าชุดแบบที่ 2 แต่ SQL กรอง `StockType<>7` ออกก่อน — สาขานี้ไม่น่าถูกเรียก (ยังไม่ verify) | `:276-293` |
| ต้นทุนมาตรฐาน | `SELECT TOP 1 standardcost FROM bcitemstandardcost WHERE itemcode=... ORDER BY startdate DESC` — ใช้ค่าล่าสุดเสมอ ไม่เทียบกับวันที่เอกสาร | `:799-809` |
| สิทธิ์ | คอลัมน์ 11-14 ต้องมี `bCanSeeItemCost` ; คอลัมน์ 12 (กำไร) ต้องมี `IsShowProfitLoss` | `:530-534,586-589` |
| สกุลเงิน | คูณ `ExchangeRate` ทุกค่า ; บล็อกเลือกสกุลบริษัท/แสดงสัญลักษณ์ถูกคอมเมนต์ทิ้ง — **ยอด "จำนวนเงิน/สุทธิ" จึงคูณอัตราแลกเปลี่ยนเสมอแม้ไม่ใช่รายงานสกุลบริษัท** | `:453-479` |
| คลิกทะลุ | เลขที่เอกสาร -> ใบขาย (`BCArInvoice.DOCNO`) ; รหัสสินค้า -> ข้อมูลหลักสินค้า | `:98,105` |
| ชื่อ "(เก่า)" | เมนูระบุ "(เก่า)" — BC ควรใช้ชื่อรายงานกำไรขั้นต้นปัจจุบัน (ดูหมายเหตุท้าย) | `menuconfig.xml:705-733` |

#### D3. 5184 รายงาน Product Movement Analysis — จัดอันดับยอดขาย/กำไรต่อสินค้า พร้อมสัดส่วน

ช่องเงื่อนไข (`BLRepBCArInvoiceDlg.cpp`):

| # | ป้าย | ชนิด | SQL / พฤติกรรม | หลักฐาน |
|---|---|---|---|---|
| 1 | พิมพ์จากสินค้า / ถึงสินค้า | รหัสสินค้า | `ItemCode BETWEEN` | `:158` |
| 2 | กลุ่มสินค้า จาก–ถึง | lookup | `GroupCode BETWEEN` | `:175,179` |
| 3 | วันที่ จาก–ถึง | วันที่ | `DocDate BETWEEN` (จากอย่างเดียว = `=`) | `:193` |
| 4 | เลขที่ใบขาย จาก–ถึง (ป้ายเงื่อนไข `FROM_SALES_NO`/`TO_SALES_NO`) | ข้อความ + ปุ่มค้นหา | `DocNo BETWEEN` | `:206-216` |
| 5 | พนักงานขาย จาก–ถึง ; แคชเชียร์ จาก–ถึง | lookup `BCSale` ทั้งคู่ | `SaleCode BETWEEN` ; `CashierCode BETWEEN` | `:223-235`, `:241-253` |
| 6 | พิมพ์เรียงตาม: | combo | 0 `SUM_PRICE` ยอดขาย, 1 `NET_PROFIT` กำไร | `:99-101` |
| 7 | พิมพ์เรียงจาก: | combo | `SORT_ASC` / `SORT_DESC` | `:103-104` |

ตัวแปรเรียง `m_nSortList` 0..3 = ยอดขาย น้อย->มาก / มาก->น้อย ; กำไร น้อย->มาก / มาก->น้อย (`:259-279`) ; ป้ายเงื่อนไขของโหมดกำไรใช้ `SORT_QTY`/`SORT_AMOUNT` ผิดป้าย (เช่นเดียวกับ 5248). SQL: `SELECT BCARInvoiceSub.*, BCARInvoice.TaxNo, TaxType AS TypeTax ... WHERE TransState=0 AND IsCancel=0 [AND PosStatus=0] AND ItemCode<>'.' AND StockType<>7 ... ORDER BY ItemCode,UnitCode` (`:27-33,281`).

คอลัมน์ (`BLRepBCArInvoiceView.cpp:234-254`): หัวกลุ่ม (4 ช่อง) `DETAILITEMPRODUCT`, `SALES_BEHIND_DISC` (ยอดขายหลังหักส่วนลด), (ว่าง), `EXCHANGE_PROFIT` ; รายการ 14 คอลัมน์: `ITEMCODE`, `ITEMNAME`, `UNITCODE`, `ITEMQTY`, `SUM_PRICE` (ยอดขาย), `DISCOUNT_ITEM` (ส่วนลดสินค้า), `INCVAT` (รวม VAT), `NOTINCVAT` (ไม่รวม VAT), `ITEMRATE` (สัดส่วน), `RATE_COLLECT` (สัดส่วนสะสม), `PRICECOST` (ราคาต้นทุน), `BATH` (บาท), `ITEMPERCENT` (เปอร์เซ็น), `ITEMRATE` (สัดส่วน — ซ้ำ key กับคอลัมน์ 8). ป้ายภาษาไทยในวงเล็บมาจากความหมายของ key (อนุมาน — ไทยจริงใน `myresource`).

| หัวข้อ | พฤติกรรม | หลักฐาน |
|---|---|---|
| รวมต่อสินค้า | รวมในหน่วยความจำต่อ `ItemCode` (โครงสร้าง `SALELEVEL_ARINVOICE`) ; จำนวนแปลงด้วย packing ; ถ้า `TaxType=0` (ราคาไม่รวม VAT) และบรรทัดมีเลขผู้เสียภาษี (`TaxNo` ไม่ว่าง) บวก `amount*TaxRate/100` เข้ายอด | `BLRepBCArInvoiceView.cpp:337-342` |
| สัดส่วน | `ITEMRATE` = ยอดสุทธิของสินค้า / ยอดสุทธิรวมทุกสินค้า × 100 ; `RATE_COLLECT` = สะสมตามลำดับที่เรียง ; ยอดสุทธิรวมได้จาก SQL เพิ่มอีกครั้ง | `:299` |
| กำไร | `กำไร = NetAmount − SumOfCost` ; `ITEMPERCENT = กำไร/ยอดสุทธิ × 100` (**เทียบกับยอดขาย** ต่างจาก 5248/5259 ที่เทียบกับต้นทุน) | `:441-446,511-535` |
| คอลัมน์ "รวม VAT" | `INCVAT` = ยอดขายเดียวกับคอลัมน์ `SUM_PRICE`; `NOTINCVAT` = ยอดสุทธิ — ชื่อคอลัมน์ไม่ตรงข้อมูล | `:511-535` |
| ยอดรวม | แถวรวม: ผลรวมทุกคอลัมน์ ; เปอร์เซ็น = `dSumBath/dSumNotVat*100` ; ข้อความ `SUMTOTALDETAIL_WORD n LIST` | `:511-535` |
| คลิกทะลุ | รหัสสินค้า -> ข้อมูลหลักสินค้า | `:142-160` |
| กรอง | `OnFilterReport` นับสินค้าไม่ซ้ำเพื่อตั้งแถบความคืบหน้า | `:196` |

เทียบ 5248 กับ 5184: 5248 รวมลดหนี้/เพิ่มหนี้ใน UNION และเรียงจำนวน/มูลค่า ; 5184 ใช้เฉพาะใบขาย, เรียงยอดขาย/กำไร, มีสัดส่วนและสัดส่วนสะสม และกรองพนักงานขาย/แคชเชียร์/เลขที่เอกสารได้.

---

### กลุ่ม E. รายงานเปรียบเทียบยอดขาย 12 เดือน (5271, 5916)

| menu id | ชื่อ (Champ) | handler | view class (ไฟล์, บรรทัด) | dialog (IDD : rc, caption) | key | สถานะเมนู |
|---|---|---|---|---|---|---|
| 5271 | รายงานเปรียบเทียบยอดขาย12เดือน | `OnBillsysAnalyrep8` `MainFrm.cpp:5259` (เปิดฟอร์ม `m_pBLRepCMP12MonthFrmSheetOrder`) | `CBL_Rep_Compare12MonthView` (`BL_Rep_Compare12MonthView.cpp`, 555) | `IDD_BL_REP_CMP12MONTH` 365x190 : `BCRepBilling.rc:5164` "รายงานเปรียบเทียบยอดขาย12เดือน" | 35800 (แบบแสดงรายการย่อย 35805) | ใช้งานได้ |
| 5916 | รายงานเปรียบเทียบยอดขายสินค้า 12 เดือน(ตามลูกค้า,สินค้า) | **ไม่พบ handler** — `ID_BILLSYS_ANALYREP_303 = 5916` (`resource.h:4677`) ไม่มี `ON_COMMAND`/`CreateOrActivateFrame` ใน `MainFrm.cpp` ; view ถูกเพียง `#include` ใน `BCRepBillingDll.h:213-215` | `CBLRepCompare12MonthByArItemView` (`BLRepCompare12MonthByArItemView.cpp`, 622) | `IDD_BL_REP_COMPARE_ARITEM` 382x175 : `BCRepBilling.rc:8647` caption "รายงานการขออนุมัติซื้อสินค้า - ตามโครงการ" (ไม่ตรงเนื้อหา) | 31103 | **เมนูมี แต่ไม่ได้ต่อสาย** (กดแล้วไม่เปิดรายงาน — อนุมานจากการไม่พบ handler) |

#### E1. 5271 ช่องเงื่อนไข (`BL_Rep_Compare12MonthDlg.cpp`)

| # | ป้าย | ชนิด | SQL / พฤติกรรม | หลักฐาน |
|---|---|---|---|---|
| 1 | เลือกปี (`AT_YEAR`) | ตัวเลข 4 หลัก ค่าเริ่มต้น = ปีปัจจุบัน (พ.ศ. ถ้า `YearType=0`) | `Year(DocDate) = ปี` (ถ้า `YearType=0` ลบ 543) ; ข้อความเงื่อนไขท้ายรายงาน "AT_YEAR <ปี>" | `:100-106,149-153` |
| 2 | พิมพ์จากรหัสสินค้า / ถึงสินค้า | รหัสสินค้า | `ItemCode BETWEEN` (จากอย่างเดียว = `=`) | `:157-174` |
| 3 | กลุ่มสินค้า จาก–ถึง (`FROM_ITEMGROUP`/`TO_ITEMGROUP` — ไทยหายใน rc ; key เดียวกันที่ `rc:5017,5019` = "พิมพ์จากกลุ่มสินค้า:" / "พิมพ์ถึงกลุ่มสินค้า:") | lookup | `GroupCode BETWEEN` | `:176-191` |
| 4 | ยี่ห้อ จาก–ถึง (`FROM_BRAND`) | lookup | `BrandCode BETWEEN` | `:193-209` |
| 5 | แผนก จาก–ถึง (`FROM_DEPART`/`TO_DEPART`) | lookup | `DepartCode BETWEEN` | `:211-226` |
| 6 | ช่องติ๊ก "แสดงรายการย่อย" (`IDC_CK_SHOWDETAIL`) | checkbox ค่าเริ่มต้นไม่ติ๊ก | ติ๊ก = รายงานรายสินค้า × 12 เดือน (แนวนอน) ; ไม่ติ๊ก = สรุป 12 เดือนรวม (แนวตั้ง) | `:41,60,230` ; `BL_Rep_Compare12MonthView.cpp:23` |
| — | combo เรียงตาม/เรียงจาก | ซ่อน (`NOT WS_VISIBLE`) ไม่ใช้งาน | — | rc `:5164-` |

ไทยของป้ายหลัก ("เลือกปี", "แสดงรายการย่อย") มาจากข้อความในไฟล์ rc ; ป้าย 3 ป้าย 4 ป้าย 5 อนุมานจาก key.

SQL ฐาน (แบบแสดงรายการย่อย): `SELECT ItemCode, Month(DocDate) AS Month_Date, ExchangeRate, SUM(Qty*PackingRate1/PackingRate2) AS SumQty, SUM(NetAmount) AS SumNetAmount FROM BCArInvoiceSub WHERE TransState=0 AND IsCancel=0 [AND PosStatus=0] AND ItemCode<>'.' AND StockType<>7 AND Year(DocDate)=ปี [เงื่อนไขช่องกรอก] GROUP BY ItemCode, Month(DocDate), ExchangeRate ORDER BY ItemCode, Month(DocDate)` (`:25-39,149-227`) — เฉพาะใบขาย ไม่รวมลดหนี้/เพิ่มหนี้. แบบสรุป: 12 คำสั่งย่อยต่อด้วย `UNION` จากตารางตัวช่วย `BCConfigurations` ต่อเดือน ได้ `YearAt, MonthAt, SaleAmount (SUM NetAmount ใบขาย), CNAmount (SUM NetAmount ลดหนี้, ItemType<>7), DNAmount (SUM NetAmount เพิ่มหนี้, ItemType<>7)` โดยเงื่อนไขช่องกรอก (สินค้า/กลุ่ม/ยี่ห้อ/แผนก) ใส่ซ้ำให้ทั้งสามตาราง (`:230-277`) ; **คำสั่งย่อยของแบบสรุปไม่กรอง `TransState`/`PosStatus`** (กรองแค่ `IsCancel=0`) ต่างจากแบบรายการย่อย.

#### E2. 5271 คอลัมน์และสูตร

| โหมด | คอลัมน์ | การจัดแถว | หลักฐาน |
|---|---|---|---|
| แบบสรุป (ไม่ติ๊ก) แนวตั้ง | `MONTH` เดือน ; `YEAR` ปี (พ.ศ./ค.ศ. ตาม `YearType`) ; `SALE_AMOUNT`+`CB_INVOICETYPE_1` มูลค่าขายสินค้า ; `DIFF_AMOUNT` มูลค่าลดหนี้ ; `ADD_AMOUNT` มูลค่าเพิ่มหนี้ ; `NETAMOUNT` ยอดสุทธิ | 12 แถว (ม.ค.-ธ.ค.) + แถวรวม "TOTAL_SUM_AMOUNT" ; `NETAMOUNT = ขาย − ลดหนี้ + เพิ่มหนี้` | `BL_Rep_Compare12MonthView.cpp:177-216,268,289-302` |
| แบบรายการย่อย (ติ๊ก) แนวนอน | `ITEMCODE`, `ITEMNAME`, `UNITCODE`, `JANUARY` ... `DECEMBER` (12 คอลัมน์), `SUMAMOUNT` รวม (รวม 16 คอลัมน์) | ต่อสินค้า 2 แถว: แถวจำนวน (แปลงหน่วยด้วย `OnGetQtyPacking` : `qty*(Rate2/Rate1)`) และแถวจำนวนเงิน (ป้าย `SUM_TOTAL` ในคอลัมน์ชื่อ) ; ท้ายรายงาน `SUMTOTALDETAIL_WORD n <คำ>` ; จำนวนเงินคูณ `ExchangeRate` เมื่อ `MultiCurrency=1` และ `bHomeCurrencyReport=1` | `:218-250,399-409,531-553` |

คอลัมน์เดือนเป็นชื่อเดือนภาษาอังกฤษ (key `JANUARY`..`DECEMBER` — ไทยจริงอยู่ใน `myresource`). `OnGetQtyPacking` ไม่กันกรณี `Rate1 = 0` (หารศูนย์) (`:531-553`).

#### E3. 5916 เปรียบเทียบยอดขายสินค้า 12 เดือน ตามลูกค้า/สินค้า

เหตุผลที่ยังเขียนไว้: ชื่อเมนูอยู่ใน `menuconfig.xml` และโค้ด view/dialog ครบ แต่เมนูไม่มีตัวรับคำสั่ง — BC ไม่ควรทำจอนี้โดยถือเป็นรายงานใช้งานจริงของ Champ จนกว่าลุงจืดยืนยัน.

| หัวข้อ | รายละเอียด | หลักฐาน |
|---|---|---|
| ช่องเงื่อนไข | ระบุปี ; ลูกค้า จาก–ถึง (`ArCode`) ; กลุ่มสินค้า ; สินค้า ; เดือน จาก–ถึง (combo ม.ค.-ธ.ค. ค่าเริ่มต้น ม.ค.-ธ.ค. ; `m_iMonth` = จำนวนเดือนในช่วง) | `BLRepCompare12MonthByArItemDlg.cpp` |
| SQL | `SELECT ArCode, ItemCode, MONTH(DocDate), SUM(Qty*PackingRate1/PackingRate2) AS SumQty, SUM(Amount) AS SumBeforeTaxAmount` (คอมเมนต์โค้ด "ยอดก่อนภาษี" — ใช้ `Amount` ไม่ใช่ `NetAmount`) + ซับคิวรีสหสัมพันธ์ลดหนี้/เพิ่มหนี้ต่อ (ลูกค้า, สินค้า, เดือน): `SUM(DiscQty*packing)` และ `SUM(Amount)` ; `FROM BCARInvoiceSub WHERE TransState=0 AND IsCancel=0 AND ItemCode<>'.'` (ไม่กรอง `StockType`/`PosStatus`) `GROUP BY ArCode, ItemCode, MONTH(DocDate)` | `...Dlg.cpp:222-229` |
| สูตร | `สุทธิ = ขาย − ลดหนี้ + เพิ่มหนี้` ทั้งจำนวนและจำนวนเงิน ; จำนวนสุทธิแปลงหน่วยอีกครั้ง `(dQty*dPackingRate2)/dPackingRate1` | `...View.cpp:381-392` |
| คอลัมน์ | หัวกลุ่ม `ARCODE` รหัสลูกค้า + `AR_NAME` ; รายการ: `ITEMCODE`, `ITEMNAME`, `CODE_UNIT`, `JANUARY`..`DECEMBER`, `SUMTOTAL` ; ต่อสินค้า 2 แถว (จำนวน, จำนวนเงิน ป้าย `SUM_TOTAL`) ; รวมต่อลูกค้า (`QUANTITY_ITEM`, `SUM_TOTAL`, `SUMDETAIL_WORD`) | `...View.cpp:150-169` |
| ข้อผิดพลาด | caption ของ dialog ไม่ตรงเนื้อหา ("รายงานการขออนุมัติซื้อสินค้า - ตามโครงการ") ; ไม่กรอง `StockType<>7` (ต่างจาก 5271) จึงอาจรวมสินค้าชุดแบบที่ 2 ซ้ำกับรายการลูก ; ซับคิวรีสหสัมพันธ์ต่อกลุ่มทำให้ช้าเมื่อข้อมูลมาก (อนุมาน ยังไม่ได้วัด) | rc `:8647` ; `...Dlg.cpp:222-229` |

---

### กลุ่ม F. Daily Sale By Department (5290, 5291, 5295, 5296, 5293, 5294)

"Department" ในชื่อรายงาน = **กลุ่มสินค้า** (`BCItemGroup`/`GroupCode`) ไม่ใช่ตาราง `BCDepartment` (คอลัมน์ `ITEMGROUP`, `ITEMGROUP_NAME`) ; อ่าน `BCARInvoiceSub` join `BCARInvoice`. 6 รายงานเป็นเมทริกซ์ 3×2: ช่วงเวลา (วัน / วัน+สะสมปี / เดือน+สะสมปี) × ระดับรายละเอียด (กลุ่มสินค้า / กลุ่มสินค้า+รายการสินค้าย่อย).

| menu id | ชื่อ (Champ) | handler (`MainFrm.cpp`) | view class (ไฟล์, บรรทัด) | dialog (IDD : rc) | key | ช่วงเวลา | รายละเอียด |
|---|---|---|---|---|---|---|---|
| 5290 | Daily Sale By Department(รายวัน) | `OnDailyDay` :6528 | `CRepDailySaleDepartment_Day_View` (`RepDailySaleDepartment_Day_View.cpp`, 373) | `IDD_DAILY_SALE_DEPARTMENT_DAY` 242x122 : `rc:5208` "Daily Sale By Department (รายวัน)" | 35802 | วันเดียว | กลุ่มสินค้า |
| 5291 | Daily Sale By Department(รายวันแสดงรายการย่อย) | `OnDailyDayShowItem` :6566 | `CRepDailySaleGroupCode_Day_View` (`RepDailySaleGroupCode_Day_View.cpp`, 483) | `IDD_DAILY_SALE_GROUPCODE_DAY` : `rc:5315` caption "Dialog" | 358021 | วันเดียว | กลุ่ม + สินค้า |
| 5295 | Daily Sale By Department(วัน,ปี) | `OnDailyYear` :6541 | `CRepDailySaleDepartment_DayYear_View` (`...DayYear_View.cpp`, 547) | `IDD_DAILY_SALE_DEPARTMENT_DAY_YEAR` : `rc:5233` "Daily Sale By Department (วัน,ปี)" | 35803 | วัน / สะสมต้นปี | กลุ่มสินค้า |
| 5296 | Daily Sale By Department(วัน,ปี แสดงรายการย่อย) | `OnDailyYearShowDetail` :6576 | `CRepDailySaleGroupCode_DayYearDetail_View` (`...DayYearDetail_View.cpp`, 684) | ใช้ `IDD_DAILY_SALE_DEPARTMENT_DAY_YEAR` ร่วม (`RepDailySaleGroupCode_DayYearDetail_Dlg.h`) | 358031 | วัน / สะสมต้นปี | กลุ่ม + สินค้า |
| 5293 | Daily Sale By Department(เดือน,ปี) | `OnDailyMonth` :6554 | `CRepDailySaleDepartment_MonthYear_View` (`...MonthYear_View.cpp`, 548) | `IDD_DAILY_SALE_DEPARTMENT_MONTH_YEAR` : `rc:5258` "Daily Sale By Department (เดือน,ปี)" | 35804 | เดือน / สะสมต้นปีถึงเดือน | กลุ่มสินค้า |
| 5294 | Daily Sale By Department(เดือน,ปี แสดงรายการย่อย) | `OnDailyMonthShowDetail` :6586 | `CRepDailySaleGroupCode_MonthYearDetail_View` (`...MonthYearDetail_View.cpp`, 669) | ใช้ `IDD_DAILY_SALE_DEPARTMENT_MONTH_YEAR` ร่วม | 358041 | เดือน / สะสมต้นปีถึงเดือน | กลุ่ม + สินค้า |

หลักฐาน `ON_COMMAND` : `MainFrm.cpp:753-758` (`ID_BILLSYS_DAILY_DAY/YEAR/MONTH/DAY_SHOWITEM/YEAR_DETAIL/MONTH_DETAIL`).

#### F1. ช่องเงื่อนไข

| # | ป้าย | 5290/5291 (วัน) | 5295/5296 (วัน,ปี) | 5293/5294 (เดือน,ปี) | หลักฐาน |
|---|---|---|---|---|---|
| 1 | พิมพ์ประจำวันที่: (`AT_DATE`) | มี บังคับกรอก (`INSERT_DATE` เตือน "โปรดกรอกวันที่ด้วย") ; `DocDate = วันที่` | มี บังคับ ; ใช้เป็นวันที่ของ "วัน" และปลายช่วงสะสม | ไม่มี | `RepDailySaleDepartment_Day_Dlg.cpp:160-168` |
| 2 | พิมพ์ประจำเดือน: (`AT_MONTH`) combo ม.ค.-ธ.ค. ค่าเริ่มต้น มกราคม | ไม่มี | ไม่มี | มี | `RepDailySaleDepartment_MonthYear_Dlg.cpp` |
| 3 | ปี: (`YEAR`) | ไม่มี | ไม่มี (ปีมาจากวันที่) | มี | เดียวกัน |
| 4 | กลุ่มสินค้า จาก–ถึง (`FROM_ITEMGROUP`/`TO_ITEMGROUP` ; ไทยหายใน rc ของ dialog นี้ ; key เดียวกันที่ `rc:5017,5019` = "พิมพ์จากกลุ่มสินค้า:" / "พิมพ์ถึงกลุ่มสินค้า:") | มี | มี | มี | `...Day_Dlg.cpp:28-31` |
| 5 | พิมพ์จากสินค้า: / พิมพ์ถึงสินค้า: (ไทยแท้) | มี | มี | มี | rc:5208-5340 |

โครงสร้างการรัน: dialog ดึงรายการกลุ่มสินค้าที่มียอดขาย `SELECT DISTINCT GroupCode FROM BCArInvoiceSub WHERE IsCancel=0 AND TransState=0 AND StockType<>7 AND ItemCode<>'.' [AND PosStatus=0] ... ORDER BY GroupCode` (5290 `Day_Dlg.cpp:28-31`) ; จากนั้น view วนแต่ละกลุ่มเพื่อรวมยอด (`ProcessDetail`) และนับเอกสาร (`ProcessBill`). เงื่อนไขช่องสินค้าส่งผ่าน `m_SQL`; วันที่และกลุ่มใส่ใน view เอง.

#### F2. คอลัมน์ — baseline 5290 และตารางความต่าง

5290 (แถวต่อกลุ่มสินค้า ; `RepDailySaleDepartment_Day_View.cpp:162-170`): 0 `ITEMGROUP` กลุ่มสินค้า ; 1 `ITEMGROUP_NAME` ; 2 `COUNT_DOCNO` จำนวนเอกสารขาย (นับเลขที่เอกสารไม่ซ้ำต่อกลุ่ม) ; 3 `SUM_QTY` รวมปริมาณสินค้า ; 4 `COST_ITEM`+"(1)" ต้นทุน ; 5 `SUM_PRICE`+"(2)" ยอดขาย ; 6 `PROFIT_LOSS`+"(2)-(1)" กำไรขาดทุน ; 7 `Margin Percent(%)` (ตัวอักษรอังกฤษฝังในโค้ด). ไทยของ key มาจากความหมายชื่อ key (อนุมาน — ไทยจริงใน `myresource`).

| รายงาน | คอลัมน์ต่างจาก baseline | หลักฐาน |
|---|---|---|
| 5291 | หัวกลุ่ม = คอลัมน์ 5290 ; ใต้แต่ละกลุ่มมีแถวสินค้า: ว่าง, `ITEMCODE`, `ITEMNAME`, `COUNT_DOCNO`, `QUANTITY_ITEM`, `COST_ITEM`, `SUM_PRICE` (**ไม่มีกำไร/Margin ระดับสินค้า**) ; เรียงสินค้าด้วย `ItemCode` ; แถวรวมกลุ่ม `SUMDETAIL_WORD n TRANS_TYPE` และรวมทั้งสิ้น | `RepDailySaleGroupCode_Day_View.cpp:193,253,259,375` (เรียงสินค้า `:253,259`) |
| 5295 | ทุกเซลล์เป็นข้อความ "วัน/สะสมปี" : `DAY_COUNT_DOCNO_YEAR`, `DAY_SUM_QTY_YEAR`, `DAY_COST_YEAR`, `DAY_SUMPRICE_YEAR`, `DAY_PROFITLOSS_YEAR`, `Margin Percent(Day/YTD)` ; สะสมปี = `01/01/<ปี ของวันที่ที่เลือก>` ถึงวันที่ที่เลือก (ปีปฏิทิน ไม่ใช่ปีบัญชี) | `RepDailySaleDepartment_DayYear_View.cpp:441,514` |
| 5296 | เหมือน 5295 + แถวสินค้า (`ITEMCODE`, `ITEMNAME`, ตามด้วยชุดค่า วัน/สะสมปี เดียวกัน 6 คอลัมน์) | `RepDailySaleGroupCode_DayYearDetail_View.cpp` |
| 5293 | เหมือน 5295 แต่คอลัมน์เป็น `MONTH_COUNT_DOCNO_YEAR`, `MONTH_SUM_QTY_YEAR`, `MONTH_COST_YEAR`, `MONTH_SUMPRICE_YEAR`, `MONTH_PROFITLOSS_YEAR`, `Margin Percent(Month/YTD)` ; สะสม = เดือนที่ 1 ถึงเดือนที่เลือก ของปีที่เลือก (`MONTH BETWEEN 1 AND m_nMonth AND YEAR = m_nYear`) | `RepDailySaleDepartment_MonthYear_View.cpp:458-469` |
| 5294 | เหมือน 5293 + แถวสินค้า | `RepDailySaleGroupCode_MonthYearDetail_View.cpp` |

สูตรร่วม: `ต้นทุน = SUM(SumOfCost)` ; `ยอดขาย = SUM(NetAmount)` ; `กำไร = ยอดขาย − ต้นทุน` ; `Margin = กำไร*100/ยอดขาย` (เทียบยอดขาย) (`5290 View.cpp:216-219`) ; สินค้าไม่มีกลุ่ม แสดงป้าย `NOT_GROUP` ("ไม่ระบุกลุ่มสินค้า") ; จำนวน `SUM(Qty)` ดิบ (ไม่แปลงหน่วยนับ) ; ยอดรวมทั้งสิ้น `SUMTOTALDETAIL_WORD n TRANS_TYPE` (ไม่มีช่อง Margin) ; รวมต่อหน้า `OnAddPageSummery` (5290 `:333-346`) ; คลิกทะลุ: เนื้อหาของ `OnCallTrans` ใน 5290 ถูกคอมเมนต์ทิ้ง (ไม่มีการทะลุ).

#### F3. ข้อผิดพลาดใน Champ ของกลุ่ม F ที่ BC ห้ามคัดลอก

| # | ข้อผิดพลาด | หลักฐาน |
|---|---|---|
| 1 | ปีอธิกสุรทินผิด: `(y%4==0 && y%100==0) || y%400==0` ทำให้ปี 2024/2028 ได้ ก.พ. 28 วัน (และ 1900 ได้ 29 วัน) ; วันสิ้นเดือนเขียนมือในโค้ด | `RepDailySaleDepartment_MonthYear_Dlg.cpp:182-189` |
| 2 | ยอดสะสมปี (ต้นทุน/ยอดขาย/กำไร/Margin) ถูกจัดรูปด้วยทศนิยมของ "จำนวน" (`QtyPoint`) แทน `AmountPoint` | `RepDailySaleDepartment_DayYear_View.cpp:280,285,290,295` |
| 3 | `ProcessDetail` ไม่กรอง `TransState`/`PosStatus` ขณะที่รายการกลุ่มที่ดึงมากรอง — ยอดอาจรวมแถวที่ตัดต่อแล้ว/หลังร้านโดยไม่ได้ตั้งใจ | `RepDailySaleDepartment_Day_View.cpp:263-295` เทียบ `...Day_Dlg.cpp:28-31` |
| 4 | หัวคอลัมน์ `Margin Percent(...)` เป็นข้อความอังกฤษ ไม่ผ่าน key ภาษา | `...Day_View.cpp:169` |
| 5 | ตารางผู้ใช้ออกแบบคอลัมน์ลงทะเบียนผิดชื่อ (`BCAPInvoice`) ในรุ่นแสดงรายการย่อย | `RepDailySaleGroupCode_Day_View.cpp:116-117`, `...DayYearDetail_View.cpp:123-124`, `...MonthYearDetail_View.cpp:132-133` |
| 6 | สะสมต้นปีใช้ปีปฏิทิน (`01/01`) ฝังในโค้ด ไม่รองรับปีบัญชีที่เริ่มเดือนอื่น | `RepDailySaleDepartment_DayYear_View.cpp:441,514` |

---

### กลุ่ม G. รายงานยอดขายแยกเจ้าหนี้รับฝากขาย (5370)

| หัวข้อ | รายละเอียด | หลักฐาน |
|---|---|---|
| ชื่อ / ตำแหน่ง | รายงานยอดขายแยกเจ้าหนี้รับฝากขาย ; handler `OnVenderConsignMent` `MainFrm.cpp:6515` (`ID_BILLSYS_VENDER_CONSIGNMENT`, `MainFrm.cpp:752`) ; key 35806 | `MainFrm.cpp:6515` |
| view / dialog | `CBILLRepVenderConsignMentView` (`BILLRepVenderConsignMentView.cpp`, 587) ; `IDD_BILLSYS_VENDER_CONSIGNMENT` 378x244 : `BCRepBilling.rc:5547` "รายงานยอดขายแยกเจ้าหนี้รับฝากขาย" | — |
| ช่องเงื่อนไข | พิมพ์จากวันที่ / ถึงวันที่ (`DocDate BETWEEN`) ; พิมพ์จากเจ้าหนี้ / ถึงเจ้าหนี้ (`BCITEM.Vendercode`) ; กลุ่มสินค้า (`BCITEM.GroupCode`) ; สินค้า (`BCArInvoiceSub.ItemCode`) | `BILLRepVenderConsignMentDlg.cpp` |
| SQL ฐาน | `BCArInvoiceSub INNER JOIN BCArInvoice INNER JOIN BCITEM WHERE [PosStatus=0] AND BCArInvoice.IsCompleteSave=1 AND IsCancel=0 AND BCArInvoiceSub.TransState=0 AND BCITEM.StockType = 4` (4 = สินค้ารับฝากขาย ตามคอมเมนต์คอมโบสินค้า `icdlgitemmasterpage1.cpp:208`) `ORDER BY BCITEM.Vendercode, ItemCode, DocDate` | `...Dlg.cpp:24-42,203` |
| คอลัมน์ | หัวกลุ่ม: `APCODE` รหัสเจ้าหนี้, `APNAME` ; รายการ 11 คอลัมน์: `ITEMCODE`, `ITEMNAME`, `ITEMGROUP` (แสดง `รหัส/ชื่อ`), `DATE_SALE` วันที่เอกสารขาย, `DOC_SALE` เลขที่เอกสารขาย, `UINTCODE`, `PERPRICE` ราคาขาย/หน่วย, `ITEMQTY`, `DISCOUNTM`, `SUM_AMOUNT` จำนวนเงินรวม, `NETAMOUNT` ; รูปแบบคอลัมน์ชุดที่สองอยู่ในคอมเมนต์ `/* */` (ไม่ทำงาน) | `...View.cpp:166-180` (ใช้งาน), `:181-199` (ปิด ; โค้ดเก่า `:404-587` ก็อยู่ในคอมเมนต์) |
| ยอดรวม | แถวรวมต่อเจ้าหนี้ (`SUMDETAIL_WORD n TRANS_TYPE` : จำนวน/ส่วนลด/จำนวนเงิน/สุทธิ) ; รวมทั้งสิ้น ; สินค้าที่ไม่มีเจ้าหนี้ แสดง `AP_NOT_SPECIFY` ("ไม่ระบุเจ้าหนี้") | `...View.cpp:255-275` |
| ขอบเขต BC | ต้องใช้ข้อมูลหลักสินค้า (รหัสเจ้าหนี้ `Vendercode`, ประเภทสินค้ารับฝากขาย) และระบบขาย — อยู่นอกขอบเขต GL-only ของ BC ตอนนี้ (ระบบ AR/AP/ERP เป็นของอนาคต) | กฎ AGENTS.md "GL-only" |

---

---

## BC เทียบ Champ และสิ่งที่ต้องปรับ

> หมายเหตุขอบเขต (กฎลุงจืด): ตอนนี้ BC ทำเฉพาะบัญชีแยกประเภท (GL) — ระบบขาย/ลูกหนี้/คลังเป็นงานอนาคต หัวข้อนี้จึงเป็น "รายการที่ต้องปรับเมื่อถึงเวลาสร้างโมดูลขาย" ไม่ใช่งานสั่งทำทันที ; ฐานข้อมูล PostgreSQL เท่านั้น (ตารางของ Champ ใน SQL Server เป็นหลักฐานอ้างอิง ไม่ใช่ชื่อที่ BC จะใช้) ; เงินเป็น decimal ไม่ใช้ float (Rule 20) ; ไม่มีรายการเงินเดือน/ภ.ง.ด.1/สปส. ในเมนู Bill ของ Champ ; ทุกข้อเสนอห้ามผูกพฤติกรรมกับรหัสบัญชี

### ภาพรวมสถานะ BC ของเมนู Bill 15 รายการ

ที่มา BC: `frontend/src/lib/menu-data.ts` กลุ่ม `bill-orders` / `bill-payment` / `bill-transactions` (:242-271) ; สถานะ: ทุก route ที่เป็น ERP transaction ถูกนับเป็น "รอพัฒนา" (`frontend/src/lib/menu-screen-status.ts:46-51`)

| Champ | ชื่อเมนู | BC menu id | BC route (บรรทัดใน `menu-data.ts`) | สถานะ BC |
|---|---|---|---|---|
| 5303 | บันทึกใบเสนอราคาสินค้า | `quotation` | `/transaction/quotation` (:244) | รอพัฒนา |
| 5307 | อนุมัติใบเสนอราคาสินค้า | `quotation-approve` | `/sales/quotation-approval` (:245) | รอพัฒนา |
| 5000 | ยกเลิกใบเสนอราคาสินค้า | `quotation-cancel` | `/sales/quotation-cancellation` (:246) | รอพัฒนา (ความหมายต่าง — ดูข้อ 5) |
| 5326 | บันทึกใบสั่งขาย/สั่งจองสินค้า | `sale-order` | `/transaction/saleorder` (:247) | รอพัฒนา |
| 5305 | อนุมัติใบสั่งขาย/สั่งจองสินค้า | `sale-order-approve` | `/sales/order-approval` (:248) | รอพัฒนา (คำอธิบายต่าง) |
| 5308 | ยกเลิกใบสั่งขาย/สั่งจองสินค้า | `sale-order-cancel` | `/sales/order-cancellation` (:249) | รอพัฒนา (คำอธิบายต่าง) |
| 5001 | Flow ใบสั่งจองสินค้า | `sale-reservation-flow` | `/sales/reservations` (:250) | รอพัฒนา (คำอธิบายต่าง) |
| 591 | ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า | `sale-order-date-check` | `/sales/order-dates` (:251) | รอพัฒนา (คำอธิบายต่าง) |
| 5327 | บันทึกใบรับเงินมัดจำ | `receive-deposit` | `/transaction/receivedeposit` (:258) | รอพัฒนา |
| 5329 | บันทึกใบรับเงินล่วงหน้า | `receive-advance` | `/transaction/paidadvance` (:259) | รอพัฒนา |
| 5306 | บันทึกคืนเงินรับล่วงหน้า | `return-advance` | `/transaction/paidadvancerefund` (:260) | รอพัฒนา |
| 5328 | บันทึกขายสินค้า,บริการ | `sale` | `/transaction/sale` (:267) | รอพัฒนา |
| 5330 | บันทึกใบรับคืนสินค้า,ลดหนี้ | `sale-return` | `/transaction/salereturn` (:268) | รอพัฒนา |
| 5331 | บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้) | `debit-note` | `/transaction/debitnote` (:269) | รอพัฒนา |
| 5304 | ปรับปรุงวันที่ส่งของให้ลูกค้า | `sale-delivery-date` | `/sales/delivery-dates` (:270) | รอพัฒนา (คำอธิบายต่าง) |
| (ไม่มีใน Bill ของ Champ) | ใบกำกับภาษีอิเล็กทรอนิกส์ | `etax-invoice` | `/transaction/etax` (:271) | ของ BC เอง — เป็นข้อยกเว้นที่เก็บไว้ตามกฎ (e-Tax) |

ชื่อไทยและลำดับเมนูของ BC ตรงกับ Champ ครบ 15 รายการ ; ความต่างอยู่ที่เนื้อหาจอ ไม่ใช่ที่ชื่อ

### รายงานขาย: Champ 45 รายการ เทียบ BC 7 รายการ

- Champ: ส่วน 6 (ชุดรายวัน/จัดขาย 24 รายการ : 5344 5333 5345 5354 5363 5364 5346 5334 5335 5355 5341 5337 5336 5347 5349 5340 5338 5339 5350 5342 5343 5302 5686 5362) + ส่วน 7 (เชิงวิเคราะห์ 21 รายการ : 5180 5185 5204 5277 5310 5232 5236 5240 5244 5248 5259 5184 5271 5916 5290 5291 5295 5296 5293 5294 5370)
- BC: กลุ่ม `bill-reports` มี 7 รายการ (`menu-data.ts:278-284`) = `sales` `/report/reportdedebisales`, `sales-daily` `/report/reportdedebisalesdaily`, `sales-by-seller` `/report/salesbyseller`, `sales-by-document`, `gross-profit-by-document`, `gross-profit-by-product`, `sale-return-report`
- จับคู่ได้: ชื่อซ้ำแต่เนื้อหาต่าง 2 (`sales-daily` ↔ 5337, `sales-by-seller` ↔ 5302) ; ใกล้เคียงบางส่วน 3-4 (`sale-return-report` ↔ 5340/5338, `sales-by-document` ↔ 5336, `sales` ↔ 5204) ; ที่เหลือไม่มีใน BC ; รายละเอียดรายรายงานอยู่ในตาราง "BC เทียบ Champ" ของส่วน 6 และ 7 ด้านล่าง
- `product-movement` (`menu-data.ts:539`) เป็นรายงานสต๊อก ห้ามจับคู่กับ 5184 (จัดอันดับสินค้าตามยอดขาย/กำไร)

### ความต่างสำคัญที่ BC ต้องปรับ (เรียงตามผลกระทบ)

1. **จอเอกสารขายทุกตัวเป็น "รอพัฒนา" และไม่มี backend ขาย** — ไม่มี package ใบขายใน `backend/internal/` ; `saleinvoice` ใน `backend/pkg/microservice/storage_name.go:90` เป็นของค้างยุคเดิม ; จอที่จะใช้คือ `ErpCrudWorkbench` กลาง (`frontend/src/app/crud/erp-crud-workbench.tsx`) ซึ่งเป็นฟอร์มเอกสาร+รายการสินค้าทั่วไป ต่างจากจอของ Champ ที่มีหัวเอกสาร 2 แท็บ (29-33 ช่องต่อแท็บ) ตารางรายการ 11 คอลัมน์ + คอลัมน์ซ่อนอีกหลายสิบ (ส่วน 4) ปุ่ม F-key และ dialog ย่อย (F3 ภาษี F5 รายละเอียดอื่น F7 รับเงิน F8 มัดจำ F2 บัญชี)
2. **เงินเป็นตัวเลข JS/float** — `ErpDetailItem.qty/price/sumamount` และ `ErpTransactionDoc.totalamount` เป็น `number` (`frontend/src/lib/erp-transaction.ts:11-15,28`) ; รายงานขายฝั่ง Go ใช้ `float64` (`backend/internal/goapi/handlers/sales_report.go:382,446-451`) ; เปอร์เซ็นต์กำไรคำนวณในเบราว์เซอร์ (`frontend/src/lib/erp-reports.ts:596-607`) ขัด Rule 20 และกฎ backend-first → ต้องเป็น decimal string / `NUMERIC` และคำนวณฝั่ง Go ; Champ ใช้ double และ `%f` ใน SQL เช่นกัน (ห้ามลอก)
3. **เลขที่เอกสาร ภาษี และประเภทเอกสารไม่ตรง Champ** — workbench สร้างเลขเป็น `<prefix>-<6 หลักท้ายของ Date.now()>` (`erp-crud-workbench.tsx:173-175`) และตั้ง VAT 7% ตายตัว (:184,:268,:321) ; Champ ใช้กลุ่มเลขเอกสารตามประเภท มีประเภทการขาย 4 แบบ (สด/เชื่อ × สินค้า/บริการ) ประเภทภาษี 3 แบบ (แยกนอก / รวมใน / อัตราศูนย์) เลขที่ใบกำกับแยกจากเลขเอกสาร และอ่านอัตราภาษีจากข้อมูลที่บันทึก ; BC มีแค่ `vattype` กับ `status` 0/1/2 (`erp-transaction.ts:34,36`) ขาดธงสถานะอิสระ `is_cancel`, `is_confirm`, `pay_bill_status`, `bill_balance`, `is_post_gl`
4. **ไม่มีการตรวจวงเงินเครดิตตอนบันทึก** — BC มีเฉพาะรายงาน `ar_credit_limit` (`backend/internal/goapi/handlers/debt_report.go:403`) ; Champ ตรวจตอนบันทึกตามค่าตั้งบริษัท `DebtOverWarning` และสถานะลูกค้า (เตือนอย่างเดียว หรือขออนุมัติผ่าน dialog `CMyConfirmCreditLimit` ซึ่งเลือก อนุมัติเดี๋ยวนี้ / พัก / ไม่บันทึก ได้ และเก็บผู้อนุมัติ `CrAuthorizeMan`) — ใบสั่งขายตรวจตอนบันทึก (`ChmBLFrmSaleOrder.cpp:517-565`) ไม่ใช่ตอนอนุมัติ ; รายละเอียดใน ส่วน 2 หัวข้อ 1.11 และส่วน 4
5. **คำอธิบายหน้าจอของ BC ไม่ตรงพฤติกรรม Champ 5 จุด** (`frontend/src/lib/erp-operations.ts`) — :68 (อนุมัติใบสั่งขายเขียนว่าตรวจวงเงินและกันสต๊อก แต่ Champ อนุมัติเพียงเปลี่ยน `HoldingStatus` 1→0 และ `IsConfirm=1`) ; :77 (ยกเลิกใบสั่งขายเขียนว่าคืนสต๊อก แต่ Champ สร้างเอกสาร `BCSOVoid` ยกเลิกยอดค้างรายแถว) ; :88 (5001 คือการติดตามสายเอกสาร SO → PO → GR → ใบขาย → ยกเลิก พร้อมปุ่มสร้าง PO) ; :96 (591 คือบันทึกวันที่ยืนยัน/ส่งจริงรายแถว) ; :104 (5304 คือปรับ `DeliveryDate`/`DueDate` ของใบขายเชื่อ ไม่ใช่ตารางรถขนส่ง) ; และ [5000] ของ Champ เป็นการตัดยอดค้างรายแถว (สร้าง `BCCancelBill` ชนิด 2) ส่วนการยกเลิกทั้งใบทำด้วย Ctrl+F5 ในจอ 5303
6. **หน้าต่างรับเงินกลาง (F7) และกลไกสถานะเอกสารยังไม่มี** — ใบมัดจำ ใบรับล่วงหน้า และใบขายเงินสด/เชื่อใช้หน้าต่างรับเงินร่วมกัน (เงินสด/โอน+สมุดธนาคาร/เช็ค/บัตร) พร้อมตรวจยอดรับครบก่อนบันทึก (`IsMoneyOk`, `myespeciallyformview.cpp:4266-4318`) ; BC GL มีสถานะร่าง/ผ่านรายการ/กลับรายการ/ยกเลิกเฉพาะสมุดรายวัน ; เอกสารขายที่ผ่าน GL แล้วต้องแก้ด้วยรายการกลับ ไม่แก้/ลบเงียบ (กฎ GL ของ BC) ; ฉบับร่างเท่านั้นที่ลบได้
7. **ความอะตอมของการบันทึก** — Champ commit หัว/รายการก่อน แล้วค่อยตัดสต๊อก ลงภาษีขาย รับเงิน GL และปรับลูกหนี้เป็นขั้นแยก (`ChmFrmBilling.cpp:3800-3861`) ถ้าล้มกลางทางข้อมูลไม่ครบ ; BC ต้องทำทั้งหมดใน `sql.Tx` เดียวพร้อม idempotency key
8. **รายงาน: จอรายงานกลางไม่มีกล่องเงื่อนไข** — Champ ให้เลือกช่วงเลขที่เอกสาร/ลูกค้า/สินค้า/พนักงาน/ประเภท + combo "เรียงตาม" 11-15 โหมดที่เปลี่ยนหัวกลุ่ม ; BC มีเพียงปุ่มช่วงวันที่สำเร็จรูป (`frontend/src/app/report/erp-report-viewer.tsx:51-63`) และไม่มีนิยามตัวกรองต่อรายงานใน `ErpReportConfig` (`erp-reports.ts:11-26`) ; รายงานที่ API พร้อมในกลุ่มขายมีแค่ `sales_by_document` กับ `gross_profit_document` (`erp-reports.ts:530-543`)
9. **ฐานของเปอร์เซ็นต์กำไรใน Champ ไม่เหมือนกันทุกรายงาน** — 5248 และคอลัมน์ `PERCENT` ของ 5259 เทียบกับต้นทุน ; 5184, 5290-5296 และ `PERCENT1` ของ 5259 เทียบกับยอดขาย ; BC `marginpercent` เทียบกับยอดขาย (`erp-reports.ts:602-607`) — ห้ามเปลี่ยนความหมายเงียบ ๆ ให้แยกคอลัมน์ "% เทียบต้นทุน" ; Champ ซ่อนคอลัมน์ต้นทุน/กำไรตามสิทธิ์ (`bCanSeeItemCost`, `IsShowProfitLoss`) — BC ต้องมีสิทธิ์ระดับคอลัมน์
10. **ของที่ BC มีเกิน Champ (ขัดกฎ Strict Parity — ให้ลุงจืดตัดสิน)** — คอลัมน์ "เป้า / % บรรลุเป้า / คอมมิชชั่น" ใน `sales_by_seller` (`erp-reports.ts:169-175`) และคอลัมน์ "เงินโอน" ใน `sales_daily` (:158) ; ส่วนคอมมิชชั่นอยู่ใกล้เรื่องเงินเดือนซึ่งนอกขอบเขตผลิตภัณฑ์

### ข้อผิดพลาดของ Champ ที่ BC ห้ามลอก (รวมจากทุกส่วน)

| จุด | หลักฐาน |
|---|---|
| ราคาต่อหน่วยรวมภาษีคูณ `1.07` ตายตัว | `ChmFrmBillingPage1.cpp:540` ; `ChmBLFrmSaleOrderPage1.cpp:429,519` |
| คำเตือนวงเงินใน `BillCalc` ของใบสั่งขายไม่เคยทำงาน | `ChmBLFrmSaleOrderPage1.cpp` (~725) |
| ตัวกรองพนักงานขายในหน้ายกเลิกใบสั่งขายใช้ผิดคอลัมน์ | `POFrmCancelPoSoView.cpp:404-537` |
| รายงานบางตัวสั่ง `UPDATE` ข้อมูลตอนกดในกล่องเงื่อนไข | `BLRepCreditNoteDateDlg.cpp:141-143` ; `BL_Rep_Compare12MonthDlg.cpp:139-146` |
| 5248: จำนวนรับคืนหายเป็น 0 แต่มูลค่ายังถูกหัก | `bl_rep_rankbyitemview.cpp:510-517` |
| 5236: เงื่อนไขเงินสร้าง SQL ผิดไวยากรณ์ ; หัวคอลัมน์ส่วนลด/จำนวนเงินสลับ (5236 และ 5244 โหมดแผนก) | `BLRepCreditNoteSDateDlg.cpp:343-348` ; `BLRepCreditNoteSDateView.cpp:223-224` |
| 5293: วันสิ้นเดือนฝังค่า ทำให้ปีอธิกสุรทินผิด | `RepDailySaleDepartment_MonthYear_Dlg.cpp:182` |
| ยกเลิกใบเพิ่มหนี้แล้วยอดลูกหนี้อาจไม่ถูกกลับ (เรียก `-1` หลังตั้ง `IsCancel=1` แต่ฟังก์ชันค้นเอกสารที่ `IsCancel=0` ; ต้องส่ง `-2`) — ยังไม่ได้รันพิสูจน์ | `billfrmardeposit.cpp:505` ; `bcprocess.cpp:2349-2365` |
| ยกเลิกใบลดหนี้แล้วยอดใบขายเดิมอาจไม่ถูกคืน — ยังไม่ได้รันพิสูจน์ | `bcprocess.cpp:2016` |
| ข้อมูลเฉพาะลูกค้า Champ บางราย (SCG ID/Coupon/Wallet/Promotion Msg, QARefno, รหัสเครื่อง, integration ขนส่ง/marketplace ใน `OnSaveData`) ไม่ใช่ parity | ส่วน 4 [5328] — ให้ลุงจืดยืนยันก่อนตัด |

### คำถามที่ต้องให้ลุงจืดตัดสินก่อนเริ่มสร้างโมดูลขาย

1. **เปิดโมดูลขายเต็มรูปหรือยังคง GL อย่างเดียว** — ทางที่ BC มีตอนนี้คือบันทึกยอดขายเป็นใบสำคัญ GL ที่มีรายละเอียดภาษีขาย `details.vats` (`backend/internal/goapi/handlers/tax_report.go:89`) ซึ่งไม่มีรายการสินค้า สต๊อก การรับเงินหลายช่องทาง หรือการคุมวงเงิน ; การทำให้ครบเท่า Champ = เปิดโมดูลขาย/ลูกหนี้/คลังเต็มรูป
2. **คำนำหน้าเลขเอกสาร** — BC ใช้ `SR`/`DN` (`erp-transaction.ts:167,178`) แต่ Champ ใช้ `CN`/`DBN` (`billfrminvcreditnote.cpp:53`, `billfrmardeposit.cpp:42`)
3. **รายงานที่ชื่อหรือทางเปิดผิดปกติ** — 5916 ไม่มี handler ใน `MainFrm.cpp` (เมนูไม่ได้ต่อสาย) ; 5259 มีคำว่า "(เก่า)" ไม่ทราบว่ารุ่นปัจจุบันคือรายงานไหน ; 5184 ชื่อเมนูไม่ตรงเนื้อหา — จะทำตามชื่อเมนูหรือพฤติกรรมจริง
4. **ฟีเจอร์ที่ BC มีเกิน Champ** (ข้อ 10 ด้านบน) จะเก็บหรือตัด
5. **ใบกำกับภาษีที่ยกเลิก** — เงื่อนไขการเก็บ/แสดงยังไม่ได้ตรวจกับแหล่งทางการ ต้องตรวจตาม skill `thai-accounting-tax` และ `docs/kms/21-thai-tax-form-references.md` ก่อนออกแบบ (ห้ามเดา)
6. **ถ้อยคำไทยที่ติดป้าย "(อนุมาน)"** — ไทยจริงของ key `GetMyRes` อยู่ในตาราง `myresource` ของ DB Champ ซึ่งไม่อยู่ใน repo ต้องขอตารางจริงก่อนเขียน `backend/assets/language/languages.tsv` (คอลัมน์ th เป็นของจริง ภาษาอื่นใส่ placeholder ตามกฎ 2 ภาษา)

### รายละเอียดรายหน้าจอ (คัดจากแต่ละส่วน)

### ส่วน 1 — BC เทียบ Champ (ส่วนของ ใบเสนอราคา 5303/5307/5000)

**สถานะรวมฝั่ง BC**: ทั้ง 3 เมนูอยู่ในกลุ่ม `bill-orders` ("ใบเสนอราคาและใบสั่งขาย") ของโมดูล `bill` (`frontend/src/lib/menu-data.ts:242`) และจัดเป็น "ยังไม่พร้อม" ตามกฎ `isMenuBackendRetired` (`frontend/src/lib/menu-screen-status.ts:46-49` คืน true สำหรับ `isErpTransactionRoute || isOperationsRoute`) ; ไม่มี handler `/transaction/quotation` ใน Go (`backend/internal/goapi/bootstrap.go:348-351` มีแค่ `/api/transaction/calculate`, `quick-calc`, `validate-payment`, `purchase-history`) ; หน้าจอปัจจุบันเป็น workbench ทั่วไป (`frontend/src/app/crud/erp-crud-workbench.tsx`, 1074 บรรทัด) ที่ยังไม่มีฟิลด์แบบ Champ

**ขอบเขต (กฎเจ้าของ)**: ลูกหนี้/ERP เป็นงานอนาคต ตอนนี้ทำเฉพาะ GL — กลุ่มนี้จึงเป็น "ข้อมูลศึกษา ยังไม่ต้องสร้าง" ; ใบเสนอราคาไม่ลง GL อยู่แล้ว (§1.8) ; ไม่เกี่ยวกับเงินเดือน ; PostgreSQL เท่านั้น ; ชื่อคอลัมน์ BC ที่เสนอเป็น snake_case และเป็น `numeric` ไม่ใช่ `MONEY`/`SMALLMONEY` (DDL Champ ของ `BCQUOTATION`, `BCQUOTATIONSUB`, `BCQUOTATIONSTATUS` ใช้ MONEY — ห้ามลอก)

#### [5303] บันทึกใบเสนอราคาสินค้า

| BC menu id | route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ (BC file:line) |
|---|---|---|---|
| `quotation` (`menu-data.ts:244`, `tx("quotation", ..., "/transaction/quotation")`) | `/transaction/quotation` | **รอพัฒนา** (ไม่มี backend ; ฝั่งหน้าจอแปลง 404 เป็น `module_not_available` ใน `erp-transaction.ts` ~866-880) | (1) ตั้ง backend ใหม่: ตารางหัว+รายการแบบ `quotations` / `quotation_lines` (snake_case, ยอดเงินเป็น `numeric`), เขียนใน `sql.Tx` เดียว (Champ ไม่มี transaction — `ChmFrmQuotation.cpp:1617-1883`) ; BFF อนุญาต module อยู่แล้ว (`frontend/src/app/api/erp-transaction/[...erpPath]/route.ts:10`, ส่งต่อ `:61`) แต่ปลายทาง Go ยังไม่มี. (2) หัวเอกสารใน workbench มีแค่ เลขที่/วันที่/ชื่อคู่ค้าแบบ text/หมายเหตุ (`erp-crud-workbench.tsx:720-775`) — ต้องเพิ่ม: ประเภทภาษี 3 ค่า, ประเภทการเสนอราคา 4 ค่า, ลูกค้าแบบ lookup พร้อมดึงค่าเริ่มต้น (ประเภทภาษี/อัตรา/พนักงาน/ผู้ติดต่อ/ส่วนลด), แท็บ 2 (พนักงาน แผนก โครงการ ผู้ติดต่อ การจัดสรร ยืนราคา วันหมดอายุ/ส่งมอบ/เครดิต พร้อมการผูกวัน↔วันที่ ตอบรับภายใน คำตอบลูกค้า สกุลเงิน/อัตรา ค่าขนส่ง กลุ่มเอกสาร ช่องทาง เรื่อง สิ่งที่ส่งมาด้วย หมายเหตุ 1/2) (3) grid ของ BC มี 6 คอลัมน์ (`erp-crud-workbench.tsx:804-809`: รหัสสินค้า/บาร์โค้ด, ชื่อ, หน่วย, จำนวน, ราคา, จำนวนเงิน) — ขาด ส่วนลดรายแถว (คำส่วนลดคั่น `,` มี `%`), เงื่อนไขรับเอง/ส่งให้, ปุ่มขยายแถว, shadow ที่จำเป็น (ยกเว้นภาษี, สถานะการเสนอราคา, ค่าขนส่งต่อแถว, คลัง/ชั้น ฯลฯ — ตัดที่ไม่ใช่ข้อบังคับตามกฎ "ไม่เพิ่มเกิน Champ") (4) ยอดรวม BC มีแค่ "รวมมูลค่าสินค้า / รวมทั้งสิ้น" (`erp-crud-workbench.tsx:877`, `:889`) — ต้องมี ยกเว้นภาษี, ส่วนลดบิล (คำ+ยอด), ภาษี, รวมภาษี, สุทธิ และคำนวณที่ **backend** (กฎ Backend-First) ตามสูตร §1.4 ; อัตราภาษี 7 ปัจจุบันฮาร์ดโค้ด (`erp-crud-workbench.tsx:168-203`, vatrate ที่ ~176-177) ต้องเปลี่ยนเป็นค่าตั้งของบริษัท/ลูกค้า. (5) เลขที่เอกสาร BC สร้างเป็น `QT-<6 หลักท้ายของ Date.now()>` (`erp-crud-workbench.tsx:168-203`) — ต้องเปลี่ยนเป็นรันตามรูปแบบเอกสาร (กลุ่ม `@`, ปี/เดือน/วัน, `#` เลขรัน, พ.ศ.) และตรวจซ้ำ ; ห้ามคำนวณเลขที่ฝั่ง browser. (6) สถานะรายการ BC มีแค่ draft/approve (`erp-transaction.ts:4-47` status 0/1/2) — ต้องเพิ่มธง ยืนยัน/ยกเลิก/`BillStatus` และป้าย CC/CF/N/0.00 ตาม §1.7. (7) คีย์ลัด F11/Ctrl+N/Ctrl+F/F12/Shift+F5/Ctrl+F5/F5 สลับแท็บ ยังไม่มี. (8) ไม่ต้องทำ: ระบบเชื่อมภายนอก (`HomeBuddy`), คอลัมน์ `SCG*`, ฟอร์ม legacy |

#### [5307] อนุมัติใบเสนอราคาสินค้า

| BC menu id | route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ (BC file:line) |
|---|---|---|---|
| `quotation-approve` (`menu-data.ts:245`, group `approval`) | `/sales/quotation-approval` | **รอพัฒนา** — `isApprovalApiReady` เป็น false เพราะ `APPROVAL_KINDS` มีแค่ `pr_approval` (`frontend/src/lib/erp-operations.ts:212-214`) ; เทสต์ยืนยัน `isMenuDataPending("/sales/quotation-approval")` เป็น true (`menu-screen-status.test.ts:82`) ; config `qt_approval` อยู่ `erp-operations.ts:45-62` (documentType "QT") | (1) หน้าจอแบบรายการ+เงื่อนไขกรองช่วง: เลขที่ใบเสนอราคา, วันที่, รหัสลูกค้า, แผนก, พนักงาน, "เลือกทั้งหมด" (§3) — ปัจจุบัน config ทั่วไปไม่มีชุดตัวกรองนี้. (2) คอลัมน์: เลขที่, วันที่, ลูกค้า, เรื่อง, ภาษี, ยอดสุทธิ (แก้ให้หัวคอลัมน์ตรงกับค่า — Champ ติดป้าย TAXRATE แต่โชว์ยอดภาษี). (3) การอนุมัติ = ตั้ง `confirmed_by`, `confirmed_at`, `is_confirmed=true` ต่อใบที่ติ๊ก ใน `sql.Tx` เดียว และ **ต้องเพิ่ม guard ที่ Champ ไม่มี**: ตรวจสิทธิ์ยืนยัน, ตรวจงวดปิด, `WHERE is_confirmed=false AND is_cancelled=false` ตอนอัปเดต, ตัวกรองแบบ parameter ไม่ต่อสตริง. (4) ดับเบิลคลิกแถวเปิดใบใน [5303]. (5) เพิ่ม `APPROVAL_KINDS`/ตัวบอกความพร้อมของ API เมื่อ backend เสร็จ (`erp-operations.ts:212-214`) และถอดออกจากรายการ retired (`menu-screen-status.ts:46-49`). (6) log ผู้ใช้ตาม audit trail |

#### [5000] ยกเลิกใบเสนอราคาสินค้า

| BC menu id | route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ (BC file:line) |
|---|---|---|---|
| `quotation-cancel` (`menu-data.ts:246`, group `approval`) | `/sales/quotation-cancellation` | **รอพัฒนา** (เหตุผลเดียวกับ [5307]: `erp-operations.ts:212-214` ; config `qt_cancellation` ที่ `erp-operations.ts:45-62`) | (1) **ความหมายต่างจากชื่อเมนูที่คนคาดหวัง**: Champ [5000] คือ "ตัดยอดคงเหลือเป็นจำนวน" ที่ออกเป็นเอกสารของตัวเอง ไม่ใช่ยกเลิกทั้งใบ (ยกเลิกทั้งใบอยู่ที่ Ctrl+F5 ใน [5303]) — ต้องตัดสินใจกับลุงจืดว่า BC ทำตาม Champ (แยก 2 เส้นทาง) หรือไม่ ; ห้ามเพิ่มเอง. (2) ต้องมี: ตัวกรอง (ช่วงวันที่ เลขที่ ลูกค้า พนักงาน สินค้า เลือกทั้งหมด), grid 11 คอลัมน์ (§4) พร้อมแก้จำนวนที่ยกเลิก, เหตุผล (`BCCancelReason` → ตารางเหตุผลยกเลิก), เลขที่เอกสารรัน (`CQO`), ตรวจ `VoidQty <= RemainQty`. (3) backend: ตารางหัว/รายการยกเลิก (`cancel_type=2`) + ปรับ `remain_qty` ของบรรทัดใบเสนอราคา และคำนวณ `bill_status` ใหม่ (=1 เมื่อคงเหลือรวม <=0) พร้อมย้อนกลับเมื่อแก้/ลบ ใน `sql.Tx` เดียว ; ต้องเพิ่ม guard ที่ Champ ไม่มี: งวดปิด, ใบเสนอราคาต้องยืนยันแล้ว/ยังไม่ยกเลิก, ลบก่อนเพิ่มด้วยเลขที่เดิมที่ถูกต้อง. (4) เทียบกับ [5303]: ฟิลด์ `remain_qty` และ `bill_status` ต้องเก็บตั้งแต่ตอนบันทึกใบเสนอราคา (`RemainQty=Qty`, §1.8) |

#### ข้อสรุปสำคัญสำหรับผู้พัฒนา

- ลำดับที่ควรทำเมื่อเริ่มโมดูลขาย: [5303] ก่อน (เป็นแหล่งข้อมูลของ [5307]/[5000]) → [5307] → [5000]
- ที่ยัง **ไม่ verify**: การผูก control→คอลัมน์ของ Page2 รายตัว (§1.2) · สิทธิ์ซ่อนของจอ [5307] · ข้อความ `EXPCREDIT` ฉบับเต็ม · ความหมายพารามิเตอร์ของ `SetComboData(int)` · ไทยของป้าย [5000] ทั้งหมด · ไทยของค่าใน combo (`CB_TAXTYPE_*`, `CB_DOCTYPE_*`, `STATUS_CUSTOMER*`, `OWN_RECEIVE`, `SEND_TO`) และรายการ "Key"

### ส่วน 2 — BC เทียบ Champ (ส่วนของ ใบสั่งขาย/สั่งจอง อนุมัติ ยกเลิก Flow ตรวจวันที่ ปรับวันที่ส่งของ)

ภาพรวม BC: ทั้ง 6 เมนูอยู่ใน menu-data แต่ "รอพัฒนา" ทั้งหมด — `isMenuBackendRetired` คืนค่า true สำหรับ route ERP transaction/operations (`frontend/src/lib/menu-screen-status.ts:46-48`) และ `main-menu-screen.tsx:2817` แสดงการ์ด "รอพัฒนา" ; ไม่มี handler/ตาราง backend ของใบสั่งขาย (grep ใน backend พบแค่ key ใน `languages.tsv` และ `goapi/myglobal/utils.go`, `pkg/microservice/storage_name.go`) ; ป้ายเมนูตรงกับ Champ แล้ว (menu-data.ts:130-132 search aliases)

| รหัสเมนู Champ | BC menu id + route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ (เมื่อถึงคิวทำ ERP) |
|---|---|---|---|
| 5326 บันทึกใบสั่งขาย/สั่งจองสินค้า | `sale-order` → `/transaction/saleorder` กลุ่ม "transaction" (`frontend/src/lib/menu-data.ts:247`) | รอพัฒนา (ต่าง) | (1) เปลี่ยนจาก `ErpCrudWorkbench` ทั่วไป (`frontend/src/lib/erp-transaction.ts:138-147`, moduleKey sale_order, apiPath sale-order, prefix "SO") เป็นหน้า 2 แท็บแบบ Champ: หน้า 1 = หัว 8 ช่อง + grid 10 คอลัมน์ + ท้ายบิล 6 ช่อง, หน้า 2 = 34 ช่อง "รายละเอียดอื่นๆ" ตาม 1.2-1.4 ; (2) แยก 2 ตัวแปร: ประเภทเอกสาร SOStatus (สั่งขาย/สั่งจอง) กับ BillType 4 ค่า (สินค้าสด/เชื่อ, บริการสด/เชื่อ) ; (3) ปุ่ม Ctrl+F3 ดึงใบเสนอราคา (กรอง ลูกค้า/IsConfirm/ค้างรับ/ประเภทภาษี/BillType, ตัวล็อกตาม option SONoEdit*) ; (4) ตรวจวงเงินเครดิต 3 ทาง (ไม่/อนุมัติทันที/พักรออนุมัติ) + ตัวบล็อก DebtOverWarning==3 ; (5) พักบิล F8/F9/Ctrl+F9 ; (6) ตารางที่ควรมี (snake_case, PostgreSQL): `sale_orders`, `sale_order_lines` ; คอลัมน์ตัวอย่าง `holding_status`, `is_confirm`, `credit_authorize_man`, `remain_qty`, `po_qty`, `quotation_ref_no` ; เงิน = `NUMERIC(19,4)` ห้าม float ; ห้ามลอก `1.07` hard-code และ `RefType=1` ทุกแถว (1.12) ; **ขัดกฎเจ้าของ**: ขอบเขตตอนนี้คือ GL อย่างเดียว (AR/ERP เป็นโมดูลอนาคต) — เสนอไว้เป็น blueprint ไม่ใช่งานเร่งด่วน |
| 5305 อนุมัติใบสั่งขาย/สั่งจองสินค้า | `sale-order-approve` → `/sales/order-approval` กลุ่ม "approval" (`menu-data.ts:248`) ; คำอธิบายใน `frontend/src/lib/erp-operations.ts:64-69` (`so_approval`, doc type "SO", บรรทัด 68 "ตรวจสอบวงเงินสินเชื่อคงเหลือ และกันสต็อกสินค้า") | รอพัฒนา (ต่าง) | (1) ความหมายของ Champ = เปลี่ยน HoldingStatus 1→0 + บันทึก ApproveCode/ApproveDateTime + IsConfirm=1 ของใบที่ถูกพักเพราะเกินวงเงิน (การตรวจวงเงินทำที่ตอนบันทึก 5326 ไม่ใช่ที่หน้านี้) — แก้คำอธิบายบรรทัด 68 ให้ตรง ; (2) ช่องกรอง: เลขที่ใบสั่งขาย จาก–ถึง, วันที่ จาก–ถึง, รหัสลูกค้า, รหัสแผนก, รหัสพนักงาน, เลือกทั้งหมด ; ปุ่ม "ประมวลผลเงื่อนไข" (F5) / "อนุมัติ" (F11) ; คอลัมน์ 5 ตัวตามหัวข้อ 2 ; (3) `APPROVAL_KINDS` ใน `erp-operations.ts:208-210` มีเฉพาะ `pr_approval` → ไม่มี API อนุมัติใบสั่งขาย (comment บรรทัด 206-207 ยอมรับว่าไม่มี API) ; (4) ข้อควรระวังจาก Champ: อนุมัติตั้ง IsConfirm=1 ทำให้ใบแก้ไม่ได้ ควรกำหนดค่าให้เป็นระบบเดียว (อย่าแยก 1 กับ 2) ; รอลุงจืดตัดสินใจ |
| 5308 ยกเลิกใบสั่งขาย/สั่งจองสินค้า | `sale-order-cancel` → `/sales/order-cancellation` กลุ่ม "approval" (`menu-data.ts:249`) ; `erp-operations.ts:73-78` (`so_cancellation`, บรรทัด 77 "ยกเลิกคำสั่งขายและคืนยอดสต็อกที่จองไว้") | รอพัฒนา (ต่างหลักการ) | (1) Champ ไม่ยกเลิกทั้งใบที่หน้านี้ — สร้างเอกสารยกเลิกยอดค้าง (`sale_order_voids` + `sale_order_void_lines`) ต่อแถวตามจำนวนที่ยกเลิก (ไม่เกินยอดค้าง) พร้อมเหตุผล (`BCCancelReason`) ; การยกเลิกทั้งใบเป็น Ctrl+F5 ในหน้า 5326 (บล็อกเมื่อยืนยันแล้ว/ถูกอ้างอิง/โพสต์ GL แล้ว) — แก้คำอธิบายบรรทัด 77 ; (2) ช่องกรอง: เลขที่เอกสาร, รหัสลูกค้า, วันที่, รหัสพนักงานขาย (จาก–ถึง) ; grid 11 คอลัมน์ตามหัวข้อ 3 ; (3) ห้ามลอกบั๊ก Champ ที่กรองรหัสพนักงานขายด้วย ArCode ; ใช้ตัวกรองให้ถูก ; (4) route `/sales/order-cancellation` ถูกจัดกลุ่ม "approval" แต่ Champ เป็นหน้ากรอกเอกสาร (มีเลขที่เอกสาร รันเลข) ควรตรวจกลุ่มสิทธิ์ |
| 5001 Flow ใบสั่งจองสินค้า | `sale-reservation-flow` → `/sales/reservations` กลุ่ม "transaction" (`menu-data.ts:250`) ; `erp-operations.ts:84-89` (`sales_reservations`, บรรทัด 88 "ติดตามสถานะสินค้าที่ถูกจอง… กำหนดวันเบิกส่งมอบ") | รอพัฒนา (ต่าง) | (1) Champ = ติดตามสายเอกสาร SO → ใบสั่งซื้อ → ใบรับสินค้า → ใบขาย → ใบยกเลิก (รายการซ้าย) + ยอดสั่งขาย/สั่งซื้อ/GP/ขาย/ยกเลิก ต่อสินค้า (รายการขวา) ไม่ใช่ "วันเบิกส่งมอบ" — แก้บรรทัด 88 ; (2) ปุ่ม "สร้างใบสั่งซื้อ" และ "คำนวณยอดสั่งจองที่สามารถทำ PO ได้" ; เมนูคลิกขวา 5 รายการตามหัวข้อ 4 ; (3) `AGENTS.md` (หัวข้อ "ตัดเมนูที่ Champ ไม่มี") ระบุว่า `sale-reservation-flow` มีใน Champ จริง ห้ามตัด ; (4) ต้องมี `po_qty` ที่แถวใบสั่งขายและ `so_ref_no` ที่แถวใบสั่งซื้อ/ใบขาย — ยังไม่มีทั้งหมดใน BC |
| 591 ตรวจสอบวันที่ใบสั่งขาย/สั่งจองสินค้า | `sale-order-date-check` → `/sales/order-dates` กลุ่ม "transaction" (`menu-data.ts:251`) ; `erp-operations.ts:92-97` (`order_dates`, บรรทัด 96 "ตรวจสอบและปรับปรุงวันครบกำหนดและลำดับคิวการส่งมอบ") | รอพัฒนา (ต่าง) | (1) Champ = บันทึกวันที่ยืนยัน/วันที่ส่งจริงรายแถวของใบสั่งขาย โดยมีวันที่เดิม (read-only) ; กติกา เดิม ≤ ยืนยัน ≤ ส่งจริง ; แก้ actualdate ได้เมื่อแถวมีใบขายอ้างอิง ; ไม่ใช่ "ลำดับคิวการส่งมอบ" — แก้บรรทัด 96 ; (2) คอลัมน์ที่ควรเพิ่มใน `sale_order_lines`: `original_date`, `confirm_date`, `actual_date` (type date) ; (3) บันทึกการแก้ไขลง log (ก่อน→หลัง) เหมือน Champ ; (4) คำนึงว่า Champ ทำเป็น DLL เสริม (CHAMPMEGA) ไม่ใช่ตัวหลัก — ลำดับความสำคัญต่ำ |
| 5304 ปรับปรุงวันที่ส่งของให้ลูกค้า | `sale-delivery-date` → `/sales/delivery-dates` กลุ่ม "transaction" (`menu-data.ts:270`) ; `erp-operations.ts:100-105` (`delivery_dates`, บรรทัด 104 "จัดการตารางเวลารถขนส่งและนัดหมายวันส่งมอบ") | รอพัฒนา (ต่าง) | (1) Champ = เลือกลูกค้า + เลือกใบขายเชื่อที่ยังมียอดค้าง แล้วกำหนดวันที่ส่งของ + เครดิตนับจากวันที่ส่งของ → คำนวณวันครบกำหนดใหม่ → อัปเดต `DeliveryDate`/`DueDate` ของใบขาย — เป็นงานของระบบลูกหนี้ ไม่ใช่ตารางเวลารถขนส่ง — แก้บรรทัด 104 ; (2) ตัวกรองบิล: BillType ไม่ใช่เงินสด (≠0 และ ≠2), BillBalance≠0, IsCancel=0, IsCompleteSave=1 ; (3) Champ ไม่ตรวจงวดปิด/โพสต์ GL ก่อนแก้วันครบกำหนด และไม่เขียน log — ถ้าทำใน BC ควรเพิ่ม guard (งวดปิด, IsPostGL) และ audit log ตามหลักบัญชี (ส่วนที่เพิ่ม ไม่ใช่ลอก Champ) แจ้งลุงจืดตัดสินใจ ; (4) **ขัดกฎเจ้าของ**: ระบบลูกหนี้/ขายยังนอกขอบเขต GL ตอนนี้ |

ประเด็นข้ามรายการ (สรุป)
1. ทั้ง 6 เมนูไม่มี backend — ถ้าจะทำ ต้องเริ่มจากตาราง PostgreSQL (snake_case lowercase) + Go sql.Tx แบบ synchronous ตามกฎเจ้าของ ไม่ใช่ Kafka/Mongo/Redis (`AGENTS.md`)
2. คำอธิบาย 5 บรรทัดใน `frontend/src/lib/erp-operations.ts` (68, 77, 88, 96, 104) อธิบายพฤติกรรมไม่ตรง Champ ทั้งหมด — แก้ถ้อยคำก่อนทำหน้าจอจริง (กฎ "ยึด Champ เป็นต้นแบบ" และ "ห้ามแต่งฟังก์ชันใหม่")
3. ชุดเอกสารนี้ไม่ลง GL (ไม่มี journal) — ผลบัญชีเกิดเมื่อออกใบขาย/มัดจำ (BCARDeposit + BCOutputTax) ซึ่งเป็นงานของโมดูลอื่น
4. ข้อควรห้ามลอกจาก Champ: ค่าภาษี 1.07 hard-code (shadow 41), เตือนวงเงินที่คำนวณจากตัวแปร 0, RefType=1 ทุกแถว, ตัวกรองรหัสพนักงานขายผิดคอลัมน์ในหน้ายกเลิก, สตริงดิบ "CANNOT_SAVE_BILLSTATUS_REF"
5. ข้อความภาษาไทยทุกจุด: ต้องใช้ถ้อยคำเดียวกับ Champ ตามกฎ (ป้ายเมนูตรงแล้ว) — ป้ายคอลัมน์/ข้อความเตือนที่เป็นไทยอนุมานต้องขอให้ลุงจืดหรือ DB `myresource` ยืนยันก่อนเปลี่ยน `languages.tsv`

### ส่วน 3 — BC เทียบ Champ (ส่วนของ 5327/5329/5306)

ฝั่ง BC ทั้ง 3 เมนูอยู่กลุ่ม "เงินมัดจำและรับล่วงหน้า" (menu-data.ts:256) เรียงลำดับเดียวกับ Champ และใช้ชื่อไทยตรงกับ Champ; ทุก route อยู่ใน `ERP_MODULE_CONFIGS` → `isMenuBackendRetired()` คืน true (menu-screen-status.ts:46-48) = **"ยังไม่พร้อม/รอพัฒนา"** — backend ฝั่ง Mongo ถูกถอดแล้ว (2026-09-23) และงานลูกหนี้/ขายเป็นงานอนาคตตามขอบเขต **GL-only** ปัจจุบัน → ทุกข้อด้านล่างคือ "สิ่งที่ต้องทำเมื่อลุงจืดสั่งเปิดโมดูลขาย/ลูกหนี้" ไม่ใช่งานตอนนี้

| Champ | BC menu id / route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ |
|---|---|---|---|
| [5327] บันทึกใบรับเงินมัดจำ | `receive-deposit` → `/transaction/receivedeposit` (menu-data.ts:258; config erp-transaction.ts:545-555) | รอพัฒนา (ชื่อเมนู/ลำดับ **เหมือน**) | 1) จอ workbench กลางใช้ `ErpTransactionDoc` (erp-transaction.ts:19-42) ที่ไม่มี tax_no/tax_date/tax_type แยก 3 แบบ/bill_balance/sum_of_wtax/ชื่อ-ที่อยู่ออกบิล/ref_no ใบจอง/credit_day/due_date → ต้องเป็นจอเฉพาะ 30 ช่องตามตาราง 5327 2) money เป็น JS `number` (erp-transaction.ts:28) ขัด Rule 20 → decimal string + PostgreSQL `numeric(19,2)`, อัตราแลกเปลี่ยน `numeric(19,10)` 3) prefix `DEP-IN` (erp-transaction.ts:553) ≠ bill code Champ "DEP" (billfrmreceiveearnest.cpp:36) 4) `apiPath: "paid"` ใช้ร่วม 3 เมนู (erp-transaction.ts:548) → แยก endpoint/ตาราง เช่น `ar_deposits` + `ar_deposit_uses` (doc_no, deposit_no, amount, balance, depos_tax_type, net_amount) + `money_receipt_lines` (save_from, payment_type, …) 5) VAT ของมัดจำ: คิดเฉพาะเมื่อมีเลขที่ใบกำกับ, ยอดคงเหลือแยกนอก = ก่อนภาษี / รวมใน = รวมภาษี; ต้องคำนวณที่ backend (กฎ Backend-First) และข้อกฎหมายเรื่องจังหวะออกใบกำกับต้องตรวจตาม skill thai-accounting-tax + docs/kms/21 ก่อน (งานนี้ไม่ได้ตรวจ) 6) ใบขายต้องมี F8 ตัดมัดจำ (`CLIBDlgDeposit`) + ใบลดหนี้เงินมัดจำ — ยังไม่มีใน BC 7) GL ใช้สูตรลงบัญชีกลุ่ม 7 ที่ผู้ใช้กำหนด ห้าม if รหัสบัญชี |
| [5329] บันทึกใบรับเงินล่วงหน้า | `receive-advance` → `/transaction/paidadvance` (menu-data.ts:259; erp-transaction.ts:523-533) | รอพัฒนา (ชื่อเมนู **เหมือน**; route/คำอังกฤษ "Customer Advance Payment" สับสนกับฝั่งจ่าย) | 1) ช่องตาม 5329 (17 ช่อง) รวม วิธีการชำระเงิน (BCPaymentType), เลขที่ใบจอง, รับชำระ/ยอดคงเหลือ — ไม่มีภาษี 2) ยอดเงินล่วงหน้าคงเหลือรายลูกค้า (BCAR.EarnestAmount) ต้องคำนวณจากเอกสาร (สูตร BCARProcess.cpp:873-933) เป็น query/ฟิลด์ใน `ar_customers` หรือ view 3) การใช้ผ่านใบเสร็จ (BillType 8) + ตั้ง "ใช้แล้ว" อัตโนมัติ — ต้องทำพร้อมจอใบเสร็จ 4) ห้ามลอกบั๊กลบภาษีขาย source 5 (BILLFrmEarnestSpecail.cpp:297-386) 5) money decimal + prefix `RCV-ADV` (erp-transaction.ts:531) ต่างจาก Champ (Champ ไม่ตั้ง bill code :33) |
| [5306] บันทึกคืนเงินรับล่วงหน้า | `return-advance` → `/transaction/paidadvancerefund` (menu-data.ts:260; erp-transaction.ts:534-544) | รอพัฒนา (ชื่อเมนู **เหมือน**) | 1) ช่องตาม 5306 (17 ช่อง): เลือกใบรับเงินล่วงหน้าแล้วเติมลูกค้า/มิติ/ยอดอัตโนมัติ; คืนเกินคงเหลือต้อง **แจ้งเตือนชัด** (Champ รีเซ็ตเงียบ :1103-1124 — ขัดกฎ UX "Feedback ทุก action") 2) ตรวจ IsPostGL ตอนลบ (Champ ขาด :230-263) และตรวจใบรับเงินล่วงหน้าที่พิมพ์เองว่าไม่ยกเลิก/คงเหลือ > 0 (Champ ขาด :1076) 3) ห้ามลอกบั๊ก ChqOnHand BCAR/BCAP (:492 vs :588), document flow ตาราง AP (:1408), print check ผิดตาราง (:1507) 4) บันทึก+ปรับยอดใบต้นทางใน `sql.Tx` เดียว (Champ ใช้ transaction :485-624 — ให้คงไว้) 5) GL สูตรกลุ่ม 41 ฝั่งจ่าย (`ProcessConditionGroup44`) |
| (ร่วม) หน้าต่างรับ/จ่ายเงิน F7 | ไม่มีใน BC | ยังไม่มี | ต้องมีคอมโพเนนต์รับเงินกลาง (เงินสด/โอน+สมุดธนาคาร/เช็คใหม่-ตัดเช็คเดิม/บัตรเครดิต+ค่าธรรมเนียม/เงินสดย่อย/รายได้-ค่าใช้จ่ายอื่น/เงินเกิน) + สูตร `IsMoneyOk` ที่ backend; ผูกกับโมดูลเช็ค/ธนาคาร (งานอนาคต) |
| (ร่วม) ยกเลิก/อนุมัติ + ป้าย N/CF/0.00/CC | BC GL มีสถานะร่าง/ผ่าน/กลับ/ยกเลิก เฉพาะสมุดรายวัน | ต่าง | เอกสารขายต้องมี is_cancel/is_confirm + cancel_code/confirm_code + เวลา + เหตุผล, guard ตามตาราง Guard ของแต่ละจอ; ยกเลิกต้อง `ReverseAccount` (กลับรายการ) ไม่ลบเงียบ (ตรงกฎ "เอกสารผ่านบัญชีแล้วแก้ด้วยรายการกลับ") |

ความขัดแย้งกับกฎเจ้าของ: ไม่มีเรื่องเงินเดือน; ฟิลด์ SCG (SCGID/JobID/Get Promotion Msg) เป็นของลูกค้าเฉพาะราย — **ไม่เสนอให้ทำ** เว้นลุงจืดสั่ง; ทั้ง 3 เมนูอยู่นอกขอบเขต GL-only ตอนนี้ (สำนักงานบัญชีบันทึกเงินมัดจำ/รับล่วงหน้าผ่านสมุดรายวัน GL ได้อยู่แล้ว)

### ส่วน 4 — BC เทียบ Champ (ส่วนของ [5328])
| Champ | BC menu id / route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ (BC file:line) |
|---|---|---|---|
| [5328] บันทึกขายสินค้า,บริการ | `sale` → `/transaction/sale` ชื่อไทยตรงกับ Champ (frontend/src/lib/menu-data.ts:267, กลุ่ม "งานขายสินค้า" :265) | **รอพัฒนา** | ดูรายการด้านล่าง |

**สถานะ "รอพัฒนา"**
- route นี้เป็น ERP transaction ทำให้ `isMenuBackendRetired()` คืนค่า true (frontend/src/lib/menu-screen-status.ts:46-50) จอจึงแสดงการ์ด "รอพัฒนา" (frontend/src/app/menu/main-menu-screen.tsx:2817)
- backend PostgreSQL ไม่มี package ใบขาย (`backend/internal/` ไม่มีโมดูลขาย; `backend/pkg/microservice/storage_name.go:90` `saleinvoice` เป็นของค้างยุค Mongo)

**สิ่งที่ต้องปรับเมื่อสร้างจริง**
1. **ขอบเขตงาน — ชนกับกฎเจ้าของ (ต้องให้ลุงจืดตัดสินก่อน)**
   - กฎตอนนี้คือ GL-only: AR/ERP เป็นงานอนาคต
   - ทางที่ BC มีอยู่ตอนนี้คือ บันทึกยอดขายเป็นใบสำคัญ GL ที่มีรายละเอียดภาษีขาย `details.vats` (backend/internal/goapi/handlers/tax_report.go:89) คู่กับเอกสารลูกหนี้ใน subledger (backend/internal/generalledger/subledger_types.go:45-58: มีแค่ partner, เลขที่, วันที่, ครบกำหนด, ยอด)
   - ทางนี้ไม่มีรายการสินค้า สต๊อก การรับเงินหลายช่องทาง หรือการคุมวงเงิน
   - การทำให้ครบเท่า Champ = เปิดโมดูลขาย/ลูกหนี้เต็มรูป **ห้ามเริ่มเองจนกว่าลุงจืดสั่ง**
2. **ชนิดข้อมูล — ชน Rule 20**
   - `ErpDetailItem.qty/price/sumamount` และ `ErpTransactionDoc.totalamount` เป็น JS `number` (frontend/src/lib/erp-transaction.ts:11-15, :28)
   - ต้องเปลี่ยนเป็น decimal string และฝั่ง PostgreSQL ใช้ `numeric(18,2)` สำหรับเงิน, `numeric(18,4)` สำหรับจำนวน/ราคา ตามจำนวนตำแหน่ง QtyPoint/PricePoint/AmountPoint
   - ต้องมีนโยบายปัดเศษที่ชัดเจน เพราะ `RoundNumber` ของ Champ ยังไม่ verify
3. **ประเภทเอกสาร/ภาษีไม่ครบ**
   - BC มีแค่ `vattype?: number` และ `status 0/1/2` (erp-transaction.ts:34, :36) และใช้คำนำหน้า `INV` (:156) ส่วนชื่ออังกฤษ "Sale & Dispatch" (:153) ไม่ตรงกับ menu-data "Sale" (:267)
   - ต้องเพิ่ม `bill_type` (0–3 สด/เชื่อ × สินค้า/บริการ) และ `tax_type` (0 แยกนอก / 1 รวมใน / 2 อัตราศูนย์)
   - เลขที่ใบกำกับควรตั้งค่าเริ่ม = เลขที่เอกสาร (ยกเว้นขายบริการเชื่อ) และต้องมีกลุ่มเลขเอกสารตามประเภท
   - สถานะต้องแยกเป็น flag อิสระ: `is_cancel`, `is_confirm` (+ผู้/เวลา), `pay_bill_status`, `bill_balance`, `is_post_gl`, `cr_authorize_man` แทนการใช้ status ตัวเดียว
4. **ความถูกต้องของ transaction — ต้องดีกว่า Champ**
   - Champ commit หัว/บรรทัด (ChmFrmBilling.cpp:3800) **ก่อน** ตัดสต๊อก ภาษีขาย รับเงิน GL และลูกหนี้ (:3806-3861) ถ้าล้มกลางทางข้อมูลจะไม่ครบ
   - BC ต้องทำทั้งหมดใน `sql.Tx` เดียว (กฎ Synchronous ACID)
   - เอกสารที่ลงบัญชีแล้วให้ "ยกเลิก = กลับรายการ" (เหมือน `ReverseAccount`) และอนุญาต "ลบ" เฉพาะฉบับร่าง (ตามกฎ GL ของ BC)
   - ข้อกำหนดการเก็บใบกำกับภาษีที่ยกเลิก **ยังไม่ verify กับแหล่งทางการ** ต้องตรวจตาม skill thai-accounting-tax ก่อนออกแบบ
5. **วงเงินลูกหนี้**
   - BC มีแค่รายงาน `ar_credit_limit` (backend/internal/goapi/handlers/debt_report.go:403)
   - ยังไม่มีการตรวจตอนบันทึก, ตัวเลือก 3 ระดับ (`DebtOverWarning` 0/1/2 + `BillDebtWarning` + สถานะลูกค้า `IsMakeTrans`) และ dialog อนุมัติพร้อมเก็บ `cr_authorize_man`
6. **ฟิลด์ที่ต้องเรียงตาม Champ**
   - แท็บ 1 ตามลำดับตารางด้านบน: ประเภทการขาย/ประเภทภาษี → เลขที่ใบขายสินค้า/วันที่ออกเอกสาร → เลขที่ใบกำกับภาษี/วันที่ใบกำกับ → รหัสลูกค้า → ตาราง 11 คอลัมน์ → ยอด 6 บรรทัดท้ายบิล + พนักงานขาย/เงื่อนไขส่งมอบ/แผนก/โครงการ/น้ำหนัก
   - แท็บ 2 "รายการอื่นๆ" 33 ช่อง
   - ชื่อคอลัมน์ BC เป็น snake_case เช่น `doc_no, doc_date, tax_no, tax_date, bill_type, tax_type, ar_code, credit_day, due_date, sale_code, is_condition_send, depart_code, project_code, sum_of_item_amount, except_tax_amount, discount_word, discount_amount, before_tax_amount, tax_rate, tax_amount, total_amount, net_debt_amount, sum_of_wtax, sum_of_deposit1, bill_balance`
7. **คีย์ลัด**
   - ต้องใช้ชุด F2/F3/F5/F6/F7/F8/F11/F12, Shift+F2/F5/F9, Ctrl+F2/F3/F5/F7/F8/F9/F12, Ctrl+P/N/F, PageUp/PageDown ตามตาราง และปุ่ม "Key" แสดงรายการคีย์
8. **ตัดทิ้ง (เฉพาะลูกค้า Champ บางราย ไม่ใช่ parity)**: SCG ID/Coupon/Wallet/Promotion Msg, QARefno, integration ขนส่ง/marketplace ใน OnSaveData, รหัสเครื่อง (`nUseMachineCode`) — เสนอให้ลุงจืดยืนยันก่อนตัด
9. **ไม่ชนกฎ**: ไม่มีส่วนเงินเดือน, ใช้ PostgreSQL อย่างเดียว (ตาราง SQL Server ข้างบนใช้เป็นหลักฐานเท่านั้น)

### ส่วน 5 — BC เทียบ Champ (ส่วนของ 5330/5331)

**ภาพรวม BC ตอนนี้**
- ทั้ง 2 เมนูมีชื่อ Champ ครบใน `frontend/src/lib/menu-data.ts:268-269` (กลุ่ม "งานขายสินค้า")
- route ทั้งสองเป็น ERP transaction route จึงถูกนับว่า backend ปลดระวาง → แสดง "ยังไม่พร้อม/รอพัฒนา" (`menu-screen-status.ts:46-50` เรียก `isErpTransactionRoute`)
- config ปัจจุบันเป็นฟอร์มกลางแบบ generic (`erp-transaction.ts:159-180`)
  - apiPath ชี้ไป API ยุค MongoDB ที่ถูกถอดแล้ว (`sale-invoice-return`, `bank/saledebitnote` — ชื่อ collection เดิมใน `backend/pkg/microservice/storage_name.go:87-89`)
- ไม่มี API บน PostgreSQL สำหรับเอกสาร 2 ชนิดนี้

| เมนู Champ | BC id → route | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ (เมื่อลุงจืดสั่งทำโมดูลขาย/ลูกหนี้) |
|---|---|---|---|
| [5330] บันทึกใบรับคืนสินค้า,ลดหนี้ | `sale-return` → `/transaction/salereturn` (menu-data.ts:268; erp-transaction.ts:159-169) | **รอพัฒนา** (และนอกขอบเขต GL-only ตอนนี้) | ดูรายการปรับด้านล่าง |
| [5331] บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้) | `debit-note` → `/transaction/debitnote` (menu-data.ts:269; erp-transaction.ts:170-180) | **รอพัฒนา** (นอกขอบเขต GL-only) | ดูรายการปรับด้านล่าง |

**สิ่งที่ต้องปรับ — [5330] ใบรับคืนสินค้า/ลดหนี้**
1. **หัวเอกสาร** เรียงตาม Champ: เลขที่ใบรับคืน/ลดหนี้ → วันที่ออกเอกสาร → เลขที่ใบกำกับ → วันที่ใบกำกับ → ประเภทภาษี → รหัสแผนก → รหัสลูกค้า+ชื่อ → ประเภทรับคืน/ลดหนี้ (0/1/2) → 4 checkbox ("รับคืนเงินสด" ซ่อนและตั้งอัตโนมัติ)
   - ฟอร์ม generic ตอนนี้มีแค่ docno / docdatetime / custcode / vattype / vatrate / remark (erp-transaction.ts:19-42)
2. **ตารางใบขายที่อ้าง** 5 คอลัมน์ (เลขที่, ส่วนลดเดิม, ยอดเดิม, ยอดผลต่าง, ยอดที่ถูกต้อง) + เลือกได้ 3 แหล่ง (ใบขาย / เอกสารยกมา / ลูกหนี้อื่น) + มัดจำ — BC ยังไม่มีเลย
3. **ตารางสินค้า** 10 คอลัมน์ตาม Champ: เพิ่มเลขใบขายอ้างอิง คลัง ที่เก็บ บรรทัดอ้างอิง และเพดานจำนวนคืนได้ (`cnqty`)
   - `ErpDetailItem` ตอนนี้ไม่มี wh / shelf / ref (erp-transaction.ts:3-17)
4. **Guard ห้ามลดหนี้เกินใบเดิม** (เทียบ IsInvoiceOverCN / NUM_ITEM_NOTMORE / INCORRECT_CREDIT_PRICE)
   - ทำที่ backend Go ภายใน `sql.Tx` เดียว
   - ล็อกแถวใบขายเดิมเพื่อกันกดซ้ำหรือสองคนลดหนี้ใบเดียวกันพร้อมกัน
5. **สถานะเอกสาร** ตาม Champ: บันทึก / ยืนยัน (Shift+F5) / ยกเลิก (Ctrl+F5, ต้องกรอกเหตุผล) / ห้ามแก้เมื่อผ่าน GL, วางบิลแล้ว หรือถูกตัดชำระแล้ว
   - BC มี status 0/1/2 แบบ generic (erp-transaction.ts:36)
6. **ภาษีขาย**: เขียนแถว VAT แบบใบลดหนี้ (ลดภาษี) และเก็บวันที่ใบกำกับในแถว VAT
   - ใน BC GL ใช้ `document_type 3` ซึ่งบังคับเลขที่และวันที่ใบกำกับเดิม (subledger_vat.go:23, :116) — **เข้มกว่า Champ ให้คงไว้** (Champ เก็บแค่ใน BCInvCreditNote)
7. **ลูกหนี้**: แยก "ลดยอดใบกำกับเดิม" (ตัดกับใบขายทันที) กับ "เอกสารเครดิตรอตัด"
   - ใน BC GL เทียบได้กับ `gl_subledger_documents` + `gl_subledger_settlements` (subledger.sql:11-36)
   - อย่าลอกบั๊กยกเลิกแล้วไม่คืนยอดใบขาย (bcprocess.cpp:2016)
8. **เลขที่เอกสาร**: Champ ใช้รหัสบิล "CN" (billfrminvcreditnote.cpp:53) ส่วน BC ใช้ prefix "SR" (erp-transaction.ts:167) → ต้องเลือกตัวใดตัวหนึ่ง (ถามลุงจืด)
9. **คีย์ลัด**: F2 GL, F3 ภาษี, F5 / Shift+F5 / Ctrl+F5, F6 WHT, F7 เงิน, F9 ดึงรายการ, F11 บันทึก, F12 เส้นทางเอกสาร, PageUp / PageDown สลับตาราง

**สิ่งที่ต้องปรับ — [5331] ใบเพิ่มหนี้**
1. **หัวเอกสาร**: เลขที่ใบเพิ่มหนี้ → วันที่ออกเอกสาร → เลขที่ใบกำกับ (ค่าเริ่มต้น = เลขที่เอกสาร) → วันที่ใบกำกับ → ประเภทภาษี → รหัสแผนก → รหัสลูกค้า → **เครดิต(วัน) ↔ ลงวันที่ (คำนวณกลับไปมาได้)** → ประเภทเพิ่มหนี้ (0 เพิ่มหนี้ / 1 ราคาผิด)
2. **ตารางใบขาย 5 คอลัมน์** (ยอดที่ถูกต้อง = ยอดเดิม + ผลต่าง) และตารางสินค้า 10 คอลัมน์ที่ค้นจากสินค้าหลัก + F9 ดึงจากใบขาย
3. **ใบเพิ่มหนี้มียอดคงค้างของตัวเอง**: ถูกตัดในใบวางบิล/ใบเสร็จเป็นชนิดเอกสาร "ใบเพิ่มหนี้" และไม่แตะยอดคงค้างของใบขายเดิม
   - ใน GL ใช้ VAT `document_type 2` ซึ่งต้องมีใบกำกับเดิม (subledger_vat.go:116)
4. **สต๊อก**: ประเภท 0 ตัดสต๊อกออกตามคลัง/ที่เก็บ และตรวจห้ามติดลบ ณ วันที่เอกสาร (ConfirmSave :805-853); ประเภท 1 ไม่กระทบสต๊อก
5. **เลขที่เอกสาร**: Champ "DBN" (billfrmardeposit.cpp:42) ส่วน BC "DN" (erp-transaction.ts:178) → ถามลุงจืด
6. **อย่าลอกบั๊ก**: ยกเลิกต้องกลับรายการลูกหนี้จริง (เทียบ billfrmardeposit.cpp:505 กับ bcprocess.cpp:2357-2365) และควรเพิ่ม guard เพดานยอด/จำนวน — เป็นข้อเสนอ ต้องถามลุงจืดก่อน เพราะ Champ ไม่มี

**ข้อขัดกฎเจ้าของ / ข้อควรระวัง**
- **ขอบเขต GL-only ตอนนี้**: โมดูลขาย/ลูกหนี้/สต๊อกเป็นงานอนาคต — ห้ามเปิดจอ 2 จอนี้เองจนกว่าลุงจืดสั่ง
  - ระหว่างนี้งานลดหนี้/เพิ่มหนี้ของสำนักงานบัญชีบันทึกผ่าน **สมุดรายวัน GL + รายละเอียด VAT ประเภท 2/3** (subledger_vat.go:23,116; tax_vat_summary.go:213-219) และ subledger ลูกหนี้ (subledger.sql:11-36)
- **ตัวเลขเงิน**: `ErpTransactionDoc` / `ErpDetailItem` ใช้ JS `number` กับยอดเงินทุกช่อง (erp-transaction.ts:12-15, :28-33) ซึ่ง **ขัด Rule 20**
  - ตอนสร้างจริงต้องส่งเป็น decimal string และเก็บใน PostgreSQL เป็น `numeric(p,s)`
  - ชื่อคอลัมน์ที่เสนอ (snake_case): `doc_no, doc_date, tax_no, tax_date, tax_type, cause_type, cause_code, ar_code, credit_day, due_date, return_status, return_money, debt_limit_state, is_cn_deposit, sum_old_amount, sum_true_amount, sum_before_tax, sum_tax_amount, net_debt_amount, bill_balance`; บรรทัด: `invoice_no, ref_line_number, wh_code, shelf_code, unit_code, qty, disc_qty, price, amount, net_amount, stock_process`
- **PostgreSQL อย่างเดียว**: ไม่มีคิวแบบ async แบบ proc job ของ Champ — ลูกหนี้/สต๊อก/ภาษี/GL ต้องทำ synchronous ใน `sql.Tx` เดียว (กฎ 100% PostgreSQL)
- ไม่มีประเด็น payroll; ไม่มีการอ้างชื่อผลิตภัณฑ์คู่แข่ง

### ส่วน 6 — BC เทียบ Champ (ส่วนของรายงานขายประจำวัน)

#### ข้อเท็จจริงฝั่ง BC ที่ใช้เทียบ (ตรวจแล้ว)

| # | ข้อเท็จจริง | หลักฐาน |
|---|---|---|
| ก | เมนู "รายงานขาย" (`bill-reports`) ของ BC มีแค่ 7 รายการ : `sales`, `sales-daily`, `sales-by-seller`, `sales-by-document`, `gross-profit-by-document`, `gross-profit-by-product`, `sale-return-report` — ไม่มีรายการไหนตรงชื่อกับรายงานของ Champ ทั้ง 24 รายการ ยกเว้น `sales-daily` (ชื่อ "รายงานรายวันขายสินค้าและบริการ" = ชื่อของ 5337) และ `sales-by-seller` (ชื่อ = 5302) | `frontend/src/lib/menu-data.ts:275-284` |
| ข | จอรายงานเป็นจอกลางตัวเดียว: ช่วงวันที่เลือกได้แค่ปุ่มสำเร็จรูป วันนี้ / 7 วัน / เดือนนี้ / ปีนี้ (ไม่มีกล่องเงื่อนไขจาก–ถึง ไม่มีช่องเลขที่เอกสาร ลูกค้า สินค้า พนักงานขาย ไม่มี checkbox/combo ตามรายงาน) ; มีช่อง "ค้นหาในรายงาน" กรองข้อความในผลที่โหลดมาแล้ว (ฝั่ง browser) เรียงตามคอลัมน์ ส่งออก CSV และพิมพ์ | `frontend/src/app/report/erp-report-viewer.tsx:51-63,94,112-113,136-139,180-185,239,254,283-290` |
| ค | คำนิยามรายงาน (คอลัมน์/ป้าย) อยู่ใน `erp-reports.ts` ; รายงานที่มี API จริงมีแค่ `sales_by_document`, `gross_profit_document`, stock 2 รายการ, `ap_*` 4 รายการ, `ar_*` 4 รายการ ; ที่เหลือแสดงเป็น "ยังไม่มีข้อมูล" (`isMenuDataPending`) — รวม `sales_daily`, `sales_by_seller`, `sales_return_report`, `sales_summary` | `frontend/src/lib/erp-reports.ts:147-162,163-177,226-243,530-547` ; `frontend/src/lib/menu-screen-status.ts:79-83` |
| ง | ฝั่ง backend รายงานขายมี 2 endpoint คือ `/api/report/sales/by-document` และ `/api/report/sales/summary` ; ตัวกรองที่รับ: holding, ช่วงวันที่, สาขา, รหัสสินค้า, เรียง, limit/offset (ไม่มีลูกค้า/พนักงานขาย/ประเภทการขาย) ; endpoint `/api/report/sales/summary` ไม่ถูกเรียกจาก `erp-reports.ts` (ไม่พบสตริง `sales/summary` ในไฟล์) และ `fetchErpReportData` คืน `report_not_available` สำหรับรหัสรายงานที่ไม่อยู่ในชุด API-ready (`erp-reports.ts:835-837`) ; อ่านจาก `stock_ledger` + ตาราง `doc` ที่ `transflag = 44` ไม่ได้อ่านตารางหัว-รายการใบขายแยกต่างหาก | `backend/internal/goapi/bootstrap.go:354-355` ; `backend/internal/goapi/handlers/sales_report.go:17-26,187-196,265-282` |
| จ | ~~ยอดรวมของรายงานขายฝั่ง BC ใช้ `float64`~~ **แก้แล้ว 2026-10-11** (`b394e44d`): รวมยอดและ % GP ด้วย `decimal.Decimal` (`salesMarginPercent` ปัด 2 ตำแหน่ง) ส่งเป็นข้อความทศนิยม ; Champ ใช้ `double` | `backend/internal/goapi/handlers/sales_report.go:391,496` |
| ฉ | เอกสารกลุ่มขายที่เป็นต้นทางของรายงาน (ใบเสนอราคา, ใบสั่งขาย, ใบรับเงินมัดจำ, ใบรับเงินล่วงหน้า, ใบเพิ่มหนี้ ฯลฯ) มีเมนูบันทึก (`menu-data.ts:244-260,269`) และคำนิยามจอ (`erp-transaction.ts:127-129,138-139`) แต่ฝั่ง Go ไม่พบ handler ของ `/api/erp-transaction` (ค้น `grep -rn "erp-transaction" backend --include=*.go` ไม่พบ ; ในฝั่ง frontend มีแค่ route proxy `frontend/src/app/api/erp-transaction/[...erpPath]/route.ts`) — จอเหล่านี้ตอบ `module_not_available` (404) | `frontend/src/lib/erp-transaction.ts:874,888` |
| ช | ขอบเขตตอนนี้ของเจ้าของ: เน้น GL เป็นหลัก ระบบ ERP ขาย/ซื้อยังไม่ต้องทำ ; PostgreSQL เท่านั้น ; ยอดเงินเป็น decimal ; ไม่ทำเงินเดือน ; คอลัมน์ BC เป็น snake_case ตัวพิมพ์เล็ก — ข้อมูลตามไฟล์กฎโปรเจกต์ (`D:\bccode\AGENTS.md`) | `AGENTS.md` (หัวข้อ GL-only / PostgreSQL / ยึด Champ เป็นต้นแบบ) |

#### ตารางเทียบรายรายงาน (24 รายการ)

คำย่อในช่อง "สิ่งที่ต้องปรับ": **[GL]** = นอกขอบเขต GL-only ตอนนี้ ไม่ควรเริ่มก่อนลุงจืดสั่ง ; **[ต้นทาง]** = ต้องมีตารางต้นทาง + endpoint ก่อน (ข้อ ฉ) ; **[กรอง]** = ต้องมีกล่องเงื่อนไขจาก–ถึงตามหัวข้อที่อ้าง (ข้อ ข — จอกลางปัจจุบันทำไม่ได้) ; **[เงิน]** = ยอดเงินเป็น `NUMERIC` / สตริงทศนิยม ไม่ใช้ float, ปัดตามจุดทศนิยมที่ระบุ.

| menu id | รายการ (Champ) | BC menu id / route | สถานะ | สิ่งที่ต้องปรับ |
|---|---|---|---|---|
| 5344 | รายงานใบเสนอราคา | ไม่มี (เอกสารต้นทาง `quotation` `/transaction/quotation`, `menu-data.ts:244`) | ยังไม่มี | [GL][ต้นทาง][กรอง 1.1] ; ตัวกรอง: วันที่, เลขที่เอกสาร, ลูกค้า, สินค้า, คำตอบจากลูกค้า, ติ๊กนับจำนวนตามวัน ; เสนอ (รอลุงจืดอนุมัติ) ยุบ 5344/5333/5354 เป็นจอเดียว มีตัวเลือก "แสดงรายการย่อย" และ "ปกติ/ยกเลิก" ตามกฎยุบรวมของ AGENTS.md |
| 5333 | รายงานการบันทึกใบเสนอราคาพร้อมรายการย่อย | ไม่มี | ยังไม่มี | รวมกับ 5344 (ตัวเลือกแสดงรายการย่อย) ; กรองสินค้าต้องกำหนดชัดว่ากรองระดับเอกสารหรือระดับบรรทัด — Champ กรองระดับเอกสารแล้วแสดงทุกบรรทัด (1.6 ข้อ 3) ซึ่งทำให้แถวที่ไม่ตรงสินค้าโผล่มา ; BC ควรกรองระดับบรรทัด หรือทำป้ายบอก |
| 5345 | รายงานสถานะใบเสนอราคา | ไม่มี | ยังไม่มี | [GL][ต้นทาง][กรอง 1.1] ; ไม่มีตัวกรองลูกค้า ; เก็บเป็น "ตัวกรองสถานะ" ของจอรวม ; อย่าลอกบั๊กยอดรวมย่อยของเอกสารสุดท้ายที่ใช้ยอดรวมทั้งรายงาน (1.6 ข้อ 1) |
| 5354 | รายงานการยกเลิกใบเสนอราคาสินค้าพร้อมรายการย่อย | ไม่มี | ยังไม่มี | รวมกับ 5344 (ตัวกรอง "ยกเลิก") ; ถอดช่อง "คำตอบจากลูกค้า" ที่อยู่บนจอแต่ไม่มีผล (1.6 ข้อ 2) |
| 5363 | รายงานการใบสั่งขาย/จองสินค้าพร้อมรายการย่อย | ไม่มี (ต้นทาง `sale-order` `/transaction/saleorder`, `menu-data.ts:247`) | ยังไม่มี | [GL][ต้นทาง][กรอง 2.1][เงิน 2.4] ; สกุลเงิน: ถ้าเปิดหลายสกุลและเลือกแปลงเป็นบาท คูณอัตราแลกเปลี่ยนแล้วปัดตามจุดทศนิยมยอดเงิน (ตาราง 2.4) |
| 5364 | รายงานการยกเลิกใบสั่งขาย/สั่งจองสินค้าพร้อมรายการย่อย | ไม่มี | ยังไม่มี | รวมกับ 5363 เป็นจอเดียว (ตัวกรอง "ยกเลิก") เสนอรอลุงจืดอนุมัติ |
| 5346 | รายงานการบันทึกใบรับเงินมัดจำ | ไม่มี (ต้นทาง `receive-deposit` `/transaction/receivedeposit`, `menu-data.ts:258`) | ยังไม่มี | [GL][ต้นทาง][กรอง 3.1][เงิน 3.4] ; ตัวเลือก "เลือกเอกสาร" ยอดคงค้าง (ค่าเริ่มต้น) / ทั้งหมด ; ยอดคงค้างคือ `BillBalance > 0` ของเอกสาร |
| 5334 | รายงานการตัดใบรับเงินมัดจำ | ไม่มี | ยังไม่มี | [ต้นทาง] ต้องมีตารางการนำมัดจำไปหักใบขาย (เอกสารขาย, จำนวนที่หัก เป็น decimal) ; Champ ไม่พิมพ์ใบมัดจำที่ยังไม่มีรายการตัด (3.6) — ถามลุงจืดว่าตั้งใจให้ซ่อนหรือไม่ ; แก้ฐานยอดหัว/แถวตัดให้ใช้ฐานเดียว (3.6 ข้อ 9) |
| 5335 | รายงานใบรับเงินล่วงหน้า | ไม่มี (ต้นทาง `receive-advance` `/transaction/paidadvance`, `menu-data.ts:259`) | ยังไม่มี | [GL][ต้นทาง][กรอง 3.1] ; กติกายอดรวมไม่นับใบยกเลิก (3.4) — BC ต้องกำหนดกติกาเดียวกับ 5341 (ซึ่ง Champ นับใบยกเลิก) |
| 5355 | รายงานการตัดใบรับเงินล่วงหน้า | ไม่มี | ยังไม่มี | [ต้นทาง] ผูกกับใบรับชำระ (Champ อ่านรายการรับชำระ `BillType = 8`) — BC ใช้ตาราง link ของตนเอง ไม่ลอกรหัสชนิด ; ข้อสังเกต 3.6 (ไม่พิมพ์ใบที่ยังไม่มีรายการตัด) |
| 5341 | รายงานบันทึกคืนเงินรับล่วงหน้า | ไม่มี (ต้นทาง `return-advance` `/transaction/paidadvancerefund`, `menu-data.ts:260`) | ยังไม่มี | [GL][ต้นทาง][เงิน] ; กติกายอดรวมนับใบยกเลิก (3.4) ต่างจาก 5335 — ตั้งกติกาเดียว ; Champ เรียกตารางตัวเลือกผู้ใช้ผิดตาราง (3.6) อย่าลอก |
| 5337 | รายงานรายวันขายสินค้าและบริการ | `sales-daily` `/report/reportdedebisalesdaily` (`menu-data.ts:279`; `erp-reports.ts:147-162`) | ชื่อซ้ำ-เนื้อหาต่างกัน | Champ = 1 ใบขาย 1 แถว 18 คอลัมน์ (4.3) ; BC = สรุปรายวัน 6 คอลัมน์ (`salesdate, doccount, cashsales, transfersales, creditsales, totalamount`) ที่ Champ ไม่มี รวมทั้งคอลัมน์ "เงินโอน" — ขัดกฎ Strict Parity (ห้ามเพิ่มของที่ Champ ไม่มี) ต้องเสนอลุงจืดก่อน ; ถ้าจะตาม Champ ต้องเปลี่ยนเป็นรายเอกสาร + [กรอง 4.1] + ตัวเลือกประเภทการขาย (ค่าเริ่มต้น "ทั้งหมด") + กติกา POS/ใบยกเลิก (4.4) ; ปัจจุบันยังไม่มี API (ข้อ ค) |
| 5336 | รายงานบันทึกขายสินค้า,บริการพร้อมรายการย่อย | ไม่มี (ใกล้สุด `sales-by-document` `/report/salesreportbydocument` `menu-data.ts:281` ที่มี API แต่ระดับเอกสาร/รายการต่างโครง) | ใกล้เคียงบางส่วน | `sales_by_document` มี 2 โหมด header/detail (`sales_report.go:17-26`) ; ไม่พบคำว่า serial ใน `sales_report.go` (ไม่มีระดับ Serial) และอ่านจาก `stock_ledger` ไม่ใช่ใบขาย — ต้องตัดสินว่าจะเติมระดับ Serial/ฟิลเตอร์ตามหัวข้อ 4.1 หรือไม่ ; ระวังไม่ให้ซ้ำกับจอเดิมตามกฎยุบรวม |
| 5347 | รายงานการตัดใบขายสินค้า,บริการ | ไม่มี | ยังไม่มี | [GL][ต้นทาง] เป็นรายงาน "การตัด" ใบขาย — ท้ายแต่ละใบพิมพ์ยอดตัดจ่ายและยอดคงเหลือ (4.4) ; ในโครง BC ที่ใช้ GL เป็นศูนย์กลาง ควรอ่านยอดตัดจากตารางเชื่อมการจัดสรรที่เก็บไว้แล้ว (many-to-many) ไม่คำนวณซ้ำ |
| 5349 | รายงานบันทึกขายสินค้า,บริการตามประเภทการขาย | ไม่มี | ยังไม่มี | [GL][กรอง 4.1] ; เป็นรายงานเดียวกับ 5337 ที่ต่างกันที่การจัดกลุ่มตามประเภทการขาย — เสนอรวมเป็นตัวเลือก "จัดกลุ่มตามประเภทการขาย" ของ 5337 |
| 5340 | รายงานบันทึกใบรับคืนสินค้า,ลดหนี้ | `sale-return-report` `/report/reportdedebisalereturn` (`menu-data.ts:284`; `erp-reports.ts:226-243`) | ใกล้เคียงบางส่วน | BC มี 6 คอลัมน์ (`docdate, docno, refdocno, custname, reason, totalamount`) ; Champ 4 ระดับ (ใบลดหนี้ > ใบกำกับเดิม > สินค้า > Serial, 5.3) + แยกสาเหตุ (`CauseType` 0/1/2, R0) ; ต้องเพิ่มรายละเอียดตามหัวข้อ 5.3 และ [กรอง 5.1] ; ปัจจุบันยังไม่มี API (ข้อ ค) |
| 5338 | รายงานใบรับคืนสินค้า/ลดหนี้พร้อมรายการย่อย | `sale-return-report` (เดียวกับ 5340) | ใกล้เคียงบางส่วน | รวม 5340/5338 เป็นจอเดียว (ตัวเลือกแสดงรายการย่อย) เสนอรอลุงจืดอนุมัติ ; ผูกปลายทางคลิกให้ครบทุกระดับ (Champ ตั้งแค่ช่อง 0 ทำให้คลิกระดับล่างไม่มีปลายทาง — 5.6 ข้อ 2) |
| 5339 | รายงานการตัดใบรับคืนสินค้า,ลดหนี้ | ไม่มี | ยังไม่มี | [GL][ต้นทาง] ความหมาย `ReturnStatus`/`ReturnMoney` ยังไม่ verify (หัวข้อ 7 รายการที่ 7) — ห้ามเขียนสเปกฝั่ง BC จนกว่าจะยืนยัน |
| 5350 | รายงานสถานะใบรับคืนสินค้า/ลดหนี้ | ไม่มี | ยังไม่มี | [GL][ต้นทาง] 1 ใบ = 1 แถว ; Champ มี 2 กิ่ง (ทั่วไป/วัสดุก่อสร้าง `m_bChmMode`) และ `m_bChmMode = TRUE` เสมอ (`BC5Account.cpp:256`) — BC ใช้กิ่งเดียว ไม่เก็บ flag |
| 5342 | รายงานใบเพิ่มหนี้/เพิ่มสินค้า(ลูกหนี้)พร้อมรายการย่อย | ไม่มี (ต้นทาง `debit-note` `/transaction/debitnote`, `menu-data.ts:269`) | ยังไม่มี | [GL][ต้นทาง][กรอง 5.1][เงิน 5.4] ; ใบเพิ่มหนี้แยกจากใบลดหนี้ (ปลายทางคลิก 5331) — ถ้ายุบรวมกับ 5340/5338 ต้องมีคอลัมน์ "ประเภทเอกสาร" |
| 5343 | รายงานยอดขายสุทธิของพนักงานขาย | ไม่มี (ใกล้สุด `sales-by-seller`, ดูแถว 5302) | ยังไม่มี | [GL][ต้นทาง][กรอง 6.1] ; คอลัมน์ตาม 6.3 : มูลค่าขาย, ลดหนี้, เพิ่มหนี้, คงเหลือ (และฝั่งสุทธิของแต่ละตัว) ; โหมด "แสดงตามกลุ่มพนักงาน" ของ Champ มีโค้ดทำงานซ้ำ (หัวข้อ 7 รายการที่ 9) — BC ทำเป็นการจัดกลุ่มแบบเดียว |
| 5302 | รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย | `sales-by-seller` `/report/salesbyseller` (`menu-data.ts:280`; `erp-reports.ts:163-177`) | ชื่อซ้ำ-เนื้อหาต่างกัน | BC แสดง 5 คอลัมน์ (`sellername, target, actual, achievedpercent, commission`) — "เป้ายอดขาย / % บรรลุเป้า / คอมมิชชั่น" ไม่มีใน Champ (6.3) ; ขัดกฎ Strict Parity และค่าคอมมิชชั่นใกล้เคียงเรื่องเงินเดือนซึ่งอยู่นอกขอบเขต — ต้องขอลุงจืดตัดสิน ; ถ้าตาม Champ: [กรอง 6.1] เพิ่มตัวกรองสินค้า/กลุ่มพนักงาน และคอลัมน์ตามหัวข้อ 6.3 ; ปัจจุบันยังไม่มี API (ข้อ ค) |
| 5686 | รายงานการเก็บเงินของพนักงานขาย | ไม่มี | ยังไม่มี | [GL][ต้นทาง] ไม่มีปลายทางคลิกใน Champ (ส่วนหัวของหัวข้อ 6) ; มีการแยกชนิดการชำระ 4 แบบ (CA/CR/CH/CO) — BC ใช้รหัสชนิดการรับชำระของตัวเอง ไม่ฝังตัวย่ออังกฤษในโค้ด ; ป้ายช่องเงื่อนไขของ Champ ยังไม่ verify (ข้อ 2) |
| 5362 | รายงานการนำส่งเงินประจำวัน | ไม่มี (BC `sales-daily` มีคอลัมน์เงินสด/เงินโอน/ขายเชื่อ แต่เป็นสรุปรายวันของการขาย ไม่ใช่การนำส่งเงิน) | ยังไม่มี | [GL][ต้นทาง][กรอง 6.1] ; กรองด้วยช่วง "จำนวนเงิน" ที่ Champ ปัดทศนิยมตาม `QtyPoint` (ทศนิยมของจำนวน ไม่ใช่ยอดเงิน — 6.1 แถว 7) อย่าลอก BC ใช้จุดทศนิยมยอดเงิน ; เทียบงานฝั่งการเงิน/ธนาคารที่ BC มีอยู่ก่อนตัดสินใจทำซ้ำ |

#### ข้อแตกต่างสำคัญที่สุด (สรุป)

1. **ครอบคลุมน้อยมาก** — จาก 24 รายงาน BC มีรายการที่ "ชื่อซ้ำ" แค่ 2 (`sales-daily` = 5337, `sales-by-seller` = 5302) และ "ใกล้เคียงบางส่วน" 3 (`sale-return-report` ↔ 5340/5338, `sales-by-document` ↔ 5336) ที่เหลืออีก 19 ไม่มีเลย (`menu-data.ts:275-284`) ; 2 รายการที่ชื่อซ้ำมีเนื้อหาไม่ตรงกับ Champ และยังไม่มี API (`erp-reports.ts:147-177`, ข้อ ค).
2. **จอกลางไม่มีกล่องเงื่อนไข** — Champ ให้ผู้ใช้เลือกช่วงเลขที่เอกสาร/ลูกค้า/สินค้า/พนักงาน/ประเภท + combo + checkbox ต่อรายงาน ; BC มีแค่ปุ่มช่วงวันที่สำเร็จรูป (`erp-report-viewer.tsx:51-63,254`) — ต้องออกแบบ "เงื่อนไขรายงาน" ที่กำหนดต่อรายงานเป็นข้อมูล (ไม่ใช่ฝังต่อจอ) ก่อนรองรับกลุ่มนี้.
3. **ไม่มีข้อมูลต้นทาง** — ใบเสนอราคา ใบสั่งขาย เงินมัดจำ/ล่วงหน้า ใบลดหนี้/เพิ่มหนี้ ไม่มี handler ฝั่ง Go (ข้อ ฉ) และรายงานขายปัจจุบันอ่านจาก `stock_ledger` + `doc` (`sales_report.go:187-196`) ไม่ใช่ตารางใบขาย ; ทั้งหมดนี้อยู่นอกขอบเขต GL-only ตอนนี้.
4. **ตัวเลข float ทั้งสองฝั่ง** — Champ ใช้ `double` + `%f` ใน SQL ; BC เองใช้ `float64` ที่ `sales_report.go:382,446` — ต้องย้ายเป็น decimal/`NUMERIC` + สตริงทศนิยมใน JSON ก่อนต่อยอดรายงานกลุ่มนี้ (Rule 20).
5. **BC มีฟีเจอร์ที่ Champ ไม่มี** — คอลัมน์ "เงินโอน" ใน `sales_daily` และ "เป้า/% บรรลุเป้า/คอมมิชชั่น" ใน `sales_by_seller` (`erp-reports.ts:158,169-175`) ขัดกฎ Strict Parity และเรื่องคอมมิชชั่นใกล้กับเงินเดือน (นอกขอบเขต) — ต้องเสนอลุงจืดตัดสินว่าจะคงหรือถอด.

### ส่วน 7 — BC เทียบ Champ (ส่วนของรายงานขายเชิงวิเคราะห์)

#### สถานะ BC ที่ตรวจแล้ว (อ่านโค้ดวันที่จัดทำเอกสาร ; ไม่ได้รันระบบ)

- กลุ่มเมนู "รายงานขาย" (`bill-reports`) มี 7 รายการ: `sales` `/report/reportdedebisales`, `sales-daily` `/report/reportdedebisalesdaily`, `sales-by-seller` `/report/salesbyseller`, `sales-by-document` `/report/salesreportbydocument`, `gross-profit-by-document` `/report/reportgrossprofitbydocument`, `gross-profit-by-product` `/report/reportgrossprofitbyproduct`, `sale-return-report` `/report/reportdedebisalereturn` (`frontend/src/lib/menu-data.ts:274-285`) — **ไม่มีรายการใดตรงกับรายงานเชิงวิเคราะห์ 21 รายการของ Champ**.
- เครื่องยนต์รายงาน BC เป็นค่าตั้งแบบตายตัว `ERP_REPORT_CONFIGS` (route / code / category / title / columns / defaultSortKey) — **ไม่มีนิยามช่องเงื่อนไข (จาก–ถึง), ตัวเลือก "เรียงตาม" หลายโหมด, หัวกลุ่ม** (`frontend/src/lib/erp-reports.ts:11-26`). หน้าแสดงผลมีเฉพาะปุ่มช่วงวันที่ วันนี้/สัปดาห์/เดือน/ปี (`frontend/src/app/report/erp-report-viewer.tsx:51-63,254-262`), ค้นหา/เรียงฝั่งเบราว์เซอร์ (`:91-150`) และรวมยอดท้ายตารางด้วย `reduce` ฝั่งเบราว์เซอร์ (`:154-163`, แถวรวม `:394-418` ฝังคำว่า "รวม").
- รายงานที่ API พร้อมใช้ในกลุ่มขายมีแค่ `sales_by_document`, `gross_profit_document` (`erp-reports.ts:530-543`) ; รายงานอื่นของกลุ่มคืน `report_not_available` (`erp-reports.ts:835`) และเมนูแสดงสถานะ "ยังไม่พร้อม" (`frontend/src/lib/menu-screen-status.ts:77-84`). เส้นทางธุรกรรมขาย (ใบเสนอราคา ใบสั่งขาย ฯลฯ) ถูกจัดเป็น backend ที่ถอดแล้ว (`menu-screen-status.ts:35-45`) — ข้อมูลต้นทางของรายงานทั้ง 21 ยังไม่มีใน BC.
- ฝั่ง backend: `SalesReportByDocumentHandler`, `SalesReportSummaryHandler` (`backend/internal/goapi/handlers/sales_report.go:30,330` ; ลงทะเบียน `backend/internal/goapi/bootstrap.go:354-355`) อ่าน `public.stock_ledger`/`public.doc` ; รวมยอดและคำนวณ `marginpercent` ที่ backend ด้วย `decimal.Decimal` (`sales_report.go:391,467-500`) ; หน้าเว็บแสดงผลอย่างเดียว (`erp-reports.ts:918`) — แก้ 2026-10-11 `b394e44d` (เดิมใช้ `float64` และ JS Number).

#### ตารางเทียบรายรายการ

| Champ id / ชื่อ | BC menu id + route ที่ใกล้ที่สุด | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ |
|---|---|---|---|
| 5180 รายงานวิเคราะห์ใบเสนอราคาสินค้า | ไม่มี (ต้นทาง `quotation` `/transaction/quotation` `menu-data.ts:244`) | ยังไม่มี | รอระบบขาย ; ถ้าทำ: 11 โหมดเรียงตาม + เงื่อนไข 14 ช่อง (กลุ่ม A) ; ระวังบั๊ก `TypeCode BETWEEN start AND start` อย่าคัดลอก |
| 5185 วิเคราะห์ใบสั่งขาย/จองฯ | ไม่มี (`sale-order` `menu-data.ts:130` เป็นคำค้น) | ยังไม่มี | เหมือน 5180 + ช่อง "เลือกเอกสาร" (ใบสั่งขาย / ใบสั่งจอง ตาม `SOStatus` 0/1) |
| 5204 วิเคราะห์การขายแบบสรุป | `sales` `/report/reportdedebisales` (`menu-data.ts:278`, `erp-reports.ts:131-146`) | ต่าง | BC เป็นรายการต่อเอกสาร (วันที่ เลขที่ ลูกค้า ก่อนภาษี ภาษี รวม) ; Champ เป็นต่อสินค้า×หน่วยนับ ภายใต้หัวกลุ่ม 12 โหมด (จำนวน, ส่วนลด, จำนวนเงิน, สุทธิ) ; ต้องมี combo "เรียงตาม" ที่ backend รวมให้ |
| 5277 วิเคราะห์การขายสุทธิ | ไม่มี (ใกล้สุด `sale-return-report` `menu-data.ts:284`) | ยังไม่มี | ต้อง UNION ขาย/ลดหนี้/เพิ่มหนี้ (สุทธิ = ขาย − ลดหนี้ + เพิ่มหนี้), 14 โหมด, ช่อง "Display summary" และ "เฉพาะสินค้ามีโปรโมชั่น", จำกัดสิทธิ์ตามพนักงานขาย |
| 5310 รายงานสรุปยอดขายสินค้า | ไม่มี (`sales` ใกล้สุด) | ยังไม่มี | แยก 3 ช่องตามประเภทภาษี (คิดภาษี / ยกเว้น / อัตราศูนย์) + ส่วนลด + ยอดขายรวม ; `sales_summary` ฝังคอลัมน์ "ภาษีขาย 7%" ตายตัว (`erp-reports.ts:143`) — ต้องไม่ฝังอัตราภาษีในชื่อคอลัมน์ |
| 5232 วิเคราะห์ใบรับคืน/ลดหนี้ แบบแจกแจง | `sale-return-report` (`menu-data.ts:284`, `erp-reports.ts:226-241`) | ต่าง | BC ต่อเอกสาร (เลขที่ใบลดหนี้ อ้างอิงบิลเดิม ลูกค้า สาเหตุ ยอด) ; Champ ต่อรายการสินค้า (จำนวนรับคืน ส่วนลด จำนวนเงิน สุทธิ) ตามวันที่/แผนก/กลุ่ม/โครงการ ; ต้องมีกฎ "ราคาผิด ไม่นับจำนวน" ; ห้ามทำ `UPDATE` ในรายงาน |
| 5236 วิเคราะห์ใบรับคืน/ลดหนี้ แบบสรุป | `sale-return-report` | ต่าง | เหมือน 5232 แต่รวมต่อวัน×สินค้า ; แก้บั๊กหัวคอลัมน์สลับ (ส่วนลด/จำนวนเงิน) และ HAVING |
| 5240 วิเคราะห์ใบเพิ่มสินค้า(ลูกหนี้) แจกแจง | ไม่มี | ยังไม่มี | ตารางต้นทาง = ใบเพิ่มหนี้ลูกหนี้ ; ฟิลเตอร์ตรง 5232 ; แก้บั๊กจำนวนจาก–ถึง |
| 5244 วิเคราะห์ใบเพิ่มสินค้า(ลูกหนี้) สรุป | ไม่มี | ยังไม่มี | เหมือน 5240 รวมต่อวัน×สินค้า |
| 5248 รายงานจัดอันดับยอด | `gross-profit-by-product` (`menu-data.ts:283`, `erp-reports.ts:209-225`) ใกล้สุด | ต่าง | BC เรียงตามกำไรขั้นต้นตายตัว (`defaultSortKey: grossprofit`) ; Champ เลือกเรียงจำนวน/มูลค่า × น้อย->มาก/มาก->น้อย และเป็นขาย − ลดหนี้ + เพิ่มหนี้ ; % ของ Champ เทียบต้นทุน (ไม่ใช่ยอดขาย) ; BC ต้องแยก "% เทียบต้นทุน" กับ "% GP" ให้ชัดและห้ามเปลี่ยนความหมายของ `marginpercent` ที่เทียบยอดขาย (`erp-reports.ts:602-607`) ; API ของ `gross_profit_product` ยังไม่พร้อม |
| 5259 รายงานกำไรขั้นต้น(เก่า) | `gross-profit-by-document` `/report/reportgrossprofitbydocument` (`menu-data.ts:282`, `erp-reports.ts:193-208`) | ต่าง | BC มีระดับหัวเอกสาร 6 คอลัมน์ ; Champ ระดับบรรทัดสินค้าจัดกลุ่มตามวัน 15 คอลัมน์ (ราคา/หน่วย, ยอดรวม, ส่วนลด, จำนวนเงิน, Bottom line discount, สุทธิ, ต้นทุน, กำไร, % ต้นทุน, % ยอดขาย) + ช่องติ๊ก 3 ช่อง ; ชื่อ "(เก่า)" — ต้องถามลุงจืดว่ารุ่นปัจจุบันของ Champ คือรายงานไหนก่อนทำ |
| 5184 รายงาน Product Movement Analysis | **ห้ามจับคู่กับ** `product-movement` `/report/reportstockmovement` (`menu-data.ts:539`) ซึ่งเป็นรายงานสต๊อก | ยังไม่มี | เป็นรายงานจัดอันดับสินค้าตามยอดขาย/กำไรพร้อมสัดส่วนและสัดส่วนสะสม ; ชื่อเมนูไทยที่เหมาะต้องขอลุงจืดตัดสินเพราะชื่อเดิมของ Champ ไม่สื่อเนื้อหา |
| 5271 เปรียบเทียบยอดขาย12เดือน | ไม่มี | ยังไม่มี | ต้องมีมุมมอง 12 เดือน 2 แบบ (สรุปแนวตั้ง ; รายสินค้าแนวนอน 2 แถว จำนวน/จำนวนเงิน) ; BC viewer ไม่มีโครงตาราง pivot (`erp-report-viewer.tsx:51-163`) |
| 5916 เปรียบเทียบยอดขายสินค้า 12 เดือน(ตามลูกค้า,สินค้า) | ไม่มี | ไม่ควรทำ | Champ ไม่ได้ต่อเมนูกับรายงานนี้ (ไม่พบ handler) — ถือเป็นข้อยกเว้นของกฎ "ครบเท่า Champ" ต้องให้ลุงจืดยืนยันก่อน |
| 5290 / 5291 Daily Sale By Department (รายวัน / แสดงรายการย่อย) | `sales-daily` (`menu-data.ts:279`, `erp-reports.ts:147-162`) | ต่าง | BC แยกตามวิธีรับเงิน (เงินสด/โอน/ขายเชื่อ) ; Champ แยกตามกลุ่มสินค้า (จำนวนเอกสาร จำนวน ต้นทุน ยอดขาย กำไร Margin) |
| 5295 / 5296 (วัน,ปี) | ไม่มี | ยังไม่มี | เซลล์แบบ "วัน/สะสมปี" ; สะสมปีต้องรองรับปีบัญชี ไม่ใช่ `01/01` ตายตัว |
| 5293 / 5294 (เดือน,ปี) | ไม่มี | ยังไม่มี | สะสมเดือน 1 ถึงเดือนที่เลือก ; คำนวณวันสิ้นเดือนด้วยปฏิทินจริง ไม่ฝังวัน (แก้บั๊กอธิกสุรทิน) |
| 5370 รายงานยอดขายแยกเจ้าหนี้รับฝากขาย | ไม่มี | นอกขอบเขต | ต้องใช้ข้อมูลเจ้าหนี้รับฝากขาย/ระบบขาย — อยู่นอก GL-only ของ BC ตอนนี้ |

#### ข้อปรับร่วมของเครื่องยนต์รายงาน BC (ถ้าลุงจืดสั่งทำ)

1. เพิ่มนิยามตัวกรองต่อรายงานใน `ErpReportConfig` (`erp-reports.ts:11-26`): ช่วง จาก–ถึง (ใส่จากอย่างเดียว = เท่ากับ), checkbox, combo "เรียงตาม" ที่เปลี่ยนหัวกลุ่ม/คอลัมน์ ; ส่งค่าให้ backend ผ่านพารามิเตอร์ ไม่กรองฝั่งเบราว์เซอร์ ; ใช้ข้อความภาษาไทยของ Champ (ป้าย "พิมพ์จากวันที่:" ฯลฯ ตามตารางด้านบน) ผ่าน key ภาษา `languages.tsv` ตามกฎ i18n.
2. ย้ายการรวมยอด, สัดส่วน, % และการจัดอันดับไป backend (Go + `NUMERIC`) ; หน้าเว็บแสดงผลอย่างเดียว — แก้ `erp-report-viewer.tsx:154-163` และ `erp-reports.ts:596-607` ; ในโมดูลที่มีอยู่ `sales_report.go:382,446-451` ใช้ `float64` ต้องเปลี่ยน (กฎ Rule 20).
3. ตัวกรองมาตรฐานของเอกสารขายที่ Champ ใช้ (`IsCancel=0`, `TransState=0`, ตัดแถวข้อความ `.`, ตัดสินค้าชุดแบบที่ 2, ตัดบิลหลังร้านตามตัวเลือกบริษัท) ต้องกำหนดเป็นกติกาเดียวทั้งโมดูล — Champ แต่ละรายงานใช้ไม่สม่ำเสมอ (ดู 5310, 5271 แบบสรุป, 5290 `ProcessDetail`, 5916) ซึ่งทำให้ยอดรายงานคนละตัวไม่ตรงกัน ; BC ไม่ควรลอกความไม่สม่ำเสมอ.
4. รายงานของ Champ ซ่อนคอลัมน์ต้นทุน/กำไรตามสิทธิ์ (`bCanSeeItemCost`, `IsShowProfitLoss` — 5248, 5259, 5277) — BC ต้องมีสิทธิ์ระดับคอลัมน์ในโมเดลสิทธิ์เดียวกับ role-permission ; ยังไม่พบกลไกนี้ใน BC (ยังไม่ verify).
5. ข้อความ/ป้ายที่ Champ เป็นอังกฤษหรือสลับ ("Margin Percent(%)", "Display summary", "Use Bottom Line Discount") — ใช้ key + คำไทยตามความหมาย ไม่คัดอังกฤษ ; ป้ายที่ไทยหายใน rc (กลุ่ม "Daily Sale By Department") ต้องยืนยันคำกับลุงจืดจาก `myresource` จริง.

#### สิ่งที่ขัดกฎเจ้าของ / ควรถามก่อน

- **ขอบเขตปัจจุบัน = GL อย่างเดียว** (AGENTS.md "ขอบเขต GL ล่าสุด") — รายงาน 21 รายการต้องใช้เอกสารขาย/ใบลดหนี้/ใบเพิ่มหนี้/ใบเสนอราคา/สินค้า ทั้งหมดเป็นระบบขาย/ERP ซึ่งยังไม่ทำ ; จึงไม่ควรเพิ่มเมนู/ไฟล์ใด ๆ ตอนนี้ — เสนอเป็นแผนเมื่อลุงจืดเริ่มโมดูลขาย (กฎ "ยึด Champ เป็นต้นแบบ, เสนอก่อนเพิ่ม").
- 5916 ไม่มีทางเปิดใน Champ ; 5259 "(เก่า)" ; 5184 ชื่อไม่ตรงเนื้อหา — 3 รายการนี้ต้องถามลุงจืดว่าจะทำตามชื่อเมนูหรือตามพฤติกรรมจริง.
- ไม่มีรายการใดเกี่ยวกับเงินเดือน/ภ.ง.ด.1/สปส. ; ไม่พบการอ้างอิงผลิตภัณฑ์ภายนอกในเนื้อหาของส่วนนี้.

#### สิ่งที่ยังไม่ verify / อนุมานในส่วนนี้

- ไทยจริงของ key หัวกลุ่ม/คอลัมน์ (เช่น `ITEMGROUP`, `QTY_INCDEBT`, `RATE_COLLECT`, ป้ายโหมดเรียง) อยู่ใน DB `myresource` ซึ่งไม่อยู่ใน repo — ที่ระบุไว้เป็นไทยแท้มาจากคอมเมนต์โค้ดหรือป้ายใน rc เท่านั้น ; ที่เหลือเป็น "(อนุมาน)".
- ไม่ได้รัน Champ จริงและไม่ได้เทียบกับข้อมูลจริง: พฤติกรรมของ SQL ที่ผิดไวยากรณ์ (5236 เงื่อนไขเงิน), จำนวนรับคืนที่หายใน 5248, ผลของ `UPDATE` ในกล่องเงื่อนไข ล้วนมาจากการอ่านโค้ด.
- ลำดับช่องบนจอ (พิกัด x,y) ของ 5277/5310 ไม่ได้ไล่เทียบรายช่อง ใช้ลำดับใน `GetReportCondition` แทน.
- ลำดับ `StockType` (ตัวเลข 4 = รับฝากขาย, 7 = สินค้าชุดแบบที่ 2) อ้างจากคอมโบสินค้า (`icdlgitemmasterpage1.cpp:203-211`) — ยังไม่ verify ว่าตัวเลขเลื่อนเมื่อบริษัทถอดประเภท Serial ออก.
- ไม่ได้เปิดตรวจ `reportCenterView`/`reportcenterconfig.xml` (ไม่อยู่ใน repo) จึงไม่ทราบว่า Report Center เปิดรายงานใดบ้างจริง (โดยเฉพาะ 5916).
