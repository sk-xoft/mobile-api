package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/smartseaman/mobile-api/internal/config"
)

type Client struct {
	client *minio.Client
	bucket string
}

func New(cfg config.Storage) (*Client, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return &Client{}, nil
	}
	parsed, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse object storage endpoint: %w", err)
	}
	client, err := minio.New(parsed.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: parsed.Scheme == "https",
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create object storage client: %w", err)
	}
	return &Client{client: client, bucket: cfg.Bucket}, nil
}

func (c *Client) Put(ctx context.Context, key, contentType string, content []byte) error {
	if c.client == nil {
		return fmt.Errorf("object storage is not configured")
	}
	_, err := c.client.PutObject(ctx, c.bucket, strings.TrimPrefix(key, "/"), bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, string, error) {
	if c.client == nil {
		return nil, "", fmt.Errorf("object storage is not configured")
	}
	object, err := c.client.GetObject(ctx, c.bucket, strings.TrimPrefix(key, "/"), minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer object.Close()
	info, err := object.Stat()
	if err != nil {
		return nil, "", err
	}
	content, err := io.ReadAll(object)
	if err != nil {
		return nil, "", err
	}
	return content, info.ContentType, nil
}
