import { displayAmountUnits, formatAmount, type GLAccount } from "./general-ledger";

export interface ComparativeRow {
  accountcode: string;
  accountname: string;
  accounttype?: string;
  period1Amount: string;
  period2Amount: string;
  varianceAmount: string;
  percentChange: string;
  direction: "increase" | "decrease" | "unchanged";
  isNew?: boolean;
}

export interface ComparativeReportResult {
  period1Label: string;
  period2Label: string;
  columns: { key: string; label: string; amount?: boolean }[];
  rows: ComparativeRow[];
  totals: {
    period1Total: string;
    period2Total: string;
    varianceTotal: string;
    percentChangeTotal: string;
    direction: "increase" | "decrease" | "unchanged";
  };
}

/**
 * Extracts numeric value from a row using prioritized amount keys
 */
export function extractRowAmountUnits(row: Record<string, string>, preferKey?: string): bigint {
  if (preferKey && row[preferKey] !== undefined) {
    try {
      return displayAmountUnits(row[preferKey] || "0");
    } catch {
      // fallback
    }
  }

  // Priority: amount -> balance -> endingdebit - endingcredit -> debit - credit
  if (row.amount !== undefined) {
    try {
      return displayAmountUnits(row.amount || "0");
    } catch {
      // fallback
    }
  }

  if (row.balance !== undefined) {
    try {
      return displayAmountUnits(row.balance || "0");
    } catch {
      // fallback
    }
  }

  if (row.endingdebit !== undefined || row.endingcredit !== undefined) {
    try {
      const d = displayAmountUnits(row.endingdebit || "0");
      const c = displayAmountUnits(row.endingcredit || "0");
      return d - c;
    } catch {
      // fallback
    }
  }

  if (row.debit !== undefined || row.credit !== undefined) {
    try {
      const d = displayAmountUnits(row.debit || "0");
      const c = displayAmountUnits(row.credit || "0");
      return d - c;
    } catch {
      // fallback
    }
  }

  return 0n;
}

/**
 * Calculates percentage change between two scaled BigInt values
 */
export function calculatePercentChange(p1Units: bigint, p2Units: bigint): {
  percentStr: string;
  direction: "increase" | "decrease" | "unchanged";
  isNew: boolean;
} {
  if (p2Units === 0n) {
    if (p1Units === 0n) {
      return { percentStr: "0.00%", direction: "unchanged", isNew: false };
    }
    if (p1Units > 0n) {
      return { percentStr: "+100.00%", direction: "increase", isNew: true };
    }
    return { percentStr: "-100.00%", direction: "decrease", isNew: true };
  }

  const p1Num = Number(p1Units) / 100000000;
  const p2Num = Number(p2Units) / 100000000;
  const absP2 = Math.abs(p2Num);

  const diff = p1Num - p2Num;
  if (Math.abs(diff) < 0.0001) {
    return { percentStr: "0.00%", direction: "unchanged", isNew: false };
  }

  const pct = (diff / absP2) * 100;
  const sign = pct > 0 ? "+" : "";
  const direction = pct > 0 ? "increase" : "decrease";

  return {
    percentStr: `${sign}${pct.toFixed(2)}%`,
    direction,
    isNew: false,
  };
}

/**
 * Build multi-period comparative financial statement
 */
export function buildComparativeReport({
  period1Label,
  period1Rows = [],
  period2Label,
  period2Rows = [],
  accounts = [],
  preferAmountKey,
}: {
  period1Label: string;
  period1Rows: Record<string, string>[];
  period2Label: string;
  period2Rows: Record<string, string>[];
  accounts?: GLAccount[];
  preferAmountKey?: string;
}): ComparativeReportResult {
  const accountsMap = new Map<string, GLAccount>();
  for (const acc of accounts) {
    accountsMap.set(acc.accountcode, acc);
  }

  const p1Map = new Map<string, Record<string, string>>();
  for (const row of period1Rows) {
    const code = row.accountcode;
    if (code) p1Map.set(code, row);
  }

  const p2Map = new Map<string, Record<string, string>>();
  for (const row of period2Rows) {
    const code = row.accountcode;
    if (code) p2Map.set(code, row);
  }

  // Union of all account codes
  const allCodesSet = new Set<string>([...p1Map.keys(), ...p2Map.keys()]);
  const allCodes = Array.from(allCodesSet).filter((c) => c !== "__current_earnings__");
  allCodes.sort((a, b) => a.localeCompare(b, undefined, { numeric: true }));

  const rows: ComparativeRow[] = [];
  let totalP1Units = 0n;
  let totalP2Units = 0n;

  for (const code of allCodes) {
    const r1 = p1Map.get(code);
    const r2 = p2Map.get(code);
    const accMaster = accountsMap.get(code);

    const name =
      r1?.accountname ||
      r2?.accountname ||
      accMaster?.names?.find((n) => n.code === "th")?.name ||
      code;

    const type = r1?.accounttype || r2?.accounttype || accMaster?.accounttype;

    const u1 = r1 ? extractRowAmountUnits(r1, preferAmountKey) : 0n;
    const u2 = r2 ? extractRowAmountUnits(r2, preferAmountKey) : 0n;

    totalP1Units += u1;
    totalP2Units += u2;

    const varianceUnits = u1 - u2;
    const { percentStr, direction, isNew } = calculatePercentChange(u1, u2);

    const p1Str = formatAmount((Number(u1) / 100000000).toFixed(2));
    const p2Str = formatAmount((Number(u2) / 100000000).toFixed(2));
    const varStr = formatAmount((Number(varianceUnits) / 100000000).toFixed(2));

    rows.push({
      accountcode: code,
      accountname: name,
      accounttype: type,
      period1Amount: p1Str,
      period2Amount: p2Str,
      varianceAmount: varStr,
      percentChange: percentStr,
      direction,
      isNew,
    });
  }

  const totalVarianceUnits = totalP1Units - totalP2Units;
  const { percentStr: totalPct, direction: totalDir } = calculatePercentChange(
    totalP1Units,
    totalP2Units
  );

  const columns = [
    { key: "accountcode", label: "รหัสบัญชี", amount: false },
    { key: "accountname", label: "ชื่อบัญชี", amount: false },
    { key: "period1Amount", label: period1Label, amount: true },
    { key: "period2Amount", label: period2Label, amount: true },
    { key: "varianceAmount", label: "ผลต่าง (Variance)", amount: true },
    { key: "percentChange", label: "% เปลี่ยนแปลง", amount: true },
  ];

  return {
    period1Label,
    period2Label,
    columns,
    rows,
    totals: {
      period1Total: formatAmount((Number(totalP1Units) / 100000000).toFixed(2)),
      period2Total: formatAmount((Number(totalP2Units) / 100000000).toFixed(2)),
      varianceTotal: formatAmount((Number(totalVarianceUnits) / 100000000).toFixed(2)),
      percentChangeTotal: totalPct,
      direction: totalDir,
    },
  };
}

export interface MonthlyPivotResult {
  columns: { key: string; label: string; amount?: boolean }[];
  rows: Record<string, string>[];
  monthTotals: Record<string, string>;
  grandTotal: string;
}

const MONTH_NAMES_TH = [
  "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
  "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."
];

/**
 * Pivot 12-month annual-balances report into a wide monthly trend matrix
 */
export function pivotAnnualBalances(rawRows: Record<string, string>[]): MonthlyPivotResult {
  const accountRowsMap = new Map<
    string,
    {
      code: string;
      name: string;
      type: string;
      months: Map<number, bigint>; // month index 1-12 -> units
    }
  >();

  for (const row of rawRows) {
    const code = row.accountcode;
    if (!code || code === "__current_earnings__") continue;

    // Month in raw rows is typically "YYYY-MM" or "MM" or "2026-01"
    const monthStr = row.month || "";
    let monthNum = 0;
    if (monthStr.includes("-")) {
      monthNum = parseInt(monthStr.split("-")[1], 10);
    } else {
      monthNum = parseInt(monthStr, 10);
    }

    if (isNaN(monthNum) || monthNum < 1 || monthNum > 12) {
      monthNum = 1;
    }

    let accData = accountRowsMap.get(code);
    if (!accData) {
      accData = {
        code,
        name: row.accountname || code,
        type: row.accounttype || "asset",
        months: new Map<number, bigint>(),
      };
      accountRowsMap.set(code, accData);
    }

    const units = extractRowAmountUnits(row);
    const prev = accData.months.get(monthNum) || 0n;
    accData.months.set(monthNum, prev + units);
  }

  const columns: { key: string; label: string; amount?: boolean }[] = [
    { key: "accountcode", label: "รหัสบัญชี", amount: false },
    { key: "accountname", label: "ชื่อบัญชี", amount: false },
  ];

  for (let m = 1; m <= 12; m++) {
    columns.push({
      key: `m${m}`,
      label: MONTH_NAMES_TH[m - 1],
      amount: true,
    });
  }
  columns.push({ key: "total", label: "รวมทั้งปี", amount: true });

  const monthlySums = new Map<number, bigint>();
  let grandTotalUnits = 0n;

  const pivotedRows: Record<string, string>[] = [];
  const sortedAccounts = Array.from(accountRowsMap.values()).sort((a, b) =>
    a.code.localeCompare(b.code, undefined, { numeric: true })
  );

  for (const acc of sortedAccounts) {
    const r: Record<string, string> = {
      accountcode: acc.code,
      accountname: acc.name,
      accounttype: acc.type,
    };

    let rowTotalUnits = 0n;
    for (let m = 1; m <= 12; m++) {
      const val = acc.months.get(m) || 0n;
      rowTotalUnits += val;
      r[`m${m}`] = formatAmount((Number(val) / 100000000).toFixed(2));

      const prevMonthSum = monthlySums.get(m) || 0n;
      monthlySums.set(m, prevMonthSum + val);
    }

    grandTotalUnits += rowTotalUnits;
    r.total = formatAmount((Number(rowTotalUnits) / 100000000).toFixed(2));
    pivotedRows.push(r);
  }

  const monthTotals: Record<string, string> = {};
  for (let m = 1; m <= 12; m++) {
    const s = monthlySums.get(m) || 0n;
    monthTotals[`m${m}`] = formatAmount((Number(s) / 100000000).toFixed(2));
  }

  return {
    columns,
    rows: pivotedRows,
    monthTotals,
    grandTotal: formatAmount((Number(grandTotalUnits) / 100000000).toFixed(2)),
  };
}
