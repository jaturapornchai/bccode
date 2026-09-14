import { describe, expect, it } from "vitest";
import { filterThaiBankPresets, findThaiBankPreset, thaiBankPresets } from "./thai-banks";

describe("Thai Bank Presets", () => {
  it("has a non-empty list of banks with unique codes and botCodes", () => {
    expect(thaiBankPresets.length).toBeGreaterThanOrEqual(20);

    const codes = thaiBankPresets.map((b) => b.code.toUpperCase());
    const uniqueCodes = new Set(codes);
    expect(uniqueCodes.size).toBe(codes.length);

    const botCodes = thaiBankPresets.map((b) => b.botCode.toLowerCase());
    const uniqueBotCodes = new Set(botCodes);
    expect(uniqueBotCodes.size).toBe(botCodes.length);
  });

  it("all presets have valid names, color hex, and logo path", () => {
    for (const bank of thaiBankPresets) {
      expect(bank.code).toBeTruthy();
      expect(bank.nameTh).toBeTruthy();
      expect(bank.nameEn).toBeTruthy();
      expect(bank.shortNameTh).toBeTruthy();
      expect(bank.color).toMatch(/^#[0-9A-Fa-f]{6}$/);
      expect(bank.logo).toMatch(/^\/banks\/[A-Za-z0-9_]+\.png$/);
    }
  });

  it("finds a bank by symbol or BOT code", () => {
    const kbank = findThaiBankPreset("kbank");
    expect(kbank).toBeDefined();
    expect(kbank?.nameTh).toBe("ธนาคารกสิกรไทย");
    expect(kbank?.botCode).toBe("004");

    const scbByBot = findThaiBankPreset("014");
    expect(scbByBot).toBeDefined();
    expect(scbByBot?.code).toBe("SCB");

    expect(findThaiBankPreset("nonexistent")).toBeUndefined();
  });

  it("filters banks by Thai name, English name, or symbol", () => {
    const kResults = filterThaiBankPresets("กสิกร");
    expect(kResults.some((b) => b.code === "KBANK")).toBe(true);

    const bblResults = filterThaiBankPresets("Bangkok");
    expect(bblResults.some((b) => b.code === "BBL")).toBe(true);

    const promptPayResults = filterThaiBankPresets("PromptPay");
    expect(promptPayResults.some((b) => b.code === "PromptPay")).toBe(true);

    const emptyFilter = filterThaiBankPresets("");
    expect(emptyFilter.length).toBe(thaiBankPresets.length);
  });
});
