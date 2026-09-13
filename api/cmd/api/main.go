package main

import (
	"context"
	"log"
	"net/http"

	adapterhttp "food-store-apis/internal/adapter/http"
	"food-store-apis/internal/adapter/http/handler"
	"food-store-apis/internal/docs"
	"food-store-apis/internal/domain/service"
	"food-store-apis/internal/infra/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			service.NewProductService,
			service.NewOrderService,
			handler.NewProductHandler,
			handler.NewOrderHandler,
			adapterhttp.NewRouter,
		),
		fx.Invoke(registerHTTPServer),
	).Run()
}

func registerHTTPServer(lc fx.Lifecycle, cfg *config.Config, router *gin.Engine) {
	docs.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Printf("listening on :%s", cfg.Port)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
