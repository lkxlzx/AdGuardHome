# Prefetch 配置指南

## 概述

Prefetch（预取）功能可以自动刷新热门域名的DNS缓存，在缓存过期前主动查询，从而减少DNS查询延迟，提升用户体验。

## ⚠️ 重要前提条件

**Prefetch 功能依赖 DNS 缓存才能正常工作！**

在启用 Prefetch 之前，必须确保：

```yaml
dns:
  # 必须启用DNS缓存
  cache_enabled: true
  cache_size: 4194304  # 建议至少 4MB
```

**如果关闭缓存**：
- ❌ Prefetch 会发送刷新查询，但结果不会被缓存
- ❌ 无法减少DNS查询延迟
- ❌ 浪费网络带宽和CPU资源
- ⚠️ 系统会在启动时输出警告日志

**推荐配置**：
- 启用缓存：`cache_enabled: true`（必需）
- 合理的缓存大小：`cache_size: 4194304`（4MB，推荐）
- 乐观缓存：`cache_optimistic: true`（可选，作为兜底机制）

**关于乐观缓存**：
- Prefetch可以在关闭乐观缓存的情况下正常工作
- 乐观缓存是可选的，但建议启用作为双保险
- 如果Prefetch失败刷新，乐观缓存可以提供兜底

## 配置参数

### 在 `AdGuardHome.yaml` 中添加以下配置：

```yaml
dns:
  # ... 其他DNS配置 ...
  
  # ========== Prefetch 配置 ==========
  
  # 是否启用DNS预取功能
  # 启用后，频繁访问的域名会在缓存过期前自动刷新
  # 默认: false
  prefetch_enabled: true
  
  # 热门域名阈值
  # 域名在时间窗口内访问次数达到此值后才会被视为"热门"并进行预取
  # 范围: 1-100
  # 默认: 5
  # 建议:
  #   - 家庭用户: 3-5
  #   - 小型企业: 5-10
  #   - 大型部署: 10-20
  prefetch_threshold: 5
  
  # 热度统计时间窗口
  # 只统计此时间窗口内的访问次数
  # 格式: 时间字符串 (如 "1h", "30m", "2h", "24h")
  # 默认: "1h"
  # 说明: 例如设置为 "1h"，则只有1小时内命中5次才算热门域名
  # 建议:
  #   - 家庭用户: "1h" - 快速响应用户习惯变化
  #   - 小型企业: "2h" - 平衡响应速度和稳定性
  #   - 大型部署: "30m" - 快速适应高频访问模式
  prefetch_time_window: "1h"
  
  # 最大跟踪条目数
  # 限制内存使用，防止无限增长
  # 范围: 1000-100000
  # 默认: 10000
  # 建议:
  #   - 家庭用户: 5000-10000
  #   - 小型企业: 10000-20000
  #   - 大型部署: 20000-50000
  prefetch_max_entries: 10000
  
  # 清理间隔
  # 定期清理不活跃的域名记录
  # 格式: 时间字符串 (如 "1h", "30m", "2h30m")
  # 默认: "1h"
  # 建议: "1h" (大多数场景)
  prefetch_cleanup_interval: "1h"
  
  # ========== 动态并发控制（推荐） ==========
  
  # 软限制 - 正常并发数
  # 范围: 10-500, 默认: 50
  prefetch_soft_limit: 50
  
  # 硬限制 - 最大并发数
  # 范围: 50-1000, 默认: 150
  prefetch_hard_limit: 150
  
  # 紧急队列大小
  # 范围: 100-5000, 默认: 500
  prefetch_urgent_queue_size: 500
  
  # 正常队列大小
  # 范围: 500-10000, 默认: 2000
  prefetch_normal_queue_size: 2000
  
  # 紧急优先级阈值
  # 范围: 50-90, 默认: 70
  prefetch_urgent_threshold: 70
```

## 配置示例

### 示例1: 家庭用户（默认配置）

```yaml
dns:
  prefetch_enabled: true
  prefetch_threshold: 5
  prefetch_time_window: "1h"
  prefetch_max_entries: 10000
  prefetch_cleanup_interval: "1h"
  prefetch_soft_limit: 30
  prefetch_hard_limit: 80
  prefetch_urgent_queue_size: 200
  prefetch_normal_queue_size: 1000
  prefetch_urgent_threshold: 70
```

**适用场景**:
- 家庭网络（2-10台设备）
- DNS查询量: 10-100 QPS
- 内存使用: ~1MB

---

### 示例2: 小型企业

```yaml
dns:
  prefetch_enabled: true
  prefetch_threshold: 8
  prefetch_time_window: "2h"
  prefetch_max_entries: 20000
  prefetch_cleanup_interval: "1h"
  prefetch_soft_limit: 50
  prefetch_hard_limit: 150
  prefetch_urgent_queue_size: 500
  prefetch_normal_queue_size: 2000
  prefetch_urgent_threshold: 70
```

**适用场景**:
- 小型办公室（10-50台设备）
- DNS查询量: 100-500 QPS
- 内存使用: ~2MB

---

### 示例3: 大型部署

```yaml
dns:
  prefetch_enabled: true
  prefetch_threshold: 15
  prefetch_time_window: "30m"
  prefetch_max_entries: 50000
  prefetch_cleanup_interval: "30m"
  prefetch_soft_limit: 100
  prefetch_hard_limit: 300
  prefetch_urgent_queue_size: 1000
  prefetch_normal_queue_size: 5000
  prefetch_urgent_threshold: 70
```

**适用场景**:
- 企业网络（100+台设备）
- DNS查询量: 1000+ QPS
- 内存使用: ~5MB

---

### 示例4: 低性能设备（树莓派等）

```yaml
dns:
  prefetch_enabled: true
  prefetch_threshold: 3
  prefetch_time_window: "1h"
  prefetch_max_entries: 5000
  prefetch_cleanup_interval: "2h"
  prefetch_soft_limit: 20
  prefetch_hard_limit: 50
  prefetch_urgent_queue_size: 100
  prefetch_normal_queue_size: 500
  prefetch_urgent_threshold: 70
```

**适用场景**:
- 树莓派、嵌入式设备
- 有限的CPU和内存
- 内存使用: ~0.5MB

---

## 参数详解

### prefetch_enabled

**类型**: Boolean  
**默认值**: false  
**说明**: 主开关，控制是否启用预取功能

**何时启用**:
- ✅ 有稳定的上游DNS服务器
- ✅ 网络带宽充足
- ✅ 希望减少DNS查询延迟

**何时禁用**:
- ❌ 网络带宽有限
- ❌ 上游DNS服务器不稳定
- ❌ 设备性能极低

---

### prefetch_threshold

**类型**: Integer  
**范围**: 1-100  
**默认值**: 5  
**说明**: 域名需要被访问多少次才会被预取

**调优建议**:
- **值太低** (1-2): 
  - 优点: 更多域名被预取，延迟更低
  - 缺点: 内存使用增加，可能预取不必要的域名
  
- **值适中** (3-10):
  - 平衡性能和资源使用
  - 适合大多数场景
  
- **值太高** (15+):
  - 优点: 只预取真正热门的域名
  - 缺点: 可能错过一些应该预取的域名

**推荐值**:
- 家庭用户: 3-5
- 小型企业: 5-10
- 大型部署: 10-20

---

### prefetch_max_entries

**类型**: Integer  
**范围**: 1000-100000  
**默认值**: 10000  
**说明**: 最多跟踪多少个域名

**内存估算**:
```
每个条目 ≈ 100 bytes
10,000 条目 ≈ 1 MB
50,000 条目 ≈ 5 MB
```

**调优建议**:
- 根据可用内存调整
- 监控实际使用量（通过日志或API）
- 如果经常触发清理，可以适当增加

---

### prefetch_cleanup_interval

**类型**: Duration String  
**默认值**: "1h"  
**说明**: 多久清理一次不活跃的域名

**格式示例**:
- "30m" - 30分钟
- "1h" - 1小时
- "2h30m" - 2小时30分钟
- "24h" - 24小时

**调优建议**:
- **频繁清理** (30m):
  - 优点: 内存使用更稳定
  - 缺点: 略微增加CPU使用
  - 适合: 大型部署、内存受限环境
  
- **正常清理** (1h):
  - 平衡性能和资源
  - 适合大多数场景
  
- **不频繁清理** (2h+):
  - 优点: 减少CPU开销
  - 缺点: 内存可能增长
  - 适合: 内存充足的环境

---

### prefetch_max_concurrent_refresh

**类型**: Integer  
**范围**: 10-200  
**默认值**: 50  
**说明**: 同时进行的域名刷新操作数量

**调优建议**:
- **值太低** (10-20):
  - 优点: 资源使用少
  - 缺点: 刷新速度慢，可能来不及刷新所有域名
  - 适合: 低性能设备
  
- **值适中** (40-80):
  - 平衡刷新速度和资源使用
  - 适合大多数场景
  
- **值太高** (100+):
  - 优点: 刷新速度快
  - 缺点: 可能占用较多网络和CPU资源
  - 适合: 高性能服务器

---

## 监控和调优

### 查看日志

启动时会显示配置信息：
```
INFO prefetch manager initialized threshold=5 max_entries=10000 cleanup_interval=1h0m0s max_concurrent_refresh=50
INFO prefetch enabled and started
```

清理时会显示统计：
```
INFO prefetch cleanup completed removed=123 remaining_hits=456 remaining_domains=78
```

### 性能指标

通过API查看统计（如果已实现）：
```bash
curl http://localhost:3000/api/prefetch/stats
```

响应示例：
```json
{
  "hits": 1234,
  "domains": 567,
  "tracked": 1234,
  "cache_size_bytes": 102400
}
```

### 调优流程

1. **启用并观察**
   ```yaml
   prefetch_enabled: true
   # 使用默认值
   ```

2. **监控内存使用**
   - 如果内存增长过快 → 降低 `prefetch_max_entries`
   - 如果内存稳定 → 保持当前配置

3. **监控刷新效果**
   - 查看日志中的 `refreshing hot domains count=X`
   - 如果数量过多 → 提高 `prefetch_threshold`
   - 如果数量过少 → 降低 `prefetch_threshold`

4. **调整并发数**
   - 如果CPU使用率高 → 降低 `prefetch_max_concurrent_refresh`
   - 如果刷新不及时 → 提高 `prefetch_max_concurrent_refresh`

---

## 常见问题

### Q1: Prefetch会增加多少内存使用？

**A**: 取决于 `prefetch_max_entries` 配置：
- 10,000 条目 ≈ 1 MB
- 20,000 条目 ≈ 2 MB
- 50,000 条目 ≈ 5 MB

### Q2: Prefetch会增加上游DNS服务器的负载吗？

**A**: 会略微增加，但影响很小：
- 只刷新热门域名（达到阈值的）
- 只在缓存即将过期时刷新
- 有并发限制，不会突发大量请求

### Q3: 如何知道Prefetch是否在工作？

**A**: 查看日志：
```
DEBUG refreshing hot domains count=10
DEBUG refreshed domain domain=example.com.
```

### Q4: 可以动态修改配置吗？

**A**: 目前需要：
1. 修改 `AdGuardHome.yaml`
2. 重启 AdGuard Home

未来版本可能支持热重载。

### Q5: Prefetch对DNS查询延迟的改善有多大？

**A**: 对于热门域名：
- 缓存过期时: 从 50-200ms 降至 <1ms
- 平均改善: 50-70%
- 用户体验: 页面加载更快

---

## 最佳实践

### ✅ 推荐做法

1. **从默认配置开始**
   ```yaml
   prefetch_enabled: true
   # 其他参数使用默认值
   ```

2. **监控一周后再调整**
   - 观察内存使用趋势
   - 查看日志中的统计信息
   - 根据实际情况微调

3. **根据设备性能调整**
   - 低性能设备: 降低所有限制
   - 高性能设备: 可以提高限制

4. **定期检查日志**
   - 确保没有错误
   - 确认清理正常进行
   - 验证刷新操作成功

### ❌ 避免做法

1. **不要设置过高的 max_entries**
   - 可能导致内存耗尽
   - 大多数场景 10,000 已足够

2. **不要设置过低的 threshold**
   - 会预取太多不必要的域名
   - 浪费资源

3. **不要在网络不稳定时启用**
   - 可能导致大量刷新失败
   - 增加日志噪音

---

## 故障排查

### 问题: Prefetch没有启动

**检查**:
1. 配置文件中 `prefetch_enabled: true`
2. 查看启动日志是否有 "prefetch enabled and started"
3. 检查配置文件语法是否正确

### 问题: 内存持续增长

**解决**:
1. 降低 `prefetch_max_entries`
2. 缩短 `prefetch_cleanup_interval`
3. 提高 `prefetch_threshold`

### 问题: 刷新失败过多

**检查**:
1. 上游DNS服务器是否正常
2. 网络连接是否稳定
3. 查看错误日志

**解决**:
- 降低 `prefetch_max_concurrent_refresh`
- 检查上游DNS配置

---

## 前端UI集成建议

### 配置界面设计

```
┌─────────────────────────────────────────┐
│ DNS Prefetch (缓存预取)                  │
├─────────────────────────────────────────┤
│                                         │
│ [✓] 启用DNS预取                         │
│                                         │
│ 热门域名阈值: [5    ] (1-100)          │
│ 说明: 访问次数达到此值后开始预取         │
│                                         │
│ 最大跟踪条目: [10000] (1000-100000)    │
│ 说明: 限制内存使用                      │
│                                         │
│ 清理间隔: [1] 小时                      │
│                                         │
│ 最大并发刷新: [50  ] (10-200)          │
│ 说明: 同时刷新的域名数量                │
│                                         │
│ [应用配置]  [恢复默认]                  │
│                                         │
│ 当前状态:                               │
│ - 跟踪域名: 1,234                       │
│ - 活跃域名: 567                         │
│ - 内存使用: ~1.2 MB                     │
│                                         │
└─────────────────────────────────────────┘
```

### API端点建议

```
GET  /api/prefetch/config  - 获取配置
POST /api/prefetch/config  - 更新配置
GET  /api/prefetch/stats   - 获取统计信息
POST /api/prefetch/clear   - 清理缓存
```

---

**文档版本**: 1.0  
**更新日期**: 2024-11-26  
**适用版本**: V3 Optimized+
