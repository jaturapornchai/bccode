# API Request: เพิ่ม MCP Tools สำหรับ Resolve Swagger $ref + Helpers

## บริบท
ตอนนี้ MCP Frontend Dev Tools ทำงานได้ดีแล้ว (5 tools, 1,427 endpoints)
แต่มีปัญหาที่เหลืออยู่ 3 ข้อ ซึ่งแก้ได้โดยเพิ่ม/ปรับ MCP tools

Frontend จะสร้าง **Claude Code Skill** ที่ wrap MCP tools เหล่านี้
เพื่อให้ AI frontend dev ใช้งานได้สมบูรณ์ โดยไม่ต้องอ่าน backend source code

## สิ่งที่ต้องการ (3 ข้อ)

---

### 1. แก้ `get_api_spec` — Resolve `$ref` ให้อัตโนมัติ (สำคัญที่สุด)

**ปัญหา:**
MainAPI endpoints ที่ดึงจาก Swagger แสดง schema เป็น `$ref` ซึ่ง AI อ่านไม่ได้:
```json
{
  "path": "/login",
  "request_body": {
    "schema": { "$ref": "#/definitions/models.UserLoginRequest" }
  },
  "response": {
    "schema": { "$ref": "#/definitions/models.AuthResponse" }
  }
}
```

**สิ่งที่ต้องการ (เลือกวิธีใดวิธีหนึ่ง):**

#### วิธี A: แก้ `get_api_spec` ให้ expand $ref เลย (แนะนำ)
เมื่อ response มี `$ref` → resolve จาก Swagger definitions → เพิ่ม `fields[]` ออกมาด้วย:
```json
{
  "path": "/login",
  "request_body": {
    "schema": { "$ref": "#/definitions/models.UserLoginRequest" },
    "fields": [
      {"name": "username", "type": "string", "required": true},
      {"name": "password", "type": "string", "required": true}
    ]
  },
  "response": {
    "schema": { "$ref": "#/definitions/models.AuthResponse" },
    "fields": [
      {"name": "success", "type": "boolean"},
      {"name": "data.token", "type": "string"},
      {"name": "data.refresh_token", "type": "string"}
    ]
  }
}
```

#### วิธี B: เพิ่ม MCP tool `resolve_swagger_ref` แยกต่างหาก
- **Tool name:** `resolve_swagger_ref`
- **Description:** Resolve Swagger `$ref` path ให้เป็น field list พร้อม types
- **Parameters:**
  - `ref` (string, required) — เช่น `"#/definitions/models.AuthResponse"` หรือ `"models.AuthResponse"`
- **Expected Response:**
```json
{
  "ref": "#/definitions/models.AuthResponse",
  "model_name": "AuthResponse",
  "source": "swagger",
  "fields": [
    {"name": "success", "type": "boolean", "required": true},
    {"name": "data", "type": "object", "fields": [
      {"name": "token", "type": "string"},
      {"name": "refresh_token", "type": "string"},
      {"name": "expired_at", "type": "string", "format": "date-time"}
    ]}
  ],
  "dart_types": {
    "success": "bool",
    "data.token": "String",
    "data.refresh_token": "String",
    "data.expired_at": "DateTime"
  }
}
```

#### วิธี C: เพิ่ม MCP tool `get_swagger_definitions`
- **Tool name:** `get_swagger_definitions`
- **Description:** ดู Swagger definitions ทั้งหมดหรือค้นหาตามชื่อ
- **Parameters:**
  - `keyword` (string, optional) — ค้นหา definition name เช่น `"Auth"`, `"Shop"`, `"Product"`
  - `name` (string, optional) — ชื่อ definition ตรงๆ เช่น `"models.AuthResponse"`
- **Expected Response:**
```json
{
  "definitions": [
    {
      "name": "models.AuthResponse",
      "type": "object",
      "properties": {
        "success": {"type": "boolean"},
        "data": {
          "type": "object",
          "properties": {
            "token": {"type": "string"},
            "refresh_token": {"type": "string"}
          }
        }
      }
    }
  ],
  "total": 1
}
```

**แนะนำวิธี A** เพราะไม่ต้องเพิ่ม tool ใหม่ แก้ที่ `get_api_spec` เดิมได้เลย

---

### 2. แก้ MainAPI Description ที่ไม่ถูกต้อง

**ปัญหา:**
Swagger annotation หลายตัวเขียน description ไม่ดี:
```
POST /login          → "get struct array by ID"    (ผิด)
GET /shop/branch/list → "search limit offset"      (ไม่ชัด)
GET /product/list     → "search limit offset"      (ไม่ชัด)
```

**วิธีแก้ที่แนะนำ:**
เพิ่ม description override ใน MCP catalog สำหรับ endpoints ที่สำคัญ:

```go
// ใน api_catalog.go หรือที่เกี่ยวข้อง
var mainAPIDescOverrides = map[string]string{
    // Auth
    "POST /login":                "Login with username/password, returns JWT access token and refresh token",
    "POST /login/email":          "Login with email/password (POS login)",
    "POST /register":             "Register new user account",
    "POST /register-phonenumber": "Register with phone number (OTP verification)",
    "POST /refresh":              "Refresh expired JWT access token using refresh token",
    "POST /forgot-password":      "Request password reset via email",
    "POST /forgot-password-phonenumber": "Request password reset via phone OTP",

    // Shop
    "GET /shop/{id}":             "Get shop details by GUID",
    "POST /shop":                 "Create new shop",
    "PUT /shop/{id}":             "Update shop details",
    "GET /shop/branch/list":      "List all branches with search and pagination",
    "GET /shop/employee":         "List employees with search and pagination",

    // Product
    "GET /product/list":          "List products with search, category filter, and pagination",
    "GET /product/{id}":          "Get product details by GUID",
    "POST /product":              "Create new product",
    "PUT /product/{id}":          "Update product details",
    "DELETE /product":            "Delete products by GUID list",

    // Transaction
    "GET /transaction/purchase-order":  "List purchase orders with filters",
    "POST /transaction/purchase-order": "Create new purchase order",
    "GET /transaction/purchase-return": "List purchase returns with filters",

    // Warehouse
    "GET /warehouse/list":        "List warehouses with search and pagination",
    "GET /warehouse/{id}":        "Get warehouse details by GUID",

    // Organization
    "GET /organization/branch/list":    "List organization branches",
    "PUT /organization/branch/{guid}":  "Update branch settings and details",
}
```

**Logic:**
```go
// เมื่อ build endpoint catalog จาก Swagger
if override, ok := mainAPIDescOverrides[method+" "+path]; ok {
    endpoint.Description = override
}
```

---

### 3. เพิ่ม Enums ที่ Flutter ใช้แต่ยังไม่มีใน `list_enums`

เพิ่ม enums ต่อไปนี้เข้า enum catalog:

```go
// doc_status — สถานะเอกสาร (ใช้ทุกหน้า transaction)
{
    Name: "doc_status",
    Category: "transaction",
    Description: "Document status — สถานะเอกสาร",
    Values: []EnumValue{
        {Key: 0, Label: "Draft", Description: "ร่าง"},
        {Key: 1, Label: "Active", Description: "ใช้งาน"},
        {Key: 2, Label: "Approved", Description: "อนุมัติ"},
        {Key: 3, Label: "Void", Description: "ยกเลิก"},
        {Key: 4, Label: "Cancelled", Description: "ยกเลิก (ลบ)"},
    },
}

// user_role — บทบาท user (ใช้ใน permission check)
{
    Name: "user_role",
    Category: "system",
    Description: "User role — บทบาทผู้ใช้",
    Values: []EnumValue{
        {Key: "owner", Label: "Owner", Description: "เจ้าของร้าน"},
        {Key: "admin", Label: "Admin", Description: "ผู้ดูแลระบบ"},
        {Key: "manager", Label: "Manager", Description: "ผู้จัดการ"},
        {Key: "staff", Label: "Staff", Description: "พนักงาน"},
        {Key: "viewer", Label: "Viewer", Description: "ดูอย่างเดียว"},
    },
}

// shop_status — สถานะร้าน
{
    Name: "shop_status",
    Category: "system",
    Description: "Shop status — สถานะร้านค้า",
    Values: []EnumValue{
        {Key: 0, Label: "Active", Description: "ใช้งาน"},
        {Key: 1, Label: "Suspended", Description: "ระงับ"},
        {Key: 2, Label: "Deleted", Description: "ลบแล้ว (soft delete)"},
    },
}

// product_type — ประเภทสินค้า
{
    Name: "product_type",
    Category: "master",
    Description: "Product type — ประเภทสินค้า",
    Values: []EnumValue{
        {Key: 0, Label: "Goods", Description: "สินค้าทั่วไป (มี stock)"},
        {Key: 1, Label: "Service", Description: "บริการ (ไม่มี stock)"},
        {Key: 2, Label: "Non-Stock", Description: "ไม่ติดตาม stock"},
    },
}

// year_type — ประเภทปี (ใช้ใน branch settings)
{
    Name: "year_type",
    Category: "system",
    Description: "Year type — ประเภทปีปฏิทิน",
    Values: []EnumValue{
        {Key: "buddhist", Label: "Buddhist Era", Description: "พุทธศักราช (+543)"},
        {Key: "christian", Label: "Christian Era", Description: "คริสต์ศักราช"},
    },
}

// vat_type — ประเภทภาษี (ใช้ใน transaction)
{
    Name: "vat_type",
    Category: "transaction",
    Description: "VAT type — ประเภทภาษีมูลค่าเพิ่ม",
    Values: []EnumValue{
        {Key: 0, Label: "Excluded", Description: "ราคาไม่รวม VAT"},
        {Key: 1, Label: "Included", Description: "ราคารวม VAT แล้ว"},
        {Key: 2, Label: "Non-Taxable", Description: "ไม่คิดภาษี"},
    },
}
```

**หมายเหตุ:** ค่า enum ข้างบนเป็นตัวอย่าง — กรุณาตรวจสอบกับ source code จริงและแก้ไขให้ตรงก่อน register

---

## สรุป

| # | สิ่งที่ต้องทำ | Priority | วิธีแก้ |
|---|-------------|----------|---------|
| 1 | Resolve `$ref` ใน `get_api_spec` | 🟡 สำคัญสุด | แก้ `get_api_spec` ให้ expand fields จาก Swagger definitions |
| 2 | แก้ MainAPI description | 🟡 Medium | เพิ่ม description override mapping |
| 3 | เพิ่ม enums ที่ขาด | 🟢 Nice to Have | เพิ่ม 6 enums เข้า catalog |

## Frontend จะทำอะไรต่อ
หลัง backend แก้แล้ว frontend จะสร้าง **Claude Code Skill** ที่:
1. Wrap MCP tools ให้ใช้ง่ายขึ้น (เช่น `/api-spec login` แทนการเรียก MCP ตรง)
2. Auto-generate Dart model จาก `get_api_spec` + `get_model_schema`
3. แสดงผลสรุปให้อ่านง่าย (แทน raw JSON)
