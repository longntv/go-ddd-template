package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
)

// publishBestEffort publishes evt after the change it describes is already
// saved. A failure must not fail the request, so it is logged instead of
// returned; the log carries the event ID so the event can be replayed.
// logger must not be nil.
func publishBestEffort(ctx context.Context, publisher gateway.EventPublisher, logger *zap.Logger, evt *event.UserEvent) {
	if err := publisher.Publish(ctx, evt); err != nil {
		logger.Error("failed to publish event",
			zap.String("event_id", evt.ID.String()),
			zap.String("event_type", evt.Type),
			zap.String("subject", evt.Subject),
			zap.Error(err),
		)
	}
}
