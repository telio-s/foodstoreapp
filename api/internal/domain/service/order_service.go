package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"food-store-apis/internal/domain/apperror"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"
)

const (
	pairDiscountRate   = 0.05
	memberDiscountRate = 0.10
)

// pairDiscountProducts lists the product names eligible for the pair
// discount (business rule 1 in CLAUDE.md).
var pairDiscountProducts = map[string]bool{
	"Orange": true,
	"Pink":   true,
	"Green":  true,
}

type orderService struct {
	products port.ProductRepository
	orders   port.OrderRepository
	logger   *slog.Logger
}

func NewOrderService(products port.ProductRepository, orders port.OrderRepository, logger *slog.Logger) port.OrderService {
	return &orderService{
		products: products,
		orders:   orders,
		logger:   logger.With("component", "order_service"),
	}
}

func (s *orderService) CreateOrder(ctx context.Context, input port.CreateOrderInput) (*model.Order, error) {
	s.logger.InfoContext(ctx, "create order: input",
		"member_card_number", input.MemberCardNumber,
		"items", input.Items,
	)

	if len(input.Items) == 0 {
		return nil, apperror.ErrOrderEmptyItems
	}

	order := &model.Order{
		MemberCardNumber: input.MemberCardNumber,
	}

	var subtotal float64
	for _, item := range input.Items {
		product, err := s.products.GetByID(ctx, item.ProductID)
		if err != nil {
			s.logger.ErrorContext(ctx, "create order: product not found",
				"product_id", item.ProductID,
				"error", err,
			)
			return nil, apperror.ErrProductNotFound
		}

		price, err := strconv.ParseFloat(product.Price, 64)
		if err != nil {
			s.logger.ErrorContext(ctx, "create order: invalid product price",
				"product_id", product.ID,
				"price", product.Price,
				"error", err,
			)
			return nil, apperror.Internal(fmt.Errorf("invalid product price %q: %w", product.Price, err))
		}

		order.Items = append(order.Items, &model.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   product.Price,
		})
		subtotal += price * float64(item.Quantity)
	}

	discount, err := calculateDiscount(order.Items, order.MemberCardNumber, subtotal)
	if err != nil {
		s.logger.ErrorContext(ctx, "create order: failed to calculate discount",
			"member_card_number", order.MemberCardNumber,
			"subtotal", subtotal,
			"error", err,
		)
		return nil, apperror.Internal(err)
	}

	order.DiscountAmount = strconv.FormatFloat(discount, 'f', 2, 64)
	order.TotalPrice = strconv.FormatFloat(subtotal-discount, 'f', 2, 64)

	if err := s.orders.Create(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "create order: failed to persist order",
			"member_card_number", order.MemberCardNumber,
			"total_price", order.TotalPrice,
			"discount_amount", order.DiscountAmount,
			"error", err,
		)
		return nil, apperror.Internal(err)
	}

	return order, nil
}

// calculateDiscount applies the two order-level discount rules described in
// CLAUDE.md. Both are computed independently against the pre-discount
// subtotal and summed -- they do not compound.
//
//  1. Pair discount: every complete pair of the same discount-eligible
//     product (Orange, Pink, Green) gets 5% off that pair's subtotal
//     (2 x unit_price x 5%). A leftover odd unit is charged at full price.
//  2. Member discount: an additional 10% off the order subtotal when a
//     member card number is present.
func calculateDiscount(items []*model.OrderItem, memberCardNumber string, subtotal float64) (float64, error) {
	quantityByProduct := make(map[string]int)
	priceByProduct := make(map[string]float64)

	for _, item := range items {
		if !pairDiscountProducts[item.ProductName] {
			continue
		}

		price, err := strconv.ParseFloat(item.UnitPrice, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid unit price %q: %w", item.UnitPrice, err)
		}

		quantityByProduct[item.ProductName] += item.Quantity
		priceByProduct[item.ProductName] = price
	}

	var discount float64
	for name, qty := range quantityByProduct {
		pairs := qty / 2
		discount += float64(pairs) * 2 * priceByProduct[name] * pairDiscountRate
	}

	if memberCardNumber != "" {
		discount += subtotal * memberDiscountRate
	}

	return discount, nil
}
