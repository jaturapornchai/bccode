---
description: Single-file agent prompt for auditing BC Account marketplace readiness and deriving a canonical marketplace JSON/model foundation.
---

# Agent Prompt: Marketplace Foundation Audit For BC Account

You are the agent working in `D:\bccode`.

## Communication
- Reply to Jead in Thai.
- Address Jead as `ลุงจืด`.
- Keep the final report concise but evidence-backed.
- Do not expose chain-of-thought. Explain decisions with source evidence, tradeoffs, and verification only.

## Mission
Audit the BC Account product, barcode/SKU, inventory, price, media, and marketplace-related system. Compare the current implementation against real marketplace product JSON/API structures from:

- Shopee
- Lazada
- TikTok Shop
- AliExpress

The goal is to find weaknesses and design a canonical BC Account marketplace-ready standard that can support future marketplace import/export/sync for many business types such as restaurant, clothing, mobile phones, SIM cards, computers, serialized goods, options/variants, prices, stock, images, and specifications.

This is an audit/design task first. Do not implement production code unless Jead explicitly says `แก้เลย`.

## Non-Negotiable Project Rules
1. Use active repo source as the system truth:
   - `D:\bccode\AI_INDEX.md`
   - `D:\bccode\AGENTS.md`
   - `D:\bccode\.agents\rules\bc-account-core-rules.md`
   - `D:\bccode\.agents\skills\bc-account-expert\SKILL.md`
   - current frontend/backend/source/test/migration files
2. Do not use `D:\bccode-model` as a source of truth.
3. Do not create or restore files under `manual/`.
4. Do not guess APIs, schemas, enum values, model fields, workflow, or business behavior.
5. If evidence is missing, write `ยังยืนยันไม่ได้` and list exactly what was checked.
6. Do not commit secrets, tokens, passwords, marketplace credentials, or real customer data.
7. Use `holdingcode` as tenant/workspace scope and `guidfixed`/`guid` as immutable CRUD identity.
8. New persistent/API contracts must use lower `snake_case`.
9. User-facing UI must not show raw JSON for normal users. Raw payloads are allowed only in importer/debug/admin tools.
10. Marketplace stock and price must not break accounting rules:
    - Accounting stock is product-based and supports total balance, warehouse balance, and storage-location balance.
    - Costing uses accounting stock only.
    - Normal costing is product-level.
    - If `costbywarehouse` is enabled, calculate cost per warehouse first, then aggregate to product cost.
    - Do not calculate inventory cost by marketplace dimensions such as color, size, capacity, network, or SIM package.
    - Marketplace stock is a separate availability projection that can expose product-level and dimension-level available stock.
    - Selling price can be product, barcode/SKU, price level, marketplace, and detailed dimension level.

## Required Source Inspection
Start with a short plan, then inspect source using `rg` and targeted reads. Avoid broad slow scans.

Minimum local files/areas to inspect:
- `.agents/rules/bc-account-core-rules.md`
- `.agents/skills/bc-account-expert/SKILL.md`
- `frontend/src/lib/product-barcode/types.ts`
- `frontend/src/lib/product-barcode/utils.ts`
- `frontend/src/components/product-barcode/barcode-form.tsx`
- `frontend/src/app/menu/product-barcode-screen.tsx`
- `frontend/src/app/menu/tab-product-marketplace.tsx`
- `frontend/src/app/menu/product-screen.tsx`
- `frontend/src/app/menu/product-set-screen.tsx`
- `backend/internal/goapi/models/barcode-model.go`
- `backend/internal/goapi/models/mongo-barcode-model.go`
- `backend/internal/goapi/models/mongo-product-model.go`
- `backend/internal/goapi/handlers/barcode_list.go`
- `backend/internal/goapi/dataimport/xlsx_product.go`
- `backend/internal/goapi/handlers/dataimport/xlsx_product.go`
- `backend/internal/goapi/inventory/**`
- `backend/migrations/**inventory**`
- `scratch/aliexpress-product-foundation-reference.json` if present

If a listed file is missing, note it and continue with `rg` evidence.

Useful searches:
```powershell
rg -n "marketplace|shopee|lazada|tiktok|aliexpress|sku|barcode|variant|dimension|media_assets|specification|payload|stock|price" frontend backend .agents scratch
rg -n "balanceqty|availableqty|costbywarehouse|marketplacestockbalances|marketplacedimensionprices|inventorystockbalances" backend frontend
rg -n "raw json|payload|schema|matrix|mapping|GUID|route|slug" frontend/src
```

## Required External Research
Use web/docs search only for marketplace API structures and current examples. Prefer official docs and primary sources. If official docs are behind login or incomplete, use the best accessible public technical source and clearly mark it as non-official.

For every external source, record:
- Platform
- URL
- Source type: official docs, SDK, public example, community example, or inferred
- Date accessed
- What exact fields were evidenced
- Confidence level: high / medium / low

Research these payload families:

### Shopee
Find current JSON examples or schemas for:
- Product/item create and update
- Product/item detail/get
- Category attributes
- Model/variation/SKU
- Stock and warehouse/location if available
- Price and promotion-related fields if product payload includes them
- Images, video, size chart, and media handling
- Package dimensions/weight/logistics
- Brand, category, condition, GTIN/barcode/identifier fields
- Status/sync/error fields

### Lazada
Find current JSON/XML/SDK structures for:
- Product create/update/get
- SKU list and variation attributes
- Category attribute requirements
- Price, sale price, special price period
- Stock and warehouse if available
- Images, package dimensions, logistics, brand, warranty, identifiers
- Status/sync/error fields

### TikTok Shop
Find current JSON examples or schemas for:
- Product create/update/get
- SKU, seller SKU, sales attributes/variants
- Inventory/warehouse
- Price/currency
- Main images, SKU images, video, size chart
- Category attributes, certifications, package dimensions, logistics
- Brand, GTIN/product identifiers
- Status/sync/error fields

### AliExpress
Treat AliExpress as the most detailed baseline. Find current JSON examples or schemas for:
- Product create/update/get
- Product properties/specifications
- Sale properties and SKU properties
- SKU list / `aeopAeProductSKUs` or equivalent current structure
- SKU price, stock, discount/sale fields
- Product images, SKU images, detail images, video if available
- Package, logistics, shipping template, product unit
- Category, brand, identifiers, language/localized fields
- Status/sync/error fields

Do not use private accounts, credentials, or paid-only content.

## Analysis Questions
Answer these with source-backed evidence:

1. What does BC Account currently model well?
2. What is missing for marketplace readiness?
3. Which current fields are too technical or raw JSON for normal users?
4. Where does the current product/barcode/SKU model mix responsibilities incorrectly?
5. Is barcode/SKU being treated as stock owner anywhere? If yes, identify exact files/fields.
6. Is marketplace dimension stock separated from accounting stock? If not, identify gaps.
7. Is dimension-level selling price supported well enough? If not, propose a new model/table.
8. Are images/media modeled well enough for:
   - main image
   - gallery images
   - SKU/variant images
   - videos
   - size charts
   - detail/description images
   - external marketplace image IDs/URLs
9. Are product specifications modeled well enough for:
   - category attributes
   - sale attributes/options
   - SKU attributes
   - package/shipping attributes
   - compliance/certification attributes
   - raw marketplace attributes for lossless import
10. Can the model support:
    - clothing: color/size/material/gender/size chart
    - mobile phones: color/storage/RAM/IMEI/serial
    - SIM cards: package/network/phone number/activation state
    - computers: CPU/RAM/storage/serial/warranty
    - restaurant products: normal items, options, modifiers, product sets
    - serialized goods and lot/expiry goods
11. What should be combined to reduce user confusion?
12. What should be separate because it has a different lifecycle, permission, or scale?

## Canonical BC Account Model Design
Propose a canonical model using lower `snake_case`. It must separate normal user-facing data from importer/debug raw payloads.

Required model areas:

### 1. Product Core
Fields for product identity, code, names, type, category, brand, tax, unit, status, and high-level attributes.

### 2. Sellable Identity
Explain relationship among:
- product
- barcode
- SKU
- marketplace model/variant
- seller SKU
- GTIN/EAN/UPC/ISBN/serial/IMEI

Make clear why barcode/SKU is a sellable/scan/mapping identity, not the authoritative accounting stock owner.

### 3. Option And Dimension Model
Design:
- option groups such as color, size, storage, RAM, network, package
- option values
- SKU combinations
- dimension key strategy
- display labels in Thai/simple user language
- duplicate prevention

### 4. Media Model
Design `media_assets[]` or equivalent:
- media kind
- storage path
- external marketplace IDs
- original external URL
- local R2/S3 object path
- alt text
- variant/SKU association
- sort order
- source/import metadata

### 5. Specification Model
Design:
- `specification_groups[]`
- `attributes[]`
- `raw_attributes[]`
- required vs optional category attributes
- platform-specific preservation without forcing raw JSON into normal UI

### 6. Marketplace Mapping Model
Design:
- platform
- account/shop/channel ID
- external item ID
- external model/SKU ID
- category/brand mapping
- sync flags
- last sync status/error
- payload version/source

### 7. Stock Model
Design with this rule:
- accounting inventory balances remain product/warehouse/location based
- marketplace availability projection can be product-level and dimension-level
- reservation/available quantities should be explicit
- costing must not use marketplace dimensions

### 8. Price Model
Design:
- product price
- barcode/SKU price
- price level
- marketplace price
- dimension-level price
- currency
- effective date/time
- sale price/compare-at price
- promotion compatibility

If existing price arrays are insufficient, propose a separate `marketplacedimensionprices` or better model name.

### 9. Raw Payload Archive
Design a safe importer/debug-only model:
- source platform
- endpoint/action
- payload shape/version
- raw request/response body
- normalized record links
- import errors/warnings
- no secrets
- retention and privacy considerations

### 10. UI/UX Plan
Normal user UI must use plain business terms:
- Do not show raw JSON textareas in normal product forms.
- Use tabs/sections such as:
  - ข้อมูลสินค้า
  - ตัวเลือกสินค้า
  - รูปภาพ/วิดีโอ
  - รายละเอียดสินค้า
  - ราคา
  - สต๊อกพร้อมขาย
  - เชื่อม Marketplace
- Use table/list editors, chips, image previews, and guided forms instead of JSON editors.
- Keep importer/debug raw payloads behind admin/debug tools only.

## Required Output
Produce one Thai report with these sections:

```markdown
# Marketplace Foundation Audit

## RESULT
สรุปสิ่งที่ตรวจและภาพรวมความพร้อม

## SOURCE EVIDENCE
รายการไฟล์ใน repo ที่อ่าน พร้อมบรรทัดสำคัญ

## MARKETPLACE JSON EVIDENCE
ตาราง Shopee/Lazada/TikTok/AliExpress: URL, source type, field families found, confidence

## CURRENT SYSTEM MAP
แผนผัง product/barcode/SKU/stock/price/media/spec ที่ระบบมีอยู่ตอนนี้

## GAP AND WEAKNESS
จุดอ่อนเรียงตามความเสี่ยงและผลกระทบ

## CANONICAL BC STANDARD
โมเดลกลางที่เสนอ แยก product, barcode/SKU, options, media, specs, marketplace mapping, stock, price, raw payload

## JSON EXAMPLES
ตัวอย่าง JSON กลางของ BC Account ที่ normalize จาก marketplace แล้ว

## MIGRATION PLAN
แผนปรับ model/API/DB/frontend ทีละขั้น พร้อม rollback idea

## TEST PLAN
คำสั่งและกรณีทดสอบจริง

## OPEN QUESTIONS
เรื่องที่ยังต้องให้ลุงจืดตัดสินใจ
```

## Required JSON Examples
Include compact but realistic JSON examples for:

1. Normal product with no variants.
2. Clothing product with color/size variants.
3. Mobile phone with color/storage and serial/IMEI handling.
4. SIM product with network/package/activation-related attributes.
5. Computer product with CPU/RAM/storage and serial/warranty.
6. Restaurant product with options/modifiers and product set compatibility.

Rules for JSON examples:
- Use lower `snake_case`.
- Include `holdingcode`.
- Include immutable `guidfixed` where it is a BC record.
- Include user-facing codes such as `businesscode`, `itemcode`, `barcode`, `sku_code`.
- Do not include secrets or real customer data.
- Keep examples readable; do not dump huge raw payloads.

## Implementation Guidance If Jead Later Says "แก้เลย"
If Jead asks to implement after the audit:
1. Start with a concise plan.
2. Patch the smallest safe scope.
3. Update central rules/skills if business model changes.
4. Add migration only when needed.
5. Use MongoDB as operational source of truth.
6. Use PostgreSQL/ClickHouse only for rebuildable projections/reporting.
7. Run targeted verification:
   - `cd frontend; npm run typecheck`
   - targeted backend package tests only
   - browser check if UI changed
8. Run secret scan before commit/push.
9. Do not create or restore `manual/`.

## Final Reminder
Do not guess. If marketplace documentation is incomplete or inconsistent, preserve the uncertainty and design a flexible model that can store unknown raw attributes safely while keeping normal UI simple.
