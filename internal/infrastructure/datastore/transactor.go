package datastore

import (
	"context"

	"gorm.io/gorm"
)

// Transactor handles database transactions.
type Transactor struct {
	db *DB
}

// NewTransactor creates a new Transactor.
func NewTransactor(db *DB) *Transactor {
	return &Transactor{db: db}
}

// RunInTx runs a function within a transaction.
func (t *Transactor) RunInTx(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}
