package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"avito/internal/application/auth"
	"avito/internal/application/product"
	"avito/internal/application/pvz"
	"avito/internal/application/reception"
)

type Server struct {
	server *http.Server
	logger *slog.Logger
}

func New(
	addr string,
	authSvc *auth.Service,
	pvzSvc *pvz.Service,
	receptionSvc *reception.Service,
	productSvc *product.Service,
	logger *slog.Logger,
) *Server {
	router := NewRouter(authSvc, pvzSvc, receptionSvc, productSvc, logger)

	server := &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           router.Handler(),
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger: logger,
	}

	return server
}

func (s *Server) Start() error {
	s.logger.Info("Запуск HTTP-сервера", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Остановка HTTP-сервера")
	return s.server.Shutdown(ctx)
}
