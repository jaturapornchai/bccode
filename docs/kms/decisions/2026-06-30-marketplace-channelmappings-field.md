---
tags: [decision, bc-account, marketplace, datamodel]
date: 2026-06-30
---

# Marketplace mapping = single product/SKU master (channelmappings reverted)

## Context
Jead: ระบบรองรับ marketplace (Shopee/Lazada/TikTok/AliExpress) ข้อมูลเข้า-ออก 100%, ต้องอ้างเอกสารทางการ. Design principle จาก Jead: **"master มีตัวเดียว รองรับทุก marketplace เพราะใช้ยอดคงเหลือเดียวกัน แต่ราคาอาจต่างกัน ยอดพร้อมขายต่างกัน"**.

## Decision
**Marketplace mapping ผูกกับ product/SKU master ตัวเดียว** (ไม่ duplicate ต่อ channel). หลักการ:
- **1 master / 1 ยอดคงเหลือ (accounting stock)** ร่วมทุก channel.
- **ราคาต่อ channel** ต่างกันได้.
- **ยอดพร้อมขายต่อ channel** ต่างกันได้ (slice/buffer ของ stock เดียว).

นี่คือสิ่งที่ **contract เดิมทำอยู่แล้ว** — ใช้/ต่อยอดอันนี้ ห้ามสร้าง marketplace structure ขนานบน entity อื่น:
- Go `MarketplaceProductMap` / `MarketplaceSKUMap` / `MarketplaceDimensionStock` — `backend/internal/product/product/models/product.go` (+ `productbarcode/models/product_barcode.go`, มี test).
- Frontend `frontend/src/lib/product-barcode/types.ts`; UI `tab-product-marketplace.tsx`, `marketplace-screen.tsx`; ราคาต่อ channel `/channelprice`.
- ราคาต่อ channel = `MarketplaceSKUMap.customprice`/`platformprice`; ยอดพร้อมขายต่อ channel = `MarketplaceDimensionStock.availableqty`/`oversellbufferqty`; stock ร่วม = accounting stock ของสินค้า.

## Reverted
ก่อนหน้านี้เผลอเพิ่ม `channelmappings` (structured field + nested SKU map) บน `productvariantmatrix` (option-set TEMPLATE) — **ซ้ำซ้อน + ผิด entity** (template ไม่มี stock, marketplace ผูกกับสินค้าจริง). **Revert ออกหมดแล้ว** (frontend `system-setting-screens.ts` + `system-settings-screen.tsx`), tsc 0.
- เก็บไว้ (แยกประเด็น, มีประโยชน์เอง): dimension `items` default `[]` fix ([[2026-06-30-dimension-items-json-object-vs-array]]); json default placeholder-aware (`[`-placeholder → `[]`).
- บทเรียน: เช็ค product/barcode datamodel ก่อนเพิ่ม field marketplace — มี contract สมบูรณ์อยู่แล้ว.

## Docs
- Rule: `.agents/rules/bc-account-core-rules.md` "System Scope & Marketplace Support".
- Field mapping (contract ↔ 4 marketplace API จริง + Shopee example + official URLs): `.agents/wiki/marketplace-field-mapping.md`.

## Still NOT built
Connectors จริง (OAuth/token, API client ต่อเจ้า, sync push/pull, category+attribute fetch). Schema พร้อม — 100% in/out ทำได้เมื่อสร้าง connector + verify field กับ live doc ตอน build.
