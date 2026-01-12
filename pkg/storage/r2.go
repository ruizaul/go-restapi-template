package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// R2Config holds Cloudflare R2 configuration
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string // Public URL for accessing files
}

// R2Client wraps the AWS S3 client for Cloudflare R2
type R2Client struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewR2Client creates a new R2 storage client
func NewR2Client(cfg R2Config) (*R2Client, error) {
	// Validate configuration
	if cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("configuración de R2 incompleta: se requieren account_id, access_key_id, secret_access_key y bucket_name")
	}

	// Construct R2 endpoint
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	// Create AWS config with R2 credentials
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"), // R2 uses "auto" as region
	)
	if err != nil {
		return nil, fmt.Errorf("error al cargar configuración de AWS SDK: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Client{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}, nil
}

// UploadFile uploads a file to R2 and returns the public URL
func (r *R2Client) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error al leer archivo: %w", err)
	}

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
	}

	// Generate unique filename with original extension
	ext := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Construct object key (path in bucket)
	objectKey := filepath.Join(folder, uniqueFilename)
	// Normalize path separators for S3 (always use forward slashes)
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	// Upload to R2
	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("error al subir archivo a R2: %w", err)
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(r.publicURL, "/"), objectKey)

	return publicURL, nil
}

// DeleteFile deletes a file from R2 by URL
func (r *R2Client) DeleteFile(ctx context.Context, fileURL string) error {
	// Extract object key from URL
	objectKey := strings.TrimPrefix(fileURL, r.publicURL+"/")
	objectKey = strings.TrimPrefix(objectKey, r.publicURL)
	objectKey = strings.TrimPrefix(objectKey, "/")

	if objectKey == "" {
		return fmt.Errorf("URL de archivo inválida: no se pudo extraer la clave del objeto")
	}

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("error al eliminar archivo de R2: %w", err)
	}

	return nil
}

// GeneratePresignedURL generates a presigned URL for temporary upload access
func (r *R2Client) GeneratePresignedURL(ctx context.Context, objectKey string, expirationMinutes int) (string, error) {
	presignClient := s3.NewPresignClient(r.client)

	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationMinutes) * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("error al generar URL presignada: %w", err)
	}

	return presignResult.URL, nil
}

// ListFiles lists all files in a folder
func (r *R2Client) ListFiles(ctx context.Context, folder string) ([]string, error) {
	prefix := folder
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	result, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("error al listar archivos de R2: %w", err)
	}

	var files []string
	for _, obj := range result.Contents {
		publicURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(r.publicURL, "/"), *obj.Key)
		files = append(files, publicURL)
	}

	return files, nil
}

// LoadConfigFromEnv loads R2 configuration from environment variables
func LoadConfigFromEnv() R2Config {
	return R2Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("R2_BUCKET_NAME"),
		PublicURL:       os.Getenv("R2_PUBLIC_URL"),
	}
}
