package pvz

import (
	"time"

	"avito/internal/domain/product"
	"avito/internal/domain/reception"

	"github.com/google/uuid"
)

type City string

const (
	CityMoscow          City = "Москва"
	CitySaintPetersburg City = "Санкт-Петербург"
	CityKazan           City = "Казань"
)

func (c City) Validate() bool {
	switch c {
	case CityMoscow, CitySaintPetersburg, CityKazan:
		return true
	default:
		return false
	}
}

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             City      `json:"city"`
}

type CreatePVZRequest struct {
	City City `json:"city"`
}

type GetPVZsRequest struct {
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
	City      *City      `json:"city,omitempty"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

type WithReceptions struct {
	PVZ        PVZ                  `json:"pvz"`
	Receptions []ReceptionWithItems `json:"receptions"`
}

type ReceptionWithItems struct {
	Reception reception.Reception `json:"reception"`
	Products  []product.Product   `json:"products"`
}
