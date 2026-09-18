import { describe, it, expect } from "vitest";
import {
  isVatAccount,
  calculateVat7,
  analyzeGLTaxAndBalance,
  autoBalanceJournalLines,
  setExactVatLine,
  appendVatLine,
} from "./gl-smart-guard";
import { amountUnits, emptyAccount, emptyLine, type GLAccount, type GLLine } from "./general-ledger";

describe("gl-smart-guard (Thai Accounting Tax & Balance Guard)", () => {
  const sampleAccounts: GLAccount[] = [
    { ...emptyAccount(), accountcode: "1111", names: [{ code: "th", name: "เงินสด" }] },
    { ...emptyAccount(), accountcode: "1151", names: [{ code: "th", name: "ภาษีซื้อ" }] },
    { ...emptyAccount(), accountcode: "2111", names: [{ code: "th", name: "เจ้าหนี้การค้า" }] },
    { ...emptyAccount(), accountcode: "2141", names: [{ code: "th", name: "ภาษีขาย" }] },
    { ...emptyAccount(), accountcode: "4111", names: [{ code: "th", name: "รายได้จากการขาย" }] },
    { ...emptyAccount(), accountcode: "5111", names: [{ code: "th", name: "ค่าใช้จ่ายในการดำเนินงาน" }] },
  ];

  describe("isVatAccount", () => {
    it("detects input VAT from account code 1151", () => {
      const res = isVatAccount("1151", "ภาษีซื้อ", "");
      expect(res.isVat).toBe(true);
      expect(res.type).toBe("input_tax");
    });

    it("detects output VAT from account code 2141", () => {
      const res = isVatAccount("2141", "ภาษีขาย", "");
      expect(res.isVat).toBe(true);
      expect(res.type).toBe("output_tax");
    });

    it("detects VAT from description if account name is generic", () => {
      const res = isVatAccount("9999", "เบ็ดเตล็ด", "บันทึกภาษีซื้อ 7%");
      expect(res.isVat).toBe(true);
      expect(res.type).toBe("input_tax");
    });

    it("returns false for regular cash or expense accounts", () => {
      const res = isVatAccount("1111", "เงินสด", "เบิกเงินสดสำรอง");
      expect(res.isVat).toBe(false);
      expect(res.type).toBe(null);
    });
  });

  describe("calculateVat7", () => {
    it("computes 7% of 100.00 accurately as 7.00", () => {
      const base = amountUnits("100.00");
      const vat = calculateVat7(base, 2);
      expect(vat).toBe(amountUnits("7.00"));
    });

    it("computes 7% of 142.86 rounded to 10.00", () => {
      const base = amountUnits("142.86");
      const vat = calculateVat7(base, 2);
      expect(vat).toBe(amountUnits("10.00"));
    });

    it("computes 7% of 15.50 rounded to 1.09", () => {
      const base = amountUnits("15.50");
      const vat = calculateVat7(base, 2);
      expect(vat).toBe(amountUnits("1.09"));
    });

    it("returns 0 for 0 base", () => {
      expect(calculateVat7(0n)).toBe(0n);
    });
  });

  describe("analyzeGLTaxAndBalance", () => {
    it("reports balanced journal with exact 7% VAT", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "1151", debit: "70.00", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "1070.00" },
      ];

      const res = analyzeGLTaxAndBalance(lines, sampleAccounts, 2);
      expect(res.isBalanced).toBe(true);
      expect(res.balanceStatus).toBe("balanced");
      expect(res.differenceUnits).toBe(0n);
      expect(res.vat.hasVatLine).toBe(true);
      expect(res.vat.isExactVat).toBe(true);
      expect(res.vat.actualVatFormatted).toBe("70.00");
      expect(res.vat.expectedVatFormatted).toBe("70.00");
    });

    it("reports penny variance when supplier invoice has 1 satang rounding diff", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "1151", debit: "70.01", credit: "0" }, // 1 satang variance
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "1070.01" },
      ];

      const res = analyzeGLTaxAndBalance(lines, sampleAccounts, 2);
      expect(res.isBalanced).toBe(true);
      expect(res.vat.hasVatLine).toBe(true);
      expect(res.vat.isExactVat).toBe(false);
      expect(res.vat.isCloseVat).toBe(true); // Within 5 satang
      expect(res.vat.actualVatFormatted).toBe("70.01");
      expect(res.vat.expectedVatFormatted).toBe("70.00");
      expect(res.vat.varianceUnits).toBe(amountUnits("0.01"));
    });

    it("suggests 7% VAT when user inputs expense without VAT line", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "5000.00", credit: "0" },
      ];

      const res = analyzeGLTaxAndBalance(lines, sampleAccounts, 2);
      expect(res.vat.hasVatLine).toBe(false);
      expect(res.vat.suggestedVatType).toBe("input_tax");
      expect(res.vat.suggestedVatFormatted).toBe("350.00");
    });
  });

  describe("autoBalanceJournalLines", () => {
    it("balances debit surplus by adjusting the credit side", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "1151", debit: "70.00", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "1069.99" }, // Short by 0.01
      ];

      const balanced = autoBalanceJournalLines(lines, 2);
      const res = analyzeGLTaxAndBalance(balanced, sampleAccounts, 2);
      expect(res.isBalanced).toBe(true);
      expect(balanced[2].credit).toBe("1070.00");
    });

    it("balances credit surplus by adjusting the debit side", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "4111", debit: "0", credit: "1000.05" }, // Credit is 0.05 higher
      ];

      const balanced = autoBalanceJournalLines(lines, 2);
      const res = analyzeGLTaxAndBalance(balanced, sampleAccounts, 2);
      expect(res.isBalanced).toBe(true);
      expect(balanced[0].debit).toBe("1000.05");
    });
  });

  describe("setExactVatLine", () => {
    it("updates VAT line to exact expected amount", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "1151", debit: "70.02", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "1070.02" },
      ];

      const exactUnits = amountUnits("70.00");
      const updated = setExactVatLine(lines, 1, exactUnits, 2);
      expect(updated[1].debit).toBe("70.00");
    });
  });

  describe("appendVatLine", () => {
    it("appends input VAT line with account 1151", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "2000.00", credit: "0" },
      ];

      const vatUnits = amountUnits("140.00");
      const updated = appendVatLine(lines, sampleAccounts, "input_tax", vatUnits, 2);
      expect(updated.length).toBe(2);
      expect(updated[1].accountcode).toBe("1151");
      expect(updated[1].debit).toBe("140.00");
      expect(updated[1].credit).toBe("0");
      expect(updated[1].description).toContain("7%");
    });
  });
});
