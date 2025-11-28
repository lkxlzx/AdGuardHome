# Build AdGuardHome for main platforms
# Version: V5 with hot domains fix and prefetch UI improvements

$ErrorActionPreference = "Stop"

Write-Host "=== Building AdGuardHome for Main Platforms ===" -ForegroundColor Cyan
Write-Host ""

# Create output directory
$outputDir = "dist_v5_main"
if (Test-Path $outputDir) {
    Remove-Item -Recurse -Force $outputDir
}
New-Item -ItemType Directory -Path $outputDir | Out-Null

# Get version info
$version = "v5.0.0"
$buildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

Write-Host "Version: $version" -ForegroundColor Green
Write-Host "Build Time: $buildTime" -ForegroundColor Green
Write-Host ""

# Platform configurations
$platforms = @(
    @{OS="windows"; ARCH="amd64"; EXT=".exe"; NAME="Windows x64"},
    @{OS="windows"; ARCH="386"; EXT=".exe"; NAME="Windows x86"},
    @{OS="linux"; ARCH="amd64"; EXT=""; NAME="Linux x64"},
    @{OS="linux"; ARCH="arm64"; EXT=""; NAME="Linux ARM64"},
    @{OS="darwin"; ARCH="amd64"; EXT=""; NAME="macOS x64"},
    @{OS="darwin"; ARCH="arm64"; EXT=""; NAME="macOS ARM64"}
)

$successCount = 0
$failCount = 0
$buildResults = @()

foreach ($platform in $platforms) {
    $os = $platform.OS
    $arch = $platform.ARCH
    $ext = $platform.EXT
    $name = $platform.NAME
    
    $outputFile = "AdGuardHome_${os}_${arch}${ext}"
    $outputPath = Join-Path $outputDir $outputFile
    
    Write-Host "Building for $name ($os/$arch)..." -ForegroundColor Yellow
    
    try {
        $env:GOOS = $os
        $env:GOARCH = $arch
        $env:CGO_ENABLED = "0"
        
        $startTime = Get-Date
        go build -ldflags "-s -w" -o $outputPath 2>&1 | Out-Null
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalSeconds
        
        if (Test-Path $outputPath) {
            $fileSize = (Get-Item $outputPath).Length
            $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
            
            Write-Host "  Success! Size: $fileSizeMB MB, Time: $([math]::Round($duration, 1))s" -ForegroundColor Green
            
            $buildResults += [PSCustomObject]@{
                Platform = $name
                OS = $os
                Arch = $arch
                File = $outputFile
                Size = "$fileSizeMB MB"
                Time = "$([math]::Round($duration, 1))s"
                Status = "Success"
            }
            
            $successCount++
        } else {
            throw "Output file not created"
        }
    }
    catch {
        Write-Host "  Failed: $_" -ForegroundColor Red
        
        $buildResults += [PSCustomObject]@{
            Platform = $name
            OS = $os
            Arch = $arch
            File = $outputFile
            Size = "N/A"
            Time = "N/A"
            Status = "Failed"
        }
        
        $failCount++
    }
    
    Write-Host ""
}

# Reset environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

# Summary
Write-Host "=== Build Summary ===" -ForegroundColor Cyan
Write-Host ""
$buildResults | Format-Table -AutoSize
Write-Host ""

Write-Host "Total: $($platforms.Count) platforms" -ForegroundColor White
Write-Host "Success: $successCount" -ForegroundColor Green
Write-Host "Failed: $failCount" -ForegroundColor $(if ($failCount -gt 0) { "Red" } else { "Green" })
Write-Host ""

# Create README
$readmeContent = @"
# AdGuardHome V5 Main Platforms Build

## Version Information
- Version: $version
- Build Time: $buildTime
- Branch: v5

## Features
- Hot domains tracking fix (uses tracked_domains instead of hot_domains)
- Prefetch UI improvements (always show last prefetch time)
- All V4 features and optimizations

## Build Results

| Platform | OS | Architecture | File | Size | Status |
|----------|----|--------------| -----|------|--------|
"@

foreach ($result in $buildResults) {
    $readmeContent += "`n| $($result.Platform) | $($result.OS) | $($result.Arch) | $($result.File) | $($result.Size) | $($result.Status) |"
}

$readmeContent += @"


## Installation

### Windows
1. Download ``AdGuardHome_windows_amd64.exe`` (64-bit) or ``AdGuardHome_windows_386.exe`` (32-bit)
2. Rename to ``AdGuardHome.exe``
3. Run ``AdGuardHome.exe -s install`` to install as service
4. Access web interface at http://localhost:3000

### Linux
1. Download ``AdGuardHome_linux_amd64`` (x64) or ``AdGuardHome_linux_arm64`` (ARM64)
2. Make executable: ``chmod +x AdGuardHome_linux_*``
3. Run: ``./AdGuardHome_linux_*``
4. Access web interface at http://localhost:3000

### macOS
1. Download ``AdGuardHome_darwin_amd64`` (Intel) or ``AdGuardHome_darwin_arm64`` (Apple Silicon)
2. Make executable: ``chmod +x AdGuardHome_darwin_*``
3. Run: ``./AdGuardHome_darwin_*``
4. Access web interface at http://localhost:3000

## Changes in This Build

### Hot Domains Fix
- Fixed hot domains count display issue
- Now uses ``tracked_domains`` metric for stable count
- Count only decreases during cleanup (not during prefetch scheduling)

### Prefetch UI Improvements
- "Last Prefetch" field now always visible
- Shows "--" when no prefetch has occurred yet
- Better user experience and consistency

## Notes
- All binaries are statically compiled (CGO_ENABLED=0)
- Binaries are stripped for smaller size (-ldflags "-s -w")
- Based on V5 branch with all V4 improvements

## Support
For issues or questions, please visit:
https://github.com/lkxlzx/AdGuardHome

---
Built on: $buildTime
"@

$readmeContent | Out-File -FilePath (Join-Path $outputDir "README.md") -Encoding UTF8

Write-Host "Build complete! Files saved to: $outputDir" -ForegroundColor Green
Write-Host ""

# Open output directory
if ($successCount -gt 0) {
    Write-Host "Opening output directory..." -ForegroundColor Yellow
    Start-Process explorer.exe $outputDir
}
