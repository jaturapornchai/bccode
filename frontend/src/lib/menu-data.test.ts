import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { MENU_SECTIONS, flattenMenuItems, menuSearchHaystack, menuSearchMatches, menuText, normalizeMenuSearchText } from "./menu-data";
import { getSystemSettingConfig } from "./system-setting-screens";
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
  it("preserves all pre-upgrade permission IDs and routes without duplicate destinations", () => {
    const baseline = JSON.parse(readFileSync(resolve(process.cwd(), "src/lib/__fixtures__/menu-before-champ-upgrade.json"), "utf8")) as { id: string; route: string }[];
    const items = flattenMenuItems();
    expect(baseline).toHaveLength(194);
    expect(items).toHaveLength(194);
    for (const previous of baseline) {
      expect(items.find((item) => item.id === previous.id)?.route, previous.id).toBe(previous.route);
    }
    expect(new Set(items.map((item) => item.route)).size).toBe(items.length);
  });

  it("keeps Champ cheque workflows separate from corporate card expenses", () => {
    const cash = MENU_SECTIONS.find((section) => section.id === "cash-bank")!;
    expect(cash.groups.find((group) => group.id === "cheques-received")?.items.map((item) => item.id)).toEqual([
      "cheque-received", "cheque-deposit", "cheque-received-clear", "cheque-received-return",
      "cheque-redeposit", "cheque-received-cancel", "cheque-discount",
    ]);
    expect(cash.groups.find((group) => group.id === "cheques-issued")?.items.map((item) => item.id)).toEqual([
      "cheque-issued", "cheque-issued-clear", "cheque-issued-cancel",
    ]);
    expect(cash.groups.find((group) => group.id === "card-settlement")?.items.some((item) => item.id === "credit-card-receipts")).toBe(true);
  });

  it("excludes payroll while retaining partner withholding taxes and employee advances", () => {
    const items = flattenMenuItems();
    const forbidden = /payroll|salary|pnd1(?:\D|$)|เงินเดือน|ภ\.?\s*ง\.?\s*ด\.?\s*1(?:\D|$)|ประกันสังคม|กท\.?\s*20/i;
    expect(items.filter((item) => forbidden.test(`${item.id} ${item.route} ${item.label.th}`))).toEqual([]);
    for (const id of ["vat-pnd2", "vat-pnd3", "vat-pnd53", "wht-certificate", "employee-advance"]) {
      expect(items.some((item) => item.id === id), id).toBe(true);
    }
  });
  it("keeps every Thai menu label unchanged after the backend dictionary loads", () => {
    const rows = backendLanguageRows();
    const dictionary = readyDictionary(Object.fromEntries([...rows].map(([key, row]) => [key, row.th])));
    for (const item of flattenMenuItems()) {
      expect(rows.get(item.label.key ?? "")?.th, item.id).toBe(item.label.th);
      expect(menuText(item.label, "th", dictionary), item.id).toBe(item.label.th);
    }
  });

  it("does not let a stale Thai dictionary rename a business action", () => {
    const item = flattenMenuItems().find((item) => item.id === "sale-order")!;
    expect(menuText(item.label, "th", readyDictionary({ sale_order: "ใบเสนอราคา/ใบแจ้งหนี้" }))).toBe("บันทึกใบสั่งขาย/สั่งจองสินค้า");
    expect(menuText(item.label, "en", readyDictionary({ sale_order: "Sales Order" }))).toBe("Sales Order");
  });

  it("provides the pending status in every supported dictionary language", () => {
    const row = backendLanguageRows().get("menu_pending_development");
    for (const language of ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi", "ms", "id", "fil"]) {
      expect(row?.[language], language).toBeTruthy();
    }
    expect(row?.th).toBe("รอพัฒนา");
  });

  it("uses the same Thai title in menus and connected setting screens", () => {
    for (const item of flattenMenuItems()) {
      const config = getSystemSettingConfig(item.route);
      if (config) expect(config.title.th, item.id).toBe(item.label.th);
    }
  });

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

  it("has no organization-people, dimensions, product-tools, or product-assistant groups in any menu section (2026-09-10)", () => {
    const allGroupIds = MENU_SECTIONS.flatMap((section) => section.groups.map((group) => group.id));
    expect(allGroupIds).not.toContain("organization-people");
    expect(allGroupIds).not.toContain("dimensions");
    expect(allGroupIds).not.toContain("product-tools");
    expect(allGroupIds).not.toContain("product-assistant");

    const allRoutes = flattenMenuItems().map((item) => item.route);
    expect(allRoutes).not.toContain("/employee");
    expect(allRoutes).not.toContain("/user");
    expect(allRoutes).not.toContain("/permissiongroup");
    expect(allRoutes).not.toContain("/useraccessaudit");
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

  it("orders purchasing like Champ: transactions first, then requisition and price inquiry (2026-09-19)", () => {
    const poSection = MENU_SECTIONS.find((section) => section.id === "po");
    const transactions = poSection?.groups.find((group) => group.id === "po-transactions");
    const procurement = poSection?.groups.find((group) => group.id === "po-procurement");

    expect(transactions?.items.map((item) => item.id).slice(0, 4)).toEqual([
      "purchase-order",
      "purchase-order-cancel",
      "purchase-partial",
      "accrual-receive",
    ]);
    expect(procurement?.items.map((item) => item.id)).toEqual([
      "purchase-requisition",
      "purchase-requisition-approve",
      "rfq",
      "rfq-price-table",
      "purchase-order-generate",
    ]);
    expect(procurement?.items.find((item) => item.id === "rfq")?.label.th).toBe("บันทึกใบสืบราคาสินค้ารวม");
  });

  it("drops the non-Champ items and carries every Champ item after the 2026-09-19 parity cut", () => {
    const ids = new Set(flattenMenuItems().map((item) => item.id));
    const removed = [
      "procurement-dashboard", "import-documents", "recurring-expense", "document-vault", "inter-company-inbox", "deposit-refund",
      "recurring-invoice", "tax-invoice", "combined-receipt", "credit-note", "return-deposit", "sales-by-customer", "sales-by-channel",
      "creditor-group", "import-partner", "payment-voucher", "combined-payment", "debtor-group", "sale-invoice", "petty-cash",
      "director-advance", "credit-card-expense", "cash-drawer", "bank-payment-file", "slip-in", "slip-out", "bank-statement", "bank-reconcile",
      "import-product", "import-product-file", "import-product-image", "fifo-cost-layers", "stock-lot", "cost-adjustment",
      "stock-balance-location", "expiring-stock-alert", "stock-lot-movement", "audit-data", "rebuild-products", "rebuild-product-balance",
      "asset-purchase", "asset-disposal", "purchase-tax-invoice-register", "unreceived-tax-invoice", "deferred-tax", "cash-flow", "project-pnl",
    ];
    const added = [
      "purchase-reduce-debt", "rfq-price-table", "ap-movement", "ap-status", "ap-outstanding", "ap-daily-payment", "ar-movement", "ar-status",
      "ar-outstanding", "ar-credit-limit", "cheque-received-report", "cheque-issued-report", "credit-card-report", "bank-statement-report",
      "cash-movement-report", "petty-cash-movement-report", "monthly-payment-book-report", "max-stock-report", "no-movement-stock-report",
      "stock-count-variance-report", "pending-receive-report", "pending-delivery-report", "serial-movement-report", "depreciation-monthly-report",
      "depreciation-yearly-report", "depreciation-pnd50-report", "asset-disposal-report", "vat-summary-report", "gl-account-groups",
      "gl-journal-books", "gl-reprocess", "gl-daily-report", "budget-comparison-report",
    ];
    expect(removed).toHaveLength(47);
    expect(added).toHaveLength(33);
    expect(removed.filter((id) => ids.has(id))).toEqual([]);
    expect(added.filter((id) => !ids.has(id))).toEqual([]);
    expect(flattenMenuItems().find((item) => item.id === "gl-opening-balance")?.label.th).toBe("บันทึกยอดสะสมประจำปี");
  });

  it("keeps core product group clean without auxiliary tool items", () => {
    const icSection = MENU_SECTIONS.find((section) => section.id === "ic");
    const productGroup = icSection?.groups.find((group) => group.id === "ic-master");
    const productIds = productGroup?.items.map((item) => item.id) ?? [];

    expect(productIds).not.toContain("price-history");
    expect(productIds).not.toContain("label-print");
    expect(productIds).not.toContain("add-product-branch");
    expect(productIds).not.toContain("add-product-department");
    expect(productIds).not.toContain("add-product-kitchen");
  });

  it("structures menus into the 8 Champ modules with tax inside General Ledger (2026-09-23)", () => {
    const sectionIds = MENU_SECTIONS.map((s) => s.id);
    expect(sectionIds).toEqual([
      "po",
      "bill",
      "ap",
      "ar",
      "cash-bank",
      "ic",
      "fa",
      "gl",
    ]);

    const glSection = MENU_SECTIONS.find((s) => s.id === "gl");
    expect(glSection?.title.th).toBe("บัญชีแยกประเภท");
    // ผู้ใช้ GL ตัวเดียวต้องเข้าถึงรายงานภาษี/แบบยื่น/50 ทวิ ได้ในหมวดเดียว (Champ วางเมนูภาษีไว้ใต้ GL)
    expect(glSection?.groups.map((g) => g.id)).toEqual(["gl-master", "gl-journals", "gl-posting", "gl-reports", "vat-transactions", "vat-reports"]);

    const icSection = MENU_SECTIONS.find((s) => s.id === "ic");
    expect(icSection?.title.th).toBe("สินค้าคงคลัง");

    const poSection = MENU_SECTIONS.find((s) => s.id === "po");
    expect(poSection?.title.th).toBe("ซื้อ/สั่งซื้อสินค้า");

    const apSection = MENU_SECTIONS.find((s) => s.id === "ap");
    expect(apSection?.title.th).toBe("เจ้าหนี้");

    const arSection = MENU_SECTIONS.find((s) => s.id === "ar");
    expect(arSection?.title.th).toBe("ลูกหนี้");

    const billSection = MENU_SECTIONS.find((s) => s.id === "bill");
    expect(billSection?.title.th).toBe("ใบสั่งของ/ใบกำกับสินค้า");

    const cashBankSection = MENU_SECTIONS.find((s) => s.id === "cash-bank");
    expect(cashBankSection?.title.th).toBe("เงินสดและธนาคาร");

    expect(MENU_SECTIONS.find((s) => s.id === "vat")).toBeUndefined();

    const faSection = MENU_SECTIONS.find((s) => s.id === "fa");
    expect(faSection?.title.th).toBe("สินทรัพย์และค่าเสื่อมราคา");
  });

  it("uses plain business label for category structure", () => {
    const allItems = flattenMenuItems();
    const labelsById = new Map(allItems.map((item) => [item.id, item.label.th]));

    expect(labelsById.get("product-group")).toBe("กลุ่มสินค้า");
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

  it.each([
    ["purchase-requisition", "บันทึกใบเสนอซื้อสินค้า"],
    ["sale-order", "ใบสั่งจอง"],
    ["purchase-landed-cost", "Weight cost"],
    ["product-serial-registry", "Serial Number"],
    ["gl-account-mapping", "Account Mapping"],
  ])("finds %s by familiar Champ wording %s", (id, query) => {
    const item = flattenMenuItems().find((item) => item.id === id)!;
    expect(menuSearchMatches(item.label, query)).toBe(true);
  });

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
