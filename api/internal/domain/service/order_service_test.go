package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"food-store-apis/internal/domain/apperror"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

func TestCreateOrder_NoItems_ReturnsInvalid(t *testing.T) {
	products := newFakeProductRepository()
	svc := NewOrderService(products, newFakeUnitOfWork(&fakeOrderRepository{}, products), testLogger())

	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{})

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalid {
		t.Fatalf("expected CodeInvalid error, got %v", err)
	}
}

func TestCreateOrder_ProductNotFound_ReturnsNotFound(t *testing.T) {
	products := newFakeProductRepository()
	svc := NewOrderService(products, newFakeUnitOfWork(&fakeOrderRepository{}, products), testLogger())

	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{{ProductID: "missing", Quantity: 1}},
	})

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound error, got %v", err)
	}
}

func TestCreateOrder_PersistsAndReturnsComputedOrder(t *testing.T) {
	products := newFakeProductRepository(blue)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	order, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		MemberCardNumber: "",
		Items:            []port.CreateOrderItemInput{{ProductID: "blue", Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.TotalPrice != "60.00" {
		t.Errorf("total price = %q, want 60.00", order.TotalPrice)
	}
	if order.DiscountAmount != "0.00" {
		t.Errorf("discount amount = %q, want 0.00", order.DiscountAmount)
	}
	if orders.created != order {
		t.Errorf("order was not persisted via the repository")
	}
	if len(order.Items) != 1 || order.Items[0].ProductName != "Blue" {
		t.Errorf("unexpected order items: %+v", order.Items)
	}
}

func TestCreateOrder_AppliesPairAndMemberDiscounts(t *testing.T) {
	products := newFakeProductRepository(orange, pink, green, blue)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	order, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		MemberCardNumber: "MC-1001",
		Items: []port.CreateOrderItemInput{
			{ProductID: "orange", Quantity: 2}, // (120+120) - 5%          = 228
			{ProductID: "pink", Quantity: 4},   // (80+80-5%)+(80+80-5%)   = 304
			{ProductID: "green", Quantity: 3},  // (40+40-5%)+40           = 116
			{ProductID: "blue", Quantity: 1},   // no rule applies         = 30
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// subtotal    = 240 + 320 + 120 + 30 = 710
	// pairDiscount = 12 (orange) + 16 (pink) + 4 (green)  = 32
	// memberDiscount = 10% * 710 = 71.00
	// total discount = 32 + 71.00 = 103.00
	if order.DiscountAmount != "103.00" {
		t.Errorf("discount amount = %q, want 103.00", order.DiscountAmount)
	}
	if order.TotalPrice != "607.00" {
		t.Errorf("total price = %q, want 607.00", order.TotalPrice)
	}
}

func TestCreateOrder_RepositoryError_ReturnsInternal(t *testing.T) {
	products := newFakeProductRepository(blue)
	orders := &fakeOrderRepository{createErr: errors.New("db down")}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{{ProductID: "blue", Quantity: 1}},
	})

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("expected CodeInternal error, got %v", err)
	}
}

func TestCreateOrder_LimitedProductOrderedWithinWindow_ReturnsInvalid(t *testing.T) {
	recentlyOrdered := time.Now().Add(-30 * time.Minute)
	limitedRed := &model.Product{ID: "red", Name: "Red", Price: "50.00", IsLimited: true, LastOrderAt: &recentlyOrdered}

	products := newFakeProductRepository(limitedRed)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{{ProductID: "red", Quantity: 1}},
	})

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalid || appErr.ErrorCode != apperror.ErrProductLimited.ErrorCode {
		t.Fatalf("expected ErrProductLimited, got %v", err)
	}
	if orders.created != nil {
		t.Errorf("order should not have been persisted")
	}
}

func TestCreateOrder_LimitedProductOrderedOutsideWindow_Allowed(t *testing.T) {
	orderedLongAgo := time.Now().Add(-2 * time.Hour)
	limitedRed := &model.Product{ID: "red", Name: "Red", Price: "50.00", IsLimited: true, LastOrderAt: &orderedLongAgo}

	products := newFakeProductRepository(limitedRed)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	order, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{{ProductID: "red", Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orders.created != order {
		t.Errorf("order was not persisted via the repository")
	}
}

func TestCreateOrder_LimitedProductNeverOrdered_Allowed(t *testing.T) {
	neverOrderedRed := &model.Product{ID: "red", Name: "Red", Price: "50.00", IsLimited: true}

	products := newFakeProductRepository(neverOrderedRed)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{{ProductID: "red", Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateOrder_SuccessfulOrder_UpdatesLastOrderAtForLimitedProducts(t *testing.T) {
	orderedLongAgo := time.Now().Add(-2 * time.Hour)
	limitedRed := &model.Product{ID: "red", Name: "Red", Price: "50.00", IsLimited: true, LastOrderAt: &orderedLongAgo}

	products := newFakeProductRepository(limitedRed, blue)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	before := time.Now()
	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{
			{ProductID: "red", Quantity: 1},
			{ProductID: "blue", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claimedAt, ok := products.claimedAt["red"]
	if !ok {
		t.Fatalf("expected ClaimLimitedProduct to be called for the limited product")
	}
	if claimedAt.Before(before) {
		t.Errorf("last_order_at = %v, want a time at or after %v", claimedAt, before)
	}
	if _, ok := products.claimedAt["blue"]; ok {
		t.Errorf("non-limited product should not be claimed")
	}
}

func TestCreateOrder_MultipleLimitedProducts_ClaimedInSortedIDOrder(t *testing.T) {
	orderedLongAgo := time.Now().Add(-2 * time.Hour)
	limitedB := &model.Product{ID: "b-limited", Name: "B", Price: "10.00", IsLimited: true, LastOrderAt: &orderedLongAgo}
	limitedA := &model.Product{ID: "a-limited", Name: "A", Price: "10.00", IsLimited: true, LastOrderAt: &orderedLongAgo}

	products := newFakeProductRepository(limitedB, limitedA)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	// Requested in "b then a" order; claims should still happen sorted by
	// ID ("a-limited" before "b-limited") to keep lock order consistent
	// across concurrent orders touching the same products.
	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{
			{ProductID: "b-limited", Quantity: 1},
			{ProductID: "a-limited", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := products.claimOrder, []string{"a-limited", "b-limited"}; !reflect.DeepEqual(got, want) {
		t.Errorf("claim order = %v, want %v", got, want)
	}
}

func TestCreateOrder_ClaimFails_OrderNotPersisted(t *testing.T) {
	recentlyOrdered := time.Now().Add(-30 * time.Minute)
	limitedRed := &model.Product{ID: "red", Name: "Red", Price: "50.00", IsLimited: true, LastOrderAt: &recentlyOrdered}

	products := newFakeProductRepository(limitedRed, blue)
	orders := &fakeOrderRepository{}
	svc := NewOrderService(products, newFakeUnitOfWork(orders, products), testLogger())

	_, err := svc.CreateOrder(context.Background(), port.CreateOrderInput{
		Items: []port.CreateOrderItemInput{
			{ProductID: "red", Quantity: 1},
			{ProductID: "blue", Quantity: 1},
		},
	})

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.ErrorCode != apperror.ErrProductLimited.ErrorCode {
		t.Fatalf("expected ErrProductLimited, got %v", err)
	}
	if orders.created != nil {
		t.Errorf("order should not have been persisted when a limited-product claim fails")
	}
}

func TestCalculateDiscount(t *testing.T) {
	tests := []struct {
		name             string
		items            []*model.OrderItem
		memberCardNumber string
		subtotal         float64
		want             float64
	}{
		{
			name: "orange pair gets 5% off the pair",
			items: []*model.OrderItem{
				{ProductName: "Orange", Quantity: 2, UnitPrice: "120.00"},
			},
			subtotal: 240,
			want:     12, // (120+120) * 5%
		},
		{
			name: "pink two pairs each get 5% off",
			items: []*model.OrderItem{
				{ProductName: "Pink", Quantity: 4, UnitPrice: "80.00"},
			},
			subtotal: 320,
			want:     16, // 2 pairs * (80+80)*5%
		},
		{
			name: "green odd quantity: one pair discounted, one at full price",
			items: []*model.OrderItem{
				{ProductName: "Green", Quantity: 3, UnitPrice: "40.00"},
			},
			subtotal: 120,
			want:     4, // 1 pair * (40+40)*5%
		},
		{
			name: "single eligible item has no pair, no discount",
			items: []*model.OrderItem{
				{ProductName: "Orange", Quantity: 1, UnitPrice: "120.00"},
			},
			subtotal: 120,
			want:     0,
		},
		{
			name: "non-eligible product never discounted",
			items: []*model.OrderItem{
				{ProductName: "Blue", Quantity: 4, UnitPrice: "30.00"},
			},
			subtotal: 120,
			want:     0,
		},
		{
			name: "member card alone gives 10% off the subtotal",
			items: []*model.OrderItem{
				{ProductName: "Blue", Quantity: 2, UnitPrice: "30.00"},
			},
			memberCardNumber: "MC-1001",
			subtotal:         60,
			want:             6,
		},
		{
			name: "pair discount and member discount stack additively",
			items: []*model.OrderItem{
				{ProductName: "Orange", Quantity: 2, UnitPrice: "120.00"},
			},
			memberCardNumber: "MC-1001",
			subtotal:         240,
			want:             12 + 24, // pair 5% + member 10% of subtotal
		},
		{
			name: "multiple eligible products combine independently",
			items: []*model.OrderItem{
				{ProductName: "Orange", Quantity: 2, UnitPrice: "120.00"},
				{ProductName: "Pink", Quantity: 4, UnitPrice: "80.00"},
				{ProductName: "Green", Quantity: 3, UnitPrice: "40.00"},
			},
			subtotal: 240 + 320 + 120,
			want:     12 + 16 + 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateDiscount(tt.items, tt.memberCardNumber, tt.subtotal)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := got - tt.want; diff < -0.0001 || diff > 0.0001 {
				t.Errorf("calculateDiscount() = %v, want %v", got, tt.want)
			}
		})
	}
}
