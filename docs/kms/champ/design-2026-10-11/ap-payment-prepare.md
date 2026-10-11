# Design: บันทึกเตรียมจ่ายชำระหนี้ (audit C1 / ORD-08) + วันครบกำหนดจากจำนวนวันเครดิต (คำถามประกอบ C2)

สถานะ: ข้อเสนอการออกแบบ วันที่ 2026-10-11 ยังไม่ได้ลงมือเขียนโค้ด (ปรับหลังตรวจทาน 2026-10-11 ดูท้ายเอกสาร)
หมายเหตุเลขบรรทัด: `posting.go`/`save.go` มีงานอื่นแก้ค้างใน working tree เลขบรรทัดจึงขยับได้ ให้ยึดชื่อฟังก์ชันที่อ้างคู่กัน (`postingAmounts`, `subledgerSpecs`, `writeSubledger`, `debtOutstanding`, `cleanSubledger`, `releasePosting`)
ขอบเขต: เมนู `ap-payment-prepare` (`frontend/src/lib/menu-data.ts:319`) การดึงใบเตรียมจ่ายเข้าบันทึกจ่ายชำระหนี้ (PV, เมนู `pay` ที่ `menu-data.ts:320`) และวันครบกำหนดของเอกสารหนี้ฝั่ง ERP

## 0. คำตอบสั้น

- ใน Champ ใบเตรียมจ่ายเป็นเอกสารจริง ไม่ใช่แค่ worklist
  - มีเลขที่เอกสารเป็น PK มีหัวเอกสาร และบรรทัดที่อ้างเอกสารหนี้เจ้าหนี้ ซึ่งเก็บยอดคงเหลือและยอดตัดจ่ายของแต่ละใบ
  - ระบบจองยอดไว้บนเอกสารหนี้ (`PrePayStatus`/`PrePayAmount`)
  - ใบจ่ายชำระหนี้ (PV) ดึงใบเตรียมจ่ายได้ด้วย Ctrl+F3 แล้วอ้างกลับผ่าน `PaybillNo` + `RefType=1`
- ใบเตรียมจ่ายไม่มี workflow อนุมัติ ไม่มี WHT ไม่มีธนาคาร/วิธีจ่าย และไม่ลงบัญชีจริง (GL source 52 เป็นแบบจำลอง)
- เช็คไม่ได้เกิดตอนเตรียมจ่าย แต่เกิดที่ PV ผ่าน F7=บันทึกจ่ายเงิน
- คำแนะนำ:
  - ทำเป็น **เอกสารเตรียมจ่าย + PV ดึงไปจ่าย (Ctrl+F3)**
  - **ไม่สร้าง PV + เช็คในขั้นเดียว**
  - เหตุผลอยู่ใน §3
- ยอดที่จ่ายได้คำนวณใน PostgreSQL ด้วย `NUMERIC`: `ยอดหนี้ที่ผ่านบัญชี − ยอดตัดชำระที่ยังไม่กลับรายการ − ยอดที่ใบเตรียมจ่ายอื่นจองไว้และยังไม่ถูก PV ใช้`
- วันครบกำหนด (คำถามประกอบ C2 — C2 ตัวจริงคือ "ช่วงอายุหนี้ตั้งค่าได้" ซึ่งทำเสร็จแล้วใน commit `387aa183`; รายงานอายุหนี้จัดช่วงตาม `payload.due_date` จึงต้องได้วันครบกำหนดที่ถูกก่อน):
  - เอกสารหนี้ ERP ต้องรับ `creditday`/วันครบกำหนดที่หัวเอกสาร แล้วเก็บลง `doc.perioddatetime` และ `gl_subledger_documents.payload.due_date`
  - ค่าตั้งต้นของจำนวนวันเครดิตมาจากทะเบียนเจ้าหนี้ ตาม Champ `BCAP.CreditDay`
  - โค้ดตอนนี้ใช้วันที่เอกสารเป็นวันครบกำหนด (`posting.go` `writeSubledger` บรรทัด `payload["due_date"] = date` ราว `:415`)

## 1. พฤติกรรม Champ (หลักฐาน)

### 1.1 หน้าจอ (bcspec ชนะ .rc)

ภาพ `D:/project-champ/bcspec/หน้าจอเมนูเจ้าหนี้/บันทึกเตรียมจ่ายจ่ายชำระหนี้.jpg` ชื่อจอ "บันทึกเตรียมจ่ายชำระหนี้" เรียงฟิลด์ตามนี้:

| ส่วน | ฟิลด์ (ลำดับตามจอ) |
|---|---|
| หัว | เลขที่เอกสาร (บังคับ, ช่องสีชมพู), วันที่จ่ายชำระหนี้ (ค่าเริ่ม = วันนี้), รหัสเจ้าหนี้ + ปุ่มค้น + ชื่อ, หมายเหตุ (หลายบรรทัด) |
| ตาราง | เลขที่เอกสาร, ลงวันที่ (เทา=อ่านอย่างเดียว), รายวัน/คำอธิบาย, มูลค่าคงเหลือ (เทา), ยอดตัดจ่าย, S |
| ท้าย | โครงการ, การจัดสรร, กลุ่มของบิล, รหัสแผนก, กำไร(ขาดทุน), รวมเป็นเงิน(บาท) (ปิด), รวมเป็นเงิน, ยอดส่วนลด, ยอดจ่าย |
| ปุ่ม | New, Save, Del, Find, Prev, Next, Close, Info, F1 Key, Recur |

- ภาพ `.../บันทึกจ่ายชำระหนี้.jpg` (PV) ใช้ layout เดียวกัน ต่างกันที่มีป้าย PAY และคำแนะนำ **"Ctrl+F3=ใบเตรียมจ่าย"** แปลว่า PV ดึงใบเตรียมจ่ายได้
- เมนูเจ้าหนี้ (`.../หน้าจอเมนูเจ้าหนี้/เมนู-ประกอบ.jpg`) เรียง "บันทึกใบรับวางบิล | บันทึกเตรียมจ่ายชำระหนี้, บันทึกจ่ายชำระหนี้, บันทึกตัดหนี้สูญ(เจ้าหนี้)"
- ใน `menuconfig.xml` ไม่มีรายการเตรียมจ่าย มีเพียง `:175` "บันทึกจ่ายชำระหนี้" และใน `BC5Account.rc` ไม่มีคำว่า "เตรียม" ใช้ภาพเป็นหลักตาม AGENTS.md
- ตารางของ PV กำหนดคอลัมน์ไว้ที่ `D:/project-champ/champ/champ/apfrmappayment1view.cpp:207-212`
  - DOCNO, DATED, DAY/EXPLAIN, ITEMAMOUNT, CUT_PAY_AMOUNT แล้วตามด้วย `"S"` (กว้าง 15, `FCS_NULL` = อ่านอย่างเดียว)
  - คอลัมน์ S คือ **BillType** (`:759` อ่าน, `:970` บันทึก)
  - ความหมายของรหัส (loader ค้นเอกสาร `:3860-3896`, ป้าย MyDesc = ชื่อประเภทเอกสาร):
    - 0=ใบรับสินค้า/ใบส่งของ (`BCApInvoice`), 1=ยอดยกมาต้นปี (`BCApInvBalance`), 2=ใบลดหนี้ (`BCStkRefund`), 3=ใบเพิ่มหนี้ (`BCDebitNote2`), 4=เจ้าหนี้อื่นๆ (`BCAPOtherDebt`)
    - 5=มัดจำ (`BCAPDeposit`) โหลดเฉพาะใต้ `#ifdef TOHOME_PROJECT` (`:3882-3887`) จึง **ไม่อยู่ใน parity** ของ Champ มาตรฐาน
    - 7=ใบรับวางบิล (`BCStatementSub`) และ 8=**จ่ายเงินล่วงหน้า** (`BCAPDepositSpecial`, `:3892-3894`; GL source 42 'บันทึกจ่ายเงินล่วงหน้า (เจ้าหนี้)' SQL `:15406`) — ไม่ใช่ "รับเงินล่วงหน้า"
  - ใบเตรียมจ่ายใช้ตารางชุดเดียวกัน (DDL §1.2) จึงถือว่าคอลัมน์ S มีความหมายเดียวกัน

### 1.2 DDL (`D:/project-champ/champ/champ/Script/SQLSERVER_Script.sql`, decode windows-874)

- `BCPaymentTemp` (`:13419-13483`, PK `DocNo`)
  - โครงเหมือน `BCPayment` (`:7248-7316`) แต่มี `PayStatus smallint default 0` (`:13481`) และ `BillStatus` (0=ยัง, 1=หมด) เพิ่ม
- `BCPaymentTempSub` (`:13488-13515`) มี InvoiceDate, InvoiceNo, InvBalance, PayAmount, MyDescription, BillType, PaybillNo, TransState, IsCancel, LineNumber และฟิลด์สกุลเงิน (ไม่มี RefType)
- หัว `BCPaymentTemp` มีคอลัมน์ IsConfirm/ConfirmCode/IsCancel/IsPostGL/GLFormat ด้วย แต่ไม่มีโค้ดจอใบเตรียมจ่ายใน repo (อ้างถึงแค่ใน `gltlsprocess.cpp`) จึงไม่มีหลักฐานว่าใช้ขั้นอนุมัติ → รอบนี้ไม่ทำ confirm
- การจองยอดบนเอกสารหนี้ใช้ `PrePayStatus smallint`/`PrePayAmount money` ใน 7 ตาราง
  - BCAPINVBALANCE `:5774`, BCAPOtherDebt `:5828`, BCAPINVOICE `:5999`, BCStkRefund `:6370`, BCDebitNote2 `:6587`, BCStatement `:7208`, BCAPDepositSpecial `:12327`
  - ค่าเริ่มต้นตั้งที่ `:14461-14475`
- การอ้างจาก PV:
  - `BCPaymentSub.PaybillNo` + `RefType` (`:7341`): 0=ใบรับวางบิล, 1=ใบเตรียมจ่ายค่าสินค้า
  - ใน `apfrmappayment1view.cpp:220` shadow `PayBillNo` ระบุว่า "เฉพาะใบวางบิล" ซึ่งเป็นโค้ดรุ่นเก่ากว่าภาพ 5.5
  - โค้ดบันทึก PV ใน repo **ไม่เขียน RefType** (บรรทัด `:982` ถูก comment ไว้ พร้อมคำอธิบาย `:979-981`) บันทึกแค่ PayBillNo จาก shadow (`:971`)
- ใบรับวางบิล (`BCStatement`) ก็มี PrePayStatus/PrePayAmount (`:7208`) จึงจองยอดและถูกดึงเข้าใบเตรียมจ่ายได้เช่นกัน → เป็นคำถามให้ลุงจืด (§7)

### 1.3 การลงบัญชี

- `BCGLSource` 52 = `bcpaymenttemp` 'เตรียมจ่ายชำระ' (SQL `:15409`)
- `D:/project-champ/champ-lib/BC5PROCESS/gltlsprocess.cpp:5941-5960` (ProcessConditionGroup9) เขียนลง `bctrans_tmp` ซึ่งเป็นการจำลองเท่านั้น **ไม่ลงสมุดรายวันจริง**
- ข้อสรุปคือใบเตรียมจ่ายไม่ต้องมี Debit=Credit เพราะไม่แตะ GL

### 1.4 การดึงไป PV และการคำนวณ (จาก `apfrmappayment1view.cpp`)

- การดึงใบรับวางบิล `StatementBill` (`:1426-1584`) ซึ่งเป็นกลไกเดียวกับ Ctrl+F3:
  1. เลือกได้หลายใบของเจ้าหนี้รายเดียวกับ PV (`ApCode` ตรงกัน) ที่ `BillStatus<>1` และไม่ใช่ใบที่ดึงมาแล้วในจอ
  2. เพิ่มทุกบรรทัดที่ `PayBalance>0` โดยตั้ง `InvBalance = PayAmount = PayBalance`
  3. ตั้ง shadow `PaybillNo` = เลขที่ใบต้นทาง; คำอธิบายบรรทัดเป็นป้าย "STATEMENT"
  - ตัวจัดการ Ctrl+F3 ของใบเตรียมจ่ายไม่อยู่ใน repo (repo เป็นรุ่นที่ F3=ภาษีซื้อ, `:2366-2400`) จึงอนุมานจากใบรับวางบิล **(สมมติฐาน)**
- สูตร (`:1330-1409`, `:2043-2073`, `:3320-3339`):
  - BillType 2/5/8 เป็นยอดลบ ประเภทอื่นเป็นยอดบวก
  - รวมเป็นเงิน(บาท) = ΣHomeAmount2
  - ยอดจ่าย = รวมเป็นเงิน − ยอดส่วนลด − WHT
  - กำไร(ขาดทุน) = ΣExchangeProfit
- เมนู More ของ PV (`:2366-2400`): F2 GL, F3 ภาษีซื้อ, F6 หัก ณ ที่จ่าย, F7 บันทึกจ่ายเงิน (เช็ค/โอน), F8 รายการตัดจ่าย
  - WHT และเช็คจึงอยู่ที่ PV
- ข้อเท็จจริงภาษี: ประมวลรัษฎากร ม.50 ให้ผู้จ่ายเงินได้ "หักภาษีเงินได้ไว้ทุกคราวที่จ่ายเงินได้พึงประเมิน" ([rd.go.th/5937.html](https://www.rd.go.th/5937.html)) ตรงกับที่ Champ หัก WHT ตอนจ่าย (PV) ไม่ใช่ตอนเตรียมจ่าย
  - ม.50 ครอบเงินได้ของบุคคลธรรมดา (ภ.ง.ด.3) ส่วนผู้รับเป็นนิติบุคคล (ภ.ง.ด.53) ใช้ ม.3 เตรส + คำสั่ง ท.ป.4/2528 — **ยังไม่ได้ตรวจกับ rd.go.th ในรอบนี้** ต้องตรวจและลงทะเบียนใน `docs/kms/21-thai-tax-form-references.md` ก่อนทำ WHT ของ PV (ข้อสรุป "หักตอนจ่าย" ไม่เปลี่ยน เพราะ Champ ทำที่ PV F6)

### 1.5 เช็คเตรียมจ่ายเป็นคนละเรื่อง

- ใน `cqfrmcheqout.cpp:148-156` "เตรียมจ่าย" (PREPAIR_PAYMENT) เป็น **สถานะเช็ค**
- จอ "บันทึกเช็คเตรียมจ่าย" (`bcspec/หน้าจอเมนูเช็ค บัตร/บันทึกเช็คเตรียมจ่าย.jpg`) ใช้ลงทะเบียนเลขเช็คล่วงหน้า มีฟิลด์ เลขที่เช็คจ่าย, จำนวนที่ Run Auto, ธนาคาร, สมุดเงินฝาก
- จอนี้ไม่เกี่ยวกับใบเตรียมจ่ายชำระหนี้ จึงอยู่นอก design นี้ (เมนู `cheque-issued-prepare`, `menu-data.ts:415`)

### 1.6 วันครบกำหนด (คำถามประกอบ C2)

- `ChmFrmGoodsPurchasePage1.cpp:813-825, 910-926`: `DueDate = DocDate + CreditDay`
  - CreditDay ตั้งต้นจาก `BCAP.CreditDay` (SQL `:5599`) และแก้ได้
  - เมื่อ nDocType = 0/2 ระบบล้างและปิดช่อง CreditDay แล้วใช้ due = วันที่เอกสาร
  - ความหมายประเภท (`champ-report/BC5REP_AR/ApRepApInvBalanceView.cpp:188-193`): 0=ซื้อสด, 1=ซื้อเชื่อ, 2=บริการสด, 3=บริการเชื่อ → 0/2 คือ **ประเภทเงินสด**
- `ChmFrmGoodsPurchase.cpp:2221-2227`: ถ้า DueDate ว่างตอนบันทึก ใช้ DocDate แทน **เฉพาะ BillType 1/3 (เชื่อ)** ตามคอมเมนต์ "บิลขายเชื่อ บังคับให้ใส่วันที่ครบกำหนด"
- `BCAPINVOICE` มีคอลัมน์ CreditDay/DueDate ที่ `:5925-5926`

### 1.7 ลำดับประมวลผลของ Champ ที่ยึดเป็นต้นแบบ (กฎ 2026-10-11 "ประมวลผล/workflow ยึด Champ เสมอ")

ไม่มีโค้ดจอใบเตรียมจ่ายใน repo จึงใช้ **ใบรับวางบิล** (`champ/champ/apfrmapstatement1view.cpp`) เป็นต้นแบบการจองยอด และ **PV** (`apfrmappayment1view.cpp` + `champ-lib/BC5PROCESS/bcprocess.cpp`) เป็นต้นแบบการตัดหนี้ ตารางนี้คือสิ่งที่ BC ต้องทำตาม และจุดที่ BC ตั้งใจต่าง:

| ขั้น | Champ (ลำดับ/ตรวจ/ข้อความ/ตารางที่แตะ) | BC | ต่างโดยตั้งใจ |
|---|---|---|---|
| ตรวจก่อนบันทึกใบรับวางบิล | `ConfirmSaveData :623-701` ตามลำดับ: ไม่มียอด `CANNOT_SAVE_NOT_MONEY` → ไม่มีเลขที่ → ปิดงวด `IsPeriodClose` → ไม่มีเจ้าหนี้ → ไม่มีบรรทัด → (แก้ไข) อนุมัติแล้ว `NOT_SAVE_ISCONFIRM` / ถูกใช้แล้ว `BillStatus!=0` → `CANNOT_UPDATE_STATEMENT` → เลขซ้ำ | §4.3 ข้อ 1-6, 11, 15 เรียงตามนี้ | ไม่มีขั้นอนุมัติ (ไม่มีหลักฐานจอ) |
| บันทึกใบรับวางบิล | `OnSaveData :280-401`: StartTransaction → insert/update หัว → (แก้ไข) `StatementProcessBill(เดิม,-1)` คืนยอดจองเก่า → ลบ/เขียนบรรทัดใหม่ `PAYBALANCE=PayAmount` → Commit → `StatementProcessBill(ใหม่,+1)` จองยอดบนตารางหนี้ (`bcprocess.cpp:886-941` `PayBillAmount`, `StatementState`) | tx เดียว: lock หนี้ → คำนวณยอดที่จ่ายได้สด (§4.2) → เขียนหัว+บรรทัด → event | Champ จองแบบเก็บค่าบนตารางหนี้และประมวลผลหลัง Commit (ไม่ atomic); BC **คำนวณยอดจองสด** ใน tx ACID เดียว (ไม่มีค่าค้าง, กฎ PostgreSQL sync ACID) |
| ตรวจก่อนลบ/ยกเลิก | `ConfirmDeleteData :593-621`: ไม่มีเลขที่ → ปิดงวด → อนุมัติแล้ว `CANCEL_DELETE_CONFIRM` → `BillStatus!=0` → `CANNOT_DELETE_STATEMENT` "ไม่สามารถลบได้ เนื่องจากจ่ายชำระหนี้แล้ว" | §4.3 ข้อ 11, 15 | - |
| ลบ/ยกเลิก | `OnDeleteData :523-555`: `StatementProcessBill(-1)` คืนยอดจอง → DELETE จริง | ยกเลิกระดับหัว + เหตุผล + event (แถวยังอยู่) ยอดจองหายเองเพราะคำนวณสด | ลบจริง → ยกเลิกระดับหัว ตามคำตัดสินลุงจืด "ERP cancel stays header-level" + audit trail |
| PV ตรวจก่อนบันทึก | `ConfirmSaveData :425-539`: ยอดว่าง → ปิดงวด → ยกเลิกแล้ว → อนุมัติแล้ว → ผ่าน GL แล้ว → เลขที่ → เจ้าหนี้ → ไม่มีบรรทัด → `IsMoneyOk` (`AMOUNT_NOT_EQUAL_PAYMENT`) → (แก้ไข) เช็ค/บัตรถูกอ้าง → เลขซ้ำ | W5 คงลำดับนี้ เพิ่มการตรวจยอดคงเหลือในใบเตรียมจ่าย | - |
| PV บันทึก | `OnSaveData :804-1063`: (แก้ไข) คืนผลเดิมก่อน `:810-816` → หัว → StartTransaction → ลบ/เขียนบรรทัด (`PayBillNo` `:971`) → Commit `:1008-1012` → ภาษีซื้อ/WHT/รับจ่ายเงิน/มัดจำ → GL `ProcessConditionGroup9` หรือมือ `:1029-1037` → `PaymentProcessAP` + `PaymentProcessBill` `:1039-1040` → `BCAP.LastPayDate` `:1046` | W5: tx เดียว: (แก้ไข) `releasePosting` คืนผลเดิม → เขียน doc/docdetail → GL → settlement ต่อบรรทัด (§4.5) | Champ ประมวลผลหลัง Commit; BC ทำใน tx เดียว |
| PV ตัดหนี้ | `PaymentProcessBill` (`bcprocess.cpp:348-470`): ต่อบรรทัด `BillBalance -= PayAmount×mark` บนตารางตาม BillType → ถ้ามี PaybillNo เรียก `PaymentProcessStatement` (`:490-515`) ลด `PayBalance` บรรทัดใบต้นทาง แล้วตั้ง `BillStatus`=1 เมื่อไม่มีบรรทัดค้าง | settlement ต่อบรรทัด; `used`/`billstatus` ของใบเตรียมจ่ายคำนวณจาก settlement ที่ยังมีผล | เก็บค่า → คำนวณสด |
| PV ลบ | `OnDeleteData :593-634`: `OnProcessEarnest(-1)`, `PaymentProcessAP(-1)`, `PaymentProcessBill(1)` คืนยอด → DELETE หัว/บรรทัด → ลบภาษีซื้อ/WHT/รับจ่ายเงิน/GL source 9 | ยกเลิก PV → `releasePosting` → `cleanSubledger` ลบ settlement ทุกบรรทัด → ใบเตรียมจ่ายกลับเป็น "ยังไม่จ่าย" เอง | ลบจริง → ยกเลิกระดับหัว (เหมือนเดิมของ ERP) |
| ลงบัญชีใบเตรียมจ่าย | GL source 52 เขียน `bctrans_tmp` (จำลอง) `gltlsprocess.cpp:5941-5960` | ไม่ลง GL | ไม่ทำ GL จำลอง (ไม่มีผลต่อบัญชี) |

## 2. สถานะ BC ปัจจุบัน

| เรื่อง | หลักฐาน | ผล |
|---|---|---|
| จอเตรียมจ่าย | `frontend/src/components/ap-payment-prepare-screen.tsx` (808 บรรทัด) render ที่ `frontend/src/app/menu/main-menu-screen.tsx:2377-2378` | mock ล้วน ไม่เรียก API มีฟิลด์ที่ Champ ไม่มี (whtRate `:29`, paymentMethod `:44`, bankAccount `:45`, สถานะอนุมัติ `:47`) ข้อมูลตั้งต้นสมมติ `:58` และใช้ `isTh ?` ราว 45 จุด ซึ่งผิดกฎ i18n |
| ถูกนับว่าเสร็จ | `frontend/src/lib/menu-screen-status.ts:26` | ผิดกติกา "UI-only facade = รอพัฒนา" |
| อัปเดต 2026-10-11 (ตรวจทาน) | `git status`: `ap-payment-prepare-screen.tsx/.test.ts` ถูกลบ, `menu-screen-status.ts` และ `main-menu-screen.tsx` ถูกแก้ และไม่มีคำว่า `payment-prepare` เหลือ (ยังไม่ commit, งานอื่น) | W0 น่าจะเสร็จในงานคู่ขนาน — ตรวจ commit ก่อนเริ่ม W0 ถ้ามีแล้วข้าม |
| PV | doctype `pay` = 29 "PV" (`backend/internal/doctype/doctype.go:90`), config workbench สินค้า (`frontend/src/lib/erp-transaction.ts:335-354`) | ตัดหนี้ได้ **ใบเดียว** ผ่านหัว `DocRef` (`posting.go` `writeSubledger` ราว `:430-465`, settlement id `SET-AP-<docno>`, ยอด = ยอดรวมทั้ง PV) |
| ยอดที่ PV ลงบัญชี | `posting.go` `postingAmounts` (`:153`) | มีแค่ยอดรวม/VAT/มูลค่าสินค้า **ยังไม่มี WHT, ส่วนลด, เช็ค/โอน** (F6/F7 ของ Champ) |
| ตัวตรวจ settlement ของ GL | `generalledger/subledger_allocations.go:159` | ผ่านเฉพาะ `debt.Side==1 && payment.Side==2` + สกุลเงินตรงกัน; คู่จาก ERP (หนี้ side 2, จ่าย side 1, payload ไม่มี `currency_code`) **ไม่มีวันผ่าน** จึงใช้ตรงๆ กับ PV ไม่ได้ |
| ลบหนี้เมื่อแก้/ยกเลิก ERP | `cleanSubledger` (`posting.go:505-530`) เรียกจาก `releasePosting` (แก้ไข `save.go:124`, ลบ `save.go:319`, ยกเลิก `service.go:504`) | **DELETE จริง** ทั้ง settlement และ `gl_subledger_documents` ของตัวเอง (ตรวจแค่ settlement/allocation อื่น); ฝั่ง GL (`subledger.go:371`) ลบเฉพาะเอกสารที่ไม่มี allocation/settlement |
| เงื่อนไขที่ ERP สร้างหนี้ในบัญชีย่อย | `posting.go:224` (`glStatusNoMapping` คืนก่อนเขียน), `writeSubledger` (ข้ามเมื่อบัญชีคุมเป็นเงินสด `:397-401` และเมื่อไม่พบคู่ค้าใน `gl_subledger_partners`) | ต้องมีครบ 3 ข้อ: (1) mapping ลงบัญชีของประเภทเอกสาร (2) เจ้าหนี้มีใน `gl_subledger_partners` (3) บัญชีคุมไม่ใช่เงินสด — ขาดข้อใด ใบซื้อจะไม่โผล่ในรายการค้างจ่าย |
| ปิดเอกสารที่ถูกอ้าง | `backend/internal/transaction/httpapi/save.go:233-237` | `UPDATE doc SET isref=true, isclosed=true` ทำให้จ่ายบางส่วนแล้วเอกสารหนี้ถูกปิดทั้งใบ |
| เงินใน model | `backend/internal/transaction/httpapi/models.go` (float64 11 จุด เช่น `:14,:16,:17,:35`) | ขัด Rule 20 ฝั่ง JSON/Go (ฝั่ง DB เป็น numeric) |
| บัญชีย่อยเจ้าหนี้ | `backend/internal/generalledger/subledger.sql:11-36` (documents/allocations/settlements numeric(34,8)), audit `:59-64` | ใช้รองรับได้ทั้งเอกสารจาก GL และ ERP |
| เอกสาร ERP ที่เขียนบัญชีย่อย | `posting.go` `subledgerSpecs` (`:360`) | ฝั่ง AP มีแค่ `purchase`(หนี้, side 2), `purchase-return`, `pay` ส่วน `creditorbeginningbalance`(53), `apotherdebt`(406), `apdebitnote`(404), `apcreditnote`(405) ไม่เขียน จึงยังเตรียมจ่ายไม่ได้ |
| side ไม่ตรงกัน | GL เก็บ `side`=balance_side (1=หนี้); ERP เก็บด้านบัญชีคุม (AP ซื้อ=2, จ่าย=1) `posting.go` `subledgerSpecs` | query ต้อง normalize แบบที่ `goapi/handlers/debt_report.go` ใช้ (`payload ? 'balance_side'`) |
| ฐานยอดหนี้ | GL: posted allocations (`subledger_queries.go:14-33,44`); ERP: `d.amount` (ไม่มี allocation) | ต้องใช้สูตรรวมสองแบบ (§4.2) |
| payload key | ERP `doc_no/doc_date/due_date` (`posting.go` `writeSubledger`), GL `document_no/document_date` | ต้องใช้ COALESCE (ดู `debt_report.go` `buildAgingQuery`) |
| วันครบกำหนด | `posting.go` `writeSubledger` `payload["due_date"] = date` (ราว `:415`); `save.go` ไม่เขียน `doc.perioddatetime` (อ่าน/แก้แค่ใน `delivery_dates.go:47,146,164`) | C2 ยังขาด |
| วันเครดิตเจ้าหนี้ | ทะเบียนเจ้าหนี้ ERP (`system-setting-screens.ts` ฟิลด์ `creditday`) ไม่มี backend จริง; `gl_subledger_partners` (`subledger.sql:3-6`, `subledger_types.go:18-30`) ไม่มีวันเครดิต | ต้องเพิ่ม `credit_days` ใน payload ของคู่ค้า |
| มิติ | `generalledger/dimensions.go:20-21` มีแค่ departments/projects | ยังไม่มี master การจัดสรร/กลุ่มของบิล (audit D2) |

## 3. Worklist อย่างเดียว vs สร้าง PV+เช็คในขั้นเดียว: เลือก "เอกสารเตรียมจ่าย + PV ดึง (Ctrl+F3)"

| ตัวเลือก | ข้อดี | ข้อเสีย |
|---|---|---|
| A. worklist อย่างเดียว (ไม่เก็บเอกสาร) | ง่ายที่สุด | ไม่ตรง Champ (ไม่มี DocNo, ค้นหา Find/Prev/Next ไม่ได้, จองยอดไม่ได้) |
| **B. เอกสาร + PV ดึง (แนะนำ)** | ตรง Champ (BCPaymentTemp + RefType=1 + Ctrl+F3); WHT/เช็ค/GL อยู่ที่ PV จุดเดียวตาม ม.50; ใบเตรียมจ่ายไม่แตะ GL จึงยกเลิกได้ง่าย | ต้องทำ PV ให้ตัดหนี้ได้หลายใบต่อบรรทัด (W5) |
| C. สร้าง PV+เช็คทันที | ขั้นตอนสั้น | ไม่มีใน Champ (ผิด Strict Parity); สร้างรายการ GL/เช็คก่อนผู้มีอำนาจตัดสินใจ; ยกเลิกต้องกลับรายการ GL; จอเช็คเตรียมจ่ายเป็นคนละฟังก์ชัน |

## 4. การออกแบบ

### 4.1 Schema (ensureSchema `CREATE/ALTER ... IF NOT EXISTS`, ฐานข้อมูลของ holding เดียวกับ `doc`)

```sql
CREATE TABLE IF NOT EXISTS ap_payment_prepare (
 company text NOT NULL,
 docno text NOT NULL CHECK (length(btrim(docno)) > 0),
 paydate date NOT NULL,                       -- วันที่จ่ายชำระหนี้
 partner_code text NOT NULL,                  -- รหัสเจ้าหนี้ (gl_subledger_partners.code)
 branch_code text NOT NULL DEFAULT '',
 remark text NOT NULL DEFAULT '',
 projectcode text NOT NULL DEFAULT '',
 allocatecode text NOT NULL DEFAULT '',
 billgroupcode text NOT NULL DEFAULT '',
 departmentcode text NOT NULL DEFAULT '',
 total_amount numeric(18,2) NOT NULL DEFAULT 0,      -- รวมเป็นเงิน (สุทธิตามเครื่องหมาย BillType)
 total_home_amount numeric(18,2) NOT NULL DEFAULT 0, -- รวมเป็นเงิน(บาท)
 discount_amount numeric(18,2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
 net_amount numeric(18,2) NOT NULL DEFAULT 0,        -- ยอดจ่าย
 exchange_profit numeric(18,2) NOT NULL DEFAULT 0,   -- กำไร(ขาดทุน) (THB อย่างเดียว = 0)
 iscancel boolean NOT NULL DEFAULT false,
 cancel_reason text NOT NULL DEFAULT '',
 cancelled_at timestamptz, cancelled_by text NOT NULL DEFAULT '',
 version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
 created_at timestamptz NOT NULL, created_by text NOT NULL,
 modified_at timestamptz NOT NULL, modified_by text NOT NULL,
 PRIMARY KEY (company, docno),
 CHECK (NOT iscancel OR length(btrim(cancel_reason)) > 0)
);
CREATE INDEX IF NOT EXISTS ap_payment_prepare_partner ON ap_payment_prepare (company, partner_code, paydate);

CREATE TABLE IF NOT EXISTS ap_payment_prepare_lines (
 company text NOT NULL,
 docno text NOT NULL,
 linenumber int NOT NULL CHECK (linenumber > 0),
 debt_document_id text NOT NULL,              -- gl_subledger_documents.id (ไม่มี FK — ดูหมายเหตุใต้ SQL)
 document_no text NOT NULL,                   -- snapshot เลขที่เอกสาร (แสดงผล/ให้ PV อ้าง แม้หนี้ถูกลบ)
 document_date date NOT NULL,                 -- snapshot ลงวันที่
 debt_amount numeric(18,2) NOT NULL,          -- snapshot ยอดหนี้ ณ เวลาบันทึก
 billtype smallint NOT NULL,                  -- S (รหัส Champ ตามตาราง §4.5.3)
 sign smallint NOT NULL DEFAULT 1 CHECK (sign IN (1,-1)),
 description text NOT NULL DEFAULT '',        -- รายวัน/คำอธิบาย
 invbalance numeric(18,2) NOT NULL,           -- มูลค่าคงเหลือ ณ เวลาบันทึก (snapshot)
 payamount numeric(18,2) NOT NULL CHECK (payamount > 0), -- ยอดตัดจ่าย
 PRIMARY KEY (company, docno, linenumber),
 UNIQUE (company, docno, debt_document_id),
 FOREIGN KEY (company, docno) REFERENCES ap_payment_prepare (company, docno)
);
CREATE INDEX IF NOT EXISTS ap_payment_prepare_lines_debt ON ap_payment_prepare_lines (company, debt_document_id);

CREATE TABLE IF NOT EXISTS ap_payment_prepare_events (   -- audit append-only
 event_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 company text NOT NULL, docno text NOT NULL, action text NOT NULL,  -- create|update|cancel
 actor text NOT NULL, occurred_at timestamptz NOT NULL, payload jsonb NOT NULL  -- before/after + reason
);
CREATE INDEX IF NOT EXISTS ap_payment_prepare_events_doc ON ap_payment_prepare_events (company, docno, event_no);

-- PV อ้างใบเตรียมจ่าย (ชื่อตาม Champ BCPaymentSub.PaybillNo/RefType) — แก้ใน transaction/httpapi/schema.go
ALTER TABLE docdetail ADD COLUMN IF NOT EXISTS paybillno text NOT NULL DEFAULT '';
ALTER TABLE docdetail ADD COLUMN IF NOT EXISTS reftype smallint NOT NULL DEFAULT 0 CHECK (reftype IN (0,1,2));   -- 0=ไม่อ้างอิง 1=ใบเตรียมจ่าย 2=ใบรับวางบิล (BC ต่างจากเลข Champ โดยตั้งใจ)
ALTER TABLE docdetail ADD COLUMN IF NOT EXISTS debt_document_id text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_docdetail_paybillno ON docdetail (businesscode, paybillno) WHERE paybillno <> '';
```

- ทุกชื่อเป็น lowercase snake_case
- เงินทุกช่องเป็น numeric(18,2) ให้ตรงกับ `docdetail.sumamount` (`schema.go` ตาราง docdetail) ไม่มี float
- ไม่มีคอลัมน์ `paystatus` ที่เก็บไว้ ใช้การคำนวณแทน (§4.2) เพื่อไม่ให้สถานะค้างเมื่อยกเลิก PV
- มีสถานะเดียวที่แสดง คือ `billstatus` 0=ยังไม่จ่าย, 1=จ่ายหมดแล้ว ตาม Champ
- **ไม่มี FK ไป `gl_subledger_documents`** เพราะ `cleanSubledger` (ERP แก้/ลบ/ยกเลิกใบซื้อ) DELETE แถวหนี้จริง FK จะทำให้ได้ error ดิบ 23503 แทนข้อความไทย
  - กันแทนด้วยการตรวจใน `cleanSubledger`: ถ้าหนี้ยังถูกจองโดยใบเตรียมจ่ายที่ยังไม่ยกเลิกและบรรทัดยังใช้ไม่หมด → คืน `gl_err_ap_prepare_reserved` (409) พร้อมเลขใบเตรียมจ่าย; ใบที่ยกเลิกแล้วหรือบรรทัดที่ใช้ครบแล้วไม่บล็อก
  - เทียบ Champ: ใบรับวางบิลล็อกการแก้/ลบเมื่อถูกใช้ (`apfrmapstatement1view.cpp:612-615, 662-665`) และยอดจองอยู่บนตารางหนี้ (`PrePayAmount`/`PayBillAmount`)
  - ฝั่งหนี้จาก GL ไม่ต้องตรวจเพิ่ม เพราะ `subledger.go:371` ไม่ลบเอกสารที่มี allocation อยู่แล้ว
  - snapshot `document_no/document_date/debt_amount` ทำให้ใบเตรียมจ่ายที่ยกเลิกแล้วยังแสดงได้แม้หนี้ถูกลบภายหลัง
- `reftype` ใช้ 0=ไม่อ้างอิง เพื่อให้บรรทัด PV เดิมและบรรทัดที่ไม่อ้างใบใด ไม่ถูกตีความเป็นใบรับวางบิล (Champ ใช้ 0=ใบรับวางบิล แต่โค้ดบันทึก PV ไม่เขียน RefType เลย `:982` จึงไม่มีข้อมูลที่ต้องตรงเลข)

### 4.2 ยอดที่จ่ายได้ (คำนวณใน PG, ไม่เก็บ)

ให้ export ฟังก์ชันเดียวใน `generalledger` (เช่น `gl.APOpenItems(ctx, tx, scope, partner, asOf, excludeDocNo)`) ที่ต่อจาก `subledgerReadCTE` (`subledger_queries.go:15-33`) และใช้สูตรยอดคงเหลือเดียวกับรายงานอายุหนี้ `buildAgingQuery` (`goapi/handlers/debt_report.go:202-275`): หนี้จาก ERP นับเมื่อรายการสร้างหนี้ผ่านบัญชีแล้ว, หนี้จาก GL นับจาก allocation ที่ผ่านบัญชี, settlement นับเฉพาะที่ผ่านบัญชีและ as-of และนับทั้งฝั่ง `debt_id` และ `payment_id`. ทุกที่ (จอเตรียมจ่าย, PV, `cleanSubledger`) ต้องเรียกฟังก์ชันนี้ ห้ามเขียนสูตรซ้ำ:

```sql
-- ใช้ลำดับ param ของ subledgerReadCTE: $1 company, $2 branch ('' = ทุกสาขา), $3 q (ILIKE, '' = ไม่กรอง), $4 as-of date
-- เพิ่ม: $5 partner_code, $6 docno ใบเตรียมจ่ายที่กำลังแก้ ('' = ไม่มี), $7 limit, $8 offset
<subledgerReadCTE>,
posted_alloc AS (
  SELECT a.document_id, SUM(a.amount) AS amt
  FROM active_allocations a JOIN posted_journals j ON j.id=a.journal_id GROUP BY 1
),
posted_settle AS (   -- นับทั้งสองขา เหมือน buildAgingQuery
  SELECT x.document_id, SUM(s.amount) AS amt
  FROM active_settlements s CROSS JOIN LATERAL (VALUES (s.debt_id),(s.payment_id)) x(document_id)
  GROUP BY 1
),
debt AS (
  SELECT d.id, d.amount, d.payload, (d.payload ? 'balance_side') AS from_gl, cj.id IS NOT NULL AS created_posted
  FROM gl_subledger_documents d
  LEFT JOIN posted_journals cj ON cj.id=d.created_journal_id
  WHERE d.company=$1 AND d.ledger='ap' AND d.partner_code=$5 AND ($2='' OR d.branch_code=$2)
    AND CASE WHEN d.payload ? 'balance_side' THEN d.side=1 ELSE d.side=2 END  -- normalize side (เฟสแรกเฉพาะฝั่งหนี้, D7)
    AND ($3='' OR d.id ILIKE $3 OR COALESCE(d.payload->>'doc_no', d.payload->>'document_no') ILIKE $3)
),
used AS (   -- ยอดที่ PV ใช้จากแต่ละบรรทัดเตรียมจ่าย = settlement ที่ยังมีผลซึ่ง payload ระบุใบเตรียมจ่าย/บรรทัด
  SELECT s.payload->>'prepare_docno' AS prepare_docno, s.debt_id, SUM(s.amount) AS amt
  FROM active_settlements s
  WHERE s.payload ? 'prepare_docno'
  GROUP BY 1,2
),
reserved AS (
  SELECT l.debt_document_id, SUM(GREATEST(l.payamount - COALESCE(u.amt,0), 0)) AS amt
  FROM ap_payment_prepare_lines l
  JOIN ap_payment_prepare p ON p.company=l.company AND p.docno=l.docno AND NOT p.iscancel
  LEFT JOIN used u ON u.prepare_docno=l.docno AND u.debt_id=l.debt_document_id
  WHERE l.company=$1 AND l.docno<>$6
  GROUP BY 1
),
open_items AS (
  SELECT d.*, (CASE WHEN d.from_gl THEN COALESCE(pa.amt,0) WHEN d.created_posted THEN d.amount ELSE 0 END)
              - COALESCE(ps.amt,0) AS remaining
  FROM debt d LEFT JOIN posted_alloc pa ON pa.document_id=d.id LEFT JOIN posted_settle ps ON ps.document_id=d.id
)
SELECT o.id,
       COALESCE(o.payload->>'doc_no', o.payload->>'document_no') AS document_no,
       COALESCE(NULLIF(o.payload->>'doc_date',''), o.payload->>'document_date')::date AS document_date,
       COALESCE(NULLIF(o.payload->>'due_date',''), NULLIF(o.payload->>'doc_date',''), o.payload->>'document_date')::date AS due_date,
       o.remaining AS remaining_amount,
       COALESCE(r.amt,0) AS reserved_amount,
       o.remaining - COALESCE(r.amt,0) AS available_amount
FROM open_items o LEFT JOIN reserved r ON r.debt_document_id=o.id
WHERE o.remaining > 0
ORDER BY due_date, document_date, document_no, o.id      -- o.id ปิดท้ายให้ลำดับคงที่เมื่อแบ่งหน้า
LIMIT $7 OFFSET $8;
```

- `used` มาจาก settlement ที่ยังมีผล (ไม่ใช่จาก `docdetail`) จึงตัดกรณี PV ที่ถูกลบแบบ soft delete (`doc.isdelete=true`, `save.go:328`), PV ที่ยกเลิก และ PV ที่บันทึกแล้วแต่ไม่เกิด settlement (ไม่มี mapping/ร่าง) ออกเองทั้งหมด
  - PV ต้องเขียน `prepare_docno` และ `prepare_linenumber` ลง payload ของ settlement ต่อบรรทัด (§4.5)
  - ข้อเสีย: PV ฉบับร่างที่ยังไม่ผ่านบัญชีจะยังไม่ลดยอดจอง → การตรวจของ PV ต้องนับ `docdetail` ของ PV ที่ยังไม่ลบ/ไม่ยกเลิกเพิ่มด้วย (`AND NOT COALESCE(h.isdelete,false) AND NOT COALESCE(h.iscancel,false)`) เพื่อกันจ่ายซ้ำระหว่างร่าง
- เมื่อ PV ผ่านบัญชี ยอด `posted_settle` จะเพิ่ม และ `used` ก็เพิ่มเท่ากัน ทำให้ `reserved` ลดลงเท่ากัน จึงไม่นับซ้ำ
- เมื่อยกเลิกหรือลบ PV settlement ถูกลบใน `cleanSubledger` ทั้งสองค่าจะคืนกลับเอง
- การรับค่า:
  - ยอดรับเป็น decimal string และต้องมีทศนิยมไม่เกิน 2 ตำแหน่ง
  - Go ใช้ `shopspring/decimal`
  - การเปรียบเทียบ `payamount ≤ available_amount` ทำใน SQL หรือด้วย decimal เท่านั้น
- การกันแย่งยอด: ก่อนคำนวณให้ `SELECT id FROM gl_subledger_documents WHERE company=$1 AND id = ANY($2) ORDER BY id FOR UPDATE` เพื่อให้สองคนที่เตรียมจ่ายหนี้ใบเดียวกันทำงานต่อกันทีละราย
- `billstatus` ของหัวเอกสาร = 1 เมื่อทุกบรรทัดมี `used ≥ payamount` ถ้าไม่เช่นนั้นเป็น 0 (คำนวณใน list/get)

### 4.3 กฎตรวจ (backend)

1. `docno` ต้องไม่ว่างและไม่ซ้ำ
2. `partner_code` ต้องมีใน `gl_subledger_partners` และเป็นเจ้าหนี้
3. ต้องมีอย่างน้อย 1 บรรทัด
4. หนี้ทุกบรรทัดต้องเป็นของเจ้าหนี้รายนี้ ledger `ap`
5. ห้ามมีหนี้ใบเดียวกันซ้ำในใบเตรียมจ่ายเดียวกัน
6. `payamount > 0` และ `≤ available_amount`
7. ส่วนลดต้องอยู่ในช่วง `0 ≤ discount_amount ≤ total_amount`
8. ยอดรวมทุกช่องคำนวณที่ backend
   - `total_amount = Σ sign×payamount`
   - `total_home_amount = total_amount` (THB)
   - `net_amount = total_amount − discount_amount`
   - ห้ามเชื่อยอดรวมที่ frontend ส่งมา
9. `projectcode` และ `departmentcode` (ถ้ากรอก) ต้องมีใน dimensions (`dimensions.go:20-21`)
10. `allocatecode` และ `billgroupcode` ยังรับไม่ได้จนกว่าจะมี master (audit D2, §7 D8) ระหว่างนี้ปิดช่องในจอ และ backend ต้องรับเฉพาะค่าว่าง
11. แก้ไขหรือยกเลิกได้เฉพาะใบที่ยังไม่ถูก PV ใช้ (`Σused = 0`) และยังไม่ยกเลิก
    - ตาม Champ ใบรับวางบิล (เอกสารจองยอดแบบเดียวกัน): แก้ไม่ได้เมื่อ `BillStatus!=0` → `CANNOT_UPDATE_STATEMENT` (`apfrmapstatement1view.cpp:662-665`), ลบไม่ได้ → `CANNOT_DELETE_STATEMENT` "ไม่สามารถลบได้ เนื่องจากจ่ายชำระหนี้แล้ว" (`:612-615`)
    - ต่างเล็กน้อยโดยตั้งใจ: Champ ใบรับวางบิลล็อกเมื่อ `BillStatus!=0` (ซึ่งตั้งเป็น 1 เมื่อจ่ายหมดเท่านั้น `bcprocess.cpp:490-515`) แต่ BC ล็อกตั้งแต่ถูกใช้บางส่วน เพราะ BC คำนวณยอดจองสด การแก้บรรทัดที่ถูกใช้ไปแล้วจะทำให้ `used > payamount` (ถ้าลุงจืดต้องการให้ตรง Champ เป๊ะ ให้ยอมแก้ได้แต่ห้ามลด payamount ต่ำกว่า used)
12. ยกเลิกต้องมีเหตุผล ปุ่ม Del ของ Champ ให้ทำเป็นยกเลิกระดับหัวเอกสาร ตามคำตัดสิน "ERP cancel stays header-level" แล้วบันทึกเหตุการณ์ลง `ap_payment_prepare_events`
13. ใช้ `version` แบบ optimistic lock สำหรับ PUT และ cancel
14. ไม่ลง GL และไม่แตะ `gl_journals`
15. ตรวจปิดงวดทั้งตอนบันทึกและยกเลิก ด้วย `paydate` (Champ ใบรับวางบิล `IsPeriodClose` `:602`, `:638`; PV `:436`, `:544`) ใช้ตัวตรวจปิดงวดตัวเดียวกับเอกสาร ERP อื่น (`generalledger/postgres_guards.go:43` `periodLocked` → key เดิม `gl_err_source_period_locked`, ดูการใช้ใน `transaction/httpapi/period_lock_test.go`) ไม่สร้าง key ใหม่
16. เตรียมจ่ายได้เฉพาะหนี้ที่ผ่านบัญชีแล้ว (สูตร §4.2) ตาม Champ ที่ดึงเฉพาะเอกสาร `BillBalance>0` และ `IsCancel=0`; ถ้าเจ้าหนี้ไม่มีหนี้ค้างเลย จอแสดง `erp_ap_prepare_no_posted_debt` อธิบายว่าต้องมี mapping ลงบัญชี, เจ้าหนี้ในทะเบียนบัญชีย่อย และใบซื้อแบบเชื่อที่ผ่านบัญชีแล้ว (§2)

### 4.4 REST API

- Backend อยู่ที่ `/ap-payment-prepare/...` เป็น **route group ของตัวเอง** (handler แยกใน `transaction/httpapi` หรือ package ใหม่) **ไม่ผ่าน `/transaction/:module`**
  - เหตุผล: `authorize` (`http.go`) เรียก `lookupDocType(module)` ก่อนเสมอ ถ้าไม่ใช่ doctype จะคืน `erp_err_invalid_module`; และ D2 ตั้งใจไม่ลงทะเบียน doctype
  - ห้ามเพิ่ม `"ap-payment-prepare"` ใน `ERP_TRANSACTION_MODULES` (`frontend/src/lib/erp-transaction-modules.ts`) เพราะ test `erp-transaction.test.ts:125-131` บังคับให้รายการนี้เท่ากับทะเบียน doctype (63 ประเภท) ทุกตัว
  - ตรวจสิทธิ์เองด้วย session + สิทธิ์เมนู `ap-payment-prepare` และ action `""/create/update/cancel` ตามกติกาเดียวกับ `<เมนู>` + `<เมนู>:<action>` (`http.go` ค่าคงที่ `actionRead..actionCancel`) โดยเรียกตัวช่วย session/สิทธิ์ตัวเดียวกับ `authorize` ไม่เขียนกฎสิทธิ์ใหม่
- Frontend เรียกผ่าน BFF ใหม่ `frontend/src/app/api/ap-payment-prepare/[...path]/route.ts` ที่มี allowlist path (`list`, `outstanding`, `<docno>`, `<docno>/outstanding`, `<docno>/cancel`, `<docno>/events`) และ allowlist query (`q, page, limit, fromdate, todate, partner, status`) แบบเดียวกับ BFF `erp-transaction`; ส่ง `Accept-Language` ต่อ และ relay `success:false` ตามเดิม

| Method + path | Request | Response `data` |
|---|---|---|
| GET `/list?q&fromdate&todate&partner&status&page&limit` | `status`: `open`/`paid`/`cancelled` | `[{docno, paydate, partner_code, partner_name, total_amount, discount_amount, net_amount, billstatus, iscancel, version}]`, `total` |
| GET `/outstanding?partner&todate&q&page&limit` | สร้างใบใหม่ | `[{debt_document_id, document_no, document_date, due_date, billtype, sign, description, remaining_amount, reserved_amount, available_amount}]` |
| GET `/:docno/outstanding?...` | แก้ใบเดิม (ไม่นับยอดจองของตัวเอง) | เหมือนแถวบน |
| GET `/:docno` | - | หัวเอกสาร + `lines[{linenumber, debt_document_id, document_no, document_date, billtype, description, invbalance, payamount, used_amount, open_amount}]` + ยอดรวม + `billstatus` + `version` |
| POST `/` | `{docno, paydate, partner_code, remark, projectcode, departmentcode, discount_amount:"0.00", lines:[{debt_document_id, payamount:"1500.00", description}]}` | เอกสารที่บันทึกแล้ว (201) |
| PUT `/:docno` | เหมือน POST + `version` | เอกสาร |
| POST `/:docno/cancel` | `{reason, version}` | `{docno, iscancel:true}` |
| GET `/:docno/events` | - | audit trail |

- เงินทุกช่องเป็น decimal string `"1234.50"` ไม่ใช่ JSON number
- รูปแบบผิดพลาดเป็น `{success:false, code, message, field, args}` ผ่าน `erpError(code, field, args...)` โดยข้อความมาจาก `languages.tsv` แถว `gl_err_<code>` (กติกาของงานที่กำลังแก้อยู่ใน transaction/httpapi) และรองรับ `Accept-Language`

| code (`gl_err_` + code) | HTTP | field |
|---|---|---|
| `ap_prepare_docno_required` | 400 | docno |
| `ap_prepare_doc_exists` | 409 | docno |
| `ap_prepare_not_found` | 404 | docno |
| `ap_prepare_creditor_required` / `ap_prepare_creditor_not_found` | 400 | partner_code |
| `ap_prepare_no_lines` | 400 | lines |
| `ap_prepare_line_duplicate` | 400 | lines[i].debt_document_id |
| `ap_prepare_debt_not_found` (หนี้ไม่ใช่ของเจ้าหนี้รายนี้/ไม่ใช่ ap) | 400 | lines[i] |
| `ap_prepare_amount_invalid` (≤0, เกิน 2 ตำแหน่ง, ไม่ใช่ตัวเลข) | 400 | lines[i].payamount |
| `ap_prepare_over_balance` args `{docno, available}` | 409 | lines[i].payamount |
| `ap_prepare_discount_invalid` | 400 | discount_amount |
| `ap_prepare_dimension_not_found` | 400 | projectcode/departmentcode |
| `ap_prepare_paid_locked` | 409 | - |
| `ap_prepare_cancelled_locked` | 409 | - |
| `ap_prepare_cancel_reason_required` | 400 | reason |
| `ap_prepare_version_conflict` | 409 | version |
| `ap_pay_prepare_creditor_mismatch` (PV ดึงใบของเจ้าหนี้อื่น) | 400 | paybillno |
| `ap_pay_prepare_over_open` (PV จ่ายเกินยอดที่เหลือในใบเตรียมจ่าย) | 409 | lines[i].sumamount |
| `ap_prepare_reserved` args `{debt_docno, prepare_docno}` (แก้/ลบ/ยกเลิกเอกสารหนี้ที่ยังถูกจอง — คืนจาก `cleanSubledger`) | 409 | - |
| `ap_pay_settle_mismatch` (PV ตัดหนี้ที่เจ้าหนี้/บัญชีคุม/สกุลเงินไม่ตรง หรือ side ไม่ใช่หนี้ — §4.5.1) | 400 | lines[i].debt_document_id |
| ปิดงวด → ใช้ key เดิม `source_period_locked` | 409 | paydate |
| (ไม่ใช่ error) `erp_ap_prepare_no_posted_debt` ข้อความ empty state ของรายการค้างจ่าย | - | - |
| `erp_due_before_doc` (วันครบกำหนด) | 400 | duedate |
| `erp_credit_days_invalid` (วันครบกำหนด, ต้องเป็นจำนวนเต็ม ≥0) | 400 | creditday |

### 4.5 PV ดึงใบเตรียมจ่าย (Ctrl+F3) และการตัดหนี้หลายใบ

1. หน้าจอ PV เพิ่ม action "ใบเตรียมจ่าย" (Ctrl+F3)
   - เปิดรายการ `GET /ap-payment-prepare/list?partner=<เจ้าหนี้ของ PV>&status=open` ให้เลือกได้หลายใบ
   - เพิ่มบรรทัดจาก `GET /:docno` เฉพาะบรรทัดที่ `open_amount > 0`
   - แต่ละบรรทัดตั้ง `docref=document_no`, `debt_document_id`, `paybillno=<docno ใบเตรียมจ่าย>`, `reftype=1`, `sumamount=open_amount`
2. PV ต้องเปลี่ยนจาก "หัว DocRef ใบเดียว" (`writeSubledger` ราว `:430-465`) เป็น **settlement ต่อบรรทัด** ตาม Champ `PaymentProcessBill` ที่ลดยอดทีละบรรทัดตาม BillType (`bcprocess.cpp:348-470`)
   - id `SET-AP-<pvdocno>-<linenumber>`
   - amount = `sumamount` (ยอดตัดจ่ายเต็มของบรรทัด ก่อนหัก WHT/ส่วนลด — §4.5.2)
   - debt = `debt_document_id`, payment = เอกสารจ่ายของ PV เอง (`AP-...` side 1)
   - payload ใส่ `prepare_docno`/`prepare_linenumber` เมื่อบรรทัดมาจากใบเตรียมจ่าย (ใช้คำนวณ `used` §4.2)
   - ตรวจด้วยตัวตรวจ §4.5.1 (**ไม่ใช่** ตัวตรวจ GL settlement เดิม เพราะตัวนั้นปฏิเสธคู่จาก ERP ทุกคู่)
   - ลบ path หัว `DocRef` เดิมในการเปลี่ยนแปลงเดียวกัน ตามกฎ Dev เดินหน้าอย่างเดียว
3. ใน tx เดียวกับการบันทึก PV:
   - lock หัวใบเตรียมจ่ายที่ถูกอ้าง (`SELECT ... FOR UPDATE`)
   - ตรวจว่า `Σ sumamount` ต่อ `(paybillno, debt_document_id)` ไม่เกิน `payamount − used` และเจ้าหนี้ตรงกัน
4. ยกเลิก/ลบ/แก้ PV ผ่าน `releasePosting` → `cleanSubledger` (`posting.go:505-530`, ปัจจุบันใช้ `DELETE` ไม่ใช่การกลับรายการ) ลบ settlement ทุกบรรทัด แล้ว `used` จะลดลงเอง ทำให้ใบเตรียมจ่ายกลับมาเป็น "ยังไม่จ่าย" — เทียบ Champ PV `OnDeleteData :593-634` ที่เรียก `PaymentProcessBill(1)` คืนยอดก่อนลบ
   - `cleanSubledger` ต้องเพิ่มการตรวจยอดจองของใบเตรียมจ่าย (§4.1) ก่อน DELETE หนี้ของเอกสารฝั่งหนี้
5. ห้ามปิดเอกสารหนี้ทั้งใบเมื่อถูกอ้าง (`save.go:233-237`) ในกรณีจ่ายบางส่วน
   - เอกสารหนี้ "ปิด" เมื่อ `remaining = 0` ซึ่งคำนวณจากบัญชีย่อย
6. เปลี่ยนเงินใน `models.go` ที่ PV ใช้ (`sumamount`, `totalamount` ฯลฯ) เป็น decimal string หรือ `decimal.Decimal` (Rule 20)
7. WHT (F6) และเช็ค/โอน (F7) เป็นหน้าที่ของ PV ตาม Champ ไม่อยู่ในใบเตรียมจ่าย — แต่ **PV ของ BC ยังไม่มี** (§2, §4.5.2) จึงไม่ใช่ "เหมือนเดิม"
   - ส่วนลดของใบเตรียมจ่ายไม่ย้ายไป PV อัตโนมัติ เพราะไม่พบหลักฐานใน Champ ดู §7

#### 4.5.1 ตัวตรวจ settlement ต่อบรรทัดของ PV (ตัวเดียว ใช้ทั้งหนี้จาก ERP และ GL)

- normalize side ตามแหล่ง: หนี้ = (payload มี `balance_side` → side 1) หรือ (หนี้จาก ERP ฝั่ง AP → side 2); เอกสารจ่ายของ PV = ฝั่งตรงข้าม
- สกุลเงิน: payload ที่ไม่มี `currency_code` (ERP ทุกใบตอนนี้) ถือเป็น `THB`; ต้องตรงกันทั้งสองขา
- เจ้าหนี้ (`partner_code`), ledger `ap` และบัญชีคุม (`control_account_code`) ต้องตรงกัน
- ยอด ≤ ยอดคงเหลือ as-of จากฟังก์ชันเดียวของ §4.2 (ไม่ใช่ `debtOutstanding` ที่ settlement ไม่ as-of)
- PV ของ ERP ตัดหนี้ที่บันทึกจาก GL ได้ (หลัง normalize) เพราะเจ้าหนี้รายเดียวมีหนี้ได้ทั้งสองแหล่ง และ Champ ตัดหนี้ทุกประเภทจากจอเดียว (`PaymentProcessBill` เลือกตารางตาม BillType)
- ผิดเงื่อนไข → `gl_err_ap_pay_settle_mismatch`
- test: (1) PV(ERP) ตัด PUR(ERP) (2) PV(ERP) ตัดหนี้จาก GL `balance_side=1` (3) จับคู่ side ผิด (เอาเอกสารจ่ายมาเป็นหนี้) ต้องถูกปฏิเสธ (4) สกุลเงินว่าง vs `THB` ผ่าน

#### 4.5.2 WHT, ส่วนลด, เช็ค/โอนของ PV (งานแยก — ใบเตรียมจ่ายไม่ต้องรอ แต่ "จ่ายจริงครบวงจร" ต้องรอ)

- ตอนนี้ `postingAmounts` (`posting.go:153`) มีแค่ยอดรวม/VAT/มูลค่าสินค้า
- งานแยก (ไม่อยู่ใน W5):
  - B-1 WHT ของ PV (F6) — ต้องตรวจ ม.50 / ม.3 เตรส กับ rd.go.th และลง `docs/kms/21` ก่อน (§1.4)
  - B-2 ส่วนลดรับของ PV
  - B-3 บันทึกจ่ายเงิน/เช็ค/โอนของ PV (F7)
- invariant ที่ W5 ต้องรักษาไว้ตั้งแต่วันแรก (Champ `ConfirmSaveData` `IsMoneyOk` + `OnSaveData` net = total − discount − WHT):
  - settlement ต่อบรรทัดหนี้ = `sumamount` เต็ม (ยอดที่ตัดหนี้)
  - เงินออกจริงของ PV = Σ`sumamount` − ส่วนลด − WHT
  - test: PV ที่มี WHT ต้องตัดหนี้ที่เตรียมจ่ายไว้ได้หมด (remaining = 0, ใบเตรียมจ่าย billstatus = 1) ไม่ใช่เหลือยอดเท่า WHT

#### 4.5.3 BillType (S) และเครื่องหมายของเอกสารหนี้ BC

backend กำหนดจาก **ประเภทเอกสาร** (BC doctype ของ ERP หรือ `document_kind` ของ GL) เท่านั้น — ห้ามดูจากรหัสบัญชี (กฎ 2026-09-24). คำอธิบายบรรทัดเป็น key ชื่อประเภทเอกสารที่ backend ส่งกลับ (แบบ MyDesc ของ Champ)

| แหล่ง | ประเภท | BillType | sign | Champ |
|---|---|---|---|---|
| ERP | `purchase` | 0 | + | BCApInvoice |
| ERP | `creditorbeginningbalance` | 1 | + | BCApInvBalance |
| ERP | `purchase-return`, `apcreditnote` | 2 | − | BCStkRefund |
| ERP | `apdebitnote` | 3 | + | BCDebitNote2 |
| ERP | `apotherdebt` | 4 | + | BCAPOtherDebt |
| ERP | `advancepayment` | 8 | − | BCAPDepositSpecial (จ่ายเงินล่วงหน้า) |
| GL | kind 5 ยอดยกมา | 1 | + | BCApInvBalance |
| GL | kind 3 เพิ่มหนี้ | 3 | + | BCDebitNote2 |
| GL | kind 4 ลดหนี้ | 2 | − | BCStkRefund |
| GL | kind 1 ตั้งหนี้ | 4 | + | **สมมติฐาน**: หนี้ที่ตั้งจาก GL ≈ ตั้งเจ้าหนี้อื่นๆ (ถามลุงจืด §7) |
| GL | kind 2 ผลชำระ | - | - | ไม่ใช่หนี้ ไม่แสดง |

- 5 (มัดจำ TOHOME) ไม่ใช้; 7 (ใบรับวางบิล) ใช้เมื่อลุงจืดอนุมัติให้ดึงใบรับวางบิล (§7)
- ฝั่ง − (2, 8) เปิดในใบเตรียมจ่ายหลัง PV ตัดหนี้กับฝั่งลดได้ (D7 ตาม parity: Champ ให้ฝั่งลดอยู่ในใบเดียวกัน)

### 4.6 วันครบกำหนดจากจำนวนวันเครดิต (คำถามประกอบ C2 — สนับสนุนรายงานอายุหนี้ที่ตั้งช่วงได้แล้วใน `387aa183`)

1. ทะเบียนเจ้าหนี้: เพิ่ม `credit_days` (int ≥0, ไม่บังคับ) ใน payload ของ `gl_subledger_partners`
   - ตาม pattern ช่องเสริมใน payload (`subledger_types.go:24-29`) และตรวจใน `validatePartner`
   - เป็น parity กับ `BCAP.CreditDay` ไม่ใช่ของใหม่
2. หัวเอกสาร ERP ฝั่งหนี้ (`purchase`, `creditorbeginningbalance`, `apotherdebt`, `apdebitnote`) รับ `creditday` (int) และ `duedate` (date)
   - ค่าตั้งต้น `creditday` = `credit_days` ของเจ้าหนี้ และ `duedate = docdate + creditday`
   - ถ้า `duedate` ว่างตอนบันทึก ให้ใช้ `docdate` (Champ `ChmFrmGoodsPurchase.cpp:2221-2227` ทำเฉพาะประเภทเชื่อ 1/3; BC ไม่มีประเภทเงินสดในเอกสารหนี้ จึงใช้กับทุกใบที่เกิดหนี้)
   - ถ้า `duedate < docdate` ให้คืน `erp_due_before_doc`
3. `save.go` เขียน `doc.perioddatetime = duedate`
   - `writeSubledger` ใช้ `payload["due_date"] = duedate` แทน `date` (ราว `posting.go:415`)
   - `delivery_dates.go:143-166` ที่ซิงก์อยู่แล้วยังใช้ได้
   - ลำดับตาม Champ: วันเครดิตตั้งต้นจากเจ้าหนี้เมื่อเลือกเจ้าหนี้ (`ChmFrmGoodsPurchasePage1.cpp:813-825`) → แก้ได้ → คำนวณวันครบกำหนดใหม่ทุกครั้งที่แก้วันที่/วันเครดิต → ว่างตอนบันทึกใช้วันที่เอกสาร
4. nDocType 0/2 ของ Champ = ประเภทเงินสด (ซื้อสด/บริการสด, `ApRepApInvBalanceView.cpp:188-193`) — **ปิดคำถามแล้ว ไม่ต้องถาม**: ประเภทเงินสดไม่เกิดหนี้ ใน BC `writeSubledger` ข้ามเอกสารที่บัญชีคุมจาก mapping (`controlAccount(mapping.Rules, ...)`) เป็นบัญชีเงินสด (`:397-401`) จึงไม่เกิดหนี้และวันครบกำหนดไม่ไปถึงรายงานอายุหนี้หรือใบเตรียมจ่าย; ใบที่เกิดหนี้ใช้กติกาวันเครดิตของประเภทเชื่อทั้งหมด
   - การที่จอซื้อของ BC ไม่มีช่อง "ประเภทซื้อ สด/เชื่อ" เป็นช่องว่าง parity ของจอซื้อ แยกเป็นงานของจอซื้อ ไม่อยู่ในเอกสารนี้
5. ฝั่ง AR (`sale-invoice`) ใช้ช่องเดียวกันได้ แต่ต้องตรวจ Champ ฝั่งขาย (`BCAR`) ก่อน จึงอยู่นอกขอบเขตเอกสารนี้

### 4.7 Frontend

- **ทำทันที (ตามคำตัดสิน "UI-only facade = รอพัฒนา")**
  - เอา `/ap/payment-prepare` ออกจาก `menu-screen-status.ts:26` แล้วเมนูจะแสดง "รอพัฒนา" (`menu_pending_development`)
  - ถอด branch ที่ `main-menu-screen.tsx:2377-2378`
  - ลบ `ap-payment-prepare-screen.tsx` และ `ap-payment-prepare-screen.test.ts` (dead code ของ mock)
  - ต้องรอให้ workflow ที่กำลังแก้ `menu-screen-status.ts` เสร็จก่อน (ตรวจทาน 2026-10-11: working tree ทำส่วนนี้แล้วแต่ยังไม่ commit — §2)
- **จอจริง** `frontend/src/components/ap-payment-prepare-screen.tsx` (เขียนใหม่)
  - รับ prop `language` และใช้ `useBackendLanguage(language, backendUrl)` กับ `backendText`
  - ห้ามใช้ `isTh ?`
  - layout ตาม §1.1 ทุกฟิลด์ตามลำดับ Champ
  - ตาราง:
    - ช่องเลขที่เอกสารพิมพ์ได้และมีปุ่มค้น (เปิดรายการ `/outstanding`)
    - ลงวันที่และมูลค่าคงเหลืออ่านอย่างเดียว
    - ยอดตัดจ่ายตั้งต้นเท่ากับ `available_amount`
    - S แสดงรหัส BillType
  - ท้ายจอ: รวมเป็นเงิน(บาท), กำไร(ขาดทุน) และยอดจ่ายเป็น read-only จากผลที่ backend ส่งกลับ
    - frontend รวมยอดเพื่อแสดงผลชั่วคราวได้ด้วย decimal string (`decimal.js`) แต่ค่าจริงเอาจาก response
  - การจัดสรรและกลุ่มของบิลแสดงแต่ปิดไว้จนกว่า master (audit D2) จะเสร็จ (§7 D8)
  - ปุ่ม: New, Save (enabled ทันทีที่มีการแก้ และมี dirty guard), ยกเลิก (Del) พร้อม dialog ไทยให้กรอกเหตุผล, Find, Prev, Next
    - Info, F1 Key และ Recur อยู่นอกรอบนี้ (§7 D6)
  - ทำตามมาตรฐาน `ui-scale-polish`: ตัวอักษร ≥0.9rem, ปุ่ม ≥2.6em, ทดสอบทั้ง light และ dark
  - เมื่อเสร็จแล้วใส่ `/ap/payment-prepare` กลับเข้า `CUSTOM_MENU_SCREEN_ROUTES`
- **PV**: เพิ่ม action "ใบเตรียมจ่าย" (Ctrl+F3) ในจอ `pay` (`erp-transaction.ts:335-354`) และแสดง hint `Ctrl+F3=ใบเตรียมจ่าย`
- **ทะเบียนคู่ค้า**: เพิ่มช่อง "จำนวนวันเครดิต" (ไม่บังคับ, key เดิม `credit_day`)
- **หัวเอกสารหนี้ ERP**: เพิ่มช่อง "เครดิต (วัน)" (key เดิม `credit_days`) และ "วันครบกำหนด" (key เดิม `gl_detail_documents_due_date`)
- **รายการค้างจ่ายว่าง**: แสดงข้อความ `erp_ap_prepare_no_posted_debt` (§4.3 ข้อ 16)

### 4.8 Language keys (13 คอลัมน์ `key th en cn ja km ko lo my vi ms id fil`; คอลัมน์ cn..fil ใส่ค่าอังกฤษเป็น placeholder)

ใช้ key เดิมได้:

- `docno`:1315 "เลขที่เอกสาร"
- `creditor_code`:1039
- `remark`:3210 "หมายเหตุ"
- `balance_amount`:398 "มูลค่าคงเหลือ"
- `project`:2788 "โครงการ"
- `department_code`:1198 "รหัสแผนก"
- `doc_discount_amount`:1288 "ยอดส่วนลด"
- `menu_pending_development`:4656 "รอพัฒนา"
- `ap_payment_prepare`:8733
- `credit_day`:1030 "จำนวนวันเครดิต" (ทะเบียนเจ้าหนี้), `credit_days`:1031 "เครดิต (วัน)" (หัวเอกสารหนี้)
- `gl_detail_documents_due_date`:7550 "วันครบกำหนด" (หัวเอกสารหนี้)
- `gl_err_source_period_locked` (ปิดงวด)
- ห้ามใช้ `due_date`:1359 เพราะ th คือ "วันที่สั่งจ่าย" ความหมายไม่ตรง

ห้ามใช้ key ต่อไปนี้ เพราะถ้อยคำไม่ตรง Champ: `profit_loss` "งบกำไรขาดทุน", `pay_amount` "ชำระเงิน", `discount_amount` "มูลค่าส่วนลด", `doc_date` "เอกสารวันที่"

key ใหม่ (th / en):

| key | th | en |
|---|---|---|
| ap_prepare_pay_date | วันที่จ่ายชำระหนี้ | Payment Date |
| ap_prepare_dated | ลงวันที่ | Dated |
| ap_prepare_day_explain | รายวัน/คำอธิบาย | Journal/Description |
| ap_prepare_cut_pay_amount | ยอดตัดจ่าย | Amount Applied |
| ap_prepare_col_s | S | S |
| ap_prepare_allocate | การจัดสรร | Allocation |
| ap_prepare_bill_group | กลุ่มของบิล | Bill Group |
| ap_prepare_exchange_profit | กำไร(ขาดทุน) | Gain (Loss) |
| ap_prepare_total_home | รวมเป็นเงิน(บาท) | Total (THB) |
| ap_prepare_total | รวมเป็นเงิน | Total |
| ap_prepare_net_pay | ยอดจ่าย | Net Payment |
| ap_prepare_outstanding_lookup | เอกสารค้างจ่าย | Outstanding Documents |
| ap_prepare_status_unpaid | ยังไม่จ่าย | Unpaid |
| ap_prepare_status_paid | จ่ายหมดแล้ว | Fully Paid |
| ap_prepare_cancel_reason | เหตุผลการยกเลิก | Cancellation Reason |
| ap_pay_pull_prepare | ใบเตรียมจ่าย | Payment Preparation |
| ap_pay_pull_prepare_hint | Ctrl+F3=ใบเตรียมจ่าย | Ctrl+F3=Payment Preparation |
| erp_ap_prepare_no_posted_debt | เจ้าหนี้รายนี้ไม่มีหนี้ค้างที่ผ่านบัญชีแล้ว ตรวจว่าใบซื้อผ่านบัญชีแล้ว และตั้งผังลงบัญชีกับทะเบียนเจ้าหนี้ไว้แล้ว | This creditor has no posted open debt; check posting setup and creditor registration |
| ap_prepare_billtype_purchase / _beginning / _creditnote / _debitnote / _otherdebt / _advance | ใบรับสินค้า / ยอดยกมาต้นปี / ใบลดหนี้ / ใบเพิ่มหนี้ / เจ้าหนี้อื่นๆ / จ่ายเงินล่วงหน้า (ถ้าไม่มี key ชื่อประเภทเอกสารเดิมที่ตรงคำ — ตรวจ tsv ก่อนเพิ่ม) | Goods Receipt / Opening Balance / Credit Note / Debit Note / Other Payable / Advance Payment |
| gl_err_ap_prepare_reserved | เอกสาร {0} ถูกจองในใบเตรียมจ่าย {1} กรุณายกเลิกใบเตรียมจ่ายหรือลดยอดก่อน | Document {0} is reserved by payment preparation {1}; cancel or reduce it first |
| gl_err_ap_pay_settle_mismatch | เอกสาร {0} ไม่ใช่หนี้ของเจ้าหนี้รายนี้ หรือบัญชีคุม/สกุลเงินไม่ตรงกับใบจ่าย | Document {0} does not match this payment's creditor, control account or currency |
| gl_err_ap_prepare_docno_required | กรุณากรอกเลขที่เอกสาร | Please enter the document number |
| gl_err_ap_prepare_doc_exists | เลขที่เอกสารนี้มีอยู่แล้ว กรุณาใช้เลขอื่น | This document number already exists |
| gl_err_ap_prepare_not_found | ไม่พบใบเตรียมจ่ายนี้ | Payment preparation not found |
| gl_err_ap_prepare_creditor_required | กรุณาเลือกรหัสเจ้าหนี้ | Please select a creditor |
| gl_err_ap_prepare_creditor_not_found | ไม่พบรหัสเจ้าหนี้นี้ในทะเบียน | Creditor not found |
| gl_err_ap_prepare_no_lines | กรุณาเลือกเอกสารที่จะจ่ายอย่างน้อย 1 รายการ | Select at least one document |
| gl_err_ap_prepare_line_duplicate | เอกสาร {0} ถูกเลือกซ้ำ กรุณาลบบรรทัดที่ซ้ำ | Document {0} is selected twice |
| gl_err_ap_prepare_debt_not_found | เอกสาร {0} ไม่ใช่หนี้ค้างของเจ้าหนี้รายนี้ | Document {0} is not an open debt of this creditor |
| gl_err_ap_prepare_amount_invalid | ยอดตัดจ่ายต้องมากกว่า 0 และมีทศนิยมไม่เกิน 2 ตำแหน่ง | Amount must be above 0 with at most 2 decimals |
| gl_err_ap_prepare_over_balance | ยอดตัดจ่ายของ {0} เกินยอดที่จ่ายได้ {1} บาท | Amount for {0} exceeds the payable balance {1} |
| gl_err_ap_prepare_discount_invalid | ยอดส่วนลดต้องไม่ติดลบและไม่เกินรวมเป็นเงิน | Discount must be between 0 and the total |
| gl_err_ap_prepare_dimension_not_found | ไม่พบรหัสโครงการหรือรหัสแผนกนี้ | Project or department code not found |
| gl_err_ap_prepare_paid_locked | ใบเตรียมจ่ายนี้ถูกนำไปจ่ายแล้ว แก้ไขหรือยกเลิกไม่ได้ ให้ยกเลิกใบจ่ายชำระหนี้ก่อน | Already used by a payment; cancel the payment first |
| gl_err_ap_prepare_cancelled_locked | ใบเตรียมจ่ายนี้ถูกยกเลิกแล้ว | This preparation is cancelled |
| gl_err_ap_prepare_cancel_reason_required | กรุณากรอกเหตุผลการยกเลิก | Please enter a cancellation reason |
| gl_err_ap_prepare_version_conflict | มีผู้อื่นแก้ไขเอกสารนี้ไปแล้ว กรุณาเปิดใหม่ | Document changed by another user; reopen it |
| gl_err_ap_pay_prepare_creditor_mismatch | ใบเตรียมจ่าย {0} เป็นของเจ้าหนี้รายอื่น | Preparation {0} belongs to another creditor |
| gl_err_ap_pay_prepare_over_open | ยอดจ่ายของ {0} เกินยอดคงเหลือในใบเตรียมจ่าย {1} | Amount for {0} exceeds the open amount in preparation {1} |
| gl_err_erp_due_before_doc | วันครบกำหนดต้องไม่ก่อนวันที่เอกสาร | Due date cannot be before the document date |
| gl_err_erp_credit_days_invalid | จำนวนวันเครดิตต้องเป็นจำนวนเต็มตั้งแต่ 0 ขึ้นไป | Credit days must be a whole number ≥ 0 |

## 5. การทดสอบ

- **Go unit (sqlmock)** — ไฟล์ `ap_payment_prepare_test.go`
  - validation ทุก error code ใน §4.4
  - ทุก query ใช้ `$n` (ไม่มี string concat)
  - decimal: `"0.10"+"0.20"` ต้องได้ `"0.30"` และผ่านเมื่อ available `"0.30"`
  - ขอบเขตยอด: เท่ากับ available ผ่าน, เกิน 0.01 ไม่ผ่าน, ทศนิยม 3 ตำแหน่งไม่ผ่าน
  - ยอดรวมและส่วนลดคำนวณที่ backend
  - cancel ต้องมีเหตุผลและเขียน event
  - ใบที่ถูก PV ใช้แล้วต้องโดน lock
  - ตรวจว่าไม่มีการเขียน `gl_journals`
- **Go integration** — ใช้ `BC_GL_TEST_POSTGRES_DSN` ตามแบบ `internal/fixedasset/gl_poster_integration_test.go:38-40` หรือ `transaction/httpapi/save_integration_test.go`
  1. สร้าง PUR (หนี้ 10,000.00) ไว้ก่อน แล้วสร้างใบเตรียมจ่าย A ยอด 6,000.00 ใบเตรียมจ่าย B ต้องจ่ายได้แค่ 4,000.00
  2. PV ดึง A แบบจ่ายบางส่วน 2,500.00 แล้ว remaining ต้องเป็น 7,500.00, reserved 3,500.00, และ PUR ต้องยังไม่ปิด
  3. ยกเลิก PV แล้วค่าต้องกลับเป็นแบบเดิม
  4. ทดสอบหนี้ที่มาจาก GL (balance_side=1) ด้วย ว่า side normalize ถูก
  5. ยิงสอง goroutine พร้อมกันให้จองเกิน ต้องสำเร็จได้ทางเดียว (FOR UPDATE)
  6. วันครบกำหนด: บันทึก PUR ที่ creditday=30 แล้ว `doc.perioddatetime` และ `payload.due_date` ต้องเป็น docdate+30
  7. fixture ต้อง seed mapping ลงบัญชีของ `purchase`/`pay` (บัญชีคุมไม่ใช่เงินสด) และเจ้าหนี้ใน `gl_subledger_partners` ก่อน ไม่งั้น PUR ไม่เกิดหนี้ (§2); เคสไม่มี mapping → รายการค้างจ่ายว่าง
  8. PV ที่ลบแบบ soft delete (`doc.isdelete=true`) หรือยกเลิก ต้องทำให้ใบเตรียมจ่ายกลับมาเป็น "ยังไม่จ่าย" และยอดจองกลับมา
  9. as-of: settlement ที่ journal ยังไม่ผ่านบัญชี หรือ settlement_date หลังวันที่ถาม ต้องไม่ลด remaining; settlement ที่หนี้อยู่ฝั่ง `payment_id` ต้องถูกนับ (เทียบผลกับ `buildAgingQuery` ชุดข้อมูลเดียวกัน)
  10. แก้/ยกเลิก PUR ที่ถูกจองในใบเตรียมจ่ายที่ยังเปิด → `ap_prepare_reserved` (409) และแถวหนี้ยังอยู่; ยกเลิกใบเตรียมจ่ายก่อนแล้วยกเลิก PUR ได้ และใบเตรียมจ่ายที่ยกเลิกแล้วยังเปิดดูได้จาก snapshot
  11. ตัวตรวจ §4.5.1: PV ตัดหนี้จาก GL (`balance_side=1`) ได้, side ผิดถูกปฏิเสธ, สกุลเงินว่างเทียบ `THB` ผ่าน
  12. PV ที่มี WHT (เมื่อ B-1 เสร็จ) ตัดหนี้ที่เตรียมไว้ได้หมด remaining = 0
  13. ปิดงวด: บันทึก/ยกเลิกใบเตรียมจ่ายที่ `paydate` อยู่ในงวดที่ปิด → `source_period_locked`
- **vitest**
  - จอใหม่: render ลำดับฟิลด์ตาม Champ, ไม่มีข้อความไทย hard-code (guard test i18n เดิม), ปุ่ม Save และ dirty guard, แสดง error ตาม key
  - `menu-screen-status` ช่วงที่ยังไม่มีจอต้องแสดง "รอพัฒนา"
  - BFF route ใหม่ `api/ap-payment-prepare/[...path]`: รับเฉพาะ path/query ใน allowlist, ปฏิเสธ path อื่น
  - `erp-transaction.test.ts` ต้องยังผ่านโดย **ไม่** เพิ่ม `ap-payment-prepare` ใน `ERP_TRANSACTION_MODULES`
- **UAT CRUD** (Demo → `rungrueng` → บริษัท `01` → สาขา `00000`) ตรวจ `psql` ใน container `postgres` ทันทีหลังแต่ละขั้น
  1. Create
     - `SELECT docno,paydate,partner_code,total_amount,net_amount,version FROM ap_payment_prepare WHERE company=$1 AND docno='<เลขทดสอบ>'`
     - และ `SELECT linenumber,debt_document_id,payamount FROM ap_payment_prepare_lines WHERE ...` ต้องได้ 1 แถวหัว และบรรทัดต้องตรงจอ
  2. Read: ค้นหาด้วย Find, Prev และ Next
  3. Update: แก้ยอด แล้วตรวจว่าเป็นแถวเดิม `version` เพิ่มขึ้น 1 และมี event `update`
  4. จองเกิน: เปิดใบใหม่ด้วยหนี้ใบเดียวกัน ต้องเห็น `available_amount` ลดลงและบันทึกเกินไม่ได้
  5. PV Ctrl+F3: ตรวจ `docdetail.paybillno/reftype=1` และมี `gl_subledger_settlements` ต่อบรรทัด
  6. ยกเลิกใบที่ถูกใช้แล้ว ต้องได้ error `paid_locked`
  7. ยกเลิก PV แล้วยกเลิกใบเตรียมจ่าย ต้องได้ `iscancel=true` และ `cancel_reason` ไม่ว่าง และแถวยังอยู่
  8. ตรวจ edge cases: docno ว่าง, docno ซ้ำ, อักขระพิเศษ, double-click Save
  9. ล้างข้อมูลด้วย docno เป้าหมายเท่านั้น และตรวจว่า record อื่นไม่ถูกแตะ

## 6. ความเสี่ยง

1. workflow ที่ทำงานพร้อมกันกำลังแก้ `transaction/httpapi/**`, `posting.go`, `save.go`, `menu-screen-status.ts`, `debt_report.go`
   - W0, W1, W2 และ W5 ต้องเริ่มหลังงานนั้น commit แล้ว และต้องอ่าน diff ใหม่ก่อนลงมือ
2. side convention ของ GL กับ ERP ไม่เหมือนกัน ถ้าลืม normalize จะเลือกเอกสารจ่ายเงินมาเป็นหนี้
   - ต้องมี integration test ทั้งสองแหล่ง
3. PV ยังตัดหนี้ได้ใบเดียวและปิดเอกสารทั้งใบ (`save.go:233-237`) ถ้าไม่ทำ W5 ใบเตรียมจ่ายจะใช้จ่ายจริงไม่ได้
4. `models.go` ยังใช้ float64 จุดเชื่อมระหว่าง Go กับ JSON อาจเพี้ยนสตางค์ ต้องแก้พร้อม W5 (Rule 20)
5. เอกสารหนี้ AP ส่วนใหญ่ใน ERP ยังไม่เขียนบัญชีย่อย (ยกมา, ตั้งเจ้าหนี้อื่น, เพิ่มหนี้) ถ้าไม่ทำ W2 จะเตรียมจ่ายได้แค่ใบซื้อและหนี้ที่บันทึกจาก GL
6. ยังไม่มี master การจัดสรรและกลุ่มของบิล (audit D2) ช่องจึงต้องปิดไว้ก่อน เป็นช่องว่างของ parity ชั่วคราว
7. ไม่มีใน repo: โค้ดจอใบเตรียมจ่าย ตัวจัดการ Ctrl+F3 และการใช้ PrePayStatus ของ Champ
   - ลำดับประมวลผลและกฎ lock อนุมานจากใบรับวางบิลซึ่งมีโค้ดครบ (§1.7) ควรเทียบกับ Champ 5.5 ที่รันจริงก่อนส่งมอบ
8. ระหว่างที่ PV ยังไม่มี WHT/ส่วนลด/เช็ค (B-1..B-3) ผู้ใช้เตรียมจ่ายได้ แต่จ่ายจริงครบวงจรใน BC ยังไม่ได้สำหรับรายการที่ต้องหัก ณ ที่จ่าย
9. ข้อมูลเดิมที่ PV เคยตัดหนี้ด้วย settlement id `SET-AP-<docno>` (แบบหัว) จะไม่มี `prepare_docno` จึงไม่นับเป็น `used` — ไม่กระทบเพราะยังไม่มีใบเตรียมจ่ายจริง (ข้อมูลทิ้งได้ตามกฎ Dev)
10. snapshot `invbalance` เป็นค่า ณ เวลาบันทึก ถ้าภายหลังมีการตัดหนี้ทางอื่น ค่านี้จะเก่า แต่การตรวจใช้ `available_amount` สดเสมอ
11. คำนวณ `reserved` ทุกครั้งจาก settlement + lines ถ้าข้อมูลโตมากจะช้า
   - index `gl_subledger_settlement_debt`, `ap_payment_prepare_lines(company,debt_document_id)` ช่วยได้ และ query กรองตามเจ้าหนี้รายเดียว; ถ้ายังช้าค่อยเพิ่ม expression index บน `payload->>'prepare_docno'`

## 7. การตัดสินใจที่เปิดอยู่

ปิดแล้วด้วยหลักฐาน Champ หรือเป็นการออกแบบของ BC ที่ไม่ต้องถาม (หลังตรวจทาน 2026-10-11):

- D1 worklist vs PV+เช็ค → **เอกสารเตรียมจ่าย + PV ดึง** (Champ มีตาราง `BCPaymentTemp` + `BCPaymentSub.RefType=1` + ป้าย Ctrl+F3 ในภาพ PV)
- D2 ตารางเฉพาะ `ap_payment_prepare*` → การออกแบบของ BC (ไม่ใช่เอกสารสินค้า ไม่ลง GL, เลี่ยง float model และ test ทะเบียน doctype §4.4)
- D3 คำนวณยอดจองสด → ต่างจาก Champ โดยตั้งใจ (sync ACID, ไม่มีค่าค้าง) บันทึกใน §1.7
- D5 nDocType 0/2 = เงินสด → ปิด (§1.6, §4.6 ข้อ 4)
- D7 ฝั่งลด (BillType 2/8) → parity: Champ ให้อยู่ในใบเดียวกัน ทำหลัง PV ตัดหนี้กับฝั่งลดได้ (เป็นลำดับงาน ไม่ใช่คำถาม); 5 (มัดจำ) เป็นของ TOHOME ไม่ทำ
- D9 เลขที่เอกสาร → ผู้ใช้กรอกเอง (Champ ช่องบังคับสีชมพู) backend ตรวจซ้ำ
- PV ของ ERP ตัดหนี้ที่บันทึกจาก GL ได้ → Champ ตัดหนี้ทุกประเภทจากจอเดียว (§4.5.1)

คำถามที่เหลือให้ลุงจืด (ไม่มีข้อใดขวางการเริ่มงาน):

| # | คำถาม | แนะนำ | ทางเลือก |
|---|---|---|---|
| D4 | ยอดส่วนลดในใบเตรียมจ่ายให้ยกไปใส่ใบจ่ายชำระหนี้อัตโนมัติไหม | ไม่ยกไป ผู้ใช้กรอกส่วนลดที่ใบจ่ายเอง (ไม่มีหลักฐาน Champ ว่ายกไป) | ยกไปเป็นค่าตั้งต้นที่แก้ได้ |
| D10 | ให้สร้างใบเตรียมจ่ายโดยดึงจากใบรับวางบิลได้ด้วยไหม | ได้ ตาม Champ ที่ใบรับวางบิลมีช่องจองยอดเตรียมจ่าย ทำหลังจอหลักเสร็จ | ดึงได้เฉพาะเอกสารหนี้รายใบ |
| D11 **[ตัดสินแล้ว 2026-10-11: ทำตาม Champ]** | ใบเตรียมจ่ายที่ถูกจ่ายไปบางส่วนแล้ว ยังให้แก้ไขได้ไหม | ล็อกทันทีที่ถูกจ่ายบางส่วน กันยอดเพี้ยน | ให้แก้ได้แบบ Champ จนกว่าจะจ่ายครบ แต่ห้ามลดยอดต่ำกว่าที่จ่ายไปแล้ว |
| D6 | ปุ่ม Info / F1 Key / Recur ในจอนี้ | ยังไม่ทำรอบนี้ ทำพร้อมกันทุกจอเมื่อมีแบบกลาง | ทำครบตาม Champ ในจอนี้เลย |
| D8 | ช่อง "การจัดสรร" และ "กลุ่มของบิล" ก่อนมีทะเบียนรหัส | แสดงช่องแต่ปิดไว้ เปิดเมื่อมีทะเบียน | ซ่อนไว้ก่อน หรือให้พิมพ์ข้อความอิสระ |
| D12 | หนี้ที่ตั้งจากสมุดรายวันโดยตรง ให้แสดงประเภทเป็น "เจ้าหนี้อื่นๆ" ไหม | ใช่ แสดงเป็นเจ้าหนี้อื่นๆ | แสดงเป็นประเภทใหม่ "ตั้งหนี้จากบัญชี" |
| D13 | เปิดใช้ใบเตรียมจ่ายก่อนที่ใบจ่ายชำระหนี้จะหัก ณ ที่จ่าย/ส่วนลด/เช็คได้ไหม | ได้ เปิดใช้เตรียมจ่ายและดึงไปจ่ายก่อน แล้วทำหัก ณ ที่จ่ายต่อทันที | รอให้ใบจ่ายครบทุกอย่างก่อนค่อยเปิด |

## 8. ลำดับงาน

เงื่อนไขก่อนใช้งานจริง (prerequisite, ไม่ใช่โค้ดใหม่): holding ต้องมี mapping ลงบัญชีของ `purchase`/`pay` ที่บัญชีคุมไม่ใช่เงินสด และเจ้าหนี้ต้องอยู่ใน `gl_subledger_partners` (§2) — ทั้ง fixture ของ test และ UAT ต้อง seed

ลำดับที่แนะนำ: W0 → W2 → W1 → W3 → W5 → W4 → W6 (B-1..B-3 ขนานหลัง W5)

- W0 (S): เปลี่ยนจอ mock เป็น "รอพัฒนา" และลบ mock
  - ไฟล์: `menu-screen-status.ts`, `main-menu-screen.tsx`, `ap-payment-prepare-screen*.ts(x)`
  - working tree ทำไว้แล้ว (ยังไม่ commit) → ตรวจ/commit กับงานต้นทาง ไม่ทำซ้ำ
- W2 (M): ให้ `creditorbeginningbalance`, `apotherdebt`, `apdebitnote` (หนี้) และ `apcreditnote` (ฝั่งลด) เขียนบัญชีย่อย (`posting.go` `subledgerSpecs`) + ตาราง BillType/sign ตามประเภทเอกสาร (§4.5.3)
- W1 (M): วันครบกำหนดจากวันเครดิต (คำถามประกอบ C2)
  - backend: `credit_days` ใน partner payload, `creditday`/`duedate` ใน ERP, `perioddatetime` และ `payload.due_date`
  - frontend: ช่องในจอ ใช้ key เดิม `credit_day`/`credit_days`/`gl_detail_documents_due_date`
- W3 (L): backend ใบเตรียมจ่าย
  - schema (ไม่มี FK ไปหนี้, snapshot), ฟังก์ชันยอดคงเหลือเดียว `gl.APOpenItems` (posted + as-of + สองขา), route group `/ap-payment-prepare` ของตัวเอง + ตรวจสิทธิ์, ตรวจปิดงวด, audit
  - `cleanSubledger` ตรวจยอดจอง → `ap_prepare_reserved`
  - tests ตาม §5
- W5 (L): PV Ctrl+F3
  - ตัวตรวจ settlement §4.5.1, settlement ต่อบรรทัดพร้อม `prepare_docno/prepare_linenumber`, docdetail `paybillno/reftype/debt_document_id`, เลิกปิดเอกสารหนี้ทั้งใบ, เงินใน PV เป็น decimal, invariant ยอดเต็มต่อบรรทัด (§4.5.2)
- W4 (M): จอใหม่ตาม Champ, BFF ใหม่ `api/ap-payment-prepare/[...path]` (ไม่แตะ `ERP_TRANSACTION_MODULES`), language keys และ vitest
- B-1 (M): WHT ของ PV (F6) — ตรวจกฎหมายกับ rd.go.th + ลง `docs/kms/21` ก่อน
- B-2 (S): ส่วนลดของ PV; B-3 (M): บันทึกจ่ายเงิน/เช็ค/โอนของ PV (F7)
- W6 (S): UAT CRUD พร้อม psql และ smoke test บน production หลัง deploy

## 9. ส่วนเสริมเล็กนอก Champ (ต้องรายงานลุงจืด ถอดได้)

- ช่อง "ยอดจองโดยใบอื่น" และ "ยอดที่จ่ายได้" ในรายการค้นเอกสารค้างจ่าย ช่วยป้องกันการจองซ้ำ (Champ เก็บเป็น PrePayAmount แต่ไม่ได้แสดง)
- เหตุผลการยกเลิกพร้อมประวัติ (`/events`) ตามกฎ audit trail
- optimistic lock (`version`)
- ข้อความอธิบายเมื่อรายการค้างจ่ายว่าง (`erp_ap_prepare_no_posted_debt`) บอกวิธีแก้ตามกฎ error ไทย

## แหล่งอ้างอิง

- ประมวลรัษฎากร มาตรา 50 (หักภาษีทุกคราวที่จ่ายเงินได้): https://www.rd.go.th/5937.html
- ผู้รับเงินเป็นนิติบุคคล (ภ.ง.ด.53): ประมวลรัษฎากร ม.3 เตรส + คำสั่งกรมสรรพากร ท.ป.4/2528 — **ยังไม่ได้ตรวจกับ rd.go.th** ต้องตรวจและลงทะเบียนใน `docs/kms/21-thai-tax-form-references.md` ก่อนทำ B-1

## ปรับปรุงหลังตรวจทาน (2026-10-11)

- P1 รับ: ตัวตรวจ settlement ของ GL (`subledger_allocations.go:159`) ปฏิเสธคู่จาก ERP ทุกคู่ → เพิ่ม §4.5.1 ตัวตรวจต่อบรรทัดที่ normalize side/สกุลเงิน + test สองทิศ
- P2 รับ: ตัด FK ไป `gl_subledger_documents` เพราะ `cleanSubledger` DELETE หนี้จริง → snapshot + ตรวจยอดจองใน `cleanSubledger` คืน `ap_prepare_reserved` (409)
- P3 รับ: `used` นับ PV ที่ soft delete/ไม่เกิด settlement → เปลี่ยนเป็นนับจาก settlement ที่ยังมีผลซึ่งมี `prepare_docno` + PV ต้องกรอง `isdelete` ตอนตรวจร่าง
- P4 รับ: สูตรคงเหลือไม่ posted/as-of, ไม่นับขา `payment_id`, param ชน → ใช้ฟังก์ชันเดียวตาม `buildAgingQuery`/`subledgerReadCTE`, เรียง param ใหม่, เพิ่ม `id` ท้าย ORDER BY
- P5 รับ: เพิ่มใน `ERP_TRANSACTION_MODULES` จะทำ test ทะเบียนพัง และ `authorize` ตอบ invalid module → route group + BFF ของตัวเอง
- P6 รับ: เพิ่มเงื่อนไข 3 ข้อที่ทำให้ใบซื้อเกิดหนี้ (mapping, เจ้าหนี้ในบัญชีย่อย, บัญชีคุมไม่ใช่เงินสด) ใน §2/§8 + fixture + ข้อความ empty state
- P7 รับ: PV ยังไม่มี WHT/ส่วนลด/เช็ค → แยกเป็น B-1..B-3 + invariant "ตัดหนี้ยอดเต็ม, เงินออก = ยอดเต็ม − ส่วนลด − WHT" + test
- P8 รับ: 8 = จ่ายเงินล่วงหน้า, 5 เป็น TOHOME (ไม่อยู่ใน parity), DueDate ว่างใช้วันที่เอกสารเฉพาะประเภทเชื่อ 1/3, nDocType 0/2 = เงินสด → ปิด D5
- P9 รับ: เพิ่มตาราง BillType/sign จากประเภทเอกสาร ERP และ `document_kind` ของ GL (§4.5.3) ไม่ใช้รหัสบัญชี; GL kind 1 → 4 เป็นสมมติฐาน (D12)
- P10 รับ: ใช้ key เดิม `credit_day`/`credit_days`/`gl_detail_documents_due_date` แทน key ใหม่ 3 ตัว, แก้เลขบรรทัด `ap_payment_prepare` เป็น 8733
- P11 รับ: เปลี่ยนป้าย "audit C2" เป็น "คำถามประกอบ C2" (C2 ตัวจริง = ช่วงอายุหนี้ตั้งค่าได้ เสร็จใน `387aa183`)
- P12 รับ: เพิ่มคำถามดึงจากใบรับวางบิล (D10) และเปลี่ยน `reftype` เป็น 0=ไม่อ้างอิง 1=ใบเตรียมจ่าย 2=ใบรับวางบิล (ต่างเลข Champ โดยตั้งใจ; Champ ไม่บันทึก RefType จริง `:982`)
- P13 รับแบบยังไม่ยืนยัน: เพิ่มหมายเหตุ ม.3 เตรส/ท.ป.4/2528 สำหรับนิติบุคคล ยังไม่ได้ตรวจ rd.go.th — ต้องตรวจก่อน B-1
- เพิ่มเอง (กฎ 2026-10-11 ยึดลำดับประมวลผล Champ): §1.7 ตารางลำดับตรวจ/บันทึก/ลบของใบรับวางบิลและ PV พร้อม file:line และจุดต่างโดยตั้งใจ; กฎ lock ข้อ 11 อ้าง Champ แทน "กฎของ BC"; เพิ่มตรวจปิดงวดข้อ 15; ปิด D1/D2/D3/D5/D7/D9 ด้วยหลักฐาน; พบว่า W0 ทำไว้แล้วใน working tree
