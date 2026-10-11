import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// vitest ของโปรเจกต์รันบน node (ไม่มี DOM) จึงตรวจ "สัญญา" ของจอเช็คจากซอร์ส: กฎที่ผิดแล้วผู้ใช้เสียหายจริง
// (ลืมกัน dirty, ลบโดยไม่ถาม, ปุ่มที่ยังไม่ควรมี, สีฝัง) ส่วนตรรกะฟอร์ม/คำสั่งทดสอบใน src/lib/cheque-*.test.ts

const read = (file: string) => readFileSync(resolve(process.cwd(), file), "utf8");

function stripComments(source: string) {
  return source
    .replace(/\{?\/\*[\s\S]*?\*\/\}?/g, (comment) => comment.replace(/[^\n]/g, ""))
    .replace(/(?:^|[ \t])\/\/[^\n]*$/gm, "");
}

const register = stripComments(read("src/components/cheque-register-screen.tsx"));
const statusDoc = stripComments(read("src/components/cheque-status-doc-screen.tsx"));
const dialogs = stripComments(read("src/components/cheque-doc-dialogs.tsx"));
const screens = { "cheque-register-screen": register, "cheque-status-doc-screen": statusDoc, "cheque-doc-dialogs": dialogs };

describe("both cheque screens", () => {
  it.each(["cheque-register-screen", "cheque-status-doc-screen"] as const)("%s takes its language from the prop and warns before unsaved data is lost", (name) => {
    const source = screens[name];
    expect(source).toContain("useBackendLanguage(language");
    expect(source).not.toContain("localStorage");
    expect(source).toContain('addEventListener("beforeunload"');
    expect(source).toContain('removeEventListener("beforeunload"');
    expect(source).toContain("cheque_confirm_discard");
  });

  it.each(Object.keys(screens) as (keyof typeof screens)[])("%s uses theme variables only, no hard-coded colors", (name) => {
    const source = screens[name];
    expect(source).not.toMatch(/#[0-9a-fA-F]{3,8}\b/);
    expect(source).not.toMatch(/\b(?:rgb|rgba|hsl|hsla)\(/);
    expect(source).not.toMatch(/\b(?:text|bg|border)-(?:red|green|amber|yellow|blue|gray|slate|zinc)-\d/);
  });

  it("never keeps a key action icon-only: remove-line has both aria-label and title", () => {
    expect(statusDoc).toMatch(/aria-label=\{`\$\{tr\("cheque_remove_line"[\s\S]{0,160}?title=\{tr\("cheque_remove_line"/);
  });
});

describe("cheque register screen (cheque-received / cheque-issued)", () => {
  it("asks the owner to confirm a risky edit (D8) with the server's message, then resends with confirm", () => {
    expect(register).toContain('["cheque_err_referenced", "cheque_err_status_confirm"]');
    expect(register).toContain("cause.status !== 409");
    expect(register).toContain("send(false)");
    expect(register).toContain("send(true)");
    expect(register).toContain("updateChequeRegister(");
  });

  it("deletes with 'registers cancel' only after a danger confirm dialog", () => {
    const before = register.indexOf('tone: "danger"');
    const call = register.indexOf("cancelChequeRegister({");
    expect(before).toBeGreaterThan(-1);
    expect(call).toBeGreaterThan(before);
    expect(register).toContain("cheque_confirm_delete_register");
  });

  it("has no prepare-cheque button yet; it lands with CB-9", () => {
    expect(register).not.toContain("cheque_btn_prepare");
    expect(register).not.toContain("prepare/unused");
  });

  it("shows the references table (item, document no., date, amount, description) and ignores a stale open", () => {
    for (const key of ["cheque_col_item", "cheque_ref_col_date", "cheque_ref_col_amount", "cheque_references_empty"]) expect(register, key).toContain(key);
    expect(register).toContain("openSeq");
  });
});

describe("cheque status document screen (419-426)", () => {
  it("creates and cancels through the cheque API, and cancel always asks for a reason", () => {
    expect(statusDoc).toContain("createChequeDoc(");
    expect(statusDoc).toContain("cancelChequeDoc({");
    expect(statusDoc).toContain("<ChequeReasonDialog");
    expect(dialogs).toMatch(/reason\.trim\(\)/);
  });

  it("shows the journal voucher number and the GL posting status", () => {
    expect(statusDoc).toContain("selected.journal_docno");
    expect(statusDoc).toContain("chequeGlStatus(selected.gl_status)");
  });

  it("picks cheques through the selectable list with the document date and the 423 / 426 conditions", () => {
    expect(statusDoc).toContain("<ChequePickerDialog");
    for (const field of ["transflag: flag", "docDate:", "cancelType: config.cancelTypeCombo", "cancelPassed: config.cancelPassedBox"]) expect(statusDoc, field).toContain(field);
  });

  it("drives every difference between the eight documents from the config, not from route or flag checks", () => {
    for (const flagField of ["config.cancelTypeCombo", "config.cancelPassedBox", "config.hasFee", "config.showRemark"]) expect(statusDoc, flagField).toContain(flagField);
    expect(statusDoc).not.toMatch(/flag === 4(?:19|2[0-6])/);
  });

  it("warns in red under the 423 cancel type and clears the lines only after a confirm when the rule changes", () => {
    expect(statusDoc).toMatch(/role="note"/);
    expect(statusDoc).toContain("text-destructive");
    expect(statusDoc).toContain("cheque_cancel_notice");
    expect(statusDoc).toContain("changeCondition(");
    expect(statusDoc).toContain("cheque_change_kind_confirm");
  });
});

describe("menu wiring", () => {
  const menu = stripComments(read("src/app/menu/main-menu-screen.tsx"));

  it("sends the ten cheque routes to the dedicated screens before the generic ERP screen", () => {
    const cheque = menu.indexOf("chequeScreenFromRoute(activeTab.route)");
    const erp = menu.indexOf("isErpTransactionRoute(activeTab.route)");
    expect(cheque).toBeGreaterThan(-1);
    expect(erp).toBeGreaterThan(cheque);
    expect(menu).toContain("<ChequeRegisterScreen key={activeTab.route}");
    expect(menu).toContain("<ChequeStatusDocScreen key={activeTab.route}");
  });

  it("no longer has the cheque doctypes in the ERP configs or the module registry", () => {
    const erp = read("src/lib/erp-transaction.ts");
    const modules = read("src/lib/erp-transaction-modules.ts");
    for (const key of ["chequereceived", "chequeissued", "chequedeposit", "chequeissuedcancel"]) {
      expect(erp, key).not.toMatch(new RegExp(`\\b${key}\\b`));
      expect(modules, key).not.toMatch(new RegExp(`\\b${key}\\b`));
    }
  });
});
