# Verify DNS Routing Cache Configuration
# This script checks if the configuration parameters are correctly defined

Write-Host "=== Configuration Verification ===" -ForegroundColor Cyan
Write-Host ""

# Check if AdGuardHome.yaml exists
if (Test-Path "AdGuardHome.yaml") {
    Write-Host "SUCCESS: AdGuardHome.yaml found" -ForegroundColor Green
    
    # Read configuration file
    $config = Get-Content "AdGuardHome.yaml" -Raw
    
    # Check for routing cache configuration
    Write-Host ""
    Write-Host "Checking for routing cache configuration..." -ForegroundColor Yellow
    
    if ($config -match "routing_cache_enabled") {
        Write-Host "  SUCCESS: routing_cache_enabled found" -ForegroundColor Green
        
        # Extract value
        if ($config -match "routing_cache_enabled:\s*(\w+)") {
            $enabled = $matches[1]
            Write-Host "    Value: $enabled" -ForegroundColor Cyan
        }
    } else {
        Write-Host "  INFO: routing_cache_enabled not found (will use default: true)" -ForegroundColor Yellow
    }
    
    if ($config -match "routing_cache_size") {
        Write-Host "  SUCCESS: routing_cache_size found" -ForegroundColor Green
        
        # Extract value
        if ($config -match "routing_cache_size:\s*(\d+)") {
            $size = $matches[1]
            Write-Host "    Value: $size" -ForegroundColor Cyan
        }
    } else {
        Write-Host "  INFO: routing_cache_size not found (will use default: 10000)" -ForegroundColor Yellow
    }
    
    if ($config -match "routing_cache_ttl") {
        Write-Host "  SUCCESS: routing_cache_ttl found" -ForegroundColor Green
        
        # Extract value
        if ($config -match "routing_cache_ttl:\s*(\d+)") {
            $ttl = $matches[1]
            Write-Host "    Value: $ttl minutes" -ForegroundColor Cyan
        }
    } else {
        Write-Host "  INFO: routing_cache_ttl not found (will use default: 5 minutes)" -ForegroundColor Yellow
    }
    
} else {
    Write-Host "WARNING: AdGuardHome.yaml not found" -ForegroundColor Yellow
}

Write-Host ""

# Check if example configuration exists
if (Test-Path "AdGuardHome_cache_example.yaml") {
    Write-Host "SUCCESS: AdGuardHome_cache_example.yaml found" -ForegroundColor Green
    Write-Host "  You can copy configuration from this example file" -ForegroundColor Cyan
} else {
    Write-Host "WARNING: AdGuardHome_cache_example.yaml not found" -ForegroundColor Yellow
}

Write-Host ""

# Check if executable exists
$exeFiles = @(
    "AdGuardHome_v10.3_CACHE_CONFIG.exe",
    "AdGuardHome_v10.3_FINAL_COMPLETE.exe",
    "AdGuardHome_test.exe"
)

Write-Host "Checking for executable files..." -ForegroundColor Yellow
$foundExe = $false
foreach ($exe in $exeFiles) {
    if (Test-Path $exe) {
        Write-Host "  SUCCESS: $exe found" -ForegroundColor Green
        $foundExe = $true
    }
}

if (-not $foundExe) {
    Write-Host "  WARNING: No executable files found" -ForegroundColor Yellow
}

Write-Host ""

# Check source code files
Write-Host "Checking source code modifications..." -ForegroundColor Yellow

$sourceFiles = @{
    "internal/home/config.go" = "routing_cache"
    "internal/home/dns.go" = "RoutingCache"
    "internal/dnsforward/http.go" = "routing_cache"
}

foreach ($file in $sourceFiles.Keys) {
    if (Test-Path $file) {
        $content = Get-Content $file -Raw
        $pattern = $sourceFiles[$file]
        
        if ($content -match $pattern) {
            Write-Host "  SUCCESS: $file contains routing cache code" -ForegroundColor Green
        } else {
            Write-Host "  WARNING: $file may not contain routing cache code" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  ERROR: $file not found" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Verification Complete ===" -ForegroundColor Cyan
Write-Host ""

# Recommendations
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "1. Add routing cache configuration to AdGuardHome.yaml:" -ForegroundColor White
Write-Host ""
Write-Host "dns:" -ForegroundColor Gray
Write-Host "  routing_cache_enabled: true" -ForegroundColor Gray
Write-Host "  routing_cache_size: 10000" -ForegroundColor Gray
Write-Host "  routing_cache_ttl: 5" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Start AdGuardHome:" -ForegroundColor White
Write-Host "   .\AdGuardHome_test.exe" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Test the configuration:" -ForegroundColor White
Write-Host "   .\test-cache-simple.ps1" -ForegroundColor Gray
