"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Eye, FileText, Pencil, Plus, Printer, RefreshCw, Save, Search, Trash2, X, ClipboardPaste, Scale, Table2, CheckCircle2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { accountName, activeJournalBooks, amountString, amountUnits, defaultJournalBookCode, findJournalBook, fiscalYearForDate, journalBookName, journalBookProblem, labelText, type GLLabel, emptyJournal, workspaceBranchCode, emptyLine, formatAmount, journalTotals, localDate, normalizeJournalLines, validateJournal, type GLJournal, type GLLine } from "@/lib/general-ledger";
import { GLCommandError, glRequest } from "@/lib/general-ledger-api";
import { workspaceStorageKeys } from "@/lib/workspace-models";
import { formatAppDate } from "@/lib/date-time";
import { AccountSelect, AmountInput, Combobox, Field, Notice, Pager, SearchInput, SplitWorkbench, UnsavedBadge, actionClass, control, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useRowDensity, useGLLanguage, useGLText } from "./gl-common";
import { useFormShortcuts } from "@/hooks/use-form-shortcuts";
import { isTabularPaste, parseClipboardJournalLines, type ClipboardJournalIssue } from "@/lib/clipboard-journal-parser";
import { useTabularEnterNav } from "@/hooks/use-tabular-enter-nav";
import { analyzeGLTaxAndBalance, autoBalanceJournalLines, setExactVatLine } from "@/lib/gl-smart-guard";
import { GLJournalDetailsPanel } from "./gl-journal-details";
import { GLVoucherPrint, printCompanyName, useGLPrint } from "./gl-print";
import { detailTaxIdTarget, journalDetailsProblem, normalizeJournalDetails, reconciliationChanges, type GLDetailVat } from "@/lib/gl-journal-details";
import { GLJournalReviewPanel } from "./gl-journal-review";
import { ChampField, ChampGroupBox, ChampKey, ChampSpacer, ChampTaxBox } from "./gl-journal-boxes";
import { JOURNAL_SAVE_OPTIONS, dropBlankTaxBoxRow, isBlankVatRow, journalFunctionKey, journalSaveOptionStorageKey, linesHaveAmounts, nextJournalAfterSave, parseJournalSaveOption, patchTaxBox, setTaxBoxType, taxBoxCalcInput, taxBoxView, taxLinesFromCalc, vatErrorTarget, type JournalSaveOption, type VatCalcResult } from "@/lib/gl-journal-entry";
import { getAuthSession } from "@/lib/client-auth-session";

const statusLabel: Record<string, GLLabel> = { draft: ["gl_draft", "ฉบับร่าง"], posted: ["gl_posted", "ผ่านรายการแล้ว"], reversed: ["gl_reversed", "กลับรายการแล้ว"] };

/** สีป้ายสมุดตามประเภทสมุด (booktype) ด้วย theme token เท่านั้น — รหัส/ชื่อสมุดผู้ใช้กำหนดเอง จึงห้ามเลือกสีจากรหัส; ป้ายแสดงรหัสเป็นข้อความเสมอ (ไม่สื่อด้วยสีอย่างเดียว) */
export function journalBookBadgeClass(booktype?: number): string {
  switch (Number(booktype)) {
    case 2: return "bg-destructive/10 text-destructive border-destructive/30";
    case 3: return "bg-primary/10 text-primary border-primary/30";
    case 4: return "bg-accent text-accent-foreground border-primary/20";
    case 5: return "bg-secondary text-secondary-foreground border-border";
    case 6: return "bg-background text-muted-foreground border-dashed border-border";
    default: return "bg-muted text-foreground border-border";
  }
}

/** ข้อความแจ้งแถวที่วางจาก Excel ไม่ได้ (บอกแถว ช่อง ค่า และวิธีแก้) — ทั้งชุดไม่ถูกวางจนกว่าจะแก้ */
export function pasteIssueText(issue: ClipboardJournalIssue, tr: (key: string, fallback: string) => string): string {
  const field = issue.field === "debit" ? tr("gl_debit", "เดบิต") : issue.field === "credit" ? tr("gl_credit", "เครดิต") : tr("gl_amount", "จำนวนเงิน");
  const text = issue.problem === "negative"
    ? tr("gl_paste_issue_negative", "แถวที่ {0} ช่อง{1} เป็นยอดติดลบ {2} — ใส่ยอดเป็นบวกในช่องฝั่งตรงข้ามแทน")
    : tr("gl_paste_issue_invalid", "แถวที่ {0} ช่อง{1} ไม่ใช่ตัวเลข ({2}) — ใส่เป็นตัวเลข เช่น 1,500.00");
  return text.replace("{0}", String(issue.row)).replace("{1}", field).replace("{2}", issue.value);
}

/**
 * Champ search box date (QList.cpp 1719-1758): when the whole text is d/m/yyyy (separator / . or -) the list is
 * filtered by document date instead of text. Day 0 = the whole month. A year of 2400 or more is Buddhist (minus 543),
 * 1900-2399 is Gregorian. A day or month out of range, or a day that does not exist (31/02), is not a date: null.
 */
export function parseChampDateSearch(text: string): { from: string; to: string } | null {
  const parts = text.trim().split(/[/.-]/);
  if (parts.length !== 3 || !/^[0-9]{1,2}$/.test(parts[0]) || !/^[0-9]{1,2}$/.test(parts[1]) || !/^[0-9]{4}$/.test(parts[2])) return null;
  const day = Number(parts[0]), month = Number(parts[1]);
  let year = Number(parts[2]);
  if (year < 1900 || year > 4000 || month < 1 || month > 12) return null;
  if (year >= 2400) year -= 543;
  const lastDay = new Date(Date.UTC(year, month, 0)).getUTCDate();
  if (day > lastDay) return null;
  const prefix = `${String(year).padStart(4, "0")}-${String(month).padStart(2, "0")}-`;
  const pad = (value: number) => String(value).padStart(2, "0");
  return day === 0 ? { from: `${prefix}01`, to: `${prefix}${pad(lastDay)}` } : { from: `${prefix}${pad(day)}`, to: `${prefix}${pad(day)}` };
}

/** Query string of the journal document list: book tab, kind, status and the inclusive document-date range (YYYY-MM-DD) from a date typed in the search box. */
export function journalListFilters(input: { bookcode?: string; kind?: string; status?: string; from?: string; to?: string }): string {
  return new URLSearchParams({
    ...(input.bookcode ? { bookcode: input.bookcode } : {}),
    ...(input.kind ? { kind: input.kind } : {}),
    ...(input.status ? { status: input.status } : {}),
    ...(input.from ? { from: input.from } : {}),
    ...(input.to ? { to: input.to } : {}),
  }).toString();
}

/**
 * Intelligent helper to create the next journal row with automatic debit/credit balance.
 * If debit > credit, pre-fills remaining credit.
 * If credit > debit, pre-fills remaining debit.
 */
export function createSmartNextLine(lines: GLLine[], scale = 2): GLLine {
  let totalDebit = 0n;
  let totalCredit = 0n;
  for (const line of lines) {
    try {
      if (line.debit && line.debit !== "0") totalDebit += amountUnits(line.debit);
    } catch {}
    try {
      if (line.credit && line.credit !== "0") totalCredit += amountUnits(line.credit);
    } catch {}
  }
  const diff = totalDebit - totalCredit;
  const line = emptyLine();
  if (diff > 0n) {
    line.credit = amountString(diff, scale);
  } else if (diff < 0n) {
    line.debit = amountString(-diff, scale);
  }
  return line;
}

/** พาไปที่ช่องรายละเอียดที่ต้องแก้: เปิดหมวด เลื่อนมาให้เห็น (ไม่กระโดดถ้าผู้ใช้ปิด motion) แล้ว focus — แถบแจ้งด้านบนบอกเหตุผล
 *  เดิมบอกแค่แถบแจ้ง ผู้ใช้ต้องไล่หาเองว่าแถวไหนผิด (UAT S3 2026-09-24) */
function focusDetailField(target: { section: string; row: number; field: string } | null) {
  if (!target || typeof document === "undefined") return;
  requestAnimationFrame(() => {
    const cell = document.querySelector<HTMLElement>(`[data-detail-field="${target.section}.${target.row}.${target.field}"]`);
    const input = cell?.querySelector<HTMLElement>("input, select, textarea, button");
    const group = cell?.closest("details");
    if (group) group.open = true;
    const reduceMotion = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
    input?.scrollIntoView({ block: "center", behavior: reduceMotion ? "auto" : "smooth" });
    input?.focus({ preventScroll: true });
  });
}

/** ปุ่มคำสั่งด้านขวาแบบ Champ: แผงกว้าง (@4xl) = คอลัมน์ปุ่มเต็มความกว้าง ข้อความชิดซ้าย ปุ่มลัดชิดขวา */
const champCommandClass = `${actionClass} @4xl:w-full @4xl:justify-start`;

/** F3 ตารางภาษี / F5 รายละเอียดอื่นๆ: เปิดหมวดในแผงรายละเอียด (null = ทั้งแผง) เลื่อนมาให้เห็น แล้ว focus หัวหมวด */
function openDetailsSection(section: string | null) {
  if (typeof document === "undefined") return;
  requestAnimationFrame(() => {
    const panel = document.querySelector<HTMLElement>("[data-gl-details-panel]");
    const group = section ? panel?.querySelector<HTMLDetailsElement>(`[data-detail-section="${section}"]`) : null;
    if (group) group.open = true;
    const target = group?.querySelector<HTMLElement>("summary") ?? panel;
    const reduceMotion = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
    target?.scrollIntoView({ block: "start", behavior: reduceMotion ? "auto" : "smooth" });
    target?.focus({ preventScroll: true });
  });
}

/** F2 เลือกรายการ: กลับไปที่ช่องค้นหาของรายการใบสำคัญ (Champ เปิดหน้าต่างค้นหาเอกสาร) */
function focusJournalSearch() {
  document.querySelector<HTMLInputElement>("[data-gl-journal-search] input")?.focus();
}

export function GLJournals({ route, book = "", kind = "", mode = "edit" }: { route: string; book?: string; kind?: string; mode?: "edit" | "reverse" }) {
  const tr = useGLText();
  const [selectedBook, setSelectedBook] = useState(book || "");
  const activeBook = selectedBook || "";
  const [search, setSearch] = useState("");
  // reverse = ยกเลิกการผ่านรายการ (กลับรายการทีละใบ): แสดงเฉพาะใบที่ผ่านรายการแล้ว; ผ่านรายการทำเป็นชุดที่ gl-post-batch แบบ Champ 5512
  const activeStatus = mode === "reverse" ? "posted" : "";
  const dateSearch = parseChampDateSearch(search);
  const filters = journalListFilters({ bookcode: activeBook, kind, status: activeStatus, from: dateSearch?.from, to: dateSearch?.to });
  const list = useGLList<GLJournal>("journals", dateSearch ? "" : search, filters), refs = useReferences();
  const language = useGLLanguage();
  const voucherPrint = useGLPrint();
  const books = useMemo(() => refs.books ?? [], [refs.books]);
  const usableBooks = useMemo(() => activeJournalBooks(books), [books]);
  const bookLabel = (code: string) => journalBookName(findJournalBook(books, code), code, language);
  // สมุดของใบใหม่: แท็บที่เลือก → (ยอดยกมา) สมุดประเภทยอดยกมา → สมุดทั่วไป — ไม่ผูกกับรหัสสมุดใดตายตัว
  const newBookCode = () => defaultJournalBookCode(books, activeBook || (kind === "opening" ? usableBooks.find((item) => Number(item.booktype) === 6)?.code ?? "" : ""));
  const searchDebounce = useDebouncedSearch({
    onSearch: (val) => {
      list.setPage(1);
      setSearch(val);
    },
    debounceMs: 400,
  });
  const density = useRowDensity();
  const [journal, setJournal] = useState<GLJournal | null>(null), [original, setOriginal] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [detailsEditing, setDetailsEditing] = useState(false);
  const [detailsBusy, setDetailsBusy] = useState(false);
  const [message, setMessage] = useState(""), [error, setError] = useState("");
  const [reason, setReason] = useState(""), [reverseDate, setReverseDate] = useState(localDate()), [reverseDocno, setReverseDocno] = useState("");
  const [reviewDirty, setReviewDirty] = useState(false), [reviewBusy, setReviewBusy] = useState(false);
  const [reviewLoad, setReviewLoad] = useState(0);
  // ตัวเลือกหลังบันทึกแบบ Champ — จำแยกผู้ใช้ในเครื่องนี้ (อ่านหลัง mount เพราะตอน render ฝั่ง server ไม่มี localStorage)
  const [saveOption, setSaveOption] = useState<JournalSaveOption>("keep");
  useEffect(() => { setSaveOption(parseJournalSaveOption(localStorage.getItem(journalSaveOptionStorageKey(getAuthSession()?.username)))); }, []);
  const changeSaveOption = (value: JournalSaveOption) => { setSaveOption(value); localStorage.setItem(journalSaveOptionStorageKey(getAuthSession()?.username), value); };
  const [vatCalcBusy, setVatCalcBusy] = useState(false);
  const { busy: journalBusy, execute } = useGLCommand(), { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const busy = journalBusy || reviewBusy || detailsBusy;
  const isJournalDirty = journal !== null && JSON.stringify(journal) !== original;
  const isReasonDirty = reason.trim() !== "";
  const detailsDirty = journal?.status === "posted" && JSON.stringify(journal.details ?? {}) !== JSON.stringify(original ? (JSON.parse(original) as GLJournal).details ?? {} : {});
  const dirty = reviewDirty || detailsDirty || (isEditing && journal !== null && (isJournalDirty || isReasonDirty));

  useDirtyGuard(route, dirty);
  const year = refs.years.find((item) => item.code === journal?.fiscalyear);
  const accountNames = useMemo(() => new Map(refs.accounts.map((account) => [account.accountcode, accountName(account)])), [refs.accounts]);
  // Champ OnCheckSeparate: แผนก/โครงการใส่ได้เฉพาะบัญชีที่ตั้ง "แยกแผนก"/"แยกโครงการ" — ช่องที่มีค่าอยู่แล้วยังแก้/ล้างได้
  const accountsByCode = useMemo(() => new Map(refs.accounts.map((account) => [account.accountcode, account])), [refs.accounts]);
  const departLocked = (line: { accountcode: string; departmentcode?: string }) => !accountsByCode.get(line.accountcode)?.departseparate && !line.departmentcode;
  const projectLocked = (line: { accountcode: string; projectcode?: string }) => !accountsByCode.get(line.accountcode)?.projectseparate && !line.projectcode;
  const totals = useMemo(() => { try { return journal ? journalTotals(journal.lines) : null; } catch { return null; } }, [journal]);
  const smartGuard = useMemo(() => {
    try {
      return journal ? analyzeGLTaxAndBalance(journal.lines, refs.accounts, year?.scale ?? 2) : null;
    } catch {
      return null;
    }
  }, [journal, refs.accounts, year?.scale]);
  const patch = (value: Partial<GLJournal>) => setJournal((current) => current ? { ...current, ...value } : current);
  const patchLine = (index: number, value: Partial<GLLine>) => patch({ lines: journal!.lines.map((line, i) => i === index ? { ...line, ...value } : line) });
  // Champ รันเลขที่ให้ใบใหม่ (BRunningEdit) — แสดงเลขถัดไปของสมุดตามวันที่เป็นตัวอย่าง; เลขจริงออกตอนบันทึก (backend journal_docno.go)
  const [nextDocNo, setNextDocNo] = useState("");
  // บันทึกใบใหม่แล้วสมุด/วันที่มักเท่าเดิม (ตัวเลือก "ไม่ล้างหน้าจอ") — นับรอบบันทึกเพื่อขอเลขถัดไปใหม่ ไม่ค้างเลขที่เพิ่งใช้
  const [savedNewCount, setSavedNewCount] = useState(0);
  const previewBook = isEditing && journal && !journal.id ? journal.bookcode : "", previewDate = journal?.date ?? "";
  useEffect(() => {
    if (!previewBook || !previewDate) { setNextDocNo(""); return; }
    const controller = new AbortController();
    glRequest<{ docno: string }>(`journal-next-docno?${new URLSearchParams({ bookcode: previewBook, from: previewDate })}`, { signal: controller.signal })
      .then((value) => setNextDocNo(value?.docno ?? ""))
      .catch(() => { if (!controller.signal.aborted) setNextDocNo(""); });
    return () => controller.abort();
  }, [previewBook, previewDate, savedNewCount]);

  const addNewLine = () => {
    if (!journal || journal.lines.length >= 500) return;
    const nextLine = createSmartNextLine(journal.lines, year?.scale ?? 2);
    patch({ lines: [...journal.lines, nextLine] });
    requestAnimationFrame(() => {
      const table = document.querySelector(".gl-workbench table");
      if (!table) return;
      const rows = table.querySelectorAll("tbody tr");
      const lastRow = rows[rows.length - 1];
      if (lastRow) {
        const input = lastRow.querySelector<HTMLElement>("input:not([disabled]):not([readonly]), button:not([disabled])");
        input?.focus?.();
      }
    });
  };

  const tabularEnterNav = useTabularEnterNav({
    onAddNewRow: () => {
      addNewLine();
    },
  });

  // หัวเอกสาร: Enter เลื่อนไปช่องถัดไปแบบ Champ (ไม่ส่งฟอร์มบันทึก) — ช่องสุดท้ายพาเข้าตารางบัญชี
  const linesRef = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const headerEnterNav = useTabularEnterNav({
    onAddNewRow: () => linesRef.current?.querySelector<HTMLElement>("input:not([disabled]):not([readonly]), button:not([disabled])")?.focus(),
  });

  const applyPastedLines = (clipboardText: string) => {
    if (!journal || !clipboardText.trim()) return;
    const { lines: parsed, issues } = parseClipboardJournalLines(clipboardText);
    if (issues.length > 0) {
      setMessage("");
      setError(tr("gl_paste_rejected", "ยังไม่ได้วางข้อมูล — แก้ใน Excel แล้ววางใหม่: {0}").replace("{0}", issues.slice(0, 3).map((issue) => pasteIssueText(issue, tr)).join(" · ")));
      return;
    }
    if (parsed.length === 0) {
      setError(tr("gl_paste_no_rows", "ไม่พบแถวข้อมูลที่วางได้ — คัดลอกคอลัมน์ รหัสบัญชี คำอธิบาย เดบิต เครดิต จาก Excel"));
      return;
    }

    const isEmptyJournal = journal.lines.every(
      (l) => !l.accountcode && !l.description && (!l.debit || l.debit === "0") && (!l.credit || l.credit === "0")
    );
    const combined = isEmptyJournal ? parsed : [...journal.lines, ...parsed];
    if (combined.length > 500) {
      setError(tr("gl_paste_too_many_lines", "ใบสำคัญมีได้ไม่เกิน 500 บรรทัด — ถ้าวางจะมี {0} บรรทัด กรุณาแบ่งเป็นหลายใบ").replace("{0}", String(combined.length)));
      return;
    }
    patch({ lines: combined });
    setError("");
    setMessage(tr("gl_paste_rows_added", "วางข้อมูลจาก Excel แล้ว {0} บรรทัด — ตรวจบัญชีและยอดก่อนบันทึก").replace("{0}", String(parsed.length)));
  };

  const handlePasteClick = async () => {
    try {
      if (!navigator?.clipboard?.readText) {
        setError(tr("gl_clipboard_unsupported", "เบราว์เซอร์นี้อ่านคลิปบอร์ดไม่ได้ — คลิกในตารางบรรทัดบัญชีแล้วกด Ctrl+V แทน"));
        return;
      }
      const text = await navigator.clipboard.readText();
      if (text) {
        applyPastedLines(text);
      }
    } catch {
      setError(tr("gl_clipboard_denied", "เบราว์เซอร์ไม่อนุญาตให้อ่านคลิปบอร์ด — คลิกในตารางบรรทัดบัญชีแล้วกด Ctrl+V แทน"));
    }
  };

  async function closeView() {
    if (busy) return;
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    setJournal(null); setOriginal(""); setIsEditing(false);
  }

  async function openView(item: GLJournal) {
    if (busy) return;
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLJournal>(`journals/${encodeURIComponent(item.id)}`) : emptyJournal(newBookCode(), kind || "manual", refs.years, workspaceBranchCode(localStorage.getItem(workspaceStorageKeys.workspace)));
      setJournal(value);
      setDetailsEditing(false);
      setReviewLoad((value) => value + 1);
      setOriginal(JSON.stringify(value));
      setIsEditing(false);
      setError("");
      setMessage("");
      setReason("");
      setReverseDocno(""); setReverseDate(localDate());
    } catch (e) {
      setError((e as Error).message);
    }
  }

  async function openEdit(item: GLJournal) {
    if (busy) return;
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLJournal>(`journals/${encodeURIComponent(item.id)}`) : emptyJournal(newBookCode(), kind || "manual", refs.years, workspaceBranchCode(localStorage.getItem(workspaceStorageKeys.workspace)));
      setJournal(value);
      setDetailsEditing(false);
      setOriginal(JSON.stringify(value));
      setIsEditing(true);
      setError("");
      setMessage("");
      setReason("");
      setReverseDocno(""); setReverseDate(localDate());
    } catch (e) {
      setError((e as Error).message);
    }
  }

  async function openCreate() {
    if (busy) return;
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    const value = emptyJournal(newBookCode(), kind || "manual", refs.years, workspaceBranchCode(localStorage.getItem(workspaceStorageKeys.workspace)));
    setJournal(value);
    setOriginal(JSON.stringify(value));
    setIsEditing(true);
    setError("");
    setMessage("");
    setReason("");
    setReverseDocno(""); setReverseDate(localDate());
  }

  async function cancelEdit() {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_entries", "ละทิ้งรายการที่ยังไม่บันทึก?"), description: tr("gl_unsaved_entries_not_saved", "รายการที่กรอกอยู่จะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    setDetailsEditing(false);
    if (journal?.id) {
      setJournal(JSON.parse(original));
      setIsEditing(false);
    } else {
      setJournal(null);
      setOriginal("");
      setIsEditing(false);
    }
    setError("");
    setReason("");
  }

  async function save() {
    if (!journal || busy || !isEditing) return;
    const saved = journal.id && original ? (JSON.parse(original) as GLJournal) : undefined;
    const bookProblem = journalBookProblem(journal.bookcode, books, saved?.bookcode, tr);
    if (bookProblem) { setError(bookProblem); return; }
    const candidate: GLJournal = { ...journal, lines: normalizeJournalLines(journal.lines) };
    const problem = validateJournal(candidate, year, refs.accounts, tr);
    if (problem) { setError(problem); return; }
    // กรอบภาษีที่เลือกประเภทไว้แต่ไม่ได้กรอกอะไร ไม่ต้องกลายเป็นใบกำกับเปล่า (Champ ไม่บันทึกภาษีเมื่อไม่มีเลขที่ใบกำกับ)
    const details = dropBlankTaxBoxRow(journal.details);
    const detailProblem = journalDetailsProblem(details, tr);
    if (detailProblem) { setError(detailProblem.message); focusDetailField(detailProblem); return; }
    const payload: GLJournal = { ...candidate, details: normalizeJournalDetails(details) };
    if (journal.id && !await confirm({ title: tr("gl_save_draft_changes", "บันทึกการแก้ไขฉบับร่าง?"), description: journal.docno, confirmLabel: tr("gl_save_draft", "บันทึกฉบับร่าง"), tone: "info" })) return;
    try {
      const result = await execute({ resource: "journals", id: journal.id, version: journal.version, action: journal.id ? "update" : "create", journal: payload });
      // backend เติมค่าที่คำนวณเอง (เช่น ภาษีหัก ณ ที่จ่าย = ฐาน × อัตรา) — แสดงฉบับที่บันทึกจริง; โหลดไม่ได้ใช้ค่าที่ส่งไป (คำสั่ง commit แล้ว)
      const reloaded = await glRequest<GLJournal>(`journals/${encodeURIComponent(result.id)}`).catch(() => ({ ...payload, id: result.id, version: result.version }));
      list.reload();
      setError("");
      if (!journal.id) {
        // Champ: ใบใหม่บันทึกแล้วขึ้นใบถัดไปทันทีตามตัวเลือกการบันทึก (ไม่ล้างหน้าจอ = เก็บหัวเอกสารและรหัสบัญชีไว้ ล้างเลขที่และยอด)
        const next = nextJournalAfterSave(reloaded, saveOption, emptyJournal(newBookCode(), kind || "manual", refs.years, workspaceBranchCode(localStorage.getItem(workspaceStorageKeys.workspace))));
        setJournal(next);
        setOriginal(JSON.stringify(next));
        setSavedNewCount((count) => count + 1);
        setMessage(tr("gl_saved_next_voucher", "บันทึกฉบับร่าง {0} แล้ว — พร้อมบันทึกใบถัดไป").replace("{0}", String(reloaded.docno ?? "")));
        requestAnimationFrame(() => headerRef.current?.querySelector<HTMLElement>("input:not([disabled]):not([readonly]):not([type=\"hidden\"]), select:not([disabled]), button:not([disabled])")?.focus());
        return;
      }
      setJournal(reloaded);
      setOriginal(JSON.stringify(reloaded));
      setIsEditing(false);
      setMessage(result.projectionpending ? tr("gl_draft_saved_updating_report_data", "บันทึกฉบับร่างแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน") : tr("gl_draft_saved_review_and_post", "บันทึกฉบับร่างแล้ว ตรวจสอบและกดผ่านรายการเมื่อพร้อม"));
    } catch (e) {
      setError((e as Error).message);
      if (e instanceof GLCommandError) focusDetailField(detailTaxIdTarget(payload.details, e.field) ?? vatErrorTarget(payload.details, e.code, e.field));
    }
  }

  // กรอบภาษี: เปลี่ยนเป็น "ไม่มีภาษี" ขณะใบกำกับมีข้อมูล = ถามก่อนนำออก (กันพลาด)
  async function changeTaxType(value: "" | "1" | "2") {
    if (!journal) return;
    const view = taxBoxView(journal.details);
    if (value === "" && view.mode === "single" && !isBlankVatRow(view.row) && !await confirm({ title: tr("gl_tax_remove_title", "นำใบกำกับภาษีออก?"), description: tr("gl_tax_remove_desc", "ข้อมูลใบกำกับภาษีที่กรอกในกรอบนี้จะถูกนำออกจากใบสำคัญ"), confirmLabel: tr("gl_tax_remove", "นำใบกำกับออก"), tone: "warning" })) return;
    setJournal((current) => current ? { ...current, details: setTaxBoxType(current.details, value, current.date) } : current);
  }
  const patchTaxBoxRow = (value: Partial<GLDetailVat>) => setJournal((current) => current ? { ...current, details: patchTaxBox(current.details, value) } : current);

  // ภาษีแยกนอก / ภาษีรวมใน แบบ Champ (OnTaxAuto): backend คำนวณ (GET /gl/v2/vat-calc) แล้ววางยอดในกรอบภาษีและ 3 บรรทัดแรก — ไม่เติมรหัสบัญชี
  async function calcTax(mode: "exclusive" | "inclusive") {
    if (!journal || vatCalcBusy) return;
    const input = taxBoxCalcInput(journal.details);
    if (!input.ok) {
      if (input.problem === "rate") { setError(tr("gl_tax_calc_need_rate", "ใส่อัตราภาษี (%) ก่อนคำนวณ — ระบบไม่ใส่ 7% ให้เอง")); focusDetailField({ section: "vats", row: 1, field: "vat_rate" }); }
      if (input.problem === "amount") { setError(tr("gl_tax_calc_need_amount", "ใส่จำนวนเงินก่อนคำนวณภาษี")); focusDetailField({ section: "vats", row: 1, field: "base_amount" }); }
      return;
    }
    setVatCalcBusy(true);
    try {
      const calc = await glRequest<VatCalcResult>(`vat-calc?${new URLSearchParams({ mode, amount: input.amount, rate: input.rate, ...(journal.fiscalyear ? { fiscalyear: journal.fiscalyear } : {}) })}`);
      // Champ ล้างยอดทุกบรรทัดแล้วเติมใหม่เงียบ ๆ — BC ถามก่อนถ้ามียอดเดิม (ปฏิเสธ = คำนวณเฉพาะกรอบภาษี)
      const fill = !linesHaveAmounts(journal.lines) || await confirm({ title: tr("gl_tax_fill_lines_title", "เติมยอดลงบรรทัดบัญชีด้วย?"), description: tr("gl_tax_fill_lines_desc", "ยอดเดบิตและเครดิตเดิมทุกบรรทัดจะถูกล้าง แล้วเติมฐานภาษี ภาษี และยอดรวมลง 3 บรรทัดแรก — รหัสบัญชีเลือกเอง"), confirmLabel: tr("gl_tax_fill_lines", "เติมยอดลงบรรทัด"), cancelLabel: tr("gl_tax_box_only", "คำนวณเฉพาะกรอบภาษี"), tone: "warning" });
      setJournal((current) => current ? { ...current, details: patchTaxBox(current.details, { base_amount: calc.base_amount, vat_amount: calc.vat_amount }), ...(fill ? { lines: taxLinesFromCalc(current.lines, input.taxType, calc) } : {}) } : current);
      setError("");
      setMessage(tr("gl_tax_calc_done", "คำนวณภาษีแล้ว: ฐานภาษี {0} · ภาษี {1} · รวม {2}").replace("{0}", formatAmount(calc.base_amount, year?.scale)).replace("{1}", formatAmount(calc.vat_amount, year?.scale)).replace("{2}", formatAmount(calc.total_amount, year?.scale)));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setVatCalcBusy(false);
    }
  }

  // ปุ่มฟังก์ชันแบบ Champ ทำงานเมื่อเปิดใบสำคัญอยู่ (F2 เลือกรายการ · F3 ตารางภาษี · F5 รายละเอียดอื่นๆ)
  const journalOpen = journal !== null;
  useEffect(() => {
    if (!journalOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      const action = journalFunctionKey(event);
      if (!action) return;
      event.preventDefault();
      if (action === "list") focusJournalSearch();
      else openDetailsSection(action === "tax" ? "vats" : null);
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [journalOpen]);

  async function saveReconciliation() {
    if (!journal?.id || !detailsDirty || busy) return;
    if (!reason.trim()) { setError(tr("gl_details_ui_30","ระบุเหตุผลการกระทบยอดก่อนบันทึก")); return; }
    try {
      const before = JSON.parse(original) as GLJournal;
      const detailProblem = journalDetailsProblem(journal.details, tr);
      if (detailProblem) { setError(detailProblem.message); focusDetailField(detailProblem); return; }
      const details = reconciliationChanges(before.details, normalizeJournalDetails(journal.details));
      const result = await execute({ resource: "journals", action: "reconcile", id: journal.id, version: journal.version, reason, journal: { details } });
      // The command has committed even if the following refresh fails. Clear pending additions before reload.
      const committed = { ...journal, version: result.version };
      setJournal(committed); setOriginal(JSON.stringify(committed)); setDetailsEditing(false); setReason("");
      setMessage(tr("gl_details_ui_31","บันทึกผลกระทบยอดแล้ว"));
      const saved = await glRequest<GLJournal>(`journals/${encodeURIComponent(journal.id)}`);
      setJournal(saved); setOriginal(JSON.stringify(saved)); setDetailsEditing(false); setReason(""); setReviewLoad(value => value + 1);
      setError(""); setMessage(tr("gl_details_ui_32","บันทึกผลกระทบยอดแล้ว กรุณาตรวจผลตรวจเอกสารอีกครั้ง")); list.reload();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : tr("gl_details_ui_33","บันทึกผลกระทบยอดไม่สำเร็จ"));
      if (cause instanceof GLCommandError) focusDetailField(detailTaxIdTarget(journal.details, cause.field));
    }
  }

  // Global Keyboard Shortcuts (Ctrl+S, Alt+N, Esc)
  useFormShortcuts({
    onSave: () => void (detailsDirty ? saveReconciliation() : save()),
    onNew: mode === "edit" ? () => void openCreate() : undefined,
    onCancel: isEditing || detailsEditing ? () => void cancelEdit() : undefined,
    canSave: (isEditing || detailsDirty) && dirty && !busy,
    disabled: busy,
  });

  async function act(action: "reverse" | "delete") {
    if (!journal?.id || busy || reviewDirty || detailsDirty || (isEditing && isJournalDirty)) return;
    if (!reason.trim()) { setError(tr("gl_enter_reason_before_posting", "กรุณาระบุเหตุผลก่อนทำรายการ")); return; }
    if (action === "reverse" && (!reverseDate || !reverseDocno.trim())) { setError(tr("gl_enter_date_and_reversal_doc_no", "กรุณาระบุวันที่และเลขที่ใบกลับรายการ")); return; }
    const label = action === "reverse" ? tr("gl_create_reversing_entry", "สร้างรายการกลับบัญชี") : tr("gl_delete_draft", "ลบฉบับร่าง");
    if (!await confirm({ title: `${label}?`, description: action === "reverse" ? tr("gl_create_reversal_doc", "สร้างเอกสาร {0} วันที่ {1} กลับเดบิตและเครดิตของ {2} โดยเก็บรายการเดิมไว้").replace("{0}", String(reverseDocno)).replace("{1}", formatAppDate(reverseDate, language)).replace("{2}", String(journal.docno)) : tr("gl_delete_draft_keep_history", "ลบฉบับร่าง {0} พร้อมเก็บประวัติ").replace("{0}", String(journal.docno)), details: tr("gl_debit_credit", "เดบิต {0} · เครดิต {1}").replace("{0}", String(totals ? formatAmount(amountString(totals.debit), year?.scale) : "—")).replace("{1}", String(totals ? formatAmount(amountString(totals.credit), year?.scale) : "—")), confirmLabel: label, tone: action === "delete" ? "danger" : "warning" })) return;
    try {
      await execute({ resource: "journals", action, id: journal.id, version: journal.version, reason, ...(action === "reverse" ? { date: reverseDate, docno: reverseDocno } : {}) });
      setJournal(null);
      setOriginal("");
      setIsEditing(false);
      list.reload();
      setError("");
      setMessage(tr("gl_success_message", "{0}เรียบร้อยแล้ว").replace("{0}", String(label)));
    } catch (e) { setError((e as Error).message); }
  }

  async function deleteDraftDirect(item: GLJournal) {
    if (!item.id || busy || reviewDirty || item.status !== "draft") return;
    if (!await confirm({
      title: tr("gl_delete_draft_confirm", "ลบฉบับร่าง?"),
      description: tr("gl_delete_draft_keep_history", "ลบฉบับร่าง {0} พร้อมเก็บประวัติ").replace("{0}", String(item.docno)),
      confirmLabel: tr("gl_delete_draft", "ลบฉบับร่าง"),
      tone: "danger",
    })) return;
    try {
      await execute({ resource: "journals", action: "delete", id: item.id, version: item.version, reason: tr("gl_delete_draft", "ลบฉบับร่าง") });
      if (journal?.id === item.id) {
        setJournal(null);
        setOriginal("");
        setIsEditing(false);
      }
      list.reload();
      setError("");
      setMessage(tr("gl_delete_draft_success", "ลบฉบับร่าง {0} เรียบร้อยแล้ว").replace("{0}", String(item.docno)));
    } catch (e) {
      setError((e as Error).message);
    }
  }

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="shrink-0 flex flex-col gap-2">
        <Notice error text={error || list.error || refs.error} />
        <Notice text={message} />
        {kind === "opening" && <Notice text={tr("gl_opening_balance_dr_cr_entries", "บันทึกยอดยกมาเป็นรายการเดบิตและเครดิตที่สมดุล เลือกปีบัญชี วันที่ และบัญชีคู่รายการตามยอดปิดที่ตรวจสอบแล้ว")} />}
      </div>
      <SplitWorkbench
        list={
          <div className="flex flex-col flex-1 min-h-0 gap-2">
            <div className="shrink-0 pb-3 border-b border-border/70 flex flex-col gap-2">
              <form className="flex flex-wrap items-center gap-2 w-full" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
                <div data-gl-journal-search className="contents" title={tr("gl_journal_search_date_hint", "เลขที่ คำอธิบาย หรือวันที่ เช่น 15/10/2569 (00/10/2569 = ทั้งเดือน)")}>
                <SearchInput
                  className="min-w-36 flex-1 basis-44"
                  ariaLabel={tr("gl_search_number_or_desc", "ค้นหาเลขที่หรือคำอธิบาย")}
                  placeholder={tr("gl_journal_search_date_hint", "เลขที่ คำอธิบาย หรือวันที่ เช่น 15/10/2569 (00/10/2569 = ทั้งเดือน)")}
                  value={searchDebounce.query}
                  onChange={searchDebounce.setQuery}
                  onClear={searchDebounce.clear}
                  onSearch={searchDebounce.searchNow}
                />
                </div>
                <div className="flex flex-wrap items-center gap-2 min-w-0">
                  <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />{tr("gl_search", "ค้นหา")}</Button>
                  <Button type="button" variant="outline" className={actionClass} onClick={() => { list.reload(); refs.reload(); }} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />{tr("gl_reload", "โหลดใหม่")}</Button>
                  {mode === "edit" && (
                    <Button type="button" className={actionClass} disabled={busy} onClick={() => void openCreate()}>
                      <Plus className="size-4 mr-1.5" />
                      <span>{tr("gl_add_row", "เพิ่มรายการ")}</span>
                      <kbd className="ml-1.5 hidden sm:inline-block rounded border border-primary-foreground/30 bg-primary-foreground/15 px-1.5 py-0.5 text-[10px] font-mono text-primary-foreground">
                        Alt+N
                      </kbd>
                    </Button>
                  )}
                </div>
              </form>

              {/* Journal Book Category Tabs (แยกกันข้างในตามคำสั่งลุงจืด) */}
              <div className="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none" role="tablist" aria-label={tr("gl_journal", "สมุดรายวัน")}>
                {[
                  { code: "", label: tr("gl_all", "ทั้งหมด") },
                  ...usableBooks.map((item) => ({ code: item.code, label: journalBookName(item, item.code, language) })),
                  // สมุดจากเมนู/ลิงก์ที่ไม่อยู่ในรายการใช้งาน (ปิดใช้งานแล้ว) ยังแสดงให้เห็นว่ากำลังกรองอยู่
                  ...(activeBook && !usableBooks.some((item) => item.code === activeBook) ? [{ code: activeBook, label: bookLabel(activeBook) }] : []),
                ].map((tab) => {
                  const isActive = activeBook === tab.code;
                  return (
                    <button
                      key={tab.code || "ALL"}
                      type="button"
                      role="tab"
                      aria-selected={isActive}
                      onClick={() => {
                        setSelectedBook(tab.code);
                        list.setPage(1);
                      }}
                      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold whitespace-nowrap transition-all cursor-pointer border shadow-2xs ${
                        isActive
                          ? "bg-primary text-primary-foreground border-primary shadow-xs font-bold"
                          : "bg-muted/40 text-muted-foreground hover:text-foreground hover:bg-muted/80 border-border/70"
                      }`}
                    >
                      <span>{tab.label}</span>
                      {tab.code && (
                        <span className={`rounded px-1.5 py-0.2 text-[10px] font-mono font-bold ${isActive ? "bg-primary-foreground/20 text-primary-foreground" : "bg-muted text-muted-foreground border border-border/60"}`}>
                          {tab.code}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
              {/* .bc-list-toolbar */}
              <div className="bc-list-toolbar shrink-0 p-2.5 border-b border-border/60 bg-muted/20 flex flex-wrap items-center justify-between gap-2 rounded-t-lg">
                <span className="text-xs font-semibold text-foreground">
                  {tr("gl_all_records", "รายการทั้งหมด")}:{" "}
                  <strong className="text-primary font-bold">{list.data.total.toLocaleString("th-TH")}</strong>
                </span>
              </div>
            </div>
            <div className="flex-1 min-h-[300px] overflow-y-auto scrollbar-thin rounded-xl border border-border shadow-xs" aria-busy={list.loading}>
              <table className="w-full text-left border-collapse">
                {/* .bc-list-header */}
                <thead className="bc-list-header sticky top-0 bg-muted/95 backdrop-blur z-10 border-b border-border/80 text-[0.7rem] font-extrabold uppercase tracking-wider text-muted-foreground">
                  <tr>
                    <th className="py-2 px-2.5 text-left w-36 whitespace-nowrap">{tr("gl_date_number", "วันที่ / เลขที่")}</th>
                    <th className="py-2 px-2 text-left min-w-44">{tr("gl_description", "คำอธิบาย")}</th>
                    <th className="py-2 px-2 text-right w-24">{tr("gl_status", "สถานะ")}</th>
                    <th className="py-2 pr-2.5 pl-2 text-right w-20">{tr("gl_manage", "จัดการ")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {list.data.items.map((item, index) => {
                    const isSelected = journal?.id === item.id;
                    const isRowEditing = isSelected && isEditing;
                    const isDraft = item.status === "draft";
                    return (
                      <tr
                        key={item.id}
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
                        <td className="whitespace-nowrap py-2 px-2.5 align-middle">
                          <div className="text-[0.8rem] text-muted-foreground">{item.date}</div>
                          <div className="flex items-center gap-1.5">
                            <span className="font-mono font-semibold text-primary">{item.docno}</span>
                            {item.bookcode && (
                              <span
                                className={`inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-mono font-bold border ${journalBookBadgeClass(findJournalBook(books, item.bookcode)?.booktype)}`}
                                title={bookLabel(item.bookcode)}
                              >
                                {item.bookcode}
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="truncate py-2 px-2 align-middle" title={item.description}>{item.description}</td>
                        <td className="whitespace-nowrap py-2 px-2 text-right align-middle">
                          <span className={`inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[0.65rem] font-semibold leading-none h-5 align-middle ${
                            isDraft
                              ? "bg-muted text-muted-foreground border border-border"
                              : item.status === "posted"
                                ? "bg-primary/10 text-primary border border-primary/20"
                                : "bg-muted text-muted-foreground border border-border"
                          }`}>
                            {labelText(statusLabel, item.status, tr, tr("gl_check_status", "ตรวจสอบสถานะ"))}
                          </span>
                        </td>
                        <td className="py-2 pr-2.5 pl-2 text-right whitespace-nowrap align-middle" onClick={(e) => e.stopPropagation()}>
                          <div className="inline-flex items-center justify-end gap-1 align-middle">
                            {isDraft && mode === "edit" ? (
                              <>
                                <button
                                  type="button"
                                  className="p-1 rounded text-primary hover:bg-primary/20 transition-colors inline-flex items-center justify-center h-6 w-6"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    void openEdit(item);
                                  }}
                                  aria-label={tr("gl_edit", "แก้ไข")}
                                  title={tr("gl_edit_draft_2", "แก้ไขฉบับร่าง (Edit)")}
                                >
                                  <Pencil className="size-3.5 block" />
                                </button>
                                <button
                                  type="button"
                                  className="p-1 rounded text-destructive hover:bg-destructive/20 transition-colors inline-flex items-center justify-center h-6 w-6"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    void deleteDraftDirect(item);
                                  }}
                                  aria-label={tr("gl_delete", "ลบ")}
                                  title={tr("gl_delete_draft_2", "ลบฉบับร่าง (Delete)")}
                                >
                                  <Trash2 className="size-3.5 block" />
                                </button>
                              </>
                            ) : (
                              <button
                                type="button"
                                className="p-1 rounded text-muted-foreground hover:text-foreground hover:bg-muted/40 transition-colors inline-flex items-center justify-center h-6 w-6"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  void openView(item);
                                }}
                                aria-label={tr("gl_view_data", "แสดงข้อมูล")}
                                title={tr("gl_view", "แสดงข้อมูล (View)")}
                              >
                                <Eye className="size-3.5 block" />
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                  {!list.data.items.length && (
                    <tr>
                      <td colSpan={4} className="p-6 text-center text-muted-foreground">
                        {list.loading ? tr("gl_loading_data", "กำลังโหลดข้อมูล…") : tr("gl_no_entries_in_journal", "ยังไม่มีรายการในสมุดนี้")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
            <Pager page={list.page} total={list.data.total} onPage={list.setPage} loading={list.loading} />
          </div>
        }
        editor={
          journal ? (
            !isEditing ? (
              <div className="flex flex-col h-full min-h-0 gap-3">
                <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
                  <div className="flex items-center gap-2">
                    <div className="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary">
                      <Eye className="size-4" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h2 className="text-base font-semibold">{tr("gl_show_data", "แสดงข้อมูล: {0}").replace("{0}", String(journal.docno))}</h2>
                        <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold bg-muted text-muted-foreground border border-border">
                          {labelText(statusLabel, journal.status, tr)} · {tr("gl_display_mode", tr("gl_display_mode", "โหมดแสดงข้อมูล"))}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground">{journal.description}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <Button type="button" variant="outline" className={actionClass} disabled={busy} onClick={() => voucherPrint.print(<GLVoucherPrint journal={journal} books={refs.books ?? []} company={printCompanyName()} tr={tr} language={language} scale={year?.scale} />)} title={tr("gl_print_voucher_title_hint", "พิมพ์ใบสำคัญรายวันพร้อมช่องลายมือชื่อผู้จัดทำ ผู้ตรวจสอบ และผู้อนุมัติ")}>
                      <Printer className="size-4 mr-1.5" />{tr("gl_print_voucher", "พิมพ์ใบสำคัญ")}
                    </Button>
                    {journal.status === "draft" && mode === "edit" && (
                      <Button
                        type="button"
                        className={actionClass}
                        onClick={() => void openEdit(journal)}
                        disabled={busy}
                        title={tr("gl_edit_this_draft", "แก้ไขฉบับร่างนี้ (Edit)")}
                      >
                        <Pencil className="size-4 mr-1.5" />{tr("gl_edit_draft_3", "แก้ไขฉบับร่าง")}
                      </Button>
                    )}
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-8 rounded-lg text-muted-foreground hover:text-foreground"
                      onClick={() => void closeView()}
                      aria-label={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}
                      title={tr("gl_close_view_dialog", "ปิดหน้าต่างแสดงข้อมูล")}
                    >
                      <X className="size-4" />
                    </Button>
                  </div>
                </header>
                <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground shrink-0">
                  <Eye className="size-3.5 text-primary shrink-0" />
                  <span>
                    {tr("gl_read_only_mode", "โหมดแสดงข้อมูล (Read-only) — {0}").replace("{0}", String(journal.status === "draft" ? tr("gl_edit_draft_instruction", "หากต้องการแก้ไข ให้กดปุ่ม \"แก้ไขฉบับร่าง\"") : tr("gl_posted_cannot_edit_directly", "เอกสารนี้ผ่านรายการแล้ว ไม่สามารถแก้ไขได้โดยตรง")))}
                  </span>
                </div>
                {/* auto-rows-max: ทุกแถวสูงตามเนื้อหา — ตารางเลื่อนแนวนอน (scroll container) ที่เป็นลูกตรงของ grid ซึ่งเลื่อนแนวตั้ง จะถูกบีบแถวเหลือ ~0 แล้วการ์ดยอดรวมทับตาราง */}
                <div className="@container overflow-y-auto p-1 grid min-w-0 auto-rows-max gap-3 content-start">
                  <fieldset disabled className="grid min-w-0 gap-3 opacity-95">
                    <ChampGroupBox title={tr("gl_journal_header_box", "หัวเอกสาร")}>
                      <ChampField label={tr("gl_journal", "สมุดรายวัน")}><input className={control} aria-label={tr("gl_journal", "สมุดรายวัน")} value={journal.bookcode ? `${journal.bookcode} · ${bookLabel(journal.bookcode)}` : ""} readOnly /></ChampField>
                      <ChampSpacer />
                      <ChampField label={tr("gl_document_no", "เลขที่เอกสาร")}><input className={control} aria-label={tr("gl_document_no", "เลขที่เอกสาร")} value={journal.docno} readOnly /></ChampField>
                      <ChampField label={tr("gl_document_date", "วันที่เอกสาร")}><input className={control} aria-label={tr("gl_document_date", "วันที่เอกสาร")} type="date" value={journal.date} readOnly /></ChampField>
                      <ChampField label={tr("gl_reference_document", "เอกสารอ้างอิง")}><input className={control} aria-label={tr("gl_reference_document", "เอกสารอ้างอิง")} value={journal.reference || "-"} readOnly /></ChampField>
                      <ChampField label={tr("gl_ref_date", "วันที่อ้างอิง")}><input className={control} aria-label={tr("gl_ref_date", "วันที่อ้างอิง")} type="date" value={journal.refdate ?? ""} readOnly /></ChampField>
                      <ChampField label={tr("gl_description", "คำอธิบาย")} wide><input className={control} aria-label={tr("gl_description", "คำอธิบาย")} value={journal.description || "-"} readOnly /></ChampField>
                      <ChampField label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} aria-label={tr("gl_branch_code", "รหัสสาขา")} value={journal.branchcode || "-"} readOnly /></ChampField>
                    </ChampGroupBox>
                  </fieldset>
                  {/* กรอบภาษีแสดงอย่างเดียว (แก้ภาษีของใบที่ผ่านรายการแล้วผ่านตารางภาษี) — อยู่นอก fieldset ให้ปุ่มตารางภาษีกดได้ */}
                  <ChampTaxBox details={journal.details} scale={year?.scale ?? 2} disabled onOpenTaxTable={() => openDetailsSection("vats")} />
                  <div className="overflow-x-auto rounded-xl border border-border">
                    <table className="w-full min-w-[900px] text-left text-[0.95rem]">
                      <thead className="bg-muted">
                        <tr>
                          <th className="w-10 p-2 text-center">#</th>
                          <th className="p-2">{tr("gl_account_code", "รหัสบัญชี")}</th>
                          <th className="p-2">{tr("gl_account_name", "ชื่อบัญชี")}</th>
                          <th className="p-2">{tr("gl_extra_detail", "รายละเอียดเพิ่มเติม")}</th>
                          <th className="p-2">{tr("gl_department", "แผนก")}</th>
                          <th className="p-2">{tr("gl_project", "โครงการ")}</th>
                          <th className="p-2 text-right">{tr("gl_debit", "เดบิต")}</th>
                          <th className="p-2 text-right">{tr("gl_credit", "เครดิต")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {journal.lines.map((line, index) => (
                          <tr key={index}>
                            <td className="p-2 text-center tabular-nums text-muted-foreground">{index + 1}</td>
                            <td className="p-2 font-mono font-semibold text-primary whitespace-nowrap">{line.accountcode}</td>
                            <td className="min-w-40 p-2 leading-snug [overflow-wrap:anywhere]">{accountNames.get(line.accountcode) ?? ""}</td>
                            <td className="min-w-40 p-2 leading-snug [overflow-wrap:anywhere]">{line.description || "—"}</td>
                            <td className="p-2">{line.departmentcode || "—"}</td>
                            <td className="p-2">{line.projectcode || "—"}</td>
                            <td className="min-w-32 p-2 text-right font-mono tabular-nums whitespace-nowrap">
                              {line.debit ? formatAmount(line.debit, year?.scale) : "—"}
                            </td>
                            <td className="min-w-32 p-2 text-right font-mono tabular-nums whitespace-nowrap">
                              {line.credit ? formatAmount(line.credit, year?.scale) : "—"}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {(journal.source_system || journal.source_record_id) && <p className="rounded-lg border border-border p-3">{tr("gl_details_ui_34","แหล่งข้อมูล:")}{journal.source_system || "—"}{tr("gl_details_ui_35","· อ้างอิง:")}{journal.source_record_id || "—"}</p>}
                  <div className="grid gap-2 rounded-xl border border-primary/20 bg-primary/5 p-3 sm:grid-cols-3" aria-live="polite">
                    {([ [tr("gl_total_debit", "รวมเดบิต"), totals?.debit], [tr("gl_total_credit", "รวมเครดิต"), totals?.credit], [tr("gl_difference", "ผลต่าง"), totals?.difference] ] as const).map(([label, units]) => (
                      <div key={label}>
                        <div className="text-[0.9rem] text-muted-foreground">{label}</div>
                        <strong className="text-lg tabular-nums">{units === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(units), year?.scale)}</strong>
                      </div>
                    ))}
                  </div>
                  <GLJournalDetailsPanel key={`details:${journal.id ?? "new"}:${journal.version ?? 0}`} value={journal.details} original={original ? (JSON.parse(original) as GLJournal).details : undefined}
                    onChange={details => patch({ details })} lines={journal.lines} accounts={refs.accounts} date={journal.date} branch={journal.branchcode ?? ""} booktype={findJournalBook(books, journal.bookcode)?.booktype} journalId={journal.id}
                    editable={!busy && (isEditing || detailsEditing)} posted={journal.status === "posted"} onBusyChange={setDetailsBusy} scale={year?.scale} />
                  {journal.id && journal.version && <GLJournalReviewPanel key={`${journal.id}:${journal.version}:${reviewLoad}`}
                    journalId={journal.id} version={journal.version} onDirtyChange={setReviewDirty} onBusyChange={setReviewBusy}
                    onReload={() => void openView(journal)} />}
                  {journal.id && journal.status === "draft" && (
                    <Field label={tr("gl_reason", "เหตุผล")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_required_for_draft_deletion", "จำเป็นสำหรับลบฉบับร่าง")} /></Field>
                  )}
                  {journal.status === "posted" && (
                    <div className="grid gap-3 sm:grid-cols-2">
                      <Field label={tr("gl_reversal_document_number", "เลขที่ใบกลับรายการ")}><input className={control} value={reverseDocno} disabled={busy} onChange={(e) => setReverseDocno(e.target.value)} placeholder={tr("gl_enter_reversal_document_number", "ระบุเลขที่ใบกลับรายการ")} /></Field>
                      <Field label={tr("gl_reversal_date", "วันที่กลับรายการ")}><input className={control} type="date" value={reverseDate} disabled={busy} onChange={(e) => setReverseDate(e.target.value)} /></Field>
                    </div>
                  )}
                  {journal.status === "posted" && (
                    <Field label={tr("gl_reversal_reason", "เหตุผลการกระทบยอด / กลับรายการ")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_required_for_reversal_entry", "จำเป็นสำหรับสร้างรายการกลับบัญชี")} /></Field>
                  )}
                </div>
                <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-3 shrink-0 mt-auto">
                  <div className="flex flex-wrap gap-2">
                    {journal.status === "draft" && mode === "edit" && (
                      <Button
                        type="button"
                        className={actionClass}
                        onClick={() => void openEdit(journal)}
                        disabled={busy}
                      >
                        <Pencil className="size-4 mr-1.5" />{tr("gl_edit_draft_3", "แก้ไขฉบับร่าง")}
                      </Button>
                    )}
                    {journal.status === "posted" && (detailsEditing ? <><Button type="button" className={actionClass} disabled={busy || !detailsDirty} onClick={() => void saveReconciliation()}><Save className="mr-1.5 size-4" />{tr("gl_details_ui_36","บันทึกผลกระทบยอด")}</Button><Button type="button" variant="outline" onClick={() => void cancelEdit()}>{tr("gl_details_ui_37","ยกเลิกกระทบยอด")}</Button></> : <Button type="button" variant="outline" disabled={busy || reviewDirty} onClick={() => setDetailsEditing(true)}>{tr("gl_details_ui_38","เพิ่มผลกระทบยอด")}</Button>)}
                    {journal.status === "posted" && (
                      <Button
                        type="button"
                        className={actionClass}
                        disabled={busy}
                        onClick={() => void act("reverse")}
                      >
                        {tr("gl_create_reversing_entry", "สร้างรายการกลับบัญชี")}
                      </Button>
                    )}
                    <Button
                      type="button"
                      variant="outline"
                      className={actionClass}
                      onClick={() => void closeView()}
                    >
                      {tr("gl_close", "ปิด")}
                    </Button>
                  </div>
                  {journal.status === "draft" && journal.id && (
                    <Button
                      type="button"
                      variant="outline"
                      className={`${actionClass} text-destructive border-destructive/30 hover:bg-destructive/10 hover:border-destructive/50`}
                      disabled={busy}
                      onClick={() => void act("delete")}
                    >
                      <Trash2 className="size-4 mr-1.5" />{tr("gl_delete_draft", "ลบฉบับร่าง")}
                    </Button>
                  )}
                </footer>
              </div>
            ) : (
              <form className="flex flex-col h-full min-h-0 gap-3" onSubmit={(event) => { event.preventDefault(); void save(); }}>
                <header className="flex items-center justify-between gap-2 border-b border-border pb-3 shrink-0">
                  <div className="flex min-w-0 items-center gap-2">
                    <div className="grid size-8 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                      {journal.id ? <Pencil className="size-4" /> : <Plus className="size-4" />}
                    </div>
                    <h2 className="truncate text-base font-semibold">
                      {journal.id ? tr("gl_edit_draft", "แก้ไขฉบับร่าง {0}").replace("{0}", String(journal.docno)) : tr("gl_save_new_journal", "บันทึกรายวันใหม่")}
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
                <div className="@container overflow-y-auto p-1 grid min-w-0 gap-3 content-start">
                  <fieldset disabled={busy} className="grid min-w-0 gap-3">
                    {!journal.id && (
                      // ตัวเลือกหลังบันทึกของ Champ (combo บนสุดของจอ) — มีผลกับใบใหม่เท่านั้น ใบที่แก้จะกลับไปโหมดแสดงข้อมูลเหมือนเดิม
                      <label className="flex flex-wrap items-center gap-2 text-[0.95rem]">
                        <span className="font-medium">{tr("gl_save_option", "ตัวเลือกการบันทึก")}</span>
                        <span className="min-w-60 max-w-sm flex-1">
                          <select className={control} value={saveOption} onChange={(e) => changeSaveOption(parseJournalSaveOption(e.target.value))} title={tr("gl_save_option_hint", "ใช้หลังกดบันทึกรายการ F11 — ระบบจำค่าที่เลือกไว้ให้ผู้ใช้นี้ในเครื่องนี้")}>
                            {JOURNAL_SAVE_OPTIONS.map((option) => <option key={option.value} value={option.value}>{tr(option.key, option.th)}</option>)}
                          </select>
                        </span>
                      </label>
                    )}
                    <div className="grid min-w-0 gap-3 @4xl:grid-cols-[minmax(0,1fr)_14rem] @4xl:items-start">
                      {/* กรอบหัวเอกสาร + กรอบภาษีแบบ Champ: Enter เลื่อนช่องตามลำดับ สมุด → เลขที่ → วันที่ → อ้างอิง → วันที่อ้างอิง → คำอธิบาย → สาขา → ภาษี แล้วเข้าตารางบัญชี */}
                      <div ref={headerRef} className="grid min-w-0 gap-3" onKeyDown={(e) => { if (e.key === "Enter") headerEnterNav.onKeyDown(e); }}>
                        <ChampGroupBox title={tr("gl_journal_header_box", "หัวเอกสาร")}>
                          <ChampField label={tr("gl_journal", "สมุดรายวัน")}>
                            <Combobox
                              // backend ห้ามเปลี่ยนสมุดของใบสำคัญที่บันทึกแล้ว (journal_book_immutable) — ล็อกไว้ ไม่ให้ผู้ใช้เลือกแล้วค่อยเจอ error ตอนบันทึก
                              disabled={!isEditing || busy || Boolean(journal.id)}
                              value={journal.bookcode}
                              onChange={(val) => patch({ bookcode: String(val) })}
                              placeholder={tr("gl_journal", "สมุดรายวัน")}
                              aria-label={tr("gl_journal", "สมุดรายวัน")}
                            >
                              {!journal.bookcode && <option value="">{tr("gl_select", "เลือก")}</option>}
                              {usableBooks.map((item) => <option key={item.code} value={item.code}>{item.code} · {journalBookName(item, item.code, language)}</option>)}
                              {journal.bookcode && !usableBooks.some((item) => item.code === journal.bookcode) && (
                                <option value={journal.bookcode}>{journal.bookcode} · {bookLabel(journal.bookcode)}{!journal.id && <> {tr("gl_book_inactive_suffix", "(ปิดใช้งาน — เลือกสมุดอื่น)")}</>}</option>
                              )}
                            </Combobox>
                          </ChampField>
                          <ChampSpacer />
                          <ChampField label={tr("gl_document_no", "เลขที่เอกสาร")} hint={journal.id ? undefined : tr("gl_docno_auto_hint", "เว้นว่างไว้ ระบบจะรันเลขที่ให้ตอนบันทึก")}>
                            <input
                              className={control}
                              aria-label={tr("gl_document_no", "เลขที่เอกสาร")}
                              required={Boolean(journal.id)}
                              value={journal.docno}
                              maxLength={30}
                              placeholder={journal.id ? undefined : nextDocNo ? tr("gl_docno_auto_preview", "อัตโนมัติ: {0}").replace("{0}", nextDocNo) : tr("gl_auto", "อัตโนมัติ")}
                              onChange={(e) => patch({ docno: e.target.value })}
                            />
                          </ChampField>
                          <ChampField label={tr("gl_document_date", "วันที่เอกสาร")}>
                            <input
                              className={control}
                              aria-label={tr("gl_document_date", "วันที่เอกสาร")}
                              required
                              type="date"
                              value={journal.date}
                              onChange={(e) => {
                                const newDate = e.target.value;
                                const matched = fiscalYearForDate(refs.years, newDate);
                                patch({ date: newDate, ...(matched ? { fiscalyear: matched } : {}) });
                              }}
                            />
                          </ChampField>
                          <ChampField label={tr("gl_reference_document", "เอกสารอ้างอิง")}><input className={control} aria-label={tr("gl_reference_document", "เอกสารอ้างอิง")} value={journal.reference} maxLength={50} onChange={(e) => patch({ reference: e.target.value })} /></ChampField>
                          <ChampField label={tr("gl_ref_date", "วันที่อ้างอิง")}><input className={control} aria-label={tr("gl_ref_date", "วันที่อ้างอิง")} type="date" value={journal.refdate ?? ""} onChange={(e) => patch({ refdate: e.target.value })} /></ChampField>
                          <ChampField label={tr("gl_description", "คำอธิบาย")} wide>
                            <input className={control} aria-label={tr("gl_description", "คำอธิบาย")} value={journal.description} onChange={(e) => patch({ description: e.target.value })} maxLength={500} />
                          </ChampField>
                          <ChampField label={tr("gl_branch_code", "รหัสสาขา")}><input className={control} aria-label={tr("gl_branch_code", "รหัสสาขา")} value={journal.branchcode} onChange={(e) => patch({ branchcode: e.target.value })} /></ChampField>
                        </ChampGroupBox>
                        <ChampTaxBox details={journal.details} scale={year?.scale ?? 2} disabled={busy} calcBusy={vatCalcBusy} onTaxType={(value) => void changeTaxType(value)} onPatch={patchTaxBoxRow} onCalc={(mode) => void calcTax(mode)} onOpenTaxTable={() => openDetailsSection("vats")} />
                      </div>
                      {/* ปุ่มคำสั่งด้านขวาตามลำดับของ Champ — แผงแคบจะเรียงเป็นแถวแทน */}
                      <div className="flex flex-wrap gap-2 @4xl:flex-col @4xl:flex-nowrap" role="group" aria-label={tr("gl_journal_commands", "คำสั่งใบสำคัญ")}>
                        <Button type="button" variant="outline" className={champCommandClass} onClick={() => openDetailsSection(null)}>
                          <FileText className="size-4" />{tr("gl_other_details", "รายละเอียดอื่นๆ")}<ChampKey>F5</ChampKey>
                        </Button>
                        <Button type="submit" className={champCommandClass} disabled={busy || !dirty}>
                          <Save className="size-4" />{busy ? tr("gl_saving", "กำลังบันทึก…") : tr("gl_save_entry", "บันทึกรายการ")}<ChampKey primary>F11</ChampKey>
                        </Button>
                        <Button type="button" variant="outline" className={champCommandClass} onClick={focusJournalSearch}>
                          <Search className="size-4" />{tr("gl_pick_entry", "เลือกรายการ")}<ChampKey>F2</ChampKey>
                        </Button>
                        {mode === "edit" && (
                          <Button type="button" variant="outline" className={champCommandClass} onClick={() => void openCreate()}>
                            <Plus className="size-4" />{tr("gl_new_entry", "เพิ่มรายการใหม่")}<ChampKey>Alt+N</ChampKey>
                          </Button>
                        )}
                        <Button type="button" variant="outline" className={`${champCommandClass} text-destructive border-destructive/30 hover:bg-destructive/10 hover:border-destructive/50`} disabled={!journal.id || isJournalDirty} onClick={() => void act("delete")}>
                          <Trash2 className="size-4" />{tr("gl_delete_entry", "ลบรายการ")}
                        </Button>
                        <Button type="button" variant="outline" className={champCommandClass} onClick={() => openDetailsSection("vats")}>
                          <Table2 className="size-4" />{tr("gl_tax_table", "ตารางภาษี")}<ChampKey>F3</ChampKey>
                        </Button>
                        <Button type="button" variant="outline" className={champCommandClass} onClick={() => void cancelEdit()}>
                          <X className="size-4" />{tr("gl_cancel", "ยกเลิก")}<ChampKey>Esc</ChampKey>
                        </Button>
                      </div>
                    </div>
                    <div
                      ref={linesRef}
                      className="overflow-x-auto rounded-xl border border-border"
                      onKeyDown={tabularEnterNav.onKeyDown}
                      onPaste={(e) => {
                        const text = e.clipboardData.getData("text");
                        // One Excel cell ("1,500.00\r\n") pastes into the focused box; only a block of cells becomes lines
                        if (text && isTabularPaste(text)) {
                          e.preventDefault();
                          applyPastedLines(text);
                        }
                      }}
                    >
                      {/* คอลัมน์ตาม Champ: รหัสบัญชี | ชื่อบัญชี | รายละเอียดเพิ่มเติม | แผนก | (โครงการ — มีเฉพาะ BC) | เดบิต | เครดิต */}
                      <table className="w-full min-w-[1040px] text-left text-[0.95rem]">
                        <thead className="bg-muted">
                          <tr>
                            <th className="w-10 p-2 text-center">#</th>
                            <th className="p-2">{tr("gl_account_code", "รหัสบัญชี")}</th>
                            <th className="p-2">{tr("gl_account_name", "ชื่อบัญชี")}</th>
                            <th className="p-2">{tr("gl_extra_detail", "รายละเอียดเพิ่มเติม")}</th>
                            <th className="p-2">{tr("gl_department", "แผนก")}</th>
                            <th className="p-2">{tr("gl_project", "โครงการ")}</th>
                            <th className="p-2">{tr("gl_debit", "เดบิต")}</th>
                            <th className="p-2">{tr("gl_credit", "เครดิต")}</th>
                            <th className="p-2"><span className="sr-only">{tr("gl_manage", "จัดการ")}</span></th>
                          </tr>
                        </thead>
                        <tbody>
                          {journal.lines.map((line, index) => (
                            <tr key={index} className="border-t border-border">
                              <td className="p-2 text-center tabular-nums text-muted-foreground">{index + 1}</td>
                              <td className="min-w-36 p-2">
                                <AccountSelect codeOnly label={tr("gl_line_account", "บัญชีบรรทัด {0}").replace("{0}", String(index + 1))} value={line.accountcode} accounts={refs.accounts} onChange={(accountcode) => patchLine(index, { accountcode })} disabled={busy} />
                              </td>
                              <td className="min-w-40 p-2 leading-snug [overflow-wrap:anywhere]">{line.accountcode ? accountNames.get(line.accountcode) ?? "" : ""}</td>
                              <td className="min-w-40 p-2">
                                <input className={control} aria-label={tr("gl_line_extra_detail", "รายละเอียดเพิ่มเติมบรรทัด {0}").replace("{0}", String(index + 1))} value={line.description} onChange={(e) => patchLine(index, { description: e.target.value })} />
                              </td>
                              <td className="min-w-24 p-2">
                                <input className={control} aria-label={tr("gl_line_department", "แผนกบรรทัด {0}").replace("{0}", String(index + 1))} value={line.departmentcode} disabled={busy || departLocked(line)} title={departLocked(line) ? tr("gl_line_not_depart_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกแผนก (ตั้งที่ผังบัญชี › เงื่อนไข)") : undefined} onChange={(e) => patchLine(index, { departmentcode: e.target.value })} />
                              </td>
                              <td className="min-w-24 p-2">
                                <input className={control} aria-label={tr("gl_line_project", "โครงการบรรทัด {0}").replace("{0}", String(index + 1))} value={line.projectcode} disabled={busy || projectLocked(line)} title={projectLocked(line) ? tr("gl_line_not_project_separate", "บัญชีนี้ไม่ได้กำหนดให้แยกโครงการ (ตั้งที่ผังบัญชี › เงื่อนไข)") : undefined} onChange={(e) => patchLine(index, { projectcode: e.target.value })} />
                              </td>
                              <td className="min-w-32 p-2">
                                <AmountInput ariaLabel={tr("gl_line_debit", "เดบิตบรรทัด {0}").replace("{0}", String(index + 1))} disabled={busy} scale={year?.scale ?? 2} allowEmpty={false} value={line.debit} onChange={(debit) => patchLine(index, { debit })} />
                              </td>
                              <td className="min-w-32 p-2">
                                <AmountInput ariaLabel={tr("gl_line_credit", "เครดิตบรรทัด {0}").replace("{0}", String(index + 1))} disabled={busy} scale={year?.scale ?? 2} allowEmpty={false} value={line.credit} onChange={(credit) => patchLine(index, { credit })} />
                              </td>
                              <td className="p-2">
                                <Button type="button" variant="outline" className={actionClass} disabled={busy || journal.lines.length <= 2} onClick={() => { if ((journal.details?.allocations?.length ?? 0) + (journal.details?.bank_lines?.length ?? 0) + (journal.details?.matches?.length ?? 0) > 0) { setError(tr("gl_details_ui_39","นำการจัดสรรและจับคู่ออกก่อนลบบรรทัดบัญชี เพื่อป้องกันการอ้างอิงผิดแถว")); return; } patch({ lines: journal.lines.filter((_, i) => i !== index) }); }}>
                                  {tr("gl_remove", "นำออก")}
                                </Button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <Button type="button" variant="outline" className={actionClass} disabled={journal.lines.length >= 500} onClick={() => addNewLine()}>
                        <Plus className="size-4 mr-1.5" />{tr("gl_add_line", "เพิ่มบรรทัด")}
                      </Button>
                      <Button type="button" variant="outline" className={actionClass} disabled={journal.lines.length >= 500} onClick={() => void handlePasteClick()} title={tr("gl_paste_from_excel_hint", "วางแถวที่คัดลอกจาก Excel หรือ Google Sheets (Ctrl+V) คอลัมน์: รหัสบัญชี คำอธิบาย เดบิต เครดิต แผนก โครงการ")}>
                        <ClipboardPaste className="size-4 mr-1.5" />
                        <span>{tr("gl_paste_from_excel", "วางจาก Excel")}</span>
                      </Button>
                      {smartGuard && !smartGuard.isBalanced && (
                        <Button
                          type="button"
                          variant="outline"
                          className={`${actionClass} border-primary/40 text-primary bg-primary/5 hover:bg-primary/10 shadow-sm`}
                          onClick={() => patch({ lines: autoBalanceJournalLines(journal.lines, year?.scale ?? 2) })}
                          title={tr("gl_auto_balance_hint", "ปรับยอดให้เดบิตและเครดิตสมดุลกันอัตโนมัติ")}
                        >
                          <Scale className="size-4 mr-1.5 text-primary" />
                          <span>{tr("gl_auto_balance", "ปรับยอดให้ดุล (Auto-Balance)")}</span>
                        </Button>
                      )}
                    </div>
                  </fieldset>

                  {/* Totals & Balance Guard Card */}
                  <div
                    className={`grid gap-2 rounded-xl border p-3 sm:grid-cols-3 transition-colors ${
                      smartGuard?.isBalanced
                        ? "border-primary/30 bg-primary/5"
                        : "border-destructive/30 bg-destructive/5"
                    }`}
                    aria-live="polite"
                  >
                    <div>
                      <div className="text-[0.9rem] text-muted-foreground">{tr("gl_total_debit", "รวมเดบิต")}</div>
                      <strong className="text-lg tabular-nums">
                        {totals?.debit === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(totals.debit), year?.scale)}
                      </strong>
                    </div>
                    <div>
                      <div className="text-[0.9rem] text-muted-foreground">{tr("gl_total_credit", "รวมเครดิต")}</div>
                      <strong className="text-lg tabular-nums">
                        {totals?.credit === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(totals.credit), year?.scale)}
                      </strong>
                    </div>
                    <div>
                      <div className="flex items-center justify-between gap-1">
                        <span className="text-[0.9rem] text-muted-foreground">{tr("gl_difference", "ผลต่าง")}</span>
                        {smartGuard?.isBalanced ? (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-primary bg-primary/10 px-2 py-0.5 rounded-full">
                            <CheckCircle2 className="size-3" />
                            {tr("gl_balanced_100", "สมดุล 100%")}
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-destructive bg-destructive/10 px-2 py-0.5 rounded-full">
                            <AlertTriangle className="size-3" />
                            {smartGuard?.balanceStatus === "debit_surplus"
                              ? tr("gl_short_credit", "ขาดเครดิต")
                              : tr("gl_short_debit", "ขาดเดบิต")}
                          </span>
                        )}
                      </div>
                      <strong className={`text-lg tabular-nums ${smartGuard?.isBalanced ? "text-primary" : "text-destructive"}`}>
                        {totals?.difference === undefined ? tr("gl_verify_amount", "ตรวจจำนวนเงิน") : formatAmount(amountString(totals.difference), year?.scale)}
                      </strong>
                      {!smartGuard?.isBalanced && smartGuard && (
                        <div className="text-[11px] text-destructive/90 mt-0.5 font-medium">
                          {smartGuard.balanceStatus === "debit_surplus"
                            ? tr("gl_debit_over_credit", "เดบิตมากกว่าเครดิต {0}").replace("{0}", smartGuard.balanceDifferenceFormatted)
                            : tr("gl_credit_over_debit", "เครดิตมากกว่าเดบิต {0}").replace("{0}", smartGuard.balanceDifferenceFormatted)}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* VAT 7% Smart Checker Card */}
                  {smartGuard?.vat.hasVatLine && (
                    <div
                      className={`flex flex-wrap items-center justify-between gap-2 p-2.5 rounded-xl border text-xs ${
                        smartGuard.vat.isExactVat
                          ? "bg-primary/5 border-primary/20 text-foreground"
                          : smartGuard.vat.isCloseVat
                          ? "bg-muted border-border text-foreground"
                          : "bg-destructive/5 border-destructive/30 text-foreground"
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        {smartGuard.vat.isExactVat ? (
                          <CheckCircle2 className="size-4 shrink-0 text-primary" />
                        ) : (
                          <AlertTriangle className={`size-4 shrink-0 ${smartGuard.vat.isCloseVat ? "text-muted-foreground" : "text-destructive"}`} />
                        )}
                        <span>
                          {smartGuard.vat.isExactVat
                            ? tr("gl_vat_verified_exact", "ตรวจสอบภาษีมูลค่าเพิ่ม 7%: ยอด {0} ตรงตามฐานคำนวณเป๊ะ").replace("{0}", smartGuard.vat.actualVatFormatted)
                            : smartGuard.vat.isCloseVat
                            ? tr("gl_vat_close_satang", "ภาษีมูลค่าเพิ่มในเอกสาร {0} (ต่างจาก 7% คำนวณปกติ {1} อยู่ {2} สตางค์ — สรรพากรยอมรับได้ตามใบกำกับภาษีจริง)")
                                .replace("{0}", smartGuard.vat.actualVatFormatted)
                                .replace("{1}", smartGuard.vat.expectedVatFormatted)
                                .replace("{2}", formatAmount(amountString(smartGuard.vat.varianceUnits < 0n ? -smartGuard.vat.varianceUnits : smartGuard.vat.varianceUnits), 2))
                            : tr("gl_vat_mismatch", "ภาษีมูลค่าเพิ่มในเอกสาร {0} ต่างจาก 7% ของฐานภาษี ({1}) โปรดตรวจสอบ")
                                .replace("{0}", smartGuard.vat.actualVatFormatted)
                                .replace("{1}", smartGuard.vat.expectedVatFormatted)}
                        </span>
                      </div>
                      {!smartGuard.vat.isExactVat && (
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          className="h-6 px-2.5 text-[11px] font-medium bg-background border-border hover:bg-muted"
                          onClick={() => patch({ lines: setExactVatLine(journal.lines, smartGuard.vat.vatLineIndex, smartGuard.vat.expectedVatUnits, year?.scale ?? 2) })}
                        >
                          {tr("gl_adjust_exact_vat", "ปรับเป็น 7% พอดี ({0})").replace("{0}", smartGuard.vat.expectedVatFormatted)}
                        </Button>
                      )}
                    </div>
                  )}
                  <GLJournalDetailsPanel key={`details:${journal.id ?? "new"}:${journal.version ?? 0}`} value={journal.details} original={original ? (JSON.parse(original) as GLJournal).details : undefined}
                    onChange={details => patch({ details })} lines={journal.lines} accounts={refs.accounts} date={journal.date} branch={journal.branchcode ?? ""} booktype={findJournalBook(books, journal.bookcode)?.booktype} journalId={journal.id}
                    editable={!busy && (isEditing || detailsEditing)} posted={journal.status === "posted"} onBusyChange={setDetailsBusy} scale={year?.scale} />
                  {journal.id && (
                    <Field label={tr("gl_reason", "เหตุผล")}><input className={control} value={reason} disabled={busy} onChange={(e) => setReason(e.target.value)} placeholder={tr("gl_edit_reason_hint", "ระบุเหตุผลการแก้ไข (ถ้ามี)")} /></Field>
                  )}
                </div>
              </form>
            )
          ) : (
            <div className="flex flex-col flex-1 h-full min-h-64 items-center justify-center content-center gap-3 text-center p-6">
              <div className="mx-auto grid size-12 place-items-center rounded-2xl bg-muted text-muted-foreground">
                <FileText className="size-6" />
              </div>
              <h2 className="text-base font-semibold">{tr("gl_select_item_show_acct", "เลือกรายการเพื่อแสดงข้อมูลบัญชี")}</h2>
              <p className="text-sm text-muted-foreground">{tr("gl_click_row_or_add_journal", "คลิกที่แถวในตารางเพื่อแสดงข้อมูล หรือกดปุ่ม “+ เพิ่มรายการ” เพื่อบันทึกรายวันใหม่")}</p>
              {mode === "edit" && (
                <div className="pt-2">
                  <Button type="button" className={actionClass} onClick={() => void openCreate()} disabled={busy}>
                    <Plus className="size-4 mr-1.5" />{tr("gl_save_new_journal", "บันทึกรายวันใหม่")}
                  </Button>
                </div>
              )}
            </div>
          )
        }
      />
      {confirmationDialog}
      {voucherPrint.portal}
    </div>
  );
}
