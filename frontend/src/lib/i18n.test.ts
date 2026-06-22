import { describe, expect, it } from "vitest";
import { LANGUAGES, normalizeLanguage, t } from "./i18n";
import type { TranslationKey } from "./i18n";
import th from "@/locales/th.json";
import en from "@/locales/en.json";
import cn from "@/locales/cn.json";
import ja from "@/locales/ja.json";
import ko from "@/locales/ko.json";
import lo from "@/locales/lo.json";
import my from "@/locales/my.json";
import km from "@/locales/km.json";
import vi from "@/locales/vi.json";
import ms from "@/locales/ms.json";
import id from "@/locales/id.json";
import fil from "@/locales/fil.json";

describe("i18n helpers", () => {
  it("normalizes browser language aliases", () => {
    expect(normalizeLanguage("zh-CN")).toBe("cn");
    expect(normalizeLanguage("tl-PH")).toBe("fil");
  });

  it("falls back to Thai for unknown languages", () => {
    expect(normalizeLanguage("de-DE")).toBe("th");
  });

  it("returns translated login labels", () => {
    expect(t("th", "login")).toBe("เข้าสู่ระบบ");
    expect(t("en", "login")).toBe("Login");
  });

  it("states that Google login acts as first-time signup", () => {
    expect(t("th", "firstUseDescription")).toContain("Google");
    expect(t("th", "firstUseDescription")).toContain("สมัคร");
    expect(t("th", "createHoldingDescription")).toContain("หลัง login");
    expect(t("th", "firstUseStepGoogle")).toContain("Google");
    expect(t("th", "holdingCodeHint")).toContain("รหัสผ่าน");
    expect(t("th", "holdingCodeHint")).not.toContain("Holding");
    expect(t("th", "passwordLoginSectionTitle")).toContain("รหัสผ่าน");
    expect(t("th", "authLoginSectionTitle")).toContain("ยืนยันตัวตน");
  });

  it("returns translated theme toggle labels", () => {
    expect(t("th", "switchToDarkTheme")).toBe("เปลี่ยนเป็นธีมมืด");
    expect(t("en", "switchToLightTheme")).toBe("Switch to light theme");
  });

  it("returns translated settings labels", () => {
    expect(t("th", "settings")).toBe("ตั้งค่า");
    expect(t("en", "backToLogin")).toBe("Back to login");
  });

  it("has translated login social and first-use labels for every supported language", () => {
    const keys = [
      "firstUseTitle",
      "firstUseDescription",
      "multiCompanyDescription",
      "authLoginSectionTitle",
      "loginWithGoogle",
      "socialLoginSeparator",
      "popupBlocked",
      "googleLoginTimeout",
      "createHoldingDescription",
      "enterHoldingCode",
      "firstUseStepGoogle",
      "firstUseStepHolding",
      "firstUseStepWorkspace",
      "holdingCode",
      "holdingCodeHint",
      "holdingCodeInvalid",
      "passwordLoginSectionDescription",
      "passwordLoginSectionTitle",
    ] as const;

    for (const language of LANGUAGES) {
      for (const key of keys) {
        expect(t(language.code, key)).not.toBe(key);
      }
    }
  });

  it("does not fall back to Thai when another language is selected", () => {
    expect(t("km", "firstUseTitle")).not.toBe(t("th", "firstUseTitle"));
    expect(t("km", "loginWithGoogle")).not.toBe(t("th", "loginWithGoogle"));
    expect(t("km", "passwordLoginSectionTitle")).not.toBe(t("th", "passwordLoginSectionTitle"));
  });

  it("falls back to the key id when a selected-language value is missing", () => {
    expect(t("lo", "missingKey" as unknown as TranslationKey)).toBe("missingKey");
  });

  it("keeps every locale in sync with the Thai source of truth (no drift)", () => {
    // Every key present in th.json must also exist (non-empty) in every other locale file.
    // Catches the "added a key to th.json but forgot the other languages" mistake.
    const thKeys = Object.keys(th).sort();
    expect(thKeys.length).toBeGreaterThan(0);

    const locales: Record<string, Record<string, string>> = {
      en, cn, ja, ko, lo, my, km, vi, ms, id, fil,
    };

    for (const [lang, dict] of Object.entries(locales)) {
      const langKeys = Object.keys(dict).sort();
      const missing = thKeys.filter((key) => !langKeys.includes(key));
      expect(missing, `${lang}.json is missing keys present in th.json: ${missing.join(", ")}`).toEqual([]);
    }
  });

  it("ensures all locale files use the same key set as th.json", () => {
    // No locale may introduce keys that th.json does not have (prevents typos / extra keys).
    const thKeySet = new Set(Object.keys(th));
    const locales: Record<string, Record<string, string>> = {
      en, cn, ja, ko, lo, my, km, vi, ms, id, fil,
    };

    for (const [lang, dict] of Object.entries(locales)) {
      const extras = Object.keys(dict).filter((key) => !thKeySet.has(key));
      expect(extras, `${lang}.json has keys not in th.json: ${extras.join(", ")}`).toEqual([]);
    }
  });
});
