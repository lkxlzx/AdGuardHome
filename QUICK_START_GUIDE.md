# Prefetch 快速上手指南

## 🚀 5 分钟快速开始

### 步骤 1：启动服务（30 秒）

双击 `start_prefetch_ui.bat` 或运行：
```bash
.\AdGuardHome_prefetch_ui.exe
```

等待看到：
```
[info] AdGuard Home is available at http://127.0.0.1:3000
```

### 步骤 2：配置 Prefetch（1 分钟）

1. 浏览器打开：http://localhost:3000
2. 登录（首次使用需创建账号）
3. 进入：**设置 → DNS 设置**
4. 找到：**DNS 缓存配置** 卡片
5. 勾选：**启用 DNS 预取**
6. 点击：**保存**

### 步骤 3：测试功能（2 分钟）

打开新的 PowerShell 窗口，运行：
```powershell
.\test_prefetch.ps1
```

等待测试完成，你会看到：
```
✓ Success! 5 domains are now marked as hot!
```

### 步骤 4：监控效果（1 分钟）

再打开一个 PowerShell 窗口，运行：
```powershell
.\monitor_prefetch.ps1
```

你会看到实时更新的监控面板：
```
╔════════════════════════════════════════════════════════════╗
║          AdGuard Home - Prefetch Status Monitor           ║
╚════════════════════════════════════════════════════════════╝

┌─ Configuration ────────────────────────────────────────┐
│ Status:           ✓ Enabled
│ Threshold:        5 hits
│ Hot Domains:      5
│ Tasks Completed:  15
└────────────────────────────────────────────────────────┘
```

### 步骤 5：验证效果（1 分钟）

使用 API 查看详细状态：
```powershell
Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status"
```

## 🎯 完成！

现在 Prefetch 已经在工作了！你可以：

✅ 继续使用网络，Prefetch 会自动跟踪热门域名  
✅ 查看监控脚本观察实时效果  
✅ 根据需要调整配置参数  

## 📊 观察效果的方法

### 方法 1：监控脚本（推荐）
```powershell
.\monitor_prefetch.ps1
```
实时显示所有指标，包括热门域名数、任务完成数等。

### 方法 2：API 查询
```powershell
# 查看完整状态
Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status"

# 只看热门域名数
(Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status").hot_domains

# 只看任务完成数
(Invoke-RestMethod -Uri "http://localhost:3000/control/prefetch_status").tasks_completed
```

### 方法 3：查看日志
在 AdGuard Home 的控制台窗口中，每分钟会输出一次指标：
```
[info] prefetch metrics hot_domains=5 tasks_completed=15 ...
```

## 💡 理解指标

| 指标 | 含义 | 期望值 |
|------|------|--------|
| hot_domains | 热门域名数 | > 0 表示有域名被预取 |
| tasks_completed | 完成的预取任务 | 持续增长表示正常工作 |
| tasks_failed | 失败的任务 | 应该很少或为 0 |
| tracked_domains | 跟踪的域名总数 | 随使用增长 |

## 🔧 常见问题

### Q1: hot_domains 一直是 0？

**原因**：域名访问次数未达到阈值（默认 5 次）

**解决**：
```powershell
# 方法 1：运行测试脚本
.\test_prefetch.ps1

# 方法 2：手动查询域名多次
nslookup google.com 127.0.0.1
nslookup google.com 127.0.0.1
nslookup google.com 127.0.0.1
nslookup google.com 127.0.0.1
nslookup google.com 127.0.0.1
```

### Q2: 监控脚本报错？

**原因**：AdGuard Home 未启动或端口不是 3000

**解决**：
1. 确认 AdGuard Home 正在运行
2. 检查是否使用了自定义端口
3. 修改脚本中的 URL

### Q3: 如何知道 Prefetch 真的在工作？

**验证方法**：
1. 运行测试脚本生成流量
2. 等待 1-2 分钟
3. 查看 `tasks_completed` 是否增加
4. 查看 `hot_domains` 是否 > 0

## 📈 性能提升

启用 Prefetch 后，你应该能观察到：

- ✅ **DNS 查询延迟降低**：热门域名的查询几乎瞬间完成
- ✅ **缓存命中率提升**：更多查询从缓存响应
- ✅ **用户体验改善**：网页加载更快

## 🎓 进阶使用

### 调整配置

根据你的网络环境调整参数：

**家庭网络（默认）**：
- 阈值：5
- 时间窗口：3600 秒

**小型企业**：
- 阈值：3（更积极预取）
- 时间窗口：7200 秒（更长的统计周期）

**大型企业**：
- 阈值：10（只预取真正热门的）
- 最大跟踪域名数：50000

### 自动化监控

创建定时任务，每小时记录一次状态：
```powershell
# 添加到 Windows 任务计划程序
$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-File C:\path\to\monitor_prefetch.ps1"
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Hours 1)
Register-ScheduledTask -TaskName "Prefetch Monitor" -Action $action -Trigger $trigger
```

## 📚 更多资源

- [PREFETCH_API_GUIDE.md](./PREFETCH_API_GUIDE.md) - 完整的 API 文档
- [PREFETCH_CONFIG_GUIDE.md](./PREFETCH_CONFIG_GUIDE.md) - 详细配置说明
- [README_PREFETCH_UI.md](./README_PREFETCH_UI.md) - 功能概述

## 💬 需要帮助？

1. 查看日志文件
2. 运行监控脚本检查状态
3. 使用 API 获取详细信息
4. 参考文档进行故障排查

---

**祝使用愉快！** 🎉

如果 Prefetch 帮助改善了你的 DNS 性能，别忘了调整配置以获得最佳效果！
