"use client";

import { ArrowDownLeft, ArrowUpRight, History, Loader2, Plus, Receipt, RefreshCw, Save, Trash2, Wallet, X } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import {
  chequeErrorText,
  createPettyCashTrans,
  deletePettyCashTrans,
  listPettyCash,
  listPettyCashTrans,
  savePettyCash,
  type PettyCashItem,
  type PettyCashTransItem,
} from "@/lib/cheque-api";
import { getAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";

type Row = {
  id: number;
  code: string;
  name: string;
  amount: string;
  begin_amount: string;
  balance: string;
};

const formatMoney = (val: string): string => {
  const clean = val.replace(/,/g, "").trim();
  const num = parseFloat(clean);
  if (isNaN(num)) return "0.00";
  return num.toFixed(2);
};

export function PettyCashLimitScreen({ language }: { language: LanguageCode }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });

  const [rows, setRows] = useState<Row[]>([]);
  const [baseline, setBaseline] = useState<PettyCashItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const nextId = useRef(1);

  // Modal State สำหรับบันทึกรับ/จ่ายเงินสดย่อย (Champ 5178/5179)
  const [activePettyCash, setActivePettyCash] = useState<Row | null>(null);
  const [transModalOpen, setTransModalOpen] = useState(false);
  const [transList, setTransList] = useState<PettyCashTransItem[]>([]);
  const [transLoading, setTransLoading] = useState(false);
  const [transDocNo, setTransDocNo] = useState("");
  const [transDocDate, setTransDocDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [transType, setTransType] = useState<1 | 2>(2); // 1 = receive/replenish, 2 = expense
  const [transAmount, setTransAmount] = useState("");
  const [transDesc, setTransDesc] = useState("");
  const [transPayee, setTransPayee] = useState("");
  const [transSaving, setTransSaving] = useState(false);
  const [transError, setTransError] = useState("");

  const dirty = useMemo(() => {
    if (rows.length !== baseline.length) return true;
    for (let i = 0; i < rows.length; i++) {
      const r = rows[i];
      const b = baseline[i];
      if (
        r.code.trim() !== b.code.trim() ||
        r.name.trim() !== b.name.trim() ||
        parseFloat(r.amount || "0") !== parseFloat(b.amount || "0") ||
        parseFloat(r.begin_amount || "0") !== parseFloat(b.begin_amount || "0")
      ) {
        return true;
      }
    }
    return false;
  }, [rows, baseline]);

  const loadData = useCallback(() => {
    setLoading(true);
    setError("");
    listPettyCash(language)
      .then((items) => {
        setBaseline(items);
        setRows(
          items.map((item) => ({
            id: nextId.current++,
            code: item.code,
            name: item.name,
            amount: item.amount,
            begin_amount: item.begin_amount,
            balance: item.balance,
          }))
        );
      })
      .catch((cause: unknown) => {
        setError(chequeErrorText(cause, tr));
      })
      .finally(() => {
        setLoading(false);
      });
  }, [language, tr]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  useEffect(() => {
    if (!dirty) return;
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [dirty]);

  const handleAddRow = () => {
    setRows((prev) => [
      ...prev,
      {
        id: nextId.current++,
        code: "",
        name: "",
        amount: "0.00",
        begin_amount: "0.00",
        balance: "0.00",
      },
    ]);
  };

  const handleRemoveRow = async (id: number) => {
    const target = rows.find((r) => r.id === id);
    const label = target?.name ? `${target.code} — ${target.name}` : target?.code || tr("petty_cash_this_row", "แถวนี้");
    const ok = await confirm({
      title: tr("petty_cash_delete_title", "ยืนยันการลบวงเงินสดย่อย"),
      description: tr("petty_cash_delete_desc", "คุณต้องการลบวงเงินสดย่อย {label} ใช่หรือไม่?").replace("{label}", label),
      tone: "danger",
    });
    if (!ok) return;

    setRows((prev) => prev.filter((r) => r.id !== id));
  };

  const handleChangeRow = (id: number, field: keyof Row, value: string) => {
    setRows((prev) =>
      prev.map((row) => {
        if (row.id !== id) return row;
        const updated = { ...row, [field]: value };
        if (field === "begin_amount") {
          // ใน Champ CASHFrmPettyCash: Balance เริ่มต้นจะคำนวณตาม BeginAmount
          const num = parseFloat(value.replace(/,/g, "").trim());
          updated.balance = isNaN(num) ? "0.00" : num.toFixed(2);
        }
        return updated;
      })
    );
  };

  const handleBlurAmount = (id: number, field: "amount" | "begin_amount") => {
    setRows((prev) =>
      prev.map((row) => {
        if (row.id !== id) return row;
        const formatted = formatMoney(row[field]);
        const updated = { ...row, [field]: formatted };
        if (field === "begin_amount") {
          updated.balance = formatted;
        }
        return updated;
      })
    );
  };

  const handleSave = async () => {
    // ตรวจสอบความถูกต้องเบื้องต้น
    for (let i = 0; i < rows.length; i++) {
      const r = rows[i];
      if (!r.code.trim()) {
        setError(tr("petty_cash_code_required", "กรุณาระบุรหัสวงเงินสดย่อยแถวที่ {line}").replace("{line}", String(i + 1)));
        return;
      }
      if (!r.name.trim()) {
        setError(tr("petty_cash_name_required", "กรุณาระบุชื่อวงเงินสดย่อยแถวที่ {line}").replace("{line}", String(i + 1)));
        return;
      }
    }

    const seen = new Set<string>();
    for (let i = 0; i < rows.length; i++) {
      const code = rows[i].code.trim().toUpperCase();
      if (seen.has(code)) {
        setError(tr("petty_cash_duplicate_code", "รหัสวงเงินสดย่อย {code} ซ้ำกัน").replace("{code}", rows[i].code.trim()));
        return;
      }
      seen.add(code);
    }

    const ok = await confirm({
      title: tr("petty_cash_save_confirm_title", "ยืนยันการบันทึกวงเงินสดย่อย"),
      description: tr("petty_cash_save_confirm_desc", "ต้องการบันทึกการเปลี่ยนแปลงทั้งหมด {count} รายการ ใช่หรือไม่?").replace("{count}", String(rows.length)),
    });
    if (!ok) return;

    setBusy(true);
    setError("");
    try {
      const payload: PettyCashItem[] = rows.map((r) => ({
        code: r.code.trim(),
        name: r.name.trim(),
        amount: formatMoney(r.amount),
        begin_amount: formatMoney(r.begin_amount),
        balance: formatMoney(r.balance),
      }));
      const saved = await savePettyCash(payload, language);
      setBaseline(saved);
      setRows(
        saved.map((item) => ({
          id: nextId.current++,
          code: item.code,
          name: item.name,
          amount: item.amount,
          begin_amount: item.begin_amount,
          balance: item.balance,
        }))
      );
      pushNotice({
        type: "success",
        text: tr("petty_cash_save_success", "บันทึกกำหนดวงเงินสดย่อยเรียบร้อยแล้ว"),
      });
    } catch (cause: unknown) {
      setError(chequeErrorText(cause, tr));
    } finally {
      setBusy(false);
    }
  };

  const loadTransHistory = useCallback(async (code: string) => {
    setTransLoading(true);
    setTransError("");
    try {
      const items = await listPettyCashTrans({ petty_cash_code: code, limit: 100 }, language);
      setTransList(items);
    } catch (cause: unknown) {
      setTransError(chequeErrorText(cause, tr));
    } finally {
      setTransLoading(false);
    }
  }, [language, tr]);

  const handleOpenTransModal = (row: Row) => {
    setActivePettyCash(row);
    setTransDocNo(`PC-${row.code}-${Date.now().toString().slice(-4)}`);
    setTransDocDate(new Date().toISOString().slice(0, 10));
    setTransType(2); // default: จ่ายเงินสดย่อย
    setTransAmount("");
    setTransDesc("");
    setTransPayee("");
    setTransError("");
    setTransModalOpen(true);
    void loadTransHistory(row.code);
  };

  const handleSaveTrans = async () => {
    if (!activePettyCash) return;
    const docNo = transDocNo.trim();
    if (!docNo) {
      setTransError(tr("petty_cash_err_doc_no_required", "กรุณาระบุเลขที่เอกสาร"));
      return;
    }
    const amtNum = parseFloat(transAmount.replace(/,/g, ""));
    if (isNaN(amtNum) || amtNum <= 0) {
      setTransError(tr("petty_cash_err_amount_positive", "จำนวนเงินต้องมากกว่าศูนย์"));
      return;
    }

    setTransSaving(true);
    setTransError("");
    try {
      await createPettyCashTrans(
        {
          doc_no: docNo,
          doc_date: transDocDate,
          petty_cash_code: activePettyCash.code,
          trans_type: transType,
          amount: amtNum.toFixed(2),
          description: transDesc.trim(),
          payee_payer: transPayee.trim(),
        },
        language,
      );
      pushNotice({
        type: "success",
        text: tr("petty_cash_trans_success", "บันทึกรายการเงินสดย่อยเรียบร้อยแล้ว"),
      });
      setTransDocNo(`PC-${activePettyCash.code}-${Date.now().toString().slice(-4)}`);
      setTransAmount("");
      setTransDesc("");
      setTransPayee("");
      void loadTransHistory(activePettyCash.code);
      loadData();
    } catch (cause: unknown) {
      setTransError(chequeErrorText(cause, tr));
    } finally {
      setTransSaving(false);
    }
  };

  const handleDeleteTrans = async (item: PettyCashTransItem) => {
    const ok = await confirm({
      title: tr("petty_cash_del_trans_title", "ยืนยันการลบรายการ"),
      description: tr(
        "petty_cash_del_trans_desc",
        "ต้องการลบรายการ {doc_no} จำนวน {amount} บาท ใช่หรือไม่?",
      )
        .replace("{doc_no}", item.doc_no)
        .replace("{amount}", item.amount),
      tone: "danger",
    });
    if (!ok) return;

    try {
      await deletePettyCashTrans(item.id, language);
      pushNotice({
        type: "success",
        text: tr("petty_cash_del_trans_success", "ลบรายการเรียบร้อยแล้ว"),
      });
      if (activePettyCash) {
        void loadTransHistory(activePettyCash.code);
      }
      loadData();
    } catch (cause: unknown) {
      pushNotice({
        type: "error",
        text: chequeErrorText(cause, tr),
      });
    }
  };

  return (
    <div className="flex h-full flex-col gap-4 p-4 md:p-6">
      {confirmationDialog}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border pb-4">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Wallet className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold tracking-tight text-foreground md:text-2xl">
                {tr("petty_cash_limit_title", "กำหนดวงเงินสดย่อย")}
              </h1>
              <Badge variant="outline" className="text-xs font-normal">
                {tr("petty_cash_code_5177", "5177")}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground">
              {tr("petty_cash_limit_subtitle", "กำหนดรหัส ชื่อ วงเงิน และยอดยกมาของวงเงินสดย่อยตามระบบ Champ (CASHFrmPettyCash)")}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={loadData}
            disabled={loading || busy}
            className="h-10 gap-1.5"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            {tr("refresh", "รีเฟรช")}
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={handleAddRow}
            disabled={loading || busy}
            className="h-10 gap-1.5"
          >
            <Plus className="h-4 w-4" />
            {tr("add_row", "เพิ่มรายการ")}
          </Button>

          <Button
            size="sm"
            onClick={handleSave}
            disabled={!dirty || busy || loading}
            className="h-10 gap-1.5 min-w-[120px]"
          >
            {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {tr("save", "บันทึกข้อมูล")}
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm font-medium text-destructive">
          {error}
        </div>
      )}

      <Card className="flex-1 overflow-hidden border border-border shadow-sm">
        <CardContent className="h-full p-0">
          <div className="relative h-full overflow-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="sticky top-0 z-10 border-b border-border bg-muted/80 backdrop-blur-sm">
                <tr>
                  <th className="w-12 px-3 py-3 text-center font-semibold text-muted-foreground">#</th>
                  <th className="w-40 px-3 py-3 font-semibold text-foreground">
                    {tr("petty_cash_col_code", "รหัสวงเงินสดย่อย")}
                  </th>
                  <th className="min-w-[220px] px-3 py-3 font-semibold text-foreground">
                    {tr("petty_cash_col_name", "ชื่อวงเงินสดย่อย")}
                  </th>
                  <th className="w-44 px-3 py-3 text-right font-semibold text-foreground">
                    {tr("petty_cash_col_amount", "วงเงิน (บาท)")}
                  </th>
                  <th className="w-44 px-3 py-3 text-right font-semibold text-foreground">
                    {tr("petty_cash_col_begin", "ยอดยกมา (บาท)")}
                  </th>
                  <th className="w-44 px-3 py-3 text-right font-semibold text-foreground">
                    {tr("petty_cash_col_balance", "ยอดคงเหลือ (บาท)")}
                  </th>
                  <th className="w-16 px-3 py-3 text-center font-semibold text-muted-foreground">
                    {tr("actions", "จัดการ")}
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {loading ? (
                  <tr>
                    <td colSpan={7} className="py-12 text-center text-muted-foreground">
                      <div className="flex flex-col items-center justify-center gap-2">
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                        <span>{tr("loading", "กำลังโหลดข้อมูล...")}</span>
                      </div>
                    </td>
                  </tr>
                ) : rows.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="py-12 text-center text-muted-foreground">
                      <div className="flex flex-col items-center justify-center gap-2">
                        <Wallet className="h-10 w-10 text-muted-foreground/40" />
                        <span>{tr("petty_cash_no_data", "ยังไม่มีรายการวงเงินสดย่อย คลิก 'เพิ่มรายการ' เพื่อเริ่มต้น")}</span>
                      </div>
                    </td>
                  </tr>
                ) : (
                  rows.map((row, index) => (
                    <tr
                      key={row.id}
                      className="group transition-colors hover:bg-muted/40 focus-within:bg-muted/50"
                    >
                      <td className="px-3 py-2 text-center text-xs text-muted-foreground">
                        {index + 1}
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          value={row.code}
                          onChange={(e) => handleChangeRow(row.id, "code", e.target.value)}
                          placeholder="PC01"
                          maxLength={20}
                          className="h-9 font-mono uppercase text-sm"
                          disabled={busy}
                        />
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          value={row.name}
                          onChange={(e) => handleChangeRow(row.id, "name", e.target.value)}
                          placeholder="เงินสดย่อยสำนักงาน"
                          maxLength={100}
                          className="h-9 text-sm"
                          disabled={busy}
                        />
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          value={row.amount}
                          onChange={(e) => handleChangeRow(row.id, "amount", e.target.value)}
                          onBlur={() => handleBlurAmount(row.id, "amount")}
                          placeholder="0.00"
                          className="h-9 text-right font-mono text-sm"
                          disabled={busy}
                        />
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          value={row.begin_amount}
                          onChange={(e) => handleChangeRow(row.id, "begin_amount", e.target.value)}
                          onBlur={() => handleBlurAmount(row.id, "begin_amount")}
                          placeholder="0.00"
                          className="h-9 text-right font-mono text-sm"
                          disabled={busy}
                        />
                      </td>
                      <td className="px-3 py-2 text-right">
                        <span className="font-mono text-sm font-semibold text-primary">
                          {row.balance || "0.00"}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-center">
                        <div className="flex items-center justify-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleOpenTransModal(row)}
                            disabled={busy || !row.code.trim()}
                            className="h-8 w-8 text-primary hover:text-primary/80 hover:bg-primary/10"
                            title={tr("petty_cash_btn_trans", "บันทึกรับ/จ่ายเงินสดย่อย")}
                            aria-label={tr("petty_cash_btn_trans", "บันทึกรับ/จ่ายเงินสดย่อย")}
                          >
                            <Receipt className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleRemoveRow(row.id)}
                            disabled={busy}
                            className="h-8 w-8 text-muted-foreground hover:text-destructive"
                            title={tr("delete", "ลบแถวนี้")}
                            aria-label={tr("delete", "ลบแถวนี้")}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Modal บันทึกรับ/จ่ายเงินสดย่อย (Champ 5178 CASHFrmOtherIncome / 5179 CASHFrmOtherExpense) */}
      {transModalOpen && activePettyCash && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-xs">
          <div className="flex max-h-[90vh] w-full max-w-2xl flex-col rounded-xl border border-border bg-card shadow-2xl">
            {/* Header */}
            <div className="flex items-center justify-between border-b border-border px-5 py-4">
              <div className="flex items-center gap-2">
                <Receipt className="h-5 w-5 text-primary" />
                <h2 className="text-lg font-semibold text-foreground">
                  {tr("petty_cash_trans_title", "บันทึกรายการเงินสดย่อย")}
                </h2>
                <Badge variant="outline" className="font-mono text-xs">
                  {activePettyCash.code}
                </Badge>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setTransModalOpen(false)}
                className="h-8 w-8 text-muted-foreground hover:text-foreground"
                aria-label={tr("close", "ปิด")}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-5 space-y-5">
              {transError && (
                <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                  {transError}
                </div>
              )}

              {/* วงเงินและยอดคงเหลือปัจจุบัน */}
              <div className="grid grid-cols-3 gap-3 rounded-lg bg-muted/40 p-3 text-sm">
                <div>
                  <span className="text-muted-foreground">{tr("petty_cash_col_amount", "วงเงิน")}: </span>
                  <span className="font-mono font-semibold">{activePettyCash.amount || "0.00"}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">{tr("petty_cash_col_begin", "ยอดยกมา")}: </span>
                  <span className="font-mono font-semibold">{activePettyCash.begin_amount || "0.00"}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">{tr("petty_cash_col_balance", "คงเหลือ")}: </span>
                  <span className="font-mono font-semibold text-primary">{activePettyCash.balance || "0.00"}</span>
                </div>
              </div>

              {/* Form บันทึกรายการใหม่ */}
              <div className="rounded-lg border border-border p-4 space-y-4">
                <div className="font-medium text-sm text-foreground">
                  {tr("petty_cash_new_trans_section", "สร้างรายการใหม่ (รับชดเชย หรือจ่ายเงินสดย่อย)")}
                </div>

                <div className="flex items-center gap-4">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="trans_type"
                      checked={transType === 2}
                      onChange={() => setTransType(2)}
                      className="text-primary"
                    />
                    <span className="flex items-center gap-1 text-sm font-medium text-destructive">
                      <ArrowUpRight className="h-4 w-4" />
                      {tr("petty_cash_type_payment", "จ่ายเงินสดย่อย (5179)")}
                    </span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="trans_type"
                      checked={transType === 1}
                      onChange={() => setTransType(1)}
                      className="text-primary"
                    />
                    <span className="flex items-center gap-1 text-sm font-medium text-emerald-600 dark:text-emerald-400">
                      <ArrowDownLeft className="h-4 w-4" />
                      {tr("petty_cash_type_receipt", "รับ/เบิกชดเชย (5178)")}
                    </span>
                  </label>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <div>
                    <label className="text-xs text-muted-foreground mb-1 block">
                      {tr("doc_no", "เลขที่เอกสาร")} *
                    </label>
                    <Input
                      value={transDocNo}
                      onChange={(e) => setTransDocNo(e.target.value)}
                      placeholder="PC-001"
                      className="h-9 text-sm font-mono"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground mb-1 block">
                      {tr("doc_date", "วันที่เอกสาร")} *
                    </label>
                    <Input
                      type="date"
                      value={transDocDate}
                      onChange={(e) => setTransDocDate(e.target.value)}
                      className="h-9 text-sm"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground mb-1 block">
                      {tr("amount", "จำนวนเงิน (บาท)")} *
                    </label>
                    <Input
                      value={transAmount}
                      onChange={(e) => setTransAmount(e.target.value)}
                      placeholder="0.00"
                      className="h-9 text-right font-mono text-sm"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs text-muted-foreground mb-1 block">
                      {tr("description", "คำอธิบายรายการ")}
                    </label>
                    <Input
                      value={transDesc}
                      onChange={(e) => setTransDesc(e.target.value)}
                      placeholder="ค่าอุปกรณ์สำนักงาน / ค่าอาหารรับรอง"
                      className="h-9 text-sm"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground mb-1 block">
                      {transType === 1 ? tr("payer", "ผู้จ่ายเงิน/ผู้อนุมัติ") : tr("payee", "ผู้รับเงิน")}
                    </label>
                    <Input
                      value={transPayee}
                      onChange={(e) => setTransPayee(e.target.value)}
                      placeholder="ชื่อผู้ติดต่อ"
                      className="h-9 text-sm"
                    />
                  </div>
                </div>

                <div className="flex justify-end">
                  <Button
                    onClick={handleSaveTrans}
                    disabled={transSaving}
                    size="sm"
                    className="gap-2"
                  >
                    {transSaving && <Loader2 className="h-4 w-4 animate-spin" />}
                    <Save className="h-4 w-4" />
                    {tr("save_transaction", "บันทึกรายการ")}
                  </Button>
                </div>
              </div>

              {/* ประวัติรายการล่าสุด */}
              <div className="space-y-2">
                <div className="flex items-center gap-2 text-sm font-medium text-foreground">
                  <History className="h-4 w-4 text-muted-foreground" />
                  <span>{tr("petty_cash_recent_history", "ประวัติรายการล่าสุด")}</span>
                </div>

                <div className="max-h-60 overflow-y-auto rounded-lg border border-border">
                  <table className="w-full text-left text-sm">
                    <thead className="sticky top-0 bg-muted text-xs text-muted-foreground">
                      <tr>
                        <th className="px-3 py-2">{tr("doc_date", "วันที่")}</th>
                        <th className="px-3 py-2">{tr("doc_no", "เลขที่เอกสาร")}</th>
                        <th className="px-3 py-2">{tr("description", "รายละเอียด")}</th>
                        <th className="px-3 py-2 text-right">{tr("amount", "จำนวนเงิน")}</th>
                        <th className="w-10 px-2 py-2 text-center"></th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {transLoading ? (
                        <tr>
                          <td colSpan={5} className="py-6 text-center text-muted-foreground">
                            <Loader2 className="h-5 w-5 animate-spin mx-auto mb-1 text-primary" />
                            <span className="text-xs">{tr("loading", "กำลังโหลดประวัติ...")}</span>
                          </td>
                        </tr>
                      ) : transList.length === 0 ? (
                        <tr>
                          <td colSpan={5} className="py-6 text-center text-xs text-muted-foreground">
                            {tr("petty_cash_no_history", "ยังไม่มีประวัติการเคลื่อนไหว")}
                          </td>
                        </tr>
                      ) : (
                        transList.map((item) => (
                          <tr key={item.id} className="hover:bg-muted/30">
                            <td className="px-3 py-2 text-xs">{item.doc_date}</td>
                            <td className="px-3 py-2 font-mono text-xs">{item.doc_no}</td>
                            <td className="px-3 py-2 text-xs truncate max-w-[180px]">
                              {item.description || item.payee_payer || "-"}
                            </td>
                            <td className="px-3 py-2 text-right font-mono text-xs">
                              <span className={item.trans_type === 1 ? "text-emerald-600 dark:text-emerald-400 font-semibold" : "text-destructive font-semibold"}>
                                {item.trans_type === 1 ? `+${item.amount}` : `-${item.amount}`}
                              </span>
                            </td>
                            <td className="px-2 py-2 text-center">
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => handleDeleteTrans(item)}
                                className="h-6 w-6 text-muted-foreground hover:text-destructive"
                                title={tr("delete", "ลบรายการ")}
                                aria-label={tr("delete", "ลบรายการ")}
                              >
                                <Trash2 className="h-3 w-3" />
                              </Button>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="flex justify-end border-t border-border px-5 py-3">
              <Button variant="outline" onClick={() => setTransModalOpen(false)}>
                {tr("close", "ปิด")}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
