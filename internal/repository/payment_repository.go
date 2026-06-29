package repository

import (
	"context"
	"database/sql"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error)
	ListByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*domain.Transaction, int64, error)
	Update(ctx context.Context, tx *domain.Transaction) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type WalletRepository interface {
	Create(ctx context.Context, w *domain.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, id uuid.UUID, balance int64) error
}

type postgresTransactionRepo struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &postgresTransactionRepo{db: db}
}

func (r *postgresTransactionRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	const q = `
		INSERT INTO transactions
			(id, wallet_id, amount, currency, status, description, idempotency_key, failure_reason, refunded_transaction_id, created_at, updated_at)
		VALUES
			(:id, :wallet_id, :amount, :currency, :status, :description, :idempotency_key, :failure_reason, :refunded_transaction_id, :created_at, :updated_at)
	`
	_, err := r.current(ctx).NamedExecContext(ctx, q, tx)
	return err
}

func (r *postgresTransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.GetContext(ctx, &tx, `SELECT * FROM transactions WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTransactionNotFound
	}
	return &tx, err
}

func (r *postgresTransactionRepo) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.GetContext(ctx, &tx, `SELECT * FROM transactions WHERE idempotency_key = $1`, key)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTransactionNotFound
	}
	return &tx, err
}

func (r *postgresTransactionRepo) ListByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*domain.Transaction, int64, error) {
	var txs []*domain.Transaction
	var total int64

	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM transactions WHERE wallet_id = $1`, walletID); err != nil {
		return nil, 0, err
	}

	err := r.db.SelectContext(ctx, &txs,
		`SELECT * FROM transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		walletID, limit, offset,
	)
	return txs, total, err
}

func (r *postgresTransactionRepo) Update(ctx context.Context, tx *domain.Transaction) error {
	const q = `
		UPDATE transactions SET status = :status, failure_reason = :failure_reason, updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.current(ctx).NamedExecContext(ctx, q, tx)
	return err
}

func (r *postgresTransactionRepo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	dbTx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = dbTx.Rollback()
			panic(p)
		}
	}()

	txRepo := &postgresTransactionTxRepo{tx: dbTx}
	wrappedCtx := context.WithValue(ctx, txRepoKey{}, txRepo)

	if err := fn(wrappedCtx); err != nil {
		_ = dbTx.Rollback()
		return err
	}
	return dbTx.Commit()
}

type postgresWalletRepo struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) WalletRepository {
	return &postgresWalletRepo{db: db}
}

func (r *postgresWalletRepo) Create(ctx context.Context, w *domain.Wallet) error {
	const q = `
		INSERT INTO wallets (id, user_id, balance, currency, is_active, created_at, updated_at)
		VALUES (:id, :user_id, :balance, :currency, :is_active, :created_at, :updated_at)
	`
	_, err := r.execer(ctx).NamedExecContext(ctx, q, w)
	return err
}

func (r *postgresWalletRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.execer(ctx).GetContext(ctx, &w, `SELECT * FROM wallets WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, domain.ErrWalletNotFound
	}
	return &w, err
}

func (r *postgresWalletRepo) GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.execer(ctx).GetContext(ctx, &w, `SELECT * FROM wallets WHERE user_id = $1`, userID)
	if err == sql.ErrNoRows {
		return nil, domain.ErrWalletNotFound
	}
	return &w, err
}

func (r *postgresWalletRepo) UpdateBalance(ctx context.Context, id uuid.UUID, balance int64) error {
	_, err := r.execer(ctx).ExecContext(ctx,
		`UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`, balance, id)
	return err
}

type txRepoKey struct{}

type dbExecutor interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (r *postgresWalletRepo) execer(ctx context.Context) dbExecutor {
	if v := ctx.Value(txRepoKey{}); v != nil {
		if txr, ok := v.(*postgresTransactionTxRepo); ok {
			return txr.tx
		}
	}
	return r.db
}

type postgresTransactionTxRepo struct {
	tx *sqlx.Tx
}

func (r *postgresTransactionRepo) current(ctx context.Context) dbExecutor {
	if v := ctx.Value(txRepoKey{}); v != nil {
		if txr, ok := v.(*postgresTransactionTxRepo); ok {
			return txr.tx
		}
	}
	return r.db
}

