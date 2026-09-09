"use client";

import {
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  Check,
  CheckCircle2,
  Filter,
  Layers,
  Plus,
  RotateCcw,
  Search,
  Sparkles,
  Star,
  Trash2,
  X,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { LanguageCode } from "@/lib/i18n";
import {
  MENU_SECTIONS,
  menuSearchMatches,
  menuText,
  type MenuCategory,
  type MenuItem,
} from "@/lib/menu-data";
import { MenuRouteIcon } from "./menu-icon";
import type { FrequentMenuEntry } from "@/lib/menu-usage";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { AuthSession } from "@/lib/workspace-models";
import {
  addUserShortcut,
  clearUserShortcuts,
  moveUserShortcut,
  readUserShortcuts,
  removeUserShortcut,
  userShortcutsStorageKey,
  writeUserShortcuts,
} from "@/lib/user-shortcuts";
import { cn } from "@/lib/utils";

type CategoryFilter = "all" | MenuCategory;

interface CategoryTab {
  id: CategoryFilter;
  labelTh: string;
  labelEn: string;
}

const CATEGORY_TABS: CategoryTab[] = [
  { id: "all", labelTh: "ทั้งหมด", labelEn: "All" },
  { id: "transaction", labelTh: "งานประจำ (ซื้อ/ขาย/คลัง)", labelEn: "Operations" },
  { id: "master", labelTh: "ข้อมูลหลัก (สินค้า/คู่ค้า)", labelEn: "Master Data" },
  { id: "report", labelTh: "รายงาน", labelEn: "Reports" },
  { id: "finance", labelTh: "การเงินและบัญชี", labelEn: "Finance & Accounting" },
  { id: "settings", labelTh: "ตั้งค่าระบบ", labelEn: "Settings" },
];

export function ManageShortcutsScreen({
  auth,
  language,
  backendLanguage,
  allowedMenuIds,
  allMenuItems,
  frequentMenuEntries = [],
  onOpenItem,
  onBackToHome,
}: {
  auth: AuthSession | null;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
  allowedMenuIds: Set<string>;
  allMenuItems: MenuItem[];
  frequentMenuEntries?: FrequentMenuEntry[];
  onOpenItem?: (item: MenuItem) => void;
  onBackToHome?: () => void;
}) {
  const isThai = language === "th";
  const t = (th: string, en: string) => (isThai ? th : en);

  const itemById = useMemo(() => new Map(allMenuItems.map((item) => [item.id, item])), [allMenuItems]);

  // Default Smart Fallback shortcut IDs
  const defaultShortcutIds = useMemo(() => {
    const frequent = frequentMenuEntries.map((e) => e.item.id).filter((id) => allowedMenuIds.has(id));
    if (frequent.length >= 4) return frequent.slice(0, 8);
    const fallback = allMenuItems
      .map((item) => item.id)
      .filter((id) => allowedMenuIds.has(id) && !frequent.includes(id));
    return [...frequent, ...fallback].slice(0, 8);
  }, [frequentMenuEntries, allMenuItems, allowedMenuIds]);

  const storageKey = useMemo(() => userShortcutsStorageKey(auth), [auth]);
  const [customShortcutIds, setCustomShortcutIds] = useState<string[] | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeCategory, setActiveCategory] = useState<CategoryFilter>("all");
  const [onlyUnadded, setOnlyUnadded] = useState(false);

  // Read shortcuts from localStorage
  useEffect(() => {
    if (typeof window === "undefined") return;
    const loadShortcuts = () => {
      const stored = readUserShortcuts(window.localStorage, storageKey, allowedMenuIds);
      setCustomShortcutIds(stored);
    };
    loadShortcuts();

    const handleUpdate = () => loadShortcuts();
    window.addEventListener("bc_shortcuts_updated", handleUpdate);
    window.addEventListener("storage", handleUpdate);
    return () => {
      window.removeEventListener("bc_shortcuts_updated", handleUpdate);
      window.removeEventListener("storage", handleUpdate);
    };
  }, [storageKey, allowedMenuIds]);

  const activeShortcutIds = customShortcutIds ?? defaultShortcutIds;

  // Active shortcut items
  const shortcuts = useMemo(() => {
    return activeShortcutIds
      .map((id) => itemById.get(id))
      .filter((item): item is MenuItem => item !== undefined && allowedMenuIds.has(item.id));
  }, [activeShortcutIds, itemById, allowedMenuIds]);

  const shortcutIdSet = useMemo(() => new Set(activeShortcutIds), [activeShortcutIds]);

  const saveShortcuts = (nextIds: string[] | null) => {
    setCustomShortcutIds(nextIds);
    if (typeof window !== "undefined") {
      if (nextIds === null) {
        clearUserShortcuts(window.localStorage, storageKey);
      } else {
        writeUserShortcuts(window.localStorage, storageKey, nextIds);
      }
      window.dispatchEvent(new CustomEvent("bc_shortcuts_updated"));
    }
  };

  const handleAddShortcut = (menuId: string) => {
    const next = addUserShortcut(activeShortcutIds, menuId);
    saveShortcuts(next);
  };

  const handleRemoveShortcut = (menuId: string) => {
    const next = removeUserShortcut(activeShortcutIds, menuId);
    saveShortcuts(next);
  };

  const handleMoveShortcut = (fromIndex: number, toIndex: number) => {
    const next = moveUserShortcut(activeShortcutIds, fromIndex, toIndex);
    saveShortcuts(next);
  };

  const handleResetToDefault = () => {
    saveShortcuts(null);
  };

  // Filtered menu candidates for catalog view
  const { filteredItems, categoryCounts } = useMemo(() => {
    const allowed = allMenuItems.filter((item) => allowedMenuIds.has(item.id));

    // Calculate category counts
    const counts: Record<CategoryFilter, number> = {
      all: allowed.length,
      transaction: 0,
      master: 0,
      report: 0,
      finance: 0,
      settings: 0,
      restaurant: 0,
      approval: 0,
    };

    for (const item of allowed) {
      if (counts[item.category] !== undefined) {
        counts[item.category] += 1;
      }
    }

    const query = searchQuery.trim();
    let result = allowed;

    // Filter by category
    if (activeCategory !== "all") {
      result = result.filter((item) => item.category === activeCategory);
    }

    // Filter by unadded if checked
    if (onlyUnadded) {
      result = result.filter((item) => !shortcutIdSet.has(item.id));
    }

    // Filter by search query
    if (query) {
      result = result.filter((item) => menuSearchMatches(item.label, query, backendLanguage));
    }

    return { filteredItems: result, categoryCounts: counts };
  }, [allMenuItems, allowedMenuIds, activeCategory, onlyUnadded, searchQuery, shortcutIdSet, backendLanguage]);

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-4 pb-12 pt-1" aria-label="manage-shortcuts-page">
      {/* 1. Header Toolbar */}
      <header className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-4 shadow-xs sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3.5 min-w-0">
          <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary shadow-xs ring-1 ring-primary/20">
            <Star className="size-6 fill-primary/25 text-primary" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="truncate text-lg sm:text-xl font-bold text-foreground">
                {t("จัดการทางลัดของฉัน", "Manage My Shortcuts")}
              </h1>
              <span className="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
                {shortcuts.length} {t("รายการ", "items")}
              </span>
            </div>
            <p className="truncate text-xs text-muted-foreground sm:text-sm">
              {t(
                "เลือกเพิ่มหรือจัดลำดับเมนูที่คุณใช้งานบ่อย เพื่อเปิดทำงานได้รวดเร็วทันใจจากหน้าภาพรวม",
                "Customize and reorder your frequent menus for fast access from the overview dashboard",
              )}
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex shrink-0 items-center gap-2 pt-1 sm:pt-0">
          <button
            type="button"
            onClick={handleResetToDefault}
            title={t("รีเซ็ตทางลัดทั้งหมดเป็นค่าเริ่มต้นของระบบ", "Reset all shortcuts to system recommendations")}
            className="inline-flex h-9 items-center gap-1.5 rounded-xl border border-border bg-muted/30 px-3 text-xs font-medium text-muted-foreground transition-colors hover:border-primary/40 hover:bg-muted hover:text-foreground"
          >
            <RotateCcw className="size-3.5" aria-hidden="true" />
            <span>{t("รีเซ็ตค่าเริ่มต้น", "Reset Default")}</span>
          </button>
          {onBackToHome ? (
            <button
              type="button"
              onClick={onBackToHome}
              className="inline-flex h-9 items-center gap-1.5 rounded-xl border border-primary/30 bg-primary/10 px-3.5 text-xs font-semibold text-primary shadow-xs transition-colors hover:bg-primary hover:text-primary-foreground active:scale-[0.98]"
            >
              <ArrowLeft className="size-4" aria-hidden="true" />
              <span>{t("กลับหน้าภาพรวม", "Back to Overview")}</span>
            </button>
          ) : null}
        </div>
      </header>

      {/* 2. Live Preview Bar (ตัวอย่างแสดงผลจริงบนหน้าภาพรวม) */}
      <section className="rounded-2xl border border-primary/20 bg-primary/[0.02] p-3.5 shadow-xs">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between pb-2 border-b border-border/50">
          <div className="flex items-center gap-2 text-xs font-bold text-foreground">
            <Sparkles className="size-4 text-primary" aria-hidden="true" />
            <span>{t("ตัวอย่างแถบทางลัดบนหน้าแรก (แสดงผลทันทีตามที่คุณปรับแต่ง)", "Live Preview on Dashboard")}</span>
          </div>
          <span className="text-[11px] text-muted-foreground">
            {t("กดปุ่มลูกศรที่กล่องขวาเพื่อจัดเรียงตำแหน่ง", "Use arrows in the right panel to reorder")}
          </span>
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-2">
          {shortcuts.length === 0 ? (
            <p className="py-2 text-xs text-muted-foreground">
              {t("ยังไม่มีทางลัด — ค้นหาและกด + เพิ่มจากคลังเมนูด้านล่างได้เลยครับ", "No shortcuts yet — search and add from the catalog below.")}
            </p>
          ) : (
            shortcuts.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => onOpenItem?.(item)}
                title={t(`คลิกเพื่อเปิดหน้าจอ: ${menuText(item.label, language, backendLanguage)}`, `Click to open: ${menuText(item.label, language, backendLanguage)}`)}
                className="group inline-flex items-center gap-2.5 rounded-xl border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground shadow-2xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/60 hover:bg-primary/[0.04] hover:shadow-xs active:translate-y-0"
              >
                <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                  <MenuRouteIcon item={item} size={15} />
                </span>
                <span className="truncate group-hover:text-primary transition-colors">
                  {menuText(item.label, language, backendLanguage)}
                </span>
              </button>
            ))
          )}
        </div>
      </section>

      {/* 3. Main Workspace: Split View (Catalog Grid on Left, Current Shortcuts on Right) */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[1fr_360px] xl:grid-cols-[1fr_400px]">
        {/* ===================== LEFT COLUMN: ALL MENUS CATALOG ===================== */}
        <section className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-4 shadow-xs" aria-label="menu-catalog">
          {/* Catalog Title & Search Bar */}
          <div className="flex flex-col gap-2.5">
            <div className="flex items-center justify-between">
              <h2 className="flex items-center gap-2 text-base font-bold text-foreground">
                <Layers className="size-4.5 text-primary" aria-hidden="true" />
                {t("คลังเมนูทั้งหมดที่สามารถเพิ่มได้", "All Available Menus Catalog")}
              </h2>
              <span className="text-xs text-muted-foreground font-medium">
                {t(`แสดง ${filteredItems.length} จาก ${allowedMenuIds.size} เมนู`, `Showing ${filteredItems.length} of ${allowedMenuIds.size}`)}
              </span>
            </div>

            {/* Search Input (with !pl-10 trap prevention) */}
            <div className="relative">
              <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
              <input
                type="search"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t(
                  "พิมพ์ค้นหาชื่อเมนู, รหัสจอ หรือประเภทงาน... (เช่น ขาย, ซื้อ, ใบสั่งซื้อ, สินค้า, บาร์โค้ด)",
                  "Search menus by title, route, or category... (e.g. Sale, Purchase, Product, Barcode)",
                )}
                className="h-10 w-full rounded-xl border border-border bg-background !pl-10 !pr-10 text-sm font-medium text-foreground placeholder:text-muted-foreground/70 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 [&::-webkit-search-cancel-button]:appearance-none"
              />
              {searchQuery ? (
                <button
                  type="button"
                  onClick={() => setSearchQuery("")}
                  aria-label={t("ล้างการค้นหา", "Clear search")}
                  className="absolute right-2.5 top-1/2 grid size-6 -translate-y-1/2 place-items-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                >
                  <X className="size-3.5" aria-hidden="true" />
                </button>
              ) : null}
            </div>

            {/* Category Filter Pills (Horizontal wrap, uniform height) */}
            <div className="flex flex-wrap items-center gap-1.5 pt-0.5">
              {CATEGORY_TABS.map((cat) => {
                const isActive = activeCategory === cat.id;
                const count = categoryCounts[cat.id] ?? 0;
                return (
                  <button
                    key={cat.id}
                    type="button"
                    onClick={() => setActiveCategory(cat.id)}
                    className={cn(
                      "inline-flex h-8 items-center gap-1.5 rounded-lg px-2.5 text-xs font-medium transition-colors",
                      isActive
                        ? "border border-primary/50 bg-primary text-primary-foreground shadow-2xs"
                        : "border border-border bg-muted/30 text-muted-foreground hover:bg-muted hover:text-foreground",
                    )}
                  >
                    <span>{t(cat.labelTh, cat.labelEn)}</span>
                    <span
                      className={cn(
                        "grid h-4.5 min-w-4.5 place-items-center rounded-full px-1 text-[10px] font-bold",
                        isActive ? "bg-primary-foreground/25 text-primary-foreground" : "bg-muted text-muted-foreground",
                      )}
                    >
                      {count}
                    </span>
                  </button>
                );
              })}
            </div>

            {/* Filter Options & Stats sub-bar */}
            <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border/60 pt-2 text-xs">
              <label className="flex cursor-pointer items-center gap-2 text-muted-foreground hover:text-foreground select-none">
                <input
                  type="checkbox"
                  checked={onlyUnadded}
                  onChange={(e) => setOnlyUnadded(e.target.checked)}
                  className="size-3.5 rounded border-border text-primary focus:ring-primary/20"
                />
                <span className="font-medium">{t("แสดงเฉพาะเมนูที่ยังไม่ได้เพิ่มเข้าทางลัด", "Show only unadded menus")}</span>
              </label>

              <span className="text-[11px] text-muted-foreground">
                {t(
                  `อยู่ในทางลัดแล้ว ${activeShortcutIds.length} เมนู`,
                  `Already in shortcuts: ${activeShortcutIds.length}`,
                )}
              </span>
            </div>
          </div>

          {/* Catalog Menu Cards Grid (High Density Multi-column) */}
          {filteredItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border p-10 text-center text-muted-foreground">
              <Search className="size-8 stroke-[1.5] text-muted-foreground/50 mb-2" aria-hidden="true" />
              <p className="text-sm font-medium">{t("ไม่พบเมนูที่ตรงกับเงื่อนไขการค้นหา", "No matching menus found")}</p>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("ลองเปลี่ยนคำค้นหา หรือคลิกเลือกหมวดหมู่ 'ทั้งหมด'", "Try adjusting your search query or select 'All' categories")}
              </p>
              <button
                type="button"
                onClick={() => {
                  setSearchQuery("");
                  setActiveCategory("all");
                  setOnlyUnadded(false);
                }}
                className="mt-3 rounded-lg border border-border px-3 py-1.5 text-xs font-semibold text-primary hover:bg-muted"
              >
                {t("ล้างตัวกรองทั้งหมด", "Reset all filters")}
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">
              {filteredItems.map((item) => {
                const isAdded = shortcutIdSet.has(item.id);
                const title = menuText(item.label, language, backendLanguage);

                return (
                  <div
                    key={item.id}
                    className={cn(
                      "flex flex-col justify-between rounded-xl border p-2.5 transition-all duration-150",
                      isAdded
                        ? "border-primary/35 bg-primary/[0.03] shadow-2xs"
                        : "border-border bg-card/70 hover:border-primary/40 hover:bg-muted/30",
                    )}
                  >
                    <div className="flex items-start gap-2.5 min-w-0">
                      <span
                        className={cn(
                          "grid size-8 shrink-0 place-items-center rounded-lg text-sm",
                          isAdded ? "bg-primary text-primary-foreground" : "bg-primary/10 text-primary",
                        )}
                      >
                        <MenuRouteIcon item={item} size={16} />
                      </span>
                      <div className="min-w-0 flex-1">
                        <h3 className="truncate text-sm font-bold text-foreground" title={title}>
                          {title}
                        </h3>
                        <p className="truncate text-[11px] font-mono text-muted-foreground" title={item.route}>
                          {item.route}
                        </p>
                      </div>
                    </div>

                    {/* Card Action Row */}
                    <div className="mt-2.5 flex items-center justify-between border-t border-border/40 pt-2">
                      <span className="text-[10px] font-medium text-muted-foreground uppercase">
                        {item.category}
                      </span>

                      {isAdded ? (
                        <div className="flex items-center gap-1.5">
                          <span className="inline-flex items-center gap-1 rounded-md bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-bold text-emerald-600 dark:text-emerald-400">
                            <Check className="size-3" aria-hidden="true" />
                            {t("เพิ่มแล้ว", "Added")}
                          </span>
                          <button
                            type="button"
                            onClick={() => handleRemoveShortcut(item.id)}
                            title={t("ลบออกจากทางลัด", "Remove from shortcuts")}
                            className="inline-flex h-7 items-center gap-1 rounded-md border border-border px-2 text-[11px] font-medium text-destructive transition-colors hover:border-destructive/40 hover:bg-destructive/10"
                          >
                            <Trash2 className="size-3" aria-hidden="true" />
                            <span>{t("ลบ", "Remove")}</span>
                          </button>
                        </div>
                      ) : (
                        <button
                          type="button"
                          onClick={() => handleAddShortcut(item.id)}
                          className="inline-flex h-7.5 items-center gap-1 rounded-lg bg-primary/10 px-2.5 text-xs font-semibold text-primary transition-colors hover:bg-primary hover:text-primary-foreground active:scale-[0.97]"
                        >
                          <Plus className="size-3.5" aria-hidden="true" />
                          <span>{t("เพิ่มเป็นทางลัด", "Add")}</span>
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>

        {/* ===================== RIGHT COLUMN: CURRENT SHORTCUTS (STICKY) ===================== */}
        <aside className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-4 shadow-xs lg:sticky lg:top-3 lg:max-h-[calc(100vh-6rem)] lg:overflow-hidden" aria-label="current-shortcuts">
          <div className="flex items-center justify-between border-b border-border/70 pb-2.5">
            <div>
              <h2 className="flex items-center gap-2 text-base font-bold text-foreground">
                <Star className="size-4 text-primary fill-primary/30" aria-hidden="true" />
                {t("ทางลัดปัจจุบันของคุณ", "Your Shortcuts")}
              </h2>
              <p className="text-[11px] text-muted-foreground">
                {t("จัดลำดับตามความสะดวกด้วยปุ่ม ▲ / ▼", "Reorder with ▲ / ▼ buttons")}
              </p>
            </div>
            <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-bold text-primary">
              {shortcuts.length} {t("รายการ", "items")}
            </span>
          </div>

          {/* Reorderable List */}
          <div className="flex-1 space-y-1.5 overflow-y-auto pr-0.5">
            {shortcuts.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border p-6 text-center text-xs text-muted-foreground">
                {t("ยังไม่มีรายการทางลัด — เลือกกดปุ่ม '+ เพิ่มเป็นทางลัด' จากคลังด้านซ้ายได้เลยครับ", "No shortcuts yet — click '+ Add' from the left catalog.")}
              </div>
            ) : (
              shortcuts.map((item, index) => {
                const isFirst = index === 0;
                const isLast = index === shortcuts.length - 1;
                const title = menuText(item.label, language, backendLanguage);

                return (
                  <div
                    key={item.id}
                    className="flex items-center justify-between gap-2 rounded-xl border border-border/80 bg-background/60 p-2 shadow-2xs transition-colors hover:border-primary/50 hover:bg-muted/30"
                  >
                    <div className="flex min-w-0 items-center gap-2">
                      <span className="grid size-5.5 shrink-0 place-items-center rounded-full bg-primary/15 text-[11px] font-bold text-primary">
                        {index + 1}
                      </span>
                      <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                        <MenuRouteIcon item={item} size={14} />
                      </span>
                      <div className="min-w-0">
                        <span className="block truncate text-xs font-bold text-foreground" title={title}>
                          {title}
                        </span>
                        <span className="block truncate text-[10px] font-mono text-muted-foreground" title={item.route}>
                          {item.route}
                        </span>
                      </div>
                    </div>

                    {/* Up / Down / Delete Controls */}
                    <div className="flex shrink-0 items-center gap-1">
                      <button
                        type="button"
                        disabled={isFirst}
                        onClick={() => handleMoveShortcut(index, index - 1)}
                        title={t("เลื่อนขึ้น", "Move up")}
                        aria-label={t("เลื่อนขึ้น", "Move up")}
                        className="grid size-7.5 place-items-center rounded-lg border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:pointer-events-none disabled:opacity-25"
                      >
                        <ArrowUp className="size-3.5" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        disabled={isLast}
                        onClick={() => handleMoveShortcut(index, index + 1)}
                        title={t("เลื่อนลง", "Move down")}
                        aria-label={t("เลื่อนลง", "Move down")}
                        className="grid size-7.5 place-items-center rounded-lg border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:pointer-events-none disabled:opacity-25"
                      >
                        <ArrowDown className="size-3.5" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        onClick={() => handleRemoveShortcut(item.id)}
                        title={t("ลบออกจากทางลัด", "Remove from shortcuts")}
                        aria-label={t("ลบออกจากทางลัด", "Remove from shortcuts")}
                        className="grid size-7.5 place-items-center rounded-lg border border-border text-destructive transition-colors hover:border-destructive hover:bg-destructive/10"
                      >
                        <Trash2 className="size-3.5" aria-hidden="true" />
                      </button>
                    </div>
                  </div>
                );
              })
            )}
          </div>

          {/* Sticky Footer: Auto-save status & back button */}
          <div className="border-t border-border/70 pt-2.5 flex flex-col gap-2">
            <div className="flex items-center gap-1.5 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="size-3.5 shrink-0" aria-hidden="true" />
              <span>{t("บันทึกการเปลี่ยนแปลงอัตโนมัติแล้ว", "Changes auto-saved to device")}</span>
            </div>

            {onBackToHome ? (
              <button
                type="button"
                onClick={onBackToHome}
                className="inline-flex h-9 w-full items-center justify-center gap-2 rounded-xl bg-primary px-3 text-xs font-semibold text-primary-foreground shadow-xs transition-opacity hover:opacity-90 active:scale-[0.98]"
              >
                <ArrowLeft className="size-4" aria-hidden="true" />
                <span>{t("เสร็จสิ้นและกลับหน้าภาพรวม", "Done and Back to Overview")}</span>
              </button>
            ) : null}
          </div>
        </aside>
      </div>
    </div>
  );
}
