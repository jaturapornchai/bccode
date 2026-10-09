---
date: 2026-10-09
severity: medium
component: [backend]
tags: [bc-account, demo, holding, auth]
fixed: code-fixed-prod-pending
---

# Symptom

ปุ่ม "ทดลองใช้ระบบ (Demo)" บน prod ล็อกอินเป็นบัญชี `somkid@rungrueng.co.th` ซึ่งเป็น OWNER ของ 2 กลุ่มกิจการ: `RUNGRUENG` (ข้อมูลตัวอย่าง) และ `TOFFT1` (ว่าง: 23 ตาราง, `gl_records`/`gl_lines`/`doc` 0 แถว, สร้าง 2026-10-01, Demo เป็นสมาชิกคนเดียว) — ใครกดปุ่ม Demo ก็สร้างกลุ่มกิจการใหม่ที่ติดบัญชีร่วมถาวรได้

# Root cause

`organization.requireOrganizationCreator` ยกเว้นเงื่อนไข "ต้องมี Google identity ที่ยืนยันแล้ว" ให้บัญชี Demo ทุกกรณี (`demo.IsDemoUser`) — รวมถึง `RequireEmailedAccount` ที่ใช้ตอนสร้าง Holding (`shop_http.go`, `ListShopCanAccess` → `cancreateholding`) ทั้งที่ Demo ต้องการแค่จัดการบริษัท/สาขาใน Holding ที่ seed ไว้

# Fix

ยกเว้น Demo เฉพาะ path `requireAdmin` (สร้างบริษัท/สาขาใน Holding ของตัวเอง); สร้าง Holding ใหม่ต้องมี Google identity เหมือนทุกบัญชี → `cancreateholding=false` ซ่อนปุ่มเพิ่มกลุ่มกิจการ และ `POST /shop` ตอบ 403 (`creator_access.go`, เทสต์ `TestDemoAccountCannotBootstrapHolding`)

# Regression test

`backend/internal/organization/creator_access_demo_test.go`; UAT local: Demo `list-shop` → `cancreateholding=false`, `POST /shop` → 403 "an active verified Google identity is required"

# ยังค้าง (ต้องให้ลุงจืดสั่งเฉพาะข้อ — แตะ prod)

`TOFFT1` ที่ Demo สร้างไว้บน prod ยังอยู่: ตัวเลือก (ก) ตั้ง membership ของ Demo ใน `tofft1` เป็น inactive (กลุ่มกิจการจะไม่มีสมาชิก active) (ข) ปิด `BCAI_DEMO_LOGIN_ENABLED` ใน `/etc/bcai-account/backend.env` แล้ว recreate mainapi (ค) สร้างผู้ใช้ Demo แยกที่มี holding ของตัวเอง — โค้ดที่แก้นี้กันไม่ให้เกิดซ้ำเท่านั้น ไม่ลบของที่มีอยู่
