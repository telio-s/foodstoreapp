package service

import (
	"context"

	"food-store-apis/internal/domain/apperror"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

type orderService struct {
	orders   port.OrderRepository
	products port.ProductRepository
}

func NewOrderService(orders port.OrderRepository, products port.ProductRepository) port.OrderService {
	return &orderService{orders: orders, products: products}
}

func (s *orderService) CreateOrder(ctx context.Context, input port.CreateOrderInput) (*model.Order, error) {
	if len(input.Items) == 0 {
		return nil, apperror.Invalid("order must contain at least one item")
	}

	order := &model.Order{
		MemberCardNumber: input.MemberCardNumber,
		DiscountAmount:   input.DiscountAmount,
	}

	var total float64
	for _, item := range input.Items {
		product, err := s.products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, apperror.NotFound("product not found")
		}

		order.Items = append(order.Items, &model.OrderItem{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			UnitPrice: product.Price,
		})
		total += product.Price * float64(item.Quantity)
	}
	order.TotalPrice = total - order.DiscountAmount

	if err := s.orders.Create(ctx, order); err != nil {
		return nil, apperror.Internal(err)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("order not found")
	}
	return order, nil
}
