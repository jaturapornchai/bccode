import { type GLAccount, displayAmountUnits, formatAmount } from "./general-ledger";

export type AnomalySeverity = "critical" | "warning" | "info";

export type AnomalyCategory =
  | "trial_balance_imbalance"
  | "normal_balance_violation"
  | "cash_bank_overdrawn"
  | "inactive_account_with_balance"
  | "suspense_account_uncleared"
  | "dormant_account";

export interface GLAnomaly {
  id: string;
  category: AnomalyCategory;
  severity: AnomalySeverity;
  accountCode?: string;
  accountName?: string;
  accountType?: string;
  amount?: string;
  messageTh: string;
  messageEn: string;
  recommendationTh: string;
  recommendationEn: string;
  canDrillLedger: boolean;
}

export interface GLHealthReport {
  score: number; // 0 - 100
  status: "excellent" | "good" | "warning" | "critical";
  statusTextTh: string;
  statusTextEn: string;
  summary: {
    totalAccountsAnalyzed: number;
    criticalCount: number;
    warningCount: number;
    infoCount: number;
    totalDebitEnding: string;
    totalCreditEnding: string;
    imbalanceAmount: string;
  };
  anomalies: GLAnomaly[];
}

/** Check if account is a recognized contra-asset (normal balance is Credit) */
function isContraAsset(name: string, normalBalance?: string): boolean {
  if (normalBalance === "credit") return true;
  const n = name.toLowerCase();
  return (
    n.includes("ค่าเสื่อมราคาสะสม") ||
    n.includes("accumulated depreciation") ||
    n.includes("ค่าตัดจำหน่ายสะสม") ||
    n.includes("accumulated amortization") ||
    n.includes("ค่าเผื่อหนี้สงสัยจะสูญ") ||
    n.includes("ค่าเผื่อผลขาดทุน") ||
    n.includes("allowance for doubtful") ||
    n.includes("allowance for expected credit") ||
    n.includes("ค่าเผื่อการลดมูลค่า") ||
    n.includes("allowance for devaluation") ||
    n.includes("allowance for inventory")
  );
}

/** Check if equity account is expected to carry a Debit balance (retained loss / deficit / dividends) */
function isContraEquity(name: string, normalBalance?: string): boolean {
  if (normalBalance === "debit") return true;
  const n = name.toLowerCase();
  return (
    n.includes("ขาดทุนสะสม") ||
    n.includes("accumulated loss") ||
    n.includes("retained loss") ||
    n.includes("deficit") ||
    n.includes("หุ้นทุนซื้อคืน") ||
    n.includes("treasury") ||
    n.includes("เงินปันผล") ||
    n.includes("dividend")
  );
}

/** Check if revenue account is contra-revenue (Sales returns / discounts) */
function isContraRevenue(name: string, normalBalance?: string): boolean {
  if (normalBalance === "debit") return true;
  const n = name.toLowerCase();
  return (
    n.includes("รับคืน") ||
    n.includes("sales return") ||
    n.includes("ส่วนลดจ่าย") ||
    n.includes("sales discount")
  );
}

/** Check if expense account is contra-expense (Purchase returns / discounts) */
function isContraExpense(name: string, normalBalance?: string): boolean {
  if (normalBalance === "credit") return true;
  const n = name.toLowerCase();
  return (
    n.includes("ส่งคืน") ||
    n.includes("purchase return") ||
    n.includes("ส่วนลดรับ") ||
    n.includes("purchase discount")
  );
}

/** Check if account is cash or bank account */
function isCashOrBank(code: string, name: string, isCashFlag?: boolean): { isCash: boolean; isPureCashInHand: boolean } {
  if (isCashFlag) {
    const isHand = code.startsWith("1111") || name.includes("เงินสดในมือ") || name.includes("cash on hand") || name.includes("เงินสดย่อย");
    return { isCash: true, isPureCashInHand: isHand };
  }
  const cleanCode = code.replace(/[^0-9]/g, "");
  const isHand = cleanCode.startsWith("1111") || name.includes("เงินสดในมือ") || name.includes("cash on hand") || name.includes("เงินสดย่อย");
  const isBank = cleanCode.startsWith("1112") || cleanCode.startsWith("1113") || name.includes("เงินฝาก") || name.includes("bank");
  return {
    isCash: isHand || isBank,
    isPureCashInHand: isHand,
  };
}

/** Check if account is a clearing or suspense account */
function isSuspenseAccount(name: string): boolean {
  const n = name.toLowerCase();
  return (
    n.includes("พัก") ||
    n.includes("clearing") ||
    n.includes("suspense") ||
    n.includes("รอตรวจสอบ") ||
    n.includes("รอกระทบยอด") ||
    n.includes("unclassified")
  );
}

/**
 * Runs comprehensive GL Health & Anomaly Audit
 */
export function auditGLHealth({
  reportRows = [],
  totals = {},
  accounts = [],
}: {
  reportRows?: Record<string, string | undefined>[];
  totals?: Record<string, string>;
  accounts?: GLAccount[];
}): GLHealthReport {
  const anomalies: GLAnomaly[] = [];
  const accountsMap = new Map<string, GLAccount>();
  for (const acc of accounts) {
    accountsMap.set(acc.accountcode, acc);
  }

  // 1. Check Trial Balance debit vs credit balance
  let endingDebitSum = 0n;
  let endingCreditSum = 0n;
  let hasParsedEnding = false;

  for (const row of reportRows) {
    try {
      if (row.endingdebit !== undefined || row.endingcredit !== undefined) {
        endingDebitSum += displayAmountUnits(row.endingdebit || "0");
        endingCreditSum += displayAmountUnits(row.endingcredit || "0");
        hasParsedEnding = true;
      }
    } catch {
      // Ignore unparseable
    }
  }

  // Check from totals if not found in rows
  if (!hasParsedEnding) {
    try {
      if (totals.debit || totals.credit) {
        endingDebitSum = displayAmountUnits(totals.debit || "0");
        endingCreditSum = displayAmountUnits(totals.credit || "0");
      }
    } catch {
      // Ignore
    }
  }

  const imbalanceUnits = endingDebitSum - endingCreditSum;
  const absImbalance = imbalanceUnits < 0n ? -imbalanceUnits : imbalanceUnits;

  if (absImbalance !== 0n) {
    const formattedDiff = formatAmount(
      ((Number(absImbalance) / 100000000)).toFixed(2)
    );
    anomalies.push({
      id: "imbalance-trial-balance",
      category: "trial_balance_imbalance",
      severity: "critical",
      amount: formattedDiff,
      messageTh: `งบทดลองมียอดเดบิตและเครดิตไม่สมดุล มีผลต่าง ${formattedDiff} บาท`,
      messageEn: `Trial Balance is out of balance by ${formattedDiff} THB (Debit != Credit)`,
      recommendationTh: "ตรวจสอบรายการสมุดรายวันฉบับร่างหรือรายการที่เพิ่งผ่านรายการ และรันคำนวณยอดยกไปใหม่",
      recommendationEn: "Review recently posted vouchers and re-run GL balance calculation.",
      canDrillLedger: false,
    });
  }

  // 2. Analyze individual account rows
  let totalAccountsAnalyzed = 0;

  for (const row of reportRows) {
    const code = row.accountcode;
    if (!code || code === "__current_earnings__") continue;
    totalAccountsAnalyzed++;

    const accMaster = accountsMap.get(code);
    const name = row.accountname || accMaster?.names?.find((n) => n.code === "th")?.name || code;
    const type = row.accounttype || accMaster?.accounttype || "asset";
    const normal = accMaster?.normalbalance || (["asset", "expense"].includes(type) ? "debit" : "credit");
    const isCashFlag = accMaster?.iscash;

    let endingDebit = 0n;
    let endingCredit = 0n;
    let periodDebit = 0n;
    let periodCredit = 0n;

    try {
      endingDebit = displayAmountUnits(row.endingdebit || "0");
      endingCredit = displayAmountUnits(row.endingcredit || "0");
      periodDebit = displayAmountUnits(row.debit || "0");
      periodCredit = displayAmountUnits(row.credit || "0");
    } catch {
      continue;
    }

    const netEndingDebit = endingDebit - endingCredit; // Positive = Debit balance, Negative = Credit balance
    const absEndingVal = netEndingDebit < 0n ? -netEndingDebit : netEndingDebit;
    const formattedEnding = formatAmount((Number(absEndingVal) / 100000000).toFixed(2));

    // A. Cash & Bank Checks
    const { isCash, isPureCashInHand } = isCashOrBank(code, name, isCashFlag);
    if (isCash) {
      if (netEndingDebit < 0n) {
        if (isPureCashInHand) {
          anomalies.push({
            id: `cash-overdrawn-${code}`,
            category: "cash_bank_overdrawn",
            severity: "critical",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: formattedEnding,
            messageTh: `บัญชีเงินสดในมือ (${code} ${name}) มียอดเครดิตคงค้าง ${formattedEnding} บาท ซึ่งเป็นไปไม่ได้ในทางปฏิบัติ`,
            messageEn: `Cash on hand (${code} ${name}) has a negative/credit balance of ${formattedEnding} THB, which is abnormal in practice.`,
            recommendationTh: "ตรวจสอบว่ามีการลงรายจ่ายล่วงหน้าก่อนได้รับเงิน หรือลืมบันทึกรายรับ/เงินทดรองจ่าย ตรวจนับเงินสดในมือทันที",
            recommendationEn: "Count physical cash and check for missing revenue/advance entries or backdated expenses.",
            canDrillLedger: true,
          });
        } else {
          anomalies.push({
            id: `bank-overdrawn-${code}`,
            category: "cash_bank_overdrawn",
            severity: "warning",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: formattedEnding,
            messageTh: `บัญชีเงินฝากธนาคาร (${code} ${name}) มียอดติดลบ/เครดิตคงค้าง ${formattedEnding} บาท (อาจเป็นเงินเบิกเกินบัญชี O/D)`,
            messageEn: `Bank account (${code} ${name}) has an overdrawn/credit balance of ${formattedEnding} THB.`,
            recommendationTh: "กระทบยอดกับรายงานธนาคาร (Bank Statement) หากเป็นวงเงินเบิกเกินบัญชี (O/D) ควรจัดประเภทเป็นหนี้สินในงบการเงิน",
            recommendationEn: "Reconcile with bank statement. If this is an overdraft (O/D), reclassify to current liability in financial statements.",
            canDrillLedger: true,
          });
        }
      }
    }

    // B. Normal Balance Violations
    if (!isCash) {
      if (type === "asset") {
        const contra = isContraAsset(name, normal);
        if (!contra && netEndingDebit < 0n && absEndingVal > 0n) {
          anomalies.push({
            id: `norm-viol-asset-${code}`,
            category: "normal_balance_violation",
            severity: "warning",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: formattedEnding,
            messageTh: `สินทรัพย์ (${code} ${name}) มียอดคงเหลือด้านเครดิต ${formattedEnding} บาท (ผิด Normal Balance)`,
            messageEn: `Asset account (${code} ${name}) has a credit balance of ${formattedEnding} THB (violates normal debit balance).`,
            recommendationTh: "ตรวจสอบการบันทึกรายการจ่ายซ้ำซ้อน หรือยังไม่ได้บันทึกรับสินทรัพย์เข้ามาในระบบ",
            recommendationEn: "Check for duplicate disposals or missing asset receipt vouchers.",
            canDrillLedger: true,
          });
        }
      } else if (type === "liability") {
        if (netEndingDebit > 0n && absEndingVal > 0n) {
          anomalies.push({
            id: `norm-viol-liab-${code}`,
            category: "normal_balance_violation",
            severity: "warning",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: formattedEnding,
            messageTh: `หนี้สิน (${code} ${name}) มียอดคงเหลือด้านเดบิต ${formattedEnding} บาท (ชำระเกินยอดหนี้)`,
            messageEn: `Liability account (${code} ${name}) has a debit balance of ${formattedEnding} THB (potential overpayment).`,
            recommendationTh: "ตรวจสอบว่ามีการจ่ายชำระหนี้เกิน หรือบันทึกเงินจ่ายล่วงหน้าผิดบัญชี",
            recommendationEn: "Check if payments exceeded invoices or if advances were misposted.",
            canDrillLedger: true,
          });
        }
      } else if (type === "equity") {
        const contra = isContraEquity(name, normal);
        if (!contra && netEndingDebit > 0n && absEndingVal > 0n) {
          anomalies.push({
            id: `norm-viol-equity-${code}`,
            category: "normal_balance_violation",
            severity: "warning",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: formattedEnding,
            messageTh: `ส่วนของเจ้าของ (${code} ${name}) มียอดคงเหลือด้านเดบิต ${formattedEnding} บาท`,
            messageEn: `Equity account (${code} ${name}) has an unexpected debit balance of ${formattedEnding} THB.`,
            recommendationTh: "ตรวจสอบว่ามีการบันทึกเงินถอนหรือรายการลดทุนถูกต้องตามมติที่ประชุมหรือไม่",
            recommendationEn: "Verify capital reductions or shareholder withdrawals against board resolutions.",
            canDrillLedger: true,
          });
        }
      } else if (type === "income") {
        const contra = isContraRevenue(name, normal);
        // Revenue is normally Credit. If net period movement is Debit > Credit
        if (!contra && periodDebit > periodCredit) {
          const diffVal = formatAmount((Number(periodDebit - periodCredit) / 100000000).toFixed(2));
          anomalies.push({
            id: `norm-viol-income-${code}`,
            category: "normal_balance_violation",
            severity: "warning",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: diffVal,
            messageTh: `รายได้ (${code} ${name}) มียอดเดบิตเคลื่อนไหวมากกว่าเครดิต (${diffVal} บาท)`,
            messageEn: `Revenue account (${code} ${name}) has net debit movement of ${diffVal} THB.`,
            recommendationTh: "ตรวจสอบว่ามีการกลับรายการรายได้ผิดพลาด หรือบันทึกค่าใช้จ่ายเข้ามาในหมวดรายได้",
            recommendationEn: "Check for errant income reversal entries or misclassified expenses.",
            canDrillLedger: true,
          });
        }
      } else if (type === "expense") {
        const contra = isContraExpense(name, normal);
        // Expense is normally Debit. If net period movement is Credit > Debit
        if (!contra && periodCredit > periodDebit) {
          const diffVal = formatAmount((Number(periodCredit - periodDebit) / 100000000).toFixed(2));
          anomalies.push({
            id: `norm-viol-expense-${code}`,
            category: "normal_balance_violation",
            severity: "warning",
            accountCode: code,
            accountName: name,
            accountType: type,
            amount: diffVal,
            messageTh: `ค่าใช้จ่าย (${code} ${name}) มียอดเครดิตเคลื่อนไหวมากกว่าเดบิต (${diffVal} บาท)`,
            messageEn: `Expense account (${code} ${name}) has net credit movement of ${diffVal} THB.`,
            recommendationTh: "ตรวจสอบว่ามีการรับคืนค่าใช้จ่าย หรือลงบัญชีฝั่งผิด",
            recommendationEn: "Check for expense refunds or incorrect side postings.",
            canDrillLedger: true,
          });
        }
      }
    }

    // C. Inactive Account with Non-zero Balance
    if (accMaster && accMaster.isactive === false && absEndingVal > 0n) {
      anomalies.push({
        id: `inactive-balance-${code}`,
        category: "inactive_account_with_balance",
        severity: "warning",
        accountCode: code,
        accountName: name,
        accountType: type,
        amount: formattedEnding,
        messageTh: `บัญชีที่ปิดใช้งานแล้ว (${code} ${name}) ยังมียอดคงค้าง ${formattedEnding} บาท`,
        messageEn: `Deactivated account (${code} ${name}) still has an outstanding balance of ${formattedEnding} THB.`,
        recommendationTh: "โอนย้ายยอดคงเหลือไปยังบัญชีที่ใช้งานอยู่ หรือเปิดใช้งานบัญชีหากยังจำเป็นต้องลงรายการ",
        recommendationEn: "Transfer remaining balance to an active account or reactivate the account.",
        canDrillLedger: true,
      });
    }

    // D. Uncleared Suspense / Clearing Account
    if (isSuspenseAccount(name) && absEndingVal > 0n) {
      anomalies.push({
        id: `suspense-uncleared-${code}`,
        category: "suspense_account_uncleared",
        severity: "info",
        accountCode: code,
        accountName: name,
        accountType: type,
        amount: formattedEnding,
        messageTh: `บัญชีพักรอตรวจสอบ (${code} ${name}) มียอดค้างชำระ/รอเคลียร์ ${formattedEnding} บาท`,
        messageEn: `Clearing/Suspense account (${code} ${name}) has an uncleared balance of ${formattedEnding} THB.`,
        recommendationTh: "กระทบยอดเอกสารและปรับปรุงรายการออกจากบัญชีพักเข้าบัญชีจริงก่อนปิดรอบบัญชี",
        recommendationEn: "Reconcile documentation and clear suspense items before closing period.",
        canDrillLedger: true,
      });
    }

    // E. Dormant Account Check (Only if total debit & credit in period are both zero, and had opening balance)
    if (periodDebit === 0n && periodCredit === 0n && absEndingVal > 0n && accMaster?.allowposting) {
      anomalies.push({
        id: `dormant-${code}`,
        category: "dormant_account",
        severity: "info",
        accountCode: code,
        accountName: name,
        accountType: type,
        amount: formattedEnding,
        messageTh: `บัญชี (${code} ${name}) ไม่มีการเคลื่อนไหวเลยในงวดนี้ (ยอดยกมาค้างอยู่ ${formattedEnding} บาท)`,
        messageEn: `Account (${code} ${name}) has had no debit/credit movements this period.`,
        recommendationTh: "ตรวจสอบว่าบัญชีนี้ยังมีความจำเป็นในธุรกิจหรือไม่ หากไม่มีความจำเป็นสามารถวางแผนปิดบัญชี",
        recommendationEn: "Evaluate if this account is still needed or schedule for consolidation.",
        canDrillLedger: true,
      });
    }
  }

  // Calculate counts and deductions
  let criticalCount = 0;
  let warningCount = 0;
  let infoCount = 0;

  for (const a of anomalies) {
    if (a.severity === "critical") criticalCount++;
    else if (a.severity === "warning") warningCount++;
    else if (a.severity === "info") infoCount++;
  }

  // Score calculation:
  // Initial 100.
  // Critical: -25 pts each
  // Warning: -8 pts each
  // Info: -1 pt each (cap info deduction at 10 pts)
  const infoDeduction = Math.min(infoCount * 1, 10);
  const totalDeduction = criticalCount * 25 + warningCount * 8 + infoDeduction;
  const score = Math.max(0, 100 - totalDeduction);

  let status: "excellent" | "good" | "warning" | "critical" = "excellent";
  let statusTextTh = "ยอดเยี่ยม สมบูรณ์ 100%";
  let statusTextEn = "Excellent, fully balanced";

  if (score < 50 || criticalCount > 0) {
    status = "critical";
    statusTextTh = "วิกฤติ ต้องตรวจสอบทันที";
    statusTextEn = "Critical, requires immediate audit";
  } else if (score < 75 || warningCount >= 3) {
    status = "warning";
    statusTextTh = "ควรระวัง มียอดผิดฝั่งหรือค้างเคลียร์";
    statusTextEn = "Warning, contains balance anomalies";
  } else if (score < 95) {
    status = "good";
    statusTextTh = "ดี มีข้อสังเกตเล็กน้อย";
    statusTextEn = "Good, minor remarks";
  }

  const endingDebitFmt = formatAmount((Number(endingDebitSum) / 100000000).toFixed(2));
  const endingCreditFmt = formatAmount((Number(endingCreditSum) / 100000000).toFixed(2));
  const imbalanceFmt = formatAmount((Number(absImbalance) / 100000000).toFixed(2));

  return {
    score,
    status,
    statusTextTh,
    statusTextEn,
    summary: {
      totalAccountsAnalyzed,
      criticalCount,
      warningCount,
      infoCount,
      totalDebitEnding: endingDebitFmt,
      totalCreditEnding: endingCreditFmt,
      imbalanceAmount: imbalanceFmt,
    },
    anomalies,
  };
}
