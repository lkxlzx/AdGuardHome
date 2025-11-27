# Build AdGuard Home with DNS Routing Filter Fix

$ErrorActionPreference = "Stop"

Write-Host "Building AdGuard Home with DNS Routing Filter Fix..." -ForegroundColor Cyan
Write-Host ""

# Get version
try {
    $VERSION = git describe --tags --abbrev=4 HEAD 2>&1
    if ($LASTEXITCODE -ne 0) {
        $commitHash = git rev-parse --short HEAD
        $VERSION = "v0.107.0-dns-routing-fix-$commitHash"
    } else {
        $VERSION = "$VERSION-dns-routing-fix"
    }
} catch {
    $VERSION = "v0.107.0-dns-routing-fix"
}

Write-Host "Version: $VERSION" -ForegroundColor Green
Write-Host ""

# Build flags
$LDFLAGS = "-s -w -X github.com/AdguardTeam/AdGuardHome/internal/version.version=$VERSION"

# Build for Windows amd64
Write-Host "Building for Windows 64-bit..." -ForegroundColor Yellow

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

$outputFile = "AdGuardHome_dns_routing_fix.exe"

Write-Host "Compiling..." -NoNewline
$buildResult = go build -ldflags="$LDFLAGS" -o "$outputFile" 2>&1

if ($LASTEXITCODE -eq 0 -and (Test-Path $outputFile)) {
    $size = (Get-Item $outputFile).Length
    $sizeMB = [math]::Round($size / 1MB, 2)
    Write-Host " Done" -ForegroundColor Green
    Write-Host ""
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "  File: $outputFile" -ForegroundColor White
    Write-Host "  Size: $sizeMB MB" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host " Failed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Build failed!" -ForegroundColor Red
    if ($buildResult) {
        Write-Host $buildResult -ForegroundColor DarkRed
    }
    exit 1
}

# Clean up
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host "The fixed version is ready: .\$outputFile" -ForegroundColor Cyan
