package handlers

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

// S3FileProxyHandler - Proxy private file downloads from S3/R2 through goapi
// GET /s3/file/*
// goapi ดาวน์โหลดไฟล์จาก private storage แล้ว stream ให้ client
func S3FileProxyHandler(c echo.Context) error {
	// Get object key from wildcard path parameter
	objectKey := storageNormalizeObjectKey(c.Param("*"))
	if objectKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "file key is required",
		})
	}

	if storageObjectHoldingCode(objectKey) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid file key",
		})
	}

	holdingCode := storageContextHoldingCode(c)
	if holdingCode == "" {
		logger.Warn("S3 proxy blocked: missing shop in auth context")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "shop not selected",
		})
	}
	if !storageObjectBelongsToShop(objectKey, holdingCode) {
		logger.Warn("S3 proxy blocked: requested_shop=%s token_shop=%s", storageObjectHoldingCode(objectKey), holdingCode)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "forbidden",
		})
	}

	client, err := GetR2Client()
	if err != nil || client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Storage service not available",
		})
	}

	return streamStorageObject(c, client, objectKey, "")
}

func streamStorageObject(c echo.Context, client *s3.Client, objectKey string, downloadName string) error {
	objectKey = storageNormalizeObjectKey(objectKey)
	if storageObjectHoldingCode(objectKey) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid file key",
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

	c.Response().Header().Set("Cache-Control", "private, no-store")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	if downloadName != "" {
		c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": downloadName,
		}))
	}

	// Stream the file to client
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response(), output.Body)
	if err != nil {
		logger.Error("S3 proxy: failed to stream object '%s': %v", objectKey, err)
	}

	return nil
}
