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

- **ขั้น 1** (commit `835cf27f`, deploy r20260927-3): catch-all + blocklist → allowlist 5 เส้นใน `rewrites()`; SSR preload ภาษาใช้ `serverGoApiBase()` เสมอ (`frontend/src/lib/backend-language-preload.ts:25-51`) และ cookie `backend_url` ไม่ถูก server fetch (`frontend/src/lib/backend-language-server.ts:19-21`); ลบโค้ดที่สร้าง URL `/backend/images/*`, `/backend/api/*`; `scripts/seed-demo.mjs` เรียก mainapi ตรงที่ `SEED_API` — รายละเอียดเดิม `git show f75e59f9:docs/kms/bugs/2026-09-27-backend-proxy-blocklist-bypass.md`
- **ขั้น 2** (2026-09-27, ยังไม่ deploy): ลบ `rewrites()` ทั้งหมดจาก `frontend/next.config.ts` (เหลือ comment `:7-9`) → ไม่มีทางใดจาก browser ถึง mainapi นอกจาก BFF; บริษัท/สาขาผ่าน `frontend/src/app/api/organization/[...orgPath]/route.ts`, ไฟล์ S3 ผ่าน `frontend/src/app/api/files/[...key]/route.ts` (ตรวจ segment + ส่งต่อแค่ `variant=thumbnail`), สร้างสาขาผ่าน `POST /api/workspace/branch` เดิม; ตัวแสดงรูปแปลง `/goapi/s3/file/<key>` → `/api/files/<key>` (`frontend/src/lib/image-upload-proxy.ts:231`); `scripts/seed-gl-tax-rungrueng-2569.mjs:145-157` ใช้ BFF; ลบ `frontend/src/app/system-settings/warehouse-tree-view.tsx` (เรียก `/warehouse/*` ที่ backend ไม่มี) และบล็อก `@internal_health` ใน `deploy/account/Caddyfile.account` — รายละเอียดใน [ADR](../decisions/2026-09-27-backend-proxy-allowlist.md)

# Regression Test

- `frontend/src/lib/next-config-rewrites.test.ts` (4 ข้อ) — `rewrites`/`redirects` ต้อง undefined; config ที่ serialize ต้องไม่มี host backend, `mainapi`, `:8888`, `/backend`; ไม่มีไฟล์ `middleware`/`proxy` ที่ root และ `src`; ไม่มี `src/app/backend` และ `getSystemSettingConfig("backend")` = undefined
- `frontend/src/app/api/organization/[...orgPath]/route.test.ts` (8 ข้อ) — path/method นอก allowlist = 404 ไม่ fetch; รหัสที่สร้างได้ทุกแบบ (`A&B`, `K+1(ก)!`, `a%2Fb`, ช่องว่าง, `X,Y;Z=1@2:$`) ส่งต่อด้วย escaping แบบ Go (`goPathEscape` — ตรวจกับ Echo v4.13.3 จริงว่า `:id` ได้รหัสตรงตัว); id `..`, `.`, ว่าง, `01/../x`, backslash, อักขระควบคุม, ยาว 129 = 400; body ไม่ใช่ object = 400; ไม่มี Bearer = 401; host จาก `x-bc-backend-url`/body ไม่ถูก fetch; status/JSON ของ mainapi ผ่านตรง
- `frontend/src/app/api/files/[...key]/route.test.ts` (7 ข้อ) — สตรีม + header ที่อนุญาต ไม่ส่ง `set-cookie`; ส่ง IP แรกของ `X-Forwarded-For` ต่อ (ค่าที่ไม่ใช่ IP ไม่ส่ง) ให้ rate limiter ของ mainapi นับรายคน; ส่งต่อแค่ `variant=thumbnail`; 304/403 ผ่าน; redirect จากต้นทาง = 502 ไม่มี `location`; traversal/ว่าง/`/`/`\`/อักขระควบคุม/ยาวเกิน = 400 ไม่ fetch; ไม่มี Bearer = 401; host จาก client ถูกเมิน
- `frontend/src/components/image-display-url.test.ts` (3 ข้อ) + `frontend/src/lib/image-upload-proxy.test.ts` (12 ข้อ) — URI ที่เก็บ → `/api/files/<key>`; ไม่มีกรณีใดสร้าง URL `/backend`; แนบ Bearer เฉพาะ `/api/files` origin เดียวกัน
- `frontend/src/lib/backend-language-preload.test.ts` (4 ข้อ), `frontend/src/lib/backend-language-server.test.ts` (2 ข้อ) — จากขั้น 1: preload fetch แค่ `BCAI_LOCAL_BACKEND_URL`, cookie `backend_url` ของผู้โจมตีไม่ถูก fetch

# ยังค้าง

- ขั้น 2 ยังไม่ deploy prod — prod ยังเป็นขั้น 1 (allowlist 5 เส้น, r20260927-3) จนกว่าจะ build + deploy frontend แล้วยิงซ้ำ (`/backend/organization/company` + Bearer ต้องเป็น 404, `/api/files/<key>` + Bearer ต้อง 200)
- mainapi `/profile/link-line*` ยังเชื่อ `code` + `lineuserid` ที่ผู้เรียกส่งมา (งาน LINE พักไว้ตามคำสั่งลุงจืด) — อินเทอร์เน็ตถึงไม่ได้เพราะไม่มี `/backend` แล้ว แต่เครื่องที่เปิด `:8888` ออก LAN ยังยิงตรงได้
- ~~Caddy บน prod มี vhost HTTP ของ IP / Caddyfile ไม่ตรง repo / ไม่มี access log~~ — แก้แล้ว 2026-09-27: ลบ vhost `http://159.223.43.229`, เปิด access log `/var/log/caddy/account-access.log` (Authorization/Cookie = REDACTED), block ของ account ตรงกับ `deploy/account/Caddyfile.account` (ดู `docs/kms/10-infra-deploy.md`)
- path เดี่ยว `/backend` ยังตกไปที่หน้า `[systemSetting]` (`notFound()` ตอบ 200 `text/html` — ตรวจ local 2026-09-27; `/backend/<อะไรก็ได้>` = 404) — ไม่ถึง mainapi แต่ไม่ใช่ 404 จริง
