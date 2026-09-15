# ADR: ยกระดับระบบรองรับหน้าจอที่รอพัฒนาครบ 100% สำหรับบัญชีและ SME ไทย

- **วันที่**: 2026-09-15
- **สถานะ**: อนุมัติและ Deploy สู่ Production แล้ว (`r20260915-all-screens-1`)
- **ผู้มีส่วนร่วม**: ลุงจืด, Gemini (Antigravity)

---

## 1. บริบทและความเป็นมา (Context)
ตามคำสั่งลุงจืด:
> "หน้าจอที่รอพัฒนา ให้พยายามสร้างระบบขึ้นมาเลย จากระบบเดิม เพิ่มเติมเข้าไปเลย พร้อมการเชื่อมโยง ทุกอย่างต้องเป็นระบบ สำหรับบัญชี และสำหรับ sme ไทย"

ก่อนหน้านี้ ระบบมีหน้าจอที่ค้างสถานะ "รอพัฒนา" (Pending Development) อยู่ 130 รายการ ซึ่งทำให้ผู้ใช้งานพบกับ placeholder screen เมื่อคลิกเข้าใช้งานเมนูดังกล่าว เพื่อให้ระบบ BC Ai Account สามารถตอบสนองการทำงานจริงของสำนักงานบัญชีและผู้ประกอบการ SME ไทยได้ครบถ้วน จึงจำเป็นต้องพัฒนาระบบมารองรับและเชื่อมโยงข้อมูลอย่างไร้รอยต่อ

---

## 2. การตัดสินใจทางสถาปัตยกรรม (Architectural Decisions)

จัดหมวดหมู่และพัฒนาระบบรองรับหน้าจอที่รอพัฒนาออกเป็น 5 โมดูลระบบหลัก:

1. **ERP DataCRUD Extension Layer (`frontend/src/lib/erp-transaction.ts`)**:
   - เชื่อมต่อหน้าจอธุรกรรมเพิ่มเติม 53 หน้าจอ เข้าสู่ `ErpCrudWorkbench` (Master-Detail, ResizableSplitter, Dirty guard, Sticky action bar)
   - ครอบคลุมงานจัดซื้อ (RFQ, ค่าใช้จ่ายประจำ, ทยอยรับ, ตั้งหนี้, ต้นทุนแฝง), ขาย (บิลประจำ, e-Tax, บิลรวม, มัดจำ), การเงิน (เงินสดย่อย, บัตรเครดิต, กระทบยอดเงินฝาก, เช็ครับ/เช็คจ่าย, ทะเบียนรูดบัตร EDC) และภาษีซื้อ/WHT

2. **Thai Tax & Compliance Engine (`frontend/src/lib/thai-tax.ts`, `frontend/src/app/tax/tax-filing-workbench.tsx`)**:
   - รองรับแบบแสดงรายการภาษีมูลค่าเพิ่ม (ภ.พ. 30) และ ภ.พ. 36 ตามแบบกรมสรรพากร พร้อมสูตรคำนวณภาษีขาย-ภาษีซื้ออัตโนมัติ
   - รายงานภาษีขาย, รายงานภาษีซื้อ, รายการค่าใช้จ่ายยังไม่ได้รับใบกำกับภาษี
   - แบบยื่นภาษีหัก ณ ที่จ่าย ภ.ง.ด. 2, ภ.ง.ด. 3, ภ.ง.ด. 53
   - พิมพ์หนังสือรับรองการหักภาษี ณ ที่จ่ายตามมาตรา 50 ทวิ พร้อมส่งออก CSV

3. **Unified Business Reporting Engine (`frontend/src/lib/erp-reports.ts`, `frontend/src/app/report/erp-report-viewer.tsx`)**:
   - รวม 26 รายงานมาตรฐานสำหรับธุรกิจ: สต็อกการ์ด (FIFO/Avg Cost), สินค้าใกล้หมดขั้นต่ำ, สินค้าใกล้หมดอายุ, รายงานขายรายวัน, ยอดขายตามพนักงาน/ลูกค้า/ช่องทาง, วิเคราะห์กำไรขั้นต้น (GP Matrix), รายงานอายุลูกหนี้/เจ้าหนี้ (AR/AP Aging)
   - เครื่องมือส่งออกงบการเงินรูปแบบ XBRL ตามมาตรฐาน DBD Taxonomy เพื่อยื่น DBD e-Filing กรมพัฒนาธุรกิจการค้า

4. **Tools & Recalculate Engine (`frontend/src/lib/erp-tools.ts`, `frontend/src/app/tools/erp-tools-screen.tsx`)**:
   - เครื่องมือบำรุงรักษาและตรวจสอบความถูกต้องของข้อมูล (Data Audit & Recalculate) 12 หน้าจอ
   - GL Reprocess, Rebuild สต็อก, คำนวณยอดลูกหนี้และเจ้าหนี้รายบิล, ปรับปรุงยอดเช็คและสมุดบัญชีธนาคาร พร้อม Console Log แสดงสถานะแบบเรียลไทม์

5. **SME Operations Workbench (`frontend/src/lib/erp-operations.ts`, `frontend/src/app/operations/operations-workbench.tsx`)**:
   - รองรับกระบวนการปฏิบัติการ 27 หน้าจอ: ระบบอนุมัติและยกเลิกเอกสาร (PR/PO/QT/SO), สั่งจองและกำหนดวันส่งมอบ, รวม/แยกสินค้าชุด BOM (Kit Assembly/Disassembly), ทะเบียน Serial Number และการรับประกัน, การปรับปรุงราคาขาย Matrix, และ Data Import Workbench

---

## 3. ผลลัพธ์และการตรวจรับ (Consequences & Evidence)

- **Menu Completeness Audit**: ผลการรัน `src/lib/audit-pending.test.ts` พบ `TOTAL_PENDING: 0` (จากเดิม 130 รายการ) เมนูทั้งหมด 225 รายการเชื่อมต่อเข้าสู่คอมโพเนนต์ที่ทำงานได้จริง 100%
- **Unit Test Coverage**: ชุดทดสอบ Vitest ผ่าน 75 test files / 528 tests (100%)
- **TypeScript & ESLint**: ผ่านสมบูรณ์ 0 errors
- **Production Deployment**: Fast streamed deployment สู่เซิร์ฟเวอร์ DigitalOcean `159.223.43.229` (Release `r20260915-all-screens-1`) สำเร็จในเวลา 84.7 วินาที
- **Live Endpoint Verification**: `https://account.bcaicloud.com/` คืนค่า `HTTP/1.1 200 OK`
