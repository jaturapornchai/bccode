import { describe, it, expect } from "vitest";
import {
  calculateCFOFinancialHealth,
  calculateCashFlowStatement,
  type CFOBalanceMetrics,
} from "./cfo-financial-health";

describe("CFO Financial Health & Cash Flow Statement Engine (Group E)", () => {
  it("calculates accurate financial ratios and health score for a healthy company", () => {
    const metrics: CFOBalanceMetrics = {
      cashAndBank: 1200000,
      tradeReceivables: 800000,
      inventory: 600000,
      otherCurrentAssets: 100000,
      nonCurrentAssets: 3000000,
      totalAssets: 5700000,

      tradePayables: 500000,
      shortTermLoans: 300000,
      otherCurrentLiabilities: 200000,
      longTermLiabilities: 500000,
      totalLiabilities: 1500000,

      totalEquity: 4200000,

      revenue: 5000000,
      costOfGoodsSold: 3000000,
      grossProfit: 2000000, // 40% GP
      operatingExpenses: 1200000,
      netProfit: 800000, // 16% NP

      monthlyOperatingBurnRate: 100000, // 12 months runway
    };

    const result = calculateCFOFinancialHealth(metrics);

    // Current assets = 2.7M, Current liab = 1.0M => Current Ratio = 2.7x
    const curRatio = result.ratios.find((r) => r.key === "current_ratio");
    expect(curRatio?.value).toBe(2.7);
    expect(curRatio?.status).toBe("HEALTHY");

    // Quick assets = 2.0M, Current liab = 1.0M => Quick Ratio = 2.0x
    const qkRatio = result.ratios.find((r) => r.key === "quick_ratio");
    expect(qkRatio?.value).toBe(2.0);
    expect(qkRatio?.status).toBe("HEALTHY");

    // D/E = 1.5M / 4.2M = 0.36x
    const deRatio = result.ratios.find((r) => r.key === "debt_to_equity");
    expect(deRatio?.value).toBe(0.36);
    expect(deRatio?.status).toBe("HEALTHY");

    // Runway = 1.2M / 100k = 12.0 months
    expect(result.cashRunwayMonths).toBe(12.0);

    // High health score
    expect(result.healthScore).toBeGreaterThanOrEqual(80);
    expect(result.overallStatus).toBe("HEALTHY");
  });

  it("identifies liquidity warning and high leverage", () => {
    const stressedMetrics: CFOBalanceMetrics = {
      cashAndBank: 100000,
      tradeReceivables: 200000,
      inventory: 800000,
      otherCurrentAssets: 50000,
      nonCurrentAssets: 1000000,
      totalAssets: 2150000,

      tradePayables: 900000,
      shortTermLoans: 400000,
      otherCurrentLiabilities: 200000, // Current liab = 1.5M
      longTermLiabilities: 800000,
      totalLiabilities: 2300000,

      totalEquity: 500000, // D/E = 4.6x

      revenue: 1000000,
      costOfGoodsSold: 850000,
      grossProfit: 150000, // 15% GP
      operatingExpenses: 200000,
      netProfit: -50000, // Loss

      monthlyOperatingBurnRate: 60000, // 1.7 months runway
    };

    const result = calculateCFOFinancialHealth(stressedMetrics);

    // Current ratio: 1.15M / 1.5M = 0.77x (Critical < 1.0)
    const curRatio = result.ratios.find((r) => r.key === "current_ratio");
    expect(curRatio?.status).toBe("CRITICAL");

    // D/E ratio: 2.3M / 0.5M = 4.6x (Critical > 2.5)
    const deRatio = result.ratios.find((r) => r.key === "debt_to_equity");
    expect(deRatio?.status).toBe("CRITICAL");

    expect(result.cashRunwayMonths).toBe(1.7);
    expect(result.healthScore).toBeLessThan(50);
    expect(result.overallStatus).toBe("CRITICAL");
  });

  it("calculates Indirect Cash Flow Statement according to TFRS", () => {
    const cf = calculateCashFlowStatement({
      beginningCash: 500000,
      netProfit: 300000,
      depreciation: 50000, // Non-cash addback (+)
      beginningAR: 200000,
      endingAR: 250000, // AR increased by 50,000 => Cash outflow (-50,000)
      beginningInventory: 300000,
      endingInventory: 280000, // Inventory decreased by 20,000 => Cash inflow (+20,000)
      beginningAP: 150000,
      endingAP: 180000, // AP increased by 30,000 => Cash inflow (+30,000)
      capexPurchases: 100000, // Capex outflow (-100,000)
      assetDisposalProceeds: 20000, // Inflow (+20,000)
      loanProceeds: 50000, // Financing inflow (+50,000)
      loanRepayments: 20000, // Financing outflow (-20,000)
      dividendsPaid: 100000, // Financing outflow (-100,000)
    });

    // Operating = 300,000 + 50,000 - 50,000 + 20,000 + 30,000 = 350,000
    expect(cf.operating.netOperatingCashFlow).toBe(350000);

    // Investing = -100,000 + 20,000 = -80,000
    expect(cf.investing.netInvestingCashFlow).toBe(-80000);

    // Financing = 50,000 - 20,000 - 100,000 = -70,000
    expect(cf.financing.netFinancingCashFlow).toBe(-70000);

    // Net Cash Change = 350,000 - 80,000 - 70,000 = 200,000
    expect(cf.summary.netCashChange).toBe(200000);

    // Ending Cash = 500,000 + 200,000 = 700,000
    expect(cf.summary.endingCash).toBe(700000);
  });
});
