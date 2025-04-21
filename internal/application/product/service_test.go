package product_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"avito/internal/application/product"
	"avito/internal/application/product/mocks"
	domainProduct "avito/internal/domain/product"
	domainPVZ "avito/internal/domain/pvz"
	domainReception "avito/internal/domain/reception"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_AddProduct(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name              string
		request           domainProduct.CreateProductRequest
		mockSetup         func(*mocks.Repository, *mocks.ReceptionRepository, *mocks.PVZRepository, *mocks.Transactor)
		expectedResult    *domainProduct.Product
		expectedErrorText string
	}{
		{
			name: "Успешное добавление товара",
			request: domainProduct.CreateProductRequest{
				Type:  domainProduct.TypeElectronics,
				PVZID: pvzID,
			},
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				reception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				receptionRepo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(reception, nil)
				receptionRepo.On("GetReceptionByID", mock.Anything, receptionID).Return(reception, nil)

				createdProduct := &domainProduct.Product{
					ID:          uuid.New(),
					DateTime:    time.Now(),
					Type:        domainProduct.TypeElectronics,
					ReceptionID: receptionID,
				}
				repo.On("AddProduct", mock.Anything, domainProduct.TypeElectronics, receptionID).Return(createdProduct, nil)

				tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedResult: &domainProduct.Product{
				Type:        domainProduct.TypeElectronics,
				ReceptionID: receptionID,
			},
			expectedErrorText: "",
		},
		{
			name: "Пустой тип товара",
			request: domainProduct.CreateProductRequest{
				Type:  "",
				PVZID: pvzID,
			},
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				// Моки не должны вызываться
			},
			expectedResult:    nil,
			expectedErrorText: "тип товара не может быть пустым",
		},
		{
			name: "Неверный тип товара",
			request: domainProduct.CreateProductRequest{
				Type:  "неверный_тип",
				PVZID: pvzID,
			},
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				// Моки не должны вызываться
			},
			expectedResult:    nil,
			expectedErrorText: "неверный тип товара",
		},
		{
			name: "ПВЗ не найден",
			request: domainProduct.CreateProductRequest{
				Type:  domainProduct.TypeElectronics,
				PVZID: pvzID,
			},
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(nil, &domainPVZ.ErrPVZNotFound{})
			},
			expectedResult:    nil,
			expectedErrorText: "ПВЗ не найден",
		},
		{
			name: "Приемка не найдена",
			request: domainProduct.CreateProductRequest{
				Type:  domainProduct.TypeElectronics,
				PVZID: pvzID,
			},
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				receptionRepo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(nil, &domainReception.ErrReceptionNotFound{})
			},
			expectedResult:    nil,
			expectedErrorText: "приемка не найдена",
		},
		{
			name: "Приемка закрыта",
			request: domainProduct.CreateProductRequest{
				Type:  domainProduct.TypeElectronics,
				PVZID: pvzID,
			},
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				reception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusClosed,
				}
				receptionRepo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(reception, nil)
			},
			expectedResult:    nil,
			expectedErrorText: "приемка уже закрыта",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockReceptionRepo := new(mocks.ReceptionRepository)
			mockPVZRepo := new(mocks.PVZRepository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockReceptionRepo, mockPVZRepo, mockTx)
			}

			service := product.NewService(mockRepo, mockReceptionRepo, mockPVZRepo, mockTx)

			result, err := service.AddProduct(context.Background(), tt.request)

			if tt.expectedErrorText != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErrorText),
					"Ожидалось сообщение об ошибке, содержащее '%s', получено: '%s'",
					tt.expectedErrorText, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Type, result.Type)
				assert.Equal(t, tt.expectedResult.ReceptionID, result.ReceptionID)
			}

			mockRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
			mockPVZRepo.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

func TestService_DeleteLastProduct(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name              string
		pvzID             uuid.UUID
		mockSetup         func(*mocks.Repository, *mocks.ReceptionRepository, *mocks.PVZRepository, *mocks.Transactor)
		expectedErrorText string
	}{
		{
			name:  "Успешное удаление товара",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				reception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				receptionRepo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(reception, nil)
				receptionRepo.On("GetReceptionByID", mock.Anything, receptionID).Return(reception, nil)

				repo.On("DeleteLastProduct", mock.Anything, receptionID).Return(nil)

				tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedErrorText: "",
		},
		{
			name:  "ПВЗ не найден",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(nil, &domainPVZ.ErrPVZNotFound{})
			},
			expectedErrorText: "ПВЗ не найден",
		},
		{
			name:  "Приемка не найдена",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				receptionRepo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(nil, &domainReception.ErrReceptionNotFound{})
			},
			expectedErrorText: "приемка не найдена",
		},
		{
			name:  "Нет товаров для удаления",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				reception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				receptionRepo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(reception, nil)
				receptionRepo.On("GetReceptionByID", mock.Anything, receptionID).Return(reception, nil)

				repo.On("DeleteLastProduct", mock.Anything, receptionID).Return(&domainProduct.ErrNoProductsToDelete{})

				tx.On("WithTransaction", mock.Anything,
					mock.AnythingOfType("func(context.Context) error")).Return(&domainProduct.ErrNoProductsToDelete{}).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedErrorText: "нет товаров для удаления",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockReceptionRepo := new(mocks.ReceptionRepository)
			mockPVZRepo := new(mocks.PVZRepository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockReceptionRepo, mockPVZRepo, mockTx)
			}

			service := product.NewService(mockRepo, mockReceptionRepo, mockPVZRepo, mockTx)

			err := service.DeleteLastProduct(context.Background(), tt.pvzID)

			if tt.expectedErrorText != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErrorText),
					"Ожидалось сообщение об ошибке, содержащее '%s', получено: '%s'",
					tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
			mockPVZRepo.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

func TestService_GetProductsByReceptionID(t *testing.T) {
	receptionID := uuid.New()

	tests := []struct {
		name              string
		receptionID       uuid.UUID
		mockSetup         func(*mocks.Repository, *mocks.ReceptionRepository)
		expectedResult    []domainProduct.Product
		expectedErrorText string
	}{
		{
			name:        "Успешное получение товаров",
			receptionID: receptionID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository) {
				reception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    uuid.New(),
					Status:   domainReception.StatusInProgress,
				}
				receptionRepo.On("GetReceptionByID", mock.Anything, receptionID).Return(reception, nil)

				products := []domainProduct.Product{
					{
						ID:          uuid.New(),
						DateTime:    time.Now(),
						Type:        domainProduct.TypeElectronics,
						ReceptionID: receptionID,
					},
					{
						ID:          uuid.New(),
						DateTime:    time.Now(),
						Type:        domainProduct.TypeClothes,
						ReceptionID: receptionID,
					},
				}
				repo.On("GetProductsByReceptionID", mock.Anything, receptionID).Return(products, nil)
			},
			expectedResult: []domainProduct.Product{
				{
					Type:        domainProduct.TypeElectronics,
					ReceptionID: receptionID,
				},
				{
					Type:        domainProduct.TypeClothes,
					ReceptionID: receptionID,
				},
			},
			expectedErrorText: "",
		},
		{
			name:        "Приемка не найдена",
			receptionID: receptionID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository) {
				receptionRepo.On("GetReceptionByID", mock.Anything, receptionID).Return(nil, &domainReception.ErrReceptionNotFound{})
			},
			expectedResult:    nil,
			expectedErrorText: "приемка не найдена",
		},
		{
			name:        "Пустой список товаров",
			receptionID: receptionID,
			mockSetup: func(repo *mocks.Repository, receptionRepo *mocks.ReceptionRepository) {
				reception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    uuid.New(),
					Status:   domainReception.StatusInProgress,
				}
				receptionRepo.On("GetReceptionByID", mock.Anything, receptionID).Return(reception, nil)

				products := []domainProduct.Product{}
				repo.On("GetProductsByReceptionID", mock.Anything, receptionID).Return(products, nil)
			},
			expectedResult:    []domainProduct.Product{},
			expectedErrorText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockReceptionRepo := new(mocks.ReceptionRepository)
			mockPVZRepo := new(mocks.PVZRepository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockReceptionRepo)
			}

			service := product.NewService(mockRepo, mockReceptionRepo, mockPVZRepo, mockTx)

			products, err := service.GetProductsByReceptionID(context.Background(), tt.receptionID)

			if tt.expectedErrorText != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErrorText),
					"Ожидалось сообщение об ошибке, содержащее '%s', получено: '%s'",
					tt.expectedErrorText, err.Error())
				assert.Nil(t, products)
			} else {
				assert.NoError(t, err)
				assert.Len(t, products, len(tt.expectedResult))

				if len(products) > 0 && len(tt.expectedResult) > 0 {
					for i, p := range products {
						assert.Equal(t, tt.expectedResult[i].Type, p.Type)
						assert.Equal(t, tt.expectedResult[i].ReceptionID, p.ReceptionID)
					}
				}
			}

			mockRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
		})
	}
}
