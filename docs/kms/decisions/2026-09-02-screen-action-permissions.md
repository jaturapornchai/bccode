---
date: 2026-09-02
status: accepted
tags: [bc-account, backend, frontend, permission]
---

# ADR: สิทธิ์ต่อจอแยก เข้า/เพิ่ม/แก้ไข/ลบ

## บริบท
ลุงจืดสั่ง "แต่ละจอจะมี เข้า, เพิ่ม, แก้ไข, ลบ" บนจอกำหนดสิทธิ์ตามบทบาท เดิม `role_permission.permissions` เป็นรายการรหัสจอ (1 ติ๊ก = เข้าได้ทั้งจอ) ใช้ซ่อนเมนูฝั่ง frontend เท่านั้น

## การตัดสินใจ
- คงโครง `[]string` เดิม เพิ่ม entry รูปแบบ `รหัสจอ:create|update|delete` (`รหัสจอ` = เข้า, `*` = ทั้งหมด) → ข้อมูลเก่าใช้ได้ ไม่ต้อง migrate
- Backend: validate รูปแบบใน `NormalizeRequest` (rolepermission/models) เท่านั้น ยังไม่บังคับต่อ API (ระยะ 2)
- Frontend: editor เดิม (`PermissionLinkMultiSelectEditor`, isGroup) เพิ่ม 4 ช่องต่อจอ; hook `useScreenActions(auth, workspace, route)` อ่าน `/permissiongroup/me` แล้ว map รหัสจอจาก `flattenMenuItems()` ด้วย route → ซ่อนปุ่ม เพิ่ม/คัดลอก, แก้ไข, ลบ ใน SettingDataList/SettingDetailPanel (prop `actions` optional, default เปิดหมด)
- จอที่ไม่อยู่ในเมนู (เช่น wizard) / โหลดสิทธิ์ไม่ได้ → ไม่ล็อกปุ่ม (fail-open เฉพาะ UI; การเข้าจอถูกคุมที่เมนูหลักแล้ว)

## ผลกระทบ / งานต่อ
ระยะ 2: บังคับที่ backend (POST/PUT/DELETE ของ system-settings ตรวจ `รหัสจอ:action` จาก role permission ของ membership) · commit `9af96103` · docs/organization.md อัปเดต

## UX (ปรับตามลุงจืด "เลือกยาก")
เปลี่ยนจาก card grid เป็นตาราง `RoleScreenMatrix` (components/system-settings/field-editors/role-screen-matrix.tsx): แถวจัดกลุ่มตามเมนู (section › group) คอลัมน์ เข้า/เพิ่ม/แก้ไข/ลบ/ทั้งหมด, ค้นหา + กรอง เลือกแล้ว/ยังไม่เลือก, หัวตารางติ๊กทีเดียวกับจอที่แสดง · verify: 128 จอ → ค้น "ซื้อ" เหลือ 8, header เข้า = 8, ติ๊กแก้ไขแล้วเข้าอัตโนมัติ, ยกเลิกเข้า = ล้าง action
