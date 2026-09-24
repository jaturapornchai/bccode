import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { GLDetailWithholding } from "@/lib/gl-journal-details";
import type { WhtReportRow } from "@/lib/thai-tax";
import type { ConfirmDialogOptions } from "@/components/ui/confirm-dialog";
import { WhtCertificatePanel, WhtReceivedCertificateDetails, findRecordedWithholding, leaveWhtCertificate, whtCertificateFigures, whtCertificateIssueDate, whtRecordRef, whtSnapshotView } from "./wht-certificate-panel";

const row = (overrides: Partial<WhtReportRow> = {}): WhtReportRow => ({
  rowid: "wht-J1#0",
  journalid: "J1",
  docno: "RV6909-0001",
  docdate: "2026-09-10",
  partnercode: "CUS-001",
  partnername: "บริษัท ก่อสร้างโครงการเอ จำกัด",
  taxid: "0105558002001",
  address: "99 ถนนพระราม 2 กรุงเทพมหานคร",
  description: "ค่าขนส่งวัสดุก่อสร้าง",
  baseamount: "20000.00",
  whtamount: "200.00",
  whtamounttext: "สองร้อยบาทถ้วน",
  netamount: "19800.00",
  ratepercent: "1.00",
  taxbasesource: "recorded",
  formtype: "PND53",
  incometype: "3_tres",
  condition: 1,
  paiddate: "2026-09-10",
  certificateno: "WT-0042",
  note: "",
  ...overrides,
});

const company = { code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234" };
const panelProps = { company, holdingcode: "rungrueng", businesscode: "01", language: "th" as const };

// ค่าที่เลือกไว้ของช่อง "ลำดับที่ในแบบ" ใน HTML ที่ render (React ใส่ selected ให้ option ที่ตรง value)
const selectedForm = (html: string) => {
  const select = html.slice(html.indexOf('data-field="form"'), html.indexOf("</select>", html.indexOf('data-field="form"')));
  return /<option value="([^"]*)"(?: disabled="")? selected="">/.exec(select)?.[1];
};

describe("WhtCertificatePanel — แบบยื่นเริ่มต้นตามที่บันทึก", () => {
  it("PND3 → ภ.ง.ด.3, PND2 → ภ.ง.ด.2, PND53 → ภ.ง.ด.53", () => {
    expect(selectedForm(renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row({ formtype: "PND3" }) })))).toBe("3");
    expect(selectedForm(renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row({ formtype: "PND2" }) })))).toBe("2");
    expect(selectedForm(renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row({ formtype: "PND53" }) })))).toBe("53");
  });

  // review 2026-09-24: ไม่รู้แบบ (ฐานประมาณ) → ยังไม่เลือก ให้ผู้ใช้เลือกเอง (เดิมตั้ง ภ.ง.ด.53 ให้ ผู้รับเงินบุคคลธรรมดาได้ใบติ๊กแบบผิด)
  it("ไม่มีแบบที่บันทึก (ฐานประมาณ) → ยังไม่เลือกแบบ และมีตัวเลือกให้เลือก", () => {
    const html = renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row({ formtype: "", taxbasesource: "inferred" }) }));
    expect(selectedForm(html)).toBe("");
    expect(html).toContain("เลือกแบบยื่น");
  });

  it("ฐานประมาณที่ backend อ่านแบบยื่นจากชื่อบัญชี (ภ.ง.ด.3) → เริ่มที่ ภ.ง.ด.3 ไม่ใช่ 53", () => {
    expect(selectedForm(renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row({ formtype: "PND3", taxbasesource: "inferred" }) })))).toBe("3");
  });
});

describe("WhtReceivedCertificateDetails — ภาษีถูกหัก ดูอย่างเดียว", () => {
  it("ลูกค้าเป็นผู้หักภาษี กิจการเราเป็นผู้ถูกหัก และไม่มีปุ่มออก/พิมพ์ 50 ทวิ", () => {
    const html = renderToStaticMarkup(createElement(WhtReceivedCertificateDetails, { row: row(), company }));
    const payer = html.indexOf("ผู้มีหน้าที่หักภาษี (ลูกค้า/ผู้จ่ายเงิน)");
    const payee = html.indexOf("ผู้ถูกหักภาษี (กิจการของเรา)");
    expect(payer).toBeGreaterThan(-1);
    expect(payee).toBeGreaterThan(payer);
    // ชื่อลูกค้าอยู่ใต้ป้ายผู้หัก ชื่อบริษัทเราอยู่ใต้ป้ายผู้ถูกหัก
    expect(html.indexOf("บริษัท ก่อสร้างโครงการเอ จำกัด")).toBeGreaterThan(payer);
    expect(html.indexOf("บริษัท ก่อสร้างโครงการเอ จำกัด")).toBeLessThan(payee);
    expect(html.indexOf(company.name)).toBeGreaterThan(payee);
    expect(html).toContain("ภ.ง.ด.53 (นิติบุคคล)");
    expect(html).toContain("WT-0042");
    expect(html).toContain("20,000.00");
    expect(html).not.toContain("<button");
    expect(html).not.toContain("<iframe");
  });

  it("ค่าที่ไม่ได้บันทึกแสดง 'ยังไม่ระบุ' ไม่เดา", () => {
    const html = renderToStaticMarkup(createElement(WhtReceivedCertificateDetails, { row: row({ certificateno: "", paiddate: "", formtype: "", condition: 0 }), company: null }));
    expect(html.match(/ยังไม่ระบุ/g)?.length).toBeGreaterThanOrEqual(5);
  });
});

describe("WhtCertificatePanel — snapshot ผู้จ่าย/ผู้รับเงินในใบสำคัญ", () => {
  const item = (overrides: Partial<GLDetailWithholding> = {}): GLDetailWithholding => ({
    id: "W1", wht_direction: 1, form_type: "PND53", partner_code: "CUS-001", payment_date: "2026-09-10", income_tax_type: "3_tres",
    condition_type: 1, wht_rate: "1", base_amount: "20000", tax_amount: "200", wht_cert_no: "WT-0042", ...overrides,
  });

  it("finds the recorded withholding that matches the report row (amounts compared exactly)", () => {
    const found = findRecordedWithholding({ details: { withholdings: [item({ id: "OTHER", partner_code: "CUS-999" }), item()] } }, row());
    expect("item" in found && found.item.id).toBe("W1");
  });

  it("does not guess when two rows match or none matches", () => {
    expect(findRecordedWithholding({ details: { withholdings: [item(), item({ id: "W2" })] } }, row())).toEqual({ problem: "ambiguous" });
    expect(findRecordedWithholding({ details: { withholdings: [item({ base_amount: "19999.99" })] } }, row())).toEqual({ problem: "not_found" });
    expect(findRecordedWithholding({ details: { withholdings: [item({ wht_direction: 2 })] } }, row())).toEqual({ problem: "not_found" });
  });

  it("matches the exact withholding id from the report row, even when amounts repeat", () => {
    const twins = { details: { withholdings: [item(), item({ id: "W2" })] } };
    const found = findRecordedWithholding(twins, row({ withholdingid: "W2" }));
    expect("item" in found && found.item.id).toBe("W2");
    expect(findRecordedWithholding(twins, row({ withholdingid: "GONE" }))).toEqual({ problem: "not_found" });
  });

  it("shows the stored snapshot first, then the registry values", () => {
    const view = whtSnapshotView(item({ payee_name: "ชื่อที่บันทึกในใบสำคัญ", payee_branch_no: "00001", wht_book_no: "12", remark: "ออกใบแทน" }), row(), company);
    expect(view.payee_name).toBe("ชื่อที่บันทึกในใบสำคัญ");
    expect(view.payee_tax_id).toBe("0105558002001");
    expect(view.payee_branch_no).toBe("00001");
    expect(view.payer_name).toBe(company.name);
    expect(view.payer_address).toBe("");
    expect(view.wht_book_no).toBe("12");
    expect(view.remark).toBe("ออกใบแทน");
  });

  it("renders payer/payee branch, remark and the save-to-voucher button; payer name is read-only when the registry has it", () => {
    const html = renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row() }));
    expect(html).toContain('data-field="payee.branch"');
    expect(html).toContain('data-field="payer.branch"');
    expect(html).toContain('data-field="remark"');
    expect(html).toContain("บันทึกลงใบสำคัญ");
    const payerName = html.slice(html.indexOf('data-field="payer.name"') - 200, html.indexOf('data-field="payer.name"'));
    expect(payerName).toContain("readOnly");
  });

  it("explains why saving is not possible for an estimated (not recorded) row", () => {
    const html = renderToStaticMarkup(createElement(WhtCertificatePanel, { ...panelProps, row: row({ taxbasesource: "inferred" }) }));
    expect(html).toContain("ยังไม่ได้บันทึกรายการภาษีหัก");
  });
});

// adversarial review 2026-09-24: edits were dropped silently when the tab or period changed
describe("leaveWhtCertificate — asks before unsaved payer/payee edits are dropped", () => {
  const tr = (_key: string, fallback: string) => fallback;
  it("goes on at once when nothing is waiting to be saved", async () => {
    const confirm = vi.fn<(options: ConfirmDialogOptions) => Promise<boolean>>(async () => false);
    expect(await leaveWhtCertificate(false, confirm, tr)).toBe(true);
    expect(confirm).not.toHaveBeenCalled();
  });
  it("asks, tells the user to press บันทึกลงใบสำคัญ, and follows the answer", async () => {
    const cancel = vi.fn<(options: ConfirmDialogOptions) => Promise<boolean>>(async () => false);
    expect(await leaveWhtCertificate(true, cancel, tr)).toBe(false);
    expect(String(cancel.mock.calls[0]?.[0]?.description)).toContain("บันทึกลงใบสำคัญ");
    expect(await leaveWhtCertificate(true, vi.fn(async () => true), tr)).toBe(true);
  });
});

// 50 ทวิ ของรายการที่บันทึกแล้ว: อ้างรายการด้วย journalid + withholdingid และไม่ส่งยอด/แบบ/เงื่อนไขจากจอ (backend ใช้ค่าที่บันทึก)
describe("whtRecordRef / whtCertificateFigures — recorded withholdings print from the voucher", () => {
  const values = { form: "7", condition: "1", conditionNote: "", incomeType: "3_tres", incomeNote: "ค่าขนส่ง", paidDate: "2026-09-10" };
  it("references a recorded row by journalid + withholdingid (from the row, else the matched record)", () => {
    expect(whtRecordRef(row({ withholdingid: "W-1" }))).toEqual({ journalid: "J1", withholdingid: "W-1" });
    expect(whtRecordRef(row(), { id: "W-9" })).toEqual({ journalid: "J1", withholdingid: "W-9" });
  });
  it("sends no reference for an estimated row or when the withholding id is unknown", () => {
    expect(whtRecordRef(row({ taxbasesource: "inferred", withholdingid: "W-1" }))).toBeNull();
    expect(whtRecordRef(row())).toBeNull();
    expect(whtRecordRef(row({ journalid: "" }), { id: "W-1" })).toBeNull();
  });
  it("leaves form, condition and amounts blank when locked so mainapi fills them from the record", () => {
    const locked = whtCertificateFigures(true, values, row());
    expect(locked).toEqual({ form: "", condition: "", conditionnote: "", incomes: [{ type: "3_tres", paiddate: "", amount: "", tax: "", note: "" }] });
    // review 2026-09-24: ประเภท "อื่น ๆ" ส่งประเภทที่บันทึกคู่กับรายละเอียดที่พิมพ์ — backend ใช้ข้อความจากจอเฉพาะเมื่อประเภทตรงกับรายการ
    // (ส่งประเภทว่าง = ข้อความหายเงียบ ๆ)
    expect(whtCertificateFigures(true, { ...values, incomeType: "other" }, row()).incomes[0]).toMatchObject({ type: "other", note: "ค่าขนส่ง" });
  });
  // review 2026-09-24: รายการที่บันทึกวันที่ออกหนังสือรับรองแล้ว backend พิมพ์วันที่นั้นเสมอ → จอแสดงอ่านอย่างเดียวและส่งค่านั้น
  it("uses the recorded certificate date (read-only) when the record has one", () => {
    expect(whtCertificateIssueDate(true, "2026-09-20", "2026-09-24")).toEqual({ value: "2026-09-20", fromRecord: true });
    expect(whtCertificateIssueDate(true, " ", "2026-09-24")).toEqual({ value: "2026-09-24", fromRecord: false });
    expect(whtCertificateIssueDate(true, undefined, "2026-09-24")).toEqual({ value: "2026-09-24", fromRecord: false });
    expect(whtCertificateIssueDate(false, "2026-09-20", "2026-09-24")).toEqual({ value: "2026-09-24", fromRecord: false });
  });
  it("sends the screen values with the report amounts when not locked", () => {
    const open = whtCertificateFigures(false, values, row());
    expect(open.form).toBe("7");
    expect(open.incomes[0]).toMatchObject({ type: "3_tres", paiddate: "2026-09-10", amount: "20000.00", tax: "200.00" });
  });
});
