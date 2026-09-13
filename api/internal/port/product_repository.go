package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*model.Product, error)
	List(ctx context.Context) ([]*model.Product, error)
	Create(ctx context.Context, product *model.Product) error
}
