# Prefetch Diagnosis Script
Write-Host "=== Prefetch Mechanism Diagnosis ===" -ForegroundColor Cyan
Write-Host ""

# Start service
Write-Host "Starting service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\AdGuardHome_optimized.exe" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 5

if ($process.HasExited) {
    Write-Host "[ERROR] Service failed to start" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Query domains multiple times
Write-Host "=== Querying test domains (5 times each) ===" -ForegroundColor Cyan
$domains = @("test1.example.com", "test2.example.com", "test3.example.com")
for ($round = 1; $round -le 5; $round++) {
    Write-Host "Round $round..." -ForegroundColor Yellow
    foreach ($domain in $domains) {
        try {
            $result = Resolve-DnsName -Name $domain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue
        } catch {}
    }
    Start-Sleep -Milliseconds 200
}
Write-Host ""

# Check metrics
Write-Host "=== Checking Metrics via API ===" -ForegroundColor Cyan
Start-Sleep -Seconds 2

# Get auth token (base64 of lkxlzx:123456)
$auth = "Basic bGt4bHp4OjEyMzQ1Ng=="

try {
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth} `
        -ErrorAction Stop
    
    Write-Host "Prefetch Metrics:" -ForegroundColor Green
    $metrics.PSObject.Properties | ForEach-Object {
        Write-Host "  $($_.Name): $($_.Value)"
    }
    Write-Host ""
    
    # Analysis
    if ($metrics.tracked_hits -gt 0) {
        Write-Host "[OK] Domains are being tracked!" -ForegroundColor Green
        Write-Host "  Tracked hits: $($metrics.tracked_hits)" -ForegroundColor Green
        Write-Host "  Hot domains: $($metrics.hot_domains)" -ForegroundColor Green
    } else {
        Write-Host "[PROBLEM] No domains tracked - Record() may not be called" -ForegroundColor Red
    }
    
    if ($metrics.tasks_completed -gt 0) {
        Write-Host "[OK] Prefetch tasks are running!" -ForegroundColor Green
    } else {
        Write-Host "[INFO] No prefetch tasks completed yet (normal if TTL is long)" -ForegroundColor Yellow
    }
    
} catch {
    Write-Host "[ERROR] Failed to get metrics: $_" -ForegroundColor Red
    Write-Host "  Make sure you're logged in with correct credentials" -ForegroundColor Yellow
}
Write-Host ""

# Check error log
Write-Host "=== Error Log Content ===" -ForegroundColor Cyan
if (Test-Path "prefetch_errors.log") {
    Get-Content "prefetch_errors.log" | ForEach-Object {
        Write-Host "  $_"
    }
} else {
    Write-Host "  [No error log found]"
}
Write-Host ""

# Stop service
Write-Host "Stopping service..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force
Write-Host "[OK] Service stopped" -ForegroundColor Green
Write-Host ""

Write-Host "=== Diagnosis Summary ===" -ForegroundColor Cyan
Write-Host "If tracked_hits > 0: Record() is working ✓"
Write-Host "If hot_domains > 0: Threshold logic is working ✓"
Write-Host "If tasks_completed = 0: TTL is too long, prefetch hasn't triggered yet"
Write-Host ""
Write-Host "Solution: Wait for cache entries to approach expiration (TTL - 5 seconds)"
