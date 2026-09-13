package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type OrderRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	Create(ctx context.Context, order *model.Order) error
}
