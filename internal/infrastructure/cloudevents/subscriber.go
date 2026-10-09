package cloudevents

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cloudevents/sdk-go/v2/event"
	"go.uber.org/zap"

	"github.com/longntv/go-ddd-template/internal/infrastructure/aws/sqs"
	"github.com/longntv/go-ddd-template/internal/infrastructure/cloudevents/handler"
)

// Subscriber subscribes to SQS messages and converts them to CloudEvents.
type Subscriber struct {
	sqsClient *sqs.Subscriber
	mux       *handler.Mux
	logger    *zap.Logger
	wg        sync.WaitGroup
	stopOnce  sync.Once
	stopChan  chan struct{}
}

// NewSubscriber creates a new Subscriber.
func NewSubscriber(
	sqsClient *sqs.Subscriber,
	mux *handler.Mux,
	logger *zap.Logger,
) *Subscriber {
	return &Subscriber{
		sqsClient: sqsClient,
		mux:       mux,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}
}

// Receive receives messages from SQS.
func (s *Subscriber) Receive(ctx context.Context) ([]*EventMessage, error) {
	messages, err := s.sqsClient.Receive(ctx)
	if err != nil {
		return nil, err
	}

	eventMessages := make([]*EventMessage, 0, len(messages))
	for i := range messages {
		msg := &messages[i]
		eventMsg := &EventMessage{
			SQSMessage: msg,
		}

		// Parse body as CloudEvent
		var cloudevent event.Event
		if msg.Body != nil {
			if err := json.Unmarshal([]byte(*msg.Body), &cloudevent); err != nil {
				// Invalid message format, skip
				continue
			}
		}

		eventMsg.CloudEvent = &cloudevent
		eventMessages = append(eventMessages, eventMsg)
	}

	return eventMessages, nil
}

// Delete deletes a message from SQS.
func (s *Subscriber) Delete(ctx context.Context, msg *EventMessage) error {
	return s.sqsClient.Delete(ctx, msg.SQSMessage)
}

// EventMessage wraps a CloudEvent with its SQS message.
type EventMessage struct {
	CloudEvent *event.Event
	SQSMessage *types.Message
}

// Run starts the subscriber to continuously process messages.
func (s *Subscriber) Run(ctx context.Context) error {
	s.logger.Info("subscriber running")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("subscriber context cancelled")
			return ctx.Err()
		case <-s.stopChan:
			s.logger.Info("subscriber stop signal received")
			return nil
		default:
			// Process messages
			if err := s.processMessages(ctx); err != nil {
				s.logger.Error("failed to process messages", zap.Error(err))
				// Continue processing even if there's an error
				// Add a small delay to avoid tight loops on persistent errors
				time.Sleep(time.Second)
			}
		}
	}
}

// processMessages receives and processes a batch of messages.
func (s *Subscriber) processMessages(ctx context.Context) error {
	messages, err := s.Receive(ctx)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	s.logger.Debug("received messages", zap.Int("count", len(messages)))

	for _, msg := range messages {
		if err := s.handleMessage(ctx, msg); err != nil {
			s.logger.Error("failed to handle message",
				zap.Error(err),
				zap.String("message_id", *msg.SQSMessage.MessageId),
			)
			// Continue with other messages even if one fails
			continue
		}

		// Delete successfully processed message
		if err := s.Delete(ctx, msg); err != nil {
			s.logger.Error("failed to delete message",
				zap.Error(err),
				zap.String("message_id", *msg.SQSMessage.MessageId),
			)
		}
	}

	return nil
}

// handleMessage processes a single event message using the registered handlers.
func (s *Subscriber) handleMessage(ctx context.Context, msg *EventMessage) error {
	s.logger.Debug("handling message",
		zap.String("event_type", msg.CloudEvent.Type()),
		zap.String("event_id", msg.CloudEvent.ID()),
	)

	if err := s.mux.Handle(ctx, msg.CloudEvent); err != nil {
		return err
	}

	s.logger.Debug("message handled successfully",
		zap.String("event_type", msg.CloudEvent.Type()),
		zap.String("event_id", msg.CloudEvent.ID()),
	)

	return nil
}

// Shutdown gracefully shuts down the subscriber.
func (s *Subscriber) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down subscriber")

	s.stopOnce.Do(func() {
		close(s.stopChan)
	})

	// Wait for all goroutines to finish or context to be done
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("subscriber shutdown completed")
		return nil
	case <-ctx.Done():
		s.logger.Warn("subscriber shutdown timed out")
		return ctx.Err()
	}
}
