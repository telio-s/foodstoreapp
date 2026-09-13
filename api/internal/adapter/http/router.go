package http

import (
	"food-store-apis/internal/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter(products *handler.ProductHandler, orders *handler.OrderHandler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/products", products.List)
		v1.POST("/order", orders.Create)
	}

	return r
}
