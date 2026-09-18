import { describe, expect, it } from "vitest";
import { buildComparativeReport, pivotAnnualBalances, calculatePercentChange } from "./gl-comparative-report";

describe("Multi-Period Comparative Financial Statements (gl-comparative-report)", () => {
  it("calculates percentage change correctly with edge cases", () => {
    // 0 to 0
    expect(calculatePercentChange(0n, 0n)).toEqual({
      percentStr: "0.00%",
      direction: "unchanged",
      isNew: false,
    });

    // 0 to 100
    expect(calculatePercentChange(10000000000n, 0n)).toEqual({
      percentStr: "+100.00%",
      direction: "increase",
      isNew: true,
    });

    // 100 to 0
    expect(calculatePercentChange(0n, 10000000000n)).toEqual({
      percentStr: "-100.00%",
      direction: "decrease",
      isNew: false,
    });

    // 100 to 150 (+50%)
    expect(calculatePercentChange(15000000000n, 10000000000n)).toEqual({
      percentStr: "+50.00%",
      direction: "increase",
      isNew: false,
    });

    // 100 to 80 (-20%)
    expect(calculatePercentChange(8000000000n, 10000000000n)).toEqual({
      percentStr: "-20.00%",
      direction: "decrease",
      isNew: false,
    });
  });

  it("builds comparative report across two periods with variance and % change", () => {
    const period1Rows = [
      { accountcode: "4111", accountname: "รายได้จากการขาย", amount: "120000.00" },
      { accountcode: "5111", accountname: "ต้นทุนขาย", amount: "80000.00" },
      { accountcode: "5322", accountname: "ค่าโฆษณาออนไลน์", amount: "15000.00" }, // New in P1
    ];

    const period2Rows = [
      { accountcode: "4111", accountname: "รายได้จากการขาย", amount: "100000.00" },
      { accountcode: "5111", accountname: "ต้นทุนขาย", amount: "80000.00" },
      { accountcode: "5321", accountname: "ค่าเช่าสำนักงานเดิม", amount: "20000.00" }, // Only in P2
    ];

    const result = buildComparativeReport({
      period1Label: "ปี 2569",
      period1Rows,
      period2Label: "ปี 2568",
      period2Rows,
    });

    expect(result.rows.length).toBe(4);

    // 4111 Sales: 120,000 vs 100,000 (+20,000, +20.00%)
    const salesRow = result.rows.find((r) => r.accountcode === "4111");
    expect(salesRow).toBeDefined();
    expect(salesRow?.period1Amount).toBe("120,000.00");
    expect(salesRow?.period2Amount).toBe("100,000.00");
    expect(salesRow?.varianceAmount).toBe("20,000.00");
    expect(salesRow?.percentChange).toBe("+20.00%");
    expect(salesRow?.direction).toBe("increase");

    // 5111 COGS: 80,000 vs 80,000 (0 variance, 0.00%)
    const cogsRow = result.rows.find((r) => r.accountcode === "5111");
    expect(cogsRow).toBeDefined();
    expect(cogsRow?.varianceAmount).toBe("0.00");
    expect(cogsRow?.percentChange).toBe("0.00%");
    expect(cogsRow?.direction).toBe("unchanged");

    // 5322 Online Ads: New in P1
    const adsRow = result.rows.find((r) => r.accountcode === "5322");
    expect(adsRow).toBeDefined();
    expect(adsRow?.isNew).toBe(true);
    expect(adsRow?.percentChange).toBe("+100.00%");

    // 5321 Old Rent: Only in P2
    const rentRow = result.rows.find((r) => r.accountcode === "5321");
    expect(rentRow).toBeDefined();
    expect(rentRow?.period1Amount).toBe("0.00");
    expect(rentRow?.period2Amount).toBe("20,000.00");
    expect(rentRow?.varianceAmount).toBe("-20,000.00");
    expect(rentRow?.percentChange).toBe("-100.00%");

    // Totals
    // P1 Total: 120k + 80k + 15k = 215k
    // P2 Total: 100k + 80k + 20k = 200k
    // Variance: +15,000 (+7.50%)
    expect(result.totals.period1Total).toBe("215,000.00");
    expect(result.totals.period2Total).toBe("200,000.00");
    expect(result.totals.varianceTotal).toBe("15,000.00");
    expect(result.totals.percentChangeTotal).toBe("+7.50%");
  });

  it("pivots 12-month annual-balances report into a monthly trend matrix", () => {
    const rawRows = [
      { month: "2026-01", accountcode: "4111", accountname: "รายได้ขาย", amount: "10000.00" },
      { month: "2026-02", accountcode: "4111", accountname: "รายได้ขาย", amount: "12000.00" },
      { month: "2026-03", accountcode: "4111", accountname: "รายได้ขาย", amount: "15000.00" },
      { month: "2026-01", accountcode: "5111", accountname: "ต้นทุนขาย", amount: "6000.00" },
      { month: "2026-02", accountcode: "5111", accountname: "ต้นทุนขาย", amount: "7000.00" },
    ];

    const matrix = pivotAnnualBalances(rawRows);

    expect(matrix.columns.length).toBe(15); // accountcode, accountname, m1..m12, total
    expect(matrix.rows.length).toBe(2);

    const salesRow = matrix.rows.find((r) => r.accountcode === "4111");
    expect(salesRow).toBeDefined();
    expect(salesRow?.m1).toBe("10,000.00");
    expect(salesRow?.m2).toBe("12,000.00");
    expect(salesRow?.m3).toBe("15,000.00");
    expect(salesRow?.m4).toBe("0.00");
    expect(salesRow?.total).toBe("37,000.00");

    const cogsRow = matrix.rows.find((r) => r.accountcode === "5111");
    expect(cogsRow).toBeDefined();
    expect(cogsRow?.m1).toBe("6,000.00");
    expect(cogsRow?.m2).toBe("7,000.00");
    expect(cogsRow?.total).toBe("13,000.00");

    expect(matrix.monthTotals.m1).toBe("16,000.00");
    expect(matrix.monthTotals.m2).toBe("19,000.00");
    expect(matrix.grandTotal).toBe("50,000.00");
  });
});
