package product

import (
	"context"
	"errors"
	"fmt"

	"avito/internal/domain/product"
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

func (r *Repository) AddProduct(ctx context.Context, productType product.Type, receptionID uuid.UUID) (*product.Product, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var nextSequence int
	err := q.QueryRow(ctx, `
        SELECT COALESCE(MAX(sequence_number), 0) + 1
        FROM products
        WHERE reception_id = $1
    `, receptionID).Scan(&nextSequence)

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении номера последовательности: %w", err)
	}

	var productObj product.Product
	err = q.QueryRow(ctx, `
        INSERT INTO products (type, reception_id, sequence_number)
        VALUES ($1, $2, $3)
        RETURNING id, date_time, type, reception_id, sequence_number
    `, productType, receptionID, nextSequence).Scan(
		&productObj.ID,
		&productObj.DateTime,
		&productObj.Type,
		&productObj.ReceptionID,
		&productObj.SequenceNumber,
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка при добавлении товара: %w", err)
	}

	return &productObj, nil
}

func (r *Repository) DeleteLastProduct(ctx context.Context, receptionID uuid.UUID) error {
	q := txs.GetQuerier(ctx, r.pool)

	var productID uuid.UUID

	var sequenceNumber int
	err := q.QueryRow(ctx, `
        SELECT id, sequence_number
        FROM products
        WHERE reception_id = $1
        ORDER BY sequence_number DESC
        LIMIT 1
    `, receptionID).Scan(&productID, &sequenceNumber)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &product.ErrNoProductsToDelete{}
		}

		return fmt.Errorf("ошибка при поиске последнего товара: %w", err)
	}

	cmdTag, err := q.Exec(ctx, `
        DELETE FROM products
        WHERE id = $1
    `, productID)

	if err != nil {
		return fmt.Errorf("ошибка при удалении товара: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return &product.ErrProductNotFound{}
	}

	return nil
}

func (r *Repository) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]product.Product, error) {
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
