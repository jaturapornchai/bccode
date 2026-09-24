import { describe, expect, it } from "vitest";
import { type GLDetailVat, type GLDetailWithholding, defaultVatTaxType, detailTaxIdTarget, validThaiTaxId, journalDetailsProblem, normalizeBranchNo, normalizeJournalDetails, parseStatementCsv, reconciliationChanges, taxDigits, supportLabel, vatClaimTiming, vatPeriodChoices, vatPeriodPatch, vatPeriodValue } from "./gl-journal-details";

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
  it("sends an explicit empty set when every row was removed (clears them on a posted journal)", () => {
    expect(reconciliationChanges({ withholdings: [wht("100000")], vats: [vat("100000")] }, { withholdings: [], vats: [] })).toMatchObject({ withholdings: [], vats: [] });
    expect(reconciliationChanges({ withholdings: [wht("100000")] }, {}).withholdings).toEqual([]);
  });
  it("omits the key when there were no rows before and none after (unchanged)", () => {
    const delta = reconciliationChanges({}, { withholdings: [] });
    expect("withholdings" in delta).toBe(false);
    expect("vats" in delta).toBe(false);
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
  it("defaults to output VAT from the book type (4 = sales journal), never from the book code", () => {
    expect(defaultVatTaxType(4)).toBe(2);
    expect(defaultVatTaxType(5)).toBe(1);
    expect(defaultVatTaxType(1)).toBe(1);
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

// ประกาศอธิบดีฯ VAT ฉบับที่ 4 ข้อ 2 (แก้โดยฉบับที่ 76): ใช้สิทธิภายหลังได้ไม่เกิน 6 เดือนนับแต่เดือนถัดจากเดือนที่ออกใบกำกับ
describe("vatClaimTiming — เดือนใช้สิทธิภาษีซื้อเทียบเดือนที่ออกใบกำกับ", () => {
  const purchase = (invoiceDate: string, year?: number, month?: number, extra: Record<string, unknown> = {}) =>
    ({ tax_type: 1, document_type: 1, claim_status: 1, tax_invoice_date: invoiceDate, tax_period_year: year, tax_period_month: month, ...extra });
  it("ใบกำกับ ม.ค. ใช้สิทธิ ม.ค. = ปกติ; ก.พ.–ก.ค. = ต้องเขียนข้อความในใบกำกับ; ส.ค. = เกินกำหนด", () => {
    expect(vatClaimTiming(purchase("2026-01-31", 2026, 1))).toBeNull();
    expect(vatClaimTiming(purchase("2026-01-31", 2026, 2))).toEqual({ kind: "late", months: 1, periodYear: 2026, periodMonth: 2 });
    expect(vatClaimTiming(purchase("2026-01-01", 2026, 7))).toEqual({ kind: "late", months: 6, periodYear: 2026, periodMonth: 7 });
    expect(vatClaimTiming(purchase("2026-01-15", 2026, 8))).toEqual({ kind: "exceeded" });
  });
  it("ข้ามปี: ใบกำกับ ธ.ค. 2568 ใช้สิทธิได้ถึง มิ.ย. 2569 (ไม่ใช้ Date จึงไม่เพี้ยนตามเขตเวลา)", () => {
    expect(vatClaimTiming(purchase("2025-12-31", 2026, 6))).toEqual({ kind: "late", months: 6, periodYear: 2026, periodMonth: 6 });
    expect(vatClaimTiming(purchase("2025-12-31", 2026, 7))).toEqual({ kind: "exceeded" });
    expect(vatClaimTiming(purchase("2025-12-01", 2025, 12))).toBeNull();
  });
  it("งวดก่อนเดือนที่ออกใบกำกับ = before", () => {
    expect(vatClaimTiming(purchase("2026-03-01", 2026, 2))).toEqual({ kind: "before" });
    expect(vatClaimTiming(purchase("2026-01-10", 2025, 12))).toEqual({ kind: "before" });
  });
  it("ใบเพิ่มหนี้: ช่วง 6 เดือนเหมือนใบกำกับ แต่ไม่แนะนำข้อความในใบกำกับ (ประกาศพูดถึงใบกำกับภาษี)", () => {
    expect(vatClaimTiming(purchase("2026-01-10", 2026, 3, { document_type: 2 }))).toBeNull();
    expect(vatClaimTiming(purchase("2026-01-10", 2026, 9, { document_type: 2 }))).toEqual({ kind: "exceeded" });
    expect(vatClaimTiming(purchase("2026-01-10", 2025, 12, { document_type: 2 }))).toEqual({ kind: "before" });
  });
  it("ใบลดหนี้ (ม.82/10 ลดในเดือนที่ได้รับ): ไม่มีเพดาน 6 เดือน ตรวจแค่งวดก่อนเดือนของใบลดหนี้ เหมือน backend", () => {
    expect(vatClaimTiming(purchase("2026-01-10", 2026, 1, { document_type: 3 }))).toBeNull();
    expect(vatClaimTiming(purchase("2026-01-10", 2026, 9, { document_type: 3 }))).toBeNull();
    expect(vatClaimTiming(purchase("2026-01-10", 2025, 12, { document_type: 3 }))).toEqual({ kind: "credit_note_before" });
  });
  it("ปีงวดเป็น พ.ศ. (≥ 2400) = buddhist_year ทุกประเภทแถว เหมือน backend vat_period_year_buddhist", () => {
    expect(vatClaimTiming(purchase("2026-01-10", 2569, 1))).toEqual({ kind: "buddhist_year" });
    expect(vatClaimTiming(purchase("2026-01-10", 2569, 1, { tax_type: 2, claim_status: undefined }))).toEqual({ kind: "buddhist_year" });
    expect(vatClaimTiming(purchase("2026-01-10", 2399, 1))).toEqual({ kind: "exceeded" });
  });
  it("ไม่เข้าเงื่อนไข → null: ภาษีขาย, ไม่ได้ใช้สิทธิ, ไม่มีงวด, วันที่ใบกำกับผิดรูป", () => {
    expect(vatClaimTiming(purchase("2026-01-10", 2026, 9, { tax_type: 2, claim_status: undefined }))).toBeNull();
    for (const status of [2, 3, 4]) expect(vatClaimTiming(purchase("2026-01-10", 2026, 9, { claim_status: status }))).toBeNull();
    expect(vatClaimTiming(purchase("2026-01-10"))).toBeNull();
    expect(vatClaimTiming(purchase("", 2026, 9))).toBeNull();
    expect(vatClaimTiming(purchase("2026-13-01", 2026, 9))).toBeNull();
  });
});

describe("tax id / branch normalisation", () => {
  it("accepts a pasted tax id with dashes and spaces without truncating first", () => {
    expect(taxDigits("0-1055-12345-67-8")).toBe("0105512345678");
    expect(taxDigits(" 0 1055 12345 67 8 ")).toBe("0105512345678");
    expect(taxDigits("\u0E50-\u0E51\u0E50\u0E55\u0E55-12345-67-8")).toBe("0105512345678");
  });
  it("pads branch numbers to 5 digits and keeps blank blank", () => {
    expect(normalizeBranchNo("0")).toBe("00000");
    expect(normalizeBranchNo("12")).toBe("00012");
    expect(normalizeBranchNo("")).toBe("");
    expect(normalizeBranchNo("123456")).toBe("123456");
  });
});

describe("normalizeJournalDetails — never sends an empty amount", () => {
  it("turns blank amounts into 0, omits blank tax amounts, strips tax ids and pads branches", () => {
    const details = normalizeJournalDetails({
      vats: [{ ...vat("100000"), zero_rate_amount: "", exempt_amount: "", vat_amount: "", partner_tax_id: "0-1055-12345-67-8", partner_branch_no: "0" }],
      withholdings: [{ ...wht("5000"), tax_amount: "", payee_tax_id: "3 1005 01234 56 7", payee_branch_no: "" }],
      documents: [{ id: "d1", ledger: "ar", partner_code: "C1", document_no: "IV1", document_date: "2026-09-01", branch_code: "00000", document_kind: 1, balance_side: 1, amount: "", currency_code: "THB", control_account_code: "" }],
    });
    const vatRow = details?.vats?.[0];
    expect(vatRow?.zero_rate_amount).toBe("0");
    expect(vatRow?.exempt_amount).toBe("0");
    expect(vatRow && "vat_amount" in vatRow).toBe(false);
    expect(vatRow?.partner_tax_id).toBe("0105512345678");
    expect(vatRow?.partner_branch_no).toBe("00000");
    const whtRow = details?.withholdings?.[0];
    expect(whtRow && "tax_amount" in whtRow).toBe(false);
    expect(whtRow?.payee_tax_id).toBe("3100501234567");
    expect(whtRow?.payee_branch_no).toBe("");
    expect(details?.documents?.[0].amount).toBe("0");
  });
});

describe("journalDetailsProblem — names the section, row and field", () => {
  const tr = (_key: string, fallback: string) => fallback;
  it("requires the withholding rate instead of silently using 0%", () => {
    const problem = journalDetailsProblem({ withholdings: [wht("5000"), { ...wht("3000"), wht_rate: "" }] }, tr);
    expect(problem).toMatchObject({ section: "withholdings", row: 2, field: "wht_rate" });
    expect(problem?.message).toContain("แถวที่ 2");
    expect(problem?.message).toContain("อัตราภาษี");
  });
  it("requires the VAT rate", () => {
    expect(journalDetailsProblem({ vats: [{ ...vat("100000"), vat_rate: " " }] }, tr)).toMatchObject({ section: "vats", row: 1, field: "vat_rate" });
  });
  it("reports a tax id that is not 13 digits with the digit count", () => {
    const problem = journalDetailsProblem({ vats: [{ ...vat("100000"), partner_tax_id: "0-1055-12345" }] }, tr);
    expect(problem).toMatchObject({ field: "partner_tax_id" });
    expect(problem?.message).toContain("10 หลัก");
  });
  it("accepts a pasted 13-digit tax id with dashes and a short branch", () => {
    expect(journalDetailsProblem({ vats: [{ ...vat("100000"), partner_tax_id: "0-1055-12345-67-8", partner_branch_no: "0" }] }, tr)).toBeNull();
  });
  it("reports a remark longer than 500 characters", () => {
    expect(journalDetailsProblem({ withholdings: [{ ...wht("5000"), remark: "ก".repeat(501) }] }, tr)).toMatchObject({ field: "remark" });
  });
  it("starts a partner error with the noun คู่ค้า from languages.tsv, not the screen heading", async () => {
    const { readFileSync } = await import("node:fs");
    const { resolve } = await import("node:path");
    const rows = new Map(readFileSync(resolve(process.cwd(), "..", "backend", "assets", "language", "languages.tsv"), "utf8").split(/\r?\n/).map((line) => line.split("\t")).map((cols) => [cols[0], cols[1] ?? ""] as const));
    const thai = (key: string, fallback: string) => rows.get(key) || fallback;
    const problem = journalDetailsProblem({ partners: [{ partner_code: "C001", name_th: "บริษัท สยามวัสดุ จำกัด", tax_id: "0-1055-12345" }] } as Parameters<typeof journalDetailsProblem>[0], thai);
    expect(problem?.message.startsWith("คู่ค้า แถวที่ 1:")).toBe(true);
  });
});

// คู่ค้า: คำนำหน้า/อำเภอ/จังหวัด/รหัสไปรษณีย์ ไม่บังคับ — ถ้ากรอกต้องถูกรูปแบบ และข้อความบอกแถว ช่อง และวิธีแก้
describe("partner optional address fields", () => {
  const tr = (_key: string, fallback: string) => fallback;
  const partner = (extra: Record<string, string> = {}) => ({ partners: [{ partner_code: "S001", name_th: "บริษัท ขนส่งไทยเร็ว จำกัด", is_customer: false, is_supplier: true, ...extra }] }) as Parameters<typeof journalDetailsProblem>[0];
  it("accepts blank optional fields and a 5-digit postcode (Thai digits pasted too)", () => {
    expect(journalDetailsProblem(partner(), tr)).toBeNull();
    expect(journalDetailsProblem(partner({ title_name: "บริษัท", addr_district: "บางนา", addr_province: "กรุงเทพมหานคร", addr_postcode: "๑๐๒๖๐" }), tr)).toBeNull();
  });
  it("names the row and field of a postcode that is not 5 digits", () => {
    const problem = journalDetailsProblem(partner({ addr_postcode: "1026" }), tr);
    expect(problem).toMatchObject({ section: "partners", row: 1, field: "addr_postcode" });
    expect(problem?.message).toContain("5 หลัก");
    expect(problem?.message).toContain("ตอนนี้ 4 หลัก");
  });
  it("counts Thai characters (not bytes) for the text limits", () => {
    expect(journalDetailsProblem(partner({ addr_province: "ก".repeat(50) }), tr)).toBeNull();
    expect(journalDetailsProblem(partner({ addr_province: "ก".repeat(51) }), tr)).toMatchObject({ field: "addr_province" });
  });
  it("trims the optional fields and drops them when blank before sending", () => {
    const [row] = normalizeJournalDetails(partner({ title_name: "  นาย ", addr_district: "   ", addr_postcode: "10-260" }))?.partners ?? [];
    expect(row).toMatchObject({ title_name: "นาย", addr_postcode: "10260" });
    expect(row).not.toHaveProperty("addr_district");
  });
});

describe("supportLabel wording", () => {
  it("names a GL line as บรรทัดที่ {0} and takes every word from the dictionary", () => {
    const row = { id: "J1:2", docno: "PV6909-0001", line_number: 2, bank_account_code: "B01", amount: "500.00", remaining_amount: "100.00" };
    expect(supportLabel("bank-lines", row)).toBe("PV6909-0001 · บรรทัดที่ 2 · B01 · 500.00 · คงเหลือ 100.00");
    const english: Record<string, string> = { gl_line_nth: "Line {0}", gl_support_remaining: "Remaining {0}", gl_support_this_journal: "This voucher" };
    expect(supportLabel("bank-lines", { ...row, docno: undefined }, (key, fallback) => english[key] ?? fallback)).toBe("This voucher · Line 2 · B01 · 500.00 · Remaining 100.00");
  });
});

// หลักตรวจสอบเลขผู้เสียภาษี (สูตรเดียวกับ backend whtcert.ValidThaiTaxID) + หาแถวจากชื่อช่องที่ backend ส่งมา (UAT S3 2026-09-24)
describe("Thai tax id check digit", () => {
  it("accepts a valid id with or without dashes and rejects a wrong check digit", () => {
    expect(validThaiTaxId("0105561001239")).toBe(true);
    expect(validThaiTaxId("0-1055-61001-23-9")).toBe(true);
    expect(validThaiTaxId("0105561001238")).toBe(false);
    expect(validThaiTaxId("010556100123")).toBe(false);
    expect(validThaiTaxId("")).toBe(false);
  });
  it("finds the row whose tax id the backend rejected", () => {
    const details = { partners: [
      { partner_code: "OK", name_th: "ถูก", tax_id: "0105561001239", is_customer: true, is_supplier: false, is_active: true },
      { partner_code: "BAD", name_th: "ผิด", tax_id: "0105561001238", is_customer: true, is_supplier: false, is_active: true },
    ] };
    expect(detailTaxIdTarget(details, "tax_id")).toEqual({ section: "partners", row: 2, field: "tax_id" });
    expect(detailTaxIdTarget(details, "partner_code")).toBeNull();
    expect(detailTaxIdTarget({ partners: [details.partners[0]] }, "tax_id")).toBeNull();
  });
});
