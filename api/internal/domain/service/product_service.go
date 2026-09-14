package service

import (
	"context"

	"food-store-apis/internal/domain/apperror"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

type productService struct {
	products port.ProductRepository
}

func NewProductService(products port.ProductRepository) port.ProductService {
	return &productService{products: products}
}

func (s *productService) ListProducts(ctx context.Context) ([]*model.Product, error) {
	products, err := s.products.List(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return products, nil
}
