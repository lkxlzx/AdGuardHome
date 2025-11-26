# Dashboard V2 设计文档

## 概述

重新设计了 DNS 缓存命中率和 Prefetch 状态两个卡片，完全符合 AdGuard Home 原有卡片风格，并在 Dashboard 顶部横向排列。

## 设计特点

### 1. DNS 缓存命中率卡片

**样式**: 采用波形图样式（类似 Statistics 卡片）

**特性**:
- 使用 `StatsCard` 相同的布局和样式
- 显示缓存命中率百分比（大号数字）
- 显示命中数/总查询数（小号文字）
- 使用 Nivo Line 图表显示 24 小时历史趋势
- 蓝色主题（与 DNS Query 卡片一致）
- 当缓存禁用时显示灰色状态

**数据来源**: `/control/cache_metrics` API

**响应格式**:
```json
{
  "cache_enabled": true,
  "cache_hit_rate": 75.0,
  "cache_size": 4194304,
  "total_queries": 1000,
  "cache_hits": 750,
  "cache_misses": 250,
  "history": [73.5, 74.2, 75.0, ...]  // 24小时数据
}
```

### 2. Prefetch 状态卡片

**样式**: 信息展示卡片（类似 Counters 卡片）

**特性**:
- 顶部显示状态图标和状态徽章（运行中/空闲）
- 中间显示三个关键指标（大号数字）:
  - 成功率（绿色高亮）
  - 热门域名数量
  - 队列大小
- 底部显示详细统计:
  - 已完成数量（绿色）
  - 已失败数量（红色）
  - 最后预取时间（灰色）
- 当 Prefetch 禁用时显示简化状态

**数据来源**: `/control/prefetch_metrics` API

**响应格式**:
```json
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

## 布局设计

### Dashboard 布局结构

```
┌─────────────────────────────────────────────────────────────┐
│  Page Title & Controls                                       │
└─────────────────────────────────────────────────────────────┘

┌──────────────────────────┬──────────────────────────────────┐
│  DNS 缓存命中率           │  Prefetch 状态                    │
│  (波形图卡片)             │  (信息卡片)                       │
│  col-sm-6 col-lg-3       │  col-sm-6 col-lg-3               │
└──────────────────────────┴──────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Performance Monitoring (原有组件)                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Statistics (4个统计卡片)                                     │
└─────────────────────────────────────────────────────────────┘

┌──────────────────────────┬──────────────────────────────────┐
│  General Statistics      │  Clients                          │
└──────────────────────────┴──────────────────────────────────┘

... (其他原有卡片)
```

## 技术实现

### 前端组件

1. **CacheMetrics.tsx**
   - 使用 `Card` 组件（type="card--full"）
   - 使用 `Line` 组件显示波形图
   - 每 30 秒自动刷新数据
   - 完全匹配 `StatsCard` 的样式

2. **PrefetchMetrics.tsx**
   - 使用 `Card` 组件（type="card--full"）
   - 自定义布局显示多层信息
   - 每 30 秒自动刷新数据
   - 使用 Badge 组件显示状态

3. **MetricsCards.css**
   - 定义 Prefetch 卡片的自定义样式
   - 响应式设计支持移动端
   - 保持与原有样式一致的间距和颜色

### 后端 API

1. **GET /control/cache_metrics**
   - 返回缓存命中率和历史数据
   - 包含缓存配置信息
   - 支持缓存禁用状态

2. **GET /control/prefetch_metrics**
   - 返回 Prefetch 运行状态
   - 包含详细的统计数据
   - 支持 Prefetch 禁用状态

### 国际化

添加了以下翻译键：

**英文 (en.json)**:
- `dns_cache_hit_rate`: "DNS Cache Hit Rate"
- `hot_domains`: "Hot Domains"
- `queue_size`: "Queue Size"
- `completed`: "Completed"
- `failed`: "Failed"
- `last_prefetch`: "Last Prefetch"
- `normal`: "Normal"

**中文 (zh-cn.json)**:
- `dns_cache_hit_rate`: "DNS 缓存命中率"
- `hot_domains`: "热门域名"
- `queue_size`: "队列大小"
- `completed`: "已完成"
- `failed`: "已失败"
- `last_prefetch`: "最后预取"
- `normal`: "正常"

## 样式特点

### 颜色方案

- **DNS 缓存命中率**: 蓝色主题 (`text-blue`, `STATUS_COLORS.blue`)
- **Prefetch 状态**: 绿色主题 (`icon--green`, `text-success`)
- **成功指标**: 绿色 (`text-success`, `badge-success`)
- **失败指标**: 红色 (`text-danger`)
- **禁用状态**: 灰色 (`text-gray`, `badge-secondary`)

### 响应式设计

- **桌面 (≥992px)**: 两个卡片各占 3 列（col-lg-3），横向排列
- **平板 (≥576px)**: 两个卡片各占 6 列（col-sm-6），横向排列
- **移动 (<576px)**: 两个卡片各占 12 列，纵向堆叠

## 测试

使用 `test_dashboard_v2.ps1` 脚本测试 API 端点：

```powershell
.\test_dashboard_v2.ps1
```

测试内容：
1. `/control/cache_metrics` - 缓存指标
2. `/control/prefetch_metrics` - Prefetch 指标
3. `/control/dashboard_metrics` - 原有仪表板指标

## 文件清单

### 新增文件
- `client/src/components/Dashboard/MetricsCards.css` - 卡片样式
- `test_dashboard_v2.ps1` - API 测试脚本
- `DASHBOARD_V2_DESIGN.md` - 本设计文档

### 修改文件
- `client/src/components/Dashboard/index.tsx` - 添加新卡片
- `client/src/components/Dashboard/CacheMetrics.tsx` - 重新设计
- `client/src/components/Dashboard/PrefetchMetrics.tsx` - 重新设计
- `client/src/__locales/en.json` - 添加英文翻译
- `client/src/__locales/zh-cn.json` - 添加中文翻译
- `internal/dnsforward/http.go` - 添加新 API 端点

## 下一步

1. 启动服务器测试新卡片显示效果
2. 验证响应式布局在不同屏幕尺寸下的表现
3. 实现真实的缓存统计数据收集（当前使用占位数据）
4. 添加更多交互功能（如点击查看详情）
5. 优化性能和数据刷新策略

## 注意事项

- 当前缓存命中率使用模拟数据，需要实现真实的统计收集
- Prefetch 最后预取时间需要在 Prefetch 模块中记录
- 历史数据需要持久化存储以支持服务器重启后的数据恢复
- 考虑添加数据导出功能用于性能分析
