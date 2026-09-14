"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { localDate, type GLLabel, type GLReport } from "@/lib/general-ledger";
import { Field, Notice, Pager, YearSelect, actionClass, control, panel, useDirtyGuard, useGLCommand, useReferences, useGLText } from "./gl-common";
import { ReportGrid, emptyReportFilters, fetchReport } from "./gl-reports";

const processText: Record<string, { button: GLLabel; text: GLLabel }> = {
  close: { button: ["gl_create_close_period_draft", "สร้างฉบับร่างปิดงวด"], text: ["gl_review_ws_create_close_draft", "ตรวจสอบกระดาษทำการตั้งแต่ต้นปีถึงวันที่ดำเนินการ แล้วสร้างรายการปิดงวดเป็นฉบับร่างแยกตามสาขา หลังตรวจและผ่านรายการทุกฉบับให้ล็อกช่วงวันที่ในเมนูล็อกงวดบัญชี"] },
  "year-end": { button: ["gl_process_year_end", "ประมวลผลสิ้นปี"], text: ["gl_close_income_expense_first", "ต้องผ่านรายการปิดรายได้และค่าใช้จ่ายก่อน เลือกปีถัดไปที่ต่อเนื่องกับปีเดิม ระบบจะปิดปีเดิมและสร้างยอดยกมาเป็นฉบับร่างในปีถัดไปให้ตรวจและผ่านรายการ"] },
  recalculate: { button: ["gl_recalc_posted_balances", "คำนวณยอดผ่านรายการใหม่"], text: ["gl_verify_balances_before_update", "ตรวจยอดจากรายการที่ผ่านบัญชีแล้ว ก่อนปรับปรุงยอดประมวลผลสำหรับรายงาน"] },
  reprocess: { button: ["gl_process_source_docs", "ประมวลผลเอกสารต้นทาง"], text: ["gl_create_entries_from_source", "สร้างรายการบัญชีจากเอกสารต้นทางและรูปแบบการเชื่อมโยงบัญชีที่กำหนด"] },
};
export function GLProcesses({ route, action }: { route: string; action: "close" | "year-end" | "recalculate" | "reprocess" }) {
  const tr = useGLText();
  const refs = useReferences(), { busy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const [yearCode, setYear] = useState(""), [date, setDate] = useState(localDate()), [docno, setDocno] = useState(""), [reason, setReason] = useState("");

  const [targetYear, setTargetYear] = useState(""), [previewPage, setPreviewPage] = useState(1);
  const [preview, setPreview] = useState<GLReport | null>(null), [previewFor, setPreviewFor] = useState("");
  const [loading, setLoading] = useState(false), [message, setMessage] = useState(""), [error, setError] = useState("");
  const year = refs.years.find((item) => item.code === yearCode), config = processText[action];
  const previewKey = JSON.stringify({ year: yearCode, to: action === "close" ? date : year?.enddate, targetYear });
  useDirtyGuard(route, !!docno || !!reason);
  async function inspect(page = 1) {
    if (!year || loading) return;
    setLoading(true); setError("");
    try { const report = await fetchReport(action === "year-end" ? "trialbalance" : "workingpaper", { ...emptyReportFilters, fiscalyear: yearCode, from: year.startdate, to: action === "close" ? date : year.enddate }, page, 50, page > 1 && previewFor === previewKey ? preview?.sequence : undefined); setPreview(report); setPreviewFor(previewKey); setPreviewPage(page); }
    catch (e) { setError((e as Error).message); } finally { setLoading(false); }
  }
  async function process() {
    if (!year || !preview || previewFor !== previewKey || busy) return;
    if (!reason.trim() || !date || !docno.trim()) { setError(tr("gl_specify_date_doc_reason", "กรุณาระบุวันที่ เลขที่เอกสาร และเหตุผลก่อนดำเนินการ")); return; }
    if (action === "year-end" && !targetYear) { setError(tr("gl_select_next_fy_opening", "กรุณาเลือกปีบัญชีถัดไปสำหรับยอดยกมา")); return; }
    if (!await confirm({ title: `${tr(...config.button)}?`, description: tr("gl_fy_doc_date", "ปีบัญชี {0} · เอกสาร {1} · วันที่ {2}").replace("{0}", String(yearCode)).replace("{1}", String(docno)).replace("{2}", String(date)), details: reason, confirmLabel: tr(...config.button), tone: "warning" })) return;
    try {
      const result = await execute({ resource: "processes", action, id: yearCode, date, docno, reason, version: year.version, ...(action === "year-end" ? { targetyear: targetYear } : {}) });
      const count = result.createdjournals ?? 1;
      const documents = tr("gl_x_docs_no", "{0} เอกสาร เลขที่ {1}{2}").replace("{0}", String(count)).replace("{1}", String(docno)).replace("{2}", String(count > 1 ? tr("gl_001_etc", "-001 เป็นต้น") : ""));
      setMessage(action === "close" ? tr("gl_draft_created_open_gj_lock_period", "สร้างฉบับร่าง {0} กรุณาเปิดสมุดรายวันทั่วไปเพื่อตรวจและผ่านรายการทุกฉบับ แล้วล็อกช่วงวันที่ในเมนูล็อกงวดบัญชี").replace("{0}", String(documents)) : action === "year-end" ? count === 0 ? tr("gl_fy_closed_no_balance_cf", "ปิดปีบัญชีแล้ว ไม่มียอดคงเหลือที่ต้องยกไปปีถัดไป") : tr("gl_close_year_create_ob_draft", "ปิดปีเดิมและสร้างฉบับร่างยอดยกมา {0} กรุณาตรวจสอบและผ่านรายการยอดยกมาในปีถัดไป").replace("{0}", String(documents)) : tr("gl_completed_download_report", "ดำเนินการเรียบร้อยแล้ว กรุณาโหลดรายงานเพื่อตรวจสอบยอด"));
      setDocno(""); setReason(""); setPreview(null); setError("");
    } catch (e) { setError((e as Error).message); }
  }
  return <section className={`${panel} grid gap-3`}><Notice text={tr(...config.text)} /><Notice error text={error || refs.error} /><Notice text={message} />
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={refs.years} value={yearCode} onChange={setYear} /></Field><Field label={tr("gl_processing_date", "วันที่ดำเนินการ")}><input className={control} type="date" value={date} onChange={(e) => setDate(e.target.value)} /></Field><Field label={tr("gl_document_no", "เลขที่เอกสาร")}><input className={control} value={docno} onChange={(e) => setDocno(e.target.value)} /></Field><Field label={tr("gl_reason", "เหตุผล")}><input className={control} value={reason} onChange={(e) => setReason(e.target.value)} /></Field></div>
    {action === "year-end" && <Field label={tr("gl_next_fy_opening_balance", "ปีบัญชีถัดไปสำหรับยอดยกมา")}><YearSelect label={tr("gl_next_fy_opening_balance", "ปีบัญชีถัดไปสำหรับยอดยกมา")} years={refs.years.filter((item) => item.code !== yearCode && !item.closed && item.isactive)} value={targetYear} onChange={(value) => { setTargetYear(value); setDate(refs.years.find((item) => item.code === value)?.startdate ?? ""); }} /></Field>}
    {year && <p className="text-[0.95rem]">{tr("gl_range_to_period_info", "{0} ถึง {1} · {2} · {3}").replace("{0}", String(year.startdate)).replace("{1}", String(year.enddate)).replace("{2}", String(year.closed ? tr("gl_year_closed", "ปิดปีแล้ว") : tr("gl_active_fiscal_year", "ปีบัญชีเปิดใช้งาน")))}</p>}
    <div className="flex flex-wrap gap-2"><Button className={actionClass} variant="outline" disabled={!year || busy || loading} onClick={() => void inspect()}>{loading ? tr("gl_checking", "กำลังตรวจสอบ…") : tr("gl_verify_balances_before", "ตรวจสอบยอดก่อนดำเนินการ")}</Button><Button className={actionClass} disabled={!preview || previewFor !== previewKey || busy || loading || action === "reprocess" || action === "year-end" && !targetYear} onClick={() => void process()}>{config.button}</Button></div>
    {action === "reprocess" && <Notice text={tr("gl_source_doc_not_enabled_config", "ยังไม่เปิดประมวลผลเอกสารต้นทางในหน้านี้ ต้องกำหนดชนิดเอกสารและตรวจสอบการเชื่อมจำนวนเงินก่อน เพื่อป้องกันลงบัญชีซ้ำหรือเลือกยอดผิด")} />}
    {preview && <><ReportGrid report={preview} /><Pager page={previewPage} total={preview.totalrows} onPage={(page) => void inspect(page)} loading={loading || busy} limit={50} /></>}{confirmationDialog}
  </section>;
}
