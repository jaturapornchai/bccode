import { MENU_SECTIONS } from "./menu-data";
import type { GLJournalDetails } from "./gl-journal-details";

export type GLIdentity = { id?: string; version?: number; isdeleted?: boolean };
export type GLAccount = GLIdentity & {
  accountcode: string; names: { code: string; name: string }[];
  accounttype: "asset" | "liability" | "equity" | "income" | "expense";
  parentaccountcode: string | null; normalbalance: "debit" | "credit";
  allowposting: boolean; isactive: boolean; accountgroup: string; iscash: boolean;
  level?: number;
};
export type GLFiscalYear = GLIdentity & {
  code: string; startdate: string; enddate: string; scale: number;
  retainedearningsaccount: string; profitlossaccount: string; isactive: boolean; closed: boolean;
};
export type GLRule = { accountcode: string; side: string; source: string };
export type GLAllocationRule = { branchcode: string; departmentcode: string; projectcode: string; accountcode: string; rate: string };
export type GLMaster = GLIdentity & {
  code: string; name: string; isactive: boolean; accountcode: string; fiscalyear: string;
  startdate: string; enddate: string; locked: boolean; amount: string;
  branchcode: string; departmentcode: string; projectcode: string; direction: string;
  bookcode: string; rules: GLRule[]; itemaccount: string; costaccount: string; revenueaccount: string;
  allocatemode: string; allocaterules: GLAllocationRule[];
};
export type GLLine = {
  accountcode: string; accountname?: string; description: string; debit: string; credit: string;
  departmentcode: string; projectcode: string; cashflow: string;
};
export type GLJournal = GLIdentity & {
  details?: GLJournalDetails;
  source_type?: number; source_system?: string; source_record_id?: string;
  docno: string; date: string; bookcode: string; fiscalyear: string; description: string;
  reference: string; branchcode: string; kind: string; status: string;
  lines: GLLine[]; reversalof?: string; reason?: string;
};
export type GLRecord = GLAccount | GLFiscalYear | GLMaster | GLJournal | GLStatementTemplate;
export type GLReviewStatus = 1 | 2 | 3;
export type GLReviewEvent = { eventno: number; version: number; status: GLReviewStatus; note: string; reviewedby: string; reviewedat: string };
export type GLJournalReview = { journalid: string; version: number; status: GLReviewStatus; eventno: number; events: GLReviewEvent[] };
export type GLPage<T> = { items: T[]; total: number; page: number; limit: number; sequence: number };
export type GLReport = {
  columns: { key: string; label: string; amount?: boolean }[];
  rows: Record<string, string>[]; totals: Record<string, string>;
  totalrows: number; warnings: string[]; asof: string; sequence: number;
};
export const GL_RESOURCES = ["accounts", "fiscal-years", "account-groups", "product-account-groups", "mappings", "budgets", "periods", "forecast", "allocations", "journals", "statement-templates", "journal-books"] as const;
export type GLResource = typeof GL_RESOURCES[number];
export type GLCommand = {
  resource: GLResource | "processes"; id?: string; action: string; requestid: string;
  version?: number; reason?: string; date?: string; docno?: string; targetyear?: string;
  account?: GLAccount; fiscalyear?: GLFiscalYear; master?: GLMaster; journal?: GLJournal | Pick<GLJournal, "details">; statementtemplate?: GLStatementTemplate;
  review?: { status: GLReviewStatus; note: string; expectedEventNo: number };
};
export const GL_REPORTS = ["ledger", "trialbalance", "pnl", "balancesheet", "workingpaper", "gljournal", "budgetcomparison", "ar-outstanding", "ap-outstanding", "bank-unmatched"] as const;
// เฉพาะกลุ่ม gl-* ที่จอ GL เปิดเอง; กลุ่มภาษี (vat-*) อยู่ในหมวดเดียวกันแต่ใช้จอภาษี
export const GL_MENU_ITEMS = MENU_SECTIONS.find((section) => section.id === "gl")!.groups.filter((group) => group.id.startsWith("gl-")).flatMap((group) => group.items);
export function isGeneralLedgerRoute(route: string) { const clean = route.split("?")[0]; return GL_MENU_ITEMS.some((item) => item.route === clean) || clean.startsWith("/gl/journal/") || clean === "/gl/unposting"; }
/** Screen text follows the selected language (AGENTS.md 2026-09-14): [languages.tsv key, Thai fallback]. */
export type GLLabel = readonly [key: string, thai: string];
export type GLTextFn = (key: string, fallback: string) => string;
export function labelText(labels: Record<string, GLLabel>, value: string, tr: GLTextFn, fallback = value) {
  const label = labels[value]; return label ? tr(label[0], label[1]) : fallback;
}
export function thaiLabels<K extends string>(labels: Record<K, GLLabel>): Record<K, string> {
  return Object.fromEntries(Object.entries<GLLabel>(labels).map(([code, label]) => [code, label[1]])) as Record<K, string>;
}
export const accountTypeLabels: Record<GLAccount["accounttype"], GLLabel> = { asset: ["gl_asset", "สินทรัพย์"], liability: ["gl_liability", "หนี้สิน"], equity: ["gl_equity", "ส่วนของเจ้าของ"], income: ["gl_revenue", "รายได้"], expense: ["gl_expense", "ค่าใช้จ่าย"] };
export const bookLabels: Record<string, GLLabel> = { JV: ["gl_general_journal", "รายวันทั่วไป"], UV: ["gl_sales_journal", "รายวันขาย"], SV: ["gl_purchase_journal", "รายวันซื้อ"], RV: ["gl_cash_receipts_journal", "รายวันรับเงิน"], PV: ["gl_cash_payments_journal", "รายวันจ่ายเงิน"] };
export const accountTypes = thaiLabels(accountTypeLabels);
export const books = thaiLabels(bookLabels);
export function accountName(account: GLAccount) { return account.names?.find((name) => name.code === "th")?.name ?? account.accountcode; }
export function emptyAccount(): GLAccount { return { accountcode: "", names: [{ code: "th", name: "" }], accounttype: "asset", parentaccountcode: null, normalbalance: "debit", allowposting: true, isactive: true, accountgroup: "", iscash: false, level: 1 }; }

export const ACCOUNT_TYPE_ORDER: Record<string, number> = {
  asset: 1,
  liability: 2,
  equity: 3,
  income: 4,
  expense: 5,
};

export function getEffectiveAccountLevel(acc: GLAccount, accountsMap: Map<string, GLAccount>, visited = new Set<string>()): number {
  if (visited.has(acc.accountcode)) return acc.level && acc.level >= 1 ? acc.level : 1;
  visited.add(acc.accountcode);

  const rawLevel = typeof acc.level === "number" && acc.level >= 1 ? acc.level : 1;
  if (acc.parentaccountcode?.trim()) {
    const parent = accountsMap.get(acc.parentaccountcode.trim());
    if (parent && parent.accountcode !== acc.accountcode) {
      const parentLevel = getEffectiveAccountLevel(parent, accountsMap, visited);
      return Math.max(rawLevel, parentLevel + 1);
    }
  }
  return rawLevel;
}

export function sortAccountsHierarchically(accounts: GLAccount[]): (GLAccount & { effectiveLevel: number })[] {
  const accountsMap = new Map<string, GLAccount>();
  for (const acc of accounts) {
    accountsMap.set(acc.accountcode, acc);
  }

  const effectiveLevels = new Map<string, number>();
  for (const acc of accounts) {
    effectiveLevels.set(acc.accountcode, getEffectiveAccountLevel(acc, accountsMap));
  }

  const childrenMap = new Map<string, GLAccount[]>();
  const rootsByCategory: Record<string, GLAccount[]> = {
    asset: [],
    liability: [],
    equity: [],
    income: [],
    expense: [],
  };

  for (const acc of accounts) {
    const parentCode = acc.parentaccountcode?.trim();
    if (parentCode && accountsMap.has(parentCode) && parentCode !== acc.accountcode) {
      const list = childrenMap.get(parentCode) || [];
      list.push(acc);
      childrenMap.set(parentCode, list);
    } else {
      const cat = acc.accounttype in rootsByCategory ? acc.accounttype : "asset";
      rootsByCategory[cat].push(acc);
    }
  }

  const codeCompare = (a: GLAccount, b: GLAccount) =>
    a.accountcode.localeCompare(b.accountcode, undefined, { numeric: true, sensitivity: "base" });

  for (const [, children] of childrenMap) {
    children.sort(codeCompare);
  }

  for (const cat of Object.keys(rootsByCategory)) {
    rootsByCategory[cat].sort(codeCompare);
  }

  const categories: ("asset" | "liability" | "equity" | "income" | "expense")[] = [
    "asset",
    "liability",
    "equity",
    "income",
    "expense",
  ];

  const result: (GLAccount & { effectiveLevel: number })[] = [];
  const added = new Set<string>();

  function addSubtree(node: GLAccount) {
    if (added.has(node.accountcode)) return;
    added.add(node.accountcode);
    result.push({
      ...node,
      effectiveLevel: effectiveLevels.get(node.accountcode) ?? (node.level ?? 1),
    });
    const children = childrenMap.get(node.accountcode) || [];
    for (const child of children) {
      addSubtree(child);
    }
  }

  for (const cat of categories) {
    for (const root of rootsByCategory[cat]) {
      addSubtree(root);
    }
  }

  for (const acc of accounts) {
    if (!added.has(acc.accountcode)) {
      result.push({
        ...acc,
        effectiveLevel: effectiveLevels.get(acc.accountcode) ?? (acc.level ?? 1),
      });
    }
  }

  return result;
}

export function emptyFiscalYear(): GLFiscalYear { return { code: "", startdate: "", enddate: "", scale: 2, retainedearningsaccount: "", profitlossaccount: "", isactive: true, closed: false }; }
export function emptyAllocationRule(): GLAllocationRule { return { branchcode: "", departmentcode: "", projectcode: "", accountcode: "", rate: "0" }; }
export function emptyMaster(): GLMaster { return { code: "", name: "", isactive: true, accountcode: "", fiscalyear: "", startdate: "", enddate: "", locked: false, amount: "0", branchcode: "", departmentcode: "", projectcode: "", direction: "in", bookcode: "JV", rules: [], itemaccount: "", costaccount: "", revenueaccount: "", allocatemode: "percent", allocaterules: [] }; }
export function emptyLine(): GLLine { return { accountcode: "", description: "", debit: "0", credit: "0", departmentcode: "", projectcode: "", cashflow: "" }; }
export function localDate() { const date = new Date(); return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`; }
export function emptyJournal(bookcode = "JV", kind = "manual"): GLJournal { return { docno: "", date: localDate(), bookcode, fiscalyear: "", description: "", reference: "", branchcode: "", kind, status: "draft", lines: [emptyLine(), emptyLine()] }; }

const amountPattern = /^-?(0|[1-9]\d{0,25})(\.\d{1,8})?$/;
const factor = 100000000n;
/** Accounting values remain strings at the boundary; BigInt is only for exact validation. */
export function amountUnits(value: string): bigint {
  if (!amountPattern.test(value)) throw new Error("จำนวนเงินต้องเป็นเลขทศนิยมไม่เกิน 8 ตำแหน่ง ไม่ใส่จุลภาค");
  return decimalUnits(value);
}
/** Report aggregates may exceed the precision of a single input; never truncate them for display. */
export function displayAmountUnits(value: string): bigint {
  if (!/^-?(0|[1-9]\d{0,99})(\.\d{1,8})?$/.test(value)) throw new Error("จำนวนเงินไม่ถูกต้อง");
  return decimalUnits(value);
}
function decimalUnits(value: string): bigint {
  const negative = value.startsWith("-");
  const [whole, fraction = ""] = value.replace(/^-/, "").split(".");
  const units = BigInt(whole) * factor + BigInt(fraction.padEnd(8, "0"));
  return negative ? -units : units;
}
export function amountString(units: bigint, scale = 8): string {
  if (!Number.isInteger(scale) || scale < 0 || scale > 8) throw new Error("จำนวนตำแหน่งทศนิยมไม่ถูกต้อง");
  const abs = units < 0n ? -units : units;
  const tail = (abs % factor).toString().padStart(8, "0");
  // Display must never round away persisted precision.
  const effectiveScale = Math.max(scale, tail.replace(/0+$/, "").length);
  return `${units < 0n ? "-" : ""}${abs / factor}${effectiveScale ? `.${tail.slice(0, effectiveScale)}` : ""}`;
}
/** Formats a numeric string with thousands commas, respecting scale and preserving persisted precision */
export function formatAmount(value: string, scale = 2): string {
  if (!value || !value.trim()) return "";
  const clean = value.replace(/,/g, "").trim();
  const isNegative = clean.startsWith("-");
  const unsigned = isNegative ? clean.slice(1) : clean;
  if (!unsigned || unsigned === ".") return "";
  const parts = unsigned.split(".");
  if (parts.length > 2) return "จำนวนเงินไม่ถูกต้อง";
  const whole = parts[0] || "0";
  const fraction = parts[1] ?? "";
  if (!/^\d+$/.test(whole) || (fraction && !/^\d+$/.test(fraction))) return "จำนวนเงินไม่ถูกต้อง";

  const isZero = (whole === "0" || whole === "") && fraction.padEnd(scale, "0").slice(0, Math.max(scale, fraction.length)).replace(/0/g, "") === "";
  const sign = isNegative && !isZero ? "-" : "";
  const wholeFormatted = whole.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  const effectiveScale = Math.max(scale, fraction.replace(/0+$/, "").length);
  if (effectiveScale <= 0) return `${sign}${wholeFormatted}`;
  const fracFormatted = fraction.padEnd(effectiveScale, "0").slice(0, effectiveScale);
  return `${sign}${wholeFormatted}.${fracFormatted}`;
}

export function journalTotals(lines: GLLine[]) {
  const debit = lines.reduce((sum, line) => sum + amountUnits(line.debit), 0n);
  const credit = lines.reduce((sum, line) => sum + amountUnits(line.credit), 0n);
  return { debit, credit, difference: debit - credit };
}
export function validateJournal(journal: GLJournal, year: GLFiscalYear | undefined, accounts: GLAccount[], tr: GLTextFn = (_key, fallback) => fallback): string | null {
  if (!journal.docno.trim() || !journal.date || !journal.description.trim()) return tr("gl_required_no_date_description", "กรุณาระบุเลขที่ วันที่ และคำอธิบายรายการ");
  if (!year || !year.isactive || year.closed || journal.date < year.startdate || journal.date > year.enddate) return tr("gl_select_active_fiscal_year_date", "กรุณาเลือกปีบัญชีที่เปิดใช้งานและวันที่ภายในปีบัญชี");
  if (journal.lines.length < 2 || journal.lines.length > 500) return tr("gl_entries_2_500_lines", "รายการบัญชีต้องมี 2–500 บรรทัด");
  try {
    for (const [index, line] of journal.lines.entries()) {
      const account = accounts.find((item) => item.accountcode === line.accountcode);
      if (!account?.isactive || !account.allowposting || account.isdeleted) return tr("gl_line_select_active_account", "บรรทัด {0}: เลือกบัญชีที่เปิดใช้งานและลงรายการได้").replace("{0}", String(index + 1));
      const debit = amountUnits(line.debit), credit = amountUnits(line.credit);
      if (debit < 0n || credit < 0n || (debit === 0n) === (credit === 0n)) return tr("gl_line_enter_amount_one_side_only", "บรรทัด {0}: ใส่ยอดมากกว่าศูนย์เพียงด้านเดียว").replace("{0}", String(index + 1));
      const precision = 10n ** BigInt(8 - year.scale);
      if (debit % precision !== 0n || credit % precision !== 0n) return tr("gl_line_amount_exceeds_decimal_places", "บรรทัด {0}: จำนวนเงินเกิน {1} ตำแหน่งทศนิยม").replace("{0}", String(index + 1)).replace("{1}", String(year.scale));
    }
    if (journalTotals(journal.lines).difference !== 0n) return tr("gl_dr_cr_equal_before_save", "ยอดเดบิตและเครดิตต้องเท่ากันก่อนบันทึก");
  } catch (error) { return error instanceof Error ? error.message : tr("gl_please_check_amount", "กรุณาตรวจสอบจำนวนเงิน"); }
  return null;
}

/** Spreadsheet apps must treat untrusted labels as text, never formulas. Amounts are raw numbers. */
export function csvCell(value: string, isAmount = false) {
  const needsApostrophe = !isAmount && /^[\s]*[=+\-@\t\r]/.test(value);
  return `"${(needsApostrophe ? "'" : "") + value.replace(/"/g, '""')}"`;
}
export function reportCsv(report: GLReport) {
  return "\uFEFF" + [report.columns.map((column) => csvCell(column.label)).join(","), ...report.rows.map((row) => report.columns.map((column) => csvCell(row[column.key] ?? "", Boolean(column.amount))).join(","))].join("\r\n");
}

export type StatementType = "balance_sheet" | "pnl" | "production_cost" | "cash_flow" | "custom";
export type StatementRowType = "header" | "account" | "formula" | "subtotal" | "blank" | "divider";
export type StatementRowUnderline = "none" | "single" | "double" | "top_single" | "top_single_bottom_double";

export type StatementRowStyle = {
  fontfamily?: string;
  fontsize?: string;
  fontweight?: "normal" | "bold" | "semibold";
  fontstyle?: "normal" | "italic";
  align?: "left" | "center" | "right";
  color?: string;
  indent?: number;
  underline?: StatementRowUnderline;
};

export type StatementRow = {
  id: string;
  rowno: number;
  rowtype: StatementRowType;
  title: string;
  noteno?: string;
  accountcodes?: string[];
  accountgroup?: string;
  normalbalance?: "debit" | "credit" | "net";
  formula?: string;
  reversesign?: boolean;
  showzero?: boolean;
  style?: StatementRowStyle;
};

export type StatementGlobalStyle = {
  fontfamily?: string;
  fontsize?: string;
  scale?: number;
  compact?: boolean;
  shownotecolumn?: boolean;
  comparisontype?: "none" | "previous_year" | "budget";
};

export type GLStatementTemplate = GLIdentity & {
  code: string;
  name: string;
  statementtype: StatementType;
  isactive: boolean;
  globalstyle: StatementGlobalStyle;
  rows: StatementRow[];
};

export const statementTypeLabels: Record<StatementType, GLLabel> = {
  balance_sheet: ["gl_stmt_fin_position_bs", "งบแสดงฐานะการเงิน (งบดุล)"],
  pnl: ["gl_income_statement", "งบกำไรขาดทุน"],
  production_cost: ["gl_cost_prod_cogs_stmt", "งบต้นทุนผลิตและต้นทุนขาย"],
  cash_flow: ["gl_cash_flow_stmt", "งบกระแสเงินสด"],
  custom: ["gl_custom_fin_stmt", "งบการเงินกำหนดเอง"],
};

export const statementRowTypeLabels: Record<StatementRowType, GLLabel> = {
  header: ["gl_heading", "หัวข้อ"],
  account: ["gl_account_balance", "ยอดบัญชี"],
  formula: ["gl_calculation_formula", "สูตรคำนวณ"],
  subtotal: ["gl_total_amount", "รวมยอด"],
  blank: ["gl_blank_line_2", "บรรทัดว่าง"],
  divider: ["gl_separator_line", "เส้นคั่น"],
};

export function emptyStatementTemplate(): GLStatementTemplate {
  return {
    code: "",
    name: "",
    statementtype: "custom",
    isactive: true,
    globalstyle: {
      fontfamily: "sarabun",
      fontsize: "15px",
      scale: 2,
      compact: false,
      shownotecolumn: true,
      comparisontype: "none",
    },
    rows: [],
  };
}

export function evaluateStatementFormula(
  formula: string,
  rowValues: Map<number, bigint>,
  currentRowNo?: number,
  visited: Set<number> = new Set(),
): bigint {
  const clean = formula.trim();
  if (!clean) return 0n;

  // Range sum: SUM(R10:R50) or SUM(10:50) or SUM(R10..R50)
  const rangeMatch = clean.match(/^SUM\s*\(\s*R?(\d+)\s*(?::|\.\.)\s*R?(\d+)\s*\)$/i);
  if (rangeMatch) {
    const from = parseInt(rangeMatch[1], 10);
    const to = parseInt(rangeMatch[2], 10);
    const start = Math.min(from, to);
    const end = Math.max(from, to);
    let sum = 0n;
    for (const [rno, val] of rowValues.entries()) {
      if (rno >= start && rno <= end && rno !== currentRowNo) {
        sum += val;
      }
    }
    return sum;
  }

  // Tokenize arithmetic expression
  try {
    const tokens = tokenizeFormula(clean);
    return parseFormulaExpr(tokens, rowValues, currentRowNo, visited);
  } catch {
    return 0n;
  }
}

function tokenizeFormula(formula: string): string[] {
  const matches = formula.match(/[A-Za-z]+[0-9]*|[0-9]+(?:\.[0-9]+)?|\+|-|\*|\/|\(|\)/g);
  return matches ?? [];
}

function parseFormulaExpr(
  tokens: string[],
  rowValues: Map<number, bigint>,
  currentRowNo?: number,
  visited: Set<number> = new Set(),
): bigint {
  let pos = 0;

  function peek(): string | undefined {
    return tokens[pos];
  }

  function consume(): string {
    return tokens[pos++];
  }

  function expr(): bigint {
    let result = term();
    while (peek() === "+" || peek() === "-") {
      const op = consume();
      const right = term();
      result = op === "+" ? result + right : result - right;
    }
    return result;
  }

  function term(): bigint {
    let result = parsePrimary();
    while (peek() === "*" || peek() === "/") {
      const op = consume();
      const right = parsePrimary();
      if (op === "*") {
        result = (result * right) / factor;
      } else {
        result = right === 0n ? 0n : (result * factor) / right;
      }
    }
    return result;
  }

  function parsePrimary(): bigint {
    const token = peek();
    if (!token) return 0n;

    if (token === "+") {
      consume();
      return parsePrimary();
    }
    if (token === "-") {
      consume();
      return -parsePrimary();
    }
    if (token === "(") {
      consume();
      const val = expr();
      if (peek() === ")") consume();
      return val;
    }

    consume();

    // Row reference like R10, r10
    if (/^[Rr]\d+$/.test(token)) {
      const rowNo = parseInt(token.slice(1), 10);
      if (rowNo === currentRowNo || visited.has(rowNo)) return 0n;
      return rowValues.get(rowNo) ?? 0n;
    }

    // Pure number
    if (/^\d+(\.\d+)?$/.test(token)) {
      try {
        return decimalUnits(token);
      } catch {
        return 0n;
      }
    }

    return 0n;
  }

  return expr();
}

export type CalculatedStatementRow = StatementRow & {
  amount: bigint;
  amountFormatted: string;
};

export type CalculatedStatement = {
  template: GLStatementTemplate;
  rows: CalculatedStatementRow[];
  totals: Record<string, string>;
  isBalanced?: boolean;
  difference?: string;
};

export function calculateStatement(
  template: GLStatementTemplate,
  trialBalanceRows: Record<string, string>[],
): CalculatedStatement {
  const rowValues = new Map<number, bigint>();
  const accountBalances = new Map<string, { balance: bigint; debit: bigint; credit: bigint; type: string }>();

  for (const tr of trialBalanceRows) {
    const code = tr.accountcode ?? "";
    if (!code) continue;
    try {
      const deb = displayAmountUnits(tr.endingdebit ?? tr.debit ?? "0");
      const cre = displayAmountUnits(tr.endingcredit ?? tr.credit ?? "0");
      const bal = displayAmountUnits(tr.balance ?? "0");
      accountBalances.set(code, { balance: bal, debit: deb, credit: cre, type: tr.accounttype ?? "asset" });
    } catch {
      // Ignore unparseable values
    }
  }

  // Pass 1: Account rows
  for (const row of template.rows) {
    if (row.rowtype === "account") {
      let sum = 0n;
      const codes = row.accountcodes ?? [];
      for (const code of codes) {
        const acc = accountBalances.get(code);
        if (!acc) continue;
        const norm = row.normalbalance ?? (["asset", "expense"].includes(acc.type) ? "debit" : "credit");
        let val = 0n;
        if (norm === "debit") val = acc.balance;
        else if (norm === "credit") val = -acc.balance;
        else val = acc.balance;
        sum += val;
      }
      if (row.reversesign) sum = -sum;
      rowValues.set(row.rowno, sum);
    }
  }

  // Pass 2: Formulas & Subtotals
  for (const row of template.rows) {
    if (row.rowtype === "formula" || row.rowtype === "subtotal") {
      let sum = 0n;
      if (row.formula?.trim()) {
        sum = evaluateStatementFormula(row.formula, rowValues, row.rowno);
      }
      if (row.reversesign) sum = -sum;
      rowValues.set(row.rowno, sum);
    }
  }

  const scale = template.globalstyle?.scale ?? 2;
  const calculatedRows: CalculatedStatementRow[] = template.rows.map((row) => {
    const isAmountRow = ["account", "formula", "subtotal"].includes(row.rowtype);
    const amount = isAmountRow ? (rowValues.get(row.rowno) ?? 0n) : 0n;
    const amountFormatted = isAmountRow ? (amount === 0n && !row.showzero ? "-" : formatAmount(amountString(amount, scale), scale)) : "";
    return {
      ...row,
      amount,
      amountFormatted,
    };
  });

  return {
    template,
    rows: calculatedRows,
    totals: {},
  };
}

export function generateStarterTemplates(): GLStatementTemplate[] {
  return [
    {
      code: "BS-DBD",
      name: "งบแสดงฐานะการเงิน (แบบ DBD กรมพัฒนาธุรกิจการค้า)",
      statementtype: "balance_sheet",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: true, comparisontype: "none" },
      rows: [
        { id: "bs-1", rowno: 10, rowtype: "header", title: "สินทรัพย์", style: { fontweight: "bold", indent: 0 } },
        { id: "bs-2", rowno: 20, rowtype: "header", title: "สินทรัพย์หมุนเวียน", style: { fontweight: "bold", indent: 1 } },
        { id: "bs-3", rowno: 30, rowtype: "account", title: "เงินสดและรายการเทียบเท่าเงินสด", noteno: "3", accountcodes: ["1111-01", "1111-02", "1112-01"], normalbalance: "debit", style: { indent: 2 } },
        { id: "bs-4", rowno: 40, rowtype: "account", title: "ลูกหนี้การค้าและลูกหนี้อื่น", noteno: "4", accountcodes: ["1131-01", "1131-02"], normalbalance: "debit", style: { indent: 2 } },
        { id: "bs-5", rowno: 50, rowtype: "account", title: "สินค้าคงเหลือ", noteno: "5", accountcodes: ["1141-01"], normalbalance: "debit", style: { indent: 2 } },
        { id: "bs-6", rowno: 60, rowtype: "account", title: "สินทรัพย์หมุนเวียนอื่น", noteno: "6", accountcodes: ["1151-01"], normalbalance: "debit", style: { indent: 2 } },
        { id: "bs-7", rowno: 70, rowtype: "subtotal", title: "รวมสินทรัพย์หมุนเวียน", formula: "SUM(R30:R60)", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "bs-8", rowno: 80, rowtype: "blank", title: "" },
        { id: "bs-9", rowno: 90, rowtype: "header", title: "สินทรัพย์ไม่หมุนเวียน", style: { fontweight: "bold", indent: 1 } },
        { id: "bs-10", rowno: 100, rowtype: "account", title: "ที่ดิน อาคาร และอุปกรณ์", noteno: "7", accountcodes: ["1211-01", "1221-01"], normalbalance: "debit", style: { indent: 2 } },
        { id: "bs-11", rowno: 110, rowtype: "account", title: "ค่าเสื่อมราคาสะสม", noteno: "7", accountcodes: ["1222-01"], normalbalance: "credit", reversesign: true, style: { indent: 2 } },
        { id: "bs-12", rowno: 120, rowtype: "account", title: "สินทรัพย์ไม่หมุนเวียนอื่น", noteno: "8", accountcodes: ["1251-01"], normalbalance: "debit", style: { indent: 2 } },
        { id: "bs-13", rowno: 130, rowtype: "subtotal", title: "รวมสินทรัพย์ไม่หมุนเวียน", formula: "SUM(R100:R120)", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "bs-14", rowno: 140, rowtype: "formula", title: "รวมสินทรัพย์ทั้งสิ้น", formula: "R70 + R130", style: { fontweight: "bold", indent: 0, underline: "double" } },
        { id: "bs-15", rowno: 150, rowtype: "blank", title: "" },
        { id: "bs-16", rowno: 160, rowtype: "header", title: "หนี้สินและส่วนของเจ้าของ", style: { fontweight: "bold", indent: 0 } },
        { id: "bs-17", rowno: 170, rowtype: "header", title: "หนี้สินหมุนเวียน", style: { fontweight: "bold", indent: 1 } },
        { id: "bs-18", rowno: 180, rowtype: "account", title: "เจ้าหนี้การค้าและเจ้าหนี้อื่น", noteno: "9", accountcodes: ["2121-01", "2131-01"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-19", rowno: 190, rowtype: "account", title: "เงินกู้ยืมระยะสั้น", noteno: "10", accountcodes: ["2111-01"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-20", rowno: 200, rowtype: "account", title: "หนี้สินหมุนเวียนอื่น", noteno: "11", accountcodes: ["2191-01"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-21", rowno: 210, rowtype: "subtotal", title: "รวมหนี้สินหมุนเวียน", formula: "SUM(R180:R200)", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "bs-22", rowno: 220, rowtype: "blank", title: "" },
        { id: "bs-23", rowno: 230, rowtype: "header", title: "หนี้สินไม่หมุนเวียน", style: { fontweight: "bold", indent: 1 } },
        { id: "bs-24", rowno: 240, rowtype: "account", title: "เงินกู้ยืมระยะยาว", noteno: "12", accountcodes: ["2211-01"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-25", rowno: 250, rowtype: "subtotal", title: "รวมหนี้สินไม่หมุนเวียน", formula: "R240", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "bs-26", rowno: 260, rowtype: "formula", title: "รวมหนี้สินทั้งสิ้น", formula: "R210 + R250", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "bs-27", rowno: 270, rowtype: "blank", title: "" },
        { id: "bs-28", rowno: 280, rowtype: "header", title: "ส่วนของเจ้าของ", style: { fontweight: "bold", indent: 1 } },
        { id: "bs-29", rowno: 290, rowtype: "account", title: "ทุนจดทะเบียนและชำระแล้ว", noteno: "13", accountcodes: ["3111-01"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-30", rowno: 300, rowtype: "account", title: "กำไรสะสม", noteno: "14", accountcodes: ["3211-01"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-31", rowno: 310, rowtype: "account", title: "กำไร (ขาดทุน) สุทธิประจำงวด", noteno: "14", accountcodes: ["__current_earnings__"], normalbalance: "credit", style: { indent: 2 } },
        { id: "bs-32", rowno: 320, rowtype: "subtotal", title: "รวมส่วนของเจ้าของ", formula: "SUM(R290:R310)", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "bs-33", rowno: 330, rowtype: "formula", title: "รวมหนี้สินและส่วนของเจ้าของทั้งสิ้น", formula: "R260 + R320", style: { fontweight: "bold", indent: 0, underline: "double" } },
      ],
    },
    {
      code: "PNL-FUNCTION",
      name: "งบกำไรขาดทุน (จำแนกค่าใช้จ่ายตามหน้าที่)",
      statementtype: "pnl",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: true, comparisontype: "none" },
      rows: [
        { id: "pnl-1", rowno: 10, rowtype: "header", title: "รายได้", style: { fontweight: "bold", indent: 0 } },
        { id: "pnl-2", rowno: 20, rowtype: "account", title: "รายได้จากการขายและบริการ", noteno: "15", accountcodes: ["4111-01", "4111-02"], normalbalance: "credit", style: { indent: 1 } },
        { id: "pnl-3", rowno: 30, rowtype: "account", title: "รายได้อื่นๆ", noteno: "16", accountcodes: ["4211-01"], normalbalance: "credit", style: { indent: 1 } },
        { id: "pnl-4", rowno: 40, rowtype: "subtotal", title: "รวมรายได้", formula: "R20 + R30", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "pnl-5", rowno: 50, rowtype: "blank", title: "" },
        { id: "pnl-6", rowno: 60, rowtype: "header", title: "ค่าใช้จ่าย", style: { fontweight: "bold", indent: 0 } },
        { id: "pnl-7", rowno: 70, rowtype: "account", title: "ต้นทุนขายและบริการ", noteno: "17", accountcodes: ["5111-01"], normalbalance: "debit", style: { indent: 1 } },
        { id: "pnl-8", rowno: 80, rowtype: "formula", title: "กำไร (ขาดทุน) ขั้นต้น", formula: "R20 - R70", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "pnl-9", rowno: 90, rowtype: "account", title: "ค่าใช้จ่ายในการขาย", noteno: "18", accountcodes: ["5211-01"], normalbalance: "debit", style: { indent: 1 } },
        { id: "pnl-10", rowno: 100, rowtype: "account", title: "ค่าใช้จ่ายในการบริหาร", noteno: "19", accountcodes: ["5311-01"], normalbalance: "debit", style: { indent: 1 } },
        { id: "pnl-11", rowno: 110, rowtype: "subtotal", title: "รวมค่าใช้จ่ายในการดำเนินงาน", formula: "R90 + R100", style: { fontweight: "bold", indent: 1, underline: "single" } },
        { id: "pnl-12", rowno: 120, rowtype: "formula", title: "กำไร (ขาดทุน) ก่อนต้นทุนทางการเงินและภาษี", formula: "R40 - R70 - R110", style: { fontweight: "bold", indent: 0 } },
        { id: "pnl-13", rowno: 130, rowtype: "account", title: "ต้นทุนทางการเงิน (ดอกเบี้ยจ่าย)", noteno: "20", accountcodes: ["5411-01"], normalbalance: "debit", style: { indent: 1 } },
        { id: "pnl-14", rowno: 140, rowtype: "account", title: "ค่าใช้จ่ายภาษีเงินได้", noteno: "21", accountcodes: ["5511-01"], normalbalance: "debit", style: { indent: 1 } },
        { id: "pnl-15", rowno: 150, rowtype: "formula", title: "กำไร (ขาดทุน) สุทธิสำหรับงวด", formula: "R120 - R130 - R140", style: { fontweight: "bold", indent: 0, underline: "double" } },
      ],
    },
    {
      code: "COGS-STMT",
      name: "งบต้นทุนผลิตและต้นทุนขาย",
      statementtype: "production_cost",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: false, comparisontype: "none" },
      rows: [
        { id: "cog-1", rowno: 10, rowtype: "header", title: "วัตถุดิบทางตรงที่ใช้ไป", style: { fontweight: "bold", indent: 0 } },
        { id: "cog-2", rowno: 20, rowtype: "account", title: "วัตถุดิบต้นงวด", accountcodes: ["1141-10"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-3", rowno: 30, rowtype: "account", title: "บวก: ซื้อวัตถุดิบสุทธิ", accountcodes: ["5111-10"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-4", rowno: 40, rowtype: "account", title: "หัก: วัตถุดิบปลายงวด", accountcodes: ["1141-10"], normalbalance: "debit", reversesign: true, style: { indent: 1 } },
        { id: "cog-5", rowno: 50, rowtype: "subtotal", title: "วัตถุดิบทางตรงใช้ไปในการผลิต", formula: "R20 + R30 + R40", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cog-6", rowno: 60, rowtype: "account", title: "ค่าแรงทางตรง", accountcodes: ["5111-20"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-7", rowno: 70, rowtype: "account", title: "ค่าใช้จ่ายการผลิต", accountcodes: ["5111-30"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-8", rowno: 80, rowtype: "subtotal", title: "รวมต้นทุนการผลิตงวดนี้", formula: "R50 + R60 + R70", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cog-9", rowno: 90, rowtype: "account", title: "บวก: งานระหว่างทำต้นงวด", accountcodes: ["1141-20"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-10", rowno: 100, rowtype: "account", title: "หัก: งานระหว่างทำปลายงวด", accountcodes: ["1141-20"], normalbalance: "debit", reversesign: true, style: { indent: 1 } },
        { id: "cog-11", rowno: 110, rowtype: "formula", title: "ต้นทุนสินค้าสำเร็จรูป", formula: "R80 + R90 + R100", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cog-12", rowno: 120, rowtype: "account", title: "บวก: สินค้าสำเร็จรูปต้นงวด", accountcodes: ["1141-01"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-13", rowno: 130, rowtype: "account", title: "หัก: สินค้าสำเร็จรูปปลายงวด", accountcodes: ["1141-01"], normalbalance: "debit", reversesign: true, style: { indent: 1 } },
        { id: "cog-14", rowno: 140, rowtype: "formula", title: "ต้นทุนขายทั้งสิ้น", formula: "R110 + R120 + R130", style: { fontweight: "bold", indent: 0, underline: "double" } },
      ],
    },
    {
      code: "CASH-FLOW-IND",
      name: "งบกระแสเงินสด (วิธีทางอ้อม)",
      statementtype: "cash_flow",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: false, comparisontype: "none" },
      rows: [
        { id: "cf-1", rowno: 10, rowtype: "header", title: "กระแสเงินสดจากกิจกรรมดำเนินงาน", style: { fontweight: "bold", indent: 0 } },
        { id: "cf-2", rowno: 20, rowtype: "account", title: "กำไร (ขาดทุน) สุทธิประจำงวด", accountcodes: ["__current_earnings__"], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-3", rowno: 30, rowtype: "account", title: "ปรับปรุง: ค่าเสื่อมราคาและค่าตัดจำหน่าย", accountcodes: ["5311-90"], normalbalance: "debit", style: { indent: 1 } },
        { id: "cf-4", rowno: 40, rowtype: "account", title: "การเปลี่ยนแปลงในลูกหนี้การค้า (เพิ่มขึ้น) ลดลง", accountcodes: ["1131-01"], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-5", rowno: 50, rowtype: "account", title: "การเปลี่ยนแปลงในสินค้าคงเหลือ (เพิ่มขึ้น) ลดลง", accountcodes: ["1141-01"], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-6", rowno: 60, rowtype: "account", title: "การเปลี่ยนแปลงในเจ้าหนี้การค้า เพิ่มขึ้น (ลดลง)", accountcodes: ["2121-01"], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-7", rowno: 70, rowtype: "subtotal", title: "เงินสดสุทธิได้มาจาก (ใช้ไปใน) กิจกรรมดำเนินงาน", formula: "SUM(R20:R60)", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cf-8", rowno: 80, rowtype: "blank", title: "" },
        { id: "cf-9", rowno: 90, rowtype: "header", title: "กระแสเงินสดจากกิจกรรมลงทุน", style: { fontweight: "bold", indent: 0 } },
        { id: "cf-10", rowno: 100, rowtype: "account", title: "เงินสดจ่ายเพื่อซื้อที่ดิน อาคาร และอุปกรณ์", accountcodes: ["1211-01"], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-11", rowno: 110, rowtype: "subtotal", title: "เงินสดสุทธิได้มาจาก (ใช้ไปใน) กิจกรรมลงทุน", formula: "R100", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cf-12", rowno: 120, rowtype: "blank", title: "" },
        { id: "cf-13", rowno: 130, rowtype: "header", title: "กระแสเงินสดจากกิจกรรมจัดหาเงิน", style: { fontweight: "bold", indent: 0 } },
        { id: "cf-14", rowno: 140, rowtype: "account", title: "เงินสดรับจากเงินกู้ยืมระยะสั้น / ระยะยาว", accountcodes: ["2111-01", "2211-01"], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-15", rowno: 150, rowtype: "subtotal", title: "เงินสดสุทธิได้มาจาก (ใช้ไปใน) กิจกรรมจัดหาเงิน", formula: "R140", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cf-16", rowno: 160, rowtype: "formula", title: "เงินสดและรายการเทียบเท่าเงินสดเพิ่มขึ้น (ลดลง) สุทธิ", formula: "R70 + R110 + R150", style: { fontweight: "bold", indent: 0, underline: "double" } },
      ],
    },
  ];
}
