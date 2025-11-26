# Prefetch UI Integration Guide

## 概述

本文档说明如何在 AdGuard Home 的前端 UI 中配置 DNS Prefetch（预取）功能。

## 功能说明

DNS Prefetch 是一个智能缓存预热功能，它会自动跟踪热门域名并在缓存过期前主动刷新，从而降低 DNS 查询延迟。

### 工作原理

1. **访问跟踪**：系统跟踪每个域名的访问次数和时间
2. **热度判断**：当域名在时间窗口内的访问次数超过阈值时，标记为"热门域名"
3. **主动刷新**：在缓存即将过期时，自动向上游服务器查询并更新缓存
4. **自动清理**：定期清理不再活跃的域名，释放内存

## UI 配置位置

前端路径：**设置 → DNS 设置 → DNS 缓存配置**

在 DNS 缓存配置卡片中，新增了"DNS 预取设置"部分。

## 配置参数

### 1. 启用 DNS 预取 (prefetch_enabled)
- **类型**：布尔值
- **默认值**：false
- **说明**：主开关，启用后才会进行预取操作

### 2. 访问阈值 (prefetch_threshold)
- **类型**：整数
- **范围**：1-100
- **默认值**：5
- **说明**：域名被视为"热门"所需的最小访问次数
- **示例**：设置为 5 表示域名在时间窗口内被访问 5 次后才会被预取

### 3. 时间窗口 (prefetch_time_window)
- **类型**：整数（秒）
- **范围**：60-86400 秒（1 分钟 - 24 小时）
- **默认值**：3600 秒（1 小时）
- **说明**：统计访问次数的时间窗口
- **示例**：设置为 3600 表示只统计最近 1 小时内的访问

### 4. 最大跟踪域名数 (prefetch_max_entries)
- **类型**：整数
- **范围**：1000-100000
- **默认值**：10000
- **说明**：系统最多跟踪的域名数量，防止内存无限增长
- **建议**：根据服务器内存调整，每个域名约占用 100-200 字节

### 5. 清理间隔 (prefetch_cleanup_interval)
- **类型**：整数（秒）
- **范围**：900-14400 秒（15 分钟 - 4 小时）
- **默认值**：3600 秒（1 小时）
- **说明**：自动清理过期条目的间隔时间

### 6. 软并发限制 (prefetch_soft_limit)
- **类型**：整数
- **范围**：10-500
- **默认值**：50
- **说明**：正常情况下的最大并发刷新数
- **建议**：根据服务器性能和网络带宽调整

### 7. 硬并发限制 (prefetch_hard_limit)
- **类型**：整数
- **范围**：50-1000
- **默认值**：150
- **说明**：紧急情况下的最大并发刷新数
- **约束**：必须大于或等于软限制

## 配置示例

### 场景 1：家庭网络（默认配置）
```yaml
prefetch_enabled: true
prefetch_threshold: 5
prefetch_time_window: 3600      # 1 小时
prefetch_max_entries: 10000
prefetch_cleanup_interval: 3600  # 1 小时
prefetch_soft_limit: 50
prefetch_hard_limit: 150
```

### 场景 2：小型企业网络
```yaml
prefetch_enabled: true
prefetch_threshold: 3            # 更低的阈值，更积极预取
prefetch_time_window: 7200       # 2 小时
prefetch_max_entries: 20000      # 跟踪更多域名
prefetch_cleanup_interval: 1800  # 30 分钟，更频繁清理
prefetch_soft_limit: 100
prefetch_hard_limit: 300
```

### 场景 3：大型企业网络
```yaml
prefetch_enabled: true
prefetch_threshold: 10           # 更高的阈值，只预取真正热门的域名
prefetch_time_window: 3600       # 1 小时
prefetch_max_entries: 50000      # 跟踪大量域名
prefetch_cleanup_interval: 3600  # 1 小时
prefetch_soft_limit: 200
prefetch_hard_limit: 500
```

## API 端点

### 获取配置
```
GET /control/dns_info
```

响应示例：
```json
{
  "cache_enabled": true,
  "cache_size": 4194304,
  "prefetch_enabled": true,
  "prefetch_threshold": 5,
  "prefetch_time_window": 3600,
  "prefetch_max_entries": 10000,
  "prefetch_cleanup_interval": 3600,
  "prefetch_soft_limit": 50,
  "prefetch_hard_limit": 150
}
```

### 设置配置
```
POST /control/dns_config
Content-Type: application/json

{
  "prefetch_enabled": true,
  "prefetch_threshold": 5,
  "prefetch_time_window": 3600,
  "prefetch_max_entries": 10000,
  "prefetch_cleanup_interval": 3600,
  "prefetch_soft_limit": 50,
  "prefetch_hard_limit": 150
}
```

## 配置文件

配置会保存在 `AdGuardHome.yaml` 文件中：

```yaml
dns:
  # ... 其他 DNS 配置 ...
  
  # Prefetch 配置
  prefetch_enabled: true
  prefetch_threshold: 5
  prefetch_time_window: 1h
  prefetch_max_entries: 10000
  prefetch_cleanup_interval: 1h
  prefetch_soft_limit: 50
  prefetch_hard_limit: 150
```

## 验证配置

### 1. 检查配置是否生效
启动 AdGuard Home 后，查看日志：
```
[info] prefetch manager initialized threshold=5 time_window=1h0m0s max_entries=10000 ...
```

### 2. 监控预取状态
查看日志中的预取指标（每分钟输出一次）：
```
[info] prefetch metrics active_tasks=10 urgent_queue=5 normal_queue=20 ...
```

### 3. 前端验证
1. 打开 AdGuard Home Web 界面
2. 进入"设置 → DNS 设置"
3. 找到"DNS 缓存配置"卡片
4. 展开"DNS 预取设置"部分
5. 确认所有参数正确显示

## 注意事项

1. **内存使用**：启用预取会增加内存使用，建议根据服务器配置调整 `prefetch_max_entries`
2. **网络带宽**：预取会产生额外的上游查询，注意监控网络流量
3. **配置重启**：修改预取配置需要重启 DNS 服务才能生效
4. **参数验证**：
   - 软限制必须 ≤ 硬限制
   - 所有参数必须在有效范围内
   - 时间参数以秒为单位

## 故障排查

### 问题 1：配置保存后不生效
- **原因**：配置需要重启 DNS 服务
- **解决**：保存配置后，系统会自动重启 DNS 服务

### 问题 2：内存使用过高
- **原因**：`prefetch_max_entries` 设置过大
- **解决**：降低 `prefetch_max_entries` 值

### 问题 3：预取任务被丢弃
- **原因**：并发限制过低或队列已满
- **解决**：增加 `prefetch_soft_limit` 和 `prefetch_hard_limit`

## 性能建议

1. **初始配置**：使用默认值，观察一段时间后再调整
2. **阈值调整**：
   - 访问量大：提高阈值（减少预取）
   - 访问量小：降低阈值（增加预取）
3. **时间窗口**：
   - 访问模式稳定：使用较长时间窗口
   - 访问模式多变：使用较短时间窗口
4. **并发限制**：
   - 服务器性能好：提高限制
   - 服务器性能差：降低限制

## 相关文档

- [PREFETCH_CONFIG_GUIDE.md](./PREFETCH_CONFIG_GUIDE.md) - 详细配置指南
- [PREFETCH_MEMORY_LEAK_FIX.md](./PREFETCH_MEMORY_LEAK_FIX.md) - 内存泄漏修复说明
- [V3_ALL_OPTIMIZATIONS_COMPLETE.md](./V3_ALL_OPTIMIZATIONS_COMPLETE.md) - V3 版本优化总结
