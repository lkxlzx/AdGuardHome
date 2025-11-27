# 简单的LRU测试 - 手动观察
Write-Host "=== LRU增量清理测试 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "此脚本将启动AdGuard Home，您需要手动测试并观察日志" -ForegroundColor Yellow
Write-Host ""

# 启动服务
Write-Host "启动 AdGuard Home..." -ForegroundColor Green
Start-Process -FilePath ".\AdGuardHome_lru_optimized.exe" -ArgumentList "-c", "AdGuardHome.yaml", "--no-check-update"

Write-Host ""
Write-Host "服务已启动！" -ForegroundColor Green
Write-Host ""
Write-Host "测试步骤:" -ForegroundColor Yellow
Write-Host "1. 运行以下命令发送大量查询（触发清理）:"
Write-Host "   for (\$i=1; \$i -le 12000; \$i++) { Resolve-DnsName -Name `"test-\$i.example.com`" -Server 127.0.0.1 -DnsOnly -ErrorAction SilentlyContinue }" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. 观察日志输出，查找以下消息:"
Write-Host "   - 'incremental cleanup completed' (增量清理完成)"
Write-Host "   - 'shards_cleaned' 应该是 4"
Write-Host "   - 'performing full cleanup after incremental rounds' (每4轮后的完整清理)"
Write-Host ""
Write-Host "3. 按 Ctrl+C 停止服务"
Write-Host ""
