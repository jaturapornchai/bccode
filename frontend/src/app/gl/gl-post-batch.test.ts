import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { readFileSync } from "node:fs";
import { describe, expect, it, vi } from "vitest";
import type { GLPostCheckRow } from "@/lib/general-ledger";
import { EMPTY_POST_BATCH_FILTER, EMPTY_POST_BATCH_SIGNATORIES, GLPostBatch, POST_BATCH_PERSON_MAX_LENGTH, postBatchCreatePayload, postBatchFilterPayload, postBatchItems, postBatchShortcut, postBatchSourceText, postBatchSummary, readyRowIds, toggleAllReady } from "./gl-post-batch";

vi.mock("./gl-common", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./gl-common")>();
  return {
    ...actual,
    useReferences: vi.fn().mockReturnValue({ accounts: [], years: [], books: [{ id: "b1", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true }], error: "", reload: vi.fn() }),
    useGLCommand: vi.fn().mockReturnValue({ busy: false, execute: vi.fn() }),
  };
});

const row = (id: string, ok: boolean, debit = "0", extra: Partial<GLPostCheckRow> = {}): GLPostCheckRow => ({
  id, version: 2, docno: `JV-${id}`, date: "2026-01-15", bookcode: "JV", fiscalyear: "2569", branchcode: "00000",
  description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. เป็นเงินเชื่อ", debit, credit: debit, ok, ...extra,
});

describe("batch posting helpers (Champ 5512)", () => {
  it("Source column ('ที่มา'): the importing system, or 'keyed in the ledger' when there is none", () => {
    const tr = (_key: string, fallback: string) => fallback;
    expect(postBatchSourceText(row("a", true), tr)).toBe("คีย์ในบัญชีแยกประเภท");
    expect(postBatchSourceText(row("b", true, "0", { sourcesystem: "pos-branch" }), tr)).toBe("pos-branch");
    expect(postBatchSourceText(row("c", true, "0", { sourcesystem: "  " }), tr)).toBe("คีย์ในบัญชีแยกประเภท");
  });
  it("sends only the filled filters, trimmed", () => {
    expect(postBatchFilterPayload(EMPTY_POST_BATCH_FILTER)).toEqual({});
    expect(postBatchFilterPayload({ bookcode: " JV ", datefrom: "2026-01-01", dateto: "", docnofrom: "  ", docnoto: "JV0099 " }))
      .toEqual({ bookcode: "JV", datefrom: "2026-01-01", docnoto: "JV0099" });
  });

  it("sends the trimmed signatories with the batch; blank reviewers are left out, the approver is always sent", () => {
    const items = [{ id: "1", version: 2 }];
    expect(postBatchCreatePayload(items, { remark: "  ผ่านรายการประจำเดือนมกราคม 2569 ", check1: " นางสาวสุดา มั่นคง ", check2: "   ", approver: "  นายสมชาย ใจดี " }))
      .toEqual({ items, remark: "ผ่านรายการประจำเดือนมกราคม 2569", check1: "นางสาวสุดา มั่นคง", approver: "นายสมชาย ใจดี" });
    expect(postBatchCreatePayload(items, { remark: "", ...EMPTY_POST_BATCH_SIGNATORIES })).toEqual({ items, remark: "", approver: "" });
    expect(POST_BATCH_PERSON_MAX_LENGTH).toBe(50);
  });

  it("selects only rows that passed the check, and F2 toggles between all ready and none", () => {
    const rows = [row("1", true), row("2", false, "0", { code: "period_locked", message: "งวดถูกล็อก" }), row("3", true)];
    expect(readyRowIds(rows)).toEqual(["1", "3"]);
    const all = toggleAllReady(rows, new Set());
    expect([...all]).toEqual(["1", "3"]);
    expect(toggleAllReady(rows, all).size).toBe(0);
    // partly selected → select every ready row (never the failing one)
    expect([...toggleAllReady(rows, new Set(["3", "2"]))]).toEqual(["1", "3"]);
    expect(toggleAllReady([row("x", false)], new Set()).size).toBe(0);
  });

  it("posts selected ready rows in table order with the checked version", () => {
    const rows = [row("1", true), row("2", false), row("3", true, "5", { version: 7 })];
    expect(postBatchItems(rows, new Set(["3", "2", "1"]))).toEqual([{ id: "1", version: 2 }, { id: "3", version: 7 }]);
    expect(postBatchItems(rows, new Set())).toEqual([]);
  });

  it("totals the selected debit exactly without floating point", () => {
    const rows = [row("1", true, "0.1"), row("2", true, "0.2"), row("3", false, "999"), row("4", true, "12345678901234567890.05")];
    const summary = postBatchSummary(rows, new Set(["1", "2", "3"]));
    expect(summary).toEqual({ checked: 4, ready: 3, issues: 1, selected: 2, taxcount: 0, debit: "0.30" });
    expect(postBatchSummary(rows, new Set(["1", "4"])).debit).toBe("12345678901234567890.15");
    expect(postBatchSummary([row("1", true, "1,000")], new Set(["1"])).debit).toBe("");
  });

  it("sums the tax count (จำนวนภาษี) of the selected ready rows only; a missing count is 0", () => {
    const rows = [row("1", true, "1", { taxcount: 2 }), row("2", true, "1", { taxcount: 1 }), row("3", false, "1", { taxcount: 5 }), row("4", true, "1")];
    expect(postBatchSummary(rows, new Set(["1", "2", "3", "4"])).taxcount).toBe(3);
    expect(postBatchSummary(rows, new Set(["2"])).taxcount).toBe(1);
    expect(postBatchSummary(rows, new Set()).taxcount).toBe(0);
  });

  it("maps the Champ function keys", () => {
    expect(postBatchShortcut("F7")).toBe("check");
    expect(postBatchShortcut("F2")).toBe("select-all");
    expect(postBatchShortcut("F11")).toBe("post");
    expect(postBatchShortcut("F9")).toBeNull();
    expect(postBatchShortcut("Enter")).toBeNull();
  });

  it("renders Thai labels with keyboard hints and the reverse tab", () => {
    const html = renderToStaticMarkup(createElement(GLPostBatch, { route: "/gl/posting" }));
    for (const text of ["ผ่านรายการทีละชุด", "กลับรายการที่ผ่านแล้ว", "ทุกสมุดรายวัน", "จากวันที่", "ถึงวันที่", "จากเลขที่เอกสาร", "ถึงเลขที่เอกสาร", "ประมวลผล", "บันทึกรายการ", "ผู้ตรวจสอบ 1", "ผู้ตรวจสอบ 2", "ผู้อนุมัติ *", "หมายเหตุชุดผ่านรายการ", "(F7)", "(F2)", "(F11)"]) expect(html).toContain(text);
    expect(html).toContain("aria-required=\"true\"");
    expect(html.match(/maxLength="50"/gi)?.length ?? 0).toBeGreaterThanOrEqual(5); // doc no x2 + check1 + check2 + approver
    expect(html).toContain("role=\"tablist\"");
  });

  it("has no hard-coded Thai outside tr() fallbacks and comments", () => {
    const source = readFileSync(new URL("./gl-post-batch.tsx", import.meta.url), "utf8")
      .replace(/\/\*[\s\S]*?\*\//g, "").replace(/\/\/.*$/gm, "")
      .replace(/tr\("[a-z0-9_]+", "[^"]*"\)/g, "");
    expect(source).not.toMatch(/[฀-๿]/);
  });
});
