"use client";

import Link from "next/link";
import { KeyRound } from "lucide-react";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import { getAuthSession } from "@/lib/client-auth-session";
import type { LanguageCode } from "@/lib/i18n";

/** Keep the destination discoverable; the screen and APIs enforce admin membership. */
export function MCPTokenAdminLink({ language, collapsed = false }: { language: LanguageCode; collapsed?: boolean }) {
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const title = backendText(dictionary, "mcp_title", "จัดการ API / MCP token");
  return <div className="col-span-2 sm:col-span-3 md:col-span-1 flex shrink-0 flex-col gap-1 mb-14">
    {!collapsed && <div className="hidden md:block px-2.5 pt-2 pb-1 text-[11px] font-bold text-muted-foreground/80 tracking-wide uppercase">{backendText(dictionary, "mcp_connections", "การเชื่อมต่อภายนอก")}</div>}
    <Link href="/mcp-tokens" title={collapsed ? title : undefined} aria-label={title}
    className={`col-span-2 sm:col-span-3 md:col-span-1 w-full min-h-11 gap-3 rounded-xl border border-transparent p-2.5 text-left text-sm font-bold text-foreground/90 transition-all duration-200 hover:border-border/70 hover:bg-card hover:text-foreground md:hover:translate-x-0.5 ${collapsed ? "hidden md:flex md:justify-center md:p-2" : "flex"} items-center`}>
    <span className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><KeyRound className="size-4" /></span>
    <span className={`min-w-0 flex-1 ${collapsed ? "md:hidden" : ""}`}><span className="block truncate leading-tight">{title}</span></span>
  </Link></div>;
}
