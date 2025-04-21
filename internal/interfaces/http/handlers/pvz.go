package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"avito/internal/domain/pvz"
	"avito/internal/interfaces/http/dto"
	"avito/internal/metrics"

	"log/slog"

	"github.com/google/uuid"
)

type PVZService interface {
	CreatePVZ(ctx context.Context, req pvz.CreatePVZRequest) (*pvz.PVZ, error)
	GetPVZs(ctx context.Context, req pvz.GetPVZsRequest) ([]pvz.PVZWithReceptions, error)
}

type PVZHandler struct {
	service PVZService
	logger  *slog.Logger
}

func NewPVZHandler(service PVZService, logger *slog.Logger) *PVZHandler {
	return &PVZHandler{
		service: service,
		logger:  logger,
	}
}

func (h *PVZHandler) CreatePVZ(w http.ResponseWriter, r *http.Request) {
	var req dto.PostPvzJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса", err, h.logger)
		return
	}

	var city pvz.City

	switch req.City {
	case dto.PVZCity("Москва"):
		city = pvz.CityMoscow
	case dto.PVZCity("Санкт-Петербург"):
		city = pvz.CitySaintPetersburg
	case dto.PVZCity("Казань"):
		city = pvz.CityKazan
	default:
		respondWithError(w, http.StatusBadRequest, "В этом городе нельзя открыть ПВЗ", nil, h.logger)
		return
	}

	createReq := pvz.CreatePVZRequest{
		City: city,
	}

	newPVZ, err := h.service.CreatePVZ(r.Context(), createReq)
	if err != nil {
		h.logger.Error("Ошибка при создании ПВЗ", "error", err)
		respondWithError(w, http.StatusBadRequest, "Неверный запрос", err, h.logger)

		return
	}

	metrics.PVZCreatedTotal.Inc()

	id, _ := uuid.Parse(newPVZ.ID.String())
	response := dto.PVZ{
		Id:               &id,
		RegistrationDate: &newPVZ.RegistrationDate,
	}

	switch newPVZ.City {
	case pvz.CityMoscow:
		response.City = dto.PVZCity("Москва")
	case pvz.CitySaintPetersburg:
		response.City = dto.PVZCity("Санкт-Петербург")
	case pvz.CityKazan:
		response.City = dto.PVZCity("Казань")
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *PVZHandler) GetPVZs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var startDate, endDate *time.Time

	if sd := query.Get("startDate"); sd != "" {
		parsedDate, err := time.Parse(time.RFC3339, sd)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "неверный формат начальной даты", err, h.logger)
			return
		}

		startDate = &parsedDate
	}

	if ed := query.Get("endDate"); ed != "" {
		parsedDate, err := time.Parse(time.RFC3339, ed)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "неверный формат конечной даты", err, h.logger)
			return
		}

		endDate = &parsedDate
	}

	page := 1

	if p := query.Get("page"); p != "" {
		parsedPage, err := strconv.Atoi(p)
		if err != nil || parsedPage < 1 {
			respondWithError(w, http.StatusBadRequest, "неверный параметр page", err, h.logger)
			return
		}

		page = parsedPage
	}

	limit := 10

	if l := query.Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil || parsedLimit < 1 || parsedLimit > 30 {
			respondWithError(w, http.StatusBadRequest, "неверный параметр limit", err, h.logger)
			return
		}

		limit = parsedLimit
	}

	requestedCity := query.Get("city")

	req := pvz.GetPVZsRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
	}

	if requestedCity != "" {
		c := pvz.City(requestedCity)
		if !c.Validate() {
			respondWithError(w, http.StatusBadRequest, "неверный параметр city", nil, h.logger)
			return
		}

		req.City = &c
	}

	pvzList, err := h.service.GetPVZs(r.Context(), req)
	if err != nil {
		h.logger.Error("Ошибка при получении списка ПВЗ", "error", err)
		respondWithError(w, http.StatusInternalServerError, "ошибка при получении списка ПВЗ", err, h.logger)

		return
	}

	filteredList := pvzList
	if requestedCity != "" {
		filteredList = make([]pvz.PVZWithReceptions, 0)

		for _, item := range pvzList {
			if string(item.PVZ.City) == requestedCity {
				filteredList = append(filteredList, item)
			}
		}
	}

	response := make([]map[string]interface{}, 0, len(filteredList))

	for _, p := range filteredList {
		receptions := make([]map[string]interface{}, 0, len(p.Receptions))

		for _, r := range p.Receptions {
			products := make([]dto.Product, 0, len(r.Products))

			for _, pr := range r.Products {
				prID, _ := uuid.Parse(pr.ID.String())
				recID, _ := uuid.Parse(pr.ReceptionID.String())

				var productType dto.ProductType

				switch pr.Type {
				case "электроника":
					productType = dto.ProductType("электроника")
				case "одежда":
					productType = dto.ProductType("одежда")
				case "обувь":
					productType = dto.ProductType("обувь")
				}

				products = append(products, dto.Product{
					Id:          &prID,
					DateTime:    &pr.DateTime,
					Type:        productType,
					ReceptionId: recID,
				})
			}

			recID, _ := uuid.Parse(r.Reception.ID.String())
			pvzID, _ := uuid.Parse(r.Reception.PVZID.String())

			var status dto.ReceptionStatus

			switch r.Reception.Status {
			case "in_progress":
				status = dto.InProgress
			case "close":
				status = dto.Close
			}

			recDTO := map[string]interface{}{
				"reception": dto.Reception{
					Id:       &recID,
					DateTime: r.Reception.DateTime,
					PvzId:    pvzID,
					Status:   status,
				},
				"products": products,
			}

			receptions = append(receptions, recDTO)
		}

		pvzID, _ := uuid.Parse(p.PVZ.ID.String())

		var city dto.PVZCity

		switch p.PVZ.City {
		case pvz.CityMoscow:
			city = dto.PVZCity("Москва")
		case pvz.CitySaintPetersburg:
			city = dto.PVZCity("Санкт-Петербург")
		case pvz.CityKazan:
			city = dto.PVZCity("Казань")
		}

		pvzDTO := map[string]interface{}{
			"pvz": dto.PVZ{
				Id:               &pvzID,
				RegistrationDate: &p.PVZ.RegistrationDate,
				City:             city,
			},
			"receptions": receptions,
		}

		response = append(response, pvzDTO)
	}

	respondWithJSON(w, http.StatusOK, response)
}
