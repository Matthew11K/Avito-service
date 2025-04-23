package pvz

import (
	"context"
	"fmt"
	"time"

	"avito/internal/domain/pvz"

	"github.com/google/uuid"
)

type Transactor interface {
	WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type Repository interface {
	CreatePVZ(ctx context.Context, city pvz.City) (*pvz.PVZ, error)
	GetPVZByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error)
	GetPVZs(ctx context.Context, startDate, endDate *time.Time, city *pvz.City, page, limit int) ([]pvz.WithReceptions, error)
}

type Service struct {
	repo      Repository
	txManager Transactor
}

func NewService(repo Repository, txManager Transactor) *Service {
	return &Service{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *Service) CreatePVZ(ctx context.Context, req pvz.CreatePVZRequest) (*pvz.PVZ, error) {
	if req.City == "" {
		return nil, &pvz.ErrCityEmpty{}
	}

	if !req.City.Validate() {
		return nil, &pvz.ErrInvalidCity{}
	}

	pvzObj, err := s.repo.CreatePVZ(ctx, req.City)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании ПВЗ: %w", err)
	}

	return pvzObj, nil
}

func (s *Service) GetPVZs(ctx context.Context, req pvz.GetPVZsRequest) ([]pvz.WithReceptions, error) {
	if req.Page <= 0 {
		req.Page = 1
	}

	if req.Limit <= 0 || req.Limit > 30 {
		req.Limit = 10
	}

	return s.repo.GetPVZs(ctx, req.StartDate, req.EndDate, req.City, req.Page, req.Limit)
}

func (s *Service) GetPVZByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	return s.repo.GetPVZByID(ctx, id)
}
