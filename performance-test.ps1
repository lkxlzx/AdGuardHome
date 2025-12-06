# AdGuardHome v10.3 Performance Testing Script
# Tests DNS routing performance and identifies bottlenecks

$ErrorActionPreference = "Stop"

Write-Host "=== AdGuardHome v10.3 Performance Testing ===" -ForegroundColor Cyan
Write-Host ""

# Configuration
$TEST_DURATION = 30  # seconds
$CONCURRENT_QUERIES = 100
$DNS_SERVER = "127.0.0.1"
$DNS_PORT = 53
$OUTPUT_DIR = "performance-results"

# Create output directory
if (!(Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Test Duration: $TEST_DURATION seconds" -ForegroundColor Gray
Write-Host "  Concurrent Queries: $CONCURRENT_QUERIES" -ForegroundColor Gray
Write-Host "  DNS Server: ${DNS_SERVER}:${DNS_PORT}" -ForegroundColor Gray
Write-Host "  Output Directory: $OUTPUT_DIR" -ForegroundColor Gray
Write-Host ""

# Test domains
$TEST_DOMAINS = @(
    "google.com",
    "youtube.com",
    "facebook.com",
    "twitter.com",
    "github.com",
    "stackoverflow.com",
    "reddit.com",
    "amazon.com",
    "wikipedia.org",
    "microsoft.com"
)

# Function to measure DNS query time
function Measure-DnsQuery {
    param(
        [string]$Domain,
        [string]$Server = "127.0.0.1"
    )
    
    $startTime = Get-Date
    try {
        $result = Resolve-DnsName -Name $Domain -Server $Server -Type A -ErrorAction Stop -DnsOnly
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalMilliseconds
        return @{
            Success = $true
            Duration = $duration
            Domain = $Domain
        }
    } catch {
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalMilliseconds
        return @{
            Success = $false
            Duration = $duration
            Domain = $Domain
            Error = $_.Exception.Message
        }
    }
}

# Test 1: Single Query Latency
Write-Host "Test 1: Single Query Latency" -ForegroundColor Yellow
Write-Host "Testing each domain 10 times..." -ForegroundColor Gray

$singleQueryResults = @()
foreach ($domain in $TEST_DOMAINS) {
    Write-Host "  Testing $domain..." -ForegroundColor Gray
    $domainResults = @()
    
    for ($i = 0; $i -lt 10; $i++) {
        $result = Measure-DnsQuery -Domain $domain -Server $DNS_SERVER
        $domainResults += $result.Duration
        $singleQueryResults += $result
    }
    
    $avg = ($domainResults | Measure-Object -Average).Average
    $min = ($domainResults | Measure-Object -Minimum).Minimum
    $max = ($domainResults | Measure-Object -Maximum).Maximum
    
    Write-Host "    Avg: $([math]::Round($avg, 2))ms, Min: $([math]::Round($min, 2))ms, Max: $([math]::Round($max, 2))ms" -ForegroundColor Cyan
}

Write-Host ""

# Test 2: Concurrent Query Performance
Write-Host "Test 2: Concurrent Query Performance" -ForegroundColor Yellow
Write-Host "Running $CONCURRENT_QUERIES concurrent queries..." -ForegroundColor Gray

$concurrentResults = @()
$jobs = @()

$startTime = Get-Date

for ($i = 0; $i -lt $CONCURRENT_QUERIES; $i++) {
    $domain = $TEST_DOMAINS[$i % $TEST_DOMAINS.Count]
    
    $job = Start-Job -ScriptBlock {
        param($domain, $server)
        $start = Get-Date
        try {
            $result = Resolve-DnsName -Name $domain -Server $server -Type A -ErrorAction Stop -DnsOnly
            $end = Get-Date
            return @{
                Success = $true
                Duration = ($end - $start).TotalMilliseconds
                Domain = $domain
            }
        } catch {
            $end = Get-Date
            return @{
                Success = $false
                Duration = ($end - $start).TotalMilliseconds
                Domain = $domain
                Error = $_.Exception.Message
            }
        }
    } -ArgumentList $domain, $DNS_SERVER
    
    $jobs += $job
}

Write-Host "  Waiting for jobs to complete..." -ForegroundColor Gray
$jobResults = $jobs | Wait-Job | Receive-Job
$jobs | Remove-Job

$endTime = Get-Date
$totalDuration = ($endTime - $startTime).TotalSeconds

$successCount = ($jobResults | Where-Object { $_.Success }).Count
$failCount = $CONCURRENT_QUERIES - $successCount
$avgLatency = ($jobResults | Where-Object { $_.Success } | Measure-Object -Property Duration -Average).Average
$qps = $CONCURRENT_QUERIES / $totalDuration

Write-Host "  Total Duration: $([math]::Round($totalDuration, 2))s" -ForegroundColor Cyan
Write-Host "  Success: $successCount, Failed: $failCount" -ForegroundColor Cyan
Write-Host "  Average Latency: $([math]::Round($avgLatency, 2))ms" -ForegroundColor Cyan
Write-Host "  QPS: $([math]::Round($qps, 2))" -ForegroundColor Cyan

Write-Host ""

# Test 3: Memory and CPU Usage
Write-Host "Test 3: Resource Usage" -ForegroundColor Yellow
Write-Host "Monitoring AdGuardHome process..." -ForegroundColor Gray

$process = Get-Process -Name "AdGuardHome*" -ErrorAction SilentlyContinue | Select-Object -First 1

if ($process) {
    $cpuBefore = $process.CPU
    $memoryBefore = $process.WorkingSet64 / 1MB
    
    Write-Host "  Initial CPU Time: $([math]::Round($cpuBefore, 2))s" -ForegroundColor Gray
    Write-Host "  Initial Memory: $([math]::Round($memoryBefore, 2)) MB" -ForegroundColor Gray
    
    # Run load test
    Write-Host "  Running 30-second load test..." -ForegroundColor Gray
    $loadJobs = @()
    $loadStartTime = Get-Date
    
    while (((Get-Date) - $loadStartTime).TotalSeconds -lt 30) {
        for ($i = 0; $i -lt 10; $i++) {
            $domain = $TEST_DOMAINS[$i % $TEST_DOMAINS.Count]
            $job = Start-Job -ScriptBlock {
                param($domain, $server)
                Resolve-DnsName -Name $domain -Server $server -Type A -ErrorAction SilentlyContinue -DnsOnly | Out-Null
            } -ArgumentList $domain, $DNS_SERVER
            $loadJobs += $job
        }
        Start-Sleep -Milliseconds 100
    }
    
    $loadJobs | Wait-Job | Remove-Job
    
    # Refresh process info
    $process = Get-Process -Id $process.Id -ErrorAction SilentlyContinue
    
    if ($process) {
        $cpuAfter = $process.CPU
        $memoryAfter = $process.WorkingSet64 / 1MB
        
        $cpuUsed = $cpuAfter - $cpuBefore
        $memoryDiff = $memoryAfter - $memoryBefore
        
        Write-Host "  Final CPU Time: $([math]::Round($cpuAfter, 2))s" -ForegroundColor Cyan
        Write-Host "  Final Memory: $([math]::Round($memoryAfter, 2)) MB" -ForegroundColor Cyan
        Write-Host "  CPU Used: $([math]::Round($cpuUsed, 2))s" -ForegroundColor Cyan
        Write-Host "  Memory Change: $([math]::Round($memoryDiff, 2)) MB" -ForegroundColor Cyan
    }
} else {
    Write-Host "  Warning: AdGuardHome process not found" -ForegroundColor Yellow
}

Write-Host ""

# Generate Report
Write-Host "Generating Performance Report..." -ForegroundColor Yellow

$report = @"
# AdGuardHome v10.3 Performance Test Report

**Test Date**: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Test Duration**: $TEST_DURATION seconds
**Concurrent Queries**: $CONCURRENT_QUERIES

---

## Test Results Summary

### Test 1: Single Query Latency

| Domain | Avg (ms) | Min (ms) | Max (ms) |
|--------|----------|----------|----------|
"@

foreach ($domain in $TEST_DOMAINS) {
    $domainResults = $singleQueryResults | Where-Object { $_.Domain -eq $domain -and $_.Success }
    if ($domainResults.Count -gt 0) {
        $avg = ($domainResults | Measure-Object -Property Duration -Average).Average
        $min = ($domainResults | Measure-Object -Property Duration -Minimum).Minimum
        $max = ($domainResults | Measure-Object -Property Duration -Maximum).Maximum
        $report += "| $domain | $([math]::Round($avg, 2)) | $([math]::Round($min, 2)) | $([math]::Round($max, 2)) |`n"
    }
}

$overallAvg = ($singleQueryResults | Where-Object { $_.Success } | Measure-Object -Property Duration -Average).Average
$overallMin = ($singleQueryResults | Where-Object { $_.Success } | Measure-Object -Property Duration -Minimum).Minimum
$overallMax = ($singleQueryResults | Where-Object { $_.Success } | Measure-Object -Property Duration -Maximum).Maximum

$report += @"

**Overall Statistics:**
- Average Latency: $([math]::Round($overallAvg, 2)) ms
- Minimum Latency: $([math]::Round($overallMin, 2)) ms
- Maximum Latency: $([math]::Round($overallMax, 2)) ms

### Test 2: Concurrent Query Performance

- **Total Queries**: $CONCURRENT_QUERIES
- **Successful**: $successCount
- **Failed**: $failCount
- **Success Rate**: $([math]::Round(($successCount / $CONCURRENT_QUERIES) * 100, 2))%
- **Total Duration**: $([math]::Round($totalDuration, 2)) seconds
- **Average Latency**: $([math]::Round($avgLatency, 2)) ms
- **Queries Per Second (QPS)**: $([math]::Round($qps, 2))

### Test 3: Resource Usage

"@

if ($process) {
    $report += @"
- **Initial Memory**: $([math]::Round($memoryBefore, 2)) MB
- **Final Memory**: $([math]::Round($memoryAfter, 2)) MB
- **Memory Change**: $([math]::Round($memoryDiff, 2)) MB
- **CPU Time Used**: $([math]::Round($cpuUsed, 2)) seconds

"@
} else {
    $report += "- Process monitoring unavailable`n`n"
}

$report += @"
---

## Performance Analysis

### Bottleneck Identification

"@

# Analyze bottlenecks
$bottlenecks = @()

if ($overallAvg -gt 50) {
    $bottlenecks += "- **High Average Latency** ($([math]::Round($overallAvg, 2))ms): DNS query latency is higher than optimal (<50ms)"
}

if ($qps -lt 100) {
    $bottlenecks += "- **Low QPS** ($([math]::Round($qps, 2))): System can handle fewer than 100 queries per second"
}

if ($failCount -gt 0) {
    $bottlenecks += "- **Query Failures** ($failCount failures): Some queries failed during concurrent testing"
}

if ($process -and $memoryDiff -gt 50) {
    $bottlenecks += "- **Memory Growth** (+$([math]::Round($memoryDiff, 2))MB): Significant memory increase during load test"
}

if ($bottlenecks.Count -eq 0) {
    $report += "No significant bottlenecks detected. Performance is optimal.`n`n"
} else {
    $report += ($bottlenecks -join "`n") + "`n`n"
}

$report += @"
### Recommendations

"@

$recommendations = @()

if ($overallAvg -gt 50) {
    $recommendations += "1. **Optimize DNS Routing Logic**: Review router.Match() performance"
    $recommendations += "2. **Enable Caching**: Implement DNS response caching"
    $recommendations += "3. **Reduce Lock Contention**: Minimize mutex lock duration"
}

if ($qps -lt 100) {
    $recommendations += "1. **Increase Concurrency**: Use goroutine pools for parallel processing"
    $recommendations += "2. **Optimize Data Structures**: Use more efficient lookup structures"
}

if ($memoryDiff -gt 50) {
    $recommendations += "1. **Check Memory Leaks**: Review resource cleanup in file manager"
    $recommendations += "2. **Implement Memory Pooling**: Reuse objects to reduce allocations"
}

if ($recommendations.Count -eq 0) {
    $report += "No specific recommendations. System is performing well.`n"
} else {
    $report += ($recommendations -join "`n") + "`n"
}

$report += @"

---

**Report Generated**: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
"@

# Save report
$reportPath = Join-Path $OUTPUT_DIR "performance-report.md"
$report | Out-File -FilePath $reportPath -Encoding UTF8

Write-Host "Report saved to: $reportPath" -ForegroundColor Green
Write-Host ""
Write-Host "=== Performance Testing Complete ===" -ForegroundColor Cyan
