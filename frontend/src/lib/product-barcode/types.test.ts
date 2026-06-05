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
      sellersku: "",
      shopsku: "",
      gtin: "",
      platformprice: 0,
      platformstock: 0,
      syncstock: false,
      syncprice: false,
      mediaassets: [],
      specificationgroups: [],
      rawattributes: [],
      payloadexamples: [],
    });
  });

  it("builds marketplace SKU mappings with dimension stock projections", () => {
    expect(emptyMarketplaceSKUMap("shopee", "shop-a", "ITEM-1")).toMatchObject({
      platform: "shopee",
      holdingcode: "shop-a",
      marketitemid: "ITEM-1",
      platformstock: 0,
      marketplacedimensionstocks: [],
    });
  });
});
