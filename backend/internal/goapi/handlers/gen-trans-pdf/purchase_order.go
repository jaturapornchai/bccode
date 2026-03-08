package gentranspdf

import (
	"fmt"
	"sort"

	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GeneratePurchaseOrderPDF - สร้าง PDF ใบสั่งซื้อ (รองรับ multi-currency)
func GeneratePurchaseOrderPDF(document map[string]interface{}, payload GenPDFPayload) (*gofpdf.Fpdf, error) {
	// Set default title
	if payload.Title == "เอกสาร" || payload.Title == "" {
		payload.Title = "ใบสั่งซื้อ"
	}

	// ตรวจสอบว่าเป็น multi-currency หรือไม่
	docCurrency := GetStringValue(document, "doc_currency")
	exchangeRate := GetFloatValue(document, "exchangerate")

	// ถ้าไม่มี doc_currency หรือ exchangeRate = 0/1 → ใช้ base PDF ปกติ
	if docCurrency == "" || exchangeRate == 0 || exchangeRate == 1 {
		return GenerateBasePDF(document, payload)
	}

	// สร้าง PDF แบบ multi-currency
	return generatePurchaseOrderMultiCurrencyPDF(document, payload)
}

// generatePurchaseOrderMultiCurrencyPDF - สร้าง PDF ใบสั่งซื้อแบบ multi-currency
// แสดง doc currency เป็นหลัก และแสดง base currency (THB) เป็นรอง
func generatePurchaseOrderMultiCurrencyPDF(document map[string]interface{}, payload GenPDFPayload) (*gofpdf.Fpdf, error) {
	theme := GetEffectiveTheme(payload)
	fontSizes := GetEffectiveFontSizes(payload)
	template := GetEffectiveTemplate(payload)
	labels := GetLabels(payload.Language)

	pdf := gofpdf.New(payload.Orientation, "mm", payload.PageSize, "fonts")
	pdf.SetMargins(template.MarginLeft, template.MarginTop, template.MarginRight)
	pdf.SetAutoPageBreak(true, template.MarginBottom)

	fontFamily := RegisterFontsWithPreferred(pdf, payload.Language, payload.FontFamily)
	if fontFamily == "" {
		fontFamily = "Arial"
	}

	setupFooter(pdf, document, fontFamily, theme, fontSizes, template, labels)
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	contentWidth := pageWidth - leftMargin - rightMargin

	// === DOCUMENT TITLE ===
	if template.ShowDocTitle {
		RenderDocHeader(pdf, document, payload, fontFamily, contentWidth, theme, fontSizes, template)
	}

	// === COMPANY & CUSTOMER INFO ===
	if template.ShowCompanyInfo || template.ShowCustomerInfo {
		RenderCompanyCustomer(pdf, document, payload, fontFamily, contentWidth, theme, fontSizes, template)
	}

	// === CURRENCY INFO (แสดงข้อมูลสกุลเงิน) ===
	renderPOCurrencyInfo(pdf, document, fontFamily, contentWidth, theme, fontSizes, labels)

	// === DETAILS TABLE (multi-currency) ===
	renderPODetailsTable(pdf, document, payload, fontFamily, contentWidth, payload.Orientation, theme, fontSizes, template)

	// === SUMMARY (multi-currency) ===
	renderPOSummary(pdf, document, payload, fontFamily, contentWidth, theme, fontSizes, template, labels)

	// === SIGNATURE ===
	RenderSignatureSection(pdf, fontFamily, contentWidth, theme, fontSizes, template, labels)

	// === WATERMARK ===
	if template.ShowWatermark && template.WatermarkText != "" {
		RenderWatermark(pdf, template.WatermarkText, fontFamily)
	}
	if payload.IsPreview {
		RenderPreviewWatermark(pdf, fontFamily)
	}

	// === BORDER ===
	if template.ShowBorder {
		RenderDocumentBorder(pdf, template, theme)
	}

	return pdf, nil
}

// renderPOCurrencyInfo - แสดงข้อมูลสกุลเงิน (doc currency + อัตราแลกเปลี่ยน)
func renderPOCurrencyInfo(pdf *gofpdf.Fpdf, doc map[string]interface{}, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, labels Labels) {
	docCurrency := GetStringValue(doc, "doc_currency")
	docCurrencySymbol := GetStringValue(doc, "doc_currencysymbol")
	exchangeRate := GetFloatValue(doc, "exchangerate")

	fontSize := float64(fontSizes.Header) - 1

	pdf.SetFont(fontFamily, "", fontSize)
	r, g, b := HexToRGB(theme.Section.LabelColor)
	pdf.SetTextColor(r, g, b)

	// แสดง: สกุลเงินเอกสาร: USD ($) | อัตราแลกเปลี่ยน: 1 USD = 35.5000 THB
	currencyText := fmt.Sprintf("%s: %s (%s)  |  %s: 1 %s = %s %s",
		labels.Currency, docCurrency, docCurrencySymbol,
		"อัตราแลกเปลี่ยน", docCurrency,
		FormatNumber(exchangeRate, 4), "THB")

	pdf.CellFormat(contentWidth, 5, currencyText, "", 1, "L", false, 0, "")
	pdf.Ln(2)
}

// renderPODetailsTable - ตารางรายละเอียดสินค้า (multi-currency)
// แสดง doc currency เป็นหลัก, base currency เป็นรอง
func renderPODetailsTable(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, orientation string, theme Theme, fontSizes FontSizes, template Template) {
	styleConfig := GetTableStyleConfig(template.TableStyle)
	detailFontSize := float64(fontSizes.Detail)
	labels := GetLabels(payload.Language)

	docCurrencySymbol := GetStringValue(doc, "doc_currencysymbol")
	exchangeRate := GetFloatValue(doc, "exchangerate")

	// ตรวจสอบว่าต้องการแสดง 2 สกุลเงินหรือไม่ (default: true)
	showDualCurrency := true
	if payload.ShowDualCurrency != nil && !*payload.ShowDualCurrency {
		showDualCurrency = false
	}

	// Column widths: ลำดับ, รหัส, รายการ, หน่วย, จำนวน, ราคา/หน่วย(doc), จำนวนเงิน(doc)
	var colWidthsPercent []float64
	if orientation == "L" {
		colWidthsPercent = []float64{5, 10, 50, 8, 8, 10, 9}
	} else {
		colWidthsPercent = []float64{5, 12, 38, 9, 10, 12, 14}
	}

	colWidths := make([]float64, len(colWidthsPercent))
	for i, percent := range colWidthsPercent {
		colWidths[i] = contentWidth * percent / 100.0
	}

	// === TABLE HEADER ===
	fontStyle := ""
	if styleConfig.HeaderBold {
		fontStyle = "B"
	}
	pdf.SetFont(fontFamily, fontStyle, detailFontSize)

	if styleConfig.HeaderFill && template.ShowHeaderBackground {
		r, g, b := HexToRGB(theme.Table.HeaderBgColor)
		pdf.SetFillColor(r, g, b)
		r, g, b = HexToRGB(theme.Table.HeaderTextColor)
		pdf.SetTextColor(r, g, b)
	} else {
		r, g, b := HexToRGB(theme.Table.HeaderBgColor)
		pdf.SetTextColor(r, g, b)
		pdf.SetFillColor(255, 255, 255)
	}

	r, g, b := HexToRGB(theme.Table.BorderColor)
	pdf.SetDrawColor(r, g, b)
	pdf.SetLineWidth(styleConfig.BorderWidth)

	// Header ใส่สัญลักษณ์สกุลเงินเอกสาร
	unitPriceHeader := fmt.Sprintf("%s(%s)", labels.UnitPrice, docCurrencySymbol)
	amountHeader := fmt.Sprintf("%s(%s)", labels.Amount, docCurrencySymbol)

	headers := []string{labels.No, labels.ItemCode, labels.Item, labels.Unit, labels.Qty, unitPriceHeader, amountHeader}
	aligns := []string{"C", "C", "C", "C", "C", "C", "C"}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], styleConfig.HeaderHeight, header, styleConfig.HeaderBorder, 0, aligns[i], styleConfig.HeaderFill && template.ShowHeaderBackground, 0, "")
	}
	pdf.Ln(-1)

	// === TABLE ROWS ===
	pdf.SetFont(fontFamily, "", detailFontSize)
	r, g, b = HexToRGB(theme.Table.RowTextColor)
	pdf.SetTextColor(r, g, b)
	r, g, b = HexToRGB(theme.Table.BorderColor)
	pdf.SetDrawColor(r, g, b)

	details, ok := doc["details"].(primitive.A)
	if !ok {
		r, g, b = HexToRGB(theme.Table.RowBgColor)
		pdf.SetFillColor(r, g, b)
		pdf.CellFormat(contentWidth, styleConfig.RowHeight, labels.NoItems, "1", 1, "C", true, 0, "")
		return
	}

	// Sort details by linenumber
	type detailItem struct {
		lineNumber int
		data       map[string]interface{}
	}
	sortedDetails := make([]detailItem, 0, len(details))
	for _, d := range details {
		if detailMap, ok := d.(map[string]interface{}); ok {
			lineNum := GetIntValue(detailMap, "linenumber")
			sortedDetails = append(sortedDetails, detailItem{lineNumber: lineNum, data: detailMap})
		}
	}
	sort.Slice(sortedDetails, func(i, j int) bool {
		return sortedDetails[i].lineNumber < sortedDetails[j].lineNumber
	})

	lineSpacing := payload.LineSpacing
	if lineSpacing <= 0 {
		lineSpacing = 1.0
	}
	lineHeight := 4.0 * lineSpacing

	for idx, item := range sortedDetails {
		detail := item.data

		// Get item name
		itemName := ""
		if itemNames, ok := detail["itemnames"].(primitive.A); ok && len(itemNames) > 0 {
			if nameObj, ok := itemNames[0].(map[string]interface{}); ok {
				itemName = GetStringValue(nameObj, "name")
			}
		}

		itemCode := GetStringValue(detail, "itemcode")
		unitCode := GetStringValue(detail, "unitcode")
		qty := GetFloatValue(detail, "qty")

		// ใช้ราคา doc currency เป็นหลัก
		priceDoc := GetFloatValue(detail, "price_doc")
		sumAmountDoc := GetFloatValue(detail, "sumamount_doc")

		// ถ้าไม่มี price_doc → คำนวณจาก price / exchangeRate
		priceBase := GetFloatValue(detail, "price")
		sumAmountBase := GetFloatValue(detail, "sumamount")
		if priceDoc == 0 && priceBase > 0 && exchangeRate > 0 {
			priceDoc = priceBase / exchangeRate
		}
		if sumAmountDoc == 0 && sumAmountBase > 0 && exchangeRate > 0 {
			sumAmountDoc = sumAmountBase / exchangeRate
		}

		// Calculate text lines for item name
		itemNameLines := WrapThaiText(pdf, itemName, colWidths[2]-2*styleConfig.CellPadding)
		numLines := len(itemNameLines)
		if numLines < 1 {
			numLines = 1
		}

		// แถว multi-currency ต้องการพื้นที่เพิ่มสำหรับ base currency line
		extraLine := 0.0
		if showDualCurrency {
			extraLine = lineHeight // พื้นที่สำหรับแสดง base currency
		}
		rowHeight := float64(numLines)*lineHeight + extraLine
		if rowHeight < styleConfig.RowHeight+extraLine {
			rowHeight = styleConfig.RowHeight + extraLine
		}

		// Alternate row colors
		if styleConfig.RowFill {
			if template.AlternateRowColor && idx%2 == 1 {
				r, g, b = HexToRGB(theme.Table.RowAltBgColor)
			} else {
				r, g, b = HexToRGB(theme.Table.RowBgColor)
			}
			pdf.SetFillColor(r, g, b)
		}

		r, g, b = HexToRGB(theme.Table.BorderColor)
		pdf.SetDrawColor(r, g, b)
		pdf.SetLineWidth(styleConfig.BorderWidth)

		startY := pdf.GetY()
		startX := template.MarginLeft

		// Column 0: ลำดับ
		pdf.SetXY(startX, startY)
		pdf.CellFormat(colWidths[0], rowHeight, fmt.Sprintf("%d", idx+1), styleConfig.RowBorder, 0, "C", styleConfig.RowFill, 0, "")

		// Column 1: รหัสสินค้า
		pdf.SetXY(startX+colWidths[0], startY)
		pdf.CellFormat(colWidths[1], rowHeight, itemCode, styleConfig.RowBorder, 0, "L", styleConfig.RowFill, 0, "")

		// Column 2: รายการ (with text wrapping)
		pdf.SetXY(startX+colWidths[0]+colWidths[1], startY)
		pdf.CellFormat(colWidths[2], rowHeight, "", styleConfig.RowBorder, 0, "L", styleConfig.RowFill, 0, "")
		pdf.SetXY(startX+colWidths[0]+colWidths[1]+styleConfig.CellPadding, startY+styleConfig.CellPadding)
		for i, line := range itemNameLines {
			pdf.SetX(startX + colWidths[0] + colWidths[1] + styleConfig.CellPadding)
			pdf.CellFormat(colWidths[2]-2*styleConfig.CellPadding, lineHeight, line, "", 0, "L", false, 0, "")
			if i < len(itemNameLines)-1 {
				pdf.Ln(lineHeight)
			}
		}

		// Column 3: หน่วย
		pdf.SetXY(startX+colWidths[0]+colWidths[1]+colWidths[2], startY)
		pdf.CellFormat(colWidths[3], rowHeight, unitCode, styleConfig.RowBorder, 0, "C", styleConfig.RowFill, 0, "")

		// Column 4: จำนวน
		pdf.SetXY(startX+colWidths[0]+colWidths[1]+colWidths[2]+colWidths[3], startY)
		pdf.CellFormat(colWidths[4], rowHeight, FormatNumber(qty, 0), styleConfig.RowBorder, 0, "R", styleConfig.RowFill, 0, "")

		// Column 5: ราคา/หน่วย (doc currency หลัก + base currency รอง)
		col5X := startX + colWidths[0] + colWidths[1] + colWidths[2] + colWidths[3] + colWidths[4]
		pdf.SetXY(col5X, startY)
		pdf.CellFormat(colWidths[5], rowHeight, "", styleConfig.RowBorder, 0, "R", styleConfig.RowFill, 0, "")
		// Doc currency (บรรทัดหลัก)
		pdf.SetFont(fontFamily, "", detailFontSize)
		r2, g2, b2 := HexToRGB(theme.Table.RowTextColor)
		pdf.SetTextColor(r2, g2, b2)
		pdf.SetXY(col5X, startY+1)
		pdf.CellFormat(colWidths[5]-1, lineHeight, FormatNumber(priceDoc, 2), "", 0, "R", false, 0, "")
		// Base currency (บรรทัดรอง - สีจาง)
		if showDualCurrency {
			pdf.SetFont(fontFamily, "", detailFontSize-1)
			pdf.SetTextColor(130, 130, 130)
			pdf.SetXY(col5X, startY+1+lineHeight)
			pdf.CellFormat(colWidths[5]-1, lineHeight, fmt.Sprintf("฿%s", FormatNumber(priceBase, 2)), "", 0, "R", false, 0, "")
		}

		// Column 6: จำนวนเงิน (doc currency หลัก + base currency รอง)
		col6X := col5X + colWidths[5]
		pdf.SetXY(col6X, startY)
		pdf.CellFormat(colWidths[6], rowHeight, "", styleConfig.RowBorder, 0, "R", styleConfig.RowFill, 0, "")
		// Doc currency
		pdf.SetFont(fontFamily, "", detailFontSize)
		pdf.SetTextColor(r2, g2, b2)
		pdf.SetXY(col6X, startY+1)
		pdf.CellFormat(colWidths[6]-1, lineHeight, FormatNumber(sumAmountDoc, 2), "", 0, "R", false, 0, "")
		// Base currency
		if showDualCurrency {
			pdf.SetFont(fontFamily, "", detailFontSize-1)
			pdf.SetTextColor(130, 130, 130)
			pdf.SetXY(col6X, startY+1+lineHeight)
			pdf.CellFormat(colWidths[6]-1, lineHeight, fmt.Sprintf("฿%s", FormatNumber(sumAmountBase, 2)), "", 0, "R", false, 0, "")
		}

		// Reset font
		pdf.SetFont(fontFamily, "", detailFontSize)
		r, g, b = HexToRGB(theme.Table.RowTextColor)
		pdf.SetTextColor(r, g, b)

		pdf.SetY(startY + rowHeight)
	}

	// Bottom border
	if styleConfig.ShowBottomBorder {
		r, g, b = HexToRGB(theme.Colors.Primary)
		pdf.SetDrawColor(r, g, b)
		pdf.SetLineWidth(styleConfig.BorderWidth)
		totalTableWidth := colWidths[0] + colWidths[1] + colWidths[2] + colWidths[3] + colWidths[4] + colWidths[5] + colWidths[6]
		pdf.Line(template.MarginLeft, pdf.GetY(), template.MarginLeft+totalTableWidth, pdf.GetY())
	}

	pdf.Ln(template.TableSpacing)
}

// renderPOSummary - ส่วนสรุปยอด multi-currency
// แสดง doc currency เป็นหลัก และ base currency (THB) เป็นรอง
func renderPOSummary(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)

	docCurrency := GetStringValue(doc, "doc_currency")
	docCurrencySymbol := GetStringValue(doc, "doc_currencysymbol")
	exchangeRate := GetFloatValue(doc, "exchangerate")

	// ตรวจสอบว่าต้องการแสดง 2 สกุลเงินหรือไม่ (default: true)
	showDualCurrency := true
	if payload.ShowDualCurrency != nil && !*payload.ShowDualCurrency {
		showDualCurrency = false
	}

	// ยอดรวม base currency (THB)
	totalAmount := GetFloatValue(doc, "totalamount")
	totalDiscount := GetFloatValue(doc, "totaldiscount")
	totalVat := GetFloatValue(doc, "totalvatvalue")
	totalAfterVat := GetFloatValue(doc, "totalaftervat")

	// ยอดรวม doc currency
	totalAmountDoc := GetFloatValue(doc, "totalamount_doc")
	totalDiscountDoc := GetFloatValue(doc, "totaldiscount_doc")
	totalVatDoc := GetFloatValue(doc, "totalvatvalue_doc")
	totalAfterVatDoc := GetFloatValue(doc, "totalaftervat_doc")

	// ถ้าไม่มียอด doc → คำนวณจาก base / exchangeRate
	if totalAmountDoc == 0 && totalAmount > 0 && exchangeRate > 0 {
		totalAmountDoc = totalAmount / exchangeRate
	}
	if totalDiscountDoc == 0 && totalDiscount > 0 && exchangeRate > 0 {
		totalDiscountDoc = totalDiscount / exchangeRate
	}
	if totalVatDoc == 0 && totalVat > 0 && exchangeRate > 0 {
		totalVatDoc = totalVat / exchangeRate
	}
	if totalAfterVatDoc == 0 && totalAfterVat > 0 && exchangeRate > 0 {
		totalAfterVatDoc = totalAfterVat / exchangeRate
	}

	// If no totals, calculate from details
	if totalAmount == 0 {
		if details, ok := doc["details"].(primitive.A); ok {
			for _, d := range details {
				if detailMap, ok := d.(map[string]interface{}); ok {
					totalAmount += GetFloatValue(detailMap, "sumamount")
				}
			}
		}
		if exchangeRate > 0 {
			totalAmountDoc = totalAmount / exchangeRate
		}
	}

	// Layout
	summaryWidth := 100.0
	pageWidth, _ := pdf.GetPageSize()
	_, _, rightMargin, _ := pdf.GetMargins()
	summaryX := pageWidth - rightMargin - summaryWidth

	labelWidth := 55.0
	valueWidth := 45.0
	rowHeight := 5.5

	var r, g, b int

	// === สร้างรายการสรุป ===
	type summaryRow struct {
		label      string
		docValue   float64
		baseValue  float64
		highlight  bool
	}

	var rows []summaryRow

	if template.ShowSubtotal {
		rows = append(rows, summaryRow{labels.Subtotal, totalAmountDoc, totalAmount, false})
	}
	if template.ShowTotalDiscount && totalDiscount > 0 {
		rows = append(rows, summaryRow{labels.Discount, totalDiscountDoc, totalDiscount, false})
		afterDiscDoc := totalAmountDoc - totalDiscountDoc
		afterDiscBase := totalAmount - totalDiscount
		rows = append(rows, summaryRow{labels.AfterDiscount, afterDiscDoc, afterDiscBase, false})
	}
	if template.ShowTotalTax {
		rows = append(rows, summaryRow{labels.Tax, totalVatDoc, totalVat, false})
	}
	rows = append(rows, summaryRow{labels.GrandTotal, totalAfterVatDoc, totalAfterVat, true})

	// === วาด summary rows ===
	for _, row := range rows {
		pdf.SetX(summaryX)

		if row.highlight {
			// แถว Grand Total — เน้น
			pdf.SetFont(fontFamily, "B", summaryFontSize+1)
			r, g, b = HexToRGB(theme.Summary.HighlightBgColor)
			pdf.SetFillColor(r, g, b)
			r, g, b = HexToRGB(theme.Summary.HighlightTextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetDrawColor(r, g, b)

			// Doc currency (หลัก)
			docValueText := fmt.Sprintf("%s %s", docCurrencySymbol, FormatNumber(row.docValue, 2))
			pdf.CellFormat(labelWidth, rowHeight+1, row.label, "1", 0, "R", true, 0, "")
			pdf.CellFormat(valueWidth, rowHeight+1, docValueText, "1", 1, "R", true, 0, "")

			// Base currency (รอง)
			if showDualCurrency {
				pdf.SetX(summaryX)
				pdf.SetFont(fontFamily, "", summaryFontSize-1)
				r, g, b = HexToRGB(theme.Summary.BgColor)
				pdf.SetFillColor(r, g, b)
				r, g, b = HexToRGB(theme.Summary.BorderColor)
				pdf.SetDrawColor(r, g, b)
				pdf.SetTextColor(130, 130, 130)
				baseValueText := fmt.Sprintf("฿%s", FormatNumber(row.baseValue, 2))
				pdf.CellFormat(labelWidth, rowHeight-1, fmt.Sprintf("เทียบเท่า %s", "THB"), "LR", 0, "R", true, 0, "")
				pdf.CellFormat(valueWidth, rowHeight-1, baseValueText, "LR", 1, "R", true, 0, "")
			}
		} else {
			// แถวปกติ
			pdf.SetFont(fontFamily, "", summaryFontSize)
			r, g, b = HexToRGB(theme.Summary.BgColor)
			pdf.SetFillColor(r, g, b)
			r, g, b = HexToRGB(theme.Summary.BorderColor)
			pdf.SetDrawColor(r, g, b)
			r, g, b = HexToRGB(theme.Summary.TextColor)
			pdf.SetTextColor(r, g, b)

			// Doc currency (หลัก)
			docValueText := fmt.Sprintf("%s %s", docCurrencySymbol, FormatNumber(row.docValue, 2))
			pdf.CellFormat(labelWidth, rowHeight, row.label, "1", 0, "R", true, 0, "")
			pdf.CellFormat(valueWidth, rowHeight, docValueText, "1", 1, "R", true, 0, "")

			// Base currency (รอง — สีจาง)
			if showDualCurrency {
				pdf.SetX(summaryX)
				pdf.SetFont(fontFamily, "", summaryFontSize-2)
				pdf.SetTextColor(150, 150, 150)
				baseValueText := fmt.Sprintf("฿%s", FormatNumber(row.baseValue, 2))
				pdf.CellFormat(labelWidth, rowHeight-2, "", "LR", 0, "R", true, 0, "")
				pdf.CellFormat(valueWidth, rowHeight-2, baseValueText, "LR", 1, "R", true, 0, "")

				// Reset text color
				r, g, b = HexToRGB(theme.Summary.TextColor)
				pdf.SetTextColor(r, g, b)
			}
		}
	}

	// === จำนวนเงินเป็นตัวอักษร ===
	pdf.Ln(1)

	// Doc currency text (ถ้าเป็น THB ใช้ NumberToThaiText)
	pdf.SetX(summaryX)
	pdf.SetFont(fontFamily, "", summaryFontSize-1)
	r, g, b = HexToRGB(theme.Summary.TextColor)
	pdf.SetTextColor(r, g, b)

	totalDocForText := totalAfterVatDoc
	if totalDocForText == 0 {
		totalDocForText = totalAmountDoc
	}
	totalBaseForText := totalAfterVat
	if totalBaseForText == 0 {
		totalBaseForText = totalAmount
	}

	// แสดงยอดรวมเป็นตัวอักษร
	if totalDocForText > 0 {
		// แสดง doc currency amount
		docAmountText := fmt.Sprintf("%s %s %s", docCurrencySymbol, FormatNumber(totalDocForText, 2), docCurrency)
		pdf.SetX(summaryX)
		pdf.CellFormat(labelWidth+valueWidth, 3.5, docAmountText, "", 1, "R", false, 0, "")

		// แสดง base currency เป็นตัวอักษรไทย (เฉพาะเมื่อแสดง 2 สกุลเงิน)
		if showDualCurrency && totalBaseForText > 0 {
			pdf.SetX(summaryX)
			pdf.MultiCell(labelWidth+valueWidth, 3.5, NumberToThaiText(totalBaseForText), "", "R", false)
		}
	}

	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(template.SummarySpacing)
}
