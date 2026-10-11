"use client";

import { Loader2, Plus, RefreshCcw, Save, Search, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Combobox } from "@/components/ui/combobox";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import {
  chequeErrorText,
  ChequeApiError,
  changeBankBookNumber,
  createBankBook,
  deleteBankBook,
  listBankBooks,
  listBankCodes,
  listBankContacts,
  listBankGLAccounts,
  updateBankBook,
  type BankBook,
  type BankCode,
  type BankContact,
  type BankGLAccount,
} from "@/lib/cheque-api";
import { BANK_BOOK_LIMITS, bookInputFromDraft, canSaveBookDraft, draftFromBook, emptyBookDraft, formatDecimalText, isBookDraftDirty, type BookDraft } from "@/lib/cheque-forms";
import { getAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 50;
const PICKER_LIMIT = 500;
const FORM_FIELDS = ["bank_account_code", "account_number", "account_name", "bank_code", "bank_branch_code", "account_type_code", "gl_account_code"];

type FormError = { field: string; message: string };

/** จอ "บันทึกสมุดเงินฝากธนาคาร" (Champ 5169): รายการสมุด + ฟอร์ม; ยอดคงเหลือและผู้ติดต่อเป็นข้อมูลอ่านอย่างเดียว */
export function BookBankScreen({ language }: { language: LanguageCode }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });

  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<BankBook[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [reloadTick, setReloadTick] = useState(0);
  const [banks, setBanks] = useState<BankCode[]>([]);
  const [accountTypes, setAccountTypes] = useState<BankCode[]>([]);
  const [glAccounts, setGlAccounts] = useState<BankGLAccount[]>([]);
  const [branches, setBranches] = useState<BankCode[]>([]);
  const [contacts, setContacts] = useState<BankContact[]>([]);
  const [contactsLoading, setContactsLoading] = useState(false);
  const [lookupError, setLookupError] = useState<unknown>(null);
  const [selected, setSelected] = useState<BankBook | null>(null);
  const [draft, setDraft] = useState<BookDraft>(emptyBookDraft);
  const [baseline, setBaseline] = useState<BookDraft>(emptyBookDraft);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<FormError | null>(null);
  const [changing, setChanging] = useState(false);
  const [newNumber, setNewNumber] = useState("");
  const [numberError, setNumberError] = useState("");

  const dirty = isBookDraftDirty(draft, baseline);
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const reload = useCallback(() => setReloadTick((tick) => tick + 1), []);
  const pageInfo = tr("bank_code_page_info", "หน้า {page} จาก {pages}").replace("{page}", String(page)).replace("{pages}", String(pages));

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
    listBankBooks({ q: appliedSearch, page, limit: PAGE_SIZE }, language)
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
  }, [appliedSearch, page, language, reloadTick]);

  // รายการให้เลือก: ธนาคาร / ประเภทบัญชี (จากทะเบียนรหัส CB-1) และบัญชีผังบัญชีที่ผูกได้
  useEffect(() => {
    let current = true;
    setLookupError(null);
    Promise.all([
      listBankCodes("bank", { activeOnly: true, limit: PICKER_LIMIT }, language),
      listBankCodes("bank_account_type", { activeOnly: true, limit: PICKER_LIMIT }, language),
      listBankGLAccounts(language),
    ])
      .then(([bankPage, typePage, accounts]) => {
        if (!current) return;
        setBanks(bankPage.items ?? []);
        setAccountTypes(typePage.items ?? []);
        setGlAccounts(accounts);
      })
      .catch((cause: unknown) => {
        if (current) setLookupError(cause);
      });
    return () => {
      current = false;
    };
  }, [language, reloadTick]);

  // สาขาของธนาคารที่เลือก (สาขาที่ไม่ระบุธนาคารใช้ได้ทุกธนาคาร — backend รวมให้แล้ว)
  useEffect(() => {
    if (!draft.bank_code) {
      setBranches([]);
      return;
    }
    let current = true;
    listBankCodes("bank_branch", { bankCode: draft.bank_code, activeOnly: true, limit: PICKER_LIMIT }, language)
      .then((result) => {
        if (current) setBranches(result.items ?? []);
      })
      .catch(() => {
        if (current) setBranches([]);
      });
    return () => {
      current = false;
    };
  }, [draft.bank_code, language]);

  // ตารางผู้ติดต่อ (อ่านอย่างเดียว) ของธนาคาร + สาขาที่เลือก — แก้ไขที่จอบันทึกรายละเอียดผู้ติดต่อธนาคาร
  useEffect(() => {
    if (!draft.bank_code || !draft.bank_branch_code) {
      setContacts([]);
      setContactsLoading(false);
      return;
    }
    let current = true;
    setContactsLoading(true);
    listBankContacts(draft.bank_code, draft.bank_branch_code, language)
      .then((rows) => {
        if (current) setContacts(rows);
      })
      .catch(() => {
        if (current) setContacts([]);
      })
      .finally(() => {
        if (current) setContactsLoading(false);
      });
    return () => {
      current = false;
    };
  }, [draft.bank_code, draft.bank_branch_code, language]);

  // กันปิดแท็บ/รีเฟรชเบราว์เซอร์ทิ้งข้อมูลที่ยังไม่บันทึก
  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  // ตัวเลือกต้องมีค่าที่เลือกอยู่เสมอ แม้รหัสนั้นถูกปิดการใช้งานภายหลัง (สมุดเดิมต้องยังแสดงค่าเดิมได้)
  const withCurrent = (options: { value: string; label: string }[], current: string) =>
    current && !options.some((option) => option.value === current) ? [...options, { value: current, label: current }] : options;

  const bankOptions = useMemo(
    () => withCurrent(banks.map((bank) => ({ value: bank.code, label: `${bank.code} — ${bank.name}` })), draft.bank_code),
    [banks, draft.bank_code],
  );
  const branchOptions = useMemo(
    () =>
      withCurrent(
        [{ value: "", label: tr("book_bank_no_branch", "ไม่ระบุสาขา") }, ...branches.map((branch) => ({ value: branch.code, label: `${branch.code} — ${branch.name}` }))],
        draft.bank_branch_code,
      ),
    [branches, draft.bank_branch_code, tr],
  );
  const typeOptions = useMemo(
    () =>
      withCurrent(
        [{ value: "", label: tr("book_bank_no_account_type", "ไม่ระบุประเภทบัญชี") }, ...accountTypes.map((type) => ({ value: type.code, label: `${type.code} — ${type.name}` }))],
        draft.account_type_code,
      ),
    [accountTypes, draft.account_type_code, tr],
  );
  const glOptions = useMemo(
    () => withCurrent(glAccounts.map((account) => ({ value: account.code, label: account.name ? `${account.code} — ${account.name}` : account.code })), draft.gl_account_code),
    [glAccounts, draft.gl_account_code],
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

  const showForm = useCallback((book: BankBook | null) => {
    const next = book ? draftFromBook(book) : emptyBookDraft();
    setSelected(book);
    setDraft(next);
    setBaseline(next);
    setFormError(null);
    setChanging(false);
    setNewNumber("");
    setNumberError("");
  }, []);

  const startNew = async () => {
    if (await canLeave()) showForm(null);
  };

  const openBook = async (book: BankBook) => {
    if (selected?.bank_account_code === book.bank_account_code) return;
    if (await canLeave()) showForm(book);
  };

  const fieldError = (cause: unknown): FormError => ({ field: cause instanceof ChequeApiError ? cause.field : "", message: chequeErrorText(cause, tr) });

  // อ่านสมุดที่เพิ่งบันทึกกลับมา (ได้ชื่อธนาคารและยอดคงเหลือล่าสุด); อ่านไม่ได้ก็ไม่ทำให้การบันทึกที่สำเร็จแล้วดูเหมือนล้มเหลว
  const fetchSaved = async (code: string, version: number): Promise<BankBook> => {
    const fallback: BankBook = { ...bookInputFromDraft(draft), bank_name: "", version, gl_balance: selected?.gl_balance };
    try {
      const result = await listBankBooks({ q: code, limit: PAGE_SIZE }, language);
      return result.items?.find((item) => item.bank_account_code.toUpperCase() === code.toUpperCase()) ?? fallback;
    } catch {
      return fallback;
    }
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (busy || !canSaveBookDraft(draft, baseline)) return;
    setBusy(true);
    setFormError(null);
    try {
      const input = bookInputFromDraft(draft);
      const saved = selected ? await updateBankBook(input, selected.version, language) : await createBankBook(input, language);
      showForm(await fetchSaved(saved.id, saved.version));
      pushNotice({ type: "success", text: tr("book_bank_saved", "บันทึกสมุดเงินฝากเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
      // สมุดถูกคนอื่นแก้ไปก่อน: โหลดรายการใหม่ให้เห็นค่าล่าสุด
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
        "book_bank_confirm_delete",
        "ลบสมุดเงินฝาก {code} ({name}) สมุดที่มีรายการเดินบัญชีหรือมียอดคงเหลือจะลบไม่ได้ ให้ปิดการใช้งานแทน",
      )
        .replace("{code}", selected.bank_account_code)
        .replace("{name}", selected.account_name),
      confirmLabel: tr("delete", "ลบ"),
      tone: "danger",
    });
    if (!ok) return;
    setBusy(true);
    setFormError(null);
    try {
      await deleteBankBook(selected.bank_account_code, selected.version, language);
      showForm(null);
      pushNotice({ type: "success", text: tr("book_bank_deleted", "ลบสมุดเงินฝากเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
    } finally {
      setBusy(false);
    }
  };

  // ปุ่ม "เปลี่ยนเลขที่บัญชี": เปลี่ยนเฉพาะเลขที่บัญชีธนาคาร รหัสสมุดและเอกสารที่อ้างสมุดไม่เปลี่ยน
  const applyNumberChange = async () => {
    const number = newNumber.trim();
    if (!selected || busy || !number) return;
    const ok = await confirm({
      title: tr("book_bank_change_account_no", "เปลี่ยนเลขที่บัญชี"),
      description: tr("book_bank_confirm_change_no", "ยืนยันเปลี่ยนเลขที่บัญชีจาก {old} เป็น {new} หรือไม่ เอกสารเดิมทั้งหมดจะอ้างเลขใหม่")
        .replace("{old}", selected.account_number)
        .replace("{new}", number),
      confirmLabel: tr("confirm", "ยืนยัน"),
      tone: "warning",
    });
    if (!ok) return;
    setBusy(true);
    setNumberError("");
    try {
      const result = await changeBankBookNumber(selected.bank_account_code, selected.version, number, language);
      // เปลี่ยนเฉพาะเลขที่บัญชีกับ version — ค่าอื่นที่ผู้ใช้กำลังแก้อยู่ยังคงอยู่
      setSelected({ ...selected, account_number: number, version: result.version });
      setDraft((current) => ({ ...current, account_number: number }));
      setBaseline((current) => ({ ...current, account_number: number }));
      setChanging(false);
      setNewNumber("");
      pushNotice({ type: "success", text: tr("book_bank_number_changed", "เปลี่ยนเลขที่บัญชีเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setNumberError(chequeErrorText(cause, tr));
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
  const labelClass = "text-[0.95rem] font-semibold";
  const ready = Boolean(draft.bank_code && draft.bank_branch_code);

  return (
    <section className="grid w-full min-w-0 gap-3">
      <header className="flex flex-wrap items-center gap-2">
        <h1 className="text-xl font-bold tracking-tight">{tr("book_bank_title", "บันทึกสมุดเงินฝากธนาคาร")}</h1>
        <Badge variant="outline" className="text-[0.9rem]">
          {tr("bank_code_list_count", "ทั้งหมด {count} รายการ").replace("{count}", String(total))}
        </Badge>
        {dirty ? <Badge variant="warning" className="text-[0.85rem]">{tr("cheque_unsaved", "ยังไม่ได้บันทึก")}</Badge> : null}
      </header>

      <div className="grid min-w-0 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.3fr)]">
        <Card className="h-fit min-w-0 shadow-sm">
          <CardContent className="grid gap-3 p-3">
            <div className="flex flex-wrap items-center gap-2">
              <label className="relative block min-w-48 flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
                <Input
                  className="!pl-10 text-[0.95rem]"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder={tr("book_bank_search_placeholder", "ค้นหารหัสสมุด เลขที่บัญชี หรือชื่อบัญชี")}
                  aria-label={tr("search", "ค้นหา")}
                />
              </label>
              <Button type="button" variant="outline" onClick={reload} disabled={loading}>
                <RefreshCcw className={cn("size-4", loading && "animate-spin")} aria-hidden="true" />
                {tr("refresh", "รีเฟรช")}
              </Button>
            </div>

            <div className="bc-list-header flex flex-nowrap items-center gap-1.5 border-b border-border px-2 pb-1 text-[0.9rem] font-semibold text-muted-foreground">
              <span className="w-24 shrink-0">{tr("code", "รหัส")}</span>
              <span className="w-32 shrink-0">{tr("book_bank_account_no", "เลขที่บัญชีธนาคาร")}</span>
              <span className="min-w-0 flex-1">{tr("book_bank_account_name", "ชื่อบัญชี")}</span>
              <span className="w-24 shrink-0 text-right">{tr("status", "สถานะ")}</span>
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
                {appliedSearch ? tr("no_data", "ไม่มีข้อมูล") : tr("book_bank_list_empty", "ยังไม่มีสมุดเงินฝาก กดปุ่ม \"เพิ่ม\" เพื่อเพิ่มสมุดแรก")}
              </p>
            ) : (
              <ul className="grid gap-0.5" aria-busy={loading}>
                {items.map((item) => (
                  <li key={item.bank_account_code}>
                    <button
                      type="button"
                      onClick={() => void openBook(item)}
                      aria-current={selected?.bank_account_code === item.bank_account_code ? "true" : undefined}
                      className={cn(
                        "bc-list-row flex min-h-9 w-full flex-nowrap items-center gap-1.5 rounded-lg px-2 py-1 text-left text-[0.95rem] transition-colors hover:bg-accent",
                        selected?.bank_account_code === item.bank_account_code && "bg-primary/10 ring-1 ring-primary/40",
                      )}
                    >
                      <span className="w-24 shrink-0 font-semibold" title={item.bank_account_code}>
                        <span className="bc-cell-text">{item.bank_account_code}</span>
                      </span>
                      <span className="w-32 shrink-0" title={item.account_number}>
                        <span className="bc-cell-text">{item.account_number}</span>
                      </span>
                      <span className="min-w-0 flex-1" title={item.account_name}>
                        <span className="bc-cell-text">{item.account_name}</span>
                      </span>
                      <span className="flex w-24 shrink-0 justify-end">
                        <Badge variant={item.is_active ? "success" : "warning"} className="text-[0.85rem]">
                          {item.is_active ? tr("active", "ใช้งาน") : tr("inactive", "ปิดใช้งาน")}
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
                <h2 className="text-base font-bold">{selected ? tr("book_bank_form_edit", "แก้ไขสมุดเงินฝาก") : tr("book_bank_form_new", "สมุดเงินฝากใหม่")}</h2>
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

              <div className="grid gap-1">
                <label htmlFor="book-bank-code" className={labelClass}>
                  {tr("book_bank_code", "รหัสสมุดบัญชี")}
                </label>
                <Input
                  id="book-bank-code"
                  value={draft.bank_account_code}
                  maxLength={BANK_BOOK_LIMITS.code}
                  readOnly={selected !== null}
                  aria-invalid={invalid("bank_account_code")}
                  className={cn("text-[0.95rem]", selected && "bg-muted")}
                  onChange={(event) => setDraft({ ...draft, bank_account_code: event.target.value })}
                />
                <span className="text-[0.85rem] text-muted-foreground">
                  {tr("book_bank_code_hint", "ไม่เกิน {max} ตัวอักษร และแก้ไขไม่ได้หลังบันทึก").replace("{max}", String(BANK_BOOK_LIMITS.code))}
                </span>
                {fieldMessage("bank_account_code")}
              </div>

              <div className="grid gap-1">
                <label htmlFor="book-bank-number" className={labelClass}>
                  {tr("book_bank_account_no", "เลขที่บัญชีธนาคาร")}
                </label>
                <div className="flex flex-wrap items-center gap-2">
                  <Input
                    id="book-bank-number"
                    className={cn("min-w-48 flex-1 text-[0.95rem]", selected && "bg-muted")}
                    value={draft.account_number}
                    maxLength={BANK_BOOK_LIMITS.accountNumber}
                    readOnly={selected !== null}
                    aria-invalid={invalid("account_number")}
                    onChange={(event) => setDraft({ ...draft, account_number: event.target.value })}
                  />
                  {selected && !changing ? (
                    <Button
                      type="button"
                      variant="outline"
                      disabled={busy}
                      onClick={() => {
                        setChanging(true);
                        setNewNumber("");
                        setNumberError("");
                      }}
                    >
                      {tr("book_bank_change_account_no", "เปลี่ยนเลขที่บัญชี")}
                    </Button>
                  ) : null}
                </div>
                {fieldMessage("account_number")}
                {selected && changing ? (
                  <div className="grid gap-2 rounded-xl border border-border bg-muted/40 p-2">
                    <label htmlFor="book-bank-new-number" className={labelClass}>
                      {tr("book_bank_new_account_no", "เลขที่บัญชีใหม่")}
                    </label>
                    <Input
                      id="book-bank-new-number"
                      className="text-[0.95rem]"
                      value={newNumber}
                      maxLength={BANK_BOOK_LIMITS.accountNumber}
                      autoFocus
                      aria-invalid={numberError !== ""}
                      onChange={(event) => setNewNumber(event.target.value)}
                    />
                    {numberError ? (
                      <span className="text-[0.9rem] font-medium text-destructive" role="alert">
                        {numberError}
                      </span>
                    ) : null}
                    <div className="flex flex-wrap items-center gap-2">
                      <Button type="button" disabled={busy || newNumber.trim() === ""} onClick={() => void applyNumberChange()}>
                        {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : null}
                        {tr("confirm", "ยืนยัน")}
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        disabled={busy}
                        onClick={() => {
                          setChanging(false);
                          setNumberError("");
                        }}
                      >
                        {tr("cancel", "ยกเลิก")}
                      </Button>
                    </div>
                  </div>
                ) : null}
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <label htmlFor="book-bank-type" className={labelClass}>
                    {tr("book_bank_account_type", "ประเภทบัญชี")}
                  </label>
                  <Combobox<string>
                    id="book-bank-type"
                    value={draft.account_type_code}
                    options={typeOptions}
                    searchable
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("book_bank_account_type", "ประเภทบัญชี")}
                    onChange={(value) => setDraft({ ...draft, account_type_code: value })}
                  />
                  {fieldMessage("account_type_code")}
                </div>
                <div className="grid gap-1">
                  <label htmlFor="book-bank-name" className={labelClass}>
                    {tr("book_bank_account_name", "ชื่อบัญชี")}
                  </label>
                  <Input
                    id="book-bank-name"
                    value={draft.account_name}
                    maxLength={BANK_BOOK_LIMITS.accountName}
                    aria-invalid={invalid("account_name")}
                    className="text-[0.95rem]"
                    onChange={(event) => setDraft({ ...draft, account_name: event.target.value })}
                  />
                  {fieldMessage("account_name")}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <label htmlFor="book-bank-bank" className={labelClass}>
                    {tr("bank_contact_bank", "ธนาคาร")}
                  </label>
                  <Combobox<string>
                    id="book-bank-bank"
                    value={draft.bank_code}
                    options={bankOptions}
                    searchable
                    placeholder={tr("bank_contact_pick_bank", "เลือกธนาคาร")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("bank_contact_bank", "ธนาคาร")}
                    // เปลี่ยนธนาคารแล้วสาขาเดิมอาจไม่ใช่ของธนาคารใหม่ จึงล้างให้เลือกสาขาใหม่
                    onChange={(value) => setDraft(value === draft.bank_code ? draft : { ...draft, bank_code: value, bank_branch_code: "" })}
                  />
                  {fieldMessage("bank_code")}
                </div>
                <div className="grid gap-1">
                  <label htmlFor="book-bank-branch" className={labelClass}>
                    {tr("bank_contact_branch", "สาขาธนาคาร")}
                  </label>
                  <Combobox<string>
                    id="book-bank-branch"
                    value={draft.bank_branch_code}
                    options={branchOptions}
                    searchable
                    disabled={!draft.bank_code}
                    placeholder={tr("bank_contact_pick_branch", "เลือกสาขา")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("bank_contact_branch", "สาขาธนาคาร")}
                    onChange={(value) => setDraft({ ...draft, bank_branch_code: value })}
                  />
                  {fieldMessage("bank_branch_code")}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="grid gap-1">
                  <label htmlFor="book-bank-gl" className={labelClass}>
                    {tr("book_bank_gl_account", "รหัสผังบัญชี")}
                  </label>
                  <Combobox<string>
                    id="book-bank-gl"
                    value={draft.gl_account_code}
                    options={glOptions}
                    searchable
                    placeholder={tr("book_bank_pick_gl", "เลือกบัญชีในผังบัญชี")}
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("book_bank_gl_account", "รหัสผังบัญชี")}
                    onChange={(value) => setDraft({ ...draft, gl_account_code: value })}
                  />
                  {fieldMessage("gl_account_code")}
                </div>
                {selected ? (
                  <div className="grid gap-1">
                    <span id="book-bank-balance-label" className={labelClass}>
                      {tr("book_bank_balance", "ยอดคงเหลือ (บาท)")}
                    </span>
                    <output
                      aria-labelledby="book-bank-balance-label"
                      className="flex min-h-10 items-center justify-end rounded-xl border border-border bg-muted px-3 text-[0.95rem] font-semibold tabular-nums"
                    >
                      {formatDecimalText(selected.gl_balance) || "0.00"}
                    </output>
                    <span className="text-[0.85rem] text-muted-foreground">
                      {tr("book_bank_balance_hint", "ยอดจากบัญชีแยกประเภทของรหัสผังบัญชีที่ผูกไว้ แก้ไขที่นี่ไม่ได้")}
                    </span>
                  </div>
                ) : null}
              </div>

              <label className="flex w-auto cursor-pointer items-center gap-2 text-[0.95rem] font-semibold">
                <Checkbox checked={draft.is_active} onCheckedChange={(checked) => setDraft({ ...draft, is_active: checked })} />
                {tr("active", "ใช้งาน")}
              </label>

              {formError && !FORM_FIELDS.includes(formError.field) ? (
                <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
                  {formError.message}
                </p>
              ) : null}

              <div className="flex flex-wrap items-center gap-2">
                <Button type="submit" disabled={busy || !canSaveBookDraft(draft, baseline)}>
                  {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : <Save className="size-4" aria-hidden="true" />}
                  {tr("save", "บันทึก")}
                </Button>
                {selected ? (
                  <Button type="button" variant="outline" className="text-destructive" onClick={() => void remove()} disabled={busy}>
                    <Trash2 className="size-4" aria-hidden="true" />
                    {tr("delete", "ลบ")}
                  </Button>
                ) : null}
              </div>

              <div className="grid gap-1 border-t border-border pt-3">
                <h3 className="text-[0.95rem] font-bold">{tr("book_bank_contacts", "ชื่อผู้ติดต่อ")}</h3>
                {!ready ? (
                  <p className="text-[0.95rem] text-muted-foreground">
                    {tr("book_bank_contacts_hint", "เลือกธนาคารและสาขาเพื่อดูผู้ติดต่อ แก้ไขได้ที่หน้าบันทึกรายละเอียดผู้ติดต่อธนาคาร")}
                  </p>
                ) : contactsLoading ? (
                  <p className="flex items-center gap-2 text-[0.95rem] text-muted-foreground">
                    <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                    {tr("loading", "กำลังโหลดข้อมูล...")}
                  </p>
                ) : contacts.length === 0 ? (
                  <p className="text-[0.95rem] text-muted-foreground">{tr("book_bank_contacts_empty", "ยังไม่มีผู้ติดต่อ")}</p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full min-w-[34rem] border-separate border-spacing-y-1 text-left text-[0.95rem]">
                      <thead>
                        <tr className="text-[0.9rem] font-semibold text-muted-foreground">
                          <th className="w-28 px-1">{tr("bank_contact_code", "รหัสผู้ติดต่อ")}</th>
                          <th className="px-1">{tr("bank_contact_name", "ชื่อ/รายละเอียด")}</th>
                          <th className="w-36 px-1">{tr("bank_contact_position", "ตำแหน่ง")}</th>
                          <th className="w-36 px-1">{tr("bank_contact_phone", "โทรศัพท์")}</th>
                          <th className="w-24 px-1 text-right">{tr("status", "สถานะ")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {contacts.map((contact) => (
                          <tr key={contact.contact_code} className={cn(!contact.is_contact && "text-muted-foreground")}>
                            <td className="px-1 font-semibold">{contact.contact_code}</td>
                            <td className="px-1">{contact.name}</td>
                            <td className="px-1">{contact.position}</td>
                            <td className="px-1">{contact.telephone}</td>
                            <td className="px-1 text-right">
                              <Badge variant={contact.is_contact ? "success" : "warning"} className="text-[0.85rem]">
                                {contact.is_contact ? tr("active", "ใช้งาน") : tr("inactive", "ปิดใช้งาน")}
                              </Badge>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
      {confirmationDialog}
    </section>
  );
}
