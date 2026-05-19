"use client";

import { BookOpen } from "lucide-react";
import Link from "next/link";
import { t, type LanguageCode } from "@/lib/i18n";

type ManualLinkProps = {
  screen: string;
  language: LanguageCode;
  compact?: boolean;
};

export function ManualLink({ compact = false, language, screen }: ManualLinkProps) {
  return (
    <Link
      aria-label={t(language, "manual")}
      className={compact ? "icon-button manual-link compact" : "manual-link"}
      href={`/manual/${screen}?lang=${language}`}
      rel="noreferrer"
      target="_blank"
      title={t(language, "manual")}
    >
      <BookOpen aria-hidden="true" size={18} />
      {compact ? null : <span>{t(language, "manual")}</span>}
    </Link>
  );
}
