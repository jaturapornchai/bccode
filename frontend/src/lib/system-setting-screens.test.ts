import { describe, expect, it } from "vitest";
import { THAI_ADDRESS_SUBKEYS } from "@/components/system-settings/types";
import { getSystemSettingConfig, USER_FORM_SECTION_KEYS } from "./system-setting-screens";

describe("system setting screen configs", () => {
  // UAT S7 2026-09-24: "ใช้งานได้ถึงวันที่" was declared on the user config but no form section listed it
  it("shows every user field in a user form section, including the access expiry date", () => {
    const sectionKeys = new Set<string>(Object.values(USER_FORM_SECTION_KEYS).flat());
    const fields = getSystemSettingConfig("user")?.fields.map((field) => field.key) ?? [];

    expect(fields).toContain("accessexpirydate");
    expect(USER_FORM_SECTION_KEYS.accessStatus).toContain("accessexpirydate");
    expect(fields.filter((key) => !sectionKeys.has(key))).toEqual([]);
  });

  it("uses the immutable guidfixed field when deleting product units and allows all companies", () => {
    const config = getSystemSettingConfig("productunit");

    expect(config?.basePath).toBe("/unit");
    expect(config?.idField).toBe("guidfixed");
    expect(config?.fields.map((field) => field.key)).toEqual(["unitcode", "names"]);
    expect(config?.fields.some((field) => field.key === "businesscodes")).toBe(false);
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
    const addressField = config?.fields.find((field) => field.key === "contact");
    expect(addressField?.type).toBe("thai-address");
    expect(THAI_ADDRESS_SUBKEYS).toEqual([
      "countrycode",
      "provincecode",
      "districtcode",
      "subdistrictcode",
      "zipcode",
    ]);
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

  it("keeps company single-record fields limited to logo, name and address", () => {
    const config = getSystemSettingConfig("company");
    const keys = config?.fields.map((field) => field.key) ?? [];

    expect(keys).toEqual(["logouri", "names", "address"]);
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

  it("provides configuration for book bank master screen with main-crud API and redirects bank alias", () => {
    const config = getSystemSettingConfig("bookbankscreen");
    expect(config).toBeDefined();
    expect(config?.route).toBe("/bookbankscreen");
    expect(config?.kind).toBe("main-crud");
    expect(config?.icon).toBe("bank");
    expect(config?.basePath).toBe("/payment/bookbank");
    expect(config?.listPath).toBe("/payment/bookbank/list");
    expect(config?.idField).toBe("guidfixed");
    expect(config?.title.th).toBe("บันทึกสมุดเงินฝากธนาคาร");
    expect(config?.fields.map((f) => f.key)).toEqual([
      "bookcode",
      "names",
      "passbook",
      "bankbranch",
      "accountname",
      "bankcode",
      "banknames",
      "accountcode",
      "logo",
    ]);
    expect(config?.fields[0]?.businessCode).toBe(true);
    expect(config?.fields[8]?.type).toBe("image-upload");

    const aliasConfig = getSystemSettingConfig("bank");
    expect(aliasConfig?.slug).toBe("bookbankscreen");
  });
});
