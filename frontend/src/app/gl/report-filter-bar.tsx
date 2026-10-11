"use client";

// Collapsible report conditions, shared by the GL report screen and the central report screen: a toggle beside the heading,
// a one-line summary bar with the run / export / print actions while the form is folded, key hints on the buttons and the
// desktop shortcuts Alt+P (or Ctrl/⌘+P) print, Alt+E export, Alt+R show report (Champ parity).
import { useEffect, useRef } from "react";
import { ChevronDown, ChevronUp, Download, Filter, Printer, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { GLTextFn } from "@/lib/general-ledger";

type Shortcut = (() => void) | null;
type ShortcutActions = { print: Shortcut; exportCsv: Shortcut; refresh: Shortcut };

/** Alt+P / Ctrl+P print, Alt+E export, Alt+R show report; null = not a report key, or its action is unavailable right now. */
export function reportShortcut(event: Pick<KeyboardEvent, "key" | "altKey" | "ctrlKey" | "metaKey">, actions: ShortcutActions): Shortcut {
  const key = event.key.toLowerCase();
  if (key === "p" && (event.altKey || event.ctrlKey || event.metaKey)) return actions.print;
  if (key === "e" && event.altKey) return actions.exportCsv;
  if (key === "r" && event.altKey) return actions.refresh;
  return null;
}

/** Window shortcuts of a report screen; a null action leaves the key to the browser (e.g. Ctrl+P before any result). */
export function useReportShortcuts(actions: ShortcutActions) {
  const latest = useRef(actions);
  useEffect(() => { latest.current = actions; });
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      const action = reportShortcut(event, latest.current);
      if (!action) return;
      event.preventDefault();
      action();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);
}

/** Key hint inside an action button ("Enter", "Alt+P"); hidden on narrow screens. */
export function KeyHint({ keys, onPrimary = false }: { keys: string; onPrimary?: boolean }) {
  return <kbd className={`hidden sm:inline-block ml-1 text-[10px] font-sans font-normal opacity-70 ${onPrimary ? "bg-black/15 dark:bg-white/15" : "bg-muted"} px-1 rounded`}>{keys}</kbd>;
}

/** Heading button that folds / unfolds the condition form. */
export function ReportFilterToggle({ collapsed, onToggle, tr }: { collapsed: boolean; onToggle: () => void; tr: GLTextFn }) {
  const label = collapsed ? tr("gl_expand_all", "ขยายเงื่อนไข") : tr("gl_collapse_all", "ย่อเงื่อนไข");
  return <Button type="button" variant="outline" size="sm" className="h-8 gap-1.5 text-xs text-muted-foreground hover:text-foreground" onClick={onToggle} title={label}>
    <Filter className="size-3.5" />
    <span>{label}</span>
    {collapsed ? <ChevronDown className="size-3.5" /> : <ChevronUp className="size-3.5" />}
  </Button>;
}

/** Folded form: the applied conditions on one line plus the actions; an absent onExport / onPrint hides that button. */
export function ReportFilterSummary({ summary, busy, onExpand, onRun, onExport, onPrint, tr }: {
  summary: string;
  busy: boolean;
  onExpand: () => void;
  onRun: () => void;
  onExport?: () => void;
  onPrint?: () => void;
  tr: GLTextFn;
}) {
  return <div className="shrink-0 flex flex-wrap items-center justify-between gap-2 p-2.5 rounded-xl border border-border bg-muted/40 text-foreground">
    <div className="flex items-center gap-2 text-sm text-muted-foreground min-w-0">
      <Filter className="size-4 shrink-0 text-primary" />
      <span className="truncate font-medium text-foreground">{summary || tr("gl_report_all_conditions", "แสดงตามเงื่อนไขที่เลือก")}</span>
    </div>
    <div className="flex flex-wrap items-center gap-2">
      <Button type="button" variant="outline" size="sm" className="h-8 gap-1 text-xs" onClick={onExpand}>
        <ChevronDown className="size-3.5" />
        <span>{tr("gl_expand_all", "ขยายเงื่อนไข")}</span>
      </Button>
      <Button type="button" size="sm" className="h-8 gap-1 text-xs" disabled={busy} onClick={onRun} title="Enter / Alt+R">
        <RefreshCw className="size-3.5" />
        <span>{busy ? tr("gl_processing", "กำลังประมวลผล…") : tr("gl_show_report", "แสดงรายงาน")}</span>
      </Button>
      {onExport && <Button type="button" variant="outline" size="sm" className="h-8 gap-1 text-xs" disabled={busy} onClick={onExport} title="Alt+E">
        <Download className="size-3.5" />
        <span>{tr("gl_export_table", "ส่งออกตาราง")}</span>
      </Button>}
      {onPrint && <Button type="button" variant="outline" size="sm" className="h-8 gap-1 text-xs" disabled={busy} onClick={onPrint} title="Alt+P">
        <Printer className="size-3.5" />
        <span>{tr("gl_print", "พิมพ์")}</span>
      </Button>}
    </div>
  </div>;
}
