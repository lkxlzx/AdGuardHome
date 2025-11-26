# Prefetch API 使用指南

## API 端点

### GET /control/prefetch_status

获取 Prefetch 的实时状态和运行指标。

## 请求示例

### 使用 curl

```bash
curl http://localhost:3000/control/prefetch_status
```

### 使用 PowerShell

```powershell
Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status" -Method Get
```

### 使用浏览器

直接访问：
```
http://localhost:3000/control/prefetch_status
```

## 响应格式

```json
{
  "enabled": true,
  "threshold": 5,
  "time_window": 3600,
  "max_entries": 10000,
  "cleanup_interval": 3600,
  "soft_limit": 50,
  "hard_limit": 150,
  "current_active": 5,
  "urgent_queue": 2,
  "normal_queue": 10,
  "soft_limit_hits": 15,
  "hard_limit_hits": 0,
  "tasks_upgraded": 3,
  "tasks_dropped": 0,
  "tasks_completed": 150,
  "tasks_failed": 2,
  "tracked_hits": 500,
  "hot_domains": 50,
  "tracked_domains": 500
}
```

## 字段说明

### 配置信息

| 字段 | 类型 | 说明 |
|------|------|------|
| enabled | boolean | Prefetch 是否启用 |
| threshold | int | 访问阈值（域名被视为热门的最小访问次数） |
| time_window | int | 时间窗口（秒） |
| max_entries | int | 最大跟踪域名数 |
| cleanup_interval | int | 清理间隔（秒） |
| soft_limit | int | 软并发限制 |
| hard_limit | int | 硬并发限制 |

### 运行指标

| 字段 | 类型 | 说明 |
|------|------|------|
| current_active | int64 | 当前正在执行的刷新任务数 |
| urgent_queue | int64 | 紧急队列中的任务数 |
| normal_queue | int64 | 普通队列中的任务数 |
| soft_limit_hits | int64 | 触发软限制的次数（累计） |
| hard_limit_hits | int64 | 触发硬限制的次数（累计） |
| tasks_upgraded | int64 | 任务从普通升级到紧急的次数（累计） |
| tasks_dropped | int64 | 因队列满而丢弃的任务数（累计） |
| tasks_completed | int64 | 成功完成的任务数（累计） |
| tasks_failed | int64 | 失败的任务数（累计） |
| tracked_hits | int64 | 当前跟踪的域名数（有访问记录） |
| hot_domains | int64 | 当前热门域名数（会被预取） |
| tracked_domains | int64 | 当前跟踪的域名总数 |

## 使用场景

### 1. 监控 Prefetch 运行状态

创建一个监控脚本，定期查询状态：

```bash
#!/bin/bash
while true; do
    echo "=== $(date) ==="
    curl -s http://localhost:3000/control/prefetch_status | jq '.'
    echo ""
    sleep 10
done
```

### 2. 检查性能指标

```bash
# 查看任务完成率
curl -s http://localhost:3000/control/prefetch_status | jq '{
  completed: .tasks_completed,
  failed: .tasks_failed,
  dropped: .tasks_dropped,
  success_rate: (.tasks_completed / (.tasks_completed + .tasks_failed) * 100)
}'
```

### 3. 监控队列状态

```bash
# 查看队列积压情况
curl -s http://localhost:3000/control/prefetch_status | jq '{
  urgent_queue: .urgent_queue,
  normal_queue: .normal_queue,
  active_tasks: .current_active,
  total_pending: (.urgent_queue + .normal_queue)
}'
```

### 4. 检查热门域名

```bash
# 查看热门域名统计
curl -s http://localhost:3000/control/prefetch_status | jq '{
  hot_domains: .hot_domains,
  tracked_domains: .tracked_domains,
  hot_ratio: (.hot_domains / .tracked_domains * 100)
}'
```

## PowerShell 监控脚本

创建 `monitor_prefetch.ps1`：

```powershell
# Prefetch 监控脚本
$url = "http://localhost:3000/control/prefetch_status"

while ($true) {
    Clear-Host
    Write-Host "=== Prefetch Status Monitor ===" -ForegroundColor Cyan
    Write-Host "Time: $(Get-Date)" -ForegroundColor Gray
    Write-Host ""
    
    try {
        $status = Invoke-RestMethod -Uri $url -Method Get
        
        Write-Host "Configuration:" -ForegroundColor Yellow
        Write-Host "  Enabled: $($status.enabled)"
        Write-Host "  Threshold: $($status.threshold)"
        Write-Host "  Time Window: $($status.time_window)s"
        Write-Host "  Max Entries: $($status.max_entries)"
        Write-Host ""
        
        Write-Host "Runtime Metrics:" -ForegroundColor Yellow
        Write-Host "  Active Tasks: $($status.current_active)"
        Write-Host "  Urgent Queue: $($status.urgent_queue)"
        Write-Host "  Normal Queue: $($status.normal_queue)"
        Write-Host ""
        
        Write-Host "Statistics:" -ForegroundColor Yellow
        Write-Host "  Completed: $($status.tasks_completed)"
        Write-Host "  Failed: $($status.tasks_failed)"
        Write-Host "  Dropped: $($status.tasks_dropped)"
        Write-Host "  Hot Domains: $($status.hot_domains)"
        Write-Host "  Tracked Domains: $($status.tracked_domains)"
        Write-Host ""
        
        # 计算成功率
        $total = $status.tasks_completed + $status.tasks_failed
        if ($total -gt 0) {
            $successRate = [math]::Round(($status.tasks_completed / $total) * 100, 2)
            Write-Host "  Success Rate: $successRate%" -ForegroundColor Green
        }
        
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Gray
    Start-Sleep -Seconds 5
}
```

运行：
```powershell
.\monitor_prefetch.ps1
```

## Python 监控脚本

创建 `monitor_prefetch.py`：

```python
#!/usr/bin/env python3
import requests
import time
import json
from datetime import datetime

def get_prefetch_status():
    """获取 Prefetch 状态"""
    try:
        response = requests.get('http://localhost:3000/control/prefetch_status')
        response.raise_for_status()
        return response.json()
    except Exception as e:
        print(f"Error: {e}")
        return None

def display_status(status):
    """显示状态信息"""
    print("\n" + "="*50)
    print(f"Prefetch Status - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("="*50)
    
    print("\n📋 Configuration:")
    print(f"  Enabled: {status['enabled']}")
    print(f"  Threshold: {status['threshold']}")
    print(f"  Time Window: {status['time_window']}s")
    print(f"  Max Entries: {status['max_entries']}")
    
    print("\n⚡ Runtime Metrics:")
    print(f"  Active Tasks: {status['current_active']}")
    print(f"  Urgent Queue: {status['urgent_queue']}")
    print(f"  Normal Queue: {status['normal_queue']}")
    
    print("\n📊 Statistics:")
    print(f"  Completed: {status['tasks_completed']}")
    print(f"  Failed: {status['tasks_failed']}")
    print(f"  Dropped: {status['tasks_dropped']}")
    print(f"  Hot Domains: {status['hot_domains']}")
    print(f"  Tracked Domains: {status['tracked_domains']}")
    
    # 计算成功率
    total = status['tasks_completed'] + status['tasks_failed']
    if total > 0:
        success_rate = (status['tasks_completed'] / total) * 100
        print(f"\n✅ Success Rate: {success_rate:.2f}%")

def main():
    """主函数"""
    print("Starting Prefetch Monitor...")
    print("Press Ctrl+C to stop\n")
    
    try:
        while True:
            status = get_prefetch_status()
            if status:
                display_status(status)
            time.sleep(5)
    except KeyboardInterrupt:
        print("\n\nMonitor stopped.")

if __name__ == "__main__":
    main()
```

运行：
```bash
python monitor_prefetch.py
```

## 性能分析

### 检查是否需要调整配置

```bash
# 获取状态
STATUS=$(curl -s http://localhost:3000/control/prefetch_status)

# 检查队列积压
URGENT=$(echo $STATUS | jq -r '.urgent_queue')
NORMAL=$(echo $STATUS | jq -r '.normal_queue')

if [ $URGENT -gt 100 ] || [ $NORMAL -gt 500 ]; then
    echo "⚠️  Queue backlog detected! Consider increasing soft_limit."
fi

# 检查任务丢弃
DROPPED=$(echo $STATUS | jq -r '.tasks_dropped')
if [ $DROPPED -gt 0 ]; then
    echo "⚠️  Tasks are being dropped! Consider increasing queue sizes."
fi

# 检查失败率
COMPLETED=$(echo $STATUS | jq -r '.tasks_completed')
FAILED=$(echo $STATUS | jq -r '.tasks_failed')
TOTAL=$((COMPLETED + FAILED))

if [ $TOTAL -gt 0 ]; then
    FAIL_RATE=$(echo "scale=2; $FAILED * 100 / $TOTAL" | bc)
    if (( $(echo "$FAIL_RATE > 5" | bc -l) )); then
        echo "⚠️  High failure rate: ${FAIL_RATE}%"
    fi
fi
```

## 集成到监控系统

### Prometheus 格式

可以创建一个简单的导出器：

```python
from prometheus_client import start_http_server, Gauge
import requests
import time

# 定义指标
prefetch_enabled = Gauge('prefetch_enabled', 'Prefetch enabled status')
prefetch_active_tasks = Gauge('prefetch_active_tasks', 'Current active tasks')
prefetch_completed = Gauge('prefetch_tasks_completed', 'Total completed tasks')
prefetch_failed = Gauge('prefetch_tasks_failed', 'Total failed tasks')
prefetch_hot_domains = Gauge('prefetch_hot_domains', 'Number of hot domains')

def collect_metrics():
    """收集指标"""
    try:
        response = requests.get('http://localhost:3000/control/prefetch_status')
        data = response.json()
        
        prefetch_enabled.set(1 if data['enabled'] else 0)
        prefetch_active_tasks.set(data['current_active'])
        prefetch_completed.set(data['tasks_completed'])
        prefetch_failed.set(data['tasks_failed'])
        prefetch_hot_domains.set(data['hot_domains'])
    except Exception as e:
        print(f"Error: {e}")

if __name__ == '__main__':
    start_http_server(8000)
    while True:
        collect_metrics()
        time.sleep(15)
```

## 故障排查

### 问题 1：API 返回 404

**原因**：使用的是旧版本  
**解决**：确保使用最新编译的 `AdGuardHome_prefetch_ui.exe`

### 问题 2：所有指标都是 0

**原因**：Prefetch 未启用  
**解决**：在 Web UI 中启用 Prefetch

### 问题 3：hot_domains 一直是 0

**原因**：没有域名达到访问阈值  
**解决**：
1. 降低 threshold 值
2. 多次查询同一域名
3. 等待一段时间积累访问

## 总结

通过 `/control/prefetch_status` API，你可以：

✅ 实时监控 Prefetch 运行状态  
✅ 分析性能指标和效果  
✅ 及时发现配置问题  
✅ 集成到监控系统  
✅ 自动化运维管理  

建议定期查看这些指标，根据实际情况调整配置参数以获得最佳性能。
