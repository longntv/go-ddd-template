package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/wire"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
)

// Subscriber receives messages from SQS.
type Subscriber struct {
	client      *sqs.Client
	queueURL    string
	maxMessages int32
	waitTime    int32
}

// NewSubscriberFromConfig creates a new SQS subscriber from config.
func NewSubscriberFromConfig(client *sqs.Client, cfg *appconfig.Config) *Subscriber {
	return &Subscriber{
		client:      client,
		queueURL:    cfg.SQS.QueueURL,
		maxMessages: int32(cfg.SQS.MaxMessages),
		waitTime:    int32(cfg.SQS.WaitTimeSeconds),
	}
}

// Receive receives messages from SQS.
func (s *Subscriber) Receive(ctx context.Context) ([]types.Message, error) {
	output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(s.queueURL),
		MaxNumberOfMessages:   s.maxMessages,
		WaitTimeSeconds:       s.waitTime,
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		return nil, err
	}
	return output.Messages, nil
}

// Delete deletes a message from SQS.
func (s *Subscriber) Delete(ctx context.Context, message *types.Message) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}

// NewClient creates a new SQS client.
func NewClient(cfg aws.Config) *sqs.Client {
	return sqs.NewFromConfig(cfg)
}

// WireSet holds the Wire providers for SQS.
var WireSet = wire.NewSet(
	NewClient,
	NewSubscriberFromConfig,
)
