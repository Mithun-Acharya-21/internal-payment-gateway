# internal payment gateway

a simple internal payment processing api built in go.

features:
- jwt auth
- atomic payments and wallet deductions
- idempotency keys to prevent duplicate charges
- docker support

## setup

you need go 1.22 and docker installed.

1. clone the repo and copy the env file
cp .env.example .env

2. start the services
docker compose up --build

api runs on localhost:8080.

## local development without docker

if you just want to run the go app directly:
export DATABASE_URL="postgres://pguser:pgpass@localhost:5432/paymentdb?sslmode=disable"
export JWT_SECRET="your-super-secret-key-at-least-32-chars"
go run ./cmd/server

## testing

go test ./... -race -cover

## api

all routes under /api/v1 need a bearer token.

POST /api/v1/wallets
create a wallet for a user.

POST /api/v1/wallets/{id}/topup
add money to a wallet.

POST /api/v1/payments
initiate a payment. pass X-Idempotency-Key in the header so it doesn't double charge on retry.

GET /api/v1/payments/{id}
get payment details.

POST /api/v1/payments/{id}/refund
refund a completed payment.
