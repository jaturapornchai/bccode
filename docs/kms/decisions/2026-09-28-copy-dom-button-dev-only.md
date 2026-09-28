---
date: 2026-09-28
status: accepted
tags: [bc-account, ui, developer-experience, dev-dom-inspector, production]
supersedes: บางส่วนของ `2026-09-12-enable-dom-inspector-on-production.md` (ปุ่มลอยบน production เท่านั้น — Alt+คลิกยังอยู่)
---

# ปุ่มลอย Copy DOM แสดงเฉพาะตอน dev — Alt+คลิกคัดลอก DOM ยังใช้บน production ได้

## Context
- 2026-09-12 ลุงจืดสั่งให้ `account.bcaicloud.com` มี Copy DOM เหมือน localhost จึงถอด `NODE_ENV` gate ทั้งคอมโพเนนต์ `DevDomInspector` ออก (ADR [[2026-09-12-enable-dom-inspector-on-production]]) — บน production จึงมีทั้ง Alt+คลิกและปุ่มลอย `Copy DOM [Alt+คลิก]` มุมล่างซ้าย
- 2026-09-28 ลุงจืดเลือกให้ **ซ่อนปุ่มลอยบน production** เพราะ:
  1. ปุ่มอยู่ `fixed bottom-4 left-4 z-[999999]` — บนจอเตี้ยไปบังปุ่มบันทึกของฟอร์มที่อยู่มุมล่างซ้าย
  2. คลิกปุ่มครั้งเดียวเข้าโหมดคัดลอกต่อเนื่อง (Toggle Mode) ที่ทุกคลิกกลายเป็น Copy DOM แทนการทำงานปกติ — ผู้ใช้จริงกดโดนโดยไม่ตั้งใจแล้วจอ "กดอะไรไม่ติด"
- พบบั๊ก Alt ค้างในคราวเดียวกัน: Alt+Tab ไปโปรแกรมอื่น (นักบัญชีสลับไป Excel บ่อย) ทำให้ keyup ของ Alt ไปตกที่หน้าต่างอื่น state `altHeld` ค้างเป็น true → คลิกปกติครั้งถัดไปใน BC ถูก `preventDefault` + `stopImmediatePropagation` แล้วคัดลอก DOM แทน (รายละเอียด [[2026-09-28-copy-dom-alt-stuck-after-alt-tab]] ใน `bugs/`)

## Decision
1. **ปุ่มลอยเป็น dev-only** — ครอบเฉพาะ wrapper ปุ่มด้วย `{process.env.NODE_ENV === "development" && (` (`frontend/src/components/dev-dom-inspector.tsx:251`); Next แทนค่า `NODE_ENV` ตอน build เป็น `"production"` ทุกคำสั่งยกเว้น `next dev` กิ่งนี้จึงหายจาก bundle ของ production; บน production ไม่มีทางเข้า Toggle Mode เพราะ `setActive(next)` อยู่ในปุ่มนี้เท่านั้น (Escape ตั้งได้แค่ `false`)
2. **Alt+คลิกคัดลอก DOM คงไว้บน production** — listener คีย์ (:70-114) และเมาส์ (:117-186) ไม่มี gate ยังดักคลิกใน capture phase + `preventDefault`/`stopPropagation`/`stopImmediatePropagation` + กรอบไฮไลต์ + toast เหมือนเดิม
3. **ตัดสินจาก event ไม่ใช่ state ที่จำไว้** — `handleClick` ใช้ `if (!activeRef.current && !e.altKey) return;` (:156); `handleMouseMove` ซิงก์ `altHeld` จาก `e.altKey` เมื่อไม่ตรงกัน (:120-126) state จึงซ่อมตัวเองแม้ blur ไม่เกิด
4. **รีเซ็ต Alt เมื่อหน้าต่างเสียโฟกัส** — `resetAlt` (:91-97) ผูกกับ `window` `blur` และ `document` `visibilitychange` ตอน `visibilityState === "hidden"` (:99-112) ถอด listener ครบใน cleanup
5. ทั้งไฟล์มี `process.env.NODE_ENV` ได้ **จุดเดียว** — test บังคับ (`frontend/src/components/dev-dom-inspector.test.ts:13-32`)

## Alternatives considered
- **ถอด DevDomInspector ออกจาก production ทั้งหมด** (ใส่ gate กลับทั้งคอมโพเนนต์) — ไม่เลือก: ลุงจืดยังใช้ Alt+คลิกคัดลอก DOM จากจอจริงส่งให้ AI แก้ (ที่มาของ ADR 2026-09-12)
- **ย้ายปุ่มไปมุมอื่น / ย่อให้เล็กลง** — ไม่เลือก: มุมไหนก็มีโอกาสบังปุ่มของบางจอบนจอเตี้ย และยังเหลือ Toggle Mode ที่ผู้ใช้จริงกดโดนแล้วติดโหมดได้; ผู้ใช้ทั่วไปไม่ต้องใช้ปุ่มนี้เลย

## Consequences
- บน production ไม่มีสิ่งใดบนจอบอกว่า Alt+คลิกคัดลอก DOM ได้ — ตั้งใจตามลุงจืด แต่ทีมที่ช่วยผู้ใช้ต้องรู้คีย์ลัดนี้เอง
- พฤติกรรมใหม่: คลิกที่กด Alt ค้างถูกดักแม้ BC ไม่เคยเห็น keydown ของ Alt (เช่น กด Alt ก่อนหน้าต่างได้โฟกัส); Alt+คลิกที่เป็นคีย์ลัดของจอในแอปยังถูกกลืนเหมือนเดิม
- **ข้อจำกัดที่รู้แล้ว (ยังไม่แก้):** AltGr — เดิม `altHeld` ตั้งจาก `e.key === "Alt"` เท่านั้น (AltGr รายงานเป็น `"AltGraph"`) AltGr+คลิกจึงไม่เคยถูกดัก; ตอนนี้เชื่อ `MouseEvent.altKey` ซึ่งบน Windows AltGr ส่งเป็น LCtrl+RAlt จึงน่าจะถูกดักเป็น Copy DOM บนแป้นยุโรป (ยังไม่ได้ทดสอบจริง; แป้นไทยเกษมณีไม่ใช้ AltGr) — ถ้าต้องแก้: `const altCopy = e.altKey && !e.getModifierState("AltGraph")` ทั้งใน click และ mousemove
- Alt+คลิกยังยิง `pointerdown`/`mousedown` ของ element (เช่น trigger ของ Radix) ก่อนคลิกถูกกลืน — มีมาก่อนการเปลี่ยนครั้งนี้
- test เป็นแบบอ่านซอร์ส (vitest environment `node` ไม่มี jsdom) ตรวจแค่ลำดับ index ของ gate กับปุ่ม — ยังไม่ pin ว่า `resetAlt` เรียก `setAltHeld(false)` และบรรทัดซิงก์ `if (altStale) setAltHeld(e.altKey);`
- ADR นี้ทับเฉพาะเรื่องปุ่มลอยของ [[2026-09-12-enable-dom-inspector-on-production]]; เหตุผลเรื่องความปลอดภัย (คัดลอกเฉพาะ `outerHTML` ในเครื่อง ไม่ส่งออก) ยังใช้ตามเดิม
- skill `ui-scale-polish` §8.17 มีข้อแก้ไขต้นหัวข้อแล้ว

## Evidence
- vitest `src/components/dev-dom-inspector.test.ts` 8/8 ผ่าน; `tsc --noEmit` 0 error; eslint 0 error (warning 2 ตัวเดิมที่ :58, :61 — warning exhaustive-deps เรื่อง `highlight` หายไป)
- ทดสอบใน browser บน `next dev` (localhost:3000/login) ด้วย synthetic event: ปุ่มแสดงใน dev; หลัง `window` blur คลิกปกติไม่ถูกดัก; Alt ค้าง (keydown แต่ไม่มี keyup) คลิกปกติไม่ถูกดักและ mousemove รีเซ็ตป้าย; คลิกที่ `altKey: true` ถูกดัก; Toggle Mode + Escape ทำงาน; `visibilitychange` hidden รีเซ็ต
- ยังไม่ได้ `next build` ในเครื่อง (ห้ามรันใน tree ที่ใช้ร่วมกันเพราะทับ `.next` ของ dev) — ตรวจหลัง deploy: `document.querySelectorAll('[data-dev-dom-inspector]').length === 0` บน account.bcaicloud.com ขณะไม่กด Alt
- (เติมหลังทดสอบ)
