package product

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeElectronics Type = "электроника"
	TypeClothes     Type = "одежда"
	TypeShoes       Type = "обувь"
)

func (t Type) Validate() bool {
	switch t {
	case TypeElectronics, TypeClothes, TypeShoes:
		return true
	default:
		return false
	}
}

type Product struct {
	ID             uuid.UUID `json:"id"`
	DateTime       time.Time `json:"dateTime"`
	Type           Type      `json:"type"`
	ReceptionID    uuid.UUID `json:"receptionId"`
	SequenceNumber int       `json:"-"`
	CreatedAt      time.Time `json:"-"`
}

type CreateProductRequest struct {
	Type  Type      `json:"type"`
	PVZID uuid.UUID `json:"pvzId"`
}
