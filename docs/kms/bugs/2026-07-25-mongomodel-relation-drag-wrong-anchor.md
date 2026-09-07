# 2026-07-25 — mongomodel: ลากเส้น relation แล้วเส้นออกไม่ตรงจุดที่กด

#mongomodel #react-flow #ux

## Symptom
กดจุดเชื่อม (handle) ในการ์ด collection แล้วลากสร้าง relation ใหม่ → เส้นที่ลากออกมา**เริ่มจากจุดอื่น** ไม่ใช่จุดที่กด

## Root cause (วัดจาก react-flow store จริง)
1. **`connectionRadius={48}`** — react-flow ใช้รัศมีนี้หาจุดเชื่อมที่ใกล้ที่สุดตอนกด แต่ระยะจริงระหว่าง handle เล็กกว่านั้นมาก:
   - source ↔ target ของ field เดียวกัน = **14px**
   - handle ของ field แถวถัดไป = **~39px**
   รัศมี 48 กินทั้งคู่ → คว้า handle ตัวข้างเคียงที่ระยะสั้นกว่าแทนตัวที่กด
2. **จุดเชื่อมเล็กแค่ 6px ทั้งที่ CSS สั่ง 10px** — `globals.css` เขียน `.react-flow__handle { width: 10px }` แต่ stylesheet ของ react-flow มี specificity เท่ากัน (0,1,0) และโหลดทีหลัง จึงชนะ; computed จริง = 6px, border 0.8px → เป้าเล็ก กดพลาดง่าย (CSS นี้ไม่เคยมีผลเลยทั้งเวอร์ชันเก่าและใหม่)

## Fix
- `app/page.tsx`: `connectionRadius={10}` (แคบกว่าครึ่งของ 14px = คว้าได้เฉพาะจุดที่กด)
- `app/globals.css`: เปลี่ยน selector เป็น `.react-flow .react-flow__handle` (specificity 0,2,0 ชนะ react-flow) → handle 10px จริง

## Verify
computed handle = 10px · `connectionRadius` = 10 · tsc ผ่าน · `demo()` regression ผ่าน · ลุงจืดลากทดสอบเอง = **ผ่าน**

## Regression guard
ยังไม่มี test อัตโนมัติ (จำลอง pointer event ของ react-flow ไม่ผ่าน guard) — ถ้าจะกัน ต้องเช็คด้วย Playwright จริง: กด handle แล้ววัดจุดเริ่มของ `.react-flow__connectionline path` เทียบ handle center

## Trade-off
รัศมี 10 ทำให้ตอน**ปล่อย**เส้นต้องเล็งแม่นขึ้น (เดิม 48 = snap ง่าย) ถ้าต่อเส้นยากขึ้น ขยับได้ถึง 13 (ต้องต่ำกว่า 14)

เกี่ยวข้อง: [[2026-07-25-names-canonical-code-name]]

---

## ต่อเนื่อง: ลูกศรถูกจุดเชื่อมทับ (เจอหลังแก้อันบน)

**Symptom:** หัวลูกศร relation มองไม่เห็น — ถูกวงจุดเชื่อมทับ

**Root cause:** react-flow ตรึง marker ไว้ที่ `refX=0` (ปลายแหลม = จุดปลาย path = ศูนย์กลาง handle พอดี, แก้ผ่าน props ไม่ได้) ลูกศรยาว 7.5px แต่จุดเชื่อมรัศมีที่มองเห็น 8px → จมทั้งดุ้น + สีเดียวกัน `#38bdf8`

**Fix:** custom edge `RelEdgeView` ขยับปลาย path ก่อนเรียก `getBezierPath` (ลูก 17px / แม่ 12px) → gap จริง 22px
⚠ ต้องบังคับ `type: "rel"` ใน `displayEdges` ด้วย ไม่ใช่แค่ `defaultEdgeOptions` — edge ที่ persist ไว้มี `type` ของตัวเองและชนะ default

**บทเรียนเรื่องการ verify:** browser pane ใน IDE ไม่ compositing frames → ResizeObserver ไม่ fire → react-flow ไม่วัด `handleBounds` → **ไม่ render edge เลย** ทำให้เข้าใจผิดว่าโค้ดพัง ต้องใช้ Playwright headless จริง (`D:/bccode/tmp/verify-edge-gap-label.mjs`)
