import { describe, it, expect } from "vitest";
import { computeOfficialPp30 } from "./thai-pp30-form";

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
