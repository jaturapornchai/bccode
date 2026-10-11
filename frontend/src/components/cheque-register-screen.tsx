"use client";

import { Loader2, Plus, RefreshCcw, Save, Search, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { ChequeFieldLabel } from "@/components/cheque-doc-dialogs";
import { useBankBooks, useCodeList, withCurrentOption } from "@/components/cheque-lookups";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Combobox } from "@/components/ui/combobox";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import {
  cancelChequeRegister,
  ChequeApiError,
  chequeErrorText,
  createChequeRegister,
  getChequeRegister,
  listChequeRegisters,
  updateChequeRegister,
  type ChequeReference,
  type ChequeRegister,
} from "@/lib/cheque-api";
import {
  CHEQUE_REGISTER_TITLES,
  CHEQUE_TYPE_ISSUED,
  chequeDocTitleByKind,
  chequeStatusLabel,
  chequeStatusTone,
  registerStatusChoices,
  type ChequeKind,
} from "@/lib/cheque-doc-config";
import {
  CHEQUE_LIMITS,
  canSaveRegisterDraft,
  chequeStamp,
  draftFromRegister,
  emptyRegisterDraft,
  isRegisterDraftDirty,
  moneyInput,
  registerBookRequired,
  registerHomeAmount,
  registerInputFromDraft,
  type RegisterDraft,
} from "@/lib/cheque-doc-forms";
import { formatDecimalText, todayIso } from "@/lib/cheque-forms";
import { getAuthSession } from "@/lib/client-auth-session";
import { formatAppDate } from "@/lib/date-time";
import type { LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 50;
const FORM_FIELDS = [
  "doc_no", "doc_date", "partner_code", "due_date", "status", "cheque_no", "ext_status_code", "bank_code",
  "bank_branch_code", "book_code", "amount", "description", "sale_code", "department_code",
];
// D8: backend ถามยืนยันก่อน (HTTP 409 + รหัสเหล่านี้) แล้วรับคำสั่งเดิมซ้ำพร้อม confirm:true
const CONFIRM_CODES = ["cheque_err_referenced", "cheque_err_status_confirm"];

type FormError = { field: string; message: string };

/**
 * จอ "บันทึกเช็ครับ / บันทึกเช็คจ่าย" (Champ 5377/5378, cqfrmcheqin.cpp / cqfrmcheqout.cpp): รายการเช็ค + ฟอร์มตามลำดับช่องของจอจริง
 * ตารางเอกสารอ้างอิงใต้ฟอร์มอ่านอย่างเดียว; ปุ่ม "เช็คเตรียมจ่าย" ของเช็คจ่ายจะมากับ CB-9 (ต้องมี /prepare/unused ก่อน)
 */
export function ChequeRegisterScreen({ chequeType, language }: { chequeType: ChequeKind; language: LanguageCode }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });
  const issued = chequeType === CHEQUE_TYPE_ISSUED;

  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<ChequeRegister[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [reloadTick, setReloadTick] = useState(0);
  const [selected, setSelected] = useState<ChequeRegister | null>(null);
  const [references, setReferences] = useState<ChequeReference[]>([]);
  const [detailLoading, setDetailLoading] = useState(false);
  const [draft, setDraft] = useState<RegisterDraft>(() => emptyRegisterDraft(todayIso()));
  const [baseline, setBaseline] = useState<RegisterDraft>(draft);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<FormError | null>(null);
  const openSeq = useRef(0);

  const { books, error: booksError } = useBankBooks(language, reloadTick);
  const { codes: banks, error: banksError } = useCodeList("bank", language, reloadTick);
  const { codes: branches } = useCodeList("bank_branch", language, reloadTick, draft.bank_code);
  const { codes: extStatuses } = useCodeList("cheque_ext_status", language, reloadTick);
  const lookupError = booksError ?? banksError;

  const dirty = isRegisterDraftDirty(draft, baseline);
  const canSave = canSaveRegisterDraft(chequeType, draft, baseline);
  const bookRequired = registerBookRequired(chequeType, draft.status);
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const reload = useCallback(() => setReloadTick((tick) => tick + 1), []);
  const pageInfo = tr("bank_code_page_info", "หน้า {page} จาก {pages}").replace("{page}", String(page)).replace("{pages}", String(pages));
  const [titleKey, titleFallback] = CHEQUE_REGISTER_TITLES[chequeType];

  // ค้นหาแบบหน่วงเวลา: พิมพ์ต่อเนื่องไม่ยิงคำขอทุกตัวอักษร และกลับไปหน้า 1 เสมอเมื่อคำค้นเปลี่ยน
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
    listChequeRegisters({ type: chequeType, status: statusFilter === "" ? null : Number(statusFilter), q: appliedSearch, page, limit: PAGE_SIZE }, language)
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
  }, [chequeType, statusFilter, appliedSearch, page, language, reloadTick]);

  // กันปิดแท็บ/รีเฟรชเบราว์เซอร์ทิ้งข้อมูลที่ยังไม่บันทึก
  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const statusLabel = useCallback(
    (status: number) => {
      const [key, fallback] = chequeStatusLabel(chequeType, status);
      return tr(key, fallback);
    },
    [chequeType, tr],
  );
  const statusOptions = useMemo(
    () => registerStatusChoices(chequeType, selected === null).map((status) => ({ value: String(status), label: statusLabel(status) })),
    [chequeType, selected, statusLabel],
  );
  const filterOptions = useMemo(
    () => [
      { value: "", label: tr("cheque_filter_status_all", "ทุกสถานะ") },
      ...registerStatusChoices(chequeType, false).map((status) => ({ value: String(status), label: statusLabel(status) })),
    ],
    [chequeType, statusLabel, tr],
  );
  const bankOptions = useMemo(
    () => withCurrentOption(banks.map((bank) => ({ value: bank.code, label: `${bank.code} — ${bank.name}` })), draft.bank_code),
    [banks, draft.bank_code],
  );
  const branchOptions = useMemo(
    () => withCurrentOption(branches.map((branch) => ({ value: branch.code, label: `${branch.code} — ${branch.name}` })), draft.bank_branch_code),
    [branches, draft.bank_branch_code],
  );
  const extStatusOptions = useMemo(
    () =>
      withCurrentOption(
        [{ value: "", label: tr("cheque_ext_status_none", "ไม่ระบุ") }, ...extStatuses.map((code) => ({ value: code.code, label: `${code.code} — ${code.name}` }))],
        draft.ext_status_code,
      ),
    [extStatuses, draft.ext_status_code, tr],
  );
  const bookOptions = useMemo(
    () =>
      withCurrentOption(
        [
          ...(bookRequired ? [] : [{ value: "", label: tr("cheque_book_none", "ไม่ระบุสมุดเงินฝาก") }]),
          ...books.map((book) => ({ value: book.bank_account_code, label: `${book.bank_account_code} — ${book.account_name}` })),
        ],
        draft.book_code,
      ),
    [books, bookRequired, draft.book_code, tr],
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

  const showForm = useCallback((register: ChequeRegister | null, refs: ChequeReference[]) => {
    const next = register ? draftFromRegister(register) : emptyRegisterDraft(todayIso());
    setSelected(register);
    setReferences(refs);
    setDraft(next);
    setBaseline(next);
    setFormError(null);
  }, []);

  const fieldError = (cause: unknown): FormError => ({ field: cause instanceof ChequeApiError ? cause.field : "", message: chequeErrorText(cause, tr) });

  const startNew = async () => {
    if (!(await canLeave())) return;
    openSeq.current += 1;
    setDetailLoading(false);
    showForm(null, []);
  };

  // เปิดเช็คจากรายการทันที แล้วโหลดรายละเอียด (ตารางเอกสารอ้างอิง + version ล่าสุด) ตามมา
  const openItem = async (item: ChequeRegister) => {
    if (selected?.id === item.id) return;
    if (!(await canLeave())) return;
    openSeq.current += 1;
    const seq = openSeq.current;
    showForm(item, []);
    setDetailLoading(true);
    try {
      const detail = await getChequeRegister(item.id, language);
      if (seq !== openSeq.current) return;
      setSelected(detail);
      setReferences(detail.references);
    } catch (cause) {
      if (seq === openSeq.current) setFormError(fieldError(cause));
    } finally {
      if (seq === openSeq.current) setDetailLoading(false);
    }
  };

  // หลังบันทึก: อ่านเช็คกลับมา (ได้ชื่อคู่ค้า ยอดบาท และเอกสารอ้างอิงล่าสุด); อ่านไม่ได้ก็ไม่ทำให้การบันทึกที่สำเร็จแล้วดูเหมือนล้มเหลว
  const showSaved = async (saved: ChequeRegister) => {
    try {
      const detail = await getChequeRegister(saved.id, language);
      showForm(detail, detail.references);
    } catch {
      showForm(saved, selected?.id === saved.id ? references : []);
    }
  };

  // D8: ถ้า backend ถามยืนยัน (เช็คถูกอ้างอิงแล้ว / เปลี่ยนสถานะตรง) แสดงข้อความของ backend แล้วส่งซ้ำพร้อม confirm; ผู้ใช้ไม่ยืนยัน = null
  const saveRegister = async (): Promise<ChequeRegister | null> => {
    const input = registerInputFromDraft(chequeType, draft);
    const send = (confirmed: boolean) =>
      selected ? updateChequeRegister({ ...input, id: selected.id, version: selected.version }, confirmed, language) : createChequeRegister(input, language);
    try {
      return await send(false);
    } catch (cause) {
      if (!(cause instanceof ChequeApiError) || cause.status !== 409 || !CONFIRM_CODES.includes(cause.code)) throw cause;
      const ok = await confirm({
        title: tr("cheque_confirm_save_title", "ยืนยันการบันทึก"),
        description: chequeErrorText(cause, tr),
        confirmLabel: tr("cheque_confirm_save_ok", "ยืนยันและบันทึก"),
        tone: "warning",
      });
      return ok ? send(true) : null;
    }
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (busy || detailLoading || !canSave) return;
    setBusy(true);
    setFormError(null);
    try {
      const saved = await saveRegister();
      if (!saved) return;
      await showSaved(saved);
      pushNotice({ type: "success", text: tr("cheque_register_saved", "บันทึกเช็คเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
      // เช็คถูกคนอื่นแก้ไขไปก่อน: โหลดรายการใหม่ให้เห็นค่าล่าสุด
      if (cause instanceof ChequeApiError && cause.status === 409) reload();
    } finally {
      setBusy(false);
    }
  };

  const remove = async () => {
    if (!selected || busy) return;
    const ok = await confirm({
      title: tr("confirm_delete", "ต้องการลบข้อมูลหรือไม่"),
      description: tr(
        "cheque_confirm_delete_register",
        "ลบเช็คเลขที่ {cheque_no} ออกจากทะเบียน เช็คที่ถูกใช้ในเอกสารอื่นหรือสร้างจากเอกสารต้นทางจะลบไม่ได้ และเมื่อลบแล้วจะเลือกเช็คใบนี้ในเอกสารอื่นไม่ได้อีก",
      ).replace("{cheque_no}", selected.cheque_no),
      confirmLabel: tr("delete", "ลบ"),
      tone: "danger",
    });
    if (!ok) return;
    setBusy(true);
    setFormError(null);
    try {
      await cancelChequeRegister({ chequeType, id: selected.id, version: selected.version }, language);
      showForm(null, []);
      pushNotice({ type: "success", text: tr("cheque_register_deleted", "ลบเช็คออกจากทะเบียนเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
      if (cause instanceof ChequeApiError && cause.status === 409) reload();
    } finally {
      setBusy(false);
    }
  };

  const invalid = (field: string) => formError?.field === field;
  const fieldMessage = (field: string) =>
    invalid(field) ? (
      <span className="text-[0.9rem] font-medium text-destructive" role="alert">
        {formError?.message}
      </span>
    ) : null;
  const set = <K extends keyof RegisterDraft>(field: K, value: RegisterDraft[K]) => setDraft((current) => ({ ...current, [field]: value }));
  const readOnlyBox = "flex min-h-10 items-center rounded-xl border border-border bg-muted px-3 text-[0.95rem] font-semibold";

  return (
    <section className="grid w-full min-w-0 gap-3">
      <header className="flex flex-wrap items-center gap-2">
        <h1 className="text-xl font-bold tracking-tight">{tr(titleKey, titleFallback)}</h1>
        <Badge variant="outline" className="text-[0.9rem]">
          {tr("bank_code_list_count", "ทั้งหมด {count} รายการ").replace("{count}", String(total))}
        </Badge>
        {dirty ? <Badge variant="warning" className="text-[0.85rem]">{tr("cheque_unsaved", "ยังไม่ได้บันทึก")}</Badge> : null}
      </header>

      <div className="grid min-w-0 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.7fr)]">
        <Card className="h-fit min-w-0 shadow-sm">
          <CardContent className="grid gap-3 p-3">
            <div className="flex flex-wrap items-center gap-2">
              <label className="relative block min-w-48 flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
                <Input
                  className="!pl-10 text-[0.95rem]"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder={tr("cheque_register_search_placeholder", "ค้นหาเลขที่เช็ค เลขที่เอกสาร หรือรหัสคู่ค้า")}
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
              options={filterOptions}
              aria-label={tr("status", "สถานะ")}
              onChange={(value) => {
                setStatusFilter(value);
                setPage(1);
              }}
            />

            <div className="bc-list-header flex flex-nowrap items-center gap-1.5 border-b border-border px-2 pb-1 text-[0.9rem] font-semibold text-muted-foreground">
              <span className="min-w-0 flex-1">{tr("cheque_number", "เลขที่เช็ค")}</span>
              <span className="w-28 shrink-0">{tr("cheque_field_due_date", "วันที่ครบกำหนด")}</span>
              <span className="w-28 shrink-0 text-right">{tr("cheque_ref_col_amount", "จำนวนเงิน")}</span>
            </div>

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
                {appliedSearch || statusFilter ? tr("no_data", "ไม่มีข้อมูล") : tr("cheque_register_list_empty", "ยังไม่มีเช็คในทะเบียน กดปุ่มเพิ่มเพื่อบันทึกเช็คใบแรก")}
              </p>
            ) : (
              <ul className="grid gap-0.5" aria-busy={loading}>
                {items.map((item) => (
                  <li key={item.id}>
                    <button
                      type="button"
                      onClick={() => void openItem(item)}
                      aria-current={selected?.id === item.id ? "true" : undefined}
                      className={cn(
                        "bc-list-row grid min-h-9 w-full gap-0.5 rounded-lg px-2 py-1 text-left text-[0.95rem] transition-colors hover:bg-accent",
                        selected?.id === item.id && "bg-primary/10 ring-1 ring-primary/40",
                      )}
                    >
                      <span className="flex flex-nowrap items-center gap-1.5">
                        <span className="min-w-0 flex-1 font-semibold" title={item.cheque_no}>
                          <span className="bc-cell-text">{item.cheque_no}</span>
                        </span>
                        <span className="w-28 shrink-0">{formatAppDate(item.due_date, language)}</span>
                        <span className="w-28 shrink-0 text-right tabular-nums">{formatDecimalText(item.amount)}</span>
                      </span>
                      <span className="flex flex-wrap items-center gap-1.5 text-[0.9rem] text-muted-foreground">
                        <span>{[item.partner_code, item.partner_name].filter(Boolean).join(" ")}</span>
                        <Badge variant={chequeStatusTone(chequeType, item.status)} className="text-[0.85rem]">
                          {statusLabel(item.status)}
                        </Badge>
                      </span>
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
                <h2 className="text-base font-bold">{selected ? tr("cheque_register_form_edit", "แก้ไขเช็ค") : tr("cheque_register_form_new", "เช็คใหม่")}</h2>
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
                    readOnly={selected !== null}
                    aria-invalid={invalid("doc_no")}
                    className={cn("text-[0.95rem]", selected && "bg-muted")}
                    onChange={(event) => set("doc_no", event.target.value)}
                  />
                  {selected ? null : <span className="text-[0.85rem] text-muted-foreground">{tr("cheque_doc_no_auto_hint", "เว้นว่างไว้เพื่อให้ระบบออกเลขที่ให้อัตโนมัติ")}</span>}
                  {fieldMessage("doc_no")}
                </div>
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-doc-date" required>
                    {issued ? tr("cheque_field_pay_date", "วันที่จ่ายเช็ค") : tr("cheque_field_receive_date", "วันที่รับเช็ค")}
                  </ChequeFieldLabel>
                  <Input id="cheque-doc-date" type="date" value={draft.doc_date} aria-invalid={invalid("doc_date")} className="text-[0.95rem]" onChange={(event) => set("doc_date", event.target.value)} />
                  {fieldMessage("doc_date")}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-partner" required>
                    {issued ? tr("cheque_field_ap_code", "รหัสเจ้าหนี้") : tr("cheque_field_ar_code", "รหัสลูกค้า")}
                  </ChequeFieldLabel>
                  <Input
                    id="cheque-partner"
                    value={draft.partner_code}
                    maxLength={CHEQUE_LIMITS.refCode}
                    aria-invalid={invalid("partner_code")}
                    className="text-[0.95rem]"
                    onChange={(event) => set("partner_code", event.target.value)}
                  />
                  <span className="text-[0.85rem] text-muted-foreground">{tr("cheque_partner_hint", "พิมพ์รหัสให้ตรงกับที่กำหนดไว้ ระบบจะแสดงชื่อให้หลังบันทึก")}</span>
                  {fieldMessage("partner_code")}
                </div>
                <div className="grid gap-1">
                  <span id="cheque-partner-name-label" className="text-[0.95rem] font-semibold">
                    {issued ? tr("cheque_field_ap_name", "ชื่อเจ้าหนี้") : tr("cheque_field_ar_name", "ชื่อลูกค้า")}
                  </span>
                  <output aria-labelledby="cheque-partner-name-label" className={readOnlyBox}>
                    {selected?.partner_name ?? ""}
                  </output>
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-due-date" required>{tr("cheque_field_due_date", "วันที่ครบกำหนด")}</ChequeFieldLabel>
                  <Input id="cheque-due-date" type="date" value={draft.due_date} aria-invalid={invalid("due_date")} className="text-[0.95rem]" onChange={(event) => set("due_date", event.target.value)} />
                  {fieldMessage("due_date")}
                </div>
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-status" required>
                    {issued ? tr("cheque_field_status_out", "สถานะเช็คจ่าย") : tr("cheque_field_status_in", "สถานะเช็ครับ")}
                  </ChequeFieldLabel>
                  <Combobox<string>
                    id="cheque-status"
                    value={String(draft.status)}
                    options={statusOptions}
                    aria-label={issued ? tr("cheque_field_status_out", "สถานะเช็คจ่าย") : tr("cheque_field_status_in", "สถานะเช็ครับ")}
                    onChange={(value) => set("status", Number(value))}
                  />
                  {fieldMessage("status")}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-no" required>
                    {issued ? tr("cheque_field_cheque_no_out", "เลขที่เช็คจ่าย") : tr("cheque_field_cheque_no_in", "เลขที่เช็ครับ")}
                  </ChequeFieldLabel>
                  <Input
                    id="cheque-no"
                    value={draft.cheque_no}
                    maxLength={CHEQUE_LIMITS.chequeNo}
                    aria-invalid={invalid("cheque_no")}
                    className="text-[0.95rem]"
                    onChange={(event) => set("cheque_no", event.target.value)}
                  />
                  {fieldMessage("cheque_no")}
                </div>
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-ext-status">{tr("cheque_field_ext_status", "สถานะเพิ่มเติม")}</ChequeFieldLabel>
                  <Combobox<string>
                    id="cheque-ext-status"
                    value={draft.ext_status_code}
                    options={extStatusOptions}
                    searchable
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("cheque_field_ext_status", "สถานะเพิ่มเติม")}
                    onChange={(value) => set("ext_status_code", value)}
                  />
                  {fieldMessage("ext_status_code")}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-bank" required>{tr("bank_code", "รหัสธนาคาร")}</ChequeFieldLabel>
                  <Combobox<string>
                    id="cheque-bank"
                    value={draft.bank_code}
                    options={bankOptions}
                    searchable
                    placeholder={tr("bank_contact_pick_bank", "เลือกธนาคาร")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("bank_code", "รหัสธนาคาร")}
                    // เปลี่ยนธนาคารแล้วสาขาเดิมอาจไม่ใช่ของธนาคารใหม่ จึงล้างให้เลือกสาขาใหม่
                    onChange={(value) => setDraft((current) => (value === current.bank_code ? current : { ...current, bank_code: value, bank_branch_code: "" }))}
                  />
                  {fieldMessage("bank_code")}
                </div>
                <div className="grid gap-1">
                  <span id="cheque-currency-label" className="text-[0.95rem] font-semibold">{tr("cheque_field_currency", "รหัสสกุลเงิน")}</span>
                  <output aria-labelledby="cheque-currency-label" className={readOnlyBox}>THB</output>
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-branch" required>{tr("cheque_field_bank_branch", "สาขาของธนาคาร")}</ChequeFieldLabel>
                  <Combobox<string>
                    id="cheque-branch"
                    value={draft.bank_branch_code}
                    options={branchOptions}
                    searchable
                    disabled={!draft.bank_code}
                    placeholder={tr("bank_contact_pick_branch", "เลือกสาขา")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("cheque_field_bank_branch", "สาขาของธนาคาร")}
                    onChange={(value) => set("bank_branch_code", value)}
                  />
                  {fieldMessage("bank_branch_code")}
                </div>
                <div className="grid gap-1">
                  <span id="cheque-rate-label" className="text-[0.95rem] font-semibold">{tr("cheque_field_rate", "อัตราแลกเปลี่ยน")}</span>
                  <output aria-labelledby="cheque-rate-label" className={cn(readOnlyBox, "justify-end tabular-nums")}>1.00</output>
                  <span className="text-[0.85rem] text-muted-foreground">{tr("cheque_currency_thb_only", "ตอนนี้รับเฉพาะเงินบาท อัตราแลกเปลี่ยน 1")}</span>
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-book" required={bookRequired}>
                    {issued ? tr("cheque_field_book_out", "สมุดเงินฝาก") : tr("cheque_field_book_in", "สมุดเงินฝากธนาคาร")}
                  </ChequeFieldLabel>
                  <Combobox<string>
                    id="cheque-book"
                    value={draft.book_code}
                    options={bookOptions}
                    searchable
                    placeholder={tr("cheque_pick_book", "เลือกสมุดเงินฝาก")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={issued ? tr("cheque_field_book_out", "สมุดเงินฝาก") : tr("cheque_field_book_in", "สมุดเงินฝากธนาคาร")}
                    onChange={(value) => set("book_code", value)}
                  />
                  {fieldMessage("book_code")}
                </div>
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-amount" required>{tr("cheque_field_amount_currency", "จำนวนเงิน (ตามสกุลเงิน)")}</ChequeFieldLabel>
                  <Input
                    id="cheque-amount"
                    inputMode="decimal"
                    value={draft.amount}
                    maxLength={19}
                    aria-invalid={invalid("amount")}
                    className="text-right text-[0.95rem] tabular-nums"
                    onChange={(event) => set("amount", moneyInput(event.target.value, draft.amount))}
                  />
                  {fieldMessage("amount")}
                </div>
              </div>

              <div className="grid gap-1 sm:ml-auto sm:w-1/2 sm:pl-1.5">
                <span id="cheque-home-label" className="text-[0.95rem] font-semibold">{tr("cheque_field_amount_home", "จำนวนเงิน (เงินบาท)")}</span>
                <output aria-labelledby="cheque-home-label" className={cn(readOnlyBox, "justify-end tabular-nums")}>
                  {formatDecimalText(registerHomeAmount(draft.amount)) || "0.00"}
                </output>
              </div>

              <div className="grid gap-1">
                <ChequeFieldLabel htmlFor="cheque-description">{tr("cheque_field_description", "คำอธิบาย")}</ChequeFieldLabel>
                <Input
                  id="cheque-description"
                  value={draft.description}
                  maxLength={CHEQUE_LIMITS.description}
                  aria-invalid={invalid("description")}
                  className="text-[0.95rem]"
                  onChange={(event) => set("description", event.target.value)}
                />
                {fieldMessage("description")}
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                {issued ? null : (
                  <div className="grid gap-1">
                    <ChequeFieldLabel htmlFor="cheque-sale">{tr("cheque_field_sale_code", "รหัสพนักงานขาย")}</ChequeFieldLabel>
                    <Input
                      id="cheque-sale"
                      value={draft.sale_code}
                      maxLength={CHEQUE_LIMITS.refCode}
                      aria-invalid={invalid("sale_code")}
                      className="text-[0.95rem]"
                      onChange={(event) => set("sale_code", event.target.value)}
                    />
                    {fieldMessage("sale_code")}
                  </div>
                )}
                <div className="grid gap-1">
                  <ChequeFieldLabel htmlFor="cheque-department">{tr("cheque_field_department", "รหัสแผนก")}</ChequeFieldLabel>
                  <Input
                    id="cheque-department"
                    value={draft.department_code}
                    maxLength={CHEQUE_LIMITS.refCode}
                    aria-invalid={invalid("department_code")}
                    className="text-[0.95rem]"
                    onChange={(event) => set("department_code", event.target.value)}
                  />
                  {fieldMessage("department_code")}
                </div>
              </div>

              {selected ? (
                <div className="grid gap-1">
                  <span id="cheque-stamp-label" className="text-[0.95rem] font-semibold">{tr("cheque_field_update_info", "การปรับปรุงเอกสาร")}</span>
                  <output aria-labelledby="cheque-stamp-label" className="grid gap-0.5 rounded-xl border border-border bg-muted px-3 py-2 text-[0.9rem]">
                    <span>{tr("cheque_stamp_created", "สร้างโดย")} {chequeStamp(language, selected.created_by, selected.created_at)}</span>
                    <span>{tr("cheque_stamp_updated", "แก้ไขล่าสุดโดย")} {chequeStamp(language, selected.updated_by, selected.updated_at)}</span>
                  </output>
                </div>
              ) : null}

              {formError && !FORM_FIELDS.includes(formError.field) ? (
                <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
                  {formError.message}
                </p>
              ) : null}

              <div className="flex flex-wrap items-center gap-2">
                <Button type="submit" disabled={busy || detailLoading || !canSave}>
                  {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : <Save className="size-4" aria-hidden="true" />}
                  {tr("save", "บันทึก")}
                </Button>
                {selected ? (
                  <Button type="button" variant="outline" className="text-destructive" onClick={() => void remove()} disabled={busy}>
                    <Trash2 className="size-4" aria-hidden="true" />
                    {tr("delete", "ลบ")}
                  </Button>
                ) : null}
                {dirty && !canSave ? <span className="text-[0.9rem] text-muted-foreground">{tr("cheque_save_hint_required", "กรอกช่องที่มีเครื่องหมาย * ให้ครบก่อนบันทึก")}</span> : null}
              </div>

              {selected ? (
                <div className="grid gap-1 border-t border-border pt-3">
                  <h3 className="text-[0.95rem] font-bold">{tr("cheque_references_title", "เอกสารที่อ้างถึงเช็คนี้")}</h3>
                  {detailLoading ? (
                    <p className="flex items-center gap-2 text-[0.95rem] text-muted-foreground">
                      <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                      {tr("loading", "กำลังโหลดข้อมูล...")}
                    </p>
                  ) : references.length === 0 ? (
                    <p className="text-[0.95rem] text-muted-foreground">{tr("cheque_references_empty", "ยังไม่มีเอกสารใช้เช็คใบนี้")}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full min-w-[34rem] border-separate border-spacing-y-1 text-left text-[0.95rem]">
                        <thead>
                          <tr className="text-[0.9rem] font-semibold text-muted-foreground">
                            <th className="px-1">{tr("cheque_col_item", "รายการ")}</th>
                            <th className="w-36 px-1">{tr("cheque_field_doc_no", "เลขที่เอกสาร")}</th>
                            <th className="w-32 px-1">{tr("cheque_ref_col_date", "วัน")}</th>
                            <th className="w-32 px-1 text-right">{tr("cheque_ref_col_amount", "จำนวนเงิน")}</th>
                            <th className="px-1">{tr("cheque_field_description", "คำอธิบาย")}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {references.map((reference) => {
                            const known = chequeDocTitleByKind(reference.kind_key);
                            return (
                              <tr key={`${reference.kind_key}:${reference.doc_no}`}>
                                <td className="px-1 font-semibold">{known ? tr(known[0], known[1]) : tr(reference.kind_key, reference.kind_key)}</td>
                                <td className="px-1">{reference.doc_no}</td>
                                <td className="px-1">{formatAppDate(reference.doc_date, language)}</td>
                                <td className="px-1 text-right tabular-nums">{formatDecimalText(reference.amount)}</td>
                                <td className="px-1">{reference.description}</td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              ) : null}
            </form>
          </CardContent>
        </Card>
      </div>
      {confirmationDialog}
    </section>
  );
}
