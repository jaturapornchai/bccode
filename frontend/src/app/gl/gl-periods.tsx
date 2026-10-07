"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { CalendarPlus, Lock, LockOpen, RefreshCw, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { fillText, localDate, type GLFiscalYear, type GLMaster } from "@/lib/general-ledger";
import { GLCommandError, commandFailure, glAllRecords } from "@/lib/general-ledger-api";
import { formatAppDate } from "@/lib/date-time";
import { toast } from "@/lib/toast";
import { cn } from "@/lib/utils";
import { Field, Notice, YearSelect, actionClass, control, panel, useGLCommand, useGLLanguage, useGLText, useReferences } from "./gl-common";

/** ปีบัญชีที่เปิดจอมาแล้วเลือกให้: ปีที่ครอบวันนี้ → ปีที่เริ่มล่าสุด → ว่าง */
export function defaultPeriodYear(years: readonly GLFiscalYear[], today: string): string {
  const live = years.filter((year) => !year.isdeleted);
  return live.find((year) => today >= year.startdate && today <= year.enddate)?.code
    ?? [...live].sort((a, b) => b.startdate.localeCompare(a.startdate))[0]?.code ?? "";
}
/** งวดของปีที่เลือก เรียงตามวันเริ่มต้น (แล้วตามรหัส) */
export function yearPeriods(periods: readonly GLMaster[], year: string): GLMaster[] {
  return periods.filter((period) => !period.isdeleted && period.fiscalyear === year)
    .sort((a, b) => (a.startdate ?? "").localeCompare(b.startdate ?? "") || (a.code ?? "").localeCompare(b.code ?? "", undefined, { numeric: true }));
}
export function periodSummary(periods: readonly GLMaster[]) {
  return { total: periods.length, locked: periods.filter((period) => period.locked).length };
}
/** ล็อก/ปลดล็อกงวดต้องมีเหตุผล (backend postgres_guards.go mutatePeriod) — ตรวจก่อนส่ง */
export function periodReasonMissing(reason: string): boolean {
  return !reason.trim();
}
/** ปุ่มลัดแบบ Champ 5511: F9 สร้างงวดอัตโนมัติ */
export function periodShortcut(key: string): "generate" | null {
  return key === "F9" ? "generate" : null;
}

/** กำหนดงวดบัญชี (Champ 5511): ตารางงวดของปีบัญชี + สร้างงวดรายเดือนอัตโนมัติ + ล็อก/ปลดล็อกพร้อมเหตุผล */
export function GLPeriods() {
  const tr = useGLText(), language = useGLLanguage();
  const refs = useReferences();
  const years = useMemo(() => refs.years.filter((year) => !year.isdeleted).sort((a, b) => b.startdate.localeCompare(a.startdate)), [refs.years]);
  const [pickedYear, setPickedYear] = useState("");
  const yearCode = pickedYear || defaultPeriodYear(years, localDate());
  const year = years.find((item) => item.code === yearCode);
  const [periods, setPeriods] = useState<GLMaster[]>([]);
  const [loading, setLoading] = useState(false);
  const [revision, setRevision] = useState(0);
  const [reason, setReason] = useState("");
  const [error, setError] = useState(""), [info, setInfo] = useState("");
  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const rootRef = useRef<HTMLElement>(null), reasonRef = useRef<HTMLInputElement>(null);
  const rows = useMemo(() => yearPeriods(periods, yearCode), [periods, yearCode]);
  const summary = periodSummary(rows);
  const working = busy || loading;
  const closed = Boolean(year?.closed);

  useEffect(() => {
    let active = true;
    setLoading(true);
    glAllRecords<GLMaster>("periods", "", 1000)
      .then((items) => { if (active) { setPeriods(items); setError(""); } })
      .catch((e: unknown) => { if (active) setError(commandFailure(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).message); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [revision, tr]);
  const reload = useCallback(() => setRevision((value) => value + 1), []);
  const failed = useCallback((e: unknown) => setError(commandFailure(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).message), [tr]);

  const generate = useCallback(async () => {
    if (working) return;
    setInfo("");
    if (!yearCode) { setError(tr("gl_periods_select_year_first", "กรุณาเลือกปีบัญชีก่อน")); return; }
    if (!await confirm({
      title: fillText(tr("gl_periods_generate_confirm", "สร้างงวดรายเดือนที่ยังไม่มีของปีบัญชี {0}?"), yearCode),
      description: tr("gl_periods_generate_detail", "ระบบสร้างงวดละ 1 เดือนตามวันเริ่มต้นและวันสิ้นสุดของปีบัญชี เดือนที่มีงวดอยู่แล้วจะข้ามไป"),
      confirmLabel: tr("gl_periods_generate", "สร้างงวดอัตโนมัติ"),
      tone: "info",
    })) return;
    try {
      setError("");
      const result = await execute({ resource: "periods", action: "generate", targetyear: yearCode });
      toast.success(fillText(tr("gl_periods_generate_done", "สร้างงวดบัญชีเพิ่ม {0} งวดแล้ว"), result?.createdperiods ?? 0));
    } catch (e) {
      // ปีมีงวดครบแล้ว = แจ้งให้ทราบ ไม่ใช่ข้อผิดพลาด
      if (e instanceof GLCommandError && e.code === "periods_complete") setInfo(e.message);
      else failed(e);
    }
    reload();
  }, [working, yearCode, confirm, execute, failed, reload, tr]);

  async function setLock(period: GLMaster, lock: boolean) {
    if (working || !period.id) return;
    setInfo("");
    if (periodReasonMissing(reason)) { setError(tr("gl_periods_reason_required", "กรุณาระบุเหตุผลก่อนล็อกหรือปลดล็อกงวด")); reasonRef.current?.focus(); return; }
    const label = lock ? tr("gl_lock_period_2", "ล็อกงวดบัญชี") : tr("gl_unlock_period_2", "ปลดล็อกงวดบัญชี");
    if (!await confirm({
      title: `${label} ${period.code}?`,
      description: lock ? tr("gl_periods_lock_detail", "ล็อกแล้วจะบันทึก แก้ไข หรือผ่านรายการที่ลงวันที่ในงวดนี้ไม่ได้ จนกว่าจะปลดล็อก") : tr("gl_periods_unlock_detail", "ปลดล็อกแล้วจะบันทึกและแก้ไขรายการที่ลงวันที่ในงวดนี้ได้อีกครั้ง"),
      details: `${tr("gl_reason", "เหตุผล")}: ${reason.trim()}`,
      confirmLabel: label,
      tone: "warning",
    })) return;
    try {
      setError("");
      await execute({ resource: "periods", id: period.id, version: period.version, action: lock ? "lock" : "unlock", reason: reason.trim() });
      toast.success(fillText(tr("gl_success_message", "{0}เรียบร้อยแล้ว"), `${label} ${period.code}`));
      setReason("");
    } catch (e) { failed(e); }
    reload();
  }

  async function remove(period: GLMaster) {
    if (working || !period.id || period.locked) return;
    setInfo("");
    const label = fillText(tr("gl_periods_delete_period", "ลบงวด {0}"), period.code);
    if (!await confirm({ title: `${label}?`, description: tr("gl_periods_delete_detail", "ลบได้เฉพาะงวดที่ยังไม่ล็อกและไม่มีรายการบัญชีในช่วงวันที่ของงวด"), confirmLabel: tr("gl_delete_item", "ลบรายการ"), tone: "danger" })) return;
    try {
      setError("");
      await execute({ resource: "periods", id: period.id, version: period.version, action: "delete", reason: reason.trim() || label });
      toast.success(fillText(tr("gl_success_message", "{0}เรียบร้อยแล้ว"), label));
    } catch (e) { failed(e); }
    reload();
  }

  const dialogOpen = confirmationDialog !== null;
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (periodShortcut(event.key) !== "generate" || event.repeat || dialogOpen || !rootRef.current?.getClientRects().length) return;
      event.preventDefault();
      if (!closed) void generate();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [dialogOpen, closed, generate]);

  const date = (value: string) => (value ? formatAppDate(value, language) : "-");
  return <section ref={rootRef} className={cn(panel, "flex min-h-0 flex-1 flex-col gap-3")} aria-busy={working} data-gl-periods="true">
    <div className="grid items-end gap-3 sm:grid-cols-2 xl:grid-cols-[minmax(12rem,16rem)_minmax(16rem,1fr)_auto]">
      <Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect years={years} value={yearCode} disabled={working} onChange={(value) => { setPickedYear(value); setError(""); setInfo(""); }} /></Field>
      <Field label={tr("gl_reason", "เหตุผล")} hint={tr("gl_periods_reason_hint", "เหตุผลจะถูกเก็บไว้ในประวัติการแก้ไข")}>
        <input ref={reasonRef} className={control} value={reason} maxLength={500} disabled={working} onChange={(event) => setReason(event.target.value)} />
      </Field>
      <div className="flex flex-wrap gap-2 pb-[1.6rem]">
        <Button className={actionClass} disabled={working || !yearCode || closed} onClick={() => void generate()}><CalendarPlus className="mr-1.5 size-4" aria-hidden="true" />{tr("gl_periods_generate", "สร้างงวดอัตโนมัติ")} (F9)</Button>
        <Button className={actionClass} variant="outline" disabled={working} onClick={() => { refs.reload(); reload(); }}><RefreshCw className={cn("mr-1.5 size-4", loading && "animate-spin")} aria-hidden="true" />{tr("gl_reload", "โหลดใหม่")}</Button>
      </div>
    </div>
    {year && <p className="text-[0.95rem] text-muted-foreground">
      {fillText(tr("gl_periods_year_range", "ปีบัญชี {0}: {1} ถึง {2}"), year.code, date(year.startdate), date(year.enddate))}
      {closed && <span className="ml-2 font-semibold text-foreground">· {tr("gl_closed", "ปิดแล้ว")}</span>}
      <span className="ml-2">· {fillText(tr("gl_periods_summary", "{0} งวด · ล็อกแล้ว {1} งวด"), summary.total, summary.locked)}</span>
    </p>}
    <Notice text={error} error />
    {refs.error && <Notice text={refs.error} error />}
    <Notice text={info} />
    {!yearCode && !refs.error && <Notice text={tr("gl_periods_select_year_first", "กรุณาเลือกปีบัญชีก่อน")} />}
    {yearCode && <div className="min-h-0 flex-1 overflow-auto rounded-xl border border-border">
      <table className="w-full min-w-[760px] border-collapse text-[0.95rem]">
        <thead className="sticky top-0 z-10 bg-muted text-left">
          <tr className="[&_th]:px-3 [&_th]:py-2 [&_th]:font-semibold">
            <th className="w-24">{tr("gl_periods_no", "งวดที่")}</th>
            <th>{tr("gl_code", "รหัส")}</th>
            <th>{tr("gl_start_date", "วันเริ่มต้น")}</th>
            <th>{tr("gl_end_date", "วันสิ้นสุด")}</th>
            <th>{tr("gl_status", "สถานะ")}</th>
            <th className="text-right">{tr("gl_manage", "จัดการ")}</th>
          </tr>
        </thead>
        <tbody>
          {!loading && rows.length === 0 && <tr><td colSpan={6} className="px-3 py-6 text-center text-muted-foreground">{tr("gl_periods_empty", "ปีบัญชีนี้ยังไม่มีงวด กด สร้างงวดอัตโนมัติ (F9) เพื่อสร้างงวดรายเดือน")}</td></tr>}
          {rows.map((period, index) => <tr key={period.id ?? period.code} className="border-t border-border [&_td]:px-3 [&_td]:py-2 hover:bg-muted/40" data-period-locked={period.locked ? "true" : "false"}>
            <td className="tabular-nums">{index + 1}</td>
            <td className="font-medium">{period.code}{period.name ? <span className="ml-2 font-normal text-muted-foreground">{period.name}</span> : null}</td>
            <td className="whitespace-nowrap">{date(period.startdate)}</td>
            <td className="whitespace-nowrap">{date(period.enddate)}</td>
            <td>{period.locked
              ? <span className="inline-flex items-center gap-1.5 rounded-full border border-border bg-muted px-2.5 py-1 font-semibold text-foreground"><Lock className="size-4" aria-hidden="true" />{tr("gl_locked", "ล็อกแล้ว")}</span>
              : <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-1 font-semibold text-primary"><LockOpen className="size-4" aria-hidden="true" />{tr("gl_periods_open", "เปิดอยู่")}</span>}</td>
            <td><div className="flex flex-wrap justify-end gap-2">
              <Button className={actionClass} variant="outline" disabled={working || closed} onClick={() => void setLock(period, !period.locked)}>
                {period.locked ? <LockOpen className="mr-1.5 size-4" aria-hidden="true" /> : <Lock className="mr-1.5 size-4" aria-hidden="true" />}
                {period.locked ? tr("gl_unlock_period", "ปลดล็อกงวด") : tr("gl_lock_period", "ล็อกงวด")}
              </Button>
              {!period.locked && <Button className={actionClass} variant="outline" disabled={working || closed} aria-label={fillText(tr("gl_periods_delete_period", "ลบงวด {0}"), period.code)} onClick={() => void remove(period)}>
                <Trash2 className="mr-1.5 size-4 text-destructive" aria-hidden="true" />{tr("gl_delete_item", "ลบรายการ")}
              </Button>}
            </div></td>
          </tr>)}
        </tbody>
      </table>
    </div>}
    {confirmationDialog}
  </section>;
}
