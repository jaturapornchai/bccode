"use client";

import { useGLText } from "./gl-common";
import { useState, useEffect, useRef, useMemo, useCallback } from "react";
import {
  Search,
  X,
  Maximize2,
  Minimize2,
  Check,
  BookOpen,
  FolderOpen,
  ShieldCheck,
  CheckSquare,
  Square,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/combobox";
import { Checkbox } from "@/components/ui/checkbox";
import {
  type GLAccount,
  accountName,
  accountTypeLabels,
  accountTypes,
  labelText,
  type GLLabel,
  sortAccountsHierarchically,
} from "@/lib/general-ledger";


export interface AccountSearchDialogProps {
  open: boolean;
  onClose: () => void;
  onSelect?: (accountCode: string, account: GLAccount) => void;
  onSelectMultiple?: (accountCodes: string[]) => void;
  accounts: GLAccount[];
  selectedCode?: string;
  selectedCodes?: string[];
  multiSelect?: boolean;
  all?: boolean; // if false, defaults to showing only allowposting && isactive
  title?: string;
  allowEmpty?: boolean;
}

const CATEGORY_TABS: { key: string; label: GLLabel; type?: GLAccount["accounttype"] }[] = [
  { key: "all", label: ["gl_all", "ทั้งหมด"] },
  { key: "asset", label: ["gl_1_assets", "1. สินทรัพย์"], type: "asset" },
  { key: "liability", label: ["gl_2_liabilities", "2. หนี้สิน"], type: "liability" },
  { key: "equity", label: ["gl_3_equity", "3. ส่วนของเจ้าของ"], type: "equity" },
  { key: "income", label: ["gl_4_revenue", "4. รายได้"], type: "income" },
  { key: "expense", label: ["gl_5_expenses", "5. ค่าใช้จ่าย"], type: "expense" },
];

export function AccountSearchDialog({
  open,
  onClose,
  onSelect,
  onSelectMultiple,
  accounts = [],
  selectedCode = "",
  selectedCodes = [],
  multiSelect = false,
  all = false,
  title: titleProp,
  allowEmpty = true,
}: AccountSearchDialogProps) {
  const tr = useGLText();
  const title = titleProp ?? tr("gl_search_select_coa", "ค้นหาและเลือกผังบัญชี (Chart of Accounts)");
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string>("all");
  const [onlyPosting, setOnlyPosting] = useState<boolean>(false);
  const [levelFilter, setLevelFilter] = useState<string>("all");
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(0);

  // Multi-select internal state
  const [multiChecked, setMultiChecked] = useState<Set<string>>(
    () => new Set(selectedCodes.length > 0 ? selectedCodes : selectedCode ? [selectedCode] : [])
  );

  const searchInputRef = useRef<HTMLInputElement>(null);
  const tableContainerRef = useRef<HTMLDivElement>(null);
  const rowRefs = useRef<(HTMLTableRowElement | null)[]>([]);

  // Initialize on open only (false → true). Depending on selectedCodes alone re-ran this on every render
  // because the default `[]` is a new array each time, wiping the search box on each keystroke.
  const wasOpenRef = useRef(false);
  useEffect(() => {
    const justOpened = open && !wasOpenRef.current;
    wasOpenRef.current = open;
    if (justOpened) {
      setSearch("");
      setCategory("all");
      setOnlyPosting(false);
      setLevelFilter("all");
      setHighlightedIndex(0);

      if (multiSelect) {
        setMultiChecked(new Set(selectedCodes.length > 0 ? selectedCodes : selectedCode ? [selectedCode] : []));
      }

      setTimeout(() => {
        searchInputRef.current?.focus();
        searchInputRef.current?.select();
      }, 50);
    }
  }, [open, multiSelect, selectedCodes, selectedCode]);

  // Hierarchical sort with calculated effective levels
  const hierarchicallySortedAccounts = useMemo(() => {
    return sortAccountsHierarchically(accounts);
  }, [accounts]);

  // Account category counts
  const categoryCounts = useMemo(() => {
    const counts: Record<string, number> = { all: hierarchicallySortedAccounts.length };
    for (const acc of hierarchicallySortedAccounts) {
      counts[acc.accounttype] = (counts[acc.accounttype] || 0) + 1;
    }
    return counts;
  }, [hierarchicallySortedAccounts]);

  // Filter accounts while preserving hierarchy order
  const filteredAccounts = useMemo(() => {
    const query = search.trim().toLowerCase();
    return hierarchicallySortedAccounts.filter((acc) => {
      // Category filter
      if (category !== "all" && acc.accounttype !== category) {
        return false;
      }

      // Posting filter
      if (onlyPosting && (!acc.allowposting || !acc.isactive)) {
        return false;
      }

      // Level filter
      const level = acc.effectiveLevel;
      if (levelFilter === "level-1" && level !== 1) return false;
      if (levelFilter === "level-sub" && level <= 1) return false;
      if (levelFilter.startsWith("lvl-") && level !== parseInt(levelFilter.replace("lvl-", ""), 10)) return false;

      // Text search query
      if (!query) return true;

      const codeMatch = acc.accountcode.toLowerCase().includes(query);
      if (codeMatch) return true;

      const nameMatch = acc.names?.some((n) => n.name.toLowerCase().includes(query));
      if (nameMatch) return true;

      const typeThai = accountTypes[acc.accounttype] ?? "";
      if (typeThai.toLowerCase().includes(query)) return true;

      return false;
    });
  }, [hierarchicallySortedAccounts, search, category, onlyPosting, levelFilter]);


  // Reset highlighted index when filtered list changes
  useEffect(() => {
    if (filteredAccounts.length === 0) {
      setHighlightedIndex(-1);
    } else {
      const targetCode = multiSelect ? undefined : selectedCode;
      const selIdx = targetCode ? filteredAccounts.findIndex((a) => a.accountcode === targetCode) : 0;
      setHighlightedIndex(selIdx >= 0 ? selIdx : 0);
    }
  }, [filteredAccounts, selectedCode, multiSelect]);

  // Scroll highlighted row into view
  useEffect(() => {
    if (highlightedIndex >= 0 && rowRefs.current[highlightedIndex]) {
      rowRefs.current[highlightedIndex]?.scrollIntoView({
        block: "nearest",
        behavior: "smooth",
      });
    }
  }, [highlightedIndex]);

  const selectSingleAccount = useCallback(
    (account: GLAccount) => {
      if (!all && !account.allowposting) {
        return;
      }
      if (onSelect) {
        onSelect(account.accountcode, account);
      }
      onClose();
    },
    [all, onSelect, onClose]
  );

  const toggleMultiCheck = useCallback((accountCode: string) => {
    setMultiChecked((prev) => {
      const next = new Set(prev);
      if (next.has(accountCode)) {
        next.delete(accountCode);
      } else {
        next.add(accountCode);
      }
      return next;
    });
  }, []);

  const confirmMultiSelect = useCallback(() => {
    if (onSelectMultiple) {
      onSelectMultiple(Array.from(multiChecked));
    }
    onClose();
  }, [onSelectMultiple, multiChecked, onClose]);

  // Keyboard navigation
  useEffect(() => {
    if (!open) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
        return;
      }

      if (filteredAccounts.length === 0) return;

      if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIndex((prev) => (prev < filteredAccounts.length - 1 ? prev + 1 : prev));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : prev));
      } else if (e.key === "Enter") {
        e.preventDefault();
        if (highlightedIndex >= 0 && highlightedIndex < filteredAccounts.length) {
          const acc = filteredAccounts[highlightedIndex];
          if (multiSelect) {
            toggleMultiCheck(acc.accountcode);
          } else if (all || acc.allowposting) {
            selectSingleAccount(acc);
          }
        }
      } else if (e.key === " " && multiSelect && document.activeElement !== searchInputRef.current) {
        // Space to toggle in multi-select mode if not typing in search box
        e.preventDefault();
        if (highlightedIndex >= 0 && highlightedIndex < filteredAccounts.length) {
          toggleMultiCheck(filteredAccounts[highlightedIndex].accountcode);
        }
      } else if (e.key === "PageDown") {
        e.preventDefault();
        setHighlightedIndex((prev) => Math.min(filteredAccounts.length - 1, prev + 8));
      } else if (e.key === "PageUp") {
        e.preventDefault();
        setHighlightedIndex((prev) => Math.max(0, prev - 8));
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, filteredAccounts, highlightedIndex, multiSelect, selectSingleAccount, toggleMultiCheck, onClose]);

  if (!open) return null;

  const currentHighlighted =
    highlightedIndex >= 0 && highlightedIndex < filteredAccounts.length
      ? filteredAccounts[highlightedIndex]
      : null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-2 sm:p-4 backdrop-blur-xs text-foreground animate-in fade-in duration-150"
      role="dialog"
      aria-modal="true"
      aria-labelledby="account-search-dialog-title"
      onPointerDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className={`flex flex-col bg-card border border-border shadow-2xl transition-all duration-200 overflow-hidden ${
          isFullscreen
            ? "fixed inset-0 w-full h-full rounded-none"
            : "w-full max-w-6xl h-[92vh] max-h-[920px] rounded-2xl"
        }`}
      >
        {/* Top Header */}
        <header className="flex items-center justify-between gap-3 border-b border-border px-4 py-3 bg-muted/30">
          <div className="flex items-center gap-3 min-w-0">
            <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <BookOpen className="size-5" />
            </span>
            <div className="min-w-0">
              <h2
                id="account-search-dialog-title"
                className="text-lg font-bold tracking-tight truncate text-foreground"
              >
                {title}
              </h2>
              <p className="text-xs text-muted-foreground flex items-center gap-2">
                <span>
                  {(() => { const [before, after = ""] = tr("gl_found_x_of_y_accounts", "พบ {0} จาก {1} บัญชี").split("{0}"); return <>{before}<strong className="text-primary font-semibold">{filteredAccounts.length}</strong>{after.replace("{1}", String(accounts.length))}</>; })()}
                </span>
                {multiSelect && (
                  <span className="rounded-md bg-primary/10 px-2 py-0.5 text-[11px] font-semibold text-primary">
                    {tr("gl_selected_accounts_2", "เลือกอยู่ {0} บัญชี").replace("{0}", String(multiChecked.size))}
                  </span>
                )}
                {onlyPosting && (
                  <span className="hidden sm:inline-block rounded-md bg-muted px-1.5 py-0.5 text-[11px] font-medium text-muted-foreground">
                    {tr("gl_postable_accounts_only", "เฉพาะบัญชีลงรายการ")}
                  </span>
                )}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-1.5">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-9 rounded-xl hover:bg-muted"
              onClick={() => setIsFullscreen(!isFullscreen)}
              aria-label={isFullscreen ? tr("gl_minimize_window", "ย่อหน้าต่างลง") : tr("gl_maximize_fullscreen", "ขยายเต็มจอ")}
              title={isFullscreen ? tr("gl_minimize_window", "ย่อหน้าต่างลง") : tr("gl_maximize_fullscreen", "ขยายเต็มจอ")}
            >
              {isFullscreen ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-9 rounded-xl hover:bg-destructive/10 hover:text-destructive"
              onClick={onClose}
              aria-label={tr("gl_close_window", "ปิดหน้าต่าง (Esc)")}
              title={tr("gl_close_window", "ปิดหน้าต่าง (Esc)")}
            >
              <X className="size-5" />
            </Button>
          </div>
        </header>

        {/* Search Bar & Primary Actions */}
        <div className="border-b border-border bg-background p-3 sm:p-4 space-y-3">
          <div className="relative flex items-center">
            <Search className="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 size-5 text-muted-foreground z-10" aria-hidden="true" />
            <input
              ref={searchInputRef}
              type="text"
              className="w-full rounded-xl border border-input bg-card py-2.5 !pl-11 !pr-10 text-base sm:text-lg text-foreground placeholder:text-muted-foreground/80 focus:border-primary focus:outline-hidden focus:ring-2 focus:ring-primary/25 transition-all"
              placeholder={tr("gl_search_account_hint", "พิมพ์ค้นหารหัสบัญชี เช่น 1101, ชื่อบัญชี เช่น เงินสด, เงินฝาก, ลูกหนี้... (↑ ↓ เลื่อน, Enter เลือก)")}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            {search && (
              <button
                type="button"
                tabIndex={-1}
                className="absolute right-3 top-1/2 -translate-y-1/2 z-10 grid size-7 place-items-center rounded-full hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                onClick={() => {
                  setSearch("");
                  searchInputRef.current?.focus();
                }}
                title={tr("gl_clear_search", "ล้างคำค้นหา")}
              >
                <X className="size-4" />
              </button>
            )}
          </div>

          {/* Category Filter Pills & Options */}
          <div className="flex flex-wrap items-center justify-between gap-2 pt-1">
            {/* 5 Categories Pills */}
            <div className="flex flex-wrap items-center gap-1.5 overflow-x-auto pb-1 max-w-full">
              {CATEGORY_TABS.map((tab) => {
                const active = category === tab.key;
                const count = categoryCounts[tab.key] ?? (tab.key === "all" ? accounts.length : 0);
                return (
                  <button
                    key={tab.key}
                    type="button"
                    onClick={() => setCategory(tab.key)}
                    className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs sm:text-sm font-medium transition-all ${
                      active
                        ? "bg-primary text-primary-foreground shadow-sm font-semibold"
                        : "bg-muted/70 text-muted-foreground hover:bg-muted hover:text-foreground"
                    }`}
                  >
                    <span>{tr(...tab.label)}</span>
                    <span
                      className={`rounded-full px-1.5 py-0.2 text-[11px] ${
                        active ? "bg-primary-foreground/20 text-primary-foreground" : "bg-background/80 text-muted-foreground"
                      }`}
                    >
                      {count}
                    </span>
                  </button>
                );
              })}
            </div>

            {/* Sub Filter Toggles */}
            <div className="flex flex-wrap items-center gap-2 min-w-0">
              <label className="flex items-center gap-2 cursor-pointer text-xs sm:text-sm text-muted-foreground hover:text-foreground select-none">
                <Checkbox
                  checked={onlyPosting}
                  onCheckedChange={(checked) => setOnlyPosting(checked)}
                />
                <span>{tr("gl_postable_accounts_only", "เฉพาะบัญชีลงรายการ")}</span>
              </label>

              <Combobox
                value={levelFilter}
                onChange={(val) => setLevelFilter(String(val))}
                aria-label={tr("gl_all_account_levels", "ทุกระดับบัญชี")}
                className="w-48 text-xs sm:text-sm"
                buttonClassName="min-h-[2.2em] py-1 text-xs sm:text-sm"
              >
                <option value="all">{tr("gl_all_account_levels", "ทุกระดับบัญชี")}</option>
                <option value="level-1">{tr("gl_level1_control_account", "ระดับ 1 (บัญชีคุมหลัก)")}</option>
                <option value="level-sub">{tr("gl_level2_up_sub_account", "ระดับ 2 ขึ้นไป (บัญชีย่อย)")}</option>
                <option value="lvl-2">{tr("gl_only_level2", "เฉพาะระดับ 2")}</option>
                <option value="lvl-3">{tr("gl_only_level3", "เฉพาะระดับ 3")}</option>
                <option value="lvl-4">{tr("gl_only_level4", "เฉพาะระดับ 4")}</option>
              </Combobox>
            </div>
          </div>
        </div>

        {/* Table Content */}
        <div ref={tableContainerRef} className="flex-1 overflow-auto p-2 sm:p-4">
          {filteredAccounts.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-center p-6 rounded-2xl border border-dashed border-border bg-muted/10">
              <span className="grid size-12 place-items-center rounded-2xl bg-muted text-muted-foreground mb-3">
                <Search className="size-6" />
              </span>
              <p className="text-base font-semibold text-foreground">{tr("gl_no_coa_match_criteria", "ไม่พบผังบัญชีที่ตรงกับเงื่อนไข")}</p>
              <p className="text-sm text-muted-foreground mt-1 max-w-sm">
                {tr("gl_try_spelling_category_posting_only", "ลองตรวจสอบตัวสะกด หรือเปลี่ยนหมวดบัญชี หรือปิดตัวเลือก \"เฉพาะบัญชีลงรายการ\"")}
              </p>
              {(search || category !== "all" || !onlyPosting || levelFilter !== "all") && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="mt-4"
                  onClick={() => {
                    setSearch("");
                    setCategory("all");
                    setOnlyPosting(false);
                    setLevelFilter("all");
                  }}
                >
                  {tr("gl_clear_all_filters", "ล้างตัวกรองทั้งหมด")}
                </Button>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto rounded-xl border border-border bg-card">
              <table className="w-full border-collapse text-left text-sm leading-normal">
                <thead>
                  <tr className="border-b border-border bg-muted/60 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                    {multiSelect && <th className="py-2.5 px-3 w-12 text-center">{tr("gl_select", "เลือก")}</th>}
                    <th className="py-2.5 px-3 w-36">{tr("gl_account_code", "รหัสบัญชี")}</th>
                    <th className="py-2.5 px-3 min-w-[240px]">{tr("gl_account_name", "ชื่อบัญชี")}</th>
                    <th className="py-2.5 px-3 w-32">{tr("gl_account_category", "หมวดบัญชี")}</th>
                    <th className="py-2.5 px-3 w-24 text-center">{tr("gl_level", "ระดับ")}</th>
                    <th className="py-2.5 px-3 w-24 text-center">{tr("gl_normal_side", "ด้านปกติ")}</th>
                    <th className="py-2.5 px-3 w-28 text-center">{tr("gl_posting_rights", "สิทธิ์ลงรายการ")}</th>
                    <th className="py-2.5 px-3 w-28 text-right">{tr("gl_process_2", "ดำเนินการ")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredAccounts.map((acc, index) => {
                    const isHighlighted = highlightedIndex === index;
                    const isChecked = multiChecked.has(acc.accountcode);
                    const isSelected = !multiSelect && acc.accountcode === selectedCode;
                    const level = acc.effectiveLevel ?? acc.level ?? 1;
                    const name = accountName(acc);
                    const engName = acc.names?.find((n) => n.code === "en")?.name;

                    // Type color & badge
                    const typeBadgeColors: Record<string, string> = {
                      asset: "bg-primary/10 text-primary border-primary/20",
                      liability: "bg-muted text-foreground border-border",
                      equity: "bg-primary/15 text-primary border-primary/30",
                      income: "bg-primary/10 text-primary border-primary/20",
                      expense: "bg-destructive/10 text-destructive border-destructive/20",
                    };

                    return (
                      <tr
                        key={acc.id ?? acc.accountcode}
                        ref={(el) => {
                          rowRefs.current[index] = el;
                        }}
                        onClick={() => {
                          setHighlightedIndex(index);
                          if (multiSelect) toggleMultiCheck(acc.accountcode);
                        }}
                        onDoubleClick={() => {
                          if (multiSelect) {
                            toggleMultiCheck(acc.accountcode);
                          } else if (all || acc.allowposting) {
                            selectSingleAccount(acc);
                          }
                        }}
                        className={`group cursor-pointer transition-colors ${
                          isHighlighted
                            ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
                            : isChecked || isSelected
                            ? "bg-muted/70"
                            : "hover:bg-muted/40"
                        }`}
                      >
                        {/* Multi-select checkbox */}
                        {multiSelect && (
                          <td className="py-2.5 px-3 text-center" onClick={(e) => e.stopPropagation()}>
                            <Checkbox
                              checked={isChecked}
                              onCheckedChange={() => toggleMultiCheck(acc.accountcode)}
                            />
                          </td>
                        )}

                        {/* Account Code */}
                        <td className="py-2.5 px-3 font-mono font-bold text-base text-primary whitespace-nowrap">
                          {acc.accountcode}
                        </td>

                        {/* Account Name with Indent Tree */}
                        <td className="py-2.5 px-3">
                          <div
                            className="flex items-center gap-1.5"
                            style={{ paddingLeft: `${Math.max(0, level - 1) * 18}px` }}
                          >
                            {level > 1 && (
                              <span className="font-mono text-xs text-muted-foreground select-none shrink-0">
                                └─
                              </span>
                            )}
                            <div>
                              <div className="text-foreground text-[0.95rem] font-medium leading-tight">
                                {name}
                              </div>
                              {engName && engName !== name && (
                                <div className="text-xs text-muted-foreground font-normal">
                                  {engName}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>

                        {/* Category */}
                        <td className="py-2.5 px-3 whitespace-nowrap">
                          <span
                            className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium border ${
                              typeBadgeColors[acc.accounttype] ?? "bg-muted text-muted-foreground border-border"
                            }`}
                          >
                            {labelText(accountTypeLabels, acc.accounttype, tr)}
                          </span>
                        </td>

                        {/* Level */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          <span className="inline-flex items-center justify-center rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">
                            {tr("gl_level_2", "ระดับ {0}").replace("{0}", String(level))}
                          </span>
                        </td>

                        {/* Normal Balance */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap text-xs text-muted-foreground">
                          {acc.normalbalance === "debit" ? (
                            <span className="text-primary font-medium">{tr("gl_debit", "เดบิต")}</span>
                          ) : (
                            <span className="text-muted-foreground font-medium">{tr("gl_credit", "เครดิต")}</span>
                          )}
                        </td>

                        {/* Posting Status */}
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          {acc.allowposting ? (
                            <span className="inline-flex items-center gap-1 text-xs text-primary font-medium">
                              <ShieldCheck className="size-3.5" />
                              {tr("gl_post_entry", "ลงรายการ")}
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
                              <FolderOpen className="size-3.5" />
                              {tr("gl_control_account", "บัญชีคุม")}
                            </span>
                          )}
                        </td>

                        {/* Action Select Button */}
                        <td className="py-2.5 px-3 text-right whitespace-nowrap">
                          {multiSelect ? (
                            <Button
                              type="button"
                              size="sm"
                              variant={isChecked ? "default" : "outline"}
                              className="h-8 px-2.5 text-xs font-medium rounded-lg"
                              onClick={(e) => {
                                e.stopPropagation();
                                toggleMultiCheck(acc.accountcode);
                              }}
                            >
                              {isChecked ? (
                                <span className="flex items-center gap-1">
                                  <Check className="size-3.5" /> {tr("gl_selected", "เลือกแล้ว")}
                                </span>
                              ) : (
                                tr("gl_select", "เลือก")
                              )}
                            </Button>
                          ) : !all && !acc.allowposting ? (
                            <span className="inline-flex items-center gap-1 rounded-md bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground select-none">
                              <FolderOpen className="size-3" /> {tr("gl_control_account", "บัญชีคุม")}
                            </span>
                          ) : (
                            <Button
                              type="button"
                              size="sm"
                              variant={isSelected ? "default" : isHighlighted ? "default" : "outline"}
                              className="h-8 px-3 text-xs font-medium rounded-lg"
                              onClick={(e) => {
                                e.stopPropagation();
                                selectSingleAccount(acc);
                              }}
                            >
                              {isSelected ? (
                                <span className="flex items-center gap-1">
                                  <Check className="size-3.5" /> {tr("gl_this_account", "บัญชีนี้")}
                                </span>
                              ) : (
                                tr("gl_select", "เลือก")
                              )}
                            </Button>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Footer Bar */}
        <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-border px-4 py-3 bg-muted/20">
          <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[11px] font-semibold text-foreground">
                ↑
              </kbd>
              <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[11px] font-semibold text-foreground">
                ↓
              </kbd>
              <span>{tr("gl_move_row", "เลื่อนแถว")}</span>
            </span>
            <span className="flex items-center gap-1">
              <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[11px] font-semibold text-foreground">
                Enter
              </kbd>
              <span>{multiSelect ? tr("gl_toggle_selection", "สลับเลือก") : tr("gl_select_account", "เลือกบัญชี")}</span>
            </span>
            <span className="flex items-center gap-1">
              <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[11px] font-semibold text-foreground">
                Esc
              </kbd>
              <span>{tr("gl_close_window_2", "ปิดหน้าต่าง")}</span>
            </span>
            <span className="hidden sm:inline-block">{tr("gl_double_click_to_select", "| ดับเบิ้ลคลิกเพื่อเลือก")}</span>
          </div>

          <div className="flex items-center gap-2">
            {!multiSelect && allowEmpty && selectedCode && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="rounded-xl"
                onClick={() => {
                  if (onSelect) onSelect("", null as never);
                  onClose();
                }}
              >
                {tr("gl_clear_selected_values", "ล้างค่าที่เลือก")}
              </Button>
            )}
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="rounded-xl"
              onClick={onClose}
            >
              {tr("gl_cancel_esc", "ยกเลิก (Esc)")}
            </Button>
            {multiSelect ? (
              <Button
                type="button"
                variant="default"
                size="sm"
                className="rounded-xl"
                onClick={confirmMultiSelect}
              >
                {tr("gl_ok_select_accounts", "ตกลงเลือก ({0} บัญชี)").replace("{0}", String(multiChecked.size))}
              </Button>
            ) : (
              <Button
                type="button"
                variant="default"
                size="sm"
                className="rounded-xl"
                disabled={!currentHighlighted}
                onClick={() => {
                  if (currentHighlighted) selectSingleAccount(currentHighlighted);
                }}
              >
                {tr("gl_select_this_account_enter", "เลือกบัญชีนี้ (Enter)")}
              </Button>
            )}
          </div>
        </footer>
      </div>
    </div>
  );
}
