---
tags: [decision, bc-account, marketplace, datamodel]
date: 2026-06-30
status: superseded
---

# Marketplace mapping = single product/SKU master (channelmappings reverted)

> ⚠️ **ถูกแทนที่โดย** [2026-09-03-product-two-layer-marketplace-model.md](2026-09-03-product-two-layer-marketplace-model.md) — หลักการ "master ตัวเดียว/สต๊อกเดียว" ยังใช้ได้ แต่ที่เก็บข้อมูลช่องทางย้ายไป collection `channel_shops`/`channel_listings`/`channel_category_maps`/`channel_brand_maps` แล้ว (`backend/internal/goapi/handlers/product_v2_types.go:20-23`) ส่วน Go `MarketplaceProductMap`/`MarketplaceSKUMap`/`marketplaceproducts[]`/`marketplaceskumappings` (`backend/internal/product/product/models/product.go:98,163,408,466`) **ถูกกำหนดให้ deprecate แต่ยังรอลุงจืดยืนยัน จึงยังอยู่ในโค้ด — ห้ามต่อยอดเพิ่ม และห้ามลบทิ้งเอง**

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
- Rule: `.agents/rules/bc-account-core-rules.md` (หัวข้อ "System Scope & Marketplace Support") **ถูกลบไปแล้วที่ commit `07c9c21c` (2026-08-23) และยังไม่มีกฎ marketplace ใน `AGENTS.md` ปัจจุบัน** — ถ้าต้องการขอบเขต/กฎ ให้ถามลุงจืด ห้ามเดา
- Field mapping (contract ↔ marketplace API จริง): เอกสารเดิม `.agents/wiki/marketplace-field-mapping.md` **ถูกลบพร้อมกันที่ commit `07c9c21c` ยังไม่มีไฟล์ทดแทนใน `docs/`** — สัญญาที่ใช้อ้างอิงตอนนี้คือ `docs/kms/architecture/product-listing-api-v2.md`

## Still NOT built
Connectors จริง (OAuth/token, API client ต่อเจ้า, sync push/pull, category+attribute fetch). Schema พร้อม — 100% in/out ทำได้เมื่อสร้าง connector + verify field กับ live doc ตอน build.
