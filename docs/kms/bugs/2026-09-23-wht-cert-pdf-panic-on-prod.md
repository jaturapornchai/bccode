---
date: 2026-09-23
severity: high
component: [backend, db]
tags: [bc-account, go, postgres, tax]
fixed: true
---

# Symptom

บน production กดสร้างหนังสือรับรอง 50 ทวิ แล้วได้ 500 `config problem: mkdir /home/appuser: permission denied` (panic recovered) — บนเครื่อง dev สร้างได้ปกติ
ก่อนถึงจุดนั้น บริษัท `rungrueng/01` บน prod ถูกปฏิเสธด้วย `wht_cert_taxid_checksum` ทุกครั้ง เพราะเลขผู้เสียภาษีของบริษัท (`0105566123450`) หลักตรวจสอบผิด

พบพร้อมกัน: เลือกบริษัทแล้ว mainapi log `column "business_code" of relation "shop_user_access_logs" does not exist` ทุกครั้ง

# Root Cause

1. pdfcpu (`api.*` ใน `backend/internal/whtcert/render.go`) สร้างโฟลเดอร์ config ใต้ `os.UserConfigDir()` ครั้งแรกที่ใช้ — container รันด้วย `appuser` ที่ไม่มี home (`adduser -H` ใน `backend/Dockerfile`) จึง mkdir ไม่ได้ และ pdfcpu panic
2. seed ข้อมูลบริษัท (`deploy_fresh_database.sql`, `seed_*.sql`, `scripts/demo-data.json`) ใช้เลขผู้เสียภาษีที่หลักตรวจสอบไม่ถูก — backend ใช้ข้อมูลบริษัทเป็นผู้จ่ายเสมอ (กันปลอมผู้จ่าย) ผู้ใช้จึงแก้จากจอไม่ได้
3. ตาราง `shop_user_access_logs` บน prod ถูกสร้างก่อนมีคอลัมน์ `business_code`; `CREATE TABLE IF NOT EXISTS` ไม่เติมคอลัมน์ให้ตารางเดิม

# Fix

- `whtcert/render.go`: `init()` เรียก `api.DisableConfigDir()` — pdfcpu ใช้ค่าเริ่มต้นในหน่วยความจำ ไม่แตะ `$HOME`
- `centraldb.go`: เพิ่ม `ALTER TABLE shop_user_access_logs ADD COLUMN IF NOT EXISTS business_code ...` ในชุด schema (โค้ดคือ migration)
- แก้เลขผู้เสียภาษีบริษัทใน seed ทั้ง 4 ไฟล์ให้หลักตรวจสอบถูก และแก้แถว `companies` บน prod 1 แถว (`rungrueng/01` → `0105566123456`)

# Regression Test

- `TestRenderWithoutHomeDir` (`backend/internal/whtcert/whtcert_test.go`): ตั้ง HOME/XDG_CONFIG_HOME/AppData ไปที่ path ที่ไม่มีจริง แล้วต้องสร้าง PDF ได้โดยไม่สร้างโฟลเดอร์ — ยืนยันแล้วว่า fail เมื่อถอด `DisableConfigDir`
- `TestEnsureSchemaAddsBusinessCodeToOldAccessLog` (`backend/internal/centraldb/centraldb_integration_test.go`): สร้างตารางรุ่นเก่าแล้ว `EnsureSchema` ต้องเติมคอลัมน์ให้
- บทเรียน: ทดสอบ PDF/ไฟล์บน image จริง (ผู้ใช้ non-root ไม่มี home) ไม่ใช่แค่ `go test` บนเครื่อง dev; ตรวจข้อมูลใน PostgreSQL ทุกครั้งก่อนสรุป
