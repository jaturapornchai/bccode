---
date: 2026-07-02
status: accepted
tags: [bc-account, architecture, mongodb, bom, recipe]
---

# สูตรผลิต (BOM) สูตรย่อยเปลี่ยนจาก snapshot เป็น live reference

## Context
ลุงจืดถาม: "สูตรย่อย ทำเป็น BOM ซ้อน BOM ได้หรือไม่ เพราะเวลาแก้สูตรย่อย จะได้ปรับสูตรใหญ่ทุกตัวเลย" — ระบบเดิมซ้อนได้ไม่จำกัดชั้นอยู่แล้ว (โครงสร้างเป็น recursive) แต่แก้สูตรย่อยไม่ปรับสูตรใหญ่อัตโนมัติ เพราะระบบ**ก็อปปี้สแนปช็อต**ตอนเพิ่มสูตรย่อยเข้าไป ไม่ใช่การอ้างอิงแบบ live. Trade-off ที่คุยกันก่อนเปลี่ยน: เสียความแม่นยำเชิงประวัติศาสตร์ (เอกสารเก่าจะไม่ freeze ค่าตอนนั้น) แลกกับความสะดวก. ลุงจืด confirm "เปลี่ยนเลย" พร้อมกำชับเรื่องกัน loop.

## Decision
เปลี่ยนสูตรใหญ่ให้เก็บแค่ **reference** (barcode+qty+yield) สำหรับ ingredient ที่เป็น reftype="recipe" ไม่ฝัง bom[] เต็ม — resolve เนื้อหาสดทุกครั้งที่อ่าน (`InfoBOM`) ผ่าน recursive resolver พร้อม cycle-guard 2 ชั้น (save-time เช็คก่อนรับ + read-time depth-cap กันค้าง).

พบและแก้บั๊กคำนวณต้นทุนที่ซ่อนอยู่ระหว่างทาง: ต้นทุนสูตรย่อยไม่เคยหารด้วย output quantity ของตัวมันเอง (เอา rollup ทั้งแบทช์มาคิดตรงๆ) — เพิ่ม `subrecipeoutputqty` แก้ให้ถูก.

## Alternatives
- **คงสแนปช็อตไว้ + เพิ่มปุ่ม "sync สูตรย่อย" แบบ manual**: ปลอดภัยกว่า (ไม่มี race condition) แต่ไม่ตรงคำขอ ("แก้ครั้งเดียว ปรับทุกตัว" อัตโนมัติ) — ไม่เลือก
- **Reference + freeze ตาม version ("ใช้ version ที่ active ตอนบันทึกเดิม")**: รักษาความแม่นยำเชิงประวัติศาสตร์ได้ดีกว่า แต่ซับซ้อนกว่ามาก (ต้อง version-pin แต่ละ reference) — ไม่เลือกในรอบนี้ เพราะลุงจืดต้องการ "ปรับสูตรใหญ่ทุกตัวเลย" ตรงๆ

## Consequences
- ✅ Verify ครบ 3 ระดับ (API/MongoDB/UI จริง) — แก้สูตรน้ำปรุงรสตัวเดียว ผัดไทย+ส้มตำอัปเดตทันทีไม่ต้อง resave
- ✅ กัน loop พิสูจน์จริงด้วย unit test ยิงเข้า pre-existing cycle ตรงๆ (goroutine+timeout) — ไม่ค้าง ทั้ง save-time reject (0.28s) และ read-time (unit test <3s)
- ✅ มีสูตรตัวอย่างไทยจริงให้ดู (ผัดไทย/ส้มตำ/ข้าวผัด) ใน dev DB
- ⚠️ **Race condition ที่ยังไม่ปิด**: 2 สูตรบันทึกพร้อมกันอ้างกันเองอาจหลุดผ่าน cycle-check ทั้งคู่ (ไม่มี Mongo transaction/lock) — ความน่าจะเป็นต่ำมาก + มี read-time safety net (ไม่ค้างแม้เกิดจริง แค่ error) แต่เป็นช่องโหว่ทางทฤษฎีจริง ต้องเพิ่ม transaction/optimistic-lock ถ้าจะปิดสนิท (ยังไม่ทำ, นอก scope)
- ⚠️ เสียความแม่นยำเชิงประวัติศาสตร์ตามที่คุยกันไว้ — เอกสารผลิตเก่าจะโชว์เนื้อหาสูตรย่อยปัจจุบัน ไม่ใช่ตอนที่ผลิตจริง (ยอมรับแล้วตอนคุย trade-off)
- ⚠️ ต้นไม้กว้างมาก (ไม่ cyclic) อาจช้า — bound ด้วย depth-cap+context-timeout, ไม่ค้างถาวรแต่อาจ timeout ถ้าสูตรซับซ้อนเกินจริง (unlikely)

related: [[2026-07-02-warehouse-subroutes-4-bugs]], [[2026-07-02-product-put-wipes-tenant-identity]]
