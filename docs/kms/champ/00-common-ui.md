# 00 — กรอบหน้าจอกลางของ Champ (Common UI Framework) และสิ่งที่ BC ต้องปรับ

> สำรวจ 2026-10-07 จากซอร์ส `D:\project-champ` (C++/MFC) เทียบกับ BC (`D:\bccode\frontend\src`) — เอกสารนี้อธิบายสิ่งที่ **ทุกหน้าจอ Champ ใช้ร่วมกัน** (ปุ่ม คีย์ ช่องกรอก ตาราง ค้นหา เลขที่เอกสาร สถานะ สิทธิ์ พิมพ์ วันที่ ตัวเลข ข้อความ) ส่วนฟิลด์/คอลัมน์ของแต่ละเอกสารอยู่ในเอกสารรายโมดูล
> คำสั่งเจ้าของ 2026-10-07: **หน้าเข้าระบบ / เลือก holding / บริษัท / สาขา ของ BC คงไว้ตามเดิม ไม่ปรับให้เหมือน Champ** — Champ parity ใช้กับหน้าจอ *ภายใน* แอปหลังเข้าระบบเท่านั้น; ข้อมูลเก่าทิ้งได้ (เดินหน้าอย่างเดียว) จึงเพิ่ม/แก้ตาราง–ฟิลด์ได้โดยไม่ต้อง migration

ตัวย่อพาธ (ใช้ `ไฟล์:บรรทัด`; เลขบรรทัดเป็นของไฟล์ต้นฉบับ — ถอดรหัสแล้วบรรทัดตรงกัน)

| ตัวย่อ | พาธจริง |
|---|---|
| `FW` | `D:\project-champ\champ-lib\BC5FRMWRK\` (คลาสฐานของฟอร์ม/ไดอะล็อก, MainFrame ฐาน) |
| `OC` | `D:\project-champ\champ-lib\BOLEDBCOMM\` (ช่องกรอก ตาราง popup ค้นหา เลขที่เอกสาร) |
| `QL` | `D:\project-champ\champ-lib\QOledblib\` (หน้าต่างรายการ/ค้นหา `CQList`) |
| `RP` | `D:\project-champ\champ-lib\BOLEDBREPORT\` (หน้าต่างรายงาน) |
| `AP` | `D:\project-champ\champ\champ\` (แอป BC5Account: MainFrm, ฟอร์มเอกสารทุกใบ) |
| `FE` | `D:\bccode\frontend\src\` (BC) |

---

## 0. สรุป 10 เรื่องที่ BC ต้องปรับก่อน (เรียงตามผลต่อผู้ใช้)

รายละเอียด/หลักฐานอยู่หัวข้อ 15

| # | เรื่อง | Champ ทำอย่างไร | BC ตอนนี้ |
|---|---|---|---|
| 1 | **Enter = ไปช่องถัดไป ทั้งฟอร์ม** | `FW/BaseFormView.cpp:207-277` Enter/↓ = ช่องถัดไป ↑ = ช่องก่อนหน้า | ช่องหัวเอกสารไม่มีตัวจัดการ Enter → ฟอร์ม `onSubmit` ทำให้ **Enter = บันทึกฉบับร่าง** (`FE/app/gl/gl-journals.tsx:797`) |
| 2 | **ช่องรหัส: พิมพ์รหัสแล้ว Enter, ไม่พบ = เปิดรายการให้เลือกเอง, F4 = เปิดรายการ** | `OC/BMergeEdit.cpp:426, 678-760` | ช่องบัญชีเป็น `<select>` ที่เปิด dialog ตอนคลิก/Space/Enter/F2 พิมพ์รหัสตรง ๆ ไม่ได้ (`FE/app/gl/gl-common.tsx:468-493`) |
| 3 | **แถบปุ่มเอกสารมาตรฐานเดียวทุกจอ** ใหม่·บันทึก·ลบ·ค้นหา·ก่อนหน้า·ถัดไป·ปิด·ข้อมูล·คีย์ลัด | `FW/BC5FrmWrk.rc:999-1015` | ปุ่มต่างกันตามโหมดของแต่ละจอ ไม่มี ก่อนหน้า/ถัดไป/คีย์ลัด (`gl-journals.tsx:395-403, 636-650, 1052-1067`) |
| 4 | **คีย์ลัดเอกสารครบชุด** Ctrl+N / Ctrl+F / F11 / Ctrl+P / Ctrl+D / Ctrl+Q,W / Shift+F5 / Ctrl+F5 / F12 | `FW/BC5FrmWrk.rc:1085-1136`, `AP/BC5Account.rc:977-1010` | มีแค่ Ctrl+S, Alt+N, Esc (`FE/hooks/use-form-shortcuts.ts:45-75`) |
| 5 | **บันทึกแล้วหน้าจอเคลียร์ พร้อมคีย์ใบใหม่ต่อทันที** (มีสวิตช์ keep/clear) | `OC/BaseDBFormView.cpp:884-981`, `FW/mybasedbformview.cpp:4113-4124` | บันทึกแล้วค้างที่ใบเดิมโหมดดู ต้องกด "บันทึกรายวันใหม่" อีกที (`gl-journals.tsx:289-294`) |
| 6 | **เลขที่เอกสารรันอัตโนมัติ** ตามรูปแบบ `DocFormat` (ปี พ.ศ./เดือน/กลุ่ม/เลขรัน) | `OC/BRunningBttn.cpp:121-433` | ผู้ใช้พิมพ์เลขที่เอง ค่าเริ่มต้นว่าง (`FE/lib/general-ledger.ts:280`, `gl-journals.tsx:825`) |
| 7 | **ถามก่อนทิ้งข้อมูลที่ยังไม่บันทึก 3 ทาง** ใช่(บันทึกแล้วไปต่อ)/ไม่/ยกเลิก | `OC/BaseDBFormView.cpp:687, 728, 814` | 2 ทาง ละทิ้ง/อยู่ต่อ (`gl-journals.tsx:247, 259`) |
| 8 | **พิมพ์: เลือกแบบฟอร์ม · ดูตัวอย่าง · จำกัดจำนวนครั้ง · บันทึกประวัติการพิมพ์ · พิมพ์ได้เฉพาะใบที่บันทึกแล้ว** | `FW/SelectBillFormDlg.cpp`, `FW/mybasedbformview.cpp:2538-2700`, ตาราง `BCPRINTLOG` | `window.print()` ตรง ๆ ไม่มีจำกัด/ประวัติ (`FE/components/thai-document-print-modal.tsx:82-90`, `FE/app/gl/gl-print.tsx:58-69`) |
| 9 | **วันที่พิมพ์เอง `วว/เดือน/พ.ศ.`** + วันนี้อัตโนมัติ + ตัวเลขไม่ผิดรูปถูกล้าง | `OC/BMergeEdit.cpp:858-870, 1250-1275` | `<input type="date">` ของเบราว์เซอร์ (`FE/components/ui/date-time-field.tsx:35-39`, `gl-journals.tsx:830`) |
| 10 | **สถานะเอกสารแบบ Champ** CC / 0.00 / CF / N + สีแถวในรายการ + ยืนยัน/ยกเลิกต้องเลือก YES | `AP/myespeciallyformview.cpp:5590-5720`, `AP/MyDlgBillCancel.cpp:51-85`, `OC/MyExBillQList.cpp:118-160` | ERP workbench มีแค่ ฉบับร่าง/อนุมัติ (`FE/app/crud/erp-crud-workbench.tsx:373-374, 608-613`) |

---

## 1. ลำดับชั้นคลาสของหน้าจอ Champ

| ชั้น | คลาส | หน้าที่ | หลักฐาน |
|---|---|---|---|
| 1 | `CFormView` (MFC) → `CBaseFormView` | Enter/↑/↓ เป็น Tab, tooltip | `FW/BaseFormView.cpp:204-277` |
| 2 | `CBaseDBFormView` | วงจรบันทึก/ลบ/ใหม่/ค้นหา/ก่อนหน้า/ถัดไป, โหลดรูปแบบรายวัน `BCSysTRFormat`, สิทธิ์ใหม่/แก้/ลบ, log | `OC/BaseDBFormView.cpp:177, 678, 717, 803, 884, 994-1426, 1642` |
| 3 | `CMyBaseDBFormView` | แถบปุ่ม, F1 วิธีใช้, Shift+F11, ชื่อหน้าต่าง, พิมพ์, สลับ keep/clear | `FW/mybasedbformview.cpp:128-182, 255-262, 2219-2275, 2538, 4097-4124` |
| 4 | `CMyEspeciallyFormView` | ภาษี/หัก ณ ที่จ่าย/มัดจำ/สต็อก/GL, ยืนยัน/ยกเลิก, ป้ายสถานะ, Posting_GL | `AP/myespeciallyformview.cpp:5590, 5646, 5685, 6467, 6939` |
| 5 | ฟอร์มเอกสารจริง (`CChmFrm*`, `CBILLFrm*`) | ผูกคีย์ลัด `ID_HK_*` ของตัวเอง, ช่องกรอก, ตาราง | เช่น `AP/ChmFrmGoodsPurchase.cpp:90-115`, `AP/billfrmsubmitprice.cpp` |
| — | ไดอะล็อก: `CBaseDialog` → `CMyBaseDialog` | Enter/↑/↓ เป็น Tab เหมือนฟอร์ม; ข้อความบนหัว/ปุ่มแปลผ่านคีย์ `KEY$ไทย` | `OC/BaseDialog.cpp:79-152`, `FW/mybasedialog.cpp` |

กติกาข้อความบนจอ: ป้ายใน `.rc` เขียนเป็น `คีย์$ไทย` (เช่น `"OK$ตกลง"`, `"CANCEL$ยกเลิก"`) แล้วโปรแกรมหาข้อความจริงจากตาราง `myresource` (คอลัมน์ thaires/engres/userres1-9) ผ่าน `GetRes`/`GetMyRes` — ไม่พบคีย์จะแสดง `[KEY]` (`FW/BCSWinApp.cpp:614-675`) ข้อความ `AfxMessageBox` จำนวนมากในโค้ดเป็นอังกฤษตรง ๆ (เช่น `"Complete"`, `"Your permission can't delete record"`) — **ข้อความไทยของคีย์ส่วนใหญ่ไม่อยู่ในซอร์ส** (อยู่ใน DB `myresource` ที่ไม่มีในรีโป) จึงอ้างได้เฉพาะคีย์ + ข้อความไทยที่ปรากฏในคอมเมนต์/ไดอะล็อก

### ตารางตั้งค่ารูปแบบเอกสาร (ขับทุกจอเอกสาร) — จอ "รูปแบบรายวัน" (`ID_SYS_TRANSCONFIG`)

| ตาราง | คอลัมน์ที่ใช้ | ใช้ทำอะไร |
|---|---|---|
| `BCSysTRFormat` | `TransKey, MainTable, JoinTable, FilterBy, SortBy, TransTitle, FindRect, DocFormat, TaxFormat, SlipFomName, WTaxFomName, SearchStyle, DisplayStyle, IsSearchFirst, PrintFrmInText, PrintLimit` | ตารางหลัก/เรียง/กรอง, รูปแบบเลขที่เอกสาร (`DocFormat`), จำกัดจำนวนครั้งพิมพ์ (`PrintLimit`), โหมดค้นหา |
| `BCSysTRFind` | (คอลัมน์หน้าต่างค้นหา) | คอลัมน์ในหน้าต่างค้นหาเอกสาร |
| `BCSysTRFom` | `TransKey, FormFileName` | รายชื่อแบบฟอร์มพิมพ์ของเอกสารนั้น |
| `BCSysTRDocG` | (กลุ่มเอกสาร/ภาษี) | กลุ่มเลขที่เอกสาร (`@`) |

หลักฐานตารางจาก `D:\bccode\scratch\champ-atlas\atlas.json` (DDL `SQLSERVER_Script.sql`); โหลดค่าที่ `OC/BaseDBFormView.cpp:994-1426`

---

## 2. หน้าต่างหลัก (MDI) — `CBCSMainFrame` / `CMainFrame`

| เรื่อง | พฤติกรรม Champ | หลักฐาน |
|---|---|---|
| หนึ่งจอ = หนึ่งหน้าต่าง | เปิดจอที่เปิดอยู่แล้ว → สลับไปหน้าต่างเดิม (วนหา View ชนิดเดียวกันทุก doc template) ไม่เปิดซ้ำ | `FW/BCSWinApp.cpp:2542-2593` (`CreateOrActivateFrame`) |
| ชื่อหน้าต่างของฟอร์ม | `ชื่อจอ : ชื่อบริษัท : รหัสผู้ใช้ : ชื่อฐานข้อมูล : เซิร์ฟเวอร์` | `FW/mybasedbformview.cpp:255-262` |
| แถบทางลัดซ้าย | โฟลเดอร์ "ทางลัด BC": ข้อมูลบริษัท · ข้อมูลหลัก · รูปแบบรายวัน · สินค้า · ลูกค้า · เจ้าหนี้ · ผังบัญชี; โฟลเดอร์ "ทางลัดของฉัน": "สร้างทางลัด.." | `AP/MainFrm.cpp:1132-1186` |
| หน้าเริ่มต้น | หน้า HTML เมนูโมดูล (คลิกลิงก์ → `OnHtmlMenuClick` เปิดจอ) | `AP/MainFrm.cpp:1327` |
| แถบเครื่องมือหลัก (rebar 4 แถบ) | ค้นหา (Edit + ปุ่ม "Go" + กล่อง query ตัวอย่าง `@DocNo='001'`, `@ALL`, `%`) · ระเบียน (`IDR_RECORD_TOOLBAR`) · วิเคราะห์ (`IDR_ANALYSE_TOOLBAR`) · รายงาน (`IDR_REP_TOOLBAR3`) | `FW/BCSWinApp.cpp:1921-2072`, `FW/BC5FrmWrk.rc:1017-1065` — ยังไม่ verify ว่าแถบไหนแสดงจริงตอนรัน (มี `SetBandInfo(...,FALSE)` ซ่อนแถบที่ 4) |
| แถบสถานะ | ช่องแรกข้อความ ("opening..." ระหว่างโหลดรายการ, "Ready" เมื่อเสร็จ) + CAPS/NUM/SCRL | `FW/BCSWinApp.cpp:382-451`, `OC/BaseDBFormView.cpp:205-211, 960-970` |
| ปิดทั้งหมด | ปุ่ม `ID_WINDOW_CLOSE_ALL` "Close All Windows"; ปิดหน้าต่าง `Ctrl+F4`; ปิดไม่ได้ขณะมีงานรัน ("Job is running please wait...") | `FW/BC5FrmWrk.rc:1026-1027, 1104, 1131`, `FW/mybasedbformview.cpp:2299-2326` |
| สิทธิ์เปิดจอ | เปิดเมนูได้ต้องผ่าน `IsMenuEnable` (ดูหัวข้อ 11) | `FW/BCSWinApp.cpp:885-922` |

---

## 3. แถบปุ่มของเอกสาร (toolbar) — ชุดจริงและคีย์ลัด

ชุดปุ่มในฟอร์ม (`IDR_STDFORM_TOOLBAR`) เรียงตามลำดับจริง — ข้อความบนปุ่มตั้งในโค้ดเป็น **อังกฤษ** (`FW/mybasedbformview.cpp:2244-2257`)

| ลำดับ | คำสั่ง (ID) | ข้อความปุ่ม | tooltip + คีย์ลัด (จาก `.rc`) | การทำงาน |
|---|---|---|---|---|
| 1 | `ID_REC_NEW` | New | "New record (Ctrl+N)" | ล้างจอเป็นเอกสารใหม่ (ถ้ามีการแก้ ถามบันทึกก่อน) `OC/BaseDBFormView.cpp:678` |
| 2 | `ID_REC_SAVE` | Save | "Save this Record (F11)" | `OnRecSave` `OC/BaseDBFormView.cpp:884` |
| — | (คั่น) | | | |
| 3 | `ID_REC_DELETE` | Del | "Delete this record (Ctrl+D)" | ลบ (เช็คสิทธิ์ + `ConfirmDeleteData`) `OC/BaseDBFormView.cpp:177` |
| — | (คั่น) | | | |
| 4 | `ID_REC_FIND` | Find | "Find Data (Ctrl+F)" | เปิดหน้าต่างรายการ/ค้นหา `OC/BaseDBFormView.cpp:404` |
| 5 | `ID_REC_PREV` | Prev | "Prev Record (Ctrl+Q)" | เอกสารก่อนหน้าตามลำดับ `SortBy` `OC/BaseDBFormView.cpp:803` |
| 6 | `ID_REC_NEXT` | Next | "Next Record (Ctrl+W)" | เอกสารถัดไป `OC/BaseDBFormView.cpp:717` |
| — | (คั่น) | | | |
| 7 | `ID_CLOSE_FRM` | Close | "Close" | ปิดหน้าต่าง (ปุ่มลัด `Ctrl+F4` ที่ `ID_DB_FRM_CLOSE`) |
| — | (คั่น) | | | |
| 8 | `ID_REC_INFO` | Info | "Information" | เมนูเด้ง: "แสดงเส้นทางเอกสาร" / "แสดงยอดตัวเลขต่างๆ" `AP/ChmFrmGoodsPurchase.cpp:402-425` |
| 9 | `IDC_MORE` | Key | "More" | เมนูเด้งรายการ **คีย์ลัดทั้งหมดของจอนั้น** พร้อมคำอธิบายไทย (หัวข้อ 4) `AP/ChmFrmGoodsPurchase.cpp:506-531` |
| 10 | `IDC_RECUR` | Recur | "Recur" | ปุ่ม drop-down คัดลอก/เรียกซ้ำจากเอกสารเดิม `FW/mybasedbformview.cpp:2412 (BaseRecur)`, `OC/BaseDBFormView.cpp OnBaseRecur` |

หมายเหตุ: **ไม่มีปุ่ม "แก้ไข"/"ยกเลิกการแก้"/"หน้าแรก"/"หน้าสุดท้าย"** ในแถบมาตรฐาน — แก้ไขคือโหลดเอกสารด้วย ค้นหา/ก่อนหน้า/ถัดไป แล้วแก้และกดบันทึก (ระบบรู้ว่าเป็นโหมดแก้จาก `m_bInsertDataMode=FALSE`)

แถบอื่นของ Champ

| แถบ | ปุ่ม (เรียงตามจริง) | หลักฐาน |
|---|---|---|
| ระเบียน (บนหน้าต่างหลัก) | ใหม่ · บันทึก · ลบ · ค้นหา · ก่อนหน้า · ถัดไป · ปิดจอ · ปิดทุกหน้าต่าง · "keep/clear screen after save" · "Edit and Save Some Data (Shift+F11)" | `FW/BC5FrmWrk.rc:1017-1030, 1130-1136` |
| วิเคราะห์ | Find Detail (Ctrl+R) · Find Primary Key (Ctrl+K) · Document Movement · Check Data · Analyse Data · Check Account | `FW/BC5FrmWrk.rc:1038-1047, 1120-1125` |
| รายงาน | พิมพ์ (Ctrl+P) · หน้าก่อน · หน้าถัดไป · ไปหน้า · ค้นหาข้อมูล · ส่งออก · รีเฟรช · คุณสมบัติคอลัมน์ · แสดง/ซ่อนคอลัมน์ · สีคอลัมน์ | `FW/BC5FrmWrk.rc:1049-1065` |

เมนูเด้งฟอร์มเอกสาร Champ พร้อมป้ายไทย (จากคอมเมนต์ในโค้ด; ข้อความจริงมาจาก `myresource`) — `AP/ChmFrmGoodsPurchase.cpp:536-556`, `AP/billfrmsubmitprice.cpp:3236-3246`

| คีย์ | ข้อความ |
|---|---|
| F2 | บัญชีแยกประเภท (ดู/จำลองรายการ GL) |
| F3 | บันทึกภาษีซื้อ (ซื้อ) |
| F5 | รายละเอียดอื่นๆ |
| F6 | ภาษีหัก ณ ที่จ่าย |
| F7 | บันทึกจ่ายเงิน |
| F8 | หักเงินมัดจำรับ |
| F12 | ตรวจสอบเอกสาร |
| Shift+F2 | ขยายรายการสินค้า (แถวที่เลือก) |
| Shift+F7 | เลือก PO รับสินค้า |
| Shift+F8 | หน่วยนับขนาน |
| Shift+F9 | Lot สินค้า |
| Ctrl+F2 | ยอดคงเหลือ, สินค้าทดแทน (ขาย) |
| Ctrl+F3 | เลือกใบสั่งซื้อสินค้า |
| Ctrl+F7 | Default คลัง, ที่เก็บเริ่มต้น |
| Ctrl+F8 | Default คลังสินค้าอัตโนมัติ |
| Ctrl+F9 | Serial Number |
| Ctrl+F12 | รายละเอียดเกี่ยวกับสินค้า |
| Shift+F12 | ตรวจสอบราคาสินค้า (ขาย) |

---

## 4. ปุ่มลัด (hotkey) — ตารางเร่งปุ่มกลางและการผูกในแต่ละจอ

ตารางเร่งปุ่มกลางของแอป `IDR_MAINFRAME_CHM` (`AP/BC5Account.rc:977-1010`) ส่งเป็น `WM_COMMAND ID_HK_*` ไปยังฟอร์มที่เปิดอยู่ — **ฟอร์มใดไม่ผูก `ON_COMMAND(ID_HK_xxx)` คีย์นั้นก็ไม่ทำงานในจอนั้น** (ฟอร์มส่วนใหญ่ผูก NEW/FIND/PRINT: นับ `ON_COMMAND` ใน `champ/**/*.cpp` = NEW 72, FIND 72, PRINT 64 จอ) ทุกตัวมีตัวกันกดซ้อน `m_bRespondHotKey` (เช่น `AP/ChmFrmGoodsPurchase.cpp:3500-3510`)

| คีย์ | ID | ความหมาย (ดูจากตัวจัดการจริง) | ผูกใน |
|---|---|---|---|
| Ctrl+N | `ID_HK_NEW` | เอกสารใหม่ → `OnRecNew` | ทุกฟอร์มเอกสาร |
| Ctrl+F | `ID_HK_FIND` | ค้นหา → `OnRecFind` | ทุกฟอร์มเอกสาร |
| Ctrl+P | `ID_HK_PRINT` | พิมพ์ → `OnCheckBeforePrint` แล้ว `OnTBPrintForm` | ทุกฟอร์มที่พิมพ์ได้ |
| F11 | `ID_HK_SAVE11` | บันทึก → `OnRecSave` (ฟอร์มฐานดัก F11 เองด้วย `OC/BaseDBFormView.cpp:1626-1637`) | ทุกฟอร์ม |
| Shift+F11 | (ใน `PreTranslateMessage`) | "แก้ไขบางฟิลด์แล้วบันทึก" `OnEditDesc` `FW/mybasedbformview.cpp:2262-2276` | ฟอร์มที่เปิดใช้ |
| F1 | (ใน `PreTranslateMessage`) | วิธีใช้ของจอ (`m_strHelpKey`); Ctrl+F1 = ตัวออกแบบวิธีใช้ | ทุกฟอร์ม |
| F2 | `ID_HK_GL` | ดูรายการบัญชี GL ของเอกสาร (`OnAccount`) | ฟอร์มซื้อ/ขาย/รับ-จ่าย |
| F3 | `ID_HK_TAX` | บันทึก/ดูภาษี (`OnInputTax`) แล้วอัปเดตเลขที่/วันที่ใบกำกับบนจอ | ซื้อ/ขาย |
| F5 | `ID_HK_DOCDESC` | สลับแท็บ "รายละเอียดอื่นๆ" ↔ หน้าหลัก (`m_TabCtrl`) | ฟอร์มที่มี 2 แท็บ |
| F6 | `ID_HK_WTAX` | ภาษีหัก ณ ที่จ่าย (คำนวณใหม่แล้วใส่ยอดสุทธิ) | ซื้อ/จ่าย |
| F7 | `ID_HK_MONEY` | บันทึกรับ/จ่ายเงิน (`OnPay`) | ซื้อ/ขาย |
| F8 | `ID_HK_DEPOSIT` | หักเงินมัดจำ (ถ้าตั้ง `IsDepositCash` และบิลสด → เตือน "EARNEST_CASH") | ซื้อ/ขาย |
| F9 / Shift+F9 / Ctrl+F9 | `ID_HK_CANCEL_HOLD` / `ID_HK_ITEMLOT` / `ID_HK_ITEMSERIAL` | ยกเลิกพักบิล / Lot / Serial | หน้าขายสด/คลัง |
| F12 | `ID_HK_DOCFLOW` | ตรวจสอบเอกสาร / เส้นทางเอกสาร (`OnDocumentFlow`) | ทุกฟอร์มเอกสาร |
| Shift+F5 | `ID_HK_CONFIRMDOC` | ยืนยันเอกสาร (ต้องมีสิทธิ์ `bCanConfirmBill`) | ซื้อ/ขาย |
| Ctrl+F5 | `ID_HK_CANCELDOC` | ยกเลิกเอกสาร (ต้องไม่ถูกยืนยัน) | ซื้อ/ขาย |
| Ctrl+F3 / Shift+F3 | `ID_HK_GETSO` / `ID_HK_GETIR` | ดึงเอกสารต้นทาง (ใบสั่งซื้อ/ขาย, ใบขอซื้อ) | ซื้อ/ขาย |
| Ctrl+F7 / Ctrl+F8 | `ID_HK_SETDEFWH` / `ID_HK_SETDEFSH` | ตั้งคลัง/ที่เก็บ default (ใช้กับทุกแถว) | ซื้อ/รับสินค้า |
| Shift+F2 | `ID_HK_ROWEXPAND` | ขยายสินค้าชุด/แถวที่เลือก | ซื้อ/ขาย |
| Shift+F4 | `ID_HK_FIND_F4` | ค้นหาแบบ F4 | บางจอ |
| Ctrl+F12 | `ID_HK_ITEMINFO` | ประวัติ/ราคา/รายละเอียดสินค้า | ซื้อ/ขาย |
| Ctrl+B / Ctrl+G | `ID_HK_BARCODE` / `ID_HK_CTRL_G` | บาร์โค้ด / เฉพาะจอ — ยังไม่ verify ความหมาย Ctrl+G | จอขาย/คลัง |
| Ctrl+Home | `ID_HK_CANCEL_ALERT` | ยกเลิกแจ้งเตือน — ยังไม่ verify | — |
| PageUp / PageDown | `ID_HK_PAGE_UP/DOWN` | โฟกัสไปช่องเลขที่เอกสาร (หน้าแรก) / ไปช่องถัดไปของหน้า | ฟอร์มเอกสาร |

ตัวอย่างการผูกจริง (ฟอร์มซื้อสินค้า): `AP/ChmFrmGoodsPurchase.cpp:90-115`; ตัวจัดการยืนยัน/ยกเลิก `:3545-3575`

---

## 5. การเลื่อนโฟกัสและปุ่มพื้นฐานในฟอร์ม/ไดอะล็อก

| ปุ่ม | พฤติกรรม | หลักฐาน |
|---|---|---|
| Enter | ไปช่องถัดไป (`NextDlgCtrl`) ยกเว้นเมื่ออยู่บนปุ่ม (กดปุ่ม) และช่องหลายบรรทัดที่รับ Enter (Enter = ขึ้นบรรทัดใหม่, **Ctrl+Enter = ไปช่องถัดไป**) | `FW/BaseFormView.cpp:228-262`; ไดอะล็อกเหมือนกัน `OC/BaseDialog.cpp:79-152` |
| ↓ / ↑ | ช่องถัดไป/ก่อนหน้า (ยกเว้นช่องหลายบรรทัดและ combo ที่ใช้ลูกศรเลือกค่า) | `FW/BaseFormView.cpp:222-272` |
| พิมพ์ตัวอักษรใด ๆ | ตั้งธง "มีการแก้ไข" (`m_bChangeData`) เพื่อใช้ถามก่อนทิ้ง | `OC/BaseDBFormView.cpp:1626-1630` |
| F11 | บันทึก | `OC/BaseDBFormView.cpp:1631-1636` |
| Esc | ปิดตัวแก้ไขในช่องตาราง (`DestroyControl`) / ยกเลิกไดอะล็อก — ในฟอร์มไม่มีปุ่ม "ยกเลิกการแก้" | `OC/BInputListWnd.cpp:2243-2246` |
| Tab ที่ช่องสุดท้ายของตาราง | ออกไปช่องถัดไปของฟอร์ม (Shift+Tab ที่ช่องแรก = ย้อนออก) | `OC/BInputListWnd.cpp:2204-2217` |

---

## 6. ตารางรายการ (detail grid) — `CBInputListWnd`

คลาสตารางหลักของทุกเอกสาร (`OC/BInputListWnd.cpp`, 5,900+ บรรทัด); ตัวแก้ไขในเซลล์คือ `CBMergeEdit` (หัวข้อ 7) หนึ่งตัวที่เปลี่ยนชนิดตามคอลัมน์

| ปุ่ม/การกระทำ | ผล | หลักฐาน |
|---|---|---|
| Enter / Tab / → | ไปเซลล์ถัดไป (คอลัมน์ถัดไป ต่อด้วยแถวถัดไป) | `OC/BInputListWnd.cpp:2204-2217, 2265`; `SelectColNext` 2091 |
| ← / Shift+Tab | เซลล์ก่อนหน้า | `:2259, 2204` |
| ↓ / ↑ | แถวถัดไป/ก่อนหน้า; Shift+↓↑ = เลือกหลายแถว | `:2219-2240` |
| Shift+← / Shift+→ | เลือกหลายคอลัมน์ | `:2246-2262` |
| Space ร่วมกับ Shift / Ctrl | เลือกแถวทั้งแถว / คอลัมน์ทั้งคอลัมน์ | `:2283-2290` |
| **Insert** | แทรกแถวว่างที่ตำแหน่งปัจจุบัน | `:593-607` |
| Ctrl+Insert / Shift+Insert | คัดลอก / วาง (จาก/ไป Excel — `OnCopy` 5749, `OnPaste` 5864) | `:593-600` |
| **Delete** | ถ้ามีแถว/คอลัมน์ที่เลือก = ลบแถว (`DeleteSelectRow` 2646); ถ้าไม่ได้เลือก = ลบตัวอักษรในช่อง | `:637, 2268, 2716` |
| **F4** | เปิดรายการค้นหาของคอลัมน์ชนิด LINK (เรียก `ILDN_LINK` ให้ฟอร์มเปิด popup) | `:635, 2272-2278`; `OnBnClick` 3579 |
| Ctrl+A / Ctrl+Space | ปุ่มพิเศษของคอลัมน์ ("other command" → `ILDN_OTHER_BNCKICKED`) | `:565-585`; `OnBnOtherClick` 5718 |
| แถวใหม่ท้ายตาราง | เมื่อเลื่อนเลยแถวสุดท้ายจะเพิ่มแถวใหม่เอง (`AddNew` 1037) | `:1037` |
| คอลัมน์ "เงา" (shadow) | เก็บค่าที่ไม่แสดง (ประเภทสินค้า, รหัสหน่วย ฯลฯ) ใช้ `GetShadowData` | เช่น `AP/billfrmsubmitprice.cpp:560` |
| ยอดรวม | **ไม่อยู่ในตาราง** — เป็นช่องตัวเลขใต้ตารางที่ฟอร์มคำนวณและเขียนลง (`m_edSumDetail`, `m_edNetAmount` ...) ผ่านการแจ้งเหตุการณ์ `ILDN_*` | `AP/ChmFrmGoodsPurchase.cpp:3577-3591` |

ชนิดคอลัมน์ในตาราง = ชนิดช่องกรอกด้านล่าง (ฟิลด์ `FCS_*`: LINK, NUMBER/CALC, DATE, COMBO, CHECK ฯลฯ)

---

## 7. ช่องกรอก (`CBMergeEdit` + `CBMergeBttn`)

ช่องเดียวที่รวมกล่องข้อความ + ปุ่มด้านขวา เปลี่ยนชนิดด้วย `m_nEditStyle` (`OC/BMergeEdit.cpp:27`, `BMergeBttn.cpp`)

| ชนิด | พฤติกรรม | หลักฐาน |
|---|---|---|
| ข้อความปกติ (`BME_NORMAL`) | พิมพ์ได้ทุกตัวอักษร | `BMergeEdit.cpp:27` |
| ตัวเลข (`BME_NUMBER`) / จำนวนเงิน (`BME_CALC`) | จุดทศนิยมตามที่ตั้ง (`m_nPoint`); ออกจากช่อง: ตัดเครื่องหมาย `,` แล้วจัดรูปใหม่; ช่องว่าง = `0.00` (ตามจำนวนทศนิยม); ปุ่มด้านขวา = **เครื่องคิดเลข** | `:604-612, 991-1023, 1290-1300`; `BMergeBttn.cpp OnPopup` |
| วันที่ไทย/อังกฤษ (`BME_DATE_TH/ENG`) | ปุ่มด้านขวา = ปฏิทิน; รูปแบบ `วว/เดือน/ปี` (หัวข้อ 13) | `BMergeEdit.cpp:833-870` (OnSetFocus), `1231-1275` (UpdateData) |
| วันที่อัตโนมัติ (`..._AUTO`) | ว่าง = ใส่วันนี้เมื่อออกจากช่อง | `:591-596` |
| เวลา (`BME_TIME`) | `HH:MM` ยาวไม่ครบ 5 ตัว = ล้างค่า | `:582-585` |
| รหัสอ้างอิง (`BME_LINK` / `BME_LINK_CONFIG`) | พิมพ์รหัสแล้วตรวจกับตารางหลัก; **ตัวพิมพ์ใหญ่อัตโนมัติ**; ปุ่มด้านขวา/F4 = เปิดรายการ (หัวข้อ 8) | `:225, 266, 587-590` |
| Combo ข้อความ/ตาราง (`BME_COMBO_NM*`, `BME_COMBO_TB`) | เลือกจากรายการ; ว่าง = ค่าแรก; พิมพ์ค่าที่ไม่มี → แจ้ง "ไม่พบข้อมูล" (`EXPORT_strMasterNotFound`) แล้วเก็บโฟกัส | `:600-665` |
| เลือกหลายรายการ (`MultiSel`) | ปุ่มเปิดรายการติ๊กหลายตัว | `BMergeBttn.cpp` — ยังไม่ verify รายละเอียด |

---

## 8. ค้นหา / เลือกจากรายการ (lookup)

### 8.1 ช่องรหัส (LINK) — พิมพ์รหัสแล้วตรวจ

| สถานการณ์ | ผล | หลักฐาน |
|---|---|---|
| พิมพ์รหัสแล้วกด Enter / Tab / ↑ / ↓ | ตรวจ `SELECT <ฟิลด์> FROM <ตาราง> WHERE <ฟิลด์>='<รหัส>'` ถ้าเจอ ไปช่องถัดไป | `OC/BMergeEdit.cpp:678-722` |
| ไม่พบ | แจ้ง "Data Not Found ...!" → **ล้างช่อง → โฟกัสกลับ → เปิดรายการค้นหาให้เลือกทันที** | `:724-730` |
| พิมพ์ `?` หรือ `#` นำหน้า แล้ว Enter | เปิดรายการค้นหาทันที (ใช้ข้อความที่เหลือเป็นคำค้น) | `:690-694` |
| กด **F4** ในช่อง | เปิดรายการค้นหา | `:426` |
| ลูกค้า: พิมพ์ `*บาร์โค้ด` | ค้นลูกค้าจากบาร์โค้ด (`BCArBarCodeMaster`) | `:700-716` |
| ออกจากช่อง | แปลงเป็นตัวพิมพ์ใหญ่ | `:587-590` |

### 8.2 รายการเลือก (popup) — `CBMasterSel`

หน้าต่างเล็กไม่มีขอบหัว วางใต้ช่อง: ตารางรายการ + ช่อง "Search:" + ปุ่ม "Ok"/"Cancel" (`OC/BOLEDBComm.rc:139-151`)
- พิมพ์ในช่อง Search = กรองสด `รหัส LIKE '%คำ%' OR ชื่อ LIKE '%คำ%'` (**ค้นส่วนใดของรหัสหรือชื่อก็ได้**) `OC/BMasterSel.cpp:549-590`
- คลิกหัวคอลัมน์ = เรียงใหม่ตามคอลัมน์นั้น `:458-466`; Enter = เลือกแถวที่ไฮไลต์ `:436`

### 8.3 หน้าต่างรายการ/ค้นหาเอกสาร — `CQList` / `CExQList` / `CMyExBillQList`

เปิดจากปุ่ม Find (Ctrl+F) และจากช่องรหัสบางชนิด; caption "Browse for Select" (`QL/QOLEDBLib.rc:256-275`)

| ส่วน | รายละเอียด |
|---|---|
| ช่องค้นหา "What?=F11" | พิมพ์คำค้น (เริ่มค้นทันที หรือกด Enter ถ้าตั้ง `search_after_enter=yes` `QL/QList.cpp:1421, 1948`) |
| "Display=Ctrl+F11" | `1=Display all rows` / `2=Display top of rows` (`QOLEDBLib.rc:603-615`; สลับ `QList.cpp:1674`) |
| "Mode=Ctrl+F12" | `1=Begin of value` (ขึ้นต้นด้วย) / `2=Part of value` (มีอยู่ส่วนใดก็ได้) (`QList.cpp:2275-2290`) |
| ปุ่ม | "OK=Enter", "Cancel=Esc", "Help=F1" (ซ่อน), ข้อความนับ "Records" |
| ปุ่มลัด | F11 โฟกัสช่องค้น · Ctrl+F11 ดู SQL · F12 ตัวเลือก · F2 ล้างตัวกรองทั้งหมด · F3 ตัวกรอง · F4 ค้นเต็ม · F5 รีเฟรช · **F6 หน้าแรก · F7 ก่อนหน้า · F8 ถัดไป · F9 หน้าสุดท้าย** · Enter เลือก (`QOLEDBLib.rc:627-642`) |
| สีแถว (ฉบับค้นหาเอกสาร) | ยกเลิก = ตัวแดง; ยอดคงค้าง 0.00 หรือใช้หมดแล้ว = ตัวเทา; ยืนยันแล้ว = ตัวเขียวน้ำเงิน (0,128,128) (`OC/MyExBillQList.cpp:118-160, 262-305`) |
| คอลัมน์ | กำหนดต่อเอกสารใน `BCSysTRFind` (ตั้งจาก "รูปแบบรายวัน") |

---

## 9. เลขที่เอกสาร (running number) — `CBRunningEdit` / `CBRunningBttn`

รูปแบบเก็บใน `BCSysTRFormat.DocFormat` ต่อเอกสาร (`OC/BRunningBttn.cpp:127-175, 290-433`)

| สัญลักษณ์ | ความหมาย |
|---|---|
| `YY` | ปี 2 หลักท้าย — **ปีตามชนิดปีของบริษัท**: ถ้าเป็น พ.ศ. บวก 543 ก่อนตัดเหลือ 2 หลัก (`:121-123, 352-360`) |
| `MM` / `DD` | เดือน / วัน 2 หลัก (`:357-377`) |
| `@` (หนึ่งตัว ต่อหนึ่งหลักของกลุ่ม) | **กลุ่มผู้ใช้เลือกเอง** (เช่น ชุดเอกสาร); ผู้ใช้ต้องเลือกให้ครบจำนวน `@` ไม่งั้นเตือน `EXPORT_strDocRunningMessage_4` (`:175-340`) |
| `#...#` ท้ายสุด | เลขรัน เติม 0 ซ้ายตามจำนวน `#` (`%0Ni` `:428-433`); ต้องอยู่ท้ายรูปแบบ |
| ตัวอักษรอื่น | คงตามตัวอักษร เช่น `IV-`, `-` |

อัลกอริทึม: แทนค่า `YY/MM/DD/@` แล้วหา `SELECT MAX(<ฟิลด์เลขที่>) FROM <ตาราง> WHERE <ฟิลด์> LIKE '<ส่วนหน้า>%'` นำ 'ส่วนที่เป็นเลขรัน' + 1 (`:383-433`) — ถ้า `DocFormat` ว่าง ใช้ `MAX` ทั้งฟิลด์ (`:139-165`) เลือกวันที่ฐานจาก `m_strRunningDate` (วันที่เอกสาร) ถ้ามี (`:361-374`)
- F4 ในช่องเลขที่ = เรียกรันใหม่/เลือกกลุ่ม (`OC/BRunningEdit.cpp:140`)
- **ล็อกเลขที่**: ผู้ใช้ที่ไม่มีสิทธิ์ `bCanChangeBillNo` แก้เลขที่ไม่ได้ (ช่องถูก disable; ระบบรันให้ใหม่ทุกครั้งที่เริ่มเอกสาร) `AP/ChmFrmGoodsPurchase.cpp:369-380`, `AP/billfrmsubmitprice.cpp:342-345`
- **ล็อกวันที่**: ผู้ใช้ที่ไม่มีสิทธิ์ `IsCanEditDocDate` (ยกเว้น `sa`) แก้วันที่เอกสารไม่ได้ `AP/ChmFrmGoodsPurchase.cpp:209`, `AP/billfrmsubmitprice.cpp:398-399`
- ความเสี่ยง: ใช้ `SELECT MAX` ไม่มีการล็อก (ผู้ใช้สองคนคีย์พร้อมกันอาจได้เลขซ้ำ) — BC ห้ามลอกส่วนนี้ ใน PostgreSQL ต้องจองเลขใน transaction (ดูหัวข้อ 15 ข้อ 6)

---

## 10. วงจรบันทึก · สถานะ · ล็อก · ยืนยัน/ยกเลิก

### 10.1 ลำดับของปุ่ม "บันทึก" (F11) — `OC/BaseDBFormView.cpp:884-981`

1. ตรวจสิทธิ์: แก้ไม่ได้ → "Your permission can't change record"; เพิ่มไม่ได้ → "Your permission can't insert new record" (`:891-900`)
2. `ConfirmSaveData()` (แต่ละฟอร์ม override ตรวจความครบถ้วน — ตัวอย่างหัวข้อ 10.4)
3. บันทึกประวัติการแก้ (change log) ถ้าเปิด `logcfg` → `OnSaveData()` → ถ้า error code 100/101 หยุด
4. `Posting_GL()` ลงบัญชีทันที (`AP/myespeciallyformview.cpp:6939`) → `OnPrintForm()` (พิมพ์อัตโนมัติถ้าบริษัทตั้ง `PrintBill=1`) → ล้างฟิลด์ชั่วคราว
5. `m_bClearScreen` = จริง (ค่าเริ่มต้น): **ล้างจอ** ; เท็จ: ล้างเฉพาะ key หลัก (`OnClearPrimary`) — สลับด้วยปุ่ม "keep/clear screen after save" ที่ถามด้วย "Do you want to keep screen after save" / "clear screen after save" (`FW/mybasedbformview.cpp:4113-4124`)
6. โหลดรายการเอกสารใหม่ (แถบสถานะ "opening..." → "Ready") แล้ว `OnBeginNewRecord()` — **กลับสู่โหมดเอกสารใหม่ พร้อมคีย์ใบถัดไป**

ไม่มีสถานะ "ฉบับร่าง" ในเอกสาร ERP ของ Champ — กดบันทึกคือบันทึกจริงและลงบัญชีตามที่ตั้ง

### 10.2 ป้ายสถานะมุมซ้ายบนของฟอร์ม — `AP/myespeciallyformview.cpp:5685-5720`

| ป้าย | เงื่อนไข (เรียงลำดับตรวจ) | สี (พื้น/ตัวอักษร) |
|---|---|---|
| **CC** | `IsCancel=1` (ยกเลิกแล้ว) | แดง / ขาว |
| **0.00** | ยอดคงค้างของบิลเป็นศูนย์ (ชำระ/ใช้ครบ) | น้ำตาล (128,64,0) / ขาว |
| **CF** | `IsConfirm=1` (ยืนยันแล้ว) | เขียวน้ำเงิน (0,128,128) / ขาว |
| **N** | ปกติ | ขาว / ดำ |

### 10.3 ยืนยัน (Confirm) และ ยกเลิก (Cancel) — ไดอะล็อก

| เรื่อง | ยืนยันเอกสาร | ยกเลิกเอกสาร |
|---|---|---|
| คีย์ลัด | Shift+F5 | Ctrl+F5 |
| ต้องมีสิทธิ์ | `bCanConfirmBill` (ไม่มี → ข้อความคีย์ `NOT_CONFIRM_NO`) `AP/myespeciallyformview.cpp:5646` | `bCanCancelBill` (`NOT_CANCEL_NO`) `:5590` |
| ห้ามทำซ้ำ | ถ้ายืนยันแล้ว → `MSG_CANNOT_CONFIRM` | ถ้ายกเลิกแล้ว → `MSG_CANNOT_CANCEL`; ถ้ายืนยันแล้ว → `CANNOT_CANCEL` (`AP/billfrmsubmitprice.cpp:4123-4134`); ถ้าลงบัญชี GL แล้ว (`IsPostGL`) → ออกทันที |
| ไดอะล็อก | caption "Confirm Document" มี combobox **Yes / No** (เลือก Yes แล้วกด "ตกลง") `AP/BC5Account.rc:3490-3509` | caption "Cancel Document" มี combobox **Yes / No** `AP/BC5Account.rc:3511-3530` |
| คำเตือนบนจอ (ไทย) | "*** กรุณาอ่านคำเตือนนี้ก่อน***" + "ดั้งนั้นก่อนยืนยันเอกสาร(Confirm) ท่านต้องตรวจสอบแน่ใจก่อนว่าจะไม่มีการแก้ไขเอกสารนี้อีก เพื่อให้แน่ใจกรุณาเลือกคำว่า YES ในช่องยืนยัน…" ป้าย "นี่คือเอกสารที่ท่านต้องการยืนยัน:" / "ท่านต้องการยืนยันเอกสารนี้จริงหรือไม่:" | "*** กรุณาอ่านคำเตือนนี้ก่อน***\nเมื่อท่านยกเลิกเอกสารนี้แล้วท่านจะไม่สามารถนำเอกสารนี้ไปใช้ในระบบได้อีก" + "และไม่สามารถเรียกกลับคืนได้ ดังนั้นกรุณาตรวจสอบให้แน่ใจอีกครั้งก่อนยกเลิก เพื่อให้แน่ใจกรุณาเลือกคำว่า YES ในช่องยืนยืนก่อนกดที่ปุ๋มตกลง" ป้าย "นี่คือเอกสารที่ท่านต้องการยกเลิก:" / "ท่านต้องการยกเลิกเอกสารนี้จริงหรือไม่:" |
| ปุ่ม | "ตกลง" (`OK$ตกลง`), "ยกเลิก" (`CANCEL$ยกเลิก`) | เหมือนกัน |
| บันทึกอะไรลงเอกสาร | `ConfirmCode`=ผู้ใช้, `ConfirmDateTime`, `IsConfirm=1` (`AP/MyDlgBillConfirm.cpp:52-74`) | `CancelCode`=ผู้ใช้, `CancelDateTime`, `IsCancel=1` — **ไม่มีช่องเหตุผล** (`AP/MyDlgBillCancel.cpp:51-85`); ใบรับเงิน (`BCReceipt1`) ที่ยกเลิก → เช็ครับ `BCChqIn.Status=4`; ใบจ่าย (`BCPayment`) → เช็คจ่าย `BCChqOut.Status=2` |
| หลังสำเร็จ | แจ้ง "Complete" แล้วล้างจอ เข้าโหมดใหม่ | แจ้ง "Complete" แล้ว `BillCancelUpdateData` (ตั้ง `IsCancel=1` ที่ตารางหลักและตารางย่อย) + บันทึก logbook + ล้างจอ (`AP/billfrmsubmitprice.cpp:482-497`) |

### 10.4 กติกาล็อก / แก้ไข / ลบ (ฟอร์มเสนอราคา `AP/billfrmsubmitprice.cpp` เป็นตัวอย่างอ้างอิง)

| กฎ | ผล | หลักฐาน |
|---|---|---|
| ต้องมีรหัสลูกค้า/เจ้าหนี้ | ไม่มี → ข้อความ `CANNOT_SAVE_NOTARCODE` | `:525-529` (ConfirmSaveData เริ่ม `:518`) |
| แก้เอกสารที่ยกเลิกแล้ว | ห้ามบันทึก (`NOT_SAVE_ISCANCEL`) | `:533-536` |
| แก้เอกสารที่ **ยืนยันแล้ว** | ห้ามบันทึก (`NOT_SAVE_ISCONFIRM`) | `:538-541` |
| แก้เอกสารที่ถูกอ้างอิงต่อแล้ว | ห้ามบันทึก (`CANNOT_SAVE_BILLSTATUS_REF`) | `:543-546` |
| ไม่มีรายการสินค้าในตาราง | ห้ามบันทึก (`CANNOT_SAVE_NOTITEM`) | `:549-552` |
| จำนวนสินค้า = 0 | ตรวจทุกแถวก่อนบันทึก | `:554-560` |
| ลบ | ต้องมีเลขที่; ห้ามลบถ้ายกเลิกแล้ว (`NOT_DELETE_ISCANCEL`) หรือยืนยันแล้ว (`NOT_DELETE_ISCONFIRM`); แล้วถาม Yes/No `CONFIRM_DELETEDATA`; **ลบจริง** (`DELETE FROM` หัวและตารางย่อย) + logbook | `:457-476` (ConfirmDeleteData), `:499-516` (OnDeleteData) |
| ลงบัญชี GL แล้ว (`IsPostGL=1`) | ห้ามลบ/ยกเลิก/แก้ — แสดง "POSTGL_WARN" พร้อมข้อมูลชุดผ่านรายการ: WorkStation · BatchDate · BatchOwner · BatchCheck1 · BatchNo (จาก `BCPostTrans`) | `FW/mybasedbformview.cpp:4126-4162`; `AP/ChmFrmGoodsPurchase.cpp:1592-1596, 3454` |
| งวดบัญชีปิด | วันที่เอกสารอยู่ในงวดที่ `BCPeriod.Status=0` หรือไม่อยู่ในงวดใด → "DATE_CLOSE_PERIOD" ห้ามทำรายการ | `AP/mytools.cpp:834-865` |
| เอกสารถูกใช้ต่อ | ใบซื้อที่มีใบเพิ่มหนี้/ลดหนี้/คืนสินค้า/วางบิล/จ่ายชำระอ้างอยู่ (ที่ไม่ยกเลิก) ห้ามลบ (`CANNOT_DELETE_APBILL`) | `AP/ChmFrmGoodsPurchase.cpp:1597-1625` |
| เตือนก่อนทิ้งข้อมูลที่แก้ | `New / Find / Prev / Next / ปิด` เมื่อมีการแก้ถาม **"Do you want to save any change ?"** ปุ่ม **Yes (บันทึกแล้วไปต่อ) / No (ทิ้ง) / Cancel (อยู่ต่อ)** | `OC/BaseDBFormView.cpp:50, 309, 442, 687, 728, 814` |
| สิทธิ์ใหม่/แก้/ลบ | ใหม่ไม่ได้ → "Your permission can't add new record"; ลบไม่ได้ → "Your permission can't delete record" | `OC/BaseDBFormView.cpp:179-181, 681-683` |
| ประวัติการทำงาน | `SaveLog / SaveCancelLog / SaveDeleteLog` ลง `BCUserLogBook` (WorkMode, WorkType, MenuID, DocNo, UserID, ComputerName…) | `OC/BaseDBFormView.cpp:2447-2490`; ตาราง `BCUSERLOGBOOK` |

---

## 11. สิทธิ์ผู้ใช้

| ระดับ | กลไก | หลักฐาน |
|---|---|---|
| เปิดจอ (เมนู) | `IsMenuEnable(MenuID, "ISMAKE")`: `sa`/`ADMIN` ผ่านทุกจอ; ผู้ใช้อื่นต้องมีแถว `MenuID` ที่ `IsMake=1` ไม่งั้น "คุณไม่มีสิทธิเข้าหน้าจอนี้" (คีย์ `CANNOT_OPEN`) | `FW/BCSWinApp.cpp:885-922` |
| ใหม่ / แก้ / ลบ รายจอ | `SetUserAuthorization`: `IsInsert / IsEdit / IsDelete` ต่อ `MenuID` จากตาราง `BCUserPermission` (ผู้ใช้) หรือ `BCRolePermission` (บทบาท) ตามโหมด (`nMode` 0/2/3; ตั้งค่า `separate_permission=yes` บังคับใช้บทบาท) | `OC/BaseDBFormView.cpp:1642-1760`; ตาราง `BCROLEPERMISSION/BCUSERPERMISSION (menuid, programstring, ismake, isdelete, isinsert, isedit)` |
| ธง 16 ข้อ (ค่า `BCOption1-16` ใน `BCRolePermission2/BCUserPermission2`) | ดูตารางล่าง | `AP/BC5Account.cpp:6150-6340`; ค่าเริ่มต้นทั้งหมด TRUE `FW/BCSWinApp.cpp:1636-1657` |
| ผู้ใช้ | ตาราง `BCUSER (code, name, rolecode)`, `BCROLE (code, name)`, `BCUSERGROUP` | atlas.json |

ธงสิทธิ์ (ป้ายไทยจากคอมเมนต์ในโค้ด) — `AP/BC5Account.cpp:6190-6215`, `FW/BCSWinApp.cpp:1636-1657`

| ตัวเลือก | ธง | ความหมาย |
|---|---|---|
| BCOption1 (อนุมานจากลำดับ — บรรทัดก่อน BCOption2) | `bCanPrintReport` | พิมพ์รายงานได้ |
| BCOption2 | `bCanChangePrice` | แก้ไขราคาขายสินค้า |
| BCOption3 | `bCanChangeItemDiscount` | แก้ไขส่วนลดรายการสินค้า |
| BCOption4 | `bCanChangeBillDiscount` | แก้ไขส่วนลดการค้า (ท้ายบิล) |
| BCOption5 | `bCanSeeItemCost` | ดูต้นทุนสินค้า |
| BCOption6 | `bCanChangeReportOption` | เปลี่ยน Report Option |
| BCOption7 | `bCanPrintBill` | พิมพ์เอกสารต่างๆ |
| BCOption8 | `bCanConfirmBill` | ยืนยันเอกสาร |
| BCOption9 | `bCanChangeBillConfirmed` | แก้ไขเอกสารที่ถูกยืนยันแล้ว (**โค้ดตั้งค่าถูกคอมเมนต์ปิด** — ค่าเป็น TRUE เสมอ) |
| BCOption10 | `bCanAuthorizeCreditLimit` | อนุมัติวงเงิน |
| BCOption11 | `bCanChangeBillNo` | แก้ไขเลขที่เอกสาร |
| BCOption12 | `bCanSaleLost` | ขายขาดทุน |
| BCOption13 | `bCanExportReportData` | ส่งออกข้อมูลรายงาน |
| BCOption14 | `bCanCancelBill` | ยกเลิกบิล |
| BCOption15 | `bCanChangeBillAmount` | แก้ไขตัวเลขท้ายบิล |
| BCOption16 | `bCanChangeGRBuyPrice` | แก้ไขราคาซื้อของใบรับสินค้า |
| อื่น ๆ | `IsCanEditDocDate`, `isPrintOverLimit`, `IsUserEditTax` | แก้วันที่เอกสาร / พิมพ์เกินจำนวนที่จำกัด / แก้อัตรา-ยอดภาษี (`AP/BC5Account.cpp:6126-6160`, `FW/mybasedbformview.cpp:2641`) |

---

## 12. พิมพ์ / ตัวอย่างก่อนพิมพ์ / รายงาน

### 12.1 พิมพ์เอกสาร (Ctrl+P หรือปุ่มพิมพ์) — `FW/mybasedbformview.cpp:2538-2760, 4097-4112`

1. ปุ่มพิมพ์ใช้ได้เฉพาะ **เอกสารที่บันทึกแล้ว** (ไม่ใช่โหมดเอกสารใหม่) และมีสิทธิ์ `bCanPrintBill` (`:220`); ไม่มีสิทธิ์ → "Use not have authorization..."; ไม่มีแบบฟอร์ม → "Please enter form for this transaction ..." (`:4097-4108`)
2. ตรวจก่อนพิมพ์ `OnCheckBeforePrint(ตาราง, เลขที่)` ของแต่ละจอ (เช่น `AP/ChmFrmGoodsPurchase.cpp` `OnHKPrint`)
3. เปิดไดอะล็อก **เลือกแบบฟอร์ม** (`CSelectBillFormDlg`, caption "Select Form", id `IDD_SELECT_DO_FORM_DLG` ที่ `FW/BC5FrmWrk.rc:124-155`, โค้ด `FW/SelectBillFormDlg.cpp`) — **ป้ายในไดอะล็อกเป็นอังกฤษทั้งหมด** (ไม่มีรูปแบบ `KEY$ไทย`):
   - กลุ่ม "Data Section": **From:** / **To:** (ช่วงเลขที่เอกสาร `IDC_FORM_START/STOP` — พิมพ์ช่วงได้), ช่อง "Select Form:" (combo "Form List" ของแบบฟอร์มจาก `BCSysTRFom`)
   - ติ๊กเลือกชนิดที่จะพิมพ์: "Document Form" · "GL Slip Form" · "Withhold. Tax Form" · "Print Serial Number" (ตัวติ๊กที่ไม่เกี่ยวกับเอกสารนั้นถูก disable)
   - ปุ่ม "Print" (`ID_PRINT`), "Preview" (`ID_PREVIEW`), "Printer Setup..." (`ID_PRINTER_SETUP` ปิดไว้), "Cancel"
   - Preview เปิดได้เมื่อเลือก *ตัวเลือกเดียว*; Print ต้องเลือกอย่างน้อยหนึ่งตัวเลือก (`FW/SelectBillFormDlg.cpp:160-196`)
4. **จำกัดจำนวนครั้งพิมพ์**: ค่าจาก `BCSysTRFormat.PrintLimit` (0 = ไม่จำกัด) หรือรายแบบฟอร์ม `BCSysTRFom.PrintLimit` ถ้า config `PRINT_LIMIT_BY_FORM=1` (โค้ด `FW/mybasedbformview.cpp:2662-2672` อ่านคอลัมน์นี้ แต่ DDL ใน `SQLSERVER_Script.sql` ของตาราง `BCSysTRFom` ไม่มีคอลัมน์ `PrintLimit` — ยังไม่ verify ว่ามาจากสคริปต์ปรับปรุงอื่น); นับจาก `BCPrintLog` (`docno, tablename, printtype=1, formname`); เกิน → "CANNOT_PRINT_OVERLIMIT" (ผู้มีสิทธิ์ `isPrintOverLimit` ข้ามได้) `FW/mybasedbformview.cpp:2641-2700`, `FW/SelectBillFormDlg.cpp:312+`
5. พิมพ์จริงโดยเอนจินแบบฟอร์ม (`m_Form.GetFileData` อ่านไฟล์แบบฟอร์มจากโฟลเดอร์ `GetFomFolder`; เอนจิน/ตัวออกแบบแบบฟอร์มอยู่ในโฟลเดอร์ champ-lib BOLEDBFORM และ BOLEDBFORMEDIT (ยังไม่ verify โครงสร้างไฟล์แบบฟอร์ม)) และบันทึกประวัติพิมพ์ลง `BCPRINTLOG (PrintType, DocNo, TableName, UserID, FormName, Printdate)`
6. พิมพ์อัตโนมัติหลังบันทึกเมื่อบริษัทตั้ง `PrintBill=1` (`FW/mybasedbformview.cpp:2543-2546`); โหมดพิมพ์ข้อความ (dot-matrix) เมื่อ `PrintFrmInText` (`:2557-2590`)

### 12.2 หน้าต่างรายงาน — `RP/BaseReportView.cpp`

| ส่วน | ฟังก์ชัน | หลักฐาน |
|---|---|---|
| แถบเครื่องมือรายงาน | พิมพ์ (Ctrl+P) · หน้าก่อน · หน้าถัดไป · ไปหน้า · ค้นหา · ส่งออก · รีเฟรช · คุณสมบัติคอลัมน์ · แสดง/ซ่อน · สี | `FW/BC5FrmWrk.rc:1049-1065` |
| คีย์ | F11 = ตัวอย่างก่อนพิมพ์; PageUp/PageDown = หน้าก่อน/ถัดไป; Ctrl+F = ค้นหาข้อมูลในรายงาน | `RP/BaseReportView.cpp:783-797, 1706` |
| คลิกขวา (หัวคอลัมน์) | "Column Property..." · "Display Hide Column" · "Select Column Color" · "Add My Column..." · "PrintPreview" | `:1697-1704` |
| คลิกขวา (เนื้อหา) | "Find...\tCtrl+F" · "Export..." · "PrintPreview" · "Print..." · "Save report config" · "Get report config" (เก็บ/ดึงค่าแสดงผลรายงานของผู้ใช้ไว้ที่ server) | `:1706-1714` |
| สิทธิ์ | เปลี่ยน Report Option = `bCanChangeReportOption`; ส่งออก = `bCanExportReportData`; พิมพ์ = `bCanPrintReport` | หัวข้อ 11 |

---

## 13. รูปแบบวันที่ · ตัวเลข · ข้อความ

### 13.1 วันที่ (ปี พ.ศ./ค.ศ. ตามบริษัท `YearType`: 0 = ไทย, 1 = อังกฤษ — `OC/BEXTools.cpp:53`, `FW/BCSWinApp.cpp:1719-1733`)

| เรื่อง | พฤติกรรม | หลักฐาน |
|---|---|---|
| ช่องวันที่ว่างเมื่อโฟกัสแรก | เติมแม่แบบ `วว/MM/ปี พ.ศ.` (ไทย) หรือ `dd/MM/yyyy` (อังกฤษ) โดยเดือน/ปีเป็นของวันนี้ ผู้ใช้พิมพ์เฉพาะวัน | `OC/BMergeEdit.cpp:864-868` |
| การแปลงค่า | แยก วัน/เดือน/ปี (`XSeparateDate`) → ถ้าไทย ลบ 543 ก่อนตรวจจริง; ไม่ใช่วันที่ถูกต้อง หรือปีนอกช่วง (อังกฤษ >2300, ไทยแปลงแล้ว <1900) → **ล้างช่อง** | `:1256-1269` |
| วันที่อัตโนมัติ | ชนิด `_AUTO` เติมวันนี้เมื่อออกจากช่องว่าง | `:591-596` |
| ปฏิทิน | ปุ่มด้านขวา เลือกวันแล้วใส่ตามชนิดปี | `OC/BMergeBttn.cpp OnPopup` |
| เลขที่เอกสารอ้างอิงปี | `YY` ใช้ปี พ.ศ. 2 หลักท้ายถ้าบริษัทเป็นปีไทย | `OC/BRunningBttn.cpp:121-123` |

### 13.2 ตัวเลข (`XMoneyFormat`, `OC/BEXTools.cpp:337-376`)

| เรื่อง | กฎ |
|---|---|
| ตัวคั่นหลักพัน | `,` ทุก 3 หลัก |
| จุดทศนิยม | ตามจำนวนที่ตั้งของบริษัท ส่งเป็นพารามิเตอร์ `nPoint` — แยกตามชนิด: `QtyPoint` (จำนวน), `PricePoint` (ราคา), `BuyPricePoint` (ราคาซื้อ), `AmountPoint` (จำนวนเงิน), `ItemAmountPoint` (เงินรายแถว), `CostPoint` (ต้นทุน; ค่าเริ่มต้น 4) — ค่าเริ่มต้นคือ 2 `FW/companyinfo.cpp:80-85` |
| ค่าศูนย์ | **แสดงเป็นช่องว่าง** (คืน `" "`) ในรายการ/รายงานที่ใช้ฟังก์ชันนี้ (`:343-345`) |
| ค่าติดลบ | ใส่เครื่องหมายนำหน้า (`-`) / ครอบ (ผู้เรียกกำหนดได้; ฟอร์มส่ง `"-"`,`""`) |
| ช่องกรอกจำนวนเงิน (`BME_CALC`) | ว่าง = `0.00`; ออกจากช่องตัด `,` แล้วจัดรูปใหม่ตามจำนวนทศนิยม (`OC/BMergeEdit.cpp:996-1023`) |
| การคำนวณเงิน | ใช้ `double` (ต่างจากกฎ BC Rule 20 — ไม่ลอก) |

### 13.3 ข้อความและการแจ้งเตือน

| ชนิด | รูปแบบ Champ |
|---|---|
| ยืนยัน/เตือน/สำเร็จ | `AfxMessageBox` modal ปุ่ม OK หรือ Yes/No(/Cancel): สำเร็จ = "Complete" |
| ข้อความ validation | เรียกคีย์ `GetMyRes("KEY")` (ภาษาตามบริษัท `nResMode`) — ไม่แสดงรหัส error ดิบ |
| ข้อความภาษาอังกฤษตรง ๆ ที่ผู้ใช้เห็น | "Your permission can't ... record", "Can not Load Configuration [...]", "Job is running please wait...", "Complete", "Data Not Found ...!", "Record Empty", "QRowset Not Open" — ใน BC ต้องแปลเป็นไทยที่บอกวิธีแก้ |
| ป้ายปุ่ม | ไดอะล็อก: `OK$ตกลง`, `CANCEL$ยกเลิก`, `Help=F1`; แถบเครื่องมือฟอร์ม: อังกฤษ (New/Save/Del/Find/Prev/Next/Close/Info/Key/Recur) |

---

## 14. เข้าระบบและเลือกบริษัท (บันทึกไว้เพื่ออ้างอิงเท่านั้น — BC คงของเดิม ตามคำสั่งเจ้าของ)

| เรื่อง | Champ | หลักฐาน |
|---|---|---|
| หน้าเข้าระบบ | ช่อง **ชื่อบริษัท** (เลือกจากตาราง `DBCONNECTION` ใน Access: `SERVER_DSN, DATABASE_NAME, COMPANY_NAME, PROVIDER`) + **รหัสผู้ใช้** + **รหัสผ่าน** (ใช้เป็นผู้ใช้/รหัสผ่านของฐานข้อมูลโดยตรง); ไม่มีกล่องเลือกสาขา — จำบริษัท/ผู้ใช้ล่าสุดใน registry | `AP/XpLogin.cpp:147, 200-210, 862-960`, `FW/userlogon.cpp:44-95` |
| ผิดพลาด | ไม่ใส่ชื่อบริษัท/ผู้ใช้ → "Please enter company name/user ID"; เชื่อมไม่ได้ → คีย์ `NOT_CONNET`; หลังเชื่อมต่อ ตรวจลิขสิทธิ์/อายุบริการ/จำนวนผู้ใช้พร้อมกัน | `XpLogin.cpp:874-956, 1037-1124` |

> **BC ไม่ปรับส่วนนี้** (เข้าระบบ + เลือก holding/บริษัท/สาขา คงตามเดิม)

---

## 15. BC เทียบ Champ และสิ่งที่ต้องปรับ

ขอบเขต: กรอบที่ทุกจอใช้ร่วมกัน (เรียงตามผลต่อผู้ใช้ — มากไปน้อย) · "คอมโพเนนต์กลาง" = ทำครั้งเดียวใน `FE/components` / `FE/hooks` แล้วทุกจอใช้ตามกฎ skill `datacrud` / `ui-scale-polish` (ผู้ใช้หลักคือคนไทยอายุ 40+ → ป้ายไทย ปุ่มใหญ่ ไม่ใช้ไอคอนเปล่า)

| # | หัวข้อ | Champ | BC ตอนนี้ | สถานะ | สิ่งที่ต้องปรับให้เหมือน Champ |
|---|---|---|---|---|---|
| 1 | Enter = ช่องถัดไป ทั้งฟอร์ม | Enter/↓ ถัดไป, ↑ ก่อนหน้า, ช่องหลายบรรทัด Ctrl+Enter (`FW/BaseFormView.cpp:207-277`) | Enter ในช่องหัวเอกสารส่งฟอร์ม → บันทึกฉบับร่าง (`FE/app/gl/gl-journals.tsx:797`); มีเฉพาะตาราง `useTabularEnterNav` (`FE/hooks/use-tabular-enter-nav.ts:46-120`, ใช้ `gl-journals.tsx:863`) | **ต่าง (เสี่ยงบันทึกพลาด)** | ทำ hook `useFormEnterNav` ครอบ `<form>` ทุกจอ: Enter/↓ → ช่องถัดไป, ↑ → ก่อนหน้า (ยกเว้น textarea/combobox), Ctrl+Enter ใน textarea = ถัดไป; **ห้ามให้ Enter ส่งฟอร์ม** — บันทึกด้วยปุ่ม/F11/Ctrl+S เท่านั้น; ข้ามเฉพาะปุ่ม (Enter = กดปุ่ม); ช่องรหัสให้ใช้ข้อ 2 |
| 2 | ช่องรหัสอ้างอิง (บัญชี ลูกค้า สินค้า …) | พิมพ์รหัส + Enter/Tab ตรวจตรง; ไม่พบ → ข้อความ + ล้างช่อง + เปิดรายการ; F4 / `?` / `#` เปิดรายการ; รหัสเป็นตัวพิมพ์ใหญ่ (`OC/BMergeEdit.cpp:426, 587-590, 678-760`) | `AccountSelect` เป็น `<select>` พิมพ์รหัสไม่ได้; Enter/Space/F2/คลิก เปิด dialog (`FE/app/gl/gl-common.tsx:468-493`); dialog ค้นหาแบบ Enter เลือก (`FE/app/gl/account-search-dialog.tsx:218-251`) | **ต่าง** | ทำคอมโพเนนต์กลาง `CodeLookupField` (input + ปุ่มค้นหา): พิมพ์รหัสแล้ว Enter/Tab = ตรวจกับรายการที่โหลดไว้; ไม่พบ → ข้อความไทย "ไม่พบข้อมูล" + ล้างช่อง + เปิดรายการค้นหาอัตโนมัติ; **F4** เปิดรายการ (เก็บ F2 ไว้เป็นตัวสำรอง); คำค้นขึ้นต้น `?`/`#` เปิดรายการ; รายการค้นหา "มีอยู่ส่วนใดของ **รหัสหรือชื่อ**" (`LIKE %คำ%`) + คลิกหัวคอลัมน์เรียง; รหัสพิมพ์ใหญ่อัตโนมัติ; Enter ในช่องต้องไม่ชนกับ `useTabularEnterNav` (ยังไม่ verify ในเบราว์เซอร์ว่าปัจจุบัน Enter บน select เปิด dialog และเลื่อนโฟกัสพร้อมกันหรือไม่ — handler ตาราง `use-tabular-enter-nav.ts:46-75` ไม่เช็ค `defaultPrevented`) |
| 3 | แถบปุ่มเอกสารมาตรฐาน | ใหม่ · บันทึก · ลบ · ค้นหา · ก่อนหน้า · ถัดไป · ปิด · ข้อมูล · คีย์ลัด (ลำดับคงที่ทุกจอ `FW/BC5FrmWrk.rc:999-1015`) | ปุ่มวางต่างตามจอ/โหมด: ค้นหา/โหลดใหม่/เพิ่มรายการ (`gl-journals.tsx:395-403`), พิมพ์/แก้ไขฉบับร่าง/ปิด (`:636-650`), บันทึก/ยกเลิก/ลบ (`:1052-1067`) ไม่มี ก่อนหน้า/ถัดไป/ข้อมูล/คีย์ลัด | **ต่าง** | คอมโพเนนต์กลาง `DocToolbar` ติดบนหัวฟอร์ม ลำดับตาม Champ ป้ายไทย (ใหม่ · บันทึก · ลบ · ค้นหา · ก่อนหน้า · ถัดไป · ปิด · ข้อมูล · คีย์ลัด) ใส่ป้ายคีย์ลัดบนปุ่มเหมือน tooltip Champ; ปุ่มที่จอนั้นไม่รองรับให้ disabled ไม่ซ่อน (ตำแหน่งคงที่); ป้ายภาษาไทย (Champ ใช้อังกฤษ New/Save/Del… — ตัดสินใจแล้วตามกฎ Thai-First ของ AGENTS.md; ให้ lead ยืนยัน) |
| 4 | คีย์ลัด | Ctrl+N ใหม่ · Ctrl+F ค้นหา · F11 บันทึก · Ctrl+P พิมพ์ · Ctrl+D ลบ · Ctrl+Q/W ก่อนหน้า/ถัดไป · Ctrl+F4 ปิด · F12 ตรวจเอกสาร · Shift+F5 ยืนยัน · Ctrl+F5 ยกเลิก · F1 ช่วยเหลือ (`FW/BC5FrmWrk.rc:1085-1136`, `AP/BC5Account.rc:977-1010`) | Ctrl+S บันทึก · Alt+N ใหม่ · Esc ยกเลิก (`FE/hooks/use-form-shortcuts.ts:45-75`) | **ขาดมาก** | ขยาย `useFormShortcuts` เป็นแผนที่คีย์เดียวทั้งระบบ: เพิ่ม F11 (บันทึก), Ctrl+F (ค้นหาในจอ), Ctrl+P (พิมพ์), Ctrl+D (ลบ), Ctrl+Q/Ctrl+W (ก่อนหน้า/ถัดไป), F12 (ตรวจเอกสาร/เส้นทางเอกสาร), Shift+F5 (ยืนยัน), Ctrl+F5 (ยกเลิก), F1 (วิธีใช้) + ปุ่ม "คีย์ลัด" แสดงรายการ **ยังไม่ verify:** เบราว์เซอร์ดักบางคีย์ (Ctrl+N/Ctrl+W/Ctrl+T สงวนโดยเบราว์เซอร์; F5/F11/F12 ถูกใช้โดยเบราว์เซอร์) — ต้องลองจริงว่า `preventDefault` ใช้ได้ไหม ถ้าไม่ได้ให้ใช้ชุดสำรอง (Alt+N ที่มีอยู่, Ctrl+S, F-key ที่ดักได้) และบอกในปุ่ม "คีย์ลัด" |
| 5 | บันทึกแล้วพร้อมคีย์ใบถัดไป | บันทึก → ล้างจอ → โหมดเอกสารใหม่ (สวิตช์ keep/clear) (`OC/BaseDBFormView.cpp:884-981`, `FW/mybasedbformview.cpp:4113-4124`) | บันทึก → แสดงใบเดิมโหมดดู ต้องกด "บันทึกรายวันใหม่" (`gl-journals.tsx:289-294`, ปุ่มที่ `:1089`) | **ต่าง** | หลังบันทึกสำเร็จ: ค่าเริ่มต้น = เปิดใบใหม่ว่างทันทีและโฟกัสช่องแรก (เก็บค่าที่ผู้ใช้เลือกซ้ำ เช่น สมุดรายวัน/วันที่); มีสวิตช์ "คงข้อมูลไว้/ล้างจอหลังบันทึก" จำค่าใน localStorage; แสดงข้อความสำเร็จเป็น toast แล้วไปต่อ (ไม่ขวางด้วย modal) |
| 6 | เลขที่เอกสารอัตโนมัติ | `DocFormat` + `YY/MM/DD/@/#` + กลุ่ม + ล็อกด้วยสิทธิ์ (`OC/BRunningBttn.cpp:121-433`) | เลขที่พิมพ์เอง ค่าเริ่มต้นว่าง `maxLength 60` (`FE/lib/general-ledger.ts:280`, `gl-journals.tsx:825`); backend ห้ามแก้เลขที่หลังบันทึก (`backend/internal/generalledger/journal_books.go:59`) | **ยังไม่มี** | เพิ่มรูปแบบเลขที่ต่อสมุดรายวัน/ประเภทเอกสาร (ฟิลด์ใน master สมุดรายวัน + ตารางตัวนับ PostgreSQL) รองรับ `YY` (พ.ศ. 2 หลัก) `MM` `DD` `@กลุ่ม` `####`; จองเลขใน transaction เดียวกับการบันทึก (ห้าม `SELECT MAX` แบบ Champ — เลขซ้ำเมื่อคีย์พร้อมกัน); แสดงเลขถัดไปในช่องเมื่อเปิดใบใหม่; ผู้ไม่มีสิทธิ์ "แก้ไขเลขที่" ช่องเป็นอ่านอย่างเดียว; ตัวเลขในสมุดรายวัน (ผู้ใช้เพิ่ม/ลบได้) ดู `journal-books-user-defined` — เพิ่มเมื่อ lead อนุมัติ (ไม่อยู่ใน Champ GL เป็นฟิลด์เดียวกัน ยังไม่ verify กับ GL ของ Champ) |
| 7 | ถามก่อนทิ้งข้อมูลที่ยังไม่บันทึก | ใหม่/ค้นหา/ก่อนหน้า/ถัดไป/ปิด → Yes (บันทึกแล้วไปต่อ) / No (ทิ้ง) / Cancel (อยู่ต่อ) (`OC/BaseDBFormView.cpp:687, 728, 814`) | ไดอะล็อก 2 ทาง "ละทิ้ง/อยู่ต่อ" (`gl-journals.tsx:247, 259`; `FE/app/crud/erp-crud-workbench.tsx:133-140`) | **ต่าง** | `useConfirmDialog` ชนิด 3 ปุ่ม: "บันทึกแล้วไปต่อ" · "ไม่บันทึก" · "ยกเลิก" (ข้อความไทยตามเจตนา "ต้องการบันทึกการเปลี่ยนแปลงหรือไม่?") ใช้กับทุกการออกจากฟอร์มที่แก้ค้าง (เปลี่ยนแถว/ปิดแท็บ/ใหม่/ค้นหา/ก่อนหน้า/ถัดไป) — ปุ่ม "บันทึกแล้วไปต่อ" ถ้าตรวจพบ error ต้องอยู่ต่อ (ตามที่ Champ ทำ `getLastError 100/101`) |
| 8 | ก่อนหน้า / ถัดไป | เอกสารก่อนหน้า/ถัดไปตามลำดับรายการ (Ctrl+Q / Ctrl+W) (`OC/BaseDBFormView.cpp:717, 803`) | ไม่มี — ต้องกลับไปคลิกแถวในรายการ | **ยังไม่มี** | เพิ่มใน `DocToolbar` — เลื่อนเอกสารตามรายการที่กรอง/เรียงอยู่ (ใช้ list ที่โหลดอยู่ + เปลี่ยนหน้าอัตโนมัติ) พร้อม dirty-guard ข้อ 7 |
| 9 | สถานะเอกสาร + สีแถว | ป้าย CC (แดง) / 0.00 (น้ำตาล) / CF (เขียวน้ำเงิน) / N; แถวในรายการ: ยกเลิก=แดง, คงค้าง 0=เทา, ยืนยัน=เขียวน้ำเงิน (`AP/myespeciallyformview.cpp:5685-5720`, `OC/MyExBillQList.cpp:118-160`) | GL: draft/posted/reversed ผ่านป้ายสถานะ (`gl-journals.tsx:628-630`); ERP workbench: ฉบับร่าง/อนุมัติ (`FE/app/crud/erp-crud-workbench.tsx:373-374, 608-613`) | **ต่าง** | ป้ายสถานะมุมบนฟอร์มเอกสาร ERP ตามกติกา Champ (ยกเลิก/ยอดคงค้างศูนย์/ยืนยันแล้ว/ปกติ) **ต้องมี "ข้อความ" กำกับสี** (กฎ 40+ ห้ามสื่อด้วยสีอย่างเดียว) — ใช้ข้อความไทย + ตัวย่อ Champ ในวงเล็บ; GL ใช้สถานะของ BC ต่อ (ดูหัวข้อ "ชนกฎเจ้าของ") |
| 10 | ยืนยัน / ยกเลิกเอกสาร | ไดอะล็อกต้องเลือก **Yes** ใน combobox แล้วกดตกลง พร้อมคำเตือนไทยยาว; เก็บผู้ทำ + เวลา; ห้ามทำซ้ำ; ยกเลิกแล้วห้ามนำมาใช้ต่อ (`AP/BC5Account.rc:3490-3530`, `AP/MyDlgBill*.cpp`) | GL: ผ่านรายการ/สร้างรายการกลับ ต้องระบุเหตุผล + วันที่ + เลขที่ใบกลับ + confirm dialog (`gl-journals.tsx:332-345`) | **BC เข้มกว่า (คงไว้)** | ใช้ถ้อยคำเตือนไทยของ Champ ในไดอะล็อกยืนยัน/ยกเลิกของ ERP (ข้อความเต็มหัวข้อ 10.3) แต่เปลี่ยน combobox Yes/No เป็นปุ่ม "ยืนยัน/ยกเลิก" แบบ confirm-dialog ของ BC; **คงช่องเหตุผลของ BC** (Champ ไม่มี — BC เข้มกว่าเพื่อ audit) |
| 11 | ล็อกเอกสาร | ล็อกเมื่อ: ยืนยันแล้ว/ยกเลิกแล้ว/ถูกอ้างอิงต่อ/ลงบัญชีแล้ว (`IsPostGL`)/งวดปิด (`BCPeriod.Status=0`); ข้อความบอกเหตุผล (`AP/billfrmsubmitprice.cpp:457-560`, `FW/mybasedbformview.cpp:4126-4162`, `AP/mytools.cpp:834-865`) | backend GL ห้ามแก้เลขที่/สมุดหลังบันทึก (`journal_books.go:59`); ผ่านรายการแล้วแก้ผ่านรายการกลับ | **บางส่วน** | UI แสดงเหตุผลที่แก้/ลบไม่ได้เป็นข้อความไทยบนฟอร์ม (ไม่ใช่แค่ disable ปุ่ม): "เอกสารนี้ยืนยันแล้ว/ยกเลิกแล้ว/ลงบัญชีแล้ว/อยู่ในงวดที่ปิดแล้ว" พร้อมข้อมูลชุดผ่านรายการ (ผู้ทำ/วันที่/เลขชุด) ตาม `POSTGL_WARN`; ตรวจงวดปิดก่อนบันทึกทุกครั้ง |
| 12 | ตารางรายการ (กริด) | Insert แทรกแถว · Delete ลบแถวที่เลือก · Ctrl+Insert/Shift+Insert คัดลอก/วาง · Shift+↑↓ เลือกหลายแถว · F4 ค้นหาในเซลล์ · Tab ท้ายตารางออกไปช่องถัดไป (`OC/BInputListWnd.cpp:561-686, 2199-2294`) | Enter ไปเซลล์ถัดไป/ท้ายตารางเพิ่มแถว (`FE/hooks/use-tabular-enter-nav.ts:46-120`); วาง Excel (`gl-journals.tsx:866-873`); ลบแถวด้วยปุ่ม "นำออก" (`:905`) | **บางส่วน** | เพิ่มคีย์ในกริดกลาง: **Insert** แทรกแถว, **Delete** (เมื่อเลือกแถว) ลบแถว, **Ctrl+Insert/Shift+Insert** คัดลอก/วางแถว (เก็บ Ctrl+C/V ปกติของเบราว์เซอร์ด้วย), **F4** เปิดรายการในเซลล์ชนิดรหัส, **Shift+↑/↓** เลือกหลายแถว, **↑/↓** ย้ายแถว; Tab ที่เซลล์สุดท้ายของแถวสุดท้าย → ไปช่องถัดไปนอกตาราง (ไม่ใช่เพิ่มแถว — Champ ให้ **Enter** เพิ่มแถว/Tab ออกจากตาราง; ตอนนี้ BC ให้ Tab ที่เซลล์สุดท้ายเพิ่มแถวใหม่ `use-tabular-enter-nav.ts:62-70`) |
| 13 | พิมพ์ | เลือกแบบฟอร์ม + ช่วงเลขที่ + ตัวเลือกพิมพ์ (แบบฟอร์ม/Slip/หัก ณ ที่จ่าย/Serial) + ดูตัวอย่าง; พิมพ์ได้เฉพาะใบที่บันทึกแล้ว; จำกัดจำนวนครั้ง `PrintLimit`; บันทึก `BCPRINTLOG` (`FW/SelectBillFormDlg.cpp`, `FW/mybasedbformview.cpp:2538-2700`) | `window.print()` ตรง ๆ จากโมดัลพิมพ์ (`FE/components/thai-document-print-modal.tsx:82-90`; `FE/app/gl/gl-print.tsx:58-69`); ไม่มีประวัติ/จำกัด (ค้นไม่พบ `printcount/printlimit/printlog` ใน backend/frontend) | **ต่าง** | ไดอะล็อก "พิมพ์เอกสาร" กลาง: เลือกแบบฟอร์ม (ตามประเภทเอกสาร), ช่วงเลขที่, ดูตัวอย่าง/พิมพ์; บันทึกประวัติพิมพ์ในตาราง PostgreSQL ใหม่ (`print_log`: ประเภท, เลขที่, ตาราง, ผู้ใช้, ชื่อแบบฟอร์ม, เวลา UTC — เพิ่มได้ ไม่ต้อง migration); จำกัดจำนวนครั้งต่อประเภทเอกสาร + สิทธิ์พิมพ์เกิน; ปุ่มพิมพ์เปิดเฉพาะเอกสารที่บันทึกแล้ว; ถามก่อนพิมพ์ซ้ำเมื่อเคยพิมพ์ ("พิมพ์ครั้งที่ N") |
| 14 | วันที่ | พิมพ์ `วว/เดือน/พ.ศ.` มีแม่แบบ; วันที่อัตโนมัติ; ผิดรูป/ปีนอกช่วง = ล้างช่อง (`OC/BMergeEdit.cpp:858-870, 1250-1275`) | `<input type="date">` ของเบราว์เซอร์ (รูปแบบขึ้นกับเครื่องผู้ใช้) + ข้อความ พ.ศ. ใต้ช่อง (`FE/components/ui/date-time-field.tsx:35-39`, `FE/lib/date-time.ts:16, 52-54`); `gl-journals.tsx:673, 830` | **ต่าง** | ช่องวันที่กลางพิมพ์ได้ `dd/MM/yyyy` (ปีไทย พ.ศ. ตามบริษัท) พร้อมปุ่มปฏิทิน; Enter/Tab ตรวจรูปแบบ, ผิด = ล้าง + แจ้ง; ช่องวันที่เอกสารใหม่เติมวันนี้; พิมพ์เฉพาะวันได้ (เดือน/ปีเติมให้) |
| 15 | ตัวเลข/จำนวนเงิน | ทศนิยมตามชนิด (จำนวน/ราคา/ราคาซื้อ/จำนวนเงิน/ต้นทุน) จากข้อมูลบริษัท; ศูนย์แสดงว่าง; ว่าง=0.00 ตอนออกจากช่อง (`FW/companyinfo.cpp:80-85`, `OC/BEXTools.cpp:337-376`, `OC/BMergeEdit.cpp:996-1023`) | `AmountInput` + `formatAmount` ตามสเกลของปีบัญชี (`FE/app/gl/gl-common.tsx:326-443`, `FE/lib/general-ledger.ts:317-336`); ศูนย์แสดง "—" ในโหมดดู (`gl-journals.tsx:700`) | **ใกล้เคียง** | คงของ BC (ใช้สตริงทศนิยม ไม่ใช้ double ตาม Rule 20); เพิ่มชนิดจำนวนทศนิยมของ "จำนวน/ราคา/ต้นทุน" ในตั้งค่าบริษัทเมื่อทำเอกสาร ERP; ยืนยันกฎแสดงศูนย์ (Champ ว่าง, BC "—") ให้ลุงจืดเลือกหนึ่งแบบทั้งระบบ |
| 16 | ค้นหาเอกสาร / รายการ | หน้าต่างรายการ: ค้นสดหรือ Enter, ขึ้นต้น/มีอยู่ส่วนใด, แสดงทั้งหมด/บนสุด, F6-F9 เปลี่ยนหน้า, F2 ล้างตัวกรอง, สีแถวตามสถานะ (`QL/QOLEDBLib.rc:256-275, 627-642`) | ช่องค้นหา + ดีเลย์ (`FE/app/gl/gl-common.tsx:115-243`) + `Pager` 30 แถว (`:659`) | **ใกล้เคียง** | เพิ่มตัวเลือก "ขึ้นต้นด้วย / มีอยู่ส่วนใด" และคีย์เปลี่ยนหน้า/ล้างตัวกรองในรายการกลาง; ใส่สีแถว/ป้ายสถานะข้อ 9 |
| 17 | หน้าต่าง/แท็บ | หนึ่งจอหนึ่งหน้าต่าง; ปิดทุกหน้าต่าง; ชื่อหน้าต่างมีบริษัท:ผู้ใช้ (`FW/BCSWinApp.cpp:2542-2593`, `FW/mybasedbformview.cpp:255-262`) | แท็บต่อเส้นทาง (ใช้ซ้ำแท็บเดิม `FE/app/menu/main-menu-screen.tsx:702`); ปิดแท็บ + ถามทิ้งข้อมูล (`:777-784`); ลากสลับแท็บ (`:786+`); ไม่มี "ปิดทั้งหมด" | **ใกล้เคียง** | เพิ่มปุ่ม/เมนู "ปิดทุกหน้าต่าง" (มี dirty-guard ทีละแท็บ); แท็บแสดงบริษัท/สาขาที่ทำงานอยู่ถ้ายังไม่แสดง (ยังไม่ verify หัวแท็บปัจจุบัน) |
| 18 | ทางลัดซ้าย | "ทางลัด BC" 7 รายการเริ่มต้น + "ทางลัดของฉัน" สร้างเอง (`AP/MainFrm.cpp:1132-1186`) | จอ "จัดการทางลัด" + ทางลัดของผู้ใช้ (`FE/app/menu/manage-shortcuts-screen.tsx`, `main-menu-screen.tsx:737-748`) | **มีแล้ว** | ตรวจรายการเริ่มต้นให้ตรง Champ (ข้อมูลบริษัท · ข้อมูลหลัก · รูปแบบรายวัน · สินค้า · ลูกค้า · เจ้าหนี้ · ผังบัญชี) — ยังไม่ verify รายการเริ่มต้นของ BC |
| 19 | สิทธิ์รายจอ | `IsMake/IsInsert/IsEdit/IsDelete` ต่อเมนู + ธง 16 ข้อ (แก้ราคา/ส่วนลด/ดูต้นทุน/ยืนยัน/ยกเลิก/แก้เลขที่/ส่งออก/พิมพ์ ฯลฯ) (`FW/BCSWinApp.cpp:885-922`, `OC/BaseDBFormView.cpp:1642-1760`, `AP/BC5Account.cpp:6190-6215`) | บทบาท + `permission_sets` + `access_scopes` ที่ระดับ holding (`backend/internal/generalledger/httpapi/scope.go:50-96`); การซ่อน/ปิดปุ่มตามสิทธิ์รายจอใน UI — ยังไม่ verify (`gl-journals.tsx` ใช้เฉพาะ `mode` ของเส้นทาง `:89-93`) | **ยังไม่ verify** | UI ปิดปุ่ม ใหม่/บันทึก/ลบ/ยืนยัน/ยกเลิก/พิมพ์/ส่งออก ตามสิทธิ์พร้อมข้อความไทย "คุณไม่มีสิทธิ์ …" (ข้อความ Champ คีย์ `CANNOT_OPEN` = "คุณไม่มีสิทธิเข้าหน้าจอนี้" ตามคอมเมนต์); backend เป็นตัวบังคับจริง (Rule 10) — จัดชุดสิทธิ์ของ BC ให้ครอบคลุมธง 16 ข้อ **เฉพาะข้อที่อยู่ในขอบเขต GL ตอนนี้** (ยืนยัน/ยกเลิก/พิมพ์/ส่งออก/แก้เลขที่/แก้วันที่/ดูต้นทุนไม่เกี่ยว GL) |
| 20 | วิธีใช้ / ข้อมูลเสริม | F1 = วิธีใช้ของจอ; ปุ่ม Info = เส้นทางเอกสาร + ยอดตัวเลขต่าง ๆ; F12 = ตรวจสอบเอกสาร (`FW/mybasedbformview.cpp:2262-2300`, `AP/ChmFrmGoodsPurchase.cpp:402-425`) | มีลิงก์คู่มือ (`FE/app/manual/`, `manual-link.tsx`); ไม่มีเส้นทางเอกสารในฟอร์ม | **ต่าง** | ปุ่ม "ข้อมูล" บน `DocToolbar`: "แสดงเส้นทางเอกสาร" (เอกสารต้นทาง/ปลายทาง) + "แสดงยอดตัวเลขต่าง ๆ"; F1 เปิดคู่มือตรงหัวข้อของจอ |
| 21 | ข้อความ/การแจ้งเตือน | modal "Complete"; ข้อความตามคีย์; ยังมีอังกฤษหลุด | toast + `Notice` (`FE/components/toast-viewport.tsx`, `gl-common.tsx:94`); ข้อความไทยผ่านคีย์ | **BC ดีกว่า (คงไว้)** | ใช้ถ้อยคำไทยของ Champ ที่ได้ยืนยันจากไดอะล็อก (หัวข้อ 10.3) เป็นต้นแบบ; ข้อความที่ Champ ปล่อยอังกฤษ ให้เขียนไทยที่บอกวิธีแก้; แจ้งรายการแปลที่ยังไม่แน่ใจให้ลุงจืด |
| 22 | เข้าระบบ / เลือก holding–บริษัท–สาขา | บริษัท + ผู้ใช้ + รหัสผ่านของ DB (`AP/XpLogin.cpp`) | Google/Demo login + เลือก holding/บริษัท/สาขา (`FE/app/login-screen.tsx`, `FE/app/holding/`) | **ไม่ปรับ — คำสั่งเจ้าของ** | คงไว้ตามเดิม |

### ชนกับกฎของเจ้าของ — ไม่เสนอให้เปลี่ยน (ต้องให้ลุงจืดตัดสิน)

| เรื่อง | Champ | กฎ/ข้อมูลฝั่ง BC | ข้อสรุป |
|---|---|---|---|
| บันทึก = ลงบัญชีทันที ไม่มีฉบับร่าง | `Posting_GL` ต่อท้ายการบันทึก (`OC/BaseDBFormView.cpp:927`) | GL ของ BC แยก ฉบับร่าง → ผ่านรายการ (`gl-journals.tsx:332-345`) ตามสเปก GL ของเจ้าของ (ยังไม่ได้เปิด `mydocs/` เทียบ — ห้ามแตะ) | **ไม่ลอก** คงฉบับร่าง/ผ่านรายการ; ข้อ 5 (พร้อมคีย์ใบถัดไป) ทำได้ทั้งสองแบบ |
| ลบจริง (`DELETE FROM`) เอกสารที่ยังไม่ยืนยัน | `AP/billfrmsubmitprice.cpp:511-525` | กฎ GL: ผ่านบัญชีแล้วแก้ด้วยรายการกลับ ห้ามแก้/ลบเงียบ; ห้ามลบผังบัญชีที่มีรายการอ้างอิง | ฉบับร่างลบได้ตามที่ BC ทำอยู่ (`gl-journals.tsx` ปุ่ม "ลบฉบับร่าง"); ผ่านรายการแล้วห้ามลบ |
| ยกเลิกเอกสารไม่ต้องระบุเหตุผล | `AP/MyDlgBillCancel.cpp:51-85` | BC บังคับเหตุผล (audit) | คงของ BC |
| เลขที่เอกสารด้วย `SELECT MAX` | `OC/BRunningBttn.cpp:383-433` | PostgreSQL only + Rule 20 (ledger ต้อง idempotent/auditable) | ใช้ตารางตัวนับ + transaction (ข้อ 6) |
| คำนวณเงินด้วย `double` | `AP/ChmFrmGoodsPurchase.cpp:3577-3591` | Rule 20 ห้าม float; ประมวลผลที่ backend | ไม่ลอก |
| ป้ายปุ่มแถบเครื่องมือเป็นอังกฤษ | `FW/mybasedbformview.cpp:2246` | กฎ Thai-First / ผู้ใช้ 40+ | ใช้ป้ายไทย (ให้ lead ยืนยัน) |
| ส่วนที่นอกขอบเขต GL ตอนนี้ | F7 บันทึกจ่ายเงิน, F8 หักมัดจำ, Shift+F9 Lot, Ctrl+F9 Serial, Ctrl+F7/F8 คลัง, พักบิล F8/F9 | ขอบเขตงาน GL-only ปัจจุบัน | ทำเป็นกรอบกลางไว้ (ผูกคีย์ได้เมื่อมีจอ ERP) ยังไม่ทำจอเหล่านั้น |
| ระบบเงินเดือน | ไม่เกี่ยวกับกรอบกลาง | ห้ามทำ | — |
| เข้าระบบ/เลือกบริษัท | หัวข้อ 14 | คำสั่งเจ้าของ 2026-10-07 | ไม่ปรับ |

### ผลต่อโครงสร้างข้อมูล (ข้อมูลเก่าทิ้งได้ — ไม่ต้อง migration)

เพิ่มได้ตรง ๆ ในสคีมาที่ backend ถือ (`CREATE ... IF NOT EXISTS`): รูปแบบเลขที่ + ตัวนับต่อสมุดรายวัน/ประเภทเอกสาร (ข้อ 6), `print_log` + `print_limit` (ข้อ 13), ธงสิทธิ์รายจอ (ข้อ 19) — ชื่อ/คอลัมน์ lowercase `snake_case`, วันที่เวลาเก็บ UTC, ไม่ใช้ float

---

## 16. สิ่งที่ยังไม่ verify (อย่าถือเป็นข้อเท็จจริง)

| เรื่อง | เหตุผล |
|---|---|
| ข้อความไทยจริงของคีย์ใน `myresource` (เช่น `CANNOT_OPEN`, `CONFIRM_DELETEDATA`, `NOT_SAVE_ISCONFIRM`, `POSTGL_WARN`, `DATE_CLOSE_PERIOD`) | ข้อความอยู่ในตาราง DB ที่ไม่มีในรีโป — อ้างได้เฉพาะคีย์ และข้อความไทยที่ปรากฏในคอมเมนต์โค้ด (`CANNOT_OPEN` = "คุณไม่มีสิทธิเข้าหน้าจอนี้") |
| แถบเครื่องมือไหนแสดงจริงตอนรัน (rebar หลัก 4 แถบ vs แถบในฟอร์ม) | อ่านจากโค้ดสร้างแถบ ไม่ได้รันโปรแกรม (`FW/BCSWinApp.cpp:1921-2072`) |
| ความหมาย `ID_HK_CTRL_G`, `ID_HK_CANCEL_ALERT`, `ID_HK_FIND_F4` ในแต่ละจอ | ไม่ได้ไล่ทุกฟอร์ม (ตรวจจริงเฉพาะฟอร์มซื้อสินค้าและเสนอราคา) |
| `BCOption1` = `bCanPrintReport` | อนุมานจากลำดับบรรทัด (ส่วนหัวของ if ถูกตัดตอนอ่าน) |
| ปุ่ม "Recur" รายละเอียดเมนู drop-down | อ่านเฉพาะทางเข้า (`FW/mybasedbformview.cpp:2412`, `OC/BaseDBFormView.cpp OnBaseRecur`) |
| ชนิด "เลือกหลายรายการ" (`MultiSel`) | ไม่ได้อ่าน `BMergeBttn` ส่วนนี้ละเอียด |
| เบราว์เซอร์ดักคีย์ Ctrl+N/Ctrl+W/F5/F11/F12 ได้หรือไม่ | ต้องทดสอบในเบราว์เซอร์จริง (ข้อ 4) |
| ปัจจุบัน Enter บน `AccountSelect` ในตารางเปิด dialog และเลื่อนโฟกัสพร้อมกันหรือไม่ | ต้องทดสอบจริง (ข้อ 2) |
| การเก็บสิทธิ์รายจอฝั่ง BC ที่ UI ใช้ปิด/เปิดปุ่ม | ไม่พบใน `gl-journals.tsx` (ข้อ 19) แต่ไม่ได้ไล่ทุกจอ |
| รายการทางลัดเริ่มต้นของ BC, หัวแท็บแสดงบริษัท/สาขา | ไม่ได้ตรวจ |
