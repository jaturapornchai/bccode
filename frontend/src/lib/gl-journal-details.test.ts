import { describe, expect, it } from "vitest";
import { type GLDetailVat, type GLDetailWithholding, defaultVatTaxType, parseStatementCsv, reconciliationChanges, supportLabel, vatPeriodChoices, vatPeriodPatch, vatPeriodValue } from "./gl-journal-details";

describe("statement evidence import",()=>{
  it("preserves decimal strings and stable file-row identities across retries",async()=>{
    const csv='transaction_date,direction,amount,bank_reference,description\n2026-09-20,1,0.10000001,R1,"ค่าธรรมเนียม, รายการทดสอบ"\n2026-09-20,1,0.2,R2,รับเงิน';
    const a=await parseStatementCsv(csv,"B1"), b=await parseStatementCsv(csv,"B1"), other=await parseStatementCsv(csv,"B2");
    expect(a).toEqual(b);expect(a[0].amount).toBe("0.10000001");expect(a[0].description).toBe("ค่าธรรมเนียม, รายการทดสอบ");
    expect(a[0].id).not.toBe(a[1].id);expect(other[0].id).not.toBe(a[0].id);expect(a[0].source_key).toMatch(/^sha256:[a-f0-9]{64}:2$/);
  });
  it("rejects missing bank, invalid amount and unfinished CSV quoting",async()=>{
    await expect(parseStatementCsv("transaction_date,direction,amount\n2026-09-20,1,0.1","")).rejects.toThrow();
    await expect(parseStatementCsv("transaction_date,direction,amount\n2026-09-20,1,1e3","B1")).rejects.toThrow();
    await expect(parseStatementCsv('transaction_date,direction,amount\n"2026-09-20,1,1',"B1")).rejects.toThrow();
  });
  it("sends only newly added reconciliation records, including allocation replacements",()=>{
    const old={id:"old",statement_line_id:"s1",line_number:1,amount:"0.1"};
    const delta=reconciliationChanges({matches:[old]},{matches:[old,{...old,id:"new",amount:"0.2"}],allocations:[{id:"alloc",ledger:"ar",document_id:"doc",line_number:1,amount:"1"}]});
    expect(delta.matches).toEqual([{...old,id:"new",amount:"0.2"}]);expect(delta.allocations).toEqual([{id:"alloc",ledger:"ar",document_id:"doc",line_number:1,amount:"1"}]);
  });
  it("labels evidence with business numbers and remaining amounts rather than UUIDs",()=>{
    const label=supportLabel("documents",{id:"uuid-hidden",ledger:"ar",document_no:"INV001",partner_code:"C001",amount:"1.20",remaining_amount:"0.20"});
    expect(label).toContain("INV001");expect(label).toContain("0.20");expect(label).not.toContain("uuid-hidden");
  });
});
it("retains optional statement values and original row positions", async()=>{
  const rows=await parseStatementCsv("transaction_date,direction,amount,value_date,balance_after\n\n2026-09-20,1,0.1,2026-09-21,-0.20000001", "B1");
  expect(rows[0]).toMatchObject({value_date:"2026-09-21",balance_after:"-0.20000001"});
  expect(rows[0].source_key).toMatch(/:3$/);
});
it("distinguishes withdrawals sharing an ID across relation types",()=>{
  expect(reconciliationChanges({withdrawals:[{kind:"allocation",id:"same",reason:"old"}]},{withdrawals:[{kind:"allocation",id:"same",reason:"old"},{kind:"match",id:"same",reason:"new"}]}).withdrawals).toEqual([{kind:"match",id:"same",reason:"new"}]);
});

const wht = (base: string): GLDetailWithholding => ({ id: "W1", wht_direction: 1, form_type: "PND53", partner_code: "TRANS", payment_date: "2026-09-15", income_tax_type: "3_tres", condition_type: 1, wht_rate: "3", base_amount: base, tax_amount: "3000" });

describe("reconciliationChanges — withholding tax base (ต้องแก้ได้เสมอ)", () => {
  it("sends the whole withholding set when the base was edited after posting", () => {
    const changed = { ...wht("95000"), tax_amount: undefined };
    expect(reconciliationChanges({ withholdings: [wht("100000")] }, { withholdings: [changed] }).withholdings).toEqual([changed]);
  });
  it("sends nothing for withholdings that did not change", () => {
    expect(reconciliationChanges({ withholdings: [wht("100000")] }, { withholdings: [wht("100000")] }).withholdings).toBeUndefined();
  });
});

const vat = (base: string): GLDetailVat => ({ id: "V1", tax_type: 2, document_type: 1, tax_invoice_no: "IV6909-001", tax_invoice_date: "2026-09-10", tax_period_year: 2026, tax_period_month: 9, partner_name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", base_amount: base, zero_rate_amount: "0", exempt_amount: "0", vat_rate: "7", vat_amount: "7000" });

describe("reconciliationChanges — VAT tax base (ต้องแก้ได้เสมอ)", () => {
  it("sends the whole VAT set when a line was edited after posting", () => {
    const changed = { ...vat("95000"), vat_amount: undefined };
    expect(reconciliationChanges({ vats: [vat("100000")] }, { vats: [changed] }).vats).toEqual([changed]);
  });
  it("sends nothing for VAT lines that did not change", () => {
    expect(reconciliationChanges({ vats: [vat("100000")] }, { vats: [vat("100000")] }).vats).toBeUndefined();
  });
});

describe("VAT detail helpers", () => {
  it("defaults to output VAT only for the sales journal (UV)", () => {
    expect(defaultVatTaxType("UV")).toBe(2);
    expect(defaultVatTaxType("SV")).toBe(1);
    expect(defaultVatTaxType(undefined)).toBe(1);
  });
  it("round-trips the tax period between YYYY-MM and year/month fields", () => {
    expect(vatPeriodValue({ tax_period_year: 2026, tax_period_month: 9 })).toBe("2026-09");
    expect(vatPeriodValue({})).toBe("");
    expect(vatPeriodValue({ tax_period_year: 2026, tax_period_month: 13 })).toBe("");
    expect(vatPeriodPatch("2026-09")).toEqual({ tax_period_year: 2026, tax_period_month: 9 });
    // ว่าง = ไม่มีงวดทั้งคู่ และหายไปจาก JSON ที่ส่ง backend
    expect(JSON.stringify(vatPeriodPatch(""))).toBe("{}");
  });
  it("offers 12 months back to 6 months ahead of the voucher and keeps an older saved period", () => {
    const choices = vatPeriodChoices("2026-01-15");
    expect(choices).toHaveLength(19);
    expect(choices[0]).toBe("2025-01");
    expect(choices).toContain("2026-01");
    expect(choices[choices.length - 1]).toBe("2026-07");
    expect(vatPeriodChoices("2026-01-15", "2024-03")[0]).toBe("2024-03");
    expect(vatPeriodChoices("2026-01-15", "2026-01")).toHaveLength(19);
  });
});
