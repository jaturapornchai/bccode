import { describe, expect, it } from "vitest";
import { isTabularPaste, parseClipboardAmount, parseClipboardJournalLines, thousandsCommasValid } from "./clipboard-journal-parser";

// adversarial review 2026-09-24: one Excel cell, European amounts and non-canonical numbers
describe("isTabularPaste — only a block of cells becomes journal lines", () => {
  it("lets one Excel cell (text + CRLF) paste into the focused box", () => {
    expect(isTabularPaste("1,500.00\r\n")).toBe(false);
    expect(isTabularPaste("12,345,678.90\r\n")).toBe(false);
    expect(isTabularPaste("ค่าน้ำ, ค่าไฟ\r\n")).toBe(false);
    expect(isTabularPaste("1500\n")).toBe(false);
    expect(isTabularPaste("1500")).toBe(false);
  });
  it("takes over a row with tabs or two or more lines", () => {
    expect(isTabularPaste("51300\tค่าไฟฟ้า\t3,210.00\t\r\n")).toBe(true);
    expect(isTabularPaste("51300,3210\r\n11110,-3210\r\n")).toBe(true);
  });
});

describe("parseClipboardJournalLines — single-column amounts never become accounts", () => {
  it("keeps a thousands-comma number in one cell (no account 12 / debit 345)", () => {
    expect(parseClipboardJournalLines("12,345,678.90\r\n1,500.00\r\n").lines).toEqual([]);
  });
  it("still reads comma CSV journal rows", () => {
    expect(parseClipboardJournalLines("51300,ค่าไฟฟ้า,3210,0\n11110,จ่ายค่าไฟฟ้า,0,3210").lines.map((line) => [line.accountcode, line.debit, line.credit])).toEqual([
      ["51300", "3210", "0"],
      ["11110", "0", "3210"],
    ]);
  });
});

describe("parseClipboardAmount — value never changes silently", () => {
  it("rejects commas that are not thousands separators (European format)", () => {
    expect(parseClipboardAmount("1.500,00").valid).toBe(false);
    expect(parseClipboardAmount("1 500,00").valid).toBe(false);
    expect(parseClipboardAmount("1,5").valid).toBe(false);
    expect(parseClipboardJournalLines("1110\tค่าวัสดุ\t1.500,00\t").issues).toEqual([{ row: 1, field: "debit", value: "1.500,00", problem: "invalid" }]);
  });
  it("returns canonical numbers that save-time validation accepts", () => {
    expect(parseClipboardAmount(".5")).toEqual({ value: "0.5", negative: false, valid: true });
    expect(parseClipboardAmount("0500")).toEqual({ value: "500", negative: false, valid: true });
    expect(parseClipboardAmount("00.50")).toEqual({ value: "0.50", negative: false, valid: true });
    expect(parseClipboardAmount("1500.")).toEqual({ value: "1500", negative: false, valid: true });
  });
  it("knows thousands commas", () => {
    expect(thousandsCommasValid("1,234,567.89")).toBe(true);
    expect(thousandsCommasValid("1500")).toBe(true);
    expect(thousandsCommasValid("1.500,00")).toBe(false);
    expect(thousandsCommasValid("12,34")).toBe(false);
  });
});
