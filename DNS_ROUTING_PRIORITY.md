# DNS 路由规则优先级

## 优先级顺序

DNS 查询的上游服务器选择按以下优先级顺序：

```
1. 自定义域名规则 (custom_domain_rules) - 最高优先级
   ↓ 未匹配
2. URL规则文件 (dns_routing_filters)
   ↓ 未匹配
3. 客户端自定义上游 (client-based custom upstreams)
   ↓ 未匹配
4. 默认上游 DNS (upstream_dns)
```

## 配置示例

### AdGuardHome.yaml

```yaml
dns:
  # 默认上游 DNS（优先级最低）
  upstream_dns:
    - https://dns10.quad9.net/dns-query
  
  # 上游组定义
  upstream_groups:
    - id: group_cn
      name: 国内DNS
      upstreams:
        - 114.114.114.114
        - 223.5.5.5
      enabled: true
    
    - id: group_global
      name: 海外DNS
      upstreams:
        - 8.8.8.8
        - 1.1.1.1
      enabled: true
    
    - id: group_default
      name: 默认DNS
      upstreams:
        - 223.6.6.6
      enabled: true
      is_default: true
  
  # 自定义域名规则（优先级最高）
  custom_domain_rules:
    - domain: google.com
      match_type: DOMAIN
      upstream_group: group_global
    
    - domain: qq.com
      match_type: DOMAIN-KEYWORD
      upstream_group: group_cn

# URL规则文件（优先级中等）
dns_routing_filters:
  - enabled: true
    url: https://.../China_Classical.yaml
    name: CN
    id: 1764005946
    upstream_group: group_cn
```

## 查询示例

### 示例 1：匹配自定义规则

**查询**：`google.com`

**处理流程**：
1. 检查自定义规则
2. ✅ 匹配：`google.com` (DOMAIN)
3. 使用上游组：`group_global` (海外DNS)
4. DNS服务器：`8.8.8.8` 或 `1.1.1.1`

**结果**：使用海外DNS解析

---

### 示例 2：匹配URL规则

**查询**：`baidu.com`

**处理流程**：
1. 检查自定义规则
2. ❌ 未匹配
3. 检查URL规则文件
4. ✅ 匹配：`||baidu.com^` (在 CN 规则文件中)
5. 使用上游组：`group_cn` (国内DNS)
6. DNS服务器：`114.114.114.114` 或 `223.5.5.5`

**结果**：使用国内DNS解析

---

### 示例 3：匹配关键词规则

**查询**：`www.qq.com`

**处理流程**：
1. 检查自定义规则
2. ✅ 匹配：`qq.com` (DOMAIN-KEYWORD)
3. 使用上游组：`group_cn` (国内DNS)
4. DNS服务器：`114.114.114.114` 或 `223.5.5.5`

**结果**：使用国内DNS解析（优先于URL规则）

---

### 示例 4：无匹配规则

**查询**：`example.com`

**处理流程**：
1. 检查自定义规则
2. ❌ 未匹配
3. 检查URL规则文件
4. ❌ 未匹配
5. 使用默认上游DNS

**结果**：使用默认DNS解析

## 匹配类型说明

### DOMAIN（精确匹配）

```yaml
- domain: google.com
  match_type: DOMAIN
```

**匹配**：
- ✅ `google.com`

**不匹配**：
- ❌ `www.google.com`
- ❌ `mail.google.com`

### DOMAIN-SUFFIX（后缀匹配）

```yaml
- domain: google.com
  match_type: DOMAIN-SUFFIX
```

**匹配**：
- ✅ `google.com`
- ✅ `www.google.com`
- ✅ `mail.google.com`
- ✅ `any.subdomain.google.com`

**不匹配**：
- ❌ `google.cn`
- ❌ `notgoogle.com`

### DOMAIN-KEYWORD（关键词匹配）

```yaml
- domain: google
  match_type: DOMAIN-KEYWORD
```

**匹配**：
- ✅ `google.com`
- ✅ `www.google.com`
- ✅ `google.cn`
- ✅ `mygoogle.com`

**不匹配**：
- ❌ `baidu.com`

## 优先级冲突处理

### 场景：同一域名有多个规则

**配置**：
```yaml
custom_domain_rules:
  - domain: baidu.com
    match_type: DOMAIN
    upstream_group: group_global  # 海外DNS

dns_routing_filters:
  - name: CN
    upstream_group: group_cn  # 国内DNS
    # 规则文件包含 ||baidu.com^
```

**查询**：`baidu.com`

**结果**：使用 `group_global`（海外DNS）

**原因**：自定义规则优先级更高

## 实现细节

### 代码位置

**internal/dnsforward/process.go**:
```go
// 优先级顺序
s.setCustomUpstream(ctx, pctx, dctx.clientID)  // 客户端自定义

// 1. 检查自定义域名规则（最高优先级）
if groupID := s.matchCustomDomainRule(domain); groupID != "" {
    s.setDNSRoutingUpstream(ctx, pctx, groupID)
} 
// 2. 检查URL规则文件
else if dctx.result != nil && dctx.result.UpstreamGroup != "" {
    s.setDNSRoutingUpstream(ctx, pctx, dctx.result.UpstreamGroup)
}
```

### 匹配函数

```go
func (s *Server) matchCustomDomainRule(domain string) string {
    // 遍历所有自定义规则
    for _, rule := range s.conf.CustomDomainRules {
        if s.matchDomainPattern(domain, rule.Domain, rule.MatchType) {
            return rule.UpstreamGroup
        }
    }
    return ""
}

func (s *Server) matchDomainPattern(domain, pattern, matchType string) bool {
    switch matchType {
    case "DOMAIN":
        return domain == pattern
    case "DOMAIN-SUFFIX":
        return domain == pattern || strings.HasSuffix(domain, "."+pattern)
    case "DOMAIN-KEYWORD":
        return strings.Contains(domain, pattern)
    default:
        return domain == pattern
    }
}
```

## 日志示例

### 匹配自定义规则

```
[DEBUG] started processing upstream
[DEBUG] matched custom domain rule domain=google.com rule=google.com match_type=DOMAIN upstream_group=group_global
[DEBUG] using dns routing upstream group group_id=group_global group_name=海外DNS
[DEBUG] finished processing upstream
```

### 匹配URL规则

```
[DEBUG] started processing upstream
[DEBUG] no custom domain rule matched
[DEBUG] dns routing upstream group found filter_id=1764005946 upstream_group=group_cn
[DEBUG] using dns routing upstream group group_id=group_cn group_name=国内DNS
[DEBUG] finished processing upstream
```

## 测试验证

### 测试 1：自定义规则优先

1. **配置自定义规则**：
   ```yaml
   custom_domain_rules:
     - domain: baidu.com
       match_type: DOMAIN
       upstream_group: group_global  # 海外DNS
   ```

2. **查询**：`baidu.com`

3. **预期结果**：
   - ✅ 使用海外DNS（8.8.8.8）
   - ❌ 不使用国内DNS（即使URL规则文件中有 ||baidu.com^）

### 测试 2：URL规则作为回退

1. **查询**：`qq.com`（假设自定义规则中没有）

2. **预期结果**：
   - ✅ 使用URL规则文件中的匹配
   - ✅ 使用国内DNS（114.114.114.114）

### 测试 3：关键词匹配

1. **配置自定义规则**：
   ```yaml
   custom_domain_rules:
     - domain: qq
       match_type: DOMAIN-KEYWORD
       upstream_group: group_cn
   ```

2. **查询**：`www.qq.com`, `mail.qq.com`, `myqq.com`

3. **预期结果**：
   - ✅ 所有包含 `qq` 的域名都使用国内DNS

## 性能考虑

### 优化建议

1. **自定义规则数量**：建议不超过 100 条
2. **匹配类型选择**：
   - 精确匹配（DOMAIN）最快
   - 后缀匹配（DOMAIN-SUFFIX）次之
   - 关键词匹配（DOMAIN-KEYWORD）最慢

3. **规则顺序**：
   - 将常用域名放在前面
   - 考虑添加缓存机制

## 总结

DNS 路由规则现在按正确的优先级顺序处理：

1. ✅ **自定义规则优先**：用户手动配置的规则最优先
2. ✅ **URL规则回退**：自定义规则未匹配时使用URL规则
3. ✅ **默认上游保底**：所有规则都未匹配时使用默认上游

这样的设计既灵活又高效，满足各种使用场景！
