import { describe, expect, it } from "vitest";
import type { BankBook, BankCode, BankContact } from "./cheque-api";
import {
  areContactsDirty,
  BANK_BOOK_LIMITS,
  blankContact,
  bookInputFromDraft,
  canSaveBookDraft,
  canSaveCodeDraft,
  codeInputFromDraft,
  contactsToSave,
  draftFromBook,
  draftFromCode,
  duplicateContactCode,
  emptyBookDraft,
  emptyCodeDraft,
  formatDecimalText,
  isBookDraftDirty,
  isCodeDraftDirty,
  removedContactCodes,
} from "./cheque-forms";

const branch: BankCode = { kind: "bank_branch", code: "0123", name: "สาขาสีลม", bank_code: "BBL", user_group: "บัญชี", is_active: true, version: 3 };
const contact = (code: string, name = "คุณสมศรี", extra: Partial<BankContact> = {}): BankContact => ({ ...blankContact(), contact_code: code, name, ...extra });

describe("code draft", () => {
  it("starts empty and active, and a loaded code is not dirty", () => {
    expect(emptyCodeDraft()).toEqual({ code: "", name: "", bank_code: "", user_group: "", is_active: true });
    expect(isCodeDraftDirty(draftFromCode(branch), draftFromCode(branch))).toBe(false);
  });

  it("ignores whitespace-only edits but sees real ones", () => {
    const base = draftFromCode(branch);
    expect(isCodeDraftDirty({ ...base, name: "  สาขาสีลม " }, base)).toBe(false);
    expect(isCodeDraftDirty({ ...base, name: "สาขาสีลม 2" }, base)).toBe(true);
    expect(isCodeDraftDirty({ ...base, is_active: false }, base)).toBe(true);
    expect(isCodeDraftDirty({ ...base, bank_code: "" }, base)).toBe(true);
  });

  it("allows saving only with code and name and a real change", () => {
    const base = emptyCodeDraft();
    expect(canSaveCodeDraft(base, base)).toBe(false);
    expect(canSaveCodeDraft({ ...base, code: "BBL" }, base)).toBe(false);
    expect(canSaveCodeDraft({ ...base, code: "BBL", name: "  " }, base)).toBe(false);
    expect(canSaveCodeDraft({ ...base, code: "BBL", name: "ธนาคารกรุงเทพ" }, base)).toBe(true);
    const loaded = draftFromCode(branch);
    expect(canSaveCodeDraft(loaded, loaded)).toBe(false);
  });

  it("sends the parent bank only for branches and trims every text", () => {
    const draft = { code: " 0123 ", name: " สาขาสีลม ", bank_code: " BBL ", user_group: " บัญชี ", is_active: true };
    expect(codeInputFromDraft("bank_branch", draft)).toEqual({ kind: "bank_branch", code: "0123", name: "สาขาสีลม", bank_code: "BBL", user_group: "บัญชี", is_active: true });
    expect(codeInputFromDraft("bank", draft)).not.toHaveProperty("bank_code");
  });
});

describe("contact rows", () => {
  it("drops untouched rows and trims the rest before saving", () => {
    const rows = [contact(" C1 ", " คุณสมศรี "), blankContact(), contact("C2", "คุณวิชัย", { is_contact: false })];
    expect(contactsToSave(rows)).toEqual([contact("C1", "คุณสมศรี"), contact("C2", "คุณวิชัย", { is_contact: false })]);
  });

  it("is not dirty after only adding an empty row, and dirty after a real edit", () => {
    const base = [contact("C1")];
    expect(areContactsDirty([...base, blankContact()], base)).toBe(false);
    expect(areContactsDirty([contact("C1", "คุณสมศรี ใจดี")], base)).toBe(true);
    expect(areContactsDirty([contact("C1", "คุณสมศรี", { is_contact: false })], base)).toBe(true);
  });

  it("reports saved contacts that the new set no longer contains (case-insensitive)", () => {
    const base = [contact("C1"), contact("c2"), contact("C3")];
    expect(removedContactCodes([contact("c1"), contact("C2")], base)).toEqual(["C3"]);
    expect(removedContactCodes(base, base)).toEqual([]);
    expect(removedContactCodes([], base)).toEqual(["C1", "c2", "C3"]);
  });

  it("finds a duplicate code ignoring case and ignores blank rows", () => {
    expect(duplicateContactCode([contact("ab1"), blankContact(), contact(" AB1 ")])).toBe(" AB1 ".trim());
    expect(duplicateContactCode([contact("C1"), contact("C2"), blankContact(), blankContact()])).toBeNull();
  });
});

const book: BankBook = {
  bank_account_code: "KBANK-01",
  bank_name: "ธนาคารกสิกรไทย",
  account_number: "123-4-56789-0",
  account_name: "บริษัท รุ่งเรืองการค้า จำกัด",
  gl_account_code: "11120",
  currency_code: "THB",
  is_active: true,
  version: 4,
  bank_code: "KBANK",
  bank_branch_code: "0123",
  account_type_code: "CURRENT",
  gl_balance: "1250000.50",
};

describe("bank book draft", () => {
  it("starts empty and active, and a loaded book is not dirty or savable", () => {
    expect(emptyBookDraft()).toMatchObject({ bank_account_code: "", account_number: "", bank_code: "", is_active: true });
    const loaded = draftFromBook(book);
    expect(isBookDraftDirty(loaded, draftFromBook(book))).toBe(false);
    expect(canSaveBookDraft(loaded, draftFromBook(book))).toBe(false);
  });

  it("maps a legacy book without bank, branch or type to empty strings", () => {
    const legacy = draftFromBook({ ...book, bank_code: undefined, bank_branch_code: undefined, account_type_code: undefined });
    expect(legacy).toMatchObject({ bank_code: "", bank_branch_code: "", account_type_code: "" });
  });

  it("ignores whitespace-only edits but sees real ones, including the optional fields", () => {
    const base = draftFromBook(book);
    expect(isBookDraftDirty({ ...base, account_name: " " + base.account_name + "  " }, base)).toBe(false);
    expect(isBookDraftDirty({ ...base, account_type_code: "" }, base)).toBe(true);
    expect(isBookDraftDirty({ ...base, bank_branch_code: "0999" }, base)).toBe(true);
    expect(isBookDraftDirty({ ...base, is_active: false }, base)).toBe(true);
  });

  it("enables save only when the required fields are filled and something changed", () => {
    const blank = emptyBookDraft();
    expect(canSaveBookDraft(blank, blank)).toBe(false);
    const filled = { ...blank, bank_account_code: "KBANK-01", account_number: "123", account_name: "บัญชีหลัก", bank_code: "KBANK" };
    expect(canSaveBookDraft(filled, blank)).toBe(false); // ยังไม่ได้เลือกบัญชีผังบัญชี
    expect(canSaveBookDraft({ ...filled, gl_account_code: "11120" }, blank)).toBe(true);
    expect(canSaveBookDraft({ ...filled, gl_account_code: "11120", bank_code: "  " }, blank)).toBe(false);
  });

  it("sends trimmed values, always in baht, and never the version or balance", () => {
    const input = bookInputFromDraft({ ...draftFromBook(book), account_name: "  บัญชีหลัก ", account_number: " 123 " });
    expect(input).toMatchObject({ account_name: "บัญชีหลัก", account_number: "123", currency_code: "THB", bank_code: "KBANK" });
    expect(input).not.toHaveProperty("version");
    expect(input).not.toHaveProperty("gl_balance");
    expect(input).not.toHaveProperty("bank_name");
  });

  it("limits match the backend columns", () => {
    expect(BANK_BOOK_LIMITS).toEqual({ code: 20, accountNumber: 50, accountName: 255 });
  });
});

describe("formatDecimalText", () => {
  it("groups thousands and keeps exact decimals without rounding", () => {
    expect(formatDecimalText("1250000.5")).toBe("1,250,000.50");
    expect(formatDecimalText("0.1")).toBe("0.10");
    expect(formatDecimalText("12345678901234567890.12345678")).toBe("12,345,678,901,234,567,890.12345678");
    expect(formatDecimalText("1000")).toBe("1,000.00");
    expect(formatDecimalText("100.000000")).toBe("100.00");
  });

  it("handles negative and zero values", () => {
    expect(formatDecimalText("-4500.25")).toBe("-4,500.25");
    expect(formatDecimalText("-0.00")).toBe("0.00");
    expect(formatDecimalText("0")).toBe("0.00");
    expect(formatDecimalText("007.5")).toBe("7.50");
  });

  it("returns an empty string for anything that is not a plain decimal", () => {
    for (const value of [undefined, "", " ", "abc", "1e5", "1,000.00", "1.2.3", "--1"]) expect(formatDecimalText(value), String(value)).toBe("");
  });
});
