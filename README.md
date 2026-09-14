# BC Ai Account — ระบบบัญชีและการเงินอัจฉริยะ

ระบบ ERP และบัญชีสำหรับธุรกิจไทย รองรับการทำงานแบบ Multi-Tenant (Holding / Company / Branch) ประมวลผลรวดเร็วด้วยสถาปัตยกรรมแบบ 2-Tier (MongoDB Storage + PostgreSQL Processing Engine) พร้อมการออกแบบ UX/UI ที่เป็นมิตรกับคนไทยอายุ 40+ ใช้งานง่าย ชัดเจน และปลอดภัย

---

## 📋 บันทึกประวัติการพัฒนาและแก้ไขระบบ (Project Activity Log)

> **กฎเหล็กของระบบ**: ทุกครั้งที่มีการแก้ไขโค้ด, เพิ่มฟีเจอร์, แก้บั๊ก, ปรับ UI หรือคอนฟิก **ต้องเพิ่มบันทึกรายการในส่วนนี้เสมอ** (เรียงลำดับจากล่าสุดอยู่บนสุด) และ commit ไปพร้อมกับโค้ดใน commit เดียวกันเสมอ

### 2026-09-14 — ลบ Dead Code แม่แบบธนาคารไทยและ alias เก่าในหน้าตั้งค่าระบบ

- [Refactor] ลบ `ThaiBankTemplateDialog`, `saveThaiBanks`, `thaiBankDialogOpen`, และการตรวจ `slug === "bank"` ออกจาก `frontend/src/app/system-settings/system-settings-screen.tsx` เนื่องจากระบบรวมการจัดการธนาคารเข้าสู่หน้าสมุดบัญชีเงินฝาก (`bookbankscreen`) ซึ่งมีระบบค้นหาและเติมข้อมูลธนาคารไทยอัตโนมัติ (`BookBankFieldEditor`) อยู่แล้ว
- [Refactor] ลบ alias คอนฟิกธนาคารเก่าที่ไม่ได้ใช้งานใน `frontend/src/lib/system-setting-screens.ts`
- [Docs] อัปเดต `docs/reference/CODE-MAP.md` สำหรับไฟล์ขนาดใหญ่ `>= 950` บรรทัดด้วย `tools/gen-code-map.ps1`
- ไฟล์: `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/lib/system-setting-screens.ts`, `docs/reference/CODE-MAP.md`
- หลักฐาน: `tsc --noEmit` ผ่าน, vitest 472/472 ผ่าน

### 2026-09-14 — ผังบัญชี (Chart of Accounts): ปรับปรุงการแจ้งเตือนข้อผิดพลาดและย้ายโฟกัสไปยังช่องที่ผิด

- [Fix] ปรับ `errorStatePatch` ให้รับ `fallbackField` และเพิ่ม `saveFailureTarget` เพื่อชี้เป้าหมายช่องที่ผิดพลาด (เช่น `accountcode`) เสมอ แม้ API จะไม่ได้ระบุฟิลด์
- [Fix] ปรับปรุง `useEffect` ย้ายโฟกัสใน `frontend/src/app/gl/gl-masters.tsx` ให้รอจนกว่าสถานะ `busy` จะเสร็จสิ้น เพื่อไม่ให้ติดสถานะ `disabled` ของฟิลด์เซ็ต
- [Test] เพิ่ม Unit Test ใน `gl-masters.test.ts` และอัปเดต Playwright E2E assertion ใน `gl-chart-of-accounts-error.spec.ts`
- ไฟล์: `frontend/src/app/gl/gl-masters.tsx`, `frontend/src/app/gl/gl-masters.test.ts`, `frontend/e2e/gl-chart-of-accounts-error.spec.ts`
- หลักฐาน: `tsc --noEmit` ผ่าน, vitest 472/472 ผ่าน

### 2026-09-14 — ยกเลิกการแต่งชื่อธนาคารจากรหัสใน API Proxy ของสมุดบัญชีธนาคาร

- [Fix] แก้ไข `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`: ยกเลิกการแต่งค่า `banknames` จาก `bankcode` เมื่อผู้ใช้ไม่กรอกชื่อธนาคาร ซึ่งเดิมทำให้ผ่าน validation ไปบันทึกชื่อธนาคารเป็นรหัส (เช่น KBANK/BBL) ในฐานข้อมูล
- [Fix] เพิ่มการตรวจสอบ `validateSystemSettingWrite`: หากไม่มีการระบุชื่อธนาคาร (`banknames` หรือ `names`) ให้ส่ง HTTP 400 ภาษาไทย "กรุณาระบุชื่อธนาคาร" กลับทันที
- ไฟล์: `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`, `route.test.ts`
- หลักฐาน: `tsc --noEmit` ผ่าน, vitest 472/472 ผ่าน (รวมเคสทดสอบปฏิเสธบันทึกสมุดบัญชีที่ไม่มีชื่อธนาคารด้วย HTTP 400)

### 2026-09-14 — เพิ่มการตรวจอ้างอิงแม่แบบงบการเงินก่อนลบผังบัญชี (ป้องกันสูตรคำนวณงบได้ 0)

- [Fix] แก้ไข `backend/internal/generalledger/references.go` ฟังก์ชัน `accountMasterReferences`: เพิ่มฟิลด์ `rows.accountcodes` ในการตรวจสอบเอกสารอ้างอิงข้ามคอลเลกชัน (`gl_statement_templates`)
- [Fix] ป้องกันไม่ให้ลบผังบัญชีที่ถูกนำไปผูกไว้ในแถวของแม่แบบงบการเงิน (Statement Templates) ซึ่งเดิมไม่ได้ถูกตรวจ ทำให้ลบบัญชีสำเร็จแล้วหน้างบการเงินคำนวณยอดเงินได้ 0 เงียบ ๆ
- ไฟล์: `backend/internal/generalledger/references.go`
- หลักฐาน: Go build `./...`, `go vet`, และ `go test` ใน Docker `golang:1.26` ผ่าน 100%

### 2026-09-14 — แก้บั๊กค้นหารายการ GL ใน PostgreSQL ไม่ตรง (False Positive จากคีย์ JSON)

- [Fix] แก้ไขการค้นหาใน `backend/internal/generalledger/postgres.go` ฟังก์ชัน `List`: เดิมแปลง JSON ทั้งก้อนเป็นข้อความ (`payload::text`) ทำให้การค้นหาคำทั่วไป เช่น `th`, `true`, `code`, `isactive` คืนค่าทุกแถวเนื่องจากไปตรงกับชื่อคีย์หรือค่าแฟล็กใน JSON
- [Fix] เปลี่ยนมาค้นหาเจาะจงเฉพาะฟิลด์เนื้อหาจริง: รหัส (`code`), ชื่อบัญชีทุกภาษา (`jsonb_path_query_array(payload, '$.names[*].name')`), ชื่อแม่แบบ/กลุ่ม (`payload->>'name'`), คำอธิบายและเลขอ้างอิงสมุดรายวัน (`description`, `reference`)
- ไฟล์: `backend/internal/generalledger/postgres.go`
- หลักฐาน: Go build `./...`, `go vet`, และ `go test` ใน Docker `golang:1.26` ผ่าน 100%, ทดสอบกับ PostgreSQL 18 ตรงตามสเปก ค้นหา `th`/`true` ได้ผลลัพธ์เป็น false และค้นหาชื่อ/รหัสได้ผลลัพธ์เป็น true ถูกต้อง

### 2026-09-14 — แก้บั๊ก reportCsv ส่งออกตัวเลขไม่ติดเครื่องหมายคำพูดเดี่ยว (Apostrophe)

- [Fix] แก้ไขฟังก์ชัน `reportCsv` ใน `frontend/src/lib/general-ledger.ts`: เดิมใส่เครื่องหมาย `'` นำหน้ายอดเงินทุกช่อง ทำให้เปิดในโปรแกรมสเปรดชีต (เช่น Excel) แล้วกลายเป็นข้อความและไม่สามารถคำนวณผลรวม (SUM) ได้
- [Fix] ปรับ `csvCell` ให้รับพารามิเตอร์ `isAmount`: ป้องกัน Formula Injection สำหรับคอลัมน์ข้อความ แต่เว้นคอลัมน์ตัวเลขให้ส่งออกเป็นค่าตัวเลขบริสุทธิ์ (ทั้งค่าบวก ค่าลบ และศูนย์)
- ไฟล์: `frontend/src/lib/general-ledger.ts`, `frontend/src/lib/general-ledger.test.ts`
- หลักฐาน: `tsc --noEmit` ผ่าน, vitest 470/470 ผ่าน (รวมเคสส่งออกตัวเลขบวกและลบใน `general-ledger.test.ts`)

### 2026-09-14 — ตรวจและแก้เมนูผังบัญชี (Chart of Accounts) — แจ้งข้อผิดพลาดเป็นภาษาไทย

- [Fix][UI/UX] เดิมกดปุ่ม “บันทึกข้อมูล” แล้วบันทึกไม่ผ่าน (เช่นใส่รหัสบัญชีซ้ำกับที่มีอยู่) หน้าจอไม่บอกอะไรเลย และเบราว์เซอร์ขึ้น error เพิ่มอีก 1 รายการ — ตอนนี้ขึ้นข้อความไทยชัดเจนในหน้าต่างแก้ไข เช่น “รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น” ตัวอักษร ≥0.9rem และไม่โชว์ข้อความอังกฤษ/รหัสเทคนิคให้ผู้ใช้เห็น
- [Fix] บันทึกไม่ผ่านแล้วระบบไม่ล้างค่าที่พิมพ์ไว้ และเลื่อนตำแหน่งเคอร์เซอร์ไปที่ช่อง “รหัสบัญชี” ทันที โดยไม่เลื่อนหน้าจอ (กันผู้ใช้ว่าพิมพ์ผิดช่องไหน)
- [UI/UX] เลิกใช้สีแดงแบบเขียนตายตัวในปุ่มลบ 3 จุด เปลี่ยนมาใช้สีจากธีม (text-destructive / ขอบและพื้น hover จากโทเคนธีม) จึงถูกต้องครบทั้ง 10 พาเลตและโหมดมืด
- [Feature] เพิ่มช่อง “ชื่อบัญชีภาษาอังกฤษ (ไม่บังคับ)” ให้เทียบเท่า Name2 ของระบบเดิม ใช้โครงสร้าง names[] ที่มีอยู่แล้ว ไม่ต้องแก้ฐานข้อมูล
- [Fix] ชั้น BFF: ข้อผิดพลาดที่ผู้ใช้แก้เองได้ (4xx ที่มีรหัส code เช่น duplicate_code) ส่งกลับเป็น HTTP 200 + success:false เพื่อไม่ให้ console ของเบราว์เซอร์ขึ้น error ซ้ำซ้อน ส่วน 401/403/5xx ยังคงสถานะเดิมไว้ตามจริง
- ไฟล์: `frontend/src/app/gl/gl-masters.tsx`, `frontend/src/app/gl/gl-common.tsx`, `frontend/src/lib/general-ledger-api.ts`, `frontend/src/lib/workspace-api.ts`, `frontend/src/app/api/gl/[...glPath]/route.ts`, `backend/internal/generalledger/httpapi/http.go`, `backend/internal/generalledger/errors.go`
- หลักฐาน: `npx tsc --noEmit` ผ่าน, eslint 0 error (226 warning ที่มีอยู่เดิม), vitest โฟลเดอร์ `src/app/gl` + `src/lib` ผ่าน 273/273; Playwright `frontend/e2e/gl-chart-of-accounts-error.spec.ts` ผ่าน 1/1 — พบ role="alert" ภาษาไทย 1 อัน ข้อความ “รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น”, console error ของการบันทึกที่ล้มเหลว 0 รายการ, จำนวนบัญชีในลิสต์ 38 → 38 รายการ (ไม่มีข้อมูลทดสอบตกค้าง) ภาพหน้าจอ 11 ภาพอยู่ที่ `frontend/test-results/gl-chart-of-accounts-error-71f9b-ept-zero-new-console-errors/` (light/dark × 1600/1280/1024/768 + hover/focus/disabled)
### 2026-09-14 — แก้บั๊กจาก code review ก่อน commit (GL v2, คลังสินค้า, ช่องตัวเลข)

- [Fix] ตรวจโค้ดค้างใน working tree 8 มุม แล้วแก้ 9 จุดที่ยืนยันแล้ว: mainapi ไม่ล่มทั้งตัวเมื่อ Kafka ของ GL ตั้งค่าผิด, GL worker ใช้ consumer group เดียว (ไม่ apply ซ้ำ 2 รอบ), ปิดงบ/ยอดยกมารองรับเกิน 500 บรรทัด, งบการเงินในตัวออกแบบอ่านครบทุกหน้า (ไม่ตัดที่ 1000 บัญชี)
- [Fix] ตารางที่เก็บสินค้า: เตือนก่อนเปลี่ยนคลังทำงานจริง (await confirm), แก้ชื่อที่เก็บไม่ทำให้สิทธิ์บริษัท/สถานะเดิมหาย, บันทึกล้มเหลวกลางทางแล้วกดซ้ำไม่ยิงรายการที่สำเร็จแล้วซ้ำ
- [Fix] ช่องตัวเลข (NumericInput) ไม่ปัดค่าที่เก็บ (0.125 โชว์ 0.125); DevDomInspector คงไว้บน production ตาม ADR 2026-09-12 (ลุงจืดสั่ง)
- ไฟล์: `backend/main.go`, `backend/internal/generalledger/httpapi/http.go`, `backend/internal/generalledger/models.go`, `frontend/src/app/system-settings/warehouse-tree-view.tsx`, `frontend/src/components/ui/numeric-input.tsx(+test)`, `frontend/src/app/gl/gl-statement-designer.tsx`, `docs/kms/bugs/2026-09-14-code-review-gl-warehouse-fixes.md`
- ยังไม่แก้ (รอตัดสินใจ): เมนู `/line-oa` หายจากผังเมนูใหม่ และสิทธิ์ปุ่มของหน้าจอระดับ Holding ที่ไม่อยู่ในเมนู
- หลักฐาน: `npx tsc --noEmit` ผ่าน, vitest numeric-input 7/7 + menu 38/38 ผ่าน, Go build/vet/test แพ็กเกจ generalledger ใน Docker (ดูผลด้านล่างใน log งาน)

### 2026-09-14 — เปลี่ยนผู้ช่วย AI เหลือ DeepSeek ตัวเดียว ("Fable คิด, DeepSeek ทำ")

- [Docs] ลุงจืดสั่งถอด Kimi K3 + GLM ออกจากกฎผู้ช่วย ให้ Claude (Fable) เป็นคนคิด/แบ่งงาน/ตรวจ และ DeepSeek เป็นคนร่างโค้ด/เอกสาร เพื่อประหยัด token Claude
- ไฟล์: `AGENTS.md` (section ผู้ช่วย), `docs/kms/17-dev-gotchas.md`, global `~/.claude/CLAUDE.md` (Orchestration Rule)
- หลักฐาน: ยิง `py ~/.claude/tools/deepseek-ask.py` ทดสอบจริง ตอบกลับ 180 token สำเร็จ; ไม่แตะโค้ด frontend/backend

### 2026-09-11 — Deploy บัญชีแยกประเภทผ่าน Kafka

- ปล่อย mainapi/worker/frontend `r20260911-gl-kafka-1` เวลา 18:11:35 น. ไทย ทุกบริการ healthy; MongoDB เก็บต้นฉบับ → Kafka reference → PostgreSQL ประมวลผล → Mongo delivered → commit offset
- GL Linux 24 tests + 63 subtests, frontend 393 tests และ release backend/outbox/projection ผ่าน; คงข้อจำกัด 15 packages ใน quarantine และ 2 เมนูรอข้อมูลต้นทาง/แม่แบบ DBD
- Topic 6 partitions / RF1; backup MongoDB/PG/config รอบใหม่ก่อน deploy เป็น same-server safety copy ยังไม่ใช่ restore drill; ข้อมูลตัวอย่างวัสดุก่อสร้าง 93 records / 110 events ผ่าน seed และกระทบยอด Mongo/Kafka/PG พร้อม 5 reports แล้ว; frontend patch `r20260911-gl-demo-thai-1` deploy 18:29:44 น. ไทย ผ่าน 401 tests และ final UI ผ่าน 5 reports / 35 routes (33 ใช้งาน / 2 รอข้อมูล), console errors 0
- [Release, หลักฐาน และ rollback](docs/kms/decisions/2026-09-11-deploy-gl-kafka-demo.md)

### 2026-09-11 — Deploy ระบบบัญชีแยกประเภทขึ้น account.bcaicloud.com

- ปล่อย mainapi/worker/frontend รุ่น `r20260911-gl-v2-1` ทุกบริการ healthy และ HTTPS 200; เชื่อม 33 เมนู ส่วนประมวลผลเอกสารเดิม/XBRL ยังรอข้อมูล
- ผ่าน frontend 384 tests, backend/outbox/projection และ GL Linux 17 tests + 38 subtests; production smoke เปิดครบ 35 เมนู, GL GET 82 ครั้งตอบ 200 และ console errors 0
- สำรอง MongoDB/PG/config ก่อนปล่อยและทดสอบ restore Mongo ในฐานแยก; สร้างเฉพาะฐาน PG `test` ของ holding ที่ขาด โดยไม่สร้างรายการบัญชีจริง เก็บ image เดิมสำหรับ rollback
- แก้เฉพาะ Kafka test setup ให้รอ leader ภายในเวลาจำกัด และปรับ E2E Account Mapping เป็นพร้อมใช้ตามระบบใหม่
- [หลักฐาน ข้อจำกัด และ rollback](docs/kms/decisions/2026-09-11-deploy-general-ledger-v2.md)
### 2026-09-11 — ระบบบัญชีแยกประเภทใหม่ตาม Champ

- เชื่อม 35 เส้นทาง: 33 เมนูมีหน้าจอและ API; ประมวลผลเอกสารซื้อขายเดิมรอต้นทางที่ตรวจสอบได้ และ XBRL เป็นหน้าเตรียมข้อมูลรอแม่แบบบริษัท
- เพิ่ม ledger เงินแม่นยำ MongoDB Decimal128 → PostgreSQL numeric, CRUD/ผ่าน/กลับรายการ, งวด, รายงาน, ปิดงบและยกยอดแบบคงสาขา/แผนก/โครงการ พร้อมกันคำขอซ้ำและประวัติแก้ไม่ได้
- ตรวจ Go GL 16 tests + 38 subtests, frontend 29 tests/TypeScript/lint และ browser UAT บน production build ผ่าน 29.6 วินาที; ภาพ light/dark ครบ 8 แบบ, Mongo 19 ขั้นตอนและยอด PG ตรงกัน ล้างข้อมูลทดสอบแล้ว ไม่ย้ายข้อมูลจริง; deploy ภายหลังตามบันทึก release ด้านบน
- MongoModel projectRev 1423 → 1480; diagram/relations/workflow และ lint ผ่าน อัปเดต UI skill §8.12
- รายละเอียดและ Markdown ครบ 35 เมนู: [คู่มือและผลตรวจ](docs/kms/architecture/2026-09-11-general-ledger-v2.md)
### 2026-09-11 — Deploy เมนูล่าสุดขึ้น account.bcaicloud.com

- **[Deploy]** อัปเดต frontend เป็น `bcai-account-frontend:r20260911-menu-top-1`: 223 เมนูตาม Champ, ชื่อหมวดไทยล้วน และเมนูบนเป็นค่าเริ่มต้น
- **ตรวจ:** image healthy / HTTPS 200; frontend 357 tests + TypeScript + Docker build ผ่าน; backend/outbox/projection ผ่าน; Playwright บน production ผ่านครบ 3 tests
- **ข้อจำกัด:** wrapper `verify:all` เรียก npm ผ่าน Git Bash ไม่ได้ จึงรันชุด frontend เทียบเท่าผ่าน PowerShell; lint ยังมี 219 warnings / 0 errors ไม่ได้แก้ backend หรือข้อมูลบัญชี
- **Rollback/หลักฐาน:** [บันทึก release](docs/kms/decisions/2026-09-11-deploy-champ-menu.md) — เก็บ image และ release.env เดิมไว้บนเซิร์ฟเวอร์

### 2026-09-11 — ตั้งเมนูบนเป็นค่าเริ่มต้น

- **[UI/UX]** ผู้ที่ยังไม่ได้เลือกรูปแบบเมนูจะเริ่มที่เมนูบน; จำรูปแบบที่ผู้ใช้เลือกไว้และคืนค่าเมนูซ้ายได้ถูกหลังรีโหลด
- **ไฟล์หลัก:** `frontend/src/app/menu/main-menu-screen.tsx`, E2E เมนู 3 ไฟล์ และ skill/KMS
- **ตรวจ:** TypeScript และ Playwright 3 tests ผ่าน; ตรวจ default top, สลับ left และจำค่าหลังรีโหลด รวม Light/Dark × 4 ขนาดจอ

### 2026-09-11 — ชื่อหมวดเมนูเป็นภาษาไทยล้วน

- **[UI/UX]** เอาข้อความอังกฤษในวงเล็บออกจากชื่อหมวดหลักทั้ง 9 ระบบตามคำสั่งลุงจืด ชื่ออังกฤษในโหมดภาษาอังกฤษยังอยู่ตามเดิม
- **ไฟล์หลัก:** `frontend/src/lib/menu-data.ts` และเทสต์ชื่อหมวด; ปรับผัง KMS และ `ui-scale-polish` ให้ตรงกัน
- **ตรวจ:** Vitest เฉพาะเมนู 37 ข้อและ Playwright 2 tests ผ่าน; ตรวจ Light/Dark × 4 ขนาดจอ

### 2026-09-11 — เมนูอัปเกรดจาก Champ ผสานงานใหม่

- **[UI/UX]** จัด 9 ระบบเดิม เพิ่ม 69 เมนูเป็น 223 รายการ เก็บ 154 id/route เดิมครบ แยกงานอนุมัติ เช็ค สินค้าชุด และผ่านบัญชี พร้อมค้นด้วยชื่อเดิมจาก Champ
- **[Fix]** หน้า “รอพัฒนา” ใช้คำอธิบายไทย; ทะเบียนเลขเครื่องที่ runtime API 404 แสดง pending จนเชื่อมพร้อม ป้องกันเปิดจอที่ดึงข้อมูลไม่ได้
- **ไฟล์หลัก:** `frontend/src/lib/menu-data.ts`, `menu-icons.ts`, `menu-screen-status.ts`, `frontend/src/app/menu/main-menu-screen.tsx`, `backend/assets/language/languages.tsv` และ skill/KMS
- **ตรวจ:** Vitest 46 ข้อ, tsc ผ่าน, Playwright เมนู 3 tests ผ่าน; light/dark × 4 ขนาดจอ; ไม่ได้ทำ CRUD หรือเปลี่ยนข้อมูลบัญชี
- **หลักฐาน/ข้อจำกัด:** [ผังและตารางเทียบ Champ](docs/kms/decisions/2026-09-11-champ-upgrade-menu-workflows.md) — เมนูเป็นแผนงาน ไม่ใช่รับรองว่า business workflow/รายงานทุกแบบเสร็จแล้ว

### [แม่แบบการบันทึก (Template)]
<!--
### YYYY-MM-DD — <หัวข้อการแก้ไขสั้นกระชับ>
- **ประเภท**: `[Feature]` / `[Fix]` / `[UI/UX]` / `[Refactor]` / `[Deploy]` / `[Docs]`
- **สิ่งที่ทำ**:
  1. <รายละเอียดภาษาไทยชัดเจน คนอายุ 40+ อ่านแล้วเข้าใจทันที>
- **ไฟล์สำคัญ**:
  - `<path/to/file>`
- **ผลการทดสอบ (Evidence)**:
  - <ผลการทดสอบ เช่น ผ่าน vitest ... tests, typecheck 0 errors, curl 200 OK>
-->

### 2026-09-10 — ย้ายธนาคารไปไว้ในข้อมูลหลัก ขยายเป็น "สมุดบัญชี" (Book Bank) รองรับเลือกจากแม่แบบและเพิ่มธนาคารเอง
- **ประเภท**: `[Feature]` `[UI/UX]` `[Refactor]`
- **สิ่งที่ทำ**: ดำเนินการตามคำสั่งของลุงจืด: "ธนาคารให้ย้ายไปไว้ในข้อมูลหลัก และเพิ่ม สาขา เลขที่บัญชี ฯลฯ เปลี่ยนชื่อเป็นสมุดบัญชี ธนาคารสามารถเลือกจาก template และสามารถเพิ่มธนาคารเองได้ด้วย"
  1. **ย้ายตำแหน่งและเปลี่ยนชื่อในเมนู**:
     - ย้ายจากกลุ่ม "ค่าเริ่มต้น" (`defaults`) ไปยังกลุ่ม "ข้อมูลหลัก" (`master`) ภายใต้กลุ่ม "สมุดบัญชี" (`bank-accounts`)
     - เปลี่ยนชื่อเมนูและหน้าจอเป็น **"สมุดบัญชี"** (`Book Bank` / `Bank Accounts`), route `/bookbankscreen`
     - กำหนด redirect อัตโนมัติจาก `/bank` ไปยัง `/bookbankscreen` ป้องกันลิงก์เดิมขาด
  2. **ขยายโครงสร้างข้อมูลสมุดบัญชี (Book Bank Master Data)**:
     - รองรับฟิลด์: รหัสสมุดบัญชี (`bookcode`), ชื่อสมุดบัญชีหลายภาษา (`names`), เลขที่บัญชี (`passbook`), สาขาธนาคาร (`bankbranch`), ชื่อบัญชี (`accountname`), รหัสธนาคาร (`bankcode`), ชื่อธนาคารหลายภาษา (`banknames`), รหัสผังบัญชี (`accountcode`), และรูปภาพ/โลโก้ (`logo` / `images`)
     - เชื่อมต่อ backend API `/payment/bookbank` สำหรับ CRUD แบบสมบูรณ์
  3. **ฟอร์มเลือกและกำหนดธนาคารแบบ 2 ระบบ (Dual Bank Selector)**:
     - **เลือกจากแม่แบบ (Template)**: เลือกจากแม่แบบธนาคารไทยทางการ 20+ ธนาคาร (`thaiBankPresets`) เติมรหัสธนาคาร, ชื่อไทย/อังกฤษ, และโลโก้ความละเอียดสูงให้อัตโนมัติ พร้อมแสดงป้าย BOT Code และตัวอย่างการแสดงผล
     - **กำหนดธนาคารเอง (Custom Bank)**: สามารถสลับไปโหมดกำหนดเองเพื่อกรอกรหัสธนาคาร, ชื่อไทย-อังกฤษ, และอัปโหลดโลโก้ธนาคารผ่าน S3 ได้อย่างอิสระ
  4. **การแสดงผลในตารางและแผงรายละเอียด (Table & Detail View)**:
     - แสดงโลโก้ธนาคาร, รหัสและชื่อสมุดบัญชี, ป้ายเลขที่บัญชี (`font-mono`), สาขา, และชื่อบัญชี
  5. **ปรับปรุงชุดทดสอบ**:
     - ปรับ `menu-data.test.ts`, `menu-icons.test.ts`, `menu-screen-status.test.ts`, `system-setting-screens.test.ts`, และ `route.test.ts` ให้ตรงกับสมุดบัญชี ผ่าน 100% (48 files, 346 tests)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/system-setting-screens.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`, `frontend/src/components/system-settings/utils.ts`, `frontend/src/app/[systemSetting]/page.tsx`
- **ผลการทดสอบ**: Vitest 48 ไฟล์ผ่าน 346/346 (100%); TypeScript compile 0 errors (`tsc --noEmit` pass)

### 2026-09-10 — ตรวจสอบและยกระดับมาตรฐาน CRUD Workbench (datacrud skill) พร้อมติดตั้ง Universal Resizable Splitter ให้ทุกหน้าจอที่ยังเลื่อนไม่ได้ครบ 100%
- **ประเภท**: `[UI/UX]` `[Feature]` `[Refactor]` `[Docs]`
- **สิ่งที่ทำ**: ดำเนินการตามคำสั่งของลุงจืด: "ตรวจ skill crud ใหม่ เพราะบางจอ ยังเลื่อนไม่ได้"
  1. **ฟื้นฟูและยกระดับมาตรฐานกลาง `docs/skills/datacrud/SKILL.md`**:
     - ร่างสัญญา CRUD Workbench ฉบับสมบูรณ์ กำหนดให้ทุกหน้าจอที่มีโครงสร้าง 2 ฝั่ง (Master-Detail, Tree-Detail, Catalog-Selection, List-Editor) ต้องมี `ResizableSplitter` ที่ปรับขนาดได้ เลื่อนได้อย่างอิสระ ไม่ล็อคความกว้างตายตัว
     - กำหนดมาตรฐานความหนาแน่น `.bc-list-*`, การใช้ `<NamesEditor>` หลายภาษา, การมี Dirty Form Guard ป้องกันข้อมูลสูญหาย, และ Pinned Action Header/Footer
  2. **ปรับปรุงคอมโพเนนต์กลาง `ResizableSplitter`**:
     - เพิ่ม breakpoint `"md"` รองรับทั้ง `"md" | "lg" | "xl"` สำหรับจอที่มีจุดตัดการแสดงผลต่างกัน
  3. **ติดตั้งตัวเลื่อนปรับความกว้างครบทุกจอที่เดิมเป็น Fixed Width ("ยังเลื่อนไม่ได้")**:
     - `frontend/src/app/system-settings/company-branch-tree-view.tsx`: โครงสร้างองค์กร (บริษัทและสาขา) ปรับจาก clamp กว้างตายตัว เป็น `--org-sidebar-width` (min 260px, max 620px, default 340px) พร้อม `ResizableSplitter` breakpoint `lg`
     - `frontend/src/app/menu/product-set-screen.tsx`: หน้าจอสินค้าชุด ปรับจาก `md:w-80` เป็น `--product-set-sidebar-width` (min 260px, max 620px, default 340px) พร้อม `ResizableSplitter` breakpoint `md`
     - `frontend/src/app/system-settings/system-settings-screen.tsx`: หน้าจอกลุ่มสินค้า (`productgroup`) และกลุ่มสินค้าย่อย (`productsubgroup`) เพิ่ม `--tree-split-basis` พร้อม `ResizableSplitter` breakpoint `xl`
     - `frontend/src/app/system-settings/system-settings-screen.tsx`: หน้าจอหมวดหมู่สินค้า (`productcategorylist` / `productcategorygroupselectscreen`) เพิ่ม `--category-split-width` พร้อม `ResizableSplitter` breakpoint `md`
     - `frontend/src/app/menu/product-barcode-shelf-screen.tsx`: หน้าจอพิมพ์ป้ายสินค้า ปรับจาก `xl:grid-cols-[minmax(0,1fr)_420px]` เป็น `--shelf-split-basis` พร้อม `ResizableSplitter` breakpoint `xl`
     - `frontend/src/app/menu/manage-shortcuts-screen.tsx`: หน้าจอจัดการทางลัด ปรับจาก `lg:grid-cols-[1fr_360px]` เป็น `--shortcuts-split-basis` พร้อม `ResizableSplitter` breakpoint `lg`
  4. **ชุดทดสอบและเอกสาร**:
     - อัปเดต `frontend/src/components/ui/resizable-splitter.test.ts` ตรวจสอบความครอบคลุมทั้ง 11 หน้าจอ ผ่าน 11/11 tests (100%)
     - อัปเดต `docs/skills/ui-scale-polish/SKILL.md` (หัวข้อ 8.7 Universal Resizable Splitter Standard)
- **ไฟล์สำคัญ**: `docs/skills/datacrud/SKILL.md`, `docs/skills/ui-scale-polish/SKILL.md`, `frontend/src/components/ui/resizable-splitter.tsx`, `frontend/src/components/ui/resizable-splitter.test.ts`, `frontend/src/app/system-settings/company-branch-tree-view.tsx`, `frontend/src/app/menu/product-set-screen.tsx`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/app/menu/product-barcode-shelf-screen.tsx`, `frontend/src/app/menu/manage-shortcuts-screen.tsx`, `README.md`
- **ผลการทดสอบ**: Vitest 48 ไฟล์ผ่าน 346/346 (100%); TypeScript compile 0 errors (`tsc --noEmit` pass); resizable-splitter.test.ts 11/11 tests pass

### 2026-09-10 — ปรับปรุงมาตรฐานตัวแบ่งและปรับความกว้างแนวตั้ง (Universal Resizable Splitter) สวยงามเหมือนหน้าจอคลังสินค้าทั้งโปรเจ็กต์
- **ประเภท**: `[UI/UX]` `[Refactor]`
- **สิ่งที่ทำ**: ปรับปรุงแถบปรับขนาดความกว้างระหว่างคอลัมน์ซ้าย-ขวาตามคำสั่งของลุงจืด: "ไม่สวย ให้ดูหน้าจอ คลัง ไล่แก้ ทั้ง project ให้เหมือนกัน"
  1. **สร้างคอมโพเนนต์กลาง `ResizableSplitter`**:
     - รูปลักษณ์พรีเมี่ยมตามหน้าจอผังคลังสินค้า: เส้นแกนแนวตั้งบางประณีต (`w-0.5 rounded-full bg-border/60` -> ชี้ hover เป็น `bg-primary/50` -> ลากเป็น `bg-primary`) พร้อมปุ่มเม็ดยาลอยตรงกลาง (Floating Pill) และไอคอนกริป `GripVertical` (`h-8 w-3.5` -> hover `h-10` -> resizing `h-12 bg-primary text-primary-foreground`)
     - มี Hitbox กว้างพอสำหรับการใช้เมาส์และหน้าจอสัมผัส (`w-3 cursor-col-resize select-none touch-none -mx-1`)
     - รองรับ Keyboard Accessibility (`ArrowLeft`, `ArrowRight`, `Home`, `End`), ดับเบิ้ลคลิกเพื่อคืนค่าความกว้างเริ่มต้น (Double-click Reset), และค่า ARIA ครบถ้วน
  2. **ปรับใช้มาตรฐานเดียวกันทั้งโปรเจ็กต์ (ครบทั้ง 5 จุด)**:
     - `frontend/src/app/system-settings/warehouse-tree-view.tsx` (หน้าจอผังคลังสินค้า — ต้นแบบ)
     - `frontend/src/app/system-settings/system-settings-screen.tsx` (หน้าต่างตั้งค่าระบบ SettingMasterDetail — จุดที่ลุงจืดทักว่าไม่สวย)
     - `frontend/src/app/system-settings/system-settings-screen.tsx` (หน้าจอแก้ไขสูตรการผลิต BOM)
     - `frontend/src/app/menu/product-screen.tsx` (หน้าจอข้อมูลสินค้า Product List-Detail)
     - `frontend/src/app/menu/product-barcode-screen.tsx` (หน้าจอบาร์โค้ดสินค้า Barcode List-Detail)
  3. **เขียนชุดทดสอบ Unit Tests**:
     - สร้าง `frontend/src/components/ui/resizable-splitter.test.ts` เพื่อรับรองการใช้งานคอมโพเนนต์กลางและพฤติกรรม Accessibility ทุกจุด
  4. **อัปเกรดมาตรฐานระบบ**: อัปเดต `docs/skills/ui-scale-polish/SKILL.md` (หัวข้อ Resizable Splitter)
- **ไฟล์สำคัญ**: `frontend/src/components/ui/resizable-splitter.tsx`, `frontend/src/components/ui/resizable-splitter.test.ts`, `frontend/src/app/system-settings/warehouse-tree-view.tsx`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/app/menu/product-screen.tsx`, `frontend/src/app/menu/product-barcode-screen.tsx`, `docs/skills/ui-scale-polish/SKILL.md`, `README.md`
- **ผลการทดสอบ**: Vitest 48 ไฟล์ผ่าน 342/342 (100%); TypeScript 0 errors; Unit test `resizable-splitter.test.ts` ผ่าน 7/7 (100%)

### 2026-09-10 — ปรับโครงสร้างคลังสินค้าเป็น 2 ระดับ คลัง → ที่เก็บ (ตัดที่วางสินค้า Bins ออกทั้งหมด)
- **ประเภท**: `[UI/UX]` `[Refactor]`
- **สิ่งที่ทำ**: ปรับปรุงหน้าจอคลังสินค้า (`warehouse-tree-view.tsx`) ตามคำสั่งของลุงจืด: "ไม่ต้องมีที่วาง ให้มีแค่ คลัง -> ที่เก็บ"
  1. **ตัดระดับที่วางสินค้า (Bins) ออกทั้งหมด**:
     - ลบคอลัมน์ "ที่วางสินค้า (Bins)" ออกจากตารางที่เก็บสินค้าทางฝั่งขวา
     - ลบหน้าต่างจัดการที่วางสินค้า (Bins Management Modal) และฟอร์ม/สเตตที่เกี่ยวข้องทั้งหมด
     - ทำให้ตารางที่เก็บสินค้าเหลือคอลัมน์ที่ชัดเจน: ลำดับ (#), รหัสที่เก็บสินค้า, ชื่อที่เก็บสินค้า (ไทย), ชื่อที่เก็บสินค้า (EN) และปุ่มลบ (Action)
  2. **ปรับปรุงสเปกและข้อความ**:
     - แก้ไข `frontend/src/lib/system-setting-screens.ts` ให้ title เป็น "คลัง" (`Warehouse → Location`), subtitle เป็น "จัดการคลังสินค้าและที่เก็บสินค้า", และ field label เป็น "ที่เก็บสินค้า" โดยตัดคำว่า "ชั้นวาง" ออกทั้งหมด
  3. **ปรับปรุงชุดทดสอบ E2E**:
     - ปรับปรุง `frontend/e2e/product-warehouse-crud.spec.ts` ให้ทดสอบเฉพาะ CRUD ของ คลังสินค้า และ ที่เก็บสินค้า ผ่าน 100%
  4. **อัปเกรดมาตรฐานระบบ**: บันทึกแบบแผนลงใน `docs/skills/ui-scale-polish/SKILL.md`
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`, `frontend/src/lib/system-setting-screens.ts`, `frontend/e2e/product-warehouse-crud.spec.ts`, `docs/skills/ui-scale-polish/SKILL.md`, `README.md`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; Playwright E2E ผ่าน 100%

### 2026-09-10 — ปรับปรุงหน้าจอคลังสินค้า: ตาราง Datalist เรียงแถวเรียบเสมอ (Single-line), ตรึง Action Bar ด้านล่าง (Pinned Footer) และเพิ่มตัวเลื่อนปรับความกว้าง (Resizable Splitter)
- **ประเภท**: `[UI/UX]` `[Feature]`
- **สิ่งที่ทำ**: ปรับปรุงหน้าจอผังโครงสร้างคลังสินค้า (`warehouse-tree-view.tsx`) ตามคำสั่งของลุงจืด: "datalist เละ เพิ่ม ให้เลื่อนความกว้างได้ด้วย"
  1. **เพิ่มตัวเลื่อนปรับความกว้าง (Accessible Draggable Splitter)**:
     - ติดตั้งแถบปรับขนาดความกว้างระหว่างฝั่งซ้าย (รายชื่อคลัง) และฝั่งขวา (ตารางที่เก็บสินค้า)
     - รองรับการลากด้วยเมาส์และหน้าจอสัมผัส (Pointer capture) พร้อมจำค่าลง `localStorage` (`bc_warehouse_sidebar_width`)
     - ปรับความกว้างได้ระหว่าง 220px - 520px (ค่าเริ่มต้น 280px)
     - รองรับ Double-click คืนค่าเริ่มต้น และรองรับ Keyboard accessibility (`ArrowLeft`, `ArrowRight`, `Home`, `End`)
  2. **ปรับปรุงตารางที่เก็บสินค้า (Datalist Polish)**:
     - **Pinned Bottom Footer**: แยก Action bar (ปุ่ม `+ เพิ่มแถว` และ `บันทึกทั้งหมด`) ออกมาอยู่นอกพื้นที่ Scroll ตรึงติดขอบล่างของการ์ดเสมอ ไม่เลื่อนหลุดหายตามแถวข้อมูล
     - **Single-line Baseline Rhythm**: แยกคอลัมน์ชื่อภาษาไทย และภาษาอังกฤษ ออกจากกันชัดเจน (แสดงคอลัมน์ EN เฉพาะเมื่อเปิดใช้ภาษาอังกฤษ) ทำให้ทุกแถวมีความสูงบรรทัดเรียบเสมอกัน (~40-42px) ไม่โป่งบวม
     - **Sticky Header with Backdrop Blur**: ตรึงหัวตารางด้านบนด้วย backdrop-blur
     - **Subtle Row State Indicators**: ปรับสถานะแถวใหม่/แถวแก้ไขด้วยเส้นขอบซ้าย `border-l-2` และ Badge `NEW`/`MOD` ในคอลัมน์ลำดับ ไม่ย้อมสีพื้นหลังหนาจนรกสายตา
     - **ลบรหัสคลังซ้ำซ้อน**: รายชื่อคลังฝั่งซ้ายแสดงรหัสใน Badge เพียงครั้งเดียว ไม่พิมพ์รหัสซ้ำข้างชื่อ
  3. **ยกเลิกการเปิด Dialog เพิ่มคลังสินค้าอัตโนมัติ**: ยกเลิก `useEffect` ที่สั่งเปิด Dialog ทันทีเมื่อเข้าหน้าจอ เพื่อให้ผู้ใช้เข้าไปดูข้อมูลเดิมได้ก่อนเสมอตามความต้องการของลุงจืด
  4. **อัปเกรดมาตรฐานระบบ**: อัปเดต `docs/skills/ui-scale-polish/SKILL.md` (หัวข้อ 8.6)
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`, `docs/skills/ui-scale-polish/SKILL.md`, `README.md`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; Playwright E2E ผ่าน 100%

### 2026-09-09 — ออกแบบหน้าจอคลังสินค้า: ตารางที่เก็บสินค้าแบบ Editable Table Grid พร้อมบันทึกทีเดียว (Batch Save All)
- **ประเภท**: `[UI/UX]` `[Feature]`
- **สิ่งที่ทำ**: ปรับปรุงหน้าจอคลังสินค้า (`warehouse-tree-view.tsx`) ตามคำสั่งของลุงจืด: "แก้ใหม่ พอเลือกคลัง ข้างในเป็นที่เก็บแบบ table save พร้อมกันทีเดียว"
  1. **ฝั่งซ้าย (~280px)**: รายการคลังสินค้าแบบ Compact List แสดงรหัส, ชื่อ, จำนวนที่เก็บ (`ลูก X`), ปุ่มเพิ่ม/แก้ไข/ลบคลังสินค้า
  2. **ฝั่งขวา (1fr)**: ตารางแก้ไขข้อมูลที่เก็บสินค้าโดยตรง (Inline Editable Table Grid):
     - แก้ไขรหัส (`code`) และชื่อหลายภาษา (`names.th`, `names.en`) ได้ในช่องตารางทันที
     - ปุ่ม `+ เพิ่มแถว` สำหรับเพิ่มแถวใหม่ต่อเนื่องหลายรายการโดยไม่ต้องรอบันทึกทีละตัว
     - สถานะสีแยกชัดเจน: แถวใหม่ (`isNew: true`), แถวที่แก้ไข (`isModified: true`)
     - ปุ่มเด่น `บันทึกทั้งหมด` (Save All) รวบรวมการเพิ่ม แก้ไข และลบ ยิง Concurrent Requests ผ่าน API (`POST`, `PUT`, `DELETE`) ในคราวเดียว
     - ระบบตรวจสอบความถูกต้อง (Validation) เตือนรหัส/ชื่อว่าง หรือรหัสซ้ำกันเองก่อนบันทึก
     - หน้าต่างจัดการที่วางสินค้า (Bins Management Modal) แยกออกเป็น Dialog ย่อยมาตรฐาน `.dialog-backdrop`
  3. **อัปเกรดมาตรฐานระบบ**: บันทึกแบบแผนลงใน `docs/skills/ui-scale-polish/SKILL.md` (หัวข้อ 8.6) ตามกฎ Mandatory Skill Upgrade
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`, `frontend/e2e/product-warehouse-crud.spec.ts`, `docs/skills/ui-scale-polish/SKILL.md`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; Playwright E2E `product-warehouse-crud.spec.ts` ผ่านครบวงจร CRUD คลัง/ที่เก็บ/ที่วางสินค้า 100%

### 2026-09-09 — ออกแบบ UX/UI หน้าจอคลังสินค้าใหม่ (Master-Detail 3 คอลัมน์ รองรับที่เก็บสินค้าจำนวนมาก)
- **ประเภท**: `[UI/UX]` `[Refactor]`
- **สิ่งที่ทำ**: ออกแบบและปรับปรุง UX/UI หน้าจอคลังสินค้า (`warehouse-tree-view.tsx`) ใหม่ทั้งหมดตามคำสั่งของลุงจืด เพื่อแก้ปัญหาเมื่อคลังสินค้ามีที่เก็บสินค้าจำนวนมาก (50–200+ แห่ง) โดยเปลี่ยนจาก Tree View แคบๆ ซ้อน 3 ชั้น เป็นสถาปัตยกรรม **Master-Detail 3 คอลัมน์**:
  1. **คอลัมน์ซ้าย (~260px)**: คลังสินค้า (Warehouses) เป็นการ์ดกระชับ พร้อม Badge จำนวนที่เก็บ และปุ่ม Action ประจำแถว
  2. **คอลัมน์กลาง (1fr กว้างสุด)**: ศูนย์จัดการที่เก็บสินค้าและที่วางสินค้า (Location Workspace) เป็นพื้นที่หลัก มี **ช่องค้นหาด่วนแบบ Real-time (Instant Search)**, สถิติสรุปจำนวนที่เก็บและที่วาง, ปุ่มคุมการแสดงผล (กางทั้งหมด/ยุบทั้งหมด), ตารางรายการที่เก็บสินค้าแบบ Single-line Baseline Rhythm อ่านง่าย สบายตา เหมาะกับคนไทยอายุ 40+ พร้อมตารางย่อยแสดงที่วางสินค้า (Bins)
  3. **คอลัมน์ขวา (~380px)**: ฟอร์มจัดการข้อมูล (Active Form Panel) ปรับปรุงให้ใช้งานง่าย พร้อมบันทึกได้ทันที
  - อัปเกรดมาตรฐาน UX/UI ลงใน `docs/skills/ui-scale-polish/SKILL.md` (หัวข้อ 8.5) ตามกฎ Mandatory Skill Upgrade
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`, `docs/skills/ui-scale-polish/SKILL.md`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบฟิลด์ "บริษัทที่ใช้คลังนี้ได้" ออกจากฟอร์มคลังสินค้า (Warehouse)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบส่วนกำหนด "บริษัทที่ใช้คลังนี้ได้" (CompanyScopePicker) ออกจากฟอร์มสร้าง/แก้ไขคลังสินค้า (Warehouse) ในหน้าจอผังโครงสร้างคลังสินค้า (`warehouse-tree-view.tsx`) ตามคำสั่งของลุงจืด เพื่อให้ฟอร์มคลังสินค้าเรียบง่ายและไม่ซับซ้อนเกินจำเป็น
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบฟิลด์เพิ่มเติม (ประเภทสินค้า, วัตถุอันตราย, ลำดับ, กฎเข้า-ออก, ขอบเขตบริษัท) ออกจากฟอร์มที่เก็บสินค้า
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบฟิลด์ที่ไม่ได้ใช้งานออกจากฟอร์มสร้าง/แก้ไขที่เก็บสินค้า (Location) ในหน้าจอผังโครงสร้างคลังสินค้า (`warehouse-tree-view.tsx`) ได้แก่ ประเภทสินค้าที่อนุญาต, ประเภทวัตถุอันตราย, ลำดับการจัดเรียง, กฎการเข้า-ออก (Allow Putaway/Pick/Blocked), และขอบเขตบริษัทที่ใช้ที่เก็บสินค้านี้ได้ เพื่อให้ฟอร์มเรียบง่าย กระชับ เหมาะสมกับการใช้งานของคนไทยอายุ 40+ คงเหลือเฉพาะรหัสและชื่อที่เก็บสินค้าหลายภาษา
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบฟิลด์ประเภทที่เก็บสินค้าออกจากฟอร์มที่เก็บสินค้า (Warehouse Location)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบส่วนเลือก "ประเภทที่เก็บสินค้า" (RadioChipPicker: จัดเก็บ, หยิบสินค้า, รับสินค้า, ตรวจสอบคุณภาพ, งานระหว่างทำ, สินค้าชำรุด, ระหว่างขนส่ง) ออกจากฟอร์มสร้าง/แก้ไขที่เก็บสินค้า (Location) ในหน้าจอผังโครงสร้างคลังสินค้า (`warehouse-tree-view.tsx`) ตามคำสั่งของลุงจืด เพื่อลดความซ้ำซ้อนและกระชับฟอร์ม
- **ไฟล์สำคัญ**: `frontend/src/app/system-settings/warehouse-tree-view.tsx`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — เพิ่มปุ่ม "เพิ่มธนาคารไทย" พร้อมแม่แบบธนาคารและโลโก้ (Bank Templates)
- **ประเภท**: `[Feature]` `[UI/UX]`
- **สิ่งที่ทำ**: เพิ่มปุ่ม "เพิ่มธนาคารไทย" บนหน้าจอธนาคาร (`/bank`) ในส่วน Action Toolbar และ Empty State พร้อมหน้าต่าง Dialog แม่แบบธนาคารไทย (`ThaiBankTemplateDialog`) รวบรวม 20 ธนาคารในประเทศไทย (KBANK, SCB, KTB, BBL, BAY, TTB, GSB, BAAC, GHB, KKP, CIMB, TISCO, UOB, LHB, ICBC, TCRB, IBANK, CITI, HSBC, PromptPay) ครบถ้วนทั้งชื่อภาษาไทย ภาษาอังกฤษ รหัส BOT สีประจำธนาคาร และไฟล์โลโก้ความละเอียดสูงที่จัดเก็บแบบ Local Static Assets (`frontend/public/banks/*.png`); รองรับการเลือกเพิ่มหลายธนาคารพร้อมกัน (Bulk Add) มีระบบตรวจจับธนาคารที่มีอยู่ในระบบแล้วเพื่อป้องกันการเพิ่มซ้ำ; พร้อมทั้งเพิ่มตัวเลือกเติมข้อมูลอัตโนมัติจากแม่แบบในฟอร์มเพิ่มธนาคารเดี่ยว
- **ไฟล์สำคัญ**: `frontend/public/banks/*.png`, `frontend/src/lib/thai-banks.ts`, `frontend/src/lib/thai-banks.test.ts`, `frontend/src/components/logo-avatar.tsx`, `frontend/src/components/authenticated-image.tsx`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/e2e/bank-template.spec.ts`
- **ผลการทดสอบ**: Vitest 47 ไฟล์ผ่าน 335/335 (100%); TypeScript 0 errors; Playwright E2E `bank-template.spec.ts` และ `bank-crud.spec.ts` (ตรวจสอบบันทึกและ query MongoDB จริงแบบ Live) ผ่าน 100%; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — พัฒนาหน้าจอธนาคาร (/bank) เชื่อมต่อ API CRUD และปลดสถานะรอพัฒนา
- **ประเภท**: `[Feature]` `[UI/UX]`
- **สิ่งที่ทำ**: พัฒนาหน้าจอ "ธนาคาร" (`/bank`) เชื่อมต่อ Backend API `/payment/bankmaster` แบบเต็มระบบ CRUD (สร้าง, ค้นหา/แสดงรายการ, แก้ไข, ลบแบบ Soft Delete ตามมาตรฐานระบบ) พร้อมรองรับอัปโหลดโลโก้ธนาคาร (`logo`), ปลดป้าย "รอพัฒนา" ออกจากหน้าจอเมนูและทางลัด (`isMenuScreenPending("/bank") === false`); แก้ไข header `x-bc-backend-url` ใน `use-screen-actions.ts` ป้องกัน HTTP 400; ปรับปรุงสถิติสถานะหน้าจอ (เชื่อมต่อแล้วเพิ่มเป็น 19 หน้าจอ, รอพัฒนาลดเหลือ 152 หน้าจอ จากยอดรวม 171 หน้าจอ)
- **ไฟล์สำคัญ**: `frontend/src/lib/system-setting-screens.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/lib/use-screen-actions.ts`, `frontend/src/lib/system-setting-screens.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.test.ts`, `frontend/e2e/bank-crud.spec.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 331/331 (100%); TypeScript 0 errors; Playwright E2E `menu-consistency.spec.ts` และ `bank-crud.spec.ts` (ตรวจสอบ CRUD ใน MongoDB จริงแบบ Live) ผ่าน 100%; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — รวมกลุ่มเมนูในหมวดค่าเริ่มต้นเป็นระดับเดียว (7 เมนู)
- **ประเภท**: `[UI/UX]` `[Refactor]`
- **สิ่งที่ทำ**: ยุบโครงสร้างกลุ่มย่อยทั้งหมดในหมวด "ค่าเริ่มต้น" (`defaults`) รวมเป็นกลุ่มเดียว `defaults` ("ค่าเริ่มต้น") เพื่อให้เมนูทั้งหมด 7 เมนู (หน่วยนับสินค้า, กลุ่มสินค้า, ยี่ห้อสินค้า, คลัง, กลุ่มลูกหนี้, กลุ่มเจ้าหนี้, ธนาคาร) แสดงผลเป็นระดับเดียวโดยตรง (Single level) ไม่มีการซ้อน accordion ย่อย; ปรับปรุง unit test รองรับ groupOrder ใหม่ คงจำนวนเมนูรวมทั้งระบบไว้ที่ 171 เมนู
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-data.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบกลุ่มเมนู "เอกสารและบิล" (4 เมนู, ยอดรวมเหลือ 171 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบกลุ่มเมนู "เอกสารและบิล" (`sales-documents`) ออกจากหมวดค่าเริ่มต้นทั้งหมด 4 เมนู ได้แก่ รูปแบบเอกสาร (`/docformat`), ออกแบบบิล (`/billdesign`), ตั้งค่าใบกำกับภาษีอิเล็กทรอนิกส์ (`/etaxsetting`), และขอใบกำกับภาษีออนไลน์ (`/taxinvoicerequestsetting`); ปลด icon mapping และปรับปรุง unit/integration tests ยอดเมนูรวมปรับลดจาก 175 เหลือ 171 เมนู (สถานะรอพัฒนาเหลือ 153 เมนู, หน้าจอเชื่อมต่อคงเดิม 18 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/lib/menu-data.test.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบกลุ่มเมนู "อนุมัติ" (7 เมนู, ยอดรวมเหลือ 175 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบกลุ่มเมนู "อนุมัติ" (`approval`) ออกจากหมวดค่าเริ่มต้นทั้งหมด 7 เมนู ได้แก่ ประเภทซื้อ (`/purchasetypescreen`), อนุมัติใบสั่งซื้อ (`/poapprovalsettingscreen`), ประเภทใบเสนอราคา (`/quotationtypescreen`), อนุมัติใบเสนอราคา (`/qtapprovalsettingscreen`), ประเภทใบสั่งขาย (`/saleordertypescreen`), อนุมัติใบสั่งขาย (`/soapprovalsettingscreen`), และอนุมัติรายจ่ายและสมุดรายวัน (`/expenseapprovalsettingscreen`); ปลด icon mapping และปรับปรุง unit/integration tests ยอดเมนูรวมปรับลดจาก 182 เหลือ 175 เมนู (สถานะรอพัฒนาเหลือ 157 เมนู, หน้าจอเชื่อมต่อคงเดิม 18 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/lib/menu-data.test.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบกลุ่มเมนู "หน้าร้าน POS" (4 เมนู, ยอดรวมเหลือ 182 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบกลุ่มเมนู "หน้าร้าน POS" (`sales-pos`) ออกจากหมวดค่าเริ่มต้นทั้งหมด 4 เมนู ได้แก่ ตั้งค่าเครื่องขายหน้าร้าน POS (`/possetting`), รูป/สื่อหน้าจอขาย (`/posmedia`), เครื่องพิมพ์ใบเสร็จเทอร์มัล (`/posprintersetting`), และสีสำหรับงานขาย (`/colorscreen`); ปลด icon mapping และปรับปรุง unit/integration tests ยอดเมนูรวมปรับลดจาก 186 เหลือ 182 เมนู (สถานะรอพัฒนาเหลือ 164 เมนู, หน้าจอเชื่อมต่อคงเดิม 18 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/lib/menu-data.test.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบกลุ่มเมนู "สมาชิก คูปอง และโปรโมชัน" (4 เมนู, ยอดรวมเหลือ 186 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบกลุ่มเมนู "สมาชิก คูปอง และโปรโมชัน" (`sales-loyalty`) ออกจากหมวดค่าเริ่มต้นทั้งหมด 4 เมนู ได้แก่ ตั้งค่าคะแนนสะสม (`/pointsetting`), ตั้งค่าคูปอง (`/couponsetting`), โปรโมชั่น (`/promotionscreen`), และรอบซื้อลูกค้าประจำ (`/customerpurchasecycle`); ปลด icon mapping และปรับปรุง unit/integration tests ยอดเมนูรวมปรับลดจาก 190 เหลือ 186 เมนู (สถานะรอพัฒนาเหลือ 168 เมนู, หน้าจอเชื่อมต่อคงเหลือ 18 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/lib/menu-data.test.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบ 3 เมนูสถานะรอพัฒนาในกลุ่มการรับเงินและบัญชีธนาคาร (ยอดรวมเหลือ 190 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบ 3 เมนูสถานะรอพัฒนาออกจากกลุ่ม "การรับเงินและบัญชีธนาคาร" (`sales-payment-banking`) ในหมวดค่าเริ่มต้น ได้แก่ ผู้ให้บริการรับเงิน QR (`/qrprovider`), อัตราแลกเปลี่ยน (`/exchangerate`), และกฎจับคู่บัญชีอัตโนมัติ (`/banking/rules`) คงเหลือเฉพาะเมนู "ธนาคาร" (`/bank`); ปลด icon mapping และปรับปรุง unit/status tests ยอดเมนูรวมปรับลดจาก 193 เหลือ 190 เมนู (สถานะรอพัฒนาเหลือ 171 เมนู, หน้าจอเชื่อมต่อคงเดิม 19 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบ 6 เมนูย่อยในกลุ่มสินค้าและบาร์โค้ด คงเหลือเฉพาะสินค้าและบาร์โค้ด (ยอดรวมเหลือ 193 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบ 6 เมนูย่อยออกจากกลุ่ม "สินค้าและบาร์โค้ด" (`products`) ในหมวดข้อมูลหลัก ได้แก่ สินค้าบริการ (`/serviceproduct`), สินค้าไม่นับสต็อก (`/nonstockproduct`), ข้อมูลเสริมสินค้า (`/productextension`), จัดหมวดสินค้า (`/productcategorygroupselectscreen`), สินค้าชุด (`/productset`), และสูตรผลิต (`/productbom`) คงเหลือไว้เฉพาะ "สินค้า" (`/product`) และ "บาร์โค้ด" (`/productbarcode`); ปลดไอคอน, custom screen routes, dispatcher ในแท็บงาน และปรับปรุง unit/integration tests ยอดเมนูรวมปรับลดจาก 199 เหลือ 193 เมนู (สถานะรอพัฒนาเหลือ 174 เมนู, หน้าจอเชื่อมต่อ 19 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/lib/menu-screen-status.ts`, `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/lib/menu-data.test.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน (100%); `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — ลบกลุ่มช่องทางขาย/ราคา/ขนส่ง และกลุ่มเชื่อมต่อตลาดออนไลน์ (7 เมนู)
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**: ลบกลุ่มเมนู "ช่องทางขาย ราคา และขนส่ง" (`sales-channel-pricing` 3 เมนู: `/salechannelscreen`, `/transportchannelscreen`, `/channelprice`) ออกจากหมวดค่าเริ่มต้น, ลบกลุ่มเมนู "เชื่อมข้อมูลตลาดออนไลน์" (`marketplace` 3 เมนู: `/marketplace/shopee`, `/marketplace/lazada`, `/marketplace/tiktok`) ออกจากหมวดข้อมูลหลัก, และลบเมนู "ดึงคำสั่งซื้อจากร้านค้าออนไลน์" (`/transaction/marketplaceorder`) ออกจากหมวดงานประจำ › ขาย; ปลดไอคอนและการ dispatch แท็บที่ไม่ใช้ออก ยอดเมนูรวมปรับลดจาก 206 เหลือ 199 เมนู (สถานะรอพัฒนาเหลือ 176 เมนู)
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`, `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/lib/menu-screen-status.ts`, `frontend/src/lib/menu-data.test.ts`, `frontend/src/lib/menu-icons.test.ts`, `frontend/src/lib/menu-screen-status.test.ts`, `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ**: Vitest 46 ไฟล์ผ่าน 329/329; TypeScript 0 errors; Playwright e2e `menu-consistency.spec.ts` ผ่าน; `gen-code-map.ps1 -Check` ผ่านสมบูรณ์

### 2026-09-09 — แก้ชื่อเมนูให้ตรงกัน ติดป้ายรอพัฒนา และรวมกลุ่มสินค้าในค่าเริ่มต้น
- **ประเภท**: `[Fix]` `[UI/UX]`
- **สิ่งที่ทำ**: ซิงค์ชื่อไทยในผังเมนู คำแปล backend และหน้าจอตั้งค่า; ป้องกัน dictionary เก่าทับชื่อไทย เปลี่ยนชื่อเมนูตรวจสอบผู้ใช้งานให้ตรงกับรายงานสิทธิ์ปัจจุบัน แสดงป้าย “รอพัฒนา” สำหรับเมนูที่ยังไม่มีหน้าจอ โดยคงสิทธิ์การเปิดเดิม; รวมกลุ่มเมนูในหมวดค่าเริ่มต้น (Defaults) โดยเปลี่ยน “จัดกลุ่มสินค้า” เป็น “สินค้า” (`product-setup`) และย้าย “ยี่ห้อสินค้า” กับ “คลัง” เข้ามารวมอยู่ในกลุ่มนี้ พร้อมยุบกลุ่มเดิมที่ว่างลง
- **ไฟล์สำคัญ**: `frontend/src/lib/menu-data.ts`, `system-setting-screens.ts`, `menu-screen-status.ts`, `frontend/src/app/menu/menu-pending-badge.tsx`, หน้ารายการเมนู/ทางลัด, `backend/assets/language/languages.tsv`, `frontend/e2e/menu-consistency.spec.ts`, `docs/skills/ui-scale-polish/SKILL.md`
- **ผลการทดสอบ**: Vitest ผ่าน; TypeScript ผ่าน; Playwright ตรวจเมนูด้วยบัญชี Demo ผ่านทั้ง 1600/1280/1024/768 × light/dark กดสลับธีมจริง ตรวจ hover/focus, ป้ายไม่ถูกตัด, การนำทางและชื่อหน้าปลายทาง ไม่มี console error ในหน้าเมนู; ไม่สร้าง/แก้ข้อมูลธุรกิจ

### 2026-09-09 — ลบกลุ่มเมนู "ร้านอาหาร/คาเฟ่" ทั้งหมด 7 เมนูออกจากระบบ
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบกลุ่มเมนูร้านอาหาร/คาเฟ่ออกจากระบบ**:
     - ลบกลุ่มเมนู `ร้านอาหาร/คาเฟ่` (`restaurant-setup`) ทั้งหมด 7 เมนู ได้แก่ `โซน` (`/zonegroupselectscreen`), `โต๊ะ` (`/tablegroupselectscreen`), `ผังโต๊ะ` (`/tablemapgroupselectscreen`), `ครัว` (`/kitchengroupselectscreen`), `ตั้งค่าเครื่องสั่งอาหาร` (`/ordertemplatsetting`), `ตั้งค่าการสั่งอาหาร` (`/ordersetting`), `สั่งอาหารด้วย QR` (`/qrcodeordergroupselectscreen`) ออกจาก `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอนทั้ง 7 เส้นทางใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 213 เหลือ 206 เมนู
     - อัปเดต unit tests ใน `menu-data.test.ts` เอา `restaurant-setup` ออกจาก `groupOrder`
  2. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/menu-data.test.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (32/32 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "รุ่นสินค้า" (Model) และตัดการเชื่อมโยงจากระบบอื่นอย่างสมบูรณ์
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบเมนูและไอคอนรุ่นสินค้า**:
     - ลบรายการเมนู `รุ่นสินค้า` (`/mastermodelscreen`) ออกจากกลุ่มรายละเอียดประกอบสินค้าใน `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอน `"/mastermodelscreen": "design"` ใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 214 เหลือ 213 เมนู
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**:
     - ลบคอนฟิก `master_model_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts`
  3. **ตัดการเชื่อมโยงจากหน้าจอสินค้า (Product)**:
     - ลบฟิลด์เลือก `model` ออกจากแถบจัดหมวดหมู่สินค้าใน `frontend/src/app/menu/tab-product-classification.tsx`
     - ลบฟิลด์ `model` ออกจาก `classificationFields`, `clearFields` และการแสดงผลรายละเอียดสินค้าใน `frontend/src/app/menu/product-screen.tsx`
  4. **ตัดการเชื่อมโยง API Proxy Master Picker**:
     - ลบ endpoint mapping `model: "/aicloud/model"` ออกจาก `frontend/src/app/api/product-barcode/master/[master]/route.ts`
     - ลบ `| "model"` ออกจากประเภท `MasterName` ใน `frontend/src/lib/product-barcode/api.ts`
  5. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `frontend/src/app/menu/tab-product-classification.tsx`
  - `frontend/src/app/menu/product-screen.tsx`
  - `frontend/src/app/api/product-barcode/master/[master]/route.ts`
  - `frontend/src/lib/product-barcode/api.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (32/32 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "คุณลักษณะสินค้า" และกลุ่มเมนู "สี ไซซ์ และตัวเลือก" ออกจากระบบ
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบเมนูและกลุ่มเมนูออกจากระบบ**:
     - ลบรายการเมนู `คุณลักษณะสินค้า` (`/mastercategoryscreen`) ออกจากกลุ่มรายละเอียดประกอบสินค้าใน `frontend/src/lib/menu-data.ts`
     - ลบกลุ่มเมนู `สี ไซซ์ และตัวเลือก` (`product-sku-options`) ทั้งกลุ่ม ซึ่งประกอบด้วย `สีสินค้า` (`/productcolor`), `ไซซ์/ขนาดสินค้า` (`/productsize`), `ชุดตัวเลือกสินค้า` (`/productvariantmatrix`) ออกจาก `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอนทั้ง 4 เส้นทางใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 218 เหลือ 214 เมนู
     - อัปเดต unit tests ใน `menu-data.test.ts` ให้สอดคล้องกับโครงสร้างเมนูใหม่
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**:
     - ลบคอนฟิก `master_category_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts`
  3. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/menu-data.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (32/32 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "รูปทรงสินค้า", "ระดับสินค้า", "เกรดสินค้า", "มิติสินค้า" และตัดการเชื่อมโยงจากระบบอื่นอย่างสมบูรณ์
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบ 4 เมนูและไอคอนออกจากระบบ**:
     - ลบรายการเมนู `ขนาด/มิติสินค้า` (`/productdimension`), `เกรดสินค้า` (`/mastergradescreen`), `ระดับสินค้า` (`/masterclassscreen`), `รูปทรงสินค้า` (`/masterdesignscreen`) ออกจากกลุ่มข้อมูลหลักใน `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอนทั้ง 4 เส้นทางใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 222 เหลือ 218 เมนู
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**:
     - ลบคอนฟิก `productdimension`, `master_class_screen`, `master_design_screen`, `master_grade_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts`
  3. **ตัดการเชื่อมโยงจากหน้าจอสินค้า (Product)**:
     - ลบฟิลด์เลือก `class`, `design`, `grade` ออกจากแถบจัดหมวดหมู่สินค้าใน `frontend/src/app/menu/tab-product-classification.tsx`
     - ลบฟิลด์ `class`, `design`, `grade` ออกจาก `classificationFields`, `clearFields` และการแสดงผลรายละเอียดสินค้าใน `frontend/src/app/menu/product-screen.tsx`
  4. **ตัดการเชื่อมโยง API Proxy Master Picker**:
     - ลบ endpoint mapping `class`, `design`, `grade` ออกจาก `frontend/src/app/api/product-barcode/master/[master]/route.ts`
     - ลบ `| "class" | "design" | "grade"` ออกจากประเภท `MasterName` ใน `frontend/src/lib/product-barcode/api.ts`
  5. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `frontend/src/app/menu/tab-product-classification.tsx`
  - `frontend/src/app/menu/product-screen.tsx`
  - `frontend/src/app/api/product-barcode/master/[master]/route.ts`
  - `frontend/src/lib/product-barcode/api.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (33/33 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "รูปแบบสินค้า" และตัดการเชื่อมโยงจากระบบอื่นอย่างสมบูรณ์
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบเมนูและไอคอนรูปแบบสินค้า**: ลบรายการเมนู `รูปแบบสินค้า` (`/masterpatternscreen`) ออกจากกลุ่มข้อมูลหลักใน `frontend/src/lib/menu-data.ts`, ลบไอคอนใน `frontend/src/lib/menu-icons.ts` และอัปเดตจำนวนเมนูใน `menu-icons.test.ts` จาก 223 เหลือ 222 เมนู
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**: ลบคอนฟิก `master_pattern_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts` ไม่ให้เข้าถึงผ่าน `/[systemSetting]`
  3. **ตัดการเชื่อมโยงจากหน้าสินค้า (Product)**:
     - ลบช่องเลือก `pattern` ออกจากแถบจัดหมวดหมู่สินค้าใน `frontend/src/app/menu/tab-product-classification.tsx`
     - ลบฟิลด์ `pattern` ออกจาก `classificationFields`, `clearFields` และการแสดงผลรายละเอียดสินค้าใน `frontend/src/app/menu/product-screen.tsx`
  4. **ตัดการเชื่อมโยง API Proxy Master Picker**:
     - ลบ endpoint mapping `pattern: "/aicloud/pattern"` ออกจาก `frontend/src/app/api/product-barcode/master/[master]/route.ts`
     - ลบ `| "pattern"` ออกจากประเภท `MasterName` ใน `frontend/src/lib/product-barcode/api.ts`
  5. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `frontend/src/app/menu/tab-product-classification.tsx`
  - `frontend/src/app/menu/product-screen.tsx`
  - `frontend/src/app/api/product-barcode/master/[master]/route.ts`
  - `frontend/src/lib/product-barcode/api.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (33/33 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ตั้งกฎห้ามอ้างอิงบุคคลภายนอก (ซอฟต์แวร์คู่แข่ง) และชำระล้างเอกสารทั้งระบบ พร้อม Pre-commit Guard
- **ประเภท**: `[Docs]` `[Tooling]` `[Compliance]`
- **สิ่งที่ทำ**:
  1. **ตั้งกฎ Zero Reference Policy ใน AGENTS.md**: สั่งเด็ดขาดห้ามมีชื่อ ยี่ห้อ หรือการอ้างอิงถึงซอฟต์แวร์ภายนอกในโค้ด, คอมเมนต์, ชื่อไฟล์, ตัวแปร, หน้าจอ UI, commit message, และเอกสารทุกชนิด เพื่อป้องกันปัญหาลิขสิทธิ์และเครื่องหมายการค้า โดยให้ใช้คำกลาง ("มาตรฐานโปรแกรมบัญชีในตลาด") แทน
  2. **ลบโฟลเดอร์เอกสารวิจัยคู่แข่งเดิม**: ลบโฟลเดอร์เอกสารวิจัยเดิม 2 โฟลเดอร์และ handoff เก่า รวม 23 ไฟล์ออกจาก repository
  3. **ชำระล้างเอกสารทั้งระบบ**: เปลี่ยนชื่อไฟล์และปรับถ้อยคำใน `docs/kms/` (บทความ 19, ADRs, ดัชนี), `docs/README.md`, `docs/skills/ui-scale-polish/SKILL.md` และ `docs/handoff/` ให้เป็นคำกลางทั้งหมด
  4. **เพิ่มระบบตรวจจับอัตโนมัติ (Git Pre-commit Guard)**: อัปเดต `.githooks/pre-commit` ให้สแกนทุกไฟล์ที่ staged หากพบคำต้องห้ามจะสกัดและปฏิเสธ commit ทันที
- **ไฟล์สำคัญ**:
  - `AGENTS.md`
  - `.githooks/pre-commit`
  - `docs/kms/19-menu-coverage-market-standard.md`
  - `docs/kms/decisions/2026-09-08-menu-parity-market-standard.md`
  - `docs/README.md`
- **ผลการทดสอบ (Evidence)**:
  - สแกนทั้ง repository: ปลอดคำต้องห้าม 100%
  - ทดสอบ Pre-commit Guard: สกัดคำต้องห้ามสำเร็จทุกกรณี (`ExitCode: 1`)

### 2026-09-09 — ตั้งกฎและระบบ Activity Log ใน README.md พร้อม Git Pre-commit Hook
- **ประเภท**: `[Docs]` `[Tooling]`
- **สิ่งที่ทำ**:
  1. สร้างไฟล์ `README.md` ที่ root ของโปรเจกต์ เพื่อเป็นหน้าแรกของ Repository บน GitHub และเป็นจุดบันทึกประวัติงานหลัก
  2. เพิ่มหัวข้อกฎใน `AGENTS.md`: บังคับให้ AI ทุกตัว (Gemini, Claude, Codex) ต้องบันทึกสิ่งที่แก้ลงใน `README.md` และ commit พร้อมโค้ดทุกครั้ง
  3. เพิ่มตัวตรวจจับใน Git Pre-commit Hook (`.githooks/pre-commit`): หากมีการ stage โค้ดใน `frontend/src` หรือ `backend` แต่ไม่มี `README.md` ระบบจะแจ้งเตือนและปฏิเสธ commit เพื่อป้องกันการลืม
- **ไฟล์สำคัญ**:
  - `README.md`
  - `AGENTS.md`
  - `.githooks/pre-commit`
- **ผลการทดสอบ (Evidence)**:
  - ทดสอบ Pre-commit Hook ดักจับกรณีไม่มี README.md ได้ถูกต้อง
  - `npm run hooks:install` อัปเดต hook ลง `.git/hooks/` สำเร็จ

### 2026-09-09 — ย้าย "จัดหมวดสินค้า" ไปข้อมูลหลัก, บาร์โค้ดในหมวด, Tree View UX & Deploy Production
- **ประเภท**: `[UI/UX]` `[Feature]` `[Deploy]`
- **สิ่งที่ทำ**:
  1. **ย้ายเมนูจัดหมวดสินค้า**: ย้ายจากกลุ่ม "ตั้งค่าระบบ" (`/defaults`) ไปไว้ที่ "ข้อมูลหลัก › สินค้าและบาร์โค้ด" (`/product-category`) ต่อจากเมนูบาร์โค้ด
  2. **เปลี่ยนเป็นเพิ่มบาร์โค้ด**: ปรับระบบจัดการรายการในหมวดสินค้า จากเดิมที่เลือกสินค้า ให้เป็นการเลือกและค้นหา "บาร์โค้ด" เข้าหมวดแทน ผ่าน `POST /api/product-barcode/list`
  3. **ยกระดับ Tree View & Row Actions**: ปรับดีไซน์ Tree View ให้ตรงกับหน้า Group Tree View โดยมีปุ่มเลือกกลุ่มแบบเต็มจอ (Full-width CSS Grid) และมี Row Actions (แก้ไข, ลบ, จัดลำดับ, จัดการบาร์โค้ด) ประจำแถว
  4. **อัปเดต UI Skill**: บันทึกแบบแผน CSS Grid Full-width Selector และ Anti-pattern ลงใน `docs/skills/ui-scale-polish/SKILL.md` (§8.4)
  5. **Deploy ขึ้น Cloud Production**: Build frontend Docker image และ deploy ไปยังเซิร์ฟเวอร์ DigitalOcean `https://account.bcaicloud.com/` พร้อม push ขึ้น GitHub branch `dev`
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/app/menu/product-category-screen.tsx`
  - `docs/skills/ui-scale-polish/SKILL.md`
- **ผลการทดสอบ (Evidence)**:
  - Frontend Typecheck: 0 errors
  - Vitest: 45 test files, 324/324 tests passed
  - Pre-push hook & Code map check: ผ่าน
  - Production Health Check: `https://account.bcaicloud.com/` ตอบ 200 OK, Google Sign-in ใช้งานได้ปกติ

---

## 📚 แผนที่เอกสารและการเรียนรู้ระบบ

โปรเจกต์นี้มีเอกสารและคลังความรู้ที่บันทึกไว้อย่างเป็นระบบในโฟลเดอร์ `docs/`:

- **[คู่มือการเลือกอ่านเอกสาร (`docs/README.md`)](docs/README.md)**: แผนที่ On-Demand Context สำหรับเลือกอ่านเอกสารเฉพาะที่ตรงกับงาน
- **[คลังความรู้ระบบ (`docs/kms/README.md`)](docs/kms/README.md)**: รวบรวมสถาปัตยกรรมระบบ 20 บทความ (`00`–`19`), การตัดสินใจทางเทคนิค (ADR), และประวัติบั๊ก
- **[ทักษะและมาตรฐาน UI/UX (`docs/skills/ui-scale-polish/SKILL.md`)](docs/skills/ui-scale-polish/SKILL.md)**: มาตรฐานการออกแบบสำหรับผู้ใช้คนไทยอายุ 40+, สี Palette, และแบบแผน UI
- **[มาตรฐานการจัดการฐานข้อมูล (`docs/skills/audit-mongomodel-sync/SKILL.md`)](docs/skills/audit-mongomodel-sync/SKILL.md)**: กฎการเชื่อมประสานระหว่าง MongoModel และ PostgreSQL

---

## 🛠️ คำสั่งที่ใช้บ่อยในการพัฒนา (Developer Commands)

```bash
# ติดตั้ง Git Hooks ประจำเครื่อง (ต้องรันหลังจาก clone หรือแก้ไข .githooks/)
npm run hooks:install

# รันโหมดพัฒนา Frontend (Next.js)
npm run dev:frontend

# ตรวจสอบความถูกต้องของโค้ดแบบเร็ว (Code map + Frontend lint/typecheck/vitest)
npm run verify

# ตรวจสอบความถูกต้องแบบเต็มระบบ (รวม Backend integration suites)
npm run verify:all

# อัปเดตแผนที่ระบุบรรทัดของไฟล์ขนาดใหญ่ (CODE-MAP)
npm run codemap
```
