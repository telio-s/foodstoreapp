package http

import (
	"net/http"
	"strconv"
)

func NewRouter(products *ProductHandler, orders *OrderHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /products", products.List)

	mux.HandleFunc("POST /orders", orders.Create)
	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid order id", http.StatusBadRequest)
			return
		}
		orders.Get(w, r, id)
	})

	return mux
}
