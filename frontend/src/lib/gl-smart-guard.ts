import {
  amountUnits,
  amountString,
  type GLLine,
} from "./general-ledger";

export interface GLBalanceAnalysis {
  totalDebitUnits: bigint;
  totalCreditUnits: bigint;
  differenceUnits: bigint; // debit - credit
  isBalanced: boolean;
  balanceStatus: "balanced" | "debit_surplus" | "credit_surplus";
  balanceDifferenceFormatted: string;
}

/**
 * Analyzes journal lines for debit-credit balance only. Tax checks (VAT rate, base, rounding) are done by the
 * backend from the user-typed rate — no default rate lives in the browser (mydocs/specs/gl/tax-design-notes.md).
 */
export function analyzeGLBalance(lines: GLLine[], scale = 2): GLBalanceAnalysis {
  let totalDebitUnits = 0n;
  let totalCreditUnits = 0n;

  for (const line of lines) {
    try {
      if (line.debit && line.debit !== "0") totalDebitUnits += amountUnits(line.debit);
    } catch {
      // invalid amount text counts as 0 while the user is still typing
    }
    try {
      if (line.credit && line.credit !== "0") totalCreditUnits += amountUnits(line.credit);
    } catch {
      // invalid amount text counts as 0 while the user is still typing
    }
  }

  const differenceUnits = totalDebitUnits - totalCreditUnits;
  const isBalanced = differenceUnits === 0n;
  const balanceStatus: GLBalanceAnalysis["balanceStatus"] =
    differenceUnits === 0n
      ? "balanced"
      : differenceUnits > 0n
      ? "debit_surplus"
      : "credit_surplus";

  let balanceDifferenceFormatted: string;
  try {
    balanceDifferenceFormatted = amountString(differenceUnits < 0n ? -differenceUnits : differenceUnits, scale);
  } catch {
    balanceDifferenceFormatted = "0.00";
  }

  return {
    totalDebitUnits,
    totalCreditUnits,
    differenceUnits,
    isBalanced,
    balanceStatus,
    balanceDifferenceFormatted,
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
