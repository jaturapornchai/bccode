/**
 * Thai e-Tax Invoice & e-Receipt XML Generator
 * Standards: ETDA (Electronic Transactions Development Agency) & Thai Revenue Department
 * Standard: TaxInvoice_CrossIndustryInvoice:2
 */

export type ETaxTypeCode =
  | "388" // ใบกำกับภาษี (Tax Invoice)
  | "380" // ใบแจ้งหนี้ (Commercial Invoice)
  | "80"  // ใบเพิ่มหนี้ (Debit Note)
  | "81"  // ใบลดหนี้ (Credit Note)
  | "T03"; // ใบเสร็จรับเงิน (Receipt)

export type ETaxParty = {
  taxId: string; // 13 digits
  branchId: string; // 5 digits, e.g. "00000" for head office
  name: string;
  addressLine?: string;
  province?: string;
  postalCode?: string;
  phoneNumber?: string;
  email?: string;
};

export type ETaxLineItem = {
  sequence: number;
  description: string;
  quantity: number;
  unitPrice: number;
  lineTotal: number;
};

export type ETaxInvoiceData = {
  invoiceNumber: string;
  issueDateTime: string; // YYYY-MM-DDTHH:mm:ss
  typeCode: ETaxTypeCode;
  seller: ETaxParty;
  buyer: ETaxParty;
  items: ETaxLineItem[];
  subtotal: number;
  vatRate: number; // 7.00
  vatAmount: number;
  grandTotal: number;
  remark?: string;
};

/**
 * Validates Thai 13-digit National ID / Tax ID checksum according to official formula.
 */
export function validateThaiTaxId13(taxId: string): boolean {
  const clean = taxId.replace(/\D/g, "");
  if (clean.length !== 13) return false;

  let sum = 0;
  for (let i = 0; i < 12; i++) {
    sum += parseInt(clean[i], 10) * (13 - i);
  }

  const checkDigit = (11 - (sum % 11)) % 10;
  return checkDigit === parseInt(clean[12], 10);
}

/**
 * Validates e-Tax Invoice Data completeness and arithmetic.
 */
export function validateETaxInvoice(data: ETaxInvoiceData): {
  isValid: boolean;
  errors: string[];
} {
  const errors: string[] = [];

  if (!data.invoiceNumber || !data.invoiceNumber.trim()) {
    errors.push("เลขที่เอกสาร (invoiceNumber) ต้องไม่ว่างเปล่า");
  }

  if (!data.seller.taxId || !validateThaiTaxId13(data.seller.taxId)) {
    errors.push("เลขประจำตัวผู้เสียภาษีของผู้ขายไม่ถูกต้อง (ต้องเป็นเลข 13 หลักและผ่าน Checksum)");
  }

  if (!data.seller.name || !data.seller.name.trim()) {
    errors.push("ชื่อผู้ขายต้องไม่ว่างเปล่า");
  }

  if (!data.buyer.taxId || !validateThaiTaxId13(data.buyer.taxId)) {
    errors.push("เลขประจำตัวผู้เสียภาษีของผู้ซื้อไม่ถูกต้อง (ต้องเป็นเลข 13 หลักและผ่าน Checksum)");
  }

  if (!data.buyer.name || !data.buyer.name.trim()) {
    errors.push("ชื่อผู้ซื้อต้องไม่ว่างเปล่า");
  }

  if (!data.items || data.items.length === 0) {
    errors.push("ต้องมีรายการสินค้า/บริการอย่างน้อย 1 รายการ");
  }

  // Arithmetic verification
  const calcSubtotal = Math.round(data.items.reduce((s, i) => s + i.lineTotal, 0) * 100) / 100;
  if (Math.abs(calcSubtotal - data.subtotal) > 0.05) {
    errors.push(`ยอดรวมก่อนภาษี (${data.subtotal}) ไม่ตรงกับผลรวมรายการ (${calcSubtotal})`);
  }

  const calcVat = Math.round((data.subtotal * data.vatRate) / 100 * 100) / 100;
  if (Math.abs(calcVat - data.vatAmount) > 0.05) {
    errors.push(`ยอดภาษีมูลค่าเพิ่ม (${data.vatAmount}) ไม่ตรงกับการคำนวณ (${calcVat})`);
  }

  const calcGrand = Math.round((data.subtotal + data.vatAmount) * 100) / 100;
  if (Math.abs(calcGrand - data.grandTotal) > 0.05) {
    errors.push(`ยอดรวมทั้งสิ้น (${data.grandTotal}) ไม่ตรงกับยอดก่อนภาษี + VAT (${calcGrand})`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

/**
 * Escapes special XML characters.
 */
function escapeXml(unsafe: string): string {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

/**
 * Formats 5-digit branch code (e.g. "00000" for head office).
 */
function formatBranchId(branch: string): string {
  const clean = branch.replace(/\D/g, "");
  return clean.padStart(5, "0").slice(0, 5);
}

/**
 * Generates official ETDA-compliant XML for Thai e-Tax Invoice.
 */
export function generateETaxInvoiceXml(data: ETaxInvoiceData): string {
  const typeCode = data.typeCode || "388";
  const typeNameTh =
    typeCode === "388"
      ? "ใบกำกับภาษี"
      : typeCode === "80"
        ? "ใบเพิ่มหนี้"
        : typeCode === "81"
          ? "ใบลดหนี้"
          : "ใบแจ้งหนี้";

  const sellerBranch = formatBranchId(data.seller.branchId || "00000");
  const buyerBranch = formatBranchId(data.buyer.branchId || "00000");

  let linesXml = "";
  for (const item of data.items) {
    linesXml += `
    <ram:IncludedSupplyChainTradeLineItem>
      <ram:AssociatedDocumentLineDocument>
        <ram:LineID>${item.sequence}</ram:LineID>
      </ram:AssociatedDocumentLineDocument>
      <ram:SpecifiedTradeProduct>
        <ram:Name>${escapeXml(item.description)}</ram:Name>
      </ram:SpecifiedTradeProduct>
      <ram:SpecifiedLineTradeAgreement>
        <ram:GrossPriceProductTradePrice>
          <ram:ChargeAmount>${item.unitPrice.toFixed(2)}</ram:ChargeAmount>
        </ram:GrossPriceProductTradePrice>
      </ram:SpecifiedLineTradeAgreement>
      <ram:SpecifiedLineTradeDelivery>
        <ram:BilledQuantity>${item.quantity}</ram:BilledQuantity>
      </ram:SpecifiedLineTradeDelivery>
      <ram:SpecifiedLineTradeSettlement>
        <ram:SpecifiedTradeSettlementLineMonetarySummation>
          <ram:LineTotalAmount>${item.lineTotal.toFixed(2)}</ram:LineTotalAmount>
        </ram:SpecifiedTradeSettlementLineMonetarySummation>
      </ram:SpecifiedLineTradeSettlement>
    </ram:IncludedSupplyChainTradeLineItem>`;
  }

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<rsm:TaxInvoice_CrossIndustryInvoice xmlns:rsm="urn:etda:uncefact:data:standard:TaxInvoice_CrossIndustryInvoice:2" xmlns:ram="urn:etda:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:2">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter>
      <ram:ID schemeAgencyID="ETDA" schemeVersionID="v2.0">ER3-2560</ram:ID>
    </ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument>
    <ram:ID>${escapeXml(data.invoiceNumber)}</ram:ID>
    <ram:Name>${escapeXml(typeNameTh)}</ram:Name>
    <ram:TypeCode>${typeCode}</ram:TypeCode>
    <ram:IssueDateTime>${escapeXml(data.issueDateTime)}</ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:SellerTradeParty>
        <ram:Name>${escapeXml(data.seller.name)}</ram:Name>
        <ram:SpecifiedTaxRegistration>
          <ram:ID schemeAgencyName="RD" schemeID="TXID">${data.seller.taxId}</ram:ID>
        </ram:SpecifiedTaxRegistration>
        <ram:DefinedTradeContact>
          <ram:PersonName>${sellerBranch === "00000" ? "สำนักงานใหญ่" : `สาขาที่ ${sellerBranch}`}</ram:PersonName>
        </ram:DefinedTradeContact>
      </ram:SellerTradeParty>
      <ram:BuyerTradeParty>
        <ram:Name>${escapeXml(data.buyer.name)}</ram:Name>
        <ram:SpecifiedTaxRegistration>
          <ram:ID schemeAgencyName="RD" schemeID="TXID">${data.buyer.taxId}</ram:ID>
        </ram:SpecifiedTaxRegistration>
        <ram:DefinedTradeContact>
          <ram:PersonName>${buyerBranch === "00000" ? "สำนักงานใหญ่" : `สาขาที่ ${buyerBranch}`}</ram:PersonName>
        </ram:DefinedTradeContact>
      </ram:BuyerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeDelivery/>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>THB</ram:InvoiceCurrencyCode>
      <ram:ApplicableTradeTax>
        <ram:TypeCode>VAT</ram:TypeCode>
        <ram:CalculatedRate>${data.vatRate.toFixed(2)}</ram:CalculatedRate>
        <ram:BasisAmount>${data.subtotal.toFixed(2)}</ram:BasisAmount>
        <ram:CalculatedAmount>${data.vatAmount.toFixed(2)}</ram:CalculatedAmount>
      </ram:ApplicableTradeTax>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>${data.subtotal.toFixed(2)}</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>${data.subtotal.toFixed(2)}</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>${data.vatAmount.toFixed(2)}</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>${data.grandTotal.toFixed(2)}</ram:GrandTotalAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
    ${linesXml.trim()}
  </rsm:SupplyChainTradeTransaction>
</rsm:TaxInvoice_CrossIndustryInvoice>`.trim();

  return xml;
}
