package postgres

import (
	"context"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"

	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepository struct {
	// pool is set only when q is not already bound to an externally
	// managed transaction (see newTxOrderRepository / UnitOfWork.Execute).
	// When set, Create opens and manages its own transaction so the order
	// and its items are still written atomically on their own.
	pool *pgxpool.Pool
	q    *Queries
}

func NewOrderRepository(pool *pgxpool.Pool) port.OrderRepository {
	return &orderRepository{pool: pool, q: New(pool)}
}

// newTxOrderRepository returns an OrderRepository whose Create executes
// directly against q instead of opening its own transaction, so it composes
// with sibling writes (e.g. ProductRepository.UpdateLastOrderAt) made under
// the same caller-managed transaction. Used by UnitOfWork.Execute.
func newTxOrderRepository(q *Queries) port.OrderRepository {
	return &orderRepository{q: q}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	if r.pool == nil {
		return r.insert(ctx, r.q, order)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.insert(ctx, r.q.WithTx(tx), order); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// insert writes order and its items using q, without making any decision
// about transaction boundaries -- that's Create's job.
func (r *orderRepository) insert(ctx context.Context, q *Queries, order *model.Order) error {
	row, err := q.CreateOrder(ctx, CreateOrderParams{
		MemberCardNumber: nullableString(order.MemberCardNumber),
		TotalPrice:       order.TotalPrice,
		DiscountAmount:   order.DiscountAmount,
	})
	if err != nil {
		return err
	}
	order.ID = row.ID
	order.CreatedAt = row.CreatedAt.Time

	for _, item := range order.Items {
		itemRow, err := q.CreateOrderItem(ctx, CreateOrderItemParams{
			OrderID:   row.ID,
			ProductID: item.ProductID,
			Quantity:  int32(item.Quantity),
			UnitPrice: item.UnitPrice,
		})
		if err != nil {
			return err
		}
		item.ID = itemRow.ID
		item.OrderID = itemRow.OrderID
	}

	return nil
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	row, err := r.q.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	itemRows, err := r.q.ListOrderItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}

	order := toModelOrder(row)
	for _, itemRow := range itemRows {
		order.Items = append(order.Items, toModelOrderItem(itemRow))
	}
	return order, nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toModelOrder(o Order) *model.Order {
	var memberCardNumber string
	if o.MemberCardNumber != nil {
		memberCardNumber = *o.MemberCardNumber
	}

	return &model.Order{
		ID:               o.ID,
		MemberCardNumber: memberCardNumber,
		TotalPrice:       o.TotalPrice,
		DiscountAmount:   o.DiscountAmount,
		CreatedAt:        o.CreatedAt.Time,
	}
}

func toModelOrderItem(i OrderItem) *model.OrderItem {
	return &model.OrderItem{
		ID:        i.ID,
		OrderID:   i.OrderID,
		ProductID: i.ProductID,
		Quantity:  int(i.Quantity),
		UnitPrice: i.UnitPrice,
	}
}
