# Detailed Prefetch Test Script
# Tests the optimistic cache refresh mechanism with proper hit counting

Write-Host "=== Detailed Prefetch Test ===" -ForegroundColor Cyan
Write-Host ""

# Configuration
$exeName = "AdGuardHome_optimized.exe"
$threshold = 2  # From config
$testDomains = @("www.google.com", "www.baidu.com", "github.com")

# Check if executable exists
if (-not (Test-Path $exeName)) {
    Write-Host "[ERROR] $exeName not found" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Found optimized version: $exeName" -ForegroundColor Green

# Backup old error log
if (Test-Path "prefetch_errors.log") {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    Move-Item "prefetch_errors.log" "prefetch_errors_old_$timestamp.log" -Force
    Write-Host "[OK] Backed up old error log" -ForegroundColor Green
}

# Start the service
Write-Host "Starting optimized version..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\$exeName" -PassThru -WindowStyle Hidden
Write-Host "Waiting for service to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check if process is still running
if ($process.HasExited) {
    Write-Host "[ERROR] Service failed to start" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Phase 1: Initial queries (below threshold)
Write-Host "=== Phase 1: Initial Queries (1 hit each) ===" -ForegroundColor Cyan
foreach ($domain in $testDomains) {
    Write-Host "Query: $domain" -ForegroundColor Yellow
    try {
        $result = Resolve-DnsName -Name $domain -Server 127.0.0.1 -Type A -ErrorAction Stop
        $ip = $result[0].IPAddress
        Write-Host "  [OK] Success: $ip" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] No response" -ForegroundColor Red
    }
}
Write-Host ""

# Phase 2: Repeat queries to reach threshold
Write-Host "=== Phase 2: Reaching Threshold ($threshold hits) ===" -ForegroundColor Cyan
for ($i = 2; $i -le $threshold; $i++) {
    Write-Host "Round $i of $threshold..." -ForegroundColor Yellow
    foreach ($domain in $testDomains) {
        try { $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -Type A -ErrorAction Stop } catch {}
    }
    Start-Sleep -Milliseconds 500
}
Write-Host "[OK] All domains should now be tracked as hot" -ForegroundColor Green
Write-Host ""

# Phase 3: Additional queries to trigger more hits
Write-Host "=== Phase 3: Additional Queries (trigger prefetch) ===" -ForegroundColor Cyan
for ($i = 1; $i -le 3; $i++) {
    Write-Host "Extra round $i..." -ForegroundColor Yellow
    foreach ($domain in $testDomains) {
        try { $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -Type A -ErrorAction Stop } catch {}
    }
    Start-Sleep -Milliseconds 500
}
Write-Host ""

# Wait for prefetch mechanism to work
Write-Host "=== Waiting for Prefetch Mechanism ===" -ForegroundColor Cyan
Write-Host "Waiting 30 seconds for cache entries to approach expiration..." -ForegroundColor Yellow
Start-Sleep -Seconds 30
Write-Host ""

# Check error log
Write-Host "=== Checking Error Log ===" -ForegroundColor Cyan
if (Test-Path "prefetch_errors.log") {
    $logContent = Get-Content "prefetch_errors.log"
    $totalLines = $logContent.Count
    $errorCount = ($logContent | Select-String -Pattern "Error:" -AllMatches).Count
    $successCount = ($logContent | Select-String -Pattern "SUCCESS" -AllMatches).Count
    $cacheRefreshCount = ($logContent | Select-String -Pattern "cache_refresh" -AllMatches).Count
    
    Write-Host "Log Statistics:" -ForegroundColor Yellow
    Write-Host "  Total lines: $totalLines"
    Write-Host "  Error count: $errorCount"
    Write-Host "  Success count: $successCount"
    Write-Host "  Using cache_refresh: $cacheRefreshCount"
    Write-Host ""
    
    Write-Host "Last 20 log entries:" -ForegroundColor Yellow
    $logContent | Select-Object -Last 20 | ForEach-Object {
        Write-Host "  $_"
    }
    Write-Host ""
    
    if ($cacheRefreshCount -gt 0) {
        Write-Host "[OK] cache_refresh mechanism is working!" -ForegroundColor Green
    } else {
        Write-Host "[WARN] cache_refresh not detected" -ForegroundColor Yellow
    }
} else {
    Write-Host "[WARN] Error log file not found" -ForegroundColor Yellow
}
Write-Host ""

# Check prefetch metrics via API
Write-Host "=== Checking Prefetch Metrics ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "http://127.0.0.1/control/prefetch_metrics" `
        -Method Get `
        -Headers @{Authorization = "Basic bGt4bHp4OjEyMzQ1Ng=="} `
        -ErrorAction Stop
    
    Write-Host "Prefetch Metrics:" -ForegroundColor Yellow
    Write-Host "  Tracked hits: $($response.tracked_hits)"
    Write-Host "  Hot domains: $($response.hot_domains)"
    Write-Host "  Tracked domains: $($response.tracked_domains)"
    Write-Host "  Tasks completed: $($response.tasks_completed)"
    Write-Host "  Tasks failed: $($response.tasks_failed)"
    Write-Host "  Current active: $($response.current_active)"
    Write-Host ""
    
    if ($response.hot_domains -gt 0) {
        Write-Host "[OK] Hot domains detected: $($response.hot_domains)" -ForegroundColor Green
    } else {
        Write-Host "[WARN] No hot domains detected" -ForegroundColor Yellow
    }
    
    if ($response.tasks_completed -gt 0) {
        Write-Host "[OK] Prefetch tasks completed: $($response.tasks_completed)" -ForegroundColor Green
    } else {
        Write-Host "[WARN] No prefetch tasks completed yet" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to get metrics: $_" -ForegroundColor Red
}
Write-Host ""

# Stop the service
Write-Host "=== Stopping Service ===" -ForegroundColor Cyan
try {
    Stop-Process -Id $process.Id -Force -ErrorAction Stop
    Write-Host "[OK] Service stopped" -ForegroundColor Green
} catch {
    Write-Host "[WARN] Service may have already stopped" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "=== Test Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Analysis:" -ForegroundColor Yellow
Write-Host "  - Threshold: $threshold hits required"
Write-Host "  - Test domains: $($testDomains.Count)"
Write-Host "  - Total queries per domain: $($threshold + 3)"
Write-Host ""
Write-Host "Tip: Check prefetch_errors.log for detailed prefetch behavior" -ForegroundColor Cyan
