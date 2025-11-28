# DNS路由过滤优先级测试脚本
# 测试广告拦截规则是否能正确作用于DNS路由规则的域名

Write-Host "=== DNS路由过滤优先级测试 ===" -ForegroundColor Cyan
Write-Host ""

# 配置
$baseUrl = "http://localhost:3000"
$username = "admin"
$password = "admin"

# 测试域名 - 假设这些域名同时存在于CN域名列表和广告拦截列表中
$testDomains = @(
    "ad.example.cn",      # 广告域名（应该被拦截，即使在DNS路由规则中）
    "tracker.example.cn", # 追踪域名（应该被拦截）
    "normal.example.cn"   # 正常域名（应该走DNS路由）
)

Write-Host "测试场景：" -ForegroundColor Yellow
Write-Host "1. 启用了DNS路由规则（CN域名使用特定上游）"
Write-Host "2. 启用了广告拦截过滤规则"
Write-Host "3. 测试广告域名是否会被拦截（而不是走DNS路由）"
Write-Host ""

# 测试DNS查询
function Test-DnsQuery {
    param(
        [string]$domain
    )
    
    Write-Host "测试域名: $domain" -ForegroundColor Green
    
    try {
        $result = Resolve-DnsName -Name $domain -Server "127.0.0.1" -Type A -ErrorAction SilentlyContinue
        
        if ($result) {
            Write-Host "  ✓ 查询成功" -ForegroundColor Green
            Write-Host "  IP地址: $($result.IPAddress -join ', ')"
            
            # 检查是否是拦截IP（0.0.0.0）
            if ($result.IPAddress -contains "0.0.0.0") {
                Write-Host "  ✓ 域名被拦截（返回0.0.0.0）" -ForegroundColor Yellow
                return "blocked"
            } else {
                Write-Host "  ✓ 域名未被拦截（正常解析）" -ForegroundColor Cyan
                return "allowed"
            }
        } else {
            Write-Host "  ✗ 查询失败或被拦截（NXDOMAIN）" -ForegroundColor Red
            return "blocked"
        }
    } catch {
        Write-Host "  ✗ 查询出错: $($_.Exception.Message)" -ForegroundColor Red
        return "error"
    }
    
    Write-Host ""
}

# 检查查询日志
function Get-QueryLog {
    param(
        [string]$domain
    )
    
    Write-Host "检查查询日志..." -ForegroundColor Cyan
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/control/querylog" `
            -Method Post `
            -Headers @{
                "Authorization" = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${username}:${password}"))
            } `
            -Body (@{
                "search" = @{
                    "domain" = $domain
                }
                "limit" = 1
            } | ConvertTo-Json) `
            -ContentType "application/json"
        
        if ($response.data -and $response.data.Count -gt 0) {
            $entry = $response.data[0]
            Write-Host "  查询时间: $($entry.time)"
            Write-Host "  域名: $($entry.question.name)"
            Write-Host "  过滤原因: $($entry.reason)"
            
            if ($entry.rules) {
                Write-Host "  匹配规则:" -ForegroundColor Yellow
                foreach ($rule in $entry.rules) {
                    Write-Host "    - $($rule.text) (FilterID: $($rule.filter_list_id))"
                }
            }
            
            if ($entry.upstream_group) {
                Write-Host "  上游组: $($entry.upstream_group)" -ForegroundColor Cyan
            }
        } else {
            Write-Host "  未找到查询记录" -ForegroundColor Gray
        }
    } catch {
        Write-Host "  获取查询日志失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Write-Host ""
}

# 执行测试
Write-Host "开始测试..." -ForegroundColor Cyan
Write-Host ""

$results = @{}

foreach ($domain in $testDomains) {
    $result = Test-DnsQuery -domain $domain
    $results[$domain] = $result
    
    # 获取查询日志详情
    Get-QueryLog -domain $domain
    
    Start-Sleep -Seconds 1
}

# 总结
Write-Host "=== 测试总结 ===" -ForegroundColor Cyan
Write-Host ""

$blocked = ($results.Values | Where-Object { $_ -eq "blocked" }).Count
$allowed = ($results.Values | Where-Object { $_ -eq "allowed" }).Count
$errors = ($results.Values | Where-Object { $_ -eq "error" }).Count

Write-Host "总测试数: $($testDomains.Count)"
Write-Host "被拦截: $blocked" -ForegroundColor Yellow
Write-Host "未拦截: $allowed" -ForegroundColor Green
Write-Host "错误: $errors" -ForegroundColor Red
Write-Host ""

# 验证结果
Write-Host "=== 验证结果 ===" -ForegroundColor Cyan
Write-Host ""

$passed = $true

# 广告域名应该被拦截
if ($results["ad.example.cn"] -ne "blocked") {
    Write-Host "✗ 失败: ad.example.cn 应该被拦截但未被拦截" -ForegroundColor Red
    $passed = $false
} else {
    Write-Host "✓ 通过: ad.example.cn 正确被拦截" -ForegroundColor Green
}

if ($results["tracker.example.cn"] -ne "blocked") {
    Write-Host "✗ 失败: tracker.example.cn 应该被拦截但未被拦截" -ForegroundColor Red
    $passed = $false
} else {
    Write-Host "✓ 通过: tracker.example.cn 正确被拦截" -ForegroundColor Green
}

# 正常域名应该走DNS路由
if ($results["normal.example.cn"] -eq "blocked") {
    Write-Host "✗ 失败: normal.example.cn 不应该被拦截" -ForegroundColor Red
    $passed = $false
} else {
    Write-Host "✓ 通过: normal.example.cn 未被拦截" -ForegroundColor Green
}

Write-Host ""

if ($passed) {
    Write-Host "=== 所有测试通过！===" -ForegroundColor Green
    Write-Host "广告拦截规则正确优先于DNS路由规则" -ForegroundColor Green
} else {
    Write-Host "=== 测试失败 ===" -ForegroundColor Red
    Write-Host "请检查过滤规则优先级配置" -ForegroundColor Red
}

Write-Host ""
Write-Host "测试完成！" -ForegroundColor Cyan
