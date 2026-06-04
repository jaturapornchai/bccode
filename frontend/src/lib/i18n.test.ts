import { describe, expect, it } from "vitest";
import { LANGUAGES, normalizeLanguage, t } from "./i18n";

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
    expect(t("th", "firstUseDescription")).toContain("สมัครสมาชิก");
    expect(t("th", "createHoldingDescription")).toContain("หลัง login");
    expect(t("th", "firstUseStepGoogle")).toContain("ผูกบัญชี");
    expect(t("th", "holdingCodeHint")).toContain("ต้องมีรหัส Holding");
    expect(t("th", "holdingCodeHint")).not.toContain("กดสร้าง Holding");
    expect(t("th", "passwordLoginSectionTitle")).toContain("User , Password");
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
      "localGoogleTestLogin",
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
    expect(t("lo", "missingKey" as unknown as Parameters<typeof t>[1])).toBe("missingKey");
  });
});
