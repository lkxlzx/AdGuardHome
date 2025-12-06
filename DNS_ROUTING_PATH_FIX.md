# DNS 路由规则文件路径修复

## 问题描述

DNS 路由规则在启动时无法加载，导致：
1. 规则文件存在但加载失败（rules_count=0）
2. 域名查询使用默认上游而不是配置的上游
3. 需要"禁用再启用"才能触发重新加载

## 根本原因

**文件路径不一致**：

- **实际文件位置**：`data/dns_routing_rules/{id}.txt`
- **代码查找位置**：`data/filters/{id}.txt` ❌

在 `internal/dnsforward/dnsforward.go` 的 `loadRoutingRuleFile` 函数中使用了错误的路径。

## 修复内容

### 修改文件
`internal/dnsforward/dnsforward.go`

### 修改前
```go
func (s *Server) loadRoutingRuleFile(ctx context.Context, filterID int64, upstreamGroup string, priority int) []dnsrouting.Rule {
	// The file should be at data/filters/{filterID}.txt
	filePath := fmt.Sprintf("data/filters/%d.txt", filterID)
	// ...
}
```

### 修改后
```go
func (s *Server) loadRoutingRuleFile(ctx context.Context, filterID int64, upstreamGroup string, priority int) []dnsrouting.Rule {
	// The file should be at data/dns_routing_rules/{filterID}.txt
	filePath := fmt.Sprintf("data/dns_routing_rules/%d.txt", filterID)
	// ...
}
```

## 验证结果

### 修复前
```
[info] dnsforward: added routing source id=4 name=GFW rules_count=0  ❌
```

### 修复后
```
[info] dnsforward: loaded routing rule file filter_id=4 format=3 total_lines=6872 valid_rules=6872
[info] dnsforward: added routing source id=4 name=GFW rules_count=6872  ✅
[info] dnsforward: DNS routing matched domain=google.com upstream_group=708dd52d-f6fb-4863-bfce-75a7f33c899d
[info] dnsforward: applied custom upstream config domain=google.com upstream_group=708dd52d-f6fb-4863-bfce-75a7f33c899d
```

## 测试验证

### 文件存在性检查
```powershell
PS> Get-ChildItem "data\dns_routing_rules"

Name              Length LastWriteTime
----              ------ -------------
3.txt            1834267 2025/12/6 14:50:51  ✅
4.txt             114307 2025/12/6 14:51:36  ✅
custom_rules.txt       2 2025/12/6 14:47:17  ✅
```

### DNS 查询测试
```powershell
PS> nslookup google.com 127.0.0.1
服务器:  UnKnown
Address:  127.0.0.1

非权威应答:
名称:    google.com
Address:  8.7.198.46  ✅ 使用国内 DNS (223.6.6.6) 解析
```

## 相关路径说明

系统中有两个不同的目录用于不同目的：

1. **`data/filters/`** - 用于广告过滤规则
   - 由 filtering 模块管理
   - 存储 AdGuard 格式的过滤规则

2. **`data/dns_routing_rules/`** - 用于 DNS 路由规则
   - 由 dnsroutingfiles 模块管理
   - 存储域名路由规则（决定使用哪个上游 DNS）

## 其他相关代码

以下代码已经使用正确的路径：

- `internal/home/dns.go` 中的 `reloadDnsRoutingRules` 使用 `GetRuleFilePath()` 方法
- `internal/dnsroutingfiles/storage.go` 中定义了正确的常量：
  ```go
  rulesDirName = "data/dns_routing_rules"
  ```

## 影响范围

### 修复前的问题
- ❌ 启动时规则文件无法加载
- ❌ 规则不生效，使用默认上游
- ❌ 需要手动"禁用再启用"触发重新加载

### 修复后的效果
- ✅ 启动时自动加载规则文件
- ✅ 规则立即生效
- ✅ 无需任何手动操作
- ✅ 配置的上游分组正确应用

## 编译版本

修复版本：`AdGuardHome_path_fix.exe`

## 总结

这是一个简单但关键的路径错误。修复后，DNS 路由功能完全正常工作，规则可以在启动时自动加载并立即生效。
