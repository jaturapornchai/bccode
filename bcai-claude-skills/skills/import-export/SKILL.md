---
name: import-export
description: Import Excel images Export CSV PDF upload — Import & Export
user-invocable: true
---

# Import & Export

## Overview
Import products from Excel/images, export reports as CSV/PDF, upload files.

## System Structure

### Product Import
| Component | File | Purpose |
|-----------|------|---------|
| Screen | `screens/import/import_product_screen.dart` | Main import screen |
| From File | `screens/import/import_product_from_file_screen.dart` | Import from Excel/CSV |
| Validation | `screens/import/import_productdetail_screen.dart` | Verify data before import |
| From Image | `screens/import/import_product_image_screen.dart` | Import product images |
| BLoC | `bloc/import_product/` | State management |
| Backend | `backend/internal/productimport/` | Import processing |

### Stock Import
| Component | File | Purpose |
|-----------|------|---------|
| Backend | `backend/internal/stockbalanceimport/` | Import opening stock balance |
| Shelf Excel | `bloc/shelf_excel_import/` | Import shelf data from Excel |

### File Upload
| Component | File | Purpose |
|-----------|------|---------|
| Images | `backend/internal/images/` | Azure blob + R2 storage |
| Slip | `backend/internal/slipimage/` | Upload slip/receipt |
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

### Export
| Component | File | Purpose |
|-----------|------|---------|
| CSV | `bloc/export_csv/` | Export CSV |
| PDF | `screens/report/pdf_report_*` | Generate PDF reports |
| Download Web | `screens/report/file_download_web.dart` | Download via browser |
| Download Mobile | `screens/report/file_download_mobile.dart` | Save to mobile device |
| Download Desktop | `screens/report/file_download_desktop.dart` | Save to PC |

### GoAPI PDF/Report
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/genpdf` | Generate PDF |
| GET | `/genpdf/history` | PDF history |
| GET | `/genpdf/reprint/:id` | Reprint |
| POST | `/resulttopdf` | Convert result to PDF |
| POST | `/xlsx/product/start` | Start import from Excel |

### Image
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/image/upload` | Upload image |
| POST | `/image/list` | List images |
| POST | `/image/get` | Get image |
| POST | `/image/delete` | Delete image |
| POST | `/image/promptpayverify` | Verify PromptPay QR code |

## Product Import Flow
```
Select Excel/CSV file
  |
System reads data -> show preview
  |
Validate: duplicates? complete data? correct format?
  |
Confirm -> save to MongoDB
  |
Kafka -> sync PG + ClickHouse
```

## Important Notes
- Chunked upload is for large files (many images, large Excel files)
- Product import must check for duplicate barcodes before saving
- PDF uses fixed PdfColors — do not change based on theme
- Image storage uses R2 (Cloudflare) or SeaweedFS
- OCR (`backend/internal/ocr/`) reads data from document images
- `promptpayverify` verifies PromptPay QR codes to confirm actual payment
