# Handoff: Marketplace Mappings Split (Shopee / Lazada / TikTok)

- Feature: Separate mappings menu and screens for Shopee, Lazada, and TikTok
- Author: Gemini
- Date: 2026-05-29
- Scope: frontend-done / backend-todo
- Status: Gemini=DONE | Codex=TODO | Claude=TODO

## 1. Frontend change (done by Gemini)
Split the single Marketplace configuration tab out of the product edit form, and created three dedicated master menu routes to handle Shopee, Lazada, and TikTok shop mappings independently.
- Added three routes to `frontend/src/lib/menu-data.ts`:
  - `Shopee Mappings` -> `/marketplace/shopee`
  - `Lazada Mappings` -> `/marketplace/lazada`
  - `TikTok Mappings` -> `/marketplace/tiktok`
- Implemented `MarketplaceMappingsScreen` in `frontend/src/app/menu/marketplace-screen.tsx` which accepts a `platform` prop and displays:
  - **Connections**: Manage active shop ID bindings for the specific platform.
  - **Mappings**: View all system product barcodes, customize mapping details (Seller SKU, Market Variant/Model ID), and perform Client-side Bulk Imports by dropping a CSV file or copying-pasting Excel columns directly.
  - **Logs**: Simulated platform sync logs to be replaced with real API database logs.
- Integrated the route matching in `WorkTabPanel` of `frontend/src/app/menu/main-menu-screen.tsx`.

## 2. Backend / model changes required (for Codex)
To fully support the separate marketplace mapping interface:
- **API Mappings Optimization (TODO)**:
  - Add a dedicated GET endpoint `/goapi/api/product/marketplace/mappings?platform={platform}` to fetch only mapping-specific structures rather than loading the heavy product barcode lists.
- **Bulk Update Endpoint (TODO)**:
  - Create a POST endpoint `/goapi/api/product/marketplace/bulk-save` to save multiple product mappings in a single request. Currently, the frontend loops and calls the single product update API (`/api/product-barcode/:guid`), which causes overhead when updating many mappings at once.
- **Real Event Logs (TODO)**:
  - Add an endpoint `/goapi/api/product/marketplace/logs?platform={platform}` to retrieve actual stock/price sync logs from ClickHouse/PostgreSQL instead of using the frontend mock list.
- **Database Model Update (If Any)**:
  - Verify that `marketplace_products` and `refbarcodes.marketplace_sku_mappings` schemas under backend Go structures mirror the parsed payload correctly.

## 3. API contract expected by the frontend
Currently, mappings are submitted via `updateBarcode` API. When Codex designs the new bulk-save endpoint, the expected payload structure will be:
```json
{
  "platform": "shopee",
  "shop_id": "shop_01",
  "mappings": [
    {
      "barcode": "8850123456789",
      "seller_sku": "sku-red-01",
      "market_item_id": "123456789",
      "market_model_id": "98765432"
    }
  ]
}
```

## 4. Affected files
- Frontend (done):
  - `frontend/src/lib/menu-data.ts` (Added 3 submenu entries)
  - `frontend/src/app/menu/main-menu-screen.tsx` (Mapped routes to MarketplaceMappingsScreen)
  - `frontend/src/app/menu/marketplace-screen.tsx` (Created mapping workspace with bulk import processor)
- Backend (todo):
  - Add mapping API handlers, bulk save logic, and logger hooks.

## 5. Verification checklist
- [x] Run `npm run typecheck` inside `frontend/` -> Completed successfully with 0 errors.

## 6. Notes for Claude (review + plan)
- Look into adding proper loading spinners or notifications when calling backend API saves.
- Verify security checks when adding OAuth integrations for Shopee, Lazada, and TikTok in later stages.
