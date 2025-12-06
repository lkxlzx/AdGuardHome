# Run Go benchmarks and generate performance profile

$ErrorActionPreference = "Stop"

Write-Host "=== Running Go Benchmarks ===" -ForegroundColor Cyan
Write-Host ""

$OUTPUT_DIR = "performance-results"
if (!(Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

# Test 1: DNS Routing Benchmarks
Write-Host "Test 1: DNS Routing Benchmarks" -ForegroundColor Yellow
go test -bench=. -benchmem -benchtime=10s ./internal/dnsrouting/... | Tee-Object -FilePath "$OUTPUT_DIR/routing-bench.txt"

Write-Host ""

# Test 2: File Manager Benchmarks
Write-Host "Test 2: File Manager Benchmarks" -ForegroundColor Yellow
go test -bench=. -benchmem -benchtime=10s ./internal/dnsroutingfiles/... | Tee-Object -FilePath "$OUTPUT_DIR/filemanager-bench.txt"

Write-Host ""

# Test 3: CPU Profile
Write-Host "Test 3: Generating CPU Profile" -ForegroundColor Yellow
go test -bench=BenchmarkMatch -cpuprofile="$OUTPUT_DIR/cpu.prof" -benchtime=30s ./internal/dnsrouting/

Write-Host ""

# Test 4: Memory Profile
Write-Host "Test 4: Generating Memory Profile" -ForegroundColor Yellow
go test -bench=BenchmarkMatch -memprofile="$OUTPUT_DIR/mem.prof" -benchtime=30s ./internal/dnsrouting/

Write-Host ""

# Test 5: Analyze profiles
Write-Host "Test 5: Analyzing Profiles" -ForegroundColor Yellow

if (Test-Path "$OUTPUT_DIR/cpu.prof") {
    Write-Host "  Analyzing CPU profile..." -ForegroundColor Gray
    go tool pprof -text -top10 "$OUTPUT_DIR/cpu.prof" | Tee-Object -FilePath "$OUTPUT_DIR/cpu-analysis.txt"
}

if (Test-Path "$OUTPUT_DIR/mem.prof") {
    Write-Host "  Analyzing Memory profile..." -ForegroundColor Gray
    go tool pprof -text -top10 "$OUTPUT_DIR/mem.prof" | Tee-Object -FilePath "$OUTPUT_DIR/mem-analysis.txt"
}

Write-Host ""
Write-Host "=== Benchmarks Complete ===" -ForegroundColor Cyan
Write-Host "Results saved to: $OUTPUT_DIR/" -ForegroundColor Green
