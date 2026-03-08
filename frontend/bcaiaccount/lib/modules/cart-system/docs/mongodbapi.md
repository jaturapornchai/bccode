# MongoDB Atlas API - ตัวอย่างการใช้งาน

## กฎสำคัญ

### สำหรับ UPDATE และ DELETE:
- **shopid, email, cartid** ต้องระบุทั้ง 3 ตัวเสมอ
- **ห้ามเป็นค่าว่าง** ทั้ง 3 ตัว
- **ชุดค่าทั้ง 3 ต้องไม่ซ้ำกัน** (unique combination)

### สำหรับ GET (Query):
- **ไม่บังคับทั้ง 3 ตัว** - ใช้ตัวใดก็ได้
- **ต้องมีอย่างน้อย 1 ตัว** (shopid, email, หรือ cartid)
- สามารถใช้เพียง 1 ตัว หรือ 2 ตัว หรือครบ 3 ตัว

---

## 1. เพิ่มข้อมูล (INSERT)

```powershell
$body = @{
    collection = "carts"
    shopid = "shop001"
    email = "customer@example.com"
    cartid = "cart_20250115_001"
    data = @{
        items = @(
            @{ product_id = "P001"; name = "เสื้อยืด"; qty = 2; price = 299 }
        )
        total = 598
        status = "active"
    }
    upsert = $true
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/update" `
    -Method POST -Body $body -ContentType "application/json"
```

**Response (สำเร็จ):**
```json
{
  "status": "success",
  "code": 200,
  "matched_count": 0,
  "modified_count": 0,
  "upserted_id": "67a1b2c3d4e5f6a7b8c9d0e1"
}
```

**Response (ซ้ำ - ถ้า upsert=false):**
```json
{
  "status": "success",
  "code": 200,
  "matched_count": 0,
  "modified_count": 0,
  "upserted_id": null
}
```

---

## 2. GET (Query) - ดึงข้อมูล

### ตัวอย่าง 1: ค้นหาด้วยทั้ง 3 ตัว (Exact Match)
```powershell
$body = @{
    shopid = "SHOP001"
    email = "user@example.com"
    cartid = "CART123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/atlas/get" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

### ตัวอย่าง 2: ค้นหาตาม shopid อย่างเดียว (ดู cart ทั้งหมดของร้าน)
```powershell
$body = @{
    shopid = "SHOP001"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/atlas/get" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

### ตัวอย่าง 3: ค้นหาตาม email อย่างเดียว (ดู cart ทั้งหมดของ user)
```powershell
$body = @{
    email = "user@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/atlas/get" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

### ตัวอย่าง 4: ค้นหาตาม cartid อย่างเดียว
```powershell
$body = @{
    cartid = "CART123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/atlas/get" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

### ตัวอย่าง 5: ค้นหา 2 ตัว (shopid + email)
```powershell
$body = @{
    shopid = "SHOP001"
    email = "user@example.com"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/atlas/get" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

**Response ตัวอย่าง:**
```json
{
  "status": "success",
  "results": [
    {
      "shopid": "SHOP001",
      "email": "user@example.com",
      "cartid": "CART123",
      "data": {
        "items": [{"product_id": "P001", "qty": 2}],
        "total": 500
      },
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

---

```powershell
$body = @{
    collection = "carts"
    shopid = "shop001"
    email = "customer@example.com"
    cartid = "cart_20250115_001"
    data = @{
        status = "checked_out"
        total = 598
        payment_method = "credit_card"
        checkout_at = (Get-Date).ToString("o")
    }
    upsert = $false
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/update" `
    -Method POST -Body $body -ContentType "application/json"
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "matched_count": 1,
  "modified_count": 1,
  "upserted_id": null
}
```

---

## 3. ลบข้อมูล (DELETE)

### 3.1 ลบ 1 รายการ

```powershell
$body = @{
    collection = "carts"
    shopid = "shop001"
    email = "customer@example.com"
    cartid = "cart_20250115_001"
    delete_many = $false
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/delete" `
    -Method POST -Body $body -ContentType "application/json"
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "deleted_count": 1
}
```

### 3.2 ลบหลายรายการ (ระวัง!)

```powershell
# ลบทุก cart ที่มี shopid, email, cartid ตรงกัน
$body = @{
    collection = "carts"
    shopid = "shop001"
    email = "customer@example.com"
    cartid = "cart_20250115_001"
    delete_many = $true
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/delete" `
    -Method POST -Body $body -ContentType "application/json"
```

---

## 4. ค้นหาข้อมูล (GET)

```powershell
$body = @{
    collection = "carts"
    shopid = "shop001"
    email = "customer@example.com"
    cartid = "cart_20250115_001"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/get" `
    -Method POST -Body $body -ContentType "application/json"
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "count": 1,
  "data": [
    {
      "_id": "67a1b2c3d4e5f6a7b8c9d0e1",
      "shopid": "shop001",
      "email": "customer@example.com",
      "cartid": "cart_20250115_001",
      "data": {
        "items": [...],
        "total": 598,
        "status": "active"
      },
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

## 5. List ข้อมูล (GET with Pagination)

```powershell
$body = @{
    collection = "carts"
    shopid = "shop001"
    email = "customer@example.com"
    cartid = "cart_20250115_001"
    limit = 10
    skip = 0
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/get" `
    -Method POST -Body $body -ContentType "application/json"
```

**หมายเหตุ:** เนื่องจากต้องระบุทั้ง 3 ตัว (shopid, email, cartid) ซึ่งเป็น unique combination 
จะได้ผลลัพธ์สูงสุด 1 รายการเท่านั้น

---

## ตัวอย่างข้อผิดพลาด

### Error 1: ไม่ระบุ shopid
```json
{
  "status": "error",
  "code": 400,
  "message": "shopid is required and cannot be empty"
}
```

### Error 2: ไม่ระบุ email
```json
{
  "status": "error",
  "code": 400,
  "message": "email is required and cannot be empty"
}
```

### Error 3: ไม่ระบุ cartid
```json
{
  "status": "error",
  "code": 400,
  "message": "cartid is required and cannot be empty"
}
```

---

## ตัวอย่างการใช้งานจริง

### Scenario 1: Shopping Cart

```powershell
# 1. สร้าง cart ใหม่
$createCart = @{
    collection = "carts"
    shopid = "shop001"
    email = "john@example.com"
    cartid = "cart_$(Get-Date -Format 'yyyyMMddHHmmss')"
    data = @{
        items = @(
            @{ sku = "SKU001"; name = "สินค้า A"; qty = 1; price = 500 }
        )
        total = 500
        status = "active"
        created_at = (Get-Date).ToString("o")
    }
    upsert = $true
} | ConvertTo-Json -Depth 10

$result = Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/update" `
    -Method POST -Body $createCart -ContentType "application/json"

Write-Host "Cart created: $($result | ConvertTo-Json)"

# 2. อัพเดท cart (เพิ่มสินค้า)
$cartId = "cart_20250115123456"  # ใช้ cartid จากขั้นตอนที่ 1

$updateCart = @{
    collection = "carts"
    shopid = "shop001"
    email = "john@example.com"
    cartid = $cartId
    data = @{
        items = @(
            @{ sku = "SKU001"; name = "สินค้า A"; qty = 1; price = 500 }
            @{ sku = "SKU002"; name = "สินค้า B"; qty = 2; price = 300 }
        )
        total = 1100
        status = "active"
    }
    upsert = $false
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/update" `
    -Method POST -Body $updateCart -ContentType "application/json"

# 3. ดึงข้อมูล cart
$getCart = @{
    collection = "carts"
    shopid = "shop001"
    email = "john@example.com"
    cartid = $cartId
} | ConvertTo-Json

$cart = Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/get" `
    -Method POST -Body $getCart -ContentType "application/json"

Write-Host "Cart data: $($cart.data | ConvertTo-Json -Depth 10)"

# 4. Checkout - ลบ cart
$deleteCart = @{
    collection = "carts"
    shopid = "shop001"
    email = "john@example.com"
    cartid = $cartId
    delete_many = $false
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/delete" `
    -Method POST -Body $deleteCart -ContentType "application/json"
```

### Scenario 2: User Session

```powershell
# สร้าง session
$sessionId = [guid]::NewGuid().ToString()

$session = @{
    collection = "sessions"
    shopid = "shop001"
    email = "user@example.com"
    cartid = $sessionId
    data = @{
        token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        login_at = (Get-Date).ToString("o")
        expires_at = (Get-Date).AddHours(24).ToString("o")
        device = "mobile"
        ip = "192.168.1.100"
    }
    upsert = $true
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/update" `
    -Method POST -Body $session -ContentType "application/json"

# ลบ session เมื่อ logout
$deleteSession = @{
    collection = "sessions"
    shopid = "shop001"
    email = "user@example.com"
    cartid = $sessionId
    delete_many = $false
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9090/goapi/atlas/delete" `
    -Method POST -Body $deleteSession -ContentType "application/json"
```

---

## สรุปสำคัญ

| การทำงาน | Endpoint | ต้องระบุ | upsert | delete_many |
|----------|----------|----------|--------|-------------|
| เพิ่มใหม่ | /atlas/update | shopid, email, cartid, data (ครบทั้ง 3) | true | - |
| แก้ไข | /atlas/update | shopid, email, cartid, data (ครบทั้ง 3) | false | - |
| ลบ 1 รายการ | /atlas/delete | shopid, email, cartid (ครบทั้ง 3) | - | false |
| ลบหลายรายการ | /atlas/delete | shopid, email, cartid (ครบทั้ง 3) | - | true |
| ค้นหา | /atlas/get | อย่างน้อย 1 ตัว (shopid หรือ email หรือ cartid) | - | - |
| List (Pagination) | /atlas/get | อย่างน้อย 1 ตัว + limit, skip | - | - |

**ข้อควรระวัง:**
- ทั้ง 3 ตัว (shopid, email, cartid) เป็น **unique key** ร่วมกัน
- **UPDATE และ DELETE:** ต้องระบุครบทั้ง 3 ตัวเสมอ และห้ามเป็นค่าว่าง
- **GET (Query):** ใช้ได้ 1-3 ตัว (flexible) แต่ต้องมีอย่างน้อย 1 ตัว
- `data` field สามารถเป็น JSON โครงสร้างใดก็ได้
