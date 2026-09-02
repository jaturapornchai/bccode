"use client";

import { Search } from "lucide-react";
import { useMemo, useState } from "react";
import { Input } from "@/components/ui/input";
import type { LanguageCode } from "@/lib/i18n";
import { MENU_SECTIONS, menuText } from "@/lib/menu-data";
import {
  PERMISSION_ACTIONS,
  PERMISSION_ACTION_LABELS,
  actionEntry,
  isActionEntry,
  type PermissionAction,
} from "@/lib/permission-actions";
import { cn } from "@/lib/utils";

export type RoleScreenOption = {
  code: string;
  name: string;
  description: string;
  isActive: boolean;
};

type Filter = "all" | "selected" | "unselected";

const TEXT = {
  search: { th: "ค้นหาจอ รหัส หรือเส้นทาง", en: "Search screen, code or route" },
  all: { th: "ทั้งหมด", en: "All" },
  selected: { th: "เลือกแล้ว", en: "Selected" },
  unselected: { th: "ยังไม่เลือก", en: "Not selected" },
  screen: { th: "จอ", en: "Screen" },
  enter: { th: "เข้า", en: "Enter" },
  every: { th: "ทั้งหมด", en: "All" },
  noMatch: { th: "ไม่พบจอที่ตรงกับคำค้น", en: "No screens match" },
  other: { th: "อื่น ๆ", en: "Other" },
  visibleHint: { th: "ติ๊กที่หัวตารางเพื่อเลือกทุกจอที่แสดงอยู่", en: "Header checkboxes apply to the screens shown" },
};

/**
 * ตารางสิทธิ์ต่อจอ: แถว = จอ (จัดกลุ่มตามเมนู) คอลัมน์ = เข้า/เพิ่ม/แก้ไข/ลบ/ทั้งหมด
 * ค้นหาได้ กรองได้ และหัวตารางติ๊กทีเดียวกับทุกจอที่แสดงอยู่
 * รูปแบบค่า: "<code>" = เข้า, "<code>:create|update|delete", ดู lib/permission-actions.
 */
export function RoleScreenMatrix({
  language,
  onChange,
  options,
  readOnly,
  selected,
}: {
  language: LanguageCode;
  onChange: (next: string[]) => void;
  options: RoleScreenOption[];
  readOnly: boolean;
  selected: string[];
}) {
  const lang = language === "th" ? "th" : "en";
  const t = (key: keyof typeof TEXT) => TEXT[key][lang];
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<Filter>("all");
  const selectedSet = useMemo(() => new Set(selected), [selected]);

  // จอ → ชื่อกลุ่มเมนู (section › group) เพื่อจัดหมวดให้กวาดตาง่าย
  const sectionByCode = useMemo(() => {
    const map = new Map<string, string>();
    for (const section of MENU_SECTIONS) {
      for (const group of section.groups) {
        const title = `${menuText(section.title, language)} › ${menuText(group.title, language)}`;
        for (const item of group.items) map.set(item.id, title);
      }
    }
    return map;
  }, [language]);

  const groups = useMemo(() => {
    const needle = query.trim().toLowerCase();
    const rows = options.filter((option) => {
      const hasEntry = selectedSet.has(option.code);
      if (filter === "selected" && !hasEntry) return false;
      if (filter === "unselected" && hasEntry) return false;
      if (!needle) return true;
      return `${option.name} ${option.code} ${option.description}`.toLowerCase().includes(needle);
    });
    const byTitle = new Map<string, RoleScreenOption[]>();
    for (const row of rows) {
      const title = sectionByCode.get(row.code) ?? TEXT.other[lang];
      const list = byTitle.get(title);
      if (list) list.push(row);
      else byTitle.set(title, [row]);
    }
    return [...byTitle.entries()];
  }, [filter, lang, options, query, sectionByCode, selectedSet]);

  const visible = useMemo(() => groups.flatMap(([, rows]) => rows), [groups]);
  const selectedCount = selected.filter((entry) => !isActionEntry(entry)).length;

  const has = (code: string, action?: PermissionAction) =>
    selectedSet.has(action ? actionEntry(code, action) : code);
  const hasAll = (code: string) => has(code) && PERMISSION_ACTIONS.every((action) => has(code, action));

  function apply(mutate: (next: Set<string>) => void) {
    if (readOnly) return;
    const next = new Set(selected);
    mutate(next);
    onChange([...next].sort());
  }
  const grant = (next: Set<string>, code: string, action?: PermissionAction) => {
    next.add(code);
    if (action) next.add(actionEntry(code, action));
  };
  const revoke = (next: Set<string>, code: string, action?: PermissionAction) => {
    if (action) {
      next.delete(actionEntry(code, action));
      return;
    }
    next.delete(code);
    for (const each of PERMISSION_ACTIONS) next.delete(actionEntry(code, each));
  };
  const grantAll = (next: Set<string>, code: string) => {
    grant(next, code);
    for (const each of PERMISSION_ACTIONS) grant(next, code, each);
  };

  const setRow = (code: string, checked: boolean, action?: PermissionAction) =>
    apply((next) => (checked ? grant(next, code, action) : revoke(next, code, action)));
  const setRowAll = (code: string, checked: boolean) =>
    apply((next) => (checked ? grantAll(next, code) : revoke(next, code)));
  const setColumn = (checked: boolean, action?: PermissionAction) =>
    apply((next) => {
      for (const row of visible) {
        if (checked) grant(next, row.code, action);
        else revoke(next, row.code, action);
      }
    });
  const setColumnAll = (checked: boolean) =>
    apply((next) => {
      for (const row of visible) {
        if (checked) grantAll(next, row.code);
        else revoke(next, row.code);
      }
    });

  const columnChecked = (action?: PermissionAction) =>
    visible.length > 0 && visible.every((row) => has(row.code, action));
  const columnAllChecked = visible.length > 0 && visible.every((row) => hasAll(row.code));

  const box = "size-5 shrink-0 cursor-pointer accent-primary disabled:cursor-not-allowed";
  const cell = "px-2 py-2 text-center align-middle";

  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center gap-2">
        <label className="relative min-w-0 flex-1 basis-56">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="!pl-9 h-10 text-sm"
            placeholder={t("search")}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
        </label>
        <div className="flex gap-1" role="group">
          {(["all", "selected", "unselected"] as const).map((key) => (
            <button
              className={cn(
                "min-h-10 rounded-lg border px-3 text-sm font-semibold transition-colors",
                filter === key
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-border bg-card text-muted-foreground hover:bg-muted/60",
              )}
              key={key}
              onClick={() => setFilter(key)}
              type="button"
            >
              {t(key)}
              {key === "selected" ? ` (${selectedCount})` : ""}
            </button>
          ))}
        </div>
      </div>

      <div className="max-h-[60vh] overflow-auto rounded-xl border border-border bg-card">
        <table className="w-full border-collapse text-sm">
          <thead className="sticky top-0 z-10 bg-card shadow-[inset_0_-1px_0_var(--line)]">
            <tr className="text-muted-foreground">
              <th className="px-3 py-2 text-left text-sm font-bold text-foreground">
                {t("screen")}
                <span className="ml-2 font-normal">({visible.length})</span>
              </th>
              <th className={cell}>
                <label className="inline-grid cursor-pointer justify-items-center gap-1 text-xs font-bold text-foreground">
                  <input
                    className={box}
                    checked={columnChecked()}
                    disabled={readOnly || visible.length === 0}
                    onChange={(event) => setColumn(event.target.checked)}
                    title={t("visibleHint")}
                    type="checkbox"
                  />
                  {t("enter")}
                </label>
              </th>
              {PERMISSION_ACTIONS.map((action) => (
                <th className={cell} key={action}>
                  <label className="inline-grid cursor-pointer justify-items-center gap-1 text-xs font-bold text-foreground">
                    <input
                      className={box}
                      checked={columnChecked(action)}
                      disabled={readOnly || visible.length === 0}
                      onChange={(event) => setColumn(event.target.checked, action)}
                      title={t("visibleHint")}
                      type="checkbox"
                    />
                    {PERMISSION_ACTION_LABELS[action][lang]}
                  </label>
                </th>
              ))}
              <th className={cell}>
                <label className="inline-grid cursor-pointer justify-items-center gap-1 text-xs font-bold text-foreground">
                  <input
                    className={box}
                    checked={columnAllChecked}
                    disabled={readOnly || visible.length === 0}
                    onChange={(event) => setColumnAll(event.target.checked)}
                    title={t("visibleHint")}
                    type="checkbox"
                  />
                  {t("every")}
                </label>
              </th>
            </tr>
          </thead>
          <tbody>
            {groups.length === 0 ? (
              <tr>
                <td className="px-3 py-6 text-center text-muted-foreground" colSpan={6}>
                  {t("noMatch")}
                </td>
              </tr>
            ) : null}
            {groups.map(([title, rows]) => (
              <GroupRows
                box={box}
                cell={cell}
                has={has}
                hasAll={hasAll}
                key={title}
                readOnly={readOnly}
                rows={rows}
                setRow={setRow}
                setRowAll={setRowAll}
                title={title}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function GroupRows({
  box,
  cell,
  has,
  hasAll,
  readOnly,
  rows,
  setRow,
  setRowAll,
  title,
}: {
  box: string;
  cell: string;
  has: (code: string, action?: PermissionAction) => boolean;
  hasAll: (code: string) => boolean;
  readOnly: boolean;
  rows: RoleScreenOption[];
  setRow: (code: string, checked: boolean, action?: PermissionAction) => void;
  setRowAll: (code: string, checked: boolean) => void;
  title: string;
}) {
  return (
    <>
      <tr>
        <td
          className="bg-muted px-3 py-1.5 text-xs font-bold text-muted-foreground"
          colSpan={6}
        >
          {title}
        </td>
      </tr>
      {rows.map((row) => {
        const entered = has(row.code);
        return (
          <tr
            className={cn(
              "border-t border-border/60 transition-colors",
              entered ? "bg-primary/5" : "hover:bg-muted/40",
              !readOnly && "cursor-pointer",
              !row.isActive && "opacity-60",
            )}
            key={row.code}
            onClick={(event) => {
              // คลิกที่ชื่อ/พื้นที่ว่างของแถว = สลับ "เข้า" (ช่องติ๊กจัดการตัวเองอยู่แล้ว)
              if (readOnly || (event.target as HTMLElement).closest("input")) return;
              setRow(row.code, !entered);
            }}
          >
            <td className="px-3 py-2">
              <div className="grid gap-0.5">
                <span className={cn("text-sm font-semibold leading-snug", entered && "text-primary")}>{row.name || row.code}</span>
                <span className="text-xs leading-snug text-muted-foreground">
                  {row.code}
                  {row.description ? ` · ${row.description}` : ""}
                </span>
              </div>
            </td>
            <td className={cell}>
              <input
                aria-label={`${TEXT.enter.th} ${row.name || row.code}`}
                checked={entered}
                className={box}
                disabled={readOnly}
                onChange={(event) => setRow(row.code, event.target.checked)}
                type="checkbox"
              />
            </td>
            {PERMISSION_ACTIONS.map((action) => (
              <td className={cell} key={action}>
                <input
                  aria-label={`${PERMISSION_ACTION_LABELS[action].th} ${row.name || row.code}`}
                  checked={has(row.code, action)}
                  className={box}
                  disabled={readOnly}
                  onChange={(event) => setRow(row.code, event.target.checked, action)}
                  type="checkbox"
                />
              </td>
            ))}
            <td className={cell}>
              <input
                aria-label={`${TEXT.every.th} ${row.name || row.code}`}
                checked={hasAll(row.code)}
                className={box}
                disabled={readOnly}
                onChange={(event) => setRowAll(row.code, event.target.checked)}
                type="checkbox"
              />
            </td>
          </tr>
        );
      })}
    </>
  );
}
