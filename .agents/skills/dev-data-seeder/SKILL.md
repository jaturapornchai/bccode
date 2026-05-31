---
name: dev-data-seeder
description: Use when creating or verifying real DEV seed data for BC Account screens, especially tenant-scoped product, product set, barcode, menu, unit, or category data that must appear in the active UI.
---

# DEV Data Seeder

## Core Rules
- Seed only into DEV and only through real application/backend APIs unless Jead explicitly approves another path.
- Resolve the active tenant first. For UI issues, match the browser-selected company/shop to the backend auth token `shopid`; do not seed into a guessed tenant.
- Do not store tokens, passwords, MongoDB URIs, R2 keys, or server credentials in scripts, docs, commits, or chat.
- Prefix/generated codes must be deterministic and easy to audit. Use business-readable names, but keep technical codes stable enough to re-run safely.
- Verify after seeding through the same API path the screen uses, and separately verify specialized flows such as product set classification or barcode lookup.
- Product Set screens must query `/product` with `item_type=2&materialtype=3`; do not rely on frontend-only filtering of the first page.

## Product / Barcode / Product Set Seeding
- Product Set records must use `item_type = 2` and `materialtype = 3`.
- Normal restaurant menu products should use `item_type = 0` and `materialtype = 0` unless the task requires another type.
- For PowerShell JSON payloads, force arrays to stay arrays (`@(...)`) for `names`, `prices`, `barcodes`, `refbarcodes`, and `bom`.
- Prefer `curl.exe` for local backend API calls; `Invoke-RestMethod` has timed out in this workspace.
- Run bundled PowerShell scripts with `pwsh` (PowerShell 7+) so UTF-8 Thai seed names are parsed correctly.
- If using the bundled script, pass the current `shopid` and username; the script reads the matching runtime Redis auth token without printing it.

## Bundled Script
- `scripts/seed-restaurant-menu-dev.ps1`
  - Creates real DEV restaurant menu products, barcodes, and Product Set records for the selected `shopid`.
  - Default count is 200 normal products plus 5 product sets.
  - Uses `POST /product`, `POST /product/barcode/bulk`, then item-level `PUT /product/barcode/:guid` for Product Set BOM, and verifies `/product`, `/product/barcode`, and `/product/barcode/pk/:barcode`.

Example:
```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File .agents/skills/dev-data-seeder/scripts/seed-restaurant-menu-dev.ps1 `
  -ShopId "selected-shopid" `
  -Username "user@example.com" `
  -BaseUrl "http://localhost:8888" `
  -Count 200
```
