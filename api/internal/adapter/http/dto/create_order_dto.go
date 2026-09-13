package dto

// CreateOrderRequest is the body for POST /api/v1/order.
type CreateOrderRequest struct {
	MemberCardNumber string                   `json:"member_card_number"`
	Items            []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// CreateOrderResponse is returned as-is (no envelope) for POST /api/v1/order.
type CreateOrderResponse struct {
	OrderID        string                   `json:"order_id"`
	TotalPrice     string                   `json:"total_price"`
	DiscountAmount string                   `json:"discount_amount"`
	Products       []CreateOrderProductItem `json:"products"`
}

type CreateOrderProductItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}
