# Design: เช็คและธนาคาร — ทะเบียนเช็ค/เอกสารเปลี่ยนสถานะ (B0), พิมพ์เช็ค (B1), การเคลื่อนไหวเช็ค (B2), เช็คเตรียมจ่าย (B3), รายละเอียดธนาคาร (B4)

- สถานะ: ข้อเสนอการออกแบบ วันที่ 2026-10-11 ยังไม่ได้แก้โค้ด เอกสารนี้เป็นไฟล์เดียวที่เขียน
- audit id: `B0` `B1` `B2` `B3` `B4`
- คำตัดสินของลุงจืด 2026-10-11 ที่ใช้ในแบบนี้:
  - ทำรายการบัญชีก่อน สินค้าทีหลัง
  - ยกเลิกเอกสาร ERP ทำระดับหัวเอกสาร
  - รายงานทุกชื่อตาม Champ อยู่ในเมนู แต่เปิด**จอรายงานกลางจอเดียว** ผ่าน report registry ฝั่ง backend
  - จอที่เป็น UI อย่างเดียวให้แสดง "รอพัฒนา" จนกว่าจะสร้างจริง
- งานที่ต้องรอ: workflow อื่นกำลังแก้ `backend/internal/transaction/**`, `generalledger/postgres_guards.go`, `generalledger/source_posting.go`, `goapi/handlers/debt_report.go`, `erp-transaction`, `erp-operations`, `erp-reports`, `menu-screen-status` และ `route.ts` → work item ที่แตะไฟล์เหล่านี้ต้องรอให้ workflow นั้น commit ก่อน
- เอกสารพี่น้องที่เกี่ยวข้อง:
  - `facade-screens.md` แถว 5–7 กับงาน F2: เปลี่ยน 3 จอเงา `cheque-movement`, `bank-service`, `cheque-issued-prepare` เป็น "รอพัฒนา" ก่อน → แบบนี้คือของจริงที่จะมาแทน
  - `facade-screens.md` §5 (CRIT-U2): รายชื่อรายงานเช็ค 16+8 ชื่อ
  - `ap-payment-prepare.md` §1.5: เช็คเตรียมจ่ายเป็นคนละเรื่องกับใบเตรียมจ่ายชำระหนี้
- ไฟล์ Champ ที่อ่าน:
  - `D:/project-champ/champ/champ/*.cpp` (CP874 ถอดด้วย Node)
  - `champ/Script/SQLSERVER_Script.sql` (UTF-16 ถอดแล้วไว้ที่ `C:/tmp/champ_ddl_utf8.sql`)
  - `champ-lib/BOLEDBCOMM/BEXTools.cpp`
  - ภาพหน้าจอ `bcspec/` (ถ้าขัดกับ `.rc` ให้ยึดภาพหน้าจอ)
  - ความรู้ที่สรุปไว้แล้ว: `docs/kms/champ/05-cheque.md`, `docs/kms/champ/06-cash-bank.md`

---

## 0. คำตอบสั้น

1. **Champ:** ระบบเช็คของ Champ ทำงานเป็นเครื่องสถานะ
   - เช็ครับมี 7 สถานะ (0–6) เช็คจ่ายมี 4 สถานะ (0–3)
   - ทะเบียนเช็ค (5377/5378) ไม่ลงบัญชี GL
   - เอกสารเปลี่ยนสถานะ 8 ตัว (5398–5405) เลือกเช็คจากกริดตามตัวกรองสถานะและวันที่ แล้วลงบัญชีตามแม่แบบที่ผู้ใช้กำหนด
   - ยกเว้น "นำฝาก" (5398) ที่ไม่ลงบัญชี
   - ลบหรือแก้เอกสาร = คำนวณสถานะใหม่จากเอกสารที่เหลือ — ทำได้เฉพาะเมื่อเช็คในเอกสารยังไม่ถูกเปลี่ยนสถานะต่อ (`IsChqReference` ห้ามจริง) และงวดยังไม่ปิด
2. **BC วันนี้:** มีแค่ 10 doctype (transflag 61, 418–426) ที่วิ่งผ่านเอกสาร ERP ทั่วไป (ทั้งจอ และ backend `/transaction/:module`)
   - เก็บเงินเป็น `float64`
   - ไม่มีทะเบียนเช็ค ไม่มีสถานะ และไม่มีตัวกรอง
   - จอสมุดเงินฝาก `/bookbankscreen` ถูกปิดเพราะ backend เดิมถูกถอด
   - จอ movement, service และ prepare เป็นข้อมูลจำลอง
3. **แบบที่เสนอ:** สร้างโมดูลใหม่ `backend/internal/cheque` (Go + PostgreSQL แบบ ACID) แยกจากเอกสาร ERP
   - `cheque_registers` (ทะเบียน) + `cheque_events` (ประวัติแบบ append-only)
   - `cheque_docs` / `cheque_doc_lines` (เอกสารเปลี่ยนสถานะ) ลง GL ผ่าน `gl.ExecuteSourceTx` และรูปแบบลงบัญชี (mapping) ตาม doctype key เดิม
   - mapping ได้แหล่งยอดใหม่ 3 ตัว คือ `cheque_total`, `cheque_fee`, `cheque_net` และ `accountfrom: "bank_book"` ที่ดึงบัญชี GL จากสมุดเงินฝากของเอกสาร
   - ไม่มีเงื่อนไขตามรหัสบัญชี
4. **ยอดคงเหลือ derive จากรายการทั้งหมด:** ยอดเช็คในมือของลูกหนี้/เจ้าหนี้ ยอดสมุดเงินฝากส่วนเช็ค และยอดคงเหลือของเช็คเตรียมจ่าย
   - ไม่เก็บยอดสะสมแบบ Champ ที่มีโอกาสเพี้ยน
   - เมนู "คำนวณยอดคงเหลือของเช็คใหม่" กลายเป็นตัวตรวจสอบและสร้าง cache ใหม่
5. **B1 พิมพ์เช็ค:** สร้าง PDF ฝั่ง server
   - ตำแหน่งเป็นมิลลิเมตรต่อสมุดเงินฝาก พร้อมปุ่ม "พิมพ์ตรวจตำแหน่ง"
   - ตัวอักษรคำอ่านจำนวนเงินมาจาก backend (`whtcert.BahtText`)
   - ระบบห้ามวางข้อความในแถบ MICR 15.875 มม. ล่างสุด และห้ามใช้ตัวอักษรเล็กกว่า 10 pt ตามมาตรฐาน ธปท. 2565
   - ตำแหน่งช่อง "ผู้รับเงิน" และ "จำนวนเงินตัวอักษร" มาตรฐานไม่ได้กำหนด จึงห้ามเดาค่าเริ่มต้น ต้องให้ผู้ใช้ calibrate
6. **B2:** จอสืบค้นอ่านอย่างเดียว ข้อมูลจาก `cheque_events` และ `cheque_source_links`
7. **B3:** ใบเตรียมเลขเช็คต่อสมุดเงินฝาก (เลขเริ่ม + จำนวน) สร้างรายการเลขที่ "ยังไม่ใช้"
   - เลขถูกใช้เมื่อบันทึกเช็คจ่ายหรือใบจ่ายที่ใช้เลขนั้น และคืนเมื่อลบ
   - กติกาการรันเลขของ Champ หาโค้ดไม่พบ → เสนอให้รันเป็นตัวเลขโดยคงความกว้าง (ยังไม่ verify)
8. **B4:** Champ จริงเก็บแค่ "รายละเอียด" เป็นบรรทัดข้อความต่อธนาคาร แยกต่อแท็บ (อัตราดอกเบี้ย/อัตราแลกเปลี่ยน/บริการธนาคาร) คอลัมน์ตัวเลขถูก comment ทิ้ง → ทำตามนั้น
9. **ข้อขัดขอบเขต:** `mydocs/specs/gl/ar-ap-bank-design-notes.md:49` เขียนว่า "workflow สั่งจ่าย/คืนเงิน/เช็คเป็นหน้าที่ระบบต้นทาง" ซึ่งขัดกับการทำวงจรเช็คครบแบบ Champ → **ต้องให้ลุงจืดตัดสินก่อนลงมือ** (open decision D1)

---

## 1. Champ ทำงานอย่างไร (หลักฐาน)

### 1.1 B0 — ทะเบียนเช็ค (5377 บันทึกเช็ครับ / 5378 บันทึกเช็คจ่าย)

**ลำดับฟิลด์ (ตามจอจริง — ถ้อยคำยึด screenshot `bcspec/หน้าจอเมนูเช็ค บัตร/ทะเบียนหลักต่างๆ-บันทึกเช็ครับ.jpg`, `...-บันทึกเช็คจ่าย.jpg` ซึ่งชนะ `.rc`; ปรับ 2026-10-11)**
- ที่มาโครง: `05-cheque.md` §3.1–3.2 (`BC5Account.rc:6601`, `cqfrmcheqin.cpp`, `cqfrmcheqout.cpp`)
- แถวที่ 1: เลขที่เอกสาร | วันที่รับเช็ค (เช็คจ่าย: เลขที่เอกสาร + **ปุ่ม "เช็คเตรียมจ่าย"** | วันที่จ่ายเช็ค)
- แถวที่ 2: รหัสลูกค้า (เช็คจ่าย: รหัสเจ้าหนี้) + ชื่อ
- แถวที่ 3: วันที่ครบกำหนด | สถานะเช็ครับ (เช็คจ่าย: สถานะเช็คจ่าย) ค่าเริ่มต้น "เช็คในมือ"
- แถวที่ 4: เลขที่เช็ครับ (เช็คจ่าย: เลขที่เช็คจ่าย) | สถานะเพิ่มเติม
- แถวที่ 5: รหัสธนาคาร | รหัสสกุลเงิน
- แถวที่ 6: สาขาของธนาคาร | อัตราแลกเปลี่ยน
- แถวที่ 7: สมุดเงินฝากธนาคาร (เช็คจ่าย: สมุดเงินฝาก) | จำนวนเงิน ( ตามสกุลเงิน)
- แถวที่ 8: จำนวนเงิน ( เงินบาท) (อ่านอย่างเดียว)
- แถวที่ 9: คำอธิบาย
- แถวที่ 10: รหัสพนักงานขาย (เฉพาะเช็ครับ) | รหัสแผนก
- แถวที่ 11: การปรับปรุงเอกสาร (อ่านอย่างเดียว)
- **ตารางอ้างอิงใต้หัวเอกสาร** (`cqfrmcheqin.cpp:192-196` หัวคอลัมน์, `:1227-1300` ข้อมูล): รายการ | เลขที่เอกสาร | วัน | จำนวนเงิน | คำอธิบาย
  - ข้อมูล = UNION ของเอกสารรับ/จ่ายเงินที่ใช้เช็คใบนี้ + ตารางย่อยของเอกสารเปลี่ยนสถานะทุกชนิด; คอลัมน์ "รายการ" แสดงชื่อชนิดเอกสาร (ตาม `SaveFrom`/transflag)
- **ปุ่ม "เช็คเตรียมจ่าย" (เฉพาะเช็คจ่าย)** — มีบนจอจริงใน screenshot แม้ handler ใน source ถูก comment (`cqfrmcheqout.cpp:94`); ตรรกะอยู่ที่ `OnSelectPrepare` `:1817-1848`
  - เปิดรายการเลขเช็คที่เตรียมไว้และยังไม่ถูกใช้ (`BCChqOutAutoSub.isuse=0`) → ใส่เลขที่เช็ค/สมุด/ธนาคาร/สาขา ตั้งสถานะ 3 เตรียมจ่าย และออกเลขเอกสารอัตโนมัติ
  - บันทึกแล้วตั้ง isuse=1 (`:537-539`); ลบทะเบียนแล้วคืน isuse=0 (`:451-453`)

**กติกา**
- **ฟิลด์บังคับ:**
  - เลขที่เอกสาร, วันที่, คู่ค้า, วันที่ถึงกำหนด, เลขที่เช็ค, ธนาคาร, สาขา, จำนวนเงิน
  - สมุดเงินฝาก: บังคับเมื่อสถานะ ≠ 0 สำหรับเช็ครับ และบังคับเสมอสำหรับเช็คจ่าย
- **เลขซ้ำ:**
  - เช็ครับ: ห้ามซ้ำที่ (ChqNumber + BankCode + BankBranchCode)
  - เช็คจ่าย: ระบบหาเช็คด้วย (ChqNumber + BookNo) (`myespeciallyformview.cpp:2255-2258`)
- **วันที่:** เช็คจ่ายที่วันที่จ่าย > วันที่ถึงกำหนด = ผิด (MSG_CHQOUT_DUEDATE_ERR)
- **สกุลเงิน:** จำนวนเงินบาท = อัตราแลกเปลี่ยน × จำนวนเงินตามสกุลเงิน คำนวณใหม่ทุกครั้งที่ค่าใดค่าหนึ่งเปลี่ยน
- **ค่าในคอมโบสถานะของเช็คจ่าย (`cqfrmcheqout.cpp:150-153`):** CB_CHEQ_TYPE_1 (ในมือ), CB_CHEQ_TYPE_3 (ผ่าน), CB_CHEQ_TYPE_5 (ยกเลิก), PREPAIR_PAYMENT (เตรียมจ่าย)
  - เมื่อสถานะเป็น 3 ตอนสร้างใหม่ ระบบตั้ง `balance = จำนวนเงิน` (`:499-501`)
- **การกันแก้:** ถ้าเช็คถูกอ้างแล้ว (ใน `BCRecMoney` หรือในตารางย่อยของเอกสารเปลี่ยนสถานะ) ระบบจะ**ถามยืนยัน** CANNOT_UPDATE_REFERENCE (`cqfrmcheqin.cpp:1329-1350`, ฝั่งบันทึก `:320-327`) ไม่ได้ห้าม
  - เช็คที่สร้างจากเอกสารต้นทาง (`SaveFrom` ≠ 4) ถามยืนยัน CANNOT_UPDATE_REFERENCE_OTHER (`:328-336`)
- **การกันลบ (ห้ามจริง ไม่ใช่ถาม):** เช็คที่ถูกอ้างแล้ว → MSG_REFERANCE (`cqfrmcheqin.cpp:438-443`); เช็คที่มาจากเอกสารต้นทาง → MSG_FROM_OTHER (`:444-449`)
- **งวดบัญชี:** จอทะเบียนไม่ตรวจงวดปิด (ไม่มี `IsPeriodClose` ในฟอร์มทะเบียน) เพราะทะเบียนไม่ลงบัญชี
- **ยอดเงินบาท:** `OnCalculateBathAmount` (`cqfrmcheqin.cpp:1129-1139`): ถ้าบริษัทไม่เปิดหลายสกุลเงิน (`MultiCurrency≠1`) หรืออัตราว่าง อัตรา = 1; แสดงผลด้วยทศนิยมของบริษัท (`AmountPoint`)
  - ตอนบันทึกเอกสารสถานะ ยอดบาทปัดด้วย `XRound` ตามทศนิยมของบริษัท (`cqfrmcheqindeposit.cpp:475-476`) = ปัดครึ่งขึ้น `floor(v×10^n+0.5)/10^n` (`champ-lib/QOledblib/QOledb.cpp:1052-1102`)
- **GL:** ฟอร์มทะเบียนไม่มีการเรียก GL เลย
  - ตรวจด้วย `grep BCTrans|GLFormat|PostGL` ใน `cqfrmcheqin.cpp` ได้ 0 ครั้ง
  - `cqfrmcheqout.cpp` มี 1 ครั้งที่เป็น query พิมพ์ GL slip (`:1624`)

**เช็คที่เกิดจากเอกสารต้นทาง**
- ใบเสร็จรับเงิน, ใบจ่าย และจ่ายอื่นๆ เขียน `BCChqIn`/`BCChqOut` เอง (`myespeciallyformview.cpp:2281-2316`, `ARFrmReceiptMultiAr.cpp:5930-5995`)
  - ใส่ `SaveFrom` = ที่มา และ `RefChqRowOrder`
  - ถ้าเป็นเช็คจ่ายที่เลขตรงกับเลขเตรียมจ่าย ระบบตั้ง `BCChqOutAutoSub.isuse=1` (`:2312-2316`)
- **ตัดยอดจากเช็คเดิม** (`m_nPayChqState == 1`, `myespeciallyformview.cpp:2240-2247`)
  - การใช้เช็คใบเดิมจ่ายหลายเอกสารจะลด `Balance` ของเช็ค
  - การลบเอกสารคืนยอด (`:1443-1448`, `:2047`, `ARFrmReceiptMultiAr.cpp:5781`)
  - → **เช็คหนึ่งใบผูกได้หลายเอกสารต้นทาง โดยแต่ละเอกสารมียอดจัดสรรของตัวเอง**

**ทะเบียนรหัส ("กำหนดรหัสต่างๆ" จอเดียวมีคอมโบชนิดรหัส `myfrmbasemaster.cpp:229-236`)**
- ชนิดรหัส: กำหนดรหัสธนาคาร (`BCBank`), กำหนดรหัสสาขาของธนาคาร (`BCBankBranch`), กำหนดรหัสบริการธนาคาร (`BCBANKSERVICE`), กำหนดรหัสประเภทบัญชีธนาคาร (`BCBANKACCOUNT`), กำหนดรหัสสถานะเช็คเพิ่มเติม (`BCChqStatus`)
- กำหนดประเภทบัตรเครดิต (`BPSCreditType`) อยู่นอกขอบเขตนี้
- ทุกตารางมีโครงเดียวกัน คือ `Code VARCHAR(10)` PK, `Name VARCHAR(50)`, `UserGroup VARCHAR(255)` + ผู้สร้าง/ผู้แก้ (DDL `:10512`, `:10528`, `:10659`, `:10771`, `:10804`)
- **สาขาธนาคารเป็น PK ทั่วระบบ ไม่ผูกกับธนาคาร**

**สมุดเงินฝาก 5169 (`BCBankBook` DDL `:10544`; `06-cash-bank.md` §2.1)**
- PK = `BookNo` (เลขที่บัญชี)
- ฟิลด์อื่น: `Name`, `AccountType` (→ `BCBankAccount`), `BankCode`, `BranchCode`, `AccountCode` (บัญชี GL), `Balance` (สะสม)
- จอจริง (`bcspec/หน้าจอเมนูธนาคาร เงินสด/บันทึกสมุดเงินฝากธนาคาร.jpg`): เลขที่บัญชีธนาคาร + **ปุ่ม "เปลี่ยนเลขที่บัญชี"** | ประเภทบัญชี / ชื่อบัญชี / ธนาคาร | สาขาธนาคาร / รหัสผังบัญชี | ยอดคงเหลือ + **ตาราง "ชื่อผู้ติดต่อ"** (รหัสผู้ติดต่อ, ชื่อ/รายละเอียด, ตำแหน่ง, ที่อยู่, โทรศัพท์)
  - ตารางผู้ติดต่อ**แสดงอย่างเดียว** ดึงจาก `BCBankContact` ตาม ธนาคาร+สาขา ของสมุด (`CASHFrmBankBookNew.cpp:292-296`; DDL `:10588-10614`) — ข้อมูลกรอกที่จอ 5680 "บันทึกรายละเอียดผู้ติดต่อธนาคาร" (คีย์ ธนาคาร+สาขา, คอลัมน์ A = ติดต่อ/ไม่ติดต่อ `DefaultStatus`)
  - "เปลี่ยนเลขที่บัญชี" (`CASHFrmBankBookNew.cpp:456-480`, `CASHDlgChangeBookNo.cpp:59-130`): ถามยืนยัน แล้ว UPDATE `BookNo` ทุกตารางที่อ้าง (เงินฝาก/ถอน/ค่าใช้จ่าย/รายได้/โอน/ทะเบียนเช็ค/บัตร/เอกสารสถานะทุกชนิด) ใน transaction เดียว เพราะเลขบัญชีเป็น PK
  - ลบ: ห้ามเมื่อไม่ระบุเลขบัญชี MSG_890 หรือยอดคงเหลือ ≠ 0 MSG_891 (`:186-207`) แล้วถามยืนยัน

### 1.2 B0 — เอกสารเปลี่ยนสถานะ (ฟอร์มแม่ `cqfrmbasetransstatus.cpp`)

**สถานะเช็ครับ (`BCChqIn.Status`)**

| ค่า | ป้าย |
|---|---|
| 0 | เช็คในมือ |
| 1 | เช็คฝากธนาคาร |
| 2 | เช็คผ่าน |
| 3 | เช็คคืน |
| 4 | เช็คยกเลิก |
| 5 | ขายลดเช็ค |
| 6 | เช็คคืนฝากธนาคาร |

**สถานะเช็คจ่าย (`BCChqOut.Status`)**

| ค่า | ป้าย |
|---|---|
| 0 | เช็คในมือ |
| 1 | เช็คผ่าน |
| 2 | เช็คยกเลิก |
| 3 | เตรียมจ่าย |

ที่มา: `05-cheque.md` §2.1–2.2 (ป้ายยืนยันจาก `REP:BCRepCashChq.rc` CB_CHEQ_TYPE_1..7)

| Champ id | เอกสาร (เมนู BC) | เช็คที่เลือกได้ | สถานะใหม่ | GL format เริ่มต้น | สมุดเงินฝาก |
|---|---|---|---|---|---|
| 5398 | บันทึกนำฝากเช็ครับ (`cheque-deposit`, 419) | รับ: 0 และ วันที่รับ ≤ วันที่เอกสาร | 1 | ไม่ลงบัญชี | มี |
| 5401 | บันทึกเช็ครับผ่าน (`cheque-received-clear`, 420) | รับ: 0/1 และ วันที่ถึงกำหนด ≤ วันที่เอกสาร | 2 | 16 | บังคับ |
| 5408 | บันทึกเช็ครับคืน (`cheque-received-return`, 421) | รับ: 2/6 และ วันที่รับ ≤ วันที่เอกสาร | 3 | 17 | มี |
| 5399 | บันทึกนำเช็คเข้าใหม่ (`cheque-redeposit`, 422) | รับ: 3 | 6 | 19 | มี |
| 5402 | บันทึกยกเลิกเช็ครับ (`cheque-received-cancel`, 423) | ตามประเภทการยกเลิก (ตารางล่าง) | 4 | 18 | ตามประเภท |
| 5403 | บันทึกขายลดเช็ครับ (`cheque-discount`, 424) | รับ: 0/1 และ วันที่รับ ≤ วันที่เอกสาร | 5 | 27 | มี |
| 5404 | บันทึกเช็คจ่ายผ่าน (`cheque-issued-clear`, 425) | จ่าย: 0 และ วันที่ถึงกำหนด ≤ วันที่เอกสาร | 1 | 20 | กรองได้ |
| 5405 | บันทึกยกเลิกเช็คจ่าย (`cheque-issued-cancel`, 426) | จ่าย: 0 หรือ (1 เมื่อติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว" + สมุดเงินฝาก) | 2 | 22 | ตามติ๊ก |

**ประเภทการยกเลิกของ 5402 (`CancelStatus`; ตารางคืนสถานะ `cqfrmbasetransstatus.cpp:147-165`)**

| ค่า | ความหมาย | เลือกเช็คสถานะ | ต้องมีสมุดเงินฝาก |
|---|---|---|---|
| 0 | ยกเลิกเช็คในมือ | 0 | ไม่ |
| 1 | ยกเลิกเช็คที่ผ่านแล้ว | 2 | ใช่ |
| 2 | ยกเลิกเช็คฝาก | 1 | — |
| 3 | ยกเลิกเช็คคืนฝากธนาคาร | 6 | ใช่ |

**โครงร่วมของเอกสาร (`05-cheque.md` §4)**
- **หัวเอกสาร:** เลขที่เอกสาร → วันที่เอกสาร → รหัสสมุดเงินฝาก → หมายเหตุ
  - 5402 มีคอมโบ "เช็คที่ต้องการยกเลิก"
  - 5405 มีช่องติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว"
- **หัวเอกสาร (ถ้อยคำตาม screenshot):** เลขที่เอกสาร | วันที่เอกสาร / รหัสสมุดเงินฝาก / หมายเหตุ
- **กริด (ถ้อยคำตาม screenshot `บันทึกเช็ครับผ่าน.jpg`):**
  - คอลัมน์: เลขที่เช็ค (ลิงก์เปิดรายการเช็คที่เลือกได้), ธนาคาร, วันที่ครบกำหนด, จำนวนเงินบนเช็ค, จำนวนเงิน (บาท), **ค่าใช้จ่าย (แก้ได้ต่อแถว)**, ยอดที่หักค่าใช้จ่ายแล้ว (`cqfrmcheqinpass.cpp:594-619`)
  - **ยกเว้นนำฝาก 5398 (`บันทึกนำฝากเช็ครับ.jpg`):** เลขที่เช็ค, ธนาคาร, วันที่ครบกำหนด, จำนวนเงินบนเช็ค, จำนวนเงิน, หมายเหตุ — **ไม่มีคอลัมน์ค่าใช้จ่าย/ยอดหัก**; คอลัมน์หมายเหตุอ่านอย่างเดียว แสดงคำอธิบายของทะเบียนเช็ค (`cqfrmcheqindeposit.cpp:319`, `:484`, `:500`; ตารางย่อย `BCChqInDeposSub` ไม่มีคอลัมน์หมายเหตุ DDL `:8022`)
- **ท้ายเอกสาร:** จำนวนเงินรวม (`SumChqAmount`), รวมค่าใช้จ่าย (`SumExpend`), ยอดที่หักค่าใช้จ่ายแล้ว (`NetAmount = SumChqAmount − SumExpend`)
  - ถ้อยคำ "รวมค่าใช้จ่าย" ตรวจกับ screenshot จอรับผ่าน รับคืน นำเข้าใหม่ ยกเลิกรับ ยกเลิกจ่าย แล้ว
  - **นำฝาก 5398 มีแค่ "จำนวนเงินรวม"**

**ลำดับตรวจตอนบันทึก (5401 `cqfrmcheqinpass.cpp:314-436` อ่านทีละบรรทัดแล้ว; ฟอร์มอื่นใช้โครงเดียวกัน)**
1. เอกสารเดิมที่มีเช็คถูกเปลี่ยนสถานะต่อแล้ว → ห้ามแก้ (`IsChqReference` `:316-321`, ตัวตรวจ `:1623-1650`)
2. งวดปิด → ห้าม (`IsPeriodClose` `:325`)
3. ผ่านบัญชี GL แล้ว → ห้าม POSTGL_WARN (`IsPostGL` `:345`; `champ-lib/BC5FRMWRK/mybasedbformview.cpp:4126-4166`)
4. เอกสารอนุมัติแล้ว (`isconfirm`) → ห้าม (`:349-354`) — แต่ไม่มีโค้ดใดตั้ง isconfirm ของตาราง BCChq\* (ดู §7)
5. ไม่มีเลขที่หรือวันที่ → ห้าม
6. ยอดเป็น 0 หรือไม่มีแถว → ห้าม (`:372`)
7. ไม่มีสมุดเงินฝาก → ห้าม (`:379`)
8. แถวที่เช็คมีสถานะ > 2 แล้ว → ห้าม CANNOT_EDIT_REFERENCE (`:399-418`)
9. ยอดรวมในหัวไม่เท่ากับยอดในกริด → ห้าม MSG_CHQ_FEEAMOUNT_ERR (`:421`)

**ลบเอกสารสถานะ (ห้ามจริง ไม่ใช่ถาม):**
- นำฝาก: เช็คถูกใช้ต่อแล้ว ห้ามแก้ `cqfrmcheqindeposit.cpp:117-125` / ห้ามลบ `:186-194` (ตัวตรวจ `IsChqReference` `:1521-1550`); อนุมัติแล้วห้ามแก้ `:145-150` / ห้ามลบ `:200-205`
- รับผ่าน: ห้ามลบ `cqfrmcheqinpass.cpp:439-444`; อนุมัติแล้ว `:455-460`
- รับคืน: `cqfrmcheqinreturn.cpp:1621-1641`

**ตัวกรองเช็คที่เลือกได้ของ 5401–5402 (ตรวจแล้ว 2026-10-11)**
- นำเช็คเข้าใหม่: ต้องเลือกสมุดก่อน CHQ_STATUS_WARN (`cqfrmcheqindeposit2.cpp:711-714`); เลือกได้เฉพาะสถานะ 3 + สมุดเดียวกัน + วันที่รับ ≤ วันที่เอกสาร (`:721`, `:724`)
- ยกเลิกเช็ครับ (`cqfrmcheqincancel.cpp:757-789`): ทุกชนิดใช้ วันที่รับ ≤ วันที่เอกสาร
  - ชนิด 0 ยกเลิกเช็คในมือ: สถานะ 0 ไม่ต้องมีสมุด (ช่องสมุดเป็นสีเทาในจอจริง)
  - ชนิด 1/2/3: ต้องมีสมุด (CHQ_STATUS_WARN) และสถานะ 2 / 1 / 6 ตามลำดับ

**ลบหรือแก้ (เมื่อผ่านตัวกันข้างบนแล้ว):** ย้อนผลทุกแถว แล้วคำนวณสถานะของเช็คใหม่จากเอกสารที่เหลือ
- ใช้ UNION ของตารางย่อย เรียงวันที่เอกสารจากใหม่ไปเก่า (`cqfrmbasetransstatus.cpp:264-300`)
- ถ้าสถานะกลับเป็น 0 ระบบล้าง BookNo ของเช็ครับ
- เพราะ `IsChqReference` ห้ามแก้/ลบเอกสารที่เช็คถูกเปลี่ยนสถานะต่อแล้ว การคำนวณใหม่จึงไม่ข้ามขั้นตอน (ต้องยกเลิกเอกสารที่ใหม่กว่าก่อน)

**ผลต่อยอดสะสมของ Champ (BC จะ derive แทน — `05-cheque.md` §2.5–2.6)**
- `BCBankBook.Balance ± NetAmount`
- `BCAR.ChqOnHand` / `ChqReturn` / `DebtLimitBal`
- `BCAP.ChqOnHand` / `ChqReturn`

### 1.3 B0 — การลงบัญชี GL ของเอกสารเช็ค

- **แม่แบบ (GLFormat):** ผู้ใช้กำหนดเอง ไม่ได้ฝังบัญชีไว้ในโค้ด (`05-cheque.md` §5.1; `LIB:gltlsprocess.cpp:7637-8026`)
- **เงื่อนไขของบรรทัดแม่แบบ:** 1 = `SumChqAmount`, 2 = `SumExpend`, 3 = `NetAmount` โดยใช้บัญชี GL ของสมุดเงินฝากในเอกสาร
- **บัญชีของบรรทัดเงื่อนไข 3 (`ProcessConditionChqGroup` `gltlsprocess.cpp:7637`, อ่านแล้ว 2026-10-11):**
  - ตั้งต้นใช้บัญชีในแม่แบบ (`:7766`) แล้วเงื่อนไข 3 ล้างใหม่ (`:7795`)
  - เอกสารทุกชนิดยกเว้นยกเลิกเช็ครับ: มีสมุด → บัญชีของสมุด (`GetBookAccountCode` `:7802-7804`); ไม่มีสมุด → ยอดเป็น 0 (`:7806-7809`)
  - ยกเลิกเช็ครับ (`BCChqInCancel`): สถานะ ≠ 0 และมีสมุด → บัญชีของสมุด (`:7818-7819`); ไม่มีสมุดและสถานะ ≠ 0 → 0 (`:7821-7822`); **ไม่มีสมุดและสถานะ 0 (ยกเลิกเช็คในมือ) → ใช้บัญชีในแม่แบบ** (`:7824`)
  - บรรทัดที่ยอด 0 หรือไม่มีบัญชีถูกข้าม (`:7870`); ผลต่างปัดเศษลงบรรทัดปัดเศษเฉพาะเมื่อ ≤ 1 (`:7916-7958`)
  - ตารางที่ลงบัญชี (`:7603-7634`): รับผ่าน, รับคืน, ยกเลิกรับ, นำเข้าใหม่, ขายลด, จ่ายผ่าน, ยกเลิกจ่าย
  - รหัสแม่แบบเริ่มต้นถูกใส่ตอนบันทึก (เช่น รับผ่าน `cqfrmcheqinpass.cpp:431-432` แม่แบบ 16)
- **โหมด:** ลงบัญชีอัตโนมัติ หรือคีย์มือ (`GLTransData`)
- **นำฝาก 5398 ไม่ลงบัญชี**
- เนื้อหาแม่แบบเริ่มต้น (บัญชีเดบิต/เครดิต) อยู่ในฐานข้อมูลลูกค้า ไม่อยู่ใน repo → **BC ห้ามกำหนดบัญชีเริ่มต้นเอง** ผู้ใช้ต้องตั้งรูปแบบลงบัญชีเอง

### 1.4 B0 — รายละเอียดเช็คใต้สมุดรายวัน GL (`BCGLChq`)

- จอบันทึกรายวันของ Champ มีแท็บเช็ค
  - เก็บที่ `BCGLChq` (BookCode, DocDate, DocNo, Source, BankCode, ChqNo, Amount, DueDate, BookNo, ChqType, ExchangeRate, BranchID — DDL `:9100`)
  - อ้าง `glfrmjournal.cpp:754`, `:831`, `:2550`, `:2622-2662`, `:2673`, `:3409`
- เป็นหลักฐานแนบรายวัน ไม่ได้สร้างหรือเปลี่ยนแถวใน `BCChqIn`/`BCChqOut`

### 1.5 B1 — พิมพ์เช็คจ่าย

- **ปุ่มพิมพ์ของ 5378** เปิด `CLIBDlgPrintChqOut`
  - ต้องระบุช่วงเลขที่เอกสาร**และ**ช่วงเลขที่เช็ค ห้ามว่าง (`LIBDlgPrintChqOut.cpp:136-145`)
  - ติ๊กเลือกได้: Document Form (ค่าเริ่มต้น), GL Slip Form, Withhold. Tax Form, Print Serial Number (`:22-38`)
  - ปุ่ม Preview ใช้ได้เมื่อเลือกแบบเดียว (`:178-196`)
- **ฟอร์มเช็ค** = ไฟล์ฟอร์มจากโฟลเดอร์ฟอร์มของ form designer (`cqfrmcheqout.cpp:1643-1700` `theApp.GetFomFolder()` + `m_Form.GetFileData`)
  - ข้อมูลมาจาก `SELECT * FROM BCCHQOUT WHERE DOCNO=… AND CHQNUMBER=…` (`:1601-1616`)
  - → **Champ ไม่มีแม่แบบเช็คในตัว ผู้ใช้ออกแบบฟอร์มเอง**
- **พิมพ์แล้ว:** `PrintCount = PrintCount + 1` (`:1806-1814`)
- **คำอ่านจำนวนเงิน:** `XMoneyWord(double, TRUE)` (`champ-lib/BOLEDBCOMM/BEXTools.cpp:1001`)
  - ขึ้นต้นด้วย `"( "` (`:1030`) และใช้ `%0.2lf` จาก double
  - BC ต้องใช้ decimal (กฎ 20)

### 1.6 B2 — ข้อมูลการเคลื่อนไหวเช็ค (`CQFrmChqFlow.cpp`; ภาพ `bcspec/หน้าจอเมนูเช็ค บัตร/ข้อมูลการเคลื่อนไหวเช็ค.jpg`)

- **หัวจอ (ตามภาพ):**
  - ซ้าย: เลขที่เช็ค (+ค้นหา), วันที่ครบกำหนด, ธนาคาร/สาขา, ชื่อลูกค้า / เจ้าหนี้
  - ขวา: ปุ่มวิทยุ เช็ครับ / เช็คจ่าย, มูลค่าเช็ค, เลขที่บัญชีธนาคาร, สถานะเช็ค
  - ทุกช่องอ่านอย่างเดียว ยกเว้นเลขที่เช็ค
- **กริด 1 "รายวันซื้อ/ขาย":** รายการ, เลขที่เอกสาร, วันที่เอกสาร, จำนวนเงิน, คำอธิบาย
  - เช็ครับดึงจาก `BCRecMoney WHERE PaymentType=2` (`:479-481`)
  - เช็คจ่ายดึงจาก `BCPayMoney` (`:555`)
- **กริด 2 "รายวันเช็ค":** รายการ, เลขที่เอกสาร, วันที่เอกสาร, คำอธิบาย
  - เป็น UNION ของตารางย่อยของเอกสารเปลี่ยนสถานะ (`:617-652`; เช็คจ่าย `:663-671`)
- **บั๊กของ Champ ที่ BC จะไม่ลอก:**
  - ป้ายสถานะเช็คจ่ายในหัวจอใช้ CB_CHEQ_TYPE_4 กับสถานะ 2 และ CB_CHEQ_TYPE_5 กับสถานะ 3 (`:449-453`)
  - แต่จอบันทึกเช็คจ่ายใช้ TYPE_5 = ยกเลิก กับสถานะ 2 และ "เตรียมจ่าย" กับสถานะ 3 (`cqfrmcheqout.cpp:150-153`)
  - ผล: เช็คจ่ายที่ยกเลิกจะแสดงเป็น "เช็คคืน"
  - เป็นการอ่านโค้ด ยังไม่ได้รันจริง

### 1.7 B3 — บันทึกเช็คเตรียมจ่าย (ภาพ `bcspec/หน้าจอเมนูเช็ค บัตร/บันทึกเช็คเตรียมจ่าย.jpg`, หัวจอ CQA)

- **ลำดับฟิลด์ (ตามภาพ):**
  - เลขที่เอกสาร (บังคับ), วันที่เอกสาร, เลขที่เช็คจ่าย (เลขเริ่ม), จำนวนที่ Run Auto
  - รหัสธนาคาร (บังคับ), สาขาของธนาคาร (บังคับ), สมุดเงินฝาก (บังคับ), คำอธิบาย
  - ปุ่ม **"รายการอัตโนมัติ"**
  - กริดขวา "เลขที่เช็ค"
- **ตาราง:**
  - `BCChqOutAuto`: DocNo PK, DocDate, BankCode, BankBranchCode, BookNo, MyDescription, ChqNoStart varchar(20), ChqRun smallint (DDL `:14834`)
  - `BCChqOutAutoSub`: ChqNumber, IsUse (DDL `:14853`)
- **การใช้เลข:**
  - เลขถูกตั้ง `isuse=1` เมื่อเอกสารจ่ายสร้างเช็คที่เลขตรงกัน (`myespeciallyformview.cpp:2312-2316`)
  - หรือเมื่อบันทึกเช็คจ่ายจากรายการเตรียมจ่าย (`cqfrmcheqout.cpp:537-539`)
  - ลบเช็คจ่าย → `isuse=0` (`:445-453`)
  - รายการให้เลือกแสดงเฉพาะ `isuse=0` (`:1817-1840`)
  - ใน source บรรทัด message map ของปุ่มถูก comment ไว้ (`:94`) **แต่ screenshot `ทะเบียนหลักต่างๆ-บันทึกเช็คจ่าย.jpg` มีปุ่ม "เช็คเตรียมจ่าย" ข้างเลขที่เอกสาร** → ยึดจอจริง: BC ทำปุ่มนี้ (ดู §1.1)
- **ไม่พบซอร์สของฟอร์ม 5xxx "บันทึกเช็คเตรียมจ่าย"** ใน repo (ค้น `ChqNoStart` เจอแค่ใน DDL) → **กติกาการรันเลขของ Champ = ไม่ทราบ**

### 1.8 B4 — บันทึกรายละเอียดธนาคาร (`CashFrmBankService*.cpp`; ภาพ `bcspec/หน้าจอเมนูธนาคาร เงินสด/บันทึกรายละเอียดธนาคาร.jpg`)

- **หัวจอ:** รหัสธนาคาร
- **แท็บ:** อัตราดอกเบี้ย | อัตราแลกเปลี่ยน | บริการธนาคาร
  - ทุกแท็บเป็นกริดคอลัมน์เดียว "รายละเอียด"
  - คอลัมน์ประเภทบัญชี/อัตรา/ณ วันที่ ถูก comment ทิ้ง (`CashFrmBankServicePage1.cpp:79-84`, `Page3.cpp:79-81`)
- **ตาราง:**
  - แท็บ 1 → `BCBANKINTERESTS` (`Page1.cpp:164`)
  - แท็บ 2 → `BCBANKCURRENCY` (`Page2.cpp:164`)
  - แท็บ 3 → `BCSERVICEFEE` (`Page3.cpp:161`)
- **การบันทึก:** ลบทั้งหมดของธนาคารนั้นแล้วใส่ใหม่ (`Page1.cpp:210`, `Page2.cpp:210`, `Page3.cpp:201`)
  - แท็บ 1–2 บันทึก log ก่อนลบ (`BCINTERESTSLOG` `Page1.cpp:245`, `BCCURRENCYLOG` `Page2.cpp:242`)
  - แท็บ 3 ไม่มี log

---

## 2. ข้อเท็จจริงทางการ (ใช้กับ B1)

| เรื่อง | ค่า | แหล่ง |
|---|---|---|
| ขนาดเช็ค (ไม่รวมต้นขั้ว) | ยาว 178 มม. (+2), สูง 89 มม. (+2) | ธปท. แนวนโยบายมาตรฐานเช็ค… 1 เม.ย. 2565 ข้อ (1) ขนาดของเช็ค — https://www.bot.or.th/content/dam/bot/documents/th/our-roles/payment-systems/about-payment-systems/BOT_2_2565_1.pdf |
| แถบ MICR Clear Band | 5/8 นิ้ว = **15.875 มม.** ชิดขอบล่าง ตลอดความยาว | ฉบับเดียวกัน ส่วนที่ 1 MICR Clear Band และ 3.2.7 (1) "ห้ามพิมพ์ หรือเขียนข้อความใด ๆ ในแถบว่าง…" |
| ช่องวันที่สั่งจ่าย | มุมบนขวา 8 ช่อง ยาวประมาณ 2 นิ้ว สูงจากขอบล่างประมาณ 3 1/8 นิ้ว รูปแบบตาม Date Field Indicator เช่น DDMMYYYY / ววดดปปปป | ฉบับเดียวกัน 1.1 และ 3.2.7 (8) |
| ช่องจำนวนเงินตัวเลข | ด้านขวา ใต้วันที่ ยาวประมาณ 2 5/8 นิ้ว สูงจากขอบล่างประมาณ 1 3/4 นิ้ว และควรกรอกชิดสัญลักษณ์ | ฉบับเดียวกัน 1.2 และ 3.2.7 (7) |
| ตัวอักษรที่พิมพ์เพื่อสั่งจ่าย (ชื่อผู้รับ, วันที่, จำนวนเงินตัวเลข/ตัวอักษร) | ไม่น้อยกว่า **10 Point (3.51 มม.)** | ฉบับเดียวกัน (2.3.2) |
| ตำแหน่งชื่อผู้รับเงิน / จำนวนเงินตัวอักษร | ไม่กำหนดพิกัด (กลุ่มข้อมูลที่ไม่บังคับบริเวณ) → **ไม่ทราบค่า ต้อง calibrate ต่อแบบพิมพ์ของธนาคาร** | ฉบับเดียวกัน (กลุ่มที่ 1 กำหนดเฉพาะ 1.1–1.4) |
| รายการที่เช็คต้องมี | ป.พ.พ. มาตรา 988 (คำบอกชื่อว่าเป็นเช็ค, คำสั่งไม่มีเงื่อนไขให้ใช้เงินจำนวนแน่นอน, ธนาคาร, ผู้รับเงินหรือ "ผู้ถือ", สถานที่ใช้เงิน, วันและสถานที่ออก, ลายมือชื่อผู้สั่งจ่าย) | ประมวลกฎหมายแพ่งและพาณิชย์ — สำเนาจากสถาบันพัฒนาข้าราชการฝ่ายตุลาการ ศาลยุติธรรม https://jla.coj.go.th/th/file/get/file/20260602dfff49c01105cbf669287811f0b3dbe4145952.pdf (มาตรา 987–988) |
| เช็คสั่งจ่ายผู้ถือ | โอนได้เพียงส่งมอบ (มาตรา 918 ใช้กับเช็คตามมาตรา 989) | ฉบับเดียวกัน |
| ขีดคร่อม | เส้นขนานคู่ด้านหน้า = ขีดคร่อมทั่วไป (994); ผู้สั่งจ่ายหรือผู้ทรงขีดคร่อมได้ และเติม "ห้ามเปลี่ยนมือ" ได้ (995 (1)(3)) | ฉบับเดียวกัน มาตรา 994–995 |
| จำนวนเงินตัวอักษรกับตัวเลขไม่ตรงกัน | ให้ถือตัวอักษร (เมื่อหยั่งเจตนาไม่ได้) → ตัวอักษรต้องสร้างจาก decimal ค่าเดียวกับตัวเลข | ฉบับเดียวกัน มาตรา 12 |
| คำว่า "A/C PAYEE ONLY" | **ไม่พบใน ป.พ.พ.** (ข้อความในกฎหมายคือ "ห้ามเปลี่ยนมือ") → เป็นข้อความที่ผู้ใช้พิมพ์เอง | — |

> ไฟล์ทั้งสองถูกดาวน์โหลดและแปลงเป็นข้อความเพื่ออ่าน (`C:/tmp/chqstd/bot.txt`, `ccc.txt`) ไฟล์ทำงานชั่วคราวอยู่นอก repo

---

## 3. สถานะ BC ปัจจุบัน (อ้าง code)

| ส่วน | หลักฐาน | สภาพ |
|---|---|---|
| doctype เช็ค 10 ตัว | `backend/internal/doctype/doctype.go:80` (chequereceived 61 CHQ-IN), `:114-122` (418 CHQ-OUT, 419 CQDP, 420 CQIP, 421 CQIR, 422 CQRD, 423 CQIC, 424 CQDS, 425 CQOP, 426 CQOC) | ใช้เป็น key ของเลขเอกสารและ GL mapping ได้ |
| จอเช็ค 10 จอ | `frontend/src/lib/erp-transaction.ts:387-406` (ทะเบียน), `:647-725` (เอกสารสถานะ) | เป็นเอกสาร ERP ทั่วไป (สินค้า/ยอดรวม) ไม่มีสถานะเช็ค ไม่มีตัวกรอง ไม่อ้างทะเบียน |
| โมเดลเงิน ERP | `backend/internal/transaction/httpapi/models.go:25-62` เป็น `float64` | **ผิดกฎ 20 ห้ามใช้กับเช็ค** |
| ทางเดิมฝั่ง backend ยังเปิด (ตรวจ 2026-10-11) | `backend/internal/transaction/httpapi/http.go:42-58` ลง route `/transaction/:module` (list/get/create/update/delete/approve/cancel); `:196-198` `lookupDocType` → `doctype.ByKey` ซึ่งรู้จัก chequereceived…chequeissuedcancel (`doctype.go:80`, `:114-122`) | ถ้าลบแค่ config หน้าจอ ยังเขียนเช็คเป็นเอกสาร ERP แบบ float ลงตาราง `doc` ได้ → ต้องปิดใน change เดียวกับจอใหม่ (CB-15) |
| GL จาก ERP | `transaction/httpapi/posting.go:170-200` `mappingLines` (total/goods/vat), `:203` `postDocument` → `gl.CheckSourceDateOpen` → `gl.MappingForSourceDoc(dt.Key)` → ไม่มี mapping = `glstatus no_mapping` | ใช้แนวเดียวกันได้ แต่ต้องมีแหล่งยอดของเช็ค |
| ตรวจ mapping | `generalledger/mapping_posting.go:17-19` (แหล่งยอด 3 ตัว), `:87` ตรวจแหล่ง, `:122-139` `mappingBalances` ตรวจโครงสองส่วน [มูลค่า, ภาษี] | ขยายได้: แหล่งยอดของเช็คใช้โครงสองส่วนเดียวกัน [สุทธิ, ค่าใช้จ่าย] |
| `MappingRule` | `generalledger/models.go:220-224` (accountcode, side, source) | ยังไม่มีบัญชีจากสมุดเงินฝาก |
| สมุดเงินฝากฝั่ง GL | `generalledger/subledger.sql:7-10` `gl_subledger_bank_accounts(company, code, version, payload jsonb)`; `subledger_types.go:35-44` (bank_account_code, bank_name, account_number, account_name, gl_account_code, currency_code, is_active) | ไม่มี bank_code, branch หรือประเภทบัญชี; ถูกบันทึกผ่านรายละเอียดรายวัน (`subledger_records.go:183-215`) |
| ข้อกำหนด mydocs | `mydocs/datamodels/gl/bank.sql:3-14` `gl_bank_accounts` (currency `THB` เท่านั้น) | อ่านอย่างเดียว ใช้เป็นกรอบ |
| จอสมุดเงินฝาก | `frontend/src/lib/system-setting-screens.ts:893-922` (`/payment/bookbank` ไม่มี handler ใน Go แล้ว); `menu-screen-status.test.ts:67` `isMenuBackendRetired("/bookbankscreen") === true` | แสดงรอพัฒนา |
| จอเงา | `cheque-movement-screen.tsx`, `bank-service-screen.tsx`, `cheque-issued-prepare-screen.tsx` | ข้อมูลจำลอง → F2 ใน `facade-screens.md` |
| คำอ่านเงินบาท | `backend/internal/whtcert/bahttext.go:16` `BahtText(decimal.Decimal)` (กติกาเดียวกับ `frontend/src/lib/thai-baht-text.ts`) | ใช้ซ้ำได้ |
| PDF | `backend/internal/pdftext/pdftext.go:123-189` `Overlay.NewPage(w,h)`, `Text`, `Fit`, `Check`, `Bytes` | ยังไม่มีเส้นตรง (`Line`) สำหรับขีดคร่อม |
| แบบโมดูลที่จะลอก | `backend/internal/fixedasset` (`records.go:24` ตารางมี `company`, `httpapi/http.go:61-70` `/fa/v2/command` + GET, `:233-238` list = `{success, items, total, page, limit}`), BFF `frontend/src/app/api/fa/[...faPath]/route.ts` | ใช้เป็นต้นแบบ |
| error contract | `backend/pkg/apperr/errors.go:15-27` + `:119-138` → `{success:false, errorcode, code, message, message_th, field, statuscode}` | ใช้ร่วม |
| เมนู | `frontend/src/lib/menu-data.ts:395-418` (เช็ครับ/จ่าย), `:441-446` (สมุดเงินฝาก/บริการ), `:460-475` (เครื่องมือ/รายงาน); `menu-bars.ts:175-197` (แถบ เช็ค/บัตร), `:198-210` (ธนาคาร/เงินสด), `:300-311` (กำหนดรหัสต่างๆ) | id/route/key คงเดิม |

---

## 4. การออกแบบ

### 4.1 หลักการ

1. **โมดูลใหม่ `backend/internal/cheque`:**
   - ไม่ใช้ `ErpTransactionDoc` (เพราะเป็น float64) และไม่ใช้ตาราง `doc`
   - เงินเป็น `numeric(18,2)` และใน Go ใช้ `shopspring/decimal`
   - JSON ส่งเงินเป็นสตริง เช่น `"125000.00"`
2. **ทุกธุรกรรมอยู่ใน `sql.Tx` เดียว:** ทะเบียน + event + GL (`gl.ExecuteSourceTx`) + ลิงก์ ทำครบหรือไม่ทำเลย
   - ล็อกเช็คด้วย `SELECT … FOR UPDATE` ก่อนเปลี่ยนสถานะ
3. **สถานะของเช็คเป็นค่าที่ derive จาก event ที่ยังไม่ถูกยกเลิก:**
   - เก็บ `status` ในทะเบียนเป็น cache ที่อัปเดตใน tx เดียวกัน
   - `cheque-recalculate` คำนวณใหม่จาก event และรายงานผลต่าง
4. **ไม่ลบ ไม่แก้เงียบ:**
   - เอกสารเปลี่ยนสถานะที่ลงบัญชีแล้วจะ "ยกเลิก" ทั้งฉบับ
     - กลับรายการ GL ด้วย `gl.SourceReversalDate` + void event
     - เก็บเหตุผล ผู้ทำ และเวลา
   - การแก้ไข = ยกเลิกฉบับเดิม (revision) + สร้างฉบับใหม่ แบบเดียวกับ `erp_doc_journals`
5. **ชื่อทุกตัวเป็น lowercase snake_case** และใช้ parameterized SQL ทั้งหมด
6. **สร้างโครงสร้างด้วย `ensureSchema`** (`CREATE … IF NOT EXISTS` / `ALTER … ADD COLUMN IF NOT EXISTS`) ไม่มี migration (ช่วง dev)
7. **ข้อความทุกตัวเป็น key** ใน `languages.tsv` (13 คอลัมน์) และ backend ตอบด้วย `language.Text(key, lang)`
8. **ไม่มีเงื่อนไขตามรหัสบัญชี:** บัญชีมาจากรูปแบบลงบัญชีที่ผู้ใช้ตั้ง หรือจาก `gl_account_code` ของสมุดเงินฝาก

### 4.2 DDL (PostgreSQL; ไฟล์ `backend/internal/cheque/schema.go`)

```sql
-- ทะเบียนรหัส 5 ชนิด (ตาม "กำหนดรหัสต่างๆ" myfrmbasemaster.cpp:229-236)
CREATE TABLE IF NOT EXISTS bank_codes (
  company     text         NOT NULL,
  kind        text         NOT NULL CHECK (kind IN ('bank','bank_branch','bank_account_type','bank_service','cheque_ext_status')),
  code        varchar(10)  NOT NULL CHECK (btrim(code) <> ''),
  name        varchar(50)  NOT NULL CHECK (btrim(name) <> ''),
  bank_code   varchar(10),                     -- เฉพาะ bank_branch: ตัวกรองไม่บังคับ (ส่วนเสริม) Champ ไม่มี
  user_group  varchar(255) NOT NULL DEFAULT '',
  is_active   boolean      NOT NULL DEFAULT true,
  version     bigint       NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at  timestamptz  NOT NULL DEFAULT now(), created_by text NOT NULL,
  updated_at  timestamptz  NOT NULL DEFAULT now(), updated_by text NOT NULL,
  PRIMARY KEY (company, kind, code),
  CHECK (kind = 'bank_branch' OR bank_code IS NULL)
);

-- ทะเบียนเช็ค (BCChqIn + BCChqOut รวมตารางเดียว แยกด้วย cheque_type)
CREATE TABLE IF NOT EXISTS cheque_registers (
  company          text          NOT NULL,
  id               uuid          NOT NULL,
  cheque_type      smallint      NOT NULL CHECK (cheque_type IN (0,1)),       -- 0 = เช็ครับ, 1 = เช็คจ่าย
  doc_no           varchar(30)   NOT NULL,                                     -- เลขที่เอกสาร (CHQ-IN / CHQ-OUT)
  doc_date         date          NOT NULL,                                     -- วันที่รับเช็ค / วันที่จ่าย
  due_date         date          NOT NULL,                                     -- วันที่ถึงกำหนด
  partner_code     text          NOT NULL,                                     -- ลูกหนี้ / เจ้าหนี้ (gl_subledger_partners)
  cheque_no        varchar(20)   NOT NULL CHECK (btrim(cheque_no) <> ''),
  bank_code        varchar(10)   NOT NULL,
  bank_branch_code varchar(10)   NOT NULL,
  book_code        text,                                                        -- สมุดเงินฝาก (gl_subledger_bank_accounts.code)
  status           smallint      NOT NULL DEFAULT 0,
  ext_status_code  varchar(10),                                                -- สถานะเพิ่มเติม (bank_codes kind cheque_ext_status)
  currency_code    varchar(3)    NOT NULL DEFAULT 'THB',                      -- คงคอลัมน์ตาม Champ; จนกว่า D-FX จะตัดสิน backend รับเฉพาะ 'THB' (mydocs bank.sql:10)
  exchange_rate    numeric(19,10) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0), -- จนกว่า D-FX จะตัดสิน backend รับเฉพาะ 1 (Champ บังคับ 1 เมื่อบริษัทไม่เปิดหลายสกุลเงิน cqfrmcheqin.cpp:1135)
  amount           numeric(18,2) NOT NULL CHECK (amount > 0),                 -- จำนวนเงินบนเช็คตามสกุลเงิน
  home_amount      numeric(18,2) NOT NULL CHECK (home_amount > 0),            -- เงินบาท = ปัดครึ่งขึ้น(amount*rate) ตามทศนิยมเงิน (Champ XRound cqfrmcheqindeposit.cpp:475-476)
  description      varchar(255)  NOT NULL DEFAULT '',
  sale_code        text          NOT NULL DEFAULT '',                          -- เฉพาะเช็ครับ
  department_code  text          NOT NULL DEFAULT '',
  source_kind      text          NOT NULL DEFAULT 'register' CHECK (source_kind IN ('register','erp','opening')),
  prepare_doc_no   varchar(30),                                                -- มาจากเลขเตรียมจ่าย (B3)
  print_count      integer       NOT NULL DEFAULT 0 CHECK (print_count >= 0),
  version          bigint        NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL DEFAULT now(), created_by text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(), updated_by text NOT NULL,
  PRIMARY KEY (company, id),
  CHECK ((cheque_type = 0 AND status BETWEEN 0 AND 6) OR (cheque_type = 1 AND status BETWEEN 0 AND 3)),
  CHECK (cheque_type = 0 OR (book_code IS NOT NULL AND doc_date <= due_date)),   -- เช็คจ่าย: สมุดบังคับ, วันที่จ่าย <= ถึงกำหนด
  CHECK (cheque_type = 1 OR status = 0 OR book_code IS NOT NULL),               -- เช็ครับ: สถานะ <> 0 ต้องมีสมุด
  CHECK (cheque_type = 0 OR sale_code = '')
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_in_no   ON cheque_registers (company, cheque_no, bank_code, bank_branch_code) WHERE cheque_type = 0;
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_out_no  ON cheque_registers (company, cheque_no, book_code) WHERE cheque_type = 1;
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_doc_no  ON cheque_registers (company, cheque_type, doc_no);
CREATE INDEX IF NOT EXISTS ix_cheque_select ON cheque_registers (company, cheque_type, status, due_date);
CREATE INDEX IF NOT EXISTS ix_cheque_partner ON cheque_registers (company, cheque_type, partner_code);

-- ประวัติสถานะ (append-only; ยกเลิกด้วย voided_*)
CREATE TABLE IF NOT EXISTS cheque_events (
  company       text        NOT NULL,
  id            bigserial,
  cheque_id     uuid        NOT NULL,
  event_kind    text        NOT NULL CHECK (event_kind IN ('create','update','status_doc','print','ext_status')),
  doc_flag      integer,                                   -- transflag ของเอกสารสถานะ 419..426
  doc_no        varchar(30),
  event_date    date        NOT NULL,
  from_status   smallint,
  to_status     smallint,
  book_code     text,
  detail        jsonb       NOT NULL DEFAULT '{}'::jsonb,  -- ฟิลด์ที่เปลี่ยน (ก่อน/หลัง)
  reason        text        NOT NULL DEFAULT '',
  actor         text        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  voided_at     timestamptz, voided_by text, void_reason text,
  PRIMARY KEY (company, id),
  FOREIGN KEY (company, cheque_id) REFERENCES cheque_registers (company, id),
  CHECK ((voided_at IS NULL AND voided_by IS NULL AND void_reason IS NULL) OR (voided_at IS NOT NULL AND voided_by IS NOT NULL AND btrim(void_reason) <> ''))
);
CREATE INDEX IF NOT EXISTS ix_cheque_events_cheque ON cheque_events (company, cheque_id, event_date DESC, id DESC) WHERE voided_at IS NULL;

-- เอกสารเปลี่ยนสถานะ 8 ตัว
CREATE TABLE IF NOT EXISTS cheque_docs (
  company        text          NOT NULL,
  transflag      integer       NOT NULL CHECK (transflag BETWEEN 419 AND 426),
  doc_no         varchar(30)   NOT NULL,
  doc_date       date          NOT NULL,
  branch_code    text          NOT NULL,                     -- scope.Branch ตอนลงบัญชี
  book_code      text,
  cancel_type    smallint CHECK (cancel_type BETWEEN 0 AND 3), -- 423 เท่านั้น (CancelStatus)
  cancel_passed  boolean       NOT NULL DEFAULT false,         -- 426 "ยกเลิกเช็คที่ผ่านแล้ว"
  description    varchar(255)  NOT NULL DEFAULT '',
  total_amount   numeric(18,2) NOT NULL CHECK (total_amount > 0),
  fee_amount     numeric(18,2) NOT NULL DEFAULT 0 CHECK (fee_amount >= 0),
  net_amount     numeric(18,2) NOT NULL,
  gl_status      text          NOT NULL DEFAULT 'none' CHECK (gl_status IN ('none','posted','no_mapping','not_required')),
  status         text          NOT NULL DEFAULT 'active' CHECK (status IN ('active','cancelled')),
  cancel_reason  text, cancelled_at timestamptz, cancelled_by text,
  version        bigint        NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL DEFAULT now(), created_by text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(), updated_by text NOT NULL,
  PRIMARY KEY (company, transflag, doc_no),
  CHECK (net_amount = total_amount - fee_amount),
  CHECK (fee_amount <= total_amount),
  CHECK (transflag <> 419 OR fee_amount = 0),                 -- นำฝากไม่มีค่าใช้จ่าย (จอ 5398 ไม่มีคอลัมน์ค่าใช้จ่าย)
  CHECK ((transflag = 423) = (cancel_type IS NOT NULL)),
  CHECK (transflag = 426 OR cancel_passed = false),
  CHECK (status = 'active' OR (cancelled_at IS NOT NULL AND btrim(cancel_reason) <> ''))
);
CREATE TABLE IF NOT EXISTS cheque_doc_lines (
  company     text          NOT NULL,
  transflag   integer       NOT NULL,
  doc_no      varchar(30)   NOT NULL,
  line_no     integer       NOT NULL CHECK (line_no > 0),
  cheque_id   uuid          NOT NULL,
  from_status smallint      NOT NULL,                          -- ใช้คืนสถานะ/ตรวจลำดับ
  amount      numeric(18,2) NOT NULL CHECK (amount > 0),       -- = home_amount ของเช็ค
  fee_amount  numeric(18,2) NOT NULL DEFAULT 0 CHECK (fee_amount >= 0 AND fee_amount <= amount),
  net_amount  numeric(18,2) NOT NULL,
  PRIMARY KEY (company, transflag, doc_no, line_no),
  FOREIGN KEY (company, transflag, doc_no) REFERENCES cheque_docs (company, transflag, doc_no),
  FOREIGN KEY (company, cheque_id) REFERENCES cheque_registers (company, id),
  UNIQUE (company, transflag, doc_no, cheque_id),
  CHECK (net_amount = amount - fee_amount),
  CHECK (transflag <> 419 OR fee_amount = 0)
);
-- ใบสำคัญ GL ต่อรอบแก้ไข (แบบเดียวกับ erp_doc_journals schema.go:206-218)
CREATE TABLE IF NOT EXISTS cheque_doc_journals (
  company text NOT NULL, transflag integer NOT NULL, doc_no varchar(30) NOT NULL, revision integer NOT NULL,
  journal_id text NOT NULL, journal_docno text NOT NULL, voided_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (company, transflag, doc_no, revision)
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_doc_journals_active ON cheque_doc_journals (company, transflag, doc_no) WHERE voided_at IS NULL;

-- ลิงก์เช็ค ↔ เอกสารต้นทาง (ใบเสร็จ/ใบจ่าย/รายวัน) แทน BCRecMoney/BCPayMoney + Balance
CREATE TABLE IF NOT EXISTS cheque_source_links (
  company      text          NOT NULL,
  id           uuid          NOT NULL,
  cheque_id    uuid          NOT NULL,
  source_kind  text          NOT NULL CHECK (source_kind IN ('erp_doc','gl_journal')),
  source_flag  integer,                                   -- transflag ของ ERP
  source_doc_no text         NOT NULL,
  source_date  date          NOT NULL,
  amount       numeric(18,2) NOT NULL CHECK (amount > 0),  -- ยอดจัดสรรของเอกสารนี้
  created_at timestamptz NOT NULL DEFAULT now(), created_by text NOT NULL,
  voided_at timestamptz, voided_by text, void_reason text,
  PRIMARY KEY (company, id),
  FOREIGN KEY (company, cheque_id) REFERENCES cheque_registers (company, id)
);
CREATE INDEX IF NOT EXISTS ix_cheque_links_cheque ON cheque_source_links (company, cheque_id) WHERE voided_at IS NULL;

-- B3 ใบเตรียมเลขเช็ค (BCChqOutAuto / BCChqOutAutoSub)
CREATE TABLE IF NOT EXISTS cheque_prepare_docs (
  company text NOT NULL, doc_no varchar(30) NOT NULL, doc_date date NOT NULL,
  bank_code varchar(10) NOT NULL, bank_branch_code varchar(10) NOT NULL, book_code text NOT NULL,
  start_no varchar(20) NOT NULL, run_count smallint NOT NULL CHECK (run_count BETWEEN 1 AND 32767),
  description varchar(255) NOT NULL DEFAULT '',
  version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at timestamptz NOT NULL DEFAULT now(), created_by text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(), updated_by text NOT NULL,
  PRIMARY KEY (company, doc_no)
);
CREATE TABLE IF NOT EXISTS cheque_prepare_numbers (
  company text NOT NULL, doc_no varchar(30) NOT NULL, line_no integer NOT NULL CHECK (line_no > 0),
  cheque_no varchar(20) NOT NULL, book_code text NOT NULL,
  used_cheque_id uuid,                                     -- แทน IsUse (NULL = ยังไม่ใช้)
  PRIMARY KEY (company, doc_no, line_no),
  FOREIGN KEY (company, doc_no) REFERENCES cheque_prepare_docs (company, doc_no) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_prepare_no ON cheque_prepare_numbers (company, book_code, cheque_no);

-- B4 รายละเอียดธนาคาร (บรรทัดข้อความต่อธนาคารต่อแท็บ) + log
CREATE TABLE IF NOT EXISTS bank_service_notes (
  company text NOT NULL, bank_code varchar(10) NOT NULL,
  tab text NOT NULL CHECK (tab IN ('interest','exchange','service')),
  line_no integer NOT NULL CHECK (line_no > 0),
  detail varchar(255) NOT NULL,
  PRIMARY KEY (company, bank_code, tab, line_no)
);
CREATE TABLE IF NOT EXISTS bank_service_note_logs (
  company text NOT NULL, id bigserial, bank_code varchar(10) NOT NULL, tab text NOT NULL,
  before_lines jsonb NOT NULL, actor text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (company, id)
);

-- B1 ตำแหน่งพิมพ์ต่อสมุดเงินฝาก (มิลลิเมตร)
CREATE TABLE IF NOT EXISTS cheque_print_layouts (
  company     text         NOT NULL,
  book_code   text         NOT NULL,
  page_width_mm  numeric(6,2) NOT NULL CHECK (page_width_mm  BETWEEN 50 AND 400),
  page_height_mm numeric(6,2) NOT NULL CHECK (page_height_mm BETWEEN 50 AND 400),
  offset_x_mm numeric(6,2) NOT NULL DEFAULT 0,  -- ปรับทั้งแผ่นตามเครื่องพิมพ์
  offset_y_mm numeric(6,2) NOT NULL DEFAULT 0,
  fields      jsonb        NOT NULL,  -- {"date":{x,y,w,size,align},"payee":{…},"amount_words":{…},"amount":{…}}
  date_format text         NOT NULL CHECK (date_format IN ('ddmmyyyy')),
  date_era    text         NOT NULL CHECK (date_era IN ('be','ce')),   -- ต้องเลือกให้ตรง Date Field Indicator ของธนาคาร
  date_box_pitch_mm numeric(5,2),                                        -- ระยะช่องวันที่ 8 ช่อง (NULL = พิมพ์ต่อกัน)
  wrap_parens boolean      NOT NULL DEFAULT true,                        -- "( … )" ตาม XMoneyWord
  amount_prefix text       NOT NULL DEFAULT '',  amount_suffix text NOT NULL DEFAULT '',
  cross_lines boolean      NOT NULL DEFAULT false,                       -- ขีดคร่อม (ป.พ.พ. 994)
  cross_text  varchar(40)  NOT NULL DEFAULT '',                          -- เช่น ห้ามเปลี่ยนมือ / A/C PAYEE ONLY (ผู้ใช้พิมพ์เอง)
  strike_bearer boolean    NOT NULL DEFAULT false,                       -- ขีดฆ่า "หรือผู้ถือ"
  bearer_strike jsonb,                                                   -- {x1,y1,x2,y2} มม.
  version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
  updated_at timestamptz NOT NULL DEFAULT now(), updated_by text NOT NULL,
  PRIMARY KEY (company, book_code)
);

-- รายละเอียดเช็คใต้รายวัน GL (BCGLChq) — อยู่ใน generalledger/subledger.sql
CREATE TABLE IF NOT EXISTS gl_journal_cheques (
  company text NOT NULL, journal_id text NOT NULL, line_no integer NOT NULL CHECK (line_no > 0),
  cheque_type smallint NOT NULL CHECK (cheque_type IN (0,1)),
  cheque_no varchar(20) NOT NULL, bank_code varchar(10) NOT NULL, bank_branch_code varchar(10) NOT NULL DEFAULT '',
  book_code text, due_date date NOT NULL,
  amount numeric(18,2) NOT NULL CHECK (amount > 0), exchange_rate numeric(19,10) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
  PRIMARY KEY (company, journal_id, line_no)
);
```

**ส่วนขยายสมุดเงินฝาก**
- ไม่มี DDL ใหม่ เพราะ `gl_subledger_bank_accounts.payload` เป็น jsonb
- เพิ่มฟิลด์ใน `SubledgerBankAccount` (`subledger_types.go:35-44`):
  - `bank_code`
  - `bank_branch_code`
  - `account_type_code`
- ตรวจกับ `bank_codes` ใน Go ตอนบันทึก
- คงฟิลด์ `bank_name` ไว้เป็นข้อความแสดงผล (อ่านมาจากรหัส)
- `currency_code` ของสมุด: จนกว่า D-FX จะตัดสิน backend รับเฉพาะ `THB` (ตรงกับ `mydocs/datamodels/gl/bank.sql:10`)
- **"เปลี่ยนเลขที่บัญชี" (ตาม Champ `CASHDlgChangeBookNo.cpp:59-130`):** ใน BC คีย์อ้างอิงคือ `bank_account_code` (`code`) ซึ่งไม่เปลี่ยน ส่วนเลขบัญชีธนาคารเก็บใน `account_number`
  - ปุ่มเปิดกล่องยืนยัน เลขเดิม → เลขใหม่ (ห้ามว่าง ห้ามซ้ำกับสมุดอื่น) แล้วแก้ `account_number` + บันทึกประวัติแก้ไข ใน tx เดียว
  - **เบี่ยงจาก Champ โดยตั้งใจ:** ไม่ต้องไล่ UPDATE ทุกตารางแบบ Champ เพราะเอกสารอ้างด้วยรหัสที่ไม่เปลี่ยน (ผลต่อผู้ใช้เหมือนเดิม)
- **ตารางผู้ติดต่อบนจอสมุด:** แสดงอย่างเดียวจากผู้ติดต่อของ ธนาคาร+สาขา ของสมุด (Champ `CASHFrmBankBookNew.cpp:292-296`); ข้อมูลมาจากจอ 5680 → work item CB-16 (ตาราง `bank_contacts` คีย์ company+bank_code+bank_branch_code+contact_code; คอลัมน์ ติดต่อ/ไม่ติดต่อ, รหัส, ชื่อ/รายละเอียด, ตำแหน่ง, ที่อยู่, โทรศัพท์ ตาม DDL `BCBankContact` `:10588-10614`)
  - ถ้า CB-16 ยังไม่เสร็จ ตารางบนจอสมุดแสดงว่างพร้อมข้อความ "ยังไม่มีผู้ติดต่อ" (ไม่ซ่อนตาราง เพื่อให้ผังจอตรง Champ)

**การตรวจข้ามตารางที่ทำใน Go** (เพราะอยู่ใน payload jsonb หรือข้ามโมดูล)
- partner มีอยู่และเป็นลูกหนี้หรือเจ้าหนี้ตามชนิดเช็ค
- book_code มีอยู่และใช้งานอยู่
- bank/branch/ext_status/account_type มีอยู่ใน `bank_codes`

### 4.3 เครื่องสถานะและกติกา (backend `cheque/state.go`)

| transflag | เลือกเช็คได้ (ชนิด, สถานะ) | เงื่อนไขวันที่ | สถานะใหม่ | สมุดเงินฝาก | ลงบัญชี |
|---|---|---|---|---|---|
| 419 นำฝาก | รับ, {0} | วันที่รับ ≤ วันที่เอกสาร | 1 | บังคับ | `not_required` |
| 420 รับผ่าน | รับ, {0,1} | ถึงกำหนด ≤ วันที่เอกสาร | 2 | บังคับ | mapping |
| 421 รับคืน | รับ, {2,6} | วันที่รับ ≤ วันที่เอกสาร | 3 | ไม่บังคับ | mapping |
| 422 นำเข้าใหม่ | รับ, {3} + สมุดของเช็คตรงกับสมุดของเอกสาร | วันที่รับ ≤ วันที่เอกสาร (`cqfrmcheqindeposit2.cpp:721,724`) | 6 | บังคับ (`:711-714`) | mapping |
| 423 ยกเลิกรับ | รับ, {0 \| 2 \| 1 \| 6} ตาม cancel_type 0/1/2/3 + สมุดตรงกันเมื่อ cancel_type ≠ 0 | วันที่รับ ≤ วันที่เอกสาร ทุกชนิด (`cqfrmcheqincancel.cpp:757-789`) | 4 | บังคับเมื่อ cancel_type ∈ {1,2,3}; ชนิด 0 ไม่ใช้สมุด | mapping |
| 424 ขายลด | รับ, {0,1} | วันที่รับ ≤ วันที่เอกสาร | 5 | บังคับ | mapping |
| 425 จ่ายผ่าน | จ่าย, {0} | ถึงกำหนด ≤ วันที่เอกสาร | 1 | กรองได้ (ถ้าระบุต้องตรงสมุดของเช็ค) | mapping |
| 426 ยกเลิกจ่าย | จ่าย, {0} หรือ {1} เมื่อ cancel_passed | — | 2 | บังคับเมื่อ cancel_passed | mapping |

**กติกาที่ backend บังคับ** (ทุกข้อคืน error key และระบุ field)
1. **เช็คอยู่ในเอกสารได้ครั้งเดียว** (`cheque_err_duplicate_line`)
   - และต้องอยู่ในสถานะที่เลือกได้ ณ ตอนบันทึก (`cheque_err_status_not_selectable`) เพราะตรวจใหม่ใน tx หลังล็อกแถว
2. **เงื่อนไขวันที่ตามตาราง:** `cheque_err_not_due`, `cheque_err_received_after_doc`
3. **ยอด:**
   - ค่าใช้จ่ายต่อแถวอยู่ในช่วง 0..amount (`cheque_err_fee_invalid`)
   - ยอดหัวเอกสาร**คำนวณใหม่ที่ backend** ถ้า client ส่งยอดมาแล้วไม่ตรงกัน → `cheque_err_totals_mismatch` (เทียบเท่า MSG_CHQ_FEEAMOUNT_ERR)
4. **สมุดเงินฝาก:** ใส่ `book_code` ของเอกสารลงทุกเช็ค (เหมือน `BookNo` ของ Champ)
   - เมื่อสถานะกลับเป็น 0 สำหรับเช็ครับ ให้ล้าง book_code
5. **ลำดับเหตุการณ์ (ตาม Champ — แก้ 2026-10-11 เดิมเขียนผิดว่าเบี่ยง):**
   - ห้ามยกเลิกหรือแก้เอกสารสถานะ ถ้าเช็คในเอกสารมี event ที่ใหม่กว่าและยังไม่ถูกยกเลิก (`cheque_err_newer_event`)
   - ต้องยกเลิกเอกสารที่ใหม่กว่าก่อน
   - Champ ห้ามแบบเดียวกันด้วย `IsChqReference` (นำฝาก `cqfrmcheqindeposit.cpp:117-125`, `:186-194`, `:1521-1550`; รับผ่าน `cqfrmcheqinpass.cpp:316-321`, `:439-444`, `:1623-1650`; รับคืน `cqfrmcheqinreturn.cpp:1621-1641`) และแถวที่เช็คสถานะ > 2 แล้ว CANNOT_EDIT_REFERENCE (`cqfrmcheqinpass.cpp:399-418`)
   - ลำดับตรวจตอนบันทึกเรียงตาม Champ §1.2: อ้างอิง → งวดปิด (`period_closed`) → ผ่าน GL แล้ว (BC: ลงบัญชีพร้อมบันทึก จึงใช้ revision + กลับรายการแทน ดู §4.4) → (อนุมัติแล้ว: Champ มีตัวกันแต่ไม่มีโค้ดตั้งค่า จึงไม่ทำ ดู §7) → เลขที่/วันที่ → ยอด/แถว → สมุด → สถานะเช็ค → ยอดรวม
6. **ยกเลิกเอกสารสถานะ:**
   - void event ทุกแถว แล้วคำนวณสถานะใหม่
     - ใช้ event ล่าสุดที่ยังไม่ void (`event_date DESC, id DESC`) เทียบเท่า `GetChqInStatus`
     - ถ้าไม่มี event สถานะ = 0
   - กลับรายการ GL (`gl.SourceReversalDate`, ใช้วันเดิมถ้างวดเปิด ถ้าไม่เปิดใช้วันทำงาน)
   - บันทึก `cancel_reason`
7. **ทะเบียนเช็คที่ถูกอ้างแล้ว** (มี event สถานะหรือ source link ที่ยังไม่ void):
   - **ห้ามลบ = ตาม Champ** MSG_REFERANCE (`cqfrmcheqin.cpp:438-443`) → `cheque_err_referenced_delete`; ให้ยกเลิกเอกสารที่อ้างก่อน
   - **เช็คที่เกิดจากเอกสารต้นทาง (`source_kind = 'erp'` เทียบ `SaveFrom ≠ 4`) ห้ามลบจากจอทะเบียน = ตาม Champ** MSG_FROM_OTHER (`:444-449`) → `cheque_err_from_source_doc`; ต้องลบ/ยกเลิกที่เอกสารต้นทาง
   - **แก้ฟิลด์หลัก** cheque_no / bank / branch / book / amount / rate / partner / type: Champ แค่ถามยืนยัน CANNOT_UPDATE_REFERENCE (`:320-327`) / CANNOT_UPDATE_REFERENCE_OTHER (`:328-336`); **D8 ตัดสินแล้ว 2026-10-11 = ทำตาม Champ**: ถามยืนยันแล้วยอมให้แก้; ทุกการแก้บันทึก event `edit` (ค่าเดิม→ค่าใหม่ ผู้แก้ เวลา) และยอดคงเหลือ derive ใหม่ทันที (`balances.go`/recalculate) แทนการล็อก
   - แก้ได้เสมอ: description, ext_status, sale, department; due_date แก้ได้เฉพาะเมื่อยังไม่มี event สถานะ
   - ทะเบียนไม่ตรวจงวดปิด (Champ ไม่ตรวจ เพราะทะเบียนไม่ลงบัญชี)
8. **สถานะในจอทะเบียน:**
   - เลือกได้เฉพาะตอนสร้าง (ยกมา/เช็คที่ไม่มีเอกสารต้นทาง)
     - เช็ครับ: 0..6 แต่ ≠ 0 ต้องมีสมุด
     - เช็คจ่าย: 0 หรือ 3 เตรียมจ่าย
   - หลังสร้าง: เปลี่ยนด้วยเอกสารสถานะ (ลงบัญชี) หรือแก้ตรงในจอทะเบียนตาม Champ หลังถามยืนยัน (ไม่ลงบัญชี เหมือน Champ)
   - **D8 ตัดสินแล้ว 2026-10-11 = ทำตาม Champ**: แก้สถานะในจอทะเบียนได้โดยตรง; ประวัติยังครบเพราะบันทึก event `manual_status` (from→to, ผู้แก้, เวลา) แบบ append-only
   - เช็คจ่ายที่สร้างด้วยปุ่ม "เช็คเตรียมจ่าย": สถานะ 3, `prepare_doc_no` + ตั้ง `cheque_prepare_numbers.used_cheque_id` ใน tx เดียว; ลบทะเบียน → คืน `used_cheque_id = NULL` (Champ `cqfrmcheqout.cpp:537-539`, `:451-453`)
9. **ยอดคงเหลือของเช็ค** (เทียบ `Balance` ของ Champ) = `home_amount − Σ cheque_source_links.amount (ที่ยังไม่ void)` → ต้อง ≥ 0 (`cheque_err_over_allocated`)
   - เช็คจ่ายสถานะ 3 เตรียมจ่าย = เช็คที่ออกไว้ก่อนและรอใบจ่ายมาตัดยอด

### 4.4 การลงบัญชี GL

**ขยาย `MappingRule`** (`generalledger/models.go:220`)
```go
type MappingRule struct {
	AccountCode string `json:"accountcode" bson:"accountcode"`
	AccountFrom string `json:"accountfrom,omitempty" bson:"accountfrom,omitempty"` // "" = บัญชีที่ระบุ, "bank_book" = บัญชี GL ของสมุดเงินฝากในเอกสาร
	Side        string `json:"side" bson:"side"`
	Source      string `json:"source" bson:"source"`
}
```
- **แหล่งยอดใหม่** (`mapping_posting.go:17-19`):
  - `cheque_total` (ยอดเช็ครวม = สองส่วน)
  - `cheque_net` (ส่วนที่ 1)
  - `cheque_fee` (ส่วนที่ 2)
  - `mappingBalances` (`:122-139`) ใช้ได้ทันทีถ้าจัดให้ `cheque_total`≡`total`, `cheque_net`≡`goods`, `cheque_fee`≡`vat` ในการนับส่วน
- **ทะเบียนตระกูลแหล่งยอดต่อ source doc:**
  - โมดูลเช็ค register ว่า doctype key 420–426 ใช้แหล่งยอดตระกูล cheque เท่านั้น
  - เอกสาร ERP ใช้ total/goods/vat เท่านั้น
  - ตรวจตอนบันทึก mapping (`mapping_rule_source_invalid`)
- **กติกา `accountfrom: "bank_book"` (ตาม Champ `gltlsprocess.cpp:7766-7824`; ปรับ 2026-10-11):**
  - อนุญาตเฉพาะ source doc ตระกูล cheque; `accountcode` **ใส่ได้เป็นบัญชีสำรอง** (เทียบบัญชีในแม่แบบที่ Champ ตั้งเป็นค่าเริ่มต้น `:7766`)
  - มีสมุดในหัวเอกสาร → ใช้ `gl_account_code` ของสมุด (`:7802-7804`, `:7818-7819`)
  - **ยกเลิกเช็ครับชนิด 0 (เช็คในมือ ไม่มีสมุด) → ใช้บัญชีสำรองใน rule** (`:7824`); ถ้า rule ไม่มีบัญชีสำรอง → `cheque_err_book_no_gl_account`
  - กรณีอื่นที่ไม่มีสมุดหรือสมุดไม่มีบัญชี: Champ ตั้งยอดบรรทัดเป็น 0 แล้วข้ามบรรทัด (`:7806-7809`, `:7821-7822`, `:7870`) ซึ่งทำให้ใบสำคัญไม่สมดุล → **BC เบี่ยงโดยตั้งใจ:** ไม่บันทึกและคืน `cheque_err_book_no_gl_account` (กฎ เดบิต = เครดิต 100%); ปกติไม่เกิดเพราะ 420/421/422/424/425 และ 423 ชนิด 1–3 บังคับสมุด
  - บัญชีนั้นต้อง `allowposting` (ใช้การตรวจ `loadLineAccounts` เดิม)
  - เลือกบัญชีจากข้อมูลสมุด/rule เท่านั้น ไม่มีเงื่อนไขตามรหัสบัญชี
- **ผลต่างปัดเศษ:** ยอดทุกส่วนคำนวณเป็น numeric ทศนิยม 2 ตำแหน่งจากยอดที่ปัดแล้วต่อแถว ผลรวมจึงตรงกันเสมอ ไม่ต้องมีบรรทัดปัดเศษแบบ Champ (`:7916-7958`); ถ้าไม่สมดุล → ไม่บันทึก
- **ตัวอย่างการตั้งค่า** (ผู้ใช้เลือกบัญชีเอง ระบบไม่สร้างรูปแบบเริ่มต้นให้)
  - รับผ่าน (420): เดบิต สมุดเงินฝาก ← `cheque_net` + เดบิต บัญชีค่าธรรมเนียม ← `cheque_fee` / เครดิต บัญชีเช็ครับในมือ ← `cheque_total`
  - ยอดสมดุลตามโครงสองส่วน
- **การโพสต์:**
  - `cheque/posting.go` ทำตามลำดับ `gl.CheckSourceDateOpen` → `gl.MappingForSourceDoc(ctx, tx, scope, dt.Key)` → สร้างบรรทัดจาก rule
    - แต่ละบรรทัดใช้ `department_code` ของเช็ค ถ้าเช็คในเอกสารเป็นแผนกเดียวกันทั้งหมด ถ้าไม่ใช่ใช้ค่าว่าง (open: ยังไม่ verify ว่า Champ แยกแผนกอย่างไร)
    - จากนั้น `gl.ExecuteSourceTx` แล้วบันทึก `cheque_doc_journals`
  - ไม่มี mapping → `gl_status='no_mapping'` และยังบันทึกเอกสารได้ (แนวเดียวกับ ERP `posting.go:203`)
  - 419 → `not_required`
  - **เบี่ยงจาก Champ โดยตั้งใจ (กฎ project: ธุรกรรมบัญชี synchronous ACID):** Champ ลงบัญชีอัตโนมัติตอนบันทึกหรือให้คีย์ใบสำคัญเอง (`GLTransData`) และเมื่อ "ผ่าน GL แล้ว" ห้ามแก้ (POSTGL_WARN `mybasedbformview.cpp:4126-4166`); BC ลงบัญชีใน tx เดียวกับการบันทึก และเอกสารที่ลงบัญชีแล้วแก้ = ยกเลิกใบเดิม (กลับรายการ) แล้วออก revision ใหม่ ไม่ลบเงียบ
- **ไฟล์ที่ชนกับ workflow อื่น:**
  - `source_posting.go` แค่**เรียกใช้** ไม่แก้
  - `mapping_posting.go` และ `models.go` ไม่อยู่ในรายการไฟล์ที่ workflow อื่นแก้ แต่อยู่แพ็กเกจเดียวกัน → รอให้ workflow นั้น commit ก่อน
- **UI รูปแบบลงบัญชี** (`frontend/src/app/gl/gl-masters.tsx:1458-1460` ตัวเลือกแหล่งยอด):
  - เพิ่มตัวเลือก 3 แหล่ง เฉพาะเมื่อ source doc เป็นตระกูลเช็ค
  - เพิ่มช่องเลือก "บัญชีจากสมุดเงินฝาก"

### 4.5 ยอดที่ derive (แทนยอดสะสมของ Champ)

| ยอด | สูตร (SQL บน `cheque_registers` + `cheque_events`) | ผู้ใช้ |
|---|---|---|
| เช็ครับในมือของลูกหนี้ (ChqOnHand) | Σ home_amount เช็ครับ status ∈ {0,1} | รายงานลูกหนี้ / B2 |
| เช็ครับคืนของลูกหนี้ (ChqReturn) | Σ status ∈ {3,6} | รายงานลูกหนี้ |
| เช็คจ่ายค้างของเจ้าหนี้ | Σ เช็คจ่าย status ∈ {0,3} | รายงานเจ้าหนี้ |
| ยอดเช็คที่ผ่านเข้า/ออกสมุดเงินฝาก | Σ net ของ doc_lines ที่ active ต่อ book_code ตามเครื่องหมาย (420/422/424 +, 421/425 −, 423 ตาม cancel_type, 426 + เมื่อ cancel_passed) | ตรวจกับยอด GL ของบัญชีสมุดเงินฝาก |
| ยอดคงเหลือของเช็ค (Balance) | home_amount − Σ links | ใบเสร็จ/ใบจ่ายที่ตัดยอดจากเช็คเดิม |

- **การตรวจวงเงินลูกหนี้ (`DebtLimitBal`):** อยู่นอกขอบเขต (ERP ขายเชื่อ) แต่ใช้สูตรข้างบนได้
- **`cheque-recalculate`:**
  - คำนวณ status จาก event และเทียบกับ cache
  - คืนรายการที่ต่างกัน
  - ปุ่ม "ปรับให้ตรง" เขียน cache + event `update` (reason = recalculate)
- **ผลของ 5402 CancelStatus 2 และ 5405 ที่ไม่ติ๊ก:** การกระทบยอดสมุดยังไม่ verify (`05-cheque.md` §10) → ใช้กติกา "ยกเลิกเช็คที่ยังไม่ผ่านไม่กระทบสมุด"
  - ต้องให้ลุงจืดยืนยันในการทดสอบรับงาน

### 4.6 B1 การพิมพ์เช็คจ่าย

1. **ตำแหน่งพิมพ์:** ตั้งค่า `cheque_print_layouts` ต่อสมุดเงินฝาก (หน่วยมิลลิเมตร)
   - **ไม่มีค่าเริ่มต้นของช่องผู้รับเงินและช่องตัวอักษร** เพราะมาตรฐานไม่ได้กำหนด
   - ช่องวันที่และจำนวนเงินมีค่าแนะนำที่คำนวณจากระยะ "ประมาณ" ของ ธปท. (§2) แสดงเป็นข้อความช่วย ไม่ใช้เป็นค่าเริ่มต้นอัตโนมัติ
2. **ตรวจก่อนบันทึก layout และก่อนพิมพ์** (backend):
   - ทุกช่องต้องอยู่ในหน้า (`cheque_print_err_outside_page`)
   - ทุกช่องต้องอยู่เหนือขอบล่าง 15.875 มม. ทั้งเส้น baseline และส่วนล่างของตัวอักษร (`cheque_print_err_in_micr_band`)
   - ขนาดตัวอักษร ≥ 10 pt (`cheque_print_err_font_too_small`)
   - ช่องที่ต้องพิมพ์แต่ยังไม่ตั้งค่า → `cheque_print_err_field_unset`
3. **สร้าง PDF ด้วย `pdftext.Overlay.NewPage(w×72/25.4, h×72/25.4)`**
   - ตัวเลขจำนวนเงินมาจาก decimal string (`"125,000.00"`)
   - ตัวอักษรใช้ `whtcert.BahtText(home_amount)` (ใช้ฟังก์ชันเดิม ไม่ย้ายแพ็กเกจ = minimal change)
     - ครอบด้วย "( … )" เมื่อ `wrap_parens`
     - ถ้ายาวเกินช่องใช้ `Fit` ลดขนาดได้ถึง 10 pt แล้วตัดเป็นสองบรรทัด
     - ถ้ายังไม่พอ → `cheque_print_err_text_overflow`
   - วันที่ = 8 หลัก DDMMYYYY ตาม `date_era` วางตาม `date_box_pitch_mm`
   - ขีดคร่อม, ข้อความในเส้นคร่อม และการขีดฆ่า "หรือผู้ถือ" ต้องเพิ่ม `Overlay.Line(p, x1, y1, x2, y2, width)` ใน `pdftext` (ตอนนี้ยังไม่มี)
4. **ปุ่ม "พิมพ์ตรวจตำแหน่ง":** พิมพ์ PDF ขนาดเช็คที่มีกรอบของทุกช่อง + เส้นแถบ MICR + ข้อความตัวอย่างลงบนกระดาษเปล่า เพื่อวางทาบเช็คจริง
   - ส่วนเสริมเล็ก — Champ ไม่มี เพราะใช้ form designer
5. **พิมพ์จริง:**
   - เลือกได้เฉพาะเช็คจ่ายที่สถานะ ∈ {0,3} (`cheque_print_err_status`)
   - เงื่อนไขช่วงตาม Champ: ช่วงเลขที่เอกสาร + ช่วงเลขที่เช็ค ห้ามว่าง และเป็นสมุดเงินฝากเดียวกัน (`cheque_print_err_mixed_books`)
   - หลังพิมพ์ทำใน tx: `print_count = print_count + 1` + event `print`
   - พิมพ์ซ้ำได้ (Champ ไม่ล็อก) แต่ UI แสดงจำนวนครั้งที่พิมพ์แล้ว
6. **นอกขอบเขต:** ตัวเลือก GL Slip / แบบภาษีหัก ณ ที่จ่ายใน dialog ของ Champ
   - GL slip ใช้จอพิมพ์รายวันเดิมของ GL
   - หนังสือรับรองหัก ณ ที่จ่ายใช้ `whtcert`
   - ใน dialog เป็นลิงก์ไปจอเหล่านั้น

### 4.7 REST API (backend `/cheque/v1/*`, BFF `/api/cheque/[...chequePath]`)

**รูปแบบ**
- อ่านใช้ `GET` ตอบ `{success:true, items, total, page, limit}` หรือ `{success:true, data}`
- เขียนใช้ `POST /cheque/v1/command` body `{resource, action, id?, version?, requestid, reason?, <payload>}` เหมือน `/fa/v2/command` และ `/gl/v2/command`
  - `requestid` ใช้กันกดซ้ำ (idempotency)
- **Error** = `apperr.Response` (`errors.go:119-138`): `{success:false, errorcode:"cheque_err_…", code, message:(language.Text), message_th, field, statuscode}`
  - 400 = ตรวจไม่ผ่าน
  - 403 = ไม่มีสิทธิ์
  - 404 = ไม่พบ
  - 409 = version หรือสถานะชน
  - 423 = งวดปิด (ใช้ error ของ GL เดิม)
- **สิทธิ์:** ตาม screen id ของเมนู (`cheque-received` ฯลฯ) ผ่าน `glhttp.ResolveSessionScope` แบบ FA

| ฟังก์ชัน | Method + path / command | request (ย่อ) | response | error keys หลัก |
|---|---|---|---|---|
| รายการรหัส | GET `/cheque/v1/codes/:kind?q=&page=&limit=&bank_code=` | — | items `{code,name,bank_code,user_group,is_active,version}` | `bank_code_err_kind_invalid` |
| เพิ่ม/แก้/ลบรหัส | command `codes` `create`/`update`/`delete` | `{kind, code, name, bank_code?, user_group?}` | `data` แถวล่าสุด | `bank_code_err_duplicate`, `bank_code_err_in_use`, `bank_code_err_name_required`, `cheque_err_version_conflict` |
| สมุดเงินฝาก | GET `/gl/v2/bank-accounts`; command GL `resource:"bank-accounts"` `create`/`update`/`delete` (`bankaccount:{…}`) | `{bank_account_code, account_number, account_name, account_type_code, bank_code, bank_branch_code, gl_account_code, currency_code}` (currency_code รับเฉพาะ `THB` จนกว่า D-FX จะตัดสิน) + action `change-account-number` `{bank_account_code, new_account_number}` | `data` | `gl_bank_account_duplicate`, `gl_bank_account_in_use`, `bank_code_err_not_found`, `gl_account_not_postable`, `cheque_err_currency_not_supported`, `bank_book_err_account_number_duplicate` |
| ผู้ติดต่อธนาคาร (CB-16, จอ 5680 + ตารางบนจอสมุด) | GET `/cheque/v1/bank-contacts?bank_code=&bank_branch_code=`; command `bank-contacts` `save` (ทั้งชุดของ ธนาคาร+สาขา) | `{bank_code, bank_branch_code, contacts:[{contact_code, name, position, address, telephone, is_contact}]}` | items | `bank_contact_err_required`, `bank_contact_err_duplicate_code` |
| ทะเบียนเช็ค (รายการ) | GET `/cheque/v1/registers?type=0\|1&status=&partner=&book=&due_from=&due_to=&q=&page=&limit=` | — | items + `balance_amount` | — |
| ทะเบียนเช็ค (ฉบับ) | GET `/cheque/v1/registers/:id` | — | data + `last_event` ("การปรับปรุงเอกสาร") + `references[]` `{kind_key, doc_no, doc_date, amount, description}` สำหรับตารางอ้างอิงใต้หัวเอกสาร (จาก `cheque_source_links` + `cheque_events` ที่ยังไม่ void; เทียบ `cqfrmcheqin.cpp:1227-1300`) | `cheque_err_not_found` |
| สร้าง/แก้/ยกเลิกทะเบียน | command `registers` `create`/`update`/`cancel` | ฟิลด์ตาม §4.2 (เงินเป็นสตริง) | data | `cheque_err_required_field`, `cheque_err_duplicate_number`, `cheque_err_due_before_issue`, `cheque_err_book_required`, `cheque_err_amount_invalid`, `cheque_err_rate_invalid`, `cheque_err_referenced`, `cheque_err_referenced_delete`, `cheque_err_from_source_doc`, `cheque_err_currency_not_supported`, `cheque_err_partner_invalid` |
| เช็คที่เลือกได้ | GET `/cheque/v1/registers/selectable?transflag=420&doc_date=2569-..&book=&cancel_type=` (วันที่ส่งเป็น ISO `YYYY-MM-DD`) | — | items (ตัวกรองตาม §4.3) | `cheque_err_transflag_invalid` |
| การเคลื่อนไหว (B2) | GET `/cheque/v1/registers/:id/movement` และ GET `/cheque/v1/movement?type=&cheque_no=&bank_code=&bank_branch_code=&book=` | — | `{cheque, source_docs[], status_docs[]}` | `cheque_err_not_found`, `cheque_err_ambiguous_number` (เลขเดียวกันหลายธนาคาร → คืนตัวเลือก) |
| เอกสารสถานะ (รายการ/ฉบับ) | GET `/cheque/v1/docs/:transflag`, GET `/cheque/v1/docs/:transflag/:doc_no` | — | หัว + lines + gl (journal docno, status) | `cheque_err_doc_not_found` |
| บันทึก/แก้เอกสารสถานะ | command `docs` `create`/`update` | `{transflag, doc_no?, doc_date, book_code?, cancel_type?, cancel_passed?, description, lines:[{cheque_id, fee_amount}], totals?}` | data + `gl_status` | `cheque_err_status_not_selectable`, `cheque_err_not_due`, `cheque_err_received_after_doc`, `cheque_err_fee_invalid`, `cheque_err_totals_mismatch`, `cheque_err_duplicate_line`, `cheque_err_no_lines`, `cheque_err_book_required`, `cheque_err_cancel_type_invalid`, `cheque_err_book_no_gl_account`, `gl_period_locked` |
| ยกเลิกเอกสารสถานะ | command `docs` `cancel` | `{transflag, doc_no, version, reason}` | data | `cheque_err_newer_event`, `cheque_err_reason_required`, `cheque_err_already_cancelled` |
| ประวัติเอกสาร | GET `/cheque/v1/docs/:transflag/:doc_no/events` | — | items | — |
| ใบเตรียมเลข (B3) | GET `/cheque/v1/prepare`, GET `/cheque/v1/prepare/:doc_no`; command `prepare` `generate` (พรีวิว ไม่บันทึก) / `create` / `update` / `delete` | `{doc_no, doc_date, bank_code, bank_branch_code, book_code, start_no, run_count, description, numbers?:[…]}` | data + numbers `{cheque_no, used:true/false, cheque_doc_no}` | `cheque_prep_err_start_not_numeric`, `cheque_prep_err_overflow`, `cheque_prep_err_count_range`, `cheque_prep_err_number_exists`, `cheque_prep_err_number_used`, `cheque_prep_err_book_mismatch` |
| เลขที่ยังไม่ใช้ | GET `/cheque/v1/prepare/unused?book=` | — | items `{cheque_no, book_code, bank_code, bank_branch_code, doc_no}` | — |
| รายละเอียดธนาคาร (B4) | GET `/cheque/v1/bank-service/:bank_code`; command `bank-service` `save` | `{bank_code, tab, lines:[string ≤255]}` | data | `bank_service_err_bank_not_found`, `bank_service_err_line_too_long`, `bank_service_err_tab_invalid` |
| log รายละเอียด | GET `/cheque/v1/bank-service/:bank_code/logs?tab=` | — | items | — |
| ตำแหน่งพิมพ์ (B1) | GET `/cheque/v1/print-layouts/:book_code`; command `print-layouts` `save` | ฟิลด์ §4.2 | data | `cheque_print_err_outside_page`, `cheque_print_err_in_micr_band`, `cheque_print_err_font_too_small`, `cheque_print_err_era_required` |
| พิมพ์ตรวจตำแหน่ง | GET `/cheque/v1/print-layouts/:book_code/calibration.pdf` | — | `application/pdf` | `cheque_print_err_layout_missing` |
| พิมพ์เช็ค | POST `/cheque/v1/print` | `{doc_no_from, doc_no_to, cheque_no_from, cheque_no_to, book_code}` | `application/pdf` (+ อัปเดต print_count ใน tx) | `cheque_print_err_range_required`, `cheque_print_err_status`, `cheque_print_err_mixed_books`, `cheque_print_err_field_unset`, `cheque_print_err_text_overflow` |
| คำนวณยอดเช็คใหม่ | GET `/cheque/v1/recalculate` (ตรวจ); command `recalculate` `apply` | `{reason}` | `{checked, mismatches[]}` | `cheque_err_reason_required` |

**ตัวอย่าง** (บันทึกเช็ครับผ่าน)
```json
POST /cheque/v1/command
{"resource":"docs","action":"create","requestid":"5b0c7a1e-…",
 "transflag":420,"doc_date":"2026-10-11","book_code":"KB-SAV-01","description":"เช็คลูกค้าผ่านเข้าบัญชีออมทรัพย์",
 "lines":[{"cheque_id":"0f4c…","fee_amount":"10.00"}]}
→ 200 {"success":true,"data":{"doc_no":"CQIP6910-0003","total_amount":"125000.00","fee_amount":"10.00","net_amount":"124990.00","gl_status":"posted","journal_docno":"JV6910-0042"}}
→ 400 {"success":false,"errorcode":"cheque_err_not_due","code":"cheque_err_not_due","message":"เช็คเลขที่ 10004521 ยังไม่ถึงวันที่ถึงกำหนด (15/10/2569)","field":"lines[0].cheque_id","statuscode":400}
```

- **MCP:** โครงคำสั่งพร้อมต่อ MCP แต่**ยังไม่ทำ** (memory: งาน MCP รอลุงจืดสั่ง)
- **BFF ใหม่ `frontend/src/app/api/cheque/[...chequePath]/route.ts`:**
  - ไฟล์ใหม่ ไม่แตะ `erp-transaction` / `goapi` route ที่ workflow อื่นแก้
  - ใช้ allowlist ของ path แบบ `api/fa`

### 4.8 Frontend: จอและเมนู (id/route/key เดิมทั้งหมด)

**ทะเบียนเช็ค**

| เมนู id | route | จอ | หมายเหตุ |
|---|---|---|---|
| `cheque-received` / `cheque-issued` | `/transaction/chequereceived`, `/transaction/chequeissued` | `components/cheque-register-screen.tsx` (prop `chequeType`) | ลำดับฟิลด์และถ้อยคำตาม §1.1 (screenshot); **ตารางอ้างอิงใต้หัวเอกสาร** (รายการ \| เลขที่เอกสาร \| วัน \| จำนวนเงิน \| คำอธิบาย) อ่านอย่างเดียวจาก `references[]`; **ปุ่ม "เช็คเตรียมจ่าย"** บนจอเช็คจ่าย เปิดรายการจาก `/prepare/unused`; เปลี่ยน routing จาก ERP generic → จอใหม่; **ลบ config 2 ตัวใน `erp-transaction.ts:387-406`** และปิด backend ทางเดิม (CB-15) ใน change เดียวกัน |

**เอกสารเปลี่ยนสถานะ 8 ตัว**
- เมนู id: `cheque-deposit`, `cheque-received-clear`, `cheque-received-return`, `cheque-redeposit`, `cheque-received-cancel`, `cheque-discount`, `cheque-issued-clear`, `cheque-issued-cancel`
- route: `/banking/cheques/*` เดิม
- จอเดียว `components/cheque-status-doc-screen.tsx` ตั้งค่าด้วย transflag
- หัวเอกสารตาม §1.2:
  - 423 มีคอมโบ "เช็คที่ต้องการยกเลิก"
  - 426 มีช่องติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว"
- กริดเลือกเช็คจาก `/selectable` และท้ายเอกสารแสดง 3 ยอด (จำนวนเงินรวม / รวมค่าใช้จ่าย / ยอดที่หักค่าใช้จ่ายแล้ว)
  - **ยกเว้นนำฝาก 419:** กริดไม่มีคอลัมน์ค่าใช้จ่าย/ยอดหัก มีคอลัมน์หมายเหตุอ่านอย่างเดียว (= คำอธิบายของทะเบียนเช็ค) และท้ายเอกสารแสดงแค่ "จำนวนเงินรวม"
  - 423 แสดงข้อความเตือนสีแดงใต้หัวเอกสารตามจอจริง (`cheque_cancel_notice`); 426 ไม่มีข้อความนี้ (ตรวจ screenshot `บันทึกยกเลิกเช็คจ่าย.jpg` แล้ว)
  - 423 ชนิด 0 ปิดช่องสมุด (สีเทา) ตามจอจริง
- ลบ config 8 ตัวใน `erp-transaction.ts:647-725` และปิด backend ทางเดิม (CB-15) ใน change เดียวกัน

**จอที่มาแทนจอเงา (ทำต่อจาก F2 ของ `facade-screens.md`)**

| เมนู id | route | จอ | หมายเหตุ |
|---|---|---|---|
| `cheque-issued-prepare` | `/banking/cheque-issued-prepare` | `components/cheque-prepare-screen.tsx` | ตามภาพ CQA; ปุ่ม "รายการอัตโนมัติ" = `generate`; กริดแก้เลขได้ก่อนบันทึก; แถวที่ใช้แล้วล็อกและแสดงเลขที่เอกสาร |
| `cheque-movement` | `/banking/cheque-movement` | `components/cheque-movement-screen.tsx` (ไฟล์ชื่อเดิมถูกลบใน F2 สร้างใหม่) | อ่านอย่างเดียว ตามภาพ §1.6 |
| `bank-service` | `/banking/bank-service` | `components/bank-service-notes-screen.tsx` | รหัสธนาคาร + 3 แท็บ กริด "รายละเอียด" |

**สมุดเงินฝากและรหัส**

| เมนู id | route | จอ | หมายเหตุ |
|---|---|---|---|
| `book-bank` | `/bookbankscreen` | `components/book-bank-screen.tsx` | ลำดับตาม Champ 5169: เลขที่บัญชีธนาคาร + ปุ่ม "เปลี่ยนเลขที่บัญชี" \| ประเภทบัญชี / ชื่อบัญชี / ธนาคาร \| สาขาธนาคาร / รหัสผังบัญชี \| ยอดคงเหลือ (อ่านอย่างเดียว = ยอด GL) + ตาราง "ชื่อผู้ติดต่อ" อ่านอย่างเดียว (จาก CB-16); **ลบ config bookbank ใน `system-setting-screens.ts:893-922` + การ map ใน `api/system-settings/[[...settingPath]]/route.ts:322-355`** และยกเลิก retired ใน `menu-screen-status` (หลัง workflow อื่น commit) |
| `bank-code-*` (ใหม่ 5 id) | `/settings/bank-codes/:kind` | `components/bank-code-screen.tsx` | ใต้ submenu `system-codes` (`menu-bars.ts:308-311`) ชื่อเมนูตาม Champ: กำหนดรหัสธนาคาร, กำหนดรหัสสาขาของธนาคาร, กำหนดรหัสบริการธนาคาร, กำหนดรหัสประเภทบัญชีธนาคาร, กำหนดรหัสสถานะเช็คเพิ่มเติม (ตามแนว submenu ที่ BC ตัดสินแล้ว `menu-bars.ts:300`; D11 ตัดออก) |
| `bank-contact` (5680, มีอยู่แล้วเป็น "รอพัฒนา") | `/banking/contacts` (เดิม) | `components/bank-contact-screen.tsx` (CB-16) | ตามภาพ `บันทึกรายละเอียดผู้ติดต่อธนาคาร.jpg`: ธนาคาร / สาขาธนาคาร + กริด A (ติดต่อ) \| รหัสผู้ติดต่อ \| ชื่อ/รายละเอียด \| ตำแหน่ง \| ที่อยู่ \| โทรศัพท์ |

**เครื่องมือและรายงาน**

| เมนู id | route | จอ | หมายเหตุ |
|---|---|---|---|
| `cheque-recalculate` | `/tools/cheque-balance…` (เดิม) | `components/cheque-recalculate-screen.tsx` | ตรวจ → แสดงผลต่าง → ปรับ (ต้องใส่เหตุผล) |
| `cheque-received-report` / `cheque-issued-report` | `/report/chequereceived`, `/report/chequeissued` | จอรายงานกลาง | report id 5380–5384, 5386–5391 + รายวัน 5409–5413, 5416 ใน registry (รายชื่อ `facade-screens.md` §5) |
| พิมพ์เช็ค | ปุ่มในจอเช็คจ่าย | dialog `cheque-print-dialog.tsx` | ช่วงเลขที่เอกสาร/เลขที่เช็ค + ปุ่ม "ตั้งตำแหน่งพิมพ์" → `cheque-print-layout-dialog.tsx` + "พิมพ์ตรวจตำแหน่ง" |

**กติกาทุกจอ**
- รับ prop `language` + `useBackendLanguage`
- ไม่มีข้อความ hard-code
- เงินรับเป็นสตริงทศนิยม (`numeric-input`) และไม่มีการคำนวณบัญชีในเบราว์เซอร์
  - ยอดในกริดมาจาก backend หรือคำนวณเพื่อแสดงผลด้วย `decimal.js` แล้ว backend คำนวณซ้ำ
- ปุ่มยกเลิกเปิด dialog ไทยที่อธิบายผล (สถานะเช็คจะกลับเป็นอะไร และ GL จะกลับรายการ) + ช่องเหตุผลบังคับ
- dirty guard
- คีย์ลัด New/Save/Find ตามแถบเครื่องมือ Champ เท่าที่จอ ERP เดิมมี

### 4.9 Language keys (เพิ่มใน `backend/assets/language/languages.tsv` ครบ 13 คอลัมน์; ภาษาอื่นใส่อังกฤษเป็น placeholder)

**key ที่มีอยู่แล้ว ใช้ซ้ำ:**
- `cheque_number`, `cheque_amount`, `bank_code`, `bank_name`, `bank_book_code`, `bank_charge`
- เมนู `cheque_*`, `book_bank`, `bank_service`, `cheque_movement`, `cheque_recalculate`

**key ใหม่:**

| key | th | en |
|---|---|---|
| `cheque_field_doc_no` | เลขที่เอกสาร | Document No. |
| `cheque_field_receive_date` | วันที่รับเช็ค | Received Date |
| `cheque_field_pay_date` | วันที่จ่ายเช็ค | Payment Date |
| `cheque_field_doc_date` | วันที่เอกสาร | Document Date |
| `cheque_field_due_date` | วันที่ครบกำหนด | Due Date |
| `cheque_field_ar_code` | รหัสลูกค้า | Customer Code |
| `cheque_field_ap_code` | รหัสเจ้าหนี้ | Supplier Code |
| `cheque_field_status_in` | สถานะเช็ครับ | Received Cheque Status |
| `cheque_field_status_out` | สถานะเช็คจ่าย | Issued Cheque Status |
| `cheque_field_cheque_no_in` | เลขที่เช็ครับ | Received Cheque No. |
| `cheque_field_cheque_no_out` | เลขที่เช็คจ่าย | Issued Cheque No. |
| `cheque_field_ext_status` | สถานะเพิ่มเติม | Extended Status |
| `cheque_field_bank_branch` | สาขาของธนาคาร | Bank Branch |
| `cheque_field_book_in` | สมุดเงินฝากธนาคาร | Bank Book |
| `cheque_field_book_out` | สมุดเงินฝาก | Bank Book |
| `cheque_field_book_code` | รหัสสมุดเงินฝาก | Bank Book Code |
| `cheque_field_currency` | รหัสสกุลเงิน | Currency Code |
| `cheque_field_rate` | อัตราแลกเปลี่ยน | Exchange Rate |
| `cheque_field_amount_currency` | จำนวนเงิน (ตามสกุลเงิน) | Amount (Currency) |
| `cheque_field_amount_home` | จำนวนเงิน (เงินบาท) | Amount (THB) |
| `cheque_field_description` | คำอธิบาย | Description |
| `cheque_field_sale_code` | รหัสพนักงานขาย | Salesperson Code |
| `cheque_field_department` | รหัสแผนก | Department Code |
| `cheque_field_update_info` | การปรับปรุงเอกสาร | Document Update Info |
| `cheque_btn_prepare` | เช็คเตรียมจ่าย | Prepared Cheques |
| `cheque_ref_col_date` | วัน | Date |
| `cheque_ref_col_amount` | จำนวนเงิน | Amount |
| `cheque_col_cheque_amount` | จำนวนเงินบนเช็ค | Cheque Amount |
| `cheque_cancel_notice` | \*\*\*เช็คที่ยกเลิกไปแล้ว ไม่สามารถที่นำไปทำรายการได้อีก (สถานะเป็นเช็คยกเลิก)\*\*\* | \*\*\*Cancelled cheques cannot be used in any further transaction (status: Cancelled)\*\*\* |
| `cheque_field_note` | หมายเหตุ | Remark |
| `cheque_field_cancel_type` | เช็คที่ต้องการยกเลิก | Cheques to Cancel |
| `cheque_field_cancel_passed` | ยกเลิกเช็คที่ผ่านแล้ว | Cancel Cleared Cheques |
| `cheque_col_bank` | ธนาคาร | Bank |
| `cheque_col_fee` | ค่าใช้จ่าย | Expense |
| `cheque_col_net` | ยอดที่หักค่าใช้จ่ายแล้ว | Net Amount |
| `cheque_sum_total` | จำนวนเงินรวม | Total Amount |
| `cheque_sum_fee` | รวมค่าใช้จ่าย | Total Expense |
| `cheque_status_in_0` … `_6` | เช็คในมือ / เช็คฝากธนาคาร / เช็คผ่าน / เช็คคืน / เช็คยกเลิก / ขายลดเช็ค / เช็คคืนฝากธนาคาร | On Hand / Deposited / Cleared / Returned / Cancelled / Discounted / Redeposited |
| `cheque_status_out_0` … `_3` | เช็คในมือ / เช็คผ่าน / เช็คยกเลิก / เตรียมจ่าย | On Hand / Cleared / Cancelled / Prepared |
| `cheque_cancel_type_0` … `_3` | ยกเลิกเช็คในมือ / ยกเลิกเช็คที่ผ่านแล้ว / ยกเลิกเช็คฝาก / ยกเลิกเช็คคืนฝากธนาคาร | Cancel On-hand / Cleared / Deposited / Redeposited |
| `cheque_move_type_in` / `_out` | เช็ครับ / เช็คจ่าย | Received / Issued |
| `cheque_move_value` | มูลค่าเช็ค | Cheque Value |
| `cheque_move_bank_branch` | ธนาคาร/สาขา | Bank/Branch |
| `cheque_move_partner` | ชื่อลูกค้า / เจ้าหนี้ | Customer / Supplier |
| `cheque_move_account_no` | เลขที่บัญชีธนาคาร | Bank Account No. |
| `cheque_move_cheque_status` | สถานะเช็ค | Cheque Status |
| `cheque_move_grid_trade` | รายวันซื้อ/ขาย | Sales/Purchase Journals |
| `cheque_move_grid_cheque` | รายวันเช็ค | Cheque Journals |
| `cheque_col_item` | รายการ | Item |
| `cheque_prep_start_no` | เลขที่เช็คจ่าย | Cheque No. |
| `cheque_prep_run_count` | จำนวนที่ Run Auto | Auto Run Count |
| `cheque_prep_generate` | รายการอัตโนมัติ | Auto Generate |
| `cheque_prep_used_by` | ใช้แล้วในเอกสาร | Used in Document |
| `bank_service_tab_interest` | อัตราดอกเบี้ย | Interest Rates |
| `bank_service_tab_exchange` | อัตราแลกเปลี่ยน | Exchange Rates |
| `bank_service_tab_service` | บริการธนาคาร | Bank Services |
| `bank_service_detail` | รายละเอียด | Details |
| `bank_code_kind_bank` … | กำหนดรหัสธนาคาร / กำหนดรหัสสาขาของธนาคาร / กำหนดรหัสบริการธนาคาร / กำหนดรหัสประเภทบัญชีธนาคาร / กำหนดรหัสสถานะเช็คเพิ่มเติม | Bank / Bank Branch / Bank Service / Bank Account Type / Extended Cheque Status Codes |
| `bank_code_user_group` | กลุ่มผู้ใช้ | User Group |
| `book_bank_account_type` | ประเภทบัญชี | Account Type |
| `book_bank_account_name` | ชื่อบัญชี | Account Name |
| `book_bank_account_no` | เลขที่บัญชีธนาคาร | Bank Account No. |
| `book_bank_gl_account` | รหัสผังบัญชี | Account Code |
| `book_bank_change_account_no` | เปลี่ยนเลขที่บัญชี | Change Account No. |
| `book_bank_confirm_change_no` | ยืนยันเปลี่ยนเลขที่บัญชีจาก {old} เป็น {new} หรือไม่ เอกสารเดิมทั้งหมดจะอ้างเลขใหม่ | Change account number from {old} to {new}? All existing documents will show the new number. |
| `book_bank_contacts` | ชื่อผู้ติดต่อ | Contacts |
| `bank_contact_code` | รหัสผู้ติดต่อ | Contact Code |
| `bank_contact_name` | ชื่อ/รายละเอียด | Name/Details |
| `bank_contact_position` | ตำแหน่ง | Position |
| `bank_contact_address` | ที่อยู่ | Address |
| `bank_contact_phone` | โทรศัพท์ | Telephone |
| `bank_contact_active` | ติดต่อ | Active Contact |
| `book_bank_contacts_empty` | ยังไม่มีผู้ติดต่อ | No contacts yet |
| `cheque_print_title` | พิมพ์เช็ค | Print Cheque |
| `cheque_print_layout` | ตั้งตำแหน่งพิมพ์ | Print Layout |
| `cheque_print_calibrate` | พิมพ์ตรวจตำแหน่ง | Print Alignment Test |
| `cheque_print_count` | พิมพ์แล้ว (ครั้ง) | Printed (times) |
| `cheque_print_cross` | ขีดคร่อม | Crossed Cheque |
| `cheque_print_strike_bearer` | ขีดฆ่าคำว่า "หรือผู้ถือ" | Strike "or Bearer" |
| `cheque_print_era` | ปีที่พิมพ์ (พ.ศ./ค.ศ.) | Year Era (BE/CE) |
| `cheque_recalc_mismatch` | สถานะไม่ตรงกับประวัติ | Status differs from history |
| `cheque_confirm_cancel_doc` | ยกเลิกเอกสารนี้ เช็คจะกลับเป็นสถานะ {status} และกลับรายการบัญชี | Cancel this document? Cheques revert to {status} and the journal is reversed. |

**error keys** (th ต้องบอกวิธีแก้ เช่น `cheque_err_newer_event` = "ยกเลิกไม่ได้ เพราะเช็คเลขที่ {no} มีเอกสาร {doc} ที่ใหม่กว่า กรุณายกเลิกเอกสารนั้นก่อน")
- ทั้งหมดที่ระบุใน §4.7
- ใช้ prefix: `cheque_err_*`, `cheque_prep_err_*`, `bank_code_err_*`, `bank_service_err_*`, `cheque_print_err_*`, `bank_contact_err_*`, `bank_book_err_*`
- เพิ่ม 2026-10-11: `cheque_err_referenced_delete` (ลบไม่ได้ เช็คถูกใช้ในเอกสาร {doc} แล้ว ให้ยกเลิกเอกสารนั้นก่อน), `cheque_err_from_source_doc` (เช็คนี้สร้างจากเอกสาร {doc} ให้ลบหรือยกเลิกที่เอกสารนั้น), `cheque_err_currency_not_supported` (ตอนนี้รับเฉพาะเงินบาท อัตราแลกเปลี่ยน 1), `bank_book_err_account_number_duplicate`

**ป้ายที่ยังไม่ verify:** ป้ายหัวคอลัมน์จาก resource key ของ Champ (ARNAME/BANK/STATUS/SUM_FEE ฯลฯ) ยังไม่ได้ verify ภาษาไทย (`05-cheque.md` §10) → ใช้ป้ายจากภาพ bcspec ที่ verify แล้ว และตรวจซ้ำกับภาพหน้าจอตอนทำ

---

## 5. การทดสอบ

**Go sqlmock** (`backend/internal/cheque/*_test.go`)
- validate ทะเบียนทั้งรับและจ่าย
  - เลขซ้ำ (unique index violation → `cheque_err_duplicate_number`)
  - วันที่จ่าย > ถึงกำหนด
  - สถานะ ≠ 0 แต่ไม่มีสมุด
  - amount/rate ที่ไม่ใช่ทศนิยม 2 ตำแหน่ง
  - home_amount = round half up (เช่น `1000.005 × 1` → ไม่ให้ใส่ 3 ตำแหน่ง; `333.33 × 3.0000000000`)
  - 0.1 + 0.2 = 0.3
- ตัวกรองเลือกเช็คครบ 8 transflag × สถานะทั้งหมด (table-driven)
- `cheque_err_status_not_selectable` เมื่อมี tx อื่นเปลี่ยนสถานะไปก่อน (ล็อก FOR UPDATE)
- ยอดหัวเอกสารคำนวณใหม่ และ `cheque_err_totals_mismatch`
- คำนวณสถานะใหม่ตอนยกเลิกเอกสาร (event ล่าสุดที่ยังไม่ void), `cheque_err_newer_event`, ล้าง book_code เมื่อกลับเป็น 0
- ตัวกรอง 422/423 ตาม Champ (เพิ่ม 2026-10-11): 422 ไม่มีสมุด → `cheque_err_book_required`; วันที่รับ > วันที่เอกสาร → `cheque_err_received_after_doc` (ทั้ง 422 และ 423 ทุกชนิด); 423 ชนิด 2 ไม่มีสมุด → `cheque_err_book_required`; 423 ชนิด 0 ไม่ต้องมีสมุด
- 419: ส่ง fee_amount ≠ 0 → `cheque_err_fee_invalid` (และ CHECK ในตารางกันซ้ำ)
- ทะเบียน: ลบเช็คที่ถูกอ้าง → `cheque_err_referenced_delete`; ลบเช็คที่ `source_kind='erp'` → `cheque_err_from_source_doc`; currency ≠ THB หรือ rate ≠ 1 → `cheque_err_currency_not_supported`
- `references[]` ของ GET ทะเบียน: เรียงตามวันที่ ไม่รวม event/link ที่ void
- mapping:
  - `accountfrom: bank_book` + มีสมุด → ใช้บัญชีสมุดแม้ rule มีบัญชีสำรอง
  - 423 ชนิด 0 ไม่มีสมุด + rule มีบัญชีสำรอง → ใช้บัญชีสำรอง; ไม่มีบัญชีสำรอง → `cheque_err_book_no_gl_account`
  - ไม่มีสมุดในชนิดอื่น → `cheque_err_book_no_gl_account` และไม่มีแถวใดถูกเขียน
  - แหล่งยอด cheque กับ source doc ของ ERP → `mapping_rule_source_invalid`
- route เดิม: `POST /transaction/chequedeposit` (และ doctype เช็คทั้ง 10) → ถูกปฏิเสธ (CB-15)
- เปลี่ยนเลขที่บัญชี: เลขซ้ำ → `bank_book_err_account_number_duplicate`; สำเร็จ → `bank_account_code` เดิม, เอกสารเช็คยังอ้างสมุดเดิม
  - `mappingBalances` ของชุด net+fee=total
- ใบเตรียมเลข:
  - `000123` + 5 → `000123`..`000127`
  - `999` + 2 → `cheque_prep_err_overflow`
  - `A12` → `cheque_prep_err_start_not_numeric`
  - ลบใบที่มีเลขใช้แล้ว → `cheque_prep_err_number_used`
- layout:
  - ช่องอยู่ในแถบ 15.875 มม. → error
  - ขนาด 9.5 pt → error
  - era ว่าง → error
- คำอ่านเงินบาท: เทียบ `whtcert.BahtText` กับตัวเลขค่าเดียวกัน (มาตรา 12) + กรณี `0.50`, `21.00`, `1000001.00`

**Integration** (`BC_GL_TEST_POSTGRES_DSN`, แบบ `fixedasset/gl_poster_integration_test.go`)
- สร้างเช็ครับ → นำฝาก (ไม่มี GL)
- → รับผ่าน + ค่าธรรมเนียม (GL ลงจริง เดบิต = เครดิต, บัญชีสมุดมาจาก bank book)
- → รับคืน → นำเข้าใหม่ → ยกเลิก (cancel_type 3)
- ยกเลิกย้อนทีละใบจนกลับเป็น 0
- ตรวจ `gl_journals` ทุกขั้น และว่าใบกลับรายการอยู่ครบ
- ไม่มี mapping → `no_mapping`
- งวดปิด → error และไม่มีแถวใดถูกเขียน (rollback)
- ใบเตรียมเลข → เช็คจ่ายใช้เลข → `used_cheque_id` ถูกตั้ง → ยกเลิกเช็ค → คืนค่า

**vitest**
- `cheque-register-screen.test.ts`: ลำดับฟิลด์และ key ป้ายแยกรับ/จ่าย, ซ่อนพนักงานขายในเช็คจ่าย, เงินเป็นสตริง, ตารางอ้างอิง 5 คอลัมน์, ปุ่มเช็คเตรียมจ่ายมีเฉพาะเช็คจ่าย
- `cheque-status-doc-screen.test.ts`: config 8 ตัว, คอมโบ cancel_type, ช่องติ๊ก 426, 419 ไม่มีคอลัมน์ค่าใช้จ่าย/ยอดหักและท้ายเอกสารมียอดเดียว, ข้อความเตือนเฉพาะ 423, 423 ชนิด 0 ปิดช่องสมุด
- `book-bank-screen.test.ts`, `bank-contact-screen.test.ts`: ปุ่มเปลี่ยนเลขที่บัญชี, ตารางผู้ติดต่ออ่านอย่างเดียว
- `cheque-prepare-screen.test.ts`, `cheque-movement-screen.test.ts`, `bank-service-notes-screen.test.ts`, `bank-code-screen.test.ts`
- `api/cheque/[...chequePath]/route.test.ts`: allowlist, ส่ง Accept-Language, relay `success:false`
- `menu-screen-status.test.ts`: route ใหม่ไม่ retired แล้ว
- `erp-transaction.test`: ไม่มี config เช็คเหลือ
- guard ภาษา: ทุก key มีใน tsv

**UAT CRUD + psql** (Demo → `rungrueng` → บริษัท `01` → สาขา `00000`)
- **วิธีรัน:** `docker exec -it postgres psql -U <db_user> -d <holding_db>` (ไม่ใส่รหัสผ่านในเอกสาร) รันหลังทุกขั้น
- ข้อมูลสมจริง:
  - ลูกหนี้ "บริษัท ศรีสยามวัสดุก่อสร้าง จำกัด"
  - เช็คเลขที่ `10004521`
  - ธนาคาร `BBL` "ธนาคารกรุงเทพ"
  - สาขา `0123` "สาขาสีลม"
  - ยอด `125000.00`

| ขั้น | ทำบนจอ | ตรวจ psql |
|---|---|---|
| C | เพิ่มรหัสธนาคาร/สาขา | `SELECT kind,code,name,version FROM bank_codes WHERE company='01' AND code IN ('BBL','0123');` → 2 แถว version 1 |
| C | บันทึกเช็ครับ | `SELECT cheque_type,cheque_no,status,amount,home_amount FROM cheque_registers WHERE company='01' AND cheque_no='10004521';` → 0, 0, 125000.00 |
| U | แก้คำอธิบาย | แถวเดิม (id เดิม) version +1, `cheque_events` มี `update` 1 แถว |
| C | นำฝาก 419 | status=1, book_code ตรง, `cheque_docs.gl_status='not_required'` |
| C | รับผ่าน 420 ค่าธรรมเนียม 10.00 | status=2; `SELECT total_amount,fee_amount,net_amount,gl_status FROM cheque_docs WHERE company='01' AND transflag=420 AND doc_no='<เลข>';` → 125000.00/10.00/124990.00/posted; journal ใน GL เดบิต=เครดิต |
| edge | กดบันทึกซ้ำสองครั้ง (requestid เดิม) | มีเอกสาร 1 ฉบับ |
| edge | ยกเลิก 419 ขณะที่ 420 ยัง active | error `cheque_err_newer_event`; ไม่มีแถวใดเปลี่ยน |
| D(ยกเลิก) | ยกเลิก 420 พร้อมเหตุผล | `cheque_docs.status='cancelled'`, event voided, status กลับเป็น 1, ใบกลับรายการ GL มีอยู่ |
| edge | บันทึกเช็คเลขซ้ำ ธนาคาร/สาขาเดียวกัน | error; จำนวนแถวไม่เพิ่ม |
| edge | เลขเช็คมีอักขระพิเศษ/ช่องว่าง | trim แล้วตรวจ; ช่องว่างล้วน → required |
| D | ลบเช็คที่ไม่มี event (ยกเลิกทะเบียน) | ไม่มีแถว active เหลือ; เช็คอื่นยังอยู่ (`count(*)` ก่อน/หลัง) |
| B3 | ใบเตรียมเลข 000123+5 → บันทึกเช็คจ่ายเลข 000124 | `SELECT cheque_no,used_cheque_id IS NOT NULL FROM cheque_prepare_numbers WHERE company='01' AND doc_no='<เลข>' ORDER BY line_no;` |
| B4 | บันทึก 3 แท็บ แล้วแก้แท็บ 1 | `bank_service_notes` ตรงกับจอ; `bank_service_note_logs` มีบรรทัดก่อนแก้ |
| B1 | ตั้ง layout แล้วพิมพ์ตรวจตำแหน่ง/พิมพ์จริง | PDF ขนาด 178×89 มม.; `print_count`=1; layout ที่ y < 15.875 มม. ถูกปฏิเสธ |
| ล้าง | ลบข้อมูลทดสอบตาม id/doc_no เป้าหมายเท่านั้น | ห้าม regex กว้าง |

---

## 6. ความเสี่ยง

| # | ความเสี่ยง | ผล | ทางลด |
|---|---|---|---|
| R1 | ขอบเขตขัดกับ `mydocs/specs/gl/ar-ap-bank-design-notes.md:49` | สร้างของที่ลุงจืดไม่ต้องการ | D1 ตัดสินแล้ว 2026-10-11: ทำครบแบบ Champ เป็นขั้น |
| R2 | เอกสาร ERP เป็น `float64` (`models.go:25-62`) | ถ้าเอาเช็คไปใช้ร่วมจะได้เงินเพี้ยน | โมดูลแยก + decimal; CB-14 ต้องแปลงเงินที่ขอบด้วยสตริง |
| R3 | workflow อื่นกำลังแก้ transaction/GL guards/erp-* | conflict หรือทับงาน | CB-2 (`menu-screen-status`), CB-4/CB-6 (`erp-transaction`), CB-5 (generalledger), CB-13 (`erp-reports`), CB-14 (transaction) เริ่มหลัง workflow นั้น commit แล้ว; CB-1/3/8/9/10/11/12 เป็นไฟล์ใหม่ เริ่มได้ก่อน (CB-1 แตะ `menu-data.ts`/`menu-bars.ts` ซึ่งไม่อยู่ในรายการที่ห้าม) |
| R4 | prod ยังไม่มีรูปแบบลงบัญชีเลย | เอกสารเช็คทุกใบได้ `no_mapping` | จอแจ้งชัดและมีลิงก์ไปตั้ง mapping; ไม่สร้างบัญชีเริ่มต้นให้ (ห้ามเดา) |
| R5 | ตำแหน่งผู้รับเงิน/ตัวอักษรของแต่ละธนาคารไม่ทราบ | พิมพ์ผิดตำแหน่ง เช็คเสีย | ไม่มีค่าเริ่มต้น + พิมพ์ตรวจตำแหน่งก่อน + ตรวจแถบ MICR |
| R6 | `pdftext` ยังไม่มีเส้น | ขีดคร่อมไม่ได้ | CB-11 เพิ่ม `Line` + test byte output |
| R7 | เลขเช็คซ้ำข้ามธนาคาร (เช็ครับ) | สืบค้น B2 ด้วยเลขอย่างเดียวได้หลายใบ | คืน `cheque_err_ambiguous_number` พร้อมตัวเลือก |
| R8 | กติกาของ Champ บางข้อยังไม่ verify (การรันเลขเตรียมจ่าย, ผลต่อสมุดของ 5402 type 2/5405 ไม่ติ๊ก, ป้ายหัวคอลัมน์ของจอที่ยังไม่ได้เทียบ screenshot, แผนกในบรรทัด GL) — การปัดเศษ home_amount ตรวจแล้ว 2026-10-11 (ปัดครึ่งขึ้น) | ผลต่างจาก Champ | ระบุใน §10 และให้ลุงจืดยืนยันตอน UAT |
| R9 | ใบเสร็จ/ใบจ่ายของ ERP ยังไม่มีบรรทัดชำระด้วยเช็คที่ persist | B2 กริด "รายวันซื้อ/ขาย" ว่าง; เช็คจาก ERP ไม่เข้าทะเบียนเอง | CB-14 รอ workflow ERP; ระหว่างนั้นคีย์ทะเบียนเองได้ |
| R10 | ~~การเบี่ยงจาก Champ ที่เหลือจริง~~ ไม่เบี่ยงแล้ว (D8 = Champ 2026-10-11); (ห้ามแก้ฟิลด์หลักของเช็คที่ถูกอ้าง ซึ่ง Champ แค่ถามยืนยัน; ห้ามแก้สถานะในจอทะเบียนหลังสร้าง) — ห้ามข้ามลำดับและห้ามลบเช็คที่ถูกอ้าง**เป็นของ Champ อยู่แล้ว** | ผู้ใช้เดิมต้องปรับตัวเล็กน้อย | D8 ให้ลุงจืดเลือก; ข้อความ error บอกวิธีแก้ |
| R11 | `gl_subledger_bank_accounts` เป็น payload jsonb | ตรวจ FK ด้วย SQL ไม่ได้ | ตรวจใน Go ใน tx เดียวกัน + test |
| R12 | backend `/transaction/:module` ยังรับ doctype เช็ค 10 ตัว (`http.go:42-58`, `doctype.go:80,114-122`) | เขียนเช็คเป็นเอกสาร ERP แบบ float ซ้อนกับโมดูลใหม่ เลขเอกสาร/ใบสำคัญชนกัน | CB-15 ปิดใน change เดียวกับ CB-4/CB-6 + test |
| R13 | ยังไม่ตัดสินเรื่องหลายสกุลเงิน (D15) | ถ้าเปิดรับสกุลอื่นก่อนตัดสิน จะขัด `mydocs` ที่จำกัด THB | backend รับเฉพาะ THB / อัตรา 1 จนกว่าจะตัดสิน |

---

## 7. Open decisions (ลุงจืดตัดสิน)

| # | คำถาม | แนะนำ | ทางเลือกอื่น |
|---|---|---|---|
| D1 **[ตัดสินแล้ว 2026-10-11: ทำตาม Champ]** (a) | ทำวงจรเช็คครบแบบ Champ หรือไม่ ทั้งที่ `mydocs/specs/gl/ar-ap-bank-design-notes.md:49` เขียนว่า "workflow สั่งจ่าย/คืนเงิน/เช็คเป็นหน้าที่ระบบต้นทาง" | (a) ทำครบแบบเป็นขั้น: CB-1→CB-6 ก่อน (รหัส + สมุดเงินฝาก + ทะเบียน + เอกสารสถานะ + GL) เพราะเมนูเช็คอยู่ใน Champ และลุงจืดสั่ง "รายการบัญชีก่อน" | (b) GL-only: ทำแค่รายละเอียดเช็คใต้รายวัน (CB-7 ใน §8) + รหัส/สมุดเงินฝาก (CB-1/CB-2) และให้เมนูเช็คอื่นเป็นรอพัฒนา |
| D4 | วิธีพิมพ์เช็คจ่าย (รวม D6, D12 เดิม) | (a) ตั้งตำแหน่งเป็นมม. ต่อสมุด + พิมพ์ตรวจตำแหน่ง + ตรวจแถบ MICR/10 pt; ขีดคร่อม/ข้อความคร่อม/ขีดฆ่า "หรือผู้ถือ" เป็นตัวเลือกปิดไว้ (ข้อความผู้ใช้พิมพ์เอง เช่น "ห้ามเปลี่ยนมือ" ป.พ.พ. 995); ปี พ.ศ./ค.ศ. บังคับเลือกตอนตั้งตำแหน่ง ไม่มีค่าเริ่มต้น | (b) form designer แบบ Champ (งานใหญ่มาก); (c) preset ต่อธนาคาร (ไม่มีแหล่งทางการของพิกัด = ห้ามเดา) |
| D8 **[ตัดสินแล้ว 2026-10-11: ทำตาม Champ]** | เช็คที่ถูกอ้างแล้ว (หดเหลือส่วนที่เบี่ยงจริง 2026-10-11) | ห้ามแก้ฟิลด์หลัก (เลขที่/ธนาคาร/สาขา/สมุด/ยอด/อัตรา/คู่ค้า) และห้ามแก้สถานะในจอทะเบียนหลังสร้าง | ทำตาม Champ: ถามยืนยันแล้วยอมให้แก้ และแก้สถานะในจอทะเบียนได้ |
| D13 | การรันเลขเตรียมจ่าย | ตัวเลขล้วน รันทีละ 1 โดยคงความกว้าง (มีศูนย์นำหน้า); ไม่ใช่ตัวเลข = error; ผู้ใช้แก้เลขในกริดได้ก่อนบันทึก | รองรับคำนำหน้าตัวอักษร (เช่น `A000123`) — ไม่มีหลักฐานจาก Champ |
| D15 (D-FX ใหม่) | เช็คสกุลเงินต่างประเทศ: Champ มีช่องสกุลเงิน/อัตรา แต่ `mydocs/datamodels/gl/bank.sql:10` และ `ar-ap-bank-design-notes.md:31,49` จำกัด THB | แสดงช่องตามจอ Champ แต่รับเฉพาะเงินบาท อัตรา 1 (Champ เองบังคับอัตรา 1 เมื่อบริษัทไม่เปิดหลายสกุลเงิน `cqfrmcheqin.cpp:1135`) | เปิดหลายสกุลเงินเต็มแบบ Champ ตอนนี้ |

**ตัดออก 2026-10-11 เพราะ Champ หรือข้อกำหนดของลุงจืดตอบแล้ว:**
- สถานะ "อนุมัติ" ของเอกสารเช็ค (ข้อเสนอจากการตรวจทาน) → **ไม่ทำขั้นอนุมัติ**: ฟอร์มเช็คมีตัวกันแก้/ลบเมื่อ isconfirm = 1 (`cqfrmcheqindeposit.cpp:145-150`, `:200-205`; `cqfrmcheqinpass.cpp:349-354`, `:455-460`; DDL `:7993`, `:8128`, `:8199`) แต่**ไม่มีโค้ดใดใน `champ/champ` และ `champ-lib` ตั้ง isconfirm ของตาราง BCChq\*** (ค้น `SET ... IsConfirm` ทั้งสองโฟลเดอร์ เจอเฉพาะใบสั่งขาย/มัดจำ/ตรวจนับ/ใบรับวางบิล) → ใน Champ ตัวกันนี้ไม่เคยทำงาน ผู้ใช้ไม่มีขั้นอนุมัติเช็ค; BC จึงไม่เพิ่มขั้นอนุมัติ (ถ้าภายหลังมีระบบอนุมัติกลาง ค่อยเพิ่มตัวกันตามลำดับ §1.2 ข้อ 4)
- D2 สมุดเงินฝาก → ขยายข้อมูลสมุดเดิม (คีย์ = รหัสสมุด + เลขบัญชีแยกฟิลด์) ตรงกับ `mydocs/datamodels/gl/bank.sql:3-13` (`bank_account_code` PK + `account_number`) และผู้ใช้เห็นเหมือน Champ (มีปุ่มเปลี่ยนเลขที่บัญชี)
- D3 สาขาธนาคาร → รหัสเดียวทั้งบริษัทแบบ Champ (DDL `BCBankBranch` PK = Code)
- D5 ครอบคำอ่านจำนวนเงินด้วย "( … )" → เปิดตาม XMoneyWord ของ Champ
- D6, D12 → รวมเข้า D4
- D7 รายละเอียดธนาคาร → บรรทัดข้อความแบบ Champ จริง (คอลัมน์ตัวเลขถูก comment ทิ้ง `CashFrmBankServicePage1.cpp:79-84`)
- D9 เมนูคำนวณยอดเช็คใหม่ → คงไว้ (มีใน Champ)
- D10 เช็คใต้รายวัน GL → แยกจากทะเบียนแบบ Champ (`BCGLChq`)
- D11 เมนูรหัส → ใช้แนวที่ BC ตัดสินแล้วสำหรับ "กำหนดรหัสต่างๆ" (`frontend/src/lib/menu-bars.ts:300` submenu ของจอรหัส)
- D14 การปัดเงินบาท → ปัดครึ่งขึ้นตามทศนิยมเงิน แบบ Champ `XRound` (`cqfrmcheqindeposit.cpp:475-476`, `QOledb.cpp:1052-1102`)

---

## 8. Work items

| id | งาน | ไฟล์หลัก | ขนาด | ต้องรอ |
|---|---|---|---|---|
| CB-1 | ทะเบียนรหัส 5 ชนิด + API + จอรหัส + เมนู | `backend/internal/cheque/{schema.go,codes.go,codes_test.go}`, `backend/internal/cheque/httpapi/http.go`, `backend/main.go`, `frontend/src/app/api/cheque/[...chequePath]/route.ts`(+test), `frontend/src/lib/cheque-api.ts`, `frontend/src/components/bank-code-screen.tsx`(+test), `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-bars.ts`, `backend/assets/language/languages.tsv` | M | — (ทั้งทาง a และ b ของ D1 ต้องมีรหัสธนาคาร) |
| CB-2 | ขยายสมุดเงินฝาก + CRUD (GL resource `bank-accounts`) + จอ `/bookbankscreen` + ปุ่ม "เปลี่ยนเลขที่บัญชี" + ตารางผู้ติดต่ออ่านอย่างเดียว + รับเฉพาะ THB + ลบ config bookbank เดิม | `backend/internal/generalledger/{subledger_types.go,subledger_records.go}`, `backend/internal/generalledger/httpapi/{http.go,subledger.go}`, `frontend/src/components/book-bank-screen.tsx`, `frontend/src/lib/system-setting-screens.ts`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`, `frontend/src/lib/menu-screen-status.ts`, `languages.tsv` | M | CB-1, งานของ workflow อื่น (`menu-screen-status`) |
| CB-3 | แกนทะเบียนเช็ค + event + เครื่องสถานะ + ยอด derive + recalculate API | `backend/internal/cheque/{schema.go,register.go,events.go,state.go,balances.go,recalculate.go}` + `*_test.go` + integration test | L | CB-1, CB-2, D1, D8 |
| CB-4 | จอทะเบียนเช็ครับ/จ่าย แทนเอกสาร ERP ทั่วไป (ถ้อยคำตาม screenshot + ตารางอ้างอิง + ปุ่ม "เช็คเตรียมจ่าย") | `frontend/src/components/cheque-register-screen.tsx`(+test), `frontend/src/lib/erp-transaction.ts` (ลบ 2 config), `frontend/src/app/menu/main-menu-screen.tsx` (routing), `languages.tsv` | M | CB-3, workflow อื่นต้อง commit `erp-transaction` ก่อน |
| CB-5 | เอกสารเปลี่ยนสถานะ 8 ตัว + ลงบัญชี + แหล่งยอด cheque_total/net/fee + `accountfrom: bank_book` + ตัวเลือกใน UI mapping | `backend/internal/cheque/{docs.go,posting.go}`+tests, `backend/internal/generalledger/{models.go,mapping_posting.go}`+tests, `frontend/src/app/gl/gl-masters.tsx`, `frontend/src/lib/general-ledger.ts` | L | CB-3, workflow อื่นต้อง commit generalledger ก่อน |
| CB-6 | จอเอกสารเปลี่ยนสถานะ (จอเดียว 8 config) + ลบ 8 config ERP | `frontend/src/components/cheque-status-doc-screen.tsx`(+test), `frontend/src/lib/erp-transaction.ts`, `frontend/src/app/menu/main-menu-screen.tsx`, `languages.tsv` | M | CB-5, F2 (`facade-screens.md`) |
| CB-7 | รายละเอียดเช็คใต้รายวัน GL (`details.cheques` → `gl_journal_cheques`) | `backend/internal/generalledger/{subledger_types.go,subledger.sql,subledger_records.go}`, จอรายวัน GL (แท็บเช็ค) | M | CB-1 |
| CB-8 | B2 จอการเคลื่อนไหวเช็ค + API movement | `backend/internal/cheque/movement.go`+test, `frontend/src/components/cheque-movement-screen.tsx`(+test) | S | CB-3, F2 |
| CB-9 | B3 ใบเตรียมเลขเช็ค + การใช้/คืนเลข | `backend/internal/cheque/prepare.go`+test, `frontend/src/components/cheque-prepare-screen.tsx`(+test) | M | CB-3, F2, D13 |
| CB-10 | B4 รายละเอียดธนาคาร (บรรทัดข้อความ + log) | `backend/internal/cheque/bank_service.go`+test, `frontend/src/components/bank-service-notes-screen.tsx`(+test) | S | CB-1, F2 |
| CB-11 | `pdftext.Overlay.Line` + test | `backend/internal/pdftext/pdftext.go`, `pdftext_test.go` | S | — |
| CB-12 | B1 ตำแหน่งพิมพ์ + พิมพ์ตรวจตำแหน่ง + พิมพ์เช็ค (PDF) | `backend/internal/cheque/{print_layout.go,print.go}`+tests, `frontend/src/components/{cheque-print-dialog.tsx,cheque-print-layout-dialog.tsx}`(+tests) | L | CB-3, CB-11, D4 |
| CB-13 | รายงานเช็ค 11 ชื่อ + รายวัน 6 ชื่อ ผ่าน report registry (ยึดรายชื่อใน `menuconfig.xml:329-362`; PDF ที่ไม่อยู่ในเมนู เช่น ประวัติเช็ครับ / ตามพนักงาน / ตามสถานะเช็ค ใช้เป็นแบบคอลัมน์เท่านั้น ตาม `facade-screens.md` §5 และ `report-registry.md:16-18`) | `backend/internal/cheque/reports.go`, ไฟล์ registry กลาง (ตามงานรายงานกลาง), `frontend/src/lib/erp-reports.ts` (ย้ายเข้า registry) | M | CB-5, งาน report registry, workflow อื่น (`erp-reports`) |
| CB-14 | ใบเสร็จ/ใบจ่าย ERP สร้างเช็คในทะเบียน + `cheque_source_links` + ตัดยอดเช็คเดิม ใน tx เดียวกัน | `backend/internal/transaction/httpapi/{save.go,posting.go}`, `backend/internal/cheque/links.go`+tests | L | CB-3, CB-9, workflow อื่นต้อง commit transaction ก่อน |

| CB-15 | ปิดทางเดิม: backend `/transaction/:module` ปฏิเสธ doctype เช็คทั้ง 10 ตัว (คง doctype ไว้เป็น key เลขเอกสาร/รูปแบบลงบัญชี) ใน change เดียวกับ CB-4/CB-6 | `backend/internal/transaction/httpapi/http.go` (`lookupDocType` หรือรายการ doctype ที่โมดูลอื่นเป็นเจ้าของ) + test, `languages.tsv` | S | CB-4/CB-6, workflow อื่นต้อง commit transaction ก่อน |
| CB-16 | ผู้ติดต่อธนาคาร 5680 (ตาราง `bank_contacts` + API + จอ `bank-contact-screen.tsx` แทนจอรอพัฒนา) และตารางอ่านอย่างเดียวบนจอสมุด | `backend/internal/cheque/{schema.go,bank_contacts.go}`+test, `frontend/src/components/bank-contact-screen.tsx`(+test), `menu-screen-status.ts`, `languages.tsv` | S | CB-1 |

**ลำดับที่แนะนำ:** CB-11, CB-1 → CB-2 / CB-16 → CB-3 → CB-5 → CB-4/CB-6 + CB-15 (change เดียวกัน) → CB-8/CB-9/CB-10 → CB-12 → CB-7 → CB-13 → CB-14

---

## 9. ส่วนเสริมเล็กที่ต้องรายงาน (additive, ไม่ขวางทาง Champ)

1. `bank_code` ตัวกรองแบบไม่บังคับบนรหัสสาขาธนาคาร (ไม่ใช่ key)
2. log ของแท็บ "บริการธนาคาร" (Champ log เฉพาะแท็บ 1–2)
3. ปุ่ม "พิมพ์ตรวจตำแหน่ง" และการตรวจแถบ MICR/ขนาดตัวอักษรตามมาตรฐาน ธปท.
4. ตัวเลือกขีดคร่อม / ข้อความคร่อม / ขีดฆ่า "หรือผู้ถือ" (ค่าเริ่มต้นปิด)
5. แสดงจำนวนครั้งที่พิมพ์ในจอเช็คจ่าย (Champ เก็บ PrintCount แต่ไม่ได้ยืนยันว่าแสดง)
6. ประวัติการแก้ไขทะเบียนและเอกสาร (event) + เหตุผลยกเลิกบังคับ (กฎ audit trail)
7. ไม่มีการเบี่ยงตาม D8 แล้ว (ลุงจืดเลือกทำตาม Champ 2026-10-11: ถามยืนยันแล้วแก้ได้ + แก้สถานะในทะเบียนได้, บันทึก event ทุกครั้ง) — การห้ามลบเช็คที่ถูกอ้างอิง/มาจากเอกสารอื่นเป็น parity ไม่ใช่ส่วนเสริม
8. "เปลี่ยนเลขที่บัญชี" แก้เฉพาะ `account_number` บนรหัสสมุดที่คงที่ ไม่ cascade ทุกตารางแบบ `CASHDlgChangeBookNo.cpp:59-130` (การเบี่ยงโดยเจตนา เพราะ BC ผูกด้วยรหัสสมุดที่ไม่เปลี่ยน)

## 10. สิ่งที่ยังไม่ verify

| เรื่อง | ทำไม |
|---|---|
| กติการันเลขเตรียมจ่ายของ Champ | ไม่มีซอร์สฟอร์มใน repo (มีแค่ DDL + ภาพ) |
| ผลต่อสมุดของ 5402 CancelStatus 2 และ 5405 ไม่ติ๊ก | อ่านโค้ด ไม่ได้รัน (`05-cheque.md` §10) |
| ป้ายไทยทีละช่องของจอ เช็ครับคืน 421, นำเช็คเข้าใหม่ 422, ขายลด 424, เช็คจ่ายผ่าน 425, เตรียมจ่าย | ทะเบียนรับ/จ่าย, รับผ่าน 420, นำฝาก 419, ยกเลิก 423/426, สมุดเงินฝาก, ผู้ติดต่อ เทียบภาพ `bcspec` แล้ว; จอที่เหลือมีภาพใน `bcspec/หน้าจอเมนูเช็ค บัตร/` แต่ยังไม่ได้เทียบทีละป้าย → เทียบตอนทำ CB-6/CB-8 |
| ตารางผู้ติดต่อบนจอสมุดเงินฝาก 5169 แก้ไขได้หรือไม่ | โค้ดตั้งอ่านอย่างเดียว (`CASHFrmBankBookNew.cpp:292-296`) แต่ยังไม่ได้ลองรัน |
| การแยกแผนกในบรรทัด GL ของเอกสารเช็ค | ยังไม่ได้อ่าน `ProcessConditionGroupNN` ครบ |
| ตำแหน่งช่องผู้รับเงิน/ตัวอักษรของเช็คแต่ละธนาคาร | มาตรฐาน ธปท. ไม่กำหนด |
| ปี พ.ศ./ค.ศ. บน Date Field Indicator ของแต่ละธนาคาร | ขึ้นกับแบบพิมพ์ของธนาคาร |
| บั๊กป้ายสถานะเช็คจ่ายในจอ movement ของ Champ | อ่านโค้ด ไม่ได้รัน |

## ปรับปรุงหลังตรวจทาน (2026-10-11)

- P1 รับ (ปรับ): กติกา `bank_book` มีรหัสบัญชีสำรองใน MappingRule สำหรับ 423 ประเภท 0 (Champ เงื่อนไข 3 ข้ามบรรทัดที่ไม่มีบัญชี `gltlsprocess.cpp:7766-7824`); กรณีอื่นที่ไม่มีสมุดยังคืน error เพื่อรักษาเดบิต=เครดิต (เบี่ยงโดยเจตนา) — §4.4, §5
- P2 รับ: เพิ่ม CB-15 ปิด `/transaction/:module` สำหรับ doctype เช็ค 10 ตัวใน change เดียวกับ CB-4/CB-6 — §0, §3, §8, R12
- P3 รับ: ป้ายทะเบียนรับ/จ่ายตามภาพ `bcspec` (ภาพชนะ .rc) และแยก key รับ/จ่าย — §1.1, §4.9
- P4 รับ: เพิ่มตารางเอกสารอ้างอิงใต้หัวเช็ค (`cqfrmcheqin.cpp:192-196,1227-1300`) + `references[]` ใน API — §1.1, §4.7, §4.8
- P5 รับ: คืนปุ่ม "เช็คเตรียมจ่าย" ตามภาพ (`cqfrmcheqout.cpp:1817-1848`) ใช้ `used_cheque_id` — §1.1, §1.7, §4.8
- P6 รับบางส่วน: 419 ไม่มีค่าธรรมเนียม/ยอดสุทธิ + CHECK fee=0 + ท้ายจอยอดรวมอย่างเดียว + ป้าย "รวมค่าใช้จ่าย"; ไม่รับคอลัมน์หมายเหตุใหม่ เพราะ Champ แสดงคำอธิบายของทะเบียนแบบอ่านอย่างเดียว (`cqfrmcheqindeposit.cpp:319,484,500`) และ DDL ไม่มีคอลัมน์หมายเหตุบรรทัด — §1.2, §4.2
- P7 รับ: ห้ามลบเช็คที่ถูกอ้างอิง/มาจากเอกสารอื่น = parity (`cqfrmcheqin.cpp:438-449`, ลำดับบันทึกผ่าน 1–9); D8/R10 เหลือเฉพาะการห้ามแก้ฟิลด์หลักและสถานะ — §1.2, §4.3, §7, §9
- P8 รับ: ตัวกรอง 422/423 เพิ่มเงื่อนไขวันที่ ≤ วันที่เอกสาร และสมุดบังคับตามประเภท — §4.3, §5
- P9 รับ: ปุ่ม "เปลี่ยนเลขที่บัญชี" (ไม่ cascade, เบี่ยงโดยเจตนา) + ตารางผู้ติดต่ออ่านอย่างเดียว + CB-16 จอผู้ติดต่อธนาคาร 5680 — §1.1, §4.2, §4.7, §4.8, §8, §9
- P10 รับ: เพิ่ม D15 สกุลเงินต่างประเทศ; ระหว่างรอ backend รับเฉพาะ THB อัตรา 1 — §4.2, §6 R13, §7
- P11 ไม่รับ (บันทึกไว้): Champ ไม่มีจุดตั้ง isconfirm ของตาราง BCChq* ทำให้ด่าน confirm ไม่เคยทำงาน จึงไม่เพิ่มขั้นอนุมัติ — §4.3, §7 ตัดออก
- P12 รับบางส่วน: ข้อความแจ้งคงที่เพิ่ม key `cheque_cancel_notice` เฉพาะจอ 423 ตามภาพ; ภาพ 426 ไม่มีข้อความนี้ — §4.8, §4.9
- P13 ไม่รับ: ตัดสินแล้วว่าเมนูยึด `menuconfig.xml` และ PDF ใช้เป็นแบบคอลัมน์ (`facade-screens.md` §5, `report-registry.md:16-18`); เพิ่มหมายเหตุใน CB-13 — §8
