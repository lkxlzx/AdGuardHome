# Long Wait Prefetch Test Script
# Tests the optimistic cache refresh mechanism with extended wait time

Write-Host "=== Long Wait Prefetch Test ===" -ForegroundColor Cyan
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"
$threshold = 2
$testDomains = @("www.google.com", "www.baidu.com", "github.com")

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
    Write-Host "[OK] Backed up old error log" -ForegroundColor Green
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

# Phase 1: Query domains to reach threshold
Write-Host "=== Phase 1: Building Hit Count ===" -ForegroundColor Cyan
Write-Host "Querying each domain $($threshold + 3) times..." -ForegroundColor Yellow

$ttlValues = @{}
for ($round = 1; $round -le ($threshold + 3); $round++) {
    Write-Host "  Round $round..." -ForegroundColor Gray
    foreach ($domain in $testDomains) {
        try {
            $result = Resolve-DnsName -Name $domain -Server 127.0.0.1 -Type A -ErrorAction Stop
            # Capture TTL from first query
            if ($round -eq 1 -and $result) {
                $ttlValues[$domain] = $result[0].TTL
                Write-Host "    $domain - TTL: $($result[0].TTL)s" -ForegroundColor Cyan
            }
        } catch {
            Write-Host "    $domain - Query failed" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 500
}
Write-Host "[OK] All domains queried multiple times" -ForegroundColor Green
Write-Host ""

# Calculate wait time based on minimum TTL
$minTTL = ($ttlValues.Values | Measure-Object -Minimum).Minimum
if ($minTTL -gt 0) {
    # Wait until 3 seconds before expiration (prefetch triggers at 5s before)
    # This ensures we're in the prefetch window
    $waitTime = [Math]::Max(5, $minTTL - 3)
    Write-Host "=== Phase 2: Waiting for Cache to Approach Expiration ===" -ForegroundColor Cyan
    Write-Host "Minimum TTL detected: $minTTL seconds" -ForegroundColor Yellow
    Write-Host "Waiting $waitTime seconds (TTL - 3s)..." -ForegroundColor Yellow
    Write-Host "This will put us in the prefetch window (5s before expiration)" -ForegroundColor Gray
    Write-Host ""
    
    # Progress bar
    $startTime = Get-Date
    for ($i = 0; $i -lt $waitTime; $i++) {
        $elapsed = $i + 1
        $remaining = $waitTime - $elapsed
        $percent = [Math]::Round(($elapsed / $waitTime) * 100)
        
        Write-Progress -Activity "Waiting for cache expiration" `
            -Status "$elapsed / $waitTime seconds ($percent%)" `
            -PercentComplete $percent
        
        Start-Sleep -Seconds 1
        
        # Check every 30 seconds
        if ($i % 30 -eq 0 -and $i -gt 0) {
            Write-Host ""
            Write-Host "  [$elapsed/$waitTime] Checking error log..." -ForegroundColor Gray
            if (Test-Path "prefetch_errors.log") {
                $logLines = Get-Content "prefetch_errors.log"
                $cacheRefreshCount = ($logLines | Select-String -Pattern "cache_refresh" -AllMatches).Count
                if ($cacheRefreshCount -gt 0) {
                    Write-Host "  [OK] Detected $cacheRefreshCount cache_refresh operations!" -ForegroundColor Green
                }
            }
        }
    }
    Write-Progress -Activity "Waiting for cache expiration" -Completed
    Write-Host ""
    Write-Host "[OK] Wait complete" -ForegroundColor Green
} else {
    Write-Host "[WARN] Could not determine TTL, waiting 60 seconds..." -ForegroundColor Yellow
    Start-Sleep -Seconds 60
}
Write-Host ""

# Phase 3: Trigger additional queries to ensure prefetch happens
Write-Host "=== Phase 3: Final Queries (trigger prefetch) ===" -ForegroundColor Cyan
Write-Host "Querying domains again to trigger prefetch..." -ForegroundColor Yellow
foreach ($domain in $testDomains) {
    try {
        $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -Type A -ErrorAction Stop
        Write-Host "  $domain - OK" -ForegroundColor Green
    } catch {}
}
Write-Host ""

# Wait a bit for prefetch to complete
Write-Host "Waiting 10 seconds for prefetch operations to complete..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
Write-Host ""

# Check results
Write-Host "=== Results ===" -ForegroundColor Cyan
Write-Host ""

# Error log analysis
Write-Host "Error Log Analysis:" -ForegroundColor Yellow
if (Test-Path "prefetch_errors.log") {
    $logContent = Get-Content "prefetch_errors.log"
    $totalLines = $logContent.Count
    $errorCount = ($logContent | Select-String -Pattern "Error:" -AllMatches).Count
    $successCount = ($logContent | Select-String -Pattern "SUCCESS" -AllMatches).Count
    $cacheRefreshCount = ($logContent | Select-String -Pattern "cache_refresh" -AllMatches).Count
    
    Write-Host "  Total lines: $totalLines"
    Write-Host "  Errors: $errorCount"
    Write-Host "  Successes: $successCount"
    Write-Host "  cache_refresh operations: $cacheRefreshCount"
    Write-Host ""
    
    if ($cacheRefreshCount -gt 0) {
        Write-Host "[SUCCESS] Optimistic cache refresh is working!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Sample log entries:" -ForegroundColor Yellow
        $logContent | Select-String -Pattern "cache_refresh" | Select-Object -First 5 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Cyan
        }
    } else {
        Write-Host "[PROBLEM] No cache_refresh detected" -ForegroundColor Red
        Write-Host ""
        Write-Host "Full log content:" -ForegroundColor Yellow
        $logContent | ForEach-Object {
            Write-Host "  $_"
        }
    }
} else {
    Write-Host "  [WARN] Error log file not found" -ForegroundColor Yellow
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

Write-Host "=== Test Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  - Domains tested: $($testDomains.Count)"
Write-Host "  - Queries per domain: $($threshold + 3)"
Write-Host "  - Wait time: $waitTime seconds"
Write-Host "  - Check prefetch_errors.log for details"
