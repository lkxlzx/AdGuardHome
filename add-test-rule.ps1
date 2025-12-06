# Add test DNS routing rule with authentication

$baseUrl = "http://127.0.0.1:80"
$username = "lkxlzx"
$password = "19821012"

# Create credentials
$pair = "$($username):$($password)"
$encodedCreds = [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($pair))
$headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Basic $encodedCreds"
}

Write-Host "Adding custom domain rule for baidu.com..." -ForegroundColor Yellow

$body = @{
    domain = "baidu.com"
    matchType = "DOMAIN-SUFFIX"
    upstreamGroup = "708dd52d-f6fb-4863-bfce-75a7f33c899d"
    enabled = $true
} | ConvertTo-Json

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/control/dns_routing/custom_rules/add" -Method Post -Headers $headers -Body $body
    Write-Host "Success! Added rule:" -ForegroundColor Green
    Write-Host "  Domain: $($result.domain)" -ForegroundColor Gray
    Write-Host "  Match Type: $($result.matchType)" -ForegroundColor Gray
    Write-Host "  Upstream Group: $($result.upstreamGroup)" -ForegroundColor Gray
    Write-Host "  Enabled: $($result.enabled)" -ForegroundColor Gray
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Red
    }
}
