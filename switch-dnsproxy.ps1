# DNSProxy 开发模式切换脚本
# 用于在本地开发和远程版本之间快速切换

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("local", "remote")]
    [string]$Mode,
    
    [string]$Version = "v0.79.2"
)

$ErrorActionPreference = "Stop"

$goModPath = "go.mod"

Write-Host ""
Write-Host "=== DNSProxy Mode Switcher ===" -ForegroundColor Cyan
Write-Host ""

# 读取 go.mod
$content = Get-Content $goModPath -Raw

if ($Mode -eq "local") {
    Write-Host "Switching to LOCAL development mode..." -ForegroundColor Yellow
    Write-Host "  Using: ./vendor-dev/dnsproxy" -ForegroundColor Gray
    
    # 检查本地 dnsproxy 是否存在
    if (-not (Test-Path "vendor-dev/dnsproxy")) {
        Write-Host ""
        Write-Host "ERROR: vendor-dev/dnsproxy directory not found!" -ForegroundColor Red
        Write-Host ""
        Write-Host "Cloning dnsproxy into workspace..." -ForegroundColor Yellow
        git clone -b v0.79.2 https://github.com/lkxlzx/dnsproxy.git vendor-dev/dnsproxy
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host ""
            Write-Host "ERROR: Failed to clone dnsproxy!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "Successfully cloned dnsproxy!" -ForegroundColor Green
        Write-Host ""
    }
    
    # 替换为本地路径
    $content = $content -replace 'replace github\.com/AdguardTeam/dnsproxy => github\.com/lkxlzx/dnsproxy .*', 'replace github.com/AdguardTeam/dnsproxy => ./vendor-dev/dnsproxy'
    
} else {
    Write-Host "Switching to REMOTE version mode..." -ForegroundColor Yellow
    Write-Host "  Using: github.com/lkxlzx/dnsproxy $Version" -ForegroundColor Gray
    
    # 替换为远程版本
    $content = $content -replace 'replace github\.com/AdguardTeam/dnsproxy => \./vendor-dev/dnsproxy', "replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy $Version"
}

# 写回 go.mod
$content | Set-Content $goModPath -NoNewline

Write-Host ""
Write-Host "Updating dependencies..." -ForegroundColor Yellow
go mod tidy

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ERROR: go mod tidy failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "SUCCESS! Switched to $Mode mode" -ForegroundColor Green
Write-Host ""
Write-Host "Current dnsproxy configuration:" -ForegroundColor Cyan
go list -m github.com/AdguardTeam/dnsproxy

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
if ($Mode -eq "local") {
    Write-Host "  1. Edit dnsproxy code in vendor-dev/dnsproxy" -ForegroundColor Gray
    Write-Host "  2. Rebuild: go build -o AdGuardHome_dev.exe" -ForegroundColor Gray
    Write-Host "  3. Test: go test ./..." -ForegroundColor Gray
} else {
    Write-Host "  1. Build: go build" -ForegroundColor Gray
    Write-Host "  2. Test: go test ./..." -ForegroundColor Gray
    Write-Host "  3. Commit changes" -ForegroundColor Gray
}
Write-Host ""
