package microservice

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// PersisterS3 implement IPersisterFile สำหรับ SeaweedFS S3 gateway
type PersisterS3 struct {
	client         *s3.Client
	bucketName     string
	publicEndpoint string
	initOnce       sync.Once
	initErr        error
}

func NewPersisterS3() *PersisterS3 {
	return &PersisterS3{}
}

func (p *PersisterS3) init() {
	p.initOnce.Do(func() {
		endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
		accessKeyID := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
		secretAccessKey := strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY"))
		p.bucketName = strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))
		p.publicEndpoint = strings.TrimSpace(os.Getenv("S3_PUBLIC_ENDPOINT"))

		if endpoint == "" || accessKeyID == "" || secretAccessKey == "" || p.bucketName == "" {
			p.initErr = fmt.Errorf("missing S3 config (S3_ENDPOINT, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY, S3_BUCKET_NAME)")
			return
		}

		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
			}, nil
		})

		cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithEndpointResolverWithOptions(resolver),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
			awsconfig.WithRegion("us-east-1"),
		)
		if err != nil {
			p.initErr = fmt.Errorf("S3 SDK config error: %w", err)
			return
		}

		p.client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
		})

		// Auto-create bucket
		_, headErr := p.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: aws.String(p.bucketName),
		})
		if headErr != nil {
			p.client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: aws.String(p.bucketName),
			})
		}
	})
}

func (p *PersisterS3) Save(fh *multipart.FileHeader, fileName string, fileExtension string) (string, error) {
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

	_, err = p.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(p.bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(fh.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("S3 upload error: %w", err)
	}

	// Return public URL (เหมือน Azure Blob ที่ return full URL)
	publicBase := p.publicEndpoint
	if publicBase == "" {
		publicBase = strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	}
	objectURI := fmt.Sprintf("%s/%s/%s", strings.TrimRight(publicBase, "/"), p.bucketName, objectKey)

	return objectURI, nil
}

func (p *PersisterS3) LoadFile(fileName string) (string, *bytes.Buffer, error) {
	p.init()
	if p.initErr != nil {
		return "", nil, p.initErr
	}

	// ถ้า fileName เป็น URL → HTTP GET (เหมือน Azure Blob pattern)
	if strings.HasPrefix(fileName, "http") {
		resp, err := http.Get(fileName)
		if err != nil {
			return "", nil, err
		}
		defer resp.Body.Close()

		buf := &bytes.Buffer{}
		buf.ReadFrom(resp.Body)
		return fileName, buf, nil
	}

	// ถ้าไม่ใช่ URL → GetObject จาก S3
	output, err := p.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return "", nil, fmt.Errorf("S3 download error: %w", err)
	}
	defer output.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, output.Body)
	if err != nil {
		return "", nil, err
	}

	return fileName, buf, nil
}
