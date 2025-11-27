# LRU增量清理性能测试脚本
# 测试增量清理在大规模场景下的性能表现

Write-Host "=== LRU增量清理性能测试 ===" -ForegroundColor Cyan
Write-Host ""

# 配置
$adguardExe = ".\AdGuardHome_lru_optimized.exe"
$configFile = "AdGuardHome.yaml"
$testDuration = 180  # 测试持续时间（秒）
$queryRate = 50      # 每秒查询数
$uniqueDomains = 15000  # 唯一域名数量（超过maxEntries以触发清理）

# 检查可执行文件
if (-not (Test-Path $adguardExe)) {
    Write-Host "错误: 找不到 $adguardExe" -ForegroundColor Red
    exit 1
}

# 检查配置文件
if (-not (Test-Path $configFile)) {
    Write-Host "错误: 找不到 $configFile" -ForegroundColor Red
    exit 1
}

Write-Host "测试配置:" -ForegroundColor Yellow
Write-Host "  - 可执行文件: $adguardExe"
Write-Host "  - 测试时长: $testDuration 秒"
Write-Host "  - 查询速率: $queryRate 查询/秒"
Write-Host "  - 唯一域名: $uniqueDomains 个"
Write-Host ""

# 启动AdGuard Home
Write-Host "启动 AdGuard Home..." -ForegroundColor Green
$process = Start-Process -FilePath $adguardExe -ArgumentList "-c", $configFile, "--no-check-update" -PassThru -WindowStyle Hidden

# 等待启动
Write-Host "等待服务启动..."
Start-Sleep -Seconds 5

# 检查进程是否运行
if ($process.HasExited) {
    Write-Host "错误: AdGuard Home 启动失败" -ForegroundColor Red
    exit 1
}

Write-Host "AdGuard Home 已启动 (PID: $($process.Id))" -ForegroundColor Green
Write-Host ""

try {
    # 生成测试域名列表
    Write-Host "生成测试域名列表..." -ForegroundColor Yellow
    $domains = @()
    for ($i = 1; $i -le $uniqueDomains; $i++) {
        $domains += "test-lru-$i.example.com"
    }
    Write-Host "已生成 $($domains.Count) 个测试域名" -ForegroundColor Green
    Write-Host ""

    # 阶段1: 填充缓存（触发清理）
    Write-Host "=== 阶段1: 填充缓存 ===" -ForegroundColor Cyan
    $phase1Duration = 60
    $phase1Queries = 0
    $phase1Errors = 0
    $startTime = Get-Date

    Write-Host "开始发送查询（持续 $phase1Duration 秒）..."
    while (((Get-Date) - $startTime).TotalSeconds -lt $phase1Duration) {
        # 随机选择域名
        $domain = $domains | Get-Random
        
        try {
            $result = Resolve-DnsName -Name $domain -Server 127.0.0.1 -DnsOnly -ErrorAction SilentlyContinue
            $phase1Queries++
        } catch {
            $phase1Errors++
        }
        
        # 控制查询速率
        Start-Sleep -Milliseconds (1000 / $queryRate)
        
        # 每10秒显示进度
        if ($phase1Queries % ($queryRate * 10) -eq 0) {
            $elapsed = ((Get-Date) - $startTime).TotalSeconds
            Write-Host "  进度: $([int]$elapsed)秒 | 查询: $phase1Queries | 错误: $phase1Errors"
        }
    }

    Write-Host ""
    Write-Host "阶段1完成:" -ForegroundColor Green
    Write-Host "  - 总查询数: $phase1Queries"
    Write-Host "  - 错误数: $phase1Errors"
    Write-Host "  - 成功率: $([math]::Round(($phase1Queries - $phase1Errors) / $phase1Queries * 100, 2))%"
    Write-Host ""

    # 获取初始统计
    Write-Host "获取清理前统计..." -ForegroundColor Yellow
    try {
        $statsUrl = "http://127.0.0.1:3000/control/stats"
        $auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("admin:admin"))
        $headers = @{
            "Authorization" = "Basic $auth"
        }
        $statsBefore = Invoke-RestMethod -Uri $statsUrl -Headers $headers -Method Get
        Write-Host "  - DNS查询总数: $($statsBefore.num_dns_queries)" -ForegroundColor Cyan
    } catch {
        Write-Host "  警告: 无法获取统计信息" -ForegroundColor Yellow
    }
    Write-Host ""

    # 阶段2: 持续查询（观察增量清理）
    Write-Host "=== 阶段2: 持续查询（观察增量清理） ===" -ForegroundColor Cyan
    $phase2Duration = $testDuration - $phase1Duration
    $phase2Queries = 0
    $phase2Errors = 0
    $startTime = Get-Date

    Write-Host "继续发送查询（持续 $phase2Duration 秒）..."
    Write-Host "观察日志中的 'incremental cleanup completed' 消息..." -ForegroundColor Yellow
    Write-Host ""

    while (((Get-Date) - $startTime).TotalSeconds -lt $phase2Duration) {
        # 随机选择域名（偏向热门域名）
        if ((Get-Random -Minimum 0 -Maximum 100) -lt 70) {
            # 70%概率选择前1000个域名（热门域名）
            $domain = $domains | Select-Object -First 1000 | Get-Random
        } else {
            # 30%概率选择所有域名
            $domain = $domains | Get-Random
        }
        
        try {
            $result = Resolve-DnsName -Name $domain -Server 127.0.0.1 -DnsOnly -ErrorAction SilentlyContinue
            $phase2Queries++
        } catch {
            $phase2Errors++
        }
        
        # 控制查询速率
        Start-Sleep -Milliseconds (1000 / $queryRate)
        
        # 每10秒显示进度
        if ($phase2Queries % ($queryRate * 10) -eq 0) {
            $elapsed = ((Get-Date) - $startTime).TotalSeconds
            Write-Host "  进度: $([int]$elapsed)秒 | 查询: $phase2Queries | 错误: $phase2Errors"
        }
    }

    Write-Host ""
    Write-Host "阶段2完成:" -ForegroundColor Green
    Write-Host "  - 总查询数: $phase2Queries"
    Write-Host "  - 错误数: $phase2Errors"
    Write-Host "  - 成功率: $([math]::Round(($phase2Queries - $phase2Errors) / $phase2Queries * 100, 2))%"
    Write-Host ""

    # 获取最终统计
    Write-Host "获取清理后统计..." -ForegroundColor Yellow
    try {
        $statsAfter = Invoke-RestMethod -Uri $statsUrl -Headers $headers -Method Get
        Write-Host "  - DNS查询总数: $($statsAfter.num_dns_queries)" -ForegroundColor Cyan
        Write-Host "  - 新增查询: $($statsAfter.num_dns_queries - $statsBefore.num_dns_queries)" -ForegroundColor Cyan
    } catch {
        Write-Host "  警告: 无法获取统计信息" -ForegroundColor Yellow
    }
    Write-Host ""

    # 总结
    Write-Host "=== 测试总结 ===" -ForegroundColor Cyan
    $totalQueries = $phase1Queries + $phase2Queries
    $totalErrors = $phase1Errors + $phase2Errors
    Write-Host "总查询数: $totalQueries"
    Write-Host "总错误数: $totalErrors"
    Write-Host "总成功率: $([math]::Round(($totalQueries - $totalErrors) / $totalQueries * 100, 2))%"
    Write-Host ""
    Write-Host "请检查日志文件以查看增量清理的详细信息:" -ForegroundColor Yellow
    Write-Host "  - 查找 'incremental cleanup completed' 消息"
    Write-Host "  - 查找 'performing full cleanup after incremental rounds' 消息"
    Write-Host "  - 观察 'shards_cleaned', 'removed', 'remaining' 等指标"
    Write-Host ""

} finally {
    # 停止AdGuard Home
    Write-Host "停止 AdGuard Home..." -ForegroundColor Yellow
    Stop-Process -Id $process.Id -Force
    Write-Host "测试完成" -ForegroundColor Green
}
