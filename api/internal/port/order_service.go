package port

import (
	"context"

	"food-store-apis/internal/domain/model"
)

type CreateOrderItemInput struct {
	ProductID string
	Quantity  int
}

type CreateOrderInput struct {
	MemberCardNumber string
	Items            []CreateOrderItemInput
}

type OrderService interface {
	CreateOrder(ctx context.Context, input CreateOrderInput) (*model.Order, error)
}
