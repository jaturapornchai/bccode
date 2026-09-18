import { describe, it, expect } from "vitest";
import {
  THAI_BUSINESS_METADATA,
  THAI_JOURNAL_PATTERNS,
  calculateVatSatang,
  calculateWhtSatang,
  suggestJournalPatterns,
  type ThaiBusinessType,
} from "./thai-accounting-business-patterns";

describe("Thai Accounting Business Patterns (8 Industries)", () => {
  const allBusinessTypes: ThaiBusinessType[] = [
    "trading",
    "service",
    "manufacturing",
    "restaurant_cafe",
    "construction",
    "ecommerce",
    "real_estate_rental",
    "transport_logistics",
  ];

  it("covers all 8 business industries with metadata and accounts", () => {
    for (const bType of allBusinessTypes) {
      const meta = THAI_BUSINESS_METADATA[bType];
      expect(meta).toBeDefined();
      expect(meta.type).toBe(bType);
      expect(meta.titleTh.length).toBeGreaterThan(0);
      expect(meta.recommendedAccounts.length).toBeGreaterThanOrEqual(8);
      expect(meta.taxCharacteristicsTh.length).toBeGreaterThanOrEqual(2);
    }
  });

  it("contains valid journal pattern templates for each business type", () => {
    for (const bType of allBusinessTypes) {
      const patterns = THAI_JOURNAL_PATTERNS.filter((p) => p.businessType === bType);
      expect(patterns.length).toBeGreaterThanOrEqual(2);

      for (const pattern of patterns) {
        expect(pattern.keywords.length).toBeGreaterThanOrEqual(2);
        expect(pattern.lines.length).toBeGreaterThanOrEqual(2);
        expect(pattern.taxNotesTh.length).toBeGreaterThan(0);

        // ตรวจสอบว่ามีทั้งฝั่ง debit และ credit
        const hasDebit = pattern.lines.some((l) => l.side === "debit");
        const hasCredit = pattern.lines.some((l) => l.side === "credit");
        expect(hasDebit).toBe(true);
        expect(hasCredit).toBe(true);
      }
    }
  });

  describe("Tax Satang Calculation Utilities", () => {
    it("calculates 7% VAT in satang with standard rounding", () => {
      // 100 บาท (10,000 satang) * 7% = 7 บาท (700 satang)
      expect(calculateVatSatang(10000n, 7)).toBe(700n);

      // 100.50 บาท (10,050 satang) * 7% = 7.035 บาท -> ปัดเป็น 7.04 บาท (704 satang)
      expect(calculateVatSatang(10050n, 7)).toBe(704n);

      // 10.33 บาท (1033 satang) * 7% = 0.7231 บาท -> ปัดเป็น 0.72 บาท (72 satang)
      expect(calculateVatSatang(1033n, 7)).toBe(72n);
    });

    it("calculates WHT in satang accurately for 1%, 2%, 3%, 5%", () => {
      const baseSatang = 100000n; // 1,000 บาท

      // WHT 1% (ขนส่ง) = 10 บาท (1000 satang)
      expect(calculateWhtSatang(baseSatang, 1)).toBe(1000n);

      // WHT 2% (โฆษณา) = 20 บาท (2000 satang)
      expect(calculateWhtSatang(baseSatang, 2)).toBe(2000n);

      // WHT 3% (บริการ / จ้างทำของ) = 30 บาท (3000 satang)
      expect(calculateWhtSatang(baseSatang, 3)).toBe(3000n);

      // WHT 5% (ค่าเช่าอสังหาริมทรัพย์) = 50 บาท (5000 satang)
      expect(calculateWhtSatang(baseSatang, 5)).toBe(5000n);
    });
  });

  describe("Smart Journal Suggestion (Auto-Suggest)", () => {
    it("suggests credit purchase for trading when querying 'ซื้อสินค้า'", () => {
      const suggestions = suggestJournalPatterns("ซื้อสินค้า", "trading", 1000);
      expect(suggestions.length).toBeGreaterThan(0);
      const topMatch = suggestions[0];
      expect(topMatch.pattern.id).toBe("trading_purchase_credit");

      // ตรวจสอบยอดคำนวณ: ฐาน 1000, VAT 7% = 70, เจ้าหนี้รวม = 1070
      const vatLine = topMatch.calculatedLines.find((l) => l.accountCode === "1151");
      const apLine = topMatch.calculatedLines.find((l) => l.accountCode === "2111");
      expect(vatLine?.amount).toBe(70);
      expect(apLine?.amount).toBe(1070);
    });

    it("suggests service revenue with WHT 3% when querying 'ค่าที่ปรึกษา'", () => {
      const suggestions = suggestJournalPatterns("ค่าที่ปรึกษา", "service", 10000);
      expect(suggestions.length).toBeGreaterThan(0);
      const topMatch = suggestions[0];
      expect(topMatch.pattern.id).toBe("service_income_wht3");

      // WHT 3% ของ 10,000 = 300
      const whtLine = topMatch.calculatedLines.find((l) => l.accountCode === "1161");
      expect(whtLine?.amount).toBe(300);
    });

    it("suggests WIP transfer for manufacturing when querying 'เบิกผลิต'", () => {
      const suggestions = suggestJournalPatterns("เบิกผลิต", "manufacturing", 50000);
      expect(suggestions.length).toBeGreaterThan(0);
      const topMatch = suggestions[0];
      expect(topMatch.pattern.id).toBe("mfg_issue_material_wip");
      expect(topMatch.calculatedLines[0].accountCode).toBe("1143"); // งานระหว่างทำ
    });

    it("suggests delivery platform settlement when querying 'gp grab'", () => {
      const suggestions = suggestJournalPatterns("gp grab", "restaurant_cafe", 5000);
      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions[0].pattern.id).toBe("rest_delivery_settlement");
    });

    it("suggests construction progress billing when querying 'ค่างวดงาน'", () => {
      const suggestions = suggestJournalPatterns("ค่างวดงาน", "construction", 200000);
      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions[0].pattern.id).toBe("const_progress_billing");
    });

    it("suggests overseas ads spend (ภ.พ.36) when querying 'facebook ads'", () => {
      const suggestions = suggestJournalPatterns("facebook ads", "ecommerce", 15000);
      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions[0].pattern.id).toBe("ecom_facebook_ads_pp36");
    });

    it("suggests rental billing when querying 'ค่าเช่าตึก'", () => {
      const suggestions = suggestJournalPatterns("ค่าเช่าตึก", "real_estate_rental", 30000);
      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions[0].pattern.id).toBe("rental_monthly_billing");
    });

    it("suggests freight transport with 1% WHT when querying 'รับจ้างขนส่ง'", () => {
      const suggestions = suggestJournalPatterns("รับจ้างขนส่ง", "transport_logistics", 8000);
      expect(suggestions.length).toBeGreaterThan(0);
      const topMatch = suggestions[0];
      expect(topMatch.pattern.id).toBe("trans_freight_service");
      const wht1Line = topMatch.calculatedLines.find((l) => l.accountCode === "1163");
      expect(wht1Line?.amount).toBe(80); // 1% ของ 8,000 = 80 บาท
    });

    it("handles empty or unmatched query safely", () => {
      expect(suggestJournalPatterns("")).toEqual([]);
      expect(suggestJournalPatterns("   ")).toEqual([]);
      expect(suggestJournalPatterns("xyz999nonsenseabc")).toEqual([]);
    });
  });
});
