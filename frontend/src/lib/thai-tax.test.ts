import { describe, it, expect } from "vitest";
import {
  THAI_TAX_CONFIGS,
  isThaiTaxRoute,
  getThaiTaxConfig,
  calculatePp30Summary,
  getSampleTaxRecords,
} from "./thai-tax";

describe("Thai Tax Engine (PP.30, PP.36, VAT, WHT, 50 Twi)", () => {
  it("registers all 12 Thai Tax routes", () => {
    expect(THAI_TAX_CONFIGS.length).toBe(12);

    expect(isThaiTaxRoute("/report/reportvatsale")).toBe(true);
    expect(isThaiTaxRoute("/report/reportvatbuy")).toBe(true);
    expect(isThaiTaxRoute("/report/vatpp30")).toBe(true);
    expect(isThaiTaxRoute("/report/vatpp36")).toBe(true);
    expect(isThaiTaxRoute("/report/vatpnd2")).toBe(true);
    expect(isThaiTaxRoute("/report/vatpnd3")).toBe(true);
    expect(isThaiTaxRoute("/report/vatpnd53")).toBe(true);
    expect(isThaiTaxRoute("/report/whtcertificate")).toBe(true);
    expect(isThaiTaxRoute("/report/whtreceived")).toBe(true);
    expect(isThaiTaxRoute("/report/wht-reports")).toBe(true);
    expect(isThaiTaxRoute("/report/deferredtax")).toBe(true);
    expect(isThaiTaxRoute("/report/unreceivedtaxinvoice")).toBe(true);
  });

  it("calculates PP.30 summary correctly according to Revenue Department formula", () => {
    const summary = calculatePp30Summary(
      2026,
      9,
      { taxable: 100000, zeroRated: 10000, exempt: 5000 },
      { claimable: 60000, exempt: 0 },
      500,
    );

    expect(summary.taxyear).toBe(2026);
    expect(summary.taxmonth).toBe(9);
    expect(summary.totalsales).toBe(115000);
    expect(summary.taxablesales).toBe(100000);
    expect(summary.outputvat).toBe(7000); // 7% of 100,000
    expect(summary.claimablepurchases).toBe(60000);
    expect(summary.inputvat).toBe(4200); // 7% of 60,000
    expect(summary.nettaxpayable).toBe(2800); // 7,000 - 4,200
    expect(summary.previousoverpayment).toBe(500);
    expect(summary.finaltaxpayable).toBe(2300); // 2,800 - 500
  });

  it("provides sample tax records for sales, purchases, and withholding tax", () => {
    const sales = getSampleTaxRecords("vat_sale");
    expect(sales.length).toBeGreaterThan(0);
    expect(sales[0].vatamount).toBe(Math.round(sales[0].amountbeforevat * 0.07));

    const buys = getSampleTaxRecords("vat_buy");
    expect(buys.length).toBeGreaterThan(0);

    const wht = getSampleTaxRecords("pnd53");
    expect(wht.length).toBeGreaterThan(0);
    expect(wht[0].incometype).toBeDefined();
    expect(wht[0].taxrate).toBe(3);
    expect(wht[0].whtamount).toBe(1500);
  });
});
