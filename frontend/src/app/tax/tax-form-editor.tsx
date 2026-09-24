"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AlertTriangle, Calculator, Download, FilePlus2, FileText, Loader2, RefreshCw, Save, Search, Trash2, X } from "lucide-react";
import { useBackendText } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { formatAmount } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import {
  attachmentChanged, cleanDocument, computeTaxForm, deleteTaxForm, fetchTaxFormCatalog, fetchTaxFormSchema, formChanges, ledgerBaseline, ledgerDrift,
  listTaxForms, loadTaxForm, noteText, prefillTaxForm, requestTaxFormPdf, requestTaxFormRdFile, saveTaxForm,
  type TaxFiling, type TaxFormCatalogItem, type TaxFormDocument, type TaxFormDrift, type TaxFormError, type TaxFormNote, type TaxFormSchema,
  type TaxRdFileIssue, type TaxRdFileOptions,
} from "@/lib/tax-forms";
import { focusTaxFormField, TaxFormFieldGroups, TaxFormRowsTable, TaxFormSheets, taxFormFieldId, type FieldInvalid } from "./tax-form-fields";
import { TaxRdFileButton, TaxRdFilePanel } from "./tax-rdfile-panel";

// แบบยื่นภาษีกรมสรรพากร: เลือกแบบ + งวด → เปิดฉบับที่บันทึกไว้ หรือให้ backend ดึงยอดจากบัญชีแยกประเภท →
// แก้ได้ทุกช่อง (รวมฐานภาษี) → backend คำนวณบรรทัดรวมตามสูตรบนแบบ → บันทึก → backend พิมพ์ลงแบบฟอร์มทางการ (PDF)

const MONTH_KEYS = ["january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"];
// แบบรายปีที่ยื่นระหว่างปีของงวดเอง (ครึ่งปี/ก่อนกำหนด) — ค่าเริ่มต้นเป็นปีปัจจุบัน ที่เหลือเป็นปีที่แล้ว
const CURRENT_YEAR_FORMS = new Set(["pnd51", "pnd94", "pnd93"]);

type Props = {
  language?: LanguageCode;
  holdingcode?: string;
  businesscode?: string;
  initialCode?: string;
};

export type Period = { year: number; month: number };
// ข้อความสถานะเก็บเป็น key + ข้อความสำรอง แล้วแปลตอนแสดง (สลับภาษาแล้วเปลี่ยนตาม)
type Status = { key: string; fallback: string; version?: number } | null;
// basis = เทียบยอดจากบัญชีตอนนี้กับยอดจากบัญชีที่ฉบับนี้ตั้งต้น (ดึงยอดในรอบนี้); saved = ไม่มียอดตั้งต้น (เปิดจากฉบับที่บันทึก)
// จึงเทียบกับค่าในฉบับที่บันทึก ยกเว้นช่องที่ผู้ใช้แก้ในรอบนี้ และต้องบอกผู้ใช้ว่าความต่างอาจมาจากการแก้เองครั้งก่อน
type Drift = { mode: "basis" | "saved"; rows: TaxFormDrift[] } | null;
// ยอดจากบัญชี (ผล prefill) ที่ค่าในจอตั้งต้นมา — backend ยังไม่เก็บค่านี้คู่กับฉบับที่บันทึก จึงรู้เฉพาะในรอบที่ดึงยอดเอง
type LedgerBasis = { code: string; year: number; month: number; doc: TaxFormDocument };

function defaultPeriod(item: TaxFormCatalogItem | undefined, now = new Date()): Period {
  if (item?.period === "year") return { year: now.getFullYear() - (CURRENT_YEAR_FORMS.has(item.code) ? 0 : 1), month: 0 };
  // เดือนก่อนหน้า (1-12) — แบบรายเดือนยื่นภายในเดือนถัดไป
  const month = now.getMonth();
  return month === 0 ? { year: now.getFullYear() - 1, month: 12 } : { year: now.getFullYear(), month };
}

/** งวดเมื่อเปลี่ยนแบบ: แบบรายเดือน → รายเดือน (หรือรายปี → รายปี) คงงวดที่ผู้ใช้เลือกไว้ — เดิมกลับไปเดือนก่อนหน้าทุกครั้ง
 *  เสี่ยงยื่นผิดงวด (UAT 2026-09-24: ตั้ง ต.ค. ใน ภ.ง.ด.53 แล้วเปลี่ยนเป็น ภ.ง.ด.3 เดือนกลายเป็น ส.ค.); ต่างชนิดงวดใช้ค่าเริ่มต้น */
export function periodForForm(item: TaxFormCatalogItem | undefined, previous: Period | null, now = new Date()): Period {
  const monthly = item?.period !== "year";
  if (previous && (monthly ? previous.month > 0 : previous.month === 0)) return previous;
  return defaultPeriod(item, now);
}

const emptyDocument = (): TaxFormDocument => ({ values: {} });
// จำนวนช่องที่ยอดต่างกันที่แสดงในคำเตือน (ที่เหลือบอกเป็นจำนวน) — ไม่ให้กล่องเตือนยาวจนดันแบบลงไป
const DRIFT_ROWS = 6;

export function TaxFormEditor({ language = "th", holdingcode = "", businesscode = "", initialCode = "" }: Props) {
  const tr = useBackendText();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultCancelLabel: tr("cancel", "ยกเลิก") });
  const scope = useMemo(() => ({ holdingcode, businesscode }), [holdingcode, businesscode]);

  const [catalog, setCatalog] = useState<TaxFormCatalogItem[]>([]);
  const [code, setCode] = useState(initialCode);
  const [schema, setSchema] = useState<TaxFormSchema | null>(null);
  const [period, setPeriod] = useState<Period | null>(null);
  // งวดล่าสุดที่เลือก — effect ที่โหลดแบบใหม่อ่านค่านี้โดยไม่ต้องขึ้นกับ period (ไม่งั้นเปลี่ยนเดือนแล้วโหลด schema ซ้ำ)
  const periodRef = useRef<Period | null>(null);
  useEffect(() => { periodRef.current = period; }, [period]);
  const [doc, setDoc] = useState<TaxFormDocument>(emptyDocument);
  const [notes, setNotes] = useState<TaxFormNote[]>([]);
  const [filing, setFiling] = useState<TaxFiling | null>(null);
  const [filings, setFilings] = useState<TaxFiling[]>([]);
  const [dirty, setDirty] = useState(false);
  const [busy, setBusy] = useState<string | null>("load");
  const [error, setError] = useState<TaxFormError | null>(null);
  const [status, setStatus] = useState<Status>(null);
  const [search, setSearch] = useState("");
  const [pdfUrl, setPdfUrl] = useState("");
  const [invalid, setInvalid] = useState<FieldInvalid>(null);
  const [drift, setDrift] = useState<Drift>(null);
  const [rdOpen, setRdOpen] = useState(false);
  const pdfUrlRef = useRef("");
  // ค่าล่าสุดสำหรับงานเบื้องหลัง (ตรวจยอดบัญชีเปลี่ยน) ที่เริ่มก่อน state รอบใหม่ render เสร็จ
  const schemaRef = useRef<TaxFormSchema | null>(null);
  const catalogRef = useRef<TaxFormCatalogItem[]>([]);
  const driftToken = useRef(0);
  const basisRef = useRef<LedgerBasis | null>(null);
  // ช่องที่ผู้ใช้แก้เองนับจากดึงยอด/เปิดฉบับล่าสุด — ไม่นับเป็น "ยอดบัญชีเปลี่ยน"
  const editedRef = useRef<Set<string>>(new Set());

  const item = catalog.find((c) => c.code === code);
  const money = (v: string) => formatAmount(v, 2);
  const monthName = (m: number) => tr(`month_${MONTH_KEYS[m - 1]}`, String(m));
  const yearText = (ce: number) => (language === "th" ? String(ce + 543) : String(ce));

  const showPdf = (url: string) => {
    if (pdfUrlRef.current) URL.revokeObjectURL(pdfUrlRef.current);
    pdfUrlRef.current = url;
    setPdfUrl(url);
  };
  useEffect(() => () => {
    if (pdfUrlRef.current) URL.revokeObjectURL(pdfUrlRef.current);
  }, []);

  // เตือนก่อนปิดแท็บเบราว์เซอร์เมื่อมีค่าที่ยังไม่บันทึก
  useEffect(() => {
    if (!dirty) return;
    const handler = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [dirty]);

  const fail = (e: TaxFormError) => {
    setError(e);
    setStatus(null);
  };

  // guard - ทำงานที่ทับค่าในจอ: ถามก่อนเมื่อมีค่าที่แก้แล้วยังไม่บันทึก
  const guard = useCallback(async () => {
    if (!dirty) return true;
    return confirm({
      title: tr("tax_form_discard_title", "ทิ้งค่าที่แก้ไว้?"),
      description: tr("tax_form_discard_desc", "ค่าที่แก้ในจอนี้ยังไม่ได้บันทึก ถ้าทำต่อค่าที่แก้จะหายไป"),
      confirmLabel: tr("tax_form_discard_confirm", "ทิ้งค่าที่แก้"),
      tone: "warning",
    });
  }, [confirm, dirty, tr]);

  useEffect(() => {
    if (!holdingcode || !businesscode) return;
    let alive = true;
    void fetchTaxFormCatalog(scope).then((r) => {
      if (!alive) return;
      if (!r.ok) {
        fail(r.error);
        setBusy(null);
        return;
      }
      catalogRef.current = r.data;
      setCatalog(r.data);
      setCode((current) => (r.data.some((c) => c.code === current) ? current : r.data[0]?.code ?? ""));
    });
    return () => {
      alive = false;
    };
  }, [scope, holdingcode, businesscode]);

  // clearDrift - ยกเลิกการตรวจที่ค้างอยู่ และซ่อนคำเตือน (ดึงยอดใหม่/เปลี่ยนงวด/เปลี่ยนแบบแล้ว คำเตือนเดิมไม่ตรงกับจอ)
  const clearDrift = useCallback(() => {
    driftToken.current += 1;
    setDrift(null);
  }, []);

  // checkDrift - หลังเปิด/บันทึกฉบับ: ดึงยอดจากบัญชี (อ่านอย่างเดียว ไม่แตะค่าในจอ) แล้วเทียบช่องเงิน; ต่างกัน = มีรายการบัญชี
  // เพิ่ม/แก้หลังเตรียมฉบับนี้ ต้องเตือนก่อนผู้ใช้ยื่นหรือพิมพ์ตัวเลขเก่า — ยอดที่ผู้ใช้แก้เองต้องไม่ถูกเตือนว่า "บัญชีเปลี่ยน"
  const checkDrift = useCallback(async (saved: TaxFiling) => {
    clearDrift();
    const token = driftToken.current;
    const formSchema = schemaRef.current;
    const source = catalogRef.current.find((c) => c.code === saved.code)?.source;
    if (!formSchema || formSchema.code !== saved.code || !saved.document || source === "manual") return;
    const fresh = await prefillTaxForm(scope, { code: saved.code, year: saved.year, month: saved.month });
    if (!fresh.ok || token !== driftToken.current) return;
    const basis = basisRef.current;
    if (basis && basis.code === saved.code && basis.year === saved.year && basis.month === saved.month) {
      const rows = ledgerDrift(formSchema, basis.doc, fresh.data);
      setDrift(rows.length > 0 ? { mode: "basis", rows } : null);
      return;
    }
    const ledger = await computeTaxForm(scope, saved.code, cleanDocument(ledgerBaseline(formSchema, saved.document, fresh.data, editedRef.current)));
    if (!ledger.ok || token !== driftToken.current) return;
    const rows = ledgerDrift(formSchema, saved.document, ledger.data);
    setDrift(rows.length > 0 ? { mode: "saved", rows } : null);
  }, [clearDrift, scope]);

  // applyPrefill - ใช้ผลดึงยอดจากบัญชีเป็นค่าในจอ และจำไว้เป็นยอดตั้งต้นของฉบับนี้ (ใช้เทียบว่าบัญชีเปลี่ยนหลังบันทึกหรือไม่)
  const applyPrefill = useCallback((p: Period, data: TaxFormDocument, prefillNotes: TaxFormNote[], keep: TaxFiling | null) => {
    basisRef.current = { code, year: p.year, month: p.month, doc: data };
    editedRef.current = new Set();
    setDoc(data);
    setNotes(prefillNotes);
    setDirty(keep !== null);
    setError(null);
    setStatus({ key: "tax_form_prefilled", fallback: "ดึงยอดจากบัญชีแล้ว — ตรวจทุกช่องก่อนยื่น" });
  }, [code]);

  const prefill = useCallback(async (p: Period, keep: TaxFiling | null) => {
    clearDrift();
    setBusy("prefill");
    const r = await prefillTaxForm(scope, { code, ...p });
    setBusy(null);
    if (!r.ok) return fail(r.error);
    applyPrefill(p, r.data, r.notes ?? [], keep);
  }, [applyPrefill, clearDrift, code, scope]);

  const openFiling = useCallback(async (id: number) => {
    setBusy("load");
    const r = await loadTaxForm(scope, id);
    setBusy(null);
    if (!r.ok) return fail(r.error);
    setFiling(r.data);
    setDoc(r.data.document ?? emptyDocument());
    basisRef.current = null;
    editedRef.current = new Set();
    setNotes([]);
    setDirty(false);
    setError(null);
    setStatus(null);
    void checkDrift(r.data);
  }, [checkDrift, scope]);

  // openPeriod - มีฉบับยื่นปกติของงวดนี้แล้ว → เปิดฉบับนั้น; ยังไม่มี → ดึงยอดจากบัญชี
  const openPeriod = useCallback(async (p: Period) => {
    setBusy("load");
    clearDrift();
    const r = await listTaxForms(scope, code, p.year);
    const list = r.ok ? r.data : [];
    setFilings(list);
    const saved = list.find((f) => f.month === p.month && f.filingseq === 0);
    setFiling(null);
    setInvalid(null);
    showPdf("");
    if (saved) return openFiling(saved.id);
    return prefill(p, null);
  }, [clearDrift, code, openFiling, prefill, scope]);

  useEffect(() => {
    if (!code || !item) return;
    let alive = true;
    setBusy("load");
    setSchema(null);
    void fetchTaxFormSchema(scope, code).then((r) => {
      if (!alive) return;
      if (!r.ok) {
        setBusy(null);
        return fail(r.error);
      }
      schemaRef.current = r.data;
      setSchema(r.data);
      const p = periodForForm(item, periodRef.current);
      setPeriod(p);
      void openPeriod(p);
    });
    return () => {
      alive = false;
    };
    // เปิดใหม่เมื่อเปลี่ยนแบบเท่านั้น (เปลี่ยนงวดเรียก openPeriod เอง)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [code, item?.code, scope]);


  const changeCode = async (next: string) => {
    if (next === code || !(await guard())) return;
    clearDrift();
    setDirty(false);
    setRdOpen(false);
    setCode(next);
  };

  const changePeriod = async (next: Period) => {
    if (!(await guard())) return;
    setPeriod(next);
    void openPeriod(next);
  };

  const edit = (next: TaxFormDocument) => {
    for (const key of new Set([...Object.keys(doc.values), ...Object.keys(next.values)])) {
      if ((doc.values[key] ?? "") !== (next.values[key] ?? "")) editedRef.current.add(key);
    }
    setDoc(next);
    setDirty(true);
    setStatus(null);
  };

  const reportError = (e: TaxFormError) => {
    fail(e);
    setInvalid(e.field ? { key: e.field, row: e.row ?? 0 } : null);
  };

  // runCompute - บรรทัดรวม/ยอดสุทธิตามสูตรบนแบบ (backend) ก่อนบันทึกและพิมพ์ทุกครั้ง ให้ตัวเลขบนกระดาษตรงกันเสมอ
  const runCompute = async (): Promise<TaxFormDocument | null> => {
    const r = await computeTaxForm(scope, code, cleanDocument(doc));
    if (!r.ok) {
      reportError(r.error);
      return null;
    }
    setDoc(r.data);
    const rechecked = r.rechecked ?? [];
    if (rechecked.length > 0) setNotes((prev) => [...prev.filter((n) => !rechecked.includes(n.key)), ...(r.notes ?? [])]);
    setInvalid(null);
    setError(null);
    return r.data;
  };

  const onCompute = async () => {
    setBusy("compute");
    const computed = await runCompute();
    setBusy(null);
    if (computed) {
      setDirty(true);
      setStatus({ key: "tax_form_computed", fallback: "คำนวณยอดรวมตามสูตรบนแบบแล้ว" });
    }
  };

  const onSave = async () => {
    if (!period) return;
    setBusy("save");
    const computed = await runCompute();
    if (!computed) return setBusy(null);
    const r = await saveTaxForm(scope, { code, ...period }, computed, filing ? { id: filing.id, version: filing.version } : undefined);
    setBusy(null);
    if (!r.ok) return reportError(r.error);
    setFiling(r.data);
    setDirty(false);
    setStatus({ key: "tax_form_saved", fallback: "บันทึกแล้ว (รุ่นที่ {version})", version: r.data.version });
    void checkDrift({ ...r.data, document: r.data.document ?? computed });
    const list = await listTaxForms(scope, code, period.year);
    if (list.ok) setFilings(list.data);
  };

  const onPdf = async () => {
    if (!period) return;
    setBusy("pdf");
    const computed = await runCompute();
    if (!computed) return setBusy(null);
    const r = await requestTaxFormPdf(scope, { code, ...period }, computed);
    setBusy(null);
    if (!r.ok) return reportError(r.error);
    showPdf(URL.createObjectURL(r.data));
    setStatus({ key: "tax_form_pdf_ready", fallback: "สร้างแบบสำหรับพิมพ์แล้ว — พิมพ์หรือดาวน์โหลดจากกรอบตัวอย่าง" });
  };

  // onPrefill - ดึงยอดจากบัญชีมาก่อน (ยังไม่แตะจอ) → มีช่องที่จะเปลี่ยน = ถามยืนยันเสมอพร้อมรายการช่อง ค่าเดิม → ค่าใหม่
  // (ค่าที่แก้เองหรือยอดในฉบับที่บันทึกจะถูกทับ) → ยืนยันแล้วจึงแทนค่าในจอ; ฉบับที่บันทึกไว้เปลี่ยนเมื่อกดบันทึกเท่านั้น
  const onPrefill = async () => {
    if (!period || !schema) return;
    setBusy("prefill");
    const r = await prefillTaxForm(scope, { code, ...period });
    setBusy(null);
    if (!r.ok) return fail(r.error);
    const changes = formChanges(schema, doc, r.data);
    const attachment = attachmentChanged(doc, r.data);
    if ((changes.length > 0 || attachment) && !(await confirmPrefill(changes, attachment))) return;
    clearDrift();
    applyPrefill(period, r.data, r.notes ?? [], filing);
  };

  const confirmPrefill = (changes: TaxFormDrift[], attachment: boolean) =>
    confirm({
      title: tr("tax_ui_prefill_confirm_title", "แทนค่าในจอด้วยยอดจากบัญชี?"),
      description: (
        <div className="grid gap-1.5 leading-[1.5]">
          <p>{tr("tax_form_prefill_confirm_desc", "ระบบจะดึงยอดจากบัญชีมาแทนค่าในจอทั้งหมด ค่าที่แก้เองจะหายไป")}</p>
          {changes.length > 0 ? (
            <>
              <p className="font-medium text-foreground">{tr("tax_ui_prefill_changes", "ช่องที่จะเปลี่ยน (ฉบับที่บันทึกไว้ยังไม่เปลี่ยนจนกว่าจะกดบันทึก):")}</p>
              <ul className="list-disc pl-5 text-foreground">
                {changes.slice(0, DRIFT_ROWS).map((c) => (
                  <li key={c.key} className="[overflow-wrap:anywhere]">
                    {tr("tax_ui_change_row", "{field}: {before} → {after}")
                      .replace("{field}", fieldLabel(c.key))
                      .replace("{before}", displayValue(c.key, c.before))
                      .replace("{after}", displayValue(c.key, c.after))}
                  </li>
                ))}
              </ul>
              {changes.length > DRIFT_ROWS ? <p>{moreText(tr("tax_ui_prefill_more", "ค่าในจอจะเปลี่ยนอีก {n} ช่อง (รวมทั้งหมด {total} ช่อง)"), changes.length)}</p> : null}
            </>
          ) : null}
          {attachment ? <p className="text-foreground">{tr("tax_ui_prefill_attachment", "ใบแนบทั้งหมดจะถูกแทนด้วยรายการจากบัญชี")}</p> : null}
        </div>
      ),
      confirmLabel: tr("tax_ui_prefill_confirm", "ใช้ยอดจากบัญชี"),
      tone: "warning",
    });

  const onDelete = async () => {
    if (!filing || !period) return;
    const ok = await confirm({
      title: tr("tax_form_delete_confirm_title", "ลบแบบที่บันทึกไว้?"),
      description: tr("tax_form_delete_confirm_desc", "ฉบับนี้จะหายจากรายการ (ประวัติทุกรุ่นยังเก็บไว้ตรวจสอบย้อนหลัง) ยอดในบัญชีไม่เปลี่ยน"),
      confirmLabel: tr("tax_form_delete", "ลบฉบับนี้"),
      tone: "danger",
    });
    if (!ok) return;
    setBusy("delete");
    const r = await deleteTaxForm(scope, filing.id, filing.version);
    setBusy(null);
    if (!r.ok) return reportError(r.error);
    clearDrift();
    setFiling(null);
    setDirty(true);
    setStatus({ key: "tax_form_deleted", fallback: "ลบฉบับที่บันทึกแล้ว — ค่าในจอยังอยู่ ยังไม่ได้บันทึก" });
    const list = await listTaxForms(scope, code, period.year);
    if (list.ok) setFilings(list.data);
  };

  // onAdditional - ยื่นเพิ่มเติม: ฉบับใหม่ของงวดเดียวกัน (ยกค่าในจอไป) ครั้งที่ถัดจากที่บันทึกไว้
  const onAdditional = () => {
    if (!period) return;
    const next = Math.max(0, ...filings.filter((f) => f.month === period.month).map((f) => f.filingseq)) + 1;
    setFiling(null);
    edit({ ...doc, values: { ...doc.values, filing_type: "additional", additional_no: String(next) } });
  };

  // onRdFile - ไฟล์ยื่นกรมสรรพากรสร้างจากฉบับที่บันทึก (id + version) — มีค่าที่แก้ค้างอยู่ต้องบันทึกก่อน
  const onRdFile = async (options: TaxRdFileOptions) => {
    if (!filing || dirty) return null;
    setBusy("rdfile");
    const r = await requestTaxFormRdFile(scope, { id: filing.id, version: filing.version }, options);
    setBusy(null);
    return r;
  };

  // gotoIssue - จุดที่ต้องแก้ก่อนสร้างไฟล์: ไฮไลต์ช่อง (หัวแบบ row 0 / แถวใบแนบ row n) แล้วพาไปช่องนั้นหลังจอ render
  const gotoIssue = (issue: TaxRdFileIssue) => {
    if (!schema || !issue.field) return;
    const field = issue.field;
    setInvalid({ key: field, row: issue.row });
    setSearch("");
    const prefix = issue.row > 0 && schema.attachment?.mode === "sheets" ? `sheet${issue.row - 1}` : `tf-${schema.code}`;
    const id = schema.attachment?.mode === "sheets" ? taxFormFieldId(prefix, field) : taxFormFieldId(prefix, field, issue.row);
    window.requestAnimationFrame(() => focusTaxFormField(id));
  };

  const fieldLabel = (key: string) =>
    schema?.fields.find((f) => f.key === key)?.label ?? schema?.attachment?.columns.find((c) => c.key === key)?.label ?? key;
  // ป้ายช่องของจุดที่ต้องแก้ในไฟล์: แถวใบแนบ (row > 0) ใช้ป้ายคอลัมน์ก่อน
  const issueFieldLabel = (key: string, row: number) =>
    (row > 0 ? schema?.attachment?.columns.find((c) => c.key === key)?.label : undefined) ?? fieldLabel(key);
  // กล่องข้อผิดพลาด: แถวใบแนบ (row > 0) ต้องได้ป้ายคอลัมน์ใบแนบ ไม่ใช่ป้ายช่องหัวแบบที่ key ซ้ำกัน (เช่น เลขผู้เสียภาษี);
  // ไม่รู้จักช่อง (ป้าย = รหัสดิบ) ไม่ต่อท้ายรหัสเทคนิคให้ผู้ใช้เห็น
  const errorFieldText = (e: TaxFormError) => {
    const label = e.field ? issueFieldLabel(e.field, e.row ?? 0) : "";
    return label && label !== e.field ? ` — ${tr("tax_form_field", "ช่อง")} “${label}”` : "";
  };
  // "และอีก n ช่อง" ของกล่องเตือนยอดบัญชีเปลี่ยน กับ dialog ดึงยอด นับคนละเรื่อง (ยอดบัญชีที่เปลี่ยน / ค่าในจอที่จะถูกแทน)
  // จึงใช้ถ้อยคำแยกกันและบอกจำนวนรวม ผู้ใช้จะไม่เห็นเลขสองเลขที่ถ้อยคำเหมือนกันแต่ไม่ตรงกัน (UAT 2026-09-24)
  const moreText = (template: string, total: number) =>
    template.replace("{n}", String(total - DRIFT_ROWS)).replace("{total}", String(total));

  // ค่าช่องสำหรับข้อความยืนยัน: เงินมีคอมมา, ตัวเลือกเป็นป้าย, ว่างบอกว่าว่าง (ไม่ใช่ช่องเปล่าที่อ่านแล้วงง)
  const displayValue = (key: string, value: string) => {
    if (!value.trim()) return tr("tax_ui_blank_value", "(ว่าง)");
    const field = schema?.fields.find((f) => f.key === key);
    if (field?.type === "money") return money(value);
    return field?.options?.find((o) => o.value === value)?.label ?? value;
  };

  const years = useMemo(() => {
    const now = new Date().getFullYear();
    return Array.from({ length: 7 }, (_, i) => now + 1 - i);
  }, []);

  const filingLabel = (f: TaxFiling) => {
    const when = f.month > 0 ? `${monthName(f.month)} ${yearText(f.year)}` : yearText(f.year);
    const kind = f.filingseq > 0 ? tr("tax_form_filing_additional", "ยื่นเพิ่มเติมครั้งที่ {n}").replace("{n}", String(f.filingseq)) : tr("tax_form_filing_normal", "ยื่นปกติ");
    return `${when} · ${kind} · v${f.version}`;
  };

  const working = busy !== null;
  const spin = (name: string) => (busy === name ? <Loader2 className="animate-spin" /> : null);

  return (
    <div className="flex flex-col gap-4 p-4 lg:p-6">
      {confirmationDialog}
      <section className="grid gap-3 rounded-2xl border border-border bg-card p-4 shadow-[var(--shadow-card)]">
        <div className="flex items-start gap-3">
          <span className="grid size-12 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
            <FileText className="size-6" />
          </span>
          <div className="min-w-0">
            <h1 className="text-xl font-bold leading-[1.45]">{tr("tax_form_screen_title", "แบบยื่นภาษีกรมสรรพากร")}</h1>
            <p className="text-sm leading-[1.5] text-muted-foreground">
              {tr("tax_form_screen_hint", "เลือกแบบและงวด ระบบดึงยอดจากบัญชีแยกประเภทให้ ทุกช่องแก้ได้ แล้วพิมพ์ลงแบบฟอร์มทางการของกรมสรรพากร")}
            </p>
          </div>
        </div>
        <div className="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(min(100%,15rem),1fr))]">
          <label className="grid gap-1 text-sm font-medium lg:col-span-2">
            {tr("tax_form_select", "แบบยื่น")}
            <ChoiceSelect
              value={code}
              onChange={(v) => void changeCode(String(v))}
              options={catalog.map((c) => ({ value: c.code, label: tr(`tax_form_title_${c.code}`, c.title) }))}
              radioThreshold={0}
              aria-label={tr("tax_form_select", "แบบยื่น")}
            />
          </label>
          {period ? (
            <label className="grid gap-1 text-sm font-medium">
              {item?.source === "cit" ? tr("tax_form_year_fiscal", "ปีที่สิ้นสุดรอบบัญชี") : tr("tax_form_year", "ปีภาษี")}
              <ChoiceSelect
                value={period.year}
                onChange={(v) => void changePeriod({ ...period, year: Number(v) })}
                options={years.map((y) => ({ value: y, label: yearText(y) }))}
                radioThreshold={0}
                aria-label={tr("tax_form_year", "ปีภาษี")}
              />
            </label>
          ) : null}
          {period && item?.period === "month" ? (
            <label className="grid gap-1 text-sm font-medium">
              {tr("tax_form_month", "เดือนภาษี")}
              <ChoiceSelect
                value={period.month}
                onChange={(v) => void changePeriod({ ...period, month: Number(v) })}
                options={MONTH_KEYS.map((_, i) => ({ value: i + 1, label: monthName(i + 1) }))}
                radioThreshold={0}
                aria-label={tr("tax_form_month", "เดือนภาษี")}
              />
            </label>
          ) : null}
          <label className="grid gap-1 text-sm font-medium">
            {tr("tax_form_saved_list", "ฉบับที่บันทึกไว้")}
            <ChoiceSelect
              value={filing?.id ?? 0}
              onChange={async (v) => {
                if (Number(v) > 0 && Number(v) !== filing?.id && (await guard())) void openFiling(Number(v));
              }}
              options={[
                { value: 0, label: filing ? tr("tax_form_unsaved", "ยังไม่ได้บันทึก") : filings.length ? tr("tax_form_unsaved", "ยังไม่ได้บันทึก") : tr("tax_form_empty_filings", "ยังไม่มีฉบับที่บันทึก") },
                ...filings.map((f) => ({ value: f.id, label: filingLabel(f) })),
              ]}
              radioThreshold={0}
              aria-label={tr("tax_form_saved_list", "ฉบับที่บันทึกไว้")}
            />
          </label>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button type="button" className="h-11 gap-2 px-5" onClick={() => void onSave()} disabled={working || !schema}>
            {spin("save") ?? <Save />}
            {tr("tax_form_save", "บันทึก")}
          </Button>
          <Button type="button" variant="outline" className="h-11 gap-2" onClick={() => void onPdf()} disabled={working || !schema}>
            {spin("pdf") ?? <Download />}
            {tr("tax_form_pdf", "พิมพ์แบบ (PDF)")}
          </Button>
          {item?.rdfile ? (
            <TaxRdFileButton ready={filing !== null && !dirty} working={working} busy={busy === "rdfile"} open={rdOpen} onToggle={() => setRdOpen((open) => !open)} />
          ) : null}
          <Button type="button" variant="outline" className="h-11 gap-2" onClick={() => void onCompute()} disabled={working || !schema}>
            {spin("compute") ?? <Calculator />}
            {tr("tax_form_compute", "คำนวณยอดรวม")}
          </Button>
          <Button type="button" variant="outline" className="h-11 gap-2" onClick={() => void onPrefill()} disabled={working || !schema}>
            {spin("prefill") ?? <RefreshCw />}
            {tr("tax_form_prefill", "ดึงยอดจากบัญชี")}
          </Button>
          {filing ? (
            <>
              <Button type="button" variant="outline" className="h-11 gap-2" onClick={onAdditional} disabled={working}>
                <FilePlus2 />
                {tr("tax_form_additional", "ยื่นเพิ่มเติม (ฉบับใหม่)")}
              </Button>
              <Button type="button" variant="outline" className="h-11 gap-2 text-destructive" onClick={() => void onDelete()} disabled={working}>
                {spin("delete") ?? <Trash2 />}
                {tr("tax_form_delete", "ลบฉบับนี้")}
              </Button>
            </>
          ) : null}
          {dirty ? <span className="text-sm font-medium text-primary">● {tr("tax_form_unsaved", "ยังไม่ได้บันทึก")}</span> : null}
          {busy === "load" ? <Loader2 className="size-5 animate-spin text-muted-foreground" aria-label={tr("tax_form_loading", "กำลังโหลด")} /> : null}
        </div>
        {item?.rdfile && rdOpen ? (
          <TaxRdFilePanel
            filing={filing}
            ready={filing !== null && !dirty}
            working={working}
            branchNo={doc.values.branch_no}
            fieldLabel={issueFieldLabel}
            onCreate={onRdFile}
            onIssue={gotoIssue}
            onClose={() => setRdOpen(false)}
          />
        ) : null}
        {error ? (
          <div role="alert" className="rounded-xl border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm leading-[1.5] text-destructive">
            {tr(error.code, error.message ?? tr("tax_form_failed", "ทำรายการไม่สำเร็จ"))}
            {errorFieldText(error)}
            {error.row ? ` (${tr("tax_form_row_label", "แถวที่ {n}").replace("{n}", String(error.row))})` : ""}
          </div>
        ) : null}
        {status ? (
          <p role="status" className="text-sm font-medium text-primary">
            {tr(status.key, status.fallback).replace("{version}", String(status.version ?? ""))}
          </p>
        ) : null}
        {drift ? (
          <div role="alert" className="grid gap-2 rounded-xl border border-primary/40 bg-primary/10 px-3 py-2 text-sm leading-[1.5] text-foreground">
            <p className="flex items-center gap-2 font-semibold">
              <AlertTriangle className="size-5 shrink-0 text-primary" aria-hidden />
              {tr("tax_form_drift_title", "ยอดในบัญชีตอนนี้ไม่ตรงกับฉบับที่บันทึก")}
            </p>
            <p>{tr("tax_form_drift_desc", "มีรายการบัญชีเพิ่มหรือแก้ไขหลังบันทึกฉบับนี้ ตรวจก่อนยื่นหรือพิมพ์ — กด “ดึงยอดจากบัญชี” เพื่อใช้ยอดล่าสุด")}</p>
            {drift.mode === "saved" ? (
              <p>{tr("tax_ui_drift_no_basis", "ฉบับนี้ไม่มียอดจากบัญชี ณ ตอนเตรียมให้เทียบ ระบบจึงเทียบกับค่าในฉบับที่บันทึก — ถ้าเคยแก้ยอดเองในช่องใด ช่องนั้นจะขึ้นว่าต่างด้วย (ช่องที่แก้ในรอบนี้ไม่นับ)")}</p>
            ) : null}
            <ul className="grid gap-0.5">
              {drift.rows.slice(0, DRIFT_ROWS).map((d) => (
                <li key={d.key}>
                  {(drift.mode === "basis"
                    ? tr("tax_ui_drift_basis_row", "{field}: ยอดจากบัญชีตอนเตรียมฉบับนี้ {saved} · ยอดจากบัญชีตอนนี้ {ledger}")
                    : tr("tax_form_drift_row", "{field}: ฉบับที่บันทึก {saved} · ยอดจากบัญชีตอนนี้ {ledger}"))
                    .replace("{field}", fieldLabel(d.key))
                    .replace("{saved}", money(d.before) || "0.00")
                    .replace("{ledger}", money(d.after) || "0.00")}
                </li>
              ))}
            </ul>
            {drift.rows.length > DRIFT_ROWS ? <p>{moreText(tr("tax_ui_drift_more", "ยอดจากบัญชีต่างกันอีก {n} ช่อง (รวมทั้งหมด {total} ช่อง)"), drift.rows.length)}</p> : null}
            <div>
              <Button type="button" variant="outline" className="h-11 gap-2" onClick={() => void onPrefill()} disabled={working}>
                {spin("prefill") ?? <RefreshCw />}
                {tr("tax_form_prefill", "ดึงยอดจากบัญชี")}
              </Button>
            </div>
          </div>
        ) : null}
        {notes.length > 0 ? (
          <ul className="grid gap-1 rounded-xl border border-primary/30 bg-primary/10 px-3 py-2 text-sm leading-[1.5] text-foreground">
            {notes.map((n) => (
              <li key={n.key}>⚠ {noteText(tr(n.key, n.key), n, money)}</li>
            ))}
          </ul>
        ) : null}
        {item?.individual ? (
          <p className="text-sm leading-[1.5] text-muted-foreground">{tr("tax_form_individual_hint", "แบบของบุคคลธรรมดา — กรอกชื่อและเลขประจำตัวผู้เสียภาษีของผู้ยื่นเอง (ไม่ใช่ของบริษัท)")}</p>
        ) : null}
      </section>

      {schema ? (
        <div className={pdfUrl ? "grid items-start gap-4 xl:grid-cols-2" : "grid gap-4"}>
          <div className="grid min-w-0 gap-4">
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={tr("tax_form_search", "ค้นหาช่องในแบบ")}
                aria-label={tr("tax_form_search", "ค้นหาช่องในแบบ")}
                className="min-h-[2.6em] !pl-10"
              />
            </div>
            <TaxFormFieldGroups
              fields={schema.fields}
              values={doc.values}
              onChange={(key, value) => edit({ ...doc, values: { ...doc.values, [key]: value } })}
              invalid={invalid}
              search={search}
              idPrefix={`tf-${schema.code}`}
            />
            {schema.attachment ? (
              <section className="grid gap-3 rounded-2xl border border-border bg-card p-4 shadow-[var(--shadow-card)]">
                <h2 className="text-base font-semibold leading-[1.45]">{schema.attachment.title}</h2>
                {schema.attachment.mode === "rows" ? (
                  <TaxFormRowsTable attachment={schema.attachment} rows={doc.rows ?? []} onChange={(rows) => edit({ ...doc, rows })} invalid={invalid} idPrefix={`tf-${schema.code}`} />
                ) : (
                  <TaxFormSheets attachment={schema.attachment} sheets={doc.sheets ?? []} onChange={(sheets) => edit({ ...doc, sheets })} invalid={invalid} />
                )}
              </section>
            ) : null}
          </div>
          {pdfUrl ? (
            // จอ xl: คอลัมน์ขวาแบบ sticky ดูคู่กับช่องกรอก; จอเล็กกว่า: overlay กลางจอ (ไม่ไปอยู่ท้ายหน้าที่มองไม่เห็น)
            <div className="fixed inset-0 z-50 grid place-items-center bg-black/45 p-3 xl:sticky xl:inset-auto xl:top-4 xl:z-auto xl:block xl:bg-transparent xl:p-0">
              <section
                role="dialog"
                aria-label={tr("tax_form_preview_title", "ตัวอย่างแบบที่จะพิมพ์")}
                className="grid w-full max-w-4xl gap-2 rounded-2xl border border-border bg-card p-3 shadow-[var(--shadow-card)] xl:max-w-none"
                onKeyDown={(e) => {
                  if (e.key === "Escape") showPdf("");
                }}
              >
                <header className="flex items-center justify-between gap-2">
                  <h2 className="text-base font-semibold">{tr("tax_form_preview_title", "ตัวอย่างแบบที่จะพิมพ์")}</h2>
                  <Button type="button" variant="outline" size="sm" className="gap-1" onClick={() => showPdf("")} autoFocus>
                    <X />
                    {tr("tax_form_close_preview", "ปิดตัวอย่าง")}
                  </Button>
                </header>
                <iframe src={pdfUrl} title={tr("tax_form_preview_title", "ตัวอย่างแบบที่จะพิมพ์")} className="h-[calc(100dvh-7rem)] w-full rounded-xl border border-border bg-white xl:h-[80dvh]" />
              </section>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
