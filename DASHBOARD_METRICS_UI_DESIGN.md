# Dashboard Metrics UI 实现文档

## 概述

成功在 AdGuard Home 的 Dashboard 页面添加了两个新的性能监控卡片：
1. **Cache Metrics（缓存指标）** - 显示 DNS 缓存的配置和状态
2. **Prefetch Metrics（预取指标）** - 显示 DNS 预取的运行指标

## 实现内容

### 1. 后端 API（已完成）

#### API 端点：`GET /control/dashboard_metrics`

**响应数据结构：**
```json
{
  "cache_enabled": true,
  "cache_size": 4194304,
  "cache_ttl_min": 0,
  "cache_ttl_max": 0,
  "cache_optimistic": false,
  "prefetch_enabled": true,
  "prefetch_hot_domains": 150,
  "prefetch_completed": 1234,
  "prefetch_failed": 56,
  "prefetch_success_rate": 95.7
}
```

**实现位置：** `internal/dnsforward/http.go`
- 函数：`handleGetDashboardMetrics()`
- 路由注册：`s.conf.HTTPReg.Register(http.MethodGet, "/control/dashboard_metrics", s.handleGetDashboardMetrics)`

### 2. 前端组件

#### 2.1 CacheMetrics 组件

**文件：** `client/src/components/Dashboard/CacheMetrics.tsx`

**功能：**
- 从 `/control/dashboard_metrics` API 获取缓存指标
- 显示缓存配置信息：
  - 缓存大小（自动格式化为 B/KB/MB/GB）
  - 最小 TTL
  - 最大 TTL
  - 乐观缓存状态
- 当缓存禁用时显示提示信息

**显示效果：**
```
┌─────────────────────────────────┐
│ Cache Metrics            [刷新] │
├─────────────────────────────────┤
│ Cache size        4 MB      ⓘ  │
│ Min TTL           0s        ⓘ  │
│ Max TTL           0s        ⓘ  │
│ Optimistic cache  Disabled  ⓘ  │
└─────────────────────────────────┘
```

#### 2.2 PrefetchMetrics 组件

**文件：** `client/src/components/Dashboard/PrefetchMetrics.tsx`

**功能：**
- 从 `/control/dashboard_metrics` API 获取预取指标
- 显示预取运行指标：
  - 热门域名数量
  - 完成的预取操作数
  - 失败的预取操作数
  - 成功率（百分比）
- 当预取禁用时显示提示信息

**显示效果：**
```
┌─────────────────────────────────┐
│ Prefetch Metrics         [刷新] │
├─────────────────────────────────┤
│ Hot domains       150       ⓘ  │
│ Completed         1,234     ⓘ  │
│ Failed            56        ⓘ  │
│ Success rate      95.7%     ⓘ  │
└─────────────────────────────────┘
```

#### 2.3 Dashboard 主页面更新

**文件：** `client/src/components/Dashboard/index.tsx`

**修改内容：**
1. 导入新组件：
   ```typescript
   import CacheMetrics from './CacheMetrics';
   import PrefetchMetrics from './PrefetchMetrics';
   ```

2. 在 Dashboard 布局中添加两个新卡片：
   ```tsx
   <div className="col-lg-6">
       <CacheMetrics refreshButton={refreshButton} />
   </div>

   <div className="col-lg-6">
       <PrefetchMetrics refreshButton={refreshButton} />
   </div>
   ```

### 3. 国际化翻译

#### 3.1 英文翻译（en.json）

添加了 16 个新的翻译键：

```json
{
  "cache_metrics": "Cache Metrics",
  "cache_disabled": "DNS cache is disabled",
  "cache_size_hint": "Total size of DNS cache in bytes",
  "cache_ttl_min_hint": "Minimum TTL for cached DNS responses",
  "cache_ttl_max_hint": "Maximum TTL for cached DNS responses",
  "cache_optimistic_hint": "Serve expired cache entries while refreshing",
  "prefetch_metrics": "Prefetch Metrics",
  "prefetch_disabled": "DNS prefetch is disabled",
  "prefetch_hot_domains": "Hot domains",
  "prefetch_hot_domains_hint": "Number of domains currently being prefetched",
  "prefetch_completed": "Completed",
  "prefetch_completed_hint": "Total number of successful prefetch operations",
  "prefetch_failed": "Failed",
  "prefetch_failed_hint": "Total number of failed prefetch operations",
  "prefetch_success_rate": "Success rate",
  "prefetch_success_rate_hint": "Percentage of successful prefetch operations"
}
```

#### 3.2 中文翻译（zh-cn.json）

添加了对应的 16 个中文翻译键。

## 技术特点

### 1. 组件设计

- **独立数据获取**：每个组件独立调用 API，不依赖 Redux store
- **错误处理**：包含完整的错误处理和加载状态
- **条件渲染**：功能禁用时显示友好提示
- **响应式布局**：使用 Bootstrap 的 `col-lg-6` 实现响应式

### 2. 数据格式化

- **字节格式化**：自动将字节转换为 B/KB/MB/GB
- **数字格式化**：使用 `formatNumber()` 添加千位分隔符
- **百分比格式化**：保留一位小数显示成功率

### 3. 用户体验

- **工具提示**：每个指标都有详细的说明
- **刷新按钮**：支持手动刷新数据
- **一致性**：与现有 Dashboard 卡片保持一致的设计风格

## 文件清单

### 新增文件
1. `client/src/components/Dashboard/CacheMetrics.tsx` - 缓存指标组件
2. `client/src/components/Dashboard/PrefetchMetrics.tsx` - 预取指标组件

### 修改文件
1. `client/src/components/Dashboard/index.tsx` - Dashboard 主页面
2. `client/src/__locales/en.json` - 英文翻译
3. `client/src/__locales/zh-cn.json` - 中文翻译

### 后端文件（已存在）
1. `internal/dnsforward/http.go` - API 端点实现

## 测试验证

### 编译测试
```bash
cd client
npm run build-prod
```
✅ 编译成功，无错误

### 功能测试清单

#### 1. 缓存指标卡片
- [ ] 缓存启用时正确显示所有指标
- [ ] 缓存禁用时显示提示信息
- [ ] 字节大小正确格式化（B/KB/MB/GB）
- [ ] 工具提示正确显示
- [ ] 刷新按钮正常工作

#### 2. 预取指标卡片
- [ ] 预取启用时正确显示所有指标
- [ ] 预取禁用时显示提示信息
- [ ] 数字正确格式化（千位分隔符）
- [ ] 成功率正确计算和显示
- [ ] 工具提示正确显示
- [ ] 刷新按钮正常工作

#### 3. API 测试
- [ ] GET /control/dashboard_metrics 返回正确数据
- [ ] 缓存禁用时 API 返回正确状态
- [ ] 预取禁用时 API 返回正确状态
- [ ] API 错误处理正常

#### 4. 国际化测试
- [ ] 英文界面显示正确
- [ ] 中文界面显示正确
- [ ] 所有翻译键都有对应的翻译

## 使用说明

### 1. 启动服务
```bash
# 编译前端
cd client
npm run build-prod

# 编译后端
cd ..
go build

# 启动服务
./AdGuardHome
```

### 2. 访问 Dashboard
1. 打开浏览器访问 AdGuard Home Web 界面
2. 进入 Dashboard 页面
3. 滚动到页面底部
4. 查看新增的两个卡片：
   - Cache Metrics（缓存指标）
   - Prefetch Metrics（预取指标）

### 3. 启用功能
如果卡片显示功能已禁用：
1. 进入 **设置 → DNS 设置**
2. 找到 **DNS 缓存配置** 卡片
3. 启用 **DNS 缓存** 和/或 **DNS 预取**
4. 保存配置
5. 返回 Dashboard 查看指标

## 性能考虑

1. **API 调用频率**：
   - 组件挂载时调用一次
   - 用户点击刷新按钮时调用
   - 不会自动轮询，避免不必要的请求

2. **数据缓存**：
   - 使用 React 的 useState 缓存数据
   - 避免重复渲染

3. **错误处理**：
   - API 失败时不会阻塞页面
   - 控制台输出错误信息便于调试

## 未来改进

1. **实时更新**：
   - 添加自动刷新功能（可配置间隔）
   - 使用 WebSocket 推送实时数据

2. **图表可视化**：
   - 添加成功率趋势图
   - 添加热门域名数量变化图

3. **详细视图**：
   - 点击卡片查看更详细的统计信息
   - 显示热门域名列表

4. **性能对比**：
   - 显示启用预取前后的性能对比
   - 显示平均查询延迟改善

## 相关文档

- [PREFETCH_UI_INTEGRATION.md](./PREFETCH_UI_INTEGRATION.md) - Prefetch UI 集成指南
- [PREFETCH_UI_IMPLEMENTATION_SUMMARY.md](./PREFETCH_UI_IMPLEMENTATION_SUMMARY.md) - Prefetch UI 实现总结
- [PREFETCH_CONFIG_GUIDE.md](./PREFETCH_CONFIG_GUIDE.md) - Prefetch 配置指南

## 总结

本次实现完成了 Dashboard 性能监控卡片的开发，包括：
- ✅ 2 个新的 React 组件
- ✅ 完整的国际化支持（中英文）
- ✅ 与现有 Dashboard 一致的设计风格
- ✅ 完善的错误处理和用户提示
- ✅ 前端编译通过，无语法错误

用户现在可以在 Dashboard 页面直观地查看 DNS 缓存和预取功能的运行状态和性能指标。
