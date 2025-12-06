# DNS Routing 和 Filtering 系统分离修复总结

## 问题描述

1. **死锁问题**：点击更新时出现死锁并目录错乱
2. **API 混用**：前端使用 filtering API 操作 DNS routing 规则
3. **错误提示**：启用规则时出现 "DNS routing filters must be updated through /control/dns_routing/update endpoint"

## 根本原因

- DNS routing 规则仍在使用旧的 filtering 系统进行更新
- 前端调用 `/control/filtering/set_url` 而不是 `/control/dns_routing/update`
- 两个系统职责不清晰，导致死锁和混乱

## 修复方案

### 1. 后端修改

#### `internal/filtering/http.go`
阻止通过 filtering API 操作 DNS routing 规则：

```go
// handleFilteringAddURL - 拒绝 dns_routing=true
if fj.DnsRouting {
    return error("DNS routing filters must be added through /control/dns_routing/add endpoint")
}

// handleFilteringSetURL - 拒绝 dns_routing=true
if fj.DnsRouting || fj.Data.DnsRouting {
    return error("DNS routing filters must be updated through /control/dns_routing/update endpoint")
}

// handleFilteringRemoveURL - 拒绝 dns_routing=true
if req.DnsRouting {
    return error("DNS routing filters must be removed through /control/dns_routing/delete endpoint")
}

// handleFilteringRefresh - 拒绝 dns_routing=true
if req.DnsRouting {
    return error("DNS routing filters must be refreshed through /control/dns_routing/refresh endpoint")
}
```

#### `internal/filtering/filter.go`
确保内部方法也跳过 DNS routing 规则：

```go
// tryRefreshSingleFilter - 跳过 DNS routing 规则
if dnsRouting {
    return error("DNS routing filters must be refreshed through the DNS routing file manager")
}

// filterSetProperties - 拒绝更新 DNS routing 规则
if flt.DnsRouting {
    return error("DNS routing filters must be updated through the DNS routing file manager")
}
```

### 2. 前端修改

#### `client/src/api/Api.ts`
添加专用的 DNS routing API 方法：

```typescript
// DNS Routing
getDnsRoutingRules()
addDnsRoutingRule(data)
updateDnsRoutingRule(data)
deleteDnsRoutingRule(data)
refreshDnsRoutingRule(data)
```

#### `client/src/actions/dnsRouting.ts`
更新所有 actions 使用新的 API：

```typescript
// 使用 ID 而不是 URL
getDnsRoutingFilters() -> getDnsRoutingRules()
addDnsRoutingFilter() -> addDnsRoutingRule()
editDnsRoutingFilter(id, data) -> updateDnsRoutingRule({id, ...data})
removeDnsRoutingFilter(id) -> deleteDnsRoutingRule({id})
toggleDnsRoutingFilter(filter) -> updateDnsRoutingRule({id, ...filter, enabled: !filter.enabled})
refreshDnsRoutingFilters(id) -> refreshDnsRoutingRule({id})
```

#### `client/src/reducers/dnsRouting.ts`
添加 `modalFilter` 字段存储正在编辑的 filter 对象

#### `client/src/components/Filters/DnsRouting.tsx`
更新组件使用新的 action 签名和 modalFilter

## 关键特性

### ✅ 启用/禁用不触发下载

`UpdateDomainListRule` 方法智能检测：

```go
// 只有 URL 改变时才重新下载
urlChanged := existingRule.URL != rule.URL
if urlChanged {
    // 重新下载和解析
    data, err := m.downloadRuleFile(ctx, rule.URL)
    // ...
} else {
    // 保留现有文件信息
    rule.RulesCount = existingRule.RulesCount
    rule.LastUpdated = existingRule.LastUpdated
    rule.FilePath = existingRule.FilePath
}
```

**行为说明**：
- ✅ **启用/禁用**：只更新 `enabled` 标志，不触发下载，`last_updated` 不变
- ✅ **更新 URL**：触发重新下载，更新 `last_updated`
- ✅ **添加规则**：触发下载
- ✅ **手动刷新**：触发重新下载

### ✅ 完全分离

```
DNS Routing 规则:
  API: /control/dns_routing/*
  存储: data/dns_routing_rules/
  管理: dnsroutingfiles.Manager

Filtering 规则:
  API: /control/filtering/*
  存储: data/filters/
  管理: filtering.DNSFilter
```

### ✅ 避免死锁

- File Manager 在调用 router 更新前释放锁
- 使用 goroutine 异步通知 router
- 清晰的锁顺序和范围

## 测试

### 测试脚本

1. **`test-frontend-fix.ps1`** - 前端修复测试说明
2. **`test-toggle-no-download.ps1`** - 验证启用/禁用不触发下载

### 测试步骤

```powershell
# 1. 运行新版本
.\AdGuardHome_final.exe

# 2. 测试启用/禁用不触发下载
.\test-toggle-no-download.ps1

# 3. 在浏览器中测试
# - 访问 http://localhost:3000
# - 进入 DNS 路由页面
# - 测试启用/禁用、编辑、删除、刷新功能
```

### 预期结果

- ✅ 所有操作通过 `/control/dns_routing/*` 端点
- ✅ 启用/禁用不改变 `last_updated`
- ✅ 不再出现 "must be updated through" 错误
- ✅ 不再出现死锁
- ✅ 不再出现目录混乱

## 文件清单

### 修改的文件

**后端**:
- `internal/filtering/http.go` - 阻止 filtering API 操作 DNS routing
- `internal/filtering/filter.go` - 内部方法跳过 DNS routing
- `internal/home/dns_routing.go` - DNS routing 专用处理器（已存在）
- `internal/dnsroutingfiles/rules.go` - 智能更新逻辑（已存在）

**前端**:
- `client/src/api/Api.ts` - 添加 DNS routing API 方法
- `client/src/actions/dnsRouting.ts` - 使用新 API
- `client/src/reducers/dnsRouting.ts` - 添加 modalFilter
- `client/src/components/Filters/DnsRouting.tsx` - 更新组件逻辑

### 新增的文件

- `AdGuardHome_final.exe` - 修复后的可执行文件
- `test-toggle-no-download.ps1` - 测试脚本
- `test-frontend-fix.ps1` - 测试说明
- `SEPARATION_FIX_SUMMARY.md` - 本文档

## 总结

通过完全分离 DNS routing 和 filtering 系统：

1. **清晰的职责分离** - 每个系统有自己的 API、存储和管理器
2. **避免死锁** - 正确的锁管理和异步通知
3. **智能更新** - 只在必要时重新下载
4. **用户友好** - 清晰的错误消息指导正确使用

所有功能现在都应该正常工作！
