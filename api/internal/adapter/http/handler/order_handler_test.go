package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newOrderTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/order", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestOrderHandler_Create_Success(t *testing.T) {
	svc := &fakeOrderService{order: &model.Order{
		ID:             "order-1",
		TotalPrice:     "228.00",
		DiscountAmount: "12.00",
		Items: []*model.OrderItem{
			{ProductID: "orange", ProductName: "Orange", Quantity: 1},
			{ProductID: "orange", ProductName: "Orange", Quantity: 1},
			{ProductID: "blue", ProductName: "Blue", Quantity: 1},
		},
	}}
	h := NewOrderHandler(svc, newTestValidator())

	c, w := newOrderTestContext(`{"member_card_number":"","items":[{"product_id":"orange","quantity":2},{"product_id":"blue","quantity":1}]}`)
	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	if !svc.called {
		t.Fatalf("expected order service to be called")
	}

	var resp dto.CreateOrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.OrderID != "order-1" || resp.TotalPrice != "228.00" || resp.DiscountAmount != "12.00" {
		t.Errorf("unexpected response: %+v", resp)
	}

	// The two "orange" order lines returned by the service must be merged
	// into a single product entry with the summed quantity.
	if len(resp.Products) != 2 {
		t.Fatalf("got %d products, want 2 (duplicate orange lines merged), products=%+v", len(resp.Products), resp.Products)
	}
	if resp.Products[0].ProductID != "orange" || resp.Products[0].Quantity != 2 {
		t.Errorf("orange line not aggregated: %+v", resp.Products[0])
	}
	if resp.Products[1].ProductID != "blue" || resp.Products[1].Quantity != 1 {
		t.Errorf("unexpected blue line: %+v", resp.Products[1])
	}
}

func TestOrderHandler_Create_ValidationErrors_ReturnBadRequestWithoutCallingService(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"malformed json", `{`},
		{"missing items", `{"member_card_number":""}`},
		{"empty items", `{"items":[]}`},
		{"missing product_id", `{"items":[{"quantity":1}]}`},
		{"quantity below minimum", `{"items":[{"product_id":"blue","quantity":0}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeOrderService{}
			h := NewOrderHandler(svc, newTestValidator())

			c, w := newOrderTestContext(tt.body)
			h.Create(c)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
			}
			if svc.called {
				t.Errorf("order service should not be called when validation fails")
			}
		})
	}
}

func TestOrderHandler_Create_ServiceErrorMapping(t *testing.T) {
	validBody := `{"items":[{"product_id":"blue","quantity":1}]}`

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", port.NotFound("product not found"), http.StatusNotFound},
		{"invalid", port.Invalid("bad input"), http.StatusBadRequest},
		{"internal apperror", port.Internal(errors.New("db down")), http.StatusInternalServerError},
		{"generic non-apperror", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeOrderService{err: tt.err}
			h := NewOrderHandler(svc, newTestValidator())

			c, w := newOrderTestContext(validBody)
			h.Create(c)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
