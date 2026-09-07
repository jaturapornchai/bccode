---
date: 2026-09-03
status: accepted
tags: [bc-account, go, mongodb, architecture]
---

# tier/init + tier/update รับ "สถานะสุดท้ายทั้งชุด" แทนการ remap ด้วยชื่อตัวเลือก

## Context

API v2 ชั้นลงขาย (ดู [[2026-09-03-product-two-layer-marketplace-model]]) ต้องให้เจ้าของกิจการกำหนดชั้นตัวเลือกสินค้า (สี/ขนาด ≤2 ชั้น, ≤50 ชุดผสม)
โดย 1 ชุดผสม = 1 เอกสาร `productbarcode` ซึ่งเป็นกุญแจสต๊อกจริง (stockprocess คีย์ที่ barcode+whcode)

ร่างออกแบบเดิมให้ `tier/update` จับคู่ตัวเลือกเดิมกับตัวเลือกใหม่ด้วย **ชื่อ option** ปัญหาคือถ้าผู้ใช้แก้คำผิด (เช่น "แดวง" → "แดง")
ระบบจะมองว่าเป็นการลบตัวเลือกเดิม + เพิ่มตัวใหม่ → บาร์โค้ดหลุดตำแหน่ง สินค้าที่ขายอยู่ถูกปิดขายโดยไม่ตั้งใจ
ทางแก้ที่เสนอกันคือเพิ่ม `optionid` ถาวรลง data model

## Decision

ทั้ง `tier/init` และ `tier/update` รับ body รูปแบบเดียวกัน = **สถานะสุดท้ายทั้งชุด**:
`tiers[]` ใหม่ทั้งหมด + `models[]` ที่ครบทุกชุดผสมพอดี โดยผู้เรียกระบุเองว่าชุดผสมไหนใช้บาร์โค้ดไหน (`tierindex` + `barcode` หรือ `generate:true`)

เซิร์ฟเวอร์จึงไม่ต้องเดาความสัมพันธ์เก่า-ใหม่เลย และ **ไม่ต้องเพิ่ม `optionid` ใน data model**
ชุดผสมเดิมที่ไม่อยู่ในคำขอ → `listing.tierindex=[]` + `listing.isforsale=false` (ปิดขาย ไม่ลบเอกสาร เพราะผูกประวัติสต๊อก)

## Alternatives

- **remap ด้วยชื่อ option** — พังเมื่อ rename ตามที่อธิบายข้างบน
- **เพิ่ม `optionid` (uuid) ต่อ option** — แก้ปัญหาได้ แต่เพิ่มฟิลด์ถาวรใน schema + ภาระให้ทุก client ต้องเก็บ id และเป็น YAGNI เมื่อหน้าจอส่งสถานะเต็มอยู่แล้ว
- **แยก endpoint ย่อย (add/rename/remove option)** — API เยอะขึ้น 3 เท่า และยังต้องแก้ปัญหา atomicity เดิม

## Consequences

- ผู้เรียกต้องส่ง payload ใหญ่ขึ้น (ทุกชุดผสม) แต่เพดานคือ 50 ชุด จึงเล็กมาก
- หน้าจอต้องโหลดสถานะปัจจุบันก่อนแก้เสมอ (`item/get`) ซึ่งเป็นสิ่งที่ต้องทำอยู่แล้วเพื่อได้ `__v`
- การเขียนเป็น transaction เดียว (`products` + `productbarcodes`); MongoDB standalone ไม่รองรับ transaction → fallback เขียนเรียงลำดับพร้อม `logger.Warn` เพื่อให้ dev local ใช้งานได้
- บาร์โค้ดที่ระบบสร้าง = prefix `200` + สุ่ม 9 หลัก + check digit; ยังไม่คัดลอก `prices` จากบาร์โค้ดหลัก (รอลุงจืดตัดสิน)
- ไฟล์: `backend/internal/goapi/handlers/product_v2_tier.go`, `product_v2_tier_validate.go` (+test 11 เคส), contract `docs/kms/architecture/product-listing-api-v2.md` §5.4

#bc-account #go #mongodb
