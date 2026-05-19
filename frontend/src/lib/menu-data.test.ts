import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { MENU_SECTIONS, flattenMenuItems, menuText } from "./menu-data";
import type { BackendLanguageDictionary } from "./backend-language";

function readyDictionary(values: BackendLanguageDictionary): BackendLanguageDictionary {
  Object.defineProperty(values, "__backendLanguageReady", { enumerable: false, value: "true" });
  return values;
}

function backendLanguageKeys(): Set<string> {
  const path = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");
  return new Set(
    readFileSync(path, "utf8")
      .split(/\r?\n/)
      .filter((line) => line && !line.startsWith("#"))
      .map((line) => line.split("\t")[0] ?? ""),
  );
}

function backendLanguageRows(): Map<string, Record<string, string>> {
  const path = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");
  const [headerLine = "", ...lines] = readFileSync(path, "utf8").split(/\r?\n/);
  const headers = headerLine.split("\t");
  const rows = new Map<string, Record<string, string>>();

  for (const line of lines) {
    if (!line || line.startsWith("#")) continue;
    const columns = line.split("\t");
    const key = columns[0] ?? "";
    rows.set(
      key,
      Object.fromEntries(headers.map((header, index) => [header, columns[index] ?? ""])),
    );
  }
  return rows;
}

function collectMenuLanguageKeys(): string[] {
  const keys = new Set<string>();
  for (const section of MENU_SECTIONS) {
    if (section.title.key) keys.add(section.title.key);
    for (const group of section.groups) {
      if (group.title.key) keys.add(group.title.key);
      for (const item of group.items) {
        if (item.label.key) keys.add(item.label.key);
      }
    }
  }

  const files = [
    "src/app/menu/main-menu-screen.tsx",
    "src/app/menu/menu-data-table.tsx",
    "src/app/menu/menu-kpi-chart.tsx",
    "src/app/menu/menu-dashboard-data.ts",
  ];
  for (const file of files) {
    const source = readFileSync(resolve(process.cwd(), file), "utf8");
    for (const match of source.matchAll(/backendText\([^,]+,\s*"([^"]+)"/g)) {
      keys.add(match[1] ?? "");
    }
    for (const match of source.matchAll(/"(close_tab|dashboard|frequent_menu|hide_menu|erp_navigation|no_frequent_menu|open_tabs|no_menu_found|overview|erp_overview|search_menu_document_route|show_menu|items)"/g)) {
      keys.add(match[1] ?? "");
    }
  }

  return [...keys].filter(Boolean).sort();
}

describe("menu language labels", () => {
  it("uses backend language dictionary by language key", () => {
    expect(
      menuText(
        { key: "system_settings", th: "ตั้งค่าระบบ", en: "System Settings" },
        "vi",
        readyDictionary({ system_settings: "Cài đặt hệ thống" }),
      ),
    ).toBe("Cài đặt hệ thống");
  });

  it("falls back to the key id when a backend translation is missing", () => {
    expect(menuText({ key: "unknown_key", th: "ตั้งค่า", en: "Settings" }, "th", readyDictionary({}))).toBe("unknown_key");
  });

  it("uses local label while backend language dictionary is still loading", () => {
    expect(menuText({ key: "settings", th: "ตั้งค่า", en: "Settings" }, "th", {})).toBe("ตั้งค่า");
  });

  it("keeps employee outside access control and before form design", () => {
    const settingsGroup = MENU_SECTIONS.find((section) => section.id === "settings")?.groups.find((group) => group.id === "company-system");
    const itemIds = settingsGroup?.items.map((item) => item.id) ?? [];

    expect(itemIds.indexOf("employee")).toBeGreaterThan(itemIds.indexOf("holiday"));
    expect(itemIds.indexOf("employee")).toBeLessThan(itemIds.indexOf("form-design"));
  });

  it("has backend language keys for every menu item", () => {
    const keys = backendLanguageKeys();
    const missing = MENU_SECTIONS.flatMap((section) =>
      section.groups.flatMap((group) =>
        group.items.map((item) => item.label.key ?? "").filter((key) => key && !keys.has(key)),
      ),
    );

    expect(missing).toEqual([]);
  });

  it("keeps menu item ids unique for permission codes", () => {
    const ids = flattenMenuItems().map((item) => item.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("has backend language keys for every keyed menu section and group", () => {
    const keys = backendLanguageKeys();
    const missing = MENU_SECTIONS.flatMap((section) => [
      section.title.key ?? "",
      ...section.groups.map((group) => group.title.key ?? ""),
    ]).filter((key) => key && !keys.has(key));

    expect(missing).toEqual([]);
  });

  it("has Lao translations for visible menu shell labels", () => {
    const rows = backendLanguageRows();
    expect(rows.get("overview")?.lo).toBe("ພາບລວມ");
    expect(rows.get("operations")?.lo).toBe("ວຽກປະຈຳ");
    expect(rows.get("report")?.lo).toBe("ລາຍງານ");
    expect(rows.get("master_data")?.lo).toBe("ຂໍ້ມູນຫຼັກ");
    expect(rows.get("settings")?.lo).toBe("ການຕັ້ງຄ່າ");
    expect(rows.get("frequent_menu")?.lo).toBe("ໃຊ້ປະຈໍາ");
    expect(rows.get("search_menu_document_route")?.lo).toContain("ຄົ້ນຫາເມນູ");
  });

  it("has all supported language cells for menu labels, placeholders, hints, and helper text", () => {
    const rows = backendLanguageRows();
    const languageColumns = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"];
    const missing = collectMenuLanguageKeys().flatMap((key) => {
      const row = rows.get(key);
      if (!row) return [`${key}:row`];
      return languageColumns
        .filter((language) => !row[language]?.trim())
        .map((language) => `${key}:${language}`);
    });

    expect(missing).toEqual([]);
  });
});
