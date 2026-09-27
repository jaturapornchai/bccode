"use client";
// พิมพ์ชุดงบการเงิน (ลุงจืดอนุมัติ 2026-09-27 — Champ ไม่มี, ไม่มีเมนูใหม่: เปิดจากจอออกแบบงบการเงิน)
// งบที่เลือก + หมายเหตุประกอบงบการเงิน ต่อเนื่องในงานพิมพ์เดียว เรียงตามแบบ 2 ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566
// ตัวเลขทุกตัวคำนวณที่ backend ในคำขอเดียว (GET reports/statement-set — backend/internal/generalledger/statement_set.go):
// ทุกงบคำนวณแบบเดียวกับ reports/statement + หมายเหตุของปี ใน snapshot เดียว และเรียงตามแบบ 2 มาแล้ว — จอนี้แค่ส่งรายการที่เลือก
// แสดงผลตรวจ แล้วพิมพ์ตามลำดับที่ได้ (รายการให้เลือกเรียงด้วย statementSetTemplates ใน lib/general-ledger.ts ซึ่ง test ตรวจว่าตรงกับ backend)
import { useEffect, useId, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { AlertTriangle, CheckCircle2, ClipboardCheck, Loader2, Printer, X, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { formatAppDate } from "@/lib/date-time";
import { glAllRecords, glRequest } from "@/lib/general-ledger-api";
import { labelText, statementNotesFromRecord, statementSetDefaultSelection, statementSetTemplates, statementTypeLabels, type GLReport, type GLStatementTemplate, type StatementNote } from "@/lib/general-ledger";
import { YearSelect, actionClass, useGLLanguage, useGLText, useReferences, type GLTextFn } from "./gl-common";
import { GLReportWarnings, GLStatementChecks, GLStatementSetPrint, GLStatementUnassigned, printCompanyName, statementPeriodLine, statementPeriodText, statementSetRootOrientation, useGLPrint, type GLStatementSetSection } from "./gl-print";
import { loadStatementNotes } from "./gl-statement-notes";

/** ผลของ GET reports/statement-set: งบเรียงตามแบบ 2 แล้ว; งบที่คำนวณไม่สำเร็จมี error (ข้อความตามภาษาที่เลือก) แทน report; ยอดเงินในรายงานเป็นสตริงทศนิยม */
export type GLStatementSetSectionResult = { code: string; name: string; statementtype: string; shownotecolumn: boolean; scale: number; report?: GLReport | null; error?: string };
export type GLStatementSetResult = { sequence: number; fiscalyear: string; from: string; to: string; sections: GLStatementSetSectionResult[] | null; notes: StatementNote[] | null };
type PreparedSection = { code: string; title: string; statementtype: string; showNote: boolean; scale: number; report: GLReport | null; error: string };
/** error = ทั้งคำขอไม่สำเร็จ (เช่น รูปแบบงบถูกลบ/ปิดใช้ระหว่างเปิดจอ หรือเชื่อมต่อไม่ได้) — ไม่มีงบใดพิมพ์ได้ */
type PreparedSet = { year: string; sections: PreparedSection[]; notes: { items: StatementNote[]; error: string } | null; error: string };
type NotesInfo = { loading: boolean; count: number; error: string };

/** คำขอชุดงบของทั้งปีบัญชี: templates ส่งเสมอ (ว่าง = หมายเหตุอย่างเดียว; backend ถือว่าไม่ส่ง = ชุดเริ่มต้น) */
export function statementSetPath(year: { code: string; startdate: string; enddate: string }, codes: string[], withNotes: boolean): string {
  return `reports/statement-set?${new URLSearchParams({ fiscalyear: year.code, from: year.startdate, to: year.enddate, templates: codes.join(","), notes: String(withNotes) })}`;
}

/** ผลจาก backend → รายการตรวจก่อนพิมพ์ ตามลำดับที่ backend ส่งมา (ไม่เรียงใหม่ที่ browser) */
export function preparedStatementSet(result: GLStatementSetResult, year: string, withNotes: boolean, tr: GLTextFn): PreparedSet {
  const sections = (result.sections ?? []).map((section): PreparedSection => ({
    code: section.code,
    title: section.name || section.code,
    statementtype: section.statementtype,
    showNote: section.shownotecolumn ?? true,
    scale: section.scale ?? 2,
    report: section.report ?? null,
    error: section.report ? "" : section.error?.trim() || tr("gl_statement_set_failed", "คำนวณไม่สำเร็จ"),
  }));
  const items = withNotes ? statementNotesFromRecord({ notes: result.notes }) : [];
  return {
    year: result.fiscalyear || year,
    sections,
    notes: withNotes ? { items, error: items.length ? "" : tr("gl_statement_set_notes_none", "ปีบัญชีนี้ยังไม่มีหมายเหตุประกอบงบการเงิน — เขียนได้ที่โหมด “หมายเหตุประกอบงบการเงิน”") } : null,
    error: "",
  };
}

const rowClass = "flex min-h-12 cursor-pointer items-center gap-3 rounded-xl border border-border bg-background px-3 py-2 text-[0.95rem] leading-snug shadow-[0_2px_8px_rgba(0,0,0,0.06)] transition-colors hover:border-primary/60 has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-ring has-[:disabled]:cursor-default has-[:disabled]:opacity-70";

// ผลคำนวณที่มีคำเตือนจาก backend ผลตรวจยอดกับบัญชีที่ไม่ตรง หรือบัญชีที่มียอดแต่ไม่อยู่ในบรรทัดใด: พิมพ์ได้ แต่ต้องบอกให้ตรวจก่อนออกงบ
export function reportNeedsReview(report: GLReport) {
  return (report.warnings ?? []).some((warning) => warning?.trim()) || (report.checks ?? []).some((check) => !check.matched) || (report.unassigned?.length ?? 0) > 0;
}

export function GLStatementSetDialog({ open, onClose, unsavedChanges }: { open: boolean; onClose: () => void; unsavedChanges: boolean }) {
  const tr = useGLText();
  const language = useGLLanguage();
  const refs = useReferences();
  const print = useGLPrint();
  const titleId = useId();
  const descriptionId = useId();
  const dialogRef = useRef<HTMLDivElement>(null);
  const printButtonRef = useRef<HTMLButtonElement>(null);
  const resultsHeadingRef = useRef<HTMLHeadingElement>(null);
  // โฟกัสที่พักไว้ที่ตัว dialog ระหว่างคำนวณ/พิมพ์ — ทำเสร็จแล้วต้องส่งต่อไปยังปุ่มถัดไป
  const focusParkedRef = useRef(false);
  // เลขคำขอ: เปลี่ยนปี/รายการ/ปิดจอระหว่างคำนวณ = ผลที่ตอบกลับทีหลังถูกทิ้ง ไม่ทับรายการใหม่
  const requestRef = useRef(0);
  const [year, setYear] = useState("");
  const [templates, setTemplates] = useState<GLStatementTemplate[]>([]);
  const [templatesLoading, setTemplatesLoading] = useState(false);
  const [templatesError, setTemplatesError] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [includeNotes, setIncludeNotes] = useState(false);
  const [notesInfo, setNotesInfo] = useState<NotesInfo>({ loading: false, count: 0, error: "" });
  const [preparing, setPreparing] = useState(false);
  const [prepared, setPrepared] = useState<PreparedSet | null>(null);

  const fiscalYear = useMemo(() => refs.years.find((item) => item.code === year), [refs.years, year]);
  const chosen = templates.filter((template) => selected.includes(template.code));
  const nothingSelected = chosen.length === 0 && !includeNotes;

  function invalidate() {
    requestRef.current += 1;
    setPrepared(null);
    setPreparing(false);
  }

  // ปิดจอ = ทิ้งผลที่เตรียมไว้และคำขอที่ค้าง; เปิดใหม่ต้องตรวจใหม่จากข้อมูลล่าสุด
  useEffect(() => {
    if (open) return;
    requestRef.current += 1;
    setPrepared(null);
    setPreparing(false);
  }, [open]);

  // ปีตั้งต้น = ปีบัญชีที่เปิดใช้และยังไม่ปิด (กติกาเดียวกับพรีวิวงบและหมายเหตุ)
  useEffect(() => {
    if (year || !refs.years.length) return;
    const active = refs.years.find((item) => item.isactive && !item.closed) ?? refs.years[0];
    if (active) setYear(active.code);
  }, [refs.years, year]);

  // รูปแบบงบที่บันทึกแล้วทุกหน้า (glAllRecords ไล่หน้าจนครบ total — backend จำกัด 1000 รายการต่อหน้า)
  useEffect(() => {
    if (!open) return;
    let active = true;
    setTemplatesLoading(true);
    setTemplatesError("");
    glAllRecords<GLStatementTemplate>("statement-templates", "", 10000)
      .then((items) => {
        if (!active) return;
        const sorted = statementSetTemplates(items);
        setTemplates(sorted);
        setSelected(statementSetDefaultSelection(sorted));
      })
      .catch((e: Error) => { if (active) { setTemplates([]); setSelected([]); setTemplatesError(e.message); } })
      .finally(() => { if (active) setTemplatesLoading(false); });
    return () => { active = false; };
  }, [open]);

  // จำนวนหมายเหตุของปีที่เลือก: มีหมายเหตุ = เลือกให้เป็นค่าเริ่มต้น
  useEffect(() => {
    if (!open || !year) return;
    let active = true;
    setNotesInfo({ loading: true, count: 0, error: "" });
    setIncludeNotes(false);
    loadStatementNotes(year)
      .then((result) => {
        if (!active) return;
        const count = result?.notes.length ?? 0;
        setNotesInfo({ loading: false, count, error: "" });
        setIncludeNotes(count > 0);
      })
      .catch((e: Error) => { if (active) setNotesInfo({ loading: false, count: 0, error: e.message }); });
    return () => { active = false; };
  }, [open, year]);

  // เปิดแล้วโฟกัสที่ตัว dialog; ปิดแล้วคืนโฟกัสให้ปุ่มที่เปิด
  useEffect(() => {
    if (!open) return;
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    dialogRef.current?.focus();
    return () => { opener?.focus(); };
  }, [open]);

  // ปุ่มตรวจ/พิมพ์ disabled ตัวเองตอนถูกกด → เบราว์เซอร์ย้ายโฟกัสไป <body>: Escape ไม่ถึง dialog และ Tab หลุดไปจอข้างหลัง
  // ระหว่างทำงาน = พักโฟกัสไว้ที่ตัว dialog; เสร็จแล้ว = ปุ่มพิมพ์ (ถ้ากดได้) หรือหัวข้อผลการตรวจ
  const busy = preparing || print.printing;
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!open || !dialog) { focusParkedRef.current = false; return; }
    const active = document.activeElement;
    const lost = !(active instanceof HTMLElement) || !dialog.contains(active) || active.matches(":disabled");
    if (busy) {
      if (lost) { focusParkedRef.current = true; dialog.focus(); }
      return;
    }
    if (!lost && !(focusParkedRef.current && active === dialog)) return;
    focusParkedRef.current = false;
    const printButton = printButtonRef.current;
    (printButton && !printButton.disabled ? printButton : resultsHeadingRef.current ?? dialog).focus();
  }, [open, busy]);

  function changeYear(value: string) {
    invalidate();
    setYear(value);
  }
  function toggleTemplate(code: string, checked: boolean) {
    invalidate();
    setSelected((current) => (checked ? [...current.filter((item) => item !== code), code] : current.filter((item) => item !== code)));
  }
  function toggleNotes(checked: boolean) {
    invalidate();
    setIncludeNotes(checked);
  }

  async function prepare() {
    if (!fiscalYear || nothingSelected) return;
    const request = ++requestRef.current;
    const withNotes = includeNotes;
    const year = fiscalYear.code;
    setPrepared(null);
    setPreparing(true);
    let next: PreparedSet;
    try {
      const result = await glRequest<GLStatementSetResult>(statementSetPath(fiscalYear, chosen.map((template) => template.code), withNotes));
      next = preparedStatementSet(result, year, withNotes, tr);
    } catch (e) {
      next = { year, sections: [], notes: null, error: (e as Error)?.message || tr("gl_statement_set_failed", "คำนวณไม่สำเร็จ") };
    }
    if (request !== requestRef.current) return;
    setPrepared(next);
    setPreparing(false);
  }

  const readySections = prepared?.sections.filter((section) => section.report) ?? [];
  const failedCount = (prepared?.sections.filter((section) => !section.report).length ?? 0) + (prepared?.notes?.error ? 1 : 0) + (prepared?.error ? 1 : 0);
  const reviewCount = readySections.filter((section) => section.report && reportNeedsReview(section.report)).length;
  const readyCount = readySections.length + (prepared?.notes && !prepared.notes.error && prepared.notes.items.length ? 1 : 0);
  const canPrint = Boolean(prepared) && !preparing && failedCount === 0 && readyCount > 0 && !print.printing;

  function printSet() {
    if (!prepared || !canPrint) return;
    const preparedYear = refs.years.find((item) => item.code === prepared.year);
    const sections: GLStatementSetSection[] = prepared.sections.flatMap(({ code, title, statementtype, showNote, scale, report }) => report ? [{
      code,
      title,
      period: statementPeriodLine(statementtype, report, refs.years, tr, language),
      report,
      showNote,
      scale,
    }] : []);
    const notesPeriod = preparedYear ? statementPeriodText({ from: preparedYear.startdate, to: preparedYear.enddate }, false, true, tr, language) : "";
    print.print(<GLStatementSetPrint sections={sections} notes={prepared.notes?.items ?? []} company={printCompanyName()} notesPeriod={notesPeriod} tr={tr} />, statementSetRootOrientation(sections));
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    // Combobox ปิดรายการตัวเลือกด้วย Escape เอง (preventDefault) — กดครั้งนั้นต้องไม่ปิดทั้ง dialog
    if (event.key === "Escape" && !event.defaultPrevented) {
      event.preventDefault();
      if (!preparing) onClose();
      return;
    }
    if (event.key !== "Tab" || !dialogRef.current) return;
    // วนโฟกัสอยู่ใน dialog ไม่หลุดไปจอข้างหลัง (:disabled รวมช่องใน fieldset ที่ปิดระหว่างคำนวณ; ไม่มีช่องให้ไป = โฟกัสอยู่ที่ dialog)
    const focusable = Array.from(dialogRef.current.querySelectorAll<HTMLElement>("button:not(:disabled), input:not(:disabled), select:not(:disabled), [tabindex]:not([tabindex='-1'])"));
    if (!focusable.length) { event.preventDefault(); return; }
    const first = focusable[0], last = focusable[focusable.length - 1];
    if (event.shiftKey && (document.activeElement === first || document.activeElement === dialogRef.current)) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  const notesDisabledReason = notesInfo.loading
    ? tr("gl_loading_data", "กำลังโหลดข้อมูล…")
    : notesInfo.error
      ? tr("gl_statement_set_notes_failed", "โหลดหมายเหตุประกอบงบการเงินไม่สำเร็จ: {0}").replace("{0}", notesInfo.error)
      : !year
        ? tr("gl_select_fiscal_year", "เลือกปีบัญชี")
        : notesInfo.count === 0
        ? tr("gl_statement_set_notes_none", "ปีบัญชีนี้ยังไม่มีหมายเหตุประกอบงบการเงิน — เขียนได้ที่โหมด “หมายเหตุประกอบงบการเงิน”")
        : "";

  // บอกเหตุผลที่กดปุ่มตรวจไม่ได้ก่อนชวนให้กด (ไม่มีปีบัญชี = ปุ่ม disabled)
  const footerHint = templatesLoading || notesInfo.loading || preparing || prepared
    ? ""
    : !fiscalYear
      ? refs.error
        ? tr("gl_statement_set_years_failed", "โหลดปีบัญชีไม่สำเร็จ: {0}").replace("{0}", refs.error)
        : tr("gl_statement_set_select_year", "เลือกปีบัญชีก่อน — ถ้ายังไม่มีปีบัญชี ให้สร้างที่เมนู “ปีบัญชีและบัญชีปิดปี”")
      : nothingSelected
        ? tr("gl_statement_set_select_one", "เลือกงบหรือหมายเหตุอย่างน้อย 1 รายการ")
        : tr("gl_statement_set_prepare_hint", "กด “ตรวจและเตรียมพิมพ์” เพื่อคำนวณงบทุกรายการจากข้อมูลล่าสุดก่อนพิมพ์");

  return (
    <>
      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-xs">
          <div
            ref={dialogRef}
            role="dialog"
            aria-modal="true"
            aria-labelledby={titleId}
            aria-describedby={descriptionId}
            aria-busy={preparing}
            tabIndex={-1}
            onKeyDown={onKeyDown}
            className="flex max-h-[92vh] w-full max-w-3xl flex-col rounded-2xl border border-border bg-card text-foreground shadow-xl outline-none"
          >
            <header className="flex items-start justify-between gap-3 border-b border-border px-5 py-4">
              <div className="min-w-0">
                <h2 id={titleId} className="flex items-center gap-2 text-lg font-bold leading-snug">
                  <Printer aria-hidden className="size-5 shrink-0 text-primary" /> {tr("gl_statement_set_print", "พิมพ์ชุดงบการเงิน")}
                </h2>
                <p id={descriptionId} className="mt-1 text-[0.95rem] leading-relaxed text-muted-foreground">
                  {tr("gl_statement_set_desc", "พิมพ์งบที่เลือกและหมายเหตุประกอบงบการเงินต่อเนื่องในงานพิมพ์เดียว เรียงตามแบบ 2 ของกรมพัฒนาธุรกิจการค้า ใช้รูปแบบงบที่บันทึกแล้วเท่านั้น")}
                </p>
              </div>
              <Button type="button" variant="outline" className={actionClass} onClick={onClose} disabled={preparing}>
                <X aria-hidden className="mr-1.5 h-4 w-4" /> {tr("gl_close", "ปิด")}
              </Button>
            </header>

            <div className="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
              {unsavedChanges && (
                <div role="note" className="flex items-start gap-2 rounded-xl border border-primary/40 bg-primary/10 p-3 text-[0.95rem] leading-relaxed text-foreground">
                  <AlertTriangle aria-hidden className="mt-1 size-4 shrink-0 text-primary" />
                  <span>{tr("gl_statement_set_unsaved", "จอนี้มีการแก้ไขที่ยังไม่บันทึก — ชุดงบที่พิมพ์ใช้ฉบับที่บันทึกแล้วเท่านั้น การแก้ไขที่ค้างอยู่จะไม่ถูกพิมพ์")}</span>
                </div>
              )}

              <div className="grid gap-3 sm:grid-cols-[minmax(12rem,16rem)_1fr] sm:items-end">
                <div className="grid gap-1 text-[0.95rem]">
                  <span className="font-medium">{tr("gl_fiscal_year", "ปีบัญชี")}</span>
                  <YearSelect value={year} onChange={changeYear} years={refs.years} disabled={preparing} />
                </div>
                <p className="pb-2 text-[0.95rem] leading-relaxed text-muted-foreground">
                  {fiscalYear ? tr("gl_statement_set_period", "งวด: ทั้งปีบัญชี {0} ถึง {1}").replace("{0}", formatAppDate(fiscalYear.startdate, language)).replace("{1}", formatAppDate(fiscalYear.enddate, language)) : ""}
                </p>
              </div>

              <fieldset className="grid gap-2" disabled={preparing}>
                <legend className="mb-2 text-[0.95rem] font-semibold">{tr("gl_statement_set_statements", "งบที่จะพิมพ์")}</legend>
                {templatesLoading && <p role="status" className="text-[0.95rem] text-muted-foreground">{tr("gl_loading_data", "กำลังโหลดข้อมูล…")}</p>}
                {templatesError && (
                  <p role="alert" className="flex items-start gap-2 text-[0.95rem] text-destructive">
                    <XCircle aria-hidden className="mt-1 size-4 shrink-0" /> {tr("gl_statement_set_load_failed", "โหลดรายการรูปแบบงบการเงินไม่สำเร็จ: {0}").replace("{0}", templatesError)}
                  </p>
                )}
                {!templatesLoading && !templatesError && templates.length === 0 && (
                  <p className="rounded-xl border border-dashed border-border p-3 text-[0.95rem] leading-relaxed text-muted-foreground">
                    {tr("gl_statement_set_empty", "ยังไม่มีรูปแบบงบการเงินที่เปิดใช้งาน — สร้างและบันทึกรูปแบบงบก่อน")}
                  </p>
                )}
                {templates.length > 0 && (
                  <ul className="grid gap-2">
                    {templates.map((template) => (
                      <li key={template.code}>
                        <label className={rowClass}>
                          <Checkbox checked={selected.includes(template.code)} onCheckedChange={(checked) => toggleTemplate(template.code, checked)} />
                          <span className="min-w-0 flex-1">
                            <span className="block font-semibold [overflow-wrap:anywhere]">{template.name || template.code}</span>
                            <span className="block text-[0.9rem] text-muted-foreground">
                              {labelText(statementTypeLabels, template.statementtype, tr)} · {tr("gl_statement_set_template_code", "รหัส {0}").replace("{0}", template.code)}
                            </span>
                          </span>
                        </label>
                      </li>
                    ))}
                  </ul>
                )}
                <label className={rowClass}>
                  <Checkbox checked={includeNotes} disabled={notesInfo.loading || notesInfo.count === 0} onCheckedChange={toggleNotes} />
                  <span className="min-w-0 flex-1">
                    <span className="block font-semibold">{tr("gl_statement_set_notes_count", "หมายเหตุประกอบงบการเงิน ({0} ข้อ)").replace("{0}", String(notesInfo.count))}</span>
                    {notesDisabledReason && <span className="block text-[0.9rem] leading-relaxed text-muted-foreground">{notesDisabledReason}</span>}
                  </span>
                </label>
              </fieldset>

              {preparing && (
                <p role="status" className="flex items-center gap-2 text-[0.95rem] font-medium">
                  <Loader2 aria-hidden className="size-4 shrink-0 motion-safe:animate-spin" /> {tr("gl_statement_set_preparing", "กำลังคำนวณงบ…")}
                </p>
              )}

              {prepared && (
                <section aria-labelledby={`${titleId}-results`} className="space-y-3 rounded-xl border border-border bg-muted/20 p-3">
                  <h3 ref={resultsHeadingRef} id={`${titleId}-results`} tabIndex={-1} className="text-[0.95rem] font-semibold outline-none">{tr("gl_statement_set_results", "ผลการตรวจก่อนพิมพ์")}</h3>
                  {prepared.error ? (
                    <p role="alert" className="flex items-start gap-2 text-[0.95rem] leading-relaxed text-destructive [overflow-wrap:anywhere]">
                      <XCircle aria-hidden className="mt-1 size-4 shrink-0" /> {tr("gl_statement_set_request_failed", "ตรวจชุดงบไม่สำเร็จ: {0}").replace("{0}", prepared.error)}
                    </p>
                  ) : <>
                  <ol className="space-y-3">
                    {prepared.sections.map(({ code, title, report, error, scale }) => (
                      <li key={code} className="space-y-2">
                        <div className="flex flex-wrap items-center justify-between gap-2 text-[0.95rem]">
                          <span className="font-semibold [overflow-wrap:anywhere]">{title}</span>
                          {!report ? (
                            <span className="inline-flex items-center gap-1.5 font-semibold text-destructive"><XCircle aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_set_failed", "คำนวณไม่สำเร็จ")}</span>
                          ) : reportNeedsReview(report) ? (
                            <span className="inline-flex items-center gap-1.5 font-semibold text-primary"><AlertTriangle aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_set_has_warnings", "พร้อมพิมพ์ มีข้อควรตรวจสอบ")}</span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 font-semibold text-primary"><CheckCircle2 aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_set_ready", "พร้อมพิมพ์")}</span>
                          )}
                        </div>
                        {error && <p role="alert" className="text-[0.95rem] leading-relaxed text-destructive [overflow-wrap:anywhere]">{error}</p>}
                        {report && <GLReportWarnings warnings={report.warnings} tr={tr} />}
                        {report && <GLStatementChecks checks={report.checks} scale={scale} tr={tr} />}
                        {report && <GLStatementUnassigned items={report.unassigned} scale={scale} tr={tr} />}
                      </li>
                    ))}
                    {prepared.notes && (
                      <li className="space-y-2">
                        <div className="flex flex-wrap items-center justify-between gap-2 text-[0.95rem]">
                          <span className="font-semibold">{tr("gl_statement_set_notes_count", "หมายเหตุประกอบงบการเงิน ({0} ข้อ)").replace("{0}", String(prepared.notes.items.length))}</span>
                          {prepared.notes.error ? (
                            <span className="inline-flex items-center gap-1.5 font-semibold text-destructive"><XCircle aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_set_failed", "คำนวณไม่สำเร็จ")}</span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 font-semibold text-primary"><CheckCircle2 aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_set_ready", "พร้อมพิมพ์")}</span>
                          )}
                        </div>
                        {prepared.notes.error && <p role="alert" className="text-[0.95rem] leading-relaxed text-destructive [overflow-wrap:anywhere]">{prepared.notes.error}</p>}
                      </li>
                    )}
                  </ol>
                  <div role="status" className="space-y-1 border-t border-border pt-2 text-[0.95rem] leading-relaxed">
                    <p className="font-semibold">{tr("gl_statement_set_summary_ready", "พร้อมพิมพ์ {0} รายการ").replace("{0}", String(readyCount))}</p>
                    {reviewCount > 0 && <p className="flex items-start gap-1.5 text-primary"><AlertTriangle aria-hidden className="mt-1 size-4 shrink-0" /> {tr("gl_statement_set_summary_warnings", "มีข้อควรตรวจสอบ {0} รายการ — พิมพ์ได้ แต่ควรตรวจก่อนออกงบ").replace("{0}", String(reviewCount))}</p>}
                    {failedCount > 0 && <p className="flex items-start gap-1.5 text-destructive"><XCircle aria-hidden className="mt-1 size-4 shrink-0" /> {tr("gl_statement_set_summary_errors", "คำนวณไม่สำเร็จ {0} รายการ — แก้ไข หรือยกเลิกการเลือกรายการนั้นก่อนพิมพ์").replace("{0}", String(failedCount))}</p>}
                  </div>
                  </>}
                </section>
              )}
            </div>

            <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-border px-5 py-4">
              <p className="min-w-0 flex-1 text-[0.95rem] leading-relaxed text-muted-foreground">
                {footerHint}
              </p>
              <div className="flex flex-wrap gap-2">
                <Button type="button" variant={prepared ? "outline" : "default"} className={actionClass} onClick={() => void prepare()} disabled={preparing || templatesLoading || notesInfo.loading || !fiscalYear || nothingSelected}>
                  {preparing ? <Loader2 aria-hidden className="mr-1.5 h-4 w-4 motion-safe:animate-spin" /> : <ClipboardCheck aria-hidden className="mr-1.5 h-4 w-4" />} {tr("gl_statement_set_prepare", "ตรวจและเตรียมพิมพ์")}
                </Button>
                <Button type="button" variant={prepared ? "default" : "outline"} ref={printButtonRef} className={actionClass} onClick={printSet} disabled={!canPrint}>
                  <Printer aria-hidden className="mr-1.5 h-4 w-4" /> {tr("gl_statement_set_print", "พิมพ์ชุดงบการเงิน")}
                </Button>
              </div>
            </footer>
          </div>
        </div>
      )}
      {print.portal}
    </>
  );
}
