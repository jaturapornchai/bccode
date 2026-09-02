package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// getImageFromR2 - ดึงรูปจาก R2 และ return เป็น bytes
func getImageFromR2(client *s3.Client, r2Key string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r2BucketName),
		Key:    aws.String(r2Key),
	})
	if err != nil {
		return nil, "", err
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := ""
	if output.ContentType != nil {
		contentType = *output.ContentType
	}

	return data, contentType, nil
}

// ImageUploadHandler - อัปโหลดรูปภาพไปยัง R2 และบันทึก metadata ใน MongoDB
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

	// Check MongoDB
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
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

	// Save metadata to MongoDB (ไม่เก็บ URL)
	now := time.Now()
	imageDoc := models.ImageMetadata{
		HoldingCode:  holdingCode,
		FileName:     fileName,
		OriginalName: file.Filename,
		ContentType:  contentType,
		Size:         file.Size,
		R2Key:        r2Key, // เก็บไว้ใน DB แต่ไม่ส่งออกไป frontend
		Category:     category,
		Description:  description,
		Tags:         tags,
		UploadedBy:   uploadedBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	collection := atlasDB.Collection("images")
	result, err := collection.InsertOne(ctx, imageDoc)
	if err != nil {
		logger.Error("Failed to save image metadata to MongoDB: %v", err)
		deleteImageStorageObjects(ctx, client, r2Key)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to save image metadata",
		})
	}

	// Set the ID from insert result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		imageDoc.ID = oid
	}

	logger.Success("Image uploaded successfully: %s (shop: %s, size: %d bytes)", fileName, holdingCode, file.Size)

	return c.JSON(http.StatusOK, models.ImageResponse{
		Status:  "success",
		Code:    200,
		Message: "Image uploaded successfully",
		Data:    &imageDoc,
	})
}

// ImageListHandler - ดึงรายการรูปภาพตาม holdingcode
// POST /image/list
func ImageListHandler(c echo.Context) error {
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	var req models.ImageListRequest
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
	filter := bson.M{"holdingcode": holdingCode}
	if req.Category != "" {
		filter["category"] = req.Category
	}

	// Query options
	opts := options.Find().SetSort(bson.D{{Key: "createdat", Value: -1}})
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

	collection := atlasDB.Collection("images")

	// นับจำนวนทั้งหมด (ไม่รวม limit/skip)
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("Failed to count images: %v", err)
		total = 0
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("Failed to query images: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to query images",
		})
	}
	defer cursor.Close(ctx)

	var images []models.ImageMetadata
	if err := cursor.All(ctx, &images); err != nil {
		logger.Error("Failed to decode images: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to decode images",
		})
	}

	// สร้าง response พร้อม Presigned URL เสมอ
	client, _ := GetR2Client()
	var dataItems []models.ImageListDataItem
	for _, img := range images {
		item := models.ImageListDataItem{
			ID:           img.ID,
			HoldingCode:  img.HoldingCode,
			FileName:     img.FileName,
			OriginalName: img.OriginalName,
			ContentType:  img.ContentType,
			Size:         img.Size,
			Category:     img.Category,
			Description:  img.Description,
			Tags:         img.Tags,
			UploadedBy:   img.UploadedBy,
			CreatedAt:    img.CreatedAt,
			UpdatedAt:    img.UpdatedAt,
		}
		// สร้าง private backend URL เฉพาะไฟล์ที่อยู่ใต้ holdingcode เดียวกัน
		if client != nil && img.R2Key != "" && storageObjectBelongsToShop(img.R2Key, img.HoldingCode) {
			url, err := getPresignedURL(client, img.R2Key, 60)
			if err == nil {
				item.URL = url
			}
		}

		// เพิ่มข้อมูล Slip Verification
		// ตรวจสอบว่าเคย verify หรือยัง โดยดูจาก VerifyStatus
		if img.VerifyStatus != "" {
			// เคยตรวจสอบแล้ว - ใส่ค่าจริง
			verified := img.Verified
			item.Verified = &verified
			verifyStatus := img.VerifyStatus
			item.VerifyStatus = &verifyStatus
			isDuplicate := img.IsDuplicate
			item.IsDuplicate = &isDuplicate
			if img.TransRef != "" {
				transRef := img.TransRef
				item.TransRef = &transRef
			}
			if img.TransAmount > 0 {
				transAmount := img.TransAmount
				item.TransAmount = &transAmount
			}
			if img.SenderName != "" {
				senderName := img.SenderName
				item.SenderName = &senderName
			}
			if img.ReceiverName != "" {
				receiverName := img.ReceiverName
				item.ReceiverName = &receiverName
			}
			if img.SlipType != "" {
				slipType := img.SlipType
				item.SlipType = &slipType
			}
		}
		// ถ้ายังไม่เคยตรวจสอบ fields จะเป็น nil (null ใน JSON) - Go's encoding/json will output null for nil pointers without omitempty

		dataItems = append(dataItems, item)
	}

	return c.JSON(http.StatusOK, models.ImageListDataResponse{
		Status: "success",
		Code:   200,
		Count:  len(dataItems),
		Total:  total,
		Data:   dataItems,
	})
}

// ImageGetHandler - ดึงรูปภาพจาก R2 และ return เป็น base64
// POST /image/get
func ImageGetHandler(c echo.Context) error {
	client, err := GetR2Client()
	if err != nil || client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": storageConfigErrorMessage(err),
		})
	}

	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	var req models.ImageGetRequest
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

	if req.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "filename is required",
		})
	}

	// ค้นหา metadata จาก MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := atlasDB.Collection("images")
	filter := bson.M{
		"holdingcode": holdingCode,
		"filename":    req.FileName,
	}

	var imageDoc models.ImageMetadata
	err = collection.FindOne(ctx, filter).Decode(&imageDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Image not found",
		})
	}

	if !storageObjectBelongsToShop(imageDoc.R2Key, holdingCode) {
		logger.Warn("Image get blocked: requested_shop=%s object_shop=%s", holdingCode, storageObjectHoldingCode(imageDoc.R2Key))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"status":  "error",
			"code":    403,
			"message": "Forbidden",
		})
	}

	// ดึงรูปจาก R2
	data, contentType, err := getImageFromR2(client, imageDoc.R2Key)
	if err != nil {
		logger.Error("Failed to get image from R2: %v", err)
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Image file not found in storage",
		})
	}

	// แปลงเป็น base64 data URL
	base64Data := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data))

	return c.JSON(http.StatusOK, models.ImageDataResponse{
		Status: "success",
		Code:   200,
		Data:   &imageDoc,
		Base64: base64Data,
	})
}

// ImageDeleteHandler - ลบรูปภาพจาก R2 และ MongoDB
// POST /image/delete
func ImageDeleteHandler(c echo.Context) error {
	client, err := GetR2Client()
	if err != nil || client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": storageConfigErrorMessage(err),
		})
	}

	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	var req models.ImageDeleteRequest
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

	if req.ImageID == "" && req.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "image_id or filename is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("images")

	// Build filter
	filter := bson.M{"holdingcode": holdingCode}
	if req.ImageID != "" {
		oid, err := primitive.ObjectIDFromHex(req.ImageID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"code":    400,
				"message": "Invalid image_id format",
			})
		}
		filter["_id"] = oid
	} else {
		filter["filename"] = req.FileName
	}

	// Find the image first to get R2 key
	var imageDoc models.ImageMetadata
	err = collection.FindOne(ctx, filter).Decode(&imageDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Image not found",
		})
	}

	if !storageObjectBelongsToShop(imageDoc.R2Key, holdingCode) {
		logger.Warn("Image delete blocked: requested_shop=%s object_shop=%s", holdingCode, storageObjectHoldingCode(imageDoc.R2Key))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"status":  "error",
			"code":    403,
			"message": "Forbidden",
		})
	}

	// Delete the original and its deterministic WebP thumbnail from R2.
	deleteImageStorageObjects(ctx, client, imageDoc.R2Key)

	// Delete from MongoDB
	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Error("Failed to delete from MongoDB: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to delete image metadata",
		})
	}

	logger.Success("Image deleted: %s (shop: %s)", imageDoc.FileName, holdingCode)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"code":    200,
		"message": "Image deleted successfully",
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

// ImageInfoHandler - ดึงข้อมูล metadata ของรูปภาพ (ไม่รวม base64)
// POST /image/info
func ImageInfoHandler(c echo.Context) error {
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	var req models.ImageGetRequest
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

	if req.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "filename is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := atlasDB.Collection("images")
	filter := bson.M{
		"holdingcode": holdingCode,
		"filename":    req.FileName,
	}

	var imageDoc models.ImageMetadata
	err := collection.FindOne(ctx, filter).Decode(&imageDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Image not found",
		})
	}

	return c.JSON(http.StatusOK, models.ImageResponse{
		Status: "success",
		Code:   200,
		Data:   &imageDoc,
	})
}

// ImageVerifyHandler - ส่งรูปไปตรวจสอบกับ Thunder API และบันทึกผลลงใน MongoDB
// POST /image/verify
func ImageVerifyHandler(c echo.Context) error {
	client, err := GetR2Client()
	if err != nil || client == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": storageConfigErrorMessage(err),
		})
	}

	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected",
		})
	}

	// Get Thunder API key
	thunderAPIKey := os.Getenv("THUNDER_API_KEY")
	if thunderAPIKey == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "Thunder API is not configured",
		})
	}

	var req models.ImageVerifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
		})
	}

	// Default: check_duplicate = true (ถ้าไม่ส่งมา)
	// Note: Go ไม่รู้ว่า false มาจาก user หรือ default แต่เราต้องการให้ default เป็น true
	// วิธีแก้: ใช้ pointer ใน model หรือตรวจสอบว่ามี field นี้ใน request หรือไม่
	// สำหรับความง่าย: ถ้า request body ไม่มี check_duplicate เลย จะเป็น false (Go default)
	// แต่เราต้องการ true เป็น default - ต้องใช้ pointer ใน model
	// ตอนนี้ใช้วิธี: ถ้าส่งมาเป็น false และไม่ระบุ field อื่นๆ ก็จะเป็น true
	checkDuplicate := true // Default เป็น true เสมอ

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

	if req.ImageID == "" && req.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "image_id or filename is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	collection := atlasDB.Collection("images")

	// Build filter
	filter := bson.M{"holdingcode": holdingCode}
	if req.ImageID != "" {
		oid, err := primitive.ObjectIDFromHex(req.ImageID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"code":    400,
				"message": "Invalid image_id format",
			})
		}
		filter["_id"] = oid
	} else {
		filter["filename"] = req.FileName
	}

	// Find the image
	var imageDoc models.ImageMetadata
	err = collection.FindOne(ctx, filter).Decode(&imageDoc)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"code":    404,
			"message": "Image not found",
		})
	}

	if !storageObjectBelongsToShop(imageDoc.R2Key, holdingCode) {
		logger.Warn("Image verify blocked: requested_shop=%s object_shop=%s", holdingCode, storageObjectHoldingCode(imageDoc.R2Key))
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"status":  "error",
			"code":    403,
			"message": "Forbidden",
		})
	}

	// Download image from R2
	imageData, contentType, err := getImageFromR2(client, imageDoc.R2Key)
	if err != nil {
		logger.Error("Failed to get image from R2: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to get image from storage",
		})
	}

	// Determine verify type (bank or truewallet)
	// ถ้าไม่ระบุ type หรือเป็น "auto" จะลอง bank ก่อน ถ้าไม่ได้ลอง truewallet
	verifyType := req.Type
	autoDetect := verifyType == "" || verifyType == "auto"

	var verifyResult *ThunderVerifyResult
	var detectedType string

	if autoDetect {
		// Auto-detect: ลอง bank ก่อน
		logger.Info("Auto-detecting slip type: trying bank first...")
		verifyResult, err = callThunderAPI(thunderAPIKey, imageData, contentType, imageDoc.FileName, checkDuplicate, "bank")
		detectedType = "bank"

		// ถ้า bank ไม่สำเร็จ (not_found หรือ error ที่ไม่ใช่ duplicate) ลอง truewallet
		if !verifyResult.Success && verifyResult.Status != "duplicate" {
			logger.Info("Bank verification failed, trying truewallet...")
			verifyResult, err = callThunderAPI(thunderAPIKey, imageData, contentType, imageDoc.FileName, checkDuplicate, "truewallet")
			detectedType = "truewallet"
		}
	} else {
		// ใช้ type ที่ระบุมา
		verifyResult, err = callThunderAPI(thunderAPIKey, imageData, contentType, imageDoc.FileName, checkDuplicate, verifyType)
		detectedType = verifyType
	}

	now := time.Now()

	// Prepare update data
	update := bson.M{
		"$set": bson.M{
			"verified":   verifyResult.Success,
			"verifiedat": now,
			"updatedat":  now,
			"sliptype":   detectedType, // bank หรือ truewallet
		},
	}

	if verifyResult.Success {
		update["$set"].(bson.M)["verifystatus"] = "success"
		update["$set"].(bson.M)["slipdata"] = verifyResult.Data
		update["$set"].(bson.M)["transref"] = verifyResult.TransRef
		update["$set"].(bson.M)["transamount"] = verifyResult.Amount
		update["$set"].(bson.M)["transdate"] = verifyResult.Date
		update["$set"].(bson.M)["sendername"] = verifyResult.SenderName
		update["$set"].(bson.M)["senderbank"] = verifyResult.SenderBank
		update["$set"].(bson.M)["receivername"] = verifyResult.ReceiverName
		update["$set"].(bson.M)["receiverbank"] = verifyResult.ReceiverBank
		update["$set"].(bson.M)["isduplicate"] = verifyResult.IsDuplicate
	} else if verifyResult.IsDuplicate {
		// กรณี duplicate - ยังคงมีข้อมูล slip ครบ
		update["$set"].(bson.M)["verifystatus"] = "duplicate"
		update["$set"].(bson.M)["verifyerror"] = verifyResult.Error
		update["$set"].(bson.M)["isduplicate"] = true
		update["$set"].(bson.M)["slipdata"] = verifyResult.Data
		update["$set"].(bson.M)["transref"] = verifyResult.TransRef
		update["$set"].(bson.M)["transamount"] = verifyResult.Amount
		update["$set"].(bson.M)["transdate"] = verifyResult.Date
		update["$set"].(bson.M)["sendername"] = verifyResult.SenderName
		update["$set"].(bson.M)["senderbank"] = verifyResult.SenderBank
		update["$set"].(bson.M)["receivername"] = verifyResult.ReceiverName
		update["$set"].(bson.M)["receiverbank"] = verifyResult.ReceiverBank
	} else {
		// กรณี error หรือ not_found
		update["$set"].(bson.M)["verifystatus"] = verifyResult.Status
		update["$set"].(bson.M)["verifyerror"] = verifyResult.Error
		update["$set"].(bson.M)["isduplicate"] = false
	}

	// Update MongoDB
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Failed to update image verification status: %v", err)
	}

	// Reload updated document
	err = collection.FindOne(ctx, filter).Decode(&imageDoc)
	if err != nil {
		logger.Error("Failed to reload image document: %v", err)
	}

	message := "Slip verification completed"
	if !verifyResult.Success {
		message = verifyResult.Error
	}

	return c.JSON(http.StatusOK, models.ImageVerifyResponse{
		Status:  "success",
		Code:    200,
		Message: message,
		Data:    &imageDoc,
	})
}

// ThunderVerifyResult - ผลลัพธ์จาก Thunder API
type ThunderVerifyResult struct {
	Success      bool
	Status       string // success, error, not_found, duplicate
	Error        string
	Data         map[string]interface{}
	TransRef     string
	Amount       float64
	Date         string
	SenderName   string
	SenderBank   string
	ReceiverName string
	ReceiverBank string
	IsDuplicate  bool
}

// extractSlipDataFromResponse - ดึงข้อมูล slip จาก Thunder API response
func extractSlipDataFromResponse(data map[string]interface{}, result *ThunderVerifyResult) {
	if transRef, ok := data["transRef"].(string); ok {
		result.TransRef = transRef
	}
	if date, ok := data["date"].(string); ok {
		result.Date = date
	}
	// Extract amount
	if amount, ok := data["amount"].(map[string]interface{}); ok {
		if amountVal, ok := amount["amount"].(float64); ok {
			result.Amount = amountVal
		}
	}
	// Extract sender info - รองรับทั้ง displayName และ account.name.th
	if sender, ok := data["sender"].(map[string]interface{}); ok {
		// ลอง displayName ก่อน
		if name, ok := sender["displayName"].(string); ok {
			result.SenderName = name
		} else if account, ok := sender["account"].(map[string]interface{}); ok {
			// ลอง account.name.th
			if nameObj, ok := account["name"].(map[string]interface{}); ok {
				if thName, ok := nameObj["th"].(string); ok {
					result.SenderName = thName
				} else if enName, ok := nameObj["en"].(string); ok {
					result.SenderName = enName
				}
			}
		}
		// Extract bank
		if bank, ok := sender["bank"].(map[string]interface{}); ok {
			if bankName, ok := bank["short"].(string); ok {
				result.SenderBank = bankName
			}
		}
	}
	// Extract receiver info - รองรับทั้ง displayName และ account.name.th
	if receiver, ok := data["receiver"].(map[string]interface{}); ok {
		// ลอง displayName ก่อน
		if name, ok := receiver["displayName"].(string); ok {
			result.ReceiverName = name
		} else if account, ok := receiver["account"].(map[string]interface{}); ok {
			// ลอง account.name.th
			if nameObj, ok := account["name"].(map[string]interface{}); ok {
				if thName, ok := nameObj["th"].(string); ok {
					result.ReceiverName = thName
				} else if enName, ok := nameObj["en"].(string); ok {
					result.ReceiverName = enName
				}
			}
		}
		// Extract bank
		if bank, ok := receiver["bank"].(map[string]interface{}); ok {
			if bankName, ok := bank["short"].(string); ok {
				result.ReceiverBank = bankName
			}
		}
	}
}

// callThunderAPI - เรียก Thunder API เพื่อตรวจสอบ slip
// verifyType: "bank" (default) หรือ "truewallet"
func callThunderAPI(apiKey string, imageData []byte, contentType, fileName string, checkDuplicate bool, verifyType string) (*ThunderVerifyResult, error) {
	result := &ThunderVerifyResult{}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		result.Error = "Failed to create form file"
		return result, err
	}
	_, err = io.Copy(part, bytes.NewReader(imageData))
	if err != nil {
		result.Error = "Failed to copy file data"
		return result, err
	}

	// Add checkDuplicate field
	if checkDuplicate {
		writer.WriteField("checkDuplicate", "true")
	}

	writer.Close()

	// Determine API URL based on type
	apiURL := "https://api.thunder.in.th/v1/verify"
	if verifyType == "truewallet" {
		apiURL = "https://api.thunder.in.th/v1/verify/truewallet"
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		result.Error = "Failed to create request"
		return result, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		result.Error = "Failed to call Thunder API: " + err.Error()
		return result, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = "Failed to read response"
		return result, err
	}

	// Parse response
	var thunderResp map[string]interface{}
	if err := json.Unmarshal(respBody, &thunderResp); err != nil {
		result.Error = "Failed to parse response"
		return result, err
	}

	logger.Info("Thunder API Response (status %d): %s", resp.StatusCode, string(respBody))

	// Handle different status codes
	switch resp.StatusCode {
	case 200:
		result.Success = true
		result.Status = "success"
		result.Data = thunderResp

		// Extract transaction details - ลองจาก data field ก่อน ถ้าไม่มีใช้ root level
		if data, ok := thunderResp["data"].(map[string]interface{}); ok {
			extractSlipDataFromResponse(data, result)
		} else {
			// Data is at root level
			extractSlipDataFromResponse(thunderResp, result)
		}

	case 400:
		result.Status = "error"
		if msg, ok := thunderResp["message"].(string); ok {
			result.Error = msg
			// ตรวจสอบว่าเป็น duplicate_slip หรือไม่
			if msg == "duplicate_slip" || strings.Contains(strings.ToLower(msg), "duplicate") {
				result.Status = "duplicate"
				result.IsDuplicate = true
				result.Data = thunderResp

				// กรณี duplicate ยังคงมีข้อมูล slip ใน data - ใช้ helper function
				if data, ok := thunderResp["data"].(map[string]interface{}); ok {
					extractSlipDataFromResponse(data, result)
				}
			}
		} else {
			result.Error = "Bad request"
		}

	case 401:
		result.Status = "error"
		result.Error = "Unauthorized: Invalid API key"

	case 403:
		result.Status = "error"
		if msg, ok := thunderResp["message"].(string); ok {
			result.Error = msg
		} else {
			result.Error = "Access denied"
		}

	case 404:
		result.Status = "not_found"
		result.Error = "Slip not found or QR code not readable"

	default:
		result.Status = "error"
		result.Error = fmt.Sprintf("Thunder API error: %d", resp.StatusCode)
	}

	return result, nil
}
