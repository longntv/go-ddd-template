package s3

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Writer writes files to S3.
type Writer struct {
	client *s3.Client
}

// NewWriter creates a new S3 writer.
func NewWriter(client *s3.Client) *Writer {
	return &Writer{
		client: client,
	}
}

// Put writes a file to S3.
func (w *Writer) Put(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	_, err := w.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

// Delete deletes a file from S3.
func (w *Writer) Delete(ctx context.Context, bucket, key string) error {
	_, err := w.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}
