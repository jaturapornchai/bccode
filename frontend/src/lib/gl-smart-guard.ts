import {
  amountUnits,
  amountString,
  type GLAccount,
  type GLLine,
} from "./general-ledger";

const PENNY_FACTOR = 1000000n; // 10^6 for 2 decimals (1 Satang)

export type TaxType = "input_tax" | "output_tax";

export interface VatAnalysis {
  hasVatLine: boolean;
  vatLineIndex: number;
  vatAccountCode: string;
  vatType: TaxType | null;
  actualVatUnits: bigint;
  baseAmountUnits: bigint;
  expectedVatUnits: bigint;
  varianceUnits: bigint;
  isExactVat: boolean;
  isCloseVat: boolean; // within 5 Satang (e.g. invoice rounding variance)
  actualVatFormatted: string;
  expectedVatFormatted: string;
  suggestedVatFormatted: string;
  suggestedVatType: TaxType | null;
  suggestedVatUnits: bigint;
}

export interface GLBalanceAnalysis {
  totalDebitUnits: bigint;
  totalCreditUnits: bigint;
  differenceUnits: bigint; // debit - credit
  isBalanced: boolean;
  balanceStatus: "balanced" | "debit_surplus" | "credit_surplus";
  balanceDifferenceFormatted: string;
  vat: VatAnalysis;
}

// Withholding tax (ภาษีเงินได้ถูกหัก/ค้างจ่าย, ภ.ง.ด.51 จ่ายล่วงหน้า) is not VAT even when it sits next to VAT in the chart.
const NOT_VAT_NAME = /หัก\s*ณ\s*ที่จ่าย|ภาษีเงินได้|withholding/i;
// Undue VAT (ภาษีซื้อ/ภาษีขายยังไม่ถึงกำหนด) waits for the tax point; it is not the account for a normal VAT line.
const UNDUE_VAT_NAME = /ยังไม่ถึงกำหนด|undue/i;

/**
 * Checks whether a line is a Thai VAT line (ภาษีซื้อ / ภาษีขาย) from the account name, account type and line
 * description only. Account codes are never a condition: every company keeps its own chart (ลุงจืด 2026-09-24).
 * Income/expense accounts are never the VAT line, even when named e.g. "รายได้ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม".
 */
export function isVatAccount(
  accountName: string,
  lineDescription: string,
  accountType = ""
): { isVat: boolean; type: TaxType | null } {
  const name = (accountName || "").trim().toLowerCase();
  const desc = (lineDescription || "").trim().toLowerCase();
  const none = { isVat: false, type: null };

  if (NOT_VAT_NAME.test(name) || accountType === "income" || accountType === "expense") {
    return none;
  }
  if (/ภาษีซื้อ|input\s*(tax|vat)|vat\s*buy/i.test(name) || /ภาษีซื้อ/i.test(desc)) {
    return { isVat: true, type: "input_tax" };
  }
  if (/ภาษีขาย|output\s*(tax|vat)|vat\s*sale/i.test(name) || /ภาษีขาย/i.test(desc)) {
    return { isVat: true, type: "output_tax" };
  }
  if (/ภาษีมูลค่าเพิ่ม|\bvat\b/i.test(name) || /ภาษีมูลค่าเพิ่ม|vat\s*7%/i.test(desc)) {
    if (accountType === "asset") return { isVat: true, type: "input_tax" };
    if (accountType === "liability") return { isVat: true, type: "output_tax" };
    return { isVat: true, type: null };
  }

  return none;
}

/**
 * Computes 7% VAT rounded to decimal scale (default 2 decimals = 1 satang) using BigInt precision
 */
export function calculateVat7(baseAmountUnits: bigint, scale = 2): bigint {
  if (baseAmountUnits <= 0n) return 0n;
  const rawVatUnits = (baseAmountUnits * 7n) / 100n;
  const penny = 10n ** BigInt(8 - scale);
  const halfPenny = penny / 2n;
  return ((rawVatUnits + halfPenny) / penny) * penny;
}

/**
 * Analyzes journal lines for debit-credit balance and 7% VAT consistency
 */
export function analyzeGLTaxAndBalance(
  lines: GLLine[],
  accounts: GLAccount[] = [],
  scale = 2
): GLBalanceAnalysis {
  let totalDebitUnits = 0n;
  let totalCreditUnits = 0n;

  const accountsMap = new Map<string, GLAccount>();
  for (const acc of accounts) {
    accountsMap.set(acc.accountcode, acc);
  }

  const parsedLines = lines.map((line, index) => {
    let debit = 0n;
    let credit = 0n;
    try {
      if (line.debit && line.debit !== "0") debit = amountUnits(line.debit);
    } catch {
      debit = 0n;
    }
    try {
      if (line.credit && line.credit !== "0") credit = amountUnits(line.credit);
    } catch {
      credit = 0n;
    }
    totalDebitUnits += debit;
    totalCreditUnits += credit;

    const acc = accountsMap.get(line.accountcode);
    const accName = acc?.names?.[0]?.name || "";
    const vatInfo = isVatAccount(accName, line.description, acc?.accounttype ?? "");

    return {
      index,
      debit,
      credit,
      vatInfo,
    };
  });

  const differenceUnits = totalDebitUnits - totalCreditUnits;
  const isBalanced = differenceUnits === 0n;
  const balanceStatus: "balanced" | "debit_surplus" | "credit_surplus" =
    differenceUnits === 0n
      ? "balanced"
      : differenceUnits > 0n
      ? "debit_surplus"
      : "credit_surplus";

  const vatLine = parsedLines.find((p) => p.vatInfo.isVat);
  let hasVatLine = false;
  let vatLineIndex = -1;
  let vatAccountCode = "";
  let vatType: TaxType | null = null;
  let actualVatUnits = 0n;
  let baseAmountUnits = 0n;
  let expectedVatUnits = 0n;
  let varianceUnits = 0n;
  let isExactVat = false;
  let isCloseVat = false;

  let suggestedVatType: TaxType | null = null;
  let suggestedVatUnits = 0n;

  if (vatLine) {
    hasVatLine = true;
    vatLineIndex = vatLine.index;
    vatAccountCode = lines[vatLine.index].accountcode;
    vatType = vatLine.vatInfo.type;

    if (vatLine.debit > 0n) {
      actualVatUnits = vatLine.debit;
      vatType = vatType || "input_tax";
      baseAmountUnits = parsedLines
        .filter((p) => p.index !== vatLine.index && p.debit > 0n)
        .reduce((sum, p) => sum + p.debit, 0n);
    } else if (vatLine.credit > 0n) {
      actualVatUnits = vatLine.credit;
      vatType = vatType || "output_tax";
      baseAmountUnits = parsedLines
        .filter((p) => p.index !== vatLine.index && p.credit > 0n)
        .reduce((sum, p) => sum + p.credit, 0n);
    }

    if (baseAmountUnits > 0n) {
      expectedVatUnits = calculateVat7(baseAmountUnits, scale);
      varianceUnits = actualVatUnits - expectedVatUnits;
      isExactVat = varianceUnits === 0n;
      const absVar = varianceUnits < 0n ? -varianceUnits : varianceUnits;
      isCloseVat = absVar <= 5n * PENNY_FACTOR;
    }
  } else {
    const nonZeroDebits = parsedLines.filter((p) => p.debit > 0n);
    const nonZeroCredits = parsedLines.filter((p) => p.credit > 0n);

    if (nonZeroDebits.length > 0 && nonZeroCredits.length === 0) {
      suggestedVatType = "input_tax";
      const totalBase = nonZeroDebits.reduce((sum, p) => sum + p.debit, 0n);
      suggestedVatUnits = calculateVat7(totalBase, scale);
    } else if (nonZeroCredits.length > 0 && nonZeroDebits.length === 0) {
      suggestedVatType = "output_tax";
      const totalBase = nonZeroCredits.reduce((sum, p) => sum + p.credit, 0n);
      suggestedVatUnits = calculateVat7(totalBase, scale);
    } else if (nonZeroDebits.length > 0 && nonZeroCredits.length > 0) {
      if (totalDebitUnits > totalCreditUnits) {
        suggestedVatType = "output_tax";
        suggestedVatUnits = calculateVat7(totalDebitUnits - totalCreditUnits, scale);
      } else if (totalCreditUnits > totalDebitUnits) {
        suggestedVatType = "input_tax";
        suggestedVatUnits = calculateVat7(totalCreditUnits - totalDebitUnits, scale);
      }
    }
  }

  const formatUnits = (units: bigint) => {
    try {
      return amountString(units, scale);
    } catch {
      return "0.00";
    }
  };

  return {
    totalDebitUnits,
    totalCreditUnits,
    differenceUnits,
    isBalanced,
    balanceStatus,
    balanceDifferenceFormatted: formatUnits(
      differenceUnits < 0n ? -differenceUnits : differenceUnits
    ),
    vat: {
      hasVatLine,
      vatLineIndex,
      vatAccountCode,
      vatType,
      actualVatUnits,
      baseAmountUnits,
      expectedVatUnits,
      varianceUnits,
      isExactVat,
      isCloseVat,
      actualVatFormatted: formatUnits(actualVatUnits),
      expectedVatFormatted: formatUnits(expectedVatUnits),
      suggestedVatFormatted: formatUnits(suggestedVatUnits),
      suggestedVatType,
      suggestedVatUnits,
    },
  };
}

/**
 * Automatically adjusts lines so that totalDebit === totalCredit.
 * Modifies the deficient side's last active row.
 */
export function autoBalanceJournalLines(lines: GLLine[], scale = 2): GLLine[] {
  if (!lines || lines.length === 0) return lines;

  let totalDebit = 0n;
  let totalCredit = 0n;
  for (const line of lines) {
    try {
      if (line.debit && line.debit !== "0") totalDebit += amountUnits(line.debit);
    } catch {}
    try {
      if (line.credit && line.credit !== "0") totalCredit += amountUnits(line.credit);
    } catch {}
  }

  const diff = totalDebit - totalCredit;
  if (diff === 0n) return lines;

  const nextLines = lines.map((line) => ({ ...line }));

  if (diff > 0n) {
    let targetIndex = -1;
    for (let i = nextLines.length - 1; i >= 0; i--) {
      try {
        if (nextLines[i].credit && amountUnits(nextLines[i].credit) > 0n) {
          targetIndex = i;
          break;
        }
      } catch {}
    }

    if (targetIndex >= 0) {
      const currentCredit = amountUnits(nextLines[targetIndex].credit);
      nextLines[targetIndex].credit = amountString(currentCredit + diff, scale);
    } else {
      let emptyIdx = -1;
      for (let i = nextLines.length - 1; i >= 0; i--) {
        if (!nextLines[i].debit || nextLines[i].debit === "0") {
          emptyIdx = i;
          break;
        }
      }
      if (emptyIdx >= 0) {
        nextLines[emptyIdx].credit = amountString(diff, scale);
      } else {
        const lastIdx = nextLines.length - 1;
        nextLines[lastIdx].credit = amountString(diff, scale);
      }
    }
  } else {
    const neededDebit = -diff;
    let targetIndex = -1;
    for (let i = nextLines.length - 1; i >= 0; i--) {
      try {
        if (nextLines[i].debit && amountUnits(nextLines[i].debit) > 0n) {
          targetIndex = i;
          break;
        }
      } catch {}
    }

    if (targetIndex >= 0) {
      const currentDebit = amountUnits(nextLines[targetIndex].debit);
      nextLines[targetIndex].debit = amountString(currentDebit + neededDebit, scale);
    } else {
      let emptyIdx = -1;
      for (let i = nextLines.length - 1; i >= 0; i--) {
        if (!nextLines[i].credit || nextLines[i].credit === "0") {
          emptyIdx = i;
          break;
        }
      }
      if (emptyIdx >= 0) {
        nextLines[emptyIdx].debit = amountString(neededDebit, scale);
      } else {
        const lastIdx = nextLines.length - 1;
        nextLines[lastIdx].debit = amountString(neededDebit, scale);
      }
    }
  }

  return nextLines;
}

/**
 * Adjusts an existing VAT line to match the exact calculated 7% VAT
 */
export function setExactVatLine(
  lines: GLLine[],
  vatLineIndex: number,
  exactVatUnits: bigint,
  scale = 2
): GLLine[] {
  if (vatLineIndex < 0 || vatLineIndex >= lines.length) return lines;
  const nextLines = lines.map((line) => ({ ...line }));
  const line = nextLines[vatLineIndex];
  const exactStr = amountString(exactVatUnits, scale);

  if (line.debit && line.debit !== "0") {
    line.debit = exactStr;
  } else {
    line.credit = exactStr;
  }

  return nextLines;
}

/**
 * Appends a new 7% VAT line with automatic account lookup
 */
export function appendVatLine(
  lines: GLLine[],
  accounts: GLAccount[],
  vatType: TaxType,
  vatUnits: bigint,
  scale = 2
): GLLine[] {
  const vatStr = amountString(vatUnits, scale);
  const isInput = vatType === "input_tax";

  // Find the VAT account by name among active posting accounts; no code guessing. Not found → leave the account
  // empty so the user picks it (the save validates it) rather than writing to a code that may not exist.
  // Charts often keep "ภาษีซื้อ/ภาษีขายยังไม่ถึงกำหนด" next to the normal VAT account; a plain 7% VAT line
  // belongs to the normal one, so the undue account is only the fallback when it is the sole match.
  const candidates = accounts.filter((a) =>
    a.allowposting !== false &&
    a.isactive !== false &&
    isVatAccount(a.names?.[0]?.name ?? "", "", a.accounttype).type === vatType
  );
  const matchingAcc = candidates.find((a) => !UNDUE_VAT_NAME.test(a.names?.[0]?.name ?? "")) ?? candidates[0];

  const accountCode = matchingAcc?.accountcode ?? "";
  const description = isInput ? "ภาษีซื้อ 7%" : "ภาษีขาย 7%";

  const newLine: GLLine = {
    accountcode: accountCode,
    description,
    debit: isInput ? vatStr : "0",
    credit: isInput ? "0" : vatStr,
    departmentcode: "",
    projectcode: "",
    cashflow: "",
  };

  return [...lines, newLine];
}
