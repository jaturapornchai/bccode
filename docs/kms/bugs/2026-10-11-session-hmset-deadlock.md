---
date: 2026-10-11
severity: high
component: [backend, db]
tags: [bc-account, go, postgres, auth, session, uat]
fixed: true
---

# Symptom

เปิดจอ "บัญชีเข้าระบบ" (`/user`) แล้วบางครั้ง (~2 ใน 7 ครั้ง) `GET /api/workspace/holdings` ได้ 500 หรือ 401 "คืนค่าบริษัทที่เลือกไม่สำเร็จ" — log ของ `mainapi` เห็น `POST /select-holding` ใช้เวลา 1.0 วินาทีพอดีแล้วล้ม (พบจาก `npm run uat`)

## Root Cause
BFF `frontend/src/app/api/workspace/[...workspacePath]/route.ts` (GET holdings) เรียก `select-holding` กับ session เดียวกันซ้อนกัน 2 ชุดเมื่อจอ mount ซ้ำ
`backend/pkg/microservice/cacher.go` `HMSet` upsert ทีละ field ตามลำดับการวน map ของ Go (สุ่มทุกครั้ง) ใน transaction เดียว → 2 transaction ล็อกแถว `cache_entries` คนละลำดับ → PostgreSQL ตรวจพบ deadlock หลัง `deadlock_timeout` 1 วินาที แล้ว abort ตัวหนึ่ง (`pq: deadlock detected`) → select-holding ล้ม

สาเหตุที่ 2 (เจอหลังแก้ข้อแรก ยังได้ 401 บางครั้ง): ทุก request ที่ผ่าน auth ต่ออายุ session ด้วย `Cacher.Expire` (`touchSession` ใน `backend/pkg/microservice/auth.go`) ซึ่งเป็น `UPDATE ... WHERE cache_key = ANY($1)` ล็อกแถวตามลำดับใน heap ไม่ใช่ลำดับ field → ชนกับ `HMSet` ของ select-holding อีกตัวบน session เดียวกัน → PostgreSQL ยกเลิก Expire → `touchSession` error → request ได้ 401 แล้ว browser ต้อง `/api/auth/refresh` ใหม่
- log ของ `mainapi` ไม่มีบรรทัด 401 เพราะ request ที่ auth ปฏิเสธไม่ผ่าน `RequestLoggerMiddleware` — เห็นได้จาก select-holding ที่ใช้ 1.0 วินาทีพอดี + `POST /refresh` ตามมาทันที

## Fix
`HMSet` วน field ตามลำดับที่เรียงแล้ว (`slices.Sorted(maps.Keys(...))`) → ทุก writer ล็อกแถวลำดับเดียวกัน รอคิวแทน deadlock
- `Expires` ล็อกแถวด้วย `SELECT ... ORDER BY cache_key, field FOR UPDATE` ก่อน UPDATE → ลำดับเดียวกับ `HMSet`

## Regression Test
- `backend/pkg/microservice/cacher_hmset_concurrency_test.go` (`TestCacherHMSetConcurrentSameKeyNoDeadlock`, ต้องตั้ง `BC_GL_TEST_POSTGRES_DSN`) — ก่อนแก้ fail `pq: deadlock detected`, หลังแก้ pass
- `TestCacherExpireConcurrentWithHMSetNoDeadlock` (ไฟล์เดียวกัน) — ก่อนแก้ fail `Expire: pq: deadlock detected` (77 วินาที), หลังแก้ pass 3 วินาที
- กฎทั่วไป: ทุก statement/transaction ที่เขียนหลายแถวของ key เดียวกัน (`cache_entries` ฯลฯ) ต้องล็อกตามลำดับเดียวกัน — เรียง key ใน Go หรือ `ORDER BY ... FOR UPDATE` ใน SQL
