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
    it("detects input VAT from the account name", () => {
      const res = isVatAccount("ภาษีซื้อ", "", "asset");
      expect(res.isVat).toBe(true);
      expect(res.type).toBe("input_tax");
    });

    it("detects output VAT from the account name", () => {
      const res = isVatAccount("ภาษีขาย", "", "liability");
      expect(res.isVat).toBe(true);
      expect(res.type).toBe("output_tax");
    });

    it("detects VAT from description if account name is generic", () => {
      const res = isVatAccount("เบ็ดเตล็ด", "บันทึกภาษีซื้อ 7%");
      expect(res.isVat).toBe(true);
      expect(res.type).toBe("input_tax");
    });

    it("returns false for regular cash or expense accounts", () => {
      const res = isVatAccount("เงินสด", "เบิกเงินสดสำรอง", "asset");
      expect(res.isVat).toBe(false);
      expect(res.type).toBe(null);
    });

    // UAT 2026-09-24: 1153 "ภาษีเงินได้ถูกหัก ณ ที่จ่าย" was flagged as input VAT because its code starts with 115
    it("never treats withholding-tax accounts as VAT", () => {
      expect(isVatAccount("ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)", "", "asset").isVat).toBe(false);
      expect(isVatAccount("ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53", "ภาษีขาย", "liability").isVat).toBe(false);
    });

    it("never treats income/expense accounts as the VAT line even when the name mentions VAT", () => {
      expect(isVatAccount("รายได้จากการให้บริการ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม", "", "income").isVat).toBe(false);
      expect(isVatAccount("รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%", "", "income").isVat).toBe(false);
    });

    it("types a generic VAT account by account type, not by code", () => {
      expect(isVatAccount("ภาษีมูลค่าเพิ่ม", "", "asset").type).toBe("input_tax");
      expect(isVatAccount("ภาษีมูลค่าเพิ่ม", "", "liability").type).toBe("output_tax");
      expect(isVatAccount("Private fund", "", "asset").isVat).toBe(false);
    });
  });

  describe("account codes are never a condition", () => {
    const chart: GLAccount[] = [
      { ...emptyAccount(), accountcode: "1111", names: [{ code: "th", name: "เงินสดในมือ" }] },
      { ...emptyAccount(), accountcode: "1153", names: [{ code: "th", name: "ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)" }] },
      { ...emptyAccount(), accountcode: "4122", accounttype: "income", normalbalance: "credit", names: [{ code: "th", name: "รายได้จากการให้บริการ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม" }] },
    ];

    it("does not flag a withheld-tax receipt as a VAT mismatch", () => {
      const res = analyzeGLTaxAndBalance([
        { ...emptyLine(), accountcode: "1111", debit: "9700.00", credit: "0" },
        { ...emptyLine(), accountcode: "1153", debit: "300.00", credit: "0" },
        { ...emptyLine(), accountcode: "4122", debit: "0", credit: "10000.00" },
      ], chart);
      expect(res.isBalanced).toBe(true);
      expect(res.vat.hasVatLine).toBe(false);
    });

    it("recognises VAT by name in a chart that uses unusual codes", () => {
      const other: GLAccount[] = [
        { ...emptyAccount(), accountcode: "9001", names: [{ code: "th", name: "ภาษีซื้อ" }] },
        { ...emptyAccount(), accountcode: "7000", accounttype: "expense", names: [{ code: "th", name: "ค่าวัสดุสิ้นเปลือง" }] },
      ];
      const res = analyzeGLTaxAndBalance([
        { ...emptyLine(), accountcode: "7000", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "9001", debit: "70.00", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "1070.00" },
      ], other);
      expect(res.vat.hasVatLine).toBe(true);
      expect(res.vat.vatAccountCode).toBe("9001");
      expect(res.vat.isExactVat).toBe(true);
    });

    it("leaves the account empty when the chart has no VAT account (no hard-coded fallback code)", () => {
      const updated = appendVatLine([{ ...emptyLine(), accountcode: "1111", debit: "100.00", credit: "0" }], chart, "input_tax", amountUnits("7.00"), 2);
      expect(updated[1].accountcode).toBe("");
      expect(updated[1].debit).toBe("7.00");
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

    const line: GLLine = { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" };

    it("skips an inactive VAT account", () => {
      const chart: GLAccount[] = [
        { ...emptyAccount(), accountcode: "8801", isactive: false, names: [{ code: "th", name: "ภาษีซื้อ" }] },
        { ...emptyAccount(), accountcode: "8802", names: [{ code: "th", name: "ภาษีซื้อ" }] },
      ];
      expect(appendVatLine([line], chart, "input_tax", amountUnits("70.00"), 2)[1].accountcode).toBe("8802");
    });

    it("prefers the normal VAT account over an undue-VAT account listed first", () => {
      const chart: GLAccount[] = [
        { ...emptyAccount(), accountcode: "8801", names: [{ code: "th", name: "ภาษีซื้อยังไม่ถึงกำหนด" }] },
        { ...emptyAccount(), accountcode: "8802", names: [{ code: "th", name: "ภาษีซื้อ" }] },
        { ...emptyAccount(), accountcode: "8803", accounttype: "liability", names: [{ code: "th", name: "ภาษีขายยังไม่ถึงกำหนด" }] },
        { ...emptyAccount(), accountcode: "8804", accounttype: "liability", names: [{ code: "th", name: "ภาษีขาย" }] },
      ];
      expect(appendVatLine([line], chart, "input_tax", amountUnits("70.00"), 2)[1].accountcode).toBe("8802");
      expect(appendVatLine([line], chart, "output_tax", amountUnits("70.00"), 2)[1].accountcode).toBe("8804");
    });

    it("leaves the account empty when no active posting VAT account matches", () => {
      const chart: GLAccount[] = [
        { ...emptyAccount(), accountcode: "8801", isactive: false, names: [{ code: "th", name: "ภาษีซื้อ" }] },
        { ...emptyAccount(), accountcode: "8802", allowposting: false, names: [{ code: "th", name: "ภาษีซื้อ" }] },
      ];
      const updated = appendVatLine([line], chart, "input_tax", amountUnits("70.00"), 2);
      expect(updated[1].accountcode).toBe("");
      expect(updated[1].debit).toBe("70.00");
    });
  });
});
