import { describe, it, expect } from "vitest";
import {
  detectDuplicateInvoices,
  detectExpenseSpikes,
  parseSmartSlipOrReceipt,
  runPreClosingChecklist,
} from "./ai-audit-guard";

describe("AI Audit Copilot & Pre-Closing Checklist Engine (Group F)", () => {
  it("detects exact duplicate invoice numbers from same vendor", () => {
    const invoices = [
      {
        docNo: "PV-01",
        taxId: "0105558000121",
        vendorOrCustomer: "บริษัท เจริญการค้า จำกัด",
        invoiceNo: "INV-9988",
        date: "2026-09-01",
        totalAmount: 10700,
      },
      {
        docNo: "PV-02",
        taxId: "0105558000121",
        vendorOrCustomer: "บริษัท เจริญการค้า จำกัด",
        invoiceNo: "INV-9988", // duplicate invoice number!
        date: "2026-09-10",
        totalAmount: 10700,
      },
    ];

    const alerts = detectDuplicateInvoices(invoices);
    expect(alerts).toHaveLength(1);
    expect(alerts[0].severity).toBe("CRITICAL");
    expect(alerts[0].invoiceNo).toBe("INV-9988");
  });

  it("detects suspected double payment with same amount in close dates", () => {
    const invoices = [
      {
        docNo: "PV-10",
        taxId: "0105558000121",
        vendorOrCustomer: "บจก. เคมีภัณฑ์ไทย",
        invoiceNo: "INV-100",
        date: "2026-09-01",
        totalAmount: 45000,
      },
      {
        docNo: "PV-15",
        taxId: "0105558000121",
        vendorOrCustomer: "บจก. เคมีภัณฑ์ไทย",
        invoiceNo: "INV-101",
        date: "2026-09-05", // 4 days later, exact same amount
        totalAmount: 45000,
      },
    ];

    const alerts = detectDuplicateInvoices(invoices);
    expect(alerts).toHaveLength(1);
    expect(alerts[0].severity).toBe("WARNING");
    expect(alerts[0].reasonTh).toContain("พบรายการตั้งเบิกยอดเงินตรงกัน");
  });

  it("detects statistical expense spikes (> 150% jump)", () => {
    const spike = detectExpenseSpikes({
      accountCode: "530101",
      accountName: "ค่าไฟฟ้าและพลังงาน",
      currentPeriod: "2026-09",
      currentAmount: 95000,
      historicalAmounts: [30000, 28000, 32000, 31000, 29000, 30000], // avg = 30,000
    });

    expect(spike).not.toBeNull();
    expect(spike?.spikePercent).toBeGreaterThanOrEqual(200);
    expect(spike?.zScore).toBeGreaterThanOrEqual(2.5);
  });

  it("parses Thai bank mobile transfer slip OCR text and drafts journal", () => {
    const slipText = `
      โอนเงินสำเร็จ
      ธนาคารกสิกรไทย Kasikornbank
      วันที่ 18 ก.ย. 2569 เวลา 14:32 น.
      จาก: นายสมคิด การค้า
      ไปยัง: บจก. บีซี ไอที
      จำนวนเงิน: 12,500.00 บาท
      รหัสอ้างอิง: KBANK20260918991234
    `;

    const parsed = parseSmartSlipOrReceipt(slipText);
    expect(parsed.bankName).toContain("กสิกรไทย");
    expect(parsed.amount).toBe(12500.0);
    expect(parsed.referenceNo).toBe("KBANK20260918991234");
    expect(parsed.transactionDate).toBe("2026-09-18");
    expect(parsed.draftJournal.type).toBe("PV");
    expect(parsed.draftJournal.lines).toHaveLength(2);
  });

  it("evaluates Pre-Closing 12-point checklist and calculates readiness score", () => {
    // 10 pass, 2 warnings
    const result = runPreClosingChecklist({
      cashOnHandBalance: 25000,
      isBankReconciled: true,
      unbalancedJournalsCount: 0,
      draftVouchersCount: 0,
      depreciationPosted: true,
      stockCountAdjustmentDone: true,
      vatClosingDone: true,
      pendingWhtCertificatesCount: 2, // warning
      unclearedSuspenseCount: 0,
      overdueArDays90Count: 1, // warning
      branchIntercompanyMatched: true,
      taxProvisionEstimated: true,
    });

    expect(result.items).toHaveLength(12);
    expect(result.criticalCount).toBe(0);
    expect(result.warningCount).toBe(2);
    expect(result.passCount).toBe(10);
    expect(result.overallStatus).toBe("WARNING");
    expect(result.readinessScore).toBeGreaterThanOrEqual(90);
  });
});
