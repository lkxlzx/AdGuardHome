# Prefetch 功能测试脚本
# 用法: .\test_prefetch.ps1

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║        AdGuard Home - Prefetch Function Test              ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# 测试域名列表
$testDomains = @(
    "google.com",
    "youtube.com",
    "facebook.com",
    "twitter.com",
    "github.com"
)

$dnsServer = "127.0.0.1"
$repeatCount = 10  # 每个域名查询次数

Write-Host "Test Configuration:" -ForegroundColor Yellow
Write-Host "  DNS Server: $dnsServer"
Write-Host "  Test Domains: $($testDomains.Count)"
Write-Host "  Queries per domain: $repeatCount"
Write-Host "  Total queries: $($testDomains.Count * $repeatCount)"
Write-Host ""

# 检查 Prefetch 状态
Write-Host "Checking Prefetch status..." -ForegroundColor Cyan
try {
    $status = Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status" -Method Get
    
    if (-not $status.enabled) {
        Write-Host "⚠️  Warning: Prefetch is disabled!" -ForegroundColor Yellow
        Write-Host "   Enable it in Web UI (Settings -> DNS Settings -> DNS Cache Config)" -ForegroundColor Yellow
        Write-Host ""
        $continue = Read-Host "Continue anyway? (y/n)"
        if ($continue -ne "y") {
            exit
        }
    } else {
        Write-Host "✓ Prefetch is enabled" -ForegroundColor Green
        Write-Host "  Threshold: $($status.threshold) hits"
        Write-Host "  Time Window: $($status.time_window)s"
        Write-Host ""
    }
} catch {
    Write-Host "⚠️  Warning: Could not check Prefetch status" -ForegroundColor Yellow
    Write-Host "   Make sure AdGuard Home is running" -ForegroundColor Yellow
    Write-Host ""
}

Write-Host "Starting DNS queries..." -ForegroundColor Cyan
Write-Host ""

$totalQueries = 0
$successQueries = 0
$failedQueries = 0

foreach ($domain in $testDomains) {
    Write-Host "Testing: $domain" -ForegroundColor Yellow
    
    for ($i = 1; $i -le $repeatCount; $i++) {
        $totalQueries++
        
        try {
            # 使用 nslookup 查询
            $result = nslookup $domain $dnsServer 2>&1
            
            if ($LASTEXITCODE -eq 0) {
                $successQueries++
                Write-Host "  [$i/$repeatCount] ✓" -ForegroundColor Green -NoNewline
            } else {
                $failedQueries++
                Write-Host "  [$i/$repeatCount] ✗" -ForegroundColor Red -NoNewline
            }
        } catch {
            $failedQueries++
            Write-Host "  [$i/$repeatCount] ✗" -ForegroundColor Red -NoNewline
        }
        
        # 每 5 个查询换行
        if ($i % 5 -eq 0) {
            Write-Host ""
        }
        
        # 短暂延迟
        Start-Sleep -Milliseconds 100
    }
    
    Write-Host ""
    Write-Host ""
}

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                      Test Results                          ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""
Write-Host "Total Queries:    $totalQueries"
Write-Host "Successful:       $successQueries" -ForegroundColor Green
Write-Host "Failed:           $failedQueries" -ForegroundColor $(if ($failedQueries -gt 0) { "Red" } else { "Green" })
Write-Host ""

# 再次检查 Prefetch 状态
Write-Host "Checking Prefetch status after test..." -ForegroundColor Cyan
try {
    $statusAfter = Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status" -Method Get
    
    if ($statusAfter.enabled) {
        Write-Host ""
        Write-Host "Prefetch Metrics:" -ForegroundColor Yellow
        Write-Host "  Tracked Domains:  $($statusAfter.tracked_domains)"
        Write-Host "  Hot Domains:      $($statusAfter.hot_domains)"
        Write-Host "  Tasks Completed:  $($statusAfter.tasks_completed)"
        Write-Host ""
        
        if ($statusAfter.hot_domains -gt 0) {
            Write-Host "✓ Success! $($statusAfter.hot_domains) domains are now marked as hot!" -ForegroundColor Green
            Write-Host "  These domains will be automatically refreshed before cache expiry." -ForegroundColor Green
        } else {
            Write-Host "ℹ️  No hot domains yet." -ForegroundColor Yellow
            Write-Host "   Domains need $($statusAfter.threshold) hits within $($statusAfter.time_window)s to become hot." -ForegroundColor Yellow
            Write-Host "   Try running this script again or wait a moment." -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "⚠️  Could not fetch final status" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Test completed!" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Run .\monitor_prefetch.ps1 to see real-time metrics"
Write-Host "  2. Check Web UI: http://localhost:3000"
Write-Host "  3. View logs for detailed information"
Write-Host ""
