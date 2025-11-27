# LRU增量清理性能基准测试
# 对比优化前后的性能差异

param(
    [int]$Queries = 10000,
    [int]$UniqueDomains = 15000
)

Write-Host "=== LRU增量清理性能基准测试 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "测试参数:" -ForegroundColor Yellow
Write-Host "  - 查询数量: $Queries"
Write-Host "  - 唯一域名: $UniqueDomains (超过maxEntries 10000)"
Write-Host ""

# 生成测试域名
Write-Host "生成测试域名..." -ForegroundColor Yellow
$domains = @()
for ($i = 1; $i -le $UniqueDomains; $i++) {
    $domains += "bench-$i.example.com"
}
Write-Host "已生成 $($domains.Count) 个域名" -ForegroundColor Green
Write-Host ""

# 测试函数
function Test-Performance {
    param(
        [string]$ExePath,
        [string]$TestName
    )
    
    Write-Host "=== 测试: $TestName ===" -ForegroundColor Cyan
    
    # 启动服务
    Write-Host "启动服务..." -ForegroundColor Yellow
    $process = Start-Process -FilePath $ExePath -ArgumentList "-c", "AdGuardHome.yaml", "--no-check-update" -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 5
    
    if ($process.HasExited) {
        Write-Host "错误: 服务启动失败" -ForegroundColor Red
        return $null
    }
    
    Write-Host "服务已启动 (PID: $($process.Id))" -ForegroundColor Green
    
    # 发送查询
    Write-Host "发送 $Queries 个查询..." -ForegroundColor Yellow
    $startTime = Get-Date
    $success = 0
    $failed = 0
    
    for ($i = 1; $i -le $Queries; $i++) {
        $domain = $domains | Get-Random
        try {
            $null = Resolve-DnsName -Name $domain -Server 127.0.0.1 -DnsOnly -ErrorAction SilentlyContinue
            $success++
        } catch {
            $failed++
        }
        
        if ($i % 1000 -eq 0) {
            Write-Host "  进度: $i / $Queries" -ForegroundColor Gray
        }
    }
    
    $duration = (Get-Date) - $startTime
    
    # 停止服务
    Write-Host "停止服务..." -ForegroundColor Yellow
    Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
    
    # 返回结果
    $result = @{
        TestName = $TestName
        Duration = $duration.TotalSeconds
        QPS = [math]::Round($Queries / $duration.TotalSeconds, 2)
        Success = $success
        Failed = $failed
    }
    
    Write-Host "完成!" -ForegroundColor Green
    Write-Host "  - 耗时: $($result.Duration) 秒"
    Write-Host "  - QPS: $($result.QPS)"
    Write-Host "  - 成功: $success"
    Write-Host "  - 失败: $failed"
    Write-Host ""
    
    return $result
}

# 检查可执行文件
$optimizedExe = ".\AdGuardHome_lru_optimized.exe"
$originalExe = ".\AdGuardHome.exe"

if (-not (Test-Path $optimizedExe)) {
    Write-Host "错误: 找不到 $optimizedExe" -ForegroundColor Red
    Write-Host "请先运行: go build -o AdGuardHome_lru_optimized.exe" -ForegroundColor Yellow
    exit 1
}

# 测试优化版本
$optimizedResult = Test-Performance -ExePath $optimizedExe -TestName "LRU增量清理版本"

# 如果有原始版本，也测试它
if (Test-Path $originalExe) {
    Write-Host "发现原始版本，进行对比测试..." -ForegroundColor Yellow
    Write-Host ""
    $originalResult = Test-Performance -ExePath $originalExe -TestName "原始版本"
    
    # 对比结果
    Write-Host "=== 性能对比 ===" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "原始版本:" -ForegroundColor Yellow
    Write-Host "  - 耗时: $($originalResult.Duration) 秒"
    Write-Host "  - QPS: $($originalResult.QPS)"
    Write-Host ""
    Write-Host "LRU优化版本:" -ForegroundColor Yellow
    Write-Host "  - 耗时: $($optimizedResult.Duration) 秒"
    Write-Host "  - QPS: $($optimizedResult.QPS)"
    Write-Host ""
    
    $improvement = [math]::Round((($originalResult.Duration - $optimizedResult.Duration) / $originalResult.Duration) * 100, 2)
    if ($improvement -gt 0) {
        Write-Host "性能提升: $improvement%" -ForegroundColor Green
    } else {
        Write-Host "性能变化: $improvement%" -ForegroundColor Yellow
    }
} else {
    Write-Host "未找到原始版本 ($originalExe)，跳过对比测试" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "测试完成！" -ForegroundColor Green
