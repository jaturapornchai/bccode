import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { ChampMenuBar } from "@/app/menu/champ-menu-bar";
import type { BackendLanguageDictionary } from "./backend-language";
import { MENU_BARS, menuBarItems, type MenuBarNode } from "./menu-bars";
import type { MenuLabel } from "./menu-data";
import { flattenMenuItems } from "./menu-data";
import { isMenuScreenPending } from "./menu-screen-status";

/** Heads of the real Champ 5.5 menu bar, left to right (bcspec screenshot "เมนูโปรแกรมทั้งหมด"). */
const CHAMP_HEADS = [
  "ไฟล์", "จัดซื้อ", "จัดขาย", "เจ้าหนี้", "ลูกหนี้", "เช็ค/บัตร", "ธนาคาร/เงินสด",
  "สินค้าคงคลัง", "สินทรัพย์", "บัญชี", "จัดการระบบ", "หน้าต่าง", "ช่วยเหลือ",
];

/** Items the real Champ menu has but BC has no screen for yet — added as "รอพัฒนา" (no real route). */
const CHAMP_ONLY_PENDING_IDS = [
  "purchase-requisition-cancel", "purchase-requisition-approval-cancel", "purchase-order-approve",
  "purchase-receipt-date-adjust", "sale-reservation", "sale-dispatch-order", "ap-payment-prepare",
  "billing-note-auto", "billing-note-auto-print", "customer-label-print", "cheque-issue-prepare",
  "cheque-movement", "bank-service", "product-supplier-adjust", "stock-issue-request-cancel",
  "stock-transfer-request-cancel", "stock-count-handheld", "transaction-lock", "product-status-report",
  "product-price-setting-report",
];

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
};

const tsvPath = resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv");

function languageRows(): { lines: string[][]; rows: Map<string, string[]> } {
  const lines = readFileSync(tsvPath, "utf8").split(/\r?\n/).filter((line) => line.length > 0).map((line) => line.split("\t"));
  return { lines, rows: new Map(lines.map((columns) => [columns[0] ?? "", columns])) };
}

function collectLabels(nodes: MenuBarNode[]): MenuLabel[] {
  return nodes.flatMap((node) => {
    if (node.kind === "item") return [node.item.label];
    if (node.kind === "submenu") return [node.label, ...collectLabels(node.nodes)];
    if (node.kind === "action") return [node.label];
    return [];
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
    for (const previous of baseline) expect(shown.get(previous.id), previous.id).toBe(previous.route);
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
    for (const id of ["barcode", "product-unit", "product-group", "brand", "warehouse"]) expect(where(id), id).toBe("ic");
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

  it("adds the Champ-only screens as pending items and reverses posting through the existing reverse screen", () => {
    const items = new Map(flattenMenuItems().map((item) => [item.id, item]));
    for (const id of CHAMP_ONLY_PENDING_IDS) {
      const item = items.get(id);
      expect(item?.route, id).toBe(`/pending/${id}`);
      expect(isMenuScreenPending(item?.route ?? ""), id).toBe(true);
    }
    // ยกเลิกเอกสารที่ผ่านรายการแล้ว = แท็บในจอผ่านรายการ ไม่ใช่เมนูแยก (AGENTS.md)
    expect(items.has("gl-unpost")).toBe(false);
  });

  it("only binds system heads to functions the program already has", () => {
    const actions = MENU_BARS.flatMap((bar) => bar.nodes).filter((node) => node.kind === "action").map((node) => (node.kind === "action" ? node.id : ""));
    expect(actions).toEqual(["change-company", "change-password", "exit-program", "create-shortcut", "close-all-tabs", "open-manual"]);
    for (const id of ["file", "system", "window", "help"]) {
      expect(menuBarItems(MENU_BARS.find((bar) => bar.id === id)?.nodes ?? [])).toEqual([]);
    }
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
