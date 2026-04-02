---
name: enum-list
description: List backend enum values (doc status, product type, payment type) — use to check or sync enums
user-invocable: true
---

# Enum Catalog — List Enums

## Usage
`/enum-list` — Show all enums
`/enum-list <keyword>` — Search enums by keyword

## Steps

### 1. List Enums
Call MCP tool `list_enums`:
```
Tool: list_enums
Parameters:
  - keyword: "payment" (optional — filter by keyword)
```

### 2. Display Results
```
## Enum: DocStatus

| Value | Label (TH) | Label (EN) |
|-------|-----------|-----------|
| 0 | ร่าง | Draft |
| 1 | รออนุมัติ | Pending |
| 2 | อนุมัติ | Approved |
| 3 | ยกเลิก | Cancelled |

## Enum: PaymentType

| Value | Label |
|-------|-------|
| cash | Cash |
| transfer | Bank Transfer |
| credit | Credit |
| qr | QR Payment |
```

### 3. Suggest Flutter Usage
```dart
// Create Dart enum matching backend
enum DocStatus {
  draft(0, 'Draft'),
  pending(1, 'Pending'),
  approved(2, 'Approved'),
  cancelled(3, 'Cancelled');

  final int value;
  final String label;
  const DocStatus(this.value, this.label);
}
```

## Examples
```
/enum-list                → show all
/enum-list payment        → payment-related enums
/enum-list doc            → document-related enums
/enum-list status         → status-related enums
/enum-list product        → product-related enums
```

## MCP Tools Used
| Tool | Purpose |
|------|---------|
| `list_enums` | List all enum values |

## Notes
- `list_enums` is on MCP Dev endpoint only
- Enum values must match between frontend and backend
- If backend adds new enum values → frontend must handle them (default case)
