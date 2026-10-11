# HANDOFF 2026-10-11 — งานต่อจาก Claude (Champ parity รอบ facade / เช็ค / เตรียมจ่าย / รายงานกลาง / มิติ) → Gemini

> ผู้เขียน: Claude (Opus 5.5) · ผู้รับ: Gemini (CLI/Antigravity) · คุยกับลุงจืดเป็นภาษาไทยเสมอ เรียก "ลุงจืด"
> ต่อจาก `docs/handoff/HANDOFF-2026-10-11-CHAMP-FULL-PARITY-CRUD.md` (ลำดับงาน §0: facade → ทะเบียนเช็ค/สมุดเงินฝาก → เตรียมจ่ายเจ้าหนี้ → มิติ → report registry → ข้อมูลหลักสินค้า)

## 0. อ่านก่อนเริ่ม (บังคับ)
1. กฎ project = `AGENTS.md` (Gemini CLI โหลดผ่าน `.gemini/settings.json` → `context.fileName`) — ฉบับเต็ม `docs/kms/rules/agents-full.md` เปิดเฉพาะหมวดที่ใช้
2. skill อยู่ที่ `.agents/skills/<name>/SKILL.md` — งาน UI อ่าน `ui-scale-polish`, งานภาษี/บัญชีอ่าน `thai-accounting-tax` + `docs/kms/21-thai-tax-form-references.md`
3. `mydocs/` = ข้อกำหนดของลุงจืด **อ่านอย่างเดียว ห้ามสร้าง/แก้/ลบ** (รวม `mydocs.zip`)
4. งาน processing ทุกชนิดต้องไล่โค้ด Champ จริงก่อน (`D:\project-champ`) แล้วอ้าง `file:line` ใน design/ADR/commit body — ไม่พบ = ถามลุงจืด ห้ามแต่ง workflow
5. ต้นทุนสินค้า = ถัวเฉลี่ยอย่างเดียว (ADR `docs/kms/decisions/2026-10-11-stock-cost-average-only.md`) ห้าม FIFO/LIFO/lot/standard ทุกรูปแบบ
6. คำตอบที่ลุงจืดตัดสินแล้ว: D1 / D8 / D11 / Q2 = **"ทำตาม Champ"** (`docs/kms/decisions/2026-10-11-champ-parity-handoff-decisions.md`)

## 1. สถานะ repo ณ ส่งมอบ
| รายการ | ค่า |
|---|---|
| branch | `dev` ใน `D:\bccode` (ไม่ใช่ worktree) |
| HEAD | commit `docs(handoff)` ของไฟล์นี้ (ถัดจาก `1ae4b95f`) — ดู `git log --oneline -5` |
| push | **ยังไม่ push** — dev นำหน้า `origin/dev` ราว 70 commit; push เมื่อลุงจืดสั่งเท่านั้น |
| prod | `https://account.bcaicloud.com` release `r20261011-1` (image สร้าง 2026-10-11 09:02 น.) — น่าจะเท่ากับ dev ที่ `1961f4c8`; **งานตั้งแต่ `8ff8c9cd` ทั้งหมดยังไม่ขึ้น prod** |
| ไฟล์ค้างที่ไม่ใช่ของงานนี้ | `tools/uat-live-*.mjs` (7 ไฟล์), `mydocs.zip`, `mydocs/` — **อย่าแตะ อย่า add** |
| งานมิติ (ยังไม่ merge) | branch `worktree-agent-ad74ce45d0d6363c3` tip `4d529581` ที่ worktree `D:\bccode\.claude\worktrees\agent-ad74ce45d0d6363c3` (ดู §5) |

## 2. งานที่เสร็จและ commit แล้วในรอบนี้ (`86e5392a..dev`)
| หมวด | commits | ผล |
|---|---|---|
| ERP Tier A + ล็อกงวด | `4d6f742a`, `fd371ff7`, `387aa183`, `b394e44d`, `70771aa3` | ล็อกงวดเอกสาร ERP, อนุมัติ/ยกเลิก, aging ตั้งค่าได้, รายงานสต็อก/ขายตามบริษัท + ทศนิยมตรง |
| facade → รอพัฒนา | `45070d4c` | จอที่มีแต่ UI ไม่มี backend แสดง "รอพัฒนา" (รายการใน `frontend/src/lib/audit-pending.test.ts`) |
| เช็ค/ธนาคาร | `5238571c`, `f521f67f`, `a44186a2`, `96be990a`, `aed1ca00`, `d3f9c5ad`, `a9ef3725` | รหัสธนาคาร 5 ชนิด, ผู้ติดต่อธนาคาร, สมุดเงินฝาก, ทะเบียนเช็ค + event append-only, เอกสารเปลี่ยนสถานะ 8 ชนิด + ลง GL, จอทะเบียนเช็ค/จอเอกสารสถานะ, แก้ deadlock/ช่องสมุดธนาคาร/เช็คขายลดไม่มีสมุด |
| รายงานกลาง | `a2436acd`, `7c82b8b1` | registry กลาง (R1–R3, R5, R6 ส่วน GL คลื่น 1) + เมนู/ไอคอน |
| สิทธิ์พิมพ์/ส่งออก Champ | `99fed30d`, `6050d9f2`, `587a3d56` | ค่าเริ่มต้นให้สิทธิ์พิมพ์/ส่งออก เว้นกลุ่มที่ติ๊กออก (`!report:print`/`!report:export`), ตรวจที่ backend ก่อนพิมพ์งบ/ทะเบียน/หมายเหตุ |
| guard ใบจากโมดูล | `2503b528`, `20fca388` | GL แก้/ลบ/กลับรายการ/ถอนตัดยอด ใบที่มาจาก ERP/เช็คไม่ได้ |
| เตรียมจ่ายเจ้าหนี้ | `30384f2e`, `1339b31c`, `5dbe6273` | ใบเตรียมจ่าย + การจองยอด + PV ตัดหนี้ (W0–W5), จ่ายตรงได้ไม่เกินยอดที่ไม่ถูกจอง |
| ต้นทุนเฉลี่ยอย่างเดียว | `0d911846`, `c9912d56` | กฎ + ตัด engine วิธีอื่นออก |
| แก้จาก verify รอบท้าย | `4867ec72`, `1ae4b95f` | lint ตัวแปร `module` ใน BFF รายงาน; race สร้างตาราง `cache_entries` (advisory lock + test) |
| อื่น ๆ | `fb942e7f`, `5cac0ed5`, `3231e8fb`, `64f9c91e`, `7084c2e3` | `npm run uat`, แก้ ภ.ง.ด.50 tab, จอสาขา, deadlock session cache |

เอกสารออกแบบของรอบนี้: `docs/kms/champ/design-2026-10-11/{facade-screens,cheque-bank,ap-payment-prepare,report-registry,gl-dimensions}.md`

## 3. ผล verify (หลักฐาน)
- `npm run verify:all` รอบแรกบน `a9ef3725` **ไม่ผ่าน 2 ขั้น** → แก้แล้วทั้งคู่:
  - frontend lint: `@next/next/no-assign-module-variable` ที่ `frontend/src/app/api/report/[...reportPath]/route.ts:25` → `4867ec72`
  - postgres: `TestLineLinkCodeOnPostgresCache` ตก `duplicate key ... pg_type_typname_nsp_index` (หลาย package สร้าง `cache_entries` พร้อมกันบน DB ใหม่) → `1ae4b95f` + `TestNewCacherConcurrentOnFreshSchema` (ตกก่อนแก้, ผ่าน 10/10 หลังแก้)
  - เทสต์ `general-ledger.test.ts` ตก 1 ตัวเพราะไฟล์ในเครื่องเป็น CRLF (ดู gotcha §9.1) → คืน EOL แล้วผ่าน 72/72
- รอบที่สองบน `1ae4b95f`: **ผ่านทั้งหมด** (codemap, frontend lint 0 error + tsc + vitest 154 ไฟล์ 1657 เทสต์, next build, backend unit, postgres integration) — 2026-10-11 13:01–13:04 น.
- ยังไม่ได้ทำ (ต้องทำก่อน deploy): `npm run uat`, screenshot รอบรวม (โดยเฉพาะจอเช็ค/จอเอกสารสถานะ/จอเตรียมจ่าย ยังไม่เคยเปิดใน browser), review `git diff 86e5392a..dev` รวม (เงิน/สิทธิ์ ใช้ reviewer ที่เก่งสุด)

## 4. งานถัดไปที่แนะนำ (เรียงลำดับ)
1. **ปิด batch นี้ให้ขึ้น prod** (ก่อนเริ่มงานใหม่)
   - `npm run uat` (Demo → `rungrueng` → บริษัท `01` → สาขา `00000`, psql หลังทุก operation)
   - screenshot จอใหม่ light/dark: จอทะเบียนเช็ครับ/จ่าย + จอเอกสารสถานะเช็ค (เปิดจากเมนูเช็ค; ผูกจอที่ `frontend/src/app/menu/main-menu-screen.tsx:2308-2311`), `/bookbankscreen`, จอเตรียมจ่ายเจ้าหนี้, จอรายงานกลาง GL
   - review diff รวม → รวบรวม fail ทั้งหมด → จัดกลุ่มตาม root cause → แก้รวดเดียว (fix ละ regression test) → `npm run verify:all` ซ้ำ
   - deploy: ตรวจ chain ก่อน `ssh root@159.223.43.229 "ls -t /opt/bcai-account/releases | head -3"` (ต้องเป็น `r20261011-1` ล่าสุด) แล้ว `PYTHONIOENCODING=utf-8 py tools/fast-deploy.py --all --tag r20261011-2` จาก `D:\bccode` (dev) เท่านั้น → `docker compose ps` + `docker logs --tail 50` + `npm run uat -- --target=prod` (อ่านอย่างเดียว RUNGRUENG/01; ห้ามแตะ TOFFT1)
2. **merge งานมิติ** (§5) — ใหญ่สุดและเสี่ยงสุด เพราะ base ตามหลัง dev 66 commit
3. เช็ค CB-8 → CB-9 → CB-10 → CB-12 (รอ D4) → CB-7 → CB-13 → CB-14 (§6.1)
4. รายงานกลาง R7–R10 (+ CB-13) และ R0 ของโมดูลอื่น (§6.3)
5. ข้อมูลหลักสินค้า (product master) — ยังไม่มีเอกสารออกแบบ ต้องไล่ Champ ก่อน

## 5. งานมิติวิเคราะห์ (GL dimensions) — WIP ยังไม่ verify
- ออกแบบ: `docs/kms/champ/design-2026-10-11/gl-dimensions.md` (6 มิติ: แผนก, โครงการ, การจัดสรร, งาน, หน่วยงาน, ฝ่าย + สวิตช์ระดับบริษัท 7 ตัว ค่าเริ่มต้นปิดตาม Champ)
- branch `worktree-agent-ad74ce45d0d6363c3`: `3ff73794` (backend + frontend lib บางส่วน, 52 ไฟล์) + `4d529581` (คอลัมน์มิติในจอรายวัน/รายงาน/งบประมาณ/พิมพ์, 8 ไฟล์) — agent ถูกหยุดกลางคัน **ยังไม่เคยรัน integration test, tsc, vitest**
- base `8b543c08` ตามหลัง dev 66 commit; `go build`/`go vet`/unit test ของ `internal/generalledger/...` ผ่านบน branch
- เอกสารออกแบบนิยามเฉพาะ W2, W4, W8 — การแบ่ง W1–W7 ด้านล่างเป็นการจัดกลุ่มของผู้ตรวจ ให้ยึดเนื้อหาใน §3 ของเอกสารออกแบบเป็นหลัก

| W | งาน | สถานะ | ที่เหลือ |
|---|---|---|---|
| W1 | registry 6 มิติ + schema + model (`dimensions.go`, `models.go`, `schema.sql`, `budget.sql`, `internal/database/{fresh_provision,fresh_provision_all,schema_full}.sql`) | บางส่วน | index `gl_lines_dimension2_idx` (§3.3) ยังไม่มี; ไฟล์ SQL 3 ไฟล์แก้มือ ต้องตรวจให้ตรง `schema.sql` |
| W2 | projection / reversal / ปิดงวด-ปิดปี / งบประมาณ / "เลือกทั้งหมด" + `dimension_provision.go` + CLI `backend/cmd/gldimensions` (dry-run เป็นค่าเริ่มต้น, `-apply` จึงเขียน) | บางส่วน | รัน integration; ตัดสินว่า holding ใหม่จะ provision อัตโนมัติหรือไม่ (ความเสี่ยง R5/R9 ในเอกสาร: prod rungrueng ที่ใช้แผนก/โครงการอยู่จะเห็นคอลัมน์หายถ้าไม่เปิดสวิตช์) |
| W3 | สวิตช์บริษัท `dimension_settings.go` (เก็บ `gl_records` id `default`) + REST | บางส่วน | `mutateDimensionSettings` ยังใช้ `fmt.Errorf` ข้อความไทย → ต้องเป็น key + `language.Text`; permission id `company-defaults` เป็นข้อสมมติ ต้องตรวจ metadata (Demo ต้องใช้ได้) |
| W4 | รายงานกรอง/จัดกลุ่ม 6 มิติ (`reports_dimensions.go`, `reports.go`, BFF `app/api/gl/[...glPath]/route.ts`) | บางส่วน | **รายงานกลาง `backend/internal/report/*` (จาก `a2436acd`) รู้จักแค่แผนก/โครงการ** (`defs_gl.go`, `source_gl.go`, `errors.go`) — ไม่ขยาย = 4 มิติใหม่หายเงียบในรายงานผ่าน registry; `DimensionExists` ที่ `ap_open_items.go` ต้องรู้จัก 4 ชนิดใหม่ |
| W5 | frontend: รายวัน รายงาน งบประมาณ พิมพ์ ตัวอย่างปิดงวด drilldown | บางส่วน | `gl-masters.tsx` ยังไม่รู้กฎ "ติ๊กแยกแต่ไม่มีรายการ = ห้ามใส่รหัส"; `gl-journals.tsx` ใช้ key `gl_dim_columns_dropped` ที่ยังไม่มีใน tsv |
| W6 | จอ `gl-dimensions.tsx` จัดการ 6 ชนิด (จอเดียว Tab/Action) + เมนูตาม `menuconfig.xml` + จอสวิตช์ใน "ค่าเริ่มต้นบริษัท" | **ยังไม่เริ่ม** | มีแค่ `frontend/src/lib/gl-dimension-register.ts`(+test) |
| W7 | `languages.tsv` + tests | บางส่วน | 93 แถวใหม่มี **14 คอลัมน์ (ต้อง 13)**, 36 แถวมีไทยในช่อง en, ขาด ~21 key ตาม §3.10 + `gl_dim_columns_dropped`; integration ปิดงวด/ปิดปี/reversal/in-use/งบประมาณ |
| W8 | กฎ ERP (`job_no` บนหัวเอกสาร ฯลฯ §1.7, §3.8) | ยังไม่เริ่ม | ทำหลัง W1–W7 |

**conflict ที่คาด** (`git merge-tree --write-tree --name-only dev worktree-agent-ad74ce45d0d6363c3`): `backend/assets/language/languages.tsv` (เก็บทั้งสองฝั่ง), `backend/internal/generalledger/httpapi/http.go` (union `resourceScreens` + คงตัวตรวจ purpose print/export ของ `6050d9f2`), `backend/internal/generalledger/models.go` (คง `fromModule` ของ guard โมดูล + `DimensionSettings`), `frontend/src/app/api/gl/[...glPath]/route.ts`, `frontend/src/app/gl/gl-reports.tsx` (dev เพิ่มรายงานกลาง/purpose)

**ขั้นตอน merge ที่แนะนำ**
1. ทำใน worktree เดิม: `git -C D:/bccode/.claude/worktrees/agent-ad74ce45d0d6363c3 -c core.autocrlf=false merge dev` → แก้ 5 conflict
2. frontend ใน worktree: `npm ci` ใน `frontend/` ของ worktree (**ห้ามทำ junction** — ดู §9.2) แล้ว `npx tsc --noEmit` + vitest ไฟล์ที่แตะ
3. แก้ของค้าง W1/W3/W4/W5/W7 ข้างบน แล้วทำ W6 (อ่าน `ui-scale-polish` ก่อน)
4. integration: `go test -tags=integration -count=1 ./internal/generalledger/... ./internal/report/... ./internal/transaction/...` (DSN §9.3)
5. เข้า dev: `git -C D:/bccode -c core.autocrlf=false merge --squash worktree-agent-ad74ce45d0d6363c3` → verify → commit แยกตาม root cause
6. `npm run verify:all` + `npm run uat` + screenshot Demo

**ตัดสินใจค้างของมิติ**: OD-2, OD-6, OD-8 (§8)

## 6. งานค้างตามเอกสารออกแบบ
### 6.1 เช็ค/ธนาคาร (`cheque-bank.md` §8 Work items)
เสร็จ: CB-1, CB-2, CB-3, CB-4, CB-5, CB-6, CB-11, CB-15, CB-16
| id | งาน | รอ |
|---|---|---|
| CB-7 | รายละเอียดเช็คใต้รายวัน GL (`details.cheques` → `gl_journal_cheques`) | — |
| CB-8 | จอการเคลื่อนไหวเช็ค + API movement (`cheque/movement.go`, `cheque-movement-screen.tsx`) | — |
| CB-9 | ใบเตรียมเลขเช็ค + ใช้/คืนเลข (ปุ่ม "เช็คเตรียมจ่าย" ตาม Champ `cqfrmcheqout.cpp:1817-1848`) | D13 |
| CB-10 | รายละเอียดธนาคาร (บรรทัดข้อความ + log) | — |
| CB-12 | ตำแหน่งพิมพ์ + พิมพ์ตรวจตำแหน่ง + พิมพ์เช็ค PDF | **D4** |
| CB-13 | รายงานเช็ค 11 ชื่อ + รายวัน 6 ชื่อ ผ่าน report registry (`menuconfig.xml:329-362`) | R7–R10 |
| CB-14 | ใบเสร็จ/ใบจ่าย ERP สร้างเช็คในทะเบียน + `cheque_source_links` ใน tx เดียว; แสดงชื่อเอกสารต้นทางแทน `kind_key` ดิบ | CB-9 |

ช่องว่างของจอเช็คที่ทำแล้ว: รหัสคู่ค้าพิมพ์เอง (ควรใช้ picker ลูกหนี้/เจ้าหนี้ที่มีอยู่ `cheque-register-screen.tsx`), เอกสารสถานะบน UI ทำได้แค่สร้าง/ดู/ยกเลิก (backend มี `UpdateDoc` แล้ว), บาทอย่างเดียว (D15), เช็คประเภท 424 ไม่มีสมุด + กฎ mapping `bank_book` → error `cheque_err_book_no_gl_account` (ตั้งใจ); ตรวจรายการ "ยังไม่ได้ตรวจ" §10 ของเอกสาร

### 6.2 เตรียมจ่ายเจ้าหนี้ (`ap-payment-prepare.md` §8)
เสร็จ: W0–W5. ค้าง: **B-1** WHT ของ PV (F6) — ต้องตรวจกฎหมาย/อัตราจาก rd.go.th ลง `docs/kms/21-thai-tax-form-references.md` ก่อนเขียนโค้ด · **B-2** ส่วนลด PV · **B-3** บันทึกจ่ายเงิน/เช็ค/โอน (F7, ผูก CB-9) · **W6** UAT CRUD + psql + smoke prod · คำถามค้าง D4, D6, D10, D12, D13

### 6.3 รายงานกลาง (`report-registry.md` §13)
เสร็จ: R1, R2, R3, R5, R6 (GL คลื่น 1), R0 เฉพาะ GL (`docs/kms/champ/report-parity-r0-gl.md`), สิทธิ์พิมพ์/ส่งออก
ค้าง: R0 ของ AR/AP/เช็ค/สต็อก/ภาษี · **R7** ภาษี #5–#9 · **R8** AR/AP ฐานเดียวจาก `buildAgingQuery` + ลบ endpoint เดิม · **R9** ธนาคาร-เงินสด 5164/5152 · **R10** เมนู planned "รอพัฒนา" · **R11** ลบ `Number()` ใน `frontend/src/app/report/erp-report-viewer.tsx` และ `frontend/src/lib/erp-reports.ts` (Rule 20) · **R12** tests §9 · **R13** xlsx (หลัง Q1) · `gl_statement` ยัง planned ใน `defs_gl.go` · กระทบยอดจำนวนรายงาน PDF 415 vs menuconfig 366

### 6.4 facade / อื่น ๆ
- จอ "รอพัฒนา" ที่เหลือ (รายการใน `frontend/src/lib/audit-pending.test.ts`) — เปิดจอได้เมื่อมี backend จริงเท่านั้น; ล็อกรายการตาม Champ `ICDlgLockCostDlg.cpp:132-270`; รับเงินมัดจำต้องมีช่องจำนวนเงิน + posting; คิว e-Tax = ความสามารถใหม่ ต้องเสนอลุงจืด; `help-standard-software` ต้องมีเลขทะเบียนจริงหรือถอด (D-5)
- ERP: ยกเลิกรายบรรทัด (Tier B) เลื่อนไว้ตาม ADR `2026-10-11-champ-parity-handoff-decisions.md` ข้อ 2; รายการ ERP ที่เหลือดู handoff เดิม §2ก
- ข้อมูลหลักสินค้า: ยังไม่ออกแบบ

## 7. หนี้เทคนิค/ช่องว่างที่ตรวจกับโค้ดแล้ว
| # | เรื่อง | หลักฐาน | ระดับ |
|---|---|---|---|
| G1 | GL reconcile ยังแทน WHT/VAT ของใบที่มาจากโมดูลได้ (guard บล็อกแค่ Settlements/Allocations) — ตรวจ Champ ว่า Source≠0 แก้ภาษีได้หรือไม่ แล้วปิดพร้อม test | `backend/internal/generalledger/subledger_reconcile.go:34-35,47` | กลาง |
| G2 | engine ต้นทุนสต็อกใช้ `float64` ~60 จุด ทั้งที่ DB เป็น `NUMERIC(18,4)` (ผิด Rule 20) — เปลี่ยนเป็น decimal = เปลี่ยน API contract (R1 บอก why ก่อน) | `backend/internal/goapi/inventory/models/models.go` (50), `service.go` (7), `costing/*.go` | สูง |
| G3 | agent ต้นทุนใน `c9912d56` ตัดสินเอง: ลบนโยบายสต็อกติดลบแบบ 4, ลบโหมด BOM standard, คง `fixedcost` — ขอลุงจืดยืนยัน | `backend/internal/goapi/process/stockengine/calc.go` | ถามเจ้าของ |
| G4 | `/organization/role-permission/me` คืนสิทธิ์ดิบ ไม่ผ่าน `report.GrantChampDefaults` → frontend อาจเห็นปุ่มพิมพ์ต่างจาก backend | `backend/internal/report/run.go:50`, ใช้แค่ `generalledger/httpapi/scope.go:97` | กลาง |
| G5 | กลุ่มสิทธิ์ที่เคยติ๊กออกผ่านเว็บก่อน `99fed30d` จะอ่านเป็น "ให้สิทธิ์" (dev DB ทิ้งได้ จึงไม่ migrate) | `99fed30d` | ต่ำ |
| G6 | ข้อความไทย hard-code (`fmt.Errorf("...ไทย")`) ใน `backend/internal/generalledger/*` หลายจุด ผิดกฎ key+`language.Text` | เช่น `subledger_reconcile.go:30` | กลาง |
| G7 | รายงาน ERP ฝั่ง frontend รวมเงินด้วย `Number()` | `frontend/src/lib/erp-reports.ts`, `erp-report-viewer.tsx` | สูง (R11) |

## 8. คำถามที่รอลุงจืดตัดสิน
| id | เรื่อง | แนะนำ | เอกสาร |
|---|---|---|---|
| OD-2 | สาขาเป็นมิติระดับบรรทัด (Champ เก็บ BranchCode ต่อบรรทัด) หรือไม่ | [2] คงระดับหัวเอกสาร + รายงาน 5529 "รอพัฒนา" | `gl-dimensions.md:452` |
| OD-6 | หัวคอลัมน์มิติใหม่ในจอรายวัน | ใช้ชื่อทะเบียน (การจัดสรร/งาน/หน่วยงาน/ฝ่าย) | `gl-dimensions.md:453` |
| OD-8 | ปุ่ม F5 ตัวอย่างของมิติ | "รอพัฒนา" ก่อน | `gl-dimensions.md:454` |
| D4 (เช็ค) | วิธีพิมพ์เช็คจ่าย | (a) ตั้งตำแหน่ง มม. ต่อสมุด + พิมพ์ตรวจตำแหน่ง | `cheque-bank.md` §7 |
| D13 (เช็ค) | การรันเลขเตรียมจ่าย | ตัวเลขล้วน รันทีละ 1 คงความกว้าง | `cheque-bank.md` §7 |
| D15 (เช็ค) | เช็คสกุลเงินต่างประเทศ | ดูเอกสาร (กระทบ Rule 20/FX) | `cheque-bank.md` §7 |
| D4/D6/D10/D12/D13 (AP) | ส่วนลดยกไป PV, ปุ่ม Info/F1/Recur, ดึงจากใบรับวางบิล, ประเภท "เจ้าหนี้อื่นๆ", เปิดใช้ก่อน PV มี WHT | ดูคอลัมน์ "แนะนำ" ในเอกสาร | `ap-payment-prepare.md` §7 |
| Q1 / Q3 (รายงาน) | ทำ xlsx ไหม / รายงานที่ BC มีแต่ Champ ไม่มี | ดูเอกสาร | `report-registry.md` §11 |
| D-5 | เมนูซอฟต์แวร์มาตรฐานต้องมีเลขทะเบียนจริง | ให้เลขหรือถอดเมนู | `facade-screens.md` |
| G3 | ยืนยันการตัดสินใจของ agent ต้นทุน | — | §7 |

## 9. Gotchas (เจอจริงในรอบนี้)
1. **CRLF หลัง merge**: เครื่องนี้ `core.autocrlf=true` แต่ไฟล์ใน git เป็น LF — `git merge --squash` เขียนไฟล์ที่ merge เป็น CRLF → เทสต์ที่อ่าน source ตรง ๆ (เช่น `frontend/src/lib/general-ledger.test.ts`) ตก. ป้องกัน: ใส่ `-c core.autocrlf=false` ทุกครั้งที่ merge/checkout. ตรวจ/แก้หลังเกิด (คืนเป็น blob เดิมเฉพาะไฟล์ที่เนื้อหาตรง HEAD):
   ```bash
   git diff --name-only 86e5392a HEAD | while read p; do [ -f "$p" ] && [ "$(tr -cd '\r' < "$p" | wc -c)" -gt 0 ] && [ "$(git show HEAD:"$p" | tr -cd '\r' | wc -c)" -eq 0 ] && [ "$(git hash-object "$p")" = "$(git rev-parse HEAD:"$p")" ] && git show HEAD:"$p" > "$p"; done
   ```
2. **worktree + node_modules**: ห้ามทำ junction `frontend/node_modules` ไปที่ main checkout — `git worktree remove`/`rm -rf` จะตามเข้าไปลบแพ็กเกจของ main (เกิดจริง ต้อง `npm ci` ใหม่) และ Turbopack ใช้ junction ไม่ได้. ใน worktree ให้ `npm ci` แยก; ก่อนลบ worktree ตรวจ `[ -L <wt>/frontend/node_modules ]` และใช้ `cmd //c "rmdir <junction>"`. ถ้า `npm ci` ใน main ติด EPERM ให้หยุด dev server :3000 ก่อน (ล็อกไฟล์ `.node` ของ lightningcss)
3. **DSN integration test** (รันจาก `backend/`, ห้ามพิมพ์รหัสผ่านลงแชท/log):
   ```bash
   PW=$(grep -E "^\s+POSTGRES_PASSWORD:" docker-compose.local.yml | head -1 | sed -E 's/^\s+POSTGRES_PASSWORD:\s*//; s/^"//; s/"$//'); export BC_GL_TEST_POSTGRES_DSN="postgres://postgres:${PW}@127.0.0.1:5432/appdb?sslmode=disable"; export GL_AUTH_TEST_DSN="$BC_GL_TEST_POSTGRES_DSN"; unset PW # pragma: allowlist secret
   ```
   `npm run verify:all` ขั้น postgres ใช้ PostgreSQL 18 ชั่วคราวแยกของมันเอง (ไม่ต้องตั้ง DSN)
4. **CODE-MAP**: ไฟล์ ≥950 บรรทัดหรือที่อยู่ใน map ต้องรัน `pwsh -NoProfile -File tools/gen-code-map.ps1` แล้ว add `docs/reference/CODE-MAP.md` ใน commit เดียวกัน (pre-commit ตรวจ); ข้ามครั้งเดียวได้ด้วย `SKIP_CODE_MAP_CHECK=1 git commit ...` เมื่อ commit ถัดไปจะรวม map
5. **languages.tsv**: 13 คอลัมน์ `key th en cn ja km ko lo my vi ms id fil`; th ต้นฉบับ, en แปลจริง, อื่น ๆ ใส่อังกฤษ ห้ามว่าง; merge แบบเก็บทั้งสองฝั่งแล้วตรวจ `awk -F'\t' 'NF!=13' backend/assets/language/languages.tsv`
6. **แก้ไฟล์ด้วย Node**: ใช้ function replacer `s.replace(a, () => b)` (กัน `$1`/`$&` ใน string แทนที่); `frontend/src/app/globals.css` เป็น EOL ผสม ต้องแก้แบบ byte-preserving
7. **commit**: `git add <path>` เฉพาะไฟล์ของงาน (ห้าม `-A`/`.`), สแกน `git diff --cached` หา `sk-|Bearer |password=|api_key=|AIza|BEGIN .*PRIVATE KEY` เจอ = หยุดถามลุงจืด; บรรทัดแรกอังกฤษ conventional, body ไทย; ห้าม `--no-verify`, ห้าม `--amend`, ห้าม `git pull`/`reset --hard origin`
8. **ทดสอบหน้าจอ**: ใช้ปุ่ม "ทดลองใช้ระบบ(Demo)" เท่านั้น ห้ามพิมพ์รหัสผ่านจริง; dev server หลักอยู่ที่ :3000
9. **ห้ามใส่ชื่อแบรนด์คู่แข่ง** ในโค้ด/docs/commit (pre-commit จับ)

## 10. เอกสารอ้างอิง
- handoff ก่อนหน้า: `docs/handoff/HANDOFF-2026-10-11-CHAMP-FULL-PARITY-CRUD.md`
- ADR รอบนี้: `docs/kms/decisions/2026-10-11-champ-parity-handoff-decisions.md`, `2026-10-11-stock-cost-average-only.md`, `2026-10-09-champ-parity-purpose-additive-extras.md`, `2026-10-11-champ-processing-workflow-template.md`
- บั๊กที่บันทึกรอบนี้: `docs/kms/bugs/2026-10-11-*.md`
- ทดสอบ/verify: `docs/kms/11-testing-quality.md` §5, `tools/verify.sh`, `tools/uat/`
