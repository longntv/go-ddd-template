package bedrock

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/google/wire"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
)

// WireSet holds the Wire providers for Bedrock.
var WireSet = wire.NewSet(
	NewClient,
)

// NewClient creates a new Bedrock runtime client.
func NewClient(cfg aws.Config, appCfg *appconfig.Config) *bedrockruntime.Client {
	return bedrockruntime.NewFromConfig(cfg, func(o *bedrockruntime.Options) {
		if appCfg.AWS.EndpointURL != "" {
			o.BaseEndpoint = aws.String(appCfg.AWS.EndpointURL)
		}
	})
}
