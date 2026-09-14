"use client";

import { useEffect, useRef, useState } from "react";
import { Eye, FileText, Pencil, Plus, RefreshCw, Save, Search, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { accountName, accountTypes, books, emptyAccount, emptyFiscalYear, emptyMaster, formatAmount, type GLAccount, type GLFiscalYear, type GLMaster, type GLRecord, type GLResource } from "@/lib/general-ledger";
import { GLCommandError, commandFailure, glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Check, Field, Notice, Pager, SearchInput, SplitWorkbench, YearSelect, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity } from "./gl-common";

type MasterResource = Exclude<GLResource, "journals">;
function newRecord(resource: MasterResource): GLRecord { return resource === "accounts" ? emptyAccount() : resource === "fiscal-years" ? emptyFiscalYear() : emptyMaster(); }
/** Exactly one alert is visible: the page-level notice for list failures, the pane notice for
 *  an open record. Errors are additive - the editor keeps the values the user already typed. */
export function editorAlert({ error, listError, refsError, hasRecord }: { error: string; listError?: string; refsError?: string; hasRecord: boolean }) {
  const pane = hasRecord ? error : "";
  const page = hasRecord ? "" : [error, listError, refsError].map((text) => text ?? "").filter(Boolean).join(" ");
  return { pane, page, count: (pane ? 1 : 0) + (page ? 1 : 0) };
}
export function pageErrorText(error: string, hasRecord: boolean) { return hasRecord ? "" : error; }
export function paneErrorText(error: string, hasRecord: boolean) { return hasRecord ? error : ""; }
export function FormErrorAlert({ text }: { text: string }) { return text ? <div className="shrink-0"><Notice error text={text} /></div> : null; }
/** The only state a failed save may touch: the message and the field to focus. Entered values are kept. */
export function errorStatePatch(info: { message: string; field: string }, code = "") { return { error: info.message, errorField: code === "duplicate_code" || code === "immutable_code" || code === "validation_failed" ? "accountcode" : info.field }; }

function recordCode(record: GLRecord) { return "accountcode" in record && "names" in record ? record.accountcode : "code" in record ? record.code : ""; }
function recordName(record: GLRecord) { return "names" in record ? accountName(record) : "name" in record ? record.name : "currency" in record ? record.currency : ""; }

export function GLMasters({ resource, route }: { resource: MasterResource; route: string }) {
  const [search, setSearch] = useState("");
  const list = useGLList<GLRecord>(resource, search);
  const searchDebounce = useDebouncedSearch({
    onSearch: (val) => {
      list.setPage(1);
      setSearch(val);
    },
    debounceMs: 2000,
  });
  const density = useRowDensity();
  const [revision, setRevision] = useState(0), refs = useReferences(revision);
  const [record, setRecord] = useState<GLRecord | null>(null), [original, setOriginal] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [message, setMessage] = useState(""), [error, setError] = useState(""), [reason, setReason] = useState("");
  const [errorField, setErrorField] = useState("");
  const formRef = useRef<HTMLFormElement>(null);
  const { busy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog();
  const dirty = isEditing && record !== null && JSON.stringify(record) !== original;
  useDirtyGuard(route, dirty);
  const set = (patch: object) => setRecord((current) => current ? { ...current, ...patch } as GLRecord : current);

  // Failures must never be silent: keep the Thai text in the pane the user is looking at and move
  // focus to the offending field (no scrolling, so the screen does not jump).
  function showFormError(cause: unknown, fallback: string, field = "") {
    const patch = errorStatePatch(commandFailure(cause, fallback, field), cause instanceof GLCommandError ? cause.code : "");
    setError(patch.error);
    setErrorField(patch.errorField);
  }
  useEffect(() => {
    if (!error || !errorField) return;
    const root = formRef.current?.querySelector<HTMLElement>(`[data-field="${errorField}"]`);
    const node = root?.matches("input,select,textarea,button") ? root : root?.querySelector<HTMLElement>("input,select,textarea,button");
    node?.focus({ preventScroll: true });
  }, [error, errorField]);

  async function openView(item: GLRecord) {
    if (dirty && !await confirm({ title: "ละทิ้งข้อมูลที่ยังไม่บันทึก?", description: "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
    try {
      const value = item?.id ? await glRequest<GLRecord>(`${resource}/${encodeURIComponent(item.id)}`) : newRecord(resource);
      const normalized = { ...newRecord(resource), ...value } as GLRecord;
      setRecord(normalized);
      setOriginal(JSON.stringify(normalized));
      setIsEditing(false);
      setMessage("");
      setError("");
      setReason("");
    } catch (e) {
      showFormError(e, "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง");
    }
  }

  async function openEdit(item: GLRecord) {
    if (dirty && !await confirm({ title: "ละทิ้งข้อมูลที่ยังไม่บันทึก?", description: "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
    try {
      const value = item?.id ? await glRequest<GLRecord>(`${resource}/${encodeURIComponent(item.id)}`) : newRecord(resource);
      const normalized = { ...newRecord(resource), ...value } as GLRecord;
      setRecord(normalized);
      setOriginal(JSON.stringify(normalized));
      setIsEditing(true);
      setMessage("");
      setError("");
      setReason("");
    } catch (e) {
      showFormError(e, "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง");
    }
  }

  async function openCreate() {
    if (dirty && !await confirm({ title: "ละทิ้งข้อมูลที่ยังไม่บันทึก?", description: "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
    const fresh = newRecord(resource);
    setRecord(fresh);
    setOriginal(JSON.stringify(fresh));
    setIsEditing(true);
    setMessage("");
    setError("");
    setReason("");
  }

  async function cancelEdit() {
    if (dirty && !await confirm({ title: "ละทิ้งข้อมูลที่ยังไม่บันทึก?", description: "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
    if (record?.id) {
      setRecord(JSON.parse(original));
      setIsEditing(false);
    } else {
      setRecord(null);
      setOriginal("");
      setIsEditing(false);
    }
    setError("");
    setReason("");
  }

  async function save() {
    if (!record || busy || !isEditing) return;
    if (!recordCode(record).trim() || resource !== "fiscal-years" && !recordName(record).trim()) { showFormError(null, "กรุณาระบุรหัสและชื่อให้ครบ", recordCode(record).trim() ? (isAcc ? "accountnameth" : "name") : (isAcc ? "accountcode" : "code")); return; }
    if (record.id && !await confirm({ title: "บันทึกการแก้ไขข้อมูล?", description: `แก้ไข ${recordCode(record)} โดยเก็บประวัติการเปลี่ยนแปลง`, confirmLabel: "บันทึกการแก้ไข", tone: "info" })) return;
    try {
      setError("");
      const field = resource === "accounts" ? { account: record as GLAccount } : resource === "fiscal-years" ? { fiscalyear: record as GLFiscalYear } : { master: record as GLMaster };
      const result = await execute({ resource, action: record.id ? "update" : "create", id: record.id, version: record.version, reason: reason.trim() || (record.id ? "แก้ไขข้อมูล" : "สร้างข้อมูลใหม่"), ...field });
      const saved = { ...record, id: result.id, version: result.version };
      setRecord(saved);
      setOriginal(JSON.stringify(saved));
      setIsEditing(false);
      list.reload();
      setRevision((value) => value + 1);
      setMessage(result.projectionpending ? "บันทึกแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน กดโหลดใหม่เพื่อตรวจสอบ" : "บันทึกเรียบร้อยแล้ว");
    } catch (e) { showFormError(e, "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"); }
  }

  async function deleteItem(item: GLRecord) {
    if (!item.id || busy) return;
    const isAcc = resource === "accounts";
    if (!await confirm({
      title: "ยืนยันลบรายการ?",
      description: `${recordCode(item)} · ${recordName(item)}`,
      details: isAcc ? "ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ใช้งานให้ปิดใช้งานแทน" : undefined,
      confirmLabel: "ลบรายการ",
      tone: "danger",
    })) return;
    try {
      setError("");
      await execute({ resource, id: item.id, version: item.version, action: "delete", reason: "ลบรายการ" });
      if (record?.id === item.id) {
        setRecord(null);
        setOriginal("");
        setIsEditing(false);
      }
      list.reload();
      setRevision((value) => value + 1);
      setMessage(`ลบ ${recordCode(item)} เรียบร้อยแล้ว`);
    } catch (e) {
      showFormError(e, "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง");
    }
  }

  async function runAction(action: "delete" | "lock" | "unlock") {
    if (!record?.id || busy) return;
    const label = action === "delete" ? "ลบรายการ" : action === "lock" ? "ล็อกงวดบัญชี" : "ปลดล็อกงวดบัญชี";
    const defaultReason = action === "delete" ? "ลบรายการ" : action === "lock" ? "ล็อกงวดบัญชี" : "ปลดล็อกงวดบัญชี";
    const effectiveReason = reason.trim() || defaultReason;
    if (!await confirm({
      title: `${label}?`,
      description: `${recordCode(record)} · ${effectiveReason}`,
      details: resource === "accounts" && action === "delete" ? "ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ใช้งานให้ปิดใช้งานแทน" : undefined,
      confirmLabel: label,
      tone: action === "delete" ? "danger" : "warning"
    })) return;
    try {
      setError("");
      await execute({ resource, id: record.id, version: record.version, action, reason: effectiveReason });
      setRecord(null); setOriginal(""); setIsEditing(false); list.reload(); setRevision((value) => value + 1); setMessage(`${label}เรียบร้อยแล้ว`);
    } catch (e) { showFormError(e, "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"); }
  }

  const hasAmount = ["budgets", "forecast"].includes(resource);
  const isAcc = resource === "accounts";
  // One alert at a time: the pane owns the message while a record is open (the page notice stays quiet),
  // otherwise the page notice shows the operation/list/reference failure.
  const alerts = editorAlert({ error, listError: list.error, refsError: refs.error, hasRecord: !!record });
  const colSpan = 4 + (hasAmount ? 1 : 0) + (isAcc ? 1 : 0);

  return <div className="flex flex-col flex-1 min-h-0 gap-2">
    <div className="shrink-0 flex flex-col gap-2"><Notice error text={alerts.page} /><Notice text={message} /></div>
    <SplitWorkbench list={<div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2 shrink-0">
        <form className="flex flex-wrap gap-2 flex-1 min-w-0" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
          <SearchInput
            className="min-w-40 flex-1"
            ariaLabel="ค้นหารหัสหรือชื่อ"
            placeholder="ค้นหารหัสหรือชื่อ"
            value={searchDebounce.query}
            onChange={searchDebounce.setQuery}
            onClear={searchDebounce.clear}
            onSearch={searchDebounce.searchNow}
          />
          <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />ค้นหา</Button>
          <Button type="button" variant="outline" className={actionClass} onClick={() => { list.reload(); refs.reload(); }} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />โหลดใหม่</Button>
          <Button type="button" className={actionClass} onClick={() => void openCreate()} disabled={busy}><Plus className="size-4 mr-1.5" />เพิ่มรายการ</Button>
          <Button type="button" variant="outline" className={actionClass} aria-pressed={density.compact} onClick={density.toggle}>{density.compact ? "ขยายบรรทัด" : "ย่อบรรทัด"}</Button>
        </form>
      </div>
      <div className="flex items-center justify-between px-1 text-xs text-muted-foreground shrink-0">
        <span>{list.data.total.toLocaleString("th-TH")} รายการ</span>
        {density.compact && <span className="text-[11px]">โหมดย่อบรรทัด</span>}
      </div>
      <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border" aria-busy={list.loading}>
        <table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}>
          <thead className="sticky top-0 bg-muted z-10">
            <tr>
              <th className="p-2.5">รหัส</th>
              <th className="p-2.5 min-w-44">ชื่อ / รายละเอียด</th>
              {hasAmount && <th className="p-2.5 text-right w-36">จำนวนเงิน</th>}
              {isAcc && <th className="p-2.5 text-center w-24">ระดับ</th>}
              <th className="p-2.5 text-center w-28">สถานะ</th>
              <th className="p-2.5 text-right pr-3 w-24">จัดการ</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {list.data.items.map((item, index) => {
              const accLevel = isAcc && "level" in item && typeof item.level === "number" ? item.level : 1;
              const isSelected = record?.id === item.id;
              const isRowEditing = isSelected && isEditing;
              const isLocked = "locked" in item && item.locked;
              const isClosed = "closed" in item && item.closed;
              const isActive = !("isactive" in item) || item.isactive;
              return (
                <tr
                  key={item.id}
                  className={`group cursor-pointer transition-colors ${
                    isRowEditing
                      ? "bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 ring-1 ring-inset ring-amber-500/50 font-medium"
                      : isSelected
                        ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
                        : index % 2 === 0
                          ? "bg-background hover:bg-accent/60"
                          : "bg-muted/20 hover:bg-accent/60"
                  }`}
                  onClick={() => void openView(item)}
                >
                  <td className="p-2 whitespace-nowrap">
                    <span className={`font-mono font-bold ${isRowEditing ? "text-amber-900 dark:text-amber-300" : "text-primary"}`}>{recordCode(item)}</span>
                  </td>
                  <td className="max-w-72 truncate p-2" title={recordName(item)}>
                    {isAcc && accLevel > 1 ? (
                      <span style={{ paddingLeft: `${(accLevel - 1) * 16}px` }} className="inline-flex items-center gap-1.5">
                        <span className="text-muted-foreground select-none font-mono">└─</span>
                        <span>{recordName(item)}</span>
                      </span>
                    ) : (
                      recordName(item)
                    )}
                  </td>
                  {hasAmount && (
                    <td className="p-2 text-right font-mono font-semibold tabular-nums whitespace-nowrap">
                      {"amount" in item ? formatAmount(item.amount) : "-"}
                    </td>
                  )}
                  {isAcc && (
                    <td className="whitespace-nowrap p-2 text-center">
                      <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
                        ระดับ {accLevel}
                      </span>
                    </td>
                  )}
                  <td className="whitespace-nowrap p-2 text-center">
                    {isLocked ? (
                      <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20">
                        ล็อกแล้ว
                      </span>
                    ) : isClosed ? (
                      <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground border border-border">
                        ปิดปีแล้ว
                      </span>
                    ) : !isActive ? (
                      <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground border border-border">
                        ปิดใช้งาน
                      </span>
                    ) : (
                      <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20">
                        ใช้งาน
                      </span>
                    )}
                  </td>
                  <td className="p-2 text-right whitespace-nowrap pr-2" onClick={(e) => e.stopPropagation()}>
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-md bg-background text-primary border-primary/30 hover:bg-primary/15 hover:border-primary/50 shadow-none transition-colors shrink-0"
                        onClick={(e) => {
                          e.stopPropagation();
                          void openEdit(item);
                        }}
                        aria-label="แก้ไข"
                        title="แก้ไข (Edit)"
                      >
                        <Pencil className="size-3.5 shrink-0" />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-md bg-background text-destructive border-border hover:bg-destructive/10 hover:border-destructive/40 focus-visible:border-destructive/50 focus-visible:ring-destructive/30 shadow-none transition-colors shrink-0"
                        onClick={(e) => {
                          e.stopPropagation();
                          void deleteItem(item);
                        }}
                        aria-label="ลบ"
                        title="ลบ (Delete)"
                      >
                        <Trash2 className="size-3.5 shrink-0" />
                      </Button>
                    </div>
                  </td>
                </tr>
              );
            })}
            {!list.data.items.length && (
              <tr>
                <td colSpan={colSpan} className="p-6 text-center text-muted-foreground">
                  {list.loading ? "กำลังโหลดข้อมูล…" : "ยังไม่มีรายการ กดเพิ่มรายการเพื่อเริ่มต้น"}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <Pager page={list.page} total={list.data.total} onPage={list.setPage} loading={list.loading} />
    </div>} editor={record ? (
      !isEditing ? (
        <div className="flex flex-col h-full min-h-0 gap-3">
          <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
            <div className="flex items-center gap-2">
              <div className="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary">
                <Eye className="size-4" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-base font-semibold">แสดงข้อมูล: {recordCode(record)}</h2>
                  <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold bg-muted text-muted-foreground border border-border">
                    โหมดแสดงข้อมูล
                  </span>
                </div>
                <p className="text-xs text-muted-foreground">{recordName(record)}</p>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              <Button
                type="button"
                className={actionClass}
                onClick={() => setIsEditing(true)}
                disabled={busy}
                title="แก้ไขข้อมูลนี้ (Edit)"
              >
                <Pencil className="size-4 mr-1.5" />แก้ไข
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                onClick={() => { setRecord(null); setOriginal(""); setIsEditing(false); }}
                title="ปิดหน้าต่างแสดงข้อมูล"
              >
                <X className="size-4" />
              </Button>
            </div>
          </header>
          <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground shrink-0">
            <Eye className="size-3.5 text-primary shrink-0" />
            <span>โหมดแสดงข้อมูล (Read-only) — หากต้องการแก้ไข ให้กดปุ่ม <strong>แก้ไข</strong></span>
          </div>
          {alerts.pane && <FormErrorAlert text={alerts.pane} />}
          <fieldset disabled className="grid gap-3 opacity-95 content-start overflow-y-auto p-1">
            {resource === "accounts" ? (
              <AccountFields value={record as GLAccount} set={set} accounts={refs.accounts} />
            ) : resource === "fiscal-years" ? (
              <FiscalYearFields value={record as GLFiscalYear} set={set} accounts={refs.accounts} />
            ) : (
              <MasterFields resource={resource} value={record as GLMaster} set={set} accounts={refs.accounts} years={refs.years} />
            )}
          </fieldset>
          <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                className={actionClass}
                onClick={() => setIsEditing(true)}
                disabled={busy}
              >
                <Pencil className="size-4 mr-1.5" />แก้ไขข้อมูล
              </Button>
              <Button
                type="button"
                variant="outline"
                className={actionClass}
                onClick={() => { setRecord(null); setOriginal(""); setIsEditing(false); }}
              >
                ปิด
              </Button>
              {resource === "periods" && record.id && (
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  disabled={busy}
                  onClick={() => void runAction((record as GLMaster).locked ? "unlock" : "lock")}
                >
                  {(record as GLMaster).locked ? "ปลดล็อกงวด" : "ล็อกงวด"}
                </Button>
              )}
            </div>
            {record.id && (
              <Button
                type="button"
                variant="outline"
                className={`${actionClass} text-destructive border-border hover:bg-destructive/10 hover:border-destructive/40 focus-visible:border-destructive/50 focus-visible:ring-destructive/30`}
                disabled={busy}
                onClick={() => void runAction("delete")}
              >
                <Trash2 className="size-4 mr-1.5" />ลบรายการนี้
              </Button>
            )}
          </footer>
        </div>
      ) : (
        <form ref={formRef} className="flex flex-col h-full min-h-0 gap-3" onSubmit={(event) => { event.preventDefault(); void save(); }}>
          <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
            <div className="flex items-center gap-2">
              <div className="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary">
                {record.id ? <Pencil className="size-4" /> : <Plus className="size-4" />}
              </div>
              <div>
                <h2 className="text-base font-semibold">{record.id ? `แก้ไข ${recordCode(record)}` : "เพิ่มรายการใหม่"}</h2>
                {dirty && <span className="text-xs text-amber-600 dark:text-amber-400">● มีการเปลี่ยนแปลงที่ยังไม่บันทึก</span>}
              </div>
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
              onClick={() => void cancelEdit()}
              title="ปิดหน้าต่างแก้ไข"
            >
              <X className="size-4" />
            </Button>
          </header>
          {alerts.pane && <FormErrorAlert text={alerts.pane} />}
          <fieldset disabled={busy} className="grid gap-3 opacity-95 content-start overflow-y-auto p-1">
            {resource === "accounts" ? (
              <AccountFields value={record as GLAccount} set={set} accounts={refs.accounts} />
            ) : resource === "fiscal-years" ? (
              <FiscalYearFields value={record as GLFiscalYear} set={set} accounts={refs.accounts} />
            ) : (
              <MasterFields resource={resource} value={record as GLMaster} set={set} accounts={refs.accounts} years={refs.years} />
            )}
            {record.id && (
              <Field label="เหตุผลการเปลี่ยนแปลง (ไม่บังคับ)">
                <input className={control} value={reason} onChange={(event) => setReason(event.target.value)} placeholder="ระบุเหตุผลการแก้ไข (ถ้ามี)" />
              </Field>
            )}
          </fieldset>
          <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
            <div className="flex flex-wrap gap-2">
              <Button type="submit" className={actionClass} disabled={busy || !dirty}>
                <Save className="size-4 mr-1.5" />{busy ? "กำลังบันทึก…" : "บันทึกข้อมูล"}
              </Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()}>
                ยกเลิก
              </Button>
              {resource === "periods" && record.id && (
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  disabled={busy || dirty}
                  onClick={() => void runAction((record as GLMaster).locked ? "unlock" : "lock")}
                >
                  {(record as GLMaster).locked ? "ปลดล็อกงวด" : "ล็อกงวด"}
                </Button>
              )}
            </div>
            {record.id && (
              <Button
                type="button"
                variant="outline"
                className={`${actionClass} text-destructive border-border hover:bg-destructive/10 hover:border-destructive/40 focus-visible:border-destructive/50 focus-visible:ring-destructive/30`}
                disabled={busy}
                onClick={() => void runAction("delete")}
              >
                <Trash2 className="size-4 mr-1.5" />ลบรายการนี้
              </Button>
            )}
          </footer>
        </form>
      )
    ) : (
      <div className="flex flex-col flex-1 h-full min-h-64 items-center justify-center content-center gap-3 text-center p-6">
        <div className="mx-auto grid size-12 place-items-center rounded-2xl bg-muted text-muted-foreground">
          <FileText className="size-6" />
        </div>
        <h2 className="text-base font-semibold">เลือกรายการเพื่อแสดงข้อมูล</h2>
        <p className="text-sm text-muted-foreground">คลิกที่แถวในตารางเพื่อแสดงข้อมูล หรือกดปุ่ม &ldquo;+ เพิ่มรายการ&rdquo; เพื่อสร้างข้อมูลใหม่</p>
        <div className="pt-2">
          <Button type="button" className={actionClass} onClick={() => void openCreate()}>
            <Plus className="size-4 mr-1.5" />เพิ่มรายการใหม่
          </Button>
        </div>
      </div>
    )} />
    {confirmationDialog}
  </div>;
}

function AccountFields({ value, set, accounts }: { value: GLAccount; set: (patch: object) => void; accounts: GLAccount[] }) {
  const onParentChange = (parentaccountcode: string) => {
    const parent = accounts.find((a) => a.accountcode === parentaccountcode);
    const suggestedLevel = parent ? (parent.level || 1) + 1 : 1;
    set({ parentaccountcode, level: Math.min(12, Math.max(1, suggestedLevel)) });
  };
  return <>{value.id && <Notice text="ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ใช้งานให้ปิดใช้งานแทน" />}<div className="grid gap-3 sm:grid-cols-2">
    <Field label="รหัสบัญชี"><input className={control} data-field="accountcode" required disabled={!!value.id} value={value.accountcode} onChange={(e) => set({ accountcode: e.target.value })} maxLength={60} /></Field>
    <Field label="ชื่อบัญชีภาษาไทย"><input className={control} data-field="accountnameth" required value={accountName(value)} onChange={(e) => set({ names: [...value.names.filter((name) => name.code !== "th"), { code: "th", name: e.target.value }] })} maxLength={300} /></Field>
    <Field label="ชื่อบัญชีภาษาอังกฤษ (ไม่บังคับ)"><input className={control} data-field="accountnameen" value={value.names.find((name) => name.code === "en")?.name ?? ""} onChange={(e) => set({ names: [...value.names.filter((name) => name.code !== "en"), { code: "en", name: e.target.value }] })} maxLength={300} placeholder="เช่น Cash on hand" /></Field>
    <Field label="หมวดบัญชี"><select className={control} value={value.accounttype} onChange={(e) => set({ accounttype: e.target.value })}>{Object.entries(accountTypes).map(([code, name]) => <option key={code} value={code}>{name}</option>)}</select></Field>
    <Field label="ยอดคงเหลือปกติ"><select className={control} value={value.normalbalance} onChange={(e) => set({ normalbalance: e.target.value })}><option value="debit">เดบิต</option><option value="credit">เครดิต</option></select></Field>
    <Field label="บัญชีแม่"><AccountSelect field="parentaccountcode" value={value.parentaccountcode} onChange={onParentChange} accounts={accounts.filter((item) => item.accountcode !== value.accountcode)} all label="บัญชีแม่" /></Field>
    <Field label="ระดับบัญชี (1–12)"><select className={control} data-field="level" aria-label="ระดับบัญชี" value={value.level ?? 1} onChange={(e) => set({ level: Number(e.target.value) })}>{Array.from({ length: 12 }, (_, i) => i + 1).map((lvl) => <option key={lvl} value={lvl}>ระดับ {lvl}</option>)}</select></Field>
    <Field label="รหัสกลุ่มผังบัญชี"><input className={control} value={value.accountgroup} onChange={(e) => set({ accountgroup: e.target.value })} /></Field>
  </div><div className="flex flex-wrap gap-x-4 gap-y-2"><Check label="เปิดใช้งาน" checked={value.isactive} onChange={(isactive) => set({ isactive })} /><Check label="อนุญาตให้ลงรายการ" checked={value.allowposting} onChange={(allowposting) => set({ allowposting })} /><Check label="บัญชีเงินสดและรายการเทียบเท่าเงินสด" checked={value.iscash} onChange={(iscash) => set({ iscash })} /></div></>;
}
function FiscalYearFields({ value, set, accounts }: { value: GLFiscalYear; set: (patch: object) => void; accounts: GLAccount[] }) {
  return <><Notice text="กำหนดปีบัญชีและสกุลเงินก่อนบันทึกรายวัน ระบบตรวจจำนวนทศนิยมตามปีบัญชีและไม่ปัดยอดให้อัตโนมัติ" /><div className="grid gap-3 sm:grid-cols-2">
    <Field label="รหัสปีบัญชี"><input className={control} required value={value.code} disabled={!!value.id} onChange={(e) => set({ code: e.target.value })} /></Field>
    <Field label="สกุลเงิน 3 ตัวอักษร"><input className={control} required placeholder="เช่น THB" pattern="[A-Z]{3}" maxLength={3} value={value.currency} onChange={(e) => set({ currency: e.target.value.toUpperCase() })} /></Field>
    <Field label="วันเริ่มต้นปีบัญชี"><input className={control} required type="date" value={value.startdate} onChange={(e) => set({ startdate: e.target.value })} /></Field>
    <Field label="วันสิ้นสุดปีบัญชี"><input className={control} required type="date" value={value.enddate} onChange={(e) => set({ enddate: e.target.value })} /></Field>
    <Field label="จำนวนตำแหน่งทศนิยม"><select className={control} value={value.scale} onChange={(e) => set({ scale: Number(e.target.value) })}>{Array.from({ length: 9 }, (_, scale) => <option key={scale} value={scale}>{scale} ตำแหน่ง</option>)}</select></Field>
    <Field label="บัญชีกำไรขาดทุน"><AccountSelect label="บัญชีกำไรขาดทุน" value={value.profitlossaccount ?? ""} onChange={(profitlossaccount) => set({ profitlossaccount })} accounts={accounts} /></Field>
    <Field label="บัญชีกำไรสะสม"><AccountSelect label="บัญชีกำไรสะสม" value={value.retainedearningsaccount} onChange={(retainedearningsaccount) => set({ retainedearningsaccount })} accounts={accounts} /></Field>
  </div><Check label="เปิดใช้งานปีบัญชี" checked={value.isactive} onChange={(isactive) => set({ isactive })} />{value.closed && <Notice text="ปีบัญชีนี้ปิดแล้ว" />}</>;
}
function MasterFields({ resource, value, set, accounts, years }: { resource: MasterResource; value: GLMaster; set: (patch: object) => void; accounts: GLAccount[]; years: GLFiscalYear[] }) {
  const dateFields = ["budgets", "periods", "forecast"].includes(resource);
  return <><div className="grid gap-3 sm:grid-cols-2">
    <Field label="รหัส"><input className={control} data-field="code" required disabled={!!value.id} value={value.code} onChange={(e) => set({ code: e.target.value })} /></Field>
    <Field label="ชื่อ"><input className={control} data-field="name" required value={value.name} onChange={(e) => set({ name: e.target.value })} /></Field>
    {dateFields && <><Field label="ปีบัญชี"><YearSelect value={value.fiscalyear} onChange={(fiscalyear) => set({ fiscalyear })} years={years} /></Field><Field label="วันเริ่มต้น"><input className={control} type="date" required value={value.startdate} onChange={(e) => set({ startdate: e.target.value })} /></Field><Field label="วันสิ้นสุด"><input className={control} type="date" required value={value.enddate} onChange={(e) => set({ enddate: e.target.value })} /></Field></>}
    {["budgets", "forecast"].includes(resource) && <><Field label="บัญชี"><AccountSelect value={value.accountcode} onChange={(accountcode) => set({ accountcode })} accounts={accounts} /></Field><Field label="จำนวนเงิน"><AmountInput required value={value.amount} onChange={(amount) => set({ amount })} /></Field><Field label="รหัสสาขา"><input className={control} value={value.branchcode} onChange={(e) => set({ branchcode: e.target.value })} /></Field><Field label="รหัสแผนก"><input className={control} value={value.departmentcode} onChange={(e) => set({ departmentcode: e.target.value })} /></Field><Field label="รหัสโครงการ"><input className={control} value={value.projectcode} onChange={(e) => set({ projectcode: e.target.value })} /></Field></>}
    {resource === "forecast" && <Field label="ทิศทางเงิน"><select className={control} value={value.direction} onChange={(e) => set({ direction: e.target.value })}><option value="in">เงินเข้า</option><option value="out">เงินออก</option></select></Field>}
    {resource === "product-account-groups" && <>{([ ["itemaccount", "บัญชีสินค้า"], ["costaccount", "บัญชีต้นทุนขาย"], ["revenueaccount", "บัญชีรายได้จากการขาย"] ] as const).map(([key, label]) => <Field key={key} label={label}><AccountSelect label={label} value={value[key] ?? ""} onChange={(accountcode) => set({ [key]: accountcode })} accounts={accounts} /></Field>)}</>}
    {resource === "mappings" && <Field label="สมุดรายวัน"><select className={control} value={value.bookcode} onChange={(e) => set({ bookcode: e.target.value })}>{Object.entries(books).map(([code, name]) => <option key={code} value={code}>{name}</option>)}</select></Field>}
  </div>
    {resource === "mappings" && <div className="grid gap-2"><h3 className="font-semibold">กฎการเชื่อมบัญชี</h3>{(value.rules ?? []).map((rule, index) => <div key={index} className="grid gap-2 rounded-xl border border-border p-2 sm:grid-cols-2"><AccountSelect label={`บัญชีในกฎ ${index + 1}`} value={rule.accountcode} accounts={accounts} onChange={(accountcode) => set({ rules: value.rules.map((item, i) => i === index ? { ...item, accountcode } : item) })} /><Field label={`ด้านบัญชีกฎ ${index + 1}`}><select className={control} value={rule.side} onChange={(e) => set({ rules: value.rules.map((item, i) => i === index ? { ...item, side: e.target.value } : item) })}><option value="debit">เดบิต</option><option value="credit">เครดิต</option></select></Field><Field label={`แหล่งจำนวนเงินกฎ ${index + 1}`}><input className={control} value={rule.source} placeholder="เลือกตามเอกสารต้นทางที่ระบบรองรับ" onChange={(e) => set({ rules: value.rules.map((item, i) => i === index ? { ...item, source: e.target.value } : item) })} /></Field><Button type="button" className={actionClass} variant="outline" onClick={() => set({ rules: value.rules.filter((_, i) => i !== index) })}>นำกฎ {index + 1} ออก</Button></div>)}<Button type="button" variant="outline" className={actionClass} onClick={() => set({ rules: [...(value.rules ?? []), { accountcode: "", side: "debit", source: "" }] })}>เพิ่มกฎการเชื่อมบัญชี</Button></div>}
    <Check label="เปิดใช้งาน" checked={value.isactive} onChange={(isactive) => set({ isactive })} />
  </>;
}
