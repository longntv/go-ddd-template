package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/wire"
)

// WireSet holds the Wire providers for S3.
var WireSet = wire.NewSet(
	NewClient,
	NewReader,
	NewWriter,
)

// NewClient creates a new S3 client. With a custom endpoint (LocalStack) it
// uses path-style addressing, since virtual-hosted bucket hostnames such as
// bucket.localhost don't resolve.
func NewClient(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.BaseEndpoint != nil
	})
}
