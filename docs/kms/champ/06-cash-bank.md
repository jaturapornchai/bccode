# 06 — ระบบเงินสด-ธนาคาร (Bank/Cash) : พิมพ์เขียว Champ สำหรับ BC

> ที่มา: ลุงจืดสั่ง 2026-10-07 "ออกแบบใหม่ให้ Champ เป็นหลัก ให้เหมือนกันแทบทุกอย่าง" — ศึกษาเมนู Bank/Cash ทั้ง 28 รายการของ Champ (`D:\project-champ`) แล้วเทียบกับ BC (`D:\bccode`)
> ขอบเขต: 12 จอ/กระบวนการ (5178, 5168, 5179, 5177, 5680, 5169, 5172, 5173, 5174, 5175, 5176, 5170) + 16 รายงาน (5149, 5154, 5158, 5160, 5163, 5164, 5165, 5166, 5167, 5150, 5151, 5152, 5684, 5685, 5156, 5157) (atlas `scratch/champ-atlas/atlas-06-Bank_Cash.md`)
> วันที่ศึกษา 2026-10-07 — อ่านโค้ด ไม่ได้รัน Champ จริง; ส่วนที่ไม่ได้พิสูจน์ระบุว่า "ยังไม่ verify". เอกสารคู่กัน: `05-cheque.md` (สถานะเช็ค/ผลต่อสมุดเงินฝาก)

## 0. วิธีอ้างอิง

| ตัวย่อ | ตำแหน่งจริง |
|---|---|
| `C:` | `D:\project-champ\champ\champ\` (`cashfrm*.cpp`, `MainFrm.cpp`, `menuconfigxml\menuconfig.xml`, `BC5Account.rc`) |
| `LIB:` | `D:\project-champ\champ-lib\BC5PROCESS\` (`BookbankProcess.cpp`, `bcprocess.cpp`, `gltlsprocess.cpp`) |
| `REP:` | `D:\project-champ\champ-report\BC5REP_CASH\` (`CASHRep*View.cpp`, `BCRepCashChq.rc`) |
| DDL | atlas.json -> tables (มาจาก `champ\champ\Script\SQLSERVER_Script.sql`) |

ข้อความไทยที่เป็น resource key (เช่น `SUM_DEPOSIT`) ไม่มีไทยใน repo — เขียน key ไว้ และเขียนไทยเฉพาะที่ยืนยันจาก rc / คอมเมนต์ต้นฉบับ / atlas (ที่ไม่ยืนยันเขียน "ยังไม่ verify"). เลขบรรทัดอ้างจากไฟล์ที่ถอดไทยแล้ว (จำนวนบรรทัดเท่าต้นฉบับ)

## 1. เมนู Bank/Cash ตามลำดับของ Champ (menuconfig.xml)

กลุ่มคั่นด้วย separate เหมือนเมนูจริง:

| กลุ่ม | id | ชื่อเมนู | ฟอร์ม | ตารางหลัก |
|---|---|---|---|---|
| A. รับ/จ่ายอื่นๆ | 5178 | บันทึกรับเงินอื่นๆ | `CCASHFrmOtherIncome` | BCOtherIncome (+BCTransSub) |
| | 5168 | บันทึกขอเบิกเงินทดรองจ่าย | `CCASHFrmIssueOtherExpense2` | BCOtherExpense2 / BCOtherExpSub2 |
| | 5179 | บันทึกจ่ายเงินอื่นๆ | `CCASHFrmOtherExpense` | BCOtherExpense (+BCTransSub) |
| B. ข้อมูลและเอกสารธนาคาร | 5177 | กำหนดวงเงินสดย่อย | `CCASHFrmPettyCash` | BCPettyCash |
| | 5680 | บันทึกรายละเอียดผู้ติดต่อธนาคาร | `CCashFrmBankContact` | BCBankConHD / BCBankContact |
| | 5169 | บันทึกสมุดเงินฝากธนาคาร | `CCASHFrmBankBookNew` | BCBankBook |
| | 5172 | บันทึกนำฝากเงินสด | `CCASHFrmCashDeposit` | BCCashBankIn |
| | 5173 | บันทึกถอนเงินสด | `CCASHFrmCashWithdrawal` | BCCashBankOut |
| | 5174 | บันทึกค่าใช้จ่ายธนาคาร | `CCASHFrmBankExpense` | BCBankExpense |
| | 5175 | บันทึกรายได้จากธนาคาร | `CCASHFrmBankIncome` | BCBankIncome |
| | 5176 | บันทึกโอนเงินระหว่างธนาคาร | `CCASHFrmBankTransfer` | BCBankTransfer |
| C. กระบวนการ | 5170 | คำนวณยอดคงเหลือของสมุดบัญชีใหม่ | `CCashDlgBankReprocess` (dialog) | BCBankBook.Balance |
| D. รายงานต่างๆ (popup `disable="yes"`; อยู่ในศูนย์รายงาน) | 5149 5154 5158 5160 5163 | รายงานการฝากเงินสด / การถอนเงินสด / รายได้จากธนาคาร / ค่าใช้จ่ายธนาคาร / การโอนเงินระหว่างธนาคาร | `CCASHRepCashDeposit/Withdrawal/BankIncome/BankExpense/BankTransferView` | |
| | 5164 5165 | รายงาน Bank Statement / รายงาน Statement ล่วงหน้า | `CCASHRepBankStatementView` / `CCASHRepBankAdvanceStatementView` | |
| | 5166 5167 | รายงานการรับเงินอื่นๆ / การจ่ายเงินอื่นๆ พร้อมรายการย่อย | `CCASHRepOtherIncome/OtherExpenseView` | |
| | 5150 5151 5152 | รายงานเคลื่อนไหวเงินสดย่อย / การตั้งเบิกเงินสดย่อย / เคลื่อนไหวเงินสด | `CCASHRepPettyCashMovement/GivePettyCash/CashMovementView` | |
| | 5684 5685 | รายงานบันทึกขอเบิกเงินทดลองจ่าย เรียงตามวันที่เอกสาร / ผู้เบิก | `CCASHRepPayMoneyByDate/ByEmployeeView` | |
| | 5156 5157 | รายงานสมุดจ่ายเงินประจำเดือน เดบิต / เคดิต (สะกดตามเมนูต้นฉบับ) | `CCASHRepBookBankPaymentDebit/CreditView` | |

หมายเหตุ: popup รายงานเป็น `disable="yes"` จึงไม่ขึ้นเมนูหลัก (`XMLMenuConfig.cpp:153`) แต่ขึ้นในศูนย์รายงาน (`reportCenterView.cpp:299-330`)

## 2. สมุดเงินฝากธนาคาร (master) และยอดคงเหลือ

### 2.1 บันทึกสมุดเงินฝากธนาคาร (5169) — `CCASHFrmBankBookNew` (`cashfrmbankbooknew.cpp`) — ตาราง `BCBankBook`

ฟอร์ม `IDD_CASH_FRM_BANKBOOK_NEW` 348x244 (`BC5Account.rc:12150`). ลำดับฟิลด์ตาม rc:

| แถว | ซ้าย | ขวา |
|---|---|---|
| 1 | เลขที่บัญชี (`BookNo`) + ปุ่ม "เปลี่ยนเลขที่บัญชี" | ประเภทบัญชี (combo จาก `BCBankAccount`) |
| 2 | ชื่อบัญชี | - |
| 3 | ธนาคาร (lookup `BCBank`) | สาขาธนาคาร (lookup `BCBankBranch`) |
| 4 | รหัสผังบัญชี GL (key `Chart_Of_Account`, ลิงก์เปิดผังบัญชี) | ยอดคงเหลือ (อ่านอย่างเดียว) |
| 5 | ตาราง "รายชื่อผู้ติดต่อ" | |

| ฟิลด์ | คอลัมน์ | บังคับ | หมายเหตุ |
|---|---|---|---|
| เลขที่บัญชี | `BookNo` (รหัสหลักของสมุดเงินฝาก ไม่มีรหัสสมุดแยกจากเลขที่บัญชี) | ใช่ | ล็อกเมื่อถูกอ้างอิง (`CheckBookNoReference`); เปลี่ยนด้วยปุ่ม "เปลี่ยนเลขที่บัญชี" (ดู 2.2) |
| ประเภทบัญชี | `AccountType` | - | master `BCBankAccount` (Code, Name) |
| ชื่อบัญชี | `Name` | ใช่ | MSG_890 (ข้อความตัดสินจาก key ยังไม่ verify ไทย) |
| ธนาคาร | `BankCode` | - | master `BCBank` |
| สาขาธนาคาร | `BranchCode` | - | master `BCBankBranch` |
| รหัสผังบัญชี GL | `AccountCode` | - | ใช้เป็นบัญชีฝั่งธนาคารของแม่แบบลงบัญชี (เงื่อนไข 3 ดู `05-cheque.md` ข้อ 5.1) |
| ยอดคงเหลือ | `Balance` | - | อ่านอย่างเดียว (`m_edAmount.EnableWindow(FALSE)` `:100`) ปรับโดยเอกสารเงินสด/เช็ค หรือกระบวนการคำนวณใหม่ |
| ตารางรายชื่อผู้ติดต่อ | `BCBankContact` | - | คอลัมน์: `CODE_CONTACT`, `DESCRIPTIONNAME`, `POSITION`, `Address`, `Tel`; โหลดจาก `BCBankContact` ตาม BankCode + BranchCode (`:172-248`) |

การลบ: ถามยืนยัน DELETE_MSG; บันทึก log ผู้ใช้. ข้อความตรวจ MSG_888..891 (ข้อความไทยอยู่ใน resource ยังไม่ verify). DDL `BCBankBook`: ROWORDER, iscancel, isconfirm, confirmcode/datetime, cancelcode/datetime, BookNo, Name, IsExport, Creator/LastEditor, AccountType, BankCode, BranchCode, AccountCode, Balance.

### 2.2 เปลี่ยนเลขที่บัญชี (ปุ่มในจอ 5169) — `CCASHDlgChangeBookNo`

กล่อง `IDD_CASH_DLG_CHANGEBOOKNO` ฟิลด์ Old Book No. / New Book No. + OK / Cancel (caption ใน rc ผิดเป็น "คำนวณยอดคงเหลือของสมุดบัญชีใหม่" — ต้นฉบับพิมพ์ผิด ไม่ต้องลอก). เปลี่ยนค่า `BookNo` ในประมาณ 24 ตารางที่อ้างสมุดเงินฝาก (เอกสารเช็ค, เงินสด-ธนาคาร, ใบเสร็จ ฯลฯ — จำนวนแม่นยำ ยังไม่ verify) (`cashfrmbankbooknew.cpp:262-300`). BC: ถ้าใช้รหัสสมุดแบบ surrogate ไม่จำเป็นต้องมีปุ่มนี้

### 2.3 บันทึกรายละเอียดผู้ติดต่อธนาคาร (5680) — `CCashFrmBankContact`

ฟอร์ม `IDD_CASH_FRM_BANK_CONTACT` (`BC5Account.rc:12259`): ธนาคาร -> สาขาธนาคาร -> ตาราง "รายชื่อผู้ติดต่อ" (คอลัมน์ติ๊กเลือก, `CODE_CONTACT`, `DESCRIPTIONNAME`, `POSITION`, `Address`, `Tel`). บันทึกที่ `BCBankConHD` (BankCode, BranchCode) + `BCBankContact` (BankCode, BranchCode, BookNo, Code, Name, Address, Telephone, PersonPosition, DefaultStatus, PictureFileName, CardPicFileName)

### 2.4 สูตรคำนวณยอดคงเหลือของสมุดเงินฝาก — `LIB:BookbankProcess.cpp:60-200`

`BCBankBook.Balance` = ผลรวมต่อไปนี้ (ทุกตารางกรอง `IsCancel=0`; ถ้าระบุ "ณ วันที่" เพิ่ม `docdate <= วันที่`):

| เครื่องหมาย | แหล่งข้อมูล |
|---|---|
| − | `BCPayMoney.PayAmount` (PaymentType=3, RefNo=BookNo) — จ่ายด้วยการโอน/ตัดบัญชี |
| + | `BCRecMoney.PayAmount` (RefNo=BookNo) — รับเข้าบัญชี |
| + | `BCCashBankIn.Amount` (นำฝากเงินสด) |
| − | `BCCashBankOut.Amount` (ถอนเงินสด) |
| − | `BCBankExpense.Amount` (ค่าใช้จ่ายธนาคาร) |
| + | `BCBankIncome.Amount` (รายได้จากธนาคาร) |
| + | `BCChqInDepo2Sub.NetAmount` (นำเช็คเข้าใหม่) |
| + | `BCChqInPassSub.NetAmount` (เช็ครับผ่าน) |
| − | `BCChqInCancelSub.NetAmount` (เฉพาะ CancelStatus 1 หรือ 3) |
| − | `BCChqInRetSub.NetAmount` (เช็ครับคืน) |
| + | `BCChqInSaleSub.NetAmount` (ขายลดเช็ค) |
| − | `BCChqOutPassSub.NetAmount` (เช็คจ่ายผ่าน) |
| + | `BCChqOutCancelSub.NetAmount` (เฉพาะ CancelStatus 1 หรือ 3) |
| + | `BCCreditPassSub.NetAmount` (ขึ้นเงินบัตรเครดิต) |
| + / − | โอนเข้า (`BCBankTransfer.ToBook`) / โอนออก (`FromBook`) |

ปัดด้วย `XRound` แล้ว `UPDATE BCBankBook SET Balance`. ระหว่างใช้งาน `ProcessBankBook(bookNo, amount, rate, mark)` (`LIB:bcprocess.cpp:3114`) บวกลบทีละเอกสาร `Balance += amount*rate*mark` (rate นำมาคูณเฉพาะ `MultiCurrency=1`). ความคลาดเคลื่อนที่เป็นไปได้ระหว่างยอดสะสมกับสูตรนี้ ดู `05-cheque.md` ข้อ 2.5

### 2.5 คำนวณยอดคงเหลือของสมุดบัญชีใหม่ (5170) — `CCashDlgBankReprocess`

กล่อง `IDD_CASH_DLG_BANKRECALC`: From Book / To Book (ช่วงสมุดเงินฝาก), ปุ่ม Start / Cancel, แถบความคืบหน้า (`BC5Account.rc:10736`; caption "คำนวณยอดคงเหลือของสมุดบัญชีใหม่"). กดแล้วใช้สูตร 2.4 คำนวณใหม่ต่อสมุดเงินฝากในช่วงและเขียนทับ `Balance`. BC: เมนู `/tools/bank-balances` (`erp-tools.ts:84`)

## 3. เอกสารเงินสด-ธนาคาร (5172-5176)

### 3.1 โครงร่วมของ 5172 นำฝากเงินสด / 5173 ถอนเงินสด / 5174 ค่าใช้จ่ายธนาคาร / 5175 รายได้จากธนาคาร

ฟอร์ม `CCASHFrmCashDeposit` ฯลฯ (สืบทอด `CMyEspeciallyFormView`), dialog `IDD_CASH_FRM_CASHDEPOSIT` (`BC5Account.rc:6941`). ลำดับฟิลด์:

| แถว | ซ้าย | ขวา |
|---|---|---|
| 1 | รหัสสมุดเงินฝาก (+ปุ่ม) | ช่องติ๊ก "บันทึกยอดเงินยกมา" (เฉพาะฟอร์มนำฝาก) |
| 2 | ชื่อบัญชี (อ่านอย่างเดียว) | - |
| 3 | รหัสธนาคาร (แสดงชื่อ, อ่านอย่างเดียว) | - |
| 4 | สาขาของธนาคาร (อ่านอย่างเดียว) | - |
| 5 | เลขที่เอกสาร (+ปุ่ม) | วันที่เอกสาร (+ปฏิทิน) |
| 6 | รหัสผังบัญชี (ฝั่งตรงข้ามของสมุดเงินฝาก; key `Chart_Of_Account`, rc ใส่ข้อความสำรอง "รหัสสมุดรายวัน") | จำนวนเงิน |
| 7 | รหัสสมุดรายวัน (`GLBookCode`, lookup `BCGLBook`) | - |
| 8 | คำอธิบาย | - |
| 9 | รหัสแผนก | - |
| 10 | รหัสการจัดสรร | - |
| 11 | รหัสโครงการ | - |
| 12-14 | ผู้บันทึก/วันที่บันทึก, ผู้แก้ไข/วันที่แก้ไข, ผู้ใช้ขณะนี้/เวลาเริ่มใช้ (ข้อมูลระบบ อ่านอย่างเดียว) | |

| ฟิลด์ | คอลัมน์ | บังคับ | หมายเหตุ |
|---|---|---|---|
| รหัสสมุดเงินฝาก | BookNo | ใช่ | CANNOT_SAVE_NOBOOKBANK (`cashfrmcashdeposit.cpp:343`) |
| เลขที่เอกสาร | DocNo | ใช่ | CANNOT_SAVE_NODOCNO (`:350`) |
| วันที่เอกสาร | DocDate | ใช่ | CANNOT_SAVE_NODOCDATE (`:357`); งวดปิดห้ามบันทึก/ลบ (`IsPeriodClose` `:180,:276`) |
| รหัสผังบัญชี | AccountCode | ใช่ (เมื่อมีโมดูล GL) | CANNOT_SAVE_NOFORMATGL (`:303`) |
| จำนวนเงิน | Amount | ใช่ ต้อง > 0 | CANNOT_SAVE_AMOUNTERROR (`:365,:375`) ยกเว้นกรณียกมา |
| รหัสสมุดรายวัน | GLBookCode | ใช่ ยกเว้นติ๊กยกมา | CANNOT_SAVE_NODEFAULT_BOOKDAILY (`:313`) |
| บันทึกยอดเงินยกมา | IsBankBalance | - | ติ๊ก = ไม่ลงบัญชี (ข้ามการเรียก `ProcessConditionGroup20` `:455-470`) แต่ยังบวกยอดเข้าสมุดเงินฝาก (`:477`) |
| คำอธิบาย/แผนก/การจัดสรร/โครงการ | MyDescription, DepartCode, AllocateCode, ProjectCode | - | |
| สถานะเอกสาร | IsConfirm, IsCancel, IsPostGL, PrintCount, JobNo, ConfirmCode.., CancelCode.. | - | แก้ไม่ได้เมื่อยืนยันแล้ว (CANNOT_UPDATE_CONFIRM `:325`) / ลงบัญชีแล้ว (`IsPostGL` `:297`); ยกเลิกเอกสารได้ (CANNOT_CANCEL `:625,:988`) |

ผลของเอกสาร:

| id | เอกสาร | ตาราง | เครื่องหมายต่อ Balance | GL source / group |
|---|---|---|---|---|
| 5172 | นำฝากเงินสด | BCCashBankIn | + (`ProcessBankBook(...,1.0,+1)` `:477`) | 14 / group 20 (`gltlsprocess.cpp` `ProcessConditionGroup20`) |
| 5173 | ถอนเงินสด | BCCashBankOut | − | 15 / group 21 |
| 5174 | ค่าใช้จ่ายธนาคาร | BCBankExpense | − | 16 / group 22 |
| 5175 | รายได้จากธนาคาร | BCBankIncome | + | 17 / group 23 |

ลำดับตอนบันทึก (5172): แก้เอกสารเดิม -> ย้อนยอดเดิม (`ProcessBankBook ... -1` `:405`) -> insert/update -> ลงบัญชี (อัตโนมัติ `m_nManualKey=0` หรือลงมือ `OnSaveAccount`) -> บวกยอดใหม่ -> `SaveLog`. ลบ (`:214-262`): ถ้ายืนยันแล้วห้ามลบ (CANNOT_DELETE_CONFIRM) -> ย้อนยอด (`:252`) -> `DeleteAccount(DocNo, 14)`. หมายเหตุ Champ มี flag ปิดโมดูล GL ทั้งชุด (`LMD_REMOVE_GL`) — BC เป็นระบบ GL เป็นหลัก ไม่ต้องลอก flag นี้

### 3.2 บันทึกโอนเงินระหว่างธนาคาร (5176) — `CCASHFrmBankTransfer` — ตาราง `BCBankTransfer`

`IDD_CASH_FRM_BANKTRANSFER` (`BC5Account.rc:7025`) แบ่ง 3 กลุ่ม:

| กลุ่ม | ฟิลด์ตามลำดับ |
|---|---|
| โอนจากบัญชีเลขที่ | เลขที่บัญชี (lookup สมุดเงินฝาก) / ชื่อบัญชี / รหัสธนาคาร / สาขาของธนาคาร (3 ช่องหลังอ่านอย่างเดียว) |
| โอนเข้าบัญชีเลขที่ | เลขที่บัญชี / ชื่อบัญชี / รหัสธนาคาร / สาขาของธนาคาร |
| รายละเอียดการโอน | เลขที่เอกสาร + วันที่เอกสาร / จำนวนเงิน / รหัสสมุดรายวัน / คำอธิบาย |

ตารางมีคอลัมน์ `FromBook, ToBook, GLBookCode, DepartCode, AllocateCode, ProjectCode, Amount, ChargeAmount, TotalAmount` แต่ไดอะล็อกไม่มีช่องค่าธรรมเนียม/ยอดรวม (`ChargeAmount/TotalAmount` ไม่ถูกคีย์ — ยังไม่ verify การเติมค่า). ผล: `FromBook` ลด / `ToBook` เพิ่ม (`cashfrmbanktransfer.cpp:381-440`); GL source 18 / group 24. สถานะเอกสารเหมือน 3.1

## 4. รับเงินอื่นๆ / จ่ายเงินอื่นๆ / ขอเบิกเงินทดรองจ่าย

### 4.1 บันทึกรับเงินอื่นๆ (5178) และ บันทึกจ่ายเงินอื่นๆ (5179) — `CCASHFrmOtherIncome` / `CCASHFrmOtherExpense`

ฟอร์มแบบ "สมุดรายวันย่อ": หัว + กริดบัญชี (เดบิต/เครดิต) + วิธีรับ/จ่ายเงิน (`IDD_CASH_FRM_OTHERINCOME` `BC5Account.rc:6687`, `IDD_CASH_FRM_OTHEREXPENSE` `:6803`)

หัวเอกสาร (ลำดับ):

| แถว | 5178 รับเงินอื่นๆ | 5179 จ่ายเงินอื่นๆ |
|---|---|---|
| 1 | เลขที่เอกสาร / วันที่เอกสาร | เหมือนกัน |
| 2 | กลุ่มเอกสาร / รหัสสมุดบัญชี (`GLBookCode`) | เหมือนกัน |
| 3 | - | อ้างอิงเอกสารขอเบิกเงินทดลองจ่าย (`RefDocNo`) |
| 3/4 | รหัสการจัดสรร / รหัสโครงการ | เหมือนกัน |
| 4/5 | รับจาก(ลูกค้า) + ชื่อ (`ArCode`) | จ่ายให้(เจ้าหนี้) + ชื่อ (`ApCode`) |
| 5/6 | คำอธิบาย | คำอธิบาย |

กริด (ลำดับคอลัมน์ ยืนยันจาก `CASHRepOtherIncomeView.cpp:164-171`): รหัสผังบัญชี (`Chart_Of_Account`), ชื่อ (key `100004`), คำอธิบาย, รหัสแผนก, รหัสการจัดสรร, รหัสโครงการ, เดบิต, เครดิต
ท้ายเอกสาร: รหัสแผนก; 5178 "รวมทั้งสิ้น" / 5179 "จำนวนเงิน"; รวมเดบิต; รวมเครดิต
กติกาบันทึก: ผลรวมเดบิต = ผลรวมเครดิต และไม่เป็นศูนย์ (CANNOT_SAVE_NORIGHT_DEBITCREDIT); วิธีรับ/จ่ายเงินต้องเท่ายอดสุทธิ — ผ่านกล่องรับ/จ่ายเงินกลาง (`OnReceive2`: เงินสด / เช็ค / บัตรเครดิต / ธนาคาร / ภาษีหัก ณ ที่จ่าย / เงินสดย่อย) เก็บเป็น `SumCashAmount`, `SumChqAmount`, `SumCreditAmount`, `SumBankAmount`, `SumOfWTax`, `PettyCashAmount`; ตารางหัว `BCOtherIncome` / `BCOtherExpense` มี `SumofDebit`, `SumofCredit`, `NetAmount`, `OtherIncome`, `OtherExpense`, `ExcessAmount1/2`, `BillGroup`; รายการบัญชีเก็บใน `BCTransSub` (Source = 19 รับ / 20 จ่าย) (สรุปจากโค้ดฟอร์ม; รายละเอียดกล่องรับ/จ่ายอยู่ในเอกสารระบบลูกหนี้-เจ้าหนี้)
ผลต่อสมุดเงินฝาก: ถ้าจ่าย/รับด้วยเงินโอนธนาคาร ยอดเข้า `BCPayMoney`/`BCRecMoney` ซึ่งสูตร 2.4 นับอยู่แล้ว (ไม่เรียก `ProcessBankBook` ตรงในฟอร์มนี้ — ยังไม่ verify ทุกกรณี)

### 4.2 บันทึกขอเบิกเงินทดรองจ่าย (5168) — `CCASHFrmIssueOtherExpense2` — `BCOtherExpense2` / `BCOtherExpSub2`

หัวเอกสาร (ลำดับตาม `BC5Account.rc:6741`): เลขที่เอกสาร / วันที่เอกสาร; รหัสการจัดสรร / รหัสโครงการ; เจ้าหนี้ (+ชื่อ); คำอธิบาย; [กริด]; รหัสสมุดรายวัน / รหัสผังบัญชี; ท้ายเอกสาร: รหัสแผนก / จำนวนเงินรวม; รหัสพนักงาน (`PersonCode`, lookup `BCSale`)
กริด: `PAY_COMMENT` (คำอธิบายรายการ) + `AMOUNT` (จำนวนเงิน) -> `BCOtherExpSub2` (DepartCode, Amount, AllocateCode, ProjectCode, ApCode)
ตารางหัวมี: `SumofAmount`, `ApCode`, `PersonCode`, `BillStatus`, `GLFormat`, `GLTransData`, `IsPostGL`, `PettyCashAmount`, `SumCashAmount`, `SumChqAmount`, `SumBankAmount`, `MyDescription1/2`
GL source 37 / group 37. เป็นเอกสาร "ขอเบิกเงินทดรอง" ที่ 5179 อ้างอิงกลับ (`RefDocNo`) เมื่อจ่ายจริง; รายงาน 5684/5685 อ่านจากตารางนี้

## 5. กำหนดวงเงินสดย่อย (5177) — `CCASHFrmPettyCash` — `BCPettyCash`

ฟอร์มกริดอย่างเดียว (ไม่มีหัวเอกสาร, `IDD_CASH_FRM_PETTYCASH` `BC5Account.rc:6675`). คอลัมน์: รหัส (`ID`), ชื่อ (`NAME`), วงเงิน (key `72001`), ยอดยกมา (`BEGIN_QTY` -> `BeginAmount`), ยอดคงเหลือ (`BALANCE`, อ่านอย่างเดียว) (`cashfrmpettycash.cpp:172-180`). การบันทึก: ลบทั้งตารางแล้วเพิ่มใหม่ทุกแถว `BCPettyCash(Code, Name, Amount, BeginAmount, Balance)` (`:322-333`) — พฤติกรรม "ลบทั้งหมดแล้วใส่ใหม่" ทำให้ Balance ที่เคยสะสมถูกเขียนทับด้วยค่าในกริด (ควรระวังใน BC)

## 6. รายงาน (16 รายการ)

ไดอะล็อกเงื่อนไขของรายงานเงินสด-ธนาคารอยู่ใน `REP:BCRepCashChq.rc` (dialog id 35001-35051; ใช้กล่อง "เงื่อนไขการพิมพ์รายงานมาตรฐาน")

| id | รายงาน | ไดอะล็อก/ช่องกรอง | คอลัมน์ (ลำดับ) | ยอดรวม |
|---|---|---|---|---|
| 5149 | รายงานการฝากเงินสด (key 74001) | จาก/ถึง วันที่เอกสาร, เลขที่เอกสาร, สมุดเงินฝาก, สมุดรายวัน (`GLBook`) | วันที่ (`AS_DATE`), เลขที่เอกสาร, สมุดเงินฝาก (`BOOK_NO`), หมายเหตุ/รายละเอียด (`REMARK_DETAIL`), จำนวนเงิน (`AMOUNT`) | ยังไม่ verify |
| 5154 | รายงานการถอนเงินสด | เหมือน 5149 | เหมือน 5149 | ยังไม่ verify |
| 5158 | รายงานรายได้จากธนาคาร | เหมือน 5149 | เหมือน 5149 | ยังไม่ verify |
| 5160 | รายงานค่าใช้จ่ายธนาคาร (key 74004) | เหมือน 5149 | เหมือน 5149 | ยังไม่ verify |
| 5163 | รายงานการโอนเงินระหว่างธนาคาร (key 74005) | วันที่ + เลขที่เอกสารเท่านั้น | วันที่, เลขที่เอกสาร, โอนจากสมุดเงินฝาก (`TRANSFER_DEPOSIT_BOOK`), โอนเข้าสมุดเงินฝาก (`TRANSFER_TO_DEPOSIT_BOOK`), หมายเหตุ/รายละเอียด, จำนวนเงิน | ยังไม่ verify |
| 5164 | รายงาน Bank Statement | ดูข้อ 6.1 | ดูข้อ 6.1 | ดูข้อ 6.1 |
| 5165 | รายงาน Statement ล่วงหน้า (key 74007) | dialog 35001: จาก/ถึง วันที่ถึงกำหนด (`DUEDATE`), จาก/ถึง เลขที่เช็ค/บัตร (`CHQ_CREDIT`), จาก/ถึง เลขที่เอกสาร | วันที่ถึงกำหนด (`DUE_DATE`), เลขที่เช็ค/บัตร (`CHEQ_NO_CARD`), ประเภท (`ITEM_TYPE`: เช็ครับ `MST_CHQIN` / บัตรเครดิต `CREDITCARD` / เช็คจ่าย `MST_CHQOUT`), เลขที่เอกสาร, สมุดเงินฝาก (`BOOK_NO`), หมายเหตุ (`REMARK`), รับเงิน (`RECEIP_MONEY`), จ่ายเงิน (`PAY_AMOUNT`) | ยังไม่ verify |
| 5166 | รายงานการรับเงินอื่นๆ พร้อมรายการย่อย (key 74008) | dialog 35013/35014: จาก/ถึง วันที่, เลขที่เอกสาร, รหัสสมุดรายวัน | แถวหัว: เลขที่เอกสาร, วันที่, สมุดรายวัน (`GL_BOOKCODE`), ผู้รับ (`RECEIPT_ARCODE`), แผนก (`DEPART_CODE`), ยอดรวม (`TOTAL_AMOUNT`); แถวคำอธิบาย (`DESCRIPTION`); แถวรายการ: รหัสผังบัญชี (`Chart_Of_Account`), ชื่อ (`100004`), คำอธิบาย, แผนก, การจัดสรร, โครงการ, เดบิต, เครดิต | ยังไม่ verify |
| 5167 | รายงานการจ่ายเงินอื่นๆ พร้อมรายการย่อย (key 74009) | เหมือน 5166 | เหมือน 5166 แต่ผู้จ่ายให้ (`TO_APCODE`) | ยังไม่ verify |
| 5150 | รายงานเคลื่อนไหวเงินสดย่อย (key 74010) | dialog 35018: จาก/ถึง วันที่ + จาก/ถึง รหัสเงินสดย่อย | หัวกลุ่ม: `DEBT_TOTAL_SUB`/`DEBT_SUB`; รายการ: วันที่ (`DOCDATE`), เลขที่เอกสาร, คำอธิบาย (`COMMENT&DESCRIPT`), รับ (`RECEIPTAMOUNT`), จ่าย (`PAY_AMOUNT`), คงเหลือ (`BALANCE`) | ยังไม่ verify |
| 5151 | รายงานการตั้งเบิกเงินสดย่อย (key 74011) | dialog 35012: รหัสเงินสดย่อย, จาก/ถึง วันที่เบิก (`SPEND_DATE`) | เหมือน 5150 | ยังไม่ verify |
| 5152 | รายงานเคลื่อนไหวเงินสด (key 74012) | dialog 35009: จากวันที่ / ถึงวันที่ | วันที่, เลขที่เอกสาร, คำอธิบาย, รับ, จ่าย, คงเหลือ | ยังไม่ verify |
| 5684 | รายงานบันทึกขอเบิกเงินทดลองจ่าย - เรียงตามวันที่เอกสาร (key 74014) | dialog 35016: วันที่เอกสาร, เลขที่เอกสาร, รหัสพนักงาน, combo เงื่อนไข (`CONDITION`) | วันที่, เลขที่เอกสาร, แผนก (`DEPART`), โครงการ (`PROJECT`), ผู้เบิก (`EMPNAME`); รายการย่อย: `DESCRIPTIONNAME`, `AMOUNT` | ยังไม่ verify |
| 5685 | เหมือน 5684 - เรียงตามผู้เบิก (key 74015) | dialog 35017 | กลุ่มตามรหัสพนักงาน (`EMPCODE`, `EMPNAME`) -> วันที่, เลขที่เอกสาร, แผนก, โครงการ, หมายเหตุ/รายละเอียด; ย่อย: ชื่อรายการ, จำนวนเงิน | ยังไม่ verify |
| 5156 | รายงานสมุดจ่ายเงินประจำเดือน - เดบิต (key 74020 + เดือน ปี) | dialog 35008: ปี (`YEAR` "ปี:"), เดือน (`AT_MONTH` "พิมพ์ประจำเดือน:"), จาก/ถึง สมุดเงินฝาก, ช่องติ๊ก "แสดงยอดยกมา,ยกไปแต่ละหน้า" | วันที่ (`AS_DATE`), รายการ (`TRANSAC`), ใบสำคัญจ่ายเลขที่ (`VOUCHER_PAY`), จำนวนเงิน (`AMOUNT`), จ่ายโดยเช็คเลขที่ (`PAY_CHQNO`), จ่ายโดย (`PAYMENT_BY`), รายได้อื่นๆ, ภาษีหัก ณ ที่จ่าย, อื่นๆ, เดบิต (`DEBIT`); แถวหัวที่สอง `NO_DOC`, `BATH` | ยอดยกมา/ยกไปต่อหน้า (ตามช่องติ๊ก) |
| 5157 | รายงานสมุดจ่ายเงินประจำเดือน - เครดิต (key 74030) | เหมือน 5156 | วันที่, รายการ, ใบสำคัญรับเลขที่ (`VOUCHER_REC`), จำนวนเงิน, รับโดย (`RECEIPT_BY`), ค่าใช้จ่ายอื่นๆ, ภาษีหัก ณ ที่จ่าย, อื่นๆ, เครดิต (`CREDIT`; คอมเมนต์ต้นฉบับพิมพ์ "เดบิต" ซ้ำ); หัวสอง: `NO_DOC`, `BATH`, `CASH`, ธนาคาร 1, ธนาคาร 2 (ชื่อจากข้อมูล), `BANK_OTHER`, `PAY_OTHER`, `SUM_OF_TAX`, `OTHER` | ยอดยกมา/ยกไป |

ข้อสังเกตรายงาน 5156/5157: เมนู (id 5156/5157 = `ID_BANKSYS_REPORT_21/22`) ไม่พบ `ON_COMMAND` ใน `MainFrm.cpp`; ตัวเปิดฟอร์มจริงเป็น `ID_BANKSYS_REPORT_20` (5155) และ `ID_BANKSYS_REPORT_30` (5159) ซึ่งไม่ได้อยู่ในเมนู (`MainFrm.cpp:335-336,8630-8650`) — ยังไม่ verify ว่าผู้ใช้เข้าได้อย่างไร (มีความเป็นไปได้ที่ศูนย์รายงานเรียกด้วย resource id). รายงานเพิ่มเติมที่โค้ดมีแต่ไม่อยู่ในเมนู: ยอดเงินคงเหลือของแต่ละบัญชี (`ID_BANKSYS_REPORT_16`=5153, `CCASHRepBankStatement2View`, dialog `IDD_CASH_REP_BANKSTATEMENT_NEW` 35006, key 74016: สมุดเงินฝากจาก/ถึง + "ณ. วันที่"; คอลัมน์ `BANK_BOOK_CODE`, `BANK_BOOK_NAME`, `QTY`; ยอดรวม `SUM_BOOKBANKAMOUNT`, `MainFrm.cpp:10099-10108`), `CCASHRepSumPaymentView` (5161), `CCASHRepSumReceiptView` (5162)

### 6.1 รายงาน Bank Statement (5164) — `CCASHRepBankStatementView` (`REP:CASHRepBankStatementView.cpp`)

เงื่อนไข (dialog 35005): จาก/ถึง สมุดเงินฝาก (`BOOKNO`/`TO_BOOKNO`), จาก/ถึง งวด (`FROM_PERIOD`/`TO_PERIOD`, ตัวหมุนเลือกงวด + แสดงช่วงวันที่), ช่องติ๊ก "แสดงแบบสรุปรายวัน"
เนื้อหา: จัดกลุ่มตามสมุดเงินฝาก (หัวกลุ่ม `BANK_BOOK_CODE`/ชื่อ) แถวแรก = ยอดยกมา (`TRANSFER__BALANCE`, คำนวณจากทุกรายการก่อนวันเริ่มด้วย `GetBalance`) แล้วรายการเคลื่อนไหว
คอลัมน์: วันที่ (`DATE`), เลขที่เอกสาร (`DOCNO`), เลขที่เช็ค/บัตร (`CHEQ_NO_CARD`), ประเภทรายการ (`LIST_DAY` ข้อความจาก `OnGetTransDesc`), คำอธิบาย (`COMMENT`), `REMARK_BILL`, ฝาก (`SUM_DEPOSIT`), ถอน (`SUM_WITHDRAW`), คงเหลือ (`BALANCE`), สกุลเงิน (`NAM_CURRENCY`), อัตราแลกเปลี่ยน (`EXCHANGE_RATE`)
ที่มาของรายการ (UNION): `BCPayMoney` / `BCRecMoney` (ตามวันที่ธนาคาร `TransBankDate`, PaymentType=3 RefNo=BookNo), `BCCashBankIn/Out`, `BCBankExpense/Income`, `BCChqInDepo2Sub`, `BCChqInPassSub`, `BCChqInCancelSub` (CancelStatus 1 และ 3), `BCChqInRetSub`, `BCChqInSaleSub`, `BCChqOutPassSub`, `BCChqOutCancelSub`, `BCCreditPassSub`, `BCBankTransfer` (ฝั่ง ToBook และ FromBook); เรียง `bookno, TransBankDate, MyDocOrder, docno, MyMark, MyNo`; ค่าธรรมเนียมแสดงเป็นแถวแยกจาก `Expend`
ยอดรวม: รวมต่อสมุดและรวมทั้งสิ้น "SUMDETAIL_WORD n TRANSAC" ; ไม่มีรายการ -> "NO_LIST"; ยอดรวมต่อหน้า `SUM_LIST_PAGE`
นี่คือ "รายการเคลื่อนไหวฝั่งสมุด (book)" เรียงตามวันที่ธนาคาร — ไม่ใช่การกระทบยอดกับ Statement ของธนาคาร (ดูข้อ 7)

## 7. กระทบยอดธนาคาร (Bank Reconciliation): Champ กับ BC

### 7.1 Champ

ค้นแล้วไม่มีฟังก์ชันกระทบยอดกับ Statement ของธนาคารเลย (ไม่พบคำว่า reconcil ใน source, ไม่มีตาราง/เมนู). สิ่งที่ใกล้เคียง:
1. "รายงาน Bank Statement" (5164) = รายการฝั่งสมุดแบบสมุดบัญชีธนาคาร (ข้อ 6.1) ผู้ใช้ถือเทียบกับ Statement จริงด้วยตา
2. "รายงาน Statement ล่วงหน้า" (5165) = รายการเช็ครับ/บัตร/เช็คจ่ายที่ยังไม่ถึงกำหนด/ยังไม่ขึ้นเงิน (ใช้ประมาณการล่วงหน้า)
3. วันที่ธนาคาร (`TransBankDate`) บนใบเสร็จ/ใบสำคัญจ่าย ที่ผู้ใช้กรอกเมื่อเงินเข้าบัญชีจริง — เป็นวิธีเดียวที่ Champ ใช้ "ยืนยันว่าตรงกับธนาคาร" (ยังไม่ verify ชื่อฟิลด์ในหน้าจอ)
4. ยอดคงเหลือ `BCBankBook.Balance` + 5170 คำนวณใหม่

### 7.2 BC ปัจจุบัน (GL)

| รายการ | ที่อยู่ | สถานะ |
|---|---|---|
| ตารางบัญชีธนาคาร `gl_subledger_bank_accounts` | `backend/internal/generalledger/subledger.sql:7` | มี |
| บรรทัดสมุดฝั่งธนาคาร `gl_subledger_bank_lines` (journal_id, line_number, bank_account_code, direction 1/2, amount>0) | `subledger.sql:37` | มี |
| บรรทัด Statement `gl_subledger_statements` (id, owner_journal_id, bank_account_code, source_key UNIQUE ต่อบัญชี, transaction_date, direction, amount; payload: value_date, bank_reference, description, balance_after) | `subledger.sql:42-49`, `subledger_types.go:83-94` | มี |
| การจับคู่ many-to-many `gl_subledger_matches` (statement_id <-> journal_id+line_number, amount, ยกเลิกได้พร้อมเหตุผล, คู่ซ้ำที่ยัง active ไม่ได้) | `subledger.sql:50-58`, `subledger_bank.go:123-205` | มี; จับคู่ได้เฉพาะใบสำคัญที่ผ่านบัญชีแล้ว (`:140`) |
| รายงาน `bank-unmatched` (Statement ที่ยังจับคู่ไม่หมด: บัญชีธนาคาร/วันที่/อ้างอิง/รับ-จ่าย/ยอด Statement/จับคู่แล้ว/ยังไม่จับคู่ ณ วันที่) | `subledger_queries.go:159-165`, `reports.go:24` | มี (ฝั่ง Statement ข้างเดียว) |
| รายการฝั่งสมุด (book) ที่ยังไม่ถูกจับคู่ | - | ไม่มี |
| สรุปกระทบยอด: ยอดสมุด ณ วันที่ เทียบยอด Statement พร้อมรายการผลต่าง | - | ไม่มี |
| รายงานรายการจับคู่แล้ว / ประวัติยกเลิกจับคู่ | ตาราง `gl_subledger_audit` มี แต่ไม่มีรายงาน | ไม่มีรายงาน |
| รายงาน "Bank Statement" ฝั่งสมุดตามแบบ Champ (5164) | เมนู `/report/bankstatement` | รอพัฒนา (ยังไม่มีรหัสรายงานใน `API_READY_REPORTS` `erp-reports.ts:530-543`) |

เจตนารมณ์ของลุงจืด (`mydocs/specs/gl/ar-ap-bank-design-notes.md:54-58`): Statement เข้าภายหลังได้, ไม่สร้าง Statement ปลอมจาก GL, จับคู่หลายต่อหลายและยกเลิกได้ (ไฟล์ read-only อ่านอย่างเดียว); และบรรทัด 49: workflow สั่งจ่าย/คืนเงิน/เช็คเป็นหน้าที่ระบบต้นทาง; บรรทัด 51 ระบุว่ารายงาน `CASHRepBankStatementView` ของ Champ เป็นต้นแบบฝั่งสมุด

### 7.3 สิ่งที่ BC ต้องเพิ่มให้เป็น "สองฝั่ง" (ข้อเสนอ รอลุงจืดอนุมัติ — ไม่มีใน Champ จึงเข้าข่ายกฎ Strict Parity ต้องปรึกษาก่อน)

หมายเหตุกฎ: AGENTS.md "ยึด project-champ เป็นต้นแบบ ไม่เพิ่มฟังก์ชันเอง" — แต่ spec GL ของลุงจืดสั่งให้มีการจับคู่ Statement อยู่แล้ว ข้อเสนอด้านล่างเป็นเพียง "รายงานอ่านอย่างเดียว" ต่อยอดจากตารางที่มีอยู่ ไม่เพิ่มเมนูหน้าจอใหม่เกินความจำเป็น

| # | รายงาน/ส่วนที่ขาด | ข้อมูลที่ใช้ | หน้าตา (เลียน Champ 5164) |
|---|---|---|---|
| 1 | รายงาน Bank Statement ฝั่งสมุด (5164) | `gl_subledger_bank_lines` ต่อบัญชีธนาคาร ณ ช่วงวันที่ | กลุ่มตามบัญชีธนาคาร, แถวยกมา, คอลัมน์ วันที่ / เลขที่เอกสาร / เลขที่เช็ค / ประเภทรายการ / คำอธิบาย / ฝาก / ถอน / คงเหลือ ตามข้อ 6.1 |
| 2 | รายการสมุดที่ยังไม่จับคู่ (book-side unmatched) | `bank_lines` ที่ยอดจับคู่ ≠ ยอดบรรทัด | คอลัมน์เหมือน `bank-unmatched` แต่ฝั่งสมุด (เลขที่ใบสำคัญ, บรรทัด, วันที่, รับ-จ่าย, ยอด, จับคู่แล้ว, คงเหลือ) |
| 3 | สรุปกระทบยอด ณ วันที่ | ยอดสมุด = ผลรวมบรรทัดฝั่งธนาคาร (รับ - จ่าย) ณ วัน; ยอด Statement = `balance_after` ของบรรทัดล่าสุด ถ้ามี หรือผลรวม; ผลต่าง = รายการสมุดที่ยังไม่ขึ้น Statement +/- รายการ Statement ที่ยังไม่มีในสมุด | แถว: ยอดตามสมุด, บวก/ลบรายการสมุดที่ยังไม่จับคู่, ยอดตาม Statement, บวก/ลบรายการ Statement ที่ยังไม่จับคู่, ผลต่าง (ต้องเป็น 0) |
| 4 | รายการจับคู่แล้ว (ตรวจย้อนหลัง) | `gl_subledger_matches` active + ที่ถูกยกเลิก (reversed) | คอลัมน์: Statement, ใบสำคัญ/บรรทัด, ยอดจับคู่, ผู้จับคู่/เวลา, เหตุผลยกเลิก |

ข้อควรระวังในการออกแบบ: ยอดสมุดต้องมาจากบรรทัดใบสำคัญที่ผ่านบัญชีแล้วเท่านั้น (ตรงกับ `subledger_bank.go:140`); ห้ามใช้ยอดสะสมแบบบวก/ลบ (ดู `05-cheque.md` 2.5) ให้คำนวณจากรายการ; ตัวเลขใช้ decimal/string ตามกฎ 20; สกุลเงินต่างประเทศ spec ระบุว่ายังไม่ครอบคลุม (`ar-ap-bank-design-notes.md:49`)

## 8. BC เทียบ Champ และสิ่งที่ต้องปรับ

สถานะ BC (ตรวจ 2026-10-07): เมนู BC กลุ่ม `cash-bank` อยู่ที่ `frontend/src/lib/menu-data.ts:376-460` ใช้ชื่อไทยตาม Champ แล้วเกือบทั้งหมด; จอธุรกรรมทุกจอ (`/transaction/otherincomereceipt`, `/transaction/employeeadvance`, `/transaction/expense`, `/banking/*`, `/transaction/accounttransfer`) อยู่ใน `isErpTransactionRoute` (`erp-transaction.ts:217,377-816`) จึงแสดง "รอพัฒนา" (`menu-screen-status.ts:46-50` isMenuBackendRetired); สมุดเงินฝาก `/bookbankscreen` เป็นจอตั้งค่า (`system-setting-screens.ts:893-925`) ที่ชี้ API เดิม `/payment/bookbank` ซึ่งไม่อยู่ในรายการ PostgreSQL ที่ใช้ได้ (`menu-screen-status.ts` `POSTGRES_SETTING_BASE_PATHS` + test บรรทัด 45 ยืนยัน retired)

### 8.1 ตารางเทียบรายเมนู

| Champ id | ชื่อ Champ | BC menu id / route | สถานะ | สิ่งที่ต้องปรับ |
|---|---|---|---|---|
| 5178 | บันทึกรับเงินอื่นๆ | `other-income-receipt` `/transaction/otherincomereceipt` | ต่าง + รอพัฒนา | ทำตามข้อ 4.1: หัว (กลุ่มเอกสาร, รหัสสมุดบัญชี, รับจาก(ลูกค้า)), กริดบัญชี 8 คอลัมน์ เดบิต/เครดิต, ตรวจเดบิต=เครดิต, กล่องรับเงิน. หมายเหตุ: BC GL มีใบสำคัญรายวันอยู่แล้ว — ต้องตัดสินใจว่าเอกสารนี้ลงเป็นใบสำคัญชนิดใด (สมุดรายวันเป็น master ที่ user เพิ่มได้) |
| 5168 | บันทึกขอเบิกเงินทดรองจ่าย | `employee-advance` `/transaction/employeeadvance` | ต่าง + รอพัฒนา | ข้อ 4.2 (กริด คำอธิบาย + จำนวนเงิน, รหัสพนักงาน, เจ้าหนี้). ตรวจกฎเงินเดือน: เป็นเงินทดรองจ่ายพนักงาน (งานการเงิน) ไม่ใช่เงินเดือน — อยู่ในขอบเขตตาม AGENTS.md |
| 5179 | บันทึกจ่ายเงินอื่นๆ | `expense-record` `/transaction/expense` | ต่าง + รอพัฒนา | ข้อ 4.1 + ช่อง "อ้างอิงเอกสารขอเบิกเงินทดลองจ่าย" |
| 5177 | กำหนดวงเงินสดย่อย | `petty-cash-limit` `/banking/petty-cash-limit` | ต่าง + รอพัฒนา | ข้อ 5: กริด 5 คอลัมน์ (รหัส, ชื่อ, วงเงิน, ยอดยกมา, ยอดคงเหลือ อ่านอย่างเดียว) — ไม่ลบทั้งตารางแล้วใส่ใหม่ |
| 5680 | บันทึกรายละเอียดผู้ติดต่อธนาคาร | `bank-contact` `/banking/contacts` | ต่าง + รอพัฒนา | ข้อ 2.3: ธนาคาร + สาขา + กริดผู้ติดต่อ (รหัส, ชื่อ, ตำแหน่ง, ที่อยู่, โทร) |
| 5169 | บันทึกสมุดเงินฝากธนาคาร | `book-bank` `/bookbankscreen` | ต่าง | ฟิลด์ BC ตอนนี้: รหัสสมุดบัญชี, ชื่อสมุดบัญชี, เลขที่บัญชี, สาขาธนาคาร (ข้อความอิสระ), ชื่อบัญชี, รหัสธนาคาร (ข้อความอิสระ), ชื่อธนาคาร, รหัสผังบัญชี, โลโก้ (`system-setting-screens.ts:905-920`). ต้องปรับตาม Champ: (1) เลขที่บัญชีเป็นรหัสหลักตัวเดียว ตัดช่อง "รหัสสมุดบัญชี" แยก หรือยืนยันกับลุงจืดว่าจะคงรหัสสมุดแยก (2) เพิ่ม "ประเภทบัญชี" (master BCBankAccount) (3) ธนาคาร/สาขาเป็น lookup master (`BCBank`/`BCBankBranch`) ไม่ใช่ข้อความอิสระ (4) แสดง "ยอดคงเหลือ" อ่านอย่างเดียว (5) ตารางผู้ติดต่อในจอเดียว (6) ตัดโลโก้/รูปสมุดบัญชี (ไม่มีใน Champ — เสนอก่อนทำตามกฎ Strict Parity) (7) ลำดับแถว ตามข้อ 2.1 |
| 5172 | บันทึกนำฝากเงินสด | `cash-deposit` `/banking/cash-deposit` | ต่าง + รอพัฒนา | ข้อ 3.1: ช่อง "บันทึกยอดเงินยกมา", รหัสผังบัญชี + รหัสสมุดรายวัน, ลำดับฟิลด์ตามข้อ 3.1 |
| 5173 | บันทึกถอนเงินสด | `cash-withdraw` `/banking/cash-withdrawal` | ต่าง + รอพัฒนา | ข้อ 3.1 (ไม่มีช่องยกมา) |
| 5174 | บันทึกค่าใช้จ่ายธนาคาร | `bank-charges` `/banking/charges` | ต่าง + รอพัฒนา | ข้อ 3.1 |
| 5175 | บันทึกรายได้จากธนาคาร | `bank-income` `/banking/income` | ต่าง + รอพัฒนา | ข้อ 3.1 |
| 5176 | บันทึกโอนเงินระหว่างธนาคาร | `account-transfer` `/transaction/accounttransfer` | ต่าง + รอพัฒนา | ข้อ 3.2: 3 กลุ่ม (โอนจาก / โอนเข้า / รายละเอียดการโอน) ตามลำดับฟิลด์ |
| 5170 | คำนวณยอดคงเหลือของสมุดบัญชีใหม่ | `bank-recalculate` `/tools/bank-balances` | ต่าง (เครื่องมือทั่วไป) | ฟอร์ม Champ: จากสมุด / ถึงสมุด + Start/Cancel + แถบความคืบหน้า (ข้อ 2.5); ตรวจ `isErpToolApiReady` (ยังไม่ verify). ถ้า BC คำนวณยอดจากรายการตอนอ่านแล้ว เมนูนี้เหลือเป็น "ตรวจสอบยอด" — รอลุงจืดตัดสิน |
| 5149 | รายงานการฝากเงินสด | (ไม่มีเมนูแยก) | ยังไม่มี | ศูนย์รายงาน: 5 รายงาน (ฝาก/ถอน/รายได้/ค่าใช้จ่าย/โอน) ตามข้อ 6 |
| 5154 5158 5160 5163 | ถอน / รายได้ / ค่าใช้จ่าย / โอน | (ไม่มี) | ยังไม่มี | เช่นเดียวกัน |
| 5164 | รายงาน Bank Statement | `bank-statement-report` `/report/bankstatement` | ต่าง + ยังไม่มีข้อมูล | ทำตามข้อ 6.1 (ฝั่งสมุด) และดูส่วนกระทบยอดข้อ 7.3 |
| 5165 | รายงาน Statement ล่วงหน้า | (ไม่มี) | ยังไม่มี | ข้อ 6 (ต้องมีเช็ค/บัตรก่อน — ดู `05-cheque.md`) |
| 5166 5167 | รายงานรับ/จ่ายเงินอื่นๆ พร้อมรายการย่อย | `expense-summary-report` `/report/expensesummary` (รายงานการจ่ายเงินอื่นๆ เท่านั้น) | ต่าง | เพิ่มรายงานรับเงินอื่นๆ; ทำเลย์เอาต์หัว + รายการย่อยตามข้อ 6 |
| 5150 | รายงานเคลื่อนไหวเงินสดย่อย | `petty-cash-movement-report` `/report/pettycashmovement` | ต่าง + ยังไม่มีข้อมูล | ตามข้อ 6 |
| 5151 | รายงานการตั้งเบิกเงินสดย่อย | (ไม่มี) | ยังไม่มี | ตามข้อ 6 |
| 5152 | รายงานเคลื่อนไหวเงินสด | `cash-movement-report` `/report/cashmovement` | ต่าง + ยังไม่มีข้อมูล | ตามข้อ 6 |
| 5684 5685 | รายงานบันทึกขอเบิกเงินทดลองจ่าย 2 แบบ | (ไม่มี) | ยังไม่มี | ตามข้อ 6 |
| 5156 5157 | รายงานสมุดจ่ายเงินประจำเดือน เดบิต/เครดิต | `monthly-payment-book-report` `/report/monthlypaymentbook` (1 เมนูรวม) | ต่าง | แยกเป็น 2 รายงาน (เดบิต/เครดิต) ตามข้อ 6; ตรวจกับลุงจืดว่า Champ ต้นฉบับหาทางเข้าอย่างไร (ข้อสังเกตใต้ตาราง 6) |

หมายเหตุสถานะรายงานทุกรายการ: รหัสรายงานพร้อมใช้งานใน BC มี 12 รายการ ไม่มีรายงานเงินสด-ธนาคาร (`erp-reports.ts:530-543`) จึงแสดง "ยังไม่มีข้อมูล" (`menu-screen-status.ts` `isMenuDataPending`)

### 8.2 ความต่างที่สำคัญที่สุด

1. จอธุรกรรมเงินสด/ธนาคารทั้ง 11 ใน BC ยังเป็น "รอพัฒนา" และ API เดิมอยู่บน MongoDB ที่ถอดแล้ว (`erp-transaction.ts:217,377-816`); ต้องสร้างใหม่บน PostgreSQL ตามโครงข้อ 3-5
2. จอสมุดเงินฝากของ BC (`system-setting-screens.ts:893-925`) ต่างจาก Champ: ไม่มีประเภทบัญชี, ธนาคาร/สาขาไม่ใช่ lookup, ไม่มียอดคงเหลือ, ไม่มีผู้ติดต่อ, และมีรหัสสมุดแยกจากเลขที่บัญชี (Champ ใช้เลขที่บัญชีเป็นรหัสเดียว)
3. ยอดคงเหลือสมุดเงินฝากของ Champ ได้จากสูตรรวมเอกสาร 15 ประเภท (ข้อ 2.4) — BC ควรคำนวณจากรายการจริงเสมอ ไม่เก็บยอดสะสมแบบบวกลบ
4. Bank Statement (5164) ของ Champ คือรายการฝั่งสมุด (book) พร้อมยอดยกมา/คงเหลือ — BC ยังไม่มี; BC มีเฉพาะรายงาน Statement ที่ยังไม่จับคู่ (`subledger_queries.go:159-165`)
5. การกระทบยอดสองฝั่ง: ทั้ง Champ และ BC ยังไม่มี; BC มีโครงสร้างข้อมูล many-to-many พร้อม (`subledger.sql:37-58`) แต่ขาดรายงานฝั่งสมุดที่ยังไม่จับคู่ และสรุปกระทบยอด (ข้อ 7.3)
6. เอกสารรับ/จ่ายเงินอื่นๆ (5178/5179) ใน Champ คือ "ใบสำคัญย่อ" ที่มีกริดเดบิต/เครดิตและกล่องรับ/จ่ายเงิน — ต้องตัดสินใจร่วมกับระบบสมุดรายวัน BC (ชนิดใบสำคัญเป็น master ที่ผู้ใช้กำหนดได้ ตาม memory `journal-books-user-defined`)
7. รายงาน 16 รายการของ Champ ซ่อนอยู่ในศูนย์รายงาน; BC มี 8 เมนูรายงานในกลุ่มเดียว (`menu-data.ts:452-460`) ไม่ตรงกัน

## 9. ข้อขัดกับกฎเจ้าของ/ขอบเขต (ไม่เสนอทำเอง — รอลุงจืดตัดสิน)

- ขอบเขต GL ล่าสุด: ระบบห้องบัญชี เน้นรับข้อมูล ตรวจสอบ กระทบยอด ปิดงบ ไม่ใช่ระบบหน้าบ้านปฏิบัติงาน (AGENTS.md "คำสั่งล่าสุด"; `mydocs/specs/gl/ar-ap-bank-design-notes.md:49`) — เอกสารรับ/จ่ายเงินสด-ธนาคารของ Champ (5172-5176, 5178/5179) เป็นงานปฏิบัติการของระบบต้นทาง จึงต้องให้ลุงจืดตัดสินว่าทำเป็นจอคีย์ใน BC หรือรับเป็นใบสำคัญรายวันจากต้นทาง
- เงินทดรองจ่าย (5168): อยู่ในขอบเขต (ไม่ใช่เงินเดือน) ตาม AGENTS.md
- ห้ามเดา: ข้อความไทยของ resource key ที่ยังไม่ verify (ตารางรายงาน, MSG_888..891) ต้องให้ลุงจืดยืนยัน; ห้ามผูกเงื่อนไขกับรหัสบัญชี (memory `no-account-code-conditions`) — ใช้ `AccountCode` ของสมุดเงินฝาก
- เงิน decimal ไม่ใช่ float (กฎ 20): Champ ใช้ `MONEY` + `double`; `ProcessBankBook` คูณอัตราแลกเปลี่ยนด้วย double — BC ใช้ decimal/minor unit และกำหนดนโยบายปัดเศษ
- ฟิลด์เฉพาะลูกค้า (`IsSAP`) และ flag ปิดโมดูล GL (`LMD_REMOVE_GL`) ไม่นำมา

## 10. สิ่งที่ยังไม่ verify (สรุป)

| เรื่อง | ทำไมยัง |
|---|---|
| ข้อความไทยของ resource key (หัวคอลัมน์รายงาน, MSG_888..891, CANNOT_*) | ฐาน resource ไม่อยู่ใน repo |
| ยอดรวมท้ายรายงาน 5149-5163, 5150-5152, 5684/5685 | ยังไม่อ่านส่วนสรุปของ view ทุกตัว |
| ทางเข้ารายงาน 5156/5157 | ไม่พบ `ON_COMMAND` ของ id 5156/5157 |
| การเติมค่า `ChargeAmount/TotalAmount` ของ BCBankTransfer | ไดอะล็อกไม่มีช่องคีย์ |
| จำนวนตารางที่ปุ่ม "เปลี่ยนเลขที่บัญชี" แก้ (ประมาณ 24) | นับจากโค้ดคร่าว ๆ |
| รายละเอียดกล่องรับ/จ่ายเงินกลาง (`OnReceive2`) | อยู่ในเอกสารระบบอื่น |
| `TransBankDate` มีที่หน้าจอรับ/จ่ายใด | ยังไม่ไล่ฟอร์มใบเสร็จ |
| ผลของ BC `isErpToolApiReady("bank_balances")` | ยังไม่รันตรวจ |
