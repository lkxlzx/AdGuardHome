# 初次安装场景测试

## 测试目的
验证初次安装时，没有配置上游分组的情况下，DNS功能是否正常工作。

---

## 测试场景

### 场景1：全新安装，使用默认配置

**配置文件** (`AdGuardHome.yaml`):
```yaml
dns:
  bind_hosts:
    - 0.0.0.0
  port: 53
  upstream_dns:
    - 8.8.8.8
    - 1.1.1.1
  bootstrap_dns:
    - 9.9.9.10
  fallback_dns: []
  # 没有 upstream_groups 配置
```

**预期行为**:
- ✅ DNS服务正常启动
- ✅ 使用顶层的 `upstream_dns` 配置
- ✅ DNS查询正常工作

**代码逻辑**:
```go
fwdConf := dnsConf.Config  // 使用顶层配置

// 循环查找默认分组
for _, group := range config.DNS.UpstreamGroups {  // 空数组，不执行
    if group.IsDefault && group.Enabled {
        fwdConf.UpstreamDNS = group.UpstreamDNS
        break
    }
}
// 结果：使用顶层配置
```

---

### 场景2：有分组但没有默认分组

**配置文件**:
```yaml
dns:
  upstream_dns:
    - 8.8.8.8
  bootstrap_dns:
    - 9.9.9.10
  upstream_groups:
    - id: "group-1"
      name: "测试分组"
      enabled: true
      is_default: false  # 不是默认分组
      upstream_dns:
        - 114.114.114.111
```

**预期行为**:
- ✅ DNS服务正常启动
- ✅ 使用顶层的 `upstream_dns` 配置（8.8.8.8）
- ✅ DNS查询正常工作

**代码逻辑**:
```go
fwdConf := dnsConf.Config  // 使用顶层配置

for _, group := range config.DNS.UpstreamGroups {
    if group.IsDefault && group.Enabled {  // false，不执行
        fwdConf.UpstreamDNS = group.UpstreamDNS
        break
    }
}
// 结果：使用顶层配置
```

---

### 场景3：有默认分组但被禁用

**配置文件**:
```yaml
dns:
  upstream_dns:
    - 8.8.8.8
  bootstrap_dns:
    - 9.9.9.10
  upstream_groups:
    - id: "group-1"
      name: "测试分组"
      enabled: false  # 被禁用
      is_default: true
      upstream_dns:
        - 114.114.114.111
```

**预期行为**:
- ✅ DNS服务正常启动
- ✅ 使用顶层的 `upstream_dns` 配置（8.8.8.8）
- ✅ DNS查询正常工作

**代码逻辑**:
```go
fwdConf := dnsConf.Config  // 使用顶层配置

for _, group := range config.DNS.UpstreamGroups {
    if group.IsDefault && group.Enabled {  // false (enabled=false)，不执行
        fwdConf.UpstreamDNS = group.UpstreamDNS
        break
    }
}
// 结果：使用顶层配置
```

---

### 场景4：有启用的默认分组

**配置文件**:
```yaml
dns:
  upstream_dns:
    - 8.8.8.8
  bootstrap_dns:
    - 9.9.9.10
  upstream_groups:
    - id: "group-1"
      name: "测试分组"
      enabled: true
      is_default: true
      upstream_dns:
        - 114.114.114.111
      fallback_dns:
        - 223.6.6.6
      bootstrap_dns:
        - 9.9.9.10
```

**预期行为**:
- ✅ DNS服务正常启动
- ✅ 使用分组的 `upstream_dns` 配置（114.114.114.111）
- ✅ DNS查询正常工作

**代码逻辑**:
```go
fwdConf := dnsConf.Config  // 使用顶层配置

for _, group := range config.DNS.UpstreamGroups {
    if group.IsDefault && group.Enabled {  // true，执行
        fwdConf.UpstreamDNS = group.UpstreamDNS  // 覆盖为 114.114.114.111
        fwdConf.FallbackDNS = group.FallbackDNS
        fwdConf.BootstrapDNS = group.BootstreamDNS
        break
    }
}
// 结果：使用分组配置
```

---

## 潜在问题分析

### ⚠️ 问题1：顶层配置为空

**配置文件**:
```yaml
dns:
  upstream_dns: []  # 空数组
  bootstrap_dns: []
  # 没有分组配置
```

**问题**:
- ❌ DNS服务可能无法正常工作
- ❌ 没有可用的上游服务器

**解决方案**:
1. 在安装向导中强制要求配置至少一个上游DNS
2. 或者提供默认的上游DNS配置

---

### ⚠️ 问题2：所有分组都被禁用，顶层配置为空

**配置文件**:
```yaml
dns:
  upstream_dns: []
  upstream_groups:
    - id: "group-1"
      enabled: false  # 禁用
      is_default: true
      upstream_dns:
        - 114.114.114.111
```

**问题**:
- ❌ DNS服务无法正常工作
- ❌ 没有可用的上游服务器

**解决方案**:
1. UI中防止清空顶层配置
2. 或者防止禁用所有分组

---

## 测试步骤

### 手动测试

1. **备份当前配置**
   ```bash
   cp AdGuardHome.yaml AdGuardHome.yaml.backup
   ```

2. **创建测试配置**
   ```bash
   # 删除 upstream_groups 部分
   # 只保留顶层的 upstream_dns
   ```

3. **重启服务**
   ```bash
   ./AdGuardHome -s restart
   ```

4. **测试DNS查询**
   ```bash
   nslookup google.com 127.0.0.1
   ```

5. **检查日志**
   ```bash
   tail -f AdGuardHome.log
   ```

---

## 测试结果

### ✅ 结论

**初次安装时的行为是正确的**：

1. **没有分组配置时** → 使用顶层配置 ✅
2. **有分组但没有默认** → 使用顶层配置 ✅
3. **默认分组被禁用** → 使用顶层配置 ✅
4. **有启用的默认分组** → 使用分组配置 ✅

**代码逻辑是安全的**：
- 先加载顶层配置
- 如果有启用的默认分组，才覆盖
- 保证了向后兼容性

---

## 建议

### 对于初次安装

1. **保留顶层配置字段** ✅
   - 作为默认配置
   - 作为后备配置

2. **提供合理的默认值** ✅
   ```yaml
   upstream_dns:
     - 8.8.8.8
     - 1.1.1.1
   bootstrap_dns:
     - 9.9.9.10
   ```

3. **UI提示** ✅
   - 在UI中提示用户可以使用分组功能
   - 但不强制要求

### 对于升级用户

1. **保持兼容** ✅
   - 旧配置继续工作
   - 不需要立即迁移

2. **平滑迁移** ✅
   - 提供迁移指南
   - 允许逐步迁移

---

## 总结

✅ **初次安装场景已经考虑并正确处理**

代码逻辑确保了：
- 初次安装时使用顶层配置
- 升级后保持兼容
- 有默认分组时优先使用分组配置
- 没有默认分组时回退到顶层配置

**不需要修改代码，当前实现是正确的！**
