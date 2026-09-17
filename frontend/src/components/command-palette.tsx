"use client";

import * as React from "react";
import { createPortal } from "react-dom";
import { useRouter, usePathname } from "next/navigation";
import { Search, X, CornerDownLeft, Sparkles, FileText, Clock } from "lucide-react";
import { MENU_SECTIONS, type MenuItem, menuText, normalizeMenuSearchText } from "@/lib/menu-data";
import { cn } from "@/lib/utils";

export interface CommandItem extends MenuItem {
  sectionTitle: string;
  groupTitle: string;
  isRecent?: boolean;
}

export const RECENT_STORAGE_KEY = "bc_recent_screens";
export const MAX_RECENT_ITEMS = 8;

export function getStoredRecentRoutes(): string[] {
  if (typeof localStorage === "undefined") return [];
  try {
    const raw = localStorage.getItem(RECENT_STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function saveStoredRecentRoutes(routes: string[]) {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(RECENT_STORAGE_KEY, JSON.stringify(routes.slice(0, MAX_RECENT_ITEMS)));
  } catch {
    // Ignore storage errors
  }
}

export function CommandPalette() {
  const router = useRouter();
  const pathname = usePathname();
  const [mounted, setMounted] = React.useState(false);
  const [isOpen, setIsOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
  const [selectedIndex, setSelectedIndex] = React.useState(0);
  const [recentRoutes, setRecentRoutes] = React.useState<string[]>([]);

  const inputRef = React.useRef<HTMLInputElement>(null);
  const listRef = React.useRef<HTMLUListElement>(null);

  React.useEffect(() => {
    setMounted(true);
    setRecentRoutes(getStoredRecentRoutes());
  }, []);

  // Prepare all searchable menu items with their parent sections
  const allItems: CommandItem[] = React.useMemo(() => {
    const items: CommandItem[] = [];
    for (const section of MENU_SECTIONS) {
      const sectionTitle = section.title.th || section.title.en;
      for (const group of section.groups) {
        const groupTitle = group.title.th || group.title.en;
        for (const item of group.items) {
          items.push({
            ...item,
            sectionTitle,
            groupTitle,
          });
        }
      }
    }
    return items;
  }, []);

  // Track currently active page route and record into recent screens
  React.useEffect(() => {
    if (!pathname || pathname === "/" || allItems.length === 0) return;
    const isKnownMenu = allItems.some((item) => item.route === pathname);
    if (!isKnownMenu) return;

    setRecentRoutes((prev) => {
      const next = [pathname, ...prev.filter((r) => r !== pathname)].slice(0, MAX_RECENT_ITEMS);
      saveStoredRecentRoutes(next);
      return next;
    });
  }, [pathname, allItems]);

  // Map recent routes to command items
  const recentItems = React.useMemo(() => {
    return recentRoutes
      .map((route) => allItems.find((item) => item.route === route))
      .filter((item): item is CommandItem => Boolean(item))
      .map((item) => ({ ...item, isRecent: true }));
  }, [allItems, recentRoutes]);

  // Filter items based on query
  const filteredItems = React.useMemo(() => {
    if (!query.trim()) {
      if (recentItems.length > 0) {
        const recentRouteSet = new Set(recentRoutes);
        const popularRemaining = allItems
          .filter((item) => !recentRouteSet.has(item.route))
          .slice(0, 10);
        return [...recentItems, ...popularRemaining];
      }
      // When empty and no history, show top 15 frequently accessed menus
      return allItems.slice(0, 15);
    }

    const needle = normalizeMenuSearchText(query);
    const recentRouteSet = new Set(recentRoutes);

    return allItems
      .filter((item) => {
        const haystack = normalizeMenuSearchText(
          [
            item.label.th,
            item.label.en,
            item.route,
            item.sectionTitle,
            item.groupTitle,
            ...(item.label.aliases ?? []),
          ].join(" ")
        );
        return haystack.includes(needle);
      })
      .map((item) => ({
        ...item,
        isRecent: recentRouteSet.has(item.route),
      }))
      .slice(0, 30);
  }, [allItems, query, recentItems, recentRoutes]);

  // Reset selected index when results change
  React.useEffect(() => {
    setSelectedIndex(0);
  }, [filteredItems]);

  // Global keyboard shortcut: Ctrl+K / Cmd+K
  React.useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setIsOpen((prev) => !prev);
      }
    };

    window.addEventListener("keydown", handleGlobalKeyDown);
    return () => window.removeEventListener("keydown", handleGlobalKeyDown);
  }, []);

  // Auto focus input when opened
  React.useEffect(() => {
    if (isOpen) {
      setRecentRoutes(getStoredRecentRoutes());
      const timer = setTimeout(() => {
        inputRef.current?.focus();
        inputRef.current?.select();
      }, 50);
      return () => clearTimeout(timer);
    } else {
      setQuery("");
      setSelectedIndex(0);
    }
  }, [isOpen]);

  // Scroll active item into view
  React.useEffect(() => {
    if (!isOpen || selectedIndex < 0 || !listRef.current) return;
    const activeEl = listRef.current.children[selectedIndex] as HTMLElement | undefined;
    if (activeEl) {
      activeEl.scrollIntoView({ block: "nearest" });
    }
  }, [selectedIndex, isOpen]);

  const handleSelect = React.useCallback(
    (item: CommandItem) => {
      setRecentRoutes((prev) => {
        const next = [item.route, ...prev.filter((r) => r !== item.route)].slice(0, MAX_RECENT_ITEMS);
        saveStoredRecentRoutes(next);
        return next;
      });
      setIsOpen(false);
      router.push(item.route);
    },
    [router]
  );

  const handleClearHistory = React.useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    setRecentRoutes([]);
    saveStoredRecentRoutes([]);
  }, []);

  // Keyboard navigation within the palette
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      setIsOpen(false);
      return;
    }

    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev + 1 < filteredItems.length ? prev + 1 : 0));
      return;
    }

    if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev - 1 >= 0 ? prev - 1 : filteredItems.length - 1));
      return;
    }

    if (e.key === "Enter") {
      e.preventDefault();
      if (filteredItems[selectedIndex]) {
        handleSelect(filteredItems[selectedIndex]);
      }
      return;
    }
  };

  if (!mounted) return null;

  const showRecentHeader = !query.trim() && recentItems.length > 0;

  return (
    <>
      {/* Floating Shortcut Pill at Bottom Right for Touch / Mouse users */}
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        className="fixed bottom-3 right-3 z-40 flex items-center gap-2 rounded-full border border-border/80 bg-background/95 px-3 py-1.5 text-xs font-medium text-muted-foreground shadow-[0_4px_16px_rgba(0,0,0,0.12)] backdrop-blur-md transition-all hover:border-primary/80 hover:text-foreground hover:shadow-[0_6px_20px_rgba(0,0,0,0.18)]"
        aria-label="ค้นหาเมนูด่วน (Ctrl+K)"
      >
        <Search className="size-3.5 text-primary" />
        <span className="hidden sm:inline">เมนูด่วน</span>
        <kbd className="inline-flex items-center rounded border border-border bg-muted/60 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
          Ctrl K
        </kbd>
      </button>

      {/* Modal Dialog Portal */}
      {isOpen &&
        createPortal(
          <div
            className="fixed inset-0 z-[999999] flex items-start justify-center p-3 sm:p-6 md:pt-20 bg-background/80 backdrop-blur-sm animate-in fade-in-50 duration-150"
            onClick={(e) => {
              if (e.target === e.currentTarget) setIsOpen(false);
            }}
            role="dialog"
            aria-modal="true"
            aria-label="ค้นหาเมนูและหน้าจอ 226 รายการ"
          >
            <div
              className="relative flex w-full max-w-2xl flex-col overflow-hidden rounded-2xl border border-border/90 bg-card text-card-foreground shadow-[0_24px_60px_rgba(0,0,0,0.35),0_8px_24px_rgba(0,0,0,0.2)] dark:shadow-[0_24px_60px_rgba(0,0,0,0.9)] animate-in zoom-in-95 duration-150"
              onKeyDown={handleKeyDown}
            >
              {/* Header Search Input */}
              <div className="flex items-center gap-3 border-b border-border/80 bg-muted/20 px-4 py-3">
                <Search className="size-5 shrink-0 text-primary" aria-hidden="true" />
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="ค้นหาหน้าจอ, เมนู, หรือรายงาน 226 รายการ... (เช่น ผังบัญชี, ภาษี, กำไรขาดทุน)"
                  className="min-h-[2.4em] w-full bg-transparent text-[1rem] leading-normal text-foreground placeholder:text-muted-foreground focus:outline-none"
                  aria-label="ค้นหาเมนูและหน้าจอ"
                />
                {query ? (
                  <button
                    type="button"
                    onClick={() => {
                      setQuery("");
                      inputRef.current?.focus();
                    }}
                    className="inline-flex size-6 items-center justify-center rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
                    aria-label="ล้างคำค้นหา"
                  >
                    <X className="size-4" />
                  </button>
                ) : (
                  <kbd className="hidden sm:inline-flex items-center rounded border border-border bg-muted/60 px-2 py-0.5 font-mono text-xs text-muted-foreground">
                    ESC ปิด
                  </kbd>
                )}
              </div>

              {/* Section Header for Recent Screens */}
              {showRecentHeader && (
                <div className="flex items-center justify-between border-b border-border/50 bg-muted/30 px-4 py-1.5 text-xs font-semibold text-muted-foreground">
                  <div className="flex items-center gap-1.5">
                    <Clock className="size-3.5 text-primary" />
                    <span>หน้าที่เข้าชมล่าสุด</span>
                  </div>
                  <button
                    type="button"
                    onClick={handleClearHistory}
                    className="text-[11px] text-muted-foreground hover:text-destructive underline decoration-dotted transition-colors"
                  >
                    ล้างประวัติ
                  </button>
                </div>
              )}

              {/* Items List */}
              <ul
                ref={listRef}
                className="max-h-[60vh] overflow-y-auto p-2 divide-y divide-border/40 focus:outline-none"
                role="listbox"
                aria-label="ผลการค้นหาเมนู"
              >
                {filteredItems.length === 0 ? (
                  <li className="flex flex-col items-center justify-center py-10 text-center text-muted-foreground">
                    <Sparkles className="size-8 text-muted-foreground/40 mb-2" />
                    <p className="text-[0.95rem] font-medium text-foreground">ไม่พบเมนูหรือหน้าจอที่ตรงกับ &ldquo;{query}&rdquo;</p>
                    <p className="text-xs text-muted-foreground mt-1">ลองค้นหาด้วยคำอื่น เช่น รหัส, บัญชี, สต็อก หรือรายงาน</p>
                  </li>
                ) : (
                  filteredItems.map((item, index) => {
                    const isSelected = index === selectedIndex;
                    const title = menuText(item.label, "th");
                    const isFirstSuggested = showRecentHeader && index === recentItems.length;

                    return (
                      <React.Fragment key={`${item.id}-${index}`}>
                        {isFirstSuggested && (
                          <li className="px-3 pt-3 pb-1 text-xs font-semibold text-muted-foreground select-none list-none border-t border-border/50">
                            เมนูแนะนำ / ยอดนิยม
                          </li>
                        )}
                        <li
                          role="option"
                          aria-selected={isSelected}
                          onClick={() => handleSelect(item)}
                          onMouseEnter={() => setSelectedIndex(index)}
                          className={cn(
                            "group relative flex min-h-[3.2em] cursor-pointer items-center justify-between gap-3 rounded-xl px-3.5 py-2.5 text-[0.95rem] transition-colors select-none",
                            isSelected
                              ? "bg-primary text-primary-foreground shadow-sm"
                              : "hover:bg-muted/60 text-foreground"
                          )}
                        >
                          <div className="flex min-w-0 flex-1 items-center gap-3">
                            <span
                              className={cn(
                                "grid size-8 shrink-0 place-items-center rounded-lg border text-xs",
                                isSelected
                                  ? "border-primary-foreground/30 bg-primary-foreground/15 text-primary-foreground"
                                  : item.isRecent
                                    ? "border-primary/40 bg-primary/10 text-primary"
                                    : "border-border bg-muted/40 text-muted-foreground group-hover:border-primary/40 group-hover:text-primary"
                              )}
                            >
                              {item.isRecent ? (
                                <Clock className="size-4" />
                              ) : (
                                <FileText className="size-4" />
                              )}
                            </span>

                            <div className="flex flex-col min-w-0 flex-1">
                              <div className="flex items-center gap-2">
                                <span className="truncate font-medium leading-snug">{title}</span>
                                {item.isRecent && (
                                  <span
                                    className={cn(
                                      "inline-flex items-center rounded px-1.5 py-0.2 font-mono text-[10px]",
                                      isSelected
                                        ? "bg-primary-foreground/20 text-primary-foreground"
                                        : "bg-primary/10 text-primary"
                                    )}
                                  >
                                    ล่าสุด
                                  </span>
                                )}
                              </div>
                              <div className="flex items-center gap-1.5 text-xs opacity-80 truncate">
                                <span>{item.sectionTitle}</span>
                                <span>›</span>
                                <span>{item.groupTitle}</span>
                                <span className="font-mono opacity-60 ml-1">({item.route})</span>
                              </div>
                            </div>
                          </div>

                          {isSelected && (
                            <div className="flex items-center gap-1 shrink-0 text-xs font-medium opacity-90">
                              <span>เปิด</span>
                              <CornerDownLeft className="size-3.5" />
                            </div>
                          )}
                        </li>
                      </React.Fragment>
                    );
                  })
                )}
              </ul>

              {/* Footer Guide */}
              <div className="flex flex-wrap items-center justify-between border-t border-border/70 bg-muted/30 px-4 py-2 text-xs text-muted-foreground">
                <div className="flex items-center gap-3">
                  <span>
                    <kbd className="font-mono font-semibold">↑↓</kbd> นำทาง
                  </span>
                  <span>
                    <kbd className="font-mono font-semibold">↵</kbd> เพื่อเลือก
                  </span>
                  <span>
                    <kbd className="font-mono font-semibold">ESC</kbd> เพื่อปิด
                  </span>
                </div>
                <div>{filteredItems.length} รายการ</div>
              </div>
            </div>
          </div>,
          document.body
        )}
    </>
  );
}
