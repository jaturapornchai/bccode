package reportstock

import (
	"encoding/json"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/build"
)

func ReportProductStockMovement(shopId string, guid string, timezoneCode string, languageCode string) string {
	logger.Info("ReportProductStockMovement: shopId=%s, guid=%s, timezoneCode=%s, languageCode=%s", shopId, guid, timezoneCode, languageCode)
	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL: %v", err)
		return ""
	}

	query := "SELECT datajson FROM result WHERE guid = $1 ORDER BY linenumber"

	rows, err := db.Query(query, guid)
	if err != nil {
		logger.Info("Failed to execute query: %v", err)
		return ""
	}
	defer rows.Close()

	var products = []models.ProcessStockMovementStruct{}
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

		var pb models.ProcessStockMovementStruct
		if err := json.Unmarshal(jsonData, &pb); err != nil {
			logger.Error("Failed to unmarshal JSON: %v", err)
			continue
		}
		products = append(products, pb)
	}

	logger.Info("Successfully loaded %d products", len(products))

	datas := []models.ReportDataModel{}
	for _, p := range products {
		// สินค้า
		datas = append(datas, models.ReportDataModel{
			RowIndex: 0,
			ColumnData: []models.ReportDataColumnModel{
				{Value: p.ItemCode},
				{Value: p.Name},
				{Value: p.UnitName},
			},
		})
		for _, detail := range p.Details {
			qtyAdd := 0.0
			costAdd := 0.0
			amountAdd := 0.0
			qtySub := 0.0
			costSub := 0.0
			amountSub := 0.0

			if detail.CalcAmount >= 0 {
				qtyAdd = detail.TotalQty * (detail.UnitStand / detail.UnitDivide)
				costAdd = detail.UnitCost
				amountAdd = detail.CalcAmount
			} else {
				qtySub = (detail.TotalQty * (detail.UnitStand / detail.UnitDivide)) * -1
				costSub = detail.UnitCost
				amountSub = detail.CalcAmount * -1
			}

			datas = append(datas, models.ReportDataModel{
				RowIndex: 1,
				ColumnData: []models.ReportDataColumnModel{
					{Value: detail.DocDateTime, DateTimeStyle: 1},
					{Value: detail.DocNo},
					{Value: myglobal.TransFlagName(detail.TransFlag, detail.TotalQty)},
					{Value: detail.UnitCode},
					{Value: detail.WhCode},
					{Value: detail.LocationCode},
					{Value: qtyAdd},
					{Value: costAdd},
					{Value: amountAdd},
					{Value: qtySub},
					{Value: costSub},
					{Value: amountSub},
					{Value: detail.BalanceQty},
					{Value: detail.AverageCost},
					{Value: detail.BalanceAmount},
				},
			})
		}
	}

	// ใช้ language dictionary
	// สร้าง header ตามภาษาใหม่ (using new simplified API)
	// รายงาน ณ. วันที่ + วันที่ตามรูปแบบภาษา
	headerText := GetHeaderText("report_as_of", languageCode) + " " + FormatDateWithLanguage(time.Now(), languageCode)

	report := models.ReportModel{
		Name:   GetReportName("product_stock_movement", languageCode), // รายงานเคลื่อนไหวสินค้า
		Header: headerText,
		HeaderRows: []models.ReportRowModel{
			{
				LeftMarginPercent: 0,
				TopLine:           true,
				BottomLine:        false,
				Columns: []models.ReportColumnModel{
					{Name: GetColumnText("code", languageCode), Width: 1, Align: 0},         // รหัส/Barcode
					{Name: GetColumnText("product_name", languageCode), Width: 4, Align: 0}, // ชื่อสินค้า
					{Name: GetColumnText("unit", languageCode), Width: 1, Align: 0},         // หน่วยนับ
				}},
			{
				LeftMarginPercent: 0,
				TopLine:           false,
				BottomLine:        true,
				FontSize:          5,
				Columns: []models.ReportColumnModel{
					{Name: GetColumnText("date", languageCode), Width: 4, Align: 0},           // วันที่
					{Name: GetColumnText("document", languageCode), Width: 6, Align: 0},       // เอกสาร
					{Name: GetColumnText("type", languageCode), Width: 4, Align: 0},           // ประเภท
					{Name: GetColumnText("unit", languageCode), Width: 4, Align: 0},           // หน่วยนับ
					{Name: GetColumnText("warehouse", languageCode), Width: 4, Align: 0},      // คลัง
					{Name: GetColumnText("location", languageCode), Width: 4, Align: 0},       // ตำแหน่ง
					{Name: GetColumnText("increase_qty", languageCode), Width: 3, Align: 2},   // จำนวนเพิ่ม
					{Name: GetColumnText("cost", languageCode), Width: 3, Align: 2},           // ต้นทุน
					{Name: GetColumnText("increase_value", languageCode), Width: 3, Align: 2}, // มูลค่าบวก
					{Name: GetColumnText("decrease_qty", languageCode), Width: 3, Align: 2},   // จำนวนลด
					{Name: GetColumnText("cost", languageCode), Width: 3, Align: 2},           // ต้นทุน
					{Name: GetColumnText("decrease_value", languageCode), Width: 3, Align: 2}, // มูลค่าลบ
					{Name: GetColumnText("balance", languageCode), Width: 4, Align: 2},        // คงเหลือ
					{Name: GetColumnText("avg_cost", languageCode), Width: 3, Align: 2},       // ต้นทุนเฉลี่ย
					{Name: GetColumnText("balance_value", languageCode), Width: 4, Align: 2},  // มูลค่าคงเหลือ
				},
			},
		},
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
	return build.BuildReport(guid, "stock-movement", report, reportStyle)
}
