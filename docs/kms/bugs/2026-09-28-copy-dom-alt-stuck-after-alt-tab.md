---
date: 2026-09-28
severity: medium
component: [frontend]
tags: [bc-account, ui, dev-dom-inspector, production, keyboard]
fixed: true
---

# Symptom

บน production (account.bcaicloud.com) หลังกด Alt+Tab ไปโปรแกรมอื่น (เช่น Excel) แล้วกลับมา คลิกปกติครั้งถัดไปใน BC "ไม่ทำงาน" — ปุ่ม/ลิงก์ไม่ตอบสนอง แต่มี toast คัดลอก DOM ขึ้นแทน; นอกจากนี้ปุ่มลอย `Copy DOM` มุมล่างซ้าย (`z-[999999]`) บังปุ่มบันทึกของฟอร์มบนจอเตี้ย และคลิกโดนครั้งเดียวทำให้ทุกคลิกกลายเป็น Copy DOM (Toggle Mode)

## Root Cause

- `DevDomInspector` (`frontend/src/components/dev-dom-inspector.tsx`) ตัดสินว่าจะกลืนคลิกจาก state `altHeld` ที่จำไว้จาก keydown/keyup ของ Alt — Alt+Tab ทำให้ keyup ของ Alt ไปตกที่หน้าต่างอื่น `altHeld` จึงค้างเป็น true และ `handleClick` (capture phase) เรียก `preventDefault` + `stopPropagation` + `stopImmediatePropagation` กับคลิกปกติครั้งถัดไป
- ไม่มีการรีเซ็ตเมื่อหน้าต่างเสียโฟกัสหรือแท็บถูกซ่อน
- บรรทัด `if (highlight) setHighlight(null)` ใน mousemove ไม่เคยทำงาน: effect มี deps `[mounted]` closure จึงเห็น `highlight` = null ของรอบแรกตลอด (listener ต้องอ่านค่าสดผ่าน ref)
- ปุ่มลอยถูกเปิดบน production ทั้งคอมโพเนนต์ตาม ADR 2026-09-12 โดยไม่ได้แยกปุ่มออกจากฟีเจอร์ Alt+คลิก

## Fix

- `handleClick` ตัดสินจาก event เอง: `if (!activeRef.current && !e.altKey) return;` (`dev-dom-inspector.tsx:156`)
- `resetAlt` (:91-97) — `setAltHeld(false)` + ล้างไฮไลต์ถ้าไม่อยู่ใน Toggle Mode; ผูกกับ `window` `blur` และ `document` `visibilitychange` เมื่อ `visibilityState === "hidden"` (:99-112) ถอดครบใน cleanup
- `handleMouseMove` ซิงก์ `altHeld` จาก `e.altKey` เมื่อไม่ตรงกัน (`altStale`, :120-126) — state ซ่อมตัวเองแม้ blur ไม่เกิด; ล้างไฮไลต์เฉพาะตอนซิงก์ (แทนบรรทัด `if (highlight)` ที่ตาย — warning exhaustive-deps หายไปด้วย)
- ปุ่มลอยครอบด้วย `{process.env.NODE_ENV === "development" && (` (:251) — บน production ไม่มีปุ่มและเข้า Toggle Mode ไม่ได้; listener คีย์/เมาส์ (:70-186) ไม่มี gate ตามเดิม (ADR `decisions/2026-09-28-copy-dom-button-dev-only.md`)
- ข้อจำกัดที่ยังไม่แก้: AltGr+คลิกบนแป้นยุโรป (Windows ส่ง LCtrl+RAlt → `MouseEvent.altKey` true) น่าจะถูกดักเป็น Copy DOM แล้ว ทั้งที่เดิมไม่ถูก (ยังไม่ทดสอบจริง; แป้นไทยไม่ใช้ AltGr) — วิธีแก้ถ้าต้องการ: `e.altKey && !e.getModifierState("AltGraph")`

แบบแผนนำไปใช้ซ้ำกับฟีเจอร์ที่ใช้ปุ่ม modifier: (a) ตัดสินต่อ event จาก flag ของ event นั้น (b) รีเซ็ต state ของ modifier ตอน `blur` / `visibilitychange` hidden (c) ซิงก์ state จาก event ที่มาถึงถัดไป

## Regression Test

- `frontend/src/components/dev-dom-inspector.test.ts` (อ่านซอร์ส 8 เทสต์): `process.env.NODE_ENV` มีจุดเดียวและอยู่หลัง click listener; wrapper `fixed bottom-4 left-4 z-[999999]` และ `setActive(next)` อยู่หลัง gate, ไม่มี `setActive(true)`; click ใช้ `!e.altKey` และไม่เช็ค `altHeldRef` แล้ว; เพิ่ม/ถอด listener `blur` + `visibilitychange` และมีเช็ค `visibilityState === "hidden"`; Escape ยังปิด Toggle Mode; ของเดิม (ข้อความปุ่ม, clipboard fallback, ดักคลิก capture phase) ยังอยู่
- `npx vitest run src/components/dev-dom-inspector.test.ts` 8/8 ผ่าน; `tsc --noEmit` 0 error; eslint 0 error
- browser บน `next dev` localhost:3000/login ด้วย synthetic event: หลัง blur คลิกปกติไม่ถูกดัก (`defaultPrevented` false); Alt ค้าง (keydown ไม่มี keyup) คลิกปกติไม่ถูกดัก และ mousemove ปกติรีเซ็ตป้าย; คลิก `altKey: true` ถูกดัก; Toggle Mode + Escape ทำงาน; `visibilitychange` hidden รีเซ็ตป้าย
- ช่องว่างของเทสต์: ยังไม่ pin ว่า `resetAlt` เรียก `setAltHeld(false)` และบรรทัด `if (altStale) setAltHeld(e.altKey);`; เช็คว่าปุ่มอยู่ "ใน" gate ด้วยลำดับ index เท่านั้น (vitest environment `node` ไม่มี jsdom)
- ยังไม่ได้ยืนยันบน production build — หลัง deploy: `document.querySelectorAll('[data-dev-dom-inspector]').length === 0` บน account.bcaicloud.com ขณะไม่กด Alt; Alt+Tab จริงไปโปรแกรมอื่นจำลองด้วย `window` blur เท่านั้น
