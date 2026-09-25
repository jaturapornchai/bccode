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

- `frontend/src/app/api/auth/line/code/route.ts`: เรียก `requireBearerToken` ก่อน แล้วตรวจ session จริงกับ mainapi `GET ${serverMainApiBase()}/verify-token` (middleware เดียวกับ API ที่ป้องกันทุกตัว: token เป็นค่าสุ่มทึบ ไม่ใช่ JWT — ค้นใน `cache_entries` แล้วตรวจ session/สิทธิ์ที่ยังใช้งานอยู่ ไม่มีการตรวจลายเซ็น (`backend/pkg/microservice/auth.go` `MWFuncMixShop`); `/verify-token` อยู่ใน `exceptShopPath` ของ `backend/main.go` จึงใช้ได้ทั้งจอ workspace และเมนูหลัก) — mainapi ตอบ 401/403 → BFF ตอบ 401 body เดียวกับ `requireBearerToken` (ให้ `authFetch` refresh token แล้วลองใหม่ 1 ครั้ง); mainapi ล่ม → 502/504 และ**ไม่**ยิง bridge.
- Route ไม่รับ input ใด ๆ: bridge URL มาจาก env `BC_AUTH_BRIDGE_URL` เท่านั้น, ไม่อ่าน/ไม่ส่งต่อ request body, และไม่ส่ง token ของผู้ใช้ไปให้ bridge; ไม่มี log token.
- ผู้เรียกทั้งสองจอแนบ `Authorization: Bearer ${auth.token}` แล้ว (`main-menu-screen.tsx:825-828`, `workspace-screen.tsx:784-787`).
- ไม่เลือกลบทั้ง flow: ADR 2026-09-24 ระบุว่าการลบการผูกบัญชี LINE ต้องให้ลุงจืดตัดสินก่อน.

## Regression Test

- `frontend/src/app/api/auth/line/code/route.test.ts` — ไม่มี Authorization → 401 และไม่มี outbound fetch; Bearer ที่ mainapi ปฏิเสธ → 401 และไม่ยิง bridge; session ถูกต้อง → เรียก `/verify-token` แล้ว bridge `?action=create` ครั้งเดียว โดยไม่ส่ง Authorization/body ให้ bridge; body ที่มี `bridgeUrl` ปลอม (และ JSON เสีย) ถูกเมิน; mainapi ติดต่อไม่ได้ → 504 ไม่ออก code. เทสต์ 4 ใน 5 ข้อ fail กับโค้ดเดิม.
- ยังค้าง (นอกขอบเขตรอบนี้): code ที่ออกไม่ผูกกับผู้ใช้ที่ขอ — `api/auth/line/link/status` รับ code ใดก็ได้จาก Bearer ใดก็ได้; ถ้าจะเปิดใช้ bridge จริงควรผูก code กับ uid ก่อน.
