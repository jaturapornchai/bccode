package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AttachmentUploadHandler - อัปโหลดไฟล์แนบเอกสารไปยัง R2 และบันทึก metadata ใน MongoDB
// POST /api/attachment/upload
// Form data: file, holding_code, screen_type, docno, guidfixed, description, uploaded_by, uploaded_name
func AttachmentUploadHandler(c echo.Context) error {
	holdingCode, authStatus := storageAuthorizedHoldingCode(c, c.FormValue("holding_code"))
	if authStatus != http.StatusOK {
		message := "shop not selected"
		if authStatus == http.StatusForbidden {
			message = "Forbidden"
		}
		return c.JSON(authStatus, map[string]interface{}{
			"status":  "error",
			"code":    authStatus,
			"message": message,
		})
	}

	// Check R2 client
	client, err := GetR2Client()
	if err != nil || client == nil {
		logger.Error("❌ AttachmentUploadHandler: R2 client not available: %v", err)
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "R2 storage is not configured",
		})
	}

	// Check MongoDB
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	// Get required fields from form
	screenType := c.FormValue("screen_type")
	docNo := c.FormValue("docno")
	guidFixed := c.FormValue("guid_fixed")
	uploadedBy := c.FormValue("uploaded_by")
	uploadedName := c.FormValue("uploaded_name")

	if holdingCode == "" || screenType == "" || docNo == "" || guidFixed == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "holding_code, screen_type, docno, and guidfixed are required",
		})
	}

	if uploadedBy == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "uploaded_by is required",
		})
	}

	// Get optional description
	description := c.FormValue("description")

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "No file uploaded",
		})
	}

	// Validate file size (max 20MB สำหรับเอกสาร)
	maxSize := int64(20 * 1024 * 1024)
	if file.Size > maxSize {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": fmt.Sprintf("File too large. Maximum size is %d MB", maxSize/(1024*1024)),
		})
	}

	// Validate file type (documents + images)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]string{
		".pdf":  "pdf",
		".xlsx": "xlsx",
		".xls":  "xls",
		".jpg":  "jpg",
		".jpeg": "jpg",
		".png":  "png",
	}
	fileType, ok := allowedExts[ext]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid file type. Allowed: PDF, Excel (xlsx, xls), Images (jpg, jpeg, png)",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to process uploaded file",
		})
	}
	defer src.Close()

	// Read file content
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		logger.Error("Failed to read file content: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to read file",
		})
	}

	// Calculate hash for filename
	hash := sha256.Sum256(buf.Bytes())
	hashStr := hex.EncodeToString(hash[:])[:16] // ใช้ 16 ตัวแรก

	// Generate filename: holding_code/attachments/screentype/guidfixed/timestamp_hash.ext
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, hashStr, ext)
	r2Key := fmt.Sprintf("%s/attachments/%s/%s/%s", holdingCode, screenType, guidFixed, fileName)

	// Detect content type
	contentType := http.DetectContentType(buf.Bytes())

	// Override content type สำหรับ file types ที่รู้จัก
	switch ext {
	case ".pdf":
		contentType = "application/pdf"
	case ".xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		contentType = "application/vnd.ms-excel"
	}

	// Upload to R2
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r2BucketName),
		Key:         aws.String(r2Key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		logger.Error("Failed to upload to R2: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to upload file to storage",
		})
	}

	// Save metadata to MongoDB
	now := time.Now()
	attachmentDoc := models.AttachmentMetadata{
		HoldingCode:  holdingCode,
		ScreenType:   screenType,
		DocNo:        docNo,
		GuidFixed:    guidFixed,
		FileName:     fileName,
		OriginalName: file.Filename,
		ContentType:  contentType,
		FileType:     fileType,
		Size:         file.Size,
		R2Key:        r2Key,
		Description:  description,
		UploadedBy:   uploadedBy,
		UploadedName: uploadedName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	collection := atlasDB.Collection("attachments")
	result, err := collection.InsertOne(ctx, attachmentDoc)
	if err != nil {
		logger.Error("Failed to save attachment metadata to MongoDB: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to save attachment metadata",
		})
	}

	// Set the ID from insert result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		attachmentDoc.ID = oid
	}

	logger.Success("Attachment uploaded successfully: %s (shop: %s, doc: %s, size: %d bytes)",
		fileName, holdingCode, docNo, file.Size)

	return c.JSON(http.StatusOK, models.AttachmentResponse{
		Status:  "success",
		Code:    200,
		Message: "Attachment uploaded successfully",
		Data:    &attachmentDoc,
	})
}

// AttachmentListHandler - ดึงรายการไฟล์แนบตามเงื่อนไข
// POST /api/attachment/list
// Body: { holding_code, screen_type, docno, guidfixed, limit, skip }
func AttachmentListHandler(c echo.Context) error {
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	var req models.AttachmentListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
		})
	}

	holdingCode, authStatus := storageAuthorizedHoldingCode(c, req.HoldingCode)
	if authStatus != http.StatusOK {
		message := "shop not selected"
		if authStatus == http.StatusForbidden {
			message = "Forbidden"
		}
		return c.JSON(authStatus, map[string]interface{}{
			"status":  "error",
			"code":    authStatus,
			"message": message,
		})
	}

	// Build filter
	filter := bson.M{"holding_code": holdingCode}
	if req.ScreenType != "" {
		filter["screen_type"] = req.ScreenType
	}
	if req.DocNo != "" {
		filter["docno"] = req.DocNo
	}
	if req.GuidFixed != "" {
		filter["guid_fixed"] = req.GuidFixed
	}

	// Query options
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
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

	collection := atlasDB.Collection("attachments")

	// นับจำนวนทั้งหมด (ไม่รวม limit/skip)
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("Failed to count attachments: %v", err)
		total = 0
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("Failed to query attachments: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to query attachments",
		})
	}
	defer cursor.Close(ctx)

	var attachments []models.AttachmentMetadata
	if err := cursor.All(ctx, &attachments); err != nil {
		logger.Error("Failed to decode attachments: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to decode attachments",
		})
	}

	// สร้าง response พร้อม Presigned URL
	client, _ := GetR2Client()
	var dataItems []models.AttachmentListDataItem
	for _, att := range attachments {
		item := models.AttachmentListDataItem{
			ID:           att.ID,
			HoldingCode:  att.HoldingCode,
			ScreenType:   att.ScreenType,
			DocNo:        att.DocNo,
			GuidFixed:    att.GuidFixed,
			FileName:     att.FileName,
			OriginalName: att.OriginalName,
			ContentType:  att.ContentType,
			FileType:     att.FileType,
			Size:         att.Size,
			Description:  att.Description,
			UploadedBy:   att.UploadedBy,
			UploadedName: att.UploadedName,
			CreatedAt:    att.CreatedAt,
			UpdatedAt:    att.UpdatedAt,
		}

		// สร้าง private backend URL สำหรับ stream ผ่าน goapi เฉพาะไฟล์ใต้ holding_code เดียวกัน
		if client != nil && att.R2Key != "" && storageObjectBelongsToShop(att.R2Key, att.HoldingCode) {
			url, err := getPresignedURL(client, att.R2Key, 60)
			if err == nil {
				item.URL = url
			}
		}

		dataItems = append(dataItems, item)
	}

	return c.JSON(http.StatusOK, models.AttachmentListDataResponse{
		Status: "success",
		Code:   200,
		Count:  len(dataItems),
		Total:  total,
		Data:   dataItems,
	})
}

// AttachmentDeleteHandler - ลบไฟล์แนบจาก R2 และ MongoDB
// POST /api/attachment/delete
// Body: { holding_code, attachment_id or filename }
func AttachmentDeleteHandler(c echo.Context) error {
	client, err := GetR2Client()
	if err != nil || client == nil {
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
			"message": "MongoDB is not connected",
		})
	}

	var req models.AttachmentDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
		})
	}

	holdingCode, authStatus := storageAuthorizedHoldingCode(c, req.HoldingCode)
	if authStatus != http.StatusOK {
		message := "shop not selected"
		if authStatus == http.StatusForbidden {
			message = "Forbidden"
		}
		return c.JSON(authStatus, map[string]interface{}{
			"status":  "error",
			"code":    authStatus,
			"message": message,
		})
	}

	if req.AttachmentID == "" && req.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "attachment_id or filename is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("attachments")

	// Build filter
	filter := bson.M{"holding_code": holdingCode}
	if req.AttachmentID != "" {
		oid, err := primitive.ObjectIDFromHex(req.AttachmentID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"code":    400,
				"message": "Invalid attachment_id format",
			})
		}
		filter["_id"] = oid
	} else {
		filter["file_name"] = req.FileName
	}

	// Find the attachment first to get R2 key
	var attachmentDoc models.AttachmentMetadata
	err = collection.FindOne(ctx, filter).Decode(&attachmentDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Attachment not found",
		})
	}

	if !storageObjectBelongsToShop(attachmentDoc.R2Key, holdingCode) {
		logger.Warn("Attachment delete blocked: requested_shop=%s object_shop=%s", holdingCode, storageObjectHoldingCode(attachmentDoc.R2Key))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"status":  "error",
			"code":    403,
			"message": "Forbidden",
		})
	}

	// Delete from R2
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r2BucketName),
		Key:    aws.String(attachmentDoc.R2Key),
	})
	if err != nil {
		logger.Error("Failed to delete from R2: %v", err)
		// Continue to delete from MongoDB anyway
	}

	// Delete from MongoDB
	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Error("Failed to delete from MongoDB: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to delete attachment metadata",
		})
	}

	logger.Success("Attachment deleted: %s (shop: %s, doc: %s)", attachmentDoc.FileName, holdingCode, attachmentDoc.DocNo)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"code":    200,
		"message": "Attachment deleted successfully",
	})
}

// AttachmentDownloadHandler - ดาวน์โหลดไฟล์แนบโดยตรงผ่าน private backend stream
// GET /api/attachment/download/:id?holding_code=xxx
func AttachmentDownloadHandler(c echo.Context) error {
	attachmentID := c.Param("id")
	if attachmentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "attachment id is required",
		})
	}

	holdingCode, authStatus := storageAuthorizedHoldingCode(c, c.QueryParam("holding_code"))
	if authStatus != http.StatusOK {
		message := "shop not selected"
		if authStatus == http.StatusForbidden {
			message = "Forbidden"
		}
		return c.JSON(authStatus, map[string]interface{}{
			"status":  "error",
			"code":    authStatus,
			"message": message,
		})
	}

	client, err := GetR2Client()
	if err != nil || client == nil {
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
			"message": "MongoDB is not connected",
		})
	}

	oid, err := primitive.ObjectIDFromHex(attachmentID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid attachment id format",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("attachments")
	filter := bson.M{
		"_id":          oid,
		"holding_code": holdingCode,
	}

	var attachmentDoc models.AttachmentMetadata
	err = collection.FindOne(ctx, filter).Decode(&attachmentDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Attachment not found",
		})
	}

	if !storageObjectBelongsToShop(attachmentDoc.R2Key, holdingCode) {
		logger.Warn("Attachment download blocked: requested_shop=%s object_shop=%s", holdingCode, storageObjectHoldingCode(attachmentDoc.R2Key))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"status":  "error",
			"code":    403,
			"message": "Forbidden",
		})
	}

	return streamStorageObject(c, client, attachmentDoc.R2Key, attachmentDoc.OriginalName)
}
