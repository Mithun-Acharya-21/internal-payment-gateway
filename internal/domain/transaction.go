package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrWalletNotFound           = errors.New("wallet not found")
	ErrInsufficientBalance      = errors.New("insufficient wallet balance")
	ErrInvalidAmount            = errors.New("amount must be greater than zero")
	ErrTransactionNotRefundable = errors.New("only completed transactions can be refunded")
	ErrDuplicateTransaction     = errors.New("duplicate transaction detected")
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
)

type Currency string

const (
	CurrencyINR Currency = "INR"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

type Transaction struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	WalletID       uuid.UUID     `json:"wallet_id" db:"wallet_id"`
	Amount         int64         `json:"amount" db:"amount"`
	Currency       Currency      `json:"currency" db:"currency"`
	Status         PaymentStatus `json:"status" db:"status"`
	Description    string        `json:"description" db:"description"`
	IdempotencyKey string        `json:"-" db:"idempotency_key"`
	FailureReason  string        `json:"failure_reason,omitempty" db:"failure_reason"`
	RefundedTxID   *uuid.UUID    `json:"refunded_transaction_id,omitempty" db:"refunded_transaction_id"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

func (t *Transaction) CanRefund() bool {
	return t.Status == StatusCompleted
}

type Wallet struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Balance   int64     `json:"balance" db:"balance"`
	Currency  Currency  `json:"currency" db:"currency"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (w *Wallet) HasSufficientBalance(amount int64) bool {
	return w.Balance >= amount
}

func (w *Wallet) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !w.HasSufficientBalance(amount) {
		return ErrInsufficientBalance
	}
	w.Balance -= amount
	return nil
}

func (w *Wallet) Credit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	w.Balance += amount
	return nil
}

