"use client";

import {
  Plus,
  Settings2,
  Star,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { LanguageCode } from "@/lib/i18n";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { menuText, type MenuItem } from "@/lib/menu-data";
import { MenuRouteIcon } from "./menu-icon";
import { MenuPendingBadge } from "./menu-pending-badge";
import { ManageShortcutsScreen } from "./manage-shortcuts-screen";
import type { FrequentMenuEntry } from "@/lib/menu-usage";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import type { AuthSession } from "@/lib/workspace-models";
import {
  readUserShortcuts,
  userShortcutsStorageKey,
} from "@/lib/user-shortcuts";

/**
 * หน้าแรก = "งานของฉันวันนี้" ประกอบจากทางลัดตามสิทธิ์จอ (allowedMenuIds = union ชุดสิทธิ์)
 * วิดเจ็ตสรุปเอกสาร/ความเคลื่อนไหวล่าสุดถูกถอดออก 2026-09-23 พร้อมกับ API /transaction/{docType}/list ฝั่ง MongoDB
 * (ดูกฎ "ห้ามเพิ่ม MongoDB/Kafka/Redis/ClickHouse กลับมา" ใน AGENTS.md) — เหลือเฉพาะทางลัดของฉัน
 */

export function DashboardHome({
  auth,
  language,
  backendLanguage,
  allowedMenuIds,
  allMenuItems,
  frequentMenuEntries,
  onOpenItem,
  onOpenManageShortcuts,
}: {
  auth: AuthSession | null;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
  allowedMenuIds: Set<string>;
  allMenuItems: MenuItem[];
  frequentMenuEntries: FrequentMenuEntry[];
  onOpenItem: (item: MenuItem) => void;
  onOpenManageShortcuts?: () => void;
}) {
  const itemById = useMemo(() => new Map(allMenuItems.map((item) => [item.id, item])), [allMenuItems]);

  // ทางลัดเริ่มต้น (Smart Fallback: เมนูที่ใช้บ่อย หรือเมนูที่มีสิทธิ์ 8 รายการแรก)
  const defaultShortcutIds = useMemo(() => {
    const frequent = frequentMenuEntries.filter((e) => !isMenuScreenPending(e.item.route)).map((e) => e.item.id).filter((id) => allowedMenuIds.has(id));
    if (frequent.length >= 4) return frequent.slice(0, 8);
    // ข้ามจอที่ยังไม่พร้อม — ผู้ใช้ GL อย่างเดียวต้องไม่เจอทางลัดเริ่มต้นที่ขึ้น "รอพัฒนา"
    const fallback = allMenuItems
      .filter((item) => !isMenuScreenPending(item.route))
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

  const t = (key: string, th: string) => backendText(backendLanguage, key, th);

  return (
    <div className="grid min-w-0 gap-2" aria-label="overview">
      <section className="grid gap-1.5" aria-label="shortcuts-section">
        <div className="flex items-center justify-between gap-2">
          <h2 className="flex items-center gap-2 text-sm font-bold text-foreground">
            <Star className="size-4 text-primary" aria-hidden="true" />
            {t("menu_my_shortcuts", "ทางลัดของฉัน")}
            <span className="text-xs font-normal text-muted-foreground">({shortcuts.length})</span>
          </h2>
          <button
            type="button"
            onClick={handleOpenManage}
            className="inline-flex items-center gap-1.5 rounded-lg border border-border bg-card px-2.5 py-1 text-xs font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
          >
            <Settings2 className="size-3.5" aria-hidden="true" />
            {t("menu_manage_shortcuts_2", "จัดการทางลัด")}
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
                <MenuPendingBadge route={item.route} language={language} backendLanguage={backendLanguage} />
            </button>
          ))}
          <button
            type="button"
            onClick={handleOpenManage}
            className="group inline-flex items-center gap-1.5 rounded-lg border border-dashed border-border/90 bg-card/60 px-2.5 py-1.5 text-xs font-medium text-muted-foreground transition-all duration-200 hover:-translate-y-0.5 hover:border-primary hover:bg-primary/[0.05] hover:text-primary hover:shadow-xs active:translate-y-0"
            title={t("menu_add_or_customize_shortcuts", "เพิ่มหรือปรับแต่งทางลัด")}
          >
            <span className="grid size-6 shrink-0 place-items-center rounded-md bg-muted/60 text-muted-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
              <Plus className="size-3.5" aria-hidden="true" />
            </span>
            <span>{t("menu_add_shortcut", "เพิ่มทางลัด")}</span>
          </button>
        </div>
      </section>

      {shortcuts.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border bg-card p-6 text-center text-sm text-muted-foreground">
          {t("menu_no_screens_available_yet_ask_admin", "ยังไม่มีสิทธิ์เข้าจอใด — ติดต่อผู้ดูแลเพื่อรับสิทธิ์การใช้งาน")}
        </p>
      ) : null}
    </div>
  );
}
