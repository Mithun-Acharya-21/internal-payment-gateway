# 🏦 Internal Payment Gateway API

[![CI/CD](https://github.com/Mithun-Acharya-21/internal-payment-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/Mithun-Acharya-21/internal-payment-gateway/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A **production-grade internal payment processing API** built with Go, featuring atomic transactions, idempotent payments, JWT authentication, structured logging, and Docker-based deployment.

---

## ✨ Features

| Feature | Details |
|---|---|
| 🔐 JWT Authentication | Bearer token auth on all payment endpoints |
| 💸 Atomic Payments | Debit + transaction record in a single DB transaction |
| 🔁 Idempotency | Re-submitting same key returns existing transaction safely |
| 💳 Wallet Management | Create, top-up, balance check |
| ↩️ Refunds | Atomic refund with credit reversal |
| 📋 Structured Logs | JSON logs via `zap` with request ID tracing |
| ⚡ Rate Limiting | Token bucket (configurable RPS) |
| 🐳 Docker | Multi-stage build, distroless image, <10MB binary |
| 🔄 CI/CD | GitHub Actions: lint → test → build → push |
| 🩺 Health Checks | `/healthz` and `/readyz` endpoints |

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- `make` (optional but recommended)

### 1. Clone & Configure
```bash
git clone https://github.com/Mithun-Acharya-21/internal-payment-gateway.git
cd internal-payment-gateway

cp .env.example .env
# Edit .env — set JWT_SECRET (min 32 chars)
```

### 2. Run with Docker Compose
```bash
docker compose up --build
```

API is live at **http://localhost:8080**

### 3. Run Locally (without Docker)
```bash
export DATABASE_URL="postgres://pguser:pgpass@localhost:5432/paymentdb?sslmode=disable"
export JWT_SECRET="your-super-secret-key-at-least-32-chars"
go run ./cmd/server
```

---

## 📡 API Reference

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
All endpoints (except `/healthz`, `/readyz`) require:
```
Authorization: Bearer <jwt_token>
```

---

### Wallets

#### Create Wallet
```http
POST /api/v1/wallets
Content-Type: application/json
Authorization: Bearer <token>

{
  "user_id": "user_123",
  "currency": "INR"
}
```

**Response `201`:**
```json
{
  "success": true,
  "message": "wallet created",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "user_123",
    "balance": 0,
    "currency": "INR",
    "is_active": true,
    "created_at": "2026-04-20T10:00:00Z"
  },
  "request_id": "req_abc123"
}
```

#### Top-Up Wallet
```http
POST /api/v1/wallets/{id}/topup
Content-Type: application/json

{ "amount": 100000 }
```
> Amount in smallest unit: `100000` paise = ₹1,000.00

---

### Payments

#### Initiate Payment
```http
POST /api/v1/payments
Content-Type: application/json
X-Idempotency-Key: pay_unique_key_001

{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 25000,
  "currency": "INR",
  "description": "Subscription - Pro Plan"
}
```

**Response `201` (success):**
```json
{
  "success": true,
  "message": "payment processed",
  "data": {
    "id": "tx_abc123",
    "wallet_id": "550e8400-...",
    "amount": 25000,
    "currency": "INR",
    "status": "completed",
    "description": "Subscription - Pro Plan",
    "created_at": "2026-04-20T10:05:00Z"
  }
}
```

**Response `422` (insufficient funds):**
```json
{
  "success": false,
  "message": "insufficient wallet balance",
  "data": {
    "id": "tx_def456",
    "status": "failed",
    "failure_reason": "insufficient wallet balance"
  }
}
```

#### Get Payment
```http
GET /api/v1/payments/{id}
```

#### List Payments
```http
GET /api/v1/payments?wallet_id={uuid}&limit=20&offset=0
```

#### Refund Payment
```http
POST /api/v1/payments/{id}/refund
```

---

## 🏗️ Architecture

```
cmd/server/          → Entrypoint: wires dependencies, starts HTTP server
internal/
  config/            → Env-based config with validation
  domain/            → Core entities: Transaction, Wallet (no framework deps)
  service/           → Business logic: payment orchestration, idempotency
  handler/           → HTTP handlers: request binding, response shaping
  middleware/         → JWT, logger, rate limiter, CORS, recovery
  repository/        → DB layer: interfaces + postgres implementations
  database/          → Connection pool + migrations
pkg/
  response/          → Standard API response envelope
migrations/          → SQL schema migrations
```

### Design Principles
- **Clean Architecture** — domain layer has zero framework dependencies
- **Repository Pattern** — swap Postgres for mock in tests trivially
- **Idempotency** — safe to retry payments without double-charging
- **Atomic DB Transactions** — balance update + transaction record never split
- **Error wrapping** — `errors.Is()` throughout, no string matching

---

## 🧪 Testing

```bash
# Unit + integration tests
go test ./... -race -cover

# With coverage report
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

---

## 🐳 Docker

```bash
# Build production image (~8MB, distroless)
docker build --target production -t payment-gateway .

# Run
docker run -p 8080:8080 \
  -e DATABASE_URL="..." \
  -e JWT_SECRET="..." \
  payment-gateway
```

---

## ⚙️ Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | PostgreSQL connection string |
| `JWT_SECRET` | ✅ | — | Min 32-char secret for JWT signing |
| `PORT` | ❌ | `8080` | HTTP server port |
| `APP_ENV` | ❌ | `development` | `development` or `production` |
| `RATE_LIMIT_RPS` | ❌ | `100` | Max requests/second (token bucket) |
| `ALLOWED_ORIGINS` | ❌ | `http://localhost:3000` | Comma-separated CORS origins |

---

## 📦 Tech Stack

- **Runtime:** Go 1.22
- **HTTP:** Gin
- **Database:** PostgreSQL 16 + sqlx
- **Auth:** JWT (golang-jwt/jwt v5)
- **Logging:** Uber zap (structured JSON)
- **Container:** Docker (distroless, multi-stage)
- **CI/CD:** GitHub Actions
- **Migrations:** golang-migrate

---

## 🔒 Security Considerations

- JWT secret validated to be ≥ 32 characters at startup
- Distroless Docker image — no shell, no package manager
- Non-root container user
- SQL queries use parameterized inputs (no string interpolation)
- Idempotency keys stored with UNIQUE constraint (DB-level guarantee)
- Rate limiting prevents abuse

---

## 👤 Author

**Mithun Acharya** — Backend Engineer  
[GitHub](https://github.com/Mithun-Acharya-21)

---

*Built to demonstrate production-grade Go API design for high-scale payment systems.*
