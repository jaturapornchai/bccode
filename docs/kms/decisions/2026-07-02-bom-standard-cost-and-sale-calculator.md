---
date: 2026-07-02
status: accepted
tags: [bc-account, architecture, bom, recipe, costing]
---

# สูตรผลิต (BOM) เพิ่ม Finished Good / Labor-OH-Scrap / Standard-vs-Current Cost + calculator ขายตามจำนวน

## Context
ลุงจืดยกตัวอย่าง: "ขายส้มตำ 10 จาน ต้องการรู้ว่าประกอบด้วยต้นทุนอะไรบ้าง วัตถุดิบอะไรบ้าง ราคาต้นทุนอัตโนมัติ หรือ STANDARD COST ต้องมีให้เลือก" — ปรึกษา GLM+DeepSeek+GPT 5.5 อิสระ (ไม่เห็นคำตอบกัน) ได้ดีไซน์ตรงกันเกือบเป๊ะ.

## Decision
เพิ่ม 6 field ที่หัวสูตร: `finishedgoodbarcode` (link ไป product master), `laborcost`/`overheadcost`/`scrappercent` (ต้นทุนแรงงาน/โสหุ้ย/% เสียหาย ต่อ batch), `costmode` (current/standard) + `standardcost` (ค่าตายตัวเมื่อเลือก standard). Calculator "จำนวนที่จะขาย/ผลิต" reuse หน้าจอ Exploded BOM เดิม ไม่สร้างใหม่ — ไม่ทำใบสั่งผลิตเต็มรูปแบบตามที่ลุงจืดเลือก scope ไว้ก่อนหน้า.

สูตร: current mode = `(rollup+labor+OH)/outputqty/(1-scrap%)`, standard mode = ใช้ `standardcost` ตรงๆ (วัตถุดิบยังโชว์ไว้อ้างอิง).

## Bug ที่เจอ + แก้ (สำคัญ)
Adversarial review เจอ: สูตรย่อยที่ถูกอ้างอิงในสูตรใหญ่ **ไม่ propagate labor/OH/scrap/costmode ของตัวเอง** ไปให้สูตรใหญ่เห็น — แก้สูตรย่อยให้มีต้นทุนแรงงานแล้ว สูตรใหญ่จะคำนวณต่ำกว่าจริงเงียบๆ. Root cause: `resolveRecipeTree` (backend) copy แค่ Names/Unit/BOM ไม่ copy 5 field ใหม่. แก้แล้ว + สร้าง shared function `effectiveUnitCost()` (frontend) ใช้สูตรเดียวกันทั้ง root และ nested — กันบั๊กประเภทนี้ไม่ให้เกิดซ้ำ (เคยเจอบั๊กคล้ายกันกับ `subrecipeoutputqty` มาก่อนแล้วในการแก้ live-reference ครั้งก่อน — เป็น pattern ที่ต้องระวังทุกครั้งที่เพิ่ม field ใหม่บนสูตรย่อย: ต้องเช็คว่า resolve เข้า parent ด้วยไหม).

Verify การแก้: ตั้ง sauce labor=5/OH=3/scrap=10% → หน้าส้มตำเห็น unit cost ฿75.61 ตรงเลขคำนวณมือเป๊ะ (เดิมจะโชว์ ฿60.05 ผิด).

## Consequences
- ✅ ตรงตัวอย่างของ Jead 100% (ส้มตำ 10 จาน คำนวณตรงเลขมือ ทั้ง current และ standard mode)
- ✅ แก้บั๊ก field-propagation ก่อนส่งมอบ (ไม่ปล่อยให้ silent-wrong-cost หลุดไปใช้งานจริง)
- ⚠️ Standard cost เก็บที่ระดับหัวสูตรเท่านั้น (ไม่มี per-line standard cost ตามที่ DeepSeek/GPT แนะนำเสริม) — ตัดสินใจ scope ให้ง่ายก่อน ขยายทีหลังได้ถ้าจำเป็นจริง
- ⚠️ ยังไม่เชื่อมกับ flow การขายจริง (POS/ใบเสร็จ) — เป็น calculator ในจอ BOM เท่านั้น ตามที่ตกลง scope ไว้ (ไม่ทำใบสั่งผลิต)

related: [[2026-07-02-bom-subrecipe-live-reference]]
