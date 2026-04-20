package service

import (
	"context"
	"errors"

	"STRIPE/internal/model"
	"STRIPE/internal/repository"
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) ProcessPayment(
	ctx context.Context,
	userID string,
	amount float64,
	currency string,
	idempotencyKey string,
) (*model.Payment, error) {

	// -------- VALIDATION --------

	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	if currency == "" {
		return nil, errors.New("currency is required")
	}

	if idempotencyKey == "" {
		return nil, errors.New("idempotency key required")
	}

	// -------- IDEMPOTENCY CHECK --------

	existing, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	// -------- START TRANSACTION --------

	tx, err := s.repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	// -------- CREATE PAYMENT --------

	payment := model.NewPayment(
		userID,
		amount,
		currency,
		idempotencyKey,
	)

	// Simulated gateway success
	payment.Status = model.StatusProcessed

	err = s.repo.Create(ctx, tx, payment)
	if err != nil {
		return nil, err
	}

	// -------- COMMIT --------

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return payment, nil
}
