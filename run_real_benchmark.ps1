# AdGuardHome V5.1 Real Performance Benchmark
# Test Date: 2025-11-28

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "AdGuardHome V5.1 Real Performance Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$outputFile = "REAL_BENCHMARK_RESULTS.txt"
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Clear output file
"AdGuardHome V5.1 Real Performance Test Report" | Out-File $outputFile
"Test Time: $timestamp" | Out-File $outputFile -Append
"Test Environment: Windows, Intel i5-8259U, 8 cores" | Out-File $outputFile -Append
"" | Out-File $outputFile -Append

# Test 1: Domain Matching Performance
Write-Host "Test 1: Domain Matching Performance..." -ForegroundColor Yellow
"========================================" | Out-File $outputFile -Append
"Test 1: Domain Matching Performance" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
cd internal/dnsforward
go test -bench=BenchmarkMatchDomainPattern -benchmem -benchtime=10s -run=^$ | Tee-Object -Append -FilePath "../../$outputFile"
cd ../..

Write-Host ""

# Test 2: Prefetch Performance
Write-Host "Test 2: Prefetch Performance..." -ForegroundColor Yellow
"" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
"Test 2: Prefetch Performance" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
cd internal/dnsforward
go test -bench=BenchmarkPrefetch_Record -benchmem -benchtime=10s -run=^$ | Tee-Object -Append -FilePath "../../$outputFile"
cd ../..

Write-Host ""

# Test 3: Prefetch Cleanup Performance
Write-Host "Test 3: Prefetch Cleanup Performance..." -ForegroundColor Yellow
"" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
"Test 3: Prefetch Cleanup Performance" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
cd internal/dnsforward
go test -bench=BenchmarkPrefetch_Cleanup -benchmem -benchtime=10s -run=^$ | Tee-Object -Append -FilePath "../../$outputFile"
cd ../..

Write-Host ""

# Test 4: Cache Hit Performance
Write-Host "Test 4: Cache Hit Performance..." -ForegroundColor Yellow
"" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
"Test 4: Cache Hit Performance" | Out-File $outputFile -Append
"========================================" | Out-File $outputFile -Append
cd internal/dnsforward
go test -bench=BenchmarkParsePattern_CacheHit -benchmem -benchtime=10s -run=^$ | Tee-Object -Append -FilePath "../../$outputFile"
cd ../..

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Test Complete! Results saved to: $outputFile" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
