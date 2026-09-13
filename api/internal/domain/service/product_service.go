package service

import (
	"context"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

// mockProducts is a temporary in-memory catalog used while the database is
// not wired up yet.
var mockProducts = []*model.Product{
	{ID: "1", Name: "Cheeseburger", Price: "5.99"},
	{ID: "2", Name: "Fries", Price: "2.49"},
	{ID: "3", Name: "Soda", Price: "1.99"},
}

func findMockProduct(id string) (*model.Product, bool) {
	for _, p := range mockProducts {
		if p.ID == id {
			return p, true
		}
	}
	return nil, false
}

type productService struct{}

func NewProductService() port.ProductService {
	return &productService{}
}

func (s *productService) ListProducts(ctx context.Context) ([]*model.Product, error) {
	return mockProducts, nil
}
