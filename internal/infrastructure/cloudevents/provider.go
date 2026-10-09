package cloudevents

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/wire"

	domainevent "github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	handlercloudevents "github.com/longntv/go-ddd-template/internal/handler/cloudevents"
	"github.com/longntv/go-ddd-template/internal/infrastructure/cloudevents/handler"
)

// WireSet holds the Wire providers for CloudEvents infrastructure.
var WireSet = wire.NewSet(
	NewSubscriber,
	NewPublisher,
	wire.Bind(new(gateway.EventPublisher), new(*Publisher)),
	ProvideConfiguredMux,
)

// ProvideConfiguredMux creates and configures a Mux with registered handlers.
func ProvideConfiguredMux(ceHandler *handlercloudevents.Handler) *handler.Mux {
	mux := handler.NewMux()

	// Register handlers for different event types
	mux.Register(domainevent.UserCreatedEvent, &handlerAdapter{
		handleFunc: ceHandler.HandleUserCreated,
	})
	mux.Register(domainevent.UserUpdatedEvent, &handlerAdapter{
		handleFunc: ceHandler.HandleUserUpdated,
	})
	mux.Register(domainevent.UserDeletedEvent, &handlerAdapter{
		handleFunc: ceHandler.HandleUserDeleted,
	})

	return mux
}

// handlerAdapter adapts a handler function to the Handler interface.
type handlerAdapter struct {
	handleFunc func(ctx context.Context, evt *event.Event) error
}

func (h *handlerAdapter) Handle(ctx context.Context, evt *event.Event) error {
	return h.handleFunc(ctx, evt)
}
