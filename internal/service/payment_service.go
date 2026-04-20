package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/domain"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PaymentService struct {
	txRepo     repository.TransactionRepository
	walletRepo repository.WalletRepository
	logger     *zap.Logger
}

func NewPaymentService(txRepo repository.TransactionRepository, walletRepo repository.WalletRepository, logger *zap.Logger) *PaymentService {
	return &PaymentService{txRepo: txRepo, walletRepo: walletRepo, logger: logger}
}

type InitiatePaymentInput struct {
	WalletID       uuid.UUID
	Amount         int64
	Currency       domain.Currency
	Description    string
	IdempotencyKey string
}

func (s *PaymentService) InitiatePayment(ctx context.Context, in InitiatePaymentInput) (*domain.Transaction, error) {
	if in.IdempotencyKey != "" {
		existing, err := s.txRepo.FindByIdempotencyKey(ctx, in.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	wallet, err := s.walletRepo.GetByID(ctx, in.WalletID)
	if err != nil {
		return nil, fmt.Errorf("fetch wallet: %w", err)
	}

	tx := &domain.Transaction{
		ID:             uuid.New(),
		WalletID:       in.WalletID,
		Amount:         in.Amount,
		Currency:       in.Currency,
		Status:         domain.StatusPending,
		Description:    in.Description,
		IdempotencyKey: in.IdempotencyKey,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.txRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := wallet.Debit(in.Amount); err != nil {
			tx.Status = domain.StatusFailed
			tx.FailureReason = err.Error()
			return s.txRepo.Create(txCtx, tx)
		}

		tx.Status = domain.StatusCompleted
		if err := s.walletRepo.UpdateBalance(txCtx, wallet.ID, wallet.Balance); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}
		return s.txRepo.Create(txCtx, tx)
	}); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}

func (s *PaymentService) ListPayments(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*domain.Transaction, int64, error) {
	return s.txRepo.ListByWallet(ctx, walletID, limit, offset)
}

func (s *PaymentService) RefundPayment(ctx context.Context, txID uuid.UUID) (*domain.Transaction, error) {
	original, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if !original.CanRefund() {
		return nil, domain.ErrTransactionNotRefundable
	}

	wallet, err := s.walletRepo.GetByID(ctx, original.WalletID)
	if err != nil {
		return nil, err
	}

	refundTx := &domain.Transaction{
		ID:          uuid.New(),
		WalletID:    original.WalletID,
		Amount:      original.Amount,
		Currency:    original.Currency,
		Status:      domain.StatusRefunded,
		Description: "Refund for transaction " + original.ID.String(),
		RefundedTxID: &original.ID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.txRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := wallet.Credit(original.Amount); err != nil {
			return err
		}
		if err := s.walletRepo.UpdateBalance(txCtx, wallet.ID, wallet.Balance); err != nil {
			return err
		}
		original.Status = domain.StatusRefunded
		original.UpdatedAt = time.Now().UTC()
		if err := s.txRepo.Update(txCtx, original); err != nil {
			return err
		}
		return s.txRepo.Create(txCtx, refundTx)
	}); err != nil {
		return nil, err
	}

	return refundTx, nil
}

