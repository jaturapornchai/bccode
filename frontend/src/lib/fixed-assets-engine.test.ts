import { describe, it, expect } from "vitest";
import {
  calculateStraightLineDepreciation,
  generatePeriodicDepreciationJournalLines,
  calculateAssetDisposal,
  THAI_ASSET_CATEGORIES,
  isLeapYear,
} from "./fixed-assets-engine";

describe("Fixed Assets Engine (Group B)", () => {
  it("provides Thai Revenue Department standard categories and limits", () => {
    const car = THAI_ASSET_CATEGORIES.find((c) => c.categoryCode === "VEHICLE_PASSENGER");
    expect(car).toBeDefined();
    expect(car?.taxCostLimit).toBe(1000000);
    expect(car?.standardDeprecPercent).toBe(20.0);

    const computer = THAI_ASSET_CATEGORIES.find((c) => c.categoryCode === "COMPUTER");
    expect(computer?.standardUsefulLifeYears).toBe(3);
  });

  it("calculates full-year straight line depreciation with 1 THB scrap value", () => {
    const result = calculateStraightLineDepreciation({
      cost: 100001,
      scrapValue: 1,
      usefulLifeYears: 5,
      purchaseDate: "2024-01-01",
      fiscalYear: 2026, // full year
    });

    expect(result.depreciableCost).toBe(100000);
    expect(result.fullYearDeprec).toBe(20000);
    expect(result.periodDeprec).toBe(20000);
  });

  it("calculates pro-rata daily depreciation for mid-year acquisition", () => {
    // Purchased July 1, 2026 (non-leap year: 365 days)
    // Days from July 1 to Dec 31: 184 days
    const result = calculateStraightLineDepreciation({
      cost: 365001,
      scrapValue: 1,
      usefulLifeYears: 5,
      purchaseDate: "2026-07-01",
      fiscalYear: 2026,
    });

    expect(result.depreciableCost).toBe(365000);
    expect(result.fullYearDeprec).toBe(73000); // 365000 / 5 = 73000
    expect(result.actualDaysInYear).toBe(184);
    // 73000 * 184 / 365 = 36800.00
    expect(result.periodDeprec).toBe(36800);
    expect(result.netBookValue).toBe(365001 - 36800);
  });

  it("generates balanced periodic depreciation journal lines", () => {
    const items = [
      {
        assetCode: "FA-001",
        assetName: "รถกระบะขนส่ง",
        expenseAccountCode: "520101",
        accumAccountCode: "129101",
        periodDeprec: 15000,
      },
      {
        assetCode: "FA-002",
        assetName: "คอมพิวเตอร์สำนักงาน",
        expenseAccountCode: "520102",
        accumAccountCode: "129102",
        periodDeprec: 8500,
      },
    ];

    const result = generatePeriodicDepreciationJournalLines(items);
    expect(result.totalDeprec).toBe(23500);
    expect(result.lines).toHaveLength(4);

    const sumDr = result.lines.reduce((s, l) => s + l.debit, 0);
    const sumCr = result.lines.reduce((s, l) => s + l.credit, 0);
    expect(sumDr).toBe(sumCr);
    expect(sumDr).toBe(23500);
  });

  it("calculates asset disposal with gain and VAT correctly and balanced 100%", () => {
    // Cost: 500,000, Accum Deprec: 400,000 => NBV = 100,000
    // Sale price: 150,000 => Gain = 50,000, VAT 7% = 10,500 => Total Received = 160,500
    const disposal = calculateAssetDisposal({
      cost: 500000,
      accumDeprecAtDisposal: 400000,
      salePrice: 150000,
      hasVat: true,
    });

    expect(disposal.netBookValue).toBe(100000);
    expect(disposal.gainLoss).toBe(50000);
    expect(disposal.isGain).toBe(true);
    expect(disposal.vatAmount).toBe(10500);
    expect(disposal.totalReceived).toBe(160500);

    const sumDr = disposal.journalLines.reduce((s, l) => s + l.debit, 0);
    const sumCr = disposal.journalLines.reduce((s, l) => s + l.credit, 0);
    expect(sumDr).toBe(sumCr); // Debit = Credit 100%
  });

  it("calculates asset disposal with loss correctly and balanced 100%", () => {
    // Cost: 200,000, Accum Deprec: 120,000 => NBV = 80,000
    // Sale price: 50,000 => Loss = 30,000, No VAT
    const disposal = calculateAssetDisposal({
      cost: 200000,
      accumDeprecAtDisposal: 120000,
      salePrice: 50000,
      hasVat: false,
    });

    expect(disposal.netBookValue).toBe(80000);
    expect(disposal.gainLoss).toBe(-30000);
    expect(disposal.isGain).toBe(false);

    const sumDr = disposal.journalLines.reduce((s, l) => s + l.debit, 0);
    const sumCr = disposal.journalLines.reduce((s, l) => s + l.credit, 0);
    expect(sumDr).toBe(sumCr);
  });
});
