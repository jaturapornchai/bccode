package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/process/build"
	processdoc "smlcloudplatform/internal/goapi/process/process-doc"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"
	reportstock "smlcloudplatform/internal/goapi/process/report-stock"

	"smlcloudplatform/internal/goapi/myglobal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

// RebuildProgressSSEHandler — SSE endpoint สำหรับ stream progress ของ rebuild job
func RebuildProgressSSEHandler(c echo.Context) error {
	jobId := c.Param("jobId")
	job := build.GetJob(jobId)
	if job == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "job not found"})
	}

	// SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)

	w := c.Response().Writer
	flusher, hasFlusher := w.(http.Flusher)

	flushWriter := func() {
		if hasFlusher {
			flusher.Flush()
		}
	}

	// ยกเลิก WriteTimeout สำหรับ SSE connection
	// (ป้องกัน Go HTTP Server ตัด connection หลัง 300 วินาที)
	rc := http.NewResponseController(c.Response())
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		logger.Warn("[SSE] ไม่สามารถ reset write deadline: %v", err)
	}

	logger.Info("[SSE] Client เชื่อมต่อ rebuild progress job %s", jobId)

	// ส่ง initial event เพื่อยืนยันว่า connection สำเร็จ
	fmt.Fprintf(w, "data: {\"status\":\"connected\",\"job_id\":\"%s\",\"message\":\"เชื่อมต่อสำเร็จ\"}\n\n", jobId)
	flushWriter()

	// Heartbeat ticker — ส่ง SSE comment ทุก 15 วินาทีเพื่อ keep connection alive
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	// อ่าน events จาก job.EventsCh แล้วส่งเป็น SSE
	for {
		select {
		case event, ok := <-job.EventsCh:
			if !ok {
				// Channel ถูกปิด — job เสร็จแล้ว
				logger.Info("[SSE] Job %s channel ถูกปิด", jobId)
				return nil
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flushWriter()

			// ถ้า completed หรือ error — ปิด connection
			if event.Status == "completed" || event.Status == "error" {
				logger.Info("[SSE] Job %s จบด้วยสถานะ: %s", jobId, event.Status)
				return nil
			}
		case <-heartbeat.C:
			// ส่ง SSE comment เป็น keepalive (client จะข้ามบรรทัดที่ขึ้นต้นด้วย ":")
			fmt.Fprintf(w, ": keepalive\n\n")
			flushWriter()
		case <-c.Request().Context().Done():
			// Client disconnect
			logger.Info("[SSE] Client disconnect จาก job %s", jobId)
			return nil
		}
	}
}

// uploadReportBinToS3 - upload .bin file ไป S3 แล้วลบ local temp files (.bin + .pdf)
// คืน S3 object key สำหรับ download ภายหลัง
func uploadReportBinToS3(localBinPath string) string {
	if localBinPath == "" {
		return ""
	}

	client, err := GetR2Client()
	if err != nil {
		logger.Warn("S3 not available, keeping local path: %v", err)
		return localBinPath
	}

	// อ่านไฟล์ .bin
	fileBytes, err := os.ReadFile(localBinPath)
	if err != nil {
		logger.Error("Failed to read .bin file: %v", err)
		return localBinPath
	}

	// สร้าง S3 key: reports/{filename}
	filename := filepath.Base(localBinPath)
	objectKey := "reports/" + filename

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r2BucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		logger.Error("Failed to upload .bin to S3: %v", err)
		return localBinPath
	}

	logger.Info("Report .bin uploaded to S3: %s (%d bytes)", objectKey, len(fileBytes))

	// ลบ local temp files (.bin + .pdf)
	os.Remove(localBinPath)
	pdfPath := strings.TrimSuffix(localBinPath, ".bin") + ".pdf"
	os.Remove(pdfPath)

	return objectKey
}

// Report Post Handler - handles report generation commands
func ReportPostHandler(c echo.Context) error {
	logger.Info("Report endpoint hit")

	// รับ JSON โดยตรงจาก request body
	jsonBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("reading request body: %v", err)
		return c.String(http.StatusBadRequest, "Invalid request body")
	}
	logger.Info("Received payload: %s", string(jsonBody))

	// ตรวจสอบว่า JSON ถูก encode เป็น string หรือไม่
	var rawJSON string
	err = json.Unmarshal(jsonBody, &rawJSON)
	if err == nil {
		// กรณีที่ JSON ถูก encode เป็น string
		// ใช้ค่า rawJSON แทน jsonBody
		jsonBody = []byte(rawJSON)
	}

	var payLoad models.PayLoadCommandStruct
	err = json.Unmarshal(jsonBody, &payLoad)
	if err != nil {
		logger.Error("unmarshaling JSON: %v", err)
		return c.String(http.StatusBadRequest, "Invalid JSON format")
	}

	logger.Info("Received payload: %+v", payLoad)

	// ใช้ TimezoneCode จาก payload หรือใช้ค่าเริ่มต้นเป็น "TH"
	timezoneCode := payLoad.TimezoneCode
	if timezoneCode == "" || len(timezoneCode) < 2 {
		timezoneCode = "TH"
	}

	// ใช้ LanguageCode จาก payload หรือใช้ค่าเริ่มต้นเป็น "th"
	languageCode := payLoad.LanguageCode
	if languageCode == "" || len(languageCode) < 2 {
		languageCode = "th"
	}

	logger.Info("Using Timezone Code: %s, Language Code: %s", timezoneCode, languageCode)

	if payLoad.CommandID == "rebuild" {
		// Default: create_database = true (drop + create ใหม่)
		createDatabase := true
		if payLoad.CreateDatabase != nil {
			createDatabase = *payLoad.CreateDatabase
		}

		// Parse item_code_list (optional - ใช้เฉพาะเมื่อ create_database = false)
		itemCodes, _ := parseItemCodesFromJSON(payLoad.ItemCodeList)

		// สร้าง rebuild job สำหรับ progress tracking
		job := build.CreateJob(payLoad.ShopID)

		go func() {
			defer func() {
				// รอให้ SSE client มีเวลาอ่าน events ที่เหลือก่อนลบ job
				time.Sleep(10 * time.Second)
				build.RemoveJob(job.ID)
			}()
			if createDatabase {
				// Full rebuild: drop database + create ใหม่ทั้งหมด (พร้อม progress)
				build.PgSqlDropDatabaseAndReProcessWithProgress(payLoad.ShopID, job)
			} else {
				// Calc only mode: คำนวณสต็อกและสถานะเอกสารเท่านั้น
				if len(itemCodes) > 0 {
					// เฉพาะ items ที่ระบุ (พร้อม progress)
					build.CalcStockCostForItemsWithProgress(payLoad.ShopID, itemCodes, job)
				} else {
					// ทุก items (พร้อม progress)
					build.CalcStockCostAllWithProgress(payLoad.ShopID, job)
				}
			}
		}()

		return c.JSON(http.StatusOK, map[string]any{
			"message":         "rebuild started",
			"status":          "success",
			"code":            200,
			"job_id":          job.ID,
			"shop_id":         payLoad.ShopID,
			"command_id":      payLoad.CommandID,
			"create_database": createDatabase,
			"item_count":      len(itemCodes),
		})
	}

	if payLoad.CommandID == "rebuild_document_flow" {
		// คำนวณ flow เอกสาร (isref, iscomparedsuccess, isclosed)
		// สำหรับคำนวณสถานะเอกสารใบสั่งซื้อว่ามีการอ้างอิง (รับสินค้า) หรือยัง
		job := build.CreateJob(payLoad.ShopID)

		go func() {
			defer func() {
				time.Sleep(10 * time.Second)
				build.RemoveJob(job.ID)
			}()
			build.RebuildDocumentFlowWithProgress(payLoad.ShopID, job)
		}()

		return c.JSON(http.StatusOK, map[string]any{
			"message":    "rebuild document flow started",
			"status":     "success",
			"code":       200,
			"job_id":     job.ID,
			"shop_id":    payLoad.ShopID,
			"command_id": payLoad.CommandID,
		})
	}

	if payLoad.CommandID == "rebuild_products_only" || payLoad.CommandID == "rebuild-products" {
		// Rebuild เฉพาะสินค้า (PostgreSQL, ClickHouse)
		// รองรับทั้ง "rebuild_products_only" และ "rebuild-products"
		go func() {
			err := build.RebuildProductsOnly(payLoad.ShopID)
			if err != nil {
				logger.Error("Failed to rebuild products: %v", err)
			}
		}()
		return c.JSON(http.StatusOK, map[string]any{
			"message":    "rebuild products only started",
			"status":     "success",
			"code":       200,
			"shop_id":    payLoad.ShopID,
			"command_id": payLoad.CommandID,
		})
	}

	if strings.EqualFold(payLoad.CommandID, processStockCalcCostCommandID) {
		itemCodes, err := parseItemCodesFromJSON(payLoad.ItemCodeList)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": err.Error(),
				"code":  "INVALID_ITEM_CODE_LIST",
			})
		}

		pointQty := resolvePoint(payLoad.PointQty, myglobal.ConfigSystem.StockQtyPoint)
		pointAmount := resolvePoint(payLoad.PointAmount, myglobal.ConfigSystem.StockAmountPoint)
		pointCost := resolvePoint(payLoad.PointCost, myglobal.ConfigSystem.StockCostPoint)
		deleteFirst := true
		if payLoad.DeleteFirst != nil {
			deleteFirst = *payLoad.DeleteFirst
		}

		results, err := runProcessStockCalcCost(payLoad.ShopID, itemCodes, pointQty, pointAmount, pointCost, deleteFirst, false, false)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
				"code":  "PROCESS_STOCK_COST_ERROR",
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"message":         "process stock cost",
			"status":          "success",
			"code":            200,
			"shop_id":         payLoad.ShopID,
			"command_id":      payLoad.CommandID,
			"processed_items": len(results),
			"point_qty":       pointQty,
			"point_amount":    pointAmount,
			"point_cost":      pointCost,
			"delete_first":    deleteFirst,
			"items":           results,
		})
	}

	if payLoad.CommandID == "report_product_balance_by_whcode_barcode" {
		logger.Info("Command: %s, ShopID: %s, GUID: %s", payLoad.CommandID, payLoad.ShopID, payLoad.Guid)

		ReportProductBalanceByWareHouseBarcode(payLoad.ShopID, payLoad.Guid, payLoad.FinalDate, timezoneCode, languageCode)

		return c.JSON(http.StatusOK, map[string]any{
			"message":    "report completed",
			"status":     "success",
			"code":       200,
			"shop_id":    payLoad.ShopID,
			"command_id": payLoad.CommandID,
			"guid":       payLoad.Guid,
		})
	}

	if payLoad.CommandID == "report_product_balance_by_location_barcode" {
		logger.Info("Command: %s, ShopID: %s, GUID: %s", payLoad.CommandID, payLoad.ShopID, payLoad.Guid)

		ReportProductBalanceByLocationBarcode(payLoad.ShopID, payLoad.Guid, payLoad.FinalDate, timezoneCode, languageCode)

		return c.JSON(http.StatusOK, map[string]any{
			"message":    "report completed",
			"status":     "success",
			"code":       200,
			"shop_id":    payLoad.ShopID,
			"command_id": payLoad.CommandID,
			"guid":       payLoad.Guid,
		})
	}

	if payLoad.CommandID == "report_product_balance_by_barcode_whcode_location" {
		// Convert condition string to int
		conditionInt := 0
		if payLoad.Condition != "" {
			// Add conversion logic here if needed
		}

		logger.Info("Command: %s, ShopID: %s, GUID: %s, Condition: %d", payLoad.CommandID, payLoad.ShopID, payLoad.Guid, conditionInt)

		ReportProductBalanceByBarcodeWhCodeLocationCode(payLoad.ShopID, payLoad.Guid, conditionInt, payLoad.FinalDate, timezoneCode, languageCode)

		return c.JSON(http.StatusOK, map[string]any{
			"message":    "report completed",
			"status":     "success",
			"code":       200,
			"shop_id":    payLoad.ShopID,
			"command_id": payLoad.CommandID,
			"guid":       payLoad.Guid,
		})
	}

	if payLoad.CommandID == "report_product_stock_movement" {
		logger.Info("Command: %s, ShopID: %s, GUID: %s", payLoad.CommandID, payLoad.ShopID, payLoad.Guid)

		ReportProductStockMovement(payLoad.ShopID, payLoad.Guid, timezoneCode, languageCode)

		return c.JSON(http.StatusOK, map[string]any{
			"message":    "report completed",
			"status":     "success",
			"code":       200,
			"shop_id":    payLoad.ShopID,
			"command_id": payLoad.CommandID,
			"guid":       payLoad.Guid,
		})
	}

	if payLoad.CommandID == "stock_balance_by_product_and_warehouse_and_location_create_pdf" {
		// สร้าง Report PDF
		// รายงานสินค้าคงเหลือ ตามบาร์โค้ด คลังสินค้า ที่เก็บสินค้า
		condition, _ := strconv.Atoi(payLoad.Condition)
		balanceOnly, _ := strconv.ParseBool(payLoad.BalanceOnly)
		logger.Info("Command: %s, Shop ID: %s", payLoad.CommandID, payLoad.ShopID)
		logger.Info("Condition: %d", condition)
		logger.Info("Balance Only: %t", balanceOnly)
		logger.Info("Condition: %d, Final Date: %s", condition, payLoad.FinalDate)
		// สร้าง Report
		logger.Info("GUID: %v", payLoad.Guid)
		// สร้าง Report
		localBin := reportstock.ReportProductBalanceByItemAndWareHouseAndLocation(payLoad.ShopID, payLoad.Guid, condition, payLoad.FinalDate, timezoneCode, languageCode)
		reportPath := uploadReportBinToS3(localBin)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    payLoad.Guid,
			"path":    reportPath,
		})
	}

	if payLoad.CommandID == "stock_balance_by_warehouse_and_product_create_pdf" {
		// สร้าง Report PDF
		// รายงานสินค้าคงเหลือ คลังสินค้า ตามบาร์โค้ด
		condition, _ := strconv.Atoi(payLoad.Condition)
		balanceOnly, _ := strconv.ParseBool(payLoad.BalanceOnly)
		logger.Info("Command: %s, Shop ID: %s", payLoad.CommandID, payLoad.ShopID)
		logger.Info("Condition: %d", condition)
		logger.Info("Balance Only: %t", balanceOnly)
		logger.Info("Condition: %d, Final Date: %s", condition, payLoad.FinalDate)
		logger.Info("GUID: %v", payLoad.Guid)
		localBin := reportstock.ReportProductBalanceByWareHouseAndItem(payLoad.ShopID, payLoad.Guid, payLoad.FinalDate, timezoneCode, languageCode)
		reportPath := uploadReportBinToS3(localBin)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    payLoad.Guid,
			"path":    reportPath,
		})
	}

	if payLoad.CommandID == "stock_balance_by_location_and_product_create_pdf" {
		// สร้าง Report PDF
		// รายงานสินค้าคงเหลือ ที่เก็บสินค้า ตามบาร์โค้ด
		balanceOnly, _ := strconv.ParseBool(payLoad.BalanceOnly)
		logger.Info("Command: %s, Shop ID: %s", payLoad.CommandID, payLoad.ShopID)
		logger.Info("Balance Only: %t", balanceOnly)
		logger.Info("GUID: %v", payLoad.Guid)
		localBin := reportstock.ReportProductBalanceByLocationAndItem(payLoad.ShopID, payLoad.Guid, payLoad.FinalDate, timezoneCode, languageCode)
		reportPath := uploadReportBinToS3(localBin)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    payLoad.Guid,
			"path":    reportPath,
		})
	}

	if payLoad.CommandID == "report_product_stock_movement_create_pdf" {
		localBin := reportstock.ReportProductStockMovement(payLoad.ShopID, payLoad.Guid, timezoneCode, languageCode)
		reportPath := uploadReportBinToS3(localBin)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    payLoad.Guid,
			"path":    reportPath,
		})
	}

	if payLoad.CommandID == "stock_balance_by_product_and_warehouse_and_location_process" {
		// ประมวลผล รายงานสินค้าคงเหลือ ตามบาร์โค้ด คลังสินค้า ที่เก็บสินค้า
		condition, _ := strconv.Atoi(fmt.Sprintf("%v", payLoad.Condition))
		balanceOnly, _ := strconv.ParseBool(fmt.Sprintf("%v", payLoad.BalanceOnly))
		barcodeListArray := strings.Split(fmt.Sprintf("%v", payLoad.BarcodeList), ",")
		barcodeList := make([]string, 0)
		for _, barcode := range barcodeListArray {
			barcode = strings.TrimSpace(barcode)
			if barcode != "" {
				barcodeList = append(barcodeList, barcode)
			}
		}

		logger.Info("Barcode List: %v", barcodeList)

		finalDate := fmt.Sprintf("%v", payLoad.FinalDate)
		logger.Info("Condition: %d, Final Date: %s", condition, finalDate)

		// ใช้ TimezoneCode จาก payload หรือใช้ค่าเริ่มต้นเป็น "TH"
		timezoneCode := payLoad.TimezoneCode
		if timezoneCode == "" {
			timezoneCode = "TH"
		}
		logger.Info("Using Timezone Code: %s", timezoneCode)

		// ใช้ LanguageCode จาก payload หรือใช้ค่าเริ่มต้นเป็น "th"
		languageCode := payLoad.LanguageCode
		if languageCode == "" {
			languageCode = "th"
		}
		logger.Info("Using Language Code: %s", languageCode)

		// สร้าง Report
		// รายงานสินค้าคงเหลือ ตามบาร์โค้ด คลังสินค้า ที่เก็บสินค้า
		prepareReport := processstock.ProcessProductBalanceByItemAndWareHouseAndLocationWithTimezone(payLoad.ShopID, condition, finalDate, balanceOnly, barcodeList, payLoad.WarehouseList, timezoneCode)
		logger.Info("PrepareReport: %+v", prepareReport)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    prepareReport.Guid,
		})
	}

	if payLoad.CommandID == "stock_balance_by_warehouse_and_product_process" {
		// ประมวลผล รายงานสินค้าคงเหลือ ตามบาร์โค้ด คลังสินค้า ที่เก็บสินค้า
		balanceOnly, _ := strconv.ParseBool(fmt.Sprintf("%v", payLoad.BalanceOnly))
		barcodeListArray := strings.Split(fmt.Sprintf("%v", payLoad.BarcodeList), ",")
		barcodeList := make([]string, 0)
		for _, barcode := range barcodeListArray {
			barcode = strings.TrimSpace(barcode)
			if barcode != "" {
				barcodeList = append(barcodeList, barcode)
			}
		}

		logger.Info("Barcode List: %v", barcodeList)

		finalDate := fmt.Sprintf("%v", payLoad.FinalDate)
		prepareReport := processstock.ProcessProductBalanceByWareHouseAndItem(payLoad.ShopID, finalDate, balanceOnly, barcodeList, payLoad.WarehouseList)
		logger.Info("PrepareReport: %+v", prepareReport)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    prepareReport.Guid,
		})
	}

	if payLoad.CommandID == "stock_balance_by_location_and_product_process" {
		// ดำเนินการตามคำสั่งที่ต้องการ
		condition, _ := strconv.Atoi(fmt.Sprintf("%v", payLoad.Condition))
		balanceOnly, _ := strconv.ParseBool(fmt.Sprintf("%v", payLoad.BalanceOnly))
		barcodeListArray := strings.Split(fmt.Sprintf("%v", payLoad.BarcodeList), ",")
		barcodeList := make([]string, 0)
		for _, barcode := range barcodeListArray {
			barcode = strings.TrimSpace(barcode)
			if barcode != "" {
				barcodeList = append(barcodeList, barcode)
			}
		}

		logger.Info("Barcode List: %v", barcodeList)

		finalDate := fmt.Sprintf("%v", payLoad.FinalDate)
		logger.Info("Condition: %d, Final Date: %s", condition, finalDate)
		// สร้าง Report
		// รายงานสินค้าคงเหลือ ตามบาร์โค้ด คลังสินค้า ที่เก็บสินค้า
		prepareReport := processstock.ProcessProductBalanceByLocationAndItem(payLoad.ShopID, finalDate, balanceOnly, barcodeList, payLoad.WarehouseList)
		logger.Info("ProcessProductBalanceByLocationCodeBarcode PrepareReport: %+v", prepareReport)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    prepareReport.Guid,
		})
	}

	if payLoad.CommandID == "stock_product_movement_and_cost_process" {
		// เคลื่อนไหวสินค้า/ต้นทุน
		condition, _ := strconv.Atoi(fmt.Sprintf("%v", payLoad.Condition))
		movementOnly, _ := strconv.ParseBool(fmt.Sprintf("%v", payLoad.MovementOnly))
		itemCodeList, err := parseItemCodesFromJSON(payLoad.ItemCodeList)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": err.Error(),
				"code":  "INVALID_ITEM_CODE_LIST",
			})
		}

		logger.Info("Item Code List: %v", itemCodeList)

		fromDate := fmt.Sprintf("%v", payLoad.FromDate)
		finalDate := fmt.Sprintf("%v", payLoad.FinalDate)
		logger.Info("Condition: %d, Final Date: %s", condition, finalDate)
		// สร้าง Report
		// รายงานสินค้าคงเหลือ ตามบาร์โค้ด คลังสินค้า ที่เก็บสินค้า
		prepareReport := processstock.ProcessProductMovement(payLoad.ShopID, fromDate, finalDate, movementOnly, itemCodeList, payLoad.WarehouseList)
		logger.Info("PrepareReport: %+v", prepareReport)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"guid":    prepareReport.Guid,
		})
	}
	logger.Info("payLoad.CommandID: %v", payLoad.CommandID)
	if payLoad.CommandID == "purchase_order_info_by_docno" {
		// สถานะเอกสารใบสั่งซื้อ (PO Status)

		// แปลง DocNoList จาก []string โดยตรง
		docNoList := payLoad.DocNumberList

		logger.Info("Document Number List: %v", docNoList)

		if len(docNoList) == 0 {
			logger.Info("No valid document numbers provided")
			return c.JSON(http.StatusBadRequest, map[string]any{
				"message": "No valid document numbers provided",
				"status":  "error",
				"code":    400,
			})
		}

		prepareReport := processdoc.PurchaseStatusByDocNo(payLoad.ShopID, docNoList)
		logger.Info("PrepareReport: %+v", prepareReport)

		return c.JSON(http.StatusOK, map[string]any{
			"message": "Report",
			"status":  "success",
			"code":    200,
			"data":    prepareReport,
		})
	}

	// Default response for unknown commands
	return c.JSON(http.StatusOK, map[string]any{
		"message": "Report",
		"status":  "success",
		"code":    200,
	})
}
