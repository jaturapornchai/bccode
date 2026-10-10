# HANDOFF 2026-10-11 — ตรวจสอบและพัฒนาหน้าจอ รายงาน เท่ากับ BC (Champ) พร้อมทดสอบ Full-Stack CRUD (Antigravity → Claude)

> **ผู้รับงาน:** Claude (และ AI ทุกตัวที่รับช่วงต่อ)  
> **คำสั่งของลุงจืด:** "handoff เดี๋ยวให้ claude ตรวจสอบ เพื่อให้ หน้าจอ รายงาน เท่ากับ bc (champ) พร้อมทดสอบ front end backend database (CRUD) ให้ครบด้วย"  
> **Single Source of Truth:** `D:\project-champ` (Champ 5.5 Pro) และภาพหน้าจอ dropdown/dialog จริงใน `D:\project-champ\bcspec`  
> **กฎเหล็กที่ห้ามฝ่าฝืน:** อ่าน `AGENTS.md` และ `~/.agents/AGENTS.md` (เรียก "ลุงจืด" เสมอ, ภาษาไทยสั้นกระชับ, `mydocs/` Read-Only เด็ดขาด [Rule 18], บัญชีห้าม float [Rule 20], ห้าม hardcode secret [Rule 21], Verify ก่อน Done [Rule 2])

---

## 1. สถานะปัจจุบันของระบบ (System Baseline ณ 2026-10-11)

- **Git Branch:** `dev` (HEAD: `86e5392a`) — นำหน้า `origin/dev` อยู่ 26 commits (**ห้าม push เอง** รอคำสั่งลุงจืดตาม Rule 19)
- **Production Server:** `159.223.43.229` (Release ปัจจุบัน: `r20261010-phase8-parity-15` ผ่าน `tools/fast-deploy.py`)
- **Production URL:** [https://account.bcaicloud.com](https://account.bcaicloud.com)
- **Live Test Credentials:** ปุ่ม "ทดลองใช้ระบบ (Demo)" -> กลุ่ม `rungrueng` -> บริษัท `01` -> สาขา `00000` (สิทธิ์ `["*"]`)
- **สิ่งที่เพิ่งส่งมอบและขึ้น Production แล้ว:**
  1. **Collapsible Filter Header ในรายงานการเงิน:** ย่อ/พับแถบเงื่อนไขใน `/report/trialbalance` และ `/report/ledger` ช่วยเพิ่มพื้นที่ตารางรายงานเป็น 15–20 แถวเต็มตา
  2. **แม่แบบรายวันสำเร็จรูปสำหรับ SMEs ไทย 6 แบบ:** ค่าบริการหัก 3%, ค่าเช่าหัก 5%, ค่าน้ำไฟ VAT 7%, เงินเดือน+ประกันสังคม, รับชำระหนี้ถูกหัก 3%, รายได้ค่าบริการ
  3. **Bank Reconciliation Auto-Match:** ตรวจจับคู่ Statement กับสมุดรายวันอัตโนมัติด้วย `amountUnits` (BigInt ไร้ float) พร้อมปุ่มทางลัดจากรายงาน Statement ที่ยังไม่จับคู่
  4. **e-Tax Invoice XML Generator:** สร้างไฟล์ XML ตามมาตรฐาน มธพ. 3-2560 (ETDA)
  5. **Menu Parity Phase 1:** เมนูบาร์บน 9 หัว (230 catalog items) จัดเรียงลำดับและข้อความตรงกับภาพ Champ จริง

---

## 2. ผลการ Audit ความครบถ้วนเทียบ Champ 5.5 (งานที่ต้องทำต่อ)

อ้างอิงพิมพ์เขียวใน `docs/kms/champ/` (โดยเฉพาะ `11-menu-parity-audit-2026-10-09.md` และ `12-phase2-design-research-2026-10-09.md`):

### ก. ระบบจัดซื้อ จัดขาย และสต็อก (Approval & Cancellation Workflow Tier A)
- **ปัญหาปัจจุบัน:** จออนุมัติและยกเลิก 11 จอในเมนูจัดซื้อ จัดขาย สต็อก เปิดแล้วขึ้นสถานะ "รอพัฒนา" (`/pending/*`)
- **สิ่งที่ต้องทำ:**
  1. **Backend Endpoints:**
     - `POST /transaction/:module/:id/approve` (สำหรับ PR, PO, QT, SO)
     - `POST /transaction/:module/:id/unapprove`
     - `POST /transaction/:module/:id/cancel` (รับ `reason` บังคับกรอก, สำหรับ 6 โมดูล: PR, PO, QT, SO, Stock Issue Request, Stock Transfer Request)
     - `POST /transaction/:module/approve-batch` (อนุมัติเป็นชุดแบบ Champ)
     - `GET /transaction/:module/:id/events` (ดึงประวัติ Audit Trail)
  2. **Database Schema:** ใส่ใน `ensureSchema` ของ `backend/internal/transaction/httpapi/schema.go` (ห้าม DDL มือ):
     - ฟิลด์ `approved_by`, `approved_at`, `cancelled_by`, `cancelled_at` ในตาราง `doc`
     - ตาราง append-only `erp_doc_events` (id, businesscode, transflag, docno, action, from_status, to_status, reason, actor, created_at)
     - ดัชนี `idx_doc_pending`
  3. **Frontend Screen Wiring:**
     - เชื่อมโยง `APPROVAL_KINDS` ใน `frontend/src/lib/erp-operations.ts`
     - ปลดล็อกหน้าจอใน `frontend/src/app/operations/operations-workbench.tsx` พร้อม Modal ให้กรอกเหตุผลยกเลิก
  4. **เอกสารขายที่ยังขาด:**
     - ใบสั่งจองสินค้า (Sale Reservation Doc)
     - ใบสั่งจ่ายสินค้า (Delivery Order: DO)

### ข. ระบบเช็คและการเงิน (Cheque & Cash/Bank Parity)
- **สิ่งที่ต้องทำ:**
  1. **แบบฟอร์มพิมพ์เช็คจ่ายธนาคารไทย (Cheque Print):**
     - สร้างหน้าจอและ Template สำหรับพิมพ์เช็คจ่ายตรงตำแหน่งฟอร์มธนาคารไทย: KBANK, SCB, BBL, KTB
     - พิมพ์วันที่, ชื่อผู้รับ (Payee), จำนวนเงินตัวเลขและตัวอักษรภาษาไทย (บาทถ้วน/สตางค์), ตีเส้นคู่ `A/C PAYEE ONLY`, และขีดฆ่าคำว่า `หรือผู้ถือ`
  2. **ข้อมูลการเคลื่อนไหวเช็ค (Cheque Flow Timeline Inquiry):**
     - หน้าจอสอบถามประวัติเช็คตามเลขที่: แสดงไทม์ไลน์สถานะตั้งแต่ รับ/จ่าย -> นำฝาก -> ผ่าน/คืน/ยกเลิก
  3. **บันทึกเช็คเตรียมจ่าย (Batch Cheque Generation):**
     - สร้างชุดเลขที่เช็คสถานะเตรียมจ่าย (Pre-pair Payment)
  4. **บริการธนาคาร (Bank Service Master):**
     - อัตราดอกเบี้ย, ค่าธรรมเนียม, สกุลเงิน ผูกกับสมุดบัญชีธนาคาร

### ค. ระบบลูกหนี้และเจ้าหนี้ (AR & AP Parity)
- **สิ่งที่ต้องทำ:**
  1. **บันทึกเตรียมจ่ายชำระหนี้ (Payment Preparation Voucher):**
     - เอกสารเตรียมจ่ายก่อนออกเช็คหรือตัดเงินจริง
  2. **รายงานอายุหนี้ (Configurable Aging Ranges):**
     - Champ ให้ตั้งช่วงวันของ Aging ได้อิสระ (เช่น 1–7, 8–14, 15–21, 28+ วัน)
     - ปรับให้หน้าจอรายงานอายุหนี้ลูกหนี้และเจ้าหนี้ใน BC สามารถระบุช่วงวันวิเคราะห์ได้เอง (ไม่ใช่ฟิกซ์ 30/60/90 วัน)

### ง. ระบบบัญชีแยกประเภทและภาษี (GL & Tax Parity)
- **สิ่งที่ต้องทำ:**
  1. **Period Lock ข้ามโมดูล:**
     - ปัจจุบัน Period Lock ตรวจเฉพาะสมุดรายวัน GL (`internal/generalledger/postgres_guards.go`)
     - ต้องขยายให้ตรวจบล็อกการบันทึกเอกสารในโมดูลจัดซื้อ จัดขาย และสต็อก (`internal/transaction/`) เมื่อเอกสารตกอยู่ในงวดบัญชีที่ปิดหรือล็อกแล้ว
  2. **มิติวิเคราะห์ 5 ตัวที่เหลือ:**
     - Champ มี 7 มิติ (แผนก, โครงการ, การจัดสรร, งาน, ส่วนงาน, ด้าน, สาขา) ปัจจุบัน BC มี แผนก และ โครงการ

### จ. ชุดรายงาน (Reports Parity)
- Champ มีรายงาน 415 แบบ (BC มี 63 route)
- ยึดหลัก **"1 หน้าจอรายงาน รองรับมิติและตัวกรองยืดหยุ่น"** (ห้ามแตกเป็น 415 เมนูย่อย)
- เพิ่มตัวเลือกพิมพ์, จัดเรียง, สรุปยอด, และ Export Excel/PDF ให้ได้รูปแบบและข้อมูลเทียบเท่า Champ

---

## 3. มาตรฐานการทดสอบ Full-Stack CRUD (บังคับใช้ 100%)

ตามกฎ `AGENTS.md` (หัวข้อ UAT: CRUD+PostgreSQL ทีละ step):

### 1. ทดสอบ CRUD ทีละขั้นตอน (ห้ามรอตรวจรวดเดียวตอนจบ)
สำหรับทุกหน้าจอและ API ที่พัฒนาใหม่:
1. **Create:** บันทึกข้อมูล -> ยิงคำสั่ง `SELECT` ใน PostgreSQL ทันทีเพื่อตรวจว่าแถวใหม่ถูกสร้าง ฟิลด์ครบ และค่าถูกต้อง
2. **Read:** เรียกดูข้อมูล -> ตรวจสอบว่า API คืนค่าครบถ้วน ถูกต้องตาม schema และไม่มี error
3. **Update:** แก้ไขข้อมูล -> ตรวจสอบใน PostgreSQL ว่าแถวเดิมถูกอัปเดต ไม่เกิดแถวซ้ำซ้อนหรือ orphan rows
4. **Delete / Cancel:** ลบหรือยกเลิก -> ตรวจสอบว่าแถวถูกลบจริง หรืออัปเดตสถานะ `isdelete=true` / `iscancel=true` พร้อม Audit Log ครบถ้วน
5. **Isolation Check:** ตรวจสอบว่าข้อมูลของ record อื่นหรือบริษัท/สาขาอื่นไม่ได้รับผลกระทบ

### 2. กฎความถูกต้องของตัวเลขบัญชี (Rule 20 — STRICT)
- **ห้ามใช้ Float ในการคำนวณเงินเด็ดขาด**
- ฝั่ง Frontend: ใช้ integer minor unit (`amountUnits` ทศนิยม 2 ตำแหน่ง หรือ BigInt)
- ฝั่ง Backend Go: ใช้ `shopspring/decimal` หรือ `int64` minor units
- ฝั่ง PostgreSQL: ใช้ `numeric(18,2)` หรือ `numeric(18,4)`
- ตรวจสอบ Debit = Credit สมดุล 100%

### 3. มาตรฐานความปลอดภัย (Rule 10 & 21)
- ห้าม Hardcode Secret, API Key หรือ Credential ในโค้ด
- Query SQL ต้องใช้ Parameterized Query (`$1, $2, ...`) 100% ห้ามต่อ String

### 4. การจัดการภาษา (13 ภาษา)
- ห้าม Hardcode ข้อความในหน้าจอ
- ทุกปุ่ม ป้าย Label และ Error Message ต้องมี key ใน `backend/assets/language/languages.tsv`
- **ต้องกรอกครบทั้ง 13 คอลัมน์** (`key th en cn ja km ko lo my vi ms id fil`) ถ้าไม่ครบ key จะถูกข้ามและแสดงเป็นรหัสดิบ

---

## 4. แหล่งข้อมูลและไฟล์อ้างอิงสำคัญ

1. **ผังและรหัสเมนู Champ:**
   - `docs/kms/champ/11-menu-parity-audit-2026-10-09.md`
   - `docs/kms/champ/12-phase2-design-research-2026-10-09.md`
   - `docs/kms/champ/09-gl.md`, `05-cheque.md`, `01-purchase.md`, `02-sales.md`
   - ภาพหน้าจอและเมนูจริง: `D:\project-champ\bcspec\`
2. **ไฟล์โค้ดเมนูและโมดูล:**
   - รายการเมนู: `frontend/src/lib/menu-data.ts`
   - แถบเมนูบน: `frontend/src/lib/menu-bars.ts`
   - สถานะความพร้อมจอ: `frontend/src/lib/menu-screen-status.ts`
   - ข้อความหลายภาษา: `backend/assets/language/languages.tsv`
3. **ไฟล์ประมวลผลธุรกรรมและ GL:**
   - Transaction API: `backend/internal/transaction/httpapi/`
   - General Ledger API: `backend/internal/generalledger/`
   - Go API Handlers: `backend/internal/goapi/handlers/`

---

## 5. Checklist ก่อนส่งมอบงานและ Deploy (Verification Checklist)

- [ ] **Unit & Component Tests:** Vitest ฝั่ง frontend ผ่าน 100% (`npm test` หรือระบุไฟล์)
- [ ] **Typecheck & Build:** `npx tsc --noEmit` และ `npm run build` ผ่าน 0 error
- [ ] **Backend Tests:** `go test ./...` ฝั่ง Go ผ่าน 100%
- [ ] **Code-Map Sync:** รัน `pwsh -NoProfile -File tools/gen-code-map.ps1` หากแตะไฟล์เกิน 950 บรรทัด
- [ ] **MyDocs Untouched:** ตรวจ `git status` ยืนยันว่าไม่มีการแก้ไขหรือแตะต้องโฟลเดอร์ `mydocs/` (Strict Read-Only)
- [ ] **Clean Code Sweep:** ไม่มี dead code, commented blocks, debug log หลงเหลือ (Rule 16)
- [ ] **Automatic Production Deploy:** รัน `python tools/fast-deploy.py --all --tag <new-tag>`
- [ ] **Live UAT Evidence:** เข้าทดสอบจริงผ่าน Browser/Playwright บน `https://account.bcaicloud.com` ด้วยบัญชี Demo และบันทึก Screenshot เป็นหลักฐาน
- [ ] **การรายงานผลต่อลุงจืด:** รายงานเป็นภาษาไทย สั้น กระชับ เรียก "ลุงจืด" เสมอ พร้อมเสนอทางเลือกถัดไปแบบ `[1]`, `[2]`, `[3]`
