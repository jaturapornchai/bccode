import { describe, expect, it } from "vitest";
import { backendText, isBackendLanguageLoading, isBackendLanguageReady, type BackendLanguageDictionary } from "./backend-language";

function markedDictionary(values: BackendLanguageDictionary, ready: boolean, loading = false): BackendLanguageDictionary {
  Object.defineProperties(values, {
    __backendLanguageLoading: { enumerable: false, value: String(loading) },
    __backendLanguageReady: { enumerable: false, value: String(ready) },
  });
  return values;
}

describe("backend language helpers", () => {
  it("uses fallback text before backend dictionary is ready", () => {
    expect(backendText({}, "system_settings", "ตั้งค่าระบบ")).toBe("ตั้งค่าระบบ");
  });

  it("keeps missing translations visible as key ids after backend dictionary is ready", () => {
    expect(backendText(markedDictionary({}, true), "missing_key", "fallback")).toBe("missing_key");
  });

  it("reads language readiness markers without exposing them as translations", () => {
    const dictionary = markedDictionary({ menu: "เมนู" }, true, true);
    expect(isBackendLanguageReady(dictionary)).toBe(true);
    expect(isBackendLanguageLoading(dictionary)).toBe(true);
    expect(Object.keys(dictionary)).toEqual(["menu"]);
  });
});
