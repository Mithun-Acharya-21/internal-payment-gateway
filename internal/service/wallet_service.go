package service

import (
	"context"
	"time"

	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/domain"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WalletService struct {
	walletRepo repository.WalletRepository
	logger     *zap.Logger
}

func NewWalletService(walletRepo repository.WalletRepository, logger *zap.Logger) *WalletService {
	return &WalletService{walletRepo: walletRepo, logger: logger}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID string, currency domain.Currency) (*domain.Wallet, error) {
	w := &domain.Wallet{
		ID:        uuid.New(),
		UserID:    userID,
		Balance:   0,
		Currency:  currency,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.walletRepo.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WalletService) GetWallet(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	return s.walletRepo.GetByID(ctx, id)
}

func (s *WalletService) TopUp(ctx context.Context, id uuid.UUID, amount int64) (*domain.Wallet, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	w, err := s.walletRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := w.Credit(amount); err != nil {
		return nil, err
	}
	if err := s.walletRepo.UpdateBalance(ctx, id, w.Balance); err != nil {
		return nil, err
	}
	w.UpdatedAt = time.Now().UTC()
	return w, nil
}

