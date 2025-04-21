package product

import (
	"context"
	"fmt"

	"avito/internal/domain/product"
	domainPVZ "avito/internal/domain/pvz"
	"avito/internal/domain/reception"

	"github.com/google/uuid"
)

type Transactor interface {
	WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type Repository interface {
	AddProduct(ctx context.Context, productType product.Type, receptionID uuid.UUID) (*product.Product, error)
	DeleteLastProduct(ctx context.Context, receptionID uuid.UUID) error
	GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]product.Product, error)
}

type ReceptionRepository interface {
	GetActiveReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
	GetReceptionByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error)
}

type PVZRepository interface {
	GetPVZByID(ctx context.Context, id uuid.UUID) (*domainPVZ.PVZ, error)
}

type Service struct {
	repo          Repository
	receptionRepo ReceptionRepository
	pvzRepo       PVZRepository
	txManager     Transactor
}

func NewService(repo Repository, receptionRepo ReceptionRepository, pvzRepo PVZRepository, txManager Transactor) *Service {
	return &Service{
		repo:          repo,
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		txManager:     txManager,
	}
}

func (s *Service) AddProduct(ctx context.Context, req product.CreateProductRequest) (*product.Product, error) {
	if req.Type == "" {
		return nil, &product.ErrTypeEmpty{}
	}

	if !req.Type.Validate() {
		return nil, &product.ErrInvalidProductType{}
	}

	_, err := s.pvzRepo.GetPVZByID(ctx, req.PVZID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке ПВЗ: %w", err)
	}

	activeReception, err := s.receptionRepo.GetActiveReceptionByPVZID(ctx, req.PVZID)
	if err != nil {
		return nil, err
	}

	if activeReception.Status == reception.StatusClosed {
		return nil, &reception.ErrReceptionClosed{}
	}

	var productObj *product.Product

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		currReception, err := s.receptionRepo.GetReceptionByID(txCtx, activeReception.ID)
		if err != nil {
			return fmt.Errorf("ошибка при проверке приемки: %w", err)
		}

		if currReception.Status == reception.StatusClosed {
			return &reception.ErrReceptionClosed{}
		}

		productObj, err = s.repo.AddProduct(txCtx, req.Type, activeReception.ID)

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка при добавлении товара: %w", err)
	}

	return productObj, nil
}

func (s *Service) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	_, err := s.pvzRepo.GetPVZByID(ctx, pvzID)
	if err != nil {
		return fmt.Errorf("ошибка при проверке ПВЗ: %w", err)
	}

	activeReception, err := s.receptionRepo.GetActiveReceptionByPVZID(ctx, pvzID)
	if err != nil {
		return err
	}

	if activeReception.Status == reception.StatusClosed {
		return &reception.ErrReceptionClosed{}
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		currReception, err := s.receptionRepo.GetReceptionByID(txCtx, activeReception.ID)
		if err != nil {
			return fmt.Errorf("ошибка при проверке приемки: %w", err)
		}

		if currReception.Status == reception.StatusClosed {
			return &reception.ErrReceptionClosed{}
		}

		return s.repo.DeleteLastProduct(txCtx, activeReception.ID)
	})

	if err != nil {
		return fmt.Errorf("ошибка при удалении товара: %w", err)
	}

	return nil
}

func (s *Service) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]product.Product, error) {
	_, err := s.receptionRepo.GetReceptionByID(ctx, receptionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке приемки: %w", err)
	}

	return s.repo.GetProductsByReceptionID(ctx, receptionID)
}
