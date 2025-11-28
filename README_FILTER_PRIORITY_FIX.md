# DNS路由过滤优先级修复

## 📋 概述

修复了DNS路由规则优先级过高导致广告拦截功能失效的问题。

**问题**: DNS路由使用白名单模板，优先级高于黑名单过滤规则  
**影响**: CN域名列表中的广告域名不会被拦截  
**解决**: 调整过滤规则检查顺序，黑名单规则现在优先于DNS路由规则  

## 🚀 快速开始

```powershell
# 1. 停止当前AdGuardHome
Stop-Process -Name "AdGuardHome*" -Force -ErrorAction SilentlyContinue

# 2. 启动修复版本
.\AdGuardHome_filter_priority_fix.exe

# 3. 测试验证
.\test_dns_routing_filter_priority.ps1
```

## 📊 优先级变化

### 修复前
```
DNS路由规则 (最高) → 白名单规则 → 黑名单规则 (最低)
```

### 修复后
```
白名单规则 (最高) → 黑名单规则 → DNS路由规则 (最低)
```

## 📁 文件说明

### 可执行文件
- **AdGuardHome_filter_priority_fix.exe** (39.45 MB)
  - 包含修复的编译版本

### 文档文件（按推荐阅读顺序）

1. **快速使用指南.md** ⭐ 推荐首先阅读
   - 快速上手指南
   - 包含最常用的操作

2. **DNS路由过滤优先级修复说明.md**
   - 详细的中文说明
   - 包含问题分析、修复方案、测试方法

3. **PRIORITY_FIX_SUMMARY.md**
   - 修复总结
   - 技术细节和代码变更

4. **DNS_ROUTING_FILTER_PRIORITY_FIX.md** (English)
   - 详细技术文档（英文版）

5. **FILTER_PRIORITY_FIX_VERIFICATION.md** (English)
   - 验证指南（英文版）

### 测试脚本
- **test_dns_routing_filter_priority.ps1**
  - 自动化测试脚本
  - 验证修复效果

## ✅ 修复效果

### 场景1: 广告域名
```
域名: ad.example.cn
规则: 在CN域名列表 + 在广告拦截列表
修复前: 不被拦截 ❌
修复后: 被拦截 ✓
```

### 场景2: 正常域名
```
域名: baidu.com
规则: 仅在CN域名列表
修复前: 走DNS路由 ✓
修复后: 走DNS路由 ✓
```

### 场景3: 白名单域名
```
域名: example.com
规则: 在白名单
修复前: 不被拦截 ✓
修复后: 不被拦截 ✓
```

## 🔧 技术细节

### 修改的文件
- `internal/filtering/filtering.go` - matchHost函数

### 核心改动
```go
// 新的检查顺序
1. 白名单规则检查
2. 黑名单规则检查 ← 新增：优先于DNS路由
3. DNS路由规则检查 ← 调整：最后检查
```

### 优先级矩阵

| 规则类型 | FilteringEnabled=false | FilteringEnabled=true |
|---------|----------------------|---------------------|
| 白名单   | 不检查                | ✓ 最高优先级          |
| 黑名单   | 不检查                | ✓ 第二优先级          |
| DNS路由  | ✓ 检查                | ✓ 最低优先级          |

## 🧪 测试验证

### 自动化测试
```powershell
.\test_dns_routing_filter_priority.ps1
```

### 手动测试
```powershell
# 测试广告域名（应该被拦截）
nslookup ad.example.cn 127.0.0.1

# 测试正常域名（应该走DNS路由）
nslookup baidu.com 127.0.0.1

# 测试白名单域名（不被拦截）
nslookup example.com 127.0.0.1
```

### 验证检查清单
- [ ] 广告域名被正确拦截
- [ ] 正常域名走DNS路由
- [ ] 白名单域名不被拦截
- [ ] 查询日志显示正确的过滤原因
- [ ] DNS路由在FilteringEnabled=false时仍然工作

## 🔄 兼容性

### ✓ 保持不变
- 白名单规则仍然具有最高优先级
- DNS路由在关闭过滤功能时仍然工作
- DNS路由独立于ProtectionEnabled设置
- 现有配置无需修改

### ⚠️ 行为变化
- DNS路由规则中的域名现在会受到过滤规则的影响
- 如果某个域名同时在DNS路由规则和黑名单中，将被拦截

## 💡 特殊需求

如果需要某些域名**绕过过滤**并**强制使用DNS路由**：

```
1. 添加白名单规则: @@||example.cn^
2. 保留DNS路由规则: ||example.cn^ (使用china_dns)
3. 结果: 不被拦截，使用DNS路由
```

## 📝 日志标识

修复后的版本会在日志中显示：

```
[DEBUG] allow list matched host=example.com
[DEBUG] blocklist matched (takes priority over DNS routing) host=ad.example.cn
[DEBUG] DNS routing matched (not blocked by filters) host=baidu.com
```

## ❓ 常见问题

**Q: DNS路由规则不生效了？**  
A: 检查域名是否在黑名单中。如需强制路由，将域名添加到白名单。

**Q: 广告域名仍然不被拦截？**  
A: 确保启用了Protection Enabled和Filtering Enabled。

**Q: 如何让某些域名绕过过滤？**  
A: 将域名添加到白名单中。

**Q: 修复会影响性能吗？**  
A: 不会，保持了原有的性能优化。

**Q: 需要修改现有配置吗？**  
A: 不需要，完全向后兼容。

## 🔙 回滚方案

如果遇到问题，可以回滚到之前的版本：

```powershell
Stop-Process -Name "AdGuardHome*" -Force
.\AdGuardHome.exe  # 使用之前的版本
```

## 📊 修复状态

- ✅ 代码修改完成
- ✅ 编译成功
- ✅ 语法检查通过
- ✅ 文档完整
- ✅ 测试脚本就绪
- ⏳ 等待实际测试验证

## 📞 反馈

如果发现任何问题，请记录：
1. 测试的域名
2. 预期行为 vs 实际行为
3. 查询日志截图
4. AdGuardHome日志相关部分

## 📅 修复信息

- **修复日期**: 2025-11-28
- **修复版本**: AdGuardHome_filter_priority_fix.exe
- **修改文件**: internal/filtering/filtering.go
- **修改函数**: matchHost
- **编译状态**: ✅ 成功

## 🎯 总结

此修复确保了广告拦截等过滤功能能够正确作用于DNS路由规则中的域名，解决了DNS路由优先级过高导致过滤失效的问题。同时保持了DNS路由功能的独立性和灵活性，完全向后兼容现有配置。

---

**开始使用**: 阅读 `快速使用指南.md`  
**详细了解**: 阅读 `DNS路由过滤优先级修复说明.md`  
**技术细节**: 阅读 `PRIORITY_FIX_SUMMARY.md`
