import { describe, it, expect } from "vitest";
import {
  reconcileVatWithGl,
  computeOfficialPp30,
  type VatRegisterTotals,
  type GlVatBalances,
} from "./thai-vat-reconciliation";

describe("Thai VAT Reconciliation & Official PP.30 Engine", () => {
  describe("reconcileVatWithGl", () => {
    it("returns fully balanced when GL and Tax Register match exactly", () => {
      const register: VatRegisterTotals = {
        taxableSalesBase: 100000,
        salesVat: 7000,
        taxablePurchasesBase: 60000,
        purchasesVat: 4200,
      };

      const gl: GlVatBalances = {
        inputVatDebit: 4200,
        outputVatCredit: 7000,
      };

      const result = reconcileVatWithGl(9, 2026, register, gl);
      expect(result.isFullyBalanced).toBe(true);
      expect(result.totalVariance).toBe(0);
      expect(result.inputVat.status).toBe("matched");
      expect(result.outputVat.status).toBe("matched");
      expect(result.recommendationsTh[0]).toContain("ตรงกับรายงานภาษีและแบบ ภ.พ.30 ครบถ้วน 100%");
    });

    it("detects input VAT discrepancy and provides actionable Thai accounting recommendation", () => {
      const register: VatRegisterTotals = {
        taxableSalesBase: 100000,
        salesVat: 7000,
        taxablePurchasesBase: 60000,
        purchasesVat: 4200,
      };

      // GL มีภาษีซื้อ 4,550 บาท (สูงกว่ารายงาน 350 บาท)
      const gl: GlVatBalances = {
        inputVatDebit: 4550,
        outputVatCredit: 7000,
      };

      const result = reconcileVatWithGl(9, 2026, register, gl);
      expect(result.isFullyBalanced).toBe(false);
      expect(result.inputVat.status).toBe("discrepancy");
      expect(result.inputVat.variance).toBe(350);
      expect(result.totalVariance).toBe(350);
      expect(result.recommendationsTh.some((r) => r.includes("ภาษีซื้อใน GL สูงกว่ารายงานภาษีซื้อ 350.00 บาท"))).toBe(true);
    });

    it("detects output VAT discrepancy when register has unposted tax invoices", () => {
      const register: VatRegisterTotals = {
        taxableSalesBase: 100000,
        salesVat: 7000,
        taxablePurchasesBase: 60000,
        purchasesVat: 4200,
      };

      // GL มีภาษีขายแค่ 6,000 บาท (รายงานสูงกว่า 1,000 บาท)
      const gl: GlVatBalances = {
        inputVatDebit: 4200,
        outputVatCredit: 6000,
      };

      const result = reconcileVatWithGl(9, 2026, register, gl);
      expect(result.isFullyBalanced).toBe(false);
      expect(result.outputVat.status).toBe("discrepancy");
      expect(result.outputVat.variance).toBe(-1000);
      expect(result.recommendationsTh.some((r) => r.includes("ภาษีขายในรายงานภาษีขายสูงกว่า GL 1000.00 บาท"))).toBe(true);
    });
  });

  describe("computeOfficialPp30", () => {
    it("computes official PP.30 items 1 to 10 when tax payable is due", () => {
      const pp30 = computeOfficialPp30({
        taxId: "0105559123456",
        companyName: "บริษัท บีซีเอไอ แอคเคานต์ จำกัด",
        branchNo: "00000",
        month: 8,
        yearCe: 2026,
        taxableSales: 500000,
        zeroRatedSales: 50000,
        exemptSales: 20000,
        taxablePurchases: 300000,
        creditBroughtForward: 1000,
      });

      expect(pp30.taxPeriodYearBe).toBe(2569);
      expect(pp30.isHeadOffice).toBe(true);

      // ข้อ 1 = 500,000 + 50,000 + 20,000 = 570,000
      expect(pp30.item1_grossSales).toBe(570000);
      // ข้อ 2 = 50,000
      expect(pp30.item2_zeroRatedSales).toBe(50000);
      // ข้อ 3 = 20,000
      expect(pp30.item3_exemptSales).toBe(20000);
      // ข้อ 4 = 500,000
      expect(pp30.item4_taxableSales).toBe(500000);
      // ข้อ 5 = 35,000 (7% ของ 500,000)
      expect(pp30.item5_outputVat).toBe(35000);
      // ข้อ 6 = 300,000
      expect(pp30.item6_taxablePurchases).toBe(300000);
      // ข้อ 7 = 21,000 (7% ของ 300,000)
      expect(pp30.item7_inputVat).toBe(21000);
      // ข้อ 8 = 1,000
      expect(pp30.item8_creditBroughtForward).toBe(1000);
      // ข้อ 9 ภาษีที่ต้องชำระ = 35,000 - (21,000 + 1,000) = 13,000 บาท
      expect(pp30.item9_vatPayable).toBe(13000);
      expect(pp30.item10_vatOverpaid).toBe(0);
      expect(pp30.netPaymentAmount).toBe(13000);
    });

    it("computes overpaid VAT (item 10) when input VAT exceeds output VAT", () => {
      const pp30 = computeOfficialPp30({
        taxId: "0105559123456",
        companyName: "บริษัท บีซีเอไอ แอคเคานต์ จำกัด",
        month: 8,
        yearCe: 2026,
        taxableSales: 100000,
        outputVat: 7000,
        taxablePurchases: 200000,
        inputVat: 14000,
      });

      expect(pp30.item9_vatPayable).toBe(0);
      // ข้อ 10 ภาษีชำระเกิน = 14,000 - 7,000 = 7,000 บาท
      expect(pp30.item10_vatOverpaid).toBe(7000);
      expect(pp30.netPaymentAmount).toBe(0);
    });

    it("computes surcharge (เงินเพิ่ม 1.5% ต่อเดือน) for late filing", () => {
      const pp30 = computeOfficialPp30({
        taxId: "0105559123456",
        companyName: "บริษัท บีซีเอไอ แอคเคานต์ จำกัด",
        month: 5,
        yearCe: 2026,
        taxableSales: 100000,
        outputVat: 7000,
        taxablePurchases: 50000,
        inputVat: 3500,
        isLateFiling: true,
        lateMonths: 2, // ยื่นล่าช้า 2 เดือน -> เงินเพิ่ม 3% ของ 3,500 = 105 บาท
      });

      expect(pp30.item9_vatPayable).toBe(3500);
      expect(pp30.surcharge).toBe(105);
      expect(pp30.netPaymentAmount).toBe(3605);
    });
  });
});
