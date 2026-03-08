package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FileUploadResponse - response structure for file upload
type FileUploadResponse struct {
	Success   bool   `json:"success"`
	FileName  string `json:"fileName"`
	FileURL   string `json:"fileUrl"`
	ObjectKey string `json:"objectKey,omitempty"` // S3 object key สำหรับ download ภายหลัง
	FileSize  int64  `json:"fileSize"`
	Message   string `json:"message"`
}

// FileUploadHandler - handles file upload to SeaweedFS S3
// POST /upload
// Form data: file (multipart/form-data), shopid (optional)
// Returns: JSON with file name and presigned URL
func FileUploadHandler(c echo.Context) error {
	logger.Info("File upload request received from %s", c.RealIP())

	client, err := GetR2Client()
	if err != nil {
		logger.Error("S3 client not available: %v", err)
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{
			Success: false,
			Message: "Storage service not available",
		})
	}

	// รับไฟล์จาก form
	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to get file from form: %v", err)
		return c.JSON(http.StatusBadRequest, FileUploadResponse{
			Success: false,
			Message: "No file uploaded or invalid form data",
		})
	}

	// ตรวจสอบขนาดไฟล์ (จำกัดไว้ที่ 50MB)
	maxSize := int64(50 * 1024 * 1024) // 50MB
	if file.Size > maxSize {
		logger.Warn("File size too large: %d bytes (max: %d bytes)", file.Size, maxSize)
		return c.JSON(http.StatusBadRequest, FileUploadResponse{
			Success: false,
			Message: fmt.Sprintf("File size too large. Maximum size is %d MB", maxSize/(1024*1024)),
		})
	}

	// เปิดไฟล์
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file: %v", err)
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{
			Success: false,
			Message: "Failed to process uploaded file",
		})
	}
	defer src.Close()

	// อ่านไฟล์เข้า memory
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		logger.Error("Failed to read uploaded file: %v", err)
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{
			Success: false,
			Message: "Failed to read uploaded file",
		})
	}

	// สร้างชื่อไฟล์ใหม่ด้วย GUID
	ext := filepath.Ext(file.Filename)
	newFileName := uuid.New().String() + ext

	// สร้าง S3 key: uploads/{date}/{filename}
	shopID := c.FormValue("shopid")
	var objectKey string
	if shopID != "" {
		objectKey = fmt.Sprintf("%s/uploads/%s/%s", shopID, time.Now().Format("20060102"), newFileName)
	} else {
		objectKey = fmt.Sprintf("uploads/%s/%s", time.Now().Format("20060102"), newFileName)
	}

	// Detect content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload ไป SeaweedFS S3
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r2BucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		logger.Error("Failed to upload to S3: %v", err)
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{
			Success: false,
			Message: "Failed to upload file to storage",
		})
	}

	// สร้าง presigned URL สำหรับ download
	fileURL, _ := getPresignedURL(client, objectKey, 60)
	if fileURL == "" {
		// fallback: ใช้ direct path
		fileURL = objectKey
	}

	logger.Success("File uploaded to S3: %s (original: %s, size: %d bytes)",
		newFileName, file.Filename, len(fileBytes))

	return c.JSON(http.StatusOK, FileUploadResponse{
		Success:   true,
		FileName:  newFileName,
		ObjectKey: objectKey,
		FileURL:   fileURL,
		FileSize:  int64(len(fileBytes)),
		Message:   "File uploaded successfully",
	})
}

// FileDownloadHandler - download file from SeaweedFS S3 via presigned URL
// GET /upload/download/:key
func FileDownloadHandler(c echo.Context) error {
	objectKey := c.Param("key")
	if objectKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "file key is required",
		})
	}

	// URL decode (key อาจมี / encoded)
	objectKey = strings.ReplaceAll(objectKey, "%2F", "/")

	client, err := GetR2Client()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Storage service not available",
		})
	}

	url, err := getPresignedURL(client, objectKey, 5)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to generate download URL",
		})
	}

	return c.Redirect(http.StatusTemporaryRedirect, url)
}
