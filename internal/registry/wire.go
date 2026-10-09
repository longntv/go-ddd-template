//go:build wireinject

package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/longntv/go-ddd-template/internal/config"
	handlercloudevents "github.com/longntv/go-ddd-template/internal/handler/cloudevents"
	"github.com/longntv/go-ddd-template/internal/handler/health"
	"github.com/longntv/go-ddd-template/internal/handler/http"
	"github.com/longntv/go-ddd-template/internal/handler/http/server"
	"github.com/longntv/go-ddd-template/internal/infrastructure/aws"
	infracloudevents "github.com/longntv/go-ddd-template/internal/infrastructure/cloudevents"
	"github.com/longntv/go-ddd-template/internal/infrastructure/datastore"
	"github.com/longntv/go-ddd-template/internal/service"
)

//go:generate go run github.com/google/wire/cmd/wire@latest

// InitializeServer initializes the HTTP server with the necessary dependencies.
func InitializeServer(
	cfg *config.Config,
	logger *zap.Logger,
) (*gin.Engine, func(), error) {
	wire.Build(
		// Infrastructure
		aws.WireSet,
		datastore.WireSet,

		// CloudEvents
		infracloudevents.WireSet,

		// Services
		service.WireSet,

		// Handlers
		health.WireSet,
		server.WireSet,
		http.WireSet,
	)

	return nil, nil, nil
}

// InitializeSubscriber initializes the CloudEvents subscriber with the
// necessary dependencies.
func InitializeSubscriber(
	cfg *config.Config,
	logger *zap.Logger,
) (*infracloudevents.Subscriber, func(), error) {
	wire.Build(
		// Infrastructure
		aws.WireSet,

		// CloudEvents
		infracloudevents.WireSet,
		handlercloudevents.WireSet,
	)

	return nil, nil, nil
}
