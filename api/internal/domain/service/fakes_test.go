package service

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"food-store-apis/internal/domain/model"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type fakeProductRepository struct {
	byID    map[string]*model.Product
	listErr error
}

func newFakeProductRepository(products ...*model.Product) *fakeProductRepository {
	byID := make(map[string]*model.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	return &fakeProductRepository{byID: byID}
}

func (f *fakeProductRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return p, nil
}

func (f *fakeProductRepository) List(ctx context.Context) ([]*model.Product, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	products := make([]*model.Product, 0, len(f.byID))
	for _, p := range f.byID {
		products = append(products, p)
	}
	return products, nil
}

func (f *fakeProductRepository) Create(ctx context.Context, product *model.Product) error {
	f.byID[product.ID] = product
	return nil
}

type fakeOrderRepository struct {
	created   *model.Order
	createErr error
}

func (f *fakeOrderRepository) Create(ctx context.Context, order *model.Order) error {
	if f.createErr != nil {
		return f.createErr
	}
	order.ID = "order-1"
	f.created = order
	return nil
}

func (f *fakeOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	if f.created != nil && f.created.ID == id {
		return f.created, nil
	}
	return nil, errors.New("order not found")
}

var (
	red    = &model.Product{ID: "red", Name: "Red", Price: "50.00"}
	green  = &model.Product{ID: "green", Name: "Green", Price: "40.00"}
	blue   = &model.Product{ID: "blue", Name: "Blue", Price: "30.00"}
	yellow = &model.Product{ID: "yellow", Name: "Yellow", Price: "50.00"}
	pink   = &model.Product{ID: "pink", Name: "Pink", Price: "80.00"}
	purple = &model.Product{ID: "purple", Name: "Purple", Price: "90.00"}
	orange = &model.Product{ID: "orange", Name: "Orange", Price: "120.00"}
)
