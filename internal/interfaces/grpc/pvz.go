package grpc

import (
	"context"
	"log/slog"
	"time"

	domainPVZ "avito/internal/domain/pvz"
	pbpvz "avito/internal/interfaces/grpc/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type pvzServiceServer struct {
	pbpvz.UnimplementedPVZServiceServer
	pvzService PVZService
	logger     *slog.Logger
}

func (s *pvzServiceServer) GetPVZList(ctx context.Context, _ *pbpvz.GetPVZListRequest) (*pbpvz.GetPVZListResponse, error) {
	pvzReq := domainPVZ.GetPVZsRequest{
		Page:  1,
		Limit: 100,
	}

	pvzList, err := s.pvzService.GetPVZs(ctx, pvzReq)
	if err != nil {
		s.logger.Error("Ошибка при получении списка ПВЗ", "error", err)
		return nil, err
	}

	response := &pbpvz.GetPVZListResponse{
		Pvzs: make([]*pbpvz.PVZ, 0, len(pvzList)),
	}

	for _, item := range pvzList {
		response.Pvzs = append(response.Pvzs, &pbpvz.PVZ{
			Id:               item.PVZ.ID.String(),
			RegistrationDate: timestampFromTime(item.PVZ.RegistrationDate),
			City:             string(item.PVZ.City),
		})
	}

	return response, nil
}

func timestampFromTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
