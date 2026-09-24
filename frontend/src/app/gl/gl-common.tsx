"use client";

import { createContext, useCallback, useContext, useEffect, useId, useMemo, useRef, useState, type CSSProperties, type ReactNode } from "react";
import { Check as CheckIcon, Search, X } from "lucide-react";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { getAuthSession, restoreAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import type { GLTextFn } from "@/lib/general-ledger";
import { ResizableSplitter } from "@/components/ui/resizable-splitter";
import { Button } from "@/components/ui/button";
import { glAllRecords, glCommand, glRequest } from "@/lib/general-ledger-api";
import { accountName, formatAmount, sortAccountsHierarchically, type GLAccount, type GLCommand, type GLFiscalYear, type GLJournalBook, type GLPage, type GLRecord, type GLResource } from "@/lib/general-ledger";
import { cn } from "@/lib/utils";
import { thousandsCommasValid } from "@/lib/clipboard-journal-parser";
import { AccountSearchDialog } from "./account-search-dialog";
import { Combobox } from "@/components/ui/combobox";

export { AccountSearchDialog, Combobox };

// Screen text follows the selected language (AGENTS.md rule 2026-09-14): every
// user-visible string is an English key resolved from backend languages.tsv, with
// the Thai source text as fallback while the dictionary loads.
export type { GLTextFn };
const GLLanguageContext = createContext<{ language: LanguageCode; dictionary: BackendLanguageDictionary }>({ language: "th", dictionary: {} });
export function GLLanguageProvider({ language, children }: { language: LanguageCode; children: ReactNode }) {
  const [backendUrl, setBackendUrl] = useState(() => getAuthSession()?.backendUrl ?? "");
  useEffect(() => {
    if (backendUrl) return;
    let active = true;
    void restoreAuthSession().then((session) => { if (active && session) setBackendUrl(session.backendUrl); });
    return () => { active = false; };
  }, [backendUrl]);
  const dictionary = useBackendLanguage(language, backendUrl || undefined);
  const value = useMemo(() => ({ language, dictionary }), [language, dictionary]);
  return <GLLanguageContext.Provider value={value}>{children}</GLLanguageContext.Provider>;
}
export function useGLText(): GLTextFn {
  const { dictionary } = useContext(GLLanguageContext);
  return useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
}
/** ภาษาที่ผู้ใช้เลือก — ใช้เลือกชื่อข้อมูลหลักที่มีทั้งไทย/อังกฤษ (เช่น ชื่อสมุดรายวัน) */
export function useGLLanguage(): LanguageCode {
  return useContext(GLLanguageContext).language;
}

export const control = "min-h-[2.6em] w-full rounded-xl border border-input bg-background px-3 py-1.5 text-[0.95rem] leading-normal text-foreground shadow-[0_3px_10px_rgba(0,0,0,0.14),0_1px_3px_rgba(0,0,0,0.1)] dark:shadow-[0_3px_10px_rgba(0,0,0,0.6)] transition-[border-color,box-shadow] hover:border-primary/80 hover:shadow-[0_4px_16px_rgba(0,0,0,0.18),0_1px_4px_rgba(0,0,0,0.12)] focus-visible:outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring focus-visible:shadow-[0_4px_16px_rgba(0,0,0,0.2)] disabled:cursor-default disabled:bg-muted/20 disabled:border-border disabled:shadow-[0_2px_8px_rgba(0,0,0,0.1),0_1px_2px_rgba(0,0,0,0.07)] disabled:text-foreground";
export const panel = "min-w-0 rounded-2xl border border-border bg-card p-3 text-card-foreground shadow-sm";
export const actionClass = "h-10 !min-h-10 rounded-xl px-3.5 text-[0.95rem] font-medium leading-normal shrink-0 inline-flex items-center justify-center";
export function useRowDensity() {
  const [compact, setCompact] = useState(true);
  useEffect(() => { setCompact(localStorage.getItem("bc_gl_compact_rows") !== "false"); }, []);
  const toggle = () => setCompact((value) => { localStorage.setItem("bc_gl_compact_rows", String(!value)); return !value; });
  return { compact, toggle, tableClass: compact ? "[&_td]:py-1 [&_td_button]:min-h-8" : "[&_td]:py-2" };
}
export function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return <div className="grid min-w-0 gap-1 text-[0.95rem] leading-normal"><span className="font-medium text-foreground">{label}</span>{children}{hint && <span className="text-[0.9rem] text-muted-foreground">{hint}</span>}</div>;
}
export function Check({ label, checked, onChange, disabled, className }: { label: string; checked: boolean; onChange: (checked: boolean) => void; disabled?: boolean; className?: string }) {
  return (
    <label
      className={cn(
        "group inline-flex !w-auto !max-w-none shrink-0 items-center gap-2.5 rounded-xl border px-3.5 py-2 text-[0.95rem] font-medium leading-normal cursor-pointer select-none transition-all duration-150",
        "shadow-[0_2px_8px_rgba(0,0,0,0.06),0_1px_2px_rgba(0,0,0,0.04)] hover:shadow-[0_3px_12px_rgba(0,0,0,0.1)]",
        "focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-1 focus-within:outline-none",
        checked
          ? "border-primary/50 bg-primary/10 text-foreground hover:border-primary/70 hover:bg-primary/[0.14]"
          : "border-input/90 bg-background text-foreground/90 hover:border-border hover:bg-muted/40",
        disabled && "cursor-default opacity-60 pointer-events-none shadow-none bg-muted/20 border-border text-muted-foreground",
        className
      )}
    >
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        className="sr-only"
      />
      <span
        aria-hidden="true"
        className={cn(
          "flex size-5 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150",
          checked
            ? "border-primary bg-primary text-primary-foreground shadow-[0_1px_3px_rgba(0,0,0,0.15)]"
            : "border-muted-foreground/40 bg-background group-hover:border-primary/60 shadow-inner"
        )}
      >
        {checked && <CheckIcon className="size-3.5 stroke-[3.2]" />}
      </span>
      <span className="whitespace-nowrap">{label}</span>
    </label>
  );
}
export function Notice({ text, error = false }: { text: string; error?: boolean }) {
  return text ? <div role={error ? "alert" : "status"} className={`rounded-xl border p-3 text-[0.95rem] leading-relaxed shadow-[0_2px_8px_rgba(0,0,0,0.08)] ${error ? "border-destructive/40 bg-destructive/5 text-foreground" : "border-primary/25 bg-primary/5 text-foreground"}`}>{text}</div> : null;
}
export function UnsavedBadge({ dirty, className }: { dirty: boolean; className?: string }) {
  const tr = useGLText();
  if (!dirty) return null;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border border-primary/30 bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary",
        className,
      )}
      role="status"
      aria-live="polite"
    >
      <span className="size-1.5 rounded-full bg-primary animate-pulse" />
      {tr("gl_unsaved_changes", "ยังไม่บันทึก")}
    </span>
  );
}

export function useDebouncedSearch({
  onSearch,
  debounceMs = 2000,
  initialQuery = "",
}: {
  onSearch: (query: string) => void;
  debounceMs?: number;
  initialQuery?: string;
}) {
  const [query, setQueryState] = useState(initialQuery);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  const cancelTimer = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const searchNow = useCallback((overrideQuery?: string) => {
    cancelTimer();
    const target = typeof overrideQuery === "string" ? overrideQuery : query;
    onSearch(target);
  }, [cancelTimer, onSearch, query]);

  const setQuery = useCallback((val: string) => {
    setQueryState(val);
    cancelTimer();
    timerRef.current = setTimeout(() => {
      onSearch(val);
      timerRef.current = null;
    }, debounceMs);
  }, [cancelTimer, debounceMs, onSearch]);

  const clear = useCallback(() => {
    cancelTimer();
    setQueryState("");
    onSearch("");
  }, [cancelTimer, onSearch]);

  useEffect(() => {
    return () => cancelTimer();
  }, [cancelTimer]);

  return {
    query,
    setQuery,
    searchNow,
    clear,
    isPending: !!timerRef.current,
  };
}

export function SearchInput({
  value,
  onChange,
  onClear,
  onSearch,
  placeholder: placeholderProp,
  ariaLabel: ariaLabelProp,
  className = "min-w-28 flex-1",
  disabled = false,
  autoFocus = false,
  inputRef: externalRef,
}: {
  value: string;
  onChange: (value: string) => void;
  onClear?: () => void;
  onSearch?: (value: string) => void;
  placeholder?: string;
  ariaLabel?: string;
  className?: string;
  disabled?: boolean;
  autoFocus?: boolean;
  inputRef?: React.RefObject<HTMLInputElement | null>;
}) {
  const tr = useGLText();
  const placeholder = placeholderProp ?? tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ");
  const ariaLabel = ariaLabelProp ?? tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ");
  const localRef = useRef<HTMLInputElement>(null);
  const inputRef = externalRef ?? localRef;

  const handleClear = () => {
    onChange("");
    onClear?.();
    inputRef.current?.focus();
  };

  return (
    <div className={`relative flex items-center h-10 ${className}`}>
      <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground z-10" aria-hidden="true" />
      <input
        ref={inputRef}
        type="text"
        className={`${control} !h-10 !min-h-10 !pl-9.5 ${value ? "!pr-9" : ""} w-full`}
        aria-label={ariaLabel}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && onSearch) {
            e.preventDefault();
            onSearch(value);
          } else if (e.key === "Escape" && value) {
            e.preventDefault();
            handleClear();
          }
        }}
        disabled={disabled}
        autoFocus={autoFocus}
      />
      {value ? (
        <button
          type="button"
          tabIndex={-1}
          className="absolute right-2 top-1/2 -translate-y-1/2 z-10 inline-flex !size-7 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-full hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          onClick={handleClear}
          title={tr("gl_clear_search_text_clean", "ล้างข้อความค้นหา (Clean)")}
          aria-label={tr("gl_clear_search_text", "ล้างข้อความค้นหา")}
        >
          <X className="size-4" />
        </button>
      ) : null}
    </div>
  );
}

/** Formats a numeric string with thousands commas and fixed scale decimals (unified with formatAmount) */
export const formatAmountValue = formatAmount;
export { formatAmount };



/**
 * Normalises typed/pasted amount text without changing its value: Thai digits → 0-9, accounting
 * negatives "(1,500.00)" → "-1500.00", thousands commas/spaces/currency signs removed.
 * Anything else that has to be removed (letters, a second decimal point, a stray minus) sets
 * `dropped` so the screen tells the user instead of changing the number silently — so do commas
 * that are not thousands separators ("1.500,00" European format would otherwise become 1.50).
 * The minus sign is never dropped here — a field that does not accept negatives shows a message.
 */
export function normalizeAmountText(raw: string): { value: string; dropped: boolean } {
  // Thai digits U+0E50–U+0E59 and the baht sign U+0E3F are written as escapes (GL files keep Thai text in tr() only)
  let text = raw.replace(/[\u0E50-\u0E59]/g, (digit) => String(digit.charCodeAt(0) - 0x0e50)).trim();
  const accountingNegative = /^\((.*)\)$/.exec(text);
  if (accountingNegative) text = `-${accountingNegative[1].trim()}`;
  text = text.replace(/[\s\u0E3F$]/g, "");
  const misplacedComma = /^-?[0-9.,]+$/.test(text) && !thousandsCommasValid(text.replace(/^-/, ""));
  text = text.replace(/,/g, "");
  let dropped = misplacedComma || /[^0-9.-]/.test(text);
  text = text.replace(/[^0-9.-]/g, "");
  const negative = text.startsWith("-");
  if ((text.match(/-/g)?.length ?? 0) > (negative ? 1 : 0)) dropped = true;
  let body = text.replace(/-/g, "");
  const dot = body.indexOf(".");
  if (dot !== -1 && body.indexOf(".", dot + 1) !== -1) {
    dropped = true;
    body = body.slice(0, dot + 1) + body.slice(dot + 1).replace(/\./g, "");
  }
  return { value: (negative ? "-" : "") + body, dropped };
}

/** Amount text that the field cannot accept as typed: a negative in a positive-only field, or more decimals than the scale (never rounded). */
export function amountInputIssue(value: string, scale: number, allowNegative: boolean): "negative" | "scale" | "" {
  const text = (value ?? "").replace(/,/g, "").trim();
  if (!allowNegative && text.startsWith("-") && /[1-9]/.test(text)) return "negative";
  const fraction = text.split(".")[1] ?? "";
  if (fraction.replace(/0+$/, "").length > Math.max(0, scale)) return "scale";
  return "";
}

/** "Characters removed" is shown while editing, and afterwards only while the field still holds the value it emitted then. */
export function droppedWarningShown(droppedFor: string | null, focused: boolean, value: string): boolean {
  return droppedFor !== null && (focused || droppedFor === value);
}

/** The message a field shows: a real problem with the number (negative, too many decimals) before "characters removed". */
export function amountFieldIssue(text: string, scale: number, allowNegative: boolean, droppedShown: boolean): "negative" | "scale" | "characters" | "" {
  return amountInputIssue(text, scale, allowNegative) || (droppedShown ? "characters" : "");
}

/** Text shown when the field gains focus: plain digits without commas; a zero focuses as "" so typing "1" gives "1" (not "01"). */
export function amountFocusText(value: string): string {
  const raw = (value || "").replace(/,/g, "").trim();
  return /^-?0*(\.0*)?$/.test(raw) ? "" : raw;
}

/** Value to emit on blur — null when the user did not type (Tab/click through keeps the stored value, e.g. "0", exactly). */
export function amountBlurValue(touched: boolean, text: string, scale: number, emptyValue: string): string | null {
  return touched ? commitAmountText(text, scale, emptyValue) : null;
}

/** Blur/Enter commit: pads decimals to the scale, strips leading zeros and a negative zero — never rounds or drops digits. */
export function commitAmountText(text: string, scale: number, emptyValue: string): string {
  const clean = text.replace(/,/g, "").trim();
  if (!clean || clean === "-" || clean === "." || clean === "-.") return emptyValue;
  const negative = clean.startsWith("-");
  const [wholeRaw = "", fractionRaw = ""] = clean.replace(/^-/, "").split(".");
  const whole = wholeRaw.replace(/^0+(?=\d)/, "") || "0";
  const trimmed = fractionRaw.length > scale ? fractionRaw.replace(/0+$/, "") : fractionRaw;
  const fraction = scale > 0 ? trimmed.padEnd(scale, "0") : trimmed;
  const isZero = /^0+$/.test(whole) && /^0*$/.test(fraction);
  return `${negative && !isZero ? "-" : ""}${whole}${fraction ? `.${fraction}` : ""}`;
}

/**
 * AmountInput — Formatted numeric input for financial amounts and GL screens.
 * - Idle / Blur: Formatted with thousands commas and fixed scale decimals, right-aligned (e.g. 70,000.00).
 * - Focus / Edit: Switches to plain text without commas (e.g. 70000.00), right-aligned, select-all for rapid editing.
 * - Blur / Enter: Normalizes decimals, applies commas, emits clean string to parent.
 */
export function AmountInput({
  value,
  onChange,
  scale = 2,
  allowNegative = false,
  allowEmpty = true,
  placeholder,
  className = "",
  disabled = false,
  required = false,
  ariaLabel,
  autoFocus = false,
  id,
  onBlur: externalBlur,
}: {
  value: string;
  onChange: (value: string) => void;
  scale?: number;
  allowNegative?: boolean;
  allowEmpty?: boolean;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
  required?: boolean;
  ariaLabel?: string;
  autoFocus?: boolean;
  id?: string;
  onBlur?: () => void;
}) {
  const tr = useGLText();
  const messageId = useId();
  const [focused, setFocused] = useState(false);
  const [editText, setEditText] = useState<string | null>(null);
  // true only after the user typed/pasted: tabbing through a field must never rewrite its value
  const touchedRef = useRef(false);
  // The value this field emitted when it removed characters. The warning belongs to that value only:
  // rows are keyed by index, so after a row is deleted or a new voucher opens, this input shows another
  // value and must not carry the old warning (adversarial review 2026-09-24).
  const [droppedFor, setDroppedFor] = useState<string | null>(null);
  const droppedCharacters = droppedWarningShown(droppedFor, focused, value);
  const suppressMouseUpRef = useRef(false);
  const resolvedPlaceholder = placeholder ?? (scale > 0 ? `0.${"0".repeat(scale)}` : "0");
  const zeroText = scale > 0 ? `0.${"0".repeat(scale)}` : "0";

  // Formatted string when not focused (commas + decimals)
  const displayFormatted = useMemo(() => {
    if (!value || !value.trim()) {
      return !allowEmpty || required ? `0.${"0".repeat(scale)}` : "";
    }
    return formatAmountValue(value, scale);
  }, [value, scale, allowEmpty, required]);

  // When focused show raw plain text without commas; otherwise show formatted display
  const shownText = focused && editText !== null ? editText : displayFormatted;

  const handleFocus = (event: React.FocusEvent<HTMLInputElement>) => {
    setFocused(true);
    touchedRef.current = false;
    setDroppedFor(null);
    const focusText = amountFocusText(value);
    const isZero = focusText === "";
    setEditText(focusText);
    suppressMouseUpRef.current = true;
    const target = event.currentTarget;
    requestAnimationFrame(() => {
      if (!isZero) {
        target.select();
      }
    });
  };

  const handleMouseUp = (event: React.MouseEvent<HTMLInputElement>) => {
    if (suppressMouseUpRef.current) {
      event.preventDefault();
      suppressMouseUpRef.current = false;
    }
  };

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const normalized = normalizeAmountText(event.target.value);
    // Normalize leading zeros: e.g. typing "1" when "0" is present: "01" -> "1", "-01" -> "-1"
    const cleaned = normalized.value.replace(/^(-?)0+([1-9])/, "$1$2");
    touchedRef.current = true;
    setDroppedFor(normalized.dropped ? cleaned : null);
    setEditText(cleaned);
    onChange(cleaned);
  };

  const handleBlur = () => {
    setFocused(false);
    setEditText(null);
    // Untouched (e.g. Tab through a 0.00 field): keep the stored value exactly — never turn "0" into "".
    const next = amountBlurValue(touchedRef.current, editText ?? value ?? "", scale, !allowEmpty || required ? zeroText : "");
    touchedRef.current = false;
    if (next !== null) {
      if (droppedFor !== null) setDroppedFor(next); // keep the warning on the committed value
      onChange(next);
    }
    externalBlur?.();
  };
  const issue = amountFieldIssue(focused && editText !== null ? editText : value, scale, allowNegative, droppedCharacters);
  const issueText = issue === "negative"
    ? tr("gl_amount_negative_not_allowed", "ช่องนี้ไม่รับยอดติดลบ — กรุณาใส่ตัวเลขบวก (ถ้าเป็นรายการกลับด้าน ให้ใส่อีกฝั่ง)")
    : issue === "scale"
      ? tr("gl_amount_too_many_decimals", "ใส่ทศนิยมได้ไม่เกิน {0} ตำแหน่ง — ระบบไม่ปัดเศษให้ กรุณาแก้ตัวเลข").replace("{0}", String(scale))
      : issue === "characters"
        ? tr("gl_amount_characters_removed", "ตัดตัวอักษรที่ไม่ใช่ตัวเลขออกแล้ว — กรุณาตรวจตัวเลขในช่องนี้อีกครั้ง")
        : "";

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.currentTarget.blur();
    }
  };

  // Same wrapper in every state so the input is never remounted (focus stays) when a message appears.
  return (
    <span className="grid w-full min-w-0 gap-1">
      <input
        id={id}
        type="text"
        inputMode="decimal"
        className={`${control} text-right tabular-nums ${issueText ? "border-destructive focus-visible:border-destructive" : ""} ${className}`}
        aria-label={ariaLabel}
        aria-invalid={issueText ? true : undefined}
        aria-describedby={issueText ? messageId : undefined}
        placeholder={resolvedPlaceholder}
        value={shownText}
        disabled={disabled}
        required={required}
        autoFocus={autoFocus}
        onFocus={handleFocus}
        onMouseUp={handleMouseUp}
        onChange={handleChange}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
      />
      {issueText && <span id={messageId} role="alert" className="text-[0.9rem] leading-snug text-destructive [overflow-wrap:anywhere]">{issueText}</span>}
    </span>
  );
}

export function AccountSelect({ value, onChange, accounts, label: labelProp, all = false, disabled = false, allowEmpty = true, field }: { value: string; onChange: (value: string) => void; accounts: GLAccount[]; label?: string; all?: boolean; disabled?: boolean; allowEmpty?: boolean; field?: string }) {
  const tr = useGLText();
  const label = labelProp ?? tr("gl_account", "บัญชี");
  const [dialogOpen, setDialogOpen] = useState(false);
  const sortedAccounts = useMemo(() => sortAccountsHierarchically(accounts), [accounts]);
  return (
    <>
      <div className="relative flex items-center w-full min-w-0">
        <select
          data-field={field}
          aria-label={label}
          className={`${control} appearance-none [&::-ms-expand]:hidden ${allowEmpty && value ? "!pr-20" : "!pr-12"} cursor-pointer text-ellipsis truncate`}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          onMouseDown={(event) => {
            event.preventDefault();
            setDialogOpen(true);
          }}
          onKeyDown={(event) => {
            if (event.key === "F2" || event.key === " " || event.key === "Enter") {
              event.preventDefault();
              setDialogOpen(true);
            }
          }}
          disabled={disabled}
          title={tr("gl_f2_full_coa_search", "{0} (คลิกหรือกด F2 เพื่อค้นหาผังบัญชีแบบเต็มจอ)").replace("{0}", String(label))}
        >
          {allowEmpty && <option value="">{tr("gl_select_acct_search_full", "เลือกบัญชี (กดค้นหาเพื่อเปิดจอใหญ่)")}</option>}
          {sortedAccounts.filter((account) => all || account.allowposting && account.isactive).map((account) => {
            const level = account.effectiveLevel ?? account.level ?? 1;
            const indent = level > 1 ? `${"\u00A0\u00A0".repeat(level - 1)}└─ ` : "";
            return <option key={account.id ?? account.accountcode} value={account.accountcode}>{indent}{account.accountcode} · {accountName(account)}</option>;
          })}
        </select>
        <div className="absolute right-2 top-1/2 -translate-y-1/2 z-10 flex items-center gap-1.5 pointer-events-auto">
          {allowEmpty && value && !disabled && (
            <button
              type="button"
              tabIndex={-1}
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onChange("");
              }}
              className="inline-flex !size-6 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-full text-muted-foreground/70 hover:bg-muted hover:text-foreground transition-colors cursor-pointer shrink-0"
              title={tr("gl_clear_selected_values", "ล้างค่าที่เลือก")}
              aria-label={tr("gl_clear_selected_values", "ล้างค่าที่เลือก")}
            >
              <X className="size-3.5" />
            </button>
          )}
          <button
            type="button"
            tabIndex={-1}
            disabled={disabled}
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              setDialogOpen(true);
            }}
            className="inline-flex !size-7 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-[8px] border border-primary/20 bg-primary/10 hover:bg-primary/20 text-primary transition-all active:scale-95 disabled:pointer-events-none disabled:opacity-50 cursor-pointer shadow-[0_2px_6px_rgba(0,0,0,0.14)] hover:shadow-[0_3px_10px_rgba(0,0,0,0.2)] shrink-0"
            title={tr("gl_open_fullscreen_coa_search_f2", "เปิดระบบค้นหาผังบัญชีแบบเต็มจอ (F2)")}
            aria-label={tr("gl_search_coa", "ค้นหาผังบัญชี")}
          >
            <Search className="size-3.5" />
          </button>
        </div>
      </div>

      <AccountSearchDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        onSelect={(code) => onChange(code)}
        accounts={accounts}
        selectedCode={value}
        all={all}
        title={tr("gl_search_and_select", "ค้นหาและเลือก{0}").replace("{0}", String(label))}
        allowEmpty={allowEmpty}
      />
    </>
  );
}
export function YearSelect({ value, onChange, years, label: labelProp, disabled = false }: { value: string; onChange: (value: string) => void; years: GLFiscalYear[]; label?: string; disabled?: boolean }) {
  const tr = useGLText();
  const label = labelProp ?? tr("gl_fiscal_year", "ปีบัญชี");
  return (
    <Combobox
      aria-label={label}
      value={value}
      onChange={(val: string | number) => onChange(String(val))}
      disabled={disabled}
      placeholder={tr("gl_select_fiscal_year", "เลือกปีบัญชี")}
    >
      <option value="">{tr("gl_select_fiscal_year", "เลือกปีบัญชี")}</option>
      {years.map((year) => (
        <option key={year.id ?? year.code} value={year.code}>
          {year.code}{year.closed ? ` · ${tr("gl_closed", "ปิดแล้ว")}` : ""}
        </option>
      ))}
    </Combobox>
  );
}
interface ReferencesCache {
  accounts: GLAccount[];
  years: GLFiscalYear[];
  books: GLJournalBook[];
  loaded: boolean;
}

let referencesCache: ReferencesCache = {
  accounts: [],
  years: [],
  books: [],
  loaded: false,
};

let inFlightReferences: Promise<ReferencesCache> | null = null;

export function invalidateReferencesCache() {
  referencesCache = { accounts: [], years: [], books: [], loaded: false };
  inFlightReferences = null;
}

export function useReferences(refresh = 0) {
  const [revision, setRevision] = useState(0);
  const reload = useCallback(() => {
    invalidateReferencesCache();
    setRevision((value) => value + 1);
  }, []);
  const [accounts, setAccounts] = useState<GLAccount[]>(() => referencesCache.accounts);
  const [years, setYears] = useState<GLFiscalYear[]>(() => referencesCache.years);
  const [books, setBooks] = useState<GLJournalBook[]>(() => referencesCache.books);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    if (referencesCache.loaded && revision === 0 && refresh === 0) {
      setAccounts(referencesCache.accounts);
      setYears(referencesCache.years);
      setBooks(referencesCache.books);
      return;
    }
    if (!inFlightReferences) {
      inFlightReferences = Promise.all([
        glAllRecords<GLAccount>("accounts", "", 10000),
        glAllRecords<GLFiscalYear>("fiscal-years", "", 1000),
        glAllRecords<GLJournalBook>("journal-books", "", 1000),
      ])
        .then(([a, y, b]) => {
          referencesCache = { accounts: a, years: y, books: b, loaded: true };
          inFlightReferences = null;
          return referencesCache;
        })
        .catch((err) => {
          inFlightReferences = null;
          throw err;
        });
    }
    inFlightReferences
      .then((cache) => {
        if (active) {
          setAccounts(cache.accounts);
          setYears(cache.years);
          setBooks(cache.books);
          setError("");
        }
      })
      .catch((e: Error) => {
        if (active) setError(e.message);
      });
    return () => {
      active = false;
    };
  }, [refresh, revision]);

  return { accounts, years, books, error, reload };
}
export function useGLList<T extends GLRecord>(resource: GLResource, query = "", extra = "", defaultLimit?: number) {
  const limit = defaultLimit ?? (resource === "accounts" ? 1000 : 30);
  const [page, setPage] = useState(1), [revision, setRevision] = useState(0);
  const [data, setData] = useState<GLPage<T>>({ items: [], total: 0, page: 1, limit, sequence: 0 });
  const [loading, setLoading] = useState(false), [error, setError] = useState("");
  useEffect(() => { let active = true; setLoading(true);
    glRequest<GLPage<T>>(`${resource}?${new URLSearchParams({ q: query, page: String(page), limit: String(limit) })}${extra ? `&${extra}` : ""}`)
      .then((result) => { if (active) { setData({ ...result, items: result.items ?? [] }); setError(""); } })
      .catch((e: Error) => { if (active) setError(e.message); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [resource, query, page, revision, extra, limit]);
  const reload = useCallback(() => setRevision((value) => value + 1), []);
  return { data, loading, error, page, setPage, reload };
}
export function Pager({ page, total, onPage, loading, limit = 30 }: { page: number; total: number; onPage: (page: number) => void; loading: boolean; limit?: number }) {
  const tr = useGLText();
  return <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-2 text-[0.95rem]"><span>{tr("gl_items_page", "{0} รายการ · หน้า {1}").replace("{0}", String(total.toLocaleString("th-TH"))).replace("{1}", String(page))}</span><div className="flex gap-2"><Button className={actionClass} variant="outline" disabled={loading || page <= 1} onClick={() => onPage(page - 1)}>{tr("gl_prev", "ก่อนหน้า")}</Button><Button className={actionClass} variant="outline" disabled={loading || page * limit >= total} onClick={() => onPage(page + 1)}>{tr("gl_next", "ถัดไป")}</Button></div></div>;
}
export function useGLCommand() {
  const [busy, setBusy] = useState(false);
  const inFlight = useRef(false), retry = useRef({ body: "", id: "" });
  const tr = useGLText();
  const execute = useCallback(async (command: Omit<GLCommand, "requestid">) => {
    if (inFlight.current) throw new Error(tr("gl_saving_please_wait", "กำลังบันทึก กรุณารอสักครู่"));
    inFlight.current = true; setBusy(true);
    const body = JSON.stringify(command);
    if (retry.current.body !== body) retry.current = { body, id: crypto.randomUUID() };
    try { const result = await glCommand(command, retry.current.id); retry.current = { body: "", id: "" }; return result; }
    finally { inFlight.current = false; setBusy(false); }
  }, [tr]);
  return { busy, execute };
}
export function useDirtyGuard(route: string, dirty: boolean) {
  useEffect(() => {
    window.dispatchEvent(new CustomEvent("bc-gl-dirty", { detail: { route, dirty } }));
    const handler = (event: BeforeUnloadEvent) => { if (dirty) { event.preventDefault(); event.returnValue = ""; } };
    window.addEventListener("beforeunload", handler);
    return () => { window.removeEventListener("beforeunload", handler); window.dispatchEvent(new CustomEvent("bc-gl-dirty", { detail: { route, dirty: false } })); };
  }, [route, dirty]);
}
export function SplitWorkbench({ list, editor, className }: { list: ReactNode; editor: ReactNode; className?: string }) {
  const tr = useGLText();
  const [width, setWidth] = useState(38), container = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);
  useEffect(() => { const saved = localStorage.getItem("bc_gl_split_percent"); if (saved && /^\d+$/.test(saved)) setWidth(Math.min(65, Math.max(25, Number(saved)))); }, []);
  const update = (value: number) => { const next = Math.round(Math.min(65, Math.max(25, value))); setWidth(next); localStorage.setItem("bc_gl_split_percent", String(next)); };
  return <div ref={container} className={`flex min-w-0 flex-1 flex-col gap-2 xl:flex-row xl:h-full xl:min-h-0 ${className ?? ""}`} style={{ "--gl-list-width": `${width}%` } as CSSProperties}>
    <div data-gl-pane="list" className={`${panel} flex flex-col min-h-0 min-w-0 w-full xl:w-[var(--gl-list-width)] xl:shrink-0`}>{list}</div>
    <ResizableSplitter breakpoint="xl" value={width} min={25} max={65} label={tr("gl_adjust_acct_list_width", "ปรับความกว้างรายการบัญชี")} onDoubleClick={() => update(38)} onKeyDown={(event) => { if (["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) { event.preventDefault(); update(event.key === "Home" ? 25 : event.key === "End" ? 65 : width + (event.key === "ArrowLeft" ? -2 : 2)); } }}
      onPointerDown={(event) => { dragging.current = true; event.currentTarget.setPointerCapture(event.pointerId); }} onPointerUp={() => { dragging.current = false; }} onPointerCancel={() => { dragging.current = false; }} onPointerMove={(event) => { if (dragging.current && container.current) { const rect = container.current.getBoundingClientRect(); update((event.clientX - rect.left) / rect.width * 100); } }} />
    <div data-gl-pane="editor" className={`${panel} flex flex-col min-h-0 min-w-0 flex-1`}>{editor}</div>
  </div>;
}
export function downloadText(filename: string, text: string, type: string) { const url = URL.createObjectURL(new Blob([text], { type })); const link = document.createElement("a"); link.href = url; link.download = filename; link.click(); setTimeout(() => URL.revokeObjectURL(url), 1000); }
