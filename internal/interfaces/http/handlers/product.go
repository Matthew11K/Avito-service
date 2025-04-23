package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"avito/internal/domain/product"
	"avito/internal/interfaces/http/dto"
	"avito/internal/metrics"

	"log/slog"

	"github.com/google/uuid"
)

var (
	ErrNoActiveReceptionProduct  = errors.New("нет активной приемки")
	ErrNoProductsToDelete        = errors.New("нет товаров для удаления")
	ErrReceptionClosedForProduct = errors.New("приемка уже закрыта")
)

type ProductService interface {
	CreateProduct(ctx context.Context, pvzID uuid.UUID, productType product.Type) (*product.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type ProductHandler struct {
	service ProductService
	logger  *slog.Logger
}

func NewProductHandler(service ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	var req dto.PostProductsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса", err, h.logger)
		return
	}

	pvzID, err := uuid.Parse(req.PvzId.String())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат UUID для pvzId", err, h.logger)
		return
	}

	var productType product.Type

	//nolint:exhaustive // Обрабатываем только известные типы товаров
	switch req.Type {
	case dto.PostProductsJSONBodyType("электроника"):
		productType = product.TypeElectronics
	case dto.PostProductsJSONBodyType("одежда"):
		productType = product.TypeClothes
	case dto.PostProductsJSONBodyType("обувь"):
		productType = product.TypeShoes
	default:
		respondWithError(w, http.StatusBadRequest, "неизвестный тип товара", nil, h.logger)
		return
	}

	newProduct, err := h.service.CreateProduct(r.Context(), pvzID, productType)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoActiveReceptionProduct):
			respondWithError(w, http.StatusBadRequest, "нет активной приемки", err, h.logger)
		case errors.Is(err, ErrReceptionClosedForProduct):
			respondWithError(w, http.StatusBadRequest, "приемка закрыта, нельзя добавлять товары", err, h.logger)
		default:
			respondWithError(w, http.StatusInternalServerError, "ошибка при создании товара", err, h.logger)
		}

		return
	}

	metrics.ProductsAddedTotal.Inc()

	productID, _ := uuid.Parse(newProduct.ID.String())
	receptionID, _ := uuid.Parse(newProduct.ReceptionID.String())
	dateTime := newProduct.DateTime

	response := dto.Product{
		Id:          &productID,
		DateTime:    &dateTime,
		ReceptionId: receptionID,
	}

	switch newProduct.Type {
	case product.TypeElectronics:
		response.Type = dto.ProductType("электроника")
	case product.TypeClothes:
		response.Type = dto.ProductType("одежда")
	case product.TypeShoes:
		response.Type = dto.ProductType("обувь")
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *ProductHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "метод не поддерживается", nil, h.logger)
		return
	}

	path := r.URL.Path
	h.logger.Info("Получен запрос на удаление последнего товара", "path", path)

	trimmedPath := strings.Trim(path, "/")
	parts := strings.Split(trimmedPath, "/")

	h.logger.Info("Разбор пути", "parts", parts, "count", len(parts))

	if len(parts) != 3 || parts[0] != "pvz" || parts[2] != "delete_last_product" {
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

	err = h.service.DeleteLastProduct(r.Context(), pvzID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoActiveReceptionProduct):
			respondWithError(w, http.StatusBadRequest, "нет активной приемки", err, h.logger)
		case errors.Is(err, ErrNoProductsToDelete):
			respondWithError(w, http.StatusBadRequest, "нет товаров для удаления", err, h.logger)
		case errors.Is(err, ErrReceptionClosedForProduct):
			respondWithError(w, http.StatusBadRequest, "приемка уже закрыта", err, h.logger)
		default:
			respondWithError(w, http.StatusInternalServerError, "ошибка при удалении товара", err, h.logger)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}
