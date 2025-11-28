# DNS路由过滤优先级修复 - 验证指南

## 修复内容

已修复DNS路由规则优先级过高导致广告拦截功能失效的问题。

### 问题
- DNS路由使用白名单模板，优先级高于黑名单过滤规则
- 导致CN域名列表中的广告域名不会被拦截

### 解决方案
调整过滤检查顺序：
1. 白名单规则（最高优先级）
2. **黑名单过滤规则**（广告拦截等）
3. DNS路由规则（最低优先级）

## 快速验证

### 1. 使用新编译的版本

```powershell
# 停止当前运行的AdGuardHome
Stop-Process -Name "AdGuardHome*" -Force -ErrorAction SilentlyContinue

# 启动修复版本
.\AdGuardHome_filter_priority_fix.exe
```

### 2. 配置测试环境

确保你的AdGuardHome配置中：
- ✓ 启用了过滤功能（Filtering Enabled）
- ✓ 启用了保护功能（Protection Enabled）
- ✓ 添加了DNS路由规则（如CN域名列表）
- ✓ 添加了广告拦截规则

### 3. 测试场景

#### 场景1：广告域名在DNS路由规则中

**测试域名**: 选择一个同时在CN域名列表和广告拦截列表中的域名

```powershell
# 测试DNS查询
nslookup ad.example.cn 127.0.0.1
```

**预期结果**:
- ✓ 域名被拦截（返回0.0.0.0或NXDOMAIN）
- ✓ 查询日志显示 `FilteredBlockList` 原因
- ✓ 不会显示DNS路由的上游组

#### 场景2：正常域名在DNS路由规则中

**测试域名**: 选择一个仅在CN域名列表中的正常域名

```powershell
# 测试DNS查询
nslookup baidu.com 127.0.0.1
```

**预期结果**:
- ✓ 域名正常解析
- ✓ 查询日志显示 `NotFilteredDNSRouting` 原因
- ✓ 显示使用的上游组

#### 场景3：白名单域名

**测试域名**: 添加到白名单的域名

```powershell
# 在白名单中添加: @@||example.com^
nslookup example.com 127.0.0.1
```

**预期结果**:
- ✓ 域名正常解析（即使在黑名单中）
- ✓ 查询日志显示 `NotFilteredAllowList` 原因

### 4. 查看日志

检查AdGuardHome日志，应该看到类似的调试信息：

```
blocklist matched (takes priority over DNS routing) host=ad.example.cn rule=||ad.example.cn^
DNS routing matched (not blocked by filters) host=baidu.com upstream_group=china_dns
```

## 详细测试脚本

运行提供的自动化测试脚本：

```powershell
.\test_dns_routing_filter_priority.ps1
```

## 验证检查清单

- [ ] 广告域名被正确拦截（即使在DNS路由规则中）
- [ ] 正常域名走DNS路由（未被拦截）
- [ ] 白名单域名优先级最高
- [ ] 查询日志显示正确的过滤原因
- [ ] DNS路由在FilteringEnabled=false时仍然工作

## 回滚方案

如果遇到问题，可以回滚到之前的版本：

```powershell
# 停止当前版本
Stop-Process -Name "AdGuardHome*" -Force

# 使用之前的版本
.\AdGuardHome.exe
```

## 常见问题

### Q: DNS路由规则不生效了？
A: 检查域名是否在黑名单中。如果需要强制路由，将域名添加到白名单。

### Q: 广告域名仍然不被拦截？
A: 确保：
1. Protection Enabled 已启用
2. Filtering Enabled 已启用
3. 广告拦截规则已正确加载

### Q: 如何让某些域名绕过过滤并强制路由？
A: 将域名同时添加到白名单和DNS路由规则中。

## 技术细节

### 修改的文件
- `internal/filtering/filtering.go` - matchHost函数

### 优先级逻辑

```
FilteringEnabled = false:
  └─> DNS路由规则

FilteringEnabled = true:
  ├─> 白名单规则 (最高优先级)
  ├─> 黑名单过滤规则 (广告拦截等)
  └─> DNS路由规则 (最低优先级)
```

### 日志标识

- `allow list matched` - 白名单匹配
- `blocklist matched (takes priority over DNS routing)` - 黑名单匹配
- `DNS routing matched (not blocked by filters)` - DNS路由匹配

## 反馈

如果发现任何问题，请记录：
1. 测试的域名
2. 预期行为 vs 实际行为
3. 查询日志截图
4. AdGuardHome日志相关部分

## 总结

此修复确保了过滤功能（广告拦截、恶意软件拦截等）能够正确作用于DNS路由规则中的域名，同时保持DNS路由功能的独立性。
