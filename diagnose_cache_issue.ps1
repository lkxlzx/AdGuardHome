# DNS 缓存问题诊断脚本
# 用于诊断 DNS 路由匹配域名的缓存命中率低的问题

param(
    [string]$Domain = "shankapi.ifeng.com",
    [string]$DnsServer = "127.0.0.1",
    [int]$TestCount = 5
)

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║          DNS Cache Issue Diagnostic Tool                  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Host "Test Configuration:" -ForegroundColor Yellow
Write-Host "  Domain:      $Domain"
Write-Host "  DNS Server:  $DnsServer"
Write-Host "  Test Count:  $TestCount"
Write-Host ""

# 测试不同的查询类型
$queryTypes = @("A", "AAAA", "HTTPS")

foreach ($type in $queryTypes) {
    Write-Host "Testing $type records for $Domain" -ForegroundColor Cyan
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    
    for ($i = 1; $i -le $TestCount; $i++) {
        $startTime = Get-Date
        
        try {
            # 使用 Resolve-DnsName 进行查询
            $result = Resolve-DnsName -Name $Domain -Server $DnsServer -Type $type -ErrorAction SilentlyContinue
            
            $endTime = Get-Date
            $duration = ($endTime - $startTime).TotalMilliseconds
            
            if ($result) {
                $recordCount = ($result | Measure-Object).Count
                Write-Host "  [$i] ✓ " -NoNewline -ForegroundColor Green
                Write-Host "$recordCount records, " -NoNewline
                Write-Host "$([math]::Round($duration, 2))ms" -ForegroundColor $(if ($duration -lt 10) { "Green" } elseif ($duration -lt 50) { "Yellow" } else { "Red" })
            } else {
                Write-Host "  [$i] ○ " -NoNewline -ForegroundColor Yellow
                Write-Host "No records (NXDOMAIN/NODATA), " -NoNewline
                Write-Host "$([math]::Round($duration, 2))ms" -ForegroundColor $(if ($duration -lt 10) { "Green" } elseif ($duration -lt 50) { "Yellow" } else { "Red" })
            }
        } catch {
            $endTime = Get-Date
            $duration = ($endTime - $startTime).TotalMilliseconds
            Write-Host "  [$i] ✗ " -NoNewline -ForegroundColor Red
            Write-Host "Query failed, $([math]::Round($duration, 2))ms" -ForegroundColor Red
        }
        
        # 短暂延迟
        Start-Sleep -Milliseconds 500
    }
    
    Write-Host ""
}

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                    Analysis                                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Host "Expected Behavior:" -ForegroundColor Yellow
Write-Host "  • First query: Higher latency (50-200ms) - Cache miss"
Write-Host "  • Subsequent queries: Low latency (<10ms) - Cache hit"
Write-Host ""

Write-Host "If you see:" -ForegroundColor Yellow
Write-Host "  • All queries have high latency (>50ms)" -ForegroundColor Red
Write-Host "    → Cache is not working for this domain/type"
Write-Host ""
Write-Host "  • AAAA/HTTPS queries always high latency" -ForegroundColor Red
Write-Host "    → Empty responses (NODATA) may not be cached properly"
Write-Host ""
Write-Host "  • A queries cached but AAAA/HTTPS not" -ForegroundColor Red
Write-Host "    → Different caching behavior for different record types"
Write-Host ""

Write-Host "Possible Causes:" -ForegroundColor Yellow
Write-Host "  1. DNS routing uses separate cache per upstream group"
Write-Host "  2. Empty responses (NODATA) have very short TTL"
Write-Host "  3. HTTPS record type may not be cached by dnsproxy"
Write-Host "  4. CustomUpstreamConfig cache not shared with main cache"
Write-Host ""

Write-Host "Recommendations:" -ForegroundColor Green
Write-Host "  1. Check AdGuard Home logs for cache hit/miss info"
Write-Host "  2. Verify cache_enabled is true in configuration"
Write-Host "  3. Check if upstream returns proper TTL for empty responses"
Write-Host "  4. Consider increasing cache_ttl_min for empty responses"
Write-Host ""

# 检查 AdGuard Home 配置
Write-Host "Checking AdGuard Home Configuration..." -ForegroundColor Cyan
try {
    $config = Invoke-RestMethod -Uri "http://localhost:3000/control/dns_info" -Method Get -ErrorAction Stop
    
    Write-Host "  Cache Enabled:    " -NoNewline
    if ($config.cache_enabled) {
        Write-Host "✓ Yes" -ForegroundColor Green
    } else {
        Write-Host "✗ No" -ForegroundColor Red
        Write-Host "    → Enable cache in Web UI!" -ForegroundColor Yellow
    }
    
    Write-Host "  Cache Size:       $($config.cache_size) bytes"
    Write-Host "  Cache Min TTL:    $($config.cache_ttl_min) seconds"
    Write-Host "  Cache Max TTL:    $($config.cache_ttl_max) seconds"
    Write-Host "  Cache Optimistic: $($config.cache_optimistic)"
    
} catch {
    Write-Host "  ⚠ Could not fetch configuration" -ForegroundColor Yellow
    Write-Host "    Make sure AdGuard Home is running" -ForegroundColor Gray
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
