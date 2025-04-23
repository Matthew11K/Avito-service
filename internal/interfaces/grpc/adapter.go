package grpc

import (
	"context"

	"avito/internal/domain/pvz"

	"github.com/google/uuid"
)

type DomainPVZService interface {
	GetPVZByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error)
	GetPVZs(ctx context.Context, req pvz.GetPVZsRequest) ([]pvz.WithReceptions, error)
}

type PVZServiceAdapter struct {
	domainService DomainPVZService
}

func NewPVZServiceAdapter(domainService DomainPVZService) *PVZServiceAdapter {
	return &PVZServiceAdapter{
		domainService: domainService,
	}
}

func (a *PVZServiceAdapter) GetPVZByID(ctx context.Context, id string) (*pvz.PVZ, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return a.domainService.GetPVZByID(ctx, uuid)
}

func (a *PVZServiceAdapter) GetPVZs(ctx context.Context, req pvz.GetPVZsRequest) ([]pvz.WithReceptions, error) {
	return a.domainService.GetPVZs(ctx, req)
}
