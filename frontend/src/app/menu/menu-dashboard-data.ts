import { MENU_SECTIONS, menuText, type MenuCategory, type MenuItem } from "@/lib/menu-data";
import type { LanguageCode } from "@/lib/i18n";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";

export type ErpMenuRow = {
  id: string;
  item: MenuItem;
  module: string;
  group: string;
  name: string;
  route: string;
  category: MenuCategory;
  status: "ready" | "planned" | "migration";
  activity: number;
  updatedAt: string;
};

export type KpiCardModel = {
  label: string;
  value: string;
  change: string;
  tone: "primary" | "success" | "warning" | "neutral";
};

export type ChartDatum = {
  name: string;
  value: number;
};

export function buildErpMenuRows(language: LanguageCode, dictionary?: BackendLanguageDictionary): ErpMenuRow[] {
  const today = new Date("2026-05-17T00:00:00+07:00");
  return MENU_SECTIONS.flatMap((section, sectionIndex) =>
    section.groups.flatMap((group, groupIndex) =>
      group.items.map((item, itemIndex) => {
        const age = sectionIndex * 8 + groupIndex + itemIndex;
        const updated = new Date(today);
        updated.setDate(today.getDate() - age);
        return {
          id: item.id,
          item,
          module: menuText(section.title, language, dictionary),
          group: menuText(group.title, language, dictionary),
          name: menuText(item.label, language, dictionary),
          route: item.route,
          category: item.category,
          status: item.route.startsWith("/transaction") ? "ready" : item.category === "report" ? "migration" : "planned",
          activity: Math.max(12, 98 - age),
          updatedAt: updated.toISOString().slice(0, 10),
        };
      }),
    ),
  );
}

export async function fetchErpMenuRows(language: LanguageCode, dictionary?: BackendLanguageDictionary): Promise<ErpMenuRow[]> {
  return buildErpMenuRows(language, dictionary);
}

export function buildKpis(rows: ErpMenuRow[], dictionary: BackendLanguageDictionary = {}): KpiCardModel[] {
  const ready = rows.filter((row) => row.status === "ready").length;
  const reports = rows.filter((row) => row.category === "report").length;
  const master = rows.filter((row) => row.category === "master").length;
  return [
    { label: backendText(dictionary, "menu_all"), value: rows.length.toLocaleString("th-TH"), change: backendText(dictionary, "from_legacy_flutter"), tone: "primary" },
    { label: backendText(dictionary, "ready_as_tab"), value: ready.toLocaleString("th-TH"), change: backendText(dictionary, "routine_group"), tone: "success" },
    { label: backendText(dictionary, "report"), value: reports.toLocaleString("th-TH"), change: backendText(dictionary, "waiting_report_screen_migration"), tone: "warning" },
    { label: backendText(dictionary, "master_data"), value: master.toLocaleString("th-TH"), change: backendText(dictionary, "erp_database"), tone: "neutral" },
  ];
}

export function buildChartData(rows: ErpMenuRow[], language: LanguageCode, dictionary?: BackendLanguageDictionary): ChartDatum[] {
  return MENU_SECTIONS.map((section) => ({
    name: menuText(section.title, language, dictionary),
    value: rows.filter((row) => row.module === menuText(section.title, language, dictionary)).length,
  }));
}
