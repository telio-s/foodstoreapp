package handler

import (
	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	products port.ProductService
}

func NewProductHandler(products port.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

// List handles GET /api/v1/products.
func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.products.ListProducts(c.Request.Context())
	if err != nil {
		failure(c, "failed to list products", err)
		return
	}

	items := make([]dto.ProductItem, 0, len(products))
	for _, p := range products {
		items = append(items, dto.ProductItem{
			ID:    p.ID,
			Name:  p.Name,
			Price: p.Price,
		})
	}

	success(c, "success", dto.ListProductsResponse{Items: items})
}
