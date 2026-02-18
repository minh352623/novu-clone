package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"CONVERDA/global"
	"CONVERDA/internal/r2/service/dto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Service interface for R2 operations
type R2Service interface {
	// UploadFileBase64 uploads multiple files from base64 encoded strings
	UploadFileBase64(ctx context.Context, images []dto.ImageDTO) (*dto.UploadResponse, error)

	// UploadSingleFileBase64 uploads a single file from base64 encoded string
	UploadSingleFileBase64(ctx context.Context, key, base64Data string) (string, error)

	// DeleteFile deletes a file from R2
	DeleteFile(ctx context.Context, key string) error
}

type r2Service struct {
	client    *s3.Client
	bucket    string
	url       string
	publicUrl string
}

// NewR2Service creates a new R2 service
func NewR2Service() (R2Service, error) {
	r2Config := global.Config.R2

	// Create custom resolver for Cloudflare R2
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: r2Config.Url,
		}, nil
	})

	// Load AWS config with custom credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			r2Config.AccessKeyId,
			r2Config.SecretAccessKey,
			r2Config.Token,
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &r2Service{
		client:    client,
		bucket:    r2Config.Bucket,
		url:       r2Config.Url,
		publicUrl: r2Config.PublicUrl,
	}, nil
}

// UploadFileBase64 uploads multiple files from base64 encoded strings
func (s *r2Service) UploadFileBase64(ctx context.Context, images []dto.ImageDTO) (*dto.UploadResponse, error) {
	response := &dto.UploadResponse{
		Results:    make([]dto.UploadResult, len(images)),
		TotalCount: len(images),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, img := range images {
		wg.Add(1)
		go func(index int, image dto.ImageDTO) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					global.Logger.Error("r2_service: panic recovered in UploadFileBase64 goroutine", "panic", r, "key", image.Key)

					// Record failure for this item
					mu.Lock()
					response.Results[index] = dto.UploadResult{
						Key:     image.Key,
						Success: false,
						Error:   fmt.Sprintf("internal server error: panic recovered: %v", r),
					}
					response.FailedCount++
					mu.Unlock()
				}
			}()

			result := dto.UploadResult{
				Key: image.Key,
			}

			asyncCtx := context.WithoutCancel(ctx)
			url, err := s.UploadSingleFileBase64(asyncCtx, image.Key, image.Base64)
			if err != nil {
				result.Success = false
				result.Error = err.Error()
			} else {
				result.Success = true
				result.URL = url
			}

			mu.Lock()
			response.Results[index] = result
			if result.Success {
				response.SuccessCount++
			} else {
				response.FailedCount++
			}
			mu.Unlock()
		}(i, img)
	}

	wg.Wait()

	return response, nil
}

// UploadSingleFileBase64 uploads a single file from base64 encoded string
func (s *r2Service) UploadSingleFileBase64(ctx context.Context, key, base64Data string) (string, error) {
	// Remove data URL prefix if present (e.g., "data:image/png;base64,")
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(data)

	// Upload to R2
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err = s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to R2: %w", err)
	}

	// Generate URL
	url := s.generateURL(key)

	return url, nil
}

// DeleteFile deletes a file from R2
func (s *r2Service) DeleteFile(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	return nil
}

// generateURL generates the public URL for the uploaded file
func (s *r2Service) generateURL(key string) string {
	if s.publicUrl != "" {
		// Use the public URL if configured (e.g., https://pub-xxx.r2.dev)
		baseURL := strings.TrimSuffix(s.publicUrl, "/")
		return fmt.Sprintf("%s/%s", baseURL, key)
	}
	// Fallback to internal URL with bucket
	baseURL := strings.TrimSuffix(s.url, "/")
	return fmt.Sprintf("%s/%s/%s", baseURL, s.bucket, key)
}
