package handler

import (
	"context"
	"reflect"
	"strings"

	"food-store-apis/internal/domain/model"
	"food-store-apis/internal/port"

	"github.com/go-playground/validator/v10"
)

// newTestValidator mirrors cmd/api/main.go's NewValidator: DTOs keep gin's
// `binding:"..."` tags, and fields are reported by their json name (e.g.
// "product_id") rather than the Go struct field name.
func newTestValidator() *validator.Validate {
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

type fakeOrderService struct {
	order  *model.Order
	err    error
	called bool
}

func (f *fakeOrderService) CreateOrder(ctx context.Context, input port.CreateOrderInput) (*model.Order, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	return f.order, nil
}

type fakeProductService struct {
	products []*model.Product
	err      error
}

func (f *fakeProductService) ListProducts(ctx context.Context) ([]*model.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}
