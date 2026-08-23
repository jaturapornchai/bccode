"use client";

import { BookOpen } from "lucide-react";
import Link from "next/link";
import { t, type LanguageCode } from "@/lib/i18n";

type ManualLinkProps = {
  screen: string;
  language: LanguageCode;
  compact?: boolean;
  label?: string;
};

const manualScreensWithContent = new Set([
  "activelanguages",
  "company",
  "currency",
  "menu",
  "permissiondefinition",
  "permissiongroup",
  "settings",
  "user",
  "useraccessaudit",
  "workspace",
]);

export function ManualLink({ compact = false, label, language, screen }: ManualLinkProps) {
  const text = label ?? t(language, "manual");
  const href = manualScreensWithContent.has(screen)
    ? `/manual/${screen}?lang=${language}`
    : `/manual?lang=${language}&screen=${encodeURIComponent(screen)}`;

  return (
    <Link
      aria-label={text}
      className={compact ? "icon-button manual-link compact" : "manual-link"}
      href={href}
      rel="noreferrer"
      target="_blank"
      title={text}
    >
      <BookOpen aria-hidden="true" size={18} />
      {compact ? null : <span>{text}</span>}
    </Link>
  );
}
