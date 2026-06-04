# Handoff: New Product Set/Bundle features

- Feature: product-set-ui-upgrade
- Author: Gemini
- Date: 2026-05-29
- Scope: frontend-done / backend-todo
- Status: Gemini=DONE | Codex=TODO | Claude=TODO

## 1. Frontend change (done by Gemini)
Overwrote the Product Set (Bundle) screen in `frontend/src/app/menu/product-set-screen.tsx` with a state-of-the-art UI:
- Added a 3-column layout on desktop: Left Sidebar (sets list & stats), Middle Editor (General Info, Card-based Components, Logistics), Right Preview/Simulator (visual storefront preview with live selections, margin calculations, and stock trace diagram).
- Added a custom `BarcodePickerModal` using `listBarcodes` to search/fetch live products, prices, and stock balances for options selection.
- Added dynamic background loading of barcode details for existing options on selected product changes.

## 2. Backend / model changes required (for Codex)
Please adapt the backend/models to match the newly added frontend UI parameters:
- [ ] Save the components options/choices structure in MongoDB `Product` model. The frontend payload sends the `options` array where each option has:
  - `guid` (string)
  - `names` (NameX[])
  - `choicetype` (0 | 1)
  - `minselect` (number)
  - `maxselect` (number)
  - `choices` (ProductChoice[]): Each choice has `guid`, `names`, `refbarcode`, `refbarcodenames`, `isstock` (boolean), `isdefault` (boolean), `qty` (number), and `price` (string containing numeric delta price).
- [ ] Ensure `isusesubbarcodes` (boolean, component-level inventory vs bundle-level inventory) is properly saved and respected when a sale transaction occurs (inventory deduct trace logic).
- [ ] Ensure `condition` (boolean, fixed set price vs dynamic price) is saved and respected.
- [ ] Ensure package dimensions (`package_weight`, `package_width`, `package_length`, `package_height`) are saved in MongoDB `Product` model.
- [ ] Implement backend automatic rebuild and deploy to local Docker Desktop:
  `cd backend; docker-compose up -d --no-deps --build mainapi`

## 3. API contract expected by the frontend
The frontend expects to save changes using:
- `POST /api/product` (create)
- `PUT /api/product/:guid` (update)

The request payload shape is:
```typescript
{
  guidfixed: string;
  holding_code: string;
  code: string;
  names: NameX[];
  group_code: string;
  group_names: NameX[];
  item_type: 2; // Always SET
  condition: boolean;
  isusesubbarcodes: boolean;
  options: ProductOption[];
  package_weight: number;
  package_width: number;
  package_length: number;
  package_height: number;
  description?: string;
  // Other standard product fields
}
```

## 4. Affected files
- Frontend (done): [product-set-screen.tsx](file:///d:/bccode/frontend/src/app/menu/product-set-screen.tsx)
- Backend (todo): `backend/internal/product/...`

## 5. Verification checklist
- [ ] `cd backend; go build ./...`
- [ ] `cd frontend; npm run typecheck`
- [ ] Live API save of a product set containing multiple options and choices.

## 6. Notes for Claude (review + plan)
The frontend relies on the Next.js API router `/api/product` and `/api/product/:guid` which proxy to the Go backend. Ensure the Go models deserialize and save the components structure correctly to MongoDB.
