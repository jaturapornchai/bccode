package etax

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func sampleInvoice() Document {
	return Document{
		DocNo:         "INV6910-0012",
		TypeCode:      TypeInvoice,
		TypeName:      "ใบกำกับภาษี",
		IssueDateTime: time.Date(2026, 10, 10, 14, 30, 0, 0, time.UTC),
		CurrencyCode:  "THB",
		Seller: Party{
			ID:       "V001",
			Name:     "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด",
			TaxID:    "0105560012345",
			BranchID: "00000",
			Address: Address{
				Postcode:    "10110",
				LineOne:     "123/45 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร",
				CountryCode: "TH",
			},
		},
		Buyer: Party{
			ID:       "C008",
			Name:     "บริษัท สยามพัฒนาการก่อสร้าง จำกัด (มหาชน)",
			TaxID:    "0107537000123",
			BranchID: "00001",
			Address: Address{
				Postcode:    "10330",
				LineOne:     "888 ถนนพระราม 9 แขวงห้วยขวาง เขตห้วยขวาง กรุงเทพมหานคร",
				CountryCode: "TH",
			},
		},
		Items: []LineItem{
			{
				LineID:    1,
				ItemCode:  "CEMENT-P01",
				ItemName:  "ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง 50 กก.",
				Quantity:  decimal.NewFromInt(1000),
				UnitCode:  "C62",
				UnitPrice: decimal.NewFromFloat(150.00),
				Amount:    decimal.NewFromFloat(150000.00),
				Discount:  decimal.Zero,
				NetAmount: decimal.NewFromFloat(150000.00),
				TaxRate:   decimal.NewFromFloat(7.00),
				TaxAmount: decimal.NewFromFloat(10500.00),
			},
		},
		LineTotalAmount: decimal.NewFromFloat(150000.00),
		DiscountTotal:   decimal.Zero,
		TaxBasisTotal:   decimal.NewFromFloat(150000.00),
		TaxRate:         decimal.NewFromFloat(7.00),
		TaxTotal:        decimal.NewFromFloat(10500.00),
		GrandTotal:      decimal.NewFromFloat(160500.00),
	}
}

func TestGenerateXML_Invoice(t *testing.T) {
	doc := sampleInvoice()
	xmlBytes, err := GenerateXML(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlBytes)

	// ตรวจสอบ Header และ Namespace ตาม มธพ. 3-2560
	if !strings.Contains(xmlStr, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("missing xml declaration")
	}
	if !strings.Contains(xmlStr, `<rsm:CrossIndustryInvoice`) {
		t.Errorf("missing root CrossIndustryInvoice element")
	}
	if !strings.Contains(xmlStr, `ER3-2560`) {
		t.Errorf("missing guideline ER3-2560")
	}

	// ตรวจสอบข้อมูลเอกสาร
	if !strings.Contains(xmlStr, `<ram:ID>INV6910-0012</ram:ID>`) {
		t.Errorf("missing doc no")
	}
	if !strings.Contains(xmlStr, `<ram:TypeCode>T01</ram:TypeCode>`) {
		t.Errorf("missing type code T01")
	}

	// ตรวจสอบ Tax ID ผู้ขายและผู้ซื้อ พร้อมรหัสสาขา 5 หลัก
	if !strings.Contains(xmlStr, `010556001234500000`) {
		t.Errorf("missing seller tax ID and branch 00000")
	}
	if !strings.Contains(xmlStr, `010753700012300001`) {
		t.Errorf("missing buyer tax ID and branch 00001")
	}

	// ตรวจสอบตัวเลขทางการเงินและภาษี 7%
	if !strings.Contains(xmlStr, `<ram:GrandTotalAmount currencyID="THB">160500.00</ram:GrandTotalAmount>`) {
		t.Errorf("missing correct grand total 160500.00")
	}
	if !strings.Contains(xmlStr, `<ram:TaxTotalAmount currencyID="THB">10500.00</ram:TaxTotalAmount>`) {
		t.Errorf("missing correct tax total 10500.00")
	}
}

func TestGenerateXML_CreditNote(t *testing.T) {
	doc := sampleInvoice()
	doc.DocNo = "SR6910-0004"
	doc.TypeCode = TypeCreditNote
	doc.TypeName = "ใบลดหนี้"
	doc.PurposeCode = "CDNG05"
	doc.PurposeReason = "รับคืนสินค้าเนื่องจากสินค้าผิดคุณลักษณะ"
	doc.OriginalDocNo = "INV6910-0012"
	doc.OriginalDocDate = "2026-10-10"

	xmlBytes, err := GenerateXML(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlBytes)

	if !strings.Contains(xmlStr, `<ram:TypeCode>81</ram:TypeCode>`) {
		t.Errorf("missing type code 81 for credit note")
	}
	if !strings.Contains(xmlStr, `<ram:PurposeCode>CDNG05</ram:PurposeCode>`) {
		t.Errorf("missing PurposeCode CDNG05")
	}
	if !strings.Contains(xmlStr, `<ram:IssuerAssignedID>INV6910-0012</ram:IssuerAssignedID>`) {
		t.Errorf("missing original invoice reference")
	}
}

func TestGenerateXML_ValidationErrors(t *testing.T) {
	// Tax ID ไม่ครบ 13 หลัก
	doc := sampleInvoice()
	doc.Seller.TaxID = "12345"
	if err := doc.Validate(); err == nil {
		t.Errorf("expected error for invalid seller tax ID, got nil")
	}

	// ใบลดหนี้ไม่มี PurposeCode
	doc = sampleInvoice()
	doc.TypeCode = TypeCreditNote
	if err := doc.Validate(); err == nil {
		t.Errorf("expected error for credit note without purpose_code, got nil")
	}
}
