# Marketplace Field Mapping (Shopee / Lazada / TikTok Shop / AliExpress)

Maps this system's **canonical** marketplace contract to each marketplace's official Open Platform API fields, so product/SKU data flows in/out losslessly. Scope rule: `../rules/bc-account-core-rules.md` → "System Scope & Marketplace Support".

> Marketplace field names below come from official Open Platform docs + vendor SDKs (sources at the bottom). They are **version-sensitive** — re-verify against the live official doc when building or changing a connector (IRON RULE 6; do not fabricate field names).

## Canonical contract (source of truth)
Marketplace mapping lives on the **product / SKU layer** (the sellable item), NOT on reusable templates:
- Go: `MarketplaceProductMap`, `MarketplaceSKUMap`, `MarketplaceDimensionStock` in `backend/internal/product/product/models/product.go` (+ mirror in `productbarcode/models/product_barcode.go`).
- Frontend mirror: `frontend/src/lib/product-barcode/types.ts` (`MARKETPLACE_PLATFORMS = shopee | lazada | aliexpress | tiktok`).
- UI: `frontend/src/app/menu/tab-product-marketplace.tsx`, `marketplace-screen.tsx`. Per-channel selling price: `/channelprice`.

`MarketplaceProductMap` = item-level listing; `MarketplaceSKUMap` = per-variation/barcode; `MarketplaceDimensionStock` = per-dimension stock projection.

## Mapping table (contract field → marketplace API field)
| Contract field | Shopee (product v2) | Lazada (product) | TikTok Shop (product 202309) | AliExpress (solution.product) |
|---|---|---|---|---|
| `marketitemid` | `item_id` | `item_id` (product) | `product_id` | `product_id` |
| `marketmodelid` | `model_id` | `SkuId` | `sku_id` | sku `id` |
| `sellersku` | `item_sku` / `model_sku` | `SellerSku` | `seller_sku` | `sku_code` |
| `shopsku` | (`model_sku`) | `ShopSku` | — | — |
| `gtin` | barcode (add_item) | barcode attr | `gtin` (mandatory some cats) | — |
| `categoryid` | `category_id` | `primary_category` | `category_id` | `category_id` |
| `platformprice` | model `price_info.current_price` | sku `price` | sku `price.amount` | `sku_price` |
| `platformstock` | model `stock_info` (seller_stock) | sku `quantity` | sku `inventory[].quantity` | `ipm_sku_stock` |
| `status` | `item_status` (NORMAL/UNLIST/BANNED) | product `status` | `status` (DRAFT/LIVE/...) | product status |
| `marketplacedimensionstocks[].marketdimensionid` | model `tier_index` `[i,j]` | `SaleProp` value | `sales_attributes[].value_id` | `sku_attr` `pid:vid#name` |
| `specificationgroups` / `rawattributes` | category attributes (`get_attributes`) | `Attributes` + `SaleProp` (`GetCategoryAttributes`) | product + `sales_attributes` | aeop product/sku attributes |

### Variation model per marketplace
- **Shopee:** `tier_variation:[{name, option_list:[{option}]}]` defines axes; each `model` = `model_id`, `model_sku`, `tier_index:[i,j]` (index into tier options), `price_info[].current_price`, stock. Endpoints: `product/init_tier_variation`, `add_model`, `get_model_list`, `update_model`, `update_stock`, `update_price`.
- **Lazada:** SPU → Product → SKU. CreateProduct needs `category_id` (GetCategoryTree) + product `Attributes` + per-SKU `SaleProp` (GetCategoryAttributes). Each SKU has `SellerSku` (ours) + `SkuId`/`ShopSku` (theirs) + price/quantity. New SPU flow: SearchSPUs, UpdatePriceQuantity. Error `[SELLER_SKU_IS_EXIST]` if SellerSku dup.
- **TikTok Shop:** product `skus[]`; each = `seller_sku`, `sales_attributes[]` ({attribute_id, attribute_name, value_id, value_name} = color/size), `inventory[]` ({warehouse_id, quantity}), `price` ({amount, currency}); `sku_id` returned on create. Category via Get Category; `gtin` mandatory for some categories.
- **AliExpress:** `aeop_ae_product_sku[]`; each = `sku_code` (ours), `sku_property`/`sku_attr` (`"14:350853#Black;5:361386"` = propId:valId#name;...), `ipm_sku_stock`, `sku_price`; sku `id` returned. API `aliexpress.solution.product.post`.

## Example — Shopee (real shape, mapped to the contract)
A "เสื้อยืด" listed on Shopee with 2 variations (RED/S, RED/M). `MarketplaceProductMap` (item level):
```
platform=shopee  accountid=<shop_id>  marketitemid=22000000123  (Shopee item_id)
sellersku=TEE-RED  categoryid=100629  categoryname="Men Clothes > T-Shirts"
currency=THB  status=LIVE  (item_status NORMAL)  syncstock=true  syncprice=true
```
Two `MarketplaceSKUMap` (one per Shopee model):
```
#1 marketitemid=22000000123  marketmodelid=30000111 (model_id)  sellersku=TEE-RED-S (model_sku)
   gtin=8850000000017  platformprice=299  platformstock=50  status=NORMAL
   marketplacedimensionstocks=[{dimensionkey:"S", marketdimensionid:"[0,0]" (tier_index), availableqty:50}]
#2 marketitemid=22000000123  marketmodelid=30000112  sellersku=TEE-RED-M
   gtin=8850000000024  platformprice=299  platformstock=30
   marketplacedimensionstocks=[{dimensionkey:"M", marketdimensionid:"[0,1]", availableqty:30}]
```
Shopee specifics: `marketmodelid` = Shopee `model_id`; `marketdimensionid` carries the `tier_index` (`[colorIdx, sizeIdx]`) that maps to `tier_variation` option positions; selling price/stock are per-model (`get_model_list`).

## Status
- Datamodel + UX: ready (contract above holds item id, model id, our/their SKU, GTIN, category, price, stock, per-dimension stock, sync flags, status).
- Connectors (OAuth/token, API client per marketplace, category+attribute fetch, sync push/pull jobs): **NOT built**. 100% in/out is achievable on this schema once connectors are built and each field is re-verified against the live official doc at build time.

## Sources (official-first)
- Shopee Open Platform: https://open.shopee.com (Product module v2 — init_tier_variation / add_model / get_model_list)
- Lazada Open Platform: https://open.lazada.com/apps/doc/api?path=/product/create ; https://lazada-sellercenter.readme.io/docs/createproduct
- TikTok Shop Partner Center: https://partner.tiktokshop.com/docv2/page/products-api-overview
- AliExpress Open Platform: https://openservice.aliexpress.com/doc/api.htm (aliexpress.solution.product.post)
