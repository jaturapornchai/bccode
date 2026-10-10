import { describe, it, expect } from "vitest";
import {
  ERP_REPORT_CONFIGS,
  getErpReportConfig,
  isErpReportRoute,
  reportText,
  reportColumnLabel,
} from "./erp-reports";

describe("Cheque & Credit Card ERP Reports (Champ Parity 5380, 5386, 5393)", () => {
  it("defines received cheques report with proper columns and metadata", () => {
    const route = "/report/chequereceived";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("cheque_received_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("รายงานรายละเอียดเช็ครับ");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("docno");
    expect(columnKeys).toContain("duedate");
    expect(columnKeys).toContain("debtorcode");
    expect(columnKeys).toContain("debtorname");
    expect(columnKeys).toContain("chqnumber");
    expect(columnKeys).toContain("bankname");
    expect(columnKeys).toContain("status");
    expect(columnKeys).toContain("amount");
    expect(columnKeys).toContain("remark");

    const amountCol = config?.columns.find((c) => c.key === "amount");
    expect(amountCol?.isCurrency).toBe(true);
    expect(amountCol?.align).toBe("right");

    const thTitle = reportText(config!, "title", "th");
    expect(thTitle).toBeTruthy();
  });

  it("defines issued cheques report with proper columns and metadata", () => {
    const route = "/report/chequeissued";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("cheque_issued_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("รายงานรายละเอียดเช็คจ่าย");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("docno");
    expect(columnKeys).toContain("duedate");
    expect(columnKeys).toContain("creditorcode");
    expect(columnKeys).toContain("creditorname");
    expect(columnKeys).toContain("chqnumber");
    expect(columnKeys).toContain("bankbook");
    expect(columnKeys).toContain("status");
    expect(columnKeys).toContain("amount");
    expect(columnKeys).toContain("remark");

    const amountCol = config?.columns.find((c) => c.key === "amount");
    expect(amountCol?.isCurrency).toBe(true);
    expect(amountCol?.align).toBe("right");
  });

  it("defines credit card report with proper columns and metadata", () => {
    const route = "/report/creditcard";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("credit_card_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("รายงานรายละเอียดบัตรเครดิต");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("docno");
    expect(columnKeys).toContain("cardno");
    expect(columnKeys).toContain("cardtype");
    expect(columnKeys).toContain("holdername");
    expect(columnKeys).toContain("bankname");
    expect(columnKeys).toContain("amount");
    expect(columnKeys).toContain("fee");
    expect(columnKeys).toContain("netamount");
    expect(columnKeys).toContain("status");

    const netAmountCol = config?.columns.find((c) => c.key === "netamount");
    expect(netAmountCol?.isCurrency).toBe(true);
    expect(netAmountCol?.align).toBe("right");
  });
});
