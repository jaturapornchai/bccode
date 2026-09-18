/**
 * AI Audit Copilot & Fraud / Leakage Detection Guard
 * Standards: Thai Audit & Accounting Standards (TAS / TFRS)
 * Includes Duplicate Invoices Detection, Expense Spikes Analysis, Smart Slip OCR Parser, and Pre-Closing 12-Point Checklist
 */

export type DuplicateInvoiceAlert = {
  id: string;
  docNo: string;
  conflictingDocNo: string;
  vendorName: string;
  invoiceNo: string;
  amount: number;
  date: string;
  conflictingDate: string;
  severity: "CRITICAL" | "WARNING";
  reasonTh: string;
};

export type ExpenseAnomalyAlert = {
  accountCode: string;
  accountName: string;
  currentAmount: number;
  averageAmount: number;
  spikePercent: number;
  zScore: number;
  period: string;
  adviceTh: string;
};

export type SmartSlipResult = {
  bankName: string;
  senderName?: string;
  receiverName?: string;
  transactionDate: string;
  transactionTime?: string;
  amount: number;
  referenceNo?: string;
  channel: "promptpay" | "transfer" | "qr";
  draftJournal: {
    type: "RV" | "PV";
    description: string;
    lines: { accountCode: string; debit: number; credit: number; description: string }[];
  };
};

export type ChecklistStatus = "PASS" | "WARNING" | "CRITICAL";

export type PreClosingChecklistItem = {
  id: number;
  code: string;
  titleTh: string;
  titleEn: string;
  status: ChecklistStatus;
  detailTh: string;
  actionUrl?: string;
  actionTextTh?: string;
};

export type PreClosingChecklistResult = {
  readinessScore: number; // 0 - 100%
  overallStatus: ChecklistStatus;
  passCount: number;
  warningCount: number;
  criticalCount: number;
  items: PreClosingChecklistItem[];
};

/**
 * Detects duplicate tax invoices and suspected double payments.
 */
export function detectDuplicateInvoices(
  invoices: {
    docNo: string;
    taxId?: string;
    vendorOrCustomer: string;
    invoiceNo: string;
    date: string;
    totalAmount: number;
  }[],
): DuplicateInvoiceAlert[] {
  const alerts: DuplicateInvoiceAlert[] = [];

  for (let i = 0; i < invoices.length; i++) {
    for (let j = i + 1; j < invoices.length; j++) {
      const invA = invoices[i];
      const invB = invoices[j];

      // Match 1: Exact Same Invoice Number and Same Tax ID or Vendor
      const sameInvNo =
        invA.invoiceNo.trim().length > 0 &&
        invA.invoiceNo.trim().toLowerCase() === invB.invoiceNo.trim().toLowerCase();

      const sameVendor =
        (invA.taxId && invB.taxId && invA.taxId === invB.taxId) ||
        invA.vendorOrCustomer.trim().toLowerCase() === invB.vendorOrCustomer.trim().toLowerCase();

      if (sameInvNo && sameVendor) {
        alerts.push({
          id: `dup-inv-${invA.docNo}-${invB.docNo}`,
          docNo: invA.docNo,
          conflictingDocNo: invB.docNo,
          vendorName: invA.vendorOrCustomer,
          invoiceNo: invA.invoiceNo,
          amount: invA.totalAmount,
          date: invA.date,
          conflictingDate: invB.date,
          severity: "CRITICAL",
          reasonTh: `พบเลขที่ใบกำกับภาษีซ้ำกัน '${invA.invoiceNo}' จากผู้จำหน่ายรายเดียวกัน (${invA.vendorOrCustomer}) เสี่ยงต่อการจ่ายเงินหรือเคลมภาษีซื้อซ้ำซ้อน`,
        });
        continue;
      }

      // Match 2: Same Vendor + Exact Same Amount within 15 days window
      if (sameVendor && Math.abs(invA.totalAmount - invB.totalAmount) < 0.01) {
        const timeA = new Date(invA.date).getTime();
        const timeB = new Date(invB.date).getTime();
        const diffDays = Math.abs(timeA - timeB) / (1000 * 60 * 60 * 24);

        if (diffDays <= 15) {
          alerts.push({
            id: `dup-amt-${invA.docNo}-${invB.docNo}`,
            docNo: invA.docNo,
            conflictingDocNo: invB.docNo,
            vendorName: invA.vendorOrCustomer,
            invoiceNo: `${invA.invoiceNo} / ${invB.invoiceNo}`,
            amount: invA.totalAmount,
            date: invA.date,
            conflictingDate: invB.date,
            severity: "WARNING",
            reasonTh: `พบรายการตั้งเบิกยอดเงินตรงกัน (฿${invA.totalAmount.toLocaleString()}) ให้ผู้จำหน่ายรายเดียวกันในเวลาใกล้เคียงกัน (${diffDays.toFixed(0)} วัน) ควรตรวจสอบว่าไม่ใช่การเบิกซ้ำ`,
          });
        }
      }
    }
  }

  return alerts;
}

/**
 * Detects unusual expense spikes (> 200% or > 2.5 sigma).
 */
export function detectExpenseSpikes(params: {
  accountCode: string;
  accountName: string;
  currentPeriod: string;
  currentAmount: number;
  historicalAmounts: number[];
}): ExpenseAnomalyAlert | null {
  const { accountCode, accountName, currentPeriod, currentAmount, historicalAmounts } = params;

  if (historicalAmounts.length < 3 || currentAmount <= 0) return null;

  const n = historicalAmounts.length;
  const mean = historicalAmounts.reduce((s, v) => s + v, 0) / n;

  if (mean <= 0) return null;

  const variance =
    historicalAmounts.reduce((s, v) => s + Math.pow(v - mean, 2), 0) / (n > 1 ? n - 1 : 1);
  const stdDev = Math.sqrt(variance);

  const zScore = stdDev > 0 ? (currentAmount - mean) / stdDev : 0;
  const spikePercent = Math.round(((currentAmount - mean) / mean) * 100);

  if (spikePercent >= 150 || zScore >= 2.5) {
    return {
      accountCode,
      accountName,
      currentAmount,
      averageAmount: Math.round(mean * 100) / 100,
      spikePercent,
      zScore: Math.round(zScore * 100) / 100,
      period: currentPeriod,
      adviceTh: `ค่าใช้จ่ายกระโดดขึ้น ${spikePercent}% เมื่อเทียบกับค่าเฉลี่ยย้อนหลัง (฿${mean.toLocaleString()}) ควรตรวจสอบใบแจ้งหนี้และรายการผิดปกติ`,
    };
  }

  return null;
}

/**
 * Smart OCR Parser for Thai Bank Mobile Slips (PromptPay / Transfer).
 */
export function parseSmartSlipOrReceipt(rawText: string): SmartSlipResult {
  const lower = rawText.toLowerCase();

  // Detect bank
  let bankName = "ธนาคารทั่วไป";
  if (lower.includes("kasikorn") || lower.includes("kbank") || lower.includes("กสิกร")) {
    bankName = "ธนาคารกสิกรไทย (KBank)";
  } else if (lower.includes("scb") || lower.includes("ไทยพาณิชย์")) {
    bankName = "ธนาคารไทยพาณิชย์ (SCB)";
  } else if (lower.includes("bangkok") || lower.includes("bbl") || lower.includes("กรุงเทพ")) {
    bankName = "ธนาคารกรุงเทพ (BBL)";
  } else if (lower.includes("krungthai") || lower.includes("ktb") || lower.includes("กรุงไทย")) {
    bankName = "ธนาคารกรุงไทย (KTB)";
  }

  // Detect Amount
  let amount = 0;
  const amtMatch = rawText.match(/(?:จำนวนเงิน|ยอดเงิน|amount|baht|thb)[\s:]*([0-9,]+\.?[0-9]*)/i);
  if (amtMatch) {
    amount = parseFloat(amtMatch[1].replace(/,/g, "")) || 0;
  } else {
    // Look for standalone currency numbers
    const numMatches = rawText.match(/\b\d{1,3}(?:,\d{3})*(?:\.\d{2})\b/g);
    if (numMatches && numMatches.length > 0) {
      amount = parseFloat(numMatches[0].replace(/,/g, "")) || 0;
    }
  }

  // Detect Date
  let date = new Date().toISOString().split("T")[0];
  const dateMatch = rawText.match(/(\d{1,2})[\/\-\.](\d{1,2})[\/\-\.](\d{4})/);
  if (dateMatch) {
    const d = dateMatch[1].padStart(2, "0");
    const m = dateMatch[2].padStart(2, "0");
    let y = parseInt(dateMatch[3], 10);
    if (y > 2500) y -= 543;
    date = `${y}-${m}-${d}`;
  }

  // Detect Reference
  let ref: string | undefined = undefined;
  const refMatch = rawText.match(/(?:รหัสอ้างอิง|ref(?:\s*no)?|transaction\s*id)[\s:]*([A-Za-z0-9]{8,25})/i);
  if (refMatch) {
    ref = refMatch[1];
  }

  const isReceipt = lower.includes("โอนเงินเข้า") || lower.includes("รับเงิน") || lower.includes("deposit");
  const type: "RV" | "PV" = isReceipt ? "RV" : "PV";

  const draftJournal = isReceipt
    ? {
        type,
        description: `รับเงินโอนสลิป ${bankName} ${ref ? `Ref: ${ref}` : ""}`.trim(),
        lines: [
          {
            accountCode: "111201",
            debit: amount,
            credit: 0,
            description: "เงินฝากธนาคาร",
          },
          {
            accountCode: "219101",
            debit: 0,
            credit: amount,
            description: "เงินโอนรับพักรอเคลียร์ (จากสลิปโอนเงิน)",
          },
        ],
      }
    : {
        type,
        description: `จ่ายเงินตามสลิปโอนเงิน ${bankName} ${ref ? `Ref: ${ref}` : ""}`.trim(),
        lines: [
          {
            accountCode: "590101",
            debit: amount,
            credit: 0,
            description: "ค่าใช้จ่ายตามสลิปโอนเงิน",
          },
          {
            accountCode: "111201",
            debit: 0,
            credit: amount,
            description: "ตัดเงินฝากธนาคาร",
          },
        ],
      };

  return {
    bankName,
    transactionDate: date,
    amount: Math.round(amount * 100) / 100,
    referenceNo: ref,
    channel: lower.includes("promptpay") || lower.includes("พร้อมเพย์") ? "promptpay" : "transfer",
    draftJournal,
  };
}

/**
 * 12-Point Automated Pre-Closing Checklist.
 */
export function runPreClosingChecklist(params: {
  cashOnHandBalance: number;
  isBankReconciled: boolean;
  unbalancedJournalsCount: number;
  draftVouchersCount: number;
  depreciationPosted: boolean;
  stockCountAdjustmentDone: boolean;
  vatClosingDone: boolean;
  pendingWhtCertificatesCount: number;
  unclearedSuspenseCount: number;
  overdueArDays90Count: number;
  branchIntercompanyMatched: boolean;
  taxProvisionEstimated: boolean;
}): PreClosingChecklistResult {
  const items: PreClosingChecklistItem[] = [
    {
      id: 1,
      code: "CASH_NEGATIVE",
      titleTh: "เงินสดในมือไม่ติดลบ (Cash on Hand Balance)",
      titleEn: "Cash on Hand Balance Check",
      status: params.cashOnHandBalance < 0 ? "CRITICAL" : "PASS",
      detailTh:
        params.cashOnHandBalance < 0
          ? `เงินสดในมือติดลบ (฿${params.cashOnHandBalance.toLocaleString()}) ขัดต่อข้อเท็จจริงทางกายภาพและเสี่ยงสรรพากรประเมิน`
          : `เงินสดคงเหลือปกติ (฿${params.cashOnHandBalance.toLocaleString()})`,
    },
    {
      id: 2,
      code: "BANK_RECON",
      titleTh: "การกระทบยอดเงินฝากธนาคาร (Bank Reconciliation)",
      titleEn: "Bank Reconciliation Status",
      status: params.isBankReconciled ? "PASS" : "WARNING",
      detailTh: params.isBankReconciled
        ? "ยอดตาม Statement และสมุดบัญชีกระทบยอดลงตัวเรียบร้อย"
        : "ยังมียอดเงินฝากระหว่างทางหรือเช็คค้างจ่ายที่ยังไม่ได้กระทบยอด",
    },
    {
      id: 3,
      code: "GL_BALANCE",
      titleTh: "สมดุลเดบิต-เครดิตในสมุดรายวัน (Debit = Credit)",
      titleEn: "Balanced Journal Entries",
      status: params.unbalancedJournalsCount > 0 ? "CRITICAL" : "PASS",
      detailTh:
        params.unbalancedJournalsCount > 0
          ? `พบใบสำคัญที่ยอดเดบิตไม่เท่ากับเครดิตจำนวน ${params.unbalancedJournalsCount} ใบ`
          : "ใบสำคัญทุกใบมียอดเดบิต = เครดิต สมดุล 100%",
    },
    {
      id: 4,
      code: "DRAFT_POSTING",
      titleTh: "ใบสำคัญฉบับร่างค้างผ่านรายการ (Unposted Drafts)",
      titleEn: "Pending Draft Vouchers",
      status: params.draftVouchersCount > 0 ? "WARNING" : "PASS",
      detailTh:
        params.draftVouchersCount > 0
          ? `มีใบสำคัญฉบับร่างค้างอยู่ ${params.draftVouchersCount} ใบที่ยังไม่ได้โพสต์เข้า GL`
          : "ไม่มีใบสำคัญฉบับร่างค้าง ทุกใบผ่านรายการครบถ้วน",
    },
    {
      id: 5,
      code: "DEPRECIATION_POSTED",
      titleTh: "คำนวณและบันทึกค่าเสื่อมราคาประจำงวด (Fixed Asset Depreciation)",
      titleEn: "Depreciation Voucher Posted",
      status: params.depreciationPosted ? "PASS" : "CRITICAL",
      detailTh: params.depreciationPosted
        ? "บันทึกค่าเสื่อมราคาสินทรัพย์ถาวรเข้า GL เรียบร้อยแล้ว"
        : "ยังไม่ได้ผ่านรายการใบสำคัญค่าเสื่อมราคาประจำงวด",
    },
    {
      id: 6,
      code: "STOCK_COUNT",
      titleTh: "ตรวจนับและปรับปรุงผลต่างสต๊อก (Physical Stock Variance)",
      titleEn: "Stock Count Variance Adjustment",
      status: params.stockCountAdjustmentDone ? "PASS" : "WARNING",
      detailTh: params.stockCountAdjustmentDone
        ? "ปรับปรุงผลต่างสต๊อกเข้าต้นทุนเรียบร้อย"
        : "ยังไม่ได้บันทึกผลต่างการตรวจนับสต๊อกสินค้าปลายงวด",
    },
    {
      id: 7,
      code: "VAT_CLOSING",
      titleTh: "โอนปิดภาษีซื้อ-ภาษีขายประจำเดือน (Monthly VAT Closing)",
      titleEn: "Monthly VAT Closing Journal",
      status: params.vatClosingDone ? "PASS" : "CRITICAL",
      detailTh: params.vatClosingDone
        ? "โอนปิดภาษีซื้อ-ภาษีขายเป็นเจ้าหนี้/ลูกหนี้กรมสรรพากรเรียบร้อย"
        : "ยังไม่ได้ทำรายการโอนปิดภาษีซื้อ-ภาษีขายประจำเดือน",
    },
    {
      id: 8,
      code: "WHT_50_TWI",
      titleTh: "ออกหนังสือรับรองหัก ณ ที่จ่าย 50 ทวิ ครบถ้วน",
      titleEn: "WHT 50 Twi Certificates Issued",
      status: params.pendingWhtCertificatesCount > 0 ? "WARNING" : "PASS",
      detailTh:
        params.pendingWhtCertificatesCount > 0
          ? `มีรายการภาษีหัก ณ ที่จ่ายรอพิมพ์ใบ 50 ทวิ ${params.pendingWhtCertificatesCount} รายการ`
          : "ออกหนังสือรับรอง 50 ทวิครบถ้วนแล้ว",
    },
    {
      id: 9,
      code: "SUSPENSE_CLEARING",
      titleTh: "เคลียร์บัญชีพักและเงินโอนไม่ทราบผู้โอน (Suspense Accounts)",
      titleEn: "Suspense Accounts Cleared",
      status: params.unclearedSuspenseCount > 0 ? "WARNING" : "PASS",
      detailTh:
        params.unclearedSuspenseCount > 0
          ? `พบบัญชีพักมียอดคงเหลือค้าง ${params.unclearedSuspenseCount} บัญชี`
          : "บัญชีพักทุกบัญชีถูกโอนเคลียร์เรียบร้อย",
    },
    {
      id: 10,
      code: "AR_AGING_RESERVE",
      titleTh: "พิจารณาตั้งสำรองหนี้สงสัยจะสูญสำหรับลูกหนี้ค้างนาน > 90 วัน",
      titleEn: "Overdue AR Allowance Check",
      status: params.overdueArDays90Count > 0 ? "WARNING" : "PASS",
      detailTh:
        params.overdueArDays90Count > 0
          ? `มีลูกหนี้ค้างชำระเกิน 90 วันจำนวน ${params.overdueArDays90Count} รายการ ควรพิจารณาตั้งค่าเผื่อหนี้สงสัยจะสูญ`
          : "ไม่มีลูกหนี้ค้างชำระเกิน 90 วัน",
    },
    {
      id: 11,
      code: "BRANCH_INTERCOMPANY",
      titleTh: "กระทบยอดบัญชีระหว่างสาขา/บริษัทในเครือ (Inter-branch Matching)",
      titleEn: "Inter-branch Balance Matching",
      status: params.branchIntercompanyMatched ? "PASS" : "WARNING",
      detailTh: params.branchIntercompanyMatched
        ? "ยอดคงเหลือระหว่างสาขาสมดุลตรงกัน"
        : "พบผลต่างยอดคงค้างระหว่างสาขา",
    },
    {
      id: 12,
      code: "TAX_PROVISION",
      titleTh: "ประมาณการภาษีเงินได้นิติบุคคลประจำงวด (Corporate Income Tax Provision)",
      titleEn: "Corporate Income Tax Provision",
      status: params.taxProvisionEstimated ? "PASS" : "WARNING",
      detailTh: params.taxProvisionEstimated
        ? "คำนวณและบันทึกประมาณการภาษีเงินได้นิติบุคคลเรียบร้อย"
        : "ควรประมาณการภาระภาษีเงินได้นิติบุคคล (20% หรืออัตรา SME) ก่อนปิดงบ",
    },
  ];

  let passCount = 0;
  let warningCount = 0;
  let criticalCount = 0;

  for (const item of items) {
    if (item.status === "PASS") passCount++;
    else if (item.status === "WARNING") warningCount++;
    else if (item.status === "CRITICAL") criticalCount++;
  }

  // Score calculation
  const score = Math.round(((passCount * 1.0 + warningCount * 0.5) / items.length) * 100);
  const overallStatus: ChecklistStatus =
    criticalCount > 0 ? "CRITICAL" : warningCount > 0 ? "WARNING" : "PASS";

  return {
    readinessScore: score,
    overallStatus,
    passCount,
    warningCount,
    criticalCount,
    items,
  };
}
