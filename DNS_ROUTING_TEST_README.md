# DNS路由独立引擎 - 快速测试指南

## 🎯 目标

验证DNS路由引擎完全独立于 `FilteringEnabled` 和 `ProtectionEnabled` 设置。

## 📋 前提条件

- ✅ 已编译 `AdGuardHome_v3_latest.exe`
- ✅ 测试文件已创建（自动包含）

## 🚀 快速测试（3步）

### 步骤1：启动AdGuardHome

```bash
AdGuardHome_v3_latest.exe -c test_dns_routing_independence.yaml
```

**注意**：这个配置文件中：
- `protection_enabled: false` ❌
- `filtering_enabled: false` ❌
- DNS路由过滤器: `enabled: true` ✅

### 步骤2：运行测试

打开新的命令行窗口，运行：

```bash
test_dns_routing.bat
```

或手动测试：

```bash
# 测试中国域名（应该路由到china上游组）
nslookup baidu.com 127.0.0.1

# 测试自定义域名（应该路由到custom上游组）
nslookup example.com 127.0.0.1

# 测试普通域名（应该使用默认上游）
nslookup google.com 127.0.0.1
```

### 步骤3：检查日志

在AdGuardHome的控制台输出中，查找：

```
[debug] DNS routing matched host=baidu.com upstream_group=china
[debug] returning DNS routing rule upstream_group=china
```

## ✅ 成功标准

如果看到以下现象，说明DNS路由引擎独立工作成功：

1. ✅ `baidu.com` 查询成功
2. ✅ 日志显示 "DNS routing matched"
3. ✅ 日志显示 "upstream_group=china"
4. ✅ 即使 `protection_enabled` 和 `filtering_enabled` 都是 `false`

## 📊 预期结果对照

| 域名 | 预期上游组 | 预期日志 |
|------|----------|---------|
| baidu.com | china | DNS routing matched ... upstream_group=china |
| qq.com | china | DNS routing matched ... upstream_group=china |
| example.com | custom | DNS routing matched ... upstream_group=custom |
| google.com | default | (无DNS routing日志) |

## 🔍 详细测试

如需更详细的测试，参见：
- `DNS_ROUTING_INDEPENDENCE_TEST.md` - 完整测试指南
- `V3_DNS_ROUTING_COMPLETE.md` - 实现细节

## 🐛 故障排除

### 问题1：看不到日志
**解决**：确保配置文件中 `verbose: true`

### 问题2：DNS路由不工作
**检查**：
1. DNS路由过滤器的 `enabled` 字段是否为 `true`
2. 规则文件是否存在（`test_china_domains.txt`）
3. 域名是否匹配规则

### 问题3：端口被占用
**解决**：修改配置文件中的 `port: 53` 为其他端口（如 `5353`）

## 📁 测试文件说明

- `test_dns_routing_independence.yaml` - 测试配置（保护和过滤都关闭）
- `test_china_domains.txt` - 中国域名规则
- `test_custom_domains.txt` - 自定义域名规则
- `test_dns_routing.bat` - 自动化测试脚本

## 🎉 测试成功后

如果测试成功，说明：
1. ✅ DNS路由引擎完全独立
2. ✅ 不受 `FilteringEnabled` 影响
3. ✅ 不受 `ProtectionEnabled` 影响
4. ✅ 可以在生产环境中使用

## 📝 下一步

1. 将测试配置应用到实际环境
2. 根据需要调整DNS路由规则
3. 监控日志确保正常工作

---

**版本**: v3_latest  
**构建**: AdGuardHome_v3_latest.exe  
**状态**: ✅ 准备就绪
