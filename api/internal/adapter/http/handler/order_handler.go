package handler

import (
	"net/http"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderHandler struct {
	orders   port.OrderService
	validate *validator.Validate
}

func NewOrderHandler(orders port.OrderService, validate *validator.Validate) *OrderHandler {
	return &OrderHandler{orders: orders, validate: validate}
}

// Create handles POST /api/v1/order.
func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if appErr := bindJSON(c, h.validate, &req); appErr != nil {
		failure(c, "", appErr)
		return
	}

	input := port.CreateOrderInput{
		MemberCardNumber: req.MemberCardNumber,
	}
	for _, item := range req.Items {
		input.Items = append(input.Items, port.CreateOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.orders.CreateOrder(c.Request.Context(), input)
	if err != nil {
		failure(c, "failed to create order", err)
		return
	}

	resp := dto.CreateOrderResponse{
		OrderID:        order.ID,
		TotalPrice:     order.TotalPrice,
		DiscountAmount: order.DiscountAmount,
		Products:       aggregateProducts(order.Items),
	}

	c.JSON(http.StatusCreated, resp)
}

// aggregateProducts combines order lines for the same product into a single
// entry, summing quantities, so a product ordered across multiple lines
// (e.g. two separate "Orange" entries in the request) appears once in the
// response with its total quantity.
func aggregateProducts(items []*model.OrderItem) []dto.CreateOrderProductItem {
	products := make([]dto.CreateOrderProductItem, 0, len(items))
	indexByProductID := make(map[string]int, len(items))

	for _, item := range items {
		if idx, ok := indexByProductID[item.ProductID]; ok {
			products[idx].Quantity += item.Quantity
			continue
		}
		indexByProductID[item.ProductID] = len(products)
		products = append(products, dto.CreateOrderProductItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	return products
}
