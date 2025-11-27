# Test Cache Hit Bypasses Threshold
# Verifies that cache hits bypass threshold checking

Write-Host "=== Test: Cache Hit Bypasses Threshold ===" -ForegroundColor Cyan
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"
$testDomain = "www.google.com"
$auth = "Basic bGt4bHp4OjE5ODIxMDEy"

# Clean up
if (Test-Path "prefetch_errors.log") {
    Remove-Item "prefetch_errors.log" -Force
}

# Start service
Write-Host "Starting service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\$exeName" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 5
Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Test 1: Single query (cache miss) - should NOT be hot
Write-Host "=== Test 1: Single Query (Cache Miss) ===" -ForegroundColor Cyan
Write-Host "Threshold: 2 hits required" -ForegroundColor Yellow
Write-Host "Querying once (cache miss)..." -ForegroundColor Yellow
$result = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
Write-Host "IP: $($result[0].IPAddress)" -ForegroundColor Green
Write-Host ""

Start-Sleep -Seconds 2
try {
    $metrics1 = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth}
    
    Write-Host "After 1 query (cache miss):" -ForegroundColor Yellow
    Write-Host "  prefetch_hot_domains: $($metrics1.prefetch_hot_domains)" -ForegroundColor $(if ($metrics1.prefetch_hot_domains -eq 0) { "Green" } else { "Red" })
    
    if ($metrics1.prefetch_hot_domains -eq 0) {
        Write-Host "  ✓ Correct: Domain NOT hot (below threshold)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Error: Domain should NOT be hot yet" -ForegroundColor Red
    }
} catch {
    Write-Host "[ERROR] Failed to get metrics" -ForegroundColor Red
}
Write-Host ""

# Test 2: Second query (cache hit) - should IMMEDIATELY be hot
Write-Host "=== Test 2: Second Query (Cache Hit) ===" -ForegroundColor Cyan
Write-Host "Querying again (cache hit)..." -ForegroundColor Yellow
Write-Host "Expected: Domain becomes HOT immediately (bypasses threshold)" -ForegroundColor Yellow
Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue | Out-Null
Write-Host ""

Start-Sleep -Seconds 2
try {
    $metrics2 = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth}
    
    Write-Host "After 2nd query (cache hit):" -ForegroundColor Yellow
    Write-Host "  prefetch_hot_domains: $($metrics2.prefetch_hot_domains)" -ForegroundColor $(if ($metrics2.prefetch_hot_domains -gt 0) { "Green" } else { "Red" })
    
    if ($metrics2.prefetch_hot_domains -gt 0) {
        Write-Host "  ✓✓✓ SUCCESS: Cache hit bypassed threshold!" -ForegroundColor Green
        Write-Host "  Domain became HOT after just 1 cache hit" -ForegroundColor Green
    } else {
        Write-Host "  ✗ FAILED: Cache hit did NOT bypass threshold" -ForegroundColor Red
        Write-Host "  Domain should be hot immediately on cache hit" -ForegroundColor Red
    }
} catch {
    Write-Host "[ERROR] Failed to get metrics" -ForegroundColor Red
}
Write-Host ""

# Stop service
Write-Host "Stopping service..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force
Write-Host "[OK] Service stopped" -ForegroundColor Green
Write-Host ""

Write-Host "=== Test Summary ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Design Logic Verification:" -ForegroundColor Yellow
Write-Host "  1. Cache miss: Requires threshold (2 hits) ✓" -ForegroundColor $(if ($metrics1.prefetch_hot_domains -eq 0) { "Green" } else { "Red" })
Write-Host "  2. Cache hit: Bypasses threshold (immediate) ✓" -ForegroundColor $(if ($metrics2.prefetch_hot_domains -gt 0) { "Green" } else { "Red" })
Write-Host ""

if ($metrics1.prefetch_hot_domains -eq 0 -and $metrics2.prefetch_hot_domains -gt 0) {
    Write-Host "✓✓✓ PERFECT: Implementation matches design logic!" -ForegroundColor Green
} else {
    Write-Host "✗ Implementation does NOT match design logic" -ForegroundColor Red
}
