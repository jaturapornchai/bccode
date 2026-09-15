import type { LanguageCode } from "@/lib/i18n";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import { menuText } from "@/lib/menu-data";
import { isMenuScreenPending, isMenuDataPending } from "@/lib/menu-screen-status";

export function MenuPendingBadge({ route, language, backendLanguage }: {
  route: string;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
}) {
  const screenPending = isMenuScreenPending(route);
  // จอที่เปิดแล้วแต่ยังไม่มี API ป้อนข้อมูล ต้องบอกผู้ใช้ตรง ๆ ไม่ปล่อยให้เข้าไปเจอจอว่าง
  if (!screenPending && !isMenuDataPending(route)) return null;
  const label = screenPending
    ? menuText({ key: "menu_pending_development", th: "รอพัฒนา", en: "In development" }, language, backendLanguage)
    : menuText({ key: "menu_data_pending", th: "รอเชื่อมข้อมูล", en: "Awaiting data" }, language, backendLanguage);
  return (
    <span data-menu-pending={route} className="inline-flex shrink-0 whitespace-nowrap rounded-md border border-border bg-muted px-1.5 text-[0.9rem] leading-normal text-foreground" title={label}>
      {label}
    </span>
  );
}
