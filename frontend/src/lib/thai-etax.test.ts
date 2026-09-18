import { describe, it, expect } from "vitest";
import {
  validateThaiTaxId13,
  validateETaxInvoice,
  generateETaxInvoiceXml,
  type ETaxInvoiceData,
} from "./thai-etax";

describe("Thai e-Tax Invoice & e-Receipt XML Engine (Group C)", () => {
  it("validates 13-digit Thai Tax ID checksum accurately", () => {
    // Valid Thai ID / Tax ID sample: 0105558000121 (Real DBD registered format)
    // Formula verification:
    // Let's test a known valid ID: 1100701831411 (or mock valid checksum)
    // Let's verify our checksum implementation:
    const validId = "1100701831411"; // checksum verified
    expect(validateThaiTaxId13(validId)).toBe(true);

    // Invalid length or wrong check digit
    expect(validateThaiTaxId13("1100701831412")).toBe(false);
    expect(validateThaiTaxId13("12345")).toBe(false);
  });

  it("validates ETaxInvoiceData arithmetic and required fields", () => {
    const validInvoice: ETaxInvoiceData = {
      invoiceNumber: "INV2026/001",
      issueDateTime: "2026-09-18T10:00:00",
      typeCode: "388",
      seller: {
        taxId: "1100701831411",
        branchId: "00000",
        name: "บริษัท บีซี ไอที จำกัด",
      },
      buyer: {
        taxId: "1100701831411",
        branchId: "00001",
        name: "บริษัท สยาม การค้า จำกัด",
      },
      items: [
        {
          sequence: 1,
          description: "ค่าบริการซอฟต์แวร์ ERP รายปี",
          quantity: 1,
          unitPrice: 10000,
          lineTotal: 10000,
        },
      ],
      subtotal: 10000,
      vatRate: 7,
      vatAmount: 700,
      grandTotal: 10700,
    };

    const result = validateETaxInvoice(validInvoice);
    expect(result.isValid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it("detects arithmetic mismatch in subtotal or VAT", () => {
    const invalidInvoice: ETaxInvoiceData = {
      invoiceNumber: "INV-ERR",
      issueDateTime: "2026-09-18T10:00:00",
      typeCode: "388",
      seller: { taxId: "1100701831411", branchId: "00000", name: "Seller" },
      buyer: { taxId: "1100701831411", branchId: "00000", name: "Buyer" },
      items: [{ sequence: 1, description: "Item 1", quantity: 1, unitPrice: 1000, lineTotal: 1000 }],
      subtotal: 1000,
      vatRate: 7,
      vatAmount: 999, // incorrect VAT
      grandTotal: 1070,
    };

    const result = validateETaxInvoice(invalidInvoice);
    expect(result.isValid).toBe(false);
    expect(result.errors.some((e) => e.includes("ยอดภาษีมูลค่าเพิ่ม"))).toBe(true);
  });

  it("generates valid ETDA standard e-Tax Invoice XML", () => {
    const invoice: ETaxInvoiceData = {
      invoiceNumber: "INV2026-0901",
      issueDateTime: "2026-09-18T14:30:00",
      typeCode: "388",
      seller: {
        taxId: "1100701831411",
        branchId: "00000",
        name: "บริษัท ทดสอบ ผู้ขาย จำกัด & Partners",
      },
      buyer: {
        taxId: "1100701831411",
        branchId: "00002",
        name: "บริษัท ทดสอบ ผู้ซื้อ จำกัด",
      },
      items: [
        {
          sequence: 1,
          description: "สินค้า ก & บริการ ข",
          quantity: 2,
          unitPrice: 500,
          lineTotal: 1000,
        },
      ],
      subtotal: 1000,
      vatRate: 7,
      vatAmount: 70,
      grandTotal: 1070,
    };

    const xml = generateETaxInvoiceXml(invoice);

    expect(xml).toContain('<?xml version="1.0" encoding="UTF-8"?>');
    expect(xml).toContain("TaxInvoice_CrossIndustryInvoice");
    expect(xml).toContain("<ram:ID>INV2026-0901</ram:ID>");
    expect(xml).toContain("<ram:TypeCode>388</ram:TypeCode>");
    expect(xml).toContain("บริษัท ทดสอบ ผู้ขาย จำกัด &amp; Partners"); // Escaped XML
    expect(xml).toContain("<ram:ChargeAmount>500.00</ram:ChargeAmount>");
    expect(xml).toContain("<ram:GrandTotalAmount>1070.00</ram:GrandTotalAmount>");
    expect(xml).toContain("สำนักงานใหญ่");
    expect(xml).toContain("สาขาที่ 00002");
  });
});
