package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewS3Client(endpoint, region, bucket, accessKey, secretKey, publicBaseURL string) (*S3Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	return &S3Client{client: client, bucket: bucket, publicURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

type UploadResult struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// Upload stores a file in S3 and returns its public URL.
func (c *S3Client) Upload(ctx context.Context, reader io.Reader, contentType, originalFilename string) (*UploadResult, error) {
	ext := path.Ext(originalFilename)
	key := fmt.Sprintf("images/%s/%s%s", time.Now().UTC().Format("2006/01"), uuid.New().String(), ext)

	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to s3: %w", err)
	}

	// The public URL is configured, not inferred: Cloudflare R2 and other
	// S3-compatible stores serve from their own domain, so assuming the AWS
	// hostname returned a URL that did not resolve.
	base := c.publicURL
	if base == "" {
		base = fmt.Sprintf("https://%s.s3.amazonaws.com", c.bucket)
	}

	return &UploadResult{URL: base + "/" + key, Key: key}, nil
}
