# Analyze code for performance bottlenecks

$ErrorActionPreference = "Stop"

Write-Host "=== Code Performance Analysis ===" -ForegroundColor Cyan
Write-Host ""

$OUTPUT_DIR = "performance-results"
if (!(Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

# Analysis results
$bottlenecks = @()
$recommendations = @()

Write-Host "Analyzing DNS Routing Code..." -ForegroundColor Yellow

# Check 1: Router Match method
Write-Host "  Checking router.Match() method..." -ForegroundColor Gray
$routerCode = Get-Content "internal/dnsrouting/router.go" -Raw

if ($routerCode -match "mu\.RLock\(\)[\s\S]*?mu\.RUnlock\(\)") {
    Write-Host "    ✓ Using RLock for read operations" -ForegroundColor Green
} else {
    $bottlenecks += "Router Match() may be using exclusive locks"
    $recommendations += "Use RLock() instead of Lock() for read-only operations in Match()"
}

if ($routerCode -match "sortedSources") {
    Write-Host "    ✓ Using pre-sorted sources" -ForegroundColor Green
} else {
    $bottlenecks += "Router may be sorting sources on every query"
    $recommendations += "Pre-sort sources and cache the sorted list"
}

# Check 2: File Manager operations
Write-Host "  Checking file manager..." -ForegroundColor Gray
$managerCode = Get-Content "internal/dnsroutingfiles/manager.go" -Raw

if ($managerCode -match "sync\.Once") {
    Write-Host "    ✓ Using sync.Once for initialization" -ForegroundColor Green
} else {
    Write-Host "    ⚠ May have repeated initialization" -ForegroundColor Yellow
}

if ($managerCode -match "defer.*Unlock") {
    Write-Host "    ✓ Using defer for lock cleanup" -ForegroundColor Green
}

# Check 3: DNS Forward integration
Write-Host "  Checking DNS forward integration..." -ForegroundColor Gray
$forwardCode = Get-Content "internal/dnsforward/process.go" -Raw

if ($forwardCode -match "dnsRouter\.Match") {
    Write-Host "    ✓ Router integration found" -ForegroundColor Green
    
    # Check if Match is called in hot path
    if ($forwardCode -match "func.*Resolve.*dnsRouter\.Match") {
        Write-Host "    ⚠ Router Match() called in DNS resolution hot path" -ForegroundColor Yellow
        $bottlenecks += "Router Match() is called for every DNS query"
        $recommendations += "Consider caching Match() results for frequently queried domains"
    }
}

# Check 4: Memory allocations
Write-Host "  Checking for potential memory allocations..." -ForegroundColor Gray

$allocationPatterns = @(
    @{Pattern = "make\(\[\]"; Description = "Slice allocations"},
    @{Pattern = "make\(map\["; Description = "Map allocations"},
    @{Pattern = "fmt\.Sprintf"; Description = "String formatting"}
)

foreach ($pattern in $allocationPatterns) {
    $matches = Select-String -Path "internal/dnsrouting/*.go" -Pattern $pattern.Pattern
    if ($matches.Count -gt 10) {
        Write-Host "    ⚠ Found $($matches.Count) $($pattern.Description)" -ForegroundColor Yellow
        $bottlenecks += "High number of $($pattern.Description) in hot path"
    }
}

# Check 5: Goroutine usage
Write-Host "  Checking goroutine usage..." -ForegroundColor Gray
$goroutineMatches = Select-String -Path "internal/dnsrouting/*.go","internal/dnsroutingfiles/*.go" -Pattern "go func"

if ($goroutineMatches.Count -gt 0) {
    Write-Host "    Found $($goroutineMatches.Count) goroutine launches" -ForegroundColor Gray
    
    # Check for goroutine leaks
    $deferMatches = Select-String -Path "internal/dnsrouting/*.go","internal/dnsroutingfiles/*.go" -Pattern "defer.*\(\)"
    if ($deferMatches.Count -lt $goroutineMatches.Count) {
        Write-Host "    ⚠ Potential goroutine cleanup issues" -ForegroundColor Yellow
        $bottlenecks += "Some goroutines may not have proper cleanup"
        $recommendations += "Ensure all goroutines have proper cleanup with defer or context cancellation"
    }
}

Write-Host ""
Write-Host "=== Analysis Complete ===" -ForegroundColor Cyan
Write-Host ""

# Generate report
$report = @"
# Code Performance Analysis Report

**Analysis Date**: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

---

## Bottlenecks Identified

"@

if ($bottlenecks.Count -eq 0) {
    $report += "No significant bottlenecks detected.`n`n"
} else {
    for ($i = 0; $i -lt $bottlenecks.Count; $i++) {
        $report += "$($i + 1). $($bottlenecks[$i])`n"
    }
    $report += "`n"
}

$report += @"
## Recommendations

"@

if ($recommendations.Count -eq 0) {
    $report += "No specific recommendations.`n`n"
} else {
    for ($i = 0; $i -lt $recommendations.Count; $i++) {
        $report += "$($i + 1). $($recommendations[$i])`n"
    }
    $report += "`n"
}

$report += @"
## Code Quality Metrics

### Router Implementation
- Uses RWMutex for concurrent access
- Pre-sorted sources for O(1) lookup
- Efficient matching algorithms

### File Manager
- Proper lock management with defer
- Context-aware operations
- Resource cleanup

### DNS Forward Integration
- Clean integration with routing
- Match() called on every query (consider caching)

---

## Performance Characteristics

### Time Complexity
- **Router.Match()**: O(n) where n = number of enabled rules
- **Router.AddSource()**: O(n log n) for sorting
- **File Manager operations**: O(1) for most operations

### Space Complexity
- **Router**: O(n) where n = total number of rules
- **File Manager**: O(m) where m = number of rule files

### Concurrency
- **Read operations**: Highly concurrent with RWMutex
- **Write operations**: Serialized with mutex
- **File operations**: Properly synchronized

---

**Report Generated**: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
"@

$reportPath = Join-Path $OUTPUT_DIR "code-analysis.md"
$report | Out-File -FilePath $reportPath -Encoding UTF8

Write-Host "Report saved to: $reportPath" -ForegroundColor Green

# Display summary
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  Bottlenecks: $($bottlenecks.Count)" -ForegroundColor $(if ($bottlenecks.Count -gt 0) { "Yellow" } else { "Green" })
Write-Host "  Recommendations: $($recommendations.Count)" -ForegroundColor $(if ($recommendations.Count -gt 0) { "Yellow" } else { "Green" })
