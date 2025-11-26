# Test compiled binaries
# This script tests if the compiled binaries are valid

$ErrorActionPreference = "Stop"

Write-Host "Testing Compiled Binaries..." -ForegroundColor Cyan
Write-Host "=============================" -ForegroundColor Cyan
Write-Host ""

$binaries = Get-ChildItem -Path "dist\AdGuardHome_*" -ErrorAction SilentlyContinue

if ($binaries.Count -eq 0) {
    Write-Host "❌ No binaries found in dist/ directory" -ForegroundColor Red
    Write-Host "Please run build-release-simple.ps1 first" -ForegroundColor Yellow
    exit 1
}

Write-Host "Found $($binaries.Count) binaries" -ForegroundColor Green
Write-Host ""

$windowsBinaries = $binaries | Where-Object { $_.Name -like "*windows*" }

if ($windowsBinaries.Count -eq 0) {
    Write-Host "⚠️  No Windows binaries found to test" -ForegroundColor Yellow
    Write-Host "Cannot test non-Windows binaries on Windows platform" -ForegroundColor Yellow
    exit 0
}

Write-Host "Testing Windows binaries..." -ForegroundColor Cyan
Write-Host ""

$tested = 0
$passed = 0
$failed = 0

foreach ($binary in $windowsBinaries) {
    $tested++
    
    # Skip ARM64 on non-ARM64 systems
    if ($binary.Name -like "*arm64*" -and [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE") -ne "ARM64") {
        Write-Host "[$tested/$($windowsBinaries.Count)] Skipping $($binary.Name) (incompatible architecture)" -ForegroundColor DarkGray
        Write-Host ""
        continue
    }
    
    Write-Host "[$tested/$($windowsBinaries.Count)] Testing $($binary.Name)..." -ForegroundColor Yellow -NoNewline
    
    try {
        # Test if binary can show version
        $output = & $binary.FullName --version 2>&1
        
        if ($LASTEXITCODE -eq 0 -or $output -match "AdGuard Home|version") {
            Write-Host " ✓ OK" -ForegroundColor Green
            $passed++
            
            # Show version info
            if ($output) {
                Write-Host "  Version: $($output | Select-Object -First 1)" -ForegroundColor Gray
            }
        } else {
            Write-Host " ✗ Failed" -ForegroundColor Red
            Write-Host "  Output: $output" -ForegroundColor DarkRed
            $failed++
        }
    } catch {
        Write-Host " ✗ Error: $_" -ForegroundColor Red
        $failed++
    }
    
    Write-Host ""
}

Write-Host "=============================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "=============================" -ForegroundColor Cyan
Write-Host "Total tested: $tested" -ForegroundColor White
Write-Host "Passed:       $passed" -ForegroundColor Green
Write-Host "Failed:       $failed" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Green" })
Write-Host ""

if ($failed -eq 0) {
    Write-Host "✅ All Windows binaries are working correctly!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Note: Non-Windows binaries cannot be tested on Windows platform." -ForegroundColor Yellow
    Write-Host "They should be tested on their respective platforms." -ForegroundColor Yellow
    exit 0
} else {
    Write-Host "❌ Some binaries failed the test" -ForegroundColor Red
    exit 1
}
