# Build AdGuard Home for all platforms
# This script builds release binaries for all supported platforms

$ErrorActionPreference = "Stop"

Write-Host "Building AdGuard Home for all platforms..." -ForegroundColor Cyan
Write-Host ""

# Get version from git
try {
    $VERSION = git describe --tags --abbrev=4 HEAD 2>&1
    if ($LASTEXITCODE -ne 0) {
        # If no tags, use commit hash
        $VERSION = "v0.107.0-dev-$(git rev-parse --short HEAD)"
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

# Platform configurations: OS, ARCH, ARM, MIPS, Extension
$platforms = @(
    @{OS="windows"; ARCH="amd64"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_amd64"}
    @{OS="windows"; ARCH="386"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_386"}
    @{OS="windows"; ARCH="arm64"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_arm64"}
    @{OS="linux"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="linux_amd64"}
    @{OS="linux"; ARCH="386"; ARM=""; MIPS=""; EXT=""; NAME="linux_386"}
    @{OS="linux"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="linux_arm64"}
    @{OS="linux"; ARCH="arm"; ARM="5"; MIPS=""; EXT=""; NAME="linux_armv5"}
    @{OS="linux"; ARCH="arm"; ARM="6"; MIPS=""; EXT=""; NAME="linux_armv6"}
    @{OS="linux"; ARCH="arm"; ARM="7"; MIPS=""; EXT=""; NAME="linux_armv7"}
    @{OS="linux"; ARCH="mips"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mips_softfloat"}
    @{OS="linux"; ARCH="mipsle"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mipsle_softfloat"}
    @{OS="linux"; ARCH="mips64"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mips64_softfloat"}
    @{OS="linux"; ARCH="mips64le"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mips64le_softfloat"}
    @{OS="linux"; ARCH="ppc64le"; ARM=""; MIPS=""; EXT=""; NAME="linux_ppc64le"}
    @{OS="linux"; ARCH="riscv64"; ARM=""; MIPS=""; EXT=""; NAME="linux_riscv64"}
    @{OS="darwin"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="darwin_amd64"}
    @{OS="darwin"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="darwin_arm64"}
    @{OS="freebsd"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="freebsd_amd64"}
    @{OS="freebsd"; ARCH="386"; ARM=""; MIPS=""; EXT=""; NAME="freebsd_386"}
    @{OS="freebsd"; ARCH="arm"; ARM="5"; MIPS=""; EXT=""; NAME="freebsd_armv5"}
    @{OS="freebsd"; ARCH="arm"; ARM="6"; MIPS=""; EXT=""; NAME="freebsd_armv6"}
    @{OS="freebsd"; ARCH="arm"; ARM="7"; MIPS=""; EXT=""; NAME="freebsd_armv7"}
    @{OS="freebsd"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="freebsd_arm64"}
    @{OS="openbsd"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="openbsd_amd64"}
    @{OS="openbsd"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="openbsd_arm64"}
)

$total = $platforms.Count
$current = 0
$failed = @()

foreach ($platform in $platforms) {
    $current++
    $percent = [math]::Round(($current / $total) * 100)
    
    Write-Host "[$current/$total] Building $($platform.NAME)..." -ForegroundColor Yellow
    
    $env:GOOS = $platform.OS
    $env:GOARCH = $platform.ARCH
    $env:GOARM = $platform.ARM
    $env:GOMIPS = $platform.MIPS
    
    $outputFile = "dist\AdGuardHome_$($platform.NAME)$($platform.EXT)"
    
    try {
        $buildCmd = "go build -ldflags=`"$LDFLAGS`" -o `"$outputFile`""
        Invoke-Expression $buildCmd 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            $size = (Get-Item $outputFile).Length
            $sizeMB = [math]::Round($size / 1MB, 2)
            Write-Host "  ✓ Success ($sizeMB MB)" -ForegroundColor Green
        } else {
            throw "Build failed with exit code $LASTEXITCODE"
        }
    } catch {
        Write-Host "  ✗ Failed: $_" -ForegroundColor Red
        $failed += $platform.NAME
    }
}

# Clean up environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
Remove-Item Env:\GOMIPS -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Build Summary:" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Total platforms: $total" -ForegroundColor White
Write-Host "Successful: $($total - $failed.Count)" -ForegroundColor Green
Write-Host "Failed: $($failed.Count)" -ForegroundColor $(if ($failed.Count -gt 0) { "Red" } else { "Green" })

if ($failed.Count -gt 0) {
    Write-Host ""
    Write-Host "Failed platforms:" -ForegroundColor Red
    foreach ($f in $failed) {
        Write-Host "  - $f" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "All binaries are in the 'dist' directory." -ForegroundColor Cyan
Write-Host ""

# List built files
Write-Host "Built files:" -ForegroundColor Cyan
Get-ChildItem -Path "dist\AdGuardHome_*" | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) - $size MB" -ForegroundColor Gray
}
