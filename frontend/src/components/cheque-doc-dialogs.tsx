"use client";

import { Loader2, Search, ShieldAlert } from "lucide-react";
import { useEffect, useRef, useState, type ReactNode } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { isEarlyConfirmClick } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { chequeErrorText, listSelectableCheques, type ChequeRegister } from "@/lib/cheque-api";
import { CHEQUE_LIMITS } from "@/lib/cheque-doc-forms";
import { formatDecimalText } from "@/lib/cheque-forms";
import { formatAppDate } from "@/lib/date-time";
import type { LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export type Translate = (key: string, fallback: string) => string;

const overlayClass = "fixed inset-0 z-[100] grid place-items-center bg-black/45 p-2 text-foreground backdrop-blur-[2px]";

/**
 * หน้าต่างยกเลิกเอกสารเช็ค: อธิบายผลที่จะเกิด + ช่องเหตุผลที่ต้องกรอก (backend ปฏิเสธถ้าไม่มีเหตุผล)
 * ค้างอยู่จนกว่าจะสำเร็จหรือผู้ใช้ปิดเอง — ความผิดพลาดจาก backend แสดงในหน้าต่างนี้เพื่อให้แก้เหตุผลหรือปิดได้ทันที
 */
export function ChequeReasonDialog({
  tr,
  title,
  description,
  busy,
  error,
  onConfirm,
  onCancel,
}: {
  tr: Translate;
  title: string;
  description: ReactNode;
  busy: boolean;
  error: string;
  onConfirm: (reason: string) => void;
  onCancel: () => void;
}) {
  const [reason, setReason] = useState("");
  const openedAt = useRef(0);
  useEffect(() => {
    openedAt.current = performance.now();
  }, []);
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !busy) onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [busy, onCancel]);

  const trimmed = reason.trim();
  // กันการดับเบิลคลิกที่เปิดหน้าต่างนี้ไปกดยืนยันทับ (เหมือน confirm-dialog กลาง)
  const guarded = (action: () => void) => (event: { detail: number }) => {
    if (!isEarlyConfirmClick(event.detail, openedAt.current, performance.now())) action();
  };

  return (
    <div className={overlayClass} role="dialog" aria-modal="true" aria-labelledby="cheque-reason-title">
      <div className="grid gap-3 rounded-2xl border border-destructive/30 bg-card p-3 text-card-foreground shadow-xl" style={{ width: "min(520px, calc(100vw - 16px))" }}>
        <header className="flex min-w-0 items-start gap-2">
          <span className="grid size-10 shrink-0 place-items-center rounded-2xl bg-destructive/10 text-destructive">
            <ShieldAlert size={21} aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <h2 id="cheque-reason-title" className="text-base font-semibold leading-6">
              {title}
            </h2>
            <div className="mt-1 text-[0.95rem] leading-6 text-muted-foreground">{description}</div>
          </div>
        </header>
        <div className="grid gap-1">
          <label htmlFor="cheque-reason-text" className="text-[0.95rem] font-semibold">
            {tr("cheque_cancel_reason_label", "เหตุผลที่ยกเลิก")}
            <span aria-hidden="true" className="text-destructive"> *</span>
          </label>
          <Textarea
            id="cheque-reason-text"
            className="min-h-24 text-[0.95rem]"
            value={reason}
            maxLength={CHEQUE_LIMITS.description}
            autoFocus
            aria-invalid={error !== ""}
            placeholder={tr("cheque_cancel_reason_placeholder", "ระบุเหตุผล เช่น ออกเอกสารผิดวันที่ หรือเลือกเช็คผิดใบ")}
            onChange={(event) => setReason(event.target.value)}
          />
          {error ? (
            <span className="text-[0.9rem] font-medium text-destructive" role="alert">
              {error}
            </span>
          ) : null}
        </div>
        <footer className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" disabled={busy} onClick={guarded(onCancel)}>
            {tr("cheque_cancel_keep", "ไม่ยกเลิก")}
          </Button>
          <Button type="button" variant="destructive" disabled={busy || trimmed === ""} onClick={guarded(() => onConfirm(trimmed))}>
            {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : null}
            {tr("cheque_cancel_confirm", "ยืนยันยกเลิกเอกสาร")}
          </Button>
        </footer>
      </div>
    </div>
  );
}

const PICKER_PAGE_SIZE = 50;

/**
 * หน้าต่างเลือกเช็คเข้าเอกสารสถานะ: รายการมาจาก registers/selectable ตามวันที่เอกสาร สมุด และชนิดการยกเลิก
 * (backend ตรวจสิทธิ์เลือกซ้ำตอนบันทึก) — เลือกได้หลายใบ เช็คที่อยู่ในเอกสารแล้วติ๊กไม่ได้
 */
export function ChequePickerDialog({
  tr,
  language,
  filter,
  excludeIds,
  bankNames,
  onPick,
  onClose,
}: {
  tr: Translate;
  language: LanguageCode;
  filter: { transflag: number; docDate: string; book: string; cancelType: number | null; cancelPassed: boolean };
  excludeIds: ReadonlySet<string>;
  bankNames: ReadonlyMap<string, string>;
  onPick: (picked: ChequeRegister[]) => void;
  onClose: () => void;
}) {
  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<ChequeRegister[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [chosen, setChosen] = useState<Map<string, ChequeRegister>>(new Map());
  const { transflag, docDate, book, cancelType, cancelPassed } = filter;

  useEffect(() => {
    const timer = setTimeout(() => {
      setAppliedSearch(search.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  useEffect(() => {
    let current = true;
    setLoading(true);
    setLoadError(null);
    listSelectableCheques({ transflag, docDate, book, cancelType, cancelPassed, q: appliedSearch, page, limit: PICKER_PAGE_SIZE }, language)
      .then((result) => {
        if (!current) return;
        setItems(result.items ?? []);
        setTotal(result.total ?? 0);
      })
      .catch((cause: unknown) => {
        if (current) setLoadError(cause);
      })
      .finally(() => {
        if (current) setLoading(false);
      });
    return () => {
      current = false;
    };
  }, [transflag, docDate, book, cancelType, cancelPassed, appliedSearch, page, language]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onClose]);

  const pages = Math.max(1, Math.ceil(total / PICKER_PAGE_SIZE));
  const pageInfo = tr("bank_code_page_info", "หน้า {page} จาก {pages}").replace("{page}", String(page)).replace("{pages}", String(pages));
  const toggle = (item: ChequeRegister, checked: boolean) =>
    setChosen((current) => {
      const next = new Map(current);
      if (checked) next.set(item.id, item);
      else next.delete(item.id);
      return next;
    });

  return (
    <div className={overlayClass} role="dialog" aria-modal="true" aria-labelledby="cheque-picker-title">
      <div className="grid max-h-[calc(100dvh-16px)] gap-3 overflow-y-auto rounded-2xl border border-border bg-card p-3 text-card-foreground shadow-xl" style={{ width: "min(780px, calc(100vw - 16px))" }}>
        <header className="flex flex-wrap items-center justify-between gap-2">
          <h2 id="cheque-picker-title" className="text-base font-bold">
            {tr("cheque_picker_title", "เลือกเช็คเข้าเอกสาร")}
          </h2>
          <Badge variant="outline" className="text-[0.9rem]">
            {tr("bank_code_list_count", "ทั้งหมด {count} รายการ").replace("{count}", String(total))}
          </Badge>
        </header>
        <p className="text-[0.9rem] text-muted-foreground">{tr("cheque_picker_hint", "แสดงเฉพาะเช็คที่ใช้กับเอกสารนี้ได้ ณ วันที่เอกสาร ติ๊กเลือกได้หลายใบ")}</p>
        <label className="relative block">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
          <Input
            className="!pl-10 text-[0.95rem]"
            value={search}
            autoFocus
            onChange={(event) => setSearch(event.target.value)}
            placeholder={tr("cheque_picker_search", "ค้นหาเลขที่เช็ค เลขที่เอกสาร หรือรหัสคู่ค้า")}
            aria-label={tr("search", "ค้นหา")}
          />
        </label>

        {loadError ? (
          <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
            {chequeErrorText(loadError, tr)}
          </p>
        ) : loading && items.length === 0 ? (
          <p className="flex items-center gap-2 p-2 text-[0.95rem] text-muted-foreground">
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
            {tr("loading", "กำลังโหลดข้อมูล...")}
          </p>
        ) : items.length === 0 ? (
          <p className="p-2 text-[0.95rem] text-muted-foreground">{tr("cheque_picker_empty", "ไม่มีเช็คที่เลือกได้ในเงื่อนไขนี้")}</p>
        ) : (
          <div className="overflow-x-auto" aria-busy={loading}>
            <table className="w-full min-w-[40rem] border-separate border-spacing-y-1 text-left text-[0.95rem]">
              <thead>
                <tr className="text-[0.9rem] font-semibold text-muted-foreground">
                  <th className="w-10 px-1">
                    <span className="sr-only">{tr("select", "เลือก")}</span>
                  </th>
                  <th className="px-1">{tr("cheque_number", "เลขที่เช็ค")}</th>
                  <th className="px-1">{tr("cheque_col_bank", "ธนาคาร")}</th>
                  <th className="w-32 px-1">{tr("cheque_field_due_date", "วันที่ครบกำหนด")}</th>
                  <th className="w-36 px-1 text-right">{tr("cheque_col_cheque_amount", "จำนวนเงินบนเช็ค")}</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => {
                  const taken = excludeIds.has(item.id);
                  return (
                    <tr key={item.id} className={cn("rounded-lg", chosen.has(item.id) && "bg-primary/10", taken && "text-muted-foreground")}>
                      <td className="px-1">
                        <Checkbox
                          checked={taken || chosen.has(item.id)}
                          disabled={taken}
                          aria-label={`${tr("select", "เลือก")} ${item.cheque_no}`}
                          onCheckedChange={(checked) => toggle(item, checked)}
                        />
                      </td>
                      <td className="px-1 font-semibold">
                        {item.cheque_no}
                        {item.partner_name ? <span className="block text-[0.85rem] font-normal text-muted-foreground">{item.partner_name}</span> : null}
                      </td>
                      <td className="px-1">{bankNames.get(item.bank_code) ?? item.bank_code}</td>
                      <td className="px-1">{formatAppDate(item.due_date, language)}</td>
                      <td className="px-1 text-right tabular-nums">{formatDecimalText(item.amount)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {pages > 1 ? (
          <nav className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-2" aria-label={pageInfo}>
            <Button type="button" variant="outline" disabled={page <= 1 || loading} onClick={() => setPage(page - 1)}>
              {tr("bank_code_prev", "ก่อนหน้า")}
            </Button>
            <span className="text-[0.95rem]">{pageInfo}</span>
            <Button type="button" variant="outline" disabled={page >= pages || loading} onClick={() => setPage(page + 1)}>
              {tr("bank_code_next", "ถัดไป")}
            </Button>
          </nav>
        ) : null}

        <footer className="flex flex-wrap items-center justify-between gap-2">
          <span className="text-[0.95rem] font-semibold">{tr("cheque_picker_chosen", "เลือกแล้ว {count} ใบ").replace("{count}", String(chosen.size))}</span>
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {tr("cancel", "ยกเลิก")}
            </Button>
            <Button type="button" disabled={chosen.size === 0} onClick={() => onPick([...chosen.values()])}>
              {tr("cheque_picker_add", "เพิ่มเช็คที่เลือก")}
            </Button>
          </div>
        </footer>
      </div>
    </div>
  );
}

/** ป้ายชื่อช่องของจอเช็ค: ช่องบังคับมีเครื่องหมาย * (ซ่อนจากโปรแกรมอ่านจอ — ช่องที่ขาดถูกแจ้งตอนบันทึก) */
export function ChequeFieldLabel({ htmlFor, required, children }: { htmlFor?: string; required?: boolean; children: ReactNode }) {
  return (
    <label htmlFor={htmlFor} className="text-[0.95rem] font-semibold">
      {children}
      {required ? <span aria-hidden="true" className="text-destructive"> *</span> : null}
    </label>
  );
}
