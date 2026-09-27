---
date: 2026-09-27
status: accepted
tags: [bc-account, security, nextjs, frontend, proxy]
supersedes: catch-all `/backend/:path*` + blocklist ใน `frontend/next.config.ts` (ขั้น 1 → allowlist 5 เส้น, commit `835cf27f`; ขั้น 2 → ไม่มี `/backend` เลย)
---

# ไม่มี proxy สาธารณะ `/backend/*` — browser เรียก mainapi ผ่าน BFF ใต้ `frontend/src/app/api/**` เท่านั้น

## Context
ทางเดียวจากอินเทอร์เน็ตถึง mainapi คือ Next (Caddy ส่งทุก path ไป frontend `127.0.0.1:3200`, mainapi ผูก `127.0.0.1:8888`). เดิม `rewrites()` เป็น catch-all `/backend/:path*` → mainapi + blocklist ซึ่งถูกเลี่ยงได้ด้วย alias `/v1` ของ mainapi และ URL-encoding และเปิด `/metrics`, `/healthz`, `/goapi/*` สู่สาธารณะ ([บั๊ก](../bugs/2026-09-27-backend-proxy-blocklist-bypass.md)). ขั้น 1 (deploy r20260927-3) ลดเหลือ allowlist 5 เส้นที่ browser ใช้จริง: ไฟล์ S3 `/backend/goapi/s3/file/*` และบริษัท/สาขา `/backend/organization/{company,branch}[/:id]`. ขั้น 2 (ลุงจืดสั่ง "ปิดให้หมด") ย้าย 5 เส้นนั้นเข้า BFF แล้วลบ `/backend` ทิ้ง.

## Decision (ขั้น 2 — 2026-09-27)
- `frontend/next.config.ts` ไม่มี `rewrites()`/`redirects()` (comment `:7-9`), ไม่มี middleware/proxy file → ทุก `/backend/*` = Next 404
- BFF บริษัท/สาขา `frontend/src/app/api/organization/[...orgPath]/route.ts`:
  - `GET /api/organization/{company|branch}` → mainapi `GET /organization/{resource}?management=true` (`:31-42`; query ของ client ไม่ถูกส่งต่อ)
  - `POST /api/organization/company` → `POST /organization/company` (`:44-60`; ตัด `backendUrl` ออกจาก body)
  - `PUT /api/organization/{company|branch}/:id` → `PUT /organization/{resource}/:id` (`:62-79`)
  - id = รหัสบริษัท/สาขาตามที่เก็บ ซึ่งเป็นข้อความอิสระ (backend `NormalizeCompanyCode` แค่ trim + ตัวพิมพ์ใหญ่): รับทุกอักขระยกเว้น `/`, backslash, อักขระควบคุม, `.`/`..` ยาวไม่เกิน 128 (`:27`) แล้วส่งต่อด้วย `goPathEscape` (`:87`) ที่ escape แบบเดียวกับ Go net/url (encodePath) — Echo v4.13.3 route ด้วย RawPath และไม่ unescape `:id` ส่วน Go ตั้ง RawPath ว่างก็ต่อเมื่อ path ตรงกับ escaping ของตัวเองทุกไบต์ จึงได้รหัสตรงตัว (ทดสอบกับ Echo จริง: `A&B`, `K+1(ก)!`, `a%2Fb`, `a b`); `encodeURIComponent` ใช้ไม่ได้เพราะ escape `& + , ; = @ : $` ที่ Go ไม่ escape (`A%26B` ค้างใน `:id`); path อื่น = 404 `not_found`, method ที่ไม่มี = 405 ของ Next, ต้องมี Bearer, ส่ง Authorization + Accept-Language, status/JSON ของ mainapi ผ่านตรง (`proxyMainApiJson`)
  - สร้างสาขาใช้ BFF เดิม `POST /api/workspace/branch` (`frontend/src/app/api/workspace/[...workspacePath]/route.ts:132` → `POST /organization/branch`, body `{ branch }`) — ไม่ทำเส้นซ้ำ
- BFF ไฟล์ `frontend/src/app/api/files/[...key]/route.ts`: `GET /api/files/<key>[?variant=thumbnail]` → `GET ${serverGoApiBase()}/s3/file/<key>` (`:44`) ด้วย Bearer ของผู้เรียก; query เดียวที่ส่งต่อคือ `variant=thumbnail` (อื่น = 400); encode ทีละ segment และปฏิเสธ segment ว่าง/`.`/`..`/มี `/` `\` อักขระควบคุม/ยาวเกิน 255 (`isSafeSegment`); ส่ง IP ผู้ใช้ต่อเป็น `X-Forwarded-For` (`forwardedClientIp` `:99` — ค่าแรกของ header ที่ Caddy ตั้ง ซึ่ง Caddy ทับค่าจาก client ที่ไม่อยู่ใน trusted_proxies; พอร์ต frontend ผูก 127.0.0.1) เพราะ rate limiter ของ `/goapi` นับตาม `ctx.RealIP()` (`backend/internal/goapi/bootstrap.go` `createTieredRateLimiter`, ค่า default 100/s burst 10) — ไม่ส่งจะนับรูปของทุกคนรวมที่ IP ของ container frontend; `redirect: "manual"` — 3xx จากต้นทาง = 502; สตรีม body พร้อม header ที่อนุญาต (`:15` content-type/length, cache-control, etag, last-modified, vary, x-content-type-options); ส่ง If-None-Match ต่อจึงได้ 304; ไม่ส่ง set-cookie
- จอ/ตัวแสดงรูป: `company-branch-tree-view.tsx:965,978,1250,1262,1275,1305` และ `currency-screen.tsx:342,349` เรียก BFF ข้างบน; `imageDisplayUrl` (`frontend/src/components/authenticated-image.tsx:101`) และ `resolveDisplayUrl` (`frontend/src/components/logo-avatar.tsx:81`) แปลง URI ที่เก็บ `/goapi/s3/file/<key>` → `/api/files/<key>` ด้วย `fileBffUrl` (`frontend/src/lib/image-upload-proxy.ts:231`); `imageNeedsAuthenticatedFetch` (`:237`) แนบ Bearer เฉพาะ URL origin เดียวกันที่ขึ้นต้น `/api/files/`; DB ยังเก็บ `/goapi/s3/file/<key>` เหมือนเดิม
- `auth.backendUrl`/header `x-bc-backend-url` (`<origin>/backend/goapi`, `frontend/src/lib/backend-url.ts:6-11`) เหลือเป็นตัวระบุที่ BFF ตรวจรูปแบบ (`getMainApiUrl`) แล้วทิ้ง — ปลายทางจริงคือ `serverMainApiBase()`/`serverGoApiBase()` (`:120-130`) จาก env เสมอ
- SSR preload ภาษาใช้ `serverGoApiBase()` ไม่ใช้ host จาก cookie (ขั้น 1, `frontend/src/lib/backend-language-preload.ts:25-51`)
- **กติกา:** browser ต้องการ endpoint ใหม่ของ mainapi/goapi → ทำ BFF ใต้ `frontend/src/app/api/**` ที่ใช้ `serverMainApiBase()`/`serverGoApiBase()` พร้อม allowlist path+method แคบ และ validate param; ห้ามเพิ่ม `rewrites()`/middleware ที่ชี้ mainapi (`frontend/src/lib/next-config-rewrites.test.ts` จะตก)

## Alternatives considered
- แก้ blocklist ให้ครบ — ไม่เลือก: ต้องไล่ตามทุก route/alias/การเข้ารหัสใหม่ตลอดไป และ path ที่ mainapi ถือเป็น public ยังหลุด
- ให้ Caddy บล็อก — ไม่เลือก: ปัญหาเดียวกับ blocklist และ Caddyfile บน prod ไม่ตรงกับ repo
- คง allowlist 5 เส้นของขั้น 1 ไว้ถาวร — ไม่เลือก: ยังเปิด mainapi สู่อินเทอร์เน็ตโดยตรง 5 เส้น, `has` แบบไม่ใส่ค่าทำ token รั่วเข้า query ปลายทางได้ (ต้องพึ่ง regex `Bearer .+`) และ rewrites ถูก bake ตอน build; BFF ให้ validate id/key และคุม header ได้เอง

## Consequences
- สคริปต์: `scripts/seed-demo.mjs` เรียก mainapi ตรงที่ `SEED_API` (ตั้งแต่ขั้น 1); `scripts/seed-gl-tax-rungrueng-2569.mjs:145-157` ใช้ `/api/organization/company` + `/api/organization/company/:code` ด้วย Bearer และไม่ส่ง `x-bc-backend-url`
- จอคลัง `/productwarehousescreen` (basePath `/warehouse` ไม่มีใน backend PostgreSQL) — ลบ `frontend/src/app/system-settings/warehouse-tree-view.tsx` (เรียก `/backend/warehouse/*` ที่ไม่มีแล้ว); route เดี่ยวแสดงการ์ด "รอพัฒนา" ตัวเดียวกับเมนู (`frontend/src/app/menu/menu-planned-card.tsx`, guard `frontend/src/app/system-settings/system-settings-screen.tsx:995,1452`); เมนูยังอยู่
- `deploy/account/Caddyfile.account` ลบบล็อก `@internal_health` (`/backend/goapi/api/health/*` → 404) เพราะไม่มี `/backend` ให้บล็อกแล้ว — Caddyfile บน prod ยังต้องเทียบเอง
- ลบ build-arg `BCAI_LOCAL_BACKEND_URL` ออกจาก `frontend/Dockerfile` และ `tools/fast-deploy.py` แล้ว: ที่เดียวที่ใช้ค่านี้ตอน build คือ `rewrites()` ซึ่งไม่มีแล้ว — Next ไม่ฝัง env ที่ไม่ขึ้นต้น `NEXT_PUBLIC_` และ route/SSR ทุกตัวอ่าน `serverMainApiBase()` ตอน runtime จาก `/etc/bcai-account/frontend.env` (`deploy/account/compose.yml` `env_file`, `provision-server.sh:119`)
- BFF อื่นที่เรียก mainapi (`proxyMainApiJson` ใน `frontend/src/lib/workspace-api.ts`) ยังไม่ส่ง IP ผู้ใช้ต่อ — ปลายทางที่อยู่ใต้ rate limiter ตาม IP (เช่น `/goapi/*` ที่เรียกผ่าน `/api/goapi`) จึงนับรวมที่ container frontend เหมือนก่อนงานนี้ (ยังค้าง)
- ไฟล์ส่วนตัวผ่าน Node เพิ่มหนึ่ง hop (สตรีม ไม่ buffer); cache ฝั่ง browser ยังใช้ ETag/`Cache-Control: private` ของ mainapi

## Evidence
- ขั้น 1: deploy r20260927-3 (commit `835cf27f`) + UAT prod — รายละเอียดเดิม `git show f75e59f9:docs/kms/decisions/2026-09-27-backend-proxy-allowlist.md`
- ขั้น 2 unit (2026-09-27, vitest): `next-config-rewrites.test.ts` 4/4, `api/organization/[...orgPath]/route.test.ts` 7/7, `api/files/[...key]/route.test.ts` 6/6, `image-upload-proxy.test.ts` 12/12, `image-display-url.test.ts` 3/3, `system-settings-screen.test.ts` 2/2, `settings-language-keys.test.ts` 2/2, `menu-screen-status.test.ts` 4/4, `resizable-splitter.test.ts` 12/12; `npx tsc --noEmit -p .` ผ่าน, eslint ไฟล์ที่แตะ 0 error (warning ไม่เกิน HEAD)
- ขั้น 2 UAT local (next dev + mainapi docker, ปุ่ม Demo → rungrueng/01/00000): `/api/organization/company|branch` 200; `/backend/organization/company`, `/backend/goapi/s3/file/...`, `/backend/metrics`, `/backend/v1/profile/link-line` = Next 404; `/api/files` ไฟล์ไม่มี 404, `%2e%2e%2f` 400, ไม่มี token 401, holding อื่น 403; รูปทดสอบ original 200 `image/png` + ETag → If-None-Match 304, thumbnail 200 `image/webp` → 304; โลโก้บริษัทโหลดผ่าน `/api/files/...?variant=thumbnail` 200; แก้ชื่อผู้จัดการสาขาบนจอ → `PUT /api/organization/branch/00000` 200 → PG `branches.settings.managername` = ค่าที่พิมพ์ → คืนค่า → PG ว่าง 1 แถว; PUT โลโก้บริษัท → PG `companies.logo_uri` = URI ทดสอบ → คืนค่า → ว่าง; ลบ object ทดสอบ 2 ตัวใน MinIO (bucket เปิด versioning = delete marker); `/productwarehousescreen` แสดงการ์ด "รอพัฒนา" ไม่มี request `/warehouse` ไม่มี console error
- **ขั้น 2 ยังไม่ deploy prod** — ต้อง build + deploy frontend แล้วยิงซ้ำบน prod
