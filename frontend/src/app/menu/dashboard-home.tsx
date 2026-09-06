"use client";

import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  ClipboardList,
  Clock3,
  Plus,
  RotateCcw,
  Search,
  Settings2,
  Star,
  Trash2,
  X,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { authFetch } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { menuSearchMatches, menuText, type MenuItem } from "@/lib/menu-data";
import { MenuRouteIcon } from "./menu-icon";
import type { FrequentMenuEntry } from "@/lib/menu-usage";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import {
  addUserShortcut,
  clearUserShortcuts,
  moveUserShortcut,
  readUserShortcuts,
  removeUserShortcut,
  userShortcutsStorageKey,
  writeUserShortcuts,
} from "@/lib/user-shortcuts";

/**
 * หน้าแรก = "งานของฉันวันนี้" ประกอบจาก widget ตามสิทธิ์จอ (allowedMenuIds = union ชุดสิทธิ์)
 * แถว 1 เอกสารที่ดูแล (จำนวน + กดไปจอ) · แถว 2 ทางลัดที่ใช้บ่อย · แถว 3 ความเคลื่อนไหวล่าสุด
 * Widget แสดงเฉพาะเมื่อมีสิทธิ์เข้าจอนั้น (fail-closed) — ไม่มี = ไม่แสดง ไม่ใช่เลข 0
 */

// เอกสารที่มี list endpoint (mainapi root) — key = menu id ที่ใช้ตรวจสิทธิ์
const DOC_WIDGETS: { menuId: string; path: string; th: string; en: string; party: "custnames" | "creditornames" | null }[] = [
  { menuId: "quotation", path: "/transaction/quotation/list", th: "ใบเสนอราคา", en: "Quotations", party: "custnames" },
  { menuId: "sale-order", path: "/transaction/sale-order/list", th: "ใบสั่งขาย", en: "Sale orders", party: "custnames" },
  { menuId: "sale", path: "/transaction/sale-invoice/list", th: "ขายสินค้า", en: "Sale invoices", party: "custnames" },
  { menuId: "purchase-requisition", path: "/transaction/purchase-requisition/list", th: "ใบขอซื้อ", en: "Purchase requisitions", party: null },
  { menuId: "purchase-order", path: "/transaction/purchase-order/list", th: "ใบสั่งซื้อ", en: "Purchase orders", party: "creditornames" },
  { menuId: "purchase", path: "/transaction/purchase/list", th: "ซื้อสินค้า", en: "Purchases", party: "creditornames" },
  { menuId: "stock-transfer", path: "/transaction/stock-transfer/list", th: "โอนสินค้า", en: "Stock transfers", party: null },
  { menuId: "stock-adjust", path: "/transaction/stock-adjustment/list", th: "ปรับปรุงสต็อก", en: "Stock adjustments", party: null },
];

type DocRecord = Record<string, unknown>;
type DocStat = { total: number; latest: DocRecord[] };

function localizedName(value: unknown, language: LanguageCode): string {
  if (!Array.isArray(value)) return "";
  const list = value as { code?: string; name?: string }[];
  return list.find((n) => n.code === language && n.name)?.name ?? list.find((n) => n.name)?.name ?? "";
}

function dateText(value: unknown, language: LanguageCode): string {
  const d = typeof value === "string" ? new Date(value) : null;
  if (!d || Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString(language === "th" ? "th-TH" : "en-GB", { day: "numeric", month: "short" });
}

function money(value: unknown, language: LanguageCode): string {
  const n = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(n) || n === 0) return "";
  return n.toLocaleString(language === "th" ? "th-TH" : "en-US", { maximumFractionDigits: 2 });
}

export function DashboardHome({
  auth,
  workspace,
  language,
  backendLanguage,
  mainApiUrl,
  allowedMenuIds,
  allMenuItems,
  frequentMenuEntries,
  onOpenItem,
}: {
  auth: AuthSession | null;
  workspace: WorkspaceSession | null;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
  mainApiUrl: string;
  allowedMenuIds: Set<string>;
  allMenuItems: MenuItem[];
  frequentMenuEntries: FrequentMenuEntry[];
  onOpenItem: (item: MenuItem) => void;
}) {
  const isThai = language === "th";
  const itemById = useMemo(() => new Map(allMenuItems.map((item) => [item.id, item])), [allMenuItems]);
  const widgets = useMemo(() => DOC_WIDGETS.filter((w) => allowedMenuIds.has(w.menuId) && itemById.has(w.menuId)), [allowedMenuIds, itemById]);
  const [stats, setStats] = useState<Record<string, DocStat>>({});
  const [loading, setLoading] = useState(false);
  const branchCode = workspace?.branch?.code ?? "";
  const token = auth?.token ?? "";

  useEffect(() => {
    if (!token || !mainApiUrl || widgets.length === 0) {
      setStats({});
      return;
    }
    let cancelled = false;
    setLoading(true);
    (async () => {
      const next: Record<string, DocStat> = {};
      await Promise.all(
        widgets.map(async (w) => {
          try {
            const response = await authFetch(`${mainApiUrl}${w.path}?limit=5&offset=0&page=1&q=`, {
              headers: { Authorization: `Bearer ${token}` },
              cache: "no-store",
            });
            if (!response.ok) return;
            const payload = (await response.json()) as { data?: unknown; total?: unknown };
            const latest = Array.isArray(payload.data) ? (payload.data as DocRecord[]) : [];
            next[w.menuId] = { total: typeof payload.total === "number" ? payload.total : latest.length, latest };
          } catch {
            // widget เงียบเมื่อโหลดไม่ได้ — จอเข้าได้อยู่แล้วผ่านเมนู
          }
        }),
      );
      if (!cancelled) {
        setStats(next);
        setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token, mainApiUrl, widgets, branchCode]);

  // ทางลัดเริ่มต้น (Smart Fallback: เมนูที่ใช้บ่อย หรือเมนูที่มีสิทธิ์ 8 รายการแรก)
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
  const [isManageOpen, setIsManageOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    if (typeof window === "undefined") return;
    const stored = readUserShortcuts(window.localStorage, storageKey, allowedMenuIds);
    setCustomShortcutIds(stored);
  }, [storageKey, allowedMenuIds]);

  useEffect(() => {
    if (!isManageOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setIsManageOpen(false);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isManageOpen]);

  const activeShortcutIds = customShortcutIds ?? defaultShortcutIds;

  const shortcuts = useMemo(() => {
    return activeShortcutIds
      .map((id) => itemById.get(id))
      .filter((item): item is MenuItem => item !== undefined && allowedMenuIds.has(item.id));
  }, [activeShortcutIds, itemById, allowedMenuIds]);

  const handleAddShortcut = (menuId: string) => {
    const next = addUserShortcut(activeShortcutIds, menuId);
    setCustomShortcutIds(next);
    if (typeof window !== "undefined") {
      writeUserShortcuts(window.localStorage, storageKey, next);
    }
  };

  const handleRemoveShortcut = (menuId: string) => {
    const next = removeUserShortcut(activeShortcutIds, menuId);
    setCustomShortcutIds(next);
    if (typeof window !== "undefined") {
      writeUserShortcuts(window.localStorage, storageKey, next);
    }
  };

  const handleMoveShortcut = (fromIndex: number, toIndex: number) => {
    const next = moveUserShortcut(activeShortcutIds, fromIndex, toIndex);
    setCustomShortcutIds(next);
    if (typeof window !== "undefined") {
      writeUserShortcuts(window.localStorage, storageKey, next);
    }
  };

  const handleResetToDefault = () => {
    setCustomShortcutIds(null);
    if (typeof window !== "undefined") {
      clearUserShortcuts(window.localStorage, storageKey);
    }
  };

  const availableCandidates = useMemo(() => {
    const remaining = allMenuItems.filter(
      (item) => allowedMenuIds.has(item.id) && !activeShortcutIds.includes(item.id),
    );
    const query = searchQuery.trim();
    if (!query) return remaining.slice(0, 10);
    return remaining
      .filter((item) => menuSearchMatches(item.label, query, backendLanguage))
      .slice(0, 15);
  }, [allMenuItems, allowedMenuIds, activeShortcutIds, searchQuery, backendLanguage]);

  // ความเคลื่อนไหว: รวมเอกสารล่าสุดทุกประเภท เรียงวันที่
  const recent = useMemo(() => {
    const rows: { widget: (typeof DOC_WIDGETS)[number]; doc: DocRecord; at: number }[] = [];
    for (const w of widgets) {
      for (const doc of stats[w.menuId]?.latest ?? []) {
        const at = typeof doc.docdatetime === "string" ? Date.parse(doc.docdatetime) : 0;
        rows.push({ widget: w, doc, at: Number.isFinite(at) ? at : 0 });
      }
    }
    return rows.sort((a, b) => b.at - a.at).slice(0, 8);
  }, [widgets, stats]);

  const t = (th: string, en: string) => (isThai ? th : en);

  return (
    <div className="grid min-w-0 gap-3" aria-label="overview">
      {widgets.length > 0 ? (
        <section className="grid gap-2">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <ClipboardList className="size-4 text-primary" aria-hidden="true" />
            {t("เอกสารที่ดูแล", "Your documents")}
            <span className="text-xs font-normal text-muted-foreground">
              {workspace?.branch ? `· ${localizedName(workspace.branch.names, language) || workspace.branch.code}` : ""}
            </span>
          </h2>
          <div className="grid grid-cols-2 gap-2 md:grid-cols-4">
            {widgets.map((w) => {
              const item = itemById.get(w.menuId)!;
              const stat = stats[w.menuId];
              return (
                <button
                  key={w.menuId}
                  type="button"
                  onClick={() => onOpenItem(item)}
                  className="group grid min-w-0 gap-1 rounded-xl border border-border bg-card p-3 text-left shadow-sm transition-colors hover:border-primary/50 hover:bg-primary/5"
                >
                  <span className="truncate text-xs text-muted-foreground">{menuText(item.label, language, backendLanguage) || t(w.th, w.en)}</span>
                  <span className="text-2xl font-bold tabular-nums text-foreground">
                    {loading && !stat ? "…" : (stat?.total ?? 0).toLocaleString(isThai ? "th-TH" : "en-US")}
                  </span>
                  <span className="flex items-center gap-1 text-xs text-primary opacity-0 transition-opacity group-hover:opacity-100">
                    {t("เปิดจอ", "Open")} <ArrowRight className="size-3" aria-hidden="true" />
                  </span>
                </button>
              );
            })}
          </div>
        </section>
      ) : null}

      <section className="grid gap-2" aria-label="shortcuts-section">
        <div className="flex items-center justify-between gap-2">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <Star className="size-4 text-primary" aria-hidden="true" />
            {t("ทางลัดของฉัน", "My shortcuts")}
            <span className="text-xs font-normal text-muted-foreground">({shortcuts.length})</span>
          </h2>
          <button
            type="button"
            onClick={() => {
              setSearchQuery("");
              setIsManageOpen(true);
            }}
            className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-2.5 py-1 text-xs font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
          >
            <Settings2 className="size-3.5" aria-hidden="true" />
            {t("จัดการทางลัด", "Manage shortcuts")}
          </button>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          {shortcuts.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onOpenItem(item)}
              className="group inline-flex items-center gap-2.5 rounded-xl border border-border bg-card px-3 py-2 text-sm font-semibold text-foreground shadow-xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/60 hover:bg-primary/[0.04] hover:shadow-md hover:shadow-primary/10 active:translate-y-0 active:scale-[0.98]"
            >
              <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                <MenuRouteIcon item={item} size={15} />
              </span>
              <span className="truncate group-hover:text-primary transition-colors">
                {menuText(item.label, language, backendLanguage)}
              </span>
            </button>
          ))}
          <button
            type="button"
            onClick={() => {
              setSearchQuery("");
              setIsManageOpen(true);
            }}
            className="group inline-flex items-center gap-2 rounded-xl border border-dashed border-border/90 bg-card/60 px-3 py-2 text-sm font-medium text-muted-foreground transition-all duration-200 hover:-translate-y-0.5 hover:border-primary hover:bg-primary/[0.05] hover:text-primary hover:shadow-xs active:translate-y-0"
            title={t("เพิ่มหรือปรับแต่งทางลัด", "Add or customize shortcuts")}
          >
            <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-muted/60 text-muted-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
              <Plus className="size-4" aria-hidden="true" />
            </span>
            <span>{t("เพิ่มทางลัด", "Add shortcut")}</span>
          </button>
        </div>
      </section>

      {/* Modal Dialog: จัดการทางลัดของฉัน */}
      {isManageOpen ? (
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="manage-shortcuts-title"
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-xs"
          onClick={() => setIsManageOpen(false)}
        >
          <div
            className="flex max-h-[85vh] w-full max-w-2xl flex-col rounded-2xl border border-border bg-card shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Modal Header */}
            <div className="flex items-start justify-between border-b border-border p-4 pb-3">
              <div className="flex items-center gap-3">
                <span className="grid size-9 place-items-center rounded-xl bg-primary/10 text-primary">
                  <Star className="size-5 fill-primary/20" aria-hidden="true" />
                </span>
                <div>
                  <h3 id="manage-shortcuts-title" className="text-base font-bold text-foreground">
                    {t("จัดการทางลัดของฉัน", "Manage My Shortcuts")}
                  </h3>
                  <p className="text-xs text-muted-foreground">
                    {t(
                      "เพิ่ม ลบ หรือจัดลำดับเมนูที่คุณใช้งานบ่อย เพื่อเปิดเข้าใช้งานได้รวดเร็ว",
                      "Add, remove, or reorder menus you use frequently for fast access",
                    )}
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setIsManageOpen(false)}
                className="grid size-8 place-items-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                aria-label={t("ปิด", "Close")}
              >
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 space-y-4 overflow-y-auto p-4">
              {/* 1. Add Section */}
              <div className="space-y-2 rounded-xl border border-border bg-muted/20 p-3">
                <label htmlFor="shortcut-search" className="block text-xs font-semibold text-foreground">
                  {t("ค้นหาและเพิ่มเมนูเข้าทางลัด", "Search and add menu to shortcuts")}
                </label>
                <div className="relative">
                  <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
                  <input
                    id="shortcut-search"
                    type="search"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder={t("พิมพ์ชื่อเมนูเพื่อค้นหา... (เช่น สินค้า, บาร์โค้ด, ขาย, ซื้อ)", "Type menu name... (e.g. Product, Barcode, Sale)")}
                    className="h-9 w-full rounded-lg border border-border bg-background !pl-10 !pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                {/* Available Candidates */}
                <div className="space-y-1 pt-1">
                  <p className="text-[11px] font-medium text-muted-foreground">
                    {searchQuery.trim()
                      ? t(`ผลการค้นหา (${availableCandidates.length} รายการ)`, `Search results (${availableCandidates.length})`)
                      : t("เมนูแนะนำที่ยังไม่ได้เพิ่มเข้าทางลัด:", "Suggested menus:")}
                  </p>
                  {availableCandidates.length === 0 ? (
                    <p className="py-2 text-center text-xs text-muted-foreground">
                      {t("ไม่พบเมนูที่ตรงกับคำค้นหา", "No matching menus found")}
                    </p>
                  ) : (
                    <div className="grid max-h-36 grid-cols-1 gap-1 overflow-y-auto sm:grid-cols-2">
                      {availableCandidates.map((item) => (
                        <div
                          key={item.id}
                          className="flex items-center justify-between gap-2 rounded-xl border border-border bg-card px-2.5 py-1.5 text-xs shadow-xs transition-colors hover:border-primary/40 hover:bg-muted/20"
                        >
                          <div className="flex min-w-0 items-center gap-2">
                            <span className="grid size-6.5 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                              <MenuRouteIcon item={item} size={14} />
                            </span>
                            <span className="truncate font-medium text-foreground">
                              {menuText(item.label, language, backendLanguage)}
                            </span>
                          </div>
                          <button
                            type="button"
                            onClick={() => handleAddShortcut(item.id)}
                            className="inline-flex shrink-0 items-center gap-1 rounded-md bg-primary/10 px-2 py-1 text-[11px] font-semibold text-primary transition-colors hover:bg-primary hover:text-primary-foreground"
                          >
                            <Plus className="size-3" aria-hidden="true" />
                            {t("เพิ่ม", "Add")}
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* 2. Current Shortcuts List */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="text-xs font-semibold text-foreground">
                    {t(`ลำดับทางลัดปัจจุบัน (${shortcuts.length} รายการ)`, `Current Shortcuts (${shortcuts.length})`)}
                  </h4>
                  <span className="text-[11px] text-muted-foreground">
                    {t("กดลูกศรเพื่อย้ายตำแหน่ง หรือกดลบออก", "Use arrows to reorder or remove")}
                  </span>
                </div>

                {shortcuts.length === 0 ? (
                  <div className="rounded-xl border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
                    {t("ยังไม่มีรายการทางลัด — เลือกค้นหาและกดเพิ่มเมนูด้านบนได้เลยครับ", "No shortcuts yet — search and add menus above.")}
                  </div>
                ) : (
                  <div className="divide-y divide-border/60 rounded-xl border border-border bg-card overflow-hidden">
                    {shortcuts.map((item, index) => {
                      const isFirst = index === 0;
                      const isLast = index === shortcuts.length - 1;
                      return (
                        <div
                          key={item.id}
                          className="flex items-center justify-between gap-3 px-3 py-2 text-sm transition-colors hover:bg-muted/30"
                        >
                          <div className="flex min-w-0 items-center gap-2.5">
                            <span className="grid size-6 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-bold text-primary">
                              {index + 1}
                            </span>
                            <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                              <MenuRouteIcon item={item} size={15} />
                            </span>
                            <span className="truncate font-medium text-foreground">
                              {menuText(item.label, language, backendLanguage)}
                            </span>
                          </div>

                          <div className="flex items-center gap-1 shrink-0">
                            <button
                              type="button"
                              disabled={isFirst}
                              onClick={() => handleMoveShortcut(index, index - 1)}
                              title={t("เลื่อนขึ้น", "Move up")}
                              aria-label={t("เลื่อนขึ้น", "Move up")}
                              className="grid size-8 place-items-center rounded-lg border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:opacity-30 disabled:pointer-events-none"
                            >
                              <ArrowUp className="size-4" aria-hidden="true" />
                            </button>
                            <button
                              type="button"
                              disabled={isLast}
                              onClick={() => handleMoveShortcut(index, index + 1)}
                              title={t("เลื่อนลง", "Move down")}
                              aria-label={t("เลื่อนลง", "Move down")}
                              className="grid size-8 place-items-center rounded-lg border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:opacity-30 disabled:pointer-events-none"
                            >
                              <ArrowDown className="size-4" aria-hidden="true" />
                            </button>
                            <button
                              type="button"
                              onClick={() => handleRemoveShortcut(item.id)}
                              title={t("ลบออกจากทางลัด", "Remove from shortcuts")}
                              aria-label={t("ลบออกจากทางลัด", "Remove from shortcuts")}
                              className="inline-flex items-center gap-1 rounded-lg border border-border px-2 py-1.5 text-xs font-medium text-destructive transition-colors hover:border-destructive/50 hover:bg-destructive/10"
                            >
                              <Trash2 className="size-3.5" aria-hidden="true" />
                              <span className="hidden sm:inline">{t("ลบ", "Remove")}</span>
                            </button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            {/* Modal Footer */}
            <div className="flex items-center justify-between border-t border-border bg-muted/20 p-3">
              <button
                type="button"
                onClick={handleResetToDefault}
                className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              >
                <RotateCcw className="size-3.5" aria-hidden="true" />
                {t("รีเซ็ตเป็นค่าเริ่มต้น", "Reset to default")}
              </button>
              <button
                type="button"
                onClick={() => setIsManageOpen(false)}
                className="inline-flex items-center justify-center rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-sm transition-opacity hover:opacity-90"
              >
                {t("เสร็จสิ้น", "Done")}
              </button>
            </div>
          </div>
        </div>
      ) : null}

      {widgets.length > 0 ? (
        <section className="grid gap-2">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <Clock3 className="size-4 text-primary" aria-hidden="true" />
            {t("ความเคลื่อนไหวล่าสุด", "Recent activity")}
          </h2>
          {recent.length === 0 ? (
            <p className="rounded-xl border border-dashed border-border bg-card p-4 text-sm text-muted-foreground">
              {loading ? t("กำลังโหลด…", "Loading…") : t("ยังไม่มีเอกสาร — เริ่มจากทางลัดด้านบนได้เลย", "No documents yet — start from a shortcut above.")}
            </p>
          ) : (
            <ul className="grid gap-1 rounded-xl border border-border bg-card p-2 shadow-sm">
              {recent.map(({ widget, doc }, index) => {
                const item = itemById.get(widget.menuId)!;
                const party = widget.party ? localizedName(doc[widget.party], language) : "";
                return (
                  <li key={`${widget.menuId}-${String(doc.docno ?? index)}`}>
                    <button
                      type="button"
                      onClick={() => onOpenItem(item)}
                      className="grid w-full grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 rounded-lg px-2 py-1.5 text-left text-sm transition-colors hover:bg-primary/5"
                    >
                      <span className="w-12 shrink-0 text-xs text-muted-foreground">{dateText(doc.docdatetime, language)}</span>
                      <span className="min-w-0 truncate">
                        <b className="font-semibold">{t(widget.th, widget.en)}</b>
                        {doc.docno ? <span className="text-muted-foreground"> · {String(doc.docno)}</span> : null}
                        {party ? <span className="text-muted-foreground"> · {party}</span> : null}
                      </span>
                      <span className="text-xs font-semibold tabular-nums text-foreground">{money(doc.totalamount ?? doc.totalvalue, language)}</span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </section>
      ) : null}

      {widgets.length === 0 && shortcuts.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border bg-card p-6 text-center text-sm text-muted-foreground">
          {t("ยังไม่มีสิทธิ์เข้าจอใด — ติดต่อผู้ดูแลเพื่อรับสิทธิ์การใช้งาน", "No screens available yet — ask your administrator for permissions.")}
        </p>
      ) : null}
    </div>
  );
}
