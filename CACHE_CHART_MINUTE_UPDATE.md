# 缓存图表分钟级别显示更新

## 更新日期
2025-11-26

## 更新内容

### 问题描述
之前的缓存命中率图表使用小时级别的时间粒度，无法提供足够细致的实时监控。

### 解决方案
将图表时间粒度从**每小时**调整为**每分钟**，提供更精细的实时监控能力。

---

## 技术实现

### 后端实现（已完成）

#### DNSCacheStats 结构
```go
type DNSCacheStats struct {
    mu           sync.RWMutex
    totalQueries int64
    cacheHits    int64
    cacheMisses  int64
    history      [60]float64  // 60 分钟历史数据
    historyIndex int
    lastUpdate   time.Time
}
```

**关键特性：**
- 维护 60 个数据点（60 分钟）
- 每分钟更新一次历史数据
- 使用环形缓冲区存储数据
- 按时间顺序返回历史数据

#### 更新逻辑
```go
func (cs *DNSCacheStats) RecordQuery(hit bool) {
    // 每次 DNS 查询都记录
    cs.totalQueries++
    if hit {
        cs.cacheHits++
    } else {
        cs.cacheMisses++
    }
    
    // 每分钟更新一次历史数据
    now := time.Now()
    if cs.lastUpdate.IsZero() || now.Sub(cs.lastUpdate) >= time.Minute {
        cs.updateHistory()
        cs.lastUpdate = now
    }
}
```

### 前端实现（本次更新）

#### 修改文件
`client/src/components/Dashboard/CacheMetrics.tsx`

#### 关键改进

##### 1. 直接使用 ResponsiveLine 组件
不再依赖通用的 `Line` 组件，而是直接使用 `@nivo/line` 的 `ResponsiveLine`，以便完全控制时间格式化。

##### 2. 分钟级别时间格式化
```typescript
xFormat={(x: number) => {
    // x 是分钟索引 (0-59)
    // 计算实际时间: 当前时间 - (59 - x) 分钟
    const minutesAgo = 59 - x;
    const time = subMinutes(Date.now(), minutesAgo);
    return dateFormat(time, 'HH:mm');  // 格式: 14:30
}}
```

**时间计算逻辑：**
- 数据点 0 = 60 分钟前
- 数据点 29 = 30 分钟前
- 数据点 59 = 当前时间

##### 3. Y 轴格式化
```typescript
yFormat={(y: number) => `${round(y, 1)}%`}
```
显示格式：`95.5%`

##### 4. 图表配置
```typescript
xScale={{
    type: 'linear',
    min: 0,
    max: 59,  // 60 个数据点 (0-59)
}}
yScale={{
    type: 'linear',
    min: 0,
    max: 100,  // 百分比范围
}}
```

##### 5. 工具提示
鼠标悬停时显示：
- **命中率**：95.5%（粗体）
- **时间**：14:30（小字）

---

## 用户体验改进

### 之前（小时级别）
- **时间范围**：24 小时
- **数据点数**：24 个
- **时间粒度**：1 小时/点
- **时间格式**：`D MMM HH:00`（如 "26 Nov 14:00"）
- **实时性**：低（1 小时才更新一次）

### 现在（分钟级别）
- **时间范围**：60 分钟
- **数据点数**：60 个
- **时间粒度**：1 分钟/点
- **时间格式**：`HH:mm`（如 "14:30"）
- **实时性**：高（每分钟更新一次）

---

## 性能影响

### 内存使用
- **之前**：24 个 float64 = 192 字节
- **现在**：60 个 float64 = 480 字节
- **增加**：288 字节（可忽略）

### CPU 使用
- **历史更新频率**：每分钟一次（与之前相同的更新机制）
- **前端刷新频率**：每 30 秒一次
- **影响**：极小

---

## 测试验证

### 1. 启动服务
```bash
./AdGuardHome_cache_minutes.exe
```

### 2. 运行测试脚本
```powershell
.\test_cache_minutes.ps1
```

### 3. 浏览器验证
1. 打开 `http://localhost:3000`
2. 进入 Dashboard 页面
3. 查看 "DNS 缓存命中率" 卡片
4. 鼠标悬停在图表上
5. 验证时间格式为 `HH:mm`
6. 验证时间范围为最近 60 分钟

### 4. 预期结果
- ✅ 图表显示 60 个数据点
- ✅ 时间标签格式为 `14:30`
- ✅ 从左到右显示从 60 分钟前到现在
- ✅ 鼠标悬停显示精确的时间和命中率
- ✅ 每 30 秒自动刷新数据

---

## API 响应示例

### GET /control/cache_metrics

```json
{
  "cache_enabled": true,
  "cache_hit_rate": 85.5,
  "cache_size": 4194304,
  "total_queries": 1234,
  "cache_hits": 1055,
  "cache_misses": 179,
  "history": [
    82.1, 83.5, 84.2, 85.0, 85.5, 86.1, 87.0, 86.5,
    85.8, 85.2, 84.9, 85.3, 85.7, 86.0, 85.5, 85.1,
    // ... 共 60 个数据点
  ],
  "next_update_in": 45
}
```

**字段说明：**
- `history`: 60 个数据点，按时间顺序排列（从 60 分钟前到现在）
- `next_update_in`: 距离下次历史更新的秒数

---

## 代码变更总结

### 修改的文件
1. `client/src/components/Dashboard/CacheMetrics.tsx` - 前端图表组件

### 新增的文件
1. `test_cache_minutes.ps1` - 测试脚本
2. `CACHE_CHART_MINUTE_UPDATE.md` - 本文档

### 编译的文件
1. `AdGuardHome_cache_minutes.exe` - 新的可执行文件

---

## 兼容性

### 后端兼容性
- ✅ 不影响现有 API
- ✅ 不影响其他功能
- ✅ 向后兼容

### 前端兼容性
- ✅ 不影响其他图表组件
- ✅ 不影响其他 Dashboard 卡片
- ✅ 响应式设计保持不变

---

## 未来改进建议

### 1. 可配置时间范围
允许用户选择查看：
- 最近 30 分钟
- 最近 60 分钟（当前）
- 最近 2 小时
- 最近 24 小时

### 2. 数据导出
添加导出功能，允许用户下载历史数据为 CSV 格式。

### 3. 对比视图
显示与上一小时的对比，帮助识别趋势。

### 4. 告警阈值
当命中率低于设定阈值时显示警告。

---

## 相关文档

- [P0_FIXES_COMPLETE.md](./P0_FIXES_COMPLETE.md) - P0 问题修复
- [DASHBOARD_METRICS_UI_DESIGN.md](./DASHBOARD_METRICS_UI_DESIGN.md) - Dashboard 设计
- [FRONTEND_TASKS_COMPLETE.md](./FRONTEND_TASKS_COMPLETE.md) - 前端任务总结

---

**更新完成时间**: 2025-11-26  
**编译状态**: ✅ 成功  
**测试状态**: ⏳ 待测试  
**可执行文件**: `AdGuardHome_cache_minutes.exe`
