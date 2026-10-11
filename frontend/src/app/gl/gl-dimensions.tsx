"use client";

import { useCallback, useEffect, useState } from "react";
import { Printer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { DIMENSIONS, DIMENSION_RESOURCES, dimensionDef, fillText, type GLDimension, type GLDimensionResource, type GLTextFn } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { cn } from "@/lib/utils";
import { Field, Notice, actionClass, control, useGLLanguage, useGLText } from "./gl-common";
import { TemplateModal as Modal } from "./gl-journal-templates";
import { GLMasters } from "./gl-masters";
import { codeRangeLines, codeRangePrintRows } from "./gl-print-conditions";
import { GLDimensionsPrint, printCompanyName, useGLPrint } from "./gl-print";

/** Rows of a register in print order (the shared code-range loader of gl-print-conditions.tsx). `request` is injectable for tests; `complete` is false when the backend reported more rows than were read. */
export function dimensionPrintRows(kind: GLDimensionResource, from: string, to: string, request: <T>(path: string) => Promise<T> = glRequest): Promise<{ rows: GLDimension[]; complete: boolean }> {
  return codeRangePrintRows<GLDimension>(kind, from, to, request);
}

/** Print-condition line of the report footer: nothing for the whole register, "equals" for one code, a range for from + to. */
export function dimensionConditionLines(kind: GLDimensionResource, from: string, to: string, tr: GLTextFn): string[] {
  const def = dimensionDef(kind);
  return codeRangeLines(tr(...def.label), from, to, tr);
}

function PrintDialog({ kind, print, onClose }: { kind: GLDimensionResource; print: ReturnType<typeof useGLPrint>["print"]; onClose: () => void }) {
  const tr = useGLText(), language = useGLLanguage();
  const [from, setFrom] = useState(""), [to, setTo] = useState(""), [busy, setBusy] = useState(false), [error, setError] = useState("");
  const def = dimensionDef(kind);
  const noun = tr(...def.label);
  const title = tr(...def.reportTitle);
  async function run() {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const { rows, complete } = await dimensionPrintRows(kind, from, to);
      if (!complete) setError(tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
      else if (!rows.length) setError(tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก"));
      else {
        print(<GLDimensionsPrint kind={kind} rows={rows} conditions={dimensionConditionLines(kind, from, to, tr)} company={printCompanyName()} tr={tr} language={language} />, "portrait");
        onClose();
      }
    } catch (cause) {
      setError(cause instanceof Error && cause.message ? cause.message : tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    } finally {
      setBusy(false);
    }
  }
  const submitOnEnter = (event: { key: string; preventDefault: () => void }) => { if (event.key === "Enter") { event.preventDefault(); void run(); } };
  return (
    <Modal label={title} onClose={onClose}>
      <h3 className="font-semibold">{title}</h3>
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label={fillText(tr("gl_dimension_filter_from", "จาก{0}"), noun)}>
          <input className={control} data-field="codefrom" autoFocus value={from} onChange={(event) => { setFrom(event.target.value); setError(""); }} onKeyDown={submitOnEnter} />
        </Field>
        <Field label={fillText(tr("gl_dimension_filter_to", "ถึง{0}"), noun)}>
          <input className={control} data-field="codeto" disabled={!from.trim()} value={to} onChange={(event) => { setTo(event.target.value); setError(""); }} onKeyDown={submitOnEnter} />
        </Field>
      </div>
      <p className="text-muted-foreground">{tr("gl_dimension_print_hint", "เว้นว่างทั้งสองช่อง = พิมพ์ทุกรายการ · ใส่เฉพาะช่อง “จาก” = พิมพ์เฉพาะรหัสนั้น")}</p>
      <Notice error text={error} />
      <div className="flex flex-wrap justify-end gap-2">
        <Button type="button" variant="outline" className={actionClass} disabled={busy} onClick={onClose}>{tr("gl_cancel", "ยกเลิก")}</Button>
        <Button type="button" className={actionClass} disabled={busy} onClick={() => void run()}><Printer className="size-4 mr-1.5" />{tr("gl_print", "พิมพ์")}</Button>
      </div>
    </Modal>
  );
}

/** "Print" button of one register tab; opens the code-range dialog. The print job lives here, not in the dialog, because the dialog closes as the job starts. */
function PrintAction({ kind, autoOpen = false }: { kind: GLDimensionResource; autoOpen?: boolean }) {
  const tr = useGLText();
  const { print, portal } = useGLPrint();
  const [open, setOpen] = useState(autoOpen);
  useEffect(() => {
    if (autoOpen) setOpen(true);
  }, [autoOpen]);
  const def = dimensionDef(kind);
  const title = tr(...def.reportTitle);
  return <>
    <Button type="button" variant="outline" className={actionClass} title={title} onClick={() => setOpen(true)}><Printer className="size-4 mr-1.5" />{tr("gl_print", "พิมพ์")}</Button>
    {open && <PrintDialog kind={kind} print={print} onClose={() => setOpen(false)} />}
    {portal}
  </>;
}

function tabFromRoute(route: string): GLDimensionResource {
  for (const resource of DIMENSION_RESOURCES) {
    if (route.includes(`tab=${resource}`) || route.includes(resource)) {
      return resource;
    }
  }
  if (route.includes("allocat")) return "allocate-codes";
  if (route.includes("job")) return "job-codes";
  if (route.includes("part")) return "part-codes";
  if (route.includes("side")) return "side-codes";
  if (route.includes("project")) return "projects";
  return "departments";
}

export function GLDimensions({ route }: { route: string }) {
  const tr = useGLText();
  const [tab, setTab] = useState<GLDimensionResource>(() => tabFromRoute(route)), [dirty, setDirty] = useState(false);
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const onDirtyChange = useCallback((value: boolean) => setDirty(value), []);

  useEffect(() => {
    if (!route) return;
    setTab(tabFromRoute(route));
  }, [route]);

  async function select(next: GLDimensionResource) {
    if (next === tab) return;
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    setDirty(false);
    setTab(next);
  }

  return <div className="flex min-h-0 flex-1 flex-col gap-2" data-gl-dimensions="true">
    <div role="tablist" aria-label={tr("gl_dimension_tabs_label", "ประเภทรหัส")} className="flex shrink-0 flex-wrap gap-2">
      {DIMENSIONS.map((def) => <button key={def.resource} type="button" role="tab" aria-selected={tab === def.resource} onClick={() => void select(def.resource)}
        className={cn("min-h-11 rounded-xl border px-4 py-2 font-semibold transition-colors", tab === def.resource ? "border-primary bg-primary text-primary-foreground shadow-sm" : "border-border bg-card text-foreground hover:border-primary/60 hover:bg-primary/5")}>
        {tr(...def.label)}
      </button>)}
    </div>
    <GLMasters key={tab} resource={tab} route={route} extraActions={<PrintAction kind={tab} autoOpen={route.includes("print=true")} />} onDirtyChange={onDirtyChange} />
    {confirmationDialog}
  </div>;
}
