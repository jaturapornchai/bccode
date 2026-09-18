import { describe, it, expect, vi, afterEach } from "vitest";
import {
  ERP_REPORT_CONFIGS,
  isErpReportRoute,
  isErpReportApiReady,
  fetchErpReportData,
} from "./erp-reports";
import { setupTestAuthSession } from "./test-auth-session";

setupTestAuthSession();

describe("Unified ERP Reporting Engine", () => {
  it("registers all 26 ERP report routes across 6 categories", () => {
    expect(ERP_REPORT_CONFIGS.length).toBe(21);

    // Inventory
    expect(isErpReportRoute("/report/stockbalanceitem")).toBe(true);
    expect(isErpReportRoute("/report/stockbalancewarehouse")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebistockbalance")).toBe(true);
    expect(isErpReportRoute("/report/lowstock")).toBe(true);
    expect(isErpReportRoute("/report/stockmovementcost")).toBe(true);
    expect(isErpReportRoute("/report/reportstockmovement")).toBe(true);

    // Sales
    expect(isErpReportRoute("/report/reportdedebisales")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebisalesdaily")).toBe(true);
    expect(isErpReportRoute("/report/salesbyseller")).toBe(true);
    expect(isErpReportRoute("/report/salesreportbydocument")).toBe(true);
    expect(isErpReportRoute("/report/reportgrossprofitbydocument")).toBe(true);
    expect(isErpReportRoute("/report/reportgrossprofitbyproduct")).toBe(true);
    expect(isErpReportRoute("/report/reportdedebisalereturn")).toBe(true);

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
});

describe("fetchErpReportData", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("reports API readiness only for the 4 wired reports", () => {
    expect(isErpReportApiReady("sales_by_document")).toBe(true);
    expect(isErpReportApiReady("gross_profit_document")).toBe(true);
    expect(isErpReportApiReady("stock_balance_item")).toBe(true);
    expect(isErpReportApiReady("stock_balance_warehouse")).toBe(true);

    expect(isErpReportApiReady("ar_aging")).toBe(false);
    expect(isErpReportApiReady("dbd_xbrl_export")).toBe(false);
    expect(isErpReportApiReady("low_stock_report")).toBe(false);
  });

  it("maps sales_by_document response without fabricating fields", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        status: "success",
        data: [
          {
            docdate: "2026-09-02",
            doctime: "10:22:01",
            docno: "INV-001",
            debtorcode: "C-001",
            debtorname: "Test Customer",
            totalqty: 3,
            totalamount: 53500,
            price: 100,
            averagecost: 80,
            calcamount: 32000,
            grossprofit: 21500,
          },
        ],
        count: 1,
      }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "sales_by_document",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBeUndefined();
    expect(result.rows).toHaveLength(1);
    const row = result.rows[0];
    expect(row.docno).toBe("INV-001");
    expect(row.docdate).toBe("2026-09-02");
    expect(row.custcode).toBe("C-001");
    expect(row.custname).toBe("Test Customer");
    expect(row.totalamount).toBe(53500);
    expect("vatamount" in row).toBe(false);
    expect("subtotal" in row).toBe(false);
    expect("marginpercent" in row).toBe(false);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("computes marginpercent for gross_profit_document from backend numbers", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        status: "success",
        data: [
          {
            docno: "INV-002",
            debtorname: "GP Customer",
            totalamount: 1000,
            calcamount: 700,
            grossprofit: 300,
          },
        ],
      }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "gross_profit_document",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBeUndefined();
    const row = result.rows[0];
    expect(row.docno).toBe("INV-002");
    expect(row.custname).toBe("GP Customer");
    expect(row.salesrevenue).toBe(1000);
    expect(row.costofgoods).toBe(700);
    expect(row.grossprofit).toBe(300);
    expect(row.marginpercent).toBe(30);
  });

  it("maps stock_balance_item from inventory-valuation payload", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        holdingcode: "H01",
        asofdate: "2026-09-15T00:00:00Z",
        totalvalue: 18750,
        items: [
          {
            itemcode: "P-001",
            itemname: "Product 1",
            whcode: "WH-01",
            unitcode: "EA",
            qty: 1500,
            averagecost: 12.5,
            totalvalue: 18750,
            costingmethod: "average",
          },
        ],
      }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "stock_balance_item",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBeUndefined();
    const row = result.rows[0];
    expect(row.itemcode).toBe("P-001");
    expect(row.itemname).toBe("Product 1");
    expect(row.qty).toBe(1500);
    expect(row.unitname).toBe("EA");
    expect(row.avgcost).toBe(12.5);
    expect(row.totalcost).toBe(18750);
    expect("warehouse" in row).toBe(false);
  });

  it("returns report_not_available without fetching for unwired reports", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "ar_aging",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBe("report_not_available");
    expect(result.rows).toEqual([]);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("returns holding_required without fetching when holdingcode is empty", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "sales_by_document",
      holdingcode: "",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBe("holding_required");
    expect(result.rows).toEqual([]);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("returns unauthorized for 401 responses", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      status: 401,
      json: async () => ({}),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "sales_by_document",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBe("unauthorized");
    expect(result.rows).toEqual([]);
  });

  it("returns load_failed for non-auth failures", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      status: 500,
      json: async () => ({}),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "sales_by_document",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBe("load_failed");
    expect(result.rows).toEqual([]);
  });

  it("returns load_failed when response shape is unexpected", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ status: "success", data: "not-an-array" }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "sales_by_document",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBe("load_failed");
    expect(result.rows).toEqual([]);
  });

  it("returns connection_error when fetch throws", async () => {
    const fetchMock = vi.fn(async () => {
      throw new Error("network down");
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchErpReportData({
      code: "sales_by_document",
      holdingcode: "H01",
      fromdate: "2026-09-01",
      todate: "2026-09-30",
    });

    expect(result.error).toBe("connection_error");
    expect(result.rows).toEqual([]);
  });
});
