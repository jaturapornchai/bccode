"use client";

import { Ban, Loader2, Plus, RefreshCcw, Save, Search, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { ChequeFieldLabel, ChequePickerDialog, ChequeReasonDialog } from "@/components/cheque-doc-dialogs";
import { useBankBooks, useCodeList, withCurrentOption } from "@/components/cheque-lookups";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Combobox } from "@/components/ui/combobox";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import {
  cancelChequeDoc,
  ChequeApiError,
  chequeErrorText,
  createChequeDoc,
  getChequeDoc,
  listChequeDocs,
  type ChequeDoc,
} from "@/lib/cheque-api";
import {
  CHEQUE_CANCEL_TYPES,
  chequeDocConfig,
  chequeGlStatus,
  chequeStatusLabel,
  type ChequeDocConfig,
} from "@/lib/cheque-doc-config";
import {
  CHEQUE_LIMITS,
  addDocLines,
  canSaveDocDraft,
  chequeStamp,
  docBookPolicy,
  docDraftTotals,
  docInputFromDraft,
  docLineNet,
  docRevertStatuses,
  draftFromDoc,
  emptyDocDraft,
  invalidFeeChequeIds,
  isDocDraftDirty,
  moneyInput,
  removeDocLine,
  setDocLineFee,
  type DocDraft,
} from "@/lib/cheque-doc-forms";
import { formatDecimalText, todayIso } from "@/lib/cheque-forms";
import { getAuthSession } from "@/lib/client-auth-session";
import { formatAppDate } from "@/lib/date-time";
import type { LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 50;
const HEADER_FIELDS = ["doc_no", "doc_date", "book_code", "cancel_type", "cancel_passed", "description", "lines"];

type FormError = { field: string; message: string };

/**
 * จอเอกสารเปลี่ยนสถานะเช็ค 419-426 (นำฝาก / รับผ่าน / รับคืน / นำเข้าใหม่ / ยกเลิกเช็ครับ / ขายลด / จ่ายผ่าน / ยกเลิกเช็คจ่าย):
 * จอเดียวขับด้วย config ของเอกสาร — เลือกเช็คจาก registers/selectable, สร้างหรือยกเลิกผ่าน command, แสดงใบสำคัญ GL ที่เกิดจากเอกสาร
 */
export function ChequeStatusDocScreen({ flag, language }: { flag: number; language: LanguageCode }) {
  const config = chequeDocConfig(flag);
  return config ? <ChequeStatusDocBody config={config} language={language} /> : null;
}

function ChequeStatusDocBody({ config, language }: { config: ChequeDocConfig; language: LanguageCode }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });
  const flag = config.flag;

  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<"" | "active" | "cancelled">("");
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<ChequeDoc[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [reloadTick, setReloadTick] = useState(0);
  // selected = เอกสารที่เปิดดูอยู่ (อ่านอย่างเดียว); null = กำลังสร้างเอกสารใหม่
  const [selected, setSelected] = useState<ChequeDoc | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [draft, setDraft] = useState<DocDraft>(() => emptyDocDraft(todayIso()));
  const [baseline, setBaseline] = useState<DocDraft>(draft);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<FormError | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [cancelOpen, setCancelOpen] = useState(false);
  const [cancelBusy, setCancelBusy] = useState(false);
  const [cancelError, setCancelError] = useState("");
  const openSeq = useRef(0);

  const { books, error: booksError } = useBankBooks(language, reloadTick);
  const { codes: banks, error: banksError } = useCodeList("bank", language, reloadTick);
  const lookupError = booksError ?? banksError;
  const bankNames = useMemo(() => new Map(banks.map((bank) => [bank.code, bank.name] as const)), [banks]);

  const creating = selected === null;
  const readOnly = !creating;
  const policy = docBookPolicy(config, draft);
  const dirty = creating && isDocDraftDirty(draft, baseline);
  const feeProblems = useMemo(() => new Set(invalidFeeChequeIds(config, draft.lines)), [config, draft.lines]);
  const canSave = creating && canSaveDocDraft(config, draft);
  const totals = docDraftTotals(config, draft.lines);
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const reload = useCallback(() => setReloadTick((tick) => tick + 1), []);
  const pageInfo = tr("bank_code_page_info", "หน้า {page} จาก {pages}").replace("{page}", String(page)).replace("{pages}", String(pages));
  const [titleKey, titleFallback] = config.title;
  const cancelled = selected?.status === "cancelled";

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
    listChequeDocs(flag, { q: appliedSearch, status: statusFilter || undefined, page, limit: PAGE_SIZE }, language)
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
  }, [flag, appliedSearch, statusFilter, page, language, reloadTick]);

  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const statusLabel = useCallback(
    (status: number) => {
      const [key, fallback] = chequeStatusLabel(config.chequeType, status);
      return tr(key, fallback);
    },
    [config.chequeType, tr],
  );
  const bookOptions = useMemo(
    () =>
      withCurrentOption(
        [
          ...(policy.required ? [] : [{ value: "", label: tr("cheque_book_none", "ไม่ระบุสมุดเงินฝาก") }]),
          ...books.map((book) => ({ value: book.bank_account_code, label: `${book.bank_account_code} — ${book.account_name}` })),
        ],
        draft.book_code,
      ),
    [books, policy.required, draft.book_code, tr],
  );
  const cancelTypeOptions = useMemo(
    () => CHEQUE_CANCEL_TYPES.map(([key, fallback], value) => ({ value: String(value), label: tr(key, fallback) })),
    [tr],
  );
  const statusFilterOptions = useMemo(
    () => [
      { value: "", label: tr("cheque_filter_status_all", "ทุกสถานะ") },
      { value: "active", label: tr("cheque_doc_active", "ใช้งานอยู่") },
      { value: "cancelled", label: tr("cheque_doc_cancelled", "ยกเลิกแล้ว") },
    ],
    [tr],
  );

  const canLeave = useCallback(async () => {
    if (!dirty) return true;
    return confirm({
      title: tr("cheque_confirm_discard_title", "ทิ้งข้อมูลที่ยังไม่ได้บันทึก?"),
      description: tr("cheque_confirm_discard", "มีข้อมูลที่แก้ไขแล้วแต่ยังไม่ได้บันทึก หากไปต่อข้อมูลที่แก้ไว้จะหายไป"),
      confirmLabel: tr("cheque_confirm_discard_ok", "ทิ้งข้อมูลและไปต่อ"),
      cancelLabel: tr("cheque_confirm_keep_editing", "กลับไปแก้ไขต่อ"),
      tone: "warning",
    });
  }, [confirm, dirty, tr]);

  const show = useCallback((doc: ChequeDoc | null) => {
    const next = doc ? draftFromDoc(doc) : emptyDocDraft(todayIso());
    setSelected(doc);
    setDraft(next);
    setBaseline(next);
    setFormError(null);
    setPickerOpen(false);
  }, []);

  const fieldError = (cause: unknown): FormError => ({ field: cause instanceof ChequeApiError ? cause.field : "", message: chequeErrorText(cause, tr) });

  const startNew = async () => {
    if (!(await canLeave())) return;
    openSeq.current += 1;
    setDetailLoading(false);
    show(null);
  };

  // รายการในลิสต์ไม่มีแถวเช็ค จึงโหลดเอกสารเต็มก่อนแสดง (เปิดฟอร์มทันทีด้วยหัวเอกสาร แล้วเติมแถวตามมา)
  const openItem = async (item: ChequeDoc) => {
    if (selected?.doc_no === item.doc_no) return;
    if (!(await canLeave())) return;
    openSeq.current += 1;
    const seq = openSeq.current;
    show(item);
    setDetailLoading(true);
    try {
      const detail = await getChequeDoc(flag, item.doc_no, language);
      if (seq === openSeq.current) show(detail);
    } catch (cause) {
      if (seq === openSeq.current) setFormError(fieldError(cause));
    } finally {
      if (seq === openSeq.current) setDetailLoading(false);
    }
  };

  // เปลี่ยนเงื่อนไขที่กระทบสิทธิ์เลือกเช็ค (ชนิดการยกเลิก / ติ๊กเช็คที่ผ่านแล้ว / สมุด) ขณะมีเช็คอยู่ → ถามก่อนแล้วล้างรายการ
  const changeCondition = async (apply: (current: DocDraft) => DocDraft) => {
    if (draft.lines.length === 0) {
      setDraft(apply);
      return;
    }
    const ok = await confirm({
      title: tr("cheque_change_kind_title", "เปลี่ยนเงื่อนไขการเลือกเช็ค?"),
      description: tr("cheque_change_kind_confirm", "เช็คที่เลือกไว้แล้วอาจใช้กับเงื่อนไขใหม่ไม่ได้ ระบบจะล้างรายการเช็คทั้งหมดแล้วให้เลือกใหม่"),
      confirmLabel: tr("cheque_change_kind_ok", "เปลี่ยนและล้างรายการ"),
      tone: "warning",
    });
    if (ok) setDraft((current) => apply({ ...current, lines: [] }));
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (busy || !canSave) return;
    setBusy(true);
    setFormError(null);
    try {
      const saved = await createChequeDoc(docInputFromDraft(config, draft), language);
      // อ่านเอกสารกลับมาเพื่อให้เห็นเลขที่เอกสาร สถานะ GL และใบสำคัญล่าสุด; อ่านไม่ได้ก็ใช้ผลที่บันทึกแล้ว
      let doc = saved;
      try {
        doc = await getChequeDoc(flag, saved.doc_no, language);
      } catch {
        doc = { ...saved, lines: saved.lines ?? [] };
      }
      show(doc);
      pushNotice({ type: "success", text: tr("cheque_doc_saved", "บันทึกเอกสารเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
      if (cause instanceof ChequeApiError && cause.status === 409) reload();
    } finally {
      setBusy(false);
    }
  };

  const cancelDocument = async (reason: string) => {
    if (!selected || cancelBusy) return;
    setCancelBusy(true);
    setCancelError("");
    try {
      await cancelChequeDoc({ transflag: flag, docNo: selected.doc_no, version: selected.version, reason }, language);
      setCancelOpen(false);
      try {
        show(await getChequeDoc(flag, selected.doc_no, language));
      } catch {
        show({ ...selected, status: "cancelled", cancel_reason: reason });
      }
      pushNotice({ type: "success", text: tr("cheque_doc_cancelled_ok", "ยกเลิกเอกสารเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setCancelError(chequeErrorText(cause, tr));
      if (cause instanceof ChequeApiError && cause.status === 409) reload();
    } finally {
      setCancelBusy(false);
    }
  };

  const cancelDescription = useMemo(() => {
    if (!selected) return "";
    const statuses = docRevertStatuses(selected.lines ?? []).map(statusLabel).join(", ");
    return selected.gl_status === "posted"
      ? tr("cheque_confirm_cancel_doc", "ยกเลิกเอกสารนี้ เช็คจะกลับเป็นสถานะ {status} และกลับรายการบัญชี").replace("{status}", statuses)
      : tr("cheque_confirm_cancel_doc_nogl", "ยกเลิกเอกสารนี้ เช็คจะกลับเป็นสถานะ {status}").replace("{status}", statuses);
  }, [selected, statusLabel, tr]);

  const invalid = (field: string) => formError?.field === field;
  const fieldMessage = (field: string) =>
    invalid(field) ? (
      <span className="text-[0.9rem] font-medium text-destructive" role="alert">
        {formError?.message}
      </span>
    ) : null;
  const readOnlyBox = "flex min-h-10 items-center rounded-xl border border-border bg-muted px-3 text-[0.95rem] font-semibold";
  const pickerReady = Boolean(draft.doc_date) && (!policy.match || Boolean(draft.book_code));
  const gl = selected ? chequeGlStatus(selected.gl_status) : null;
  const remarkColumn = config.showRemark;
  const feeColumns = config.hasFee;
  const colCount = 5 + (feeColumns ? 2 : 0) + (remarkColumn ? 1 : 0) + (readOnly ? 0 : 1);

  return (
    <section className="grid w-full min-w-0 gap-3">
      <header className="flex flex-wrap items-center gap-2">
        <h1 className="text-xl font-bold tracking-tight">{tr(titleKey, titleFallback)}</h1>
        <Badge variant="outline" className="text-[0.9rem]">
          {tr("bank_code_list_count", "ทั้งหมด {count} รายการ").replace("{count}", String(total))}
        </Badge>
        {dirty ? <Badge variant="warning" className="text-[0.85rem]">{tr("cheque_unsaved", "ยังไม่ได้บันทึก")}</Badge> : null}
      </header>

      <div className="grid min-w-0 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
        <Card className="h-fit min-w-0 shadow-sm">
          <CardContent className="grid gap-3 p-3">
            <div className="flex flex-wrap items-center gap-2">
              <label className="relative block min-w-48 flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
                <Input
                  className="!pl-10 text-[0.95rem]"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder={tr("cheque_doc_search_placeholder", "ค้นหาเลขที่เอกสารหรือเลขที่เช็ค")}
                  aria-label={tr("search", "ค้นหา")}
                />
              </label>
              <Button type="button" variant="outline" onClick={reload} disabled={loading}>
                <RefreshCcw className={cn("size-4", loading && "animate-spin")} aria-hidden="true" />
                {tr("refresh", "รีเฟรช")}
              </Button>
            </div>
            <Combobox<string>
              value={statusFilter}
              options={statusFilterOptions}
              aria-label={tr("status", "สถานะ")}
              onChange={(value) => {
                setStatusFilter(value === "active" || value === "cancelled" ? value : "");
                setPage(1);
              }}
            />

            {loadError ? (
              <div className="grid gap-2 rounded-xl border border-destructive/40 bg-destructive/5 p-3" role="alert">
                <p className="text-[0.95rem] font-medium">{chequeErrorText(loadError, tr)}</p>
                <Button type="button" variant="outline" className="w-fit" onClick={reload}>
                  {tr("refresh", "รีเฟรช")}
                </Button>
              </div>
            ) : loading && items.length === 0 ? (
              <p className="flex items-center gap-2 p-3 text-[0.95rem] text-muted-foreground">
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                {tr("loading", "กำลังโหลดข้อมูล...")}
              </p>
            ) : items.length === 0 ? (
              <p className="p-3 text-[0.95rem] text-muted-foreground">
                {appliedSearch || statusFilter ? tr("no_data", "ไม่มีข้อมูล") : tr("cheque_doc_list_empty", "ยังไม่มีเอกสาร กดปุ่มเพิ่มเพื่อสร้างเอกสารใบแรก")}
              </p>
            ) : (
              <ul className="grid gap-0.5" aria-busy={loading}>
                {items.map((item) => (
                  <li key={item.doc_no}>
                    <button
                      type="button"
                      onClick={() => void openItem(item)}
                      aria-current={selected?.doc_no === item.doc_no ? "true" : undefined}
                      className={cn(
                        "bc-list-row grid min-h-9 w-full gap-0.5 rounded-lg px-2 py-1 text-left text-[0.95rem] transition-colors hover:bg-accent",
                        selected?.doc_no === item.doc_no && "bg-primary/10 ring-1 ring-primary/40",
                      )}
                    >
                      <span className="flex flex-nowrap items-center gap-1.5">
                        <span className="min-w-0 flex-1 font-semibold" title={item.doc_no}>
                          <span className="bc-cell-text">{item.doc_no}</span>
                        </span>
                        <span className="w-28 shrink-0">{formatAppDate(item.doc_date, language)}</span>
                        <span className="w-28 shrink-0 text-right tabular-nums">{formatDecimalText(item.total_amount)}</span>
                      </span>
                      {item.status === "cancelled" ? (
                        <span>
                          <Badge variant="warning" className="text-[0.85rem]">{tr("cheque_doc_cancelled", "ยกเลิกแล้ว")}</Badge>
                        </span>
                      ) : null}
                    </button>
                  </li>
                ))}
              </ul>
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
          </CardContent>
        </Card>

        <Card className="h-fit min-w-0 shadow-sm">
          <CardContent className="p-3">
            <form className="grid gap-3" onSubmit={(event) => void submit(event)} noValidate>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-bold">
                  {selected ? tr("cheque_doc_form_view", "เอกสารเลขที่ {doc_no}").replace("{doc_no}", selected.doc_no) : tr("cheque_doc_form_new", "เอกสารใหม่")}
                </h2>
                <Button type="button" variant="outline" onClick={() => void startNew()} disabled={busy}>
                  <Plus className="size-4" aria-hidden="true" />
                  {tr("add", "เพิ่ม")}
                </Button>
              </div>

              {lookupError ? (
                <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
                  {chequeErrorText(lookupError, tr)}
                </p>
              ) : null}

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-doc-no">{tr("cheque_field_doc_no", "เลขที่เอกสาร")}</ChequeFieldLabel>
                  <Input
                    id="cheque-doc-no"
                    value={draft.doc_no}
                    maxLength={CHEQUE_LIMITS.docNo}
                    readOnly={readOnly}
                    aria-invalid={invalid("doc_no")}
                    className={cn("text-[0.95rem]", readOnly && "bg-muted")}
                    onChange={(event) => setDraft((current) => ({ ...current, doc_no: event.target.value }))}
                  />
                  {creating ? <span className="text-[0.85rem] text-muted-foreground">{tr("cheque_doc_no_auto_hint", "เว้นว่างไว้เพื่อให้ระบบออกเลขที่ให้อัตโนมัติ")}</span> : null}
                  {fieldMessage("doc_no")}
                </div>
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-doc-date" required={creating}>{tr("cheque_field_doc_date", "วันที่เอกสาร")}</ChequeFieldLabel>
                  <Input
                    id="cheque-doc-date"
                    type="date"
                    value={draft.doc_date}
                    readOnly={readOnly}
                    aria-invalid={invalid("doc_date")}
                    className={cn("text-[0.95rem]", readOnly && "bg-muted")}
                    onChange={(event) => setDraft((current) => ({ ...current, doc_date: event.target.value }))}
                  />
                  {fieldMessage("doc_date")}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-doc-book" required={creating && policy.required}>
                    {tr("cheque_field_book_code", "รหัสสมุดเงินฝาก")}
                  </ChequeFieldLabel>
                  <Combobox<string>
                    id="cheque-doc-book"
                    value={draft.book_code}
                    options={bookOptions}
                    searchable
                    disabled={readOnly || policy.none}
                    placeholder={tr("cheque_pick_book", "เลือกสมุดเงินฝาก")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("cheque_field_book_code", "รหัสสมุดเงินฝาก")}
                    onChange={(value) => {
                      if (value === draft.book_code) return;
                      if (policy.match) void changeCondition((current) => ({ ...current, book_code: value }));
                      else setDraft((current) => ({ ...current, book_code: value }));
                    }}
                  />
                  {policy.none ? <span className="text-[0.85rem] text-muted-foreground">{tr("cheque_book_not_used", "เอกสารชนิดนี้ไม่ใช้สมุดเงินฝาก")}</span> : null}
                  {fieldMessage("book_code")}
                </div>

                {config.cancelTypeCombo ? (
                  <div className="grid gap-1">
                    <ChequeFieldLabel htmlFor="cheque-doc-cancel-type" required={creating}>
                      {tr("cheque_field_cancel_type", "เช็คที่ต้องการยกเลิก")}
                    </ChequeFieldLabel>
                    <Combobox<string>
                      id="cheque-doc-cancel-type"
                      value={String(draft.cancel_type)}
                      options={cancelTypeOptions}
                      disabled={readOnly}
                      aria-label={tr("cheque_field_cancel_type", "เช็คที่ต้องการยกเลิก")}
                      onChange={(value) => {
                        const next = Number(value);
                        if (next === draft.cancel_type) return;
                        // ชนิด 0 ไม่ใช้สมุด จึงล้างสมุดที่เลือกไว้ด้วย
                        void changeCondition((current) => ({ ...current, cancel_type: next, book_code: next === 0 ? "" : current.book_code }));
                      }}
                    />
                    {fieldMessage("cancel_type")}
                  </div>
                ) : null}

                {config.cancelPassedBox ? (
                  <div className="flex items-end gap-2 pb-2">
                    <Checkbox
                      id="cheque-doc-cancel-passed"
                      checked={draft.cancel_passed}
                      disabled={readOnly}
                      onCheckedChange={(checked) => void changeCondition((current) => ({ ...current, cancel_passed: checked === true }))}
                    />
                    <label htmlFor="cheque-doc-cancel-passed" className="text-[0.95rem] font-semibold">
                      {tr("cheque_field_cancel_passed", "ยกเลิกเช็คที่ผ่านแล้ว")}
                    </label>
                  </div>
                ) : null}
              </div>

              {config.cancelTypeCombo ? (
                <p className="text-[0.95rem] font-semibold text-destructive" role="note">
                  {tr("cheque_cancel_notice", "***เช็คที่ยกเลิกไปแล้ว ไม่สามารถที่นำไปทำรายการได้อีก (สถานะเป็นเช็คยกเลิก)***")}
                </p>
              ) : null}

              <div className="grid gap-1">
                <ChequeFieldLabel htmlFor="cheque-doc-description">{tr("cheque_field_note", "หมายเหตุ")}</ChequeFieldLabel>
                <Input
                  id="cheque-doc-description"
                  value={draft.description}
                  maxLength={CHEQUE_LIMITS.description}
                  readOnly={readOnly}
                  aria-invalid={invalid("description")}
                  className={cn("text-[0.95rem]", readOnly && "bg-muted")}
                  onChange={(event) => setDraft((current) => ({ ...current, description: event.target.value }))}
                />
                {fieldMessage("description")}
              </div>

              <div className="grid gap-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <h3 className="text-[0.95rem] font-bold">{tr("cheque_lines_title", "รายการเช็ค")}</h3>
                  {creating ? (
                    <Button type="button" variant="outline" disabled={!pickerReady || busy} onClick={() => setPickerOpen(true)}>
                      <Plus className="size-4" aria-hidden="true" />
                      {tr("cheque_btn_pick", "เลือกเช็ค")}
                    </Button>
                  ) : null}
                </div>
                {creating && policy.match && !draft.book_code ? (
                  <p className="text-[0.9rem] text-muted-foreground">{tr("cheque_pick_need_book_hint", "เลือกสมุดเงินฝากก่อน ระบบจะแสดงเฉพาะเช็คของสมุดนั้น")}</p>
                ) : null}

                <div className="overflow-x-auto">
                  <table className="w-full min-w-[40rem] border-separate border-spacing-y-1 text-left text-[0.95rem]">
                    <thead>
                      <tr className="text-[0.9rem] font-semibold text-muted-foreground">
                        <th className="px-1">{tr("cheque_number", "เลขที่เช็ค")}</th>
                        <th className="px-1">{tr("cheque_col_bank", "ธนาคาร")}</th>
                        <th className="w-32 px-1">{tr("cheque_field_due_date", "วันที่ครบกำหนด")}</th>
                        <th className="w-32 px-1 text-right">{tr("cheque_col_cheque_amount", "จำนวนเงินบนเช็ค")}</th>
                        <th className="w-32 px-1 text-right">{tr("cheque_col_amount_baht", "จำนวนเงิน (บาท)")}</th>
                        {feeColumns ? <th className="w-32 px-1 text-right">{tr("cheque_col_fee", "ค่าใช้จ่าย")}</th> : null}
                        {feeColumns ? <th className="w-36 px-1 text-right">{tr("cheque_col_net", "ยอดที่หักค่าใช้จ่ายแล้ว")}</th> : null}
                        {remarkColumn ? <th className="px-1">{tr("cheque_field_note", "หมายเหตุ")}</th> : null}
                        {readOnly ? null : (
                          <th className="w-12 px-1">
                            <span className="sr-only">{tr("cheque_remove_line", "เอาเช็คออกจากเอกสาร")}</span>
                          </th>
                        )}
                      </tr>
                    </thead>
                    <tbody>
                      {detailLoading && draft.lines.length === 0 ? (
                        <tr>
                          <td colSpan={colCount} className="px-1 text-muted-foreground">
                            <span className="flex items-center gap-2">
                              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                              {tr("loading", "กำลังโหลดข้อมูล...")}
                            </span>
                          </td>
                        </tr>
                      ) : draft.lines.length === 0 ? (
                        <tr>
                          <td colSpan={colCount} className={cn("px-1 text-muted-foreground", invalid("lines") && "font-medium text-destructive")}>
                            {tr("cheque_lines_empty", "ยังไม่ได้เลือกเช็ค กดปุ่มเลือกเช็คเพื่อเพิ่มเช็คเข้าเอกสาร")}
                          </td>
                        </tr>
                      ) : (
                        draft.lines.map((line) => (
                          <tr key={line.cheque_id}>
                            <td className="px-1 font-semibold">
                              {line.cheque_no}
                              {line.partner_name ? <span className="block text-[0.85rem] font-normal text-muted-foreground">{line.partner_name}</span> : null}
                            </td>
                            <td className="px-1">{bankNames.get(line.bank_code) ?? line.bank_code}</td>
                            <td className="px-1">{formatAppDate(line.due_date, language)}</td>
                            <td className="px-1 text-right tabular-nums">{formatDecimalText(line.cheque_amount)}</td>
                            <td className="px-1 text-right tabular-nums">{formatDecimalText(line.amount)}</td>
                            {feeColumns ? (
                              <td className="px-1 text-right">
                                {readOnly ? (
                                  <span className="tabular-nums">{formatDecimalText(line.fee_amount || "0")}</span>
                                ) : (
                                  <Input
                                    inputMode="decimal"
                                    value={line.fee_amount}
                                    maxLength={19}
                                    aria-label={`${tr("cheque_col_fee", "ค่าใช้จ่าย")} ${line.cheque_no}`}
                                    aria-invalid={feeProblems.has(line.cheque_id)}
                                    className="text-right text-[0.95rem] tabular-nums"
                                    onChange={(event) =>
                                      setDraft((current) => ({ ...current, lines: setDocLineFee(current.lines, line.cheque_id, moneyInput(event.target.value, line.fee_amount)) }))
                                    }
                                  />
                                )}
                                {feeProblems.has(line.cheque_id) ? (
                                  <span className="block text-[0.85rem] font-medium text-destructive" role="alert">
                                    {tr("cheque_fee_invalid_hint", "ค่าใช้จ่ายต้องไม่เกินจำนวนเงินของเช็ค")}
                                  </span>
                                ) : null}
                              </td>
                            ) : null}
                            {feeColumns ? <td className="px-1 text-right tabular-nums">{formatDecimalText(docLineNet(line))}</td> : null}
                            {remarkColumn ? <td className="px-1">{line.description}</td> : null}
                            {readOnly ? null : (
                              <td className="px-1">
                                <Button
                                  type="button"
                                  variant="outline"
                                  className="px-2"
                                  aria-label={`${tr("cheque_remove_line", "เอาเช็คออกจากเอกสาร")} ${line.cheque_no}`}
                                  title={tr("cheque_remove_line", "เอาเช็คออกจากเอกสาร")}
                                  onClick={() => setDraft((current) => ({ ...current, lines: removeDocLine(current.lines, line.cheque_id) }))}
                                >
                                  <Trash2 className="size-4" aria-hidden="true" />
                                </Button>
                              </td>
                            )}
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>

                <dl className="ml-auto grid w-full gap-1 sm:w-80">
                  <div className="flex items-center justify-between gap-3">
                    <dt className="text-[0.95rem] font-semibold">{tr("cheque_sum_total", "จำนวนเงินรวม")}</dt>
                    <dd className="text-[0.95rem] font-bold tabular-nums">{formatDecimalText(totals.total)}</dd>
                  </div>
                  {feeColumns ? (
                    <>
                      <div className="flex items-center justify-between gap-3">
                        <dt className="text-[0.95rem] font-semibold">{tr("cheque_sum_fee", "รวมค่าใช้จ่าย")}</dt>
                        <dd className="text-[0.95rem] font-bold tabular-nums">{formatDecimalText(totals.fee)}</dd>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <dt className="text-[0.95rem] font-semibold">{tr("cheque_col_net", "ยอดที่หักค่าใช้จ่ายแล้ว")}</dt>
                        <dd className="text-[0.95rem] font-bold tabular-nums">{formatDecimalText(totals.net)}</dd>
                      </div>
                    </>
                  ) : null}
                </dl>
              </div>

              {selected && gl ? (
                <div className="grid gap-1 rounded-xl border border-border bg-muted/50 p-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[0.95rem] font-semibold">{tr("cheque_gl_journal", "ใบสำคัญบัญชี")}</span>
                    <Badge variant={gl.tone} className="text-[0.85rem]">{tr(gl.label[0], gl.label[1])}</Badge>
                    {selected.journal_docno ? <span className="text-[0.95rem] font-semibold">{selected.journal_docno}</span> : null}
                  </div>
                  <span className="text-[0.85rem] text-muted-foreground">
                    {tr("cheque_stamp_created", "สร้างโดย")} {chequeStamp(language, selected.created_by, selected.created_at)}
                  </span>
                </div>
              ) : null}

              {cancelled && selected ? (
                <div className="grid gap-1 rounded-xl border border-destructive/30 bg-destructive/5 p-2" role="status">
                  <Badge variant="warning" className="w-fit text-[0.85rem]">{tr("cheque_doc_cancelled", "ยกเลิกแล้ว")}</Badge>
                  <span className="text-[0.95rem]">
                    {tr("cheque_doc_cancel_info", "ยกเลิกเอกสารโดย")} {chequeStamp(language, selected.cancelled_by, selected.cancelled_at ?? undefined)}
                  </span>
                  <span className="text-[0.95rem]">
                    {tr("cheque_cancel_reason_label", "เหตุผลที่ยกเลิก")}: {selected.cancel_reason}
                  </span>
                </div>
              ) : null}

              {formError && (formError.field === "lines" || !HEADER_FIELDS.includes(formError.field)) ? (
                <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
                  {formError.message}
                </p>
              ) : null}

              <div className="flex flex-wrap items-center gap-2">
                {creating ? (
                  <Button type="submit" disabled={busy || !canSave}>
                    {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : <Save className="size-4" aria-hidden="true" />}
                    {tr("save", "บันทึก")}
                  </Button>
                ) : null}
                {selected && !cancelled ? (
                  <Button
                    type="button"
                    variant="outline"
                    className="text-destructive"
                    disabled={detailLoading || busy}
                    onClick={() => {
                      setCancelError("");
                      setCancelOpen(true);
                    }}
                  >
                    <Ban className="size-4" aria-hidden="true" />
                    {tr("cheque_doc_cancel_btn", "ยกเลิกเอกสาร")}
                  </Button>
                ) : null}
                {creating && !canSave ? (
                  <span className="text-[0.9rem] text-muted-foreground">
                    {tr("cheque_doc_lines_required", "เลือกเช็คอย่างน้อยหนึ่งใบ และกรอกช่องที่มีเครื่องหมาย * ให้ครบก่อนบันทึก")}
                  </span>
                ) : null}
              </div>
            </form>
          </CardContent>
        </Card>
      </div>

      {pickerOpen ? (
        <ChequePickerDialog
          tr={tr}
          language={language}
          filter={{
            transflag: flag,
            docDate: draft.doc_date,
            book: policy.none ? "" : draft.book_code,
            cancelType: config.cancelTypeCombo ? draft.cancel_type : null,
            cancelPassed: config.cancelPassedBox ? draft.cancel_passed : false,
          }}
          excludeIds={new Set(draft.lines.map((line) => line.cheque_id))}
          bankNames={bankNames}
          onPick={(picked) => {
            setDraft((current) => ({ ...current, lines: addDocLines(current.lines, picked) }));
            setPickerOpen(false);
          }}
          onClose={() => setPickerOpen(false)}
        />
      ) : null}

      {cancelOpen && selected ? (
        <ChequeReasonDialog
          tr={tr}
          title={tr("cheque_cancel_dialog_title", "ยกเลิกเอกสารเลขที่ {doc_no}").replace("{doc_no}", selected.doc_no)}
          description={cancelDescription}
          busy={cancelBusy}
          error={cancelError}
          onConfirm={(reason) => void cancelDocument(reason)}
          onCancel={() => setCancelOpen(false)}
        />
      ) : null}
      {confirmationDialog}
    </section>
  );
}
