"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  getOperationsConfig,
  isApprovalApiReady,
  fetchPendingApprovals,
  submitApprovalAction,
  type PendingApprovalDoc,
} from "@/lib/erp-operations";
import { getAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import {
  CheckCircle,
  XCircle,
  Clock,
  FileSpreadsheet,
  Upload,
  Search,
  Package,
  Boxes,
  ShieldCheck,
  AlertCircle,
  Loader2,
  RotateCw,
} from "lucide-react";

interface OperationsWorkbenchProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  holdingcode?: string;
}

const MESSAGES: Record<string, { th: string; en: string }> = {
  approval_not_available: {
    th: "จอนี้ยังไม่เชื่อมกับระบบงานจริง — อยู่ระหว่างเปิดใช้งาน API",
    en: "This screen is not connected to the live system yet",
  },
  holding_required: { th: "ยังไม่ได้เลือกกิจการ", en: "No business selected" },
  docno_required: { th: "ไม่พบเลขที่เอกสาร", en: "Document number is missing" },
  unauthorized: { th: "ไม่มีสิทธิ์ดำเนินการรายการนี้", en: "You do not have permission for this action" },
  load_failed: { th: "โหลดข้อมูลไม่สำเร็จ", en: "Failed to load data" },
  connection_error: { th: "เชื่อมต่อระบบไม่ได้", en: "Unable to connect to the system" },
  action_failed: { th: "ดำเนินการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง", en: "Action failed, please try again" },
  approve_success: { th: "อนุมัติเอกสารเรียบร้อยแล้ว", en: "Document approved" },
  reject_success: { th: "บันทึกการไม่อนุมัติเรียบร้อยแล้ว", en: "Rejection recorded" },
};

function messageText(key: string, language: LanguageCode): string {
  const message = MESSAGES[key];
  if (!message) return key;
  return language === "th" ? message.th : message.en;
}

export function OperationsWorkbench({
  route,
  embedded: _embedded = false,
  language = "th",
  holdingcode = "",
}: OperationsWorkbenchProps) {
  const config = getOperationsConfig(route) || {
    route,
    code: "operations_workflow",
    category: "approval" as const,
    title: { th: "ระบบงานปฏิบัติการและเวิร์กโฟลว์", en: "Operations & Workflow" },
    description: { th: "จัดการกระบวนการทำงานและเอกสาร", en: "Process & workflow management" },
    primaryActionLabel: { th: "ดำเนินการ", en: "Execute" },
  };

  const [searchTerm, setSearchTerm] = useState<string>("");
  const [docs, setDocs] = useState<PendingApprovalDoc[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [errorKey, setErrorKey] = useState<string | null>(null);
  const [noticeKey, setNoticeKey] = useState<string | null>(null);
  const [noticeOk, setNoticeOk] = useState<boolean>(false);
  const [pendingAction, setPendingAction] = useState<{ docno: string; action: "approve" | "reject" } | null>(null);
  const [submitting, setSubmitting] = useState<boolean>(false);

  const approvalReady = config.category === "approval" && isApprovalApiReady(config.code);
  const reportCode = config.code;

  const loadDocs = useCallback(async () => {
    if (!approvalReady) return;
    setLoading(true);
    setErrorKey(null);
    const result = await fetchPendingApprovals({ code: reportCode, holdingcode });
    setDocs(result.docs);
    setErrorKey(result.error ?? null);
    setLoading(false);
  }, [approvalReady, reportCode, holdingcode]);

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

  async function confirmAction() {
    if (!pendingAction) return;
    const actionby = getAuthSession()?.username ?? "";
    setSubmitting(true);
    const result = await submitApprovalAction({
      code: reportCode,
      holdingcode,
      docno: pendingAction.docno,
      action: pendingAction.action,
      actionby,
    });
    setSubmitting(false);
    setNoticeKey(result.messageKey);
    setNoticeOk(result.success);
    setPendingAction(null);
    if (result.success) await loadDocs();
  }

  const unavailable = !approvalReady;

  return (
    <div className="flex flex-col gap-5 p-4 lg:p-6 max-w-6xl mx-auto">
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            {config.category === "approval" ? (
              <ShieldCheck className="h-6 w-6" />
            ) : config.category === "bom" ? (
              <Boxes className="h-6 w-6" />
            ) : config.category === "import" ? (
              <FileSpreadsheet className="h-6 w-6" />
            ) : (
              <Package className="h-6 w-6" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                {language === "th" ? config.title.th : config.title.en}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                {config.category}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {language === "th" ? config.description.th : config.description.en}
            </p>
          </div>
        </div>

        {approvalReady && (
          <Button
            variant="outline"
            onClick={() => void loadDocs()}
            disabled={loading}
            className="gap-2 min-h-[44px]"
          >
            <RotateCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            {language === "th" ? "โหลดรายการใหม่" : "Refresh"}
          </Button>
        )}
      </div>

      {/* Availability Notice */}
      {unavailable && (
        <div className="flex items-center gap-2 rounded-xl border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground" role="status">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{messageText("approval_not_available", language)}</span>
        </div>
      )}

      {/* Action Result */}
      {noticeKey && (
        <div
          className={`flex items-center justify-between gap-3 rounded-xl border px-4 py-3 text-sm font-semibold ${
            noticeOk
              ? "border-primary/30 bg-primary/10 text-primary"
              : "border-destructive/30 bg-destructive/10 text-destructive"
          }`}
          role={noticeOk ? "status" : "alert"}
        >
          <span className="flex items-center gap-2">
            {noticeOk ? <CheckCircle className="h-5 w-5" /> : <AlertCircle className="h-5 w-5" />}
            {messageText(noticeKey, language)}
          </span>
          <button
            type="button"
            onClick={() => setNoticeKey(null)}
            className="rounded-md px-2 py-1 text-xs underline"
            aria-label={language === "th" ? "ปิดข้อความแจ้งเตือน" : "Dismiss message"}
            title={language === "th" ? "ปิดข้อความแจ้งเตือน" : "Dismiss message"}
          >
            {language === "th" ? "ปิด" : "Close"}
          </button>
        </div>
      )}

      {/* Load Error */}
      {approvalReady && errorKey && (
        <div className="flex items-center gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{messageText(errorKey, language)}</span>
        </div>
      )}

      {/* Pending Approvals */}
      {approvalReady && (
        <Card className="overflow-hidden">
          <div className="flex items-center justify-between border-b p-4 bg-muted/40">
            <div className="flex items-center gap-2 font-semibold text-sm">
              <Clock className="h-4 w-4 text-primary" />
              <span>
                {language === "th"
                  ? `รายการรอการอนุมัติ (${filteredDocs.length} รายการ)`
                  : `Pending Approvals (${filteredDocs.length})`}
              </span>
            </div>
            <div className="relative w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={language === "th" ? "ค้นหาเอกสาร..." : "Search..."}
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
                  <th className="px-4 py-3">{language === "th" ? "เลขที่เอกสาร" : "Document No."}</th>
                  <th className="px-4 py-3">{language === "th" ? "วันที่" : "Date"}</th>
                  <th className="px-4 py-3">{language === "th" ? "ผู้ขออนุมัติ" : "Requested by"}</th>
                  <th className="px-4 py-3">{language === "th" ? "คู่ค้า / ลูกค้า" : "Counterparty"}</th>
                  <th className="px-4 py-3 text-right">{language === "th" ? "ยอดเงินรวม" : "Total"}</th>
                  <th className="px-4 py-3 text-center">{language === "th" ? "จัดการ" : "Action"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {loading ? (
                  <tr>
                    <td colSpan={6} className="py-12 text-center text-muted-foreground">
                      <span className="inline-flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        {language === "th" ? "กำลังโหลดข้อมูลจากระบบ..." : "Loading data..."}
                      </span>
                    </td>
                  </tr>
                ) : filteredDocs.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-12 text-center text-muted-foreground">
                      {errorKey
                        ? messageText(errorKey, language)
                        : language === "th"
                          ? "ไม่มีเอกสารค้างรอการอนุมัติ"
                          : "No pending documents"}
                    </td>
                  </tr>
                ) : (
                  filteredDocs.map((doc) => (
                    <tr key={doc.docno} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-semibold text-primary font-mono">{doc.docno}</td>
                      <td className="px-4 py-3 font-mono text-xs">{doc.docdate || "-"}</td>
                      <td className="px-4 py-3">{doc.requestorname || "-"}</td>
                      <td className="px-4 py-3">{doc.counterpartyname || "-"}</td>
                      <td className="px-4 py-3 text-right font-mono font-bold">
                        {doc.totalamount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                        {language === "th" ? " บาท" : " THB"}
                      </td>
                      <td className="px-4 py-3 text-center">
                        {pendingAction?.docno === doc.docno ? (
                          <div className="flex flex-col items-center gap-2">
                            <span className="text-xs font-semibold text-foreground">
                              {language === "th"
                                ? pendingAction.action === "approve"
                                  ? `ยืนยันอนุมัติเอกสาร ${doc.docno} ใช่หรือไม่? เมื่ออนุมัติแล้วเอกสารจะถูกส่งต่อตามขั้นตอนทันที`
                                  : `ยืนยันไม่อนุมัติเอกสาร ${doc.docno} ใช่หรือไม่? ผู้ขอจะต้องแก้ไขและส่งใหม่`
                                : `Confirm ${pendingAction.action} for ${doc.docno}?`}
                            </span>
                            <div className="flex items-center justify-center gap-2">
                              <Button
                                size="sm"
                                variant="default"
                                disabled={submitting}
                                onClick={() => void confirmAction()}
                                className="h-8 gap-1 text-xs"
                              >
                                {submitting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <CheckCircle className="h-3.5 w-3.5" />}
                                {language === "th" ? "ยืนยัน" : "Confirm"}
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                disabled={submitting}
                                onClick={() => setPendingAction(null)}
                                className="h-8 gap-1 text-xs"
                              >
                                {language === "th" ? "ยกเลิก" : "Cancel"}
                              </Button>
                            </div>
                          </div>
                        ) : (
                          <div className="flex items-center justify-center gap-2">
                            <Button
                              size="sm"
                              variant="default"
                              onClick={() => setPendingAction({ docno: doc.docno, action: "approve" })}
                              className="h-8 gap-1 text-xs"
                            >
                              <CheckCircle className="h-3.5 w-3.5" />
                              {language === "th" ? "อนุมัติ" : "Approve"}
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => setPendingAction({ docno: doc.docno, action: "reject" })}
                              className="h-8 gap-1 text-xs text-destructive"
                            >
                              <XCircle className="h-3.5 w-3.5" />
                              {language === "th" ? "ไม่อนุมัติ" : "Reject"}
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
        </Card>
      )}

      {/* Import screens: upload is not wired to the backend yet */}
      {config.category === "import" && (
        <Card className="border-2 border-dashed border-border p-8 text-center bg-muted/10 rounded-2xl">
          <div className="flex flex-col items-center justify-center gap-3">
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
              <Upload className="h-8 w-8" />
            </div>
            <div>
              <h3 className="text-base font-bold text-foreground">
                {language === "th" ? "ระบบนำเข้าข้อมูลยังไม่เปิดใช้งาน" : "Data import is not enabled yet"}
              </h3>
              <p className="text-sm text-muted-foreground mt-1">
                {language === "th"
                  ? "เมื่อเปิดใช้งานแล้วจะสามารถนำเข้าไฟล์ Excel หรือ CSV ผ่านระบบหลังบ้านได้จากจอนี้"
                  : "Once enabled, Excel or CSV files will be imported through the backend from this screen."}
              </p>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}
