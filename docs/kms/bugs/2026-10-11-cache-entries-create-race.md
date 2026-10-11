---
date: 2026-10-11
severity: medium
component: [backend]
tags: [bc-account, postgres, auth, verify]
fixed: true
---

# Symptom

`npm run verify:all` ขั้น postgres ตกที่ `TestLineLinkCodeOnPostgresCache` ด้วย
`new cacher: create cache_entries: pq: duplicate key value violates unique constraint "pg_type_typname_nsp_index"`
(ตกเฉพาะบน PostgreSQL ใหม่เอี่ยมที่ verify สร้างขึ้น ส่วน DB local ที่มีตารางอยู่แล้วผ่าน)

# Root Cause

`microservice.NewCacher` (`backend/pkg/microservice/cacher.go`) รัน `CREATE TABLE IF NOT EXISTS cache_entries` ตรง ๆ
`go test ./...` รันหลาย package พร้อมกัน (และ service หลายตัวเปิด cache พร้อมกันได้) — บน DB ที่ยังไม่มีตาราง
PostgreSQL สองคำสั่งที่เห็นว่า "ยังไม่มี" พร้อมกันจะชน unique index ของ `pg_type` (`IF NOT EXISTS` ไม่กันการรันพร้อมกัน)

# Fix

สร้างตารางใน transaction ที่ถือ `pg_advisory_xact_lock(hashtext('bc_cache_entries_schema'))` ก่อน
แบบเดียวกับ schema ของ generalledger (`backend/internal/generalledger/postgres.go:65`) — commit `1ae4b95f`

# Regression Test

`TestNewCacherConcurrentOnFreshSchema` (`backend/pkg/microservice/cacher_schema_test.go`) — 12 goroutine เรียก `NewCacher`
บน schema ใหม่พร้อมกัน: ตกก่อนแก้ (error เดียวกับข้างบน), ผ่าน 10/10 รอบหลังแก้ (ต้องตั้ง `BC_GL_TEST_POSTGRES_DSN`)
