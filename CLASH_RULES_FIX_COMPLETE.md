# Clash 规则处理修复完成

## 问题描述

在域名分流规则页面添加 Clash 规则 URL 时，保存的文件包含原始内容（包括 IP 规则），而不是过滤后的域名规则。

## 根本原因

在 `internal/filtering/http.go` 的 `handleFilteringAddURL` 函数中，创建 `FilterYAML` 结构时没有设置 `dnsRouting` 字段，导致 `updateIntl` 函数中的 Clash 规则处理逻辑不会被触发。

### 代码流程分析

1. **前端发送请求**：
   ```json
   {
     "name": "CN域名",
     "url": "https://.../China_Classical.yaml",
     "dns_routing": true,
     "upstream_group": "group_xxx"
   }
   ```

2. **后端接收** (`handleFilteringAddURL`):
   ```go
   fj := filterAddJSON{}
   json.NewDecoder(r.Body).Decode(&fj)
   // fj.DnsRouting = true ✅
   ```

3. **创建 FilterYAML** (修复前):
   ```go
   filt := FilterYAML{
       Enabled: true,
       URL:     fj.URL,
       Name:    fj.Name,
       white:   fj.Whitelist,
       // ❌ 缺少 dnsRouting: fj.DnsRouting
       Filter: Filter{
           ID:            d.idGen.next(),
           UpstreamGroup: fj.UpstreamGroup,
       },
   }
   ```

4. **下载过滤器** (`update` → `updateIntl`):
   ```go
   isDomainRoutingRule := flt.dnsRouting  // ❌ false (未设置)
   isClashRule := IsClashRuleURL(flt.URL) // ✅ true
   
   if isDomainRoutingRule && isClashRule {
       // ❌ 不会执行，因为 isDomainRoutingRule = false
       ProcessClashRuleFile(r, tmpFile)
   }
   ```

5. **结果**：
   - 原始 Clash 规则文件被直接保存
   - IP 规则没有被过滤掉

## 修复方案

在 `handleFilteringAddURL` 中创建 `FilterYAML` 时，添加 `dnsRouting` 字段：

```go
filt := FilterYAML{
    Enabled:    true,
    URL:        fj.URL,
    Name:       fj.Name,
    white:      fj.Whitelist,
    dnsRouting: fj.DnsRouting,  // ✅ 添加此行
    Filter: Filter{
        ID:            d.idGen.next(),
        UpstreamGroup: fj.UpstreamGroup,
    },
}
```

## 修复后的流程

1. **前端发送请求** (同上)

2. **后端接收** (同上)

3. **创建 FilterYAML** (修复后):
   ```go
   filt := FilterYAML{
       Enabled:    true,
       URL:        fj.URL,
       Name:       fj.Name,
       white:      fj.Whitelist,
       dnsRouting: fj.DnsRouting,  // ✅ true
       Filter: Filter{
           ID:            d.idGen.next(),
           UpstreamGroup: fj.UpstreamGroup,
       },
   }
   ```

4. **下载过滤器** (`update` → `updateIntl`):
   ```go
   isDomainRoutingRule := flt.dnsRouting  // ✅ true
   isClashRule := IsClashRuleURL(flt.URL) // ✅ true
   
   if isDomainRoutingRule && isClashRule {
       // ✅ 会执行
       stats, err := ProcessClashRuleFile(r, tmpFile)
       // 过滤掉 IP 规则，只保留域名规则
       
       d.logger.InfoContext(ctx, "clash rule processed",
           "id", flt.ID,
           "total_rules", stats.TotalRules,
           "valid_domains", stats.ValidDomains,
           "filtered_ip_rules", stats.IPRules,
       )
   }
   ```

5. **结果**：
   - ✅ Clash 规则被正确处理
   - ✅ IP 规则被过滤掉
   - ✅ 只保留域名规则
   - ✅ 转换为 AdGuard 格式

## 其他场景验证

### 编辑过滤器

`handleFilteringSetURL` → `filterSetProperties` → `update`

- ✅ `filterSetProperties` 直接修改已存在的过滤器
- ✅ 已存在的过滤器已经有正确的 `dnsRouting` 值
- ✅ 无需修改

### 刷新过滤器

`handleFilteringRefresh` → `tryRefreshFilters` → `updateIntl`

- ✅ 使用已存在的过滤器列表
- ✅ 已存在的过滤器已经有正确的 `dnsRouting` 值
- ✅ 无需修改

### 启动时加载

`loadFilters` → `load`

- ✅ `load` 函数只读取已存在的文件
- ✅ 不涉及下载和处理
- ✅ 无需修改

## 测试验证

### 测试步骤

1. **添加 Clash 规则**：
   ```
   URL: https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml
   名称: CN域名
   上游组: 国内DNS
   ```

2. **等待下载完成**

3. **检查规则文件** (`data/filters/[id].txt`):
   ```
   google.com
   ||baidu.com^
   ||qq.com^
   *taobao*
   ```

4. **验证没有 IP 规则**：
   - ❌ 不应该包含 `192.168.0.0/16`
   - ❌ 不应该包含 `IP-CIDR`

5. **查看日志**：
   ```
   [INFO] clash rule processed id=xxx total_rules=1000 valid_domains=800 filtered_ip_rules=200
   ```

### 预期结果

- ✅ 规则文件只包含域名规则
- ✅ IP 规则被过滤掉
- ✅ 日志显示处理统计信息
- ✅ DNS 查询使用指定的上游组

## 额外修复：重启后启用规则的问题

### 问题描述

重启程序后，如果操作启用 DNS 路由规则，会导致文件被重新下载为原始内容（包含 IP 规则），而不是处理后的内容。

### 根本原因

`dnsRouting` 是内部标志，不会保存到配置文件。程序重启后从配置文件加载过滤器时，这个标志没有被设置，导致启用操作触发的更新不会进行 Clash 规则处理。

### 修复方案

**1. 添加公共方法** (`internal/filtering/filter.go`):
```go
// MarkAsDnsRouting marks this filter as a DNS routing filter.
func (filter *FilterYAML) MarkAsDnsRouting() {
    filter.dnsRouting = true
}
```

**2. 在加载配置时设置标志** (`internal/home/home.go`):
```go
conf.DnsRoutingFilters = slices.Clone(config.DnsRoutingFilters)

// Mark DNS routing filters with internal flag
for i := range conf.DnsRoutingFilters {
    conf.DnsRoutingFilters[i].MarkAsDnsRouting()
}
```

这确保了程序重启后，DNS 路由过滤器的 `dnsRouting` 标志被正确设置，后续的启用/禁用操作会正确触发 Clash 规则处理。

## 修改文件

- `internal/filtering/http.go` (1 行修改) - 添加过滤器时设置 dnsRouting
- `internal/filtering/filter.go` (6 行新增) - 添加 MarkAsDnsRouting 方法
- `internal/home/home.go` (5 行新增) - 加载配置时设置 dnsRouting 标志

## 编译状态

✅ 编译成功，无错误

## 总结

通过在 `handleFilteringAddURL` 中添加一行代码 `dnsRouting: fj.DnsRouting`，成功修复了 Clash 规则处理问题。现在添加 Clash 规则 URL 时，系统会自动：

1. 检测 URL 是否为 Clash 规则
2. 下载并解析 Clash 规则
3. 过滤掉 IP 规则
4. 只保留域名规则
5. 转换为 AdGuard 格式
6. 保存到本地文件

修复简单但关键，确保了 Clash 规则处理功能的正常工作。
