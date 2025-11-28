# DNS路由过滤优先级修复说明

## 问题描述

之前DNS路由功能使用了白名单模板，导致优先级过高，使得广告拦截功能对DNS路由规则中的域名失效。

**具体表现**：
- 加载了CN域名规则作为DNS路由
- 启用了广告拦截功能
- 结果：CN域名列表中的广告域名不会被拦截 ❌

## 修复内容

调整了过滤规则的检查顺序，确保广告拦截等过滤规则优先于DNS路由规则。

### 新的优先级顺序

```
1. 白名单规则（最高优先级）
   ↓
2. 黑名单过滤规则（广告拦截、恶意软件拦截等）
   ↓
3. DNS路由规则（最低优先级）
```

### 修复效果

**修复前**：
```
域名: ad.example.cn
- 在CN域名列表中（DNS路由规则）
- 在广告拦截列表中（黑名单规则）
结果: 使用DNS路由，不被拦截 ❌
```

**修复后**：
```
域名: ad.example.cn
- 在CN域名列表中（DNS路由规则）
- 在广告拦截列表中（黑名单规则）
结果: 被广告拦截规则拦截 ✓
```

## 使用方法

### 1. 使用修复版本

```powershell
# 停止当前运行的AdGuardHome
Stop-Process -Name "AdGuardHome*" -Force -ErrorAction SilentlyContinue

# 启动修复版本
.\AdGuardHome_filter_priority_fix.exe
```

### 2. 验证修复

测试一个同时在DNS路由规则和广告拦截列表中的域名：

```powershell
# 测试DNS查询
nslookup ad.example.cn 127.0.0.1
```

**预期结果**：
- 域名被拦截（返回0.0.0.0或NXDOMAIN）
- 查询日志显示被过滤规则拦截

### 3. 自动化测试

运行测试脚本：

```powershell
.\test_dns_routing_filter_priority.ps1
```

## 兼容性说明

### ✓ 不影响的功能

- 白名单规则仍然具有最高优先级
- DNS路由在关闭过滤功能时仍然工作
- 现有配置无需修改

### ⚠️ 行为变化

- DNS路由规则中的域名现在会受到过滤规则的影响
- 如果某个域名同时在DNS路由规则和黑名单中，将被拦截而不是路由

## 特殊需求

如果需要某些域名**绕过过滤规则**并**强制使用DNS路由**：

1. 将域名添加到**白名单**中
2. 同时保留在DNS路由规则中
3. 白名单会优先生效，然后DNS路由规则会处理路由

**示例**：
```
白名单规则: @@||example.cn^
DNS路由规则: ||example.cn^ (使用china_dns上游组)
结果: example.cn不会被拦截，并使用china_dns上游组解析
```

## 测试场景

### 场景1：广告域名（应该被拦截）
```powershell
nslookup ad.example.cn 127.0.0.1
# 预期：被拦截，返回0.0.0.0
```

### 场景2：正常域名（应该走DNS路由）
```powershell
nslookup baidu.com 127.0.0.1
# 预期：正常解析，使用DNS路由的上游组
```

### 场景3：白名单域名（最高优先级）
```powershell
# 添加白名单规则: @@||example.com^
nslookup example.com 127.0.0.1
# 预期：正常解析，即使在黑名单中
```

## 查看日志

启用调试日志可以看到详细的匹配过程：

```
[DEBUG] blocklist matched (takes priority over DNS routing) host=ad.example.cn
[DEBUG] DNS routing matched (not blocked by filters) host=baidu.com upstream_group=china_dns
```

## 常见问题

**Q: DNS路由规则不生效了？**
A: 检查域名是否在黑名单中。如果需要强制路由，将域名添加到白名单。

**Q: 广告域名仍然不被拦截？**
A: 确保启用了Protection Enabled和Filtering Enabled。

**Q: 如何让某些域名绕过过滤？**
A: 将域名添加到白名单中。

## 文件说明

- `AdGuardHome_filter_priority_fix.exe` - 修复版本的可执行文件
- `test_dns_routing_filter_priority.ps1` - 自动化测试脚本
- `DNS_ROUTING_FILTER_PRIORITY_FIX.md` - 详细技术文档（英文）
- `FILTER_PRIORITY_FIX_VERIFICATION.md` - 验证指南（英文）

## 总结

此修复确保了广告拦截等过滤功能能够正确作用于DNS路由规则中的域名，解决了之前DNS路由优先级过高导致过滤失效的问题。同时保持了DNS路由功能的独立性和灵活性。

**核心改进**：
- ✓ 广告拦截规则现在优先于DNS路由规则
- ✓ 白名单规则仍然具有最高优先级
- ✓ DNS路由功能保持独立性
- ✓ 向后兼容，无需修改现有配置
