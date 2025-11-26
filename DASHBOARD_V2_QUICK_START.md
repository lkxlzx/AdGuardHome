# Dashboard V2 快速启动指南

## 快速测试新卡片

### 1. 启动服务器

```powershell
# 使用新编译的版本
.\AdGuardHome_dashboard_v2.exe
```

### 2. 测试 API 端点

在另一个终端窗口运行：

```powershell
.\test_dashboard_v2.ps1
```

预期输出：
```
=== Testing Dashboard V2 API Endpoints ===

1. Testing /control/cache_metrics endpoint...
✓ Cache Metrics Response:
{
  "cache_enabled": true,
  "cache_hit_rate": 75.0,
  "cache_size": 4194304,
  "total_queries": 1000,
  "cache_hits": 750,
  "cache_misses": 250,
  "history": [73.5, 74.2, 75.0, ...]
}

2. Testing /control/prefetch_metrics endpoint...
✓ Prefetch Metrics Response:
{
  "prefetch_enabled": true,
  "prefetch_status": "active",
  "prefetch_hot_domains": 150,
  "prefetch_completed": 5420,
  "prefetch_failed": 28,
  "prefetch_success_rate": 99.5,
  "prefetch_queue_size": 12,
  "last_prefetch_time": "2025-11-26T10:30:00Z"
}
```

### 3. 查看 Web UI

1. 打开浏览器访问: http://localhost:3000
2. 登录到 AdGuard Home
3. 进入 Dashboard 页面
4. 在页面顶部应该看到两个新卡片：
   - **DNS 缓存命中率** - 左侧，蓝色波形图
   - **Prefetch 状态** - 右侧，绿色信息卡片

## 卡片显示效果

### DNS 缓存命中率卡片

```
┌─────────────────────────┐
│  75.0%                  │  ← 大号蓝色数字
│  DNS Cache Hit Rate     │  ← 标题
│  750 / 1000            │  ← 命中数/总数
│  ╱╲╱╲╱╲╱╲╱╲           │  ← 波形图
└─────────────────────────┘
```

### Prefetch 状态卡片

```
┌─────────────────────────┐
│  🔄 Prefetch 状态  [运行中] │  ← 标题和状态
│                         │
│  99.5%    150    12    │  ← 三个关键指标
│  成功率  热门域名 队列大小 │
│                         │
│  已完成: 5420          │  ← 详细统计
│  已失败: 28            │
│  最后预取: 10:30:00    │
└─────────────────────────┘
```

## 验证清单

- [ ] 两个卡片在 Dashboard 顶部横向排列
- [ ] DNS 缓存命中率显示蓝色波形图
- [ ] Prefetch 状态显示完整信息
- [ ] 卡片样式与原有卡片一致
- [ ] 数据每 30 秒自动刷新
- [ ] 响应式布局在移动端正常显示
- [ ] 缓存禁用时显示正确状态
- [ ] Prefetch 禁用时显示正确状态
- [ ] 中英文翻译正确显示

## 常见问题

### Q: 卡片不显示？
A: 检查浏览器控制台是否有 API 错误，确认服务器正在运行。

### Q: 数据显示为 0 或 --？
A: 这是正常的，因为当前使用模拟数据。实际使用时会显示真实统计。

### Q: 波形图不显示？
A: 确认 Nivo 图表库已正确加载，检查浏览器控制台错误。

### Q: 样式不正确？
A: 清除浏览器缓存，确认 MetricsCards.css 已加载。

## 下一步开发

1. **实现真实缓存统计**
   - 在 dnsProxy 中添加缓存命中/未命中计数器
   - 持久化历史数据

2. **优化 Prefetch 数据**
   - 在 Prefetch 模块中记录最后预取时间
   - 添加更多运行时指标

3. **增强交互**
   - 点击卡片查看详细信息
   - 添加时间范围选择器
   - 支持数据导出

4. **性能优化**
   - 实现增量数据更新
   - 优化图表渲染性能
   - 添加数据缓存

## 编译前端（如需修改）

如果修改了前端代码，需要重新编译：

```bash
cd client
npm install
npm run build
cd ..
```

然后重新编译后端：

```powershell
go build -o AdGuardHome_dashboard_v2.exe
```

## 反馈

如有问题或建议，请查看 `DASHBOARD_V2_DESIGN.md` 了解详细设计。
