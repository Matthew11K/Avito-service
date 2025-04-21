package reception_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"avito/internal/application/reception"
	"avito/internal/application/reception/mocks"
	domainPVZ "avito/internal/domain/pvz"
	domainReception "avito/internal/domain/reception"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateReception(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name              string
		request           domainReception.CreateReceptionRequest
		mockSetup         func(*mocks.Repository, *mocks.PVZRepository, *mocks.Transactor)
		expectedResult    *domainReception.Reception
		expectedErrorText string
	}{
		{
			name: "Успешное создание приемки",
			request: domainReception.CreateReceptionRequest{
				PVZID: pvzID,
			},
			mockSetup: func(_repo *mocks.Repository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				_repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(nil, &domainReception.ErrNoActiveReception{})

				newReception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				_repo.On("CreateReception", mock.Anything, pvzID).Return(newReception, nil)

				tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedResult: &domainReception.Reception{
				ID:     receptionID,
				PVZID:  pvzID,
				Status: domainReception.StatusInProgress,
			},
			expectedErrorText: "",
		},
		{
			name: "ПВЗ не найден",
			request: domainReception.CreateReceptionRequest{
				PVZID: pvzID,
			},
			mockSetup: func(_repo *mocks.Repository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(nil, &domainPVZ.ErrPVZNotFound{})
			},
			expectedResult:    nil,
			expectedErrorText: "ПВЗ не найден",
		},
		{
			name: "Уже есть активная приемка",
			request: domainReception.CreateReceptionRequest{
				PVZID: pvzID,
			},
			mockSetup: func(_repo *mocks.Repository, pvzRepo *mocks.PVZRepository, tx *mocks.Transactor) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				existingReception := &domainReception.Reception{
					ID:       uuid.New(),
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				_repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(existingReception, nil)

				tx.On("WithTransaction", mock.Anything,
					mock.AnythingOfType("func(context.Context) error")).Return(&domainReception.ErrActiveReceptionExists{}).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedResult:    nil,
			expectedErrorText: "уже есть незакрытая приемка",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockPVZRepo := new(mocks.PVZRepository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockPVZRepo, mockTx)
			}

			service := reception.NewService(mockRepo, mockPVZRepo, mockTx)

			result, err := service.CreateReception(context.Background(), tt.request)

			if tt.expectedErrorText != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErrorText),
					"Ожидалось сообщение об ошибке, содержащее '%s', получено: '%s'",
					tt.expectedErrorText, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.PVZID, result.PVZID)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
			}

			mockRepo.AssertExpectations(t)
			mockPVZRepo.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

func TestService_CloseReception(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name              string
		pvzID             uuid.UUID
		mockSetup         func(*mocks.Repository, *mocks.Transactor)
		expectedResult    *domainReception.Reception
		expectedErrorText string
	}{
		{
			name:  "Успешное закрытие приемки",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, _tx *mocks.Transactor) {
				activeReception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(activeReception, nil)

				closedReception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusClosed,
				}
				repo.On("CloseReception", mock.Anything, receptionID).Return(closedReception, nil)

				_tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedResult: &domainReception.Reception{
				ID:     receptionID,
				PVZID:  pvzID,
				Status: domainReception.StatusClosed,
			},
			expectedErrorText: "",
		},
		{
			name:  "Нет активной приемки",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, _tx *mocks.Transactor) {
				repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(nil, &domainReception.ErrNoActiveReception{})
			},
			expectedResult:    nil,
			expectedErrorText: "нет активной приемки",
		},
		{
			name:  "Ошибка при закрытии приемки",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, _tx *mocks.Transactor) {
				activeReception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(activeReception, nil)

				repo.On("CloseReception", mock.Anything, receptionID).Return(nil, assert.AnError)

				_tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(assert.AnError).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedResult:    nil,
			expectedErrorText: "ошибка при закрытии приемки",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockPVZRepo := new(mocks.PVZRepository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockTx)
			}

			service := reception.NewService(mockRepo, mockPVZRepo, mockTx)

			result, err := service.CloseReception(context.Background(), tt.pvzID)

			if tt.expectedErrorText != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErrorText),
					"Ожидалось сообщение об ошибке, содержащее '%s', получено: '%s'",
					tt.expectedErrorText, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.PVZID, result.PVZID)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
			}

			mockRepo.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

func TestService_GetActiveReception(t *testing.T) {
	pvzID := uuid.New()
	receptionID := uuid.New()

	tests := []struct {
		name              string
		pvzID             uuid.UUID
		mockSetup         func(*mocks.Repository, *mocks.PVZRepository)
		expectedResult    *domainReception.Reception
		expectedErrorText string
	}{
		{
			name:  "Успешное получение активной приемки",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, pvzRepo *mocks.PVZRepository) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				activeReception := &domainReception.Reception{
					ID:       receptionID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   domainReception.StatusInProgress,
				}
				repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(activeReception, nil)
			},
			expectedResult: &domainReception.Reception{
				ID:     receptionID,
				PVZID:  pvzID,
				Status: domainReception.StatusInProgress,
			},
			expectedErrorText: "",
		},
		{
			name:  "ПВЗ не найден",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, pvzRepo *mocks.PVZRepository) {
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(nil, &domainPVZ.ErrPVZNotFound{})
			},
			expectedResult:    nil,
			expectedErrorText: "ПВЗ не найден",
		},
		{
			name:  "Нет активной приемки",
			pvzID: pvzID,
			mockSetup: func(repo *mocks.Repository, pvzRepo *mocks.PVZRepository) {
				pvz := &domainPVZ.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					City:             domainPVZ.CityMoscow,
				}
				pvzRepo.On("GetPVZByID", mock.Anything, pvzID).Return(pvz, nil)

				repo.On("GetActiveReceptionByPVZID", mock.Anything, pvzID).Return(nil, &domainReception.ErrNoActiveReception{})
			},
			expectedResult:    nil,
			expectedErrorText: "нет активной приемки",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockPVZRepo := new(mocks.PVZRepository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockPVZRepo)
			}

			service := reception.NewService(mockRepo, mockPVZRepo, mockTx)

			result, err := service.GetActiveReception(context.Background(), tt.pvzID)

			if tt.expectedErrorText != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErrorText),
					"Ожидалось сообщение об ошибке, содержащее '%s', получено: '%s'",
					tt.expectedErrorText, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.PVZID, result.PVZID)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
			}

			mockRepo.AssertExpectations(t)
			mockPVZRepo.AssertExpectations(t)
		})
	}
}
