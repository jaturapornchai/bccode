// แบบยื่นภาษีของกรมสรรพากร (ภ.ง.ด.3/53/2/2ก, ภ.พ.30/36, ภ.ธ.40, ภ.ง.ด.50/51/93/94)
// backend เป็นเจ้าของทุกอย่าง: สเปกช่องของแบบ, ค่าเริ่มต้นจากบัญชีแยกประเภท, สูตรยอดรวมที่พิมพ์บนแบบ, การบันทึก และ PDF
// จอนี้เก็บค่าที่ผู้ใช้กรอกเป็น string ตามที่พิมพ์ (ยอดเงินห้ามแปลงเป็น number — กฎ Accounting Number)

import { apiFetch } from "./client-auth-session";

const BASE = "/api/goapi/api/report/tax/form";

export type TaxFormPeriod = "month" | "year";

export interface TaxFormCatalogItem {
  code: string;
  title: string;
  period: TaxFormPeriod;
  source: "wht" | "vat" | "cit" | "manual";
  individual: boolean;
  hasattachment: boolean;
  // rdfile = backend สร้างไฟล์ยื่นกรมสรรพากร (Format กลาง V2.0) ของแบบนี้ได้ — ไม่มีในคำตอบ = ไม่รองรับ
  rdfile?: boolean;
}

export interface TaxFormOption {
  value: string;
  label: string;
}

export interface TaxFormField {
  key: string;
  type: "text" | "digits" | "taxid" | "money" | "int" | "check" | "choice";
  label: string;
  group?: string;
  note?: string;
  options?: TaxFormOption[];
}

export interface TaxFormAttachment {
  code: string;
  title: string;
  label: string;
  mode: "rows" | "sheets";
  rowspersheet?: number;
  columns: TaxFormField[];
}

export interface TaxFormSchema {
  code: string;
  title: string;
  fields: TaxFormField[];
  attachment?: TaxFormAttachment;
}

export type TaxFormValues = Record<string, string>;

export interface TaxFormDocument {
  values: TaxFormValues;
  rows?: TaxFormValues[];
  sheets?: TaxFormValues[];
}

export interface TaxFormNote {
  key: string;
  count?: number;
  amount?: string;
}

export interface TaxFiling {
  id: number;
  code: string;
  year: number;
  month: number;
  filingseq: number;
  version: number;
  updatedby: string;
  updatedat: string;
  document?: TaxFormDocument;
}

export interface TaxFormError {
  code: string;
  message?: string;
  field?: string;
  row?: number;
}

// rechecked = key หมายเหตุที่ backend ตรวจใหม่จากเอกสารปัจจุบัน: ผู้เรียกแทนหมายเหตุ key เหล่านั้นด้วย notes ชุดนี้
export type TaxFormResult<T> = { ok: true; data: T; notes?: TaxFormNote[]; rechecked?: string[] } | { ok: false; error: TaxFormError };

export interface TaxFormScope {
  holdingcode: string;
  businesscode: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

const text = (value: unknown) => (typeof value === "string" ? value : "");

// toValues - ค่าในเอกสารต้องเป็น string เท่านั้น (ค่าอื่นจาก backend ที่ผิดสัญญาถูกทิ้ง ไม่เดา)
export function toValues(value: unknown): TaxFormValues {
  const out: TaxFormValues = {};
  if (!isRecord(value)) return out;
  for (const [k, v] of Object.entries(value)) {
    if (typeof v === "string") out[k] = v;
  }
  return out;
}

export function toDocument(value: unknown): TaxFormDocument {
  const record = isRecord(value) ? value : {};
  const list = (v: unknown) => (Array.isArray(v) ? v.map(toValues) : undefined);
  return { values: toValues(record.values), rows: list(record.rows), sheets: list(record.sheets) };
}

function toError(payload: unknown, fallback: string): TaxFormError {
  if (!isRecord(payload)) return { code: fallback };
  return {
    code: text(payload.code) || fallback,
    message: text(payload.message) || undefined,
    field: text(payload.field) || undefined,
    row: typeof payload.row === "number" ? payload.row : undefined,
  };
}

async function post(path: string, body: unknown): Promise<{ ok: true; payload: Record<string, unknown> } | { ok: false; error: TaxFormError }> {
  const res = await apiFetch(`${BASE}/${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }).catch(() => null);
  if (res === null) return { ok: false, error: { code: "connection_error" } };
  if (res.status === 401 || res.status === 403) return { ok: false, error: { code: "unauthorized" } };
  const payload: unknown = await res.json().catch(() => null);
  if (!isRecord(payload) || payload.success !== true) return { ok: false, error: toError(payload, "load_failed") };
  return { ok: true, payload };
}

function toNotes(value: unknown): TaxFormNote[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((n) => ({
    key: text(n.key),
    count: typeof n.count === "number" ? n.count : undefined,
    amount: text(n.amount) || undefined,
  })).filter((n) => n.key);
}

function toFiling(value: unknown): TaxFiling {
  const r = isRecord(value) ? value : {};
  const num = (v: unknown) => (typeof v === "number" ? v : 0);
  return {
    id: num(r.id), code: text(r.code), year: num(r.year), month: num(r.month), filingseq: num(r.filingseq),
    version: num(r.version), updatedby: text(r.updatedby), updatedat: text(r.updatedat),
    document: r.document ? toDocument(r.document) : undefined,
  };
}

export async function fetchTaxFormCatalog(scope: TaxFormScope): Promise<TaxFormResult<TaxFormCatalogItem[]>> {
  const r = await post("catalog", scope);
  if (!r.ok) return r;
  const list = Array.isArray(r.payload.data) ? r.payload.data.filter(isRecord) : [];
  return {
    ok: true,
    data: list.map((x) => ({
      code: text(x.code), title: text(x.title), period: x.period === "year" ? "year" : "month",
      source: (["wht", "vat", "cit"].includes(text(x.source)) ? text(x.source) : "manual") as TaxFormCatalogItem["source"],
      individual: x.individual === true, hasattachment: x.hasattachment === true, rdfile: x.rdfile === true,
    })),
  };
}

export async function fetchTaxFormSchema(scope: TaxFormScope, code: string): Promise<TaxFormResult<TaxFormSchema>> {
  const r = await post("schema", { ...scope, code });
  if (!r.ok) return r;
  const data = r.payload.data;
  if (!isRecord(data) || !Array.isArray(data.fields)) return { ok: false, error: { code: "load_failed" } };
  return { ok: true, data: data as unknown as TaxFormSchema };
}

export interface TaxFormPeriodInput {
  code: string;
  year: number; // ค.ศ.
  month: number; // 0 = แบบรายปี
}

export async function prefillTaxForm(scope: TaxFormScope, period: TaxFormPeriodInput): Promise<TaxFormResult<TaxFormDocument>> {
  const r = await post("prefill", { ...scope, ...period });
  if (!r.ok) return r;
  return { ok: true, data: toDocument(r.payload.data), notes: toNotes(r.payload.notes) };
}

export async function computeTaxForm(scope: TaxFormScope, code: string, document: TaxFormDocument): Promise<TaxFormResult<TaxFormDocument>> {
  const r = await post("compute", { ...scope, code, document });
  if (!r.ok) return r;
  const rechecked = Array.isArray(r.payload.rechecked) ? r.payload.rechecked.filter((k): k is string => typeof k === "string") : [];
  return { ok: true, data: toDocument(r.payload.data), notes: toNotes(r.payload.notes), rechecked };
}

export async function saveTaxForm(
  scope: TaxFormScope,
  period: TaxFormPeriodInput,
  document: TaxFormDocument,
  current?: { id: number; version: number },
): Promise<TaxFormResult<TaxFiling>> {
  const r = await post("save", { ...scope, ...period, id: current?.id ?? 0, version: current?.version ?? 0, document });
  if (!r.ok) return r;
  return { ok: true, data: toFiling(r.payload.data) };
}

export async function listTaxForms(scope: TaxFormScope, code: string, year: number): Promise<TaxFormResult<TaxFiling[]>> {
  const r = await post("list", { ...scope, code, year });
  if (!r.ok) return r;
  return { ok: true, data: Array.isArray(r.payload.data) ? r.payload.data.map(toFiling) : [] };
}

export async function loadTaxForm(scope: TaxFormScope, id: number): Promise<TaxFormResult<TaxFiling>> {
  const r = await post("load", { ...scope, id });
  if (!r.ok) return r;
  return { ok: true, data: toFiling(r.payload.data) };
}

export async function deleteTaxForm(scope: TaxFormScope, id: number, version: number): Promise<TaxFormResult<null>> {
  const r = await post("delete", { ...scope, id, version });
  if (!r.ok) return r;
  return { ok: true, data: null };
}

// requestTaxFormPdf - backend ตรวจค่าและพิมพ์ลงแบบฟอร์มทางการ (รวมใบแนบที่แบ่งแผ่นเอง) — จอแค่แสดง
export async function requestTaxFormPdf(scope: TaxFormScope, period: TaxFormPeriodInput, document: TaxFormDocument): Promise<TaxFormResult<Blob>> {
  const res = await apiFetch(`${BASE}/pdf`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...scope, ...period, document }),
  }).catch(() => null);
  if (res === null) return { ok: false, error: { code: "connection_error" } };
  if ((res.headers.get("content-type") ?? "").startsWith("application/pdf")) return { ok: true, data: await res.blob() };
  if (res.status === 401 || res.status === 403) return { ok: false, error: { code: "unauthorized" } };
  return { ok: false, error: toError(await res.json().catch(() => null), "load_failed") };
}

// ---- ไฟล์ยื่นกรมสรรพากร (รูปแบบข้อมูล Format กลาง V2.0 สำหรับโปรแกรม SWC-UI) ----
// backend สร้างไฟล์จาก "ฉบับที่บันทึกแล้ว" (id + version) เท่านั้น ไม่ใช้ค่าที่ค้างในจอ — ตัวเลขจึงตรงกับฉบับที่ตรวจ/พิมพ์ PDF

export type TaxRdFileLto = "" | "0" | "1";
export type TaxRdFileBranchType = "" | "V" | "S";

export interface TaxRdFileOptions {
  dept_name: string;
  submission_no: string; // "00"–"99" (ส่วนท้ายของชื่อไฟล์)
  lto: TaxRdFileLto;
  branch_type: TaxRdFileBranchType;
}

/** จุดที่ต้องแก้ก่อนสร้างไฟล์: row 0 = ทั้งไฟล์/หัวแบบ, row n = แถวที่ n ของใบแนบ; args = ค่าที่แทน {char}/{max} ในข้อความ */
export interface TaxRdFileIssue {
  key: string;
  field?: string;
  row: number;
  message?: string;
  args?: TaxFormValues;
}

export interface TaxRdFileError extends TaxFormError {
  issues: TaxRdFileIssue[];
  total: number; // จำนวนจุดทั้งหมด (backend ส่งรายการมาไม่เกิน 100)
}

export type TaxRdFileResult = { ok: true; data: { blob: Blob; filename: string } } | { ok: false; error: TaxRdFileError };

// ค่าช่อง DEPT_NAME ของสำนักงานใหญ่ในไฟล์ — เป็นข้อมูลที่เขียนลงไฟล์ของกรม (ภาษาไทยเสมอ ไม่ขึ้นกับภาษาจอ)
// Format กลาง ภ.ง.ด.53 ข้อ 9: "ระบุชื่อแผนก/ส่วน/ฝ่ายที่นำส่ง หรือสำนักงานใหญ่ กรณีไม่แยกนำส่งเป็นแผนก"
export const RD_HEAD_OFFICE_DEPT_NAME = "สำนักงานใหญ่";
// ใช้เมื่อ backend ไม่ส่ง Content-Disposition มา (ไม่ควรเกิด — SWC-UI ตรวจชื่อไฟล์ ผู้ใช้ต้องใช้ชื่อที่ backend ตั้ง)
export const RD_FILE_FALLBACK_NAME = "rdfile.txt";

// isHeadOfficeBranch - เลขสาขาบนแบบเป็นศูนย์ล้วน (00000) = สำนักงานใหญ่
export function isHeadOfficeBranch(branch: string | undefined): boolean {
  return /^0+$/.test((branch ?? "").replace(/\s/g, ""));
}

// normalizeSubmissionNo - ครั้งที่ส่ง 2 หลัก: ว่าง = "00", หลักเดียวเติม 0 หน้า; ค่าอื่นคืน null (ให้ backend ตอบ tax_rdfile_submission_invalid)
export function normalizeSubmissionNo(value: string): string | null {
  const v = value.trim();
  if (v === "") return "00";
  if (/^\d$/.test(v)) return `0${v}`;
  return /^\d{2}$/.test(v) ? v : null;
}

// contentDispositionFilename - ชื่อไฟล์จาก header (filename*=UTF-8''… ก่อน แล้ว filename="…") ตัด path และอักขระควบคุมทิ้ง
export function contentDispositionFilename(header: string | null): string {
  if (!header) return "";
  let name = "";
  const extended = /filename\*\s*=\s*[\w-]+'[^']*'([^;]+)/i.exec(header);
  if (extended) {
    try {
      name = decodeURIComponent(extended[1].trim());
    } catch {
      name = "";
    }
  }
  if (!name) {
    const quoted = /filename\s*=\s*"((?:[^"\\]|\\.)*)"/i.exec(header);
    const bare = /filename\s*=\s*([^;"\s]+)/i.exec(header);
    name = quoted ? quoted[1].replace(/\\(.)/g, "$1") : bare?.[1] ?? "";
  }
  const base = name.split(/[\\/]/).pop() ?? "";
  return [...base].filter((ch) => ch >= " " && ch !== "\u007f").join("").trim();
}

function toRdFileIssues(value: unknown): TaxRdFileIssue[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((i) => ({
    key: text(i.key),
    field: text(i.field) || undefined,
    row: typeof i.row === "number" && Number.isInteger(i.row) && i.row > 0 ? i.row : 0,
    message: text(i.message) || undefined,
    args: Object.keys(toValues(i.args)).length > 0 ? toValues(i.args) : undefined,
  })).filter((i) => i.key);
}

function toRdFileError(payload: unknown): TaxRdFileError {
  const base = toError(payload, "load_failed");
  const record = isRecord(payload) ? payload : {};
  const issues = toRdFileIssues(record.issues);
  const total = typeof record.total === "number" && Number.isInteger(record.total) ? Math.max(record.total, issues.length) : issues.length;
  return { ...base, issues, total };
}

// requestTaxFormRdFile - สำเร็จ = ไฟล์ .txt (Blob ตามไบต์ที่ backend ส่ง: BOM + CRLF ต้องคงเดิม ห้ามอ่านเป็น text แล้วสร้างใหม่)
// ไม่ผ่าน = code + issues[] (BFF ส่ง 4xx ที่มี code มาเป็น 200 + success:false จึงอ่าน JSON ทุกกรณีที่ไม่ใช่ไฟล์)
export async function requestTaxFormRdFile(
  scope: TaxFormScope,
  filing: { id: number; version: number },
  options: TaxRdFileOptions,
): Promise<TaxRdFileResult> {
  const rdfile: TaxRdFileOptions = {
    dept_name: options.dept_name.trim(),
    submission_no: normalizeSubmissionNo(options.submission_no) ?? options.submission_no.trim(),
    lto: options.lto,
    branch_type: options.branch_type,
  };
  const res = await apiFetch(`${BASE}/rdfile`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...scope, id: filing.id, version: filing.version, rdfile }),
  }).catch(() => null);
  const failure = (code: string): TaxRdFileResult => ({ ok: false, error: { code, issues: [], total: 0 } });
  if (res === null) return failure("connection_error");
  if (res.ok && (res.headers.get("content-type") ?? "").startsWith("text/plain")) {
    const filename = contentDispositionFilename(res.headers.get("content-disposition")) || RD_FILE_FALLBACK_NAME;
    return { ok: true, data: { blob: await res.blob(), filename } };
  }
  if (res.status === 401 || res.status === 403) return failure("unauthorized");
  return { ok: false, error: toRdFileError(await res.json().catch(() => null)) };
}

// ---- ช่วยจอ (ไม่มีการคำนวณยอดเงิน) ----

// ช่องเงิน: ตัวเลข คอมมา และทศนิยมไม่เกิน 2 ตำแหน่ง (backend ตรวจซ้ำก่อนบันทึก/พิมพ์)
const MONEY_TYPING = /^[\d,]{0,20}(\.\d{0,2})?$/;
const DIGIT_TYPING = /^[\d\s-]{0,40}$/;
const INT_TYPING = /^\d{0,9}$/;

export function acceptsTyping(type: TaxFormField["type"], value: string): boolean {
  switch (type) {
    case "money":
      return MONEY_TYPING.test(value);
    case "digits":
    case "taxid":
      return DIGIT_TYPING.test(value);
    case "int":
      return INT_TYPING.test(value);
    default:
      return !/[\r\n\t]/.test(value) && value.length <= 300;
  }
}

// groupFields - หมวดของแบบตามลำดับที่ปรากฏบนกระดาษ
export function groupFields(fields: TaxFormField[]): { group: string; fields: TaxFormField[] }[] {
  const out: { group: string; fields: TaxFormField[] }[] = [];
  const index = new Map<string, number>();
  for (const f of fields) {
    const g = f.group || "other";
    let i = index.get(g);
    if (i === undefined) {
      i = out.length;
      index.set(g, i);
      out.push({ group: g, fields: [] });
    }
    out[i].fields.push(f);
  }
  return out;
}

// cleanDocument - ตัดช่องว่างทิ้ง (ไม่ส่งช่องว่างไปเก็บ) และตัดแถวใบแนบที่ว่างทั้งแถว
export function cleanDocument(doc: TaxFormDocument): TaxFormDocument {
  const trim = (values: TaxFormValues) => {
    const out: TaxFormValues = {};
    for (const [k, v] of Object.entries(values)) {
      const t = v.trim();
      if (t) out[k] = t;
    }
    return out;
  };
  const list = (rows?: TaxFormValues[]) => {
    const kept = (rows ?? []).map(trim).filter((r) => Object.keys(r).length > 0);
    return kept.length > 0 ? kept : undefined;
  };
  return { values: trim(doc.values), rows: list(doc.rows), sheets: list(doc.sheets) };
}

// noteText - ข้อความเตือนจาก backend: แทน {count} / {amount} ในข้อความของพจนานุกรม
export function noteText(template: string, note: TaxFormNote, formatMoney: (v: string) => string): string {
  return template.replace("{count}", String(note.count ?? 0)).replace("{amount}", note.amount ? formatMoney(note.amount) : "");
}

// normalizeMoneyText - ข้อความทศนิยมในรูปที่เทียบกันได้ (ตัดจุลภาค ศูนย์นำหน้า ศูนย์ท้ายทศนิยม; ว่าง = 0)
// เทียบเป็นข้อความ ไม่แปลงเป็น number (ตัวเลขบัญชีห้ามใช้ float — Rule 20)
export function normalizeMoneyText(value: string | undefined): string {
  let raw = (value ?? "").replace(/[,\s]/g, "");
  const negative = raw.startsWith("-");
  if (negative || raw.startsWith("+")) raw = raw.slice(1);
  const [whole = "", fraction = ""] = raw.split(".");
  const digits = whole.replace(/^0+/, "") || "0";
  const decimals = fraction.replace(/0+$/, "");
  if (digits === "0" && !decimals) return "0";
  return `${negative ? "-" : ""}${digits}${decimals ? `.${decimals}` : ""}`;
}

// ledgerBaseline - เอกสาร "ถ้าดึงยอดจากบัญชีตอนนี้" สำหรับเทียบกับฉบับที่บันทึก (ใช้เมื่อไม่มียอดจากบัญชีที่ฉบับนั้นตั้งต้น):
// ช่องเงินที่บัญชีเติมให้ (ค่าจาก prefill ไม่ว่าง) และใบแนบใช้ยอดจากบัญชี ส่วนช่องกรอกเอง (ภาษีชำระเกินยกมา, ประเภทการยื่น)
// และช่องที่ผู้ใช้แก้ในรอบนี้ (userEdited) ยกจากฉบับที่บันทึก — ต้องส่ง compute ก่อนเทียบ เพื่อให้บรรทัดรวมคิดจากค่าชุดเดียวกัน
// ไม่เตือนผิดว่า "ยอดบัญชีเปลี่ยน" เพียงเพราะผู้ใช้แก้ยอดที่ดึงมาเอง
export function ledgerBaseline(
  schema: TaxFormSchema,
  saved: TaxFormDocument,
  fresh: TaxFormDocument,
  userEdited: ReadonlySet<string> = new Set(),
): TaxFormDocument {
  const values: TaxFormValues = { ...saved.values };
  for (const field of schema.fields) {
    const ledger = fresh.values[field.key] ?? "";
    if (field.type === "money" && ledger.trim() && !userEdited.has(field.key)) values[field.key] = ledger;
  }
  return { values, rows: fresh.rows, sheets: fresh.sheets };
}

/** ช่องที่ค่าต่างกันระหว่างเอกสาร 2 ชุด: before = ค่าเดิม (ในจอ/ฉบับที่บันทึก/ยอดบัญชีตอนเตรียม), after = ค่าใหม่ */
export interface TaxFormDrift {
  key: string;
  before: string;
  after: string;
}

// ledgerDrift - ช่องเงินที่ before ต่างจาก after ตามลำดับช่องบนแบบ ใช้ 2 แบบ:
// (1) ยอดจากบัญชีที่ฉบับนี้ตั้งต้น (prefill ตอนเตรียม) เทียบกับ prefill ตอนนี้ — แม่นที่สุด ค่าที่ผู้ใช้แก้เองไม่ถูกนับ
// (2) ไม่มียอดตั้งต้น: ฉบับที่บันทึกเทียบกับ ledgerBaseline ที่ compute แล้ว (UAT V22 2026-09-24: ภ.พ.30 เปิดยอดเก่าโดยไม่เตือน)
export function ledgerDrift(schema: TaxFormSchema, before: TaxFormDocument, after: TaxFormDocument): TaxFormDrift[] {
  return schema.fields
    .filter((field) => field.type === "money")
    .map((field) => ({ key: field.key, before: before.values[field.key] ?? "", after: after.values[field.key] ?? "" }))
    .filter((d) => normalizeMoneyText(d.before) !== normalizeMoneyText(d.after));
}

// formChanges - ช่องบนแบบที่จะเปลี่ยนถ้าแทนค่าในจอ (current) ด้วยเอกสารใหม่ (next) — ใช้บอกผู้ใช้ก่อนยืนยันทับค่า
// ช่องเงินเทียบแบบข้อความทศนิยม (คอมมา/ศูนย์ท้ายไม่นับ) ช่องอื่นเทียบหลังตัดช่องว่างหัวท้าย
export function formChanges(schema: TaxFormSchema, current: TaxFormDocument, next: TaxFormDocument): TaxFormDrift[] {
  const same = (field: TaxFormField, a: string, b: string) =>
    field.type === "money" ? normalizeMoneyText(a) === normalizeMoneyText(b) : a.trim() === b.trim();
  return schema.fields
    .map((field) => ({ field, before: current.values[field.key] ?? "", after: next.values[field.key] ?? "" }))
    .filter(({ field, before, after }) => !same(field, before, after))
    .map(({ field, before, after }) => ({ key: field.key, before, after }));
}

// attachmentChanged - ใบแนบ (rows/sheets) จะถูกแทนด้วยชุดใหม่ที่ต่างจากในจอหรือไม่ (แถวว่างทั้งแถวไม่นับ)
export function attachmentChanged(current: TaxFormDocument, next: TaxFormDocument): boolean {
  // เรียง key ก่อนเทียบ: ลำดับ key ในแถวไม่ใช่ความต่างของข้อมูล
  const stable = (doc: TaxFormDocument) => {
    const clean = cleanDocument(doc);
    const list = (rows?: TaxFormValues[]) => (rows ?? []).map((row) => Object.keys(row).sort().map((key) => [key, row[key]]));
    return JSON.stringify([list(clean.rows), list(clean.sheets)]);
  };
  return stable(current) !== stable(next);
}

// เมนูแบบยื่น → แบบที่เปิดให้ทันที ("" = ให้เลือกเองจากรายการทั้งหมด); รหัสเมนู/route เดิมคงไว้ (ใช้เป็นรหัสสิทธิ์)
const TAX_FORM_ROUTES: Record<string, string> = {
  "/report/vatpp30": "pp30",
  "/report/vatpp36": "pp36",
  "/report/vatpnd2": "pnd2",
  "/report/vatpnd3": "pnd3",
  "/report/vatpnd53": "pnd53",
  "/report/taxforms": "",
};

export function taxFormRouteCode(route: string): string | undefined {
  return TAX_FORM_ROUTES[route.split("?")[0]];
}
