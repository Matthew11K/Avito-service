package auth_test

import (
	"context"
	"testing"
	"time"

	"avito/internal/application/auth"
	"avito/internal/application/auth/mocks"
	domainAuth "avito/internal/domain/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Register(t *testing.T) {
	tests := []struct {
		name          string
		request       domainAuth.RegisterRequest
		mockSetup     func(*mocks.Repository, *mocks.Transactor)
		expectedUser  *domainAuth.User
		expectedError error
	}{
		{
			name: "Успешная регистрация",
			request: domainAuth.RegisterRequest{
				Email:    "test@example.com",
				Password: "password",
				Role:     domainAuth.RoleEmployee,
			},
			mockSetup: func(repo *mocks.Repository, tx *mocks.Transactor) {
				userID := uuid.New()
				expectedUser := &domainAuth.User{
					ID:           userID,
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					Role:         domainAuth.RoleEmployee,
				}

				repo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(nil, &domainAuth.ErrUserNotFound{})

				repo.On("CreateUser", mock.Anything, "test@example.com", mock.AnythingOfType("string"), domainAuth.RoleEmployee).
					Return(expectedUser, nil)

				tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedUser: &domainAuth.User{
				Email: "test@example.com",
				Role:  domainAuth.RoleEmployee,
			},
			expectedError: nil,
		},
		{
			name: "Пустой email",
			request: domainAuth.RegisterRequest{
				Email:    "",
				Password: "password",
				Role:     domainAuth.RoleEmployee,
			},
			mockSetup: func(repo *mocks.Repository, tx *mocks.Transactor) {
				// Моки не должны вызываться
			},
			expectedUser:  nil,
			expectedError: &domainAuth.ErrEmailEmpty{},
		},
		{
			name: "Пустой пароль",
			request: domainAuth.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
				Role:     domainAuth.RoleEmployee,
			},
			mockSetup: func(repo *mocks.Repository, tx *mocks.Transactor) {
				// Моки не должны вызываться
			},
			expectedUser:  nil,
			expectedError: &domainAuth.ErrPasswordEmpty{},
		},
		{
			name: "Пустая роль",
			request: domainAuth.RegisterRequest{
				Email:    "test@example.com",
				Password: "password",
				Role:     "",
			},
			mockSetup: func(repo *mocks.Repository, tx *mocks.Transactor) {
				// Моки не должны вызываться
			},
			expectedUser:  nil,
			expectedError: &domainAuth.ErrRoleEmpty{},
		},
		{
			name: "Неверная роль",
			request: domainAuth.RegisterRequest{
				Email:    "test@example.com",
				Password: "password",
				Role:     "invalid_role",
			},
			mockSetup: func(repo *mocks.Repository, tx *mocks.Transactor) {
				// Моки не должны вызываться
			},
			expectedUser:  nil,
			expectedError: &domainAuth.ErrInvalidRole{},
		},
		{
			name: "Пользователь уже существует",
			request: domainAuth.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password",
				Role:     domainAuth.RoleEmployee,
			},
			mockSetup: func(repo *mocks.Repository, tx *mocks.Transactor) {
				existingUser := &domainAuth.User{
					ID:           uuid.New(),
					Email:        "existing@example.com",
					PasswordHash: "hashed_password",
					Role:         domainAuth.RoleEmployee,
				}

				repo.On("GetUserByEmail", mock.Anything, "existing@example.com").
					Return(existingUser, nil)

				tx.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(&domainAuth.ErrUserAlreadyExists{}).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					})
			},
			expectedUser:  nil,
			expectedError: &domainAuth.ErrUserAlreadyExists{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockTx)
			}

			service := auth.NewService(mockRepo, mockTx, "test_secret_key", 24*time.Hour)

			actualUser, err := service.Register(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, actualUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualUser)

				if tt.expectedUser != nil {
					assert.Equal(t, tt.expectedUser.Email, actualUser.Email)
					assert.Equal(t, tt.expectedUser.Role, actualUser.Role)
				}
			}

			mockRepo.AssertExpectations(t)
			mockTx.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name          string
		request       domainAuth.LoginRequest
		mockSetup     func(*mocks.Repository)
		expectedAuth  *domainAuth.Auth
		expectedError error
	}{
		{
			name: "Пустой email",
			request: domainAuth.LoginRequest{
				Email:    "",
				Password: "password",
			},
			mockSetup: func(repo *mocks.Repository) {
				// Моки не должны вызываться
			},
			expectedAuth:  nil,
			expectedError: &domainAuth.ErrEmailEmpty{},
		},
		{
			name: "Пустой пароль",
			request: domainAuth.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			mockSetup: func(repo *mocks.Repository) {
				// Моки не должны вызываться
			},
			expectedAuth:  nil,
			expectedError: &domainAuth.ErrPasswordEmpty{},
		},
		{
			name: "Пользователь не найден",
			request: domainAuth.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password",
			},
			mockSetup: func(repo *mocks.Repository) {
				repo.On("GetUserByEmail", mock.Anything, "notfound@example.com").
					Return(nil, &domainAuth.ErrUserNotFound{})
			},
			expectedAuth:  nil,
			expectedError: &domainAuth.ErrUserNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockTx := new(mocks.Transactor)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := auth.NewService(mockRepo, mockTx, "test_secret_key", 24*time.Hour)

			actualAuth, err := service.Login(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, actualAuth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualAuth)
				assert.NotEmpty(t, actualAuth.Token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_DummyLogin(t *testing.T) {
	tests := []struct {
		name          string
		request       domainAuth.DummyLoginRequest
		expectedError error
	}{
		{
			name: "Успешное получение токена для сотрудника",
			request: domainAuth.DummyLoginRequest{
				Role: domainAuth.RoleEmployee,
			},
			expectedError: nil,
		},
		{
			name: "Успешное получение токена для модератора",
			request: domainAuth.DummyLoginRequest{
				Role: domainAuth.RoleModerator,
			},
			expectedError: nil,
		},
		{
			name: "Пустая роль",
			request: domainAuth.DummyLoginRequest{
				Role: "",
			},
			expectedError: &domainAuth.ErrRoleEmpty{},
		},
		{
			name: "Неверная роль",
			request: domainAuth.DummyLoginRequest{
				Role: "invalid_role",
			},
			expectedError: &domainAuth.ErrInvalidRole{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockTx := new(mocks.Transactor)

			service := auth.NewService(mockRepo, mockTx, "test_secret_key", 24*time.Hour)

			actualAuth, err := service.DummyLogin(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, actualAuth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualAuth)
				assert.NotEmpty(t, actualAuth.Token)

				userID, role, err := service.ParseToken(actualAuth.Token)
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
				assert.Equal(t, tt.request.Role, role)
			}
		})
	}
}

func TestService_ParseToken(t *testing.T) {
	mockRepo := new(mocks.Repository)
	mockTx := new(mocks.Transactor)

	service := auth.NewService(mockRepo, mockTx, "test_secret_key", 24*time.Hour)

	validToken, err := service.DummyLogin(context.Background(), domainAuth.DummyLoginRequest{
		Role: domainAuth.RoleEmployee,
	})
	require.NoError(t, err)
	require.NotNil(t, validToken)

	tests := []struct {
		name           string
		token          string
		expectedUserID uuid.UUID
		expectedRole   domainAuth.Role
		expectedError  bool
	}{
		{
			name:           "Валидный токен",
			token:          validToken.Token,
			expectedUserID: uuid.Nil,
			expectedRole:   domainAuth.RoleEmployee,
			expectedError:  false,
		},
		{
			name:           "Невалидный токен",
			token:          "invalid.token.string",
			expectedUserID: uuid.Nil,
			expectedRole:   "",
			expectedError:  true,
		},
		{
			name:           "Пустой токен",
			token:          "",
			expectedUserID: uuid.Nil,
			expectedRole:   "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, role, err := service.ParseToken(tt.token)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.expectedUserID == uuid.Nil {
					assert.NotEqual(t, uuid.Nil, userID)
				} else {
					assert.Equal(t, tt.expectedUserID, userID)
				}

				assert.Equal(t, tt.expectedRole, role)
			}
		})
	}
}
