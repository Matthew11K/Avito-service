package middleware

import (
	"context"
	"net/http"
	"strings"

	"avito/internal/domain/auth"
)

type ContextKey string

const (
	UserIDKey   ContextKey = "user_id"
	UserRoleKey ContextKey = "user_role"
)

type AuthTokenParser interface {
	ParseToken(tokenString string) (userID string, role auth.Role, err error)
}

func RequireAuth(tokenParser AuthTokenParser, logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "отсутствует токен авторизации", nil, logger)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "неверный формат токена", nil, logger)
				return
			}

			userID, role, err := tokenParser.ParseToken(tokenParts[1])
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "невалидный токен", err, logger)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(requiredRole auth.Role, logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(auth.Role)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "отсутствует информация о пользователе", nil, logger)
				return
			}

			if role != requiredRole {
				respondWithError(w, http.StatusForbidden, "недостаточно прав для выполнения операции", nil, logger)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAnyRole(requiredRoles []auth.Role, logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(auth.Role)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "отсутствует информация о пользователе", nil, logger)
				return
			}

			var hasRequiredRole bool

			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				respondWithError(w, http.StatusForbidden, "недостаточно прав для выполнения операции", nil, logger)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
