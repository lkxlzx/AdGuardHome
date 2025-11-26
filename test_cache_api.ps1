# Test Cache Metrics API
# Reads credentials from AdGuardHome.yaml

$configFile = "AdGuardHome.yaml"

if (-not (Test-Path $configFile)) {
    Write-Host "Error: $configFile not found!" -ForegroundColor Red
    exit 1
}

# Read config file and parse manually
$configContent = Get-Content $configFile

# Find bind address
$bindLine = $configContent | Where-Object { $_ -match "^\s*address:\s*(.+)" }
if ($bindLine) {
    $bindHost = $matches[1].Trim()
    if ($bindHost -match ":(\d+)$") {
        $port = $matches[1]
        $hostname = $bindHost -replace ":\d+$", ""
    } else {
        $port = 80
        $hostname = $bindHost
    }
} else {
    $hostname = "localhost"
    $port = 80
}

if ([string]::IsNullOrEmpty($hostname) -or $hostname -eq "0.0.0.0") {
    $hostname = "localhost"
}

# Find username
$usernameLine = $configContent | Where-Object { $_ -match "^\s*-\s*name:\s*(.+)" } | Select-Object -First 1
if ($usernameLine) {
    $username = $matches[1].Trim()
} else {
    $username = "admin"
}

# Use password directly
$password = "19821012"

Write-Host "Connecting to: http://${hostname}:${port}" -ForegroundColor Cyan
Write-Host "Username: $username" -ForegroundColor Cyan
Write-Host ""

# Create auth header
$base64AuthInfo = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes(("{0}:{1}" -f $username, $password)))

try {
    $response = Invoke-RestMethod -Uri "http://${hostname}:${port}/control/cache_metrics" -Method Get -Headers @{Authorization=("Basic {0}" -f $base64AuthInfo)}
    
    Write-Host "=== Cache Metrics Response ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Cache Enabled: $($response.cache_enabled)" -ForegroundColor Yellow
    Write-Host "Cache Hit Rate: $($response.cache_hit_rate)%" -ForegroundColor Yellow
    Write-Host "Total Queries: $($response.total_queries)" -ForegroundColor Yellow
    Write-Host "Cache Hits: $($response.cache_hits)" -ForegroundColor Yellow
    Write-Host "Cache Misses: $($response.cache_misses)" -ForegroundColor Yellow
    Write-Host ""
    
    Write-Host "History Data (24 points):" -ForegroundColor Cyan
    for ($i = 0; $i -lt $response.history.Length; $i++) {
        $value = $response.history[$i]
        if ($value -gt 0) {
            Write-Host "  [$i] $value%" -ForegroundColor Green
        } else {
            Write-Host "  [$i] $value%" -ForegroundColor Gray
        }
    }
    Write-Host ""
    Write-Host "Non-zero values: $(($response.history | Where-Object { $_ -gt 0 }).Count) / 24" -ForegroundColor Yellow
    
    Write-Host ""
    Write-Host "=== Full JSON Response ===" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 5
    
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please make sure:" -ForegroundColor Yellow
    Write-Host "1. AdGuard Home is running"
    Write-Host "2. The service is accessible at http://${hostname}:${port}"
    Write-Host "3. Credentials in $configFile are correct"
}
