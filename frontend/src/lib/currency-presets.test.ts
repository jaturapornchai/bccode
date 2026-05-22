import { describe, expect, it } from "vitest";
import {
  applyCurrencySymbolPreset,
  currencyPresetSource,
  currencySymbolPresets,
  filterCurrencySymbolPresets,
  findCurrencySymbolPreset,
} from "@/lib/currency-presets";

describe("currency presets", () => {
  it("fills currency code and name when selecting a symbol preset", () => {
    const preset = currencySymbolPresets.find((item) => item.code === "VND");
    expect(preset).toBeDefined();

    expect(applyCurrencySymbolPreset({ code: "USD", name: "US Dollar", symbol: "$" }, preset!, false)).toEqual({
      code: "VND",
      name: "Vietnamese Dong",
      symbol: "₫",
    });
  });

  it("keeps code when an existing currency code is locked for edit", () => {
    const preset = currencySymbolPresets.find((item) => item.code === "EUR");
    expect(preset).toBeDefined();

    expect(applyCurrencySymbolPreset({ code: "USD", name: "US Dollar", symbol: "$" }, preset!, true)).toEqual({
      code: "USD",
      name: "Euro",
      symbol: "€",
    });
  });

  it("searches presets by code, name, or symbol", () => {
    expect(filterCurrencySymbolPresets("dong").map((item) => item.code)).toContain("VND");
    expect(filterCurrencySymbolPresets("renminbi").map((item) => item.code)).toContain("CNY");
    expect(filterCurrencySymbolPresets("₭").map((item) => item.code)).toContain("LAK");
    expect(filterCurrencySymbolPresets("cad").map((item) => item.code)).toEqual(["CAD"]);
  });

  it("confirms a selected currency exists only when code name and symbol match", () => {
    expect(findCurrencySymbolPreset({ code: "THB", name: "Thai Baht", symbol: "฿" })?.code).toBe("THB");
    expect(findCurrencySymbolPreset({ code: "THB", name: "Baht", symbol: "฿" })?.code).toBe("THB");
    expect(findCurrencySymbolPreset({ code: "USD", name: "US Dollar", symbol: "US$" })?.code).toBe("USD");
    expect(findCurrencySymbolPreset({ code: "CHF", name: "Swiss Franc", symbol: "Fr." })?.code).toBe("CHF");
    expect(findCurrencySymbolPreset({ code: "CAD", name: "Canadian Dollar", symbol: "C$" })?.code).toBe("CAD");
    expect(findCurrencySymbolPreset({ code: "THB", name: "Wrong", symbol: "฿" })).toBeUndefined();
  });

  it("keeps official ISO 4217 names from the verified source", () => {
    expect(currencyPresetSource).toMatchObject({
      standard: "ISO 4217",
      authority: "SIX Financial Information",
      list: "List One: Current Currency & Funds",
      symbolSource: "Wikipedia currency symbol and currency pages",
    });
    expect(currencySymbolPresets.find((item) => item.code === "CNY")?.isoName).toBe("Yuan Renminbi");
    expect(currencySymbolPresets.find((item) => item.code === "JPY")?.isoName).toBe("Yen");
    expect(currencySymbolPresets.find((item) => item.code === "CAD")?.symbol).toBe("Can$");
    expect(currencySymbolPresets.find((item) => item.code === "CHF")?.symbolKind).toBe("abbreviation");
  });
});
