# Handoff: Product Marketplace 3-Tab (Shopee / Lazada / TikTok)

- Feature: Marketplace mapping tabs on the product/barcode form
- Author: Claude (full-stack this round — rule was set after the work)
- Date: 2026-05-29
- Scope: frontend-done / model-done / backend-sync-todo
- Status: Gemini=N/A | Codex=TODO (Phase 2 sync) | Claude=DONE (model + UI)

## 1. Frontend change (done)
Added 3 marketplace tabs (Shopee / Lazada / TikTok) to the barcode edit form, each a separate tab next to basic/pricing/media. Per platform: toggle to link, then fields for listing id/SKU/category/brand/price/stock/status/sync. Unified shape — all 3 tabs share one `MarketplaceProductMap`, filtered by `platform`.
- `frontend/src/components/product-barcode/barcode-form.tsx` — `TabMarketplace`, `MARKETPLACE_TABS`, tab wiring
- `frontend/src/lib/product-barcode/types.ts` — `MarketplaceProductMap`/`MarketplaceSKUMap`, `MARKETPLACE_PLATFORMS/STATUS`, `emptyMarketplaceProductMap`, `marketplace_products` on `ProductBarcode`
- `frontend/src/lib/product-barcode/language.ts` — 33 TH/EN label keys (`tabShopee`, `mk*`)

## 2. Backend / model changes
**DONE (this round):**
- `backend/internal/product/product/models/product.go:295` — `MarketplaceProductMap` (4→26 fields), `MarketplaceSKUMap` (7→17 fields), unified, backward-compatible.
- `backend/internal/product/productbarcode/models/product_barcode.go:379` — identical structs (duplicated by package).
- `backend/internal/product/productbarcode/models/marketplace_map_test.go` — JSON round-trip + contract guard (2 tests PASS).
- No migration needed: MongoDB is schemaless; new fields default to zero on old docs. CRUD save/load already carries `marketplace_products` via the existing barcode bson path.

**TODO — Phase 2 (Codex), when Jead asks for real sync:**
- [ ] OAuth + API client per platform (Shopee Open API v2, Lazada Open Platform, TikTok Shop Partner API). Credentials from env/secret only (no hardcode, no fallback — core-rules No Fallback).
- [ ] Inbound sync service: pull listing → fill `platform_price`, `platform_stock`, `status`, `last_sync_at`/`last_sync_error`.
- [ ] Outbound sync: push `custom_price`/stock when `sync_price`/`sync_stock` enabled.
- [ ] Unit normalization on push: Shopee weight=KG/dim=cm, Lazada weight=g/dim=mm — store one base unit, convert at the boundary.
- [ ] Validate `platform` ∈ {shopee,lazada,tiktok} on save; tenant-scope all marketplace queries by `holding_code`.
- [ ] MCP tool for marketplace status (optional) so the AI agent can query/sync.

## 3. API contract expected by the frontend
No new endpoint this round — `marketplace_products` rides on the existing product-barcode create/update payload (`ProductBarcodeBase.marketplace_products`). Phase 2 sync endpoints are open for Codex to design under `/goapi/...` and route through MainAPI.

## 4. Affected files
- Frontend (done): the 3 files in §1.
- Backend (done): the 3 files in §2 "DONE".
- Backend (todo): new `internal/...marketplace/...` package for Phase 2 (Codex to place).

## 5. Verification checklist
- [x] `go build`/`go vet` on both model packages — OK
- [x] `go test ./internal/product/productbarcode/models/` — 2 PASS
- [x] `cd frontend; npm run typecheck` — exit 0
- [ ] Full backend Docker build — blocked by pre-existing host kafka/CGO + WIP S3→R2 in working tree; run `cd backend; docker-compose up -d --no-deps --build mainapi` + verify `/healthz` once tree is clean
- [ ] Light/dark + mobile check of the 3 tabs (reused themed components; not yet opened in a browser)

## 6. Notes for Claude / Codex
- Status enum uses radio (6 options) to avoid native-select dark-mode issues; swap to a themed combobox if a dropdown is preferred.
- `languages.tsv` not yet synced with the new `mk*` keys (local pack used, per `language.ts` interim pattern) — sync when convenient.
- Phase 2 is a separate, larger effort (per-platform OAuth + workers). Keep `v1` contract intact when adding sync endpoints.
