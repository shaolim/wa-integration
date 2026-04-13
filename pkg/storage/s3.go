package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}

type S3Uploader struct {
	client *s3.Client
	bucket string
}

func NewS3Uploader(ctx context.Context, region, bucket, accessKey, secretKey string) (*S3Uploader, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &S3Uploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// ListObjects returns all objects in the bucket.
func (u *S3Uploader) ListObjects(ctx context.Context) ([]Object, error) {
	out, err := u.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(u.bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 list objects: %w", err)
	}

	objects := make([]Object, 0, len(out.Contents))
	for _, obj := range out.Contents {
		objects = append(objects, Object{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
		})
	}
	return objects, nil
}

// GetPresignedURL returns a temporary URL for reading the object at key.
func (u *S3Uploader) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(u.client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign get object key=%s: %w", key, err)
	}
	return req.URL, nil
}

// Upload puts data into S3 at the given key with the provided MIME type.
func (u *S3Uploader) Upload(ctx context.Context, key string, data []byte, mimeType string) error {
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(mimeType),
	})
	if err != nil {
		return fmt.Errorf("s3 put object key=%s: %w", key, err)
	}
	return nil
}
