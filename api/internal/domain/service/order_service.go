package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
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
	uow      port.UnitOfWork
	logger   *slog.Logger
}

func NewOrderService(products port.ProductRepository, uow port.UnitOfWork, logger *slog.Logger) port.OrderService {
	return &orderService{
		products: products,
		uow:      uow,
		logger:   logger.With("component", "order_service"),
	}
}

// CreateOrder loads/validates products and computes pricing up front, ahead
// of any database transaction -- none of that depends on the limited-product
// race this rule cares about. Only the part that actually needs atomicity
// against a concurrent order -- claiming each limited product and creating
// the order -- runs inside port.UnitOfWork, so the transaction stays as
// short as possible and holds row locks for as little time as possible.
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

	subtotal, limitedProductIDs, err := loadOrderItems(ctx, s.products, input.Items, order)
	if err != nil {
		s.logger.ErrorContext(ctx, "create order: failed to load order items", "error", err)
		return nil, toOrderError(err)
	}

	discount, err := calculateDiscount(order.Items, order.MemberCardNumber, subtotal)
	if err != nil {
		s.logger.ErrorContext(ctx, "create order: failed to calculate discount", "error", err)
		return nil, apperror.Internal(err)
	}
	order.DiscountAmount = strconv.FormatFloat(discount, 'f', 2, 64)
	order.TotalPrice = strconv.FormatFloat(subtotal-discount, 'f', 2, 64)

	// Sorted by ID so two orders that both touch the same set of limited
	// products always try to lock their rows in the same order, avoiding a
	// deadlock between them.
	sort.Strings(limitedProductIDs)

	// BEGIN: everything inside this closure runs inside one database
	// transaction (see postgres.unitOfWork.Execute, which issues the actual
	// BEGIN/COMMIT/ROLLBACK). Returning an error here rolls the whole thing
	// back -- including any limited-product claim already made earlier in
	// this same order -- returning nil commits it.
	err = s.uow.Execute(ctx, func(repos port.TxRepositories) error {
		for _, productID := range limitedProductIDs {
			claimed, err := repos.Products.ClaimLimitedProduct(ctx, productID)
			if err != nil {
				return fmt.Errorf("claim limited product %s: %w", productID, err)
			}
			if claimed == nil {
				s.logger.WarnContext(ctx, "create order: limited product already claimed within the window",
					"product_id", productID,
				)
				return apperror.ErrProductLimited
			}
		}

		if err := repos.Orders.Create(ctx, order); err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "create order: failed to persist order",
			"member_card_number", order.MemberCardNumber,
			"error", err,
		)
		return nil, toOrderError(err)
	}

	return order, nil
}

// toOrderError maps an error from the create-order flow to the AppError
// callers expect: an existing AppError (ErrProductNotFound, ErrProductLimited,
// ...) passes through unchanged; anything else becomes apperror.Internal.
func toOrderError(err error) *apperror.AppError {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.Internal(err)
}

// loadOrderItems fetches the product behind each requested line item,
// appends the corresponding model.OrderItem to order, and returns the
// pre-discount subtotal plus the deduplicated IDs of every limited product
// involved (still needing ClaimLimitedProduct before the order can proceed).
func loadOrderItems(ctx context.Context, products port.ProductRepository, items []port.CreateOrderItemInput, order *model.Order) (float64, []string, error) {
	productByID := make(map[string]*model.Product, len(items))
	limited := make(map[string]bool)

	var subtotal float64
	for _, item := range items {
		product, ok := productByID[item.ProductID]
		if !ok {
			var err error
			product, err = products.GetByID(ctx, item.ProductID)
			if err != nil {
				return 0, nil, apperror.ErrProductNotFound
			}
			productByID[item.ProductID] = product
			if product.IsLimited {
				limited[product.ID] = true
			}
		}

		price, err := strconv.ParseFloat(product.Price, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid product price %q: %w", product.Price, err)
		}

		order.Items = append(order.Items, &model.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   product.Price,
		})
		subtotal += price * float64(item.Quantity)
	}

	limitedProductIDs := make([]string, 0, len(limited))
	for id := range limited {
		limitedProductIDs = append(limitedProductIDs, id)
	}
	return subtotal, limitedProductIDs, nil
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
		slog.Debug("calculateDiscount: pair discount", "product", name, "pairs", pairs)
		discount += float64(pairs) * 2 * priceByProduct[name] * pairDiscountRate
	}

	if memberCardNumber != "" {
		discount += subtotal * memberDiscountRate
	}

	return discount, nil
}
