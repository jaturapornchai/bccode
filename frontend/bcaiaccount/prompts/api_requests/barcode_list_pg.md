# API Request: Barcode List จาก PostgreSQL

## สถานะ: ✅ แก้ Backend แล้ว (AI แก้โดยตรง)

## สิ่งที่ต้องการ
เปลี่ยนหน้าจอรายการบาร์โค้ด (`product_barcode_screen.dart`) จากดึงข้อมูลผ่าน MongoDB API
ไปใช้ PostgreSQL เพื่อ:
- **เร็วกว่า** — PG มี index, ILIKE search
- **Query ค้นหาได้ดีกว่า** — filter หลาย field พร้อมกัน (กลุ่ม, ยี่ห้อ, หมวด, ช่วงราคา)
- **Search UI ใหม่** — มี advanced filter panel + font size (A- A+)

## Endpoint ที่สร้าง
- **Method:** POST
- **Path:** `/goapi/api/product/barcode/list`
- **Auth:** ไม่ต้อง (ใช้ shopid จาก body)

## Request Body
```json
{
  "shopid": "string — required, shop ID",
  "keyword": "string — ค้นหา barcode, name0, itemcode, groupnames (ILIKE)",
  "groupcode": "string — filter กลุ่มสินค้า",
  "brandcode": "string — filter ยี่ห้อ",
  "categorycode": "string — filter หมวดสินค้า",
  "classcode": "string — filter ระดับ",
  "designcode": "string — filter รูปทรง",
  "gradecode": "string — filter เกรด",
  "modelcode": "string — filter รุ่น",
  "patterncode": "string — filter รูปแบบ",
  "price_min": "number|null — ราคาต่ำสุด",
  "price_max": "number|null — ราคาสูงสุด",
  "limit": "int — จำนวนต่อหน้า (default 50, max 500)",
  "offset": "int — เริ่มจากรายการที่",
  "sort_field": "string — barcode|name0|itemcode|groupnames|price1|unitname (default barcode)",
  "sort_order": "string — asc|desc (default asc)"
}
```

## Response Format
```json
{
  "success": true,
  "data": [
    {
      "guidfixed": "xxx-xxx-xxx",
      "barcode": "8850001001001",
      "names": [{"code": "th", "name": "ปูนซีเมนต์ตราเสือ 50กก."}],
      "itemunitcode": "BAG",
      "itemunitnames": [{"code": "th", "name": "ถุง"}],
      "itemcode": "CMT001",
      "groupcode": "CONST",
      "groupnames": [{"code": "th", "name": "วัสดุก่อสร้าง"}],
      "prices": [{"keynumber": 1, "price": 180.0}],
      "imageuri": "https://..."
    }
  ],
  "total": 500
}
```

**หมายเหตุ:** Response format ตรงกับ `ProductBarcodeModel.fromJson()` ของ Flutter
— PG เก็บ flat string (name0) แต่ handler แปลงเป็น array format ให้

## ไฟล์ Backend ที่แก้ไข

### 1. `internal/goapi/process/build/create-database.go`
เพิ่ม columns ใน `productbarcode` table:
```sql
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS guidfixed TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS shopid TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS brandcode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS brandnames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS categorycode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS categorynames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS classcode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS classnames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS designcode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS designnames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS gradecode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS gradenames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS modelcode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS modelnames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS patterncode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS patternnames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS imageuri TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS isusesubbarcodes BOOLEAN DEFAULT FALSE;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubonecode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubonenames TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubtwocode TEXT;
ALTER TABLE productbarcode ADD COLUMN IF NOT EXISTS groupsubtwonames TEXT;
```

เพิ่ม indexes:
```sql
CREATE INDEX IF NOT EXISTS idx_productbarcode_shopid ON productbarcode (shopid);
CREATE INDEX IF NOT EXISTS idx_productbarcode_guidfixed ON productbarcode (guidfixed);
CREATE INDEX IF NOT EXISTS idx_productbarcode_groupcode ON productbarcode (groupcode);
CREATE INDEX IF NOT EXISTS idx_productbarcode_brandcode ON productbarcode (brandcode);
CREATE INDEX IF NOT EXISTS idx_productbarcode_categorycode ON productbarcode (categorycode);
```

### 2. `internal/goapi/models/mongo-product-model.go`
เพิ่ม fields ใน `MongoProductBarcodeModel`:
- BrandCode, BrandNames
- CategoryCode, CategoryNames
- ClassCode, ClassNames
- DesignCode, DesignNames
- GradeCode, GradeNames
- ModelCode, ModelNames
- PatternCode, PatternNames
- GroupSubOneCode, GroupSubOneNames
- GroupSubTwoCode, GroupSubTwoNames

### 3. `internal/goapi/mypg/product.go`
เขียนใหม่ `ProductBarcodeUpdate()` — sync 37 columns จาก MongoDB ลง PG:
- ใช้ checksum (MD5) เพื่อ skip ถ้าข้อมูลไม่เปลี่ยน
- DELETE + INSERT ใน transaction
- เพิ่ม helper: `firstLangName()`, `boolToStock()`

### 4. `internal/goapi/handlers/barcode_list.go` (ไฟล์ใหม่)
Handler สำหรับ `POST /api/product/barcode/list`:
- Dynamic SQL query builder
- ILIKE keyword search (barcode, name0, itemcode, groupnames)
- Exact match filters (groupcode, brandcode, categorycode, etc.)
- Price range filter
- Sort field whitelist (ป้องกัน SQL injection)
- Parameterized queries ($1, $2, ...)
- Response format แปลง flat PG data → array format ที่ Flutter เข้าใจ

### 5. `internal/goapi/bootstrap.go`
เพิ่ม route:
```go
g.POST("/api/product/barcode/list", handlers.BarcodeListHandler)
```

### 6. `assets/language/languages.json`
เพิ่ม language keys:
- `price_min` — ราคาต่ำสุด
- `price_max` — ราคาสูงสุด
- `clear_filter` — ล้างตัวกรอง

## ไฟล์ Frontend ที่แก้ไข

| ไฟล์ | การแก้ไข |
|------|---------|
| `lib/repositories/product_barcode_repository.dart` | เพิ่ม `searchBarcodeListPg()` ใช้ `goApiPost()` |
| `lib/bloc/product_barcode/product_barcode_event.dart` | เพิ่ม `ProductBarcodeLoadListPg` event |
| `lib/bloc/product_barcode/product_barcode_state.dart` | เพิ่ม `ProductBarcodeLoadPgSuccess/Failed` (มี total) |
| `lib/bloc/product_barcode/product_barcode_bloc.dart` | เพิ่ม `onProductBarcodeLoadListPg` handler |
| `lib/screens/config/product_barcode_screen.dart` | Search UI ใหม่ + filter panel + A- A+ |

## ขั้นตอนหลัง Deploy
1. **Build backend** → restart server
2. **Rebuild products** — เพื่อ sync ข้อมูลใหม่ลง PG columns ที่เพิ่ม
3. ทดสอบหน้าบาร์โค้ด → รายการต้องแสดงจาก PG
4. ทดสอบค้นหาภาษาไทย → ILIKE ทำงานถูกต้อง
5. ทดสอบ filter กลุ่ม/ยี่ห้อ/หมวด/ช่วงราคา
6. ทดสอบ A- A+ → font size เปลี่ยน
7. ทดสอบ infinite scroll → โหลดเพิ่มได้
8. กดเลือก row → เปิด edit ได้ปกติ (ยังใช้ MongoDB)

## SQL Query ที่ใช้ (Handler)
```sql
-- Count
SELECT COUNT(*) FROM productbarcode WHERE ...

-- Data
SELECT COALESCE(guidfixed,''), barcode, COALESCE(name0,''), COALESCE(unitcode,''),
       COALESCE(unitname,''), COALESCE(itemcode,''), COALESCE(groupcode,''),
       COALESCE(groupnames,''), COALESCE(price1,0), COALESCE(imageuri,'')
FROM productbarcode
WHERE (barcode ILIKE $1 OR name0 ILIKE $1 OR itemcode ILIKE $1 OR groupnames ILIKE $1)
  AND groupcode = $2
  AND brandcode = $3
  AND price1 >= $4 AND price1 <= $5
ORDER BY barcode ASC
LIMIT $6 OFFSET $7
```
