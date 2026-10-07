# Champ โมดูล 7 "IC" (ระบบสินค้าคงคลัง) — พิมพ์เขียวสำหรับทำให้ BC เหมือน Champ

> สถานะ: ร่างเพื่อ review โดย Opus (เขียนโดย agent ลูกรุ่น Sonnet 5.5, 2026-10-07). Champ = `D:\project-champ` (C++/MFC + SQL Server). BC = `D:\bccode`.
> ขอบเขตไฟล์นี้: 68 รายการเมนู IC (รวมสินค้าชุด + รายงานสินค้า) ตาม `menuconfig.xml:407-511`. **ไม่แก้โค้ด BC** — เป็นสเปกสำหรับสร้างทีหลัง (BC ฝั่ง IC ยังเป็น "รอพัฒนา" เกือบทั้งหมด และลุงจืดสั่ง GL-only ก่อน).

## 0. วิธีอ่านและข้อจำกัดของหลักฐาน

1. **อ้างอิง**: `ไฟล์:บรรทัด` ของ Champ อยู่ใต้ `D:\project-champ\champ\champ\` เว้นแต่ระบุ `champ-lib/...` หรือ `champ-report/...`. `atlas-07:N` = `D:\bccode\scratch\champ-atlas\atlas-07-IC.md` บรรทัด N. BC อ้างเป็น `frontend/src/...` / `backend/internal/...`.
2. **ข้อความภาษาไทยบนจอ Champ ไม่ได้อยู่ในโค้ด** — ฟอร์มเรียก `theApp.GetMyRes("KEY")` ซึ่งอ่านจากไฟล์ทรัพยากรภายนอก `BCResData.mdb` (`champ-lib/BC5FRMWRK/BCSWinApp.cpp:80,614`; ไม่อยู่ใน repo). จึงกู้คำไทยได้เฉพาะจาก (ก) ข้อความ `KEY$ไทย` ใน `BC5Account.rc` (ข) คอมเมนต์ไทยข้างโค้ด `//("สินค้าทั่วไป")` (ค) ชื่อเมนู `menuconfig.xml`. ในเอกสารนี้:
   - `«KEY»` = key ทรัพยากร ที่กู้ไทยไม่ได้ (ต้องไปถามลุงจืด/ดูจากโปรแกรมจริง) — **ห้ามเดาคำ**. ถ้ามีคำไทยจาก (ก)-(ค) จะเขียนต่อท้าย `= ...`.
   - คำที่ใส่ `(~)` = อนุมานจากชื่อ key/ตำแหน่ง ยังไม่ verify กับหน้าจอจริง.
3. **ชื่อฟิลด์ฐานข้อมูล** ใช้ตามสคริปต์ `champ/Script/SQLSERVER_Script.sql` (ผ่าน `atlas.json` → tables). BC ต้องเป็น PostgreSQL lowercase `snake_case` (กฎ repo) — ตารางเทียบชื่ออยู่ท้ายแต่ละหัวข้อเป็นแนวทาง ไม่ใช่ข้อบังคับ.
4. **รายงานทั้งกลุ่มในเมนู IC ถูกซ่อนจากเมนูหลัก** (`disable="yes"` ที่ระดับ popup — `menuconfig.xml:440,446,482`) และ `XMLMenuConfig.cpp:124-179` ข้ามรายการที่ disable จากการสร้างเมนู. รายงานเปิดผ่าน **ศูนย์รายงาน (Report Center)** ที่โหลดโครงต้นไม้จากไฟล์ `reportcenterconfig.xml` ข้างตัว exe (`reportCenterView.cpp:226-231`) — **ไฟล์นี้ไม่อยู่ใน repo ⇒ ผังต้นไม้ของศูนย์รายงานยังไม่ verify**; แต่คำสั่งเปิดรายงานใช้ command id เดียวกับเมนู (`reportCenterView.cpp:158` ส่ง `WM_COMMAND` ด้วย id เมนู) จึงถือ id/ชื่อใน `menuconfig.xml` เป็นรายการจริง.
5. **ค่านับ**: IC มี 68 รายการเมนู (นับจาก atlas `grep -c "^- \*\*"` = 68; 10 รายการเอกสาร/ข้อมูลหลัก, 6 สินค้าชุด+Serial+ตรวจรับ+ลำดับ+คำนวณใหม่ ฯลฯ, ที่เหลือเป็นรายงาน).

### 0.1 กลไกร่วมของเอกสารสต็อก (ใช้ซ้ำทุกเอกสารด้านล่าง)

| หัวข้อ | พฤติกรรมใน Champ | หลักฐาน |
|---|---|---|
| ฟอร์มเอกสาร | ทุกเอกสารสืบจาก `CMyEspeciallyFormView` — หัวเอกสาร (เลขที่/วันที่/แผนก/…) + กริดรายการ `BInputListWnd` + คอลัมน์เงา (shadow) เก็บค่าที่ไม่แสดง | `icfrmstockissue.cpp:140-236` |
| เลขที่เอกสาร | `ED_DOCNO` + ปุ่มค้นหา, ตรวจซ้ำก่อนบันทึก ("Document No is duplicated … try again?" Yes/No) | `icfrmstockissue.cpp:729-737` |
| การบันทึก | `ConfirmSaveData()` ตรวจ → `OnSaveData()` เขียนหัว (`InsertDB`/`UpdateDB`) → ลบรายการย่อยเก่าที่ `TransState=1` แล้วเขียนรายการย่อยใหม่ → `UPDATE IsCompleteSave=1` → `Commit` → `INSERT INTO processdoc(DocNo,DocType)` → เรียก process ตัดสต็อก → ลงบัญชี GL → บันทึก log book | `icfrmstockissue.cpp:845-1131` |
| การตัดสต็อกไม่ได้ทำในฟอร์ม | ฟอร์มเรียก `m_BCProcess.StkItemIssueProcess(docNo)` (ยกเว้นเมื่อ `AutoServerProc=1` ซึ่งให้ Robot/เซิร์ฟเวอร์ทำแทน) — ตัวประมวลผลจริงอยู่ที่ `champ-lib/BC5PROCESS/bcdumpprocjob.cpp:2473 StockCenterProcess` | `icfrmstockissue.cpp:1092-1093` |
| แก้ไขเอกสารเดิม | บันทึกแถวใหม่เป็น `TransState=0`, แถวเดิมตั้ง `TransState=1` = "รายการเก่า" → process ย้อน (กลับรายการ, mark −1) ก่อนแล้วลบ แล้วค่อยตัดแถวใหม่ | `bcdumpprocjob.cpp:4511-4529` (issue); `bcdumpprocjob.cpp:2577` ("0 = new transaction 1 = old transaction") |
| ลบเอกสาร/ยกเลิก | `m_nDeleteMode`: 1 = ลบ (ย้อนแถว `TransState=0` แล้ว `DELETE ...Sub`), 2 = ยกเลิก (ย้อนแถว `IsCancel=1`) | `bcdumpprocjob.cpp:4483-4508` |
| ยืนยัน / ยกเลิกเอกสาร | ปุ่ม/ปุ่มลัด Shift+F5 = ยืนยัน (`ID_HK_CONFIRMDOC`), Ctrl+F5 = ยกเลิก (`ID_HK_CANCELDOC`); เอกสารที่ยืนยันแล้วหรือยกเลิกแล้ว **บันทึกซ้ำไม่ได้** (`CANNOT_SAVE_DOCUMENTCONFIRM`, `NOT_SAVE_ISCANCEL`) และเอกสารที่ผ่านบัญชีแล้วแก้ไม่ได้ (`CANNOT_SAVE_POSTGL`) | `BC5Account.rc:982-983`; `icfrmstockissue.cpp:537-550,2771-2781` |
| ปิดงวด | บันทึกไม่ได้ถ้าวันที่อยู่ในงวดที่ปิดแล้ว | `icfrmstockissue.cpp:531` (`IsPeriodClose`) |
| ลงบัญชี GL | เลือกรูปแบบการลงบัญชี (`GLFormat`, ค่าเริ่มต้นตามกลุ่ม `GetDefaultGLFormat(group)`); ถ้า `GLTransData(m_nManualKey)=0` และมี GLFormat → ลงอัตโนมัติผ่าน `m_GLTLSProcess`, ถ้าไม่ → เปิดหน้าจอ F2 ให้ลงมือ (`OnSaveAccount`) | `icfrmstockissue.cpp:1096-1106` |
| ที่เก็บ/คลัง | แต่ละบรรทัดต้องมี คลัง + ที่เก็บ + หน่วย + จำนวน (ไม่งั้น `CANNOT_SAVE_NEEDDETAIL`/`PUT_NOT_QTY`) | `icfrmstockissue.cpp:621-640` |
| Serial / Lot | ปุ่มลัด Ctrl+F9 = Serial, Shift+F9 = Lot, Ctrl+F2 = ตรวจสต็อก, Shift+F2 = ขยายแถว, F2 = ดู/ลง GL | `BC5Account.rc:977-1011`; `icfrmstockissue.cpp:1286-1330` |
| ตัวเลขทศนิยม | จำนวน = `CompanyInfo.QtyPoint`, ต้นทุน = `CostPoint`, ราคา = `PricePoint`, จำนวนเงิน = `AmountPoint` (ตั้งค่าบริษัท, ไม่ตายตัว) | `icfrmstockissue.cpp:187-198`; `bcdumpprocjob.cpp:2617-2763` |
| รหัสชนิดรายการสต็อก (`MyType`) | 0 ยกยอดมา · 1 รับสินค้าสำเร็จรูป · 2 ซื้อ · 3 เพิ่มหนี้เจ้าหนี้ · 4 ขาย · 5 รับคืน/ลดหนี้ลูกหนี้ · 6 เพิ่มหนี้ลูกหนี้ · 7 ส่งคืน/ลดหนี้เจ้าหนี้ · 8 เบิก · 9 รับคืนจากเบิก · 10 ปรับปรุงเกิน (+) · 11 ปรับปรุงขาด (−) · 12 โอนเข้า · 13 โอนออก · 20 รับคืนสินค้าสำเร็จรูป | `bcdumpprocjob.cpp:2600-2681` |

## 1. ผังเมนู IC ของ Champ (68 รายการ — `menuconfig.xml:407-511`)

คอลัมน์: id เมนู · ชื่อเมนู (ตาม Champ) · command → handler ใน `MainFrm` · คลาสฟอร์ม · บรรทัดใน atlas. รายงาน (5603 เป็นต้นไป) ไม่มีคลาสฟอร์ม เปิดผ่าน `REPComm_InsertCommand` (ดูหัวข้อรายงาน). `*(ปิดในเมนู)*` = `disable="yes"` ใน `menuconfig.xml`.

| id | ชื่อเมนู | command → handler | คลาสฟอร์ม | atlas |
|---|---|---|---|---|
| 5602 | รายละเอียดสินค้า | ID_ICSYS_ITEMMASTER → OnIcsysItemmaster | CICFrmItemMaster (icfrmitemmaster.cpp) | atlas-07:6 |
| 5332 | กำหนดราคาขายสินค้า *(ปิดในเมนู)* | ID_BILLSYS_TRANS_9 → OnBillPriceList | CBILLFrmPriceList (BILLFrmPriceList.cpp) | atlas-07:353 |
| 5005 | กำหนดราคาขาย/โปรโมชั่นสินค้า | ID_80000_30019 → OnPriceStructure | CChmFrmPriceStructureView (ChmFrmPriceStructureView.cpp) | atlas-07:376 |
| 5624 | บันทึกสินค้ายกมา | ID_ICSYS_TRANS_1 → OnIcsysTrans1 | CICFrmStockBalance (icfrmstockbalance.cpp) | atlas-07:409 |
| 5601 | บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ | ID_ICSYS_ISSUE_REQUEST → OnIcIssueRequest | CICFrmStkIssueRequst (ICFrmStkIssueRequst.cpp) | atlas-07:511 |
| 5625 | บันทึกเบิกใช้สินค้า,วัตถุดิบ | ID_ICSYS_TRANS_2 → OnIcsysTrans2 | CICFrmStockIssue (icfrmstockissue.cpp) | atlas-07:575 |
| 5626 | บันทึกรับคืนสินค้า,วัตถุดิบ | ID_ICSYS_TRANS_3 → OnIcsysTrans3 | CICFrmStockIssueRet (icfrmstockissueret.cpp) | atlas-07:885 |
| 5627 | บันทึกรับสินค้าสำเร็จรูป | ID_ICSYS_TRANS_4 → OnFinishGoods | CICFrmFinishGood (icfrmfinishgood.cpp) | atlas-07:954 |
| 5631 | บันทึกใบขอโอนสินค้า | ID_ICSYS_TRANSFER_REQUEST → OnIcTransRequest | CICFrmStkTransferRequest (ICFrmStkTransferRequest.cpp) | atlas-07:1025 |
| 5628 | บันทึกโอนสินค้าระหว่างคลัง | ID_ICSYS_TRANS_5 → OnIcsysTrans5 | CICFrmStockTransfer (icfrmstocktransfer.cpp) | atlas-07:1085 |
| 5629 | บันทึกเอกสารเพื่อตรวจนับสินค้า | ID_ICSYS_TRANS_7 → OnIcsysTrans7 | CICFrmStockInspect (icfrmstockinspect.cpp) | atlas-07:1163 |
| 5599 | บันทึกผลการตรวจนับสินค้า | ID_ICSYS_CHECK_1 → OnICsysCheck1 | CICFrmStockInspectResult (ICFrmStockInspectResult.cpp) | atlas-07:1258 |
| 5630 | บันทึกปรับปรุงสินค้าหลังตรวจนับ | ID_ICSYS_TRANS_8 → OnIcsysTrans8 | CICFrmStockAdjust (ICFrmStockAdjust.cpp) | atlas-07:1270 |
| 5613 | บันทึกสินค้าชุด | ID_ICSYS_ITEMSET → OnIcsysItemset | CICFrmItemSet (icfrmitemset.cpp) | atlas-07:1355 |
| 5615 | ตรวจสอบสินค้าที่สามารถรวมเป็นสินค้าชุดได้ | ID_ICSYS_ITEMSET_BUILD → OnIcsysItemSetBuild | CICFrmItemSetBuild (ICFrmItemSetBuild.cpp) | atlas-07:1388 |
| 5616 | ตรวจสอบชุดสินค้าที่สามารถแยกเป็นสินค้าได้ | ID_ICSYS_ITEMSET_PART → OnIcsysItemSetPart | CICFrmItemSetPart (ICFrmItemSetPart.cpp) | atlas-07:1398 |
| 5614 | บันทึกรายการย่อยสินค้าชุดแบบที่ 2 | ID_ICSYS_ITEMSET_2 → OnICSysItemSetType2 | CICFrmItemSetType2 (ICFrmItemSetType2.cpp) | atlas-07:1408 |
| 5618 | บันทึก Serial Number | ID_ICSYS_SERIAL_NUM → OnIcsysSerialNum | CICFrmItemSerialNumber (ICFrmItemSerialNumber.cpp) | atlas-07:1433 |
| 5600 | ใบตรวจรับสินค้า | ID_ICSYS_GOODS_REC → OnIcsysGoodsRec | CICFrmGoodsRec (ICFrmGoodsRec.cpp) | atlas-07:1479 |
| 5623 | กำหนดลำดับรายวัน | ID_ICSYS_TOOLS_999 → OnICSysTools999 | CICDlgUpdateBehindIndex (ICDlgUpdateBehindIndex.cpp) | atlas-07:1502 |
| 5621 | คำนวณยอดสินค้าใหม่ | ID_ICSYS_TOOLS_1 → OnIcsysTools1 | CICFrmStockReProcess (ICFrmStockReProcess.cpp) | atlas-07:1512 |
| 5564 | ปรับปรุงราคาขายสินค้า *(ปิดในเมนู)* | ID_ICSYS_ADJ_PRICE → OnAdjustPrice | CMyFrmUpdatePrice (MyFrmUpdatePrice.cpp) | atlas-07:1548 |
| 5603 | รายงานรายละเอียดสินค้า | ID_ICSYS_ITEMMASTERREP_1 → OnIcsysItemmasterrep1 | รายงาน (REPComm) | atlas-07:1597 |
| 5608 | รายงานรายละเอียดสินค้า-ตาม Supplier | ID_ICSYS_ITEMMASTERREP_5 → OnIcsysItemmasterrep5 | รายงาน (REPComm) | atlas-07:1601 |
| 5609 | รายงานรายการสินค้าแบบมี Serial | ID_ICSYS_ITEMMASTERREP_6 → OnIcsysItemmasterrep6 | รายงาน (REPComm) | atlas-07:1605 |
| 5610 | รายงานเคลื่อนไหว Serial Number | ID_ICSYS_ITEMMASTERREP_8 → OnIcsysItemmasterrep8 | รายงาน (REPComm) | atlas-07:1608 |
| 5632 | รายงานรายการสินค้ายกมาต้นปี | ID_ICSYS_TRANSREP_0 → OnIcsysTransrep0 | รายงาน (REPComm) | atlas-07:1613 |
| 5633 | รายงานการบันทึกยอดสินค้ายกมาต้นปี | ID_ICSYS_TRANSREP_01 → OnStkBalanceStartYear | รายงาน (REPComm) | atlas-07:1616 |
| 5634 | รายงานโอนสินค้าระหว่างคลังและรายการย่อย | ID_ICSYS_TRANSREP_1_1 → OnIcsysTransrep11 | รายงาน (REPComm) | atlas-07:1621 |
| 5635 | รายงานโอนสินค้าระหว่างคลัง-ตามคลังโอนออก | ID_ICSYS_TRANSREP_1_2 → OnIcsysTransrep12 | รายงาน (REPComm) | atlas-07:1624 |
| 5638 | รายงานโอนสินค้าระหว่างคลัง-ตามคลังโอนเข้า | ID_ICSYS_TRANSREP_1_8 → OnIcsysTransrep18 | รายงาน (REPComm) | atlas-07:1627 |
| 5636 | รายงานโอนสินค้าระหว่างคลัง-ตามสินค้า | ID_ICSYS_TRANSREP_1_3 → OnIcsysTransrep13 | รายงาน (REPComm) | atlas-07:1630 |
| 5637 | รายงานโอนสินค้าระหว่างคลัง-ตามวันที่ | ID_ICSYS_TRANSREP_1_4 → OnIcsysTransrep14 | รายงาน (REPComm) | atlas-07:1633 |
| 5639 | รายงานการเบิกสินค้า,วัตถุดิบ-วันที่เบิก | ID_ICSYS_TRANSREP_2_1 → OnIcsysTransrep21 | รายงาน (REPComm) | atlas-07:1638 |
| 5640 | รายงานการเบิกสินค้า,วัตถุดิบ-ผู้เบิก | ID_ICSYS_TRANSREP_2_2 → OnIcsysTransrep22 | รายงาน (REPComm) | atlas-07:1641 |
| 5641 | รายงานการเบิกสินค้า,วัตถุดิบ-แผนก | ID_ICSYS_TRANSREP_2_3 → OnIcsysTransrep23 | รายงาน (REPComm) | atlas-07:1644 |
| 5642 | รายงานการเบิกสินค้า,วัตถุดิบ-สินค้า | ID_ICSYS_TRANSREP_2_4 → OnIcsysTransrep24 | รายงาน (REPComm) | atlas-07:1647 |
| 5643 | รายงานการเบิกสินค้า,วัตถุดิบ-โครงการ | ID_ICSYS_TRANSREP_2_5 → OnIcsysTransrep25 | รายงาน (REPComm) | atlas-07:1650 |
| 5645 | รายงานการรับคืนเบิก-วันที่ | ID_ICSYS_TRANSREP_3_1 → OnIcsysTransrep31 | รายงาน (REPComm) | atlas-07:1655 |
| 5646 | รายงานการรับคืนเบิก-ผู้เบิก | ID_ICSYS_TRANSREP_3_2 → OnIcsysTransrep32 | รายงาน (REPComm) | atlas-07:1658 |
| 5647 | รายงานการรับคืนเบิก-แผนก | ID_ICSYS_TRANSREP_3_3 → OnIcsysTransrep33 | รายงาน (REPComm) | atlas-07:1661 |
| 5648 | รายงานการรับคืนเบิก-สินค้า | ID_ICSYS_TRANSREP_3_4 → OnIcsysTransrep34 | รายงาน (REPComm) | atlas-07:1664 |
| 5649 | รายงานการรับสินค้าสำเร็จรูป-วันที่ | ID_ICSYS_TRANSREP_4_1 → OnIcsysTransrep41 | รายงาน (REPComm) | atlas-07:1669 |
| 5650 | รายงานการรับสินค้าสำเร็จรูป-แผนก | ID_ICSYS_TRANSREP_4_2 → OnIcsysTransrep42 | รายงาน (REPComm) | atlas-07:1672 |
| 5651 | รายงานการรับสินค้าสำเร็จรูป-สินค้า | ID_ICSYS_TRANSREP_4_3 → OnIcsysTransrep43 | รายงาน (REPComm) | atlas-07:1675 |
| 5652 | พิมพ์เอกสารเพื่อตรวจนับสินค้า | ID_ICSYS_TRANSREP_5_1 → OnIsSysRepInspect | รายงาน (REPComm) | atlas-07:1680 |
| 5653 | รายงานใบตรวจนับสินค้าพร้อมรายการย่อย | ID_ICSYS_TRANSREP_5_2 → OnIsSysRepInspectItem | รายงาน (REPComm) | atlas-07:1683 |
| 5654 | รายงานผลต่างจากการตรวจนับ | ID_ICSYS_TRANSREP_5_3 → OnIsSysRepInspectReult | รายงาน (REPComm) | atlas-07:1686 |
| 5655 | รายงานการปรับปรุงยอดสินค้า | ID_ICSYS_TRANSREP_5_4 → OnIsSysRepAdujust | รายงาน (REPComm) | atlas-07:1689 |
| 5656 | รายงานประเมินการรับสินค้า | ID_ICSYS_TRANSREP_6 → OnAssessGoods | รายงาน (REPComm) | atlas-07:1692 |
| 5566 | รายงาน Lot สินค้า | ID_ICSYS_ANALYREP_1_1 → OnIcsysAnalyrepLotByDate | รายงาน (REPComm) | atlas-07:1697 |
| 5574 | รายงานสินค้าและวัตถุดิบ | ID_ICSYS_ANALYREP_2_4 → OnIcsysAnalyrep24 | รายงาน (REPComm) | atlas-07:1702 |
| 5578 | รายงานบัญชีคุมพิเศษสินค้า | ID_ICSYS_ANALYREP_2_8 → OnIcsysAnalyrep28 | รายงาน (REPComm) | atlas-07:1706 |
| 5571 | รายงานบัญชีคุมพิเศษสินค้าตามกลุ่มสินค้า | ID_ICSYS_ANALYREP_2_10 → OnIcsysAnalyrep210 | รายงาน (REPComm) | atlas-07:1710 |
| 5596 | รายงานเคลื่อนไหวสินค้าต้นทุนมาตรฐาน | ID_ICSYS_ANALYREP_811 → OnICsysAnalyrep811 | รายงาน (REPComm) | atlas-07:1715 |
| 5570 | รายงานเคลื่อนไหวสินค้า | ID_ICSYS_ANALYREP_2_1 → OnIcsysAnalyrep21 | รายงาน (REPComm) | atlas-07:1719 |
| 5575 | รายงานสรุปเคลื่อนไหวปริมาณสินค้า | ID_ICSYS_ANALYREP_2_5 → OnIcsysAnalyrep25 | รายงาน (REPComm) | atlas-07:1723 |
| 5577 | รายงานสินค้าที่ไม่มีการเคลื่อนไหว | ID_ICSYS_ANALYREP_2_7 → OnIcsysAnalyrep27 | รายงาน (REPComm) | atlas-07:1727 |
| 5580 | รายงานยอดคงเหลือสินค้า-ตามสินค้า | ID_ICSYS_ANALYREP_3_1 → OnIcsysAnalyrep31 | รายงาน (REPComm) | atlas-07:1733 |
| 5588 | รายงานยอดคงเหลือสินค้าที่ถึงจุดสั่งซื้อ | ID_ICSYS_ANALYREP_4_1 → OnIclsysAnaByStkMin | รายงาน (REPComm) | atlas-07:1737 |
| 5589 | รายงานยอดคงเหลือสินค้าที่ถึงจุดสูงสุด | ID_ICSYS_ANALYREP_4_2 → OnIcsysAnaByStkMax | รายงาน (REPComm) | atlas-07:1740 |
| 5597 | รายงานยอดคงเหลือสินค้า-ตามคลัง,สินค้า | ID_ICSYS_ANALYREP_86302 → OnICSysAnalyRepStkRemainByWH_Item | รายงาน (REPComm) | atlas-07:1743 |
| 5590 | รายงานประมาณการสั่งสินค้าในแต่ละเดือน | ID_ICSYS_ANALYREP_4_3 → ? | รายงาน (REPComm) | atlas-07:1748 |
| 5592 | รายงานสรุปยอดสินค้าค้างรับแยกตามเจ้าหนี้ | ID_ICSYS_ANALYREP_4_5 → OnStkRemainInOutDetail | รายงาน (REPComm) | atlas-07:1750 |
| 5591 | รายงานสรุปยอดสินค้าค้างรับ | ID_ICSYS_ANALYREP_4_4 → OnIcsysAnaByRemainInOut | รายงาน (REPComm) | atlas-07:1753 |
| 5595 | รายงานสรุปยอดสินค้าค้างส่ง | ID_ICSYS_ANALYREP_4_8 → ? | รายงาน (REPComm) | atlas-07:1756 |
| 5612 | รายงานสินค้าใหม่ *(ปิดในเมนู)* | ID_ICSYS_ITEMNEW → OnltemNew | รายงาน (REPComm) | atlas-07:1758 |
| 5657 | รายงานการกำหนดราคาขายสินค้า *(ปิดในเมนู)* | ID_ICSYS_TRANSREP_PRICELIST → OnIcsysTransrepPriceList | รายงาน (REPComm) | atlas-07:1761 |

กลุ่มตามเมนูจริง (ผังต้นไม้):

| กลุ่มใน Champ | รายการ (id) |
|---|---|
| (ระดับบนของ IC) ข้อมูลหลัก/ราคา | 5602 รายละเอียดสินค้า · 5332 กำหนดราคาขายสินค้า (ปิด) · 5005 กำหนดราคาขาย/โปรโมชั่นสินค้า |
| (ระดับบน) เอกสารสต็อก | 5624 ยกมา · 5601 ใบขอเบิก · 5625 เบิกใช้ · 5626 รับคืนเบิก · 5627 รับสินค้าสำเร็จรูป · 5631 ใบขอโอน · 5628 โอนระหว่างคลัง · 5629 เอกสารตรวจนับ · 5599 ผลตรวจนับ · 5630 ปรับปรุงหลังตรวจนับ |
| popup สินค้าชุด | 5613 · 5615 · 5616 · 5614 |
| (ระดับบน) อื่น ๆ | 5618 Serial Number · 5600 ใบตรวจรับสินค้า · 5623 กำหนดลำดับรายวัน · 5621 คำนวณยอดสินค้าใหม่ · 5564 ปรับปรุงราคาขาย (ปิด) |
| popup รายงานสินค้า (ปิด) | 5603 · 5608 · 5609 · 5610 |
| popup รายงานรายวัน (ปิด) | ยกมา 5632/5633 · โอน 5634/5635/5638/5636/5637 · เบิก 5639-5643 · รับคืนเบิก 5645-5648 · รับสำเร็จรูป 5649-5651 · ตรวจนับ 5652-5655 · 5656 ประเมินการรับสินค้า |
| popup รายงานเชิงวิเคราะห์สินค้า (ปิด) | 5566 Lot · เคลื่อนไหว 5574/5578/5571/5596/5570/5575/5577 · ยอดคงเหลือ 5580/5588/5589/5597 · อื่น ๆ 5590/5592/5591/5595 |
| ปิดในเมนู | 5612 รายงานสินค้าใหม่ · 5657 รายงานการกำหนดราคาขายสินค้า |

> ข้อสังเกต (verify แล้ว): **ไม่มี** เมนูสำหรับ "คลังสินค้า", "ที่เก็บ/ชั้นวาง", "หน่วยนับ", "กลุ่ม/ชนิด/หมวด/รูปแบบ/ยี่ห้อ/สี" ใน IC (`menuconfig.xml:407-511`). มาสเตอร์เหล่านี้เปิดจากปุ่มค้นหา/ลิงก์ในฟอร์มสินค้า (เช่น `OnSysShelfMaster`, `OnSysWarehouseMaster` อยู่ใน MainFrm แต่ไม่ผูกเมนู — ยังไม่ verify ว่าเรียกจากจุดไหนใน UI) — ดูหัวข้อ 3.

## 2. ข้อมูลหลักสินค้า — 5602 "รายละเอียดสินค้า"

### 2.0 คลาสที่ใช้จริง (ยืนยันแล้ว)

เมนู 5602 → `ID_ICSYS_ITEMMASTER` → `OnIcsysItemmaster` (atlas-07:6) แต่ **คลาสที่ MainFrm เปิดจริงในรุ่นนี้คือ `CChmICFrmItemMaster`** (ชุด "Chm" แบบมีแท็บ, ใช้แผงปุ่มลัด `IDR_MAINFRAME_CHM` `BC5Account.rc:977-1011`). คลาสเก่า `CICFrmItemMaster` (`icfrmitemmaster.cpp`) ไม่ถูกเรียกจาก MainFrm — **ห้ามยึดเป็นต้นแบบ**. (ยังไม่ verify: เงื่อนไขเลือกคลาสใน `MainFrm.cpp`.)

โครงฟอร์ม (`IDD_CHM_IC_FRM_ITEMMASTER` 560×310):

| ส่วน | รายละเอียด |
|---|---|
| ตัวควบคุมซ่อน | `ED_ITEM_CODE` (เก็บรหัสสินค้าปัจจุบันให้เฟรม) |
| ตัวเลือกหลังบันทึก (มุมขวาบน) | combo `ED_SAVEOPTION`: 0 «SAVE_NOTCLEAR_SCREEN» = "บันทึกแล้วไม่ล้างหน้าจอ", 1 «SAVE_NEW_REC» = "บันทึกแล้วเพิ่มรายการใหม่" (`ChmICFrmItemMaster.cpp:101-173`) |
| แท็บหลัก 2 แท็บ | แท็บ 0 «MST_HEADER» = ส่วนหัวสินค้า (`CChmICFrmItemMaster1`) · แท็บ 1 «MST_OTHER» = ข้อมูลอื่น (`CChmICFrmItemMaster2`) |
| แท็บย่อยของแท็บ 0 | «DETAIL_1» → `ChmICFrmItemMaster1_1` (ยอดคงเหลือ) · «DETAIL_2» → `1_2` (หน่วยนับ+กลุ่มสินค้า) · «DETAIL_3» → `1_3` (คลัง/ที่เก็บ) |
| แท็บย่อยของแท็บ 1 | «DETAIL_4» → `2_1` (บาร์โค้ด/Supplier/ห้ามขายร่วม-ขายพ่วง) · «DETAIL_5» → `2_2` (บัญชี/เงื่อนไข/ขนาด) · «DETAIL_6» → `2_3` (ชื่อในเอกสาร/จุดรับ/ราคาประกอบ/สินค้าทดแทน) |
| แถบเครื่องมือ | `ICTOOLBAR_RECT` ด้านล่าง + `LB_AUTOPOS` (toolbar มาตรฐานของ `CMyBaseFormView` — ดูเอกสาร common-UI) |

ปุ่มลัด (`ChmICFrmItemMaster.cpp:65-81`; accelerator `BC5Account.rc:977-1011`):

| ปุ่ม | ทำอะไร |
|---|---|
| F11 | บันทึก |
| Ctrl+N | เพิ่มใหม่ |
| Ctrl+F | ค้นหา |
| Ctrl+P | พิมพ์ |
| Ctrl+B | ค้นหาด้วยบาร์โค้ด |
| Ctrl+G | คัดลอกสินค้า (`OnCopyItem`) |
| Ctrl+F12 | เปิดกล่องราคาขาย (`OnPriceList` → `CICDlgPriceList`) |
| Shift+F12 | เปลี่ยนหน่วยนับ (`OnChangeUnit`) |
| Ctrl+F8 | `OnBillGroup` (ยังไม่ verify ว่าทำอะไร) |
| F5 / Shift+F5 / Ctrl+F5 | คำอธิบายเอกสาร / ยืนยัน / ยกเลิก (แผงปุ่มลัดกลาง — ไม่มีผลกับสินค้า, ยังไม่ verify) |
| F2 / Shift+F9 / Ctrl+F9 / Ctrl+F2 / Shift+F2 / F12 | GL / Lot / Serial / ตรวจสต็อก / ขยายแถว / ผังเอกสาร (แผงกลาง) |
| PgUp / PgDn | ไปสินค้าก่อนหน้า/ถัดไป |

### 2.1 แท็บ 0 — ส่วนหัวสินค้า (`IDD_CHM_IC_FRM_ITEMMASTER1`, 600×300) ลำดับฟิลด์บนจอ

| # | ป้าย | ชนิด/ตัวควบคุม | ตาราง.คอลัมน์ | กติกา | หลักฐาน |
|---|---|---|---|---|---|
| 1 | รหัสสินค้า: | text `ED_ITEMCODE` + ปุ่มค้นหา `BTTN_ITEMCODE` + ปุ่ม "บันทึกเป็น…" `BTTN_SAVEAS` | BCItem.Code | จำเป็น (`CANNOT_SAVE_ITEM`); ซ้ำไม่ได้ (ถามลองใหม่); **ปิดแก้เมื่อ `ProductCentralize=1`** (`ChmICFrmItemMaster1.cpp:77`) | `ChmICFrmItemMaster.cpp:394+` |
| 2 | «ITEMS_TYPE» = ประเภทสินค้า (~) | combo ค่าคงที่ `ED_TYPE` | BCItem.StockType | 0 สินค้าทั่วไป · 1 บริการ · 2 สินค้าที่มี Serial (แสดงเฉพาะเมื่อไม่ได้ถอด `LMD_REMOVE_SERIAL_NUMBER`; ถ้าถอด ดัชนีถัดไปเลื่อนขึ้น) · สินค้าเช่า · สินค้ารับฝากขาย · สินค้าชุด · พิเศษ(บุหรี่) · สินค้าชุดแบบที่ 2; ค่าเริ่มต้น 0 | `ChmICFrmItemMaster1.cpp:81-95` |
| 3 | «ItemName1» = ชื่อสินค้า (~) | text `ED_ITEMNAME1` | BCItem.Name1 | จำเป็น (focus กลับ) | `…Master.cpp:394+` |
| 4 | ชื่อที่2: | text `ED_ITEMNAME2` | BCItem.Name2 | ไม่บังคับ | |
| 5 | «COSTTYPE» = วิธีต้นทุน (~) | combo `ED_COST_TYPE` | BCItem.CostType | **เหลือตัวเลือกเดียว «AVENGECOST» = "ต้นทุนเฉลี่ย:"** (FIFO/มาตรฐานถูก comment ออก 27/9/2553 โดย Jum ตามคำขอ K.Boy "remove from chmap2") — แต่ engine ยังรองรับ 0 เฉลี่ย / 1 ล็อต(FIFO) / 2 มาตรฐาน | `ChmICFrmItemMaster1.cpp:97-111`; `bcdumpprocjob.cpp:2756-2766` |
| 6 | หน่วยนับ | combo `ED_UNIT_TYPE` | BCItem.UnitType | 0 «SINGLEUNIT» หน่วยนับเดียว · 1 «MULTIUNIT» หลายหน่วยนับ; ค่าเริ่มต้น 0; **เปลี่ยนจากหลายหน่วยกลับเป็นหน่วยเดียวไม่ได้** (`SINGLEUNITNOPACKING`) | `…1.cpp:127-144` |
| 7 | «STANDARDCOST» = ต้นทุนมาตรฐาน (~) | ตัวเลข `ED_STANDARD_COST` + ปุ่ม "…" | BCItem.StandardCost (+ BCItemStandardCost) | **ปิดไว้ยกเว้น CostType = 2** (`…1.cpp:149`) | |
| 8 | ราคาซื้อล่าสุด | ตัวเลข `ED_LAST_BUY_PRICE` (อ่านอย่างเดียว ~) | BCItem.LastBuyPrice | ตั้งโดยเอกสารซื้อ ไม่ใช่ผู้ใช้ | atlas DDL |
| 9 | «GENERICNAME» = ชื่อสามัญ (~) | combo-table `ED_GENERIC` | BCItem.GenericCode ← BCGeneric | `SELECT * FROM BCGeneric` | `…1.cpp:146` |
| 10 | สถานะ | combo `ED_ACTIVE1` | BCItem.ActiveStatus | «INSALE» / «ICSALE» (ความหมายไทยยังไม่ verify; คอมเมนต์โค้ด "InActive") — **ค่าเริ่มต้น = ตัวเลือกที่ 2 (index 1)** | `…1.cpp:120-125` |
| 11 | ต้นทุนการจัดเก็บ | ตัวเลข `ED_SAVE_COST` + ปุ่ม "ประมวลผล" `BTTN_CHANGE_CODE` | BCItem.CostOfArea (~) | ยังไม่ verify ความหมายปุ่ม "ประมวลผล" | |
| 12 | ภาษี (`ED_VAT`; rc ไม่มีป้ายชัด) | combo | BCItem.ExceptTax | 0 «TAXCASE» (~ มีภาษี) · 1 «IC_EXCEPT_TAX» (~ ยกเว้นภาษี); ค่าเริ่มต้น 0 | `…1.cpp:113-118` |
| 13 | หมายเหตุ: | text `ED_DESCRIPTION` | BCItem.MyDescription | | |
| 14 | รูปสินค้า 4 ด้าน | กรอบภาพ `FRAME_PIC_FRONT/BACK/RIGHT/LEFT` (ขวาบน) | BCItem.PicFileName1-4 (หน้า/หลัง/ขวา/ซ้าย) | เก็บเป็นพาธไฟล์ (ใน BC **ต้องเก็บ S3 + thumbnail ตามกฎ repo**) | `…Master.cpp:1111-1217` |

> ป้ายที่ขึ้นเป็น «KEY» ถูกกำหนดตอนรันจาก `BCResData.mdb` (ไม่อยู่ใน repo) — ขอลุงจืดเปิดโปรแกรมจริงยืนยันคำไทยก่อนสร้างจอ BC.

### 2.2 แท็บย่อย 1_1 «DETAIL_1» — ยอดคงเหลือ (อ่านอย่างเดียว) (`IDD_CHM_IC_FRM_ITEMMASTER_PAGE1`)

| ป้าย (ตามหน้าจอ) | ตัวควบคุม | ที่มา |
|---|---|---|
| จำนวนคงเหลือ | `ED_QTY_BALANCE` + `ED_QTY_WORD` (~ข้อความหน่วยย่อย) | BCItem.StockQty |
| ต้นทุนเฉลี่ย: | `ED_AVERAGE_COST` | BCItem.AverageCost |
| มูลค่าคงเหลือ | `ED_AMOUNT_BALANCE` | BCItem.Amount (กลไกหัวข้อ 5.2) |
| สั่งซื้อคงค้าง | `ED_PO_QTY` | ใบสั่งซื้อค้างรับ (~BCItem.RemainInQty) |
| สั่งจองคงค้าง | `ED_RESERVE_QTY` | BCItem.ReserveQty |
| สั่งขายคงค้าง | `ED_SO_QTY` | ~BCItem.RemainOutQty |
| ปุ่ม "แสดงยอดคงเหลือตามคลังสินค้า" | `BTN_WH_QTY_BALANCE` | แสดงกริดตามคลัง |
| ปุ่ม "Serial Number" | `BTN_SERIAL_NO` | เปิดกล่อง serial ของสินค้านี้ |
| ปุ่ม "…" ต้นทุนเฉลี่ย / มูลค่าคงเหลือ / «STANDARDCOST» | `BTN_AVERAGE_COST`, `BTN_AMOUNT_BALANCE`, `BTN_STD_COST_SPEC` | รายละเอียด/ประวัติต้นทุน (ยังไม่ verify) |

กริดคงเหลือแยกคลัง (BCStkWarehouse): คลัง/ชื่อ (`WHCODENAME`), ที่เก็บ/ชื่อ (`SHEFTCODENAME`), หน่วยนับ (`UINTCODE`), จำนวน (`QTY`), วันที่รับล่าสุด (`LAST_RECEIVE_DATE`) — `ChmICFrmItemMaster1_1.cpp`.

**กฎสำคัญ (Champ quirk): จำนวน/ต้นทุน/มูลค่าบน BCItem ไม่ใช่ฟิลด์ที่ผู้ใช้พิมพ์** — เปลี่ยนโดยเอกสารสต็อกผ่าน process เท่านั้น (หัวข้อ 5).

### 2.3 แท็บย่อย 1_2 «DETAIL_2» — หน่วยนับ + กลุ่มสินค้า (`IDD_CHM_IC_FRM_ITEMMASTER_PAGE2`)

กรอบซ้าย "หน่วยนับ [Unit]":

| ป้าย | ตัวควบคุม | คอลัมน์ BCItem | หมายเหตุ |
|---|---|---|---|
| หน่วยนับมาตรฐาน | `ED_STD_UNIT` | DefStkUnitCode (+DefFixUnitCode) | จำเป็น (`NOT_SET_UNITCODE`) |
| หน่วยนับขาย | `ED_SALE_UNIT` | DefSaleUnitCode | จำเป็นเมื่อหลายหน่วย |
| หน่วยนับซื้อ | `ED_BUY_UNIT` | DefBuyUnitCode | จำเป็นเมื่อหลายหน่วย |
| หน่วยซื้อ E-ordering | `ED_BUY_UNIT2` | (ยังไม่ verify) | |
| หน่วยนับขนาน1 / ขนาน2 | `ED_PARALLEL_UNIT1/2` | OppositeUnit / OppositeUnit2 | «OPPOSITEUNIT» ถูกถอดจาก combo หน่วยนับ (comment 27/9/2553) แต่ช่องยังอยู่บนจอ |
| ราคาขายหน่วย1 | `ED_SALE_PRICE_UNIT1` | Opposite1Price | |
| `ED_RATE1 ED_UNIT_RATE1 = ED_RATE2 ED_UNIT_RATE3` | อัตราส่วนหน่วยขนาน | OppositeRate | |

> **หน่วยเดียว (`UnitType=0`)**: โค้ดตั้ง DefStk/DefFix/DefSale/DefBuyUnitCode = หน่วยนับมาตรฐานทั้งหมดตอนบันทึก (`ChmICFrmItemMaster.cpp:1111-1217`).

กรอบกลาง «MULTI_UNIT_RATE» [PackingRate] = ตารางอัตราแปลงหน่วย (`REP_LIST`; ตาราง BCStkPacking: ItemCode, UnitCode, Rate1, Rate2, Rate, Running, Runner, SCGUnitcode) พร้อมปุ่ม "เพิ่ม/บันทึก/ลบ" (`BTN_NEW/BTN_SAVE/BTN_DEL`), ช่อง `ED_UNIT_RATE` + หมายเหตุ; หลายหน่วยต้องมีอย่างน้อย 1 แถว (`SET_RATIO_ITEM`).

กรอบขวา «DETAILITEMGROUP» = รายละเอียดกลุ่มสินค้า (ลำดับบนจอ):

| ป้าย (rc) | ตัวควบคุม | คอลัมน์ BCItem | มาสเตอร์ | จำเป็น |
|---|---|---|---|---|
| กลุ่มสินค้า: | `ED_CATE` | GroupCode | BCItemGroup | **ใช่** (`CANNOT_SAVE_NO_ITEMCATEGORY`) |
| ชนิดสินค้า: | `ED_SUB_CATE` | TypeCode | BCItemType | ใช่ (`…SUBCATEGORY`) |
| หมวดสินค้า: | `ED_CLASS` | CategoryCode | BCItemCategory | ใช่ (`…CLASS`) |
| รูปแบบสินค้า: | `ED_SUB_CLASS` | FormatCode | BCItemFormat | ใช่ (`…SUBCLASS`) |
| รหัสสี | `ED_COLOR` | ColorCode | (ยังไม่ verify) | ไม่ |
| รหัสยี่ห้อ: | `ED_BRAND` | BrandCode | BCItemBrand | ไม่ |
| เกรด: | `ED_GRADE` | MyGrade | | ไม่ |

(เงื่อนไขเช็คจำเป็น 4 ฟิลด์แรกอยู่ที่ `ChmICFrmItemMaster.cpp:636-678` — ยังไม่ verify เงื่อนไขครอบ, สรุปเบื้องต้น = บังคับ.)

### 2.4 แท็บย่อย 1_3 «DETAIL_3» — คลัง/ที่เก็บ (`IDD_CHM_IC_FRM_ITEMMASTER_PAGE3`)

กรอบซ้าย «ORIGINALWH» = คลังตั้งต้น — แต่ละแถวมี 2 ช่อง (คลัง + ที่เก็บ):

| ประเภท | ช่องคลัง / ช่องที่เก็บ | คอลัมน์ BCItem |
|---|---|---|
| ซื้อ | `ED_WH_BUY` / `ED_CATE` | DefBuyWHCode / DefBuyShelf |
| ขาย | `ED_WH_SALE` / `ED_SUB_CATE` | DefSaleWHCode / DefSaleShelf |
| สินค้าสำเร็จรูป (Finish goods) | `ED_WH_FG` / `ED_CLASS` | DefFGWHCode / DefFGShelf |
| วัตถุดิบ (Raw material) | `ED_WH_RAW` / `ED_SUB_CLASS` | DefRawWHCode / DefRawShelf |
| ของเสีย/ชำรุด | `ED_WH_BAD` / `ED_COLOR` | DefBadWHCode / DefBadShelf |
| ซ่อม | `ED_WH_REPAIR` / `ED_BRAND` | DefRepairWHCode / DefRepairShelf |

(ป้ายใน .rc ของหน้านี้คัดลอกมาจากหน้าอื่น — ชื่อตัวควบคุม `ED_WH_*` และคอลัมน์ `Def*` เป็นตัวบอกความหมายที่เชื่อถือได้.)

กรอบขวา «WHANDSHELF» = รายการ "คลัง + ที่เก็บ" ที่สินค้านี้ใช้ได้ (กริด → ตาราง **BCItemWarehouse**: ItemCode, WHCode, ShelfCode) + ปุ่ม "ทั้งหมด" 2 ปุ่ม (`BTN_ALL1/BTN_ALL2` ~เติมทุกคลัง/ที่เก็บ — ยังไม่ verify). **กฎบันทึก:** ต้องมีคลัง+ที่เก็บตั้งต้น และกริดนี้ต้องมี ≥ 1 แถว (`NOT_SET_WHCODE_SHELF`), ทุกแถวมีทั้งคลังและที่เก็บ — กริดนี้คือ **รายการที่ combo คลัง/ที่เก็บในเอกสารสต็อกดึงมาให้เลือก** (`icfrmstockissue.cpp` combo จาก BCItemWarehouse).

### 2.5 แท็บย่อย 2_1 «DETAIL_4» (`IDD_CHM_IC_FRM_ITEMMASTER_PAGE4`)

| กริด (ตามตำแหน่งบนจอ) | ตาราง | คอลัมน์ในกริด |
|---|---|---|
| "หน่วยนับ:" (บนซ้าย) = **บาร์โค้ด** | BCBarCodeMaster (BarCode, BarcodeName, UnitCode, ActiveStatus, Remark) | BARCODE, BARCODE_NAME, UNITCODE, ACTIVE, REMARK, ปุ่มพิมพ์บาร์โค้ด — บันทึกต้องไม่ซ้ำ (`BARCODE_REPEAT`) |
| «SUPPLIER» = ผู้จำหน่ายของสินค้านี้ | BCItemSupplier | 15 คอลัมน์: SUPPCODE, SHIPDATE, STDPURPRICE, UNITSOLD, REDUCECOSTMETHOD, OTHERCONDITION, PRIORITY, DATESTART, DATESTOP, BUYPRICE, ITEMCODE_BYSUPPLIER, ITEMNAME_BYSUPPLIER (+3 ช่องยังไม่ verify) |
| «NOTMATCHSALE» = สินค้าห้ามขายร่วมกัน | BCItemMatch (MatchStatus = 1) | ITEMCODE, ITEMNAME, UNITCODE |
| «MATCHAPPENDED» = สินค้าขายพ่วง | BCItemMatch (MatchStatus = 0) | ITEMCODE, ITEMNAME, UNITCODE, ITEMQTY |

### 2.6 แท็บย่อย 2_2 «DETAIL_5» (`IDD_CHM_IC_FRM_ITEMMASTER_PAGE5`) ลำดับบนจอ

| คอลัมน์ซ้าย | คอลัมน์ BCItem | คอลัมน์ขวา | คอลัมน์ BCItem |
|---|---|---|---|
| กลุ่มการ Post บัญชี (`ED_ITEM_ACC` → **BCItemAccGroup**) | AccGroupCode | จุดต่ำสุด (`ED_MIN_POINT`) | StockMin |
| รหัสสูตร1 (`ED_FORMULAR1` → BCFormula) | Formula1 | จุดสูงสุด (`ED_MAX_POINT`) | StockMax |
| รหัสสูตร2 | Formula2 | จุดที่ควรสั่งซื้อ (`ED_PO_POINT`) | OrderPoint |
| รหัสสูตร3 | Formula3 | ค่าขนส่ง (`ED_DELI_CHARGE`) | DeliveryCharge |
| ไฟล์อ้างอิง (`ED_FILE` + ปุ่ม Add) | SpecFileName | ค่าติดตั้ง (`ED_INSTALL_CHARGE`) | InstallCharge |
| ไฟล์ AVI (`ED_AVI_FILE`) | AviFileName | ฐานคิดสินค้าพิเศษ (`ED_BASE_CALC`) | BasePrice |
| อัตราภาษี ณ ที่จ่าย (`ED_WTAX_RATE`) | WTaxRate | สถานะสินค้า (`ED_ITEM_STATUS` → BCItemStatus) | ItemStatus |
| การรับคืนสินค้า (`ED_RET_STATUS`): «RECEIVE» / «NOTRECEIVE» / «CONDITIONRECV» | ReturnStatus | การรับซ่อมสินค้า (`ED_RENOVATION`): «RenovateFree» / «RENOVATEPAY» / «RENOVATNOT» | RenovateStatus |
| หักส่วนลดรับคืน (`ED_CN_DISCOUNT`) | ReturnCharge | ชื่อย่อ (`ED_SHORTNAME`) | ShortName |
| สถานะขาย (`ED_SALE_STATUS`): «SPECLTD» / «NOT_OVERSTK» | StockProcess | ☐ «DISCOUNT_CASHCARD» (`CHK_CASH_CARD`) | DiscCashCard |
| หมายเหตุการคืนสินค้า (`ED_CN_REMARK`) | ReturnRemark | | |
| วันที่เคลื่อนไหวล่าสุด (`ED_LAST_MOVE`, อ่านอย่างเดียว) | (คำนวณ) | รหัสกลุ่ม/รุ่น (`ED_MODEL`) | MyClass |
| รับเงินมัดจำ (`ED_DEPOSIT_REC`) | DepositCondit | ขนาด (`ED_SIZE`) | MySize |
| รับฝากขาย (`ED_CONSIGNMENT`) | VenderCode | กว้าง (`ED_WIDTH`) m | Wide |
| กลุ่มพนักงาน (`ED_EMP_GROUP`) | UserGroup | ยาว (`ED_LONG`) m | Long |
| กลุ่มลูกค้า (`ED_CUST_GROUP`) | CustGroup | สูง (`ED_HIGH`) m | High |
| | | น้ำหนัก/หน่วยนับมาตรฐาน (`ED_W_PER_UNIT`) Kg. | Weight |

(ฟิลด์→คอลัมน์จาก `OnSaveTab*` `ChmICFrmItemMaster.cpp:1111-1217`; ป้ายไทยจาก .rc ของ `IDD_CHM_IC_FRM_ITEMMASTER_PAGE5` — ป้ายบางตัวใน rc ซ้ำกัน เช่น "หักส่วนลดรับคืน" ปรากฏหลายแถว ให้ถือชื่อตัวควบคุม/คอลัมน์เป็นตัวบอกความหมาย; ยังไม่ verify กับหน้าจอจริง.)

> **BCItemAccGroup (Code, Name, ItemAccount, CostAccount, RevenueAccount)** คือจุดเชื่อม "สินค้า → บัญชี GL" ที่เอกสารสต็อก/ขาย/ซื้อใช้ลงบัญชี (หัวข้อ 5.5). ใน BC ผูกที่ master นี้ ห้ามแตกเงื่อนไขตามรหัสบัญชี (กฎ repo).

### 2.7 แท็บย่อย 2_3 «DETAIL_6» (`IDD_CHM_IC_FRM_ITEMMASTER_PAGE6`; กริด 6 ชุด)

| ตำแหน่ง | หัวข้อ (key) | ตาราง | คอลัมน์กริด (ลำดับ) | หลักฐาน |
|---|---|---|---|---|
| ซ้ายบน | «BILLINGNAME» = ชื่อสินค้าในเอกสาร (~) | BCItemBillDesc (DescGroup, BillDesc, LineNumber) | รหัสกลุ่ม (`GROUPCODE`), «BILLINGDES» (ชื่อที่พิมพ์) | `ChmICFrmItemMaster2_3.cpp:82-85` |
| ขวาบน | «CODE_FOR_BILLING» = รหัสสินค้าสำหรับขาย (~) | BCItemCodeForSale (MyDesc, หน่วย, คลัง, ที่เก็บ) | «DETAILS», หน่วยนับ, คลัง, ที่เก็บ | `…2_3.cpp:101-104` |
| ซ้ายกลาง | «LEAD_KEEP» (~ จุดเก็บตามลำดับ) | BCRecProduct (WH, Shelf, Priority, Unit, Min, Max) | คลัง, ที่เก็บ, Priority, หน่วยนับ, MIN, MAX | `…2_3.cpp:165-170` |
| ขวากลาง | «SET_PRICE_ERECT» (~ ตั้งราคาติดตั้ง) | BCPriceErect (Unit, PriceErect, MyDescription) | หน่วยนับ, «PRICE_ERECT» | `…2_3.cpp:225-226` |
| ซ้ายล่าง | «SET_POINTDOWN» (~ จุดรับสินค้า) | BCRecPoint (BCReceiveMaster code, Priority, IsWarehouse) | «CODE» (combo จาก BCReceiveMaster), Priority, «STOCK_DEFAULT» (combo: «NO1» "รับเอง" / «YES» "ส่งให้") | `…2_3.cpp:135-152` |
| ขวาล่าง | «ITEMSUBSTITUTE» = สินค้าทดแทน | BCItemSubstitute | รหัสสินค้า (link), ชื่อสินค้า, หน่วยนับ, หมายเหตุ | `…2_3.cpp:200-203` |

กริด Lot ในไฟล์เดียวกัน (LOTNUM, คลัง, ที่เก็บ, วันที่ซื้อ, เลขที่เอกสาร, วันที่ผลิต, อายุ(วัน), วันหมดอายุ, หมายเหตุ, จำนวน, ต้นทุน, มูลค่า) **ถูก comment ไม่สร้าง** (`ChmICFrmItemMaster2_3.cpp:247`) — ไม่ต้องทำใน BC.

### 2.8 กติกาตอนบันทึก / ลบ / ตารางที่แตะ

ลำดับ `ConfirmSaveData` (`ChmICFrmItemMaster.cpp:394-700`) — BC ต้องเรียงเหมือนกัน:

1. รหัสสินค้าว่าง → `CANNOT_SAVE_ITEM`. 2. รหัสซ้ำ → ถาม Yes/No. 3. ชื่อสินค้า 1 ว่าง → โฟกัสช่องชื่อ. 4. `MULTIUNIT` ต้องมีแถวอัตราแปลงหน่วย ≥ 1 (`SET_RATIO_ITEM`). 5. เปลี่ยนหลายหน่วย→หน่วยเดียวไม่ได้ (`SINGLEUNITNOPACKING`). 6. บันทึกเป็นรหัสใหม่แต่ชื่อเดิม → ถาม `SAVE_NAME1RE`. 7. หน่วยนับมาตรฐานต้องตั้ง (`NOT_SET_UNITCODE`); หลายหน่วยต้องตั้งหน่วยขาย+ซื้อ. 8. คลัง/ที่เก็บตั้งต้นต้องตั้ง; กริดคลัง+ที่เก็บต้องมีแถวและครบ (`NOT_SET_WHCODE_SHELF`). 9. กลุ่ม/ชนิด/หมวด/รูปแบบ ครบ 4 ช่อง. 10. บาร์โค้ดใหม่ต้องไม่ซ้ำ (`BARCODE_REPEAT`). 11. เปลี่ยนหน่วยนับ → ถามยืนยัน (`OLDUNITX` / `CHANG_UNIT_DAILY`).

`OnSaveData` (`:702-1109`): INSERT/UPDATE `BCItem` (สร้าง: `CreatorCode`+`CreateDateTime`; แก้: `LastEditorCode`+`LastEditDateT`) → แต่ละหน้าบันทึกตารางลูก (BCBarCodeMaster, BCStkPacking, BCItemWarehouse, BCItemMatch, BCItemSupplier, BCItemBillDesc, BCItemCodeForSale, BCRecPoint, BCRecProduct, BCItemSubstitute, BCPriceErect) → log ผู้ใช้ → **ถ้าแก้สินค้าเดิม ถาม `RECALC_ITEM`**: ถ้าตอบใช่ UPDATE `packingrate1/2` ในตารางรายการย่อยของเอกสาร 15 ตาราง (BCStkBalanceSub, BCFinishGoodsSub, BCPurchaseOrderSub, BCPOVoidSub, BCAPInvoiceSub, BCStkTransfSub3, BCDebitNoteSub2, BCSaleOrderSub, BCSOVoidSub, BCARInvoiceSub, BCCreditNoteSub, BCDebitNoteSub1, BCStkRefundSub, BCStkIssueSub, BCStkAdjustSub) และ `INSERT ProcessStock(ItemCode, DocDate NULL, ProcessFlag 1, FlowStatus 1, ProcessType 0, ProcessCase 0)` (`:897-899`) ให้ Robot คิดสต็อกใหม่; ถ้าเปลี่ยนหน่วย → `UpdateNewUnit`. ตัวเลือก "บันทึกแล้วไม่ล้างหน้าจอ" (0) = คงหน้าจอ ล้างเฉพาะรหัส/บาร์โค้ด/จำนวน/รูป แล้วสลับเป็นโหมดเพิ่ม.

ลบสินค้า (`ConfirmDeleteData` `:1874`, `ChkInvoice` `:2695`): **ลบไม่ได้ถ้าสินค้าอยู่ในเอกสาร** BCARInvoiceSub, BCAPInvoiceSub, BCStkIssueSub, BCStkIssRetSub, BCStkBalanceSub, BCFinishGoodsSub, BCStkTransfSub, BCStkTransfSub2 (`CANNOT_DELETE_USE`); ถาม `WANT_DELETE_ITEM`; แล้วลบจาก (`:2218-2277`) bcbarcodemaster, bcstkpacking, bcitemmatch, bcitemsupplier, bcitembilldesc, bcstkwarehouse, bcstklocation, bcitemsubstitute, bcsupplierpromo, bcitemcomponents(parentcode), bcpricelist, bcitem, bcarpricelist, bcitemstandardcost, bcitemwarehouse, bcrecpoint, bcrecproduct, bcitemcodeforsale.

> ข้อสังเกต: การตรวจ "ใช้แล้วลบไม่ได้" ใน Champ ตรวจแค่ 8 ตาราง (ไม่รวมใบสั่งซื้อ/ใบสั่งขาย/ใบลดหนี้ ฯลฯ) — BC ควรตรวจกว้างกว่า (FK) แต่ Champ = baseline ให้แจ้งลุงจืดก่อน.

## 3. ราคาขาย, คลัง/ที่เก็บ และมาสเตอร์ที่เกี่ยวข้อง

### 3.1 5005 "กำหนดราคาขาย/โปรโมชั่นสินค้า" (`CChmFrmPriceStructureView` + Views1-4, `IDD_CHM_IC_PRICE_STRUCT*`; atlas-07:376)

เมนูนี้เป็นหน้าจอ **ตั้งราคาแบบเงื่อนไข** (เลือกกลุ่มสินค้าแล้วประมวลผลเข้ากริด → แก้ → บันทึก) ไม่ใช่รายการทีละสินค้า. ส่วนหัว (ตัวกรอง):

| กลุ่มตัวกรอง | ฟิลด์ (จาก–ถึง) | มาสเตอร์ |
|---|---|---|
| สินค้า | จากรหัสสินค้า – ถึงรหัสสินค้า | BCItem |
| ภาษี | ประเภทภาษี | combo (CB_TAXTYPE_1/2 + ทั้งหมด) |
| กลุ่มสินค้า | จาก–ถึง กลุ่ม(Category)/ชนิด(Sub-Category)/หมวด(Class)/รูปแบบ(Sub-Class)/ยี่ห้อ | BCItemGroup / BCItemCategory / BCItemType / BCItemFormat / BCItemBrand |
| ชนิดราคาที่ต้องการแก้ | ☐ ราคาขาย ☐ โปรโมชั่นราคาขาย ☐ โปรโมชั่นแถมตามจำนวน ☐ โปรโมชั่นแถมตามมูลค่า | เลือกได้ทีละหลายชนิด (แต่ละชนิด = 1 แท็บ) |
| ปุ่ม | "ประมวลผลเงื่อนไข" · "Clear screen" · "บันทึก" | |

แท็บ (4 หน้า; `ChmFrmPriceStructureView1-4.cpp`):

| แท็บ | บันทึกลง | คอลัมน์กริด (ลำดับ) |
|---|---|---|
| 1 ราคาขาย | `BCPriceList` (+`BCPriceList_Log`) | ITEMCODE, ITEMNAME, UNITCODE, FROM_QTY (ค่าเริ่ม 1), TO_QTY (ค่าเริ่ม 99999), FROM_DATE1, STOPDATE, SALETYPE (ประเภทการขาย; ค่าเริ่ม 0 หรือ 1 ตาม `DefItemBillType`), CONDIT_TRANS (เงื่อนไขขนส่ง), SALEPRICE1…SALEPRICE8, TAX_TYPE (ค่าเริ่มตาม `DefaultTaxType`) — เครื่องมือปรับ: ประเภทการปรับ, ระดับราคา, อัตราเปลี่ยน/หน่วย, "แสดงรายการย่อย" |
| 2 โปรโมชั่นราคา | `BCPromoPrice` (+Log) | FROM_QTY, TO_QTY (1 / 99999), วันที่เริ่ม-สิ้นสุด, FROM_AR_GROUP, TO_AR_GROUP (กลุ่มลูกหนี้ → BCARGroup), PROMOTION_PRICE, LIST_SALE, TAX_TYPE |
| 3 แถมตามจำนวน | `BCPromotionSub` (+Log) | แท็บ 2 + PREMIUM_ITEM_CODE, PREMIUM_ITEM_NAME, PREMIUM_UNIT_CODE, ADDITION_AMOUNT |
| 4 แถมตามมูลค่า | `BCPromotionValue` (+_Log) | TAX_TYPE, PROMOTION_FROM_AMT1, PROMOTION_TO_AMT1, วันที่, กลุ่มลูกหนี้, สินค้าแถม/หน่วย/จำนวน |

DDL `BCPriceList`: ItemCode, UnitCode, FromQty/ToQty, StartDate/StopDate, SalePrice1-20 (Champ เก็บไว้ถึง 20 ระดับ แต่หน้าจอใช้ 1-8), LowPrice, RetailPrice, TaxType, SaleType, TransportType (atlas.json tables). ทุกการบันทึกเขียน `_Log` คู่กัน (ประวัติราคา).

ตัวกริดเดียวกันในจอเก่า **5332 "กำหนดราคาขายสินค้า" (ปิดในเมนู, `CBILLFrmPriceList`)**: ส่วนหัว = รหัสสินค้า + ชื่อ, กลุ่มพนักงาน (`ED_SALEGROUP`), กลุ่มลูกค้า (`ED_CUSTGROUP`), ปุ่ม "Update date" (กล่อง "ปรับปรุงวันที่ส่งของให้ลูกค้า" ช่วงรหัส/ช่วงวันที่ → วันที่ใหม่); 3 กริด: "โครงสร้างราคาสินค้า", "กำหนดราคาขายสินค้าตามจำนวนที่ลูกค้าซื้อ", "กำหนดช่วงโปรโมชั่นของสินค้าแยกตามสินค้า" (ใช้คอลัมน์ชุดเดียวกับแท็บ 1 ข้างบน — `BILLFrmPriceList.cpp:443-470`). **BC: ใช้ 5005 เป็นหน้าหลัก; 5332 เป็นหน้าเดิมที่ Champ ซ่อนไว้ ไม่ต้องทำซ้ำ** (แต่ปุ่ม Ctrl+F12 ในสินค้าเปิดกล่องราคาของสินค้านั้น `CICDlgPriceList`).

### 3.2 5564 "ปรับปรุงราคาขายสินค้า" (ปิดในเมนู; `CMyFrmUpdatePrice`, `IDD_MY_FRM_UPDATEPRICE`; atlas-07:1548)

ปรับราคาขายเป็นชุด: ตัวกรอง จาก–ถึง ของ รหัสสินค้า / กลุ่ม / ชนิด / หมวด / รูปแบบ / ยี่ห้อ / รหัสกลุ่ม(MyClass) / เกรด / สี / ขนาด / หน่วยนับ / ประเภทภาษี → ปุ่ม "แสดงรายการ" (`OnBttnShowItem`) เติมกริด (ITEMCODE, UNITCODE, FROM_QTY, TO_QTY, FROM_DATE1, STOPDATE, SALETYPE, CONDIT_TRANS, SALEPRICE1-8, TAX_TYPE) → แถบ "ปรับปรุง : [ชนิดการปรับ] [อัตราการเปลี่ยน]" + ช่องกำหนดราคาขาย 1-8 → "ปรับราคา" (`OnBttnAdjustPrice`) → "ตกลง F11" (`OnBttnProcessPrice`) บันทึก (`MyFrmUpdatePrice.cpp:78-83,124-162`). ซ่อนในเมนู = ฟังก์ชันซ้ำกับ 5005 (BC ไม่ต้องทำเป็นเมนูแยก — ยุบรวมตามกฎ repo "ซ้ำซ้อนให้ยุบ").

### 3.3 คลังสินค้า / ที่เก็บ (Warehouse / Shelf)

| มาสเตอร์ | คลาส / ฟอร์ม | ฟิลด์ | หลักฐาน |
|---|---|---|---|
| คลังสินค้า | `CICFrmItemWarehouse` (handler `OnSysWarehouseMaster`) | BCWarehouse: Code, Name, MapPicFileName, UserGroup | atlas.json tables; MainFrm handler |
| ที่เก็บ | `CMyFrmWarehouseMaster` (`IDD_MY_FRM_WAREHOUSE`, `MyFrmWarehouseMaster.cpp`; handler `OnSysShelfMaster`) | BCShelf: Code, Name, WHCode — กริด 3 คอลัมน์: SHELF_ID, SHELF_NAME, WAREHOUSE_ITEM (combo จาก BCWarehouse) | `MyFrmWarehouseMaster.cpp` (คอลัมน์) |
| คลังโอนอัตโนมัติ | `CICFrmItemWHTransDefault` (`OnSysWHTransDefault`) | ค่าเริ่มต้นการโอน (ยังไม่ verify ฟิลด์) | MainFrm handler |

- ปุ่มลัดฟอร์มที่เก็บ: F5 = คัดลอก, F11 = บันทึก. **บันทึก = ลบทั้งตาราง `BCShelf` แล้ว INSERT ใหม่ทุกแถว** (`DELETE FROM BCShelf` ตามด้วย INSERT (Code, Name, WHCode)) — ตรวจก่อนว่าชื่อที่เก็บต้องมีรหัสคลัง และชื่อต้องไม่ว่าง (`MyFrmWarehouseMaster.cpp:131-187`). **อย่า copy พฤติกรรม "ลบแล้วสร้างใหม่" ไปใน BC** (ที่เก็บถูกอ้างอิงจากสต็อก — ต้อง UPSERT) แจ้งลุงจืดเป็น "ต่างโดยตั้งใจ".
- **ไม่มีเมนูของ คลัง/ที่เก็บ/หน่วยนับ/กลุ่ม-ชนิด-หมวด-รูปแบบ-ยี่ห้อ ใน IC** (`menuconfig.xml:407-511`) — เข้าถึงจากปุ่มค้นหาที่ฟิลด์ในฟอร์ม (คลิกปุ่ม lookup → กล่องค้นหา → ปุ่มเพิ่ม ~ยังไม่ verify) หรือจากเมนูกลุ่มอื่น (ข้อมูลหลักกลาง — อยู่นอกขอบเขตไฟล์นี้).
- ค่าเริ่มต้นสต็อก: คลัง+ที่เก็บของแต่ละสินค้า = BCItemWarehouse (หัวข้อ 2.4); ยอดคงเหลือต่อคลัง/ที่เก็บ/หน่วย = **BCStkWarehouse** (หัวข้อ 5.3).

## 4. เอกสารสต็อก (เมนูระดับบนของ IC)

> ทุกเอกสารใช้โครงร่วมเดียวกันตามหัวข้อ 0.1 (หัว + กริด + คอลัมน์เงา + ConfirmSaveData → OnSaveData → process). หัวข้อนี้ระบุเฉพาะส่วนที่ต่างกัน. **สิทธิ์เห็นต้นทุน:** คอลัมน์ต้นทุนในเอกสารเบิก/รับคืน (`UNITCOST`, `RETURN_COST`) แก้ไม่ได้ถ้าผู้ใช้ไม่มีสิทธิ์ `bCanSeeItemCost` (`icfrmstockissueret.cpp:710-717`) — BC ต้องมีสิทธิ์นี้ระดับผู้ใช้. **โหมด "Chm"** (`theApp.m_bChmMode`) ซ่อนคอลัมน์ "จำนวนเงิน" ในใบรับคืนเบิก (`icfrmstockissueret.cpp:718-719`) และติ๊ก "ภาษี(TAXCASE)" ไว้ล่วงหน้าในใบปรับปรุง (`ICFrmStockAdjust.cpp:760`) — เป็นตัวเลือกเชิงรุ่น (ยังไม่ verify ว่า Chm = รุ่นใด) ให้ลุงจืดเลือก.

ป้ายคอลัมน์ที่ขึ้นเป็น `«KEY»` ใช้คำไทยจากคอมเมนต์ข้างโค้ด (ระบุ "ตามคอมเมนต์") เมื่อมี.

### 4.1 5624 "บันทึกสินค้ายกมา" (`CICFrmStockBalance`, `IDD_IC_FRM_STOCKBALANCE` 436×231; atlas-07:409)

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร (`ED_DOCNO` + ปุ่มค้นหา) · วันที่ (`ED_DOCDATE`, ปิดแก้ไข) · หมายเหตุ (`ED_DESC`) · แผนก (`ED_DEPART_CODE` ← BCDepartment) · (ท้ายจอ) จำนวนเงิน (`ED_AMOUNT` = ผลรวมกริด, ทศนิยม `AmountPoint`) |
| กริด (ตามลำดับ) | 1 รหัสสินค้า (แก้ได้ + ลิงก์ค้นหา) · 2 ชื่อสินค้า (อ่านอย่างเดียว) · 3 คลัง (combo) · 4 ที่เก็บ (combo) · 5 หน่วยนับ (combo) · 6 จำนวน («ITEMQTY», `QtyPoint`) · 7 ต้นทุน/หน่วย («UNITCOST», ตามคอมเมนต์ "ต้นทุน/หน่วย", `CostPoint`) · 8 รวมเงิน («TOTALS», ตามคอมเมนต์ "รวมเงิน", `AmountPoint`) |
| การคำนวณ | แก้ "รวมเงิน" (คอลัมน์ 7) → ต้นทุน/หน่วย = รวมเงิน ÷ จำนวน; แก้จำนวนหรือต้นทุน → รวมเงิน = จำนวน × ต้นทุน; ผลรวมกริด → จำนวนเงินท้ายจอ (`DetailCalc` `icfrmstockbalance.cpp:1173-1206`, `CalSumAmount` `:1207`) |
| คอลัมน์เงา | RowOrder, MyType, MyDescription, OppositeQty, OppositeUnit, BarCode, LotScript(ล็อต), … BehindIndex, รหัสกลุ่ม/ชนิด/ยี่ห้อ/รูปแบบ/สี/ขนาด/สไตล์ ของสินค้า |
| ตรวจก่อนบันทึก | ปิดงวดแล้ว → ไม่ให้บันทึก · เอกสารยืนยัน/ยกเลิกแล้ว (`CANNOT_SAVE_DOCUMENTCONFIRM` / `CANNOT_SAVE_DOCCANCEL`) · เลขที่เอกสารว่าง (`CANNOT_SAVE_NO_DOCNO`) · ทุกบรรทัดที่มีรหัสสินค้าต้องมี คลัง + ที่เก็บ + หน่วยนับ (`CANNOT_SAVE_NEEDDETAIL`) · สินค้า Serial (`StockType=2` และตั้งค่าบริษัท `SerialItem=1`): จำนวน Serial ต้องเท่ากับจำนวน; แก้เอกสารแล้วลบ serial ที่ขาย/เบิกไปแล้วไม่ได้ (`NOTEDIT_SERIALSTATUS`) และเปลี่ยนคลังของ serial ไม่ได้ (`CANNOT_EDIT_SERIALWH`) · สินค้าต้นทุนแบบล็อต (CostType=1) ต้องระบุ Lot เมื่อตั้งค่า `nLotStatus=1` และ `nUseLotStatus=1` (`NOT_SET_LOT`) · เลขที่เอกสารซ้ำ → ถามลองใหม่ (`icfrmstockbalance.cpp:123-262`) |
| บันทึกลง | `BCStkBalance` (DocNo, DocDate, DepartCode, SumOfAmount, MyDescription, IsCompleteSave, Creator/CreateDateTime หรือ LastEditor/LastEditDateT) + `BCStkBalanceSub` (DocNo, DocDate, DepartCode, ItemCode, WHCode, ShelfCode, UnitCode, Qty, Cost, Amount, MyDescription, OppositeQty/Unit, BarCode, LotNumber/LotExpireDate/LotMyDesc, Category/Group/Brand/Type/Format/Color/Size/Style Code, TaxRate=0, PackingRate1/2 (ค่าเริ่ม 1.0), behindindex, IsProcess, LineNumber) (`:368-415,924-1065`) |
| ผลข้างเคียง | **ไม่มีการลงบัญชี GL** (ฟอร์มไม่มีฟิลด์ GLFormat) · ไม่เรียก process รายเอกสาร แต่เรียก `OnStockReprocess(item, docDate)` ต่อทุกสินค้า → `INSERT ProcessStock(ItemCode, DocDate, ProcessFlag=1, FlowStatus=1)` (`mytools.cpp:2750-2757`) เพื่อให้ Robot คิดสต็อก/ต้นทุนของสินค้านั้นใหม่ตั้งแต่วันที่เอกสาร · บันทึก Serial/Lot (`OnSaveSerial(0,…)`, `OnSaveLotItem(0,…)`) · log book (`:1043-1057`) |
| ยืนยัน/ยกเลิก | `BillConfirm("BCStkBalance", docNo)` / `BillCancel("BCStkBalance", docNo)` (`:1508-1522`); ยืนยันได้เมื่อมีเลขที่เอกสาร (`NOT_CONFIRM_NO`) |
| ปุ่มลัด | F5 = อื่น ๆ, F8, F9 = Serial/Lot, Enter, PgDn (`:514-589`) |

### 4.2 5601 "บันทึกใบขอเบิกใช้สินค้า,วัตถุดิบ" (`CICFrmStkIssueRequst`, `IDD_IC_FRM_STOCKISSUE_REQ` 339×175; atlas-07:511)

เอกสารขอเบิก **ไม่กระทบสต็อกและไม่ลง GL** (บันทึกลง `BCStkIssue2`/`BCStkIssueSub2` เท่านั้น; ไม่เรียก process).

| ส่วน | รายละเอียด |
|---|---|
| หัว (ลำดับบนจอ) | เลขที่เอกสาร · วันที่ · รหัสแผนก (BCDepartment) · การจัดสรร (BCAllocate) · โครงการ (BCProject) · วันที่นำส่งคืน (`ED_DUE_DATE`) · ประเภทการเบิก (`ED_ISSUETYPE` ← BCIssueType) · จำนวนวันที่เบิก (`ED_DAY_OF_USE`) · หมายเหตุ · ลูกหนี้/ลูกค้า (`ED_DEMO` ลิงก์, ป้ายใน rc ซ้ำกับ "วันที่นำส่งคืน" — ยังไม่ verify) · ผู้เบิก (`ED_PERSONCODE` ← BCSale) · จำนวนเงิน (`ED_SUM_OF_AMOUNT`) |
| กริด | รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · จำนวน · ต้นทุน/หน่วย («UNITCOST») · รวมต้นทุน («COST_SUM») (`ICFrmStkIssueRequst.cpp:491-498`) |
| บันทึกลง | `BCStkIssue2` (หัว) + `BCStkIssueSub2` (… `IssueType`, `Qty`, **`RemainQty` = จำนวนที่ยังไม่เบิก**, `Price`, `Amount`, DepartCode/AllocateCode/ProjectCode ต่อบรรทัด) (`:751-822`) |
| ตรวจ | ปิดงวด/ยืนยัน/ยกเลิก/เลขที่ว่าง/ซ้ำ เหมือน 4.1; **แก้ไม่ได้ถ้าถูกใบเบิกอ้างอิงแล้ว** (`CANNOT_EDIT_USE`, `:154`) |
| ใช้ต่อ | ปุ่ม "ใบสั่งซื้อสินค้า"/PreIssue ในใบเบิก (4.3) ดึงรายการจากใบขอเบิก (การลด `RemainQty` เมื่อเบิกแล้ว ยังไม่ verify) |
| ปุ่มลัด | F5, F12 (`:431-463`) |

### 4.3 5625 "บันทึกเบิกใช้สินค้า,วัตถุดิบ" (`CICFrmStockIssue`, `IDD_IC_FRM_STOCKISSUE` 443×245; atlas-07:575) — เอกสารต้นแบบของกลุ่มสต็อก

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร · วันที่ · รหัสแผนก (BCDepartment) · การจัดสรร (BCAllocate) · โครงการ (BCProject) · วันที่นำส่งคืน · ประเภทการเบิก (BCIssueType) · จำนวนวันที่เบิก · ผู้เบิก (BCSale) · หมายเหตุ · ลูกหนี้ (ลิงก์) · ปุ่ม "ใบสั่งซื้อสินค้า"/PreIssue (ดึงจากใบขอเบิก 4.2) · ปุ่ม "ตรวจสอบสต๊อกย้อนหลัง" · ผลรวมจำนวนเงิน (`SumOfAmount`) |
| กริด (โหมดทั่วไป) | รหัสสินค้า (แก้ได้+ลิงก์) · ชื่อสินค้า · คลัง (combo จาก BCItemWarehouse ของสินค้านั้น) · ที่เก็บ · หน่วยนับ · จำนวน («ITEMQTY») · ราคาขาย/ราคาต่อหน่วย («PERPRICE», ตามคอมเมนต์ "ราคาขาย") · จำนวนเงิน («AMOUNT» = "จำนวนเงิน:") · ต้นทุน/หน่วย («UNITCOST», อ่านอย่างเดียว, ระบบคิด) · รวมต้นทุน («COST_SUM», อ่านอย่างเดียว) · จำนวนที่รับคืนแล้ว («STKREFUND_NUMBER», อ่านอย่างเดียว) — เมื่อใช้งานแบบงาน (Job) เพิ่มคอลัมน์ "JOBNO" นำหน้า (`icfrmstockissue.cpp:140-330`) |
| คอลัมน์เงา (0-29) | RowOrder, OppositeQty, OppositeUnit, BarCode, DepartCode, AllocateCode, ProjectCode, MyDescription, Category/Group/Brand/Type/FormatCode, MyType, RefNo, LotScript, RefAmt, StockType, LineNumber, RefLineNumber, OldPrice/OldAmount/OldQty, LotNumber, OldUnitCode, Color/Size/StyleCode, BehindIndex, StockProcess |
| ปุ่มลัด (`:1277-1352`) | F2 (Ctrl+F2 = ตรวจสต็อก), F5 = อื่น ๆ, F8 = มัดจำ, F9 (Ctrl/Shift = Serial/Lot), Enter, PgDn, Home |
| ตรวจก่อนบันทึก (`ConfirmSaveData` `:527-743`) | ปิดงวด · ผ่านบัญชีแล้วแก้ไม่ได้ (`CANNOT_SAVE_POSTGL`) · เลขที่ว่าง · ยืนยัน/ยกเลิกแล้ว · **ต้นทุนแบบ FIFO ต้องระบุ Lot และจำนวน Lot ≥ จำนวนเบิก** (เมื่อ `nLotStatus=1`) · ห้ามอ้างอิงสถานะบิล (`CANNOT_SAVE_BILLSTATUS_REF`) · จำนวนต้อง > 0 (`PUT_NOT_QTY`) · คลัง/ที่เก็บ/หน่วยครบ · Serial ครบและเท่าจำนวน · **ตรวจสต็อกไม่พอเมื่อ `CompanyInfo.StockStatus = 2`** (ยอดคงเหลือ < จำนวนเบิก → บล็อก "STOCK_NOT_ENOUGHLIST"; `StockStatus = 1` ~ เตือนแล้วให้ไปต่อ, 0 = ไม่ตรวจ — ค่า 1 ยังไม่ verify ในใบเบิก แต่ใบโอนใช้เตือน Yes/No `icfrmstocktransfer.cpp:184-198`) · ต้องมีรูปแบบลงบัญชี (GLFormat) · เลขที่ซ้ำ → ถามลองใหม่ |
| บันทึกลง (`OnSaveData` `:845-1132`) | `BCStkIssue` (หัว 38 คอลัมน์ รวม DocNo, DocDate, DepartCode, AllocateCode, ProjectCode, IssueType, DueDate, DayOfUse, PersonCode, ArCode, SumOfAmount, SumOfCost, GLFormat, GLTransData, GLStartPosting, IsCompleteSave, IsCancel, …) + `BCStkIssueSub` (55 คอลัมน์: ItemCode, WHCode, ShelfCode, UnitCode, Qty, Price, Amount, **SumOfCost**, TaxRate=0, PackingRate1/2 ค่าเริ่ม 1.0, BehindIndex, IsProcess, TransState, RefNo/RefLineNumber, Lot*, …) → `UPDATE IsCompleteSave=1` → Commit → **`INSERT processdoc(DocNo, DocType=8)`** → ถ้า `AutoServerProc ≠ 1` เรียก `StkItemIssueProcess(docNo)` (ตัดสต็อก/ต้นทุน หัวข้อ 5.2) → GL (5.5) → `OnSaveSerial(8,…)`, `OnSaveLotItem(8,…)`, `OnStockReprocess` ต่อสินค้าที่มีวันที่ย้อนหลัง → log book |
| GL | รูปแบบบัญชีเริ่มต้นกลุ่ม **24**; ตั้ง `m_GLTLSProcess.m_nSource=21`, `m_nGLGroup=13`; ถ้า GLTransData ≠ 0 (ลงมือ) เปิดจอ F2 ผ่าน `OnSaveAccount(...,21,...)` (`:744-747,1096-1106`) — เงื่อนไข GL ดูหัวข้อ 5.5 |
| ยืนยัน/ยกเลิก | `BillConfirm("BCStkIssue", docNo)` / `BillCancel("BCStkIssue", docNo)` |

### 4.4 5626 "บันทึกรับคืนสินค้า,วัตถุดิบ" (`CICFrmStockIssueRet`, `IDD_IC_FRM_STOCKISSUERET` 339×175; atlas-07:885)

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร · วันที่ · **เลขที่ใบเบิก** (`ED_ISSUE_REF_NO` ลิงก์ → BCStkIssue; ~เลือกแล้วดึงรายการจากใบเบิก — ยังไม่ verify) · รหัสแผนก · โครงการ · การจัดสรร · ผู้รับ/ผู้เบิก (`ED_PERSONCODE` ← BCSale) · หมายเหตุ |
| กริด | รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · จำนวน · ต้นทุน/หน่วย («UNITCOST») · ต้นทุนรับคืนรวม («RETURN_COST», ตามคอมเมนต์ "ต้นทุนรับคืนเบิกรวม") · (ถ้าไม่ใช่โหมด Chm) จำนวนเงิน («AMOUNT»); คอลัมน์ต้นทุนแก้ไม่ได้ถ้าไม่มีสิทธิ์เห็นต้นทุน (`icfrmstockissueret.cpp:703-719`) |
| บันทึกลง | `BCStkIssueRet` (32 คอลัมน์) + `BCStkIssRetSub` (56 คอลัมน์: … `IssueRefNo`, `RefLineNumber` (บรรทัดของใบเบิก), `Price`, `ReturnCost`, `Amount`, `HomeAmount`, TaxRate=0, PackingRate1/2) (`:402-466,1253-1433`) |
| ต้นทุนรับคืน | ต้นทุนมาจากบรรทัดใบเบิกที่อ้างอิง (ไม่ใช่ต้นทุนเฉลี่ยปัจจุบัน — หัวข้อ 5.2) |
| ผลข้างเคียง | `INSERT processdoc(DocNo, DocType=9)` → `StkItemIssueRetProcess` (เว้น AutoServerProc=1) · GL: รูปแบบเริ่มต้นกลุ่ม **25** (`GetDefaultGLFormat(25)`), `m_nSource=22`, `m_nGLGroup=26` (`:247,268-271,1388-1412`) · `OnSaveSerial(9,…)`/`OnSaveLotItem(9,…)` · คอลัมน์ `STKREFUND_NUMBER` ในใบเบิกแสดงจำนวนที่รับคืนแล้ว (~ยังไม่ verify จุดอัปเดต) |

### 4.5 5627 "บันทึกรับสินค้าสำเร็จรูป" (`CICFrmFinishGood`, `IDD_IC_FRM_FINISHGOOD` 341×250; atlas-07:954)

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร · วันที่ · รหัสแผนก · การจัดสรร · โครงการ · ผู้รับ (`ED_PERSONCODE` ← BCSale) · หมายเหตุ · (ท้ายจอ) จำนวนเงิน (`ED_AMOUNT1`) |
| กริด | รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · จำนวน · ต้นทุน («COST_ITEM», ตามคอมเมนต์ "ต้นทุน") · มูลค่า («SUM_TOTAL», ตามคอมเมนต์ "มูลค่า") (`icfrmfinishgood.cpp:659-666`) — **ต้นทุนของสินค้าสำเร็จรูปผู้ใช้เป็นคนกรอก** (ไม่มีสูตรประกอบ BOM ในเอกสารนี้) |
| บันทึกลง | `BCFinishGoods` (33 คอลัมน์) + `BCFinishGoodsSub` (52 คอลัมน์: Qty, Cost, Amount, …) (`:364-430,1066-1228`) |
| ผลข้างเคียง | `processdoc DocType=1` → process รับสินค้าสำเร็จรูป (MyType 1 +) · GL: รูปแบบเริ่มต้นกลุ่ม **26**, `m_nSource=23`, `m_nGLGroup=27` (`:251,462,1191-1212`) |
| ปุ่มลัด | F2, F5 (`:542-625`) |

### 4.6 5631 "บันทึกใบขอโอนสินค้า" (`CICFrmStkTransferRequest`, `IDD_IC_FRM_STOCKTRANSFER_REQ` 339×211; atlas-07:1025)

เอกสารขอโอน **ไม่กระทบสต็อก/ไม่ลง GL** (`BCStkTransfer2` / `BCStkTransfSub2`).

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร · วันที่ · รหัสแผนก (BCDepartment) · หมายเหตุ · กรอบ "คลังและที่เก็บ": จากคลัง (`ED_FROM_WH` ← BCWarehouse) · ถึงคลัง (`ED_TO_WH`) · จากที่เก็บ (`ED_FROM_SHELF` ← BCShelf) · ถึงที่เก็บ (`ED_TO_SHELF`) (ค่าตั้งต้นของทุกบรรทัด) · ปุ่ม "ใบสั่งซื้อสินค้า" (`BTTN_PROCESS` = ปุ่มประมวลผล/เติมรายการ ~ ป้าย rc คัดลอกมา ยังไม่ verify) · จำนวนเงิน |
| กริด | รหัสสินค้า · ชื่อสินค้า · **WH//O** (คลังต้นทาง) · **SH//O** (ที่เก็บต้นทาง) · **WH//I** (คลังปลายทาง) · **SH//I** (ที่เก็บปลายทาง) · หน่วยนับ · จำนวน (`ICFrmStkTransferRequest.cpp:507-514`; ป้ายหัวคอลัมน์เป็นข้อความอังกฤษตายตัว `WH//O` ฯลฯ) |
| บันทึกลง | `BCStkTransfer2` + `BCStkTransfSub2` (FromWh, FromShelf, ToWh, ToShelf, UnitCode, Qty, **RemainQty = Qty** ตอนสร้าง, SumOfCost, Price, Amount, …) (`:267-390`) |
| ตรวจ | เอกสารยืนยัน/ยกเลิกแล้ว · เลขที่ว่าง · อ้างอิงสถานะบิล (`CANNOT_SAVE_BILLSTATUS_REF`) · รายการครบ (`PLEASE_CHECKDATA`) · ถ้า `StockStatus = 2` และจำนวนขอโอนเกินคงเหลือต้นทาง → `CANNOTTRANS` (บล็อก) |

### 4.7 5628 "บันทึกโอนสินค้าระหว่างคลัง" (`CICFrmStockTransfer`, `IDD_IC_FRM_STOCKTRANSFER` 346×234; atlas-07:1085)

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร · วันที่ · เขตการขาย (`EDIT_SALEAREA`) · รหัสแผนก · พนักงานขาย (`EDIT_SALECODE`) · หมายเหตุ · ปุ่ม **"ใบขอโอน"** (`BTTN_PRETRANS` ดึงจากใบขอโอน 4.6) · กรอบ "คลังและที่เก็บ": จากคลัง · ถึงคลัง · จากที่เก็บ · ถึงที่เก็บ · ปุ่ม "พิมพ์บาร์โค้ด" · ปุ่มประมวลผล (`BTTN_PROCESS`) · จำนวนเงิน |
| กริด | รหัสสินค้า · ชื่อสินค้า · WH//O · SH//O · WH//I · SH//I · หน่วยนับ · จำนวน · «Print_Barcode» (จำนวนดวงบาร์โค้ดที่จะพิมพ์) · ☑ "S" (เลือกพิมพ์) · ☑ "P" (พิมพ์ราคา) (`icfrmstocktransfer.cpp:919-929`) |
| บันทึกลง | `BCStkTransfer` (23 คอลัมน์) + `BCStkTransfSub` (บรรทัดเอกสาร) **และ `BCStkTransfSub3` (42 คอลัมน์: ลูกคู่ "ออก/เข้า" ที่ process ใช้ — fromwhcode/fromshelfcode/reftrlinenum/MyAmount/WHCode/ShelfCode)** — ลบ Sub3 ของเลขเก่าก่อนบันทึกใหม่ (`:506-575,576-725`) |
| ตรวจ | ยืนยัน/ยกเลิกแล้ว · ถูกแก้ไม่ได้ (`CANNOT_EDIT_TRANS`) · เลขที่ว่าง · รายการครบ (`PLEASE_CHECKDATA`) · **`StockStatus ≠ 0`: ถ้าสต็อกต้นทางไม่พอ → ถาม Yes/No (`STOCK_NOT_ENOUGH`); `StockStatus = 2`: บล็อก** แสดง "ยอดรวม … รหัสสินค้า … ยอดคงเหลือ … (STOCK_NOT_ENOUGHLIST)" · Serial: เปลี่ยนคลัง → ถามยืนยัน "ย้าย Serial ไปคลังใหม่" (`CHANGWH_NEWSERIAL`) · Lot: ต้องระบุ Lot ออก (`LotOutNumber`) และจำนวนพอ |
| ผลข้างเคียง | `StkItemTransferProcess(docNo)` (`:687`) → ตัดออกจากต้นทาง/เพิ่มเข้าปลายทางด้วยต้นทุนเดียวกัน (หัวข้อ 5.2) · **ไม่มี GL** (ฟอร์มไม่มีฟิลด์ GL) · `OnStockReprocess` ทั้งต้นทาง/ปลายทาง |

### 4.8 5629 "บันทึกเอกสารเพื่อตรวจนับสินค้า" (`CICFrmStockInspect`, `IDD_IC_FRM_STOCKINSPECT` 346×233; atlas-07:1163)

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสาร · วันที่เอกสาร · หมายเหตุการตรวจนับ (`ED_BILL_DES`) · หมายเหตุผลการตรวจนับ (`ED_RESULT_DES`) · กรอบ **"รายการอัตโนมัติ"**: จากคลัง–ถึงคลัง · จากสินค้า–ถึงสินค้า · จากที่เก็บ–ถึงที่เก็บ · เงื่อนไข (`ED_CONDITION` + ปุ่ม "…" นำเข้าจากไฟล์ข้อความ `IMPORT_TEXT`) · การตรวจโดย (`ED_INSPECT_BY` combo) · ปุ่ม **"ประมวลผล"** (`BTTN_GET_INSPECT` เติมกริดจากยอดคงเหลือตามช่วงที่เลือก) · ☐ "แสดงยอดคงเหลือตามหน่วยนับ" (`CHK_UNIT`) |
| กริด | รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · จำนวน («QTY» = ยอดคงเหลือ ณ ตอนออกเอกสาร) (`ICFrmStockInspect.cpp:155-160`) |
| บันทึกลง | `BCStkInspect` (18 คอลัมน์) + `BCStkInspectSub` (21 คอลัมน์: DocNo, DocDate, InspectBy, ItemCode, WHCode, ShelfCode, UnitCode, **StkQty**, **InspectQty** (ว่าง/คำนวณ), **DiffQty**, **isexport = 0** (ธงส่งออกให้เครื่องตรวจนับ BCDTS), LineNumber, AdjustStatus, InspectDesc) (`:280-330,814-865`) |
| ผลข้างเคียง | **ไม่กระทบสต็อก / ไม่ลง GL**; ใช้พิมพ์ใบตรวจนับ (รายงาน 5652/5653) และเป็นฐานให้ 5599 กรอกผลนับ |
| ตรวจ | ยืนยันแล้วบันทึกไม่ได้ · เลขที่ว่าง/ซ้ำ |
| ปุ่มลัด | F5, F12 (`:406-431`); ยืนยัน: `BillConfirm("BCStkInspect", docNo)` |

### 4.9 5599 "บันทึกผลการตรวจนับสินค้า" (`CICFrmStockInspectResult` : `CMyBaseFormView`, `IDD_IC_FRM_STOCKINSPECT_RESULT` 390×210; atlas-07:1258)

จอนี้ **ไม่ใช่เอกสารใหม่** — เลือกเอกสารตรวจนับเดิมแล้วกรอกผลนับลงบรรทัดเดิม.

| ส่วน | รายละเอียด |
|---|---|
| หัว | เลขที่เอกสารการตรวจนับ (`ED_DOC_NO` ลิงก์ → BCStkInspect; ดึงหัว+รายการมาแสดง) · วันที่เอกสาร · หมายเหตุการตรวจนับ · หมายเหตุผลการตรวจนับ · ปุ่ม **"ตรวจสอบสต๊อกย้อนหลัง"** (`CHECK_STOCKREWOAD`) · ปุ่ม (`BTTN_UPDATE_DATA` ป้าย "ประม…" ถูกตัดใน atlas ~ ประมวลผล/บันทึกผล) |
| กริด | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · คลัง · ที่เก็บ · จำนวน (ยอดในระบบ, อ่านอย่างเดียว) · «STKCHKQTY» **จำนวนที่นับได้ (แก้ได้)** · «VARSTKCHKQTY» **ผลต่าง (แก้ได้ — ระบบคิด = นับได้ − ในระบบ)** · หมายเหตุ (`ICFrmStockInspectResult.cpp:111-119`) |
| บันทึก (`OnUpdateData` `:427-473`) | `UPDATE BCStkInspect SET ConfirmCode, ConfirmDateTime` (ผู้ยืนยัน + เวลา) แล้วต่อบรรทัด `UPDATE BCStkInspectSub SET InspectQty, DiffQty, StkQty, InspectDesc, AdjustStatus = 0 WHERE DocNo AND RowOrder`; บันทึก log book 88001 |
| ตรวจสต๊อกย้อนหลัง (`OnBttnCheckStockReword` `:510-600`) | ถ้าวันที่เอกสารตรวจนับ < วันนี้ → คำนวณยอดคงเหลือ ณ วันนั้นด้วยตารางชั่วคราว `FunctionStock` / `FunctionStockResult` (คีย์ SessionId = "88001"+เลขที่เอกสาร) แล้วแสดง (`CHECKSTOCK_TURNBACK`); ปิดปุ่มเมื่อตั้งค่า `IsNoStockReword=1` |

### 4.10 5630 "บันทึกปรับปรุงสินค้าหลังตรวจนับ" (`CICFrmStockAdjust`, `IDD_IC_FRM_STOCKADJUST` 339×175; atlas-07:1270)

| ส่วน | รายละเอียด |
|---|---|
| หัว (ลำดับบนจอ) | เลขที่เอกสาร · วันที่ · **เลขที่เอกสารการตรวจนับ** (`ED_INSPECT_NO` ลิงก์ → BCStkInspect; ดึงรายการผลต่างมา) · **ประเภทการปรับปรุง** (`ED_ADJUST_TYPE` combo 4 ค่า) · หมายเหตุการปรับปรุง · ☐ «TAXCASE» (= `OutputTaxStatus`, โหมด Chm ติ๊กไว้ล่วงหน้า) · ☐ "แสดงเฉพาะรายการที่ปรับปรุง" |
| ประเภทการปรับปรุง (`AdjustType`) | 0 «OVERSTK» สินค้าเกิน (+ เท่านั้น) · 1 «BREAKSTK» สินค้าชำรุด (− เท่านั้น) · 2 «LOSTSTK» สินค้าสูญหาย (− เท่านั้น) · 3 «ADJUSTITEM» ปรับปรุงสินค้า (+ หรือ −) — ความหมายไทย ~ตามชื่อ key ยังไม่ verify (`ICFrmStockAdjust.cpp:748-758,173-178`) |
| ยอดรวมท้ายจอ | จำนวนเงิน (`SumOfAmount`) · ยอดยกเว้นภาษี (`SumOfExceptTax`) · ภาษี (`TaxAmount`) · ยอดรวม 3 ช่อง: `SumAmount` (ช่อง `m_edPositiveTotal`), `SumAmount2` (ช่อง `m_edTotal`), `TotalAmount` (ช่อง `m_edTotalAmount`) (`OnSaveData` `:425-470`) — ตอนลง GL เงื่อนไข 1 อ่าน `SumAmount` เป็น "ยอดรวมต้นทุนปรับปรุงลบ" และเงื่อนไข 2 อ่าน `SumAmount2` เป็น "ยอดรวมต้นทุนปรับปรุงบวก" (คอมเมนต์ `gltlsprocess.cpp:9132-9136`) — ชื่อช่องบนฟอร์มขัดกับคอมเมนต์ GL **ยังไม่ verify ว่าช่องไหนเก็บบวก/ลบ** |
| กริด | รหัสสินค้า · ชื่อสินค้า · คลัง · ที่เก็บ · หน่วยนับ · จำนวนปรับปรุง («ADJUSTEDQTY») · ต้นทุนปรับปรุง («ADJUSTEDCOST», ตามคอมเมนต์ "ต้นทุนปรับปรุง") · รวมเงิน («TOTALS») · **+/-** (เครื่องหมาย 1 ตัวอักษร) · ราคา/หน่วย («UNITPRICE») · มูลค่า («SUM_TOTAL») (`:785-795`) |
| กฎบันทึก | ทุกบรรทัดที่จำนวน > 0 **ต้องระบุ +/-** และต้องตรงประเภท (0: `+`, 1/2: `-`, 3: `+` หรือ `-`) ไม่งั้น `CANNOT_SAVE_NOSPECIFY_UPDATE`; ต้องมี คลัง+ที่เก็บ+หน่วย; ถ้า `StockStatus = 2` ปรับลบเกินคงเหลือไม่ได้; Lot ต้องระบุเมื่อตั้งค่า (`:152-345`); ผ่านบัญชีแล้วแก้ไม่ได้ |
| บันทึกลง | `BCStkAdjust` (37 คอลัมน์: DocNo, DocDate, InspectNo, AdjustType, TaxAmount, TotalAmount, SumOfExceptTax, SumAmount2, SumAmount, OutputTaxStatus, GLTransData, GLFormat, GLStartPosting, IsCompleteSave, MyDescription …) + `BCStkAdjustSub` (47 คอลัมน์: **AdjMark** 0 = บวก (`MyType=10`), 1 = ลบ (`MyType=11`); InspectNo, **InspectRow** (อ้างบรรทัด BCStkInspectSub), Qty, AdjCost, Amount, Price, NetAmount, ExceptTax, SumOfCost, Lot*, IsProcess …) (`:1172-1365`) |
| เชื่อมใบตรวจนับ | บันทึก/แก้: `UPDATE BCStkInspectSub SET AdjustStatus = 0` ของบรรทัดที่เอกสารปรับปรุงเดิมอ้างอิง แล้วบันทึกใหม่ (ตั้ง AdjustStatus ตามการปรับ) + `UpdateDocumentFlow(-1, 10, inspectNo)` (`:489-495`) |
| ผลข้างเคียง | `INSERT processdoc(DocNo, DocType=10)` → process ปรับปรุง (บวก/ลบตาม AdjMark, MyType 10/11) · **GL**: รูปแบบเริ่มต้นตามประเภท — 0 → กลุ่ม **34**, 1 → **44**, 2 → **45**, 3 → **46** (`GetGLFormat` `:2437-2448`); `m_nSource=34`, `m_nGLGroup=34`, `OnSaveAccount(...,34,...)` (`:1305-1330`) |
| ปุ่มลัด | (ชุดเอกสาร: F2 = GL, F5 …) |

### 4.11 ผังความสัมพันธ์ของเอกสารสต็อก

```
ใบขอเบิก 5601 (BCStkIssue2) ──ดึง──▶ ใบเบิก 5625 (BCStkIssue) ──อ้าง IssueRefNo──▶ ใบรับคืนเบิก 5626 (BCStkIssueRet)
ใบขอโอน 5631 (BCStkTransfer2) ──ดึง──▶ ใบโอน 5628 (BCStkTransfer + Sub + Sub3)
เอกสารตรวจนับ 5629 (BCStkInspect) ──กรอกผล──▶ 5599 (BCStkInspectSub.InspectQty/DiffQty) ──ดึง──▶ ใบปรับปรุง 5630 (BCStkAdjust, InspectRow)
ยกมา 5624 (BCStkBalance) · รับสินค้าสำเร็จรูป 5627 (BCFinishGoods) — เอกสารเดี่ยว
```


---

## 5. ประมวลผลสต็อก · ต้นทุน · ยอดงวด · ลงบัญชี GL

> ที่มาหลักของหัวข้อนี้: เครื่องยนต์สต็อกของ Champ อยู่ใน **`champ-lib\BC5PROCESS\bcdumpprocjob.cpp`** (ฟังก์ชัน `CBCDumpProcJob::StockCenterProcess` บรรทัด 2473-3323) และการลงบัญชีอยู่ใน **`champ-lib\BC5PROCESS\gltlsprocess.cpp`**. ฟอร์มเอกสารสต็อกไม่คำนวณต้นทุนเอง — แค่บันทึกแล้วสั่งให้เครื่องยนต์ทำ (หัวข้อ 0.1).

### 5.1 ลำดับการทำงานหลังกด "บันทึก" เอกสารสต็อก

| ขั้น | เกิดอะไรขึ้น | หลักฐาน |
|---|---|---|
| 1 | ฟอร์มบันทึกหัว + รายการย่อย (`TransState=0` = รายการใหม่, แถวเดิมที่ถูกแก้ตั้ง `TransState=1` = "รายการเก่า") แล้วตั้ง `IsCompleteSave = 1` | หัวข้อ 0.1 |
| 2 | ลงคิวงาน: `INSERT processdoc(DocNo, DocType, ProcessFlag, guid)` (ถ้า `AutoServerProc = 1` ให้ Robot/เซิร์ฟเวอร์รับไปทำ ไม่งั้นฟอร์มเรียก `StkItem*Process` เอง) | `bcdumpprocjob.cpp` (ตาราง `ProcessDoc` ใน atlas.json: RowOrder, DocNo, DocType, ProcessFlag, guid) |
| 3 | เครื่องยนต์ดึงแถวรายการย่อยของเอกสารนั้น (เก่า → ย้อนด้วย `nExMark = -1`, ใหม่ → `nExMark = +1`) แล้ววนทีละแถวเข้า `StockCenterProcess` | `bcdumpprocjob.cpp:2473-2513, 4483-4508` |
| 4 | ต่อแถว: โหลดข้อมูลสินค้า (`LoadItemInfo`: `CostType`, `UnitType`, `StockType`, `AverageCost`, `StandardCost`, `StockQty`, `Amount`, `DefStkUnitCode`) → คำนวณต้นทุน → ปรับยอดสินค้า/คลัง/ยอดงวด → เขียน `SumOfCost` กลับไปที่แถวเอกสาร | `:2523-3266`, `LoadItemInfo` `:5773-5876` |
| 5 | ถ้ามี GLFormat และไม่ใช่ลงบัญชีมือ → ลงบัญชี GL อัตโนมัติ (หัวข้อ 5.6) | หัวข้อ 0.1 |
| ทางลัดสำหรับ "คำนวณใหม่" | `OnStockReprocess(ItemCode, DocDate)` = `INSERT ProcessStock(ItemCode, DocDate, ProcessFlag=1, FlowStatus=1)` — ใช้เมื่อเอกสารย้อนหลังกระทบสินค้านั้น (เช่น สินค้าชุดประกอบ/แยก, ปรับลำดับรายวัน) ให้เซิร์ฟเวอร์คำนวณสินค้านั้นใหม่ตั้งแต่วันที่นั้น | `champ\champ\mytools.cpp:2750-2757` |

ข้ามสินค้าที่ **`StockType = 1` (บริการ)**: เครื่องยนต์ไม่ปรับยอดสต็อก/ต้นทุน/ยอดงวดของสินค้าบริการ (`if(nStockType != 1)` `:2743, :3102`); รหัสสินค้า `"."` (แถวคำอธิบาย) ถูกข้าม (`:2543`).

### 5.2 ตารางแปลงชนิดเอกสาร → ผลกับสต็อก (`MyType`) — `bcdumpprocjob.cpp:2600-2681`

`nMark` = ทิศทางของสต็อก (+1 เข้า / −1 ออก) แล้วคูณกับ `nExMark` (+1 ปกติ / −1 ย้อนกลับ) → `nMark = nMark × nExMark`. "ช่องจำนวน/จำนวนเงิน" = ชื่อคอลัมน์ในตารางรายการย่อยที่เครื่องยนต์อ่านเป็นจำนวนและมูลค่า (ค่าเริ่มต้น `Qty` / `HomeAmount`).

| MyType | ชนิดรายการ | ตารางรายการย่อยที่อ่าน | nMark | ช่องจำนวน | ช่องจำนวนเงิน | ปรับ `SumOfCost` ของแถวเอกสารด้วย | Lot (ตั้งยอด Lot เมื่อเป็นต้นทุน FIFO/Lot) |
|---|---|---|---|---|---|---|---|
| 0 | ยกยอดมา | `BCStkBalanceSub` | +1 | Qty | Amount | ไม่ | ใช่ |
| 1 | รับสินค้าสำเร็จรูป | `BCFinishGoodsSub` | +1 | Qty | Amount | ใช่ | ใช่ |
| 2 | ซื้อ | `BCApInvoiceSub` | +1 | Qty | HomeAmount (+ ต้นทุนแฝง `SumOfExpCost` ถ้า `UseExpCost=1`) | ใช่ | ใช่ |
| 3 | เพิ่มหนี้เจ้าหนี้ | `BCDebitNoteSub2` | +1 | DiscQty | HomeAmount | ใช่ | ใช่ |
| 4 | ขาย | `BCARInvoiceSub` | −1 | Qty | HomeAmount | ใช่ | — |
| 5 | รับคืน/ลดหนี้ลูกหนี้ | `BCCreditNoteSub` | +1 | DiscQty | HomeAmount | ใช่ | ใช่ |
| 6 | เพิ่มหนี้ลูกหนี้ | `BCDebitNoteSub1` | −1 | DiscQty | HomeAmount | ใช่ | — |
| 7 | ส่งคืน/ลดหนี้เจ้าหนี้ | `BCStkRefundSub` | −1 | DiscQty | HomeAmount | ใช่ | — |
| 8 | เบิก | `BCStkIssueSub` | −1 | Qty | SumOfCost | ใช่ | — |
| 9 | รับคืนจากเบิก | `BCStkIssRetSub` | +1 | Qty | SumOfCost | ใช่ | ใช่ |
| 10 | ปรับปรุงเกิน (+) | `BCStkAdjustSub` (AdjMark=0) | +1 | Qty | Amount | ใช่ | ใช่ |
| 11 | ปรับปรุงขาด (−) | `BCStkAdjustSub` (AdjMark=1) | −1 | Qty | Amount | ใช่ | — |
| 12 | โอนเข้า | `BCStkTransfSub3` | +1 | Qty | MyAmount | ใช่ | — |
| 13 | โอนออก | `BCStkTransfSub3` | −1 | Qty | MyAmount | ใช่ (ผ่าน `StkItemTransferProcess2`) | — |
| 20 | รับคืนสินค้าสำเร็จรูป | (ไม่ระบุตาราง) | −1 | Qty | Amount | ไม่ | — |

หมายเหตุสำคัญ:
- **เอกสารโอน** แยกเป็น 2 แถวใน `BCStkTransfSub3` (โอนออก 13 / โอนเข้า 12) — ฝั่งโอนออกคำนวณต้นทุนก่อน แล้วส่งต้นทุนไปให้ฝั่งโอนเข้าผ่าน `StkItemTransferProcess2(RowOrder, ต้นทุน, CostType)`; ยอดของ **BCItem ไม่ถูกแตะเมื่อ MyType = 13** (`:3171`) เพราะโอนออก+เข้าหักล้างกัน; `BCStkWarehouse` ของโอนออกถูกปรับภายใน `StkItemTransferProcess2` ไม่ใช่ใน `UpdateStockWarehouse` ตรง ๆ (`:3119, :3134-3140`).
- **ใบขอเบิก / ใบขอโอน / เอกสารตรวจนับ / ผลตรวจนับ ไม่มี MyType** = ไม่กระทบสต็อก (หัวข้อ 4.2, 4.6, 4.8, 4.9). มีแต่ใบปรับปรุง (10/11) ที่ตัดสต็อกจริง.
- ใบสั่งขาย (SO) ปรากฏในเครื่องมือจัดลำดับรายวัน/ล็อกต้นทุนเป็น `MyType = 20` (`ICDlgUpdateBehindIndex.cpp:200`, `ICDlgLockCostDlg.cpp:160`) ทั้งที่ในเครื่องยนต์ 20 = "รับคืนสินค้าสำเร็จรูป" — เป็นการใช้รหัสซ้ำคนละบริบท (ยังไม่ verify ว่าเครื่องยนต์รับ 20 เข้ามาจริงจากแหล่งใด) → **BC ห้ามคัดลอกการใช้รหัสซ้ำ** ให้ใช้ enum แยกชัดเจน.

### 5.3 วิธีคำนวณต้นทุน (`BCItem.CostType`)

ค่า `CostType`: **0 = เฉลี่ย (Average)**, **1 = FIFO/Lot** (ใช้ตาราง Lot), **2 = มาตรฐาน (Standard)** (`StockCenterProcess` ตรวจ `nCostType == 0/1/2`). หน้าจอสินค้าปัจจุบันเลือกได้ **เฉพาะ "ต้นทุนเฉลี่ย"** (หัวข้อ 2.1 ข้อ 5) แต่เครื่องยนต์รองรับครบสามแบบ.

**สูตรหลัก (ทุกแบบ) ต่อแถวรายการ** (`:2760-2763, :2820-2821`)

```
dTransRate / dItemRate = อัตราแปลงหน่วยของหน่วยในเอกสาร / หน่วยสต็อก   (ถ้า UnitType = 1 → GetStkPackingRate2 จาก BCStkPacking)
dSumOfCost = dQty × ((AverageCost ÷ dItemRate) × dTransRate)             ปัดด้วย CostPoint   ← ต้นทุนเฉลี่ยเป็นฐานของทุกกรณี
dProcQty   = dQty × dTransRate ÷ dItemRate                               ปัดด้วย QtyPoint     ← จำนวนในหน่วยสต็อก
```
(`AverageCost` เปลี่ยนเป็น `GetStandardCost(สินค้า, หน่วย, วันที่เอกสาร, 1)` เมื่อ `CostType = 2`.)

| กรณี | ต้นทุนที่ใช้ (`dSumOfCost`) / จำนวนเงินที่บวกเข้ายอดสินค้า | หลักฐาน |
|---|---|---|
| ซื้อ/รับเข้า (2, 3, 0, 1, 10, 12 ฯลฯ — รายการ "เข้า") | ใช้ **จำนวนเงินของเอกสาร** (`dTransNetAmount`) เป็นมูลค่าที่เพิ่ม (รวมต้นทุนแฝงเมื่อ `UseExpCost=1`) | `:2729-2732, :2826` |
| ขาย (4) · เพิ่มหนี้ลูกหนี้ (6) | `dSumOfCost` = ต้นทุนเฉลี่ย × จำนวน; **ไม่ใช้ราคาขาย** | `:2842-2843, :3157` |
| เบิก (8) · ปรับขาด (11) · โอนเข้า/ออก (12/13) · รับคืนลูกหนี้ (5) · รับคืนเบิก (9) | มูลค่า = `dSumOfCost` (ต้นทุนเฉลี่ย × จำนวน) เมื่อ CostType = 0 | `:2832-2841` |
| **รับคืนลูกหนี้ (5)** เมื่อเฉลี่ย | ใช้ **ต้นทุนของบรรทัดใบขายที่อ้างอิง** (`GetReferCost(InvoiceNo, reflinenumber)` × จำนวน); ถ้าหาไม่เจอ ใช้จำนวนเงินของใบรับคืน | `:2770-2785` |
| **ส่งคืนเจ้าหนี้ (7)** เมื่อเฉลี่ย | ใช้ **จำนวนเงินของใบส่งคืน** | `:2787-2791` |
| **รับคืนจากเบิก (9)** เมื่อเฉลี่ย | ใช้ **ต้นทุนต่อหน่วยของบรรทัดใบเบิกที่อ้างอิง** (`BCStkIssueSub.SumOfCost ÷ Qty` ของ `IssueRefNo` + `reflinenumber`) × จำนวน | `:2793-2815` |
| มาตรฐาน (`CostType = 2`) | ทุกรายการที่ "ปรับ SumOfCost" ใช้ `dSumOfCost` จากต้นทุนมาตรฐาน ณ วันที่เอกสาร; ตาราง `BCItemStandardCost` (UnitCode, StartDate, StopDate, StandardCost, IsRevalue, RevalueDate, TransState) | `:2756-2760, :2828-2829` |
| FIFO/Lot (`CostType = 1`, สินค้าไม่ใช่บริการ) | ถ้า `nLotStatus ≠ 0` (Lot กำหนดเอง): ออก = อ่าน `BCLotOut.CostOfLot` ของเอกสารนั้น; เข้า = อ่าน `BCLotMaster.Cost × จำนวน`. ถ้า `nLotStatus = 0` (Lot อัตโนมัติ): รับคืน (5, 9) จัดสรรกลับเข้า Lot ด้วย `ProcessCreditNote2`/`ProcessIssueRet2`; ชนิดเข้า (`bLotMaster`) → `UpdateLotMaster`; ชนิดออก → `UpdateLotOut` (คืนต้นทุนที่ตัดจาก Lot); โอน (12/13) → `StkItemTransferProcess2` | `:2979-3099` |

**การปรับยอดสินค้าหลังคำนวณ** (`:3167-3199`)
```
BCItem.StockQty = StockQty + (dProcQty × nMark)
BCItem.Amount   = Amount   + (dItemAmountNew)          dItemAmountNew = (ชนิดออกที่ปรับ SumOfCost ยกเว้น ซื้อ(2) ? ต้นทุนที่ตัด : จำนวนเงินเอกสาร) × nMark
ถ้า MyType ∈ {0,1,2,3,5,7,9,10,13}:  AverageCost = Amount ÷ StockQty   (เฉพาะ StockQty ≠ 0; ปัดด้วย CostPoint)
ถ้า StockQty = 0:                      AverageCost = 0 และ Amount = 0   ← ล้างมูลค่าตกค้างเมื่อสต็อกเป็นศูนย์
```
(**ขายออก/เบิกออก ไม่คำนวณ AverageCost ใหม่** — ต้นทุนเฉลี่ยเปลี่ยนเฉพาะตอน "เข้า" และตอนรับคืน/ส่งคืนบางชนิดตามรายการด้านบน.)

**บันทึกต้นทุนกลับที่เอกสาร:** `UPDATE <ตารางรายการย่อย> SET SumOfCost = ... WHERE RowOrder = ...` ทุกชนิดที่ `bUpdateSumOfCost` (ยกเว้นโอน MyType 13); เฉพาะใบเบิก (8) ถ้า `Price` ว่าง ให้ตั้ง `Price = SumOfCost ÷ Qty` (`:3250-3263`).

**การปัดเศษ:** ใช้จุดทศนิยมของบริษัท — จำนวน `QtyPoint`, ต้นทุน `CostPoint`, ราคา `PricePoint`, จำนวนเงินลงบัญชี `AmountPoint` (`XRound`/`RoundNumber`); ไม่มีการเก็บ "ส่วนต่างปัดเศษ" แยกต่างหากในเครื่องยนต์สต็อก (ส่วนต่างปัดเศษอยู่ฝั่ง GL ผ่านเงื่อนไข 90/91 ในหัวข้อ 5.6).

### 5.4 ตารางยอดคงเหลือและยอดงวดที่เครื่องยนต์เขียน

| ตาราง | คีย์ | คอลัมน์ | เขียนอย่างไร | หลักฐาน |
|---|---|---|---|---|
| `BCItem` | Code | StockQty, Amount, AverageCost, RemainInQty, RemainOutQty, ReserveQty, LastBuyPrice/LastBuyPriceHome/LastBuyCurrencyCode | บวก/ลบตามหัวข้อ 5.3; ซื้อ (2) อัปเดตราคาซื้อล่าสุดด้วย (และ `BCItemSupplier.LastBuyPrice/LastBuyDate/LastBuyUnitCode` ตามเจ้าหนี้+หน่วย ถ้า `nIsSkipAddItemSupplier = 0`) | `:3168-3213, :2919-2966` |
| `BCStkWarehouse` | ItemCode + WHCode + ShelfCode + UnitCode (**หน่วยของเอกสาร ไม่ใช่หน่วยสต็อก**) | Qty, Amount, ReserveQty, RemainInQty, RemainOutQty, ReadyQtyAll, ReserveQtyAll, RemainQtyAll, RemainOutQtyAll, ItemPriority, averagecost, DocDate | `UPDATE ... SET Qty = Qty+(จำนวน×nMark), Amount = Amount+(มูลค่า×nMark)`; ไม่พบแถว → `INSERT` (ต้องมีรหัสสินค้า/คลัง/ที่เก็บ/หน่วยครบ); ถ้าจำนวน=0 และมูลค่า=0 → ตั้งศูนย์; ถ้าจำนวนเป็น 0 ให้ล้าง `Amount` | `:1950-2031` |
| `BCStkPacking` | ItemCode + UnitCode | Qty, BalQtyPeriod1…25, Rate, Rate1, Rate2, Wide/Long/High, … | สินค้า `UnitType=1` (หลายหน่วย) บวก/ลบยอดตามหน่วยของเอกสารและยอดงวด 1-25 (สูงสุด 25 งวด) | `:1182-1247, :2746-2752` |
| `BCStkPeriodBal` | ItemCode + PeriodNo | BeginQty/Amount, SaleQty/Amount, BuyQty/Amount, StkRetQty/Amount, CNQty/Amount, Db1Qty/Amount, Db2Qty/Amount, IssueQty/Amount, IssueRetQty/Amount, FGQty/Amount, InQty/Amount, OutQty/Amount, EndQty/Amount, IsReProcess, UnitCode (หน่วยสต็อก) | ดูหัวข้อ 5.4.1 | `:2076-2278` |
| `BCStkWHPeriodBal` | ItemCode + WHCode + ShelfCode + UnitCode + PeriodNo | BeginQty/BeginAmount, Qty, Amount | `UPDATE qty=qty+(จำนวน×nMark), amount=amount+(มูลค่า×nMark)`; ไม่พบ → INSERT พร้อมยอดต้นงวดจากงวดก่อนหน้า (งวดล่าสุดที่ `PeriodNo <` ปัจจุบัน) | `:2281-2405` |
| `BCPeriod` | PeriodNo | StartDate, StopDate, Status | หา `PeriodNo` จากวันที่เอกสาร (`StartDate ≤ วันที่ ≤ StopDate`); ถ้าไม่พบงวด = −1 → **ไม่ปรับยอดงวด** | `BC5PROCESS\bc5proctools.cpp:59-81` |

#### 5.4.1 `BCStkPeriodBal` — ใครบวกคอลัมน์ไหน (`UpdateStkPeriodBalance` `:2100-2183`)

| MyType | คอลัมน์ที่บวก (จำนวน / มูลค่า) | ทิศ (`nEndMark`) | หมายเหตุ |
|---|---|---|---|
| 0 ยกมา | InQty / InAmount | + | |
| 1 รับสำเร็จรูป | FGQty / FGAmount | + | |
| 2 ซื้อ | BuyQty / BuyAmount | + | ใช้จำนวนเงินเป็นต้นทุน |
| 3 เพิ่มหนี้เจ้าหนี้ | Db2Qty / Db2Amount | + | |
| 4 ขาย | SaleQty / SaleAmount | − | มูลค่า = ต้นทุน (ไม่ใช่ราคาขาย) |
| 5 รับคืนลูกหนี้ | CNQty / CNAmount | + | |
| 6 เพิ่มหนี้ลูกหนี้ | Db1Qty / Db1Amount | − | |
| 7 ส่งคืนเจ้าหนี้ | StkRetQty / StkRetAmount | − | |
| 8 เบิก | IssueQty / IssueAmount | − | |
| 9 รับคืนเบิก | IssueRetQty / IssueRetAmount | + | |
| 10 ปรับเกิน | InQty / InAmount | + | ใช้คอลัมน์เดียวกับยกมา |
| 11 ปรับขาด | OutQty / OutAmount | − | |
| 12, 13 โอน | OutQty / OutAmount บวก **0** (จำนวนและมูลค่าตั้งเป็นศูนย์) | + | โอนไม่เปลี่ยนยอดรวมของสินค้า — แค่ให้มีแถวงวดอยู่ |
| 20 | InQty / InAmount | **+** | ⚠ ดูหัวข้อ "ที่ยังไม่ verify" — สต็อกจริงถูกหัก (nMark=−1) แต่ยอดงวดถูกบวก |

สูตรยอดคงเหลือปลายงวดที่ใช้ซ้ำทุกครั้งหลังบวก (`:2217-2218, :2262-2263`):
```
EndQty    = BeginQty    − SaleQty + BuyQty − StkRetQty + CNQty − Db1Qty + Db2Qty − IssueQty + IssueRetQty + FGQty + InQty − OutQty
EndAmount = BeginAmount − SaleAmount + BuyAmount − StkRetAmount + CNAmount − Db1Amount + Db2Amount − IssueAmount + IssueRetAmount + FGAmount + InAmount − OutAmount
```
(สตริง SQL เขียนเป็นลูกโซ่ `begin − sale + buy − …` ซึ่งตามลำดับการคำนวณซ้าย→ขวาให้เครื่องหมายถูกต้องตามชนิด.)
หลังแถวของงวดปัจจุบันถูกต้อง เครื่องยนต์ **ไล่งวดถัด ๆ ไปทุกงวด**: `BeginQty/BeginAmount` = `EndQty/EndAmount` ของงวดก่อนหน้าที่ใกล้ที่สุด แล้วคำนวณ End ใหม่ (`:2241-2268`); สุดท้ายตั้ง `EndAmount = 0` เมื่อ `EndQty = 0` ทุกงวดตั้งแต่งวดปัจจุบันเป็นต้นไป (`:2274-2276`). `IsReProcess` = 0 เมื่อมาจากคำนวณใหม่ทั้งชุด (`m_nProcMode = −1`), 1 เมื่อประมวลผลรายเอกสาร.

> ข้อสังเกต: การบวกยอดงวดและการไล่งวดถัดไปเป็นการ "เขียนซ้ำทุกแถวเอกสาร" ต่อสินค้า (ประสิทธิภาพ O(จำนวนงวดที่เหลือ) ต่อแถว) — **BC ไม่ต้องลอกวิธีบวกทีละแถว** แต่ผลลัพธ์ที่ต้องได้ตรงกัน: ยอดต้นงวด/ปลายงวดต่อสินค้า (และต่อคลัง/ที่เก็บ/หน่วย) ต่องวดบัญชี, แยกตามชนิดรายการด้านบน (ตอบรายงานหัวข้อ 7).

### 5.5 "คำนวณยอดสินค้าใหม่" — วิธีที่ Champ คำนวณซ้ำทั้งชุด (`StockReprocess` `bcdumpprocjob.cpp:1249-1500`)

1. ตั้งค่าสินค้าบริการ (`StockType = 1`) เป็นศูนย์ และลบ `BCStkWHPeriodBal`/`BCStkPeriodBal`/`BCStkWarehouse` ของสินค้าบริการ (`:1268-1278`).
2. ต่อสินค้า: ตั้ง `BCItem.StockQty/Amount/AverageCost` ตั้งต้นจาก **`EndQty/EndAmount` ของงวดก่อน "งวดเริ่มคำนวณ"** (`GetStkPeriodBalance(งวดเริ่ม−1)` `:1360-1379`); ลบ `BCStkPeriodBal` ตั้งแต่งวดเริ่ม, สร้าง `BCStkWarehouse` ใหม่จาก `BCStkWHPeriodBal` งวดล่าสุดก่อนงวดเริ่ม แล้วลบ `BCStkWHPeriodBal` ตั้งแต่งวดเริ่ม (`:1385-1401`); ลบ `BCLotMaster`/`BCLotOut` ของงวดเริ่มเป็นต้นไป (`:1410-1417`).
3. ดึงรายการทุกตารางรายการย่อย (13 แหล่ง) แล้วเรียงตาม **`DocDate, BehindIndex, Atype, MyType, RowOrder`** แล้วเดินเข้า `StockCenterProcess` ทีละแถว (`:1315-1368`). แหล่งที่ดึง (`atype`): ยกมา(1) · รับสำเร็จรูป(2) · ซื้อ(3) · เพิ่มหนี้เจ้าหนี้(10, เฉพาะ `StockProcess < 2`) · ขาย(5) · ลดหนี้ลูกหนี้(6, `StockProcess = 0` และ `DiscQty > 0`) · เพิ่มหนี้ลูกหนี้(10) · ส่งคืนเจ้าหนี้(10, `StockProcess < 2`) · เบิก(10) · รับคืนเบิก(11) · ปรับ+ (10, MyType 10) · ปรับ− (10, MyType 11) · โอนออก(4, `BCStkTransfSub3` MyType 13).
4. ตัวกรอง: เฉพาะแถว `TransState = 0 AND IsCancel = 0`.

→ **ลำดับในวันเดียวกันถูกกำหนดด้วย `BehindIndex`** (หัวข้อ 6.6) และชนิดเอกสาร (`Atype`/`MyType`); ถ้า `BehindIndex` เท่ากัน ลำดับเอกสารที่ต่างชนิดเป็นตามตัวเลข `Atype` — ผลต้นทุนเฉลี่ยของเอกสารวันเดียวกันขึ้นกับลำดับนี้ **BC ต้องกำหนดกติกาลำดับให้ตรงกัน (วันที่ → ลำดับรายวัน → ชนิด → RowOrder/เลขแถว) และเก็บ `BehindIndex` เป็นฟิลด์ของแถวเอกสาร**.

### 5.6 ลงบัญชี GL ของเอกสารสต็อก

กลไกร่วม (หัวข้อ 0.1): เอกสารมี `GLFormat` (รหัสรูปแบบการลงบัญชี) และ `GLTransData` (0 = อัตโนมัติ / 1 = มือ); เมื่ออัตโนมัติ เครื่องลงบัญชี (`CGLTLSProcess`) อ่านรูปแบบแล้วสร้างรายการรายวันตาม **"เงื่อนไข" (Condition)** ของแต่ละบรรทัดในรูปแบบ — บรรทัดรูปแบบมี `nCondition` + `nIsDebit` (0 = เดบิต, 1 = เครดิต) + บัญชี. ผูกกับบัญชีของสินค้าผ่าน **กลุ่มบัญชีสินค้า `BCItemAccGroup`** (Code, Name, **ItemAccount** บัญชีสินค้าคงเหลือ, **CostAccount** บัญชีต้นทุน, **RevenueAccount** บัญชีรายได้ — `atlas.json` tables; ฟังก์ชัน `getAccGroup`); สินค้าแต่ละตัวชี้กลุ่มด้วย `BCItem.AccGroupCode`.

| เอกสาร | รูปแบบเริ่มต้น (`GetDefaultGLFormat`) | `m_nSource` | กลุ่มประมวลผล (`m_nGLGroup` / `ProcessConditionGroupNN`) | ตารางที่รวมยอด |
|---|---|---|---|---|
| เบิกใช้ 5625 | 24 | 21 | 13 | `BCStkIssue` / `BCStkIssueSub` |
| รับคืนจากเบิก 5626 | 25 | 22 | 26 | `BCStkIssRetSub` |
| รับสินค้าสำเร็จรูป 5627 | 26 | 23 | 27 | `BCFinishGoodsSub` |
| ปรับปรุงหลังตรวจนับ 5630 | 34 (เกิน) / 44 (ชำรุด) / 45 (สูญหาย) / 46 (ปรับปรุง) ตาม `AdjustType` | 34 | 34 (+ 44/45/46 มีฟังก์ชัน `ProcessConditionGroup44/45/46` ของตัวเอง) | `BCStkAdjust` / `BCStkAdjustSub` |
| ยกมา 5624 · โอน 5628 · ขอเบิก 5601 · ขอโอน 5631 · ตรวจนับ 5629 · ผลตรวจนับ 5599 | **ไม่ลง GL** (ฟอร์มไม่มีรูปแบบ/ไม่เรียก `m_GLTLSProcess`) | — | — | — |

(หลักฐาน: `icfrmstockissue.cpp:1098-1099, 1189`; `icfrmstockissueret.cpp:247, 1405-1406`; `icfrmfinishgood.cpp:251, 462, 1204-1205`; `ICFrmStockAdjust.cpp:1321-1322, 2437-2448`; `gltlsprocess.cpp:8390, 8962, 10350, 10786`.)

**เงื่อนไขที่ใช้ในรูปแบบ (จำนวนเงินที่ดึงมาลงแต่ละบรรทัด)**

| Condition | ความหมาย (ตามคอมเมนต์/โค้ด) | ใช้กับ | ที่มาของจำนวนเงิน |
|---|---|---|---|
| 1 | ยอดรวมต้นทุน (ทั้งใบ) | เบิก 13 · รับคืนเบิก 26 · สำเร็จรูป 27 | `GetSummeryCostStkIssueSub(docNo, ตาราง)` = ผลรวม `SumOfCost` ของใบนั้น (13); รับคืนเบิก/สำเร็จรูปคิดจากต้นทุนสินค้ารวมปรับส่วนต่าง (`dGoodsCost + dGoodsCostDiff_CR + dGoodsCostDiff_DR`) |
| 2 | **แยกตามกลุ่มบัญชีสินค้า — ด้านต้นทุน (CostAccount)** | 13 · 26 · 27 | `PostStkIssueSub(..., nAccMode=2)` : รวม `SumOfCost` ตาม (`AllocateCode`, `ProjectCode`, `DepartCode`, กลุ่มบัญชีสินค้า) แล้วลงบัญชี `CostAccount` ของกลุ่มนั้น |
| 3, 5 | **แยกตามกลุ่มบัญชีสินค้า — ด้านสินค้าคงเหลือ (ItemAccount)** | 13 (3/5) · 26 (3) · 27 (3/5) | `nAccMode=0` → `ItemAccount` (โหมด BMS ใช้บัญชี "IN") |
| 4, 6, 7, 8, 9 | บัญชีออก / ชำรุด / รางวัล / แจก-โปรโมชัน / โฆษณา (บัญชีเสริมของ **โหมด BMS** เท่านั้น — ใน Champ ปกติ `getAccGroup` คืนแค่ 3 บัญชีด้านบน จึงเป็นค่าว่าง) | 13 | `nAccMode = 8/6/5/3/4` |
| 10 | ภาษี (`taxamount`) | 13 | ฟิลด์ใบเบิก |
| 11 / 12 | เดบิต / เครดิต **จำนวนเงินการเบิกแยกรายตัว** (`BCStkIssueSub.Amount`, แยก `JobNo`) | 13 | `PostStkIssueSub(..., nIsDebit=11/12)` |
| 13 | ภาษีหัก ณ ที่จ่าย (`sumofwtax`) | 13 | ฟิลด์ใบเบิก |
| 80 / 81 | **ส่วนต่างต้นทุน ด้านเดบิต / เครดิต** (`dGoodsCostDiff_DR/CR`) | 26 · 27 · 34 | คำนวณในฟังก์ชันกลุ่ม |
| 82 | ต้นทุนสินค้า (`dGoodsCost` / `dSumOfCost`) | 26 · 27 · 34 | |
| 83 | **บัญชีเดินสะพัดระหว่างสาขา** — ใช้บัญชีจาก `BCSaleBranch.AccountCode` ของสาขา (`CatCode`) | 13 · 27 · 34 | ใช้จำนวนเงินเป็นเงินบาท (`bIsHomeAmount`) |
| 90 / 91 | **ส่วนต่างปัดเศษ** ด้านเดบิต / เครดิต (สร้างบรรทัดปัดเศษให้เดบิต = เครดิต) | 13 · 26 · 27 | `MakeRoundScriptDr/Cr` |

เงื่อนไขของ **ใบปรับปรุง (กลุ่ม 34)** (`gltlsprocess.cpp:9132-9192`): 1 = ยอดรวมต้นทุนปรับปรุงลบ (อ่าน `BCStkAdjust.SumAmount`) · 2 = ยอดรวมต้นทุนปรับปรุงบวก (อ่าน `SumAmount2`) · 3/5 = รายตัวด้านบัญชี **RevenueAccount** ของกลุ่มบัญชีสินค้า (`PostStkAdjustSub nAccMode=1`) · 4/6 = รายตัวด้าน **ItemAccount** (`nAccMode=0`) · 7 = ภาษี (`TaxAmount`) · 80/81/82/83 ตามตารางด้านบน. แถวรายตัวของใบปรับปรุงรวมตาม `AdjMark` ผ่านพารามิเตอร์ `nStatus` (ยังไม่ verify ความหมายค่า `nStatus` ว่า 0/1 = บวก/ลบ) โดยบรรทัด GL เก็บ `debit/credit` (บาท) และ `fdebit/fcredit` (สกุลเงินเอกสาร) (`:11981-12036`).

> ⚠ **ยังไม่ verify:** `SumAmount` (ช่อง `m_edPositiveTotal`) vs `SumAmount2` (ช่อง `m_edTotal`) ของ `BCStkAdjust` ตามชื่อคอมเมนต์ "ลบ/บวก" **ไม่ตรงกับชื่อช่องบนจอ** (หัวข้อ 4.10) — ต้องรันกับข้อมูลจริงหรืออ่าน `OnSaveData` (`ICFrmStockAdjust.cpp:425-470`) ซ้ำก่อนใช้สูตรนี้ใน BC; ถ้าไม่แน่ใจให้ BC คำนวณยอดบวก/ลบจากรายการย่อย (`AdjMark`) ตรง ๆ แล้วไม่ต้องสืบทอดสองคอลัมน์นี้.

### 5.7 ข้อสังเกตด้านความถูกต้อง (สิ่งที่ BC "ห้ามลอก" และต้องทำให้ดีกว่า)

1. **ทุกจำนวนเงินใน Champ เป็น `double` + `%lf` ใน SQL** (`bcdumpprocjob.cpp:2562, 3168-3169`) — ขัดกฎ Accounting Number ของเจ้าของ (Rule 20): BC ต้องใช้ `NUMERIC` / decimal และกำหนดนโยบายปัดเศษ (เช่น ต้นทุนต่อหน่วยปัดที่ `CostPoint`, มูลค่าบรรทัดปัดที่ `AmountPoint`) พร้อมบันทึกส่วนต่างปัดเศษ.
2. **SQL ต่อสตริงด้วย `Format()`** ทุกแห่ง — BC ใช้ parameterized query เท่านั้น.
3. **เขียนยอดทีละแถว ไม่ atomic** ระหว่างตาราง (BCItem / BCStkWarehouse / BCStkPeriodBal ถูก UPDATE แยกคำสั่ง) — BC ต้องทำใน transaction เดียว (`sql.Tx`) ต่อเอกสาร (ตามกฎ Pure PostgreSQL synchronous ACID).
4. **ยอดงวดของ MyType 20** (ดูตาราง 5.4.1) บวกแทนที่จะลบ — เก็บเป็นข้อสงสัย ยังไม่ verify ว่ามีแหล่งข้อมูลส่งแถว MyType 20 เข้าเครื่องยนต์จริงหรือไม่.
5. **ลำดับรายวันขึ้นกับ `BehindIndex` ที่ผู้ใช้แก้มือได้** แล้วต้องคำนวณใหม่ — BC ต้องออกแบบลำดับเป็นข้อมูลที่ตรวจสอบย้อนหลังได้ (เหตุผล + ผู้แก้ + เวลา).

---

## 6. สินค้าชุด · Serial Number · ใบตรวจรับสินค้า · เครื่องมือจัดลำดับ/คำนวณใหม่

> กลุ่มนี้คือเมนู IC ที่ไม่ใช่เอกสารสต็อกตรง ๆ — **สินค้าชุด** ใช้เอกสาร *เบิก* + *รับสินค้าสำเร็จรูป* ที่มีอยู่แล้วเป็นตัวตัด/เพิ่มสต็อก (ไม่มีเครื่องยนต์สต็อกแยก), **Serial** เป็นทะเบียนรายตัว, **ใบตรวจรับ** ไม่กระทบสต็อก, ส่วน 5623/5621 เป็นเครื่องมือดูแลข้อมูลสต็อก. กลไกฟอร์มร่วมดูหัวข้อ 0.1.

### 6.1 5613 "บันทึกสินค้าชุด" (`CICFrmItemSet : CMyEspeciallyFormView`, `IDD_IC_FRM_ITEMSET` 342×243; atlas-07:1355)

**ความหมาย:** กำหนด "สูตรส่วนประกอบ" (BOM ชั้นเดียว) ของสินค้าที่มี `StockType = 5` (สินค้าชุด) — ไม่ใช่เอกสารรายวัน ไม่มีเลขที่เอกสาร/วันที่/สถานะ. รหัสสินค้าชุดต้องสร้างในข้อมูลหลักสินค้า (5602) ก่อนโดยตั้ง `StockType = สินค้าชุด` แล้วค่อยมากำหนดส่วนประกอบที่นี่ (`icfrmitemset.cpp:675-676` ค้นหาด้วย `StockType = 5`).

**ส่วนหัว (ลำดับบนจอ — `BC5Account.rc:4641`)**

| # | ป้าย (Thai) | ชนิด | บังคับ | ที่มา/กติกา |
|---|---|---|---|---|
| 1 | รหัสสินค้าชุด | ช่องรหัส + ปุ่มค้นหา (`IDC_BTTN_SET_CODE`) | ใช่ (`m_bMustEntry`, `:133`) | ค้นหาเฉพาะ `StockType = 5` → โหลดสูตรเดิมมาแก้ (`OnFindItem :669-690`) |
| 2 | ชื่อสินค้า | อ่านอย่างเดียว | — | ชื่อตาม `DefaultItemName` ของบริษัท (Name1/Name2) |
| — | จำนวนเงิน (ท้ายจอ, `IDC_ED_AMOUNT`) | ตัวเลข อ่านอย่างเดียว (ทศนิยม `AmountPoint`) | — | = ผลรวมคอลัมน์ "ราคารวม" ของทุกส่วนประกอบ (`:412`) |

**ตารางส่วนประกอบ (คอลัมน์ตามลำดับ — `:113-123`)**

| # | หัวคอลัมน์ | ชนิด | ทศนิยม | หมายเหตุ |
|---|---|---|---|---|
| 0 | รหัสสินค้า | พิมพ์ + ค้นหา (`FCS_EDIT` + `FCS_LINK`) | — | ตัวพิมพ์ใหญ่อัตโนมัติ; รับ `จำนวน*รหัสบาร์โค้ด` หรือ `จำนวน_รหัสบาร์โค้ด` (`:1066-1090`); ค้นหาได้เฉพาะสินค้า `StockType <> 5 AND <> 7` (ห้ามเอาสินค้าชุดเป็นส่วนประกอบของสินค้าชุด และห้ามชนิด 7) (`:528`); รหัส `.` ใช้เป็นแถวคำอธิบาย |
| 1 | ชื่อสินค้า | อ่านอย่างเดียว | — | |
| 2 | หน่วยนับ | คอมโบ (`BCItemUnit`; ถ้าสินค้าหลายหน่วยจำกัดเฉพาะหน่วยใน `BCStkPacking`) | — | ค่าเริ่มต้น = `DefSaleUnitCode` (`:536`) |
| 3 | ราคาต่อหน่วย | ตัวเลขแก้ได้ | `PricePoint` | ค่าเริ่มต้น = ราคาขาย 1 ของหน่วยนั้นจาก `BCPriceList` (โครงการ CHM ใช้ `AverageCost` แทน — ธง `#ifdef CHM_PROJ` `:540-543`) |
| 4 | จำนวน | ตัวเลขแก้ได้ | `QtyPoint` | |
| 5 | ราคารวม | ตัวเลข (คำนวณ) | `AmountPoint` | = จำนวน × ราคาต่อหน่วย ปัดด้วย `AmountPoint` (`DetailCalc :610-627`) |
| ซ่อน | RowOrder | shadow col 0 | | |

**กติกาตอนบันทึก (`ConfirmSaveData :287-313`):** ต้องมีรหัสสินค้าชุด; **ทุกแถวที่มีรหัสสินค้า ต้องมีหน่วย, จำนวน > 0 และ ราคาต่อหน่วย > 0** (ราคา 0 ใช้ไม่ได้ — ส่วนประกอบแจกฟรีทำไม่ได้) ไม่งั้นขึ้นข้อความ «CANNOT_SAVE_NEEDDETAIL» แล้วไม่บันทึก.

**การบันทึก (`OnSaveData :417-457`):** `DELETE FROM BCITEMComponents WHERE ParentCode=<รหัสชุด>` แล้ว `INSERT` ทีละแถวใหม่ทั้งหมด → ตาราง **`BCITEMComponents`** (ParentCode, ItemCode, UnitCode, Price, Qty, Amount, RowOrder; + `SaleType`, `TaxType` เฉพาะแบบที่ 2) → บันทึก logbook «80002 บันทึกรายการย่อยสินค้าชุด» (ใหม่/แก้) → **`UpdatePriceList()`**: เขียน "ราคาขาย 1" ของสินค้าชุด (หน่วย = `DefSaleUnitCode` ของสินค้าชุด) = ผลรวมราคารวมส่วนประกอบ ลง `BCPriceList` (ถ้ายังไม่มี → INSERT `FromQty=1, ToQty=99, LineNumber=0, StartDate=วันนี้, StopDate=วันนี้+100 ปี`; ถ้ามี → UPDATE) (`:1000-1040`).

> ข้อสังเกต: การแก้ไขสูตรไม่กระทบเอกสารเก่า (ตาราง BOM ไม่มีเวอร์ชัน/วันที่มีผล) และไม่ต้องคำนวณสต็อกใหม่. **การขายสินค้าชุด** (ตัดสต็อกส่วนประกอบหรือตัดตัวชุด) ทำในเอกสารขายตาม `StockType = 5` (อยู่นอกขอบเขตหัวข้อนี้ — ดูเอกสารส่วนขาย; ยังไม่ verify ในส่วนของ IC).

### 6.2 5614 "บันทึกรายการย่อยสินค้าชุดแบบที่ 2" (`CICFrmItemSetType2 : CMyBaseDBFormView`, `IDD_IC_FRM_ITEMSET_2` 509×267; atlas-07:1408)

สูตรส่วนประกอบ **แยกตาม "ประเภทการขาย (SaleType)" × "ประเภทภาษี (TaxType)"** — สินค้าชุดตัวเดียวมีหลายสูตร (เช่น ขายสด/ขายเชื่อ × ภาษีแยก/รวม) เก็บใน `BCITEMComponents` เดียวกัน โดยมีคอลัมน์ `saletype`, `taxtype` เพิ่ม.

| ส่วน | รายละเอียด |
|---|---|
| หัว | รหัสสินค้าชุด + ปุ่มค้นหา · **ประเภทการขาย** (คอมโบ) · **ประเภทภาษี** (คอมโบ) · ชื่อสินค้า · **จำนวนเงิน** (ผลรวมราคารวม) · ปุ่มเปิดจอ "เลือกราคาขายสินค้า" (`ICFrmItemSetType2Expand`) |
| ตาราง (`:513-518`) | รหัสสินค้า (ค้นหา) · ชื่อสินค้า · หน่วยนับ (คอมโบ) · **จำนวน** · ราคาต่อหน่วย · ราคารวม — *สลับลำดับจำนวน/ราคาเทียบแบบที่ 1* (ที่นี่: จำนวนก่อนราคา; ราคาใช้ `AmountPoint`) |
| โหลด | เปลี่ยนประเภทการขาย/ภาษี → โหลดเฉพาะแถวที่ `ISNULL(TaxType,0)=… AND ISNULL(SaleType,0)=…` (`:835, :1027`) |
| กฎ | ขายต่ำกว่าทุน → เตือน «SALE_LOW_COST» (`OnSalePriceLowCost :746`); ราคา ≤ 0 → เตือน «PRICE_MORE_ZERE» (`:753`); หน่วยไม่มี/ไม่ตรงกับสินค้า → เตือน (`:673-678`) |
| บันทึก | `DELETE BCITEMComponents WHERE ParentCode=… AND isnull(TaxType,0)=… AND isnull(SaleType,0)=…` แล้ว INSERT ใหม่ (`:158-193, :340`); แล้ว UPDATE/INSERT `BCPriceList` ที่ **คอลัมน์ราคาขายที่เลือก (`SalePrice1..8`)** ของ (สินค้าชุด, หน่วย, TaxType, SaleType) ด้วยผลรวม (`:380-420`) |
| ลบ | `DELETE FROM BCITEMComponents WHERE ParentCode=…` ทั้งหมด (ทุกสูตร) + logbook ลบ (`:132-139`) |
| จอ "เลือกราคาขายสินค้า" (`IDD_IC_FRM_ITEMSET_2_EXPAND` 307×170) | รหัส/ชื่อชุด · หน่วยนับ · ประเภทภาษี · ประเภทการขาย · เงื่อนไขขนส่ง · **ต้นทุนสินค้า** (= `BCItem.AverageCost` ปัดด้วย `CostPoint`, `:100-101`) · ปุ่ม "แสดงราคาขาย" (แสดงช่วงราคา `FromQty/ToQty/StartDate/StopDate/SalePrice1..8` จาก `BCPriceList`) · วิทยุ "ราคาขาย 1…8" + ช่องแสดงราคา · ตกลง/ยกเลิก; ไม่เลือก → «กรุณาเลือกราคาขายสินค้า» |

### 6.3 5615 "ตรวจสอบสินค้าที่สามารถรวมเป็นสินค้าชุดได้" (ประกอบชุด; `CICFrmItemSetBuild`, `IDD_IC_FRM_ITEMSET_BUILD` 342×272; atlas-07:1388)

**หน้าที่:** คำนวณว่าจากสต็อกส่วนประกอบที่มี ประกอบสินค้าชุดได้กี่ชุด แล้ว "ประกอบจริง" ตามจำนวนที่ต้องการ → สร้างเอกสาร **ใบเบิก (เบิกส่วนประกอบ)** + **ใบรับสินค้าสำเร็จรูป (รับตัวชุด)** อัตโนมัติ 2 ใบ.

| # | ป้าย | ชนิด | กติกา |
|---|---|---|---|
| 1 | วันที่ | วันที่ (`BME_DATE_TH_AUTO`) | วันที่ของเอกสารเบิก+รับที่จะสร้าง |
| 2 | รหัสสินค้าชุด (+ชื่อ) | ช่องรหัส + ค้นหา | ค้นหาเฉพาะ `StockType = 5`; พอเลือกแล้วโหลดส่วนประกอบ (`OnLoadDataDetail :513`) |
| 3 | จำนวนสินค้าที่คำนวณได้ | ตัวเลข อ่านอย่างเดียว | = **ค่าต่ำสุด** ของ (สต็อกส่วนประกอบ ÷ จำนวนต่อชุด) ของทุกส่วนประกอบ (`:576-591`) |
| 4 | จำนวนสินค้าที่ต้องการ | ตัวเลข (`QtyPoint`) | ต้อง > 0 และ ≤ จำนวนที่คำนวณได้ |
| 5 | เลขที่ใบเบิก / เลขที่ใบรับสำเร็จรูป | ช่องเลขที่เอกสาร (running ตามรูปแบบ `ICStkIssueConfig` / `ICFinishGoodConfig` ของ `BCSysTRFormat`/`BCSysTRDocG`) | ผู้ใช้แก้ได้ |
| ปุ่ม | ประมวลผล (`!PROCESS`) | | |

ตาราง (อ่านอย่างเดียว ทั้งหมด `FCS_NULL`, `:142-149`): รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ราคาต่อหน่วย · จำนวน(ต่อชุด) · ราคารวม · **จำนวน (คงเหลือในสต็อก)** · **จำนวนที่ประกอบได้** (= สต็อก ÷ จำนวนต่อชุด). คอลัมน์ซ่อน: RowOrder, WHCode, ShelfCode, OppositeUnit, CategoryCode, GroupCode, BrandCode, TypeCode, FormatCode.

**ขั้นตอน `OnProcess :595-639`:**
1. ตรวจ: มีรหัสชุด, `จำนวนที่คำนวณได้ > 0`, `จำนวนที่ต้องการ > 0` และ ≤ ที่คำนวณได้ — ไม่ผ่าน → «QTY_OVER_FOR_BUILD».
2. `m_dAmount = จำนวนที่ต้องการ × ต้นทุนชุด` .
3. **`ProcessIssue()`** — INSERT `BCStkIssue` (DocNo, GLTransData=0, `GLFormat = GetDefaultGLFormat(24)`, `IsCompleteSave=0`, DocDate, CreatorCode, CreateDateTime) + `BCStkIssueSub` ต่อส่วนประกอบ: `Qty = จำนวนต่อชุด × จำนวนที่ต้องการ`, WHCode/ShelfCode/OppositeUnit/กลุ่มต่าง ๆ จากสินค้า, `behindindex = GetBehindIndex()`, `PackingRate1/2`, `TaxRate=0`; หลังบันทึกทุกแถวเรียก `OnStockReprocess` ต่อสินค้า, ตั้ง `IsCompleteSave=1`, แล้ว `StkItemIssueProcess` + ลงบัญชี `m_nSource=21, m_nGLGroup=13` (เหมือนใบเบิกปกติ หัวข้อ 4.3/5.6), บันทึก Serial/Lot ของใบเบิก (`OnSaveSerial(8,…)`, `OnSaveLotItem(8,…)`).
4. **`ProcessFinishGoods()`** (`:866-966`) — INSERT `BCFinishGoods` (`SumOfAmount`, `GLFormat = GetDefaultGLFormat(26)`) + 1 แถว `BCFinishGoodsSub`: ItemCode = รหัสชุด, `Qty = จำนวนที่ต้องการ`, **`Cost = OnCalcCost(ใบเบิกที่เพิ่งสร้าง) ÷ จำนวน`** (= ต้นทุนรวมส่วนประกอบที่เบิก ÷ จำนวนชุด, `:876-878`), `Amount = Qty × Cost`, หน่วย/คลัง/ที่เก็บ = ค่าเริ่มต้นของสินค้าชุด (`DefSaleUnitCode/DefSaleWHCode/DefSaleShelf`), `behindindex`, `IsProcess = (IsBehindIndex=0 || ไม่ใช่โหมดใหม่) ? 1 : 0`; แล้ว `StkItemFinishGoodProcess` + ลงบัญชี `m_nSource=23, m_nGLGroup=27`.
5. สำเร็จ → `OnStockReprocess(รหัสชุด, วันที่)` + logbook «80015 ตรวจสอบสินค้าที่ประกอบเป็นสินค้าชุดได้» + ข้อความ «PROCESS_COMPLETE»; ล้มเหลว → «MSG_PROCESSITEMSET_FAIL». แล้วล้างจอเริ่มใหม่.

→ **ต้นทุนของสินค้าชุดที่ประกอบ = ต้นทุนเฉลี่ยของส่วนประกอบที่เบิก ณ ขณะนั้น** (ไม่ใช่ราคาในสูตร BOM). ปุ่มบันทึก/ค้นหา/ลบ/ก่อนหน้า/ถัดไปของแถบเครื่องมือถูกปิด (`OnUpdateRecSave/Find/Delete/Next/Prev = FALSE`, `ICFrmItemSetBuild.cpp:785-808`); ใช้ปุ่ม "ประมวลผล" อย่างเดียว.

### 6.4 5616 "ตรวจสอบชุดสินค้าที่สามารถแยกเป็นสินค้าได้" (แยกชุด; `CICFrmItemSetPart`, `IDD_IC_FRM_ITEMSET_PART` 342×272; atlas-07:1398)

ตรงข้ามกับ 6.3: **เบิกตัวสินค้าชุดออก** และ **รับส่วนประกอบเข้า** เป็นสินค้าสำเร็จรูป.

- ส่วนหัวเหมือน 6.3 แต่มีปุ่ม **"คำนวณ"** (`IDC_BTTN_CALCULATE`, `OnCalculate :876-899`): ต้องไม่ให้ "จำนวนสินค้าที่ต้องการแยก" > "จำนวนสินค้าที่คำนวณได้" (ไม่งั้นเตือน «PLEASE_CHECKDATA» และล้างช่อง) แล้วเติมคอลัมน์ "จำนวนที่แยกได้" ของทุกส่วนประกอบ = จำนวนต่อชุด × จำนวนที่ต้องการแยก.
- `OnProcess :372-430`: ตรวจเหมือน 6.3 + **ทุกส่วนประกอบต้องมีจำนวนที่แยกได้ (กดคำนวณก่อน)**; `ProcessIssue()` → INSERT `BCStkIssue` (`GLFormat = GetDefaultGLFormat(24)`) + `BCStkIssueSub` **1 แถว = ตัวสินค้าชุด** (`Qty = จำนวนที่ต้องการแยก`, หน่วย/คลัง/ที่เก็บ = ค่าเริ่มต้นการขายของชุด) แล้วลงบัญชีกลุ่ม 13; `ProcessFinishGoods()` → INSERT `BCFinishGoods` + `BCFinishGoodsSub` **ต่อส่วนประกอบ** (`Qty = จำนวนที่แยกได้`, `Cost` = shadow col 9 = ต้นทุนปัจจุบันของส่วนประกอบจาก `ItemDetail.duCost` (`:695`, ยังไม่ verify ว่ามาจาก `AverageCost`), `Amount = Qty × Cost`; `BCFinishGoods.SumOfAmount` = ผลรวม) แล้วลงบัญชีกลุ่ม 27 และ `OnStockReprocess` ทุกส่วนประกอบ (`:546-625`).
- ⚠ มูลค่าที่ตัดออกของตัวชุด (เฉลี่ยของชุด) **ไม่จำเป็นต้องเท่า** มูลค่าส่วนประกอบที่รับเข้า (คิดจากต้นทุนส่วนประกอบแยกกัน) — ผลต่างตกเป็นเงื่อนไขส่วนต่างต้นทุน 80/81 ในการลงบัญชี FG (หัวข้อ 5.6). ยังไม่ verify ผลบัญชีจริงของกรณีนี้.

### 6.5 5618 "บันทึก Serial Number" (`CICFrmItemSerialNumber : CMyBaseDBFormView`, `IDD_IC_FRM_ITEM_SERIALNUMBER` 344×239; atlas-07:1433)

**ทะเบียนคุมสินค้ารายตัว** (รถ/เครื่องจักร/อุปกรณ์ที่ต้องติดตามเลขเครื่อง–เลขถัง–ประกัน) — ตารางหลัก **`BCSerialMaster`** (คีย์ `CtrlNo`), ประวัติเคลื่อนไหวอยู่ **`BCSerialTrans`** (CtrlNo, ItemCode, DocNo, DocDate, `SaveFrom`) ที่เอกสารอื่นเขียนให้อัตโนมัติ (`OnSaveSerial(ชนิดเอกสาร, …)` ในเอกสาร; เช่น ใบเบิกใช้ค่า 8).

**ฟิลด์บนจอ (ลำดับ/ตำแหน่งตาม `BC5Account.rc:4796`)**

| กลุ่ม | ป้าย (Thai) | ชนิด | คอลัมน์ `BCSerialMaster` | หมายเหตุ |
|---|---|---|---|---|
| บน | รหัสควบคุมทะเบียนสินค้า | ช่องรหัส+ค้นหา | `CtrlNo` | **บังคับ** («CANNOT_SAVE_NOCONTROLCODE»); คีย์หลัก |
| | รหัสสินค้า (+ชื่อ) | ช่องรหัส+ค้นหา | `ItemCode` | |
| | MODEL | ข้อความ | `Model` | |
| | สี | ข้อความ | `ItemColor` | |
| | วันที่หมดประกัน | วันที่ | `StopInsureDate` | |
| | (สถานะใช้งาน) | คอมโบ: 0 ไม่ใช้งาน(InActive) / 1 ใช้งาน(Active), ค่าเริ่มต้น 1 | `ActiveStatus` | (`:151-155`) |
| | หมายเลขทะเบียน | ข้อความ | `RegisterNo` | |
| | วันที่ทะเบียน | วันที่ | `RegisDate` | |
| | สถานะสินค้า | คอมโบ: ใหม่ / เก่า / ไฟแนนท์ (ค่า 0/1/2, เริ่มต้น 0) | `ItemStatus` | (`:166-172`) |
| | สถานะปัจจุบัน | คอมโบ: **0 ในสต็อก / 1 เบิกใช้งาน / 2 ขายแล้ว / 3 ยกเลิก** (เริ่มต้น 0) | `StockStatus` | อัปเดตโดยเอกสารอัตโนมัติ (`m_nStockStatus`) |
| | รายละเอียดเฉพาะ | ข้อความ | `ItemRemark` | |
| ล่าง | คลัง · ที่เก็บ | คอมโบ (`BCWarehouse`, `BCShelf`) | `WHCode`, `ShelfCode` | |
| | หมายเลขตัวถัง (+ "พิมพ์อยู่ที่") | ข้อความ ×2 | `CaseNo`, `CasePosition` | |
| | หมายเลขเครื่อง (+ "พิมพ์อยู่ที่") | ข้อความ ×2 | `SerialNo`, `SerialPosition` | |
| | วันที่รับเข้า · เลขที่เอกสารรับเข้า | วันที่ · ข้อความ | `ReceiveDate`, `ReceiveDocNo` | |
| | ซื้อจาก (เจ้าหนี้) | ช่องรหัส+ค้นหา | `ApCode` | |
| | พนักงานรับสินค้า | คอมโบ `BCSale` | `ReceiveMan` | |
| | วันที่ขาย · เลขที่เอกสารขาย | วันที่ · ข้อความ | `SaleDate`, `SaleDocNo` | |
| | ขายให้ (ลูกหนี้) | ช่องรหัส+ค้นหา | `ArCode` | |
| | พนักงานขายสินค้า | คอมโบ `BCSale` | `SaleMan` | |
| | ราคาขาย | ตัวเลข | `Price` | |
| (ซ่อน) | รูปภาพ | | `PictureName` | |
| (ระบบ) | ผู้สร้าง/ผู้แก้ + เวลา | | `CreatorCode`, `CreateDateTime`, `LastEditorCode`, `LastEditDateT` | |

**กติกา**
- **แก้ไขไม่ได้เมื่อ `StockStatus = 1` (เบิกใช้ไปแล้ว → «CANNOT_EDIT_SERIAL1») หรือ `2` (ขายไปแล้ว → «CANNOT_EDIT_SERIAL2»)** (`ConfirmSaveData :279-288`).
- **ลบไม่ได้** ถ้ามีรายการใน `BCSerialTrans` ที่ `SaveFrom IN (1,2,3)` ของ CtrlNo+ItemCode นั้น («NOT_SAVE_OTHER_DOC») หรือสถานะ ≠ 0 («CANNOT_DELETE_SERIALSTATUS»); ลบได้เฉพาะ "ในสต็อก" และไม่เคยถูกรับด้วยเอกสาร; ยืนยัน «WANT_DELETE_ITEM» ก่อน (`:230-255`) → `DELETE FROM BCSerialMaster WHERE CtrlNo=…`.
- **ปุ่มเสริม:** "ดูการเคลื่อนไหวเอกสาร" (`OnDocumentFlow` → `CMyDocuemtFlow`): รวมประวัติ `BCSerialTrans` ตาม `SaveFrom` → 0 ยกมา · 1 รับสำเร็จรูป · 2 ซื้อ · 3 เพิ่มหนี้เจ้าหนี้ · 4 ขาย · 5 รับคืน/ลดหนี้ · 6 เพิ่มหนี้ลูกหนี้ · 7 ส่งคืน/ลดหนี้ · 8 เบิก · 9 รับคืนเบิก · 10 โอน (`:842-870`); ปุ่มประมวลผลยอด Serial เป็นช่วงรหัส (`CICDlgSerialProcess`: จากรหัส…ถึง… + แถบความคืบหน้า); พิมพ์ป้าย/เอกสาร (`UPDATE … SET PrintCount = PrintCount+1`, `:805-807`).
- สินค้าที่ต้องใช้ Serial ถูกกำหนดที่หน้าสินค้า (หัวข้อ 2 — "ควบคุม Serial"); **การเลือก Serial ตอนทำเอกสาร** (เบิก/ขาย/ซื้อ/ปรับปรุง) ทำในจอย่อยของเอกสารนั้น (`OnSaveSerial` → เขียน `BCSerialTrans` + ปรับ `StockStatus`) — ยังไม่ verify รายละเอียดจอเลือก Serial ในแต่ละเอกสาร.

### 6.6 5600 "ใบตรวจรับสินค้า" (`CICFrmGoodsRec : CMyEspeciallyFormView`; atlas-07:1479)

**เอกสารตรวจรับของจากเจ้าหนี้ก่อนซื้อจริง — ไม่กระทบสต็อก/บัญชี** (ไม่มี `MyType`, ไม่มี `GLFormat`; ตาราง `BCGoodsRecCheck` / `BCGoodsRecChkSub`).

| ส่วน | รายละเอียด |
|---|---|
| หัว (`:195-201`) | **เลขที่เอกสาร** (running `DocNo`; ซ้ำ → ถามลองใหม่) · **วันที่** (`BME_DATE_TH_AUTO`) · **เจ้าหนี้** (`ApCode`, **บังคับ** «PLEASE_INSERT_AP») · คำอธิบาย (`MyDescription`) · หมายเหตุบริการ (`ServiceRemark`) · กลุ่มเอกสาร (`BillGroup`) |
| ตาราง (`:344-352`) | รหัสสินค้า (ค้นหา) · ชื่อสินค้า · **คลัง** · **ที่เก็บ** · **หน่วยนับ** · **ยอดที่รับ** (`Qty`) · **สมบูรณ์** (`PerfectQty`) · **ไม่ดี** (`BadQty`) · **แตก** (`BrokenQty`); ซ่อน: RowOrder, OppositeQty, OppositeUnit, **PORefNo** (อ้างใบสั่งซื้อ), MyDescription |
| ที่มา | ดึงรายการจาก **ใบสั่งซื้อ** ที่ยังไม่ยกเลิก (`BCPurchaseOrderSub WHERE DocNo=… AND TransState=0 AND IsCancel=0`, `:991`) |
| ตรวจตอนบันทึก (`:114-137`) | ต้องมีเจ้าหนี้, เลขที่เอกสาร, อย่างน้อย 1 รายการ; **ผลรวม `สมบูรณ์+ไม่ดี+แตก` ต้องเท่ายอดที่รับ** («CHECK_AMOUNT_QTY») |
| ลบ | ลบ `BCGoodsRecChkSub` แล้ว `BCGoodsRecCheck` |
| รายงาน | 5656 "รายงานประเมินการรับสินค้า" (หัวข้อ 7.5) |

### 6.7 5623 "กำหนดลำดับรายวัน" (`CICDlgUpdateBehindIndex`, `IDD_IC_DLG_UPDATE_BEHIND_INDEX` 470×230; atlas-07:1502)

**ปัญหาที่แก้:** เอกสารสต็อกหลายใบของสินค้าเดียวกันวันเดียวกัน ต้นทุนเฉลี่ยจะแตกต่างตามลำดับการประมวลผล — ผู้ใช้กำหนดลำดับเองได้ผ่าน `BehindIndex` (เปิดใช้เมื่อ `Option.IsBehindIndex = 1`).

| ส่วน | รายละเอียด |
|---|---|
| ตัวเลือก | วิทยุ **"ตามสินค้า"** / **"ตามวันที่"** (`IDC_RDO_ITEM/DATE`) · รหัสสินค้า (+ชื่อ) · วันที่ — ต้องกรอกรหัสสินค้าและวันที่ทั้งคู่ («กรุณาบันทึกรหัสสินค้า», «กรุณาบันทึกวันที่») |
| ปุ่ม | **แสดงรายการ** (F9) · **Default** (F3) · **ประมวลผล** (F5) · **ปิด** (Esc) |
| ตาราง (`:81-88`) | **ลำดับ** (แก้ได้) · รหัสสินค้า · **ประเภทรายวัน** · เลขที่เอกสาร · คลัง · ที่เก็บ · หน่วยนับ · จำนวน; ซ่อน: RowOrder จริง + ชื่อตารางต้นทาง |
| แหล่งรายการ (UNION 19 ตาราง, เรียง `ItemCode, DocDate, BehindIndex, rowindex, DocNo, RowOrder`) | `rowindex` ค่าเริ่มต้นตามชนิด: 100 ยกมา · 101 รับสำเร็จรูป · 102 ใบสั่งซื้อ(30) · 103 ยกเลิกใบสั่งซื้อ(31) · 104 ซื้อ · 105 โอนออก(13) · 106 โอนเข้า(12) · 107 เพิ่มหนี้เจ้าหนี้ · 108 ใบสั่งขายค้าง(20) · 109 ใบสั่งขาย(21) · 110 ยกเลิกใบสั่งขาย(22) · 111 ขาย · 112 ลดหนี้ลูกหนี้ · 113 เพิ่มหนี้ลูกหนี้ · 114 ส่งคืนเจ้าหนี้ · 115 เบิก · 116 รับคืนเบิก · 117 ปรับเกิน · 118 ปรับขาด (`:192-211`) |
| ประมวลผล (`:287-310`) | ต่อแถว: `UPDATE <ตารางต้นทาง> SET behindindex=<เลขลำดับที่แก้> WHERE docno AND itemcode AND docdate AND roworder` → **`INSERT INTO ProcessStock(ItemCode, DocDate, ProcessFlag=0)`** (คิวให้เซิร์ฟเวอร์คำนวณสินค้านั้นใหม่ตั้งแต่วันที่นั้น) |
| Default | เตือน: "การบันทึกลำดับโดยใช้ค่า Default ... **ไม่สามารถกลับไปใช้ค่าลำดับเดิมได้อีก** ต้องการบันทึกลำดับตามค่า Default?" ใช่/ไม่ใช่ (`:292`) → รีเซ็ตลำดับเป็นตามชนิดข้างบน |

> ข้อสังเกต: ชุดชนิดบนจอนี้รวมใบสั่งซื้อ/ใบสั่งขายที่ไม่กระทบสต็อกจริง (ใช้แสดงลำดับสำหรับยอดค้างรับ/ค้างส่ง); ใน `StockReprocess` ลำดับเริ่มต้นใช้ `Atype`/`MyType` ต่างจากตารางนี้ (หัวข้อ 5.5) — ลำดับ default สองที่ไม่ตรงกันเป๊ะ (ยังไม่ verify ผลต่างจริงในวันเดียวกัน).

### 6.8 5621 "คำนวณยอดสินค้าใหม่" (`CICFrmStockReProcess : CMyBaseFormView`, `IDD_IC_FRM_STOCKREPROC` 335×204; atlas-07:1512)

| ฟิลด์/ปุ่ม | รายละเอียด |
|---|---|
| จากสินค้า / ถึงสินค้า | ช่วงรหัสสินค้า (ว่าง = ทั้งหมด; มีแค่ "จาก" = สินค้าเดียว) |
| จากงวดที่ / ถึงงวด | เลขงวดบัญชี + แถบเลื่อนปรับ (`m_nStartPeriod`) → **ช่วงวันที่** แสดงอ่านอย่างเดียว (`IDC_ED_SHOWDATE`; ได้จาก `BCPeriod`) |
| ☐ Display Process Step | แสดงขั้นตอนใน list |
| ☐ **คำนวณยอดยกมาของ Period** (`IDC_CHK_PERIODBAL_ONLY`) | ทำเฉพาะยอดงวด (`StockReUpdatePeriodBal(สินค้า, DefStkUnitCode, 0)` ต่อสินค้า) ไม่คำนวณต้นทุนเอกสารใหม่ (`:113-153`) |
| ปุ่ม | Start · Stop · Export.. · Stock Period (`ID_CHECK_STOCK`) · **Import Queue** (นำเข้ารายการสินค้าจากไฟล์ .txt → ใส่ช่อง "จากสินค้า") · Close |
| ความคืบหน้า | แถบ "Trans Progress" (ต่อสินค้า) + "Total Progress" + รายการผล (`IDC_REP_LIST`) |
| ก่อนเริ่ม | ถ้าไม่มีช่วงวันที่ → «Can not recalc stock» และไม่เริ่ม |
| การทำงาน (`:166-193`) | ตั้ง `m_nProcMode = -1` (คำนวณทั้งชุด) + ช่วงสินค้า + `m_nStartPeriod` + ช่วงวันที่ → รันงานเบื้องหลัง `StockReprocess` ตามหัวข้อ 5.5 |
| จอย่อย `CICDlgStockReprocess` ("Recalculate Item") | เงื่อนไขสินค้า (ตัวกรอง) · ☐ วันที่คำนวณ · ☐ **Update Packing Rate** · ☐ **คำนวณ Flow เอกสารใหม่** · ☐ **คำนวณ GL ใหม่** + "เริ่มจากวันที่/ถึงวันที่" · F9 Import Item · F5 ตกลง · Esc ปิด · **ยกเลิกสินค้ารอคำนวณ** (ล้างคิว `ProcessStock`) |

> ข้อสังเกต: การคำนวณใหม่ทำที่ฝั่งเซิร์ฟเวอร์/งานเบื้องหลัง (ไม่ใช่ใน transaction เดียว) และเปิดให้ผู้ใช้สั่งเองได้ทุกเมื่อ → BC ควรเป็นงานประมวลผลที่ backend เท่านั้น (ตามกฎ Backend-First) มีสถานะงาน/ความคืบหน้า/ยกเลิกได้ และบันทึกผู้สั่ง (ดูหัวข้อ 8).

---

## 7. รายงานของ IC (46 รายการ: 5603–5657 และกลุ่มวิเคราะห์สินค้า)

> **ขอบเขต/หลักฐาน.** เมนูรายงานของ IC เป็นรายการ popup ที่ซ่อนจากเมนูหลัก (`disable="yes"`; เปิดผ่านศูนย์รายงาน Report Center — ดู `00-common-ui.md` §12.2 เรื่องหน้าต่างรายงาน; การจัดกลุ่มดูหัวข้อ 1). ตัวจับคู่ **เมนู → handler (`MainFrm.cpp`) → คลาสรายงาน (`champ-report\BC5REP_IC\`, `BC5REP_IC2\`)** ได้จากสคริปต์ที่ตรวจ `RUNTIME_CLASS`/`REPComm_InsertCommand` ของแต่ละ handler (เช่น `MainFrm.cpp:8686-8696, 9378-9385`). ป้าย «KEY» ที่ยังไม่ resolve (ไฟล์ทรัพยากรภายนอก) เขียนเป็นคีย์/ชื่อที่อ่านได้จากโค้ด. **ตัวแปลงรหัสคอลัมน์ → ภาษาไทย** ที่เห็นในสคริปต์เป็นคำอธิบาย (gloss) ของ `R("KEY")` ซึ่งไม่ใช่ข้อความที่พิมพ์จริงทุกตัว (ยังไม่ verify).

### 7.1 กรอบร่วมของรายงาน IC

| หัวข้อ | รูปแบบ |
|---|---|
| **รุ่นเก่า (`CICRep*`)** | หน้าต่างเงื่อนไขธรรมดา: ช่อง "พิมพ์จาก … / พิมพ์ถึง …" ต่อฟิลด์ (วันที่, เลขที่เอกสาร, สินค้า, คลัง, ที่เก็บ, แผนก, ผู้เบิก, โครงการ ฯลฯ), ตัวเลือก "เรียงตาม", เช็ก "แสดงต้นทุน/หน่วย", "แสดง Serial" แล้วกด "พิมพ์/ตัวอย่างก่อนพิมพ์" — ข้อมูลอ่านตรงจากตารางเอกสาร |
| **รุ่น IC2 (`CIC2Rep*`)** | กรอบเงื่อนไขมาตรฐานเดียวกันทุกรายงาน: **เรียงตาม/รายงานแบบ (group-by)** + **พิมพ์จากรายการทั้งหมด** + **Select Config** (เลือก/สร้าง/ลบ "ชุดเงื่อนไขที่บันทึกไว้" ตาราง `FieldConfig`(IdCode, FieldData, MyDescription, LinkTable, FieldType), รายการฟิลด์ให้เลือกจาก `FieldDetail`(TableName) ผ่านจอ `CMyDlgFieldDetail`; ไม่ระบุเงื่อนไขเลย → «PLEASE_CONFIG_PRINT_CONDITION») + **โหมดคลัง 3 แบบ (วิทยุ)**: *พิมพ์คลังรวม / พิมพ์แยกคลัง / พิมพ์คลังรวมและคลังอื่น ๆ* (`m_nShowWH`) — ช่วงวันที่เริ่มต้นดึงจาก **งวดบัญชีของวันนี้** (`SELECT StartDate,StopDate FROM BCPeriod WHERE วันนี้ BETWEEN`, `IC2RepItemMovementDlg.cpp:781`). การสั่งพิมพ์ **เขียนคำขอลงตาราง `ReportStatus`** (SessionId, ReportId, Condition, GroupBy, StartDate, StopDate, ReqTime, Condition2-4) แล้วงานฝั่งเซิร์ฟเวอร์ประมวลผลใส่ตารางชั่วคราว `BCIC2Report`/`BCIC2Report2` ให้หน้าต่างรายงานอ่าน (`:287`; ใครเติมตารางชั่วคราว ยังไม่ verify) |
| ปุ่ม "รูปแบบกรมสรรพากร" | หลายรายงาน IC2 (5574, 5578, 5571, 5575…) มีเช็ก **"พิมพ์ตามรูปแบบข้อกำหนดของกรมสรรพากร"** ให้ผลลัพธ์เป็นรูปแบบบัญชีคุมสินค้า (วัน/เลขที่/ใบกำกับ/รายวัน/Lot/ปริมาณ-ต้นทุน-มูลค่า) — รายละเอียดแบบฟอร์มที่แท้จริงต้องตรวจจากเอกสารทางการ (ไม่ได้ยืนยันในที่นี้) |
| หน้าต่างรายงาน | ตามหัวข้อ 12.2 ของ `00-common-ui.md` (ตัวอย่างก่อนพิมพ์, ส่งออก, หัวกระดาษชื่อบริษัท/ช่วงวันที่, ท้ายรายงานรวมยอด) |

### 7.2 รายงานข้อมูลหลักสินค้า/ราคา/Serial

| เมนู | ชื่อ (Champ) | คลาส (key) | ทิศ | ตัวกรอง | คอลัมน์ (ตามลำดับ) | ตาราง |
|---|---|---|---|---|---|---|
| 5603 | รายงานรายละเอียดสินค้า | `CIC2RepItemMasterView` (84001) | ขวาง | วันที่(จาก-ถึง) *(ป้ายยังไม่ resolve)* · เรียงตาม (กลุ่มสินค้า / ยี่ห้อ / ประเภท / หมวด / รหัสสินค้า) · ☐ พิมพ์จากรายการทั้งหมด · Select Config | รหัสสินค้า · ชื่อสินค้า · หน่วยนับ · ประเภทสินค้า · วิธีต้นทุน · สถานะ · คิดภาษี · ชื่อ 2 · รายละเอียด · หน่วย (ชื่อ) · อัตราส่วน 1/2 · ใช้กับ(USEFLAG/USE_FORM) — หัวกลุ่มตามตัวเลือกเรียงตาม (รหัส+ชื่อกลุ่ม) | BCItem |
| 5608 | รายงานรายละเอียดสินค้า-ตาม Supplier | `CIC2RepItemMasterSuppView` (84005) | ขวาง | ☐ พิมพ์จากรายการทั้งหมด · Select Config | หัวกลุ่มเจ้าหนี้ (รหัส+ชื่อ) → รหัสสินค้า · ชื่อ · หน่วย · ประเภท · ต้นทุน · สถานะ · คิดภาษี · ชื่อ 2 · รายละเอียด · ชื่อหน่วย · อัตราส่วน | BCItem (+BCItemSupplier ยังไม่ verify) |
| 5609 | รายงานรายการสินค้าแบบมี Serial | `CICRepBCItemSerialView` (84006) | ขวาง | สินค้า(จาก-ถึง) · วันที่รับเข้า(จาก-ถึง) · วันที่หมดประกัน(จาก-ถึง) · สถานะ | รหัสสินค้า · ชื่อสินค้า · รหัสทะเบียนคุม · รายละเอียดเฉพาะ · หมายเลขตัวถัง · พิมพ์อยู่ที่ · หมายเลขเครื่อง · วันที่รับเข้า · วันหมดประกัน · สถานะปัจจุบัน · สถานะสินค้า | BCItem, BCSerialMaster |
| 5610 | รายงานเคลื่อนไหว Serial Number | `CICRepSerialMovementView` (84008) | แนวตั้ง | วันที่รับเข้า(จาก-ถึง) · สินค้า(จาก-ถึง) · Serial No.(จาก-ถึง) | Serial No. · หมายเลขตัวถัง · รหัสสินค้า · ชื่อสินค้า · หมายเลขทะเบียน · สถานะปัจจุบัน · วันที่รับเข้า · พนักงานรับสินค้า → (ประวัติ) วันที่เอกสาร · เลขที่เอกสาร · รายวัน/คำอธิบาย · คลัง · ที่เก็บ · ชื่อลูกค้า | BCSerialMaster, BCSerialTrans |
| 5612 | รายงานสินค้าใหม่ *(ปิดในเมนู)* | `CICRepItemNewView` («86002 รายงานสถานะสินค้า») | ขวาง | วันที่รับเข้า(จาก-ถึง) · กลุ่มสินค้า(จาก-ถึง) · สินค้า(จาก-ถึง) · สถานะสินค้า(จาก-ถึง) | กลุ่ม (รหัส+ชื่อ) · รหัสสินค้า · ชื่อ · วันที่เริ่มมีรายการ (FIRSTDATE) · วันที่รับล่าสุด (RECLASTDATE) · วันที่จ่ายล่าสุด (PAYLASTDATE) · จำนวนวันห่าง (DIFFDATE) | BCItem |
| 5657 | รายงานการกำหนดราคาขายสินค้า *(ปิดในเมนู)* | `CICRepPriceListView` (86003) | ขวาง | พนักงานที่แก้ไข(จาก-ถึง) · กลุ่มสินค้า(จาก-ถึง) · สินค้า(จาก-ถึง) · หน่วยนับ(จาก-ถึง) · ☐ แสดงตามกลุ่มสินค้า | รหัส/ชื่อกลุ่ม · รหัสสินค้า · ชื่อสินค้า · ชื่อที่ 2 · ยอดคงเหลือ · ราคาซื้อล่าสุด · หน่วยนับ · จากจำนวน · ถึงจำนวน · ราคาขาย 1-8 · ราคาคิดภาษี · ผู้แก้ไขล่าสุด · วันที่แก้ไข · โปรโมชั่นของสินค้า | BCItem, BCPriceList |

### 7.3 รายงานสินค้ายกมาและโอนสินค้า

| เมนู | ชื่อ (Champ) | คลาส (key) | ตัวกรอง | คอลัมน์ | ตาราง |
|---|---|---|---|---|---|
| 5632 | รายงานรายการสินค้ายกมาต้นปี | `CICRepBCStkBalanceView` (85001) | วันที่ · สินค้า · คลัง · ที่เก็บ (ทุกตัว จาก-ถึง) | รหัสสินค้า · ชื่อสินค้า → วันที่เอกสาร · เลขที่เอกสาร · คำอธิบายรายการ · แผนก · คลังสินค้า · ที่เก็บสินค้า · หน่วยนับ · จำนวนสินค้า · ต้นทุนที่เบิก · จำนวนเงิน | BCStkBalanceSub |
| 5633 | รายงานการบันทึกยอดสินค้ายกมาต้นปี | `CICRepBCStkBalanceStartYearView` (850011) | วันที่ · **เลขที่เอกสาร** · สินค้า · คลัง · ที่เก็บ (จาก-ถึง) · ☐ แสดงตามสินค้า | เลขที่เอกสาร · วันที่ · หมายเหตุ · รหัสสินค้า · ชื่อ · คลัง · ที่เก็บ · หน่วยนับ · จำนวน · ต้นทุน/ต่อหน่วย · จำนวนเงิน (+ คำอธิบาย/แผนก/ต้นทุนที่เบิกเมื่อแสดงตามสินค้า) | BCStkBalanceSub |
| 5634 | รายงานโอนสินค้าระหว่างคลังและรายการย่อย | `CICRepStkTransferView` (85100) | เรียงตาม · เลขที่เอกสาร · วันที่เอกสาร · แผนก · **คลังโอนออก** · **คลังโอนเข้า** · รหัสสินค้า (จาก-ถึง) · ☐ แสดง Serial · **การยืนยันการเบิก** (สถานะยืนยัน) | วันที่ · เลขที่ · แผนก · หมายเหตุเอกสาร → รหัสสินค้า · ชื่อ · จากคลัง · จากที่เก็บ · เข้าคลัง · เข้าที่เก็บ · หน่วย · จำนวน · ต้นทุนรวม · หมายเหตุ (+ Serial: รหัสควบคุมทะเบียน · หมายเลขทะเบียน · ตัวถัง · เลขเครื่อง) | BCStkTransfer, BCStkTransfSub ⚠ |
| 5635 | รายงานโอนสินค้าระหว่างคลัง-ตามคลังโอนออก | **ต่อผิดคลาส**: handler เปิด `CICRepBCIssue2View` (85600; หน้าต่าง/ตาราง = *รับคืนเบิก* `BCStkIssRetSub`) — คลาสที่ตรงชื่อ `CICRepBCMoveWareHouseView` (85102; "โอนตามคลัง") มีอยู่แต่ไม่ถูกเรียก (`MainFrm.cpp:9378-9385`) | (ของ 85600) เรียงตาม · วันที่ · เลขที่ · ผู้เบิก · แผนก · โครงการ · สินค้า (จาก-ถึง) · แสดงต้นทุน/หน่วย | (ของ 85600) วันที่ · เลขที่ · คำอธิบาย · รหัสสินค้า · ชื่อ · ผู้เบิก · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · … ; (ของ 85102) จากคลัง · จากที่เก็บ · เข้าคลัง · เข้าที่เก็บ · สินค้า · หน่วย · วันที่ · เลขที่ · จำนวน · ต้นทุน · Serial | BCStkIssRetSub / BCStkTransfSub |
| 5638 | รายงานโอนสินค้าระหว่างคลัง-ตามคลังโอนเข้า | `CICRepBCMoveWareHouseInView` (85108) | วันที่ · เลขที่เอกสาร · สินค้า (จาก-ถึง) · ☐ แสดง Serial | เข้าคลัง · เข้าที่เก็บ · จากคลัง · จากที่เก็บ · รหัสสินค้า · ชื่อ · หน่วย · วันที่ · เลขที่ · จำนวน · ต้นทุนรวม · Serial/หมายเหตุ | BCStkTransfSub |
| 5636 | รายงานโอนสินค้าระหว่างคลัง-ตามสินค้า | **ต่อผิดคลาส**: เปิด `CICRepStkReqTransferView` (85700 "รายงานขอโอนสินค้า…พร้อมรายการ" = *ใบขอโอน*) | เรียงตาม · เลขที่ · วันที่ · แผนก · คลังโอนออก · คลังโอนเข้า · สินค้า (จาก-ถึง) | วันที่ · เลขที่ · แผนก · หมายเหตุ · สินค้า · ชื่อ · จากคลัง/ที่เก็บ · เข้าคลัง/ที่เก็บ · หน่วย · จำนวน · หมายเหตุ | BCStkTransfer, BCStkTransfSub |
| 5637 | รายงานโอนสินค้าระหว่างคลัง-ตามวันที่ | `CICRepStkTransfSubDateView` (85104) | วันที่เอกสาร · เลขที่ · สินค้า (จาก-ถึง) · ☐ แสดง Serial | วันที่ · เลขที่ · สินค้า · ชื่อ · หมายเหตุ · จากคลัง/ที่เก็บ · เข้าคลัง/ที่เก็บ · หน่วย · จำนวน · ต้นทุนรวม · Serial | BCStkTransfSub |

> ⚠ รายงานโอนอ่านตาราง **`BCStkTransfer`/`BCStkTransfSub`** แต่เครื่องยนต์สต็อกอ่านโอนจาก **`BCStkTransfSub3`** (หัวข้อ 5.2) — ชื่อตารางโอนในรายงานกับในฟอร์มโอน (หัวข้อ 4.7) อาจไม่ตรงกัน (ยังไม่ verify ว่าเป็นตาราง/วิวเดียวกันหรือชื่อรุ่นเก่า).

### 7.4 รายงานเบิก · รับคืนเบิก · รับสินค้าสำเร็จรูป

**กลุ่มเบิก (5639–5643)** ตัวกรองพื้นฐานของทุกรายงานในกลุ่มคือ *วันที่ · เลขที่เอกสาร · สินค้า · ผู้เบิก · แผนก (จาก-ถึง)* (5639 เพิ่ม *เรียงตาม · โครงการ · ☐ แสดงต้นทุน/หน่วย*; 5643 เพิ่ม *โครงการ*).

| เมนู | ชื่อ | คลาส (key) | หัวกลุ่ม (ตัวแบ่งหน้า/รวมย่อย) | คอลัมน์ (ตามลำดับ) |
|---|---|---|---|---|
| 5639 | การเบิก-วันที่เบิก | `CICRepBCIssueDocDateView` (85201) | วันที่ → เลขที่ | วันที่ · เลขที่ · รหัสสินค้า · ชื่อ · คำอธิบายรายการ · ผู้เบิก · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · ต้นทุน/หน่วย · (รวมตามผู้เบิก/แผนก/โครงการ: จำนวนที่เบิก · ราคาขาย/หน่วย · รวมทั้งสิ้น) |
| 5640 | การเบิก-ผู้เบิก | `CICRepBCIssuePersonView` (85202) | ผู้เบิก (รหัส+ชื่อ) | วันที่ · เลขที่ · รหัสสินค้า · คำอธิบาย · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · ต้นทุน/หน่วย · Serial (รหัสควบคุม/ทะเบียน/ตัวถัง/เลขเครื่อง) · หมายเหตุ |
| 5641 | การเบิก-แผนก | `CICRepBCIssueDepartView` (85203) | แผนก (รหัส+ชื่อ) | วันที่ · เลขที่ · รหัสสินค้า · คำอธิบาย · ผู้เบิก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · Serial · หมายเหตุ |
| 5642 | การเบิก-สินค้า | `CICRepBCIssueItemView` (85204) | สินค้า (รหัส+ชื่อ) | วันที่ · เลขที่ · คำอธิบาย · ผู้เบิก · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · Serial · หมายเหตุ |
| 5643 | การเบิก-โครงการ | `CICRepBCIssueProjectView` (150011) | โครงการ (รหัส+ชื่อ) | สินค้า · วันที่ · เลขที่ · คำอธิบาย · หน่วย · จำนวนที่เบิก · ต้นทุน/หน่วย · ต้นทุนทั้งสิ้น · ราคาขาย/หน่วย · รวมทั้งสิ้น |

ตาราง: `BCStkIssue`, `BCStkIssueSub`. **กลุ่มรับคืนเบิก (5645–5648)** เป็นโครงเดียวกัน (ตัวกรอง *วันที่ · เลขที่ · สินค้า · ผู้เบิก · แผนก*; 5645 เพิ่ม *เรียงตาม · โครงการ · แสดงต้นทุน/หน่วย*):

| เมนู | คลาส (key) | หัวกลุ่ม | หมายเหตุคอลัมน์ |
|---|---|---|---|
| 5645 รับคืนเบิก-วันที่ | `CICRepBCIssueRetDocDateView` (85300) | วันที่/เลขที่ | วันที่ · เลขที่ · คำอธิบาย · สินค้า · ผู้เบิก · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุน/หน่วย · ต้นทุนทั้งสิ้น (+ รหัส/ชื่อผู้เบิก · แผนก · Project · Serial) |
| 5646 รับคืนเบิก-ผู้เบิก | `CICRepBCIssueRetPersonView` (85302) | ผู้เบิก | วันที่ · เลขที่ · สินค้า · คำอธิบาย · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · Serial |
| 5647 รับคืนเบิก-แผนก | `CICRepBCIssueRetDepartView` (85303) | แผนก | วันที่ · เลขที่ · สินค้า · ผู้เบิก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · Serial |
| 5648 รับคืนเบิก-สินค้า | `CICRepBCIssueRetItemView` (85304) | สินค้า | วันที่ · เลขที่ · คำอธิบาย · ผู้เบิก · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · ต้นทุนทั้งสิ้น · Serial |

ตาราง: `BCStkIssRetSub`. ⚠ **กรอบกล่องเงื่อนไขของกลุ่มรับคืนเบิกยังใช้ชื่อหน้าต่างของ "เบิก"** (คัดลอกมา) เช่น «รายงานการเบิกสินค้า,วัตถุดิบ - วันที่» — BC ต้องตั้งชื่อตามเมนู ไม่ลอกชื่อหน้าต่าง.

**กลุ่มรับสินค้าสำเร็จรูป (5649–5651)**

| เมนู | ชื่อ | คลาสที่เปิดจริง (key) | หมายเหตุ |
|---|---|---|---|
| 5649 | รับสินค้าสำเร็จรูป-วันที่ | `CICRepBCStkFGReceiveDateView` (85401) | ตัวกรอง: เรียงตาม · วันที่ · แผนก · คลัง · ที่เก็บ · สินค้า (จาก-ถึง) · แสดง Serial. คอลัมน์: วันที่ · รหัสสินค้า · ชื่อ · เลขที่เอกสาร · แผนก · คลัง · ที่เก็บ · หน่วย · จำนวน · **ต้นทุนที่เบิก** (`COST_ITEM`) · จำนวนเงิน · Serial (รหัสควบคุม/ทะเบียน/ตัวถัง/เลขเครื่อง/สถานะ) · หมายเหตุ. ตาราง `BCFinishGoodsSub` |
| 5650 | รับสินค้าสำเร็จรูป-แผนก | **ต่อผิดคลาส**: `CICRepItemSetView` (86005, แนวตั้ง, หน้าต่าง «รายงานรายละเอียดสินค้า-ประเภท») = *รายงานสินค้าชุดและส่วนประกอบ* | ตัวกรอง: สินค้า · หมวด · กลุ่ม · ประเภท · ยี่ห้อ · รูปแบบสินค้า (จาก-ถึง). คอลัมน์: รหัส/ชื่อสินค้าชุด → รหัส/ชื่อส่วนประกอบ · หน่วย · ราคาขาย/หน่วย · จำนวน · ราคารวม · ประเภทการขาย · ประเภทภาษี (ตรงกับ `BCITEMComponents` แบบที่ 2 — หัวข้อ 6.2) |
| 5651 | รับสินค้าสำเร็จรูป-สินค้า | **ต่อผิดคลาส**: `CICRepItemSetPromotionView` (86008) = *รายงานโปรโมชั่นสินค้าชุด* | ตัวกรอง: กลุ่มสินค้า · สินค้า · วันที่ · กลุ่มลูกค้า · พนักงานที่แก้ไข (จาก-ถึง). คอลัมน์: รหัส/ชื่อสินค้า · หน่วย · จากจำนวน/ถึงจำนวน · จากวันที่/ถึงวันที่ · กลุ่มลูกค้า (จาก/ถึง) · ราคาโปรโมชั่น · ส่วนลดรายการ · ประเภทภาษี · สินค้าแถม (รหัส/ชื่อ/หน่วย/จำนวน) · จากมูลค่าสุทธิ/ถึงมูลค่าสุทธิ |

### 7.5 รายงานตรวจนับ · ปรับปรุง · ประเมินการรับสินค้า

| เมนู | ชื่อ | คลาส | ตัวกรอง | คอลัมน์ | ตาราง |
|---|---|---|---|---|---|
| 5652 | พิมพ์เอกสารเพื่อตรวจนับสินค้า | `CICRepBCStkInspectWHView` | คลัง · วันที่ · เลขที่เอกสาร · สินค้า (จาก-ถึง) | คลัง · ที่เก็บ · วันที่ · เลขที่ → รหัสสินค้า · ชื่อ · หน่วย · **จำนวนในระบบ** (`QUANTITY_FROM_SYSTEM`) · **จำนวนที่ตรวจนับ** (`QUANTITY_FOR_INSPECT` — ช่องว่างให้เขียนมือ) | BCStkInspectSub |
| 5653 | รายงานใบตรวจนับสินค้าพร้อมรายการย่อย | `CICRepBCStkInspectItemView` | สินค้า · คลัง · วันที่ · เลขที่ (จาก-ถึง) | จัดตามสินค้า: รหัส/ชื่อ · คลัง · ที่เก็บ · วันที่ · เลขที่ · หน่วย · จำนวนในระบบ · จำนวนตรวจนับ | BCStkInspectSub, BCItem |
| 5654 | รายงานผลต่างจากการตรวจนับ | `CICRepBCStkInspectResultView` | วันที่ · เลขที่เอกสาร · คลัง · สินค้า (จาก-ถึง) | รหัส/ชื่อสินค้า · คลัง · ที่เก็บ · วันที่ · เลขที่ · หน่วย · **จำนวนตามระบบ** (`QUANTITY`) · **จำนวนที่ตรวจนับได้** (`STKCHKQTY`) · **ผลต่าง** (`VARSTKCHKQTY`) · หมายเหตุ | BCStkInspectSub, BCItem |
| 5655 | รายงานการปรับปรุงยอดสินค้า | `CICRepBCStkAdjustView` | วันที่ · เลขที่เอกสาร · คลัง · สินค้า (จาก-ถึง) | วันที่ · เลขที่ · รหัส/ชื่อสินค้า · คลัง · ที่เก็บ · หน่วย · **จำนวนปรับ** · **ต้นทุนที่ปรับ** · ราคา/หน่วย · เครื่องหมาย **+/-** · ยอดรวม | BCStkAdjustSub, BCItem |
| 5656 | รายงานประเมินการรับสินค้า | `CICRepBCAssessGoodsView` (85512) | วันที่ · เจ้าหนี้ · กลุ่มสินค้า · สินค้า (จาก-ถึง) | เจ้าหนี้ (รหัส+ชื่อ) → กลุ่มสินค้า → รหัส/ชื่อสินค้า · **เลขที่ใบสั่งซื้อ** (`ON_PO1`) · **วันที่ตรวจรับ** · วันที่รับสินค้า · จำนวนที่รับ · **สมบูรณ์ · ไม่ดี · แตก** · **เกินกำหนด** (`OVERDUEDATE`) | BCGoodsRecChkSub, BCItem (ข้อมูลจากใบตรวจรับ 5600) |

### 7.6 รายงานวิเคราะห์สินค้า (กลุ่ม `BC5REP_IC2` เป็นหลัก)

| เมนู | ชื่อ | คลาส (key) | ตัวกรอง/ตัวเลือก | คอลัมน์ (ตามลำดับ) | หมายเหตุ |
|---|---|---|---|---|---|
| 5566 | รายงาน Lot สินค้า | `CIC2RepItemLotView` (86100) | เรียงตาม · วันที่ (จาก-ถึง) · ☐ ทั้งหมด · Select Config | หัวสินค้า: รหัส · ชื่อ · หน่วย · ประเภท · วิธีต้นทุน · **คงเหลือ (ปริมาณ/มูลค่า)** → ต่อ Lot: วันที่รับ · **เลขที่ Lot** · คลัง · เลขที่เอกสาร · **วันหมดอายุ** · ปริมาณ · ต้นทุน · มูลค่า | อ่าน `BCLotMaster/BCLotOut` ตามต้นทุน FIFO/Lot (หัวข้อ 5.3) |
| 5574 | รายงานสินค้าและวัตถุดิบ | `CIC2RepItemMaterailView` (86204) | วันที่ (จาก-ถึง) · วิทยุคลัง (รวม/แยก/รวมและอื่น ๆ) · ☐ รูปแบบกรมสรรพากร · ☐ แสดง Serial/IMEI · ☐ ทั้งหมด · Select Config | ต่อสินค้า: คลัง (รหัส/ชื่อ) · ยกมา(ปริมาณ) → วันที่ · เลขที่เอกสาร · เลขที่ใบกำกับภาษี · รายวัน · คำอธิบาย · **รับ(ปริมาณ/ต้นทุน/มูลค่า)** · **จ่าย(…)** · **คงเหลือ(…)** · รวมยอดเข้า/ออก/คงเหลือ | บัญชีคุมสินค้ารายตัว |
| 5578 | รายงานบัญชีคุมพิเศษสินค้า | `CIC2RepSpecialAccountView` (86208) | วันที่ (จาก-ถึง) · **วิทยุวิธีต้นทุน: FIFO / เฉลี่ย / มาตรฐาน** · วิทยุคลัง · ☐ **แสดงต้นทุนแฝง** · ☐ Serial/IMEI · ☐ รูปแบบกรมสรรพากร · ☐ ทั้งหมด · วิทยุ **ไม่ระบุ / สำนักงานใหญ่ / สาขา** · Select Config | เหมือน 5574 + คอลัมน์ **Lot**; ท้ายสินค้า: ยอดเข้า · ยอดออก · ยอดคงเหลือ; ตามคลัง: จำนวน · ต้นทุนเฉลี่ย · ยอดรวมมูลค่า | ตัวเลือกวิธีต้นทุนในรายงานเปลี่ยนวิธีคำนวณมูลค่าที่แสดง (ยังไม่ verify ว่าคำนวณใหม่ย้อนหลังหรืออ่านจาก SumOfCost) |
| 5571 | รายงานบัญชีคุมพิเศษสินค้าตามกลุ่มสินค้า | `CIC2RepSpecAccGroupView` (862010) | วันที่ (จาก-ถึง) · Select Config · ☐ รูปแบบกรมสรรพากร | หัวกลุ่มสินค้า (รหัส+ชื่อ) → สินค้า → ยอดเข้า/ออก/คงเหลือ → ต่อคลัง (จำนวน · ต้นทุนเฉลี่ย · มูลค่า) → รายวัน (วันที่ · เลขที่ · ใบกำกับ · ปริมาณ/ต้นทุน/มูลค่า) | |
| 5596 | รายงานเคลื่อนไหวสินค้าต้นทุนมาตรฐาน | `CICRepStandardCostView2` (86211) | **จากงวดที่** · เริ่มจากวันที่ · ถึงวันที่ · คลัง (จาก-ถึง) · สินค้า (จาก-ถึง) · ☐ Serial/IMEI · ☐ **แสดงรายการเคลื่อนไหวเป็นรายวัน** | สินค้า: รหัส · ชื่อ · หน่วย · ยอดเข้า · ยอดออก · ยอดคงเหลือ → รายวัน: วันที่ · เลขที่ · ใบกำกับ · รายวัน · คำอธิบาย · ปริมาณ · ต้นทุน(มาตรฐาน) · มูลค่า · ที่เก็บ · คลัง | เฉพาะสินค้า `CostType = 2` (ยังไม่ verify) |
| 5570 | รายงานเคลื่อนไหวสินค้า | `CIC2RepItemMovementView` (86200) | ☐ รายงานแบบสรุป · **รายงานแบบ: ตามปริมาณ / ตามมูลค่า** · **เรียงตาม: รหัสสินค้า / คลังสินค้า / กลุ่มสินค้า / ประเภทสินค้า / ยี่ห้อสินค้า** · ☐ ทั้งหมด · วันที่ (จาก-ถึง) · วิทยุคลัง (รวม/แยก/รวมและอื่น ๆ) · Select Config | สินค้า (รหัส · ชื่อ · หน่วย) → คลัง (รหัส/ชื่อ · ยกมา) → รายวัน: วันที่ · เลขที่ · ใบกำกับ · **SO/SR/PO** (เลขที่อ้างอิง) · รายวัน · **รับ(ปริมาณ/ต้นทุน/มูลค่า) · จ่าย(…) · คงเหลือ(…)** | ยอดยกมาอ่านจาก `BCStkPeriodBal`/`BCStkWHPeriodBal` ของงวดก่อน (ยังไม่ verify) |
| 5575 | รายงานสรุปเคลื่อนไหวปริมาณสินค้า | `CIC2RepItemMovementSumView` (86255) | รายงานแบบ (ตามปริมาณ/ตามมูลค่า) · **เรียงตาม: รหัสสินค้า / รหัสกลุ่ม / ประเภทสินค้า / ยี่ห้อ / หมวด / รูปแบบ** · ☐ ทั้งหมด · พิมพ์จาก/ถึงวันที่ · ☐ แสดงยอดยกมาของหน้า · วิทยุคลัง · Select Config | ลำดับที่ · รายการสินค้า (รหัส/ชื่อ/กลุ่ม/หน่วย) · **ยกมาต้นงวด (BEGIN) · ซื้อ (BUY) · ส่งคืนเจ้าหนี้ (STKRET) · รับคืน/ลดหนี้ลูกหนี้ (CN) · ขาย (SALE_OUT) · เพิ่มหนี้ลูกหนี้ (DB1) · เพิ่มหนี้เจ้าหนี้ (DB2) · เบิก (ISSUE) · รับคืนเบิก (ISSUERET) · รับสำเร็จรูป (FG) · รับเข้า-ปรับยกมา (IN) · ออก-ปรับขาด (OUT) · ปลายงวด (END)** (ปริมาณหรือมูลค่าตามที่เลือก) | **คอลัมน์ตรงกับ `BCStkPeriodBal` ทีละช่อง** (หัวข้อ 5.4.1) — ใช้เป็นสเปกยอดงวดของ BC |
| 5577 | รายงานสินค้าที่ไม่มีการเคลื่อนไหว | `CIC2RepItemNotMovementView` (86207) | **ณ วันที่** · เรียงตาม · ☐ ทั้งหมด · **จากจำนวนวัน–ถึง** (ช่วงวันที่ไม่เคลื่อนไหว) · ☐ **คำนวณเฉพาะรายการขายเท่านั้น** · รายงานแบบ · Select Config | รหัส/ชื่อสินค้า · หน่วย · จำนวน · ต้นทุนเฉลี่ย · มูลค่า · **รายการเข้าล่าสุด · วันที่รับล่าสุด · รายการออกล่าสุด · วันที่ออกล่าสุด · วันที่ขายล่าสุด · จำนวนวันที่ไม่เคลื่อนไหว** | |
| 5580 | รายงานยอดคงเหลือสินค้า-ตามสินค้า | `CIC2RepItemRemainView` (86300) | เรียงตาม (รหัสสินค้า / คลังสินค้า / …) · **ณ วันที่** (พิมพ์ถึงวันที่) · ☐ **แสดงยอดคงเหลือสินค้าที่เท่ากับศูนย์** · ☐ แสดง Lot · ☐ แสดงคลังสินค้า · ☐ แสดงที่เก็บสินค้า · ☐ แสดงยอดยกมาของหน้า · ☐ Display summary · ☐ ทั้งหมด · Select Config | สินค้า: รหัส · ชื่อ · ประเภท · กลุ่ม · วิธีต้นทุน · สถานะ · คิดภาษี · หน่วย · จำนวน · ต้นทุนเฉลี่ย · มูลค่า (+ จำนวนแยกตามหน่วย) → ต่อคลัง: จำนวน · ต้นทุนเฉลี่ย · ยอดรวมมูลค่า → ต่อ Lot: หมายเลขล็อต · วันที่เอกสาร · เลขที่ · วันหมดอายุ · ปริมาณ · ต้นทุน · มูลค่า | รายงานยอดคงเหลือหลักของ Champ |
| 5597 | รายงานยอดคงเหลือสินค้า-ตามคลัง,สินค้า | `CICRepItemBalanceByWarehouseItemView` (86305) | As of Date · คลัง (จาก-ถึง) ×2 · สินค้า (จาก-ถึง) · ☐ Display qty in packing format | สินค้า · ชื่อ · คลัง (รหัส/ชื่อ) · ที่เก็บ · หน่วย · ยอดคงเหลือ · มูลค่า | ตาราง `BCStkWarehouse` (ยอดปัจจุบัน ไม่ใช่ย้อนหลัง ณ วันที่ — ยังไม่ verify) |
| 5588 | รายงานยอดคงเหลือสินค้าที่ถึงจุดสั่งซื้อ | `CICRepBCStkQtyMaxView` (86401) *(ชื่อคลาสสลับกับ 5589; เนื้อหาตรงเมนู)* | กลุ่มสินค้า (จาก-ถึง) · สินค้า (จาก-ถึง) | กลุ่ม → รหัส/ชื่อสินค้า · หน่วย · **จุดสูงสุด · จุดต่ำสุด** · **ยอดค้างรับ** · **«คจ»** (`ITEM_ARREAR`) · **«คส»** (`SEND_REMAIN_QTY`) *(ตัวย่อสองคอลัมน์นี้ความหมายยังไม่ verify — น่าจะเป็นยอดค้างจ่าย/ค้างส่ง)* · **จุดสั่งซื้อ (`REORDERPOINT`)** · **ยอดคงเหลือ** · **จำนวนที่ควรสั่งซื้อ (`QTY_PURCHASE_LOW`)** · ต้นทุนเฉลี่ย | ใช้ค่าจุดสั่งซื้อ/จุดต่ำสุด/สูงสุดจากข้อมูลหลักสินค้า (หัวข้อ 2.4/2.6) |
| 5589 | รายงานยอดคงเหลือสินค้าที่ถึงจุดสูงสุด | `CICRepBCStkQtyMinView` (86402) | เหมือน 5588 | กลุ่ม → สินค้า · หน่วย · จุดต่ำสุด · จุดสูงสุด · ยอดคงเหลือ · **จำนวนที่เกินจุดสูงสุด** (`AMOUNT_OVER`) · ต้นทุนเฉลี่ย | |
| 5590 | รายงานประมาณการสั่งสินค้าในแต่ละเดือน | **ไม่มี handler** (`ID_ICSYS_ANALYREP_4_3 → ?`) | — | — | ไม่ทำงานใน Champ เอง — BC ไม่ต้องทำ |
| 5592 | รายงานสรุปยอดสินค้าค้างรับแยกตามเจ้าหนี้ | `CICRepBCStkRemainInOutDetailView` | เจ้าหนี้ · กลุ่มสินค้า · สินค้า (จาก-ถึง) | เจ้าหนี้ (รหัส+ชื่อ) → กลุ่มสินค้า → สินค้า · **เลขที่ใบสั่งซื้อ** · **วันที่สั่งซื้อ** · **กำหนดส่ง** · หน่วย · ราคา/หน่วย · **จำนวนที่สั่ง** · **ยอดค้างรับ** | `BCGoodsRecChkSub`, `BCItem` |
| 5591 | รายงานสรุปยอดสินค้าค้างรับ | `CICRepBCStkRemainInOutView` (86404) | กลุ่มสินค้า · สินค้า (จาก-ถึง) | สินค้า · กลุ่ม · หน่วย · **มูลค่าคงเหลือ** · **ยอดคงเหลือ** · **ค้างรับ** · **ค้างส่ง** · **Available For Sale** (คำนวณ — สูตรยังไม่ verify) | ยอดค้างรับ/ค้างส่งเก็บที่ `BCItem.RemainInQty/RemainOutQty` (หัวข้อ 5.4) |
| 5595 | รายงานสรุปยอดสินค้าค้างส่ง | **ไม่มี handler** (`ID_ICSYS_ANALYREP_4_8 → ?`) | — | — | ไม่ทำงานใน Champ เอง — ถ้า BC ต้องมี ให้ใช้ 5591 (มีคอลัมน์ค้างส่ง) |

**ข้อสังเกตรวม:** (ก) ตัวกรองสินค้า/วันที่/คลังเป็น "จาก–ถึง" ทั้งหมด; (ข) รายงานเคลื่อนไหวทั้งหมดมีสูตรเดียวกัน — `คงเหลือ = ยกมา + รับ − จ่าย` และหน่วยเป็นหน่วยสต็อก; (ค) รายงานที่ต้องใช้ตัวเลขย้อนหลัง (ยกมา ณ วันที่) พึ่ง `BCStkPeriodBal`/`BCStkWHPeriodBal` ที่เครื่องยนต์เขียนไว้ — BC ที่ไม่มีตารางยอดงวดจึงต้องคำนวณจากรายการเคลื่อนไหวหรือสร้างยอดงวดคู่กัน (หัวข้อ 8).

---

## 8. BC เทียบ Champ และสิ่งที่ต้องปรับ

> **ข้อเท็จจริงสถานะ BC (ตรวจแล้ว):** เมนูกลุ่ม IC ใน BC อยู่ที่ `frontend/src/lib/menu-data.ts:467-558` (39 รายการ + 5 รายการข้อมูลหลักที่ Champ ไม่ไว้ใต้ IC); **ทุกจอเอกสาร/ข้อมูลหลักของ IC เป็น "รอพัฒนา"** เพราะ `isMenuBackendRetired` คืนค่าจริงให้ route ในกลุ่ม ERP transaction/operations และ route ข้อมูลหลักที่ backend ยังไม่มี (`menu-screen-status.ts:13-17, 36-38`) · รายงาน IC ที่ "ข้อมูลพร้อม" มีเพียง 2 (`stock_balance_item`, `stock_balance_warehouse` → `GET /api/reports/inventory-valuation`; `erp-reports.ts:545-546, 915`) · เครื่องมือ 5621/5623 มีหน้าจอแต่ไม่มี endpoint (`TOOL_ENDPOINTS` ว่าง, `erp-tools.ts:147`) · backend ที่มีจริงคือ Go package `backend/internal/goapi/inventory` (route รับ/จ่าย/โอน/ปรับ/คืนขาย/คืนซื้อ + รายงานมูลค่า/stock card, `handler.go:17-33`; ผูกที่ `bootstrap.go:337-339`) ซึ่งออกแบบต้นทุนคนละแนวกับ Champ (หัวข้อ 8.2). ตามกฎเจ้าของ **GL เป็นขอบเขตปัจจุบัน** — blueprint นี้คือสเปกสำหรับงานคลังรอบถัดไป ไม่ใช่งานที่ต้องทำตอนนี้.

### 8.1 ตารางเทียบรายเมนู (68 รายการ เรียงตามหัวข้อ 1)

สถานะ: **เหมือน** = มีจอ+ข้อมูลตรง Champ · **ต่าง** = มีแต่ทำงาน/ฟิลด์ไม่ตรง · **ยังไม่มี** = ไม่มีเมนู/จอใน BC · **รอพัฒนา** = มีเมนูใน BC แต่จอแสดง "ยังไม่พร้อม" (backend ไม่มี).

| id | Champ | BC (menu id → route) | สถานะ | สิ่งที่ต้องปรับ (ให้เหมือน Champ) |
|---|---|---|---|---|
| 5602 | รายละเอียดสินค้า | `product` → `/product` | รอพัฒนา | ทำจอ 5 แท็บตามหัวข้อ 2 (ส่วนหัว+ยอดคงเหลือ+หน่วย/กลุ่ม+คลัง/ที่เก็บ+แท็บ 2_1…2_3) พร้อมกติกาบันทึก/ลบ (2.8); ฟิลด์ต้นทุนใช้ `CostType` 0/1/2 แบบ Champ |
| 5332 | กำหนดราคาขายสินค้า *(Champ ปิด)* | `product-sale-price` → `/inventory/selling-prices` | รอพัฒนา | Champ ซ่อนเมนูนี้ (ใช้ 5005 แทน) → ตัดสินใจว่าจะซ่อนตามต้นแบบหรือเก็บ; คำอธิบายใน BC ("ตารางราคาตามกลุ่มลูกค้า" `erp-operations.ts:162`) ไม่ตรงกับหน้าจอ Champ |
| 5005 | กำหนดราคาขาย/โปรโมชั่นสินค้า | `product-promotion` → `/promotionscreen` | รอพัฒนา | ทำตามหัวข้อ 3.1 (4 มุมมอง `BCPriceList` ราคา 1-8/ช่วงจำนวน/ช่วงวันที่/กลุ่มลูกค้า/สินค้าแถม) |
| 5624 | บันทึกสินค้ายกมา | `stock-balance` → `/transaction/stockbalance` | รอพัฒนา | ใช้ฟอร์มเอกสารมาตรฐานหัวข้อ 4.1; ผลลัพธ์ = MyType 0 ในเครื่องยนต์ (5.2); BC ใช้ "counterparty" ช่องเดียว (`erp-transaction.ts:59-68`) ต้องเปลี่ยนเป็นหัวเอกสารตามจริง |
| 5601 | ใบขอเบิกใช้สินค้า,วัตถุดิบ | `stock-issue-request` → `/inventory/issue-request` | รอพัฒนา | **ต้องไม่ตัดสต็อก** — BC ชี้ `apiPath` เดียวกับใบเบิกจริง (`stock-prickup-product`, `erp-transaction.ts:585`) ต้องแยกตารางเอกสาร (4.2) |
| 5625 | บันทึกเบิกใช้สินค้า,วัตถุดิบ | `stock-pickup` → `/transaction/stockpickupproduct` | รอพัฒนา | หัวเอกสาร: วันที่/ผู้เบิก/แผนก/โครงการ/Job + ลงบัญชี GL กลุ่ม 13 (5.6) + Serial/Lot (4.3) — BC มีแค่ "แผนกผู้ขอเบิก" |
| 5626 | บันทึกรับคืนสินค้า,วัตถุดิบ | `stock-return` → `/transaction/stockreturnproduct` | รอพัฒนา | ต้องอ้างบรรทัดใบเบิก (`IssueRefNo`, `reflinenumber`) เพื่อคืนต้นทุนเดิม (5.3); GL กลุ่ม 26 |
| 5627 | บันทึกรับสินค้าสำเร็จรูป | `stock-receive` → `/transaction/stockreceiveproduct` | รอพัฒนา | ต้นทุน = ต้นทุนที่ผู้ใช้ระบุ/จากสินค้าชุดประกอบ; GL กลุ่ม 27 (4.5) |
| 5631 | บันทึกใบขอโอนสินค้า | `stock-transfer-request` → `/inventory/transfer-request` | รอพัฒนา | **ต้องไม่ตัดสต็อก** — BC ชี้ `apiPath` เดียวกับใบโอนจริง (`erp-transaction.ts:596`) |
| 5628 | บันทึกโอนสินค้าระหว่างคลัง | `stock-transfer` → `/transaction/stocktransfer` | รอพัฒนา | 1 เอกสารสร้าง 2 แถว (โอนออก 13/โอนเข้า 12) ส่งต้นทุนข้ามกัน (5.2); ไม่ลง GL (4.7) |
| 5629 | บันทึกเอกสารเพื่อตรวจนับสินค้า | `stock-count-sheet` → `/inventory/count-sheet` | รอพัฒนา | **ต้องไม่ตัดสต็อก** — BC ชี้ `stock-adjustment` (`erp-transaction.ts:618`); ต้องเป็นเอกสารแช่ยอดในระบบ (`QUANTITY_FROM_SYSTEM`) |
| 5599 | บันทึกผลการตรวจนับสินค้า | `stock-count` → `/transaction/stockcount` | รอพัฒนา | บันทึกยอดนับ + คำนวณผลต่าง (ไม่ตัดสต็อก) แยกจาก 5630; BC ใช้ `stock-adjustment` ร่วมกัน (`:574`) |
| 5630 | บันทึกปรับปรุงสินค้าหลังตรวจนับ | `stock-adjust` → `/transaction/adjust` | รอพัฒนา | บรรทัด +/- (`AdjMark`) → MyType 10/11; GL กลุ่ม 34/44/45/46 ตาม `AdjustType` (5.6) |
| 5613 | บันทึกสินค้าชุด | `product-set` → `/inventory/product-sets` | รอพัฒนา | BOM ชั้นเดียว `BCITEMComponents` + ปรับราคาขาย 1 ของชุดอัตโนมัติ (6.1); ราคา/จำนวนต้อง > 0 |
| 5615 | ตรวจสอบสินค้าที่รวมเป็นชุดได้ | `product-set-assemble` → `/inventory/set-assembly` | รอพัฒนา | ประกอบ = เบิก (กลุ่ม 13) + รับสำเร็จรูป (กลุ่ม 27) 2 ใบ ต้นทุนชุด = ต้นทุนส่วนประกอบที่เบิก (6.3); คำอธิบายใน BC ใช้ได้ |
| 5616 | ตรวจสอบชุดที่แยกเป็นสินค้าได้ | `product-set-disassemble` → `/inventory/set-disassembly` | รอพัฒนา | แยก = เบิกตัวชุด + รับส่วนประกอบ (6.4) |
| 5614 | รายการย่อยสินค้าชุดแบบที่ 2 | `product-set-components` → `/inventory/set-components` | รอพัฒนา | Champ = สูตรแยกตาม SaleType × TaxType (6.2); BC บอก "หลายระดับ" (`erp-operations.ts:144`) → **แก้เป็นชั้นเดียว แยกตามประเภทการขาย/ภาษี** |
| 5618 | บันทึก Serial Number | `product-serial-registry` → `/productserialregistry` | รอพัฒนา | ทะเบียน `BCSerialMaster` + ประวัติ `BCSerialTrans.SaveFrom 0-10` + กติกาแก้/ลบตามสถานะ (6.5); BC มีแค่ flag `serialtrackingenabled` (`database.go:57`) |
| 5600 | ใบตรวจรับสินค้า | `goods-inspection` → `/inventory/goods-inspection` | รอพัฒนา | **ต้องไม่รับสินค้าเข้าสต็อก** — BC ชี้ `stock-receive-product` (`erp-transaction.ts:607`); ต้องเป็น `BCGoodsRecCheck/Sub` (สมบูรณ์+ไม่ดี+แตก=ยอดรับ) (6.6) |
| 5623 | กำหนดลำดับรายวัน | `stock-daily-sequence` → `/inventory/daily-sequence` | รอพัฒนา (ไม่มี API) | Champ = ผู้ใช้กำหนดเลขลำดับรายเอกสารต่อสินค้า/วัน แล้วสั่งคำนวณใหม่; BC เขียนเป็นกฎอัตโนมัติ "รับก่อนจ่าย" (`erp-tools.ts:115-119`) → เปลี่ยนให้เหมือน 6.7 (มีฟิลด์ `behindindex` ที่บรรทัดเอกสาร) |
| 5621 | คำนวณยอดสินค้าใหม่ | `reprocess` → `/rebuildstockscreen` | รอพัฒนา (ไม่มี API) | งาน backend (ไม่ใช่ browser) ตามหัวข้อ 5.5/6.8: ช่วงสินค้า+ช่วงงวด + "คำนวณเฉพาะยอดงวด" + นำเข้าคิวจากไฟล์ + สถานะ/ยกเลิกงาน |
| 5564 | ปรับปรุงราคาขายสินค้า *(Champ ปิด)* | `product-price-adjust` → `/inventory/price-adjustment` | รอพัฒนา | Champ ซ่อน (3.2) → ตัดสินใจซ่อนตามต้นแบบหรือเก็บ |
| 5603 | รายงานรายละเอียดสินค้า | — | ยังไม่มี | ทำตาม 7.2 (เรียงตามกลุ่ม/ยี่ห้อ/ประเภท/หมวด) |
| 5608 | รายงานรายละเอียดสินค้า-ตาม Supplier | — | ยังไม่มี | ทำตาม 7.2 |
| 5609 | รายงานรายการสินค้าแบบมี Serial | — | ยังไม่มี | ทำตาม 7.2 (ขึ้นกับ 5618) |
| 5610 | รายงานเคลื่อนไหว Serial Number | `serial-movement-report` → `/report/serialmovement` | รอพัฒนา (ไม่มี config จอ) | ทำตาม 7.2 |
| 5632 | รายงานรายการสินค้ายกมาต้นปี | — | ยังไม่มี | ทำตาม 7.3 |
| 5633 | รายงานการบันทึกยอดสินค้ายกมาต้นปี | — | ยังไม่มี | ทำตาม 7.3 |
| 5634 | โอนสินค้าระหว่างคลังและรายการย่อย | — | ยังไม่มี | ทำตาม 7.3 |
| 5635 | โอน-ตามคลังโอนออก | — | ยังไม่มี | Champ ต่อผิดคลาส (รับคืนเบิก) → ทำรายงานโอนตามคลังให้ถูก ไม่ลอกบั๊ก |
| 5638 | โอน-ตามคลังโอนเข้า | — | ยังไม่มี | ทำตาม 7.3 |
| 5636 | โอน-ตามสินค้า | — | ยังไม่มี | Champ ต่อผิดคลาส (ใบขอโอน) → ทำรายงานโอนตามสินค้าให้ถูก |
| 5637 | โอน-ตามวันที่ | — | ยังไม่มี | ทำตาม 7.3 |
| 5639 | เบิก-วันที่เบิก | — | ยังไม่มี | ทำตาม 7.4 |
| 5640 | เบิก-ผู้เบิก | — | ยังไม่มี | ทำตาม 7.4 |
| 5641 | เบิก-แผนก | — | ยังไม่มี | ทำตาม 7.4 |
| 5642 | เบิก-สินค้า | — | ยังไม่มี | ทำตาม 7.4 |
| 5643 | เบิก-โครงการ | — | ยังไม่มี | ทำตาม 7.4 |
| 5645 | รับคืนเบิก-วันที่ | — | ยังไม่มี | ทำตาม 7.4 (ตั้งชื่อหน้าต่างให้ถูก) |
| 5646 | รับคืนเบิก-ผู้เบิก | — | ยังไม่มี | ทำตาม 7.4 |
| 5647 | รับคืนเบิก-แผนก | — | ยังไม่มี | ทำตาม 7.4 |
| 5648 | รับคืนเบิก-สินค้า | — | ยังไม่มี | ทำตาม 7.4 |
| 5649 | รับสินค้าสำเร็จรูป-วันที่ | — | ยังไม่มี | ทำตาม 7.4 |
| 5650 | รับสำเร็จรูป-แผนก | — | ยังไม่มี | Champ ต่อผิดคลาส (รายงานสินค้าชุดและส่วนประกอบ) → ทำรายงานรับสำเร็จรูปตามแผนกให้ถูก; รายงานสินค้าชุดแยกเป็นอีกรายการ (ถ้าต้องการ — ต้องถามลุงจืด) |
| 5651 | รับสำเร็จรูป-สินค้า | — | ยังไม่มี | Champ ต่อผิดคลาส (โปรโมชั่นสินค้าชุด) → ทำตามชื่อเมนู |
| 5652 | พิมพ์เอกสารเพื่อตรวจนับสินค้า | — | ยังไม่มี | ทำตาม 7.5 (มีช่องเขียนมือ "จำนวนที่ตรวจนับ") |
| 5653 | ใบตรวจนับสินค้าพร้อมรายการย่อย | — | ยังไม่มี | ทำตาม 7.5 |
| 5654 | ผลต่างจากการตรวจนับ | `stock-count-variance-report` → `/report/stockcountvariance` | รอพัฒนา (ไม่มี config จอ) | คอลัมน์: ตามระบบ · ตรวจนับได้ · ผลต่าง (7.5) |
| 5655 | การปรับปรุงยอดสินค้า | — | ยังไม่มี | ทำตาม 7.5 |
| 5656 | ประเมินการรับสินค้า | — | ยังไม่มี | ทำตาม 7.5 (ขึ้นกับ 5600) |
| 5566 | รายงาน Lot สินค้า | — | ยังไม่มี | ทำตาม 7.6 (ขึ้นกับต้นทุน Lot/FIFO) |
| 5574 | รายงานสินค้าและวัตถุดิบ | — | ยังไม่มี | ทำตาม 7.6 (บัญชีคุมสินค้าตามรูปแบบกรมสรรพากร — ตรวจแบบฟอร์มจากแหล่งทางการก่อนทำ) |
| 5578 | บัญชีคุมพิเศษสินค้า | `stock-movement-cost` → `/report/stockmovementcost` | จอพร้อม / ข้อมูลยังไม่พร้อม | เพิ่มตัวเลือกวิธีต้นทุน/คลัง/ต้นทุนแฝง/Serial/ประเภทสำนักงาน และคอลัมน์ Lot (7.6) |
| 5571 | บัญชีคุมพิเศษตามกลุ่มสินค้า | — | ยังไม่มี | ทำตาม 7.6 |
| 5596 | เคลื่อนไหวสินค้าต้นทุนมาตรฐาน | — | ยังไม่มี | ทำตาม 7.6 (เฉพาะสินค้า CostType 2) |
| 5570 | เคลื่อนไหวสินค้า | `product-movement` → `/report/reportstockmovement` | จอพร้อม / ข้อมูลยังไม่พร้อม | เพิ่มรายงานแบบสรุป/ตามปริมาณ-มูลค่า/เรียงตาม 5 แบบ/โหมดคลัง 3 แบบ + คอลัมน์ SO,SR,PO (7.6) |
| 5575 | สรุปเคลื่อนไหวปริมาณสินค้า | — | ยังไม่มี | ทำตาม 7.6 — ต้องมี `BCStkPeriodBal` ก่อน (5.4) |
| 5577 | สินค้าที่ไม่มีการเคลื่อนไหว | `no-movement-stock-report` → `/report/nomovementstock` | รอพัฒนา (ไม่มี config จอ) | ตัวกรอง "จำนวนวันจาก-ถึง" + "เฉพาะรายการขาย" + 6 คอลัมน์วันที่ล่าสุด (7.6) |
| 5580 | ยอดคงเหลือสินค้า-ตามสินค้า | `stock-balance-item` → `/report/stockbalanceitem` | **ต่าง** (API พร้อม) | BC แสดง 6 คอลัมน์ (รหัส/ชื่อ/หน่วย/จำนวน/ต้นทุนเฉลี่ย/มูลค่า, `erp-reports.ts:39-44`) และไม่มี "ณ วันที่", แสดงศูนย์, Lot, คลัง, ที่เก็บ, สรุป → เพิ่มตามหัวข้อ 7.6 (5580) |
| 5588 | ยอดคงเหลือที่ถึงจุดสั่งซื้อ | `low-stock-alert` → `/report/lowstock` | จอพร้อม / ข้อมูลยังไม่พร้อม | คอลัมน์: จุดสูงสุด/ต่ำสุด · ค้างรับ · จุดสั่งซื้อ · คงเหลือ · จำนวนที่ควรสั่ง (7.6) |
| 5589 | ยอดคงเหลือที่ถึงจุดสูงสุด | `max-stock-report` → `/report/maxstock` | รอพัฒนา (ไม่มี config จอ) | คอลัมน์: จุดต่ำสุด/สูงสุด · คงเหลือ · จำนวนที่เกิน (7.6) |
| 5597 | ยอดคงเหลือ-ตามคลัง,สินค้า | `stock-balance-warehouse` → `/report/stockbalancewarehouse` | **ต่าง** (API พร้อม) | เพิ่ม As of Date, ช่วงคลัง×2, ช่วงสินค้า, "แสดงจำนวนตามบรรจุภัณฑ์", คอลัมน์ที่เก็บ (7.6) |
| 5590 | ประมาณการสั่งสินค้าแต่ละเดือน | — | — | Champ ไม่มี handler (ทำงานไม่ได้) → **ไม่ต้องทำ** |
| 5592 | สรุปยอดค้างรับแยกตามเจ้าหนี้ | — | ยังไม่มี | ทำตาม 7.6 (อ่านใบสั่งซื้อ/ใบตรวจรับ) |
| 5591 | สรุปยอดสินค้าค้างรับ | `pending-receive-report` → `/report/pendingreceive` | รอพัฒนา (ไม่มี config จอ) | ทำตาม 7.6 (มีคอลัมน์ค้างรับ+ค้างส่ง+Available) |
| 5595 | สรุปยอดสินค้าค้างส่ง | `pending-delivery-report` → `/report/pendingdelivery` | รอพัฒนา (ไม่มี config จอ) | Champ ไม่มี handler → ใช้คอลัมน์ค้างส่งของ 5591; ถ้าจะเก็บเมนูนี้ต้องถามลุงจืด |
| 5612 | รายงานสินค้าใหม่ *(ปิด)* | — | ยังไม่มี | Champ ซ่อนเมนู → ไม่ต้องทำเว้นลุงจืดสั่ง |
| 5657 | รายงานการกำหนดราคาขายสินค้า *(ปิด)* | — | ยังไม่มี | Champ ซ่อนเมนู → ไม่ต้องทำเว้นลุงจืดสั่ง (ข้อมูลเดียวกับ 5005) |

**เมนู BC ที่ Champ ไม่มีใต้ IC (ต้องตรวจตำแหน่ง/ความจำเป็น — กฎ Strict Parity):** `barcode` `/productbarcode` · `product-unit` `/productunit` · `product-group` `/productgroup` · `brand` `/masterbrandscreen` · `warehouse` `/productwarehousescreen` (`menu-data.ts:472-477`; Champ จัดการหน่วย/กลุ่ม/ยี่ห้อ/คลัง/ที่เก็บในกลุ่มระบบ/ตั้งค่า ไม่ใช่เมนู IC — หัวข้อ 1 ข้อสังเกต; ยังไม่ verify ตำแหน่งในเมนูโมดูลอื่นของ Champ) · `inventory` `/report/reportdedebistockbalance` ("รายงานสินค้าคงเหลือ" ไม่มีใน Champ — ใกล้เคียง 5580).

### 8.2 ความต่างเชิงสถาปัตยกรรมที่ต้องตัดสินใจก่อนลงมือ

| # | หัวข้อ | Champ | BC ตอนนี้ (หลักฐาน) | ข้อเสนอให้ตรง Champ / ข้อควรระวัง |
|---|---|---|---|---|
| 1 | **เอกสารขอ/ตรวจ vs เอกสารที่ตัดสต็อก** | ใบขอเบิก/ขอโอน/ใบตรวจรับ/เอกสารตรวจนับ/ผลตรวจนับ = ไม่กระทบสต็อก มีแต่เบิก/รับคืน/สำเร็จรูป/ปรับปรุง/ยกมา/โอน (5.2) | 5 เอกสารข้างต้นใช้ `apiPath` ร่วมกับเอกสารที่ตัดสต็อกจริง (`erp-transaction.ts:572-620`; ตัวสะกด `stock-prickup-product` ที่ `:83, :585`) | แยกตารางเอกสาร 5 ชนิดให้ไม่ผ่าน engine สต็อก — ความเสี่ยงสูง (ใบขอเบิกตัดสต็อกจริง) |
| 2 | **วิธีต้นทุน** | `BCItem.CostType` 0 เฉลี่ย / 1 FIFO-Lot / 2 มาตรฐาน; ต้นทุนเฉลี่ยเก็บที่ระดับสินค้า (ไม่แยกคลัง); เฉลี่ยเปลี่ยนเมื่อ "เข้า" เท่านั้น; สินค้าบริการข้าม (5.3) | `productcostingconfig` ให้ 5 วิธี (movingaverage/FIFO/LIFO/FEFO/standard) + `costbywarehouse` + ชั้นต้นทุน `inventorycostlayers` + ต้นทุนแฝง `inventorylandedcosts` (`database.go:52-100, 204-235`; `costing/*.go`) | LIFO/FEFO/ต้นทุนแฝง/marketplace stock **ไม่มีใน Champ** → ตามกฎ Strict Parity ห้ามเพิ่มเอง: เสนอลุงจืดตัดสินใจ (คง/ซ่อน). วิธี 0/1/2 ของ Champ ต้องทำให้ตรงสูตรหัวข้อ 5.3 (โดยเฉพาะ รับคืนใช้ต้นทุนของเอกสารอ้างอิง) |
| 3 | **ตารางยอดงวด** | `BCStkPeriodBal` (ต่อสินค้า×งวด แยก 12 ชนิดรายการ) + `BCStkWHPeriodBal` (ต่อ คลัง×ที่เก็บ×หน่วย×งวด) + `BCPeriod` | มี `inventoryaccountingperiods` (เปิด/ปิดงวดอย่างเดียว) และ `inventorycosttransactions.accountingperiod` (`database.go:155-185, 236-250, 263`) ไม่มียอดต้น/ปลายงวดแยกชนิด | สร้างยอดงวดตามหัวข้อ 5.4 (หรือวิวที่ให้ผลเท่ากัน) — 5570/5574/5575/5578 ต้องใช้ |
| 4 | **ยอดคงเหลือตามคลัง** | `BCStkWarehouse` คีย์ (สินค้า, คลัง, ที่เก็บ, **หน่วยของเอกสาร**) เก็บ Qty/Amount/ReserveQty/RemainIn/Out | `inventorystockbalances` คีย์ (holdingcode, itemcode, whcode, locationcode) ไม่มีมิติหน่วย (`database.go:101-116`) | เก็บหน่วยฐานได้ แต่รายงาน 5597 "แสดงตามบรรจุภัณฑ์" และสินค้าหลายหน่วย (`UnitType=1`, `BCStkPacking` 25 งวด) ต้องรองรับ |
| 5 | **ลำดับรายวัน & คำนวณใหม่** | `behindindex` ที่แถวเอกสาร + `ProcessStock` คิว + `StockReprocess` เรียง `DocDate, BehindIndex, Atype, MyType, RowOrder` (5.5, 6.7) | ไม่มี; จอ 5623 บอกกฎ "รับก่อนจ่าย" (`erp-tools.ts:115-119`) — ไม่ใช่พฤติกรรม Champ | เพิ่มฟิลด์ลำดับรายวัน + งาน backend คำนวณใหม่ (ไม่ใช่ Kafka — ใช้คิวใน PostgreSQL ตามกฎ Pure PostgreSQL) |
| 6 | **ลงบัญชี GL ของเอกสารสต็อก** | รูปแบบ GL ต่อเอกสาร (กลุ่ม 24/25/26/34/44/45/46) + กลุ่มบัญชีสินค้า `BCItemAccGroup` (ItemAccount/CostAccount/RevenueAccount) (5.6) | แพ็กเกจ `inventory` ไม่มีการเชื่อม GL (ค้นหา `gl_/journal` ไม่พบ) | ผูกผ่าน **คุณสมบัติของกลุ่มบัญชีสินค้า** ไม่ใช่เงื่อนไขตามรหัสบัญชี (กฎ no-account-code-conditions); ใช้ journal ของ BC (`gl_journals`) แบบเดบิต=เครดิตและ idempotent |
| 7 | **ความแม่นยำตัวเลข** | `double` + `%lf` ทุกจุด; ทศนิยมตามตั้งค่าบริษัท `QtyPoint/CostPoint/PricePoint/AmountPoint` | DDL ใช้ `NUMERIC(18,4)` แต่ Go model เป็น `float64` (`models/models.go:97-101`) | แก้เป็น decimal (shopspring/apd) ตาม Rule 20 + เก็บนโยบายปัดเศษ (ต้นทุน/หน่วยปัด `CostPoint`, มูลค่าบรรทัดปัด `AmountPoint`) + บันทึกส่วนต่างปัดเศษ; จำนวนทศนิยมต้องตั้งค่าได้ต่อบริษัทแบบ Champ |
| 8 | **รายงาน** | 46 รายการ + กรอบ IC2 (group-by, Select Config, โหมดคลัง, ตัวเลือกรูปแบบกรมสรรพากร) (7.1) | 12 เมนูรายงาน, `ErpReportConfig` ระบุแค่คอลัมน์ ไม่มีนิยามตัวกรอง (`erp-reports.ts:28-46`); API พร้อมเพียง 2 | ทำกรอบรายงานกลางที่ backend (ตัวกรองจาก-ถึง + group-by + ชุดเงื่อนไขที่บันทึกได้) แล้วค่อยเพิ่มรายงานทีละกลุ่ม; ทุกรายงานคำนวณที่ backend |
| 9 | **Serial / Lot** | `BCSerialMaster/Trans` (SaveFrom 0-10), `BCLotMaster/BCLotOut`, ตัวเลือกบริษัท `nLotStatus` (4.3, 5.3, 6.5) | flag ใน `productcostingconfig`; `inventorycostlayers` มี `lotnumber` | ออกแบบ Serial/Lot ให้อ้างเอกสารต้นทาง (docno + บรรทัด) และบังคับกติกาแก้/ลบตามสถานะ (6.5) |
| 10 | **เอกสารซื้อ/ขาย กระทบสต็อก** | เอกสารซื้อ/ขาย/ลดหนี้/เพิ่มหนี้ ส่งเข้าเครื่องยนต์เดียวกัน (MyType 2-7) | Go API มี `sales-return`/`purchase-return` แยก | เครื่องยนต์สต็อกต้องรับจากเอกสารซื้อ-ขายด้วย (ประสานทีมซื้อ/ขาย — นอกขอบเขตไฟล์นี้) |

### 8.3 ตารางข้อมูล Champ → BC (ข้อเสนอชื่อ PostgreSQL ตัวพิมพ์เล็ก `snake_case`)

> เป็น **ข้อเสนอ** ให้ตัดสินใจ ไม่ใช่การออกแบบที่ยืนยัน; ชื่อฟิลด์ใน Champ ตามหัวข้อ 2-6. ใช้ NUMERIC(p,s) สำหรับจำนวน/ต้นทุน/มูลค่าทั้งหมด.

| Champ | ข้อเสนอ BC | หมายเหตุ |
|---|---|---|
| BCItem (ข้อมูลหลัก+ยอดรวม) | `items` (+ `item_stock_summary` สำหรับ `stock_qty/amount/average_cost/remain_in_qty/remain_out_qty/reserve_qty`) | แยกข้อมูลหลักกับยอดเพื่อ transaction ชัด |
| BCItemUnit / BCStkPacking | `item_units` / `item_packings` (อัตราแปลง, ยอดงวด 1-25 → ย้ายไปยอดงวด) | |
| BCStkWarehouse | `stock_warehouse_balances` | คีย์ item+wh+shelf(+unit) |
| BCStkPeriodBal / BCStkWHPeriodBal / BCPeriod | `stock_period_balances` / `stock_wh_period_balances` / `periods` | 5.4 |
| BCStkBalanceSub, BCStkIssue(+Sub), BCStkIssRetSub, BCFinishGoods(+Sub), BCStkAdjust(+Sub), BCStkTransfer(+Sub3) | เอกสารละ header+lines ตามหัวข้อ 4 (`stock_balance_docs`, `stock_issue_docs`, …) พร้อม `behind_index`, `gl_format`, `is_cancel` | ตารางเอกสารต้องไม่ลบ/แก้เงียบ ๆ — ใช้ reversal (Rule 20 audit) |
| BCStkInspect(+Sub), BCStkInspectResult | `stock_count_sheets`, `stock_count_results` | ไม่กระทบสต็อก |
| BCITEMComponents | `item_set_components` (+ `sale_type`, `tax_type`) | 6.1-6.2 |
| BCSerialMaster / BCSerialTrans | `serial_masters` / `serial_trans` | 6.5 |
| BCLotMaster / BCLotOut | `lot_masters` / `lot_outs` | 5.3 |
| BCGoodsRecCheck(+ChkSub) | `goods_receipt_checks` (+ lines) | 6.6 |
| ProcessDoc / ProcessStock / ReportStatus | `stock_process_queue` / `stock_reprocess_queue` / `report_requests` | คิวใน PostgreSQL (ไม่ใช้ Kafka/Redis) |
| BCItemAccGroup | `item_account_groups` (item_account, cost_account, revenue_account) | 5.6 |
| BCPriceList / BCItemStandardCost | `price_lists` / `item_standard_costs` | 3.1, 5.3 |

### 8.4 จุดชนกับกฎเจ้าของ (flag ไม่เสนอทำ)

1. **ขอบเขต GL เท่านั้น (ปัจจุบัน):** IC ทั้งหมดอยู่นอกขอบเขตตอนนี้ — ไม่ต้องลงมือจากไฟล์นี้จนกว่าลุงจืดสั่ง; ของที่เกี่ยวกับ GL (กลุ่มบัญชีสินค้า/รูปแบบ GL ของเอกสารสต็อก) ให้ทำร่วมกับทีม GL ตอนเปิดงานคลัง.
2. **PostgreSQL เท่านั้น:** กลไก `ReportStatus`/`BCIC2Report` (ตารางชั่วคราวโดยงานฝั่งเซิร์ฟเวอร์) ต้องแปลงเป็น query/materialized view หรือคิวในฐานข้อมูลเดียว ไม่เพิ่ม worker/Kafka/Redis.
3. **ห้ามเพิ่มฟังก์ชันที่ Champ ไม่มี:** LIFO, FEFO, ต้นทุนแฝง (landed cost), marketplace stock/price dimension และ `inventoryvariances` ใน Go API ปัจจุบันไม่มีใน Champ — ต้องถามลุงจืดก่อนคงไว้.
4. **ห้ามเงินเดือน:** ฟิลด์ "ผู้เบิก/พนักงานรับสินค้า/พนักงานขาย" อ้างทะเบียนพนักงาน/ผู้ใช้ (ไม่ใช่ payroll) — ทำได้โดยไม่ขัดกฎ.
5. **ห้ามเงื่อนไขตามรหัสบัญชี:** การลงบัญชีสต็อกต้องใช้คุณสมบัติกลุ่มบัญชีสินค้า (ดู 8.2 ข้อ 6).
6. **ภาษาไทย/อังกฤษเท่านั้น:** ป้ายทั้งหมดผ่าน key + `languages.tsv` (ไม่ฝังไทยในโค้ด); คำ «KEY» ของ Champ ที่ยังไม่ resolve ต้องให้ลุงจืดยืนยันถ้อยคำ.

### 8.5 สิ่งที่ Champ ทำ "ผิด/แปลก" ห้ามลอก

- รายงาน 5635/5636/5650/5651 ต่อกับคลาสผิด; 5590/5595 ไม่มี handler; ชื่อคลาส QtyMax/QtyMin สลับกับ 5588/5589; ชื่อหน้าต่างของกลุ่มรับคืนเบิกยังเป็น "เบิก"; หน้าต่าง 5588/5589 ตั้งชื่อ «รายงานสถานะเจ้าหนี้» (คัดลอก).
- MyType 20 ซ้ำความหมาย (ใบสั่งขายในจอลำดับรายวัน ↔ รับคืนสำเร็จรูปในเครื่องยนต์) และยอดงวดของ 20 บวกทั้งที่สต็อกถูกหัก (5.4.1).
- `DetailCalc` ของจอสินค้าชุดสลับดัชนีราคา/จำนวน (ผลคูณเหมือนเดิม จึงไม่เกิดผิด) `icfrmitemset.cpp:614-615`.
- ปุ่ม Default ของลำดับรายวัน "ย้อนกลับไม่ได้" โดยไม่มีบันทึกก่อนหน้า (6.7) → BC ต้องเก็บประวัติลำดับ.
- ค่าเริ่มต้นราคาส่วนประกอบสินค้าชุดอ่านราคาขาย 1 ของสินค้า (หรือ AverageCost ใน CHM) และบังคับราคา > 0 → ส่วนประกอบแจกฟรีทำไม่ได้ (6.1).
- เขียนยอดทีละตารางไม่อยู่ใน transaction เดียว, SQL ต่อสตริง, double (5.7).

---

## 9. ยังไม่ verify / ช่องว่างของหลักฐาน

> รายการที่ **ไม่ได้ยืนยัน** (อย่านำไปใช้เป็นสเปกจนกว่าจะตรวจต่อ) — เรียงตามความสำคัญต่อการออกแบบ BC

1. **ผลต่าง `BCStkAdjust.SumAmount` vs `SumAmount2`** — ชื่อ/ความหมาย "ยอดปรับบวก/ลบ" ในการลงบัญชี (`gltlsprocess.cpp:9132-9192`) ไม่ตรงกับชื่อช่องบนจอ (4.10); ความหมายค่า `nStatus` ใน `PostStkAdjustSub` (`:11951-12036`) ไม่แน่ชัด → ใช้ผลรวมจากรายการย่อย (`AdjMark`) แทนเมื่อสร้างใหม่.
2. **ยอดงวดของ MyType 20** — เครื่องยนต์หักสต็อก (`nMark = −1`) แต่บวก `InQty/InAmount` ของ `BCStkPeriodBal` (`bcdumpprocjob.cpp:2100-2183`); ไม่พบแหล่งข้อมูลที่ส่งแถว MyType 20 เข้าเครื่องยนต์ (จอลำดับรายวันใช้ 20-22/30-31 สำหรับใบสั่งซื้อ/ขายแบบอื่น) — น่าจะเป็นโค้ดตกค้าง.
3. **ตารางโอนของรายงาน** — รายงานอ่าน `BCStkTransfer/BCStkTransfSub` แต่เครื่องยนต์อ่านโอนจาก `BCStkTransfSub3` (5.2, 7.3): ยังไม่ยืนยันว่าเป็นตารางเดียวกันหรือวิว.
4. **ใครเติมตารางชั่วคราว `BCIC2Report`/`BCIC2Report2`** และสูตรจริงของรายงาน IC2 (7.1, 7.6) — เห็นเพียงด้าน "ส่งคำขอ" (`ReportStatus`); งานเซิร์ฟเวอร์ที่คำนวณไม่อยู่ในที่ตรวจ; ตัวเลือก "วิธีต้นทุน FIFO/เฉลี่ย/มาตรฐาน" ของ 5578 ยังไม่ทราบว่าคำนวณใหม่หรืออ่านต้นทุนที่บันทึก.
5. **ป้ายภาษาไทยที่เป็นคีย์ «KEY»** ในไฟล์ทรัพยากรภายนอก (ไม่อยู่ใน repo): ป้ายที่ไม่ resolve ในหัวข้อ 2 (บางแท็บ), 7 (ป้าย filter ที่ขึ้น `?`), และตัวย่อ «คจ»/«คส» ของ 5588 — ต้องให้ลุงจืดยืนยันถ้อยคำก่อนขึ้นจอ.
6. **สูตร "Available For Sale" ของ 5591**, ความหมาย `FIRSTDATE/RECLASTDATE/PAYLASTDATE` ของ 5612, ที่มาค่า `Cost` ของส่วนประกอบเมื่อแยกชุด (`ItemDetail.duCost`, 6.4) และผลบัญชีจริงของส่วนต่างต้นทุนเมื่อแยกชุด.
7. **การขายสินค้าชุดตัดสต็อกอย่างไร** (ตัดตัวชุดหรือส่วนประกอบ) — อยู่ที่เอกสารขาย (นอกไฟล์นี้; ดูเอกสารส่วนขาย).
8. **ตำแหน่งเมนูหน่วยนับ/กลุ่ม/ยี่ห้อ/คลัง/ที่เก็บ ใน Champ** (ไม่อยู่ใต้ IC) และโครงของ Report Center — ยังไม่ได้ตรวจเมนูโมดูลอื่น.
9. **จอเลือก Serial/Lot ตอนทำเอกสาร** (`OnSaveSerial`/`OnSaveLotItem`) และตัวเลือกบริษัท `nLotStatus 0/1/2` ผลต่อหน้าจอ (4.3, 5.3, 6.5) — เห็นเพียงจุดเรียก ไม่ได้อ่านจอย่อยทุกตัว.
10. **ไม่ได้รันโปรแกรม Champ จริง** — ทุกข้อมูลมาจากอ่านซอร์ส/`.rc`/XML/`MainFrm.cpp`; พฤติกรรมตอนรันจริง (เช่น แถบเครื่องมือ, ปุ่มลัด F-key ของบางเอกสาร) อาจต่างจากที่อ่าน. หลักฐานตรวจซ้ำได้จาก `file:line` ที่อ้างในแต่ละหัวข้อ; ตัวช่วยสคริปต์อยู่นอก repo (`%TEMP%\ic\`) ไม่จำเป็นต่อการใช้เอกสาร.
11. **ไม่ได้ verify ฝั่ง BC ด้วยการรันระบบ** — สถานะ BC (รอพัฒนา/API พร้อม) ได้จากการอ่าน `menu-data.ts`, `menu-screen-status.ts`, `erp-reports.ts`, `erp-tools.ts`, `erp-transaction.ts`, `erp-operations.ts` และแพ็กเกจ `backend/internal/goapi/inventory` ณ 2026-10-07 (ไม่ได้เปิดจอผ่านปุ่ม Demo); งานของเอเจนต์อื่นที่กำลังแก้ไฟล์เหล่านี้อาจทำให้เลขบรรทัดเลื่อน.
