import { describe, expect, it, vi } from "vitest";
import type { GLDimension, GLPage } from "@/lib/general-ledger";
import { dimensionConditionLines, dimensionPrintRows } from "./gl-dimensions";

const tr = (_key: string, fallback: string) => fallback;
const row = (code: string): GLDimension => ({ id: `id-${code}`, code, name: `ฝ่าย ${code}`, isactive: true });
const page = (items: GLDimension[], total = items.length): GLPage<GLDimension> => ({ items, total, page: 1, limit: 500, sequence: 1 });

describe("dimensionPrintRows (Champ 5538 / 5521 code range)", () => {
  it("sends the range to the backend, trimmed; the browser neither filters nor sorts", async () => {
    const request = vi.fn(async () => page([row("D03"), row("D02")]));
    const result = await dimensionPrintRows("departments", " D02 ", " D03 ", request as never);
    expect(request).toHaveBeenCalledTimes(1);
    expect(new URL(`http://x/${(request.mock.calls[0] as unknown as [string])[0]}`).searchParams.toString()).toBe("page=1&limit=500&codefrom=D02&codeto=D03");
    expect(result).toEqual({ rows: [row("D03"), row("D02")], complete: true });
  });
  it("sends no range for the whole register and ignores 'to' without 'from' like Champ", async () => {
    const request = vi.fn(async () => page([row("P01")]));
    await dimensionPrintRows("projects", "", "P09", request as never);
    expect((request.mock.calls[0] as unknown as [string])[0]).toBe("projects?page=1&limit=500");
  });
  it("sends 'from' alone as the exact-code filter", async () => {
    const request = vi.fn(async () => page([row("P01")]));
    await dimensionPrintRows("projects", "P01", "", request as never);
    expect((request.mock.calls[0] as unknown as [string])[0]).toBe("projects?page=1&limit=500&codefrom=P01");
  });
  it("reads every page until the reported total and flags a short read", async () => {
    const pages = [page([row("D01"), row("D02")], 3), page([row("D03")], 3)];
    const request = vi.fn(async () => pages.shift()!);
    expect(await dimensionPrintRows("departments", "", "", request as never)).toEqual({ rows: [row("D01"), row("D02"), row("D03")], complete: true });
    const stalled = vi.fn(async () => page([], 5));
    expect(await dimensionPrintRows("departments", "", "", stalled as never)).toEqual({ rows: [], complete: false });
    expect(stalled).toHaveBeenCalledTimes(1);
  });
});

describe("dimensionConditionLines", () => {
  it("is empty for the whole register, 'equals' for one code and a range for from + to", () => {
    expect(dimensionConditionLines("departments", "", "", tr)).toEqual([]);
    expect(dimensionConditionLines("departments", "", "D09", tr)).toEqual([]);
    expect(dimensionConditionLines("departments", "D01", "", tr)).toEqual(["จากแผนก เท่ากับ D01"]);
    expect(dimensionConditionLines("projects", " P01 ", " P05 ", tr)).toEqual(["จากโครงการ P01 ถึง P05"]);
  });
});
