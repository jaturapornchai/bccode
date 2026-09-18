import { describe, expect, it } from "vitest";
import {
  THAI_JOURNAL_PATTERNS,
  calculatePatternJournalLines,
  suggestJournalPatterns,
} from "./thai-accounting-business-patterns";

describe("Thai Accounting Patterns & Balancing Calculator", () => {
  it("ensures every pattern in THAI_JOURNAL_PATTERNS produces balanced debit and credit entries", () => {
    const testBaseAmounts = [1000, 10700, 35420.50, 100000];

    for (const pattern of THAI_JOURNAL_PATTERNS) {
      for (const baseAmount of testBaseAmounts) {
        const lines = calculatePatternJournalLines(pattern, baseAmount);

        let totalDebit = 0;
        let totalCredit = 0;

        for (const line of lines) {
          if (line.side === "debit") totalDebit += Math.round(line.amount * 100);
          else totalCredit += Math.round(line.amount * 100);
        }

        expect(
          totalDebit,
          `Pattern "${pattern.titleTh}" (${pattern.id}) must have balanced debit and credit for base ${baseAmount}`
        ).toBe(totalCredit);
      }
    }
  });

  it("calculates service income with WHT 3% and VAT 7% correctly", () => {
    const servicePattern = THAI_JOURNAL_PATTERNS.find((p) => p.id === "service_income_wht3")!;
    expect(servicePattern).toBeDefined();

    // Base 10,000 THB:
    // Revenue (Credit) = 10,000
    // Output VAT 7% (Credit) = 700
    // WHT 3% (Debit) = 300
    // Net Bank (Debit) = 10,400
    const lines = calculatePatternJournalLines(servicePattern, 10000);

    const bankLine = lines.find((l) => l.accountCode === "1112");
    const whtLine = lines.find((l) => l.accountCode === "1161");
    const revLine = lines.find((l) => l.accountCode === "4121");
    const vatLine = lines.find((l) => l.accountCode === "2141");

    expect(revLine?.amount).toBe(10000);
    expect(revLine?.side).toBe("credit");
    expect(vatLine?.amount).toBe(700);
    expect(vatLine?.side).toBe("credit");
    expect(whtLine?.amount).toBe(300);
    expect(whtLine?.side).toBe("debit");
    expect(bankLine?.amount).toBe(10400);
    expect(bankLine?.side).toBe("debit");
  });

  it("calculates office rent with WHT 5% (VAT exempt) correctly", () => {
    const rentPattern = THAI_JOURNAL_PATTERNS.find((p) => p.id === "service_office_rent")!;
    expect(rentPattern).toBeDefined();

    // Base 20,000 THB:
    // Expense (Debit) = 20,000
    // WHT 5% (Credit) = 1,000
    // Net Bank (Credit) = 19,000
    const lines = calculatePatternJournalLines(rentPattern, 20000);

    const rentLine = lines.find((l) => l.accountCode === "5321");
    const whtLine = lines.find((l) => l.accountCode === "2151");
    const bankLine = lines.find((l) => l.accountCode === "1112");

    expect(rentLine?.amount).toBe(20000);
    expect(rentLine?.side).toBe("debit");
    expect(whtLine?.amount).toBe(1000);
    expect(whtLine?.side).toBe("credit");
    expect(bankLine?.amount).toBe(19000);
    expect(bankLine?.side).toBe("credit");
  });
});
