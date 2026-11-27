package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	PublicBaseURL   string
	ForcePathStyle  bool
}

type S3Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func NewS3Storage(cfg Config) (*S3Storage, error) {
	if cfg.Region == "" || cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.PublicBaseURL == "" {
		return nil, fmt.Errorf("storage configuration is incomplete: region, bucket, access key, secret key, and public base URL are required")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})

	return &S3Storage{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func NewS3StorageFromEnv() (*S3Storage, error) {
	forcePathStyle := true
	if value := os.Getenv("S3_FORCE_PATH_STYLE"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid S3_FORCE_PATH_STYLE %q: %w", value, err)
		}
		forcePathStyle = parsed
	}

	return NewS3Storage(Config{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("S3_REGION"),
		Bucket:          os.Getenv("S3_BUCKET"),
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		PublicBaseURL:   os.Getenv("S3_PUBLIC_BASE_URL"),
		ForcePathStyle:  forcePathStyle,
	})
}

func (s *S3Storage) Upload(ctx context.Context, key, contentType string, r io.Reader, size int64) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	}
	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return "", fmt.Errorf("failed to upload %s: %w", key, err)
	}

	return s.PublicURL(key), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("failed to delete %s: %w", key, err)
	}
	return nil
}

func (s *S3Storage) DeletePrefix(ctx context.Context, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("refusing to delete an empty prefix")
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
		}
		for _, object := range page.Contents {
			if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    object.Key,
			}); err != nil {
				return fmt.Errorf("failed to delete %s: %w", aws.ToString(object.Key), err)
			}
		}
	}

	return nil
}

func (s *S3Storage) PublicURL(key string) string {
	return s.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}
