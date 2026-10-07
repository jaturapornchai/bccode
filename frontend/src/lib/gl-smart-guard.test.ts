import { describe, it, expect } from "vitest";
import { analyzeGLBalance, autoBalanceJournalLines } from "./gl-smart-guard";
import { emptyLine, type GLLine } from "./general-ledger";

describe("gl-smart-guard (debit/credit balance guard)", () => {
  describe("analyzeGLBalance", () => {
    it("reports a balanced journal", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "1151", debit: "70.00", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "1070.00" },
      ];

      const res = analyzeGLBalance(lines, 2);
      expect(res.isBalanced).toBe(true);
      expect(res.balanceStatus).toBe("balanced");
      expect(res.differenceUnits).toBe(0n);
      expect(res.balanceDifferenceFormatted).toBe("0.00");
    });

    it("reports a debit surplus with the absolute difference", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "999.75" },
      ];

      const res = analyzeGLBalance(lines, 2);
      expect(res.isBalanced).toBe(false);
      expect(res.balanceStatus).toBe("debit_surplus");
      expect(res.balanceDifferenceFormatted).toBe("0.25");
    });

    it("reports a credit surplus with the absolute difference", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "4111", debit: "0", credit: "1000.05" },
      ];

      const res = analyzeGLBalance(lines, 2);
      expect(res.isBalanced).toBe(false);
      expect(res.balanceStatus).toBe("credit_surplus");
      expect(res.balanceDifferenceFormatted).toBe("0.05");
    });

    it("counts half-typed amount text as zero instead of throwing", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "12.", credit: "0" },
        { ...emptyLine(), accountcode: "2111", debit: "0", credit: "abc" },
      ];

      expect(() => analyzeGLBalance(lines, 2)).not.toThrow();
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
      const res = analyzeGLBalance(balanced, 2);
      expect(res.isBalanced).toBe(true);
      expect(balanced[2].credit).toBe("1070.00");
    });

    it("balances credit surplus by adjusting the debit side", () => {
      const lines: GLLine[] = [
        { ...emptyLine(), accountcode: "5111", debit: "1000.00", credit: "0" },
        { ...emptyLine(), accountcode: "4111", debit: "0", credit: "1000.05" }, // Credit is 0.05 higher
      ];

      const balanced = autoBalanceJournalLines(lines, 2);
      const res = analyzeGLBalance(balanced, 2);
      expect(res.isBalanced).toBe(true);
      expect(balanced[0].debit).toBe("1000.05");
    });
  });
});
