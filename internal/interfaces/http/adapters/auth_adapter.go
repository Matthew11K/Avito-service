package adapters

import (
	"context"
	"errors"

	appAuth "avito/internal/application/auth"
	"avito/internal/domain/auth"
	"avito/internal/interfaces/http/handlers"
)

type AuthServiceAdapter struct {
	service *appAuth.Service
}

func NewAuthServiceAdapter(service *appAuth.Service) *AuthServiceAdapter {
	return &AuthServiceAdapter{
		service: service,
	}
}

func (a *AuthServiceAdapter) Register(ctx context.Context, email, password string, role auth.Role) (*auth.User, error) {
	req := auth.RegisterRequest{
		Email:    email,
		Password: password,
		Role:     role,
	}

	user, err := a.service.Register(ctx, req)
	if err != nil {
		var userExistsErr *auth.ErrUserAlreadyExists
		if errors.As(err, &userExistsErr) {
			return nil, handlers.ErrEmailAlreadyExists
		}

		return nil, err
	}

	return user, nil
}

func (a *AuthServiceAdapter) Login(ctx context.Context, email, password string) (string, error) {
	req := auth.LoginRequest{
		Email:    email,
		Password: password,
	}

	authResult, err := a.service.Login(ctx, req)
	if err != nil {
		var invalidCredentialsErr *auth.ErrInvalidCredentials
		if errors.As(err, &invalidCredentialsErr) {
			return "", handlers.ErrInvalidCredentials
		}

		return "", err
	}

	return authResult.Token, nil
}

func (a *AuthServiceAdapter) GenerateDummyToken(ctx context.Context, role auth.Role) (string, error) {
	req := auth.DummyLoginRequest{
		Role: role,
	}

	authResult, err := a.service.DummyLogin(ctx, req)
	if err != nil {
		return "", err
	}

	return authResult.Token, nil
}
