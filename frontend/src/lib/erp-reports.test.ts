import { describe, it, expect } from "vitest";
import {
  ERP_REPORT_CONFIGS,
  isErpReportRoute,
  getErpReportConfig,
  getSampleReportData,
} from "./erp-reports";

describe("Unified ERP Reporting Engine", () => {
  it("registers all 26 ERP report routes across 6 categories", () => {
    expect(ERP_REPORT_CONFIGS.length).toBe(26);

    // Inventory
    expect(isErpReportRoute("/report/stockbalanceitem")).toBe(true);
    expect(isErpReportRoute("/report/stockbalancewarehouse")).toBe(true);
    expect(isErpReportRoute("/report/stockbalancelocation")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebistockbalance")).toBe(true);
    expect(isErpReportRoute("/report/lowstock")).toBe(true);
    expect(isErpReportRoute("/report/expiringstock")).toBe(true);
    expect(isErpReportRoute("/report/stockmovementcost")).toBe(true);
    expect(isErpReportRoute("/report/reportstockmovement")).toBe(true);
    expect(isErpReportRoute("/report/stocklotmovement")).toBe(true);

    // Sales
    expect(isErpReportRoute("/report/reportdedebisales")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebisalesdaily")).toBe(true);
    expect(isErpReportRoute("/report/salesbyseller")).toBe(true);
    expect(isErpReportRoute("/report/salesreportbydocument")).toBe(true);
    expect(isErpReportRoute("/report/reportgrossprofitbydocument")).toBe(true);
    expect(isErpReportRoute("/report/reportgrossprofitbyproduct")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebisalereturn")).toBe(true);
    expect(isErpReportRoute("/report/salesbycustomer")).toBe(true);
    expect(isErpReportRoute("/report/salesbychannel")).toBe(true);

    // Purchase
    expect(isErpReportRoute("/report/reportdedebipurchase")).toBe(true);
    expect(isErpReportRoute("/report/purchasebyproduct")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebipurchasepartial")).toBe(true);
    expect(isErpReportRoute("/report/expensesummary")).toBe(true);

    // AR & AP
    expect(isErpReportRoute("/report/araging")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebipaymentdaily")).toBe(true);
    expect(isErpReportRoute("/report/apaging")).toBe(true);

    // XBRL
    expect(isErpReportRoute("/report/xbrl")).toBe(true);
  });

  it("provides column definitions and sample data for reports", () => {
    const stockConfig = getErpReportConfig("/report/stockbalanceitem");
    expect(stockConfig).toBeDefined();
    expect(stockConfig?.columns.length).toBeGreaterThan(0);

    const stockData = getSampleReportData(stockConfig!.code);
    expect(stockData.length).toBeGreaterThan(0);

    const xbrlConfig = getErpReportConfig("/report/xbrl");
    expect(xbrlConfig?.category).toBe("xbrl");
    const xbrlData = getSampleReportData(xbrlConfig!.code);
    expect(xbrlData.some((x) => x.xbrltag === "th-gaap-ci:CashAndCashEquivalents")).toBe(true);
  });
});
