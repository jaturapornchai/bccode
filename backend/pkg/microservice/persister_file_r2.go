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

// PersisterR2 implements IPersisterFile using Cloudflare R2 storage
type PersisterR2 struct {
	client     *s3.Client
	accountID  string
	bucketName string
	initOnce   sync.Once
	initErr    error
}

func NewPersisterR2() *PersisterR2 {
	return &PersisterR2{}
}

func (p *PersisterR2) init() {
	p.initOnce.Do(func() {
		p.accountID = strings.TrimSpace(os.Getenv("R2_ACCOUNT_ID"))
		accessKeyID := strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID"))
		secretAccessKey := strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY"))
		p.bucketName = strings.TrimSpace(os.Getenv("R2_BUCKET_NAME"))

		if p.accountID == "" || accessKeyID == "" || secretAccessKey == "" || p.bucketName == "" {
			p.initErr = fmt.Errorf("missing R2 config (R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME)")
			return
		}

		r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", p.accountID),
			}, nil
		})

		cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithEndpointResolverWithOptions(r2Resolver),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
			awsconfig.WithRegion("auto"),
		)
		if err != nil {
			p.initErr = fmt.Errorf("R2 SDK config error: %w", err)
			return
		}

		p.client = s3.NewFromConfig(cfg)
	})
}

func (p *PersisterR2) Save(fh *multipart.FileHeader, fileName string, fileExtension string) (string, error) {
	p.init()
	if p.initErr != nil {
		return "", p.initErr
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
		return "", fmt.Errorf("R2 upload error: %w", err)
	}

	// Construct standardized R2 S3 endpoint URL
	objectURI := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", p.accountID, p.bucketName, objectKey)
	return objectURI, nil
}

func (p *PersisterR2) LoadFile(fileName string) (string, *bytes.Buffer, error) {
	p.init()
	if p.initErr != nil {
		return "", nil, p.initErr
	}

	// If fileName is an HTTP URL, check if it's our own R2 URL first to use S3 client (authorized)
	if strings.HasPrefix(fileName, "http") {
		if u, err := url.Parse(fileName); err == nil {
			if strings.HasSuffix(u.Host, ".r2.cloudflarestorage.com") {
				parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
				if len(parts) == 2 {
					return p.downloadFromR2(parts[1])
				}
			}
		}

		// Fallback: anonymous HTTP GET for other external URLs
		return p.downloadViaHTTP(fileName)
	}

	// Otherwise, it is a raw key. Retrieve it using GetObject from R2.
	return p.downloadFromR2(fileName)
}

func (p *PersisterR2) downloadFromR2(objectKey string) (string, *bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", nil, fmt.Errorf("R2 download error: %w", err)
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
