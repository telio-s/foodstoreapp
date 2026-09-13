package main

import (
	"log"
	"net/http"

	adapterhttp "food-store-apis/internal/adapter/http"
	"food-store-apis/internal/adapter/repository"
	"food-store-apis/internal/domain/service"
	"food-store-apis/internal/infra/config"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewPostgresDB(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, productRepo)

	productHandler := adapterhttp.NewProductHandler(productService)
	orderHandler := adapterhttp.NewOrderHandler(orderService)

	router := adapterhttp.NewRouter(productHandler, orderHandler)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
