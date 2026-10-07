import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { GLDetailVat } from "@/lib/gl-journal-details";
import { ChampTaxBox } from "./gl-journal-boxes";

const vat = (patch: Partial<GLDetailVat> = {}): GLDetailVat => ({
  id: "V1", tax_type: 2, document_type: 1, tax_invoice_no: "IV6909-001", tax_invoice_date: "2026-09-10", tax_period_year: 2026, tax_period_month: 9,
  partner_name: "", base_amount: "10000.00", zero_rate_amount: "0", exempt_amount: "0", vat_rate: "7", ...patch,
});
const render = (props: Partial<Parameters<typeof ChampTaxBox>[0]>) =>
  renderToStaticMarkup(createElement(ChampTaxBox, { details: undefined, scale: 2, disabled: false, onOpenTaxTable: () => {}, onTaxType: () => {}, onPatch: () => {}, onCalc: () => {}, ...props }));

describe("ChampTaxBox", () => {
  it("shows 'no tax' with the invoice fields locked when the voucher has no tax invoice", () => {
    const html = render({});
    expect(html).toContain('<option value="" selected="">ไม่มีภาษี</option>');
    expect(html).toMatch(/aria-label="เลขที่ใบกำกับภาษี"[^>]*disabled=""/);
    expect(html).toContain("ภาษีแยกนอก");
    expect(html).toContain("ภาษีรวมใน");
  });

  it("edits the single invoice in place and marks every field for the save-check jump", () => {
    const html = render({ details: { vats: [vat()] } });
    expect(html).toContain('<option value="2" selected="">ภาษีขาย</option>');
    expect(html).toMatch(/aria-label="เลขที่ใบกำกับภาษี"[^>]*value="IV6909-001"/);
    for (const field of ["tax_type", "vat_rate", "tax_invoice_no", "tax_invoice_date", "base_amount", "vat_amount", "remark"]) expect(html).toContain(`data-detail-field="vats.1.${field}"`);
    expect(html).not.toContain("กรุณาใส่อัตราภาษี");
  });

  it("asks for the rate on a filled invoice without one (no 7% default)", () => {
    expect(render({ details: { vats: [vat({ vat_rate: "" })] } })).toContain("กรุณาใส่อัตราภาษี (%) — ระบบไม่ใส่ 0% ให้เอง");
  });

  it("sends two or more invoices to the tax table", () => {
    const html = render({ details: { vats: [vat(), vat({ id: "V2" })] } });
    expect(html).toContain("มีใบกำกับ 2 ใบ — แก้ในตารางภาษี F3");
    expect(html).not.toContain("data-detail-field");
  });

  it("is read-only without calc buttons in view mode", () => {
    const html = render({ details: { vats: [vat()] }, disabled: true, onTaxType: undefined, onPatch: undefined, onCalc: undefined });
    expect(html).not.toContain("ภาษีแยกนอก");
    expect(html).toMatch(/<select[^>]*disabled=""/);
  });
});
