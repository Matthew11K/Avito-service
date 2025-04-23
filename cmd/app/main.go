package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"avito/internal/config"
	grpcServer "avito/internal/interfaces/grpc"
	httpServer "avito/internal/interfaces/http"
	"avito/internal/metrics"
	"avito/pkg"
	"avito/pkg/txs"

	authRepository "avito/internal/infrastructure/auth"
	productRepository "avito/internal/infrastructure/product"
	pvzRepository "avito/internal/infrastructure/pvz"
	receptionRepository "avito/internal/infrastructure/reception"

	authService "avito/internal/application/auth"
	productService "avito/internal/application/product"
	pvzService "avito/internal/application/pvz"
	receptionService "avito/internal/application/reception"

	"github.com/jackc/pgx/v5/pgxpool"
)

//nolint:funlen // main агрегирует все зависимости и точки входа, разбивать на части нецелесообразно для читаемости
func main() {
	cfg := config.LoadConfig()

	logger := pkg.NewLogger(os.Stdout)
	logger.Info("Запуск приложения")

	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Ошибка при парсинге URL базы данных", "error", err)
		os.Exit(1)
	}

	//nolint:gosec // cfg.DBMaxConn и cfg.DBMinConn всегда валидируются и ограничиваются в config.LoadConfig, переполнение невозможно
	dbConfig.MaxConns = int32(cfg.DBMaxConn)
	dbConfig.MinConns = int32(cfg.DBMinConn)

	db, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		logger.Error("Ошибка при подключении к базе данных", "error", err)
		os.Exit(1)
	}

	defer func() {
		db.Close()
		logger.Info("Подключение к базе данных закрыто")
	}()

	if err := db.Ping(context.Background()); err != nil {
		logger.Error("Ошибка при пинге базы данных", "error", err)
		return
	}

	logger.Info("Подключение к базе данных установлено")

	txManager := txs.NewTxManager(db, logger)

	authRepo := authRepository.NewRepository(db)
	pvzRepo := pvzRepository.NewRepository(db)
	receptionRepo := receptionRepository.NewRepository(db)
	productRepo := productRepository.NewRepository(db)

	authSvc := authService.NewService(authRepo, txManager, cfg.JWTSecret, cfg.TokenTTL)
	pvzSvc := pvzService.NewService(pvzRepo, txManager)
	receptionSvc := receptionService.NewService(receptionRepo, pvzRepo, txManager)
	productSvc := productService.NewService(productRepo, receptionRepo, pvzRepo, txManager)

	prometheusServer := metrics.StartServer(cfg.PrometheusAddr)
	logger.Info("Prometheus metrics доступны", "addr", cfg.PrometheusAddr+"/metrics")

	httpSrv := httpServer.New(
		cfg.HTTPAddr,
		authSvc,
		pvzSvc,
		receptionSvc,
		productSvc,
		logger,
	)

	go func() {
		if err := httpSrv.Start(); err != nil {
			logger.Error("Ошибка при запуске HTTP-сервера", "error", err)
		}
	}()
	logger.Info("HTTP-сервер запущен", "addr", cfg.HTTPAddr)

	grpcSrv := grpcServer.New(pvzSvc, logger)

	go func() {
		if err := grpcSrv.Start(cfg.GRPCAddr); err != nil {
			logger.Error("Ошибка при запуске gRPC-сервера", "error", err)
		}
	}()
	logger.Info("gRPC-сервер запущен", "addr", cfg.GRPCAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Получен сигнал завершения, выполняем graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpSrv.Stop(ctx); err != nil {
		logger.Error("Ошибка при остановке HTTP-сервера", "error", err)
	}

	grpcSrv.Stop()

	if err := prometheusServer.Shutdown(ctx); err != nil {
		logger.Error("Ошибка при остановке Prometheus-сервера", "error", err)
	}

	logger.Info("Приложение остановлено")
}
