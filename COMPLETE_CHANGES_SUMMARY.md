# DNS 路由规则完整修改总结

## 核心原则

**配置文件（AdGuardHome.yaml）是唯一的数据源，metadata.json 不再使用。**

## 修改的文件清单

### 1. internal/dnsroutingfiles/storage.go
**修改内容**：
- `saveMetadata()` → 改为空函数（不再保存 metadata）
- `loadMetadata()` → 改为空函数（不再加载 metadata）
- 移除未使用的 imports（json, yaml, time）

**目的**：停止使用 metadata 文件

### 2. internal/dnsroutingfiles/manager.go
**修改内容**：
- `LoadAll()` → 保持接口不变，但内部不再加载 metadata
- 添加 `GetRuleFilePath(ruleID int64) string` 方法到接口和实现

**目的**：提供获取规则文件路径的方法

### 3. internal/dnsroutingfiles/rules.go
**修改内容**：
- `AddDomainListRule()` → 移除对 `m.domainListRules` 的检查和更新，只操作文件
- `UpdateDomainListRule()` → 移除对 `m.domainListRules` 的检查和更新，只操作文件
- `DeleteDomainListRule()` → 移除对 `m.domainListRules` 的检查和更新，只操作文件

**目的**：File Manager 只负责文件操作，不管理规则列表

### 4. internal/home/dns_routing.go
**修改内容**：

#### handleGetDnsRoutingRules
```go
// 从配置文件读取，而不是从 File Manager
config.RLock()
filters := config.Filters
config.RUnlock()

for _, filter := range filters {
    if !filter.DnsRouting {
        continue
    }
    // 返回规则
}
```

#### handleAddDnsRoutingRule
```go
// 1. File Manager 下载并解析
err = globalContext.dnsRoutingFileManager.AddDomainListRule(ctx, rule)

// 2. 添加到 config.Filtering.Filters（注意是 Filtering.Filters）
config.Lock()
newFilter := filtering.FilterYAML{...}
config.Filtering.Filters = append(config.Filtering.Filters, newFilter)
config.Unlock()

// 3. 保存配置
if !web.saveConfigIfNeeded(ctx, l, r, w) {
    return
}
```

#### handleUpdateDnsRoutingRule
```go
// 1. File Manager 重新下载
err = globalContext.dnsRoutingFileManager.UpdateDomainListRule(ctx, rule)

// 2. 更新 config.Filtering.Filters
config.Lock()
for i := range config.Filtering.Filters {
    if int64(config.Filtering.Filters[i].ID) == req.ID && config.Filtering.Filters[i].DnsRouting {
        // 更新字段
        break
    }
}
config.Unlock()

// 3. 保存配置
if !web.saveConfigIfNeeded(ctx, l, r, w) {
    return
}
```

#### handleDeleteDnsRoutingRule
```go
// 1. File Manager 删除文件
err = globalContext.dnsRoutingFileManager.DeleteDomainListRule(ctx, req.ID)

// 2. 从 config.Filtering.Filters 移除
config.Lock()
newFilters := make([]filtering.FilterYAML, 0, len(config.Filtering.Filters))
for _, filter := range config.Filtering.Filters {
    if int64(filter.ID) != req.ID || !filter.DnsRouting {
        newFilters = append(newFilters, filter)
    }
}
config.Filtering.Filters = newFilters
config.Unlock()

// 3. 保存配置
if !web.saveConfigIfNeeded(ctx, l, r, w) {
    return
}
```

**目的**：API 直接操作配置文件

### 5. internal/home/config.go
**修改内容**：
```go
// 之前：过滤掉 DNS 路由规则
regularFilters := make([]filtering.FilterYAML, 0, len(config.Filtering.Filters))
for _, f := range config.Filtering.Filters {
    if !f.DnsRouting {
        regularFilters = append(regularFilters, f)
    }
}
config.Filters = regularFilters

// 现在：保留所有规则
config.Filters = config.Filtering.Filters
```

**目的**：保存配置时不过滤 DNS 路由规则

### 6. internal/home/dns.go
**修改内容**：
```go
// reloadDnsRoutingRules 函数

// 之前：从 File Manager 读取规则
domainListRules := globalContext.dnsRoutingFileManager.GetAllDomainListRules()

// 现在：从配置文件读取规则
config.RLock()
filters := config.Filters
config.RUnlock()

for _, filter := range filters {
    if !filter.DnsRouting || !filter.Enabled {
        continue
    }
    
    // 获取文件路径
    filePath := globalContext.dnsRoutingFileManager.GetRuleFilePath(int64(filter.ID))
    
    // 读取并解析规则文件
    // 更新路由器
}
```

**目的**：路由器重新加载时从配置文件读取规则

## 关键点

### 1. 配置存储位置
- **正确**：`config.Filtering.Filters`
- **错误**：`config.Filters`（这个会被 `config.write()` 覆盖）

### 2. 数据流

#### 添加规则
```
API 请求
  ↓
File Manager: 下载并解析规则文件
  ↓
API: 添加到 config.Filtering.Filters
  ↓
config.write(): 保存到 AdGuardHome.yaml
  ↓
Router: 重新加载规则
```

#### 读取规则
```
API 请求
  ↓
读取 config.Filters（由 config.write() 从 config.Filtering.Filters 复制）
  ↓
返回规则列表
```

#### 路由器加载
```
启动/重新加载
  ↓
读取 config.Filters
  ↓
遍历 dns_routing=true 的规则
  ↓
File Manager: 获取规则文件路径
  ↓
读取并解析规则文件
  ↓
更新路由器
```

### 3. File Manager 职责
- ✅ 下载规则文件
- ✅ 解析规则文件
- ✅ 保存规则文件到磁盘
- ✅ 删除规则文件
- ✅ 提供规则文件路径
- ❌ 不管理规则列表
- ❌ 不保存 metadata

### 4. API 层职责
- ✅ 验证请求参数
- ✅ 调用 File Manager 操作文件
- ✅ 更新 config.Filtering.Filters
- ✅ 保存配置文件
- ✅ 从 config.Filters 读取规则

## 没有修改的内容

以下内容保持不变：
- Custom Rules 的处理逻辑
- Refresh 规则的逻辑
- 其他 filtering 相关的代码
- 前端代码（除了之前的修改）

## 测试检查点

1. ✅ 编译成功
2. ⏳ 添加规则 → 配置文件中应该有对应条目
3. ⏳ 重启后规则仍然存在
4. ⏳ 路由器正确加载规则（日志显示 domain_list_rules > 0）
5. ⏳ 更新规则 → 配置文件更新
6. ⏳ 删除规则 → 配置文件中移除
7. ⏳ 手动编辑配置文件 → 重启后生效

## 潜在问题

如果还有问题，可能的原因：
1. `config.Filtering.Filters` 和 `config.Filters` 的同步问题
2. `config.write()` 的调用时机
3. 配置文件的读取权限
4. 配置文件的写入权限

## 回滚方案

如果需要回滚，恢复以下文件：
1. internal/dnsroutingfiles/storage.go
2. internal/dnsroutingfiles/manager.go
3. internal/dnsroutingfiles/rules.go
4. internal/home/dns_routing.go
5. internal/home/config.go
6. internal/home/dns.go
