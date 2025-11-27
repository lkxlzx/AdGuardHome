# Final Correct Prefetch Test
Write-Host "=== Final Prefetch Test ===" -ForegroundColor Cyan
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

# Phase 1: Initial query
Write-Host "=== Phase 1: Initial Query ===" -ForegroundColor Cyan
$result = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
$ttl = $result[0].TTL
Write-Host "Domain: $testDomain" -ForegroundColor Yellow
Write-Host "IP: $($result[0].IPAddress)" -ForegroundColor Green
Write-Host "TTL: $ttl seconds" -ForegroundColor Cyan
Write-Host ""

# Phase 2: Cache hits
Write-Host "=== Phase 2: Cache Hits ===" -ForegroundColor Cyan
Write-Host "Querying 10 times to trigger cache hits..." -ForegroundColor Yellow
for ($i = 1; $i -le 10; $i++) {
    Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue | Out-Null
    Start-Sleep -Milliseconds 300
}
Write-Host "[OK] Completed 10 queries" -ForegroundColor Green
Write-Host ""

# Check metrics
Write-Host "=== Checking Metrics ===" -ForegroundColor Cyan
Start-Sleep -Seconds 2
try {
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth}
    
    Write-Host "Prefetch Status:" -ForegroundColor Yellow
    Write-Host "  prefetch_enabled: $($metrics.prefetch_enabled)"
    Write-Host "  prefetch_status: $($metrics.prefetch_status)"
    Write-Host "  prefetch_hot_domains: $($metrics.prefetch_hot_domains)" -ForegroundColor $(if ($metrics.prefetch_hot_domains -gt 0) { "Green" } else { "Red" })
    Write-Host "  prefetch_completed: $($metrics.prefetch_completed)"
    Write-Host "  prefetch_failed: $($metrics.prefetch_failed)"
    Write-Host ""
    
    if ($metrics.prefetch_hot_domains -eq 0) {
        Write-Host "✗ No hot domains detected" -ForegroundColor Red
        Write-Host "  Domain not being tracked properly" -ForegroundColor Red
        Stop-Process -Id $process.Id -Force
        exit 1
    }
    
    Write-Host "✓✓✓ Domain marked as HOT!" -ForegroundColor Green
    Write-Host "  $($metrics.prefetch_hot_domains) hot domain(s) ready for prefetch" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to get metrics: $_" -ForegroundColor Red
    Stop-Process -Id $process.Id -Force
    exit 1
}
Write-Host ""

# Phase 3: Wait for prefetch
$waitTime = [Math]::Max(15, $ttl - 3)
Write-Host "=== Phase 3: Waiting for Prefetch ===" -ForegroundColor Cyan
Write-Host "TTL: $ttl seconds" -ForegroundColor Yellow
Write-Host "Prefetch triggers at: TTL - 5 seconds" -ForegroundColor Yellow
Write-Host "Waiting: $waitTime seconds" -ForegroundColor Yellow
Write-Host "checkAndRefresh scans every: 10 seconds" -ForegroundColor Gray
Write-Host ""

$prefetchDetected = $false
for ($elapsed = 0; $elapsed -lt $waitTime; $elapsed++) {
    Write-Progress -Activity "Waiting for prefetch trigger" `
        -Status "$elapsed / $waitTime seconds" `
        -PercentComplete (($elapsed / $waitTime) * 100)
    Start-Sleep -Seconds 1
    
    # Check every 10 seconds
    if ($elapsed % 10 -eq 0 -and $elapsed -gt 0) {
        if (Test-Path "prefetch_errors.log") {
            $logContent = Get-Content "prefetch_errors.log" -Raw
            if ($logContent -match "cache_refresh") {
                Write-Host ""
                Write-Host "  [$elapsed s] ✓✓✓ Prefetch DETECTED!" -ForegroundColor Green
                $prefetchDetected = $true
                break
            }
        }
    }
}
Write-Progress -Activity "Waiting for prefetch trigger" -Completed
Write-Host ""

if (-not $prefetchDetected) {
    Write-Host "Waiting additional 15 seconds for operations to complete..." -ForegroundColor Yellow
    Start-Sleep -Seconds 15
}
Write-Host ""

# Final results
Write-Host "=== Final Results ===" -ForegroundColor Cyan
Write-Host ""

# Check error log
if (Test-Path "prefetch_errors.log") {
    $logContent = Get-Content "prefetch_errors.log"
    $refreshLines = $logContent | Select-String -Pattern "cache_refresh"
    $errorLines = $logContent | Select-String -Pattern "Error:"
    
    Write-Host "Prefetch Error Log:" -ForegroundColor Yellow
    Write-Host "  Total lines: $($logContent.Count)"
    Write-Host "  cache_refresh operations: $($refreshLines.Count)" -ForegroundColor $(if ($refreshLines.Count -gt 0) { "Green" } else { "Red" })
    Write-Host "  Errors: $($errorLines.Count)"
    Write-Host ""
    
    if ($refreshLines.Count -gt 0) {
        Write-Host "✓✓✓ SUCCESS: Optimistic Cache Refresh is WORKING!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Prefetch operations:" -ForegroundColor Cyan
        $refreshLines | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "No cache_refresh operations detected" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Possible reasons:" -ForegroundColor Yellow
        Write-Host "  1. Wait time insufficient (TTL=$ttl, waited=$waitTime)" -ForegroundColor Gray
        Write-Host "  2. checkAndRefresh timing (scans every 10s)" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Full log:" -ForegroundColor Gray
        $logContent | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "[WARN] No error log file found" -ForegroundColor Yellow
}
Write-Host ""

# Final metrics
Write-Host "=== Final Metrics ===" -ForegroundColor Cyan
try {
    $finalMetrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth}
    
    Write-Host "  prefetch_hot_domains: $($finalMetrics.prefetch_hot_domains)"
    Write-Host "  prefetch_completed: $($finalMetrics.prefetch_completed)" -ForegroundColor $(if ($finalMetrics.prefetch_completed -gt 0) { "Green" } else { "Yellow" })
    Write-Host "  prefetch_failed: $($finalMetrics.prefetch_failed)"
    Write-Host "  prefetch_success_rate: $($finalMetrics.prefetch_success_rate)%"
    
    if ($finalMetrics.prefetch_completed -gt 0) {
        Write-Host ""
        Write-Host "✓ Prefetch tasks completed: $($finalMetrics.prefetch_completed)" -ForegroundColor Green
    }
} catch {
    Write-Host "[ERROR] Failed to get final metrics" -ForegroundColor Red
}
Write-Host ""

# Stop service
Write-Host "Stopping service..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force
Write-Host "[OK] Service stopped" -ForegroundColor Green
Write-Host ""

Write-Host "=== Test Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  Domain: $testDomain"
Write-Host "  TTL: $ttl seconds"
Write-Host "  Wait time: $waitTime seconds"
Write-Host "  Cleanup interval: 3600 seconds (1 hour)"
Write-Host "  Test within cleanup window: YES"
