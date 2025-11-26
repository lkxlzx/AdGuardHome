# Prefetch 实时监控脚本
# 用法: .\monitor_prefetch.ps1

$url = "http://localhost:3000/control/prefetch_status"
$refreshInterval = 5  # 刷新间隔（秒）

Write-Host "Starting Prefetch Monitor..." -ForegroundColor Cyan
Write-Host "Refresh interval: $refreshInterval seconds" -ForegroundColor Gray
Write-Host "Press Ctrl+C to stop" -ForegroundColor Gray
Write-Host ""

while ($true) {
    Clear-Host
    
    Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║          AdGuard Home - Prefetch Status Monitor           ║" -ForegroundColor Cyan
    Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Time: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
    Write-Host ""
    
    try {
        $status = Invoke-RestMethod -Uri $url -Method Get -ErrorAction Stop
        
        # 配置信息
        Write-Host "┌─ Configuration ────────────────────────────────────────┐" -ForegroundColor Yellow
        $enabledColor = if ($status.enabled) { "Green" } else { "Red" }
        $enabledText = if ($status.enabled) { "✓ Enabled" } else { "✗ Disabled" }
        Write-Host "│ Status:           " -NoNewline
        Write-Host $enabledText -ForegroundColor $enabledColor
        Write-Host "│ Threshold:        $($status.threshold) hits"
        Write-Host "│ Time Window:      $($status.time_window)s ($([math]::Round($status.time_window/60, 1))min)"
        Write-Host "│ Max Entries:      $($status.max_entries)"
        Write-Host "│ Cleanup Interval: $($status.cleanup_interval)s ($([math]::Round($status.cleanup_interval/60, 1))min)"
        Write-Host "│ Soft Limit:       $($status.soft_limit)"
        Write-Host "│ Hard Limit:       $($status.hard_limit)"
        Write-Host "└────────────────────────────────────────────────────────┘"
        Write-Host ""
        
        if ($status.enabled) {
            # 运行时指标
            Write-Host "┌─ Runtime Metrics ──────────────────────────────────────┐" -ForegroundColor Cyan
            Write-Host "│ Active Tasks:     $($status.current_active)"
            
            $urgentColor = if ($status.urgent_queue -gt 50) { "Red" } elseif ($status.urgent_queue -gt 20) { "Yellow" } else { "Green" }
            Write-Host "│ Urgent Queue:     " -NoNewline
            Write-Host $status.urgent_queue -ForegroundColor $urgentColor
            
            $normalColor = if ($status.normal_queue -gt 500) { "Red" } elseif ($status.normal_queue -gt 200) { "Yellow" } else { "Green" }
            Write-Host "│ Normal Queue:     " -NoNewline
            Write-Host $status.normal_queue -ForegroundColor $normalColor
            
            Write-Host "│ Soft Limit Hits:  $($status.soft_limit_hits)"
            Write-Host "│ Hard Limit Hits:  $($status.hard_limit_hits)"
            Write-Host "│ Tasks Upgraded:   $($status.tasks_upgraded)"
            Write-Host "└────────────────────────────────────────────────────────┘"
            Write-Host ""
            
            # 统计信息
            Write-Host "┌─ Statistics ───────────────────────────────────────────┐" -ForegroundColor Green
            Write-Host "│ Tasks Completed:  $($status.tasks_completed)"
            Write-Host "│ Tasks Failed:     $($status.tasks_failed)"
            Write-Host "│ Tasks Dropped:    $($status.tasks_dropped)"
            Write-Host "│ Tracked Hits:     $($status.tracked_hits)"
            Write-Host "│ Hot Domains:      $($status.hot_domains)"
            Write-Host "│ Tracked Domains:  $($status.tracked_domains)"
            
            # 计算成功率
            $total = $status.tasks_completed + $status.tasks_failed
            if ($total -gt 0) {
                $successRate = [math]::Round(($status.tasks_completed / $total) * 100, 2)
                $rateColor = if ($successRate -ge 95) { "Green" } elseif ($successRate -ge 90) { "Yellow" } else { "Red" }
                Write-Host "│ Success Rate:     " -NoNewline
                Write-Host "$successRate%" -ForegroundColor $rateColor
            }
            
            # 计算热门域名比例
            if ($status.tracked_domains -gt 0) {
                $hotRatio = [math]::Round(($status.hot_domains / $status.tracked_domains) * 100, 2)
                Write-Host "│ Hot Ratio:        $hotRatio%"
            }
            
            Write-Host "└────────────────────────────────────────────────────────┘"
            Write-Host ""
            
            # 警告信息
            $warnings = @()
            
            if ($status.urgent_queue -gt 100) {
                $warnings += "⚠️  Urgent queue backlog detected!"
            }
            
            if ($status.normal_queue -gt 500) {
                $warnings += "⚠️  Normal queue backlog detected!"
            }
            
            if ($status.tasks_dropped -gt 0) {
                $warnings += "⚠️  Tasks are being dropped! ($($status.tasks_dropped) total)"
            }
            
            if ($total -gt 0) {
                $failRate = ($status.tasks_failed / $total) * 100
                if ($failRate -gt 5) {
                    $warnings += "⚠️  High failure rate: $([math]::Round($failRate, 2))%"
                }
            }
            
            if ($status.hard_limit_hits -gt 0) {
                $warnings += "⚠️  Hard limit has been hit! ($($status.hard_limit_hits) times)"
            }
            
            if ($warnings.Count -gt 0) {
                Write-Host "┌─ Warnings ─────────────────────────────────────────────┐" -ForegroundColor Red
                foreach ($warning in $warnings) {
                    Write-Host "│ $warning" -ForegroundColor Yellow
                }
                Write-Host "└────────────────────────────────────────────────────────┘"
                Write-Host ""
            }
            
            # 建议
            if ($status.hot_domains -eq 0 -and $status.tracked_domains -gt 0) {
                Write-Host "💡 Tip: No hot domains yet. Try lowering the threshold or wait for more traffic." -ForegroundColor Cyan
                Write-Host ""
            }
        } else {
            Write-Host "ℹ️  Prefetch is disabled. Enable it in Web UI to see metrics." -ForegroundColor Yellow
            Write-Host ""
        }
        
    } catch {
        Write-Host "┌─ Error ────────────────────────────────────────────────┐" -ForegroundColor Red
        Write-Host "│ Failed to fetch status:" -ForegroundColor Red
        Write-Host "│ $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "│" 
        Write-Host "│ Make sure AdGuard Home is running on localhost:3000" -ForegroundColor Yellow
        Write-Host "└────────────────────────────────────────────────────────┘"
        Write-Host ""
    }
    
    Write-Host "Next refresh in $refreshInterval seconds... (Press Ctrl+C to stop)" -ForegroundColor Gray
    Start-Sleep -Seconds $refreshInterval
}
