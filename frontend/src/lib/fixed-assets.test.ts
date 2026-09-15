import { describe, it, expect } from "vitest";
import { assetName, type FixedAsset } from "./fixed-assets";
import { isFixedAssetRoute, isMenuScreenPending } from "./menu-screen-status";

describe("Fixed Assets Client Library", () => {
  it("resolves asset name with language fallback correctly", () => {
    const asset: FixedAsset = {
      assetcode: "FA-001",
      names: [
        { code: "th", name: "เครื่องคอมพิวเตอร์สำนักงาน" },
        { code: "en", name: "Office Computer" },
      ],
      assettypecode: "COMPUTER",
      branchcode: "HQ",
      departmentcode: "IT",
      locationcode: "FL2",
      purchasedate: "2026-01-01",
      startcalcdate: "2026-01-01",
      cost: "50000.00",
      scrapvalue: "1.00",
      usefullifeyears: 3,
      deprecpercent: "33.33",
      method: "straight_line",
      firstyearpercent: "40.00",
      beginaccumdeprec: "0.00",
      assetaccountcode: "120101",
      accumdeprecaccountcode: "129101",
      deprecexpenseaccountcode: "520103",
      status: "active",
      serialnumber: "SN-998811",
      brand: "Dell",
      model: "OptiPlex",
      suppliercode: "SUP-01",
      istaxdeductible: true,
      notes: "",
    };

    expect(assetName(asset, "th")).toBe("เครื่องคอมพิวเตอร์สำนักงาน");
    expect(assetName(asset, "en")).toBe("Office Computer");
    expect(assetName(asset, "ja")).toBe("เครื่องคอมพิวเตอร์สำนักงาน"); // fallback to th
  });

  it("identifies fixed asset routes properly", () => {
    expect(isFixedAssetRoute("/asset/registry")).toBe(true);
    expect(isFixedAssetRoute("/asset/depreciation")).toBe(true);
    expect(isFixedAssetRoute("/asset/post-gl")).toBe(true);
    expect(isFixedAssetRoute("/asset/disposal")).toBe(true);
    expect(isFixedAssetRoute("/report/assetschedule")).toBe(true);
    expect(isFixedAssetRoute("/gl/chartofaccounts")).toBe(false);
    expect(isFixedAssetRoute("/product")).toBe(false);
  });

  it("ensures fixed asset screens are active and not marked as pending", () => {
    expect(isMenuScreenPending("/asset/registry")).toBe(false);
    expect(isMenuScreenPending("/asset/depreciation")).toBe(false);
    expect(isMenuScreenPending("/asset/post-gl")).toBe(false);
    expect(isMenuScreenPending("/asset/disposal")).toBe(false);
    expect(isMenuScreenPending("/report/assetschedule")).toBe(false);
  });
});
