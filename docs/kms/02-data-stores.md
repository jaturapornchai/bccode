# 02 — ที่เก็บข้อมูล (Data stores) และบทบาทจริงของแต่ละตัว
> ตรวจล่าสุด: 2026-09-25 (commit c58f0b62) — MongoDB, Kafka, Redis, ClickHouse ถูกถอดออกจากระบบเมื่อ 2026-09-23 ([ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) ห้ามเพิ่มกลับ; ระบบเก็บข้อมูลจริงเหลือ 2 อย่างเท่านั้น: **PostgreSQL** และ **MinIO/S3** (ไฟล์)

## 1. ภาพรวม
| store | บทบาทจริง | สถานะ |
|---|---|---|
| PostgreSQL — ฐานควบคุมกลาง `bcai_projection` | identity, holding/company/branch, สมาชิกและสิทธิ์, พนักงาน, audit, **session/token cache**, MCP token | LIVE |
| PostgreSQL — ฐานแยกต่อกลุ่มกิจการ (`dbname = <holdingcode>`) | บัญชีแยกประเภท (GL), แบบยื่นภาษี และสินทรัพย์ถาวรของ Holding นั้น | LIVE (สร้าง database อัตโนมัติเมื่อเชื่อมครั้งแรก, ตารางสร้างแบบ lazy ต่อโมดูล — ดู §3) |
| PostgreSQL — ฐาน global `postgres` | ตาราง `queues`/`deadletterqueue` ของ goapi | สร้างตอน boot แต่ไม่มีผู้ใช้ (ดู §3.4) |
| MinIO / S3 | ไฟล์รูป/วิดีโอ/ไฟล์แนบ + thumbnail คู่เสมอ; PostgreSQL เก็บแค่ object key | LIVE |

## 2. PostgreSQL — ฐานควบคุมกลาง `bcai_projection`
- ชื่อฐานตายตัว: `const ControlDatabase = "bcai_projection"` (`backend/internal/mcptoken/token.go:26-27`); เปิดผ่าน `mypg.PgSqlFastConnect(mcptoken.ControlDatabase)` ครั้งเดียวใน `main()` แล้วส่งต่อ (`backend/main.go:90-93`) ให้ทั้ง cacher, `AuthService`, live-authorization และ goapi (`backend/main.go:94-118,174`)
- schema สร้าง/ขยายด้วยโค้ด Go ตรง ๆ (ไม่ใช้ SQL migration tool แยก) ที่ `backend/internal/centraldb/centraldb.go` — `centraldb.Open()` (`:26-38`) เรียก `EnsureSchema` ครั้งเดียวต่อ connection pool ตอนถูกเรียกใช้ครั้งแรก (ล็อก `pg_advisory_lock(hashtextextended('bcai_central_schema', 0))`, `:57-74`)
- ตารางหลักในฐานนี้ (`centraldb.go:76-248`, ตัวแปร `schemaStatements`): `users`, `user_identities`, `holdings`, `companies`, `branches`, `holding_members`, `role_permissions`, `business_types`, `employees`, `organization_audits` (โค้ดมีแต่ `INSERT` — `backend/internal/organization/audit.go:57`, `backend/internal/shop/shopuser_postgres_repository.go:466`), `shop_user_access_logs`
- **`cache_entries`** (session/token) สร้างแยกโดย `backend/pkg/microservice/cacher.go:45-53` บนฐานเดียวกัน — คีย์ `(cache_key, field)` + `value` + `expires_at` (bearer/xapikey/refresh/session token); worker ใน `main.go:99-112` ลบแถวหมดอายุทุก 10 นาทีผ่าน `cacher.PurgeExpired`
- **ตรวจสิทธิ์สดทุกคำขอ**: `pkg/microservice/live_authorization.go` query ตรง `users.is_active` และ membership/company/branch จากฐานนี้ทุกครั้งที่มี request (timeout 2 วินาที, `liveAuthorizationTimeout`, `:30`) — ผู้ใช้/บริษัท/สาขาถูกปิดแล้ว token เดิมใช้ไม่ได้ทันที
- MCP: `mcptoken` เก็บ API token ของ Holding ในตาราง `mcp_access_tokens`, `mcp_token_companies`, `mcp_token_audit` (`backend/internal/mcptoken/schema.sql`, go:embed ที่ `http.go:22`) ในฐานเดียวกันนี้ (`token.go:136,236`) ให้ `/mcp/gl` และ `/integration/gl/v2/*` ตรวจสิทธิ์แยกจาก session ปกติ

## 3. PostgreSQL — ฐานแยกต่อกลุ่มกิจการ (per-holding)
### 3.1 การสร้างฐานข้อมูล
- ทุกจุดที่ต้องอ่าน/เขียนข้อมูลของ Holding หนึ่ง ๆ เรียกผ่าน `mypg.PgSqlFastConnect(holdingCode)` (`backend/internal/goapi/mypg/fast_utils.go:38-41` → `PgSqlFastConnectV2` ใน `mypg/compatibility.go:11` → `mydb.GetGlobalConnectionFromPool`)
- ถ้าฐานชื่อ `<holdingcode>` ยังไม่มี, `DatabaseManager.connectPostgreSQL` จะจับ error `does not exist`/`3D000` แล้วเรียก `createPostgreSQLDatabase` สร้างให้อัตโนมัติผ่าน connection ไปฐาน `postgres` ก่อน แล้วต่อใหม่ (`backend/internal/goapi/mydb/unified_database_manager.go:74-107,159-166`) — **ไม่มี allowlist ชื่อ Holding ในขั้นตอนนี้** เพราะฟังก์ชันฝั่ง HTTP ตรวจรูปแบบเองอีกชั้น (เช่น GL: `validHoldingRegex = ^[A-Za-z0-9_]{1,63}$`, `backend/internal/generalledger/httpapi/http.go:29,54`)
- ไม่มีการ provision ตารางล่วงหน้าเป็นชุด — แต่ละโมดูลสร้างตารางของตัวเองแบบ lazy (`CREATE TABLE IF NOT EXISTS`) ตอนถูกเรียกใช้ครั้งแรก และจำว่าเสร็จแล้วต่อ connection pool

### 3.2 ตารางที่โมดูลปัจจุบันสร้างจริง
| โมดูล | ตาราง | สร้างโดย | อ้างอิง |
|---|---|---|---|
| GL (บัญชีแยกประเภท) | `gl_records`, `gl_projection_state`, `gl_events`, `gl_lines`, `gl_journal_review_events`, `gl_source_journals` | `generalledger.EnsureSchema` รัน `schema.sql` + `subledger.sql` + `budget.sql` (go:embed) ใน transaction เดียว ล็อกด้วย `pg_advisory_xact_lock(hashtext('bc_general_ledger_schema'))` | `backend/internal/generalledger/postgres.go:16,52-80`, `schema.sql` |
| GL — บัญชีย่อยลูกหนี้/เจ้าหนี้/ธนาคาร | `gl_subledger_partners`, `gl_subledger_bank_accounts`, `gl_subledger_documents`, `gl_subledger_allocations`, `gl_subledger_settlements`, `gl_subledger_bank_lines`, `gl_subledger_statements`, `gl_subledger_matches`, `gl_subledger_audit` | `EnsureSchema` เดียวกัน | `backend/internal/generalledger/subledger.go:17`, `subledger.sql` |
| GL — งบประมาณ | `gl_budgets`, `gl_budget_lines` | `EnsureSchema` เดียวกัน | `backend/internal/generalledger/budgets.go:24`, `budget.sql` |
| แบบยื่นภาษี (goapi `/api/report/tax/form/*`) | `tax_filings`, `tax_filing_history` (เอกสารเป็น JSONB, ยอดเงินเป็นสตริงทศนิยม) | `ensureTaxFilingSchema` ครั้งเดียวต่อ pool (`sync.Map`) | `backend/internal/goapi/handlers/tax_form.go:454-523` |
| สินทรัพย์ถาวร (Fixed Asset) | `fa_records` (เก็บทุกชนิด: `assets/types/depreciations/disposals` แยกด้วยคอลัมน์ `kind`, payload เป็น JSONB — คล้าย `gl_records`) | `CREATE TABLE IF NOT EXISTS` ตอนเรียกใช้ครั้งแรก (`sync.Map` กันสร้างซ้ำต่อ holding) | `backend/internal/fixedasset/records.go:23-59` |

ทุกโมดูลข้างบนเชื่อมฐานเดียวกันของ Holding นั้น (`mypg.PgSqlFastConnect`) — ไม่มีการแยกฐานข้อมูลรายโมดูล

### 3.3 ตารางโดเมนสินค้า/สต๊อก/เอกสารซื้อขาย
- ไม่มีโค้ดสร้างหรือเขียนตารางสินค้า/บาร์โค้ด/คลัง/ลูกหนี้-เจ้าหนี้ master/เอกสารซื้อขายต่อ holding แล้ว (ไม่มี `INSERT INTO product`/`productbarcode` ใน backend) — ถอดพร้อม ADR 2026-09-23; จอเหล่านี้ขึ้น "รอพัฒนา" ที่ frontend (`frontend/src/lib/menu-screen-status.ts`, ดู `01-architecture-overview.md` §5)
- โค้ดค้างที่ยังอยู่: route goapi กลุ่ม inventory/stock-cost/product/stock-report (`backend/internal/goapi/bootstrap.go:337-386`, ยกเว้น `/api/report/*` `:353-370` ที่ยังใช้จริง) — route `inventory/*` สร้างตาราง `productcostingconfig`/`inventory*` เองเมื่อถูกเรียก (`backend/internal/goapi/inventory/handler.go:288` → `database.go:12`); รอลบหรือ rebuild บน PostgreSQL ล้วน

### 3.4 ฐาน global `postgres` — คิวงาน goapi
- `mypostgres.InitQueueSchema(globalDB)` สร้างตาราง `queues`/`deadletterqueue` ตอน boot (`backend/internal/goapi/bootstrap.go:94`, `backend/internal/goapi/mypostgres/init_schema.go`)
- ตรวจแล้ว 2026-09-25: `mypostgres.QueueManager` (`mypostgres/queue.go`) ไม่มีผู้เรียกนอกแพ็กเกจ — ไม่มีโค้ดใด enqueue/dequeue; health endpoint `/goapi/api/health/queue*` คืนข้อความคงที่ (`backend/internal/goapi/handlers/health.go:36-58`) — ตารางนี้เป็นโค้ดค้าง

## 4. MinIO / S3
- env key: `S3_ENDPOINT`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_BUCKET_NAME`, `S3_REGION`, `S3_FORCE_PATH_STYLE`, `S3_ACCOUNT_ID` — ถ้าไม่ตั้งจะ fallback ไปตัวแปรชื่อเดิม `R2_*` (`backend/internal/goapi/handlers/image_r2.go:67-105`)
- เส้นทางอัปโหลด: frontend `POST /api/upload/image` (`frontend/src/app/api/upload/image/route.ts` → `frontend/src/lib/image-upload-proxy.ts:128-129`) → `${mainApiUrl}/goapi/image/upload` → `ImageUploadHandler` (`image_r2.go:192`) ตรวจ holdingcode จาก token, ตั้ง object key, สร้าง WebP thumbnail แล้ว `PutObject` 2 object (ตัวเต็ม + `<key>.thumb.webp`) (`image_thumbnail.go:25-30`)
- อ่านรูป: `GET /goapi/s3/file/*` (`backend/internal/goapi/bootstrap.go:389` — อยู่ใน `authGroup` ต้องมี Bearer token)
- PostgreSQL เก็บแค่ object key/URI (เช่น `holding_members.avatar`/`avatar_thumb`, `employees.profile_picture`/`profile_picture_thumb` — `centraldb.go:175-176,216-217`) ไม่เก็บไฟล์จริงตามกฎ AGENTS.md
- compose: `minio/minio` + `minio-init` (mc สร้าง bucket, เปิด versioning, ปิด anonymous access, สร้าง policy `bcai-account-app` จาก `deploy/account/minio-app-policy.json`) ทั้ง local (`backend/docker-compose.local.yml`) และ prod (`deploy/account/compose.yml`)

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
1. route สินค้า/สต๊อก/inventory และตาราง `queues`/`deadletterqueue` (§3.3–3.4) จะถูกลบทิ้งหรือรอ rebuild บน PostgreSQL เมื่อไร ต้องถามลุงจืด
2. บทความนี้ตรวจจากซอร์สโค้ด local เท่านั้น ยังไม่ได้ตรวจ schema จริงบนฐาน PostgreSQL ของ production (`159.223.43.229`) ว่าตรงกับ `schema.sql`/`subledger.sql`/`budget.sql`/`centraldb.go` เวอร์ชันปัจจุบันครบทุกตารางหรือไม่
3. ไฟล์ SQL ที่ไม่มีโค้ด Go โหลดใช้: `backend/internal/database/{fresh_provision.sql,fresh_provision_all.sql,schema_full.sql}`, `backend/migrations/*.sql`, `backend/uat_seed.sql` — มีตาราง (เช่น `gl_accounts`, `gl_journals`, `user_sessions`) ที่ไม่ตรงกับ schema จริงข้างบน อย่าใช้เป็นแหล่งอ้างอิง schema
4. `internal/vfgl`, `internal/transaction`, `internal/stockprocess` (ดู `01-architecture-overview.md` §3) เป็นโค้ดกำพร้าที่อาจอ้างตารางเก่าเพิ่มเติม — ยังไม่ได้ไล่อ่านเพราะไม่มีผู้เรียกใช้แล้ว
