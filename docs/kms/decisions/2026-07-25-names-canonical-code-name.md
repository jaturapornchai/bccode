# ADR 2026-07-25 — names array canonical = {code, name}

#bc-account #go #nextjs #datamodel

## Decision
โครงสร้าง array `*names` ทุกจุดของ BC Ai Account = `{code, name}` เท่านั้น
ตัด `isauto` / `isdelete` ออกทั้งระบบ (diagram → Go → frontend → swagger)

## Context
mongomodel diagram normalize names 42 array เหลือ `{code, name}` ไปแล้ว (2026-07-23)
แต่ runtime ยังพก 2 flag นี้อยู่ → model กับ code ไม่ตรงกัน ลุงจืดสั่งให้ runtime ตาม diagram

## Consequences
- `models.NameX` / `models.NameNormal` / `LanguageModel` เหลือ 2 field
- `isdelete` **ระดับเอกสาร** (soft delete) ไม่เกี่ยว — ยังอยู่เหมือนเดิม
- payload เก่าที่ยังส่ง 2 flag มา: Go ignore field ที่ไม่รู้จัก → ไม่พัง (fixture test เดิมยังมี flag ไว้เป็น backward-compat case)
- ข้อมูลเดิมใน MongoDB ที่มี flag: ปล่อยตาย ตาม [[disposable-database-rule]] (pre-launch ไม่ทำ migration)
- codegen ของ mongomodel มี `namesShapeWarnings()` เตือนแล้วถ้ามี array ไหน drift ออกจาก shape นี้

## Verify
tsc + next build ผ่าน · vitest เท่า baseline · gofmt ทั้ง tree ผ่าน · `go build ./internal/models ./internal/goapi/models` ผ่าน
⚠ Go compile เต็ม repo ยังไม่ได้ verify — เครื่องไม่มี gcc + Docker ล่ม (confluent-kafka ต้อง CGO)
