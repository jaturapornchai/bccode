# Product Models

สินค้า, Barcode, หน่วยนับ, BOM, กลุ่มสินค้า

## Product (สินค้า)

**Backend:** `internal/product/product/models/product.go`
**Frontend:** `lib/model/product_model.dart` → `ProductMasterModel`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *PartitionIdentity | — | — | inline | ParID (ซ่อน) |
| Code | string | `code` | `code` | รหัสสินค้า |
| Names | *[]NameX | `names` | `names` | ชื่อสินค้า (หลายภาษา) |
| GroupCode | string | `groupcode` | `groupcode` | รหัสกลุ่มสินค้า |
| GroupNames | *[]NameX | `groupnames` | `groupnames` | ชื่อกลุ่ม |
| ManufacturerGUID | string | `manufacturerguid` | `manufacturerguid` | GUID ผู้ผลิต |
| ManufacturerCode | string | `manufacturercode` | `manufacturercode` | รหัสผู้ผลิต |
| ManufacturerNames | *[]NameX | `manufacturernames` | `manufacturernames` | ชื่อผู้ผลิต |
| Dimensions | []ProductDimension | `dimensions` | `dimensions` | มิติสินค้า (สี, ขนาด, etc.) |
| VatType | int8 | `vattype` | `vattype` | ประเภท VAT (ดู enums.md) |
| ItemType | int8 | `itemtype` | `itemtype` | 0=ทั่วไป, 1=อาหาร, 2=เครื่องดื่ม |
| UnitGuid | string | `unitguid` | `unitguid` | GUID หน่วยนับ |
| Barcodes | []Barcodes | `barcodes` | — | รายการ barcode |

### Barcodes (barcode ต่อหน่วย)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| GuidFixed | string | `guidfixed` | GUID |
| ItemUnitCode | string | `itemunitcode` | รหัสหน่วย |
| ItemUnitNames | *[]NameX | `itemunitnames` | ชื่อหน่วย |
| Barcode | string | `barcode` | เลข barcode |
| Prices | *[]ProductPrice | `prices` | ราคาตามระดับ |
| Condition | bool | `condition` | สถานะ |

### ProductPrice (ราคาตามระดับ)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| KeyNumber | int | `keynumber` | ระดับราคา (1-9) |
| Price | float64 | `price` | ราคา |

### ProductDimension (มิติสินค้า)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| Code | string | `code` | รหัสมิติ |
| Names | *[]NameX | `names` | ชื่อมิติ (สี, ขนาด, etc.) |
| Choices | []DimensionChoice | `choices` | ตัวเลือก |

---

## ProductBarcode (MongoDB Collection)

**Backend:** `internal/product/productbarcode/models/productbarcode.go`
**MongoDB Collection:** `productBarcode`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| ItemCode | string | `itemcode` | `itemcode` | รหัสสินค้า |
| Barcode | string | `barcode` | `barcode` | barcode |
| Names | *[]NameX | `names` | `names` | ชื่อสินค้า (หลายภาษา) |
| UnitCode | string | `unitcode` | `unitcode` | รหัสหน่วย |
| UnitNames | *[]NameX | `unitnames` | `unitnames` | ชื่อหน่วย |
| StandValue | float64 | `standvalue` | `standvalue` | ตัวตั้ง (unit conversion) |
| DivideValue | float64 | `dividevalue` | `dividevalue` | ตัวหาร (unit conversion) |
| Price | float64 | `price` | `price` | ราคาขาย |
| AverageCost | float64 | `averagecost` | `averagecost` | ต้นทุนเฉลี่ย |
| GroupCode | string | `groupcode` | `groupcode` | รหัสกลุ่มสินค้า |
| GroupNames | *[]NameX | `groupnames` | `groupnames` | ชื่อกลุ่ม |
| CategoryCode | string | `categorycode` | `categorycode` | รหัสหมวดหมู่ |
| CategoryNames | *[]NameX | `categorynames` | `categorynames` | ชื่อหมวดหมู่ |
| BrandCode | string | `brandcode` | `brandcode` | รหัสยี่ห้อ |
| BrandNames | *[]NameX | `brandnames` | `brandnames` | ชื่อยี่ห้อ |
| ItemType | int8 | `itemtype` | `itemtype` | ประเภทสินค้า |
| VatType | int8 | `vattype` | `vattype` | ประเภท VAT |
| ImageUri | string | `imageuri` | `imageuri` | URL รูป |
| Prices | *[]ProductPrice | `prices` | `prices` | ราคาตามระดับ |
| IsSumPoint | bool | `issumpoint` | `issumpoint` | สะสมแต้ม |

### Unit Conversion

สินค้า 1 ตัวมีหลายหน่วย เชื่อมกันด้วย `standvalue` / `dividevalue`:

```
ตัวอย่าง: น้ำขวดเล็ก (ชิ้น) → แพ็ค 12 ชิ้น → ลัง 4 แพ็ค

barcode: 8850001001 (ชิ้น)  → stand=1, divide=1
barcode: 8850001012 (แพ็ค)  → stand=12, divide=1
barcode: 8850001048 (ลัง)   → stand=48, divide=1

จำนวนหน่วยเล็ก = qty × standvalue / dividevalue
```

---

## ProductBarcode (ClickHouse Table)

**Table:** `productbarcode` (ReplacingMergeTree)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน |
| itemcode | String | รหัสสินค้า |
| barcode | String | barcode |
| name0-5 | String | ชื่อภาษาที่ 0-5 |
| checksum | String | checksum |
| groupcode | String | รหัสกลุ่ม |
| groupnames | String | ชื่อกลุ่ม (JSON string) |
| unitcode | String | รหัสหน่วย |
| unitname | String | ชื่อหน่วย |
| price1 | Float64 | ราคาขาย |
| unitstand | Float64 | ตัวตั้ง (default 1) |
| unitdivide | Float64 | ตัวหาร (default 1) |
| barcoderef | String | barcode อ้างอิง |
| imageuri | String | URL รูป |
| price_retail | Float64 | ราคาปลีก |

**Engine:** ReplacingMergeTree, Partition by shopid, Order by (shopid, barcode)

---

## Product (Frontend — Flutter)

**Frontend:** `lib/model/product_model.dart` → `ProductModel`

| Field | Dart Type | JSON Key | หมายเหตุ |
|-------|-----------|----------|----------|
| guidfixed | String | `guidfixed` | GUID |
| itemcode | String | `itemcode` | รหัสสินค้า |
| groupcode | String | `groupcode` | รหัสกลุ่ม |
| groupnames | List\<LanguageDataModel\> | `groupnames` | ชื่อกลุ่ม |
| barcodes | List\<String\> | `barcodes` | barcode ทั้งหมด |
| names | List\<LanguageDataModel\> | `names` | ชื่อสินค้า |
| multiunit | bool | `multiunit` | หลายหน่วย |
| useserialnumber | bool | `useserialnumber` | ใช้ serial number |
| units | List\<ProductUnitModel\> | `units` | รายการหน่วย |
| images | List\<ImagesModel\> | `images` | รูปภาพ |
| unitcost | String | `unitcost` | หน่วยต้นทุน |
| unitstandard | String | `unitstandard` | หน่วยมาตรฐาน |
| itemstocktype | int | `itemstocktype` | 0=stock, 1=service |
| itemtype | int | `itemtype` | 0=ทั่วไป, 1=อาหาร, 2=เครื่องดื่ม |
| vattype | int | `vattype` | ประเภท VAT |
| issumpoint | bool | `issumpoint` | สะสมแต้ม |

### ProductUnitModel (หน่วยนับ — Frontend)

| Field | Dart Type | JSON Key | หมายเหตุ |
|-------|-----------|----------|----------|
| unitcode | String | `unitcode` | รหัสหน่วย |
| names | List\<LanguageDataModel\> | `names` | ชื่อหน่วย |
| divider | double | `divider` | ตัวหาร = dividevalue |
| stand | double | `stand` | ตัวตั้ง = standvalue |
| xorder | int | `xorder` | ลำดับ |
| stockcount | bool | `stockcount` | นับ stock |

---

## BOM (Bill of Materials)

**Backend:** `internal/product/productbom/models/productbom.go`

### ProductBOM

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัส BOM |
| Names | *[]NameX | `names` | `names` | ชื่อ BOM |
| BOMType | int8 | `bomtype` | `bomtype` | ประเภท |
| ProductCode | string | `productcode` | `productcode` | รหัสสินค้าผลิต |
| ProductBarcode | string | `productbarcode` | `productbarcode` | barcode สินค้าผลิต |
| Qty | float64 | `qty` | `qty` | จำนวนที่ผลิตได้ |
| UnitCode | string | `unitcode` | `unitcode` | หน่วย |
| Materials | []BOMMaterial | `materials` | `materials` | วัตถุดิบ |

### BOMMaterial (วัตถุดิบ)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| ItemCode | string | `itemcode` | รหัสวัตถุดิบ |
| Barcode | string | `barcode` | barcode |
| ItemNames | *[]NameX | `itemnames` | ชื่อ |
| UnitCode | string | `unitcode` | หน่วย |
| Qty | float64 | `qty` | จำนวนที่ใช้ |
| WhCode | string | `whcode` | คลัง |
| LocationCode | string | `locationcode` | ตำแหน่ง |

---

## ProductGroup, Category, Brand (กลุ่มสินค้า)

ทุกตัวใช้โครงสร้างเดียวกัน:

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัส |
| Names | *[]NameX | `names` | `names` | ชื่อ (หลายภาษา) |

**Backend locations:**
- Product Group: `internal/product/productgroup/`
- Category: `internal/product/productcategory/`
- Brand: `internal/product/productbrand/`
- Manufacturer: `internal/product/manufacturer/`

---

## Unit (หน่วยนับ)

**Backend:** `internal/product/unit/models/unit.go`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัสหน่วย |
| Names | *[]NameX | `names` | `names` | ชื่อหน่วย (หลายภาษา) |
