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

    expect(config?.fields[0]?.key).toBe("languages");
    expect(keys.has("languages")).toBe(true);
    expect(keys.has("language")).toBe(false);
    expect(keys.has("contact.address")).toBe(true);
    expect(keys.has("pos.vatrate")).toBe(true);
    expect(keys.has("paymentrounding")).toBe(true);
    expect(keys.has("couponusetype")).toBe(true);
    expect(keys.has("imageuri")).toBe(true);
    expect(keys.has("logouri")).toBe(true);
    expect(keys.has("businesstype")).toBe(true);
  });

  it("requires company language setup before company details", () => {
    const config = getSystemSettingConfig("company");
    const keys = new Set(config?.fields.map((field) => field.key));

    expect(config?.fields[0]?.key).toBe("settings.languageconfigs");
    expect(keys.has("settings.languageconfigs")).toBe(true);
    expect(keys.has("settings.language")).toBe(false);
  });

  it("keeps company single-record fields in a full-width friendly order", () => {
    const config = getSystemSettingConfig("company");
    const keys = config?.fields.map((field) => field.key) ?? [];

    expect(keys.slice(0, 4)).toEqual(["settings.languageconfigs", "names", "address", "telephone"]);
  });

  it("offers full-month date formats with localized examples", () => {
    const config = getSystemSettingConfig("company");
    const dateFormat = config?.fields.find((field) => field.key === "settings.date_format");
    const fullMonth = dateFormat?.options?.find((option) => option.value === "dd MMMM yyyy");

    expect(fullMonth?.labels?.th).toContain("22 พฤษภาคม 2569");
    expect(fullMonth?.labels?.en).toContain("22 May 2026");
    expect(dateFormat?.options?.some((option) => option.value === "MMMM dd, yyyy")).toBe(true);
  });
});
