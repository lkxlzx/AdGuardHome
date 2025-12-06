# DNS 路由规则立即生效修复

## 问题描述

添加、更新或删除 DNS 路由规则后，规则不能立即生效，需要禁用再启用才能生效。

## 根本原因

在添加、更新、删除 DNS 路由规则的 API 处理函数中，虽然更新了配置文件和 File Manager，但**没有通知 DNS 服务器重新加载路由规则**。DNS 服务器的 `dnsRouter` 还在使用旧的规则集。

## 修复方案

在以下操作完成后，立即调用 `ReloadDnsRouter()` 或 `RemoveDnsRoutingSource()` 来重新加载路由规则：

### 1. 添加自定义域名规则 (`handleAddCustomDomainRule`)
- 添加规则到 File Manager 后
- 调用 `globalContext.dnsServer.ReloadDnsRouter(ctx)` 重新加载路由器

### 2. 更新自定义域名规则 (`handleUpdateCustomDomainRule`)
- 更新规则到 File Manager 后
- 调用 `globalContext.dnsServer.ReloadDnsRouter(ctx)` 重新加载路由器

### 3. 删除自定义域名规则 (`handleDeleteCustomDomainRule`)
- 删除规则从 File Manager 后
- 调用 `globalContext.dnsServer.ReloadDnsRouter(ctx)` 重新加载路由器

### 4. 添加规则文件 (`handleAddDnsRoutingRule`)
- 添加规则文件并保存配置后
- 调用 `globalContext.dnsServer.ReloadDnsRouter(ctx)` 重新加载路由器

### 5. 更新规则文件 (`handleUpdateDnsRoutingRule`)
- 更新规则文件并保存配置后
- 调用 `globalContext.dnsServer.ReloadDnsRouter(ctx)` 重新加载路由器

### 6. 删除规则文件 (`handleDeleteDnsRoutingRule`)
- 删除规则文件并保存配置后
- 调用 `globalContext.dnsServer.RemoveDnsRoutingSource(req.ID)` 移除路由源

### 7. 刷新规则文件 (`handleRefreshDnsRoutingRule`)
- 刷新规则文件并保存配置后
- 调用 `globalContext.dnsServer.ReloadDnsRouter(ctx)` 重新加载路由器

## 增强的 ReloadDnsRouter 方法

原来的 `ReloadDnsRouter` 方法只重新加载自定义域名规则，现在增强为：

1. 重新加载自定义域名规则
2. 重新加载所有启用的规则文件
3. **清除路由缓存**以确保立即生效
4. **清除上游配置缓存**以确保使用最新的上游分组配置
5. 记录详细的重新加载日志

```go
func (s *Server) ReloadDnsRouter(ctx context.Context) {
    // 1. 获取所有规则配置
    // 2. 清除路由缓存
    // 3. 清除上游配置缓存
    // 4. 重新加载自定义规则
    // 5. 重新加载规则文件
    // 6. 记录日志
}
```

## 测试验证

使用 `test-routing-reload.ps1` 脚本测试：

```powershell
.\test-routing-reload.ps1
```

测试场景：
1. 添加自定义域名规则 → 立即生效
2. 更新规则到不同的上游分组 → 立即生效
3. 删除规则 → 立即恢复默认行为

## 修改的文件

- `internal/home/dns_routing.go` - 在所有规则操作后添加重新加载调用
- `internal/dnsforward/dnsforward.go` - 增强 `ReloadDnsRouter` 方法

## 效果

- ✅ 添加规则后立即生效，无需禁用/启用
- ✅ 更新规则后立即生效，无需禁用/启用
- ✅ 删除规则后立即生效，无需禁用/启用
- ✅ 启用/禁用规则后立即生效
- ✅ 刷新规则后立即生效
- ✅ 路由缓存自动清除，确保新规则立即应用
- ✅ 上游配置缓存自动清除，确保使用最新的上游分组
