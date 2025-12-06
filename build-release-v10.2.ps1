# AdGuard Home v10.2 正式发布构建脚本
# 确保使用正确的前端和后端构建命令

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "AdGuard Home v10.2 正式发布构建" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 步骤 1: 清理旧的构建文件
Write-Host "[1/4] 清理旧的构建文件..." -ForegroundColor Yellow
Remove-Item -Path "build/static" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "AdGuardHome_v10.2_release.exe" -Force -ErrorAction SilentlyContinue
Write-Host "  ✓ 清理完成" -ForegroundColor Green
Write-Host ""

# 步骤 2: 构建前端（生产模式）
Write-Host "[2/4] 构建前端（生产模式）..." -ForegroundColor Yellow
Write-Host "  命令: npm --prefix client run build-prod" -ForegroundColor Gray
Write-Host "  ⚠️  注意：使用 build-prod 而不是 build-dev" -ForegroundColor Red
Write-Host ""

npm --prefix client run build-prod

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "前端构建失败!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "  ✓ 前端构建完成" -ForegroundColor Green

# 检查前端构建大小
$frontendSize = (Get-ChildItem -Path "build/static" -Recurse -File | Measure-Object -Property Length -Sum).Sum
$frontendSizeMB = [math]::Round($frontendSize / 1MB, 2)

Write-Host "  前端文件大小: $frontendSizeMB MB" -ForegroundColor Cyan

if ($frontendSizeMB -gt 15) {
    Write-Host ""
    Write-Host "  ⚠️  警告：前端文件过大 ($frontendSizeMB MB)" -ForegroundColor Red
    Write-Host "  正常大小应该约 10 MB" -ForegroundColor Yellow
    Write-Host "  可能使用了 build-dev 而不是 build-prod" -ForegroundColor Yellow
    Write-Host ""
    $continue = Read-Host "是否继续构建? (y/n)"
    if ($continue -ne "y") {
        exit 1
    }
}

Write-Host ""

# 步骤 3: 设置 Go 编译环境
Write-Host "[3/4] 设置 Go 编译环境..." -ForegroundColor Yellow
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

$version = "v10.2"
$channel = "release"
$committime = (git log -1 --pretty=%ct)

# 设置 ldflags 来减小文件大小
$ldflags = "-s -w"
$ldflags += " -X github.com/AdguardTeam/AdGuardHome/internal/version.version=$version"
$ldflags += " -X github.com/AdguardTeam/AdGuardHome/internal/version.channel=$channel"
$ldflags += " -X github.com/AdguardTeam/AdGuardHome/internal/version.committime=$committime"

Write-Host "  版本: $version" -ForegroundColor Gray
Write-Host "  通道: $channel" -ForegroundColor Gray
Write-Host "  CGO: 禁用" -ForegroundColor Gray
Write-Host "  优化: -s -w (去除调试符号)" -ForegroundColor Gray
Write-Host "  ✓ 环境设置完成" -ForegroundColor Green
Write-Host ""

# 步骤 4: 编译 Go 程序
Write-Host "[4/4] 编译 Go 程序..." -ForegroundColor Yellow
$output = "AdGuardHome_v10.2_release.exe"

go build -ldflags="$ldflags" -trimpath -o $output

if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ 编译完成" -ForegroundColor Green
    Write-Host ""
    
    # 显示最终文件大小
    $fileSize = (Get-Item $output).Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "构建成功!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "输出文件: $output" -ForegroundColor Cyan
    Write-Host "文件大小: $fileSizeMB MB" -ForegroundColor Cyan
    Write-Host ""
    
    # 大小检查
    if ($fileSizeMB -gt 40) {
        Write-Host "⚠️  警告：文件大小异常 ($fileSizeMB MB)" -ForegroundColor Red
        Write-Host "正常大小应该约 30-32 MB" -ForegroundColor Yellow
        Write-Host "可能的原因：" -ForegroundColor Yellow
        Write-Host "  1. 前端使用了 build-dev 而不是 build-prod" -ForegroundColor Gray
        Write-Host "  2. 编译时没有使用 -ldflags=\"-s -w\"" -ForegroundColor Gray
    } elseif ($fileSizeMB -lt 25) {
        Write-Host "⚠️  警告：文件大小异常 ($fileSizeMB MB)" -ForegroundColor Red
        Write-Host "文件可能不完整或缺少前端资源" -ForegroundColor Yellow
    } else {
        Write-Host "✓ 文件大小正常" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "构建详情:" -ForegroundColor Yellow
    Write-Host "  前端大小: $frontendSizeMB MB" -ForegroundColor Gray
    Write-Host "  可执行文件: $fileSizeMB MB" -ForegroundColor Gray
    Write-Host "  前端占比: $([math]::Round($frontendSizeMB/$fileSizeMB*100, 1))%" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "编译失败!" -ForegroundColor Red
    exit 1
}
