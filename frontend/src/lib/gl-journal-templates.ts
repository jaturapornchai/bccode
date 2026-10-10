import { emptyLine, type GLJournal, type GLJournalTemplate, type GLJournalTemplateBody, type GLJournalTemplateLine, type GLLine } from "./general-ledger";
import type { GLJournalDetails } from "./gl-journal-details";
import { setTaxBoxType } from "./gl-journal-entry";

// Helpers of the journal entry screen (Champ CGLFrmJournal, docs/kms/champ/09-gl.md 2.6 and 2.16): daily-entry templates
// (Champ "recurring": OnRecur saves a named pattern, OnSelRecure fills the screen from one) and the debit/credit jump
// (OnToDebit/OnToCredit). No money is calculated here; a template never holds an amount.

/** Champ BCTrans.RecurName is VARCHAR(40) (Script/SQLSERVER_Script.sql:9404); counted in code points like the backend counts runes. */
export const JOURNAL_TEMPLATE_NAME_MAX = 40;
export const JOURNAL_TEMPLATE_MAX_LINES = 500;

export function journalTemplateName(raw: string): string {
  return raw.trim().normalize("NFC");
}

export type JournalTemplateNameProblem = { problem: "required" } | { problem: "too_long"; length: number } | null;
export function journalTemplateNameProblem(name: string): JournalTemplateNameProblem {
  const clean = journalTemplateName(name);
  if (!clean) return { problem: "required" };
  const length = Array.from(clean).length;
  return length > JOURNAL_TEMPLATE_NAME_MAX ? { problem: "too_long", length } : null;
}

function templateTaxType(details: GLJournalDetails | undefined): 1 | 2 | undefined {
  const types = new Set((details?.vats ?? []).map((row) => Number(row.tax_type)));
  if (types.size !== 1) return undefined;
  return types.has(2) ? 2 : types.has(1) ? 1 : undefined;
}

/** Lines that carry an account code: the pattern of the entry the user is looking at (Champ keeps the whole saved document;
 *  every figure of it is dropped again when the template is loaded, so none is stored here). */
export function templateLinesFromJournal(journal: GLJournal): GLJournalTemplateLine[] {
  return journal.lines
    .filter((line) => (line.accountcode ?? "").trim() !== "")
    .slice(0, JOURNAL_TEMPLATE_MAX_LINES)
    .map((line) => ({
      accountcode: line.accountcode.trim(), description: (line.description ?? "").trim(),
      departmentcode: (line.departmentcode ?? "").trim(), projectcode: (line.projectcode ?? "").trim(), cashflow: line.cashflow ?? "",
    }));
}

/** What the Recur button stores: book, reference, description and tax type of the header plus the account/detail/dimension of
 *  every line. Never the document number, date, debit/credit amounts or invoice figures. */
export function templateBodyFromJournal(journal: GLJournal): GLJournalTemplateBody {
  const taxtype = templateTaxType(journal.details);
  return {
    bookcode: journal.bookcode, reference: journal.reference ?? "", refdate: journal.refdate ?? "", description: journal.description ?? "",
    ...(taxtype ? { taxtype } : {}), lines: templateLinesFromJournal(journal),
  };
}

/**
 * Fill the screen from a template (Champ OnSelRecure + OnLoadData with m_bLoadRecurring, glfrmjournal.cpp:988-1090):
 * book, reference, reference date, description, tax type and the lines (account, detail, department, project, cash flow) come
 * from the template; the document number stays empty (the backend numbers it), amounts are 0, the tax box has the tax type
 * but no invoice number or amount. The document date, fiscal year and branch stay those of `base` — pass the new draft on
 * the screen or a blank one, never a saved voucher. A template book that is no longer usable keeps the base book.
 */
export function journalFromTemplate(template: GLJournalTemplate, base: GLJournal, usableBookCodes: readonly string[]): GLJournal {
  const body = template.template;
  const lines: GLLine[] = body.lines.map((line) => ({
    ...emptyLine(), accountcode: line.accountcode, description: line.description ?? "",
    departmentcode: line.departmentcode ?? "", projectcode: line.projectcode ?? "", cashflow: line.cashflow ?? "",
  }));
  while (lines.length < 2) lines.push(emptyLine());
  return {
    ...base, docno: "",
    bookcode: body.bookcode && usableBookCodes.includes(body.bookcode) ? body.bookcode : base.bookcode,
    reference: body.reference ?? "", refdate: body.refdate ?? "", description: body.description ?? "",
    lines, details: body.taxtype ? setTaxBoxType(undefined, body.taxtype === 2 ? "2" : "1", base.date) : undefined,
  };
}

export type JournalJumpSide = "debit" | "credit";
export type JournalJumpKeyEvent = { key: string; code?: string; ctrlKey?: boolean; metaKey?: boolean; altKey?: boolean; shiftKey?: boolean; defaultPrevented?: boolean };

/**
 * Alt+D = go to the debit cell, Alt+C = go to the credit cell of the current line (Champ OnToDebit/OnToCredit,
 * glfrmjournal.cpp:2583-2591). Champ binds these two commands to no key and no visible control (no accelerator, menu or
 * dialog control in BC5Account.rc or menuconfig.xml), so the keys follow Alt+N: Alt + the first letter. `code` is checked too
 * because a Thai keyboard layout reports Thai letters in `key`.
 */
export function journalJumpKey(event: JournalJumpKeyEvent): JournalJumpSide | null {
  if (event.defaultPrevented || !event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return null;
  if (event.code === "KeyD" || event.key.toLowerCase() === "d") return "debit";
  if (event.code === "KeyC" || event.key.toLowerCase() === "c") return "credit";
  return null;
}

/** Row of the jump: the row that has focus, else the last row the user worked in (a click on a jump button moves focus
 *  to the button), else the first; always inside the table. */
export function jumpRowIndex(focusedRow: number | null, rememberedRow: number, lineCount: number): number {
  const last = Math.max(0, lineCount - 1);
  const wanted = focusedRow ?? rememberedRow;
  return Math.min(Math.max(0, Number.isInteger(wanted) ? wanted : 0), last);
}

export interface SMEPresetTemplate {
  id: string;
  name: string;
  category: "expense" | "income" | "payroll";
  bookcode: string;
  description: string;
  badge: string;
  lines: { accountcode: string; description: string }[];
}

/** แม่แบบบันทึกรายวันสำเร็จรูปสำหรับ SMEs ไทย ตามผังบัญชีมาตรฐานและประมวลรัษฎากร */
export const SME_PRESET_TEMPLATES: SMEPresetTemplate[] = [
  {
    id: "preset-rent",
    name: "จ่ายค่าเช่าสำนักงาน (หัก ณ ที่จ่าย 5%)",
    category: "expense",
    bookcode: "PV",
    badge: "หัก 5%",
    description: "จ่ายค่าเช่าสำนักงานประจำเดือน",
    lines: [
      { accountcode: "53210", description: "ค่าเช่าสำนักงาน" },
      { accountcode: "21264", description: "ภาษีหัก ณ ที่จ่ายค้างจ่าย (5%)" },
      { accountcode: "11131", description: "เงินฝากกระแสรายวัน" },
    ],
  },
  {
    id: "preset-service-fee",
    name: "จ่ายค่าบริการ / ค่าจ้างทำของ (หัก ณ ที่จ่าย 3%)",
    category: "expense",
    bookcode: "PV",
    badge: "หัก 3%",
    description: "จ่ายค่าบริการวิชาชีพและค่าจ้างทำของ",
    lines: [
      { accountcode: "53310", description: "ค่าสอบบัญชีและบริการวิชาชีพ" },
      { accountcode: "21264", description: "ภาษีหัก ณ ที่จ่ายค้างจ่าย (3%)" },
      { accountcode: "11131", description: "เงินฝากกระแสรายวัน" },
    ],
  },
  {
    id: "preset-utilities",
    name: "จ่ายค่าน้ำประปา / ค่าไฟฟ้า (ภาษีซื้อ 7%)",
    category: "expense",
    bookcode: "PV",
    badge: "VAT 7%",
    description: "จ่ายค่าน้ำประปา ค่าไฟฟ้า ค่าสาธารณูปโภค",
    lines: [
      { accountcode: "53210", description: "ค่าสาธารณูปโภค" },
      { accountcode: "11351", description: "ภาษีซื้อ" },
      { accountcode: "11110", description: "เงินสด" },
    ],
  },
  {
    id: "preset-salary",
    name: "จ่ายเงินเดือนพนักงาน (พร้อมประกันสังคม)",
    category: "payroll",
    bookcode: "PV",
    badge: "เงินเดือน",
    description: "จ่ายเงินเดือนพนักงานประจำเดือน",
    lines: [
      { accountcode: "53111", description: "เงินเดือนและค่าจ้าง" },
      { accountcode: "53121", description: "เงินสมทบประกันสังคม" },
      { accountcode: "21243", description: "เงินประกันสังคมค้างจ่าย" },
      { accountcode: "11131", description: "เงินฝากกระแสรายวัน" },
    ],
  },
  {
    id: "preset-ar-receive",
    name: "รับชำระหนี้จากลูกหนี้ (ถูกหัก ณ ที่จ่าย 3%)",
    category: "income",
    bookcode: "RV",
    badge: "ถูกหัก 3%",
    description: "รับชำระหนี้ค่าบริการจากลูกหนี้การค้า",
    lines: [
      { accountcode: "11131", description: "เงินฝากกระแสรายวัน" },
      { accountcode: "11353", description: "ภาษีเงินได้ถูกหัก ณ ที่จ่าย (3%)" },
      { accountcode: "11310", description: "ลูกหนี้การค้า" },
    ],
  },
  {
    id: "preset-service-income",
    name: "บันทึกรายได้ค่าบริการ",
    category: "income",
    bookcode: "SV",
    badge: "รายได้",
    description: "บันทึกรายได้ค่าบริการ",
    lines: [
      { accountcode: "11310", description: "ลูกหนี้การค้า" },
      { accountcode: "41210", description: "รายได้ค่าบริการ" },
    ],
  },
];

export function presetToJournalTemplate(preset: SMEPresetTemplate): GLJournalTemplate {
  return {
    id: preset.id,
    code: preset.id,
    version: 1,
    name: preset.name,
    isactive: true,
    template: {
      bookcode: preset.bookcode,
      description: preset.description,
      lines: preset.lines.map((line) => ({
        accountcode: line.accountcode,
        description: line.description,
      })),
    },
  };
}
