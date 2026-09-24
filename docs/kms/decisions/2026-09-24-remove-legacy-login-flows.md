---
date: 2026-09-24
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, go, auth, security]
---

# ลบ flow login เก่าที่ไม่มีทางเรียกถึง: LoginEmail, Firebase (TokenLogin), LINE login

## Context

- reviewer ชี้ว่า `backend/internal/authentication/**` มี flow login เก่าที่ไม่มีทางเรียกถึง; กฎ "เดินหน้าอย่างเดียว" ใน `AGENTS.md` ให้ลบโค้ดเก่าใน change เดียวกัน ห้ามเก็บไว้ "เผื่อ"
- หลักฐานว่าไม่มีทางเรียกถึง (ตรวจ 2026-09-24):
  - **route**: `RegisterHttp` (`backend/internal/authentication/authentication_http.go:126-164`) ไม่ register `LoginEmail`, `TokenLogin`, `LoginWithLine`, `LoginWithLineUserID`; grep ทั้ง `backend/` ไม่พบผู้เรียก handler หรือ service method เหล่านี้ นอกจากตัว handler เอง
  - **frontend**: BFF `frontend/src/app/api/auth/*` เรียก backend แค่ `/login`, `/googlelogin`, `/demo-login`, `/dev-login`, `/refresh`, `/logout`, `PUT /profile/link-line`; ไม่มีโค้ดส่วนไหนเรียก `/login/email`, `/tokenlogin`, `/linelogin`, `/login/line` (เจอแค่ใน blocklist ของ `frontend/next.config.ts`)
  - **deploy**: `deploy/account/compose.yml` อ่านค่าผ่าน `env_file`; ชื่อ key ใน `/etc/bcai-account/*.env` บน prod (อ่านเฉพาะชื่อ key ไม่อ่านค่า) และ `deploy/account/provision-server.sh` ไม่มี `LINE_CLIENT_ID`, `FIREBASE_PROJECT_ID`, `GOOGLE_APPLICATION_CREDENTIALS`
- `LoginEmail` เป็นโค้ดที่อันตราย: ถ้าไม่เจอ username จะสร้าง user ใหม่ แล้วออก token ให้ username ใดก็ได้**โดยไม่ตรวจรหัสผ่าน** ถ้าวันหนึ่งมีคน register route นี้กลับมาโดยไม่รู้ ระบบจะถูกข้ามการยืนยันตัวตนได้ทันที

## Decision

ลบ:

| flow | ที่ลบ |
|---|---|
| LoginEmail | handler และ service `LoginEmail` |
| Firebase | handler `TokenLogin` (รวมใน `IAuthenticationHttp`), service `LoginWithFirebaseToken`, package `backend/internal/firebase/` (รวม test), dependency `firebase.google.com/go/v4` + `google.golang.org/api` และ transitive ที่ไม่มีใครใช้แล้ว (`go mod tidy`: ลบอย่างเดียว go.mod −22 / go.sum −155 บรรทัด), key setup-config `service.firebaseprojectid` (loader ของ mainapi และ goapi + `frontend/src/lib/setup-config.ts`), `FIREBASE_PROJECT_ID` ใน `backend/docker-compose.yml` |
| LINE login | handler `LoginWithLine`, `LoginWithLineUserID`, service `LoginWithLineToken`, `LoginWithLineUserID`, model `LineLoginRequest`, `LineUserLoginRequest`, package `backend/internal/line/` (รวม test), `IConfig.LineClientId()` (`LINE_CLIENT_ID`) |
| ส่วนที่ตายตามไปด้วย | พารามิเตอร์ `firebaseAdapter`/`lineAdapter` ของ `NewAuthenticationService` และ mock ใน test; รายการ `/backend/login/email`, `/backend/login/line`, `/backend/linelogin`, `/backend/tokenlogin` ใน blocklist ของ `frontend/next.config.ts`; แถว `firebase`/`line` ใน `docs/kms/08-legacy-modules.md`, `LINE_CLIENT_ID` ใน `docs/kms/13-pkg-framework.md`, `FIREBASE_PROJECT_ID` ใน `docs/kms/10-infra-deploy.md`, เลขบรรทัดที่เลื่อนใน `docs/kms/*` และ `docs/reference/CODE-MAP.md` |

ไม่แตะ:

- Google login (`/googlelogin`), ปุ่ม Demo (`/demo-login`), Dev login (`/dev-login`), `/login` แบบรหัสผ่าน, `/refresh`, `/logout`
- **การเชื่อมบัญชี LINE** (`PUT/DELETE /profile/link-line`, repository `FindByLineUserID`/`SetLineIdentity`, BFF `api/auth/line/*`, ปุ่มเชื่อมต่อ LINE ใน `main-menu-screen.tsx`/`workspace-screen.tsx`) — route ยัง register อยู่และ UI ยังเรียกใช้ จึงไม่ใช่โค้ดที่ไม่มีทางเรียกถึง

ตรวจเรื่อง log ด้วย: หลังลบ `[DEBUG_LOGIN]` ไม่มีจุดไหนใน `backend/` ที่ log request body หรือรหัสผ่านอีก (grep หา middleware ที่ dump body, การ log `ReadInput()`, การอ่าน raw body, และคำสั่ง log ที่มี password/secret/credential)

## Alternatives

- **เก็บไว้เผื่อแอปมือถือกลับมาใช้** — ไม่เลือก: ใน repo ไม่มี client ตัวไหนเรียก, prod ไม่มี config และขัดกฎเดินหน้าอย่างเดียว; ถ้าต้องการกู้คืน เปิดดูได้จาก git history ก่อน commit ที่เพิ่ม ADR นี้
- **ลบการเชื่อมบัญชี LINE ไปพร้อมกัน** — ไม่เลือกรอบนี้: UI ยังเรียกใช้อยู่ ต้องให้ลุงจืดตัดสินก่อน (ดู Consequences)
- **ลบ handler อื่นที่ยังไม่มีทางเรียกถึง** (`Poslogin`, `LoginWithPhoneNumber`, `Register*`, OTP, `ResetPasswordToDefault`) — อยู่นอกขอบเขตที่ reviewer ระบุ; ยังอยู่ในตาราง DEAD ของ `docs/kms/03-auth-tenancy.md` §2 เพื่อรอตัดสิน

## Consequences

- พื้นผิวโจมตีลดลง (ไม่เหลือโค้ดที่ออก token โดยไม่ตรวจรหัสผ่าน), dependency ของ Google Cloud/gRPC หลุดออกจาก go.mod และ mainapi ไม่พิมพ์ `[Firebase] ไม่พบ ...` ตอนเริ่มทำงานอีก
- พฤติกรรมภายนอกไม่เปลี่ยน: route เหล่านี้ตอบ 404 อยู่แล้วก่อนลบ เพราะไม่เคยถูก register
- เรื่องที่ค้าง: การเชื่อมบัญชี LINE ใช้งานบน prod ไม่ได้จริง เพราะ prod ไม่มี `BC_AUTH_BRIDGE_URL` (`getAuthBridgeUrl` ใน `frontend/src/lib/auth-bridge.ts` จะ throw) — ต้องเลือกว่าจะตั้งค่า bridge หรือลบ link flow ทั้งชุด (BFF + ปุ่ม + `/profile/link-line`)
