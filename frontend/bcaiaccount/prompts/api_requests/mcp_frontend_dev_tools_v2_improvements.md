# API Request: ปรับปรุง MCP Frontend Dev Tools (V2.1)

## สถานะปัจจุบัน (ทดสอบ 26 ก.พ. 2026)

### ✅ แก้แล้ว — ขอบคุณทีม Backend!
| # | ปัญหาเดิม | ผลทดสอบ |
|---|-----------|---------|
| 1 | MainAPI endpoints หายหมด (0 ตัว) | **1,279 endpoints** จาก Swagger ✅ |
| 2 | `get_api_spec` ค้น MainAPI ไม่เจอ | `/login` → 2 results ✅ |
| 4 | `get_model_schema` ขาด MainAPI models | `Shop` → มีแล้ว (พร้อม fields, Dart types) ✅ |
| 5 | `list_enums` ขาด MainAPI enums | `approval_status`, `import_status` มาแล้ว ✅ |

### สรุปตัวเลข
- **GoAPI endpoints:** 148 ตัว
- **MainAPI endpoints:** 1,279 ตัว (จาก Swagger)
- **รวมทั้งหมด:** ~1,427 endpoints
- **MainAPI categories:** 130+ categories

---

## ปัญหาที่ยังเหลือ (3 ข้อ)

---

### 🟡 ปัญหาที่ 1: `get_api_spec` — MainAPI response schema ยังเป็น `$ref` ไม่ expand (Medium)

**อาการ:**
```json
// ตอนนี้ MainAPI endpoints แสดง schema เป็น $ref
{
  "path": "/login",
  "request_body": {
    "schema": {
      "$ref": "#/definitions/smlcloudplatform_internal_authentication_models.UserLoginRequest"
    }
  },
  "response": {
    "schema": {
      "$ref": "#/definitions/models.AuthResponse"
    }
  }
}
```

**ปัญหา:** AI อ่าน `$ref` ไม่ได้ ต้องไปค้นหา definition เอง ซึ่ง MCP ไม่มี tool ดู Swagger definitions

**สิ่งที่ควรเป็น:**
```json
{
  "path": "/login",
  "request_body": {
    "schema": {
      "$ref": "#/definitions/models.UserLoginRequest"
    },
    "fields": [
      {"name": "username", "type": "string", "required": true, "description": "Username"},
      {"name": "password", "type": "string", "required": true, "description": "Password"}
    ]
  },
  "response": {
    "schema": {
      "$ref": "#/definitions/models.AuthResponse"
    },
    "fields": [
      {"name": "success", "type": "boolean"},
      {"name": "token", "type": "string", "description": "JWT access token"},
      {"name": "refresh_token", "type": "string", "description": "Refresh token"}
    ]
  }
}
```

**วิธีแก้ที่แนะนำ:**
1. เมื่อ `get_api_spec` พบ `$ref` → resolve จาก Swagger `definitions` section → expand เป็น field list
2. หรือ link ไปที่ `get_model_schema` ให้ AI เรียกต่อได้ → เพิ่ม field `"model_name": "UserLoginRequest"` เพื่อให้ AI รู้ว่าต้อง query model ไหน

---

### 🟡 ปัญหาที่ 2: MainAPI description ไม่ถูกต้อง (Medium)

**อาการ:**
```json
// /login description ผิด
{ "path": "/login", "description": "get struct array by ID" }

// /shop/branch/list description ไม่ชัด
{ "path": "/shop/branch/list", "description": "search limit offset" }
```

**ปัญหา:** Description มาจาก Swagger annotation ที่เขียนไว้ไม่ดี → AI อ่านแล้วไม่เข้าใจว่า API ทำอะไร

**วิธีแก้ที่แนะนำ (เลือกอย่างใดอย่างหนึ่ง):**
1. **แก้ Swagger annotation ใน Go source** — เขียน `@Summary` ให้ถูกต้อง (ดีที่สุดแต่งานเยอะ)
2. **Override description ใน MCP catalog** — สร้าง mapping table สำหรับ endpoints ที่สำคัญ เช่น:
```go
descriptionOverrides := map[string]string{
    "POST /login":           "Login with username/password, returns JWT token",
    "POST /login/email":     "Login with email/password",
    "GET /shop/branch/list": "List all branches with search and pagination",
    "POST /register":        "Register new user account",
    "POST /refresh":         "Refresh expired JWT token",
}
```
3. **เพิ่ม fallback** — ถ้า description เป็น generic ("get struct array by ID", "search limit offset") → generate description จาก path + method แทน

---

### 🟢 ปัญหาที่ 3: `list_enums` ยังขาด enums บางตัวที่ Flutter ใช้ (Nice to Have)

**Enums ที่ยังไม่มี:**
| Enum | ใช้ทำอะไร | ใช้ที่ไหนใน Flutter |
|------|----------|-------------------|
| `doc_status` | สถานะเอกสาร (draft/approved/void/cancelled) | หน้า transaction list, doc detail |
| `user_role` | บทบาท user (owner/admin/staff/viewer) | permission check, user management |
| `shop_status` | สถานะร้าน (active/suspended/deleted) | shop selection screen |
| `product_type` | ประเภทสินค้า (goods/service/non-stock) | product form, stock logic |
| `year_type` | ประเภทปี (buddhist=543/christian=0) | date display, branch settings |

**วิธีแก้:** เพิ่ม enums จาก MainAPI constants/models เข้า enum catalog

---

## สรุป Priority (V2.1)

| # | ปัญหา | Priority | ผลกระทบ |
|---|--------|----------|---------|
| 1 | `$ref` ไม่ expand ใน get_api_spec | 🟡 Medium | AI อ่าน request/response schema ของ MainAPI ไม่ได้ |
| 2 | MainAPI description ไม่ถูกต้อง | 🟡 Medium | AI ไม่เข้าใจว่า API ทำอะไร ค้นหายาก |
| 3 | ขาด enums บางตัว | 🟢 Nice to Have | ต้อง hardcode enum values |

## ข้อ 1 สำคัญที่สุด
ถ้าแก้ข้อ 1 ได้ (`$ref` expand) จะทำให้ AI สร้าง Dart model จาก `get_api_spec` ได้เลย โดยไม่ต้องเรียก `get_model_schema` แยก — workflow จะ smooth ขึ้นมาก

## ผลลัพธ์ที่คาดหวัง
```bash
# ดู /login spec พร้อม expanded fields ✅
get_api_spec path="/login"
→ request fields: [username (string, required), password (string, required)]
→ response fields: [success (bool), token (string), refresh_token (string)]

# Description ที่อ่านรู้เรื่อง ✅
list_api_endpoints keyword="login"
→ "POST /login — Login with username/password, returns JWT token"

# Enums ครบ ✅
list_enums keyword="doc_status"
→ draft=0, approved=2, void=4, cancelled=6
```
