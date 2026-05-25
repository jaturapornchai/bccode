import { describe, expect, it } from "vitest";
import { getSystemSettingConfig } from "./system-setting-screens";

describe("system setting screen configs", () => {
  it("uses the legacy unit guid field when deleting product units", () => {
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
    expect(keys.has("contact.country_code")).toBe(true);
    expect(keys.has("contact.province_code")).toBe(true);
    expect(keys.has("contact.district_code")).toBe(true);
    expect(keys.has("contact.sub_district_code")).toBe(true);
    expect(keys.has("contact.zip_code")).toBe(true);
    expect(keys.has("contact.phone_number")).toBe(true);
    expect(keys.has("year_type")).toBe(true);
    expect(keys.has("company_registration_no")).toBe(true);
    expect(keys.has("is_vat_registered")).toBe(true);
    expect(keys.has("pos.tax_id")).toBe(true);
    expect(keys.has("pos.vatrate")).toBe(true);
    expect(keys.has("paymentrounding")).toBe(true);
    expect(keys.has("pointconfig")).toBe(true);
    expect(keys.has("machinetype")).toBe(true);
    expect(keys.has("couponusetype")).toBe(true);
    expect(keys.has("imageuri")).toBe(true);
    expect(keys.has("logouri")).toBe(true);
    expect(keys.has("businesstype")).toBe(true);
    for (const key of [
      "is_restaurant",
      "is_tire",
      "is_agriculture",
      "is_pharmacy",
      "is_retail",
      "is_service",
      "is_wholesale",
      "is_manufacturing",
      "is_import_export",
      "is_contractor",
      "is_rental",
      "is_ecommerce",
      "is_logistics",
      "is_education",
      "is_hotel",
      "is_beauty",
      "is_gold_shop",
      "is_accounting_firm",
      "is_construction",
      "is_electronics",
      "is_mobile_shop",
    ]) {
      expect(keys.has(key)).toBe(true);
    }
    expect(keys.has("pos.taxid")).toBe(false);
    expect(keys.has("contact.phonenumber")).toBe(false);
    expect(keys.has("yeartype")).toBe(false);
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
    const config = getSystemSettingConfig("active_languages");
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
      (field) => field.key === "date_format",
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
});
