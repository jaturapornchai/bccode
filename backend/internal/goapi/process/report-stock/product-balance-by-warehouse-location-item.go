package reportstock

import (
	"encoding/json"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/build"
)

func ReportProductBalanceByLocationAndItem(holdingCode string, guid string, finalDate string, timezoneCode string, languageCode string) string {
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

	var dataList = []models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct{}
	for rows.Next() {
		var jsonData []byte
		err := rows.Scan(&jsonData)
		if err != nil {
			logger.Error("Failed to scan row: %v", err)
			continue
		}

		if len(jsonData) == 0 {
			logger.Error("Empty JSON data from database")
			continue
		}

		var pb models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct
		if err := json.Unmarshal(jsonData, &pb); err != nil {
			logger.Error("Failed to unmarshal JSON: %v", err)
			continue
		}
		dataList = append(dataList, pb)
	}

	datas := []models.ReportDataModel{}
	for _, p := range dataList {
		unitFullName := p.UnitName
		if p.UnitCode != p.UnitName {
			unitFullName = p.UnitCode + "/" + p.UnitName
		}
		// สินค้า
		datas = append(datas, models.ReportDataModel{
			RowIndex:   0,
			TopLine:    false,
			BottomLine: true,
			ColumnData: []models.ReportDataColumnModel{
				{Value: p.WareHouseCode},
				{Value: p.LocationCode},
				{Value: p.ItemCode},
				{Value: p.ItemName},
				{Value: p.BarcodeList},
				{Value: unitFullName},
				{Value: p.BalanceQty},
				{Value: p.BalanceWord},
			},
		})
	}
	// convert finalDate to DateTime
	finalDateTime, _ := time.Parse("2006-01-02", finalDate)
	logger.Debug("finalDateTime: %s --> %v", finalDate, finalDateTime)

	// สร้าง header ตามภาษาใหม่ (using new simplified API)
	// รายงาน ณ. วันที่ + วันที่ตามรูปแบบภาษา
	headerText := GetHeaderText("report_as_of", languageCode) + " " + FormatDateWithLanguage(finalDateTime, languageCode)

	headerList := []models.ReportRowModel{}
	headerList = append(headerList, models.ReportRowModel{
		LeftMarginPercent: 0,
		TopLine:           true,
		BottomLine:        true,
		Columns: []models.ReportColumnModel{
			{Name: GetColumnText("warehouse", languageCode), Width: 1, Align: 0},    // คลัง
			{Name: GetColumnText("location", languageCode), Width: 1, Align: 0},     // ตำแหน่ง
			{Name: GetColumnText("itemcode", languageCode), Width: 1, Align: 0},     // รหัสสินค้า
			{Name: GetColumnText("product_name", languageCode), Width: 3, Align: 0}, // ชื่อสินค้า
			{Name: GetColumnText("barcodelist", languageCode), Width: 2, Align: 0},  // รหัส/Barcode (หลายบาร์โค้ด)
			{Name: GetColumnText("unit", languageCode), Width: 1, Align: 0},         // หน่วยนับ
			{Name: GetColumnText("quantity", languageCode), Width: 1, Align: 2},     // จำนวน
			{Name: GetColumnText("balance_word", languageCode), Width: 1, Align: 0}, // ยอดคงเหลือ (ตัวหนังสือ)
		},
	})

	report := models.ReportModel{
		Name:       GetReportName("product_balance_by_location", languageCode), // รายงานสินค้าคงเหลือตามตำแหน่ง
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
	return build.BuildReport(guid, "stock-balance", report, reportStyle)
}
