"use client";

import {
  ArrowRight,
  ClipboardList,
  Clock3,
  Plus,
  Settings2,
  Star,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { authFetch } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";
import { menuText, type MenuItem } from "@/lib/menu-data";
import { MenuRouteIcon } from "./menu-icon";
import { ManageShortcutsScreen } from "./manage-shortcuts-screen";
import type { FrequentMenuEntry } from "@/lib/menu-usage";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import {
  readUserShortcuts,
  userShortcutsStorageKey,
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
  onOpenManageShortcuts,
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
  onOpenManageShortcuts?: () => void;
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
  const [isManageMode, setIsManageMode] = useState(false);

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

  const shortcuts = useMemo(() => {
    return activeShortcutIds
      .map((id) => itemById.get(id))
      .filter((item): item is MenuItem => item !== undefined && allowedMenuIds.has(item.id));
  }, [activeShortcutIds, itemById, allowedMenuIds]);

  const handleOpenManage = () => {
    if (onOpenManageShortcuts) {
      onOpenManageShortcuts();
    } else {
      setIsManageMode(true);
    }
  };

  // If in standalone manage mode, render dedicated ManageShortcutsScreen
  if (isManageMode) {
    return (
      <ManageShortcutsScreen
        auth={auth}
        language={language}
        backendLanguage={backendLanguage}
        allowedMenuIds={allowedMenuIds}
        allMenuItems={allMenuItems}
        frequentMenuEntries={frequentMenuEntries}
        onOpenItem={onOpenItem}
        onBackToHome={() => setIsManageMode(false)}
      />
    );
  }

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
    <div className="grid min-w-0 gap-2" aria-label="overview">
      {widgets.length > 0 ? (
        <section className="grid gap-1.5">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <ClipboardList className="size-4 text-primary" aria-hidden="true" />
            {t("เอกสารที่ดูแล", "Your documents")}
            <span className="text-xs font-normal text-muted-foreground">
              {workspace?.branch ? `· ${localizedName(workspace.branch.names, language) || workspace.branch.code}` : ""}
            </span>
          </h2>
          <div className="grid grid-cols-2 gap-1.5 md:grid-cols-4">
            {widgets.map((w) => {
              const item = itemById.get(w.menuId)!;
              const stat = stats[w.menuId];
              return (
                <button
                  key={w.menuId}
                  type="button"
                  onClick={() => onOpenItem(item)}
                  className="group grid min-w-0 gap-0.5 rounded-xl border border-border bg-card p-2 text-left shadow-sm transition-colors hover:border-primary/50 hover:bg-primary/5"
                >
                  <span className="truncate text-xs text-muted-foreground">{menuText(item.label, language, backendLanguage) || t(w.th, w.en)}</span>
                  <span className="text-xl font-bold tabular-nums text-foreground">
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

      <section className="grid gap-1.5" aria-label="shortcuts-section">
        <div className="flex items-center justify-between gap-2">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <Star className="size-4 text-primary" aria-hidden="true" />
            {t("ทางลัดของฉัน", "My shortcuts")}
            <span className="text-xs font-normal text-muted-foreground">({shortcuts.length})</span>
          </h2>
          <button
            type="button"
            onClick={handleOpenManage}
            className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-2.5 py-1 text-xs font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
          >
            <Settings2 className="size-3.5" aria-hidden="true" />
            {t("จัดการทางลัด", "Manage shortcuts")}
          </button>
        </div>

        <div className="flex flex-wrap items-center gap-1.5">
          {shortcuts.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => onOpenItem(item)}
              className="group inline-flex items-center gap-2 rounded-lg border border-border bg-card px-2.5 py-1.5 text-xs font-semibold text-foreground shadow-xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/60 hover:bg-primary/[0.04] hover:shadow-md hover:shadow-primary/10 active:translate-y-0 active:scale-[0.98]"
            >
              <span className="grid size-6 shrink-0 place-items-center rounded-md bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                <MenuRouteIcon item={item} size={13} />
              </span>
              <span className="truncate group-hover:text-primary transition-colors">
                {menuText(item.label, language, backendLanguage)}
              </span>
            </button>
          ))}
          <button
            type="button"
            onClick={handleOpenManage}
            className="group inline-flex items-center gap-1.5 rounded-lg border border-dashed border-border/90 bg-card/60 px-2.5 py-1.5 text-xs font-medium text-muted-foreground transition-all duration-200 hover:-translate-y-0.5 hover:border-primary hover:bg-primary/[0.05] hover:text-primary hover:shadow-xs active:translate-y-0"
            title={t("เพิ่มหรือปรับแต่งทางลัด", "Add or customize shortcuts")}
          >
            <span className="grid size-6 shrink-0 place-items-center rounded-md bg-muted/60 text-muted-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
              <Plus className="size-3.5" aria-hidden="true" />
            </span>
            <span>{t("เพิ่มทางลัด", "Add shortcut")}</span>
          </button>
        </div>
      </section>

      {widgets.length > 0 ? (
        <section className="grid gap-1.5">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <Clock3 className="size-4 text-primary" aria-hidden="true" />
            {t("ความเคลื่อนไหวล่าสุด", "Recent activity")}
          </h2>
          {recent.length === 0 ? (
            <p className="rounded-xl border border-dashed border-border bg-card p-3 text-sm text-muted-foreground">
              {loading ? t("กำลังโหลด…", "Loading…") : t("ยังไม่มีเอกสาร — เริ่มจากทางลัดด้านบนได้เลย", "No documents yet — start from a shortcut above.")}
            </p>
          ) : (
            <ul className="grid gap-1 rounded-xl border border-border bg-card p-1.5 shadow-sm">
              {recent.map(({ widget, doc }, index) => {
                const item = itemById.get(widget.menuId)!;
                const party = widget.party ? localizedName(doc[widget.party], language) : "";
                return (
                  <li key={`${widget.menuId}-${String(doc.docno ?? index)}`}>
                    <button
                      type="button"
                      onClick={() => onOpenItem(item)}
                      className="grid w-full grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 rounded-lg px-2 py-1 text-left text-xs transition-colors hover:bg-primary/5"
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
