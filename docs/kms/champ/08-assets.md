# Champ -> BC : ระบบสินทรัพย์ถาวร (AS) — พิมพ์เขียวหน้าจอ การคิดค่าเสื่อม การโอนเข้า GL และรายงาน

> ที่มา: ลุงจืดสั่ง 2026-10-07 ให้ออกแบบ BC ใหม่โดยยึด Champ เป็นหลัก ให้เหมือนกันแทบทุกอย่าง (หน้าจอ ลำดับฟิลด์ ข้อความไทย กติกาสถานะ คีย์ลัด รายงาน)
> ขอบเขต: เมนู AS ของ Champ 28 รายการ (menuconfig.xml, atlas `scratch/champ-atlas/atlas-08-AS.md`) — จอบันทึก 3 + ตั้งประเภท 1 + กระบวนการ 2 + รายงาน 22
> วิธีตรวจ: อ่านซอร์สอย่างเดียว ไม่ได้รัน Champ และไม่ได้ query SQL Server ของ Champ. ข้อความไทยของ Champ ที่เรียกผ่าน `GetMyRes("KEY")` เก็บในตาราง SQL `myresource` ซึ่งไม่อยู่ใน repo — ข้อความที่ไม่พบใน .rc / ตาราง keyThai ของ atlas ผมระบุเป็น "(อนุมาน)" จากคอมเมนต์ในโค้ด ห้ามใช้เป็นคำศัพท์จริงจนกว่าลุงจืดยืนยัน
> สถานะ BC ตรวจ 2026-10-07 จาก working tree `D:\bccode` (read-only)
> ข้อควรระวังนอกขอบเขต: AGENTS.md ระบุว่าตอนนี้เน้น GL ตัวเดียว (ระบบอื่นยังไม่ต้องทำ) — สินทรัพย์ถาวรเป็นโมดูลที่ BC ทำไปแล้วก่อนหน้า (ADR `docs/kms/decisions/2026-09-15-fixed-assets-and-depreciation-engine.md`) เอกสารนี้จึงเป็น "เป้าหมายเมื่อถึงคิวสินทรัพย์" ไม่ใช่คำสั่งให้ทำทันที — ลุงจืดตัดสินลำดับงาน

ชื่อย่อ path

| ย่อ | path จริง |
|---|---|
| `C\` | `D:\project-champ\champ\champ\` (จอ/ไดอะล็อกของโปรแกรมหลัก) |
| `RG\` | `D:\project-champ\champ-report\BC5REP_GLFS\` (รายงานสินทรัพย์ + เอนจินค่าเสื่อมฝั่งรายงาน) |
| `LF\` | `D:\project-champ\champ-lib\BC5FRMWRK\` |
| `LR\` | `D:\project-champ\champ-lib\BOLEDBREPORT\` |
| `be\` | `D:\bccode\backend\internal\fixedasset\` |
| `fe\` | `D:\bccode\frontend\src\` |

เลขบรรทัดของ Champ อ้างจากไฟล์ที่ decode TIS-620/UTF-16 แล้ว (จำนวนบรรทัดเท่าต้นฉบับ)

---

## 0. สรุปความต่างสำคัญที่ BC ต้องปรับให้เหมือน Champ (เรียงตามผลกระทบ)

| # | หัวข้อ | Champ | BC ตอนนี้ | หลักฐาน |
|---|---|---|---|---|
| 1 | จอและเมนู | 6 จอ/ไดอะล็อก + 22 รายงาน = 28 เมนู แยกจอชัดเจน | หน้าเดียว 5 แท็บ (ทะเบียน/คิดค่าเสื่อม/โอน GL/ตาราง/ภาษี); เมนู 11 รายการ; `/asset/cip`, `/asset/maintenance` ไม่มีจอ (ตกไปแท็บทะเบียนเงียบ ๆ แต่ถูกนับว่า "พร้อม"); `/asset/types` ไม่มีปุ่ม/แท็บ | `C\FAFrmAssetsMaster.cpp:1`; `fe\app\asset\fixed-assets-screen.tsx:113-117,350-380`; `fe\lib\menu-screen-status.ts:19-28`; `fe\lib\menu-data.ts:558-592` |
| 2 | ฟิลด์ทะเบียนสินทรัพย์ | ~38 ฟิลด์ลำดับตายตัว: หน่วยนับ, เอกสารซื้อ(เลขที่/ลงวันที่/ซื้อจาก), กลุ่มขาย (วันที่ขาย/ขายให้/เลขที่/ราคาขาย/มูลค่าคงเหลือ), กลุ่มประกัน (6 ฟิลด์), การดูแลรักษา, รูปภาพ, วิธีคิด 3 แบบ, ค่าเสื่อมสะสมยกมา, ราคาซาก default 1 | ~20 ฟิลด์ (ไม่มีหน่วยนับ/เอกสารซื้อ/ประกัน/รูป/ซื้อจาก/ขายให้) + ฟิลด์ที่ Champ ไม่มี (serial/brand/model/supplier/PassengerCarTaxCap/IsTaxDeductible) | `be\models.go:46-82` |
| 3 | วิธีคิดค่าเสื่อม | 3 วิธี (เส้นตรง/ผลรวมจำนวนปี/ลดน้อยถอยลง) + โหมด "คำนวณเป็นรายวัน" vs รายปี (true-up งวดสุดท้าย) + งวดจริงจาก `BCPeriod` + ตัดตามวันขาย | เส้นตรงอย่างเดียว ตามเดือนปฏิทิน; ฟิลด์ `Method` มีแต่ไม่ถูกใช้ | `C\Depreciation.cpp:144-166,299-1037`; `be\models.go:69` เทียบ `be\calculator.go:34-157` |
| 4 | กระบวนการคิดค่าเสื่อม | "ประมวลผลสินทรัพย์" = เลือกช่วงรหัส + ติ๊กรายวัน → ลบแล้วสร้าง `BCAssetsOfYear/OfPeriod` ใหม่ทั้งช่วง (ไม่ลง GL) | คำนวณตารางรายสินทรัพย์ (แท็บคิดค่าเสื่อม) ไม่มีงานประมวลผลเป็นชุด | `C\FAFrmAssetsProcessDlg.cpp:1360-1595` |
| 5 | โอนเข้า GL | ไดอะล็อก: สมุดรายวัน + เลขที่ + วันที่ + ช่วงสินทรัพย์/ประเภท/แผนก + ช่วงงวด + ปี + ติ๊กแยกแผนก → ตารางตัวอย่าง (แก้ตัวเลขได้) → บันทึกเป็นรายการ Source = -2 รวมตามประเภท | ผ่านค่าเสื่อมทีละงวดตามปี/งวดเดียว (`PostDepreciation(fiscalYear, period, ...)`) | `C\FSTransferToGLDlg.cpp:74-629`; `be\gl_poster.go:239` |
| 6 | ประเภทสินทรัพย์ | ตาราง 4 คอลัมน์ (รหัสประเภท, ชื่อประเภท, เดบิต, เครดิต) เป็นไดอะล็อกกริด | 3 บัญชี (สินทรัพย์/สะสม/ค่าใช้จ่าย) + อายุ/อัตราตั้งต้น; มี API สร้างแต่ไม่มีจอ | `C\FAAssetsTypeDlg.cpp:49-52`; `be\models.go:31-41`; `be\httpapi\http.go:60-69` |
| 7 | รายงาน | 22 รายงาน 4 กลุ่ม (รายละเอียด 6 / รายปี 6 / รายเดือน 6 / อื่น 4) ผ่านกรอบ filter dialog มาตรฐาน | 1 ตารางค่าเสื่อม + 1 กระทบยอดภาษี; 4 เมนูรายงานในเมนู (รายเดือน/รายปี/ภ.ง.ด.50/ขาย) ยังไม่มี API | `fe\lib\menu-data.ts:582-590` |
| 8 | การขาย/ตัดจำหน่าย | บันทึกในทะเบียนอย่างเดียว (วันที่ขาย/ราคาขาย/มูลค่าคงเหลือ ณ วันขาย) ไม่ลง GL | BC มีจอ/คำสั่งตัดจำหน่ายที่ลง GL อัตโนมัติ (มีกำไร/ขาดทุน) | `C\FAFrmAssetsMaster.cpp:414-487`; `be\gl_poster.go:562` |
| 9 | เครื่องยนต์คำนวณ | Champ มีสองสำเนาไม่เท่ากัน (ฝั่งจอ vs ฝั่งรายงาน) — ต้องเลือกฐานเดียวก่อนลอก | decimal (`shopspring`) เหมาะ Rule 20 | `C\Depreciation.cpp`(1091 บรรทัด) เทียบ `RG\Depreciation.cpp`(1674 บรรทัด) |
| 10 | ภ.ง.ด.50 | "รายงานค่าเสื่อมตาม ภ.ง.ด.50" = รายงานค่าเสื่อมรายปีจัดกลุ่มตามประเภทเท่านั้น (ไม่มีส่วนกระทบยอดภาษี/บวกกลับ) | มีรายงานกระทบยอดภาษี (เกินกว่า Champ) | `RG\FAsRepAssetsOrderP50View.cpp:435-448` (เทียบกับ OrderType) |

หมายเหตุ: ข้อ 8 และ 10 BC "มากกว่า Champ" — ตามกฎ Strict Parity (AGENTS.md) ต้องให้ลุงจืดตัดสินว่าจะคงส่วนเกินไว้หรือถอด; ผมไม่เสนอเพิ่มเอง

---

## 1. โครงเมนู AS ของ Champ (28 รายการ) และคู่ใน BC

ที่มา: `C\menuconfigxml\menuconfig.xml` (UTF-16LE) → atlas; handler ใน `C\MainFrm.cpp`. ป๊อปอัปรายงาน 4 ชุดมี `disable="yes"` (ไม่โชว์ในเมนูหลัก เปิดผ่าน Report Center — ดู `10-reports.md`)

| menu id | ชื่อเมนู (Champ) | define | handler : บรรทัดใน MainFrm.cpp | จอ/คลาสหลัก | BC route (menu-data.ts) |
|---|---|---|---|---|---|
| 5441 | รายละเอียดสินทรัพย์ | ID_FASYS_TRANS_1 | OnFasysTrans1 : 2647 | CFAFrmAssetsMaster | `/asset/registry` |
| 5442 | บันทึกรายการซ่อมบำรุง | ID_FASYS_TRANS_2 | OnFasysTrans2 : 2668 | CFAFrmAssetsRenovate | `/asset/maintenance` |
| 5443 | บันทึกสินทรัพย์ระหว่างก่อสร้าง | ID_FASYS_TRANS_3 | OnFasysTrans3 : 2678 | CFAFrmAssetsBuild | `/asset/cip` |
| 5873 | กำหนดประเภทสินทรัพย์ | ID_SYS_FATYPE | OnSysFaType : 3112 | CFAAssetsTypeDlg (DoModal) | `/asset/types` |
| 5444 | ประมวลผลสินทรัพย์ | ID_FASYS_TRANS_6 | OnFAPROCESS : 3005 | CFAFrmAssetsProcessDlg (DoModal) | `/asset/depreciation` |
| 5491 | โอนข้อมูลเข้าสู่ GL | ID_FS_TRANSFERGL | OnFATRANSFERGL : 2996 | CFSTransferToGLDlg (DoModal) | `/asset/post-gl` |
| 5424 | รายงานรายละเอียดสินทรัพย์-ตามรหัส | ID_FASYS_MASTERREP_1 | OnFasysMasterrep1 : 7967 | CFARepAssetsItemByItemView | ไม่มี |
| 5425 | รายงานรายละเอียดสินทรัพย์-ตามประเภท | ID_FASYS_MASTERREP_2 | OnFasysMasterrep2 : 7978 | CFARepAssetsItemByTypeView | ไม่มี |
| 5426 | รายงานรายละเอียดสินทรัพย์-ตามแผนก | ID_FASYS_MASTERREP_3 | OnFasysMasterrep3 : 7990 | CFARepAssetsItemByDepartView | ไม่มี |
| 5427 | รายงานรายละเอียดสินทรัพย์-ตามที่ตั้ง | ID_FASYS_MASTERREP_4 | OnFasysMasterrep4 : 8003 | CFARepAssetsItemByPlaceView | ไม่มี |
| 5428 | รายงานรายละเอียดสินทรัพย์-ตามวันที่ซื้อ | ID_FASYS_MASTERREP_5 | OnFasysMasterrep5 : 8014 | CFARepAssetsItemByDateView | ไม่มี |
| 5429 | รายงานรายละเอียดสินทรัพย์-วันที่คำนวณ | ID_FASYS_MASTERREP_6 | OnFasysMasterrep6 : 8025 | CFARepAssetsItemByProcessView | ไม่มี |
| 5445 | รายงานค่าเสื่อมราคาเป็นปี-แผนก,ประเภท | ID_FASYS_YEARCALCREP_1 | OnFasysYearcalcrep1 : 8036 | CFAsRepAssetsOrderDepartView | `/report/depreciationyearly` (รวมทั้งกลุ่ม) |
| 5446 | รายงานค่าเสื่อมราคาเป็นปี-ประเภท,แผนก | ID_FASYS_YEARCALCREP_2 | OnFasysYearcalcrep2 : 8048 | CFAsRepAssetsOrderTypeView | (รวมข้างบน) |
| 5447 | รายงานค่าเสื่อมราคาเป็นปี-ตามที่ตั้ง | ID_FASYS_YEARCALCREP_3 | OnFasysYearcalcrep3 : 8059 | CFAsRepAssetsOrderLocateView | (รวมข้างบน) |
| 5448 | รายงานค่าเสื่อมราคาเป็นปี-ตามรหัส | ID_FASYS_YEARCALCREP_4 | OnFasysYearcalcrep4 : 8070 | CFAsRepAssetsOrderCodeView | (รวมข้างบน) |
| 5449 | รายงานสรุปค่าเสื่อมทั้งหมดเป็นปี | ID_FASYS_YEARCALCREP_6 | OnFasysYearcalcrep6 : 8081 | CFAsRepAssetsSumOfYearView | `/report/assetschedule` (ใกล้เคียง) |
| 5450 | รายงานค่าเสื่อมตาม ภ.ง.ด.50 | ID_FASYS_YEARCALCREP_7 | OnFasysYearcalcrep7 : 8092 | CFAsRepAssetsOrderP50View | `/report/depreciationpnd50` |
| 5430 | รายงานค่าเสื่อมราคาเดือน-แผนก,ประเภท | ID_FASYS_MONTHCALCREP_1 | OnFasysMonthcalcrep1 : 8103 | CFARepDeperciateDepartView | `/report/depreciationmonthly` (รวมทั้งกลุ่ม) |
| 5431 | รายงานค่าเสื่อมราคาเดือน-ประเภท,แผนก | ID_FASYS_MONTHCALCREP_2 | OnFasysMonthcalcrep2 : 8114 | CFARepDeperciateTypeView | (รวมข้างบน) |
| 5432 | รายงานค่าเสื่อมราคาเดือน-ตามที่ตั้ง | ID_FASYS_MONTHCALCREP_3 | OnFasysMonthcalcrep3 : 8125 | CFARepDeperciateLocateView | (รวมข้างบน) |
| 5433 | รายงานค่าเสื่อมราคาเดือน-ตามรหัส | ID_FASYS_MONTHCALCREP_4 | OnFasysMonthcalcrep4 : 8136 | CFARepDeperciateAssetsView | (รวมข้างบน) |
| 5434 | รายงานค่าเสื่อมราคาเดือน-วันเริ่มคำนวณ | ID_FASYS_MONTHCALCREP_5 | OnFasysMonthcalcrep5 : 8147 | CFARepDeperciateDocDateView | (รวมข้างบน) |
| 5435 | รายงานสรุปค่าเสื่อมทั้งหมดเป็นเดือน | ID_FASYS_MONTHCALCREP_6 | OnFasysMonthcalcrep6 : 8158 | CFARepDeperciateAbstractView | (รวมข้างบน) |
| 5436 | รายงานสินทรัพย์ที่ถูกขายพร้อมกำไรขั้นต้น | ID_FASYS_OTHERREP_2 | OnFasysOtherrep2 : 8169 | CFARepSellItemView | `/report/assetdisposal` |
| 5437 | รายงานสินทรัพย์ระหว่างก่อสร้าง | ID_FASYS_OTHERREP_3 | OnFasysOtherrep3 : 8179 | CFARepAssetsBuildView | ไม่มี |
| 5438 | รายงานการซ่อมบำรุงสินทรัพย์ | ID_FASYS_OTHERREP_4 | OnFasysOtherrep4 : 8190 | CFARepAssetsRenovateView | ไม่มี |
| 5439 | รายงานข้อมูลก่อน post บัญชี | ID_FASYS_OTHERREP_5 | **ไม่มี handler** (atlas: `→ ?`) | — | ไม่มี |

ลำดับเมนูใน BC (`fe\lib\menu-data.ts:564-572`) ตรงกับ Champ 6 รายการแรกอยู่แล้ว (registry, maintenance, cip, types, depreciation) แต่ BC แยก "โอนข้อมูลเข้าสู่ GL" ไปกลุ่ม `fa-posting` (`:575-579`) ขณะที่ Champ อยู่ลำดับที่ 6 ในกลุ่มเดียวกัน

---

## 2. ตาราง SQL ของระบบสินทรัพย์ (DDL จาก `C\Script\SQLSERVER_Script.sql` ผ่าน atlas.json)

ชนิด `MONEY` ของ SQL Server = ทศนิยม 4 ตำแหน่ง; ฟิลด์ที่แสดง/ปัดบนจอใช้ `theApp.m_CompanyInfo.AmountPoint` (จำนวนทศนิยมตามที่บริษัทตั้ง) — BC ต้องใช้ decimal/NUMERIC ตาม Rule 20 ห้าม float

| ตาราง | คอลัมน์ (ตามลำดับใน DDL) |
|---|---|
| `BCAssetsMaster` | ROWORDER, iscancel, isconfirm, confirmcode, confirmdatetime, cancelcode, canceldatetime, **Code** VARCHAR(25), **Name** VARCHAR(255), UnitCode VARCHAR(10), **AssetsType** VARCHAR(10), IsExport, CreatorCode, CreateDateTime, LastEditorCode, LastEditDateT, DepartCode, LocateCode, PictureFileName VARCHAR(255), **BuyDate**, BuyPrice MONEY, BuyDocNo VARCHAR(20), BuyDocDate, BuyForm VARCHAR(255), **StartCalcDate**, AmountCalc MONEY, **CalcType** SMALLINT, AssetsRate MONEY, AssetsAge SMALLINT, StopCalcDate, AssetsBalance MONEY, SumOfAssetsBal MONEY, SaleDate, SalePrice MONEY, SaleAssetsBal MONEY, SaleDocNo VARCHAR(20), SaleDocDate, CustName VARCHAR(255), InsuranceNo VARCHAR(20), SaftyNo VARCHAR(20), SaftyName VARCHAR(255), InsurePremium MONEY, InsureAge SMALLINT, InsureStartDate, InsureStopDate, MyDescription VARCHAR(255), Remain MONEY, PrintCount, FirstYearRate MONEY, IsFirstScrap SMALLINT, ProcessType SMALLINT |
| `BCAssetsType` | ROWORDER, Code VARCHAR(10), IsExport, CreatorCode, CreateDateTime, LastEditorCode, LastEditDateT, Name VARCHAR(50), **DrAccountCode** VARCHAR(20), **CrAccountCode** VARCHAR(20) |
| `BCAssetsLocate` | ROWORDER, Code VARCHAR(10), (audit cols), Name VARCHAR(50), UserGroup VARCHAR(255) |
| `BCAssetsOfYear` | ROWORDER, AssetsCode VARCHAR(25), (audit), StartCalcDate, StopCalcDate, DepartCode, AssetsType, LocateCode, DayCalc SMALLINT, BeginBalance MONEY, AssetsAmount MONEY, AssetsAmass MONEY, EndBalance MONEY |
| `BCAssetsOfPeriod` | ROWORDER, AssetsCode, (audit), StartCalcDate, StopCalcDate, DepartCode, AssetsType, LocateCode, BeginBalance, AssetsAmount, **AssetsPeriod1..AssetsPeriod13** MONEY, EndBalance |
| `BCAssetsOfBuild` (หัวสินทรัพย์ระหว่างก่อสร้าง) | ROWORDER, iscancel, isconfirm, confirm*/cancel* cols, AssetsCode VARCHAR(25), (audit), MyDescription VARCHAR(255), Amount MONEY, PrintCount |
| `BCAssetsOfBldSub` (รายการย่อย) | ROWORDER, LineNumber, (audit), AssetsCode, DocDate, DocNo VARCHAR(20), Desc1 VARCHAR(255), Desc2 VARCHAR(255), Amount MONEY |
| `BCRenovate` (หัวซ่อมบำรุง) | ROWORDER, iscancel, isconfirm, confirm*/cancel* cols, AssetsCode, Amount MONEY, (audit), PrintCount |
| `BCRenovateSub` | ROWORDER, LineNumber, AssetsCode, DocDate, (audit), DocNo, Desc1, Desc2, Amount MONEY |
| `BCPeriod` | ROWORDER, PeriodNo SMALLINT, StartDate, (audit), StopDate, ItemRemain, NetProfitLoss, SummProfitLoss, Status SMALLINT |
| `BCDepartment`, `BCItemUnit`, `BCGLBook`, `BCChartOfAccount` | ตารางอ้างอิง (lookup) |
| `BCTrans` / `BCTransSub` | รายการบัญชีที่โอนจากสินทรัพย์ (`Source = -2`) — โครงเต็มดู `09-gl.md` §6 |

ข้อสังเกต: ทั้ง `BCAssetsMaster/BCAssetsOfBuild/BCRenovate` มีคอลัมน์ iscancel/isconfirm/confirm*/cancel* แต่จอสินทรัพย์ทั้งสามไม่ใช้ (ไม่มีขั้นตอนยืนยัน/ยกเลิกเอกสาร — ดู §3.5)

---

## 3. จอ "รายละเอียดสินทรัพย์" (menu 5441, `CFAFrmAssetsMaster`)

ที่มา: `C\FAFrmAssetsMaster.cpp` (1150 บรรทัด), เลย์เอาต์ `C\BC5Account.rc:15429` (`IDD_FA_FRM_ASSETS_MASTER`, 441x273 dialog units, ฟอร์มฝังในหน้าต่างหลักแบบ MDI)

### 3.1 วัตถุประสงค์และหน้าตา

| หัวข้อ | รายละเอียด | หลักฐาน |
|---|---|---|
| ชื่อจอ | "รายการสินทรัพย์" (key 90001 ใช้ในชื่อหัวและ log) | `FAFrmAssetsMaster.cpp:31` |
| เจตนา | ทะเบียนสินทรัพย์ถาวร 1 ใบ = 1 แถวใน `BCAssetsMaster` เก็บข้อมูลซื้อ, พารามิเตอร์ค่าเสื่อม, ข้อมูลขาย, ประกัน/ดูแลรักษา, รูปภาพ | `:418-462` |
| รูปแบบ | ฟอร์มเดี่ยว (ไม่มีตารางรายการย่อย) แบ่ง 4 กลุ่ม: หัวเอกสาร (ซ้ายบน) + รูปภาพ (ขวาบน) / การคิดค่าเสื่อมราคา (ซ้ายล่าง) / การขายสินทรัพย์ (กลางขวา) / การดูแลรักษา+ประกัน (ขวาล่าง) | `rc:15429-15600` |
| แถบบนสุด | ช่อง "ตัวเลือกหลังบันทึก" (combo 2 ค่า) ที่ x=67,y=3 | `FAFrmAssetsMaster.cpp:146-150` |
| แถบล่าง | Toolbar มาตรฐานของ `CMyEspeciallyFormView` (ใหม่/บันทึก/ค้นหา/ลบ/พิมพ์) + ปุ่ม `IDC_MORE` เปิดเมนูป๊อปอัป | `:986-1006` (OnMore) |
| สถานะเอกสาร | **ไม่มี** — ไม่มี confirm/cancel/post (ทั้งที่ตารางมีคอลัมน์ iscancel/isconfirm) | ไม่พบการใช้ใน `FAFrmAssetsMaster.cpp` |
| สิทธิ์ | `SetUserAuthorization(..., ID_FASYS_TRANS_1)`; รหัสสินทรัพย์แก้ได้เฉพาะผู้มีสิทธิ์ `bCanChangeBillNo` | `:38`, `:272-274` |

ตัวเลือกหลังบันทึก (combo `IDC_ED_SAVEOPTION`):

| ลำดับ | ข้อความ (key) | พฤติกรรมหลังบันทึก (`OnClearData`) | หลักฐาน |
|---|---|---|---|
| 0 (ค่าเริ่มต้น) | "บันทึกแล้วไม่ล้างหน้าจอ" (`SAVE_NOTCLEAR_SCREEN`) | ล้างเฉพาะ "รหัสสินทรัพย์" และตั้งราคาซาก = 1 — ฟิลด์อื่นคงค่าไว้เพื่อคีย์สินทรัพย์ชุดคล้ายกันต่อ | `:579-590` |
| 1 | "บันทึกแล้วเพิ่มรายการใหม่" (`SAVE_NEW_REC`) | ล้างทุกฟิลด์ ตั้งวันที่ซื้อ/ลงวันที่อ้างอิง/เริ่มคิดค่าเสื่อม = วันที่ทำงาน (`m_strWorkingDate`) ถ้ามี, CalcType=0, หักค่าซากก่อน=Yes, ราคาซาก=1 | `:592-651` |

ข้อความไทยของสองตัวเลือก: จากคอมเมนต์ในโค้ด (`:146-147`), ไม่มีใน .rc (อนุมาน)

### 3.2 ฟิลด์ตามลำดับบนจอ (ซ้ายไปขวา บนลงล่าง) — ข้อความไทยตรงจาก `.rc` (`KEY$ไทย`)

#### กลุ่มหัวเอกสาร (กรอบ `IDC_LB_GROUP_LIST` 4,12,329x87)

| # | ป้ายไทย (key) | ชนิด control | บังคับ | ค่าเริ่มต้น | Lookup / รูปแบบ | คอลัมน์ `BCAssetsMaster` | หลักฐาน |
|---|---|---|---|---|---|---|---|
| 1 | รหัสสินทรัพย์ (`ASSETS_CODE`) | `BRunningEdit` ตัวพิมพ์ใหญ่ + ปุ่มรัน | **บังคับ** (`m_bMustEntry`) | รันเลขใหม่ทุกครั้งที่กด "ใหม่" (`RunningAgain`) ตามรูปแบบเลขที่ของ `BCAssetsMaster.Code` | ห้ามซ้ำ (`SetDuplicate`); ปิดการแก้ไขถ้าไม่มีสิทธิ์เปลี่ยนเลขที่ | `Code` VARCHAR(25) | `.cpp:133-134,272-274,560-561,777-785`; rc:15434 |
| 2 | หน่วยนับ (`UINTCODE`) | combo จากตาราง | ไม่ | ว่าง | `SELECT Code,Name FROM BCItemUnit` | `UnitCode` VARCHAR(10) | `.cpp:220`; rc:15435 |
| 3 | ประเภท (`ITEM_TYPE`) | combo จากตาราง | **บังคับ** (`ITEM_TYPE_NOTNULL`) | ว่าง | `SELECT * FROM BCAssetsType` | `AssetsType` VARCHAR(10) | `.cpp:221-222,404-409`; rc:15437 |
| 4 | แผนก (`DEPART`) | combo จากตาราง | ไม่ | ว่าง | `SELECT * FROM BCDepartment` | `DepartCode` | `.cpp:224` |
| 5 | ที่ตั้ง (`FA_LOCATE`) | combo จากตาราง | ไม่ | ว่าง | `SELECT * FROM BCAssetsLocate` | `LocateCode` | `.cpp:225` |
| 6 | ชื่อสินทรัพย์ (`ASSETS_NAME`) | text กว้าง | **ไม่บังคับ** (ไม่มีการตรวจ) | ว่าง | — | `Name` VARCHAR(255) | `.cpp:419`; rc:15441 |
| 7 | วันที่ซื้อหรือได้มา (`FA_BUYDATE`) | วันที่ (พ.ศ./ค.ศ. ตามปีบริษัท `YearType`) | ไม่ | วันที่ทำงาน | — | `BuyDate` | `.cpp:172-176,427` |
| 8 | ราคาที่ซื้อหรือได้มา (`FA_BUYPRICE`) | ตัวเลข ชิดขวา ทศนิยม=`AmountPoint` | ไม่ | ว่าง | — | `BuyPrice` MONEY | `.cpp:211,428` |
| 9 | เอกสารอ้างอิง (`REFERANCE_NO`) | text ตัวพิมพ์ใหญ่ | ไม่ | ว่าง | — | `BuyDocNo` VARCHAR(20) | `.cpp:429` |
| 10 | ลงวันที่อ้างอิง (`FA_BUYDOCDATE`) | วันที่ | ไม่ | วันที่ทำงาน | — | `BuyDocDate` | `.cpp:179-183,431` |
| 11 | ซื้อสินทรัพย์จาก (`FA_BUYFORM`) | text หลายบรรทัด | ไม่ | ว่าง | — | `BuyForm` VARCHAR(255) | `.cpp:432` |
| 12 | รูปภาพ | กรอบรูป 99x101 (`CBImageWnd`) + ช่องซ่อน `IDC_ED_PIC` เก็บชื่อไฟล์ | ไม่ | ว่าง | เลือกไฟล์รูปจากโฟลเดอร์รูป (`theApp.GetPictureFolder()`) | `PictureFileName` VARCHAR(255) (เก็บแค่ชื่อไฟล์) | `.cpp:236-245,700-708` |

#### กลุ่ม "การคิดค่าเสื่อมราคา" (`FA_CALE_DEPRECIATE`)

| # | ป้ายไทย (key) | ชนิด | บังคับ | ค่าเริ่มต้น | Lookup / กติกา | คอลัมน์ | หลักฐาน |
|---|---|---|---|---|---|---|---|
| 13 | เริ่มคิดค่าเสื่อม (`FA_START_CALCDATE`) | วันที่ | **บังคับ** (`m_bMustEntry` `.cpp:191`) | วันที่ทำงาน | ปิดการแก้ไขในโหมด Chm (`m_bChmMode`) | `StartCalcDate` | `.cpp:184-191,434,554-555` |
| 14 | อายุการใช้งาน (`FA_ASSETS_AGE`) | ตัวเลข ชิดขวา | ไม่ตรวจ | ว่าง | เมื่อออกจากช่อง (KillFocus) → คำนวณ "วันที่หยุดคำนวณ" และ "อัตรา" อัตโนมัติ (ดู 3.4) | `AssetsAge` SMALLINT | `.cpp:98,206,441` |
| 15 | อัตราค่าเสื่อมราคา (`FA_ASSETS_RATE`) | ตัวเลข % | ไม่ตรวจ | ว่าง (เติมอัตโนมัติ = 100/อายุ) | แก้ได้ | `AssetsRate` MONEY | `.cpp:208,438` |
| 16 | หักค่าเสื่อมเบื้องต้นปีแรก(%) (`FA_FIRSTYEAR_RATE`) | ตัวเลข % | ไม่ | ว่าง | 0<x≤100 จึงมีผลใน engine | `FirstYearRate` MONEY | `.cpp:209,439` |
| 17 | วันที่หยุดคำนวณ (`FA_STOP_CALCDATE`) | วันที่ | ไม่ตรวจ | ว่าง (เติมอัตโนมัติ) | = วันที่ซื้อ + อายุ(ปี) - 1 วัน | `StopCalcDate` | `.cpp:193,442` |
| 18 | ราคาที่คิดค่าเสื่อม (`FA_AMOUNT_CALC`) | ตัวเลข | ไม่ตรวจ | ว่าง | ฐานคิดค่าเสื่อม (ไม่ผูกกับราคาซื้ออัตโนมัติ) | `AmountCalc` MONEY | `.cpp:212,435` |
| 19 | มูลค่าคงเหลือยกมา (`FA_ASSETS_BALANCE`) | ตัวเลข | ไม่ | ว่าง | ถ้า >0 และ "วันที่ซื้อ" ≠ "เริ่มคิดค่าเสื่อม" → engine ใช้เป็นฐาน (สินทรัพย์ยกมาจากระบบเก่า) | `AssetsBalance` MONEY | `.cpp:213,443`; `Depreciation.cpp:122-125,314-318` |
| 20 | ค่าเสื่อมสะสมยกมา (`FA_ASSETS_SUM_BAL`) | ตัวเลข | ไม่ | ว่าง | เก็บอย่างเดียว (engine ไม่ได้อ่านค่านี้ — ดูข้อสังเกต Q4) | `SumOfAssetsBal` MONEY | `.cpp:218,445`; rc:15539 |
| 21 | ราคาซาก (`FA_REMAIN_PRICE`) | ตัวเลข | ไม่ | **1** (ฮาร์ดโค้ด คอมเมนต์ว่า "ปูนไทยให้ราคาซากเป็น 1") | — | `Remain` MONEY | `.cpp:214,462,589,629` |
| 22 | วิธีคิดค่าเสื่อม (`FA_CALCTYPE`) | combo 3 ค่า | ไม่ | เส้นตรง (0) | 0 เส้นตรง (`CB_STRAIGHT_LINE`), 1 ผลรวมจำนวนปี (`CB_SUM_OF_THE_YEAR`), 2 ลดน้อยถอยลง (`CB_DECREASING`) — ข้อความไทยของ 3 ตัวเลือกจากคอมเมนต์ | `CalcType` SMALLINT | `.cpp:155-161,436` |
| 23 | หักค่าซากก่อน (`FA_SCRAPT_FIRST`) | combo "No"/"Yes" (ข้อความอังกฤษตายตัว) | ไม่ | Yes (1) | ความหมายดู §7.4 (มีการกลับค่าระหว่างจอกับกระบวนการ) | `IsFirstScrap` SMALLINT | `.cpp:163-169,437,633` |

#### กลุ่ม "การขายสินทรัพย์" (`FA_SALE_FIXED_ASSETS`)

| # | ป้ายไทย (key) | ชนิด | บังคับ | หมายเหตุ | คอลัมน์ | หลักฐาน |
|---|---|---|---|---|---|---|
| 24 | วันที่ขาย (`FA_SALEDATE`) | วันที่ | ไม่ — แต่ถ้ากรอก จะบังคับ 26 และ 27 | ต้อง ≥ วันที่เริ่มคิดค่าเสื่อม (`MESS_ERROR_SALEDATE_LITTLE_CALCDATE`) | `SaleDate` | `.cpp:201,375-401,448` |
| 25 | ขายให้ (`FA_CUSTNAME`) | text | ไม่ | ชื่อผู้ซื้อแบบพิมพ์อิสระ (ไม่ผูกตารางลูกหนี้) | `CustName` VARCHAR(255) | `.cpp:453` |
| 26 | เอกสารอ้างอิง (`REFERANCE_NO`) | text | บังคับเมื่อมีวันที่ขาย (`MSG_REF_DOC_EMPTY` "กรุณากรอกเลขที่เอกสารอ้างอิง") | — | `SaleDocNo` VARCHAR(20) | `.cpp:393,450` |
| 27 | วันที่เอกสาร (`DOCDATE`) | วันที่ | บังคับเมื่อมีวันที่ขาย (`MSG_SALE_DATE_EMPTY` "กรุณากรอกวันที่เอกสารขาย") | — | `SaleDocDate` | `.cpp:399,452` |
| 28 | ราคาขาย (`PRICE`) | ตัวเลข | ไม่ | ไม่ตรวจเทียบกับมูลค่าคงเหลือ | `SalePrice` MONEY | `.cpp:216,449` |
| 29 | ค่าเสื่อมคงเหลือ (`SALEASSETSBAL`) + ปุ่ม **Calc** | ตัวเลข + ปุ่มเล็ก 17x13 | ไม่ | ปุ่ม Calc คำนวณมูลค่าคงเหลือ ณ วันขาย (ดู 3.5) — ป้ายใน rc ว่า "ค่าเสื่อมคงเหลือ" แต่เก็บ "มูลค่าตามบัญชีคงเหลือ ณ วันขาย" | `SaleAssetsBal` MONEY | `.cpp:217,444`; rc:15544 |

#### กลุ่ม "การดูแลรักษา" (`FA_DESCRIPTION` ใช้เป็นหัวกรอบกลุ่มประกัน+ดูแลรักษา)

| # | ป้ายไทย (key) | ชนิด | บังคับ | หมายเหตุ | คอลัมน์ | หลักฐาน |
|---|---|---|---|---|---|---|
| 30 | กรมธรรม์เลขที่ (`FA_INSURANCE_NO`) | text | ไม่ | — | `InsuranceNo` VARCHAR(20) | `.cpp:454` |
| 31 | เลขที่ประกัน (`FA_SAFTYNO`) | text | ไม่ | ชื่อ field สะกด "Safty" ตามต้นฉบับ | `SaftyNo` VARCHAR(20) | `.cpp:455` |
| 32 | บริษัทประกัน (`FA_SAFTY_NAME`) | text กว้าง | ไม่ | — | `SaftyName` VARCHAR(255) | `.cpp:456` |
| 33 | เบี้ยประกัน (`FA_INSURE_PREMIUM`) | ตัวเลข | ไม่ | — | `InsurePremium` MONEY | `.cpp:215,457` |
| 34 | อายุประกัน (`FA_INSURE_AGE`) | จำนวนเต็ม (ปี) | ไม่ | KillFocus → คำนวณวันสิ้นสุดประกัน | `InsureAge` SMALLINT | `.cpp:99-100,207,458` |
| 35 | วันที่เริ่มประกัน (`FA_INSURE_STARTDATE`) | วันที่ | ไม่ | KillFocus → คำนวณวันสิ้นสุดประกัน | `InsureStartDate` | `.cpp:195,459` |
| 36 | วันที่สิ้นสุดประกัน (`FA_INSURE_STOPDATE`) | วันที่ | ไม่ | = เริ่มประกัน + อายุ(ปี) - 1 วัน | `InsureStopDate` | `.cpp:198,460` |
| 37 | การดูแลรักษา (`FA_DESCRIPTION`) | text กว้าง | ไม่ | หมายเหตุอิสระ | `MyDescription` VARCHAR(255) | `.cpp:461` |

รวม 37 ฟิลด์ผู้ใช้กรอกได้ + ชื่อไฟล์รูป. ฟิลด์ในตารางที่จอ **ไม่** เขียน: `ProcessType` (เขียนโดยกระบวนการประมวลผล), `iscancel/isconfirm/confirm*/cancel*`, `IsExport`, `PrintCount` (`UpdatePrintLog` ว่างเปล่า `.cpp:1008-1011`), `CreatorCode/CreateDateTime/LastEditor*` (ถูกเติมโดย `InsertDB/UpdateDB` ของคลาสฐาน — ยังไม่ verify ฝั่งคลาสฐาน)

### 3.3 ปุ่มและคีย์ลัด

| ตัวควบคุม | การทำงาน | หลักฐาน |
|---|---|---|
| F2 | `OnAccount()` — เปิดบัญชีแยกประเภท (ของคลาสฐาน; จอนี้ไม่ได้ใช้ข้อมูลบัญชี) | `.cpp:676-681` |
| F5 | โอนข้อมูลจากสินทรัพย์ระหว่างก่อสร้าง (ใช้ได้เฉพาะโหมดเพิ่มใหม่; ไม่งั้นแจ้ง "Please click 'New Record'") | `.cpp:682-687,787-791` |
| F6 | แสดงตารางค่าเสื่อมทั้งหมดแบบข้อความ (MessageBox) จากค่าบนจอ — เครื่องมือตรวจของนักพัฒนา ข้อความเป็นอังกฤษตายตัว | `.cpp:688-693,1013-1121` |
| F11 | บันทึก (ผ่านเมนู "More" → `F11_SAVE`) | `.cpp:986-1006` |
| ID_HK_NEW / ID_HK_FIND / ID_HK_PRINT | ใหม่ / ค้นหา / พิมพ์ฟอร์ม (`OnCheckBeforePrint("BCAssetsMaster", Code)` แล้ว `OnTBPrintForm`) | `.cpp:1124-1150` |
| ปุ่ม `Calc` | คำนวณ "ค่าเสื่อมคงเหลือ" ณ วันขาย | `.cpp:872-985` |
| ปุ่ม `...` ข้างรหัส / ปุ่มรัน | เปิดรายการค้นหา / รันเลขใหม่ | `.cpp:777-785` |

### 3.4 กติกาเติมค่าอัตโนมัติ (ต้องเลียนแบบ)

| ทริกเกอร์ | ผล | สูตร | หลักฐาน |
|---|---|---|---|
| ออกจากช่อง "อายุการใช้งาน" | วันที่หยุดคำนวณ | `StopCalcDate = BuyDate + อายุ(ปี) - 1 วัน` (ใช้ **วันที่ซื้อ** `IDC_ED_DOCDATE` ไม่ใช่วันที่เริ่มคิดค่าเสื่อม) — อ่านวัน/เดือนด้วยตำแหน่งตัวอักษร dd/mm/yyyy | `.cpp:724-745` |
| เดียวกัน | อัตราค่าเสื่อม | `100 / อายุ` ปัดตาม `AmountPoint` เมื่อ 0 < อายุ ≤ 100; นอกช่วง = 0 | `.cpp:746-751` |
| ออกจากช่อง "อายุประกัน" หรือ "วันที่เริ่มประกัน" | วันที่สิ้นสุดประกัน | `InsureStart + อายุประกัน(ปี) - 1 วัน`; ไม่ทำถ้ายังไม่มีวันที่เริ่ม | `.cpp:753-775` |
| บันทึก (โหมดแก้ไข) | เปลี่ยนรหัสสินทรัพย์ได้ และ **ตามไปแก้** `BCAssetsOfYear.AssetsCode` และ `BCAssetsOfPeriod.AssetsCode` (ไม่ตามไปแก้ `BCRenovate`, `BCAssetsOfBuild`) | `UPDATE ... SET AssetsCode = <ใหม่> WHERE AssetsCode = <เดิม>` | `.cpp:474-477` |

### 3.5 ปุ่ม Calc และ F6 — คำนวณมูลค่าคงเหลือ ณ วันขาย

1. ปุ่ม Calc: ไม่ทำอะไรถ้ายังไม่มี "วันที่ขาย" (F6 ทำงานเสมอ) (`.cpp:872-876`, `:1013`)
2. ถ้า วันที่ขาย > วันที่หยุดคำนวณ → `SaleAssetsBal = ราคาซาก` แล้วจบ (`:889-894`)
3. ไม่งั้นโหลดงวดบัญชีจาก `BCPeriod WHERE PeriodNo > 0 AND PeriodNo < 13 ORDER BY PeriodNo` (`:899-915`)
4. ใส่พารามิเตอร์จากหน้าจอเข้า `CDepreciation` (ProcessType = **1 รายวัน**, `m_nFirstScrap` = ค่า combo ตรง ๆ **ไม่กลับค่า**) แล้ว `Process()` (`:922-945`)
5. `SaleAssetsBal = GetYearEnd(ปีสุดท้าย)` (ยอดคงเหลือสิ้นปีสุดท้ายของตาราง) (`:945`)
6. มีโค้ดตายหลัง `return;` (สูตรเก่า: ค่าเสื่อมต่อวัน = AssetsAmount/365 x วัน) ไม่ถูกเรียก (`:947-985`)

### 3.6 ลบ

| ขั้น | พฤติกรรม | หลักฐาน |
|---|---|---|
| ยืนยัน | รหัสว่าง = ไม่ลบเงียบ ๆ; ไม่งั้น `AfxMessageBox(DELETE_MSG, YES/NO)` | `.cpp:294-338` |
| ลบ | `DELETE BCAssetsMaster WHERE Code=...` ถ้าสำเร็จ → ลบ `BCAssetsOfYear` และ `BCAssetsOfPeriod` ของรหัสนั้น + บันทึก log (`DELETE_GL` + `90001`, action 3) | `.cpp:338-361` |
| **ไม่มี guard** | ไม่ตรวจว่าเคยโอนเข้า GL แล้ว / มีรายการซ่อมบำรุง / มีรายการก่อสร้าง — ข้อความ `CANNOT_DELETE` ถูกคอมเมนต์ทิ้ง (guard เดิม "ห้ามลบถ้ามีรายการคำนวณแล้ว" ถูกปิดโดยผู้พัฒนา 19/12/2005) | `.cpp:298-325` |

---

## 4. จอ "กำหนดประเภทสินทรัพย์" (menu 5873, `CFAAssetsTypeDlg`)

ที่มา: `C\FAAssetsTypeDlg.cpp` (163 บรรทัด) สืบทอด `CMyDlgMasterCode` — **ไม่มีซอร์สของคลาสฐานนี้ใน repo (มีแต่ `.obj`)** จึงยังไม่ verify พฤติกรรมปุ่ม/คีย์ของกริดมาสเตอร์ (เพิ่มแถว/ลบแถว/บันทึก) — ให้ดูแบบเดียวกับมาสเตอร์ตารางอื่นใน `00-common-ui.md`

| รายการ | รายละเอียด | หลักฐาน |
|---|---|---|
| ชนิดจอ | ไดอะล็อกโมดัล `DoModal` เป็นกริดแก้ไขในตัว ไม่ใช่ฟอร์มเอกสาร | `MainFrm.cpp:3112` |
| ชื่อ | "กำหนดประเภทสินทรัพย์" (key `110021`) | `FAAssetsTypeDlg.cpp:47` |
| ตาราง | `BCAssetsType` (Code, Name, DrAccountCode, CrAccountCode) | `:49-52` |

| คอลัมน์ | ป้าย | กว้าง | ชนิด | บังคับ/ตรวจ | หลักฐาน |
|---|---|---|---|---|---|
| 0 | รหัสประเภท (`TYPE_CODE`, ข้อความจากคอมเมนต์) | 75 | แก้ไขได้ | ถ้ามีชื่อแต่ไม่มีรหัส → ไม่ให้บันทึก (`NOT_TYPE`) | `:49,107-134` |
| 1 | ชื่อประเภท (`TYPE_NAME`, ข้อความจากคอมเมนต์) | 280 | แก้ไขได้ | ไม่บังคับ | `:50` |
| 2 | เดบิต (`DEBIT`) = `DrAccountCode` | 120 | ลิงก์ → เปิดรายการผังบัญชี (`GLChartOfAccountConfig`) | ต้องเป็นบัญชีที่ไม่ใช่บัญชีหัวข้อ (`ISHEADER = 0`) ไม่งั้น `NO_GL_CODE` แล้วล้างค่า | `:51,61-80,137-163` |
| 3 | เครดิต (`CREDIT`) = `CrAccountCode` | 120 | ลิงก์เช่นเดียวกัน | เช่นเดียวกัน | `:52,81-93,137-163` |

ความหมายทางบัญชี (จากการโอนเข้า GL §8): `DrAccountCode` = บัญชีค่าเสื่อมราคา (ค่าใช้จ่าย) ด้านเดบิต, `CrAccountCode` = ค่าเสื่อมราคาสะสม ด้านเครดิต — Champ **ไม่มี** บัญชีสินทรัพย์ในประเภท และไม่มีอายุ/อัตราตั้งต้นของประเภท

หมายเหตุ: คอมเมนต์ผู้พัฒนา 27-09-2003 ระบุว่าเงื่อนไข "ต้องมีบัญชีเดบิต/เครดิต" ถูกปิดไว้ "สำหรับประเภทที่ดิน" (ที่ดินไม่คิดค่าเสื่อม) → ประเภทไม่ต้องมีบัญชีได้ (`:125-130`) และตอนโอน GL ข้ามประเภทที่ไม่มีบัญชีทั้งสองฝั่ง (`FSTransferToGLDlg.cpp:538`)

---

## 5. จอ "บันทึกสินทรัพย์ระหว่างก่อสร้าง" (5443) และ "บันทึกรายการซ่อมบำรุง" (5442)

ทั้งสองเป็นฟอร์ม "หัว + ตารางรายการ" ขนาดเล็กโครงเดียวกัน (เทมเพลต `IDD_FA_FRM_BUILD_ASSETS` rc:15407 และ `IDD_FA_FRM_RENOVATE` rc:15385)

| หัวข้อ | สินทรัพย์ระหว่างก่อสร้าง `CFAFrmAssetsBuild` | ซ่อมบำรุง `CFAFrmAssetsRenovate` |
|---|---|---|
| ชื่อจอ (key) | 90003 | 90002 |
| สิทธิ์ | `ID_FASYS_TRANS_3` (`FAFrmAssetsBuild.cpp:33`) | `ID_FASYS_TRANS_2` (`FAFrmAssetsRenovate.cpp:32`) |
| ช่อง "รหัสสินทรัพย์" | text ตัวพิมพ์ใหญ่แบบ "รันเลข" ตามรูปแบบ `BCAssetsOfBuild.AssetsCode` — รหัสของงานก่อสร้างที่ตั้งเอง (ยังไม่ใช่สินทรัพย์จริง) (`:129`) | `InitialLink` — เลือกจากรายการสินทรัพย์ `FAFrmAssetsMasterConfig` พิมพ์ prefix กรองได้; ชื่อแสดงอัตโนมัติ (`:128,757-791`) — ไม่มีการบังคับว่ารหัสต้องมีใน master ตอนบันทึก (ยังไม่ verify) |
| ช่อง "ชื่อ" | `IDC_ED_NAME` แก้ไขได้ เก็บ `MyDescription` (`:274`) | อ่านอย่างเดียว ดึงจาก `BCAssetsMaster.Name` (`:783-791`) |
| ตารางรายการ | 4 คอลัมน์: วันที่ (`AS_DATE` "วันที่", กว้าง 100) / เลขที่เอกสาร (`DOCNO`, 120) / รายละเอียด (`COMMENT_DESC` "หมายเหตุ/คำอธิบาย", 200) / จำนวนเงิน (`AMOUNT`, 90, ตัวเลข ชิดขวา) (`:99-109`) | เหมือนกัน 4 คอลัมน์ (`:99-109`) |
| ท้ายจอ | "รวมทั้งสิ้น:" (`TOTAL_AMOUNT`, อ่านอย่างเดียว) = ผลรวมคอลัมน์จำนวนเงิน คำนวณใหม่เมื่อเปลี่ยนคอลัมน์/ลบแถว (`OnSumDetail`) | เหมือนกัน |
| ตารางหัว | `BCAssetsOfBuild(AssetsCode, MyDescription, Amount)` | `BCRenovate(AssetsCode, Amount)` |
| ตารางรายการ | `BCAssetsOfBldSub(AssetsCode, LineNumber, DocDate, DocNo, Desc1, Amount)` — `Desc2` ไม่ได้ใช้ | `BCRenovateSub(AssetsCode, LineNumber, DocDate, DocNo, Desc1, Amount)` |
| บันทึก | insert/update หัว แล้ว **ลบรายการเดิมทั้งหมดแล้วใส่ใหม่** เฉพาะแถวที่มี "วันที่" (`:300-340`) | ลบหัว+รายการเดิมของรหัสนั้น **ตั้งแต่ตอนตรวจก่อนบันทึก (ConfirmSaveData)** แล้ว insert ใหม่ (`:227-235`) — ถ้าตรวจไม่ผ่านหลังจากนั้นข้อมูลเดิมหาย (ดู Q8) |
| ข้อห้าม | **แก้เอกสารที่เปิดจากฐานไม่ได้ถ้ารหัสนั้นมีใน `BCAssetsMaster` แล้ว** (`MESS_ERROR_NOTSAVE_BUILD1`, ข้อความอนุมาน "โอนเป็นสินทรัพย์แล้ว") (`:213-228`) | ไม่มี |
| ลบ | ลบหัว+รายการ ไม่มีบันทึก log (`:184-203`) | ลบหัว+รายการ + บันทึก log (`DELETE_GL`+`90002`) (`:195-215`) |
| ตรวจก่อนบันทึก | รหัสว่าง → `CCANNOT_SAVE_NOASSETS` ("ไม่สามารถบันทึกได้ เนื่องจากไม่ใส่รหัสสินทรัพย์") | เหมือนกัน |
| การเชื่อม | F5 ในจอทะเบียนสินทรัพย์ดึงรายการนี้ไปสร้างสินทรัพย์ (§3.3) — ไม่มีการลบรายการก่อสร้างอัตโนมัติ และ GL **ไม่ได้ผูก** กับรายการเหล่านี้ | ไม่มีผลต่อมูลค่าสินทรัพย์/ค่าเสื่อม (ข้อมูลอ้างอิงล้วน ๆ) — `Depreciation.cpp` ไม่มีการอ่าน `BCRenovate` |

F5 ในจอทะเบียน: เมื่อเลือกรายการก่อสร้าง ระบบเติม รหัส, ชื่อ (= `MyDescription`), อายุการใช้งาน = **20 (ฮาร์ดโค้ด)**, ราคาที่คิดค่าเสื่อม = ราคาซื้อ = ยอดรวมก่อสร้าง (`FAFrmAssetsMaster.cpp:829-850`) — ไม่เรียก KillFocus ของอายุ จึงไม่เติมอัตราและวันที่หยุดคำนวณให้ (ผู้ใช้ต้องแก้ช่องอายุเพื่อกระตุ้น); เลือกหลายรายการได้แต่ค่าที่ค้างอยู่คือรายการสุดท้าย (วนเขียนทับฟิลด์เดียวกัน)

หมายเหตุสถานะ: ทั้งสองจอไม่มี workflow ยืนยัน/ยกเลิก; log ผู้ใช้ของ Renovate ใช้ menu id `ID_BILLSYS_TRANS_6` (ผิดโมดูล, `FAFrmAssetsRenovate.cpp:340`)

---

## 6. กระบวนการ "ประมวลผลสินทรัพย์" (menu 5444, `CFAFrmAssetsProcessDlg`)

ที่มา: `C\FAFrmAssetsProcessDlg.cpp` (3216 บรรทัด; ไฟล์มีโค้ดเก่าหลายชุด — ตัวที่ทำงานจริงคือ `OnBnStartProcessDP` บรรทัด 1360-1595), เทมเพลต `rc:15570` (`IDD_FA_FRA_ASSETS_PROCESSE`, caption "PROCESS ASSETS"), handler `MainFrm.cpp:3005` (`OnFAPROCESS`, DoModal)

### 6.1 ไดอะล็อก

| ลำดับ | ป้ายไทย (key) | ชนิด | บังคับ | หมายเหตุ | หลักฐาน |
|---|---|---|---|---|---|
| 1 | เริ่มจากสินทรัพย์ (`ASSETS_START`) | link + ช่องชื่ออ่านอย่างเดียวข้างขวา | ไม่ | ค้นผ่านรายการ `FAFrmAssetsMasterConfig`; พิมพ์ prefix กรองด้วย `Code LIKE 'xxx%'` | rc:15570+; `.cpp:3036-3072` |
| 2 | ถึงสินทรัพย์ (`TO_FIXED_ASSETS`) | เช่นเดียวกัน | ไม่ | ว่างทั้งคู่ = ทั้งหมด; มีแค่ "เริ่มจาก" = สินทรัพย์ตัวเดียว | เดียวกัน |
| 3 | คำนวณเป็นรายวัน (`PRO_DAY`) | checkbox | ไม่ | ติ๊ก = ProcessType 1 (รายวัน), ไม่ติ๊ก = 0 (ปรับรายปี) | rc:15570+ |
| 4 | ประมวลผล (`PROCESS`) | แถบความคืบหน้า | — | ขยับทีละสินทรัพย์ | rc:15570+ |
| ปุ่ม | OK / Cancel (ข้อความอังกฤษตายตัวใน rc) | | | OK เริ่มทำงาน; Cancel ปิด | `.cpp:1355-1358` |

### 6.2 ขั้นตอนเมื่อกด OK (ไม่มีหน้ายืนยัน — ลบและสร้างใหม่ทันที)

| ขั้น | การทำงาน | หลักฐาน |
|---|---|---|
| 1 | ลบแถวกำพร้า: `BCAssetsOfYear`/`BCAssetsOfPeriod` ที่ `AssetsCode` ไม่อยู่ใน `BCAssetsMaster` | `:1368-1372` |
| 2 | **`UPDATE BCAssetsMaster SET ProcessType = <0/1>` กับสินทรัพย์ทุกตัว** (ไม่จำกัดช่วงรหัส) — รายงานปีใช้ค่านี้ตัดสินโหมดคำนวณทีหลัง | `:1373-1374` |
| 3 | บันทึก log ผู้ใช้ ("ประมวลผลสินทรัพย์" key 90006 + ช่วงรหัส; ใช้ menu id `ID_BILLSYS_TRANS_6` ซึ่งผิดโมดูล) | `:1376-1396` |
| 4 | โหลดงวดบัญชีปีที่ 1-12 จาก `BCPeriod` ตามลำดับ `PeriodNo` | `:1398-1414` |
| 5 | เลือกสินทรัพย์: `SELECT Code, SaleDate, AmountCalc, AssetsRate, AssetsAge, Remain, BuyDate, StartCalcDate, StopCalcDate, DepartCode, AssetsType, LocateCode, AssetsBalance, SumOfAssetsBal, FirstYearRate, CalcType, IsFirstScrap FROM BCAssetsMaster [WHERE Code BETWEEN/= ...] ORDER BY Code`; ไม่มีข้อมูล → `NO_DATA_IN_DB` แล้วจบ | `:1433-1441` |
| 6 | ต่อสินทรัพย์: ลบ `BCAssetsOfYear` และ `BCAssetsOfPeriod` ของรหัสนั้น แล้วรัน `CDepreciation` (§7) ด้วยค่าจาก master (`m_nFirstScrap` = **กลับค่า** `IsFirstScrap`: ในฐาน 1 → engine 0) | `:1456-1500` |
| 7 | เขียนผลต่อ "ปีบัญชี" 1 แถวใน `BCAssetsOfYear` (`AssetsCode, StartCalcDate, StopCalcDate, DayCalc, BeginBalance, AssetsAmount, AssetsAmass, EndBalance, DepartCode, AssetsType, LocateCode`) และ 1 แถวใน `BCAssetsOfPeriod` (`AssetsCode, StartCalcDate, StopCalcDate, BeginBalance, AssetsAmount, AssetsPeriod1..12, EndBalance, DepartCode, AssetsType, LocateCode`; `AssetsPeriod13` ไม่ถูกเขียน) | `:1501-1568` |
| 8 | ถ้าสินทรัพย์มีวันที่ขาย → `UPDATE BCAssetsMaster SET SaleAssetsBal = EndBalance ของปีสุดท้าย` | `:1575-1584` |
| 9 | จบ: ข้อความ `PROCESS_COMPLETE` แล้วปิดไดอะล็อก | `:1590-1594` |

ผลข้างเคียงที่ต้องรู้

| หัวข้อ | ผล |
|---|---|
| GL | **ไม่ลงบัญชี** — กระบวนการนี้เขียนเฉพาะตารางผลการคำนวณ; การลง GL ทำที่เมนู "โอนข้อมูลเข้าสู่ GL" (§8) |
| ล็อกงวด | ไม่ตรวจงวดปิด |
| ทำซ้ำได้ | ได้ — ลบและคำนวณใหม่ทุกครั้ง; ไม่มีสถานะ "ประมวลแล้ว" ในตารางสินทรัพย์ |
| ผลต่อเอกสาร GL เดิม | **ไม่มี** — ถ้าเคยโอนเข้า GL ไปแล้วแล้วประมวลผลใหม่ ตัวเลขใน GL ไม่เปลี่ยน (เสี่ยงต่างจากรายงาน) |
| สิทธิ์ | เมนู `ID_FASYS_TRANS_6` (ผ่านระบบสิทธิ์เมนูกลาง) |

---

## 7. เอนจินคิดค่าเสื่อมราคา (`CDepreciation`)

ที่มา: ฝั่งจอ `C\Depreciation.cpp` (1091 บรรทัด) และฝั่งรายงาน `RG\Depreciation.cpp` (1675 บรรทัด) — **สองสำเนาไม่เหมือนกัน** (§7.5)

### 7.1 อินพุต/เอาต์พุต

| อินพุต (ตัวแปร engine) | มาจาก | หมายเหตุ |
|---|---|---|
| `m_nMethod` | `CalcType` 0/1/2 | `Process()` เลือก `StraightLineMethod / SumOfTheYearDigitsMethod / DecreasingChargeMethod` (`Depreciation.cpp:144-156`); `DoubleDecliningBalanceMethod` เป็นตัวเปล่า ไม่ถูกเรียก (`:792`) |
| `m_nProcessType` | 0 = ปรับยอดปีให้ครบ (true-up), 1 = รายวัน | จอประมวลผล: จาก checkbox; รายงาน: จาก `BCAssetsMaster.ProcessType` ที่บันทึกไว้ครั้งล่าสุด |
| `m_nFirstScrap` | `IsFirstScrap` (กลับค่าในกระบวนการ/รายงาน; **ไม่กลับ**ในปุ่ม Calc/F6) | ดู §7.4 |
| `m_dCost` | `AmountCalc` | |
| `m_dPercent` | `AssetsRate` | อัตราต่อปี (%) |
| `m_dFirstYearPercent` | `FirstYearRate` | ใช้เมื่อ 0 < x ≤ 100 |
| `m_nAge` | `AssetsAge` | `0` = ไม่คิดค่าเสื่อม (เช่นที่ดิน) → 1 แถวปี ค่าเสื่อม 0 (`Depreciation.cpp:343-360`) |
| `m_dScrap` | `Remain` | ราคาซาก |
| `m_BuyDate`, `m_StartDate`, `m_StopDate`, `m_SaleDate` | วันที่ซื้อ, เริ่มคิด, หยุดคิด, ขาย | `Validate()`: ถ้า BuyDate = StartDate ตั้ง `m_dBeginBalance = 0` (`:103-111`) |
| `m_dBeginBalance` | `AssetsBalance` | ถ้า > 0 ใช้เป็นฐานต้นปีแรก (`:314-318`) |
| งวด (`m_Period[]`) | แถว `BCPeriod` ที่ PeriodNo 1-12 (StartDate/StopDate) | ถ้าไม่มี → สร้าง 12 งวดปฏิทินของปีปัจจุบัน (`InitPeriod :90-101`) — **ปีบัญชีไม่ต้องตรงปีปฏิทิน** |
| `AmountPoint` | `CompanyInfo.AmountPoint` | ทศนิยมของการปัด |

เอาต์พุต: รายการ "ปี" `Year[]` — แต่ละปีมี `StartDate, StopDate, Days, dBegin, dAmount, dAccum, dEnd, dDepreciateDay` และรายการ "งวด" 12 ช่อง `Period[]` (`StartDate, StopDate, dPeriodDays, dPeriodAmount`) — ตัวอ่าน: `GetYearAmount/Begin/Accumulate/End/Days/StartDate/StopDate`, `GetPeriodAmount(y,p)`, `GetPeriodDays(y,p)` (`:168-296`)

### 7.2 สูตรของแต่ละวิธี (ฝั่งจอ `Depreciation.cpp`)

| ลำดับ | ขั้นตอน (เส้นตรง `StraightLineMethod :299-549`) | หลักฐาน |
|---|---|---|
| 1 | ฐาน: `dYearBalance = Cost`; ถ้ามี FirstYearPercent: `dCost = Cost - Cost x FY%/100`; ถ้ามี BeginBalance: `dCost = dYearBalance = BeginBalance` | `:307-318` |
| 2 | ถ้า `m_nFirstScrap == 0` : `m_dCost -= m_dScrap` (หักซากออกจากฐานก่อน — แก้ค่าตัวแปรสมาชิกโดยตรง) | `:322` |
| 3 | ค่าเสื่อมเฉลี่ยต่อปี `dDeprecAvgYear = round(m_dCost / (100 / percent))` = ฐาน x อัตรา% | `:327` |
| 4 | ถ้ามีวันที่ขาย และ ≤ วันหยุดคำนวณ → วันหยุดจริง = **วันขาย - 1 วัน** | `:339` |
| 5 | วนทีละ "ปีบัญชี" โดยหาช่วงปีจากงวดที่ 1 ถึงงวดสุดท้ายของ `BCPeriod` (รองรับปีบัญชีข้ามปีปฏิทิน และกุมภาพันธ์สิ้นเดือน 28/29) นับ `nYearDays` (วันทั้งปีบัญชี) และ `nYearDeprecDays` (วันที่ทรัพย์สินอยู่ในปีนั้น) | `Depreciation.cpp:369-460` (ยังไม่ได้ไล่อ่านทุกกิ่งของการหาช่วงปี) |
| 6 | ต่อ "งวด" j: ตัดช่วงวันให้อยู่ในช่วง [StartDate, StopDate] ของสินทรัพย์; ข้ามงวดที่ยังไม่ถึง/เลยแล้ว; หยุดเมื่อยอดคงเหลือปี ≤ ราคาซาก | `:455,462-466,541` |
| 7 | **ค่าเสื่อมต่อวัน** `= (m_dCost x percent/100) / nYearDays`; **ค่าเสื่อมงวด** `= round(ค่าต่อวัน x จำนวนวันของงวด)` (ปัดตาม `AmountPoint`) | `:470`, `:477` |
| 8 | FirstYearPercent (0 < x ≤ 100): บวก `m_dCost x FY%/100` เข้า "งวดที่เดือนเริ่มต้นตรงกับเดือนเริ่มคิดค่าเสื่อม" ของปีแรกเท่านั้น (เงื่อนไขปีใช้ `StartDate.Year == StartDate.Year + i` จึงจริงเฉพาะ i = 0) | `:478-483` |
| 9 | กันไม่ให้ต่ำกว่าซาก: ถ้า `round(yearBalance - round(accum + amount)) < scrap` → `amount = amount + round(yearBalance - round(accum+amount)) - scrap` | `:486-489` |
| 10 | **ProcessType 0** และเป็นงวดสุดท้ายของ `m_Period` และปีนั้นมี `nYearDeprecDays` ≥ 365: ปรับให้ยอดสะสมของปีเท่า `dDeprecAvgYear` (`amount += avgYear - accum`) — ทำให้รายปีลงตัวพอดีกับสูตรปีเต็ม | `:494-500` |
| 11 | งวดที่ตรงกับวันหยุดคำนวณ (และเป็นเดือนสุดท้ายของปีบัญชี): ถ้าไม่มีวันขาย → ดันยอดที่เหลือ (`yearBalance - accum`) ลงงวดนี้ แล้วถ้า `m_nFirstScrap == 1` หักซากออกจากงวดและจากสะสม; ถ้าขายหลัง `m_StopDate` → ดันยอดที่เหลือเหมือนกันแต่ไม่หักซาก; ถ้าขายก่อน → ปรับเฉพาะกรณียอดติดลบ | `:502-524` |
| 12 | สรุปปี (เฉพาะเมื่อ Σงวด > 0 จึงเพิ่มปีนี้เข้ารายการ): `dAmount = round(Σงวด)`, `dBegin = round(yearBalance)`, `dEnd = round(yearBalance - dAmount)`, `dAccum = round(Σสะสม)` หรือถ้ามี BeginBalance = `round((m_dCost - BeginBalance) + Σสะสม)`; `yearBalance -= dAmount` | `:526-540` |

| วิธี | ต่างจากเส้นตรงที่ | หลักฐาน |
|---|---|---|
| 1 ผลรวมจำนวนปี (`SumOfTheYearDigitsMethod :551-791`) | ค่าเสื่อมต่อปี `dDeprecAvgYear = round((nAgeDec / (N(N+1)/2)) x m_dCost)` โดย `N = อายุ`, `nAgeDec` เริ่ม `N` แล้วลด `(ปีเริ่มคิด - ปีซื้อ)` ถ้าซื้อก่อนปีเริ่มคิด และลดทีละ 1 ทุกปี; ค่าต่อวัน = `avgYear / nYearDays`; วันหยุดเมื่อขาย = **วันขาย** (ไม่ลบ 1 วัน); **หักซากท้ายตารางเสมอ** (ไม่ดู `m_nFirstScrap`) และไม่มีการหักซากออกจากฐานตอนต้น | `:576-577,605-613,718,753-754` (ตัวแปร `nAgeDec` ประกาศที่ `:~600` ยังไม่ยืนยันเลขบรรทัด) |
| 2 ลดน้อยถอยลง (`DecreasingChargeMethod :797-1037`) | ค่าเสื่อมต่อปี `= round(dYearBalance x percent/100)` (ยอดคงเหลือต้นปี x อัตราที่ผู้ใช้กรอก — ไม่ใช่สูตร 2/อายุ); ค่าต่อวัน = `(dYearBalance x percent/100) / nYearDays`; วันหยุดเมื่อขาย = วันขาย; หักซากท้ายตารางเสมอ | `:825-826,964-965,1001-1002` |

### 7.3 ฟังก์ชัน `GetBookValue(begin, end, &dBegin, &dAccu, &dBalance)` (ใช้ในรายงานประจำปี/รายงานขาย)

1. ตั้ง `m_SaleDate = begin` แล้ว `Process()` → เอา "ปีสุดท้าย" ของตาราง = สถานะ ณ วัน begin (ค่าเสื่อมสะสมถึงก่อน begin)
2. ตั้ง `m_SaleDate = end + 1 วัน` แล้ว `Process()` อีกครั้ง → สถานะ ณ สิ้น end
3. ค่าเสื่อมในช่วง = `สะสม(end) - สะสม(begin)` (ถ้าไม่มีสะสมตอน begin ใช้ค่าเสื่อมปีสุดท้ายตรง ๆ); ยอดยกมา = `Cost - สะสม(end) + ค่าเสื่อมช่วง`; ยอดยกไป = `EndBalance(end)`; ถ้ายอดยกไป = 0 และมีวันขาย → สะสม -= ซาก, ยอดยกไป = ซาก
(ฝั่งจอ `Depreciation.cpp:1039-1091`; ฝั่งรายงานเพิ่ม `GetDepreciatePeriod/…2`, `GetBookValue2` ที่ `RG\Depreciation.cpp:1281-1675` — ยังไม่ได้อ่านรายละเอียดทั้งหมด)

### 7.4 ความหมายของ "หักค่าซากก่อน" (Yes/No) — **สับสนจริงในต้นฉบับ**

| จุดใช้งาน | ค่าที่ส่งเข้า engine | ผลใน engine เส้นตรง |
|---|---|---|
| จอทะเบียน (บันทึก/ปุ่ม Calc/F6) | `IsFirstScrap` ตามที่ผู้ใช้เลือก (Yes = 1) | `m_nFirstScrap = 1` → **หักซากตอนท้าย** (ไม่หักจากฐาน) |
| กระบวนการประมวลผล, รายงานปี, รายงานขาย | `m_nFirstScrap = (IsFirstScrap != 0) ? 0 : 1` (กลับค่า) | Yes → `m_nFirstScrap = 0` → **หักซากจากฐานตั้งแต่ต้น** |
| ผลสุทธิ | ปุ่ม Calc ที่หน้าจอกับตัวเลขหลังประมวลผลอาจไม่เท่ากันเมื่อเลือก "หักซากก่อน" | จอทะเบียน `FAFrmAssetsMaster.cpp:924,1065` (ใช้ค่า combo ตรง ๆ); ประมวลผล `FAFrmAssetsProcessDlg.cpp:1512-1513`; รายงานปี `FAsRepAssetsOrderCodeView.cpp:178,211`; รายงานขาย `FARepSellItemView.cpp:209,243` (กลับค่า) |

### 7.5 สองสำเนา engine ไม่เหมือนกัน (ยังไม่ verify ผลต่อเลขจริง)

| เรื่อง | ฝั่งจอ (`C\Depreciation.cpp`) | ฝั่งรายงาน (`RG\Depreciation.cpp`) |
|---|---|---|
| งวดตั้งต้นถ้าไม่มี | สร้าง 12 งวดปฏิทินถ้า **ไม่มีเลย** (`:104`) | สร้างถ้า **น้อยกว่า 12** (`:105`) |
| หักซากจากฐาน | `m_dCost -= m_dScrap` (แก้ตัวแปรสมาชิก; กระทบ `dAccum` ที่ใช้ `m_dCost - BeginBalance`) (`:322`) | ใช้ตัวแปรท้องถิ่น `dCostAmount` ไม่แตะ `m_dCost` (`:312`) |
| จำนวนวันตลอดอายุ | `StopDate - m_StartDate + 1` (`:328`) | `m_StopDate - m_BuyDate + 1` (`:328`) |
| วันหยุดเมื่อขาย (เส้นตรง) | วันขาย - 1 (`:339`) | วันขาย (`:337`) |
| ค่าต่อวันต่อปี | คำนวณ `(m_dCost x rate%)/nYearDays` เสมอ (`:470`) | คำนวณเฉพาะ `ProcessType == 0`; โหมดรายวัน (1) ใช้ `dCostAmount / จำนวนวันตลอดอายุ` (`:473-479`) |
| true-up สิ้นปี | เติมยอดต่างเสมอ | เติมเฉพาะเมื่อค่าเสื่อมงวดสุดท้าย > 0 (`:502`) |
| หักซากท้ายตาราง | เฉพาะ `m_nFirstScrap == 1` | **ไม่ตรวจเงื่อนไข** (คอมเมนต์ปิดไว้) (`:515`) |
| ฟังก์ชันเพิ่ม | — | `GetDepreciatePeriod`, `GetBookValue2`, `GetDepreciatePeriod2` |

ผลสำหรับ BC: ต้องเลือก "สูตรฐานเดียว" แล้วให้ทุกหน้าจอ/ประมวลผล/รายงานเรียกตัวเดียวกัน. จากตำแหน่งไฟล์ `FAFrmAssetsProcessDlg.cpp` อยู่ในโปรเจกต์ `champchamp` (ใช้ `Depreciation.cpp` ฝั่งจอ — ยังไม่ verify ผ่าน include เพราะไฟล์รวม `gain.h`) ส่วนรายงานปีคำนวณสดด้วยฝั่งรายงาน จึงมีโอกาสที่ "รายงานรายเดือน (อ่านค่าที่เก็บ)" กับ "รายงานรายปี (คำนวณสด)" ต่างกันเล็กน้อยเมื่อมีวันขาย/ซาก — ต้องให้ลุงจืดตัดสินว่า BC ยึดสูตรไหนเป็นหลัก และเทียบตัวเลขกับ Champ ด้วยข้อมูลชุดเดียวกันก่อนลอกสูตร (ดู §13).
---

## 8. "โอนข้อมูลเข้าสู่ GL" (menu 5491, `CFSTransferToGLDlg`)

ที่มา: `C\FSTransferToGLDlg.cpp` (1168 บรรทัด), เทมเพลต `rc:5685` (`IDD_FS_TRANSFERTOGL`, 289x281), handler `MainFrm.cpp:2996` (`OnFATRANSFERGL`, DoModal) — ชื่อไดอะล็อก "โอนข้อมูลเข้าสู่ GL" (key 90007)

### 8.1 ไดอะล็อก (ตามลำดับบนจอ)

| กลุ่ม | ป้ายไทย (key) | ชนิด | บังคับ | หมายเหตุ | หลักฐาน |
|---|---|---|---|---|---|
| เอกสาร (กรอบ "Document") | เลขที่เอกสาร (`DOCNO`, อนุมาน — .rc เสียรหัสภาษา) + ปุ่มรันเลข + ปุ่ม `...` เลือกเอกสารเดิม | `BRunningEdit` ตามรูปแบบ `GLJournalConfig` (ตั้งในหน้ารูปแบบเลขที่ของสมุดรายวัน) | บังคับ (`INSERT_DOCNO`) | ใช้เลขที่เอกสาร GL ชุดเดียวกับสมุดรายวันทั่วไป | `.cpp:74-95,166-170,1022-1070` |
| | วันที่เอกสาร (`DOCDATE`, อนุมาน) | วันที่ (พ.ศ.) | — | เปิดจอ/บันทึกจะตรวจ "งวดบัญชีปิดหรือไม่" (`IsPeriodClose`) | `:96,129-131,361-365` |
| | สมุดรายวัน (`BOOKCODE`, อนุมาน) | combo `BCGLBook` | บังคับ (`INSERT_BOOKNO`) | | `:98,172-177` |
| | คำอธิบาย (`DESCRIPTION`, อนุมาน) | text หลายบรรทัด | ไม่ | เก็บ `BCTrans.MyDescription` | rc:5685+ |
| เงื่อนไข (`CONDITION` "เงื่อนไข") | รหัสสินทรัพย์ (`ASSETS_CODE`) จาก / ถึง (`TO` "ถึง") | link → `FAFrmAssetsMasterConfig` | ไม่ | | `:95-96,918-940` |
| | ประเภทสินทรัพย์ จาก / ถึง (`FROM_TYPE_ASSETS`) | combo `BCAssetsType` | **บังคับ "จาก"** (ยกเว้นโหลดเอกสารเดิม) — ข้อความเตือนฮาร์ดโค้ดอังกฤษ "Please enter assetsType" | | `:178-182` |
| | แผนก จาก / ถึง (`FROM_DEPART_CODE` "จากแผนก") | combo `BCDepartment` | ไม่ | | `:101-102` |
| | งวดที่ (`PERIOD_AT`) จาก / ถึง | ตัวเลข 1-12 | ไม่ (ว่าง = 1 ถึง 12; ใส่แค่ "จาก" = งวดเดียว) | ใช้ลำดับงวดที่ 1-12 ของ `AssetsPeriodN` | `:710-720` |
| | ประจำปี (`FIXED_YEAR` ":ประจำปี") | ตัวเลข ปี (พ.ศ. หรือ ค.ศ. > 2400 จะลบ 543) | **บังคับ** (ยกเว้นโหลดเอกสารเดิม; ข้อความเตือนซ้ำกับข้อความประเภท) | | `:183-193` |
| | แยกตามแผนก (`SEPARATE_DEPARTMENT`, อนุมาน) | checkbox | ไม่ | ติ๊ก + มี "แผนกจาก" → แยกบรรทัดต่อ (ประเภท, แผนก) และใส่ `DepartCode` ในรายการ | `:48,223-229,1098-1105` |
| ตาราง | รหัสบัญชี / ชื่อบัญชี / เดบิต / เครดิต | กริด 4 คอลัมน์ ชื่อบัญชีอ่านอย่างเดียว; เดบิต/เครดิต **แก้ไขได้** | — | ชื่อบัญชีเลือก `Name1` (ไทย) หรือ `Name2` ตามโหมดภาษาของผู้ใช้ | `:104-124,640-660` |
| ปุ่ม | Process / Save / Cancel (ข้อความอังกฤษใน rc); `Save` ปิดอยู่จนกว่าจะ Process สำเร็จ หรือโหลดเอกสารเดิม | | | | `:127,350,1133-1135` |
| แถบความคืบหน้า | `PROCESS` | | | | rc:5685+ |

### 8.2 ขั้น Process (ดูตัวอย่าง — ยังไม่เขียนฐานข้อมูล)

| ขั้น | การทำงาน | หลักฐาน |
|---|---|---|
| 1 | ตรวจ: เลขที่เอกสาร, สมุดรายวัน, ประเภทจาก, ปี (ถ้าไม่ได้โหลดเอกสารเดิม) | `:166-192` |
| 2 | หา "ประเภท (+แผนก)" ที่มีข้อมูล: `SELECT DISTINCT ASSETSTYPE [,DEPARTCODE] FROM BCASSETSOFPERIOD WHERE YEAR(STARTCALCDATE)=ปี AND YEAR(STOPCALCDATE)=ปี [AND ASSETSCODE ...] [AND ASSETSTYPE ...] [AND DEPARTCODE ...] ORDER BY ASSETSTYPE[,DEPARTCODE]` | `:227-265` |
| 3 | ต่อ (ประเภท, แผนก): `ProcessAmount` = Σ `AssetsPeriod[จากงวด..ถึงงวด]` ของ `BCAssetsOfPeriod` ที่ `StartCalcDate` อยู่ระหว่าง **1 ม.ค.** ของปี ถึงวันสุดท้ายของเดือนที่เท่ากับ "งวดถึง" (กำหนดตายตัวเป็นเดือนปฏิทิน), ตรงตามช่วงรหัส/ประเภท/แผนก, และสินทรัพย์ยังมีใน `BCAssetsMaster` | `:682-860` |
| 4 | ดึงบัญชีจาก `BCAssetsType` : เดบิต = `DrAccountCode`, เครดิต = `CrAccountCode` | `:662-680` |
| 5 | เติมกริด 2 บรรทัดต่อ 1 ประเภท: บรรทัดเดบิต (ยอดในช่อง "เดบิต") และบรรทัดเครดิต (ยอดในช่อง "เครดิต") | `:326-345` |
| 6 | เปิดปุ่ม Save | `:349-350` |

### 8.3 ขั้น Save (เขียน `BCTrans` + `BCTransSub`)

| ขั้น | การทำงาน | หลักฐาน |
|---|---|---|
| 1 | งวดบัญชีของวันที่เอกสารปิดแล้ว → หยุด | `:361-365` |
| 2 | ตรวจเงื่อนไขซ้ำ (เหมือนขั้น Process) | `:367-369` |
| 3 | **ตรวจเลขที่เอกสารซ้ำ**: ถ้ามี `BCTrans.DocNo` นี้อยู่แล้ว ถามยืนยัน (`HAVE_DUPLICATE` หรือ `HAVE_DUPLICATE2` เมื่อโหลดเอกสารเดิม — ข้อความตามคอมเมนต์ "ข้อมูลซ้ำซ้อน ต้องการเปลี่ยนแปลงข้อมูลหรือไม่?"): ไม่ → ล้างเลขที่แล้วหยุด; ใช่ → ลบ `BCTrans`/`BCTransSub` ที่ `DocNo` นั้นและ `Source = -2` | `:371-373,1073-1103` |
| 4 | ตรวจ "ยอดสมดุล" ของกริด: Σ คอลัมน์เดบิต = Σ คอลัมน์เครดิต (เทียบด้วยทศนิยม `AmountPoint`) ไม่งั้น `CANNOT_SAVE_NORIGHT_DEBITCREDIT` | `:386-415` |
| 5 | วนตามประเภท (+แผนก) อีกรอบ **โดยใช้ยอดจากกริดช่อง "เดบิต" ของแต่ละคู่บรรทัด** (แก้ตัวเลขในช่องเดบิตได้ แก้ช่องเครดิตมีผลแค่การตรวจสมดุล) ข้ามประเภทที่ไม่มีบัญชีทั้งสองฝั่ง | `:501-538` |
| 6 | ตัวแรก: ลบ `BCTrans`/`BCTransSub` ที่ `DocNo + BookCode + Source=-2` แล้ว insert หัว `BCTrans` (`BookCode, DocDate, DocNo, Amount = Σยอดเดบิตในกริด, Source = -2, MyDescription`, ถ้า `CompanyInfo.AutoPostGL = 1` → `IsConfirm = 1` และเรียก `CGLAutoProcess.OnProcessPeriod(docNo, docDate, bookNo, -2)`) | `:547-575` |
| 7 | ต่อประเภท: insert `BCTransSub` 2 บรรทัด (เดบิต `AccountCode = DrAccountCode`, เครดิต `AccountCode = CrAccountCode`; `DepartCode` ถ้าแยกแผนก; `Source = -2`; `LineNumber` เรียง 0,1,2…; `IsConfirm` ตาม AutoPostGL) — **ยอด 0 ไม่ถูกบันทึก** | `:577-614` |
| 8 | สำเร็จ → `PROCESS_COMPLETE` ปิดไดอะล็อก; ไม่มีข้อมูล → "Can't process,Haven't data." (อังกฤษตายตัว) | `:624-628` |

โหลดเอกสารเดิม (ปุ่ม `...` ข้างเลขที่): รายการ `BCTrans` ที่ `Source = -2 AND isconfirm = 0` (เอกสารที่ยังไม่ผ่านบัญชี) → เติมสมุด/เลขที่/วันที่ และโหลดบรรทัดจาก `BCTransSub` เข้ากริด; ปิดปุ่ม Process เปิดปุ่ม Save (`:1106-1168`)

หลักการสำคัญที่ BC ต้องเก็บ

| หัวข้อ | กติกา Champ |
|---|---|
| หนึ่งเอกสาร | ต่อ (ประเภท[, แผนก]) 1 คู่ เดบิตค่าเสื่อมราคา / เครดิตค่าเสื่อมราคาสะสม; ไม่มีแยกสินทรัพย์รายตัวใน GL |
| รวมยอดจาก | ค่าที่ **เก็บไว้** ใน `BCAssetsOfPeriod` ณ ตอนประมวลผล (ไม่คำนวณใหม่) |
| แก้ตัวเลขก่อนบันทึก | ได้ในกริด (ผู้ใช้ปรับเศษ/ปรับรายการ) แต่ต้องดุล |
| ซ้ำ | เลขที่เอกสารเดียวกันยืนยันแล้วทับของเดิม (เฉพาะ Source -2) |
| ย้อนรายการ | ลบ/เขียนทับ ณ เอกสารที่ยังไม่ผ่านบัญชี (`isconfirm = 0`); ไม่มีเมนู "กลับรายการ" ของสินทรัพย์ — เมื่อผ่านบัญชีแล้วแก้ที่สมุดรายวัน |
| งวดปิด | บล็อกที่ `IsPeriodClose` |
| ภาษี | ไม่เกี่ยว (ไม่มี VAT/WHT) |

---

## 9. การขาย / ตัดจำหน่ายสินทรัพย์

| หัวข้อ | Champ | หลักฐาน |
|---|---|---|
| บันทึกการขาย | กรอกในกลุ่ม "การขายสินทรัพย์" ของจอทะเบียน (§3.2 ข้อ 24-29) — **ไม่มีเมนู/เอกสารขายสินทรัพย์แยก** และไม่เชื่อมใบกำกับภาษี/ลูกหนี้ | `FAFrmAssetsMaster.cpp:414-487` |
| ผลต่อค่าเสื่อม | วันหยุดคำนวณเปลี่ยนเป็นวันขาย (-1 วัน เส้นตรง ฝั่งจอ) → ประมวลผลครั้งถัดไปไม่คิดค่าเสื่อมหลังวันขาย; กระบวนการประมวลผลอัปเดต `SaleAssetsBal` อัตโนมัติ | `Depreciation.cpp:339`; `ProcessDlg.cpp:1575-1584` |
| ผลต่อรายงานปี | แถวที่ขายก่อนต้นช่วง → ไม่แสดง; ขายก่อนสิ้นช่วง → ยอดยกไป = 0 | `FAsRepAssetsOrderCodeView.cpp:228-262,271-272` |
| ผลต่อ GL | **ไม่มี** — Champ ไม่ลงบัญชีการขาย/ตัดจำหน่าย (ตัดสินทรัพย์ ค่าเสื่อมสะสม และรับรู้กำไร/ขาดทุน ต้องคีย์สมุดรายวันเอง) | ไม่พบโค้ดเขียน `BCTrans` ใน `FAFrmAssetsMaster.cpp` |
| รายงาน | "รายงานสินทรัพย์ที่ถูกขายพร้อมกำไรขั้นต้น" (§10 หมายเลข 22) กำไร = ราคาขาย - มูลค่าคงเหลือที่เก็บ | `FARepSellItemView.cpp:296` |
| ตัดจำหน่ายแบบไม่ขาย (เสีย/สูญหาย) | **ไม่มีฟังก์ชันเฉพาะ** (ต้องใช้วันที่ขาย + ราคาขาย 0) | — |
| ปัญหาที่ตัดสินใจ | ข้อมูลขายที่ผ่านกำหนดคิดค่าเสื่อม (วันขายหลังวันหยุด) → ใช้ราคาซากเป็นมูลค่าคงเหลือ (ปุ่ม Calc) | `FAFrmAssetsMaster.cpp:889-894` |

---

## 10. รายงานสินทรัพย์ 22 รายการ

ที่มา: คลาส View/Dlg อยู่ที่ `RG\` (โปรเจกต์ `BC5REP_GLFS`); ทุกรายงานเปิดผ่านกรอบรายงานมาตรฐาน (กล่อง filter → พรีวิว F11 → พิมพ์/ส่งออก) ตามที่อธิบายใน `10-reports.md` — ไฟล์นี้บันทึกเฉพาะ "ตัวกรอง/คอลัมน์/การเรียง/ยอดรวม" ของแต่ละรายงาน. ข้อความไทยของคอลัมน์มาจากตาราง `myresource` ที่ไม่อยู่ใน repo: คีย์ที่ถอดไทยได้จากไฟล์ .rc/.aps แสดงเป็นไทย, ที่เหลือแสดงเป็น "(อนุมาน)" พร้อมคีย์

### 10.1 กล่อง filter (ทุกรายงานสินทรัพย์)

| ลำดับ | ป้ายไทย | ช่อง (resource id) | ชนิด | บังคับ | หลักฐาน |
|---|---|---|---|---|---|
| 1 | พิมพ์จากวันที่เริ่มคำนวณ: / พิมพ์ถึงวันที่เริ่มคำนวณ: (`FROM_START_CALCDATE`, `TO_STOP_CALCDATE`) — รายงานขายใช้ "วันที่ขาย" | `IDC_ED_START_DATE`, `IDC_ED_STOP_DATE` | วันที่ (พ.ศ.ตามตั้งค่าบริษัท) | กลุ่มรายละเอียด: ไม่บังคับ; **กลุ่มรายเดือน: บังคับทั้งสองช่อง และ "ถึง" ต้องไม่น้อยกว่า "จาก"** (ข้อความ `SETDATE_COMPUTE` "กรุณากำหนดวันที่เริ่มคำนวณด้วย", `ENDDATE_COMPUTE`, `ENDDATE_COMPUTE_CORRECT`) | `RG\FARepDeperciateDepartDlg.cpp:141-165`; `FARepAssetsItemByItemDlg.cpp:142` |
| 2 | พิมพ์จากสินทรัพย์: / พิมพ์ถึงสินทรัพย์: | `IDC_ED_START_FIXED`, `IDC_ED_STOP_FIXED` | link → `BCAssetsMaster.Code` (แสดง รหัส 100 + ชื่อ 400) | ไม่ | `FAsRepAssetsOrderCodeDlg.cpp:97-100` |
| 3 | พิมพ์จากแผนก: / พิมพ์ถึงแผนก: | `IDC_ED_START_DEPART`, `IDC_ED_STOP_DEPART` | combo `BCDepartment` | ไม่ | `:107-108` |
| 4 | พิมพ์จากประเภท: / พิมพ์ถึงประเภท: | `IDC_ED_START_TYPE`, `IDC_ED_STOP_TYPE` | combo `BCAssetsType` | ไม่ | `:110-111` |
| 5 | พิมพ์จากที่ตั้ง: / พิมพ์ถึงที่ตั้ง: (`FROM_PLACE`) | `IDC_ED_START_PLACE`, `IDC_ED_STOP_PLACE` | combo `BCAssetsLocate` | ไม่ | `FARepDeperciateDepartDlg.cpp:57-58` |
| 6 | คำนวณเป็นวัน (`CK_DAY`) — เฉพาะรายงานรายปี | `IDC_CK_DAY` | checkbox | ไม่ | `FAsRepAssetsOrderCodeDlg.cpp:61` |
| — | ช่วงวันที่ของรายงานรายปี: ค่าเริ่มต้น `01/01/<ปีปัจจุบัน>` ถึง `31/12/<ปีปัจจุบัน>` (พ.ศ. เมื่อ `YearType = 0`) — **ฮาร์ดโค้ดปฏิทิน ไม่ใช้ปีบัญชีใน `BCPeriod`** | `IDC_ED_START_DATE/STOP` | | | `FAsRepAssetsOrderCodeDlg.cpp:85-94` |

กติการ่วม: ช่องจาก/ถึงว่างทั้งคู่ = ไม่กรอง; ใส่แค่ "จาก" = เท่ากับค่านั้น (`= '...'`); ข้อความเงื่อนไขที่เลือกถูกเก็บใน `m_saReportCondition` แล้วพิมพ์ใต้หัวรายงาน (ตัวอย่าง `FARepDeperciateDepartDlg.cpp:212`). รายงานขายใช้ `SaleDate BETWEEN` (และกรณีวันเดียวใช้ชื่อคอลัมน์ผิด `SaleDates` — `FARepSellItemDlg.cpp:160-165`). ปุ่มและคีย์ของกล่อง filter เป็นของกรอบกลาง (ดู `10-reports.md` §framework).

### 10.2 กลุ่ม "รายละเอียดสินทรัพย์" (6 รายงาน, อ่าน `BCAssetsMaster` ตรง ๆ)

SQL: `SELECT * FROM BCAssetsMaster` + เงื่อนไข `StartCalcDate BETWEEN`, `Code`, `DepartCode`, `AssetsType`, `LocateCode` (`RG\FARepAssetsItemBy*Dlg.cpp:26,142-222`)

| menu | รายงาน (คลาส) | คีย์ชื่อรายงาน | เรียงตาม | คอลัมน์แรกสุด | หลักฐาน |
|---|---|---|---|---|---|
| 5424 | ตามรหัส (`FARepAssetsItemByItemView`) | 94001 | `Code, SaleDate, SaleDocNo, StartCalcDate, LocateCode` | รหัส | Dlg:236; View:28-29,140 |
| 5425 | ตามประเภท (`...ByType`) | 94002 | `AssetsType, Code, SaleDate, SaleDocNo, StartCalcDate, LocateCode` | รหัส (หัวกลุ่ม = ประเภท `TYPE_CODE`) | Dlg:236; View:140 |
| 5426 | ตามแผนก (`...ByDepart`) | 94003 | `DepartCode, Code, SaleDate, SaleDocNo, StartCalcDate, LocateCode` | รหัส (หัวกลุ่ม = แผนก `DEPART_CODE`) | Dlg:240; View:137 |
| 5427 | ตามที่ตั้ง (`...ByPlace`) | 94004 | `LocateCode, Code, SaleDate, SaleDocNo, StartCalcDate` | รหัส (หัวกลุ่ม = ที่ตั้ง `LOCATION_CODE`) | Dlg:240; View:138 |
| 5428 | ตามวันที่ซื้อ (`...ByDate`) | 94005 | `BuyDate, Code, ...` | **วันที่ซื้อ** (คอลัมน์ที่ 0) แล้ว รหัส | Dlg:238; View:140 |
| 5429 | วันที่คำนวณ (`...ByProcess`) | 94006 | `StartCalcDate, Code, SaleDate, SaleDocNo, LocateCode` | **วันที่เริ่มคำนวณ** (คอลัมน์ที่ 0) แล้ว รหัส | Dlg:236; View:138 |

คอลัมน์ (ตัวอย่าง `...ByItem`, `View:140-149`): รหัส (`CODE`) → ชื่อสินทรัพย์ (`ASSETS_NAME` "ชื่อสินทรัพย์:") → หน่วยนับ (`UNIT`) → วันที่เริ่มคำนวณ (`START_DATE_CALC`, อนุมาน) → จำนวนเงิน (`AMOUNT` = `AmountCalc`, ชิดขวา) → อัตราค่าเสื่อมราคา (`FA_ASSETS_RATE`, อนุมาน = `AssetsRate`) → ราคาซาก (`FA_REMAIN_PRICE`, อนุมาน = `Remain`). ทศนิยมตาม `CompanyInfo.AmountPoint`. ทุกตัวต่างกันเฉพาะหัวกลุ่ม/คอลัมน์ที่ย้ายมาอยู่หน้าสุด (5428, 5429)
ยอดรวม: ท้ายกลุ่ม/ท้ายรายงานแสดง "รวม n รายการ" (`SUMCOUNT` + จำนวน + `INSTANCE`) และผลรวมจำนวนเงินกับราคาซาก (`View:222-225`); **ไม่มีผลรวมของอัตรา**; เมื่อ `CompanyInfo.IsPrintSumPage = 0` มีแถวสรุปต่อหน้า (`:232-236`) (เงื่อนไขนี้ใช้ในทุกรายงานด้านล่าง)
ข้อสังเกต: `...ByDate` มีโค้ดท้ายกลุ่มที่รวมคอลัมน์ 4-9 (`View:258-263`, ใช้ชื่อตัวแปรของรายงานปี) เกินจำนวนคอลัมน์ 8 ของจอนี้ — ยังไม่ verify ว่าพิมพ์ออกจริงหรือไม่

### 10.3 กลุ่ม "ค่าเสื่อมราคาเป็นปี" (6 รายงาน)

| menu | รายงาน (คลาส) | คีย์ชื่อ | แหล่งตัวเลข | เรียงตาม | หลักฐาน |
|---|---|---|---|---|---|
| 5445 | แผนก,ประเภท (`FAsRepAssetsOrderDepartView`) | 95001 | **คำนวณสดต่อสินทรัพย์** ด้วย `GetBookValue(begin,end)` | `DepartCode, AssetsType, Code` | View:549 |
| 5446 | ประเภท,แผนก (`...OrderTypeView`) | 95002 | คำนวณสด | `AssetsType, DepartCode, Code` | View:557 |
| 5447 | ตามที่ตั้ง (`...OrderLocateView`) | 95003 | คำนวณสด | `LocateCode, Code` | View:458 |
| 5448 | ตามรหัส (`...OrderCodeView`) | 95004 | คำนวณสด | `Code` | View:407 |
| 5449 | สรุปค่าเสื่อมทั้งหมดเป็นปี (`...SumOfYearView`) | 95006 | **อ่านค่าที่เก็บ** จาก `BCAssetsOfYear` (ไม่คำนวณใหม่) | `AssetsCode, StartCalcDate` | Dlg:38,116; View:103,138-188 |
| 5450 | ตาม ภ.ง.ด.50 (`...OrderP50View`) | 95007 | คำนวณสด | `AssetsType, Code` | View:469 |

กฎของ 5445-5448 และ 5450 (คอลัมน์เดียวกันทั้งหมด `OrderCodeView.cpp:358-368`)

| ลำดับ | คอลัมน์ | คีย์ | หมายเหตุ |
|---|---|---|---|
| 0 | รหัสสินทรัพย์ | `ASSETS_CODE` | |
| 1 | ชื่อสินทรัพย์ | `ASSETS_NAME` | |
| 2 | หน่วยนับ | `UINTCODE` | |
| 3 | วันที่ซื้อหรือได้มา | `FA_BUYDATE` | |
| 4 | วันที่เริ่มคำนวณ | `START_DATE_CALC` (อนุมาน) | |
| 5 | จำนวนเงิน | `AMOUNT` | = ราคาที่คิดค่าเสื่อม |
| 6 | ยอดยกมา | `TRANSFER__BALANCE` (อนุมาน) | `GetBookValue` → `dBegin` |
| 7 | อัตราส่วน | `RATIO` | |
| 8 | ค่าเสื่อมปีนี้ | `ASSETSAMOUNT` (อนุมาน) | |
| 9 | ค่าเสื่อมสะสม | `ASSETSAMASS` (อนุมาน) | |
| 10 | ยอดยกไป | `ENDTRANSFER__BALANCE` (อนุมาน) | |

กติกา: เลือกสินทรัพย์ที่ `StartCalcDate <= วันที่สิ้นสุดที่กรอก`; ช่วงวันที่ = ค่าที่กรอก (ค่าเริ่มต้นปีปฏิทิน §10.1); `ProcessType` มาจากทะเบียนสินทรัพย์ (ค่าที่กระบวนการประมวลผลบันทึกไว้ล่าสุด — ไม่ใช่ checkbox "คำนวณเป็นวัน" ที่หน้ากรอง, ยังไม่ verify ว่า CK_DAY ถูกนำไปใช้ที่ใด) `View:176,209`; แถวที่ **ขายก่อนวันต้นช่วง ไม่แสดง**, แถวที่ขายก่อนวันสิ้นช่วง **ยอดยกไปเป็น 0** (`:228-262,271-272`); ท้ายกลุ่ม/ท้ายรายงาน "รวม n รายการ" + ผลรวมของคอลัมน์ 5-10 (`:328`); **5450 ไม่มีคอลัมน์เฉพาะภาษี** (ไม่มีส่วนเพิ่ม/ส่วนต่างทางภาษี) — เป็นเพียงรายงานปีจัดกลุ่มตามประเภทสำหรับแนบ ภ.ง.ด.50
5449: แต่ละสินทรัพย์ 1 บรรทัดหัว (รหัส, ชื่อ) แล้วรายการต่อปีของ `BCAssetsOfYear` : วันที่เริ่มคำนวณ, วันที่สิ้นสุดคำนวณ (`FA_STOP_CALCDATE`), หน่วยนับ, จำนวนเงิน, ยอดยกมา, อัตราส่วน, ค่าเสื่อมปีนี้, ค่าเสื่อมสะสม, ยอดยกไป (`View:248-259`); ค่า 0 ในคอลัมน์เงินพิมพ์เป็นว่าง (`:164,176,182,188`); กล่อง filter มีเฉพาะช่วงรหัสสินทรัพย์ (`FAsRepAssetsSumOfYearDlg`)

### 10.4 กลุ่ม "ค่าเสื่อมราคาเป็นเดือน" (6 รายงาน, อ่านค่าที่เก็บใน `BCAssetsOfPeriod`)

| menu | รายงาน (คลาส) | คีย์ชื่อ | หัวกลุ่ม | เรียงตาม | หลักฐาน (Dlg) |
|---|---|---|---|---|---|
| 5430 | แผนก,ประเภท (`FARepDeperciateDepartView`) | 96001 | รหัสแผนก + ชื่อแผนก (`DEPART_CODE`, `DEPART_NAME`) | `DepartCode, AssetsType, AssetsCode, StartCalcDate, LocateCode` | :358-361; View:132-133 |
| 5431 | ประเภท,แผนก (`...TypeView`) | 96002 | รหัสประเภท + ชื่อประเภท (`TYPE_CODE`, `TYPE_NAME`) | `AssetsType, DepartCode, AssetsCode, StartCalcDate, LocateCode` | :351-354; View:130-131 |
| 5432 | ตามที่ตั้ง (`...LocateView`) | 96003 | รหัสที่ตั้ง + ชื่อที่ตั้ง (`LOCATION_CODE`, `LOCATION_NAME`) | `LocateCode, StartCalcDate, AssetsCode` | :352-355; View:130-131 |
| 5433 | ตามรหัส (`...AssetsView`) | 96004 | ไม่มี | `AssetsCode, StartCalcDate, DepartCode, AssetsType, LocateCode` | :351-355 |
| 5434 | วันเริ่มคำนวณ (`...DocDateView`) | 96005 | วันที่เริ่ม + วันที่สิ้นสุดคำนวณ | `StartCalcDate, AssetsCode` | :355-358; View:131-132 |
| 5435 | สรุปทั้งหมดเป็นเดือน (`...AbstractView`) | 96006 | ไม่มี (แสดงวันที่เริ่ม/สิ้นสุดเป็นคอลัมน์) | `AssetsCode, StartCalcDate, AssetsType, DepartCode, LocateCode` | Dlg:196; View:130-136 |

คอลัมน์ของ 5430-5434 (`...DepartView:135-151`): รหัสสินทรัพย์, ชื่อสินทรัพย์, ยอดยกมา (`BEGIN_QTY`, อนุมาน), 12 คอลัมน์ "มกราคม … ธันวาคม" (คีย์ `JANUARY`…`DECEMBER` = `AssetsPeriod1..12`), รวมทั้งสิ้น (`TOTAL_AMOUNT` "รวมทั้งสิ้น :"), ยอดยกไป (`ENDTRANSFER__BALANCE`). 5435 เพิ่มคอลัมน์วันที่เริ่ม/สิ้นสุดคำนวณก่อนยอดยกมา (`:133-136`)
กติกาสำคัญ

| หัวข้อ | กติกา Champ | หลักฐาน |
|---|---|---|
| แหล่งข้อมูล | `BCAssetsOfPeriod a JOIN BCAssetsMaster b` + **UNION** แถวจาก `BCAssetsMaster` ที่ `SaleDate IS NULL` (สินทรัพย์ที่ยังไม่เคยประมวลผล แสดง ยอดยกมา/ยอดยกไป = ราคาซาก, ค่าเสื่อมทุกเดือน 0; ส่วน UNION ไม่ถูกกรองด้วยวันที่) | Dlg:26-37,343-354 |
| ชื่อเดือน | พิมพ์ชื่อเดือนปฏิทิน แต่ข้อมูลเป็น "งวดที่ 1-12 ของปีบัญชี" — ถ้าปีบัญชีไม่เริ่ม ม.ค. หัวคอลัมน์จะไม่ตรงกับเดือนจริง | View:138-149; `Deperciate...TypeView:199` (อ่าน `BCPeriod` เฉพาะลำดับงวด) |
| ขอบเขตวันที่ | บังคับ; กรองด้วย `StartCalcDate BETWEEN` ของแถวงวด | Dlg:141-165,210 |
| ผลรวม | ท้ายกลุ่ม/ท้ายรายงาน รวมแต่ละคอลัมน์; แถวสรุปต่อหน้าตาม `IsPrintSumPage` | View:174-176 |
| ความต่างจากรายงานปี | **รายเดือนไม่คำนวณสด** — ถ้าไม่ได้ "ประมวลผลสินทรัพย์" ใหม่หลังแก้ทะเบียน ตัวเลขเก่าค้าง ขณะที่รายงานปีคำนวณทันที | §7.5 |

### 10.5 กลุ่ม "รายงานอื่น ๆ" (4 รายการ)

| menu | รายงาน | ตัวกรอง | คอลัมน์ | ยอดรวม | หลักฐาน |
|---|---|---|---|---|---|
| 5436 | สินทรัพย์ที่ถูกขายพร้อมกำไรขั้นต้น (`FARepSellItemView`, 97002) | วันที่ขายจาก-ถึง, รหัส, แผนก, ประเภท, ที่ตั้ง; `ORDER BY SaleDate, Code` | วันที่ขาย (`FA_SALEDATE` "วันที่ขาย :", ค่าที่แสดง = `SaleDocDate`), เลขที่เอกสาร (`DOCNO` ← `SaleDocNo`), รหัส, ชื่อสินทรัพย์, วันที่ซื้อ, จำนวนเงิน, ยอดยกมา, ค่าเสื่อม (`ASSETSAMOUNT`), มูลค่าคงเหลือ (`ITEMAMOUNT`), ราคาขาย (`PRICE`), กำไร(ขาดทุน) (`EXCHANGE_PROFIT`) | ท้ายรายงาน: รวมราคาขาย และ รวมกำไร(ขาดทุน) รูปแบบ `(...)` | Dlg:160-254; View:141-153,280-316,338,356 |
| 5437 | สินทรัพย์ระหว่างก่อสร้าง (`FARepAssetsBuildView`, 97003) | รหัสสินทรัพย์จาก-ถึง + ช่วงวันที่เอกสาร; ตาราง `BCAssetsOfBldSub` | หัวกลุ่ม: รหัส (`FROM_CODE`) + ชื่อ (`DESCRIPTIONNAME`); รายการ: วันที่, เลขที่เอกสาร, คำอธิบาย (`COMMENT_DESC`), จำนวนเงิน | รวม n รายการ + ยอดรวมต่อสินทรัพย์ | Dlg:26; View:141-147,165-186 |
| 5438 | การซ่อมบำรุงสินทรัพย์ (`FARepAssetsRenovateView`, 97004) | เหมือน 5437; ตาราง `BCRenovateSub` | เหมือน 5437 | เหมือน 5437 | View:111 |
| 5439 | ข้อมูลก่อน post บัญชี | — | **ไม่มี handler/คลาสในซอร์ส** (มีแต่ resource id) | — | atlas `→ ?` |

หมายเหตุ 5436: คอลัมน์ "มูลค่าคงเหลือ" ที่แสดงคำนวณใหม่จาก `GetBookValue` (ยกมา - ค่าเสื่อม) แต่ "กำไร(ขาดทุน)" = ราคาขาย - `SaleAssetsBal` ที่เก็บไว้ในทะเบียน (`View:288-296`) — สองฐานนี้ไม่จำเป็นต้องเท่ากัน

### 10.6 รายงานนอกเมนู 28 รายการ (ไม่ต้องทำ)

| รายการ | สถานะ | หลักฐาน |
|---|---|---|
| รายงานค่าเสื่อมตาม ภ.ง.ด.50 ประจำเดือน (id 5440 `ID_FASYS_REPMONT50`, คีย์ 95009, คลาส `CFARep50MonthView` และ `CFARep50MonthYearView`) | **ปิดอยู่** — การผูกคำสั่งใน `MainFrm.cpp:1011` ถูกคอมเมนต์ และตัว handler อยู่ใต้ `#ifdef FIRM_PROJECT` (รุ่นลูกค้าเฉพาะ); ไม่มีในเมนู 28 รายการ | `C\MainFrm.cpp:1011,10026-10077`; `RG\FARep50MonthView.cpp:30` |

---

## 11. จุดแปลก/ข้อควรระวังของ Champ (BC ไม่ต้องลอกบั๊ก แต่ต้องตัดสินใจ)

| รหัส | เรื่อง | รายละเอียด | หลักฐาน | ข้อเสนอสำหรับ BC |
|---|---|---|---|---|
| Q1 | ค่าเริ่มต้น "หักค่าซากก่อน" | ฮาร์ดโค้ด = 1 ที่ engine (`m_nFirstScrap = 1`) และการกลับค่าตอนประมวลผล/รายงาน (§7.4) ทำให้ค่าเสื่อมจากปุ่ม Calc ต่างจากที่ประมวลผล | `Depreciation.cpp:33`; §7.4 | ใช้ความหมายเดียว (ฐานหรือท้ายตาราง) ทั้งระบบ — ให้ลุงจืดเลือก |
| Q2 | ประมวลผลตั้ง `ProcessType` ให้สินทรัพย์ทุกตัว | `UPDATE BCAssetsMaster SET ProcessType` ไม่กรองช่วงรหัส | `FAFrmAssetsProcessDlg.cpp:1373-1374` | เก็บ ProcessType ต่อสินทรัพย์เฉพาะที่ประมวลผล หรือเป็นพารามิเตอร์ของรอบประมวลผล |
| Q3 | log ผู้ใช้ผิดเมนู | บันทึกด้วย `ID_BILLSYS_TRANS_6` | `:1376-1396` | log ด้วยรหัสเมนูของตัวเอง |
| Q4 | `SumOfAssetsBal` (ค่าเสื่อมสะสมยกมา) เก็บแต่ไม่ถูกนำไปคำนวณ | ผู้ใช้กรอกแล้วไม่มีผลต่อค่าเสื่อมที่ประมวลผล (ใช้เฉพาะ `AssetsBalance`); โค้ดบวกค่านี้เข้าค่าเสื่อมสะสมถูกคอมเมนต์ปิดไว้ | `Depreciation.cpp` ไม่มีการอ้าง `SumOfAssetsBal`; `FAFrmAssetsProcessDlg.cpp:918-921,942` (คอมเมนต์) | แสดงฟิลด์นี้เป็นข้อมูลอ้างอิง และ/หรือใช้ตรวจสอบ ไม่ใช่ตัวแปรคำนวณ (ต้องถามลุงจืด) |
| Q5 | GL transfer ใช้ยอดจากช่องเดบิตเท่านั้น | แก้ช่องเครดิตมีผลแค่ตรวจดุล | `FSTransferToGLDlg.cpp:386-415,501-538` | BC: ให้แก้ยอดได้ที่เดียว เครดิตคำนวณตามเดบิต |
| Q6 | ข้อความผิด/อังกฤษตายตัว | "Please enter assetsType" ใช้ซ้ำกับปี; "Can't process,Haven't data."; ปุ่ม OK/Cancel/Process/Save | `FSTransferToGLDlg.cpp:178-193,624-628` | ใช้ข้อความไทยผ่าน key (กฎ i18n ของ BC) |
| Q7 | กล่อง filter รายงานขายใช้คอลัมน์ผิด | `SaleDates` (ไม่มีในตาราง) เมื่อกรอกวันเดียว | `FARepSellItemDlg.cpp:165` | ใช้ `SaleDate = ...` |
| Q8 | หน้าซ่อมบำรุง (Renovate) ลบข้อมูลเดิมก่อนตรวจผ่าน | ถ้า validate ไม่ผ่านหลังลบ รายการหาย | `FAFrmAssetsRenovate.cpp` (ConfirmSaveData) — ยังไม่ verify เลขบรรทัด | ตรวจก่อน แล้วค่อยลบ/แทนในทรานแซกชัน |
| Q9 | ปีปฏิทินฮาร์ดโค้ด | ค่าเริ่มต้นรายงานปีและรายงานเดือน, `ProcessAmount` ของ GL transfer ใช้ 1 ม.ค. เป็นต้นปี ทั้งที่ engine รองรับปีบัญชีข้ามปี | `FAsRepAssetsOrderCodeDlg.cpp:91-92`; `FSTransferToGLDlg.cpp:796` | ใช้งวดบัญชี `BCPeriod` ของปีบัญชีเดียวกันทุกที่ |
| Q10 | สองเอนจิน | ดู §7.5 | | เลือกสูตรเดียว |
| Q11 | ตั้งค่าอายุ 20 ปี (F5 จากงานก่อสร้าง) | ฮาร์ดโค้ด | §3.3 | เป็นค่าจากประเภทสินทรัพย์ |
| Q12 | ลบสินทรัพย์ไม่มีตัวกัน; เปลี่ยนรหัสไม่ตามไปที่ Build/Renovate | ดู §3.6 | | BC: บล็อกลบเมื่อมีรายการอ้างอิง (กฎของ BC เองที่ใช้กับผังบัญชี) — ยังไม่ใช่การลอก Champ |
| Q13 | เมนู 5439 "ข้อมูลก่อน post บัญชี" ไม่มีโค้ด | | atlas | ไม่ทำ |
| Q14 | หัวคอลัมน์รายงานเดือนใช้ชื่อเดือนปฏิทิน | §10.4 | | ใช้ชื่อ "งวดที่ n (ช่วงวันที่)" |

---

## 12. BC เทียบ Champ และสิ่งที่ต้องปรับ

คำชี้แจงขอบเขต: AGENTS.md กำหนดว่าตอนนี้ลุงจืดเน้น **GL ตัวเดียว** (ระบบ ERP อื่นรอ) — ตารางนี้บอก "ถ้าจะทำสินทรัพย์ตาม Champ" ไม่ใช่คำสั่งให้เริ่มทำ. สถานะ: **เหมือน** = พฤติกรรมตรงแล้ว, **ต่าง** = มีแต่ต่างจาก Champ, **ยังไม่มี** = ไม่มีโค้ด/หน้าจอ, **รอพัฒนา** = มีเมนู แต่หน้าจอ/ API ยังไม่เสร็จ (`menu-screen-status.ts`)

หลักฐานฝั่ง BC ที่ใช้ตลอดตาราง

| ที่ | ข้อเท็จจริง |
|---|---|
| `fe\lib\menu-data.ts:558-592` | กลุ่มเมนู FA 11 รายการ (ทะเบียน, ซ่อมบำรุง, ระหว่างก่อสร้าง, ประเภท, ประมวลผล, โอน GL, ตารางสินทรัพย์, รายเดือน, รายปี, ภ.ง.ด.50, ขาย) |
| `fe\lib\menu-screen-status.ts:19-28` | `/asset/*` และ `/report/assetschedule` ถูกนับเป็น "พร้อม" |
| `fe\app\asset\fixed-assets-screen.tsx:113-117` | route `/asset/types` ตั้งแท็บ "types" แต่ไม่มีบล็อก render ของแท็บนี้; `/asset/cip`, `/asset/maintenance` ไม่เปลี่ยนแท็บ — **จอที่ตกเป็น "พร้อม" แต่ไม่มีเนื้อหา** (ยืนยันจากโค้ด ยังไม่ได้เปิดดูบนเบราว์เซอร์) |
| `fe\lib\menu-screen-status.test.ts:15-24` | `/report/depreciationmonthly`, `yearly`, `pnd50`, `assetdisposal` อยู่ในรายการ "ยังไม่มีจอ/ API" |
| `be\models.go:31-82` | `AssetType` (3 บัญชี + อายุ/อัตราเริ่มต้น), `Asset` ~28 ฟิลด์ (มี `Method` แต่ `calculator.go` ไม่อ้างถึง) |
| `be\calculator.go:34-157` | เส้นตรงรายเดือนปฏิทินเท่านั้น |
| `be\gl_poster.go:239,562` | `PostDepreciation(fiscalYear, period, ...)`; `DisposeAsset` ลง GL พร้อมกำไร/ขาดทุน |
| `be\store.go` | เก็บใน `fa_records` / `fa_schedule_posted` (รูปแบบเอกสาร JSON — ขัดกฎคอลัมน์ snake_case ของ AGENTS.md ที่ให้ทุกฟิลด์เป็นคอลัมน์ lowercase) |
| `be\reports.go:234,380` | รายงานตารางสินทรัพย์ และรายงานกระทบยอดภาษี (ไม่มีใน Champ) |

| menu Champ | BC menu id / route | สถานะ | สิ่งที่ต้องปรับ |
|---|---|---|---|
| 5441 รายละเอียดสินทรัพย์ | `asset-registry` `/asset/registry` | ต่าง | ใช้ลำดับ/ป้าย 37 ฟิลด์ของ §3.2 (กลุ่ม หัว / การคิดค่าเสื่อมราคา / การขาย / การดูแลรักษา), เพิ่มหน่วยนับ, เอกสารซื้อ (เลขที่/วันที่/ซื้อจาก), ราคาที่คิดค่าเสื่อม, มูลค่าคงเหลือยกมา, ค่าเสื่อมสะสมยกมา, ราคาซาก, หักค่าเสื่อมเบื้องต้นปีแรก %, วันหยุดคำนวณ, วิธีคิด 0/1/2, หักซากก่อน, ประกัน, ซ่อมบำรุง; ปุ่ม F5 ดึงงานก่อสร้าง, Calc, F6 ตารางค่าเสื่อม; ไม่มี "สถานะเอกสาร" ในทะเบียน; ฟิลด์ BC ที่ Champ ไม่มี (ยี่ห้อ/รุ่น/ซีเรียล/ผู้ขาย/รถยนต์นั่ง/ลดหย่อนภาษี) ให้ลุงจืดตัดสินว่าจะเก็บหรือตัด |
| 5442 บันทึกรายการซ่อมบำรุง | `asset-maintenance` `/asset/maintenance` | รอพัฒนา (ถูกนับ "พร้อม" แต่ไม่มีจอ) | สร้างจอตาม §5 (หัว: รหัสสินทรัพย์ + ชื่อ, กริด: วันที่, เลขที่เอกสาร, คำอธิบาย, จำนวนเงิน; ท้าย: รวมทั้งสิ้น); บันทึกแบบลบแล้วใส่ใหม่ในทรานแซกชันเดียว (ตรวจก่อน — แก้ Q8) |
| 5443 บันทึกสินทรัพย์ระหว่างก่อสร้าง | `asset-construction-in-progress` `/asset/cip` | รอพัฒนา | เช่นเดียวกับ 5442 ตาราง `BCAssetsOfBldSub`/`BCAssetsOfBuild`; ผูกกับ F5 ของทะเบียน |
| 5873 กำหนดประเภทสินทรัพย์ | `asset-type` `/asset/types` | ต่าง/รอพัฒนา | Champ: 4 คอลัมน์ (รหัส, ชื่อ, **เดบิต = ค่าเสื่อมราคา**, **เครดิต = ค่าเสื่อมสะสม**); BC มี 3 บัญชีในโมเดลแต่ไม่มีจอ — ทำจอ 4 คอลัมน์; บัญชีที่สามของ BC (ต้นทุนสินทรัพย์) ให้ลุงจืดตัดสินว่าจำเป็นหรือไม่ |
| 5444 ประมวลผลสินทรัพย์ | `asset-depreciation` `/asset/depreciation` | ต่าง | Champ: ไดอะล็อก เริ่มจาก/ถึงสินทรัพย์ + "คำนวณเป็นรายวัน" → ลบและสร้าง `OfYear/OfPeriod` ใหม่ ไม่ลง GL; BC: คำนวณเส้นตรงตามเดือนปฏิทิน (`becalculator.go:34-157`) แล้ว `PostDepreciation(fiscalYear, period)` โพสต์ GL ต่องวด (`begl_poster.go:239`) — ไม่มีขั้นเก็บผลรายปี/รายงวดแยกก่อนโอน; ต้องแยกเป็น 2 ขั้น (ประมวลผล, แล้วโอน) ตาม §6/§8 และรองรับ 3 วิธี + ProcessType + งวดปีบัญชี |
| 5491 โอนข้อมูลเข้าสู่ GL | `asset-post-gl` `/asset/post-gl` | ต่าง | ใช้ตัวกรอง ช่วงรหัส/ประเภท/แผนก/ตั้งแต่งวด-ถึงงวด/ปี/แยกตามแผนก + สมุดรายวัน + เลขที่ + วันที่ + คำอธิบาย; ตัวอย่างกริดแก้ไขได้ก่อนบันทึก; ยอด = Σ `AssetsPeriod[k]` ที่เก็บไว้; บันทึกลง GL เป็นเอกสารเดียวต่อรอบ (ในระบบ BC ใช้สมุดรายวัน/สถานะร่าง-ผ่านรายการของ GL เอง); ตรวจงวดปิด + เลขที่ซ้ำ |
| 5424-5429 รายละเอียดสินทรัพย์ 6 รายงาน | ไม่มี | ยังไม่มี | ทำตาม §10.2 (รายงานเดียวที่เลือกเรียง/จัดกลุ่มได้ หรือ 6 เมนูตาม Champ — ตามกฎ "ยุบซ้ำซ้อนได้" ของ AGENTS.md ข้อเสนอ: 1 จอ + ตัวเลือกการจัดกลุ่ม ให้ลุงจืดตัดสิน) |
| 5445-5448 รายปี 4 รายงาน | `depreciation-yearly-report` `/report/depreciationyearly` | รอพัฒนา | §10.3 คอลัมน์ 11 ช่อง; ตัวกรอง 5 ช่วง + วันที่ + "คำนวณเป็นวัน"; คำนวณสดต่อสินทรัพย์ |
| 5449 สรุปค่าเสื่อมทั้งหมดเป็นปี | `fixed-asset-schedule` `/report/assetschedule` | ต่าง | BC อ่านจากตารางที่ BC สร้างเอง; Champ อ่าน `BCAssetsOfYear` รายสินทรัพย์ × ปี — ปรับคอลัมน์ให้ตรง §10.3 |
| 5450 ค่าเสื่อมตาม ภ.ง.ด.50 | `depreciation-pnd50-report` `/report/depreciationpnd50` | รอพัฒนา | ทำเป็นรายงานปีจัดกลุ่มตามประเภท (เรียง `AssetsType, Code`) ตามต้นฉบับ; **อย่าเพิ่มรายงานกระทบยอดภาษีของ BC เข้าไปในเมนูนี้** (`be\reports.go:380` ไม่มีใน Champ — รอลุงจืดตัดสิน) |
| 5430-5435 รายเดือน 6 รายงาน | `depreciation-monthly-report` `/report/depreciationmonthly` | รอพัฒนา | §10.4 : อ่านค่าที่ประมวลผลเก็บไว้ 12 งวด + UNION สินทรัพย์ที่ยังไม่ประมวลผล; หัวคอลัมน์ใช้ "งวดที่ n" ตามปีบัญชี (Q14) |
| 5436 สินทรัพย์ที่ถูกขายพร้อมกำไรขั้นต้น | `asset-disposal-report` `/report/assetdisposal` | รอพัฒนา | §10.5 ตัวกรองวันที่ขาย + 4 ช่วง; ใช้ฐานมูลค่าคงเหลือเดียวกันทั้งคอลัมน์แสดงและกำไร (Champ ใช้สองฐาน — §10.5) |
| 5437-5438 ก่อสร้าง / ซ่อมบำรุง | ไม่มี | ยังไม่มี | §10.5 |
| 5439 ข้อมูลก่อน post บัญชี | ไม่มี | ไม่ทำ | Champ เองไม่มีโค้ด |
| (นอกเมนู) การตัดจำหน่ายลง GL | `be\gl_poster.go:562` `DisposeAsset` | **BC มากกว่า Champ** | Champ ไม่ลง GL ตอนขาย (§9) — ตามกฎ Strict Parity ต้องให้ลุงจืดตัดสินว่าจะคงไว้ (เป็นประโยชน์ทางบัญชี แต่ไม่ใช่ของ Champ) |
| (นอกเมนู) รายงานกระทบยอดภาษีสินทรัพย์ | `be\reports.go:380` | **BC มากกว่า Champ** | เช่นเดียวกัน |

ประเด็นที่ชนกับกฎของลุงจืด (ไม่เสนอให้ทำ)

| ประเด็น | กฎ | ผล |
|---|---|---|
| ระบบสินทรัพย์ทั้งชุด | AGENTS.md: ตอนนี้เน้น GL; ระบบอื่นรอ | ตารางนี้เป็นพิมพ์เขียวไว้ก่อน ไม่เริ่มทำเอง |
| `fa_records` เก็บเป็นเอกสาร JSON | กฎชื่อคอลัมน์ PostgreSQL lowercase + ตัวเลขบัญชีเป็น decimal (Rule 20) | ถ้าจะทำจริง ต้องแปลงเป็นตารางคอลัมน์ `numeric(p,s)` (Champ ใช้ MONEY ของ SQL Server — ห้ามลอกชนิดเป็น float) |
| ข้อความ/ป้ายไทยของ Champ | ใช้ถ้อยคำเดิมของ Champ (memory 2026-09-16) | ป้ายใน §3.2/§10 ที่ระบุ "(อนุมาน)" ต้องให้ลุงจืดยืนยันก่อนใส่ `languages.tsv` |
| ระบบเงินเดือน | ไม่ทำ | ไม่เกี่ยวกับสินทรัพย์ (ไม่มีรายการในโมดูลนี้) |
| ชื่อคู่แข่ง | ห้ามอ้างอิง | ไม่มีในเอกสารนี้ |

---

## 13. สิ่งที่ยังไม่ verify

| # | เรื่อง | เหตุผล | วิธีปิด |
|---|---|---|---|
| 1 | ข้อความไทยของคีย์ที่ต้องระบุ "(อนุมาน)" (หัวคอลัมน์รายงานปี/เดือน, ป้ายไดอะล็อก GL transfer บางตัว) | ตาราง `myresource` ไม่อยู่ใน repo; `.rc` บางป้ายมีรหัสภาษาเสีย | ขอไฟล์/ตาราง `myresource` ของ Champ จากลุงจืด หรือดูจากหน้าจอจริง |
| 2 | พฤติกรรมกริดของไดอะล็อกประเภทสินทรัพย์ (`CMyDlgMasterCode`) | มีเฉพาะ `.obj` ไม่มีซอร์ส | เปิดโปรแกรม Champ จริง |
| 3 | ผลต่างตัวเลขระหว่างเอนจินฝั่งจอกับฝั่งรายงาน | ไม่ได้รันโปรแกรม | รันชุดข้อมูลเดียวกันทั้งสองทางแล้วเทียบ (§7.5) |
| 4 | การใช้ checkbox "คำนวณเป็นวัน" ของกล่อง filter รายงานปี | พบแค่ DDX (`FAsRepAssetsOrderCodeDlg.cpp:61`) ไม่พบการอ่านค่าใน View | ติดตามค่า `m_CalcByDay` |
| 5 | เนื้อหาการหาช่วงปีใน engine (`Depreciation.cpp:369-460`) และบรรทัดประกาศ `nAgeDec` | ไม่ได้ไล่ทุกกิ่ง | อ่านต่อหากจะลอกสูตรปีบัญชีข้ามปี |
| 6 | ฟังก์ชันเพิ่มของเอนจินรายงาน `GetDepreciatePeriod/GetBookValue2/GetDepreciatePeriod2` | อ่านเฉพาะส่วนหัว | อ่านต่อหากรายงานที่เรียกใช้จะถูกลอก |
| 7 | `FAFrmAssetsRenovate` ลำดับลบ/ตรวจ (Q8) และการผูก `FAFrmAssetsBuild` | อ่านผ่านสรุป | ยืนยันเลขบรรทัด |
| 8 | จอ `/asset/types`, `/asset/cip`, `/asset/maintenance` ของ BC ว่างจริงบนเบราว์เซอร์ | ยืนยันจากซอร์สเท่านั้น | เปิดผ่านปุ่ม Demo |
| 9 | ยอดรวมท้ายกลุ่มของรายงาน `...ByDate` (§10.2) | พบโค้ด 4-9 เกินคอลัมน์ | พิมพ์รายงานจริง |
| 10 | ตารางสิทธิ์/รหัสเมนู log ของรายงานแต่ละตัว | ไม่เกี่ยวกับเลย์เอาต์ | ไม่จำเป็นสำหรับ BC |
