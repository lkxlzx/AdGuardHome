# 完整修复总结

## 修复的问题

### 问题 1：白名单操作导致 DNS 路由规则数据丢失 ✅

**症状**：添加或编辑白名单规则后，配置文件中 `dns_routing_filters` 的 `upstream_group` 字段被删除。

**修复**：`internal/filtering/filtering.go` - WriteDiskConfig 函数
```go
c.DnsRoutingFilters = slices.Clone(d.conf.DnsRoutingFilters)  // 添加此行
```

---

### 问题 2：添加 Clash 规则时未处理 ✅

**症状**：在域名分流规则页面添加 Clash 规则 URL 时，保存的文件包含原始内容（包括 IP 规则）。

**修复**：`internal/filtering/http.go` - handleFilteringAddURL 函数
```go
filt := FilterYAML{
    // ...
    dnsRouting: fj.DnsRouting,  // 添加此行
    // ...
}
```

---

### 问题 3：重启后启用规则时未处理 ✅

**症状**：重启程序后，点击启用 DNS 路由规则，文件被重新下载为原始内容（包含 IP 规则）。

**原因**：`dnsRouting` 是内部标志，不保存到配置文件。重启后加载配置时，这个标志没有被设置。

**修复 A**：`internal/filtering/filter.go` - 添加公共方法
```go
// MarkAsDnsRouting marks this filter as a DNS routing filter.
func (filter *FilterYAML) MarkAsDnsRouting() {
    filter.dnsRouting = true
}
```

**修复 B**：`internal/home/home.go` - 加载配置时设置标志
```go
conf.DnsRoutingFilters = slices.Clone(config.DnsRoutingFilters)

// Mark DNS routing filters with internal flag
for i := range conf.DnsRoutingFilters {
    conf.DnsRoutingFilters[i].MarkAsDnsRouting()
}
```

---

### 问题 4：禁用再启用 DNS 路由规则时 upstream_group 丢失 ✅

**症状**：在 DNS 路由规则页面，禁用再启用规则后，原来设置的 DNS 分组（upstream_group）会丢失。

**原因**：前端在切换启用/禁用状态时，只发送了 `name`, `url`, `enabled` 字段，没有包含 `upstreamGroup` 字段。

**修复**：`client/src/components/Filters/Table.tsx` - renderCheckbox 方法
```typescript
// 修复前
const data = { name, url, enabled: !enabled };

// 修复后
const data = showUpstreamGroup 
    ? { name, url, enabled: !enabled, upstreamGroup }
    : { name, url, enabled: !enabled };
```

这样只有 DNS 路由规则页面（`showUpstreamGroup={true}`）才会包含 `upstreamGroup` 字段，不影响黑名单和白名单页面。

---

## 修改统计

| 文件 | 修改类型 | 行数 |
|------|---------|------|
| `internal/filtering/filtering.go` | 修改 | 1 |
| `internal/filtering/http.go` | 修改 | 1 |
| `internal/filtering/filter.go` | 新增 | 6 |
| `internal/home/home.go` | 新增 | 5 |
| `client/src/components/Filters/Table.tsx` | 修改 | 4 |
| **总计** | | **17** |

---

## 测试场景

### 场景 1：白名单不影响 DNS 路由规则

1. 添加 DNS 路由规则（带 upstream_group）
2. 添加白名单规则
3. ✅ 检查配置文件，`dns_routing_filters` 的 `upstream_group` 仍然存在

### 场景 2：添加 Clash 规则自动处理

1. 添加 Clash 规则 URL 到 DNS 路由规则
2. ✅ 检查 `data/filters/[id].txt`，只包含域名规则
3. ✅ 查看日志，看到 "clash rule processed" 信息

### 场景 3：重启后启用规则正常处理

1. 添加 Clash 规则并保存
2. 重启程序
3. 禁用该规则
4. 再次启用该规则
5. ✅ 检查 `data/filters/[id].txt`，仍然只包含域名规则（没有 IP 规则）

### 场景 4：禁用再启用规则保留 upstream_group

1. 添加 DNS 路由规则（选择上游组）
2. 记录配置文件中的 `upstream_group` 值
3. 禁用该规则
4. 再次启用该规则
5. ✅ 检查配置文件，`upstream_group` 仍然存在且值不变

---

## Clash 规则处理流程

### 检测条件

URL 满足以下任一条件即被识别为 Clash 规则：
- 包含 `/Clash/` 或 `/clash/`
- 以 `.yaml` 结尾
- 以 `.yml` 结尾

### 处理逻辑

1. **下载** Clash 规则文件
2. **解析** YAML 格式
3. **过滤** IP 规则（IP-CIDR, IP-CIDR6）
4. **保留** 域名规则（DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD）
5. **转换** 为 AdGuard 格式
6. **保存** 到本地文件

### 规则转换

| Clash 格式 | AdGuard 格式 | 说明 |
|-----------|-------------|------|
| `DOMAIN,google.com` | `google.com` | 精确匹配 |
| `DOMAIN-SUFFIX,google.com` | `\|\|google.com^` | 后缀匹配 |
| `DOMAIN-KEYWORD,google` | `*google*` | 关键词匹配 |
| `IP-CIDR,192.168.0.0/16` | ❌ 过滤掉 | IP 规则 |

---

## 编译和部署

### 编译

```bash
go build -o AdGuardHome.exe
```

✅ 编译成功，无错误

### 部署步骤

1. 停止旧版本 AdGuard Home
2. 备份配置文件 `AdGuardHome.yaml`
3. 替换可执行文件 `AdGuardHome.exe`
4. 启动新版本
5. 验证功能正常

---

## 验证清单

- [ ] 白名单操作不影响 DNS 路由规则
- [ ] 添加 Clash 规则自动处理
- [ ] 重启后启用规则正常处理
- [ ] 禁用再启用规则保留 upstream_group
- [ ] 配置文件正确保存
- [ ] DNS 查询使用正确的上游组
- [ ] 日志显示处理信息

---

## 相关文档

- `WHITELIST_DNS_ROUTING_SEPARATION_FIX.md` - 白名单分离修复详情
- `CLASH_RULES_FIX_COMPLETE.md` - Clash 规则处理修复详情
- `CLASH_RULES_TEST_GUIDE.md` - Clash 规则测试指南
- `QUICK_TEST_STEPS.md` - 快速测试步骤

---

## 总结

通过修改 5 个文件共 17 行代码，成功修复了四个关键问题：

1. ✅ 白名单操作不再影响 DNS 路由规则数据
2. ✅ 添加 Clash 规则时自动处理
3. ✅ 重启后启用规则时正确处理
4. ✅ 禁用再启用规则时保留 upstream_group

所有修复已完成并编译成功，可以开始测试部署！
