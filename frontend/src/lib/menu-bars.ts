import { flattenMenuItems, type MenuItem, type MenuLabel } from "./menu-data";

/**
 * Champ-style menu bar layout (2026-10-07, ลุงจืด: "เมนูหลักคล้าย Champ ที่สุด").
 *
 * `MENU_SECTIONS` stays the permission catalog (one `id` per screen, grouped by module);
 * this file only decides how those same items are laid out in the top bar and the left
 * tree: 13 heads in Champ order, divider-separated blocks, nested submenus. Every catalog
 * item must appear exactly once (checked by menu-bars.test.ts) — never add a second copy.
 */
export type MenuBarActionId =
  | "change-company"
  | "change-password"
  | "exit-program"
  | "create-shortcut"
  | "close-all-tabs"
  | "open-manual";

export type MenuBarNode =
  | { kind: "item"; item: MenuItem }
  // Opens an existing system-settings screen (src/lib/system-setting-screens.ts) in a work tab. `item` is a synthetic
  // MenuItem that is NOT in MENU_SECTIONS: no permission id here; real enforcement is the backend (e.g. OWNER/ADMIN on /holding/permission), the menu does not lock these entries.
  | { kind: "route"; item: MenuItem }
  | { kind: "divider" }
  | { kind: "submenu"; id: string; label: MenuLabel; nodes: MenuBarNode[] }
  | { kind: "action"; id: MenuBarActionId; label: MenuLabel };

export type MenuBar = { id: string; label: MenuLabel; nodes: MenuBarNode[] };

const catalog = new Map(flattenMenuItems().map((item) => [item.id, item]));

const label = (key: string, th: string, en: string): MenuLabel => ({ key, th, en });

const DIVIDER: MenuBarNode = { kind: "divider" };

function item(id: string): MenuBarNode {
  const found = catalog.get(id);
  if (!found) throw new Error(`menu-bars: unknown menu item "${id}"`);
  return { kind: "item", item: found };
}

const items = (...ids: string[]): MenuBarNode[] => ids.map(item);

const routeItemIds = new Set<string>();

function routeNode(id: string, path: string, key: string, th: string, en: string): MenuBarNode {
  routeItemIds.add(id);
  return { kind: "route", item: { id, label: label(key, th, en), route: path, category: "settings" } };
}

/** True for the synthetic items of `route` nodes (they are not catalog items, so no permission id or usage stats). */
export const isMenuBarRouteItem = (menuItem: MenuItem): boolean => routeItemIds.has(menuItem.id);

const action = (id: MenuBarActionId, key: string, th: string, en: string): MenuBarNode => ({
  kind: "action",
  id,
  label: label(key, th, en),
});

/** Drops leading/trailing/double dividers so an empty block never leaves a stray line. */
function tidy(nodes: MenuBarNode[]): MenuBarNode[] {
  const result: MenuBarNode[] = [];
  for (const node of nodes) {
    if (node.kind === "divider" && (!result.length || result[result.length - 1].kind === "divider")) continue;
    result.push(node);
  }
  while (result.length && result[result.length - 1].kind === "divider") result.pop();
  return result;
}

const submenu = (id: string, key: string, th: string, en: string, ...nodes: MenuBarNode[]): MenuBarNode[] => {
  const children = tidy(nodes);
  return children.length ? [{ kind: "submenu", id, label: label(key, th, en), nodes: children }] : [];
};

const bar = (id: string, head: MenuLabel, ...groups: (MenuBarNode | MenuBarNode[])[]): MenuBar => ({
  id,
  label: head,
  nodes: tidy(groups.flat()),
});

const REPORTS_DAILY = ["menu_sub_reports_daily", "รายงานรายวัน", "Daily Reports"] as const;
const REPORTS_ANALYSIS = ["menu_sub_reports_analysis", "รายงานเชิงวิเคราะห์", "Analytical Reports"] as const;
const RECALCULATE = ["menu_sub_recalculate_balance", "คำนวณยอดคงเหลือใหม่", "Recalculate Balances"] as const;

// Order, dividers and submenu names follow the real Champ menus (scratch/champ-atlas/menu-diff.md §1.2).
// Items Champ does not have (sale-reservation-flow, etax-invoice, barcode …, gl-year-end, tax-forms …)
// sit at the end of the nearest Champ block so nothing in the permission catalog is lost.
export const MENU_BARS: MenuBar[] = [
  bar(
    "file",
    label("menu_bar_file", "ไฟล์", "File"),
    action("change-company", "menu_action_change_company", "เลือกบริษัททำงานใหม่", "Select Company"),
    routeNode("bar-users", "/user", "menu_route_user_accounts", "กำหนดรหัสผู้ใช้งาน", "User Accounts"),
    action("change-password", "menu_action_change_password", "เปลี่ยนรหัสผ่าน", "Change Password"),
    routeNode("bar-permissions", "/permissiongroup", "menu_route_permissions", "กำหนดสิทธิ์การใช้งาน", "User Permissions"),
    DIVIDER,
    action("exit-program", "menu_action_exit_program", "ออกจากโปรแกรม", "Exit"),
  ),
  bar(
    "po",
    label("menu_bar_purchase", "จัดซื้อ", "Purchasing"),
    items("purchase-requisition", "purchase-requisition-cancel", "purchase-requisition-approve", "purchase-requisition-approve-cancel"),
    submenu(
      "po-price-inquiry", "menu_sub_price_inquiry", "ระบบสืบราคาซื้อสินค้า", "Price Inquiry System",
      ...items("rfq", "rfq-price-table", "purchase-order-generate"),
    ),
    DIVIDER,
    items("purchase-order", "purchase-order-approve", "purchase-order-cancel"),
    DIVIDER,
    items("deposit", "advance-payment", "advance-payment-refund"),
    DIVIDER,
    submenu(
      "po-credit-purchase", "menu_sub_credit_purchase", "บันทึกซื้อสินค้าเชื่อ", "Credit Purchases",
      ...items("purchase-partial", "accrual-receive", "purchase-return", "purchase-reduce-debt"),
    ),
    items("purchase", "purchase-debit-note", "purchase-credit-note"),
    DIVIDER,
    items("purchase-landed-cost", "purchase-receipt-adjustment", "purchase-receipt-date-adjust"),
    DIVIDER,
    submenu("po-reports-daily", ...REPORTS_DAILY, ...items("purchase-report", "gradual-receipt")),
    submenu("po-reports-analysis", ...REPORTS_ANALYSIS, ...items("purchase-by-product", "purchase-price-comparison")),
  ),
  bar(
    "bill",
    label("menu_bar_sales", "จัดขาย", "Sales"),
    items("quotation", "quotation-approve", "quotation-cancel"),
    DIVIDER,
    items(
      "sale-reservation", "sale-order", "sale-order-approve", "sale-order-cancel", "sale-dispatch", "receive-deposit",
      "sale-reservation-flow", "sale-order-date-check",
    ),
    DIVIDER,
    items("receive-advance", "return-advance"),
    DIVIDER,
    items("sale", "sale-return", "debit-note", "etax-invoice"),
    DIVIDER,
    items("product-sale-price", "sale-delivery-date", "product-promotion"),
    DIVIDER,
    submenu(
      "bill-reports-daily", "menu_sub_reports_daily_data", "รายงานข้อมูลรายวัน", "Daily Data Reports",
      ...items("sales", "sales-daily", "sales-by-document", "sale-return-report"),
    ),
    submenu(
      "bill-reports-analysis", ...REPORTS_ANALYSIS,
      ...items("sales-by-seller", "gross-profit-by-document", "gross-profit-by-product"),
    ),
  ),
  bar(
    "ap",
    label("menu_account_payable_ap", "เจ้าหนี้", "Account Payable (AP)"),
    items("creditor", "creditor-beginning-balance", "ap-other-debt", "ap-billing-receipt"),
    DIVIDER,
    items("ap-payment-prepare", "pay", "ap-bad-debt"),
    DIVIDER,
    submenu("ap-balance-tools", ...RECALCULATE, ...items("ap-recalculate", "ap-bill-recalculate")),
    submenu("ap-reports-daily", ...REPORTS_DAILY, ...items("ap-movement", "ap-outstanding", "ap-daily-payment")),
    submenu("ap-reports-analysis", ...REPORTS_ANALYSIS, ...items("ap-aging", "ap-status")),
  ),
  bar(
    "ar",
    label("menu_account_receivable_ar", "ลูกหนี้", "Account Receivable (AR)"),
    items("debtor", "debtor-beginning-balance", "ar-other-debt", "billing-note"),
    DIVIDER,
    items("billing-note-auto", "billing-note-auto-print"),
    DIVIDER,
    items("temporary-receipt", "paid", "ar-bad-debt"),
    DIVIDER,
    items("customer-label-print", "ar-credit-limit"),
    DIVIDER,
    submenu("ar-balance-tools", ...RECALCULATE, ...items("ar-recalculate", "ar-bill-recalculate")),
    submenu("ar-reports-daily", ...REPORTS_DAILY, ...items("payment-daily", "ar-movement", "ar-outstanding")),
    submenu("ar-reports-analysis", ...REPORTS_ANALYSIS, ...items("ar-aging", "ar-status")),
  ),
  bar(
    "cheque",
    label("menu_bar_cheque", "เช็ค/บัตร", "Cheques & Cards"),
    submenu(
      "cheque-registers", "menu_sub_cheque_registers", "ทะเบียนหลักต่างๆ", "Master Registers",
      ...items("cheque-received", "cheque-issued", "credit-card-receipts"),
      DIVIDER,
      ...items("cheque-issued-prepare"),
    ),
    items("cheque-movement"),
    DIVIDER,
    items("cheque-deposit", "cheque-received-clear", "cheque-received-return", "cheque-redeposit", "cheque-received-cancel", "cheque-discount"),
    DIVIDER,
    items("cheque-issued-clear", "cheque-issued-cancel"),
    DIVIDER,
    items("credit-card-settle", "credit-card-cancel"),
    DIVIDER,
    submenu(
      "cheque-reports-master", "menu_sub_reports_master_data", "รายงานข้อมูลหลัก", "Master Data Reports",
      ...items("cheque-received-report", "cheque-issued-report", "credit-card-report"),
    ),
    items("cheque-recalculate"),
  ),
  bar(
    "bank-cash",
    label("menu_bar_bank_cash", "ธนาคาร/เงินสด", "Bank & Cash"),
    items("other-income-receipt", "employee-advance", "expense-record"),
    DIVIDER,
    items("bank-service", "bank-contact", "book-bank", "petty-cash-limit", "cash-deposit", "cash-withdraw", "bank-charges", "bank-income", "account-transfer"),
    DIVIDER,
    items("bank-recalculate"),
    submenu(
      "bank-reports", "menu_sub_reports_misc", "รายงานต่างๆ", "Reports",
      ...items("bank-statement-report", "cash-movement-report", "petty-cash-movement-report", "expense-summary-report", "monthly-payment-book-report"),
    ),
  ),
  bar(
    "ic",
    label("menu_inventory_control_ic", "สินค้าคงคลัง", "Inventory Control (IC)"),
    items("product"),
    submenu(
      "ic-product-sets", "menu_sub_product_sets", "สินค้าชุด", "Product Sets",
      ...items("product-set", "product-set-assemble", "product-set-disassemble"),
      DIVIDER,
      ...items("product-set-components"),
    ),
    items("product-serial-registry"),
    DIVIDER,
    items("product-price-adjust", "product-supplier-update", "goods-inspection"),
    DIVIDER,
    items("stock-balance", "stock-receive"),
    DIVIDER,
    items("stock-issue-request", "stock-issue-request-cancel", "stock-pickup", "stock-return"),
    DIVIDER,
    items("stock-transfer-request", "stock-transfer-request-cancel", "stock-transfer"),
    submenu(
      "ic-counting", "menu_sub_stock_counting", "การตรวจนับสินค้า", "Stock Counting",
      ...items("stock-count-sheet", "stock-count", "stock-adjust"),
      DIVIDER,
      ...items("stock-count-handheld"),
    ),
    DIVIDER,
    items("reprocess", "stock-daily-sequence", "transaction-lock"),
    DIVIDER,
    submenu(
      "ic-reports-product", "menu_sub_reports_product", "รายงานสินค้า", "Product Reports",
      ...items("stock-balance-item", "stock-balance-warehouse", "inventory", "low-stock-alert", "max-stock-report"),
    ),
    submenu(
      "ic-reports-daily", ...REPORTS_DAILY,
      ...items("product-movement", "stock-movement-cost", "pending-receive-report", "pending-delivery-report", "serial-movement-report"),
    ),
    submenu(
      "ic-reports-analysis", "menu_sub_reports_product_analysis", "รายงานเชิงวิเคราะห์สินค้า", "Product Analysis Reports",
      ...items("no-movement-stock-report", "stock-count-variance-report"),
    ),
    items("product-status-report", "product-price-setting-report"),
  ),
  bar(
    "fa",
    label("menu_bar_fixed_assets", "สินทรัพย์", "Fixed Assets"),
    items("asset-registry", "asset-maintenance", "asset-construction-in-progress", "asset-type"),
    DIVIDER,
    items("asset-depreciation", "asset-post-gl"),
    DIVIDER,
    submenu("fa-reports-asset", "menu_sub_reports_asset", "รายงานสินทรัพย์", "Asset Reports", ...items("fixed-asset-schedule")),
    submenu(
      "fa-reports-yearly", "menu_sub_reports_asset_yearly", "รายงานคิดค่าเสื่อมราคาเป็นปี", "Yearly Depreciation Reports",
      ...items("depreciation-yearly-report", "depreciation-pnd50-report"),
    ),
    submenu(
      "fa-reports-monthly", "menu_sub_reports_asset_monthly", "รายงานคิดค่าเสื่อมราคาเป็นเดือน", "Monthly Depreciation Reports",
      ...items("depreciation-monthly-report"),
    ),
    submenu("fa-reports-other", "menu_sub_reports_other", "รายงานอื่นๆ", "Other Reports", ...items("asset-disposal-report")),
  ),
  bar(
    "gl",
    label("menu_bar_accounting", "บัญชี", "Accounting"),
    items("chart-of-accounts", "gl-journals"),
    DIVIDER,
    items("gl-reprocess", "gl-post", "gl-unpost", "gl-recalculate-posted"),
    DIVIDER,
    items("gl-account-mapping", "period-lock", "gl-journal-books", "gl-budget", "gl-account-groups", "gl-product-account-groups"),
    DIVIDER,
    items("gl-opening-balance", "financial-close", "purchase-vat-adjust", "withholding-tax-deduction", "gl-fiscal-years", "gl-year-end"),
    DIVIDER,
    // Champ order (docs/kms/champ/report-parity-r0-gl.md §4): master-data reports nested first, then the daily reports, budget comparison last
    submenu(
      "gl-reports-daily", ...REPORTS_DAILY,
      ...submenu(
        "gl-reports-master", "gl_reports_master", "รายงานข้อมูลหลักต่างๆ", "Master Data Reports",
        ...items("gl-journal-books-report", "gl-department-report", "gl-project-report", "gl-allocate-report", "gl-job-report", "gl-part-report", "gl-side-report", "gl-budget-detail", "gl-chart-of-accounts-report"),
      ),
      DIVIDER,
      ...items("gl-daily-report", "gl-daily-summary-report", "general-ledger", "general-ledger-summary"),
      DIVIDER,
      ...items("budget-comparison-report"),
    ),
    submenu(
      "gl-reports-financial", "menu_sub_reports_financial", "รายงานงบการเงิน", "Financial Statement Reports",
      ...items(
        "trial-balance", "trial-balance-2", "working-paper", "financial-statement-designer", "profit-loss", "balance-sheet",
        "gl-ar-outstanding", "gl-ap-outstanding", "gl-bank-unmatched",
      ),
    ),
    submenu(
      "gl-reports-tax", "menu_sub_reports_tax", "รายงานภาษี", "Tax Reports",
      ...items(
        "vat-buy", "vat-sale", "vat-summary-report", "vat-pp30", "vat-pp36", "withholding-tax-report",
        "withholding-tax-received", "vat-pnd2", "vat-pnd3", "vat-pnd53", "tax-forms", "wht-certificate",
      ),
    ),
  ),
  // จัดการระบบ follows the real Champ dropdown (bcspec หน้าจอเมนูจัดการระบบ/เมนู-ประกอบ.jpg). Omitted on purpose:
  // ขอดูข้อมูล (raw-SQL viewer, security) and แสดง ToolTip — awaiting the owner's decision.
  // Champ opens กำหนดรหัสต่างๆ as one screen with a code-type combo; here it is a submenu of the existing code screens.
  bar(
    "system",
    label("menu_bar_system", "จัดการระบบ", "System"),
    routeNode("bar-company-defaults", "/branch", "menu_route_company_defaults", "กำหนดค่าเริ่มต้นบริษัท", "Company Defaults"),
    items("system-journal-format", "system-work-date"),
    DIVIDER,
    submenu(
      "system-codes", "menu_sub_system_codes", "กำหนดรหัสต่างๆ", "Master Codes",
      ...items("gl-dimensions", "product-group", "brand", "warehouse", "product-unit", "barcode",
        "bank-code-kind-bank", "bank-code-kind-bank-branch", "bank-code-kind-bank-service", "bank-code-kind-bank-account-type", "bank-code-kind-cheque-ext-status"),
    ),
    routeNode("bar-warehouse-locations", "/productwarehousescreen", "menu_route_warehouse_locations", "กำหนดที่เก็บสินค้าตามคลัง", "Storage Locations by Warehouse"),
    items("system-product-by-warehouse", "system-default-warehouse", "system-currency", "system-main-map"),
    routeNode("bar-employees", "/employee", "menu_route_employees_sellers", "กำหนดพนักงาน/พนักงานขาย", "Employees and Salespeople"),
    items("system-credit-debit-reasons"),
    routeNode("bar-debtor-groups", "/debtorgroup", "menu_route_debtor_groups", "กำหนดกลุ่มลูกหนี้", "Debtor Groups"),
    items("system-payment-terms"),
    DIVIDER,
    action("create-shortcut", "menu_action_create_shortcut", "สร้าง Shortcut", "Create Shortcut"),
    items("system-customer-inquiry", "system-product-inquiry"),
  ),
  bar(
    "window",
    label("menu_bar_window", "หน้าต่าง", "Window"),
    action("close-all-tabs", "menu_action_close_all_tabs", "ปิดหน้าจอทั้งหมด", "Close All Screens"),
    submenu(
      "window-fonts", "menu_sub_window_fonts", "เลือกตัวอักษรในการแสดงผลต่าง", "Choose Display Fonts",
      ...items(
        "window-font-screen",
        "window-font-report",
        "window-font-report-title",
        "window-font-report-column",
        "window-font-report-detail",
        "window-font-report-company",
      ),
    ),
  ),
  bar(
    "help",
    label("menu_bar_help", "ช่วยเหลือ", "Help"),
    items("help-about", "help-standard-software"),
    DIVIDER,
    action("open-manual", "menu_action_open_manual", "คู่มือการใช้งาน", "User Manual"),
  ),
];

/** Every catalog item shown in a bar (submenus included), in display order — the permission catalog view, route nodes excluded. */
export function menuBarItems(nodes: MenuBarNode[]): MenuItem[] {
  return nodes.flatMap((node) => {
    if (node.kind === "item") return [node.item];
    if (node.kind === "submenu") return menuBarItems(node.nodes);
    return [];
  });
}

/** Everything a bar row can open — catalog items plus settings routes — in display order (menu UI only). */
export function menuBarEntries(nodes: MenuBarNode[]): MenuItem[] {
  return nodes.flatMap((node) => {
    if (node.kind === "item" || node.kind === "route") return [node.item];
    if (node.kind === "submenu") return menuBarEntries(node.nodes);
    return [];
  });
}
