package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type fakeProductRepository struct {
	byID     map[string]*model.Product
	listErr  error
	claimErr error

	claimedAt  map[string]time.Time
	claimOrder []string
}

func newFakeProductRepository(products ...*model.Product) *fakeProductRepository {
	byID := make(map[string]*model.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	return &fakeProductRepository{byID: byID, claimedAt: make(map[string]time.Time)}
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

// claimWindow mirrors the 1-hour cooldown baked into the ClaimLimitedProduct
// SQL query, so this fake's behavior matches Postgres's for these tests.
const claimWindow = time.Hour

// ClaimLimitedProduct mimics the atomic conditional UPDATE the real query
// performs: it only "claims" (advances LastOrderAt to now and returns the
// product) when the product is limited and outside its cooldown window;
// otherwise it returns (nil, nil), exactly like a zero-row UPDATE.
func (f *fakeProductRepository) ClaimLimitedProduct(ctx context.Context, id string) (*model.Product, error) {
	if f.claimErr != nil {
		return nil, f.claimErr
	}

	f.claimOrder = append(f.claimOrder, id)

	p, ok := f.byID[id]
	if !ok || !p.IsLimited {
		return nil, nil
	}
	if p.LastOrderAt != nil && time.Since(*p.LastOrderAt) < claimWindow {
		return nil, nil
	}

	now := time.Now()
	p.LastOrderAt = &now
	f.claimedAt[id] = now
	return p, nil
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

// fakeUnitOfWork runs fn directly against the given fakes instead of a real
// transaction -- sufficient for unit tests, which don't exercise rollback
// behavior (that belongs to an integration test against real Postgres).
type fakeUnitOfWork struct {
	orders   *fakeOrderRepository
	products *fakeProductRepository
}

func newFakeUnitOfWork(orders *fakeOrderRepository, products *fakeProductRepository) *fakeUnitOfWork {
	return &fakeUnitOfWork{orders: orders, products: products}
}

func (u *fakeUnitOfWork) Execute(ctx context.Context, fn func(repos port.TxRepositories) error) error {
	return fn(port.TxRepositories{Orders: u.orders, Products: u.products})
}

var (
	red    = &model.Product{ID: "red", Name: "Red", Price: "50.00", IsLimited: true}
	green  = &model.Product{ID: "green", Name: "Green", Price: "40.00"}
	blue   = &model.Product{ID: "blue", Name: "Blue", Price: "30.00"}
	yellow = &model.Product{ID: "yellow", Name: "Yellow", Price: "50.00"}
	pink   = &model.Product{ID: "pink", Name: "Pink", Price: "80.00"}
	purple = &model.Product{ID: "purple", Name: "Purple", Price: "90.00"}
	orange = &model.Product{ID: "orange", Name: "Orange", Price: "120.00"}
)
