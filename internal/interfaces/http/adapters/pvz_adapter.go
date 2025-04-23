package adapters

import (
	"context"

	appPVZ "avito/internal/application/pvz"
	"avito/internal/domain/pvz"
)

type PVZServiceAdapter struct {
	service *appPVZ.Service
}

func NewPVZServiceAdapter(service *appPVZ.Service) *PVZServiceAdapter {
	return &PVZServiceAdapter{
		service: service,
	}
}

func (a *PVZServiceAdapter) CreatePVZ(ctx context.Context, req pvz.CreatePVZRequest) (*pvz.PVZ, error) {
	return a.service.CreatePVZ(ctx, req)
}

func (a *PVZServiceAdapter) GetPVZs(ctx context.Context, req pvz.GetPVZsRequest) ([]pvz.WithReceptions, error) {
	return a.service.GetPVZs(ctx, req)
}
