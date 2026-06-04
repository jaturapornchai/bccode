package reportstock

import (
	"encoding/json"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/build"
	// "smlcloudplatform/internal/goapi/process"
)

func ReportProductBalanceByItemAndWareHouseAndLocation(holdingCode string, guid string, condition int, finalDate string, timezoneCode string, languageCode string) string {
	// เพิ่ม panic recovery
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		logger.Info("PDF Generation panic recovered for GUID %s: %v", guid, r)
	// 	}
	// }()

	// ตรวจสอบ parameters
	if holdingCode == "" || guid == "" {
		logger.Info("Invalid parameters: holdingCode=%s, guid=%s", holdingCode, guid)
		return ""
	}

	logger.Info("Starting PDF generation for GUID: %s, Shop: %s", guid, holdingCode)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL: %v", err)
		return ""
	}

	query := "SELECT datajson FROM result WHERE guid = $1 ORDER BY linenumber"
	logger.Debug("query: %s", query)

	rows, err := db.Query(query, guid)
	if err != nil {
		logger.Info("Failed to execute query: %v", err)
		return ""
	}
	defer rows.Close()

	var products = []models.ProductBalanceByCodeStruct{}
	for rows.Next() {
		var jsonData []byte
		err := rows.Scan(&jsonData)
		if err != nil {
			logger.Info("Failed to scan row: %v", err)
			continue
		}

		if len(jsonData) == 0 {
			logger.Error("Empty JSON data from database")
			continue
		}

		var pb models.ProductBalanceByCodeStruct
		if err := json.Unmarshal(jsonData, &pb); err != nil {
			logger.Error("Failed to unmarshal JSON: %v", err)
			continue
		}
		products = append(products, pb)
	}

	topLine := false
	if condition == 2 || condition == 3 {
		topLine = true
	}
	datas := []models.ReportDataModel{}
	for _, p := range products {
		if p.BalanceQty == 0 {
			continue
		}
		unitFullName := p.UnitName
		if p.UnitCode != p.UnitName {
			unitFullName = p.UnitCode + "/" + p.UnitName
		}
		// สินค้า
		datas = append(datas, models.ReportDataModel{
			RowIndex:   0,
			TopLine:    topLine,
			BottomLine: false,
			ColumnData: []models.ReportDataColumnModel{
				{Value: p.ItemCode},
				{Value: p.ItemName},
				{Value: p.BarcodeList},
				{Value: unitFullName},
				{Value: p.BalanceQty},
				{Value: p.AverageCost},
				{Value: p.BalanceAmount},
				{Value: p.BalanceWord},
			},
		})
		if condition == 2 || condition == 3 {
			// 2=ตามคลัง
			// 3=ตามคลังและตำแหน่ง
			for _, wh := range p.WareHouses {
				// คลัง
				datas = append(datas, models.ReportDataModel{
					RowIndex: 1,
					ColumnData: []models.ReportDataColumnModel{
						{Value: wh.WareHouseCode},
						{Value: wh.BalanceQty},
						{Value: wh.AverageCost},
						{Value: wh.BalanceAmount},
						{Value: wh.BalanceWord},
					},
				})
				if condition == 3 {
					// 3=ตามคลังและตำแหน่ง
					for _, location := range wh.Locations {
						datas = append(datas, models.ReportDataModel{
							RowIndex: 2,
							ColumnData: []models.ReportDataColumnModel{
								{Value: location.LocationCode},
								{Value: location.BalanceQty},
								{Value: location.BalanceWord},
							},
						})
					}
				}
			}
		}
	}
	// convert finalDate to DateTime
	finalDateTime, _ := time.Parse("2006-01-02", finalDate)
	logger.Debug("finalDateTime: %s --> %v", finalDate, finalDateTime)
	bottomLine := false
	if condition == 1 {
		bottomLine = true
	}

	// ใช้ language dictionary สำหรับ headers (using new simplified API)
	headerList := []models.ReportRowModel{}
	headerList = append(headerList, models.ReportRowModel{
		LeftMarginPercent: 0,
		TopLine:           true,
		BottomLine:        bottomLine,
		Columns: []models.ReportColumnModel{
			{Name: GetColumnText("item_code", languageCode), Width: 1, Align: 0},    // รหัสสินค้า
			{Name: GetColumnText("product_name", languageCode), Width: 3, Align: 0}, // ชื่อสินค้า
			{Name: GetColumnText("barcode_list", languageCode), Width: 2, Align: 0}, // รหัส/Barcode (หลายบาร์โค้ด)
			{Name: GetColumnText("unit", languageCode), Width: 1, Align: 0},         // หน่วยนับ
			{Name: GetColumnText("quantity", languageCode), Width: 1, Align: 2},     // จำนวน
			{Name: GetColumnText("cost", languageCode), Width: 1, Align: 2},         // ต้นทุน
			{Name: GetColumnText("total_value", languageCode), Width: 1, Align: 2},  // มูลค่ารวม
			{Name: GetColumnText("balance_word", languageCode), Width: 2, Align: 0}, // ยอดคงเหลือ (ตัวหนังสือ)
		},
	})
	if condition == 2 || condition == 3 {
		// 2=ตามคลัง
		// 3=ตามคลังและตำแหน่ง
		headerList = append(headerList, models.ReportRowModel{
			LeftMarginPercent: 5,
			TopLine:           false,
			BottomLine:        false,
			FontSize:          6,
			Columns: []models.ReportColumnModel{
				{Name: GetColumnText("warehouse", languageCode), Width: 2, Align: 0},     // คลัง
				{Name: GetColumnText("quantity", languageCode), Width: 2, Align: 2},      // จำนวน
				{Name: GetColumnText("avg_cost", languageCode), Width: 2, Align: 2},      // ต้นทุนเฉลี่ย
				{Name: GetColumnText("balance_value", languageCode), Width: 2, Align: 2}, // มูลค่าคงเหลือ
				{Name: GetColumnText("balance_unit", languageCode), Width: 5, Align: 0},  // คงเหลือ (หน่วย)
			},
		})
		if condition == 3 {
			headerList = append(headerList, models.ReportRowModel{
				LeftMarginPercent: 10,
				TopLine:           false,
				BottomLine:        true,
				FontSize:          6,
				Columns: []models.ReportColumnModel{
					{Name: GetColumnText("location", languageCode), Width: 2, Align: 0},     // ตำแหน่ง
					{Name: GetColumnText("quantity", languageCode), Width: 2, Align: 2},     // จำนวน
					{Name: GetColumnText("balance_unit", languageCode), Width: 5, Align: 0}, // คงเหลือ (หน่วย)
				},
			})
		}
	}

	// สร้าง header ตามภาษาใหม่ (using new simplified API)
	// รายงาน ณ. วันที่ + วันที่ตามรูปแบบภาษา
	headerText := GetHeaderText("report_as_of", languageCode) + " " + FormatDateWithLanguage(finalDateTime, languageCode)

	// เลือกชื่อรายงานตาม condition
	var reportName string
	switch condition {
	case 2:
		reportName = GetReportName("product_balance_by_warehouse", languageCode) // รายงานสินค้าคงเหลือตามคลัง
	case 3:
		reportName = GetReportName("product_balance_by_location", languageCode) // รายงานสินค้าคงเหลือตามตำแหน่ง
	default:
		reportName = GetReportName("product_balance", languageCode) // รายงานสินค้าคงเหลือ
	}

	report := models.ReportModel{
		Name:       reportName,
		Header:     headerText,
		HeaderRows: headerList,
		FooterRows: []models.ReportRowModel{
			{
				LeftMarginPercent: 0,
				TopLine:           true,
				BottomLine:        false,
				Columns:           []models.ReportColumnModel{},
			},
		},
		DataRows: datas,
	}
	// สร้างรายงาน
	reportStyle := models.ReportStyleModel{
		PaperType: "P", // P=Portrait, L=Landscape
	}

	// ตรวจสอบข้อมูลก่อนสร้าง PDF
	if len(datas) == 0 {
		logger.Info("No data available for PDF generation, GUID: %s", guid)
		return ""
	}

	logger.Info("Building PDF with %d data rows for GUID: %s", len(datas), guid)

	result := build.BuildReport(guid, "stock-balance", report, reportStyle)

	if result == "" {
		logger.Info("PDF generation failed for GUID: %s", guid)
		return ""
	}

	logger.Info("PDF generation completed successfully for GUID: %s, Path: %s", guid, result)
	return result
}
