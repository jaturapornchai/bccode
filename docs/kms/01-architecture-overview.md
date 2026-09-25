# ภาพรวมสถาปัตยกรรมระบบ BC Ai Account
> ตรวจล่าสุด: 2026-09-25 (commit c58f0b62) — PostgreSQL ตัวเดียว: MongoDB, Kafka, Redis, ClickHouse ถูกถอดออกจากระบบเมื่อ 2026-09-23 ([ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — ห้ามเพิ่มกลับ

## 1. สรุปสั้น
- ระบบเป็น monorepo 2 ส่วนหลัก: Go backend (module `smlcloudplatform`, binary เดียวจาก `backend/main.go`, 192 บรรทัด) และ Next.js frontend (`frontend/`); ทั้งคู่ deploy เป็น Docker container ตาม `deploy/account/compose.yml`
- binary backend มีบทบาทเดียว ไม่มีโหมดสลับ: `func main()` (`backend/main.go:76-182`) เชื่อม PostgreSQL ฐานควบคุมกลาง → ตั้ง cacher → ลงทะเบียน HTTP module → mount goapi ใต้ `/goapi` แล้ว `ms.Start()`; env `DEV_API_MODE=2` ที่ยังตั้งอยู่ใน `backend/Dockerfile:42`, `backend/docker-compose.yml:39` และ `backend.env` ของ prod (`deploy/account/provision-server.sh:85`) ถูกอ่านเข้า config (`backend/internal/setupconfig/loader.go:34`) แต่ไม่มีโค้ดใดใช้ค่าแล้ว
- goapi ไม่ใช่ service แยกใน deploy — เป็น sub-server ที่สร้างด้วย `goapi.New()` แล้ว mount กลุ่ม `/goapi` บน Echo instance เดียวกับ mainapi (`backend/main.go:168-179`); ไม่มี `cmd/goapi`/`Dockerfile.goapi` standalone แล้ว (ลบ 2026-09-23)
- browser ไม่คุย mainapi ตรง: ทุก request ผ่าน Next.js (route handler ใต้ `frontend/src/app/api/**`, 34 ไฟล์ `route.ts`, หรือ rewrite `/backend/*`) แล้ว Next.js ค่อย fetch ไป `BCAI_LOCAL_BACKEND_URL` ฝั่ง server (`frontend/next.config.ts:19,71`)

## 2. แผนภาพ (ASCII)
```
 Browser (Thai 40+ users)
   |  HTTPS account.bcaicloud.com  (prod: Caddy -> 127.0.0.1:3200, deploy/account/Caddyfile.account;
   |                                Caddy ตอบ 404 ให้ /backend/goapi/api/health/{background,queue*,database,system})
   v
 Next.js frontend :3000 (prod publish 127.0.0.1:3200, deploy/account/compose.yml)
   |-- route handlers  frontend/src/app/api/**  (34 route.ts)  -- forward Authorization: Bearer
   `-- rewrite         /backend/:path*  ->  ${BCAI_LOCAL_BACKEND_URL}/:path*   (next.config.ts:71, บล็อก login/dangerous routes ก่อน :26-69)
   v
 mainapi :8888  (backend/main.go)  -- Echo via pkg/microservice, ตรวจสิทธิ์สดทุกคำขอ (AuthService.MWFuncMixShop, main.go:118,148)
   |-- โมดูล (main.go:153-166): /login /googlelogin /dev-login /demo-login /refresh /logout /holding/* /shop/* /profile/*
   |                            /organization/* /fa/v2/* /gl/v2/* /mcp/gl /integration/gl/v2/* /mcp-tokens /media/upload/*
   |-- /goapi/*  <- goapi sub-server (bootstrap.go RegisterRoutes :306) มี middleware + bearer auth ของตัวเอง (bootstrap.go:130-209)
   |-- /api/language/:lang (public alias, main.go:178)   /healthz (:149)   /metrics (Prometheus, :114)   /reload-config (:138-146)
   `-- background: purge session หมดอายุทุก 10 นาที (main.go:99-112), goapi health checker + cleanup prepared statement ทุก 30 นาที (bootstrap.go:103-110)
   |
   |--> PostgreSQL: ฐานควบคุมกลาง `bcai_projection` (identity/holding/company/branch/member/สิทธิ์/session/MCP token) + ฐานแยกต่อกลุ่มกิจการ (GL, สินทรัพย์ถาวร, แบบยื่นภาษี) — ดู `02-data-stores.md`
   `--> MinIO/S3 (รูปภาพ/ไฟล์แนบ + thumbnail คู่เสมอ)

 (ไม่มี worker/migrate service แยก — ทุกอย่างรันใน mainapi process เดียว)
```

## 3. โมดูล HTTP ที่ลงทะเบียนจริงใน `main.go` (`backend/main.go:153-166`)
| โมดูล | หน้าที่ | อ้างอิง |
|---|---|---|
| `authentication.NewAuthenticationHttp` | login/googlelogin/dev-login/demo-login, refresh, logout, LINE link (`/profile/link-line*`) | `backend/internal/authentication/authentication_http.go:127-154` |
| `shop.NewShopHttp` / `NewShopMemberHttp` | holding/shop selection, สมาชิก | `backend/internal/shop/` |
| `employee.NewEmployeeHttp` | ทะเบียนพนักงาน (`/holding/employee*`, `/shop/employee*`) | `backend/internal/shop/employee/employee_http.go:36` |
| `fahttp.NewHttp` (fixedasset) | สินทรัพย์ถาวร + ตารางค่าเสื่อม (`/fa/v2/*`) | `backend/internal/fixedasset/httpapi/http.go:61-62` |
| `glhttp.NewHttp` (generalledger) | `/gl/v2/*` command/report/resource + `/mcp/gl` (MCP Streamable HTTP ผ่าน `internal/mcpgateway`) + `/integration/gl/v2/*` (API token) | `backend/internal/generalledger/httpapi/http.go:67-73`, `mcp.go:20-24`, `token_api.go:14-20` |
| `mcptoken.NewHttp` | ออก/จัดการ MCP API token ต่อ Holding (`/mcp-tokens`) | `backend/internal/mcptoken/http.go:34` |
| `company.NewCompanyHttp` / `branch.NewBranchHttp` / `businesstype.NewBusinessTypeHttp` | บริษัท/สาขา/ประเภทธุรกิจ | `backend/internal/organization/{company,branch,businesstype}/` |
| `rolepermission.NewRolePermissionHttp` | ชุดสิทธิ์ | `backend/internal/organization/rolepermission/` |
| `media.InitMediaUploadHttp` | อัปโหลดไฟล์ทั่วไป (`/media/upload/*`) | `backend/internal/media/media_http.go:37-38` |
| `goapi.New()` | ดูตาราง §5 | `backend/internal/goapi/` |

แพ็กเกจเก่าที่ยังอยู่ใน `backend/internal/` แต่ **ไม่มีจุดใดเรียกใช้แล้ว** (ตรวจด้วย `grep -rl "smlcloudplatform/internal/<pkg>" backend --include=*.go` ไม่พบผู้เรียกนอกแพ็กเกจตัวเอง): `internal/vfgl`, `internal/transaction`, `internal/stockprocess` — รอการลบหรือ decision จากลุงจืด ไม่ใช่สถาปัตยกรรมที่ใช้งานจริง

## 4. binary / Dockerfile / compose ที่ใช้จริง
| artifact | build จาก | หมายเหตุ |
|---|---|---|
| `backend/Dockerfile` | `go build -ldflags="-w -s" main.go` (multi-stage `golang:1.26-alpine` → `alpine:3.21`), `CGO_ENABLED=0` | ใช้ทั้ง local (`backend/docker-compose.yml`) และ prod (`deploy/account/compose.yml`, image `${MAINAPI_IMAGE}`) |
| `backend/Dockerfile.local` | เหมือนกัน ใช้กับ overlay dev (`backend/docker-compose.local.yml`) | compose local mount `bootstrap.local.json` + `custom_config.local.json` |
| `backend/cmd/glseed` | เครื่องมือ CLI แยก (ไม่ใช่ service) เติมข้อมูลหลักฐานลูกหนี้/เจ้าหนี้/ธนาคาร/งบประมาณ/ใบสำคัญของ GL ผ่าน service layer เดียวกับ backend (`store.Execute`), ค่าเริ่มต้น dry-run | `backend/cmd/glseed/main.go:1-11` |

`backend/cmd/*` microservice ยุคเก่า (29 โฟลเดอร์), `backend/cluster/`, Dockerfile/Makefile ยุคเก่า และ swagger เก่า ถูกลบเมื่อ 2026-09-23 (commit `550d5489`, ADR ข้อ 6) — ปัจจุบัน `backend/cmd/` มีแค่ `glseed/`; ที่ยังค้างคือ `backend/.github/workflows/*.yaml` (ไม่ถูกรันเพราะไม่ได้อยู่ที่ root — ห้ามย้ายขึ้น root ตามกฎ GitHub ใน `AGENTS.md`)

Stack prod (`deploy/account/compose.yml`): 5 service — `postgres` (18-alpine), `minio`, `minio-init` (สร้าง bucket/policy แล้วจบ), `mainapi` (publish `127.0.0.1:8888`), `frontend` (publish `127.0.0.1:3200`); `deploy/account/compose.8gb.yml` ลด `mem_limit` ให้พอดีเครื่อง 8 GB; edge = Caddy บน host proxy ไป `:3200` (`deploy/account/Caddyfile.account`)

## 5. goapi sub-server (`backend/internal/goapi/bootstrap.go`, 517 บรรทัด)
`Init()` (`:45-127`): โหลด bootstrap config ของ goapi เอง → `handlers.InitR2Client()` (S3/MinIO, `:52`) → `mydb.InitManagerPool` (PostgreSQL เท่านั้น, `:65`) → `mypostgres.InitQueueSchema` บนฐาน global `postgres` (`:94`) → goroutine `StartConnectionHealthChecker` (`:103`) + ticker เคลียร์ prepared statement ทุก 30 นาที (`:105-108`) → ตั้ง stock-engine callback (`stockengine.AfterRecalculate`, `:116`) และ `stockengine.StartWorkers` (`:121-123`, ปิดได้ด้วย `BCAI_STOCK_WORKER=0|false`)

`RegisterRoutes(g, prefix, cacher, authorizationFinders...)` (`:306-404`): route สาธารณะ `/goapi/`, `/goapi/version`, `/goapi/api/health*`, `/goapi/api/language/:lang`, `/goapi/api/address/thailand`; ที่เหลืออยู่ใต้ `authGroup` (`:308`, bearer token ตรวจกับ PostgreSQL ผ่าน `createGoAPIAuthMiddleware`, `:169-209`) ได้แก่ inventory costing (`inventory.RegisterRoutes`, `:337-339`), stock-cost (`:341-345`), transaction-calculator (`:347-351`), รายงานขาย/ภาษี/หนี้ + แบบยื่นภาษีจาก GL (`/api/report/*`, `:353-370`), product search/alias/cache (`:372-382`), stock-report (`:384-386`), ไฟล์/รูป/วิดีโอ (`:388-397`)

**หมายเหตุสำคัญ**: route กลุ่มสินค้า/บาร์โค้ด/สต๊อก (inventory `:337-339`, stock-cost `:341-345`, transaction-calculator `:347-351`, product `:372-382`, stock-report `:384-386`) ยังลงทะเบียนอยู่ในโค้ด แต่ไม่มีจุดใดใน backend เขียน `INSERT INTO product`/`productbarcode` แล้ว และฝั่ง frontend ทำเครื่องหมายจอเหล่านี้เป็น "รอพัฒนา" ผ่าน `isMenuBackendRetired` (`frontend/src/lib/menu-screen-status.ts:40-55` — `RETIRED_BACKEND_ROUTES`, `isErpTransactionRoute`, `isOperationsRoute` และจอตั้งค่าที่ basePath ไม่อยู่ใน `POSTGRES_SETTING_BASE_PATHS`) — ถือเป็นโค้ดค้างที่ยังไม่ได้ลบ ไม่ใช่เส้นทางข้อมูลที่ใช้งานจริง; ส่วน `/api/report/tax/*` (`:353-370`) ยังใช้งานจริงกับเมนูภาษีในหมวด GL

## 6. pkg/microservice (framework กลาง)
- `NewMicroservice(cfg)` (`microservice.go:66`) สร้าง Echo + logger; `SetCacher` (`:322`) ผูก session store เข้ากับ `*Cacher` ที่เก็บใน PostgreSQL (`backend/pkg/microservice/cacher.go`)
- `AuthService.MWFuncMixShop` (`auth.go:160`) ตรวจ token จาก cacher แล้วยืนยันสิทธิ์สดจากฐานควบคุมกลางทุกคำขอผ่าน `live_authorization.go` (ผู้ใช้/สมาชิกภาพ/บริษัท/สาขาถูกปิด → token เดิมใช้ไม่ได้ทันที)
- `RegisterBackgroundWorker(fn)` (`microservice.go:160`) เก็บ goroutine ไว้รันตอน `Start()`; ผู้ใช้ปัจจุบันในโค้ด mainapi มีตัวเดียวคือ purge session หมดอายุ (`main.go:99-112`)
- persister ที่เหลือ: `persister.go` (GORM บน PostgreSQL, `gorm.io/driver/postgres`) และ `persister_file.go`/`persister_file_r2.go`/`persister_image.go` (ไฟล์/รูป) — persister ของฐานที่ถูกถอดไม่มีแล้ว

## 7. เส้นทาง request: browser → Next.js → mainapi
1. browser เรียก `/api/<domain>/...` ของ Next.js เอง; route handler ดึง `Authorization` header แล้ว fetch ไป mainapi ด้วย `BCAI_LOCAL_BACKEND_URL`
2. เส้นทางที่สอง: rewrite `/backend/:path*` → `${BCAI_LOCAL_BACKEND_URL}/:path*` (`frontend/next.config.ts:71`) โดยบล็อกเส้นทาง login (`/backend/login`, `/backend/googlelogin`, `/backend/dev-login`, `/backend/demo-login` ฯลฯ, `:26-38`) และเส้นทางอันตราย (`/backend/goapi/get|exec|getdoc`, `/backend/reportm/*`, `/backend/goapi/api/setup|mcp/*`, `/backend/goapi/mcp/*`, `/backend/reload-config`, `/backend/profile/link-line*`, `:51-64`) ไปที่ `/_blocked-auth-route` ก่อน (`next.config.ts:66-69`)
3. ที่ mainapi: `publicPath` ไม่ต้อง token (`/mcp/gl`, `/integration/gl/v2/*`, login family, `/healthz`, `/metrics`, `/reload-config`, `/goapi/*`, `/api/language/*` — `main.go:119-135`); `exceptShopPaths` ต้องมี token แต่ยังไม่ต้องเลือก holding/shop (`main.go:47-73`)
4. `BCAI_LOCAL_BACKEND_URL` เป็น build ARG/ENV ของ image frontend (`frontend/Dockerfile:26-28`), image ฟัง `:3000`

## 8. Background workers ที่ยังทำงานจริง
| ตัวประมวลผล | รันที่ไหน | กลไก | อ้างอิง |
|---|---|---|---|
| Purge session หมดอายุ | mainapi | ticker ทุก 10 นาที เรียก `cacher.PurgeExpired(ctx)` | `backend/main.go:99-112` |
| goapi connection health checker | mainapi (goapi sub-server) | goroutine ตรวจการเชื่อมต่อ | `backend/internal/goapi/bootstrap.go:103` (`handlers.StartConnectionHealthChecker`) |
| เคลียร์ prepared statement | mainapi (goapi) | ticker ทุก 30 นาที | `backend/internal/goapi/bootstrap.go:105-108` (`myPg.CleanupPreparedStatements`) |
| Stock engine (ต้นทุนสินค้า) | mainapi (goapi) | `StartWorkers` แค่เปิดสวิตช์; worker ต่อกลุ่มกิจการเริ่มผ่าน `EnsureWorker` ซึ่งไม่มีผู้เรียกนอกแพ็กเกจ และผู้ตั้งงาน (`MarkDocumentDirty` ใน `mypg/doc.go`) ก็ไม่มีผู้เรียก — ในทางปฏิบัติไม่มีงานประมวลผล | `backend/internal/goapi/bootstrap.go:116-123`, `process/stockengine/manager.go:36,78` |

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- `internal/vfgl`, `internal/transaction`, `internal/stockprocess` ไม่มีผู้เรียกแล้ว — ยังไม่ได้ถามลุงจืดว่าจะลบทิ้งเมื่อไร
- route สินค้า/สต๊อก/inventory ใน goapi (§5) และ `stockengine` (ใช้ตาราง `stock_dirty` ที่ไม่มีโค้ด production สร้าง — มีแค่ใน `process/stockengine/store_integration_test.go:95`) ยังไม่ได้ตัดสินว่าจะลบหรือใช้ต่อเมื่อสร้าง backend สต๊อกบน PostgreSQL (ตามลำดับความสำคัญที่ลุงจืดตั้งไว้ว่า "เน้น GL ก่อน")
- ยังไม่ได้ไล่ทุก route ของ `/gl/v2/*` และ goapi ทีละเส้นว่ามี frontend เรียกจริงกี่เส้น — ดู `07-goapi-layer.md` และ `09-frontend.md` (นอกขอบเขตบทความนี้)
