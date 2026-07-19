---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# MarketplaceSKUMap

โครงสร้าง `MarketplaceSKUMap` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Platform | string | platform | platform | รหัสแพลตฟอร์ม |
| AccountID | string | accountid | accountid | ค่าของ AccountID ตามฟิลด์ `accountid` ในซอร์ส |
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| MarketItemID | string | marketitemid | marketitemid | ค่าของ MarketItemID ตามฟิลด์ `marketitemid` ในซอร์ส |
| MarketModelID | string | marketmodelid | marketmodelid | ค่าของ MarketModelID ตามฟิลด์ `marketmodelid` ในซอร์ส |
| SellerSKU | string | sellersku | sellersku | ค่าของ SellerSKU ตามฟิลด์ `sellersku` ในซอร์ส |
| ShopSKU | string | shopsku | shopsku | ค่าของ ShopSKU ตามฟิลด์ `shopsku` ในซอร์ส |
| GTIN | string | gtin | gtin | ค่าของ GTIN ตามฟิลด์ `gtin` ในซอร์ส |
| MediaAssets | *[][[product-MarketplaceMediaAsset\|MarketplaceMediaAsset]] | mediaassets | mediaassets | ค่าของ MediaAssets ตามฟิลด์ `mediaassets` ในซอร์ส |
| RawAttributes | *[][[product-MarketplaceAttribute\|MarketplaceAttribute]] | rawattributes | rawattributes | ค่าของ RawAttributes ตามฟิลด์ `rawattributes` ในซอร์ส |
| Currency | string | currency | currency | รหัสสกุลเงิน |
| SyncStock | bool | syncstock | syncstock | ค่าของ SyncStock ตามฟิลด์ `syncstock` ในซอร์ส |
| SyncPrice | bool | syncprice | syncprice | ค่าของ SyncPrice ตามฟิลด์ `syncprice` ในซอร์ส |
| CustomPrice | float64 | customprice | customprice | ค่าของ CustomPrice ตามฟิลด์ `customprice` ในซอร์ส |
| PlatformPrice | float64 | platformprice | platformprice | ค่าของ PlatformPrice ตามฟิลด์ `platformprice` ในซอร์ส |
| PlatformStock | int | platformstock | platformstock | ค่าของ PlatformStock ตามฟิลด์ `platformstock` ในซอร์ส |
| MarketplaceDimensionStocks | *[][[product-MarketplaceDimensionStock\|MarketplaceDimensionStock]] | marketplacedimensionstocks | marketplacedimensionstocks | ค่าของ MarketplaceDimensionStocks ตามฟิลด์ `marketplacedimensionstocks` ในซอร์ส |
| Status | string | status | status | สถานะ |
| SyncEnabled | bool | syncenabled | syncenabled | ค่าของ SyncEnabled ตามฟิลด์ `syncenabled` ในซอร์ส |
| LastSyncAt | string | lastsyncat | lastsyncat | ค่าของ LastSyncAt ตามฟิลด์ `lastsyncat` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[product-MarketplaceMediaAsset|MarketplaceMediaAsset]], [[product-MarketplaceAttribute|MarketplaceAttribute]], [[product-MarketplaceDimensionStock|MarketplaceDimensionStock]]
