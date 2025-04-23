//nolint:revive // структура теста требует неиспользуемых параметров для поддержания единообразия
package pvz_test

import (
	"context"
	"testing"
	"time"

	"avito/internal/application/pvz"
	"avito/internal/application/pvz/mocks"
	domainPvz "avito/internal/domain/pvz"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreatePVZ(t *testing.T) {
	tests := []struct {
		name          string
		request       domainPvz.CreatePVZRequest
		mockSetup     func(*mocks.Repository, *mocks.Transactor)
		expectedPVZ   *domainPvz.PVZ
		expectedError error
	}{
		{
			name: "Успешное создание ПВЗ",
			request: domainPvz.CreatePVZRequest{
				City: domainPvz.CityMoscow,
			},
			mockSetup: func(_repo *mocks.Repository, tx *mocks.Transactor) {
				expectedPVZ := &domainPvz.PVZ{
					ID:               uuid.New(),
					RegistrationDate: time.Now(),
					City:             domainPvz.CityMoscow,
				}
				_repo.On("CreatePVZ", mock.Anything, domainPvz.CityMoscow).Return(expectedPVZ, nil)
			},
			expectedPVZ: &domainPvz.PVZ{
				City: domainPvz.CityMoscow,
			},
			expectedError: nil,
		},
		{
			name: "Пустой город",
			request: domainPvz.CreatePVZRequest{
				City: "",
			},
			mockSetup: func(_repo *mocks.Repository, tx *mocks.Transactor) {
				// Мок не должен вызываться
			},
			expectedPVZ:   nil,
			expectedError: &domainPvz.ErrCityEmpty{},
		},
		{
			name: "Неверный город",
			request: domainPvz.CreatePVZRequest{
				City: "Неверный город",
			},
			mockSetup: func(_repo *mocks.Repository, tx *mocks.Transactor) {
				// Мок не должен вызываться
			},
			expectedPVZ:   nil,
			expectedError: &domainPvz.ErrInvalidCity{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockTx)
			}

			service := pvz.NewService(mockRepo, mockTx)

			actualPVZ, err := service.CreatePVZ(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.IsType(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualPVZ)
				assert.Equal(t, tt.expectedPVZ.City, actualPVZ.City)
			}

			mockRepo.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

func TestService_GetPVZs(t *testing.T) {
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now
	moscow := domainPvz.CityMoscow

	tests := []struct {
		name          string
		request       domainPvz.GetPVZsRequest
		mockSetup     func(*mocks.Repository)
		expectedItems []domainPvz.WithReceptions
		expectedError error
	}{
		{
			name: "Успешное получение ПВЗ с параметрами по умолчанию",
			request: domainPvz.GetPVZsRequest{
				Page:  0,
				Limit: 0,
			},
			mockSetup: func(repo *mocks.Repository) {
				expectedItems := []domainPvz.WithReceptions{
					{
						PVZ: domainPvz.PVZ{
							ID:               uuid.New(),
							RegistrationDate: now,
							City:             domainPvz.CityMoscow,
						},
						Receptions: []domainPvz.ReceptionWithItems{},
					},
				}
				repo.On("GetPVZs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, 1, 10).Return(expectedItems, nil)
			},
			expectedItems: []domainPvz.WithReceptions{
				{
					PVZ: domainPvz.PVZ{
						City: domainPvz.CityMoscow,
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Получение ПВЗ с фильтрацией",
			request: domainPvz.GetPVZsRequest{
				StartDate: &startDate,
				EndDate:   &endDate,
				City:      &moscow,
				Page:      2,
				Limit:     5,
			},
			mockSetup: func(repo *mocks.Repository) {
				expectedItems := []domainPvz.WithReceptions{}
				repo.On("GetPVZs", mock.Anything, mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz.City"), 2, 5).Return(expectedItems, nil)
			},
			expectedItems: []domainPvz.WithReceptions{},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := pvz.NewService(mockRepo, mockTx)

			actualItems, err := service.GetPVZs(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualItems)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetPVZByID(t *testing.T) {
	pvzID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*mocks.Repository)
		expectedPVZ   *domainPvz.PVZ
		expectedError error
	}{
		{
			name: "Успешное получение ПВЗ по ID",
			id:   pvzID,
			mockSetup: func(repo *mocks.Repository) {
				expectedPVZ := &domainPvz.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPvz.CityMoscow,
				}
				repo.On("GetPVZByID", mock.Anything, pvzID).Return(expectedPVZ, nil)
			},
			expectedPVZ: &domainPvz.PVZ{
				ID:   pvzID,
				City: domainPvz.CityMoscow,
			},
			expectedError: nil,
		},
		{
			name: "ПВЗ не найден",
			id:   pvzID,
			mockSetup: func(repo *mocks.Repository) {
				repo.On("GetPVZByID", mock.Anything, pvzID).Return(nil, &domainPvz.ErrPVZNotFound{})
			},
			expectedPVZ:   nil,
			expectedError: &domainPvz.ErrPVZNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := pvz.NewService(mockRepo, mockTx)

			actualPVZ, err := service.GetPVZByID(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, actualPVZ)
			} else {
				require.NoError(t, err)
				require.NotNil(t, actualPVZ)
				assert.Equal(t, tt.expectedPVZ.ID, actualPVZ.ID)
				assert.Equal(t, tt.expectedPVZ.City, actualPVZ.City)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
