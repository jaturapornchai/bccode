---
name: product
description: >
  End-to-end product management: product data, barcodes, units, BOM, options, AI classification,
  and product search with Thai NLP. Use when the user mentions product creation, barcode
  management, BOM, price setup, product categories, or AI classification.
user-invocable: true
---

# Product Management System

## Overview
End-to-end product management: product data, barcodes, units, categories, BOM, options, images

## When to Use
- User wants to create/edit a product, set price, or add barcode
- User asks about BOM setup (e.g., "Rice Curry Set" composed of sub-items)
- User needs to configure product options (Color/Size/Model variants)
- User wants to set up AI classification levels (Brand, Category, Class... 10 levels)
- User asks about product search, Thai NLP search, or search aliases

## Anti-Patterns
- Do NOT create a barcode without linking it to a product (`itemcode` is required)
- Do NOT delete a unit that is already used by existing barcodes — check references first
- Do NOT set duplicate barcodes across different products — system will reject
- Do NOT modify `smlaiproduct` (AI classification) levels without syncing to MongoDB + Kafka
- Do NOT skip Kafka sync after product changes — PG + ClickHouse will go out of sync

## System Structure

### Product (Main)
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/product/product_add.dart` | Create/edit product |
| Screen | `screens/product/product_screen.dart` | Product list |
| Screen | `screens/product/product_image.dart` | Manage product images |
| BLoC | `bloc/product/` | Product state management |
| Backend | `backend/internal/product/product/` | Product CRUD |

### Barcode
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/config/product_barcode_screen.dart` | Manage barcodes |
| BLoC | `bloc/product_barcode/` | State management |
| Backend | `backend/internal/product/productbarcode/` | CRUD + price history |
| Consumer | `backend/cmd/productbarcodeconsumer/` | Process barcode events |

### Unit
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/config/product_unit_screen.dart` | Manage units |
| BLoC | `bloc/unit/` | State management |
| Backend | `backend/internal/product/unit/` | Unit CRUD |

### Category & Group
| Level | Screen | Backend |
|-------|--------|---------|
| Category | `product_category_screen.dart` | `product/productcategory/` |
| Group | `product_group_screen.dart` | `product/productgroup/` |
| Type | `product_type_screen.dart` | `product/producttype/` |
| Dimension | `product_dimension_screen.dart` | `backend/internal/dimension/` |
| Color | `color_screen.dart` | `product/color/` |

### AI Product Classification (smlaiproduct — 10 Levels)
| Level | Screen | Backend |
|-------|--------|---------|
| Brand | `master_brand_screen.dart` | `smlaiproduct/brandproduct/` |
| Category | `master_category_screen.dart` | `smlaiproduct/categoryproduct/` |
| Class | `master_class_screen.dart` | `smlaiproduct/classproduct/` |
| Design | `master_design_screen.dart` | `smlaiproduct/designproduct/` |
| Grade | `master_grade_screen.dart` | `smlaiproduct/gradeproduct/` |
| Group | `master_group_screen.dart` | `smlaiproduct/groupproduct/` |
| Sub-Group 1 | `master_group_sub1_screen.dart` | `smlaiproduct/groupsuboneproduct/` |
| Sub-Group 2 | `master_group_sub2_screen.dart` | `smlaiproduct/groupsubtwoproduct/` |
| Model | `master_model_screen.dart` | `smlaiproduct/modelproduct/` |
| Pattern | `master_pattern_screen.dart` | `smlaiproduct/patternproduct/` |

### BOM (Bill of Materials)
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/config/product_barcode_bom_screen.dart` | Manage BOM |
| Widget | `screens/config/product_bom_widget.dart` | Display BOM |
| Backend | `backend/internal/product/bom/` | BOM CRUD |

### Product Options (Color/Size/Model)
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/option/option_list.dart` | Option list |
| Screen | `screens/option/option_add.dart` | Create option |
| Backend Option | `backend/internal/product/option/` | Option CRUD |
| Backend Pattern | `backend/internal/product/optionpattern/` | Option pattern |

## API Endpoints (GoAPI)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/product/search` | Search products |
| POST | `/api/product/barcode` | Search by barcode |
| POST | `/api/product/barcode/list` | Barcode list |
| POST | `/api/product/search/unified` | Unified search (Thai NLP) |
| POST | `/api/process/product-balance` | Calculate product balance |
| GET | `/api/product/cache/stats` | Product cache statistics |
| POST | `/api/product/cache/clear` | Clear cache |
| GET | `/api/search/aliases` | Search aliases |
| POST | `/xlsx/product/start` | Import from Excel |

## Product Search
- **Thai NLP**: Thai word tokenization + transliteration
- **Unified Search**: Search across name + barcode + code in a single query
- **Search Aliases**: Set alternative names, e.g. "Coke" -> "Coca-Cola"
- **Product Cache**: Cache frequently accessed product data in memory

## Frontend (Flutter)

### BLoCs (12 total)
| BLoC | Purpose |
|------|---------|
| `product_bloc` | Main product CRUD |
| `productmaster_bloc` | Product master data |
| `product_list_bloc` | Product list |
| `product_barcode_bloc` | Barcode management |
| `product_category_bloc` | Category |
| `product_group_bloc` | Product group |
| `product_dimension_bloc` | Dimension |
| `product_type_bloc` | Type |
| `master_brand_bloc` ~ `master_pattern_bloc` | AI classification (10 levels) |
| `option_bloc` | Product options |
| `selected_products_bloc` | Multi-product selection |
| `import_product_bloc` | Product import |

### Repositories (14 total)
`product_repository`, `product_master_repository`, `product_barcode_repository`, `product_category_repository`, `product_group_repository`, `product_dimension_reporsitory`, `product_type_repository`, `product_section_repository`, `product_import_repository`, `master_brand_repository` ~ `master_pattern_repository` (10 total)

### Key Models
`product_model`, `product_barcode_model`, `product_category_model`, `product_group_model`, `product_type_model`, `product_dimension_model`, `product_bom_model`, `product_condition_model`, `option_model`, `import_product_model`

### Cart System Module
Cart system located at `lib/modules/cart_system/` (35 files):
- **Cubit**: `cart_cubit`, `product_search_cubit`
- **Services**: Thai word tokenizer, transliteration, ClickHouse/PG/MongoDB product search
- **Widgets**: cart selector, product search, creditor/debtor/warehouse selector dialogs

## Important Notes
- A single product can have multiple barcodes (different units/prices)
- BOM = Bill of Materials, e.g. "Rice Curry Set" consists of rice + curry + water
- Options = choices, e.g. Color: Red/Blue, Size: S/M/L
- AI classification has 10 levels — used for automatic categorization
- Price history is stored in barcode (every price change is recorded)
- Products must sync: MongoDB -> Kafka -> PG + ClickHouse

## Related Skills
- `/stock-inventory` — manage stock balances, receive, transfer, and costing per product
- `/master-data` — CRUD for units, barcodes, categories, creditors, debtors via MCP tools
- `/auto-packing` — BOM-based pick & pack, unit conversion, packing list
- `/import-export` — bulk import products from Excel, export reports
- `/coupon-promotion` — product-specific coupons and promotions
