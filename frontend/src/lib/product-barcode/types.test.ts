import { describe, expect, it } from "vitest";
import { MARKETPLACE_PLATFORMS, emptyMarketplaceProductMap } from "./types";

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
});
