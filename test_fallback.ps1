# Test fallback DNS mechanism
# This script tests if fallback DNS works when primary upstream fails

Write-Host "Testing Fallback DNS Mechanism" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Test domain
$testDomain = "www.google.com"

Write-Host "Test Configuration:" -ForegroundColor Yellow
Write-Host "- Primary Upstream: 114.114.114.111" -ForegroundColor White
Write-Host "- Fallback Upstream: 223.6.6.6" -ForegroundColor White
Write-Host "- Test Domain: $testDomain" -ForegroundColor White
Write-Host ""

# Test 1: Query when primary is working
Write-Host "[Test 1] Normal query (primary should respond):" -ForegroundColor Green
$result1 = Measure-Command {
    $response1 = Resolve-DnsName -Name $testDomain -Server "127.0.0.1" -DnsOnly -ErrorAction SilentlyContinue
}
if ($response1) {
    Write-Host "  ✓ Success - Response time: $($result1.TotalMilliseconds)ms" -ForegroundColor Green
    Write-Host "  IP: $($response1[0].IPAddress)" -ForegroundColor White
} else {
    Write-Host "  ✗ Failed" -ForegroundColor Red
}
Write-Host ""

# Test 2: Simulate primary failure by using a non-existent upstream
Write-Host "[Test 2] Testing fallback mechanism:" -ForegroundColor Green
Write-Host "  Note: To properly test fallback, you need to:" -ForegroundColor Yellow
Write-Host "  1. Temporarily change primary upstream to an invalid IP (e.g., 192.0.2.1)" -ForegroundColor Yellow
Write-Host "  2. Keep fallback as 223.6.6.6" -ForegroundColor Yellow
Write-Host "  3. Restart AdGuardHome" -ForegroundColor Yellow
Write-Host "  4. Query should still work using fallback" -ForegroundColor Yellow
Write-Host ""

# Test 3: Direct query to fallback server
Write-Host "[Test 3] Direct query to fallback server (223.6.6.6):" -ForegroundColor Green
$result3 = Measure-Command {
    $response3 = Resolve-DnsName -Name $testDomain -Server "223.6.6.6" -DnsOnly -ErrorAction SilentlyContinue
}
if ($response3) {
    Write-Host "  ✓ Success - Response time: $($result3.TotalMilliseconds)ms" -ForegroundColor Green
    Write-Host "  IP: $($response3[0].IPAddress)" -ForegroundColor White
} else {
    Write-Host "  ✗ Failed" -ForegroundColor Red
}
Write-Host ""

Write-Host "Recommendation:" -ForegroundColor Cyan
Write-Host "- Fallback DNS only activates when ALL primary upstreams fail" -ForegroundColor White
Write-Host "- If primary upstream is slow but responds, fallback won't be used" -ForegroundColor White
Write-Host "- Consider adding multiple primary upstreams for better reliability" -ForegroundColor White
Write-Host ""
Write-Host "Example configuration:" -ForegroundColor Yellow
Write-Host "  Primary Upstreams: 114.114.114.111, 114.114.115.115" -ForegroundColor White
Write-Host "  Fallback Upstreams: 223.6.6.6, 223.5.5.5" -ForegroundColor White
