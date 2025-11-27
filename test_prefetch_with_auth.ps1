# Test Prefetch with Correct Auth
Write-Host "=== Prefetch Test with Authentication ===" -ForegroundColor Cyan
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"
$testDomain = "www.google.com"
$auth = "Basic bGt4bHp4OjE5ODIxMDEy"  # lkxlzx:19821012

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

# Phase 1: Cache miss
Write-Host "=== Phase 1: Initial Query (Cache Miss) ===" -ForegroundColor Cyan
$result = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
$ttl = $result[0].TTL
Write-Host "Domain: $testDomain" -ForegroundColor Yellow
Write-Host "IP: $($result[0].IPAddress)" -ForegroundColor Green
Write-Host "TTL: $ttl seconds" -ForegroundColor Cyan
Write-Host ""

# Phase 2: Cache hits
Write-Host "=== Phase 2: Cache Hits (10 queries) ===" -ForegroundColor Cyan
for ($i = 1; $i -le 10; $i++) {
    Write-Host "  Query $i..." -ForegroundColor Gray
    Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue | Out-Null
    Start-Sleep -Milliseconds 300
}
Write-Host ""

# Check metrics
Write-Host "=== Checking Metrics ===" -ForegroundColor Cyan
Start-Sleep -Seconds 2
try {
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth} `
        -ErrorAction Stop
    
    Write-Host "Prefetch Metrics:" -ForegroundColor Yellow
    Write-Host "  tracked_hits: $($metrics.tracked_hits)" -ForegroundColor $(if ($metrics.tracked_hits -gt 0) { "Green" } else { "Red" })
    Write-Host "  hot_domains: $($metrics.hot_domains)" -ForegroundColor $(if ($metrics.hot_domains -gt 0) { "Green" } else { "Red" })
    Write-Host "  tracked_domains: $($metrics.tracked_domains)"
    Write-Host "  tasks_completed: $($metrics.tasks_completed)"
    Write-Host "  tasks_failed: $($metrics.tasks_failed)"
    Write-Host ""
    
    if ($metrics.tracked_hits -eq 0) {
        Write-Host "✗ CRITICAL: No hits tracked!" -ForegroundColor Red
        Write-Host "  Record() is not being called" -ForegroundColor Red
        Write-Host ""
        Write-Host "Stopping test - need to fix Record() first" -ForegroundColor Red
        Stop-Process -Id $process.Id -Force
        exit 1
    }
    
    Write-Host "✓ Record() is working! ($($metrics.tracked_hits) hits tracked)" -ForegroundColor Green
    
    if ($metrics.hot_domains -eq 0) {
        Write-Host "⚠ Domain not yet marked as hot" -ForegroundColor Yellow
        Write-Host "  Threshold: 2 hits (from config)" -ForegroundColor Gray
        Write-Host "  Current hits: $($metrics.tracked_hits)" -ForegroundColor Gray
    } else {
        Write-Host "✓ Domain marked as HOT! ($($metrics.hot_domains) hot domains)" -ForegroundColor Green
    }
} catch {
    Write-Host "[ERROR] Failed to get metrics: $_" -ForegroundColor Red
    Stop-Process -Id $process.Id -Force
    exit 1
}
Write-Host ""

# If domain is hot, wait for prefetch
if ($metrics.hot_domains -gt 0) {
    # Calculate wait time: TTL - 3 seconds (to be in prefetch window)
    $waitTime = [Math]::Max(10, $ttl - 3)
    
    Write-Host "=== Waiting for Prefetch Trigger ===" -ForegroundColor Cyan
    Write-Host "Domain is HOT, waiting for cache to approach expiration..." -ForegroundColor Yellow
    Write-Host "TTL: $ttl seconds" -ForegroundColor Yellow
    Write-Host "Prefetch window: TTL - 5 seconds = $($ttl - 5) seconds" -ForegroundColor Yellow
    Write-Host "Waiting: $waitTime seconds" -ForegroundColor Yellow
    Write-Host "checkAndRefresh scans every: 10 seconds" -ForegroundColor Gray
    Write-Host ""
    
    for ($elapsed = 0; $elapsed -lt $waitTime; $elapsed++) {
        Write-Progress -Activity "Waiting for prefetch" `
            -Status "$elapsed / $waitTime seconds" `
            -PercentComplete (($elapsed / $waitTime) * 100)
        Start-Sleep -Seconds 1
        
        # Check every 10 seconds
        if ($elapsed % 10 -eq 0 -and $elapsed -gt 0) {
            if (Test-Path "prefetch_errors.log") {
                $logContent = Get-Content "prefetch_errors.log" -Raw
                if ($logContent -match "cache_refresh") {
                    Write-Host ""
                    Write-Host "  [$elapsed s] ✓✓✓ Prefetch detected!" -ForegroundColor Green
                    break
                }
            }
        }
    }
    Write-Progress -Activity "Waiting for prefetch" -Completed
    Write-Host ""
    
    # Wait for operations to complete
    Write-Host "Waiting 10 seconds for operations to complete..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    Write-Host ""
}

# Final results
Write-Host "=== Final Results ===" -ForegroundColor Cyan
Write-Host ""

# Check error log
if (Test-Path "prefetch_errors.log") {
    $logContent = Get-Content "prefetch_errors.log"
    $refreshLines = $logContent | Select-String -Pattern "cache_refresh"
    
    Write-Host "Prefetch Error Log:" -ForegroundColor Yellow
    Write-Host "  Total lines: $($logContent.Count)"
    Write-Host "  cache_refresh operations: $($refreshLines.Count)" -ForegroundColor $(if ($refreshLines.Count -gt 0) { "Green" } else { "Red" })
    Write-Host ""
    
    if ($refreshLines.Count -gt 0) {
        Write-Host "✓✓✓ SUCCESS: Optimistic cache refresh is WORKING!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Sample operations:" -ForegroundColor Cyan
        $refreshLines | Select-Object -First 5 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "No cache_refresh operations yet" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Full log:" -ForegroundColor Gray
        $logContent | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    }
}
Write-Host ""

# Final metrics
Write-Host "=== Final Metrics ===" -ForegroundColor Cyan
try {
    $finalMetrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth} `
        -ErrorAction Stop
    
    Write-Host "  tracked_hits: $($finalMetrics.tracked_hits)"
    Write-Host "  hot_domains: $($finalMetrics.hot_domains)"
    Write-Host "  tasks_completed: $($finalMetrics.tasks_completed)" -ForegroundColor $(if ($finalMetrics.tasks_completed -gt 0) { "Green" } else { "Yellow" })
    Write-Host "  tasks_failed: $($finalMetrics.tasks_failed)"
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
