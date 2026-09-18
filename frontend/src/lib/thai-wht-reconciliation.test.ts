import { describe, it, expect } from "vitest";
import {
  reconcileWhtWithGl,
  type GlWhtBalances,
  type WhtRegisterData,
  type ThaiWhtPendingCertificate,
} from "./thai-wht-reconciliation";
import type { ThaiWhtRecord } from "./thai-wht";

describe("thai-wht-reconciliation", () => {
  const mockPnd3: ThaiWhtRecord[] = [
    {
      id: "w1",
      docNo: "50TWI-001",
      docDate: "2026-09-10",
      filingType: "pnd3",
      payeeType: "individual",
      payeeTaxId: "1100500123451",
      payeeName: "นาย สมศักดิ์ กิจการ",
      payeeAddress: "กทม.",
      payeeBranchNo: "00000",
      isHeadOffice: true,
      incomeType: "service_subcontract_40_8",
      incomeDescription: "ค่าจ้างทำของ",
      taxRate: 3,
      paymentAmount: 10000,
      whtAmount: 300,
      condition: "deducted",
      status: "active",
    },
  ];

  const mockPnd53: ThaiWhtRecord[] = [
    {
      id: "w2",
      docNo: "50TWI-002",
      docDate: "2026-09-12",
      filingType: "pnd53",
      payeeType: "corporate",
      payeeTaxId: "0105558012345",
      payeeName: "บริษัท ซอฟต์แวร์ เฮ้าส์ จำกัด",
      payeeAddress: "กทม.",
      payeeBranchNo: "00000",
      isHeadOffice: true,
      incomeType: "rent_40_5",
      incomeDescription: "ค่าเช่าห้องเซิร์ฟเวอร์",
      taxRate: 5,
      paymentAmount: 20000,
      whtAmount: 1000,
      condition: "deducted",
      status: "active",
    },
  ];

  const mockReceived: ThaiWhtRecord[] = [
    {
      id: "w3",
      docNo: "REC-001",
      docDate: "2026-09-14",
      filingType: "pnd53",
      payeeType: "corporate",
      payeeTaxId: "0105559123456",
      payeeName: "บริษัท ลูกค้า เอ จำกัด",
      payeeAddress: "กทม.",
      payeeBranchNo: "00000",
      isHeadOffice: true,
      incomeType: "service_subcontract_40_8",
      incomeDescription: "ค่าบริการพัฒนาระบบ",
      taxRate: 3,
      paymentAmount: 50000,
      whtAmount: 1500,
      condition: "deducted",
      status: "active",
    },
  ];

  it("should report fully balanced when GL and Registers match exactly", () => {
    // ภ.ง.ด.3 (300) + ภ.ง.ด.53 (1000) = 1300
    // ทะเบียนรับ = 1500
    const gl: GlWhtBalances = {
      whtPayableCredit: 1300,
      whtReceivableDebit: 1500,
    };

    const registers: WhtRegisterData = {
      pnd3Records: mockPnd3,
      pnd53Records: mockPnd53,
      receivedRecords: mockReceived,
    };

    const report = reconcileWhtWithGl(9, 2026, registers, gl);

    expect(report.isFullyBalanced).toBe(true);
    expect(report.isPayableMatched).toBe(true);
    expect(report.isReceivableMatched).toBe(true);
    expect(report.totalVariance).toBe(0);
    expect(report.cashOutflowRequired).toBe(1300);
    expect(report.duePaymentDatePaperTh).toBe("7 ตุลาคม 2569");
    expect(report.duePaymentDateOnlineTh).toBe("15 ตุลาคม 2569");
  });

  it("should detect payable discrepancy when GL 2151 does not match filing sum", () => {
    // GL มี 1500 แต่แบบยื่นมีแค่ 1300 (ผลต่าง 200)
    const gl: GlWhtBalances = {
      whtPayableCredit: 1500,
      whtReceivableDebit: 1500,
    };

    const registers: WhtRegisterData = {
      pnd3Records: mockPnd3,
      pnd53Records: mockPnd53,
      receivedRecords: mockReceived,
    };

    const report = reconcileWhtWithGl(9, 2026, registers, gl);

    expect(report.isFullyBalanced).toBe(false);
    expect(report.isPayableMatched).toBe(false);
    expect(report.payableItem.variance).toBe(200);
    expect(report.payableItem.status).toBe("discrepancy");
    expect(report.recommendationsTh.some((r) => r.includes("GL 2151 มียอดสูงกว่าแบบยื่น"))).toBe(true);
  });

  it("should detect receivable discrepancy and pending certificates", () => {
    // GL บันทึกถูกหัก 2100 แต่มีใบ 50 ทวิแค่ 1500 (ค้างรับ 600)
    const gl: GlWhtBalances = {
      whtPayableCredit: 1300,
      whtReceivableDebit: 2100,
    };

    const registers: WhtRegisterData = {
      pnd3Records: mockPnd3,
      pnd53Records: mockPnd53,
      receivedRecords: mockReceived,
    };

    const pendingCerts: ThaiWhtPendingCertificate[] = [
      {
        docNo: "INV-202609-042",
        docDate: "2026-09-05",
        customerName: "บริษัท บีซี อินเตอร์ จำกัด",
        customerTaxId: "0105553099999",
        incomeDescription: "ค่าบริการที่ปรึกษา",
        baseAmount: 20000,
        whtAmount: 600,
        daysOutstanding: 13,
      },
    ];

    const report = reconcileWhtWithGl(9, 2026, registers, gl, pendingCerts);

    expect(report.isReceivableMatched).toBe(false);
    expect(report.receivableItem.variance).toBe(600);
    expect(report.pendingCertificates.length).toBe(1);
    expect(report.recommendationsTh.some((r) => r.includes("หนังสือรับรอง 50 ทวิค้างรับ"))).toBe(true);
  });

  it("should handle leap year and year crossover correctly for next month due dates", () => {
    const gl: GlWhtBalances = { whtPayableCredit: 0, whtReceivableDebit: 0 };
    const registers: WhtRegisterData = { pnd3Records: [], pnd53Records: [] };

    // เดือนธันวาคม ข้ามปีไปมกราคมถัดไป
    const decReport = reconcileWhtWithGl(12, 2026, registers, gl);
    expect(decReport.duePaymentDatePaperTh).toBe("7 มกราคม 2570");
    expect(decReport.duePaymentDateOnlineTh).toBe("15 มกราคม 2570");
  });
});
