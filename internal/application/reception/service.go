package reception

import (
	"context"
	"fmt"

	domainPVZ "avito/internal/domain/pvz"
	"avito/internal/domain/reception"

	"github.com/google/uuid"
)

type Transactor interface {
	WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type Repository interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
	GetActiveReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
	GetReceptionByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error)
	CloseReception(ctx context.Context, id uuid.UUID) (*reception.Reception, error)
}

type PVZRepository interface {
	GetPVZByID(ctx context.Context, id uuid.UUID) (*domainPVZ.PVZ, error)
}

type Service struct {
	repo      Repository
	pvzRepo   PVZRepository
	txManager Transactor
}

func NewService(repo Repository, pvzRepo PVZRepository, txManager Transactor) *Service {
	return &Service{
		repo:      repo,
		pvzRepo:   pvzRepo,
		txManager: txManager,
	}
}

func (s *Service) CreateReception(ctx context.Context, req reception.CreateReceptionRequest) (*reception.Reception, error) {
	_, err := s.pvzRepo.GetPVZByID(ctx, req.PVZID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке ПВЗ: %w", err)
	}

	var receptionObj *reception.Reception

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		_, err := s.repo.GetActiveReceptionByPVZID(txCtx, req.PVZID)
		if err == nil {
			return &reception.ErrActiveReceptionExists{}
		}

		if _, ok := err.(*reception.ErrNoActiveReception); !ok {
			return fmt.Errorf("ошибка при проверке активной приемки: %w", err)
		}

		receptionObj, err = s.repo.CreateReception(txCtx, req.PVZID)

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании приемки: %w", err)
	}

	return receptionObj, nil
}

func (s *Service) CloseReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	activeReception, err := s.repo.GetActiveReceptionByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}

	var closedReception *reception.Reception

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		closedReception, err = s.repo.CloseReception(txCtx, activeReception.ID)

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка при закрытии приемки: %w", err)
	}

	return closedReception, nil
}

func (s *Service) GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	_, err := s.pvzRepo.GetPVZByID(ctx, pvzID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке ПВЗ: %w", err)
	}

	return s.repo.GetActiveReceptionByPVZID(ctx, pvzID)
}
