import { Card, CardContent } from "@/components/ui/card";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import { menuText, type MenuItem } from "@/lib/menu-data";
import { HomeMenuIcon, MenuRouteIcon } from "./menu-icon";
import { MenuPendingBadge } from "./menu-pending-badge";

/**
 * "เมนูในแผนพัฒนา" card with the "รอพัฒนา" badge. Shown by the menu tab (main-menu-screen.tsx) and by the
 * standalone settings route (system-settings-screen.tsx) for screens that have no API yet — including the
 * ones whose MongoDB-era API was removed (isMenuBackendRetired in src/lib/menu-screen-status.ts).
 */
export function MenuPlannedCard({ route, item, title, language, backendLanguage }: {
  route: string;
  item?: MenuItem;
  title: string;
  language: LanguageCode;
  backendLanguage: BackendLanguageDictionary;
}) {
  return (
    <Card className="min-h-[420px]">
      <CardContent className="grid min-w-0 grid-cols-[minmax(0,1fr)] gap-4 p-5">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="flex min-w-0 gap-4">
            <span className="grid h-14 w-14 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              {item ? <MenuRouteIcon item={item} size={28} /> : <HomeMenuIcon size={28} />}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-medium text-muted-foreground">{menuText({ key: "menu_planned_workflow", th: "เมนูในแผนพัฒนา", en: "Planned Workflow" }, language, backendLanguage)}</p>
              <h2 className="break-words text-2xl font-semibold leading-relaxed">{item ? menuText(item.label, language, backendLanguage) : title}</h2>
            </div>
          </div>
          <MenuPendingBadge route={route} language={language} backendLanguage={backendLanguage} />
        </div>

        <div className="rounded-2xl border border-dashed border-border bg-muted/30 p-5 text-base leading-relaxed text-muted-foreground">
          {menuText({ key: "menu_planned_description", th: "หน้าจอนี้ยังอยู่ระหว่างเตรียมพัฒนา จึงยังบันทึกหรือประมวลผลข้อมูลไม่ได้ เลือกใช้งานเมนูอื่นจากแถบเมนูได้ตามปกติ", en: "This screen is planned and cannot save or process data yet. You can continue using other menus." }, language, backendLanguage)}
        </div>
      </CardContent>
    </Card>
  );
}
