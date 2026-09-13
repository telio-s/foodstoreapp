package http

import (
	"net/http"

	"food-store-apis/internal/port"
)

type ProductHandler struct {
	products port.ProductService
}

func NewProductHandler(products port.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.products.ListProducts(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, products)
}
