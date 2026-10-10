import { describe, it, expect } from "vitest";
import {
  getErpReportConfig,
  isErpReportRoute,
  reportText,
} from "./erp-reports";

describe("Bank & Cash ERP Reports (Champ Parity 5164, 5152, 5150, 5156/5157)", () => {
  it("defines bank statement report with proper columns and metadata", () => {
    const route = "/report/bankstatement";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("bank_statement_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("Bank Statement");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("docno");
    expect(columnKeys).toContain("bookcode");
    expect(columnKeys).toContain("chequeno");
    expect(columnKeys).toContain("transdesc");
    expect(columnKeys).toContain("description");
    expect(columnKeys).toContain("deposit");
    expect(columnKeys).toContain("withdraw");
    expect(columnKeys).toContain("balance");

    const depCol = config?.columns.find((c) => c.key === "deposit");
    expect(depCol?.isCurrency).toBe(true);
    expect(depCol?.align).toBe("right");

    const balCol = config?.columns.find((c) => c.key === "balance");
    expect(balCol?.isCurrency).toBe(true);
    expect(balCol?.align).toBe("right");

    const thTitle = reportText(config!, "title", "th");
    expect(thTitle).toBeTruthy();
  });

  it("defines cash movement report with proper columns and metadata", () => {
    const route = "/report/cashmovement";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("cash_movement_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("เคลื่อนไหวเงินสด");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("docno");
    expect(columnKeys).toContain("description");
    expect(columnKeys).toContain("receipt");
    expect(columnKeys).toContain("payment");
    expect(columnKeys).toContain("balance");

    const recCol = config?.columns.find((c) => c.key === "receipt");
    expect(recCol?.isCurrency).toBe(true);
    expect(recCol?.align).toBe("right");
  });

  it("defines petty cash movement report with proper columns and metadata", () => {
    const route = "/report/pettycashmovement";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("petty_cash_movement_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("เคลื่อนไหวเงินสดย่อย");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("docno");
    expect(columnKeys).toContain("pettycashcode");
    expect(columnKeys).toContain("description");
    expect(columnKeys).toContain("receipt");
    expect(columnKeys).toContain("payment");
    expect(columnKeys).toContain("balance");
  });

  it("defines monthly payment book report with proper columns and metadata", () => {
    const route = "/report/monthlypaymentbook";
    expect(isErpReportRoute(route)).toBe(true);

    const config = getErpReportConfig(route);
    expect(config).toBeDefined();
    expect(config?.code).toBe("monthly_payment_book_report");
    expect(config?.category).toBe("finance");
    expect(config?.title.th).toContain("สมุดจ่ายเงินประจำเดือน");
    expect(config?.defaultSortKey).toBe("docdate");

    const columnKeys = config?.columns.map((c) => c.key);
    expect(columnKeys).toContain("docdate");
    expect(columnKeys).toContain("voucherno");
    expect(columnKeys).toContain("description");
    expect(columnKeys).toContain("payto");
    expect(columnKeys).toContain("chequeno");
    expect(columnKeys).toContain("whtamount");
    expect(columnKeys).toContain("amount");

    const amountCol = config?.columns.find((c) => c.key === "amount");
    expect(amountCol?.isCurrency).toBe(true);
    expect(amountCol?.align).toBe("right");
  });
});
