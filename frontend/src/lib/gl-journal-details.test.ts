import { describe, expect, it } from "vitest";
import { parseStatementCsv, reconciliationChanges, supportLabel } from "./gl-journal-details";

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
