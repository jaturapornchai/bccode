import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { GLStatementAccountFix, GLStatementSuggestionTarget } from "@/lib/general-ledger";
import { GLStatementSuggestionDialog, statementFixMessage, statementSuggestReasonText, statementTargetLabel } from "./gl-statement-suggestions";

const tr = (_key: string, fallback: string) => fallback;

const target: GLStatementSuggestionTarget = {
  target: "row",
  id: "bs-30",
  rowno: 30,
  suggestkey: "cash_and_equivalents",
  accounttype: "asset",
  accounts: [
    { accountcode: "11110", accountname: "เงินสดในมือ", isactive: true, reasons: [{ source: "iscash", count: 1 }] },
    { accountcode: "11120", accountname: "เงินฝากกระแสรายวัน", isactive: false, reasons: [{ source: "bank_accounts", count: 2 }, { source: "other_templates", count: 1, templates: ["BS-2568"] }] },
  ],
  inuse: [{ accountcode: "11130", accountname: "เงินฝากออมทรัพย์", target: "row", id: "bs-140", rowno: 140, title: "เงินฝากธนาคารที่มีภาระค้ำประกัน" }],
};

function renderDialog(overrides: Partial<Parameters<typeof GLStatementSuggestionDialog>[0]> = {}) {
  return renderToStaticMarkup(createElement(GLStatementSuggestionDialog, {
    open: true,
    target,
    targetLabel: "เงินสดและรายการเทียบเท่าเงินสด",
    inUseLabel: (ref) => statementTargetLabel(ref, tr),
    onClose: () => undefined,
    onAccept: () => undefined,
    tr,
    ...overrides,
  }));
}

// แนะนำบัญชีเป็นคำแนะนำเท่านั้น (ADR 2026-09-27-gl-statement-account-suggestions): ผู้ใช้ติ๊กเองทุกบัญชี ระบบไม่ติ๊กให้และไม่เพิ่มให้เอง
describe("GLStatementSuggestionDialog", () => {
  it("is a labelled modal dialog with nothing ticked and the add button disabled", () => {
    const html = renderDialog();
    expect(html).toContain('role="dialog"');
    expect(html).toContain('aria-modal="true"');
    expect(html).toMatch(/aria-labelledby="[^"]+"/);
    expect(html).toContain("บัญชีที่แนะนำสำหรับ “เงินสดและรายการเทียบเท่าเงินสด”");
    expect(html).toContain("แนะนำเฉพาะบัญชีหมวดสินทรัพย์ที่ลงรายการได้");
    expect(html.match(/type="checkbox"/g)).toHaveLength(2);
    expect(html).not.toMatch(/checked=""/);
    expect(html).toMatch(/<button[^>]*disabled=""[^>]*>.*เพิ่มบัญชีที่เลือก \(0\)/);
  });

  it("explains every suggestion and marks inactive accounts in words, not colour only", () => {
    const html = renderDialog();
    expect(html).toContain("ตั้งเป็นบัญชีเงินสดในผังบัญชี");
    expect(html).toContain("ผูกกับบัญชีธนาคาร 2 บัญชี · ผูกไว้ในบรรทัดชนิดเดียวกันของแม่แบบงบอื่น 1 แม่แบบ (BS-2568)");
    expect(html).toContain("ปิดใช้งาน");
  });

  it("lists accounts already used elsewhere without a checkbox", () => {
    const html = renderDialog();
    const inUse = html.slice(html.indexOf("อยู่ในบรรทัดอื่นของแม่แบบนี้แล้ว"));
    expect(inUse).toContain("11130");
    expect(inUse).toContain("บรรทัด 140 “เงินฝากธนาคารที่มีภาระค้ำประกัน”");
    expect(inUse).not.toContain('type="checkbox"');
    expect(renderDialog({ target: { ...target, inuse: [] } })).not.toContain("อยู่ในบรรทัดอื่นของแม่แบบนี้แล้ว");
  });

  it("renders nothing when closed or without a target", () => {
    expect(renderDialog({ open: false })).toBe("");
    expect(renderDialog({ target: null })).toBe("");
  });

  it("closes on Escape unless a child handled it, traps Tab and returns focus to the opener", () => {
    const source = readFileSync(resolve(process.cwd(), "src", "app", "gl", "gl-statement-suggestions.tsx"), "utf8");
    expect(source).toContain('event.key === "Escape" && !event.defaultPrevented');
    expect(source).toContain(`"button:not(:disabled), input:not(:disabled), select:not(:disabled), [tabindex]:not([tabindex='-1'])"`);
    expect(source).toContain("if (opener?.isConnected) opener.focus();");
    // ปุ่มปิดมุมขวาบนต้องมีชื่อไทยทั้ง aria-label และ title (ห้าม icon เปล่า)
    expect(source).toMatch(/aria-label=\{cancelLabel\}[^>]*title=\{cancelLabel\}|title=\{cancelLabel\}[^>]*aria-label=\{cancelLabel\}/);
  });
});

describe("statement suggestion texts", () => {
  it("names a line by number and title, a column by title", () => {
    expect(statementTargetLabel({ target: "row", rowno: 30, title: "เงินสด" }, tr)).toBe("บรรทัด 30 “เงินสด”");
    expect(statementTargetLabel({ target: "column", title: "กำไรสะสม" }, tr)).toBe("คอลัมน์ “กำไรสะสม”");
    expect(statementTargetLabel({ target: "row", rowno: 40, title: "  " }, tr)).toBe("บรรทัด 40 “—”");
  });

  it("shows the other-template codes and marks when the backend sent only some of them", () => {
    expect(statementSuggestReasonText({ source: "other_templates", count: 7, templates: ["A", "B", "C", "D", "E"] }, tr)).toBe("ผูกไว้ในบรรทัดชนิดเดียวกันของแม่แบบงบอื่น 7 แม่แบบ (A, B, C, D, E, …)");
    expect(statementSuggestReasonText({ source: "other_templates", count: 2, templates: ["A", "B"] }, tr)).toBe("ผูกไว้ในบรรทัดชนิดเดียวกันของแม่แบบงบอื่น 2 แม่แบบ (A, B)");
    expect(statementSuggestReasonText({ source: "unknown_source", count: 1 }, tr)).toBe("");
  });

  it("tells the user per line what was replaced, skipped or removed", () => {
    const fixes: GLStatementAccountFix[] = [
      { target: "row", id: "bs-30", rowno: 30, title: "เงินสด", headers: [{ accountcode: "11100", accountname: "เงินสดและรายการเทียบเท่าเงินสด", descendants: ["11110", "11120"], skipped: [{ accountcode: "11130", accountname: "เงินฝากออมทรัพย์", target: "row", id: "bs-140", rowno: 140, title: "เงินฝาก" }] }], removed: ["ZZZ"] },
      { target: "column", id: "eq-c4", title: "กำไรสะสม", headers: [{ accountcode: "33000", accountname: "กำไรสะสม", descendants: [], skipped: [] }], removed: [] },
    ];
    const message = statementFixMessage(fixes, (fix) => statementTargetLabel(fix, tr), tr);
    expect(message).toContain("บรรทัด 30 “เงินสด”: ");
    expect(message).toContain("แทนด้วยบัญชีย่อยที่ลงรายการได้ 2 บัญชี");
    expect(message).toContain("บัญชีย่อย 11130 อยู่ในบรรทัดอื่นของแม่แบบนี้แล้ว");
    expect(message).toContain("นำรหัสบัญชี ZZZ ออก");
    expect(message).toContain(" · คอลัมน์ “กำไรสะสม”: ");
    expect(message).toContain("ยังไม่มีบัญชีย่อยที่ลงรายการได้");
    expect(statementFixMessage([], () => "", tr)).toBe("");
  });
});
