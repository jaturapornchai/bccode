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

  it("falls back to the local label before the key id when a backend translation is missing", () => {
    expect(menuText({ key: "unknown_key", th: "ตั้งค่า", en: "Settings" }, "th", readyDictionary({}))).toBe("ตั้งค่า");
  });

  it("does not show the active language menu slug when the backend dictionary is missing that key", () => {
    expect(menuText({ key: "active_languages", th: "ภาษาที่ใช้งาน", en: "Active Languages" }, "th", readyDictionary({}))).toBe("ภาษาที่ใช้งาน");
  });

  it("uses local label while backend language dictionary is still loading", () => {
    expect(menuText({ key: "settings", th: "ตั้งค่า", en: "Settings" }, "th", {})).toBe("ตั้งค่า");
  });

  it("keeps employee before form design", () => {
    const settingsGroup = MENU_SECTIONS.find((section) => section.id === "settings")?.groups.find((group) => group.id === "company-system");
    const itemIds = settingsGroup?.items.map((item) => item.id) ?? [];

    expect(itemIds.indexOf("employee")).toBeLessThan(itemIds.indexOf("form-design"));
  });

  it("hides sub-screens that were merged into the branch screen", () => {
    const settingsGroup = MENU_SECTIONS.find((section) => section.id === "settings")?.groups.find((group) => group.id === "company-system");
    const itemIds = settingsGroup?.items.map((item) => item.id) ?? [];

    expect(itemIds).not.toContain("department");
    expect(itemIds).not.toContain("workday");
    expect(itemIds).not.toContain("holiday");
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

  it("keeps product operations in product tools instead of the core product group", () => {
    const masterSection = MENU_SECTIONS.find((section) => section.id === "master");
    const productGroup = masterSection?.groups.find((group) => group.id === "products");
    const productToolsGroup = masterSection?.groups.find((group) => group.id === "product-tools");
    const productIds = productGroup?.items.map((item) => item.id) ?? [];
    const toolIds = productToolsGroup?.items.map((item) => item.id) ?? [];

    expect(productIds).not.toContain("price-history");
    expect(productIds).not.toContain("label-print");
    expect(productIds).not.toContain("add-product-branch");
    expect(productIds).not.toContain("add-product-department");
    expect(productToolsGroup?.title.th).toBe("เครื่องมือสินค้า");
    expect(toolIds).toEqual(["product-serial-registry", "price-history", "label-print"]);
  });

  it("splits product setup into user-focused groups instead of one long technical list", () => {
    const masterSection = MENU_SECTIONS.find((section) => section.id === "master");
    const groupsById = new Map(masterSection?.groups.map((group) => [group.id, group]) ?? []);

    expect(groupsById.get("products")?.title.th).toBe("สินค้าและบาร์โค้ด");
    expect(groupsById.get("products")?.items.map((item) => item.id)).toEqual([
      "product",
      "barcode",
      "productset",
      "product-unit",
    ]);
    const salesSettingIds = groupsById.get("sales-settings")?.items.map((item) => item.id) ?? [];
    expect(salesSettingIds).toContain("promotion");
    expect(salesSettingIds).toContain("channel-price");
    expect(groupsById.get("product-classification")?.title.th).toBe("จัดกลุ่มสินค้า");
    expect(groupsById.get("product-classification")?.items.map((item) => item.id)).toEqual([
      "product-group",
      "product-category",
      "product-category-list",
    ]);
    expect(groupsById.get("product-sku-options")?.title.th).toBe("สี ไซซ์ และตัวเลือก");
    expect(groupsById.get("product-sku-options")?.items.map((item) => item.id)).toEqual([
      "product-variant-matrix",
      "product-color",
      "product-size",
    ]);
    expect(groupsById.get("product-stock-production")?.items.map((item) => item.id)).toEqual([
      "warehouse",
      "bom",
    ]);
  });

  it("uses distinct product category labels for category structure and attribute category", () => {
    const masterSection = MENU_SECTIONS.find((section) => section.id === "master");
    const masterItems = masterSection?.groups.flatMap((group) => group.items) ?? [];
    const labelsById = new Map(masterItems.map((item) => [item.id, item.label.th]));

    expect(labelsById.get("product-category")).toBe("จัดหมวดสินค้า");
    expect(labelsById.get("product-category-list")).toBe("สินค้าในหมวด");
    expect(labelsById.get("category")).toBe("คุณลักษณะสินค้า");
    expect(masterItems.filter((item) => item.label.th === "หมวดสินค้า")).toHaveLength(0);
  });

  it("uses plain Thai business words for visible menu labels", () => {
    const visibleLabels = MENU_SECTIONS.flatMap((section) => [
      section.title.th,
      ...section.groups.flatMap((group) => [
        group.title.th,
        ...group.items.map((item) => item.label.th),
      ]),
    ]);
    const technicalWords = [
      "Dashboard",
      "Mappings",
      "Template",
      "route",
      "MCP Token",
      "AI Provider",
      "Matrix",
      "Schema",
      "Payload",
      "Raw JSON",
      "GUID",
      "Collection",
      "Slug",
    ];

    const offenders = visibleLabels.filter((label) => technicalWords.some((word) => label.includes(word)));
    expect(offenders).toEqual([]);
  });

  it("adds neutral SKU option masters for import-ready product variants", () => {
    const masterSection = MENU_SECTIONS.find((section) => section.id === "master");
    const skuGroup = masterSection?.groups.find((group) => group.id === "product-sku-options");
    const routesById = new Map(skuGroup?.items.map((item) => [item.id, item.route]) ?? []);

    expect(routesById.get("product-color")).toBe("/productcolor");
    expect(routesById.get("product-size")).toBe("/productsize");
    expect(routesById.get("product-variant-matrix")).toBe("/productvariantmatrix");
    expect(skuGroup?.items.find((item) => item.id === "product-variant-matrix")?.label.th).toBe("ชุดตัวเลือกสินค้า");
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
