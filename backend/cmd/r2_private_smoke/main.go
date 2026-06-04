package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type r2SmokeConfig struct {
	Provider        string
	AccountID       string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	HoldingCode     string
}

type bootstrapConfig struct {
	Integrations map[string]string `json:"integrations"`
	Storage      map[string]string `json:"storage"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "r2 private smoke failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client, err := newR2Client(cfg)
	if err != nil {
		return err
	}

	suffix, err := randomHex(8)
	if err != nil {
		return err
	}
	objectKey := fmt.Sprintf("%s/__r2_private_smoke__/%s.txt", cfg.HoldingCode, suffix)
	body := []byte("bc account r2 private smoke " + time.Now().UTC().Format(time.RFC3339Nano))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("%s smoke: bucket=%s holding_code=%s key=%s\n", cfg.Provider, cfg.BucketName, cfg.HoldingCode, objectKey)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("text/plain; charset=utf-8"),
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	defer func() {
		deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer deleteCancel()
		_, _ = client.DeleteObject(deleteCtx, &s3.DeleteObjectInput{
			Bucket: aws.String(cfg.BucketName),
			Key:    aws.String(objectKey),
		})
	}()

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	defer output.Body.Close()

	got, err := io.ReadAll(output.Body)
	if err != nil {
		return fmt.Errorf("read object: %w", err)
	}
	if !bytes.Equal(got, body) {
		return fmt.Errorf("downloaded object body mismatch")
	}

	status, err := checkUnsignedS3URL(ctx, cfg, objectKey)
	if err != nil {
		return err
	}
	if status < 400 {
		return fmt.Errorf("unsigned S3 endpoint returned %d; expected private access to reject anonymous read", status)
	}

	fmt.Printf("%s smoke: upload/read/delete path works and unsigned object read is blocked\n", cfg.Provider)
	return nil
}

func loadConfig() (r2SmokeConfig, error) {
	bootstrap := loadBootstrap()
	cfg := r2SmokeConfig{
		AccountID:       firstNonEmpty(os.Getenv("R2_ACCOUNT_ID"), bootstrapValue(bootstrap, "r2_account_id")),
		Endpoint:        firstNonEmpty(os.Getenv("S3_ENDPOINT"), bootstrapValue(bootstrap, "s3_endpoint")),
		AccessKeyID:     firstNonEmpty(os.Getenv("R2_ACCESS_KEY_ID"), bootstrapValue(bootstrap, "r2_access_key_id")),
		SecretAccessKey: firstNonEmpty(os.Getenv("R2_SECRET_ACCESS_KEY"), bootstrapValue(bootstrap, "r2_secret_access_key")),
		BucketName:      firstNonEmpty(os.Getenv("R2_BUCKET_NAME"), bootstrapValue(bootstrap, "r2_bucket_name")),
		HoldingCode:     strings.Trim(strings.TrimSpace(os.Getenv("BC_R2_SMOKE_HOLDING_CODE")), "/"),
	}

	if cfg.AccountID != "" {
		cfg.Provider = "R2"
	} else if cfg.Endpoint != "" {
		cfg.Provider = "S3"
		cfg.AccessKeyID = firstNonEmpty(os.Getenv("S3_ACCESS_KEY_ID"), bootstrapValue(bootstrap, "s3_access_key_id"))
		cfg.SecretAccessKey = firstNonEmpty(os.Getenv("S3_SECRET_ACCESS_KEY"), bootstrapValue(bootstrap, "s3_secret_access_key"))
		cfg.BucketName = firstNonEmpty(os.Getenv("S3_BUCKET_NAME"), bootstrapValue(bootstrap, "s3_bucket_name"))
	} else {
		cfg.Provider = "R2"
	}

	var missing []string
	if cfg.Provider == "R2" && cfg.AccountID == "" {
		missing = append(missing, "R2_ACCOUNT_ID")
	}
	if cfg.Provider == "S3" && cfg.Endpoint == "" {
		missing = append(missing, "S3_ENDPOINT")
	}
	if cfg.AccessKeyID == "" {
		if cfg.Provider == "S3" {
			missing = append(missing, "S3_ACCESS_KEY_ID")
		} else {
			missing = append(missing, "R2_ACCESS_KEY_ID")
		}
	}
	if cfg.SecretAccessKey == "" {
		if cfg.Provider == "S3" {
			missing = append(missing, "S3_SECRET_ACCESS_KEY")
		} else {
			missing = append(missing, "R2_SECRET_ACCESS_KEY")
		}
	}
	if cfg.BucketName == "" {
		if cfg.Provider == "S3" {
			missing = append(missing, "S3_BUCKET_NAME")
		} else {
			missing = append(missing, "R2_BUCKET_NAME")
		}
	}
	if cfg.HoldingCode == "" {
		missing = append(missing, "BC_R2_SMOKE_HOLDING_CODE")
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}
	if strings.Contains(cfg.HoldingCode, "..") || strings.Contains(cfg.HoldingCode, "\\") || strings.Contains(cfg.HoldingCode, "/") {
		return cfg, fmt.Errorf("BC_R2_SMOKE_HOLDING_CODE must be a single holding_code segment")
	}
	return cfg, nil
}

func newR2Client(smokeCfg r2SmokeConfig) (*s3.Client, error) {
	endpoint := smokeCfg.Endpoint
	region := "us-east-1"
	usePathStyle := true
	if smokeCfg.Provider == "R2" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", smokeCfg.AccountID)
		region = "auto"
		usePathStyle = false
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(smokeCfg.AccessKeyID, smokeCfg.SecretAccessKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("load %s SDK config: %w", smokeCfg.Provider, err)
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	}), nil
}

func checkUnsignedS3URL(ctx context.Context, cfg r2SmokeConfig, objectKey string) (int, error) {
	publicURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(cfg.Endpoint, "/"), cfg.BucketName, objectKey)
	if cfg.Provider == "R2" {
		publicURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", cfg.AccountID, cfg.BucketName, objectKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, publicURL, nil)
	if err != nil {
		return 0, fmt.Errorf("build unsigned request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("unsigned S3 request: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func loadBootstrap() bootstrapConfig {
	for _, path := range []string{"bootstrap.json", filepath.Join("config", "bootstrap.json")} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg bootstrapConfig
		if err := json.Unmarshal(data, &cfg); err == nil {
			return cfg
		}
	}
	return bootstrapConfig{}
}

func bootstrapValue(cfg bootstrapConfig, key string) string {
	if value := strings.TrimSpace(cfg.Integrations[key]); value != "" {
		return value
	}
	return strings.TrimSpace(cfg.Storage[key])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func randomHex(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random suffix: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
