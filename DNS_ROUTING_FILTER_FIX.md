# DNS 路由筛选器修复

## 问题描述

在查询日志页面中，当用户选择"DNS 路由"筛选器时，会出现以下错误：

```
Error: controlQuerylog/
search: "response_status=dns_routing"
Loading params: invalid value
dns_routing | #40
```

## 根本原因

前端代码在 `client/src/helpers/constants.ts` 中定义了 `DNS_ROUTING` 筛选器选项：

```typescript
DNS_ROUTING: {
    QUERY: 'dns_routing',
    LABEL: 'dns_routing',
}
```

但后端代码在 `internal/querylog/searchcriterion.go` 中的 `filteringStatusValues` 数组中没有包含 `dns_routing` 这个值，导致后端拒绝了这个筛选参数。

## 修复方案

在 `internal/querylog/searchcriterion.go` 文件中进行了以下修改：

### 1. 添加 DNS 路由常量

```go
const (
    // ... 其他常量
    filteringStatusDNSRouting = "dns_routing" // DNS routing
    // ... 其他常量
)
```

### 2. 更新支持的筛选状态列表

```go
var filteringStatusValues = []string{
    filteringStatusAll, filteringStatusFiltered, filteringStatusBlocked,
    filteringStatusBlockedService, filteringStatusBlockedSafebrowsing, filteringStatusBlockedParental,
    filteringStatusWhitelisted, filteringStatusDNSRouting, filteringStatusRewritten, filteringStatusSafeSearch,
    filteringStatusProcessed,
}
```

### 3. 实现 DNS 路由筛选逻辑

在 `ctFilteringStatusCase` 函数中添加了对 DNS 路由的处理：

```go
case filteringStatusDNSRouting:
    return reason == filtering.NotFilteredDNSRouting
```

这个逻辑会匹配所有 reason 为 `NotFilteredDNSRouting` 的查询记录，即通过 DNS 路由规则处理的查询。

## 编译和测试

### 编译修复后的版本

使用提供的构建脚本：

```powershell
.\build-dns-routing-fix.ps1
```

或手动编译：

```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o AdGuardHome_dns_routing_fix.exe
```

### 验证可执行文件

```powershell
.\AdGuardHome_dns_routing_fix.exe --version
.\AdGuardHome_dns_routing_fix.exe --check-config
```

### 测试 DNS 路由筛选功能

启动修复后的版本，然后使用测试脚本验证：

```powershell
.\test_dns_routing_filter.ps1
```

或在 Web 界面中：
1. 访问 http://localhost:3000
2. 进入"查询日志"页面
3. 在筛选器下拉菜单中选择"DNS 路由"
4. 确认不再出现错误，可以正常筛选 DNS 路由记录

## 影响范围

- 修复了查询日志页面中"DNS 路由"筛选器的功能
- 用户现在可以正常筛选和查看通过 DNS 路由规则处理的查询记录
- 不影响其他筛选器的功能

## 相关文件

- `internal/querylog/searchcriterion.go` - 后端筛选逻辑
- `client/src/helpers/constants.ts` - 前端筛选器定义
- `internal/filtering/result.go` - DNS 路由 Reason 定义
