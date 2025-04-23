package http

import (
	"log/slog"
	"net/http"
	"strings"

	"avito/internal/application/auth"
	"avito/internal/application/product"
	"avito/internal/application/pvz"
	"avito/internal/application/reception"
	domainAuth "avito/internal/domain/auth"
	"avito/internal/interfaces/http/adapters"
	"avito/internal/interfaces/http/handlers"
	"avito/internal/interfaces/http/middleware"
)

type Router struct {
	handler http.Handler
	logger  *slog.Logger
}

//nolint:funlen // функция инициализации роутера должна быть целостной для наглядности и поддержки
func NewRouter(
	authSvc *auth.Service,
	pvzSvc *pvz.Service,
	receptionSvc *reception.Service,
	productSvc *product.Service,
	logger *slog.Logger,
) *Router {
	router := &Router{
		logger: logger,
	}

	authAdapter := adapters.NewAuthServiceAdapter(authSvc)
	pvzAdapter := adapters.NewPVZServiceAdapter(pvzSvc)
	receptionAdapter := adapters.NewReceptionServiceAdapter(receptionSvc)
	productAdapter := adapters.NewProductServiceAdapter(productSvc)

	authHandler := handlers.NewAuthHandler(authAdapter, logger)
	pvzHandler := handlers.NewPVZHandler(pvzAdapter, logger)
	receptionHandler := handlers.NewReceptionHandler(receptionAdapter, logger)
	productHandler := handlers.NewProductHandler(productAdapter, logger)

	publicMux := http.NewServeMux()

	publicMux.HandleFunc("/dummyLogin", authHandler.DummyLogin)
	publicMux.HandleFunc("/login", authHandler.Login)
	publicMux.HandleFunc("/register", authHandler.Register)

	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("/pvz", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pvz" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			pvzHandler.GetPVZs(w, r)
		case http.MethodPost:
			role, ok := r.Context().Value(middleware.UserRoleKey).(domainAuth.Role)
			if !ok || role != domainAuth.RoleModerator {
				middleware.RespondWithError(w, http.StatusForbidden, "недостаточно прав для выполнения операции", nil, logger)
				return
			}

			pvzHandler.CreatePVZ(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	protectedMux.HandleFunc("/pvz/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		role, ok := r.Context().Value(middleware.UserRoleKey).(domainAuth.Role)
		if !ok || role != domainAuth.RoleEmployee {
			middleware.RespondWithError(w, http.StatusForbidden, "недостаточно прав для выполнения операции", nil, logger)
			return
		}

		if strings.HasSuffix(path, "/close_last_reception") {
			receptionHandler.CloseLastReception(w, r)
			return
		}

		if strings.HasSuffix(path, "/delete_last_product") {
			productHandler.DeleteLastProduct(w, r)
			return
		}

		http.NotFound(w, r)
	})

	protectedMux.HandleFunc("/receptions", func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(middleware.UserRoleKey).(domainAuth.Role)
		if !ok || role != domainAuth.RoleEmployee {
			middleware.RespondWithError(w, http.StatusForbidden, "недостаточно прав для выполнения операции", nil, logger)
			return
		}

		receptionHandler.CreateReception(w, r)
	})

	protectedMux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(middleware.UserRoleKey).(domainAuth.Role)
		if !ok || role != domainAuth.RoleEmployee {
			middleware.RespondWithError(w, http.StatusForbidden, "недостаточно прав для выполнения операции", nil, logger)
			return
		}

		productHandler.CreateProduct(w, r)
	})

	loggerMiddleware := middleware.RequestLogging(logger)
	recoveryMiddleware := middleware.Recovery(logger)
	metricsMiddleware := middleware.Metrics()

	tokenParser := &tokenParser{
		authService: authSvc,
	}

	authMiddleware := middleware.RequireAuth(tokenParser, logger)
	protectedHandler := authMiddleware(protectedMux)

	finalMux := http.NewServeMux()

	finalMux.Handle("/dummyLogin", publicMux)
	finalMux.Handle("/login", publicMux)
	finalMux.Handle("/register", publicMux)

	finalMux.Handle("/pvz", protectedHandler)
	finalMux.Handle("/pvz/", protectedHandler)
	finalMux.Handle("/receptions", protectedHandler)
	finalMux.Handle("/products", protectedHandler)

	handler := loggerMiddleware(metricsMiddleware(recoveryMiddleware(finalMux)))

	router.handler = handler

	return router
}

func (r *Router) Handler() http.Handler {
	return r.handler
}

type tokenParser struct {
	authService *auth.Service
}

func (tp *tokenParser) ParseToken(tokenString string) (string, domainAuth.Role, error) {
	userID, role, err := tp.authService.ParseToken(tokenString)
	if err != nil {
		return "", "", err
	}

	return userID.String(), role, nil
}
