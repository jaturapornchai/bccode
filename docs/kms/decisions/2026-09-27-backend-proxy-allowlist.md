---
date: 2026-09-27
status: accepted
tags: [bc-account, security, nextjs, frontend, proxy]
supersedes: blocklist `blockedAuthRoutes`/`blockedDangerousRoutes` + catch-all `/backend/:path*` ใน `frontend/next.config.ts`
---

# proxy สาธารณะ `/backend/*` ของ Next เป็น allowlist — การเรียก mainapi อื่นต้องผ่าน BFF

## Context
ทางเดียวจากอินเทอร์เน็ตถึง mainapi คือ rewrite ของ Next (Caddy ส่งทุก path ไป frontend `127.0.0.1:3200`, mainapi ผูก `127.0.0.1:8888`). เดิม rewrite เป็น catch-all `/backend/:path*` → mainapi แล้วปิดเส้นทางอันตรายเป็นรายตัวใน `beforeFiles` → `/_blocked-auth-route`. ตรวจ production (อ่านอย่างเดียว) 2026-09-27 พบว่า blocklist ถูกเลี่ยงได้ — alias `/v1` ที่ mainapi ลงทะเบียนให้ทุก route (`backend/pkg/microservice/microservice_http.go:28-37`), URL-encoding (`/backend/goapi/%67et` หลุด blocklist ไปถึง mainapi — Echo เลือก route จาก raw path จึงจบที่ auth ของ goapi ไม่ใช่ handler `/goapi/get` แต่พิสูจน์ว่า blocklist รั่ว) และ `/backend/metrics`, `/backend/healthz`, `/backend/goapi/*` สาธารณะตอบ 200 (รายละเอียด: [บั๊ก](../bugs/2026-09-27-backend-proxy-blocklist-bypass.md)).

ตรวจโค้ด browser แล้วเรียก `/backend` จริงแค่ 2 กลุ่ม และส่ง `Authorization: Bearer` ผ่าน `authFetch` ทุกครั้ง:
- รูป/ไฟล์ `GET /backend/goapi/s3/file/<key>[?variant=thumbnail]` (`frontend/src/components/authenticated-image.tsx`, `frontend/src/components/logo-avatar.tsx`, `frontend/src/lib/image-upload-proxy.ts:224-256`) — DB เก็บ `/goapi/s3/file/<key>`
- บริษัท/สาขา `GET|POST /backend/organization/company|branch`, `PUT .../:id` (`frontend/src/app/system-settings/company-branch-tree-view.tsx`, `frontend/src/app/currency/currency-screen.tsx`)

BFF ทุกตัวใต้ `frontend/src/app/api/**` (รวม LINE `/api/auth/line/*`, `/mcp/gl`, `/api/integration/gl/*`) เรียก mainapi ฝั่ง server ด้วย `serverMainApiBase()`/`serverGoApiBase()` (`frontend/src/lib/backend-url.ts:118-128`) ไม่ผ่าน `/backend`.

## Decision
- `frontend/next.config.ts:34-44` เหลือ `afterFiles` 5 rule เท่านั้น: `/backend/goapi/s3/file/:path+`, `/backend/organization/company`, `/backend/organization/company/:id`, `/backend/organization/branch`, `/backend/organization/branch/:id` — ปลายทาง = path เดียวกันบน `BCAI_LOCAL_BACKEND_URL` แบบตายตัว; ไม่มี `beforeFiles`, ไม่มี catch-all, ไม่มี `/_blocked-auth-route`
- ทุก rule มี `has: [{ type: "header", key: "authorization", value: "Bearer .+" }]` — ตัดเสียงรบกวนจาก scanner ที่ไม่มี token เท่านั้น **ไม่ใช่** ขอบเขตความปลอดภัย (mainapi ตรวจ token/สิทธิ์เอง); ต้องมีค่า regex ที่ไม่มี named group เพราะ `has` แบบไม่ใส่ค่าทำให้ Next ก๊อปค่า header เข้า params แล้วต่อท้าย query ของปลายทางที่ไม่มี param (`/organization/company?authorization=Bearer...`) = token รั่วเข้า URL/log (ตรวจกับ `matchHas`/`prepareDestination` ของ Next 16.3.0)
- ไม่เปิด: `/v1/*`, `/metrics`, `/healthz`, login/register/googlelogin, `/profile/*` (รวม LINE), language, `/goapi/*` อื่น, `/warehouse/*` (backend ไม่มีแล้ว), `/mcp*`, `/integration/*`, `/gl/*`, `/fa/*`, `/holding/*`
- SSR preload ภาษาใช้ `serverGoApiBase()` แทน host จาก cookie `backend_url` (`frontend/src/lib/backend-language-preload.ts:25-51`)
- **กติกาต่อจากนี้:** browser ต้องการ endpoint ใหม่ของ mainapi/goapi → ทำ BFF route ใต้ `frontend/src/app/api/**` ที่ใช้ `serverMainApiBase()`/`serverGoApiBase()`; ห้ามเพิ่ม rule ใน allowlist โดยไม่มี ADR ใหม่ และต้องเพิ่มกรณีใน `frontend/src/lib/next-config-rewrites.test.ts`

## Alternatives considered
- **แก้ blocklist ให้ครบ** (เพิ่ม `/backend/v1/*`, decode ก่อนเทียบ ฯลฯ) — ไม่เลือก: ต้องไล่ตามทุก route/alias/การเข้ารหัสใหม่ตลอดไป พลาดครั้งเดียวก็เปิด และ `/metrics`/`/goapi/*` ที่ mainapi ถือเป็น public ยังหลุดอยู่ดี
- **ให้ Caddy บล็อกแทน** — ไม่เลือก: ปัญหาเดียวกับ blocklist และ Caddyfile บน prod ไม่ตรงกับ repo อยู่แล้ว
- **ขั้นที่ 2: ย้ายบริษัท/สาขา + การสตรีมไฟล์ S3 เข้า BFF แล้วลบ `/backend` ทั้งหมด** — เป็นเป้าหมายที่ถูกต้องกว่า แต่ยังไม่ทำตอนนี้: ต้องแก้จอ `company-branch-tree-view.tsx`/`currency-screen.tsx` + ตัวแสดงรูป (`authenticated-image.tsx`, `logo-avatar.tsx`) และทำ BFF สตรีม binary (thumbnail/cache header) ซึ่งใหญ่กว่างานปิดช่องโหว่; allowlist ปิดความเสี่ยงได้ทันทีด้วยการเปลี่ยนไฟล์เดียว
- **เปิด `has` แบบไม่มีค่า (ตามแผนแรก)** — ไม่เลือก: ทำให้ token ถูกต่อท้าย query (ดู Decision)

## Consequences
- สคริปต์ที่เคยยิง mainapi ผ่าน `${SEED_BASE}/backend` ต้องเรียก mainapi ตรง: `scripts/seed-demo.mjs` ใช้ `SEED_API` (ค่าเริ่มต้น `http://127.0.0.1:8888`, prod ผ่าน `ssh -L 8888:127.0.0.1:8888`); `scripts/seed-gl-tax-rungrueng-2569.mjs` ใช้ `/backend/organization/company/:code` + Bearer ซึ่งยังอยู่ใน allowlist
- โค้ดที่สร้าง URL `/backend/images/*` (mainapi ไม่มี route นี้) และ `/backend/api/*` (mainapi มีแค่ `/api/language/:lang` ซึ่งไม่ใช่รูป) ถูกลบ
- `rewrites()` ถูก bake ตอน build (`tools/fast-deploy.py:97`) — มีผลบน prod หลัง build + deploy frontend เท่านั้น
- `/backend/organization/branch/:id` ครอบ `GET /organization/branch/list` ด้วย (ต้องมี token เหมือนกัน)
- `..%2f` ภายใน `:id`/`:path+` ยังถูกส่งต่อ แต่ปลายทางถูกตรึง prefix — ถึงได้แค่ handler ของ route นั้น (การตรวจ key ภายใน handler S3 ไม่อยู่ในขอบเขต ADR นี้)

## Evidence
- Unit: `frontend/src/lib/next-config-rewrites.test.ts` 7/7, `frontend/src/lib/backend-language-preload.test.ts` 4/4, `frontend/src/lib/backend-language-server.test.ts` 2/2, `frontend/src/components/image-display-url.test.ts` 3/3 (2026-09-27)
- UAT production: pending
