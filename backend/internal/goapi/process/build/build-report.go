package build

import (
	"fmt"
	"os"
	"path/filepath"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"smlcloudplatform/internal/goapi/myglobal"
)

// ฟังก์ชันเพิ่ม \n อัตโนมัติเมื่อข้อความกว้างเกินคอลัมน์
func AddLineBreaks(pdf *gofpdf.Fpdf, text string, maxWidth float64) string {
	// ถ้ามี \n อยู่แล้ว ให้แยกตามนั้นก่อน
	paragraphs := strings.Split(text, "\n")
	result := []string{}

	for _, paragraph := range paragraphs {
		words := strings.Split(paragraph, " ")
		currentLine := ""

		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if pdf.GetStringWidth(testLine) <= maxWidth {
				currentLine = testLine
			} else {
				if currentLine != "" {
					result = append(result, currentLine)
				}
				currentLine = word
			}
		}

		if currentLine != "" {
			result = append(result, currentLine)
		}
	}

	// รวมบรรทัดด้วย \n
	return strings.Join(result, "\n")
}

func BuildReport(guid string, reportName string, report models.ReportModel, reportStyle models.ReportStyleModel) string {
	// ตรวจสอบ input parameters
	if guid == "" || reportName == "" {
		logger.Error("Empty guid or reportName")
		return ""
	}

	// ใช้ system temp directory และสร้าง subfolder tmpdede
	tempDir := os.TempDir()

	// สร้าง tmpdede directory ใน system temp ถ้ายังไม่มี
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.MkdirAll(tempDir, 0755)
		if err != nil {
			logger.Error("creating tmpdede directory: %v", err)
			return ""
		}
	}

	buildPdf := func() {
		leftMargin := 10.0
		rightMargin := 10.0
		topMargin := 10.0
		bottomMargin := 10.0
		pageNumber := 0
		// คำนวณความกว้าง Column ตามเปอร์เซ็นต์
		for indexRow, headerRow := range report.HeaderRows {
			totalWidth := 0.0
			for _, column := range headerRow.Columns {
				totalWidth += column.Width
			}
			for indexColumn, column := range headerRow.Columns {
				// คำนวณความกว้างของ Column ตามเปอร์เซ็นต์
				report.HeaderRows[indexRow].Columns[indexColumn].WidthCalc = (column.Width * (100 - headerRow.LeftMarginPercent)) / totalWidth
			}
		}
		// ลบ file ใน folder tmpdede ทั้งหมด
		errx := myglobal.DeleteFilesInFolder(tempDir, reportName)
		if errx != nil {
			logger.Error("delete files: %v", errx)
			return
		}

		pdf := gofpdf.New(reportStyle.PaperType, "mm", "A4", "")
		if pdf == nil {
			logger.Error("Failed to create PDF")
			return
		}

		pdf.SetMargins(0, 0, 0) // ตั้งค่า margin ซ้าย, บน, ขวา เป็น 0
		pdf.SetAutoPageBreak(true, 0)
		pdf.SetLeftMargin(leftMargin)
		pdf.SetRightMargin(rightMargin)

		// โหลดฟอนต์ภาษาไทย - เพิ่ม error handling
		if _, err := os.Stat("fonts/NotoSansThai-Light.ttf"); err == nil {
			pdf.AddUTF8Font("noto", "", "fonts/NotoSansThai-Light.ttf")
		} else {
			logger.Warn("Font file not found: fonts/NotoSansThai-Light.ttf")
		}

		if _, err := os.Stat("fonts/NotoSansThai-Bold.ttf"); err == nil {
			pdf.AddUTF8Font("noto", "B", "fonts/NotoSansThai-Bold.ttf")
		} else {
			logger.Warn("Font file not found: fonts/NotoSansThai-Bold.ttf")
		}

		// ดึงความกว้างกระดาษ
		pageWidthMM, pageHeightMM := pdf.GetPageSize()
		pageWidthMM -= (leftMargin + rightMargin)
		logger.Info("Page size: %f x %f", pageWidthMM, pageHeightMM)

		// ตำแหน่งปัจจุบันในหน้า
		currentHeight := topMargin

		printHeader := func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("Panic in printHeader: %v", r)
				}
			}()

			pageNumber = pageNumber + 1
			currentHeight = topMargin
			pdf.AddPage()
			{
				// top margin
				pdf.SetY(currentHeight)
			}
			{
				// Name - เพิ่มการตรวจสอบ
				if report.Name != "" {
					pdf.SetFont("noto", "B", 10)

					// วัดความสูงของข้อความ
					lineHeight := pdf.PointToUnitConvert(16) // แปลงขนาดฟอนต์เป็นหน่วย mm

					// แสดง header
					pdf.CellFormat(pageWidthMM, lineHeight, report.Name, "", 1, "C", false, 0, "")

					currentHeight += lineHeight
				}
			}
			{
				// Header
				pdf.SetFont("noto", "", 8)

				// วัดความสูงของข้อความ
				lineHeight := pdf.PointToUnitConvert(16) // แปลงขนาดฟอนต์เป็นหน่วย mm

				// แสดง header
				pdf.CellFormat(pageWidthMM, lineHeight, report.Header, "", 1, "C", false, 0, "")

				// วันที่พิมพ์
				pdf.SetFont("noto", "", 6)
				pdf.SetY(currentHeight)
				pdf.SetX(leftMargin)
				pdf.CellFormat(20, lineHeight, fmt.Sprintf("วันที่พิมพ์ : %s", myglobal.FormatDateFullThai(time.Now())), "", 0, "L", false, 0, "")

				// แสดงเลขหน้า
				pdf.SetFont("noto", "", 6)
				pdf.SetY(currentHeight)
				pdf.SetX(pageWidthMM - rightMargin)
				pdf.CellFormat(20, lineHeight, fmt.Sprintf("หน้า %d", pageNumber), "", 0, "R", false, 0, "")

				// คำนวณความสูงที่ใช้จริง (รวมระยะห่างเพิ่มเติม)
				currentHeight += lineHeight
			}
			{
				// วาดแถวหัวตาราง
				for _, headerRow := range report.HeaderRows {
					fontSize := 6
					if headerRow.FontSize > 0 {
						fontSize = int(headerRow.FontSize)
					}
					pdf.SetFont("noto", "", float64(fontSize))
					fontHeight := pdf.PointToUnitConvert(float64(fontSize)) // แก้ไข: เพิ่มการคำนวณ fontHeight

					// คำนวณความกว้างจริงของแต่ละคอลัมน์
					leftMarginRow := (headerRow.LeftMarginPercent * pageWidthMM) / 100
					columnWidths := make([]float64, len(headerRow.Columns))
					for i, column := range headerRow.Columns {
						columnWidths[i] = (column.WidthCalc * (pageWidthMM - leftMarginRow)) / 100
					}

					// 1. แบ่งข้อความและเพิ่ม \n เมื่อข้อความกว้างเกินคอลัมน์
					modifiedTexts := make([]string, len(headerRow.Columns))
					maxLines := 1

					for i, column := range headerRow.Columns {
						// แปลงข้อความให้มี \n เมื่อข้อความกว้างเกินคอลัมน์
						modifiedText := AddLineBreaks(pdf, column.Name, columnWidths[i])
						modifiedTexts[i] = modifiedText

						// นับจำนวนบรรทัดในข้อความที่แปลงแล้ว
						lineCount := 1 + strings.Count(modifiedText, "\n")
						if lineCount > maxLines {
							maxLines = lineCount
						}
					}

					// 2. คำนวณความสูงของแถวจากจำนวนบรรทัดสูงสุด
					rowHeight := float64(maxLines) * (fontHeight + (fontHeight / 2))

					// 3. วาดเส้นบนและล่างของแถว
					if headerRow.TopLine {
						pdf.Line(leftMargin, currentHeight, leftMargin+pageWidthMM, currentHeight) // แก้ไข: เพิ่ม leftMargin
						currentHeight += fontHeight / 2
					}

					// 4. วาดข้อความในแต่ละคอลัมน์
					xStart := leftMargin
					for i, modifiedText := range modifiedTexts {
						colStartX := xStart
						colStartY := currentHeight

						// แยกข้อความตามบรรทัด
						lines := strings.Split(modifiedText, "\n")
						align := "L"
						switch headerRow.Columns[i].Align {
						case 1:
							align = "C"
						case 2:
							align = "R"
						}
						for j, line := range lines {
							lineY := colStartY + float64(j)*(fontHeight+2) // แก้ไข: ใช้ fontHeight ในการคำนวณ
							pdf.SetXY(colStartX+leftMarginRow, lineY)
							pdf.CellFormat(columnWidths[i], fontHeight, line, "", 0, align, false, 0, "") // แก้ไข: กำหนดความสูงของเซลล์
						}
						xStart += columnWidths[i]
					}

					if headerRow.BottomLine {
						// 5. วาดเส้นล่าง
						pdf.Line(leftMargin, currentHeight+rowHeight, leftMargin+pageWidthMM, currentHeight+rowHeight) // แก้ไข: เพิ่ม leftMargin
						rowHeight += (fontHeight / 2)
					}

					// อัปเดตตำแหน่งปัจจุบัน
					currentHeight += rowHeight
					pdf.SetY(currentHeight)
				}
			}
		}

		printFooter := func() {
			// วาดแถวท้ายรายงาน
			// ขีดเส้นบน
			yStart := pageHeightMM - bottomMargin
			pdf.Line(leftMargin, yStart, leftMargin+pageWidthMM, yStart) // แก้ไข: เพิ่ม leftMargin

			// ข้อความ
			pdf.SetFont("noto", "", 6)
			for _, footerRow := range report.FooterRows {
				yStart += 5
				pdf.SetY(yStart)
				xStart := 10.0
				for _, column := range footerRow.Columns {
					pdf.SetX(xStart)
					pdf.CellFormat(column.Width, 5, column.Name, "", 0, "L", false, 0, "")
					xStart += column.Width
				}
			}
		}

		printData := func() {
			// วาดข้อมูล
			for _, dataRow := range report.DataRows {
				rowIndex := dataRow.RowIndex
				pdf.SetY(currentHeight)
				headerRow := report.HeaderRows[rowIndex]
				leftMarginRow := (headerRow.LeftMarginPercent * pageWidthMM) / 100 // แก้ไข: ลบ leftMargin ออก
				xStart := leftMargin                                               // แก้ไข: เริ่มที่ leftMargin

				// คำนวณความกว้างจริงของแต่ละคอลัมน์
				columnWidths := make([]float64, len(headerRow.Columns))
				for i, column := range headerRow.Columns {
					columnWidths[i] = (column.WidthCalc * (pageWidthMM - leftMarginRow)) / 100
				}

				// แบ่งข้อความและเพิ่ม \n เมื่อข้อความกว้างเกินคอลัมน์
				modifiedTexts := make([]string, len(dataRow.ColumnData))
				maxLines := 1
				fontSize := 6
				if headerRow.FontSize > 0 {
					fontSize = int(headerRow.FontSize)
				}
				pdf.SetFont("noto", "", float64(fontSize))
				fontHeight := pdf.PointToUnitConvert(float64(fontSize)) // แก้ไข: เพิ่มการคำนวณ fontHeight

				for i, columnData := range dataRow.ColumnData {
					// แปลงข้อความให้มี \n เมื่อข้อความกว้างเกินคอลัมน์
					text := ""
					if columnData.Value != nil {
						switch columnData.Value.(type) {
						case time.Time:
							if columnData.DateTimeStyle == 1 {
								text = myglobal.FormatDateFull(columnData.Value.(time.Time))
							} else {
								text = myglobal.FormatDateFullThai(columnData.Value.(time.Time))
							}
						case float64:
							point := 2
							if columnData.Point > 0 {
								point = columnData.Point
							}
							text = myglobal.FormatNumber(columnData.Value.(float64), point)
						case int:
							text = myglobal.FormatNumber(float64(columnData.Value.(int)), 0)
						default:
							text = columnData.Value.(string)
						}
					}
					if i < len(columnWidths) { // ป้องกันข้อผิดพลาด index out of range
						modifiedText := AddLineBreaks(pdf, text, columnWidths[i])
						modifiedTexts[i] = modifiedText

						// นับจำนวนบรรทัดในข้อความที่แปลงแล้ว
						lineCount := 1 + strings.Count(modifiedText, "\n")
						if lineCount > maxLines {
							maxLines = lineCount
						}
					}
				}

				// คำนวณความสูงของแถวจากจำนวนบรรทัดสูงสุด
				rowHeight := float64(maxLines) * (fontHeight + (fontHeight / 2))

				if dataRow.TopLine {
					pdf.Line(leftMargin, currentHeight, leftMargin+pageWidthMM, currentHeight)
					currentHeight += (fontHeight / 2)
				}

				for i, modifiedText := range modifiedTexts {
					if i >= len(headerRow.Columns) { // ป้องกันข้อผิดพลาด index out of range
						continue
					}

					// ตรวจสอบเพื่อขึ้นหน้าใหม่
					if currentHeight+rowHeight > pageHeightMM-bottomMargin {
						printFooter()
						printHeader()
					}

					// แยกข้อความตามบรรทัด
					lines := strings.Split(modifiedText, "\n")
					align := "L"
					switch headerRow.Columns[i].Align {
					case 1:
						align = "C"
					case 2:
						align = "R"
					}
					if dataRow.ColumnData[i].Style == 1 {
						pdf.SetFont("noto", "B", float64(fontSize))
					} else {
						pdf.SetFont("noto", "", float64(fontSize))
					}
					for j, line := range lines {
						lineY := currentHeight + float64(j)*(fontHeight+(fontHeight/2))
						pdf.SetXY(xStart+leftMarginRow, lineY)
						pdf.CellFormat(columnWidths[i], fontHeight, line, "", 0, align, false, 0, "") // แก้ไข: ใช้ columnWidths[i] และกำหนดความสูงของเซลล์
					}

					xStart += columnWidths[i]
				}

				if dataRow.BottomLine {
					pdf.Line(leftMargin, currentHeight+rowHeight, leftMargin+pageWidthMM, currentHeight+rowHeight)
					rowHeight += (fontHeight / 2)
				}

				currentHeight += rowHeight // แก้ไข: ใช้ rowHeight ที่คำนวณแล้ว
			}
		}

		printHeader()
		printData()
		printFooter()

		// บันทึกไฟล์ PDF ใน tmpdede
		pdfPath := filepath.Join(tempDir, reportName+"-"+guid+".pdf")

		logger.Info("Writing PDF to: %s", pdfPath)
		err := pdf.OutputFileAndClose(pdfPath)
		if err != nil {
			logger.Error("writing PDF: %v", err)
			return
		}

		// save pdf to .bin file (zip)
		binPath := filepath.Join(tempDir, reportName+"-"+guid+".bin")
		err = myglobal.ConvertPDFToGzip(pdfPath, binPath)
		if err != nil {
			logger.Error("saving PDF to bin: %v", err)
			return
		}
	}

	logger.Info("Start Build PDF: %s for GUID %s", reportName, guid)

	buildPdf()
	// return absolute path สำหรับให้หาไฟล์เจอ
	return filepath.Join(tempDir, reportName+"-"+guid+".bin")
}
