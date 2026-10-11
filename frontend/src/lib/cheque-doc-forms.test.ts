import { describe, expect, it } from "vitest";
import type { ChequeDoc, ChequeRegister } from "./cheque-api";
import { chequeDocConfig, type ChequeDocConfig } from "./cheque-doc-config";
import {
  addDocLines,
  canSaveDocDraft,
  canSaveRegisterDraft,
  chequeStamp,
  docBookPolicy,
  docDraftTotals,
  docInputFromDraft,
  docLineFromRegister,
  docLineNet,
  docMissingFields,
  docRevertStatuses,
  draftFromDoc,
  draftFromRegister,
  emptyDocDraft,
  emptyRegisterDraft,
  invalidFeeChequeIds,
  isDocDraftDirty,
  isRegisterDraftDirty,
  moneyInput,
  registerBookRequired,
  registerHomeAmount,
  registerInputFromDraft,
  registerMissingFields,
  removeDocLine,
  setDocLineFee,
  type DocDraft,
} from "./cheque-doc-forms";

const config = (flag: number): ChequeDocConfig => {
  const found = chequeDocConfig(flag);
  if (!found) throw new Error(`no config for ${flag}`);
  return found;
};

const register = (overrides: Partial<ChequeRegister> = {}): ChequeRegister => ({
  id: "0b8e6a52-3c1d-4f0a-9e2b-5d7c8a1f4e30",
  cheque_type: 0,
  doc_no: "CQIN6910-0001",
  doc_date: "2026-10-01",
  due_date: "2026-10-20",
  partner_code: "AR-0012",
  partner_name: "บริษัท สยามพาณิชย์ จำกัด",
  cheque_no: "1023456",
  bank_code: "BBL",
  bank_branch_code: "0123",
  book_code: "SCB-01",
  status: 0,
  ext_status_code: "",
  currency_code: "THB",
  exchange_rate: "1",
  amount: "15000.00",
  home_amount: "15000.00",
  balance_amount: "0.00",
  description: "ชำระค่าสินค้างวดที่ 2",
  sale_code: "",
  department_code: "",
  source_kind: "",
  prepare_doc_no: "",
  print_count: 0,
  version: 1,
  created_at: "2026-10-01T03:00:00Z",
  created_by: "somchai",
  updated_at: "2026-10-01T03:00:00Z",
  updated_by: "somchai",
  ...overrides,
});

const second = register({ id: "9c1f0a52-3c1d-4f0a-9e2b-5d7c8a1f4e31", cheque_no: "1023457", amount: "2500.50", home_amount: "2500.50" });

describe("moneyInput", () => {
  it("strips commas and spaces pasted into an amount field", () => {
    expect(moneyInput("1,234.5", "")).toBe("1234.5");
    expect(moneyInput(" 12 000", "")).toBe("12000");
  });

  it("keeps the previous value when the typed text is not a money amount", () => {
    expect(moneyInput("12a", "12")).toBe("12");
    expect(moneyInput("1.234", "1.23")).toBe("1.23");
    expect(moneyInput("-5", "")).toBe("");
    expect(moneyInput("12345678901234567", "1")).toBe("1");
  });

  it("allows an empty field and a trailing decimal point while typing", () => {
    expect(moneyInput("", "5")).toBe("");
    expect(moneyInput("5.", "5")).toBe("5.");
  });
});

describe("chequeStamp", () => {
  it("joins user and date-time, in Buddhist years for Thai", () => {
    const thai = chequeStamp("th", "สมชาย", "2026-10-11T08:30:00Z");
    expect(thai.startsWith("สมชาย — ")).toBe(true);
    expect(thai).toContain("2569");
    expect(chequeStamp("en", "somchai", "2026-10-11T08:30:00Z")).toContain("2026");
  });

  it("returns whatever part exists and an empty string for nothing", () => {
    expect(chequeStamp("th", "สมชาย", undefined)).toBe("สมชาย");
    expect(chequeStamp("th", undefined, undefined)).toBe("");
  });
});

describe("register draft", () => {
  it("starts a new cheque on today with status on-hand", () => {
    const draft = emptyRegisterDraft("2026-10-11");
    expect(draft).toMatchObject({ doc_date: "2026-10-11", due_date: "2026-10-11", status: 0, amount: "", doc_no: "" });
  });

  it("reads a saved cheque back, turning missing optional fields into empty text", () => {
    const draft = draftFromRegister(register({ book_code: undefined as unknown as string, sale_code: undefined as unknown as string }));
    expect(draft.book_code).toBe("");
    expect(draft.sale_code).toBe("");
    expect(draft).toMatchObject({ partner_code: "AR-0012", cheque_no: "1023456", amount: "15000.00", status: 0 });
  });

  it("does not count extra spaces as an edit", () => {
    const base = draftFromRegister(register());
    expect(isRegisterDraftDirty({ ...base, cheque_no: " 1023456 ", description: `${base.description}  ` }, base)).toBe(false);
    expect(isRegisterDraftDirty({ ...base, amount: "15000.01" }, base)).toBe(true);
  });

  it("requires a book for issued cheques and for received cheques that are not on-hand, cancelled or discounted", () => {
    expect(registerBookRequired(1, 0)).toBe(true);
    expect(registerBookRequired(0, 0)).toBe(false);
    expect(registerBookRequired(0, 4)).toBe(false);
    expect(registerBookRequired(0, 5)).toBe(false);
    for (const status of [1, 2, 3, 6]) expect(registerBookRequired(0, status), String(status)).toBe(true);
  });

  it("lists missing required fields in screen order", () => {
    const received = emptyRegisterDraft("2026-10-11");
    expect(registerMissingFields(0, received)).toEqual(["partner_code", "cheque_no", "bank_code", "bank_branch_code", "amount"]);
    expect(registerMissingFields(1, received)).toEqual(["partner_code", "cheque_no", "bank_code", "bank_branch_code", "book_code", "amount"]);
    expect(registerMissingFields(0, { ...received, doc_date: "", due_date: "" })).toEqual(expect.arrayContaining(["doc_date", "due_date"]));
  });

  it("rejects zero, negative and over-precise amounts", () => {
    const base = { ...draftFromRegister(register()) };
    for (const amount of ["", "0", "0.00", "-5", "1.234", "abc"]) expect(registerMissingFields(0, { ...base, amount }), amount).toContain("amount");
    expect(registerMissingFields(0, { ...base, amount: "0.01" })).toEqual([]);
  });

  it("enables save only when complete and actually changed", () => {
    const saved = draftFromRegister(register());
    expect(canSaveRegisterDraft(0, saved, saved)).toBe(false);
    expect(canSaveRegisterDraft(0, { ...saved, description: "แก้ไขคำอธิบาย" }, saved)).toBe(true);
    expect(canSaveRegisterDraft(0, { ...saved, cheque_no: "" }, saved)).toBe(false);
    const blank = emptyRegisterDraft("2026-10-11");
    expect(canSaveRegisterDraft(0, { ...saved, doc_no: "" }, blank)).toBe(true);
  });

  it("computes the baht amount as text without float rounding", () => {
    expect(registerHomeAmount("15000")).toBe("15000.00");
    expect(registerHomeAmount("12.5")).toBe("12.50");
    expect(registerHomeAmount("0.1")).toBe("0.10");
    expect(registerHomeAmount("abc")).toBe("");
  });

  it("builds a THB command with rate 1, trimmed text, and the salesperson only for received cheques", () => {
    const draft = { ...draftFromRegister(register()), cheque_no: " 1023456 ", sale_code: "S-01", department_code: " D1 " };
    const received = registerInputFromDraft(0, draft);
    expect(received).toMatchObject({ cheque_type: 0, currency_code: "THB", exchange_rate: "1", cheque_no: "1023456", sale_code: "S-01", department_code: "D1" });
    expect(registerInputFromDraft(1, draft)).toMatchObject({ cheque_type: 1, sale_code: "" });
  });
});

describe("status document lines", () => {
  it("takes the cheque amount and the baht amount from the register row", () => {
    expect(docLineFromRegister(register({ amount: "100.00", home_amount: "100.00" }))).toMatchObject({
      cheque_id: "0b8e6a52-3c1d-4f0a-9e2b-5d7c8a1f4e30",
      cheque_amount: "100.00",
      amount: "100.00",
      fee_amount: "",
      description: "ชำระค่าสินค้างวดที่ 2",
    });
  });

  it("adds picked cheques once, removes by id and sets the fee of one line", () => {
    const first = addDocLines([], [register(), second]);
    expect(first.map((line) => line.cheque_no)).toEqual(["1023456", "1023457"]);
    expect(addDocLines(first, [register(), second])).toHaveLength(2);
    expect(removeDocLine(first, second.id).map((line) => line.cheque_no)).toEqual(["1023456"]);
    const fee = setDocLineFee(first, second.id, "10.25");
    expect(fee.map((line) => line.fee_amount)).toEqual(["", "10.25"]);
  });

  it("nets a line's fee off its amount, blank fee counting as zero", () => {
    const [line] = addDocLines([], [second]);
    expect(docLineNet(line!)).toBe("2500.50");
    expect(docLineNet({ ...line!, fee_amount: "0.50" })).toBe("2500.00");
  });
});

describe("status document totals and validation", () => {
  const lines = setDocLineFee(addDocLines([], [register(), second]), second.id, "10.25");

  it("totals amount, fee and net exactly for a document that has fees", () => {
    expect(docDraftTotals(config(420), lines)).toEqual({ total: "17500.50", fee: "10.25", net: "17490.25" });
  });

  it("gives the deposit slip (419) only a total: no fee, net equal to the total", () => {
    expect(docDraftTotals(config(419), lines)).toEqual({ total: "17500.50", fee: "0.00", net: "17500.50" });
  });

  it("flags fees that are not money or exceed the cheque amount, never a blank fee", () => {
    const bad = setDocLineFee(lines, register().id, "15000.01");
    expect(invalidFeeChequeIds(config(420), bad)).toEqual([register().id]);
    expect(invalidFeeChequeIds(config(420), setDocLineFee(lines, register().id, "abc"))).toEqual([register().id]);
    expect(invalidFeeChequeIds(config(420), setDocLineFee(lines, register().id, "15000.00"))).toEqual([]);
    expect(invalidFeeChequeIds(config(420), lines)).toEqual([]);
    expect(invalidFeeChequeIds(config(419), bad)).toEqual([]);
  });

  const draftWith = (overrides: Partial<DocDraft>): DocDraft => ({ ...emptyDocDraft("2026-10-25"), ...overrides });

  it("needs a date, the book when the policy requires it, and at least one cheque", () => {
    expect(docMissingFields(config(420), draftWith({}))).toEqual(["book_code", "lines"]);
    expect(docMissingFields(config(420), draftWith({ book_code: "SCB-01", lines }))).toEqual([]);
    expect(docMissingFields(config(420), draftWith({ doc_date: "", book_code: "SCB-01", lines }))).toEqual(["doc_date"]);
  });

  it("does not ask 423 for a book while the cancel type is on-hand, but does for the other types", () => {
    expect(docMissingFields(config(423), draftWith({ cancel_type: 0, lines }))).toEqual([]);
    expect(docMissingFields(config(423), draftWith({ cancel_type: 2, lines }))).toEqual(["book_code"]);
    expect(docBookPolicy(config(423), draftWith({ cancel_type: 0 })).none).toBe(true);
  });

  it("keeps 424's book optional", () => {
    expect(docMissingFields(config(424), draftWith({ lines }))).toEqual([]);
  });

  it("enables save only with a complete draft and valid fees", () => {
    const ready = draftWith({ book_code: "SCB-01", lines });
    expect(canSaveDocDraft(config(420), ready)).toBe(true);
    expect(canSaveDocDraft(config(420), { ...ready, lines: setDocLineFee(lines, register().id, "99999.00") })).toBe(false);
    expect(canSaveDocDraft(config(420), draftWith({ book_code: "SCB-01" }))).toBe(false);
  });

  it("treats only user-entered data as a change, not the default date or stray spaces", () => {
    const baseline = emptyDocDraft("2026-10-25");
    expect(isDocDraftDirty({ ...baseline, doc_no: "  " }, baseline)).toBe(false);
    expect(isDocDraftDirty({ ...baseline, lines }, baseline)).toBe(true);
    expect(isDocDraftDirty({ ...baseline, description: "ฝากเช็คประจำวัน" }, baseline)).toBe(true);
  });
});

describe("docInputFromDraft", () => {
  const lines = setDocLineFee(addDocLines([], [register(), second]), second.id, " 10.25 ");

  it("sends ids and per-line fees plus the totals, with no cancel fields for an ordinary document", () => {
    const input = docInputFromDraft(config(420), { ...emptyDocDraft("2026-10-25"), book_code: " SCB-01 ", description: " ผ่านแล้ว ", lines });
    expect(input).toEqual({
      transflag: 420,
      doc_no: "",
      doc_date: "2026-10-25",
      book_code: "SCB-01",
      cancel_type: undefined,
      cancel_passed: undefined,
      description: "ผ่านแล้ว",
      lines: [
        { cheque_id: register().id, fee_amount: "" },
        { cheque_id: second.id, fee_amount: "10.25" },
      ],
      totals: { total: "17500.50", fee: "10.25", net: "17490.25" },
    });
  });

  it("sends no fee for the deposit slip, even if a fee was typed before", () => {
    const input = docInputFromDraft(config(419), { ...emptyDocDraft("2026-10-25"), book_code: "SCB-01", lines });
    expect(input.lines.every((line) => line.fee_amount === "")).toBe(true);
    expect(input.totals).toEqual({ total: "17500.50", fee: "0.00", net: "17500.50" });
  });

  it("sends the cancel type for 423 and drops the book when the type is on-hand", () => {
    const base = { ...emptyDocDraft("2026-10-25"), lines, book_code: "SCB-01" };
    expect(docInputFromDraft(config(423), { ...base, cancel_type: 0 })).toMatchObject({ cancel_type: 0, book_code: "", cancel_passed: undefined });
    expect(docInputFromDraft(config(423), { ...base, cancel_type: 2 })).toMatchObject({ cancel_type: 2, book_code: "SCB-01" });
  });

  it("sends the cleared-cheque tick for 426 only", () => {
    const base = { ...emptyDocDraft("2026-10-25"), lines, book_code: "SCB-01", cancel_passed: true };
    expect(docInputFromDraft(config(426), base)).toMatchObject({ transflag: 426, cancel_passed: true, cancel_type: undefined });
    expect(docInputFromDraft(config(425), base).cancel_passed).toBeUndefined();
  });
});

describe("saved documents", () => {
  const saved: ChequeDoc = {
    transflag: 420,
    doc_no: "CQRC6910-0001",
    doc_date: "2026-10-25",
    branch_code: "00000",
    book_code: "SCB-01",
    cancel_type: null,
    cancel_passed: false,
    description: "",
    total_amount: "17500.50",
    fee_amount: "10.25",
    net_amount: "17490.25",
    gl_status: "posted",
    journal_docno: "JV6910-0007",
    status: "active",
    cancel_reason: "",
    cancelled_at: null,
    cancelled_by: "",
    version: 1,
    created_at: "2026-10-25T03:00:00Z",
    created_by: "somchai",
    updated_at: "2026-10-25T03:00:00Z",
    updated_by: "somchai",
    lines: [
      { line_no: 1, cheque_id: register().id, cheque_no: "1023456", partner_code: "AR-0012", partner_name: "บริษัท สยามพาณิชย์ จำกัด", bank_code: "BBL", bank_branch_code: "0123", cheque_date: "2026-10-01", due_date: "2026-10-20", description: "", from_status: 1, amount: "15000.00", fee_amount: "0.00", net_amount: "15000.00" },
      { line_no: 2, cheque_id: second.id, cheque_no: "1023457", partner_code: "AR-0012", partner_name: "บริษัท สยามพาณิชย์ จำกัด", bank_code: "BBL", bank_branch_code: "0123", cheque_date: "2026-10-01", due_date: "2026-10-20", description: "", from_status: 6, amount: "2500.50", fee_amount: "10.25", net_amount: "2490.25" },
    ],
  };

  it("shows a saved document on the form with the amount used for both cheque and baht columns", () => {
    const draft = draftFromDoc(saved);
    expect(draft).toMatchObject({ doc_no: "CQRC6910-0001", book_code: "SCB-01", cancel_type: 0, cancel_passed: false });
    expect(draft.lines.map((line) => [line.cheque_amount, line.amount, line.fee_amount])).toEqual([
      ["15000.00", "15000.00", "0.00"],
      ["2500.50", "2500.50", "10.25"],
    ]);
  });

  it("lists the statuses the cheques revert to on cancel, once each and sorted", () => {
    expect(docRevertStatuses(saved.lines)).toEqual([1, 6]);
    expect(docRevertStatuses([{ from_status: 2 }, { from_status: 2 }, { from_status: 1 }])).toEqual([1, 2]);
    expect(docRevertStatuses([])).toEqual([]);
  });
});
