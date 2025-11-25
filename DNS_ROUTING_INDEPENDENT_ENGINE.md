# DNS路由独立引擎实现

## 问题描述

之前的实现中，DNS路由过滤器被添加到 `allowFilters` 中，这导致它们受到 `FilteringEnabled` 和 `ProtectionEnabled` 设置的影响。当用户关闭这些保护开关时，DNS路由功能也会停止工作。

## 解决方案

创建了一个完全独立的DNS路由引擎 (`filteringEngineDnsRouting`)，它：
- 不受 `FilteringEnabled` 影响
- 不受 `ProtectionEnabled` 影响
- 只受各个DNS路由过滤器自己的 `Enabled` 字段控制
- 总是优先检查（在其他过滤规则之前）

## 核心问题

之前即使创建了独立的DNS路由引擎，但在 `matchHost` 函数开始处有一个提前返回：

```go
if !setts.FilteringEnabled {
    return Result{}, nil  // ❌ 整个函数直接返回，DNS路由引擎根本不会被检查！
}
```

这导致当 `FilteringEnabled` 为 false 时，DNS路由引擎永远不会被调用。

## 实现细节

### 1. 新增字段

在 `DNSFilter` 结构中添加了两个新字段：

```go
// filteringEngineBlock is for blocklist filters.
filteringEngineBlock *urlfilter.DNSEngine

// filteringEngineDnsRouting is for DNS routing filters (independent of FilteringEnabled).
filteringEngineDnsRouting *urlfilter.DNSEngine
```

### 2. 修改函数签名

更新了以下函数以支持独立的DNS路由过滤器列表：

- `setFilters(ctx, blockFilters, allowFilters, dnsRoutingFilters, async)`
- `initFiltering(ctx, allowFilters, blockFilters, dnsRoutingFilters)`
- `filtersInitializerParams` 结构添加了 `dnsRoutingFilters` 字段

### 3. 分离过滤器处理

在 `enableFiltersLocked` 函数中：
- DNS路由过滤器不再添加到 `allowFilters`
- 创建独立的 `dnsRoutingFilters` 列表
- 按优先级排序后传递给 `setFilters`

### 4. 独立引擎初始化

在 `initFiltering` 函数中：
- 为DNS路由过滤器创建独立的 `RuleStorage`
- 创建独立的 `DNSEngine`
- 在锁保护下设置 `filteringEngineDnsRouting`

### 5. 优先检查逻辑（关键修复）

**这是最关键的修复！** 在 `matchHost` 函数中重新组织了检查顺序：

```go
func (d *DNSFilter) matchHost(...) (res Result, err error) {
    // 准备请求对象
    ufReq := &urlfilter.DNSRequest{...}
    
    d.engineLock.RLock()
    defer d.engineLock.RUnlock()
    
    // 1. 首先检查DNS路由引擎（总是活动，在FilteringEnabled检查之前！）
    if d.filteringEngineDnsRouting != nil {
        dnsres, ok := d.filteringEngineDnsRouting.MatchRequest(ufReq)
        if ok && result.UpstreamGroup != "" {
            return result, nil  // ✅ 立即返回，不受任何设置影响
        }
    }
    
    // 2. 然后才检查FilteringEnabled（DNS路由已经检查过了）
    if !setts.FilteringEnabled {
        return Result{}, nil  // 其他过滤被跳过
    }
    
    // 3. 检查允许列表（仅当保护开启时）
    if setts.ProtectionEnabled && d.filteringEngineAllow != nil {
        // 检查允许列表
    }
    
    // 4. 检查阻止列表（仅当保护开启时）
    if d.filteringEngine != nil {
        // 检查阻止列表
    }
}
```

**关键点**：DNS路由引擎的检查在 `FilteringEnabled` 检查**之前**，确保即使过滤被禁用，DNS路由仍然工作。

## 行为对照表

| 场景 | ProtectionEnabled | FilteringEnabled | DNS路由过滤器Enabled | 结果 |
|------|------------------|------------------|-------------------|------|
| 1 | ✅ true | ✅ true | ✅ true | ✅ DNS路由工作 |
| 2 | ✅ true | ❌ false | ✅ true | ✅ DNS路由工作 |
| 3 | ❌ false | ✅ true | ✅ true | ✅ DNS路由工作 |
| 4 | ❌ false | ❌ false | ✅ true | ✅ DNS路由工作 |
| 5 | any | any | ❌ false | ❌ DNS路由不工作 |

## 关键优势

1. **完全独立**：DNS路由功能不再依赖全局保护设置
2. **灵活控制**：每个DNS路由过滤器可以单独启用/禁用
3. **优先处理**：DNS路由规则优先于其他过滤规则
4. **向后兼容**：不影响现有的过滤和保护功能

## 测试建议

1. 关闭 `ProtectionEnabled` 和 `FilteringEnabled`
2. 确保DNS路由过滤器的 `Enabled` 字段为 `true`
3. 测试DNS查询是否仍然被正确路由到指定的上游组
4. 检查日志中的 "DNS routing matched" 消息

## 相关文件

- `internal/filtering/filtering.go` - 核心过滤逻辑
- `internal/filtering/filter.go` - 过滤器管理
- `AdGuardHome_v3_latest.exe` - 包含此修复的最新版本
