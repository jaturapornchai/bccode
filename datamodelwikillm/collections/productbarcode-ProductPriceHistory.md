---
source: backend/internal/product/productbarcode/models/product_price_history.go
collection: productbarcodespricehistory
tags: [datamodel, mongodb, mongo-root]
---

# ProductPriceHistory

โครงสร้าง `ProductPriceHistory` เป็น root document ที่ repository โมดูล `productbarcode` ใช้กับ MongoDB collection `productbarcodespricehistory`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeGUID | string | productbarcodeguid | productbarcodeguid | ค่าของ ProductBarcodeGUID ตามฟิลด์ `productbarcodeguid` ในซอร์ส |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| ProductName | string | productname | productname | ค่าของ ProductName ตามฟิลด์ `productname` ในซอร์ส |
| PriceType | string | pricetype | pricetype | ค่าของ PriceType ตามฟิลด์ `pricetype` ในซอร์ส |
| KeyNumber | int | keynumber | keynumber | ค่าของ KeyNumber ตามฟิลด์ `keynumber` ในซอร์ส |
| OldPrice | float64 | oldprice | oldprice | ค่าของ OldPrice ตามฟิลด์ `oldprice` ในซอร์ส |
| NewPrice | float64 | newprice | newprice | ค่าของ NewPrice ตามฟิลด์ `newprice` ในซอร์ส |
| PriceDifference | float64 | pricedifference | pricedifference | ค่าของ PriceDifference ตามฟิลด์ `pricedifference` ในซอร์ส |
| Action | string | action | action | ค่าของ Action ตามฟิลด์ `action` ในซอร์ส |
| CreatedBy | string | createdby | createdby | ค่าของ CreatedBy ตามฟิลด์ `createdby` ในซอร์ส |
| CreatedAt | time.Time | createdat | createdat | วันเวลาที่สร้าง |
| Remark | string | remark | remark | ค่าของ Remark ตามฟิลด์ `remark` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[DocIdentity]]
- MongoDB collection: `productbarcodespricehistory`
