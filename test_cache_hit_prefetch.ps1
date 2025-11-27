# Test Cache Hit Prefetch
# This test verifies that cache hits trigger prefetch recording

Write-Host "=== Cache Hit Prefetch Test ===" -ForegroundColor Cyan
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"
$testDomain = "www.google.com"

# Backup error log
if (Test-Path "prefetch_errors.log") {
    Remove-Item "prefetch_errors.log" -Force
}

# Start service
Write-Host "Starting service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\$exeName" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 5
Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Phase 1: First query (cache miss)
Write-Host "=== Phase 1: First Query (Cache Miss) ===" -ForegroundColor Cyan
Write-Host "Querying $testDomain..." -ForegroundColor Yellow
$result1 = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
$ttl = $result1[0].TTL
Write-Host "  IP: $($result1[0].IPAddress)" -ForegroundColor Green
Write-Host "  TTL: $ttl seconds" -ForegroundColor Cyan
Write-Host ""

# Phase 2: Second query (cache hit)
Write-Host "=== Phase 2: Second Query (Cache Hit) ===" -ForegroundColor Cyan
Write-Host "Querying $testDomain again (should hit cache)..." -ForegroundColor Yellow
Start-Sleep -Seconds 1
$result2 = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
Write-Host "  IP: $($result2[0].IPAddress)" -ForegroundColor Green
Write-Host "  TTL: $($result2[0].TTL) seconds (decreased)" -ForegroundColor Cyan
Write-Host ""

# Phase 3: Multiple cache hits
Write-Host "=== Phase 3: Multiple Cache Hits ===" -ForegroundColor Cyan
Write-Host "Querying 5 more times to ensure cache hits..." -ForegroundColor Yellow
for ($i = 1; $i -le 5; $i++) {
    Write-Host "  Query $i..." -ForegroundColor Gray
    Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue | Out-Null
    Start-Sleep -Milliseconds 500
}
Write-Host ""

# Wait a bit
Write-Host "Waiting 5 seconds..." -ForegroundColor Yellow
Start-Sleep -Seconds 5
Write-Host ""

# Check metrics
Write-Host "=== Checking Metrics ===" -ForegroundColor Cyan
try {
    $auth = "Basic bGt4bHp4OjEyMzQ1Ng=="
    $metrics = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = $auth} `
        -ErrorAction Stop
    
    Write-Host "Prefetch Metrics:" -ForegroundColor Yellow
    Write-Host "  tracked_hits: $($metrics.tracked_hits)" -ForegroundColor $(if ($metrics.tracked_hits -gt 0) { "Green" } else { "Red" })
    Write-Host "  hot_domains: $($metrics.hot_domains)" -ForegroundColor $(if ($metrics.hot_domains -gt 0) { "Green" } else { "Yellow" })
    Write-Host "  tracked_domains: $($metrics.tracked_domains)"
    Write-Host ""
    
    if ($metrics.tracked_hits -gt 0) {
        Write-Host "✓ SUCCESS: Cache hits are being recorded for prefetch!" -ForegroundColor Green
        Write-Host "  Domain is being tracked with $($metrics.tracked_hits) hit(s)" -ForegroundColor Green
    } else {
        Write-Host "✗ PROBLEM: No hits tracked" -ForegroundColor Red
    }
    
    if ($metrics.hot_domains -gt 0) {
        Write-Host "✓ Domain marked as HOT and ready for prefetch!" -ForegroundColor Green
    } else {
        Write-Host "  Domain not yet marked as hot (may need more hits)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to get metrics: $_" -ForegroundColor Red
}
Write-Host ""

# Now wait for cache to approach expiration
$waitTime = [Math]::Max(10, $ttl - 10)
Write-Host "=== Waiting for Prefetch Trigger ===" -ForegroundColor Cyan
Write-Host "Cache TTL: $ttl seconds" -ForegroundColor Yellow
Write-Host "Waiting $waitTime seconds for cache to approach expiration..." -ForegroundColor Yellow
Write-Host ""

for ($elapsed = 0; $elapsed -lt $waitTime; $elapsed++) {
    $remaining = $waitTime - $elapsed
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
                Write-Host "  [$elapsed s] ✓ Prefetch triggered!" -ForegroundColor Green
            }
        }
    }
}
Write-Progress -Activity "Waiting for prefetch" -Completed
Write-Host ""

# Wait for prefetch to complete
Write-Host "Waiting 10 seconds for prefetch operations to complete..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
Write-Host ""

# Check results
Write-Host "=== Final Results ===" -ForegroundColor Cyan
if (Test-Path "prefetch_errors.log") {
    $logContent = Get-Content "prefetch_errors.log"
    $cacheRefreshCount = ($logContent | Select-String -Pattern "cache_refresh").Count
    
    Write-Host "Error Log:" -ForegroundColor Yellow
    Write-Host "  Total lines: $($logContent.Count)"
    Write-Host "  cache_refresh operations: $cacheRefreshCount" -ForegroundColor $(if ($cacheRefreshCount -gt 0) { "Green" } else { "Red" })
    Write-Host ""
    
    if ($cacheRefreshCount -gt 0) {
        Write-Host "✓✓✓ SUCCESS: Optimistic cache refresh is working!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Sample operations:" -ForegroundColor Cyan
        $logContent | Select-String -Pattern "cache_refresh" | Select-Object -First 5 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "✗ No cache_refresh detected" -ForegroundColor Red
        Write-Host ""
        Write-Host "Log content:" -ForegroundColor Yellow
        $logContent | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "[WARN] No error log found" -ForegroundColor Yellow
}
Write-Host ""

# Stop service
Write-Host "Stopping service..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force
Write-Host "[OK] Service stopped" -ForegroundColor Green
Write-Host ""

Write-Host "=== Test Complete ===" -ForegroundColor Cyan
