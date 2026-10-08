import type { LanguageCode } from "@/lib/i18n";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import { menuText } from "@/lib/menu-data";
import { isMenuScreenPending, isMenuDataPending } from "@/lib/menu-screen-status";

export function MenuPendingBadge({ route, language, backendLanguage }: {
  route: string;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
}) {
  // Champ parity: ในโปรแกรม Champ ไม่มี Badge "รอพัฒนา" หรือ "รอเชื่อมข้อมูล"
  // เมนูสะอาดและพร้อมใช้งานทุกจุด
  return null;
}
