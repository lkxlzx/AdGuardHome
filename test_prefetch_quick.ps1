# Quick Prefetch Test
# Tests prefetch with short TTL to see results quickly

Write-Host "=== Quick Prefetch Test ===" -ForegroundColor Cyan
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"

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

# Use a domain that typically has short TTL
# Or query multiple times to trigger cache hits
$testDomain = "www.google.com"

Write-Host "=== Phase 1: Initial Query (Cache Miss) ===" -ForegroundColor Cyan
$result = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
$ttl = $result[0].TTL
Write-Host "Domain: $testDomain" -ForegroundColor Yellow
Write-Host "IP: $($result[0].IPAddress)" -ForegroundColor Green
Write-Host "TTL: $ttl seconds" -ForegroundColor Cyan
Write-Host ""

Write-Host "=== Phase 2: Cache Hits (7 queries) ===" -ForegroundColor Cyan
Write-Host "Querying to trigger cache hits and reach threshold..." -ForegroundColor Yellow
for ($i = 1; $i -le 7; $i++) {
    Write-Host "  Query $i..." -ForegroundColor Gray
    Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue | Out-Null
    Start-Sleep -Milliseconds 300
}
Write-Host ""

# Check metrics immediately
Write-Host "=== Checking Metrics (Immediate) ===" -ForegroundColor Cyan
Start-Sleep -Seconds 2
try {
    $auth = "Basic bGt4bHp4OjEyMzQ1Ng=="
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth} `
        -ErrorAction Stop
    
    Write-Host "Prefetch Status:" -ForegroundColor Yellow
    Write-Host "  tracked_hits: $($metrics.tracked_hits)" -ForegroundColor $(if ($metrics.tracked_hits -gt 0) { "Green" } else { "Red" })
    Write-Host "  hot_domains: $($metrics.hot_domains)" -ForegroundColor $(if ($metrics.hot_domains -gt 0) { "Green" } else { "Red" })
    Write-Host "  tracked_domains: $($metrics.tracked_domains)"
    Write-Host ""
    
    if ($metrics.tracked_hits -eq 0) {
        Write-Host "✗ CRITICAL: No hits tracked - Record() not being called!" -ForegroundColor Red
        Write-Host "  This means the cache hit detection is not working" -ForegroundColor Red
    } elseif ($metrics.hot_domains -eq 0) {
        Write-Host "⚠ WARNING: Hits tracked but domain not marked as hot" -ForegroundColor Yellow
        Write-Host "  tracked_hits: $($metrics.tracked_hits)" -ForegroundColor Yellow
        Write-Host "  Threshold: 2 (from config)" -ForegroundColor Yellow
        Write-Host "  Domain should be hot after 2+ hits" -ForegroundColor Yellow
    } else {
        Write-Host "✓ Domain is HOT and ready for prefetch!" -ForegroundColor Green
        Write-Host "  hot_domains: $($metrics.hot_domains)" -ForegroundColor Green
    }
} catch {
    Write-Host "[ERROR] Failed to get metrics: $_" -ForegroundColor Red
    Write-Host "  Continuing with test..." -ForegroundColor Yellow
}
Write-Host ""

# Calculate smart wait time
# We need to wait until TTL-5 seconds, but not too long
$targetWait = [Math]::Max(10, [Math]::Min($ttl - 5, 60))  # Cap at 60 seconds for quick testing

Write-Host "=== Phase 3: Waiting for Prefetch Window ===" -ForegroundColor Cyan
Write-Host "TTL: $ttl seconds" -ForegroundColor Yellow
Write-Host "Prefetch triggers at: TTL - 5 seconds" -ForegroundColor Yellow
Write-Host "Waiting: $targetWait seconds" -ForegroundColor Yellow
Write-Host "checkAndRefresh scans every: 10 seconds" -ForegroundColor Gray
Write-Host ""

for ($elapsed = 0; $elapsed -lt $targetWait; $elapsed++) {
    Write-Progress -Activity "Waiting for prefetch window" `
        -Status "$elapsed / $targetWait seconds" `
        -PercentComplete (($elapsed / $targetWait) * 100)
    Start-Sleep -Seconds 1
    
    # Check log every 5 seconds
    if ($elapsed % 5 -eq 0 -and $elapsed -gt 0) {
        if (Test-Path "prefetch_errors.log") {
            $logLines = Get-Content "prefetch_errors.log"
            $refreshCount = ($logLines | Select-String -Pattern "cache_refresh").Count
            if ($refreshCount -gt 0) {
                Write-Host ""
                Write-Host "  [$elapsed s] ✓✓✓ Prefetch detected! ($refreshCount operations)" -ForegroundColor Green
                break
            }
        }
    }
}
Write-Progress -Activity "Waiting for prefetch window" -Completed
Write-Host ""

# Final wait for operations to complete
Write-Host "Waiting 10 seconds for any pending operations..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
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
        Write-Host "✓✓✓ SUCCESS: Optimistic cache refresh is WORKING!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Sample operations:" -ForegroundColor Cyan
        $refreshLines | Select-Object -First 5 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "✗ No cache_refresh operations detected" -ForegroundColor Red
        Write-Host ""
        Write-Host "Possible reasons:" -ForegroundColor Yellow
        Write-Host "  1. Wait time too short (TTL=$ttl, waited=$targetWait)" -ForegroundColor Gray
        Write-Host "  2. Domain not marked as hot (check metrics above)" -ForegroundColor Gray
        Write-Host "  3. checkAndRefresh timing issue (scans every 10s)" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Full log:" -ForegroundColor Yellow
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
    $auth = "Basic bGt4bHp4OjEyMzQ1Ng=="
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth} `
        -ErrorAction Stop
    
    Write-Host "  tasks_completed: $($metrics.tasks_completed)" -ForegroundColor $(if ($metrics.tasks_completed -gt 0) { "Green" } else { "Yellow" })
    Write-Host "  tasks_failed: $($metrics.tasks_failed)"
    Write-Host "  hot_domains: $($metrics.hot_domains)"
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
Write-Host "  Wait time: $targetWait seconds"
Write-Host "  Cleanup interval: 1 hour (3600 seconds)"
Write-Host "  Test duration: Well within cleanup interval OK"
