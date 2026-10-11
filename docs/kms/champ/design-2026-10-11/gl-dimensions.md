# D2 — มิติวิเคราะห์ GL (แผนก/โครงการ/การจัดสรร/งาน/หน่วยงาน/ฝ่าย/สาขา) ให้ตรง Champ 5.5

สถานะ: ออกแบบ (2026-10-11) · audit id `D2` · เจ้าของการตัดสินใจ: ลุงจืด
ขอบเขต: GL อย่างเดียวก่อน ตามการตัดสินใจ 2026-10-11 ("ทำรายการบัญชีก่อน, สินค้าทีหลัง") ส่วน ERP ทำใน W8 (ระยะหลัง)
แหล่งอ้างอิงทางการ: **ไม่จำเป็น** เพราะมิติวิเคราะห์เป็นข้อมูลบริหารภายใน ไม่มีข้อเท็จจริงด้านภาษีหรือกฎหมายไทย ไม่มีอัตรา/มาตรา/แบบฟอร์มให้อ้าง
ไฟล์ Champ ที่อ่าน: `D:/project-champ` — `SQLSERVER_Script.sql` (CP874/TIS-620 ไม่ใช่ UTF-16 — ถอดด้วย Node `TextDecoder('windows-874')`; เลขบรรทัดในเอกสารนี้นับจากไฟล์ที่ถอดแล้ว), `.cpp` (CP874), `menuconfig.xml` (UTF-16LE), ภาพหน้าจอใน `bcspec/` (ถ้าขัดกับ `.rc` ให้ถือภาพหน้าจอ)

---

## 1. Champ ทำงานอย่างไร (หลักฐาน)

### 1.1 จำนวนมิติ = 7 (ไม่นับสาขาเหลือ 6)
`BCTransSub` (`SQLSERVER_Script.sql:9430-9475`, index `:9465-9475`) มีคอลัมน์มิติรายบรรทัดดังนี้

| # | คอลัมน์ Champ | ขนาดในบรรทัด | ทะเบียน (master) | ขนาดของ master (code/name) | ชื่อที่ใช้ในเมนู Champ |
|---|---|---|---|---|---|
| 1 | DepartCode | VARCHAR(10) | BCDepartment `:460` | 10 / 50 | แผนก |
| 2 | ProjectCode | VARCHAR(10) | BCProject `:590` | 10 / 50 | โครงการ |
| 3 | AllocateCode | VARCHAR(10) | BCAllocate `:370` | 10 / 50 | การจัดสรร |
| 4 | JobCode | VARCHAR(20) | BCJob `:9177` | 15 / 255 | งาน |
| 5 | BranchCode | VARCHAR(15) | BCBranch `:9192` | 15 / 255 | สาขาของบริษัท |
| 6 | PartCode | VARCHAR(10) | BCPart `:9148` | 15 / 255 | หน่วยงาน |
| 7 | SideCode | VARCHAR(10) | BCSide `:9162` | 15 / 255 | ฝ่าย |

- ความยาวที่ใช้ได้จริง = ค่าที่น้อยที่สุดระหว่าง master, บรรทัด และตารางยอดคงเหลือ/งบประมาณ: Depart/Project/Allocate/Part/Side = 10, Job = 15 (master 15, บรรทัด/ยอด 20), **Branch = 10** (บรรทัด `BCTransSub` 15 แต่ `BCPeriodBalance :9321`, `BCYearPeriodBal :9356`, `BCGLBudget :9213` เป็น VARCHAR(10))
- ทุก master มี `UserGroup VARCHAR(255)` และคอลัมน์ audit (`CreatorCode`, `CreateDateTime`, `LastEditorCode`, `LastEditDateT`)
- ที่เก็บรหัส: **Part/Side มีเฉพาะในตาราง GL** ได้แก่ `BCTransSub(_tmp/_exp)`, `BCPeriodBalance :9317`, `BCYearPeriodBal :9352` และ `BCGLBudget :9207`
  - **Job** มีในตาราง GL ชุดเดียวกัน **และ**บนหัวเอกสาร ERP 51 ตารางเป็น `JobNo VARCHAR(20), -- Job สำหรับ GL` (เช่น `:1361, :1479, :1719`) — แก้จากฉบับก่อนที่เขียนว่า Job ไม่มีใน ERP
- **Depart/Project/Allocate** อยู่บนหัวและบรรทัดเอกสาร ERP ประมาณ 80 ตาราง (เช่น BCAPInvoice(Sub), BCPurchaseOrder(Sub), BCSaleOrder(Sub), BCReceipt*, BCPayment*, BCStkIssue*, BCCashBankIn/Out, BCOtherExpense)

### 1.2 สวิตช์ระดับบริษัท ("ใช้ผังบัญชีแบบแยก…")
- คอลัมน์ `DepartSeparate`, `ProjectSeparate`, `AllocateSeparate`, `JobSeparate`, `PartSeparate`, `SideSeparate`, `BranchSeparate` (ทั้งหมด `SMALLINT DEFAULT 0`) มีอยู่ใน 3 ตาราง
  - `BCConfigurations` `:743-750`
  - `BCConfigurations2` `:947-954`
  - `BCChartOfAccount` `:9069-9075` ซึ่งเป็นธงรายบัญชี
- จอสมุดรายวันอ่านค่าจาก `BCConfigurations` (`glfrmjournal.cpp:1569-1599`) โดย Job/Part/Side/Branch อยู่ใต้ `#ifndef EXPRESS_PROJ`
- จอกำหนดค่าบริษัทอยู่ที่ `myfrmsysconfig.cpp:927-946` และบันทึกที่ `:1176-1182`
- ภาพ `bcspec/.../กำหนดค่าเริ่มต้นบริษัท -Option 2.jpg` แสดงคำตามตัวอักษรดังนี้
  - 7) ใช้คำอธิบายในรายการย่อย GL
  - 8) ใช้ผังบัญชีแบบแยกแผนก
  - 9) ใช้ผังบัญชีแบบแยกโครงการ
  - 10) ใช้ผังบัญชีแบบแยกการจัดสรร
  - 11) ใช้ผังบัญชีแบบแยก JOB
  - 12) ใช้ผังบัญชีแบบแยกส่วนงาน (Part)
  - 13) ใช้ผังบัญชีแบบแยก Site
  - 14) ใช้ผังบัญชีแบบแยกสาขา
- ค่าเริ่มต้นตาม DDL คือ 0 (ปิด) ส่วนในภาพตัวอย่างลูกค้ารายหนึ่งเปิดข้อ 8

### 1.3 จอบันทึกรายวัน
- ลำดับคอลัมน์ (`glfrmjournal.cpp:227-256`): รหัสบัญชี, ชื่อบัญชี, [รายละเอียดเพิ่มเติม], **แผนก, โครงการ, การจัดสรร, งาน, สาขา, หน่วยงาน, ฝ่าย**, เดบิต, เครดิต
- คอลัมน์มิติแสดงก็ต่อเมื่อสวิตช์ของบริษัท = 1 ภาพ `บันทึกข้อมูลรายวัน.jpg` ยืนยันว่าเมื่อเปิดเฉพาะแผนกจะเห็นแค่คอลัมน์ "แผนก"
- ตัวเลือกรหัสเป็น combo `SELECT Code,Name FROM BCxxx` (`:272-299`) ไม่กรองด้วย UserGroup
- การตรวจตอนบันทึก (`OnCheckSeparate :639-686`, ข้อความ error `:584-589`): ถ้ากรอกรหัสในบัญชีที่ `XSeparate=0` จะปฏิเสธด้วย "Chart_Of_Account <code> NOT_SEPARATE" และ**ไม่บังคับกรอก**รหัสมิติ
  - ตรวจ**เฉพาะใบใหม่** (`if(m_bInsertDataMode)` `:574-590`) และอ่านค่าเฉพาะมิติที่สวิตช์บริษัทเปิด (`:575-581`)
  - ตอนเขียนบรรทัด (`:890-903` แล้ว `InsertDB :922` / `UpdateDB :927`) ใส่คอลัมน์มิติเฉพาะที่สวิตช์เปิด → สวิตช์ปิด = ไม่แตะค่าเดิมในบรรทัด
  - ระหว่างกรอกช่อง (`OnCheckComboData :2812-2923` เรียกจาก `:1492`): บัญชีปิดธง หรือรหัสไม่อยู่ในรายการที่อนุญาต → **ล้างช่องเงียบ ๆ** ไม่มีข้อความ (ใบแก้ไขจึงถูกกันที่ช่องกรอก ไม่ใช่ตอนบันทึก)
  - ใบที่แก้ไข: ห้ามบันทึกเมื่อยกเลิกแล้ว `NOT_SAVE_ISCANCEL :510-516`, ผ่านรายการแล้ว `CANNOT_SAVE_POSTOK :517-523`, มาจากระบบอื่น (`Source<>0`) `JONAL_MESS_NOT_SAVE :524-529`
- รหัสที่อนุญาตต่อบัญชีเก็บใน `BCAccSeparate` (`:9563-9576`) เป็นแถว (SeparateType, AccountCode, SeparateCode)
  - SeparateType: 0 แผนก, 1 โครงการ, 2 การจัดสรร, 3 ฝ่าย, 4 หน่วยงาน, 5 สาขา, 6 งาน
  - SeparateCode เป็นรายการคั่นด้วยจุลภาค หรือ `_ALL`
  - `glfrmjournal.cpp:2815-2923`: `_ALL` = ทุกรหัส; ระบุรายการ = จำกัดเฉพาะรายการนั้น; **เปิดธงแต่ไม่มีแถว = ไม่อนุญาตรหัสใดเลย** (ค่าเริ่มต้น "X9X9X9X9X9")

### 1.4 ผังบัญชี
- `GLChartOfAccPage1DetailDlg.cpp:134-145, :225-278`: แต่ละมิติมี checkbox แยก, ตัวเลือก "SELECT_ALL" (`_ALL`) และปุ่มเลือกหลายรหัส
- control จะถูก disable เมื่อสวิตช์บริษัทของมิตินั้นปิด (`:250-278`)
- ตอนบันทึกจะลบแล้ว insert `BCAccSeparate` ใหม่ (`GLFrmChartofAccountPage1.cpp:721-755`)

### 1.5 จอ "กำหนดรหัสต่างๆ"
- เป็นจอ generic ใน `myfrmbasemaster.cpp:214-274` มี combo เลือกประเภทรหัสราว 35 แบบ ส่วนที่เป็น GL ได้แก่
  - กำหนดรหัสการจัดสรร `:219`
  - กำหนดรหัสโครงการ `:223`
  - กำหนดรหัสงาน `:224`
  - กำหนดรหัสแผนก `:242`
  - กำหนดรหัสฝ่าย `:243`
  - กำหนดรหัสสาขาของบริษัท `:247`
  - กำหนดรหัสหน่วยงาน `:248`
- ภาพ `กำหนดรหัสต่างๆ.jpg`: grid มีคอลัมน์ รหัส, ชื่อ, กลุ่มพนักงาน/ใช้กับรายวัน และปุ่ม Import/Export, F5 ตัวอย่าง, F11 บันทึก, พิมพ์, ปิด
- ขั้นตอนบันทึก (F11, `OnMasterSave :393-560`): ตรวจสิทธิ์อ่านอย่างเดียว `:401-404` → วนทุกแถวของ grid `:509-536`: แถวใหม่ที่มีรหัส+ชื่อ = INSERT; ลบรหัสในแถวเดิมจนว่าง = DELETE ตาม RowOrder; อื่น ๆ = UPDATE รหัส/ชื่อ/UserGroup → ข้อความ `SAVE_COMPLETED :557`
  - **ไม่มีการตรวจว่ารหัสถูกใช้แล้ว** ก่อนลบหรือเปลี่ยนรหัส
- Import (`:743-864`): อ่านไฟล์คั่นด้วย tab (`CODE<TAB>NAME`) ใส่ grid เท่านั้น ยังไม่บันทึกจนกด F11
- Export (`:685-742`): เขียนไฟล์ `CODE<TAB>NAME` ทุกแถว
- F5 ตัวอย่าง (`OnExamples :1116-1135`): ขอรหัสผ่าน → ยืนยันด้วยข้อความ `USE_DATAEXAMPLE_DELETEOLEDTA` ("ต้องลบข้อมูลเดิมก่อน") → ลบของเดิมแล้วคัดจากฐานข้อมูลตัวอย่าง

### 1.6 รายงาน
- **รายงานทะเบียน 7 ตัว** (`menuconfig.xml:575-585`, `MainFrm.cpp:7777-7815`)

  | เมนู Champ | resource | ชื่อรายงาน |
  |---|---|---|
  | 5538 | 104009 | รายงานรายละเอียดแผนก |
  | 5521 | 104010 | โครงการ |
  | 5524 | 104011 | การจัดสรร |
  | 5526 | 104013 | งาน |
  | 5527 | 104014 | หน่วยงาน |
  | 5528 | 104015 | ฝ่าย |
  | 5529 | 104016 | สาขา |

  มี PDF อยู่ใน `bcspec/Report_ระบบบัญชี/`
- **รายงานที่กรองหรือแยกยอดตามมิติ** (`champ-report/BC5REP_GLFS`): งบทดลอง (`GLRepTrialBalanceDlg` 5506, `GLRepTrialBalance2Dlg`), บัญชีแยกประเภท (`GLRepGLDlg`, `GLRepGL2Dlg` 5534), สรุปแยกประเภท (`GLRepGLSumDlg` 5535, `GLRepGLSummerizeDlg`), รายการตามรหัส (`GLRepGLTransByCodeDlg`, `GLRepGLTransDlg`), กระดาษทำการ (`GLRepWorkingPaperDlg` 5499) และงบประมาณ (`GLRepBudgetDlg`, `GLRepBudgetCompDlg`)
- กลไกตาม `GLRepTrialBalanceDlg.cpp:163-188, :694-715`
  - ตัวกรอง**แสดงเสมอ แต่ถูก disable** (`EnableWindow(FALSE)` `:160-190`) เมื่อสวิตช์บริษัทของมิตินั้นปิด
  - งบการเงิน (`MyXpFinanceSheet.cpp:103-111, :409-474, :707-717`) ไม่มีตัวกรองมิติระดับรายงาน แต่สูตรในแต่ละบรรทัดของรูปแบบงบระบุมิติได้ (`ACC,PERIOD,DEP,ALLOCATE,PROJ,SIDE,JOB,BRANCH,PART`) → เป็นงานของการออกแบบรูปแบบงบการเงิน ไม่ใช่เอกสารนี้
  - เลือกหลายรหัสได้ → `DepartCode IN (...)`
  - checkbox `m_iDepart` จะเพิ่มมิติเข้า `GROUP BY/ORDER BY`
  - สรุปไว้แล้วที่ `docs/kms/champ/09-gl.md:493-547`

### 1.7 ผ่านรายการ/ยอดคงเหลือ
- `GLFrmPostTrans.cpp:1249-1266`: ยอดงวดจัดกลุ่มตามบัญชีบวกมิติที่เปิด
- ในจอทบทวนก่อนผ่านรายการ (`:1042-1062`) แก้ได้เฉพาะ แผนก/โครงการ/การจัดสรร ของบรรทัดต้นทาง
- ผ่านรายการ (`OnBnProcess :305-313` เลือก `BCTrans` ที่ `IsConfirm<>1 AND IsCancel=0` → `OnCheckTransSub :722-756` ตรวจเดบิต=เครดิต (iError 1/2) → `OnPostGL :757-800` ใส่ BatchNo + `IsConfirm=1` ทั้งหัวและบรรทัด) **ไม่ตรวจมิติซ้ำ**
- จอทบทวน (`:1040-1072`) UPDATE บัญชี/คำอธิบาย/แผนก/โครงการ/การจัดสรร/เดบิต/เครดิต ตรงลง `BCTransSub` โดยไม่เรียก OnCheckSeparate
- คำนวณยอดงวด (`OnPeriodProcess :1228-1246`, `OnProcess :1681-1700`): ลบ `BCPeriodBalance` ของงวด → `OnSetPeriodLink :1249-1266` จัดกลุ่มบัญชี+มิติที่เปิด → `OnProcessLink :1373-1489` INSERT; สวิตช์อ่านที่ `OnInitialPeriod :1346-1370` (สาขา/งาน/หน่วยงาน/ฝ่าย อยู่ใน `#ifndef EXPRESS_PROJ`)
- ยกเลิกผ่านรายการ (`GLDlgCancelPost.cpp`): อ่านสวิตช์ `:207-229` → เลือกบรรทัดที่ผ่านแล้ว `:349` → **หักยอด** `BCPeriodBalance` ตามบัญชี+มิติที่เปิด `:377-417` (`Debit = Debit + (-x)`); ข้อความ `MESS_CELPOST_NOTVALID :319`
- ปิดปี (`YearEndProcess_Page1.cpp`): อ่านสวิตช์ `:245` → อ่านมิติต่อแถวยอด `:305-343` → คัด `BCPeriodBalance` → `BCYearPeriodBal` พร้อมทุกมิติ `:566` → INSERT `BCTrans :619` + บรรทัดยอดยกมาพร้อมมิติที่เปิด `:655-707`
- **ERP → GL** (`champ-lib/BC5PROCESS/gltlsprocess.cpp`, trace แล้ว): อ่าน**หัวเอกสาร** (เช่น `select bcapinvoice.* :692/:730`)
  - แผนก/โครงการ/การจัดสรร มาจากหัว (`:290-292, :879-881`)
  - งาน = `JobNo` ของหัว เมื่อสวิตช์บริษัท `JobSeparate == 1` (`:293-295, :883-885`)
  - `IsSeparate(บัญชี, ...)` `:463-464` (โค้ด `:9354-9381`) อ่านธงของบัญชี → ตอน INSERT บรรทัด `:497-500`: แผนก/โครงการ/การจัดสรร ใส่เมื่อธงบัญชีเปิด ไม่เช่นนั้นว่าง; งานใส่เสมอ (ไม่ดูธงบัญชี)
  - ไม่มีการตรวจสวิตช์บริษัท/รายการรหัสที่อนุญาตสำหรับ แผนก/โครงการ/การจัดสรร; บรรทัดปัดเศษ (`:424-427, :447`) คัดมิติหัวโดยไม่ดูธง
  - หน่วยงาน/ฝ่าย ไม่มาจาก ERP
  - DDL มีคอลัมน์ `JobNo VARCHAR(20), -- Job สำหรับ GL` บนหัวเอกสาร ERP 51 ตาราง (เช่น `:1361, :1479, :1719`)

### 1.8 ขั้นตอนประมวลผล: Champ (ต้นแบบ) เทียบ BC
กฎ 2026-10-11 "ประมวลผล/workflow ยึด Champ เป็นต้นแบบเสมอ" → ทุกขั้นด้านล่างทำตาม Champ ยกเว้นช่อง "BC ต่างโดยตั้งใจ" ที่ระบุเหตุผล

| ขั้นตอน | Champ: ลำดับ / สถานะ / ตรวจ+ข้อความ / ตารางที่แตะ / การกลับรายการ | BC ต่างโดยตั้งใจ (เหตุผล) |
|---|---|---|
| บันทึกทะเบียนรหัส | `myfrmbasemaster.cpp:393-560` สิทธิ์ `:401-404` → วนแถว INSERT/DELETE/UPDATE `:509-536` → `SAVE_COMPLETED :557`; ไม่ตรวจการใช้งาน | ห้ามลบ/เปลี่ยนรหัสที่ถูกใช้ (`dimensionInUse`) = ส่วนเสริมเล็กที่มีอยู่แล้ว (§7) |
| บันทึกสวิตช์บริษัท | `myfrmsysconfig.cpp:927-946` (แถว 8-14), บันทึก `:1176-1183`; ไม่ตรวจข้อมูลเดิม | ใช้ version กันเขียนทับ (sync ACID) |
| บันทึกธงบัญชี | `GLFrmChartofAccountPage1.cpp:721-755` ลบแล้ว INSERT `BCAccSeparate` (`_ALL` เมื่อเลือกทั้งหมด); control disable เมื่อสวิตช์ปิด `GLChartOfAccPage1DetailDlg.cpp:134-217, :250-278` | เก็บเป็นคอลัมน์ในบัญชีแทนตารางแยก (ความหมายเท่ากัน) |
| สร้างรายวัน | `glfrmjournal.cpp:572-591` อ่านเฉพาะมิติที่สวิตช์เปิด → `OnCheckSeparate :639-686` → `NOT_SEPARATE :585-588` → INSERT `:922` → `AddUserLogBook :970-971` | ตรวจรหัสที่อนุญาตที่ backend ตอนบันทึก (Champ ตรวจที่ช่องกรอก `:2812-2923`) เพราะ API รับข้อมูลได้โดยไม่ผ่านจอ |
| แก้รายวัน | `:510-529` ห้ามเมื่อยกเลิก/ผ่านแล้ว/มาจากระบบอื่น → ไม่เรียก OnCheckSeparate → UPDATE `:927` เฉพาะมิติที่สวิตช์เปิด `:890-903` | ตรวจกฎมิติเหมือนตอนสร้าง (เหตุผลเดียวกับข้างบน); สวิตช์ปิด = คงค่าเดิมเหมือน Champ |
| ผ่านรายการ | `GLFrmPostTrans.cpp:305-313` → `:722-756` เดบิต=เครดิต → `:757-800` `IsConfirm=1`; ไม่ตรวจมิติ | เหมือน Champ: ไม่ตรวจมิติซ้ำ (`checkSeparate=false`) |
| ทบทวนก่อนผ่าน | `:1040-1072` แก้ แผนก/โครงการ/การจัดสรร ตรงลงบรรทัด ไม่ตรวจ | ไม่พบจอนี้ใน BC (grep `review/ทบทวน` ใน `generalledger`, `app/gl` ไม่พบ) → แก้ใบร่างผ่านการแก้รายวันปกติ (ตรวจกฎมิติ); ถ้าจะทำจอนี้ ให้แยกเป็นงาน parity ของการผ่านรายการ |
| ยกเลิกผ่านรายการ | `GLDlgCancelPost.cpp:207-245` → `:349` → หักยอด `BCPeriodBalance :377-417`; `MESS_CELPOST_NOTVALID :319` | BC ใช้รายการกลับบัญชี (ADR เดิม: ผ่านแล้วห้ามแก้เงียบ) → รายการกลับต้องคัดครบ 6 มิติ ไม่ตรวจกฎมิติ |
| คำนวณยอดงวด | `:1228-1246`/`:1681-1700` ลบแล้วสร้างใหม่ จัดกลุ่มบัญชี+มิติที่เปิด `:1249-1266, :1373-1489` | BC คำนวณจาก `gl_lines` จัดกลุ่มทุกมิติเสมอ (ไม่ขึ้นกับสวิตช์) เพราะไม่มีตารางยอดที่ต้องสร้างใหม่; รายงานจะรวมเองเมื่อไม่เลือกแยก |
| ปิดปี/ยอดยกมา | `YearEndProcess_Page1.cpp:245, :305-343, :566, :619, :655-707` ยกยอดพร้อมมิติที่เปิด | key ยอดยกมาขยายเป็นทุกมิติ; ไม่ตรวจกฎมิติ (`checkSeparate=false`) |
| ERP → GL | `gltlsprocess.cpp:290-295, :463-464, :497-500, :9354-9381` (§1.7) | BC บันทึกแบบ sync ACID ใน tx เดียวกับเอกสาร; กฎคัดมิติเหมือน Champ |
| งบประมาณ | `GLFrmBudget.cpp:138-151` แสดงคอลัมน์มิติตามสวิตช์ → บันทึก `ConfrimSaveData :421-480` (รหัสซ้ำ `MESS_ERROR_DUPCODE`/`CANNOT_SAVE_CODE_DUP`, ธงบัญชี `NOT_SEPARATE :471`); `BCGLBudget :9207-9230` มีครบ 7 มิติ | เหมือน Champ (§3.11) |

---

## 2. สถานะ BC ปัจจุบัน

| เรื่อง | BC มีแล้ว | หลักฐาน |
|---|---|---|
| ทะเบียน | แผนก (`departments`) และโครงการ (`projects`) เก็บใน `gl_records` ผ่าน Master CRUD แบบ generic | `generalledger/dimensions.go:1-60`, `models.go:206-216` |
| registry | map 2 รายการ (label, column, jsonKey, accountListKey); ความยาวตายตัว 10/50 | `dimensions.go` const + `dimensionKinds` |
| บรรทัด GL | `gl_lines.branch_code, department_code, project_code text NOT NULL`; index `gl_lines_dimension_idx` | `schema.sql` (embed ที่ `postgres.go:17`) |
| งบประมาณ | `gl_budgets.department_code/project_code` | `budget.sql:12-13` |
| ธงบัญชี | `DepartSeparate/ProjectSeparate` + `DepartCodes/ProjectCodes`; **รายการว่าง = ทุกรหัส** (ต่างจาก Champ) | `models.go:50-58` |
| ตรวจบรรทัด | `checkLineSeparation` คืนข้อความไทยผ่าน `fmt.Errorf` → code `validation_failed` | `models.go:517-538` |
| ตรวจรหัส | `loadLineDimensions` และ `checkAccountDimensionCodes` เขียนตายตัวไว้ 2 มิติ | `dimensions.go:212-274` |
| ลบ/แก้รหัส | `dimensionInUse` ตรวจ `gl_lines`, `gl_budgets`, journals, templates, allocations, forecast และรายการของบัญชี | `dimensions.go:128-165` |
| projection | `projectedLine`, `insertProjectedLines`, `verifyProjectedJournal` ใช้ 2 มิติ | `postgres_lines.go:44-90` |
| ยอดปิดงวด | จัดกลุ่มตาม branch/department/project | `postgres_processes.go:225`, `reports_process.go:63-66` |
| ขอบเขตการตรวจมิติ | `validateJournalLines(..., requireActiveBook, checkSeparate)`; `checkSeparate=true` เฉพาะสร้าง `:587` / แก้ `:681`; ผ่านรายการ `:708` และกลับรายการ `:792` = `(false,false)`; ปิดปี `postgres_processes.go:215` = `(true,false)`; เหตุผลใน `models.go:512-516` (เปลี่ยนผังทีหลังต้องไม่ล็อกเอกสารเก่า) | `postgres_store.go:820-848` |
| กลับรายการ | บรรทัดกลับคัดแค่ `DepartmentCode/ProjectCode` | `postgres_store.go:752-765` |
| ปิดปี/ยอดยกมา | key มิติ = {branch, department, project} | `postgres_processes.go:143-187` |
| งบประมาณ (Go) | struct มีแค่ Branch/Department/ProjectCode; ตรวจ `checkCode :225`; INSERT `:296`; UPDATE `:335`; `budgetHeaderColumns :406`; `gl_budget_lines.amount numeric(18,2)` | `budgets.go:48-56`, `budget.sql:30` |
| งบการเงิน | กรอง branch/department/project อยู่แล้ว (ส่วนเสริมเดิมของ BC) | `statements.go:165-172` |
| รายงาน | กรอง TB/ledger ด้วย `$7/$8` และ breakdown `ByDepartment/ByProject` | `reports.go:52, :218-220, :373-383`; งบประมาณ `:861, :942, :957` |
| รายงานทะเบียน | รองรับ `departments`/`projects` | `reports.go:79, :123, :1110-1141` |
| HTTP | `POST /gl/v2/command`, `GET /gl/v2/:resource(/:id)`, `GET /gl/v2/reports/:report`; ทั้ง 2 มิติ map ไปจอ `gl-dimensions`; ผู้ใช้ ledger ทุกคนอ่านได้ | `httpapi/http.go:69-72, :94, :139, :158` |
| สาขา | ระดับหัวรายการ (`gl_journals.branch_code` ใช้กับทุกบรรทัด) ไม่มีทะเบียน BCBranch แยก | `postgres_lines.go` |
| ERP → GL | `mappingLines` ไม่ส่งมิติ | `transaction/httpapi/posting.go:170-195` (อีก workflow กำลังแก้อยู่) |
| UI ทะเบียน | 2 tab | `frontend/src/app/gl/gl-dimensions.tsx` (123 บรรทัด) |
| UI รายวัน | คอลัมน์แผนก/โครงการแสดงเสมอ ล็อกตามธงบัญชี | `gl-journals.tsx:213-214, :915-921, :1196-1221` |
| UI ผังบัญชี | ธง+รายการ 2 มิติ | `gl-masters.tsx:128-131, :161-162, :503-504, :1403-1404` |
| เมนู | `tx("gl-dimensions","กำหนดรหัสต่างๆ (แผนก/โครงการ)",...,"/gl/dimensions")` | `frontend/src/lib/menu-data.ts:632`, alias `:95` |
| BFF | allowlist query `departmentcode, projectcode, bydepartment, byproject, dimtotals` | `frontend/src/app/api/gl/[...glPath]/route.ts:27` |
| audit | ทุก command เขียน `gl_events` ใน tx เดียวกัน | `postgres_store.go:220` |

**ช่องว่าง**
- ไม่มีมิติ การจัดสรร/งาน/หน่วยงาน/ฝ่าย
- ไม่มีสวิตช์บริษัท จึงแสดงคอลัมน์มิติเสมอ
- error ยังเป็นข้อความไทยฝังโค้ด
- ไม่มีรายงานทะเบียน 5 ตัว
- ERP ไม่ส่งมิติไป GL
- `docs/kms/champ/09-gl.md:49-55` ตั้งชื่อเมนู 5527/5528 ผิดเป็น "ส่วนงาน/ด้าน" (ที่ถูกตามเมนู Champ คือ หน่วยงาน/ฝ่าย)

---

## 3. การออกแบบ

### 3.1 ชื่อ (เลี่ยงชนกับของเดิม)
- ชนที่ 1: `allocations` = `gl_allocations` ซึ่งเป็นการปันส่วนต้นทุน (`models.go:209`, `journal_books.go:300`)
- ชนที่ 2: `side` ถูกใช้แล้วใน `MappingRule.Side` (เดบิต/เครดิต)
- ชื่อใหม่ทั้งหมดเติม `-codes` และใช้ชื่อเดียวกันทุกชั้น

| มิติ | resource/kind | คอลัมน์ PG | JSON บรรทัด | ธงบัญชี | รายการรหัสในบัญชี | code/name max |
|---|---|---|---|---|---|---|
| แผนก (มีแล้ว) | `departments` | `department_code` | `departmentcode` | `departseparate` | `departcodes` | 10/50 |
| โครงการ (มีแล้ว) | `projects` | `project_code` | `projectcode` | `projectseparate` | `projectcodes` | 10/50 |
| การจัดสรร | `allocate-codes` | `allocate_code` | `allocatecode` | `allocateseparate` | `allocatecodes` | 10/50 |
| งาน | `job-codes` | `job_code` | `jobcode` | `jobseparate` | `jobcodes` | 15/255 |
| หน่วยงาน | `part-codes` | `part_code` | `partcode` | `partseparate` | `partcodes` | 10/255 |
| ฝ่าย | `side-codes` | `side_code` | `sidecode` | `sideseparate` | `sidecodes` | 10/255 |

- หมายเหตุ: audit E4 เสนอ `allocation_code` บนเอกสาร ERP → ให้ใช้ `allocate_code` ชื่อเดียวกัน

### 3.2 Registry (`dimensions.go`)
- ขยาย `dimensionKind` เป็น `{column, jsonKey, accountFlagKey, accountListKey, errKey, labelKey, reportTitleKey, codeMax, nameMax, order}`
- `errKey` เป็นชื่อสั้น: `depart`, `project`, `allocate`, `job`, `part`, `side`
- `order` ใช้ลำดับ Champ: depart 1, project 2, allocate 3, job 4, part 6, side 7
- `loadLineDimensions`, `checkAccountDimensionCodes`, `checkLineSeparation`, `dimensionInUse` และ `validateDimensionMaster` ทั้งหมดวน loop ตาม registry
  - ลบ `if kind == DepartmentKind` และ `fmt.Errorf` ไทยออก (ไม่ผูกตรรกะกับรหัสบัญชี ใช้ธงจาก master บัญชีเท่านั้น)
- ข้อความไทยใน `UserError.Message` คงไว้เป็น fallback ตาม pattern เดิมของ `fieldError` (`errors.go:80-90`) ส่วนจอแปลผ่าน `gl_err_<code>`
- `MasterCollections` เพิ่ม 4 kind
  - collection name ใช้เพื่อ log/grouping เท่านั้น ข้อมูลอยู่ใน `gl_records` และ `recordKind` (`postgres.go:160`) รับ kind เหล่านี้อัตโนมัติ
- ทะเบียนเก็บ `usergroup` (text, ไม่บังคับ) ใน payload และแสดงในคอลัมน์ "กลุ่มพนักงาน/ใช้กับรายวัน" ตาม Champ (`myfrmbasemaster.cpp:509-536` UPDATE UserGroup) — **ไม่ใช้กรองตัวเลือก** เพราะ combo ของรายวัน Champ ไม่กรอง (`glfrmjournal.cpp:272-299`)

### 3.3 Schema (ensureSchema ตาม pattern เดิม ไม่มี migration)
```sql
-- schema.sql (gl_lines)
ALTER TABLE gl_lines ADD COLUMN IF NOT EXISTS allocate_code text NOT NULL DEFAULT '';
ALTER TABLE gl_lines ADD COLUMN IF NOT EXISTS job_code      text NOT NULL DEFAULT '';
ALTER TABLE gl_lines ADD COLUMN IF NOT EXISTS part_code     text NOT NULL DEFAULT '';
ALTER TABLE gl_lines ADD COLUMN IF NOT EXISTS side_code     text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS gl_lines_dimension2_idx
  ON gl_lines(company, fiscal_year, allocate_code, job_code, part_code, side_code, entry_date);

-- budget.sql (gl_budgets) — งบประมาณมีครบมิติตาม Champ BCGLBudget :9207-9230 (สเปกเต็ม §3.11)
ALTER TABLE gl_budgets ADD COLUMN IF NOT EXISTS allocate_code text NOT NULL DEFAULT '';
ALTER TABLE gl_budgets ADD COLUMN IF NOT EXISTS job_code      text NOT NULL DEFAULT '';
ALTER TABLE gl_budgets ADD COLUMN IF NOT EXISTS part_code     text NOT NULL DEFAULT '';
ALTER TABLE gl_budgets ADD COLUMN IF NOT EXISTS side_code     text NOT NULL DEFAULT '';
```
- ทะเบียนใหม่ไม่ต้องมีตารางใหม่ ใช้ `gl_records(company, kind, id, code, version, payload jsonb)` แบบเดียวกับ departments
- `gl_budgets` มี PK `(company, code)` (`budget.sql:21`) ไม่รวมมิติ จึงไม่ต้องแก้ key
- เงินไม่เปลี่ยน (Rule 20): `gl_lines` debit/credit = `numeric(38,8)` (`schema.sql:24`), `gl_budget_lines.amount` = `numeric(18,2)` (`budget.sql:30`) — ไม่ขยาย ไม่แตะคอลัมน์เงินในงานนี้
- Go
  - `Line` และ `projectedLine` เพิ่ม `AllocateCode/JobCode/PartCode/SideCode`
  - `insertProjectedLines` (jsonb_to_recordset) และ `verifyProjectedJournal` เปรียบเทียบครบ 6 มิติ
  - `processBalancesTx`, `reports_process`, `close_preview` group เพิ่ม 4 คอลัมน์
  - รายการกลับ (`postgres_store.go:752-765`) คัดครบ 6 มิติ (ตอนนี้คัดแค่ 2) — เทียบ Champ ยกเลิกผ่านรายการหักยอดตามมิติ `GLDlgCancelPost.cpp:377-417`
  - ปิดปี/ยอดยกมา (`postgres_processes.go:143-187`) key ขยายเป็น {branch + 6 มิติ} ตาม Champ `YearEndProcess_Page1.cpp:305-343, :655-707`
  - บัญชี: เพิ่ม `<k>allcodes` (bool "เลือกทั้งหมด") ต่อมิติ ตาม Champ `_ALL` (`GLFrmChartofAccountPage1.cpp:721-755`)

### 3.4 สวิตช์บริษัท `dimension-settings`
- เก็บใน `gl_records` kind `dimension-settings`, id/code `default` (1 แถวต่อ company) และมี `version`
- payload ตามลำดับ Champ (บรรทัด 8-14 ของภาพ Option 2):
  `{departseparate, projectseparate, allocateseparate, jobseparate, partseparate, sideseparate, branchseparate}` เป็น bool
- **ค่าเริ่มต้น = ปิดทั้งหมด** ตาม DDL Champ `DEFAULT 0` (`SQLSERVER_Script.sql:743-750, :947-954, :9069-9075`) — **ไม่มีข้อยกเว้น ไม่สแกนข้อมูลเดิม** (กฎเดินหน้าอย่างเดียว)
  - holding ตัวอย่าง (rungrueng) ที่ใช้แผนก/โครงการอยู่ → เปิดสวิตช์ผ่าน seed/provision script หรือจอค่าเริ่มต้นบริษัท แล้วรายงานว่าเปิดอะไร
- ผลของสวิตช์ (ตาม Champ)
  1. จอรายวัน: ซ่อนคอลัมน์มิติที่ปิด (`glfrmjournal.cpp:227-256`); จอผังบัญชี: disable checkbox/ปุ่ม (`GLChartOfAccPage1DetailDlg.cpp:250-278`); ตัวกรองรายงาน: **แสดงแต่ disable** (`GLRepTrialBalanceDlg.cpp:160-190`)
  2. แก้รายวันตอนสวิตช์ปิด: คงค่าเดิมของบรรทัด ไม่ล้าง (Champ `:890-903` ไม่เขียนคอลัมน์ที่ปิด)
  3. backend ปฏิเสธ **รหัสใหม่** ในมิติที่ปิด (`journal_line_<k>_disabled`) = ส่วนเสริมเล็ก (Champ ซ่อนคอลัมน์จึงกรอกไม่ได้ แต่ API รับได้) — §7
  4. ปิดสวิตช์ได้เสมอแม้มีข้อมูล (Champ `myfrmsysconfig.cpp:1176-1183` ไม่ตรวจ); ข้อมูลเดิมยังอยู่และรายงานได้
- อ่านสวิตช์: ผู้ใช้ ledger ทุกคน (จอรายวัน/ผังบัญชี/รายงานต้องใช้); แก้สวิตช์: สิทธิ์ของจอค่าเริ่มต้นบริษัท (`menu-bars.ts:304` → `/branch`) — id สิทธิ์จริงต้องตรวจในโค้ดตอนทำ W3 ห้ามตั้งชื่อเอง
- `branchseparate` เก็บไว้ให้ครบ 7 ข้อตาม Champ แต่ยังไม่มีผลจนกว่าจะตัดสิน OD-2

### 3.5 กติกาตรวจ (เฉพาะผู้ใช้สร้าง/แก้รายวัน)
- ข้อ 2-5 ทำงานเฉพาะเมื่อ `checkSeparate=true` คือสร้าง (`postgres_store.go:587`) และแก้ (`:681`) เท่านั้น
- ผ่านรายการ (`:708`), กลับรายการ (`:792`), ปิดปี/ยอดยกมา (`postgres_processes.go:215`), คำนวณยอดใหม่ และ ERP → GL **ไม่ตรวจข้อ 2-5** — ต้องผ่านได้แม้สวิตช์/ธง/รายการที่อนุญาตเปลี่ยนภายหลัง (`models.go:512-516`; Champ ผ่านรายการไม่ตรวจมิติ `GLFrmPostTrans.cpp:722-800`)
- Champ ตรวจตอนบันทึกเฉพาะใบใหม่ (`glfrmjournal.cpp:574-590`) ส่วนใบแก้ไขกันที่ช่องกรอก (`:2812-2923`); BC ตรวจทั้งสร้างและแก้ที่ backend = ต่างโดยตั้งใจ เพราะ API ไม่มีช่องกรอก

ลำดับต่อบรรทัด ต่อมิติ `k` ตาม registry

| # | เงื่อนไข | ผล | code |
|---|---|---|---|
| 1 | รหัสว่าง | ผ่าน (ไม่บังคับ ตาม Champ) | — |
| 2 | สวิตช์บริษัทปิด และรหัสไม่ใช่ค่าเดิมของบรรทัด | 400 | `journal_line_<k>_disabled` |
| 3 | บัญชี `<k>separate=false` | 400 | `journal_line_<k>_not_separate` (Champ "NOT_SEPARATE") |
| 4 | รหัสไม่มีในทะเบียน หรือ `isactive=false` | 400 | `journal_line_<k>_not_found` |
| 5 | บัญชีไม่ได้ติ๊ก "เลือกทั้งหมด" และรหัสไม่อยู่ในรายการ (รายการว่าง = ไม่อนุญาตรหัสใด) | 400 | `journal_line_<k>_not_allowed` |

- ข้อ 5 ตาม Champ (`glfrmjournal.cpp:2815-2923`): ธงเปิด + เลือกทั้งหมด (`_ALL`) = ทุกรหัส; ธงเปิด + รายการ = เฉพาะรายการ; ธงเปิด + ไม่มีอะไร = ไม่อนุญาตรหัสใด
  - เปลี่ยนจาก BC เดิม "รายการว่าง = ทุกรหัส" (เลือกโดย AI 2026-10-07 ไม่ใช่คำสั่งลุงจืด) → seed บัญชีที่เปิดธงอยู่ให้ติ๊ก "เลือกทั้งหมด"
- Champ ไม่มีข้อความ error ข้อ 4-5 (ล้างช่องเงียบ ๆ); BC คืน error key เพราะ API ไม่มีช่องกรอกให้ล้าง
- บัญชี (`mutateAccount`): `<k>codes` ไม่ว่างแต่ธงปิด → `account_<k>_codes_need_flag`; รหัสในรายการไม่มีในทะเบียน → `account_<k>_code_not_found`
- ทะเบียน
  - `dimension_name_required`
  - `dimension_name_too_long` (field `name`)
  - `code_too_long` (ของเดิม `checkCode` `models.go:313-345` ส่ง `codeMax` ของ kind)
  - ลบหรือเปลี่ยนรหัสที่ถูกใช้แล้ว → 409 `dimension_in_use_delete` / `dimension_in_use_code` (ส่วนเสริมเดิมของ BC; Champ ไม่ตรวจ `myfrmbasemaster.cpp:509-536` — §7)
- ทุก error เป็น `UserError{code, field, message, status}` (`errors.go:10-25, :80-95`; ค่าคงที่ `duplicate_code`/`stale_version` มีแล้ว): 400 = แก้ช่องข้อมูล, 409 = ชนกับข้อมูลที่เก็บแล้วหรือ version เก่า

### 3.6 REST (ใช้ route เดิม ไม่เพิ่ม route ใหม่)

| ใช้ทำอะไร | Request | Response | Error keys |
|---|---|---|---|
| รายการทะเบียน | `GET /gl/v2/{departments\|projects\|allocate-codes\|job-codes\|part-codes\|side-codes}?q=&page=&limit=` (ตาม list handler เดิม `http.go:502-547` → `h.pg.List(... page, limit ...)`) | `{items:[{id,code,name,usergroup,isactive,version,createdat,updatedat}], total}` เรียง `code` | `forbidden` |
| อ่านทะเบียน 1 รายการ | `GET /gl/v2/{kind}/{id}` | master | `not_found` |
| สร้าง/แก้/ลบทะเบียน | `POST /gl/v2/command {resource:"job-codes", action:"create\|update\|delete", id?, version?, requestid, reason?, master:{code,name,isactive}}` | `Result{changes:[...], version}` และเขียน `gl_events` | `dimension_name_required`, `dimension_name_too_long`, `code_required`, `code_too_long`, `code_has_space`, `code_invalid_char`, `duplicate_code`, `stale_version`, `dimension_in_use_delete`, `dimension_in_use_code` |
| อ่านสวิตช์ | `GET /gl/v2/dimension-settings` — **เพิ่ม special-case ใน list handler** (แบบ `journal-support`, `http.go:502-547`) คืน object เดียว ไม่ใช่ `{items,total}`; ยังไม่มีแถว = ค่าเริ่มต้นปิดทั้งหมด `version:0` | `{departseparate,…,branchseparate,version}` | `forbidden` |
| แก้สวิตช์ | `POST /gl/v2/command {resource:"dimension-settings", action:"update", id:"default", version, dimensionsettings:{…}}` | `Result` และเขียน `gl_events` (audit) | `stale_version` (409 version ชน), `forbidden` |
| บันทึกรายวัน | เดิม `resource:"journals"`; บรรทัดเพิ่ม `allocatecode, jobcode, partcode, sidecode` | เดิม | กลุ่ม `journal_line_<k>_*` ตาม 3.5 |
| ผังบัญชี | เดิม `resource:"accounts"`; เพิ่ม `allocateseparate…sideseparate` และ `allocatecodes…sidecodes` | เดิม | `account_<k>_codes_need_flag`, `account_<k>_code_not_found` |
| รายงาน | `GET /gl/v2/reports/{trial-balance\|ledger\|ledger-summary\|working-paper\|budget*}` เพิ่ม `allocatecode, jobcode, partcode, sidecode` (คั่นจุลภาค = หลายรหัส) และ `byallocate, byjob, bypart, byside` | เดิม แต่ breakdown เพิ่มมิติ | `report_dimension_too_many` (`reports.go:39`) |
| รายงานทะเบียน | `GET /gl/v2/reports/{kind}?codefrom=&codeto=` สำหรับ kind ทั้ง 6 (ขยาย `reports.go:79` และ `:1110-1141` ให้ใช้ registry) และลงทะเบียนใน central report registry (E4) | `{columns, rows}` | `report_code_range_invalid` (`reports.go:1024`) |

- สิทธิ์ (`httpapi/http.go`)
  - `resourceScreens` (`:94`): kind ใหม่ทั้งหมด → `gl-dimensions`; `dimension-settings` → จอค่าเริ่มต้นบริษัท (Champ Option 2 ข้อ 8-14) — id สิทธิ์ตรวจในโค้ดตอน W3
  - `ledgerLookup` (`:158-160`): เพิ่ม kind ใหม่ทั้ง 4 **และ `dimension-settings`** (อ่านอย่างเดียว) เพราะจอรายวัน/ผังบัญชี/รายงานต้องรู้สวิตช์เพื่อซ่อน/disable คอลัมน์ ไม่เช่นนั้นได้ 403
- MCP: **ยังไม่ทำ** (memory "MCP work on hold") จด flag ไว้ใน W4

### 3.7 รายงาน (`reports.go`)
- `parseReportCodeList` เพิ่ม 4 array (`pq.Array`) เข้า `r.args` ฐาน (`:218-220` ตอนนี้ `$1-$11`) → เป็น `$12-$15`
  - ทุก query ที่ใช้ `r.args` ต้องอ้างครบทุก param ตามคอมเมนต์ที่ `:218` โดยเพิ่ม `AND (cardinality($12::text[])=0 OR allocate_code=ANY($12::text[]))` … `$15`
- param ที่ต่อท้ายภายหลัง **ห้ามเขียนเลขตายตัว** ให้สร้าง placeholder จาก `len(args)+1` แบบเดียวกับ limit/offset (`:282-283`)
  - ช่วงเลขที่เอกสาร `docNoFilter`/`withDocNoRange` (`:730-731` ตอนนี้ `$12/$13`) → เลื่อนตาม `len(r.args)`
  - เปรียบเทียบงบประมาณ (`:851-859` ต่อท้าย `$12` งวด, `$13` รหัสงบ) → เลื่อนตาม `len(r.args)`
  - รายละเอียดงบประมาณ (`:942, :950-955`) มี args ของตัวเอง `$1-$9` → มิติใหม่เป็น `$10-$13`
- test integration ทุกเส้นทาง: TB, ledger พร้อมช่วงเลขที่เอกสาร, เปรียบเทียบงบประมาณ, รายละเอียดงบประมาณ — ใส่ตัวกรองมิติครบ + ช่วงเลขที่พร้อมกัน
- `breakdown()` (`:373-383`) อ่านจาก registry แทนการเขียนตายตัว
- ป้ายคอลัมน์ใช้ key ภาษาแทนข้อความไทยฝังโค้ด เฉพาะส่วนที่แตะ
- งบการเงิน (`statements.go`): **ไม่แตะในงานนี้** — ตัวกรอง branch/department/project เดิม (`:165-172`) คงไว้; Champ ไม่มีตัวกรองมิติระดับงบ แต่ใส่มิติในสูตรแต่ละบรรทัดของรูปแบบงบ (`MyXpFinanceSheet.cpp:103-111, :409-474`) → ไปอยู่ในการออกแบบรูปแบบงบการเงิน
- BFF `frontend/src/app/api/gl/[...glPath]/route.ts:27` เพิ่ม 8 query key ใน allowlist
- รายงานทะเบียน 7 ตัว
  - **เมนูทุกชื่อของ Champ ยังอยู่** แต่เปิดจอรายงานกลางจอเดียว ตามการตัดสินใจ 2026-10-11 (registry ของ E4)
  - entry ใหม่ (ยังไม่มีใน `menu-data.ts`; id สุดท้ายให้ตรงรูปแบบ registry ของ E4): `report-gl-depart` (5538), `report-gl-project` (5521), `report-gl-allocate` (5524), `report-gl-job` (5526), `report-gl-part` (5527), `report-gl-side` (5528), `report-gl-branch` (5529)
  - 5529 (สาขา) แสดง "รอพัฒนา" จนกว่าจะตัดสิน OD-2

### 3.8 ERP → GL (W8 ทำหลัง workflow ERP เสร็จ)
- กฎยืนยันจาก Champ `gltlsprocess.cpp` (§1.7) — มิติมาจาก**หัวเอกสาร**เท่านั้น
- เพิ่ม `department_code, project_code, allocate_code, job_no` ระดับหัวเอกสาร ใน `transaction/httpapi/schema.go` (ตอนนี้ไม่มีคอลัมน์มิติเลย) ตาม Champ DDL (`JobNo VARCHAR(20)` บน 51 ตาราง)
- `posting.go mappingLines` คัดลง `gl.Line` ตาม Champ `:497-500`
  - แผนก/โครงการ/การจัดสรร: ใส่เมื่อ**ธงบัญชี**ของบรรทัดเปิด (`IsSeparate :9354-9381`) ไม่เช่นนั้น `''` — ไม่ดูสวิตช์บริษัท ไม่ดูรายการรหัสที่อนุญาต (Champ ไม่ดู)
  - งาน: ใส่ `job_no` ของหัวเมื่อสวิตช์บริษัท `jobseparate` เปิด (`:293-295, :883-885`) ไม่ดูธงบัญชี
  - บรรทัดปัดเศษ: คัดมิติหัวโดยไม่ดูธง (`:424-427, :447`)
  - หน่วยงาน/ฝ่าย: ไม่มาจาก ERP
- ไม่ error เรื่องมิติบนเส้นทาง ERP (เหมือน Champ) — การบันทึก ERP ไม่ล้มเพราะผังบัญชี; GL ที่ได้ใช้ `checkSeparate=false`
- BC ต่างโดยตั้งใจ: บันทึก GL แบบ sync ACID ใน tx เดียวกับเอกสาร (ไม่ใช่ batch process ของ Champ)

### 3.9 Frontend + เมนู
- **`gl-dimensions.tsx`** → จอ "กำหนดรหัสต่างๆ" แบบ Champ
  - combo เลือกประเภท 6 ตัวตามลำดับ Champ (การจัดสรร, โครงการ, งาน, แผนก, ฝ่าย, หน่วยงาน) แทน 2 tab
  - grid: รหัส, ชื่อ, กลุ่มพนักงาน/ใช้กับรายวัน (`usergroup`), ใช้งาน
  - ปุ่มตามภาพ Champ: Import, Export, F5 ตัวอย่าง, บันทึก (F11), พิมพ์ (เปิดรายงานกลางตามประเภท), ปิด
  - บันทึก (F11): ส่งทุกแถวที่เปลี่ยนเป็น command create/update/delete ทีละแถวผ่าน backend (ตาม `myfrmbasemaster.cpp:509-536`); error ของแถวใดแสดงที่แถวนั้น
  - Import: อ่านไฟล์ `CODE<TAB>NAME` (Champ `:743-864`) ใส่ grid เท่านั้น ยังไม่บันทึก → ผู้ใช้กด F11 แล้ว backend ตรวจตามกติกาเดิม (ไม่มี endpoint import ใหม่); ไฟล์ผิดรูปแบบ → `gl_err_dimension_import_format`
  - Export: ดาวน์โหลด `CODE<TAB>NAME` ของประเภทที่เลือก (Champ `:685-742`) จากรายการที่โหลดแล้ว
  - F5 ตัวอย่าง: Champ ลบของเดิมแล้วคัดจากฐานข้อมูลตัวอย่าง (`:1116-1135`) แต่ BC ไม่มีฐานข้อมูลตัวอย่าง → แสดง "รอพัฒนา" จนกว่าจะตัดสิน OD-8
  - ประเภทที่สวิตช์บริษัทปิดยังแก้ทะเบียนได้ (Champ ไม่ล็อกจอนี้)
  - `frontend/src/lib/general-ledger.ts:114` (`GL_RESOURCES`) และ `:171` (`DIMENSION_RESOURCES`) เพิ่ม 4 kind
- **เมนู** `menu-data.ts:632` id `gl-dimensions` (ห้ามเปลี่ยน id/route)
  - ป้ายเป็น "กำหนดรหัสต่างๆ" (ตัด "(แผนก/โครงการ)")
  - alias `:95` เพิ่ม กำหนดรหัสการจัดสรร/งาน/หน่วยงาน/ฝ่าย
  - เพิ่ม 7 entry รายงานทะเบียนในหัวรายงาน GL ตาม `menuconfig.xml:575-585`
- **สวิตช์บริษัท**: checklist 7 ข้อพร้อมป้ายตามภาพ Option 2 ในจอ "กำหนดค่าเริ่มต้นบริษัท" (`menu-bars.ts:304` → `/branch`, `app/[systemSetting]/page.tsx`) ตำแหน่งเดียวกับ Champ (`myfrmsysconfig.cpp:927-946`) — ไม่ทำ panel ในจอ `gl-dimensions`
- **`gl-journals.tsx`**
  - คอลัมน์ view (`:915-921`) และ edit grid (`:1196-1221`) เรียงตาม Champ: แผนก, โครงการ, การจัดสรร, งาน, หน่วยงาน, ฝ่าย
  - แสดงเมื่อสวิตช์เปิด หรือใบนั้นมีค่าอยู่แล้ว
  - ล็อกช่องเมื่อบัญชีปิดธง (ขยาย `departLocked/projectLocked :213-214` ให้วนตาม registry)
  - `DimensionSelect` (`gl-dimension-select.tsx:70`) รับ kind ใหม่; `useDimensionRegisters` (`:39`) โหลดเฉพาะ kind ที่เปิด
- **`gl-masters.tsx`** (ผังบัญชี): ธง, ช่อง "เลือกทั้งหมด" และตัวเลือกหลายรหัสสำหรับทุกมิติ (pattern `:128-131, :503-504, :1403-1404`) และ disable เมื่อสวิตช์ปิด (`GLChartOfAccPage1DetailDlg.cpp:134-217, :250-278`)
- **ไฟล์อื่นที่ส่งต่อมิติ**
  - `gl-reports.tsx`, `gl-print-conditions.tsx`, `account-ledger-drilldown.tsx`: ตัวกรองและ group-by — ตัวกรองมิติที่สวิตช์ปิด **แสดงแต่ disable** (`GLRepTrialBalanceDlg.cpp:160-190`)
  - `gl-budgets.tsx`: คอลัมน์มิติแสดงตามสวิตช์ (`GLFrmBudget.cpp:138-151`) §3.11
  - `lib/gl-journal-entry.ts`, `lib/gl-journal-templates.ts`, `lib/clipboard-journal-parser.ts` (Smart Paste รับคอลัมน์ใหม่)
- ทุกจอรับ prop `language` และใช้ `useBackendLanguage` ไม่ฝังข้อความ

### 3.10 Language keys (`backend/assets/language/languages.tsv`, 13 คอลัมน์; th ตามคำ Champ, en เป็นคำแปล, อีก 10 ภาษาใส่ en เป็น placeholder)
- มีแล้วและใช้ต่อ: `gl_department`, `gl_project`, `gl_line_department`, `gl_line_project`, `gl_department_code`, `gl_project_code`, `gl_project_separate`, `gl_dimensions`, `gl_dimension_*` (`:8930-8936`), `gl_err_dimension_*` (`:8937-8940`), `gl_err_account_depart/project_*` (`:8942-8945`), `gl_err_journal_line_depart/project_*` (`:8946-8949`)
- ใหม่ ต่อมิติ `<k>` ∈ {allocate, job, part, side}
  - `gl_<k>` (การจัดสรร/งาน/หน่วยงาน/ฝ่าย)
  - `gl_line_<k>`
  - `gl_<k>_code`
  - `gl_<k>_separate`
  - `gl_<k>_codes_label`
  - `gl_<k>_codes_pick`
  - `gl_<k>_report_title`
  - `gl_err_account_<k>_codes_need_flag`
  - `gl_err_account_<k>_code_not_found`
  - `gl_err_journal_line_<k>_not_found`
  - `gl_err_journal_line_<k>_not_allowed`
- ใหม่ ทุกมิติ (รวม depart/project)
  - `gl_err_journal_line_<k>_not_separate`
  - `gl_err_journal_line_<k>_disabled`
  - `gl_<k>_all_codes` (เลือกทั้งหมด)
  - `gl_err_budget_<k>_not_found`, `gl_err_budget_<k>_not_separate`, `gl_err_budget_<k>_not_allowed` (§3.11)
- ใหม่ ทะเบียน
  - `gl_dimension_usergroup` "กลุ่มพนักงาน/ใช้กับรายวัน"
  - `gl_dimension_import`, `gl_dimension_export`, `gl_dimension_examples`
  - `gl_err_dimension_import_format` (ไฟล์ไม่ใช่ `รหัส<TAB>ชื่อ`)
- สวิตช์บริษัท ใช้ข้อความ Champ ตามตัวอักษร
  - `gl_opt_sep_depart` "ใช้ผังบัญชีแบบแยกแผนก"
  - `gl_opt_sep_project` "ใช้ผังบัญชีแบบแยกโครงการ"
  - `gl_opt_sep_allocate` "ใช้ผังบัญชีแบบแยกการจัดสรร"
  - `gl_opt_sep_job` "ใช้ผังบัญชีแบบแยก JOB"
  - `gl_opt_sep_part` "ใช้ผังบัญชีแบบแยกส่วนงาน (Part)"
  - `gl_opt_sep_side` "ใช้ผังบัญชีแบบแยก Site"
  - `gl_opt_sep_branch` "ใช้ผังบัญชีแบบแยกสาขา"
  - `gl_dimension_settings_title`
- เมนูและจอ
  - `gl_dimension_type` (ประเภทรหัส)
  - `menu_report_gl_depart`, `menu_report_gl_project`, `menu_report_gl_allocate`, `menu_report_gl_job`, `menu_report_gl_part`, `menu_report_gl_side`, `menu_report_gl_branch` (ข้อความตาม `menuconfig.xml:575-585`)
- หมายเหตุ: Champ ใช้คำต่างกันระหว่างเมนูทะเบียน (หน่วยงาน/ฝ่าย) กับสวิตช์บริษัท (ส่วนงาน (Part)/Site) ให้ใช้ตามที่ปรากฏในแต่ละจอ

### 3.11 มิติในงบประมาณ (ตาม Champ `BCGLBudget :9207-9230`, `GLFrmBudget.cpp:138-151`)
- `budget.sql`: เพิ่ม `allocate_code, job_code, part_code, side_code` (§3.3); PK `(company, code)` เดิม; `amount numeric(18,2)` เดิม
- `budgets.go`
  - struct `Budget` (`:48-56`) เพิ่ม `AllocateCode/JobCode/PartCode/SideCode`
  - ตรวจ (`:225`): `checkCode` ตามความยาวของ registry + รหัสต้องมีในทะเบียนและใช้งานอยู่ (`not_found`) + **ธงบัญชี** ตาม Champ: บันทึกงบ (`GLFrmBudget.cpp ConfrimSaveData :421-480`) อ่านเฉพาะมิติที่สวิตช์เปิด แล้ว `OnCheckSeparate :482-530` → "Chart_Of_Account <code> NOT_SEPARATE" `:471`; ช่องกรอกกันรหัสที่ไม่อนุญาต (`OnCheckComboData :1575`, `BCAccSeparate :1589`) → BC ใช้ข้อ 3 และ 5 ของ §3.5 เหมือนรายวัน
  - INSERT (`:296`), UPDATE (`:335`), `budgetHeaderColumns` (`:406`) เพิ่ม 4 คอลัมน์
- REST: แถวงบประมาณเพิ่ม `allocatecode, jobcode, partcode, sidecode`; error key `gl_err_budget_<k>_not_found`, `gl_err_budget_<k>_not_separate`, `gl_err_budget_<k>_not_allowed`
- รายงาน: เปรียบเทียบงบประมาณ (`reports.go:851-859`) และรายละเอียดงบประมาณ (`:942-959`) กรองครบมิติ ตามการนับ param ใน §3.7
- จอ: คอลัมน์มิติแสดงตามสวิตช์บริษัท
- `dimensionInUse` อ่านคอลัมน์ใหม่ของ `gl_budgets` ได้จริง (ไม่มีคอลัมน์ตาย)

---

## 4. Tests
- **Go unit** (`generalledger`)
  - registry: ทุก kind มี column/jsonKey/errKey ไม่ซ้ำ และ codeMax/nameMax ตาม §3.1
  - `validateDimensionMaster`: job 15 ผ่าน, 16 ไม่ผ่าน; part ชื่อ 255 ผ่าน
  - กติกา §3.5 ทั้ง 5 ข้อ × 6 มิติ (table test)
  - `TestParseReportCodeList` เพิ่มกรณีของ param ใหม่
- **sqlmock** (`gopkg.in/DATA-DOG/go-sqlmock.v1`)
  - store ของ `dimension-settings`: อ่านค่าเริ่มต้น, update ที่มี version, `stale_version`
  - `dimensionInUse` สร้าง SQL ครบตาม kind ใหม่
- **Integration** (`BC_GL_TEST_POSTGRES_DSN`, ต่อยอดไฟล์เดิม)
  - `dimensions_integration_test.go`: lifecycle CRUD ต่อ kind ใหม่ (`:76` pattern), รหัสบรรทัด (`:238`), template (`:300`), รายการบัญชี (`:317`)
  - `report_dimension_filter_integration_test.go`: กรองและ breakdown สำหรับ 4 มิติ, งบประมาณ, ปิดงวดจัดกลุ่มครบ 6 มิติ
  - ทุกเส้นทาง `r.args` (§3.7): TB, ledger + ช่วงเลขที่เอกสาร, เปรียบเทียบงบประมาณ, รายละเอียดงบประมาณ — ใส่ตัวกรองครบ 6 มิติ + ช่วงเลขที่พร้อมกัน ต้องไม่ error "could not determine data type of parameter" และยอดถูก
  - ผ่านรายการหลังเปลี่ยนผัง: บันทึกฉบับร่าง → เปิดธงบัญชี/ปิดสวิตช์/ลบรหัสออกจากรายการที่อนุญาต → ผ่านรายการ**สำเร็จ** (§3.5)
  - กลับรายการ: บรรทัดกลับมีครบ 6 มิติเท่าต้นฉบับ; ปิดปี: ยอดยกมาแยกตามครบทุกมิติ
  - ธงบัญชี: เปิดธง + ไม่เลือกทั้งหมด + ไม่มีรายการ → `not_allowed`; เลือกทั้งหมด → ผ่าน
  - งบประมาณ: บันทึก/แก้แถวที่มี 4 มิติใหม่, รหัสไม่มีในทะเบียน → `budget_<k>_not_found`, ธงบัญชีปิด → `budget_<k>_not_separate`
  - `dimension-settings`: ไม่มีแถว = ปิดหมด `version:0`; ผู้ใช้ ledger (ไม่มีสิทธิ์จอค่าเริ่มต้นบริษัท) อ่านได้ แก้ไม่ได้ (403)
  - ensureSchema รันซ้ำ 2 ครั้งไม่ error
- **vitest**
  - `gl-dimensions` เปลี่ยน combo แล้วโหลด kind ถูก
  - `gl-journals` ซ่อนหรือแสดงคอลัมน์ตามสวิตช์ และแสดงคอลัมน์เมื่อใบมีค่า
  - ตัวกรองรายงาน: สวิตช์ปิด = แสดงแต่ disable
  - Import: ไฟล์ `รหัส<TAB>ชื่อ` เติม grid ไม่เรียก API จนกด F11; ไฟล์ผิดรูปแบบ → `gl_err_dimension_import_format`; Export ได้ไฟล์ตรงรูปแบบเดียวกัน
  - `gl-language-keys.test.ts` ครอบ key ใหม่
- **UAT CRUD** (Demo → `rungrueng` → `01` → `00000`) ตรวจ `psql` ทันทีหลังแต่ละขั้น
  1. สร้างรหัสงาน `J-2569-001` "งานติดตั้งระบบโรงงานบางปู" → `SELECT code,payload->>'name',version FROM gl_records WHERE company='01' AND kind='job-codes' AND code='J-2569-001';` = 1 แถว
  2. แก้ชื่อ → แถวเดิม `version`+1 ไม่มีแถวใหม่
  3. เปิดสวิตช์ JOB และเปิดธงบัญชี 1 บัญชี → `SELECT payload FROM gl_records WHERE company='01' AND kind='dimension-settings';`
  4. บันทึกรายวันที่ใส่ job → `SELECT job_code,debit,credit FROM gl_lines WHERE journal_id=$id;` และตรวจ Σdebit=Σcredit
  5. ลบรหัสที่ถูกใช้ → ได้ 409 `dimension_in_use_delete` และแถวยังอยู่
  6. ลบรายวันฉบับร่าง แล้วลบรหัส → แถวหาย
  7. edge cases: รหัสว่าง, ซ้ำ, อักขระไทยและเครื่องหมายกำกับ, ยาว 16, double-click บันทึก
  8. ตรวจว่า record อื่นไม่ถูกแตะ; ลบข้อมูลทดสอบด้วย code เป้าหมายเท่านั้น
- `npm run verify:all` ก่อน deploy

## 5. ความเสี่ยง
- R1 `reports.go` ใช้ positional param ร่วมกัน: `$12/$13` ถูกใช้แล้วโดยช่วงเลขที่เอกสาร (`:730-731`) และงบประมาณ (`:851-859`) → ต้องเลื่อนด้วย `len(args)+1` ไม่เขียนเลขตายตัว; รายละเอียดงบประมาณนับใหม่ `$10-$13` (§3.7) และมี integration test ครอบทุกเส้นทาง
- R2 `dimensionInUse` ต้องมีคอลัมน์ใน `gl_budgets` ก่อน (§3.3) ถ้าไม่มี การลบทะเบียนจะล้ม → ทำพร้อม §3.11 ใน W2
- R3 อีก workflow กำลังแก้ `postgres_guards.go`, `source_posting.go` และ `transaction/**` อยู่ → W2 ต้อง rebase หลังจากนั้น ส่วน W8 ทำหลังงาน ERP ปิด
- R4 การซ่อนคอลัมน์ในจอรายวันอาจบังข้อมูลเดิม → ใช้กฎ "แสดงถ้าใบมีค่า" (ส่วนเสริม §7); ตัวกรองรายงานแสดงแต่ disable ตาม Champ (`GLRepTrialBalanceDlg.cpp:160-190`) ไม่ซ่อน
- R5 ค่าเริ่มต้น "ปิดหมด" ทำให้ tenant ที่ใช้แผนก/โครงการอยู่แล้ว (prod rungrueng) เห็นคอลัมน์หาย → เปิดสวิตช์ผ่าน seed/provision แล้วรายงาน (ไม่สแกนข้อมูลอัตโนมัติ — กฎเดินหน้าอย่างเดียว)
- R9 เปลี่ยนความหมาย "รายการว่าง" เป็นแบบ Champ (ไม่อนุญาตรหัสใด) → บัญชีที่เปิดธงแต่ไม่มีรายการจะบันทึกมิติไม่ได้ทันที → seed ต้องติ๊ก "เลือกทั้งหมด" ให้บัญชีเดิมใน W2 และ UAT ตรวจ prod rungrueng หลัง deploy
- R6 ชื่อเอกสารผิด: `09-gl.md:49-55` และ handoff เรียก 5527/5528 ว่า "ส่วนงาน/ด้าน" → แก้ docs ใน commit ของ W4
- R7 ความยาวรหัส job (15) ต่างจากมิติอื่น (10) → UI ต้องใช้ `codeMax` จาก registry ห้ามฝังค่า 10
- R8 ข้อความหัวคอลัมน์ ALLOCATE/JOB/PART/SIDE ในจอรายวันของ Champ อยู่ใน string table ของ exe ที่หาไม่พบ (`BCSWinApp.cpp:524 GetRes`) → ใช้ชื่อทะเบียนไปก่อน (OD-6)

## 6. Open decisions (หลังตรวจทาน เหลือ 3 ข้อ ไม่มีข้อใดขวางการเริ่มงาน)
- OD-2 สาขาเป็นมิติระดับบรรทัดหรือไม่ — แนะนำ: คงระดับหัวรายการแบบเดิม, รายงาน 5529 แสดง "รอพัฒนา"
- OD-6 หัวคอลัมน์ในจอรายวัน — แนะนำ: ใช้ชื่อทะเบียน (การจัดสรร/งาน/หน่วยงาน/ฝ่าย) เพราะหา string table ของ Champ ไม่พบ
- OD-8 ปุ่ม F5 ตัวอย่าง — แนะนำ: แสดง "รอพัฒนา" ก่อน (ของ Champ ลบข้อมูลเดิมและ BC ไม่มีฐานข้อมูลตัวอย่าง)
- ตัดออกเพราะ Champ ตอบแล้ว: OD-1 (ค่าเริ่มต้นปิด, อยู่จอค่าเริ่มต้นบริษัท, ปิดได้เสมอ), OD-3 (ERP คัดจากหัว ตาม `gltlsprocess.cpp`), OD-4 (รายการว่าง = ไม่อนุญาต, มี "เลือกทั้งหมด"), OD-5 (เก็บและแสดง UserGroup ไม่กรอง), OD-7 (ชื่อ kind เป็นเรื่องเทคนิค ตัดสินใน §3.1)

## 7. ส่วนเสริมนอก Champ (เล็กและไม่บังคับ ต้องรายงาน)
1. กฎ "แสดงคอลัมน์ถ้าใบมีค่า" แม้สวิตช์ปิด เพื่อกันข้อมูลหาย
2. ช่อง "ใช้งาน" (`isactive`) ในทะเบียน ซึ่ง BC มีอยู่แล้ว
3. ค้นหาหรือกรองในจอกำหนดรหัส
4. error ที่บอก field และวิธีแก้เป็นภาษาไทย (Champ ล้างช่องเงียบ ๆ)
5. backend ปฏิเสธรหัสใหม่ในมิติที่สวิตช์ปิด (`journal_line_<k>_disabled`) — Champ ซ่อนคอลัมน์จึงกรอกไม่ได้ แต่ API รับได้
6. ห้ามลบ/เปลี่ยนรหัสทะเบียนที่ถูกใช้แล้ว (`dimensionInUse`, BC มีอยู่แล้ว) — Champ ไม่ตรวจ
7. งบการเงินกรอง branch/department/project (`statements.go:165-172`, BC มีอยู่แล้ว) — คงไว้ ไม่ขยาย

ไม่มีเมนูหรือจอใหม่: รายงานทะเบียน 7 ตัวเป็นเมนูของ Champ และทุกตัวเปิดจอรายงานกลาง

## 8. Work items
ดู StructuredOutput: W1 ถึง W7 คือ GL (ทำได้ทันที), W8 คือ ERP (ทำหลังงาน ERP)
- W2 รวมมิติในงบประมาณ (§3.11), รายการกลับ/ปิดปีครบ 6 มิติ, ช่อง "เลือกทั้งหมด" + seed
- W4 ไม่แตะ `statements.go` แล้ว
- W8 ใช้กฎ ERP ที่ยืนยันจาก Champ (§1.7, §3.8) รวม `job_no` บนหัวเอกสาร

## ปรับปรุงหลังตรวจทาน (2026-10-11)
- P1 รับ: param รายงานชน `$12/$13` → มิติใหม่ `$12-$15` ใน `r.args`, param ต่อท้ายใช้ `len(args)+1`, รายละเอียดงบ `$10-$13`, test ทุกเส้นทาง (§3.7, R1, §4)
- P2 รับ: กติกามิติตรวจเฉพาะสร้าง/แก้ (`checkSeparate=true`); ผ่านรายการ/กลับรายการ/ปิดปี/ERP ไม่ตรวจ + test ผ่านรายการหลังเปลี่ยนผัง (§3.5, §1.8)
- P3 รับ: trace `gltlsprocess.cpp` แล้ว — ERP คัด แผนก/โครงการ/การจัดสรร ตามธงบัญชี, งานจาก `JobNo` ตามสวิตช์บริษัท, ไม่มีเงื่อนไขเพิ่ม; เพิ่ม `job_no` หัวเอกสาร; ตัด OD-3 (§1.7, §3.8)
- P4 รับ: `dimension-settings` เข้า `ledgerLookup` (อ่าน), แก้ด้วยสิทธิ์จอค่าเริ่มต้นบริษัท (§3.4, §3.6)
- P5 รับ: special-case object เดียว + version, list ใช้ `page/limit` (§3.6)
- P6 รับ: ค่าเริ่มต้นปิดหมดตาม Champ ไม่สแกนข้อมูล, เปิดผ่าน seed แล้วรายงาน, ตัด `dimension_disable_in_use` และ OD-1 (§3.4, R5)
- P7 รับบางส่วน: ตัดการแก้ `statements.go`; ไม่รับหลักฐาน `BCGLFormat` (เป็นรูปแบบผ่านรายการ ERP ไม่ใช่รูปแบบงบ) — Champ ใส่มิติในสูตรบรรทัดของรูปแบบงบ (`MyXpFinanceSheet.cpp`) ไปอยู่ในงานออกแบบรูปแบบงบ (§1.6, §3.7)
- P8 รับ: Import/Export แบบ Champ (`CODE<TAB>NAME`, Import เติม grid แล้วบันทึกด้วย F11) + F5 เป็น OD-8 (§1.5, §3.9)
- P9 รับทางเลือก (a): มิติงบประมาณครบ — struct, ตรวจทะเบียน+ธงบัญชีตาม Champ `GLFrmBudget.cpp:421-530`, INSERT/UPDATE, REST, ตัวกรองรายงาน (§3.11)
- P10 รับทั้ง 4: encoding CP874, สาขามีผล 10, เงินงบ `numeric(18,2)` ไม่ขยาย, ตัวกรองรายงานแสดงแต่ disable (หัวเอกสาร, §1.1, §3.3, §3.4, R4)
- เพิ่มเติมตามกฎ "ประมวลผลยึด Champ": §1.8 ตารางขั้นตอนพร้อม file:line และข้อต่างโดยตั้งใจ; ข้อ 5 รายการว่างเปลี่ยนเป็นแบบ Champ + "เลือกทั้งหมด" (ตัด OD-4); เก็บ UserGroup (ตัด OD-5)
