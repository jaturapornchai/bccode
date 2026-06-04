import { describe, expect, it } from "vitest";
import { MARKETPLACE_PLATFORMS, emptyMarketplaceProductMap, emptyMarketplaceSKUMap } from "./types";

describe("product barcode marketplace contracts", () => {
  it("supports all planned marketplace import and sync channels", () => {
    expect(MARKETPLACE_PLATFORMS).toEqual([
      "shopee",
      "lazada",
      "aliexpress",
      "tiktok",
    ]);
  });

  it("builds the unified mapping payload for AliExpress", () => {
    expect(emptyMarketplaceProductMap("aliexpress")).toMatchObject({
      platform: "aliexpress",
      seller_sku: "",
      shop_sku: "",
      gtin: "",
      platform_price: 0,
      platform_stock: 0,
      sync_stock: false,
      sync_price: false,
      media_assets: [],
      specification_groups: [],
      raw_attributes: [],
      payload_examples: [],
    });
  });

  it("builds marketplace SKU mappings with dimension stock projections", () => {
    expect(emptyMarketplaceSKUMap("shopee", "shop-a", "ITEM-1")).toMatchObject({
      platform: "shopee",
      holding_code: "shop-a",
      market_item_id: "ITEM-1",
      platform_stock: 0,
      marketplace_dimension_stocks: [],
    });
  });
});
