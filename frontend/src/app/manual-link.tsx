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

export function ManualLink({ compact = false, label, language, screen }: ManualLinkProps) {
  const text = label ?? t(language, "manual");

  return (
    <Link
      aria-label={text}
      className={compact ? "icon-button manual-link compact" : "manual-link"}
      href={`/manual/${screen}?lang=${language}`}
      rel="noreferrer"
      target="_blank"
      title={text}
    >
      <BookOpen aria-hidden="true" size={18} />
      {compact ? null : <span>{text}</span>}
    </Link>
  );
}
