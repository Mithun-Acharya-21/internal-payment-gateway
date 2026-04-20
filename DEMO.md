# Demo Runbook

## 1) Start dependencies

```powershell
docker compose up -d postgres
```

## 2) Export env and run API

```powershell
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $parts = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1], 'Process')
}
go run ./cmd/server
```

API will be available at `http://localhost:8080`.

## 3) Run full payment demo flow (new terminal)

```powershell
cd c:\Users\mh129\OneDrive\Desktop\internal-payment
$env:JWT_SECRET="demo-secret-key-at-least-32-chars"
.\scripts\demo.ps1
```

This runs:
- create wallet
- top-up wallet
- initiate payment
- refund payment

## 4) Optional UI

Open:
- `http://localhost:8080/` (dashboard)
- `http://localhost:8080/healthz`

