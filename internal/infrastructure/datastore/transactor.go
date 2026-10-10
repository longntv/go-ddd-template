package datastore

import (
	"context"

	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/gateway"
)

// txKey is the context key under which RunInTx stores the open transaction.
type txKey struct{}

// conn returns the transaction that Transactor.RunInTx bound to ctx, or db
// when ctx has none. Every reader and writer queries through it so its calls
// join the caller's transaction.
func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := txFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

func txFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok
}

// transactor implements gateway.Transactor.
type transactor struct {
	db *gorm.DB
}

// NewTransactor creates a new Transactor.
func NewTransactor(db *gorm.DB) gateway.Transactor {
	return &transactor{db: db}
}

// RunInTx runs fn in a transaction. Called inside another RunInTx, it runs fn
// in a savepoint of the outer transaction.
func (t *transactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return conn(ctx, t.db).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}
