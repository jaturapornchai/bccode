# docs/kms — ฐานความรู้ BC Ai Account (Knowledge Management)

ที่เก็บความรู้ที่ "ต้องไม่ลืม" ของโปรเจ็กต์ — ใช้ร่วมกันโดยลุงจืดและ AI ทุกตัว (Claude / Codex / Gemini) เพื่อให้เข้าใจตรงกัน
ตั้งโดยลุงจืด 2026-09-07 · กฎบังคับอยู่ใน `AGENTS.md` (ไฟล์นั้นชนะเสมอ) · skill ส่วนตัวอยู่ที่ `docs/skills/` · handoff ล่าสุดอยู่ที่ `docs/handoff/`

## ลำดับการอ่าน (On-Demand: โหลดเฉพาะเมื่อจำเป็น ไม่เปลือง Context)

> [!IMPORTANT]
> **ห้ามอ่านเอกสารทั้งหมดพร้อมกันเด็ดขาด**: อ่านเฉพาะไฟล์ที่ตรงกับงานเพื่อประหยัด Context Window ของ AI
> 1. **งานเล็ก / แก้บั๊ก 1 บรรทัด / คำถามทั่วไป** → ดูโค้ดจริงโดยตรง (`code = truth`) ไม่ต้องเปิด docs
> 2. **งาน UX/UI** → อ่านเฉพาะ `docs/skills/ui-scale-polish/SKILL.md`
> 3. **งาน Schema / MongoModel** → อ่านเฉพาะ `docs/skills/audit-mongomodel-sync/SKILL.md`
> 4. **งานสถาปัตยกรรม / โดเมนเฉพาะเรื่อง** → ดูตารางสรุป 1 บรรทัดด้านล่าง แล้วเลือกเปิดเฉพาะ **1 บทความที่เกี่ยวข้อง**
> 5. **Handoff (`docs/handoff/`)** → อ่านเฉพาะเมื่อลุงจืดถามสถานะงานค้าง/ความเสี่ยง หรือเริ่มงานสถาปัตยกรรมใหญ่ข้ามระบบ

## กติกา

- **โค้ด = ความจริง, docs ตามโค้ด** — ทุกข้อเท็จจริงอ้าง `path:line` และระบุวันที่ตรวจ (บรรทัด 2 ของทุกบทความ); โค้ดเปลี่ยน → แก้ docs ใน commit เดียวกัน
- ห้ามเดา: ตรวจไม่ได้ให้เขียน "ยังไม่ตรวจ"; ไม่เก็บ secret/PII (ชื่อ key ได้ ค่าไม่ได้)
- หัวข้อละ 1 ไฟล์ kebab-case; เพิ่มไฟล์แล้วต้องเพิ่มบรรทัดในดัชนีนี้
- ADR ใหม่ → `decisions/YYYY-MM-DD-<slug>.md` (template ใน `decisions/README.md`); บั๊กที่แก้แล้ว → `bugs/YYYY-MM-DD-<symptom>.md` (template ใน `bugs/README.md`)
- `D:\bcdev` (project เก่า Flutter) ไม่เกี่ยวกับ repo นี้ — อย่านำความรู้จากที่นั่นมาปน

## บทความหลัก (อ่านโค้ดทั้ง repo 2026-09-07 ที่ commit d93a210d — 15 นักอ่าน + 15 fact-checker เปิดไฟล์ตรวจทุก citation; ยังไม่ได้ตรวจซ้ำทั้งชุดที่ HEAD `098107e1`)

| # | ไฟล์ | เรื่อง | สรุปบรรทัดเดียว | อ้างอิง path:line | ตรวจซ้ำ (แก้/ยังไม่ตรวจ) |
|---|---|---|---|---|---|
| 00 | [00-source-router.md](00-source-router.md) | แผนที่หาโค้ดตาม feature | router เดิม (AI_INDEX) — ชี้ว่า feature ไหนอยู่ package ไหน | — | — |
| 01 | [01-architecture-overview.md](01-architecture-overview.md) | ภาพรวมสถาปัตยกรรมระบบ BC Ai Account | binary Go ตัวเดียวสลับ 3 บทบาทด้วย DEV_API_MODE, goapi ซ้อนใต้ /goapi ของ mainapi:8888, browser ผ่าน Next.js เสมอ, bootstrap.json ชนะ env | 65 | 175 claims (12/3) |
| 02 | [02-data-stores.md](02-data-stores.md) | 02 — ที่เก็บข้อมูล (Data stores) และบทบาทจริงของแต่ละตัว | สถาปัตยกรรม 2-Tier: Mongo (appdb) เป็น Storage Layer รับเขียน/เก็บย่อ (compact); PG per-holding เป็น Processing Engine ประมวลผลเร็วแบบ self-contained ไม่พึ่ง Mongo อีก; Redis แค่ cache token; MinIO เก็บรูป+thumb; ClickHouse พักแล้ว | 87 | 168 claims (17/6) |
| 03 | [03-auth-tenancy.md](03-auth-tenancy.md) | การยืนยันตัวตน เซสชัน และ Tenancy (Holding / Company / Branch) | token ไม่ใช่ JWT แต่เป็น Redis session 15 นาที/12 ชม. ที่ตรวจสิทธิ์สดจาก Mongo ทุก request; holding ถูกเลือกฝั่ง server ผ่าน /select-holding ไม่ใช่ header; /systemadm/* ไม่มี role guard | 138 | 118 claims (22/0) |
| 04 | [04-product-domain.md](04-product-domain.md) | โดเมนสินค้า (Product / Barcode / Unit / BOM) — Mongo SoT, outbox, projection PG, import, listing v2, จอ frontend | สินค้าเขียนลง Mongo + outbox แล้วค่อย project ไป PG ต่อ holding ผ่าน Kafka; import/barcode2 ยังผูก ClickHouse ที่ถอดไปแล้ว | 148 | 168 claims (5/0) |
| 05 | [05-transaction-sales-purchase.md](05-transaction-sales-purchase.md) | ธุรกรรมซื้อ–ขาย (Transaction: Sales & Purchase) — เส้นทางจาก Mongo → Kafka → PG projection | เอกสารซื้อ-ขาย 19 module: Mongo เป็นต้นทาง, Kafka ยิงแบบ fire-and-forget, goapi project ลง PG doc/docdetail — พร้อมตาราง transflag, สูตรเลขที่เอกสาร และจุดที่โค้ดค้าง (docwaitprocess ไม่ถูกลบ, docref ซ้ำ, PP/PO prefix ชน, quotation/paid/pay ไม่มี consumer) | 180 | 168 claims (27/1) |
| 06 | [06-transaction-stock.md](06-transaction-stock.md) | ธุรกรรมสต็อกและเครื่องคำนวณต้นทุน (Stock transactions & cost engines) | สต็อกมี engine ต้นทุน 3 ชุดในโค้ด แต่ที่ทำงานจริงมีแค่ goapi process-stock (average) — legacy stockprocess ตาย, inventory/costing ยังไม่มีใครใช้, stockwaitprocess ไม่มีใคร drain | 118 | 118 claims (9/2) |
| 07 | [07-goapi-layer.md](07-goapi-layer.md) | ชั้น goapi — HTTP routes, Kafka consumers, workers, และ connection managers | goapi ไม่ใช่ binary แยก: mount ใต้ /goapi ใน mainapi, มี route LIVE ~160 เส้น, handler DEAD 55 ตัว, Kafka 25 group -v1, worker 24 ตัว busy-poll คิวว่าง, ClickHouse เป็น stub | 146 | 160 claims (11/2) |
| 08 | [08-legacy-modules.md](08-legacy-modules.md) | โมดูล legacy/อื่น ๆ ใน backend/internal — inventory, สถานะ และผู้เรียกใช้ | แผนที่โมดูล legacy 40+ ตัวใน backend/internal: ตัวไหน register จริงใน main.go (mode 1/2/3), frontend เรียกอะไรบ้าง, และ 15+ ตัวที่เป็นซากตายพร้อมหลักฐาน path:line + runtime | 221 | 212 claims (10/0) |
| 09 | [09-frontend.md](09-frontend.md) | 09 — Frontend (Next.js BFF): หน้าจอ, API proxy, session, ธีม และกฎ UX | Next 16 BFF: 12 หน้าจอ + 36 route proxy ไป mainapi/goapi, refresh cookie 12 ชม., 10 พาเลต/8 ฟอนต์/12 ภาษา, จอ /settings ตายเพราะ BFF ตอบ 410 | 118 | 148 claims (19/2) |
| 10 | [10-infra-deploy.md](10-infra-deploy.md) | โครงสร้างพื้นฐานและการ deploy (Infra & Deploy) | binary เดียวสลับ 4 โหมดด้วย DEV_API_MODE, local/CI/prod compose 3 ชุด, config โหลด 2 loader ที่ normalize key ไม่เหมือนกัน, prod ยังผูก ClickHouse ที่ stub แล้ว | 124 | 142 claims (11/1) |
| 11 | [11-testing-quality.md](11-testing-quality.md) | การทดสอบและคุณภาพโค้ด (Testing & Quality) | Go test 204 ไฟล์แต่ 15 package ถูก quarantine, ไม่มี CI แล้ว (ลบ 2026-09-09) ใช้ `tools/verify.sh` แทน, Playwright UAT ที่ root เท่านั้นที่ตรวจ Mongo ทีละ step | 71 | 118 claims (11/1) |
| 12 | [12-kafka-messaging.md](12-kafka-messaging.md) | Kafka Messaging — แคตตาล็อก topic / producer / consumer / offset semantics | Kafka มี 384 topic `when-*`, consumer 3 ตระกูล (goapi kafka-go `-v1` / legacy librdkafka / barcode `-projection`) — goapi ส่วนใหญ่ commit offset ก่อน handler เสร็จ (at-most-once) มีเพียง 9 topic product/barcode ที่ ack-after-success, DLQ เป็น log-only | 99 | 118 claims (19/2) |
| 13 | [13-pkg-framework.md](13-pkg-framework.md) | แพ็กเกจ framework กลางของ backend: pkg/microservice, pkg/*, internal/config, utils, models, repositories | แผนที่ framework กลาง (Microservice/persister/cacher/producer/config env/utils/models/repositories) พร้อมสถานะ LIVE/DEAD และกับดักที่ตรวจจากโค้ดจริง | 142 | 190 claims (29/2) |
| 14 | [14-cmd-tools-scripts.md](14-cmd-tools-scripts.md) | เครื่องมือบรรทัดคำสั่งและสคริปต์ (backend/cmd, scripts, tools, scratch, prompts, cluster) | binary จริงมีตัวเดียวคือ root main.go — backend/cmd 30 dir เป็น legacy/dead/one-off, tools/kung_* กับ prompt chat-agent ชี้ route ที่ไม่มีแล้ว, scratch 18 ไฟล์ติด git ทั้งที่ .gitignore ห้าม | 88 | 118 claims (11/0) |
| 15 | [15-known-issues.md](15-known-issues.md) | ปัญหาที่รู้แล้วและงานค้าง (Known Issues & Backlog) — ตรวจซ้ำที่ HEAD d93a210d | 27 issue จาก handoff ตรวจซ้ำทีละบรรทัดที่ HEAD: 21 ยังอยู่ / 3 แก้แล้ว / 3 ตรวจไม่ได้ + พบใหม่ outbox องค์กร 98 PENDING ไม่มี dispatcher | 76 | 118 claims (9/0) |
| 16 | [16-environments-and-servers.md](16-environments-and-servers.md) | สภาพแวดล้อมและเซิร์ฟเวอร์ | dev local / on-prem .202 / prod DigitalOcean / prod เก่า / tunnel / Google sign-in — ค่า secret ไม่อยู่ในนี้ | จาก memory/handoff | ผู้เขียนหลัก |
| 17 | [17-dev-gotchas.md](17-dev-gotchas.md) | กับดักตอนพัฒนา | กับดัก frontend/backend/UAT/เครื่องมือ พร้อมวันที่ยืนยัน — อ่านก่อนเสียเวลาซ้ำ | จาก memory/handoff | ผู้เขียนหลัก |
| 18 | [18-decisions-and-agreements.md](18-decisions-and-agreements.md) | ข้อตกลงและการตัดสินใจ | ลำดับการอ่านสำหรับ AI ทุกตัว + ไทม์ไลน์การตัดสินใจ 2026-06→09 + คำถามค้าง 6 ข้อ | จาก memory/handoff | ผู้เขียนหลัก |
| 19 | [19-menu-coverage-flowaccount-peak.md](19-menu-coverage-flowaccount-peak.md) | ความครบของเมนูเทียบ FlowAccount + PEAK | เมนู 224 รายการครอบคลุมงานของทั้งสองเจ้าเท่าที่หลักฐาน 2 ชุดครอบคลุม (ยกเว้นเงินเดือนที่ตัดออกจากขอบเขต) + วิธีตรวจจอกำพร้า | เว็บทางการ 411 ฟีเจอร์ + แหล่งนอกทางการ 12 มุม 442 ข้อกล่าวอ้าง (2026-09-08) | ตรวจซ้ำแบบหักล้างทั้งสองรอบ |

## เอกสารสถาปัตยกรรม/สัญญา (ย้ายจาก `backend/architecture/` 2026-09-07)

- [architecture/admin-access-control.md](architecture/admin-access-control.md)
- [architecture/high-scale-multitenant-bi.md](architecture/high-scale-multitenant-bi.md)
- [architecture/product-listing-api-v2-handoff.md](architecture/product-listing-api-v2-handoff.md)
- [architecture/product-listing-api-v2.md](architecture/product-listing-api-v2.md)

## การตัดสินใจ (ADR) — 21 ไฟล์ใน `decisions/`

- [decisions/2026-06-10-product-readmodel-parity-taxtype-rename.md](decisions/2026-06-10-product-readmodel-parity-taxtype-rename.md)
- [decisions/2026-06-11-productlanguage-join-table.md](decisions/2026-06-11-productlanguage-join-table.md)
- [decisions/2026-06-30-marketplace-channelmappings-field.md](decisions/2026-06-30-marketplace-channelmappings-field.md)
- [decisions/2026-07-01-warehouse-mongodb-migration.md](decisions/2026-07-01-warehouse-mongodb-migration.md)
- [decisions/2026-07-02-bom-standard-cost-and-sale-calculator.md](decisions/2026-07-02-bom-standard-cost-and-sale-calculator.md)
- [decisions/2026-07-02-bom-subrecipe-live-reference.md](decisions/2026-07-02-bom-subrecipe-live-reference.md)
- [decisions/2026-07-25-names-canonical-code-name.md](decisions/2026-07-25-names-canonical-code-name.md)
- [decisions/2026-09-02-premium-thai-ux-rule.md](decisions/2026-09-02-premium-thai-ux-rule.md)
- [decisions/2026-09-02-screen-action-permissions.md](decisions/2026-09-02-screen-action-permissions.md)
- [decisions/2026-09-02-setup-steps-permissions-before-people.md](decisions/2026-09-02-setup-steps-permissions-before-people.md)
- [decisions/2026-09-03-product-two-layer-marketplace-model.md](decisions/2026-09-03-product-two-layer-marketplace-model.md)
- [decisions/2026-09-03-product-two-user-groups-plan-proposed.md](decisions/2026-09-03-product-two-user-groups-plan-proposed.md)
- [decisions/2026-09-03-tier-endpoints-full-final-state.md](decisions/2026-09-03-tier-endpoints-full-final-state.md)
- [decisions/2026-09-04-product-form-core-vs-extension-menu.md](decisions/2026-09-04-product-form-core-vs-extension-menu.md)
- [decisions/2026-09-06-pause-clickhouse-local.md](decisions/2026-09-06-pause-clickhouse-local.md)
- [decisions/2026-09-07-consolidate-docs-for-multi-ai.md](decisions/2026-09-07-consolidate-docs-for-multi-ai.md)
- [decisions/2026-09-07-mongodb-storage-postgres-processing-clone.md](decisions/2026-09-07-mongodb-storage-postgres-processing-clone.md)
- [decisions/2026-09-07-on-demand-docs-context-efficiency.md](decisions/2026-09-07-on-demand-docs-context-efficiency.md)
- [decisions/2026-09-07-speed-and-context-hygiene.md](decisions/2026-09-07-speed-and-context-hygiene.md)
- [decisions/2026-09-08-menu-parity-flowaccount-peak.md](decisions/2026-09-08-menu-parity-flowaccount-peak.md)
- [decisions/2026-09-08-menu-parity-social-sweep.md](decisions/2026-09-08-menu-parity-social-sweep.md)
- [decisions/2026-09-09-github-storage-only.md](decisions/2026-09-09-github-storage-only.md)

## บั๊กที่แก้แล้ว (symptom → root cause → fix → regression test) — 16 ไฟล์ใน `bugs/`

- [bugs/2026-06-13-product-browser-url-drift-base-href.md](bugs/2026-06-13-product-browser-url-drift-base-href.md)
- [bugs/2026-06-13-productbarcodes-camelcase-naming.md](bugs/2026-06-13-productbarcodes-camelcase-naming.md)
- [bugs/2026-06-21-select-holding-holdingcode-invalid.md](bugs/2026-06-21-select-holding-holdingcode-invalid.md)
- [bugs/2026-06-22-pdf-docdate-off-by-one-utc-timezone.md](bugs/2026-06-22-pdf-docdate-off-by-one-utc-timezone.md)
- [bugs/2026-06-30-barcode-ghost-list-missing-deletedat.md](bugs/2026-06-30-barcode-ghost-list-missing-deletedat.md)
- [bugs/2026-06-30-dimension-items-json-object-vs-array.md](bugs/2026-06-30-dimension-items-json-object-vs-array.md)
- [bugs/2026-07-01-productcategorylist-codelist-and-emptystate.md](bugs/2026-07-01-productcategorylist-codelist-and-emptystate.md)
- [bugs/2026-07-01-warehouse-screen-3-bugs.md](bugs/2026-07-01-warehouse-screen-3-bugs.md)
- [bugs/2026-07-02-product-put-wipes-tenant-identity.md](bugs/2026-07-02-product-put-wipes-tenant-identity.md)
- [bugs/2026-07-02-warehouse-subroutes-4-bugs.md](bugs/2026-07-02-warehouse-subroutes-4-bugs.md)
- [bugs/2026-07-25-mongomodel-relation-drag-wrong-anchor.md](bugs/2026-07-25-mongomodel-relation-drag-wrong-anchor.md)
- [bugs/2026-09-02-dev-login-401-secret-drift.md](bugs/2026-09-02-dev-login-401-secret-drift.md)
- [bugs/2026-09-02-local-upload-invalid-access-key.md](bugs/2026-09-02-local-upload-invalid-access-key.md)
- [bugs/2026-09-02-login-blank-background-tab.md](bugs/2026-09-02-login-blank-background-tab.md)
- [bugs/2026-09-04-refresh-logs-out-non-https.md](bugs/2026-09-04-refresh-logs-out-non-https.md)
- [bugs/2026-09-05-projection-consumer-head-of-line-block.md](bugs/2026-09-05-projection-consumer-head-of-line-block.md)

## Snippets (รูปแบบการเขียนดูที่ `snippets/README.md`)

- [snippets/picklangname.md](snippets/picklangname.md)

## ที่อื่นใน `docs/`

- `docs/handoff/` — สถานะงานค้างระหว่าง session (`HANDOFF-2026-09-08.md` ล่าสุด: เมนู parity + ขอบเขตที่ตัดออก, `HANDOFF-2026-09-06.md` วิธีรัน/งานค้าง backend, `HANDOFF-RISKS-2026-09-05.md` รายละเอียด outbox/projection + audit 3 store, `HANDOFF-2026-09-07-FLOWPEAK.md` = เอกสารประวัติ ห้ามใช้เป็นคำสั่งงานปัจจุบัน)
- `docs/runbooks/RECOVERY-READINESS.md` — runbook กู้คืน prod (รอข้อมูล backup/RPO/RTO จากลุงจืด)
- `docs/reference/CODE-MAP.md` — ดัชนีไฟล์ใหญ่ที่สร้างอัตโนมัติด้วย `tools/gen-code-map.ps1`
- `docs/features-flowaccount-peak/` — หลักฐานฟีเจอร์/เมนูของ FlowAccount และ PEAK (ของคู่แข่ง) ใช้คู่กับบทความ 19
- `docs/skills/` — skill ส่วนตัวของลุงจืด (`ui-scale-polish`, `audit-mongomodel-sync`)
- `docs/archive/` — เอกสารประวัติที่เลิกใช้แล้ว (`flowpeak-legacy-2026-09-07/`) **ห้ามใช้อ้างอิงหรือทำตาม** (เช่น "เมนูรวม 243 รายการ" และเนื้อหาเงินเดือนในนั้นขัดกับขอบเขตปัจจุบัน) อ่านเหตุผลที่ `docs/archive/flowpeak-legacy-2026-09-07/ARCHIVE-NOTE.md` เท่านั้น

## README ที่ยังอยู่ข้างโค้ด (เอกสารเฉพาะ package — อ่านคู่กับบทความด้านบน)

- `backend/README.md`, `backend/VSCODE_RUN.md`, `backend/server/README.md`, `backend/internal/product/product/outbox/README.md`, `backend/internal/goapi/mydb/README.md`, `backend/internal/goapi/mypostgres/README.md`, `backend/internal/goapi/logger/README.md`, `backend/internal/productimport/TESTING.md`, `backend/cmd/datatransfer/README.md`, `backend/cmd/storage_name_migration/README.md`, `backend/cmd/stable_identity_migration/ROLLBACK.md`, `frontend/AGENTS.md` (สร้างโดย Next.js เอง)

## คำถามค้างจากนักอ่าน (รวมจากทุกบทความ — ดูท้ายแต่ละบทความหัวข้อ "ช่องว่าง / สิ่งที่ยังไม่ตรวจ")

รวม 96 ข้อ; ข้อที่ต้องการคำตอบจากลุงจืดโดยตรงสรุปไว้ใน `18-decisions-and-agreements.md` §คำถามที่ยังไม่มีคำตอบ
