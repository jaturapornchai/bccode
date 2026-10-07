"use client";

import { useEffect, useMemo, useRef, useId, useState, type InputHTMLAttributes } from "react";
import { AlertTriangle, Eye, FileText, Pencil, Plus, RefreshCw, Save, Search, Trash2, X, FolderTree, List, ChevronDown, ChevronRight, Printer, Download, Lock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ChoiceSelect, Combobox } from "@/components/ui/select";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { JOURNAL_BOOK_CODE_MAX, JOURNAL_BOOK_NAME_MAX, JOURNAL_BOOK_TYPES, JOURNAL_DOC_FORMAT_DEFAULT, JOURNAL_DOC_FORMAT_MAX, accountName, accountTypeLabels, activeJournalBooks, chartOfAccountsCsv, emptyAccount, emptyFiscalYear, emptyJournalBook, emptyMaster, fillText, formatAmount, isJournalBookType, journalBookName, journalBookPayload, journalBookTypeLabels, untypedJournalBooks, validateJournalBook, type GLAccount, type GLFiscalYear, type GLJournalBook, type GLLabel, type GLMaster, type GLRecord, type GLResource, type GLTextFn } from "@/lib/general-ledger";
import { GLCommandError, commandFailure, glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Check, Field, Notice, Pager, SearchInput, SplitWorkbench, UnsavedBadge, YearSelect, actionClass, control, downloadText, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity, useGLLanguage, useGLText } from "./gl-common";
import { toast } from "@/lib/toast";
import { cn } from "@/lib/utils";
import { buildChartOfAccountsTree, filterAccountTree, type AccountTreeNode } from "@/lib/chart-of-accounts-tree";
import { GLChartOfAccountsPrint, GLJournalBooksPrint, printCompanyName, useGLPrint } from "./gl-print";
import { AccountBalanceCard, AccountLedgerModal } from "./account-ledger-drilldown";

type MasterResource = Exclude<GLResource, "journals">;
function newRecord(resource: MasterResource): GLRecord { return resource === "accounts" ? emptyAccount() : resource === "fiscal-years" ? emptyFiscalYear() : resource === "journal-books" ? emptyJournalBook() : emptyMaster(); }

/** บัญชีแม่ที่เลือกได้: บัญชีคุม (ไม่รับลงรายการ) ที่เปิดใช้งานและไม่ถูกลบ — ไม่รวมตัวเองและบัญชีลูกหลานทุกชั้น (กันผังวนเป็นวง) */
export function parentAccountCandidates(accounts: GLAccount[], accountcode: string): GLAccount[] {
  const self = accountcode.trim();
  const children = new Map<string, string[]>();
  for (const account of accounts) {
    if (!account.parentaccountcode) continue;
    children.set(account.parentaccountcode, [...(children.get(account.parentaccountcode) ?? []), account.accountcode]);
  }
  const descendants = new Set<string>();
  const queue = self ? [self] : [];
  while (queue.length > 0) {
    for (const child of children.get(queue.shift()!) ?? []) {
      if (!descendants.has(child)) { descendants.add(child); queue.push(child); }
    }
  }
  return accounts
    .filter((account) => account.isactive && !account.allowposting && !account.isdeleted && account.accountcode !== self && !descendants.has(account.accountcode))
    .sort((a, b) => a.accountcode.localeCompare(b.accountcode, "en", { numeric: true }));
}
/** ยอดคงเหลือปกติตามหมวด (สินทรัพย์/ค่าใช้จ่าย = เดบิต, หมวดอื่น = เครดิต) — บัญชีปรับลด เช่น ค่าเสื่อมราคาสะสม เลือกด้านตรงข้ามได้เอง */
export function defaultNormalBalance(accounttype: string): GLAccount["normalbalance"] {
  return accounttype === "asset" || accounttype === "expense" ? "debit" : "credit";
}
/** บัญชีเงินสดหรือรายการเทียบเท่าเงินสด (iscash): backend ยอมเฉพาะบัญชีสินทรัพย์ที่บันทึกบัญชีได้; บัญชีที่มีบัญชีย่อยเป็นบัญชีคุม */
export function accountCashAllowed(account: Pick<GLAccount, "accounttype" | "allowposting">, hasChildren: boolean): boolean {
  return account.accounttype === "asset" && account.allowposting !== false && !hasChildren;
}
/** เปลี่ยนหมวด: ยอดปกติกลับไปตามหมวดใหม่เฉพาะเมื่อหมวดเปลี่ยนจริง (ด้านที่ผู้ใช้เลือกไว้ไม่หาย); ออกจากหมวดสินทรัพย์ = ไม่ใช่บัญชีเงินสด */
export function accountTypePatch(account: Pick<GLAccount, "accounttype">, accounttype: GLAccount["accounttype"]): Partial<GLAccount> {
  if (accounttype === account.accounttype) return { accounttype };
  return { accounttype, normalbalance: defaultNormalBalance(accounttype), ...(accounttype === "asset" ? {} : { iscash: false }) };
}
/** เลือกบัญชีแม่: ระดับ = แม่ + 1 และหมวดตามแม่ (ยอดปกติเปลี่ยนเฉพาะเมื่อหมวดเปลี่ยน) */
export function accountParentPatch(account: Pick<GLAccount, "accounttype">, parentCode: string, parent: GLAccount | undefined): Partial<GLAccount> {
  if (!parentCode) return { parentaccountcode: null, level: 1 };
  if (!parent) return { parentaccountcode: parentCode };
  return { parentaccountcode: parent.accountcode, level: Math.min(12, (parent.level ?? 1) + 1), ...accountTypePatch(account, parent.accounttype) };
}
/** แนะนำรหัสบัญชีถัดไปใต้บัญชีแม่: หาเลขลำดับสูงสุดของลูกที่มีอยู่แล้ว หรือสร้างเลขรหัสลูกตัวแรก */
export function suggestNextAccountCode(parent: GLAccount, allAccounts: GLAccount[] = []): string {
  if (!parent?.accountcode) return "";
  const parentCode = parent.accountcode.trim();
  const children = allAccounts.filter((a) => a.parentaccountcode === parentCode);

  if (children.length > 0) {
    const numericChildren = children
      .map((c) => c.accountcode.trim())
      .filter((code) => /^\d+$/.test(code));

    if (numericChildren.length > 0) {
      const sampleLen = numericChildren[0].length;
      const maxNum = numericChildren.reduce((max, code) => {
        const val = parseInt(code, 10);
        return val > max ? val : max;
      }, 0);
      return String(maxNum + 1).padStart(sampleLen, "0");
    }

    const suffixed = children
      .map((c) => c.accountcode.trim())
      .map((code) => {
        const match = code.match(/^(.*?)([-/_])?(\d+)$/);
        return match ? { prefix: match[1] + (match[2] ?? ""), num: parseInt(match[3], 10), pad: match[3].length } : null;
      })
      .filter((x): x is NonNullable<typeof x> => x !== null);

    if (suffixed.length > 0) {
      const max = suffixed.reduce((prev, curr) => (curr.num > prev.num ? curr : prev), suffixed[0]);
      return `${max.prefix}${String(max.num + 1).padStart(max.pad, "0")}`;
    }
  }

  if (/^\d+$/.test(parentCode)) {
    const trailingZerosMatch = parentCode.match(/(0+)$/);
    if (trailingZerosMatch) {
      const zeros = trailingZerosMatch[1];
      if (zeros.length >= 2) {
        const power = zeros.length;
        if (power >= 3) {
          const base = parseInt(parentCode, 10);
          const step = Math.pow(10, power - 1);
          return String(base + step).padStart(parentCode.length, "0");
        }
        const base = parseInt(parentCode, 10);
        return String(base + 1).padStart(parentCode.length, "0");
      } else if (zeros.length === 1) {
        const base = parseInt(parentCode, 10);
        return String(base + 1).padStart(parentCode.length, "0");
      }
    }
    return `${parentCode}01`;
  }

  return `${parentCode}-01`;
}
/** ปิด "บันทึกบัญชีได้" = เป็นบัญชีคุม จึงไม่ใช่บัญชีเงินสด */
export function accountPostingPatch(allowposting: boolean): Partial<GLAccount> {
  return allowposting ? { allowposting } : { allowposting, iscash: false };
}
/** ค่าที่ส่งบันทึก: บัญชีที่มีบัญชีย่อยบันทึกบัญชีไม่ได้; iscash เหลือเฉพาะบัญชีที่ backend ยอม */
export function accountSavePayload(account: GLAccount, hasChildren: boolean): GLAccount {
  const acc = { ...account, parentaccountcode: account.parentaccountcode && account.parentaccountcode.trim() !== "" ? account.parentaccountcode.trim() : null };
  if (hasChildren) acc.allowposting = false;
  acc.iscash = Boolean(acc.iscash) && accountCashAllowed(acc, hasChildren);
  acc.departseparate = Boolean(account.departseparate);
  acc.projectseparate = Boolean(account.projectseparate);
  return acc;
}
export function normalizeRecord(resource: MasterResource, raw: GLRecord): GLRecord {
  const base = { ...newRecord(resource), ...raw };
  if (!base.id) {
    if (resource === "accounts" && "accountcode" in base && base.accountcode) {
      base.id = `acc-${base.accountcode}`;
    } else if ("code" in base && base.code) {
      base.id = `${resource === "fiscal-years" ? "year-" : resource === "journal-books" ? "book-" : ""}${base.code.toLowerCase()}`;
    }
  }
  if (resource === "accounts") {
    const acc = base as GLAccount;
    if (!Array.isArray(acc.names) || acc.names.length === 0) {
      acc.names = [{ code: "th", name: "" }];
    } else if (!acc.names.some((n) => n.code === "th")) {
      acc.names = [{ code: "th", name: "" }, ...acc.names];
    }
    acc.accountcode = acc.accountcode || "";
    acc.parentaccountcode = acc.parentaccountcode && acc.parentaccountcode.trim() !== "" ? acc.parentaccountcode : null;
    acc.accountgroup = acc.accountgroup || "";
    acc.accounttype = acc.accounttype || "asset";
    // ค่าว่างจาก emptyAccount() เป็นเดบิตเสมอ — ยอดปกติที่ไม่ได้ส่งมาจึงต้องตามหมวดของบัญชีนี้
    acc.normalbalance = (raw as Partial<GLAccount>).normalbalance || defaultNormalBalance(acc.accounttype);
    acc.level = typeof acc.level === "number" ? acc.level : 1;
    acc.isactive = acc.isactive ?? true;
    acc.allowposting = acc.allowposting ?? true;
    acc.iscash = acc.iscash ?? false;
    acc.departseparate = Boolean(acc.departseparate);
    acc.projectseparate = Boolean(acc.projectseparate);
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

export function recordId(record: GLRecord | null | undefined): string {
  if (!record) return "";
  if ("id" in record && record.id) return String(record.id);
  if ("accountcode" in record && record.accountcode) return String(record.accountcode);
  if ("code" in record && record.code) return String(record.code);
  return "";
}
function recordCode(record: GLRecord) { return "accountcode" in record && "names" in record ? record.accountcode : "code" in record ? record.code : ""; }
function recordName(record: GLRecord) { return "names" in record ? accountName(record) : "name" in record ? record.name : ""; }
/** Confirm-dialog subtitle: which record (code · name), plus the reason only when the user typed one — never the action label, the title already says it. */
export function confirmSubject(record: GLRecord, typedReason: string, tr: GLTextFn) {
  const name = recordName(record);
  const subject = name ? `${recordCode(record)} · ${name}` : String(recordCode(record));
  const reason = typedReason.trim();
  return reason ? fillText(tr("gl_confirm_subject_reason", "{0} — เหตุผล: {1}"), subject, reason) : subject;
}

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
  onAddChild,
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
  onAddChild?: (parentAcc: GLAccount) => void;
  tr: GLTextFn;
}) {
  const acc = node.account;
  const accKey = recordId(acc);
  const isSelected = Boolean(selectedId && (selectedId === accKey || selectedId === acc.id || selectedId === acc.accountcode));
  const isRowEditing = isSelected && isEditing;
  const isControl = !acc.allowposting;
  const isActive = acc.isactive ?? true;
  const isExpanded = expandedNodes[acc.accountcode] ?? (depth < 2);
  const toggleLabel = fillText(isExpanded ? tr("gl_collapse_sub_accounts", "ยุบบัญชีย่อยของ {0}") : tr("gl_expand_sub_accounts", "ขยายบัญชีย่อยของ {0}"), acc.accountcode);
  // Badges sit in fixed w-16 columns so rows stay aligned: short label + nowrap, full meaning in the tooltip.
  // nowrap needs `!`: the unlayered `span { text-wrap: pretty }` in globals.css resets text-wrap-mode and beats plain utilities.
  const badge = "inline-flex items-center justify-center whitespace-nowrap! rounded px-1.5 py-0.5 text-[0.65rem] leading-none h-5";

  return (
    <div className="flex flex-col shrink-0">
      <div
        className={`group flex items-center justify-between gap-2 px-2.5 py-1 min-h-[34px] cursor-pointer transition-colors text-[0.75rem] border-b border-border/40 shrink-0 ${
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
              className="p-0.5 text-muted-foreground hover:text-foreground rounded hover:bg-muted shrink-0 cursor-pointer inline-flex items-center justify-center"
              onClick={(e) => {
                e.stopPropagation();
                onToggleNode(acc.accountcode);
              }}
              aria-expanded={isExpanded}
              aria-label={toggleLabel}
              title={toggleLabel}
            >
              {isExpanded ? <ChevronDown className="size-4 block" /> : <ChevronRight className="size-4 block" />}
            </button>
          ) : (
            <span className="w-4 shrink-0 text-muted-foreground/40 text-center font-mono select-none text-xs">
              {depth > 0 ? "└" : ""}
            </span>
          )}

          <span className="font-mono font-semibold text-primary shrink-0 min-w-24">
            {acc.accountcode}
          </span>

          <span className="truncate text-foreground font-medium max-w-72" title={accountName(acc)}>
            {accountName(acc)}
          </span>
        </div>

        {/* Right anchored items: Level -> Account Role -> Status -> Actions (Fixed Width Columns) */}
        <div className="flex items-center justify-end shrink-0 ml-auto gap-2" onClick={(e) => e.stopPropagation()}>
          <div className="w-16 flex justify-end shrink-0">
            <span className={`${badge} font-semibold bg-primary/10 text-primary border border-primary/20`}>
              {tr("gl_level_2", "ระดับ {0}").replace("{0}", String(node.level))}
              <span className="sr-only">L{node.level}</span>
            </span>
          </div>

          <div className="w-16 flex justify-end shrink-0">
            {isControl ? (
              <span className={`${badge} font-medium bg-muted text-muted-foreground border border-border`} title={tr("gl_control_account", "บัญชีคุม")}>
                {tr("gl_control_account_short", "บัญชีคุม")}
              </span>
            ) : (
              <span className={`${badge} font-medium bg-primary/10 text-primary border border-primary/20`} title={tr("gl_posting_account", "บัญชีย่อย (บันทึกรายการ)")}>
                {tr("gl_posting_account_short", "บัญชีย่อย")}
              </span>
            )}
          </div>

          <div className="w-16 flex justify-end shrink-0">
            {isActive ? (
              <span className={`${badge} font-semibold bg-primary/10 text-primary border border-primary/20`}>
                {tr("gl_enable", "ใช้งาน")}
              </span>
            ) : (
              <span className={`${badge} font-semibold bg-muted text-muted-foreground border border-border`}>
                {tr("gl_disable", "ปิดใช้งาน")}
              </span>
            )}
          </div>

          <div className="w-20 flex items-center justify-end gap-1 shrink-0">
            {isControl && onAddChild && (
              <button
                type="button"
                className="p-1 rounded text-primary hover:bg-primary/20 transition-colors inline-flex items-center justify-center h-6 w-6"
                onClick={(e) => {
                  e.stopPropagation();
                  onAddChild(acc);
                }}
                title={tr("gl_add_sub_account_under", "เพิ่มบัญชีย่อยใต้ {0}").replace("{0}", acc.accountcode)}
                aria-label={tr("gl_add_sub_account", "เพิ่มบัญชีย่อย")}
              >
                <Plus className="size-3.5 block" />
              </button>
            )}
            <button
              type="button"
              className="p-1 rounded text-primary hover:bg-primary/20 transition-colors inline-flex items-center justify-center h-6 w-6"
              onClick={(e) => {
                e.stopPropagation();
                onEdit(acc);
              }}
              title={tr("gl_edit_label", "แก้ไข (Edit)")}
              aria-label={tr("gl_edit", "แก้ไข")}
            >
              <Pencil className="size-3.5 block" />
            </button>
            <button
              type="button"
              className="p-1 rounded text-destructive hover:bg-destructive/20 transition-colors inline-flex items-center justify-center h-6 w-6"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(acc);
              }}
              title={tr("gl_delete", "ลบ")}
              aria-label={tr("gl_delete", "ลบ")}
            >
              <Trash2 className="size-3.5 block" />
            </button>
          </div>
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
          onAddChild={onAddChild}
          tr={tr}
        />
      ))}
    </div>
  );
}

export function GLMasters({ resource, route }: { resource: MasterResource; route: string }) {
  const tr = useGLText();
  const language = useGLLanguage();
  const { print: printJob, portal: printPortal } = useGLPrint();
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
  const [error, setError] = useState(""), [reason, setReason] = useState("");
  const [errorField, setErrorField] = useState("");
  const [ledgerModalAcc, setLedgerModalAcc] = useState<GLAccount | null>(null);
  const [ledgerFiscalYear, setLedgerFiscalYear] = useState<string>("");
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
      const id = recordId(item);
      const value = id ? await glRequest<GLRecord>(`${resource}/${encodeURIComponent(id)}`) : newRecord(resource);
      const normalized = normalizeRecord(resource, value);
      setRecord(normalized);
      setOriginal(JSON.stringify(normalized));
      setIsEditing(false);
      setError("");
      setReason("");
    } catch (e) {
      showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    }
  }

  async function openEdit(item: GLRecord) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const id = recordId(item);
      const value = id ? await glRequest<GLRecord>(`${resource}/${encodeURIComponent(id)}`) : newRecord(resource);
      const normalized = normalizeRecord(resource, value);
      setRecord(normalized);
      setOriginal(JSON.stringify(normalized));
      setIsEditing(true);
      setError("");
      setReason("");
    } catch (e) {
      showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    }
  }

  async function openCreate(preset?: Partial<GLAccount>) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    const fresh = normalizeRecord(resource, { ...newRecord(resource), ...(preset ?? {}) });
    delete fresh.id;
    setRecord(fresh);
    setOriginal(JSON.stringify(fresh));
    setIsEditing(true);
    setError("");
    setReason("");
  }

  function handleAddChild(parentAcc: GLAccount) {
    const childLevel = Math.min(12, (parentAcc.level ?? 1) + 1);
    const childType = parentAcc.accounttype;
    const suggestedCode = suggestNextAccountCode(parentAcc, treeAccounts);
    void openCreate({
      accountcode: suggestedCode,
      parentaccountcode: parentAcc.accountcode,
      level: childLevel,
      accounttype: childType,
      normalbalance: parentAcc.normalbalance || defaultNormalBalance(childType),
      allowposting: true,
      isactive: true,
      iscash: false,
      departseparate: Boolean(parentAcc.departseparate),
      projectseparate: Boolean(parentAcc.projectseparate),
    });
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
    if (resource === "journal-books") {
      // A legacy book stored without a type can be renamed or deactivated as it is (backend allowUntyped).
      const storedBookType = original ? Number((JSON.parse(original) as GLJournalBook).booktype) || 0 : 0;
      const bookProblem = validateJournalBook(record as GLJournalBook, tr, Boolean(record.id) && storedBookType === 0);
      if (bookProblem) { showFormError(null, bookProblem.message, bookProblem.field); return; }
    }
    if (!recordCode(record).trim() || resource !== "fiscal-years" && !recordName(record).trim()) { showFormError(null, tr("gl_enter_code_and_name", "กรุณาระบุรหัสและชื่อให้ครบ"), recordCode(record).trim() ? (isAcc ? "accountnameth" : "name") : (isAcc ? "accountcode" : "code")); return; }
    if (record.id && !await confirm({ title: tr("gl_save_changes_confirm", "บันทึกการแก้ไขข้อมูล?"), description: tr("gl_edit_with_history", "แก้ไข {0} โดยเก็บประวัติการเปลี่ยนแปลง").replace("{0}", String(recordCode(record))), confirmLabel: tr("gl_save_changes", "บันทึกการแก้ไข"), tone: "info" })) return;
    try {
      setError("");
      let accPayload: GLAccount | undefined;
      if (resource === "accounts") {
        const acc = record as GLAccount;
        accPayload = accountSavePayload(acc, refs.accounts.some((a) => a.parentaccountcode === acc.accountcode));
      }
      const field = resource === "accounts" ? { account: accPayload! } : resource === "fiscal-years" ? { fiscalyear: record as GLFiscalYear } : resource === "journal-books" ? { master: journalBookPayload(record as GLJournalBook) } : { master: record as GLMaster };
      const result = await execute({ resource, action: record.id ? "update" : "create", id: record.id, version: record.version, reason: reason.trim() || (record.id ? tr("gl_edit_data", "แก้ไขข้อมูล") : tr("gl_create_new", "สร้างข้อมูลใหม่")), ...field });
      const saved = normalizeRecord(resource, { ...record, id: result.id, version: result.version });
      setRecord(saved);
      setOriginal(JSON.stringify(saved));
      setIsEditing(false);
      setReason("");
      list.reload();
      setRevision((value) => value + 1);
      if (result.projectionpending) toast.info(tr("gl_saved_updating_reload", "บันทึกแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน กดโหลดใหม่เพื่อตรวจสอบ"));
      else toast.success(tr("gl_saved_successfully", "บันทึกเรียบร้อยแล้ว"));
    } catch (e) { showFormError(e, tr("gl_save_failed_retry", "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"), saveFailureTarget(resource)); }
  }

  async function deleteItem(item: GLRecord) {
    const id = recordId(item);
    if (!id || busy) return;
    const isAcc = resource === "accounts";
    if (!await confirm({
      title: tr("gl_confirm_delete", "ยืนยันลบรายการ?"),
      description: confirmSubject(item, "", tr),
      details: isAcc ? tr("gl_coa_referenced_no_delete", "บัญชีที่มีรายการในสมุดรายวัน หรือมีข้อมูลอื่นอ้างอิงอยู่ จะลบไม่ได้ — ถ้าเลิกใช้ ให้ปิดใช้งานแทน") : resource === "journal-books" ? tr("gl_journal_book_delete_hint", "สมุดที่มีใบสำคัญใช้อยู่ลบไม่ได้ — ถ้าเลิกใช้ ให้แก้เป็นปิดใช้งานแทน") : undefined,
      confirmLabel: tr("gl_delete_item", "ลบรายการ"),
      tone: "danger",
    })) return;
    try {
      setError("");
      await execute({ resource, id, version: item.version, action: "delete", reason: tr("gl_delete_item", "ลบรายการ") });
      if (record && recordId(record) === id) {
        setRecord(null);
        setOriginal("");
        setIsEditing(false);
      }
      list.reload();
      setRevision((value) => value + 1);
      toast.success(tr("gl_delete_success", "ลบ {0} เรียบร้อยแล้ว").replace("{0}", String(recordCode(item))));
    } catch (e) {
      showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"));
    }
  }

  // ล็อก/ปลดล็อกงวดย้ายไปจอกำหนดงวดบัญชี (gl-periods.tsx) — จอนี้เหลือแค่ลบ
  async function runAction(action: "delete") {
    if (!record || busy) return;
    const id = recordId(record);
    if (!id) return;
    const label = tr("gl_delete_item", "ลบรายการ");
    const effectiveReason = reason.trim() || label;
    if (!await confirm({
      title: `${label}?`,
      // Name the record as saved — the form may hold unsaved edits to the code/name.
      description: confirmSubject(original ? JSON.parse(original) as GLRecord : record, reason, tr),
      details: resource === "accounts" ? tr("gl_coa_referenced_no_delete", "บัญชีที่มีรายการในสมุดรายวัน หรือมีข้อมูลอื่นอ้างอิงอยู่ จะลบไม่ได้ — ถ้าเลิกใช้ ให้ปิดใช้งานแทน") : resource === "journal-books" ? tr("gl_journal_book_delete_hint", "สมุดที่มีใบสำคัญใช้อยู่ลบไม่ได้ — ถ้าเลิกใช้ ให้แก้เป็นปิดใช้งานแทน") : undefined,
      confirmLabel: label,
      tone: "danger"
    })) return;
    try {
      setError("");
      await execute({ resource, id, version: record.version, action, reason: effectiveReason });
      setRecord(null); setOriginal(""); setIsEditing(false); list.reload(); setRevision((value) => value + 1); toast.success(tr("gl_success_message", "{0}เรียบร้อยแล้ว").replace("{0}", String(label)));
    } catch (e) { showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")); }
  }

  const hasAmount = resource === "forecast";
  const isAcc = resource === "accounts";
  const isBook = resource === "journal-books";
  const isProductGroup = resource === "product-account-groups";
  const champLabels = champMasterLabels(resource, tr);
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

  const handlePrint = () => {
    if (!isAcc || !treeAccounts.length) return;
    printJob(
      <GLChartOfAccountsPrint
        accounts={treeAccounts}
        company={printCompanyName()}
        tr={tr}
        language={language}
      />,
      "portrait"
    );
  };

  // Champ 5531 รายงานรายละเอียดสมุดบัญชี = ปุ่มพิมพ์ในจอกำหนดสมุดรายวัน (ไม่แยกเมนูรายงาน)
  const handlePrintBooks = () => {
    if (!isBook || !refs.books.length) return;
    printJob(<GLJournalBooksPrint books={refs.books} company={printCompanyName()} tr={tr} language={language} />, "portrait");
  };

  const handleExportCsv = () => {
    if (!isAcc || !treeAccounts.length) return;
    const csv = chartOfAccountsCsv(treeAccounts, tr);
    downloadText(
      tr("gl_coa_export_filename", "ผังบัญชี-chart-of-accounts.csv"),
      csv,
      "text/csv;charset=utf-8"
    );
  };

  return <div className="flex flex-col flex-1 min-h-0 gap-2">
    {printPortal}
    <div className="shrink-0 flex flex-col gap-2"><Notice error text={alerts.page} /></div>
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
            {isBook && (
              <Button type="button" variant="outline" className={actionClass} onClick={handlePrintBooks} disabled={!refs.books.length} title={tr("gl_book_report_title", "รายงานรายละเอียดสมุดบัญชี")}>
                <Printer className="size-4 mr-1.5" />
                {tr("gl_print", "พิมพ์")}
              </Button>
            )}
            {isAcc && (
              <>
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
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  onClick={handlePrint}
                  title={tr("gl_print_coa", "พิมพ์ผังบัญชี")}
                >
                  <Printer className="size-4 mr-1.5" />
                  {tr("gl_print", "พิมพ์")}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  onClick={handleExportCsv}
                  title={tr("gl_export_csv", "ส่งออก CSV (Excel)")}
                >
                  <Download className="size-4 mr-1.5" />
                  {tr("gl_export_csv_short", "ส่งออก Excel")}
                </Button>
              </>
            )}
          </div>
        </form>
        {resource === "journal-books" && <UntypedJournalBooksNotice books={refs.books} onOpen={(book) => void openView(book)} />}
        {/* .bc-list-toolbar */}
        <div className="bc-list-toolbar shrink-0 p-2.5 border-b border-border/60 bg-muted/20 flex flex-wrap items-center justify-between gap-2 rounded-t-lg">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold text-foreground">
              {tr("gl_all_records", "รายการทั้งหมด")}:{" "}
              <strong className="text-primary font-bold">
                {isAcc && viewMode === "tree"
                  ? treeAccounts.length.toLocaleString("th-TH")
                  : list.data.total.toLocaleString("th-TH")}
              </strong>
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
        <div className="flex-1 min-h-[300px] overflow-y-auto rounded-xl border border-border shadow-xs p-2 flex flex-col gap-2.5 bg-card scrollbar-thin" aria-busy={list.loading}>
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
                        selectedId={recordId(record)}
                        isEditing={isEditing}
                        expandedNodes={expandedNodes}
                        onToggleNode={toggleNode}
                        onSelect={(acc) => void openView(acc)}
                        onEdit={(acc) => void openEdit(acc)}
                        onDelete={(acc) => void deleteItem(acc)}
                        onAddChild={handleAddChild}
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
        <div className="flex-1 min-h-[300px] overflow-y-auto scrollbar-thin rounded-xl border border-border shadow-xs" aria-busy={list.loading}>
          <table className="w-full text-left border-collapse">
            {/* .bc-list-header */}
            <thead className="bc-list-header sticky top-0 bg-muted/95 backdrop-blur z-10 border-b border-border/80 text-[0.7rem] font-extrabold uppercase tracking-wider text-muted-foreground">
              <tr>
                <th className="py-2 px-2.5 text-left w-28 whitespace-nowrap">{isBook ? tr("gl_book_code", "รหัสสมุดรายวัน") : champLabels?.code ?? tr("gl_code", "รหัส")}</th>
                <th className="py-2 px-2 text-left min-w-44">{isBook ? tr("gl_book_name", "ชื่อสมุดรายวัน") : champLabels?.name ?? tr("gl_name_description", "ชื่อ / รายละเอียด")}</th>
                {isBook && <th className="py-2 px-2 text-left min-w-36">{tr("gl_book_title_name", "หัวเอกสาร")}</th>}
                {isProductGroup && PRODUCT_GROUP_ACCOUNTS.map(({ key, label }) => <th key={key} className="py-2 px-2 text-left w-36 whitespace-nowrap">{tr(...label)}</th>)}
                {hasAmount && <th className="py-2 px-2 text-right w-36 whitespace-nowrap">{tr("gl_amount", "จำนวนเงิน")}</th>}
                {isAcc && <th className="py-2 px-2 text-right w-20 whitespace-nowrap">{tr("gl_level", "ระดับ")}</th>}
                <th className="py-2 px-2 text-right w-24 whitespace-nowrap">{tr("gl_status", "สถานะ")}</th>
                <th className="py-2 pr-2.5 pl-2 text-right w-20 whitespace-nowrap">{tr("gl_manage", "จัดการ")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/40">
              {list.data.items.map((item, index) => {
                const accLevel = isAcc && "level" in item && typeof item.level === "number" ? item.level : 1;
                const selectedKey = recordId(record);
                const itemKey = recordId(item);
                const isSelected = Boolean(selectedKey && selectedKey === itemKey);
                const isRowEditing = isSelected && isEditing;
                const isLocked = "locked" in item && item.locked;
                const isClosed = "closed" in item && item.closed;
                const isActive = !("isactive" in item) || item.isactive;
                return (
                  <tr
                    key={itemKey || index}
                    onClick={() => void openView(item)}
                    className={`bc-list-row cursor-pointer transition-colors text-[0.75rem] align-middle ${
                      isRowEditing
                        ? "bg-primary/15 font-medium text-foreground ring-1 ring-inset ring-primary/50"
                        : isSelected
                          ? "bg-primary/10 font-medium text-foreground ring-1 ring-inset ring-primary/40"
                          : index % 2 === 0
                            ? "bg-background hover:bg-muted/40 text-foreground/90"
                            : "bg-muted/20 hover:bg-muted/40 text-foreground/90"
                    }`}
                    style={{ padding: "3px 8px" }}
                  >
                    <td className="py-2 px-2.5 whitespace-nowrap align-middle">
                      <span className="font-mono font-semibold text-primary">{recordCode(item)}</span>
                    </td>
                    <td className="py-2 px-2 truncate align-middle" title={recordName(item)}>
                      {isAcc && accLevel > 1 ? (
                        <span style={{ paddingLeft: `${(accLevel - 1) * 16}px` }} className="inline-flex items-center gap-1.5 align-middle">
                          <span className="text-muted-foreground select-none font-mono">└─</span>
                          <span>{recordName(item)}</span>
                        </span>
                      ) : (
                        <span className="align-middle">{recordName(item)}{resource === "journal-books" && <JournalBookTypeBadge booktype={(item as GLJournalBook).booktype} />}</span>
                      )}
                    </td>
                    {isBook && <td className="py-2 px-2 truncate align-middle">{(item as GLJournalBook).titlename || ""}</td>}
                    {isProductGroup && PRODUCT_GROUP_ACCOUNTS.map(({ key }) => <td key={key} className="py-2 px-2 whitespace-nowrap align-middle font-mono">{(item as GLMaster)[key] || ""}</td>)}
                    {hasAmount && (
                      <td className="py-2 px-2 text-right font-mono font-medium tabular-nums whitespace-nowrap align-middle">
                        {"amount" in item ? formatAmount(item.amount) : "-"}
                      </td>
                    )}
                    {isAcc && (
                      <td className="whitespace-nowrap py-2 px-2 text-right align-middle">
                        <span className="inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[0.65rem] font-semibold bg-primary/10 text-primary border border-primary/20 leading-none h-5 align-middle">
                          {tr("gl_level_2", "ระดับ {0}").replace("{0}", String(accLevel))}
                        </span>
                      </td>
                    )}
                    <td className="whitespace-nowrap py-2 px-2 text-right align-middle">
                      {isLocked ? (
                        <span className="inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[0.65rem] font-semibold bg-muted text-muted-foreground border border-border leading-none h-5 align-middle">
                          {tr("gl_locked", "ล็อกแล้ว")}
                        </span>
                      ) : isClosed ? (
                        <span className="inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[0.65rem] font-semibold bg-muted text-muted-foreground border border-border leading-none h-5 align-middle">
                          {tr("gl_year_closed", "ปิดปีแล้ว")}
                        </span>
                      ) : !isActive ? (
                        <span className="inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[0.65rem] font-semibold bg-muted text-muted-foreground border border-border leading-none h-5 align-middle">
                          {tr("gl_disable", "ปิดใช้งาน")}
                        </span>
                      ) : (
                        <span className="inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[0.65rem] font-semibold bg-primary/10 text-primary border border-primary/20 leading-none h-5 align-middle">
                          {tr("gl_enable", "ใช้งาน")}
                        </span>
                      )}
                    </td>
                    <td className="py-2 pr-2.5 pl-2 text-right whitespace-nowrap align-middle" onClick={(e) => e.stopPropagation()}>
                      <div className="inline-flex items-center justify-end gap-1 align-middle">
                        <button
                          type="button"
                          className="p-1 rounded text-primary hover:bg-primary/20 transition-colors inline-flex items-center justify-center h-6 w-6"
                          onClick={(e) => {
                            e.stopPropagation();
                            void openEdit(item);
                          }}
                          aria-label={tr("gl_edit", "แก้ไข")}
                          title={tr("gl_edit_label", "แก้ไข (Edit)")}
                        >
                          <Pencil className="size-3.5 block" />
                        </button>
                        <button
                          type="button"
                          className="p-1 rounded text-destructive hover:bg-destructive/20 transition-colors inline-flex items-center justify-center h-6 w-6"
                          onClick={(e) => {
                            e.stopPropagation();
                            void deleteItem(item);
                          }}
                          aria-label={tr("gl_delete", "ลบ")}
                          title={tr("gl_delete_label", "ลบ (Delete)")}
                        >
                          <Trash2 className="size-3.5 block" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
              {!list.data.items.length && (
                <tr>
                  <td colSpan={colSpan} className="py-10 text-center text-muted-foreground text-xs">
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
        <div className="flex flex-col h-full min-h-0 gap-2">
          <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-2 shrink-0">
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
          {resource !== "accounts" && (
            <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-2.5 py-1 text-xs text-muted-foreground shrink-0">
              <Eye className="size-3.5 text-primary shrink-0" />
              <span>{tr("gl_readonly_mode_edit_hint", "โหมดแสดงข้อมูล (Read-only) — หากต้องการแก้ไข ให้กดปุ่ม")} <strong>{tr("gl_edit", "แก้ไข")}</strong></span>
            </div>
          )}
          {alerts.pane && <FormErrorAlert text={alerts.pane} />}
          {resource === "accounts" && (
            <AccountBalanceCard
              account={record as GLAccount}
              years={refs.years}
              books={refs.books ?? []}
              onOpenLedger={(year) => {
                setLedgerModalAcc(record as GLAccount);
                setLedgerFiscalYear(year);
              }}
            />
          )}
          <fieldset disabled data-mode="view" className="grid gap-2 opacity-95 content-start overflow-y-auto min-h-0 flex-1 p-0.5">
            {resource === "accounts" ? (
              <AccountFields value={record as GLAccount} set={set} accounts={refs.accounts} />
            ) : resource === "fiscal-years" ? (
              <FiscalYearFields value={record as GLFiscalYear} set={set} accounts={refs.accounts} />
            ) : resource === "journal-books" ? (
              <JournalBookFields value={record as GLJournalBook} set={set} />
            ) : (
              <MasterFields resource={resource} value={record as GLMaster} set={set} accounts={refs.accounts} years={refs.years} books={refs.books ?? []} />
            )}
          </fieldset>
          <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-2 shrink-0 mt-auto">
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
        <form ref={formRef} className="flex flex-col h-full min-h-0 gap-2" onSubmit={(event) => { event.preventDefault(); void save(); }}>
          <header className="flex items-center justify-between gap-2 border-b border-border pb-2 shrink-0">
            <div className="flex min-w-0 items-center gap-2">
              <div className="grid size-8 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                {record.id ? <Pencil className="size-4" /> : <Plus className="size-4" />}
              </div>
              <h2 className="truncate text-base font-semibold">
                {record.id
                  ? tr("gl_edit_2", "แก้ไข {0}").replace("{0}", String(recordCode(record)))
                  : isAcc && (record as GLAccount).parentaccountcode
                  ? tr("gl_add_sub_account_under", "เพิ่มบัญชีย่อยใต้ {0}").replace("{0}", (record as GLAccount).parentaccountcode!)
                  : tr("gl_add_new_item", "เพิ่มรายการใหม่")}
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
          <fieldset disabled={busy} className="grid gap-2 opacity-95 content-start overflow-y-auto min-h-0 flex-1 p-0.5">
            {resource === "accounts" ? (
              <AccountFields value={record as GLAccount} set={set} accounts={refs.accounts} />
            ) : resource === "fiscal-years" ? (
              <FiscalYearFields value={record as GLFiscalYear} set={set} accounts={refs.accounts} />
            ) : resource === "journal-books" ? (
              <JournalBookFields value={record as GLJournalBook} set={set} />
            ) : (
              <MasterFields resource={resource} value={record as GLMaster} set={set} accounts={refs.accounts} years={refs.years} books={refs.books ?? []} />
            )}
            {record.id && (
              <Field label={tr("gl_change_reason_optional", "เหตุผลการเปลี่ยนแปลง (ไม่บังคับ)")}>
                <input className={control} value={reason} onChange={(event) => setReason(event.target.value)} placeholder={tr("gl_edit_reason_hint", "ระบุเหตุผลการแก้ไข (ถ้ามี)")} />
              </Field>
            )}
          </fieldset>
          <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-2 shrink-0 mt-auto">
            <div className="flex flex-wrap gap-2">
              <Button type="submit" className={actionClass} disabled={busy || !dirty}>
                <Save className="size-4 mr-1.5" />{busy ? tr("gl_saving", "กำลังบันทึก…") : tr("gl_save_data", "บันทึกข้อมูล")}
              </Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()}>
                {tr("gl_cancel", "ยกเลิก")}
              </Button>
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
    {ledgerModalAcc && (
      <AccountLedgerModal
        open={Boolean(ledgerModalAcc)}
        onClose={() => setLedgerModalAcc(null)}
        account={ledgerModalAcc}
        fiscalYear={ledgerFiscalYear}
        books={refs.books ?? []}
      />
    )}
    {confirmationDialog}
  </div>;
}

/** A saved record code cannot change. The shared `control` keeps disabled inputs looking normal (read-only view mode),
 *  so a code locked inside an editable form needs its own cue: muted fill, lock icon, not-allowed cursor and a visible hint under the field.
 *  The fill sits on the wrapper because globals.css forces every input background to transparent (!important). */
function CodeInput({ locked, ...props }: InputHTMLAttributes<HTMLInputElement> & { locked: boolean }) {
  const tr = useGLText();
  const hintId = useId();
  if (!locked) return <input className={control} {...props} />;
  const hint = tr("gl_code_locked_hint", "รหัสแก้ไขไม่ได้หลังบันทึกแล้ว");
  return (
    <div className="grid gap-1">
      <div className="relative rounded-xl bg-muted/60">
        {/* pr-9 needs `!`: the shared control sets its own padding and globals.css input rules win over plain utilities. */}
        <input className={cn(control, "pr-9! disabled:cursor-not-allowed")} {...props} disabled aria-describedby={hintId} />
        <Lock aria-hidden className="pointer-events-none absolute right-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      </div>
      {/* View mode disables every field, so the hint only matters while editing. */}
      <span id={hintId} className="text-[0.9rem] text-muted-foreground in-data-[mode=view]:sr-only">{hint}</span>
    </div>
  );
}
function AccountFields({ value, set, accounts = [] }: { value: GLAccount; set: (patch: object) => void; accounts?: GLAccount[] }) {
  const tr = useGLText();
  const thName = (value.names || []).find((name) => name.code === "th")?.name ?? "";
  const enName = (value.names || []).find((name) => name.code === "en")?.name ?? "";

  const handleTypeChange = (accounttype: string) => set(accountTypePatch(value, accounttype as GLAccount["accounttype"]));

  const hasChildren = Boolean(
    value.accountcode && accounts.some((a) => a.parentaccountcode === value.accountcode)
  );

  const parentCandidates = useMemo(() => parentAccountCandidates(accounts, value.accountcode || ""), [accounts, value.accountcode]);
  // บัญชีแม่เดิมที่เลือกไม่ได้แล้ว (ปิดใช้งาน/รับลงรายการ/เป็นลูกของตัวเอง) ยังแสดงให้เห็น พร้อมบอกให้เลือกใหม่
  const invalidParent = value.parentaccountcode && !parentCandidates.some((a) => a.accountcode === value.parentaccountcode) ? value.parentaccountcode : "";

  const handleParentChange = (parentCode: string) => {
    const parent = accounts.find((a) => a.accountcode === parentCode);
    const patch = accountParentPatch(value, parentCode, parent);
    if (!value.id && parent && (!value.accountcode || accounts.some((a) => a.accountcode === value.accountcode))) {
      const suggested = suggestNextAccountCode(parent, accounts);
      if (suggested) {
        set({ ...patch, accountcode: suggested });
        return;
      }
    }
    set(patch);
  };
  const categoryBalance = defaultNormalBalance(value.accounttype || "asset");
  const normalBalance = value.normalbalance || categoryBalance;
  const balanceLabel = (side: GLAccount["normalbalance"]) => (side === "debit" ? `${tr("gl_debit", "เดบิต")} (Dr.)` : `${tr("gl_credit", "เครดิต")} (Cr.)`);
  const cashAllowed = accountCashAllowed(value, hasChildren);

  return (
    <div className="flex flex-col gap-2">
      <div className="grid items-start gap-2 sm:grid-cols-2">
        <Field label={tr("gl_account_code", "รหัสบัญชี")}>
          <CodeInput
            locked={!!value.id}
            data-field="accountcode"
            required
            value={value.accountcode || ""}
            onChange={(e) => set({ accountcode: e.target.value })}
            maxLength={20}
          />
        </Field>
        {/* Champ: ช่อง "เป็นบัญชีคุม" อยู่ข้างรหัสบัญชี (IsHeader = ลงรายการไม่ได้) — BC เก็บเป็น allowposting กลับด้าน */}
        <div className="grid min-w-0 content-start gap-1 text-[0.95rem] leading-normal sm:pt-7">
          <Check
            label={tr("gl_is_control_account", "เป็นบัญชีคุม")}
            checked={hasChildren || value.allowposting === false}
            disabled={hasChildren}
            onChange={(control) => {
              if (!hasChildren) set(accountPostingPatch(!control));
            }}
          />
          <span className="text-[0.9rem] text-muted-foreground">{hasChildren ? tr("gl_parent_has_children_no_posting", "ผังบัญชีที่เป็นหัว (มีตัวลูก) ไม่สามารถบันทึกตัวเลขได้") : tr("gl_is_control_account_hint", "บัญชีคุมใช้รวมยอดบัญชีย่อย ลงรายการในสมุดรายวันไม่ได้")}</span>
        </div>
        <div className="sm:col-span-2">
          <Field label={tr("gl_account_name_th", "ชื่อบัญชี 1")}>
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
        </div>
        <div className="sm:col-span-2">
          <Field label={tr("gl_account_name_en", "ชื่อบัญชี 2")}>
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
        </div>
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
        <Field label={tr("gl_account_level", "ระดับบัญชี")}>
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
        <Field label={tr("gl_parent_account", "หัวบัญชี (บัญชีแม่)")}>
          <Combobox
            data-field="parentaccountcode"
            aria-label={tr("gl_parent_account", "หัวบัญชี (บัญชีแม่)")}
            value={value.parentaccountcode || ""}
            onChange={handleParentChange}
          >
            <option value="">{tr("gl_no_parent_level_1", "-- ไม่มีหัวบัญชี (ระดับ 1 ผังหลัก) --")}</option>
            {invalidParent && <option value={invalidParent}>{invalidParent} {tr("gl_parent_not_allowed_suffix", "(ใช้เป็นบัญชีแม่ไม่ได้ — ต้องเป็นบัญชีคุมที่เปิดใช้งาน กรุณาเลือกใหม่)")}</option>}
            {parentCandidates.map((parent) => {
              const lvl = parent.level ?? 1;
              const indent = lvl > 1 ? `${"\u00A0\u00A0".repeat(lvl - 1)}└─ ` : "";
              return (
                <option key={parent.accountcode} value={parent.accountcode}>
                  {indent}{parent.accountcode} · {accountName(parent)} ({tr("gl_level_2", "ระดับ {0}").replace("{0}", String(lvl))})
                </option>
              );
            })}
          </Combobox>
        </Field>
        <Field label={tr("gl_normal_balance", "ยอดคงเหลือปกติ")}>
          <div title={fillText(tr("gl_normal_balance_default_hint", "ปกติของหมวดนี้: {0} — เปลี่ยนไม่ได้เมื่อบัญชีผ่านรายการแล้ว"), balanceLabel(categoryBalance))}>
            <Combobox
              data-field="normalbalance"
              aria-label={tr("gl_normal_balance", "ยอดคงเหลือปกติ")}
              value={normalBalance}
              onChange={(side) => set({ normalbalance: side === "credit" ? "credit" : "debit" })}
            >
              <option value="debit">{balanceLabel("debit")}</option>
              <option value="credit">{balanceLabel("credit")}</option>
            </Combobox>
          </div>
          {normalBalance !== categoryBalance && (
            <span role="status" className="flex items-center gap-1 text-xs font-medium text-primary mt-0.5">
              <AlertTriangle aria-hidden className="size-3.5 shrink-0" />
              {tr("gl_normal_balance_contra", "บัญชีปรับลด (ด้านตรงข้ามกับหมวด)")}
            </span>
          )}
        </Field>
        <div className="sm:col-span-2 flex flex-col gap-1.5 p-2 rounded-xl border border-border/60 bg-muted/20">
          <div className="flex flex-wrap items-center gap-2">
            <Check
              className="px-2.5 py-1 text-xs rounded-lg gap-1.5"
              label={tr("gl_enabled", "เปิดใช้งาน")}
              checked={value.isactive ?? true}
              onChange={(isactive) => set({ isactive })}
            />
            <div className="inline-flex items-center" title={tr("gl_iscash_hint", "เลือกได้เฉพาะบัญชีสินทรัพย์ที่บันทึกบัญชีได้และไม่มีบัญชีย่อย — เปลี่ยนไม่ได้เมื่อบัญชีผ่านรายการแล้ว")}>
              <Check
                className="px-2.5 py-1 text-xs rounded-lg gap-1.5"
                label={tr("gl_iscash_account", "บัญชีเงินสด / เทียบเท่าเงินสด")}
                checked={cashAllowed && Boolean(value.iscash)}
                disabled={!cashAllowed}
                onChange={(iscash) => set({ iscash })}
              />
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 pt-1 border-t border-border/40" title={tr("gl_dim_separate_hint", "บรรทัดรายวันของบัญชีนี้ระบุแผนก/โครงการได้เฉพาะเงื่อนไขที่เลือก")}>
            <span className="text-xs text-muted-foreground font-medium shrink-0">
              {tr("gl_dim_separate", "เงื่อนไข:")}
            </span>
            <Check
              className="px-2.5 py-1 text-xs rounded-lg gap-1.5"
              label={tr("gl_depart_separate", "แยกแผนก")}
              checked={Boolean(value.departseparate)}
              onChange={(departseparate) => set({ departseparate })}
            />
            <Check
              className="px-2.5 py-1 text-xs rounded-lg gap-1.5"
              label={tr("gl_project_separate", "แยกโครงการ")}
              checked={Boolean(value.projectseparate)}
              onChange={(projectseparate) => set({ projectseparate })}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
function FiscalYearFields({ value, set, accounts }: { value: GLFiscalYear; set: (patch: object) => void; accounts: GLAccount[] }) {
  const tr = useGLText();
  return <><Notice text={tr("gl_set_fy_currency_before_journal", "กำหนดปีบัญชีก่อนบันทึกรายวัน ระบบตรวจจำนวนทศนิยมตามปีบัญชีและไม่ปัดยอดให้อัตโนมัติ")} /><div className="grid gap-3 sm:grid-cols-2">
    <Field label={tr("gl_fiscal_year_code", "รหัสปีบัญชี")}><CodeInput locked={!!value.id} required value={value.code || ""} onChange={(e) => set({ code: e.target.value })} /></Field>
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
/** Champ columns of กำหนดกลุ่มบัญชีสินค้า (key = field in the record, label = [language key, Champ wording]). */
const PRODUCT_GROUP_ACCOUNTS: readonly { key: "itemaccount" | "costaccount" | "revenueaccount"; label: GLLabel }[] = [
  { key: "itemaccount", label: ["gl_inventory_account", "รหัสผังบัญชีสินค้า"] },
  { key: "costaccount", label: ["gl_cogs_account", "รหัสผังบัญชีต้นทุน"] },
  { key: "revenueaccount", label: ["gl_sales_revenue_account", "รหัสผังบัญชีรายได้"] },
];
/** Code/name wording of the Champ group and link-template screens; the shared gl_code/gl_name keys stay for the other masters. */
function champMasterLabels(resource: MasterResource, tr: GLTextFn) {
  if (resource === "account-groups" || resource === "product-account-groups") return { code: tr("gl_group_code", "รหัสกลุ่ม"), name: tr("gl_group_name", "ชื่อกลุ่ม") };
  if (resource === "mappings") return { code: tr("gl_mapping_code", "รหัสรูปแบบ"), name: tr("gl_mapping_name", "ชื่อรูปแบบ") };
  return null;
}
function MasterFields({ resource, value, set, accounts, years, books }: { resource: MasterResource; value: GLMaster; set: (patch: object) => void; accounts: GLAccount[]; years: GLFiscalYear[]; books: GLJournalBook[] }) {
  const tr = useGLText();
  const language = useGLLanguage();
  const usableBooks = activeJournalBooks(books);
  const dateFields = resource === "forecast";
  const champLabels = champMasterLabels(resource, tr);
  return <><div className="grid items-start gap-3 sm:grid-cols-2">
    <Field label={champLabels?.code ?? tr("gl_code", "รหัส")}><CodeInput locked={!!value.id} data-field="code" required value={value.code || ""} onChange={(e) => set({ code: e.target.value })} /></Field>
    <Field label={champLabels?.name ?? tr("gl_name", "ชื่อ")}><input className={control} data-field="name" required value={value.name || ""} onChange={(e) => set({ name: e.target.value })} /></Field>
    {dateFields && <><Field label={tr("gl_fiscal_year", "ปีบัญชี")}><YearSelect value={value.fiscalyear || ""} onChange={(fiscalyear) => set({ fiscalyear })} years={years} /></Field><Field label={tr("gl_start_date", "วันเริ่มต้น")}><input className={control} type="date" required value={value.startdate || ""} onChange={(e) => set({ startdate: e.target.value })} /></Field><Field label={tr("gl_end_date", "วันสิ้นสุด")}><input className={control} type="date" required value={value.enddate || ""} onChange={(e) => set({ enddate: e.target.value })} /></Field></>}
    {resource === "forecast" && <><Field label={tr("gl_account", "บัญชี")}><AccountSelect value={value.accountcode || ""} onChange={(accountcode) => set({ accountcode })} accounts={accounts} /></Field><Field label={tr("gl_amount", "จำนวนเงิน")}><AmountInput required value={value.amount || ""} onChange={(amount) => set({ amount })} /></Field><Field label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} value={value.branchcode || ""} onChange={(e) => set({ branchcode: e.target.value })} /></Field><Field label={tr("gl_department_code", "รหัสแผนก")}><input className={control} value={value.departmentcode || ""} onChange={(e) => set({ departmentcode: e.target.value })} /></Field><Field label={tr("gl_project_code", "รหัสโครงการ")}><input className={control} value={value.projectcode || ""} onChange={(e) => set({ projectcode: e.target.value })} /></Field></>}
    {resource === "forecast" && <Field label={tr("gl_money_direction", "ทิศทางเงิน")}><ChoiceSelect value={value.direction || "in"} onChange={(direction) => set({ direction })}><option value="in">{tr("gl_money_in", "เงินเข้า")}</option><option value="out">{tr("gl_money_out", "เงินออก")}</option></ChoiceSelect></Field>}
    {resource === "product-account-groups" && <>{PRODUCT_GROUP_ACCOUNTS.map(({ key, label }) => { const text = tr(...label); return <Field key={key} label={text}><AccountSelect label={text} value={value[key] ?? ""} onChange={(accountcode) => set({ [key]: accountcode })} accounts={accounts} /></Field>; })}</>}
    {resource === "mappings" && <Field label={tr("gl_book_code", "รหัสสมุดรายวัน")}><Combobox value={value.bookcode || ""} onChange={(bookcode) => set({ bookcode })}><option value="">{tr("gl_select", "เลือก")}</option>{usableBooks.map((book) => <option key={book.code} value={book.code}>{book.code} · {journalBookName(book, book.code, language)}</option>)}{value.bookcode && !usableBooks.some((book) => book.code === value.bookcode) && <option value={value.bookcode}>{value.bookcode} {tr("gl_book_inactive_suffix", "(ปิดใช้งาน — เลือกสมุดอื่น)")}</option>}</Combobox></Field>}
    <div className="sm:col-span-2 flex flex-wrap items-center gap-x-6 gap-y-2 pt-1">
      <Check label={tr("gl_enabled", "เปิดใช้งาน")} checked={value.isactive ?? true} onChange={(isactive) => set({ isactive })} />
    </div>
  </div>
    {resource === "mappings" && <div className="grid gap-2"><h3 className="font-semibold">{tr("gl_account_mapping_rules", "กฎการเชื่อมบัญชี")}</h3>{(value.rules ?? []).map((rule, index) => <div key={index} className="grid gap-2 rounded-xl border border-border p-2 sm:grid-cols-2 shadow-sm bg-muted/20"><Field label={tr("gl_accounts_in_rule", "รหัสผังบัญชี")}><AccountSelect label={tr("gl_accounts_in_rule", "รหัสผังบัญชี")} value={rule.accountcode || ""} accounts={accounts} onChange={(accountcode) => set({ rules: (value.rules ?? []).map((item, i) => i === index ? { ...item, accountcode } : item) })} /></Field><Field label={tr("gl_accounting_side_rule", "ประเภท")}><ChoiceSelect value={rule.side || "debit"} onChange={(side) => set({ rules: (value.rules ?? []).map((item, i) => i === index ? { ...item, side } : item) })}><option value="debit">{tr("gl_debit", "เดบิต")}</option><option value="credit">{tr("gl_credit", "เครดิต")}</option></ChoiceSelect></Field><Field label={tr("gl_amount_source_rule", "เงื่อนไข")}><input className={control} value={rule.source || ""} placeholder={tr("gl_select_supported_source_doc", "เลือกตามเอกสารต้นทางที่ระบบรองรับ")} onChange={(e) => set({ rules: (value.rules ?? []).map((item, i) => i === index ? { ...item, source: e.target.value } : item) })} /></Field><Button type="button" className={actionClass} variant="outline" onClick={() => set({ rules: (value.rules ?? []).filter((_, i) => i !== index) })}>{tr("gl_remove_rule", "นำกฎ {0} ออก").replace("{0}", String(index + 1))}</Button></div>)}<Button type="button" variant="outline" className={actionClass} onClick={() => set({ rules: [...(value.rules ?? []), { accountcode: "", side: "debit", source: "" }] })}>{tr("gl_add_account_mapping_rule", "เพิ่มกฎการเชื่อมบัญชี")}</Button></div>}
  </>;
}

/** สมุดเดิมที่ยังไม่กำหนดประเภทใช้บันทึกใบใหม่ไม่ได้ — แสดงรายชื่อทั้งหมด (ไม่ใช่เฉพาะหน้าที่เปิดอยู่) ให้กดเปิดไปเลือกประเภทได้ทันที
 *  เปลี่ยนชื่อหรือปิดใช้งานสมุดเหล่านี้ได้โดยไม่ต้องเลือกประเภท (backend allowUntyped) */
function UntypedJournalBooksNotice({ books, onOpen }: { books: GLJournalBook[]; onOpen: (book: GLJournalBook) => void }) {
  const tr = useGLText();
  const untyped = untypedJournalBooks(books);
  if (!untyped.length) return null;
  return (
    <div role="status" data-field="untyped-journal-books" className="rounded-xl border border-destructive/40 bg-destructive/5 p-3 text-[0.95rem] leading-relaxed text-foreground shadow-[0_2px_8px_rgba(0,0,0,0.08)]">
      <p className="font-semibold">{tr("gl_books_untyped_notice", "สมุดรายวัน {0} เล่มยังไม่ได้กำหนดประเภท — เลือกประเภทก่อนบันทึกเอกสาร").replace("{0}", String(untyped.length))}</p>
      <p className="text-[0.9rem] text-muted-foreground">{tr("gl_books_untyped_hint", "กดชื่อสมุดเพื่อเปิด แล้วกดแก้ไขและเลือกประเภท (ทั่วไป จ่าย รับ ขาย ซื้อ หรือยอดยกมา) — ถ้าเลิกใช้ ให้ปิดใช้งานได้เลยโดยไม่ต้องเลือกประเภท")}</p>
      <ul className="mt-2 flex flex-wrap gap-2">
        {untyped.map((book) => (
          <li key={book.id || book.code}>
            <Button type="button" variant="outline" className="min-h-11 h-auto whitespace-normal text-left" onClick={() => onOpen(book)}>{book.code} · {book.name}</Button>
          </li>
        ))}
      </ul>
    </div>
  );
}

/** ป้ายประเภทสมุดในรายการ — สมุดที่ยังไม่กำหนดประเภทใช้บันทึกใบใหม่ไม่ได้ จึงแสดงเตือนเป็นข้อความ */
function JournalBookTypeBadge({ booktype }: { booktype?: number }) {
  const tr = useGLText();
  const label = journalBookTypeLabels[String(booktype)];
  return (
    <span className={`ml-2 inline-flex items-center rounded px-1.5 py-0.5 text-[0.7rem] font-semibold border leading-none h-5 align-middle ${label ? "bg-muted text-muted-foreground border-border" : "bg-destructive/10 text-destructive border-destructive/30"}`}>
      {label ? tr(label[0], label[1]) : tr("gl_book_type_missing_badge", "ยังไม่กำหนดประเภท")}
    </span>
  );
}

/** ฟอร์มสมุดรายวัน (journalbook.sql): รหัส ≤ 15 ตัวอักษร, ชื่อไทยบังคับ, ชื่ออังกฤษไม่บังคับ, ประเภทสมุด 1–6 บังคับ
 *  สมุดที่มีใบสำคัญใช้อยู่ backend ไม่ให้เปลี่ยนรหัส/ประเภทหรือลบ (แจ้งข้อความไทยกลับมา) — แก้ชื่อและเปิด/ปิดใช้งานได้เสมอ */
function JournalBookFields({ value, set }: { value: GLJournalBook; set: (patch: object) => void }) {
  const tr = useGLText();
  const codeLength = [...(value.code ?? "")].length;
  return (
    <>
      {value.id && <Notice text={tr("gl_journal_book_in_use_hint", "สมุดที่มีใบสำคัญใช้อยู่แล้ว เปลี่ยนรหัสหรือประเภทไม่ได้ และลบไม่ได้ — ถ้าเลิกใช้ให้ปิดใช้งานแทน (แก้ชื่อได้เสมอ)")} />}
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label={tr("gl_book_code", "รหัสสมุดรายวัน")} hint={tr("gl_book_code_hint", "ไม่เกิน {0} ตัวอักษร เช่น JV, PV, RV, SV, UV").replace("{0}", String(JOURNAL_BOOK_CODE_MAX))}>
          <input className={control} data-field="code" required value={value.code || ""} aria-invalid={codeLength > JOURNAL_BOOK_CODE_MAX || undefined} onChange={(e) => set({ code: e.target.value })} />
          {codeLength > JOURNAL_BOOK_CODE_MAX && (
            <span role="alert" className="text-[0.9rem] leading-snug text-destructive">
              {tr("gl_book_code_too_long", "รหัสสมุดรายวันยาวได้ไม่เกิน {0} ตัวอักษร (ตอนนี้ {1} ตัว)").replace("{0}", String(JOURNAL_BOOK_CODE_MAX)).replace("{1}", String(codeLength))}
            </span>
          )}
        </Field>
        <Field label={tr("gl_book_name", "ชื่อสมุดรายวัน")}>
          <input className={control} data-field="name" required value={value.name || ""} onChange={(e) => set({ name: e.target.value })} />
        </Field>
        {/* Champ BCGLBook.TitleName — หัวใบสำคัญที่พิมพ์ (gl-print.tsx) */}
        <Field label={tr("gl_book_title_name", "หัวเอกสาร")} hint={tr("gl_book_title_hint", "ชื่อใบสำคัญที่พิมพ์ เช่น ใบสำคัญจ่าย — เว้นว่างจะพิมพ์ชื่อสมุด")}>
          <input className={control} data-field="titlename" maxLength={JOURNAL_BOOK_NAME_MAX} value={value.titlename || ""} onChange={(e) => set({ titlename: e.target.value })} />
        </Field>
        <Field label={tr("gl_book_type", "ประเภทสมุดรายวัน")} hint={value.id && !isJournalBookType(value.booktype) ? tr("gl_book_type_untyped_hint", "ยังไม่ได้กำหนดประเภท — เลือกประเภทก่อนบันทึกเอกสาร (เปลี่ยนชื่อหรือปิดใช้งานได้โดยไม่ต้องเลือก)") : undefined}>
          <Combobox data-field="booktype" aria-label={tr("gl_book_type", "ประเภทสมุดรายวัน")} value={value.booktype ? String(value.booktype) : ""} onChange={(booktype) => set({ booktype: Number(booktype) || 0 })}>
            <option value="">{tr("gl_select", "เลือก")}</option>
            {JOURNAL_BOOK_TYPES.map((type) => <option key={type} value={String(type)}>{tr(journalBookTypeLabels[type][0], journalBookTypeLabels[type][1])}</option>)}
          </Combobox>
        </Field>
        <Field label={tr("gl_name_en", "ชื่อภาษาอังกฤษ (ไม่บังคับ)")}>
          <input className={control} data-field="nameen" value={value.nameen || ""} onChange={(e) => set({ nameen: e.target.value })} />
        </Field>
        {/* เลขรันแบบ Champ (BRunningBttn) — backend journal_docno.go ออกเลขตอนบันทึกใบใหม่ที่เว้นเลขที่ไว้ */}
        <Field label={tr("gl_book_doc_format", "รูปแบบเลขที่เอกสาร")} hint={tr("gl_book_doc_format_hint", "@ = รหัสสมุด · YY = ปี พ.ศ. 2 หลัก · MM = เดือน · DD = วัน · # = เลขรัน 1 หลัก — เว้นว่างใช้ {0}").replace("{0}", JOURNAL_DOC_FORMAT_DEFAULT)}>
          <input className={control} data-field="docformat" maxLength={JOURNAL_DOC_FORMAT_MAX} placeholder={JOURNAL_DOC_FORMAT_DEFAULT} value={value.docformat || ""} onChange={(e) => set({ docformat: e.target.value.toUpperCase() })} />
        </Field>
        <div className="sm:col-span-2 flex flex-wrap items-center gap-x-6 gap-y-2 pt-1">
          <Check label={tr("gl_enabled", "เปิดใช้งาน")} checked={value.isactive ?? true} onChange={(isactive) => set({ isactive })} />
        </div>
      </div>
    </>
  );
}
