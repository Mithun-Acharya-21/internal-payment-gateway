# demo

start the db:
docker compose up -d postgres

run the app:
go run ./cmd/server

run the demo script in powershell:
cd c:\Users\mh129\OneDrive\Desktop\internal-payment
$env:JWT_SECRET="demo-secret-key-at-least-32-chars"
.\scripts\demo.ps1

this script will automatically create a wallet, add funds, run a payment, and refund it.
