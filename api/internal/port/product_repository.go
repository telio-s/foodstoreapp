package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*model.Product, error)
	List(ctx context.Context) ([]*model.Product, error)
	Create(ctx context.Context, product *model.Product) error

	// ClaimLimitedProduct atomically marks a limited product as ordered by
	// setting last_order_at = now(), but only if it wasn't already claimed
	// within the last hour. It returns the updated product on success, or
	// (nil, nil) if the product is currently within its cooldown window --
	// callers should treat that as "unavailable right now", not an error.
	ClaimLimitedProduct(ctx context.Context, id string) (*model.Product, error)
}
