package dataimport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

// ProductPrepareStatus - สถานะการประมวลผล
type ProductPrepareStatus int

const (
	StatusNotStarted ProductPrepareStatus = 0 // ยังไม่ทำ
	StatusProcessing ProductPrepareStatus = 1 // กำลังทำ
	StatusCompleted  ProductPrepareStatus = 2 // เสร็จแล้ว
	StatusError      ProductPrepareStatus = 3 // เกิดข้อผิดพลาด
)

// ProductPrepareSession - session การประมวลผล
type ProductPrepareSession struct {
	HoldingCode   string                 `json:"holdingcode"`
	FileName      string                 `json:"filename"`
	FilePath      string                 `json:"filepath"`
	JSONFilePath  string                 `json:"jsonfilepath"` // เพิ่ม: path ของ JSON result file
	JSONFileName  string                 `json:"jsonfilename"` // เพิ่ม: ชื่อไฟล์ JSON
	Status        ProductPrepareStatus   `json:"status"`
	Progress      float64                `json:"progress"`
	TotalRows     int                    `json:"totalrows"`
	ProcessedRows int                    `json:"processedrows"`
	SuccessCount  int                    `json:"successcount"`
	ErrorCount    int                    `json:"errorcount"`
	StartTime     time.Time              `json:"starttime"`
	EndTime       *time.Time             `json:"endtime,omitempty"`
	Result        []ProductPrepareResult `json:"result,omitempty"`
	ErrorMessage  string                 `json:"errormessage,omitempty"`
	Mutex         sync.RWMutex           `json:"-"`
}

// ProductPrepareResult - ผลลัพธ์แต่ละแถว
type ProductPrepareResult struct {
	RowNumber       int                    `json:"rownumber"`
	Barcode         string                 `json:"barcode"`
	Name            string                 `json:"name"`
	UnitCode        string                 `json:"unitcode"`
	ProductType     string                 `json:"producttype"`
	TaxType         string                 `json:"taxtype"`
	Code            string                 `json:"code"`
	Price           float64                `json:"price"`
	PriceMember     float64                `json:"pricemember"`
	PriceDelivery   float64                `json:"pricedelivery"`
	PriceOne        float64                `json:"priceone"`
	PriceTwo        float64                `json:"pricetwo"`
	PriceThree      float64                `json:"pricethree"`
	PriceFour       float64                `json:"pricefour"`
	PriceFive       float64                `json:"pricefive"`
	PriceSix        float64                `json:"pricesix"`
	PriceSeven      float64                `json:"priceseven"`
	PriceEight      float64                `json:"priceeight"`
	PriceNine       float64                `json:"pricenine"`
	GroupCode       string                 `json:"groupcode"`
	GroupsuboneCode string                 `json:"groupsubonecode"`
	GroupsubtwoCode string                 `json:"groupsubtwocode"`
	BrandCode       string                 `json:"brandcode"`
	DesignCode      string                 `json:"designcode"`
	ModelCode       string                 `json:"modelcode"`
	PatternCode     string                 `json:"patterncode"`
	GradeCode       string                 `json:"gradecode"`
	CategoryCode    string                 `json:"categorycode"`
	ClassCode       string                 `json:"classcode"`
	StandValue      float64                `json:"standvalue"`
	DivideValue     float64                `json:"dividevalue"`
	Status          string                 `json:"status"` // "success", "error", "warning"
	Message         string                 `json:"message"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

var (
	// เก็บ session ตาม holdingCode
	productPrepareSessions = make(map[string]*ProductPrepareSession)
	xlsxSessionMutex       sync.RWMutex
)

// StartProductPrepareHandler - เริ่มประมวลผลไฟล์ Excel
// POST /xlsx/product/start
// Form-data: holdingCode, file (Excel file)
func StartProductPrepareHandler(c echo.Context) error {
	holdingCode := c.FormValue("holdingCode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "holdingCode is required",
		})
	}

	// ตรวจสอบว่ามี session ทำงานอยู่แล้วหรือไม่
	xlsxSessionMutex.RLock()
	existingSession, exists := productPrepareSessions[holdingCode]
	xlsxSessionMutex.RUnlock()

	if exists && existingSession.Status == StatusProcessing {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"error":   "Already processing",
			"message": "Another process is already running for this shop",
			"status":  StatusProcessing,
		})
	}

	// รับไฟล์ Excel
	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to get Excel file: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "No file uploaded",
			"message": err.Error(),
		})
	}

	// ตรวจสอบนามสกุลไฟล์
	ext := filepath.Ext(file.Filename)
	if ext != ".xlsx" && ext != ".xls" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid file type",
			"message": "Only .xlsx and .xls files are supported",
		})
	}

	// บันทึกไฟล์ไปที่ temp directory
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to process file",
			"message": err.Error(),
		})
	}
	defer src.Close()

	// ใช้ system temp directory
	tempDir := os.TempDir()
	uploadDir := filepath.Join(tempDir, "uploads", "xlsx", holdingCode)
	os.MkdirAll(uploadDir, 0755)

	filePath := filepath.Join(uploadDir, file.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		logger.Error("Failed to create file: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save file",
			"message": err.Error(),
		})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		logger.Error("Failed to copy file: %v", err)
		os.Remove(filePath)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save file",
			"message": err.Error(),
		})
	}

	logger.Info("Excel file saved: %s for shop %s", filePath, holdingCode)

	// สร้าง session ใหม่
	session := &ProductPrepareSession{
		HoldingCode: holdingCode,
		FileName:    file.Filename,
		FilePath:    filePath,
		Status:      StatusProcessing,
		Progress:    0,
		StartTime:   time.Now(),
		Result:      []ProductPrepareResult{},
	}

	// เก็บ session
	xlsxSessionMutex.Lock()
	productPrepareSessions[holdingCode] = session
	xlsxSessionMutex.Unlock()

	// เริ่มประมวลผลใน goroutine
	go processProductExcel(session)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":     true,
		"message":     "Processing started",
		"holdingCode": holdingCode,
		"fileName":    file.Filename,
		"status":      StatusProcessing,
	})
}

// GetProductPrepareStatusHandler - ตรวจสอบสถานะการประมวลผล
// GET /xlsx/product/status?holdingCode=xxx
func GetProductPrepareStatusHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingCode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "holdingCode is required",
		})
	}

	xlsxSessionMutex.RLock()
	session, exists := productPrepareSessions[holdingCode]
	xlsxSessionMutex.RUnlock()

	if !exists {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"holdingCode": holdingCode,
			"status":      StatusNotStarted,
		})
	}

	session.Mutex.RLock()
	defer session.Mutex.RUnlock()

	response := map[string]interface{}{
		"holdingCode":   session.HoldingCode,
		"fileName":      session.FileName,
		"status":        session.Status,
		"progress":      session.Progress,
		"totalRows":     session.TotalRows,
		"processedRows": session.ProcessedRows,
		"successCount":  session.SuccessCount,
		"errorCount":    session.ErrorCount,
		"startTime":     session.StartTime,
	}

	if session.EndTime != nil {
		response["endTime"] = session.EndTime
		duration := session.EndTime.Sub(session.StartTime)
		response["duration"] = duration.String()
	}

	if session.ErrorMessage != "" {
		response["errorMessage"] = session.ErrorMessage
	}

	return c.JSON(http.StatusOK, response)
}

// GetProductPrepareResultHandler - ดึงผลลัพธ์การประมวลผล
// GET /xlsx/product/result?holdingCode=xxx
func GetProductPrepareResultHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingCode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "holdingCode is required",
		})
	}

	xlsxSessionMutex.RLock()
	session, exists := productPrepareSessions[holdingCode]
	xlsxSessionMutex.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "Session not found",
			"message": "No processing session found for this shop",
		})
	}

	session.Mutex.RLock()
	defer session.Mutex.RUnlock()

	if session.Status == StatusProcessing {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"holdingCode": holdingCode,
			"status":      StatusProcessing,
			"message":     "Still processing, please wait",
			"progress":    session.Progress,
		})
	}

	response := map[string]interface{}{
		"holdingCode":   session.HoldingCode,
		"fileName":      session.FileName,
		"status":        session.Status,
		"totalRows":     session.TotalRows,
		"processedRows": session.ProcessedRows,
		"successCount":  session.SuccessCount,
		"errorCount":    session.ErrorCount,
		"startTime":     session.StartTime,
		"result":        session.Result,
		"jsonFileName":  session.JSONFileName, // เพิ่ม: ชื่อไฟล์ JSON
		"jsonFilePath":  session.JSONFilePath, // เพิ่ม: path ของไฟล์ JSON
	}

	if session.EndTime != nil {
		response["endTime"] = session.EndTime
		duration := session.EndTime.Sub(session.StartTime)
		response["duration"] = duration.String()
	}

	if session.ErrorMessage != "" {
		response["errorMessage"] = session.ErrorMessage
	}

	return c.JSON(http.StatusOK, response)
}

// processProductExcel - ประมวลผลไฟล์ Excel
func processProductExcel(session *ProductPrepareSession) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in processProductExcel: %v", r)
			session.Mutex.Lock()
			session.Status = StatusError
			session.ErrorMessage = fmt.Sprintf("Panic: %v", r)
			endTime := time.Now()
			session.EndTime = &endTime
			session.Mutex.Unlock()
		}
	}()

	logger.Info("Starting Excel processing for shop %s: %s", session.HoldingCode, session.FilePath)

	// เปิดไฟล์ Excel
	f, err := excelize.OpenFile(session.FilePath)
	if err != nil {
		logger.Error("Failed to open Excel file: %v", err)
		session.Mutex.Lock()
		session.Status = StatusError
		session.ErrorMessage = fmt.Sprintf("Failed to open file: %v", err)
		endTime := time.Now()
		session.EndTime = &endTime
		session.Mutex.Unlock()
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Error("Failed to close Excel file: %v", err)
		}
	}()

	// อ่าน sheet แรก
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		session.Mutex.Lock()
		session.Status = StatusError
		session.ErrorMessage = "No sheets found in Excel file"
		endTime := time.Now()
		session.EndTime = &endTime
		session.Mutex.Unlock()
		return
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		logger.Error("Failed to read rows: %v", err)
		session.Mutex.Lock()
		session.Status = StatusError
		session.ErrorMessage = fmt.Sprintf("Failed to read rows: %v", err)
		endTime := time.Now()
		session.EndTime = &endTime
		session.Mutex.Unlock()
		return
	}

	if len(rows) <= 1 {
		session.Mutex.Lock()
		session.Status = StatusError
		session.ErrorMessage = "Excel file is empty or contains only headers"
		endTime := time.Now()
		session.EndTime = &endTime
		session.Mutex.Unlock()
		return
	}

	// กำหนดจำนวนแถวทั้งหมด (ไม่นับ header)
	totalRows := len(rows) - 1
	session.Mutex.Lock()
	session.TotalRows = totalRows
	session.Mutex.Unlock()

	logger.Info("Processing %d rows from sheet '%s'", totalRows, sheetName)

	// อ่าน header (แถวแรก)
	headers := rows[0]
	logger.Info("Headers: %v", headers)

	// ประมวลผลแต่ละแถว (เริ่มจากแถวที่ 2)
	results := []ProductPrepareResult{}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowNumber := i + 1 // เพิ่ม 1 เพราะนับรวม header

		result := processProductRow(session.HoldingCode, rowNumber, headers, row)
		results = append(results, result)

		// อัปเดตความคืบหน้า
		session.Mutex.Lock()
		session.ProcessedRows = i
		session.Progress = float64(i) / float64(totalRows) * 100

		if result.Status == "success" {
			session.SuccessCount++
		} else if result.Status == "error" {
			session.ErrorCount++
		}
		session.Mutex.Unlock()

		// Log ทุก 100 แถว
		if i%100 == 0 {
			logger.Info("Processed %d/%d rows (%.1f%%)", i, totalRows, session.Progress)
		}
	}

	// เสร็จสิ้น - บันทึก JSON file
	session.Mutex.Lock()
	session.Status = StatusCompleted
	session.Progress = 100
	session.Result = results
	endTime := time.Now()
	session.EndTime = &endTime

	// สร้าง JSON file
	jsonFileName, jsonFilePath, err := saveResultsToJSON(session.HoldingCode, results)
	if err != nil {
		logger.Error("Failed to save JSON file: %v", err)
		session.Status = StatusError
		session.ErrorMessage = fmt.Sprintf("Failed to save results: %v", err)
	} else {
		session.JSONFileName = jsonFileName
		session.JSONFilePath = jsonFilePath
		logger.Info("Results saved to JSON: %s", jsonFilePath)
	}

	session.Mutex.Unlock()

	duration := endTime.Sub(session.StartTime)
	logger.Success("Excel processing completed for shop %s: %d rows in %s (Success: %d, Error: %d)",
		session.HoldingCode, totalRows, duration, session.SuccessCount, session.ErrorCount)

	// Cleanup Excel file after processing
	os.Remove(session.FilePath)
	logger.Info("Cleaned up Excel file: %s", session.FilePath)
}

// processProductRow - ประมวลผลแต่ละแถว
func processProductRow(holdingCode string, rowNumber int, headers []string, row []string) ProductPrepareResult {
	result := ProductPrepareResult{
		RowNumber: rowNumber,
		Status:    "success",
		Data:      make(map[string]interface{}),
	}

	// สร้าง map จาก headers และ row
	rowData := make(map[string]string)
	for i, header := range headers {
		if i < len(row) {
			rowData[header] = row[i]
		} else {
			rowData[header] = ""
		}
	}

	// Map ข้อมูลตามโครงสร้างที่กำหนด
	result.Barcode = getValueFromRow(rowData, []string{"Barcode", "barcode"})
	result.Name = getValueFromRow(rowData, []string{"Name", "name"})
	result.UnitCode = getValueFromRow(rowData, []string{"Unit Code", "UnitCode", "unit_code"})
	result.ProductType = getValueFromRow(rowData, []string{"Product Type", "ProductType", "product_type"})
	result.TaxType = getValueFromRow(rowData, []string{"Tax Type", "TaxType", "tax_type"})
	result.Code = getValueFromRow(rowData, []string{"Code", "code"})

	// Prices
	result.Price = parseFloat(getValueFromRow(rowData, []string{"Price", "price"}))
	result.PriceMember = parseFloat(getValueFromRow(rowData, []string{"Price Member", "PriceMember", "price_member"}))
	result.PriceDelivery = parseFloat(getValueFromRow(rowData, []string{"Price Delivery", "PriceDelivery", "price_delivery"}))
	result.PriceOne = parseFloat(getValueFromRow(rowData, []string{"PriceOne", "price_one"}))
	result.PriceTwo = parseFloat(getValueFromRow(rowData, []string{"PriceTwo", "price_two"}))
	result.PriceThree = parseFloat(getValueFromRow(rowData, []string{"PriceThree", "price_three"}))
	result.PriceFour = parseFloat(getValueFromRow(rowData, []string{"PriceFour", "price_four"}))
	result.PriceFive = parseFloat(getValueFromRow(rowData, []string{"PriceFive", "price_five"}))
	result.PriceSix = parseFloat(getValueFromRow(rowData, []string{"PriceSix", "price_six"}))
	result.PriceSeven = parseFloat(getValueFromRow(rowData, []string{"PriceSeven", "price_seven"}))
	result.PriceEight = parseFloat(getValueFromRow(rowData, []string{"PriceEight", "price_eight"}))
	result.PriceNine = parseFloat(getValueFromRow(rowData, []string{"PriceNine", "price_nine"}))

	// Codes
	result.GroupCode = getValueFromRow(rowData, []string{"GroupCode", "group_code"})
	result.GroupsuboneCode = getValueFromRow(rowData, []string{"GroupsuboneCode", "groupsubone_code"})
	result.GroupsubtwoCode = getValueFromRow(rowData, []string{"GroupsubtwoCode", "groupsubtwo_code"})
	result.BrandCode = getValueFromRow(rowData, []string{"BrandCode", "brand_code"})
	result.DesignCode = getValueFromRow(rowData, []string{"DesignCode", "design_code"})
	result.ModelCode = getValueFromRow(rowData, []string{"ModelCode", "model_code"})
	result.PatternCode = getValueFromRow(rowData, []string{"PatternCode", "pattern_code"})
	result.GradeCode = getValueFromRow(rowData, []string{"GradeCode", "grade_code"})
	result.CategoryCode = getValueFromRow(rowData, []string{"CategoryCode", "category_code"})
	result.ClassCode = getValueFromRow(rowData, []string{"ClassCode", "class_code"})

	// Values
	result.StandValue = parseFloat(getValueFromRow(rowData, []string{"StandValue", "stand_value"}))
	result.DivideValue = parseFloat(getValueFromRow(rowData, []string{"DivideValue", "divide_value"}))

	// Validation - Code เป็น required field
	if result.Code == "" {
		result.Status = "error"
		result.Message = "Code is required"
		return result
	}

	// Warning ถ้าไม่มีชื่อ
	if result.Name == "" {
		result.Status = "warning"
		result.Message = "Name is empty"
	}

	// เก็บข้อมูลทั้งหมด
	result.Data["rowData"] = rowData
	result.Data["holdingCode"] = holdingCode

	if result.Message == "" {
		result.Message = "Processed successfully"
	}

	return result
}

// getValueFromRow - ดึงค่าจาก row โดยตรวจสอบหลาย column names
func getValueFromRow(rowData map[string]string, possibleNames []string) string {
	for _, name := range possibleNames {
		if value, exists := rowData[name]; exists && value != "" {
			return value
		}
	}
	return ""
}

// parseFloat - แปลง string เป็น float64
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}

	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

// saveResultsToJSON - บันทึกผลลัพธ์เป็น JSON file
func saveResultsToJSON(holdingCode string, results []ProductPrepareResult) (string, string, error) {
	// สร้าง folder สำหรับเก็บ JSON
	tempDir := os.TempDir()
	jsonDir := filepath.Join(tempDir, "uploads", "xlsx", holdingCode, "results")
	if err := os.MkdirAll(jsonDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create JSON directory: %v", err)
	}

	// สร้างชื่อไฟล์ JSON (ใช้ timestamp)
	timestamp := time.Now().Format("20060102_150405")
	jsonFileName := fmt.Sprintf("products_%s_%s.json", holdingCode, timestamp)
	jsonFilePath := filepath.Join(jsonDir, jsonFileName)

	// แปลง results เป็น JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// บันทึกไฟล์
	if err := os.WriteFile(jsonFilePath, jsonData, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write JSON file: %v", err)
	}

	logger.Info("JSON file created: %s (%d bytes)", jsonFilePath, len(jsonData))
	return jsonFileName, jsonFilePath, nil
}

// ClearProductPrepareSession - ลบ session (เรียกใช้หลังดึงผลลัพธ์แล้ว)
func ClearProductPrepareSession(holdingCode string) {
	xlsxSessionMutex.Lock()
	defer xlsxSessionMutex.Unlock()

	if session, exists := productPrepareSessions[holdingCode]; exists {
		// ลบไฟล์
		if session.FilePath != "" {
			os.Remove(session.FilePath)
			logger.Info("Cleaned up Excel file: %s", session.FilePath)
		}
		delete(productPrepareSessions, holdingCode)
	}
}

// GetProductPrepareSessionsJSON - ดึง sessions ทั้งหมด (สำหรับ debug)
func GetProductPrepareSessionsJSON() (string, error) {
	xlsxSessionMutex.RLock()
	defer xlsxSessionMutex.RUnlock()

	data, err := json.MarshalIndent(productPrepareSessions, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
