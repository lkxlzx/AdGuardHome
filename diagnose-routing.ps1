# Diagnostic script to check DNS routing configuration and behavior

$baseUrl = "http://localhost/control"
$auth = @{
    Authorization = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("lkxlzx:19821012"))
}

Write-Host "=== DNS Routing Diagnostic ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Get upstream groups
Write-Host "Step 1: Checking upstream groups..." -ForegroundColor Yellow
try {
    $groups = Invoke-RestMethod -Uri "$baseUrl/dns/upstream_groups" -Method Get -Headers $auth
    Write-Host "✓ Found $($groups.Count) upstream groups:" -ForegroundColor Green
    foreach ($group in $groups) {
        Write-Host "  - ID: $($group.id)" -ForegroundColor White
        Write-Host "    Name: $($group.name)" -ForegroundColor Gray
        Write-Host "    Upstreams: $($group.upstream_dns -join ', ')" -ForegroundColor Gray
        Write-Host "    Default: $($group.is_default)" -ForegroundColor Gray
        Write-Host ""
    }
} catch {
    Write-Host "✗ Failed to get upstream groups: $_" -ForegroundColor Red
}

# Step 2: Get DNS routing rules
Write-Host "Step 2: Checking DNS routing rules..." -ForegroundColor Yellow
try {
    $rules = Invoke-RestMethod -Uri "$baseUrl/dns_routing/rules" -Method Get -Headers $auth
    Write-Host "✓ Found $($rules.Count) routing rules:" -ForegroundColor Green
    foreach ($rule in $rules) {
        $groupName = ($groups | Where-Object { $_.id -eq $rule.upstream_group }).name
        Write-Host "  - Name: $($rule.name)" -ForegroundColor White
        Write-Host "    ID: $($rule.id)" -ForegroundColor Gray
        Write-Host "    Enabled: $($rule.enabled)" -ForegroundColor Gray
        Write-Host "    Upstream Group ID: $($rule.upstream_group)" -ForegroundColor Gray
        Write-Host "    Upstream Group Name: $groupName" -ForegroundColor $(if ($groupName) { "Green" } else { "Red" })
        Write-Host "    Rules Count: $($rule.rules_count)" -ForegroundColor Gray
        Write-Host "    Priority: $($rule.priority)" -ForegroundColor Gray
        Write-Host ""
    }
} catch {
    Write-Host "✗ Failed to get routing rules: $_" -ForegroundColor Red
}

# Step 3: Test specific domains
Write-Host "Step 3: Testing domain resolution..." -ForegroundColor Yellow
Write-Host ""

$testDomains = @(
    @{ Domain = "google.com"; ExpectedGroup = "GFW (should use configured group)" },
    @{ Domain = "baidu.com"; ExpectedGroup = "CN (should use configured group)" },
    @{ Domain = "youtube.com"; ExpectedGroup = "GFW (should use configured group)" }
)

foreach ($test in $testDomains) {
    Write-Host "Testing: $($test.Domain)" -ForegroundColor White
    Write-Host "  Expected: $($test.ExpectedGroup)" -ForegroundColor Gray
    
    try {
        $result = Resolve-DnsName -Name $test.Domain -Server 127.0.0.1 -ErrorAction Stop
        if ($result) {
            Write-Host "  ✓ Resolved to: $($result[0].IPAddress)" -ForegroundColor Green
        }
    } catch {
        Write-Host "  ⚠ Resolution failed or NXDOMAIN" -ForegroundColor Yellow
    }
    Write-Host ""
}

Write-Host "=== Diagnostic Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Analysis:" -ForegroundColor White
Write-Host "1. Check if upstream group IDs match between rules and groups" -ForegroundColor Gray
Write-Host "2. Verify that rules are enabled" -ForegroundColor Gray
Write-Host "3. Check server logs for routing decisions" -ForegroundColor Gray
Write-Host ""
Write-Host "To see detailed routing logs, check the AdGuardHome console output" -ForegroundColor Yellow
