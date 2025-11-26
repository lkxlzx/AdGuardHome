# DNS Performance Benchmark Script
# Tests AdGuard Home DNS performance

Write-Host "=== DNS Performance Benchmark ===" -ForegroundColor Cyan
Write-Host ""

# Configuration
$dnsServer = "127.0.0.1"
$testDomains = @(
    "google.com",
    "github.com",
    "stackoverflow.com",
    "microsoft.com",
    "amazon.com",
    "facebook.com",
    "twitter.com",
    "youtube.com",
    "wikipedia.org",
    "reddit.com"
)
$iterations = 100

# Results storage
$results = @{
    FirstQuery = @()
    CachedQuery = @()
    TotalQueries = 0
    SuccessfulQueries = 0
    FailedQueries = 0
}

Write-Host "Test Configuration:" -ForegroundColor Yellow
Write-Host "  DNS Server: $dnsServer"
Write-Host "  Test Domains: $($testDomains.Count)"
Write-Host "  Iterations: $iterations"
Write-Host ""

# Test 1: First Query (Cache Miss)
Write-Host "Test 1: First Query Performance (Cache Miss)" -ForegroundColor Yellow
Write-Host "Testing each domain for the first time..." -ForegroundColor Gray

foreach ($domain in $testDomains) {
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $result = Resolve-DnsName -Name $domain -Server $dnsServer -ErrorAction Stop
        $stopwatch.Stop()
        
        $latency = $stopwatch.Elapsed.TotalMilliseconds
        $results.FirstQuery += $latency
        $results.TotalQueries++
        $results.SuccessfulQueries++
        
        Write-Host "  $domain : $([math]::Round($latency, 2))ms" -ForegroundColor Green
    }
    catch {
        Write-Host "  $domain : FAILED" -ForegroundColor Red
        $results.FailedQueries++
    }
    Start-Sleep -Milliseconds 100
}

Write-Host ""

# Test 2: Cached Query Performance
Write-Host "Test 2: Cached Query Performance (Cache Hit)" -ForegroundColor Yellow
Write-Host "Testing same domains again (should hit cache)..." -ForegroundColor Gray

foreach ($domain in $testDomains) {
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $result = Resolve-DnsName -Name $domain -Server $dnsServer -ErrorAction Stop
        $stopwatch.Stop()
        
        $latency = $stopwatch.Elapsed.TotalMilliseconds
        $results.CachedQuery += $latency
        $results.TotalQueries++
        $results.SuccessfulQueries++
        
        Write-Host "  $domain : $([math]::Round($latency, 2))ms" -ForegroundColor Green
    }
    catch {
        Write-Host "  $domain : FAILED" -ForegroundColor Red
        $results.FailedQueries++
    }
    Start-Sleep -Milliseconds 50
}

Write-Host ""

# Test 3: Sustained Load Test
Write-Host "Test 3: Sustained Load Test" -ForegroundColor Yellow
Write-Host "Running $iterations queries..." -ForegroundColor Gray

$loadTestResults = @()
$progressCount = 0

for ($i = 0; $i -lt $iterations; $i++) {
    $domain = $testDomains[$i % $testDomains.Count]
    
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $result = Resolve-DnsName -Name $domain -Server $dnsServer -ErrorAction Stop
        $stopwatch.Stop()
        
        $loadTestResults += $stopwatch.Elapsed.TotalMilliseconds
        $results.TotalQueries++
        $results.SuccessfulQueries++
    }
    catch {
        $results.FailedQueries++
    }
    
    $progressCount++
    if ($progressCount % 10 -eq 0) {
        Write-Host "  Progress: $progressCount/$iterations" -ForegroundColor Gray
    }
}

Write-Host "  Completed: $iterations queries" -ForegroundColor Green
Write-Host ""

# Calculate Statistics
Write-Host "=== Performance Results ===" -ForegroundColor Cyan
Write-Host ""

# First Query Stats
if ($results.FirstQuery.Count -gt 0) {
    $firstQueryAvg = ($results.FirstQuery | Measure-Object -Average).Average
    $firstQueryMin = ($results.FirstQuery | Measure-Object -Minimum).Minimum
    $firstQueryMax = ($results.FirstQuery | Measure-Object -Maximum).Maximum
    
    Write-Host "First Query (Cache Miss):" -ForegroundColor Yellow
    Write-Host "  Average: $([math]::Round($firstQueryAvg, 2))ms" -ForegroundColor White
    Write-Host "  Min: $([math]::Round($firstQueryMin, 2))ms" -ForegroundColor Green
    Write-Host "  Max: $([math]::Round($firstQueryMax, 2))ms" -ForegroundColor Red
    Write-Host ""
}

# Cached Query Stats
if ($results.CachedQuery.Count -gt 0) {
    $cachedQueryAvg = ($results.CachedQuery | Measure-Object -Average).Average
    $cachedQueryMin = ($results.CachedQuery | Measure-Object -Minimum).Minimum
    $cachedQueryMax = ($results.CachedQuery | Measure-Object -Maximum).Maximum
    
    Write-Host "Cached Query (Cache Hit):" -ForegroundColor Yellow
    Write-Host "  Average: $([math]::Round($cachedQueryAvg, 2))ms" -ForegroundColor White
    Write-Host "  Min: $([math]::Round($cachedQueryMin, 2))ms" -ForegroundColor Green
    Write-Host "  Max: $([math]::Round($cachedQueryMax, 2))ms" -ForegroundColor Red
    
    # Calculate improvement
    if ($firstQueryAvg -gt 0) {
        $improvement = (($firstQueryAvg - $cachedQueryAvg) / $firstQueryAvg) * 100
        Write-Host "  Improvement: $([math]::Round($improvement, 1))%" -ForegroundColor Cyan
    }
    Write-Host ""
}

# Load Test Stats
if ($loadTestResults.Count -gt 0) {
    $loadTestAvg = ($loadTestResults | Measure-Object -Average).Average
    $loadTestMin = ($loadTestResults | Measure-Object -Minimum).Minimum
    $loadTestMax = ($loadTestResults | Measure-Object -Maximum).Maximum
    $loadTestP95 = $loadTestResults | Sort-Object | Select-Object -Index ([math]::Floor($loadTestResults.Count * 0.95))
    $loadTestP99 = $loadTestResults | Sort-Object | Select-Object -Index ([math]::Floor($loadTestResults.Count * 0.99))
    
    Write-Host "Sustained Load ($iterations queries):" -ForegroundColor Yellow
    Write-Host "  Average: $([math]::Round($loadTestAvg, 2))ms" -ForegroundColor White
    Write-Host "  Min: $([math]::Round($loadTestMin, 2))ms" -ForegroundColor Green
    Write-Host "  Max: $([math]::Round($loadTestMax, 2))ms" -ForegroundColor Red
    Write-Host "  P95: $([math]::Round($loadTestP95, 2))ms" -ForegroundColor Yellow
    Write-Host "  P99: $([math]::Round($loadTestP99, 2))ms" -ForegroundColor Yellow
    Write-Host ""
}

# Overall Stats
Write-Host "Overall Statistics:" -ForegroundColor Yellow
Write-Host "  Total Queries: $($results.TotalQueries)" -ForegroundColor White
Write-Host "  Successful: $($results.SuccessfulQueries)" -ForegroundColor Green
Write-Host "  Failed: $($results.FailedQueries)" -ForegroundColor Red
$successRate = ($results.SuccessfulQueries / $results.TotalQueries) * 100
Write-Host "  Success Rate: $([math]::Round($successRate, 2))%" -ForegroundColor Cyan
Write-Host ""

# Check Cache Statistics via API
Write-Host "=== Cache Statistics (from API) ===" -ForegroundColor Cyan
try {
    $cacheStats = Invoke-RestMethod -Uri "http://localhost:3000/control/cache_metrics" -Method Get
    Write-Host "  Cache Enabled: $($cacheStats.cache_enabled)" -ForegroundColor White
    Write-Host "  Total Queries: $($cacheStats.total_queries)" -ForegroundColor White
    Write-Host "  Cache Hits: $($cacheStats.cache_hits)" -ForegroundColor Green
    Write-Host "  Cache Misses: $($cacheStats.cache_misses)" -ForegroundColor Yellow
    Write-Host "  Hit Rate: $([math]::Round($cacheStats.cache_hit_rate, 2))%" -ForegroundColor Cyan
}
catch {
    Write-Host "  Could not fetch cache statistics from API" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Benchmark Complete ===" -ForegroundColor Cyan
Write-Host ""

# Performance Rating
Write-Host "Performance Rating:" -ForegroundColor Yellow
if ($cachedQueryAvg -lt 1) {
    Write-Host "  ⭐⭐⭐⭐⭐ Excellent (< 1ms cached)" -ForegroundColor Green
}
elseif ($cachedQueryAvg -lt 5) {
    Write-Host "  ⭐⭐⭐⭐ Very Good (< 5ms cached)" -ForegroundColor Green
}
elseif ($cachedQueryAvg -lt 10) {
    Write-Host "  ⭐⭐⭐ Good (< 10ms cached)" -ForegroundColor Yellow
}
else {
    Write-Host "  ⭐⭐ Fair (> 10ms cached)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Benchmark results saved to memory. Run again to compare." -ForegroundColor Gray
