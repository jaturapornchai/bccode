# กับดักตอนพัฒนา (dev gotchas) ที่เคยเสียเวลาไปแล้ว — อย่าเจอซ้ำ
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — รวมจาก memory ของ Claude + handoff; แต่ละข้อมีวันที่ที่ยืนยันจริง

## Frontend

| กับดัก | อาการ | วิธีที่ถูก | ยืนยัน |
|---|---|---|---|
| `frontend/src/app/globals.css` มี EOL ผสม (CRLF/CR/LF) | Edit/Write tool normalize ทั้งไฟล์ → diff 1,400+ บรรทัดจากการแก้ 1 บรรทัด | แก้ด้วย Node script แบบ byte-preserving (`readFileSync` → `String.replace` ตรง ๆ → `writeFileSync`); ตรวจ `git diff --stat` ต้องเล็ก | 2026-08 |
| ปุ่มไม่รับ `text-xs`/`text-sm` | มี reset แบบ unlayered `button{font-size:inherit}` ชนะ Tailwind v4 utilities เสมอ | ใช้ `text-xs!` (important) หรือ inline style | 2026-06-29 |
| dark mode ทดสอบผิดวิธี | flip `data-theme` ด้วย JS แล้วดูเหมือน bug | palette ถูกเขียน inline บน `<html>` โดย `applyVisualTheme()` (`src/lib/theme-data.ts`) → ต้อง**กดปุ่มสลับธีมจริง** | 2026-09-02 |
| แก้ CSS login แล้วไม่เปลี่ยน | block กลางไฟล์ (design-pass 2026-08-28) hard-code สีขาว ชนะ tail override | selector ท้ายไฟล์ต้อง `.login-shell .login-card` ขึ้นไป และใช้ตัวแปร `--login-*` | 2026-09-02 |
| `localhost:3000` โชว์ UI เก่า | มี `next start` (production build) ค้างอยู่ | `netstat -ano \| grep :3000` → kill → `npm run dev` | 2026-09-02 |
| `npm`/`npx` ใน Git Bash บน Windows พัง ("cannot find the path specified") | postinstall/npx ล้ม | `npm install --ignore-scripts`; รัน `node node_modules/next/dist/bin/next dev` ตรง ๆ ถ้า `npm run dev` ไม่ขึ้น | 2026-06-29 |
| `next.config.ts` ต้องไม่มี `output: "standalone"` | `next start` พังใน Docker on-prem | ใช้ `frontend/Dockerfile.onprem` (build && start) | 2026-06 |
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
| `.mcp.json` / MCP AI tools เขียน collection ชื่อ camelCase | `productGroups`/`productCategories` ≠ ของจริง `productgroups`/`productcategories` (ยังไม่มี doc ใน collection ผี = ยังไม่เคยถูกเรียก) — ถ้าจะใช้ MCP tools ต้องแก้ก่อน | 2026-06-21 |

## UAT ผ่าน browser (harness)

- ฟอร์มแก้ **บาร์โค้ด** ไม่ยิง PUT ถ้าใช้ `form_input` อย่างเดียว — ต้องพิมพ์คีย์จริงในช่องอย่างน้อย 1 ครั้ง (End → space → BackSpace) ก่อนกดบันทึก; ฟอร์มสินค้าไม่มีปัญหานี้
- เปิดแท็บสินค้า+บาร์โค้ดพร้อมกันจะมีปุ่ม "แก้ไข/ลบ" 2 ชุด — ref ใหม่กว่าเป็นของแท็บที่ active; เช็ค header ก่อนกด (dirty guard "ทิ้งการแก้ไข" ช่วยได้)
- หน้าจอ datacrud มีปุ่ม แก้ไข/ลบ รายแถว; ลบมี modal "ต้องการลบจริงหรือไม่" ต้องกด `ลบ` ซ้ำ; บางจอ (`warehouse-tree-view.tsx`) ใช้ `window.confirm` → Stagehand ต้อง monkey-patch `window.confirm = () => true`
- Stagehand v3 + DeepSeek: ต้องสร้าง `AISdkClient` จาก `@ai-sdk/deepseek` ที่ bundle มา แล้วส่งเป็น `llmClient` (ห้าม `modelName:"openai/deepseek-chat"`); `act/observe/extract` อยู่บน instance ไม่ใช่ `sh.page`
- Kafka poison test: `printf '%s\n' '[{"holdingcode":"demo","barcode":"X"}]' | docker exec -i kafka kafka-console-producer --bootstrap-server localhost:9092 --topic when-product-barcode-bulk-created` แล้วดู `docker logs mainapi | grep ⛔`
- **ล้างข้อมูลทดสอบด้วย id/code ที่ระบุเป้าเท่านั้น** — เคยใช้ regex กว้างแล้วลบสาขาจริงไป 2 ตัว

## ผู้ช่วย AI / เครื่องมือ

- Kimi K3 / GLM (`py ~/.claude/tools/kimi-ask.py|glm-ask.py`, ต้องใช้ `py` launcher + `PYTHONIOENCODING=utf-8`) — เหมาะกับงานสั้น; ร่างยาว 400+ บรรทัดจะ timeout
- สร้างรูป = Codex CLI `gpt-image-2` เท่านั้น (ห้าม SVG) — รายละเอียดใน `~/.claude/refs/image-gen-codex.md` ของเครื่องลุงจืด
- Magnitude (browser agent) ประเมินแล้ว 2026-07-01 **ไม่ใช้** (ต้องการ vision LLM ที่ไม่มี) — อย่าประเมินซ้ำถ้าไม่มีข้อมูลใหม่
