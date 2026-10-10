package gateway

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock/transaction.go -source=transaction.go

// Transactor runs work in a database transaction.
type Transactor interface {
	// RunInTx runs fn in a transaction, committed when fn returns nil and
	// rolled back otherwise. Gateway calls made with the ctx passed to fn take
	// part in the transaction; calls made with the outer ctx do not.
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
