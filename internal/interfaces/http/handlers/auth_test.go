//nolint:revive // структура теста требует неиспользуемых параметров для поддержания единообразия
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito/internal/domain/auth"
	"avito/internal/interfaces/http/dto"
	"avito/internal/interfaces/http/handlers"
	"avito/internal/interfaces/http/handlers/mocks"

	"log/slog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Register(t *testing.T) {
	type args struct {
		request dto.PostRegisterJSONRequestBody
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(mockSvc *mocks.AuthService)
		expectedStatus int
		expectedBody   func() *auth.User
	}{
		{
			name: "Успешная регистрация",
			args: args{
				request: dto.PostRegisterJSONRequestBody{
					Email:    "test@example.com",
					Password: "password123",
					Role:     dto.Employee,
				},
			},
			setupMock: func(mockSvc *mocks.AuthService) {
				userID := uuid.New()
				user := &auth.User{
					ID:           userID,
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					Role:         auth.RoleEmployee,
				}

				mockSvc.On("Register", mock.Anything, "test@example.com", "password123", auth.RoleEmployee).
					Return(user, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() *auth.User {
				userID := uuid.New()
				return &auth.User{
					ID:           userID,
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					Role:         auth.RoleEmployee,
				}
			},
		},
		{
			name: "Пользователь уже существует",
			args: args{
				request: dto.PostRegisterJSONRequestBody{
					Email:    "existing@example.com",
					Password: "password123",
					Role:     dto.Employee,
				},
			},
			setupMock: func(mockSvc *mocks.AuthService) {
				mockSvc.On("Register", mock.Anything, "existing@example.com", "password123", auth.RoleEmployee).
					Return(nil, handlers.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name: "Некорректная роль пользователя",
			args: args{
				request: dto.PostRegisterJSONRequestBody{
					Email:    "test@example.com",
					Password: "password123",
					Role:     "invalid_role",
				},
			},
			setupMock:      func(mockSvc *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.AuthService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewAuthHandler(mockService, nullLogger)

			requestBody, err := json.Marshal(tt.args.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.Register(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != nil {
				var responseBody dto.User
				err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
				require.NoError(t, err)

				assert.NotNil(t, responseBody.Id)
				assert.Equal(t, tt.args.request.Email, responseBody.Email)

				var expectedRole dto.UserRole

				switch tt.expectedBody().Role {
				case auth.RoleEmployee:
					expectedRole = dto.UserRoleEmployee
				case auth.RoleModerator:
					expectedRole = dto.UserRoleModerator
				}

				assert.Equal(t, expectedRole, responseBody.Role)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	type args struct {
		request dto.PostLoginJSONRequestBody
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(mockSvc *mocks.AuthService)
		expectedStatus int
		expectedToken  string
	}{
		{
			name: "Успешная авторизация",
			args: args{
				request: dto.PostLoginJSONRequestBody{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			setupMock: func(mockSvc *mocks.AuthService) {
				mockSvc.On("Login", mock.Anything, "test@example.com", "password123").
					Return("test.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "test.jwt.token",
		},
		{
			name: "Неверные учетные данные",
			args: args{
				request: dto.PostLoginJSONRequestBody{
					Email:    "wrong@example.com",
					Password: "wrongpass",
				},
			},
			setupMock: func(mockSvc *mocks.AuthService) {
				mockSvc.On("Login", mock.Anything, "wrong@example.com", "wrongpass").
					Return("", handlers.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.AuthService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewAuthHandler(mockService, nullLogger)

			requestBody, err := json.Marshal(tt.args.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.Login(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedToken != "" {
				var token string
				err = json.Unmarshal(recorder.Body.Bytes(), &token)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_DummyLogin(t *testing.T) {
	type args struct {
		request dto.PostDummyLoginJSONRequestBody
	}

	tests := []struct {
		name           string
		args           args
		setupMock      func(mockSvc *mocks.AuthService)
		expectedStatus int
		expectedToken  string
	}{
		{
			name: "Успешное получение токена для сотрудника",
			args: args{
				request: dto.PostDummyLoginJSONRequestBody{
					Role: dto.PostDummyLoginJSONBodyRoleEmployee,
				},
			},
			setupMock: func(mockSvc *mocks.AuthService) {
				mockSvc.On("GenerateDummyToken", mock.Anything, auth.RoleEmployee).
					Return("employee.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "employee.jwt.token",
		},
		{
			name: "Успешное получение токена для модератора",
			args: args{
				request: dto.PostDummyLoginJSONRequestBody{
					Role: dto.PostDummyLoginJSONBodyRoleModerator,
				},
			},
			setupMock: func(mockSvc *mocks.AuthService) {
				mockSvc.On("GenerateDummyToken", mock.Anything, auth.RoleModerator).
					Return("moderator.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "moderator.jwt.token",
		},
		{
			name: "Некорректная роль",
			args: args{
				request: dto.PostDummyLoginJSONRequestBody{
					Role: "invalid_role",
				},
			},
			setupMock:      func(mockSvc *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.AuthService)
			tt.setupMock(mockService)

			nullLogger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			handler := handlers.NewAuthHandler(mockService, nullLogger)

			requestBody, err := json.Marshal(tt.args.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			handler.DummyLogin(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedToken != "" {
				var token string
				err = json.Unmarshal(recorder.Body.Bytes(), &token)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}

			mockService.AssertExpectations(t)
		})
	}
}
