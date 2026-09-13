package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"food-store-apis/internal/domain/apperror"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

// orderService keeps created orders in memory while the database is not
// wired up yet.
type orderService struct {
	mu     sync.Mutex
	orders map[string]*model.Order
	nextID int64
}

func NewOrderService() port.OrderService {
	return &orderService{orders: make(map[string]*model.Order)}
}

func (s *orderService) CreateOrder(ctx context.Context, input port.CreateOrderInput) (*model.Order, error) {
	if len(input.Items) == 0 {
		return nil, apperror.Invalid("order must contain at least one item")
	}

	order := &model.Order{
		MemberCardNumber: input.MemberCardNumber,
		DiscountAmount:   "0.00",
		CreatedAt:        time.Now(),
	}

	var total float64
	for _, item := range input.Items {
		product, ok := findMockProduct(item.ProductID)
		if !ok {
			return nil, apperror.NotFound("product not found")
		}

		price, err := strconv.ParseFloat(product.Price, 64)
		if err != nil {
			return nil, apperror.Internal(fmt.Errorf("invalid product price %q: %w", product.Price, err))
		}

		order.Items = append(order.Items, &model.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   product.Price,
		})
		total += price * float64(item.Quantity)
	}
	order.TotalPrice = strconv.FormatFloat(total, 'f', 2, 64)

	s.mu.Lock()
	s.nextID++
	order.ID = strconv.FormatInt(s.nextID, 10)
	for _, item := range order.Items {
		item.OrderID = order.ID
	}
	s.orders[order.ID] = order
	s.mu.Unlock()

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	if !ok {
		return nil, apperror.NotFound("order not found")
	}
	return order, nil
}
