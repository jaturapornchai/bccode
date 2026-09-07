---
date: 2026-09-02
severity: high
component: [frontend]
tags: [bc-account, login, motion]
fixed: true
---

# Symptom

prod https://account.bcaicloud.com เปิดในแท็บพื้นหลัง (หรือแท็บถูกซ่อนระหว่างโหลด) → ฝั่งฟอร์ม login ว่างเปล่า กดปุ่ม Google ไม่มีอะไรเกิดขึ้น (ผู้ใช้รายงานว่า "กด Google เงียบ")

## Root Cause
`motion.section` ของ `.brand-panel` / `.form-panel` ใช้ `initial={{opacity:0,y:16}}` + `animate` — motion ขับ animation ด้วย requestAnimationFrame ซึ่ง **ไม่ทำงานในแท็บที่ `document.hidden`** → inline `opacity:0; transform:translateY(16px)` ค้างถาวร (ตรวจด้วย Chrome MCP tab: `hidden=true`, rAF ไม่ fire ใน 1.5s, style ค้าง) — bug เดิมมีก่อน premium pass

## Fix
commit `299c7f8e`: panel ใช้ `initial={false} animate="animate"` (ลูกยัง stagger ผ่าน variants) + CSS `@keyframes login-panel-enter` fill-mode both ใน globals.css (reduced-motion = none) · deploy `bcai-account-frontend:r20260902-2`

## Regression Test
1. เปิดหน้า login ในแท็บพื้นหลัง (Ctrl+click) รอ 5 วิ แล้วสลับมา → การ์ดต้องแสดง
2. JS: `document.querySelector(".form-panel").getAttribute("style")` ต้องเป็น `null` (ไม่มี inline opacity)
3. หลีกเลี่ยง motion `initial` บน container ที่บังทั้งจอ — ใช้ CSS keyframes สำหรับ entrance ระดับ panel
