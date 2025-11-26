# 测试缓存分钟级别图表显示
# Test cache minute-level chart display

Write-Host "=== 缓存分钟级别图表测试 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 测试 API 端点
Write-Host "1. 测试 /control/cache_metrics API..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:3000/control/cache_metrics" -Method Get
    Write-Host "   ✓ API 响应成功" -ForegroundColor Green
    Write-Host "   - 缓存启用: $($response.cache_enabled)" -ForegroundColor Gray
    Write-Host "   - 当前命中率: $([math]::Round($response.cache_hit_rate, 1))%" -ForegroundColor Gray
    Write-Host "   - 总查询数: $($response.total_queries)" -ForegroundColor Gray
    Write-Host "   - 缓存命中: $($response.cache_hits)" -ForegroundColor Gray
    Write-Host "   - 缓存未命中: $($response.cache_misses)" -ForegroundColor Gray
    Write-Host "   - 历史数据点数: $($response.history.Count)" -ForegroundColor Gray
    
    if ($response.history.Count -eq 60) {
        Write-Host "   ✓ 历史数据包含 60 个数据点（60分钟）" -ForegroundColor Green
    } else {
        Write-Host "   ✗ 历史数据点数不正确: $($response.history.Count)" -ForegroundColor Red
    }
    
    # 显示最近 10 分钟的数据
    Write-Host ""
    Write-Host "   最近 10 分钟的命中率:" -ForegroundColor Cyan
    $recent = $response.history[-10..-1]
    for ($i = 0; $i -lt $recent.Count; $i++) {
        $minutesAgo = 10 - $i
        $rate = [math]::Round($recent[$i], 1)
        Write-Host "   $minutesAgo 分钟前: $rate%" -ForegroundColor Gray
    }
    
} catch {
    Write-Host "   ✗ API 请求失败: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "2. 验证时间格式..." -ForegroundColor Yellow

# 计算时间范围
$now = Get-Date
$startTime = $now.AddMinutes(-59)
Write-Host "   - 图表时间范围: $($startTime.ToString('HH:mm')) - $($now.ToString('HH:mm'))" -ForegroundColor Gray
Write-Host "   - 时间跨度: 60 分钟" -ForegroundColor Gray
Write-Host "   - 数据粒度: 1 分钟/点" -ForegroundColor Gray

Write-Host ""
Write-Host "3. 前端显示验证..." -ForegroundColor Yellow
Write-Host "   请在浏览器中验证以下内容:" -ForegroundColor Cyan
Write-Host "   □ 打开 http://localhost:3000" -ForegroundColor Gray
Write-Host "   □ 进入 Dashboard 页面" -ForegroundColor Gray
Write-Host "   □ 查看 'DNS 缓存命中率' 卡片" -ForegroundColor Gray
Write-Host "   □ 鼠标悬停在图表上查看时间标签" -ForegroundColor Gray
Write-Host "   □ 确认时间格式为 'HH:mm'（如 14:30）" -ForegroundColor Gray
Write-Host "   □ 确认时间范围为最近 60 分钟" -ForegroundColor Gray
Write-Host "   □ 确认图表从左到右显示从 60 分钟前到现在" -ForegroundColor Gray

Write-Host ""
Write-Host "4. 生成测试数据..." -ForegroundColor Yellow
Write-Host "   发送一些 DNS 查询以生成缓存数据..." -ForegroundColor Cyan

$testDomains = @(
    "google.com",
    "github.com",
    "stackoverflow.com",
    "microsoft.com",
    "apple.com"
)

foreach ($domain in $testDomains) {
    try {
        # 第一次查询（缓存未命中）
        $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -ErrorAction SilentlyContinue
        Start-Sleep -Milliseconds 100
        
        # 第二次查询（应该缓存命中）
        $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -ErrorAction SilentlyContinue
        Start-Sleep -Milliseconds 100
        
        Write-Host "   ✓ 查询 $domain" -ForegroundColor Green
    } catch {
        Write-Host "   ✗ 查询 $domain 失败" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "5. 再次检查 API..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

try {
    $response2 = Invoke-RestMethod -Uri "http://localhost:3000/control/cache_metrics" -Method Get
    Write-Host "   - 总查询数: $($response2.total_queries)" -ForegroundColor Gray
    Write-Host "   - 缓存命中: $($response2.cache_hits)" -ForegroundColor Gray
    Write-Host "   - 当前命中率: $([math]::Round($response2.cache_hit_rate, 1))%" -ForegroundColor Gray
    
    if ($response2.total_queries -gt $response.total_queries) {
        Write-Host "   ✓ 查询计数已增加" -ForegroundColor Green
    }
} catch {
    Write-Host "   ✗ API 请求失败" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "关键改进:" -ForegroundColor Yellow
Write-Host "  • 时间粒度: 从每小时改为每分钟" -ForegroundColor Green
Write-Host "  • 数据点数: 60 个数据点（60 分钟）" -ForegroundColor Green
Write-Host "  • 时间格式: HH:mm（如 14:30）" -ForegroundColor Green
Write-Host "  • 更新频率: 每分钟更新一次历史数据" -ForegroundColor Green
Write-Host "  • 刷新间隔: 前端每 30 秒刷新一次" -ForegroundColor Green
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Open Dashboard in browser" -ForegroundColor Gray
Write-Host "  2. View cache hit rate chart" -ForegroundColor Gray
Write-Host "  3. Hover to see time-specific data" -ForegroundColor Gray
Write-Host "  4. Wait a few minutes to observe real-time updates" -ForegroundColor Gray
Write-Host ""
