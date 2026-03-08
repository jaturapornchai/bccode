package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

// S3FileProxyHandler - Proxy file downloads from SeaweedFS S3 through goapi
// GET /s3/file/*
// ใช้แทน presigned URL เมื่อ SeaweedFS ไม่ได้เปิด port ให้เข้าถึงจากภายนอก
// goapi ดาวน์โหลดไฟล์จาก SeaweedFS internal แล้ว stream ให้ client
func S3FileProxyHandler(c echo.Context) error {
	// Get object key from wildcard path parameter
	objectKey := c.Param("*")
	if objectKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "file key is required",
		})
	}

	// Sanitize: prevent path traversal
	if strings.Contains(objectKey, "..") {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid file key",
		})
	}

	client, err := GetR2Client()
	if err != nil || client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Storage service not available",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r2BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		logger.Error("S3 proxy: failed to get object '%s': %v", objectKey, err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "File not found",
		})
	}
	defer output.Body.Close()

	// Set content type
	contentType := "application/octet-stream"
	if output.ContentType != nil {
		contentType = *output.ContentType
	}

	// Set content length
	if output.ContentLength != nil {
		c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", *output.ContentLength))
	}

	// Cache headers (1 hour)
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")

	// Stream the file to client
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response(), output.Body)
	if err != nil {
		logger.Error("S3 proxy: failed to stream object '%s': %v", objectKey, err)
	}

	return nil
}
