package handler

import (
	"net/http"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/domain/apperror"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orders port.OrderService
}

func NewOrderHandler(orders port.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

// Create handles POST /api/v1/order.
func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failure(c, "invalid request body", apperror.Invalid(err.Error()))
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
		Products:       make([]dto.CreateOrderProductItem, 0, len(order.Items)),
	}
	for _, item := range order.Items {
		resp.Products = append(resp.Products, dto.CreateOrderProductItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	c.JSON(http.StatusCreated, resp)
}
