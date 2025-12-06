# AdGuard Home v10.2 优化构建脚本
# 使用正确的编译标志来减小可执行文件大小

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "AdGuard Home v10.2 优化构建" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 设置环境变量
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# 获取版本信息
$version = "v10.2"
$channel = "development"
$committime = (git log -1 --pretty=%ct)

# 设置 ldflags 来减小文件大小
$ldflags = "-s -w"
$ldflags += " -X github.com/AdguardTeam/AdGuardHome/internal/version.version=$version"
$ldflags += " -X github.com/AdguardTeam/AdGuardHome/internal/version.channel=$channel"
$ldflags += " -X github.com/AdguardTeam/AdGuardHome/internal/version.committime=$committime"

Write-Host "构建参数:" -ForegroundColor Yellow
Write-Host "  版本: $version" -ForegroundColor Gray
Write-Host "  通道: $channel" -ForegroundColor Gray
Write-Host "  CGO: 禁用" -ForegroundColor Gray
Write-Host "  优化: 去除调试符号 (-s -w)" -ForegroundColor Gray
Write-Host ""

# 构建
Write-Host "开始构建..." -ForegroundColor Green
$output = "AdGuardHome_v10.2_optimized.exe"

go build -ldflags="$ldflags" -trimpath -o $output

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "构建成功!" -ForegroundColor Green
    
    # 显示文件大小
    $fileSize = (Get-Item $output).Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    
    Write-Host "输出文件: $output" -ForegroundColor Cyan
    Write-Host "文件大小: $fileSizeMB MB" -ForegroundColor Cyan
    
    # 与旧版本比较
    if (Test-Path "AdGuardHome_v10.2_test.exe") {
        $oldSize = (Get-Item "AdGuardHome_v10.2_test.exe").Length
        $oldSizeMB = [math]::Round($oldSize / 1MB, 2)
        $diff = $oldSize - $fileSize
        $diffMB = [math]::Round($diff / 1MB, 2)
        
        Write-Host ""
        Write-Host "与未优化版本比较:" -ForegroundColor Yellow
        Write-Host "  旧版本: $oldSizeMB MB" -ForegroundColor Gray
        Write-Host "  新版本: $fileSizeMB MB" -ForegroundColor Gray
        Write-Host "  减少: $diffMB MB" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "提示: 可以使用以下命令启动:" -ForegroundColor Yellow
    Write-Host "  .\$output" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "构建失败!" -ForegroundColor Red
    exit 1
}
