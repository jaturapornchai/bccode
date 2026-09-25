---
date: 2026-09-08
severity: medium
component: [frontend]
tags: [bc-account, next-bff, permissions, api-proxy]
fixed: true
---

# Symptom

เปิดจอตั้งค่า (เช่น ทะเบียนพนักงาน `/employee`, สิทธิ์การใช้งาน `/permissiongroup`) หลัง Demo Login แล้ว
network log มี `GET /api/system-settings/permissiongroup/me?holdingcode=demo` ตอบ **400 Bad Request**
`{"success":false,"message":"รูปแบบ Backend URL ไม่ถูกต้อง"}` — 1 ครั้งต่อการเปิดจอ ส่วนคำขอเดิมจากเมนูหลักตอบ 200
(ยืนยันบน localhost:3000 วันที่ 2026-09-08: เปิด `/employee` ครั้งเดียว → 400 หนึ่งรายการ)

อาการภายนอก: ปุ่ม เพิ่ม/แก้ไข/ลบ ของจอนั้นถอยไปใช้ค่า default (เปิดทุกปุ่ม) เพราะ hook กลืน error เงียบ ๆ
ไม่ใช่การเปิดสิทธิ์เกินจริง (backend ยังบังคับสิทธิ์เอง) แต่ทำให้ผู้ใช้เห็นปุ่มที่ตัวเองอาจกดไม่ผ่าน

## Root Cause

ไม่ใช่ race ตอน bootstrap — เป็น **ผู้เรียกคนละเส้นทางส่งค่าไม่เท่ากัน**:

- เมนูหลัก `fetchAllowedMenuIds()` ส่ง header `x-bc-backend-url: auth.backendUrl` มาด้วย
  (`frontend/src/app/menu/main-menu-screen.tsx:265-275`) → ผ่าน
- hook `useScreenActions()` ที่จอตั้งค่าเรียก endpoint เดียวกัน แต่ส่งแค่ `Authorization`
  (เดิม `frontend/src/lib/use-screen-actions.ts:44-47`) และ `authFetch()` ไม่เติม header นี้ให้
  (`frontend/src/lib/client-auth-session.ts:174-192`)

ฝั่ง proxy `resolveBaseUrl()` เรียก `getMainApiUrl(getBackendUrlFromRequest(...))` ซึ่ง `validateBackendUrl("")`
โยน "กรุณากรอก Backend URL" ทันที → route ตอบ 400 ทั้งที่ **ค่าจาก client ถูกทิ้งอยู่แล้ว** —
การ fetch จริงใช้ `serverMainApiBase()` / `serverGoApiBase()` (`frontend/src/lib/backend-url.ts:117-127`) เสมอ

## Fix

1. `frontend/src/lib/workspace-api.ts:24-35` — `getMainApiUrl()` ตรวจ URL จาก client เฉพาะเมื่อส่งมาจริง
   (`if (rawBackendUrl.trim()) validateBackendUrl(...)`) ไม่ส่งมา = ใช้ค่าฝั่ง server ต่อ (ค่าที่ผิดรูปแบบยังตอบ 400 เหมือนเดิม)
2. `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts:137-151` — สาขา goapi ก็ตรวจเฉพาะเมื่อมีค่า
3. `frontend/src/lib/use-screen-actions.ts:44-62` — ส่ง `x-bc-backend-url` เหมือนเมนูหลัก (ใส่เฉพาะเมื่อมีค่า) และเพิ่ม `backendUrl` ใน deps

หมายเหตุความปลอดภัย: Backend URL จาก client ไม่เคยเป็นขอบเขตความปลอดภัย — ทุกคำขอยังต้องมี Bearer token
และผ่าน `validateTenantAccess()` (`route.ts:167-185`) ตามเดิม

## Regression Test

- unit: `frontend/src/app/api/system-settings/[[...settingPath]]/route.test.ts`
  - "reads role permissions without a client backend URL (menu/screen-action race)" → คำขอไม่มี backend URL ต้องได้ 200 + ยิงไป `http://localhost:8888/organization/role-permission/me?offset=0&limit=1000`
  - "falls back to the server backend base for goapi-backed slugs without a client backend URL" → สาขา goapi ใช้ `.../goapi/atlas/get`
  - "still rejects a malformed client backend URL" → ส่งค่าเพี้ยน (`ftp://bad url`) ต้องยัง 400 และไม่ยิง backend
  - ทั้งไฟล์ตรึง env ด้วย `vi.stubEnv("BCAI_LOCAL_BACKEND_URL", ...)` ไม่ให้ผลลัพธ์ขึ้นกับ env ของเครื่อง
- ตรวจ 2 เทสต์แรกกับโค้ดก่อนแก้แล้ว: fail ทั้งคู่ (400) → ผ่านหลังแก้; ทั้งชุด frontend 327/327 ผ่าน
- UAT ด้วยมือ (2026-09-08, dev server localhost:3000): Demo Login → เมนูหลัก → `/employee` → `/permissiongroup`
  ไม่มี 400 ใน network log และไม่มี console error
