//nolint:revive // структура теста требует неиспользуемых параметров для поддержания единообразия
package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito/internal/domain/product"
	"avito/internal/domain/pvz"
	"avito/internal/domain/reception"
	"avito/internal/interfaces/http/dto"
	"avito/internal/interfaces/http/handlers"
	"avito/internal/interfaces/http/handlers/mocks"

	"log/slog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPVZHandler_CreatePVZ(t *testing.T) {
	type args struct {
		request dto.PostPvzJSONRequestBody
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(mockSvc *mocks.PVZService)
		expectedStatus int
		expectedBody   func() *pvz.PVZ
	}{
		{
			name: "Успешное создание ПВЗ в Москве",
			args: args{
				request: dto.PostPvzJSONRequestBody{
					City: dto.PVZCity("Москва"),
				},
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				expectedPVZ := &pvz.PVZ{
					ID:               uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					RegistrationDate: time.Now(),
					City:             pvz.CityMoscow,
				}

				mockSvc.On("CreatePVZ", mock.Anything, pvz.CreatePVZRequest{
					City: pvz.CityMoscow,
				}).Return(expectedPVZ, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() *pvz.PVZ {
				return &pvz.PVZ{
					ID:               uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					RegistrationDate: time.Now(),
					City:             pvz.CityMoscow,
				}
			},
		},
		{
			name: "Успешное создание ПВЗ в Санкт-Петербурге",
			args: args{
				request: dto.PostPvzJSONRequestBody{
					City: dto.PVZCity("Санкт-Петербург"),
				},
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				expectedPVZ := &pvz.PVZ{
					ID:               uuid.MustParse("223e4567-e89b-12d3-a456-426614174001"),
					RegistrationDate: time.Now(),
					City:             pvz.CitySaintPetersburg,
				}

				mockSvc.On("CreatePVZ", mock.Anything, pvz.CreatePVZRequest{
					City: pvz.CitySaintPetersburg,
				}).Return(expectedPVZ, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() *pvz.PVZ {
				return &pvz.PVZ{
					ID:               uuid.MustParse("223e4567-e89b-12d3-a456-426614174001"),
					RegistrationDate: time.Now(),
					City:             pvz.CitySaintPetersburg,
				}
			},
		},
		{
			name: "Успешное создание ПВЗ в Казани",
			args: args{
				request: dto.PostPvzJSONRequestBody{
					City: dto.PVZCity("Казань"),
				},
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				expectedPVZ := &pvz.PVZ{
					ID:               uuid.MustParse("323e4567-e89b-12d3-a456-426614174002"),
					RegistrationDate: time.Now(),
					City:             pvz.CityKazan,
				}

				mockSvc.On("CreatePVZ", mock.Anything, pvz.CreatePVZRequest{
					City: pvz.CityKazan,
				}).Return(expectedPVZ, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() *pvz.PVZ {
				return &pvz.PVZ{
					ID:               uuid.MustParse("323e4567-e89b-12d3-a456-426614174002"),
					RegistrationDate: time.Now(),
					City:             pvz.CityKazan,
				}
			},
		},
		{
			name: "Неверный город",
			args: args{
				request: dto.PostPvzJSONRequestBody{
					City: dto.PVZCity("Новосибирск"),
				},
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name: "Ошибка сервиса при создании ПВЗ",
			args: args{
				request: dto.PostPvzJSONRequestBody{
					City: dto.PVZCity("Москва"),
				},
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				mockSvc.On("CreatePVZ", mock.Anything, pvz.CreatePVZRequest{
					City: pvz.CityMoscow,
				}).Return(nil, fmt.Errorf("ошибка при создании ПВЗ"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name: "Некорректный JSON в запросе",
			args: args{
				request: dto.PostPvzJSONRequestBody{},
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.PVZService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewPVZHandler(mockService, nullLogger)

			var requestBody []byte

			var err error

			if tt.name == "Некорректный JSON в запросе" {
				requestBody = []byte(`{"city": тест"}`)
			} else {
				requestBody, err = json.Marshal(tt.args.request)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.CreatePVZ(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != nil {
				var responseBody dto.PVZ
				err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
				require.NoError(t, err)

				expectedBody := tt.expectedBody()
				assert.Equal(t, expectedBody.City, pvz.City(responseBody.City))

				assert.NotNil(t, responseBody.Id)
				assert.NotNil(t, responseBody.RegistrationDate)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPVZHandler_GetPVZs(t *testing.T) {
	// Подготовка данных для тестов
	pvzID1 := uuid.New()
	receptionID1 := uuid.New()
	productID1 := uuid.New()

	pvzID2 := uuid.New()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Создаем тестовые кейсы
	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMock      func(mockSvc *mocks.PVZService)
		expectedStatus int
		expectedPVZs   int
	}{
		{
			name:        "Успешное получение списка ПВЗ без фильтров",
			queryParams: map[string]string{},
			setupMock: func(mockSvc *mocks.PVZService) {
				pvzs := []pvz.WithReceptions{
					{
						PVZ: pvz.PVZ{
							ID:               pvzID1,
							RegistrationDate: now,
							City:             pvz.CityMoscow,
						},
						Receptions: []pvz.ReceptionWithItems{
							{
								Reception: reception.Reception{
									ID:       receptionID1,
									DateTime: now,
									PVZID:    pvzID1,
									Status:   reception.StatusInProgress,
								},
								Products: []product.Product{
									{
										ID:          productID1,
										DateTime:    now,
										Type:        product.TypeElectronics,
										ReceptionID: receptionID1,
									},
								},
							},
						},
					},
					{
						PVZ: pvz.PVZ{
							ID:               pvzID2,
							RegistrationDate: yesterday,
							City:             pvz.CitySaintPetersburg,
						},
						Receptions: []pvz.ReceptionWithItems{},
					},
				}

				mockSvc.On("GetPVZs", mock.Anything, pvz.GetPVZsRequest{
					Page:  1,
					Limit: 10,
				}).Return(pvzs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPVZs:   2,
		},
		{
			name: "Фильтрация по городу",
			queryParams: map[string]string{
				"city": "Москва",
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				city := pvz.CityMoscow

				pvzs := []pvz.WithReceptions{
					{
						PVZ: pvz.PVZ{
							ID:               pvzID1,
							RegistrationDate: now,
							City:             pvz.CityMoscow,
						},
						Receptions: []pvz.ReceptionWithItems{},
					},
				}

				mockSvc.On("GetPVZs", mock.Anything, pvz.GetPVZsRequest{
					City:  &city,
					Page:  1,
					Limit: 10,
				}).Return(pvzs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPVZs:   1,
		},
		{
			name: "Фильтрация по дате",
			queryParams: map[string]string{
				"startDate": yesterday.Format(time.RFC3339),
				"endDate":   tomorrow.Format(time.RFC3339),
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				mockSvc.On("GetPVZs", mock.Anything, mock.MatchedBy(func(req pvz.GetPVZsRequest) bool {
					return req.Page == 1 && req.Limit == 10 &&
						req.StartDate != nil && req.EndDate != nil
				})).Return([]pvz.WithReceptions{
					{
						PVZ: pvz.PVZ{
							ID:               pvzID1,
							RegistrationDate: now,
							City:             pvz.CityMoscow,
						},
						Receptions: []pvz.ReceptionWithItems{},
					},
					{
						PVZ: pvz.PVZ{
							ID:               pvzID2,
							RegistrationDate: yesterday,
							City:             pvz.CitySaintPetersburg,
						},
						Receptions: []pvz.ReceptionWithItems{},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPVZs:   2,
		},
		{
			name: "Пагинация",
			queryParams: map[string]string{
				"page":  "2",
				"limit": "5",
			},
			setupMock: func(mockSvc *mocks.PVZService) {
				pvzs := []pvz.WithReceptions{
					{
						PVZ: pvz.PVZ{
							ID:               pvzID1,
							RegistrationDate: now,
							City:             pvz.CityMoscow,
						},
						Receptions: []pvz.ReceptionWithItems{},
					},
				}

				mockSvc.On("GetPVZs", mock.Anything, pvz.GetPVZsRequest{
					Page:  2,
					Limit: 5,
				}).Return(pvzs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedPVZs:   1,
		},
		{
			name: "Некорректный параметр page",
			queryParams: map[string]string{
				"page": "invalid",
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedPVZs:   0,
		},
		{
			name: "Некорректный параметр limit",
			queryParams: map[string]string{
				"limit": "invalid",
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedPVZs:   0,
		},
		{
			name: "Некорректный параметр startDate",
			queryParams: map[string]string{
				"startDate": "invalid-date",
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedPVZs:   0,
		},
		{
			name: "Некорректный параметр endDate",
			queryParams: map[string]string{
				"endDate": "invalid-date",
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedPVZs:   0,
		},
		{
			name: "Некорректный параметр city",
			queryParams: map[string]string{
				"city": "Неизвестный",
			},
			setupMock:      func(mockSvc *mocks.PVZService) {},
			expectedStatus: http.StatusBadRequest,
			expectedPVZs:   0,
		},
		{
			name:        "Ошибка сервиса",
			queryParams: map[string]string{},
			setupMock: func(mockSvc *mocks.PVZService) {
				mockSvc.On("GetPVZs", mock.Anything, pvz.GetPVZsRequest{
					Page:  1,
					Limit: 10,
				}).Return(nil, fmt.Errorf("ошибка при получении списка ПВЗ"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedPVZs:   0,
		},
	}

	// Выполняем тесты
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок сервиса
			mockService := new(mocks.PVZService)
			tt.setupMock(mockService)

			// Создаем обработчик с моком
			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewPVZHandler(mockService, nullLogger)

			// Подготавливаем запрос с параметрами
			req, err := http.NewRequest(http.MethodGet, "/pvz", http.NoBody)
			require.NoError(t, err)

			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}

			req.URL.RawQuery = q.Encode()

			// Создаем ResponseRecorder для записи ответа
			recorder := httptest.NewRecorder()

			// Вызываем обработчик
			handler.GetPVZs(recorder, req)

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Если ожидается успешный ответ, проверяем тело
			if tt.expectedStatus == http.StatusOK {
				var responseBody []map[string]interface{}
				err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
				require.NoError(t, err)

				assert.Len(t, responseBody, tt.expectedPVZs)

				// Проверяем структуру ответа для первого ПВЗ, если он есть
				if len(responseBody) > 0 {
					// Проверяем наличие ключа "pvz" в ответе
					assert.Contains(t, responseBody[0], "pvz")

					// Проверяем, что поля ПВЗ присутствуют
					pvzData, ok := responseBody[0]["pvz"].(map[string]interface{})
					require.True(t, ok)
					assert.Contains(t, pvzData, "id")
					assert.Contains(t, pvzData, "registrationDate")
					assert.Contains(t, pvzData, "city")

					// Проверяем наличие ключа "receptions" в ответе
					assert.Contains(t, responseBody[0], "receptions")
				}
			}

			// Проверяем, что все ожидания мока были выполнены
			mockService.AssertExpectations(t)
		})
	}
}
