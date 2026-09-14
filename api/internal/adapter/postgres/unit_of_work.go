package postgres

import (
	"context"

	"food-store-apis/internal/port"

	"github.com/jackc/pgx/v5/pgxpool"
)

type unitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) port.UnitOfWork {
	return &unitOfWork{pool: pool}
}

// Execute runs fn against repositories bound to a single pgx transaction:
// every write fn makes through repos commits together, or none do.
func (u *unitOfWork) Execute(ctx context.Context, fn func(repos port.TxRepositories) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := New(tx)
	repos := port.TxRepositories{
		Orders:   newTxOrderRepository(q),
		Products: &productRepository{q: q},
	}

	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
