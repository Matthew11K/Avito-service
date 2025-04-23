package grpc

import (
	"context"
	"log/slog"
	"net"

	"avito/internal/domain/pvz"
	pbpvz "avito/internal/interfaces/grpc/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type PVZService interface {
	GetPVZByID(ctx context.Context, id string) (*pvz.PVZ, error)
	GetPVZs(ctx context.Context, req pvz.GetPVZsRequest) ([]pvz.WithReceptions, error)
}

type Server struct {
	server     *grpc.Server
	pvzService PVZService
	logger     *slog.Logger
}

func New(domainPVZService DomainPVZService, logger *slog.Logger) *Server {
	pvzService := NewPVZServiceAdapter(domainPVZService)

	server := &Server{
		server:     grpc.NewServer(),
		pvzService: pvzService,
		logger:     logger,
	}

	return server
}

func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	pbpvz.RegisterPVZServiceServer(s.server, &pvzServiceServer{
		pvzService: s.pvzService,
		logger:     s.logger,
	})

	reflection.Register(s.server)

	s.logger.Info("Запуск gRPC-сервера", "addr", addr)

	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.logger.Info("Остановка gRPC-сервера")
	s.server.GracefulStop()
}
