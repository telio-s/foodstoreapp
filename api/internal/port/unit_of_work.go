package port

import "context"

// TxRepositories bundles repository instances that all share a single
// database transaction, so a caller can compose multiple writes atomically.
type TxRepositories struct {
	Orders   OrderRepository
	Products ProductRepository
}

// UnitOfWork runs fn against repositories bound to a single database
// transaction: the transaction commits if fn returns nil, and rolls back
// (discarding every write fn made through repos) otherwise.
type UnitOfWork interface {
	Execute(ctx context.Context, fn func(repos TxRepositories) error) error
}
