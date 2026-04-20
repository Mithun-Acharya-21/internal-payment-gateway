package repository

import (
	"context"
	"database/sql"
	"errors"

	"STRIPE/internal/model"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Expose DB so service can start transactions
func (r *PaymentRepository) DB() *sql.DB {
	return r.db
}

func (r *PaymentRepository) Create(ctx context.Context, tx *sql.Tx, p *model.Payment) error {

	query := `
	INSERT INTO payments
	(user_id, amount, currency, status, idempotency_key)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at
	`

	err := tx.QueryRowContext(
		ctx,
		query,
		p.UserID,
		p.Amount,
		p.Currency,
		p.Status,
		p.IdempotencyKey,
	).Scan(&p.ID, &p.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) GetByIdempotencyKey(
	ctx context.Context,
	key string,
) (*model.Payment, error) {

	query := `
	SELECT id, user_id, amount, currency, status, idempotency_key, created_at
	FROM payments
	WHERE idempotency_key = $1
	`

	row := r.db.QueryRowContext(ctx, query, key)

	var p model.Payment

	err := row.Scan(
		&p.ID,
		&p.UserID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.IdempotencyKey,
		&p.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}