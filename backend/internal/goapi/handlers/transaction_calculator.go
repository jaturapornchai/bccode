package handlers

import (
	"math"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ============================================================
// Transaction Calculator - Enhanced with VatCal, 2-level discount,
// deposit-adjusted VAT, multi-currency, and return document support
//
// Ported from Flutter:
//   - bcmerchant/lib/calamount.dart (per-item calculation)
//   - bcmerchant/lib/screens/transaction/utils/transaction_calculator.dart
// ============================================================

// TransactionItem - รายการสินค้าในธุรกรรม
type TransactionItem struct {
	ItemCode string  `json:"itemcode"`
	ItemName string  `json:"item_name"`
	Qty float64 `json:"qty"`
	Price float64 `json:"price"`
	Discount string  `json:"discount"`    // เช่น "10%" หรือ "100"
	VatType int     `json:"vat_type"`     // 0=แยกนอก, 1=รวมใน, 2=ไม่กระทบภาษี
	VatRate float64 `json:"vatrate"`     // อัตราภาษี เช่น 7
	VatCal int     `json:"vatcal"`      // 0=มีภาษี (taxable), 1=ยกเว้นภาษี (exempt)
	UnitCode string  `json:"unitcode"`
	UnitStand float64 `json:"unitstand"`
	UnitDivide float64 `json:"unitdivide"`
	WHCode string  `json:"whcode"`
	ShelfCode string  `json:"shelfcode"`
	CalcFlag int     `json:"calcflag"`
	InquiryType int     `json:"inquirytype,omitempty"`

	// Calculated fields - enhanced (ported from calamount.dart)
	SumAmount float64 `json:"sum_amount,omitempty"`
	DiscountAmount float64 `json:"discountamount,omitempty"`
	PriceExcludeVat float64 `json:"priceexcludevat,omitempty"`
	SumAmountExcludeVat float64 `json:"sumamountexcludevat,omitempty"`
	TotalValueVat float64 `json:"totalvaluevat,omitempty"`
	TotalQty float64 `json:"totalqty,omitempty"`

	// Legacy calculated fields (backward compat)
	NetAmount float64 `json:"netamount,omitempty"`
	TaxAmount float64 `json:"taxamount,omitempty"`
	TotalAmount float64 `json:"total_amount,omitempty"`
}

// PaymentMethod - วิธีการชำระเงิน
type PaymentMethod struct {
	PayCode string  `json:"paycode"`
	PayName string  `json:"payname"`
	Amount float64 `json:"amount"`
	TransferBankCode string  `json:"transfer_bank_code,omitempty"`
	CreditCardType string  `json:"credit_card_type,omitempty"`
	ChequeNo string  `json:"cheque_no,omitempty"`
	ChequeDate string  `json:"cheque_date,omitempty"`
	ChequeBankCode string  `json:"cheque_bank_code,omitempty"`
}

// TransactionCalculateRequest - request สำหรับคำนวณธุรกรรม
type TransactionCalculateRequest struct {
	Items []TransactionItem `json:"items"`
	Discount string            `json:"discount"`           // ส่วนลดท้ายบิล (discountword)
	DetailDiscount string            `json:"detail_discount"`    // ส่วนลดก่อนชำระเงิน (detaildiscountformula)
	VatType int               `json:"vat_type"`            // 0=แยกนอก, 1=รวมใน, 2=ไม่กระทบภาษี
	VatRate float64           `json:"vatrate"`            // อัตราภาษี default 7
	Payments []PaymentMethod   `json:"payments"`
	DepositAmount float64           `json:"deposit_amount"`     // เงินมัดจำ (sumdeposit)
	IsManualAmount bool              `json:"is_manual_amount"`   // ไม่คำนวณยอดเอง
	InquiryType int               `json:"inquiry_type"`       // ประเภทเอกสาร
	DocCurrency string            `json:"doc_currency"`       // สกุลเงินเอกสาร
	BaseCurrency string            `json:"base_currency"`      // สกุลเงินหลัก
	ExchangeRate float64           `json:"exchange_rate"`      // อัตราแลกเปลี่ยน
	TransactionType string            `json:"transaction_type"`   // ประเภทธุรกรรม (salereturn, purchasereturn, etc.)
	RefTotalOriginal float64           `json:"ref_total_original"` // ยอดเอกสารอ้างอิง (สำหรับ return)
	RoundAmount float64           `json:"round_amount"`       // จำนวนปัดเศษ
}

// TransactionCalculateResponse - response การคำนวณ (enhanced)
type TransactionCalculateResponse struct {
	Items []TransactionItem `json:"items"`

	// === Document totals (Flutter-compatible) ===
	TotalValue float64 `json:"total_value"`                      // ยอดรวมสินค้าทั้งหมด
	TotalQty float64 `json:"total_qty"`                        // จำนวนสินค้าทั้งหมด
	DetailTotalDiscount float64 `json:"detail_total_discount"`            // ส่วนลดก่อนชำระเงิน
	TotalDiscountVatAmount float64 `json:"total_discount_vat_amount"`        // ส่วนลดสินค้ามีภาษี
	TotalDiscountExceptVatAmount float64 `json:"total_discount_except_vat_amount"` // ส่วนลดสินค้ายกเว้นภาษี
	TotalBeforeVat float64 `json:"total_before_vat"`                 // มูลค่าก่อนภาษี
	TotalVatValue float64 `json:"total_vat_value"`                  // มูลค่าภาษี
	TotalAfterVat float64 `json:"total_after_vat"`                  // มูลค่าหลังภาษี
	TotalExceptVat float64 `json:"total_except_vat"`                 // มูลค่ายกเว้นภาษี
	DetailTotalAmount float64 `json:"detail_total_amount"`              // ยอดรวมก่อนหักส่วนลดท้ายบิล
	TotalDiscount float64 `json:"total_discount"`                   // ส่วนลดท้ายบิล
	TotalAmountAfterDiscount float64 `json:"total_amount_after_discount"`      // ยอดรวมหลังหักส่วนลดทั้งหมด
	TotalAmount float64 `json:"total_amount"`                     // ยอดรวมสุทธิ (Base Currency)
	TotalAmountDoc float64 `json:"total_amount_doc"`                // ยอดรวมในสกุลเงินเอกสาร
	TotalValueDoc float64 `json:"total_value_doc"`                 // มูลค่ารวมสินค้า (Document Currency)
	TotalDiscountDoc float64 `json:"total_discount_doc"`              // ส่วนลดท้ายบิล (Document Currency)
	TotalVatValueDoc float64 `json:"total_vat_value_doc"`             // มูลค่าภาษี (Document Currency)
	TotalBeforeVatDoc float64 `json:"total_before_vat_doc"`            // มูลค่าก่อนภาษี (Document Currency)
	TotalAfterVatDoc float64 `json:"total_after_vat_doc"`             // มูลค่าหลังภาษี (Document Currency)
	RefTotalDiff float64 `json:"ref_total_diff,omitempty"`         // ผลต่างจากเอกสารอ้างอิง
	RefTotalCorrect float64 `json:"ref_total_correct,omitempty"`      // ยอดที่ถูกต้องหลังหัก

	// === Legacy fields (backward compat) ===
	SumAmount float64            `json:"sum_amount"`
	DiscountAmount float64            `json:"discount_amount"`
	BeforeVatAmount float64            `json:"before_vat_amount"`
	VatAmount float64            `json:"vat_amount"`
	NetAmount float64            `json:"net_amount"`
	DepositAmount float64            `json:"deposit_amount"`
	PaymentTotal float64            `json:"payment_total"`
	ChangeAmount float64            `json:"change_amount"`
	RemainingAmount float64            `json:"remaining_amount"`
	IsPaymentValid bool               `json:"is_payment_valid"`
	PaymentBreakdown map[string]float64 `json:"payment_breakdown"`
}

// TransactionCalculatorHandler - คำนวณธุรกรรม (enhanced)
func TransactionCalculatorHandler(c echo.Context) error {
	var req TransactionCalculateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	// Default VAT rate
	if req.VatRate <= 0 {
		req.VatRate = 7.0
	}

	// ============================================================
	// Phase 1: Per-item calculation (ported from calamount.dart)
	// ============================================================
	calculatedItems := make([]TransactionItem, len(req.Items))
	var totalValue float64
	var totalQty float64
	var totalAmountVatCale0 float64 // รวมสินค้ามีภาษี (vatcal=0)
	var totalAmountVatCale1 float64 // รวมสินค้ายกเว้นภาษี (vatcal=1)

	for i, item := range req.Items {
		calculated := calculateItemEnhanced(item, req.VatType, req.VatRate, req.InquiryType)
		calculatedItems[i] = calculated

		totalValue += calculated.SumAmount
		totalQty += calculated.Qty

		if calculated.VatCal == 0 {
			totalAmountVatCale0 += calculated.SumAmount
		} else {
			totalAmountVatCale1 += calculated.SumAmount
		}
	}

	// If isManualAmount, only return per-item results
	if req.IsManualAmount {
		response := TransactionCalculateResponse{
			Items:            calculatedItems,
			TotalValue:       roundTo2Decimals(totalValue),
			TotalQty:         totalQty,
			SumAmount:        roundTo2Decimals(totalValue),
			PaymentBreakdown: map[string]float64{},
		}
		return c.JSON(http.StatusOK, map[string]any{
			"status": "success",
			"data":   response,
		})
	}

	// ============================================================
	// Phase 2: Detail discount (first level - detaildiscountformula)
	// Ported from transaction_calculator.dart lines 83-128
	// ============================================================
	detailTotalDiscount := calculateDiscount(totalValue, req.DetailDiscount)

	// Proportional split between taxable and exempt items
	var totalDiscountVatAmount, totalDiscountExceptVatAmount float64
	if totalValue > 0 {
		totalDiscountVatAmount = roundTo2Decimals(detailTotalDiscount * totalAmountVatCale0 / totalValue)
		if math.IsNaN(totalDiscountVatAmount) {
			totalDiscountVatAmount = 0
		}
	}
	totalDiscountExceptVatAmount = roundTo2Decimals(detailTotalDiscount - totalDiscountVatAmount)
	if math.IsNaN(totalDiscountExceptVatAmount) {
		totalDiscountExceptVatAmount = 0
	}

	// ============================================================
	// Phase 3: Deposit-adjusted VAT calculation
	// Ported from transaction_calculator.dart lines 130-233
	// ============================================================
	depositAmount := req.DepositAmount
	totalAmountVatAfterDiscount := totalAmountVatCale0 - totalDiscountVatAmount

	var totalBeforeVat, totalVatValue, totalExceptVat float64

	switch req.VatType {
	case 0: // แยกนอก (Exclusive VAT)
		if depositAmount >= totalAmountVatAfterDiscount {
			// เงินมัดจำ >= รวมมูลค่าภาษี
			totalBeforeVat = 0
			totalVatValue = 0
			excessDeposit := depositAmount - totalAmountVatAfterDiscount
			totalExceptVat = totalAmountVatCale1 - totalDiscountExceptVatAmount - excessDeposit
		} else {
			// เงินมัดจำ < รวมมูลค่าภาษี
			totalBeforeVat = roundTo2Decimals(totalAmountVatAfterDiscount - depositAmount)
			totalVatValue = roundTo2Decimals(totalBeforeVat * req.VatRate / 100)
			totalExceptVat = totalAmountVatCale1 - totalDiscountExceptVatAmount
		}

	case 1: // รวมใน (Inclusive VAT)
		if depositAmount >= totalAmountVatAfterDiscount {
			totalVatValue = 0
			totalBeforeVat = 0
			excessDeposit := depositAmount - totalAmountVatAfterDiscount
			totalExceptVat = totalAmountVatCale1 - totalDiscountExceptVatAmount - excessDeposit
		} else {
			remainingAmount := totalAmountVatAfterDiscount - depositAmount
			totalVatValue = roundTo2Decimals(remainingAmount * req.VatRate / (100 + req.VatRate))
			totalBeforeVat = roundTo2Decimals(remainingAmount - totalVatValue)
			totalExceptVat = totalAmountVatCale1 - totalDiscountExceptVatAmount
		}

	default: // ไม่กระทบภาษี (No VAT)
		if depositAmount >= totalAmountVatAfterDiscount {
			totalBeforeVat = 0
			totalVatValue = 0
			excessDeposit := depositAmount - totalAmountVatAfterDiscount
			totalExceptVat = totalAmountVatCale1 - totalDiscountExceptVatAmount - excessDeposit
		} else {
			totalBeforeVat = roundTo2Decimals(totalAmountVatAfterDiscount - depositAmount)
			totalVatValue = 0
			totalExceptVat = totalAmountVatCale1 - totalDiscountExceptVatAmount
		}
	}

	// Sanitize NaN and negative values
	totalBeforeVat = sanitizeAmount(totalBeforeVat)
	totalVatValue = sanitizeAmount(totalVatValue)
	totalExceptVat = sanitizeAmount(totalExceptVat)

	// ============================================================
	// Phase 4: Total calculations
	// Ported from transaction_calculator.dart lines 257-278
	// ============================================================
	totalAfterVat := sanitizeAmount(roundTo2Decimals(totalBeforeVat + totalVatValue))
	detailTotalAmount := sanitizeAmount(roundTo2Decimals(totalAfterVat + totalExceptVat))

	// ============================================================
	// Phase 5: Bill-end discount (second level - discountword)
	// Ported from transaction_calculator.dart lines 280-309
	// ============================================================
	totalDiscount := calculateDiscount(detailTotalAmount, req.Discount)
	totalAmountAfterDiscount := sanitizeAmount(roundTo2Decimals(detailTotalAmount - totalDiscount))
	totalAmount := totalAmountAfterDiscount

	// ============================================================
	// Phase 6: Multi-currency
	// Ported from transaction_calculator.dart lines 311-343
	// ============================================================
	var totalAmountDoc float64
	var totalValueDoc float64
	var totalDiscountDoc float64
	var totalVatValueDoc float64
	var totalBeforeVatDoc float64
	var totalAfterVatDoc float64
	isMultiCurrency := req.DocCurrency != "" && req.BaseCurrency != "" &&
		req.DocCurrency != req.BaseCurrency &&
		req.ExchangeRate > 0 && req.ExchangeRate != 1.0

	if isMultiCurrency {
		totalAmountDoc = roundTo2Decimals(totalAmount / req.ExchangeRate)
		totalValueDoc = roundTo2Decimals(totalValue / req.ExchangeRate)
		totalDiscountDoc = roundTo2Decimals(totalDiscount / req.ExchangeRate)
		totalVatValueDoc = roundTo2Decimals(totalVatValue / req.ExchangeRate)
		totalBeforeVatDoc = roundTo2Decimals(totalBeforeVat / req.ExchangeRate)
		totalAfterVatDoc = roundTo2Decimals(totalAfterVat / req.ExchangeRate)
	} else {
		totalAmountDoc = totalAmount
		totalValueDoc = totalValue
		totalDiscountDoc = totalDiscount
		totalVatValueDoc = totalVatValue
		totalBeforeVatDoc = totalBeforeVat
		totalAfterVatDoc = totalAfterVat
	}
	if math.IsNaN(totalAmountDoc) || math.IsInf(totalAmountDoc, 0) {
		totalAmountDoc = totalAmount
	}

	// ============================================================
	// Phase 7: Return documents
	// Ported from transaction_calculator.dart lines 353-360
	// ============================================================
	var refTotalDiff, refTotalCorrect float64
	if req.TransactionType == "salereturn" || req.TransactionType == "purchasereturn" {
		refTotalDiff = totalBeforeVat + totalExceptVat
		refTotalCorrect = req.RefTotalOriginal - refTotalDiff
	}

	// ============================================================
	// Phase 8: Payment calculation
	// ============================================================
	paymentTotal, paymentBreakdown := calculatePaymentTotal(req.Payments)

	netAmount := totalAmount - depositAmount
	if netAmount < 0 {
		netAmount = 0
	}

	var changeAmount, remainingAmount float64
	if paymentTotal >= netAmount {
		changeAmount = paymentTotal - netAmount
		remainingAmount = 0
	} else {
		changeAmount = 0
		remainingAmount = netAmount - paymentTotal
	}

	isPaymentValid := remainingAmount <= 0.01

	// ============================================================
	// Build response
	// ============================================================
	response := TransactionCalculateResponse{
		Items: calculatedItems,

		// Enhanced fields (Flutter-compatible)
		TotalValue:                   roundTo2Decimals(totalValue),
		TotalQty:                     totalQty,
		DetailTotalDiscount:          roundTo2Decimals(detailTotalDiscount),
		TotalDiscountVatAmount:       totalDiscountVatAmount,
		TotalDiscountExceptVatAmount: totalDiscountExceptVatAmount,
		TotalBeforeVat:               totalBeforeVat,
		TotalVatValue:                totalVatValue,
		TotalAfterVat:                totalAfterVat,
		TotalExceptVat:               totalExceptVat,
		DetailTotalAmount:            detailTotalAmount,
		TotalDiscount:                roundTo2Decimals(totalDiscount),
		TotalAmountAfterDiscount:     totalAmountAfterDiscount,
		TotalAmount:                  roundTo2Decimals(totalAmount),
		TotalAmountDoc:               totalAmountDoc,
		TotalValueDoc:                totalValueDoc,
		TotalDiscountDoc:             totalDiscountDoc,
		TotalVatValueDoc:             totalVatValueDoc,
		TotalBeforeVatDoc:            totalBeforeVatDoc,
		TotalAfterVatDoc:             totalAfterVatDoc,
		RefTotalDiff:                 refTotalDiff,
		RefTotalCorrect:              refTotalCorrect,

		// Legacy fields (backward compat)
		SumAmount:        roundTo2Decimals(totalValue),
		DiscountAmount:   roundTo2Decimals(detailTotalDiscount + totalDiscount),
		BeforeVatAmount:  totalBeforeVat,
		VatAmount:        totalVatValue,
		NetAmount:        roundTo2Decimals(netAmount),
		DepositAmount:    depositAmount,
		PaymentTotal:     roundTo2Decimals(paymentTotal),
		ChangeAmount:     roundTo2Decimals(changeAmount),
		RemainingAmount:  roundTo2Decimals(remainingAmount),
		IsPaymentValid:   isPaymentValid,
		PaymentBreakdown: paymentBreakdown,
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   response,
	})
}

// calculateItemEnhanced - คำนวณรายการสินค้าพร้อม vatcal support
// Ported from calamount.dart lines 23-65
func calculateItemEnhanced(item TransactionItem, docVatType int, defaultVatRate float64, inquiryType int) TransactionItem {
	result := item
	result.VatType = docVatType
	result.InquiryType = inquiryType
	result.TotalQty = item.Qty

	vatRate := item.VatRate
	if vatRate <= 0 {
		vatRate = defaultVatRate
	}

	// Calculate discount amount
	baseAmount := item.Qty * item.Price
	discountAmount := calculateDiscount(baseAmount, item.Discount)
	result.SumAmount = roundTo2Decimals(baseAmount - discountAmount)
	if result.SumAmount < 0 {
		result.SumAmount = 0
	}
	result.DiscountAmount = roundTo2Decimals(discountAmount)

	// Calculate VAT-related fields based on vattype and vatcal
	switch docVatType {
	case 1: // รวมใน (Inclusive VAT)
		if item.VatCal == 1 {
			// ยกเว้นภาษี (exempt)
			result.PriceExcludeVat = item.Price
			result.SumAmountExcludeVat = result.SumAmount
			result.TotalValueVat = 0
		} else {
			// มีภาษี (taxable)
			result.PriceExcludeVat = roundTo2Decimals(item.Price * 100 / (100 + vatRate))
			result.SumAmountExcludeVat = roundTo2Decimals(result.SumAmount * 100 / (100 + vatRate))
			result.TotalValueVat = roundTo2Decimals(result.SumAmount * vatRate / (100 + vatRate))
		}

	case 0: // แยกนอก (Exclusive VAT)
		result.PriceExcludeVat = item.Price
		result.SumAmountExcludeVat = result.SumAmount
		result.TotalValueVat = roundTo2Decimals(result.SumAmount * vatRate / 100)

	default: // ไม่กระทบภาษี (No VAT)
		result.PriceExcludeVat = item.Price
		result.SumAmountExcludeVat = result.SumAmount
		result.TotalValueVat = 0
	}

	// Legacy fields (backward compat)
	result.NetAmount = result.SumAmountExcludeVat
	result.TaxAmount = result.TotalValueVat
	result.TotalAmount = roundTo2Decimals(result.NetAmount + result.TaxAmount)

	return result
}

// calculateItem - คำนวณรายการสินค้า (legacy wrapper)
func calculateItem(item TransactionItem, defaultVatRate float64) TransactionItem {
	return calculateItemEnhanced(item, item.VatType, defaultVatRate, 0)
}

// sanitizeAmount - ตรวจสอบและแก้ไขค่า NaN/Inf/negative
func sanitizeAmount(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	if val < 0 {
		return 0
	}
	return val
}

// calculateDiscount - คำนวณส่วนลด รองรับ % และจำนวนเงิน
func calculateDiscount(amount float64, discount string) float64 {
	if discount == "" {
		return 0
	}

	discount = strings.TrimSpace(discount)
	if discount == "" || discount == "0" {
		return 0
	}

	// Remove commas
	discount = strings.ReplaceAll(discount, ",", "")

	// ตรวจสอบว่าเป็น % หรือไม่
	if strings.HasSuffix(discount, "%") {
		percentStr := strings.TrimSuffix(discount, "%")
		percent := parseFloat(percentStr)
		if percent > 0 && percent <= 100 {
			return roundTo2Decimals(amount * percent / 100)
		}
		return 0
	}

	// ส่วนลดเป็นจำนวนเงิน
	discountAmount := parseFloat(discount)
	if discountAmount > amount {
		return amount // ส่วนลดไม่เกินยอดรวม
	}
	return discountAmount
}

// calculatePaymentTotal - คำนวณยอดชำระทั้งหมด
func calculatePaymentTotal(payments []PaymentMethod) (float64, map[string]float64) {
	breakdown := make(map[string]float64)
	var total float64 = 0

	for _, payment := range payments {
		total += payment.Amount

		// จัดกลุ่มตาม pay code
		payType := categorizePaymentType(payment.PayCode)
		breakdown[payType] += payment.Amount
	}

	// Round all values
	for k, v := range breakdown {
		breakdown[k] = roundTo2Decimals(v)
	}

	return total, breakdown
}

// categorizePaymentType - จัดประเภทการชำระเงิน
func categorizePaymentType(payCode string) string {
	payCodeLower := strings.ToLower(payCode)

	switch {
	case strings.Contains(payCodeLower, "cash") || payCodeLower == "01":
		return "cash"
	case strings.Contains(payCodeLower, "transfer") || payCodeLower == "02":
		return "transfer"
	case strings.Contains(payCodeLower, "credit") || payCodeLower == "03":
		return "credit_card"
	case strings.Contains(payCodeLower, "cheque") || strings.Contains(payCodeLower, "check") || payCodeLower == "04":
		return "cheque"
	case strings.Contains(payCodeLower, "coupon") || payCodeLower == "05":
		return "coupon"
	case strings.Contains(payCodeLower, "qr") || payCodeLower == "06":
		return "qr_payment"
	case strings.Contains(payCodeLower, "deposit") || payCodeLower == "07":
		return "deposit"
	default:
		return "other"
	}
}

// parseFloat - แปลง string เป็น float64
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	var result float64
	_, err := parseFloatValue(s, &result)
	if err != nil {
		return 0
	}
	return result
}

// parseFloatValue - helper function to parse float
func parseFloatValue(s string, result *float64) (bool, error) {
	// Remove commas if present
	s = strings.ReplaceAll(s, ",", "")

	var val float64
	n, err := scanFloat(s, &val)
	if err != nil || n == 0 {
		return false, err
	}
	*result = val
	return true, nil
}

// scanFloat - simple float scanner
func scanFloat(s string, val *float64) (int, error) {
	var v float64
	n := 0
	negative := false
	decimal := false
	decimalPlaces := 0.1

	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if c == '.' {
			decimal = true
			continue
		}
		if c >= '0' && c <= '9' {
			digit := float64(c - '0')
			if decimal {
				v += digit * decimalPlaces
				decimalPlaces *= 0.1
			} else {
				v = v*10 + digit
			}
			n++
		}
	}

	if negative {
		v = -v
	}
	*val = v
	return n, nil
}

// roundTo2Decimals - ปัดเศษ 2 ตำแหน่ง
func roundTo2Decimals(val float64) float64 {
	return math.Round(val*100) / 100
}

// QuickCalcRequest - request สำหรับคำนวณแบบย่อ
type QuickCalcRequest struct {
	Amount float64 `json:"amount"`
	Discount string  `json:"discount"`
	VatType int     `json:"vat_type"`
	VatRate float64 `json:"vatrate"`
}

// QuickCalculatorHandler - คำนวณแบบย่อ (sum amount เดียว)
func QuickCalculatorHandler(c echo.Context) error {
	var req QuickCalcRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.VatRate <= 0 {
		req.VatRate = 7.0
	}

	// คำนวณส่วนลด
	discountAmount := calculateDiscount(req.Amount, req.Discount)
	afterDiscount := req.Amount - discountAmount

	var beforeVat, vat, total float64

	switch req.VatType {
	case 0: // แยกนอก
		beforeVat = afterDiscount
		vat = roundTo2Decimals(afterDiscount * req.VatRate / 100)
		total = beforeVat + vat
	case 1: // รวมใน
		total = afterDiscount
		beforeVat = roundTo2Decimals(afterDiscount * 100 / (100 + req.VatRate))
		vat = total - beforeVat
	default: // ไม่กระทบภาษี
		beforeVat = afterDiscount
		vat = 0
		total = afterDiscount
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data": map[string]float64{
			"amount":          req.Amount,
			"discount_amount": roundTo2Decimals(discountAmount),
			"after_discount":  roundTo2Decimals(afterDiscount),
			"before_vat":      roundTo2Decimals(beforeVat),
			"vat_amount":      roundTo2Decimals(vat),
			"total_amount":    roundTo2Decimals(total),
		},
	})
}

// ValidatePaymentRequest - request ตรวจสอบการชำระเงิน
type ValidatePaymentRequest struct {
	TotalAmount float64         `json:"total_amount"`
	DepositAmount float64         `json:"deposit_amount"`
	Payments []PaymentMethod `json:"payments"`
}

// ValidatePaymentHandler - ตรวจสอบการชำระเงิน
func ValidatePaymentHandler(c echo.Context) error {
	var req ValidatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	netAmount := req.TotalAmount - req.DepositAmount
	if netAmount < 0 {
		netAmount = 0
	}

	paymentTotal, breakdown := calculatePaymentTotal(req.Payments)

	var changeAmount, remainingAmount float64
	if paymentTotal >= netAmount {
		changeAmount = paymentTotal - netAmount
		remainingAmount = 0
	} else {
		changeAmount = 0
		remainingAmount = netAmount - paymentTotal
	}

	isValid := remainingAmount <= 0.01

	var validationMessage string
	if isValid {
		validationMessage = "Payment is complete"
	} else {
		validationMessage = "Insufficient payment amount"
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data": map[string]any{
			"net_amount":        roundTo2Decimals(netAmount),
			"payment_total":     roundTo2Decimals(paymentTotal),
			"change_amount":     roundTo2Decimals(changeAmount),
			"remaining_amount":  roundTo2Decimals(remainingAmount),
			"is_valid":          isValid,
			"message":           validationMessage,
			"payment_breakdown": breakdown,
		},
	})
}
