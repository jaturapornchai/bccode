---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# MarketplaceProductMap

โครงสร้าง `MarketplaceProductMap` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Platform | string | platform | platform | รหัสแพลตฟอร์ม |
| AccountID | string | accountid | accountid | ค่าของ AccountID ตามฟิลด์ `accountid` ในซอร์ส |
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| MarketItemID | string | marketitemid | marketitemid | ค่าของ MarketItemID ตามฟิลด์ `marketitemid` ในซอร์ส |
| MarketModelID | string | marketmodelid | marketmodelid | ค่าของ MarketModelID ตามฟิลด์ `marketmodelid` ในซอร์ส |
| ItemURL | string | itemurl | itemurl | ค่าของ ItemURL ตามฟิลด์ `itemurl` ในซอร์ส |
| SellerSKU | string | sellersku | sellersku | ค่าของ SellerSKU ตามฟิลด์ `sellersku` ในซอร์ส |
| ShopSKU | string | shopsku | shopsku | ค่าของ ShopSKU ตามฟิลด์ `shopsku` ในซอร์ส |
| GTIN | string | gtin | gtin | ค่าของ GTIN ตามฟิลด์ `gtin` ในซอร์ส |
| CategoryID | string | categoryid | categoryid | ค่าของ CategoryID ตามฟิลด์ `categoryid` ในซอร์ส |
| CategoryName | string | categoryname | categoryname | ค่าของ CategoryName ตามฟิลด์ `categoryname` ในซอร์ส |
| BrandID | string | brandid | brandid | ค่าของ BrandID ตามฟิลด์ `brandid` ในซอร์ส |
| MediaAssets | *[][[productbarcode-MarketplaceMediaAsset\|MarketplaceMediaAsset]] | mediaassets | mediaassets | ค่าของ MediaAssets ตามฟิลด์ `mediaassets` ในซอร์ส |
| SpecificationGroups | *[][[productbarcode-MarketplaceSpecificationGroup\|MarketplaceSpecificationGroup]] | specificationgroups | specificationgroups | ค่าของ SpecificationGroups ตามฟิลด์ `specificationgroups` ในซอร์ส |
| RawAttributes | *[][[productbarcode-MarketplaceAttribute\|MarketplaceAttribute]] | rawattributes | rawattributes | ค่าของ RawAttributes ตามฟิลด์ `rawattributes` ในซอร์ส |
| PayloadExamples | *[][[productbarcode-MarketplacePayloadExample\|MarketplacePayloadExample]] | payloadexamples | payloadexamples | ค่าของ PayloadExamples ตามฟิลด์ `payloadexamples` ในซอร์ส |
| Currency | string | currency | currency | รหัสสกุลเงิน |
| CustomPrice | float64 | customprice | customprice | ค่าของ CustomPrice ตามฟิลด์ `customprice` ในซอร์ส |
| PlatformPrice | float64 | platformprice | platformprice | ค่าของ PlatformPrice ตามฟิลด์ `platformprice` ในซอร์ส |
| PlatformStock | int | platformstock | platformstock | ค่าของ PlatformStock ตามฟิลด์ `platformstock` ในซอร์ส |
| SyncStock | bool | syncstock | syncstock | ค่าของ SyncStock ตามฟิลด์ `syncstock` ในซอร์ส |
| SyncPrice | bool | syncprice | syncprice | ค่าของ SyncPrice ตามฟิลด์ `syncprice` ในซอร์ส |
| Status | string | status | status | สถานะ |
| RejectReason | string | rejectreason | rejectreason | ค่าของ RejectReason ตามฟิลด์ `rejectreason` ในซอร์ส |
| DaysToShip | int | daystoship | daystoship | ค่าของ DaysToShip ตามฟิลด์ `daystoship` ในซอร์ส |
| IsPreOrder | bool | ispreorder | ispreorder | ค่าของ IsPreOrder ตามฟิลด์ `ispreorder` ในซอร์ส |
| SyncEnabled | bool | syncenabled | syncenabled | ค่าของ SyncEnabled ตามฟิลด์ `syncenabled` ในซอร์ส |
| SyncStatus | string | syncstatus | syncstatus | ค่าของ SyncStatus ตามฟิลด์ `syncstatus` ในซอร์ส |
| LastSyncAt | string | lastsyncat | lastsyncat | ค่าของ LastSyncAt ตามฟิลด์ `lastsyncat` ในซอร์ส |
| LastSyncError | string | lastsyncerror | lastsyncerror | ค่าของ LastSyncError ตามฟิลด์ `lastsyncerror` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-MarketplaceMediaAsset|MarketplaceMediaAsset]], [[productbarcode-MarketplaceSpecificationGroup|MarketplaceSpecificationGroup]], [[productbarcode-MarketplaceAttribute|MarketplaceAttribute]], [[productbarcode-MarketplacePayloadExample|MarketplacePayloadExample]]
