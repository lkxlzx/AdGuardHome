# AdGuardHome v10.3 Release Build Script
# Builds optimized binaries for multiple platforms

$ErrorActionPreference = "Stop"

Write-Host "=== AdGuardHome v10.3 Release Build ===" -ForegroundColor Cyan
Write-Host ""

# Version info
$VERSION = "v10.3-complete"
$BUILD_TIME = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$OUTPUT_DIR = "release"

# Create output directory
if (!(Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

Write-Host "Version: $VERSION" -ForegroundColor Green
Write-Host "Build Time: $BUILD_TIME" -ForegroundColor Green
Write-Host "Output Directory: $OUTPUT_DIR" -ForegroundColor Green
Write-Host ""

# Build flags for optimization
# -s: strip symbol table
# -w: strip DWARF debug info
$LDFLAGS = "-s -w"
$BUILD_FLAGS = "-trimpath"

# Platform configurations
$PLATFORMS = @(
    @{OS="windows"; ARCH="amd64"; EXT=".exe"; NAME="Windows x64"},
    @{OS="windows"; ARCH="386"; EXT=".exe"; NAME="Windows x86"},
    @{OS="linux"; ARCH="amd64"; EXT=""; NAME="Linux x64"},
    @{OS="linux"; ARCH="386"; EXT=""; NAME="Linux x86"},
    @{OS="linux"; ARCH="arm64"; EXT=""; NAME="Linux ARM64"},
    @{OS="linux"; ARCH="arm"; EXT=""; NAME="Linux ARM"},
    @{OS="darwin"; ARCH="amd64"; EXT=""; NAME="macOS x64"},
    @{OS="darwin"; ARCH="arm64"; EXT=""; NAME="macOS ARM64"}
)

$SUCCESS_COUNT = 0
$FAILED_COUNT = 0
$BUILD_RESULTS = @()

foreach ($PLATFORM in $PLATFORMS) {
    $OS = $PLATFORM.OS
    $ARCH = $PLATFORM.ARCH
    $EXT = $PLATFORM.EXT
    $NAME = $PLATFORM.NAME
    
    $OUTPUT_NAME = "AdGuardHome_${VERSION}_${OS}_${ARCH}${EXT}"
    $OUTPUT_PATH = Join-Path $OUTPUT_DIR $OUTPUT_NAME
    
    Write-Host "Building for $NAME ($OS/$ARCH)..." -ForegroundColor Yellow
    
    $env:GOOS = $OS
    $env:GOARCH = $ARCH
    $env:CGO_ENABLED = "0"
    
    try {
        $startTime = Get-Date
        
        # Build command
        $buildCmd = "go build $BUILD_FLAGS -ldflags `"$LDFLAGS`" -o `"$OUTPUT_PATH`""
        $output = Invoke-Expression $buildCmd 2>&1
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  Build output:" -ForegroundColor Gray
            $output | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        }
        
        if ($LASTEXITCODE -eq 0) {
            $endTime = Get-Date
            $duration = ($endTime - $startTime).TotalSeconds
            $fileSize = (Get-Item $OUTPUT_PATH).Length
            $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
            
            Write-Host "  ✓ Success - Size: ${fileSizeMB} MB - Time: ${duration}s" -ForegroundColor Green
            
            $SUCCESS_COUNT++
            $BUILD_RESULTS += @{
                Platform = $NAME
                OS = $OS
                Arch = $ARCH
                Status = "Success"
                Size = "${fileSizeMB} MB"
                Time = "${duration}s"
                File = $OUTPUT_NAME
            }
        } else {
            throw "Build failed with exit code $LASTEXITCODE"
        }
    } catch {
        Write-Host "  ✗ Failed: $_" -ForegroundColor Red
        $FAILED_COUNT++
        $BUILD_RESULTS += @{
            Platform = $NAME
            OS = $OS
            Arch = $ARCH
            Status = "Failed"
            Size = "N/A"
            Time = "N/A"
            File = "N/A"
        }
    }
}

# Reset environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "=== Build Summary ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Total Platforms: $($PLATFORMS.Count)" -ForegroundColor White
Write-Host "Successful: $SUCCESS_COUNT" -ForegroundColor Green
Write-Host "Failed: $FAILED_COUNT" -ForegroundColor $(if ($FAILED_COUNT -gt 0) { "Red" } else { "Green" })
Write-Host ""

# Display results table
Write-Host "=== Build Results ===" -ForegroundColor Cyan
Write-Host ""
Write-Host ("{0,-20} {1,-10} {1,-10} {2,-12} {3,-10} {4,-10}" -f "Platform", "OS", "Arch", "Status", "Size", "Time") -ForegroundColor White
Write-Host ("-" * 80) -ForegroundColor Gray

foreach ($result in $BUILD_RESULTS) {
    $color = if ($result.Status -eq "Success") { "Green" } else { "Red" }
    Write-Host ("{0,-20} {1,-10} {2,-10} {3,-12} {4,-10} {5,-10}" -f `
        $result.Platform, $result.OS, $result.Arch, $result.Status, $result.Size, $result.Time) -ForegroundColor $color
}

Write-Host ""

# Calculate total size
if ($SUCCESS_COUNT -gt 0) {
    $totalSize = 0
    Get-ChildItem $OUTPUT_DIR -File | ForEach-Object {
        $totalSize += $_.Length
    }
    $totalSizeMB = [math]::Round($totalSize / 1MB, 2)
    Write-Host "Total Size: ${totalSizeMB} MB" -ForegroundColor Cyan
    Write-Host ""
}

# List output files
Write-Host "=== Output Files ===" -ForegroundColor Cyan
Write-Host ""
Get-ChildItem $OUTPUT_DIR -File | ForEach-Object {
    $sizeMB = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) - ${sizeMB} MB" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=== Build Complete ===" -ForegroundColor Cyan

if ($FAILED_COUNT -eq 0) {
    Write-Host "All builds successful!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Some builds failed. Check the output above." -ForegroundColor Yellow
    exit 1
}
