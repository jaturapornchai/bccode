import { describe, it, expect } from "vitest";
import {
  getLastDayOfMonth,
  thaiBahtText,
  generateVatClosingJournal,
} from "./thai-vat-closing";

describe("Thai VAT Closing Engine & Utilities", () => {
  describe("getLastDayOfMonth", () => {
    it("returns correct last day for 31-day and 30-day months", () => {
      expect(getLastDayOfMonth(2026, 1)).toBe("2026-01-31");
      expect(getLastDayOfMonth(2026, 4)).toBe("2026-04-30");
      expect(getLastDayOfMonth(2026, 9)).toBe("2026-09-30");
      expect(getLastDayOfMonth(2026, 12)).toBe("2026-12-31");
    });

    it("handles February leap year vs non-leap year correctly", () => {
      // 2024 is a leap year (29 days)
      expect(getLastDayOfMonth(2024, 2)).toBe("2024-02-29");
      // 2026 is a regular year (28 days)
      expect(getLastDayOfMonth(2026, 2)).toBe("2026-02-28");
    });
  });

  describe("thaiBahtText", () => {
    it("converts numbers to Thai Baht text format accurately", () => {
      expect(thaiBahtText(0)).toBe("ศูนย์บาทถ้วน");
      expect(thaiBahtText(100)).toBe("หนึ่งร้อยบาทถ้วน");
      expect(thaiBahtText(101)).toBe("หนึ่งร้อยเอ็ดบาทถ้วน");
      expect(thaiBahtText(1250.50)).toBe("หนึ่งพันสองร้อยห้าสิบบาทห้าสิบสตางค์");
      expect(thaiBahtText(21.25)).toBe("ยี่สิบเอ็ดบาทยี่สิบห้าสตางค์");
      expect(thaiBahtText(1500000)).toBe("หนึ่งล้านห้าแสนบาทถ้วน");
    });
  });

  describe("generateVatClosingJournal", () => {
    it("generates balanced closing entry when output VAT exceeds input VAT (payable)", () => {
      const result = generateVatClosingJournal({
        year: 2026,
        month: 9,
        outputVat: 7000,
        inputVat: 4200,
      });

      expect(result.date).toBe("2026-09-30");
      expect(result.isBalanced).toBe(true);
      expect(result.netVatType).toBe("payable");
      expect(result.netAmount).toBe(2800);
      expect(result.totalDebit).toBe(7000);
      expect(result.totalCredit).toBe(7000);

      // Lines: Dr. 2141 (7000), Cr. 1151 (4200), Cr. 2142 (2800)
      const drOutput = result.lines.find((l) => l.accountCode === "2141");
      const crInput = result.lines.find((l) => l.accountCode === "1151");
      const crPayable = result.lines.find((l) => l.accountCode === "2142");

      expect(drOutput?.debit).toBe("7000.00");
      expect(crInput?.credit).toBe("4200.00");
      expect(crPayable?.credit).toBe("2800.00");
      expect(result.summaryNoteTh).toContain("สองพันแปดร้อยบาทถ้วน");
    });

    it("generates balanced closing entry when input VAT exceeds output VAT (refundable/credit)", () => {
      const result = generateVatClosingJournal({
        year: 2026,
        month: 8,
        outputVat: 5000,
        inputVat: 8500,
      });

      expect(result.isBalanced).toBe(true);
      expect(result.netVatType).toBe("refundable");
      expect(result.netAmount).toBe(3500);
      expect(result.totalDebit).toBe(8500);
      expect(result.totalCredit).toBe(8500);

      // Lines: Dr. 2141 (5000), Dr. 1152 (3500), Cr. 1151 (8500)
      const drOutput = result.lines.find((l) => l.accountCode === "2141");
      const drRefundable = result.lines.find((l) => l.accountCode === "1152");
      const crInput = result.lines.find((l) => l.accountCode === "1151");

      expect(drOutput?.debit).toBe("5000.00");
      expect(drRefundable?.debit).toBe("3500.00");
      expect(crInput?.credit).toBe("8500.00");
      expect(result.summaryNoteTh).toContain("สามพันห้าร้อยบาทถ้วน");
    });

    it("factors in credit brought forward from previous month", () => {
      const result = generateVatClosingJournal({
        year: 2026,
        month: 9,
        outputVat: 10000,
        inputVat: 6000,
        creditBroughtForward: 1500, // เครดิตยกมา 1,500 -> ภาษีที่ต้องชำระสุทธิเหลือ 2,500
      });

      expect(result.isBalanced).toBe(true);
      expect(result.netVatType).toBe("payable");
      expect(result.netAmount).toBe(2500);

      const crPayable = result.lines.find((l) => l.accountCode === "2142");
      expect(crPayable?.credit).toBe("2500.00");
    });
  });
});
