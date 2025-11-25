# DNS路由规则URL重复检查修复

## 问题描述

添加DNS路由规则时，即使URL在DNS路由列表中不存在，但如果该URL已经在黑名单或白名单中使用，系统也会报错"Filter with URL already exists"。

这是因为 `filterExists()` 函数检查了所有过滤器列表（黑名单、白名单、DNS路由），导致不同类型的过滤器不能使用相同的URL。

## 问题原因

在 `handleFilteringAddURL` 函数中：
```go
// 旧代码 - 检查所有列表
if d.filterExists(fj.URL) {
    err = errFilterExists
    // 返回错误
}
```

`filterExists()` 函数检查逻辑：
```go
func (d *DNSFilter) filterExistsLocked(url string) (ok bool) {
    // 检查黑名单
    for _, f := range d.conf.Filters {
        if f.URL == url {
            return true
        }
    }
    
    // 检查白名单
    for _, f := range d.conf.WhitelistFilters {
        if f.URL == url {
            return true
        }
    }
    
    // 检查DNS路由
    for _, f := range d.conf.DnsRoutingFilters {
        if f.URL == url {
            return true
        }
    }
    
    return false
}
```

这导致：
- 如果URL在黑名单中，就不能在DNS路由中使用
- 如果URL在白名单中，就不能在DNS路由中使用
- 反之亦然

## 解决方案

添加新函数 `filterExistsInType()`，只检查相同类型的过滤器列表。

### 修改内容

#### 1. `internal/filtering/filter.go`

添加新函数：
```go
// filterExistsInType returns true if a filter with the same url exists in the specified filter type.
// It's safe for concurrent use.
func (d *DNSFilter) filterExistsInType(url string, isWhitelist bool, isDnsRouting bool) (ok bool) {
    d.conf.filtersMu.RLock()
    defer d.conf.filtersMu.RUnlock()

    var filters []FilterYAML
    if isDnsRouting {
        filters = d.conf.DnsRoutingFilters
    } else if isWhitelist {
        filters = d.conf.WhitelistFilters
    } else {
        filters = d.conf.Filters
    }

    for _, f := range filters {
        if f.URL == url {
            return true
        }
    }

    return false
}
```

#### 2. `internal/filtering/http.go`

修改 `handleFilteringAddURL` 函数：
```go
// 新代码 - 只检查相同类型的列表
if d.filterExistsInType(fj.URL, fj.Whitelist, fj.DnsRouting) {
    err = errFilterExists
    aghhttp.ErrorAndLog(
        ctx,
        l,
        r,
        w,
        http.StatusBadRequest,
        "Filter with URL %q: %s",
        fj.URL,
        err,
    )

    return
}
```

## 行为变化

### 修复前
- ❌ 同一个URL不能在不同类型的过滤器中使用
- ❌ 例如：`https://example.com/rules.txt` 在黑名单中使用后，不能在DNS路由中使用

### 修复后
- ✅ 同一个URL可以在不同类型的过滤器中使用
- ✅ 例如：`https://example.com/rules.txt` 可以同时在黑名单和DNS路由中使用
- ✅ 但同一个URL不能在同一类型的过滤器中重复添加

## 使用场景

这个修复允许以下场景：

### 场景1：同一规则文件用于不同目的
```yaml
filters:
  - url: https://example.com/rules.txt
    name: 黑名单规则
    id: 1

dns_routing_filters:
  - url: https://example.com/rules.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"
```

### 场景2：测试不同配置
在测试时，可以将同一个规则文件添加到不同的列表中，测试不同的行为。

## 重复检查逻辑

### 黑名单过滤器
- 只检查 `filters` 列表
- 不检查 `whitelist_filters` 和 `dns_routing_filters`

### 白名单过滤器
- 只检查 `whitelist_filters` 列表
- 不检查 `filters` 和 `dns_routing_filters`

### DNS路由过滤器
- 只检查 `dns_routing_filters` 列表
- 不检查 `filters` 和 `whitelist_filters`

## 测试步骤

### 1. 测试跨类型URL使用
1. 添加一个黑名单规则：`https://example.com/rules.txt`
2. 添加一个DNS路由规则，使用相同URL：`https://example.com/rules.txt`
3. 验证：两个规则都成功添加

### 2. 测试同类型URL重复
1. 添加一个DNS路由规则：`https://example.com/rules.txt`
2. 尝试再次添加相同URL的DNS路由规则
3. 验证：第二次添加失败，提示URL已存在

### 3. 测试配置文件
检查 `AdGuardHome.yaml`：
```yaml
filters:
  - enabled: true
    url: https://example.com/rules.txt
    name: 黑名单规则
    id: 1

dns_routing_filters:
  - enabled: true
    url: https://example.com/rules.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"
```

## 注意事项

1. **ID唯一性**：即使URL相同，每个过滤器仍然需要唯一的ID
2. **规则内容**：同一个URL的规则文件内容是相同的，但在不同列表中的作用不同
3. **更新同步**：如果URL指向的规则文件更新，所有使用该URL的过滤器都会更新
4. **性能影响**：使用相同URL的多个过滤器会多次下载和解析相同的文件

## 编译和部署

```bash
# 编译后端
go build -o AdGuardHome.exe

# 编译前端（可选，前端没有改动）
cd client
npm run build-prod

# 运行
./AdGuardHome.exe
```

## 相关文件

- `internal/filtering/filter.go` - 添加 `filterExistsInType` 函数
- `internal/filtering/http.go` - 修改重复检查逻辑

## 技术细节

### 函数签名
```go
func (d *DNSFilter) filterExistsInType(url string, isWhitelist bool, isDnsRouting bool) (ok bool)
```

### 参数说明
- `url`: 要检查的过滤器URL
- `isWhitelist`: 是否为白名单过滤器
- `isDnsRouting`: 是否为DNS路由过滤器

### 返回值
- `true`: URL在指定类型的过滤器列表中已存在
- `false`: URL在指定类型的过滤器列表中不存在

### 类型判断逻辑
```
isDnsRouting=true  → 检查 DnsRoutingFilters
isDnsRouting=false, isWhitelist=true  → 检查 WhitelistFilters
isDnsRouting=false, isWhitelist=false → 检查 Filters
```
