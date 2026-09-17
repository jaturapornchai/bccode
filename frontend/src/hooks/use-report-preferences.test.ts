import { describe, expect, it } from "vitest";
import {
  getReportFontSizeClass,
  getReportContrastClass,
  useReportPreferences,
} from "./use-report-preferences";

describe("useReportPreferences", () => {
  it("returns appropriate font size classes", () => {
    expect(getReportFontSizeClass("normal")).toContain("text-[0.95rem]");
    expect(getReportFontSizeClass("medium")).toContain("text-[1.05rem]");
    expect(getReportFontSizeClass("large")).toContain("text-[1.2rem]");
  });

  it("returns contrast classes when enabled", () => {
    expect(getReportContrastClass(false)).toBe("");
    expect(getReportContrastClass(true)).toContain("bc-high-contrast");
    expect(getReportContrastClass(true)).toContain("border-foreground");
  });

  it("exports useReportPreferences hook without crashing", () => {
    expect(typeof useReportPreferences).toBe("function");
  });
});
