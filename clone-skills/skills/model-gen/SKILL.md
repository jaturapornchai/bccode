---
name: model-gen
description: >
  View backend data model schema and generate Dart classes. Inspect field types, DB tables,
  sample data, sync frontend-backend models. Use when the user mentions model schema, Dart
  class generation, or DB table inspection.
user-invocable: true
---

# Model Schema -- View & Generate Data Models

## Usage
`/model-gen <model_name>` -- View model schema
`/model-gen <model_name> dart` -- View schema + generate Dart class

## Workflow

### 1. View Model Schema
Call MCP tool `get_model_schema`:
```
Tool: get_model_schema
Parameters:
  - name: "ProductDoc" | "SaleInvoiceDoc" | "UnitDoc" | ...
```

### 2. Display Schema
```
## Model: ProductDoc

| Field | Type | JSON Tag | Required | Description |
|-------|------|----------|----------|-------------|
| guidfixed | string | guidfixed | yes | Primary key |
| code1 | string | code1 | yes | Product code |
| name1 | string | name1 | yes | Product name (TH) |
| name2 | string | name2 | no | Product name (EN) |
| price | float64 | price | no | Sale price |
```

### 3. Generate Dart Class (if requested)
```dart
class ProductModel {
  String guidfixed;
  String code1;
  String name1;
  String name2;
  double price;

  ProductModel({
    this.guidfixed = '',
    this.code1 = '',
    this.name1 = '',
    this.name2 = '',
    this.price = 0,
  });

  factory ProductModel.fromJson(Map<String, dynamic> json) {
    return ProductModel(
      guidfixed: json['guidfixed'] ?? '',
      code1: json['code1'] ?? '',
      name1: json['name1'] ?? '',
      name2: json['name2'] ?? '',
      price: (json['price'] ?? 0).toDouble(),
    );
  }

  Map<String, dynamic> toJson() => {
    'guidfixed': guidfixed,
    'code1': code1,
    'name1': name1,
    'name2': name2,
    'price': price,
  };
}
```

### 4. Inspect DB Schema (optional)
Call MCP tool `get_database_schema` for table structure:
```
Tool: get_database_schema
Parameters:
  - table: "products"
```

Call MCP tool `get_table_sample` for sample data:
```
Tool: get_table_sample
Parameters:
  - table: "products"
  - limit: 3
```

## Dart Model Rules
- **fromJson**: Use `??` default value for every field (prevent null)
- **String**: default `''`
- **int/double**: default `0` / `0.0`
- **bool**: default `false`
- **List**: default `[]`
- **DateTime**: parse from string with `DateTime.tryParse()`
- **Nested object**: create separate factory
- **File location**: `lib/model/` directory

## Examples
```
/model-gen ProductDoc
/model-gen SaleInvoiceDoc
/model-gen UnitDoc
/model-gen BarcodeDoc
/model-gen ProductDoc dart    -> Generate Dart class
```

## MCP Tools
| Tool | Purpose |
|------|---------|
| `get_model_schema` | View Go struct fields + types |
| `get_database_schema` | View PostgreSQL table structure |
| `get_table_sample` | View sample data from table |

## Notes
- `get_model_schema` is available on MCP Dev endpoint only
- Always verify JSON tags match `fromJson` keys
- New backend fields are safe (frontend ignores unknown fields)
- Removed/renamed backend fields are breaking (frontend must update)

## When to Use
- Creating a new Dart model for an API endpoint response
- Syncing frontend model after backend adds/changes fields
- Checking field types and JSON tags before writing `fromJson`
- Need to view DB table structure to understand the data model
- Creating a complex model with nested objects or List fields

## Anti-Patterns
- Do NOT manually edit `.g.dart` files — always regenerate with `build_runner`
- Do NOT guess field names from model name — JSON tag may differ from Go field name
- Do NOT forget `??` default in `fromJson` — prevents null crashes
- Do NOT copy model from old files without checking — schema may have changed
- Do NOT create model in the wrong directory — must be in `lib/model/`

## Related Skills
- `/api-spec` — View response schema before generating model
- `/enum-list` — View enum values for fields that are enums
- `/api-search` — Find endpoints that use this model
