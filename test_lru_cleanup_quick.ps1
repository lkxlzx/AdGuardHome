# LRU增量清理快速测试脚本
# 快速验证增量清理功能是否正常工作

Write-Host "=== LRU增量清理快速测试 ===" -ForegroundColor Cyan
Write-Host ""

# 配置
$adguardExe = ".\AdGuardHome_lru_optimized.exe"
$configFile = "AdGuardHome.yaml"
$testDuration = 60  # 测试持续时间（秒）
$queryRate = 100    # 每秒查询数
$uniqueDomains = 12000  # 唯一域名数量（超过默认maxEntries 10000）

# 检查可执行文件
if (-not (Test-Path $adguardExe)) {
    Write-Host "错误: 找不到 $adguardExe" -ForegroundColor Red
    Write-Host "请先运行: go build -o AdGuardHome_lru_optimized.exe" -ForegroundColor Yellow
    exit 1
}

Write-Host "测试配置:" -ForegroundColor Yellow
Write-Host "  - 测试时长: $testDuration 秒"
Write-Host "  - 查询速率: $queryRate 查询/秒"
Write-Host "  - 唯一域名: $uniqueDomains 个（超过maxEntries触发清理）"
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

# 生成测试域名列表
Write-Host "生成测试域名..." -ForegroundColor Yellow
$domains = @()
for ($i = 1; $i -le $uniqueDomains; $i++) {
    $domains += "lru-test-$i.example.com"
}
Write-Host "已生成 $($domains.Count) 个测试域名" -ForegroundColor Green
Write-Host ""

# 发送查询
Write-Host "开始发送查询（持续 $testDuration 秒）..." -ForegroundColor Cyan
$queries = 0
$errors = 0
$startTime = Get-Date

while (((Get-Date) - $startTime).TotalSeconds -lt $testDuration) {
    # 随机选择域名
    $domain = $domains | Get-Random
    
    try {
        $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -DnsOnly -ErrorAction SilentlyContinue
        $queries++
    } catch {
        $errors++
    }
    
    # 控制查询速率
    Start-Sleep -Milliseconds (1000 / $queryRate)
    
    # 每5秒显示进度
    if ($queries % ($queryRate * 5) -eq 0) {
        $elapsed = ((Get-Date) - $startTime).TotalSeconds
        Write-Host "  进度: $([int]$elapsed)秒 | 查询: $queries | 错误: $errors" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "测试完成!" -ForegroundColor Green
Write-Host "  - 总查询数: $queries"
Write-Host "  - 错误数: $errors"
if ($queries -gt 0) {
    Write-Host "  - 成功率: $([math]::Round(($queries - $errors) / $queries * 100, 2))%"
}
Write-Host ""

# 等待一下让清理运行
Write-Host "等待清理运行..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Write-Host ""
Write-Host "=== 验证结果 ===" -ForegroundColor Cyan
Write-Host "请检查以下内容:" -ForegroundColor Yellow
Write-Host "1. 查看日志中是否有 'incremental cleanup completed' 消息"
Write-Host "2. 检查 'shards_cleaned' 是否为 4（16个shard的25%）"
Write-Host "3. 观察 'removed' 和 'remaining' 数量"
Write-Host "4. 每4轮后应该看到 'performing full cleanup after incremental rounds'"
Write-Host ""
Write-Host "如果看到这些消息，说明LRU增量清理正常工作！" -ForegroundColor Green
Write-Host ""

# 停止AdGuard Home
Write-Host "停止 AdGuard Home..." -ForegroundColor Yellow
Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
Write-Host "测试完成" -ForegroundColor Green
