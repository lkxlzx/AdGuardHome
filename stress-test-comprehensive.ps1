# Comprehensive DNS Stress Test & Comparison
# Tests AdGuard Home against other DNS servers

param(
    [int]$Duration = 60,           # Test duration in seconds
    [int]$Concurrency = 10,        # Number of concurrent queries
    [int]$WarmupTime = 10,         # Warmup time in seconds
    [string]$AdGuardServer = "127.0.0.1",
    [string]$CompareServer = "",   # Optional: compare with another DNS server
    [switch]$ExportResults = $false
)

$ErrorActionPreference = "Continue"

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║     DNS Comprehensive Stress Test & Comparison            ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Test domains - mix of popular and various types
$testDomains = @(
    # Popular sites
    "google.com", "youtube.com", "facebook.com", "twitter.com", "instagram.com",
    "amazon.com", "wikipedia.org", "reddit.com", "netflix.com", "linkedin.com",
    
    # Tech sites
    "github.com", "stackoverflow.com", "microsoft.com", "apple.com", "cloudflare.com",
    
    # News sites
    "bbc.com", "cnn.com", "nytimes.com", "theguardian.com", "reuters.com",
    
    # CDN and services
    "cdn.jsdelivr.net", "ajax.googleapis.com", "fonts.googleapis.com",
    
    # Chinese sites
    "baidu.com", "taobao.com", "qq.com", "weibo.com", "jd.com"
)

# DNS servers to compare
$dnsServers = @{
    "AdGuard Home" = $AdGuardServer
    "Cloudflare" = "1.1.1.1"
    "Google" = "8.8.8.8"
    "Quad9" = "9.9.9.9"
}

if ($CompareServer) {
    $dnsServers["Custom"] = $CompareServer
}

Write-Host "Test Configuration:" -ForegroundColor Yellow
Write-Host "  Duration: $Duration seconds" -ForegroundColor White
Write-Host "  Concurrency: $Concurrency threads" -ForegroundColor White
Write-Host "  Warmup: $WarmupTime seconds" -ForegroundColor White
Write-Host "  Test Domains: $($testDomains.Count)" -ForegroundColor White
Write-Host "  DNS Servers: $($dnsServers.Count)" -ForegroundColor White
Write-Host ""

# Function to test a single DNS query
function Test-DnsQuery {
    param(
        [string]$Domain,
        [string]$Server
    )
    
    try {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        $result = Resolve-DnsName -Name $Domain -Server $Server -ErrorAction Stop -DnsOnly
        $stopwatch.Stop()
        
        return @{
            Success = $true
            Latency = $stopwatch.Elapsed.TotalMilliseconds
            Domain = $Domain
        }
    }
    catch {
        return @{
            Success = $false
            Latency = 0
            Domain = $Domain
            Error = $_.Exception.Message
        }
    }
}

# Function to run stress test on a DNS server
function Start-StressTest {
    param(
        [string]$ServerName,
        [string]$ServerIP,
        [int]$Duration,
        [int]$Concurrency,
        [int]$WarmupTime
    )
    
    Write-Host "Testing $ServerName ($ServerIP)..." -ForegroundColor Yellow
    
    $results = @{
        ServerName = $ServerName
        ServerIP = $ServerIP
        Queries = [System.Collections.ArrayList]::new()
        StartTime = Get-Date
        WarmupQueries = 0
        TestQueries = 0
        SuccessfulQueries = 0
        FailedQueries = 0
        Latencies = [System.Collections.ArrayList]::new()
    }
    
    # Warmup phase
    Write-Host "  Warmup phase ($WarmupTime seconds)..." -ForegroundColor Gray
    $warmupEnd = (Get-Date).AddSeconds($WarmupTime)
    
    while ((Get-Date) -lt $warmupEnd) {
        $domain = $testDomains | Get-Random
        $result = Test-DnsQuery -Domain $domain -Server $ServerIP
        $results.WarmupQueries++
        Start-Sleep -Milliseconds 50
    }
    
    Write-Host "  Warmup complete: $($results.WarmupQueries) queries" -ForegroundColor Green
    
    # Main test phase
    Write-Host "  Running stress test ($Duration seconds)..." -ForegroundColor Gray
    $testEnd = (Get-Date).AddSeconds($Duration)
    $lastProgress = Get-Date
    
    # Create concurrent jobs
    $jobs = @()
    for ($i = 0; $i -lt $Concurrency; $i++) {
        $job = Start-Job -ScriptBlock {
            param($domains, $server, $endTime)
            
            $results = @()
            while ((Get-Date) -lt $endTime) {
                $domain = $domains | Get-Random
                try {
                    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
                    $null = Resolve-DnsName -Name $domain -Server $server -ErrorAction Stop -DnsOnly
                    $stopwatch.Stop()
                    
                    $results += @{
                        Success = $true
                        Latency = $stopwatch.Elapsed.TotalMilliseconds
                        Time = Get-Date
                    }
                }
                catch {
                    $results += @{
                        Success = $false
                        Latency = 0
                        Time = Get-Date
                    }
                }
            }
            return $results
        } -ArgumentList $testDomains, $ServerIP, $testEnd
        
        $jobs += $job
    }
    
    # Wait for jobs and show progress
    while ($jobs | Where-Object { $_.State -eq 'Running' }) {
        if (((Get-Date) - $lastProgress).TotalSeconds -ge 5) {
            $elapsed = ((Get-Date) - $results.StartTime.AddSeconds($WarmupTime)).TotalSeconds
            $remaining = $Duration - $elapsed
            Write-Host "    Progress: $([math]::Round($elapsed, 0))s / ${Duration}s (${remaining}s remaining)" -ForegroundColor DarkGray
            $lastProgress = Get-Date
        }
        Start-Sleep -Milliseconds 500
    }
    
    # Collect results
    foreach ($job in $jobs) {
        $jobResults = Receive-Job -Job $job
        foreach ($r in $jobResults) {
            $results.TestQueries++
            if ($r.Success) {
                $results.SuccessfulQueries++
                [void]$results.Latencies.Add($r.Latency)
            }
            else {
                $results.FailedQueries++
            }
        }
        Remove-Job -Job $job
    }
    
    $results.EndTime = Get-Date
    
    Write-Host "  Test complete: $($results.TestQueries) queries" -ForegroundColor Green
    Write-Host ""
    
    return $results
}

# Function to calculate statistics
function Get-Statistics {
    param($Latencies)
    
    if ($Latencies.Count -eq 0) {
        return $null
    }
    
    $sorted = $Latencies | Sort-Object
    
    return @{
        Count = $Latencies.Count
        Min = $sorted[0]
        Max = $sorted[-1]
        Average = ($Latencies | Measure-Object -Average).Average
        Median = $sorted[[math]::Floor($sorted.Count / 2)]
        P95 = $sorted[[math]::Floor($sorted.Count * 0.95)]
        P99 = $sorted[[math]::Floor($sorted.Count * 0.99)]
        StdDev = [math]::Sqrt((($Latencies | ForEach-Object { [math]::Pow($_ - ($Latencies | Measure-Object -Average).Average, 2) } | Measure-Object -Sum).Sum / $Latencies.Count))
    }
}

# Run tests on all DNS servers
$allResults = @{}

foreach ($server in $dnsServers.GetEnumerator()) {
    $result = Start-StressTest -ServerName $server.Key -ServerIP $server.Value -Duration $Duration -Concurrency $Concurrency -WarmupTime $WarmupTime
    $allResults[$server.Key] = $result
}

# Display results
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                    Test Results                            ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$comparisonTable = @()

foreach ($serverName in $allResults.Keys | Sort-Object) {
    $result = $allResults[$serverName]
    $stats = Get-Statistics -Latencies $result.Latencies
    
    Write-Host "═══ $serverName ($($result.ServerIP)) ═══" -ForegroundColor Yellow
    Write-Host ""
    
    # Query statistics
    Write-Host "Query Statistics:" -ForegroundColor Cyan
    Write-Host "  Total Queries: $($result.TestQueries)" -ForegroundColor White
    Write-Host "  Successful: $($result.SuccessfulQueries)" -ForegroundColor Green
    Write-Host "  Failed: $($result.FailedQueries)" -ForegroundColor $(if ($result.FailedQueries -gt 0) { "Red" } else { "Green" })
    
    $successRate = if ($result.TestQueries -gt 0) { ($result.SuccessfulQueries / $result.TestQueries) * 100 } else { 0 }
    Write-Host "  Success Rate: $([math]::Round($successRate, 2))%" -ForegroundColor Cyan
    
    $qps = $result.TestQueries / $Duration
    Write-Host "  Queries/Second: $([math]::Round($qps, 2))" -ForegroundColor White
    Write-Host ""
    
    # Latency statistics
    if ($stats) {
        Write-Host "Latency Statistics:" -ForegroundColor Cyan
        Write-Host "  Average: $([math]::Round($stats.Average, 2))ms" -ForegroundColor White
        Write-Host "  Median: $([math]::Round($stats.Median, 2))ms" -ForegroundColor White
        Write-Host "  Min: $([math]::Round($stats.Min, 2))ms" -ForegroundColor Green
        Write-Host "  Max: $([math]::Round($stats.Max, 2))ms" -ForegroundColor Red
        Write-Host "  P95: $([math]::Round($stats.P95, 2))ms" -ForegroundColor Yellow
        Write-Host "  P99: $([math]::Round($stats.P99, 2))ms" -ForegroundColor Yellow
        Write-Host "  StdDev: $([math]::Round($stats.StdDev, 2))ms" -ForegroundColor Gray
        
        $comparisonTable += [PSCustomObject]@{
            Server = $serverName
            QPS = [math]::Round($qps, 2)
            SuccessRate = [math]::Round($successRate, 2)
            AvgLatency = [math]::Round($stats.Average, 2)
            MedianLatency = [math]::Round($stats.Median, 2)
            P95Latency = [math]::Round($stats.P95, 2)
            P99Latency = [math]::Round($stats.P99, 2)
        }
    }
    
    Write-Host ""
}

# Comparison table
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                  Performance Comparison                    ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$comparisonTable | Format-Table -AutoSize

# Rankings
Write-Host "Rankings:" -ForegroundColor Yellow
Write-Host ""

Write-Host "🏆 Best Average Latency:" -ForegroundColor Cyan
$bestAvg = $comparisonTable | Sort-Object AvgLatency | Select-Object -First 3
for ($i = 0; $i -lt $bestAvg.Count; $i++) {
    $medal = @("🥇", "🥈", "🥉")[$i]
    Write-Host "  $medal $($bestAvg[$i].Server): $($bestAvg[$i].AvgLatency)ms" -ForegroundColor White
}
Write-Host ""

Write-Host "🏆 Best P95 Latency:" -ForegroundColor Cyan
$bestP95 = $comparisonTable | Sort-Object P95Latency | Select-Object -First 3
for ($i = 0; $i -lt $bestP95.Count; $i++) {
    $medal = @("🥇", "🥈", "🥉")[$i]
    Write-Host "  $medal $($bestP95[$i].Server): $($bestP95[$i].P95Latency)ms" -ForegroundColor White
}
Write-Host ""

Write-Host "🏆 Best Throughput (QPS):" -ForegroundColor Cyan
$bestQPS = $comparisonTable | Sort-Object QPS -Descending | Select-Object -First 3
for ($i = 0; $i -lt $bestQPS.Count; $i++) {
    $medal = @("🥇", "🥈", "🥉")[$i]
    Write-Host "  $medal $($bestQPS[$i].Server): $($bestQPS[$i].QPS) queries/sec" -ForegroundColor White
}
Write-Host ""

# AdGuard Home specific metrics
if ($allResults.ContainsKey("AdGuard Home")) {
    Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║            AdGuard Home Detailed Metrics                   ║" -ForegroundColor Cyan
    Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
    
    try {
        # Get cache statistics
        $cacheStats = Invoke-RestMethod -Uri "http://localhost:3000/control/cache_metrics" -Method Get -ErrorAction Stop
        
        Write-Host "Cache Performance:" -ForegroundColor Yellow
        Write-Host "  Cache Enabled: $($cacheStats.cache_enabled)" -ForegroundColor White
        Write-Host "  Total Queries: $($cacheStats.total_queries)" -ForegroundColor White
        Write-Host "  Cache Hits: $($cacheStats.cache_hits)" -ForegroundColor Green
        Write-Host "  Cache Misses: $($cacheStats.cache_misses)" -ForegroundColor Yellow
        Write-Host "  Hit Rate: $([math]::Round($cacheStats.cache_hit_rate, 2))%" -ForegroundColor Cyan
        Write-Host ""
        
        # Get stats
        $stats = Invoke-RestMethod -Uri "http://localhost:3000/control/stats" -Method Get -ErrorAction Stop
        
        Write-Host "Overall Statistics:" -ForegroundColor Yellow
        Write-Host "  DNS Queries: $($stats.num_dns_queries)" -ForegroundColor White
        Write-Host "  Blocked: $($stats.num_blocked_filtering)" -ForegroundColor Red
        Write-Host "  Avg Processing Time: $([math]::Round($stats.avg_processing_time * 1000, 2))ms" -ForegroundColor Cyan
        Write-Host ""
    }
    catch {
        Write-Host "Could not fetch AdGuard Home metrics: $_" -ForegroundColor Red
        Write-Host ""
    }
}

# Export results if requested
if ($ExportResults) {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $exportFile = "stress_test_results_$timestamp.json"
    
    $exportData = @{
        TestConfig = @{
            Duration = $Duration
            Concurrency = $Concurrency
            WarmupTime = $WarmupTime
            TestDomains = $testDomains.Count
        }
        Results = $allResults
        Comparison = $comparisonTable
        Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    }
    
    $exportData | ConvertTo-Json -Depth 10 | Out-File -FilePath $exportFile -Encoding UTF8
    Write-Host "Results exported to: $exportFile" -ForegroundColor Green
    Write-Host ""
}

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                  Test Complete                             ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Performance verdict
$aghResult = $comparisonTable | Where-Object { $_.Server -eq "AdGuard Home" }
if ($aghResult) {
    $avgRank = ($comparisonTable | Sort-Object AvgLatency).IndexOf($aghResult) + 1
    $qpsRank = ($comparisonTable | Sort-Object QPS -Descending).IndexOf($aghResult) + 1
    
    Write-Host "AdGuard Home Performance Verdict:" -ForegroundColor Yellow
    Write-Host "  Latency Rank: #$avgRank of $($comparisonTable.Count)" -ForegroundColor White
    Write-Host "  Throughput Rank: #$qpsRank of $($comparisonTable.Count)" -ForegroundColor White
    
    if ($avgRank -eq 1 -and $qpsRank -eq 1) {
        Write-Host "  🏆 CHAMPION - Best overall performance!" -ForegroundColor Green
    }
    elseif ($avgRank -le 2 -or $qpsRank -le 2) {
        Write-Host "  ⭐ EXCELLENT - Top tier performance!" -ForegroundColor Green
    }
    elseif ($avgRank -le 3 -or $qpsRank -le 3) {
        Write-Host "  ✓ GOOD - Competitive performance" -ForegroundColor Cyan
    }
    else {
        Write-Host "  ℹ FAIR - Room for improvement" -ForegroundColor Yellow
    }
}

Write-Host ""
