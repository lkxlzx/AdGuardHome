# Verify Prefetch Recording
# This script verifies that domains are being recorded for prefetch

Write-Host "=== Verify Prefetch Recording ===" -ForegroundColor Cyan
Write-Host "This test verifies the Record() function is being called" -ForegroundColor Yellow
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"
$testDomain = "www.google.com"

# Start service
Write-Host "Starting service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\$exeName" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 5
Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Phase 1: Cache miss
Write-Host "=== Phase 1: Cache Miss ===" -ForegroundColor Cyan
Write-Host "First query (will miss cache)..." -ForegroundColor Yellow
$result1 = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
Write-Host "  IP: $($result1[0].IPAddress)" -ForegroundColor Green
Write-Host "  TTL: $($result1[0].TTL) seconds" -ForegroundColor Cyan
Write-Host ""

# Phase 2: Cache hits
Write-Host "=== Phase 2: Cache Hits ===" -ForegroundColor Cyan
Write-Host "Querying 5 more times (will hit cache)..." -ForegroundColor Yellow
for ($i = 1; $i -le 5; $i++) {
    Write-Host "  Query $i..." -ForegroundColor Gray
    Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue | Out-Null
    Start-Sleep -Milliseconds 500
}
Write-Host ""

# Wait for metrics to update
Write-Host "Waiting 3 seconds for metrics to update..." -ForegroundColor Yellow
Start-Sleep -Seconds 3
Write-Host ""

# Try to get metrics via API (may fail due to auth)
Write-Host "=== Attempting to Get Metrics ===" -ForegroundColor Cyan
$metricsSuccess = $false
try {
    # Try without auth first
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -ErrorAction Stop
    $metricsSuccess = $true
} catch {
    # Try with basic auth
    try {
        $auth = "Basic bGt4bHp4OjEyMzQ1Ng=="
        $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
            -Method Get `
            -Headers @{Authorization = $auth} `
            -ErrorAction Stop
        $metricsSuccess = $true
    } catch {
        Write-Host "[INFO] API not accessible (auth required)" -ForegroundColor Yellow
        Write-Host "  Will check via other methods" -ForegroundColor Gray
    }
}

if ($metricsSuccess) {
    Write-Host "Metrics Retrieved:" -ForegroundColor Green
    Write-Host "  tracked_hits: $($metrics.tracked_hits)"
    Write-Host "  hot_domains: $($metrics.hot_domains)"
    Write-Host "  tracked_domains: $($metrics.tracked_domains)"
    Write-Host ""
    
    if ($metrics.tracked_hits -gt 0) {
        Write-Host "✓✓✓ SUCCESS: Record() is being called!" -ForegroundColor Green
        Write-Host "  Domains are being tracked for prefetch" -ForegroundColor Green
        
        if ($metrics.hot_domains -gt 0) {
            Write-Host "✓ Domain marked as HOT (ready for prefetch)" -ForegroundColor Green
        } else {
            Write-Host "  Domain tracked but not yet hot (need more hits)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ PROBLEM: No domains tracked" -ForegroundColor Red
        Write-Host "  Record() may not be called or cache hits not detected" -ForegroundColor Red
    }
}
Write-Host ""

# Alternative: Check data directory for any prefetch-related files
Write-Host "=== Checking Data Directory ===" -ForegroundColor Cyan
if (Test-Path "data") {
    $dataFiles = Get-ChildItem "data" -Recurse -File | Select-Object -First 10
    if ($dataFiles) {
        Write-Host "Data files found:" -ForegroundColor Yellow
        $dataFiles | ForEach-Object {
            Write-Host "  $($_.Name)" -ForegroundColor Gray
        }
    }
}
Write-Host ""

# Stop service
Write-Host "Stopping service..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force
Start-Sleep -Seconds 2
Write-Host "[OK] Service stopped" -ForegroundColor Green
Write-Host ""

Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test completed. Key findings:" -ForegroundColor Yellow
if ($metricsSuccess) {
    if ($metrics.tracked_hits -gt 0) {
        Write-Host "  ✓ Prefetch recording is WORKING" -ForegroundColor Green
        Write-Host "  ✓ Cache hits are being detected" -ForegroundColor Green
        Write-Host "  ✓ Domains are being tracked" -ForegroundColor Green
        Write-Host ""
        Write-Host "Next step: Wait for cache to approach expiration to see prefetch trigger" -ForegroundColor Cyan
        Write-Host "  (This requires waiting TTL-5 seconds)" -ForegroundColor Gray
    } else {
        Write-Host "  ✗ Prefetch recording NOT working" -ForegroundColor Red
        Write-Host "  Need to debug Record() function" -ForegroundColor Red
    }
} else {
    Write-Host "  ? Could not verify via API (auth required)" -ForegroundColor Yellow
    Write-Host "  Manual verification needed" -ForegroundColor Yellow
}
