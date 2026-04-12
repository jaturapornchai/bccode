---
name: master-data
description: >
  Manage master data via MCP tools: units, barcodes, product groups, categories, creditors
  (suppliers), and debtors (customers). Use when the user wants to create/update/delete
  master records, search products, or inspect schemas.
user-invocable: true
---

# Master Data Management

## When to Use
- User wants to add a new unit of measure (e.g., Dozen, Box, Pallet)
- User wants to create or update a product barcode with a new price/unit
- User needs to add a new supplier (creditor) or customer (debtor) to the system
- User asks to batch-create multiple records at once (e.g., import list of units)
- User wants to inspect what fields are required before creating a record

## Anti-Patterns
- Do NOT create barcodes without checking the schema first — missing required fields cause silent failures
- Do NOT delete a unit that is referenced by existing barcodes or transactions
- Do NOT create duplicate creditor/debtor codes — the system does not auto-deduplicate
- Do NOT skip `get_ref_barcodes` when setting up multi-unit products — reference chain must be correct
- Do NOT use `search_products` to modify data — it is read-only; use dedicated MCP tools for writes

## Usage
`/master-data list <type>` -- List records
`/master-data search <keyword>` -- Search products
`/master-data create <type> <data>` -- Create record
`/master-data update <type> <data>` -- Update record
`/master-data delete <type> <guid>` -- Delete record

## Supported Types

| Type | MCP Tools | Description |
|------|-----------|-------------|
| `units` | list/create/update/delete_unit(s) | Units of measure (piece, box, dozen) |
| `barcodes` | list/create/update/delete_barcode(s) | Product barcodes |
| `product-groups` | list/create/update/delete_product_group(s) | Product groups |
| `product-categories` | list/create/update/delete_product_category(ies) | Product categories |
| `creditors` | list/create/update/delete_creditor(s) | Creditors (suppliers) |
| `debtors` | list/create/update/delete_debtor(s) | Debtors (customers) |
| `products` | search_products | Search products (read-only) |

## Workflows

### List
```
/master-data list units
-> Call MCP tool: list_units
-> Display table: code, name, description
```

### Search
```
/master-data search MAKITA
-> Call MCP tool: search_products
-> Parameters: { keyword: "MAKITA" }
-> Display: code, name, price, stock
```

### Create
```
/master-data create unit { "code": "DOZ", "name1": "Dozen", "name2": "Dozen" }
-> Call MCP tool: create_unit
-> Confirm result
```

### Batch Create
```
/master-data create units [
  { "code": "DOZ", "name1": "Dozen" },
  { "code": "BOX", "name1": "Box" }
]
-> Call MCP tool: create_units (batch)
```

### Check Schema Before Create
Inspect schema before create/update:
```
Tool: get_barcode_schema / get_unit_schema / get_creditor_schema / ...
-> Display required fields, data types, validation rules
```

## MCP Tools Reference

### Units
| Tool | Purpose |
|------|---------|
| `list_units` | List units |
| `create_unit` | Create single unit |
| `create_units` | Create multiple units |
| `update_unit` | Update unit |
| `delete_unit` | Delete single unit |
| `delete_units` | Delete multiple units |
| `get_unit_schema` | View schema |

### Barcodes
| Tool | Purpose |
|------|---------|
| `list_barcodes` | List barcodes |
| `create_barcode` | Create barcode |
| `create_barcodes` | Create multiple barcodes |
| `update_barcode` | Update barcode |
| `delete_barcode` | Delete barcode |
| `delete_barcodes` | Delete multiple barcodes |
| `get_barcode_schema` | View schema |
| `get_ref_barcodes` | View reference barcodes (reference chain) |
| `set_ref_barcode` | Set reference barcode |
| `create_multi_unit_barcode` | Create multi-unit product |

### Product Groups / Categories / Creditors / Debtors
Each type follows the same pattern: `list_`, `create_`, `create_` (batch), `update_`, `delete_`, `delete_` (batch), `get_schema`.

Replace suffix: `product_group(s)`, `product_category(ies)`, `creditor(s)`, `debtor(s)`.

### Products (read-only)
| Tool | Purpose |
|------|---------|
| `search_products` | Search products + stock balance |

## Examples
```
/master-data list units           -> All units
/master-data list barcodes        -> All barcodes
/master-data list creditors       -> All creditors
/master-data search MAKITA        -> Search products by keyword
```

## Notes
- Create/Update/Delete require an API key with write permissions
- `search_products` supports Thai language
- Batch operations (create_units, delete_barcodes) accept arrays

## Related Skills
- `/product` — full product management, BOM, options, AI classification
- `/stock-inventory` — warehouse and stock balance management
- `/import-export` — bulk import products and stock from Excel files
- `/coupon-promotion` — coupon conditions linked to specific products or customers
