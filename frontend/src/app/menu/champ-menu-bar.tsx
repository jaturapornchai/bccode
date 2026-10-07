"use client";

import { BookOpen, KeyRound, LayoutDashboard, Link2, Lock, LogOut, Store, X, type LucideIcon } from "lucide-react";
import { useRef, useState, type KeyboardEvent } from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import { MENU_BARS, type MenuBar, type MenuBarActionId, type MenuBarNode } from "@/lib/menu-bars";
import { menuText, type MenuItem } from "@/lib/menu-data";
import { cn } from "@/lib/utils";
import { MenuRouteIcon } from "./menu-icon";
import { MenuPendingBadge } from "./menu-pending-badge";

const ACTION_ICONS: Record<MenuBarActionId, LucideIcon> = {
  "change-company": Store,
  "change-password": KeyRound,
  "exit-program": LogOut,
  "create-shortcut": Link2,
  "close-all-tabs": X,
  "open-manual": BookOpen,
};

// Heads must stay on ONE row at 1024px (13 heads + overview), so they are compact text buttons;
// `!` beats the unlayered `button { font-size: inherit }` reset (see tailwind-button-font-size-gotcha).
const HEAD_CLASS = "h-9 shrink-0 gap-1 px-2 text-[0.9rem]! data-[state=open]:bg-accent data-[state=open]:text-accent-foreground";
// Dropdown rows: ≥ 44px click target and ≥ 0.9rem Thai text (UX rule: คนไทย 40+).
const ROW_CLASS = "min-h-[44px] gap-2 px-2 py-1.5 text-[0.95rem] leading-snug data-[disabled]:cursor-not-allowed";
const PANEL_CLASS =
  "max-h-[min(80vh,var(--radix-dropdown-menu-content-available-height))] max-w-[min(30rem,calc(100vw-1rem))] overflow-y-auto! p-1.5";

export type ChampMenuBarProps = {
  activeSection: string;
  backendLanguage: BackendLanguageDictionary;
  bars?: MenuBar[];
  canAccessMenuItem: (item: MenuItem) => boolean;
  language: LanguageCode;
  navigationLabel: string;
  onAction: (action: MenuBarActionId) => void;
  onOpenItem: (item: MenuItem) => void;
  onOpenNewItem: (item: MenuItem) => void;
  onOpenOverview: () => void;
  onSelectSection: (barId: string) => void;
  overviewLabel: string;
};

/**
 * Top menu in the shape of the real Champ menu bar: one compact head per module, a single-column
 * dropdown with divider lines and nested submenus. Radix gives portal rendering (no clipping),
 * Esc / ↑↓ / → keyboard handling and collision-aware placement; here we add the menubar behaviours
 * Radix does not have: hovering another head while one is open switches to it, and ←/→ move
 * between heads. Ctrl/Cmd/Shift + click opens the screen in a new tab (replaces the old row button).
 */
export function ChampMenuBar({
  activeSection,
  backendLanguage,
  bars = MENU_BARS,
  canAccessMenuItem,
  language,
  navigationLabel,
  onAction,
  onOpenItem,
  onOpenNewItem,
  onOpenOverview,
  onSelectSection,
  overviewLabel,
}: ChampMenuBarProps) {
  const [openBarId, setOpenBarId] = useState<string | null>(null);
  const triggers = useRef(new Map<string, HTMLButtonElement>());
  // Switching heads must not hand focus back to the head that just closed.
  const keepFocus = useRef(false);
  const newTabRequested = useRef(false);

  function openBar(barId: string) {
    if (barId !== openBarId) keepFocus.current = openBarId !== null;
    onSelectSection(barId);
    setOpenBarId(barId);
  }

  function neighbour(barId: string, step: 1 | -1): MenuBar {
    const index = bars.findIndex((bar) => bar.id === barId);
    return bars[(index + step + bars.length) % bars.length];
  }

  function onHeadKeyDown(event: KeyboardEvent<HTMLButtonElement>, bar: MenuBar) {
    if (openBarId || (event.key !== "ArrowRight" && event.key !== "ArrowLeft")) return;
    event.preventDefault();
    triggers.current.get(neighbour(bar.id, event.key === "ArrowRight" ? 1 : -1).id)?.focus();
  }

  function onPanelKeyDown(event: KeyboardEvent<HTMLElement>, bar: MenuBar) {
    // Radix already handled (and prevented) ArrowRight on a submenu head and ArrowLeft inside a submenu.
    if (event.defaultPrevented || (event.key !== "ArrowRight" && event.key !== "ArrowLeft")) return;
    event.preventDefault();
    openBar(neighbour(bar.id, event.key === "ArrowRight" ? 1 : -1).id);
  }

  function renderNodes(nodes: MenuBarNode[], keyPrefix: string) {
    return nodes.map((node, index) => {
      if (node.kind === "divider") return <DropdownMenuSeparator key={`${keyPrefix}-divider-${index}`} />;

      if (node.kind === "submenu") {
        return (
          <DropdownMenuSub key={node.id}>
            <DropdownMenuSubTrigger className={ROW_CLASS}>{menuText(node.label, language, backendLanguage)}</DropdownMenuSubTrigger>
            <DropdownMenuPortal>
              <DropdownMenuSubContent className={cn("min-w-[17rem]", PANEL_CLASS)} collisionPadding={8}>
                {renderNodes(node.nodes, node.id)}
              </DropdownMenuSubContent>
            </DropdownMenuPortal>
          </DropdownMenuSub>
        );
      }

      if (node.kind === "action") {
        const Icon = ACTION_ICONS[node.id];
        return (
          <DropdownMenuItem className={ROW_CLASS} key={node.id} onSelect={() => onAction(node.id)}>
            <Icon aria-hidden="true" className="h-4 w-4 shrink-0 text-muted-foreground" />
            <span className="min-w-0 flex-1">{menuText(node.label, language, backendLanguage)}</span>
          </DropdownMenuItem>
        );
      }

      const { item } = node;
      const locked = !canAccessMenuItem(item);
      return (
        <DropdownMenuItem
          className={ROW_CLASS}
          disabled={locked}
          key={item.id}
          onClick={(event) => {
            newTabRequested.current = event.ctrlKey || event.metaKey || event.shiftKey;
          }}
          onSelect={() => {
            const newTab = newTabRequested.current;
            newTabRequested.current = false;
            if (newTab) onOpenNewItem(item);
            else onOpenItem(item);
          }}
        >
          <span className="grid h-7 w-7 shrink-0 place-items-center rounded-md bg-muted text-muted-foreground">
            <MenuRouteIcon item={item} size={15} />
          </span>
          <span className="line-clamp-2 min-w-0 flex-1 break-words font-medium">{menuText(item.label, language, backendLanguage)}</span>
          <MenuPendingBadge route={item.route} language={language} backendLanguage={backendLanguage} />
          {locked ? <Lock aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-muted-foreground" /> : null}
        </DropdownMenuItem>
      );
    });
  }

  return (
    <nav className="shrink-0 border-b border-border bg-card/95 px-2 py-1" aria-label={navigationLabel}>
      <div className="flex flex-wrap items-center gap-0.5">
        <Button
          type="button"
          variant={activeSection === "all" ? "secondary" : "ghost"}
          className={HEAD_CLASS}
          onClick={() => {
            setOpenBarId(null);
            onOpenOverview();
          }}
        >
          <LayoutDashboard aria-hidden="true" className="h-4 w-4" />
          {overviewLabel}
        </Button>
        {bars.map((bar) => (
          <DropdownMenu
            key={bar.id}
            modal={false}
            open={openBarId === bar.id}
            onOpenChange={(open) => {
              if (open) openBar(bar.id);
              else setOpenBarId((current) => (current === bar.id ? null : current));
            }}
          >
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant={activeSection === bar.id ? "secondary" : "ghost"}
                className={HEAD_CLASS}
                ref={(node) => {
                  if (node) triggers.current.set(bar.id, node);
                  else triggers.current.delete(bar.id);
                }}
                onKeyDown={(event) => onHeadKeyDown(event, bar)}
                onPointerEnter={() => {
                  if (openBarId && openBarId !== bar.id) openBar(bar.id);
                }}
              >
                {menuText(bar.label, language, backendLanguage)}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="start"
              className={cn("min-w-[17rem]", PANEL_CLASS)}
              collisionPadding={8}
              onCloseAutoFocus={(event) => {
                if (!keepFocus.current) return;
                keepFocus.current = false;
                event.preventDefault();
              }}
              onKeyDown={(event) => onPanelKeyDown(event, bar)}
            >
              {renderNodes(bar.nodes, bar.id)}
            </DropdownMenuContent>
          </DropdownMenu>
        ))}
      </div>
    </nav>
  );
}
