package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

var (
	r2Client     *s3.Client
	r2BucketName string
	r2InitMu     sync.Mutex
)

// GetR2Client returns the configured S3-compatible storage client.
// Failed initialization is intentionally retryable so a configuration reload can
// recover without leaving the process permanently stuck on the first error.
func GetR2Client() (*s3.Client, error) {
	r2InitMu.Lock()
	defer r2InitMu.Unlock()

	if r2Client != nil {
		return r2Client, nil
	}

	client, bucketName, err := initR2()
	if err != nil {
		return nil, err
	}
	r2Client = client
	r2BucketName = bucketName
	return r2Client, nil
}

// firstNonEmpty returns the first trimmed non-empty string from the given values.
// Used to let initR2 prefer config-driven S3_* env (set by settings screen) over legacy R2_* env.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		s := strings.TrimSpace(v)
		if s != "" {
			return s
		}
	}
	return ""
}

// initR2 สร้าง S3 client สำหรับ Cloudflare R2 หรือ S3-compatible storage เช่น MinIO
// อ่านค่าจาก config ที่ settings screen ส่งมา (ผ่าน setup config/save → env) เป็นหลัก
// fallback ไป R2_* env ถ้า config ไม่มี
func initR2() (*s3.Client, string, error) {
	accountID := firstNonEmpty(os.Getenv("S3_ACCOUNT_ID"), os.Getenv("R2_ACCOUNT_ID"))
	accessKeyID := firstNonEmpty(os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("R2_ACCESS_KEY_ID"))
	secretAccessKey := firstNonEmpty(os.Getenv("S3_SECRET_ACCESS_KEY"), os.Getenv("R2_SECRET_ACCESS_KEY"))
	bucketName := firstNonEmpty(os.Getenv("S3_BUCKET_NAME"), os.Getenv("R2_BUCKET_NAME"))
	// Endpoint: อ่านจาก S3_ENDPOINT (ที่ settings screen ส่ง) ก่อน, fallback R2_ENDPOINT
	r2Endpoint := firstNonEmpty(os.Getenv("S3_ENDPOINT"), os.Getenv("R2_ENDPOINT"))
	usePathStyle := strings.EqualFold(strings.TrimSpace(firstNonEmpty(os.Getenv("S3_FORCE_PATH_STYLE"), os.Getenv("R2_FORCE_PATH_STYLE"))), "true") || r2Endpoint != ""

	if (accountID == "" && r2Endpoint == "") || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		missing := []string{}
		if accountID == "" && r2Endpoint == "" {
			missing = append(missing, "S3_ENDPOINT or R2_ACCOUNT_ID")
		}
		if accessKeyID == "" {
			missing = append(missing, "S3_ACCESS_KEY_ID or R2_ACCESS_KEY_ID")
		}
		if secretAccessKey == "" {
			missing = append(missing, "S3_SECRET_ACCESS_KEY or R2_SECRET_ACCESS_KEY")
		}
		if bucketName == "" {
			missing = append(missing, "S3_BUCKET_NAME or R2_BUCKET_NAME")
		}
		err := fmt.Errorf("missing object storage environment variables: %s", strings.Join(missing, ", "))
		logger.Error("❌ Object storage init error: %v", err)
		return nil, "", err
	}

	// Resolve S3 endpoint: explicit override (MinIO/on-prem) wins, otherwise Cloudflare R2.
	endpointURL := r2Endpoint
	if endpointURL == "" {
		endpointURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	}

	var cfg aws.Config
	if r2Endpoint != "" {
		// S3-compatible storage (MinIO, Wasabi, etc.) — use path-style addressing.
		region := firstNonEmpty(os.Getenv("S3_REGION"), "us-east-1")
		customCfg, customErr := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		)
		if customErr != nil {
			err := fmt.Errorf("unable to load object storage SDK config: %w", customErr)
			logger.Error("❌ Object storage config error: %v", err)
			return nil, "", err
		}
		cfg = customCfg
		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &endpointURL
			o.UsePathStyle = usePathStyle
			o.Credentials = credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
			o.Region = region
		})
		logger.Info("✅ Object storage client initialized successfully (bucket: %s, endpoint: %s, pathStyle: %v)", bucketName, endpointURL, usePathStyle)
		return client, bucketName, nil
	} else {
		// Cloudflare R2 — keep the original resolver-based config for backwards compatibility.
		r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
			}, nil
		})
		cloudflareCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithEndpointResolverWithOptions(r2Resolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
			config.WithRegion("auto"),
		)
		if err != nil {
			logger.Error("❌ R2 config error: %v", err)
			return nil, "", err
		}
		client := s3.NewFromConfig(cloudflareCfg)
		logger.Info("✅ R2 client initialized successfully (bucket: %s, endpoint: %s, pathStyle: %v)", bucketName, endpointURL, usePathStyle)
		return client, bucketName, nil
	}
}

// getPresignedURL - สร้าง URL สำหรับดูไฟล์
// Private mode (default): return proxy URL ผ่าน goapi (/s3/file/{key}) เพื่อไม่เปิด R2/S3 URL ตรง
// Direct presigned URL mode: ต้องเปิด STORAGE_ALLOW_PRESIGNED_URL=true เท่านั้น
func getPresignedURL(client *s3.Client, r2Key string, expireMinutes int) (string, error) {
	if !storageAllowsDirectPresignedURL() {
		return storageProxyURL(r2Key), nil
	}

	// Direct URL mode (explicit opt-in only)
	if expireMinutes <= 0 {
		expireMinutes = 60 // default 1 hour
	}

	presignClient := s3.NewPresignClient(client)

	presignResult, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(r2BucketName),
		Key:    aws.String(r2Key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expireMinutes) * time.Minute
	})

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

// InitR2Client - Called by main.go to trigger initialization
func InitR2Client() error {
	_, err := GetR2Client()
	return err
}

func storageConfigErrorMessage(err error) string {
	const message = "Object storage is not configured"
	if err == nil {
		return message
	}
	return fmt.Sprintf("%s: %v", message, err)
}

// ImageUploadHandler - อัปโหลดรูปภาพต้นฉบับ + thumbnail (WebP) ไปยัง S3/MinIO แล้วคืน object key/URI
// ไม่บันทึก metadata ลงฐานข้อมูล — ผู้เรียกเก็บ URI ที่ได้ไว้ในเอกสารของตัวเอง (เช่น field รูป + <field>thumb)
// POST /image/upload
func ImageUploadHandler(c echo.Context) error {
	// Get holdingcode from auth context and reject multipart tenant tampering.
	holdingCode, authStatus := storageAuthorizedHoldingCode(c, c.FormValue("holdingcode"))
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
		logger.Error("❌ ImageUploadHandler: R2 client not available: %v", err)
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": storageConfigErrorMessage(err),
		})
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "No file uploaded",
		})
	}

	// Validate file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024)
	if file.Size > maxSize {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": fmt.Sprintf("File too large. Maximum size is %d MB", maxSize/(1024*1024)),
		})
	}

	// Validate file type — only PNG and JPG may be stored.
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
	}
	if !allowedExts[ext] {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid file type. Allowed: jpg, jpeg, png",
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

	// Generate filename: holdingcode/category/timestamp_hash.ext
	// ถ้ามี category จะแยก folder เช่น holdingcode/slip_money_in/filename.png
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s%s", timestamp, hashStr, ext)
	rawCategory := strings.Trim(strings.ReplaceAll(c.FormValue("category"), "\\", "/"), "/")
	category := storageSanitizeCategory(rawCategory)
	if rawCategory != category {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "error", "code": 400, "message": "Invalid upload category",
		})
	}
	if category == "products" {
		businessCode, businessStatus := storageAuthorizedBusinessCode(c)
		if businessStatus != http.StatusOK {
			return c.JSON(businessStatus, map[string]interface{}{
				"status": "error", "code": businessStatus, "message": "an active company is required",
			})
		}
		category = fmt.Sprintf("companies/%s/products/images", storageBusinessPathSegment(businessCode))
	}
	var r2Key string
	if category != "" {
		r2Key = fmt.Sprintf("%s/%s/%s", holdingCode, category, fileName)
	} else {
		r2Key = fmt.Sprintf("%s/%s", holdingCode, fileName)
	}

	// Detect content type and enforce PNG/JPG-only storage (guards against a renamed
	// non-image or webp/gif/bmp file that slipped past the extension check).
	contentType := http.DetectContentType(buf.Bytes())
	if contentType != "image/jpeg" && contentType != "image/png" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid image content. Only PNG and JPG are allowed.",
		})
	}
	thumbnailData, err := createWebPThumbnail(buf.Bytes())
	if err != nil {
		logger.Warn("Image upload rejected while creating thumbnail: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid image dimensions or content.",
		})
	}
	thumbnailKey := storageThumbnailObjectKey(r2Key)

	// Upload to R2
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			"message": "Failed to upload image to storage",
		})
	}
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r2BucketName),
		Key:         aws.String(thumbnailKey),
		Body:        bytes.NewReader(thumbnailData),
		ContentType: aws.String("image/webp"),
	})
	if err != nil {
		logger.Error("Failed to upload WebP thumbnail to R2: %v", err)
		deleteImageStorageObjects(ctx, client, r2Key)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to upload image thumbnail to storage",
		})
	}

	// Get optional fields (category ถูกดึงไว้แล้วด้านบน)
	description := c.FormValue("description")
	uploadedBy := c.FormValue("uploadedby")
	tagsStr := c.FormValue("tags")
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	now := time.Now()
	imageDoc := models.ImageMetadata{
		ID:           r2Key,
		HoldingCode:  holdingCode,
		FileName:     fileName,
		OriginalName: file.Filename,
		ContentType:  contentType,
		Size:         file.Size,
		ObjectKey:    r2Key,
		URL:          storageProxyURL(r2Key),
		ThumbnailKey: thumbnailKey,
		ThumbnailURL: storageProxyURL(r2Key) + "?variant=" + storageImageThumbnailVariant,
		Category:     category,
		Description:  description,
		Tags:         tags,
		UploadedBy:   uploadedBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	logger.Success("Image uploaded successfully: %s (shop: %s, size: %d bytes)", fileName, holdingCode, file.Size)

	return c.JSON(http.StatusOK, models.ImageResponse{
		Status:  "success",
		Code:    200,
		Message: "Image uploaded successfully",
		Data:    &imageDoc,
	})
}

func deleteImageStorageObjects(ctx context.Context, client *s3.Client, objectKey string) {
	for _, key := range storageImageObjectKeys(objectKey) {
		if _, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(r2BucketName),
			Key:    aws.String(key),
		}); err != nil {
			logger.Error("Failed to delete image storage object '%s': %v", key, err)
		}
	}
}
