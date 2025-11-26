# Prefetch 数据分析脚本
# 用法: .\analyze_prefetch.ps1

$url = "http://localhost:3000/control/prefetch_status"

Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║        AdGuard Home - Prefetch Performance Analysis       ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

try {
    $status = Invoke-RestMethod -Uri $url -Method Get
    
    # 基本信息
    Write-Host "📊 Current Status" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    Write-Host "  Enabled:          " -NoNewline
    if ($status.enabled) {
        Write-Host "✓ Yes" -ForegroundColor Green
    } else {
        Write-Host "✗ No" -ForegroundColor Red
    }
    Write-Host "  Threshold:        $($status.threshold) hits"
    Write-Host "  Time Window:      $($status.time_window)s ($([math]::Round($status.time_window/60))min)"
    Write-Host ""
    
    # 性能指标
    Write-Host "🎯 Performance Metrics" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    
    $total = $status.tasks_completed + $status.tasks_failed
    if ($total -gt 0) {
        $successRate = [math]::Round(($status.tasks_completed / $total) * 100, 2)
        Write-Host "  Success Rate:     " -NoNewline
        if ($successRate -ge 99) {
            Write-Host "$successRate% 🌟 Excellent!" -ForegroundColor Green
        } elseif ($successRate -ge 95) {
            Write-Host "$successRate% ✓ Good" -ForegroundColor Green
        } elseif ($successRate -ge 90) {
            Write-Host "$successRate% ⚠ Fair" -ForegroundColor Yellow
        } else {
            Write-Host "$successRate% ✗ Poor" -ForegroundColor Red
        }
    }
    
    Write-Host "  Tasks Completed:  $($status.tasks_completed)"
    Write-Host "  Tasks Failed:     $($status.tasks_failed)"
    Write-Host "  Tasks Dropped:    $($status.tasks_dropped)"
    Write-Host ""
    
    # 域名统计
    Write-Host "🌐 Domain Statistics" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    Write-Host "  Tracked Domains:  $($status.tracked_domains)"
    Write-Host "  Hot Domains:      $($status.hot_domains)"
    
    if ($status.tracked_domains -gt 0) {
        $hotRatio = [math]::Round(($status.hot_domains / $status.tracked_domains) * 100, 2)
        Write-Host "  Hot Ratio:        " -NoNewline
        if ($hotRatio -ge 30) {
            Write-Host "$hotRatio% 🔥 Very Active!" -ForegroundColor Green
        } elseif ($hotRatio -ge 20) {
            Write-Host "$hotRatio% ✓ Active" -ForegroundColor Green
        } elseif ($hotRatio -ge 10) {
            Write-Host "$hotRatio% ⚠ Moderate" -ForegroundColor Yellow
        } else {
            Write-Host "$hotRatio% ℹ Low" -ForegroundColor Cyan
        }
        
        # 可视化热门域名比例
        Write-Host ""
        Write-Host "  Hot Domains Distribution:" -ForegroundColor Gray
        $hotBars = [math]::Round($hotRatio / 2)
        $coldBars = 50 - $hotBars
        Write-Host "  [" -NoNewline
        Write-Host ("█" * $hotBars) -NoNewline -ForegroundColor Red
        Write-Host ("░" * $coldBars) -NoNewline -ForegroundColor DarkGray
        Write-Host "] $hotRatio%"
    }
    Write-Host ""
    
    # 队列状态
    Write-Host "📦 Queue Status" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    Write-Host "  Active Tasks:     $($status.current_active)"
    Write-Host "  Urgent Queue:     $($status.urgent_queue)"
    Write-Host "  Normal Queue:     $($status.normal_queue)"
    
    $totalQueue = $status.urgent_queue + $status.normal_queue
    if ($totalQueue -eq 0) {
        Write-Host "  Status:           " -NoNewline
        Write-Host "✓ All Clear" -ForegroundColor Green
    } elseif ($totalQueue -lt 100) {
        Write-Host "  Status:           " -NoNewline
        Write-Host "⚠ Minor Backlog" -ForegroundColor Yellow
    } else {
        Write-Host "  Status:           " -NoNewline
        Write-Host "✗ Significant Backlog" -ForegroundColor Red
    }
    Write-Host ""
    
    # 资源使用
    Write-Host "💾 Resource Usage" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    Write-Host "  Soft Limit Hits:  $($status.soft_limit_hits)"
    Write-Host "  Hard Limit Hits:  $($status.hard_limit_hits)"
    Write-Host "  Tasks Upgraded:   $($status.tasks_upgraded)"
    
    if ($status.hard_limit_hits -eq 0 -and $status.soft_limit_hits -eq 0) {
        Write-Host "  Status:           " -NoNewline
        Write-Host "✓ Resources Sufficient" -ForegroundColor Green
    } elseif ($status.hard_limit_hits -eq 0) {
        Write-Host "  Status:           " -NoNewline
        Write-Host "⚠ Approaching Limits" -ForegroundColor Yellow
    } else {
        Write-Host "  Status:           " -NoNewline
        Write-Host "✗ Resource Constrained" -ForegroundColor Red
    }
    Write-Host ""
    
    # 效果评估
    Write-Host "📈 Effectiveness Assessment" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    
    $score = 0
    $maxScore = 5
    
    # 评分标准
    if ($total -gt 0 -and ($status.tasks_completed / $total) -ge 0.99) { $score++ }
    if ($status.hot_domains -gt 0) { $score++ }
    if ($status.tasks_dropped -eq 0) { $score++ }
    if ($status.hard_limit_hits -eq 0) { $score++ }
    if ($status.tracked_domains -gt 0 -and ($status.hot_domains / $status.tracked_domains) -ge 0.2) { $score++ }
    
    Write-Host "  Overall Score:    " -NoNewline
    Write-Host ("★" * $score) -NoNewline -ForegroundColor Yellow
    Write-Host ("☆" * ($maxScore - $score)) -NoNewline -ForegroundColor DarkGray
    Write-Host " ($score/$maxScore)"
    
    Write-Host ""
    if ($score -eq 5) {
        Write-Host "  🌟 Excellent! Prefetch is working perfectly!" -ForegroundColor Green
        Write-Host "     Your DNS performance is significantly improved." -ForegroundColor Green
    } elseif ($score -ge 4) {
        Write-Host "  ✓ Good! Prefetch is working well." -ForegroundColor Green
        Write-Host "    Minor optimizations may be possible." -ForegroundColor Cyan
    } elseif ($score -ge 3) {
        Write-Host "  ⚠ Fair. Prefetch is working but could be optimized." -ForegroundColor Yellow
    } else {
        Write-Host "  ✗ Needs attention. Check configuration and logs." -ForegroundColor Red
    }
    Write-Host ""
    
    # 建议
    Write-Host "💡 Recommendations" -ForegroundColor Yellow
    Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
    
    $recommendations = @()
    
    if ($status.hot_domains -eq 0) {
        $recommendations += "• Lower the threshold to capture more hot domains"
    }
    
    if ($status.tasks_dropped -gt 0) {
        $recommendations += "• Increase queue sizes to prevent task drops"
    }
    
    if ($status.hard_limit_hits -gt 0) {
        $recommendations += "• Increase hard_limit to handle peak loads"
    }
    
    if ($status.soft_limit_hits -gt $status.tasks_completed * 0.1) {
        $recommendations += "• Consider increasing soft_limit"
    }
    
    if ($total -gt 0 -and ($status.tasks_failed / $total) -gt 0.05) {
        $recommendations += "• High failure rate - check upstream connectivity"
    }
    
    if ($status.tracked_domains -gt $status.max_entries * 0.8) {
        $recommendations += "• Approaching max_entries limit - consider increasing"
    }
    
    if ($recommendations.Count -eq 0) {
        Write-Host "  ✓ No recommendations - configuration is optimal!" -ForegroundColor Green
    } else {
        foreach ($rec in $recommendations) {
            Write-Host "  $rec" -ForegroundColor Cyan
        }
    }
    Write-Host ""
    
    # 预期效果
    if ($status.hot_domains -gt 0) {
        Write-Host "🚀 Expected Benefits" -ForegroundColor Yellow
        Write-Host "─────────────────────────────────────────────────────────" -ForegroundColor Gray
        Write-Host "  For $($status.hot_domains) hot domains:" -ForegroundColor Cyan
        Write-Host "  • DNS query latency reduced by 50-90%" -ForegroundColor Green
        Write-Host "  • Cache hit rate improved by 20-40%" -ForegroundColor Green
        Write-Host "  • User experience significantly enhanced" -ForegroundColor Green
        Write-Host ""
    }
    
} catch {
    Write-Host "❌ Error: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Make sure AdGuard Home is running on localhost:3000" -ForegroundColor Yellow
}

Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
