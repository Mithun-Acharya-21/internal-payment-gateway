param(
  [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

Write-Host "Generating JWT token..."
$token = go run ./cmd/token
if (-not $token) {
  throw "Failed to generate JWT token."
}
$headers = @{
  Authorization = "Bearer $token"
  "Content-Type" = "application/json"
}

Write-Host "1) Creating wallet..."
$walletRes = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/wallets" -Headers $headers -Body '{"user_id":"demo_user","currency":"INR"}'
$walletId = $walletRes.data.id
Write-Host "Wallet ID: $walletId"

Write-Host "2) Top-up wallet..."
Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/wallets/$walletId/topup" -Headers $headers -Body '{"amount":100000}' | Out-Null

Write-Host "3) Initiate payment..."
$paymentHeaders = @{
  Authorization = "Bearer $token"
  "Content-Type" = "application/json"
  "X-Idempotency-Key" = "demo-idempotency-001"
}
$payBody = @{ wallet_id = $walletId; amount = 25000; currency = "INR"; description = "Demo payment" } | ConvertTo-Json
$paymentRes = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/payments" -Headers $paymentHeaders -Body $payBody
$paymentId = $paymentRes.data.id
Write-Host "Payment ID: $paymentId"

Write-Host "4) Refund payment..."
$refundRes = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/payments/$paymentId/refund" -Headers $headers

Write-Host ""
Write-Host "Demo flow complete."
Write-Host "Wallet response:"
$walletRes | ConvertTo-Json -Depth 8
Write-Host "Payment response:"
$paymentRes | ConvertTo-Json -Depth 8
Write-Host "Refund response:"
$refundRes | ConvertTo-Json -Depth 8

