# HANDOFF 2026-10-10 — เมนู Champ parity (Claude → Gemini)

ส่งต่องานทำเมนูหลักให้เหมือน Champ 5.5 (`D:\project-champ`, ภาพจริงใน `D:\project-champ\bcspec` ชนะ `.rc`/`menuconfig.xml`) — อ่าน `AGENTS.md` ของ repo ก่อนเริ่ม (กฎ Champ parity, ภาษา th+en, `mydocs/` read-only, ห้ามชื่อแบรนด์คู่แข่ง)
งาน GL parity คนละเรื่อง → ดู `HANDOFF-2026-10-08-GL-CHAMP-PARITY.md` (ยังใช้ได้ วิธีรัน Go test อยู่ใน §3 ของไฟล์นั้น)

## 1. สถานะตอนส่งต่อ

- branch `dev` HEAD `5e26affe` (+ commit ของไฟล์ handoff นี้) — **นำหน้า `origin/dev` ~9 commit ยังไม่ push** (push เมื่อลุงจืดสั่งเท่านั้น)
- **deploy แล้ว** production `https://account.bcaicloud.com/` = release `r20261009-3` (รวมงานเมนูรอบแรก, container healthy, key ภาษาใหม่ขึ้นจริง) — deploy ครั้งถัดไปใช้ tag `r20261010-1`
- working tree: ไม่ต้อง stage `mydocs/`, `mydocs.zip` (ของลุงจืด), `frontend/next-env.d.ts` (dev server แก้เอง)
- dev server `localhost:3000` อาจยังรันค้างจาก session นี้; local `mainapi` (docker) รันอยู่

## 2. ทำเสร็จแล้ว (commit `f66f059b`, audit `719b2fcc`, สถานะ `5e26affe`)

เมนูบน 9 หัว (ไฟล์, จัดซื้อ, จัดขาย, เจ้าหนี้, เช็ค/บัตร, ธนาคาร/เงินสด, สินค้าคงคลัง, บัญชี, จัดการระบบ) **ตรงภาพ dropdown จริง** ทั้งลำดับ/ตัวคั่น/ถ้อยคำ (opus เทียบภาพทีละบรรทัดแล้ว) ยกเว้นรายการที่ตั้งใจไม่ทำ (§4-A)
- catalog 204→230 id; รายการที่ยังไม่มีจอ = route `/pending/<id>` ขึ้นการ์ด "ยังไม่พร้อม" (20 รายการ + 10 `system-*`; `audit-pending` นับ 49)
- จอ ยกเลิก/อนุมัติ 5 จอใหม่อยู่ใน `OPERATIONS_CONFIGS` (17→22) — **ทั้ง 11 จอเป็น pending** เพราะ `APPROVAL_KINDS = {}` (`frontend/src/lib/erp-operations.ts:253`) และไม่มี API
- `gl-unpost` → `/gl/unposting` (จอกลับรายการ) เปิดได้จริง
- node ชนิด `route` ใหม่ใน `frontend/src/lib/menu-bars.ts` (`routeNode()` บรรทัด 46, `menuBarEntries()` 344, `isMenuBarRouteItem`) เปิดจอตั้งค่าเดิมจากแถบเมนู (`/user`, `/permissiongroup`, `/branch`, `/employee`, `/productwarehousescreen`, `/debtorgroup`) — ไม่อยู่ใน permission catalog เมนูไม่ล็อก (`main-menu-screen.tsx` `canAccessMenuItem`), backend คุมสิทธิ์ตอนบันทึก
- `/productwarehousescreen`, `/debtorgroup` อยู่ใน `RETIRED_BACKEND_ROUTES` (`frontend/src/lib/menu-screen-status.ts:35`) → การ์ด "ยังไม่พร้อม" (API เดิมถูกถอด)
- "กำหนดรหัสต่างๆ" เป็น submenu รวม gl-dimensions/product-group/brand/warehouse/product-unit/barcode — **Champ เป็นจอเดียวมี combo เลือกประเภทรหัส** (เบี่ยงตั้งใจ รายงานลุงจืดแล้ว)
- ทดสอบแล้ว: `npm run verify:all` ผ่านทั้งหมด, vitest frontend 1288 ผ่าน, ลองกดผ่านปุ่ม Demo (ไฟล์ › กำหนดรหัสผู้ใช้งาน เปิดหน้าผู้ใช้, บัญชี › ยกเลิกเอกสารที่ผ่านรายการแล้ว เปิดจอกลับรายการ, จัดซื้อ › ยกเลิกใบเสนอซื้อสินค้า ขึ้น "ยังไม่พร้อม", กลุ่มลูกหนี้ ขึ้น "ยังไม่พร้อม")

## 3. เอกสารที่ต้องอ่าน (ในลำดับนี้)

1. `docs/kms/champ/11-menu-parity-audit-2026-10-09.md` — ตารางเมนูขาด/เกินต่อหัว + **หัวข้อ "สถานะหลังแก้รอบแรก"** ท้ายไฟล์
2. `docs/kms/champ/12-phase2-design-research-2026-10-09.md` — ผลสำรวจก่อนเฟส 2/3: ออกแบบยกเลิก/อนุมัติ (Tier A/B, endpoint, SQL, ความเสี่ยง, test), ขนาดจอใหม่ 11 จอ, submenu รายงานที่ขาด
3. `docs/kms/decisions/2026-10-09-champ-parity-purpose-additive-extras.md` — เส้นแบ่ง "เสริมเล็กทำได้ / ใหญ่ต้องอนุมัติ"
4. ภาพ Champ: `D:\project-champ\bcspec\หน้าจอเมนู<หัว>\เมนู-ประกอบ*.jpg` (จัดซื้อสะกด "จัดชื้อ"; `หน้าเมนูหน้าต่าง`, `หน้าจอเมนูไฟล์`), PDF รายงาน `Report_ระบบ*`

## 4. งานที่เหลือ

### A. ต้องถามลุงจืดก่อน — ห้ามตัดสินเอง (ลุงจืดยังไม่ได้ตอบ 5 ข้อนี้)
1. **ยกเลิก/อนุมัติ 11 จอ:** Tier A (ระดับหัวเอกสาร ใช้คอลัมน์ `iscancel/approval_status` เดิม ~2 วัน — ผมแนะนำ) vs Tier B (ใบยกเลิกระดับบรรทัดแบบ Champ เป๊ะ ~5–8 วัน) — Tier A เปลี่ยน workflow ของ Champ จึงต้องอนุมัติ; รายละเอียด `12-…md` §ก
2. **ของ Champ ที่เว็บทำไม่ได้/เสี่ยง:** สร้างการเชื่อมโยงข้อมูล, เลือกใช้ข้อมูลตัวอย่างต่างๆ (หัวไฟล์), ขอดูข้อมูล (ยิง SQL ดิบ เสี่ยงความปลอดภัย), แสดง ToolTip (จัดการระบบ) — ตอนนี้ **ไม่ใส่** ในเมนู; ห้ามตัดของ Champ ถ้าไม่มีคำสั่งเป็นลายลักษณ์อักษร → ถามว่าจะตัด/ทำแบบอื่น
3. **รายงาน (415 ใน Champ vs 63 ใน BC):** รูปแบบ "หนึ่งเมนู = หนึ่งรายงานตามชื่อ Champ" หรือ "ตัวเลือกในจอรายงานเดียว" — แล้วเริ่มที่จัดขายกับเช็ค/บัตร (backend-first); submenu "รายงานการขายสินค้าชุด(สินค้าชุดแบบที่ 2)" ไม่รู้สมาชิก (ไม่มีภาพ cascade) → ขอภาพจากลุงจืด
4. **ส่วนเสริมที่ BC มีแต่ Champ ไม่มี:** Flow ใบสั่งจองสินค้า (`sale-reservation-flow`), ตรวจสอบวันที่ใบสั่งขาย (`sale-order-date-check`), `etax-invoice`, GL extras (gl-dimensions ย้ายไป จัดการระบบ แล้ว; gl-fiscal-years, gl-year-end, tax-forms ยังอยู่หัวบัญชี) — คง/เอาออก; และแก้ข้อความใน `AGENTS.md` ที่ว่า `sale-reservation-flow` "มีใน Champ ห้ามตัด" (ภาพ dropdown จริงไม่มี) หลังลุงจืดตัดสิน
5. **เสนอใหญ่อื่นๆ** (เมนู/จอ/โมดูลใหม่ที่ Champ ไม่มี) — เสนอข้อดี-ข้อเสียก่อนเสมอ

### B. ทำต่อได้เลย ไม่ต้องรออนุมัติ (เป็นของ Champ เอง — รายงานท้ายงานว่าทำอะไร)
1. **หัวหน้าต่าง:** เพิ่ม submenu "เลือกตัวอักษรในการแสดงผลต่าง" 6 รายการ (หน้าจอทั่วไป / รายงานทั่วไป / ชื่อของรายงาน / คอลัมน์ของรายงาน / รายการย่อยของรายงาน / ชื่อบริษัทในรายงาน — ตรวจสะกดจากภาพ `หน้าเมนูหน้าต่าง/`) ปัจจุบันหัวนี้มีแค่ ปิดหน้าจอทั้งหมด (`menu-bars.ts` ~บรรทัด 322); เริ่มด้วย `/pending/window-font-*` แล้วต่อ `หน้าจอทั่วไป` เข้า FontPicker เดิม (`frontend/src/app/font-picker.tsx`, `lib/font-data.ts`)
2. **หัวช่วยเหลือ:** เพิ่ม About และ "มาตรฐานซอฟท์แวร์ของกรมสรรพากร (Standard software)" (ตอนนี้มีแค่ คู่มือการใช้งาน ซึ่งเป็นส่วนเสริมของ BC) — เนื้อหา Standard software ต้องอิงแหล่งทางการของกรมสรรพากร (Rule 17) ถ้ายังไม่ชัดให้ใส่เป็น `/pending/` + ถามลุงจืด
3. **ตรวจสิทธิ์เมนูใหม่:** 26 id ใหม่ไม่ได้ถูกให้อัตโนมัติกับชุดสิทธิ์ที่ไม่ใช่ `*` (Demo/owner ได้ `*` จึงไม่เห็นปัญหา) — ตรวจว่า `permissiondefinition`/catalog ฝั่ง backend ต้องลงทะเบียน id ใหม่ไหม และผู้ใช้ชุดสิทธิ์จำกัดเห็นเมนูถูกล็อกอย่างที่ตั้งใจหรือไม่ (ยังไม่ได้ตรวจ) ; route node 6 ตัวไม่ผ่านระบบสิทธิ์เมนูเลย — ถ้าลุงจืดต้องการให้ล็อกตามสิทธิ์ ให้ map route → id สิทธิ์เดิม (เช่น `/productwarehousescreen` → `warehouse`)
4. เทียบ **หัวลูกหนี้, สินทรัพย์** กับภาพอีกครั้งหลังแก้อะไรที่กระทบ `menu-bars.ts` (รอบนี้ตรงอยู่แล้ว)

### C. หลังลุงจืดอนุมัติ
- **เฟส 2** ยกเลิก/อนุมัติ (ถ้าเลือก Tier A): ไฟล์ที่เกี่ยว `backend/internal/transaction/httpapi/{http.go,service.go,schema.go,save.go}` (DDL ใส่ใน `ensureSchema` เท่านั้น ห้ามมือ), `frontend/src/lib/erp-operations.ts` (`APPROVAL_KINDS` + fetcher ที่ยิง `/api/goapi/api/approval/...` ซึ่งไม่มีจริงต้องลบ), `frontend/src/app/operations/operations-workbench.tsx`, BFF `app/api/erp-transaction/[...erpPath]/route.ts` (POST ต้องส่ง JSON object `{}`), เทสต์ที่ล็อกสถานะ pending ต้องพลิกใน change เดียวกัน (`erp-operations.test.ts`, `menu-screen-status(.test).ts`)
- **เฟส 3** จอใหม่ 11 จอ (ใบสั่งจอง, ใบสั่งจ่าย, เตรียมจ่ายชำระหนี้, เช็คเตรียมจ่าย, ข้อมูลการเคลื่อนไหวเช็ค, บริการธนาคาร, Supplier ตามสินค้า, Hand Held, Transaction Lock, ปรับปรุงวันที่รับสินค้า, system-head masters) — ขนาด/คำถามเปิดต่อจอใน `12-…md` §ข; **ทุกจอเป็น "จอใหม่" ต้องเทียบภาพ Champ ก่อนและเสนอลุงจืดถ้าเปลี่ยน workflow**
- **เฟส 4** รายงาน (หลังตกลงรูปแบบ)

### D. งานค้างอื่นที่ไม่ใช่เมนู (ลุงจืดยังไม่ตัดสิน ห้ามแตะเอง)
RUNGRUENG prod ยังไม่มี posting mappings ของ doctype ที่ลง GL; Demo ของ prod ผูก identity somkid@rungrueng.co.th (มี TOFFT1 owner ห้ามแตะ); ความซ้ำซ้อนใน `D:\project-champ` (ห้ามลบอะไรโดยไม่สั่ง); รายการ "BC มากกว่า Champ" ใน `docs/kms/champ/*.md`; ช่องว่าง ERP (Details subledger, COGS lines, negative-stock policy, WH01 hard-code, reopen source docs on delete) ดู `docs/kms/decisions/2026-10-09-erp-posting-reverses-on-edit-delete.md`

## 5. สูตรเพิ่ม 1 เมนู (ต้องแก้ครบพร้อมกัน — `docs/README.md` แถวงานเมนูหลัก)

1. `frontend/src/lib/menu-data.ts` — `tx(id, th, en, route, category?, labelKey?)` ในกลุ่มที่ถูกต้อง (ชื่อ th = ถ้อยคำ Champ; **ห้ามเปลี่ยน id/route/language key ของของเดิม**; ชื่อเก่า → `CHAMP_SEARCH_ALIASES`)
2. `frontend/src/lib/menu-bars.ts` — วางตามลำดับ/ตัวคั่นของภาพ (ทุก catalog item ต้องอยู่ใน bar **ครั้งเดียว** — test บังคับ)
3. `backend/assets/language/languages.tsv` — 1 แถว **13 คอลัมน์** (`key th en cn ja km ko lo my vi ms id fil`; th ตรง label, ภาษาอื่น = อังกฤษ placeholder ห้ามว่าง; ไม่ครบ = key ถูกข้าม คืน key ดิบ)
4. `frontend/src/lib/menu-icons.ts` — icon ของ route (`/pending/<id>` ก็ต้องมี)
5. เทสต์: `menu-icons.test.ts` (จำนวน), `menu-data.test.ts` (catalog **230**), `audit-pending.test.ts` (pending **49**), `menu-bars.test.ts` (`CHAMP_ONLY_PENDING_IDS` ถ้าเป็น `/pending/`), `menu-screen-status.test.ts`
- ของที่ยังไม่มีจอ → `route: "/pending/<id>"` (ตัวอย่าง `billing-note-auto`, `system-*`) ห้ามทำจอเปล่า/จอปลอม

## 6. ข้อควรระวัง (เจอจริงรอบนี้)

- **EOL ผสม:** `languages.tsv` (CRLF เป็นหลัก), `menu-bars.ts` (CRLF), `menu-screen-status.ts` (CRLF+LF ปนกัน), `docs/kms/README.md` (CRLF+LF), `globals.css` (CRLF/CR/LF ปน) — **ห้ามใช้ Edit/Write แทนทั้งไฟล์**; แก้ด้วย Node patch ที่ byte-preserving (อ่าน `latin1`, replace string ที่ unique, เขียน `latin1`) แล้ว `git diff --stat` ต้องเห็นเฉพาะบรรทัดที่ตั้งใจ
- ใน bash/Node: backslash ซ้อนยุบ → regex `\s`/`\d` เพี้ยน ใช้ `[ \t]`/`[0-9]` หรือเขียนสคริปต์เป็นไฟล์ `.cjs`
- pre-commit `.githooks/pre-commit` เช็ค `docs/reference/CODE-MAP.md` เมื่อแตะไฟล์ ≥950 บรรทัด (เช่น `main-menu-screen.tsx`): ถ้าไม่ผ่านรัน `pwsh -NoProfile -File tools/gen-code-map.ps1` แล้ว `git add docs/reference/CODE-MAP.md`
- สคริปต์ตรวจเมนู `C:/tmp/menu-audit/extract-bc-v2.cjs` ถอดโครงเมนู BC เป็น JSON (ตัว `extract-bc.cjs` เก่าไม่รู้จัก node `route`) — `C:/tmp` เป็นของชั่วคราว อาจหาย; ใช้ `menu-bars.test.ts` เป็นตัวตรวจหลัก
- BFF: ข้อผิดพลาดฝั่งผู้ใช้ถูกส่งเป็น HTTP 200 + `success:false` (ดู `docs/kms/17-dev-gotchas.md`) — ตรวจ payload จริง ไม่ใช่ดูแค่ 200
- `401` ที่ `/api/auth/refresh` ก่อนล็อกอินเป็นเรื่องปกติ ไม่ใช่บั๊ก

## 7. วิธีตรวจ/ส่งมอบ (ตามกฎ repo)

- `npm run verify` (codemap + frontend) ก่อน commit ทุกครั้ง; แคบ: `cd frontend && npx vitest run src/lib/menu-bars.test.ts src/lib/menu-data.test.ts src/lib/menu-screen-status.test.ts src/lib/erp-operations.test.ts src/lib/menu-icons.test.ts src/lib/audit-pending.test.ts` + `npx tsc --noEmit -p .`; แตะ backend/languages.tsv → `npm run verify:all` ก่อน deploy
- ทดสอบจอเอง: dev server `:3000` → หน้า login กด "ทดลองใช้ระบบ (Demo)" → กลุ่ม `rungrueng` → บริษัท `01` → สาขา `00000` → เมนูหลัก (สลับ "เมนูบน"); ตรวจ payload จริง, console, ตรวจ PostgreSQL หลังแต่ละ operation ที่เขียนข้อมูล; ห้ามขอลุงจืดทดสอบแทน
- commit: `git add <path>` ระบุไฟล์ (ห้าม `-A`/`.`), บรรทัดแรกอังกฤษ conventional + body ไทย (แก้อะไร/ผลทดสอบ), ห้าม `--no-verify`, ห้าม stage `mydocs/`; จบงาน+verify ผ่าน → deploy auto: `py tools/fast-deploy.py --all --tag r20261010-1` (frontend อย่างเดียวไม่ต้อง rebuild mainapi; แต่ `languages.tsv` อยู่ฝั่ง backend ต้อง `--all`) แล้วตรวจ `docker ps` บน prod + payload จริง
- ท้ายงานทุกครั้ง: ✅ ข้อดี / ⚠️ ข้อเสีย / 💡 คำแนะนำ (งานเล็ก 1 บรรทัดต่อหมวด) + บรรทัด **"เสริมเกิน Champ: …"** (หรือ "ไม่มี") + ขั้นต่อไปเป็น `[1] [2] [3]` (ตัวแนะนำก่อน) — เรียกผู้ใช้ว่า "ลุงจืด", ตอบไทยสั้นๆ
