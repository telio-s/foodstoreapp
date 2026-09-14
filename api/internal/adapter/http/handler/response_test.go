package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
)

func TestSuccess_WritesOKEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	success(c, "ok", map[string]string{"foo": "bar"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != 0 || resp.Msg != "ok" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestFailure_MapsAppErrorCodesToHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{"not found", port.NotFound("missing"), http.StatusNotFound, 1},
		{"invalid", port.Invalid("bad"), http.StatusBadRequest, 1},
		{"internal apperror", port.Internal(errors.New("db down")), http.StatusInternalServerError, -9999},
		{"validation error", port.ErrValidation.WithErrors([]string{"product_id is required"}), http.StatusBadRequest, -9000},
		{"plain non-apperror", errors.New("boom"), http.StatusInternalServerError, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			failure(c, "", tt.err)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
			var resp dto.Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", resp.Code, tt.wantCode)
			}
		})
	}
}

func TestFailure_EmptyMsgFallsBackToAppErrorMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	failure(c, "", port.NotFound("product not found"))

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Msg != "product not found" {
		t.Errorf("msg = %q, want %q", resp.Msg, "product not found")
	}
}

func TestFailure_ExplicitMsgOverridesAppErrorMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	failure(c, "custom message", port.NotFound("product not found"))

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Msg != "custom message" {
		t.Errorf("msg = %q, want %q", resp.Msg, "custom message")
	}
}
