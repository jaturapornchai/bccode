---
name: import-export
description: >
  Data import and export: import products from Excel/CSV/images, import opening stock balances,
  export reports as CSV/PDF, chunked file upload, and OCR document reading. Use when the user
  mentions importing data, exporting reports, uploading files, or printing PDF documents.
user-invocable: true
---

# Data Import & Export

## Overview
Import products from Excel/images, export reports as CSV/PDF, file uploads.

## When to Use
- User wants to bulk-import products from an Excel or CSV file
- User needs to import opening stock balances at the start of a new period
- User wants to export a report (sales, inventory, financial) as CSV or PDF
- User is uploading large files (many product images, large Excel) — chunked upload applies
- User wants to read data from a document image using OCR

## Anti-Patterns
- Do NOT import products without previewing and validating first — duplicate barcodes will be rejected silently
- Do NOT change PDF colors dynamically — `PdfColors` are fixed, do NOT theme them
- Do NOT bypass chunked upload for large files — standard `/upload` will time out
- Do NOT import stock balances mid-period without closing the period first — opening balance only
- Do NOT store uploaded images outside R2/SeaweedFS — file references will break across environments

## System Structure

### Product Import
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/import/import_product_screen.dart` | Main import screen |
| From File | `screens/import/import_product_from_file_screen.dart` | Import from Excel/CSV |
| Validation | `screens/import/import_product_detail_screen.dart` | Review data before import |
| From Image | `screens/import/import_product_image_screen.dart` | Import product images |
| BLoC | `bloc/import_product/` | State management |
| Backend | `backend/internal/productimport/` | Process import |

### Stock Import
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/stockbalanceimport/` | Import opening stock balances |
| Shelf Excel | `bloc/shelf_excel_import/` | Import shelf data from Excel |

### File Upload
| Section | File | Purpose |
|---------|------|---------|
| Images | `backend/internal/images/` | Azure blob + R2 storage |
| Slip | `backend/internal/slipimage/` | Upload slips/receipts |
| Documents | `backend/internal/documentwarehouse/` | Document warehouse |

### GoAPI Upload (Chunked)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/upload` | Standard file upload |
| POST | `/upload/init` | Start chunked upload |
| POST | `/upload/chunk` | Send chunk |
| POST | `/upload/merge` | Merge chunks |
| GET | `/upload/status/:uploadID` | Check status |
| DELETE | `/upload/cancel/:uploadID` | Cancel upload |

### Data Export
| Section | File | Purpose |
|---------|------|---------|
| CSV | `bloc/export_csv/` | CSV export |
| PDF | `screens/report/pdf_report_*` | PDF report generation |
| Download Web | `screens/report/file_download_web.dart` | Download via browser |
| Download Mobile | `screens/report/file_download_mobile.dart` | Save to mobile device |
| Download Desktop | `screens/report/file_download_desktop.dart` | Save to PC |

### GoAPI PDF/Report
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/genpdf` | Generate PDF |
| GET | `/genpdf/history` | PDF history |
| GET | `/genpdf/reprint/:id` | Reprint |
| POST | `/resulttopdf` | Convert results to PDF |
| POST | `/xlsx/product/start` | Start import from Excel |

### Image
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/image/upload` | Upload image |
| POST | `/image/list` | List images |
| POST | `/image/get` | Get image |
| POST | `/image/delete` | Delete image |
| POST | `/image/promptpayverify` | Verify PromptPay QR |

## Product Import Flow
```
Select Excel/CSV file
  |
System reads data → Display preview
  |
Validate: Duplicates? Data complete? Format correct?
  |
Confirm → Save to MongoDB
  |
Kafka → sync PG + ClickHouse
```

## Important Notes
- Chunked upload is for large files (many images, large Excel files)
- Product import must check for duplicate barcodes before saving
- PDF uses fixed PdfColors — do not change based on theme
- Image storage uses R2 (Cloudflare) or SeaweedFS
- OCR (`backend/internal/ocr/`) reads data from document images
- `promptpayverify` verifies whether a PromptPay QR code payment was actually made

## Related Skills
- `/product` — product data that is the target of Excel import
- `/stock-inventory` — opening stock balances imported via `stockbalanceimport`
- `/master-data` — use MCP tools for smaller batch creates instead of Excel import
- `/auto-packing` — shelf structure can be imported via `shelf_excel_import`
