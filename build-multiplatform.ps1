# Multi-platform build script for AdGuard Home v10.3
# Builds optimized binaries for major platforms with frontend included

$ErrorActionPreference = "Stop"

# Version info
$VERSION = "v10.3"
$BUILD_DATE = Get-Date -Format "yyyy-MM-dd"

# Create release directory
$RELEASE_DIR = "release"
if (Test-Path $RELEASE_DIR) {
    Remove-Item -Recurse -Force $RELEASE_DIR
}
New-Item -ItemType Directory -Path $RELEASE_DIR | Out-Null

Write-Host "=== Building AdGuard Home $VERSION ===" -ForegroundColor Cyan
Write-Host "Build Date: $BUILD_DATE" -ForegroundColor Cyan
Write-Host ""

# Step 1: Build frontend
Write-Host "[1/5] Building frontend..." -ForegroundColor Yellow
Set-Location client
npm run build-prod
if ($LASTEXITCODE -ne 0) {
    Write-Host "Frontend build failed!" -ForegroundColor Red
    exit 1
}
Set-Location ..
Write-Host "Frontend build completed!" -ForegroundColor Green
Write-Host ""

# Step 2: Define build targets
$targets = @(
    @{OS="windows"; ARCH="amd64"; EXT=".exe"; NAME="AdGuardHome_windows_amd64.exe"},
    @{OS="windows"; ARCH="386"; EXT=".exe"; NAME="AdGuardHome_windows_386.exe"},
    @{OS="linux"; ARCH="amd64"; EXT=""; NAME="AdGuardHome_linux_amd64"},
    @{OS="linux"; ARCH="386"; EXT=""; NAME="AdGuardHome_linux_386"},
    @{OS="linux"; ARCH="arm64"; EXT=""; NAME="AdGuardHome_linux_arm64"},
    @{OS="linux"; ARCH="arm"; EXT=""; NAME="AdGuardHome_linux_armv7"},
    @{OS="darwin"; ARCH="amd64"; EXT=""; NAME="AdGuardHome_darwin_amd64"},
    @{OS="darwin"; ARCH="arm64"; EXT=""; NAME="AdGuardHome_darwin_arm64"},
    @{OS="freebsd"; ARCH="amd64"; EXT=""; NAME="AdGuardHome_freebsd_amd64"}
)

# Build flags for optimization
$LDFLAGS = "-s -w -X main.version=$VERSION -X main.buildtime=$BUILD_DATE"

Write-Host "[2/5] Building binaries for multiple platforms..." -ForegroundColor Yellow
$buildCount = 0
$totalBuilds = $targets.Count

foreach ($target in $targets) {
    $buildCount++
    $percent = [math]::Round(($buildCount / $totalBuilds) * 100)
    
    Write-Host "[$buildCount/$totalBuilds] Building $($target.NAME) ($percent%)..." -ForegroundColor Cyan
    
    $env:GOOS = $target.OS
    $env:GOARCH = $target.ARCH
    
    # Special handling for ARM v7
    if ($target.ARCH -eq "arm") {
        $env:GOARM = "7"
    }
    
    $outputPath = Join-Path $RELEASE_DIR $target.NAME
    
    try {
        go build -ldflags $LDFLAGS -o $outputPath .
        
        if (Test-Path $outputPath) {
            $size = (Get-Item $outputPath).Length
            $sizeMB = [math]::Round($size / 1MB, 2)
            Write-Host "  ✓ Built successfully: $sizeMB MB" -ForegroundColor Green
        } else {
            Write-Host "  ✗ Build failed!" -ForegroundColor Red
        }
    } catch {
        Write-Host "  ✗ Error: $_" -ForegroundColor Red
    }
    
    # Clean up ARM env var
    if ($env:GOARM) {
        Remove-Item Env:\GOARM
    }
}

# Reset environment
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "[3/5] Compressing binaries with UPX..." -ForegroundColor Yellow

# Check if UPX is available
$upxAvailable = $false
try {
    $upxVersion = upx --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        $upxAvailable = $true
        Write-Host "UPX found, compressing binaries..." -ForegroundColor Green
    }
} catch {
    Write-Host "UPX not found, skipping compression (binaries will be larger)" -ForegroundColor Yellow
}

if ($upxAvailable) {
    Get-ChildItem $RELEASE_DIR -File | ForEach-Object {
        Write-Host "  Compressing $($_.Name)..." -ForegroundColor Cyan
        $originalSize = $_.Length
        
        # UPX compression (skip darwin/macOS as it may cause issues with code signing)
        if ($_.Name -notlike "*darwin*") {
            upx --best --lzma $_.FullName 2>&1 | Out-Null
            
            if (Test-Path $_.FullName) {
                $newSize = (Get-Item $_.FullName).Length
                $reduction = [math]::Round((1 - ($newSize / $originalSize)) * 100, 1)
                Write-Host "    Reduced by $reduction%" -ForegroundColor Green
            }
        } else {
            Write-Host "    Skipped (macOS binary)" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "[4/5] Generating checksums..." -ForegroundColor Yellow

$checksumFile = Join-Path $RELEASE_DIR "checksums.txt"
$checksums = @()

Get-ChildItem $RELEASE_DIR -File | Where-Object { $_.Name -ne "checksums.txt" } | ForEach-Object {
    $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash
    $checksums += "$hash  $($_.Name)"
    Write-Host "  $($_.Name): $hash" -ForegroundColor Gray
}

$checksums | Out-File -FilePath $checksumFile -Encoding UTF8

Write-Host ""
Write-Host "[5/5] Build summary:" -ForegroundColor Yellow
Write-Host "===========================================" -ForegroundColor Cyan

$totalSize = 0
Get-ChildItem $RELEASE_DIR -File | Where-Object { $_.Name -ne "checksums.txt" } | ForEach-Object {
    $sizeMB = [math]::Round($_.Length / 1MB, 2)
    $totalSize += $_.Length
    Write-Host "  $($_.Name.PadRight(35)) $sizeMB MB" -ForegroundColor White
}

$totalSizeMB = [math]::Round($totalSize / 1MB, 2)
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "  Total size: $totalSizeMB MB" -ForegroundColor Green
Write-Host "  Files: $($targets.Count) binaries + 1 checksum file" -ForegroundColor Green
Write-Host ""
Write-Host "✓ Build completed successfully!" -ForegroundColor Green
Write-Host "  Output directory: $RELEASE_DIR" -ForegroundColor Cyan
