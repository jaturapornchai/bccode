# 09 — Frontend (Next.js BFF): หน้าจอ, API proxy, session, ธีม และกฎ UX
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวมและเวอร์ชัน

- Frontend = Next.js App Router ทำหน้าที่ **BFF** (Backend-for-Frontend): browser เรียก `/api/**` ของ Next แล้ว route handler ฝั่ง server fetch ไป mainapi/goapi ที่ `BCAI_LOCAL_BACKEND_URL` เสมอ (`frontend/src/lib/backend-url.ts:117-127`, `frontend/src/lib/workspace-api.ts:24-31`)
- เวอร์ชันหลัก: `next 16.3.0`, `react 19.2.8`, `typescript 6.0.3`, `tailwindcss ^4.3.3`, `vitest 4.1.10`, `@playwright/test ^1.62.1`, engines `node 24.x / npm 11.x` (`frontend/package.json:5-8,30,32,41,49-51`)
- ไลบรารีเด่น: `@tanstack/react-query`, `@tanstack/react-table`, `react-hook-form` + `zod`, `motion`, `recharts`, `leaflet`/`react-leaflet`, `lucide-react`, Radix dropdown (`frontend/package.json:20-39`)
- ไม่มี `middleware.ts`/`proxy.ts` ใน `frontend/src` (ตรวจด้วย `git ls-files frontend/src` — ไฟล์นอก `app/lib/components/locales` = 0)
- Scripts: `dev`, `build`, `lint`, `start`, `typecheck` (`tsc --noEmit`), `test` (`vitest run`), `test:e2e` (`playwright test`) (`frontend/package.json:9-17`)

## 2. เส้นทาง network: rewrite `/backend/*` และ route ที่ถูกบล็อก

- `next.config.ts` rewrite `/backend/:path*` → `${BCAI_LOCAL_BACKEND_URL}/:path*` (afterFiles) และ **throw ถ้าไม่ตั้ง env** (`frontend/next.config.ts:19-22,73-75`)
- บล็อก (beforeFiles → `/_blocked-auth-route`) กลุ่ม login/register ที่ browser ไม่ควรเรียกตรง เช่น `/backend/login`, `/backend/googlelogin`, `/backend/dev-login`, `/backend/demo-login`, `/backend/register-username` (`frontend/next.config.ts:29-45`) และกลุ่มอันตราย `/backend/goapi/get|exec|getdoc`, `/backend/reportm/*`, `/backend/goapi/api/setup/*`, `/backend/goapi/api/mcp/*`, `/backend/goapi/mcp/*`, `/backend/reload-config` (`frontend/next.config.ts:58-67`)
- Header ทุก path: `Referrer-Policy: no-referrer-when-downgrade`, `COOP: same-origin-allow-popups` (จำเป็นสำหรับ Google popup) (`frontend/next.config.ts:10-15`)
- `allowedDevOrigins` ระบุ IP Tailscale 1 ค่าเพื่อให้เครื่องอื่นเรียก dev server ได้ (`frontend/next.config.ts:6`)
- Browser ใช้ URL สาธารณะรูป `<origin>/backend/goapi` (`frontend/src/lib/backend-url.ts:6-8`) ส่วน server-side ใช้ `serverMainApiBase()`/`serverGoApiBase()` — client URL ถูก validate แล้วทิ้ง ไม่ถูก echo กลับ (`frontend/src/lib/workspace-api.ts:24-31`); `public/config.json` มี key เดียว `goapi_url` (ตรวจด้วย node; ไม่คัดลอกค่า)

## 3. Auth / Session model

| ชิ้นส่วน | พฤติกรรม | อ้างอิง |
|---|---|---|
| Refresh cookie | `bc_refresh_token`, httpOnly, sameSite=lax, maxAge 12 ชม., `secure` เฉพาะ `NODE_ENV=production` | `frontend/src/lib/auth-session-server.ts:3-8,25-34` |
| Access token ฝั่ง client | เก็บใน memory + `localStorage["bc_auth"]` (ไม่เก็บ token ลง storage: `StoredAuthSession = Omit<AuthSession,"token">`), sync ข้ามแท็บผ่าน `BroadcastChannel("bc-auth-session")` | `frontend/src/lib/client-auth-session.ts:4,24,38-57`, `frontend/src/lib/workspace-models.ts:79-83` |
| Refresh/auto-retry | `restoreAuthSession()` เรียก `POST /api/auth/refresh`; `authFetch()` แนบ Bearer; `logoutAuthSession()` เรียก `/api/auth/logout` | `frontend/src/lib/client-auth-session.ts:121,144,174,201-205` |
| Bearer guard ใน BFF | `requireBearerToken()` ตอบ 401 ไทยถ้าไม่มี header | `frontend/src/lib/workspace-api.ts:8-14` |
| JWT verify ฝั่ง Next | HS256 ด้วย `JWT_SECRET_KEY` + ตรวจ `exp` + อ่าน claim `holdingcode` — ใช้เป็น tenant guard ใน line-oa และ system-settings (ข้ามถ้าไม่ตั้ง env) | `frontend/src/lib/server-jwt.ts:7-39`, `frontend/src/app/api/line-oa/user/route.ts:98-110`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts:178-184` |
| Holding access guard | system-settings ยิง `POST /select-holding` ไป mainapi ก่อน **เฉพาะ `kind === "atlas"`** (kind อื่น return null หลังตรวจ JWT) ถ้าไม่ผ่านตอบ 401/403 | `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts:183-184,207,225` |
| Login methods | username/password → `/login`; Google → ตรวจ `oauth2.googleapis.com/tokeninfo` แล้วค่อย `/googlelogin`; Demo → `/demo-login` (backend ตัดสินด้วย `BCAI_DEMO_LOGIN_ENABLED`); Dev → BFF `/api/auth/dev-login` → `/dev-login` ต้อง `BCAI_DEV_LOGIN_ENABLED=true` และ secret ≥ 32 ตัวอักษร (**หน้า login ไม่มีปุ่ม Dev Login แล้ว** — ปุ่มถูกแทนด้วย Demo, `frontend/src/app/login-screen.tsx:304-309`; route BFF ยังอยู่) | `frontend/src/app/api/auth/login/route.ts:63`, `frontend/src/app/api/auth/google/verify/route.ts:54-55,82`, `frontend/src/app/api/auth/demo-login/route.ts:9,27`, `frontend/src/app/api/auth/dev-login/route.ts:9,22-27,34` |
| Root layout bootstrap | `<AuthSessionBootstrap>` ครอบทุกหน้า + `<ToastViewport>` | `frontend/src/app/layout.tsx:91-92` |

## 4. หน้าจอ (App Router pages)

page ส่วนใหญ่เป็น server component บาง ๆ ที่อ่านภาษาจาก cookie แล้ว render screen component client (`frontend/src/lib/backend-language-server.ts:16-19`) — ยกเว้น `/` ที่ render screen ตรงโดยไม่อ่านภาษา (`frontend/src/app/page.tsx:9-10`)

| route | หน้าที่ | screen component | สถานะ | อ้างอิง |
|---|---|---|---|---|
| `/` | Login (username, Google, Demo — ไม่มีปุ่ม Dev Login; ปุ่มที่ใช้ class `dev-login-button` คือปุ่ม Demo) | `login-wrapper.tsx` → `login-screen.tsx` (835 บรรทัด) | LIVE | `frontend/src/app/page.tsx:2,9-10`, `frontend/src/app/login-wrapper.tsx:4,73`, `frontend/src/app/login-screen.tsx:304-309,649-657` |
| `/holding` | เลือก/สร้างกลุ่มกิจการ + จัดการ admin ตาม email | `holding/holding-screen.tsx` (1547) | LIVE | `frontend/src/app/holding/page.tsx:3,11` |
| `/workspace` | เลือกบริษัท/สาขา, สร้าง holding/branch, หน่วยนับเริ่มต้น, ผูก LINE | `workspace/workspace-screen.tsx` (2315) | LIVE | `frontend/src/app/workspace/page.tsx:3,11` |
| `/menu` | เมนูหลัก + dashboard + ทางลัด + สินค้า/บาร์โค้ด/ชุดสินค้า/marketplace (tabs) | `menu/main-menu-screen.tsx` (2842) และ `menu/*-screen.tsx` | LIVE | `frontend/src/app/menu/page.tsx:3,11` |
| `/[systemSetting]` | 31 config slug (นับ `slug:` ไม่ซ้ำใน `SYSTEM_SETTING_CONFIGS` + `productMasterConfigs()`; 2 ตัวคือ `permissionlink`→redirect `/user`, `approvalsetting`→redirect `/workspace` จึงเหลือ 29 จอที่ render จริง) เช่น company, branch, employee, user, productunit, debtor; slug ไม่รู้จัก → 404 | `system-settings/system-settings-screen.tsx` (**16600 บรรทัด**) | LIVE | `frontend/src/app/[systemSetting]/page.tsx:17-24`, `frontend/src/lib/system-setting-screens.ts:287-1180,1182-1769,1771-1777` |
| `/currency` | สกุลเงิน CRUD | `currency/currency-screen.tsx` (1224) | LIVE | `frontend/src/app/currency/page.tsx:3,11` |
| `/price_history` | ประวัติราคาสินค้า | `menu/product-price-history-screen.tsx` | LIVE | `frontend/src/app/price_history/page.tsx:3,11` |
| `/product_barcode_shelf` | ชั้นวางบาร์โค้ด | `menu/product-barcode-shelf-screen.tsx` | LIVE | `frontend/src/app/product_barcode_shelf/page.tsx:3,11` |
| `/line-oa` | ผูก LINE OA กับผู้ใช้ | `line-oa/line-oa-link-screen.tsx` | LIVE | `frontend/src/app/line-oa/page.tsx:3,11` |
| `/manual`, `/manual/[screen]` | คู่มือ th/en อ่านจากไฟล์ `frontend/manual/*.json` (login, workspace, menu, currency + ทุก slug) | server-rendered | LIVE (มี json 9 ไฟล์) | `frontend/src/app/manual/page.tsx:1,30-31,60-71`, `frontend/Dockerfile:47` |
| `/favicon.ico` | route handler ส่ง icon พร้อม cache 1 ปี | — | LIVE | `frontend/src/app/favicon.ico/route.ts:8-11` |

## 5. API route handlers (36 ไฟล์ `route.ts` ใต้ `frontend/src/app/api/`; รวม `favicon.ico/route.ts` = 37 — นับจาก `git ls-files frontend/src/app | grep route.ts`) → backend path

ฐาน: **M** = mainapi (`serverMainApiBase()`), **G** = goapi (`serverGoApiBase()` = `M/goapi`)

| Next route | method | ปลายทาง backend | auth ที่ BFF | สถานะ | อ้างอิง |
|---|---|---|---|---|---|
| `/api/auth/login` | POST | M `/login` → set refresh cookie | validate ไทย 400 | LIVE | `auth/login/route.ts:13-19,63,107` |
| `/api/auth/refresh` | POST | M `/refresh` (อ่าน cookie) | cookie | LIVE | `auth/refresh/route.ts:10-21,42` |
| `/api/auth/logout` | POST | M `/logout` (+ `/refresh` ถ้าไม่มี Bearer) แล้ว clear cookie | Bearer หรือ cookie | LIVE | `auth/logout/route.ts:6-26,57` |
| `/api/auth/demo-login` | POST | M `/demo-login` | ไม่มี (backend gate) | LIVE | `auth/demo-login/route.ts:11-27,51` |
| `/api/auth/dev-login` | POST | M (หรือ `BCAI_DEV_LOGIN_BACKEND_URL`) `/dev-login` | env gate + secret ≥32 | LIVE (dev) | `auth/dev-login/route.ts:9,22-27,34,122` |
| `/api/auth/google/verify` | POST | Google tokeninfo → M `/googlelogin` | ตรวจ `GOOGLE_CLIENT_ID` | LIVE | `auth/google/verify/route.ts:29,54-55,82` |
| `/api/auth/google/session` | POST | `BC_AUTH_BRIDGE_URL` `/api/google/session` | — | LEGACY (bridge) | `auth/google/session/route.ts:8-18` |
| `/api/auth/google/status` | POST | — ตอบ 410 | — | STUBBED | `auth/google/status/route.ts:6-10` |
| `/api/auth/line/code` | POST | bridge `/api/login?action=create` | — | LIVE (ต้อง bridge) | `auth/line/code/route.ts:4-7` |
| `/api/auth/line/link/status` | POST | bridge `/api/login?code=` → M `PUT /profile/link-line` | Bearer | LIVE | `auth/line/link/status/route.ts:31,54,121` |
| `/api/auth/line/status` | POST | — ตอบ 410 (LINE login ปิด) | — | STUBBED | `auth/line/status/route.ts:5-9` |
| `/api/auth/register-username` | POST | M `/register-username` | validate | LIVE | `auth/register-username/route.ts:11,47` |
| `/api/auth/profile` | GET/PUT | M `/profile`, `/profile/password` (PUT สำเร็จ → clear cookie) | Bearer | LIVE | `auth/profile/route.ts:5-26` |
| `/api/auth/profile/reset-password` | PUT | — ตอบ 501 | — | STUBBED | `auth/profile/reset-password/route.ts:3-12` |
| `/api/auth/sessions` | GET | M `/sessions/active-count` | Bearer | LIVE | `auth/sessions/route.ts:5-8` |
| `/api/backend/check` | POST | G `/api/health`, `/version` | ไม่มี | LIVE | `backend/check/route.ts:35-39` |
| `/api/address/thailand` | GET | G `/api/address/thailand` (force-cache + revalidate) | ไม่มี | LIVE | `address/thailand/route.ts:8,24-30` |
| `/api/language/[lang]` | GET | G `/api/language/{lang}` (no-store) | ไม่มี | LIVE | `language/[lang]/route.ts:9,31` |
| `/api/workspace/[...path]` | GET/POST | M: `holdings`→`/list-holding?limit=100`+`/holding/{code}`, `holding-info`, `branches`→`/organization/branch/list`, `product-units*`→`/unit/list`,`/unit`,`/unit/bulk`; POST `select-holding`→`/select-holding`, `branch`→`/organization/branch`, `create-holding`→`/create-holding`, `update-holding`, `product-units/defaults` | Bearer | LIVE (886 บรรทัด) | `workspace/[...workspacePath]/route.ts:44-101,126-178,213,415,471,496` |
| `/api/system-settings/[[...path]]` | GET/POST/PUT/DELETE | ตาม `kind` ของ slug: `main-crud`→`basePath` (เช่น `/organization/branch`, `/holding/employee`, `/holding/permission`, `/debtaccount/debtor`, `/unit`, `/product/group`, `/warehouse`); `company`→`/shop/{id}`,`/holding/{code}`; `restaurant-setting`→`/restaurant/settings*`; `atlas`→`/atlas/get|update|delete`; `ai-provider`→`/api/v1/ai-provider/*`; `copy-uat`→`/listsourceshops`,`/copymongouattodev`,`/previewcopymongo`; `permission-catalog`; `report`; `goapi-crud`/`atlas`/`ai-provider`/`copy-uat` ใช้ G | Bearer + JWT holdingcode (ถ้าตั้ง `JWT_SECRET_KEY`) + `/select-holding` เฉพาะ kind `atlas` | LIVE (670 บรรทัด) | `system-settings/[[...settingPath]]/route.ts:25-112,136-164,166-184,252-283,366-396`, `frontend/src/lib/system-setting-screens.ts:287-1769` |
| `/api/product/[[...path]]` | GET/POST/PUT/DELETE | M `/product?…`, `/product/{id}`, `/product/resync` | Bearer | LIVE | `product/[[...productPath]]/route.ts:15-24,43-51,133,152-164` |
| `/api/product-barcode/[[...path]]` | GET/POST/PUT/DELETE | M `/product/barcode*` | Bearer | LIVE | `product-barcode/[[...barcodePath]]/route.ts:36,79,92,102-114` |
| `/api/product-barcode/list` | POST | G `/api/product/barcode/list` | Bearer | LIVE | `product-barcode/list/route.ts:16,39` |
| `/api/product-barcode/master/[master]` | GET | M ตาม `MASTER_PATHS` 9 คีย์ — ใช้งานจริงแค่ `businesstype` (`/organization/business-type`) + `branch` (`/list-holding`); ที่เหลือ (`/product/group`, `/unit`, `/product/order-type`, `/debtaccount/*`, `/product`) เป็นท่อของจอ "รอพัฒนา" หลังถอด MongoDB (2026-09-23 ลบ 6 คีย์ที่ไม่มีผู้เรียก) | Bearer (ผ่าน proxyMainApiJson) | LIVE (บางคีย์) | `product-barcode/master/[master]/route.ts:18-32` |
| `/api/product-barcode/bom/[barcode]` | GET | M `/product/barcode/bom/{code}?itemcode=` | Bearer | LIVE | `product-barcode/bom/[barcode]/route.ts:10-42` |
| `/api/product-barcode/price-history/[barcode]` | GET | M `/product/barcode/price-history/{code}` | Bearer | LIVE | `product-barcode/price-history/[barcode]/route.ts:6-30` |
| `/api/product-price-history/[[...path]]` | GET/POST | M `/product/barcode/price-history/*` | Bearer | LIVE | `product-price-history/[[...historyPath]]/route.ts:14,24,35,47` |
| `/api/product-barcode/image` | POST | M `/goapi/image/upload` (category `products`) | Bearer | LIVE | `product-barcode/image/route.ts:9-13`, `frontend/src/lib/image-upload-proxy.ts:28,129` |
| `/api/product-barcode/video` | POST | M `/goapi/video/upload` (stream + จำกัดขนาด) | Bearer | LIVE | `product-barcode/video/route.ts:7-15`, `frontend/src/lib/image-upload-proxy.ts:47-79` |
| `/api/upload/image` | POST | M `/goapi/image/upload` (category `system-settings`, client ต้องส่ง category) | Bearer | LIVE | `upload/image/route.ts:3-8` |
| `/api/currency/[[...path]]` | GET/POST/PUT/DELETE | M `/currency/*` | Bearer | LIVE | `currency/[[...currencyPath]]/route.ts:14-60,80,91` |
| `/api/holding-member` | GET | M `/holding-member/list` (อ่านอย่างเดียว; เพิ่ม/ถอดผู้ดูแลทำที่ ตั้งค่าระบบและการเข้าถึง › บัญชีเข้าระบบ) | Bearer | LIVE | `holding-member/route.ts:1-17` |
| `/api/holding-users-import` | POST | M `/holding/users/import` | Bearer | LIVE | `holding-users-import/route.ts:8-12` |
| `/api/line-oa/user` | POST | G `/api/user/lineoa/link|profile` | Bearer + tenant JWT | LIVE | `line-oa/user/route.ts:20-21,32,47,63-68` |

(path ในคอลัมน์อ้างอิงย่อจาก `frontend/src/app/api/`)

## 6. ตาราง page → endpoint ที่ใช้ (grep `"/api/…"` ใน screen component)

| page/screen | `/api/*` ที่เรียก | อ้างอิง |
|---|---|---|
| `/` login-screen | `auth/login`, `auth/google/verify`, `auth/demo-login`, `auth/profile`, `backend/check` | `frontend/src/app/login-screen.tsx` (grep) |
| `/holding` | `workspace/*`, `holding-member` | `frontend/src/app/holding/holding-screen.tsx` |
| `/workspace` | `workspace/*`, `system-settings/*`, `auth/line/code`, `auth/line/link/status` | `frontend/src/app/workspace/workspace-screen.tsx` |
| `/menu` main-menu | `auth/profile`, `auth/sessions`, `system-settings/permissiongroup/me`, `auth/line/*` | `frontend/src/app/menu/main-menu-screen.tsx` |
| `/menu` product / product-set | `product`, `product/*`, `product-barcode/*`, `product-barcode/list`, `workspace/select-holding` | `frontend/src/app/menu/product-screen.tsx`, `product-set-screen.tsx` |
| `/menu` barcode / shelf / marketplace | `product-barcode/*`, `product-barcode/list` | `frontend/src/app/menu/product-barcode-screen.tsx`, `product-barcode-shelf-screen.tsx`, `marketplace-screen.tsx` |
| `/price_history` | `product-barcode/list`, `product-price-history` | `frontend/src/app/menu/product-price-history-screen.tsx` |
| `/[systemSetting]` | `system-settings/*`, `system-settings/branch`, `workspace/holdings`, `workspace/product-units/standard|defaults`, `upload/image`, `auth/profile/reset-password` (501), `holding-users-import` | `frontend/src/app/system-settings/system-settings-screen.tsx`, `bulk-user-import.tsx` |
| `/currency` | `currency`, `currency/*` | `frontend/src/app/currency/currency-screen.tsx` |
| `/line-oa` | `line-oa/user` | `frontend/src/app/line-oa/line-oa-link-screen.tsx` |
| lib/components ทั่วไป | `auth/refresh`, `auth/logout`, `language/{lang}`, `address/thailand`, `product-barcode/master/*` (ผ่าน `listMaster()` ที่ `master-picker.tsx` import), `upload/image` | `frontend/src/lib/client-auth-session.ts:121,205`, `frontend/src/lib/backend-language.ts:135`, `frontend/src/lib/thailand-addresses.ts`, `frontend/src/lib/product-barcode/api.ts:288`, `frontend/src/components/product-barcode/master-picker.tsx:8`, `frontend/src/components/system-settings/field-editors/image-upload-editor.tsx` |

## 7. `frontend/src/lib/**` — โมดูลสำคัญ

| module | หน้าที่ | อ้างอิง |
|---|---|---|
| `backend-url.ts` | normalize/validate URL, `serverMainApiBase/serverGoApiBase`, migrate URL เก่า (`192.168.2.202`, `dev./api.bcaicloud.com`) | `frontend/src/lib/backend-url.ts:56-109,117-144` |
| `workspace-api.ts` | helper proxy กลาง: `requireBearerToken`, `getBackendUrlFromRequest` (body > header `x-bc-backend-url` > query), `proxyMainApiJson` | `frontend/src/lib/workspace-api.ts:8-31,61` |
| `client-auth-session.ts` / `auth-session-server.ts` / `server-jwt.ts` | session ฝั่ง client / cookie ฝั่ง server / JWT verify | ดู §3 |
| `auth-bridge.ts` | URL ของ auth bridge (`BC_AUTH_BRIDGE_URL`) + `postMainApiAuth` timeout 15 วิ | `frontend/src/lib/auth-bridge.ts:6-10,76-87` |
| `theme-data.ts` | 10 พาเลต (`ban-chiang` default, `sukhothai-jade`, `ayutthaya-gold`, `lanna-teak`, `andaman-blue`, `siam-rose`, `violet-bloom`, `coral-sunset`, `citrus-lime`, `berry-magenta`) × light/dark; cookie/storage key `bc_theme`, `bc_color_theme`; ตัวแปรถูก inline บน `<html style>` ตั้งแต่ server | `frontend/src/lib/theme-data.ts:47-53,100-476,530-542`, `frontend/src/app/layout.tsx:48-73` |
| `font-data.ts` | 8 ฟอนต์ (inter default, noto-sans-thai, prompt, sarabun, kanit, ibm-plex-sans-thai, mitr, bai-jamjuree) key `bc_app_font`; layout preload Google Fonts ทั้ง 8 | `frontend/src/lib/font-data.ts:31-86`, `frontend/src/app/layout.tsx:77-88` |
| `i18n.ts` + `locales/*.json` | 12 ภาษา (th, en, cn, ja, ko, lo, my, km, vi, ms, id, fil) ไฟล์ละ 93 key เท่ากัน; `t(language,key)` | `frontend/src/lib/i18n.ts:19-31,67` |
| `backend-language*.ts` | dictionary จาก goapi `/api/language/{lang}` (preload fetch goapi ตรง; client refresh ผ่าน Next `/api/language/{lang}`); cookie `user_language`, `backend_url`; server อ่าน cookie ก่อน render | `frontend/src/lib/backend-language-preload.ts:6-7,25,41`, `frontend/src/lib/backend-language.ts:135`, `frontend/src/lib/backend-language-server.ts:16-19` |
| `system-setting-screens.ts` | config 31 slug (slug/route/kind/basePath/ฟิลด์; kind มี 9 แบบ) = source ของ `/[systemSetting]` | `frontend/src/lib/system-setting-screens.ts:78-88,287-1769,1771` |
| `menu-data.ts`, `menu-icons.ts`, `menu-usage.ts`, `user-shortcuts.ts` | โครงเมนู `MENU_SECTIONS` (transactions/reports …), icon ต่อ route, นับการใช้เมนู, ทางลัดผู้ใช้ใน storage; **เพิ่ม/ลบ 1 เมนู = แก้ 4 จุดพร้อมกันเสมอ** `menu-data.ts` → `backend/assets/language/languages.tsv` (13 คอลัมน์ ครบทุกภาษา) → `ROUTE_ICON_KEYS` ใน `menu-icons.ts` → จำนวนใน `menu-icons.test.ts` (ปัจจุบัน `toHaveLength(224)`) ไม่งั้น `frontend/src/lib/menu-data.test.ts` + `frontend/src/lib/menu-icons.test.ts` fail — ดู `docs/kms/19-menu-coverage-market-standard.md:61-62` และ `.agents/skills/ui-scale-polish/SKILL.md` §8 | `frontend/src/lib/menu-data.ts:48-155` |
| `permission-actions.ts`, `use-screen-actions.ts` | สิทธิ์ต่อจอ (`hasScreenAccess/hasScreenAction`) | `frontend/src/lib/permission-actions.ts` |
| `date-time.ts`, `thailand-addresses.ts`, `thai-branch-code.ts`, `holding-code.ts`, `business-code.ts`, `currency-presets.ts` | พ.ศ./timezone, ที่อยู่ไทย, รหัสสาขาสรรพากร (`00000` = สำนักงานใหญ่), validate รหัส | `frontend/src/lib/thai-branch-code.ts`, `frontend/src/lib/holding-code.ts` |
| `image-upload-proxy.ts`, `logo-thumb.ts` | proxy multipart ไป goapi, URI thumbnail | `frontend/src/lib/image-upload-proxy.ts:24-129` |
| `toast.ts` | `notify/toast/pushNotice` → `ToastViewport` | `frontend/src/lib/toast.ts` |

## 8. Components inventory

- `components/ui/*` (shadcn-style: badge, button, card, confirm-dialog, date-time-field, dropdown-menu, input, numeric-input, select, skeleton, table, textarea) — `frontend/components.json`
- `components/system-settings/*` — field editors ของจอตั้งค่า (image upload, permission matrix/sets, role-screen-matrix, thailand-address, holding-scope, work-day, copy-uat)
- `components/product-barcode/*` — barcode-form, ean13-barcode, master-picker, names/addresses editor, business-image-editor
- อื่น ๆ: `authenticated-image.tsx` (โหลดรูปพร้อม Bearer), `leaflet-map-picker.tsx` + `map-picker-dialog.tsx`, `logo-avatar.tsx`, `toast-viewport.tsx`
- ใน `src/app/`: `theme-toggle.tsx`, `font-picker.tsx`, `language-dialog.tsx`, `zoom-control.tsx`, `app-header-controls.tsx`, `motion-config.tsx`, `shared/motion-variants.ts`, `shared/skeleton-card.tsx`, `system-settings/*-tree-view.tsx` (5 ไฟล์), `menu/tab-product-*.tsx` (12 ไฟล์ — นับจาก `ls frontend/src/app/menu`)

## 9. กฎ UX/UI ที่บังคับใช้ (สรุป non-negotiables)

ที่มา: `AGENTS.md` (กฎ "คนไทย 40+" และ "พรีเมี่ยม" + กฎ upgrade skill 2026-09-06) และ `.agents/skills/ui-scale-polish/SKILL.md:13-49,196-214`
1. ข้อความตัดสินใจ ≥ 0.9rem, Thai-first, `line-height ≥ 1.45`, ปุ่มหลัก ≥ 2.6em, ห้าม icon เปล่าใน action สำคัญ, contrast WCAG AA, ห้ามสื่อสถานะด้วยสีเดียว, dirty guard + dialog ไทยก่อนทำลาย, feedback ไทยทุก action, motion ≤ 300ms + `prefers-reduced-motion` (`SKILL.md:23-36`)
2. สีทุกจุด derive จาก `--primary` ผ่าน `color-mix` ห้าม hard-code (10 พาเลต × dark); ทดสอบ dark โดย **กดปุ่มสลับธีมจริง** เพราะตัวแปรถูก inline บน `<html>` (`frontend/src/app/layout.tsx:59-72`); popover ห้ามโดน `overflow:hidden` (`SKILL.md:39-47`)
3. CSS = skin block ต่อท้าย `globals.css` มี comment วันที่ + selector prefix, ห้ามแตะ type scale เดิม (`SKILL.md:196-200`); ตัวอย่าง block ล่าสุด 2026-09-02…09-07 (`frontend/src/app/globals.css:7500,7848,8247,8429,8487`)
4. Checklist ก่อนบอกเสร็จ: viewport **1600 / 1280 / 1024 / 768-portrait** (บังคับตาม `AGENTS.md:61` กฎพรีเมี่ยมข้อ 8 — เป้าหมาย iPad ขึ้นไป), dual theme (กดปุ่มสลับธีมจริง), Thai glyph ไม่ทับ, no clip, popover safety, console สะอาด, `npm run typecheck` + unit test ผ่าน (`SKILL.md:204-214`); ทุกงาน UI ต้อง append section ใน SKILL.md และ commit พร้อมโค้ด (`SKILL.md:13-21`) — `SKILL.md:207` แก้ให้ตรงกับ `AGENTS.md:61` แล้ว 2026-09-09 (ห้ามแก้กลับเป็น 1280/1920/2560)
5. Gotcha: reset แบบไม่มี layer `button, input, select, textarea { font: inherit; }` ใน `globals.css` ทับ Tailwind `text-*` บนปุ่ม → ใช้ `text-xs!` (important) หรือ inline style (memory `tailwind-button-font-size-gotcha`; ตรวจแล้ว `frontend/src/app/globals.css:364-369` — ไม่มี `font-size: inherit` ตรง ๆ แต่ `font: inherit` ให้ผลเดียวกัน)

## 10. `globals.css` mixed EOL gotcha

- ไฟล์ยาว 8,575 บรรทัด (`wc -l`) / 229,204 bytes; นับด้วย node ได้ **CRLF 7600, LF 975, CR 0** (ตรวจ 2026-09-09) → EOL ผสมจริง
- ห้ามใช้ Edit/Write tool ทั้งไฟล์ (จะ normalize EOL ทำให้ diff ทั้งไฟล์) — แก้ด้วย Node byte-preserving patch ต่อท้ายไฟล์ตามกฎ AGENTS.md ข้อ 9 และ memory `globals-css-mixed-eol-gotcha`
- ไฟล์แปลก: `frontend/src/app/.dialog-header` (1 บรรทัด เนื้อหา `8134:section[role=dialog].h-full.bg-card {` ดูเหมือน grep output ที่ commit หลุด) — ควรลบ (รอลุงจืดตัดสิน)

## 11. Tests

| ชุด | config | จำนวน | วิธี login | อ้างอิง |
|---|---|---|---|---|
| Unit (vitest, env node) | `frontend/vitest.config.ts:3-12` include `src/**/*.test.ts` | 45 ไฟล์ (lib, api route, security tests เช่น `login-screen.security.test.ts`, `holding-screen.security.test.ts`, `main-menu-password.security.test.ts`, `menu/main-menu-sidebar-resize.test.ts`) | — | `frontend/src/app/login-screen.security.test.ts:7-21` |
| e2e ใน frontend | `frontend/playwright.config.ts` testDir `./e2e`, baseURL `E2E_BASE_URL` (default `localhost:3000`), ไม่ start server เอง | 14 spec (product/barcode/BOM/warehouse/category-list/group-tree/set/brand/trade-partners/job-costing-channel CRUD, barcode company-scope + sample-data, dev-login, Thai address cascade) | spec คลิกปุ่ม "เข้าทดสอบระบบ (Dev Login)" → `POST /api/auth/dev-login` — **แต่ปุ่มนี้ไม่มีใน `login-screen.tsx` แล้ว** (ดูช่องว่าง) | `frontend/playwright.config.ts:6-17`, `frontend/e2e/dev-login.spec.ts:6-21` |
| UAT ที่ root | `playwright.config.ts` testDir `./tests`, baseURL `PW_BASE_URL` (default `127.0.0.1:3000`), project `setup` สร้าง `.auth/user.json` แล้ว chromium reuse | 13 spec + `auth.setup.ts` | คลิกปุ่ม /Dev Login/ → `/holding` เลือก "บ้านเชียง" → `/workspace` เลือกสาขา TST03 / 00001 (ปุ่ม Dev Login ไม่มีในหน้า login แล้ว — ดูช่องว่าง) | `playwright.config.ts:17,32,42-54`, `tests/auth.setup.ts:14-25` |
| ตรวจอัตโนมัติ | ไม่มี CI ฝั่ง GitHub แล้ว (ลบ 2026-09-09) — ใช้ `sh tools/verify.sh frontend`: `npm run lint` → `npm run typecheck` → `npm test -- --run` ด้วย `BCAI_LOCAL_BACKEND_URL=http://localhost:8888`, และ `frontend-build` = `next build` แยก target | — | — | `tools/verify.sh:64-84` |

## 12. รัน dev / build / deploy และ env (ชื่อเท่านั้น)

- Dev: `cd frontend && npm run dev` (root มี `npm run dev:frontend` = port 3001, `package.json:13`); ต้องมี `BCAI_LOCAL_BACKEND_URL` ไม่งั้น `next.config.ts` throw (`frontend/next.config.ts:20-22`); ตรวจ runtime 2026-09-07: `curl localhost:3000` → 200, container `mainapi` up ที่ 8888 (read-only)
- Env ที่โค้ดอ่าน: `BCAI_LOCAL_BACKEND_URL`, `JWT_SECRET_KEY`, `BCAI_DEV_LOGIN_ENABLED`, `BCAI_DEV_LOGIN_SECRET`, `BCAI_DEV_LOGIN_BACKEND_URL`, `GOOGLE_CLIENT_ID`, `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `BC_AUTH_BRIDGE_URL`, `NODE_ENV`, `E2E_BASE_URL`, `E2E_BROWSER_CHANNEL`, `PW_BASE_URL`, `CI` (grep `process.env.*` ใน `frontend/src`, `next.config.ts`, `playwright.config.ts`); `BCAI_DEMO_LOGIN_ENABLED` เป็นของ backend (`frontend/src/app/api/auth/demo-login/route.ts:9`)
- Docker: `frontend/Dockerfile` multi-stage `node:24.18.0-alpine`, build-args `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `BCAI_LOCAL_BACKEND_URL`, copy `manual/` + `.next` แล้ว `npm run start` (ไม่ใช้ standalone) (`frontend/Dockerfile:2,25-32,44-57`)
- Prod compose: service `frontend` image `${FRONTEND_IMAGE}` + `env_file /etc/bcai-account/frontend.env` (`deploy/account/compose.yml:309-311`)
- ~~ไฟล์ log ถูก track ใน git~~ **เลิก track แล้ว 2026-09-09** (`git rm --cached frontend/.next-dev.log frontend/.next-dev.err.log`) — root `.gitignore:53` (`*.log`) คุมอยู่แล้ว ส่วน `frontend/.gitignore:9-12` ยัง ignore เฉพาะ `npm-debug.log*`, `yarn-debug.log*`, `yarn-error.log*`, `pnpm-debug.log*` (`git ls-files frontend | grep .log`, `frontend/.gitignore:9-12`)

## 13. ตรวจว่าแถวไหนใน `languages.tsv` "ยังไม่ได้แปล" (ตรวจ 2026-09-16)

- อาการที่ผู้ใช้เห็น: เลือกภาษาญี่ปุ่นแล้วเมนูบางรายการยังเป็นอังกฤษ (เช่น "Cancel Purchase Order") ทั้งที่แถวนั้นมีครบ 13 คอลัมน์ — เพราะตอนสร้างแถวใส่ข้อความอังกฤษซ้ำลงคอลัมน์ภาษาอื่น ตัวตรวจ "13 คอลัมน์" จึงไม่จับ
- **นิยามที่ใช้ตรวจ: คอลัมน์ภาษาใดก็ตามที่มีค่าเท่ากับคอลัมน์ `en` = ยังไม่ได้แปล** (ยกเว้นคำที่ภาษานั้นใช้คำอังกฤษจริง เช่น `status`/`unit`/`menu`/`total`/`ok` ใน ms-id-fil)
- คีย์ของเมนูมาจาก 3 ที่ ต้องดึงให้ครบทั้งสาม ไม่งั้นนับพลาด: `ml("<group-id>", …)` (หัวข้อกลุ่มย่อย 21 รายการ), `tx("<id>", …, languageKey?)` (รายการเมนู) และ `key: "…"` (หมวดหลัก) — `frontend/src/lib/menu-data.ts:34,159`
- คีย์ที่จอจริงใช้ = grep `backendText(dict, "key"` + `tr("key"` + `languageKey:` + `counterpartyKey:` ใน `frontend/src` (2026-09-16 ได้ 1,159 คีย์) — ตรวจเฉพาะชุดนี้ก่อน คุ้มกว่าไล่ทั้ง 6,041 แถว
- คีย์ที่โค้ดเรียกแต่**ไม่มีแถวใน tsv** = `Text()` คืน key ดิบขึ้นจอ ต้องเติมแถวทันที (รอบนี้เจอ 6 คีย์จาก `workspace-screen.tsx`)
- ตรวจ Thai leakage ด้วย regex `[฀-฾เ-๿]` **ไม่ใช่ `[฀-๿]`** — ช่วงเต็มกินสัญลักษณ์บาท `฿` (U+0E3F) ทำให้แถวที่ถูกต้องอย่าง `product_set_price_delta` ("加/减 (฿):") ถูกตีว่าเป็นไทยหลุด
- **กับดักคีย์ชนกัน:** `getBarcodeText()` (`frontend/src/lib/product-barcode/language.ts:1115-1138`) ใช้ Proxy ไล่หา `barcode_<snake>` → `<snake>` → `<prop>` ตามลำดับ ถ้าไม่มีแถว `barcode_*` มันจะหยิบคีย์กลางที่ชื่อพ้องกันมาแทน — เดิม `text.title` จึงขึ้นว่า "dede POS" (แถว `title`) และ `text.filter` ขึ้น "ค้นหา" (แถว `filter`) แก้ด้วยการสร้างแถวเฉพาะของจอครบทั้ง 467 property (2026-09-16); helper ตัวใดที่ fallback เป็นชื่อคีย์กว้าง ๆ ต้องตรวจแบบเดียวกัน
- **จอตั้งค่า:** `fieldLabel()` (`frontend/src/components/system-settings/utils.ts:599-612`) แมป `slug.field` → คีย์ผ่าน `fieldBackendKeys` แล้ว **fallback เป็น `field.label.en`** เมื่อไม่มีแถว → คีย์ที่ไม่มีแถว = ผู้ใช้ทุกภาษาที่ไม่ใช่ไทยเห็นอังกฤษ (2026-09-16 เจอ 26 จุด); และเมื่อแถวมีอยู่ **ข้อความในแถวชนะป้ายในโค้ดเสมอ** แก้ `label.th` อย่างเดียวจึงไม่เห็นผล — มี guard test `frontend/src/components/system-settings/field-backend-keys.test.ts` คุมทั้งสองเรื่องแล้ว
- เบราว์เซอร์เก็บ dictionary ไว้ใน `localStorage` คีย์ `bc_backend_language::<backendUrl>::<lang>` แล้วโหลดใหม่ทับทุกครั้ง (`cache: "no-store"`, `backend-language.ts:116-147`) → หลังแก้ tsv ต้องรีเฟรชหนึ่งครั้งถึงเห็นของใหม่ ไม่ต้องล้าง cache เอง
- ทดสอบจริง **ไม่ต้อง restart แล้ว** (2026-09-16): สั่งครั้งเดียว `cd backend && docker compose -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.dev.yml up -d --no-deps mainapi` เพื่อให้ container ได้ `LANGUAGE_RELOAD_ENABLED=true` + mount `assets/language` เข้าไป จากนั้นแก้ไฟล์แล้ว `curl localhost:8888/goapi/api/language/ja` เห็นผลทันที (`language.go:120` เทียบ mtime/size ก่อนอ่านซ้ำ — ถ้าไม่มีธงนี้จะอ่านครั้งเดียวตลอดอายุ process). วิธีเก่า `docker cp … && docker restart mainapi` ยังใช้ได้แต่ **restart = ผู้ใช้ถูกเตะออก ต้องล็อกอินใหม่ 5 คลิกทุกรอบ**
- **กับดักที่เงียบกว่า: เรียก `getBarcodeText(language)` โดยไม่ส่ง dictionary** — ตัวฟังก์ชันคืนตาราง th/en แบบ static ทันทีโดยไม่ error (`language.ts:1122-1125`) ผู้ใช้อีก 10 ภาษาจึงเห็นอังกฤษแบบไม่มีสัญญาณเตือน; จอลูก/แท็บที่ไม่มี dictionary ให้ใช้ hook `useBarcodeText(language)` (`frontend/src/components/product-barcode/use-barcode-text.ts`) ซึ่งดึง dictionary จาก context ให้เอง แล้วครอบจอด้วย `<BackendTextProvider dictionary={backendLanguage}>` ที่ root ของจอ (`product-screen.tsx`, `product-set-screen.tsx`, `product-barcode-screen.tsx`)
- **ไล่ hard-code ต้องค้น 3 แบบ ไม่ใช่แบบเดียว**: `language === "th" ?`, `lang === "th" ?` และ `const isThai = normalizeLanguage(language) === "th"` แล้ว `isThai ? … : …` — การ audit ที่ grep แค่ `=== "th" ?` จะนับไม่ครบ (รอบ 2026-09-16 พบ `isThai` อีก 16 จุดหลังคิดว่าเก็บหมดแล้ว) และ ternary ที่ซ่อนใน template literal (`` `กรอก${label}` ``) grep ธรรมดาก็ไม่เจอ ต้องหาแยกอีกรอบ
- **แถวที่มีตัวเลข/ชื่อแปรผัน** ใช้ placeholder `{0}`/`{1}` ในทุกคอลัมน์แล้วให้โค้ด `.replace("{0}", String(n))` (`backendText`/`tr` ไม่แทนค่าให้เอง) — ตัวอย่าง `barcode_video_format_hint`, `barcode_enter_value_for_language`, `general-ledger.ts:241`
- **แคตตาล็อกจอแท็บงาน (`erp-operations.ts`, `erp-tools.ts`, `erp-reports.ts`, `thai-tax.ts`)** เก็บข้อความเป็น literal `{ th, en }` ข้าง config ของ route — 346 จุด จึงรองรับแค่ 2 ใน 12 ภาษา; 2026-09-16 เพิ่ม `const catalogKeys` (map `<code>.<part>` → คีย์ tsv) ท้ายไฟล์ + resolver (`operationsText`, `toolText`/`toolStepText`, `reportText`/`reportColumnLabel`, `taxText`) ที่เรียก `catalogText()` (`frontend/src/lib/catalog-text.ts`) — literal เดิมเหลือเป็น fallback ตอน dictionary ยังไม่โหลด; guard test `frontend/src/lib/catalog-language-keys.test.ts`
- **ตาราง message นิ่ง ๆ ในจอ** (`MESSAGES`, `RESULT_MESSAGES`, `ERROR_MESSAGES`) เดิมเป็น `{ th, en }` + helper `messageText(key, language)` — เปลี่ยน value เป็น `{ key, th }` แล้วให้ helper รับ `tr` แทน `language` (`messageText(key, tr)`) คีย์ซ้ำข้ามจอใช้แถวเดียวกันได้ (`holding_required`, `load_data_failed`, `ops_msg_unable_to_connect_to_the`)
- **กฎการใช้แถวเดิมซ้ำ: ต้องตรงทั้งไทย*และ*อังกฤษ** — จับคู่ด้วยไทยอย่างเดียวจะหยิบแถวที่แปลไว้คนละความหมาย เช่น "ทั้งหมด" ในตารางสิทธิ์แปลว่า *All* แต่แถว `common_total` แปลว่า *Total* (ja = 合計 ไม่ใช่ すべて) — 2026-09-16 เจอ 18 จุดในรอบกวาดจอ; ถ้าไทยตรงแต่อังกฤษเป็นคำพ้อง (Employees/Employee) ใช้ซ้ำได้ เพราะภาษาอื่นแปลตามไทย
- **จอที่มีตาราง `{ th: {...}, en: {...} }` ของตัวเอง** (จอพิมพ์ป้ายสินค้า, จอประวัติราคา, จอโครงสร้างข้อมูล) ใช้ `resolveTextTable(text, <screen>TextKeys, language, dictionary)` (`frontend/src/lib/catalog-text.ts`) แปลงทั้งตารางครั้งเดียวใน `useMemo` โค้ดที่เหลือยังเรียก `text.title` เหมือนเดิม
- **`useBackendLanguage(language, undefined)` ไม่ยิง API เลย คืน `{}`** (`backend-language.ts:119-122`) — component ที่ไม่มี `backendUrl` ต้องใช้ `useBackendDictionary()` จาก provider ของจอแม่แทน ไม่ใช่เรียก hook แล้วส่ง undefined (เจอตอนแปลงจอโครงสร้างข้อมูลกับตัวเลือกฟอนต์ 2026-09-16)
- **แถวที่ "ครบ 13 คอลัมน์ แต่เป็นอังกฤษทั้งแถว"** — 2026-09-16 นับได้ **728 แถว** ที่ทั้ง 10 ภาษาที่ไม่ใช่ th/en มีค่าเท่ากับคอลัมน์ `en` เป๊ะ (611 แถวเป็นวลีหลายคำ = รูโหว่จริง ไม่ใช่ตัวย่ออย่าง PO/VAT/CSV) เช่น `print_report` ที่ขึ้น "Print Report" บนจอญี่ปุ่น — ตรวจด้วย `[3..12].every(i => cells[i] === cells[2])` สคริปต์ `scratch/english-only.js` (ข้ามแถวที่ `th === en` = ยี่ห้อ/รหัส/หน่วย 38 แถว ซึ่งถูกต้องแล้ว). **แปลไปแล้ว 683 แถวในรอบเดียว → เหลือ 11** ดู README Activity Log 2026-09-16
- **11 แถวที่แปลไม่ได้เพราะคอลัมน์ `th` ว่าง** (`item_display_all|barcode|price|sku`, `kitchen_secondary`, `kitchenprinter`, `spare` (en = Kitchen name), `stand`, `xqty`, `xunit`) — ต้นฉบับคือภาษาไทย เมื่อไม่มีไทยก็ห้ามแต่งขึ้นเอง (NO MAGIC) และผู้ใช้ไทยเห็นเป็นค่าว่างอยู่แล้ว = bug คนละตัว; อีกแถวคือ `process ` ที่ **key มีช่องว่างต่อท้าย** ซึ่ง backend ข้ามไม่เสิร์ฟอยู่แล้ว (เหมือน `add `) — รอลุงจืดบอกคำไทยที่ถูกต้อง ห้ามเดา
- **`useBackendLanguage(language, "")` ก็พังเงียบแบบเดียวกับ `undefined`** — `""` เป็น falsy จึงคืน `{}` ไม่ยิง API เลย (`frontend/src/app/asset/fixed-assets-screen.tsx:31` เดิม) ผลคือ `backendText(dictionary, key, fallback)` คืน fallback ไทยเสมอ ทุกภาษา; จอที่อยู่ใต้ `WorkTabPanel` มี `BackendTextProvider` ครอบอยู่แล้ว (`main-menu-screen.tsx:1481`) ให้ใช้ `useBackendText()` แทน
- **เครื่องมือกวาดไทยฝังโค้ดทีละจอ: `node tools/i18n-sweep.mjs <scan|plan|apply> <file...> [--prefix=xx]`** (ถาวรแล้ว 2026-09-16 — เดิมเขียนสคริปต์ใหม่ทุกจอใน `scratch/`) ไฟล์ทำงานอยู่ใน `.i18n-sweep/` (git ignore)
  - `scan` แยกข้อความเป็น 3 ทรง: string literal / JSX text node / template literal ที่มี `${}` (ทรงที่สามแก้มือด้วย `{0}` เสมอ ห้ามตัดประโยค) — ข้ามฝั่ง `th:` ของ literal สองภาษา เพราะ helper ที่อ่านมันแปลให้อยู่แล้ว
  - `plan` เสนอ key: ใช้แถวเดิมซ้ำเมื่อไทยตรง **แต่ต้องอ่าน en ของผู้สมัครทุกตัวด้วยตาก่อน** (พิมพ์ให้ครบทุกตัวในบรรทัดเดียว) แล้วแก้ `.i18n-sweep/plan.json` ได้ — การรัน `plan` ซ้ำ **ไม่ทับ** ของที่แก้มือแล้ว (ใช้ `--fresh` ถ้าต้องการเริ่มใหม่)
  - `apply` แก้โค้ดเป็น `tr("key", "ไทย")` แล้ว append แถวใหม่ + ตรวจตารางให้ (13 คอลัมน์ / ไม่มี key ซ้ำ)
  - กับดักที่เจอจริง: (1) แก้จำนวนบรรทัดก่อน apply แล้วเลขบรรทัดที่สแกนไว้เลื่อนทั้งไฟล์ — เครื่องมือจึงสแกนใหม่เองใน `apply` (2) `placeholder="ไทย"` ใน JSX ต้องใส่ปีกกา `{tr(...)}` ไม่งั้น TS1145 (3) ข้อความที่อยู่นอกตัว component (helper/`TestBadge`) ไม่มี `tr` — ต้องรับป้ายเป็น argument/prop จากผู้เรียก ตามกฎ AGENTS.md ข้อ 2 (4) `settings-screen.tsx` มี EOL เป็น `\r\r\n` ที่ Edit tool หาไม่เจอ — patch ด้วย index ของ substring แทน anchor หลายบรรทัด
- **ช่วง dev ไทยอย่างเดียว: แถวใหม่ใส่ไทยทุกคอลัมน์** (AGENTS.md 2026-09-16) แถวจึงสมบูรณ์ (backend ไม่ข้าม) และหาแถวที่รอแปลได้ด้วยเงื่อนไข `cells.slice(2).every(v => v === cells[1])` — คู่กับตัวตรวจเดิม "ทุกคอลัมน์ == en" ที่ใช้หาแถวรุ่นเก่า
- **ให้ DeepSeek ตั้งชื่อ key ได้ แต่ต้องอ่านความหมายซ้ำ** — "ยังไม่ผ่าน" (ยังไม่ผ่านรายการเข้า GL) ถูกตั้งเป็น `fa_not_yet_approved` + แปลว่า "Not Yet Approved / まだ承認されていません" ทั้ง 10 ภาษา ต้องแก้เป็น `fa_not_posted` (Not posted / 未転記) — คำว่า "ผ่าน/ผ่านรายการ" ในบริบทบัญชีคือ post ไม่ใช่ approve
- ตรวจว่า dictionary ของจอครบไหม: เทียบ property ทุกตัวของ `barcodeTextTh` กับคีย์ใน tsv (`barcode_<snake>` → `<snake>` → `<prop>`) — สคริปต์แบบใช้ครั้งเดียวอยู่ใน `scratch/barcode-dict-gap.js`; 2026-09-16 เหลือ 3 property ที่ไม่มีแถวเพราะ**ไม่มีโค้ดเรียกใช้แล้ว** (`inheritedInfoDetail`, `noProductDetailFound`, `mkNotLinked`) — dead code ที่รอลุงจืดตัดสินใจ ห้ามลบเงียบ ๆ

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ

- **ยังไม่ตรวจ** ว่า backend มี route ครบตามที่ BFF เรียกทุกตัว (เช่น `/holding/users/import`, `/product/resync`, `/sessions/active-count`, `/unit/bulk`, `/api/user/lineoa/*`) — ต้อง cross-check กับ `backend/main.go` exceptShopPath และ module routes (บทความ backend)
- **ยังไม่ตรวจ** `BC_AUTH_BRIDGE_URL` ชี้ไปที่บริการใด/ยังรันอยู่ไหม — ถ้าไม่มี flow ผูก LINE (`/api/auth/line/code`, `link/status`) จะตอบ error ไทย "ระบบยังไม่ได้ตั้งค่า Auth bridge" (`frontend/src/lib/auth-bridge.ts:8`)
- ขัดกันจริงในโค้ด: `frontend/src/app/login-screen.tsx:304-309,649-657` มีเฉพาะปุ่ม "ทดลองใช้ระบบ (Demo)" (comment บอกว่าแทนปุ่ม Dev Login เดิม) และ `frontend/src/app/login-screen.security.test.ts:8-12` บังคับว่าห้ามมี `/api/auth/dev-login` ในหน้า login — แต่ `frontend/e2e/dev-login.spec.ts:8` และ `tests/auth.setup.ts:16` ยังหาปุ่ม "Dev Login" อยู่ → e2e/UAT ชุดนี้น่าจะไม่ผ่านตั้งแต่ขั้น login (**ยังไม่รันยืนยัน**); ถามลุงจืดว่าจะย้าย e2e ไป Demo หรือคืนปุ่ม Dev
- (ปิดแล้ว 2026-09-07) skill ส่วนตัวอยู่ที่ `.agents/skills/ui-scale-polish/SKILL.md` (271 บรรทัด, ตรวจ 2026-09-09) ตามกฎใน `AGENTS.md:20-25` ซึ่งชี้ path นี้ถูกต้องแล้ว — โฟลเดอร์ `.agents/skills/` ถูกย้ายออกไปแล้ว (`git ls-files .agents` ว่างเป็นเรื่องปกติ ไม่ใช่ bug) ห้ามสร้างกลับ และ **ห้ามแก้ path ใน AGENTS.md อีก**
- (ปิดแล้ว 2026-09-24) ลบจอ `/settings` + `setup-config.ts` + BFF `/api/setup/*` และ `/api/storage/health` — ไม่ตรวจ session, รับรหัสว่าง/`12345`/`admin`, เปิด SSRF ให้ใครก็ได้บน URL สาธารณะ และ config ที่บันทึกไม่มี service ใดอ่าน; guard test `frontend/src/app/api/removed-setup-api.test.ts` — ดู `bugs/2026-09-24-unauthenticated-setup-api-ssrf.md`
- `system-settings-screen.tsx` 16,600 บรรทัดในไฟล์เดียว — ไม่ได้อ่านทั้งไฟล์ (ใช้ grep) จึง**ยังไม่ตรวจ** พฤติกรรมรายจอ 31 slug
- ไม่ได้รัน `npm test`/`typecheck`/playwright ในรอบนี้ (อ่านโค้ดอย่างเดียว) — สถานะผ่าน/ตกล่าสุด**ยังไม่ตรวจ**
- คำถาม: ไฟล์ `frontend/src/app/.dialog-header` และ `.next-dev*.log` ที่ถูก track ลบได้ไหม (R2 แต่รอคำสั่ง)
