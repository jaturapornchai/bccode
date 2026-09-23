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

export type TaxFormResult<T> = { ok: true; data: T; notes?: TaxFormNote[] } | { ok: false; error: TaxFormError };

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
      individual: x.individual === true, hasattachment: x.hasattachment === true,
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
  return { ok: true, data: toDocument(r.payload.data) };
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
