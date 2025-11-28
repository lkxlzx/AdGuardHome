# Build AdGuard Home Filter Priority Fix for all major platforms
# DNS路由过滤优先级修复 - 全平台构建脚本

param(
    [switch]$SkipCompress = $false
)

Write-Host "=== AdGuard Home Filter Priority Fix - 全平台构建 ===" -ForegroundColor Cyan
Write-Host ""

# 创建输出目录
$distDir = "dist_filter_priority_fix"
if (Test-Path $distDir) {
    Write-Host "清理旧的构建目录..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $distDir
}
New-Item -ItemType Directory -Path $distDir | Out-Null

# 获取版本信息
try {
    $version = git describe --tags --abbrev=4 HEAD 2>$null
    if (-not $version) {
        $version = "v0.0.0-filter-priority-fix"
    }
} catch {
    $version = "v0.0.0-filter-priority-fix"
}

Write-Host "版本: $version" -ForegroundColor Green
Write-Host ""

# 构建标志
$ldflags = "-s -w -X main.version=$version"

# 定义要构建的平台
$platforms = @(
    @{OS="windows"; Arch="amd64"; Ext=".exe"; Name="Windows AMD64"},
    @{OS="windows"; Arch="arm64"; Ext=".exe"; Name="Windows ARM64"},
    @{OS="windows"; Arch="386"; Ext=".exe"; Name="Windows 386"},
    
    @{OS="linux"; Arch="amd64"; Ext=""; Name="Linux AMD64"},
    @{OS="linux"; Arch="arm64"; Ext=""; Name="Linux ARM64"},
    @{OS="linux"; Arch="arm"; Ext=""; Name="Linux ARMv7"; Arm="7"},
    @{OS="linux"; Arch="arm"; Ext=""; Name="Linux ARMv6"; Arm="6"},
    @{OS="linux"; Arch="386"; Ext=""; Name="Linux 386"},
    @{OS="linux"; Arch="mips"; Ext=""; Name="Linux MIPS"},
    @{OS="linux"; Arch="mipsle"; Ext=""; Name="Linux MIPSLE"},
    
    @{OS="darwin"; Arch="amd64"; Ext=""; Name="macOS AMD64"},
    @{OS="darwin"; Arch="arm64"; Ext=""; Name="macOS ARM64 (Apple Silicon)"},
    
    @{OS="freebsd"; Arch="amd64"; Ext=""; Name="FreeBSD AMD64"},
    @{OS="freebsd"; Arch="arm64"; Ext=""; Name="FreeBSD ARM64"}
)

$successCount = 0
$failCount = 0
$buildResults = @()

foreach ($platform in $platforms) {
    $os = $platform.OS
    $arch = $platform.Arch
    $ext = $platform.Ext
    $name = $platform.Name
    $arm = $platform.Arm
    
    $outputName = "AdGuardHome_filter_priority_fix_${os}_${arch}${ext}"
    $outputPath = Join-Path $distDir $outputName
    
    Write-Host "构建 $name..." -ForegroundColor Cyan
    
    # 设置环境变量
    $env:GOOS = $os
    $env:GOARCH = $arch
    if ($arm) {
        $env:GOARM = $arm
    } else {
        Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
    }
    
    # 构建
    $buildCmd = "go build -ldflags=`"$ldflags`" -o `"$outputPath`""
    
    try {
        $output = Invoke-Expression $buildCmd 2>&1
        
        if ($LASTEXITCODE -eq 0 -and (Test-Path $outputPath)) {
            $fileSize = (Get-Item $outputPath).Length
            $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
            
            Write-Host "  ✓ 成功 ($fileSizeMB MB)" -ForegroundColor Green
            
            $buildResults += [PSCustomObject]@{
                Platform = $name
                OS = $os
                Arch = $arch
                File = $outputName
                Size = "$fileSizeMB MB"
                Status = "成功"
            }
            
            $successCount++
        } else {
            Write-Host "  ✗ 失败" -ForegroundColor Red
            if ($output) {
                Write-Host "  错误: $output" -ForegroundColor Red
            }
            
            $buildResults += [PSCustomObject]@{
                Platform = $name
                OS = $os
                Arch = $arch
                File = $outputName
                Size = "N/A"
                Status = "失败"
            }
            
            $failCount++
        }
    } catch {
        Write-Host "  ✗ 失败: $($_.Exception.Message)" -ForegroundColor Red
        
        $buildResults += [PSCustomObject]@{
            Platform = $name
            OS = $os
            Arch = $arch
            File = $outputName
            Size = "N/A"
            Status = "失败"
        }
        
        $failCount++
    }
    
    Write-Host ""
}

# 清理环境变量
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\GOARM -ErrorAction SilentlyContinue

# 显示构建结果
Write-Host "=== 构建结果 ===" -ForegroundColor Cyan
Write-Host ""
$buildResults | Format-Table -AutoSize

Write-Host ""
Write-Host "总计: $($platforms.Count) 个平台" -ForegroundColor Cyan
Write-Host "成功: $successCount" -ForegroundColor Green
Write-Host "失败: $failCount" -ForegroundColor $(if ($failCount -gt 0) { "Red" } else { "Green" })
Write-Host ""

# 计算总大小
if ($successCount -gt 0) {
    $totalSize = (Get-ChildItem $distDir -File | Measure-Object -Property Length -Sum).Sum
    $totalSizeMB = [math]::Round($totalSize / 1MB, 2)
    Write-Host "总大小: $totalSizeMB MB" -ForegroundColor Cyan
    Write-Host ""
}

# 压缩文件（可选）
if (-not $SkipCompress -and $successCount -gt 0) {
    Write-Host "=== 压缩文件 ===" -ForegroundColor Cyan
    Write-Host ""
    
    $archiveName = "AdGuardHome_filter_priority_fix_${version}_all_platforms.zip"
    $archivePath = Join-Path (Get-Location) $archiveName
    
    Write-Host "创建压缩包: $archiveName" -ForegroundColor Yellow
    
    try {
        if (Test-Path $archivePath) {
            Remove-Item $archivePath -Force
        }
        
        Compress-Archive -Path "$distDir\*" -DestinationPath $archivePath -CompressionLevel Optimal
        
        $archiveSize = (Get-Item $archivePath).Length
        $archiveSizeMB = [math]::Round($archiveSize / 1MB, 2)
        
        Write-Host "压缩完成 ($archiveSizeMB MB)" -ForegroundColor Green
        Write-Host ""
    } catch {
        Write-Host "压缩失败: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
    }
}

# 创建README
$readmePath = Join-Path $distDir "README.txt"
$readmeContent = @"
AdGuard Home - DNS路由过滤优先级修复版本
Filter Priority Fix Version

版本: $version
构建时间: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

=== 修复内容 ===

修复了DNS路由规则优先级过高导致广告拦截功能失效的问题。

问题: DNS路由使用白名单模板，优先级高于黑名单过滤规则
影响: CN域名列表中的广告域名不会被拦截
解决: 调整过滤规则检查顺序，黑名单规则现在优先于DNS路由规则

=== 新的优先级顺序 ===

1. 白名单规则（最高优先级）
2. 黑名单过滤规则（广告拦截等）
3. DNS路由规则（最低优先级）

=== 使用方法 ===

1. 选择适合你系统的可执行文件
2. 停止当前运行的AdGuardHome
3. 使用修复版本替换原文件
4. 启动AdGuardHome

=== 平台说明 ===

Windows:
  - amd64: 64位Windows系统（推荐）
  - arm64: ARM64 Windows系统
  - 386: 32位Windows系统

Linux:
  - amd64: 64位Linux系统（推荐）
  - arm64: ARM64 Linux系统（如树莓派4）
  - armv7: ARMv7 Linux系统（如树莓派3）
  - armv6: ARMv6 Linux系统（如树莓派1）
  - 386: 32位Linux系统
  - mips/mipsle: MIPS路由器

macOS:
  - amd64: Intel Mac
  - arm64: Apple Silicon Mac (M1/M2/M3)

FreeBSD:
  - amd64: 64位FreeBSD系统
  - arm64: ARM64 FreeBSD系统

=== 验证修复 ===

测试一个同时在DNS路由规则和广告拦截列表中的域名：

Windows:
  nslookup ad.example.cn 127.0.0.1

Linux/macOS:
  dig @127.0.0.1 ad.example.cn

预期结果: 域名被拦截（返回0.0.0.0或NXDOMAIN）

=== 文档 ===

详细文档请查看源代码目录中的：
- 快速使用指南.md
- DNS路由过滤优先级修复说明.md
- README_FILTER_PRIORITY_FIX.md

=== 兼容性 ===

✓ 白名单规则仍然具有最高优先级
✓ DNS路由在关闭过滤功能时仍然工作
✓ 现有配置无需修改
✓ 完全向后兼容

=== 技术支持 ===

如有问题，请查看文档或提交issue。

构建信息:
- 成功: $successCount 个平台
- 失败: $failCount 个平台
- 总大小: $totalSizeMB MB

"@

Set-Content -Path $readmePath -Value $readmeContent -Encoding UTF8

Write-Host "=== 构建完成 ===" -ForegroundColor Green
Write-Host ""
Write-Host "输出目录: $distDir" -ForegroundColor Cyan
Write-Host ""

if ($successCount -gt 0) {
    Write-Host "可执行文件列表:" -ForegroundColor Cyan
    Get-ChildItem $distDir -File -Exclude "README.txt" | Select-Object Name, @{Label="大小";Expression={"{0:N2} MB" -f ($_.Length/1MB)}} | Format-Table -AutoSize
}

Write-Host ""
Write-Host "使用方法:" -ForegroundColor Yellow
Write-Host "1. 选择适合你系统的可执行文件" -ForegroundColor White
Write-Host "2. 替换原AdGuardHome可执行文件" -ForegroundColor White
Write-Host "3. 启动AdGuardHome" -ForegroundColor White
Write-Host ""

if ($failCount -gt 0) {
    Write-Host "警告: 有 $failCount 个平台构建失败" -ForegroundColor Red
    exit 1
} else {
    Write-Host "所有平台构建成功！" -ForegroundColor Green
    exit 0
}
