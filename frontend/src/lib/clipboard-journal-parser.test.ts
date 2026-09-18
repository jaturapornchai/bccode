import { describe, expect, it } from "vitest";
import {
  cleanNumericAmount,
  isHeaderRow,
  parseClipboardJournalLines,
} from "./clipboard-journal-parser";

describe("clipboard-journal-parser", () => {
  describe("cleanNumericAmount", () => {
    it("strips commas, spaces, and currency signs", () => {
      expect(cleanNumericAmount("1,500.50")).toBe("1500.50");
      expect(cleanNumericAmount("฿ 25,000.00")).toBe("25000.00");
      expect(cleanNumericAmount("$1,200")).toBe("1200");
    });

    it("returns empty string for dashes and empty inputs", () => {
      expect(cleanNumericAmount("-")).toBe("");
      expect(cleanNumericAmount("—")).toBe("");
      expect(cleanNumericAmount("")).toBe("");
      expect(cleanNumericAmount(undefined)).toBe("");
    });
  });

  describe("isHeaderRow", () => {
    it("detects Thai and English header titles", () => {
      expect(isHeaderRow(["รหัสบัญชี", "คำอธิบาย", "เดบิต", "เครดิต"])).toBe(true);
      expect(isHeaderRow(["Account", "Description", "Debit", "Credit"])).toBe(true);
      expect(isHeaderRow(["1111-01", "เงินสด", "5000", "0"])).toBe(false);
    });
  });

  describe("parseClipboardJournalLines", () => {
    it("parses 4-column tab-separated Excel copy data and skips header", () => {
      const excelClipboard =
        "รหัสบัญชี\tคำอธิบาย\tเดบิต\tเครดิต\n" +
        "1111-01\tเงินสดในมือ\t15,000.00\t0.00\n" +
        "1112-01\tเงินฝากกระแสรายวัน\t0.00\t15,000.00";

      const lines = parseClipboardJournalLines(excelClipboard);
      expect(lines).toHaveLength(2);
      expect(lines[0]).toEqual({
        accountcode: "1111-01",
        description: "เงินสดในมือ",
        debit: "15000.00",
        credit: "",
        departmentcode: "",
        projectcode: "",
        cashflow: "",
      });
      expect(lines[1]).toEqual({
        accountcode: "1112-01",
        description: "เงินฝากกระแสรายวัน",
        debit: "",
        credit: "15000.00",
        departmentcode: "",
        projectcode: "",
        cashflow: "",
      });
    });

    it("parses 3-column data (Account, Debit, Credit)", () => {
      const pasteData = "1111-01\t5000\t0\n1112-01\t0\t5000";
      const lines = parseClipboardJournalLines(pasteData);
      expect(lines).toHaveLength(2);
      expect(lines[0].accountcode).toBe("1111-01");
      expect(lines[0].debit).toBe("5000");
      expect(lines[1].credit).toBe("5000");
    });

    it("parses 5-column data with department code", () => {
      const pasteData = "5111-01\tค่าใช้จ่ายสำนักงาน\t2500\t0\tDEPT-ACC";
      const lines = parseClipboardJournalLines(pasteData);
      expect(lines).toHaveLength(1);
      expect(lines[0].accountcode).toBe("5111-01");
      expect(lines[0].departmentcode).toBe("DEPT-ACC");
    });

    it("returns empty array for invalid or empty text", () => {
      expect(parseClipboardJournalLines("")).toEqual([]);
      expect(parseClipboardJournalLines("   ")).toEqual([]);
    });
  });
});
