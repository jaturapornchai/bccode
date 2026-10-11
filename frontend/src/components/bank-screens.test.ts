import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

// vitest ของโปรเจกต์ไม่มี DOM — ตรวจกติกา UX/ความปลอดภัยที่ต้องคงอยู่ของจอกำหนดรหัส จอผู้ติดต่อธนาคาร และจอสมุดเงินฝากจากซอร์ส
// (ตรรกะของฟอร์มทดสอบจริงใน lib/cheque-forms.test.ts และ lib/cheque-api.test.ts)
const read = (file: string) => readFileSync(resolve(process.cwd(), file), "utf8");
const codeScreen = read("src/components/bank-code-screen.tsx");
const contactScreen = read("src/components/bank-contact-screen.tsx");
const bookScreen = read("src/components/book-bank-screen.tsx");

describe.each([
  ["bank-code-screen", codeScreen],
  ["bank-contact-screen", contactScreen],
  ["book-bank-screen", bookScreen],
])("%s", (_name, source) => {
  it("follows the language prop through the backend dictionary, not localStorage or a local t(th,en)", () => {
    expect(source).toContain("useBackendLanguage(language");
    expect(source).not.toContain("localStorage");
    expect(source).not.toMatch(/\bt\(\s*"[^"]*"\s*,\s*"[^"]*"\s*\)/);
  });

  it("warns before the tab closes or the user switches away while there are unsaved edits", () => {
    expect(source).toContain('window.addEventListener("beforeunload"');
    expect(source).toContain("cheque_confirm_discard_title");
    expect(source).toContain("canLeave");
  });

  it("never lets an icon-only button go unlabeled", () => {
    const buttons = source.split("<Button").slice(1).map((chunk) => chunk.split("</Button>")[0] ?? "");
    for (const button of buttons.filter((chunk) => chunk.includes('size="icon"'))) expect(button).toContain("aria-label=");
  });
});

describe("bank-code-screen", () => {
  it("remounts per kind so a draft from one register never leaks into another", () => {
    expect(codeScreen).toContain("<BankCodeWorkbench key={kind}");
  });

  it("deletes only after a danger confirmation dialog and keeps the code read-only while editing", () => {
    expect(codeScreen).toContain('tone: "danger"');
    expect(codeScreen).toContain("bank_code_confirm_delete");
    expect(codeScreen).toContain("readOnly={selected !== null}");
  });

  it("enables save only for a real change and reloads after a version conflict", () => {
    expect(codeScreen).toContain("canSaveCodeDraft(draft, baseline)");
    expect(codeScreen).toContain("cause.status === 409");
  });
});

describe("bank-contact-screen", () => {
  it("asks before a save would remove saved contacts and warns on duplicate codes first", () => {
    expect(contactScreen).toContain("removedContactCodes(values, baseline)");
    expect(contactScreen).toContain('tone: "danger"');
    expect(contactScreen.indexOf("duplicateContactCode(values)")).toBeLessThan(contactScreen.indexOf("saveBankContacts("));
  });

  it("limits every text input to the column length used by the database", () => {
    for (const limit of ["code", "name", "position", "address", "telephone"]) expect(contactScreen).toContain(`maxLength={BANK_CONTACT_LIMITS.${limit}}`);
  });

  it("only offers active banks and branches in the pickers", () => {
    expect(contactScreen.match(/activeOnly: true/g)).toHaveLength(2);
  });
});

describe("book-bank-screen", () => {
  it("keeps the book code and account number read-only while editing and changes the number only through its own button", () => {
    expect(bookScreen).toContain("readOnly={selected !== null}");
    expect(bookScreen).toContain("changeBankBookNumber(selected.bank_account_code, selected.version, number, language)");
    expect(bookScreen).toContain("book_bank_change_account_no");
    expect(bookScreen).toContain("book_bank_confirm_change_no");
  });

  it("asks for confirmation before changing the account number and before deleting", () => {
    expect(bookScreen.indexOf("book_bank_confirm_change_no")).toBeLessThan(bookScreen.indexOf("changeBankBookNumber(selected"));
    expect(bookScreen.indexOf("book_bank_confirm_delete")).toBeLessThan(bookScreen.indexOf("deleteBankBook(selected"));
    expect(bookScreen).toContain('tone: "danger"');
  });

  it("keeps the unsaved edits of other fields when only the account number changes", () => {
    expect(bookScreen).toContain("setDraft((current) => ({ ...current, account_number: number }))");
    expect(bookScreen).toContain("setBaseline((current) => ({ ...current, account_number: number }))");
  });

  it("shows the GL balance as read-only text from the backend string, never through Number()", () => {
    expect(bookScreen).toContain("formatDecimalText(selected.gl_balance)");
    expect(bookScreen).toContain("<output");
    expect(bookScreen).not.toMatch(/\b(?:Number|parseFloat)\(/);
  });

  it("shows the contacts of the chosen bank and branch as a read-only table", () => {
    expect(bookScreen).toContain("listBankContacts(draft.bank_code, draft.bank_branch_code, language)");
    const table = bookScreen.slice(bookScreen.indexOf("<table"));
    expect(table).not.toContain("<Input");
    expect(table).not.toContain("<Checkbox");
  });

  it("clears the branch when the bank changes and only offers active codes", () => {
    expect(bookScreen).toContain('bank_branch_code: ""');
    expect(bookScreen.match(/activeOnly: true/g)).toHaveLength(3);
  });

  it("limits every text input to the column length used by the backend", () => {
    for (const limit of ["code", "accountNumber", "accountName"]) expect(bookScreen).toContain(`maxLength={BANK_BOOK_LIMITS.${limit}}`);
  });

  it("enables save only for a real change and reloads after a version conflict", () => {
    expect(bookScreen).toContain("canSaveBookDraft(draft, baseline)");
    expect(bookScreen).toContain("cause.status === 409");
  });
});
