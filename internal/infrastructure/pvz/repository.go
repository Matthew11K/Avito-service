package pvz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"avito/internal/domain/product"
	"avito/internal/domain/pvz"
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

func (r *Repository) CreatePVZ(ctx context.Context, city pvz.City) (*pvz.PVZ, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var pvzObj pvz.PVZ
	err := q.QueryRow(ctx, `
        INSERT INTO pvz (city)
        VALUES ($1)
        RETURNING id, registration_date, city
    `, city).Scan(&pvzObj.ID, &pvzObj.RegistrationDate, &pvzObj.City)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании ПВЗ: %w", err)
	}

	return &pvzObj, nil
}

func (r *Repository) GetPVZByID(ctx context.Context, id uuid.UUID) (*pvz.PVZ, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var pvzObj pvz.PVZ
	err := q.QueryRow(ctx, `
        SELECT id, registration_date, city
        FROM pvz
        WHERE id = $1
    `, id).Scan(&pvzObj.ID, &pvzObj.RegistrationDate, &pvzObj.City)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &pvz.ErrPVZNotFound{}
		}

		return nil, fmt.Errorf("ошибка при поиске ПВЗ: %w", err)
	}

	return &pvzObj, nil
}

//nolint:funlen // сложный SQL-конструктор, разбиение ухудшит читаемость и поддержку кода
func (r *Repository) GetPVZs(ctx context.Context, startDate, endDate *time.Time, city *pvz.City,
	page, limit int) ([]pvz.WithReceptions, error) {
	q := txs.GetQuerier(ctx, r.pool)

	query := `
		SELECT p.id, p.registration_date, p.city
		FROM pvz p
	`

	where := []string{}
	args := []interface{}{}
	argIndex := 1

	if city != nil {
		where = append(where, fmt.Sprintf("p.city = $%d", argIndex))
		args = append(args, *city)
		argIndex++
	}

	if startDate != nil || endDate != nil {
		subquery := `EXISTS (
			SELECT 1 FROM receptions r 
			WHERE r.pvz_id = p.id
		`

		subqueryConds := []string{}

		if startDate != nil {
			subqueryConds = append(subqueryConds, fmt.Sprintf("r.date_time >= $%d", argIndex))
			args = append(args, startDate)
			argIndex++
		}

		if endDate != nil {
			subqueryConds = append(subqueryConds, fmt.Sprintf("r.date_time <= $%d", argIndex))
			args = append(args, endDate)
			argIndex++
		}

		if len(subqueryConds) > 0 {
			subquery += " AND " + strings.Join(subqueryConds, " AND ")
		}

		subquery += ")"
		where = append(where, subquery)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY p.registration_date DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	args = append(args, limit, (page-1)*limit)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка ПВЗ: %w", err)
	}
	defer rows.Close()

	var result []pvz.WithReceptions

	for rows.Next() {
		var pvzObj pvz.PVZ
		if err := rows.Scan(&pvzObj.ID, &pvzObj.RegistrationDate, &pvzObj.City); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании результатов ПВЗ: %w", err)
		}

		receptions, err := r.getReceptionsWithProductsByPVZID(ctx, pvzObj.ID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении приемок для ПВЗ: %w", err)
		}

		result = append(result, pvz.WithReceptions{
			PVZ:        pvzObj,
			Receptions: receptions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов ПВЗ: %w", err)
	}

	return result, nil
}

func (r *Repository) getReceptionsWithProductsByPVZID(ctx context.Context, pvzID uuid.UUID, startDate,
	endDate *time.Time) ([]pvz.ReceptionWithItems, error) {
	q := txs.GetQuerier(ctx, r.pool)

	query := `
        SELECT r.id, r.date_time, r.pvz_id, r.status
        FROM receptions r
        WHERE r.pvz_id = $1
    `

	args := []interface{}{pvzID}
	argIndex := 2

	if startDate != nil {
		query += fmt.Sprintf(" AND r.date_time >= $%d", argIndex)

		args = append(args, startDate)
		argIndex++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND r.date_time <= $%d", argIndex)

		args = append(args, endDate)
	}

	query += " ORDER BY r.date_time DESC"

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении приемок: %w", err)
	}
	defer rows.Close()

	var result []pvz.ReceptionWithItems

	for rows.Next() {
		var rec reception.Reception
		if err := rows.Scan(&rec.ID, &rec.DateTime, &rec.PVZID, &rec.Status); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании результатов приемок: %w", err)
		}

		products, err := r.getProductsByReceptionID(ctx, rec.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении товаров для приемки: %w", err)
		}

		result = append(result, pvz.ReceptionWithItems{
			Reception: rec,
			Products:  products,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов приемок: %w", err)
	}

	return result, nil
}

func (r *Repository) getProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]product.Product, error) {
	q := txs.GetQuerier(ctx, r.pool)

	query := `
        SELECT id, date_time, type, reception_id, sequence_number
        FROM products
        WHERE reception_id = $1
        ORDER BY sequence_number
    `

	rows, err := q.Query(ctx, query, receptionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении товаров: %w", err)
	}
	defer rows.Close()

	var products []product.Product

	for rows.Next() {
		var prod product.Product
		if err := rows.Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionID, &prod.SequenceNumber); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании результатов товаров: %w", err)
		}

		products = append(products, prod)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов товаров: %w", err)
	}

	return products, nil
}
