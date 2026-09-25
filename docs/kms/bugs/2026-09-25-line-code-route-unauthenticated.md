---
date: 2026-09-25
severity: medium
component: [frontend]
tags: [bc-account, security, auth, line]
fixed: true
---

# Symptom

BFF route `frontend/src/app/api/auth/line/code/route.ts` (POST) ไม่ตรวจ session เลย — ใครก็ตามที่เข้าถึง URL สาธารณะได้สั่ง `POST /api/auth/line/code` แบบไม่มี token แล้ว frontend server จะไปเรียก `${BC_AUTH_BRIDGE_URL}/api/login?action=create` เพื่อออก LINE bridge code + `loginUrl` ให้ทันที (สร้าง code ได้ไม่จำกัดจำนวน, ใช้ server เราเป็นทางผ่านไปยิง bridge).

ผู้เรียกจริงมีแค่ปุ่ม "เชื่อมต่อ LINE" (ผูกบัญชี LINE กับผู้ใช้ที่ login อยู่) ใน `frontend/src/app/menu/main-menu-screen.tsx:825` และ `frontend/src/app/workspace/workspace-screen.tsx:784` — ทั้งคู่เรียก `authFetch` โดยไม่แนบ `Authorization` จึงออกไปแบบ anonymous.

บน prod ตอนนี้ยังไม่ถูกใช้ประโยชน์ได้จริง เพราะ `frontend.env` ไม่มี `BC_AUTH_BRIDGE_URL` (ADR `decisions/2026-09-24-remove-legacy-login-flows.md`) → `getAuthBridgeUrl()` throw แล้ว route ตอบ 504 — แต่ถ้าวันไหนตั้งค่า bridge ช่องนี้จะเปิดทันที.

## Root Cause

- Route นี้มาพร้อม import แรกของ repo (`0902ddb8`, 2026-05-20) และไม่เคยมีการตรวจ `Authorization` เลย; คู่ของมัน `api/auth/line/link/status` ตรวจ Bearer แล้วส่งต่อ `PUT /profile/link-line` แต่ขา "ออก code" ถูกลืม.
- Commit `32f02c71` (ลบ LINE **login**) เก็บ flow **ผูกบัญชี** LINE ไว้ตามตั้งใจ (UI ยังเรียกใช้ รอลุงจืดตัดสิน) จึงไม่ได้แตะ route นี้.
- `requireBearerToken` (`frontend/src/lib/workspace-api.ts:8-14`) ตรวจแค่รูปแบบ header — route อื่นปลอดภัยเพราะส่ง token ต่อให้ mainapi ตรวจ แต่ route นี้ไม่เรียก mainapi เลย การใส่ `requireBearerToken` อย่างเดียวจึงยังโดน `Bearer x` ปลอมผ่านได้.

## Fix

- `frontend/src/app/api/auth/line/code/route.ts`: เรียก `requireBearerToken` ก่อน แล้วตรวจ session จริงกับ mainapi `GET ${serverMainApiBase()}/verify-token` (middleware เดียวกับ API ที่ป้องกันทุกตัว: token เป็นค่าสุ่มทึบ ไม่ใช่ JWT — ค้นใน `cache_entries` แล้วตรวจ session/สิทธิ์ที่ยังใช้งานอยู่ ไม่มีการตรวจลายเซ็น (`backend/pkg/microservice/auth.go` `MWFuncMixShop`); `/verify-token` อยู่ใน `exceptShopPaths` ของ `backend/main.go` จึงใช้ได้ทั้งจอ workspace และเมนูหลัก) — mainapi ตอบ 401/403 → BFF ตอบ 401 body เดียวกับ `requireBearerToken` (ให้ `authFetch` refresh token แล้วลองใหม่ 1 ครั้ง); mainapi ล่ม → 502/504 และ**ไม่**ยิง bridge.
- Route ไม่รับ input ใด ๆ: bridge URL มาจาก env `BC_AUTH_BRIDGE_URL` เท่านั้น, ไม่อ่าน/ไม่ส่งต่อ request body, และไม่ส่ง token ของผู้ใช้ไปให้ bridge; ไม่มี log token.
- ผู้เรียกทั้งสองจอแนบ `Authorization: Bearer ${auth.token}` แล้ว (`main-menu-screen.tsx:825-828`, `workspace-screen.tsx:784-787`).
- ไม่เลือกลบทั้ง flow: ADR 2026-09-24 ระบุว่าการลบการผูกบัญชี LINE ต้องให้ลุงจืดตัดสินก่อน.

## Regression Test

- `frontend/src/app/api/auth/line/code/route.test.ts` — ไม่มี Authorization → 401 และไม่มี outbound fetch; Bearer ที่ mainapi ปฏิเสธ → 401 และไม่ยิง bridge; session ถูกต้อง → เรียก `/verify-token` แล้ว bridge `?action=create` ครั้งเดียว โดยไม่ส่ง Authorization/body ให้ bridge; body ที่มี `bridgeUrl` ปลอม (และ JSON เสีย) ถูกเมิน; mainapi ติดต่อไม่ได้ → 504 ไม่ออก code. เทสต์ 4 ใน 5 ข้อ fail กับโค้ดเดิม.
- ~~ยังค้าง: code ที่ออกไม่ผูกกับผู้ใช้ที่ขอ~~ — แก้แล้วในหัวข้อ Follow-up ด้านล่าง (ลุงจืดเลือกเก็บ flow แล้วแก้ ไม่ลบ).

# Follow-up 2026-09-25 — `link/status` ตรวจ session + ผูก code กับผู้ออก

## Symptom

- (a) `frontend/src/app/api/auth/line/link/status/route.ts` ตรวจแค่รูปแบบ header (`requireBearerToken`) แล้วยิง bridge เลย — `Bearer x` ปลอมไปถึง bridge ได้.
- (b) code ไม่ผูกกับผู้ใช้ที่ออก: session ใดก็ poll `link/status` ด้วย code ที่คนอื่นออกได้ และเมื่อเจ้าของ code ยืนยันใน LINE แล้ว ผู้ที่ poll ก่อนจะได้ LINE ของคนนั้นไปผูกกับบัญชีตัวเอง (mainapi `PUT /profile/link-line` ผูก `lineuserid` กับผู้เรียกทันที).

## Root Cause

- ตอนแก้ `line/code` รอบแรกตั้งใจแก้แค่ขาออก code; `link/status` เห็น token แค่ตอน `PUT` ท้ายสุด (หลังถาม bridge แล้ว).
- ทั้ง flow ไม่มีที่ใดจำว่า code ไหนออกให้ใคร — bridge ตอบตาม code อย่างเดียว.

## Fix

- Helper กลาง `frontend/src/lib/line-link-session.ts` (แทน `verifySession` ที่เคยอยู่ใน `line/code/route.ts`): `verifyLineSession` (`GET /verify-token`), `bindLineLinkCode` (`POST /profile/link-line/code`), `checkLineLinkCode` (`POST /profile/link-line/code/check`) — mainapi 401/403 → BFF 401 body เดียวกับ `requireBearerToken`; code ถูกปฏิเสธ (400/404/409) → ส่ง status + ข้อความที่ mainapi แปลตาม `Accept-Language` ต่อ; อื่น ๆ 502; ติดต่อไม่ได้ 504.
- `line/code/route.ts:14,41`: ตรวจ session → ออก code ที่ bridge → ผูก code กับผู้ใช้ใน mainapi ก่อนคืน code (ผูกไม่ได้ = ไม่คืน code/`loginUrl`).
- `link/status/route.ts:55-58`: ตรวจ session แล้วตรวจว่า code เป็นของผู้ใช้นี้ **ก่อน** ถาม bridge; `:89` ส่ง `code` ไปกับ `PUT /profile/link-line`.
- mainapi `backend/internal/authentication/line_link_code.go`: เก็บ `linelinkcode:<code>` → uid ใน `cache_entries` (ผ่าน `ICacher` Get/SetS/Del) อายุ 10 นาที (> 5 นาทีที่ dialog poll); ผูกครั้งแรกชนะ (คนอื่นถือ code อยู่ → 409); check ตอบ 404 แบบเดียวกันทั้ง "ไม่มี/หมดอายุ/เป็นของคนอื่น"; `authentication_http.go:1109-1119` `LinkLine` ตรวจ code ก่อนเรียก service แล้วลบ code หลังผูกสำเร็จ (ใช้ครั้งเดียว); routes `authentication_http.go:153-154`; field `code` ที่ `models/authentication.go:62`; ข้อความ 3 key `auth_err_line_link_code_{invalid,taken,not_owned}` ใน `backend/assets/language/languages.tsv`.
- ไม่ใช้คุกกี้ HMAC ฝั่ง BFF: `frontend.env` บน prod (`deploy/account/provision-server.sh`) ไม่มี secret ฝั่ง server ให้ใช้ และ `/backend/*` rewrite ยิง mainapi ตรงได้อยู่แล้ว — บังคับที่ mainapi จึงครอบคลุมกว่าโดยไม่เพิ่ม infra.

## Regression Test

- `frontend/src/app/api/auth/line/link/status/route.test.ts` (6 ข้อ, fake mainapi + bridge): ไม่มี token → 401 ไม่มี fetch; token ปลอม → 401 body เดียวกัน เรียกแค่ `/verify-token`; A ออก code แล้ว B poll → 404 เรียกแค่ verify+check ไม่ถาม bridge ไม่ link; A poll code ตัวเอง → verify → check → bridge → `PUT` (มี `code`) แล้ว poll ซ้ำ → 404 (ใช้ครั้งเดียว); code ที่ไม่เคยออก/หมดอายุ → 404 ไม่ถาม bridge; mainapi ติดต่อไม่ได้ → 504 ไม่ถาม bridge — 5/6 ข้อ fail กับ route เดิม.
- `frontend/src/app/api/auth/line/code/route.test.ts` (6 ข้อ): 5 ข้อเดิม (เพิ่มขั้นผูก code ในลำดับ fetch) + mainapi ไม่ยอมผูก (409) → ไม่คืน code/`loginUrl`.
- `backend/internal/authentication/line_link_code_test.go`: ผูก/ตรวจ/คนอื่นผูกทับไม่ได้/ไม่มี uid/อักขระควบคุม/store error ไม่นับเป็นเจ้าของ/ลบหลังใช้; handler จริงผ่าน `HTTPContext` (409 ข้อความ en, 404 ข้อความ th, 400 JSON เสีย); `LinkLine` ปฏิเสธ code ของคนอื่นหรือไม่มี code ก่อนถึง service; `TestLineLinkCodeOnPostgresCache` กับ PostgreSQL จริง (อายุ > 9 นาที, หมดอายุแล้วถูกปฏิเสธ, ลบแถวหลังใช้ — ข้ามถ้าไม่ตั้ง `BC_GL_TEST_POSTGRES_DSN`).

## Follow-up 2 (review 2026-09-25)

- ปิด `/backend/profile/link-line` + `/backend/profile/link-line/:path*` ใน `blockedDangerousRoutes` (`frontend/next.config.ts:60-63`) — เดิมผู้ที่ login แล้ว `POST /backend/profile/link-line/code` ด้วย code ที่คิดเอง (mainapi ผูกให้เลยเพราะไม่รู้ว่า bridge ออก code ไหน) แล้ว `PUT /backend/profile/link-line` ด้วย `lineuserid` ของคนอื่นได้ ข้ามการผูก code ทั้งชุด. ตอนนี้เรียก 3 endpoint นี้ได้แค่จาก BFF (`/api/auth/line/*` → `serverMainApiBase()` ภายใน) ซึ่งผูกเฉพาะ code ที่ bridge ออกให้จริงและเอา `lineuserid` จาก bridge; browser ไม่เคยเรียกตรง.
- เพิ่ม `/profile/link-line`, `/profile/link-line/code`, `/profile/link-line/code/check` ใน `exceptShopPaths` (`backend/main.go:45-73`, ย้ายจากตัวแปรใน `main()` มาเป็น package var ให้ทดสอบได้) — จอ workspace ที่ยังไม่เลือกกลุ่มกิจการเคยได้ 401 "Shop not selected." → BFF แสดง "ไม่พบ token", `authFetch` refresh token ทิ้งฟรีและ bridge ออก code ทิ้ง 2 ใบ. handler ใช้แค่ `UID`/`Username` ที่ `loginOnlyUserInfo` เก็บไว้; จับคู่ตาม path จึงครอบ `DELETE /profile/link-line` (ยกเลิกเชื่อม) ด้วย.
- `LinkLine` service ไม่บอก username ของบัญชีที่ถือ LINE นั้นแล้ว (`backend/internal/authentication/services/authentication_service.go:1021-1022`).
- Regression: `frontend/src/lib/next-config-rewrites.test.ts` (ใช้ `getPathMatch` ตัวเดียวกับ Next: 3 path + ตัวพิมพ์ใหญ่ถูกบล็อก, `/backend/profile*` อื่นยังผ่าน — fail เมื่อถอด 2 บรรทัดออก); `backend/main_security_test.go` `TestExceptShopPathsAllowLineLinkWithoutHolding`; `authentication_service_test.go` `TestAuthService_LinkLineTakenDoesNotRevealOtherUsername`.

## ยังค้าง

- mainapi ยังเชื่อ `code` + `lineuserid` ที่ผู้เรียกส่งมา — ปลอดภัยเพราะ Next บล็อกทาง `/backend/*` และ prod ผูก mainapi ไว้ที่ `127.0.0.1:8888`; เครื่องที่เปิด port mainapi ออก LAN (เช่น on-prem) ยังยิงตรงได้. ทางแก้ถาวร: ให้ mainapi ถาม bridge เองว่า code นี้ยืนยันแล้วและเป็น LINE ใคร แทนรับจาก client.
- ผูก code แบบ Get แล้ว Set ไม่ atomic (`ICacher` ไม่มี set-if-absent) — ต้องเดา code ที่ bridge เพิ่งออกให้ทันช่วงมิลลิวินาทีนั้น (ตอนนี้ต้องผ่าน BFF ซึ่งผูกเฉพาะ code ที่ bridge ออกให้ผู้เรียกเอง).
