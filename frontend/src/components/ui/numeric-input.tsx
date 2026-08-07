"use client";

/**
 * NumericInput — the global numeric field for BC Ai Account.
 * Right-aligned input + calculator button that opens a popup calculator
 * (React Portal to body, flips above the trigger when space below is tight).
 * Used by every NumberField variant across product/barcode screens.
 */
import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Calculator as CalculatorIcon, Delete, Redo2, Undo2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input } from "./input";

export type NumericInputProps = {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  max?: number;
  /** Reserved for stepper UIs — the text-editing model does not use native step. */
  step?: string | number;
  disabled?: boolean;
  className?: string;
  placeholder?: string;
  ariaLabel?: string;
};

const POPUP_WIDTH = 248;
const POPUP_HEIGHT = 330;

/** Evaluate a simple + − × ÷ expression with operator precedence (no eval). */
function evaluateExpression(tokens: string[]): number | null {
  if (tokens.length === 0) return null;
  const nums: number[] = [];
  const ops: string[] = [];
  let expectNumber = true;
  for (const token of tokens) {
    if (expectNumber) {
      const n = Number(token);
      if (!Number.isFinite(n)) return null;
      nums.push(n);
      expectNumber = false;
    } else {
      if (!["+", "-", "*", "/"].includes(token)) return null;
      ops.push(token);
      expectNumber = true;
    }
  }
  if (expectNumber) return null; // trailing operator
  // first pass: × ÷
  const n2: number[] = [nums[0]];
  const o2: string[] = [];
  for (let i = 0; i < ops.length; i += 1) {
    if (ops[i] === "*" || ops[i] === "/") {
      const left = n2.pop() ?? 0;
      const right = nums[i + 1];
      n2.push(ops[i] === "*" ? left * right : right === 0 ? NaN : left / right);
    } else {
      o2.push(ops[i]);
      n2.push(nums[i + 1]);
    }
  }
  // second pass: + −
  let result = n2[0];
  for (let i = 0; i < o2.length; i += 1) {
    result = o2[i] === "+" ? result + n2[i + 1] : result - n2[i + 1];
  }
  return Number.isFinite(result) ? result : null;
}

export function NumericInput({
  value,
  onChange,
  min,
  max,
  disabled,
  className,
  placeholder,
  ariaLabel,
}: NumericInputProps) {
  const wrapRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [popupStyle, setPopupStyle] = useState<React.CSSProperties>({});
  const [tokens, setTokens] = useState<string[]>([]);
  const [entry, setEntry] = useState("0");
  const [evaluated, setEvaluated] = useState(false);
  // Local text state while editing: a controlled numeric input that coerces on
  // every keystroke makes the leading "0" immortal (click-end + type "1" gives
  // "10", clear + type "7" gives "07"). Keep raw text while focused, select-all
  // on focus so typing replaces, commit/clamp on blur.
  const [focused, setFocused] = useState(false);
  const [text, setText] = useState<string | null>(null);
  const suppressMouseUpRef = useRef(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const safeValue = Number.isFinite(value) ? value : 0;
  const shownText = focused && text !== null ? text : String(safeValue);

  // Per-field undo/redo history: every distinct value the field emits (typed,
  // blur-committed, or calculator-applied) becomes an undo step; focus always
  // follows the restored value so the user sees exactly where the undo landed.
  const [undoStack, setUndoStack] = useState<number[]>([]);
  const [redoStack, setRedoStack] = useState<number[]>([]);
  const lastEmittedRef = useRef(safeValue);

  // External value changes (parent reloads etc.) rebase the history anchor
  // without creating an undo step.
  useEffect(() => {
    lastEmittedRef.current = safeValue;
  }, [safeValue]);

  const emit = (next: number) => {
    // Capture the anchor BEFORE queueing the updater: setState updaters run at
    // render time (after lastEmittedRef is reassigned) and StrictMode may
    // double-invoke them — reading the ref inside the updater pushed the NEW
    // value onto the undo stack instead of the previous one.
    const anchor = lastEmittedRef.current;
    if (next === anchor) return;
    setUndoStack((stack) => [...stack.slice(-49), anchor]);
    setRedoStack([]);
    lastEmittedRef.current = next;
    onChange(next);
  };

  const focusAndSelect = () => {
    // onFocus handler performs select-all; just move focus to the field.
    inputRef.current?.focus();
  };

  const undo = () => {
    if (undoStack.length === 0) return;
    const previous = undoStack[undoStack.length - 1];
    setUndoStack(undoStack.slice(0, -1));
    setRedoStack([...redoStack, lastEmittedRef.current]);
    lastEmittedRef.current = previous;
    onChange(previous);
    focusAndSelect();
    // setText AFTER focus so the onFocus init can't overwrite the restored value
    setText(String(previous));
  };

  const redo = () => {
    if (redoStack.length === 0) return;
    const next = redoStack[redoStack.length - 1];
    setRedoStack(redoStack.slice(0, -1));
    setUndoStack([...undoStack, lastEmittedRef.current]);
    lastEmittedRef.current = next;
    onChange(next);
    focusAndSelect();
    setText(String(next));
  };

  useEffect(() => {
    if (!open || !wrapRef.current) return;
    const rect = wrapRef.current.getBoundingClientRect();
    const spaceBelow = window.innerHeight - rect.bottom;
    const style: React.CSSProperties = {
      position: "fixed",
      width: POPUP_WIDTH,
      zIndex: 100,
      left: Math.min(Math.max(8, rect.right - POPUP_WIDTH), window.innerWidth - POPUP_WIDTH - 8),
    };
    if (spaceBelow >= POPUP_HEIGHT + 8) {
      style.top = rect.bottom + 6;
      style.maxHeight = POPUP_HEIGHT;
    } else {
      style.bottom = window.innerHeight - rect.top + 6;
      style.maxHeight = Math.max(220, spaceBelow + POPUP_HEIGHT - 40);
    }
    setPopupStyle(style);
    setTokens([]);
    setEntry(String(safeValue));
    setEvaluated(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open]);

  const currentResult = () => {
    const chain = [...tokens, entry];
    if (chain.length === 1) {
      const n = Number(chain[0]);
      return Number.isFinite(n) ? n : null;
    }
    return evaluateExpression(chain);
  };

  const applyResult = () => {
    const result = currentResult();
    if (result === null) return;
    let next = result;
    if (min !== undefined) next = Math.max(min, next);
    if (max !== undefined) next = Math.min(max, next);
    emit(next);
    // keep the focused field's local text in sync with the applied value,
    // otherwise the blur-commit would clobber the calculator result
    setText(String(next));
    setOpen(false);
  };

  const pushDigit = (digit: string) => {
    setEvaluated(false);
    setEntry((current) => {
      if (evaluated || current === "0") return digit === "." ? "0." : digit;
      if (digit === "." && current.includes(".")) return current;
      return current + digit;
    });
  };

  const pushOperator = (op: string) => {
    const base = currentResult();
    if (base === null) return;
    setTokens([String(base), op]);
    setEntry("0");
    setEvaluated(false);
  };

  const backspace = () => {
    setEntry((current) => (current.length <= 1 ? "0" : current.slice(0, -1)));
  };

  const clearAll = () => {
    setTokens([]);
    setEntry("0");
    setEvaluated(false);
  };

  const equals = () => {
    const result = currentResult();
    if (result === null) return;
    setTokens([]);
    setEntry(String(Math.round(result * 1e10) / 1e10));
    setEvaluated(true);
  };

  const keyButton = (
    label: React.ReactNode,
    onPress: () => void,
    opts: { span?: boolean; tone?: "op" | "apply" | "muted"; aria?: string } = {},
  ) => (
    <button
      key={opts.aria ?? "key"}
      type="button"
      aria-label={opts.aria}
      onClick={onPress}
      className={cn(
        "grid h-10 place-items-center rounded-lg border text-sm font-semibold transition",
        opts.tone === "op" &&
          "border-primary/30 bg-primary/10 text-primary hover:bg-primary/20",
        opts.tone === "apply" &&
          "border-primary bg-primary text-primary-foreground hover:bg-primary/90",
        (!opts.tone || opts.tone === "muted") &&
          "border-border bg-background text-foreground hover:bg-muted",
        opts.span && "col-span-2",
      )}
    >
      {label}
    </button>
  );

  const keypad: { label: React.ReactNode; onPress: () => void; tone?: "op" | "muted"; aria?: string }[] = [
    { label: "7", onPress: () => pushDigit("7") },
    { label: "8", onPress: () => pushDigit("8") },
    { label: "9", onPress: () => pushDigit("9") },
    { label: "÷", onPress: () => pushOperator("/"), tone: "op", aria: "หาร" },
    { label: "4", onPress: () => pushDigit("4") },
    { label: "5", onPress: () => pushDigit("5") },
    { label: "6", onPress: () => pushDigit("6") },
    { label: "×", onPress: () => pushOperator("*"), tone: "op", aria: "คูณ" },
    { label: "1", onPress: () => pushDigit("1") },
    { label: "2", onPress: () => pushDigit("2") },
    { label: "3", onPress: () => pushDigit("3") },
    { label: "−", onPress: () => pushOperator("-"), tone: "op", aria: "ลบ" },
    { label: "0", onPress: () => pushDigit("0") },
    { label: ".", onPress: () => pushDigit("."), aria: "จุดทศนิยม" },
    { label: <Delete className="size-4" />, onPress: backspace, tone: "muted", aria: "ลบตัวเลขทีละตัว" },
    { label: "+", onPress: () => pushOperator("+"), tone: "op", aria: "บวก" },
  ];

  return (
    <div ref={wrapRef} className={cn("relative", className)}>
      <Input
        ref={inputRef}
        type="text"
        inputMode="decimal"
        value={shownText}
        disabled={disabled}
        placeholder={placeholder}
        aria-label={ariaLabel}
        onFocus={(event) => {
          setFocused(true);
          // A zero value focuses as EMPTY (per Jead: typing into a 0 field must
          // start fresh, no pre-filled "0"); non-zero selects all for replace.
          setText(safeValue === 0 ? "" : String(safeValue));
          // select-all synchronously so the first typed digit replaces the whole
          // value. The matching onMouseUp preventDefault keeps the browser's
          // caret placement from clearing it.
          event.target.select();
          suppressMouseUpRef.current = true;
        }}
        onMouseUp={(event) => {
          if (suppressMouseUpRef.current) {
            event.preventDefault();
            suppressMouseUpRef.current = false;
          }
        }}
        onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
          const raw = event.target.value;
          // digits, one dot, leading minus — drop everything else as typed
          const cleaned = raw.replace(/[^0-9.\-]/g, "").replace(/(\..*)\./g, "$1");
          setText(cleaned);
          if (cleaned === "" || cleaned === "-" || cleaned === ".") return; // still typing
          const n = Number(cleaned);
          if (Number.isFinite(n)) emit(n);
        }}
        onBlur={() => {
          setFocused(false);
          const n = text === null || text === "" || text === "-" || text === "." ? NaN : Number(text);
          let committed = Number.isFinite(n) ? (n as number) : safeValue;
          if (min !== undefined) committed = Math.max(min, committed);
          if (max !== undefined) committed = Math.min(max, committed);
          setText(null);
          if (committed !== safeValue) emit(committed);
        }}
        onKeyDown={(event) => {
          if (!(event.ctrlKey || event.metaKey)) return;
          const key = event.key.toLowerCase();
          if (key === "z" && !event.shiftKey) {
            event.preventDefault();
            undo();
          } else if (key === "y" || (key === "z" && event.shiftKey)) {
            event.preventDefault();
            redo();
          }
        }}
        className="text-right"
        style={{ paddingRight: `${10 + 26 * (1 + (undoStack.length > 0 ? 1 : 0) + (redoStack.length > 0 ? 1 : 0))}px` }}
      />
      <div className="absolute inset-y-0 right-0 flex items-center">
        {redoStack.length > 0 ? (
          <button
            type="button"
            aria-label="ทำซ้ำ"
            title="ทำซ้ำ (Ctrl+Y)"
            tabIndex={-1}
            disabled={disabled}
            onClick={redo}
            className="grid h-full w-7 place-items-center text-muted-foreground transition hover:text-primary disabled:opacity-50"
          >
            <Redo2 className="size-3.5" />
          </button>
        ) : null}
        {undoStack.length > 0 ? (
          <button
            type="button"
            aria-label="ย้อนกลับ"
            title="ย้อนกลับ (Ctrl+Z)"
            tabIndex={-1}
            disabled={disabled}
            onClick={undo}
            className="grid h-full w-7 place-items-center text-muted-foreground transition hover:text-primary disabled:opacity-50"
          >
            <Undo2 className="size-3.5" />
          </button>
        ) : null}
        <button
          type="button"
          aria-label="เปิดเครื่องคิดเลย"
          title="เครื่องคิดเลย"
          disabled={disabled}
          onClick={() => setOpen((current) => !current)}
          className="grid h-full w-8 place-items-center rounded-r-lg text-muted-foreground transition hover:text-primary disabled:opacity-50"
        >
          <CalculatorIcon className="size-4" />
        </button>
      </div>

      {open
        ? createPortal(
            <div
              className="fixed inset-0 z-[99] bg-transparent"
              onPointerDown={() => setOpen(false)}
              aria-hidden="true"
            >
              <div
                role="dialog"
                aria-modal="true"
                aria-label="เครื่องคิดเลย"
                style={popupStyle}
                className="grid gap-2 rounded-xl border border-border bg-card p-2.5 text-card-foreground shadow-2xl"
                onPointerDown={(event) => event.stopPropagation()}
              >
                <div
                  aria-live="polite"
                  className="rounded-lg border border-border bg-muted/40 px-2 py-1.5 text-right font-mono text-lg font-semibold text-foreground"
                >
                  {tokens.length > 0 ? (
                    <span className="mr-1 text-xs text-muted-foreground">
                      {tokens.join(" ").replace(/\*/g, "×").replace(/\//g, "÷")}
                    </span>
                  ) : null}
                  {entry}
                </div>
                <div className="grid grid-cols-4 gap-1.5">
                  {keypad.map((k, i) =>
                    keyButton(k.label, k.onPress, {
                      tone: k.tone,
                      aria: k.aria ?? (typeof k.label === "string" ? k.label : `key-${i}`),
                    }),
                  )}
                  {keyButton("C", clearAll, { tone: "muted", aria: "ล้างทั้งหมด" })}
                  {keyButton("=", equals, { tone: "op", aria: "เท่ากับ" })}
                  {keyButton("ตกลง", applyResult, { span: true, tone: "apply", aria: "ใช้ค่านี้" })}
                </div>
              </div>
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}
