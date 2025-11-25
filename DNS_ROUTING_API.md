# DNS 上游分组路由 API 文档

## 概述

DNS 上游分组功能已完成，现在提供了一组 API 接口供 DNS 分流规则使用。这些接口允许根据域名匹配规则选择不同的上游 DNS 服务器组。

## 核心接口

### 1. GetUpstreamsByGroupID

根据分组 ID 获取上游服务器列表。

**函数签名**:
```go
func (s *Server) GetUpstreamsByGroupID(groupID string) (upstreams []upstream.Upstream, err error)
```

**参数**:
- `groupID`: 分组的唯一标识符（例如："group_1701234567890"）

**返回值**:
- `upstreams`: 上游服务器列表
- `err`: 错误信息（如果有）

**行为**:
- 如果分组不存在，返回 `nil, nil`
- 如果分组被禁用（`Enabled = false`），返回 `nil, nil`
- 自动过滤空行和注释行（以 `#` 开头）
- 使用配置的 Bootstrap DNS 和超时设置

**使用示例**:
```go
// 在 DNS 查询处理中
upstreams, err := s.GetUpstreamsByGroupID("group_1701234567890")
if err != nil {
    // 处理错误
    return err
}

if upstreams != nil {
    // 使用这些上游服务器进行查询
    reply, err := upstream.ExchangeParallel(upstreams, req)
    // ...
}
```

### 2. GetDefaultUpstreamGroup

获取默认上游分组。

**函数签名**:
```go
func (s *Server) GetDefaultUpstreamGroup() *UpstreamGroup
```

**返回值**:
- `*UpstreamGroup`: 默认分组，如果没有则返回 `nil`

**行为**:
- 返回标记为默认（`IsDefault = true`）且已启用的分组
- 如果没有默认分组，返回 `nil`

**使用示例**:
```go
// 当没有路由规则匹配时
defaultGroup := s.GetDefaultUpstreamGroup()
if defaultGroup != nil {
    upstreams, err := s.GetUpstreamsByGroupID(defaultGroup.ID)
    // 使用默认分组的上游服务器
}
```

### 3. GetEnabledUpstreamGroups

获取所有已启用的上游分组。

**函数签名**:
```go
func (s *Server) GetEnabledUpstreamGroups() []UpstreamGroup
```

**返回值**:
- `[]UpstreamGroup`: 已启用的分组列表

**使用示例**:
```go
// 列出所有可用的分组（用于 UI 或配置）
groups := s.GetEnabledUpstreamGroups()
for _, group := range groups {
    fmt.Printf("Group: %s (ID: %s)\n", group.Name, group.ID)
}
```

### 4. GetUpstreamGroupByName

根据分组名称获取分组。

**函数签名**:
```go
func (s *Server) GetUpstreamGroupByName(name string) *UpstreamGroup
```

**参数**:
- `name`: 分组名称（例如："国内 DNS"）

**返回值**:
- `*UpstreamGroup`: 匹配的分组，如果没有则返回 `nil`

**使用示例**:
```go
// 通过名称查找分组
group := s.GetUpstreamGroupByName("国内 DNS")
if group != nil {
    upstreams, err := s.GetUpstreamsByGroupID(group.ID)
    // ...
}
```

## 数据结构

### UpstreamGroup

```go
type UpstreamGroup struct {
    // ID 是分组的唯一标识符
    ID string `yaml:"id" json:"id"`

    // Name 是分组的显示名称
    Name string `yaml:"name" json:"name"`

    // Upstreams 是换行分隔的上游 DNS 服务器列表
    Upstreams string `yaml:"upstreams" json:"upstreams"`

    // Enabled 表示此分组是否启用
    Enabled bool `yaml:"enabled" json:"enabled"`

    // IsDefault 表示这是否是默认分组（当路由规则不匹配时使用）
    IsDefault bool `yaml:"is_default" json:"is_default"`
}
```

## DNS 路由集成示例

### 场景 1: 基于域名的路由

```go
// 伪代码：DNS 查询处理
func (s *Server) handleDNSRequest(req *dns.Msg) (*dns.Msg, error) {
    domain := req.Question[0].Name
    
    // 1. 检查路由规则
    groupID := s.matchRoutingRule(domain)
    
    // 2. 根据规则选择上游分组
    var upstreams []upstream.Upstream
    var err error
    
    if groupID != "" {
        // 使用匹配的分组
        upstreams, err = s.GetUpstreamsByGroupID(groupID)
    } else {
        // 使用默认分组
        defaultGroup := s.GetDefaultUpstreamGroup()
        if defaultGroup != nil {
            upstreams, err = s.GetUpstreamsByGroupID(defaultGroup.ID)
        }
    }
    
    if err != nil {
        return nil, err
    }
    
    if upstreams == nil {
        // 回退到原有的上游配置
        upstreams = s.dnsProxy.Upstreams
    }
    
    // 3. 使用选定的上游服务器进行查询
    reply, err := upstream.ExchangeParallel(upstreams, req)
    return reply, err
}
```

### 场景 2: 路由规则配置

```go
// 路由规则示例
type RoutingRule struct {
    // 域名匹配模式（支持通配符）
    DomainPattern string
    
    // 目标分组 ID
    GroupID string
}

// 示例规则
var routingRules = []RoutingRule{
    {
        DomainPattern: "*.cn",
        GroupID:       "group_1701234567890", // 国内 DNS
    },
    {
        DomainPattern: "*.com",
        GroupID:       "group_1701234567891", // 国外 DNS
    },
}

// 匹配路由规则
func (s *Server) matchRoutingRule(domain string) string {
    for _, rule := range routingRules {
        if matchDomain(domain, rule.DomainPattern) {
            return rule.GroupID
        }
    }
    return "" // 没有匹配，使用默认分组
}
```

## 配置文件示例

### AdGuardHome.yaml

```yaml
dns:
  upstream_groups:
    - id: "group_1701234567890"
      name: "国内 DNS"
      upstreams: |
        223.5.5.5
        119.29.29.29
        https://dns.alidns.com/dns-query
      enabled: true
      is_default: true
      
    - id: "group_1701234567891"
      name: "国外 DNS"
      upstreams: |
        8.8.8.8
        1.1.1.1
        https://dns.google/dns-query
      enabled: true
      is_default: false
      
    - id: "group_1701234567892"
      name: "安全 DNS"
      upstreams: |
        https://dns.quad9.net/dns-query
        https://cloudflare-dns.com/dns-query
      enabled: false
      is_default: false
```

## 线程安全

所有接口都使用 `serverLock` 读锁保护，确保并发访问安全：

```go
s.serverLock.RLock()
defer s.serverLock.RUnlock()
```

## 性能考虑

1. **缓存**: 考虑缓存已解析的上游配置，避免重复解析
2. **锁粒度**: 使用读锁而不是写锁，允许并发读取
3. **快速路径**: 禁用的分组会立即返回，不进行解析

## 错误处理

接口设计遵循 Go 的错误处理惯例：

- 返回 `nil, nil` 表示分组不存在或被禁用（不是错误）
- 返回 `nil, err` 表示发生了实际错误（如解析失败）
- 调用者应该检查两个返回值

## 下一步开发

### 1. DNS 路由规则管理

需要实现：
- 路由规则的数据结构
- 路由规则的 CRUD API
- 域名匹配逻辑（支持通配符、正则表达式）
- 规则优先级管理

### 2. 集成到 DNS 查询处理

需要修改：
- `internal/dnsforward/process.go` - DNS 查询处理逻辑
- 在查询前匹配路由规则
- 根据规则选择上游分组
- 回退到默认分组

### 3. 前端路由规则 UI

需要创建：
- 路由规则列表组件
- 添加/编辑路由规则表单
- 域名匹配测试工具

## API 使用检查清单

在实现 DNS 路由功能时，请确保：

- [ ] 检查分组是否存在
- [ ] 检查分组是否启用
- [ ] 处理 `nil` 返回值
- [ ] 处理错误返回值
- [ ] 提供默认分组回退
- [ ] 记录日志（用于调试）
- [ ] 添加性能监控

## 示例：完整的路由处理流程

```go
func (s *Server) processWithRouting(req *dns.Msg) (*dns.Msg, error) {
    domain := req.Question[0].Name
    
    // 1. 匹配路由规则
    groupID := s.matchRoutingRule(domain)
    
    // 2. 获取上游服务器
    var upstreams []upstream.Upstream
    var err error
    var groupName string
    
    if groupID != "" {
        upstreams, err = s.GetUpstreamsByGroupID(groupID)
        groupName = groupID
    }
    
    // 3. 回退到默认分组
    if upstreams == nil && err == nil {
        defaultGroup := s.GetDefaultUpstreamGroup()
        if defaultGroup != nil {
            upstreams, err = s.GetUpstreamsByGroupID(defaultGroup.ID)
            groupName = defaultGroup.Name
        }
    }
    
    // 4. 错误处理
    if err != nil {
        s.logger.Error("failed to get upstreams", "error", err)
        return nil, err
    }
    
    // 5. 最终回退到原有配置
    if upstreams == nil {
        upstreams = s.dnsProxy.Upstreams
        groupName = "default"
    }
    
    // 6. 记录日志
    s.logger.Debug("routing query",
        "domain", domain,
        "group", groupName,
        "upstreams", len(upstreams))
    
    // 7. 执行查询
    reply, err := upstream.ExchangeParallel(upstreams, req)
    return reply, err
}
```

## 总结

DNS 上游分组的核心 API 已经实现并可以使用。这些接口提供了：

1. ✅ 根据 ID 获取上游服务器
2. ✅ 获取默认分组
3. ✅ 列出所有启用的分组
4. ✅ 根据名称查找分组
5. ✅ 线程安全保证
6. ✅ 错误处理机制

下一步可以基于这些接口实现 DNS 路由规则功能。

---

**文件位置**: `internal/dnsforward/upstream_groups.go`  
**创建日期**: 2024年11月24日  
**版本**: 1.0
