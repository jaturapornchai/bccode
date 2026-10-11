---
date: 2026-10-11
status: accepted
tags: [bc-account, stock, cost, champ-parity]
---

# ต้นทุนสินค้าใช้แบบถัวเฉลี่ยอย่างเดียว

## Context
โค้ดยังมีเครื่องคิดต้นทุนหลายวิธีค้างอยู่ใน `backend/internal/goapi/inventory/costing/` (FIFO, LIFO, FEFO, standard, moving average) ค่าคงที่ 7 วิธีใน `backend/internal/goapi/inventory/models/models.go:9-17` route ตั้งค่าต้นทุนลงทะเบียนที่ `backend/internal/goapi/bootstrap.go:339` และข้อความเมนู/เครื่องมือ/รายงานที่อ้าง FIFO/FEFO ใน `backend/assets/language/languages.tsv`

## Decision
ลุงจืดสั่ง 2026-10-11 ว่า "เอาแค่ต้นทุนแบบ average ไม่เอา fifo และอื่นๆ"
- ต้นทุนสินค้าคิดแบบถัวเฉลี่ยวิธีเดียว และคำนวณแบบที่ Champ คำนวณ
- เอา FIFO, LIFO, FEFO, lot/specific, standard และวิธีอื่นออกทั้งหมด ทั้ง engine, ค่าตั้ง, ตัวเลือก, เมนู, เครื่องมือ, คอลัมน์รายงาน และข้อความ
- คำสั่งนี้อยู่เหนือหลัก Champ parity ในเรื่องนี้ (ตัดได้เพราะเป็นคำสั่งเป็นลายลักษณ์อักษร)

## Alternatives
- เก็บหลายวิธีไว้ให้เลือกตามบริษัท: ไม่เลือก เพราะลุงจืดสั่งตัด และผู้ใช้กลุ่มเป้าหมาย (SME) ใช้ถัวเฉลี่ยเป็นหลัก ระบบที่มีหลายวิธีต้องดูแลชั้นต้นทุน (cost layer) เพิ่มโดยไม่มีคนใช้

## Consequences
- งานลบ engine/ตัวเลือกที่ไม่ใช่ถัวเฉลี่ยต้องทำในการเปลี่ยนแปลงเดียว และห้ามเพิ่มกลับมา (กฎใน `AGENTS.md` หัวข้อ Stock cost)
- รายงานที่มีคอลัมน์ "ต้นทุน FIFO" ให้เหลือเฉพาะต้นทุนถัวเฉลี่ย
