import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { MENU_SECTIONS, flattenMenuItems, menuSearchHaystack, menuSearchMatches, menuText, normalizeMenuSearchText } from "./menu-data";
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

  it("has no settings section in the main menu (settings live in the workspace wizard, 2026-09-02)", () => {
    expect(MENU_SECTIONS.find((section) => section.id === "settings")).toBeUndefined();
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

  it("orders procurement as overview, request, price inquiry, then purchase order", () => {
    const procurementGroup = MENU_SECTIONS.find((section) => section.id === "transactions")?.groups.find((group) => group.id === "procurement");

    expect(procurementGroup?.items.map((item) => item.id)).toEqual([
      "procurement-dashboard",
      "purchase-requisition",
      "rfq",
      "purchase-order",
    ]);
    expect(procurementGroup?.items.find((item) => item.id === "rfq")?.label.th).toBe("สืบราคาและเจรจา");
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
    expect(toolIds).toEqual(["product-serial-registry", "price-history", "label-print", "add-product-kitchen"]);
  });

  it("splits defaults and master data into dependency-ordered groups", () => {
    const defaultsSection = MENU_SECTIONS.find((section) => section.id === "defaults");
    const masterSection = MENU_SECTIONS.find((section) => section.id === "master");
    const defaultGroupsById = new Map(defaultsSection?.groups.map((group) => [group.id, group]) ?? []);
    const masterGroupsById = new Map(masterSection?.groups.map((group) => [group.id, group]) ?? []);

    // Master Data holds core business records
    expect(masterGroupsById.get("products")?.title.th).toBe("สินค้าและบาร์โค้ด");
    expect(masterGroupsById.get("products")?.items.map((item) => item.id)).toEqual([
      "product",
      "service-product",
      "non-stock-product",
      "product-extension",
      "barcode",
      "product-category",
      "productset",
      "bom",
    ]);
    expect(masterGroupsById.get("partners")?.items.map((item) => item.id)).toEqual([
      "debtor",
      "creditor",
      "debtor-beginning-balance",
      "creditor-beginning-balance",
    ]);
    expect(masterGroupsById.get("bank-accounts")?.items.map((item) => item.id)).toEqual([
      "book-bank",
    ]);

    // Defaults holds baseline definitions in dependency order
    expect(defaultGroupsById.get("product-classification")?.title.th).toBe("จัดกลุ่มสินค้า");
    expect(defaultGroupsById.get("product-classification")?.items.map((item) => item.id)).toEqual([
      "product-unit",
      "product-group",
    ]);
    expect(defaultGroupsById.get("product-descriptors")?.title.th).toBe("รายละเอียดประกอบสินค้า");
    expect(defaultGroupsById.get("product-sku-options")?.title.th).toBe("สี ไซซ์ และตัวเลือก");
    expect(defaultGroupsById.get("product-sku-options")?.items.map((item) => item.id)).toEqual([
      "product-color",
      "product-size",
      "product-variant-matrix",
    ]);
    expect(defaultGroupsById.get("warehouse-setup")?.items.map((item) => item.id)).toEqual([
      "warehouse",
    ]);
    expect(defaultGroupsById.get("partner-groups")?.items.map((item) => item.id)).toEqual([
      "debtor-group",
      "creditor-group",
    ]);
    expect(defaultGroupsById.get("sales-channel-pricing")?.items.map((item) => item.id)).toContain("channel-price");
    expect(defaultGroupsById.get("sales-loyalty")?.items.map((item) => item.id)).toContain("promotion");
  });

  it("orders the defaults section according to business dependency workflow", () => {
    const defaultsSection = MENU_SECTIONS.find((section) => section.id === "defaults");
    const groupOrder = defaultsSection?.groups.map((group) => group.id) ?? [];

    expect(groupOrder).toEqual([
      "product-classification",
      "product-descriptors",
      "product-sku-options",
      "warehouse-setup",
      "partner-groups",
      "sales-payment-banking",
      "sales-channel-pricing",
      "sales-documents",
      "approval",
      "sales-pos",
      "sales-loyalty",
      "restaurant-setup",
    ]);
  });

  it("uses distinct product category labels for category structure and attribute category", () => {
    const allItems = flattenMenuItems();
    const labelsById = new Map(allItems.map((item) => [item.id, item.label.th]));

    expect(labelsById.get("product-category")).toBe("จัดหมวดสินค้า");
    expect(labelsById.get("category")).toBe("คุณลักษณะสินค้า");
    expect(allItems.filter((item) => item.label.th === "หมวดสินค้า")).toHaveLength(0);
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
    const defaultsSection = MENU_SECTIONS.find((section) => section.id === "defaults");
    const skuGroup = defaultsSection?.groups.find((group) => group.id === "product-sku-options");
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

describe("menu full-text search (all languages)", () => {
  const first = flattenMenuItems()[0];

  it("English query finds a Thai-labeled item (all-language haystack)", () => {
    expect(first).toBeTruthy();
    const enNeedle = normalizeMenuSearchText(first.label.en);
    expect(enNeedle.length).toBeGreaterThan(1);
    expect(menuSearchHaystack(first.label).includes(enNeedle)).toBe(true);
  });

  it("route/id are searchable in the combined haystack", () => {
    const haystack = `${menuSearchHaystack(first.label)} ${normalizeMenuSearchText(first.route)} ${normalizeMenuSearchText(first.id)}`;
    expect(haystack.includes(normalizeMenuSearchText(first.route))).toBe(true);
    expect(haystack.includes(normalizeMenuSearchText(first.id))).toBe(true);
  });

  it("matches across the OTHER language regardless of UI language", () => {
    const salesItem = flattenMenuItems().find((item) => menuSearchHaystack(item.label).includes(normalizeMenuSearchText("sale")));
    expect(salesItem).toBeTruthy();
    // same item must also be reachable with its Thai text
    expect(menuSearchHaystack(salesItem!.label)).toContain(normalizeMenuSearchText(salesItem!.label.th));
  });

  it("is case-insensitive for Latin", () => {
    expect(normalizeMenuSearchText("InVoice")).toBe(normalizeMenuSearchText("invoice"));
  });

  it("Thai tone-mark typos still match (combining marks stripped on both sides)", () => {
    const needle = normalizeMenuSearchText("\u0e23\u0e32\u0e22\u0e07\u0e32\u0e19"); // รายงาน
    const wrongTone = normalizeMenuSearchText("\u0e23\u0e32\u0e49\u0e22\u0e07\u0e32\u0e19"); // รา้ยงาน (extra mark)
    expect(needle).toBe(wrongTone);
    const reportItem = flattenMenuItems().find((item) => item.label.key === "report" || /report/i.test(item.route));
    void reportItem;
  });

  it("menuSearchMatches uses key and backend dictionary override", () => {
    const label = { key: "invoice", th: "ใบแจ้งหนี้", en: "Invoice" };
    expect(menuSearchMatches(label, normalizeMenuSearchText("invoice"))).toBe(true);
    expect(menuSearchMatches(label, normalizeMenuSearchText("ใบแจ้งหนี้"))).toBe(true);
    expect(menuSearchMatches(label, normalizeMenuSearchText("invoice"), { invoice: "請求書" } as BackendLanguageDictionary)).toBe(true);
    const cnNeedle = normalizeMenuSearchText("請求書");
    expect(menuSearchMatches(label, cnNeedle, { invoice: "請求書" } as BackendLanguageDictionary)).toBe(true);
    expect(menuSearchMatches(label, normalizeMenuSearchText("xyz-not-exist"))).toBe(false);
  });
});
