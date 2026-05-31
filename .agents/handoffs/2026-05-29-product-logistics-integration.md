# Handoff: Product and Barcode Logistics Dimensions

- Feature: product-logistics-dimensions
- Author: Gemini
- Date: 2026-05-29
- Scope: frontend-done / backend-todo
- Status: Gemini=DONE | Codex=TODO | Claude=TODO

## 1. Frontend change (done by Gemini)
Exposed logistics and dimensions editing fields in the product editing screen and individual barcode variant dialogs:
- Updated [product-screen.tsx](file:///d:/bccode/frontend/src/app/menu/product-screen.tsx): added a `"logistics"` editing tab containing `package_weight`, `package_width`, `package_length`, and `package_height` input fields, a real-time volumetric weight calculator `(W * L * H) / 5000`, and a Fragile warning badge configuration. Also added a summary card in the View mode grid.
- Updated [barcode-form.tsx](file:///d:/bccode/frontend/src/components/product-barcode/barcode-form.tsx): added a `"logistics"` tab inside the barcode variant dialog to configure weight, dimensions, and Fragile stickers at the specific variant level.

## 2. Backend / model changes required (for Codex)
Please verify that the backend API properly stores and updates these fields in PostgreSQL/MongoDB (depending on the authoritative store for master products and barcode units):
- [ ] Product fields: `package_weight`, `package_width`, `package_length`, `package_height`, `isalert`, `alertdescription`.
- [ ] Barcode variant fields: `package_weight`, `package_width`, `package_length`, `package_height`, `isalert`, `alertdescription`.
- [ ] Verify that these fields are correctly sent and received via `POST /api/product` / `PUT /api/product/:guid` and `POST /api/product-barcode` / `PUT /api/product-barcode/:guid`.

## 3. API contract expected by the frontend
The frontend maps the values directly to these keys in JSON:
```typescript
{
  package_weight: number;
  package_width: number;
  package_length: number;
  package_height: number;
  isalert: boolean;
  alertdescription: string;
}
```

## 4. Affected files
- Frontend (done): [product-screen.tsx](file:///d:/bccode/frontend/src/app/menu/product-screen.tsx), [barcode-form.tsx](file:///d:/bccode/frontend/src/components/product-barcode/barcode-form.tsx)
- Backend (todo): `backend/internal/product/...`
