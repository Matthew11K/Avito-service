package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"avito/internal/domain/reception"
	"avito/internal/interfaces/http/dto"
	"avito/internal/metrics"

	"log/slog"

	"github.com/google/uuid"
)

var (
	ErrActiveReceptionExists = errors.New("уже есть незакрытая приемка")
	ErrNoActiveReception     = errors.New("нет открытых приемок")
)

type ReceptionService interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error)
}

type ReceptionHandler struct {
	service ReceptionService
	logger  *slog.Logger
}

func NewReceptionHandler(service ReceptionService, logger *slog.Logger) *ReceptionHandler {
	return &ReceptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReceptionHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	var req dto.PostReceptionsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса", err, h.logger)
		return
	}

	pvzID, err := uuid.Parse(req.PvzId.String())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат UUID", err, h.logger)
		return
	}

	rec, err := h.service.CreateReception(r.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, ErrActiveReceptionExists):
			respondWithError(w, http.StatusBadRequest, "уже есть незакрытая приемка", err, h.logger)
		default:
			respondWithError(w, http.StatusInternalServerError, "ошибка при создании приемки", err, h.logger)
		}

		return
	}

	metrics.ReceptionsCreatedTotal.Inc()

	recID, _ := uuid.Parse(rec.ID.String())
	pvzIDParsed, _ := uuid.Parse(rec.PVZID.String())

	response := dto.Reception{
		Id:       &recID,
		DateTime: rec.DateTime,
		PvzId:    pvzIDParsed,
		Status:   dto.InProgress,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *ReceptionHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	path := r.URL.Path
	h.logger.Info("Получен запрос на закрытие приемки", "path", path)

	trimmedPath := strings.Trim(path, "/")
	parts := strings.Split(trimmedPath, "/")

	h.logger.Info("Разбор пути", "parts", parts, "count", len(parts))

	if len(parts) != 3 || parts[0] != "pvz" || parts[2] != "close_last_reception" {
		respondWithError(w, http.StatusBadRequest, "неверный URL", nil, h.logger)
		return
	}

	pvzIDStr := parts[1]
	h.logger.Info("Извлечен ID PVZ", "pvzID", pvzIDStr)

	pvzID, err := uuid.Parse(pvzIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат UUID", err, h.logger)
		return
	}

	rec, err := h.service.CloseLastReception(r.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoActiveReception):
			respondWithError(w, http.StatusBadRequest, "нет открытых приемок", err, h.logger)
		default:
			respondWithError(w, http.StatusInternalServerError, "ошибка при закрытии приемки", err, h.logger)
		}

		return
	}

	recID, _ := uuid.Parse(rec.ID.String())
	pvzIDParsed, _ := uuid.Parse(rec.PVZID.String())

	response := dto.Reception{
		Id:       &recID,
		DateTime: rec.DateTime,
		PvzId:    pvzIDParsed,
		Status:   dto.Close,
	}

	respondWithJSON(w, http.StatusOK, response)
}
