package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ─── S3 client สำหรับ token storage ───
var (
	tokenS3Client *s3.Client
	tokenS3Bucket string
	tokenS3Once   sync.Once
	tokenS3Err    error
)

func getTokenS3Client() (*s3.Client, string, error) {
	tokenS3Once.Do(func() {
		endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
		accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
		secretKey := strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY"))
		tokenS3Bucket = strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))

		if endpoint == "" || accessKey == "" || secretKey == "" || tokenS3Bucket == "" {
			tokenS3Err = fmt.Errorf("S3 config not available for token storage")
			return
		}

		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
		})

		cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithEndpointResolverWithOptions(resolver),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
			awsconfig.WithRegion("us-east-1"),
		)
		if err != nil {
			tokenS3Err = err
			return
		}

		tokenS3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	})
	return tokenS3Client, tokenS3Bucket, tokenS3Err
}

// ==================== Token Export (JSON per holding_code) ====================

// TokenExportData โครงสร้างข้อมูล export สำหรับ MCP client config
type TokenExportData struct {
	APIKey              string                 `json:"api_key"`
	HoldingCode         string                 `json:"holding_code"`
	Name                string                 `json:"name"`
	ExpiresAt           *string                `json:"expires_at,omitempty"`
	CreatedAt           string                 `json:"created_at"`
	ClaudeDesktopConfig map[string]interface{} `json:"claude_desktop_config"`
	ClaudeCodeConfig    map[string]interface{} `json:"claude_code_config"`
}

// GenerateTokenExport สร้าง export data สำหรับ MCP client
// ใช้ stdio transport ผ่าน Node.js MCP server — เสถียรกว่า mcp-remote (SSE)
func GenerateTokenExport(apiKey, holdingCode, name, serverURL string, expiresAt *time.Time, createdAt time.Time) *TokenExportData {
	if serverURL == "" {
		serverURL = "http://localhost:8888"
	}

	var expiresStr *string
	if expiresAt != nil {
		s := expiresAt.Format(time.RFC3339)
		expiresStr = &s
	}

	// ทั้ง Claude Desktop และ Claude Code ใช้ stdio transport เหมือนกัน
	// ผ่าน Node.js MCP server ที่เชื่อมต่อ Go API ด้วย HTTP
	mcpServerConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"bc-erp": map[string]interface{}{
				"command": "node",
				"args":    []string{"/path/to/backend/mcp-server/index.js"},
				"env": map[string]interface{}{
					"BC_API_URL": serverURL,
					"BC_API_KEY": apiKey,
				},
			},
		},
	}

	return &TokenExportData{
		APIKey:              apiKey,
		HoldingCode:         holdingCode,
		Name:                name,
		ExpiresAt:           expiresStr,
		CreatedAt:           createdAt.Format(time.RFC3339),
		ClaudeDesktopConfig: mcpServerConfig,
		ClaudeCodeConfig:    mcpServerConfig,
	}
}

// SaveTokenFile บันทึก token config เป็น JSON file ไป S3
func SaveTokenFile(holdingCode, tokenID string, data *TokenExportData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("แปลง JSON ล้มเหลว: %w", err)
	}

	client, bucket, s3Err := getTokenS3Client()
	if s3Err != nil {
		// fallback: เก็บ local (backward compatible)
		logger.Warn("[Token Export] S3 not available, saving locally: %v", s3Err)
		return saveTokenFileLocal(holdingCode, tokenID, jsonData)
	}

	objectKey := fmt.Sprintf("mcp-tokens/%s/%s.json", holdingCode, tokenID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("upload token to S3 failed: %w", err)
	}

	logger.Info("[Token Export] บันทึกไป S3 สำเร็จ: %s", objectKey)
	return nil
}

// DeleteTokenFile ลบ token file จาก S3
func DeleteTokenFile(holdingCode, tokenID string) error {
	client, bucket, s3Err := getTokenS3Client()
	if s3Err != nil {
		// fallback: ลบ local
		return deleteTokenFileLocal(holdingCode, tokenID)
	}

	objectKey := fmt.Sprintf("mcp-tokens/%s/%s.json", holdingCode, tokenID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("delete token from S3 failed: %w", err)
	}

	logger.Info("[Token Export] ลบจาก S3 สำเร็จ: %s", objectKey)
	return nil
}

// ─── Local fallback (backward compatible) ───

func saveTokenFileLocal(holdingCode, tokenID string, jsonData []byte) error {
	dir := fmt.Sprintf("/app/mcp-tokens/%s", holdingCode)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("สร้าง directory ล้มเหลว: %w", err)
	}
	filePath := fmt.Sprintf("%s/%s.json", dir, tokenID)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("เขียนไฟล์ล้มเหลว: %w", err)
	}
	logger.Info("[Token Export] บันทึกไฟล์ local สำเร็จ: %s", filePath)
	return nil
}

func deleteTokenFileLocal(holdingCode, tokenID string) error {
	filePath := fmt.Sprintf("/app/mcp-tokens/%s/%s.json", holdingCode, tokenID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ลบไฟล์ล้มเหลว: %w", err)
	}
	return nil
}
