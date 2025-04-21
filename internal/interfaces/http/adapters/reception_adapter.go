package adapters

import (
	"context"
	"errors"

	appReception "avito/internal/application/reception"
	"avito/internal/domain/reception"
	"avito/internal/interfaces/http/handlers"

	"github.com/google/uuid"
)

type ReceptionServiceAdapter struct {
	service *appReception.Service
}

func NewReceptionServiceAdapter(service *appReception.Service) *ReceptionServiceAdapter {
	return &ReceptionServiceAdapter{
		service: service,
	}
}

func (a *ReceptionServiceAdapter) CreateReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	req := reception.CreateReceptionRequest{
		PVZID: pvzID,
	}

	rec, err := a.service.CreateReception(ctx, req)
	if err != nil {
		var activeReceptionErr *reception.ErrActiveReceptionExists
		if errors.As(err, &activeReceptionErr) {
			return nil, handlers.ErrActiveReceptionExists
		}

		return nil, err
	}

	return rec, nil
}

func (a *ReceptionServiceAdapter) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	rec, err := a.service.CloseReception(ctx, pvzID)
	if err != nil {
		var noActiveReceptionErr *reception.ErrNoActiveReception
		if errors.As(err, &noActiveReceptionErr) {
			return nil, handlers.ErrNoActiveReception
		}

		return nil, err
	}

	return rec, nil
}
