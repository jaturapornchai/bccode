---
name: enum-list
description: >
  List backend enum values (doc status, product type, payment type) — use to check or sync
  enums. Use when the user needs enum values, status codes, or type constants.
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

| Value | Label |
|-------|-------|
| 0 | Draft |
| 1 | Pending |
| 2 | Approved |
| 3 | Cancelled |

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

## When to Use
- Need to know the numeric value of `DocStatus` (e.g., is approved = 2 or 3?)
- Creating a Dart enum to match backend constants
- Checking supported payment types before implementing UI
- Debugging why a status filter is not working — values may not match
- Need Thai/English labels for each enum value

## Anti-Patterns
- Do NOT hardcode enum values like `status == 2` without lookup — use this skill to verify first
- Do NOT guess enum values from names — `cancelled` may be 3 or 9 depending on the system
- Do NOT create Dart enums before checking backend values — they must stay in sync
- Do NOT forget a default case in switch — backend may add new enum values
- Do NOT use magic numbers in filter queries — reference the enum name instead

## Related Skills
- `/model-gen` — View the fields that use this enum in the model
- `/api-spec` — Check how an endpoint accepts enum values (int or string)
- `/api-search` — Find endpoints related to this enum
