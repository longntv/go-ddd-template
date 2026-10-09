package bedrock

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/google/wire"
)

// WireSet holds the Wire providers for Bedrock.
var WireSet = wire.NewSet(
	NewClient,
)

// NewClient creates a new Bedrock runtime client. A custom endpoint
// (AWS_ENDPOINT_URL) is inherited from cfg.
func NewClient(cfg aws.Config) *bedrockruntime.Client {
	return bedrockruntime.NewFromConfig(cfg)
}
