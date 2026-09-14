---
date: 2026-09-14
status: open
tags: [bc-account, frontend, i18n, system-settings, react-hooks]
---
# จอตั้งค่าแบบเปิดตรง (`/[systemSetting]`) กดเลือกภาษาแล้ว render loop — ภาษาสลับ th/ja ไม่หยุด

**อาการ** — เปิดจอตั้งค่าผ่าน URL ตรง เช่น `/productwarehousescreen`, `/employee` (route `frontend/src/app/[systemSetting]/page.tsx` → `SystemSettingsScreen` แบบไม่มี `language` prop) แล้วกด "เลือกภาษา" → console ขึ้น `Maximum update depth exceeded` ซ้ำ ๆ, network ยิง `/api/language/th` และ `/api/language/ja` สลับกัน (วัดได้ 611 ครั้ง/6 วินาที) และจอค้าง "กำลังโหลด..."; เปิดจอเดียวกันผ่านเมนูหลัก (embedded, มี `language` prop) ไม่เป็น

**Root cause (ยังไม่แก้)** — `frontend/src/app/system-settings/system-settings-screen.tsx`:
- effect A (บรรทัด ~1309): อ่าน `localStorage.user_language` แล้ว `setLanguage(saved)` เมื่อไม่มี `externalLanguage` — deps มี `loadRecords`
- `loadRecords` (useCallback ~1083–1207) มี `language` ใน deps → ทุกครั้งที่ภาษาเปลี่ยน effect A รันซ้ำ
- effect B (~1402) เขียน `localStorage.user_language = language` แต่ประกาศ**หลัง** A จึงรันทีหลังใน commit เดียวกัน
- ลำดับ: เลือก ja → A อ่านค่าเก่า th → `setLanguage("th")` → B เขียน "ja" → render ถัดไป A อ่าน "ja" → `setLanguage("ja")` → B เขียน "th" → วนไม่จบ

**แนวทางแก้ (เสนอ)** — ให้ A อ่าน localStorage เพื่อตั้งค่าเริ่มต้น**ครั้งเดียวตอน mount** (ref flag) หรือย้ายการอ่านไปไว้ใน `useState` initializer แล้วตัด `setLanguage` ออกจาก effect ที่ผูกกับ `loadRecords`; ห้ามให้ `loadRecords` เป็นตัวจุด effect ที่แตะ `language`

**วิธี reproduce** — dev server → login demo → เปิด `http://localhost:3000/employee` → กดเลือกภาษา → English → ดู console/network; ตรวจว่าหายด้วย `window.__langFetch` counter (patch `window.fetch`) ต้องไม่เกิน 2 ครั้งต่อการเปลี่ยนภาษา

พบระหว่างงาน i18n จอ system-settings 2026-09-14 (README log วันเดียวกัน) — ไม่ได้แก้ในรอบนั้นเพราะอยู่นอกขอบเขต (กฎ SCOPE DRIFT) และจอตั้งค่าส่วนใหญ่เปิดจากเมนูหลักซึ่งไม่โดน
