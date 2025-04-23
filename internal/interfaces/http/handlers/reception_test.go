//nolint:revive // структура теста требует неиспользуемых параметров для поддержания единообразия
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestReceptionHandler_CreateReception(t *testing.T) {
	type args struct {
		request dto.PostReceptionsJSONRequestBody
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(mockSvc *mocks.ReceptionService)
		expectedStatus int
		expectedBody   func() *reception.Reception
	}{
		{
			name: "Успешное создание приемки",
			args: args{
				request: dto.PostReceptionsJSONRequestBody{
					PvzId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				},
			},
			setupMock: func(mockSvc *mocks.ReceptionService) {
				expectedReception := &reception.Reception{
					ID:       uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					DateTime: time.Now(),
					PVZID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Status:   reception.StatusInProgress,
				}

				mockSvc.On("CreateReception", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).
					Return(expectedReception, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() *reception.Reception {
				return &reception.Reception{
					ID:       uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					DateTime: time.Now(),
					PVZID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Status:   reception.StatusInProgress,
				}
			},
		},
		{
			name: "Уже есть активная приемка",
			args: args{
				request: dto.PostReceptionsJSONRequestBody{
					PvzId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				},
			},
			setupMock: func(mockSvc *mocks.ReceptionService) {
				mockSvc.On("CreateReception", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).
					Return(nil, handlers.ErrActiveReceptionExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.ReceptionService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewReceptionHandler(mockService, nullLogger)

			requestBody, err := json.Marshal(tt.args.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.CreateReception(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != nil {
				var responseBody dto.Reception
				err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
				require.NoError(t, err)

				expectedBody := tt.expectedBody()

				assert.NotNil(t, responseBody.Id)
				assert.Equal(t, expectedBody.PVZID.String(), responseBody.PvzId.String())
				assert.Equal(t, dto.InProgress, responseBody.Status)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestReceptionHandler_CloseLastReception(t *testing.T) {
	pvzID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	recID := uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		url            string
		setupMock      func(mockSvc *mocks.ReceptionService)
		expectedStatus int
	}{
		{
			name: "Успешное закрытие приемки",
			url:  "/pvz/" + pvzID.String() + "/close_last_reception",
			setupMock: func(mockSvc *mocks.ReceptionService) {
				closedReception := &reception.Reception{
					ID:       recID,
					DateTime: time.Now(),
					PVZID:    pvzID,
					Status:   reception.StatusClosed,
				}

				mockSvc.On("CloseLastReception", mock.Anything, pvzID).
					Return(closedReception, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Нет активной приемки",
			url:  "/pvz/" + pvzID.String() + "/close_last_reception",
			setupMock: func(mockSvc *mocks.ReceptionService) {
				mockSvc.On("CloseLastReception", mock.Anything, pvzID).
					Return(nil, handlers.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Неверный URL",
			url:  "/invalid/url",
			setupMock: func(mockSvc *mocks.ReceptionService) {
				// Метод не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.ReceptionService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewReceptionHandler(mockService, nullLogger)

			req, err := http.NewRequest(http.MethodPost, tt.url, http.NoBody)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			handler.CloseLastReception(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			mockService.AssertExpectations(t)
		})
	}
}
