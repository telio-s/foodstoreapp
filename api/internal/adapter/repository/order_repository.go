package repository

import (
	"context"
	"database/sql"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) port.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order

	err := r.db.QueryRowContext(ctx,
		`SELECT id, member_card_number, total_price, discount_amount, created_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.MemberCardNumber, &o.TotalPrice, &o.DiscountAmount, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, product_id, quantity, unit_price
		 FROM order_items WHERE order_id = $1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (member_card_number, total_price, discount_amount, created_at)
		 VALUES ($1, $2, $3, now()) RETURNING id, created_at`,
		order.MemberCardNumber, order.TotalPrice, order.DiscountAmount,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		item.OrderID = order.ID
		err = tx.QueryRowContext(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			 VALUES ($1, $2, $3, $4) RETURNING id`,
			item.OrderID, item.ProductID, item.Quantity, item.UnitPrice,
		).Scan(&item.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
