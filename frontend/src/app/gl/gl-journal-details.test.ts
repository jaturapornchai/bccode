import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { GLJournalDetailsPanel, VatClaimHint, clearWithholdingParty, vatClaimPeriodTh } from "./gl-journal-details";

// ภาษีซื้อที่ใช้สิทธิช้า: ข้อความบนใบกำกับต้องเป็นเดือนไทยเสมอ และทุก {period} ต้องถูกแทน (adversarial review 2026-09-24)
describe("VatClaimHint late-claim note", () => {
  const row = { tax_type: 1, document_type: 1, claim_status: 1, tax_invoice_date: "2026-01-10", tax_period_year: 2026, tax_period_month: 3 };
  const english: Record<string, string> = {
    month_march: "March",
    vat_ui_claim_late_hint: "Input VAT claimed {months} month(s) after the invoice month — this tax invoice must carry the Thai note “ถือเป็นภาษีซื้อในเดือนภาษี {period_th}” (input tax for tax month {period}; Director-General's VAT Notification No. 4, clause 2).",
  };
  it("replaces every placeholder and keeps the Thai month in the invoice note", () => {
    const html = renderToStaticMarkup(createElement(VatClaimHint, { row, tr: (key: string, fallback: string) => english[key] ?? fallback }));
    expect(html).not.toContain("{period");
    expect(html).toContain("“ถือเป็นภาษีซื้อในเดือนภาษี มีนาคม 2569”");
    expect(html).toContain("input tax for tax month March 2569");
    expect(html).toContain("claimed 2 month(s)");
  });
  it("uses the Thai fallback note when no dictionary row exists", () => {
    const html = renderToStaticMarkup(createElement(VatClaimHint, { row, tr: (_key: string, fallback: string) => fallback }));
    expect(html).toContain("“ถือเป็นภาษีซื้อในเดือนภาษี มีนาคม 2569”");
    expect(html).not.toContain("{period");
  });
  // ข้อความในใบกำกับ = ชื่อเดือนไทย + ปี พ.ศ. ตัวเลขล้วน (ไม่มีคำว่า "พ.ศ." หรือเดือนอังกฤษ) ทุกเดือน
  it("builds the Thai month name + Buddhist year for every month", () => {
    expect(vatClaimPeriodTh(2026, 9)).toBe("กันยายน 2569");
    expect(vatClaimPeriodTh(2026, 12)).toBe("ธันวาคม 2569");
    expect(vatClaimPeriodTh(2027, 1)).toBe("มกราคม 2570");
    for (let month = 1; month <= 12; month++) expect(vatClaimPeriodTh(2026, month)).toMatch(/^[ก-๎]+ 2569$/);
  });
});

// เปลี่ยนคู่ค้า/ทิศทางของแถวภาษีหัก: snapshot ฝั่งนั้นต้องถูกล้าง (undefined = ไม่ส่งไป backend) ให้ backend เติมจากทะเบียนคู่ค้าใหม่
describe("clearWithholdingParty", () => {
  it("drops every snapshot field of the given side from the saved JSON", () => {
    const row = { partner_code: "S002", payee_name: "บริษัท ขนส่งไทยเร็ว จำกัด", payee_tax_id: "0105558012349", payee_branch_no: "00000", payee_address: "กรุงเทพฯ", payer_name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" };
    const saved = JSON.parse(JSON.stringify({ ...row, ...clearWithholdingParty("payee") }));
    expect(saved).toEqual({ partner_code: "S002", payer_name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" });
  });
});

// UAT S3 2026-09-24: เลขผู้เสียภาษีหลักตรวจสอบผิดขึ้นแค่แถบแจ้งด้านบน — ต้องไฮไลต์ที่ช่อง และจอหาช่องได้จาก data-detail-field
describe("GLJournalDetailsPanel partner rows", () => {
  const render = (taxId: string) => renderToStaticMarkup(createElement(GLJournalDetailsPanel, {
    value: { partners: [{ partner_code: "ZUAT-IND", title_name: "นาย", name_th: "สมชาย รับเหมาดี", tax_id: taxId, is_customer: false, is_supplier: true, is_active: true }] },
    onChange: () => {}, lines: [], accounts: [], date: "2026-10-20", branch: "00000", editable: true, posted: false,
  }));
  it("flags a 13-digit tax id whose check digit is wrong right at the field", () => {
    const html = render("0105561001238");
    expect(html).toContain('data-detail-field="partners.1.tax_id"');
    expect(html).toContain("หลักสุดท้าย (หลักตรวจสอบ) ไม่ตรงกับ 12 หลักแรก");
    expect(html).toMatch(/aria-label="เลขผู้เสียภาษี"[^>]*aria-invalid="true"/);
  });
  it("accepts a valid tax id and names the add button without repeating เพิ่ม", () => {
    const html = render("0105561001239");
    expect(html).not.toContain("หลักตรวจสอบ) ไม่ตรง");
    expect(html).not.toContain("เพิ่มเพิ่ม");
    expect(html).toContain("เพิ่มคู่ค้าภายในห้องบัญชี");
    expect(html).toContain("ใช้งานอยู่");
  });
});
