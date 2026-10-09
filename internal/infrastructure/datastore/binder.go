package datastore

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/gateway"
)

// binder implements the gateway.Binder interface,
// providing a mechanism to bind a GormDB instance to a context.
type binder struct {
	db *gorm.DB
}

// NewBinder creates a new instance of gateway.Binder
// with the provided GormDB connection.
func NewBinder(db *gorm.DB) gateway.Binder {
	return &binder{
		db: db,
	}
}

// Bind binds the GormDB instance to the context and returns a new context
// containing the database connection, typically for use in request-scoped operations.
func (c *binder) Bind(ctx context.Context) context.Context {
	return WithDB(ctx, c.db)
}

type dbKey struct{} // type used to store the GormDB in context

// WithDB injects a GormDB instance into the provided context.
func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, &dbKey{}, db)
}

// DBFromContext retrieves the GormDB instance from the context.
func DBFromContext(ctx context.Context) (*gorm.DB, error) {
	if v := ctx.Value(&dbKey{}); v != nil {
		db, ok := v.(*gorm.DB)
		if ok {
			return db, nil
		}
	}
	return nil, errors.New("no db found in context")
}
