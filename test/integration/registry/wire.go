//go:build wireinject

package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/handler/health"
	"github.com/longntv/go-ddd-template/internal/handler/http"
	"github.com/longntv/go-ddd-template/internal/handler/http/middleware"
	"github.com/longntv/go-ddd-template/internal/handler/http/server"
	"github.com/longntv/go-ddd-template/internal/infrastructure/datastore"
	"github.com/longntv/go-ddd-template/internal/service"
)

//go:generate go run github.com/google/wire/cmd/wire@v0.6.0

// InitializeServer builds the real HTTP router for integration tests.
//
// Unlike internal/registry, external dependencies are parameters: the test
// passes its own database and a fast password hasher, so the whole stack from
// router to Postgres runs for real without AWS. Events stop at the outbox
// table; the outbox relay has its own tests in the datastore package.
func InitializeServer(
	gormDB *gorm.DB,
	passwordHasher gateway.PasswordHasher,
) (*gin.Engine, error) {
	wire.Build(
		datastore.NewTransactor,
		datastore.NewUserReader,
		datastore.NewUserWriter,
		datastore.NewEventOutbox,
		service.WireSet,
		health.WireSet,
		server.WireSet,
		wire.Value(middleware.CORSConfig{}), // CORS is covered by middleware unit tests
		http.WireSet,
	)

	return nil, nil
}
