"use client";

import { useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import { Ban, ChevronLeft, ChevronRight, FilePlus2, FolderSearch, Plus, Save, Search, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { useFormShortcuts } from "@/hooks/use-form-shortcuts";
import {
  emptyPrepareForm,
  isPrepareAmountText,
  prepareFormFromDoc,
  prepareLineFromOutstanding,
  preparePayload,
  prepareRequest,
  prepareStatus,
  prepareTotals,
  type ApCreditor,
  type ApOutstandingRow,
  type ApPrepareDoc,
  type ApPrepareForm,
  type ApPrepareFormLine,
  type ApPrepareListRow,
  type PrepareResult,
} from "@/lib/ap-payment-prepare";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { getAuthSession, restoreAuthSession } from "@/lib/client-auth-session";
import { erpDecimalSign, erpMoneyDisplay } from "@/lib/erp-transaction";
import { fillText, formatAmount } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import { useScreenActions } from "@/lib/use-screen-actions";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";

// จอ "บันทึกเตรียมจ่ายชำระหนี้" ตาม Champ (docs/kms/champ/design-2026-10-11/ap-payment-prepare.md §1.1, §4.7):
// หัว → ตารางเอกสารหนี้ → ท้ายจอ; ยอดเงินเป็นข้อความทศนิยมตลอด ยอดจริงคือที่ backend ตอบหลังบันทึก
const ROUTE = "/ap/payment-prepare";

type Tr = (key: string, fallback: string, ...args: (string | number)[]) => string;
type Notice = { tone: "error" | "success" | "info"; text: string };
type DialogKind = "creditor" | "outstanding" | "find" | "cancel" | null;

const enc = encodeURIComponent;
const money = (value: string | undefined) => formatAmount(erpMoneyDisplay(value ?? "0")) || "0.00";

/** วันทำการวันนี้ตามเวลาไทย (YYYY-MM-DD) — เครื่องผู้ใช้อยู่เขตเวลาอื่นก็ได้วันเดียวกับ backend */
function businessToday(): string {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Bangkok" }).format(new Date());
}

interface ApPaymentPrepareScreenProps {
  language?: LanguageCode;
  auth?: AuthSession | null;
  workspace?: WorkspaceSession | null;
}

export function ApPaymentPrepareScreen({ language = "th", auth = null, workspace = null }: ApPaymentPrepareScreenProps) {
  const screenActions = useScreenActions(auth, workspace, ROUTE);
  const [backendUrl, setBackendUrl] = useState(() => getAuthSession()?.backendUrl ?? "");
  useEffect(() => {
    if (backendUrl) return;
    let active = true;
    void restoreAuthSession().then((session) => {
      if (active && session) setBackendUrl(session.backendUrl);
    });
    return () => {
      active = false;
    };
  }, [backendUrl]);
  const dictionary = useBackendLanguage(language, backendUrl || undefined);
  const tr: Tr = useCallback((key, fallback, ...args) => fillText(backendText(dictionary, key, fallback), ...args), [dictionary]);

  const [form, setForm] = useState<ApPrepareForm>(() => emptyPrepareForm(businessToday()));
  const [saved, setSaved] = useState<ApPrepareDoc | null>(null);
  const [dirty, setDirty] = useState(false);
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState<Notice | null>(null);
  const [entryNo, setEntryNo] = useState("");
  const [dialog, setDialog] = useState<DialogKind>(null);
  const [outstandingQuery, setOutstandingQuery] = useState("");
  const [cancelReason, setCancelReason] = useState("");
  const [cancelError, setCancelError] = useState("");
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });

  const status = saved ? prepareStatus(saved) : null;
  // จ่ายบางส่วนยังแก้ได้ (D11) — ปิดเฉพาะใบที่ยกเลิกแล้วหรือจ่ายครบแล้ว
  const locked = status === "cancelled" || status === "paid";
  const canWrite = saved ? screenActions.update : screenActions.create;
  const editable = canWrite && !locked && !busy;
  const preview = useMemo(() => prepareTotals(form.lines, form.discount_amount), [form.lines, form.discount_amount]);
  const showSaved = saved !== null && !dirty;
  const totalAmount = showSaved ? saved.total_amount : preview.total;
  const netAmount = showSaved ? saved.net_amount : preview.net;
  const fail = (message: string, fallbackKey: string, fallback: string) => setNotice({ tone: "error", text: message || tr(fallbackKey, fallback) });

  const edit = (patch: Partial<ApPrepareForm>) => {
    setForm((prev) => ({ ...prev, ...patch }));
    setDirty(true);
  };
  const editLine = (index: number, patch: Partial<ApPrepareFormLine>) => {
    setForm((prev) => ({ ...prev, lines: prev.lines.map((line, i) => (i === index ? { ...line, ...patch } : line)) }));
    setDirty(true);
  };

  const discardOk = async () =>
    !dirty ||
    confirm({
      title: tr("unsaved_changes", "มีการเปลี่ยนแปลงที่ยังไม่ได้บันทึก"),
      description: tr("discard_edits_confirm", "ข้อมูลที่ท่านแก้ไขจะถูกละทิ้ง ยืนยันการยกเลิก?"),
      confirmLabel: tr("confirm_discard", "ยืนยันยกเลิก"),
      tone: "warning",
    });

  const showDoc = (doc: ApPrepareDoc) => {
    setSaved(doc);
    setForm(prepareFormFromDoc(doc));
    setDirty(false);
    setEntryNo("");
  };

  const loadDoc = async (docno: string): Promise<boolean> => {
    setBusy(true);
    const result = await prepareRequest<ApPrepareDoc>(`/${enc(docno)}`, language);
    setBusy(false);
    if (!result.ok) {
      fail(result.message, "ap_prepare_load_failed", "โหลดข้อมูลใบเตรียมจ่ายไม่สำเร็จ กรุณาลองใหม่");
      return false;
    }
    showDoc(result.data);
    return true;
  };

  const openDoc = async (docno: string) => {
    setDialog(null);
    if (!(await discardOk())) return;
    if (await loadDoc(docno)) setNotice(null);
  };

  const startNew = async () => {
    if (!screenActions.create || !(await discardOk())) return;
    setSaved(null);
    setForm(emptyPrepareForm(businessToday()));
    setDirty(false);
    setEntryNo("");
    setNotice(null);
  };

  const save = async () => {
    if (!editable || !dirty) return;
    const badLine = form.lines.find((line) => !isPrepareAmountText(line.payamount) || erpDecimalSign(line.payamount) <= 0);
    if (badLine) {
      setNotice({ tone: "error", text: `${badLine.document_no}: ${tr("gl_err_ap_prepare_amount_invalid", "ยอดตัดจ่ายต้องมากกว่า 0 และมีทศนิยมไม่เกิน 2 ตำแหน่ง")}` });
      return;
    }
    if (form.discount_amount.trim() && !isPrepareAmountText(form.discount_amount)) {
      setNotice({ tone: "error", text: tr("gl_err_ap_prepare_discount_invalid", "ยอดส่วนลดต้องไม่ติดลบและไม่เกินรวมเป็นเงิน") });
      return;
    }
    setBusy(true);
    const body = JSON.stringify(preparePayload(form));
    const result = saved
      ? await prepareRequest<ApPrepareDoc>(`/${enc(saved.docno)}`, language, { method: "PUT", body })
      : await prepareRequest<ApPrepareDoc>("", language, { method: "POST", body });
    setBusy(false);
    if (!result.ok) {
      fail(result.message, "save_failed", "บันทึกไม่สำเร็จ");
      return;
    }
    showDoc(result.data);
    setNotice({ tone: "success", text: tr("ap_prepare_saved", "บันทึกใบเตรียมจ่าย {0} แล้ว", result.data.docno) });
  };

  const openCancel = async () => {
    if (!saved || locked || !screenActions.cancel || !(await discardOk())) return;
    setCancelReason("");
    setCancelError("");
    setDialog("cancel");
  };

  const submitCancel = async () => {
    if (!saved) return;
    const reason = cancelReason.trim();
    if (!reason) {
      setCancelError(tr("gl_err_ap_prepare_cancel_reason_required", "กรุณากรอกเหตุผลการยกเลิก"));
      return;
    }
    setBusy(true);
    const result = await prepareRequest<unknown>(`/${enc(saved.docno)}/cancel`, language, {
      method: "POST",
      body: JSON.stringify({ reason, version: saved.version }),
    });
    setBusy(false);
    if (!result.ok) {
      setCancelError(result.message || tr("ap_prepare_load_failed", "โหลดข้อมูลใบเตรียมจ่ายไม่สำเร็จ กรุณาลองใหม่"));
      return;
    }
    setDialog(null);
    if (await loadDoc(saved.docno)) setNotice({ tone: "success", text: tr("ap_prepare_cancelled_notice", "ยกเลิกใบเตรียมจ่าย {0} แล้ว", saved.docno) });
  };

  // ก่อนหน้า = ใบที่เก่ากว่า, ถัดไป = ใบที่ใหม่กว่า (รายการเรียงวันที่จ่าย/เลขที่ จากใหม่ไปเก่า)
  const step = async (direction: "prev" | "next") => {
    const result = await prepareRequest<ApPrepareListRow[] | null>("/list?limit=200", language);
    if (!result.ok) {
      fail(result.message, "ap_prepare_load_failed", "โหลดข้อมูลใบเตรียมจ่ายไม่สำเร็จ กรุณาลองใหม่");
      return;
    }
    const rows = result.data ?? [];
    const at = saved ? rows.findIndex((row) => row.docno === saved.docno) : -1;
    const target = at < 0 ? rows[0] : rows[direction === "prev" ? at + 1 : at - 1];
    if (!target) {
      setNotice({ tone: "info", text: tr("ap_prepare_no_more", "ไม่มีใบเตรียมจ่ายถัดไปแล้ว") });
      return;
    }
    await openDoc(target.docno);
  };

  const pickCreditor = async (creditor: ApCreditor) => {
    setDialog(null);
    if (creditor.partner_code === form.partner_code && form.partner_name) return;
    if (
      form.lines.length > 0 &&
      !(await confirm({
        title: tr("ap_prepare_change_creditor", "เปลี่ยนเจ้าหนี้?"),
        description: tr("ap_prepare_change_creditor_explain", "รายการเอกสารของเจ้าหนี้เดิมทั้งหมดจะถูกล้างออกจากใบนี้"),
        tone: "warning",
      }))
    ) {
      return;
    }
    edit({ partner_code: creditor.partner_code, partner_name: creditor.name, lines: [] });
  };

  // Enter ที่ช่องรหัสเจ้าหนี้: ตรงรหัสพอดีเลือกเลย ไม่ตรงเปิดหน้าค้นหา
  const lookupCreditor = async () => {
    const code = form.partner_code.trim();
    if (code) {
      const result = await prepareRequest<ApCreditor[] | null>(`/creditors?q=${enc(code)}&limit=20`, language);
      const exact = result.ok ? (result.data ?? []).find((row) => row.partner_code === code) : undefined;
      if (exact) {
        await pickCreditor(exact);
        return;
      }
    }
    setDialog("creditor");
  };

  const outstandingPath = (q: string) => {
    const params = new URLSearchParams({ partner: form.partner_code.trim(), todate: form.paydate, limit: "200" });
    if (q) params.set("q", q);
    return `${saved ? `/${enc(saved.docno)}/outstanding` : "/outstanding"}?${params.toString()}`;
  };

  const addOutstanding = (row: ApOutstandingRow) => {
    setDialog(null);
    if (form.lines.some((line) => line.debt_document_id === row.debt_document_id)) {
      setNotice({ tone: "error", text: tr("gl_err_ap_prepare_line_duplicate", "เอกสาร {0} ถูกเลือกซ้ำ กรุณาลบบรรทัดที่ซ้ำ", row.document_no) });
      return;
    }
    edit({ lines: [...form.lines, prepareLineFromOutstanding(row)] });
    setEntryNo("");
  };

  const openOutstanding = (q: string) => {
    if (!form.partner_code.trim() || !form.partner_name) {
      setNotice({ tone: "info", text: tr("ap_prepare_pick_creditor_first", "กรุณาเลือกรหัสเจ้าหนี้ก่อนเลือกเอกสาร") });
      return;
    }
    setOutstandingQuery(q);
    setDialog("outstanding");
  };

  // Enter ที่ช่องเลขที่เอกสารแถวว่าง: ตรงเลขที่พอดีใบเดียวเพิ่มเลย ไม่อย่างนั้นเปิดรายการเอกสารค้างจ่าย
  const lookupEntry = async () => {
    const no = entryNo.trim();
    if (!no || !form.partner_name) {
      openOutstanding(no);
      return;
    }
    const result = await prepareRequest<ApOutstandingRow[] | null>(outstandingPath(no), language);
    const exact = result.ok ? (result.data ?? []).filter((row) => row.document_no === no) : [];
    if (exact.length === 1) addOutstanding(exact[0]);
    else openOutstanding(no);
  };

  const removeLine = (index: number) => {
    setForm((prev) => ({ ...prev, lines: prev.lines.filter((_, i) => i !== index) }));
    setDirty(true);
  };

  useFormShortcuts({
    onSave: save,
    onNew: screenActions.create ? startNew : undefined,
    onCancel: dialog ? () => setDialog(null) : undefined,
    canSave: editable && dirty,
    disabled: busy,
  });

  const statusLabel =
    status === "cancelled"
      ? tr("cancelled", "ยกเลิกแล้ว")
      : status === "paid"
        ? tr("ap_prepare_status_paid", "จ่ายหมดแล้ว")
        : status === "unpaid"
          ? tr("ap_prepare_status_unpaid", "ยังไม่จ่าย")
          : tr("create_new", "สร้างใหม่");
  const lockedHint =
    status === "cancelled"
      ? tr("gl_err_ap_prepare_cancelled_locked", "ใบเตรียมจ่ายนี้ถูกยกเลิกแล้ว") + (saved?.cancel_reason ? ` — ${saved.cancel_reason}` : "")
      : status === "paid"
        ? tr("gl_err_ap_prepare_paid_locked", "ใบเตรียมจ่ายนี้จ่ายครบแล้ว แก้ไขหรือยกเลิกไม่ได้ ให้ยกเลิกใบจ่ายชำระหนี้ก่อน")
        : "";
  const billTypeLabel = (line: Pick<ApPrepareFormLine, "billtype" | "billtype_key">) => tr(line.billtype_key, String(line.billtype));
  const label = "mb-1 block text-[0.85rem] font-semibold leading-[1.45] text-muted-foreground";
  const control = "h-10 text-[0.95rem]";
  const toolbarButton = "min-h-[2.6em] gap-1.5 px-3 text-[0.9rem]";

  return (
    <div className="flex flex-col gap-3 p-3 text-[0.95rem] leading-[1.5]">
      <div className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-border/70 bg-card p-3 shadow-[0_8px_24px_-18px_color-mix(in_srgb,var(--primary)_45%,transparent)]">
        <div className="flex min-w-0 items-center gap-2">
          <h1 className="text-lg font-bold text-foreground">{tr("ap_payment_prepare", "บันทึกเตรียมจ่ายชำระหนี้")}</h1>
          <span className="rounded-full border border-border bg-muted px-2.5 py-0.5 text-[0.85rem] font-semibold text-foreground">
            {saved ? `${saved.docno} · ${statusLabel}` : statusLabel}
          </span>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="outline" className={toolbarButton} onClick={() => void startNew()} disabled={busy || !screenActions.create}>
            <FilePlus2 className="size-4" aria-hidden />
            {tr("create_new", "สร้างใหม่")}
          </Button>
          <Button className={`${toolbarButton} font-semibold`} onClick={() => void save()} disabled={!editable || !dirty}>
            <Save className="size-4" aria-hidden />
            {tr("save", "บันทึก")}
            <kbd className="rounded border border-primary-foreground/30 px-1 text-[0.8rem]">F11</kbd>
          </Button>
          <Button
            variant="outline"
            className={`${toolbarButton} text-destructive`}
            onClick={() => void openCancel()}
            disabled={busy || !saved || locked || !screenActions.cancel}
          >
            <Ban className="size-4" aria-hidden />
            {tr("ap_prepare_cancel_doc", "ยกเลิกใบเตรียมจ่าย")}
          </Button>
          <Button variant="outline" className={toolbarButton} onClick={() => setDialog("find")} disabled={busy}>
            <FolderSearch className="size-4" aria-hidden />
            {tr("ap_prepare_find", "ค้นหาใบเตรียมจ่าย")}
          </Button>
          <Button variant="outline" className={toolbarButton} onClick={() => void step("prev")} disabled={busy}>
            <ChevronLeft className="size-4" aria-hidden />
            {tr("previous", "ก่อนหน้า")}
          </Button>
          <Button variant="outline" className={toolbarButton} onClick={() => void step("next")} disabled={busy}>
            {tr("next", "ถัดไป")}
            <ChevronRight className="size-4" aria-hidden />
          </Button>
        </div>
      </div>

      {notice || lockedHint ? (
        <div
          role={notice?.tone === "error" ? "alert" : "status"}
          className={`rounded-xl border px-3 py-2 text-[0.95rem] [overflow-wrap:anywhere] ${
            notice?.tone === "error"
              ? "border-destructive/40 bg-destructive/10 text-destructive"
              : "border-border bg-muted/60 text-foreground"
          }`}
        >
          {notice?.text ?? lockedHint}
        </div>
      ) : null}

      <section className="grid gap-3 rounded-2xl border border-border/70 bg-card p-3 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label className={label} htmlFor="ap-prepare-docno">
            {tr("docno", "เลขที่เอกสาร")} <span className="text-destructive" aria-hidden>*</span>
          </label>
          <Input
            id="ap-prepare-docno"
            className={`${control} font-mono`}
            value={form.docno}
            maxLength={50}
            aria-required
            autoComplete="off"
            readOnly={Boolean(saved)}
            disabled={!editable && !saved}
            onChange={(e) => edit({ docno: e.target.value })}
          />
        </div>
        <div>
          <label className={label} htmlFor="ap-prepare-paydate">
            {tr("ap_prepare_pay_date", "วันที่จ่ายชำระหนี้")}
          </label>
          <Input
            id="ap-prepare-paydate"
            type="date"
            className={control}
            value={form.paydate}
            disabled={!editable}
            onChange={(e) => edit({ paydate: e.target.value })}
          />
        </div>
        <div className="sm:col-span-2">
          <label className={label} htmlFor="ap-prepare-creditor">
            {tr("creditor_code", "รหัสเจ้าหนี้")}
          </label>
          <div className="flex gap-2">
            <Input
              id="ap-prepare-creditor"
              className={`${control} w-40 shrink-0 font-mono`}
              value={form.partner_code}
              maxLength={50}
              autoComplete="off"
              readOnly={Boolean(saved) || form.lines.length > 0}
              disabled={!editable}
              onChange={(e) => edit({ partner_code: e.target.value, partner_name: "" })}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  void lookupCreditor();
                }
              }}
            />
            <Button
              variant="outline"
              className="min-h-[2.6em] shrink-0 gap-1.5 px-3"
              onClick={() => setDialog("creditor")}
              disabled={!editable || Boolean(saved)}
              aria-label={tr("search", "ค้นหา")}
              title={tr("search", "ค้นหา")}
            >
              <Search className="size-4" aria-hidden />
            </Button>
            <Input className={`${control} min-w-0 flex-1`} value={form.partner_name} readOnly tabIndex={-1} aria-label={tr("creditor_name", "ชื่อเจ้าหนี้")} />
          </div>
        </div>
        <div className="sm:col-span-2 lg:col-span-4">
          <label className={label} htmlFor="ap-prepare-remark">
            {tr("remark", "หมายเหตุ")}
          </label>
          <Textarea
            id="ap-prepare-remark"
            rows={2}
            maxLength={500}
            className="text-[0.95rem]"
            value={form.remark}
            disabled={!editable}
            onChange={(e) => edit({ remark: e.target.value })}
          />
        </div>
      </section>

      <section className="rounded-2xl border border-border/70 bg-card p-3">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[56rem] border-collapse text-[0.9rem]">
            <thead>
              <tr className="border-b border-border text-left text-[0.85rem] text-muted-foreground">
                <th className="w-10 px-2 py-1.5 text-center">#</th>
                <th className="px-2 py-1.5">{tr("docno", "เลขที่เอกสาร")}</th>
                <th className="w-32 px-2 py-1.5">{tr("ap_prepare_dated", "ลงวันที่")}</th>
                <th className="px-2 py-1.5">{tr("ap_prepare_day_explain", "รายวัน/คำอธิบาย")}</th>
                <th className="w-40 px-2 py-1.5 text-right">{tr("balance_amount", "มูลค่าคงเหลือ")}</th>
                <th className="w-44 px-2 py-1.5 text-right">{tr("ap_prepare_cut_pay_amount", "ยอดตัดจ่าย")}</th>
                <th className="w-12 px-2 py-1.5 text-center">{tr("ap_prepare_col_s", "S")}</th>
                <th className="w-12 px-2 py-1.5" aria-label={tr("ap_prepare_remove_line", "ลบบรรทัด")} />
              </tr>
            </thead>
            <tbody>
              {form.lines.map((line, index) => {
                const paid = erpDecimalSign(line.used_amount) > 0;
                const amountOk = isPrepareAmountText(line.payamount) && erpDecimalSign(line.payamount) > 0;
                return (
                  <tr key={line.debt_document_id} className="border-b border-border/60">
                    <td className="px-2 py-1 text-center text-muted-foreground">{index + 1}</td>
                    <td className="px-2 py-1 font-mono">{line.document_no}</td>
                    <td className="px-2 py-1 font-mono text-muted-foreground">{line.document_date}</td>
                    <td className="px-2 py-1">
                      <Input
                        className="h-9 text-[0.9rem]"
                        value={line.description}
                        maxLength={500}
                        disabled={!editable}
                        aria-label={tr("ap_prepare_day_explain", "รายวัน/คำอธิบาย")}
                        onChange={(e) => editLine(index, { description: e.target.value })}
                      />
                    </td>
                    <td className="px-2 py-1 text-right font-mono text-muted-foreground">
                      {line.sign < 0 ? "-" : ""}
                      {money(line.invbalance)}
                    </td>
                    <td className="px-2 py-1">
                      <Input
                        className="h-9 text-right font-mono text-[0.9rem]"
                        inputMode="decimal"
                        autoComplete="off"
                        maxLength={19}
                        value={line.payamount}
                        disabled={!editable}
                        aria-invalid={!amountOk}
                        aria-label={`${tr("ap_prepare_cut_pay_amount", "ยอดตัดจ่าย")} ${line.document_no}`}
                        title={paid ? tr("ap_prepare_paid_hint", "ใบจ่ายใช้ไปแล้ว {0} บาท ลดยอดต่ำกว่านี้ไม่ได้", money(line.used_amount)) : undefined}
                        onChange={(e) => editLine(index, { payamount: e.target.value.replace(/,/g, "") })}
                      />
                    </td>
                    <td className="px-2 py-1 text-center font-mono" title={billTypeLabel(line)}>
                      {line.billtype}
                    </td>
                    <td className="px-2 py-1 text-center">
                      <Button
                        variant="ghost"
                        className="size-9 p-0 text-destructive"
                        disabled={!editable || paid}
                        onClick={() => removeLine(index)}
                        aria-label={`${tr("ap_prepare_remove_line", "ลบบรรทัด")} ${line.document_no}`}
                        title={paid ? tr("ap_prepare_paid_hint", "ใบจ่ายใช้ไปแล้ว {0} บาท ลดยอดต่ำกว่านี้ไม่ได้", money(line.used_amount)) : tr("ap_prepare_remove_line", "ลบบรรทัด")}
                      >
                        <Trash2 className="size-4" aria-hidden />
                      </Button>
                    </td>
                  </tr>
                );
              })}
              {editable ? (
                <tr>
                  <td className="px-2 py-1 text-center text-muted-foreground">{form.lines.length + 1}</td>
                  <td className="px-2 py-1" colSpan={7}>
                    <div className="flex max-w-xl gap-2">
                      <Input
                        className="h-9 font-mono text-[0.9rem]"
                        value={entryNo}
                        maxLength={50}
                        autoComplete="off"
                        placeholder={tr("ap_prepare_entry_hint", "พิมพ์เลขที่เอกสารแล้วกด Enter หรือกดเลือกเอกสาร")}
                        aria-label={tr("docno", "เลขที่เอกสาร")}
                        onChange={(e) => setEntryNo(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            e.preventDefault();
                            void lookupEntry();
                          }
                        }}
                      />
                      <Button variant="outline" className="min-h-[2.6em] shrink-0 gap-1.5 px-3 text-[0.9rem]" onClick={() => openOutstanding(entryNo.trim())}>
                        <Plus className="size-4" aria-hidden />
                        {tr("ap_prepare_outstanding_lookup", "เอกสารค้างจ่าย")}
                      </Button>
                    </div>
                  </td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </section>

      <section className="grid gap-3 rounded-2xl border border-border/70 bg-card p-3 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label className={label} htmlFor="ap-prepare-project">
            {tr("project", "โครงการ")}
          </label>
          <Input id="ap-prepare-project" className={control} value={form.projectcode} maxLength={50} disabled={!editable} onChange={(e) => edit({ projectcode: e.target.value })} />
        </div>
        <div>
          <label className={label} htmlFor="ap-prepare-allocate">
            {tr("ap_prepare_allocate", "การจัดสรร")}
          </label>
          {/* D8: แสดงแต่ปิดไว้จนกว่าทะเบียนการจัดสรรจะเสร็จ (เก็บค่าว่าง) */}
          <Input id="ap-prepare-allocate" className={control} value="" disabled readOnly />
        </div>
        <div>
          <label className={label} htmlFor="ap-prepare-billgroup">
            {tr("ap_prepare_bill_group", "กลุ่มของบิล")}
          </label>
          <Input id="ap-prepare-billgroup" className={control} value="" disabled readOnly />
        </div>
        <div>
          <label className={label} htmlFor="ap-prepare-department">
            {tr("department_code", "รหัสแผนก")}
          </label>
          <Input id="ap-prepare-department" className={control} value={form.departmentcode} maxLength={50} disabled={!editable} onChange={(e) => edit({ departmentcode: e.target.value })} />
        </div>
        <AmountField label={tr("ap_prepare_exchange_profit", "กำไร(ขาดทุน)")} value={money(saved?.exchange_profit)} labelClass={label} />
        <AmountField label={tr("ap_prepare_total_home", "รวมเป็นเงิน(บาท)")} value={money(showSaved ? saved.total_home_amount : preview.total)} labelClass={label} />
        <AmountField label={tr("ap_prepare_total", "รวมเป็นเงิน")} value={money(totalAmount)} labelClass={label} />
        <div>
          <label className={label} htmlFor="ap-prepare-discount">
            {tr("doc_discount_amount", "ยอดส่วนลด")}
          </label>
          <Input
            id="ap-prepare-discount"
            className={`${control} text-right font-mono`}
            inputMode="decimal"
            autoComplete="off"
            maxLength={19}
            placeholder="0.00"
            value={form.discount_amount}
            disabled={!editable}
            aria-invalid={form.discount_amount.trim() !== "" && !isPrepareAmountText(form.discount_amount)}
            onChange={(e) => edit({ discount_amount: e.target.value.replace(/,/g, "") })}
          />
        </div>
        <div className="sm:col-span-2 lg:col-span-4 flex justify-end">
          <div className="flex min-w-[18rem] items-baseline justify-between gap-4 rounded-xl border border-primary/30 bg-[color-mix(in_srgb,var(--primary)_8%,transparent)] px-4 py-2">
            <span className="font-semibold text-foreground">{tr("ap_prepare_net_pay", "ยอดจ่าย")}</span>
            <span className="font-mono text-xl font-bold text-primary">{money(netAmount)}</span>
          </div>
        </div>
      </section>

      {dialog === "creditor" ? (
        <PickerDialog<ApCreditor>
          title={tr("creditor_code", "รหัสเจ้าหนี้")}
          tr={tr}
          initialQuery={form.partner_code.trim()}
          load={(q) => prepareRequest<ApCreditor[] | null>(`/creditors?limit=100${q ? `&q=${enc(q)}` : ""}`, language)}
          rowKey={(row) => row.partner_code}
          columns={[
            { label: tr("creditor_code", "รหัสเจ้าหนี้"), render: (row) => <span className="font-mono">{row.partner_code}</span> },
            { label: tr("creditor_name", "ชื่อเจ้าหนี้"), render: (row) => row.name },
            { label: tr("credit_day", "จำนวนวันเครดิต"), render: (row) => row.credit_days, align: "right" },
          ]}
          onPick={(row) => void pickCreditor(row)}
          onClose={() => setDialog(null)}
        />
      ) : null}

      {dialog === "outstanding" ? (
        <PickerDialog<ApOutstandingRow>
          title={`${tr("ap_prepare_outstanding_lookup", "เอกสารค้างจ่าย")} — ${form.partner_code} ${form.partner_name}`}
          tr={tr}
          initialQuery={outstandingQuery}
          load={async (q) => {
            const result = await prepareRequest<ApOutstandingRow[] | null>(outstandingPath(q), language);
            if (!result.ok) return result;
            const taken = new Set(form.lines.map((line) => line.debt_document_id));
            return { ...result, data: (result.data ?? []).filter((row) => !taken.has(row.debt_document_id)) };
          }}
          emptyText={(q) => (q ? tr("no_data", "ไม่มีข้อมูล") : tr("erp_ap_prepare_no_posted_debt", "เจ้าหนี้รายนี้ไม่มีหนี้ค้างที่ผ่านบัญชีแล้ว"))}
          rowKey={(row) => row.debt_document_id}
          columns={[
            { label: tr("docno", "เลขที่เอกสาร"), render: (row) => <span className="font-mono">{row.document_no}</span> },
            { label: tr("ap_prepare_dated", "ลงวันที่"), render: (row) => <span className="font-mono">{row.document_date}</span> },
            { label: tr("gl_detail_documents_due_date", "วันครบกำหนด"), render: (row) => <span className="font-mono">{row.due_date || "-"}</span> },
            { label: tr("ap_prepare_col_s", "S"), render: (row) => tr(row.billtype_key, String(row.billtype)) },
            {
              label: tr("ap_prepare_available", "ยอดที่จ่ายได้"),
              render: (row) => (
                <span className="font-mono">
                  {row.sign < 0 ? "-" : ""}
                  {money(row.available_amount)}
                </span>
              ),
              align: "right",
            },
          ]}
          onPick={addOutstanding}
          onClose={() => setDialog(null)}
        />
      ) : null}

      {dialog === "find" ? (
        <PickerDialog<ApPrepareListRow>
          title={tr("ap_prepare_find", "ค้นหาใบเตรียมจ่าย")}
          tr={tr}
          initialQuery=""
          load={(q) => prepareRequest<ApPrepareListRow[] | null>(`/list?limit=100${q ? `&q=${enc(q)}` : ""}`, language)}
          rowKey={(row) => row.docno}
          columns={[
            { label: tr("docno", "เลขที่เอกสาร"), render: (row) => <span className="font-mono">{row.docno}</span> },
            { label: tr("ap_prepare_pay_date", "วันที่จ่ายชำระหนี้"), render: (row) => <span className="font-mono">{row.paydate}</span> },
            { label: tr("cp_vendor", "เจ้าหนี้"), render: (row) => `${row.partner_code} ${row.partner_name}` },
            { label: tr("ap_prepare_net_pay", "ยอดจ่าย"), render: (row) => <span className="font-mono">{money(row.net_amount)}</span>, align: "right" },
            {
              label: tr("status", "สถานะ"),
              render: (row) => {
                const rowStatus = prepareStatus(row);
                if (rowStatus === "cancelled") return tr("cancelled", "ยกเลิกแล้ว");
                return rowStatus === "paid" ? tr("ap_prepare_status_paid", "จ่ายหมดแล้ว") : tr("ap_prepare_status_unpaid", "ยังไม่จ่าย");
              },
            },
          ]}
          onPick={(row) => void openDoc(row.docno)}
          onClose={() => setDialog(null)}
        />
      ) : null}

      {dialog === "cancel" && saved ? (
        <DialogShell title={tr("ap_prepare_cancel_doc", "ยกเลิกใบเตรียมจ่าย") + ` ${saved.docno}`} closeLabel={tr("close", "ปิด")} onClose={() => setDialog(null)}>
          <div className="flex flex-col gap-3 p-4">
            <p className="text-[0.95rem] text-foreground [overflow-wrap:anywhere]">
              {tr("ap_prepare_cancel_explain", "การยกเลิกจะคืนยอดที่จองไว้ของเอกสารหนี้ทุกบรรทัดให้นำไปเตรียมจ่ายใหม่ได้ และแก้ไขใบนี้ไม่ได้อีก")}
            </p>
            <label className={label} htmlFor="ap-prepare-cancel-reason">
              {tr("ap_prepare_cancel_reason", "เหตุผลการยกเลิก")} <span className="text-destructive" aria-hidden>*</span>
            </label>
            <Textarea
              id="ap-prepare-cancel-reason"
              rows={3}
              maxLength={500}
              autoFocus
              className="text-[0.95rem]"
              value={cancelReason}
              aria-required
              aria-invalid={cancelError !== ""}
              onChange={(e) => {
                setCancelReason(e.target.value);
                setCancelError("");
              }}
            />
            {cancelError ? (
              <p role="alert" className="text-[0.9rem] text-destructive [overflow-wrap:anywhere]">
                {cancelError}
              </p>
            ) : null}
            <div className="flex justify-end gap-2">
              <Button variant="outline" className="min-h-[2.6em] px-4" onClick={() => setDialog(null)}>
                {tr("close", "ปิด")}
              </Button>
              <Button variant="destructive" className="min-h-[2.6em] gap-1.5 px-4" onClick={() => void submitCancel()} disabled={busy}>
                <Ban className="size-4" aria-hidden />
                {tr("ap_prepare_cancel_doc", "ยกเลิกใบเตรียมจ่าย")}
              </Button>
            </div>
          </div>
        </DialogShell>
      ) : null}

      {confirmationDialog}
    </div>
  );
}

/**
 * Ctrl+F3 บนใบจ่ายชำระหนี้ (design §4.7): เลือกใบเตรียมจ่ายที่ยังจ่ายไม่หมด — ระบุเจ้าหนี้บนใบจ่ายแล้วแสดงเฉพาะของเจ้าหนี้รายนั้น
 */
export function ApPreparePullDialog({
  language,
  dictionary,
  partnerCode,
  onPick,
  onClose,
}: {
  language: LanguageCode;
  dictionary: BackendLanguageDictionary;
  partnerCode: string;
  onPick: (row: ApPrepareListRow) => void;
  onClose: () => void;
}) {
  const tr: Tr = (key, fallback, ...args) => fillText(backendText(dictionary, key, fallback), ...args);
  const partner = partnerCode.trim();
  return (
    <PickerDialog<ApPrepareListRow>
      title={tr("ap_pay_pull_prepare", "ดึงใบเตรียมจ่าย")}
      tr={tr}
      initialQuery=""
      load={(q) => {
        const params = new URLSearchParams({ status: "open", limit: "100" });
        if (partner) params.set("partner", partner);
        if (q) params.set("q", q);
        return prepareRequest<ApPrepareListRow[] | null>(`/list?${params.toString()}`, language);
      }}
      rowKey={(row) => row.docno}
      columns={[
        { label: tr("docno", "เลขที่เอกสาร"), render: (row) => <span className="font-mono">{row.docno}</span> },
        { label: tr("ap_prepare_pay_date", "วันที่จ่ายชำระหนี้"), render: (row) => <span className="font-mono">{row.paydate}</span> },
        { label: tr("cp_vendor", "เจ้าหนี้"), render: (row) => `${row.partner_code} ${row.partner_name}` },
        { label: tr("ap_prepare_net_pay", "ยอดจ่าย"), render: (row) => <span className="font-mono">{money(row.net_amount)}</span>, align: "right" },
      ]}
      onPick={onPick}
      onClose={onClose}
    />
  );
}

function AmountField({ label, value, labelClass }: { label: string; value: string; labelClass: string }) {
  return (
    <div>
      <span className={labelClass}>{label}</span>
      <div className="flex h-10 items-center justify-end rounded-md border border-border bg-muted/50 px-3 font-mono text-[0.95rem] text-foreground" aria-label={label}>
        {value}
      </div>
    </div>
  );
}

function DialogShell({ title, closeLabel, onClose, children }: { title: string; closeLabel: string; onClose: () => void; children: ReactNode }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/30 p-4" onClick={onClose}>
      <div
        role="dialog"
        aria-modal="true"
        aria-label={title}
        className="flex max-h-[85vh] w-[min(56rem,94vw)] flex-col rounded-2xl border border-border bg-background shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between gap-2 border-b border-border px-4 py-3">
          <h2 className="text-base font-bold text-foreground [overflow-wrap:anywhere]">{title}</h2>
          <Button variant="ghost" className="size-10 p-0" onClick={onClose} aria-label={closeLabel} title={closeLabel}>
            <X className="size-5" aria-hidden />
          </Button>
        </div>
        {children}
      </div>
    </div>
  );
}

type PickerColumn<T> = { label: string; render: (row: T) => ReactNode; align?: "right" };

function PickerDialog<T>({
  title,
  tr,
  initialQuery,
  load,
  rowKey,
  columns,
  emptyText,
  onPick,
  onClose,
}: {
  title: string;
  tr: Tr;
  initialQuery: string;
  load: (q: string) => Promise<PrepareResult<T[] | null>>;
  rowKey: (row: T) => string;
  columns: PickerColumn<T>[];
  emptyText?: (q: string) => string;
  onPick: (row: T) => void;
  onClose: () => void;
}) {
  const [query, setQuery] = useState(initialQuery);
  const [rows, setRows] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searched, setSearched] = useState(initialQuery);

  const run = useCallback(
    async (q: string) => {
      setLoading(true);
      setError("");
      setSearched(q);
      const result = await load(q);
      setLoading(false);
      if (!result.ok) {
        setRows([]);
        setError(result.message || tr("ap_prepare_load_failed", "โหลดข้อมูลใบเตรียมจ่ายไม่สำเร็จ กรุณาลองใหม่"));
        return;
      }
      setRows(result.data ?? []);
    },
    // load เปลี่ยนทุก render ของจอแม่ — โหลดครั้งแรกครั้งเดียว ที่เหลือโหลดตามปุ่มค้นหา
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );
  useEffect(() => {
    void run(initialQuery);
  }, [run, initialQuery]);

  return (
    <DialogShell title={title} closeLabel={tr("close", "ปิด")} onClose={onClose}>
      <form
        className="flex gap-2 border-b border-border px-4 py-3"
        onSubmit={(e) => {
          e.preventDefault();
          void run(query.trim());
        }}
      >
        <Input className="h-10 text-[0.95rem]" value={query} autoFocus maxLength={100} aria-label={tr("search", "ค้นหา")} onChange={(e) => setQuery(e.target.value)} />
        <Button type="submit" variant="outline" className="min-h-[2.6em] shrink-0 gap-1.5 px-4">
          <Search className="size-4" aria-hidden />
          {tr("search", "ค้นหา")}
        </Button>
      </form>
      <div className="min-h-0 flex-1 overflow-y-auto px-2 py-2">
        {loading ? (
          <p className="p-4 text-center text-muted-foreground">{tr("loading", "กำลังโหลดข้อมูล...")}</p>
        ) : error ? (
          <p role="alert" className="p-4 text-center text-destructive [overflow-wrap:anywhere]">
            {error}
          </p>
        ) : rows.length === 0 ? (
          <p className="p-4 text-center text-muted-foreground [overflow-wrap:anywhere]">{emptyText ? emptyText(searched) : tr("no_data", "ไม่มีข้อมูล")}</p>
        ) : (
          <table className="w-full border-collapse text-[0.9rem]">
            <thead>
              <tr className="border-b border-border text-left text-[0.85rem] text-muted-foreground">
                {columns.map((column) => (
                  <th key={column.label} className={`px-2 py-1.5 ${column.align === "right" ? "text-right" : ""}`}>
                    {column.label}
                  </th>
                ))}
                <th className="w-24 px-2 py-1.5" />
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={rowKey(row)} className="cursor-pointer border-b border-border/60 hover:bg-muted/60" onClick={() => onPick(row)}>
                  {columns.map((column) => (
                    <td key={column.label} className={`px-2 py-1.5 ${column.align === "right" ? "text-right" : ""}`}>
                      {column.render(row)}
                    </td>
                  ))}
                  <td className="px-2 py-1 text-right">
                    <Button
                      variant="outline"
                      className="min-h-[2.4em] px-3 text-[0.9rem]"
                      onClick={(e) => {
                        e.stopPropagation();
                        onPick(row);
                      }}
                    >
                      {tr("select", "เลือก")}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </DialogShell>
  );
}
