package dataimport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
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
	HoldingCode       string                 `json:"holding_code"`
	FileName          string                 `json:"file_name"`
	FilePath          string                 `json:"file_path"`
	Status            ProductPrepareStatus   `json:"status"`
	Progress          float64                `json:"progress"`
	TotalRows         int                    `json:"total_rows"`
	ProcessedRows     int                    `json:"processed_rows"`
	SuccessCount      int                    `json:"success_count"`
	ErrorCount        int                    `json:"error_count"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           *time.Time             `json:"end_time,omitempty"`
	Result            []ProductPrepareResult `json:"result,omitempty"`
	ComparisonResult  *ComparisonSummary     `json:"comparison,omitempty"`         // เพิ่ม: ผลการเปรียบเทียบกับ MongoDB
	ErrorMessage      string                 `json:"error_message,omitempty"`      // ข้อความ error แบบ string (เก่า)
	DuplicateBarcodes []DuplicateBarcodeInfo `json:"duplicate_barcodes,omitempty"` // รายการ barcode ซ้ำแบบ JSON
	Mutex             sync.RWMutex           `json:"-"`
}

// ProductPrepareResult - ผลลัพธ์จากการ parse แถวใน Excel
type ProductPrepareResult struct {
	RowNumber       int                    `json:"row_number"` // เลขแถวใน Excel (เริ่มจาก 1)
	Barcode         string                 `json:"barcode"`
	ItemCode        string                 `json:"item_code"`
	Name            string                 `json:"name"`
	ShortName       string                 `json:"short_name"`
	NameEN          string                 `json:"name_en"`
	NameCN          string                 `json:"name_cn"`
	UnitCode        string                 `json:"unit_code"`
	ProductType     string                 `json:"product_type"`
	TaxType         string                 `json:"tax_type"`
	Code            string                 `json:"code"`
	Price           float64                `json:"price"`
	PriceMember     float64                `json:"price_member"`
	PriceDelivery   float64                `json:"price_delivery"`
	PriceOne        float64                `json:"price_one"`
	PriceTwo        float64                `json:"price_two"`
	PriceThree      float64                `json:"price_three"`
	PriceFour       float64                `json:"price_four"`
	PriceFive       float64                `json:"price_five"`
	PriceSix        float64                `json:"price_six"`
	PriceSeven      float64                `json:"price_seven"`
	PriceEight      float64                `json:"price_eight"`
	PriceNine       float64                `json:"price_nine"`
	GroupCode       string                 `json:"group_code"`
	GroupsuboneCode string                 `json:"groupsubone_code"`
	GroupsubtwoCode string                 `json:"groupsubtwo_code"`
	BrandCode       string                 `json:"brand_code"`
	DesignCode      string                 `json:"design_code"`
	ModelCode       string                 `json:"model_code"`
	PatternCode     string                 `json:"pattern_code"`
	GradeCode       string                 `json:"grade_code"`
	CategoryCode    string                 `json:"category_code"`
	ClassCode       string                 `json:"class_code"`
	StandValue      float64                `json:"stand_value"`
	DivideValue     float64                `json:"divide_value"`
	Status          string                 `json:"status"` // "success", "error", "warning"
	Message         string                 `json:"message"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

// DuplicateBarcodeInfo - ข้อมูล barcode ที่ซ้ำ
type DuplicateBarcodeInfo struct {
	Barcode    string `json:"barcode"`
	Count      int    `json:"count"`
	RowNumbers []int  `json:"row_numbers"` // เลขแถวใน Excel ที่ barcode นี้ปรากฏ
}

// ComparisonSummary - สรุปผลการเปรียบเทียบ
type ComparisonSummary struct {
	HoldingCode         string           `json:"holding_code"`
	TotalExcelRows      int              `json:"total_excel_rows"`
	TotalMongoProducts  int              `json:"total_mongo_products"`
	NewProductCount     int              `json:"new_product_count"`     // action = 1
	UpdatedProductCount int              `json:"updated_product_count"` // action = 2
	UnchangedCount      int              `json:"unchanged_count"`       // action = 0
	ProcessTime         time.Time        `json:"process_time"`
	Products            []ProductCompact `json:"products"` // รวมทุก product ไว้ที่เดียว
}

// ProductCompact - ข้อมูลสินค้าแบบย่อ (แสดงทั้ง MongoDB และ Excel)
type ProductCompact struct {
	Barcode string              `json:"barcode"`
	Mongo   *ProductCompactData `json:"mongo"`           // ข้อมูลจาก MongoDB
	Excel   *ProductCompactData `json:"excel,omitempty"` // ข้อมูลจาก Excel (ถ้ามี)
	Action  int                 `json:"action"`          // 0=match, 1=insert, 2=update
}

// ProductCompactData - ข้อมูล 5 fields ที่ compare
type ProductCompactData struct {
	Code        string  `json:"code"`         // itemcode
	Name        string  `json:"name"`         // names[code='th'].name
	UnitCode    string  `json:"unit_code"`    // itemunitcode (mongo) / unitcode (excel)
	DivideValue float64 `json:"divide_value"` // dividevalue
	StandValue  float64 `json:"stand_value"`  // standvalue
}

var (
	// เก็บ session ตาม holdingCode
	productPrepareSessions = make(map[string]*ProductPrepareSession)
	xlsxSessionMutex       sync.RWMutex
)

// StartProductPrepareHandler - เริ่มประมวลผลไฟล์ Excel
// POST /xlsx/product/start
// JSON body: {"holdingCode": "xxx", "fileName": "uploaded_file.xlsx"}
func StartProductPrepareHandler(c echo.Context) error {
	// รับ JSON body
	var request struct {
		HoldingCode string `json:"holding_code"`
		FileName    string `json:"file_name"`
		FileUrl     string `json:"file_url"` // presigned URL สำหรับ download จาก S3
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request",
			"message": err.Error(),
		})
	}

	if request.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "holdingCode is required",
		})
	}

	if request.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "fileName is required",
		})
	}

	// ตรวจสอบว่ามี session ทำงานอยู่แล้วหรือไม่
	xlsxSessionMutex.RLock()
	existingSession, exists := productPrepareSessions[request.HoldingCode]
	xlsxSessionMutex.RUnlock()

	if exists && existingSession.Status == StatusProcessing {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"error":   "Already processing",
			"message": "Another process is already running for this shop",
			"status":  StatusProcessing,
		})
	}

	// หาไฟล์ที่ upload ไว้แล้ว
	tempDir := os.TempDir()
	uploadDir := filepath.Join(tempDir, "uploads", "xlsx", request.HoldingCode)
	os.MkdirAll(uploadDir, 0755)
	filePath := filepath.Join(uploadDir, request.FileName)

	logger.Info("Looking for file: %s", filePath)

	// ตรวจสอบว่าไฟล์มีอยู่ในเครื่อง
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// ไฟล์ไม่อยู่ในเครื่อง → ลอง download จาก S3 URL
		if request.FileUrl != "" {
			logger.Info("File not found locally, downloading from S3: %s", request.FileUrl)
			if dlErr := downloadFileFromURL(request.FileUrl, filePath); dlErr != nil {
				logger.Error("Failed to download file from S3: %v", dlErr)
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error":   "Failed to download file",
					"message": dlErr.Error(),
				})
			}
			logger.Info("File downloaded from S3 to: %s", filePath)
		} else {
			// ลอง path เดิม (backward compatible)
			oldPath := filepath.Join(tempDir, "uploads", request.FileName)
			if _, statErr := os.Stat(oldPath); statErr == nil {
				filePath = oldPath
				logger.Info("File found at legacy path: %s", filePath)
			} else {
				return c.JSON(http.StatusNotFound, map[string]interface{}{
					"error":   "File not found",
					"message": fmt.Sprintf("File %s not found. Please provide fileUrl for S3 download.", request.FileName),
				})
			}
		}
	}

	// ตรวจสอบนามสกุลไฟล์
	ext := filepath.Ext(request.FileName)
	if ext != ".xlsx" && ext != ".xls" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid file type",
			"message": "Only .xlsx and .xls files are supported",
		})
	}

	logger.Info("File ready for processing: %s", filePath)
	logger.Info("Starting Excel processing for shop %s", request.HoldingCode)

	// สร้าง session ใหม่
	session := &ProductPrepareSession{
		HoldingCode: request.HoldingCode,
		FileName:    request.FileName,
		FilePath:    filePath,
		Status:      StatusProcessing,
		Progress:    0,
		StartTime:   time.Now(),
		Result:      []ProductPrepareResult{},
	}

	// เก็บ session
	xlsxSessionMutex.Lock()
	productPrepareSessions[request.HoldingCode] = session
	xlsxSessionMutex.Unlock()

	// ประมวลผล Excel แบบ synchronous (รอให้เสร็จก่อน return)
	logger.Info("Processing Excel file synchronously...")
	processProductExcel(session)

	// รอให้ประมวลผลเสร็จ แล้ว return เฉพาะผลการเปรียบเทียบ
	session.Mutex.RLock()
	response := map[string]interface{}{
		"success":      true,
		"holdingCode":  session.HoldingCode,
		"fileName":     session.FileName,
		"status":       session.Status,
		"totalRows":    session.TotalRows,
		"successCount": session.SuccessCount,
		"errorCount":   session.ErrorCount,
		"comparison":   session.ComparisonResult, // เฉพาะผลการเปรียบเทียบ
	}

	if session.EndTime != nil {
		response["duration"] = session.EndTime.Sub(session.StartTime).String()
	}

	if session.ErrorMessage != "" {
		response["errorMessage"] = session.ErrorMessage
	}

	if len(session.DuplicateBarcodes) > 0 {
		response["duplicateBarcodes"] = session.DuplicateBarcodes
	}
	session.Mutex.RUnlock()

	// Log response summary
	logger.Info("Response: success=%v, totalRows=%d, successCount=%d, errorCount=%d",
		response["success"], response["totalRows"], response["successCount"], response["errorCount"])

	if session.ComparisonResult != nil {
		logger.Info("Comparison: New=%d, Updated=%d, Unchanged=%d",
			session.ComparisonResult.NewProductCount,
			session.ComparisonResult.UpdatedProductCount,
			session.ComparisonResult.UnchangedCount)
	}

	return c.JSON(http.StatusOK, response)
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

	if len(session.DuplicateBarcodes) > 0 {
		response["duplicateBarcodes"] = session.DuplicateBarcodes
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
		"products":      session.Result,           // ข้อมูลสินค้าจาก Excel
		"comparison":    session.ComparisonResult, // ผลการเปรียบเทียบกับ MongoDB
	}

	if session.EndTime != nil {
		response["endTime"] = session.EndTime
		duration := session.EndTime.Sub(session.StartTime)
		response["duration"] = duration.String()
	}

	if session.ErrorMessage != "" {
		response["errorMessage"] = session.ErrorMessage
	}

	if len(session.DuplicateBarcodes) > 0 {
		response["duplicateBarcodes"] = session.DuplicateBarcodes
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

	// แปลง rows ให้เป็น string ทั้งหมด (รองรับทั้ง number และ string)
	stringRows := make([][]string, len(rows))
	for i, row := range rows {
		stringRow := make([]string, len(row))
		for j := range row {
			// แปลง cell เป็น string ทุกประเภท
			colName, _ := excelize.ColumnNumberToName(j + 1)
			cellRef := fmt.Sprintf("%s%d", colName, i+1)
			cellValue, _ := f.GetCellValue(sheetName, cellRef)
			stringRow[j] = cellValue
		}
		stringRows[i] = stringRow
	}
	rows = stringRows

	// กำหนดจำนวนแถวทั้งหมด (ไม่นับ header)
	totalRows := len(rows) - 1
	session.Mutex.Lock()
	session.TotalRows = totalRows
	session.Mutex.Unlock()

	logger.Info("Processing %d rows from sheet '%s'", totalRows, sheetName)

	// อ่าน header (แถวแรก) และ trim space
	headers := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		headers[i] = strings.TrimSpace(h)
	}
	logger.Info("Headers: %v", headers)

	// ประมวลผลแต่ละแถว (เริ่มจากแถวที่ 2)
	results := []ProductPrepareResult{}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowNumber := i + 1 // เพิ่ม 1 เพราะนับรวม header

		rowResults := processProductRow(session.HoldingCode, rowNumber, headers, row)

		// ถ้า barcode มีหลายตัว (แบ่งด้วย /) ให้แยกออกเป็นหลายแถว
		expandedResults := expandBarcodeRows(rowResults)
		results = append(results, expandedResults...)

		// อัปเดตความคืบหน้า
		session.Mutex.Lock()
		session.ProcessedRows = i
		session.Progress = float64(i) / float64(totalRows) * 100

		// นับตามแถวต้นฉบับ (ไม่นับแถวที่แยกออกมา)
		if rowResults.Status == "success" || rowResults.Status == "warning" {
			session.SuccessCount++
		} else if rowResults.Status == "error" {
			session.ErrorCount++
		}
		session.Mutex.Unlock()

		// Log ทุก 100 แถว
		if i%100 == 0 {
			logger.Info("Processed %d/%d rows (%.1f%%)", i, totalRows, session.Progress)
		}
	}

	// เสร็จสิ้น - เก็บผลลัพธ์และเปรียบเทียบกับ MongoDB
	session.Mutex.Lock()
	session.Status = StatusCompleted
	session.Progress = 100
	session.Result = results
	endTime := time.Now()
	session.EndTime = &endTime
	session.Mutex.Unlock()

	// กรองเฉพาะแถวที่ parse สำเร็จ (success หรือ warning) สำหรับเปรียบเทียบกับ MongoDB
	successResults := []ProductPrepareResult{}
	for _, r := range results {
		if r.Status == "success" || r.Status == "warning" {
			successResults = append(successResults, r)
		}
	}

	// ตรวจสอบ barcode ซ้ำ
	logger.Info("Checking for duplicate barcodes in %d valid rows", len(successResults))
	duplicates := checkDuplicateBarcodes(successResults)
	if len(duplicates) > 0 {
		// มี barcode ซ้ำ - สร้าง error message พร้อม row numbers
		duplicateList := []string{}
		for _, dup := range duplicates {
			// แสดง barcode, จำนวนครั้ง และ row numbers
			rowNumStr := make([]string, len(dup.RowNumbers))
			for i, rowNum := range dup.RowNumbers {
				rowNumStr[i] = fmt.Sprintf("%d", rowNum)
			}
			duplicateList = append(duplicateList,
				fmt.Sprintf("%s (ซ้ำ %d ครั้ง ที่แถว: %s)",
					dup.Barcode, dup.Count, strings.Join(rowNumStr, ", ")))
		}

		errorMsg := fmt.Sprintf("พบ barcode ซ้ำกัน %d รายการ: %s",
			len(duplicates), strings.Join(duplicateList, "; "))

		logger.Error(errorMsg)

		session.Mutex.Lock()
		session.Status = StatusError
		session.ErrorMessage = errorMsg
		session.DuplicateBarcodes = duplicates                     // เก็บ JSON structure ด้วย
		session.ErrorCount = len(duplicates)                       // นับ error = จำนวน barcode ที่ซ้ำ
		session.SuccessCount = session.TotalRows - len(duplicates) // ลด success count
		session.Mutex.Unlock()
		return
	}
	logger.Info("No duplicate barcodes found")

	// เปรียบเทียบกับ MongoDB
	logger.Info("Starting MongoDB comparison for shop %s with %d valid rows", session.HoldingCode, len(successResults))
	comparisonSummary, err := compareWithMongo(session.HoldingCode, successResults)
	if err != nil {
		logger.Error("Failed to compare with MongoDB: %v", err)
	} else {
		// เก็บผลการเปรียบเทียบใน session
		session.Mutex.Lock()
		session.ComparisonResult = comparisonSummary
		session.Mutex.Unlock()

		logger.Info("Comparison completed: New=%d, Updated=%d, Unchanged=%d",
			comparisonSummary.NewProductCount,
			comparisonSummary.UpdatedProductCount,
			comparisonSummary.UnchangedCount)
	}

	duration := endTime.Sub(session.StartTime)
	logger.Success("Excel processing completed for shop %s: %d rows in %s (Success: %d, Error: %d)",
		session.HoldingCode, totalRows, duration, session.SuccessCount, session.ErrorCount)

	// ลบไฟล์ Excel หลังประมวลผลเสร็จ (เป็น temp copy จาก S3)
	if session.FilePath != "" {
		os.Remove(session.FilePath)
		logger.Info("Cleaned up temp Excel file: %s", session.FilePath)
	}
}

// expandBarcodeRows - แยก barcode ที่มีหลายตัว (คั่นด้วย /) ออกเป็นหลายแถว
func expandBarcodeRows(result ProductPrepareResult) []ProductPrepareResult {
	// ถ้า error ให้ return เลย
	if result.Status == "error" {
		return []ProductPrepareResult{result}
	}

	// ตรวจสอบว่า barcode มี / หรือไม่
	if !strings.Contains(result.Barcode, "/") {
		return []ProductPrepareResult{result}
	}

	// แยก barcode ด้วย /
	barcodes := strings.Split(result.Barcode, "/")
	expandedResults := []ProductPrepareResult{}

	for _, barcode := range barcodes {
		barcode = strings.TrimSpace(barcode)
		if barcode == "" {
			continue
		}

		// Copy ข้อมูลทั้งหมด
		newResult := result
		newResult.Barcode = barcode

		// Copy Data map
		newResult.Data = make(map[string]interface{})
		for k, v := range result.Data {
			newResult.Data[k] = v
		}

		expandedResults = append(expandedResults, newResult)
	}

	if len(expandedResults) > 1 {
		barcodeList := []string{}
		for _, r := range expandedResults {
			barcodeList = append(barcodeList, r.Barcode)
		}
		logger.Info("Expanded row %d: '%s' -> %d barcodes: %v", result.RowNumber, result.Barcode, len(expandedResults), barcodeList)
	}

	return expandedResults
}

// processProductRow - ประมวลผลแต่ละแถว
func processProductRow(holdingCode string, rowNumber int, headers []string, row []string) ProductPrepareResult {
	result := ProductPrepareResult{
		RowNumber: rowNumber,
		Status:    "success",
		Data:      make(map[string]interface{}),
	}

	// สร้าง map จาก headers และ row (trim space ทั้ง header และ value)
	rowData := make(map[string]string)
	for i, header := range headers {
		trimmedHeader := strings.TrimSpace(header)
		if i < len(row) {
			rowData[trimmedHeader] = strings.TrimSpace(row[i])
		} else {
			rowData[trimmedHeader] = ""
		}
	}

	// Debug: Log row data สำหรับแถวที่ 2 เท่านั้น (เพื่อดู column names)
	if rowNumber == 2 {
		logger.Info("Sample row data (row 2):")
		for k, v := range rowData {
			if v != "" {
				logger.Info("  [%s] = %s", k, v)
			}
		}
	}

	// Map ข้อมูลตามโครงสร้างที่กำหนด
	result.Barcode = getValueFromRow(rowData, []string{"Barcode", "barcode", "รหัสบาร์โค้ด"})

	// Debug: log barcode value สำหรับแถวแรก
	if rowNumber == 2 {
		logger.Info("Barcode value from row 2: '%s'", result.Barcode)
		if result.Barcode == "" {
			// Log ชื่อ columns ทั้งหมด
			cols := []string{}
			for k := range rowData {
				cols = append(cols, k)
			}
			logger.Warn("Barcode is empty! Available columns: %v", cols)
		}
	}

	result.Name = getValueFromRow(rowData, []string{"Name", "name", "ชื่อสินค้า"})
	result.UnitCode = getValueFromRow(rowData, []string{"Unit Code", "UnitCode", "unit_code", "รหัสหน่วย"})
	result.ProductType = getValueFromRow(rowData, []string{"Product Type", "ProductType", "product_type", "ประเภทสินค้า"})
	result.TaxType = getValueFromRow(rowData, []string{"Tax Type", "TaxType", "tax_type", "ภาษี"})
	result.Code = getValueFromRow(rowData, []string{"itemcode", "item_code", "ItemCode", "Code", "code", "รหัสสินค้า"})

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
	result.StandValue = parseFloat(getValueFromRow(rowData, []string{"standvalue", "StandValue", "stand_value", "StandValue"}))
	result.DivideValue = parseFloat(getValueFromRow(rowData, []string{"dividevalue", "DivideValue", "divide_value", "DivideValue"}))

	// Validation - Barcode เป็น required field (สำหรับ compare กับ MongoDB)
	if result.Barcode == "" {
		result.Status = "error"
		result.Message = "Barcode is required"
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

// checkDuplicateBarcodes - ตรวจสอบ barcode ซ้ำและคืน array ของข้อมูล duplicate พร้อม row numbers
func checkDuplicateBarcodes(results []ProductPrepareResult) []DuplicateBarcodeInfo {
	barcodeMap := make(map[string][]int) // key=barcode, value=array ของ row numbers

	// เก็บ row numbers ของแต่ละ barcode
	for _, r := range results {
		barcodeMap[r.Barcode] = append(barcodeMap[r.Barcode], r.RowNumber)
	}

	// กรองเฉพาะ barcode ที่ปรากฏมากกว่า 1 ครั้ง
	duplicates := []DuplicateBarcodeInfo{}
	for barcode, rowNumbers := range barcodeMap {
		if len(rowNumbers) > 1 {
			duplicates = append(duplicates, DuplicateBarcodeInfo{
				Barcode:    barcode,
				Count:      len(rowNumbers),
				RowNumbers: rowNumbers,
			})
		}
	}

	return duplicates
}

// loadMongoProducts - ดึงข้อมูลสินค้าจาก MongoDB
func loadMongoProducts(holdingCode string) ([]models.MongoProductBarcodeModel, error) {
	ctx := context.Background()

	// Get MongoDB client
	mongoClient, err := myglobal.MongoConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect MongoDB: %v", err)
	}

	// ดึงชื่อ database จาก config
	svcConfig := config.NewServiceConfig()
	databaseName := svcConfig.MongodbDatabaseName()

	// ใช้ database และ collection
	collection := mongoClient.Database(databaseName).Collection("productBarcodes")

	// Query โดยใช้ holdingCode
	filter := bson.M{"holding_code": holdingCode}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	var products []models.MongoProductBarcodeModel
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode MongoDB results: %v", err)
	}

	logger.Info("Loaded %d products from MongoDB for shop %s", len(products), holdingCode)

	// Debug: แสดง sample product เพื่อตรวจสอบค่า
	if len(products) > 0 {
		sample := products[0]
		logger.Info("Sample MongoDB product: Barcode=%s, Code=%s, DivideValue=%.2f, StandValue=%.2f",
			sample.Barcode, sample.ItemCode, sample.DivideValue, sample.StandValue)
	}

	return products, nil
}

// compareWithMongo - เปรียบเทียบข้อมูล Excel กับ MongoDB (ใช้ Excel เป็นหลัก)
func compareWithMongo(holdingCode string, excelResults []ProductPrepareResult) (*ComparisonSummary, error) {
	// ดึงข้อมูลจาก MongoDB
	mongoProducts, err := loadMongoProducts(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("failed to load MongoDB products: %v", err)
	}

	// สร้าง map จาก MongoDB โดยใช้ barcode เป็น key
	mongoMap := make(map[string]*models.MongoProductBarcodeModel)
	for i := range mongoProducts {
		mongoMap[mongoProducts[i].Barcode] = &mongoProducts[i]
	}

	logger.Info("MongoDB map created: %d products", len(mongoMap))
	logger.Info("Excel data: %d rows", len(excelResults))

	// Debug: แสดง sample barcode จาก Excel (5 ตัวแรก)
	if len(excelResults) > 0 {
		logger.Info("Sample Excel barcodes:")
		for i := 0; i < len(excelResults) && i < 5; i++ {
			logger.Info("  [%d] %s", i+1, excelResults[i].Barcode)
		}
	}

	summary := &ComparisonSummary{
		HoldingCode:        holdingCode,
		TotalExcelRows:     len(excelResults),
		TotalMongoProducts: len(mongoProducts),
		Products:           []ProductCompact{},
		ProcessTime:        time.Now(),
	}

	// Loop Excel เป็นหลัก
	matchCount := 0
	for _, excelRow := range excelResults {
		if excelRow.Barcode == "" {
			continue // ข้าม barcode ว่าง
		}

		// ข้อมูลจาก Excel
		excelData := &ProductCompactData{
			Code:        excelRow.Code,
			Name:        excelRow.Name,
			UnitCode:    excelRow.UnitCode,
			DivideValue: excelRow.DivideValue,
			StandValue:  excelRow.StandValue,
		}

		product := ProductCompact{
			Barcode: excelRow.Barcode,
			Excel:   excelData,
			Mongo:   nil, // default ไม่มี MongoDB
		}

		// ตรวจสอบว่ามีใน MongoDB หรือไม่
		if mongoProduct, exists := mongoMap[excelRow.Barcode]; exists {
			matchCount++

			// ดึง Thai name จาก MongoDB
			mongoThaiName := ""
			for _, name := range mongoProduct.Names {
				if name.Code == "th" {
					mongoThaiName = name.Name
					break
				}
			}

			// สร้างข้อมูล MongoDB
			mongoData := &ProductCompactData{
				Code:        mongoProduct.ItemCode,
				Name:        mongoThaiName,
				UnitCode:    mongoProduct.ItemUnitCode,
				DivideValue: mongoProduct.DivideValue,
				StandValue:  mongoProduct.StandValue,
			}
			product.Mongo = mongoData

			// เปรียบเทียบ 5 fields
			hasChanges := false

			if excelRow.Name != mongoThaiName {
				hasChanges = true
			}
			if excelRow.Code != mongoProduct.ItemCode {
				hasChanges = true
			}
			if excelRow.UnitCode != mongoProduct.ItemUnitCode {
				hasChanges = true
			}
			if excelRow.DivideValue != mongoProduct.DivideValue {
				hasChanges = true
			}
			if excelRow.StandValue != mongoProduct.StandValue {
				hasChanges = true
			}

			if hasChanges {
				product.Action = 2 // UPDATE - มีทั้ง Excel และ MongoDB แต่ข้อมูลไม่ตรงกัน
				summary.UpdatedProductCount++
			} else {
				product.Action = 0 // MATCH - ข้อมูลเหมือนกันทุกอย่าง
				summary.UnchangedCount++
			}
		} else {
			// ไม่มีใน MongoDB - ต้อง INSERT
			product.Action = 1 // INSERT - มีใน Excel แต่ไม่มีใน MongoDB
			summary.NewProductCount++
		}

		summary.Products = append(summary.Products, product)
	}

	logger.Info("Excel barcode matches with MongoDB: %d out of %d", matchCount, len(excelResults))
	logger.Info("Comparison summary: Insert=%d, Update=%d, Match=%d",
		summary.NewProductCount, summary.UpdatedProductCount, summary.UnchangedCount)

	return summary, nil
}

// downloadFileFromURL - download ไฟล์จาก URL (presigned S3 URL) มาเก็บที่ local path
func downloadFileFromURL(fileURL string, destPath string) error {
	// สร้าง directory ถ้ายังไม่มี
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(destPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	logger.Info("Downloaded %d bytes to %s", written, destPath)
	return nil
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
