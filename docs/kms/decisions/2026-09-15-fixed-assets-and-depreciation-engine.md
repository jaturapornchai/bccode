# การตัดสินใจ: ระบบบริหารสินทรัพย์ถาวรและการคำนวณค่าเสื่อมราคา (Fixed Assets & Depreciation Engine) ตามต้นแบบ Champ

- **วันที่**: 2026-09-15
- **สถานะ**: อนุมัติและ Deploy สู่ Production แล้ว (`account.bcaicloud.com` release `r20260915-fa-1`)
- **ผู้มีส่วนร่วม**: ลุงจืด, Antigravity AI
- **คำสั่งต้นทาง**: `/goal ทำแบบเดียวกัน กับ ระบบสินทรัพย์ และค่าเสื่อมราคาเลย`

---

## 1. บริบทและปัญหา (Context & Problem)

หลังจากที่ระบบบัญชีแยกประเภท (General Ledger) ได้รับการทดสอบและพัฒนาตามวงจรบัญชีมาตรฐานไทยและต้นแบบ `D:\project-champ` จนเสร็จสมบูรณ์ พร้อมเปิดให้สำนักงานบัญชีใช้งานได้จริง ลุงจืดได้สั่งการให้ดำเนินการในลักษณะเดียวกันกับ **"ระบบสินทรัพย์ถาวรและค่าเสื่อมราคา" (Fixed Assets & Depreciation)**

โดยระบบสินทรัพย์ถาวรเดิมยังขาดส่วนการคำนวณค่าเสื่อมราคารายงวดอัตโนมัติ การลงบัญชีแยกประเภทสมุดรายวันทั่วไป (JV) อัตโนมัติ การบันทึกการจำหน่ายสินทรัพย์พร้อมคำนวณกำไร/ขาดทุน และรายงานทะเบียนสินทรัพย์รวมถึงรายงานกระทบยอดทางภาษีตามประมวลรัษฎากร (ภ.ง.ด.50)

---

## 2. การศึกษาต้นแบบจาก Champ (`D:\project-champ`)

จากการวิเคราะห์โค้ดต้นแบบใน `D:\project-champ`:
1. **โครงสร้างข้อมูลสินทรัพย์**:
   - `BCAssetsMaster`: เก็บข้อมูลหลัก ราคาทุน วันที่เริ่มคิดค่าเสื่อม อายุการใช้งาน อัตราค่าเสื่อมราคา ราคาซาก และผังบัญชีที่เกี่ยวข้อง (สินทรัพย์, ค่าเสื่อมราคาสะสม, ค่าเสื่อมราคาประจำงวด)
   - `BCAssetsOfYear` / `BCAssetsOfPeriod`: เก็บรวบรวมค่าเสื่อมราคารายปีและรายงวดเดือน (Period 1-12)
2. **เครื่องคำนวณค่าเสื่อมราคา (`CDepreciation`)**:
   - ใช้วิธีคิดค่าเสื่อมเส้นตรง (Straight-Line Method) คำนวณตามจำนวนวันจริงในแต่ละงวดเดือน (`daysInMonth` / `daysInYear`)
   - รองรับปีอธิกสุรทิน (Leap Year 366 วัน)
   - รองรับสิทธิประโยชน์ทางภาษีปีแรก (First-Year Initial Tax Allowance) เช่น อุปกรณ์คอมพิวเตอร์หักได้ 40% ในงวดแรก
   - มีการล็อกเพดานมูลค่าซาก (Salvage/Scrap Value Cap) ไม่ให้คิดค่าเสื่อมราคาจนมูลค่าตามบัญชีต่ำกว่าราคาซาก (ปกติเหลือ 1 บาท)

---

## 3. การตัดสินใจทางสถาปัตยกรรม (Architectural Decisions)

### 3.1 สถาปัตยกรรม 2-Tier Database (MongoDB + PostgreSQL)
- **MongoDB (`appdb`)**: ทำหน้าที่เป็น Storage Layer หลักสำหรับเก็บ Document:
  - `fixed_assets`: ทะเบียนสินทรัพย์ถาวร
  - `asset_types`: กลุ่ม/ประเภทสินทรัพย์ถาวร
  - `asset_depreciations`: ตารางค่าเสื่อมราคารายเดือน (Depreciation Schedule Items)
  - `asset_disposals`: ประวัติการขาย/ตัดจำหน่ายสินทรัพย์ (Gain/Loss on Disposal)
  - `gl_journals`: ใบสำคัญสมุดรายวันทั่วไป (JV)
- **PostgreSQL (`gl_lines`)**: ทำหน้าที่เป็น Processing Engine สำหรับรายงานทางการเงิน โดยเมื่อมีการ Post ค่าเสื่อมราคาหรือจำหน่ายสินทรัพย์ ระบบจะฉาย (project) รายการเข้าตาราง `gl_lines` ของ Holding ทันที

### 3.2 ความแม่นยำทางการเงิน (Exact Precision via `Amount`)
- สร้างประเภทข้อมูล `fixedasset.Amount` ที่ทำงานครอบคลุม BSON Type `Decimal128`, `Double`, `String`, `Int32`, `Int64` เหมือนโมดูล GL
- ขจัดปัญหา Floating Point Precision Error ทั้งหมดในการคำนวณค่าเสื่อมราคาและยอดสมดุลเดบิต-เครดิต (Debit == Credit)

### 3.3 การผ่านรายการเข้า GL อัตโนมัติ (Automated Balanced GL Posting)
- **ผ่านรายการค่าเสื่อมราคารายงวด**:
  - `Dr. ค่าใช้จ่ายค่าเสื่อมราคา (520103)`
  - `Cr. ค่าเสื่อมราคาสะสม (129101)`
  - รองรับการจัดกลุ่มตามประเภทบัญชีและสาขา
  - รองรับการยกเลิก/กลับรายการ (Reverse Depreciation) เพื่อคืนสถานะให้สามารถคำนวณใหม่ได้
- **การจำหน่ายสินทรัพย์ (Asset Disposal)**:
  - `Dr. เงินสด/ลูกหนี้ (ราคาขาย + VAT)`
  - `Dr. ค่าเสื่อมราคาสะสม (ตัดออกทั้งหมด)`
  - `Dr. ขาดทุนจากการจำหน่ายสินทรัพย์ (กรณีขาดทุน)`
  - `Cr. ราคาทุนสินทรัพย์ (ตัดราคาทุนเดิม)`
  - `Cr. กำไรจากการจำหน่ายสินทรัพย์ (กรณีกำไร)`
  - `Cr. ภาษีขาย (VAT 7% ถ้ามี)`

### 3.4 รายงานมาตรฐานสำหรับนักบัญชีและผู้สอบบัญชีไทย
1. **รายงานตารางสินทรัพย์และค่าเสื่อมราคา (Fixed Asset Schedule Report)**:
   - แสดง ราคาทุนยกมา + ซื้อเพิ่ม - จำหน่าย = ราคาทุนยกไป
   - แสดง ค่าเสื่อมยกมา + ค่าเสื่อมงวดนี้ - ค่าเสื่อมจำหน่าย = ค่าเสื่อมสะสมยกไป
   - มูลค่าตามบัญชียกไป (Net Book Value)
2. **รายงานกระทบยอดค่าเสื่อมราคาทางบัญชีและภาษีอากร (ภ.ง.ด.50)**:
   - ตรวจจับยานพาหนะนั่งไม่เกิน 10 ที่นั่ง (PASSENGER_CAR) ที่จำกัดมูลค่าทางภาษีไม่เกิน 1,000,000 บาท ตาม พ.ร.ฎ. 315
   - แสดงผลต่างเพื่อใช้ปรับปรุงกำไรสุทธิทางภาษีอากรในแบบ ภ.ง.ด.50

### 3.5 AI & MCP Tools Integration
เปิดใช้งาน Model Context Protocol (MCP) เครื่องมือ 6 รายการที่ `/fa/v2/mcp` เพื่อให้ AI ภายนอกสามารถเรียกทำงานได้:
1. `fa_list_assets`
2. `fa_create_asset`
3. `fa_get_schedule`
4. `fa_post_depreciation_to_gl`
5. `fa_dispose_asset`
6. `fa_get_schedule_report`

### 3.6 Frontend Workbench
สร้างหน้าจอครบวงจรที่ `frontend/src/app/asset/fixed-assets-screen.tsx` รองรับ 6 แถบงาน:
1. ทะเบียนสินทรัพย์ (Asset Registry & Search)
2. ตารางค่าเสื่อมราคารายงวด (Depreciation Schedule)
3. ผ่านรายการเข้าบัญชีแยกประเภท (Post to GL)
4. การจำหน่ายสินทรัพย์ (Asset Disposal Workbench)
5. รายงานตารางสินทรัพย์ (Fixed Asset Schedule Report)
6. รายงานภาษี ภ.ง.ด.50 (Tax Reconciliation Report)
พร้อมสอดคล้องกับกฎ UX/UI ผู้ใช้คนไทยอายุ 40+ และระบบ 12 ภาษา

---

## 4. ผลการทดสอบและการตรวจรับ (Verification & Evidence)

1. **Go Tests**:
   - `TestCalculator_StraightLineStandard`: PASS
   - `TestCalculator_FirstYearSpecialAllowance`: PASS
   - `TestFixedAssets_CPACycle` (Create Asset, Calculate Schedule, Post Depreciation to GL, Dispose Asset with Gain): PASS 100%
2. **Frontend Tests**:
   - Vitest: 68 test files passed / 507 tests passed (100%)
3. **Static Analysis & Compilation**:
   - TypeScript `tsc --noEmit`: 0 errors
   - ESLint: 0 errors
   - Next.js 16 Production Build (Turbopack): สำเร็จ 100%
4. **Fast Streamed Zero-Disk Deployment**:
   - สำเร็จใน 150.7 วินาที
   - อัปเดต production server `159.223.43.229` สู่ release `r20260915-fa-1`
   - ตรวจสอบ Live health check บน `https://account.bcaicloud.com/` เรียบร้อย
