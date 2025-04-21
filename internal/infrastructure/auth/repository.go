package auth

import (
	"context"
	"errors"
	"fmt"

	domainAuth "avito/internal/domain/auth"
	"avito/pkg/txs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string, role domainAuth.Role) (*domainAuth.User, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var user domainAuth.User
	err := q.QueryRow(ctx, `
        INSERT INTO users (email, password_hash, role)
        VALUES ($1, $2, $3)
        RETURNING id, email, password_hash, role
    `, email, passwordHash, role).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &domainAuth.ErrUserAlreadyExists{}
		}

		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domainAuth.User, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var user domainAuth.User
	err := q.QueryRow(ctx, `
        SELECT id, email, password_hash, role
        FROM users
        WHERE email = $1
    `, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domainAuth.ErrUserNotFound{}
		}

		return nil, fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*domainAuth.User, error) {
	q := txs.GetQuerier(ctx, r.pool)

	var user domainAuth.User
	err := q.QueryRow(ctx, `
        SELECT id, email, password_hash, role
        FROM users
        WHERE id = $1
    `, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domainAuth.ErrUserNotFound{}
		}

		return nil, fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	return &user, nil
}
