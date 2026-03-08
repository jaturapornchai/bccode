# Thai Word Tokenizer - ระบบตัดคำภาษาไทยสำหรับ Full Text Search

## จุดประสงค์
ใช้ตัดคำค้นหาภาษาไทยและภาษาอังกฤษเป็น Array เพื่อค้นหาสินค้าใน ClickHouse ได้แม่นยำขึ้น

## วิธีการทำงาน

### 1. Thai Word Segmentation
- **คำเต็ม**: เก็บคำทั้งหมดที่ user พิมพ์
- **คำย่อย 2 ตัวอักษร**: เช่น "น้ำตาล" → ["น้", "้ำ", "ำต", "ตา", "าล"]
- **คำย่อย 3 ตัวอักษร**: เช่น "น้ำตาล" → ["น้ำ", "้ำต", "ำตา", "ตาล"]
- **คำย่อย 4 ตัวอักษร**: เช่น "น้ำตาล" → ["น้ำต", "้ำตา", "ำตาล"]

### 2. Multiple Field Search
ค้นหาใน 5 fields:
- `name0` - ชื่อสินค้า
- `unitname` - ชื่อหน่วย (เช่น "ถุง", "กล่อง")
- `unitcode` - รหัสหน่วย
- `barcode` - บาร์โค้ด
- `itemcode` - รหัสสินค้า

### 3. Search Strategy
- **ตัวเลข**: ใช้ exact match (`field = 'value'`)
- **ภาษาไทย**: ใช้ LIKE (`lower(field) LIKE lower('%keyword%')`)
- **ภาษาอังกฤษ**: ใช้ LIKE (`lower(field) LIKE lower('%keyword%')`)
- **OR Operator**: รวมเงื่อนไขทุกคำด้วย OR เพื่อให้หาได้หลายคำ

## ตัวอย่างการใช้งาน

### ตัวอย่าง 1: ค้นหาด้วยชื่อสินค้าภาษาไทย
```dart
// Input
"น้ำตาลทราย"

// Tokenized keywords
["น้ำตาลทราย", "น้", "้ำ", "ำต", "ตา", "าล", "ลท", "ทร", "รา", "าย", 
 "น้ำ", "้ำต", "ำตา", "ตาล", "าลท", "ลทร", "ทรา", "ราย",
 "น้ำต", "้ำตา", "ำตาล", "าลทร", "ลทรา", "ทราย"]

// Generated SQL WHERE clause
(lower(name0) LIKE lower('%น้ำตาลทราย%') OR lower(unitname) LIKE lower('%น้ำตาลทราย%') OR ...) OR
(lower(name0) LIKE lower('%น้ำ%') OR lower(unitname) LIKE lower('%น้ำ%') OR ...) OR
(lower(name0) LIKE lower('%ตาล%') OR lower(unitname) LIKE lower('%ตาล%') OR ...) OR
...
```

### ตัวอย่าง 2: ค้นหาด้วยบาร์โค้ด
```dart
// Input
"8850123456789"

// Tokenized keywords (ตัวเลขเก็บทั้งหมด)
["8850123456789"]

// Generated SQL WHERE clause
(barcode = '8850123456789' OR itemcode = '8850123456789' OR 
 lower(name0) LIKE lower('%8850123456789%') OR ...)
```

### ตัวอย่าง 3: ค้นหาแบบผสม (ภาษาไทย + ภาษาอังกฤษ)
```dart
// Input
"น้ำอัดลม Coke"

// Tokenized keywords
["น้ำอัดลม", "coke", "น้", "้ำ", "ำอ", "อั", "ัด", "ดล", "ลม", 
 "น้ำ", "้ำอ", "ำอั", "อัด", "ัดล", "ดลม",
 "น้ำอ", "้ำอั", "ำอัด", "อัดล", "ัดลม"]

// Generated SQL WHERE clause
(lower(name0) LIKE lower('%น้ำอัดลม%') OR ...) OR
(lower(name0) LIKE lower('%coke%') OR ...) OR
(lower(name0) LIKE lower('%น้ำ%') OR ...) OR
...
```

### ตัวอย่าง 4: ค้นหาด้วยหน่วย
```dart
// Input
"กล่อง"

// Tokenized keywords
["กล่อง", "กล", "ล่", "่อ", "อง", "กล่", "ล่อ", "่อง", "กล่อ", "ล่อง"]

// Generated SQL WHERE clause
(lower(unitname) LIKE lower('%กล่อง%') OR ...) OR
(lower(unitname) LIKE lower('%กล่%') OR ...) OR
...
```

## ประโยชน์

### 1. ค้นหาได้แม่นยำขึ้น
- ค้นหาได้แม้พิมพ์ไม่ครบ (เช่น "น้ำตา" จะหา "น้ำตาลทราย" ได้)
- ค้นหาหลาย field พร้อมกัน (ชื่อ, หน่วย, บาร์โค้ด)

### 2. รองรับหลายภาษา
- ภาษาไทย (ตัดคำอัตโนมัติ)
- ภาษาอังกฤษ
- ตัวเลข (barcode/itemcode)

### 3. ยืดหยุ่น
- สามารถปรับ search fields ได้
- สามารถเปลี่ยน operator (OR/AND) ได้
- สามารถปรับขนาดคำย่อย (2-4 ตัวอักษร) ได้

## Performance Considerations

### ข้อดี
- ค้นหาได้ครอบคลุมมากขึ้น
- User experience ดีขึ้น (ไม่ต้องพิมพ์ให้ตรงทุกตัวอักษร)

### ข้อควรระวัง
- SQL query จะยาวขึ้นเมื่อมีหลายคำ
- ควรใช้ LIMIT เพื่อจำกัดผลลัพธ์
- ควรมี index ใน database สำหรับ fields ที่ค้นหา

## การปรับแต่ง

### เปลี่ยน Search Fields
```dart
final searchFields = ['name0', 'unitname', 'barcode']; // เลือกเฉพาะที่ต้องการ
```

### เปลี่ยน Operator
```dart
// OR - หาเจอคำใดคำหนึ่งก็แสดง
buildFullTextSearchQuery(keywords: keywords, operator: 'OR');

// AND - ต้องเจอทุกคำ
buildFullTextSearchQuery(keywords: keywords, operator: 'AND');
```

### ปรับ Token Length
แก้ใน `_simpleThaiSegment()`:
```dart
// ลด/เพิ่มขนาดคำย่อย
for (int i = 0; i <= thaiText.length - 2; i++) {
  segments.add(thaiText.substring(i, i + 2)); // 2 ตัวอักษร
}
```

## API Integration

ใช้ใน `ClickHouseProductService`:
```dart
// searchByName() - Full Text Search
final result = await productService.searchByName('น้ำตาล', 'SHOP001');

// universalSearch() - Auto detect + Full Text Search  
final result = await productService.universalSearch('8850123456789', 'SHOP001');
```

## Testing

ควรทดสอบกับ:
- ✅ คำภาษาไทยสั้น (1-2 คำ)
- ✅ คำภาษาไทยยาว (3+ คำ)
- ✅ ภาษาอังกฤษ
- ✅ ตัวเลข (barcode/itemcode)
- ✅ ผสมหลายภาษา
- ✅ อักขระพิเศษ (space, comma, etc.)
