---
date: 2026-09-27
severity: high
component: [frontend]
tags: [bc-account, security, nextjs, proxy, ssrf]
fixed: true
---

# Symptom

proxy สาธารณะ `/backend/*` ของ Next (`frontend/next.config.ts` ฉบับก่อน 2026-09-27: catch-all `/backend/:path*` → mainapi + blocklist ใน `beforeFiles` → `/_blocked-auth-route`) เลี่ยง blocklist ได้ ตรวจบน production แบบอ่านอย่างเดียว:

- `/backend/v1/profile/link-line`, `/backend/v1/profile/link-line/code`, `/backend/v1/login` ถึง mainapi (ตอบ 401 JSON ของ mainapi) ทั้งที่ `/backend/profile/link-line*` และ `/backend/login` อยู่ใน blocklist
- `/backend/goapi/%67et` ถึง mainapi (401) ขณะที่ `/backend/goapi/get` ถูกบล็อก
- ตอบ 200 ให้ใครก็ได้โดยไม่มี token: `/backend/metrics` (Prometheus — มี scanner ยิงเข้ามาแล้ว), `/backend/healthz`, `/backend/goapi/version`, `/backend/goapi/api/health`, `/backend/goapi/api/address/thailand`, `/backend/api/language/th`, `/backend/goapi/api/language/th`
- path ที่ "ถูกบล็อก" ตอบ 200 `text/html` (หน้า `[systemSetting]` เรียก `notFound()`) ไม่ใช่ 404 จริง

พบอีกจุดระหว่างตรวจ: SSR preload ภาษา (`getInitialBackendLanguage` → `loadBackendLanguageDictionary`) fetch `${cookie backend_url}/api/language/<lang>` ฝั่ง server — host มาจาก cookie ที่ผู้ใช้ตั้งเองได้ (http/https ใดก็ได้) = SSRF แบบจำกัด บน 7 หน้า และบน prod วนกลับเข้า `/backend` สาธารณะ (hairpin)

# Root Cause

- mainapi ลงทะเบียนทุก route ซ้ำอีกชุดใต้ `/v1` อัตโนมัติ (`backend/pkg/microservice/microservice_http.go:28-37` สำหรับ GET และแบบเดียวกันใน POST/PUT/PATCH/DELETE) — blocklist เขียนแค่ path ปกติ (มี `/backend/v1/dev-login`, `/backend/v1/demo-login` แค่สองตัว) จึงหลุดทุก route ที่เหลือ
- Next จับคู่ `source` กับ pathname ที่ยังเข้ารหัสอยู่ (`%67` ไม่ใช่ `g`) จึงเลี่ยง blocklist ได้และส่งต่อถึง mainapi — แต่ **ไม่ได้** ถึง handler ที่ถูกบล็อก: Echo v4.13.3 (`backend/go.mod:81`) เลือก route จาก `RawPath` (`GetPath` ใน `echo.go:949-955`) ซึ่งไม่ถอดรหัส segment คงที่ และ goapi ไม่มี route `/goapi/get` อยู่แล้ว — 401 ที่เห็นมาจาก middleware auth ของกลุ่ม goapi ที่ครอบทุก path ใต้ `/goapi/*` (`backend/internal/goapi/bootstrap.go:169-175,308`); ยิง mainapi local ตรง 2026-09-27: `/goapi/version` 200 แต่ `/goapi/%76ersion` 401, `/metrics` 200 แต่ `/%6detrics` 404, `/goapi/%67et` 401 เท่ากับ `/goapi/zzz-nonexistent` — การเข้ารหัสจึงพิสูจน์แค่ว่า blocklist รั่วเชิงหลักการ ช่องโหว่ที่ใช้ได้จริงคือ alias `/v1` (เช่น `/backend/v1/profile/link-line*` ถึง handler ผูก LINE ได้สำหรับผู้ใช้ที่ login แล้ว) และเส้นทางที่ mainapi ถือเป็น public ในข้อถัดไป (ส่วน `/backend/v1/login` ได้ 401 เพราะ `publicPath` เทียบแบบตรงตัว `/login` — `backend/pkg/microservice/auth.go:164-173` — ไม่ใช่ทางเลี่ยง login)
- blocklist แบบ "เปิดทุกอย่าง ปิดเป็นรายตัว" ต้องตามทุก route ใหม่/alias/การเข้ารหัส ซึ่งพลาดง่าย; ส่วน `/metrics`, `/healthz`, `/goapi/*` อยู่ใน `publicPath` ของ mainapi (`backend/main.go:119-135`) จึงเปิดสู่อินเทอร์เน็ตผ่าน catch-all ทันที
- ทางเดียวจากอินเทอร์เน็ตถึง mainapi คือ rewrite นี้ (Caddy ส่งทุก path ไป frontend `127.0.0.1:3200`, mainapi ผูก `127.0.0.1:8888`)
- preload ภาษาใช้ host จาก cookie `backend_url` แทน `serverGoApiBase()` ที่ BFF อื่นทุกตัวใช้

# Fix

- `frontend/next.config.ts:34-44` — ลบ blocklist ทั้งสองชุด, `beforeFiles` → `/_blocked-auth-route` และ catch-all แทนด้วย **allowlist 5 source** ใน `afterFiles` ที่ browser เรียกจริง: `/backend/goapi/s3/file/:path+`, `/backend/organization/company`, `/backend/organization/company/:id`, `/backend/organization/branch`, `/backend/organization/branch/:id` — แต่ละตัวชี้ปลายทาง mainapi แบบ path ตายตัว (ตัด `/backend`) ดังนั้น `/v1`, `..%2f`, ตัวพิมพ์ต่าง ๆ พาไป route อื่นไม่ได้; ทุกตัวมี `has` header `authorization` = `Bearer .+` (ตัดเสียงรบกวน scanner เท่านั้น mainapi ยังตรวจ token) — ใช้ regex ไม่มี named group เพราะ `has` แบบไม่มีค่าจะก๊อป token เข้า params แล้ว Next ต่อท้าย query ปลายทางที่ไม่มี param (`?authorization=Bearer...`) ทำให้ token รั่วเข้า URL/log ของ mainapi (ยืนยันกับ `matchHas`/`prepareDestination` ของ Next 16.3.0)
- `frontend/src/lib/backend-language-preload.ts:25-51` — `loadBackendLanguageDictionary(language, fetcher)` ไม่รับ URL แล้ว ใช้ `serverGoApiBase()` เสมอ; env หาย → `{}` ไม่ทำหน้าพัง; `frontend/src/lib/backend-language-server.ts:19-21` ยังอ่าน cookie `backend_url` เพื่อส่งให้จอฝั่ง client เป็น fallback เท่านั้น server ไม่ fetch
- ลบโค้ดที่สร้าง URL `/backend/...` ที่ไม่มีทางทำงาน: fallback `<backend>/images/...` ใน `frontend/src/components/authenticated-image.tsx` (`imageDisplayUrl`) และ `frontend/src/components/logo-avatar.tsx` (`resolveDisplayUrl`) รวมถึงการแปลง `/api/...` → `/backend/api/...` ใน `logo-avatar.tsx` (ตอนนี้ `/api/*` อยู่ที่ origin ของ frontend เหมือน `authenticated-image.tsx`); URI `/goapi/s3/file/...`, `http(s)://`, `data:`/`blob:` ทำงานเหมือนเดิม
- `scripts/seed-demo.mjs` เลิกยิงผ่าน `${SEED_BASE}/backend` — login ผ่าน BFF แล้วเรียก mainapi ตรงที่ `SEED_API` (ค่าเริ่มต้น `http://127.0.0.1:8888`; prod ใช้ ssh tunnel)
- แก้ comment ที่ผิดใน `frontend/src/lib/backend-url.ts:111-117` และ docs `01`, `03`, `09`, `10`, บั๊ก `2026-09-25-line-code-route-unauthenticated.md`; ADR `decisions/2026-09-27-backend-proxy-allowlist.md`

# Regression Test

- `frontend/src/lib/next-config-rewrites.test.ts` (7 ข้อ) — จำลองขั้นตอนของ Next ด้วยฟังก์ชันจริง `parseUrl` → `getPathMatch` → `matchHas` → `prepareDestination`: ไม่มี `beforeFiles`/`fallback`/`*`; `afterFiles` ทั้งหมด (ไม่กรองเฉพาะ source `/backend`) ต้องเท่ากับ 5 source พอดี — rule ใหม่ที่ชี้ mainapi ด้วย prefix อื่นก็ตก; ทุก rule ปลายทาง = source ตัด `/backend` + `has` Bearer; path ที่อนุญาต 6 ตัว match rule เดียวและได้ปลายทางตรง; query ผ่านได้แต่ไม่มี `authorization` ใน query; ไม่มี header / `Basic` → ไม่ proxy; path ต้องห้าม 32 ตัว (รวม `/backend/v1/profile/link-line*`, `/backend/v1/login`, `/backend/goapi/%67et`, `/backend/metrics`, `/backend/healthz`, `/backend/goapi/api/language/th`, `/backend/goapi/s3/file/../../get`, `%2e%2e`, ตัวพิมพ์ใหญ่) ไม่ match rule ใด; `/BACKEND/ORGANIZATION/COMPANY` ลงปลายทางตัวพิมพ์เล็กตายตัว; `..%2f` ใน `:id`/`:path+` อยู่ใน prefix ของ route ที่อนุญาต
- `frontend/src/lib/backend-language-preload.test.ts` (4 ข้อ) — fetch ไป `BCAI_LOCAL_BACKEND_URL/goapi/api/language/<lang>`, env ว่าง → `{}` ไม่ fetch, backend ตอบ error → `{}`
- `frontend/src/lib/backend-language-server.test.ts` (2 ข้อ) — cookie `backend_url=http://attacker.example/goapi` ไม่ถูก fetch เลย (fetch ครั้งเดียวไปที่ env) และ env หาย → `{}` ไม่ throw
- `frontend/src/components/image-display-url.test.ts` (3 ข้อ) — `imageDisplayUrl` (authenticated-image) และ `resolveDisplayUrl` (logo-avatar): URI `/goapi/s3/file/<key>` (+`?variant=thumbnail`) → `<origin>/backend/goapi/s3/file/<key>` ซึ่งเป็นรูปแบบเดียวที่ allowlist ส่งต่อ; `/api/*`, `https://`, `data:`, `/banks/*` ไม่ถูกแตะ; ไม่มีกรณีใดสร้าง `/backend/images/*` หรือ `/backend/api/*`

# ยังค้าง

- deploy แล้ว r20260927-3 (commit `835cf27f`) และยิงซ้ำบน prod ผ่านครบ — หลักฐานใน ADR `decisions/2026-09-27-backend-proxy-allowlist.md` หัวข้อ Evidence
- mainapi `/profile/link-line*` ยังเชื่อ `code` + `lineuserid` ที่ผู้เรียกส่งมา (งาน LINE พักไว้ตามคำสั่งลุงจืด) — ตอนนี้ปลอดภัยจากอินเทอร์เน็ตเพราะ allowlist ไม่มีเส้นทางนี้ แต่เครื่องที่เปิด `:8888` ออก LAN ยังยิงตรงได้
- Caddy บน prod มี vhost HTTP ของ IP เซิร์ฟเวอร์ และ Caddyfile บนเซิร์ฟเวอร์ไม่ตรงกับ `deploy/account/Caddyfile.account` ใน repo; ไม่มี access log ให้ตรวจย้อนหลังว่าเคยมีใครใช้ช่องทางนี้
- path เดี่ยวอย่าง `/backend` ยังตกไปที่หน้า `[systemSetting]` (`notFound()` ตอบ 200 `text/html`) — ไม่ถึง mainapi แต่ไม่ใช่ 404 จริง
- `frontend/src/app/system-settings/warehouse-tree-view.tsx` ยังเรียก `/warehouse/*` ผ่าน `/backend` ซึ่ง backend ไม่มีแล้ว (จอพังอยู่ก่อน) — ไม่เปิดใน allowlist
