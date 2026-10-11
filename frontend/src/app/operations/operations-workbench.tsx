"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  useBackendDictionary,
  useBackendText,
  type BackendTextFn,
} from "@/components/backend-text-provider";
import {
  getOperationsConfig,
  operationsText,
  isApprovalApiReady,
  getApprovalOperationMeta,
  fetchPendingApprovals,
  submitApprovalAction,
  submitBatchApprove,
  submitBatchUnapprove,
  submitBatchCancel,
  type PendingApprovalDoc,
} from "@/lib/erp-operations";
import { getAuthSession } from "@/lib/client-auth-session";
import type { BatchApproveItemResult } from "@/lib/erp-transaction";
import type { LanguageCode } from "@/lib/i18n";
import { ErpDocEventsDialog } from "@/components/erp-doc-events-dialog";
import { DeliveryDateOverhaulScreen } from "@/components/delivery-date-overhaul";
import { PurchaseReceiptAdjustmentScreen } from "@/components/purchase-receipt-adjustment";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import {
  CheckCircle,
  XCircle,
  Clock,
  Search,
  Package,
  Boxes,
  ShieldCheck,
  AlertCircle,
  Loader2,
  RotateCw,
  Undo2,
  Ban,
} from "lucide-react";

interface OperationsWorkbenchProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
}

const MESSAGES: Record<string, { key: string; th: string }> = {
  approval_not_available: {
    key: "ops_msg_this_screen_is_not_connected",
    th: "จอนี้ยังไม่เชื่อมกับระบบงานจริง — อยู่ระหว่างเปิดใช้งาน API",
  },
  holding_required: { key: "holding_required", th: "ยังไม่ได้เลือกกิจการ" },
  docno_required: { key: "ops_msg_document_number_is_missing", th: "ไม่พบเลขที่เอกสาร" },
  unauthorized: { key: "ops_msg_you_do_not_have_permission", th: "ไม่มีสิทธิ์ดำเนินการรายการนี้" },
  erp_err_action_forbidden: { key: "ops_msg_you_do_not_have_permission", th: "ไม่มีสิทธิ์ดำเนินการรายการนี้" },
  module_not_available: {
    key: "module_not_available",
    th: "จอนี้ยังไม่เปิดใช้งาน ระบบยังไม่รองรับเอกสารชนิดนี้ กรุณาติดต่อผู้ดูแลระบบ",
  },
  load_failed: { key: "load_data_failed", th: "โหลดข้อมูลไม่สำเร็จ" },
  connection_error: { key: "ops_msg_unable_to_connect_to_the", th: "เชื่อมต่อระบบไม่ได้" },
  action_failed: { key: "ops_msg_action_failed_please_try_again", th: "ดำเนินการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง" },
  approve_success: { key: "erp_msg_approved", th: "อนุมัติเอกสารเรียบร้อยแล้ว" },
  unapprove_success: { key: "erp_msg_unapproved", th: "ยกเลิกการอนุมัติเอกสารเรียบร้อยแล้ว" },
  cancel_success: { key: "erp_msg_cancelled", th: "ยกเลิกเอกสารเรียบร้อยแล้ว" },
  reject_success: { key: "reject_success", th: "บันทึกการไม่อนุมัติเรียบร้อยแล้ว" },
  batch_partial: {
    key: "ops_batch_partial_success",
    th: "ดำเนินการสำเร็จบางรายการ กรุณาตรวจสอบรายการที่ไม่สำเร็จด้านล่าง",
  },
};

// backend คืนได้สูงสุด 200 แถวต่อหน้า (ค่าเริ่มต้น 50)
const PAGE_SIZE = 50;

function messageText(key: string, tr: BackendTextFn): string {
  // คีย์ที่ไม่รู้จักห้ามหลุดเป็นรหัสให้ผู้ใช้เห็น
  const message = MESSAGES[key] ?? MESSAGES.action_failed;
  return tr(message.key, message.th);
}

export function OperationsWorkbench({
  route,
  embedded = false,
  language = "th",
  holdingcode = "",
}: OperationsWorkbenchProps) {
  const config = getOperationsConfig(route);
  if (config?.code === "delivery_dates" || config?.code === "purchase_receipt_dates") {
    return (
      <DeliveryDateOverhaulScreen
        mode={config.code === "purchase_receipt_dates" ? "vendor" : "customer"}
        language={language}
        holdingcode={holdingcode}
      />
    );
  }
  if (config?.code === "purchase_receipt_adjustment" || route === "/transaction/purchasereceiptadjustment") {
    return (
      <PurchaseReceiptAdjustmentScreen
        language={language}
        holdingcode={holdingcode}
      />
    );
  }
  return (
    <OperationsApprovalWorkbench
      key={route}
      route={route}
      embedded={embedded}
      language={language}
      holdingcode={holdingcode}
    />
  );
}

function OperationsApprovalWorkbench({
  route,
  embedded: _embedded = false,
  language = "th",
  holdingcode = "",
}: OperationsWorkbenchProps) {
  const tr = useBackendText();
  const dictionary = useBackendDictionary();
  const config = getOperationsConfig(route) || {
    route,
    code: "operations_workflow",
    category: "approval" as const,
    title: { th: "ระบบงานปฏิบัติการและเวิร์กโฟลว์", en: "Operations & Workflow" },
    description: { th: "จัดการกระบวนการทำงานและเอกสาร", en: "Process & workflow management" },
    primaryActionLabel: { th: "ดำเนินการ", en: "Execute" },
  };

  const meta = getApprovalOperationMeta(config.code);
  const actionType = meta?.actionType ?? "approve";

  const [searchTerm, setSearchTerm] = useState<string>("");
  const [docs, setDocs] = useState<PendingApprovalDoc[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);
  const [noticeKey, setNoticeKey] = useState<string | null>(null);
  const [noticeOk, setNoticeOk] = useState<boolean>(false);
  const [actionDetail, setActionDetail] = useState<string | null>(null);
  const [batchFailures, setBatchFailures] = useState<{ id: string; text: string }[]>([]);
  const [page, setPage] = useState<number>(1);
  const [total, setTotal] = useState<number>(0);
  const [selectedDocnos, setSelectedDocnos] = useState<Set<string>>(new Set());
  const [reasonInput, setReasonInput] = useState<string>("");
  const [pendingAction, setPendingAction] = useState<{
    docno: string;
    action: "approve" | "unapprove" | "cancel";
  } | null>(null);
  const [pendingBatchAction, setPendingBatchAction] = useState<"unapprove" | "cancel" | null>(null);
  const [historyDocNo, setHistoryDocNo] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState<boolean>(false);

  const approvalReady = config.category === "approval" && isApprovalApiReady(config.code);
  const reportCode = config.code;

  const loadDocs = useCallback(async () => {
    if (!approvalReady) return;
    setLoading(true);
    setErrorKey(null);
    const result = await fetchPendingApprovals({
      code: reportCode,
      holdingcode,
      offset: (page - 1) * PAGE_SIZE,
      limit: PAGE_SIZE,
    });
    setDocs(result.docs);
    setTotal(result.total);
    setErrorKey(result.error ?? null);
    setLoading(false);
    // ทำรายการหน้าสุดท้ายจนหมด → ถอยไปหน้าที่ยังมีข้อมูล (effect โหลดซ้ำเอง)
    if (!result.error && result.docs.length === 0 && result.total > 0 && page > 1) {
      setPage(Math.max(1, Math.ceil(result.total / PAGE_SIZE)));
    }
  }, [approvalReady, reportCode, holdingcode, page]);

  useEffect(() => {
    void loadDocs();
  }, [loadDocs]);

  const filteredDocs = useMemo(() => {
    const term = searchTerm.trim().toLowerCase();
    if (!term) return docs;
    return docs.filter((doc) =>
      [doc.docno, doc.requestorname, doc.counterpartyname].some((value) =>
        value.toLowerCase().includes(term),
      ),
    );
  }, [docs, searchTerm]);

  const toggleSelectDoc = useCallback((docno: string) => {
    setSelectedDocnos((prev) => {
      const next = new Set(prev);
      if (next.has(docno)) {
        next.delete(docno);
      } else {
        next.add(docno);
      }
      return next;
    });
  }, []);

  const toggleSelectAll = useCallback(() => {
    if (selectedDocnos.size === filteredDocs.length && filteredDocs.length > 0) {
      setSelectedDocnos(new Set());
    } else {
      setSelectedDocnos(new Set(filteredDocs.map((d) => d.docno)));
    }
  }, [selectedDocnos.size, filteredDocs]);

  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));
  // ค้นหาในหน้าที่โหลดอยู่ — มีคำค้นให้นับตามที่เห็น ไม่มีให้นับทั้งหมดจาก backend
  const shownCount = searchTerm.trim() ? filteredDocs.length : total;

  const changePage = useCallback((nextPage: number) => {
    setSelectedDocnos(new Set());
    setPage(nextPage);
  }, []);

  // ปุ่มทำรายการเป็นชุด: ข้อความที่มี {0} แทนด้วยจำนวนที่เลือก ไม่มี {0} ต่อท้ายด้วย (จำนวน)
  const withSelectedCount = (label: string) =>
    label.includes("{0}") ? label.replace("{0}", String(selectedDocnos.size)) : `${label} (${selectedDocnos.size})`;

  // backend ตอบผลทีละใบ — HTTP สำเร็จแต่บางใบไม่ผ่านต้องบอกว่าใบไหนและเพราะอะไร (ข้อความตามภาษาผ่านคีย์ gl_err_<code>)
  function showBatchOutcome(
    result: { success: boolean; messageKey: string; detail?: string; total: number; items?: BatchApproveItemResult[] },
    done: number,
    summary: { key: string; fallback: string },
  ) {
    const failed = (result.items ?? []).filter((item) => !item.success);
    setBatchFailures(
      failed.map((item) => ({
        id: item.id,
        text: item.code ? tr(`gl_err_${item.code}`, item.message ?? item.code) : (item.message ?? ""),
      })),
    );
    setNoticeOk(result.success && failed.length === 0);
    setNoticeKey(
      !result.success || failed.length === 0 ? result.messageKey : done === 0 ? "action_failed" : "batch_partial",
    );
    if (result.detail) {
      setActionDetail(result.detail);
    } else if (result.success) {
      setActionDetail(
        tr(summary.key, summary.fallback).replace("{0}", String(done)).replace("{1}", String(result.total)),
      );
    }
  }

  async function handleBatchApprove() {
    if (selectedDocnos.size === 0) return;
    setSubmitting(true);
    setActionDetail(null);
    setBatchFailures([]);
    const result = await submitBatchApprove({
      code: reportCode,
      docnos: Array.from(selectedDocnos),
      language,
    });
    setSubmitting(false);
    showBatchOutcome(result, result.approved, {
      key: "ops_batch_approve_summary",
      fallback: "อนุมัติสำเร็จ {0} จาก {1} รายการ",
    });
    setSelectedDocnos(new Set());
    await loadDocs();
  }

  async function confirmBatchAction() {
    if (!pendingBatchAction || selectedDocnos.size === 0) return;
    const docnos = Array.from(selectedDocnos);
    const reason = reasonInput.trim();
    setSubmitting(true);
    setActionDetail(null);
    setBatchFailures([]);
    if (pendingBatchAction === "unapprove") {
      const result = await submitBatchUnapprove({
        code: reportCode,
        docnos,
        comment: reason,
        language,
      });
      setSubmitting(false);
      showBatchOutcome(result, result.unapproved, {
        key: "ops_batch_unapprove_summary",
        fallback: "ยกเลิกการอนุมัติสำเร็จ {0} จาก {1} รายการ",
      });
    } else {
      const result = await submitBatchCancel({
        code: reportCode,
        docnos,
        comment: reason,
        language,
      });
      setSubmitting(false);
      showBatchOutcome(result, result.cancelled, {
        key: "ops_batch_cancel_summary",
        fallback: "ยกเลิกเอกสารสำเร็จ {0} จาก {1} รายการ",
      });
    }
    setPendingBatchAction(null);
    setReasonInput("");
    setSelectedDocnos(new Set());
    await loadDocs();
  }

  async function confirmAction() {
    if (!pendingAction) return;
    const actionby = getAuthSession()?.username ?? "";
    setSubmitting(true);
    setActionDetail(null);
    setBatchFailures([]);
    const result = await submitApprovalAction({
      code: reportCode,
      holdingcode,
      docno: pendingAction.docno,
      action: pendingAction.action,
      actionby,
      comment: reasonInput.trim(),
      language,
    });
    setSubmitting(false);
    setNoticeKey(result.messageKey);
    setNoticeOk(result.success);
    setActionDetail(result.detail ?? null);
    setPendingAction(null);
    setReasonInput("");
    if (result.success) {
      setSelectedDocnos((prev) => {
        const next = new Set(prev);
        next.delete(pendingAction.docno);
        return next;
      });
      await loadDocs();
    }
  }

  const unavailable = !approvalReady;

  return (
    <div className="flex flex-col gap-5 p-4 lg:p-6 max-w-6xl mx-auto">
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            {config.category === "approval" ? (
              actionType === "cancel" ? (
                <Ban className="h-6 w-6 text-destructive" />
              ) : actionType === "unapprove" ? (
                <Undo2 className="h-6 w-6 text-amber-500" />
              ) : (
                <ShieldCheck className="h-6 w-6" />
              )
            ) : config.category === "bom" ? (
              <Boxes className="h-6 w-6" />
            ) : (
              <Package className="h-6 w-6" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                {operationsText(config, "title", language, dictionary)}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                {config.category}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {operationsText(config, "description", language, dictionary)}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {actionType === "approve" && selectedDocnos.size > 0 && (
            <Button
              variant="default"
              onClick={() => void handleBatchApprove()}
              disabled={submitting || loading}
              className="gap-2 min-h-[44px]"
            >
              <CheckCircle className="h-4 w-4" />
              {withSelectedCount(tr("ops_batch_approve", "อนุมัติรายการที่เลือก"))}
            </Button>
          )}

          {actionType === "unapprove" && selectedDocnos.size > 0 && (
            <Button
              variant="outline"
              onClick={() => {
                setPendingBatchAction("unapprove");
                setReasonInput("");
              }}
              disabled={submitting || loading}
              className="gap-2 min-h-[44px] text-amber-600 hover:text-amber-700 border-amber-300 dark:border-amber-700"
            >
              <Undo2 className="h-4 w-4" />
              {withSelectedCount(tr("ops_batch_unapprove", "ยกเลิกการอนุมัติที่เลือก"))}
            </Button>
          )}

          {actionType === "cancel" && selectedDocnos.size > 0 && (
            <Button
              variant="destructive"
              onClick={() => {
                setPendingBatchAction("cancel");
                setReasonInput("");
              }}
              disabled={submitting || loading}
              className="gap-2 min-h-[44px]"
            >
              <Ban className="h-4 w-4" />
              {withSelectedCount(tr("ops_batch_cancel", "ยกเลิกเอกสารที่เลือก"))}
            </Button>
          )}

          {approvalReady && (
            <Button
              variant="outline"
              onClick={() => void loadDocs()}
              disabled={loading}
              className="gap-2 min-h-[44px]"
            >
              <RotateCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              {tr("ops_refresh", "โหลดรายการใหม่")}
            </Button>
          )}
        </div>
      </div>

      {/* Availability Notice */}
      {unavailable && (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{messageText("approval_not_available", tr)}</span>
        </div>
      )}

      {/* Action Result Notice */}
      {noticeKey && (
        <div
          className={`flex items-center justify-between gap-3 rounded-xl border px-4 py-3 text-sm font-semibold ${
            noticeOk
              ? "border-primary/30 bg-primary/10 text-primary"
              : "border-destructive/30 bg-destructive/10 text-destructive"
          }`}
          role={noticeOk ? "status" : "alert"}
        >
          <div className="flex flex-col gap-0.5">
            <span className="flex items-center gap-2">
              {noticeOk ? <CheckCircle className="h-5 w-5 shrink-0" /> : <AlertCircle className="h-5 w-5 shrink-0" />}
              {messageText(noticeKey, tr)}
            </span>
            {actionDetail && (
              <span className="text-xs font-normal pl-7 opacity-90">{actionDetail}</span>
            )}
            {batchFailures.length > 0 && (
              <ul className="mt-1 max-h-40 list-disc space-y-0.5 overflow-y-auto pl-12 text-xs font-normal">
                {batchFailures.map((failure) => (
                  <li key={failure.id}>
                    <span className="font-mono font-semibold">{failure.id}</span>
                    {failure.text ? `: ${failure.text}` : ""}
                  </li>
                ))}
              </ul>
            )}
          </div>
          <button
            type="button"
            onClick={() => {
              setNoticeKey(null);
              setActionDetail(null);
              setBatchFailures([]);
            }}
            className="rounded-md px-2 py-1 text-xs underline"
            aria-label={tr("ops_dismiss_message", "ปิดข้อความแจ้งเตือน")}
            title={tr("ops_dismiss_message", "ปิดข้อความแจ้งเตือน")}
          >
            {tr("close", "ปิด")}
          </button>
        </div>
      )}

      {/* Load Error */}
      {approvalReady && errorKey && (
        <div className="flex items-center gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{messageText(errorKey, tr)}</span>
        </div>
      )}

      {/* Document List */}
      {approvalReady && (
        <Card className="overflow-hidden">
          <div className="flex items-center justify-between border-b p-4 bg-muted/40">
            <div className="flex items-center gap-2 font-semibold text-sm">
              <Clock className="h-4 w-4 text-primary" />
              <span>
                {actionType === "approve"
                  ? tr("ops_pending_approvals_count", "รายการรอการอนุมัติ ({0} รายการ)").replace(
                      "{0}",
                      String(shownCount),
                    )
                  : actionType === "unapprove"
                  ? tr("ops_approved_docs_count", "รายการที่อนุมัติแล้ว ({0} รายการ)").replace(
                      "{0}",
                      String(shownCount),
                    )
                  : tr("ops_active_docs_count", "รายการเอกสารที่ใช้งานอยู่ ({0} รายการ)").replace(
                      "{0}",
                      String(shownCount),
                    )}
              </span>
            </div>
            <div className="relative w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={tr("ops_search", "ค้นหาเอกสาร...")}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-8 h-8 text-xs"
              />
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="border-b bg-muted/60 text-xs font-semibold text-muted-foreground uppercase">
                <tr>
                  <th className="px-3 py-3 w-10 text-center">
                    <input
                      type="checkbox"
                      checked={filteredDocs.length > 0 && selectedDocnos.size === filteredDocs.length}
                      onChange={toggleSelectAll}
                      aria-label={tr("ops_select_all", "เลือกทั้งหมด")}
                      className="rounded border-border text-primary focus:ring-primary h-4 w-4 cursor-pointer"
                    />
                  </th>
                  <th className="px-4 py-3">{tr("document_no", "เลขที่เอกสาร")}</th>
                  <th className="px-4 py-3">{tr("date", "วันที่")}</th>
                  <th className="px-4 py-3">{tr("ops_requested_by", "ผู้ขออนุมัติ")}</th>
                  <th className="px-4 py-3">{tr("ops_counterparty", "คู่ค้า / ลูกค้า")}</th>
                  <th className="px-4 py-3 text-right">{tr("ops_total", "ยอดเงินรวม")}</th>
                  <th className="px-4 py-3 text-center">{tr("manage", "จัดการ")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {loading ? (
                  <tr>
                    <td colSpan={7} className="py-12 text-center text-muted-foreground">
                      <span className="inline-flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        {tr("ops_loading_data", "กำลังโหลดข้อมูลจากระบบ...")}
                      </span>
                    </td>
                  </tr>
                ) : filteredDocs.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="py-12 text-center text-muted-foreground">
                      {errorKey
                        ? messageText(errorKey, tr)
                        : tr("ops_no_pending_documents", "ไม่มีเอกสารในรายการ")}
                    </td>
                  </tr>
                ) : (
                  filteredDocs.map((doc) => (
                    <tr key={doc.docno} className="hover:bg-muted/30 transition-colors">
                      <td className="px-3 py-3 text-center">
                        <input
                          type="checkbox"
                          checked={selectedDocnos.has(doc.docno)}
                          onChange={() => toggleSelectDoc(doc.docno)}
                          aria-label={`${tr("ops_select_doc", "เลือกเอกสาร")} ${doc.docno}`}
                          className="rounded border-border text-primary focus:ring-primary h-4 w-4 cursor-pointer"
                        />
                      </td>
                      <td className="px-4 py-3 font-semibold text-primary font-mono">{doc.docno}</td>
                      <td className="px-4 py-3 font-mono text-xs">{doc.docdate || "-"}</td>
                      <td className="px-4 py-3">{doc.requestorname || "-"}</td>
                      <td className="px-4 py-3">{doc.counterpartyname || "-"}</td>
                      <td className="px-4 py-3 text-right font-mono font-bold">
                        {doc.totalamount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                        {tr("baht_suffix", " บาท")}
                      </td>
                      <td className="px-4 py-3 text-center">
                        {pendingAction?.docno === doc.docno ? (
                          <div className="flex flex-col items-center gap-2 min-w-[260px] p-2 bg-muted/60 rounded-xl border border-border">
                            <span className="text-xs font-semibold text-foreground text-center">
                              {pendingAction.action === "approve"
                                ? tr(
                                    "ops_confirm_approve_doc",
                                    "ยืนยันอนุมัติเอกสาร {0} ใช่หรือไม่?",
                                  ).replace("{0}", doc.docno)
                                : pendingAction.action === "unapprove"
                                ? tr(
                                    "ops_confirm_unapprove_doc",
                                    "ยืนยันยกเลิกการอนุมัติเอกสาร {0} ใช่หรือไม่?",
                                  ).replace("{0}", doc.docno)
                                : tr(
                                    "ops_confirm_cancel_doc",
                                    "ยืนยันยกเลิกเอกสาร {0} ใช่หรือไม่?",
                                  ).replace("{0}", doc.docno)}
                            </span>

                            {pendingAction.action !== "approve" && (
                              <Input
                                placeholder={
                                  pendingAction.action === "cancel"
                                    ? tr("ops_cancel_reason_required", "ระบุเหตุผลในการยกเลิก (จำเป็น)...")
                                    : tr("ops_enter_reason", "ระบุเหตุผล (ถ้ามี)...")
                                }
                                value={reasonInput}
                                onChange={(e) => setReasonInput(e.target.value)}
                                className="h-8 text-xs bg-background"
                              />
                            )}

                            <div className="flex items-center justify-center gap-2">
                              <Button
                                size="sm"
                                variant={pendingAction.action === "cancel" ? "destructive" : "default"}
                                disabled={
                                  submitting ||
                                  (pendingAction.action === "cancel" && reasonInput.trim() === "")
                                }
                                onClick={() => void confirmAction()}
                                className="h-8 gap-1 text-xs"
                              >
                                {submitting ? (
                                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                                ) : (
                                  <CheckCircle className="h-3.5 w-3.5" />
                                )}
                                {tr("confirm", "ยืนยัน")}
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                disabled={submitting}
                                onClick={() => {
                                  setPendingAction(null);
                                  setReasonInput("");
                                }}
                                className="h-8 gap-1 text-xs"
                              >
                                {tr("cancel", "ยกเลิก")}
                              </Button>
                            </div>
                          </div>
                        ) : (
                          <div className="flex items-center justify-center gap-2">
                            {actionType === "approve" && (
                              <Button
                                size="sm"
                                variant="default"
                                onClick={() => {
                                  setPendingAction({ docno: doc.docno, action: "approve" });
                                  setReasonInput("");
                                }}
                                className="h-8 gap-1 text-xs"
                              >
                                <CheckCircle className="h-3.5 w-3.5" />
                                {tr("approve", "อนุมัติ")}
                              </Button>
                            )}

                            {actionType === "unapprove" && (
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => {
                                  setPendingAction({ docno: doc.docno, action: "unapprove" });
                                  setReasonInput("");
                                }}
                                className="h-8 gap-1 text-xs text-amber-600 hover:text-amber-700"
                              >
                                <Undo2 className="h-3.5 w-3.5" />
                                {tr("ops_unapprove_action", "ยกเลิกการอนุมัติ")}
                              </Button>
                            )}

                            {actionType === "cancel" && (
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => {
                                  setPendingAction({ docno: doc.docno, action: "cancel" });
                                  setReasonInput("");
                                }}
                                className="h-8 gap-1 text-xs text-destructive hover:text-destructive"
                              >
                                <Ban className="h-3.5 w-3.5" />
                                {tr("ops_cancel_action", "ยกเลิกเอกสาร")}
                              </Button>
                            )}

                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => setHistoryDocNo(doc.docno)}
                              className="h-8 w-8 p-0 text-muted-foreground hover:text-foreground"
                              title={tr("ops_doc_events", "ประวัติการอนุมัติ/ยกเลิก")}
                              aria-label={tr("ops_doc_events", "ประวัติการอนุมัติ/ยกเลิก")}
                            >
                              <Clock className="h-4 w-4" />
                            </Button>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {pageCount > 1 && (
            <div className="flex flex-wrap items-center justify-between gap-2 border-t p-3 text-sm">
              <span>
                {tr("ops_page_of", "หน้า {0} จาก {1} · {2} รายการ")
                  .replace("{0}", String(page))
                  .replace("{1}", String(pageCount))
                  .replace("{2}", total.toLocaleString("th-TH"))}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  className="min-h-[44px]"
                  disabled={loading || page <= 1}
                  onClick={() => changePage(page - 1)}
                >
                  {tr("previous", "ก่อนหน้า")}
                </Button>
                <Button
                  variant="outline"
                  className="min-h-[44px]"
                  disabled={loading || page >= pageCount}
                  onClick={() => changePage(page + 1)}
                >
                  {tr("next", "ถัดไป")}
                </Button>
              </div>
            </div>
          )}
        </Card>
      )}

      {/* Dialog ยืนยันการทำรายการเป็นชุด (ยกเลิกการอนุมัติ / ยกเลิกเอกสาร) */}
      {pendingBatchAction && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-xl animate-in fade-in zoom-in-95 duration-150">
            <h3 className="text-lg font-bold">
              {pendingBatchAction === "cancel"
                ? tr("ops_confirm_batch_cancel", "ยืนยันยกเลิกเอกสารรายการที่เลือก")
                : tr("ops_confirm_batch_unapprove", "ยืนยันยกเลิกการอนุมัติรายการที่เลือก")}
            </h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {tr("ops_batch_count_hint", "จำนวนที่เลือก {0} รายการ").replace("{0}", String(selectedDocnos.size))}
            </p>
            <div className="mt-4">
              <label className="text-xs font-semibold text-foreground">
                {tr("ops_reason_label", "เหตุผล")}
                {pendingBatchAction === "cancel" && <span className="text-destructive"> *</span>}
              </label>
              <Input
                value={reasonInput}
                onChange={(e) => setReasonInput(e.target.value)}
                placeholder={tr("ops_reason_placeholder", "ระบุเหตุผล...")}
                className="mt-1 text-sm"
                autoFocus
              />
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setPendingBatchAction(null);
                  setReasonInput("");
                }}
                disabled={submitting}
              >
                {tr("cancel", "ยกเลิก")}
              </Button>
              <Button
                variant={pendingBatchAction === "cancel" ? "destructive" : "default"}
                disabled={submitting || (pendingBatchAction === "cancel" && !reasonInput.trim())}
                onClick={() => void confirmBatchAction()}
              >
                {submitting ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : null}
                {tr("confirm", "ยืนยัน")}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Dialog ประวัติการอนุมัติ/ยกเลิก */}
      <ErpDocEventsDialog
        module={meta?.module ?? ""}
        docno={historyDocNo ?? ""}
        isOpen={!!historyDocNo}
        onClose={() => setHistoryDocNo(null)}
        language={language}
      />
    </div>
  );
}
