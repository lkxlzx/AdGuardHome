# Build AdGuard Home Filter Priority Fix for all major platforms

Write-Host "=== Building Filter Priority Fix for All Platforms ===" -ForegroundColor Cyan
Write-Host ""

# Create output directory
$distDir = "dist_filter_priority_fix"
if (Test-Path $distDir) {
    Remove-Item -Recurse -Force $distDir
}
New-Item -ItemType Directory -Path $distDir | Out-Null

# Get version
try {
    $version = git describe --tags --abbrev=4 HEAD 2>$null
    if (-not $version) {
        $version = "v0.0.0-filter-fix"
    }
} catch {
    $version = "v0.0.0-filter-fix"
}

Write-Host "Version: $version" -ForegroundColor Green
Write-Host ""

# Build flags
$ldflags = "-s -w"

# Define platforms
$platforms = @(
    @{OS="windows"; Arch="amd64"; Ext=".exe"; Name="Windows AMD64"},
    @{OS="windows"; Arch="arm64"; Ext=".exe"; Name="Windows ARM64"},
    @{OS="windows"; Arch="386"; Ext=".exe"; Name="Windows 386"},
    
    @{OS="linux"; Arch="amd64"; Ext=""; Name="Linux AMD64"},
    @{OS="linux"; Arch="arm64"; Ext=""; Name="Linux ARM64"},
    @{OS="linux"; Arch="arm"; Ext=""; Name="Linux ARMv7"; Arm="7"},
    @{OS="linux"; Arch="386"; Ext=""; Name="Linux 386"},
    
    @{OS="darwin"; Arch="amd64"; Ext=""; Name="macOS AMD64"},
    @{OS="darwin"; Arch="arm64"; Ext=""; Name="macOS ARM64"}
)

$successCount = 0
$failCount = 0

foreach ($platform in $platforms) {
    $os = $platform.OS
    $arch = $platform.Arch
    $ext = $platform.Ext
    $name = $platform.Name
    $arm = $platform.Arm
    
    $outputName = "AdGuardHome_${os}_${arch}${ext}"
    $outputPath = Join-Path $distDir $outputName
    
    Write-Host "Building $name..." -ForegroundColor Cyan
    
    # Set environment variables
    $env:GOOS = $os
    $env:GOARCH = $arch
    if ($arm) {
        $env:GOARM = $arm
    } else {
        Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
    }
    
    # Build
    $result = & go build -ldflags="$ldflags" -o "$outputPath" 2>&1
    
    if ($LASTEXITCODE -eq 0 -and (Test-Path $outputPath)) {
        $fileSize = (Get-Item $outputPath).Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "  Success ($fileSizeMB MB)" -ForegroundColor Green
        $successCount++
    } else {
        Write-Host "  Failed" -ForegroundColor Red
        $failCount++
    }
    
    Write-Host ""
}

# Clean up environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\GOARM -ErrorAction SilentlyContinue

# Summary
Write-Host "=== Build Summary ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Total: $($platforms.Count) platforms" -ForegroundColor Cyan
Write-Host "Success: $successCount" -ForegroundColor Green
Write-Host "Failed: $failCount" -ForegroundColor $(if ($failCount -gt 0) { "Red" } else { "Green" })
Write-Host ""

if ($successCount -gt 0) {
    $totalSize = (Get-ChildItem $distDir -File | Measure-Object -Property Length -Sum).Sum
    $totalSizeMB = [math]::Round($totalSize / 1MB, 2)
    Write-Host "Total Size: $totalSizeMB MB" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "Built files:" -ForegroundColor Cyan
    Get-ChildItem $distDir -File | Select-Object Name, @{Label="Size(MB)";Expression={[math]::Round($_.Length/1MB,2)}} | Format-Table -AutoSize
}

Write-Host ""
Write-Host "Output directory: $distDir" -ForegroundColor Green
Write-Host ""

if ($failCount -eq 0) {
    Write-Host "All platforms built successfully!" -ForegroundColor Green
} else {
    Write-Host "Warning: $failCount platform(s) failed to build" -ForegroundColor Red
}
