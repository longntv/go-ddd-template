//go:build wireinject

package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/handler/health"
	"github.com/longntv/go-ddd-template/internal/handler/http"
	"github.com/longntv/go-ddd-template/internal/handler/http/server"
	"github.com/longntv/go-ddd-template/internal/infrastructure/datastore"
	"github.com/longntv/go-ddd-template/internal/service"
)

//go:generate go run github.com/google/wire/cmd/wire@v0.6.0

// InitializeServer builds the real HTTP router for integration tests.
//
// Unlike internal/registry, external dependencies are parameters: the test
// passes its own database and a mock event publisher, so the whole stack
// from router to Postgres runs for real without AWS.
func InitializeServer(
	gormDB *gorm.DB,
	eventPublisher gateway.EventPublisher,
) (*gin.Engine, error) {
	wire.Build(
		datastore.NewUserReader,
		datastore.NewUserWriter,
		service.WireSet,
		health.WireSet,
		server.WireSet,
		http.WireSet,
	)

	return nil, nil
}
