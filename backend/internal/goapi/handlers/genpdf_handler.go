package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	serviceConfig "smlcloudplatform/internal/goapi/config"
	gentranspdf "smlcloudplatform/internal/goapi/handlers/gen-trans-pdf"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GenPDFHandler - สร้าง PDF จากเอกสารใน MongoDB, upload ไป R2, เก็บประวัติ
func GenPDFHandler(c echo.Context) error {
	logger.Info("-> GenPDFHandler called")

	// รับ JSON payload จาก request body
	var payload gentranspdf.GenPDFPayload

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid JSON payload",
		})
	}

	// Validate required parameters
	if payload.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Missing required parameter: shopid",
		})
	}

	if payload.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Missing required parameter: collection",
		})
	}

	if payload.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Missing required parameter: docno",
		})
	}

	// Set default values
	if payload.Title == "" {
		payload.Title = "เอกสาร"
	}
	if payload.PageSize == "" {
		payload.PageSize = "A4"
	}
	if payload.Orientation == "" {
		payload.Orientation = "P" // Portrait
	}
	if payload.FontSize == 0 {
		payload.FontSize = 9 // Default font size
	}
	// ColorMode default is true (color)
	if payload.ColorMode == nil {
		colorModeDefault := true
		payload.ColorMode = &colorModeDefault
	}

	// Validate orientation
	payload.Orientation = strings.ToUpper(payload.Orientation)
	if payload.Orientation != "P" && payload.Orientation != "L" {
		payload.Orientation = "P"
	}

	// Validate page size
	payload.PageSize = strings.ToUpper(payload.PageSize)
	validPageSizes := map[string]bool{"A4": true, "A3": true, "A5": true, "LETTER": true, "LEGAL": true}
	if !validPageSizes[payload.PageSize] {
		payload.PageSize = "A4"
	}

	// Validate font size (8-16)
	if payload.FontSize < 8 {
		payload.FontSize = 8
	} else if payload.FontSize > 16 {
		payload.FontSize = 16
	}

	// Set default date format if not specified
	if payload.DateFormat == "" {
		payload.DateFormat = "DD/MM/YYYY"
	}

	// Set default language if not specified
	if payload.Language == "" {
		payload.Language = "th"
	}

	// Log payload as JSON for debugging
	if payloadJSON, err := json.Marshal(payload); err == nil {
		logger.Info("GenPDF payload: %s", string(payloadJSON))
	}

	// Connect to MongoDB
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		logger.Error("MongoDB connection failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "MongoDB connection failed",
		})
	}

	// Get MongoDB database name from config
	svcConfig := serviceConfig.NewServiceConfig()
	mongoDBName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(mongoDBName).Collection(payload.Collection)

	// Query document from MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"shopid": payload.ShopID, "docno": payload.DocNo}
	logger.Info("Finding document in MongoDB collection '%s' with filter: %+v", payload.Collection, filter)

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		logger.Error("MongoDB find error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Database query error",
		})
	}
	defer cur.Close(ctx)

	var document map[string]interface{}
	if cur.Next(ctx) {
		if err := cur.Decode(&document); err != nil {
			logger.Error("Document decode error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"code":    500,
				"message": "Error decoding document",
			})
		}
	} else {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Document not found",
		})
	}

	// Generate PDF using the gen-trans-pdf package
	pdf, err := gentranspdf.GenerateDocumentPDF(document, payload)
	if err != nil {
		logger.Error("PDF generation error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to generate PDF",
		})
	}

	// Save PDF to temp directory with GUID filename
	fileGUID := uuid.New().String()
	filePath, err := gentranspdf.SavePDF(pdf, fileGUID)
	if err != nil {
		logger.Error("PDF save error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to save PDF",
		})
	}

	logger.Success("PDF created successfully: %s", filePath)

	// Upload PDF to R2 and save history (non-blocking - errors won't affect response)
	go func() {
		uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer uploadCancel()

		if err := uploadPDFToR2AndSaveHistory(uploadCtx, filePath, payload, document, mongoDBName); err != nil {
			logger.Error("Failed to upload PDF to R2: %v", err)
		}
	}()

	// Return file as download
	return c.File(filePath)
}

// uploadPDFToR2AndSaveHistory - Upload PDF ไป R2 และเก็บประวัติใน MongoDB
func uploadPDFToR2AndSaveHistory(ctx context.Context, filePath string, payload gentranspdf.GenPDFPayload, document map[string]interface{}, mongoDBName string) error {
	// Get R2 client
	r2Client, err := GetR2Client()
	if err != nil || r2Client == nil {
		return fmt.Errorf("R2 client not available: %v", err)
	}

	// Read PDF file
	pdfData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read PDF file: %v", err)
	}

	// Generate R2 key: shopid/pdf/collection_docno_timestamp.pdf
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s_%s.pdf", payload.Collection, payload.DocNo, timestamp)
	r2Key := fmt.Sprintf("%s/pdf/%s", payload.ShopID, fileName)

	// Get bucket name from env
	r2BucketName := strings.TrimSpace(os.Getenv("R2_BUCKET_NAME"))
	if r2BucketName == "" {
		return fmt.Errorf("R2_BUCKET_NAME not configured")
	}

	// Upload to R2
	_, err = r2Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r2BucketName),
		Key:         aws.String(r2Key),
		Body:        bytes.NewReader(pdfData),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to R2: %v", err)
	}

	logger.Success("PDF uploaded to R2: %s", r2Key)

	// Save history to MongoDB Atlas
	if atlasClient == nil {
		logger.Warn("MongoDB Atlas not connected, skipping history save")
		return nil
	}

	// Extract document info
	docDate := extractDocDate(document)
	customerName := extractCustomerName(document)
	vendorName := extractVendorName(document)
	totalAmount := extractTotalAmount(document)

	now := time.Now()
	historyDoc := models.PdfHistory{
		ShopID:       payload.ShopID,
		Collection:   payload.Collection,
		DocNo:        payload.DocNo,
		DocDate:      docDate,
		Title:        payload.Title,
		FileName:     fileName,
		R2Key:        r2Key,
		FileSize:     int64(len(pdfData)),
		PageSize:     payload.PageSize,
		Orientation:  payload.Orientation,
		Language:     payload.Language,
		ThemeName:    payload.ThemeName,
		TemplateID:   payload.TemplateID,
		PrintedBy:    payload.PrintedBy,
		PrintedAt:    now,
		ReprintCount: 0,
		CreatedAt:    now,
		CustomerName: customerName,
		VendorName:   vendorName,
		TotalAmount:  totalAmount,
	}

	historyCollection := atlasDB.Collection("pdfHistory")
	_, err = historyCollection.InsertOne(ctx, historyDoc)
	if err != nil {
		logger.Error("Failed to save PDF history: %v", err)
		return err
	}

	logger.Success("PDF history saved: %s/%s", payload.Collection, payload.DocNo)

	// Cleanup temp file
	os.Remove(filePath)

	return nil
}

// extractDocDate - ดึงวันที่เอกสารจาก document
func extractDocDate(doc map[string]interface{}) time.Time {
	// ลอง docdatetime ก่อน
	if dt, ok := doc["docdatetime"]; ok {
		switch v := dt.(type) {
		case time.Time:
			return v
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t
			}
		}
	}
	// ลอง docdate
	if dt, ok := doc["docdate"]; ok {
		switch v := dt.(type) {
		case time.Time:
			return v
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// extractCustomerName - ดึงชื่อลูกค้า/ลูกหนี้จาก document
func extractCustomerName(doc map[string]interface{}) string {
	// ลอง custnames (array of {name: "..."})
	if custNames, ok := doc["custnames"].([]interface{}); ok && len(custNames) > 0 {
		if first, ok := custNames[0].(map[string]interface{}); ok {
			if name, ok := first["name"].(string); ok {
				return name
			}
		}
	}
	// ลอง customername
	if name, ok := doc["customername"].(string); ok {
		return name
	}
	// ลอง arname
	if name, ok := doc["arname"].(string); ok {
		return name
	}
	return ""
}

// extractVendorName - ดึงชื่อเจ้าหนี้/ผู้ขายจาก document
func extractVendorName(doc map[string]interface{}) string {
	// ลอง vendornames (array of {name: "..."})
	if vendorNames, ok := doc["vendornames"].([]interface{}); ok && len(vendorNames) > 0 {
		if first, ok := vendorNames[0].(map[string]interface{}); ok {
			if name, ok := first["name"].(string); ok {
				return name
			}
		}
	}
	// ลอง vendorname
	if name, ok := doc["vendorname"].(string); ok {
		return name
	}
	// ลอง apname
	if name, ok := doc["apname"].(string); ok {
		return name
	}
	return ""
}

// extractTotalAmount - ดึงยอดรวมจาก document
func extractTotalAmount(doc map[string]interface{}) float64 {
	// ลอง totalaftervat ก่อน (ยอดรวมสุทธิ)
	if amt, ok := doc["totalaftervat"].(float64); ok {
		return amt
	}
	// ลอง totalamount
	if amt, ok := doc["totalamount"].(float64); ok {
		return amt
	}
	// ลอง grandtotal
	if amt, ok := doc["grandtotal"].(float64); ok {
		return amt
	}
	return 0
}

// PdfHistorySimpleItem - รายการประวัติ PDF แบบครบถ้วน
type PdfHistorySimpleItem struct {
	// ข้อมูลหลัก
	ID         string `json:"id"`
	Collection string `json:"collection"` // collection ที่ดึงข้อมูล
	DocNo      string `json:"docno"`      // เลขที่เอกสาร
	DocDate    string `json:"docdate"`    // วันที่เอกสาร (ISO format)
	Title      string `json:"title"`      // ชื่อเอกสาร

	// ข้อมูลคู่ค้า
	CustomerName string `json:"customername,omitempty"` // ชื่อลูกหนี้/ลูกค้า (AR)
	VendorName   string `json:"vendorname,omitempty"`   // ชื่อเจ้าหนี้/ผู้ขาย (AP)

	// ข้อมูลยอดเงิน
	TotalAmount     float64 `json:"totalamount"`     // ยอดรวม (ตัวเลข)
	TotalAmountText string  `json:"totalamountText"` // ยอดรวม (format แล้ว)

	// ข้อมูล PDF
	Theme        string `json:"theme"`            // theme ที่ใช้
	Template     string `json:"template"`         // template ที่ใช้
	PageSize     string `json:"pageSize"`         // A4, A3, etc.
	Orientation  string `json:"orientation"`      // P=Portrait, L=Landscape
	Language     string `json:"language"`         // th, en, etc.
	FileName     string `json:"filename"`         // ชื่อไฟล์ PDF
	FileSize     int64  `json:"filesize"`         // ขนาดไฟล์ (bytes)
	FileSizeText string `json:"filesizeText"`     // ขนาดไฟล์ (format แล้ว)
	PdfURL       string `json:"pdfurl,omitempty"` // Presigned URL

	// ข้อมูลการพิมพ์
	PrintedAt    string `json:"printedAt"`    // วันเวลาที่พิมพ์ (ISO format)
	PrintedBy    string `json:"printedBy"`    // ผู้พิมพ์
	ReprintCount int    `json:"reprintCount"` // จำนวนครั้งที่ reprint
}

// formatFileSize - แปลงขนาดไฟล์เป็นรูปแบบที่อ่านง่าย
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatAmount - แปลงยอดเงินเป็นรูปแบบที่อ่านง่าย
func formatAmount(amount float64) string {
	return gentranspdf.FormatNumber(amount, 2)
}

// buildPdfHistoryItem - สร้าง PdfHistorySimpleItem จาก PdfHistory
func buildPdfHistoryItem(h models.PdfHistory, r2Client *s3.Client) PdfHistorySimpleItem {
	// Set defaults
	theme := h.ThemeName
	if theme == "" {
		theme = "default"
	}
	template := h.TemplateID
	if template == "" {
		template = "standard"
	}
	title := h.Title
	if title == "" {
		title = "เอกสาร"
	}
	language := h.Language
	if language == "" {
		language = "th"
	}

	// Format dates
	docDateStr := ""
	if !h.DocDate.IsZero() {
		docDateStr = h.DocDate.Format(time.RFC3339)
	}
	printedAtStr := h.PrintedAt.Format(time.RFC3339)

	item := PdfHistorySimpleItem{
		// ข้อมูลหลัก
		ID:         h.ID.Hex(),
		Collection: h.Collection,
		DocNo:      h.DocNo,
		DocDate:    docDateStr,
		Title:      title,

		// ข้อมูลคู่ค้า
		CustomerName: h.CustomerName,
		VendorName:   h.VendorName,

		// ข้อมูลยอดเงิน
		TotalAmount:     h.TotalAmount,
		TotalAmountText: formatAmount(h.TotalAmount),

		// ข้อมูล PDF
		Theme:        theme,
		Template:     template,
		PageSize:     h.PageSize,
		Orientation:  h.Orientation,
		Language:     language,
		FileName:     h.FileName,
		FileSize:     h.FileSize,
		FileSizeText: formatFileSize(h.FileSize),

		// ข้อมูลการพิมพ์
		PrintedAt:    printedAtStr,
		PrintedBy:    h.PrintedBy,
		ReprintCount: h.ReprintCount,
	}

	// Generate presigned URL (valid for 60 minutes)
	if r2Client != nil && h.R2Key != "" {
		url, err := getPresignedURL(r2Client, h.R2Key, 60)
		if err == nil {
			item.PdfURL = url
		}
	}

	return item
}

// PdfHistoryGetHandler - ดึงรายการประวัติ PDF แบบ GET (query params)
// GET /genpdf/history?shopid=xxx&collection=xxx&docno=xxx
func PdfHistoryGetHandler(c echo.Context) error {
	logger.Info("-> PdfHistoryGetHandler called")

	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB Atlas is not connected",
		})
	}

	// Get query parameters
	shopID := c.QueryParam("shopid")
	collectionName := c.QueryParam("collection")
	docNo := c.QueryParam("docno")

	if shopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "shopid query parameter is required",
		})
	}

	// Build filter
	filter := bson.M{"shopid": shopID}
	if collectionName != "" {
		filter["collection"] = collectionName
	}
	if docNo != "" {
		filter["docno"] = docNo
	}

	// Query options - sort by printedat descending, limit 100
	opts := options.Find().SetSort(bson.D{{Key: "printedat", Value: -1}}).SetLimit(100)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := atlasDB.Collection("pdfHistory")

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("Failed to query PDF history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to query PDF history",
		})
	}
	defer cursor.Close(ctx)

	var histories []models.PdfHistory
	if err := cursor.All(ctx, &histories); err != nil {
		logger.Error("Failed to decode PDF history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to decode PDF history",
		})
	}

	// Get R2 client for presigned URLs
	r2Client, _ := GetR2Client()

	// Build response array with all fields
	var result []PdfHistorySimpleItem
	for _, h := range histories {
		item := buildPdfHistoryItem(h, r2Client)
		result = append(result, item)
	}

	// Return empty array if no results
	if result == nil {
		result = []PdfHistorySimpleItem{}
	}

	return c.JSON(http.StatusOK, result)
}

// PdfHistoryListHandler - ดึงรายการประวัติ PDF
// POST /genpdf/history
func PdfHistoryListHandler(c echo.Context) error {
	logger.Info("-> PdfHistoryListHandler called")

	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB Atlas is not connected",
		})
	}

	var req models.PdfHistoryListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "shopid is required",
		})
	}

	// Build filter
	filter := bson.M{"shopid": req.ShopID}
	if req.Collection != "" {
		filter["collection"] = req.Collection
	}
	if req.DocNo != "" {
		filter["docno"] = req.DocNo
	}

	// Query options
	opts := options.Find().SetSort(bson.D{{Key: "printedat", Value: -1}})
	if req.Limit > 0 {
		opts.SetLimit(req.Limit)
	} else {
		opts.SetLimit(100) // default limit
	}
	if req.Skip > 0 {
		opts.SetSkip(req.Skip)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := atlasDB.Collection("pdfHistory")

	// Count total
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("Failed to count PDF history: %v", err)
		total = 0
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("Failed to query PDF history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to query PDF history",
		})
	}
	defer cursor.Close(ctx)

	var histories []models.PdfHistory
	if err := cursor.All(ctx, &histories); err != nil {
		logger.Error("Failed to decode PDF history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to decode PDF history",
		})
	}

	// Build response with presigned URLs
	r2Client, _ := GetR2Client()
	var dataItems []models.PdfHistoryListItem
	for _, h := range histories {
		item := models.PdfHistoryListItem{
			ID:           h.ID,
			ShopID:       h.ShopID,
			Collection:   h.Collection,
			DocNo:        h.DocNo,
			Title:        h.Title,
			FileName:     h.FileName,
			FileSize:     h.FileSize,
			PageSize:     h.PageSize,
			Orientation:  h.Orientation,
			Language:     h.Language,
			ThemeName:    h.ThemeName,
			TemplateID:   h.TemplateID,
			PrintedBy:    h.PrintedBy,
			PrintedAt:    h.PrintedAt,
			ReprintCount: h.ReprintCount,
			CreatedAt:    h.CreatedAt,
		}
		// Generate presigned URL (valid for 60 minutes)
		if r2Client != nil && h.R2Key != "" {
			url, err := getPresignedURL(r2Client, h.R2Key, 60)
			if err == nil {
				item.URL = url
			}
		}
		dataItems = append(dataItems, item)
	}

	return c.JSON(http.StatusOK, models.PdfHistoryListResponse{
		Status: "success",
		Code:   200,
		Count:  len(dataItems),
		Total:  total,
		Data:   dataItems,
	})
}

// PdfReprintHandler - พิมพ์ซ้ำ PDF จาก history ID
// GET /genpdf/reprint/:id
func PdfReprintHandler(c echo.Context) error {
	logger.Info("-> PdfReprintHandler called")

	historyID := c.Param("id")
	if historyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "id is required",
		})
	}

	// Parse ObjectID
	oid, err := primitive.ObjectIDFromHex(historyID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid id format",
		})
	}

	// Check services
	r2Client, err := GetR2Client()
	if err != nil || r2Client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "R2 storage is not configured",
		})
	}

	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB Atlas is not connected",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find history record
	collection := atlasDB.Collection("pdfHistory")
	var historyDoc models.PdfHistory
	err = collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&historyDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "PDF history not found",
		})
	}

	// Get bucket name
	r2BucketName := strings.TrimSpace(os.Getenv("R2_BUCKET_NAME"))

	// Download PDF from R2
	output, err := r2Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r2BucketName),
		Key:    aws.String(historyDoc.R2Key),
	})
	if err != nil {
		logger.Error("Failed to get PDF from R2: %v", err)
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "PDF file not found in storage",
		})
	}
	defer output.Body.Close()

	// Update reprint count
	_, err = collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$inc": bson.M{"reprintcount": 1},
	})
	if err != nil {
		logger.Warn("Failed to update reprint count: %v", err)
	}

	logger.Success("PDF reprint: %s (count: %d)", historyDoc.FileName, historyDoc.ReprintCount+1)

	// Stream PDF to client
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", historyDoc.FileName))

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)
	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}
