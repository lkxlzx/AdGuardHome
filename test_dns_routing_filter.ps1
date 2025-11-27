# Test DNS Routing Filter in Query Log
# This script tests the DNS routing filter functionality

Write-Host "=== Testing DNS Routing Filter ===" -ForegroundColor Cyan
Write-Host ""

# AdGuard Home API endpoint
$baseUrl = "http://localhost:3000"
$username = "admin"
$password = "admin"

# Create credentials
$pair = "${username}:${password}"
$bytes = [System.Text.Encoding]::ASCII.GetBytes($pair)
$base64 = [System.Convert]::ToBase64String($bytes)
$headers = @{
    Authorization = "Basic $base64"
}

Write-Host "Testing query log filter with response_status=dns_routing..." -ForegroundColor Yellow

try {
    # Test the DNS routing filter
    $response = Invoke-RestMethod -Uri "$baseUrl/control/querylog" `
        -Method POST `
        -Headers $headers `
        -ContentType "application/json" `
        -Body '{"response_status":"dns_routing","older_than":""}' `
        -ErrorAction Stop
    
    Write-Host "✓ DNS routing filter request successful!" -ForegroundColor Green
    Write-Host "  Found $($response.data.Count) DNS routing entries" -ForegroundColor Green
    
    if ($response.data.Count -gt 0) {
        Write-Host ""
        Write-Host "Sample DNS routing entries:" -ForegroundColor Cyan
        $response.data | Select-Object -First 3 | ForEach-Object {
            Write-Host "  - Domain: $($_.question.name)" -ForegroundColor White
            Write-Host "    Reason: $($_.reason)" -ForegroundColor Gray
            Write-Host "    Upstream: $($_.upstream)" -ForegroundColor Gray
            Write-Host ""
        }
    }
    
} catch {
    Write-Host "✗ Error testing DNS routing filter:" -ForegroundColor Red
    Write-Host "  $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.ErrorDetails.Message) {
        Write-Host "  Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    exit 1
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
