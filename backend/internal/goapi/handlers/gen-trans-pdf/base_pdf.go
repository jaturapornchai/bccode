package gentranspdf

import (
	"fmt"
	"sort"

	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateBasePDF - ฟังก์ชันพื้นฐานสำหรับสร้าง PDF
func GenerateBasePDF(document map[string]interface{}, payload GenPDFPayload) (*gofpdf.Fpdf, error) {
	// Get effective theme, font sizes, template, and labels
	theme := GetEffectiveTheme(payload)
	fontSizes := GetEffectiveFontSizes(payload)
	template := GetEffectiveTemplate(payload)
	labels := GetLabels(payload.Language)

	// Create PDF with specified orientation and page size
	// Set font directory to "fonts" for UTF-8 font support
	pdf := gofpdf.New(payload.Orientation, "mm", payload.PageSize, "fonts")
	pdf.SetMargins(template.MarginLeft, template.MarginTop, template.MarginRight)
	pdf.SetAutoPageBreak(true, template.MarginBottom)

	// Register UTF-8 fonts - ใช้ preferred font ถ้าระบุมา, ไม่งั้นใช้ Universal Font
	fontFamily := RegisterFontsWithPreferred(pdf, payload.Language, payload.FontFamily)
	if fontFamily == "" {
		fontFamily = "Arial"
	}

	// Setup footer based on style
	setupFooter(pdf, document, fontFamily, theme, fontSizes, template, labels)

	pdf.AddPage()

	// Get page dimensions
	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	contentWidth := pageWidth - leftMargin - rightMargin

	// === DOCUMENT TITLE (Top Right) ===
	if template.ShowDocTitle {
		RenderDocHeader(pdf, document, payload, fontFamily, contentWidth, theme, fontSizes, template)
	}

	// === COMPANY & CUSTOMER INFO (Two Columns) ===
	if template.ShowCompanyInfo || template.ShowCustomerInfo {
		RenderCompanyCustomer(pdf, document, payload, fontFamily, contentWidth, theme, fontSizes, template)
	}

	// === DETAILS TABLE ===
	RenderDetailsTable(pdf, document, payload, fontFamily, contentWidth, payload.Orientation, theme, fontSizes, template)

	// === SUMMARY & SIGNATURE SECTION ===
	RenderSummarySignature(pdf, document, payload, fontFamily, contentWidth, theme, fontSizes, template)

	// === SIGNATURE SECTION (if enabled) ===
	RenderSignatureSection(pdf, fontFamily, contentWidth, theme, fontSizes, template, labels)

	// === WATERMARK (if enabled) ===
	if template.ShowWatermark && template.WatermarkText != "" {
		RenderWatermark(pdf, template.WatermarkText, fontFamily)
	}

	// === PREVIEW WATERMARK (ถ้าเป็น preview mode) ===
	if payload.IsPreview {
		RenderPreviewWatermark(pdf, fontFamily)
	}

	// === DOCUMENT BORDER (if enabled) ===
	if template.ShowBorder {
		RenderDocumentBorder(pdf, template, theme)
	}

	return pdf, nil
}

// RenderDocumentBorder - วาดกรอบรอบเอกสาร
func RenderDocumentBorder(pdf *gofpdf.Fpdf, template Template, theme Theme) {
	pageWidth, pageHeight := pdf.GetPageSize()
	margin := 5.0 // ระยะขอบจากขอบกระดาษ

	r, g, b := HexToRGB(theme.Colors.Primary)
	pdf.SetDrawColor(r, g, b)

	switch template.BorderStyle {
	case "solid":
		pdf.SetLineWidth(0.5)
		pdf.Rect(margin, margin, pageWidth-2*margin, pageHeight-2*margin, "D")

	case "dashed":
		pdf.SetLineWidth(0.5)
		// Draw dashed border using multiple short lines
		dashLen := 3.0
		gapLen := 2.0
		drawDashedRect(pdf, margin, margin, pageWidth-2*margin, pageHeight-2*margin, dashLen, gapLen)

	case "double":
		pdf.SetLineWidth(0.3)
		pdf.Rect(margin, margin, pageWidth-2*margin, pageHeight-2*margin, "D")
		pdf.Rect(margin+2, margin+2, pageWidth-2*margin-4, pageHeight-2*margin-4, "D")

	case "rounded":
		pdf.SetLineWidth(0.5)
		// gofpdf doesn't have native rounded rect, use regular rect
		pdf.Rect(margin, margin, pageWidth-2*margin, pageHeight-2*margin, "D")

	default: // "none" - do nothing
		return
	}
}

// drawDashedRect - วาดสี่เหลี่ยมเส้นประ
func drawDashedRect(pdf *gofpdf.Fpdf, x, y, w, h, dashLen, gapLen float64) {
	// Top line
	drawDashedLine(pdf, x, y, x+w, y, dashLen, gapLen)
	// Right line
	drawDashedLine(pdf, x+w, y, x+w, y+h, dashLen, gapLen)
	// Bottom line
	drawDashedLine(pdf, x+w, y+h, x, y+h, dashLen, gapLen)
	// Left line
	drawDashedLine(pdf, x, y+h, x, y, dashLen, gapLen)
}

// drawDashedLine - วาดเส้นประ (simplified - draws solid line)
func drawDashedLine(pdf *gofpdf.Fpdf, x1, y1, x2, y2, _, _ float64) {
	// Simple implementation - just draw solid line
	// Full dashed line would require SetDashPattern which gofpdf supports
	pdf.Line(x1, y1, x2, y2)
}

// RenderWatermark - แสดงลายน้ำบนหน้าเอกสาร
func RenderWatermark(pdf *gofpdf.Fpdf, text string, fontFamily string) {
	pageWidth, pageHeight := pdf.GetPageSize()

	// Save current state
	pdf.SetAlpha(0.1, "Normal")
	pdf.SetFont(fontFamily, "B", 60)
	pdf.SetTextColor(150, 150, 150)

	// Calculate position for diagonal watermark
	centerX := pageWidth / 2
	centerY := pageHeight / 2

	// Get text width
	textWidth := pdf.GetStringWidth(text)

	// Draw watermark at center with rotation
	pdf.TransformBegin()
	pdf.TransformRotate(-45, centerX, centerY)
	pdf.SetXY(centerX-textWidth/2, centerY-10)
	pdf.CellFormat(textWidth, 20, text, "", 0, "C", false, 0, "")
	pdf.TransformEnd()

	// Reset alpha
	pdf.SetAlpha(1.0, "Normal")
}

// RenderPreviewWatermark - แสดงลายน้ำ "Preview" ตัวใหญ่ เอียง 45 องศา กลางหน้า
// ใช้สำหรับ preview mode (กรณียังไม่ save ข้อมูล)
func RenderPreviewWatermark(pdf *gofpdf.Fpdf, fontFamily string) {
	pageWidth, pageHeight := pdf.GetPageSize()

	// ตั้งค่า alpha ให้โปร่งใส (0.15 = 15% opacity)
	pdf.SetAlpha(0.15, "Normal")
	pdf.SetFont(fontFamily, "B", 72) // ตัวใหญ่ 72pt
	pdf.SetTextColor(128, 128, 128)  // สีเทา

	// คำนวณตำแหน่งกลางหน้า
	centerX := pageWidth / 2
	centerY := pageHeight / 2

	text := "Preview"
	textWidth := pdf.GetStringWidth(text)

	// วาด watermark ที่กึ่งกลางหน้า เอียง 45 องศา
	pdf.TransformBegin()
	pdf.TransformRotate(-45, centerX, centerY) // หมุน -45 องศา (เอียงซ้าย)
	pdf.SetXY(centerX-textWidth/2, centerY-15)
	pdf.CellFormat(textWidth, 30, text, "", 0, "C", false, 0, "")
	pdf.TransformEnd()

	// Reset alpha กลับเป็นปกติ
	pdf.SetAlpha(1.0, "Normal")
}

// RenderDocHeader - แสดงหัวเอกสาร (เลือก layout ตาม template)
func RenderDocHeader(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	switch template.HeaderLayout {
	case "centered":
		renderHeaderCentered(pdf, doc, payload, fontFamily, contentWidth, theme, fontSizes, template)
	case "minimal":
		renderHeaderMinimal(pdf, doc, payload, fontFamily, contentWidth, theme, fontSizes, template)
	case "split":
		renderHeaderSplit(pdf, doc, payload, fontFamily, contentWidth, theme, fontSizes, template)
	case "banner":
		renderHeaderBanner(pdf, doc, payload, fontFamily, contentWidth, theme, fontSizes, template)
	case "leftAligned":
		renderHeaderLeftAligned(pdf, doc, payload, fontFamily, contentWidth, theme, fontSizes, template)
	default: // "standard"
		renderHeaderStandard(pdf, doc, payload, fontFamily, contentWidth, theme, fontSizes, template)
	}
}

// renderHeaderStandard - Layout มาตรฐาน: Title/DocNo/Date ชิดขวา (กระชับ)
func renderHeaderStandard(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)

	// Document type badge (top right)
	pdf.SetFont(fontFamily, "B", headerFontSize+4)
	r, g, b := HexToRGB(theme.Header.TitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 6, payload.Title, "", 1, "R", false, 0, "")

	// Document number and date on same line
	docNo := GetStringValue(doc, "docno")
	docDate := FormatDateWithFormat(doc["docdatetime"], payload.DateFormat)
	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Header.SubtitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 4, fmt.Sprintf("%s %s  %s %s", labels.DocNo, docNo, labels.Date, docDate), "", 1, "R", false, 0, "")

	pdf.Ln(2)
	renderHeaderLine(pdf, contentWidth, theme, template)
	pdf.Ln(template.HeaderSpacing)
}

// renderHeaderCentered - Layout กึ่งกลาง: ข้อมูลทั้งหมดจัดกลาง (กระชับ)
func renderHeaderCentered(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)

	// Document title (center)
	pdf.SetFont(fontFamily, "B", headerFontSize+5)
	r, g, b := HexToRGB(theme.Header.TitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 7, payload.Title, "", 1, "C", false, 0, "")

	// Document number and date (center)
	docNo := GetStringValue(doc, "docno")
	docDate := FormatDateWithFormat(doc["docdatetime"], payload.DateFormat)
	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Header.SubtitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 4, fmt.Sprintf("%s %s   %s %s", labels.DocNo, docNo, labels.Date, docDate), "", 1, "C", false, 0, "")

	pdf.Ln(2)
	renderHeaderLine(pdf, contentWidth, theme, template)
	pdf.Ln(template.HeaderSpacing)
}

// renderHeaderMinimal - Layout ย่อ: บรรทัดเดียว (กระชับที่สุด)
func renderHeaderMinimal(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)

	docNo := GetStringValue(doc, "docno")
	docDate := FormatDateWithFormat(doc["docdatetime"], payload.DateFormat)

	// Single line: Title | DocNo | Date
	pdf.SetFont(fontFamily, "B", headerFontSize+1)
	r, g, b := HexToRGB(theme.Header.TitleColor)
	pdf.SetTextColor(r, g, b)
	text := fmt.Sprintf("%s | %s %s | %s %s", payload.Title, labels.DocNo, docNo, labels.Date, docDate)
	pdf.CellFormat(contentWidth, 5, text, "", 1, "L", false, 0, "")

	pdf.Ln(1)
	renderHeaderLine(pdf, contentWidth, theme, template)
	pdf.Ln(template.HeaderSpacing)
}

// renderHeaderSplit - Layout แบ่ง 2 ส่วน: ซ้าย Title, ขวา DocNo/Date (กระชับ)
func renderHeaderSplit(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)
	startY := pdf.GetY()
	halfWidth := contentWidth / 2

	// Left side: Document Title
	pdf.SetFont(fontFamily, "B", headerFontSize+4)
	r, g, b := HexToRGB(theme.Header.TitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(halfWidth, 6, payload.Title, "", 0, "L", false, 0, "")

	// Right side: DocNo and Date
	docNo := GetStringValue(doc, "docno")
	docDate := FormatDateWithFormat(doc["docdatetime"], payload.DateFormat)
	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Header.SubtitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(halfWidth, 6, fmt.Sprintf("%s %s  %s %s", labels.DocNo, docNo, labels.Date, docDate), "", 1, "R", false, 0, "")

	pdf.SetY(startY + 7)
	pdf.Ln(1)
	renderHeaderLine(pdf, contentWidth, theme, template)
	pdf.Ln(template.HeaderSpacing)
}

// renderHeaderBanner - Layout แถบสี: แถบสี primary เต็มความกว้าง (กระชับ)
func renderHeaderBanner(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)
	startY := pdf.GetY()

	// Draw banner background
	r, g, b := HexToRGB(theme.Colors.Primary)
	pdf.SetFillColor(r, g, b)
	pdf.Rect(template.MarginLeft, startY, contentWidth, 8, "F")

	// Document title on banner (white text)
	pdf.SetFont(fontFamily, "B", headerFontSize+3)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(template.MarginLeft+2, startY+1)
	pdf.CellFormat(contentWidth-4, 6, payload.Title, "", 0, "L", false, 0, "")

	// Move below banner
	pdf.SetY(startY + 9)

	// Document number and date below banner
	docNo := GetStringValue(doc, "docno")
	docDate := FormatDateWithFormat(doc["docdatetime"], payload.DateFormat)
	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Header.SubtitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 4, fmt.Sprintf("%s %s  %s %s", labels.DocNo, docNo, labels.Date, docDate), "", 1, "L", false, 0, "")

	pdf.Ln(template.HeaderSpacing)
}

// renderHeaderLeftAligned - Layout ชิดซ้าย: ข้อมูลทั้งหมดชิดซ้าย (กระชับ)
func renderHeaderLeftAligned(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)

	// Document title (left)
	pdf.SetFont(fontFamily, "B", headerFontSize+4)
	r, g, b := HexToRGB(theme.Header.TitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 6, payload.Title, "", 1, "L", false, 0, "")

	// Document number and date
	docNo := GetStringValue(doc, "docno")
	docDate := FormatDateWithFormat(doc["docdatetime"], payload.DateFormat)
	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Header.SubtitleColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentWidth, 4, fmt.Sprintf("%s %s  %s %s", labels.DocNo, docNo, labels.Date, docDate), "", 1, "L", false, 0, "")

	pdf.Ln(2)
	renderHeaderLine(pdf, contentWidth, theme, template)
	pdf.Ln(template.HeaderSpacing)
}

// renderHeaderLine - วาดเส้นใต้ header (ใช้ร่วมกัน)
func renderHeaderLine(pdf *gofpdf.Fpdf, contentWidth float64, theme Theme, template Template) {
	if template.ShowHeaderLine {
		r, g, b := HexToRGB(theme.Header.LineColor)
		pdf.SetDrawColor(r, g, b)
		lineWidth := theme.Header.LineWidth
		if template.HeaderLineWidth > 0 {
			lineWidth = template.HeaderLineWidth
		}
		pdf.SetLineWidth(lineWidth)
		pdf.Line(template.MarginLeft, pdf.GetY(), template.MarginLeft+contentWidth, pdf.GetY())
	}
}

// RenderCompanyCustomer - แสดงข้อมูลบริษัทและลูกค้าแบบ 2 คอลัมน์
func RenderCompanyCustomer(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	headerFontSize := float64(fontSizes.Header)
	labels := GetLabels(payload.Language)
	startY := pdf.GetY()
	colWidth := contentWidth/2 - 5

	// === LEFT COLUMN: Company Info ===
	pdf.SetXY(15, startY)

	// Company section header
	pdf.SetFont(fontFamily, "B", headerFontSize+1)
	r, g, b := HexToRGB(theme.Section.LabelColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(colWidth, 6, labels.SellerInfo, "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Section.TextColor)
	pdf.SetTextColor(r, g, b)

	// Company name (placeholder - can be from config)
	pdf.SetX(15)
	pdf.CellFormat(colWidth, 5, labels.CompanyName, "", 1, "L", false, 0, "")
	pdf.SetX(15)
	pdf.CellFormat(colWidth, 5, "123 ถนนสุขุมวิท แขวงคลองเตย", "", 1, "L", false, 0, "")
	pdf.SetX(15)
	pdf.CellFormat(colWidth, 5, "เขตคลองเตย กรุงเทพฯ 10110", "", 1, "L", false, 0, "")
	pdf.SetX(15)
	pdf.CellFormat(colWidth, 5, fmt.Sprintf("%s 02-xxx-xxxx", labels.Phone), "", 1, "L", false, 0, "")

	leftEndY := pdf.GetY()

	// === RIGHT COLUMN: Customer Info ===
	pdf.SetXY(15+colWidth+10, startY)

	// Customer section header
	pdf.SetFont(fontFamily, "B", headerFontSize+1)
	r, g, b = HexToRGB(theme.Section.LabelColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(colWidth, 6, labels.CustomerInfo, "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", headerFontSize)
	r, g, b = HexToRGB(theme.Section.TextColor)
	pdf.SetTextColor(r, g, b)

	// Customer code
	custCode := GetStringValue(doc, "custcode")
	pdf.SetX(15 + colWidth + 10)
	pdf.CellFormat(colWidth, 5, fmt.Sprintf("%s %s", labels.Code, custCode), "", 1, "L", false, 0, "")

	// Customer name from custnames array
	custName := ""
	if custNames, ok := doc["custnames"].(primitive.A); ok && len(custNames) > 0 {
		if nameObj, ok := custNames[0].(map[string]interface{}); ok {
			custName = GetStringValue(nameObj, "name")
		}
	}
	pdf.SetX(15 + colWidth + 10)
	pdf.MultiCell(colWidth, 5, custName, "", "L", false)

	// Customer address
	custAddress := GetStringValue(doc, "fullvataddress")
	if custAddress != "" {
		pdf.SetX(15 + colWidth + 10)
		pdf.MultiCell(colWidth, 5, custAddress, "", "L", false)
	}

	// Customer phone
	custPhone := GetStringValue(doc, "customertelephone")
	if custPhone != "" {
		pdf.SetX(15 + colWidth + 10)
		pdf.CellFormat(colWidth, 5, fmt.Sprintf("%s %s", labels.Phone, custPhone), "", 1, "L", false, 0, "")
	}

	rightEndY := pdf.GetY()

	// Set Y to the lower of the two columns
	if leftEndY > rightEndY {
		pdf.SetY(leftEndY)
	} else {
		pdf.SetY(rightEndY)
	}

	pdf.Ln(8)
}

// RenderDetailsTable - แสดงตารางรายละเอียดสินค้า (รองรับหลาย styles)
func RenderDetailsTable(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, orientation string, theme Theme, fontSizes FontSizes, template Template) {
	// Get table style configuration
	styleConfig := GetTableStyleConfig(template.TableStyle)
	detailFontSize := float64(fontSizes.Detail)
	labels := GetLabels(payload.Language)

	// Column widths as percentages of contentWidth (total = 100%)
	var colWidthsPercent []float64
	if orientation == "L" {
		colWidthsPercent = []float64{5, 10, 50, 8, 8, 10, 9}
	} else {
		colWidthsPercent = []float64{5, 12, 38, 9, 10, 12, 14}
	}

	// Calculate actual widths from percentages
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

	// Set header colors based on style
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

	headers := []string{labels.No, labels.ItemCode, labels.Item, labels.Unit, labels.Qty, labels.UnitPrice, labels.Amount}
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
			lineNum := GetIntValue(detailMap, "line_number")
			sortedDetails = append(sortedDetails, detailItem{lineNumber: lineNum, data: detailMap})
		}
	}
	sort.Slice(sortedDetails, func(i, j int) bool {
		return sortedDetails[i].lineNumber < sortedDetails[j].lineNumber
	})

	var grandTotal float64
	var totalQty float64
	// คำนวณ lineHeight จาก LineSpacing (default 1.0)
	lineSpacing := payload.LineSpacing
	if lineSpacing <= 0 {
		lineSpacing = 1.0 // default
	}
	lineHeight := 4.0 * lineSpacing // Height per line of text (ปรับตาม lineSpacing)

	for idx, item := range sortedDetails {
		detail := item.data

		// Get item name from itemnames array
		itemName := ""
		if itemNames, ok := detail["itemnames"].(primitive.A); ok && len(itemNames) > 0 {
			if nameObj, ok := itemNames[0].(map[string]interface{}); ok {
				itemName = GetStringValue(nameObj, "name")
			}
		}

		itemCode := GetStringValue(detail, "itemcode")
		unitCode := GetStringValue(detail, "unitcode")
		qty := GetFloatValue(detail, "qty")
		price := GetFloatValue(detail, "price")
		sumAmount := GetFloatValue(detail, "sum_amount")

		// Calculate number of lines needed for item name
		itemNameLines := WrapThaiText(pdf, itemName, colWidths[2]-2*styleConfig.CellPadding)
		numLines := len(itemNameLines)
		if numLines < 1 {
			numLines = 1
		}

		// Calculate row height based on content
		rowHeight := float64(numLines) * lineHeight
		if rowHeight < styleConfig.RowHeight {
			rowHeight = styleConfig.RowHeight
		}

		// Alternate row colors (conditional based on template and style)
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

		// Save current Y position
		startY := pdf.GetY()
		startX := template.MarginLeft

		// Draw cells with proper height
		// Column 0: ลำดับ
		pdf.SetXY(startX, startY)
		pdf.CellFormat(colWidths[0], rowHeight, fmt.Sprintf("%d", idx+1), styleConfig.RowBorder, 0, "C", styleConfig.RowFill, 0, "")

		// Column 1: รหัสสินค้า
		pdf.SetXY(startX+colWidths[0], startY)
		pdf.CellFormat(colWidths[1], rowHeight, itemCode, styleConfig.RowBorder, 0, "L", styleConfig.RowFill, 0, "")

		// Column 2: รายการ (with text wrapping)
		pdf.SetXY(startX+colWidths[0]+colWidths[1], startY)
		pdf.CellFormat(colWidths[2], rowHeight, "", styleConfig.RowBorder, 0, "L", styleConfig.RowFill, 0, "")
		// Draw text inside cell
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

		// Column 5: ราคา/หน่วย
		pdf.SetXY(startX+colWidths[0]+colWidths[1]+colWidths[2]+colWidths[3]+colWidths[4], startY)
		pdf.CellFormat(colWidths[5], rowHeight, FormatNumber(price, 2), styleConfig.RowBorder, 0, "R", styleConfig.RowFill, 0, "")

		// Column 6: จำนวนเงิน
		pdf.SetXY(startX+colWidths[0]+colWidths[1]+colWidths[2]+colWidths[3]+colWidths[4]+colWidths[5], startY)
		pdf.CellFormat(colWidths[6], rowHeight, FormatNumber(sumAmount, 2), styleConfig.RowBorder, 0, "R", styleConfig.RowFill, 0, "")

		// Move to next row
		pdf.SetY(startY + rowHeight)

		grandTotal += sumAmount
		totalQty += qty
	}

	// Bottom border for table (if style requires)
	if styleConfig.ShowBottomBorder {
		r, g, b = HexToRGB(theme.Colors.Primary)
		pdf.SetDrawColor(r, g, b)
		pdf.SetLineWidth(styleConfig.BorderWidth)
		totalTableWidth := colWidths[0] + colWidths[1] + colWidths[2] + colWidths[3] + colWidths[4] + colWidths[5] + colWidths[6]
		pdf.Line(template.MarginLeft, pdf.GetY(), template.MarginLeft+totalTableWidth, pdf.GetY())
	}

	pdf.Ln(template.TableSpacing)
}

// RenderSummarySignature - แสดงส่วนสรุป (รองรับหลาย layouts)
func RenderSummarySignature(pdf *gofpdf.Fpdf, doc map[string]interface{}, payload GenPDFPayload, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template) {
	// Calculate totals from document
	totalAmount := GetFloatValue(doc, "total_amount")
	totalDiscount := GetFloatValue(doc, "totaldiscount")
	totalVat := GetFloatValue(doc, "totalvatvalue")
	totalAfterVat := GetFloatValue(doc, "totalaftervat")
	labels := GetLabels(payload.Language)

	// If no totals in header, calculate from details
	if totalAmount == 0 {
		if details, ok := doc["details"].(primitive.A); ok {
			for _, d := range details {
				if detailMap, ok := d.(map[string]interface{}); ok {
					totalAmount += GetFloatValue(detailMap, "sum_amount")
				}
			}
		}
	}

	// Create summary data
	summaryData := SummaryData{
		TotalAmount:   totalAmount,
		TotalDiscount: totalDiscount,
		TotalVat:      totalVat,
		TotalAfterVat: totalAfterVat,
	}

	// Render based on layout
	switch template.SummaryLayout {
	case "full":
		renderSummaryFull(pdf, summaryData, fontFamily, contentWidth, theme, fontSizes, template, labels)
	case "boxed":
		renderSummaryBoxed(pdf, summaryData, fontFamily, contentWidth, theme, fontSizes, template, labels)
	case "minimal":
		renderSummaryMinimal(pdf, summaryData, fontFamily, contentWidth, theme, fontSizes, template, labels)
	case "twoColumn":
		renderSummaryTwoColumn(pdf, doc, summaryData, fontFamily, contentWidth, theme, fontSizes, template, labels)
	case "highlighted":
		renderSummaryHighlighted(pdf, summaryData, fontFamily, contentWidth, theme, fontSizes, template, labels)
	default: // "right"
		renderSummaryRight(pdf, summaryData, fontFamily, contentWidth, theme, fontSizes, template, labels)
	}
}

// SummaryData - ข้อมูลสรุปยอด
type SummaryData struct {
	TotalAmount   float64
	TotalDiscount float64
	TotalVat      float64
	TotalAfterVat float64
}

// renderSummaryRight - Layout ชิดขวา (default)
func renderSummaryRight(pdf *gofpdf.Fpdf, data SummaryData, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)
	summaryWidth := 80.0
	pageWidth, _ := pdf.GetPageSize()
	_, _, rightMargin, _ := pdf.GetMargins()
	summaryX := pageWidth - rightMargin - summaryWidth

	labelWidth := 45.0
	valueWidth := 35.0
	rowHeight := 5.0

	renderSummaryItems(pdf, data, fontFamily, summaryX, labelWidth, valueWidth, rowHeight, summaryFontSize, theme, template, labels)
}

// renderSummaryFull - Layout เต็มความกว้าง
func renderSummaryFull(pdf *gofpdf.Fpdf, data SummaryData, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)
	summaryX := template.MarginLeft
	summaryWidth := contentWidth

	labelWidth := summaryWidth * 0.6
	valueWidth := summaryWidth * 0.4
	rowHeight := 5.0

	renderSummaryItems(pdf, data, fontFamily, summaryX, labelWidth, valueWidth, rowHeight, summaryFontSize, theme, template, labels)
}

// renderSummaryBoxed - Layout มีกรอบ
func renderSummaryBoxed(pdf *gofpdf.Fpdf, data SummaryData, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)
	summaryWidth := 85.0
	pageWidth, _ := pdf.GetPageSize()
	_, _, rightMargin, _ := pdf.GetMargins()
	summaryX := pageWidth - rightMargin - summaryWidth

	startY := pdf.GetY()

	// Draw box background
	r, g, b := HexToRGB(theme.Summary.BgColor)
	pdf.SetFillColor(r, g, b)
	r, g, b = HexToRGB(theme.Summary.BorderColor)
	pdf.SetDrawColor(r, g, b)
	pdf.SetLineWidth(0.3)

	// Calculate box height
	numItems := 1 // always total
	if template.ShowSubtotal {
		numItems++
	}
	if template.ShowTotalDiscount && data.TotalDiscount > 0 {
		numItems += 2
	}
	if template.ShowTotalTax {
		numItems++
	}
	boxHeight := float64(numItems)*5.0 + 4

	pdf.Rect(summaryX-2, startY, summaryWidth+4, boxHeight, "FD")

	labelWidth := 48.0
	valueWidth := 35.0
	rowHeight := 5.0

	pdf.SetY(startY + 2)
	renderSummaryItems(pdf, data, fontFamily, summaryX, labelWidth, valueWidth, rowHeight, summaryFontSize, theme, template, labels)
}

// renderSummaryMinimal - Layout ย่อ แสดงเฉพาะยอดรวม
func renderSummaryMinimal(pdf *gofpdf.Fpdf, data SummaryData, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)
	pageWidth, _ := pdf.GetPageSize()
	_, _, rightMargin, _ := pdf.GetMargins()

	// Only show total
	pdf.SetFont(fontFamily, "B", summaryFontSize+1)
	r, g, b := HexToRGB(theme.Summary.TextColor)
	pdf.SetTextColor(r, g, b)

	totalText := fmt.Sprintf("%s: %s %s", labels.GrandTotal, FormatNumber(data.TotalAfterVat, 2), labels.Currency)
	pdf.SetX(template.MarginLeft)
	pdf.CellFormat(pageWidth-template.MarginLeft-rightMargin, 6, totalText, "", 1, "R", false, 0, "")

	// Amount in words
	if data.TotalAfterVat > 0 {
		pdf.SetFont(fontFamily, "", summaryFontSize-1)
		pdf.SetX(template.MarginLeft)
		pdf.CellFormat(pageWidth-template.MarginLeft-rightMargin, 4, fmt.Sprintf("(%s)", NumberToThaiText(data.TotalAfterVat)), "", 1, "R", false, 0, "")
	}
}

// renderSummaryTwoColumn - Layout 2 คอลัมน์ (หมายเหตุ + สรุป)
func renderSummaryTwoColumn(pdf *gofpdf.Fpdf, doc map[string]interface{}, data SummaryData, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)
	startY := pdf.GetY()
	colWidth := contentWidth / 2

	// Left column: Notes
	pdf.SetXY(template.MarginLeft, startY)
	pdf.SetFont(fontFamily, "B", summaryFontSize)
	r, g, b := HexToRGB(theme.Section.LabelColor)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(colWidth-5, 5, labels.Notes, "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", summaryFontSize-1)
	r, g, b = HexToRGB(theme.Section.TextColor)
	pdf.SetTextColor(r, g, b)
	notes := GetStringValue(doc, "remark")
	if notes == "" {
		notes = "-"
	}
	pdf.SetX(template.MarginLeft)
	pdf.MultiCell(colWidth-5, 4, notes, "", "L", false)

	leftEndY := pdf.GetY()

	// Right column: Summary
	pdf.SetY(startY)
	summaryX := template.MarginLeft + colWidth + 5
	labelWidth := 45.0
	valueWidth := colWidth - 55
	rowHeight := 5.0

	renderSummaryItems(pdf, data, fontFamily, summaryX, labelWidth, valueWidth, rowHeight, summaryFontSize, theme, template, labels)

	rightEndY := pdf.GetY()

	// Set Y to lower position
	if leftEndY > rightEndY {
		pdf.SetY(leftEndY)
	} else {
		pdf.SetY(rightEndY)
	}
}

// renderSummaryHighlighted - Layout เน้นยอดรวม
func renderSummaryHighlighted(pdf *gofpdf.Fpdf, data SummaryData, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	summaryFontSize := float64(fontSizes.Summary)
	summaryWidth := 85.0
	pageWidth, _ := pdf.GetPageSize()
	_, _, rightMargin, _ := pdf.GetMargins()
	summaryX := pageWidth - rightMargin - summaryWidth

	labelWidth := 48.0
	valueWidth := 35.0
	rowHeight := 5.0

	// Force highlight for all styles
	templateCopy := template
	templateCopy.HighlightTotal = true

	renderSummaryItems(pdf, data, fontFamily, summaryX, labelWidth, valueWidth, rowHeight, summaryFontSize, theme, templateCopy, labels)
}

// renderSummaryItems - ฟังก์ชันร่วมสำหรับวาด summary items
func renderSummaryItems(pdf *gofpdf.Fpdf, data SummaryData, fontFamily string, summaryX, labelWidth, valueWidth, rowHeight, fontSize float64, theme Theme, template Template, labels Labels) {
	var r, g, b int

	// Build summary items based on template settings
	type summaryItem struct {
		label       string
		value       float64
		isHighlight bool
	}
	var items []summaryItem

	if template.ShowSubtotal {
		items = append(items, summaryItem{labels.Subtotal, data.TotalAmount, false})
	}
	if template.ShowTotalDiscount && data.TotalDiscount > 0 {
		items = append(items, summaryItem{labels.Discount, data.TotalDiscount, false})
		items = append(items, summaryItem{labels.AfterDiscount, data.TotalAmount - data.TotalDiscount, false})
	}
	if template.ShowTotalTax {
		items = append(items, summaryItem{labels.Tax, data.TotalVat, false})
	}
	items = append(items, summaryItem{labels.GrandTotal, data.TotalAfterVat, template.HighlightTotal})

	for _, item := range items {
		pdf.SetX(summaryX)

		if item.isHighlight {
			pdf.SetFont(fontFamily, "B", fontSize+1)
			r, g, b = HexToRGB(theme.Summary.HighlightBgColor)
			pdf.SetFillColor(r, g, b)
			r, g, b = HexToRGB(theme.Summary.HighlightTextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetDrawColor(r, g, b)
			pdf.CellFormat(labelWidth, rowHeight+1, item.label, "1", 0, "R", true, 0, "")
			pdf.CellFormat(valueWidth, rowHeight+1, FormatNumber(item.value, 2), "1", 1, "R", true, 0, "")
		} else {
			pdf.SetFont(fontFamily, "", fontSize)
			r, g, b = HexToRGB(theme.Summary.BgColor)
			pdf.SetFillColor(r, g, b)
			r, g, b = HexToRGB(theme.Summary.BorderColor)
			pdf.SetDrawColor(r, g, b)
			r, g, b = HexToRGB(theme.Summary.TextColor)
			pdf.SetTextColor(r, g, b)
			pdf.CellFormat(labelWidth, rowHeight, item.label, "1", 0, "R", true, 0, "")
			pdf.CellFormat(valueWidth, rowHeight, FormatNumber(item.value, 2), "1", 1, "R", true, 0, "")
		}
	}

	// Amount in words (Thai Baht)
	pdf.Ln(1)
	pdf.SetX(summaryX)
	pdf.SetFont(fontFamily, "", fontSize-1)
	r, g, b = HexToRGB(theme.Summary.TextColor)
	pdf.SetTextColor(r, g, b)
	if data.TotalAfterVat > 0 {
		pdf.MultiCell(labelWidth+valueWidth, 3, fmt.Sprintf("(%s)", NumberToThaiText(data.TotalAfterVat)), "", "R", false)
	}

	pdf.SetTextColor(0, 0, 0)
}

// setupFooter - ตั้งค่า footer ตาม style
func setupFooter(pdf *gofpdf.Fpdf, doc map[string]interface{}, fontFamily string, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	footerFontSize := float64(fontSizes.Footer)

	pdf.SetFooterFunc(func() {
		pageWidth, _ := pdf.GetPageSize()
		contentWidth := pageWidth - template.MarginLeft - template.MarginRight

		switch template.FooterStyle {
		case "detailed":
			// ซ้าย: วันที่พิมพ์, ขวา: หน้า
			pdf.SetY(-10)
			pdf.SetFont(fontFamily, "", footerFontSize)
			r, g, b := HexToRGB(theme.Footer.TextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth/2, 5, fmt.Sprintf("%s %s", labels.PrintedAt, GetCurrentDateTime()), "", 0, "L", false, 0, "")
			pdf.CellFormat(contentWidth/2, 5, fmt.Sprintf("%s %d", labels.Page, pdf.PageNo()), "", 0, "R", false, 0, "")

		case "minimal":
			// เลขหน้าชิดขวา
			pdf.SetY(-8)
			pdf.SetFont(fontFamily, "", footerFontSize)
			r, g, b := HexToRGB(theme.Footer.TextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth, 4, fmt.Sprintf("%d", pdf.PageNo()), "", 0, "R", false, 0, "")

		case "branded":
			// กลาง: ชื่อบริษัท, ล่าง: เลขหน้า
			pdf.SetY(-12)
			pdf.SetFont(fontFamily, "B", footerFontSize)
			r, g, b := HexToRGB(theme.Footer.TextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth, 4, labels.CompanyName, "", 1, "C", false, 0, "")
			pdf.SetFont(fontFamily, "", footerFontSize-1)
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth, 4, fmt.Sprintf("%s %d", labels.Page, pdf.PageNo()), "", 0, "C", false, 0, "")

		case "withTerms":
			// แสดงเงื่อนไข + เลขหน้า
			pdf.SetY(-15)
			pdf.SetFont(fontFamily, "", footerFontSize-1)
			r, g, b := HexToRGB(theme.Footer.TextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth, 3, labels.Terms, "", 1, "L", false, 0, "")
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth/2, 4, fmt.Sprintf("%s %s", labels.PrintedAt, GetCurrentDateTime()), "", 0, "L", false, 0, "")
			pdf.CellFormat(contentWidth/2, 4, fmt.Sprintf("%s %d", labels.Page, pdf.PageNo()), "", 0, "R", false, 0, "")

		default: // "standard"
			// เลขหน้าอยู่กลาง
			pdf.SetY(-10)
			pdf.SetFont(fontFamily, "", footerFontSize)
			r, g, b := HexToRGB(theme.Footer.TextColor)
			pdf.SetTextColor(r, g, b)
			pdf.SetX(template.MarginLeft)
			pdf.CellFormat(contentWidth, 5, fmt.Sprintf("%s %d", labels.Page, pdf.PageNo()), "", 0, "C", false, 0, "")
		}
	})
}

// RenderSignatureSection - แสดงส่วนลงนาม (เรียกหลัง summary ถ้า template.ShowSignature = true)
func RenderSignatureSection(pdf *gofpdf.Fpdf, fontFamily string, contentWidth float64, theme Theme, fontSizes FontSizes, template Template, labels Labels) {
	if !template.ShowSignature {
		return
	}

	pdf.Ln(5)
	fontSize := float64(fontSizes.Footer)
	pdf.SetFont(fontFamily, "", fontSize)
	r, g, b := HexToRGB(theme.Section.TextColor)
	pdf.SetTextColor(r, g, b)

	// 3 signature boxes
	boxWidth := contentWidth / 3
	startX := template.MarginLeft
	startY := pdf.GetY()

	signatures := []string{labels.PreparedBy, labels.CheckedBy, labels.ApprovedBy}

	for i, label := range signatures {
		x := startX + float64(i)*boxWidth

		// Draw signature line
		pdf.SetDrawColor(100, 100, 100)
		pdf.SetLineWidth(0.2)
		lineY := startY + 15
		pdf.Line(x+10, lineY, x+boxWidth-10, lineY)

		// Draw label
		pdf.SetXY(x, lineY+2)
		pdf.CellFormat(boxWidth, 4, label, "", 0, "C", false, 0, "")
	}

	pdf.SetY(startY + 25)
}
