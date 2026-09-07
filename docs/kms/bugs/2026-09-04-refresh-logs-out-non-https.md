---
date: 2026-09-04
severity: high
component: [frontend]
tags: [bc-account, auth, cookie, tailscale, nextjs]
fixed: true
---

# Symptom

ผู้ใช้กด refresh (F5) ขณะอยู่หน้าจอทำงาน (เช่น `/menu`) แล้วเด้งกลับหน้า login ทันที
เกิดเฉพาะเวลาเข้าเว็บผ่าน origin ที่ไม่ใช่ HTTP(S) `localhost` (เช่น Tailscale IP `100.118.122.7:3000` ผ่าน HTTP ธรรมดา) — บน `localhost:3000` refresh ทำงานปกติ ไม่หลุด

# Root Cause

`bc_refresh_token` cookie (httpOnly, [[login-css-override-traps|auth session]] refresh token, rotate ทุกครั้งที่ใช้) ถูก set ด้วย `secure: true` ตายตัวใน
`frontend/src/lib/auth-session-server.ts` (`setRefreshTokenCookie`/`clearRefreshTokenCookie`)

Cookie ที่มี `Secure` flag เบราว์เซอร์จะไม่ส่งกลับเลยถ้า origin ไม่ใช่ HTTPS และไม่ใช่ literal `localhost`
— ตอน reload หน้าเว็บ, client ในหน่วยความจำ (access token) หายไปตาม SPA lifecycle,
ต้องพึ่ง `POST /api/auth/refresh` (ใช้ cookie นี้) เพื่อกู้ session กลับ
แต่ cookie ไม่เคยถูกส่งมา → backend เห็นไม่มี refresh token → 401 → client เคลียร์ session → เด้งไป `/`

Verify ผ่าน network log จริง (Claude Browser pane):
- `http://localhost:3000` reload → `POST /api/auth/refresh → 200 OK` (ทำงานถูก)
- `http://100.118.122.7:3000` (Tailscale, HTTP) reload → `POST /api/auth/refresh → 401 Unauthorized` (bug)

# Fix

`frontend/src/lib/auth-session-server.ts`: เปลี่ยน `secure: true` → `secure: isSecureCookie`
โดย `const isSecureCookie = process.env.NODE_ENV === "production"`

เหตุผลที่เลือก `NODE_ENV` แทนการ detect protocol ต่อ request (เช่นอ่าน `x-forwarded-proto`):
- dev เสมอเข้าผ่าน HTTP (localhost หรือ IP ในวง LAN/Tailscale) → ไม่ต้อง secure ก็ปลอดภัยพอ (เครือข่ายไว้ใจได้)
- prod deploy ผ่าน HTTPS เสมอ (Cloudflare Tunnel / Caddy ตาม [[account-bcaicloud-deploy]] และ [[do-sgp1-new-prod-server]]) → บังคับ secure ได้แน่นอน ไม่ต้องเดา header ที่พังง่ายเวลาอยู่หลัง proxy

Verify ซ้ำหลังแก้ (Tailscale IP): `POST /api/auth/refresh → 200 OK`, reload แล้วยังอยู่หน้า `/holding` ไม่เด้ง login
Regression check บน `localhost` ผ่านปกติเหมือนเดิม
`npx tsc --noEmit` ผ่าน 0 error

## Regression Test

ไม่ได้เขียน automated test (เป็น cookie/env behavior ที่ต้องรัน dev server จริงถึงจะเห็นผล) —
วิธี manual verify future regression: เข้าเว็บผ่าน non-localhost origin (Tailscale IP หรือ LAN IP) ตอน dev,
login แล้ว reload หน้าใดก็ได้ที่ต้อง auth — ต้อง**ไม่**เด้งกลับ `/`, เช็ค Network tab ว่า `/api/auth/refresh` ตอบ 200
