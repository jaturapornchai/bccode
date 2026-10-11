import {
  DIMENSIONS, accountAllowedCodes, dimensionEnabled, emptyLine, rowsHoldDimension, visibleDimensions,
  type GLAccount, type GLDimension, type GLDimensionDef, type GLDimensionFlagField, type GLDimensionLineField, type GLDimensionResource, type GLDimensionSettingsInput,
  type GLJournal, type GLLabel, type GLLine,
} from "./general-ledger";
import { vatPeriodPatch, vatPeriodValue, type GLDetailVat, type GLJournalDetails } from "./gl-journal-details";

// ตัวช่วยของจอบันทึกข้อมูลรายวันแบบ Champ (CGLFrmJournal, docs/kms/champ/09-gl.md §2): ตัวเลือกหลังบันทึก + กล่องภาษีบนหัวเอกสาร
// ไม่มีการคำนวณเงินที่นี่ — ยอดภาษีมาจาก backend (GET /gl/v2/vat-calc, generalledger/vat_calc.go) แล้ววางตามตำแหน่งเท่านั้น

/** ตัวเลือกการบันทึก — ลำดับและข้อความตาม combo ของ Champ (glfrmjournal.cpp:338-344, SAVE_NOTCLEAR_SCREEN / SAVE_NEW_REC) */
export type JournalSaveOption = "keep" | "new";
export const JOURNAL_SAVE_OPTIONS: readonly { value: JournalSaveOption; key: string; th: string }[] = [
  { value: "keep", key: "gl_save_option_keep_screen", th: "บันทึกแล้วไม่ล้างหน้าจอ" },
  { value: "new", key: "gl_save_option_new_record", th: "บันทึกแล้วเพิ่มรายการใหม่" },
];
/** จำตัวเลือกแยกตามผู้ใช้ในเครื่องนี้ (ค่าเริ่มต้นแบบ Champ = แถวแรก "บันทึกแล้วไม่ล้างหน้าจอ") */
export function journalSaveOptionStorageKey(username: string | undefined): string {
  return `bc_gl_journal_save_option:${(username ?? "").trim().toLowerCase() || "anonymous"}`;
}
export function parseJournalSaveOption(value: string | null | undefined): JournalSaveOption {
  return value === "new" ? "new" : "keep";
}

/**
 * ใบถัดไปหลังบันทึกสำเร็จ (Champ บันทึกแล้วเข้าโหมดเพิ่มใบใหม่เสมอ — BaseDBFormView.cpp:933-939):
 * - "keep" = OnClearData แบบไม่ล้างหน้าจอ (glfrmjournal.cpp:1139-1164): เก็บสมุด วันที่ อ้างอิง คำอธิบาย สาขา และรหัสบัญชี/รายละเอียด/มิติของทุกบรรทัด
 *   ล้างเลขที่เอกสาร ยอดเดบิต/เครดิต เลขที่ใบกำกับและยอดภาษี (ประเภทภาษี วันที่ใบกำกับ อัตรา คำอธิบายภาษี คงไว้)
 * - "new" = OnClearScreen: ใบเปล่า (blank ที่ผู้เรียกสร้างด้วย emptyJournal)
 */
export function nextJournalAfterSave(saved: GLJournal, option: JournalSaveOption, blank: GLJournal): GLJournal {
  if (option === "new") return blank;
  const vats = saved.details?.vats ?? [];
  const keptVat = vats.length === 1 ? keptTaxBoxRow(vats[0]) : undefined;
  return {
    ...blank,
    bookcode: saved.bookcode, date: saved.date, fiscalyear: saved.fiscalyear, reference: saved.reference ?? "",
    refdate: saved.refdate ?? "", description: saved.description ?? "", branchcode: saved.branchcode ?? blank.branchcode, kind: saved.kind || blank.kind,
    lines: (saved.lines.length ? saved.lines : blank.lines).map((line) => ({ ...emptyLine(), accountcode: line.accountcode ?? "", description: line.description ?? "", departmentcode: line.departmentcode ?? "", projectcode: line.projectcode ?? "", ...extraDimensionCodes(line), cashflow: line.cashflow ?? "" })),
    details: keptVat ? { vats: [keptVat] } : undefined,
  };
}
/**
 * เลขที่เอกสารที่พิมพ์เองซ้ำ: Champ ถามว่าจะรันเลขที่ใหม่ไหม (glfrmjournal.cpp:500-509 + BRunningEdit.cpp:153-178 "Document No is duplicated / Do you want to try again?")
 * เฉพาะใบใหม่ที่พิมพ์เลขที่เอง และ backend ออกเลขให้สมุดนี้ได้ — สมุดที่ไม่ตั้งรูปแบบใช้ค่าเริ่มต้น @YYMM-#### จึงออกเลขได้ทุกสมุดที่มีอยู่จริง
 * (journal_docno.go); ความซ้ำเป็นของ backend (duplicate_code) ผู้เรียกส่งเฉพาะรหัสข้อผิดพลาดมา
 */
export function shouldOfferNewDocNo(errorCode: string | undefined, ctx: { isNew: boolean; typedDocNo: string; bookFound: boolean }): boolean {
  return errorCode === "duplicate_code" && ctx.isNew && ctx.bookFound && ctx.typedDocNo.trim() !== "";
}
function keptTaxBoxRow(row: GLDetailVat): GLDetailVat {
  const type = Number(row.tax_type) === 2 ? 2 : 1;
  return {
    ...newVatRow(row.tax_invoice_date, type),
    vat_rate: row.vat_rate ?? "", remark: row.remark,
    ...(vatPeriodValue(row) ? { tax_period_year: row.tax_period_year, tax_period_month: row.tax_period_month } : {}),
    ...(type === 1 ? { claim_status: row.claim_status ?? 1 } : {}),
  };
}

/** แถวภาษีมูลค่าเพิ่มใหม่ (1 แถว = 1 ใบกำกับภาษี) — อัตราภาษีว่าง ไม่ใส่ 7% ให้เอง (mydocs/specs/gl/tax-design-notes.md:19) */
export function newVatRow(date: string, taxType: 1 | 2): GLDetailVat {
  return {
    id: crypto.randomUUID(), tax_type: taxType, document_type: 1, tax_invoice_no: "", tax_invoice_date: date, ...vatPeriodPatch(date.slice(0, 7)),
    partner_name: "", partner_tax_id: "", partner_branch_no: "", base_amount: "0", zero_rate_amount: "0", exempt_amount: "0", vat_rate: "",
    ...(taxType === 1 ? { claim_status: 1 } : {}),
  };
}

/** กล่องภาษีแก้ได้เฉพาะใบสำคัญที่มีใบกำกับ 0–1 ใบ; 2 ใบขึ้นไปให้แก้ในตารางภาษี F3 (Champ CLIBDlgMultiTax) */
export type TaxBoxView = { mode: "none" } | { mode: "single"; row: GLDetailVat } | { mode: "multi"; count: number };
export function taxBoxView(details: GLJournalDetails | undefined): TaxBoxView {
  const vats = details?.vats ?? [];
  if (vats.length === 0) return { mode: "none" };
  if (vats.length === 1) return { mode: "single", row: vats[0] };
  return { mode: "multi", count: vats.length };
}

const isZeroText = (value: unknown) => /^\s*0*(\.0*)?\s*$/.test(String(value ?? ""));
/** แถวภาษีที่ยังไม่ได้กรอกอะไร (ไม่มีเลขที่ใบกำกับ ยอด หรือคู่ค้า) — Champ ไม่บันทึกภาษีเมื่อไม่มีเลขที่ใบกำกับ (TransType 3, glfrmjournal.cpp:782-784) */
export function isBlankVatRow(row: GLDetailVat): boolean {
  return !String(row.tax_invoice_no ?? "").trim() && !String(row.original_invoice_no ?? "").trim()
    && isZeroText(row.base_amount) && isZeroText(row.zero_rate_amount) && isZeroText(row.exempt_amount) && isZeroText(row.vat_amount)
    && !String(row.partner_code ?? "").trim() && !String(row.partner_name ?? "").trim() && !String(row.partner_tax_id ?? "").trim();
}
/** ก่อนบันทึก: กล่องภาษีที่เลือกประเภทไว้แต่ไม่ได้กรอกอะไรไม่ต้องกลายเป็นรายการภาษี (ไม่แตะกรณีหลายใบ — ตารางภาษีตรวจตามปกติ) */
export function dropBlankTaxBoxRow(details: GLJournalDetails | undefined): GLJournalDetails | undefined {
  const vats = details?.vats ?? [];
  if (!details || vats.length !== 1 || !isBlankVatRow(vats[0])) return details;
  return { ...details, vats: [] };
}

/** เลือกประเภทภาษีในกล่อง: "" = ไม่มีภาษี (นำใบกำกับใบเดียวออก), 1 = ภาษีซื้อ, 2 = ภาษีขาย — ผลข้างเคียงเหมือนตารางภาษี (gl-journal-details.tsx) */
export function setTaxBoxType(details: GLJournalDetails | undefined, value: "" | "1" | "2", date: string): GLJournalDetails | undefined {
  const view = taxBoxView(details);
  if (view.mode === "multi") return details;
  if (value === "") return view.mode === "none" ? details : { ...details, vats: [] };
  const taxType = value === "1" ? 1 : 2;
  if (view.mode === "none") return { ...details, vats: [newVatRow(date, taxType)] };
  const row = view.row;
  const patch: Partial<GLDetailVat> = taxType === 1
    ? { tax_type: 1, claim_status: 1 }
    : { tax_type: 2, claim_status: undefined, claim_reason: undefined, ...(vatPeriodValue(row) ? {} : vatPeriodPatch(date.slice(0, 7))) };
  return { ...details, vats: [{ ...row, ...patch }] };
}

/** แก้ช่องในกล่องภาษี (แถวเดียว) — แก้ฐานหรืออัตราแล้วยอดภาษีกลับเป็น "คำนวณตอนบันทึก" เหมือนตารางภาษี */
export function patchTaxBox(details: GLJournalDetails | undefined, patch: Partial<GLDetailVat>): GLJournalDetails | undefined {
  const view = taxBoxView(details);
  if (view.mode !== "single") return details;
  const resetVat = ("base_amount" in patch || "vat_rate" in patch) && !("vat_amount" in patch);
  return { ...details, vats: [{ ...view.row, ...patch, ...(resetVat ? { vat_amount: undefined } : {}) }] };
}

/** ผลจาก backend GET /gl/v2/vat-calc */
export type VatCalcResult = { base_amount: string; vat_amount: string; total_amount: string; scale?: number };

/** เติมยอดลง 3 บรรทัดแรกแบบ Champ OnTaxAuto (glfrmjournal.cpp:1745-1817): ล้างเดบิต/เครดิตทุกบรรทัดก่อน ไม่เติมรหัสบัญชี
 *  ภาษีซื้อ: Dr ฐาน · Dr ภาษี · Cr รวม — ภาษีขาย: Dr รวม · Cr ฐาน · Cr ภาษี (ยอดทั้งสามมาจาก backend ไม่บวกลบในเบราว์เซอร์) */
export function taxLinesFromCalc(lines: GLLine[], taxType: 1 | 2, calc: VatCalcResult): GLLine[] {
  const next = lines.map((line) => ({ ...line, debit: "0", credit: "0" }));
  while (next.length < 3) next.push(emptyLine());
  const amounts: [keyof Pick<GLLine, "debit" | "credit">, string][] = taxType === 1
    ? [["debit", calc.base_amount], ["debit", calc.vat_amount], ["credit", calc.total_amount]]
    : [["debit", calc.total_amount], ["credit", calc.base_amount], ["credit", calc.vat_amount]];
  amounts.forEach(([side, amount], index) => { next[index] = { ...next[index], [side]: amount }; });
  return next;
}
/** มียอดเดบิต/เครดิตอยู่แล้วหรือไม่ — ถ้ามีต้องถามก่อนเติมทับ (Champ ทับเงียบ ๆ) */
export function linesHaveAmounts(lines: GLLine[]): boolean {
  return lines.some((line) => !isZeroText(line.debit) || !isZeroText(line.credit));
}

/** ข้อผิดพลาดภาษีมูลค่าเพิ่มจาก backend (code ขึ้นต้น vat_) ของใบสำคัญที่มีใบกำกับใบเดียว → ช่องของแถวนั้น
 *  (ช่องที่กรอบภาษีไม่มี เช่น ชื่อผู้ขาย/ผู้ซื้อ จะไปที่ตารางภาษี F3) — หลายใบระบุแถวไม่ได้ ไม่พาไปเดา */
export function vatErrorTarget(details: GLJournalDetails | undefined, code: string, field: string): { section: "vats"; row: number; field: string } | null {
  if (!code.startsWith("vat_") || !field || (details?.vats?.length ?? 0) !== 1) return null;
  return { section: "vats", row: 1, field };
}

/** ค่าที่ส่งให้ backend คำนวณภาษีจากกล่องภาษี — ตรวจแค่ "กรอกครบไหม" (อัตราห้ามเดา ไม่ใส่ 7% ให้เอง) ตัวเลขจริงคำนวณที่ backend
 *  ส่งปุ่มที่กด (button) มาด้วยแล้ว mode คือค่าที่ส่ง backend: จำนวนเงินว่างแต่กรอกภาษีไว้ → "from_tax" (backend หาฐาน = ภาษี × 100 ÷ อัตรา
 *  แบบ Champ OnTaxAuto glfrmjournal.cpp:1691-1701); กรอกทั้งคู่ → ใช้จำนวนเงินตามปุ่ม (Champ คำนวณภาษีทับจากจำนวนเงิน) */
export type VatCalcMode = "exclusive" | "inclusive" | "from_tax";
export type TaxBoxCalcInput = { ok: true; mode?: VatCalcMode; amount: string; rate: string; taxType: 1 | 2 } | { ok: false; problem: "no_row" | "rate" | "amount" };
export function taxBoxCalcInput(details: GLJournalDetails | undefined, button?: "exclusive" | "inclusive"): TaxBoxCalcInput {
  const view = taxBoxView(details);
  if (view.mode !== "single") return { ok: false, problem: "no_row" };
  const rate = String(view.row.vat_rate ?? "").trim();
  if (!rate) return { ok: false, problem: "rate" };
  const taxType = Number(view.row.tax_type) === 2 ? 2 : 1;
  const amount = String(view.row.base_amount ?? "").trim();
  if (!isZeroText(amount)) return { ok: true, mode: button, amount, rate, taxType };
  const tax = String(view.row.vat_amount ?? "").trim();
  if (!button || isZeroText(tax)) return { ok: false, problem: "amount" };
  return { ok: true, mode: "from_tax", amount: tax, rate, taxType };
}

/** ปุ่มฟังก์ชันของจอแบบ Champ: F2 เลือกรายการ (ช่องค้นหารายการ) · F3 ตารางภาษี · F5 รายละเอียดอื่นๆ
 *  เฉพาะกดปุ่มเดี่ยว (ไม่มี Ctrl/Alt/Shift/Meta) และยังไม่มีตัวจัดการอื่นรับไป — ผู้เรียก preventDefault เฉพาะเมื่อได้ค่ากลับ
 *  (F5 ปกติคือโหลดหน้าใหม่ จึงกันไว้เฉพาะตอนเปิดใบสำคัญอยู่); F9 ตารางเช็คไม่ผูก — BC ไม่มีระบบเช็คในสมุดรายวัน */
export type JournalFunctionKey = "list" | "tax" | "details";
export function journalFunctionKey(event: { key: string; ctrlKey?: boolean; metaKey?: boolean; altKey?: boolean; shiftKey?: boolean; defaultPrevented?: boolean }): JournalFunctionKey | null {
  if (event.defaultPrevented || event.ctrlKey || event.metaKey || event.altKey || event.shiftKey) return null;
  return event.key === "F2" ? "list" : event.key === "F3" ? "tax" : event.key === "F5" ? "details" : null;
}

// ---- รหัสมิติวิเคราะห์ของบรรทัดรายวัน: แผนก โครงการ การจัดสรร งาน หน่วยงาน ฝ่าย (ลำดับตามตารางรายวันของ Champ; DIMENSIONS ใน general-ledger.ts) ----
// สวิตช์ระดับบริษัทเปิดมิติ + ธง "แยก…" ของบัญชีเป็นผู้กำหนดว่าช่องไหนกรอกได้ — backend ตรวจซ้ำตอนสร้าง/แก้ใบ (สัญญา journal_line_<มิติ>_disabled / _not_separate)

type LineDimensionCodes = Partial<Record<GLDimensionLineField, string | undefined>>;

/** สี่มิติที่เพิ่มต่อจากแผนก/โครงการ: บรรทัด แม่แบบ และข้อมูลที่วางใส่ key ของมิติเหล่านี้เฉพาะเมื่อมีรหัส เพื่อให้ข้อมูลที่ไม่มีมิติเหล่านี้คงรูปเดิม */
export const EXTRA_DIMENSIONS: readonly GLDimensionDef[] = DIMENSIONS.filter((def) => def.line !== "departmentcode" && def.line !== "projectcode");

/** รหัสของสี่มิติเพิ่มเติมที่มีค่า (ตัดช่องว่างหัวท้าย) — กระจายใส่บรรทัดได้เลย ไม่มี key ที่เป็น "" */
export function extraDimensionCodes(source: LineDimensionCodes): Partial<Record<GLDimensionLineField, string>> {
  const codes: Partial<Record<GLDimensionLineField, string>> = {};
  for (const def of EXTRA_DIMENSIONS) {
    const code = (source[def.line] ?? "").trim();
    if (code) codes[def.line] = code;
  }
  return codes;
}

/** คอลัมน์มิติของตารางบรรทัด (ลำดับตาม DIMENSIONS): บริษัทเปิดใช้มิตินั้น หรือใบนี้มีรหัสของมิตินั้นอยู่แล้ว — สวิตช์ที่ปิดภายหลังไม่ซ่อนข้อมูลเดิม */
export function journalDimensionColumns(settings: Partial<GLDimensionSettingsInput> | undefined, lines: readonly LineDimensionCodes[] | undefined): GLDimensionDef[] {
  return visibleDimensions(settings, (def) => rowsHoldDimension(lines, def));
}

export type LineDimensionCell = {
  /** ช่องนี้แก้ไม่ได้ */
  locked: boolean;
  /** ข้อความ [key, ไทย] อธิบายว่าทำไมล็อก (title ของช่อง) */
  hint: GLLabel | null;
  /** ทะเบียนรหัสที่ช่องเลือกได้ และรหัสที่บัญชีของบรรทัดอนุญาต (null = เลือกทั้งหมด, [] = ไม่อนุญาตรหัสใด) */
  register: GLDimension[];
  allowed: string[] | null;
};
const NOT_SEPARATE_HINT: Record<GLDimensionFlagField, GLLabel> = {
  departseparate: ["gl_line_not_depart_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกแผนก (ตั้งที่ผังบัญชี › เงื่อนไข)"],
  projectseparate: ["gl_line_not_project_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกโครงการ (ตั้งที่ผังบัญชี › เงื่อนไข)"],
  allocateseparate: ["gl_line_not_allocate_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกการจัดสรร (ตั้งที่ผังบัญชี › เงื่อนไข)"],
  jobseparate: ["gl_line_not_job_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกงาน (ตั้งที่ผังบัญชี › เงื่อนไข)"],
  partseparate: ["gl_line_not_part_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกหน่วยงาน (ตั้งที่ผังบัญชี › เงื่อนไข)"],
  sideseparate: ["gl_line_not_side_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกฝ่าย (ตั้งที่ผังบัญชี › เงื่อนไข)"],
};
const VALUE_KEPT_HINT: GLLabel = ["gl_dim_value_kept", "บริษัทปิดการใช้งานมิตินี้แล้ว ค่าเดิมในใบสำคัญยังคงอยู่แต่แก้ไม่ได้"];
const SWITCH_OFF_HINT: GLLabel = ["gl_dim_switch_off_hint", "ยังไม่เปิดใช้ — เปิดได้ที่เมนู “กำหนดค่าเริ่มต้นบริษัท”"];

/** สถานะช่องมิติหนึ่งช่องของบรรทัด (Champ OnCheckSeparate, glfrmjournal.cpp:2812-3056):
 *  - บริษัทปิดมิติ: ช่องล็อก ทะเบียนไม่ได้โหลด แสดงรหัสเดิมของใบตามที่เก็บไว้ (แก้ไม่ได้)
 *  - บัญชีของบรรทัดไม่ได้ตั้ง "แยก…" และบรรทัดยังไม่มีรหัส: ช่องล็อก; ถ้ามีรหัสอยู่แล้วยังแก้/ล้างได้ */
export function lineDimensionCell(
  def: GLDimensionDef, settings: Partial<GLDimensionSettingsInput> | undefined, registers: Record<GLDimensionResource, GLDimension[]>,
  account: GLAccount | undefined, line: LineDimensionCodes,
): LineDimensionCell {
  const held = (line[def.line] ?? "").trim();
  if (!dimensionEnabled(settings, def)) {
    return { locked: true, hint: held ? VALUE_KEPT_HINT : SWITCH_OFF_HINT, register: held ? [{ code: held, name: "", isactive: true }] : [], allowed: null };
  }
  const register = registers[def.resource], allowed = accountAllowedCodes(account, def);
  if (!account?.[def.flag] && !held) return { locked: true, hint: NOT_SEPARATE_HINT[def.flag], register, allowed };
  return { locked: false, hint: null, register, allowed };
}

/** บรรทัดที่เอารหัสของมิติที่บริษัทปิดใช้ออก (วางจาก Excel / เติมจากแม่แบบ — ค่าใหม่ที่ backend ไม่รับ และช่องล็อกจนแก้ไม่ได้) + มิติที่ถูกเอาออก ตามลำดับ DIMENSIONS */
export function withoutDisabledDimensions<T extends LineDimensionCodes>(lines: readonly T[], settings: Partial<GLDimensionSettingsInput> | undefined): { lines: T[]; dropped: GLDimensionDef[] } {
  const dropped = DIMENSIONS.filter((def) => !dimensionEnabled(settings, def) && rowsHoldDimension(lines, def));
  if (dropped.length === 0) return { lines: [...lines], dropped };
  return {
    lines: lines.map((line) => {
      const clear = dropped.filter((def) => (line[def.line] ?? "").trim() !== "");
      return clear.length > 0 ? { ...line, ...Object.fromEntries(clear.map((def) => [def.line, ""])) } : line;
    }),
    dropped,
  };
}
