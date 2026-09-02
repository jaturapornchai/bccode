package microservice

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// PersisterR2 implements IPersisterFile using Cloudflare R2 or S3-compatible storage.
type PersisterR2 struct {
	client     *s3.Client
	bucketName string
	endpoint   string
	initMu     sync.Mutex
}

func NewPersisterR2() *PersisterR2 {
	return &PersisterR2{}
}

func firstNonEmptyStorageEnv(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (p *PersisterR2) init() error {
	p.initMu.Lock()
	defer p.initMu.Unlock()

	if p.client != nil {
		return nil
	}

	accountID := firstNonEmptyStorageEnv(os.Getenv("S3_ACCOUNT_ID"), os.Getenv("R2_ACCOUNT_ID"))
	endpoint := firstNonEmptyStorageEnv(os.Getenv("S3_ENDPOINT"), os.Getenv("R2_ENDPOINT"))
	accessKeyID := firstNonEmptyStorageEnv(os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("R2_ACCESS_KEY_ID"))
	secretAccessKey := firstNonEmptyStorageEnv(os.Getenv("S3_SECRET_ACCESS_KEY"), os.Getenv("R2_SECRET_ACCESS_KEY"))
	bucketName := firstNonEmptyStorageEnv(os.Getenv("S3_BUCKET_NAME"), os.Getenv("R2_BUCKET_NAME"))

	if (accountID == "" && endpoint == "") || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return fmt.Errorf("missing object storage config (S3_ENDPOINT or R2_ACCOUNT_ID, access key, secret key, bucket)")
	}

	var (
		cfg    aws.Config
		err    error
		client *s3.Client
	)
	if endpoint != "" {
		endpoint = strings.TrimRight(endpoint, "/")
		region := firstNonEmptyStorageEnv(os.Getenv("S3_REGION"), "us-east-1")
		cfg, err = awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
			awsconfig.WithRegion(region),
		)
		if err == nil {
			client = s3.NewFromConfig(cfg, func(options *s3.Options) {
				options.BaseEndpoint = aws.String(endpoint)
				options.UsePathStyle = true
				options.Region = region
			})
		}
	} else {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint}, nil
		})
		cfg, err = awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithEndpointResolverWithOptions(resolver),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
			awsconfig.WithRegion("auto"),
		)
		if err == nil {
			client = s3.NewFromConfig(cfg)
		}
	}
	if err != nil {
		return fmt.Errorf("object storage SDK config error: %w", err)
	}

	p.client = client
	p.bucketName = bucketName
	p.endpoint = endpoint
	return nil
}

func (p *PersisterR2) Save(fh *multipart.FileHeader, fileName string, fileExtension string) (string, error) {
	if err := p.init(); err != nil {
		return "", err
	}

	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	objectKey := fmt.Sprintf("%s.%s", fileName, fileExtension)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("object storage upload error: %w", err)
	}

	// Preserve the existing endpoint/bucket/key URI format for stored metadata.
	objectURI := fmt.Sprintf("%s/%s/%s", strings.TrimRight(p.endpoint, "/"), p.bucketName, objectKey)
	return objectURI, nil
}

func (p *PersisterR2) LoadFile(fileName string) (string, *bytes.Buffer, error) {
	if err := p.init(); err != nil {
		return "", nil, err
	}

	// If fileName is our own object-storage URL, use the authenticated S3 client.
	if strings.HasPrefix(fileName, "http") {
		if objectKey, ok := p.objectKeyFromURI(fileName); ok {
			return p.downloadFromR2(objectKey)
		}

		// Fallback: anonymous HTTP GET for other external URLs
		return p.downloadViaHTTP(fileName)
	}

	// Otherwise, it is a raw key. Retrieve it using GetObject from R2.
	return p.downloadFromR2(fileName)
}

func (p *PersisterR2) objectKeyFromURI(fileName string) (string, bool) {
	objectURL, err := url.Parse(fileName)
	if err != nil {
		return "", false
	}
	storageURL, err := url.Parse(p.endpoint)
	if err != nil || !strings.EqualFold(objectURL.Scheme, storageURL.Scheme) || !strings.EqualFold(objectURL.Host, storageURL.Host) {
		return "", false
	}

	parts := strings.SplitN(strings.TrimPrefix(objectURL.Path, "/"), "/", 2)
	if len(parts) != 2 || parts[0] != p.bucketName || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func (p *PersisterR2) downloadFromR2(objectKey string) (string, *bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", nil, fmt.Errorf("object storage download error: %w", err)
	}
	defer output.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, output.Body)
	if err != nil {
		return "", nil, err
	}

	return objectKey, buf, nil
}

func (p *PersisterR2) downloadViaHTTP(fileURL string) (string, *bytes.Buffer, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to download via HTTP: status %d", resp.StatusCode)
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return fileURL, buf, nil
}
