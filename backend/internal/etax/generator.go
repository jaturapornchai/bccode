package etax

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

const (
	xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	schemeAgencyID = "ETDA"
	schemeVersionID = "v2.0"
	guidelineID = "ER3-2560"
)

// GenerateXML แปลง Document เป็น e-Tax Invoice XML ตามมาตรฐาน มธพ. 3-2560
func GenerateXML(doc Document) ([]byte, error) {
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("document validation failed: %w", err)
	}

	if doc.CurrencyCode == "" {
		doc.CurrencyCode = "THB"
	}
	if doc.IssueDateTime.IsZero() {
		doc.IssueDateTime = time.Now()
	}

	var buf bytes.Buffer
	buf.WriteString(xmlHeader)

	// Root Element
	buf.WriteString(`<rsm:CrossIndustryInvoice xmlns:rsm="urn:etda:uncefact:data:standard:CrossIndustryInvoice:100" `)
	buf.WriteString(`xmlns:ram="urn:etda:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100" `)
	buf.WriteString(`xmlns:qdt="urn:etda:uncefact:data:standard:QualifiedDataType:100" `)
	buf.WriteString(`xmlns:udt="urn:etda:uncefact:data:standard:UnqualifiedDataType:100">` + "\n")

	// 1. ExchangedDocumentContext
	buf.WriteString("  <rsm:ExchangedDocumentContext>\n")
	buf.WriteString("    <ram:GuidelineSpecifiedDocumentContextParameter>\n")
	fmt.Fprintf(&buf, "      <ram:ID schemeAgencyID=\"%s\" schemeVersionID=\"%s\">%s</ram:ID>\n",
		schemeAgencyID, schemeVersionID, guidelineID)
	buf.WriteString("    </ram:GuidelineSpecifiedDocumentContextParameter>\n")
	buf.WriteString("  </rsm:ExchangedDocumentContext>\n")

	// 2. ExchangedDocument
	buf.WriteString("  <rsm:ExchangedDocument>\n")
	fmt.Fprintf(&buf, "    <ram:ID>%s</ram:ID>\n", escapeXML(doc.DocNo))
	fmt.Fprintf(&buf, "    <ram:Name>%s</ram:Name>\n", escapeXML(doc.TypeName))
	fmt.Fprintf(&buf, "    <ram:TypeCode>%s</ram:TypeCode>\n", doc.TypeCode)
	fmt.Fprintf(&buf, "    <ram:IssueDateTime>%s</ram:IssueDateTime>\n", doc.IssueDateTime.Format("2006-01-02T15:04:05.000"))

	if doc.PurposeCode != "" {
		fmt.Fprintf(&buf, "    <ram:PurposeCode>%s</ram:PurposeCode>\n", escapeXML(doc.PurposeCode))
	}
	if doc.PurposeReason != "" {
		fmt.Fprintf(&buf, "    <ram:Purpose>%s</ram:Purpose>\n", escapeXML(doc.PurposeReason))
	}
	buf.WriteString("  </rsm:ExchangedDocument>\n")

	// 3. SupplyChainTradeTransaction
	buf.WriteString("  <rsm:SupplyChainTradeTransaction>\n")

	// 3.1 Header Trade Agreement (Seller / Buyer)
	buf.WriteString("    <ram:ApplicableHeaderTradeAgreement>\n")

	// Seller
	buf.WriteString("      <ram:SellerTradeParty>\n")
	fmt.Fprintf(&buf, "        <ram:Name>%s</ram:Name>\n", escapeXML(doc.Seller.Name))
	buf.WriteString("        <ram:SpecifiedTaxRegistration>\n")
	fmt.Fprintf(&buf, "          <ram:ID schemeID=\"TXID\">%s%s</ram:ID>\n", doc.Seller.TaxID, doc.Seller.BranchID)
	buf.WriteString("        </ram:SpecifiedTaxRegistration>\n")
	buf.WriteString("        <ram:PostalTradeAddress>\n")
	fmt.Fprintf(&buf, "          <ram:PostcodeCode>%s</ram:PostcodeCode>\n", escapeXML(doc.Seller.Address.Postcode))
	fmt.Fprintf(&buf, "          <ram:LineOne>%s</ram:LineOne>\n", escapeXML(doc.Seller.Address.LineOne))
	fmt.Fprintf(&buf, "          <ram:CountryID>%s</ram:CountryID>\n", escapeXML(doc.Seller.Address.CountryCode))
	buf.WriteString("        </ram:PostalTradeAddress>\n")
	buf.WriteString("      </ram:SellerTradeParty>\n")

	// Buyer
	buf.WriteString("      <ram:BuyerTradeParty>\n")
	fmt.Fprintf(&buf, "        <ram:Name>%s</ram:Name>\n", escapeXML(doc.Buyer.Name))
	buf.WriteString("        <ram:SpecifiedTaxRegistration>\n")
	fmt.Fprintf(&buf, "          <ram:ID schemeID=\"TXID\">%s%s</ram:ID>\n", doc.Buyer.TaxID, doc.Buyer.BranchID)
	buf.WriteString("        </ram:SpecifiedTaxRegistration>\n")
	buf.WriteString("        <ram:PostalTradeAddress>\n")
	fmt.Fprintf(&buf, "          <ram:PostcodeCode>%s</ram:PostcodeCode>\n", escapeXML(doc.Buyer.Address.Postcode))
	fmt.Fprintf(&buf, "          <ram:LineOne>%s</ram:LineOne>\n", escapeXML(doc.Buyer.Address.LineOne))
	fmt.Fprintf(&buf, "          <ram:CountryID>%s</ram:CountryID>\n", escapeXML(doc.Buyer.Address.CountryCode))
	buf.WriteString("        </ram:PostalTradeAddress>\n")
	buf.WriteString("      </ram:BuyerTradeParty>\n")

	// Additional Referenced Document (กรณีใบลดหนี้/ใบเพิ่มหนี้)
	if doc.OriginalDocNo != "" {
		buf.WriteString("      <ram:AdditionalReferencedDocument>\n")
		fmt.Fprintf(&buf, "        <ram:IssuerAssignedID>%s</ram:IssuerAssignedID>\n", escapeXML(doc.OriginalDocNo))
		if doc.OriginalDocDate != "" {
			fmt.Fprintf(&buf, "        <ram:IssueDateTime>%s</ram:IssueDateTime>\n", escapeXML(doc.OriginalDocDate))
		}
		fmt.Fprintf(&buf, "        <ram:ReferenceTypeCode>%s</ram:ReferenceTypeCode>\n", "1")
		buf.WriteString("      </ram:AdditionalReferencedDocument>\n")
	}

	buf.WriteString("    </ram:ApplicableHeaderTradeAgreement>\n")

	// 3.2 Header Trade Delivery
	buf.WriteString("    <ram:ApplicableHeaderTradeDelivery>\n")
	buf.WriteString("      <ram:ShipToTradeParty>\n")
	fmt.Fprintf(&buf, "        <ram:Name>%s</ram:Name>\n", escapeXML(doc.Buyer.Name))
	buf.WriteString("      </ram:ShipToTradeParty>\n")
	buf.WriteString("    </ram:ApplicableHeaderTradeDelivery>\n")

	// 3.3 Header Trade Settlement
	buf.WriteString("    <ram:ApplicableHeaderTradeSettlement>\n")
	fmt.Fprintf(&buf, "      <ram:InvoiceCurrencyCode>%s</ram:InvoiceCurrencyCode>\n", doc.CurrencyCode)

	// ApplicableTradeTax
	buf.WriteString("      <ram:ApplicableTradeTax>\n")
	buf.WriteString("        <ram:TypeCode>VAT</ram:TypeCode>\n")
	fmt.Fprintf(&buf, "        <ram:CalculatedRate>%s</ram:CalculatedRate>\n", doc.TaxRate.StringFixed(2))
	fmt.Fprintf(&buf, "        <ram:BasisAmount currencyID=\"%s\">%s</ram:BasisAmount>\n",
		doc.CurrencyCode, doc.TaxBasisTotal.StringFixed(2))
	fmt.Fprintf(&buf, "        <ram:CalculatedAmount currencyID=\"%s\">%s</ram:CalculatedAmount>\n",
		doc.CurrencyCode, doc.TaxTotal.StringFixed(2))
	buf.WriteString("      </ram:ApplicableTradeTax>\n")

	// Monetary Summation
	buf.WriteString("      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>\n")
	fmt.Fprintf(&buf, "        <ram:LineTotalAmount currencyID=\"%s\">%s</ram:LineTotalAmount>\n",
		doc.CurrencyCode, doc.LineTotalAmount.StringFixed(2))
	fmt.Fprintf(&buf, "        <ram:TaxBasisTotalAmount currencyID=\"%s\">%s</ram:TaxBasisTotalAmount>\n",
		doc.CurrencyCode, doc.TaxBasisTotal.StringFixed(2))
	fmt.Fprintf(&buf, "        <ram:TaxTotalAmount currencyID=\"%s\">%s</ram:TaxTotalAmount>\n",
		doc.CurrencyCode, doc.TaxTotal.StringFixed(2))
	fmt.Fprintf(&buf, "        <ram:GrandTotalAmount currencyID=\"%s\">%s</ram:GrandTotalAmount>\n",
		doc.CurrencyCode, doc.GrandTotal.StringFixed(2))
	buf.WriteString("      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>\n")

	buf.WriteString("    </ram:ApplicableHeaderTradeSettlement>\n")

	// 3.4 Line Items
	for _, item := range doc.Items {
		buf.WriteString("    <ram:IncludedSupplyChainTradeLineItem>\n")
		buf.WriteString("      <ram:AssociatedDocumentLineDocument>\n")
		fmt.Fprintf(&buf, "        <ram:LineID>%d</ram:LineID>\n", item.LineID)
		buf.WriteString("      </ram:AssociatedDocumentLineDocument>\n")

		// SpecifiedTradeProduct
		buf.WriteString("      <ram:SpecifiedTradeProduct>\n")
		fmt.Fprintf(&buf, "        <ram:ID>%s</ram:ID>\n", escapeXML(item.ItemCode))
		fmt.Fprintf(&buf, "        <ram:Name>%s</ram:Name>\n", escapeXML(item.ItemName))
		buf.WriteString("      </ram:SpecifiedTradeProduct>\n")

		// SpecifiedLineTradeAgreement
		buf.WriteString("      <ram:SpecifiedLineTradeAgreement>\n")
		buf.WriteString("        <ram:GrossPriceProductTradePrice>\n")
		fmt.Fprintf(&buf, "          <ram:ChargeAmount currencyID=\"%s\">%s</ram:ChargeAmount>\n",
			doc.CurrencyCode, item.UnitPrice.StringFixed(2))
		buf.WriteString("        </ram:GrossPriceProductTradePrice>\n")
		buf.WriteString("      </ram:SpecifiedLineTradeAgreement>\n")

		// SpecifiedLineTradeDelivery
		buf.WriteString("      <ram:SpecifiedLineTradeDelivery>\n")
		unitCode := item.UnitCode
		if unitCode == "" {
			unitCode = "C62" // ชิ้น / one
		}
		fmt.Fprintf(&buf, "        <ram:BilledQuantity unitCode=\"%s\">%s</ram:BilledQuantity>\n",
			escapeXML(unitCode), item.Quantity.StringFixed(2))
		buf.WriteString("      </ram:SpecifiedLineTradeDelivery>\n")

		// SpecifiedLineTradeSettlement
		buf.WriteString("      <ram:SpecifiedLineTradeSettlement>\n")
		buf.WriteString("        <ram:SpecifiedTradeSettlementLineMonetarySummation>\n")
		fmt.Fprintf(&buf, "          <ram:LineTotalAmount currencyID=\"%s\">%s</ram:LineTotalAmount>\n",
			doc.CurrencyCode, item.NetAmount.StringFixed(2))
		buf.WriteString("        </ram:SpecifiedTradeSettlementLineMonetarySummation>\n")
		buf.WriteString("      </ram:SpecifiedLineTradeSettlement>\n")

		buf.WriteString("    </ram:IncludedSupplyChainTradeLineItem>\n")
	}

	buf.WriteString("  </rsm:SupplyChainTradeTransaction>\n")
	buf.WriteString("</rsm:CrossIndustryInvoice>\n")

	return buf.Bytes(), nil
}

func escapeXML(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
