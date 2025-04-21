package reception

import (
	"context"
	"errors"
	"fmt"

	"avito/internal/domain/reception"
	"avito/pkg/txs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateReception(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	q := txs.GetQuerier(ctx, r.pool)

	existingReception, err := r.GetActiveReceptionByPVZID(ctx, pvzID)
	if err != nil && !errors.Is(err, &reception.ErrNoActiveReception{}) {
		return nil, fmt.Errorf("ошибка при проверке активной приемки: %w", err)
	}

	if existingReception != nil {
		return nil, &reception.ErrActiveReceptionExists{}
	}

	var receptionObj reception.Reception
	err = q.QueryRow(ctx, `
        INSERT INTO receptions (pvz_id, status)
        VALUES ($1, $2)
        RETURNING id, date_time, pvz_id, status
    `, pvzID, reception.StatusInProgress).Scan(&receptionObj.ID, &receptionObj.DateTime, &receptionObj.PVZID, &receptionObj.Status)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании приемки: %w", err)
	}

	return &receptionObj, nil
}

func (r *Repository) GetActiveReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*reception.Reception, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var receptionObj reception.Reception
	err := q.QueryRow(ctx, `
        SELECT id, date_time, pvz_id, status
        FROM receptions
        WHERE pvz_id = $1 AND status = $2
        ORDER BY date_time DESC
        LIMIT 1
    `, pvzID, reception.StatusInProgress).Scan(&receptionObj.ID, &receptionObj.DateTime, &receptionObj.PVZID, &receptionObj.Status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &reception.ErrNoActiveReception{}
		}

		return nil, fmt.Errorf("ошибка при поиске активной приемки: %w", err)
	}

	return &receptionObj, nil
}

func (r *Repository) GetReceptionByID(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var receptionObj reception.Reception
	err := q.QueryRow(ctx, `
        SELECT id, date_time, pvz_id, status
        FROM receptions
        WHERE id = $1
    `, id).Scan(&receptionObj.ID, &receptionObj.DateTime, &receptionObj.PVZID, &receptionObj.Status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &reception.ErrReceptionNotFound{}
		}

		return nil, fmt.Errorf("ошибка при поиске приемки: %w", err)
	}

	return &receptionObj, nil
}

func (r *Repository) CloseReception(ctx context.Context, id uuid.UUID) (*reception.Reception, error) {
	q := txs.GetQuerier(ctx, r.pool)

	currentReception, err := r.GetReceptionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if currentReception.Status == reception.StatusClosed {
		return nil, &reception.ErrReceptionClosed{}
	}

	var updatedReception reception.Reception
	err = q.QueryRow(ctx, `
        UPDATE receptions
        SET status = $1
        WHERE id = $2
        RETURNING id, date_time, pvz_id, status
    `, reception.StatusClosed, id).Scan(&updatedReception.ID, &updatedReception.DateTime, &updatedReception.PVZID, &updatedReception.Status)

	if err != nil {
		return nil, fmt.Errorf("ошибка при закрытии приемки: %w", err)
	}

	return &updatedReception, nil
}
