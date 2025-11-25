# DNS路由引擎独立性测试指南

## 测试目标

验证DNS路由引擎完全独立于 `FilteringEnabled` 和 `ProtectionEnabled` 设置。

## 关键修改点

### 1. matchHost函数重构

**之前的问题**：
```go
func (d *DNSFilter) matchHost(...) (res Result, err error) {
    if !setts.FilteringEnabled {
        return Result{}, nil  // ❌ 直接返回，DNS路由引擎不会被检查
    }
    // ... 其他逻辑
}
```

**修复后**：
```go
func (d *DNSFilter) matchHost(...) (res Result, err error) {
    // 1. 首先检查DNS路由引擎（不受任何设置影响）
    if d.filteringEngineDnsRouting != nil {
        // 检查DNS路由规则
        if result.UpstreamGroup != "" {
            return result, nil  // ✅ 立即返回路由结果
        }
    }
    
    // 2. 然后才检查FilteringEnabled
    if !setts.FilteringEnabled {
        return Result{}, nil  // 其他过滤被跳过，但DNS路由已经检查过了
    }
    
    // 3. 其他过滤逻辑...
}
```

### 2. 检查顺序

新的检查顺序确保DNS路由总是优先：

1. **DNS路由引擎** - 总是检查，不受任何设置影响
2. **FilteringEnabled检查** - 如果为false，跳过后续过滤
3. **允许列表** - 仅当 ProtectionEnabled 为 true 时检查
4. **阻止列表** - 仅当 ProtectionEnabled 为 true 时检查

## 测试场景

### 场景1：所有保护关闭，DNS路由开启

**配置**：
```yaml
dns:
  protection_enabled: false
  filtering_enabled: false

filters:
  dns_routing_filters:
    - enabled: true
      id: 1
      name: "China Domains"
      url: "file://china_domains.txt"
      upstream_group: "china"
      priority: 1
```

**预期结果**：
- ✅ DNS路由规则仍然生效
- ✅ 匹配的域名被路由到指定的上游组
- ✅ 日志显示 "DNS routing matched"

### 场景2：FilteringEnabled关闭，ProtectionEnabled开启

**配置**：
```yaml
dns:
  protection_enabled: true
  filtering_enabled: false

filters:
  dns_routing_filters:
    - enabled: true
      id: 1
      name: "China Domains"
      url: "file://china_domains.txt"
      upstream_group: "china"
      priority: 1
```

**预期结果**：
- ✅ DNS路由规则仍然生效
- ❌ 普通过滤规则不生效（因为FilteringEnabled为false）
- ✅ 日志显示 "DNS routing matched"

### 场景3：所有保护开启，DNS路由关闭

**配置**：
```yaml
dns:
  protection_enabled: true
  filtering_enabled: true

filters:
  dns_routing_filters:
    - enabled: false  # 单独关闭DNS路由
      id: 1
      name: "China Domains"
      url: "file://china_domains.txt"
      upstream_group: "china"
      priority: 1
```

**预期结果**：
- ❌ DNS路由规则不生效（因为过滤器自己的enabled为false）
- ✅ 普通过滤规则正常工作
- ❌ 日志不显示 "DNS routing matched"

### 场景4：所有开启

**配置**：
```yaml
dns:
  protection_enabled: true
  filtering_enabled: true

filters:
  dns_routing_filters:
    - enabled: true
      id: 1
      name: "China Domains"
      url: "file://china_domains.txt"
      upstream_group: "china"
      priority: 1
```

**预期结果**：
- ✅ DNS路由规则生效
- ✅ 普通过滤规则生效
- ✅ DNS路由规则优先于其他规则

## 测试步骤

### 1. 准备测试环境

创建测试用的DNS路由规则文件 `china_domains.txt`：
```
||baidu.com^
||qq.com^
||taobao.com^
```

### 2. 配置AdGuardHome

编辑 `AdGuardHome.yaml`：
```yaml
dns:
  bind_hosts:
    - 0.0.0.0
  port: 53
  protection_enabled: false  # 关闭保护
  filtering_enabled: false   # 关闭过滤
  upstream_dns:
    - https://dns.google/dns-query
  upstream_groups:
    china:
      - https://dns.alidns.com/dns-query

filters:
  dns_routing_filters:
    - enabled: true
      id: 1
      name: "China Domains"
      url: "file://china_domains.txt"
      upstream_group: "china"
      priority: 1
```

### 3. 启动并测试

```bash
# 启动AdGuardHome
./AdGuardHome_v3_latest.exe

# 测试DNS查询
nslookup baidu.com 127.0.0.1
```

### 4. 检查日志

查找以下日志消息：
```
DNS routing matched host=baidu.com upstream_group=china
returning DNS routing rule upstream_group=china
```

### 5. 验证行为

- 即使 `protection_enabled: false` 和 `filtering_enabled: false`
- DNS查询 `baidu.com` 仍然应该被路由到 `china` 上游组
- 日志应该显示 "DNS routing matched"

## 调试技巧

### 启用详细日志

在 `AdGuardHome.yaml` 中：
```yaml
log:
  file: ""
  max_backups: 0
  max_size: 100
  max_age: 3
  compress: false
  local_time: false
  verbose: true  # 启用详细日志
```

### 关键日志消息

成功的DNS路由应该显示：
```
[debug] DNS routing matched host=example.com upstream_group=china
[debug] returning DNS routing rule upstream_group=china
```

如果没有看到这些消息，检查：
1. DNS路由过滤器的 `enabled` 字段是否为 `true`
2. 规则文件是否存在且格式正确
3. 域名是否匹配规则

## 预期结果总结

| FilteringEnabled | ProtectionEnabled | DNS路由Filter.Enabled | DNS路由工作 | 普通过滤工作 |
|-----------------|-------------------|---------------------|-----------|-----------|
| false | false | true | ✅ 是 | ❌ 否 |
| false | true | true | ✅ 是 | ❌ 否 |
| true | false | true | ✅ 是 | ❌ 否 |
| true | true | true | ✅ 是 | ✅ 是 |
| any | any | false | ❌ 否 | 取决于设置 |

## 成功标准

✅ DNS路由引擎完全独立，满足以下条件：
1. 不受 `FilteringEnabled` 影响
2. 不受 `ProtectionEnabled` 影响
3. 只受各个过滤器自己的 `Enabled` 字段控制
4. 总是优先于其他过滤规则检查
5. 日志清晰显示DNS路由匹配过程
