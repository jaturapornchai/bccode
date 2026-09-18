import { describe, it, expect } from "vitest";
import {
  THAI_WHT_INCOME_CONFIGS,
  computePndSummary,
  generate50TwiCertificate,
  type ThaiWhtRecord,
  type PayerInfo,
} from "./thai-wht";

describe("Thai Withholding Tax (WHT) Engine & 50 Twi Certificate", () => {
  const mockPayer: PayerInfo = {
    taxId: "0105559123456",
    nameTh: "บริษัท บีซีเอไอ แอคเคานต์ จำกัด",
    addressTh: "123/45 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร 10110",
    branchNo: "00000",
    isHeadOffice: true,
  };

  const sampleRecords: ThaiWhtRecord[] = [
    {
      id: "wht-1",
      docNo: "50TWI-2026/09-001",
      docDate: "2026-09-05",
      filingType: "pnd3",
      payeeType: "individual",
      payeeTaxId: "1100500123456",
      payeeName: "นายสมชาย ใจดี (ผู้ให้เช่า)",
      payeeAddress: "99 หมู่ 2 ต.บางกรวย อ.บางกรวย จ.นนทบุรี 11130",
      payeeBranchNo: "00000",
      isHeadOffice: true,
      incomeType: "rent_40_5",
      incomeDescription: "ค่าเช่าสำนักงานเดือนกันยายน 2569",
      taxRate: 5,
      paymentAmount: 20000,
      whtAmount: 1000,
      condition: "deducted",
      status: "active",
    },
    {
      id: "wht-2",
      docNo: "50TWI-2026/09-002",
      docDate: "2026-09-10",
      filingType: "pnd3",
      payeeType: "individual",
      payeeTaxId: "3100600789012",
      payeeName: "นางสาวสมศรี มีสุข (โปรแกรมเมอร์ฟรีแลนซ์)",
      payeeAddress: "55/12 แขวงพญาไท เขตพญาไท กรุงเทพมหานคร 10400",
      payeeBranchNo: "00000",
      isHeadOffice: true,
      incomeType: "service_subcontract_40_8",
      incomeDescription: "ค่าจ้างพัฒนาระบบซอฟต์แวร์",
      taxRate: 3,
      paymentAmount: 30000,
      whtAmount: 900,
      condition: "deducted",
      status: "active",
    },
    {
      id: "wht-3",
      docNo: "50TWI-2026/09-003",
      docDate: "2026-09-15",
      filingType: "pnd53",
      payeeType: "corporate",
      payeeTaxId: "0105558012345",
      payeeName: "บริษัท ฟาสต์ โลจิสติกส์ จำกัด",
      payeeAddress: "888 ถนนบางนา-ตราด กม.18 อ.บางพลี จ.สมุทรปราการ 10540",
      payeeBranchNo: "00001",
      isHeadOffice: false,
      incomeType: "transportation_40_8",
      incomeDescription: "ค่าบริการขนส่งสินค้ากระจายสาขา",
      taxRate: 1,
      paymentAmount: 50000,
      whtAmount: 500,
      condition: "deducted",
      status: "active",
    },
    {
      id: "wht-4",
      docNo: "50TWI-2026/09-004",
      docDate: "2026-09-20",
      filingType: "pnd53",
      payeeType: "corporate",
      payeeTaxId: "0105556098765",
      payeeName: "บริษัท มีเดีย แอดเวอร์ไทซิ่ง จำกัด",
      payeeAddress: "100 ถนนพระราม 9 เขตห้วยขวาง กรุงเทพมหานคร 10310",
      payeeBranchNo: "00000",
      isHeadOffice: true,
      incomeType: "advertising_40_8",
      incomeDescription: "ค่าสื่อโฆษณาประชาสัมพันธ์",
      taxRate: 2,
      paymentAmount: 40000,
      whtAmount: 800,
      condition: "deducted",
      status: "active",
    },
  ];

  it("provides complete statutory income configs according to Thai Revenue Department", () => {
    expect(THAI_WHT_INCOME_CONFIGS.rent_40_5.defaultRate).toBe(5);
    expect(THAI_WHT_INCOME_CONFIGS.service_subcontract_40_8.defaultRate).toBe(3);
    expect(THAI_WHT_INCOME_CONFIGS.transportation_40_8.defaultRate).toBe(1);
    expect(THAI_WHT_INCOME_CONFIGS.advertising_40_8.defaultRate).toBe(2);
  });

  describe("computePndSummary", () => {
    it("computes PND.3 summary correctly for individuals", () => {
      const pnd3 = computePndSummary(sampleRecords, "pnd3", 9, 2026);

      expect(pnd3.formType).toBe("pnd3");
      expect(pnd3.periodMonth).toBe(9);
      expect(pnd3.periodYearBe).toBe(2569);
      expect(pnd3.totalPayees).toBe(2);
      expect(pnd3.totalPaymentAmount).toBe(50000); // 20000 + 30000
      expect(pnd3.totalWhtAmount).toBe(1900);       // 1000 + 900
      expect(pnd3.totalWhtTextTh).toBe("หนึ่งพันเก้าร้อยบาทถ้วน");
      expect(pnd3.byIncomeType.length).toBe(2);
    });

    it("computes PND.53 summary correctly for corporations", () => {
      const pnd53 = computePndSummary(sampleRecords, "pnd53", 9, 2026);

      expect(pnd53.formType).toBe("pnd53");
      expect(pnd53.totalPayees).toBe(2);
      expect(pnd53.totalPaymentAmount).toBe(90000); // 50000 + 40000
      expect(pnd53.totalWhtAmount).toBe(1300);       // 500 + 800
      expect(pnd53.totalWhtTextTh).toBe("หนึ่งพันสามร้อยบาทถ้วน");
    });
  });

  describe("generate50TwiCertificate", () => {
    it("generates a complete 50 Twi withholding certificate data", () => {
      const record = sampleRecords[0];
      const cert = generate50TwiCertificate(record, mockPayer);

      expect(cert.certNo).toBe("50TWI-2026/09-001");
      expect(cert.payer.taxId).toBe("0105559123456");
      expect(cert.payee.taxId).toBe("1100500123456");
      expect(cert.payee.name).toBe("นายสมชาย ใจดี (ผู้ให้เช่า)");
      expect(cert.totalPayment).toBe(20000);
      expect(cert.totalWht).toBe(1000);
      expect(cert.totalWhtBahtText).toBe("หนึ่งพันบาทถ้วน");
      expect(cert.conditionTextTh).toBe("(1) หัก ณ ที่จ่าย");
      expect(cert.items[0].incomeCategoryName).toContain("มาตรา 40(5)");
    });
  });
});
