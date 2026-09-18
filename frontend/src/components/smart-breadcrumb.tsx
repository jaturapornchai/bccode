"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { ChevronRight, Copy, Check as CheckIcon, Home } from "lucide-react";
import Link from "next/link";
import { MENU_SECTIONS, menuText } from "@/lib/menu-data";
import type { LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export interface SmartBreadcrumbProps {
  currentTitle?: string;
  className?: string;
  language?: LanguageCode;
}

export function getBreadcrumbTrail(pathname: string, language: LanguageCode = "th") {
  if (!pathname || pathname === "/") return null;
  const targetPath = pathname.startsWith("/gl/journal") ? "/gl/journals" : pathname;

  for (const section of MENU_SECTIONS) {
    for (const group of section.groups) {
      for (const item of group.items) {
        if (item.route === targetPath || (item.route.length > 1 && targetPath.startsWith(item.route))) {
          return {
            section: menuText(section.title, language),
            group: menuText(group.title, language),
            item: menuText(item.label, language),
            route: item.route,
          };
        }
      }
    }
  }
  return null;
}

export function SmartBreadcrumb({ currentTitle, className, language = "th" }: SmartBreadcrumbProps) {
  const pathname = usePathname();
  const [copied, setCopied] = useState(false);
  const [currentLang, setCurrentLang] = useState<LanguageCode>(language);

  useEffect(() => {
    if (typeof document !== "undefined") {
      const docLang = document.documentElement.lang as LanguageCode;
      if (docLang) {
        setCurrentLang(docLang);
      }
    }
  }, []);

  const trail = getBreadcrumbTrail(pathname, currentLang);
  if (!trail && !currentTitle) return null;

  const handleCopyLink = async () => {
    try {
      if (typeof window !== "undefined" && navigator?.clipboard?.writeText) {
        await navigator.clipboard.writeText(window.location.href);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }
    } catch {
      // Fallback or ignore
    }
  };

  const homeLabel = currentLang === "en" ? "Home" : "หน้าแรก";
  const copyLabel = currentLang === "en" ? "Copy Link" : "คัดลอกลิงก์";
  const copiedLabel = currentLang === "en" ? "Copied!" : "คัดลอกแล้ว!";

  return (
    <nav
      aria-label="Breadcrumb"
      className={cn(
        "flex flex-wrap items-center justify-between gap-2 rounded-xl border border-border/70 bg-card/60 px-3.5 py-2 text-[0.88rem] text-muted-foreground shadow-xs backdrop-blur-xs",
        className
      )}
    >
      <ol className="flex flex-wrap items-center gap-1.5 min-w-0">
        <li className="inline-flex items-center gap-1.5">
          <Link
            href="/"
            className="inline-flex items-center gap-1 text-foreground/80 hover:text-primary transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring rounded px-1"
          >
            <Home className="size-3.5 shrink-0" />
            <span className="hidden sm:inline font-medium">{homeLabel}</span>
          </Link>
        </li>

        {trail?.section && (
          <li className="inline-flex items-center gap-1.5">
            <ChevronRight className="size-3.5 shrink-0 text-muted-foreground/60" aria-hidden="true" />
            <span className="truncate max-w-[140px] sm:max-w-none">{trail.section}</span>
          </li>
        )}

        {trail?.group && (
          <li className="inline-flex items-center gap-1.5">
            <ChevronRight className="size-3.5 shrink-0 text-muted-foreground/60" aria-hidden="true" />
            <span className="truncate max-w-[140px] sm:max-w-none">{trail.group}</span>
          </li>
        )}

        {(currentTitle || trail?.item) && (
          <li className="inline-flex items-center gap-1.5 font-semibold text-foreground">
            <ChevronRight className="size-3.5 shrink-0 text-muted-foreground/60" aria-hidden="true" />
            <span className="truncate max-w-[200px] sm:max-w-none" aria-current="page">
              {currentTitle || trail?.item}
            </span>
          </li>
        )}
      </ol>

      <button
        type="button"
        onClick={() => void handleCopyLink()}
        title={copyLabel}
        className={cn(
          "inline-flex shrink-0 items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs font-medium transition-all duration-150 cursor-pointer",
          copied
            ? "border-emerald-500/50 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 shadow-xs"
            : "border-border/80 bg-background/80 text-muted-foreground hover:bg-accent/60 hover:text-foreground shadow-2xs"
        )}
      >
        {copied ? (
          <>
            <CheckIcon className="size-3.5 text-emerald-500" />
            <span>{copiedLabel}</span>
          </>
        ) : (
          <>
            <Copy className="size-3.5" />
            <span className="hidden sm:inline">{copyLabel}</span>
          </>
        )}
      </button>
    </nav>
  );
}
