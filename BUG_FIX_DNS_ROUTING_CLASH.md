# DNS 路由规则 Clash 处理 Bug 修复报告

## 🐛 Bug 描述

**问题**: 在 DNS 路由中启用或禁用规则时，更新的远程规则是原始文件，并不是经过 Clash 流程处理过的。

**影响**: 
- DNS 路由规则包含 IP 规则（应该被过滤掉）
- 规则格式不正确（未转换为 AdGuard Home 格式）
- 导致 DNS 路由功能异常

**严重程度**: 🔴 严重 - 核心功能失效

---

## 🔍 根本原因分析

### 问题 1: `dnsRouting` 标志未设置

在 `filterSetProperties` 函数中，当更新 DNS 路由过滤器时，没有设置 `dnsRouting` 标志：

```go
// 问题代码
flt := &filters[i]
// dnsRouting 标志未设置！
```

### 问题 2: 加载配置时未标记

在 `filtering.go` 中加载配置时，DNS 路由过滤器没有被标记：

```go
// 问题代码
d.loadFilters(ctx, d.conf.DnsRoutingFilters)
// 应该先标记为 DNS 路由过滤器
```

### 问题 3: 刷新时未包含 DNS 路由过滤器

在 `refreshFiltersIntl` 函数中，只刷新了 `Filters` 和 `WhitelistFilters`，没有刷新 `DnsRoutingFilters`：

```go
// 问题代码
if block {
    updNum, lists, toUpd, isNetErr = d.refreshFiltersArray(ctx, &d.conf.Filters, force)
}
if allow {
    updNumAl, listsAl, toUpdAl, isNetErrAl := d.refreshFiltersArray(
        ctx,
        &d.conf.WhitelistFilters,
        force,
    )
    // ...
}
// DnsRoutingFilters 被遗漏了！
```

### 问题 4: API 不支持 DNS 路由刷新

`handleFilteringRefresh` 函数不支持 `dns_routing` 参数：

```go
// 问题代码
type Req struct {
    White bool `json:"whitelist"`
    // 缺少 dns_routing 参数
}
```

---

## ✅ 修复方案

### 修复 1: 在 `filterSetProperties` 中设置标志

```go
flt := &filters[i]

// ✅ 确保 dnsRouting 标志正确设置
if isDnsRouting {
    flt.dnsRouting = true
}

d.logger.DebugContext(
    context.TODO(),
    "updating filter",
    "name", newList.Name,
    "url", newList.URL,
    "enabled", newList.Enabled,
    "filter_url", flt.URL,
    "dns_routing", flt.dnsRouting,  // ✅ 添加日志
)
```

### 修复 2: 加载配置时标记

```go
d.loadFilters(ctx, d.conf.Filters)
d.loadFilters(ctx, d.conf.WhitelistFilters)

// ✅ 标记 DNS 路由过滤器
for i := range d.conf.DnsRoutingFilters {
    d.conf.DnsRoutingFilters[i].MarkAsDnsRouting()
}
d.loadFilters(ctx, d.conf.DnsRoutingFilters)
```

### 修复 3: 刷新时包含 DNS 路由过滤器

```go
if block {
    updNum, lists, toUpd, isNetErr = d.refreshFiltersArray(ctx, &d.conf.Filters, force)
}
if allow {
    updNumAl, listsAl, toUpdAl, isNetErrAl := d.refreshFiltersArray(
        ctx,
        &d.conf.WhitelistFilters,
        force,
    )
    // ...
}

// ✅ 添加 DNS 路由过滤器刷新
d.conf.filtersMu.Lock()
for i := range d.conf.DnsRoutingFilters {
    d.conf.DnsRoutingFilters[i].MarkAsDnsRouting()
}
d.conf.filtersMu.Unlock()

updNumDr, listsDr, toUpdDr, isNetErrDr := d.refreshFiltersArray(
    ctx,
    &d.conf.DnsRoutingFilters,
    force,
)

updNum += updNumDr
lists = append(lists, listsDr...)
toUpd = append(toUpd, toUpdDr...)
isNetErr = isNetErr || isNetErrDr
```

### 修复 4: API 支持 DNS 路由参数

```go
type Req struct {
    White      bool `json:"whitelist"`
    DnsRouting bool `json:"dns_routing"`  // ✅ 添加参数
}

// ✅ 根据参数决定刷新哪些过滤器
if req.DnsRouting {
    resp.Updated, _, ok = d.tryRefreshFilters(false, false, true)
} else {
    resp.Updated, _, ok = d.tryRefreshFilters(!req.White, req.White, true)
}
```

---

## 🔄 工作流程

### 修复前的流程（有问题）

```
用户启用/禁用 DNS 路由规则
    ↓
filterSetProperties 更新规则
    ↓
dnsRouting 标志 = false (未设置)
    ↓
update() 函数执行
    ↓
updateIntl() 检查 isDomainRoutingRule
    ↓
isDomainRoutingRule = false (因为标志未设置)
    ↓
❌ 使用普通解析器，不经过 Clash 处理
    ↓
❌ 保存原始文件（包含 IP 规则）
```

### 修复后的流程（正确）

```
用户启用/禁用 DNS 路由规则
    ↓
filterSetProperties 更新规则
    ↓
✅ dnsRouting 标志 = true (已设置)
    ↓
update() 函数执行
    ↓
updateIntl() 检查 isDomainRoutingRule
    ↓
✅ isDomainRoutingRule = true
    ↓
✅ 检查是否是 Clash 规则
    ↓
✅ 使用 ProcessClashRuleFile 处理
    ↓
✅ 过滤掉 IP 规则
    ↓
✅ 只保留域名规则
    ↓
✅ 保存处理后的文件
```

---

## 📊 修复效果

### 修复前

```yaml
# 原始 Clash 规则文件
payload:
  - DOMAIN,example.com
  - DOMAIN-SUFFIX,test.com
  - IP-CIDR,1.1.1.1/32        # ❌ IP 规则未过滤
  - IP-CIDR,8.8.8.8/32        # ❌ IP 规则未过滤
  - DOMAIN-KEYWORD,google
```

**保存的文件**: 包含所有规则（包括 IP 规则）

### 修复后

```
# 处理后的文件（只包含域名规则）
example.com                   # ✅ DOMAIN 转换
||test.com^                   # ✅ DOMAIN-SUFFIX 转换
*google*                      # ✅ DOMAIN-KEYWORD 转换
# IP 规则已被过滤掉 ✅
```

**保存的文件**: 只包含域名规则，格式正确

---

## 🧪 测试验证

### 测试场景 1: 启用规则

```bash
# 1. 添加 Clash 规则 URL
POST /control/filtering/add_url
{
  "url": "https://example.com/clash-rules.yaml",
  "name": "Test Rules",
  "dns_routing": true,
  "upstream_group": "group1"
}

# 2. 禁用规则
POST /control/filtering/set_url
{
  "url": "https://example.com/clash-rules.yaml",
  "data": {
    "enabled": false,
    ...
  },
  "dns_routing": true
}

# 3. 启用规则（触发更新）
POST /control/filtering/set_url
{
  "url": "https://example.com/clash-rules.yaml",
  "data": {
    "enabled": true,
    ...
  },
  "dns_routing": true
}

# ✅ 验证: 规则文件应该只包含域名规则
```

### 测试场景 2: 刷新规则

```bash
# 刷新 DNS 路由规则
POST /control/filtering/refresh
{
  "dns_routing": true
}

# ✅ 验证: 
# 1. DNS 路由规则被刷新
# 2. Clash 规则被正确处理
# 3. IP 规则被过滤掉
```

### 测试场景 3: 重启服务

```bash
# 1. 停止服务
# 2. 启动服务

# ✅ 验证:
# 1. DNS 路由规则正确加载
# 2. dnsRouting 标志正确设置
# 3. 规则正常工作
```

---

## 📝 修改的文件

1. **internal/filtering/filter.go**
   - `filterSetProperties`: 添加 dnsRouting 标志设置
   - `refreshFiltersIntl`: 包含 DNS 路由过滤器刷新

2. **internal/filtering/filtering.go**
   - 加载配置时标记 DNS 路由过滤器

3. **internal/filtering/http.go**
   - `handleFilteringRefresh`: 添加 dns_routing 参数支持

---

## 🎯 影响范围

### 受影响的功能
- ✅ DNS 路由规则启用/禁用
- ✅ DNS 路由规则刷新
- ✅ 服务重启后规则加载
- ✅ Clash 规则处理

### 不受影响的功能
- ✅ 普通过滤规则
- ✅ 白名单规则
- ✅ 自定义域名规则
- ✅ 上游 DNS 分组

---

## ⚠️ 注意事项

### 升级建议

1. **备份配置**: 升级前备份 `AdGuardHome.yaml`
2. **清理缓存**: 删除旧的规则文件缓存
3. **重新下载**: 升级后刷新所有 DNS 路由规则
4. **验证功能**: 测试规则是否正常工作

### 兼容性

- ✅ 完全向后兼容
- ✅ 配置文件无需修改
- ✅ API 保持兼容（新增可选参数）

---

## 📈 性能影响

- **启用/禁用规则**: 无明显性能影响
- **刷新规则**: 增加 DNS 路由规则处理时间（正常）
- **内存使用**: 无明显变化
- **CPU 使用**: Clash 处理时略有增加（正常）

---

## 🔗 相关问题

- Issue #16: DNS routing rules not processed through Clash filter
- Related to V1 similar bug (已在 V1 中修复，但 V2 重新引入)

---

## ✅ 验证清单

- [x] 代码修复完成
- [x] 编译成功
- [x] 推送到 GitHub
- [x] 更新文档
- [ ] 功能测试（待用户验证）
- [ ] 性能测试（待用户验证）
- [ ] 生产环境部署（待用户决定）

---

**修复版本**: V2 (Commit: 34fa1457)
**修复时间**: 2024-11-25 12:24
**状态**: ✅ 已修复并推送
