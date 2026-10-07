import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { readFileSync } from "node:fs";
import { describe, expect, it, vi } from "vitest";
import type { GLFiscalYear, GLMaster, GLPeriodProfit } from "@/lib/general-ledger";
import { GLPeriods, defaultPeriodYear, periodProfitText, periodReasonMissing, periodShortcut, periodSummary, yearPeriods } from "./gl-periods";

const year = (code: string, startdate: string, enddate: string, extra: Partial<GLFiscalYear> = {}): GLFiscalYear => ({
  id: code, code, startdate, enddate, scale: 2, retainedearningsaccount: "", profitlossaccount: "", isactive: true, closed: false, ...extra,
});
const years = [year("2568", "2025-01-01", "2025-12-31"), year("2569", "2026-01-01", "2026-12-31"), year("2570", "2027-01-01", "2027-12-31", { isdeleted: true })];

vi.mock("./gl-common", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./gl-common")>();
  return {
    ...actual,
    useReferences: vi.fn().mockReturnValue({ accounts: [], years: [], books: [], error: "", reload: vi.fn() }),
    useGLCommand: vi.fn().mockReturnValue({ busy: false, execute: vi.fn() }),
  };
});

const period = (code: string, fiscalyear: string, startdate: string, extra: Partial<GLMaster> = {}) => ({ id: code, code, fiscalyear, startdate, enddate: startdate, locked: false, ...extra }) as GLMaster;

describe("period screen helpers (Champ 5511)", () => {
  it("opens on the fiscal year that contains today, else the latest live year", () => {
    expect(defaultPeriodYear(years, "2026-06-15")).toBe("2569");
    expect(defaultPeriodYear(years, "2025-12-31")).toBe("2568");
    // 2027 belongs to a deleted year → fall back to the latest live year
    expect(defaultPeriodYear(years, "2027-03-01")).toBe("2569");
    expect(defaultPeriodYear([], "2026-01-01")).toBe("");
  });

  it("lists the selected year's live periods in date order and counts locks", () => {
    const rows = yearPeriods([
      period("2569-02", "2569", "2026-02-01", { locked: true }),
      period("2569-01", "2569", "2026-01-01", { locked: true }),
      period("2568-12", "2568", "2025-12-01"),
      period("2569-03", "2569", "2026-03-01", { isdeleted: true }),
      period("2569-04", "2569", "2026-04-01"),
    ], "2569");
    expect(rows.map((item) => item.code)).toEqual(["2569-01", "2569-02", "2569-04"]);
    expect(periodSummary(rows)).toEqual({ total: 3, locked: 2 });
  });

  it("shows profit as positive and loss as negative at the year scale, or a dash before the amounts load", () => {
    const profits: Record<string, GLPeriodProfit> = {
      "2569-01": { code: "2569-01", netprofitloss: "750", summprofitloss: "1234567.5" },
      "2569-02": { code: "2569-02", netprofitloss: "-300", summprofitloss: "0" },
    };
    expect(periodProfitText(profits, "2569-01", "netprofitloss", 2)).toBe("750.00");
    expect(periodProfitText(profits, "2569-01", "summprofitloss", 2)).toBe("1,234,567.50");
    expect(periodProfitText(profits, "2569-02", "netprofitloss", 2)).toBe("-300.00");
    expect(periodProfitText(profits, "2569-02", "summprofitloss", 0)).toBe("0");
    expect(periodProfitText(profits, "2569-03", "netprofitloss", 2)).toBe("-");
    expect(periodProfitText({}, "2569-01", "summprofitloss", 2)).toBe("-");
  });

  it("requires a reason before lock/unlock and maps F9 to generate", () => {
    expect(periodReasonMissing("")).toBe(true);
    expect(periodReasonMissing("   ")).toBe(true);
    expect(periodReasonMissing("ปิดงบเดือนมกราคม 2569")).toBe(false);
    expect(periodShortcut("F9")).toBe("generate");
    expect(periodShortcut("F7")).toBeNull();
  });

  it("renders the generate action with its F9 hint and the reason field", () => {
    const html = renderToStaticMarkup(createElement(GLPeriods));
    expect(html).toContain("สร้างงวดอัตโนมัติ");
    expect(html).toContain("(F9)");
    expect(html).toContain("เหตุผล");
    expect(html).toContain("กรุณาเลือกปีบัญชีก่อน");
  });

  it("has no hard-coded Thai outside tr() fallbacks and comments", () => {
    const source = readFileSync(new URL("./gl-periods.tsx", import.meta.url), "utf8")
      .replace(/\/\*[\s\S]*?\*\//g, "").replace(/\/\/.*$/gm, "")
      .replace(/tr\("[a-z0-9_]+", "[^"]*"\)/g, "");
    expect(source).not.toMatch(/[฀-๿]/);
  });
});
