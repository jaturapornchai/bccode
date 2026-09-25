# docs/kms — ฐานความรู้ BC Ai Account (Knowledge Management)

ที่เก็บความรู้ที่ "ต้องไม่ลืม" ของโปรเจ็กต์ — ใช้ร่วมกันโดยลุงจืดและ AI ทุกตัว (Claude / Codex / Gemini) เพื่อให้เข้าใจตรงกัน
ตั้งโดยลุงจืด 2026-09-07 · กฎบังคับอยู่ใน `AGENTS.md` (ไฟล์นั้นชนะเสมอ) · skill ส่วนตัวอยู่ที่ `.agents/skills/` · handoff ล่าสุดอยู่ที่ `docs/handoff/`

## ลำดับการอ่าน (On-Demand: โหลดเฉพาะเมื่อจำเป็น ไม่เปลือง Context)

> [!IMPORTANT]
> **ห้ามอ่านเอกสารทั้งหมดพร้อมกันเด็ดขาด**: อ่านเฉพาะไฟล์ที่ตรงกับงานเพื่อประหยัด Context Window ของ AI
> 1. **งานเล็ก / แก้บั๊ก 1 บรรทัด / คำถามทั่วไป** → ดูโค้ดจริงโดยตรง (`code = truth`) ไม่ต้องเปิด docs
> 2. **งาน UX/UI** → อ่านเฉพาะ `.agents/skills/ui-scale-polish/SKILL.md`
> 3. **งาน Schema / ฐานข้อมูล** → อ่าน [02-data-stores.md](02-data-stores.md) แล้วยืนยันกับโค้ดจริง (`backend/internal/centraldb/centraldb.go`, `backend/internal/generalledger/schema.sql`) + สเปกของลุงจืดใน `mydocs/datamodels/gl/` (อ่านอย่างเดียว) — ระบบใช้ PostgreSQL ตัวเดียว
> 3.1 **งานบัญชี/ภาษีไทย (แบบยื่น อัตรา เครดิตภาษี งบการเงิน)** → อ่าน `.agents/skills/thai-accounting-tax/SKILL.md` + ทะเบียน [21-thai-tax-form-references.md](21-thai-tax-form-references.md) — ห้ามเดา ต้องมีอ้างอิงทางการ
> 4. **งานสถาปัตยกรรม / โดเมนเฉพาะเรื่อง** → ดูตารางสรุป 1 บรรทัดด้านล่าง แล้วเลือกเปิดเฉพาะ **1 บทความที่เกี่ยวข้อง**
> 5. **Handoff (`docs/handoff/` — ตอนนี้มี `HANDOFF-2026-09-19.md`, `HANDOFF-2026-09-20-GL-GLM.md`, `HANDOFF-2026-09-20-GL-TAX.md`; การอ้าง `HANDOFF-2026-09-06.md` / `HANDOFF-RISKS-2026-09-05.md` ในบทความ kms เป็นหลักฐานประวัติ ดูได้ด้วย `git show 1660b335:docs/handoff/<file>`)** → อ่านเฉพาะเมื่อลุงจืดถามสถานะงานค้าง/ความเสี่ยง หรือเริ่มงานสถาปัตยกรรมใหญ่ข้ามระบบ

## กติกา

- **โค้ด = ความจริง, docs ตามโค้ด** — ทุกข้อเท็จจริงอ้าง `path:line` และระบุวันที่ตรวจ (บรรทัด 2 ของทุกบทความ); โค้ดเปลี่ยน → แก้ docs ใน commit เดียวกัน
- ห้ามเดา: ตรวจไม่ได้ให้เขียน "ยังไม่ตรวจ"; ไม่เก็บ secret/PII (ชื่อ key ได้ ค่าไม่ได้)
- หัวข้อละ 1 ไฟล์ kebab-case; เพิ่มไฟล์แล้วต้องเพิ่มบรรทัดในดัชนีนี้
- ADR ใหม่ → `decisions/YYYY-MM-DD-<slug>.md` (template ใน `decisions/README.md`); บั๊กที่แก้แล้ว → `bugs/YYYY-MM-DD-<symptom>.md` (template ใน `bugs/README.md`)
- `D:\bcdev` (project เก่า Flutter) ไม่เกี่ยวกับ repo นี้ — อย่านำความรู้จากที่นั่นมาปน

## บทความหลัก (เขียนใหม่ตามโค้ดจริงหลังถอด MongoDB/Kafka/Redis/ClickHouse — ตรวจรอบ 2026-09-25 ที่ HEAD `c58f0b62`; รอบก่อนหน้าอ่านทั้ง repo 2026-09-07 ที่ `d93a210d`)

| # | ไฟล์ | เรื่อง | สรุปบรรทัดเดียว | ตรวจล่าสุด |
|---|---|---|---|---|
| 00 | [00-source-router.md](00-source-router.md) | แผนที่หาโค้ดตามงาน | ชี้ path จริงของ auth/tenancy, GL, ภาษี, สินทรัพย์ถาวร, API/MCP token, infra, จอที่ backend ถูกลบ และชุดตรวจ — ไม่นิยามพฤติกรรมระบบ | 2026-09-25 |
| 01 | [01-architecture-overview.md](01-architecture-overview.md) | ภาพรวมสถาปัตยกรรมระบบ BC Ai Account | binary Go ตัวเดียว (`backend/main.go`) ไม่มีโหมดสลับ, goapi mount ใต้ `/goapi` ของ mainapi:8888, browser ผ่าน Next.js BFF เสมอ, เก็บข้อมูลใน PostgreSQL + MinIO เท่านั้น | 2026-09-25 |
| 02 | [02-data-stores.md](02-data-stores.md) | ที่เก็บข้อมูลและบทบาทจริงของแต่ละตัว | PostgreSQL ฐานกลาง `bcai_projection` (identity/holding/สมาชิก/สิทธิ์/`cache_entries`/MCP token) + ฐานแยกต่อ holding (GL/ภาษี/สินทรัพย์ถาวร สร้างอัตโนมัติ) + MinIO เก็บไฟล์คู่ thumbnail | 2026-09-25 |
| 03 | [03-auth-tenancy.md](03-auth-tenancy.md) | การยืนยันตัวตน เซสชัน และ Tenancy (Holding / Company / Branch) | token ไม่ใช่ JWT แต่เป็น opaque token ในตาราง `cache_entries` (access ≤ 15 นาที, session 12 ชม.), ตรวจสิทธิ์สดจาก PostgreSQL ทุกคำขอ, holding เลือกฝั่ง server ผ่าน `/select-holding`, login 4 ทาง (password/Google/Demo/Dev) | 2026-09-25 |
| 04 | [04-product-domain.md](04-product-domain.md) | โดเมนสินค้า (Product / Barcode / Unit / BOM) | backend สินค้าถูกลบ 2026-09-23 — จอเดิมขึ้น "รอพัฒนา", เหลือ route goapi ค้างที่อ่านตารางที่ไม่มีใครเติม + อัปโหลดรูป/วิดีโอลง MinIO ที่ยังใช้ได้ | 2026-09-25 |
| 05 | [05-transaction-sales-purchase.md](05-transaction-sales-purchase.md) | ธุรกรรมซื้อ–ขาย | module ซื้อ-ขายเดิมถูกลบ 2026-09-23 — frontend มี registry `ERP_MODULE_CONFIGS` รอ backend บน PostgreSQL, route goapi รายงานขายที่ยังค้าง | 2026-09-25 |
| 06 | [06-transaction-stock.md](06-transaction-stock.md) | ธุรกรรมสต็อกและเครื่องคำนวณต้นทุน | ทางเข้าเอกสารสต็อกเดิมถูกลบทั้งหมด — เหลือ engine ต้นทุนฝั่ง goapi (`stockengine`, `process-stock`, `inventory`) ที่ไม่มีข้อมูลป้อน | 2026-09-25 |
| 07 | [07-goapi-layer.md](07-goapi-layer.md) | ชั้น goapi — HTTP routes, connection managers และ stock engine | goapi mount ใต้ `/goapi` ใน mainapi: ลำดับ `Init()`, middleware/auth ของตัวเอง, ตาราง route, handler DEAD, stock engine และคิวที่ไม่มีผู้ใช้, connection manager `mydb`/`mypg` | 2026-09-25 |
| 08 | [08-legacy-modules.md](08-legacy-modules.md) | โมดูล legacy/อื่น ๆ ใน `backend/internal` | เกือบทุกโมดูล legacy ถูกลบ 2026-09-23 (`backend/cmd/` เหลือแค่ `glseed/`) — เหลือ `media`, `demo`, `encrypt`, `models`, `utils` และสคริปต์ SQL ใน `backend/internal/database/` | 2026-09-25 |
| 09 | [09-frontend.md](09-frontend.md) | Frontend (Next.js BFF): หน้าจอ, API proxy, session, ธีม และกฎ UX | Next 16 BFF: หน้าจอ, route proxy ไป mainapi/goapi, refresh cookie, ธีม/ฟอนต์/ภาษา — BFF ที่ปลายทาง backend ถูกลบติดป้าย BACKEND-REMOVED (จอขึ้น "รอพัฒนา") | 2026-09-07 + ตรวจซ้ำบางหัวข้อ 2026-09-25 |
| 10 | [10-infra-deploy.md](10-infra-deploy.md) | โครงสร้างพื้นฐานและการ deploy | local stack (Docker Desktop) + prod compose `deploy/account/compose.yml` (postgres, minio, mainapi, frontend), config bootstrap.json → env, deploy ด้วย `tools/fast-deploy.py` | 2026-09-25 |
| 11 | [11-testing-quality.md](11-testing-quality.md) | การทดสอบและคุณภาพโค้ด | Go test บน PostgreSQL (integration ด้วย tag), ไม่มี CI แล้ว ใช้ `tools/verify.sh`, vitest/Playwright, UAT ด้วยปุ่ม Demo + ตรวจ PostgreSQL ทีละ step | 2026-09-25 |
| 13 | [13-pkg-framework.md](13-pkg-framework.md) | แพ็กเกจ framework กลางของ backend | `pkg/microservice` (echo, cacher `cache_entries`, persister ไฟล์ S3/MinIO, live authorization), `pkg/*` อื่น, ชื่อ env ใน `internal/config`, utils/models และกับดักที่ตรวจจากโค้ด | 2026-09-25 |
| 14 | [14-cmd-tools-scripts.md](14-cmd-tools-scripts.md) | เครื่องมือบรรทัดคำสั่งและสคริปต์ | binary จริงตัวเดียวคือ `backend/main.go`, `backend/cmd/glseed`, สคริปต์ seed/UAT ใน `scripts/`, เครื่องมือใน `tools/`, `scratch/` ที่ติด git, prompt assets | 2026-09-25 |
| 15 | [15-known-issues.md](15-known-issues.md) | ปัญหาที่รู้แล้วและงานค้าง | issue เดิมตรวจซ้ำที่ HEAD — ข้อที่ผูกกับ Kafka/Mongo ปิดเพราะโค้ดถูกลบ, ที่เหลือแยกยังอยู่/แก้แล้ว + คำถามเปิดถึงลุงจืด | 2026-09-25 |
| 16 | [16-environments-and-servers.md](16-environments-and-servers.md) | สภาพแวดล้อมและเซิร์ฟเวอร์ | DEV เครื่องลุงจืด + PROD DigitalOcean `159.223.43.229` (`account.bcaicloud.com`), Google Sign-In — ค่า secret ไม่อยู่ในนี้ | 2026-09-25 |
| 17 | [17-dev-gotchas.md](17-dev-gotchas.md) | กับดักตอนพัฒนา | กับดัก frontend/backend/UAT/สคริปต์/เครื่องมือ AI พร้อมวันที่ยืนยัน (รวม regex `\p{L}` ไม่รับสระ/วรรณยุกต์ไทย) — อ่านก่อนเสียเวลาซ้ำ | 2026-09-07 + ล้างแถวยุคเก่า 2026-09-25 |
| 18 | [18-decisions-and-agreements.md](18-decisions-and-agreements.md) | ข้อตกลงและการตัดสินใจ | ลำดับการอ่านสำหรับ AI ทุกตัว + ไทม์ไลน์การตัดสินใจ + คำถามที่ยังไม่มีคำตอบ | 2026-09-25 |
| 19 | [19-menu-coverage-market-standard.md](19-menu-coverage-market-standard.md) | ความครบของเมนูเทียบมาตรฐานโปรแกรมบัญชีในตลาด | ประวัติการเทียบเมนูกับมาตรฐานตลาด (เคย 224 รายการ) — **ตั้งแต่ 2026-09-19 ยึด Champ parity ดู ADR 2026-09-19-champ-parity-menu-cut** + วิธีตรวจจอกำพร้า | ผัง 2026-09-11 (เนื้อหาเทียบตลาดเป็นประวัติ 2026-09-08) |
| 20 | [20-champ-parity-gap.md](20-champ-parity-gap.md) | ช่องว่างเทียบต้นแบบ Champ (`D:\project-champ`) | ผังเมนู, รายงาน/ภาษี/เครื่องมือที่ต่างเชิงระบบ, BCProcess Windows Service ที่ BC ยังไม่มี worker | Champ 2026-09-15; ฝั่ง BC 2026-09-25 |
| 21 | [21-thai-tax-form-references.md](21-thai-tax-form-references.md) | ทะเบียนอ้างอิงทางการของแบบยื่นภาษี | ช่องไหนของแบบ (ภ.ง.ด.50/51 เครดิตภาษี + สูตรข้อ 3–8) ระบบเติม/คำนวณ เพราะคำชี้แจงกรมสรรพากรข้อไหน + สิ่งที่ยังไม่เติมเพราะข้อมูลไม่พอ (ห้ามเดา); งวดภาษีหัก/กลับรายการ (§9), ภาษีซื้อ ใบเพิ่ม/ลดหนี้ ใบกำกับซ้ำ (§10), mapping ไฟล์ยื่นด้วยสื่อ Format กลาง V2.0 (§11), หลักตรวจสอบเลขผู้เสียภาษี (§12) — คู่กับ skill `thai-accounting-tax` | 2026-09-23–24 |
| 22 | [22-sample-data-gl-tax-2569.md](22-sample-data-gl-tax-2569.md) | ข้อมูลบัญชีและภาษี ก.ค.–ก.ย. 2569 ของ rungrueng/01 | สคริปต์ `scripts/seed-gl-tax-rungrueng-2569.mjs` (dry-run เป็นค่าเริ่มต้น, `--apply`, `verify`), ใบรายวัน 31 ใบ + แบบยื่น 7 ฉบับ, สมุดเลือกตามประเภทและจำใน manifest, ค่าสมมติที่ต้องล้างก่อน go-live | 2026-09-25 |

บทความ `12-kafka-messaging.md` ถูกลบ 2026-09-25 เพราะ Kafka ถูกถอดออกจากระบบแล้ว ([ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — เลข 12 เว้นว่างไว้

## เอกสารสถาปัตยกรรม/สัญญา (`architecture/`)

- [architecture/2026-09-11-general-ledger-v2.md](architecture/2026-09-11-general-ledger-v2.md) — บัญชีแยกประเภทตาม Champ: PostgreSQL ล้วนแบบ synchronous ACID ตั้งแต่ 2026-09-23, decimal exact, ผ่าน/กลับรายการ, ปิดงบ/สิ้นปี, ผลทดสอบและข้อจำกัด XBRL/เอกสารต้นทาง
- [architecture/gl-journal-details.md](architecture/gl-journal-details.md) — หลักฐานลูกหนี้/เจ้าหนี้/Statement ธนาคารแบบ many-to-many ภายในรายวัน, source dedup, API/MCP และข้อจำกัด
- [architecture/gl-journal-review.md](architecture/gl-journal-review.md) — ผลตรวจและข้อแตกต่างในหน้ารายวัน แยกจากสถานะผ่านรายการ พร้อมประวัติผู้ตรวจ/รุ่นเอกสาร
- [architecture/gl-mcp-tokens.md](architecture/gl-mcp-tokens.md) — API / MCP token ของ Holding (`/mcp-tokens`, `/mcp/gl`): credential แยกจาก session, ตรวจสิทธิ์ผู้ออกทุกคำขอ, เพิกถอนได้
- [architecture/admin-access-control.md](architecture/admin-access-control.md) — ข้อเสนอออกแบบ admin หลายบริษัท (**ยังไม่ได้ทำ** — ตาราง/API ในเอกสารไม่มีใน `backend/`; โมเดลปัจจุบันคือ `holding_members` ในฐานกลาง)

## การตัดสินใจ (ADR) — 70 ไฟล์ใน `decisions/`

> ADR และบั๊กเป็นบันทึกประวัติ ไม่แก้ย้อนหลัง — ฉบับก่อน 2026-09-23 ที่อธิบายการออกแบบบน MongoDB/Kafka/Redis/ClickHouse (2-Tier, outbox, projection, consumer) ใช้เป็นหลักฐานประวัติเท่านั้น ระบบปัจจุบันยึด [ADR ถอดระบบ 2026-09-23](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)

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
- [decisions/2026-09-08-menu-parity-market-standard.md](decisions/2026-09-08-menu-parity-market-standard.md)
- [decisions/2026-09-08-menu-parity-social-sweep.md](decisions/2026-09-08-menu-parity-social-sweep.md)
- [decisions/2026-09-09-github-storage-only.md](decisions/2026-09-09-github-storage-only.md)
- [decisions/2026-09-10-holding-scope-menu-deduplication.md](decisions/2026-09-10-holding-scope-menu-deduplication.md)
- [decisions/2026-09-10-master-menu-streamlining.md](decisions/2026-09-10-master-menu-streamlining.md)
- [decisions/2026-09-11-erp-nine-modules-menu-restructure.md](decisions/2026-09-11-erp-nine-modules-menu-restructure.md)
- [decisions/2026-09-11-champ-menu-structure-alignment.md](decisions/2026-09-11-champ-menu-structure-alignment.md)
- [decisions/2026-09-11-deploy-champ-menu.md](decisions/2026-09-11-deploy-champ-menu.md) — Deploy frontend 223 เมนู Champ, ชื่อหมวดไทยล้วน และเมนูบนเป็นค่าเริ่มต้น (r20260911-menu-top-1)
- [decisions/2026-09-11-deploy-gl-kafka-demo.md](decisions/2026-09-11-deploy-gl-kafka-demo.md) — ประวัติ deploy GL รอบ r20260911-gl-kafka-1 + frontend gl-demo-thai-1 (เส้นทาง Kafka ในเอกสารนี้ถูกถอดแล้ว 2026-09-23)
- [decisions/2026-09-11-deploy-general-ledger-v2.md](decisions/2026-09-11-deploy-general-ledger-v2.md) — Main API/worker/frontend GL v2, สำรองและ restore แยก, ตรวจ HTTPS 35 เมนูและ rollback
- [decisions/2026-09-11-deploy-chart-of-accounts-level-and-delete-guard.md](decisions/2026-09-11-deploy-chart-of-accounts-level-and-delete-guard.md) — Deploy ผังบัญชีรองรับระดับ 1–12, ป้องกันการลบเมื่อมีข้อมูลอ้างอิงจากสมุดรายวันเด็ดขาด (r20260911-gl-level-1)
- [decisions/2026-09-11-champ-upgrade-menu-workflows.md](decisions/2026-09-11-champ-upgrade-menu-workflows.md) — ผังอัปเกรด Champ 223 เมนู, baseline 154, 69 คำสั่งเพิ่มและสถานะจอจริง
- [decisions/2026-09-19-champ-parity-menu-cut.md](decisions/2026-09-19-champ-parity-menu-cut.md) — ตัดเมนูเกิน Champ 47 + ถอด engine ฝั่ง browser (Health Audit/AI Copilot/CFO/Templates/กระทบยอด/e-Tax XML/Bank Feeds), เติมของที่ Champ มี 33, เรียงเมนูตาม menuconfig.xml, เมนู 208→194, เปิดใช้ 2 ภาษา th/en; แก้ handoff ที่ระบุ goods-inspection/low-stock-alert/sale-reservation-flow ผิด
- [decisions/2026-09-11-financial-statement-designer.md](decisions/2026-09-11-financial-statement-designer.md) — ออกแบบงบการเงิน (Financial Statement Designer), รองรับหลายประเภทงบ, ปรับแต่งฟอนต์อิสระ, สูตรคำนวณสด และสร้างได้ไม่จำกัด
- [decisions/2026-09-11-deploy-financial-statement-designer.md](decisions/2026-09-11-deploy-financial-statement-designer.md) — Deploy ระบบออกแบบงบการเงิน สู่ Production (r20260911-gl-statement-1) พร้อมผลการตรวจ Live CRUD และ Healthcheck 100%
- [decisions/2026-09-12-fullscreen-account-search-dialog.md](decisions/2026-09-12-fullscreen-account-search-dialog.md) — ระบบค้นหาผังบัญชีแบบเต็มจอ (Full-Screen Chart of Accounts Search Dialog) พร้อมตัวกรอง 5 หมวดและคีย์ลัดสำหรับผู้ใช้ 40+
- [decisions/2026-09-12-enable-dom-inspector-on-production.md](decisions/2026-09-12-enable-dom-inspector-on-production.md) — เปิดใช้งานวิดเจ็ต Copy DOM (DevDomInspector) บน Production (account.bcaicloud.com)
- [decisions/2026-09-12-deploy-search-and-copy-dom.md](decisions/2026-09-12-deploy-search-and-copy-dom.md) — Deploy ระบบค้นหาผังบัญชีแบบเต็มจอและ Copy DOM สู่ Production (r20260912-search-dom-1)
- [decisions/2026-09-12-chart-of-accounts-tree-sort-and-search-icon.md](decisions/2026-09-12-chart-of-accounts-tree-sort-and-search-icon.md) — แก้การเรียงผังบัญชีแบบต้นไม้ (ระดับที่แท้จริง) และไอคอนในช่องค้นหาโดนทับ
- [decisions/2026-09-12-baseline-search-debounce-and-clean-icon.md](decisions/2026-09-12-baseline-search-debounce-and-clean-icon.md) — ระบบค้นหาหลัก Baseline Toolbar ค้นหาอัตโนมัติ (Auto 2s debounce) และปุ่ม Clean (✕) ล้างคำค้น
- [decisions/2026-09-12-deploy-baseline-search-auto-and-clean.md](decisions/2026-09-12-deploy-baseline-search-auto-and-clean.md) — Deploy แถบค้นหาหลัก Baseline Toolbar (Auto Search 2s & Clean Icon) สู่ Production (r20260912-search-auto-1)
- [decisions/2026-09-12-formatted-numeric-input.md](decisions/2026-09-12-formatted-numeric-input.md) — มาตรฐานช่องกรอกตัวเลขและการแสดงผลจำนวนเงิน (Formatted Numeric Input Standard — Comma, Decimal, Right-Aligned & Clean Edit Mode)
- [decisions/2026-09-12-gl-masters-crud-standard.md](decisions/2026-09-12-gl-masters-crud-standard.md) — สถาปัตยกรรมตารางข้อมูลหลักระบบบัญชีมาตรฐาน CRUD (GL Masters CRUD Table Parity — Actions Column, Amount Column, Status Badges & Quick Delete)
- [decisions/2026-09-12-fast-deploy-standard.md](decisions/2026-09-12-fast-deploy-standard.md) — มาตรฐานการ Deploy แบบเร็วที่สุด (Fast Streamed Zero-Disk Deploy) และกฎเสร็จแล้ว Deploy ทันที
- [decisions/2026-09-13-crud-view-edit-separation.md](decisions/2026-09-13-crud-view-edit-separation.md) — มาตรฐานการแยกโหมดแสดงข้อมูลและโหมดแก้ไขใน CRUD Table (Row Click View Mode, No Save Button in View, Explicit Edit Mode)
- [decisions/2026-09-14-shorten-nine-module-menu-titles.md](decisions/2026-09-14-shorten-nine-module-menu-titles.md) — ตัดคำว่า "ระบบ" และ "ระบบบัญชี" ออกจากชื่อเมนูหลัก 9 ระบบ ERP เพื่อแก้ปัญหาเมนูล้นจอแนวนอน (Shorten 9 Module Menu Titles)
- [decisions/2026-09-14-menu-bar-flex-wrap.md](decisions/2026-09-14-menu-bar-flex-wrap.md) — ปรับแถบเมนูนำทางด้านบนให้ตัดขึ้นบรรทัดใหม่ (Flex Wrap) แทนการมีแถบเลื่อนแนวนอน (No Horizontal Scroll)
- [decisions/2026-09-14-gl-workbench-full-height-expanded.md](decisions/2026-09-14-gl-workbench-full-height-expanded.md) — จอ GL workbench ขยายความสูงเต็มพื้นที่ที่เหลืออัตโนมัติ (viewport lock ใน `main-menu-screen.tsx`)
- [decisions/2026-09-14-input-addon-icons-no-overlap.md](decisions/2026-09-14-input-addon-icons-no-overlap.md) — แก้ไขปัญหาไอคอนในช่องเลือกผังบัญชีซ้อนทับกัน (Fix AccountSelect Addon Icons Overlap)
- [decisions/2026-09-15-deploy-gl-thai-accounting-firm-and-champ-parity.md](decisions/2026-09-15-deploy-gl-thai-accounting-firm-and-champ-parity.md) — Deploy ระบบบัญชีแยกประเภท วงจรบัญชีมาตรฐานไทย และแก้ไขระบบรายงาน สู่ Production (r20260915-1)
- [decisions/2026-09-15-fixed-assets-and-depreciation-engine.md](decisions/2026-09-15-fixed-assets-and-depreciation-engine.md) — ระบบบริหารสินทรัพย์ถาวรและการคำนวณค่าเสื่อมราคา (Fixed Assets & Depreciation Engine) ตามต้นแบบ Champ, GL Posting, สิทธิประโยชน์ภาษี และ MCP Tools สู่ Production (r20260915-fa-1)
- [decisions/2026-09-15-background-process-engine.md](decisions/2026-09-15-background-process-engine.md) — ข้อเสนอเดิม (รออนุมัติ) ให้แทน timer ของ BCProcess ด้วย dirty-set + worker LISTEN/NOTIFY — ถูกแทนด้วย stock-cost-engine-v2 ด้านล่าง; ส่วนที่อิง Kafka ใช้ไม่ได้แล้วหลัง 2026-09-23
- [decisions/2026-09-15-stock-cost-engine-v2.md](decisions/2026-09-15-stock-cost-engine-v2.md) — ออกแบบเครื่องยนต์สต็อก/ต้นทุนใหม่ทั้งชุด: PK ทางธุรกิจ, total order key ที่ deterministic, checkpoint ต่องวด, เขียนผลใน transaction เดียว, advisory lock, dirty-set (แทน ADR worker ฉบับก่อน)
- [decisions/2026-09-15-erp-datacrud-workbench-standard.md](decisions/2026-09-15-erp-datacrud-workbench-standard.md) — ระบบธุรกรรม ERP แบบ Master-Detail DataCRUD (สินค้า, ขาย, ซื้อ, ลูกหนี้, เจ้าหนี้, เงินสดธนาคาร) ตามมาตรฐาน datacrud skill สู่ Production (r20260915-datacrud-1)
- [decisions/2026-09-15-complete-menu-coverage-standard.md](decisions/2026-09-15-complete-menu-coverage-standard.md) — ระบบรองรับหน้าจอที่รอพัฒนาครบ 100% สำหรับบัญชีและ SME ไทย (ภาษี, รายงาน, เครื่องมือประมวลผล, ปฏิบัติการ SME) สู่ Production (r20260915-all-screens-1)
- [decisions/2026-09-16-menu-labels-follow-champ-wording.md](decisions/2026-09-16-menu-labels-follow-champ-wording.md) — ป้ายเมนูและชื่อฟิลด์ใช้ถ้อยคำเดียวกับระบบเดิม Champ (เปลี่ยนป้าย 103 เมนู, ชื่อเดิมกลายเป็นคำค้น, ห้ามแตะ id/route/language key)
- [decisions/2026-09-20-single-source-ai-rules-skills.md](decisions/2026-09-20-single-source-ai-rules-skills.md) — กฎ+skill ของ AI ทุกตัวอยู่ที่เดียว: `AGENTS.md` + `.agents/skills/` (ย้ายจาก `docs/skills/`), Gemini CLI ผ่าน `.gemini/settings.json`, Claude ผ่าน junction `npm run ai:link`
- [decisions/2026-09-23-drop-readme-activity-log.md](decisions/2026-09-23-drop-readme-activity-log.md) — เลิก Activity Log ใน README (ประวัติ = git log + body ไทย), ถอด README guard จาก pre-commit, ล้างกฎ Mongo/Redis/12 ภาษาที่ปลดระวางออกจาก AGENTS.md
- [decisions/2026-09-23-wht-certificate-pdf-backend.md](decisions/2026-09-23-wht-certificate-pdf-backend.md) — ใบ 50 ทวิ: backend (`internal/whtcert`) สร้าง PDF บนแบบฟอร์มจริงของกรมสรรพากร (ตำแหน่งวัดจากฟอร์ม, ไทยจัดรูปด้วย HarfBuzz, ฉบับที่ 1/2 + สำเนาคู่ฉบับ/ใบแทน) frontend แค่แสดง
- [decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md) — ถอด MongoDB/Kafka/Redis/ClickHouse: session → `cache_entries`, ตรวจสิทธิ์สดจาก PG ทุกคำขอ, จอที่ API อยู่บน Mongo ขึ้น "รอพัฒนา"
- [decisions/2026-09-23-tax-inside-general-ledger.md](decisions/2026-09-23-tax-inside-general-ledger.md) — เมนูภาษีทั้งหมดอยู่ใต้ "บัญชีแยกประเภท" ตาม Champ; ผู้ใช้ GL ตัวเดียวใช้ได้เลย, ภ.ง.ด./50 ทวิ อ่านจาก GL
- [decisions/2026-09-23-wht-tax-base-editable.md](decisions/2026-09-23-wht-tax-base-editable.md) — ฐานภาษีหัก ณ ที่จ่ายแก้ได้เสมอ: เก็บใน `details.withholdings` ของใบสำคัญ (ภาษีว่าง = ฐาน × อัตรา), แก้หลังผ่านบัญชีผ่าน reconcile + เหตุผล + audit, รายงาน ภ.ง.ด./50 ทวิ ใช้ฐานที่บันทึก (recorded) ก่อนค่าประมาณ (inferred)
- [decisions/2026-09-23-rd-tax-forms-engine.md](decisions/2026-09-23-rd-tax-forms-engine.md) — แบบยื่นกรมสรรพากรทุกแบบใน `mydocs/sample` (ภ.ง.ด.2/2ก/3/53/50/51/93/94, ภ.พ.30/36, ภ.ธ.40): `internal/rdform` สเปก JSON → PDF ทางการ, API `/api/report/tax/form/*` ดึงยอดจาก GL/แก้ได้ทุกช่อง/บันทึก `tax_filings`, จอ `TaxFormEditor`; ลบแบบ HTML + `pp30-summary`
- [decisions/2026-09-19-champ-parity-no-bloat-rule.md](decisions/2026-09-19-champ-parity-no-bloat-rule.md) — กฎยึด D:\project-champ เป็นต้นแบบหลัก ไม่เพิ่มฟังก์ชันหรือเมนูมากเกินไป เพื่อมุ่งเน้นการ Upgrade จาก Windows สู่ Web ที่รวดเร็ว ปลอดภัย และไม่ทำให้ลูกค้าสับสน
- [decisions/2026-09-24-holding-wide-access-scope.md](decisions/2026-09-24-holding-wide-access-scope.md) — "ใช้ได้ทั้งกลุ่มกิจการ" เก็บเป็นกฎ holding กฎเดียว ขยายตอนอ่านเป็นทุกบริษัท active (รวมที่เพิ่มภายหลัง); ตรวจทุกกฎตอนบันทึก (400), ADMIN ให้เกินขอบเขตตัวเองไม่ได้ (403), frontend เลิกแปลงกฎเสียเป็น holding
- [decisions/2026-09-24-user-defined-journal-books.md](decisions/2026-09-24-user-defined-journal-books.md) — สมุดรายวันเป็น master ที่ผู้ใช้กำหนดเอง (`booktype` 1–6 ตาม `journalbook.sql`, SV=ขาย UV=ซื้อ): โค้ดเลือกสมุดตามประเภทไม่ยึดรหัส, สมุดที่มีเอกสารหรือรูปแบบการเชื่อมบัญชีอ้างถึงลบไม่ได้ให้ปิดใช้งาน, สมุดเก่าไม่มีประเภทผู้ใช้กำหนดเองตามชื่อ (ไม่เดาจากรหัส), สิทธิ์ `gl-journals`/`gl-opening-balance`/`gl-post` แทนสิทธิ์รายสมุด, สาขาหัวเอกสารบังคับเมื่อเข้าระบบระดับบริษัท
- [decisions/2026-09-24-rd-wht-file-export.md](decisions/2026-09-24-rd-wht-file-export.md) — ไฟล์ยื่นภาษีหัก ณ ที่จ่ายด้วยสื่อ = .txt Format กลาง V2.0 สำหรับโปรแกรม SWC-UI (ภ.ง.ด.53/3/2 รายเดือน) จากฉบับที่บันทึกแล้ว — **ไม่ใช่ไฟล์อัปโหลด e-Filing**; ไม่ทำ .rdx/XML/Open API/RD Prep รอบนี้
- [decisions/2026-09-24-remove-legacy-login-flows.md](decisions/2026-09-24-remove-legacy-login-flows.md) — ลบ flow login ที่ไม่มี route: `LoginEmail` (ออก token โดยไม่ตรวจรหัสผ่าน), Firebase `TokenLogin` + package/dependency, LINE login + package + `LINE_CLIENT_ID`; เก็บ Google/Demo/Dev login และการเชื่อมบัญชี LINE (`/profile/link-line` ยังมี UI เรียก แต่ prod ไม่มี `BC_AUTH_BRIDGE_URL`)
- [decisions/2026-09-25-company-tax-address.md](decisions/2026-09-25-company-tax-address.md) — ที่อยู่สำหรับภาษีของบริษัท (สำนักงานใหญ่) เก็บในทะเบียนบริษัทแยก 13 ช่องตาม key `addr_*` ของแบบ + โทรศัพท์ (Champ เก็บข้อความก้อนเดียว แต่แบบกรมสรรพากรแยกช่อง); prefill หัวแบบใช้ที่อยู่ทะเบียนแทนฉบับก่อน, 50 ทวิ ใช้ที่อยู่ทะเบียนเมื่อใบไม่มี, ยังแก้รายฉบับได้; กรุงเทพฯ ใช้ แขวง/เขต (พ.ร.บ.ฯ กทม. 2528 ม.8); สาขาและ snapshot ตอนบันทึก = งานอนาคต
- [decisions/2026-09-25-gl-monthly-budget.md](decisions/2026-09-25-gl-monthly-budget.md) — งบประมาณรายเดือนต่อบัญชี (Champ 5500 `BCGLBudget`): ตาราง `gl_budgets` + `gl_budget_lines` (บัญชี × งวด 1–12, `numeric(18,2)`), คำสั่ง `budgets` create/update/delete/spread ผ่าน `Execute`, สถานะเปิด/ปิดแบบ Champ (ไม่มีอนุมัติ), รายงาน `budgetcomparison` (Champ 5530) เทียบ `gl_lines` ที่ผ่านบัญชีตาม `normal_balance`; จอ `/gl/budget` ขึ้น "ยังไม่พร้อม" จนกว่าจะทำจองบรายเดือน; ส่วนที่ต่างจาก Champ โดยตั้งใจอยู่ในตาราง ADR
- [decisions/2026-09-25-retire-mongo-era-docs-and-skills.md](decisions/2026-09-25-retire-mongo-era-docs-and-skills.md) — ล้างเอกสาร/skill ยุค MongoDB/Kafka/Redis/ClickHouse: เขียนบทความ kms ใหม่ตามโค้ด, ลบ `12-kafka-messaging.md` + architecture/snippet ที่ตาย, ลบ skill `audit-mongomodel-sync` และ server `mongodb`/`mongomodel` ใน `.mcp.json`

## บั๊กที่แก้แล้ว (symptom → root cause → fix → regression test) — 35 ไฟล์ใน `bugs/`

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
- [bugs/2026-09-14-settings-language-switch-loop.md](bugs/2026-09-14-settings-language-switch-loop.md) — **ยังไม่แก้ (open)**: จอตั้งค่าเปิดผ่าน URL ตรงแล้วกดเลือกภาษา → render loop สลับ th/ja (effect อ่าน localStorage ผูก deps กับ loadRecords ที่มี language)
- [bugs/2026-09-23-vat-report-reads-missing-erp-tables.md](bugs/2026-09-23-vat-report-reads-missing-erp-tables.md) — **แก้แล้ว 2026-09-23**: รายงานภาษีขาย/ซื้อ + ภ.พ.30 เคยอ่านตาราง ERP ที่ไม่มีในฐาน holding (prod 500 ทุก holding) → อ่านจากรายละเอียดภาษีมูลค่าเพิ่มของใบสำคัญ GL ที่ผ่านบัญชี (`details.vats`, `generalledger.VatRecordsForPeriod`) ตาม `mydocs/datamodels/gl/vat.sql`
- [bugs/2026-09-23-wht-cert-pdf-panic-on-prod.md](bugs/2026-09-23-wht-cert-pdf-panic-on-prod.md) — 50 ทวิ บน prod panic เพราะ pdfcpu mkdir ใน home ที่ไม่มี (แก้ `api.DisableConfigDir`) + เลขผู้เสียภาษีบริษัทใน seed ผิด checksum + คอลัมน์ `business_code` ขาดใน `shop_user_access_logs`
- [bugs/2026-09-23-wht-received-report-empty.md](bugs/2026-09-23-wht-received-report-empty.md) — **แก้แล้ว 2026-09-23**: รายงานภาษีถูกหักว่างบน prod เพราะหาบัญชีจากคำว่า "ภาษีถูกหัก" ติดกัน (ชื่อมาตรฐาน "ภาษีเงินได้ถูกหัก ณ ที่จ่าย" ไม่ตรง) — ตอนนี้อ่านรายการภาษีหักที่บันทึกในใบสำคัญเป็นหลัก
- [bugs/2026-09-24-unauthenticated-setup-api-ssrf.md](bugs/2026-09-24-unauthenticated-setup-api-ssrf.md) — **แก้แล้ว 2026-09-24 (critical)**: Setup API ของจอ `/settings` ไม่ตรวจ session รับรหัสว่าง/`12345`/`admin` และเปิด SSRF/สแกนพอร์ตภายในจาก URL สาธารณะ → ลบจอ + BFF `/api/setup/*` + `/api/storage/health` ทั้งชุด (config ที่บันทึกไม่มีใครอ่าน) + guard test
- [bugs/2026-09-23-login-accounts-save-edit-broken.md](bugs/2026-09-23-login-accounts-save-edit-broken.md) — **แก้แล้ว 2026-09-23**: จอบัญชีเข้าระบบเพิ่มผู้ใช้ไม่ได้ (`permissionsets` ค่าเริ่มต้น `{}`) และแก้ไขไม่ได้ ("find failed" — repo PostgreSQL คืน error ไม่พบแทนค่าว่าง ทำให้ไม่ fallback ไป useruid)
- [bugs/2026-09-24-tax-forms-reports-uat.md](bugs/2026-09-24-tax-forms-reports-uat.md) — **แก้แล้ว 2026-09-24**: รายงาน/แบบยื่นภาษี — ใบกลับรายการ/ยกมา/ปิดบัญชี/โอนยอดระหว่างบัญชีภาษีหักไม่เป็นแถวผี, ภาษีหักที่บันทึกเข้างวดตามวันที่จ่าย (ประกาศฯ ฉบับที่ 111), ส่วนที่ยังไม่บันทึกเป็นแถวประมาณ, VAT ไม่ระบุอัตราถูกปฏิเสธ, กรอบใช้สิทธิภาษีซื้อ 6 เดือน, ใบกำกับซ้ำเทียบผู้ออก+เลขที่+วันที่ (ม.86/4, เตือนเท่านั้น), ยอดรวมแบบยื่นคำนวณใหม่ที่ backend
- [bugs/2026-09-24-gl-journal-entry-uat-findings.md](bugs/2026-09-24-gl-journal-entry-uat-findings.md) — **แก้แล้ว 2026-09-24**: UAT บันทึกรับเงินที่ถูกหักภาษี — dialog ค้นบัญชีรีเซ็ตทุก render, VAT guard เดาจากรหัสบัญชี (ห้ามใช้รหัสบัญชีเป็นเงื่อนไข), ใบใหม่ส่ง `branchcode` ว่าง, รายงานภาษีหักนับใบที่กลับรายการ
- [bugs/2026-09-24-save-blockers-audit.md](bugs/2026-09-24-save-blockers-audit.md) — **แก้แล้ว 2026-09-24**: ไล่จุด "บันทึกไม่ได้" ทั้งรอบ — รหัสไทยที่มีสระ/วรรณยุกต์, NUL → 500, คู่ค้า version ชนในใบร่างเก่า, นับความยาวเป็น byte, เลขภาษีมีขีด/สาขา 0, `wht_rate` หายเป็น 0%, ล้างแถวภาษีสุดท้ายไม่ได้, ช่องยอด Tab แล้วว่าง, วาง Excel, บัญชีแม่วนลูป, วันหมดอายุการเข้าใช้งานไม่มีผล (ใช้ได้ถึงสิ้นวัน ตรวจทุกทางรวม GL), เพิ่มผู้ใช้ทับข้อมูลเงียบ, master รหัสซ้ำ upsert ทับ, คอลัมน์ `tax_filings` แคบ; รอบ E: ใบลดหนี้ไม่ติดกรอบ 6 เดือน, ปี พ.ศ. ในงวดภาษี, กฎใหม่ตรวจเฉพาะแถวใหม่, หลักตรวจสอบเลขผู้เสียภาษี, สมุดที่ mapping อ้างถึง, กลับรายการภาษีหักเดือนหลัง, 50 ทวิ ใช้ยอดที่บันทึก
- [bugs/2026-09-24-mcp-token-exceeds-issuer-scope.md](bugs/2026-09-24-mcp-token-exceeds-issuer-scope.md) — **แก้แล้ว 2026-09-24**: ADMIN ที่จำกัดบริษัทออก API/MCP token ให้บริษัทนอกขอบเขตได้ — ตอนนี้ตรวจตอนออก + ตัดบริษัทนอกขอบเขตปัจจุบันของผู้ออกทุกครั้งที่ใช้ + GL ตรวจขอบเขตผู้ออกทุกคำขอ token
- [bugs/2026-09-25-passenger-car-tax-cap-never-applied.md](bugs/2026-09-25-passenger-car-tax-cap-never-applied.md) — **แก้แล้ว 2026-09-25**: เพดานค่าเสื่อมทางภาษีรถยนต์นั่ง 1 ล้านบาทไม่เคยทำงาน (backend เช็ครหัส `PASSENGER_CAR` แต่ฟอร์มบันทึก `VEHICLE_PASSENGER`) — เปลี่ยนเป็น flag `passengercartaxcap` รายสินทรัพย์ + คิดตามสัดส่วนระยะเวลา + เพดานอัตราร้อยละ 20 ต่อปีตามวันที่ถือรถจริง (พ.ร.ฎ. 145 ม.4(5), ม.5, ทะเบียน 21 §13)
- [bugs/2026-09-24-settings-screen-clears-company-session.md](bugs/2026-09-24-settings-screen-clears-company-session.md) — **แก้แล้ว 2026-09-24** เปิดจอตั้งค่าบัญชีเข้าระบบแล้ว session เหลือแต่ Holding (ทุกจอบัญชีขึ้น "กรุณาเลือกบริษัท") + error GL เป็นข้อความกลางเมื่อ browser เป็น en-US → `GET /session/selection` + ส่งภาษาแอป/ใช้ `message_th`
- [bugs/2026-09-24-fixed-asset-gl-guessed-accounts.md](bugs/2026-09-24-fixed-asset-gl-guessed-accounts.md) — **แก้แล้ว 2026-09-24**: ผ่านค่าเสื่อม/จำหน่ายสินทรัพย์เข้า GL — เลิกเดารหัสบัญชี (`520103`/`129101` ฯลฯ) ใช้สินทรัพย์ → ประเภทสินทรัพย์ → ถามผู้ใช้ (`fa_account_required` + field), เลขใบสำคัญขึ้นต้นด้วยสมุดประเภททั่วไปที่เลือก + ตรวจ 30 ตัวอักษร (rune), ตรวจสาขาหัวเอกสารแบบเดียวกับหน้าบันทึกรายวัน, เงินเป็น decimal ทั้งเส้น
- [bugs/2026-09-08-permissiongroup-me-400-missing-backend-url.md](bugs/2026-09-08-permissiongroup-me-400-missing-backend-url.md) — **แก้แล้ว (รวมเข้า dev 2026-09-25)**: `GET /api/system-settings/permissiongroup/me` ตอบ 400 เมื่อผู้เรียกไม่ส่ง `x-bc-backend-url` — BFF ตรวจ URL ฝั่ง client เฉพาะเมื่อส่งมา (proxy ใช้ URL ฝั่ง server อยู่แล้ว)
- [bugs/2026-09-25-line-code-route-unauthenticated.md](bugs/2026-09-25-line-code-route-unauthenticated.md) — **แก้แล้ว 2026-09-25**: `POST /api/auth/line/code` ออก LINE bridge code ให้ใครก็ได้โดยไม่ตรวจ session — BFF ตรวจรูปแบบ header แล้วถาม mainapi `/verify-token` ก่อนเรียก bridge (token ปลอม/หมดอายุ → 401 ไม่ยิง bridge)
- [bugs/2026-09-25-fa-edit-does-not-recalculate-schedule.md](bugs/2026-09-25-fa-edit-does-not-recalculate-schedule.md) — **แก้แล้ว 2026-09-25**: แก้อัตราค่าเสื่อม/% ปีแรก/ค่าเสื่อมสะสมยกมาแล้วตารางค่าเสื่อมไม่คำนวณใหม่ — ตอนนี้คำนวณใหม่เมื่อค่าที่ใช้คำนวณเปลี่ยน และปฏิเสธ (409 `fa_schedule_posted`) ทั้งการแก้และการสั่งคำนวณใหม่เมื่อมีงวดผ่านรายการ GL แล้ว
- [bugs/2026-09-25-fa-post-gl-ce-year-as-fiscal-year-code.md](bugs/2026-09-25-fa-post-gl-ce-year-as-fiscal-year-code.md) — **แก้แล้ว 2026-09-25**: ผ่านค่าเสื่อม/จำหน่ายสินทรัพย์เข้า GL ส่งปี ค.ศ. เป็นรหัสปีบัญชี บริษัทที่ใช้รหัส พ.ศ. จึงได้ "กรุณาตั้งค่าปีบัญชี" — ตอนนี้หาปีบัญชีจากวันที่ของใบสำคัญแบบ Champ (ไม่พบ/ปิด/ทับซ้อน = field error ไทย) ไม่ระบุวันที่ = วันสิ้นงวด และวันที่ใบต้องอยู่ปีบัญชีเดียวกับงวด (ปีของงวดรับ ค.ศ. เท่านั้น)

- [ชื่อเมนูและสถานะรอพัฒนา 2026-09-09](bugs/2026-09-09-menu-labels-and-pending-screens.md) — ชื่อไทยไม่ถูกแคชทับ, ชื่อหน้าจอตรงกัน และป้ายสำหรับ 177 เมนูที่ยังไม่มีหน้าจอ
- [บทเรียน code review 2026-09-14](bugs/2026-09-14-code-review-gl-warehouse-fixes.md) — confirm() เป็น Promise, PUT location ต้อง spread doc เดิม, GL consumer group คงที่, ห้าม panic ตอน register consumer, เพดานบรรทัด journal ปิดงบ, report วนหน้า, NumericInput ไม่ปัดค่า

## Snippets (รูปแบบการเขียนดูที่ `snippets/README.md`)

- [snippets/champ-menu-source-audit.md](snippets/champ-menu-source-audit.md) — บัญชีรายการต้นทางจาก `menuconfig.xml` ของ Champ สำหรับตรวจผังเมนู (ตรวจ 2026-09-11)
- [snippets/champ-upgrade-menu-map.md](snippets/champ-upgrade-menu-map.md) — ผังเมนู BC สำหรับผู้ใช้อัปเกรดจาก Champ 223 รายการ (ตรวจ 2026-09-11 — ผังปัจจุบันยึด ADR 2026-09-19-champ-parity-menu-cut)

## ที่อื่นใน `docs/`

- `docs/handoff/` — สถานะงานค้างระหว่าง session: `HANDOFF-2026-09-19.md` (รอบ Champ-parity), `HANDOFF-2026-09-20-GL-GLM.md`, `HANDOFF-2026-09-20-GL-TAX.md` (GL + ภาษี); handoff เก่ากว่านี้ดูผ่าน `git show 1660b335:docs/handoff/<file>`
- `docs/audits/` — ผลตรวจ GL 2026-09-20 (`GL-AUDIT-2026-09-20.md`, `GL-FIX-2026-09-20.md`)
- `docs/user-guide/gl-journals.md` — คู่มือใช้งานสมุดรายวัน GL
- `docs/examples/` — ชุดข้อมูล/manifest ตัวอย่าง GL และภาษี (อ้างจาก [22-sample-data-gl-tax-2569.md](22-sample-data-gl-tax-2569.md))
- `docs/evidence/` — หลักฐาน deploy/UAT รอบ 2026-09-11 (บันทึกประวัติ)
- `docs/runbooks/RECOVERY-READINESS.md` — runbook กู้คืน prod (รอข้อมูล backup/RPO/RTO จากลุงจืด)
- `docs/reference/CODE-MAP.md` — ดัชนีไฟล์ใหญ่ที่สร้างอัตโนมัติด้วย `tools/gen-code-map.ps1`
- `.agents/skills/` — skill ส่วนตัวของลุงจืด (`ui-scale-polish`, `datacrud`, `thai-accounting-tax`)

## README ที่ยังอยู่ข้างโค้ด (เอกสารเฉพาะ package — อ่านคู่กับบทความด้านบน)

- `backend/README.md`, `backend/internal/goapi/mydb/README.md`, `backend/internal/goapi/mypostgres/README.md`, `backend/internal/goapi/logger/README.md`, `tools/rdform/README.md`, `backend/prompts/language_requests/README.md`, `frontend/AGENTS.md` (สร้างโดย Next.js เอง)

## คำถามค้าง

แต่ละบทความมีหัวข้อ "ช่องว่าง / สิ่งที่ยังไม่ตรวจ" ท้ายบทความ; ข้อที่ต้องการคำตอบจากลุงจืดโดยตรงสรุปไว้ใน [18-decisions-and-agreements.md](18-decisions-and-agreements.md) §คำถามที่ยังไม่มีคำตอบ
