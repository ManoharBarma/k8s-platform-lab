// Package s3 provides a thin wrapper over the AWS SDK v2 S3 client with
// presigned URL helpers. It supports a custom endpoint + path-style addressing
// so it also works against MinIO for local development.
package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config controls the S3 client.
type Config struct {
	Region       string
	Bucket       string
	Endpoint     string // optional; set for MinIO
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

// Client wraps the S3 API and a presigner.
type Client struct {
	api     *s3.Client
	presign *s3.PresignClient
	bucket  string
}

// New builds an S3 Client. If AccessKey/SecretKey are empty, the default AWS
// credential chain is used (env, shared config, IRSA, etc.).
func New(ctx context.Context, cfg Config) (*Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Region)}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	api := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &Client{
		api:     api,
		presign: s3.NewPresignClient(api),
		bucket:  cfg.Bucket,
	}, nil
}

// Upload stores an object.
func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := c.api.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

// Delete removes an object.
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.api.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// PresignGet returns a time-limited URL to download an object.
func (c *Client) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	out, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

// PresignPut returns a time-limited URL clients can use to upload an object
// directly to S3 (offloading large uploads from the service).
func (c *Client) PresignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, error) {
	out, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}
