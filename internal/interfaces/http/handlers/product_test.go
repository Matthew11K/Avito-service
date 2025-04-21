package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito/internal/domain/product"
	"avito/internal/interfaces/http/dto"
	"avito/internal/interfaces/http/handlers"
	"avito/internal/interfaces/http/handlers/mocks"

	"log/slog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductHandler_CreateProduct(t *testing.T) {
	type args struct {
		request dto.PostProductsJSONRequestBody
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(mockSvc *mocks.ProductService)
		expectedStatus int
		expectedBody   func() *product.Product
	}{
		{
			name: "Успешное создание товара",
			args: args{
				request: dto.PostProductsJSONRequestBody{
					PvzId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Type:  dto.PostProductsJSONBodyType("электроника"),
				},
			},
			setupMock: func(mockSvc *mocks.ProductService) {
				expectedProduct := &product.Product{
					ID:          uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
					DateTime:    time.Now(),
					Type:        product.TypeElectronics,
					ReceptionID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}

				mockSvc.On("CreateProduct", mock.Anything,
					uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					product.TypeElectronics).
					Return(expectedProduct, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() *product.Product {
				return &product.Product{
					ID:          uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
					DateTime:    time.Now(),
					Type:        product.TypeElectronics,
					ReceptionID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}
			},
		},
		{
			name: "Нет активной приемки",
			args: args{
				request: dto.PostProductsJSONRequestBody{
					PvzId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Type:  dto.PostProductsJSONBodyType("одежда"),
				},
			},
			setupMock: func(mockSvc *mocks.ProductService) {
				mockSvc.On("CreateProduct", mock.Anything,
					uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					product.TypeClothes).
					Return(nil, handlers.ErrNoActiveReceptionProduct)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name: "Неизвестный тип товара",
			args: args{
				request: dto.PostProductsJSONRequestBody{
					PvzId: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Type:  dto.PostProductsJSONBodyType("мебель"),
				},
			},
			setupMock: func(mockSvc *mocks.ProductService) {
				// Метод не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.ProductService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewProductHandler(mockService, nullLogger)

			requestBody, err := json.Marshal(tt.args.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.CreateProduct(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != nil {
				var responseBody dto.Product
				err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
				require.NoError(t, err)

				expectedBody := tt.expectedBody()

				assert.NotNil(t, responseBody.Id)
				assert.NotNil(t, responseBody.DateTime)
				assert.Equal(t, expectedBody.ReceptionID.String(), responseBody.ReceptionId.String())

				var expectedType dto.ProductType

				switch expectedBody.Type {
				case product.TypeElectronics:
					expectedType = dto.ProductType("электроника")
				case product.TypeClothes:
					expectedType = dto.ProductType("одежда")
				case product.TypeShoes:
					expectedType = dto.ProductType("обувь")
				}

				assert.Equal(t, expectedType, responseBody.Type)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_DeleteLastProduct(t *testing.T) {
	pvzID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		url            string
		setupMock      func(mockSvc *mocks.ProductService)
		expectedStatus int
	}{
		{
			name: "Успешное удаление товара",
			url:  "/pvz/" + pvzID.String() + "/delete_last_product",
			setupMock: func(mockSvc *mocks.ProductService) {
				mockSvc.On("DeleteLastProduct", mock.Anything, pvzID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Нет активной приемки",
			url:  "/pvz/" + pvzID.String() + "/delete_last_product",
			setupMock: func(mockSvc *mocks.ProductService) {
				mockSvc.On("DeleteLastProduct", mock.Anything, pvzID).
					Return(handlers.ErrNoActiveReceptionProduct)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Нет товаров для удаления",
			url:  "/pvz/" + pvzID.String() + "/delete_last_product",
			setupMock: func(mockSvc *mocks.ProductService) {
				mockSvc.On("DeleteLastProduct", mock.Anything, pvzID).
					Return(handlers.ErrNoProductsToDelete)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Неверный URL",
			url:  "/invalid/url",
			setupMock: func(mockSvc *mocks.ProductService) {
				// Метод не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.ProductService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewProductHandler(mockService, nullLogger)

			req, err := http.NewRequest(http.MethodPost, tt.url, http.NoBody)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			handler.DeleteLastProduct(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			mockService.AssertExpectations(t)
		})
	}
}
