import type { LanguageCode } from "@/lib/i18n";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import { menuText } from "@/lib/menu-data";
import { isMenuScreenPending } from "@/lib/menu-screen-status";

export function MenuPendingBadge({ route, language, backendLanguage }: {
  route: string;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
}) {
  if (!isMenuScreenPending(route)) return null;
  const label = menuText({ key: "menu_pending_development", th: "รอพัฒนา", en: "In development" }, language, backendLanguage);
  return (
    <span data-menu-pending={route} className="inline-flex shrink-0 whitespace-nowrap rounded-md border border-border bg-muted px-1.5 text-[0.9rem] leading-normal text-foreground" title={label}>
      {label}
    </span>
  );
}
