package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	build "smlcloudplatform/internal/goapi/process/build"

	"github.com/mitchellh/mapstructure"
)

// TransSaleInvoiceDecode - decodes transaction sale invoice JSON data
func TransSaleInvoiceDecode(jsonData string) models.MongoDocModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal transSaleInvoice: %v", err)
		return models.MongoDocModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var docData models.MongoDocModel
	decoderConfig := mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           &docData,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339), // สำหรับแปลง string เป็น time.Time
		),
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.MongoDocModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding to struct: %v", err)
		logger.Info("JSON data sample: %s", jsonData[:min(200, len(jsonData))]) // แสดงข้อมูลสำหรับ debug
		return models.MongoDocModel{}
	}

	// Log สำเร็จ (เฉพาะส่วนสำคัญ)
	logger.Info("Successfully decoded DocNo: %s, DocDateTime: %s, TotalAmount: %.2f",
		docData.DocNo, docData.DocDateTime.Format("2006-01-02 15:04:05"), docData.TotalAmount)

	return docData
}

// min function สำหรับ Go version ที่ไม่มี built-in min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetLocalIPAddress - gets the local IP address of the machine
func GetLocalIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("ไม่พบ IP address ในระบบ local")
}

// Helper function to validate shop ID format
func ValidateShopID(shopId string) bool {
	if len(shopId) < 10 || len(shopId) > 50 {
		return false
	}
	// Add more validation logic as needed
	return true
}

// Helper function to validate GUID format
func ValidateGUID(guid string) bool {
	if len(guid) < 10 || len(guid) > 100 {
		return false
	}
	// Add more validation logic as needed
	return true
}

// TransSaleInvoiceBuild - processes transaction sale invoice data
func TransSaleInvoiceBuild(jsonData string) {
	docData := TransSaleInvoiceDecode(jsonData)

	// Validate ShopId before proceeding
	if docData.ShopId == "" {
		logger.Warn("Sale invoice data missing ShopId, skipping processing")
		return
	}

	build.DatabaseChecker(docData.ShopId, false) // เช็ค database
	logger.Info("Processing document: %s %s %s %s %.2f",
		docData.ShopId, docData.DocNo, docData.Description,
		docData.DocDateTime.UTC().Format("2006-01-02 15:04:05"), docData.TotalAmount)
	for _, detail := range docData.Details {
		logger.Info("Detail: %s %s %s", detail.Barcode, detail.ItemCode, detail.UnitCode)
	}
	/*err := mypg.DocUpdate(docData)
	if err != nil {
		logger.Error("PostgreSQL DocUpdate: %v", err)
	}*/
}

// ProductBarcodeDecode - decodes product barcode JSON data
func ProductBarcodeDecode(jsonData string) models.MongoProductBarcodeModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal productBarcode: %v", err)
		return models.MongoProductBarcodeModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var productBarcode models.MongoProductBarcodeModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &productBarcode,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.MongoProductBarcodeModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding map to struct: %v", err)
		return models.MongoProductBarcodeModel{}
	}
	return productBarcode
}

// ProductBarcodeBuild - processes product barcode data
func ProductBarcodeBuild(jsonData string) {
	productBarcode := ProductBarcodeDecode(jsonData)

	// Validate ShopId before proceeding
	if productBarcode.ShopId == "" {
		logger.Warn("Product barcode data missing ShopId, skipping processing")
		return
	}

	build.DatabaseChecker(productBarcode.ShopId, false) // เช็ค database
	err := mypg.ProductBarcodeUpdate(productBarcode)
	if err != nil {
		logger.Error("in PostgreSQL ProductBarcodeUpdate: %v", err)
	}
}

// DeleteSaleInvoiceFromPostgreSQL - deletes sale invoice from PostgreSQL
func DeleteSaleInvoiceFromPostgreSQL(saleInvoice models.MongoDocModel) error {
	logger.Info("Starting deletion of Sale Invoice: ShopID: %s, DocNo: %s", saleInvoice.ShopId, saleInvoice.DocNo)

	db, err := mypg.PgSqlFastConnect(saleInvoice.ShopId)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	ctx := context.Background()

	// Use prepared statements to prevent SQL injection
	deleteQueries := []string{
		"DELETE FROM doc WHERE docno = $1",
		"DELETE FROM docdetail WHERE docno = $1",
		"DELETE FROM docpayment WHERE docno = $1",
	}

	// Execute delete queries
	for _, query := range deleteQueries {
		logger.Info("Executing delete query: %s with docno: %s", query, saleInvoice.DocNo)
		_, err = db.ExecContext(ctx, query, saleInvoice.DocNo)
		if err != nil {
			return fmt.Errorf("error executing delete query '%s': %v", query, err)
		}
	}

	logger.Success("deleted Sale Invoice: ShopID: %s, DocNo: %s", saleInvoice.ShopId, saleInvoice.DocNo)
	return nil
}

// ParseRebuildDates - parses rebuild dates from string format
func ParseRebuildDates(startDateStr, endDateStr string) (time.Time, time.Time, bool) {
	layout := "2006-01-02"
	var startDate, endDate time.Time
	useAllData := false

	if startDateStr == "" && endDateStr == "" {
		useAllData = true
		return startDate, endDate, useAllData
	}

	now := time.Now().UTC()
	startDate = now.AddDate(-100, 0, 0) // ย้อนหลัง 100 ปี เป็นค่าเริ่มต้น
	endDate = now

	if startDateStr != "" {
		if parsedDate, err := time.Parse(layout, startDateStr); err == nil {
			startDate = parsedDate.UTC()
		} else {
			logger.Info("Invalid start date format. Using default: %s", startDate.Format(layout))
		}
	}

	if endDateStr != "" {
		if parsedDate, err := time.Parse(layout, endDateStr); err == nil {
			endDate = parsedDate.UTC().Add(time.Hour*23 + time.Minute*59 + time.Second*59)
		} else {
			logger.Info("Invalid end date format. Using default: %s", endDate.Format(layout))
		}
	}

	return startDate, endDate, useAllData
}
