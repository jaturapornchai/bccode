# 05 — ระบบเช็ค (CHQSYS) : พิมพ์เขียว Champ สำหรับ BC

> ที่มา: ลุงจืดสั่ง 2026-10-07 "ออกแบบใหม่ให้ Champ เป็นหลัก ให้เหมือนกันแทบทุกอย่าง" — เอกสารนี้ศึกษาเมนู CHQSYS ทั้ง 38 รายการของ Champ (`D:\project-champ`) แล้วเทียบกับ BC (`D:\bccode`)
> ขอบเขต: เมนู 38 รายการ = เอกสารหลัก 3 + เอกสารเปลี่ยนสถานะ 10 + กระบวนการ 1 + รายงาน 24 (atlas `scratch/champ-atlas/atlas-05-CHQSYS.md`)
> วันที่ศึกษา 2026-10-07 — ผลการอ่านโค้ด ไม่ได้รันโปรแกรม Champ จริง; จุดที่ไม่ได้พิสูจน์ระบุว่า "ยังไม่ verify"

## 0. วิธีอ้างอิง

| ตัวย่อ | ตำแหน่งจริง |
|---|---|
| `C:` | `D:\project-champ\champ\champ\` (ฟอร์ม/เมนู: `cqfrm*.cpp`, `cashfrm*.cpp`, `MainFrm.cpp`, `menuconfigxml\menuconfig.xml`) |
| `LIB:` | `D:\project-champ\champ-lib\BC5PROCESS\` (`BookbankProcess.cpp`, `gltlsprocess.cpp`, `bcprocess.cpp`) |
| `REP:` | `D:\project-champ\champ-report\BC5REP_CASH\` (รายงาน `CHMCQRep*`, `CASHRep*`, `BCRepCashChq.rc`) |
| `DDL:` | `champ\champ\Script\SQLSERVER_Script.sql` (ผ่าน atlas.json -> tables) |

- ข้อความไทยในรายงาน/ฟอร์มที่เป็น resource key (เช่น `CHEQ_NO`, `SUM_NETT`) แต่ละ key ไม่มีข้อความไทยใน repo (ฐาน resource แยก) — ตารางจึงเขียน key ไว้ และเขียนไทยเฉพาะที่ยืนยันจาก rc/คอมเมนต์/atlas (ที่ไม่ยืนยันเขียนว่า "ยังไม่ verify")
- เลขบรรทัดอ้างอิงจากไฟล์ที่ถอดรหัสไทยแล้ว (จำนวนบรรทัดเท่าต้นฉบับ)

## 1. เมนู CHQSYS ตามลำดับของ Champ (menuconfig.xml)

กลุ่มคั่นด้วยเส้น separate เหมือนเมนูจริง (`menuconfigxml\menuconfig.xml`, popup "CHQSYS"):

| กลุ่ม | id | ชื่อเมนู | ฟอร์ม/Handler | ตารางหลัก |
|---|---|---|---|---|
| A. เอกสารหลัก | 5377 | บันทึกเช็ครับ | `CCQFrmCheqIn` (cqfrmcheqin.cpp) | BCChqIn |
| | 5378 | บันทึกเช็คจ่าย | `CCQFrmCheqOut` | BCChqOut |
| | 5379 | บันทึกบัตรเครดิต | `CCQFrmCreditCard` | BCCreditCard |
| B. เช็ครับ | 5398 | บันทึกนำฝากเช็ครับ | `CCQFrmCheqInDeposit` | BCChqInDeposit / BCChqInDeposSub |
| | 5401 | บันทึกเช็ครับผ่าน | `CCQFrmCheqInPass` | BCChqInPass / BCChqInPassSub |
| | 5408 | บันทึกเช็ครับคืน | `CCQFrmCheqInReturn` | BCChqInReturn / BCChqInRetSub |
| | 5399 | บันทึกนำเช็คเข้าใหม่ | `CCQFrmCheqInDeposit2` | BCChqInDeposit2 / BCChqInDepo2Sub |
| | 5402 | บันทึกยกเลิกเช็ครับ | `CCQFrmCheqInCancel` | BCChqInCancel / BCChqInCancelSub |
| | 5403 | บันทึกขายลดเช็ครับ | `CCQFrmCheqInSale` | BCChqInSale / BCChqInSaleSub |
| C. เช็คจ่าย | 5404 | บันทึกเช็คจ่ายผ่าน | `CCQFrmCheqOutPass` | BCChqOutPass / BCChqOutPassSub |
| | 5405 | บันทึกยกเลิกเช็คจ่าย | `CCQFrmCheqOutCancel` | BCChqOutCancel / BCChqOutCancelSub |
| D. บัตรเครดิต | 5406 | บันทึกขึ้นเงินบัตรเครดิต | `CCQFrmCreditPass` | BCCreditPass / BCCreditPassSub |
| | 5407 | บันทึกยกเลิกบัตรเครดิต | `CCQFrmCreditCancel` | BCCreditCancel / BCCreditCancelSub |
| E. กระบวนการ | 5400 | คำนวณยอดคงเหลือของเช็คใหม่ | `CMyReCalcCheqe` (dialog) | BCChqIn/Out + BCAR/BCAP |
| F. รายงานข้อมูลหลัก (popup `disable="yes"`) | 5380-5384 | รายงานเช็ครับ 5 แบบ | 5380 `CCHMCQRepChqInView`; 5381-5384 `CCQRepChqInBy{DueDate,Ar,ChqNo,Depart}View` | อ่านอย่างเดียว |
| | 5386-5391 | รายงานเช็คจ่าย 6 แบบ | 5386 `CCHMCQRepChqOutView`; 5387-5391 `CCQRepChqOutBy{DueDate,Ap,ChqNo,Depart,BankBook}View` | |
| | 5393-5397 | รายงานบัตรเครดิต 5 แบบ | 5393 `CCHMCQRepCreditcardView`; 5394-5397 `CCQRepCardBy{DueDate,Ar,CardNo,Depart}View` | |
| G. รายงานรายวัน (popup `disable="yes"`) | 5409-5416 | รายงานรายวัน 8 รายการ | `OnChqsysTransRep1..8` | |

หมายเหตุเมนู: popup ที่ `disable="yes"` ไม่แสดงใน main menu (`XMLMenuConfig.cpp:153`) แต่ไปแสดงในศูนย์รายงาน (`reportCenterView.cpp:299-330`) — BC ต้องเอารายงานกลุ่ม F/G ไปไว้ใน "ศูนย์รายงาน" ไม่ใช่เมนูหลัก (รายละเอียดดูเอกสารกรอบ UI กลาง)

## 2. สถานะของเช็ค (หัวใจของระบบ)

### 2.1 เช็ครับ — `BCChqIn.Status` (SmallInt)

ข้อความไทยยืนยันจาก dialog รายงาน 35032 (`REP:BCRepCashChq.rc`, combo CB_CHEQ_TYPE_1..7) ลำดับ = ค่า 0..6:

| Status | ข้อความ | ทำให้เกิดโดย | ตัวกรองเช็คที่เลือกได้ในเอกสารนั้น |
|---|---|---|---|
| 0 | เช็คในมือ | บันทึกเช็ครับ (5377) / ใบเสร็จรับเงินสร้างให้อัตโนมัติ | ค่าเริ่มต้น |
| 1 | เช็คฝากธนาคาร | 5398 นำฝาก | ต้องเป็นเช็ค Status=0 และ ReceiveDate <= วันที่เอกสาร |
| 2 | เช็คผ่าน | 5401 เช็ครับผ่าน | เช็ค Status 0 หรือ 1 และ DueDate <= วันที่เอกสาร |
| 3 | เช็คคืน | 5408 เช็ครับคืน | เช็ค Status 2 หรือ 6 และ ReceiveDate <= วันที่เอกสาร |
| 4 | เช็คยกเลิก | 5402 ยกเลิกเช็ครับ | ดูตาราง 2.3 |
| 5 | ขายลดเช็ค | 5403 ขายลดเช็ครับ | เช็ค Status 0 หรือ 1 และ ReceiveDate <= วันที่เอกสาร |
| 6 | เช็คคืนฝากธนาคาร (ฟอร์มเรียก "นำฝากเช็คคืน"/นำเช็คเข้าใหม่) | 5399 นำเช็คเข้าใหม่ | เช็ค Status=3 |

หลักฐาน: ตัวกรอง `strExFilter` ของแต่ละฟอร์ม — Deposit `cqfrmcheqindeposit.cpp` (Status=0), InPass `cqfrmcheqinpass.cpp:742-752`, InReturn/InDeposit2/InSale/InCancel ในไฟล์ `cqfrmcheqinreturn.cpp`, `cqfrmcheqindeposit2.cpp`, `cqfrmcheqinsale.cpp`, `cqfrmcheqincancel.cpp:760-830`. ทุกตัวกรองมีเงื่อนไข `NOT IN` เช็คที่เลือกในตารางแล้ว (กันเลือกซ้ำ)

เส้นทางสถานะ (ลูกศร = เอกสารที่ทำให้เปลี่ยน):

```
          5398 นำฝาก        5401 ผ่าน
  0 ในมือ ----------> 1 ฝากธนาคาร ----------> 2 ผ่าน ---- 5408 คืน ----> 3 คืน ---- 5399 นำเข้าใหม่ ----> 6 คืนฝากธนาคาร
   |  \__________ 5401 ผ่านตรงจากในมือ ________/  ^                                                      |
   |                                                \____________ 5408 คืนได้จาก 2 หรือ 6 ____________/
   +-- 5403 ขายลด (จาก 0 หรือ 1) --> 5
   +-- 5402 ยกเลิก --> 4  (ยกเลิกได้จาก 0, 2, 1, 6 ตามประเภทที่เลือก)
```

ข้อสังเกต: เช็คสถานะ 6 ผ่านซ้ำด้วย 5401 ไม่ได้ (ตัวกรอง 5401 รับเฉพาะ 0/1) — ต้องคืนแล้ว (5408) หรือยกเลิก (5402 ประเภท 3) — "เปลี่ยนเช็ค" ไม่มีเมนูเฉพาะใน Champ; ทำโดย 5408 คืน + 5399 นำเข้าใหม่ หรือยกเลิก + บันทึกเช็คใหม่

### 2.2 เช็คจ่าย — `BCChqOut.Status`

| Status | ข้อความ | ที่มา |
|---|---|---|
| 0 | เช็คในมือ (CB_CHEQ_TYPE_1; ความหมายจริงคือ "ออกเช็คแล้วยังไม่ขึ้นเงิน") | บันทึกเช็คจ่าย 5378 |
| 1 | เช็คผ่าน (CB_CHEQ_TYPE_3) | 5404 เช็คจ่ายผ่าน (ตัวกรอง Status=0 และ DueDate <= วันที่เอกสาร, กรอง BookNo ได้) |
| 2 | เช็คยกเลิก (CB_CHEQ_TYPE_5) | 5405 ยกเลิกเช็คจ่าย (ติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว" = เลือกเช็ค Status=1 ต้องระบุสมุดเงินฝาก; ไม่ติ๊ก = Status=0; คอมเมนต์โค้ดเรียกช่องนี้ "ยกเลิกเช็คผ่านเท่านั้น" `cqfrmcheqoutcancel.cpp:442`) |
| 3 | เตรียมจ่าย (`PREPAIR_PAYMENT`, `cqfrmcheqout.cpp:153`) | ยังไม่ verify ว่าใช้งานจริง (ปุ่ม `IDC_BTN_PREPARE` ถูก comment ไว้ `cqfrmcheqout.cpp:94`) |

### 2.3 ยกเลิกเช็ครับ (5402) — `CancelStatus`

combo "เช็คที่ต้องการยกเลิก" (`IDC_CQ_STATUS`, `cqfrmcheqincancel.cpp:620-640`; rc ยังมีช่องติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว" แต่โค้ดบันทึก/ตรวจใช้ค่าจาก combo — ยังไม่ verify ว่าช่องติ๊กถูกซ่อน) พร้อมข้อความเตือนใต้ช่อง "***เช็คที่ยกเลิกไปแล้ว ไม่สามารถที่นำไปทำรายการได้อีก" (อ่านจาก atlas, ข้อความต่อท้ายถูกตัด):

| CancelStatus | ข้อความ | เช็คที่เลือกได้ | ต้องระบุสมุดเงินฝาก | สถานะกู้คืนเมื่อลบเอกสารยกเลิก |
|---|---|---|---|---|
| 0 | ยกเลิกเช็คในมือ | Status=0 | ไม่ (ช่องถูกปิดและล้าง) | 0 |
| 1 | ยกเลิกเช็คที่ผ่านแล้ว | Status=2 | ใช่ | 2 |
| 2 | ยกเลิกเช็คฝาก | Status=1 | (ไม่ถูกปิด) | 1 |
| 3 | ยกเลิกเช็คคืนฝากธนาคาร | Status=6 | ใช่ | 6 |

(ตาราง restore: `cqfrmbasetransstatus.cpp:147-165`; เช็คที่ถูกยกเลิกจะมี `Status=4`)

### 2.4 บัตรเครดิต — `BCCreditCard.Status`

0 = ยังไม่ขึ้นเงิน (ค่าเริ่มต้น), 1 = ขึ้นเงินแล้ว (5406), 2 = ยกเลิก (5407) — ข้อความไทยของ combo `CB_CREDITCARD_TYPE_1..3` ยังไม่ verify. ตัวกรอง: CreditPass = Status 0 และไม่อยู่ใน BCCreditPassSub, CreditCancel = Status 0 (`cqfrmcreditpass.cpp`, `cqfrmcreditcancel.cpp`)

### 2.5 ตารางสรุปการเปลี่ยนสถานะ + ผลต่อยอด (จากโค้ด engine)

ฟังก์ชันกลาง `ProcessChqStatus(ตารางย่อย, ตารางเช็ค, DocNo, nDeleteMode, nBookMark, nChqStatus, nReChqStatus)` (`cqfrmbasetransstatus.cpp:303-363`):

- `nDeleteMode = 0` (บันทึก): ย้อนรายการเดิม (TransState=1) แล้วลบ แล้วลงรายการใหม่ (TransState=0)
- `nDeleteMode = 1` (ลบเอกสาร): ย้อนรายการ TransState=0 แล้ว `DELETE` แถวย่อยทั้งหมดของ DocNo
- `nDeleteMode = 2` (ยกเลิกเอกสาร): ย้อนแถวย่อยที่ `IsCancel=1`
- `nReChqStatus` ไม่ถูกใช้ใน `UpdataCheqStatus` (โค้ดใช้ GetChqInStatus คำนวณใหม่แทน, บรรทัดที่เคยใช้ถูก comment ไว้ `:189`) — BC ไม่ต้องมีพารามิเตอร์นี้

ฟังก์ชัน `UpdataCheqStatus` (`:119-262`) ทำเมื่อ nMark=+1 (ลงรายการ) หรือ −1 (ย้อน) ต่อเช็คทีละใบ:
1. อัปเดตเช็ค (`WHERE RefChqRowOrder` ของเช็ครับ/จ่าย, บัตรใช้ `WHERE RowOrder`): `Status` = สถานะใหม่, `StatusDocNo` = เลขที่เอกสาร, `StatusDate` = วันที่เอกสาร, `BookNo` = สมุดเงินฝาก
2. เมื่อย้อน (nMark=−1): `Status` คำนวณใหม่จากเอกสารย่อยที่ยังเหลือ (`GetChqInStatus/GetChqOutStatus` `:264-300` = UNION ของตารางย่อยเรียง DocDate ลดลง, MyType ลดลง); ถ้าผลเป็น 0 สำหรับเช็ครับ จะล้าง `BookNo`; `StatusDocNo/StatusDate` ถูกล้าง
3. `BCBankBook.Balance += NetAmount * nMarkStatus` โดย `nMarkStatus` = nMark แต่กลับเครื่องหมายสำหรับ เช็คจ่ายผ่าน (status 1) และ เช็ครับคืน (status 3) และสำหรับยกเลิกเช็ครับ CancelStatus=1 ใช้ `nBookMark`
4. ยอดลูกหนี้/เจ้าหนี้: ดูตาราง 2.6

| เอกสาร | Status ใหม่ | nBookMark ที่ส่ง (บันทึก) | ผลต่อ BCBankBook.Balance (ลงรายการ) | บรรทัดเรียก |
|---|---|---|---|---|
| 5401 เช็ครับผ่าน | 2 | 1 | + NetAmount | `cqfrmcheqinpass.cpp:1328` |
| 5408 เช็ครับคืน | 3 | -1 | − NetAmount (flip) | `cqfrmcheqinreturn.cpp:1315` |
| 5399 นำเช็คเข้าใหม่ | 6 | 1 | + NetAmount | `cqfrmcheqindeposit2.cpp:1288` |
| 5403 ขายลดเช็ครับ | 5 | 1 | + NetAmount | `cqfrmcheqinsale.cpp:1232` |
| 5402 ยกเลิก CancelStatus 1 หรือ 3 | 4 | -1 | CancelStatus 1: − NetAmount; CancelStatus 3: ดูข้อสังเกต | `cqfrmcheqincancel.cpp:1450` |
| 5402 ยกเลิก CancelStatus 0 หรือ 2 | 4 | 0 | + NetAmount บน BookNo ของแถว (0 = ว่าง จึงไม่กระทบ; 2 ดูข้อสังเกต) | `:1452` |
| 5404 เช็คจ่ายผ่าน | 1 | -1 | − NetAmount (flip) | `cqfrmcheqoutpass.cpp:1248` |
| 5405 ยกเลิกเช็คจ่าย | 2 | 1 หรือ 0 | + NetAmount | `cqfrmcheqoutcancel.cpp:1227-1229` |
| 5406 ขึ้นเงินบัตรเครดิต | 1 | 1 | + NetAmount | `cqfrmcreditpass.cpp:1183` |
| 5407 ยกเลิกบัตรเครดิต | 2 | -1 | − NetAmount | `cqfrmcreditcancel.cpp:1028` |
| 5398 นำฝากเช็ครับ (ฟอร์มมี UpdataCheqStatus ของตัวเอง) | 1 | - | ไม่แตะ Balance, ไม่ลง GL | `cqfrmcheqindeposit.cpp:1181-1217` |

ข้อสังเกตจากการอ่านโค้ด (ยังไม่ verify โดยรันจริง): Champ เก็บ `BCBankBook.Balance` เป็นยอดสะสมแบบบวก/ลบทีละเอกสาร ส่วนสูตรคำนวณใหม่ (`LIB:BookbankProcess.cpp:60-200`, ดู 06-cash-bank.md) หักยกเลิกเช็ครับ CancelStatus 1 และ 3 และบวกยกเลิกเช็คจ่าย CancelStatus 1 และ 3 — แต่ `UpdataCheqStatus` ของ 5402 กับ CancelStatus 3 บวก NetAmount เข้ายอดสะสม (nMark คงเป็น +1; มีการกลับเครื่องหมายเฉพาะ CancelStatus 1 `:150-156`) ซึ่งสวนกับสูตรคำนวณใหม่; ส่วน 5402 กับ CancelStatus 2 และ 5405 กับ CancelStatus 0 เมื่อระบุ BookNo จะบวก NetAmount เข้ายอดสะสมโดยสูตรคำนวณใหม่ไม่นับเลย; และเมื่อลบเอกสาร 5402 (CancelStatus 1/3) โค้ดส่ง nChqStatus=2 ทำให้ไม่กลับเครื่องหมาย (`cqfrmcheqincancel.cpp:459-460`) ยอดสะสมอาจถูกหักซ้ำ — จึงมีโอกาสที่ยอดสะสมคลาดเคลื่อน และเป็นเหตุผลที่ Champ ต้องมีเมนู "คำนวณยอดคงเหลือของสมุดบัญชีใหม่" (5170). BC ไม่ควรลอก "ยอดสะสมแบบบวกลบ" ให้คำนวณจากรายการจริง (derived) เสมอ

### 2.6 ผลต่อลูกหนี้/เจ้าหนี้ (ยอดเช็ค)

(`cqfrmbasetransstatus.cpp:58-116`, เรียกจาก `:226-258`)

| ตารางที่ปรับ | เงื่อนไข | สูตร |
|---|---|---|
| `BCAR.ChqOnHand` | เช็ครับ status 0/1/2/5 และยกเลิก CancelStatus 0/2 | `ChqOnHand -= ChqAmount * nMarkStatus` |
| `BCAR.ChqReturn` | เช็ครับ status 3 (คืน) และ 6 (นำเข้าใหม่) | `ChqReturn -= ChqAmount * nMarkStatus` (คืน flip เครื่องหมาย = เพิ่มยอดคืน) |
| `BCAR.DebtLimitBal` | เมื่อ `CreditLimitFlag=1` และ `DebtLimit1 > 0` ตอน ChqOnHand ถูกปรับ | `DebtLimitBal += ChqAmount * nMarkStatus` (คืนวงเงินเมื่อเช็คผ่าน) |
| `BCAR.CreditOnHand` | บัตรเครดิต status 1 (ขึ้นเงิน) | `CreditOnHand -= CreditAmount * nMarkStatus` + คืนวงเงินเช่นเดียวกัน |
| `BCAP.ChqOnHand` | เช็คจ่าย status 1 (ผ่าน) | `ChqOnHand += ChqAmount * nMarkStatus` (nMark ถูก flip จึงลดยอด) |
| `BCAP.ChqReturn` | เช็คจ่าย status 2 (ยกเลิก) | `ChqReturn += ChqAmount * nMarkStatus` |

BC: ยอดเหล่านี้ใน Champ เป็น "ฟิลด์สะสมบนตารางลูกหนี้/เจ้าหนี้" — สำหรับ BC ที่ยึดหลัก derived + PostgreSQL ให้คำนวณจากตารางเช็คตอนอ่าน (ดูข้อ 8)

## 3. เอกสารหลัก (Master) 3 รายการ

### 3.1 บันทึกเช็ครับ (5377) — `CCQFrmCheqIn` (`cqfrmcheqin.cpp`) — ตาราง `BCChqIn`

ฟอร์ม `IDD_CQ_FRM_CHEQIN` 348x244 (`BC5Account.rc:6601`). ลำดับฟิลด์จริง (ซ้าย/ขวาเรียงตามแถว y):

| แถว | ฟิลด์ซ้าย | ฟิลด์ขวา |
|---|---|---|
| 1 | เลขที่เอกสาร (+ปุ่มค้นหา) | วันที่รับเช็ค (+ปุ่มปฏิทิน) |
| 2 | รหัสลูกหนี้ (+ปุ่ม) + ชื่อลูกหนี้ (อ่านอย่างเดียว) | - |
| 3 | วันที่ถึงกำหนด (+ปุ่มปฏิทิน) | สถานะ (combo 7 ค่า) |
| 4 | เลขที่เช็ค | สถานะเพิ่มเติม (combo จาก `BCChqStatus`) |
| 5 | รหัสธนาคาร (+ปุ่ม) | รหัสสกุลเงิน |
| 6 | รหัสสาขาของธนาคาร (+ปุ่ม) | อัตราแลกเปลี่ยน |
| 7 | รหัสสมุดเงินฝาก (+ปุ่ม) | จำนวนเงินบนเช็คตามสกุลเงิน |
| 8 | - | จำนวนเงินบนเช็ค (เงินบาท) |
| 9 | คำอธิบาย | - |
| 10 | รหัสพนักงานขาย (`BCSale` ที่ ActiveStatus<>0) | รหัสแผนก |
| 11 | ข้อมูลการปรุบปรุง (อ่านอย่างเดียว; สะกดตามต้นฉบับ) | - |

(atlas `atlas-05-CHQSYS.md:6-22`)

| ฟิลด์ | คอลัมน์ | บังคับ | หมายเหตุ/การตรวจ |
|---|---|---|---|
| เลขที่เอกสาร | DocNo | ใช่ | CANNOT_SAVE_NODOCNO |
| วันที่รับเช็ค | ReceiveDate | ใช่ | CANNOT_SAVE_NO_CHQDAY |
| รหัสลูกหนี้ | ArCode | ใช่ | lookup ลูกหนี้; CANNOT_SAVE_NOTARCODE |
| วันที่ถึงกำหนด | DueDate | ใช่ | CANNOT_SAVE_NO_DUEDATE |
| สถานะ | Status | - | 0..6 ดู 2.1; ถ้า Status<>0 ต้องมีรหัสสมุดเงินฝาก (CANNOT_SAVE_NO_BOOKBANK) |
| เลขที่เช็ค | ChqNumber | ใช่ | ซ้ำไม่ได้เมื่อ (ChqNumber + BankCode + BankBranchCode) ซ้ำใน BCChqIn — ตรวจตอนออกจากช่อง (`:1474-1520`) และตอนบันทึก (`IsDuplicate`, ข้อความ "Chq Number Duplicate !!!") |
| สถานะเพิ่มเติม | ExtendStatus | - | master `BCChqStatus` (CODE, NAME, UserGroup); ประวัติเปลี่ยนที่ `BCChqStatusLog` |
| รหัสธนาคาร | BankCode | ใช่ | lookup `BCBank`; CANNOT_SAVE_NO_BANKID |
| รหัสสาขาของธนาคาร | BankBranchCode | ใช่ | lookup `BCBankBranch`; CANNOT_SAVE_NO_BRANCEID |
| รหัสสกุลเงิน | CurrencyCode | - | ออกจากช่อง -> ดึง `BCCurrency.ExchangeRate` มาใส่ (`:1141-1153`) |
| อัตราแลกเปลี่ยน | ExchangeRate | - | ถ้า `MultiCurrency<>1` หรือว่าง = 1.0 |
| รหัสสมุดเงินฝาก | BookNo | เมื่อ Status<>0 | lookup `BCBankBook` |
| จำนวนเงินบนเช็คตามสกุลเงิน | Amount | ใช่ | CANNOT_SAVE_NO_AMOUNT |
| จำนวนเงินบนเช็ค (เงินบาท) | (คำนวณ) | - | = อัตราแลกเปลี่ยน x จำนวนเงิน อัปเดตทุกครั้งที่ amount/rate เปลี่ยน (`:1129-1139`), อ่านอย่างเดียว |
| คำอธิบาย | MyDescription | - | |
| รหัสพนักงานขาย | SaleCode | - | |
| รหัสแผนก | DepartCode | - | |
| ข้อมูลการปรุบปรุง | (GetUpdateInfo) | - | แสดง StatusDocNo/StatusDate ของเอกสารที่เปลี่ยนสถานะล่าสุด (`:1160-`) |

การกันแก้ไข: ถ้าเช็คถูกอ้างอิงแล้ว (ใน `BCRecMoney` PaymentType=2 หรือตารางย่อย Depos/Pass/Cancel/Ret/Sale — `IsReferance` `:1329-1350`) จะถามยืนยัน CANNOT_UPDATE_REFERENCE (ใช่/ไม่ใช่) ก่อนแก้; ถ้าเช็คมาจากเอกสารอื่น (SaveFrom<>4 และมี DocNo) ถามยืนยัน CANNOT_UPDATE_REFERENCE_OTHER (`:311-340`)
คอลัมน์เพิ่มในตาราง: `SaveFrom` (4 = คีย์จากจอนี้; ค่าอื่นมาจากเอกสารต้นทาง เช่นใบเสร็จรับเงิน ยังไม่ verify ความหมายทุกค่า), `ChqUseStatus` (0 = ใช้ครั้งเดียว, 1 = ใช้หลายครั้ง มี `Amount`/`Balance`), `Balance`, `ConfirmCode/ConfirmDateTime`, `CancelCode/CancelDateTime`, `PrintCount`, `ismanualkey`, `RefChqRowOrder` (คีย์อ้างอิงที่ตารางย่อยใช้), `IsSAP` (เฉพาะลูกค้าบางราย)
เช็ครับถูกสร้างอัตโนมัติจากใบเสร็จรับเงิน: `ARFrmReceiptMultiAr.cpp:5930-5995`
ปุ่ม/คีย์: `ID_HK_NEW`, `ID_HK_FIND` (ค้นหา), `IDC_MORE`, ปุ่มข้างช่องเปิด lookup (`cqfrmcheqin.cpp:97-99`)

### 3.2 บันทึกเช็คจ่าย (5378) — `CCQFrmCheqOut` — ตาราง `BCChqOut`

เลย์เอาต์เดียวกับเช็ครับ (rc ยังพิมพ์ป้าย "รหัสลูกหนี้"/"วันที่รับเช็ค" แต่โค้ดเปลี่ยนความหมาย — ป้ายที่ใช้จริง ยังไม่ verify เพราะอยู่ใน resource key; ใช้ "รหัสเจ้าหนี้" และ "วันที่จ่าย" ตามคอลัมน์ `ApCode`/`PaymentDate`)

ต่างจากเช็ครับ:

| ประเด็น | เช็คจ่าย |
|---|---|
| คู่ค้า | รหัสเจ้าหนี้ (`ApCode`, lookup เจ้าหนี้) |
| วันที่ | วันที่จ่าย (`PaymentDate`, ค่าเริ่มต้นเป็นวันที่ทำงาน) |
| รหัสพนักงานขาย | ไม่มี |
| รหัสสมุดเงินฝาก | บังคับเสมอ (CANNOT_SAVE_NO_BOOKBANK, `cqfrmcheqout.cpp:~380`) |
| ตรวจวันที่ | วันที่จ่าย > วันที่ถึงกำหนด = ผิด ข้อความ MSG_CHQOUT_DUEDATE_ERR (ล้างวันที่ถึงกำหนดและย้ายโฟกัส) `:1475-1560` |
| เลขซ้ำ | MSG_CHQOUT_DUPLICATE |
| พิมพ์ | ปุ่มพิมพ์เปิด `LIBDlgPrintChqOut`: เลือกช่วง เลขที่เอกสาร/เลขที่เช็ค, เลือกฟอร์ม, ติ๊ก Document Form / GL Slip Form / Withhold. Tax Form / Print Serial Number |
| คอลัมน์เพิ่ม | `AutoDocNo` (จากกลุ่มสร้างเช็คอัตโนมัติ `BCChqOutAuto`: ChqNoStart, ChqRun — ไม่อยู่ในเมนู 38 รายการ ยังไม่ verify ทางเข้า) |

เช็คจ่ายถูกสร้างอัตโนมัติจากใบสำคัญจ่าย (`ARFrmReceiptMultiAr.cpp`-ตระกูลเดียวกัน, `LIBMoneyChqPage.cpp:122-125`, `CASHFrmOtherExpense.cpp:2156` เขียน BCChqOut)

### 3.3 บันทึกบัตรเครดิต (5379) — `CCQFrmCreditCard` — ตาราง `BCCreditCard`

เหมือนเช็ครับแต่: เลขที่บัตรเครดิต (`CreditCardNo`) แทนเลขที่เช็ค, ไม่มี "สถานะเพิ่มเติม", มีคอลัมน์เฉพาะ `CreditType`, `ConfirmNo`, `ChargeAmount`, `CreditVatRate`, `CreditVat`, `CreditSumVat`, และฟิลด์เครื่อง EDC (`EDCBatchNo, EDCCardHolder, EDCCardExpr, EDCTID, EDCSystemTrc, EDCNII, EDCNetAmount`, `POSDOCNO`, `WalletType`, `SaveType`) (`cqfrmcreditcard.cpp:117-155`; DDL ใน atlas.json). สถานะ combo = CB_CREDITCARD_TYPE_1..3 (ข้อความไทยยังไม่ verify)

## 4. เอกสารเปลี่ยนสถานะ (Trans) — โครงร่วม

ทุกฟอร์มสืบทอด `CCQFrmBaseTransStatus` (`cqfrmbasetransstatus.cpp`) ใช้กริด `CBInputListWnd` (เลือกเช็คเป็นแถว) และมีโครงดังนี้

หัวเอกสาร (ลำดับ): เลขที่เอกสาร -> วันที่เอกสาร -> รหัสสมุดเงินฝาก (combo `BCBankBook` คอลัมน์ BookNo/Name; ฟอร์มนำฝาก 5398 ก็มี) -> หมายเหตุ -> ข้อมูลผู้สร้าง/ผู้แก้/ผู้ยืนยัน
(ตำแหน่งจริงตาม rc: 5401/5404 = แถว1 เลขที่เอกสาร + วันที่เอกสาร, แถว2 รหัสสมุดเงินฝาก, แถว3 หมายเหตุ; 5402 = แถว2 combo "เช็คที่ต้องการยกเลิก" + รหัสสมุดเงินฝาก, แถว3 หมายเหตุ, แถว4 ช่องติ๊ก; 5405 = แถว2 รหัสสมุดเงินฝาก + ช่องติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว"; ปุ่มข้างช่องยอดรวม 3 ช่อง `AMOUNTBTN/FEEBTN/NETBTN` ยังไม่ verify หน้าที่; 5406/5407 ใช้บัตรเครดิต)

กริดรายการ (ลำดับคอลัมน์ เหมือนกันทั้งกลุ่มเช็ค ตัวอย่าง `cqfrmcheqinpass.cpp:594-619`):

| # | key | ความหมาย | การแก้ไข |
|---|---|---|---|
| 0 | CHEQ_NO | เลขที่เช็ค | พิมพ์เลขเช็ค หรือกดลิงก์เปิดรายการเช็คที่เลือกได้ |
| 1 | BANK | ธนาคาร (+สาขา) | อ่านอย่างเดียว |
| 2 | SETTLE_DATE | วันที่ถึงกำหนด | อ่านอย่างเดียว |
| 3 | CHQ_AMOUNT | จำนวนเงินบนเช็ค | อ่านอย่างเดียว (เงินตามสกุล) |
| 4 | AMOUNT | จำนวนเงิน (เป็นเงินบาท) | อ่านอย่างเดียว |
| 5 | FEE_AMOUNT | ค่าใช้จ่าย | แก้ได้ |
| 6 | SUM_NETT | ยอดที่หักค่าใช้จ่ายแล้ว = AMOUNT - FEE (เช็คจ่ายใช้ `SUM_CHQOUT_NETT`) | คำนวณ |
| (7) | ISSAP | ช่องติ๊ก (เฉพาะลูกค้ารายที่เปิดค่า oomdts) | - |

บัตรเครดิตใช้ CREDITCARD_NO, BANK, SETTLE_DATE, CREDIT_AMOUNT, AMOUNT, FEE_AMOUNT, SUM_NETT
คอลัมน์เงา (ไม่แสดง) เก็บ: RowOrder, ChqRowOrder (บัตร: CreditRowOrder), CurrencyCode, ExchangeRate, OrgChqStatus, Arcode/Bankcode/BankBranchCode/ChqStatus/RefChqRowOrder (เช็คจ่าย: Bankcode/Branchcode/ApCode/RefChqRowOrder)
กริดบันทึกเฉพาะแถวที่เลขที่เช็คไม่ว่างและ `ChqAmount>0` (`cqfrmcheqinpass.cpp:~1286`)

ท้ายเอกสาร (footer): จำนวนเงินรวม (`SumChqAmount`), รวมยอดค่าใช้จ่าย (`SumExpend`), ยอดที่หักค่าใช้จ่ายแล้ว (`NetAmount`). ฟอร์มนำฝาก 5398 มีเฉพาะ "จำนวนเงินรวม". ฟอร์ม 5406 (ขึ้นเงินบัตร) เพิ่ม อัตราภาษี / ยอดก่อนภาษี / ยอดภาษี (`TaxRate`, `BeforeTaxAmount`, `TaxAmount`; `OnSaveInputTax(7,…,13)`).

ปุ่ม/คีย์ในเหตุการณ์กริด (`cqfrmcheqinpass.cpp:84-96`): ILDN_LINK (เปิดรายการเช็ค), SELCHANGE, CHANGEDATA, ADDNEW (เพิ่มแถว), DELETESELROW (ลบแถว), COMBO_CONDITION; `ID_HK_NEW`, `ID_HK_FIND`, `ID_HK_PRINT`, `IDC_RECUR` (เอกสารประจำ `RecurName`), `IDC_MORE`
สถานะเอกสาร (คอลัมน์ในตารางหัว): `IsConfirm` (ยืนยัน), `IsPostGL` (ลงบัญชีแล้ว), `IsCancel`, `GLFormat` (รูปแบบการลงบัญชี), `GLTransData` (0 = อัตโนมัติ 1 = ลงมือ `m_nManualKey`), `PrintCount`, `JobNo`

ตารางหัวมีคอลัมน์: `DocNo, DocDate, MyDescription, BookNo, GLFormat, GLStartPosting, IsPostGL, GLTransData, SumChqAmount, SumExpend, NetAmount, IsConfirm, RecurName, ConfirmCode, ConfirmDateTime, CancelCode, CancelDateTime, IsCancel, PrintCount, JobNo` (+ `CancelStatus` ใน BCChqInCancel; บัตร: `SumCreditAmount, TaxRate, BeforeTaxAmount, TaxAmount`). ตารางย่อยมี `ChqRowOrder, LineNumber, TransState, IsCancel, ChqNumber, ChqAmount, Expend, NetAmount, CurrencyCode, ExchangeRate, HomeAmount, OrgChqStatus, Arcode, Bankcode, BankBranchCode, RefChqRowOrder, IsSAP` (บัตร: `CreditRowOrder, CreditNumber, CreditAmount`). ตาราง `BCChqInDeposit` (5398) ไม่มีคอลัมน์ GL เลย (ไม่ลงบัญชี)

### 4.1 กติกาก่อนบันทึก (ตัวอย่างฟอร์มผ่าน — 5401) `cqfrmcheqinpass.cpp:314-436`

ห้ามบันทึก (แสดงข้อความ ไม่ผ่าน) เมื่อ:
1. เอกสารเช็คถูกอ้างอิง (MSG_CHQ_IS_REF1)
2. งวดบัญชีปิดแล้ว (`IsPeriodClose`)
3. ลงบัญชีแล้ว (`IsPostGL`)
4. ยืนยันแล้ว `IsConfirm=1` (CANNOT_UPDATE_CONFIRM)
5. ไม่มีเลขที่เอกสาร / วันที่เอกสาร
6. ยอดเช็คเป็น 0 หรือไม่มีแถว (MSG_CHEQAMOUNT_ERROR)
7. ไม่ได้เลือกสมุดเงินฝาก (CANNOT_SAVE_NO_BOOKBANK)
8. แก้ไขเอกสารเดิมที่มีเช็คสถานะ > 2 อยู่ (CANNOT_EDIT_REFERENCE)
9. ค่าใช้จ่ายรวม/สุทธิในหัว ไม่เท่ากับผลรวมกริด (MSG_CHQ_FEEAMOUNT_ERR)
ค่าเริ่มต้นรูปแบบลงบัญชี: `GetDefaultGLFormat(16)`; ฟอร์มอื่นใช้โครงตรวจเดียวกัน (ยังไม่ verify ทีละข้อของฟอร์มอื่น)

ลำดับตอนบันทึก (ตัวอย่าง 5401 `:1255-1330`): (1) ถ้าแก้เอกสารเดิม: ย้อนผลของแถวเดิม (`nDeleteMode=1`) แล้วลบแถวย่อย (2) แทรกแถวย่อยใหม่ (3) ลงบัญชี (อัตโนมัติ `ProcessConditionGroup11` หรือแบบลงมือ `OnSaveAccount`) (4) `ProcessChqStatus(nDeleteMode=0)` ลงผลของแถวใหม่ (5) บันทึก log (`SaveLog` รหัสเมนู `ID_CHQSYS_TRANS_2`)
ลำดับตอนลบ (`:492-516`): ลบหัว -> ย้อนผลทุกแถว + ลบแถวย่อย -> `DeleteAccount(DocNo, source)` ลบรายการบัญชี -> log

## 5. เอกสารเปลี่ยนสถานะ — รายตัว

| id | เอกสาร | ตัวกรองเช็คที่เลือกได้ | Status ใหม่ | ลงบัญชี (default format / GL source / group) | หมายเหตุเฉพาะ |
|---|---|---|---|---|---|
| 5398 | นำฝากเช็ครับ | Status=0 และ ReceiveDate <= วันที่เอกสาร | 1 | ไม่ลงบัญชี | ใช้ `UpdataCheqStatus` ของตัวเอง (ตั้ง Status/StatusDocNo/StatusDate/BookNo ไม่แตะ Balance) |
| 5401 | เช็ครับผ่าน | (Status 0 หรือ 1) และ DueDate <= วันที่เอกสาร (+ BookNo ถ้าระบุ) | 2 | 16 / 11 / group 11 | ค่าธรรมเนียมหักจากยอดรับ (Net = Amount - Fee) |
| 5408 | เช็ครับคืน | (Status 2 หรือ 6) และ ReceiveDate <= วันที่เอกสาร | 3 | 17 / 24 / group 14 | |
| 5399 | นำเช็คเข้าใหม่ | Status=3 | 6 | 19 / 32 / group 28 | |
| 5402 | ยกเลิกเช็ครับ | ตามประเภท (ดู 2.3) | 4 | 18 / 25 / group 15 | CancelStatus 0 ไม่ต้องมีสมุดเงินฝาก |
| 5403 | ขายลดเช็ครับ | (Status 0 หรือ 1) และ ReceiveDate <= วันที่เอกสาร | 5 | 27 / 29 / group 17 | ลบเอกสาร (ย้อน) ส่งสถานะ 2 (`cqfrmcheqinsale.cpp:452`) แต่บันทึกส่ง 5 (`:1232`) — ความหมายของ call ลบยังไม่ verify |
| 5404 | เช็คจ่ายผ่าน | Status=0 และ DueDate <= วันที่เอกสาร (+ BookNo ถ้าระบุ) | 1 | 20 / 12 / group 18 | ท้ายตารางเอกสารพิมพ์รายงาน "ขึ้นหน้าใหม่ตามเอกสาร" ได้ |
| 5405 | ยกเลิกเช็คจ่าย | ติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว" = Status=1 (ต้องระบุ BookNo), ไม่ติ๊ก = Status=0 | 2 | 22 / 27 / group 19 | ถ้า CancelStatus=0 จะไม่ลงบัญชีเฉพาะโปรเจกต์ CHM (`#ifdef CHM_PROJ` `:1176-1186`) |
| 5406 | ขึ้นเงินบัตรเครดิต | Status=0 และยังไม่อยู่ใน BCCreditPassSub | 1 (บัตร) | 23 / 13 | มีภาษี (`OnSaveInputTax`) |
| 5407 | ยกเลิกบัตรเครดิต | Status=0 | 2 (บัตร) | 28 / 28 | |

(เลข default GL format = `GetDefaultGLFormat(n)`, GL source = `nGLTransSource`; ที่มา: `cqfrmcheqinpass.cpp:505`, `cqfrmcheqincancel.cpp:464`, ฟอร์มอื่นไฟล์ของตัวเอง; กลุ่ม `ProcessConditionGroupNN` อยู่ `LIB:gltlsprocess.cpp:7637-8026`)

### 5.1 การลงบัญชี GL ของเอกสารเช็ค

- เป็น "แม่แบบ" (GLFormat) ที่ผู้ใช้กำหนดเอง ไม่ฝังบัญชีในโค้ด — เงื่อนไขของบรรทัดแม่แบบที่เกี่ยวกับเช็ค: เงื่อนไข 1 = จำนวนเงินรวมของเช็ค (`SumChqAmount`; ฟอร์มบัตรยกเลิกใช้ `SumCreditAmount`), เงื่อนไข 2 = ค่าใช้จ่ายรวม (`SumExpend`), เงื่อนไข 3 = ยอดสุทธิ (`NetAmount`) ผูกกับบัญชีของสมุดเงินฝาก (`BCBankBook.AccountCode` ผ่าน `GetBookAccountCode`), เงื่อนไข 90/91 = ปัดเศษ (`LIB:gltlsprocess.cpp:7637-8026`) — เนื้อหาแม่แบบเริ่มต้น (บัญชี เดบิต/เครดิต) อยู่ใน GLFormat ฐานข้อมูล ยังไม่ verify
- โหมดลงบัญชี: อัตโนมัติ (`GLTransData=0` เรียก `ProcessConditionGroupNN`) หรือลงมือ (`m_nManualKey`, `OnSaveAccount`)
- เอกสารนำฝาก 5398 ไม่ลงบัญชี
- ตาราง GL เฉพาะเช็ค: `BCGLChq` (BookCode, DocDate, DocNo, Source, BankCode, Chqno, Amount, Duedate, BookNo, ChqType, ExChangeRate, BranchID) — ใช้เป็นข้อมูลเช็คใต้ GL (วิธีเขียน ยังไม่ verify)
- GL ของ BC เป็นสมุดรายวันครั้งเดียว: การผูกกับเช็คต้องรอลุงจืดตัดสิน (ดูข้อ 9)

## 6. กระบวนการ: คำนวณยอดคงเหลือของเช็คใหม่ (5400)

ไดอะล็อก `CMyReCalcCheqe` (หัวเรื่อง "คำนวณยอดคงเหลือของเช็คใหม่"): ตัวเลือก เช็ครับ / เช็คจ่าย (radio), จากเช็คเลขที่ / ถึงเช็คเลขที่, เริ่มจากรหัสลูกค้า / ถึงรหัสลูกค้า, ปุ่ม Start / Cancel, แถบความคืบหน้า. ใช้ปรับยอดสะสม `BCAR.ChqOnHand/ChqReturn` (และ `BCAP` สำหรับเช็คจ่าย) ให้ตรงกับสถานะเช็ครับ/จ่ายจริง (`LIB:ChqProcess.cpp`; รายละเอียดสูตร ยังไม่อ่านครบทุกบรรทัด — ยังไม่ verify)
BC: เมนู "คำนวณยอดคงเหลือของเช็คใหม่" `/tools/cheque-balances` มีอยู่แล้ว (`erp-tools.ts:72-84`) — แต่ถ้า BC คำนวณยอดจากตารางเช็คตอนอ่านจะไม่ต้องมี process นี้ ให้ลุงจืดตัดสินว่าจะคงปุ่มไว้เป็น "ตรวจสอบ/ตรวจนับ" หรือเอาออก

## 7. รายงาน (24 รายการ)

### 7.1 ไดอะล็อกเงื่อนไขมาตรฐาน

กล่องกลุ่ม "เงื่อนไขการพิมพ์รายงานมาตรฐาน" (ข้อความไทยจาก rc `BCRepCashChq.rc` dialog 35001-35051). ช่องกรอง (key -> ป้ายไทยที่ยืนยัน):

| key | ป้ายไทย |
|---|---|
| FROM_/TO_RECEIVEDATE | พิมพ์จาก/ถึงวันที่รับเช็ค: |
| FROM_/TO_DUEDATE | พิมพ์จาก/ถึงวันที่ครบกำหนด (ข้อความใน rc บางรายงานยังเป็น "วันที่รับเช็ค" ตามที่ rc พิมพ์ผิด) |
| FROM_/TO_DOCNO | พิมพ์จาก/ถึงเลขที่เอกสาร |
| FROM_/TO_ARCODE | พิมพ์จาก/ถึงรหัสลูกหนี้ |
| FROM_/TO_APCODE | พิมพ์จาก/ถึงรหัสเจ้าหนี้ (ยังไม่ verify ป้ายตรงตัว) |
| FROM_/TO_CHQNO | พิมพ์จาก/ถึงเลขที่เช็ค |
| FROM_/TO_DEPART | พิมพ์จาก/ถึงรหัสแผนก |
| FROM_/TO_BKBOOK | จาก/ถึงสมุดเงินฝาก (เช็คจ่ายตามสมุดบัญชี) |
| FROM_/TO_PAYMENTDATE | จาก/ถึงวันที่จ่าย |
| CHOSE_STATUSCHECK | เลือกสถานะเช็ครับ (combo: ทั้งหมด + 7 สถานะ) |
| REPORT_BY_ACC | แสดงรายงานตามเลขที่บัญชี: (checkbox) |

### 7.2 รายงานรายละเอียดเช็ครับ (5380-5384)

เมนู 5380 เปิด `CCHMCQRepChqInView` (`MainFrm.cpp:8203-8232`, dialog `IDD_CHMCHQIN_REP` caption "Chq in report"): combo จัดเรียงตาม = [วันที่รับเช็ค, วันที่ครบกำหนด, ลูกค้า, เลขที่เช็ครับ, แผนก], combo สถานะ (ทั้งหมด + 7), ช่วง: ลูกหนี้ / วันครบกำหนด / เลขที่เช็ค / แผนก / เลขที่เอกสาร / วันที่รับเช็ค, และ checkbox ตามบัญชี. 5380 เรียก `REPComm_InsertCommand(0/1, ID_CHQSYS_MASTER_1)` เพิ่มคอลัมน์เสริมจากฟอร์ม master (กลไกคอลัมน์ผู้ใช้กำหนด ยังไม่ verify รายละเอียด) และตัวเปิดรายงานแบบเรียงตามวันที่รับเดิม `CCQRepChqInByReveiveDateView` ถูก comment ทิ้งแล้ว (`MainFrm.cpp:8212-8216`). เมนู 5381-5384 เปิดคลาสรายงานแยกต่างหาก: ตามวันครบกำหนด `CCQRepChqInByDueDateView` (`MainFrm.cpp:8227`), ตามลูกค้า `CCQRepChqInByArView` (`:8238`), ตามเลขที่ `CCQRepChqInByChqNoView`, ตามแผนก `CCQRepChqInByDepartView` — เป็น 4 หน้าจอเงื่อนไข/รายงานของตัวเอง (เงื่อนไขของแต่ละคลาส ยังไม่ verify ทีละตัว) ส่วน combo "จัดเรียงตาม" ของ 5380 ครอบคลุมคีย์เดียวกัน

คอลัมน์รายการ (โหมดไม่ติ๊กตามบัญชี): วันที่รับเช็ค (`CHEQ_RECE_DATE`), (คอลัมน์ซ่อน), เลขที่เอกสาร (`DOCNO`), วันที่ครบกำหนด (`DUE_DATE`), `ARCODE`, `ARNAME`, เลขที่เช็ค (`CHQNUMBER`), (ซ่อน), ธนาคาร/สาขา (`BANK/BRANCH`), สถานะ (`STATUS`), จำนวนเงิน (`AMOUNT`, ชิดขวา), หมายเหตุ/รายละเอียด (`COMMENT&DESCRIPT`) — ลำดับ "คอลัมน์นำหน้า" สลับตามจัดเรียง (ครบกำหนด/ลูกหนี้/เลขที่เช็ค/แผนก)
ยอดรวม: รวมกลุ่ม "รวม" (`ITEM_SUM`) พร้อม "N รายการ", รวมทั้งสิ้น `SUMTOTALDETAIL_WORD n TRANSAC`; ติ๊กตามบัญชี -> จัดกลุ่มตามสมุดเงินฝาก (หัวกลุ่ม `PAY_FROM_ACC` + ชื่อสมุดเงินฝาก) และมีรวมต่อสมุด
ชื่อรายงาน (key): 64101-64108. รายงานที่มี status checkbox แยก: dialog 35032 (เช็คในมือ, เช็คฝากธนาคาร, เช็คผ่าน, เช็คคืน, เช็คยกเลิก, ขายลดเช็ค, เช็คคืนฝากธนาคาร, "แสดงเรียงตามลูกค้า", ช่วงรหัสพนักงานขาย)

### 7.3 รายงานรายละเอียดเช็คจ่าย (5386-5391)

เมนู 5386 เปิด `CCHMCQRepChqOutView` (`IDD_CHMCHQOUT_REP`, `MainFrm.cpp:8284-8288`) ในแนวเดียวกับ 5380; 5387-5391 เปิดคลาสแยก: วันครบกำหนด `CCQRepChqOutByDueDateView`, เจ้าหนี้ `CCQRepChqOutByApView`, เลขที่ `CCQRepChqOutByChqNoView`, แผนก `CCQRepChqOutByDepartView`, สมุดบัญชี `CCQRepChqOutByBankBookView` (`:8298,8342`)
คอลัมน์: วันที่จ่าย (`CHEQ_PAY_DATE`), เลขที่เอกสาร, วันครบกำหนด, ชื่อเจ้าหนี้ (`APNAME`), เลขที่เช็คจ่าย (`CHEQ_PAY_NO`), ธนาคาร/สาขา, สถานะ, จำนวนเงิน, หมายเหตุ/รายละเอียด — คอลัมน์นำหน้าสลับตามจัดเรียง (`APNAME` / `CHEQ_PAY_NO` / `DEPART_NAME` / `BOOK_NO`)
สถานะ checkbox (dialog 35045): เช็คในมือ, เช็คผ่าน, เช็คยกเลิก; key ชื่อรายงาน 64201-64206

### 7.4 รายงานบัตรเครดิต (5393-5397)

5393 เปิด `CCHMCQRepCreditcardView` (`IDD_CHMCREDITCARD_MASTER`, `MainFrm.cpp:8365`); 5394-5397 เปิดคลาสแยก `CCQRepCardByDueDateView`, `CCQRepCardByArView`, `CCQRepCardByCardNoView`, `CCQRepCardByDepartView`. คอลัมน์ใช้ `CREDITCARD_RECE_DATE`, `CREDITCARD_NO`; key ชื่อรายงาน 64301-64305

### 7.5 รายงานรายวัน (5409-5416) — ใช้ไดอะล็อกร่วม

ฟิลเตอร์: จาก/ถึงวันที่ (key `FROM_DOCDATE`; ป้ายใน rc เขียนว่า "ถึงกำหนด" — ผิดตามต้นฉบับ), จาก/ถึงเลขที่เช็ครับ (ป้ายใน rc = "เลขที่เช็ครับ" แต่ค่าจริงคือเลขที่เอกสาร — ยังไม่ verify), จาก/ถึงรหัสสมุดเงินฝาก, ตัวเลือก "ขึ้นหน้าใหม่ตามเอกสาร" (เฉพาะ 5410 และ 5412 — เช็ครับผ่านและเช็คจ่ายผ่าน)
เลย์เอาต์: แถวหัวเอกสาร: เลขที่เอกสาร (`DOCNO`), วันที่เอกสาร (`DOCDATE`), สมุดเงินฝาก (`BOOK_CODE`), จำนวนเงิน (`SUM_AMOUNT`), ค่าใช้จ่าย (`SUM_FEE`), สุทธิ (`SUM_NETT`); แถวรายการ: `CHEQ_NO`, `BANK` (เช็คจ่ายแสดง `APNAME` + `DEBT_PAYMENT`), `SETTLE_DATE`, `CHQ_AMOUNT`, `AMOUNT`, `FEE_AMOUNT` (บัตร: `CREDITCARD_NO ... CREDIT_AMOUNT`); key ชื่อรายงาน 65001-65008

| id | ชื่อรายงาน | ตารางต้นทาง |
|---|---|---|
| 5409 | ใบนำฝากเช็ครับพร้อมรายการย่อย | BCChqInDeposit/DeposSub |
| 5410 | รายวันเช็ครับผ่านพร้อมรายการย่อย | BCChqInPass/PassSub |
| 5411 | การยกเลิกเช็ครับพร้อมรายการย่อย | BCChqInCancel/CancelSub |
| 5416 | การขายลดเช็ครับพร้อมรายการย่อย | BCChqInSale/SaleSub |
| 5412 | รายวันเช็คจ่ายผ่านพร้อมรายการย่อย | BCChqOutPass/PassSub |
| 5413 | การยกเลิกเช็คจ่ายพร้อมรายการย่อย | BCChqOutCancel/CancelSub |
| 5414 | รายวันขึ้นเงินพร้อมรายการย่อย | BCCreditPass/PassSub |
| 5415 | การยกเลิกบัตรพร้อมรายการย่อย | BCCreditCancel/CancelSub |

## 8. BC เทียบ Champ และสิ่งที่ต้องปรับ

สถานะ BC (ตรวจ 2026-10-07): เมนู BC กลุ่ม `cash-bank` อยู่ที่ `frontend/src/lib/menu-data.ts:376-460` ชื่อเมนูใช้คำ Champ แล้ว; จอธุรกรรมทุกจอ (`/transaction/cheque*`, `/banking/cheques/*`, `/banking/cards/*`) อยู่ใน `isErpTransactionRoute` (`erp-transaction.ts:388-805`) จึงแสดง "รอพัฒนา" (`menu-screen-status.ts:46-50` isMenuBackendRetired); API เดิม (`apiPath: "chequereceive/chequedeposit"` ฯลฯ) อยู่บน MongoDB ที่ถูกถอดแล้ว (AGENTS.md 2026-09-23)
BC ตั้งจอเช็คเป็น "เอกสารทั่วไป" (counterparty + prefix + `hasLineItems:false`, `erp-transaction.ts:388-410`) — ไม่มีกริดเลือกเช็ค, ไม่มีสถานะ, ไม่มีผลต่อสมุดเงินฝาก/ลูกหนี้ — ต้องทำใหม่ตามโครงใน 2-5

### 8.1 ตารางเทียบรายเมนู

| Champ id | ชื่อ Champ | BC menu id / route | สถานะ | สิ่งที่ต้องปรับ |
|---|---|---|---|---|
| 5377 | บันทึกเช็ครับ | `cheque-received` `/transaction/chequereceived` | ต่าง + รอพัฒนา | ทำฟอร์มตามข้อ 3.1: 16 ฟิลด์ตามลำดับ, สถานะ 7 ค่า, ตรวจเลขเช็คซ้ำ (ChqNumber+ธนาคาร+สาขา), คำนวณเงินบาท = อัตรา x จำนวนเงิน, สถานะเพิ่มเติม, ห้ามแก้เมื่อถูกอ้างอิง |
| 5378 | บันทึกเช็คจ่าย | `cheque-issued` `/transaction/chequeissued` | ต่าง + รอพัฒนา | ข้อ 3.2: รหัสเจ้าหนี้, วันที่จ่าย, สมุดเงินฝากบังคับ, ตรวจวันที่จ่าย <= ครบกำหนด, พิมพ์เช็ค |
| 5379 | บันทึกบัตรเครดิต | `credit-card-receipts` `/banking/cards/receipts` | ต่าง + รอพัฒนา | ข้อ 3.3 (เลขบัตร, CreditType, ChargeAmount, VAT; ฟิลด์ EDC เป็นของลูกค้าเฉพาะราย ไม่ต้องเริ่มทำ) |
| 5398 | บันทึกนำฝากเช็ครับ | `cheque-deposit` `/banking/cheques/deposit` | ต่าง + รอพัฒนา | กริดเลือกเช็ค Status=0; ตั้ง Status=1; ไม่ลงบัญชี ไม่แตะยอดสมุดเงินฝาก. ปัจจุบัน BC ใช้ `apiPath` เดียวกับบันทึกเช็ครับ (`erp-transaction.ts:695-705`) ต้องแยกเอกสาร |
| 5401 | บันทึกเช็ครับผ่าน | `cheque-received-clear` `/banking/cheques/received-clear` | ต่าง + รอพัฒนา | กริด 7 คอลัมน์ (ข้อ 4), ตัวกรอง Status 0/1 และ DueDate <= วันที่เอกสาร, Net = Amount - Fee, ลงบัญชี |
| 5408 | บันทึกเช็ครับคืน | `cheque-received-return` | ต่าง + รอพัฒนา | ตัวกรอง Status 2/6; Status -> 3 |
| 5399 | บันทึกนำเช็คเข้าใหม่ | `cheque-redeposit` | ต่าง + รอพัฒนา | ตัวกรอง Status=3; Status -> 6 |
| 5402 | บันทึกยกเลิกเช็ครับ | `cheque-received-cancel` | ต่าง + รอพัฒนา | combo 4 ประเภท (ข้อ 2.3) |
| 5403 | บันทึกขายลดเช็ครับ | `cheque-discount` | ต่าง + รอพัฒนา | ตัวกรอง Status 0/1; Status -> 5 |
| 5404 | บันทึกเช็คจ่ายผ่าน | `cheque-issued-clear` | ต่าง + รอพัฒนา | ตัวกรอง Status=0, DueDate <= วันที่เอกสาร; Status -> 1; พิมพ์ "ขึ้นหน้าใหม่ตามเอกสาร" |
| 5405 | บันทึกยกเลิกเช็คจ่าย | `cheque-issued-cancel` | ต่าง + รอพัฒนา | ช่องติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว" (Status=1 ต้องมีสมุดเงินฝาก / ไม่ติ๊ก Status=0) |
| 5406 | บันทึกขึ้นเงินบัตรเครดิต | `credit-card-settle` | ต่าง + รอพัฒนา | มีภาษี (อัตรา/ก่อนภาษี/ภาษี) |
| 5407 | บันทึกยกเลิกบัตรเครดิต | `credit-card-cancel` | ต่าง + รอพัฒนา | ตัวกรอง Status=0 |
| 5400 | คำนวณยอดคงเหลือของเช็คใหม่ | `cheque-recalculate` `/tools/cheque-balances` | ต่าง (ชื่อเหมือน; เนื้อหาไม่ตรง) | ฟอร์ม Champ มี radio เช็ครับ/จ่าย, ช่วงเลขที่เช็ค, ช่วงลูกค้า, แถบความคืบหน้า; BC เป็นเครื่องมือทั่วไป `erp-tools.ts:72-84` ตรวจ `isErpToolApiReady` (ยังไม่ verify) |
| 5380-5384 | รายงานรายละเอียดเช็ครับ 5 แบบ | `cheque-received-report` `/report/chequereceived` (รายการเดียว) | ต่าง | ทำรายงานเดียวที่มี combo "จัดเรียงตาม" 5 ค่า + สถานะ 7 ค่า + ช่วงตามข้อ 7.2 (เหมือน 5380) แล้วให้ 5 เมนูชี้ค่า sort ล่วงหน้า — เป็นการยุบคลาสแยก 5381-5384 ของ Champ (ผลเทียบเท่า ยังไม่ verify ว่าคอลัมน์ทุกตัวตรงกัน; ยืนยันกับลุงจืดว่ายุบได้) |
| 5386-5391 | รายงานรายละเอียดเช็คจ่าย 6 แบบ | `cheque-issued-report` `/report/chequeissued` | ต่าง | เช่นเดียวกัน 6 ค่าจัดเรียง (รวมตามสมุดบัญชี) |
| 5393-5397 | รายงานบัตรเครดิต 5 แบบ | `credit-card-report` `/report/creditcard` | ต่าง | เช่นเดียวกัน 5 ค่าจัดเรียง |
| 5409-5416 | รายงานรายวัน 8 รายการ | (ไม่มี) | ยังไม่มี | เพิ่มได้ภายหลังเมื่อมีเอกสารเปลี่ยนสถานะ (เป็นรายงานอ่านอย่างเดียวจากตารางหัว+ย่อย) |

หมายเหตุสถานะรายงาน: รหัสรายงานที่พร้อมใช้งานจริงใน BC มี 12 รายการ (`erp-reports.ts:530-543` — ลูกหนี้/เจ้าหนี้/ขาย/สต๊อก) ไม่มีรายงานเช็ค/เงินสด/ธนาคาร จึงทุกรายการแสดง "ยังไม่มีข้อมูล" (`menu-screen-status.ts:isMenuDataPending`)

### 8.2 ความต่างที่สำคัญที่สุด

1. BC ยังไม่มี "สถานะเช็ค" และเอกสารเปลี่ยนสถานะเลย (เป็นเอกสารทั่วไป) — Champ คือเครื่องสถานะ 7 ค่า (ข้อ 2) ที่ทุกเอกสารเลือกเช็คจากกริดตามตัวกรองสถานะ
2. ตัวกรองเลือกเช็คแต่ละเอกสารต้องตรงตามข้อ 2/5 (เช่น เช็ครับผ่านรับ Status 0 หรือ 1 และถึงกำหนดแล้ว) — เป็นกติกาธุรกิจหลัก ไม่ใช่รายละเอียด UI
3. ผลต่อสมุดเงินฝากและยอดเช็คของลูกหนี้/เจ้าหนี้ — Champ ใช้ยอดสะสมบวก/ลบทีละเอกสารและมีโอกาสเพี้ยน (ข้อ 2.5); BC ควรคำนวณยอดจากรายการเช็ค/เอกสารจริงเสมอ (derived) และเก็บ `StatusDocNo/StatusDate` ของเช็คเท่านั้น
4. ย้อนเอกสาร (ลบ/แก้) ต้องคำนวณสถานะเช็คใหม่จากเอกสารที่เหลือ (UNION ตามวันที่) ไม่ใช่ย้อนกลับเป็นค่าตายตัว (`cqfrmbasetransstatus.cpp:264-300`)
5. รายงานรายละเอียด 16 เมนู (เช็ครับ 5 / เช็คจ่าย 6 / บัตร 5) ของ Champ เป็นเนื้อหาชุดเดียวกัน 3 ชุดเรียงตามคีย์ต่างกัน (ตัวแรกของแต่ละชุดเป็นหน้าจอ CHM ที่มี combo "จัดเรียงตาม") — BC ทำ 3 รายงานที่มี combo จัดเรียง แล้วให้เมนูย่อยชี้ค่า sort ได้ (ผลลัพธ์เทียบเท่า ลดจอซ้ำซ้อน สอดคล้องกฎยุบรวมใน AGENTS.md)
6. เมนูรายงานของ Champ ซ่อนในศูนย์รายงาน (popup `disable="yes"`) ไม่แสดงในเมนูหลัก

## 9. ข้อขัดกับกฎเจ้าของ/ขอบเขต (ไม่เสนอทำเอง — รอลุงจืดตัดสิน)

- ขอบเขต GL ล่าสุดของลุงจืด: ระบบห้องบัญชี เน้นรับข้อมูล/ตรวจสอบ/กระทบยอด/ปิดงบ; `mydocs/specs/gl/ar-ap-bank-design-notes.md:49` ระบุ "workflow สั่งจ่าย/คืนเงิน/เช็คเป็นหน้าที่ระบบต้นทาง" และ `AGENTS.md` ว่า AR/AP/Statement เป็นรายละเอียดตรวจสอบ ไม่ใช่คำสั่งรับจ่ายเงินจริง -> การสร้างวงจรสถานะเช็คครบทั้ง 13 เอกสารขัดกับขอบเขตข้อนี้ จึงต้องให้ลุงจืดตัดสินก่อนเริ่ม: (ก) ทำตาม Champ ครบ หรือ (ข) ให้ GL รับ "ผลของเช็ค" เป็นข้อมูลต้นทางผ่านใบสำคัญรายวันแทน
- เงื่อนไข "ห้ามเดา" ตามกฎ: เนื้อหา GLFormat เริ่มต้น (บัญชีเดบิต/เครดิตต่อเอกสาร) ยังไม่มีใน repo — ห้ามกำหนดบัญชีเองในโค้ด ต้องให้ผู้ใช้กำหนดผ่านรูปแบบการลงบัญชี
- ห้ามเขียนเงื่อนไขตามรหัสบัญชี (memory `no-account-code-conditions`); ใช้ `AccountCode` จากสมุดเงินฝากหรือค่าที่บันทึกไว้
- เงิน/จำนวนต้องเป็น decimal ไม่ใช่ float (กฎ 20) — Champ ใช้ MONEY/DOUBLE ปนกัน (`dNetAmount*nMark`)
- ฟิลด์เฉพาะลูกค้า Champ (`IsSAP`/`ISSAP`, `CHM_PROJ`, EDC) ไม่นำมา (YAGNI) เว้นลุงจืดสั่ง

## 10. สิ่งที่ยังไม่ verify (สรุป)

| เรื่อง | ทำไมยัง |
|---|---|
| ข้อความไทยของ resource key (ป้ายหัวคอลัมน์ ARNAME/BANK/BRANCH/STATUS/SUM_FEE/PAY_FROM_ACC และ combo สถานะบัตร) | ฐาน resource ไม่อยู่ใน repo |
| ความหมายทุกค่าของ `SaveFrom` | อ่านเฉพาะค่า 4 |
| การใช้งานเช็คจ่ายสถานะ 3 "เตรียมจ่าย" | ปุ่มถูก comment |
| ผลกระทบยอดสมุดเงินฝากของ 5402 CancelStatus 2 และ 5405 CancelStatus 0 เมื่อระบุ BookNo | อ่านโค้ด ไม่ได้รันกับข้อมูลจริง |
| เนื้อหาแม่แบบ GLFormat เริ่มต้น | อยู่ในฐานข้อมูล |
| สูตรเต็มของ `CMyReCalcCheqe` | ยังไม่อ่านทุกบรรทัด |
| การตรวจก่อนบันทึกของฟอร์มเปลี่ยนสถานะอื่นนอกจาก 5401 | อ่านละเอียดเฉพาะ 5401 |
| เมนูดูแล master `BCChqStatus` และกลุ่มสร้างเช็คอัตโนมัติ `BCChqOutAuto` | ไม่อยู่ในเมนู 38 รายการ |
