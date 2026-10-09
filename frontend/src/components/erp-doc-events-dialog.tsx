"use client";

import { useEffect, useState, useCallback } from "react";
import { useBackendText } from "@/components/backend-text-provider";
import { fetchErpDocEvents, type ErpDocEvent } from "@/lib/erp-transaction";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import {
  CheckCircle,
  Undo2,
  Ban,
  Clock,
  X,
  Loader2,
  AlertCircle,
  FileText,
  User,
} from "lucide-react";

interface ErpDocEventsDialogProps {
  module: string;
  docno: string;
  isOpen: boolean;
  onClose: () => void;
  language?: LanguageCode;
}

function formatEventTime(isoString: string): string {
  if (!isoString) return "-";
  try {
    const d = new Date(isoString);
    if (isNaN(d.getTime())) return isoString;
    return d.toLocaleString("th-TH", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch {
    return isoString;
  }
}

export function ErpDocEventsDialog({
  module,
  docno,
  isOpen,
  onClose,
  language = "th",
}: ErpDocEventsDialogProps) {
  const tr = useBackendText();
  const [events, setEvents] = useState<ErpDocEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const loadEvents = useCallback(async () => {
    if (!docno || !module) return;
    setLoading(true);
    setError(null);
    const res = await fetchErpDocEvents(module, docno, language);
    if (res.success) {
      setEvents(res.events);
    } else {
      setError(res.error || "load_failed");
    }
    setLoading(false);
  }, [module, docno, language]);

  useEffect(() => {
    if (isOpen) {
      void loadEvents();
    } else {
      setEvents([]);
      setError(null);
    }
  }, [isOpen, loadEvents]);

  if (!isOpen) return null;

  function actionBadge(action: string) {
    switch (action) {
      case "approve":
        return {
          icon: <CheckCircle className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />,
          bg: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20",
          label: tr("erp_msg_approved", "อนุมัติเอกสาร"),
        };
      case "unapprove":
        return {
          icon: <Undo2 className="h-4 w-4 text-amber-600 dark:text-amber-400" />,
          bg: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
          label: tr("erp_msg_unapproved", "ยกเลิกการอนุมัติ"),
        };
      case "cancel":
        return {
          icon: <Ban className="h-4 w-4 text-destructive" />,
          bg: "bg-destructive/10 text-destructive border-destructive/20",
          label: tr("erp_msg_cancelled", "ยกเลิกเอกสาร"),
        };
      default:
        return {
          icon: <FileText className="h-4 w-4 text-primary" />,
          bg: "bg-primary/10 text-primary border-primary/20",
          label: action,
        };
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-in fade-in duration-200"
      role="dialog"
      aria-modal="true"
      aria-labelledby="events-dialog-title"
    >
      <div className="flex flex-col max-h-[85vh] w-full max-w-lg rounded-2xl border border-border bg-card shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border p-4 bg-muted/30">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Clock className="h-5 w-5" />
            </div>
            <div>
              <h2 id="events-dialog-title" className="text-base font-bold text-foreground">
                {tr("ops_doc_events", "ประวัติการอนุมัติ/ยกเลิก")}
              </h2>
              <p className="text-xs font-mono text-muted-foreground">
                {tr("document_no", "เลขที่เอกสาร")}: <span className="font-semibold text-primary">{docno}</span>
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            aria-label={tr("close", "ปิด")}
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-5">
          {loading ? (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
              <span className="text-xs">{tr("ops_loading_data", "กำลังโหลดข้อมูลจากระบบ...")}</span>
            </div>
          ) : error ? (
            <div className="flex flex-col items-center justify-center py-8 text-center gap-3">
              <AlertCircle className="h-8 w-8 text-destructive" />
              <p className="text-sm font-semibold text-destructive">
                {tr("load_data_failed", "โหลดข้อมูลไม่สำเร็จ")}
              </p>
              <Button size="sm" variant="outline" onClick={() => void loadEvents()}>
                {tr("common_retry", "ลองใหม่")}
              </Button>
            </div>
          ) : events.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground gap-2">
              <FileText className="h-10 w-10 stroke-1 opacity-50" />
              <p className="text-sm">{tr("ops_no_events_found", "ยังไม่มีประวัติการอนุมัติหรือยกเลิกสำหรับเอกสารนี้")}</p>
            </div>
          ) : (
            <div className="relative pl-6 space-y-6 before:absolute before:left-2.5 before:top-2 before:bottom-2 before:w-0.5 before:bg-border">
              {events.map((ev, index) => {
                const badge = actionBadge(ev.action);
                return (
                  <div key={ev.id || index} className="relative group">
                    {/* Dot on timeline */}
                    <div className="absolute -left-[19px] top-1 flex h-5 w-5 items-center justify-center rounded-full bg-background border border-border shadow-sm">
                      <div className="h-2 w-2 rounded-full bg-primary" />
                    </div>

                    <div className="flex flex-col gap-1.5 rounded-xl border border-border/60 bg-muted/20 p-3.5 hover:bg-muted/40 transition-colors">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span
                          className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${badge.bg}`}
                        >
                          {badge.icon}
                          {badge.label}
                        </span>
                        <span className="text-xs font-mono text-muted-foreground">
                          {formatEventTime(ev.created_at || ev.createdat || "")}
                        </span>
                      </div>

                      <div className="flex items-center gap-1.5 text-xs text-muted-foreground mt-0.5">
                        <User className="h-3.5 w-3.5" />
                        <span>{tr("ops_performed_by", "ผู้ดำเนินการ")}:</span>
                        <span className="font-semibold text-foreground">{ev.actor || ev.createdby || "-"}</span>
                      </div>

                      {ev.reason && ev.reason.trim() !== "" && (
                        <div className="mt-1.5 rounded-lg bg-background p-2.5 text-xs text-foreground/90 border border-border/50">
                          <span className="font-semibold text-muted-foreground mr-1">
                            {tr("ops_reason_label", "เหตุผล")}:
                          </span>
                          {ev.reason}
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end border-t border-border p-3 bg-muted/20">
          <Button variant="outline" size="sm" onClick={onClose} className="h-9 px-4 text-xs">
            {tr("close", "ปิด")}
          </Button>
        </div>
      </div>
    </div>
  );
}
