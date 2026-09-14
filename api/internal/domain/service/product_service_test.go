package service

import (
	"context"
	"errors"
	"testing"

	"food-store-apis/internal/domain/apperror"
)

func TestListProducts_ReturnsProductsFromRepository(t *testing.T) {
	repo := newFakeProductRepository(orange, pink, green)
	svc := NewProductService(repo)

	products, err := svc.ListProducts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(products) != 3 {
		t.Fatalf("got %d products, want 3", len(products))
	}

	byID := make(map[string]bool, len(products))
	for _, p := range products {
		byID[p.ID] = true
	}
	for _, want := range []string{orange.ID, pink.ID, green.ID} {
		if !byID[want] {
			t.Errorf("missing product %q in result", want)
		}
	}
}

func TestListProducts_EmptyRepository_ReturnsEmptySlice(t *testing.T) {
	repo := newFakeProductRepository()
	svc := NewProductService(repo)

	products, err := svc.ListProducts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("got %d products, want 0", len(products))
	}
}

func TestListProducts_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := newFakeProductRepository()
	repo.listErr = errors.New("db down")
	svc := NewProductService(repo)

	_, err := svc.ListProducts(context.Background())

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("expected CodeInternal error, got %v", err)
	}
}
