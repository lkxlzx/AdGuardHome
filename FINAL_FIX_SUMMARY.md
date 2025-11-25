# 最终修复总结

## 修复的问题

本次修复解决了两个关键问题，涉及 2 个文件的修改。

### 问题 1：白名单操作导致 DNS 路由规则数据丢失 ✅

**症状**：
- 添加或编辑白名单规则后
- 配置文件中 `dns_routing_filters` 的 `upstream_group` 字段被删除

**根本原因**：
`internal/filtering/filtering.go` 中的 `WriteDiskConfig` 函数在保存配置时，只复制了 `Filters` 和 `WhitelistFilters`，遗漏了 `DnsRoutingFilters`。

**修复代码**：
```go
// 修复前
func (d *DNSFilter) WriteDiskConfig(c *Config) {
    // ...
    c.Filters = slices.Clone(d.conf.Filters)
    c.WhitelistFilters = slices.Clone(d.conf.WhitelistFilters)
    c.UserRules = slices.Clone(d.conf.UserRules)
}

// 修复后
func (d *DNSFilter) WriteDiskConfig(c *Config) {
    // ...
    c.Filters = slices.Clone(d.conf.Filters)
    c.WhitelistFilters = slices.Clone(d.conf.WhitelistFilters)
    c.DnsRoutingFilters = slices.Clone(d.conf.DnsRoutingFilters)  // ✅ 添加此行
    c.UserRules = slices.Clone(d.conf.UserRules)
}
```

**影响文件**：
- `internal/filtering/filtering.go` (1 行修改)

**测试验证**：
1. 添加 DNS 路由规则（带 upstream_group）
2. 添加或编辑白名单规则
3. 检查配置文件，确认 `dns_routing_filters` 的 `upstream_group` 字段仍然存在

---

### 问题 2：域名分流规则的 Clash 规则处理 ✅

**需求**：
在域名分流规则页面添加的 Clash 规则 URL 应该自动处理：
- 过滤掉 IP 规则（IP-CIDR, IP-CIDR6）
- 只保留域名规则（DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD）
- 转换为 AdGuard Home 格式

**问题**：
添加 DNS 路由规则时，`dnsRouting` 标志没有被设置，导致 Clash 规则处理逻辑不会被触发。

**修复代码** (`internal/filtering/http.go`):
```go
// 修复前
filt := FilterYAML{
    Enabled: true,
    URL:     fj.URL,
    Name:    fj.Name,
    white:   fj.Whitelist,
    // ❌ 缺少 dnsRouting 字段
    Filter: Filter{
        ID:            d.idGen.next(),
        UpstreamGroup: fj.UpstreamGroup,
    },
}

// 修复后
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

**影响文件**：
- `internal/filtering/http.go` (1 行修改)

**测试验证**：
1. 添加 Clash 规则 URL 到 DNS 路由规则
2. 检查生成的规则文件（`data/filters/[id].txt`）
3. 确认文件中只包含域名规则，没有 IP 规则
4. 查看日志确认看到 "clash rule processed" 信息

**关键代码位置**：

1. **Clash 规则检测** (`internal/filtering/clash_rules.go`):
```go
func IsClashRuleURL(url string) bool {
    return strings.Contains(url, "/Clash/") ||
           strings.Contains(url, "/clash/") ||
           strings.HasSuffix(url, ".yaml") ||
           strings.HasSuffix(url, ".yml")
}
```

2. **Clash 规则处理** (`internal/filtering/clash_rules.go`):
```go
func ProcessClashRuleFile(input io.Reader, output io.Writer) (*ClashRuleStats, error) {
    domains, stats, err := ParseClashRulesFromReader(input)
    if err != nil {
        return nil, err
    }
    
    // Write domain rules to output (one per line)
    for _, domain := range domains {
        _, err := fmt.Fprintln(output, domain)
        if err != nil {
            return nil, fmt.Errorf("writing domain: %w", err)
        }
    }
    
    return stats, nil
}
```

3. **自动调用** (`internal/filtering/filter.go`):
```go
func (d *DNSFilter) updateIntl(ctx context.Context, flt *FilterYAML) (ok bool, err error) {
    // ...
    isDomainRoutingRule := flt.dnsRouting
    isClashRule := IsClashRuleURL(flt.URL)
    
    if isDomainRoutingRule && isClashRule {
        // Process Clash rules: filter out IP rules and keep only domain rules
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
}
```

**规则转换示例**：
| Clash 格式 | AdGuard 格式 |
|-----------|-------------|
| `DOMAIN,google.com` | `google.com` |
| `DOMAIN-SUFFIX,google.com` | `\|\|google.com^` |
| `DOMAIN-KEYWORD,google` | `*google*` |
| `IP-CIDR,192.168.0.0/16` | ❌ 过滤掉 |

---

## 编译状态

✅ **前端编译成功**
```bash
cd client
npm run build-prod
# webpack 5.102.1 compiled successfully
```

✅ **后端编译成功**
```bash
go build -o AdGuardHome.exe
# 编译成功，无错误
```

---

## 测试清单

### 白名单与 DNS 路由规则分离测试

- [ ] 添加 DNS 路由规则（带 upstream_group）
- [ ] 记录 `dns_routing_filters` 的内容
- [ ] 添加白名单规则
- [ ] 检查配置文件，确认 `dns_routing_filters` 未被影响
- [ ] 编辑白名单规则
- [ ] 再次检查配置文件，确认 `dns_routing_filters` 未被影响
- [ ] 删除白名单规则
- [ ] 最后检查配置文件，确认 `dns_routing_filters` 未被影响

### Clash 规则处理测试

- [ ] 添加 Clash 规则 URL 到 DNS 路由规则
  - 测试 URL: `https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml`
- [ ] 选择上游组
- [ ] 保存并等待规则下载
- [ ] 查看日志，确认看到 "clash rule processed" 信息
- [ ] 检查规则文件（`data/filters/[id].txt`），确认只包含域名规则
- [ ] 测试 DNS 查询，确认使用指定的上游组

### 配置持久化测试

- [ ] 完成上述所有操作
- [ ] 重启 AdGuard Home
- [ ] 检查所有配置是否正确加载
- [ ] 检查 DNS 路由规则是否正常工作

---

## 配置文件示例

修复后的正确配置文件结构：

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
    name: CN域名
    id: 1764000553
    upstream_group: group_1763970331409  # ✅ 此字段不会丢失
```

---

## 关键改进

1. **数据完整性**：三种过滤器类型完全独立，互不干扰
2. **自动化处理**：Clash 规则自动检测和转换，无需手动操作
3. **性能优化**：IP 规则被过滤掉，减少内存占用
4. **日志完善**：详细的处理日志，便于调试和监控

---

## 文档

- `WHITELIST_DNS_ROUTING_SEPARATION_FIX.md` - 白名单分离修复详情
- `CLASH_RULES_TEST_GUIDE.md` - Clash 规则测试指南
- `FINAL_FIX_SUMMARY.md` - 本文档

---

## 总结

✅ **问题 1 已修复**：白名单操作不再影响 DNS 路由规则数据
✅ **问题 2 已实现**：Clash 规则自动检测和处理功能完整
✅ **编译成功**：前端和后端都已成功编译
✅ **准备测试**：所有功能已就绪，可以开始测试

**下一步**：
1. 重启 AdGuard Home
2. 按照测试清单进行验证
3. 如有问题，查看日志获取详细信息
