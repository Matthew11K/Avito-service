package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"avito/internal/domain/auth"
	"avito/internal/interfaces/http/dto"

	"log/slog"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("пользователь с таким email уже существует")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
)

type AuthService interface {
	Register(ctx context.Context, email, password string, role auth.Role) (*auth.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GenerateDummyToken(ctx context.Context, role auth.Role) (string, error)
}

type AuthHandler struct {
	service AuthService
	logger  *slog.Logger
}

func NewAuthHandler(service AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	var req dto.PostDummyLoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса", err, h.logger)
		return
	}

	var role auth.Role

	switch req.Role {
	case dto.PostDummyLoginJSONBodyRoleEmployee:
		role = auth.RoleEmployee
	case dto.PostDummyLoginJSONBodyRoleModerator:
		role = auth.RoleModerator
	default:
		respondWithError(w, http.StatusBadRequest, "неизвестная роль", nil, h.logger)
		return
	}

	token, err := h.service.GenerateDummyToken(r.Context(), role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "ошибка при генерации токена", err, h.logger)
		return
	}

	respondWithJSON(w, http.StatusOK, token)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	var req dto.PostRegisterJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса", err, h.logger)
		return
	}

	var role auth.Role

	switch req.Role {
	case dto.Employee:
		role = auth.RoleEmployee
	case dto.Moderator:
		role = auth.RoleModerator
	default:
		respondWithError(w, http.StatusBadRequest, "неизвестная роль", nil, h.logger)
		return
	}

	user, err := h.service.Register(r.Context(), string(req.Email), req.Password, role)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			respondWithError(w, http.StatusBadRequest, "пользователь с таким email уже существует", err, h.logger)
		default:
			respondWithError(w, http.StatusInternalServerError, "ошибка при регистрации пользователя", err, h.logger)
		}

		return
	}

	userID, _ := uuid.Parse(user.ID.String())
	respUser := dto.User{
		Id:    &userID,
		Email: req.Email,
	}

	switch user.Role {
	case auth.RoleEmployee:
		respUser.Role = dto.UserRoleEmployee
	case auth.RoleModerator:
		respUser.Role = dto.UserRoleModerator
	}

	respondWithJSON(w, http.StatusCreated, respUser)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	var req dto.PostLoginJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса", err, h.logger)
		return
	}

	token, err := h.service.Login(r.Context(), string(req.Email), req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			respondWithError(w, http.StatusUnauthorized, "неверные учетные данные", err, h.logger)
		default:
			respondWithError(w, http.StatusInternalServerError, "ошибка при авторизации", err, h.logger)
		}

		return
	}

	respondWithJSON(w, http.StatusOK, token)
}
