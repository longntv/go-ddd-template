package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/google/wire"

	"github.com/longntv/go-ddd-template/internal/infrastructure/aws/bedrock"
	"github.com/longntv/go-ddd-template/internal/infrastructure/aws/s3"
	"github.com/longntv/go-ddd-template/internal/infrastructure/aws/sns"
	"github.com/longntv/go-ddd-template/internal/infrastructure/aws/sqs"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
)

// WireSet holds the Wire providers for AWS infrastructure.
var WireSet = wire.NewSet(
	LoadConfig,
	s3.WireSet,
	bedrock.WireSet,
	sqs.WireSet,
	sns.WireSet,
)

// LoadConfig loads the AWS configuration using the provided application
// configuration.
func LoadConfig(cfg *appconfig.Config) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.AWS.Region),
	}

	// For local development with LocalStack: every service client sends
	// requests to this endpoint instead of the AWS one.
	if cfg.AWS.EndpointURL != "" {
		opts = append(opts, config.WithBaseEndpoint(cfg.AWS.EndpointURL))
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return aws.Config{}, err
	}

	return awsConfig, nil
}
