---
name: model-gen
description: ดู data model schema จาก backend และสร้าง Dart class — ใช้เมื่อต้องการดูโครงสร้าง model, สร้าง Dart model ใหม่จาก backend schema, ตรวจสอบ field types, ดู DB table structure, ดูข้อมูลตัวอย่าง, หรือ sync model ระหว่าง frontend-backend. ใช้เมื่อพูดถึง struct, schema, fromJson, toJson, model fields
user-invocable: true
---

# Model Schema — ดูโครงสร้าง Data Model

## วิธีใช้
`/model-gen <model_name>` — ดู schema ของ model
`/model-gen <model_name> dart` — ดู schema + สร้าง Dart class

## ขั้นตอนการทำงาน

### 1. ดู Model Schema
เรียก MCP tool `get_model_schema`:
```
Tool: get_model_schema
Parameters:
  - name: "ProductDoc" | "SaleInvoiceDoc" | "UnitDoc" | ...
```

### 2. แสดง Schema
```
## Model: ProductDoc

| Field | Type | JSON Tag | Required | Description |
|-------|------|----------|----------|-------------|
| guidfixed | string | guidfixed | yes | Primary key |
| code1 | string | code1 | yes | รหัสสินค้า |
| name1 | string | name1 | yes | ชื่อสินค้า (ไทย) |
| name2 | string | name2 | no | ชื่อสินค้า (อังกฤษ) |
| price | float64 | price | no | ราคาขาย |
```

### 3. สร้าง Dart Class (ถ้าขอ)
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

### 4. ตรวจสอบ DB Schema เพิ่มเติม (ถ้าต้องการ)
เรียก MCP tool `get_database_schema` เพื่อดู table structure:
```
Tool: get_database_schema
Parameters:
  - table: "products"
```

เรียก MCP tool `get_table_sample` เพื่อดูข้อมูลตัวอย่าง:
```
Tool: get_table_sample
Parameters:
  - table: "products"
  - limit: 3
```

## กฎสร้าง Dart Model
- **fromJson**: ใช้ `??` default value ทุก field (ป้องกัน null)
- **String**: default `''`
- **int/double**: default `0` / `0.0`
- **bool**: default `false`
- **List**: default `[]`
- **DateTime**: parse จาก string ด้วย `DateTime.tryParse()`
- **Nested object**: สร้าง factory แยก
- **วางไฟล์**: `lib/model/` directory

## ตัวอย่าง
```
/model-gen ProductDoc
/model-gen SaleInvoiceDoc
/model-gen UnitDoc
/model-gen BarcodeDoc
/model-gen ProductDoc dart    → สร้าง Dart class ด้วย
```

## MCP Tools ที่ใช้
| Tool | หน้าที่ |
|------|--------|
| `get_model_schema` | ดู Go struct fields + types |
| `get_database_schema` | ดู PostgreSQL table structure |
| `get_table_sample` | ดูข้อมูลตัวอย่างจาก table |

## หมายเหตุ
- `get_model_schema` อยู่ใน MCP Dev endpoint เท่านั้น
- ตรวจสอบ JSON tag ให้ตรงกับ `fromJson` key เสมอ
- ถ้า backend เพิ่ม field ใหม่ → frontend เก่ายังทำงานได้ (safe)
- ถ้า backend ลบ/เปลี่ยนชื่อ field → frontend ต้องแก้ด้วย (breaking)
