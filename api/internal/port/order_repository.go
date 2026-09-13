package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*model.Order, error)
	Create(ctx context.Context, order *model.Order) error
}
