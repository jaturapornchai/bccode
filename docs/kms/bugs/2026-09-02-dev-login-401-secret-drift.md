---
date: 2026-09-02
severity: medium
component: [backend, docker, frontend]
tags: [bc-account, docker, dev-login]
fixed: true
---

# Symptom

กดปุ่ม "เข้าทดสอบระบบ (Dev Login)" บน `localhost:3000` → error pill "ไม่อนุญาตให้ใช้ Dev Login" (frontend route ได้ 401 จาก backend) ทั้งที่ Docker Desktop และ `mainapi` healthy และปกติเข้าได้ทันที เกิดซ้ำจากที่เคยเจอในรอบ UAT 2026-09-01

## Root Cause
`BCAI_DEV_LOGIN_SECRET` **คนละค่ากัน 2 ฝั่ง**:
- ฝั่ง frontend: `next dev` อ่านจาก **Windows User env** (ไม่ได้อยู่ใน `.env.local`) → hash8 `88e499f9`
- ฝั่ง backend: container `mainapi` (สร้างด้วย `docker run` มือ ไม่มี compose label, env ค้างตั้งแต่สร้าง 2026-09-01 21:18) → hash8 `940cc006`

backend `authentication_http.go` เทียบ constant-time แล้วตอบ 401 "dev login failed" ใน ~20µs (log ยืนยันว่าตกที่ compare ไม่ใช่ DB) ไม่มีอะไร sync ค่านี้ให้อัตโนมัติ — เปลี่ยน env ฝั่ง Windows แล้ว container ไม่รู้

## Fix
recreate `mainapi` จาก `docker inspect` เดิม (image/network/ports/mounts เดิม, env เดิมทั้งชุด ยกเว้น override `BCAI_DEV_LOGIN_ENABLED/USER_UID/SECRET` จาก env ปัจจุบัน ส่งผ่าน `--env-file` temp 0600 แล้วลบ) — rename ตัวเก่าเป็น `mainapi_old` ไว้ rollback ก่อน, `docker diff` ยืนยันว่า `/app/go-app` ไม่ได้ถูก swap จึงใช้ image เดิมได้ตรง
ผล: `curl -X POST -H "X-BC-Dev-Login-Secret: …" :8888/dev-login` → 200 `{"success":true,"token":…}` และปุ่มบนหน้าจอเข้า `/holding` ได้ (ไม่มี commit — เป็น infra local)

## Fix ถาวร (ทำต่อในวันเดียวกัน)
ย้าย `mainapi` ไปรันผ่าน compose: `cd backend && docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --no-deps mainapi` → container มี compose label, restart unless-stopped, env interpolate ใหม่ทุกครั้งที่ up (แก้ secret ใน Windows env แล้วรันคำสั่งเดิม = sync) ใช้ image เดิม ไม่ build ใหม่ · ผลพลอยได้: bootstrap.local.json ชี้ postgres/clickhouse `appdb` ที่มีจริง (ของเดิมชี้ `appdb_test`/`altcoin` ที่ไม่มี)

## Regression Test
1. เช็ค drift โดยไม่ print secret: `printf %s "$BCAI_DEV_LOGIN_SECRET" | sha256sum | cut -c1-8` เทียบ `docker exec mainapi sh -c 'printf %s "$BCAI_DEV_LOGIN_SECRET" | sha256sum'` — ต้องเท่ากัน
2. ทุกครั้งที่เปลี่ยน `BCAI_DEV_LOGIN_*` ใน Windows env → recreate `mainapi` ด้วย (ไม่ใช่แค่ `docker restart` — restart ไม่ reload env)
3. อ่านรหัสสถานะ: 404 = backend ไม่ register route (BC_ENV≠dev/flag ปิด), 401 = secret/uid ไม่ตรง, 503 จาก frontend = env ฝั่ง `next dev` ไม่ครบ

เกี่ยวข้อง: [[2026-09-02]] · memory `local-mainapi-dev-login-secret-drift`
