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
      "loginWithGoogle",
      "loginWithLine",
      "localGoogleTestLogin",
      "socialLoginSeparator",
      "popupBlocked",
      "googleLoginTimeout",
      "signUp",
      "signUpDescription",
      "displayName",
      "confirmPassword",
      "passwordMismatch",
      "registerSuccess",
      "registerFailed",
      "creatingAccount",
      "createAccount",
    ] as const;

    for (const language of LANGUAGES) {
      for (const key of keys) {
        expect(t(language.code, key)).toBeTruthy();
      }
    }
  });

  it("does not fall back to Thai when another language is selected", () => {
    expect(t("km", "firstUseTitle")).not.toBe(t("th", "firstUseTitle"));
    expect(t("km", "loginWithGoogle")).not.toBe(t("th", "loginWithGoogle"));
  });

  it("falls back to the key id when a selected-language value is missing", () => {
    expect(t("lo", "missingKey" as unknown as Parameters<typeof t>[1])).toBe("missingKey");
  });
});
