package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/domain/model"

	"github.com/gin-gonic/gin"
)

func newProductTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	return c, w
}

func TestProductHandler_List_Success(t *testing.T) {
	svc := &fakeProductService{products: []*model.Product{
		{ID: "orange", Name: "Orange", Price: "120.00"},
		{ID: "blue", Name: "Blue", Price: "30.00"},
	}}
	h := NewProductHandler(svc)

	c, w := newProductTestContext()
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("code = %d, want 0", resp.Code)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("failed to re-marshal data: %v", err)
	}
	var payload dto.ListProductsResponse
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if len(payload.Items) != 2 || payload.Items[0].ID != "orange" || payload.Items[1].ID != "blue" {
		t.Errorf("unexpected items: %+v", payload.Items)
	}
}

func TestProductHandler_List_EmptyRepository_ReturnsEmptyItemsSlice(t *testing.T) {
	svc := &fakeProductService{products: nil}
	h := NewProductHandler(svc)

	c, w := newProductTestContext()
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	// aggregateProducts-style handlers must marshal an empty slice ("[]"),
	// not a JSON null, so clients can always range over items.
	if !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Errorf("expected empty items array, got body=%s", w.Body.String())
	}
}

func TestProductHandler_List_ServiceError_ReturnsInternalServerError(t *testing.T) {
	svc := &fakeProductService{err: errors.New("db down")}
	h := NewProductHandler(svc)

	c, w := newProductTestContext()
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
}
