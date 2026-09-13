package repository

import (
	"context"
	"database/sql"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) port.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	var p model.Product

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, price FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Price)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *productRepository) List(ctx context.Context) ([]*model.Product, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, price FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	return products, rows.Err()
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO products (name, price) VALUES ($1, $2) RETURNING id`,
		product.Name, product.Price,
	).Scan(&product.ID)
}
