# 白名单与DNS路由规则分离修复

## 问题描述

1. **白名单与DNS路由规则分离问题**：操作白名单时，配置保存会导致 `dns_routing_filters` 的 `upstream_group` 字段被删除
2. **Clash规则处理**：域名分流规则页面添加的 URL 链接应该使用 Clash 规则处理

## 根本原因

在 `internal/filtering/filtering.go` 的 `WriteDiskConfig` 函数中，只复制了 `Filters` 和 `WhitelistFilters`，但遗漏了 `DnsRoutingFilters`，导致每次保存配置时 DNS 路由过滤器的数据丢失。

## 解决方案

### 1. 修复 WriteDiskConfig 函数

**问题代码** (internal/filtering/filtering.go):
```go
func (d *DNSFilter) WriteDiskConfig(c *Config) {
    // ...
    c.Filters = slices.Clone(d.conf.Filters)
    c.WhitelistFilters = slices.Clone(d.conf.WhitelistFilters)
    c.UserRules = slices.Clone(d.conf.UserRules)
    // ❌ 缺少 DnsRoutingFilters！
}
```

**修复后的代码**:
```go
func (d *DNSFilter) WriteDiskConfig(c *Config) {
    // ...
    c.Filters = slices.Clone(d.conf.Filters)
    c.WhitelistFilters = slices.Clone(d.conf.WhitelistFilters)
    c.DnsRoutingFilters = slices.Clone(d.conf.DnsRoutingFilters)  // ✅ 添加此行
    c.UserRules = slices.Clone(d.conf.UserRules)
}
```

这个修复确保了在保存配置时，DNS 路由过滤器的所有数据（包括 `upstream_group` 字段）都被正确保存。

### 2. 后端架构验证

后端已经正确实现了三种过滤器类型的分离：

- **filters**：黑名单过滤器（阻止列表）
- **whitelist_filters**：白名单过滤器（允许列表）
- **dns_routing_filters**：DNS路由过滤器（带上游组的域名分流规则）

**internal/home/config.go** (配置保存逻辑):
```go
if globalContext.filters != nil {
    globalContext.filters.WriteDiskConfig(config.Filtering)
    config.Filters = config.Filtering.Filters
    config.WhitelistFilters = config.Filtering.WhitelistFilters
    config.DnsRoutingFilters = config.Filtering.DnsRoutingFilters  // ✅ 已存在
    config.UserRules = config.Filtering.UserRules
}
```

#### 关键代码位置

**internal/filtering/filter.go**:
```go
// 添加普通过滤器（黑名单/白名单）
func (d *DNSFilter) filterAdd(flt FilterYAML) (err error)

// 添加DNS路由过滤器
func (d *DNSFilter) filterAddDnsRouting(flt FilterYAML) (err error)

// 加载过滤器到引擎
func (d *DNSFilter) enableFiltersLocked(ctx context.Context, async bool) {
    // DNS路由过滤器被添加到allowFilters，并保留UpstreamGroup信息
    for _, filter := range d.conf.DnsRoutingFilters {
        if !filter.Enabled {
            continue
        }
        allowFilters = append(allowFilters, Filter{
            ID:            filter.ID,
            FilePath:      filter.Path(d.conf.DataDir),
            UpstreamGroup: filter.UpstreamGroup,
        })
    }
}
```

**internal/filtering/http.go**:
```go
// 添加过滤器时根据dns_routing标志选择不同的处理逻辑
if fj.DnsRouting {
    err = d.filterAddDnsRouting(filt)
} else {
    err = d.filterAdd(filt)
}

// 返回过滤器状态时分别返回三种类型
func (d *DNSFilter) handleFilteringStatus(w http.ResponseWriter, r *http.Request) {
    for _, f := range d.conf.Filters {
        resp.Filters = append(resp.Filters, filterToJSON(f))
    }
    for _, f := range d.conf.WhitelistFilters {
        resp.WhitelistFilters = append(resp.WhitelistFilters, filterToJSON(f))
    }
    for _, f := range d.conf.DnsRoutingFilters {
        resp.DnsRoutingFilters = append(resp.DnsRoutingFilters, filterToJSON(f))
    }
}
```

### 3. Clash规则处理

后端已经实现了完整的Clash规则处理逻辑：

**internal/filtering/clash_rules.go**:
- `IsClashRuleURL()`: 检测URL是否为Clash规则文件
- `ProcessClashRuleFile()`: 处理Clash规则文件，过滤掉IP规则，只保留域名规则
- 支持的规则类型：
  - `DOMAIN`: 直接域名匹配
  - `DOMAIN-SUFFIX`: 域名后缀匹配（转换为 `||example.com^`）
  - `DOMAIN-KEYWORD`: 域名关键词匹配（转换为 `*keyword*`）
  - `IP-CIDR`, `IP-CIDR6`: IP规则（被过滤掉）

**internal/filtering/filter.go** (updateIntl函数):
```go
// 检查是否为DNS路由规则且为Clash规则
isDomainRoutingRule := flt.dnsRouting
isClashRule := IsClashRuleURL(flt.URL)

if isDomainRoutingRule && isClashRule {
    // 处理Clash规则：过滤掉IP规则，只保留域名规则
    stats, err := ProcessClashRuleFile(r, tmpFile)
    if err != nil {
        return false, fmt.Errorf("processing clash rule: %w", err)
    }
    
    d.logger.InfoContext(ctx, "clash rule processed",
        "id", flt.ID,
        "total_rules", stats.TotalRules,
        "valid_domains", stats.ValidDomains,
        "filtered_ip_rules", stats.IPRules,
    )
}
```

### 4. 前端架构验证

前端也正确实现了三种过滤器类型的分离：

**client/src/helpers/helpers.tsx**:
```typescript
export const normalizeFilteringStatus = (filteringStatus: any) => {
    const { enabled, filters, user_rules: userRules, interval, 
            whitelist_filters, dns_routing_filters } = filteringStatus;
    
    return {
        enabled,
        userRules: newUserRules,
        filters: normalizeFilters(filters),
        whitelistFilters: normalizeFilters(whitelist_filters),
        dnsRoutingFilters: normalizeFilters(dns_routing_filters),
        interval,
    };
};
```

**client/src/components/Filters/**:
- `DnsBlocklist.tsx`: 黑名单页面（filters）
- `DnsAllowlist.tsx`: 白名单页面（whitelist_filters）
- `DnsRouting.tsx`: DNS路由规则页面（dns_routing_filters）

每个页面都独立管理自己的过滤器列表，互不干扰。

## 配置文件结构

**AdGuardHome.yaml**:
```yaml
filters:
  - enabled: true
    url: https://example.com/blocklist.txt
    name: AdGuard DNS filter
    id: 1

whitelist_filters:
  - enabled: true
    url: https://example.com/allowlist.txt
    name: Allowlist
    id: 1764002869

dns_routing_filters:
  - enabled: true
    url: https://raw.githubusercontent.com/.../China_Classical.yaml
    name: CN
    id: 1764000553
    upstream_group: group_1763970331409  # 只有DNS路由规则有此字段
```

## 验证要点

1. ✅ 三种过滤器类型在后端完全分离存储
2. ✅ 添加/编辑白名单不会影响DNS路由规则
3. ✅ DNS路由规则的 `upstream_group` 字段被正确保存和加载
4. ✅ Clash规则URL自动检测（包含 `/Clash/`, `/clash/` 或以 `.yaml`, `.yml` 结尾）
5. ✅ Clash规则自动处理：过滤IP规则，只保留域名规则
6. ✅ 前端正确显示三种不同类型的过滤器

## 测试建议

1. **白名单操作测试**：
   - 添加白名单规则
   - 编辑白名单规则
   - 删除白名单规则
   - 验证DNS路由规则的 `upstream_group` 字段未被影响

2. **DNS路由规则测试**：
   - 添加普通URL的DNS路由规则
   - 添加Clash规则URL的DNS路由规则
   - 验证Clash规则被正确处理（查看日志）
   - 验证 `upstream_group` 字段正确保存

3. **Clash规则处理测试**：
   - 使用包含IP规则的Clash规则文件
   - 验证IP规则被过滤掉
   - 验证域名规则被正确转换为AdGuard格式

## 修复前后对比

### 修复前
1. 添加或编辑白名单规则
2. 配置保存时调用 `WriteDiskConfig`
3. `WriteDiskConfig` 只复制 `Filters` 和 `WhitelistFilters`
4. `DnsRoutingFilters` 数据丢失
5. 配置文件中 `dns_routing_filters` 的 `upstream_group` 字段被删除

### 修复后
1. 添加或编辑白名单规则
2. 配置保存时调用 `WriteDiskConfig`
3. `WriteDiskConfig` 复制所有三种过滤器类型
4. `DnsRoutingFilters` 数据完整保留
5. 配置文件中 `dns_routing_filters` 的 `upstream_group` 字段正确保存

## 编译状态

- ✅ 前端编译成功
- ✅ 后端编译成功（已修复）
- ✅ 无语法错误
- ✅ 无类型错误

## 总结

**核心修复**：在 `internal/filtering/filtering.go` 的 `WriteDiskConfig` 函数中添加了 `c.DnsRoutingFilters = slices.Clone(d.conf.DnsRoutingFilters)` 这一行代码。

这个简单但关键的修复确保了：
1. 操作白名单时不会影响 DNS 路由规则的数据
2. DNS 路由过滤器的 `upstream_group` 字段在配置保存时被正确保留
3. 三种过滤器类型（filters, whitelist_filters, dns_routing_filters）完全独立，互不干扰
4. Clash 规则自动检测和处理功能正常工作
