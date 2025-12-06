# Script to fix reversed DNS routing configuration

$baseUrl = "http://localhost/control"
$auth = @{
    Authorization = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("lkxlzx:19821012"))
}

Write-Host "=== Fixing DNS Routing Configuration ===" -ForegroundColor Cyan
Write-Host ""

# Get current rules
Write-Host "Step 1: Getting current DNS routing rules..." -ForegroundColor Yellow
try {
    $rules = Invoke-RestMethod -Uri "$baseUrl/dns_routing/rules" -Method Get -Headers $auth
    Write-Host "✓ Found $($rules.Count) rules" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to get rules: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Display current configuration
Write-Host "Current Configuration:" -ForegroundColor White
foreach ($rule in $rules) {
    Write-Host "  - $($rule.name): upstream_group = $($rule.upstream_group)" -ForegroundColor Gray
}

Write-Host ""

# Fix CN rule (should use 国内 DNS)
Write-Host "Step 2: Fixing CN rule to use 国内 DNS..." -ForegroundColor Yellow
$cnRule = $rules | Where-Object { $_.name -eq "CN" }
if ($cnRule) {
    $updateBody = @{
        id = $cnRule.id
        name = $cnRule.name
        url = $cnRule.url
        upstream_group = "708dd52d-f6fb-4863-bfce-75a7f33c899d"  # 国内
        update_interval = $cnRule.update_interval
        priority = $cnRule.priority
        enabled = $cnRule.enabled
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$baseUrl/dns_routing/update" -Method Post -Headers $auth -Body $updateBody -ContentType "application/json" | Out-Null
        Write-Host "✓ CN rule updated to use 国内 DNS (223.6.6.6)" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to update CN rule: $_" -ForegroundColor Red
    }
} else {
    Write-Host "⚠ CN rule not found" -ForegroundColor Yellow
}

Write-Host ""

# Fix GFW rule (should use 海外 DNS)
Write-Host "Step 3: Fixing GFW rule to use 海外 DNS..." -ForegroundColor Yellow
$gfwRule = $rules | Where-Object { $_.name -eq "GFW" }
if ($gfwRule) {
    $updateBody = @{
        id = $gfwRule.id
        name = $gfwRule.name
        url = $gfwRule.url
        upstream_group = "fedcfe05-772b-45df-b190-84382427ce66"  # 海外
        update_interval = $gfwRule.update_interval
        priority = $gfwRule.priority
        enabled = $gfwRule.enabled
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$baseUrl/dns_routing/update" -Method Post -Headers $auth -Body $updateBody -ContentType "application/json" | Out-Null
        Write-Host "✓ GFW rule updated to use 海外 DNS (Cloudflare/Google)" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to update GFW rule: $_" -ForegroundColor Red
    }
} else {
    Write-Host "⚠ GFW rule not found" -ForegroundColor Yellow
}

Write-Host ""

# Verify the fix
Write-Host "Step 4: Verifying configuration..." -ForegroundColor Yellow
try {
    $rules = Invoke-RestMethod -Uri "$baseUrl/dns_routing/rules" -Method Get -Headers $auth
    Write-Host "✓ Configuration verified" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to verify: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Updated Configuration:" -ForegroundColor White
foreach ($rule in $rules) {
    $upstreamName = if ($rule.upstream_group -eq "708dd52d-f6fb-4863-bfce-75a7f33c899d") { "国内 (223.6.6.6)" } else { "海外 (Cloudflare/Google)" }
    Write-Host "  - $($rule.name): $upstreamName" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Configuration Fixed ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor White
Write-Host "- CN (中国域名) → 国内 DNS (223.6.6.6)" -ForegroundColor Green
Write-Host "- GFW (被墙域名) → 海外 DNS (Cloudflare/Google)" -ForegroundColor Green
Write-Host ""
Write-Host "The rules will take effect immediately!" -ForegroundColor Green
