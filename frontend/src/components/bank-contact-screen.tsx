"use client";

import { Loader2, Plus, Save, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Combobox } from "@/components/ui/combobox";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { BANK_CONTACT_LIMITS } from "@/lib/bank-code-kinds";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import { chequeErrorText, listBankCodes, listBankContacts, saveBankContacts, type BankCode, type BankContact } from "@/lib/cheque-api";
import { areContactsDirty, blankContact, contactsToSave, duplicateContactCode, removedContactCodes } from "@/lib/cheque-forms";
import { getAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";

const PICKER_LIMIT = 500;

type Row = { id: number; value: BankContact };

const codeOptions = (codes: BankCode[]) => codes.map((item) => ({ value: item.code, label: `${item.code} — ${item.name}` }));

/** บันทึกรายละเอียดผู้ติดต่อธนาคาร (Champ 5680): เลือกธนาคาร + สาขา แล้วแก้ผู้ติดต่อทั้งชุดในตารางเดียว */
export function BankContactScreen({ language }: { language: LanguageCode }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });

  const [banks, setBanks] = useState<BankCode[]>([]);
  const [branches, setBranches] = useState<BankCode[]>([]);
  const [bankCode, setBankCode] = useState("");
  const [branchCode, setBranchCode] = useState("");
  const [rows, setRows] = useState<Row[]>([]);
  const [baseline, setBaseline] = useState<BankContact[]>([]);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [focusRow, setFocusRow] = useState<number | null>(null);
  const nextId = useRef(1);

  const values = useMemo(() => rows.map((row) => row.value), [rows]);
  const dirty = areContactsDirty(values, baseline);
  const ready = bankCode !== "" && branchCode !== "";

  const toRows = useCallback((contacts: BankContact[]): Row[] => contacts.map((value) => ({ id: nextId.current++, value })), []);

  useEffect(() => {
    let current = true;
    listBankCodes("bank", { activeOnly: true, limit: PICKER_LIMIT }, language)
      .then((result) => {
        if (current) setBanks(result.items ?? []);
      })
      .catch((cause: unknown) => {
        if (current) setError(chequeErrorText(cause, tr));
      });
    return () => {
      current = false;
    };
    // tr เปลี่ยนตามพจนานุกรมภาษา ไม่ควรทำให้โหลดรายการธนาคารซ้ำ
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [language]);

  useEffect(() => {
    if (!bankCode) return;
    let current = true;
    listBankCodes("bank_branch", { bankCode, activeOnly: true, limit: PICKER_LIMIT }, language)
      .then((result) => {
        if (current) setBranches(result.items ?? []);
      })
      .catch((cause: unknown) => {
        if (current) setError(chequeErrorText(cause, tr));
      });
    return () => {
      current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bankCode, language]);

  useEffect(() => {
    if (!bankCode || !branchCode) return;
    let current = true;
    setLoading(true);
    listBankContacts(bankCode, branchCode, language)
      .then((contacts) => {
        if (!current) return;
        setRows(toRows(contacts));
        setBaseline(contacts);
      })
      .catch((cause: unknown) => {
        if (current) setError(chequeErrorText(cause, tr));
      })
      .finally(() => {
        if (current) setLoading(false);
      });
    return () => {
      current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bankCode, branchCode, language, toRows]);

  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => event.preventDefault();
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

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

  const clearRows = () => {
    setRows([]);
    setBaseline([]);
    setError("");
  };

  const chooseBank = async (code: string) => {
    if (code === bankCode || !(await canLeave())) return;
    setBankCode(code);
    setBranchCode("");
    setBranches([]);
    clearRows();
  };

  const chooseBranch = async (code: string) => {
    if (code === branchCode || !(await canLeave())) return;
    setBranchCode(code);
    clearRows();
  };

  const updateRow = (id: number, patch: Partial<BankContact>) =>
    setRows((current) => current.map((row) => (row.id === id ? { ...row, value: { ...row.value, ...patch } } : row)));

  const addRow = () => {
    const row: Row = { id: nextId.current++, value: blankContact() };
    setRows((current) => [...current, row]);
    setFocusRow(row.id);
  };

  const removeRow = (id: number) => setRows((current) => current.filter((row) => row.id !== id));

  const save = async () => {
    if (busy || !ready || !dirty) return;
    setError("");
    const duplicate = duplicateContactCode(values);
    if (duplicate) {
      setError(tr("bank_contact_duplicate_code", "รหัสผู้ติดต่อ {code} ซ้ำกัน กรุณาแก้ให้ไม่ซ้ำกันก่อนบันทึก").replace("{code}", duplicate));
      return;
    }
    const removed = removedContactCodes(values, baseline);
    if (removed.length > 0) {
      const ok = await confirm({
        title: tr("bank_contact_confirm_removed_title", "ยืนยันการลบผู้ติดต่อ"),
        description: tr("bank_contact_confirm_removed", "การบันทึกนี้จะลบผู้ติดต่อ {count} ราย ({codes}) ออกจากระบบ ยืนยันหรือไม่")
          .replace("{count}", String(removed.length))
          .replace("{codes}", removed.join(", ")),
        confirmLabel: tr("save", "บันทึก"),
        tone: "danger",
      });
      if (!ok) return;
    }
    setBusy(true);
    try {
      const saved = await saveBankContacts(bankCode, branchCode, contactsToSave(values), language);
      setRows(toRows(saved));
      setBaseline(saved);
      pushNotice({ type: "success", text: tr("bank_contact_saved", "บันทึกผู้ติดต่อเรียบร้อยแล้ว") });
    } catch (cause) {
      setError(chequeErrorText(cause, tr));
    } finally {
      setBusy(false);
    }
  };

  const bankOptions = useMemo(() => codeOptions(banks), [banks]);
  const branchOptions = useMemo(() => codeOptions(branches), [branches]);
  const cell = "h-9 min-w-0 text-[0.95rem]";

  return (
    <section className="grid w-full min-w-0 gap-3">
      <header className="flex flex-wrap items-center gap-2">
        <h1 className="text-xl font-bold tracking-tight">{tr("bank_contact_title", "บันทึกรายละเอียดผู้ติดต่อธนาคาร")}</h1>
        {dirty ? <Badge variant="warning" className="text-[0.85rem]">{tr("cheque_unsaved", "ยังไม่ได้บันทึก")}</Badge> : null}
      </header>

      <Card className="min-w-0 shadow-sm">
        <CardContent className="grid gap-3 p-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="grid gap-1">
              <label htmlFor="bank-contact-bank" className="text-[0.95rem] font-semibold">
                {tr("bank_contact_bank", "ธนาคาร")}
              </label>
              <Combobox<string>
                id="bank-contact-bank"
                value={bankCode}
                options={bankOptions}
                searchable
                placeholder={tr("bank_contact_pick_bank", "เลือกธนาคาร")}
                searchPlaceholder={tr("search", "ค้นหา")}
                aria-label={tr("bank_contact_bank", "ธนาคาร")}
                onChange={(value) => void chooseBank(value)}
              />
            </div>
            <div className="grid gap-1">
              <label htmlFor="bank-contact-branch" className="text-[0.95rem] font-semibold">
                {tr("bank_contact_branch", "สาขาธนาคาร")}
              </label>
              <Combobox<string>
                id="bank-contact-branch"
                value={branchCode}
                options={branchOptions}
                searchable
                disabled={!bankCode}
                placeholder={tr("bank_contact_pick_branch", "เลือกสาขา")}
                searchPlaceholder={tr("search", "ค้นหา")}
                aria-label={tr("bank_contact_branch", "สาขาธนาคาร")}
                onChange={(value) => void chooseBranch(value)}
              />
            </div>
          </div>

          {error ? (
            <p className="rounded-xl border border-destructive/40 bg-destructive/5 p-2 text-[0.95rem] font-medium" role="alert">
              {error}
            </p>
          ) : null}

          {!ready ? (
            <p className="p-2 text-[0.95rem] text-muted-foreground">{tr("bank_contact_select_hint", "เลือกธนาคารและสาขาเพื่อดูหรือบันทึกผู้ติดต่อ")}</p>
          ) : loading ? (
            <p className="flex items-center gap-2 p-2 text-[0.95rem] text-muted-foreground">
              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              {tr("loading", "กำลังโหลดข้อมูล...")}
            </p>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full min-w-[56rem] border-separate border-spacing-y-1 text-left">
                  <thead>
                    <tr className="text-[0.9rem] font-semibold text-muted-foreground">
                      <th className="w-20 px-1">{tr("bank_contact_active", "ติดต่อ")}</th>
                      <th className="w-32 px-1">{tr("bank_contact_code", "รหัสผู้ติดต่อ")}</th>
                      <th className="w-56 px-1">{tr("bank_contact_name", "ชื่อ/รายละเอียด")}</th>
                      <th className="w-44 px-1">{tr("bank_contact_position", "ตำแหน่ง")}</th>
                      <th className="px-1">{tr("bank_contact_address", "ที่อยู่")}</th>
                      <th className="w-40 px-1">{tr("bank_contact_phone", "โทรศัพท์")}</th>
                      <th className="w-12 px-1">
                        <span className="sr-only">{tr("bank_contact_remove_row", "ลบแถว")}</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {rows.map((row) => (
                      <tr key={row.id}>
                        <td className="px-1">
                          <Checkbox
                            checked={row.value.is_contact}
                            onCheckedChange={(checked) => updateRow(row.id, { is_contact: checked })}
                            aria-label={tr("bank_contact_active", "ติดต่อ")}
                          />
                        </td>
                        <td className="px-1">
                          <Input
                            className={cell}
                            value={row.value.contact_code}
                            maxLength={BANK_CONTACT_LIMITS.code}
                            autoFocus={row.id === focusRow}
                            aria-label={tr("bank_contact_code", "รหัสผู้ติดต่อ")}
                            onChange={(event) => updateRow(row.id, { contact_code: event.target.value })}
                          />
                        </td>
                        <td className="px-1">
                          <Input
                            className={cell}
                            value={row.value.name}
                            maxLength={BANK_CONTACT_LIMITS.name}
                            aria-label={tr("bank_contact_name", "ชื่อ/รายละเอียด")}
                            onChange={(event) => updateRow(row.id, { name: event.target.value })}
                          />
                        </td>
                        <td className="px-1">
                          <Input
                            className={cell}
                            value={row.value.position}
                            maxLength={BANK_CONTACT_LIMITS.position}
                            aria-label={tr("bank_contact_position", "ตำแหน่ง")}
                            onChange={(event) => updateRow(row.id, { position: event.target.value })}
                          />
                        </td>
                        <td className="px-1">
                          <Input
                            className={cell}
                            value={row.value.address}
                            maxLength={BANK_CONTACT_LIMITS.address}
                            aria-label={tr("bank_contact_address", "ที่อยู่")}
                            onChange={(event) => updateRow(row.id, { address: event.target.value })}
                          />
                        </td>
                        <td className="px-1">
                          <Input
                            className={cell}
                            value={row.value.telephone}
                            maxLength={BANK_CONTACT_LIMITS.telephone}
                            inputMode="tel"
                            aria-label={tr("bank_contact_phone", "โทรศัพท์")}
                            onChange={(event) => updateRow(row.id, { telephone: event.target.value })}
                          />
                        </td>
                        <td className="px-1">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            onClick={() => removeRow(row.id)}
                            aria-label={tr("bank_contact_remove_row", "ลบแถว")}
                            title={tr("bank_contact_remove_row", "ลบแถว")}
                          >
                            <Trash2 className="size-4" aria-hidden="true" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {rows.length === 0 ? (
                <p className="p-2 text-[0.95rem] text-muted-foreground">{tr("bank_contact_empty", "ยังไม่มีผู้ติดต่อ กดปุ่ม \"เพิ่มแถว\" เพื่อเริ่มบันทึก")}</p>
              ) : null}
              <div className="flex flex-wrap items-center gap-2">
                <Button type="button" onClick={() => void save()} disabled={busy || !dirty}>
                  {busy ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : <Save className="size-4" aria-hidden="true" />}
                  {tr("save", "บันทึก")}
                </Button>
                <Button type="button" variant="outline" onClick={addRow} disabled={busy || rows.length >= BANK_CONTACT_LIMITS.rows}>
                  <Plus className="size-4" aria-hidden="true" />
                  {tr("bank_contact_add_row", "เพิ่มแถว")}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
      {confirmationDialog}
    </section>
  );
}
