# Correct Prefetch Test - Ensures continuous hits within time window
Write-Host "=== Correct Prefetch Test ===" -ForegroundColor Cyan
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"
$threshold = 2  # From config
$testDomain = "www.google.com"  # Use single domain for clarity

# Check executable
if (-not (Test-Path $exeName)) {
    Write-Host "[ERROR] $exeName not found" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Found: $exeName" -ForegroundColor Green

# Backup error log
if (Test-Path "prefetch_errors.log") {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    Move-Item "prefetch_errors.log" "prefetch_errors_old_$timestamp.log" -Force
}

# Start service
Write-Host "Starting service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\$exeName" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 5

if ($process.HasExited) {
    Write-Host "[ERROR] Service failed to start" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Phase 1: Build hit count (continuous queries within time window)
Write-Host "=== Phase 1: Building Hit Count (Continuous Queries) ===" -ForegroundColor Cyan
Write-Host "Threshold: $threshold hits required" -ForegroundColor Yellow
Write-Host "Time window: 6 minutes (from config)" -ForegroundColor Yellow
Write-Host ""

$ttl = 0
Write-Host "Querying $testDomain continuously..." -ForegroundColor Yellow
for ($i = 1; $i -le ($threshold + 3); $i++) {
    Write-Host "  Query $i..." -ForegroundColor Gray
    try {
        $result = Resolve-DnsName -Name $testDomain -Server 127.0.0.1 -Type A -ErrorAction Stop
        if ($i -eq 1 -and $result) {
            $ttl = $result[0].TTL
            Write-Host "    TTL: $ttl seconds" -ForegroundColor Cyan
        }
        Write-Host "    IP: $($result[0].IPAddress)" -ForegroundColor Green
    } catch {
        Write-Host "    Query failed" -ForegroundColor Red
    }
    Start-Sleep -Milliseconds 500
}

Write-Host ""
Write-Host "[OK] Domain queried $($threshold + 3) times continuously" -ForegroundColor Green
Write-Host "[INFO] Domain should now be tracked as 'hot' and scheduled for prefetch" -ForegroundColor Cyan
Write-Host ""

# Phase 2: Wait for cache to approach expiration
if ($ttl -gt 0) {
    # Calculate wait time: TTL - 3 seconds (to be in the 5-second prefetch window)
    $waitTime = [Math]::Max(5, $ttl - 3)
    
    Write-Host "=== Phase 2: Waiting for Prefetch Trigger ===" -ForegroundColor Cyan
    Write-Host "Cache TTL: $ttl seconds" -ForegroundColor Yellow
    Write-Host "Prefetch triggers: 5 seconds before expiration" -ForegroundColor Yellow
    Write-Host "Waiting: $waitTime seconds (TTL - 3s)" -ForegroundColor Yellow
    Write-Host ""
    
    # Progress with periodic checks
    $checkInterval = 10
    for ($elapsed = 0; $elapsed -lt $waitTime; $elapsed++) {
        $remaining = $waitTime - $elapsed
        $percent = [Math]::Round(($elapsed / $waitTime) * 100)
        
        Write-Progress -Activity "Waiting for prefetch trigger" `
            -Status "$elapsed / $waitTime seconds ($percent%) - $remaining seconds remaining" `
            -PercentComplete $percent
        
        Start-Sleep -Seconds 1
        
        # Check log every 10 seconds
        if ($elapsed % $checkInterval -eq 0 -and $elapsed -gt 0) {
            if (Test-Path "prefetch_errors.log") {
                $logContent = Get-Content "prefetch_errors.log" -Raw
                if ($logContent -match "cache_refresh") {
                    Write-Host ""
                    Write-Host "  [$elapsed s] ✓ Prefetch detected!" -ForegroundColor Green
                }
            }
        }
    }
    Write-Progress -Activity "Waiting for prefetch trigger" -Completed
    Write-Host ""
    Write-Host "[OK] Wait complete - cache should be in prefetch window" -ForegroundColor Green
} else {
    Write-Host "[WARN] Could not determine TTL, waiting 60 seconds..." -ForegroundColor Yellow
    Start-Sleep -Seconds 60
}
Write-Host ""

# Phase 3: Additional wait for prefetch to complete
Write-Host "=== Phase 3: Waiting for Prefetch Completion ===" -ForegroundColor Cyan
Write-Host "Waiting 15 seconds for prefetch operations to complete..." -ForegroundColor Yellow
Start-Sleep -Seconds 15
Write-Host ""

# Results
Write-Host "=== Results ===" -ForegroundColor Cyan
Write-Host ""

# Analyze error log
Write-Host "Error Log Analysis:" -ForegroundColor Yellow
if (Test-Path "prefetch_errors.log") {
    $logContent = Get-Content "prefetch_errors.log"
    $cacheRefreshLines = $logContent | Select-String -Pattern "cache_refresh"
    $errorLines = $logContent | Select-String -Pattern "Error:"
    $successLines = $logContent | Select-String -Pattern "SUCCESS"
    
    Write-Host "  Total lines: $($logContent.Count)"
    Write-Host "  cache_refresh operations: $($cacheRefreshLines.Count)" -ForegroundColor $(if ($cacheRefreshLines.Count -gt 0) { "Green" } else { "Red" })
    Write-Host "  Errors: $($errorLines.Count)"
    Write-Host "  Successes: $($successLines.Count)"
    Write-Host ""
    
    if ($cacheRefreshLines.Count -gt 0) {
        Write-Host "✓ SUCCESS: Optimistic cache refresh is working!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Sample cache_refresh operations:" -ForegroundColor Cyan
        $cacheRefreshLines | Select-Object -First 10 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "✗ PROBLEM: No cache_refresh operations detected" -ForegroundColor Red
        Write-Host ""
        Write-Host "Full log:" -ForegroundColor Yellow
        $logContent | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
        Write-Host ""
        Write-Host "Possible issues:" -ForegroundColor Yellow
        Write-Host "  1. Domain not marked as 'hot' (check debug logs)" -ForegroundColor Gray
        Write-Host "  2. Prefetch manager not started" -ForegroundColor Gray
        Write-Host "  3. checkAndRefresh not scanning at the right time" -ForegroundColor Gray
    }
} else {
    Write-Host "  [ERROR] Error log file not found" -ForegroundColor Red
}
Write-Host ""

# Stop service
Write-Host "=== Cleanup ===" -ForegroundColor Cyan
try {
    Stop-Process -Id $process.Id -Force -ErrorAction Stop
    Write-Host "[OK] Service stopped" -ForegroundColor Green
} catch {
    Write-Host "[WARN] Service may have already stopped" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "=== Test Summary ===" -ForegroundColor Cyan
Write-Host "  Domain: $testDomain"
Write-Host "  Queries: $($threshold + 3) (continuous)"
Write-Host "  TTL: $ttl seconds"
Write-Host "  Wait time: $waitTime seconds"
Write-Host ""
Write-Host "Check prefetch_errors.log for detailed operation logs"
