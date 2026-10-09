package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Reader reads files from S3.
type Reader struct {
	client *s3.Client
}

// NewReader creates a new S3 reader.
func NewReader(client *s3.Client) *Reader {
	return &Reader{
		client: client,
	}
}

// Get reads a file from S3.
func (r *Reader) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	output, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// GetURL returns a presigned URL for a file.
func (r *Reader) GetURL(ctx context.Context, bucket, key string) (string, error) {
	presignClient := s3.NewPresignClient(r.client)
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	return presignResult.URL, nil
}
