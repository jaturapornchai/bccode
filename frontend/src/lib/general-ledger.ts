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
/** งบประมาณรายเดือน (resource "budgets", ADR 2026-09-25-gl-monthly-budget): เงินเป็นข้อความทศนิยม 12 งวดต่อบัญชี */
export type GLBudgetLine = { accountcode: string; accountname?: string; periods: string[]; total?: string };
export type GLBudget = GLIdentity & { code: string; name: string; fiscalyear: string; branchcode: string; departmentcode: string; projectcode: string; status: "open" | "closed"; remark: string; lines?: GLBudgetLine[]; total?: string };
export const BUDGET_PERIODS = 12;
/** วันเริ่มของแต่ละงวดในปีบัญชี: งวด 1 = วันเริ่มปี งวดถัดไป = วันที่ 1 ของเดือนถัดไป ตัดงวดที่เริ่มหลังวันสิ้นปี (ตรงกับ fiscalPeriodStarts ใน backend budgets.go) */
export function budgetPeriodStarts(year: Pick<GLFiscalYear, "startdate" | "enddate"> | undefined): string[] {
  if (!year?.startdate || !year.enddate || !/^\d{4}-\d{2}-\d{2}$/.test(year.startdate)) return [];
  const [startYear, startMonth] = year.startdate.split("-").map(Number);
  const starts = [year.startdate];
  for (let offset = 1; offset < BUDGET_PERIODS; offset++) {
    const monthIndex = startMonth - 1 + offset;
    const start = `${startYear + Math.floor(monthIndex / 12)}-${String((monthIndex % 12) + 1).padStart(2, "0")}-01`;
    if (start > year.enddate) break;
    starts.push(start);
  }
  return starts;
}
/** ผลรวมยอดงวดแบบทศนิยมตรง (BigInt) — ช่องว่างนับเป็นศูนย์ */
export function budgetPeriodsTotal(periods: string[]): bigint { return periods.reduce((sum, value) => sum + amountUnits(blankAmountAsZero(value)), 0n); }
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
/** สมุดรายวันเป็นข้อมูลหลักที่ผู้ใช้กำหนดเอง (mydocs/datamodels/gl/journalbook.sql) — ห้ามอิงรหัสสมุด ใช้ booktype เท่านั้น
 *  booktype 1=ทั่วไป 2=จ่าย 3=รับ 4=ขาย 5=ซื้อ 6=ยอดยกมา; 0/ไม่มี = ยังไม่กำหนด (ใช้กับใบใหม่ไม่ได้) */
export type GLJournalBook = GLIdentity & { code: string; name: string; nameen?: string; booktype?: number; isactive: boolean };
export type GLRecord = GLAccount | GLFiscalYear | GLMaster | GLJournal | GLStatementTemplate | GLJournalBook;
export type GLReviewStatus = 1 | 2 | 3;
export type GLReviewEvent = { eventno: number; version: number; status: GLReviewStatus; note: string; reviewedby: string; reviewedat: string };
export type GLJournalReview = { journalid: string; version: number; status: GLReviewStatus; eventno: number; events: GLReviewEvent[] };
export type GLPage<T> = { items: T[]; total: number; page: number; limit: number; sequence: number };
export type GLReport = {
  columns: { key: string; label: string; amount?: boolean }[];
  rows: Record<string, string>[]; totals: Record<string, string>;
  totalrows: number; warnings: string[]; asof: string; sequence: number;
  // statement: ช่วงของแต่ละคอลัมน์ยอดเงิน (ปีนี้/ปีก่อน) สำหรับหัวงบ
  periods?: { key: string; fiscalyear: string; from: string; to: string }[];
};
export const GL_RESOURCES = ["accounts", "fiscal-years", "account-groups", "product-account-groups", "mappings", "budgets", "periods", "forecast", "allocations", "journals", "statement-templates", "journal-books"] as const;
export type GLResource = typeof GL_RESOURCES[number];
export type GLCommand = {
  resource: GLResource | "processes"; id?: string; action: string; requestid: string;
  version?: number; reason?: string; date?: string; docno?: string; targetyear?: string;
  account?: GLAccount; fiscalyear?: GLFiscalYear; master?: GLMaster | GLJournalBook; journal?: GLJournal | Pick<GLJournal, "details">; budget?: Omit<GLBudget, keyof GLIdentity | "total">; statementtemplate?: GLStatementTemplate;
  review?: { status: GLReviewStatus; note: string; expectedEventNo: number };
};
export const GL_REPORTS = ["ledger", "trialbalance", "pnl", "balancesheet", "workingpaper", "gljournal", "budgetcomparison", "ar-outstanding", "ap-outstanding", "bank-unmatched", "statement"] as const;
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
export const accountTypes = thaiLabels(accountTypeLabels);
export const JOURNAL_BOOK_TYPES = [1, 2, 3, 4, 5, 6] as const;
export type GLJournalBookType = (typeof JOURNAL_BOOK_TYPES)[number];
/** ป้ายประเภทสมุดตาม journalbook.sql (ค่าตัวเลขเป็น key; ชื่อสมุดจริงมาจากข้อมูลหลักของบริษัท) */
export const journalBookTypeLabels: Record<string, GLLabel> = { 1: ["gl_general_journal", "รายวันทั่วไป"], 2: ["gl_cash_payments_journal", "รายวันจ่ายเงิน"], 3: ["gl_cash_receipts_journal", "รายวันรับเงิน"], 4: ["gl_sales_journal", "รายวันขาย"], 5: ["gl_purchase_journal", "รายวันซื้อ"], 6: ["gl_opening_balance_2", "ยอดยกมา"] };
export const JOURNAL_BOOK_CODE_MAX = 15;
export const JOURNAL_BOOK_NAME_MAX = 100;
export function isJournalBookType(value: unknown): value is GLJournalBookType { return Number.isInteger(Number(value)) && JOURNAL_BOOK_TYPES.includes(Number(value) as GLJournalBookType); }
export function emptyJournalBook(): GLJournalBook { return { code: "", name: "", nameen: "", booktype: 0, isactive: true }; }
/** payload ของคำสั่ง create/update สมุดรายวัน: ส่งเฉพาะช่องตามสัญญา ไม่พ่วงช่องของข้อมูลหลักอื่น */
export function journalBookPayload(book: GLJournalBook): GLJournalBook {
  return { ...(book.id ? { id: book.id } : {}), ...(book.version ? { version: book.version } : {}), code: book.code.trim(), name: book.name.trim(), nameen: (book.nameen ?? "").trim(), booktype: Number(book.booktype) || 0, isactive: book.isactive ?? true };
}
/** ตรวจก่อนส่ง: ข้อความบอกช่องที่ผิดและวิธีแก้ (field = data-field ของช่อง) — นับความยาวเป็นตัวอักษร ไม่ใช่ไบต์
 *  allowUntyped: แก้สมุดเดิมที่ยังไม่มีประเภท — เปลี่ยนชื่อ/ปิดใช้งานได้โดยไม่ต้องเลือกประเภท (ตรงกับ backend allowUntyped) */
export function validateJournalBook(book: GLJournalBook, tr: GLTextFn = (_key, fallback) => fallback, allowUntyped = false): { message: string; field: string } | null {
  const code = book.code.trim(), name = book.name.trim(), nameen = (book.nameen ?? "").trim();
  if (!code) return { message: tr("gl_book_code_required", "กรุณาระบุรหัสสมุดรายวัน เช่น JV"), field: "code" };
  if ([...code].length > JOURNAL_BOOK_CODE_MAX) return { message: tr("gl_book_code_too_long", "รหัสสมุดรายวันยาวได้ไม่เกิน {0} ตัวอักษร (ตอนนี้ {1} ตัว)").replace("{0}", String(JOURNAL_BOOK_CODE_MAX)).replace("{1}", String([...code].length)), field: "code" };
  if (!name) return { message: tr("gl_book_name_required", "กรุณาระบุชื่อสมุดรายวันภาษาไทย"), field: "name" };
  if ([...name].length > JOURNAL_BOOK_NAME_MAX) return { message: tr("gl_book_name_too_long", "ชื่อสมุดรายวันยาวได้ไม่เกิน {0} ตัวอักษร").replace("{0}", String(JOURNAL_BOOK_NAME_MAX)), field: "name" };
  if ([...nameen].length > JOURNAL_BOOK_NAME_MAX) return { message: tr("gl_book_name_en_too_long", "ชื่อสมุดรายวันภาษาอังกฤษยาวได้ไม่เกิน {0} ตัวอักษร").replace("{0}", String(JOURNAL_BOOK_NAME_MAX)), field: "nameen" };
  if (!isJournalBookType(book.booktype) && !(allowUntyped && !Number(book.booktype))) return { message: tr("gl_book_type_required", "กรุณาเลือกประเภทสมุดรายวัน (ทั่วไป จ่าย รับ ขาย ซื้อ หรือยอดยกมา)"), field: "booktype" };
  return null;
}
/** สมุดที่ใช้บันทึกใบใหม่ได้: เปิดใช้งาน ไม่ถูกลบ และกำหนดประเภทแล้ว — เรียงตามประเภทแล้วตามรหัส */
/** สมุดเดิมที่เปิดใช้งานแต่ยังไม่กำหนดประเภท (ใช้บันทึกใบใหม่ไม่ได้จนกว่าจะเลือกประเภท) — เรียงตามรหัส */
export function untypedJournalBooks(books: GLJournalBook[]): GLJournalBook[] {
  return books.filter((book) => book.isactive && !book.isdeleted && !isJournalBookType(book.booktype))
    .sort((a, b) => a.code.localeCompare(b.code, undefined, { numeric: true, sensitivity: "base" }));
}
export function activeJournalBooks(books: GLJournalBook[]): GLJournalBook[] {
  return books.filter((book) => book.isactive && !book.isdeleted && isJournalBookType(book.booktype))
    .sort((a, b) => Number(a.booktype) - Number(b.booktype) || a.code.localeCompare(b.code, undefined, { numeric: true, sensitivity: "base" }));
}
export function findJournalBook(books: GLJournalBook[], code: string): GLJournalBook | undefined { return books.find((book) => book.code === code && !book.isdeleted); }
/** ชื่อสมุดตามภาษาที่เลือก: ไทยใช้ชื่อไทย ภาษาอื่นใช้ชื่ออังกฤษถ้ามี — ไม่รู้จักรหัส = แสดงรหัสตรง ๆ */
export function journalBookName(book: GLJournalBook | undefined, code: string, language = "th"): string {
  if (!book) return code;
  return language !== "th" && book.nameen?.trim() ? book.nameen.trim() : book.name || code;
}
/** สมุดเริ่มต้นของใบใหม่: สมุดที่ขอ (แท็บที่เลือก) ถ้ายังใช้ได้ → สมุดประเภททั่วไปตัวแรก → สมุดที่ใช้ได้ตัวแรก → ว่าง (ให้ผู้ใช้เลือก) */
export function defaultJournalBookCode(books: GLJournalBook[], preferred = ""): string {
  const usable = activeJournalBooks(books);
  return usable.find((book) => book.code === preferred)?.code ?? usable.find((book) => Number(book.booktype) === 1)?.code ?? usable[0]?.code ?? "";
}
/** สมุดของใบใหม่หรือเมื่อเปลี่ยนสมุดต้องเปิดใช้งานและกำหนดประเภทแล้ว; ใบเดิมที่ไม่เปลี่ยนสมุดใช้สมุดที่ปิดแล้วต่อได้ */
export function journalBookProblem(bookcode: string, books: GLJournalBook[], previousBookcode: string | undefined, tr: GLTextFn = (_key, fallback) => fallback): string | null {
  const book = findJournalBook(books, bookcode);
  if (!bookcode || !book) return tr("gl_journal_book_required", "กรุณาเลือกสมุดรายวัน — ถ้ายังไม่มีสมุดให้เพิ่มที่เมนู กำหนดสมุดรายวัน");
  if (previousBookcode !== undefined && previousBookcode === bookcode) return null;
  if (!book.isactive) return tr("gl_journal_book_inactive", "สมุดรายวัน {0} ปิดใช้งานแล้ว — เลือกสมุดอื่น หรือเปิดใช้งานที่เมนู กำหนดสมุดรายวัน").replace("{0}", bookcode);
  if (!isJournalBookType(book.booktype)) return tr("gl_journal_book_type_missing", "สมุดรายวัน {0} ยังไม่กำหนดประเภท — กำหนดประเภทที่เมนู กำหนดสมุดรายวัน ก่อนใช้บันทึก").replace("{0}", bookcode);
  return null;
}
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
export function emptyMaster(): GLMaster { return { code: "", name: "", isactive: true, accountcode: "", fiscalyear: "", startdate: "", enddate: "", locked: false, amount: "0", branchcode: "", departmentcode: "", projectcode: "", direction: "in", bookcode: "", rules: [], itemaccount: "", costaccount: "", revenueaccount: "", allocatemode: "percent", allocaterules: [] }; }
export function emptyLine(): GLLine { return { accountcode: "", description: "", debit: "0", credit: "0", departmentcode: "", projectcode: "", cashflow: "" }; }
export function localDate() { const date = new Date(); return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`; }
/** ปีบัญชีที่เปิดใช้งาน (ยังไม่ปิด) ซึ่งครอบคลุมวันที่ — ไม่มี = "" ให้ผู้ใช้เลือกเอง */
export function fiscalYearForDate(years: GLFiscalYear[], date: string): string {
  return years.find((year) => year.isactive && !year.closed && !year.isdeleted && date >= year.startdate && date <= year.enddate)?.code ?? "";
}
export function emptyJournal(bookcode = "", kind = "manual", years: GLFiscalYear[] = [], branchcode = ""): GLJournal { const date = localDate(); return { docno: "", date, bookcode, fiscalyear: fiscalYearForDate(years, date), description: "", reference: "", branchcode, kind, status: "draft", lines: [emptyLine(), emptyLine()] }; }
/** สาขาที่ผู้ใช้เลือกตอนเข้าพื้นที่ทำงาน (localStorage bc_workspace → branch.code/guidfixed) — ใบใหม่ใช้เป็นสาขาเริ่มต้น
 *  เพราะ session อาจกลายเป็นระดับบริษัท (บางจอเรียก select-holding โดยไม่ส่งสาขา) แล้ว backend บังคับให้ระบุสาขาเอง */
export function workspaceBranchCode(workspaceJson: string | null): string {
  try {
    const branch = (JSON.parse(workspaceJson ?? "") as { branch?: { code?: string; guidfixed?: string } | null } | null)?.branch;
    return String(branch?.code || branch?.guidfixed || "").trim();
  } catch { return ""; }
}

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

/** ช่องเงินที่ว่าง = ศูนย์ (backend ไม่รับข้อความว่าง) — ใช้ก่อนตรวจ/บันทึกทุกครั้ง ไม่เปลี่ยนค่าที่ผู้ใช้พิมพ์ */
export function blankAmountAsZero(value: string | undefined): string { const text = (value ?? "").trim(); return text === "" ? "0" : text; }
export function normalizeJournalLines(lines: GLLine[]): GLLine[] { return lines.map((line) => ({ ...line, debit: blankAmountAsZero(line.debit), credit: blankAmountAsZero(line.credit) })); }
export function journalTotals(lines: GLLine[]) {
  const debit = lines.reduce((sum, line) => sum + amountUnits(blankAmountAsZero(line.debit)), 0n);
  const credit = lines.reduce((sum, line) => sum + amountUnits(blankAmountAsZero(line.credit)), 0n);
  return { debit, credit, difference: debit - credit };
}
export function validateJournal(journal: GLJournal, year: GLFiscalYear | undefined, accounts: GLAccount[], tr: GLTextFn = (_key, fallback) => fallback): string | null {
  if (!journal.docno.trim() || !journal.date || !journal.description.trim()) return tr("gl_required_no_date_description", "กรุณาระบุเลขที่ วันที่ และคำอธิบายรายการ");
  if (!year || !year.isactive || year.closed || journal.date < year.startdate || journal.date > year.enddate) return tr("gl_select_active_fiscal_year_date", "กรุณาเลือกปีบัญชีที่เปิดใช้งานและวันที่ภายในปีบัญชี");
  if (journal.lines.length < 2 || journal.lines.length > 500) return tr("gl_entries_2_500_lines", "รายการบัญชีต้องมี 2–500 บรรทัด");
  try {
    for (const [index, raw] of journal.lines.entries()) {
      const line = { ...raw, debit: blankAmountAsZero(raw.debit), credit: blankAmountAsZero(raw.credit) };
      const lineNo = String(index + 1);
      const account = accounts.find((item) => item.accountcode === line.accountcode);
      if (!account?.isactive || !account.allowposting || account.isdeleted) return tr("gl_line_select_active_account", "บรรทัดที่ {0}: เลือกบัญชีที่เปิดใช้งานและลงรายการได้").replace("{0}", lineNo);
      // ตรวจทีละช่อง: บอกบรรทัดและช่องที่ผิด พร้อมวิธีแก้ (ไม่ใช้ข้อความรวมที่ไม่บอกบรรทัด)
      const sides = [[line.debit, tr("gl_debit", "เดบิต"), tr("gl_credit", "เครดิต")], [line.credit, tr("gl_credit", "เครดิต"), tr("gl_debit", "เดบิต")]] as const;
      for (const [value, side, opposite] of sides) {
        if (!amountPattern.test(value)) return tr("gl_line_amount_invalid", "บรรทัดที่ {0}: ช่อง{1}ต้องเป็นตัวเลข เช่น 1500.00 (ไม่ใส่จุลภาค ทศนิยมไม่เกิน 8 ตำแหน่ง)").replace("{0}", lineNo).replace("{1}", side);
        if (value.startsWith("-") && decimalUnits(value) !== 0n) return tr("gl_line_amount_negative", "บรรทัดที่ {0}: ยอด{1}ติดลบไม่ได้ — ถ้าเป็นยอดกลับด้าน ให้ใส่เป็นบวกในช่อง{2}").replace("{0}", lineNo).replace("{1}", side).replace("{2}", opposite);
      }
      const debit = amountUnits(line.debit), credit = amountUnits(line.credit);
      if ((debit === 0n) === (credit === 0n)) return tr("gl_line_enter_amount_one_side_only", "บรรทัดที่ {0}: ใส่ยอดมากกว่าศูนย์เพียงด้านเดียว").replace("{0}", lineNo);
      const precision = 10n ** BigInt(8 - year.scale);
      if (debit % precision !== 0n || credit % precision !== 0n) return tr("gl_line_amount_exceeds_decimal_places", "บรรทัดที่ {0}: จำนวนเงินเกิน {1} ตำแหน่งทศนิยม").replace("{0}", lineNo).replace("{1}", String(year.scale));
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

// งบการเงินคำนวณที่ backend (GET reports/statement?template=) — backend/internal/generalledger/statements.go
type StarterRow = [rowno: number, kind: "header" | "blank" | "subtotal" | "total" | "debit" | "credit", title: string, indent?: number, formula?: string, accountcodes?: string[]];
/** แถวแม่แบบงบ: debit/credit = แถวยอดบัญชี (ผู้ใช้เลือกบัญชีเอง), subtotal = รวมย่อยขีดเส้นเดี่ยว, total = ยอดรวมขีดเส้นคู่ */
function starterRows(prefix: string, rows: StarterRow[]): StatementRow[] {
  return rows.map(([rowno, kind, title, indent = 2, formula, accountcodes]) => {
    const id = `${prefix}-${rowno}`;
    if (kind === "header") return { id, rowno, rowtype: "header", title, style: { fontweight: "bold", indent } };
    if (kind === "blank") return { id, rowno, rowtype: "blank", title };
    if (kind === "subtotal" || kind === "total") return { id, rowno, rowtype: "subtotal", title, formula, style: { fontweight: "bold", indent, underline: kind === "total" ? "double" : "single" } };
    return { id, rowno, rowtype: "account", title, noteno: "", accountcodes: accountcodes ?? [], normalbalance: kind, style: { indent } };
  });
}
export function generateStarterTemplates(): GLStatementTemplate[] {
  return [
    {
      code: "BS-DBD",
      // ชื่อแม่แบบ = ชื่องบที่พิมพ์บนหัวงบ; แถวตามแบบ 2 (บริษัทจำกัด) ของประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566 — docs/kms/21
      name: "งบแสดงฐานะการเงิน",
      statementtype: "balance_sheet",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: true, comparisontype: "previous_year" },
      rows: starterRows("bs", [
        [10, "header", "สินทรัพย์", 0], [20, "header", "สินทรัพย์หมุนเวียน", 1],
        [30, "debit", "เงินสดและรายการเทียบเท่าเงินสด"], [40, "debit", "เงินลงทุนชั่วคราว"], [50, "debit", "ลูกหนี้การค้าและลูกหนี้หมุนเวียนอื่น"],
        [60, "debit", "มูลค่าของงานส่วนที่เสร็จแต่ยังไม่ถึงกำหนดเรียกชำระเงิน - หมุนเวียน"], [70, "debit", "เงินให้กู้ยืมระยะสั้น"], [80, "debit", "สินค้าคงเหลือ"],
        [90, "debit", "สินทรัพย์ชีวภาพหมุนเวียน"], [100, "debit", "สินทรัพย์หมุนเวียนอื่น"], [110, "debit", "สินทรัพย์ไม่หมุนเวียนที่ถือไว้เพื่อขาย - หมุนเวียน"],
        [120, "subtotal", "รวมสินทรัพย์หมุนเวียน", 1, "SUM(R30:R110)"],
        [130, "header", "สินทรัพย์ไม่หมุนเวียน", 1],
        [140, "debit", "เงินฝากธนาคารที่มีภาระค้ำประกัน"], [150, "debit", "เงินลงทุนในบริษัทย่อย"], [160, "debit", "เงินลงทุนในการร่วมค้า"], [170, "debit", "เงินลงทุนในบริษัทร่วม"],
        [180, "debit", "เงินลงทุนระยะยาวอื่น"], [190, "debit", "ลูกหนี้การค้าและลูกหนี้ไม่หมุนเวียนอื่น"], [200, "debit", "เงินให้กู้ยืมระยะยาว"],
        [210, "debit", "มูลค่าของงานส่วนที่เสร็จแต่ยังไม่ถึงกำหนดเรียกชำระเงิน - ไม่หมุนเวียน"], [220, "debit", "อสังหาริมทรัพย์เพื่อการลงทุน"], [230, "debit", "ที่ดิน อาคารและอุปกรณ์"],
        [240, "debit", "ค่าความนิยม"], [250, "debit", "สินทรัพย์ไม่มีตัวตน"], [260, "debit", "สินทรัพย์ชีวภาพไม่หมุนเวียน"], [270, "debit", "สินทรัพย์ไม่หมุนเวียนอื่น"],
        [280, "debit", "สินทรัพย์ไม่หมุนเวียนที่ถือไว้เพื่อขาย - ไม่หมุนเวียน"],
        [290, "subtotal", "รวมสินทรัพย์ไม่หมุนเวียน", 1, "SUM(R140:R280)"],
        [300, "total", "รวมสินทรัพย์", 0, "R120 + R290"], [310, "blank", ""],
        [320, "header", "หนี้สินและส่วนของผู้ถือหุ้น", 0], [330, "header", "หนี้สินหมุนเวียน", 1],
        [340, "credit", "เงินเบิกเกินบัญชีและเงินกู้ยืมระยะสั้นจากสถาบันการเงิน"], [350, "credit", "เจ้าหนี้การค้าและเจ้าหนี้หมุนเวียนอื่น"],
        [360, "credit", "เงินรับล่วงหน้าส่วนที่เกินกว่างานส่วนที่เสร็จ - หมุนเวียน"], [370, "credit", "ส่วนของหนี้สินระยะยาวที่ถึงกำหนดชำระภายในหนึ่งปี"],
        [380, "credit", "ส่วนของหนี้สินตามสัญญาเช่าเงินทุนที่ถึงกำหนดชำระภายในหนึ่งปี"], [390, "credit", "เงินกู้ยืมระยะสั้น"], [400, "credit", "ภาษีเงินได้นิติบุคคลค้างจ่าย"],
        [410, "credit", "ประมาณการหนี้สินหมุนเวียนสำหรับผลประโยชน์พนักงาน"], [420, "credit", "ประมาณการหนี้สินระยะสั้นอื่น"], [430, "credit", "หนี้สินหมุนเวียนอื่น"],
        [440, "subtotal", "รวมหนี้สินหมุนเวียน", 1, "SUM(R340:R430)"],
        [450, "header", "หนี้สินไม่หมุนเวียน", 1],
        [460, "credit", "เงินกู้ยืมระยะยาว"], [470, "credit", "หนี้สินตามสัญญาเช่าเงินทุน"], [480, "credit", "เจ้าหนี้การค้าและเจ้าหนี้ไม่หมุนเวียนอื่น"],
        [490, "credit", "เงินรับล่วงหน้าส่วนที่เกินกว่างานส่วนที่เสร็จ - ไม่หมุนเวียน"], [500, "credit", "ประมาณการหนี้สินไม่หมุนเวียนสำหรับผลประโยชน์พนักงาน"],
        [510, "credit", "ประมาณการหนี้สินระยะยาวอื่น"], [520, "credit", "หนี้สินไม่หมุนเวียนอื่น"],
        [530, "subtotal", "รวมหนี้สินไม่หมุนเวียน", 1, "SUM(R460:R520)"],
        [540, "subtotal", "รวมหนี้สิน", 0, "R440 + R530"],
        [550, "header", "ส่วนของผู้ถือหุ้น", 1], [560, "header", "ทุนเรือนหุ้น", 2], [570, "header", "ทุนจดทะเบียน (แสดงจำนวนหุ้นและมูลค่าในหมายเหตุ)", 3],
        [580, "credit", "ทุนที่ชำระแล้ว", 3], [590, "credit", "ส่วนเกินมูลค่าหุ้น"], [600, "credit", "ส่วนเกิน (ต่ำกว่า) ทุนอื่น"],
        [610, "header", "กำไร (ขาดทุน) สะสม", 2], [620, "header", "จัดสรรแล้ว", 3],
        [630, "credit", "ทุนสำรองตามกฎหมาย", 4], [640, "credit", "อื่น ๆ", 4],
        // กำไรขาดทุนที่ยังไม่ปิดบัญชีเป็นส่วนหนึ่งของกำไรสะสมที่ยังไม่ได้จัดสรร; ผู้ใช้เพิ่มบัญชีกำไรสะสมของตนเอง
        [650, "credit", "ยังไม่ได้จัดสรร", 3, undefined, ["__current_earnings__"]],
        [660, "credit", "ส่วนได้เสีย - ทุนอื่น"], [670, "credit", "องค์ประกอบอื่นของส่วนของผู้ถือหุ้น"],
        [680, "subtotal", "รวมส่วนของผู้ถือหุ้น", 1, "SUM(R580:R670)"],
        [690, "total", "รวมหนี้สินและส่วนของผู้ถือหุ้น", 0, "R540 + R680"],
      ]),
    },
    {
      code: "PNL-DBD",
      // แบบ 2 จำแนกค่าใช้จ่ายตามหน้าที่ แบบหลายขั้น
      name: "งบกำไรขาดทุน",
      statementtype: "pnl",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: true, comparisontype: "previous_year" },
      rows: starterRows("pnl", [
        [10, "credit", "รายได้จากการขายหรือการให้บริการ", 1], [20, "debit", "ต้นทุนขายหรือต้นทุนการให้บริการ", 1],
        [30, "subtotal", "กำไร (ขาดทุน) ขั้นต้น", 0, "R10 - R20"],
        [40, "credit", "รายได้อื่น", 1],
        [50, "subtotal", "กำไร (ขาดทุน) ก่อนค่าใช้จ่าย", 0, "R30 + R40"],
        [60, "debit", "ค่าใช้จ่ายในการขาย", 1], [70, "debit", "ค่าใช้จ่ายในการบริหาร", 1], [80, "debit", "ค่าใช้จ่ายอื่น", 1],
        [90, "subtotal", "รวมค่าใช้จ่าย", 0, "SUM(R60:R80)"],
        [100, "subtotal", "กำไร (ขาดทุน) ก่อนต้นทุนทางการเงินและภาษีเงินได้", 0, "R50 - R90"],
        [110, "debit", "ต้นทุนทางการเงิน", 1],
        [120, "subtotal", "กำไร (ขาดทุน) ก่อนภาษีเงินได้", 0, "R100 - R110"],
        [130, "debit", "ภาษีเงินได้", 1],
        [140, "total", "กำไร (ขาดทุน) สุทธิ", 0, "R120 - R130"],
      ]),
    },
    {
      code: "COGS-STMT",
      name: "งบต้นทุนผลิตและต้นทุนขาย",
      statementtype: "production_cost",
      isactive: true,
      globalstyle: { fontfamily: "sarabun", fontsize: "15px", scale: 2, compact: false, shownotecolumn: false, comparisontype: "none" },
      rows: [
        { id: "cog-1", rowno: 10, rowtype: "header", title: "วัตถุดิบทางตรงที่ใช้ไป", style: { fontweight: "bold", indent: 0 } },
        { id: "cog-2", rowno: 20, rowtype: "account", title: "วัตถุดิบต้นงวด", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-3", rowno: 30, rowtype: "account", title: "บวก: ซื้อวัตถุดิบสุทธิ", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-4", rowno: 40, rowtype: "account", title: "หัก: วัตถุดิบปลายงวด", accountcodes: [], normalbalance: "debit", reversesign: true, style: { indent: 1 } },
        { id: "cog-5", rowno: 50, rowtype: "subtotal", title: "วัตถุดิบทางตรงใช้ไปในการผลิต", formula: "R20 + R30 + R40", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cog-6", rowno: 60, rowtype: "account", title: "ค่าแรงทางตรง", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-7", rowno: 70, rowtype: "account", title: "ค่าใช้จ่ายการผลิต", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-8", rowno: 80, rowtype: "subtotal", title: "รวมต้นทุนการผลิตงวดนี้", formula: "R50 + R60 + R70", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cog-9", rowno: 90, rowtype: "account", title: "บวก: งานระหว่างทำต้นงวด", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-10", rowno: 100, rowtype: "account", title: "หัก: งานระหว่างทำปลายงวด", accountcodes: [], normalbalance: "debit", reversesign: true, style: { indent: 1 } },
        { id: "cog-11", rowno: 110, rowtype: "formula", title: "ต้นทุนสินค้าสำเร็จรูป", formula: "R80 + R90 + R100", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cog-12", rowno: 120, rowtype: "account", title: "บวก: สินค้าสำเร็จรูปต้นงวด", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cog-13", rowno: 130, rowtype: "account", title: "หัก: สินค้าสำเร็จรูปปลายงวด", accountcodes: [], normalbalance: "debit", reversesign: true, style: { indent: 1 } },
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
        { id: "cf-3", rowno: 30, rowtype: "account", title: "ปรับปรุง: ค่าเสื่อมราคาและค่าตัดจำหน่าย", accountcodes: [], normalbalance: "debit", style: { indent: 1 } },
        { id: "cf-4", rowno: 40, rowtype: "account", title: "การเปลี่ยนแปลงในลูกหนี้การค้า (เพิ่มขึ้น) ลดลง", accountcodes: [], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-5", rowno: 50, rowtype: "account", title: "การเปลี่ยนแปลงในสินค้าคงเหลือ (เพิ่มขึ้น) ลดลง", accountcodes: [], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-6", rowno: 60, rowtype: "account", title: "การเปลี่ยนแปลงในเจ้าหนี้การค้า เพิ่มขึ้น (ลดลง)", accountcodes: [], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-7", rowno: 70, rowtype: "subtotal", title: "เงินสดสุทธิได้มาจาก (ใช้ไปใน) กิจกรรมดำเนินงาน", formula: "SUM(R20:R60)", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cf-8", rowno: 80, rowtype: "blank", title: "" },
        { id: "cf-9", rowno: 90, rowtype: "header", title: "กระแสเงินสดจากกิจกรรมลงทุน", style: { fontweight: "bold", indent: 0 } },
        { id: "cf-10", rowno: 100, rowtype: "account", title: "เงินสดจ่ายเพื่อซื้อที่ดิน อาคาร และอุปกรณ์", accountcodes: [], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-11", rowno: 110, rowtype: "subtotal", title: "เงินสดสุทธิได้มาจาก (ใช้ไปใน) กิจกรรมลงทุน", formula: "R100", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cf-12", rowno: 120, rowtype: "blank", title: "" },
        { id: "cf-13", rowno: 130, rowtype: "header", title: "กระแสเงินสดจากกิจกรรมจัดหาเงิน", style: { fontweight: "bold", indent: 0 } },
        { id: "cf-14", rowno: 140, rowtype: "account", title: "เงินสดรับจากเงินกู้ยืมระยะสั้น / ระยะยาว", accountcodes: [], normalbalance: "credit", style: { indent: 1 } },
        { id: "cf-15", rowno: 150, rowtype: "subtotal", title: "เงินสดสุทธิได้มาจาก (ใช้ไปใน) กิจกรรมจัดหาเงิน", formula: "R140", style: { fontweight: "bold", indent: 0, underline: "single" } },
        { id: "cf-16", rowno: 160, rowtype: "formula", title: "เงินสดและรายการเทียบเท่าเงินสดเพิ่มขึ้น (ลดลง) สุทธิ", formula: "R70 + R110 + R150", style: { fontweight: "bold", indent: 0, underline: "double" } },
      ],
    },
  ];
}
