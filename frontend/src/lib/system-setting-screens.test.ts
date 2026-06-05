import { describe, expect, it } from "vitest";
import { getSystemSettingConfig } from "./system-setting-screens";

describe("system setting screen configs", () => {
  it("uses the immutable guidfixed field when deleting product units", () => {
    const config = getSystemSettingConfig("productunit");

    expect(config?.basePath).toBe("/unit");
    expect(config?.idField).toBe("guidfixed");
  });

  it("keeps branch settings aligned with the legacy Flutter branch model", () => {
    const config = getSystemSettingConfig("branch");
    const keys = new Set(config?.fields.map((field) => field.key));

    expect(keys.size).toBe(config?.fields.length);
    expect(config?.fields[0]?.key).toBe("languages");
    expect(keys.has("languages")).toBe(true);
    expect(keys.has("language")).toBe(false);
    expect(keys.has("contact.address")).toBe(true);
    expect(keys.has("contact.countrycode")).toBe(true);
    expect(keys.has("contact.provincecode")).toBe(true);
    expect(keys.has("contact.districtcode")).toBe(true);
    expect(keys.has("contact.subdistrictcode")).toBe(true);
    expect(keys.has("contact.zipcode")).toBe(true);
    expect(keys.has("contact.phonenumber")).toBe(true);
    expect(keys.has("yeartype")).toBe(true);
    expect(keys.has("companyregistrationno")).toBe(true);
    expect(keys.has("isvatregistered")).toBe(true);
    expect(keys.has("pos.taxid")).toBe(true);
    expect(keys.has("pos.vatrate")).toBe(true);
    expect(keys.has("paymentrounding")).toBe(true);
    expect(keys.has("pointconfig")).toBe(true);
    expect(keys.has("machinetype")).toBe(true);
    expect(keys.has("couponusetype")).toBe(true);
    expect(keys.has("imageuris")).toBe(true);
    expect(keys.has("logouri")).toBe(true);
    expect(keys.has("businesstype")).toBe(true);
    for (const key of [
      "isrestaurant",
      "istire",
      "isagriculture",
      "ispharmacy",
      "isretail",
      "isservice",
      "iswholesale",
      "ismanufacturing",
      "isimportexport",
      "iscontractor",
      "isrental",
      "isecommerce",
      "islogistics",
      "iseducation",
      "ishotel",
      "isbeauty",
      "isgoldshop",
      "isaccountingfirm",
      "isconstruction",
      "iselectronics",
      "ismobileshop",
    ]) {
      expect(keys.has(key)).toBe(true);
    }
    expect(keys.has(`pos.${"tax"}_${"id"}`)).toBe(false);
    expect(keys.has(`contact.${"phone"}_${"number"}`)).toBe(false);
    expect(keys.has(`${"year"}_${"type"}`)).toBe(false);
  });

  it("documents Thai legal branch code structure on the branch form", () => {
    const config = getSystemSettingConfig("branch");
    const codeField = config?.fields.find((field) => field.key === "code");

    expect(codeField?.label.th).toBe("รหัสสาขาภาษี");
    expect(codeField?.placeholder).toBe("00000");
    expect(codeField?.helper?.th).toContain("สำนักงานใหญ่ = 00000");
    expect(codeField?.helper?.th).toContain("สาขาที่ 1 = 00001");
  });

  it("requires active language setup config to be defined separately", () => {
    const config = getSystemSettingConfig("activelanguages");
    const keys = new Set(config?.fields.map((field) => field.key));

    expect(config?.fields[0]?.key).toBe("settings.languageconfigs");
    expect(keys.has("settings.languageconfigs")).toBe(true);
    expect(keys.has("settings.language")).toBe(false);
  });

  it("keeps company single-record fields limited to name and address", () => {
    const config = getSystemSettingConfig("company");
    const keys = config?.fields.map((field) => field.key) ?? [];

    expect(keys).toEqual(["names", "address"]);
  });

  it("offers full-month date formats with localized examples", () => {
    const config = getSystemSettingConfig("branch");
    const dateFormat = config?.fields.find(
      (field) => field.key === "dateformat",
    );
    const fullMonth = dateFormat?.options?.find(
      (option) => option.value === "dd MMMM yyyy",
    );

    expect(fullMonth?.labels?.th).toContain("22 พฤษภาคม 2569");
    expect(fullMonth?.labels?.en).toContain("22 May 2026");
    expect(
      dateFormat?.options?.some((option) => option.value === "MMMM dd, yyyy"),
    ).toBe(true);
  });

  it("defines neutral SKU option masters on MongoDB Atlas collections", () => {
    const color = getSystemSettingConfig("productcolor");
    const size = getSystemSettingConfig("productsize");
    const matrix = getSystemSettingConfig("productvariantmatrix");
    const serialRegistry = getSystemSettingConfig("productserialregistry");
    const channelPrice = getSystemSettingConfig("channelprice");

    expect(color).toMatchObject({
      route: "/productcolor",
      kind: "atlas",
      collection: "productcolors",
      idField: "guidfixed",
    });
    expect(size).toMatchObject({
      route: "/productsize",
      kind: "atlas",
      collection: "productsizes",
      idField: "guidfixed",
    });
    expect(matrix).toMatchObject({
      route: "/productvariantmatrix",
      kind: "atlas",
      collection: "productvariantmatrices",
      idField: "guidfixed",
    });
    expect(serialRegistry).toMatchObject({
      route: "/productserialregistry",
      kind: "atlas",
      collection: "productserialregistries",
      idField: "guidfixed",
    });
    expect(channelPrice).toMatchObject({
      route: "/channelprice",
      kind: "atlas",
      collection: "productchannelprices",
      idField: "guidfixed",
    });

    expect(color?.fields.map((field) => field.key)).toContain("aliases");
    expect(size?.fields.map((field) => field.key)).toContain("aliases");
    expect(matrix?.fields.map((field) => field.key)).toEqual(
      expect.arrayContaining([
        "optiontiers",
        "skucombinations",
        "mediaassets",
        "specificationgroups",
        "importattributemaps",
        "integrationprofiles",
        "payloadexamples",
        "serialtrackingmode",
        "businesscodes",
      ]),
    );
    expect(serialRegistry?.fields.map((field) => field.key)).toEqual(
      expect.arrayContaining([
        "serialno",
        "identifiertype",
        "status",
        "itemcode",
        "barcode",
        "warrantystartdate",
        "warrantyenddate",
        "businesscodes",
      ]),
    );
    expect(channelPrice?.fields.map((field) => field.key)).toEqual(
      expect.arrayContaining([
        "pricecode",
        "channelcode",
        "itemcode",
        "barcode",
        "dimensionkey",
        "saleprice",
        "startdate",
        "enddate",
        "businesscodes",
      ]),
    );
  });
});
