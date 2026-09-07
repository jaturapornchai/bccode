---
date: 2026-09-02
status: accepted
tags: [bc-account, frontend, ux]
---

# ADR: UX/UI ทั้งระบบต้อง "พรีเมี่ยม + ใช้ง่าย + เหมาะกับคนไทย"

## บริบท
หลัง premium pass หน้า login/holding (2026-09-02) ลุงจืดดูจอ /workspace (เลือกบริษัท) แล้วสั่งตั้งกฎให้ทั้งระบบเป็นมาตรฐานเดียวกัน

## การตัดสินใจ
เพิ่ม block กฎใน AGENTS.md ต่อจากกฎ "คนไทย 40+" — 9 ข้อ: accent ตระกูลเดียวจาก --primary, light+dark ผ่านตัวแปร (ทดสอบด้วยปุ่มสลับธีมจริง), พื้นผิวมีชั้น, control ภาษาเดียวกัน, motion แบบ ambient เท่านั้น, รูป hero คุณภาพ, popover ห้ามโดน clip, ตรวจรับด้วย screenshot 1600/1280/1024/768 × 2 โหมด, CSS append-only

## ผลกระทบ
จอถัดไปที่ต้องยกระดับ: /workspace, settings wizard, menu · reference = login/holding

เกี่ยวข้อง: [[2026-09-02]] · [[2026-09-02-dev-login-401-secret-drift]]
