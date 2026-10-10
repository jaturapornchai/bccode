import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { ChampMenuBar } from "@/app/menu/champ-menu-bar";
import type { BackendLanguageDictionary } from "./backend-language";
import { GL_MENU_ITEMS } from "./general-ledger";
import { MENU_BARS, isMenuBarRouteItem, menuBarEntries, menuBarItems, type MenuBarNode } from "./menu-bars";
import type { MenuLabel } from "./menu-data";
import { flattenMenuItems } from "./menu-data";
import { isMenuScreenPending } from "./menu-screen-status";
import { getSystemSettingConfig } from "./system-setting-screens";

/** Heads of the real Champ 5.5 menu bar, left to right (bcspec screenshot "เมนูโปรแกรมทั้งหมด"). */
const CHAMP_HEADS = [
  "ไฟล์", "จัดซื้อ", "จัดขาย", "เจ้าหนี้", "ลูกหนี้", "เช็ค/บัตร", "ธนาคาร/เงินสด",
  "สินค้าคงคลัง", "สินทรัพย์", "บัญชี", "จัดการระบบ", "หน้าต่าง", "ช่วยเหลือ",
];

/** Items the real Champ menu has but BC has no screen for yet — added as "รอพัฒนา" (no real route). */
// 2026-10-10: All 13 Champ Menu Bar heads are now fully connected!
const CHAMP_ONLY_PENDING_IDS: string[] = [];

const CHAMP_LABELS: Record<string, string> = {
  "purchase-debit-note": "ใบส่งคืนสินค้า/ลดหนี้",
  "purchase-credit-note": "บันทึกใบเพิ่มหนี้/เพิ่มสินค้า(เจ้าหนี้)",
  "sale-order": "ใบสั่งขายสินค้า",
  "sale-order-cancel": "ยกเลิกใบสั่งขายสินค้า",
  "debtor-beginning-balance": "บันทึกเอกสารลูกหนี้ยกมาต้นปี",
  "ar-bad-debt": "ตัดหนี้สูญ/ปรับปรุง (ลูกหนี้)",
  product: "บันทึกรายละเอียดสินค้า",
  reprocess: "คำนวณยอดคงเหลือรวมของสินค้าใหม่",
  "asset-registry": "รายการสินทรัพย์",
  "chart-of-accounts": "รายการผังบัญชี",
  "gl-journals": "บันทึกข้อมูลรายวัน",
  "gl-post": "ผ่านรายการ",
  "withholding-tax-deduction": "พิมพ์ภาษีหัก ณ ที่จ่าย (ภ.ง.ด. 3,ภ.ง.ด. 53)",
  "product-set-assemble": "บันทึกการจัดชุดสินค้า",
  "product-set-disassemble": "บันทึกการแยกสินค้าออกจากชุด",
  "product-promotion": "กำหนดโปรโมชั่น",
  "ar-credit-limit": "ตรวจสอบวงเงินเครดิต",
  // 2026-10-09 (S1): exact wording of the Champ items added to the transaction heads
  "purchase-requisition-cancel": "ยกเลิกใบเสนอซื้อสินค้า",
  "purchase-requisition-approve-cancel": "ยกเลิกใบอนุมัติเสนอซื้อสินค้า",
  "purchase-order-approve": "อนุมัติใบสั่งซื้อสินค้า",
  "purchase-receipt-date-adjust": "ปรับปรุงวันที่รับสินค้าจากเจ้าหนี้",
  "sale-reservation": "ใบสั่งจองสินค้า",
  "sale-dispatch": "ใบสั่งจ่ายสินค้า",
  "ap-payment-prepare": "บันทึกเตรียมจ่ายชำระหนี้",
  "cheque-issued-prepare": "บันทึกเช็คเตรียมจ่าย",
  "cheque-movement": "ข้อมูลการเคลื่อนไหวเช็ค",
  "bank-service": "บริการธนาคาร",
  "product-supplier-update": "ปรับปรุง ข้อมูล Supplier ตามสินค้า",
  "stock-issue-request-cancel": "ยกเลิกใบขอเบิกใช้สินค้า,วัตถุดิบ",
  "stock-transfer-request-cancel": "ยกเลิกใบขอโอนสินค้า",
  "stock-count-handheld": "ตรวจนับสต็อกโดย Hand Held",
  "transaction-lock": "Transaction Lock",
  "gl-unpost": "ยกเลิกเอกสารที่ผ่านรายการแล้ว",
};

const tsvPath = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");

function languageRows(): { lines: string[][]; rows: Map<string, string[]> } {
  const lines = readFileSync(tsvPath, "utf8").split(/\r?\n/).filter((line) => line.length > 0).map((line) => line.split("\t"));
  return { lines, rows: new Map(lines.map((columns) => [columns[0] ?? "", columns])) };
}

function collectLabels(nodes: MenuBarNode[]): MenuLabel[] {
  return nodes.flatMap((node) => {
    if (node.kind === "item" || node.kind === "route") return [node.item.label];
    if (node.kind === "submenu") return [node.label, ...collectLabels(node.nodes)];
    if (node.kind === "action") return [node.label];
    return [];
  });
}

/** Every node of a bar with submenus flattened (the submenu node itself is kept before its children). */
function allNodes(nodes: MenuBarNode[]): MenuBarNode[] {
  return nodes.flatMap((node) => (node.kind === "submenu" ? [node, ...allNodes(node.nodes)] : [node]));
}

/** One line per top-level node: Thai label, "---" for a divider, "▸" before a submenu. */
function topLevelOutline(barId: string): string[] {
  return (MENU_BARS.find((bar) => bar.id === barId)?.nodes ?? []).map((node) => {
    if (node.kind === "divider") return "---";
    if (node.kind === "submenu") return `▸ ${node.label.th}`;
    if (node.kind === "action") return node.label.th;
    return node.item.label.th;
  });
}

function strayLayout(nodes: MenuBarNode[], where: string): string[] {
  const problems: string[] = [];
  if (nodes[0]?.kind === "divider") problems.push(`${where}: starts with a divider`);
  if (nodes[nodes.length - 1]?.kind === "divider") problems.push(`${where}: ends with a divider`);
  nodes.forEach((node, index) => {
    if (node.kind === "divider" && nodes[index + 1]?.kind === "divider") problems.push(`${where}: two dividers in a row`);
    if (node.kind === "submenu") {
      if (!node.nodes.length) problems.push(`${where}/${node.id}: empty submenu`);
      problems.push(...strayLayout(node.nodes, `${where}/${node.id}`));
    }
  });
  return problems;
}

describe("Champ-style menu bar layout", () => {
  it("has the 13 heads of the real Champ menu bar in the same order", () => {
    expect(MENU_BARS.map((bar) => bar.label.th)).toEqual(CHAMP_HEADS);
    expect(new Set(MENU_BARS.map((bar) => bar.id)).size).toBe(MENU_BARS.length);
  });

  it("shows every catalog item exactly once, so no permission id is lost or duplicated", () => {
    const shown = MENU_BARS.flatMap((bar) => menuBarItems(bar.nodes)).map((item) => item.id);
    expect(new Set(shown).size).toBe(shown.length);
    expect([...shown].sort()).toEqual(flattenMenuItems().map((item) => item.id).sort());
  });

  it("keeps every pre-upgrade item id and route reachable from the menu bar", () => {
    const baseline = JSON.parse(
      readFileSync(resolve(process.cwd(), "src/lib/__fixtures__/menu-before-champ-upgrade.json"), "utf8"),
    ) as { id: string; route: string }[];
    const shown = new Map(MENU_BARS.flatMap((bar) => menuBarItems(bar.nodes)).map((item) => [item.id, item.route]));
    // Screens deliberately moved after the baseline snapshot (id stays, route changes).
    const moved: Record<string, string> = { "purchase-vat-adjust": "/gl/purchasevatadjust", "withholding-tax-deduction": "/report/whtdeduction" };
    for (const previous of baseline) expect(shown.get(previous.id), previous.id).toBe(moved[previous.id] ?? previous.route);
  });

  it("never leaves a stray divider or an empty submenu", () => {
    expect(MENU_BARS.flatMap((bar) => strayLayout(bar.nodes, bar.id))).toEqual([]);
  });

  it("moves the two pricing screens to the sales head (B2) and keeps BC-only screens next to their Champ block", () => {
    const where = (id: string) => MENU_BARS.find((bar) => menuBarItems(bar.nodes).some((item) => item.id === id))?.id;
    expect(where("product-sale-price")).toBe("bill");
    expect(where("product-promotion")).toBe("bill");
    expect(where("product-price-adjust")).toBe("ic");
    for (const id of ["sale-reservation-flow", "sale-order-date-check", "etax-invoice"]) expect(where(id), id).toBe("bill");
    // 2026-10-09 (S2): Champ keeps the code masters under จัดการระบบ › กำหนดรหัสต่างๆ, not under สินค้าคงคลัง / บัญชี
    for (const id of ["barcode", "product-unit", "product-group", "brand", "warehouse", "gl-dimensions"]) expect(where(id), id).toBe("system");
    for (const id of ["gl-fiscal-years", "gl-year-end", "gl-ar-outstanding", "tax-forms"]) expect(where(id), id).toBe("gl");
  });

  it("uses the real Champ wording for renamed screens without touching id or route", () => {
    const items = new Map(flattenMenuItems().map((item) => [item.id, item]));
    for (const [id, th] of Object.entries(CHAMP_LABELS)) expect(items.get(id)?.label.th, id).toBe(th);
    expect(items.get("gl-journals")?.route).toBe("/gl/journals");
    expect(items.get("gl-post")?.label.key).toBe("gl_post_and_reverse");
  });

  it("keeps the old names searchable as aliases", () => {
    const items = new Map(flattenMenuItems().map((item) => [item.id, item]));
    expect(items.get("gl-journals")?.label.aliases).toContain("สมุดรายวัน");
    expect(items.get("sale-order")?.label.aliases).toContain("บันทึกใบสั่งขาย/สั่งจองสินค้า");
    expect(items.get("withholding-tax-deduction")?.label.aliases).toContain("ภาษีหัก ณ ที่จ่าย");
    expect(items.get("gl-post")?.label.aliases).toContain("ผ่านรายการและยกเลิก");
  });

  it("adds the Champ-only screens as pending items and opens the existing reverse-mode journal screen from gl-unpost", () => {
    const items = new Map(flattenMenuItems().map((item) => [item.id, item]));
    for (const id of CHAMP_ONLY_PENDING_IDS) {
      const item = items.get(id);
      expect(item?.route, id).toBe(`/pending/${id}`);
      expect(isMenuScreenPending(item?.route ?? ""), id).toBe(true);
    }
    const adj = items.get("purchase-receipt-date-adjust");
    expect(adj?.route).toBe("/procurement/delivery-dates");
    expect(isMenuScreenPending("/procurement/delivery-dates")).toBe(false);
    for (const [id, route] of [
      ["window-font-screen", "/settings/font"],
      ["window-font-report", "/settings/font-report"],
      ["window-font-report-title", "/settings/font-report-title"],
      ["window-font-report-column", "/settings/font-report-column"],
      ["window-font-report-detail", "/settings/font-report-detail"],
      ["window-font-report-company", "/settings/font-report-company"],
      ["help-about", "/settings/about"],
      ["help-standard-software", "/settings/standard-software"],
    ]) {
      const f = items.get(id);
      expect(f?.route).toBe(route);
      expect(isMenuScreenPending(route)).toBe(false);
    }

    // ยกเลิกเอกสารที่ผ่านรายการแล้ว: Champ 5.5 dropdown "บัญชี" มีเป็นเมนูของตัวเองต่อจาก "ผ่านรายการ"
    // (docs/kms/champ/11-menu-parity-audit-2026-10-09.md) → เปิดจอกลับรายการ (GLJournals mode=reverse) ที่ route เดิม
    const unpost = items.get("gl-unpost");
    expect(unpost?.route).toBe("/gl/unposting");
    expect(isMenuScreenPending("/gl/unposting")).toBe(false);
    // GeneralLedgerWorkbench หา item จาก GL_MENU_ITEMS — ไม่อยู่ในนี้จะขึ้น "ไม่พบหน้าบัญชี"
    expect(GL_MENU_ITEMS.some((entry) => entry.route === "/gl/unposting")).toBe(true);
  });

  it("places the items added to the transaction heads where Champ has them (2026-10-09)", () => {
    const shownIn = (barId: string, expected: string[]) =>
      menuBarItems(MENU_BARS.find((bar) => bar.id === barId)?.nodes ?? []).map((entry) => entry.id).filter((id) => expected.includes(id));
    const order = (barId: string, expected: string[]) => expect(shownIn(barId, expected), barId).toEqual(expected);
    order("po", [
      "purchase-requisition", "purchase-requisition-cancel", "purchase-requisition-approve", "purchase-requisition-approve-cancel",
      "purchase-order", "purchase-order-approve", "purchase-order-cancel",
      "purchase-landed-cost", "purchase-receipt-adjustment", "purchase-receipt-date-adjust",
    ]);
    order("bill", ["sale-reservation", "sale-order", "sale-order-approve", "sale-order-cancel", "sale-dispatch", "receive-deposit"]);
    order("ap", ["ap-billing-receipt", "ap-payment-prepare", "pay"]);
    order("cheque", ["cheque-received", "cheque-issued", "credit-card-receipts", "cheque-issued-prepare", "cheque-movement", "cheque-deposit"]);
    order("bank-cash", ["expense-record", "bank-service", "bank-contact"]);
    order("ic", [
      "product-price-adjust", "product-supplier-update", "goods-inspection",
      "stock-issue-request", "stock-issue-request-cancel", "stock-pickup", "stock-return",
      "stock-transfer-request", "stock-transfer-request-cancel", "stock-transfer",
      "stock-adjust", "stock-count-handheld", "reprocess", "stock-daily-sequence", "transaction-lock",
    ]);
    order("gl", ["gl-post", "gl-unpost", "gl-recalculate-posted"]);

    const submenuNodes = (barId: string, submenuId: string) => {
      const found = MENU_BARS.find((bar) => bar.id === barId)?.nodes.find((node) => node.kind === "submenu" && node.id === submenuId);
      return found?.kind === "submenu" ? found.nodes.map((node) => (node.kind === "item" ? node.item.id : node.kind)) : [];
    };
    expect(submenuNodes("cheque", "cheque-registers")).toEqual(["cheque-received", "cheque-issued", "credit-card-receipts", "divider", "cheque-issued-prepare"]);
    expect(submenuNodes("ic", "ic-counting")).toEqual(["stock-count-sheet", "stock-count", "stock-adjust", "divider", "stock-count-handheld"]);
    // Champ puts the three cheque/card reports under "รายงานข้อมูลหลัก"; the other daily reports are not built yet
    expect(MENU_BARS.find((bar) => bar.id === "cheque")?.nodes.flatMap((node) => (node.kind === "submenu" ? [node.id] : []))).toEqual(["cheque-registers", "cheque-reports-master"]);
    expect(submenuNodes("cheque", "cheque-reports-master")).toEqual(["cheque-received-report", "cheque-issued-report", "credit-card-report"]);
  });

  it("only binds system heads to functions the program already has", () => {
    const actions = MENU_BARS.flatMap((bar) => bar.nodes).filter((node) => node.kind === "action").map((node) => (node.kind === "action" ? node.id : ""));
    expect(actions).toEqual(["change-company", "change-password", "exit-program", "create-shortcut", "close-all-tabs", "open-manual"]);
    expect(menuBarItems(MENU_BARS.find((bar) => bar.id === "file")?.nodes ?? [])).toEqual([]);
  });

  it("lays out the หน้าต่าง head like Champ with the font submenu (2026-10-10)", () => {
    expect(topLevelOutline("window")).toEqual(["ปิดหน้าจอทั้งหมด", "▸ เลือกตัวอักษรในการแสดงผลต่าง"]);
    const fonts = MENU_BARS.find((bar) => bar.id === "window")?.nodes.find((node) => node.kind === "submenu" && node.id === "window-fonts");
    expect(fonts?.kind === "submenu" ? menuBarItems(fonts.nodes).map((entry) => entry.id) : []).toEqual([
      "window-font-screen", "window-font-report", "window-font-report-title",
      "window-font-report-column", "window-font-report-detail", "window-font-report-company",
    ]);
  });

  it("lays out the ช่วยเหลือ head like Champ (2026-10-10)", () => {
    expect(topLevelOutline("help")).toEqual([
      "About", "มาตรฐานซอฟท์แวร์ของกรมสรรพากร (Standard software)", "---", "คู่มือการใช้งาน",
    ]);
  });

  it("lays out the ไฟล์ head like Champ, opening the user and permission settings screens (2026-10-09)", () => {
    // Champ also has สร้างการเชื่อมโยงข้อมูล and เลือกใช้ข้อมูลตัวอย่างต่างๆ (desktop-only) — left out until the owner decides.
    expect(topLevelOutline("file")).toEqual([
      "เลือกบริษัททำงานใหม่", "กำหนดรหัสผู้ใช้งาน", "เปลี่ยนรหัสผ่าน", "กำหนดสิทธิ์การใช้งาน", "---", "ออกจากโปรแกรม",
    ]);
    const entries = menuBarEntries(MENU_BARS.find((bar) => bar.id === "file")?.nodes ?? []);
    expect(entries.map((entry) => entry.route)).toEqual(["/user", "/permissiongroup"]);
  });

  it("lays out the จัดการระบบ head like Champ with the code screens in a กำหนดรหัสต่างๆ submenu (2026-10-09)", () => {
    // Left out on purpose: ขอดูข้อมูล (raw-SQL viewer) and แสดง ToolTip — awaiting the owner's decision.
    expect(topLevelOutline("system")).toEqual([
      "กำหนดค่าเริ่มต้นบริษัท", "กำหนดรูปแบบรายวัน", "กำหนดวันที่ทำงานเริ่มต้น",
      "---",
      "▸ กำหนดรหัสต่างๆ", "กำหนดที่เก็บสินค้าตามคลัง", "กำหนดสินค้าตามคลัง", "กำหนดคลังที่เก็บสินค้าเริ่มต้น", "กำหนดสกุลเงิน",
      "กำหนดแผนที่หลัก", "กำหนดพนักงาน/พนักงานขาย", "กำหนดสาเหตุการลดหนี้/เพิ่มหนี้", "กำหนดกลุ่มลูกหนี้", "เงื่อนไขการชำระเงิน",
      "---",
      "สร้าง Shortcut", "แสดงข้อมูลลูกค้า", "แสดงข้อมูลสินค้า",
    ]);
    const codes = MENU_BARS.find((bar) => bar.id === "system")?.nodes.find((node) => node.kind === "submenu" && node.id === "system-codes");
    expect(codes?.kind === "submenu" ? menuBarItems(codes.nodes).map((entry) => entry.id) : []).toEqual([
      "gl-dimensions", "product-group", "brand", "warehouse", "product-unit", "barcode",
    ]);
    // the catalog items of the head are the pending screens plus the six code screens — nothing else
    const catalogIds = menuBarItems(MENU_BARS.find((bar) => bar.id === "system")?.nodes ?? []).map((entry) => entry.id);
    expect(catalogIds.filter((id) => !CHAMP_ONLY_PENDING_IDS.includes(id)).sort()).toEqual(
      ["barcode", "brand", "gl-dimensions", "product-group", "product-unit", "system-credit-debit-reasons", "system-currency", "system-customer-inquiry", "system-default-warehouse", "system-journal-format", "system-main-map", "system-payment-terms", "system-product-by-warehouse", "system-product-inquiry", "system-work-date", "warehouse"],
    );
  });

  it("opens only existing settings screens from route nodes, outside the permission catalog (2026-10-09)", () => {
    const routes = MENU_BARS.flatMap((bar) => allNodes(bar.nodes)).filter((node) => node.kind === "route");
    expect(routes.map((node) => (node.kind === "route" ? node.item.route : ""))).toEqual([
      "/user", "/permissiongroup", "/branch", "/productwarehousescreen", "/employee", "/debtorgroup",
    ]);
    const catalogIds = new Set(flattenMenuItems().map((entry) => entry.id));
    for (const node of routes) {
      if (node.kind !== "route") continue;
      expect(getSystemSettingConfig(node.item.route), node.item.route).toBeTruthy();
      expect(catalogIds.has(node.item.id), node.item.id).toBe(false);
      expect(isMenuBarRouteItem(node.item), node.item.id).toBe(true);
    }
    // menuBarItems stays the permission-catalog view: route nodes never leak into it
    expect(MENU_BARS.flatMap((bar) => menuBarItems(bar.nodes)).some((entry) => isMenuBarRouteItem(entry))).toBe(false);
  });
});

describe("menu bar language rows", () => {
  it("has no duplicate keys and 13 columns on every row of languages.tsv", () => {
    const { lines } = languageRows();
    const keys = lines.map((columns) => columns[0]);
    expect(keys.filter((key, index) => keys.indexOf(key) !== index)).toEqual([]);
    expect(lines.filter((columns) => columns.length !== 13).map((columns) => columns[0])).toEqual([]);
  });

  it("has a row for every head, submenu, action and item label with the same Thai text", () => {
    const { rows } = languageRows();
    const missing: string[] = [];
    for (const bar of MENU_BARS) {
      for (const label of [bar.label, ...collectLabels(bar.nodes)]) {
        const row = rows.get(label.key ?? "");
        if (!row) missing.push(`${label.key}: no row`);
        else if (row[1] !== label.th) missing.push(`${label.key}: tsv "${row[1]}" != "${label.th}"`);
        else if (!row[2]) missing.push(`${label.key}: no English`);
      }
    }
    expect(missing).toEqual([]);
  });
});

describe("ChampMenuBar", () => {
  it("renders the overview button first and the 13 heads as menu triggers in Champ order", () => {
    const html = renderToStaticMarkup(
      createElement(ChampMenuBar, {
        activeSection: "all",
        backendLanguage: {} as BackendLanguageDictionary,
        canAccessMenuItem: () => true,
        language: "th",
        navigationLabel: "เมนู",
        onAction: () => undefined,
        onOpenItem: () => undefined,
        onOpenNewItem: () => undefined,
        onOpenOverview: () => undefined,
        onSelectSection: () => undefined,
        overviewLabel: "ภาพรวม",
      }),
    );
    expect(html.match(/aria-haspopup="menu"/g)).toHaveLength(13);
    let cursor = html.indexOf("ภาพรวม");
    expect(cursor).toBeGreaterThan(-1);
    for (const head of CHAMP_HEADS) {
      const next = html.indexOf(`>${head}<`, cursor);
      expect(next, head).toBeGreaterThan(cursor);
      cursor = next;
    }
  });
});
