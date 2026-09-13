package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type ProductService interface {
	ListProducts(ctx context.Context) ([]*model.Product, error)
}
