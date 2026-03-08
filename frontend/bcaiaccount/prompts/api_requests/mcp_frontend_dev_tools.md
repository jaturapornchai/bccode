# API Request: MCP Tools สำหรับ Frontend Developer

## สิ่งที่ต้องการ
เพิ่ม MCP tools ที่ช่วย AI frontend developer (Claude Code, Cursor, etc.) ให้สามารถดู API endpoints, request/response spec, และ enum values ของ backend ได้โดยตรงผ่าน MCP — ไม่ต้องไปอ่าน source code backend เอง

ตอนนี้ MCP มี tools สำหรับ Business Analytics (sales, dashboard, inventory, customer) กับ Database query แล้ว แต่ยังขาด tools ที่ช่วย frontend dev โดยเฉพาะ

## MCP Tools ที่ต้องการ (4 tools)

---

### 1. `list_api_endpoints`
- **Description:** แสดง API endpoints ทั้งหมดที่ backend มี พร้อม method, path, auth requirement, และคำอธิบายสั้นๆ
- **Parameters:**
  - `group` (string, optional) — filter ตามกลุ่ม เช่น "auth", "shop", "product", "transaction", "organization", "goapi"
  - `method` (string, optional) — filter ตาม HTTP method เช่น "GET", "POST", "PUT", "DELETE"
  - `search` (string, optional) — ค้นหา path หรือ description
- **Expected Response:**
```json
{
  "total": 45,
  "endpoints": [
    {
      "method": "GET",
      "path": "/shop/{id}",
      "group": "shop",
      "auth_required": true,
      "description": "Get shop details by ID",
      "request_content_type": "application/json",
      "response_content_type": "application/json"
    },
    {
      "method": "POST",
      "path": "/login",
      "group": "auth",
      "auth_required": false,
      "description": "Login with username/password",
      "request_content_type": "application/json",
      "response_content_type": "application/json"
    }
  ]
}
```

---

### 2. `get_api_spec`
- **Description:** ดู request/response specification ของ API endpoint ที่ระบุ พร้อม field types, required/optional, validation rules
- **Parameters:**
  - `method` (string, required) — HTTP method เช่น "GET", "POST"
  - `path` (string, required) — API path เช่น "/shop/{id}", "/login"
- **Expected Response:**
```json
{
  "method": "POST",
  "path": "/login",
  "description": "Login with username/password",
  "auth_required": false,
  "request": {
    "content_type": "application/json",
    "fields": [
      {"name": "username", "type": "string", "required": true, "description": "Username or email"},
      {"name": "password", "type": "string", "required": true, "description": "Password"}
    ]
  },
  "response": {
    "success_example": {
      "success": true,
      "data": {
        "token": "string — JWT token",
        "user": {
          "guid": "string",
          "username": "string",
          "email": "string"
        }
      }
    },
    "error_example": {
      "success": false,
      "message": "invalid credentials"
    }
  },
  "go_handler": "handlers/auth.go:Login",
  "go_model": "models/user.go:UserLoginRequest"
}
```

---

### 3. `get_api_example`
- **Description:** ดูตัวอย่าง request/response จริงจาก API endpoint (ใช้ sample data จาก DB)
- **Parameters:**
  - `method` (string, required) — HTTP method
  - `path` (string, required) — API path
- **Expected Response:**
```json
{
  "method": "GET",
  "path": "/shop/{id}",
  "example_request": {
    "url": "/shop/abc-123-def",
    "headers": {
      "Authorization": "Bearer <token>"
    }
  },
  "example_response": {
    "success": true,
    "data": {
      "guidfixed": "abc-123-def",
      "names1": "ร้านค้าตัวอย่าง",
      "taxid": "1234567890123",
      "branchnumber": "00000"
    }
  },
  "note": "Response based on actual sample data from database"
}
```

---

### 4. `list_enums`
- **Description:** แสดง enum values / constants ที่ backend ใช้ เช่น status codes, document types, transaction flags
- **Parameters:**
  - `search` (string, optional) — ค้นหา enum name หรือ value
  - `group` (string, optional) — filter ตามกลุ่ม เช่น "transaction", "document", "status"
- **Expected Response:**
```json
{
  "enums": [
    {
      "name": "TransFlag",
      "group": "transaction",
      "description": "Transaction status flags",
      "values": [
        {"value": 0, "label": "Draft", "description": "เอกสารร่าง"},
        {"value": 2, "label": "Approved", "description": "อนุมัติแล้ว"},
        {"value": 6, "label": "Completed", "description": "เสร็จสิ้น"}
      ]
    },
    {
      "name": "DocType",
      "group": "document",
      "description": "Document types",
      "values": [
        {"value": 1, "label": "Invoice", "description": "ใบแจ้งหนี้"},
        {"value": 2, "label": "Receipt", "description": "ใบเสร็จรับเงิน"}
      ]
    }
  ]
}
```

---

## Use Case (Frontend)

### AI Frontend Developer ใช้ทำอะไร:
1. **สร้าง Dart model ใหม่** → ใช้ `get_api_spec` ดู response fields + types → สร้าง model ตรง 100%
2. **เขียน repository/API call** → ใช้ `list_api_endpoints` ดูว่ามี endpoint อะไรบ้าง ไม่ต้องเดา path
3. **Debug API issues** → ใช้ `get_api_example` ดูตัวอย่าง response จริง เทียบกับที่ได้
4. **จัดการ status/type** → ใช้ `list_enums` ดู enum values ที่ถูกต้อง ไม่ต้อง hardcode ผิด
5. **ลดการอ่าน backend source code** → AI ดูผ่าน MCP ได้เลย เร็วกว่าและแม่นกว่า

### ปัญหาที่แก้ได้:
- AI frontend dev ต้องอ่าน backend Go source code เพื่อรู้ว่า API return อะไร → ช้าและอาจพลาด
- Field names/types ไม่ตรงกันระหว่าง frontend กับ backend → bugs
- ไม่รู้ว่า backend มี API อะไรบ้าง → เขียน API request prompt ที่ซ้ำกับของที่มีอยู่แล้ว
- Enum values hardcode ผิด → logic ผิดพลาด

## แนวทาง Implementation (Suggestion)

### วิธีที่ 1: Auto-generate จาก Go source code (แนะนำ)
- Parse Go route registration (e.g., `r.GET("/shop/:id", handler)`)
- Parse Go struct tags (`json:"field_name"`)
- Extract enum constants
- สร้าง tool response อัตโนมัติ

### วิธีที่ 2: Swagger/OpenAPI
- ถ้า backend มี Swagger อยู่แล้ว (`/swagger/*`) → อ่านจาก Swagger spec
- เพิ่ม MCP tools เป็น wrapper ที่อ่าน Swagger JSON แล้ว format ให้เหมาะกับ AI

### วิธีที่ 3: Manual registration
- สร้าง registry file ที่ list endpoints + specs
- ง่ายที่สุดแต่ต้อง maintain manual

## Priority
**สูง** — tools เหล่านี้จะช่วยลดเวลา dev frontend ลงมาก และลด bugs จากการ mismatch ระหว่าง frontend/backend
