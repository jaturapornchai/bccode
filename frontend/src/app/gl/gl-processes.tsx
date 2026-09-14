"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { localDate, type GLReport } from "@/lib/general-ledger";
import { Field, Notice, Pager, YearSelect, actionClass, control, panel, useDirtyGuard, useGLCommand, useReferences } from "./gl-common";
import { ReportGrid, emptyReportFilters, fetchReport } from "./gl-reports";

const processText = {
  close: { button: "สร้างฉบับร่างปิดงวด", text: "ตรวจสอบกระดาษทำการตั้งแต่ต้นปีถึงวันที่ดำเนินการ แล้วสร้างรายการปิดงวดเป็นฉบับร่างแยกตามสาขา หลังตรวจและผ่านรายการทุกฉบับให้ล็อกช่วงวันที่ในเมนูล็อกงวดบัญชี" },
  "year-end": { button: "ประมวลผลสิ้นปี", text: "ต้องผ่านรายการปิดรายได้และค่าใช้จ่ายก่อน เลือกปีถัดไปที่ต่อเนื่องกับปีเดิม ระบบจะปิดปีเดิมและสร้างยอดยกมาเป็นฉบับร่างในปีถัดไปให้ตรวจและผ่านรายการ" },
  recalculate: { button: "คำนวณยอดผ่านรายการใหม่", text: "ตรวจยอดจากรายการที่ผ่านบัญชีแล้ว ก่อนปรับปรุงยอดประมวลผลสำหรับรายงาน" },
  reprocess: { button: "ประมวลผลเอกสารต้นทาง", text: "สร้างรายการบัญชีจากเอกสารต้นทางและรูปแบบการเชื่อมโยงบัญชีที่กำหนด" },
};
export function GLProcesses({ route, action }: { route: string; action: keyof typeof processText }) {
  const refs = useReferences(), { busy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog();
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
    if (!reason.trim() || !date || !docno.trim()) { setError("กรุณาระบุวันที่ เลขที่เอกสาร และเหตุผลก่อนดำเนินการ"); return; }
    if (action === "year-end" && !targetYear) { setError("กรุณาเลือกปีบัญชีถัดไปสำหรับยอดยกมา"); return; }
    if (!await confirm({ title: `${config.button}?`, description: `ปีบัญชี ${yearCode} · เอกสาร ${docno} · วันที่ ${date}`, details: reason, confirmLabel: config.button, tone: "warning" })) return;
    try {
      const result = await execute({ resource: "processes", action, id: yearCode, date, docno, reason, version: year.version, ...(action === "year-end" ? { targetyear: targetYear } : {}) });
      const count = result.createdjournals ?? 1;
      const documents = `${count} เอกสาร เลขที่ ${docno}${count > 1 ? "-001 เป็นต้น" : ""}`;
      setMessage(action === "close" ? `สร้างฉบับร่าง ${documents} กรุณาเปิดสมุดรายวันทั่วไปเพื่อตรวจและผ่านรายการทุกฉบับ แล้วล็อกช่วงวันที่ในเมนูล็อกงวดบัญชี` : action === "year-end" ? count === 0 ? "ปิดปีบัญชีแล้ว ไม่มียอดคงเหลือที่ต้องยกไปปีถัดไป" : `ปิดปีเดิมและสร้างฉบับร่างยอดยกมา ${documents} กรุณาตรวจสอบและผ่านรายการยอดยกมาในปีถัดไป` : "ดำเนินการเรียบร้อยแล้ว กรุณาโหลดรายงานเพื่อตรวจสอบยอด");
      setDocno(""); setReason(""); setPreview(null); setError("");
    } catch (e) { setError((e as Error).message); }
  }
  return <section className={`${panel} grid gap-3`}><Notice text={config.text} /><Notice error text={error || refs.error} /><Notice text={message} />
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><Field label="ปีบัญชี"><YearSelect years={refs.years} value={yearCode} onChange={setYear} /></Field><Field label="วันที่ดำเนินการ"><input className={control} type="date" value={date} onChange={(e) => setDate(e.target.value)} /></Field><Field label="เลขที่เอกสาร"><input className={control} value={docno} onChange={(e) => setDocno(e.target.value)} /></Field><Field label="เหตุผล"><input className={control} value={reason} onChange={(e) => setReason(e.target.value)} /></Field></div>
    {action === "year-end" && <Field label="ปีบัญชีถัดไปสำหรับยอดยกมา"><YearSelect label="ปีบัญชีถัดไปสำหรับยอดยกมา" years={refs.years.filter((item) => item.code !== yearCode && !item.closed && item.isactive)} value={targetYear} onChange={(value) => { setTargetYear(value); setDate(refs.years.find((item) => item.code === value)?.startdate ?? ""); }} /></Field>}
    {year && <p className="text-[0.95rem]">{year.startdate} ถึง {year.enddate} · {year.currency} · {year.closed ? "ปิดปีแล้ว" : "ปีบัญชีเปิดใช้งาน"}</p>}
    <div className="flex flex-wrap gap-2"><Button className={actionClass} variant="outline" disabled={!year || busy || loading} onClick={() => void inspect()}>{loading ? "กำลังตรวจสอบ…" : "ตรวจสอบยอดก่อนดำเนินการ"}</Button><Button className={actionClass} disabled={!preview || previewFor !== previewKey || busy || loading || action === "reprocess" || action === "year-end" && !targetYear} onClick={() => void process()}>{config.button}</Button></div>
    {action === "reprocess" && <Notice text="ยังไม่เปิดประมวลผลเอกสารต้นทางในหน้านี้ ต้องกำหนดชนิดเอกสารและตรวจสอบการเชื่อมจำนวนเงินก่อน เพื่อป้องกันลงบัญชีซ้ำหรือเลือกยอดผิด" />}
    {preview && <><ReportGrid report={preview} /><Pager page={previewPage} total={preview.totalrows} onPage={(page) => void inspect(page)} loading={loading || busy} limit={50} /></>}{confirmationDialog}
  </section>;
}
