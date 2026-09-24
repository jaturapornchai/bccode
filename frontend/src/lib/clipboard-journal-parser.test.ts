import { describe, expect, it } from "vitest";
import { isHeaderRow, parseClipboardAmount, parseClipboardJournalLines } from "./clipboard-journal-parser";

const line = (accountcode: string, description: string, debit: string, credit: string, departmentcode = "", projectcode = "") => ({
  accountcode,
  description,
  debit,
  credit,
  departmentcode,
  projectcode,
  cashflow: "",
});

describe("clipboard-journal-parser", () => {
  describe("parseClipboardAmount", () => {
    it("strips commas, spaces, NBSP and currency signs", () => {
      expect(parseClipboardAmount("1,500.50")).toEqual({ value: "1500.50", negative: false, valid: true });
      expect(parseClipboardAmount("฿ 25,000.00")).toEqual({ value: "25000.00", negative: false, valid: true });
      expect(parseClipboardAmount("$1,200")).toEqual({ value: "1200", negative: false, valid: true });
      expect(parseClipboardAmount(" 1,070.00 ")).toEqual({ value: "1070.00", negative: false, valid: true });
    });

    it("reads blanks and Excel accounting dashes as zero \"0\" (never \"\")", () => {
      for (const cell of ["", "   ", "-", "–", "—", "0", "0.00", " - ", undefined]) {
        expect(parseClipboardAmount(cell)).toEqual({ value: "0", negative: false, valid: true });
      }
    });

    it("converts Thai digits", () => {
      expect(parseClipboardAmount("๑,๕๐๐.๐๐").value).toBe("1500.00");
    });

    it("flags accounting negatives explicitly instead of dropping them", () => {
      expect(parseClipboardAmount("(1,500.00)")).toEqual({ value: "1500.00", negative: true, valid: true });
      expect(parseClipboardAmount("-1500")).toEqual({ value: "1500", negative: true, valid: true });
    });

    it("marks text that is not an amount as invalid", () => {
      expect(parseClipboardAmount("หนึ่งพัน").valid).toBe(false);
      expect(parseClipboardAmount("1500-").valid).toBe(false);
      expect(parseClipboardAmount("12.34.56").valid).toBe(false);
      expect(parseClipboardAmount("(-100)").valid).toBe(false);
    });
  });

  describe("isHeaderRow", () => {
    it("detects Thai and English header titles", () => {
      expect(isHeaderRow(["รหัสบัญชี", "คำอธิบาย", "เดบิต", "เครดิต"])).toBe(true);
      expect(isHeaderRow(["Account", "Description", "Debit", "Credit"])).toBe(true);
      expect(isHeaderRow(["11110", "เงินสดในมือ", "5000", "0"])).toBe(false);
    });
  });

  describe("parseClipboardJournalLines — real Excel copies", () => {
    it("keeps the debit and the description when the credit cell is empty (trailing tab from Excel)", () => {
      // Excel copy of 2 rows × 4 columns: row 1 has an empty credit cell, row 2 an empty debit cell
      const excel = "11110\tรับเงินสดขายหน้าร้าน\t1,070.00\t\r\n41100\tขายวัสดุก่อสร้าง\t\t1,070.00\r\n";
      const { lines, issues } = parseClipboardJournalLines(excel);
      expect(issues).toEqual([]);
      expect(lines).toEqual([
        line("11110", "รับเงินสดขายหน้าร้าน", "1070.00", "0"),
        line("41100", "ขายวัสดุก่อสร้าง", "0", "1070.00"),
      ]);
    });

    it("skips the header row, keeps department/project and writes the zero side as \"0\"", () => {
      const excel =
        "รหัสบัญชี\tคำอธิบาย\tเดบิต\tเครดิต\tแผนก\tโครงการ\n" +
        "11310\tซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. 500 ถุง\t65,000.00\t0.00\tคลัง\tA-2569\n" +
        "11410\tภาษีซื้อ 7%\t4,550.00\t-\tคลัง\tA-2569\n" +
        "21100\tเจ้าหนี้การค้า บจ. สยามซีเมนต์\t\t69,550.00\tคลัง\tA-2569\n";
      const { lines, issues } = parseClipboardJournalLines(excel);
      expect(issues).toEqual([]);
      expect(lines).toEqual([
        line("11310", "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. 500 ถุง", "65000.00", "0", "คลัง", "A-2569"),
        line("11410", "ภาษีซื้อ 7%", "4550.00", "0", "คลัง", "A-2569"),
        line("21100", "เจ้าหนี้การค้า บจ. สยามซีเมนต์", "0", "69550.00", "คลัง", "A-2569"),
      ]);
    });

    it("rejects an accounting negative in a debit/credit column with the row number (never drops it)", () => {
      const excel = "11110\tเงินสด\t(1,500.00)\t0\n21100\tเจ้าหนี้การค้า\t\t1,500.00";
      const { issues } = parseClipboardJournalLines(excel);
      expect(issues).toEqual([{ row: 1, field: "debit", value: "(1,500.00)", problem: "negative" }]);
    });

    it("reports a non-numeric amount cell with its row number (header row counted)", () => {
      const excel = "รหัสบัญชี\tคำอธิบาย\tเดบิต\tเครดิต\n11110\tเงินสด\t1,500.00\t\n21100\tเจ้าหนี้\t\tหนึ่งพันห้าร้อย";
      const { issues } = parseClipboardJournalLines(excel);
      expect(issues).toEqual([{ row: 3, field: "credit", value: "หนึ่งพันห้าร้อย", problem: "invalid" }]);
    });

    it("2-column signed amounts: positive → debit, negative or (x) → credit (documented rule)", () => {
      const { lines, issues } = parseClipboardJournalLines("51200\t18,000.00\n11120\t(18,000.00)\n21300\t-500");
      expect(issues).toEqual([]);
      expect(lines.map((l) => [l.accountcode, l.debit, l.credit])).toEqual([
        ["51200", "18000.00", "0"],
        ["11120", "0", "18000.00"],
        ["21300", "0", "500"],
      ]);
    });

    it("3 columns: [account][debit][credit] when the 2nd cell is a number, else [account][description][signed]", () => {
      const numeric = parseClipboardJournalLines("11110\t5000\t0\n11120\t\t5000");
      expect(numeric.lines.map((l) => [l.accountcode, l.debit, l.credit])).toEqual([
        ["11110", "5000", "0"],
        ["11120", "0", "5000"],
      ]);
      const described = parseClipboardJournalLines("51300\tค่าไฟฟ้าเดือนมกราคม\t3,210.00\n11110\tจ่ายค่าไฟฟ้า\t(3,210.00)");
      expect(described.lines).toEqual([
        line("51300", "ค่าไฟฟ้าเดือนมกราคม", "3210.00", "0"),
        line("11110", "จ่ายค่าไฟฟ้า", "0", "3210.00"),
      ]);
    });

    it("reads quoted CSV cells with thousands commas as one amount", () => {
      const { lines, issues } = parseClipboardJournalLines('11110,"เงินสด, หน้าร้าน","1,500.00",0');
      expect(issues).toEqual([]);
      expect(lines).toEqual([line("11110", "เงินสด, หน้าร้าน", "1500.00", "0")]);
    });

    it("keeps a first data row whose description mentions a header word", () => {
      const { lines } = parseClipboardJournalLines("11210\tปรับปรุงรหัสลูกค้า credit note\t0\t2,000.00");
      expect(lines).toEqual([line("11210", "ปรับปรุงรหัสลูกค้า credit note", "0", "2000.00")]);
    });

    it("returns nothing for empty text or blank rows", () => {
      expect(parseClipboardJournalLines("")).toEqual({ lines: [], issues: [] });
      expect(parseClipboardJournalLines("   \n\t\t\t\n")).toEqual({ lines: [], issues: [] });
    });
  });
});
