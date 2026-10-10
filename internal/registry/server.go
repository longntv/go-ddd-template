package registry

import (
	"github.com/gin-gonic/gin"

	"github.com/longntv/go-ddd-template/internal/infrastructure/datastore"
)

// Server is what cmd/server runs: the HTTP router and, beside it, the relay
// that publishes the events the request handlers saved to the outbox.
type Server struct {
	Router      *gin.Engine
	OutboxRelay *datastore.OutboxRelay
}
