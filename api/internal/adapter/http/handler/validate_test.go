package handler

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
)

func newBindContext(body string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/api/v1/order", bytes.NewBufferString(body))
	return c
}

func TestBindJSON_ValidBody_ReturnsNoError(t *testing.T) {
	v := newTestValidator()
	c := newBindContext(`{"member_card_number":"MC-1","items":[{"product_id":"blue","quantity":1}]}`)

	var out dto.CreateOrderRequest
	if appErr := bindJSON(c, v, &out); appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if out.MemberCardNumber != "MC-1" || len(out.Items) != 1 || out.Items[0].ProductID != "blue" {
		t.Errorf("unexpected decoded request: %+v", out)
	}
}

func TestBindJSON_MalformedBody_ReturnsInvalid(t *testing.T) {
	v := newTestValidator()
	c := newBindContext(`{`)

	var out dto.CreateOrderRequest
	appErr := bindJSON(c, v, &out)
	if appErr == nil || appErr.Code != port.CodeInvalid {
		t.Fatalf("expected CodeInvalid, got %v", appErr)
	}
}

func TestBindJSON_ValidationFailures_ReportPerFieldMessages(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"missing items key", `{}`, "items is required"},
		{"empty items slice", `{"items":[]}`, "items must be at least 1"},
		{"missing product_id", `{"items":[{"quantity":1}]}`, "product_id is required"},
		{"zero quantity", `{"items":[{"product_id":"blue","quantity":0}]}`, "quantity is required"},
		{"negative quantity", `{"items":[{"product_id":"blue","quantity":-1}]}`, "quantity must be at least 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newTestValidator()
			c := newBindContext(tt.body)

			var out dto.CreateOrderRequest
			appErr := bindJSON(c, v, &out)
			if appErr == nil {
				t.Fatalf("expected a validation error, got none")
			}
			if appErr.ErrorCode != -9000 {
				t.Errorf("ErrorCode = %d, want -9000 (ErrValidation)", appErr.ErrorCode)
			}

			found := false
			for _, msg := range appErr.Errors {
				if msg == tt.wantMsg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("errors = %v, want to contain %q", appErr.Errors, tt.wantMsg)
			}
		})
	}
}
