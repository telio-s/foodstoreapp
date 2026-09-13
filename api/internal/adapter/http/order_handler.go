package http

import (
	"encoding/json"
	"net/http"

	"food-store-apis/internal/port"
)

type OrderHandler struct {
	orders port.OrderService
}

func NewOrderHandler(orders port.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

type createOrderRequest struct {
	MemberCardNumber string  `json:"member_card_number"`
	DiscountAmount   float64 `json:"discount_amount"`
	Items            []struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	} `json:"items"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := port.CreateOrderInput{
		MemberCardNumber: req.MemberCardNumber,
		DiscountAmount:   req.DiscountAmount,
	}
	for _, item := range req.Items {
		input.Items = append(input.Items, port.CreateOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.orders.CreateOrder(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request, id int64) {
	order, err := h.orders.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}
