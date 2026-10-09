package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/wire"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
)

// Publisher publishes messages to SNS.
type Publisher struct {
	client   *sns.Client
	topicARN string
}

// NewPublisher creates a new SNS publisher.
func NewPublisher(client *sns.Client, topicARN string) *Publisher {
	return &Publisher{
		client:   client,
		topicARN: topicARN,
	}
}

// NewPublisherFromConfig creates a new SNS publisher from config.
func NewPublisherFromConfig(client *sns.Client, cfg *appconfig.Config) *Publisher {
	return NewPublisher(client, cfg.SNS.TopicARN)
}

// Publish publishes a message to the SNS topic.
func (p *Publisher) Publish(ctx context.Context, message string, attributes map[string]types.MessageAttributeValue) error {
	_, err := p.client.Publish(ctx, &sns.PublishInput{
		TopicArn:          aws.String(p.topicARN),
		Message:           aws.String(message),
		MessageAttributes: attributes,
	})
	return err
}

// NewClient creates a new SNS client.
func NewClient(cfg aws.Config) *sns.Client {
	return sns.NewFromConfig(cfg)
}

// WireSet holds the Wire providers for SNS.
var WireSet = wire.NewSet(
	NewClient,
	NewPublisherFromConfig,
)
