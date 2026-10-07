# HANDOFF 2026-10-08 — GL Champ parity (Claude → Gemini)

ส่งต่องานระบบบัญชีแยกประเภท (GL) ที่ทำให้เหมือน Champ (`D:\project-champ`, จอจริงใน `D:\project-champ\bcspec` ชนะ `.rc`) — กฎทั้งหมดอยู่ที่ `AGENTS.md` ของ repo อ่านก่อนเริ่ม

## 1. สถานะตอนส่งต่อ

- branch `dev` HEAD `730ebe22` — **นำหน้า `origin/dev` 61 commit ยังไม่ push** (push เมื่อลุงจืดสั่งเท่านั้น)
- **ยังไม่ deploy** — production ยังเป็น release ก่อนรอบนี้
- working tree สะอาด ยกเว้น `mydocs/`, `mydocs.zip` (ของลุงจืด ห้ามแตะ/ห้าม stage) และ `frontend/next-env.d.ts` (ห้าม stage)
- คำสั่งที่ยังมีผล (goal ของลุงจืด): **"ทำไปเรื่อยๆก่อน ยังไม่ต้องทดสอบ ไม่ต้องสนข้อมูลเก่า สนแต่ spec champ"**
  - ไม่ทดสอบบนจอ/browser แต่ **ยังต้องรัน compile + unit + integration test ก่อน commit ทุกครั้ง**
  - body ของ commit เขียนว่า "ยังไม่ได้ทดสอบบนจอจริงตามคำสั่งลุงจืด"
- ตารางเทียบ Champ ทีละเมนู (รหัส 5497–5541, 5881) = `docs/kms/champ/09-gl.md` — แถวที่ทำแล้วมีหมายเหตุ "อัปเดต 2026-10-0x" + `file:line`

## 2. งานรอบนี้ (ใหม่สุดก่อน — รายละเอียดใน `git show <hash>`)

| commit | งาน |
|---|---|
| `730ebe22` | ปุ่ม "ใช้ผังบัญชีมาตรฐาน" (5502): `accounts`/`load-standard`, ผัง `backend/internal/generalledger/standard_chart_th.json` 198 บัญชี รหัส 5 หลัก 5 ระดับ ตาม DBD 2566 แบบ 2 (ทะเบียน `docs/kms/21` แถว `DBD-FS-2566-COA`); โหลดได้เฉพาะบริษัทที่ไม่มีแถวบัญชีเลย ไม่ล้างผังแบบ Champ |
| `9f863af6` | ชุดผ่านรายการนับ "จำนวนภาษี" (5512) |
| `e4f0d959` | ห้ามชื่อบัญชีซ้ำ, แผงรายการที่ใช้บัญชีนี้, คอลัมน์ที่มาในจอผ่านรายการ, เงื่อนไขพิมพ์ผังบัญชี/สมุด/งบประมาณ + แก้บั๊กตัวกันลบบัญชี |
| `8cf6529c` | รายงานแยกประเภท/รายวัน/งบทดลอง ตัวเลือกแบบ Champ 5534/5532/5506 |
| `32b0e3be` | ทะเบียนภาษีซื้อ/ขาย หมายเหตุ เรียง กรองขอคืน ชื่อรายงาน (5536/5537) |
| `a1adc255` | กรอกยอดภาษี → หาฐานภาษีที่ backend (`vat_calc.go` mode `from_tax`) |
| `7b32884c` | จอบันทึกรายวัน: จำนวนรายการ, ก่อนหน้า/ถัดไป, ขยายคำอธิบาย, กล่องชุดผ่านรายการ |
| `c212311b`, `7a38ea8e`, `6c312894` | เลขที่ซ้ำ → เสนอเลขใหม่, รายงานเลือกช่วงงวด, กรองหลายแผนก/โครงการ |
| `4fcca458` | ลบตัวตรวจ VAT 7% ฝั่ง browser (ห้ามมีอัตรา VAT ค่าเริ่มต้น) |
| `ce267e7a` | ตัวอย่างรายการปิดบัญชี 5504 + กลุ่มผังบัญชีแบบช่วงรหัส 5497 |

ก่อนหน้านั้น (`a274cc95`…`d4b5586b`) ดู `git log --oneline origin/dev..dev`

## 3. วิธีตรวจ (ต้องผ่านก่อน commit)

- **Go** (ต้องมี container `bc-pg-lead` = `postgres:18-alpine` + `POSTGRES_HOST_AUTH_METHOD=trust` ยังรันอยู่; ถ้าหายให้สร้างใหม่ชื่อเดิมด้วย image/env นี้):
  ```bash
  cd backend && MSYS_NO_PATHCONV=1 docker run --rm --network container:bc-pg-lead -e BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable' -v bc-go-mod:/go/pkg/mod -v "D:/bccode/backend:/src" -w /src golang:1.26 sh -c 'go vet ./internal/generalledger/... ./internal/goapi/handlers/... && go test -count=1 -tags=integration ./internal/generalledger/... ./internal/goapi/handlers/... ./internal/goapi/language/... ./internal/fixedasset/...'
  ```
  ห้ามตั้ง `BC_TAX_TEST_POSTGRES_DSN`; package ภาษา = `./internal/goapi/language` (ไม่ใช่ `./internal/language`)
- **gofmt** (ไฟล์มี CRLF): `tr -d '\r' < f.go > /tmp/x.go && gofmt -l /tmp/x.go`
- **Frontend** (ใน `frontend/`): `npx tsc --noEmit -p .` แล้ว `npx vitest run src/app/gl src/app/api src/lib src/app/menu src/app/tax` (ล่าสุด 1137/1137)
- **eslint** เทียบกับ HEAD: `git show HEAD:frontend/<p> | npx eslint --stdin --stdin-filename <p>` จำนวนปัญหาต้องไม่เพิ่ม
- **ก่อน deploy**: `npm run verify:all` (รอบก่อนผ่านทั้งหมด frontend 1271/1271)

## 4. กับดักที่เจอแล้ว

- `backend/assets/language/languages.tsv`: ข้อความใหม่ = 1 แถว 13 คอลัมน์ (`key th en` + อีก 10 ช่องใส่อังกฤษ) ต่อท้ายแบบ LF; error ของ backend ต้องมีแถว `gl_err_<code>` (`TestGLUserErrorCodesHaveEnglishRows`), key ฝั่งจอถูกตรวจโดย `gl-language-keys.test.ts`
- แก้ไฟล์ ≥950 บรรทัด → `pwsh -NoProfile -File tools/gen-code-map.ps1` แล้ว `git add docs/reference/CODE-MAP.md` (pre-commit ตรวจ)
- `git add <path>` เฉพาะไฟล์งาน; ตรวจ secret ใน `git diff --cached` ด้วยเงื่อนไข `&&` (fixture `"Bearer test-token"` ใน test เป็นค่าปลอมที่รู้แล้ว); ห้าม `--no-verify` / `--amend`
- รายงานของ GL: filter หลายรหัสใช้ `parseReportCodeList` + `cardinality($n::text[])=0 OR col=ANY($n::text[])`; GET ใหม่ต้องเพิ่มใน allowlist ของ BFF `frontend/src/app/api/gl/[...glPath]/route.ts` (POST command มี allowlist action ด้วย)
- สิทธิ์: `backend/internal/generalledger/httpapi/http.go` (`resourceScreens`, `reportScreens`, การแมป action → permission ~บรรทัด 652)
- ไฟล์ test บางไฟล์เป็น CRLF — script แก้ไฟล์ต้องรองรับ CRLF; ห้ามใช้ Edit tool กับ `frontend/src/app/globals.css` (EOL ผสม)
- ห้ามแตกเงื่อนไขตามรหัสบัญชี, ห้ามอัตรา VAT ค่าเริ่มต้น, ห้าม float กับเงิน, ห้าม N+1, ห้าม engine คำนวณฝั่ง browser

## 5. งานค้าง

**ทำต่อได้เลย (ไม่ต้องรอลุงจืด):**
1. ลบไฟล์ผังเก่า `docs/examples/thai-chart-of-accounts-standard.json` (รหัส 4 หลัก 3 ระดับ ขัดกฎ AGENTS.md) แล้วให้ `frontend/src/lib/chart-of-accounts-tree.test.ts:189` อ่าน `backend/internal/generalledger/standard_chart_th.json` แทน
2. (ทางเลือก) เพิ่มบัญชีที่ตัดออกจากผังมาตรฐาน: อสังหาริมทรัพย์เพื่อการลงทุน, เงินลงทุนระยะยาว, ดอกเบี้ยเช่าซื้อรอตัดบัญชี (ยอดปกติเดบิตใต้หนี้สิน) — ต้องผ่าน `validateStandardChart` + ชื่อไทยห้ามซ้ำ
3. Champ ยังไม่ทำ/ข้าม: ยอดรวมต่อหน้า (SUMPAPER) ในรายงาน, ตรวจงวด 0/งวดขาดช่วง (`09-gl.md` แถว 6)

**ต้องให้ลุงจืดตัดสินก่อน (ห้ามเดา):**
- ตัวกรองสถานะในรายงาน (ผ่านแล้ว/ยังไม่ผ่าน/ทั้งหมด) — ต้องอ่านฉบับร่างด้วย
- เมนู 5515 ประมวลผลข้อมูลบัญชีใหม่ ซ้ำกับ 5514 → "รอพัฒนา" หรือเปลี่ยนความหมาย
- เปลี่ยนรหัสบัญชีแบบไล่แก้ทุกตาราง (กระทบใบที่ผ่านรายการแล้ว)
- มิติเพิ่ม 5 ตัว, ธง "ทำรายการที่ GL ได้" (IsGLAccess), ช่องกลุ่มภาษี/อัตรา VAT ของบริษัท
- ป้ายเมนู 5881; เลขที่เอกสารผูกกับสมุดหรือกลุ่มเอกสาร; ปุ่มลัด Alt+D ชนกัน; ค่าเริ่มต้นยอดรวมผู้รับเงิน 5541; งบทดลองแบบ 3 (5508)
- ทะเบียนภาษีซื้อ: สถานะขอคืน 2/3/4 ตอนนี้นับเป็น "ไม่ขอคืน"; ค่าเริ่มต้นการเรียง (Champ = วันที่เอกสาร, BC = วันที่ใบกำกับ)
- ชื่อผู้ผ่านรายการในกล่องชุดผ่านรายการ ตอนนี้แสดง UID (`scope.Actor`) — GL ไม่มีแหล่งชื่อผู้ใช้
- ผังบัญชีมาตรฐานมีบัญชี ภ.ง.ด.1/ประกันสังคมค้างจ่าย (เป็นบัญชีเท่านั้น ไม่ใช่ระบบเงินเดือน) และค่าเสื่อมเครื่องจักรอยู่ใต้ค่าใช้จ่ายบริหาร

**Deploy (เมื่อลุงจืดอนุญาตให้ทดสอบบนจอ):** ทดสอบผ่านปุ่ม "ทดลองใช้ระบบ (Demo)" → `rungrueng` → บริษัท `01` → สาขา `00000` (ห้ามกรอกรหัสผ่านจริง) ตรวจ payload ใน network + console + screenshot แล้ว `py tools/fast-deploy.py --all --tag rYYYYMMDD-N`; ผังบัญชีมาตรฐานต้องทดสอบกับบริษัทที่ยังไม่มีผังบัญชี
