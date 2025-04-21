package reception

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "close"
)

type Reception struct {
	ID        uuid.UUID `json:"id"`
	DateTime  time.Time `json:"dateTime"`
	PVZID     uuid.UUID `json:"pvzId"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"-"`
}

type CreateReceptionRequest struct {
	PVZID uuid.UUID `json:"pvzId"`
}
