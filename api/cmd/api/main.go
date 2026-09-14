package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"

	adapterhttp "food-store-apis/internal/adapter/http"
	"food-store-apis/internal/adapter/http/handler"
	"food-store-apis/internal/adapter/postgres"
	"food-store-apis/internal/docs"
	"food-store-apis/internal/domain/service"
	"food-store-apis/internal/infra/config"

	charmlog "github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			NewLogger,
			NewValidator,
			providePostgresPool,
			postgres.NewProductRepository,
			postgres.NewOrderRepository,
			service.NewProductService,
			service.NewOrderService,
			handler.NewProductHandler,
			handler.NewOrderHandler,
			adapterhttp.NewRouter,
		),
		fx.Invoke(registerHTTPServer),
	).Run()
}

// NewLogger provides the base *slog.Logger as an fx dependency. Consumers
// should call logger.With("component", "<name>") so log lines can be traced
// back to the component that produced them. Backed by charmbracelet/log for
// colorized, styled CLI output; it implements slog.Handler directly.
func NewLogger() *slog.Logger {
	handler := charmlog.NewWithOptions(os.Stdout, charmlog.Options{
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
	})
	return slog.New(handler)
}

// NewValidator provides the shared *validator.Validate engine as an fx
// dependency. DTO structs (internal/adapter/http/dto) keep gin's familiar
// `binding:"..."` tags; this engine reads that same tag name and reports
// each field by its json name (e.g. "product_id") rather than its Go
// struct field name (e.g. "ProductID"), so handler.bindJSON can turn a
// failure straight into a readable apperror.ErrValidation message.
func NewValidator() *validator.Validate {
	v := validator.New()
	v.SetTagName("binding")
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return v
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
