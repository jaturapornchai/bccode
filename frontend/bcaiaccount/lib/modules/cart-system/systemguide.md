# Cart System - Guideline และสิ่งสำคัญ

## 📋 สารบัญ

1. [ภาพรวมระบบ](#ภาพรวมระบบ)
2. [โครงสร้างไฟล์](#โครงสร้างไฟล์)
3. [Data Flow](#data-flow)
4. [ขั้นตอนการใช้งาน](#ขั้นตอนการใช้งาน)
5. [Business Rules](#business-rules)
6. [API และ Database](#api-และ-database)
7. [การพัฒนาและแก้ไข](#การพัฒนาและแก้ไข)
8. [Troubleshooting](#troubleshooting)

---

## 📌 ภาพรวมระบบ

### วัตถุประสงค์
ระบบตะกร้าสินค้าแบบ Multi-Cart ที่รองรับ 3 ประเภทธุรกรรม:
- **ระบบขาย (Sales)** - สำหรับขายสินค้าให้ลูกค้า/ลูกหนี้
- **ระบบซื้อ (Purchase)** - สำหรับสั่งซื้อสินค้าจากเจ้าหนี้/ซัพพลายเออร์
- **ระบบคงคลัง (Inventory)** - สำหรับจัดการสินค้าคงคลัง (เบิก/รับ/โอน)

### คุณสมบัติหลัก
✅ แต่ละระบบมีตะกร้าได้**หลายตะกร้า** พร้อมกัน
✅ แต่ละตะกร้ามี **GUID (cartId)** ไม่เปลี่ยนแปลง (เปลี่ยนชื่อได้)
✅ **Persistent Storage** - เก็บใน MongoDB ไม่หายเมื่อปิดแอป
✅ **Active Cart** - แต่ละระบบมี active cart ที่กำลังใช้งาน
✅ **ค้นหาสินค้า** - Full-text search จาก ClickHouse (300 รายการ)
✅ **Thai Word Tokenizer** - ตัดคำภาษาไทยเพื่อค้นหาที่แม่นยำ
✅ **เลือกคลัง/ที่เก็บ** - แต่ละรายการสินค้ามีข้อมูลคลัง/ที่เก็บ
✅ **เลือกลูกหนี้/เจ้าหนี้** - สำหรับระบบขาย/ซื้อ (optional)

---

## 📂 โครงสร้างไฟล์

```
lib/modules/cart-system/
├── cartsystem.dart                    # หน้าจอหลัก (3 Tabs)
├── cartsystem_old_backup.dart         # Backup code เก่า (อ้างอิง)
│
├── bloc/
│   ├── cart_cubit.dart                # State management สำหรับตะกร้า
│   ├── cart_state.dart                # States ของตะกร้า
│   ├── product_search_cubit.dart      # State management สำหรับค้นหาสินค้า
│   └── product_search_state.dart      # States ของการค้นหา
│
├── models/
│   ├── cart_model.dart                # โมเดลตะกร้า (CartModel)
│   ├── cart_item_model.dart           # โมเดลสินค้าในตะกร้า
│   ├── cart_system_type.dart          # Enum ประเภทระบบ (sales/purchase/inventory)
│   ├── debtor_model.dart              # โมเดลลูกหนี้
│   ├── creditor_model.dart            # โมเดลเจ้าหนี้
│   ├── clickhouse_warehouse_model.dart # โมเดลคลังสินค้า
│   ├── clickhouse_location_model.dart  # โมเดลที่เก็บสินค้า
│   ├── product_search_model.dart      # โมเดลสินค้าจากการค้นหา
│   ├── product_transaction_model.dart # โมเดลธุรกรรมสินค้า
│   └── product_card_data.dart         # ข้อมูลแสดงบนการ์ดสินค้า
│
├── services/
│   ├── mongodb_cart_service.dart      # Service สำหรับ CRUD ตะกร้าใน MongoDB
│   ├── clickhouse_product_service.dart # Service ค้นหาสินค้าจาก ClickHouse
│   ├── clickhouse_warehouse_service.dart # Service ดึงข้อมูลคลัง/ที่เก็บ
│   ├── clickhouse_debtor_creditor_service.dart # Service ดึงข้อมูลลูกหนี้/เจ้าหนี้
│   ├── thai_word_tokenizer.dart       # ตัดคำภาษาไทย
│   └── transliteration_service.dart   # แปลงภาษาไทย → English
│
├── widgets/
│   ├── product_search_widget.dart     # กล่องค้นหาและแสดงสินค้า
│   ├── product_detail_dialog.dart     # Dialog แสดงรายละเอียดสินค้า
│   ├── cart_list_widget.dart          # แสดงรายการตะกร้าทั้งหมด
│   ├── cart_detail_screen.dart        # หน้าแสดงสินค้าในตะกร้า
│   ├── cart_item_widget.dart          # แสดง 1 รายการสินค้า
│   ├── debtor_selector_dialog.dart    # Dialog เลือกลูกหนี้
│   ├── creditor_selector_dialog.dart  # Dialog เลือกเจ้าหนี้
│   └── warehouse_selector_dialog.dart # Dialog เลือกคลัง/ที่เก็บ
│
└── docs/
    ├── cartsystem.md                  # เอกสารหลักของระบบ
    ├── mongodbapi.md                  # API Spec ของ MongoDB
    ├── clickhouseapi.md               # API Spec ของ ClickHouse
    ├── thai_word_tokenizer.md         # วิธีใช้ Thai Word Tokenizer
    └── transliteration_flutter_guide.md # คู่มือ Transliteration
```

---

## 🔄 Data Flow

### 1. การโหลดตะกร้าเมื่อเปิดแอป

```
┌─────────────────┐
│  initState()    │
└────────┬────────┘
         │
         v
┌─────────────────────────────┐
│ _initializeUserData()       │ ← ดึง shopId, userEmail
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│ _loadAllCarts()             │
└────────┬────────────────────┘
         │
         ├───> MongoDBCartService.getCartsByType(sales)
         ├───> MongoDBCartService.getCartsByType(purchase)
         └───> MongoDBCartService.getCartsByType(inventory)
                │
                v
         ┌─────────────────────┐
         │ Filter: status='active' │
         └─────────┬───────────┘
                   │
                   v
         ┌─────────────────────────┐
         │ setState():             │
         │  - _salesCarts          │
         │  - _purchaseCarts       │
         │  - _inventoryCarts      │
         │  - _activeSalesCart     │
         │  - _activePurchaseCart  │
         │  - _activeInventoryCart │
         └─────────────────────────┘
```

### 2. การสร้างตะกร้าใหม่

```
┌──────────────────────┐
│ User กดปุ่ม "สร้างตะกร้า" │
└──────────┬───────────┘
           │
           v
┌──────────────────────┐
│ _createNewCart()     │
└──────────┬───────────┘
           │
           v
┌───────────────────────────┐
│ สร้าง CartModel ใหม่       │
│  - cartId: UUID v4 (GUID) │
│  - cartName: auto-generate │
│  - systemType: sales/...  │
│  - status: 'active'       │
│  - items: []              │
│  - debtorId (optional)    │
│  - creditorId (optional)  │
│  - warehouseId (optional) │
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ MongoDBCartService.createCart() │
│  POST /atlas/update       │
│  {                        │
│    collection: 'carts',   │
│    shopid: ...,           │
│    email: ...,            │
│    cartid: ...,           │
│    data: {...},           │
│    upsert: true           │
│  }                        │
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ setState():               │
│  - เพิ่มเข้า _salesCarts  │
│  - ตั้งเป็น _activeSalesCart │
└───────────────────────────┘
```

### 3. การเพิ่มสินค้าลงตะกร้า

```
┌──────────────────────┐
│ User เลือกสินค้า       │
└──────────┬───────────┘
           │
           v
┌──────────────────────┐
│ _addProductToCart()  │
└──────────┬───────────┘
           │
           v
┌───────────────────────────┐
│ เปิด ProductDetailDialog  │
│  - แสดงรายละเอียดสินค้า    │
│  - ระบุจำนวน              │
│  - เลือกคลัง/ที่เก็บ      │
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ สร้าง CartItemModel       │
│  - itemId: UUID           │
│  - itemCode, barcode      │
│  - productName            │
│  - quantity, price        │
│  - warehouseId, locationId│
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ cart.addItem(newItem)     │
│  → CartModel ใหม่         │
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ MongoDBCartService.updateCart() │
│  POST /atlas/update       │
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ _refreshCart()            │
│  → อัพเดท UI             │
└───────────────────────────┘
```

### 4. การดูสินค้าในตะกร้า

```
┌──────────────────────┐
│ User กดปุ่ม "ดูสินค้า" │
└──────────┬───────────┘
           │
           v
┌──────────────────────┐
│ _openCartDetail()    │
└──────────┬───────────┘
           │
           v
┌───────────────────────────┐
│ Navigator.push()          │
│  → CartDetailScreen       │
└───────────┬───────────────┘
            │
            v
┌───────────────────────────┐
│ CartDetailScreen          │
│  - แสดงรายการสินค้าทั้งหมด │
│  - แก้ไขจำนวน              │
│  - ลบรายการ                │
│  - แก้ไขราคา               │
│  - ดูยอดรวม                │
└───────────┬───────────────┘
            │
            v (onCartUpdated)
┌───────────────────────────┐
│ _refreshCart()            │
│  → อัพเดท active cart     │
└───────────────────────────┘
```

---

## 🚀 ขั้นตอนการใช้งาน

### สำหรับ End User

#### 1. เปิดระบบตะกร้า
```
Main Menu → ระบบตะกร้า (Cart System)
```


#### 3. สร้างตะกร้าใหม่
```
กดปุ่ม "สร้างตะกร้า" (สีเขียว)
→ ระบบจะสร้างตะกร้าใหม่อัตโนมัติ
→ ตั้งชื่อแบบ "ตะกร้าขาย #001" (auto)
```

กดที่ตระกร้าเพื่อเลือกใช้งาน เมื่อกดให้ Show Dialog 
ปุ่ม ยืนยันการเปลี่ยนตะกร้า เมื่อกดก็ให้เปลลี่ยนตะกร้าตามที่เลือก
ปุ่ม แก้ไขชื่อตะกร้า ประเภะบบให้ Show Dialog
โดยสามารถเปลี่ยนชื่อตะกร้าได้ใน Dialog นี้ และเลือกประเภทตะกร้า ได้เช่นกัน
เงื่อนการเลือกตระกร้า:
ถ้าอยู่ Mode ขาย สามารถเลือกประเภทได้ คือ ขาย / สั่งขาย / เสนอราคา
ถ้าอยู่ Mode ซื้อ สามารถเลือกประเภทได้ คือ สั่งซื้อ
ถ้าอยู่ Mode คงคลัง สามารถเลือกประเภทได้ คือ เบิก / โอน
ให้เป็น Radio Button
มีปุ่ม ยืนยันการเปลี่ยนแปลง ยกเลิก

#### 4. เลือกตะกร้า (ถ้ามีหลายตะกร้า)
```
กดที่ Chip ชื่อตะกร้า (สีฟ้าอ่อน)
→ ตะกร้าที่เลือกจะเป็นสีเข้ม + ตัวหนา
```

#### 5. เลือกลูกหนี้/เจ้าหนี้ (ถ้าต้องการ)
```
Tab ขาย → กดปุ่ม "ลูกหนี้: ยังไม่เลือก"
Tab ซื้อ → กดปุ่ม "เจ้าหนี้: ยังไม่เลือก"
→ เลือกจากรายการ
```

#### 6. เลือกคลัง/ที่เก็บ (ถ้าต้องการ)
```
กดปุ่ม "คลัง: คลังหลัก"
→ เลือกคลังและที่เก็บ
→ สินค้าที่เพิ่มต่อไปจะใช้คลัง/ที่เก็บนี้เป็น default
```

#### 7. ค้นหาและเพิ่มสินค้า
```
พิมพ์ชื่อสินค้า, บาร์โค้ด, หรือรหัสสินค้า
→ เลือกสินค้าที่ต้องการ
→ ระบุจำนวน
→ ยืนยัน "เพิ่มลงตะกร้า"
```

#### 8. ดูรายการสินค้าในตะกร้า
```
กดปุ่ม "ดูสินค้า (5)" (สีน้ำเงิน)
→ แสดงรายการสินค้าทั้งหมด
→ สามารถแก้ไข/ลบรายการได้
```

#### 9. ปิดแอปและกลับมาใช้งาน
```
ตะกร้าทั้งหมดจะยังอยู่
→ ไม่หายเมื่อปิดแอป
→ สามารถทำงานต่อได้เลย
```

---

## ⚠️ Business Rules

### ❌ ห้ามแก้ไข (จากไฟล์ .claude/CLAUDE.md)

1. **โครงสร้างข้อมูล**
   - ห้ามแก้ไข CartModel schema
   - ห้ามแก้ไข CartItemModel schema
   - ห้ามแก้ไข MongoDB collection structure
   - ห้ามแก้ไข ClickHouse query structure

2. **Business Logic**
   - ห้ามแก้ไขการคำนวณราคา/ยอดรวม
   - ห้ามแก้ไขการ filter สถานะตะกร้า
   - ห้ามแก้ไขการสร้าง GUID (cartId)
   - ห้ามแก้ไขการจัดเก็บข้อมูลลง MongoDB

3. **API Calls**
   - ห้ามแก้ไข endpoint URLs
   - ห้ามแก้ไข request/response format
   - ห้ามแก้ไข authentication headers

### ✅ แก้ไขได้

1. **UI/UX**
   - สี, ขนาดตัวอักษร, spacing
   - ตำแหน่งปุ่ม, layout
   - Animation, transition
   - Icon, รูปภาพ

2. **Formatting**
   - รูปแบบการแสดงตัวเลข
   - รูปแบบวันที่/เวลา
   - การเรียงลำดับแสดงผล

3. **Helper Functions**
   - ฟังก์ชันช่วยจัดรูปแบบข้อความ
   - ฟังก์ชันช่วยแสดงผล
   - Validation UI (ไม่ใช่ server-side)

---

## 🗄️ API และ Database

### MongoDB Atlas API

**Base URL**: `{global.goApiUrlPath('atlas')}`

#### 1. GET Carts (โหลดตะกร้าทั้งหมด)
```http
POST /get
Content-Type: application/json

{
  "collection": "carts",
  "shopid": "xxx",
  "email": "user@example.com"
}

Response:
{
  "status": "success",
  "code": 200,
  "count": 3,
  "data": [
    {
      "shopId": "xxx",
      "email": "user@example.com",
      "cartId": "uuid-v4",
      "cartName": "ตะกร้าขาย #001",
      "systemType": "sales",
      "status": "active",
      "items": [...],
      "totalAmount": 1500.00,
      "createdAt": "2025-01-15T10:30:00Z",
      "updatedAt": "2025-01-15T12:00:00Z",
      "debtorId": "debtor-guid",
      "debtorCode": "D001",
      "debtorName": "บริษัท ABC",
      "warehouseId": "warehouse-id",
      "warehouseName": "คลังหลัก",
      "locationId": "location-id",
      "locationName": "ชั้น 1"
    }
  ]
}
```

#### 2. CREATE/UPDATE Cart
```http
POST /update
Content-Type: application/json

{
  "collection": "carts",
  "shopid": "xxx",
  "email": "user@example.com",
  "cartid": "uuid-v4",
  "data": {
    "shopId": "xxx",
    "email": "user@example.com",
    "cartId": "uuid-v4",
    "cartName": "ตะกร้าขาย #001",
    ...
  },
  "upsert": true
}

Response:
{
  "status": "success",
  "code": 200,
  "message": "Cart saved successfully"
}
```

#### 3. DELETE Cart
```http
POST /delete
Content-Type: application/json

{
  "collection": "carts",
  "shopid": "xxx",
  "email": "user@example.com",
  "cartid": "uuid-v4"
}
```

### ClickHouse API

**Base URL**: `{global.goApiUrlPath('clickhouse')}`

#### 1. Search Products
```http
POST /v2/product/search
Content-Type: application/json

{
  "shopId": "xxx",
  "searchText": "สินค้า",
  "limit": 300
}

Response:
{
  "products": [
    {
      "itemcode": "P001",
      "barcode": "1234567890123",
      "name0": "สินค้าทดสอบ",
      "unitcode": "PCS",
      "unitname": "ชิ้น",
      "price": 100.00,
      "stock": 50
    }
  ]
}
```

#### 2. Get Warehouses
```http
POST /v2/warehouse/list
Content-Type: application/json

{
  "shopId": "xxx"
}
```

#### 3. Get Locations
```http
POST /v2/location/list
Content-Type: application/json

{
  "shopId": "xxx",
  "warehouseId": "warehouse-id"
}
```

---

## 🛠️ การพัฒนาและแก้ไข

### ขั้นตอนการเพิ่มฟีเจอร์ใหม่

#### 1. อ่านและทำความเข้าใจ
```bash
# อ่านไฟล์เอกสารทั้งหมด
docs/cartsystem.md          # เอกสารหลัก
docs/mongodbapi.md          # API Spec
x.md                        # ไฟล์นี้
```

#### 2. ตรวจสอบกฏใน CLAUDE.md
```bash
# อ่านไฟล์กฏ
.claude/CLAUDE.md

# ตรวจสอบว่าสิ่งที่จะทำ:
- ✅ เป็นการแก้ UI/Formatting หรือไม่?
- ❌ เป็นการแก้ Business Logic หรือไม่?
```

#### 3. ทำการแก้ไข

**ตัวอย่าง: เพิ่มปุ่มใหม่**
```dart
// ใน _buildCompactPanel()
ElevatedButton.icon(
  onPressed: () {
    // เรียกฟังก์ชันที่มีอยู่แล้ว
    _someExistingFunction();
  },
  icon: const Icon(Icons.new_icon, size: 18),
  label: const Text('ปุ่มใหม่', style: TextStyle(fontSize: 13)),
  style: ElevatedButton.styleFrom(
    backgroundColor: Colors.purple[700],
    foregroundColor: Colors.white,
    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
    minimumSize: const Size(0, 36),
  ),
)
```

**ตัวอย่าง: แก้ไขสี/ขนาดตัวอักษร**
```dart
// เดิม
Text('ชื่อตะกร้า', style: TextStyle(fontSize: 13))

// ใหม่
Text('ชื่อตะกร้า', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500))
```

#### 4. ทดสอบ
```bash
# รัน app
flutter run

# ทดสอบทุก Tab
- ระบบขาย
- ระบบซื้อ
- ระบบคงคลัง

# ทดสอบทุกฟีเจอร์
- สร้างตะกร้า
- เลือกตะกร้า
- เพิ่มสินค้า
- ดูสินค้าในตะกร้า
- แก้ไข/ลบสินค้า
- ปิดแอป → เปิดใหม่ → ตรวจสอบว่าข้อมูลยังอยู่
```

### Code Conventions

#### ชื่อตัวแปร (ภาษาไทย สำหรับ UI/Formatting)
```dart
// ✅ ถูกต้อง
String ชื่อตะกร้า = cart.cartName;
int จำนวนสินค้า = cart.itemCount;

// ❌ ผิด (ใช้ภาษาอังกฤษสำหรับ business logic)
String cartName = cart.cartName; // OK for business logic
```

#### Comments
```dart
// เพิ่มคอมเมนต์อธิบายโค้ดใหม่
/// ฟังก์ชันนี้ใช้สำหรับ...
/// - ทำอะไร
/// - ทำไม
/// - ใช้งานอย่างไร
void _newFunction() {
  // Implementation
}
```

#### Helper Functions
```dart
// สร้าง helper function เพื่อให้โค้ดหลักอ่านง่าย
String _จัดรูปแบบราคา(double ราคา) {
  return global.formatNumber(ราคา, 2);
}

Widget _สร้างปุ่มสีน้ำเงิน(String ข้อความ, VoidCallback onPressed) {
  return ElevatedButton(
    onPressed: onPressed,
    child: Text(ข้อความ),
    style: ElevatedButton.styleFrom(
      backgroundColor: Colors.blue[700],
    ),
  );
}
```

---

## 🔧 Troubleshooting

### ปัญหา: ตะกร้าหายเมื่อเปิดแอปใหม่

**สาเหตุ**:
1. `_shopId` หรือ `_userEmail` เป็น null
2. MongoDB API ไม่ตอบกลับ
3. Filter สถานะผิด (ไม่ใช่ 'active')

**วิธีแก้**:
```dart
// ตรวจสอบ log
AppLogger.info('🔄 กำลังโหลดตะกร้า... shopId: $_shopId, email: $_userEmail');
AppLogger.info('📦 พบตะกร้า: ขาย ${salesCarts.length}, ...');

// ตรวจสอบ MongoDB response
// ดูใน DevTools Console
```

### ปัญหา: เพิ่มสินค้าแล้วไม่แสดง

**สาเหตุ**:
1. ไม่ได้เรียก `_refreshCart()` หลังเพิ่มสินค้า
2. MongoDB update ล้มเหลว
3. UI ไม่ได้ rebuild

**วิธีแก้**:
```dart
// หลังเพิ่มสินค้า ต้องเรียก
await _cartService.updateCart(updatedCart);
await _refreshCart(cart); // ← สำคัญ!
```

### ปัญหา: ค้นหาสินค้าไม่เจอ

**สาเหตุ**:
1. ClickHouse API ไม่ตอบกลับ
2. Thai Word Tokenizer ตัดคำผิด
3. Limit 300 รายการ ไม่พอ

**วิธีแก้**:
```dart
// ตรวจสอบ search query
AppLogger.info('🔍 ค้นหา: $searchText');
AppLogger.info('📦 พบสินค้า: ${products.length} รายการ');

// ลอง search ด้วย barcode หรือ itemcode แทน
```

### ปัญหา: สีหรือ layout ไม่ถูกต้อง

**วิธีแก้**:
```dart
// ใช้ Flutter DevTools
// ดู Widget Inspector
// ตรวจสอบ padding, margin, colors

// ตัวอย่าง debug
Container(
  color: Colors.red, // เพิ่ม border เพื่อดู boundary
  child: YourWidget(),
)
```

---

## 📚 เอกสารเพิ่มเติม

- [MongoDB Atlas API](docs/mongodbapi.md)
- [ClickHouse API](docs/clickhouseapi.md)
- [Thai Word Tokenizer](docs/thai_word_tokenizer.md)
- [Transliteration Guide](docs/transliteration_flutter_guide.md)
- [Cart System Main Doc](docs/cartsystem.md)

---

## ✅ Checklist สำหรับ Developer

### ก่อนเริ่มงาน
- [ ] อ่าน x.md (ไฟล์นี้)
- [ ] อ่าน .claude/CLAUDE.md
- [ ] อ่าน docs/cartsystem.md
- [ ] ทำความเข้าใจ Business Rules

### ขณะทำงาน
- [ ] ตรวจสอบว่าไม่แก้ Business Logic
- [ ] เพิ่ม comments อธิบายโค้ดใหม่
- [ ] ใช้ชื่อตัวแปรภาษาไทยสำหรับ UI
- [ ] สร้าง helper functions ถ้าจำเป็น

### หลังทำงาน
- [ ] ทดสอบทุก Tab (ขาย/ซื้อ/คงคลัง)
- [ ] ทดสอบสร้าง/เลือก/ลบตะกร้า
- [ ] ทดสอบเพิ่ม/แก้ไข/ลบสินค้า
- [ ] ทดสอบปิด → เปิดแอป (persistence)
- [ ] ตรวจสอบ log ไม่มี error
- [ ] อัพเดทเอกสาร (ถ้าจำเป็น)

---

**เอกสารนี้อัพเดทล่าสุด**: 2025-01-15
**เวอร์ชัน**: 1.0.0
**ผู้เขียน**: Claude Code Assistant
