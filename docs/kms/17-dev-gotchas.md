# กับดักตอนพัฒนา (dev gotchas) ที่เคยเสียเวลาไปแล้ว — อย่าเจอซ้ำ
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) · แถวที่ระบุวันที่ 2026-09-09 ตรวจซ้ำที่ commit 098107e1 — รวมจาก memory ของ Claude + handoff; แต่ละข้อมีวันที่ที่ยืนยันจริง

## Frontend

| กับดัก | อาการ | วิธีที่ถูก | ยืนยัน |
|---|---|---|---|
| `frontend/src/app/globals.css` มี EOL ผสม (CRLF/CR/LF) | Edit/Write tool normalize ทั้งไฟล์ → diff 1,400+ บรรทัดจากการแก้ 1 บรรทัด | แก้ด้วย Node script แบบ byte-preserving (`readFileSync` → `String.replace` ตรง ๆ → `writeFileSync`); ตรวจ `git diff --stat` ต้องเล็ก | 2026-08 |
| ปุ่มไม่รับ `text-xs`/`text-sm` | มี reset แบบ unlayered `button{font-size:inherit}` ชนะ Tailwind v4 utilities เสมอ | ใช้ `text-xs!` (important) หรือ inline style | 2026-06-29 |
| dark mode ทดสอบผิดวิธี | flip `data-theme` ด้วย JS แล้วดูเหมือน bug | palette ถูกเขียน inline บน `<html>` โดย `applyVisualTheme()` (`src/lib/theme-data.ts`) → ต้อง**กดปุ่มสลับธีมจริง** | 2026-09-02 |
| แก้ CSS login แล้วไม่เปลี่ยน | block กลางไฟล์ (design-pass 2026-08-28) hard-code สีขาว ชนะ tail override | selector ท้ายไฟล์ต้อง `.login-shell .login-card` ขึ้นไป และใช้ตัวแปร `--login-*` | 2026-09-02 |
| `localhost:3000` โชว์ UI เก่า | มี `next start` (production build) ค้างอยู่ | `netstat -ano \| grep :3000` → kill → `npm run dev` | 2026-09-02 |
| `npm`/`npx` ใน Git Bash บน Windows พัง ("cannot find the path specified") | postinstall/npx ล้ม | `npm install --ignore-scripts`; รัน `node node_modules/next/dist/bin/next dev` ตรง ๆ ถ้า `npm run dev` ไม่ขึ้น | 2026-06-29 |
| `next.config.ts` ต้องไม่มี `output: "standalone"` | `next start` พังใน Docker on-prem | build ด้วย `frontend/Dockerfile` (multi-stage, จบที่ `CMD ["npm","run","start"]`) และต้องส่ง build-arg `BCAI_LOCAL_BACKEND_URL` เสมอ ไม่งั้น `next build` ล้มที่ rewrites (frontend/next.config.ts:19-21) — **ไม่มีไฟล์ `frontend/Dockerfile.onprem` ในรีโป** (`git ls-files frontend/Dockerfile*` มีแค่ `frontend/Dockerfile`; ดู `10-infra-deploy.md:90`) | 2026-09-09 |
| viewport | ออกแบบไว้สำหรับ **iPad (≥768px) ขึ้นไป** — มือถือแค่ไม่พัง | ใช้ grid หลายคอลัมน์บนจอกว้าง ไม่ stack ทุก field แนวตั้ง | กฎ 2026-06-28 |
| Popover โดนตัด | `overflow: hidden` บน panel ที่มี popover ลูก | clip ที่ shell ชั้นนอกสุดเท่านั้น (กฎ premium ข้อ 7 ใน AGENTS.md) | 2026-09-02 |

## Backend / Go

| กับดัก | รายละเอียด | ยืนยัน |
|---|---|---|
| build/test บน Windows native ไม่ได้ | `CGO_ENABLED=0` + ไม่มี gcc → confluent-kafka undefined; ใช้ container `golang:1.26` + `apt-get install librdkafka-dev` + `MSYS_NO_PATHCONV=1` (Git Bash แปลง `/src` เป็น path Windows) | 2026-09-03 |
| `.ignore` ที่ root มี `**/build/` | ripgrep ข้าม `backend/internal/goapi/process/build/` ทั้งโฟลเดอร์ (rebuild/DDL ทั้งหมดอยู่ที่นั่น) → ใช้ `rg --no-ignore` | 2026-09-05 |
| kafka-go กับ librdkafka แชร์ consumer group ไม่ได้ | JoinGroup member metadata คนละรูปแบบ → rebalance วน; barcode reader ใหม่จึงใช้ group `<id>-projection` | 2026-09-05 |
| `docker compose run` (v5.3) | restart one-shot dependency ที่ exit แล้ว (เช่น `mongo-init`) → `rs.initiate` ซ้ำพัง → ใช้ `run --rm --no-deps` และ init ต้อง idempotent | 2026-09-05 |
| Mongo ต้องเป็น replica set | standalone ให้ `Transaction numbers are only allowed on a replica set member` — outbox/create-holding ใช้ transaction | 2026-06-21 |
| `mongo.CommandError` เทียบด้วย `errors.Is` ไม่ได้ | struct ไม่ comparable → ตรวจ label `HasErrorLabel("TransientTransactionError")` แทน | 2026-09-05 |
| gofmt import order | แทรก import เองแล้วลืม sort → CI gofmt แดง; รัน `gofmt -l .` ก่อน commit | 2026-09-05 |
| ไฟล์ log ถูก track ใน git | `backend/internal/goapi/process/report-stock/logs/goapi_debug.log` เปลี่ยนทุกครั้งที่รัน test → `git checkout --` ก่อน commit (ควรเลิก track) | 2026-09-05 |
| curl.exe บน Windows กับข้อความไทย | `curl -d "…ไทย…"` inline กลายเป็น `?????` ใน DB → เขียน JSON ลงไฟล์ UTF-8ก่อนแล้ว `--data-binary @file.json` | 2026-07-02 |
| Dev Login 401/404/503 | ดู `16-environments-and-servers.md` §DEV (secret drift ระหว่าง env กับ container) | 2026-09-02 |
| ชื่อ collection แบบ camelCase (`productGroups`/`productCategories`) = ชื่อ **ก่อน migration** ไม่ใช่ของจริง | ของจริงเป็นตัวพิมพ์เล็กล้วน (`productgroups`/`productcategories`) ตามตารางแปลงชื่อที่ `backend/cmd/storage_name_migration/MIGRATION_HISTORY.md:83-84`; live `appdb` ไม่มี collection ตัวพิมพ์ใหญ่เหลือแล้ว และ `.mcp.json` ปัจจุบันมีแค่ server `mongodb`/`docker`/`mongomodel`/`playwright`(disabled)/`windows` ไม่มี AI write-tool ที่ใช้ชื่อเก่า — ถ้าเจอสคริปต์/เครื่องมือ**นอกรีโป**ที่ยังเขียนชื่อ camelCase ให้แก้ที่เครื่องมือนั้น อย่าสร้าง collection ใหม่ | 2026-09-09 |
| `backend/assets/language/languages.tsv` ต้องมี **13 คอลัมน์ต่อบรรทัดพอดี** (`key,th,en,cn,ja,km,ko,lo,my,vi,ms,id,fil` ตามบรรทัดหัวไฟล์) | tab ขาดแล้วไม่มีอะไรเตือน: แถวที่มีน้อยกว่า 2 คอลัมน์ถูกข้ามเงียบ ๆ (`backend/internal/goapi/language/language.go:166-168`) → `Text()` คืน key ดิบให้ผู้ใช้เห็น (`:69-78`); แถวที่ขาดคอลัมน์กลางทำให้ค่าภาษา**เลื่อนช่อง** (`line_oa_qrcode` บรรทัด 2127 มี 12 คอลัมน์ ขาด `km` → ช่อง `km` เก็บข้อความเกาหลี, `ko` เก็บลาว, `lo` เก็บพม่า … `fil` ว่าง; key นี้ใช้จริงที่ `frontend/src/app/line-oa/line-oa-link-screen.tsx:114`); test ที่มี (`frontend/src/lib/menu-data.test.ts:293-304`) ตรวจครบทุกภาษาเฉพาะ key ที่ผูกกับเมนูเท่านั้น จึงไม่จับแถวเหล่านี้ — ตรวจทั้งไฟล์ก่อน commit ด้วย `awk -F'\t' 'NF!=13 && NF>0 {print NR": "NF}' backend/assets/language/languages.tsv` (2026-09-09 ยังค้าง 11 แถว: 893, 2025, 2127, 2273, 2550, 2551, 3130, 3199, 3841, 4389, 4408 โดย 4408 ใช้ข้อความไทยเป็น key และไม่มีแถวใดเป็น key ของเมนู) | 2026-09-09 |

## UAT ผ่าน browser (harness)

- ฟอร์มแก้ **บาร์โค้ด** ไม่ยิง PUT ถ้าใช้ `form_input` อย่างเดียว — ต้องพิมพ์คีย์จริงในช่องอย่างน้อย 1 ครั้ง (End → space → BackSpace) ก่อนกดบันทึก; ฟอร์มสินค้าไม่มีปัญหานี้
- เปิดแท็บสินค้า+บาร์โค้ดพร้อมกันจะมีปุ่ม "แก้ไข/ลบ" 2 ชุด — ref ใหม่กว่าเป็นของแท็บที่ active; เช็ค header ก่อนกด (dirty guard "ทิ้งการแก้ไข" ช่วยได้)
- หน้าจอ datacrud มีปุ่ม แก้ไข/ลบ รายแถว; ลบมี modal "ต้องการลบจริงหรือไม่" ต้องกด `ลบ` ซ้ำ; **ทุกจอใช้ confirm dialog ในแอปแล้ว ไม่มี `window.confirm` เหลือใน `frontend/src`** (`frontend/src/app/system-settings/warehouse-tree-view.tsx:855` เขียนกำกับไว้ว่า "never window.confirm" ตั้งแต่ commit b4a4898f 2026-07-03) → Stagehand ไม่ต้อง monkey-patch `window.confirm` แต่ต้องกดปุ่มยืนยันใน modal จริง
- Stagehand v3 + DeepSeek: ต้องสร้าง `AISdkClient` จาก `@ai-sdk/deepseek` ที่ bundle มา แล้วส่งเป็น `llmClient` (ห้าม `modelName:"openai/deepseek-chat"`); `act/observe/extract` อยู่บน instance ไม่ใช่ `sh.page`
- Kafka poison test: `printf '%s\n' '[{"holdingcode":"demo","barcode":"X"}]' | docker exec -i kafka kafka-console-producer --bootstrap-server localhost:9092 --topic when-product-barcode-bulk-created` แล้วดู `docker logs mainapi | grep ⛔`
- **ล้างข้อมูลทดสอบด้วย id/code ที่ระบุเป้าเท่านั้น** — เคยใช้ regex กว้างแล้วลบสาขาจริงไป 2 ตัว

## สคริปต์ / PowerShell

| กับดัก | อาการ | วิธีที่ถูก | ยืนยัน |
|---|---|---|---|
| `Get-Date -Format yyyy-MM-dd` บนเครื่องที่ตั้ง culture ไทย | ได้ปีพุทธศักราช — stamp ใน `CODE-MAP.md` ออกมาเป็น `2569-09-09` แทน `2026-09-09` | pin culture เสมอ: `[datetime]::Now.ToString('yyyy-MM-dd', [System.Globalization.CultureInfo]::InvariantCulture)` | เจอจริง 2026-09-09 ตอนทำ `-Check` mode ของ `tools/gen-code-map.ps1:139-140,154` |
| สคริปต์ `.ps1` ที่มีข้อความไทยต้องมี BOM | Windows PowerShell 5.1 อ่านไฟล์เป็น ANSI → ข้อความไทยเพี้ยน | เขียนไฟล์ด้วย `utf-8-sig` (ห้ามใช้ Write tool ที่ตัด BOM ทิ้ง) | `tools/gen-code-map.ps1` ขึ้นต้นด้วย BOM |
| hook ใน `.githooks/` ไม่ทำงานเองหลัง clone | git อ่านเฉพาะ `.git/hooks/` เท่านั้น ไม่อ่านโฟลเดอร์ที่ track ไว้ | `npm run hooks:install` ครั้งเดียวต่อ clone (คัดลอกไฟล์เข้า `.git/hooks/`) — ต้องรันซ้ำทุกครั้งที่ pull การแก้ `.githooks/` | `tools/install-hooks.mjs` |
| **ห้ามใช้ `git config core.hooksPath .githooks`** | hooksPath **แทนที่** `.git/hooks` ทั้งโฟลเดอร์ — hook ส่วนตัวที่มีอยู่เดิมหยุดทำงานเงียบ ๆ ไม่มี error | ใช้วิธีคัดลอก (`npm run hooks:install`) เท่านั้น; ถ้าเคยตั้งไปแล้ว `git config --unset core.hooksPath` | เจอจริง 2026-09-09: การตั้ง hooksPath ปิด `post-commit` ของลุงจืด (sync Obsidian vault) โดยไม่แจ้ง |
| hook ที่ commit จาก Windows ไม่มี exec bit | repo ตั้ง `core.fileMode=false` → git เก็บเป็น `100644` → บน Linux/macOS รันไม่ได้ (การ์ดเงียบ) | `git update-index --chmod=+x .githooks/<hook>` ก่อน commit; ตรวจด้วย `git ls-files -s .githooks` ต้องเห็น `100755` | `.githooks/pre-commit`, `.githooks/pre-merge-commit` |
| merge commit ไม่วิ่ง `pre-commit` | git เรียก `pre-merge-commit` แทน — สอง branch ที่ต่างก็ผ่าน พอ merge แล้วแผนที่ผิด | มีไฟล์ `.githooks/pre-merge-commit` ที่ `exec` ต่อไปยัง pre-commit | `.githooks/pre-merge-commit:5` (ทดสอบ merge จริงแล้ว 2026-09-09) |

## ผู้ช่วย AI / เครื่องมือ

- Kimi K3 / GLM (`py ~/.claude/tools/kimi-ask.py|glm-ask.py`, ต้องใช้ `py` launcher + `PYTHONIOENCODING=utf-8`) — เหมาะกับงานสั้น; ร่างยาว 400+ บรรทัดจะ timeout
- สร้างรูป = Codex CLI `gpt-image-2` เท่านั้น (ห้าม SVG) — รายละเอียดใน `~/.claude/refs/image-gen-codex.md` ของเครื่องลุงจืด
- Magnitude (browser agent) ประเมินแล้ว 2026-07-01 **ไม่ใช้** (ต้องการ vision LLM ที่ไม่มี) — อย่าประเมินซ้ำถ้าไม่มีข้อมูลใหม่
