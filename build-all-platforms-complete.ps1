# Build AdGuard Home for ALL supported platforms
# Complete version with all platforms from official build script

$ErrorActionPreference = "Stop"

Write-Host "Building AdGuard Home for ALL Platforms..." -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
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

# All platform configurations from official build-release.sh
$platforms = @(
    # Darwin (macOS)
    @{OS="darwin"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="darwin_amd64"}
    @{OS="darwin"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="darwin_arm64"}
    
    # FreeBSD
    @{OS="freebsd"; ARCH="386"; ARM=""; MIPS=""; EXT=""; NAME="freebsd_386"}
    @{OS="freebsd"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="freebsd_amd64"}
    @{OS="freebsd"; ARCH="arm"; ARM="5"; MIPS=""; EXT=""; NAME="freebsd_armv5"}
    @{OS="freebsd"; ARCH="arm"; ARM="6"; MIPS=""; EXT=""; NAME="freebsd_armv6"}
    @{OS="freebsd"; ARCH="arm"; ARM="7"; MIPS=""; EXT=""; NAME="freebsd_armv7"}
    @{OS="freebsd"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="freebsd_arm64"}
    
    # Linux
    @{OS="linux"; ARCH="386"; ARM=""; MIPS=""; EXT=""; NAME="linux_386"}
    @{OS="linux"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="linux_amd64"}
    @{OS="linux"; ARCH="arm"; ARM="5"; MIPS=""; EXT=""; NAME="linux_armv5"}
    @{OS="linux"; ARCH="arm"; ARM="6"; MIPS=""; EXT=""; NAME="linux_armv6"}
    @{OS="linux"; ARCH="arm"; ARM="7"; MIPS=""; EXT=""; NAME="linux_armv7"}
    @{OS="linux"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="linux_arm64"}
    @{OS="linux"; ARCH="mips"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mips_softfloat"}
    @{OS="linux"; ARCH="mips64"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mips64_softfloat"}
    @{OS="linux"; ARCH="mips64le"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mips64le_softfloat"}
    @{OS="linux"; ARCH="mipsle"; ARM=""; MIPS="softfloat"; EXT=""; NAME="linux_mipsle_softfloat"}
    @{OS="linux"; ARCH="ppc64le"; ARM=""; MIPS=""; EXT=""; NAME="linux_ppc64le"}
    @{OS="linux"; ARCH="riscv64"; ARM=""; MIPS=""; EXT=""; NAME="linux_riscv64"}
    
    # OpenBSD
    @{OS="openbsd"; ARCH="amd64"; ARM=""; MIPS=""; EXT=""; NAME="openbsd_amd64"}
    @{OS="openbsd"; ARCH="arm64"; ARM=""; MIPS=""; EXT=""; NAME="openbsd_arm64"}
    
    # Windows
    @{OS="windows"; ARCH="386"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_386"}
    @{OS="windows"; ARCH="amd64"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_amd64"}
    @{OS="windows"; ARCH="arm64"; ARM=""; MIPS=""; EXT=".exe"; NAME="windows_arm64"}
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
    
    Write-Host "[$current/$total] $($platform.NAME)..." -ForegroundColor Yellow -NoNewline
    
    $env:GOOS = $platform.OS
    $env:GOARCH = $platform.ARCH
    if ($platform.ARM) { $env:GOARM = $platform.ARM } else { Remove-Item Env:\GOARM -ErrorAction SilentlyContinue }
    if ($platform.MIPS) { $env:GOMIPS = $platform.MIPS } else { Remove-Item Env:\GOMIPS -ErrorAction SilentlyContinue }
    
    $outputFile = "dist\AdGuardHome_$($platform.NAME)$($platform.EXT)"
    
    try {
        $output = & go build -ldflags="$LDFLAGS" -o "$outputFile" 2>&1
        
        if ($LASTEXITCODE -eq 0 -and (Test-Path $outputFile)) {
            $size = (Get-Item $outputFile).Length
            $sizeMB = [math]::Round($size / 1MB, 2)
            Write-Host " ✓ ($sizeMB MB)" -ForegroundColor Green
            $successful += $platform.NAME
        } else {
            Write-Host " ✗" -ForegroundColor Red
            $failed += $platform.NAME
        }
    } catch {
        Write-Host " ✗" -ForegroundColor Red
        $failed += $platform.NAME
    }
}

# Clean up environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
Remove-Item Env:\GOMIPS -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "Build Summary" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "Version:    $VERSION" -ForegroundColor White
Write-Host "Total:      $total platforms" -ForegroundColor White
Write-Host "Successful: $($successful.Count)" -ForegroundColor Green
Write-Host "Failed:     $($failed.Count)" -ForegroundColor $(if ($failed.Count -gt 0) { "Red" } else { "Green" })

if ($failed.Count -gt 0) {
    Write-Host ""
    Write-Host "Failed platforms:" -ForegroundColor Yellow
    foreach ($f in $failed) {
        Write-Host "  ✗ $f" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Built binaries in 'dist' directory:" -ForegroundColor Cyan
Get-ChildItem -Path "dist\AdGuardHome_*" -ErrorAction SilentlyContinue | Sort-Object Name | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  ✓ $($_.Name) - $size MB" -ForegroundColor Gray
}

Write-Host ""
if ($failed.Count -eq 0) {
    Write-Host "🎉 All $total platforms built successfully!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "✓ $($successful.Count)/$total platforms built successfully." -ForegroundColor Green
    exit 0
}
