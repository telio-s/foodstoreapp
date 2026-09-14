package postgres

import (
	"context"
	"errors"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productRepository struct {
	q *Queries
}

func NewProductRepository(pool *pgxpool.Pool) port.ProductRepository {
	return &productRepository{q: New(pool)}
}

func (r *productRepository) List(ctx context.Context) ([]*model.Product, error) {
	rows, err := r.q.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*model.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, toModelProduct(row))
	}
	return products, nil
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	row, err := r.q.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toModelProduct(row), nil
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	row, err := r.q.CreateProduct(ctx, CreateProductParams{
		Name:  product.Name,
		Price: product.Price,
	})
	if err != nil {
		return err
	}

	product.ID = row.ID
	return nil
}

func (r *productRepository) ClaimLimitedProduct(ctx context.Context, id string) (*model.Product, error) {
	row, err := r.q.ClaimLimitedProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toModelProduct(row), nil
}

func toModelProduct(p Product) *model.Product {
	product := &model.Product{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		IsLimited: p.IsLimited,
	}
	if p.LastOrderAt.Valid {
		product.LastOrderAt = &p.LastOrderAt.Time
	}
	return product
}
