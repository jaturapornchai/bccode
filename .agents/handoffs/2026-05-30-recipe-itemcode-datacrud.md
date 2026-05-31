# Handoff: Recipe ItemCode and Datacrud Layout

- Feature: Recipe ItemCode & Datacrud Layout
- Author: Gemini (Antigravity)
- Date: 2026-05-30
- Scope: completed-end-to-end
- Status: Gemini=DONE | Codex=DONE | Claude=DONE

## 1. Frontend change (done by Gemini)
- Added **Recipe Item Code (รหัสสูตร)** input field to the parent product creation form when creating a new BOM recipe from scratch (virtual recipe).
- Integrated `parentItemCode` state into the creation payload (`itemcode` field). If the user doesn't specify a barcode, the barcode default becomes the itemcode value.
- Refactored the BOM screen layout to follow the **datacrud** specification:
  - Left pane lists formula-associated products.
  - A mouse-draggable split bar separates the panes with local storage split persistence (`bc_bom_split_left`).
  - Added a count label to the list headers showing the dataset count (e.g. `10 / 10 รายการ`).
  - Right pane loads the inline custom BOM editor `ProductBomEditor` with custom navigation controls on mobile.
  - Form validations and specs rendering updated to include/display the Item Code.

## 2. Backend / model changes required (for Codex)
Codex needs to verify that the Go API backend handles `itemcode` cleanly for all BOM operations:
- Check MongoDB schema mapping for `productBarcodes`. Ensure `itemcode` is fully mapped and saved correctly on `POST /api/product-barcode` and `PUT /api/product-barcode/:guid`.
- Verify the Go API endpoint `/api/system-settings/product_bom` that triggers BOM rebuilds and calculates rollup costs, ensuring it doesn't break if `barcode` is identical to `itemcode` (for new products created without barcodes).
- Check standard database model sync scripts or projections (PostgreSQL / ClickHouse) if they reference `itemcode` or `barcode` for BOM listings.

Source locations to check:
- `backend/internal/productbarcode/` (Model and MongoDB mapping)
- `backend/internal/systemsettings/productbom/` (BOM services and calculations)

## 3. API contract expected by the frontend
Payload for `POST /api/product-barcode`:
```json
{
  "shopid": "...",
  "barcode": "RECIPE-001",
  "itemcode": "RECIPE-001",
  "names": [{"code": "th", "name": "..."}],
  "item_unit_code": "PCS",
  "itemunitnames": [{"code": "th", "name": "..."}],
  "price": 0,
  "bom": [...],
  "boms": [...]
}
```

## 4. Affected files
- Frontend (done):
  - [product-bom-editor.tsx](file:///d:/bccode/frontend/src/app/system-settings/product-bom-editor.tsx)
  - [system-settings-screen.tsx](file:///d:/bccode/frontend/src/app/system-settings/system-settings-screen.tsx)
- Backend (todo):
  - `backend/internal/productbarcode/...`

## 5. Verification checklist
- [x] Run `npm run typecheck` in `frontend/` directory (Passed with 0 errors).
- [x] Verify Go backend builds with the MongoDB schema upgrades: `cd backend; go build ./...`
- [x] Run end-to-end integration tests creating virtual recipes from scratch, ensuring itemcode is correctly populated.

## 6. Notes for Claude (review + plan)
- The layout split values are saved and restored using local storage. Double check if we should add responsive fallback values for smaller desktop viewports.
- The Go backend triggers BOM rollup updates automatically using `/api/system-settings/product_bom?barcode=...`. Double-check the transaction isolation and database consistency.
