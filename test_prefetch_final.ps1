# Final Prefetch Test - With full logging
Write-Host "=== Final Prefetch Test ===" -ForegroundColor Cyan
Write-Host "This test will verify the complete prefetch flow" -ForegroundColor Yellow
Write-Host ""

$exeName = "AdGuardHome_optimized.exe"

# Start service and capture output
Write-Host "Starting service..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\$exeName" -PassThru -WindowStyle Hidden -RedirectStandardOutput "service_output.log" -RedirectStandardError "service_error.log"
Start-Sleep -Seconds 5

Write-Host "[OK] Service started (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

# Query domain 5 times
Write-Host "=== Querying www.google.com 5 times ===" -ForegroundColor Cyan
for ($i = 1; $i -le 5; $i++) {
    Write-Host "  Query $i..." -ForegroundColor Gray
    $result = Resolve-DnsName -Name "www.google.com" -Server 127.0.0.1 -Type A -ErrorAction SilentlyContinue
    if ($i -eq 1 -and $result) {
        $ttl = $result[0].TTL
        Write-Host "    TTL: $ttl seconds" -ForegroundColor Cyan
    }
    Start-Sleep -Milliseconds 500
}
Write-Host ""

# Check service logs for "recording domain"
Write-Host "=== Checking if Record() was called ===" -ForegroundColor Cyan
Start-Sleep -Seconds 2
if (Test-Path "service_output.log") {
    $recordingLines = Select-String -Path "service_output.log" -Pattern "recording domain for prefetch"
    if ($recordingLines) {
        Write-Host "[OK] Record() was called $($recordingLines.Count) times!" -ForegroundColor Green
        $recordingLines | Select-Object -First 3 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    } else {
        Write-Host "[ERROR] Record() was NOT called" -ForegroundColor Red
    }
}
Write-Host ""

# Check metrics
Write-Host "=== Checking Prefetch Metrics ===" -ForegroundColor Cyan
Start-Sleep -Seconds 2
$metricsLines = Select-String -Path "service_output.log" -Pattern "prefetch metrics" | Select-Object -Last 1
if ($metricsLines) {
    Write-Host "Latest metrics:" -ForegroundColor Yellow
    Write-Host "  $metricsLines" -ForegroundColor Gray
    
    if ($metricsLines -match "tracked_hits=(\d+)") {
        $trackedHits = $matches[1]
        if ([int]$trackedHits -gt 0) {
            Write-Host "[OK] Domains are being tracked! (tracked_hits=$trackedHits)" -ForegroundColor Green
        } else {
            Write-Host "[PROBLEM] No domains tracked (tracked_hits=0)" -ForegroundColor Red
        }
    }
    
    if ($metricsLines -match "hot_domains=(\d+)") {
        $hotDomains = $matches[1]
        if ([int]$hotDomains -gt 0) {
            Write-Host "[OK] Hot domains detected! (hot_domains=$hotDomains)" -ForegroundColor Green
        } else {
            Write-Host "[INFO] No hot domains yet (hot_domains=0)" -ForegroundColor Yellow
            Write-Host "  This is expected - need to reach threshold (2 hits)" -ForegroundColor Gray
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
Write-Host "Check service_output.log for full details"
