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
import { BANK_CODE_LIMITS, type BankCodeKind } from "@/lib/bank-code-kinds";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import { chequeErrorText, ChequeApiError, createBankCode, deleteBankCode, listBankCodes, updateBankCode, type BankCode } from "@/lib/cheque-api";
import { canSaveCodeDraft, codeInputFromDraft, draftFromCode, emptyCodeDraft, isCodeDraftDirty, type CodeDraft } from "@/lib/cheque-forms";
import { getAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 50;
const BANK_PICKER_LIMIT = 500;

type FormError = { field: string; message: string };

const KIND_TITLES: Record<BankCodeKind, [key: string, fallback: string]> = {
  bank: ["bank_code_kind_bank", "กำหนดรหัสธนาคาร"],
  bank_branch: ["bank_code_kind_bank_branch", "กำหนดรหัสสาขาของธนาคาร"],
  bank_service: ["bank_code_kind_bank_service", "กำหนดรหัสบริการธนาคาร"],
  bank_account_type: ["bank_code_kind_bank_account_type", "กำหนดรหัสประเภทบัญชีธนาคาร"],
  cheque_ext_status: ["bank_code_kind_cheque_ext_status", "กำหนดรหัสสถานะเช็คเพิ่มเติม"],
};

/** จอกำหนดรหัส 5 ชนิด (ธนาคาร / สาขา / บริการ / ประเภทบัญชี / สถานะเช็คเพิ่มเติม) — จอเดียวตั้งค่าด้วย kind */
export function BankCodeScreen({ kind, language }: { kind: BankCodeKind; language: LanguageCode }) {
  return <BankCodeWorkbench key={kind} kind={kind} language={language} />;
}

function BankCodeWorkbench({ kind, language }: { kind: BankCodeKind; language: LanguageCode }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });

  const [search, setSearch] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [activeOnly, setActiveOnly] = useState(false);
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<BankCode[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [reloadTick, setReloadTick] = useState(0);
  const [banks, setBanks] = useState<BankCode[]>([]);
  const [selected, setSelected] = useState<BankCode | null>(null);
  const [draft, setDraft] = useState<CodeDraft>(emptyCodeDraft);
  const [baseline, setBaseline] = useState<CodeDraft>(emptyCodeDraft);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<FormError | null>(null);

  const isBranch = kind === "bank_branch";
  const [titleKey, titleFallback] = KIND_TITLES[kind];
  const dirty = isCodeDraftDirty(draft, baseline);
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const reload = useCallback(() => setReloadTick((tick) => tick + 1), []);

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
    listBankCodes(kind, { q: appliedSearch, activeOnly, page, limit: PAGE_SIZE }, language)
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
  }, [kind, appliedSearch, activeOnly, page, language, reloadTick]);

  // รายการธนาคารสำหรับเลือกธนาคารของสาขา
  useEffect(() => {
    if (!isBranch) return;
    let current = true;
    listBankCodes("bank", { limit: BANK_PICKER_LIMIT }, language)
      .then((result) => {
        if (current) setBanks(result.items ?? []);
      })
      .catch(() => {
        if (current) setBanks([]);
      });
    return () => {
      current = false;
    };
  }, [isBranch, language]);

  // กันปิดแท็บ/รีเฟรชเบราว์เซอร์ทิ้งข้อมูลที่ยังไม่บันทึก
  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const bankOptions = useMemo(() => {
    const options = [{ value: "", label: tr("bank_code_branch_any_bank", "ไม่ระบุธนาคาร (ใช้ได้กับทุกธนาคาร)") }];
    for (const bank of banks) options.push({ value: bank.code, label: `${bank.code} — ${bank.name}` });
    if (draft.bank_code && !banks.some((bank) => bank.code === draft.bank_code)) options.push({ value: draft.bank_code, label: draft.bank_code });
    return options;
  }, [banks, draft.bank_code, tr]);

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

  const showForm = useCallback((code: BankCode | null) => {
    const next = code ? draftFromCode(code) : emptyCodeDraft();
    setSelected(code);
    setDraft(next);
    setBaseline(next);
    setFormError(null);
  }, []);

  const startNew = async () => {
    if (await canLeave()) showForm(null);
  };

  const openCode = async (code: BankCode) => {
    if (selected?.code === code.code) return;
    if (await canLeave()) showForm(code);
  };

  const fieldError = (cause: unknown): FormError => ({ field: cause instanceof ChequeApiError ? cause.field : "", message: chequeErrorText(cause, tr) });

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (busy || !canSaveCodeDraft(draft, baseline)) return;
    setBusy(true);
    setFormError(null);
    try {
      const input = codeInputFromDraft(kind, draft);
      const saved = selected ? await updateBankCode({ ...input, version: selected.version }, language) : await createBankCode(input, language);
      showForm(saved);
      pushNotice({ type: "success", text: tr("bank_code_saved", "บันทึกรหัสเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
      // รหัสถูกคนอื่นแก้ไปก่อน: โหลดรายการใหม่ให้เห็นค่าล่าสุด
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
        "bank_code_confirm_delete",
        "ลบรหัส {code} ({name}) ออกจากทะเบียน รหัสที่ถูกใช้ในเอกสารหรือข้อมูลอื่นแล้วจะลบไม่ได้ ให้ปิดการใช้งานแทน",
      )
        .replace("{code}", selected.code)
        .replace("{name}", selected.name),
      confirmLabel: tr("delete", "ลบ"),
      tone: "danger",
    });
    if (!ok) return;
    setBusy(true);
    setFormError(null);
    try {
      await deleteBankCode({ kind, code: selected.code, version: selected.version }, language);
      showForm(null);
      pushNotice({ type: "success", text: tr("bank_code_deleted", "ลบรหัสเรียบร้อยแล้ว") });
      reload();
    } catch (cause) {
      setFormError(fieldError(cause));
    } finally {
      setBusy(false);
    }
  };

  const idPrefix = `bank-code-${kind}`;
  const invalid = (field: string) => formError?.field === field;
  const fieldMessage = (field: string) =>
    invalid(field) ? (
      <span className="text-[0.9rem] font-medium text-destructive" role="alert">
        {formError?.message}
      </span>
    ) : null;

  return (
    <section className="grid w-full min-w-0 gap-3">
      <header className="flex flex-wrap items-center gap-2">
        <h1 className="text-xl font-bold tracking-tight">{tr(titleKey, titleFallback)}</h1>
        <Badge variant="outline" className="text-[0.9rem]">
          {tr("bank_code_list_count", "ทั้งหมด {count} รายการ").replace("{count}", String(total))}
        </Badge>
      </header>

      <div className="grid min-w-0 gap-3 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)]">
        <Card className="min-w-0 shadow-sm">
          <CardContent className="grid gap-3 p-3">
            <div className="flex flex-wrap items-center gap-2">
              <label className="relative block min-w-48 flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
                <Input
                  className="!pl-10 text-[0.95rem]"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder={tr("bank_code_search_placeholder", "ค้นหารหัสหรือชื่อ")}
                  aria-label={tr("search", "ค้นหา")}
                />
              </label>
              <label className="flex w-auto cursor-pointer items-center gap-2 text-[0.95rem]">
                <Checkbox checked={activeOnly} onCheckedChange={(checked) => {
                    setActiveOnly(checked);
                    setPage(1);
                  }} />
                {tr("bank_code_active_only", "เฉพาะที่ใช้งานอยู่")}
              </label>
              <Button type="button" variant="outline" onClick={reload} disabled={loading}>
                <RefreshCcw className={cn("size-4", loading && "animate-spin")} aria-hidden="true" />
                {tr("refresh", "รีเฟรช")}
              </Button>
            </div>

            <div className="bc-list-header flex flex-nowrap items-center gap-1.5 border-b border-border px-2 pb-1 text-[0.9rem] font-semibold text-muted-foreground">
              <span className="w-28 shrink-0">{tr("code", "รหัส")}</span>
              <span className="min-w-0 flex-1">{tr("name", "ชื่อ")}</span>
              {isBranch ? <span className="hidden w-28 shrink-0 sm:inline">{tr("bank_code", "รหัสธนาคาร")}</span> : null}
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
                {appliedSearch || activeOnly ? tr("no_data", "ไม่มีข้อมูล") : tr("bank_code_list_empty", "ยังไม่มีรหัส กดปุ่ม \"เพิ่ม\" เพื่อเพิ่มรหัสแรก")}
              </p>
            ) : (
              <ul className="grid gap-0.5" aria-busy={loading}>
                {items.map((item) => (
                  <li key={item.code}>
                    <button
                      type="button"
                      onClick={() => void openCode(item)}
                      aria-current={selected?.code === item.code ? "true" : undefined}
                      className={cn(
                        "bc-list-row flex min-h-9 w-full flex-nowrap items-center gap-1.5 rounded-lg px-2 py-1 text-left text-[0.95rem] transition-colors hover:bg-accent",
                        selected?.code === item.code && "bg-primary/10 ring-1 ring-primary/40",
                      )}
                    >
                      <span className="w-28 shrink-0 font-semibold" title={item.code}>
                        <span className="bc-cell-text">{item.code}</span>
                      </span>
                      <span className="min-w-0 flex-1" title={item.name}>
                        <span className="bc-cell-text">{item.name}</span>
                      </span>
                      {isBranch ? (
                        <span className="hidden w-28 shrink-0 sm:inline" title={item.bank_code}>
                          <span className="bc-cell-text">{item.bank_code || "—"}</span>
                        </span>
                      ) : null}
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
              <nav className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-2" aria-label={tr("bank_code_page_info", "หน้า {page} จาก {pages}").replace("{page}", String(page)).replace("{pages}", String(pages))}>
                <Button type="button" variant="outline" disabled={page <= 1 || loading} onClick={() => setPage(page - 1)}>
                  {tr("bank_code_prev", "ก่อนหน้า")}
                </Button>
                <span className="text-[0.95rem]">{tr("bank_code_page_info", "หน้า {page} จาก {pages}").replace("{page}", String(page)).replace("{pages}", String(pages))}</span>
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
                <h2 className="text-base font-bold">{selected ? tr("bank_code_form_edit", "แก้ไขรหัส") : tr("bank_code_form_new", "รหัสใหม่")}</h2>
                <Button type="button" variant="outline" onClick={() => void startNew()} disabled={busy}>
                  <Plus className="size-4" aria-hidden="true" />
                  {tr("add", "เพิ่ม")}
                </Button>
              </div>

              <div className="grid gap-1">
                <label htmlFor={`${idPrefix}-code`} className="text-[0.95rem] font-semibold">
                  {tr("code", "รหัส")}
                </label>
                <Input
                  id={`${idPrefix}-code`}
                  value={draft.code}
                  maxLength={BANK_CODE_LIMITS.code}
                  readOnly={selected !== null}
                  aria-invalid={invalid("code")}
                  className={cn("text-[0.95rem]", selected && "bg-muted")}
                  onChange={(event) => setDraft({ ...draft, code: event.target.value })}
                />
                <span className="text-[0.85rem] text-muted-foreground">
                  {tr("bank_code_code_hint", "ไม่เกิน {max} ตัวอักษร และแก้ไขไม่ได้หลังบันทึก").replace("{max}", String(BANK_CODE_LIMITS.code))}
                </span>
                {fieldMessage("code")}
              </div>

              <div className="grid gap-1">
                <label htmlFor={`${idPrefix}-name`} className="text-[0.95rem] font-semibold">
                  {tr("name", "ชื่อ")}
                </label>
                <Input
                  id={`${idPrefix}-name`}
                  value={draft.name}
                  maxLength={BANK_CODE_LIMITS.name}
                  aria-invalid={invalid("name")}
                  className="text-[0.95rem]"
                  onChange={(event) => setDraft({ ...draft, name: event.target.value })}
                />
                {fieldMessage("name")}
              </div>

              {isBranch ? (
                <div className="grid gap-1">
                  <label htmlFor={`${idPrefix}-bank`} className="text-[0.95rem] font-semibold">
                    {tr("bank_code", "รหัสธนาคาร")}
                  </label>
                  <Combobox<string>
                    id={`${idPrefix}-bank`}
                    value={draft.bank_code}
                    options={bankOptions}
                    searchable
                    searchPlaceholder={tr("search", "ค้นหา")}
                    aria-label={tr("bank_code", "รหัสธนาคาร")}
                    onChange={(value) => setDraft({ ...draft, bank_code: value })}
                  />
                  {fieldMessage("bank_code")}
                </div>
              ) : null}

              <div className="grid gap-1">
                <label htmlFor={`${idPrefix}-group`} className="text-[0.95rem] font-semibold">
                  {tr("bank_code_user_group", "กลุ่มผู้ใช้")}
                </label>
                <Input
                  id={`${idPrefix}-group`}
                  value={draft.user_group}
                  maxLength={BANK_CODE_LIMITS.userGroup}
                  aria-invalid={invalid("user_group")}
                  className="text-[0.95rem]"
                  onChange={(event) => setDraft({ ...draft, user_group: event.target.value })}
                />
                {fieldMessage("user_group")}
              </div>

              <label className="flex w-auto cursor-pointer items-center gap-2 text-[0.95rem] font-semibold">
                <Checkbox checked={draft.is_active} onCheckedChange={(checked) => setDraft({ ...draft, is_active: checked })} />
                {tr("active", "ใช้งาน")}
              </label>

              {formError && !["code", "name", "bank_code", "user_group"].includes(formError.field) ? (
                <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
                  {formError.message}
                </p>
              ) : null}

              <div className="flex flex-wrap items-center gap-2">
                <Button type="submit" disabled={busy || !canSaveCodeDraft(draft, baseline)}>
                  {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : <Save className="size-4" aria-hidden="true" />}
                  {tr("save", "บันทึก")}
                </Button>
                {selected ? (
                  <Button type="button" variant="outline" className="text-destructive" onClick={() => void remove()} disabled={busy}>
                    <Trash2 className="size-4" aria-hidden="true" />
                    {tr("delete", "ลบ")}
                  </Button>
                ) : null}
                {dirty ? <Badge variant="warning" className="text-[0.85rem]">{tr("cheque_unsaved", "ยังไม่ได้บันทึก")}</Badge> : null}
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
      {confirmationDialog}
    </section>
  );
}
