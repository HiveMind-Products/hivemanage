package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	MinIO = "minio"
	R2    = "r2"
	S3    = "s3"
)

type Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func New(s3Provider string) (*Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, fmt.Errorf("load storage config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		if s3Provider == MinIO {
			o.UsePathStyle = true
		}
	})
	bucket := os.Getenv("AWS_BUCKET")
	if bucket == "" {
		return nil, errors.New("environment variable AWS_BUCKET is not set")
	}
	return &Storage{client: client, presignClient: s3.NewPresignClient(client), bucket: bucket}, nil
}

func (r *Storage) UploadFile(ctx context.Context, file io.Reader, key, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key), Body: file, ContentType: aws.String(contentType)})
	if err != nil {
		return fmt.Errorf("upload file to storage: %w", err)
	}
	return nil
}

func (r *Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete storage object: %w", err)
	}
	return nil
}

func (r *Storage) SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	res, err := r.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)}, func(options *s3.PresignOptions) { options.Expires = expiry })
	if err != nil {
		return "", fmt.Errorf("create signed url: %w", err)
	}
	return res.URL, nil
}

func (r *Storage) CreateBucket(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(r.bucket)})
	if err == nil {
		slog.Info(fmt.Sprintf("bucket [%s] already exists", r.bucket))
		return nil
	}
	var nfe *types.NotFound
	if !errors.As(err, &nfe) {
		return fmt.Errorf("failed to check for bucket: %w", err)
	}
	_, err = r.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(r.bucket)})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	slog.Info(fmt.Sprintf("created private bucket: %s", r.bucket))
	return nil
}
