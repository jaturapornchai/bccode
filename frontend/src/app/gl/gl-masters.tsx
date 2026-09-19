"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Eye, FileText, Pencil, Plus, RefreshCw, Save, Search, Trash2, X, FolderTree, List, ChevronDown, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ChoiceSelect, Combobox } from "@/components/ui/select";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { accountName, accountTypeLabels, bookLabels, emptyAccount, emptyFiscalYear, emptyMaster, formatAmount, type GLAccount, type GLFiscalYear, type GLMaster, type GLRecord, type GLResource, type GLTextFn } from "@/lib/general-ledger";
import { GLCommandError, commandFailure, glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Check, Field, Notice, Pager, SearchInput, SplitWorkbench, UnsavedBadge, YearSelect, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity, useGLText } from "./gl-common";
import { buildChartOfAccountsTree, filterAccountTree, type AccountTreeNode } from "@/lib/chart-of-accounts-tree";

type MasterResource = Exclude<GLResource, "journals">;
function newRecord(resource: MasterResource): GLRecord { return resource === "accounts" ? emptyAccount() : resource === "fiscal-years" ? emptyFiscalYear() : emptyMaster(); }
export function normalizeRecord(resource: MasterResource, raw: GLRecord): GLRecord {
  const base = { ...newRecord(resource), ...raw };
  if (resource === "accounts") {
    const acc = base as GLAccount;
    if (!Array.isArray(acc.names) || acc.names.length === 0) {
      acc.names = [{ code: "th", name: "" }];
    } else if (!acc.names.some((n) => n.code === "th")) {
      acc.names = [{ code: "th", name: "" }, ...acc.names];
    }
    acc.accountcode = acc.accountcode || "";
    acc.parentaccountcode = acc.parentaccountcode || "";
    acc.accountgroup = acc.accountgroup || "";
    acc.accounttype = acc.accounttype || "asset";
    acc.normalbalance = acc.normalbalance || "debit";
    acc.level = typeof acc.level === "number" ? acc.level : 1;
    acc.isactive = acc.isactive ?? true;
    acc.allowposting = acc.allowposting ?? true;
    acc.iscash = acc.iscash ?? false;
  }
  return base as GLRecord;
}
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
export function errorStatePatch(info: { message: string; field: string }, code = "", fallbackField = "") { return { error: info.message, errorField: code === "duplicate_code" || code === "immutable_code" || code === "validation_failed" ? "accountcode" : info.field || fallbackField }; }
/** ช่องที่ต้องโฟกัสเมื่อบันทึกไม่ผ่านและเซิร์ฟเวอร์ไม่ได้ระบุ field */
export function saveFailureTarget(resource: string) { return resource === "accounts" ? "accountcode" : "code"; }

function recordCode(record: GLRecord) { return "accountcode" in record && "names" in record ? record.accountcode : "code" in record ? record.code : ""; }
function recordName(record: GLRecord) { return "names" in record ? accountName(record) : "name" in record ? record.name : ""; }

export function TreeNodeRow({
  node,
  depth,
  selectedId,
  isEditing,
  expandedNodes,
  onToggleNode,
  onSelect,
  onEdit,
  onDelete,
  tr,
}: {
  node: AccountTreeNode;
  depth: number;
  selectedId?: string;
  isEditing: boolean;
  expandedNodes: Record<string, boolean>;
  onToggleNode: (code: string) => void;
  onSelect: (acc: GLAccount) => void;
  onEdit: (acc: GLAccount) => void;
  onDelete: (acc: GLAccount) => void;
  tr: GLTextFn;
}) {
  const acc = node.account;
  const isSelected = selectedId === acc.id;
  const isRowEditing = isSelected && isEditing;
  const isControl = !acc.allowposting;
  const isActive = acc.isactive ?? true;
  const isExpanded = expandedNodes[acc.accountcode] ?? (depth < 2);

  return (
    <div className="flex flex-col shrink-0">
      <div
        className={`group flex items-center justify-between gap-2 px-3 py-1.5 min-h-[38px] cursor-pointer transition-colors text-[0.95rem] border-b border-border/40 shrink-0 ${
          isRowEditing
            ? "bg-primary/15 hover:bg-primary/20 text-foreground ring-1 ring-inset ring-primary/50 font-medium"
            : isSelected
              ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
              : "hover:bg-accent/60"
        }`}
        style={{ paddingLeft: `${8 + depth * 20}px` }}
        onClick={() => onSelect(acc)}
      >
        <div className="flex items-center gap-2 min-w-0 flex-1">
          {node.hasChildren ? (
            <button
              type="button"
              className="p-0.5 text-muted-foreground hover:text-foreground rounded hover:bg-muted shrink-0 cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                onToggleNode(acc.accountcode);
              }}
              aria-label={isExpanded ? "Collapse" : "Expand"}
            >
              {isExpanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
            </button>
          ) : (
            <span className="w-4 shrink-0 text-muted-foreground/40 text-center font-mono select-none text-xs">
              {depth > 0 ? "└" : ""}
            </span>
          )}

          <span className="font-mono font-bold text-primary shrink-0 min-w-24">
            {acc.accountcode}
          </span>

          <span className="truncate text-foreground font-medium max-w-72" title={accountName(acc)}>
            {accountName(acc)}
          </span>

          <div className="flex items-center gap-1.5 shrink-0 ml-auto mr-2">
            <span className="inline-flex items-center rounded-md px-1.5 py-0.5 text-[11px] font-semibold bg-primary/10 text-primary border border-primary/20">
              L{node.level}
            </span>

            {isControl ? (
              <span className="inline-flex items-center rounded-md px-1.5 py-0.5 text-[11px] font-medium bg-muted text-muted-foreground border border-border">
                {tr("gl_control_account", "บัญชีคุม")}
              </span>
            ) : (
              <span className="inline-flex items-center rounded-md px-1.5 py-0.5 text-[11px] font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                {tr("gl_posting_account", "บัญชีย่อย")}
              </span>
            )}

            {!isActive && (
              <span className="inline-flex items-center rounded-md px-1.5 py-0.5 text-[11px] font-medium bg-rose-500/10 text-rose-600 border border-rose-500/20">
                {tr("gl_disable", "ปิดใช้งาน")}
              </span>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" onClick={(e) => e.stopPropagation()}>
          <Button
            type="button"
            size="icon"
            variant="outline"
            className="size-7 rounded-md bg-background text-primary border-primary/30 hover:bg-primary/15 shadow-none shrink-0"
            onClick={(e) => {
              e.stopPropagation();
              onEdit(acc);
            }}
            title={tr("gl_edit_label", "แก้ไข (Edit)")}
          >
            <Pencil className="size-3.5" />
          </Button>
          <Button
            type="button"
            size="icon"
            variant="outline"
            className="size-7 rounded-md bg-background text-destructive border-border hover:bg-destructive/10 shadow-none shrink-0"
            onClick={(e) => {
              e.stopPropagation();
              onDelete(acc);
            }}
            title={tr("gl_delete", "ลบ")}
          >
            <Trash2 className="size-3.5" />
          </Button>
        </div>
      </div>

      {isExpanded && node.children.map((child) => (
        <TreeNodeRow
          key={child.account.accountcode}
          node={child}
          depth={depth + 1}
          selectedId={selectedId}
          isEditing={isEditing}
          expandedNodes={expandedNodes}
          onToggleNode={onToggleNode}
          onSelect={onSelect}
          onEdit={onEdit}
          onDelete={onDelete}
          tr={tr}
        />
      ))}
    </div>
  );
}

export function GLMasters({ resource, route }: { resource: MasterResource; route: string }) {
  const tr = useGLText();
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
  const { busy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const isRecordDirty = record !== null && JSON.stringify(record) !== original;
  const isReasonDirty = reason.trim() !== "";
  const dirty = isEditing && record !== null && (isRecordDirty || isReasonDirty);

  useDirtyGuard(route, dirty);
  const set = (patch: object) => setRecord((current) => current ? { ...current, ...patch } as GLRecord : current);

  // Failures must never be silent: keep the Thai text in the pane the user is looking at and move
  // focus to the offending field (no scrolling, so the screen does not jump).
  function showFormError(cause: unknown, fallback: string, field = "") {
    const patch = errorStatePatch(commandFailure(cause, fallback, field), cause instanceof GLCommandError ? cause.code : "", saveFailureTarget(resource));
    setError(patch.error);
    setErrorField(patch.errorField);
  }
  useEffect(() => {
    // รอให้คำสั่งจบก่อน (fieldset ถูก disable ระหว่าง busy ทำให้ focus ไม่ติด) แล้วค่อยย้ายโฟกัส
    if (!error || !errorField || busy) return;
    const find = (name: string) => {
      const root = formRef.current?.querySelector<HTMLElement>(`[data-field="${name}"]`);
      return root?.matches("input,select,textarea,button") ? root : root?.querySelector<HTMLElement>("input,select,textarea,button");
    };
    const target = () => {
      const node = find(errorField) ?? find(saveFailureTarget(resource));
      if (node && document.activeElement !== node) node.focus({ preventScroll: true });
      return node;
    };
    if (target()) return;
    const frame = requestAnimationFrame(() => { target(); });
    return () => cancelAnimationFrame(frame);
  }, [error, errorField, busy]);

  async function openView(item: GLRecord) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLRecord>(`${resource}/${encodeURIComponent(item.id)}`) : newRecord(resource);
      const normalized = normalizeRecord(resource, value);
      setRecord(normalized);
      setOriginal(JSON.stringify(normalized));
      setIsEditing(false);
      setMessage("");
      setError("");
      setReason("");
    } catch (e) {
      showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    }
  }

  async function openEdit(item: GLRecord) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLRecord>(`${resource}/${encodeURIComponent(item.id)}`) : newRecord(resource);
      const normalized = normalizeRecord(resource, value);
      setRecord(normalized);
      setOriginal(JSON.stringify(normalized));
      setIsEditing(true);
      setMessage("");
      setError("");
      setReason("");
    } catch (e) {
      showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    }
  }

  async function openCreate() {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    const fresh = normalizeRecord(resource, newRecord(resource));
    setRecord(fresh);
    setOriginal(JSON.stringify(fresh));
    setIsEditing(true);
    setMessage("");
    setError("");
    setReason("");
  }

  async function cancelEdit() {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
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
    if (!recordCode(record).trim() || resource !== "fiscal-years" && !recordName(record).trim()) { showFormError(null, tr("gl_enter_code_and_name", "กรุณาระบุรหัสและชื่อให้ครบ"), recordCode(record).trim() ? (isAcc ? "accountnameth" : "name") : (isAcc ? "accountcode" : "code")); return; }
    if (record.id && !await confirm({ title: tr("gl_save_changes_confirm", "บันทึกการแก้ไขข้อมูล?"), description: tr("gl_edit_with_history", "แก้ไข {0} โดยเก็บประวัติการเปลี่ยนแปลง").replace("{0}", String(recordCode(record))), confirmLabel: tr("gl_save_changes", "บันทึกการแก้ไข"), tone: "info" })) return;
    try {
      setError("");
      const field = resource === "accounts" ? { account: record as GLAccount } : resource === "fiscal-years" ? { fiscalyear: record as GLFiscalYear } : { master: record as GLMaster };
      const result = await execute({ resource, action: record.id ? "update" : "create", id: record.id, version: record.version, reason: reason.trim() || (record.id ? tr("gl_edit_data", "แก้ไขข้อมูล") : tr("gl_create_new", "สร้างข้อมูลใหม่")), ...field });
      const saved = normalizeRecord(resource, { ...record, id: result.id, version: result.version });
      setRecord(saved);
      setOriginal(JSON.stringify(saved));
      setIsEditing(false);
      setReason("");
      list.reload();
      setRevision((value) => value + 1);
      setMessage(result.projectionpending ? tr("gl_saved_updating_reload", "บันทึกแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน กดโหลดใหม่เพื่อตรวจสอบ") : tr("gl_saved_successfully", "บันทึกเรียบร้อยแล้ว"));
    } catch (e) { showFormError(e, tr("gl_save_failed_retry", "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"), saveFailureTarget(resource)); }
  }

  async function deleteItem(item: GLRecord) {
    if (!item.id || busy) return;
    const isAcc = resource === "accounts";
    if (!await confirm({
      title: tr("gl_confirm_delete", "ยืนยันลบรายการ?"),
      description: `${recordCode(item)} · ${recordName(item)}`,
      details: isAcc ? tr("gl_coa_referenced_no_delete", "ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ใช้งานให้ปิดใช้งานแทน") : undefined,
      confirmLabel: tr("gl_delete_item", "ลบรายการ"),
      tone: "danger",
    })) return;
    try {
      setError("");
      await execute({ resource, id: item.id, version: item.version, action: "delete", reason: tr("gl_delete_item", "ลบรายการ") });
      if (record?.id === item.id) {
        setRecord(null);
        setOriginal("");
        setIsEditing(false);
      }
      list.reload();
      setRevision((value) => value + 1);
      setMessage(tr("gl_delete_success", "ลบ {0} เรียบร้อยแล้ว").replace("{0}", String(recordCode(item))));
    } catch (e) {
      showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    }
  }

  async function runAction(action: "delete" | "lock" | "unlock") {
    if (!record?.id || busy) return;
    const label = action === "delete" ? tr("gl_delete_item", "ลบรายการ") : action === "lock" ? tr("gl_lock_period_2", "ล็อกงวดบัญชี") : tr("gl_unlock_period_2", "ปลดล็อกงวดบัญชี");
    const defaultReason = action === "delete" ? tr("gl_delete_item", "ลบรายการ") : action === "lock" ? tr("gl_lock_period_2", "ล็อกงวดบัญชี") : tr("gl_unlock_period_2", "ปลดล็อกงวดบัญชี");
    const effectiveReason = reason.trim() || defaultReason;
    if (!await confirm({
      title: `${label}?`,
      description: `${recordCode(record)} · ${effectiveReason}`,
      details: resource === "accounts" && action === "delete" ? tr("gl_coa_referenced_no_delete", "ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ใช้งานให้ปิดใช้งานแทน") : undefined,
      confirmLabel: label,
      tone: action === "delete" ? "danger" : "warning"
    })) return;
    try {
      setError("");
      await execute({ resource, id: record.id, version: record.version, action, reason: effectiveReason });
      setRecord(null); setOriginal(""); setIsEditing(false); list.reload(); setRevision((value) => value + 1); setMessage(tr("gl_success_message", "{0}เรียบร้อยแล้ว").replace("{0}", String(label)));
    } catch (e) { showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")); }
  }

  const hasAmount = ["budgets", "forecast"].includes(resource);
  const isAcc = resource === "accounts";
  // One alert at a time: the pane owns the message while a record is open (the page notice stays quiet),
  // otherwise the page notice shows the operation/list/reference failure.
  const alerts = editorAlert({ error, listError: list.error, refsError: refs.error, hasRecord: !!record });
  const colSpan = 4 + (hasAmount ? 1 : 0) + (isAcc ? 1 : 0);

  const [viewMode, setViewMode] = useState<"list" | "tree">("list");
  const [expandedCategories, setExpandedCategories] = useState<Record<string, boolean>>({
    asset: true,
    liability: true,
    equity: true,
    income: true,
    expense: true,
  });
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({});

  const toggleCategory = (cat: string) => {
    setExpandedCategories((prev) => ({ ...prev, [cat]: !prev[cat] }));
  };

  const toggleNode = (code: string) => {
    setExpandedNodes((prev) => ({ ...prev, [code]: !(prev[code] ?? true) }));
  };

  const treeAccounts = useMemo(() => {
    return refs.accounts.length > 0 ? refs.accounts : (list.data.items as GLAccount[]);
  }, [refs.accounts, list.data.items]);

  const expandAll = () => {
    setExpandedCategories({ asset: true, liability: true, equity: true, income: true, expense: true });
    const allNodes: Record<string, boolean> = {};
    for (const a of treeAccounts) {
      allNodes[a.accountcode] = true;
    }
    setExpandedNodes(allNodes);
  };

  const collapseAll = () => {
    setExpandedCategories({ asset: false, liability: false, equity: false, income: false, expense: false });
    setExpandedNodes({});
  };

  const treeGroups = useMemo(() => {
    if (!isAcc) return [];
    const base = buildChartOfAccountsTree(treeAccounts);
    if (!searchDebounce.query.trim()) return base;
    return filterAccountTree(base, searchDebounce.query);
  }, [isAcc, treeAccounts, searchDebounce.query]);

  return <div className="flex flex-col flex-1 min-h-0 gap-2">
    <div className="shrink-0 flex flex-col gap-2"><Notice error text={alerts.page} /><Notice text={message} /></div>
    <SplitWorkbench list={<div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="shrink-0 pb-3 border-b border-border/70 flex flex-col gap-2">
        <form className="flex flex-wrap items-center gap-2 w-full" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
          <SearchInput
            className="min-w-36 flex-1 basis-44"
            ariaLabel={tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ")}
            placeholder={tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ")}
            value={searchDebounce.query}
            onChange={searchDebounce.setQuery}
            onClear={searchDebounce.clear}
            onSearch={searchDebounce.searchNow}
          />
          <div className="flex flex-wrap items-center gap-2 min-w-0">
            <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />{tr("gl_search", "ค้นหา")}</Button>
            <Button type="button" variant="outline" className={actionClass} onClick={() => { list.reload(); refs.reload(); }} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />{tr("gl_reload", "โหลดใหม่")}</Button>
            <Button type="button" className={actionClass} onClick={() => void openCreate()} disabled={busy}><Plus className="size-4 mr-1.5" />{tr("gl_add_row", "เพิ่มรายการ")}</Button>
            {isAcc && (
              <Button
                type="button"
                variant={viewMode === "tree" ? "default" : "outline"}
                className={actionClass}
                onClick={() => setViewMode((m) => m === "tree" ? "list" : "tree")}
                title={viewMode === "tree" ? tr("gl_view_list", "มุมมองตาราง") : tr("gl_view_tree", "มุมมองผังต้นไม้")}
              >
                {viewMode === "tree" ? (
                  <>
                    <List className="size-4 mr-1.5" />
                    {tr("gl_view_list", "มุมมองตาราง")}
                  </>
                ) : (
                  <>
                    <FolderTree className="size-4 mr-1.5" />
                    {tr("gl_view_tree", "มุมมองผังต้นไม้")}
                  </>
                )}
              </Button>
            )}
          </div>
        </form>
        <div className="flex items-center justify-between px-1 text-xs text-muted-foreground">
          <div className="flex items-center gap-2">
            <span className="font-medium text-foreground/80">
              {isAcc && viewMode === "tree"
                ? tr("gl_x_items", "{0} รายการ").replace("{0}", String(treeAccounts.length.toLocaleString("th-TH")))
                : tr("gl_x_items", "{0} รายการ").replace("{0}", String(list.data.total.toLocaleString("th-TH")))}
            </span>
            {isAcc && viewMode === "tree" && (
              <span className="inline-flex items-center rounded-md bg-primary/10 text-primary px-2 py-0.5 text-[11px] font-medium border border-primary/20">
                {tr("gl_view_tree", "มุมมองผังต้นไม้")}
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            {isAcc && viewMode === "tree" && (
              <>
                <button
                  type="button"
                  className="text-xs text-muted-foreground hover:text-foreground hover:underline cursor-pointer"
                  onClick={expandAll}
                >
                  {tr("gl_expand_all", "ขยายทั้งหมด")}
                </button>
                <span className="text-border select-none">·</span>
                <button
                  type="button"
                  className="text-xs text-muted-foreground hover:text-foreground hover:underline cursor-pointer"
                  onClick={collapseAll}
                >
                  {tr("gl_collapse_all", "ยุบทั้งหมด")}
                </button>
              </>
            )}
          </div>
        </div>
      </div>
      {isAcc && viewMode === "tree" ? (
        <div className="flex-1 min-h-[300px] overflow-y-auto rounded-xl border border-border shadow-sm p-2 flex flex-col gap-2.5 bg-card" aria-busy={list.loading}>
          {treeGroups.map((group) => {
            const isCatExpanded = expandedCategories[group.category] ?? true;
            return (
              <div key={group.category} className="shrink-0 rounded-xl border border-border/70 overflow-hidden bg-background shadow-xs">
                {/* Category Header */}
                <button
                  type="button"
                  className="w-full flex items-center justify-between p-2.5 bg-muted/40 hover:bg-muted/70 transition-colors cursor-pointer text-left select-none shrink-0"
                  onClick={() => toggleCategory(group.category)}
                >
                  <div className="flex items-center gap-2">
                    <span className="p-0.5 rounded-md text-muted-foreground">
                      {isCatExpanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
                    </span>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-bold border ${group.colorClass.badge}`}>
                      {group.categoryNumber}. {group.nameTh} ({group.nameEn})
                    </span>
                  </div>
                  <span className="text-xs text-muted-foreground font-medium px-2 py-0.5 rounded-full bg-muted">
                    {tr("gl_x_items", "{0} รายการ").replace("{0}", String(group.totalAccounts))}
                  </span>
                </button>

                {/* Root Nodes */}
                {isCatExpanded && (
                  <div className="flex flex-col shrink-0">
                    {group.rootNodes.map((node) => (
                      <TreeNodeRow
                        key={node.account.accountcode}
                        node={node}
                        depth={0}
                        selectedId={record?.id}
                        isEditing={isEditing}
                        expandedNodes={expandedNodes}
                        onToggleNode={toggleNode}
                        onSelect={(acc) => void openView(acc)}
                        onEdit={(acc) => void openEdit(acc)}
                        onDelete={(acc) => void deleteItem(acc)}
                        tr={tr}
                      />
                    ))}
                    {group.rootNodes.length === 0 && (
                      <div className="p-3 text-xs text-muted-foreground text-center">
                        {tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
          {treeGroups.every((g) => g.totalAccounts === 0) && (
            <div className="p-8 text-center text-muted-foreground">
              {tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}
            </div>
          )}
        </div>
      ) : (
        <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm" aria-busy={list.loading}>
          <table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}>
            <thead className="sticky top-0 bg-muted z-10">
              <tr>
                <th className="p-2.5">{tr("gl_code", "รหัส")}</th>
                <th className="p-2.5 min-w-44">{tr("gl_name_description", "ชื่อ / รายละเอียด")}</th>
                {hasAmount && <th className="p-2.5 text-right w-36">{tr("gl_amount", "จำนวนเงิน")}</th>}
                {isAcc && <th className="p-2.5 text-center w-24">{tr("gl_level", "ระดับ")}</th>}
                <th className="p-2.5 text-center w-28">{tr("gl_status", "สถานะ")}</th>
                <th className="p-2.5 text-right pr-3 w-24">{tr("gl_manage", "จัดการ")}</th>
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
                        ? "bg-primary/15 hover:bg-primary/20 text-foreground ring-1 ring-inset ring-primary/50 font-medium"
                        : isSelected
                          ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
                          : index % 2 === 0
                            ? "bg-background hover:bg-accent/60"
                            : "bg-muted/20 hover:bg-accent/60"
                    }`}
                    onClick={() => void openView(item)}
                  >
                    <td className="p-2 whitespace-nowrap">
                      <span className="font-mono font-bold text-primary">{recordCode(item)}</span>
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
                          {tr("gl_level_2", "ระดับ {0}").replace("{0}", String(accLevel))}
                        </span>
                      </td>
                    )}
                    <td className="whitespace-nowrap p-2 text-center">
                      {isLocked ? (
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground border border-border">
                          {tr("gl_locked", "ล็อกแล้ว")}
                        </span>
                      ) : isClosed ? (
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground border border-border">
                          {tr("gl_year_closed", "ปิดปีแล้ว")}
                        </span>
                      ) : !isActive ? (
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground border border-border">
                          {tr("gl_disable", "ปิดใช้งาน")}
                        </span>
                      ) : (
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary border border-primary/20">
                          {tr("gl_enable", "ใช้งาน")}
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
                          aria-label={tr("gl_edit", "แก้ไข")}
                          title={tr("gl_edit_label", "แก้ไข (Edit)")}
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
                          aria-label={tr("gl_delete", "ลบ")}
                          title={tr("gl_delete_label", "ลบ (Delete)")}
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
                    {list.loading ? tr("gl_loading_data", "กำลังโหลดข้อมูล…") : tr("gl_no_entries_add", "ยังไม่มีรายการ กดเพิ่มรายการเพื่อเริ่มต้น")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
      {!isAcc ? (
        <Pager page={list.page} total={list.data.total} onPage={list.setPage} loading={list.loading} />
      ) : (
        <div className="flex items-center justify-between border-t border-border pt-2 text-[0.95rem] text-muted-foreground">
          <span>{tr("gl_x_items", "{0} รายการ").replace("{0}", list.data.total.toLocaleString("th-TH"))} ({tr("gl_all_types", "ทั้งหมด")})</span>
        </div>
      )}
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
                  <h2 className="text-base font-semibold">{tr("gl_show_data", "แสดงข้อมูล: {0}").replace("{0}", String(recordCode(record)))}</h2>
                  <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold bg-muted text-muted-foreground border border-border">
                    {tr("gl_display_mode", "โหมดแสดงข้อมูล")}
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
                title={tr("gl_edit_this_data", "แก้ไขข้อมูลนี้ (Edit)")}
              >
                <Pencil className="size-4 mr-1.5" />{tr("gl_edit", "แก้ไข")}
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                onClick={() => { setRecord(null); setOriginal(""); setIsEditing(false); }}
                aria-label={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}
                title={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}
              >
                <X className="size-4" />
              </Button>
            </div>
          </header>
          <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground shrink-0">
            <Eye className="size-3.5 text-primary shrink-0" />
            <span>{tr("gl_readonly_mode_edit_hint", "โหมดแสดงข้อมูล (Read-only) — หากต้องการแก้ไข ให้กดปุ่ม")} <strong>{tr("gl_edit", "แก้ไข")}</strong></span>
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
                <Pencil className="size-4 mr-1.5" />{tr("gl_edit_data", "แก้ไขข้อมูล")}
              </Button>
              <Button
                type="button"
                variant="outline"
                className={actionClass}
                onClick={() => { setRecord(null); setOriginal(""); setIsEditing(false); }}
              >
                {tr("gl_close", "ปิด")}
              </Button>
              {resource === "periods" && record.id && (
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  disabled={busy}
                  onClick={() => void runAction((record as GLMaster).locked ? "unlock" : "lock")}
                >
                  {(record as GLMaster).locked ? tr("gl_unlock_period", "ปลดล็อกงวด") : tr("gl_lock_period", "ล็อกงวด")}
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
                <Trash2 className="size-4 mr-1.5" />{tr("gl_delete_item_2", "ลบรายการนี้")}
              </Button>
            )}
          </footer>
        </div>
      ) : (
        <form ref={formRef} className="flex flex-col h-full min-h-0 gap-3" onSubmit={(event) => { event.preventDefault(); void save(); }}>
          <header className="flex items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
            <div className="flex min-w-0 items-center gap-2">
              <div className="grid size-8 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                {record.id ? <Pencil className="size-4" /> : <Plus className="size-4" />}
              </div>
              <h2 className="truncate text-base font-semibold">
                {record.id ? tr("gl_edit_2", "แก้ไข {0}").replace("{0}", String(recordCode(record))) : tr("gl_add_new_item", "เพิ่มรายการใหม่")}
              </h2>
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <UnsavedBadge dirty={dirty} />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                onClick={() => void cancelEdit()}
                aria-label={tr("gl_close_edit_dialog", "ปิดหน้าต่างแก้ไข")}
                title={tr("gl_close_edit_dialog", "ปิดหน้าต่างแก้ไข")}
              >
                <X className="size-4" />
              </Button>
            </div>
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
              <Field label={tr("gl_change_reason_optional", "เหตุผลการเปลี่ยนแปลง (ไม่บังคับ)")}>
                <input className={control} value={reason} onChange={(event) => setReason(event.target.value)} placeholder={tr("gl_edit_reason_hint", "ระบุเหตุผลการแก้ไข (ถ้ามี)")} />
              </Field>
            )}
          </fieldset>
          <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
            <div className="flex flex-wrap gap-2">
              <Button type="submit" className={actionClass} disabled={busy || !dirty}>
                <Save className="size-4 mr-1.5" />{busy ? tr("gl_saving", "กำลังบันทึก…") : tr("gl_save_data", "บันทึกข้อมูล")}
              </Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()}>
                {tr("gl_cancel", "ยกเลิก")}
              </Button>
              {resource === "periods" && record.id && (
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  disabled={busy || dirty}
                  onClick={() => void runAction((record as GLMaster).locked ? "unlock" : "lock")}
                >
                  {(record as GLMaster).locked ? tr("gl_unlock_period", "ปลดล็อกงวด") : tr("gl_lock_period", "ล็อกงวด")}
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
                <Trash2 className="size-4 mr-1.5" />{tr("gl_delete_item_2", "ลบรายการนี้")}
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
        <h2 className="text-base font-semibold">{tr("gl_select_entry_to_view", "เลือกรายการเพื่อแสดงข้อมูล")}</h2>
        <p className="text-sm text-muted-foreground">{tr("gl_click_row_to_view_or_add", "คลิกที่แถวในตารางเพื่อแสดงข้อมูล หรือกดปุ่ม “+ เพิ่มรายการ” เพื่อสร้างข้อมูลใหม่")}</p>
        <div className="pt-2">
          <Button type="button" className={actionClass} onClick={() => void openCreate()}>
            <Plus className="size-4 mr-1.5" />{tr("gl_add_new_item", "เพิ่มรายการใหม่")}
          </Button>
        </div>
      </div>
    )} />
    {confirmationDialog}
  </div>;
}

function AccountFields({ value, set }: { value: GLAccount; set: (patch: object) => void; accounts?: GLAccount[] }) {
  const tr = useGLText();
  const thName = (value.names || []).find((name) => name.code === "th")?.name ?? "";
  const enName = (value.names || []).find((name) => name.code === "en")?.name ?? "";

  const handleTypeChange = (accounttype: string) => {
    const normalbalance = accounttype === "asset" || accounttype === "expense" ? "debit" : "credit";
    set({ accounttype, normalbalance });
  };

  return (
    <>
      {value.id && (
        <Notice
          text={tr(
            "gl_coa_referenced_no_delete",
            "ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ใช้งานให้ปิดใช้งานแทน"
          )}
        />
      )}
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label={tr("gl_account_code", "รหัสบัญชี")}>
          <input
            className={control}
            data-field="accountcode"
            required
            disabled={!!value.id}
            value={value.accountcode || ""}
            onChange={(e) => set({ accountcode: e.target.value })}
            maxLength={20}
          />
        </Field>
        <Field label={tr("gl_account_level_range", "ระดับบัญชี (1–12)")}>
          <Combobox
            data-field="level"
            aria-label={tr("gl_account_level", "ระดับบัญชี")}
            value={value.level ?? 1}
            onChange={(level) => set({ level: Number(level) })}
          >
            {Array.from({ length: 12 }, (_, i) => i + 1).map((lvl) => (
              <option key={lvl} value={lvl}>
                {tr("gl_level_2", "ระดับ {0}").replace("{0}", String(lvl))}
              </option>
            ))}
          </Combobox>
        </Field>
        <Field label={tr("gl_account_name_th", "ชื่อบัญชีภาษาไทย")}>
          <input
            className={control}
            data-field="accountnameth"
            required
            value={thName}
            onChange={(e) => {
              const currentNames = (value.names || []).filter((name) => name.code !== "th");
              set({ names: [...currentNames, { code: "th", name: e.target.value }] });
            }}
            maxLength={100}
          />
        </Field>
        <Field label={tr("gl_account_name_en", "ชื่อบัญชีภาษาอังกฤษ")}>
          <input
            className={control}
            data-field="accountnameen"
            value={enName}
            onChange={(e) => {
              const currentNames = (value.names || []).filter((name) => name.code !== "en");
              set({ names: [...currentNames, { code: "en", name: e.target.value }] });
            }}
            maxLength={100}
            placeholder={tr("gl_optional_example_cash_on_hand", "ไม่บังคับ เช่น Cash on hand")}
          />
        </Field>
        <Field label={tr("gl_account_category", "หมวดบัญชี")}>
          <Combobox
            value={value.accounttype || "asset"}
            onChange={handleTypeChange}
          >
            {Object.entries(accountTypeLabels).map(([code, name]) => (
              <option key={code} value={code}>
                {tr(...name)}
              </option>
            ))}
          </Combobox>
        </Field>
        <Field label={tr("gl_normal_balance", "ยอดคงเหลือปกติ")}>
          <div className="flex min-h-[2.6em] w-full items-center justify-between gap-2 rounded-xl border border-input bg-muted/20 px-3 py-1.5 text-[0.95rem] leading-normal text-foreground select-none shadow-[0_3px_10px_rgba(0,0,0,0.14),0_1px_3px_rgba(0,0,0,0.1)] dark:shadow-[0_3px_10px_rgba(0,0,0,0.6)]">
            <div className="flex items-center gap-2 font-semibold">
              {value.accounttype === "asset" || value.accounttype === "expense" ? (
                <span className="text-blue-600 dark:text-blue-400">{tr("gl_debit", "เดบิต")} (Dr.)</span>
              ) : (
                <span className="text-emerald-600 dark:text-emerald-400">{tr("gl_credit", "เครดิต")} (Cr.)</span>
              )}
            </div>
            <span className="text-xs text-muted-foreground font-normal">({tr("gl_auto_by_category", "กำหนดตามหมวดบัญชี")})</span>
          </div>
        </Field>
        <div className="sm:col-span-2 flex flex-wrap items-center gap-x-6 gap-y-2 pt-1 border-t border-border/50">
          <Check
            label={tr("gl_enabled", "เปิดใช้งาน")}
            checked={value.isactive ?? true}
            onChange={(isactive) => set({ isactive })}
          />
          <Check
            label={tr("gl_allow_posting", "บันทึกบัญชีได้ (IsGLAccess)")}
            checked={value.allowposting ?? true}
            onChange={(allowposting) => set({ allowposting })}
          />
        </div>
      </div>
    </>
  );
}
function FiscalYearFields({ value, set, accounts }: { value: GLFiscalYear; set: (patch: object) => void; accounts: GLAccount[] }) {
  const tr = useGLText();
  return <><Notice text={tr("gl_set_fy_currency_before_journal", "กำหนดปีบัญชีก่อนบันทึกรายวัน ระบบตรวจจำนวนทศนิยมตามปีบัญชีและไม่ปัดยอดให้อัตโนมัติ")} /><div className="grid gap-3 sm:grid-cols-2">
    <Field label={tr("gl_fiscal_year_code", "รหัสปีบัญชี")}><input className={control} required value={value.code || ""} disabled={!!value.id} onChange={(e) => set({ code: e.target.value })} /></Field>
    <Field label={tr("gl_fiscal_year_start_date", "วันเริ่มต้นปีบัญชี")}><input className={control} required type="date" value={value.startdate || ""} onChange={(e) => set({ startdate: e.target.value })} /></Field>
    <Field label={tr("gl_fiscal_year_end_date", "วันสิ้นสุดปีบัญชี")}><input className={control} required type="date" value={value.enddate || ""} onChange={(e) => set({ enddate: e.target.value })} /></Field>
    <Field label={tr("gl_decimal_places", "จำนวนตำแหน่งทศนิยม")}><Combobox value={value.scale ?? 2} onChange={(scale) => set({ scale: Number(scale) })}>{Array.from({ length: 9 }, (_, scale) => <option key={scale} value={scale}>{tr("gl_positions_count", "{0} ตำแหน่ง").replace("{0}", String(scale))}</option>)}</Combobox></Field>
    <Field label={tr("gl_profit_loss_account", "บัญชีกำไรขาดทุน")}><AccountSelect label={tr("gl_profit_loss_account", "บัญชีกำไรขาดทุน")} value={value.profitlossaccount || ""} onChange={(profitlossaccount) => set({ profitlossaccount })} accounts={accounts} /></Field>
    <Field label={tr("gl_retained_earnings_account", "บัญชีกำไรสะสม")}><AccountSelect label={tr("gl_retained_earnings_account", "บัญชีกำไรสะสม")} value={value.retainedearningsaccount || ""} onChange={(retainedearningsaccount) => set({ retainedearningsaccount })} accounts={accounts} /></Field>
    <div className="sm:col-span-2 flex flex-wrap items-center gap-x-6 gap-y-2 pt-1">
      <Check label={tr("gl_activate_fiscal_year", "เปิดใช้งานปีบัญชี")} checked={value.isactive ?? true} onChange={(isactive) => set({ isactive })} />
    </div>
  </div>{value.closed && <Notice text={tr("gl_fiscal_year_closed", "ปีบัญชีนี้ปิดแล้ว")} />}</>;
}
function MasterFields({ resource, value, set, accounts, years }: { resource: MasterResource; value: GLMaster; set: (patch: object) => void; accounts: GLAccount[]; years: GLFiscalYear[] }) {
  const tr = useGLText();
  const dateFields = ["budgets", "periods", "forecast"].includes(resource);
  return <><div className="grid gap-3 sm:grid-cols-2">
    <Field label={tr("gl_code", "รหัส")}><input className={control} data-field="code" required disabled={!!value.id} value={value.code || ""} onChange={(e) => set({ code: e.target.value })} /></Field>
    <Field label={tr("gl_name", "ชื่อ")}><input className={control} data-field="name" required value={value.name || ""} onChange={(e) => set({ name: e.target.value })} /></Field>
    {dateFields && <><Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect value={value.fiscalyear || ""} onChange={(fiscalyear) => set({ fiscalyear })} years={years} /></Field><Field label={tr("gl_start_date", "วันเริ่มต้น")}><input className={control} type="date" required value={value.startdate || ""} onChange={(e) => set({ startdate: e.target.value })} /></Field><Field label={tr("gl_end_date", "วันสิ้นสุด")}><input className={control} type="date" required value={value.enddate || ""} onChange={(e) => set({ enddate: e.target.value })} /></Field></>}
    {["budgets", "forecast"].includes(resource) && <><Field label={tr("gl_account", "บัญชี")}><AccountSelect value={value.accountcode || ""} onChange={(accountcode) => set({ accountcode })} accounts={accounts} /></Field><Field label={tr("gl_amount", "จำนวนเงิน")}><AmountInput required value={value.amount || ""} onChange={(amount) => set({ amount })} /></Field><Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={value.branchcode || ""} onChange={(e) => set({ branchcode: e.target.value })} /></Field><Field label={tr("gl_department_code", "รหัสแผนก")}><input className={control} value={value.departmentcode || ""} onChange={(e) => set({ departmentcode: e.target.value })} /></Field><Field label={tr("gl_project_code", "รหัสโครงการ")}><input className={control} value={value.projectcode || ""} onChange={(e) => set({ projectcode: e.target.value })} /></Field></>}
    {resource === "forecast" && <Field label={tr("gl_money_direction", "ทิศทางเงิน")}><ChoiceSelect value={value.direction || "in"} onChange={(direction) => set({ direction })}><option value="in">{tr("gl_money_in", "เงินเข้า")}</option><option value="out">{tr("gl_money_out", "เงินออก")}</option></ChoiceSelect></Field>}
    {resource === "product-account-groups" && <>{([ ["itemaccount", tr("gl_inventory_account", "บัญชีสินค้า")], ["costaccount", tr("gl_cogs_account", "บัญชีต้นทุนขาย")], ["revenueaccount", tr("gl_sales_revenue_account", "บัญชีรายได้จากการขาย")] ] as const).map(([key, label]) => <Field key={key} label={label}><AccountSelect label={label} value={value[key] ?? ""} onChange={(accountcode) => set({ [key]: accountcode })} accounts={accounts} /></Field>)}</>}
    {resource === "mappings" && <Field label={tr("gl_journal", "สมุดรายวัน")}><Combobox value={value.bookcode || ""} onChange={(bookcode) => set({ bookcode })}>{Object.entries(bookLabels).map(([code, name]) => <option key={code} value={code}>{tr(...name)}</option>)}</Combobox></Field>}
    <div className="sm:col-span-2 flex flex-wrap items-center gap-x-6 gap-y-2 pt-1">
      <Check label={tr("gl_enabled", "เปิดใช้งาน")} checked={value.isactive ?? true} onChange={(isactive) => set({ isactive })} />
    </div>
  </div>
    {resource === "mappings" && <div className="grid gap-2"><h3 className="font-semibold">{tr("gl_account_mapping_rules", "กฎการเชื่อมบัญชี")}</h3>{(value.rules ?? []).map((rule, index) => <div key={index} className="grid gap-2 rounded-xl border border-border p-2 sm:grid-cols-2 shadow-sm bg-muted/20"><AccountSelect label={tr("gl_accounts_in_rule", "บัญชีในกฎ {0}").replace("{0}", String(index + 1))} value={rule.accountcode || ""} accounts={accounts} onChange={(accountcode) => set({ rules: (value.rules ?? []).map((item, i) => i === index ? { ...item, accountcode } : item) })} /><Field label={tr("gl_accounting_side_rule", "ด้านบัญชีกฎ {0}").replace("{0}", String(index + 1))}><ChoiceSelect value={rule.side || "debit"} onChange={(side) => set({ rules: (value.rules ?? []).map((item, i) => i === index ? { ...item, side } : item) })}><option value="debit">{tr("gl_debit", "เดบิต")}</option><option value="credit">{tr("gl_credit", "เครดิต")}</option></ChoiceSelect></Field><Field label={tr("gl_amount_source_rule", "แหล่งจำนวนเงินกฎ {0}").replace("{0}", String(index + 1))}><input className={control} value={rule.source || ""} placeholder={tr("gl_select_supported_source_doc", "เลือกตามเอกสารต้นทางที่ระบบรองรับ")} onChange={(e) => set({ rules: (value.rules ?? []).map((item, i) => i === index ? { ...item, source: e.target.value } : item) })} /></Field><Button type="button" className={actionClass} variant="outline" onClick={() => set({ rules: (value.rules ?? []).filter((_, i) => i !== index) })}>{tr("gl_remove_rule", "นำกฎ {0} ออก").replace("{0}", String(index + 1))}</Button></div>)}<Button type="button" variant="outline" className={actionClass} onClick={() => set({ rules: [...(value.rules ?? []), { accountcode: "", side: "debit", source: "" }] })}>{tr("gl_add_account_mapping_rule", "เพิ่มกฎการเชื่อมบัญชี")}</Button></div>}
  </>;
}
