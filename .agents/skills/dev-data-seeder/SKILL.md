---
name: dev-data-seeder
description: Use when creating or verifying real DEV seed data for BC Ai Account screens, especially tenant-scoped product, product set, barcode, menu, unit, or category data that must appear in the active UI.
---

# DEV Data Seeder

## Core Rules
- Seed only into DEV and only through real application/backend APIs unless Jead explicitly approves another path.
- Resolve the active tenant first. For UI issues, match the browser-selected company/shop to the backend auth token `holdingcode`; do not seed into a guessed tenant.
- Do not store tokens, passwords, MongoDB URIs, R2 keys, or server credentials in scripts, docs, commits, or chat.
- Prefix/generated codes must be deterministic and easy to audit. Use business-readable names, but keep technical codes stable enough to re-run safely.
- Verify after seeding through the same API path the screen uses, and separately verify specialized flows such as product set classification or barcode lookup.
- Product Set screens must query `/product` with `itemtype=2&materialtype=3`; do not rely on frontend-only filtering of the first page.

## Product / Barcode / Product Set Seeding
- Product Set records must use `itemtype = 2` and `materialtype = 3`.
- Normal restaurant menu products should use `itemtype = 0` and `materialtype = 0` unless the task requires another type.
- For PowerShell JSON payloads, force arrays to stay arrays (`@(...)`) for `names`, `prices`, `barcodes`, `refbarcodes`, and `bom`.
- Prefer `curl.exe` for local backend API calls; `Invoke-RestMethod` has timed out in this workspace.
- Run bundled PowerShell scripts with `pwsh` (PowerShell 7+) so UTF-8 Thai seed names are parsed correctly.
- If using the bundled script, pass the current `holdingcode` and username; the script reads the matching runtime Redis auth token without printing it.

## Product Category Group Seeding
- `productCategories.groupnumber` is a usage/device/channel group, not a flat product-menu category. Examples: group 1 ordering/tablet, group 2 cashier/POS, group 3 kitchen/KDS, group 4 delivery.
- Each seeded group must be a complete category tree. Create root usage/category nodes first through `POST /product/category`, capture the returned GUID, then create child/subchild nodes with `parentguid` and `parentguidall`.
- Attach `codelist` only to sellable leaf categories, or to an explicitly sellable node. Do not seed one flat root per menu type across groups.
- Verify `/product/category/list?group-number=<n>` returns multiple hierarchical nodes for a populated group, and verify at least one leaf category contains Product-master-only `codelist` entries shaped `{code,xorder,names}`. Every `code` must resolve to an active Mongo `products` record in the same Holding; never seed Barcode/unit fields as category membership.

## Bundled Script
- `scripts/seed-companies-branches-dev.ps1`
  - Creates real DEV organization companies through `POST /organization/company`.
  - Relies on backend auto-creation of head-office branch `00000`, then creates 1-3 additional branches per company with `POST /organization/branch`.
  - Uses the selected runtime Redis auth token for the provided `holdingcode` and username without printing the token.
- `scripts/seed-restaurant-menu-dev.ps1`
  - Creates real DEV restaurant menu products, barcodes, and Product Set records for the selected `holdingcode`.
  - Default count is 200 normal products plus 5 product sets.
  - Uses `POST /product`, `POST /product/barcode/bulk`, then item-level `PUT /product/barcode/:guid` for Product Set BOM, and verifies `/product`, `/product/barcode`, and `/product/barcode/pk/:barcode`.

Example:
```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File .agents/skills/dev-data-seeder/scripts/seed-restaurant-menu-dev.ps1 `
  -HoldingCode "selected-holdingcode" `
  -Username "user@example.com" `
  -BaseUrl "http://localhost:8888" `
  -Count 200
```
