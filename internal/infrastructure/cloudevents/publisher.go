package cloudevents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/infrastructure/aws/sns"
)

// Publisher publishes domain events to SNS.
type Publisher struct {
	snsClient *sns.Publisher
}

// NewPublisher creates a new Publisher.
func NewPublisher(snsClient *sns.Publisher) *Publisher {
	return &Publisher{
		snsClient: snsClient,
	}
}

// Publish publishes a domain event.
func (p *Publisher) Publish(ctx context.Context, evt *event.DomainEvent) error {
	ce, err := evt.ToCloudEvent()
	if err != nil {
		return err
	}

	// Convert CloudEvent to JSON
	data, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("marshal %s cloudevent: %w", ce.Type(), err)
	}

	// Create message attributes for filtering
	attributes := map[string]types.MessageAttributeValue{
		"type": {
			DataType:    awsString("String"),
			StringValue: awsString(ce.Type()),
		},
		"source": {
			DataType:    awsString("String"),
			StringValue: awsString(ce.Source()),
		},
	}

	if ce.Subject() != "" {
		attributes["subject"] = types.MessageAttributeValue{
			DataType:    awsString("String"),
			StringValue: awsString(ce.Subject()),
		}
	}

	// Publish to SNS
	return p.snsClient.Publish(ctx, string(data), attributes)
}

func awsString(s string) *string {
	return &s
}
