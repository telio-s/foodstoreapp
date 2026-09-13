package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type CreateOrderItemInput struct {
	ProductID int64
	Quantity  int
}

type CreateOrderInput struct {
	MemberCardNumber string
	DiscountAmount   float64
	Items            []CreateOrderItemInput
}

type OrderService interface {
	CreateOrder(ctx context.Context, input CreateOrderInput) (*model.Order, error)
	GetOrder(ctx context.Context, id int64) (*model.Order, error)
}
