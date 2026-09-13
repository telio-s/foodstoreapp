package main

import (
	"context"
	"log"
	"net/http"

	adapterhttp "food-store-apis/internal/adapter/http"
	"food-store-apis/internal/adapter/http/handler"
	"food-store-apis/internal/adapter/postgres"
	"food-store-apis/internal/docs"
	"food-store-apis/internal/domain/service"
	"food-store-apis/internal/infra/config"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			providePostgresPool,
			postgres.NewProductRepository,
			service.NewProductService,
			service.NewOrderService,
			handler.NewProductHandler,
			handler.NewOrderHandler,
			adapterhttp.NewRouter,
		),
		fx.Invoke(registerHTTPServer),
	).Run()
}

func providePostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func registerHTTPServer(lc fx.Lifecycle, cfg *config.Config, router *gin.Engine, pool *pgxpool.Pool) {
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
			pool.Close()
			return srv.Shutdown(ctx)
		},
	})
}
