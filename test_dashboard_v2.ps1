# Test Dashboard V2 API Endpoints
# This script tests the new cache_metrics and prefetch_metrics endpoints

$baseUrl = "http://localhost:3000"

Write-Host "=== Testing Dashboard V2 API Endpoints ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: Cache Metrics
Write-Host "1. Testing /control/cache_metrics endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/control/cache_metrics" -Method Get
    Write-Host "✓ Cache Metrics Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
    Write-Host ""
} catch {
    Write-Host "✗ Failed to fetch cache metrics: $_" -ForegroundColor Red
    Write-Host ""
}

# Test 2: Prefetch Metrics
Write-Host "2. Testing /control/prefetch_metrics endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/control/prefetch_metrics" -Method Get
    Write-Host "✓ Prefetch Metrics Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
    Write-Host ""
} catch {
    Write-Host "✗ Failed to fetch prefetch metrics: $_" -ForegroundColor Red
    Write-Host ""
}

# Test 3: Dashboard Metrics (existing)
Write-Host "3. Testing /control/dashboard_metrics endpoint (existing)..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/control/dashboard_metrics" -Method Get
    Write-Host "✓ Dashboard Metrics Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
    Write-Host ""
} catch {
    Write-Host "✗ Failed to fetch dashboard metrics: $_" -ForegroundColor Red
    Write-Host ""
}

Write-Host "=== Test Complete ===" -ForegroundColor Cyan
