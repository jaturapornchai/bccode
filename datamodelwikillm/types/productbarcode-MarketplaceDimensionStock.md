---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# MarketplaceDimensionStock

โครงสร้าง `MarketplaceDimensionStock` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DimensionKey | string | dimensionkey | dimensionkey | ค่าของ DimensionKey ตามฟิลด์ `dimensionkey` ในซอร์ส |
| DimensionName | string | dimensionname | dimensionname | ค่าของ DimensionName ตามฟิลด์ `dimensionname` ในซอร์ส |
| MarketDimensionID | string | marketdimensionid | marketdimensionid | ค่าของ MarketDimensionID ตามฟิลด์ `marketdimensionid` ในซอร์ส |
| AvailableQty | float64 | availableqty | availableqty | ค่าของ AvailableQty ตามฟิลด์ `availableqty` ในซอร์ส |
| ReservedQty | float64 | reservedqty | reservedqty | ค่าของ ReservedQty ตามฟิลด์ `reservedqty` ในซอร์ส |
| InboundQty | float64 | inboundqty | inboundqty | ค่าของ InboundQty ตามฟิลด์ `inboundqty` ในซอร์ส |
| OversellBufferQty | float64 | oversellbufferqty | oversellbufferqty | ค่าของ OversellBufferQty ตามฟิลด์ `oversellbufferqty` ในซอร์ส |
| LastPlatformStock | float64 | lastplatformstock | lastplatformstock | ค่าของ LastPlatformStock ตามฟิลด์ `lastplatformstock` ในซอร์ส |
| LastSyncedAt | string | lastsyncedat | lastsyncedat | ค่าของ LastSyncedAt ตามฟิลด์ `lastsyncedat` ในซอร์ส |
| LastSyncStatus | string | lastsyncstatus | lastsyncstatus | ค่าของ LastSyncStatus ตามฟิลด์ `lastsyncstatus` ในซอร์ส |
| LastSyncError | string | lastsyncerror | lastsyncerror | ค่าของ LastSyncError ตามฟิลด์ `lastsyncerror` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
