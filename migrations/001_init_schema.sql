CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS wallets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     VARCHAR(255) NOT NULL UNIQUE,
    balance     BIGINT       NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency    VARCHAR(3)   NOT NULL DEFAULT 'INR',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

CREATE TABLE IF NOT EXISTS transactions (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id                 UUID         NOT NULL REFERENCES wallets(id),
    amount                    BIGINT       NOT NULL CHECK (amount > 0),
    currency                  VARCHAR(3)   NOT NULL,
    status                    VARCHAR(20)  NOT NULL DEFAULT 'pending',
    description               TEXT,
    idempotency_key           VARCHAR(64)  UNIQUE,
    failure_reason            TEXT,
    refunded_transaction_id   UUID         REFERENCES transactions(id),
    created_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id   ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status      ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_created_at  ON transactions(created_at DESC);

