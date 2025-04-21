package adapters

import (
	"context"
	"errors"

	appProduct "avito/internal/application/product"
	"avito/internal/domain/product"
	"avito/internal/domain/reception"
	"avito/internal/interfaces/http/handlers"

	"github.com/google/uuid"
)

type ProductServiceAdapter struct {
	service *appProduct.Service
}

func NewProductServiceAdapter(service *appProduct.Service) *ProductServiceAdapter {
	return &ProductServiceAdapter{
		service: service,
	}
}

func (a *ProductServiceAdapter) CreateProduct(ctx context.Context, pvzID uuid.UUID, productType product.Type) (*product.Product, error) {
	req := product.CreateProductRequest{
		Type:  productType,
		PVZID: pvzID,
	}

	prod, err := a.service.AddProduct(ctx, req)
	if err != nil {
		var noActiveReceptionErr *reception.ErrNoActiveReception
		if errors.As(err, &noActiveReceptionErr) {
			return nil, handlers.ErrNoActiveReceptionProduct
		}

		var receptionClosedErr *reception.ErrReceptionClosed
		if errors.As(err, &receptionClosedErr) {
			return nil, handlers.ErrReceptionClosedForProduct
		}

		return nil, err
	}

	return prod, nil
}

func (a *ProductServiceAdapter) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	err := a.service.DeleteLastProduct(ctx, pvzID)
	if err != nil {
		var noActiveReceptionErr *reception.ErrNoActiveReception
		if errors.As(err, &noActiveReceptionErr) {
			return handlers.ErrNoActiveReceptionProduct
		}

		var receptionClosedErr *reception.ErrReceptionClosed
		if errors.As(err, &receptionClosedErr) {
			return handlers.ErrReceptionClosedForProduct
		}

		var noProductsErr *product.ErrNoProductsToDelete
		if errors.As(err, &noProductsErr) {
			return handlers.ErrNoProductsToDelete
		}

		return err
	}

	return nil
}
