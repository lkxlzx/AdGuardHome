# 配置文件作为数据源修复

## 问题

DNS 路由规则列表为空，因为：
1. 移除了 metadata 的加载逻辑
2. API 仍然从 File Manager 的内存读取规则
3. File Manager 的内存中没有规则数据

## 解决方案

**让 API 直接从配置文件读取规则，而不是从 File Manager 读取。**

### 修改内容

#### 1. handleGetDnsRoutingRules
**之前**：从 File Manager 读取
```go
domainListRules := globalContext.dnsRoutingFileManager.GetAllDomainListRules()
```

**现在**：从配置文件读取
```go
config.RLock()
filters := config.Filters
config.RUnlock()

for _, filter := range filters {
    if !filter.DnsRouting {
        continue
    }
    // 转换为响应格式
}
```

#### 2. handleAddDnsRoutingRule
添加规则时：
1. 使用 File Manager 下载并解析规则文件
2. **同时更新配置文件**
3. 保存配置

```go
// 1. File Manager 下载
err = globalContext.dnsRoutingFileManager.AddDomainListRule(ctx, rule)

// 2. 更新配置文件
config.Lock()
newFilter := filtering.FilterYAML{
    Enabled:        true,
    URL:            req.URL,
    Name:           req.Name,
    RulesCount:     rule.RulesCount,
    LastUpdated:    rule.LastUpdated,
    DnsRouting:     true,
    UpstreamGroup:  req.UpstreamGroup,
    Priority:       req.Priority,
}
newFilter.ID = rules.ListID(newID)
config.Filters = append(config.Filters, newFilter)
config.Unlock()

// 3. 保存配置
web.saveConfigIfNeeded(ctx, l, r, w)
```

#### 3. handleUpdateDnsRoutingRule
更新规则时：
1. 使用 File Manager 重新下载（如果 URL 改变）
2. **更新配置文件中的对应条目**
3. 保存配置

```go
// 1. File Manager 更新
err = globalContext.dnsRoutingFileManager.UpdateDomainListRule(ctx, rule)

// 2. 更新配置文件
config.Lock()
for i := range config.Filters {
    if int64(config.Filters[i].ID) == req.ID && config.Filters[i].DnsRouting {
        config.Filters[i].Name = req.Name
        config.Filters[i].URL = req.URL
        config.Filters[i].UpstreamGroup = req.UpstreamGroup
        config.Filters[i].Priority = req.Priority
        config.Filters[i].Enabled = req.Enabled
        config.Filters[i].RulesCount = rule.RulesCount
        config.Filters[i].LastUpdated = rule.LastUpdated
        break
    }
}
config.Unlock()

// 3. 保存配置
web.saveConfigIfNeeded(ctx, l, r, w)
```

#### 4. handleDeleteDnsRoutingRule
删除规则时：
1. 使用 File Manager 删除规则文件
2. **从配置文件中移除条目**
3. 保存配置

```go
// 1. File Manager 删除文件
err = globalContext.dnsRoutingFileManager.DeleteDomainListRule(ctx, req.ID)

// 2. 从配置文件移除
config.Lock()
newFilters := make([]filtering.FilterYAML, 0, len(config.Filters))
for _, filter := range config.Filters {
    if int64(filter.ID) != req.ID || !filter.DnsRouting {
        newFilters = append(newFilters, filter)
    }
}
config.Filters = newFilters
config.Unlock()

// 3. 保存配置
web.saveConfigIfNeeded(ctx, l, r, w)
```

### 数据流

#### 获取规则列表
```
API 请求
  ↓
读取 config.Filters
  ↓
过滤 dns_routing=true 的条目
  ↓
返回规则列表
```

#### 添加规则
```
API 请求
  ↓
File Manager: 下载并解析规则文件
  ↓
更新 config.Filters（添加新条目）
  ↓
保存配置文件
  ↓
返回响应
```

#### 更新规则
```
API 请求
  ↓
File Manager: 重新下载规则文件（如果 URL 改变）
  ↓
更新 config.Filters（修改对应条目）
  ↓
保存配置文件
  ↓
返回响应
```

#### 删除规则
```
API 请求
  ↓
File Manager: 删除规则文件
  ↓
更新 config.Filters（移除条目）
  ↓
保存配置文件
  ↓
返回响应
```

### 类型转换

由于 `filter.ID` 是 `rules.ListID` 类型（uint64），需要进行类型转换：

```go
// 读取时：ListID -> int64
ID: int64(filter.ID)

// 比较时
if int64(filter.ID) == req.ID

// 写入时：int64 -> ListID
newFilter.ID = rules.ListID(newID)
```

### 编译测试

```powershell
go build -o AdGuardHome_fixed.exe
```
✅ 编译成功，无错误

## 优势

1. **配置文件是唯一数据源** - 所有规则配置都在 AdGuardHome.yaml 中
2. **API 直接读取配置** - 不依赖 File Manager 的内存状态
3. **操作同步更新配置** - 添加/更新/删除规则时立即更新配置文件
4. **用户友好** - 用户可以直接编辑配置文件

## 测试

### 功能测试

1. **获取规则列表**
   ```bash
   curl http://localhost/control/dns_routing/rules
   ```
   应该返回配置文件中的所有 DNS 路由规则

2. **添加规则**
   ```bash
   curl -X POST http://localhost/control/dns_routing/add \
     -d '{"name":"Test","url":"https://example.com/rules.txt","upstream_group":"xxx","priority":0}'
   ```
   应该下载规则文件并更新配置文件

3. **更新规则**
   ```bash
   curl -X POST http://localhost/control/dns_routing/update \
     -d '{"id":123,"name":"Updated","url":"...","enabled":false}'
   ```
   应该更新配置文件中的对应条目

4. **删除规则**
   ```bash
   curl -X POST http://localhost/control/dns_routing/delete \
     -d '{"id":123}'
   ```
   应该删除规则文件并从配置文件中移除

## 总结

这次修复解决了"读取不到配置文件内容"的问题，通过：
- ✅ API 直接从配置文件读取规则
- ✅ 所有操作同步更新配置文件
- ✅ 配置文件是唯一的权威数据源
- ✅ File Manager 只负责文件操作

现在 DNS 路由规则列表应该能正确显示配置文件中的规则了。
