"use client";

import { Loader2, Search, X } from "lucide-react";
import { type RefObject, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { listMaster, type MasterEntry, type MasterName } from "@/lib/product-barcode/api";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import { type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import type { AuthSession } from "@/lib/workspace-models";

const FIELD_PICKER_VIEWPORT_MARGIN = 16;
const FIELD_PICKER_GAP = 4;
const FIELD_PICKER_MAX_WIDTH = 520;
const FIELD_PICKER_MIN_WIDTH = 420;

type FieldPickerPlacement = {
  left: number;
  top: number;
  width: number;
  maxHeight: number;
};

/**
 * Master-data picker.
 *
 * - Reuse for: group, brand, category, class, design, grade, model, pattern,
 *   unit, producttype, ordertype, businesstype, branch.
 * - Debounced server search (300ms).
 * - Returns the picked `MasterEntry` to caller via `onSelect`.
 * - Modal overlay; ESC closes; backdrop closes; tap row picks.
 *
 * Usage:
 *   <MasterPicker
 *     open={pickerOpen} onClose={() => setPickerOpen(false)}
 *     auth={auth} language={lang} master="brand"
 *     onSelect={(entry) => setBrand(entry)}
 *   />
 */
export interface MasterPickerProps {
  open: boolean;
  onClose: () => void;
  auth: AuthSession | null;
  language: LanguageCode | string;
  master: MasterName;
  title?: string;
  onSelect: (entry: MasterEntry) => void;
  /** Optional initial search keyword. */
  initialQuery?: string;
  placement?: "dialog" | "field";
  anchorRef?: RefObject<HTMLElement | null>;
}

export function MasterPicker({
  open,
  onClose,
  auth,
  language,
  master,
  title,
  onSelect,
  initialQuery = "",
  placement = "dialog",
  anchorRef,
}: MasterPickerProps) {
  const text = getBarcodeText(language);
  const [query, setQuery] = useState(initialQuery);
  const [debounced, setDebounced] = useState(initialQuery);
  const [items, setItems] = useState<MasterEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [fieldPlacement, setFieldPlacement] = useState<FieldPickerPlacement | null>(null);
  const fieldPanelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!open) return;
    setQuery(initialQuery);
    setDebounced(initialQuery);
  }, [open, initialQuery]);

  useEffect(() => {
    if (!open) return;
    const id = window.setTimeout(() => setDebounced(query), 300);
    return () => window.clearTimeout(id);
  }, [open, query]);

  const load = useCallback(async () => {
    if (!open) return;
    setLoading(true);
    setError("");
    const response = await listMaster(auth, master, { q: debounced, limit: 50, lang: String(language) });
    if (!response.success) {
      setError(response.message || text.requestFailed);
      setItems([]);
    } else {
      setItems(response.data ?? []);
    }
    setLoading(false);
  }, [open, auth, master, debounced, language, text.requestFailed]);

  useEffect(() => {
    if (open) void load();
  }, [open, load]);

  const updateFieldPlacement = useCallback(() => {
    if (!open || placement !== "field" || typeof window === "undefined") return;

    const margin = FIELD_PICKER_VIEWPORT_MARGIN;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const availableWidth = Math.max(0, viewportWidth - margin * 2);
    const anchorRect = anchorRef?.current?.getBoundingClientRect();
    const anchorWidth = anchorRect?.width ?? 0;
    const width = Math.min(
      Math.max(anchorWidth, FIELD_PICKER_MIN_WIDTH),
      FIELD_PICKER_MAX_WIDTH,
      availableWidth,
    );
    const preferredLeft = anchorRect?.left ?? margin;
    const maxLeft = Math.max(margin, viewportWidth - margin - width);
    const left = Math.min(Math.max(preferredLeft, margin), maxLeft);
    const top = Math.max(margin, (anchorRect?.bottom ?? margin) + FIELD_PICKER_GAP);
    const maxHeight = Math.max(96, viewportHeight - top - margin);

    setFieldPlacement((current) => {
      if (
        current &&
        current.left === left &&
        current.top === top &&
        current.width === width &&
        current.maxHeight === maxHeight
      ) {
        return current;
      }
      return { left, top, width, maxHeight };
    });
  }, [anchorRef, open, placement]);

  useLayoutEffect(() => {
    if (!open || placement !== "field") {
      setFieldPlacement(null);
      return;
    }

    updateFieldPlacement();

    let frameId = 0;
    const scheduleUpdate = () => {
      window.cancelAnimationFrame(frameId);
      frameId = window.requestAnimationFrame(updateFieldPlacement);
    };
    const anchorElement = anchorRef?.current;
    const resizeObserver =
      typeof ResizeObserver !== "undefined" && anchorElement ? new ResizeObserver(scheduleUpdate) : null;

    window.addEventListener("resize", scheduleUpdate);
    window.addEventListener("scroll", scheduleUpdate, true);
    resizeObserver?.observe(anchorElement as Element);

    return () => {
      window.cancelAnimationFrame(frameId);
      window.removeEventListener("resize", scheduleUpdate);
      window.removeEventListener("scroll", scheduleUpdate, true);
      resizeObserver?.disconnect();
    };
  }, [anchorRef, open, placement, updateFieldPlacement]);

  // ESC to close
  useEffect(() => {
    if (!open) return;
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onClose]);

  useEffect(() => {
    if (!open || placement !== "field") return;
    const handler = (event: PointerEvent) => {
      const target = event.target;
      if (!(target instanceof Node)) return;
      if (fieldPanelRef.current?.contains(target)) return;
      if (anchorRef?.current?.contains(target)) return;
      onClose();
    };
    window.addEventListener("pointerdown", handler, true);
    return () => window.removeEventListener("pointerdown", handler, true);
  }, [anchorRef, onClose, open, placement]);

  const headerTitle = useMemo(() => title ?? `${text.pickerSelect}: ${master}`, [title, text.pickerSelect, master]);
  const sortedItems = useMemo(() => {
    const locale = String(language).toLowerCase().startsWith("th") ? "th" : "en";
    const collator = new Intl.Collator(locale, { numeric: true, sensitivity: "base" });
    return [...items].sort((left, right) => {
      const leftName = pickName(left.names, language) || left.code || left.guidfixed;
      const rightName = pickName(right.names, language) || right.code || right.guidfixed;
      return collator.compare(leftName, rightName);
    });
  }, [items, language]);

  if (!open) return null;

  const pickerPanel = (
    <div
      className={cn(
        "flex w-full flex-col overflow-hidden rounded-lg bg-card text-card-foreground shadow-2xl",
        placement === "dialog" ? "max-h-[90vh] max-w-2xl" : "border border-border shadow-lg",
      )}
      style={placement === "field" && fieldPlacement ? { maxHeight: fieldPlacement.maxHeight } : undefined}
      onClick={(event) => event.stopPropagation()}
    >
      <div className="flex items-center justify-between border-b border-border px-3 py-2">
        <div className="text-sm font-semibold">{headerTitle}</div>
        <Button variant="ghost" size="sm" onClick={onClose} aria-label={text.pickerClose}>
          <X className="h-4 w-4" />
        </Button>
      </div>
      <div className="border-b border-border px-3 py-2">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" aria-hidden />
          <Input
            autoFocus
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={text.pickerSearch}
            className="h-9 pl-8"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {text.pickerLoading}
          </div>
        ) : error ? (
          <div className="p-4 text-sm text-destructive">{error}</div>
        ) : sortedItems.length === 0 ? (
          <div className="p-4 text-center text-sm text-muted-foreground">{text.pickerNoResult}</div>
        ) : (
          <ul className="divide-y divide-border">
            {sortedItems.map((entry) => (
              <li key={`${entry.guidfixed}:${entry.code}`}>
                <button
                  type="button"
                  onClick={() => {
                    onSelect(entry);
                    onClose();
                  }}
                  className={cn(
                    "grid min-h-9 w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-3 py-1.5 text-left transition hover:bg-muted/60",
                    "focus:bg-muted focus:outline-none",
                  )}
                >
                  <span className="min-w-0 truncate text-sm font-medium">
                    {pickName(entry.names, language) || entry.code || entry.guidfixed}
                  </span>
                  {entry.code ? (
                    <span className="shrink-0 text-xs font-medium text-muted-foreground">{entry.code}</span>
                  ) : null}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="border-t border-border px-3 py-1.5 text-right">
        <Button variant="outline" size="sm" onClick={onClose}>
          {text.pickerClose}
        </Button>
      </div>
    </div>
  );

  if (placement === "field") {
    return (
      <div
        ref={fieldPanelRef}
        className="fixed z-50"
        role="dialog"
        aria-modal="false"
        style={{
          left: fieldPlacement?.left ?? FIELD_PICKER_VIEWPORT_MARGIN,
          top: fieldPlacement?.top ?? FIELD_PICKER_VIEWPORT_MARGIN,
          width: fieldPlacement?.width ?? FIELD_PICKER_MIN_WIDTH,
          maxHeight: fieldPlacement?.maxHeight,
          visibility: fieldPlacement ? "visible" : "hidden",
        }}
      >
        {pickerPanel}
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-2 md:p-6"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
    >
      {pickerPanel}
    </div>
  );
}
