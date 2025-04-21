package auth

import (
	"context"
	"fmt"
	"time"

	"avito/internal/domain/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Transactor interface {
	WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type Repository interface {
	CreateUser(ctx context.Context, email string, passwordHash string, role auth.Role) (*auth.User, error)
	GetUserByEmail(ctx context.Context, email string) (*auth.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error)
}

type Service struct {
	repo       Repository
	txManager  Transactor
	signingKey []byte
	tokenTTL   time.Duration
}

func NewService(repo Repository, txManager Transactor, signingKey string, tokenTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		txManager:  txManager,
		signingKey: []byte(signingKey),
		tokenTTL:   tokenTTL,
	}
}

func (s *Service) Register(ctx context.Context, req auth.RegisterRequest) (*auth.User, error) {
	if req.Email == "" {
		return nil, &auth.ErrEmailEmpty{}
	}

	if req.Password == "" {
		return nil, &auth.ErrPasswordEmpty{}
	}

	if req.Role == "" {
		return nil, &auth.ErrRoleEmpty{}
	}

	if req.Role != auth.RoleEmployee && req.Role != auth.RoleModerator {
		return nil, &auth.ErrInvalidRole{}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	var user *auth.User

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		existingUser, err := s.repo.GetUserByEmail(txCtx, req.Email)
		if err == nil && existingUser != nil {
			return &auth.ErrUserAlreadyExists{}
		}

		if err != nil && !isErrUserNotFound(err) {
			return fmt.Errorf("ошибка при проверке пользователя: %w", err)
		}

		user, err = s.repo.CreateUser(txCtx, req.Email, string(hashedPassword), req.Role)

		return err
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func isErrUserNotFound(err error) bool {
	_, ok := err.(*auth.ErrUserNotFound)
	return ok
}

func (s *Service) Login(ctx context.Context, req auth.LoginRequest) (*auth.Auth, error) {
	if req.Email == "" {
		return nil, &auth.ErrEmailEmpty{}
	}

	if req.Password == "" {
		return nil, &auth.ErrPasswordEmpty{}
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, &auth.ErrInvalidCredentials{}
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации токена: %w", err)
	}

	return &auth.Auth{Token: token}, nil
}

func (s *Service) DummyLogin(_ context.Context, req auth.DummyLoginRequest) (*auth.Auth, error) {
	if req.Role == "" {
		return nil, &auth.ErrRoleEmpty{}
	}

	if req.Role != auth.RoleEmployee && req.Role != auth.RoleModerator {
		return nil, &auth.ErrInvalidRole{}
	}

	dummyUser := &auth.User{
		ID:   uuid.New(),
		Role: req.Role,
	}

	token, err := s.generateJWT(dummyUser)
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации токена: %w", err)
	}

	return &auth.Auth{Token: token}, nil
}

func (s *Service) generateJWT(user *auth.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     time.Now().Add(s.tokenTTL).Unix(),
	})

	return token.SignedString(s.signingKey)
}

func (s *Service) ParseToken(tokenString string) (uuid.UUID, auth.Role, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}

		return s.signingKey, nil
	})

	if err != nil {
		return uuid.Nil, "", fmt.Errorf("ошибка при парсинге токена: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", fmt.Errorf("невалидный токен")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, "", fmt.Errorf("отсутствует ID пользователя")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("невалидный ID пользователя")
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return uuid.Nil, "", fmt.Errorf("отсутствует роль пользователя")
	}

	role := auth.Role(roleStr)

	return userID, role, nil
}
