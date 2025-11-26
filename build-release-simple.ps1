# Build AdGuard Home for major platforms
# Simplified version focusing on most common platforms

$ErrorActionPreference = "Stop"

Write-Host "Building AdGuard Home Release Binaries..." -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Get version
try {
    $VERSION = git describe --tags --abbrev=4 HEAD 2>&1
    if ($LASTEXITCODE -ne 0) {
        $commitHash = git rev-parse --short HEAD
        $VERSION = "v0.107.0-dev-$commitHash"
    }
} catch {
    $VERSION = "v0.107.0-dev"
}

Write-Host "Version: $VERSION" -ForegroundColor Green
Write-Host ""

# Create dist directory
if (-not (Test-Path "dist")) {
    New-Item -ItemType Directory -Path "dist" | Out-Null
}

# Build flags
$LDFLAGS = "-s -w -X github.com/AdguardTeam/AdGuardHome/internal/version.version=$VERSION"

# Platform configurations for major platforms
$platforms = @(
    # Windows
    @{OS="windows"; ARCH="amd64"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_amd64"; DESC="Windows 64-bit"}
    @{OS="windows"; ARCH="386"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_386"; DESC="Windows 32-bit"}
    @{OS="windows"; ARCH="arm64"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_arm64"; DESC="Windows ARM64"}
    
    # Linux
    @{OS="linux"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="linux_amd64"; DESC="Linux 64-bit"}
    @{OS="linux"; ARCH="386"; ARM=""; MIPS=""; EXT=""; NAME="linux_386"; DESC="Linux 32-bit"}
    @{OS="linux"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="linux_arm64"; DESC="Linux ARM64"}
    @{OS="linux"; ARCH="arm"; ARM="7"; MIPS=""; EXT=""; NAME="linux_armv7"; DESC="Linux ARMv7"}
    
    # macOS
    @{OS="darwin"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="darwin_amd64"; DESC="macOS Intel"}
    @{OS="darwin"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="darwin_arm64"; DESC="macOS Apple Silicon"}
)

$total = $platforms.Count
$current = 0
$successful = @()
$failed = @()

Write-Host "Building $total platform binaries..." -ForegroundColor Cyan
Write-Host ""

foreach ($platform in $platforms) {
    $current++
    $percent = [math]::Round(($current / $total) * 100)
    
    Write-Host "[$current/$total] $($platform.DESC)..." -ForegroundColor Yellow -NoNewline
    
    $env:GOOS = $platform.OS
    $env:GOARCH = $platform.ARCH
    if ($platform.ARM) { $env:GOARM = $platform.ARM } else { Remove-Item Env:\GOARM -ErrorAction SilentlyContinue }
    if ($platform.MIPS) { $env:GOMIPS = $platform.MIPS } else { Remove-Item Env:\GOMIPS -ErrorAction SilentlyContinue }
    
    $outputFile = "dist\AdGuardHome_$($platform.NAME)$($platform.EXT)"
    
    try {
        # Redirect stderr to stdout and capture
        $output = & go build -ldflags="$LDFLAGS" -o "$outputFile" 2>&1
        
        if ($LASTEXITCODE -eq 0 -and (Test-Path $outputFile)) {
            $size = (Get-Item $outputFile).Length
            $sizeMB = [math]::Round($size / 1MB, 2)
            Write-Host " ✓ ($sizeMB MB)" -ForegroundColor Green
            $successful += $platform.NAME
        } else {
            Write-Host " ✗ Failed" -ForegroundColor Red
            if ($output) {
                Write-Host "  Error: $($output | Select-Object -Last 3)" -ForegroundColor DarkRed
            }
            $failed += $platform.NAME
        }
    } catch {
        Write-Host " ✗ Failed: $_" -ForegroundColor Red
        $failed += $platform.NAME
    }
}

# Clean up environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
Remove-Item Env:\GOMIPS -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Build Summary" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Version:    $VERSION" -ForegroundColor White
Write-Host "Total:      $total platforms" -ForegroundColor White
Write-Host "Successful: $($successful.Count)" -ForegroundColor Green
Write-Host "Failed:     $($failed.Count)" -ForegroundColor $(if ($failed.Count -gt 0) { "Red" } else { "Green" })

if ($failed.Count -gt 0) {
    Write-Host ""
    Write-Host "Failed platforms:" -ForegroundColor Red
    foreach ($f in $failed) {
        Write-Host "  ✗ $f" -ForegroundColor Red
    }
}

if ($successful.Count -gt 0) {
    Write-Host ""
    Write-Host "Successfully built binaries:" -ForegroundColor Green
    Get-ChildItem -Path "dist\AdGuardHome_*" -ErrorAction SilentlyContinue | Sort-Object Name | ForEach-Object {
        $size = [math]::Round($_.Length / 1MB, 2)
        Write-Host "  ✓ $($_.Name) - $size MB" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "All binaries are in the 'dist' directory." -ForegroundColor Cyan

if ($failed.Count -eq 0) {
    Write-Host ""
    Write-Host "🎉 All platforms built successfully!" -ForegroundColor Green
    exit 0
} else {
    Write-Host ""
    Write-Host "⚠️  Some platforms failed to build." -ForegroundColor Yellow
    exit 1
}
