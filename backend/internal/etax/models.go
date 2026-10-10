package etax

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/shopspring/decimal"
)

// DocumentTypeCode รหัสประเภทเอกสารตามมาตรฐาน มธพ. 3-2560 (ETDA / กรมสรรพากร)
type DocumentTypeCode string

const (
	TypeInvoice           DocumentTypeCode = "T01" // ใบกำกับภาษี (Tax Invoice)
	TypeInvoiceTaxInvoice DocumentTypeCode = "T02" // ใบแจ้งหนี้/ใบกำกับภาษี (Invoice / Tax Invoice)
	TypeReceiptTaxInvoice DocumentTypeCode = "T03" // ใบเสร็จรับเงิน/ใบกำกับภาษี (Receipt / Tax Invoice)
	TypeReceipt           DocumentTypeCode = "T04" // ใบเสร็จรับเงิน (Receipt)
	TypeDebitNote         DocumentTypeCode = "80"  // ใบเพิ่มหนี้ (Debit Note)
	TypeCreditNote        DocumentTypeCode = "81"  // ใบลดหนี้ (Credit Note)
)

var (
	taxIDRegex    = regexp.MustCompile(`^[0-9]{13}$`)
	branchIDRegex = regexp.MustCompile(`^[0-9]{5}$`)
)

// Address ที่อยู่สำหรับเอกสาร e-Tax Invoice
type Address struct {
	BuildingName string `json:"building_name,omitempty"`
	StreetName   string `json:"street_name,omitempty"`
	CitySubDiv   string `json:"city_sub_division,omitempty"` // ตำบล/แขวง
	CityName     string `json:"city_name,omitempty"`         // อำเภอ/เขต
	CountrySub   string `json:"country_sub_division,omitempty"` // จังหวัด
	Postcode     string `json:"postcode,omitempty"`
	CountryCode  string `json:"country_code"`                // TH
	LineOne      string `json:"line_one,omitempty"`
}

// Party ข้อมูลคู่ค้า (ผู้ขาย / ผู้ซื้อ)
type Party struct {
	ID         string  `json:"id"`                    // รหัสคู่ค้า
	Name       string  `json:"name"`                  // ชื่อบริษัท/บุคคล
	TaxID      string  `json:"tax_id"`                // เลขประจำตัวผู้เสียภาษี 13 หลัก
	BranchID   string  `json:"branch_id"`             // รหัสสาขา 5 หลัก (00000 = สำนักงานใหญ่)
	Address    Address `json:"address"`
	Email      string  `json:"email,omitempty"`
	Telephone  string  `json:"telephone,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
}

// LineItem รายการสินค้า/บริการ
type LineItem struct {
	LineID      int             `json:"line_id"`
	ItemCode    string          `json:"item_code"`
	ItemName    string          `json:"item_name"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitCode    string          `json:"unit_code"`    // รหัสหน่วยนับ (เช่น C62, HUR)
	UnitPrice   decimal.Decimal `json:"unit_price"`   // ราคาต่อหน่วย (ไม่รวม VAT)
	Amount      decimal.Decimal `json:"amount"`       // มูลค่าก่อนส่วนลด
	Discount    decimal.Decimal `json:"discount"`     // ส่วนลด
	NetAmount   decimal.Decimal `json:"net_amount"`   // มูลค่าสุทธิบรรทัด
	TaxRate     decimal.Decimal `json:"tax_rate"`     // อัตราภาษี เช่น 7.00
	TaxAmount   decimal.Decimal `json:"tax_amount"`   // จำนวนภาษีของบรรทัด
}

// Document ข้อมูลเอกสาร e-Tax Invoice ครบถ้วน
type Document struct {
	DocNo             string           `json:"doc_no"`
	TypeCode          DocumentTypeCode `json:"type_code"`
	TypeName          string           `json:"type_name"`
	IssueDateTime     time.Time        `json:"issue_date_time"`
	PurposeCode       string           `json:"purpose_code,omitempty"`       // CDNG01-CDNG05, DBNG01-DBNG02
	PurposeReason     string           `json:"purpose_reason,omitempty"`     // ข้อความเหตุผล
	OriginalDocNo     string           `json:"original_doc_no,omitempty"`    // อ้างอิงเอกสารเดิม (กรณีใบลดหนี้/ใบเพิ่มหนี้)
	OriginalDocDate   string           `json:"original_doc_date,omitempty"`  // วันที่เอกสารเดิม
	CurrencyCode      string           `json:"currency_code"`                // THB
	Seller            Party            `json:"seller"`
	Buyer             Party            `json:"buyer"`
	Items             []LineItem       `json:"items"`
	LineTotalAmount   decimal.Decimal  `json:"line_total_amount"`            // รวมมูลค่าสินค้า
	DiscountTotal     decimal.Decimal  `json:"discount_total"`               // ส่วนลดรวม
	TaxBasisTotal     decimal.Decimal  `json:"tax_basis_total"`              // ฐานภาษี
	TaxRate           decimal.Decimal  `json:"tax_rate"`                     // 7.00
	TaxTotal          decimal.Decimal  `json:"tax_total"`                    // ภาษีมูลค่าเพิ่ม
	GrandTotal        decimal.Decimal  `json:"grand_total"`                  // ยอดรวมทั้งสิ้น
	Remark            string           `json:"remark,omitempty"`
}

// Validate ตรวจสอบความถูกต้องของข้อมูลตามข้อกำหนด มธพ. 3-2560
func (d *Document) Validate() error {
	if d.DocNo == "" {
		return errors.New("doc_no is required")
	}
	if d.TypeCode == "" {
		return errors.New("type_code is required")
	}
	switch d.TypeCode {
	case TypeInvoice, TypeInvoiceTaxInvoice, TypeReceiptTaxInvoice, TypeReceipt, TypeDebitNote, TypeCreditNote:
		// Valid
	default:
		return fmt.Errorf("invalid type_code: %s", d.TypeCode)
	}

	if !taxIDRegex.MatchString(d.Seller.TaxID) {
		return fmt.Errorf("seller tax_id must be 13 digits, got '%s'", d.Seller.TaxID)
	}
	if d.Seller.BranchID == "" {
		d.Seller.BranchID = "00000"
	}
	if !branchIDRegex.MatchString(d.Seller.BranchID) {
		return fmt.Errorf("seller branch_id must be 5 digits, got '%s'", d.Seller.BranchID)
	}

	if !taxIDRegex.MatchString(d.Buyer.TaxID) && d.Buyer.TaxID != "0000000000000" {
		return fmt.Errorf("buyer tax_id must be 13 digits, got '%s'", d.Buyer.TaxID)
	}
	if d.Buyer.BranchID == "" {
		d.Buyer.BranchID = "00000"
	}
	if !branchIDRegex.MatchString(d.Buyer.BranchID) {
		return fmt.Errorf("buyer branch_id must be 5 digits, got '%s'", d.Buyer.BranchID)
	}

	if len(d.Items) == 0 {
		return errors.New("document must contain at least one line item")
	}

	// กรณีใบลดหนี้หรือใบเพิ่มหนี้ ต้องมี PurposeCode และ OriginalDocNo
	if d.TypeCode == TypeCreditNote || d.TypeCode == TypeDebitNote {
		if d.PurposeCode == "" {
			return errors.New("purpose_code is required for credit note and debit note")
		}
		if d.OriginalDocNo == "" {
			return errors.New("original_doc_no is required for credit note and debit note")
		}
	}

	return nil
}
