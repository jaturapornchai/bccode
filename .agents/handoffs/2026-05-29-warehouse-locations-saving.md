# Handoff: Save nested Warehouse Locations & Shelves in PUT API

- Feature: Warehouse locations & shelves saving support
- Author: Gemini
- Date: 2026-05-29
- Scope: frontend-done / backend-todo
- Status: Gemini=DONE | Codex=TODO | Claude=TODO

## 1. Frontend change (done by Gemini)
The frontend `WarehouseTreeView` component (in `frontend/src/app/system-settings/warehouse-tree-view.tsx`) manages the entire warehouse structure (Warehouse -> Location/Zone -> Shelf) as a single tree-structured state.
When creating, editing, reordering, or deleting locations/shelves, it sends a single `PUT` request to `/api/system-settings/product_warehouse_screen/{warehouse_guid}` containing the full nested payload:
```json
{
  "guid_fixed": "warehouse-guid",
  "code": "00000",
  "names": [...],
  "location": [
    {
      "code": "ZONE-01",
      "names": [...],
      "shelf": [
        {
          "code": "SHELF-01",
          "name": "Shelf 1"
        }
      ]
    }
  ],
  "company_guids": [...]
}
```
Currently, the UI does not display nested child locations/shelves because the backend fails to persist the nested `"location"` (Zones) and `"shelf"` (Shelves) structure.

## 2. Backend / model changes required (for Codex)
The backend Go API needs to parse and persist the nested `location` (Zones) and `shelf` (Shelves) fields during the `PUT /warehouse/:id` operation.
Source-link: [warehouse_http.go:L185-L250](file:///d:/bccode/backend/internal/warehouse/warehouse_http.go#L185-L250)

Codex must modify the `UpdateWarehouse` handler:
- [ ] Parse incoming nested `Zones` (which is tag-named `location`) and `Shelves` (tag-named `shelf`) from `UpdateWarehouseRequest`.
- [ ] Inside the GORM transaction, delete old nested relations:
  - Delete all shelves (`warehouse_shelves`) that belong to zones of this warehouse.
  - Delete all zones (`warehouse_zones`) that belong to this warehouse.
- [ ] Insert the new zones and shelves from the request:
  - Set `WarehouseGuid` to the current warehouse ID.
  - Generates new GUID for Zone/Shelf if empty, and link `ZoneGuid` in shelf records.
  - Insert them into `warehouse_zones` and `warehouse_shelves` tables.

## 3. API contract expected by the frontend
The request payload is structured as `WarehousePg` with pre-loaded `Zones` (as `location` array) and `Shelves` (as `shelf` array inside each location):
```typescript
interface WarehouseRecord {
  guid_fixed?: string;
  code?: string;
  names?: LocalizedNames;
  location?: WarehouseLocation[]; // mapped to Zones in Go struct
  latitude?: number;
  longitude?: number;
  company_guids?: string[];
}
```

## 4. Affected files
- Frontend (done): [warehouse-tree-view.tsx](file:///d:/bccode/frontend/src/app/system-settings/warehouse-tree-view.tsx)
- Backend (todo): [warehouse_http.go](file:///d:/bccode/backend/internal/warehouse/warehouse_http.go)

## 5. Verification checklist
- [ ] `cd backend; go build ./...`
- [ ] `cd frontend; npm run typecheck`
- [ ] Open Warehouse Settings screen, add a location, add a shelf, click Save, refresh the page, and ensure the child locations/shelves are persisted and visible in the tree view list.

## 6. Notes for Claude (review + plan)
Go API `UpdateWarehouse` was only updating the root warehouse attributes and ignoring the nested zones/shelves slice, leading to empty locations on save. Re-inserting them inside the transaction resolves the sync mismatch.
