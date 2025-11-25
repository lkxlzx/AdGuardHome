# DNS 域名分流处理流程设计

## 概述

DNS 域名分流功能允许根据域名规则将 DNS 查询路由到不同的上游 DNS 服务器组。例如，中国域名使用国内 DNS 服务器解析，国外域名使用海外 DNS 服务器解析。

## 配置示例

### 上游组配置

```yaml
dns:
  upstream_groups:
    - id: group_1763970331409
      name: 国内DNS
      upstreams:
        - 114.114.114.114
        - 223.5.5.5
      enabled: true
      is_default: false
    
    - id: group_1763973311919
      name: 海外DNS
      upstreams:
        - 8.8.8.8
        - 1.1.1.1
      enabled: true
      is_default: false
```

### 域名分流规则配置

```yaml
dns_routing_filters:
  - enabled: true
    url: https://raw.githubusercontent.com/.../China_Classical.yaml
    name: CN
    id: 1764005946
    upstream_group: group_1763970331409  # 关联到"国内DNS"组
```

### 规则文件内容

文件：`data/filters/1764005946.txt`

```
weixin.com
||00cdn.com^
||10010.com^
||126.com^
||baidu.com^
...
```

## DNS 查询处理流程

### 流程图

```
DNS 查询请求
    ↓
1. 接收查询（例如：baidu.com）
    ↓
2. 检查是否启用域名分流
    ↓
3. 遍历 DNS 路由规则（dns_routing_filters）
    ↓
4. 匹配域名规则
    ├─ 匹配成功 → 使用关联的上游组
    │   ↓
    │   获取 upstream_group (group_1763970331409)
    │   ↓
    │   查找上游组配置
    │   ↓
    │   使用该组的 DNS 服务器（114.114.114.114, 223.5.5.5）
    │   ↓
    │   发送查询并返回结果
    │
    └─ 未匹配 → 继续检查下一条规则
        ↓
        所有规则都未匹配 → 使用默认上游 DNS
```

### 详细步骤

#### 步骤 1：接收 DNS 查询

```go
// 用户查询：baidu.com
query := "baidu.com"
qtype := dns.TypeA
```

#### 步骤 2：检查域名分流规则

```go
// 遍历所有启用的 DNS 路由规则
for _, filter := range dnsRoutingFilters {
    if !filter.Enabled {
        continue
    }
    
    // 检查域名是否匹配规则文件中的规则
    if matchDomain(query, filter.RulesFile) {
        // 找到匹配的规则
        upstreamGroupID := filter.UpstreamGroup
        break
    }
}
```

#### 步骤 3：域名匹配逻辑

规则文件中的规则格式：

1. **直接域名匹配**：`weixin.com`
   - 只匹配 `weixin.com`
   - 不匹配 `www.weixin.com`

2. **域名后缀匹配**：`||baidu.com^`
   - 匹配 `baidu.com`
   - 匹配 `www.baidu.com`
   - 匹配 `map.baidu.com`
   - 匹配任何 `*.baidu.com`

3. **域名关键词匹配**：`*taobao*`
   - 匹配包含 `taobao` 的任何域名
   - 例如：`taobao.com`, `www.taobao.com`, `m.taobao.com`

#### 步骤 4：获取上游组

```go
// 根据 upstreamGroupID 查找上游组配置
upstreamGroup := findUpstreamGroup(upstreamGroupID)

if upstreamGroup != nil && upstreamGroup.Enabled {
    // 使用该组的 DNS 服务器
    upstreams := upstreamGroup.Upstreams
    // upstreams = ["114.114.114.114", "223.5.5.5"]
}
```

#### 步骤 5：发送 DNS 查询

```go
// 使用上游组的 DNS 服务器进行查询
for _, upstream := range upstreams {
    result := queryDNS(upstream, query, qtype)
    if result.Success {
        return result
    }
}
```

## 实际示例

### 示例 1：查询中国域名

**查询**：`baidu.com`

**处理流程**：
1. 接收查询：`baidu.com`
2. 检查 DNS 路由规则
3. 在规则文件 `1764005946.txt` 中找到匹配：`||baidu.com^`
4. 获取关联的上游组：`group_1763970331409`（国内DNS）
5. 使用国内 DNS 服务器查询：
   - 尝试 `114.114.114.114`
   - 如果失败，尝试 `223.5.5.5`
6. 返回查询结果

**结果**：使用国内 DNS 服务器解析，速度快，结果准确

### 示例 2：查询国外域名

**查询**：`google.com`

**处理流程**：
1. 接收查询：`google.com`
2. 检查 DNS 路由规则
3. 在规则文件 `1764005946.txt` 中未找到匹配
4. 继续检查其他规则文件（如果有）
5. 所有规则都未匹配
6. 使用默认上游 DNS 服务器（配置中的 `upstream_dns`）
7. 返回查询结果

**结果**：使用默认 DNS 服务器解析

### 示例 3：查询子域名

**查询**：`www.weixin.com`

**处理流程**：
1. 接收查询：`www.weixin.com`
2. 检查 DNS 路由规则
3. 在规则文件中：
   - `weixin.com` 不匹配（只匹配精确域名）
   - 如果有 `||weixin.com^` 则匹配（匹配所有子域名）
4. 根据匹配结果决定使用哪个上游组

## 代码实现位置

### 后端核心代码

#### 1. 过滤器加载（internal/filtering/filter.go）

```go
// 加载 DNS 路由过滤器
func (d *DNSFilter) enableFiltersLocked(ctx context.Context, async bool) {
    // ...
    
    // 添加 DNS 路由过滤器到 allowFilters
    for _, filter := range d.conf.DnsRoutingFilters {
        if !filter.Enabled {
            continue
        }
        
        allowFilters = append(allowFilters, Filter{
            ID:            filter.ID,
            FilePath:      filter.Path(d.conf.DataDir),
            UpstreamGroup: filter.UpstreamGroup,  // 关键：保存上游组 ID
        })
    }
}
```

#### 2. DNS 查询处理（internal/dnsforward/process.go）

```go
// 处理 DNS 查询
func (s *Server) processDNSRequest(ctx context.Context, req *dns.Msg) *dns.Msg {
    // 1. 提取查询域名
    domain := req.Question[0].Name
    
    // 2. 检查域名分流规则
    upstreamGroupID := s.matchDomainRoutingRule(domain)
    
    // 3. 根据上游组 ID 选择 DNS 服务器
    if upstreamGroupID != "" {
        upstreams := s.getUpstreamsByGroupID(upstreamGroupID)
        return s.queryUpstreams(upstreams, req)
    }
    
    // 4. 使用默认上游 DNS
    return s.queryDefaultUpstreams(req)
}
```

#### 3. 上游组管理（internal/dnsforward/upstream_groups.go）

```go
// 根据组 ID 获取上游服务器列表
func (s *Server) getUpstreamsByGroupID(groupID string) []upstream.Upstream {
    for _, group := range s.conf.UpstreamGroups {
        if group.ID == groupID && group.Enabled {
            return group.Upstreams
        }
    }
    return nil
}
```

## 性能优化

### 1. 规则缓存

```go
// 缓存域名匹配结果
type DomainCache struct {
    cache map[string]string  // domain -> upstreamGroupID
    mu    sync.RWMutex
}

func (dc *DomainCache) Get(domain string) (string, bool) {
    dc.mu.RLock()
    defer dc.mu.RUnlock()
    groupID, ok := dc.cache[domain]
    return groupID, ok
}
```

### 2. 规则索引

```go
// 为规则建立索引，加快匹配速度
type RuleIndex struct {
    exactMatch  map[string]string  // 精确匹配
    suffixMatch []SuffixRule       // 后缀匹配
    keywordMatch []KeywordRule     // 关键词匹配
}
```

### 3. 并发查询

```go
// 并发查询多个上游服务器，使用最快的响应
func (s *Server) queryUpstreamsParallel(upstreams []upstream.Upstream, req *dns.Msg) *dns.Msg {
    results := make(chan *dns.Msg, len(upstreams))
    
    for _, u := range upstreams {
        go func(upstream upstream.Upstream) {
            result := upstream.Exchange(req)
            results <- result
        }(u)
    }
    
    // 返回第一个成功的响应
    return <-results
}
```

## 配置建议

### 1. 规则优先级

建议按以下顺序配置规则：

1. **特定域名规则**（优先级最高）
   - 例如：公司内部域名

2. **国家/地区规则**
   - 例如：中国域名、美国域名

3. **默认规则**（优先级最低）
   - 未匹配任何规则的域名

### 2. 上游组配置

```yaml
upstream_groups:
  # 国内 DNS（速度快，适合国内域名）
  - id: group_cn
    name: 国内DNS
    upstreams:
      - 114.114.114.114
      - 223.5.5.5
      - 119.29.29.29
  
  # 海外 DNS（无污染，适合国外域名）
  - id: group_global
    name: 海外DNS
    upstreams:
      - 8.8.8.8
      - 1.1.1.1
      - 9.9.9.9
  
  # 默认 DNS
  - id: group_default
    name: 默认DNS
    upstreams:
      - https://dns.alidns.com/dns-query
    is_default: true
```

### 3. 规则文件管理

```yaml
dns_routing_filters:
  # 中国域名 → 国内 DNS
  - enabled: true
    url: https://.../China.yaml
    name: CN域名
    upstream_group: group_cn
  
  # 国外域名 → 海外 DNS
  - enabled: true
    url: https://.../Global.yaml
    name: 全球域名
    upstream_group: group_global
```

## 监控和日志

### 日志示例

```
[INFO] DNS query: baidu.com
[DEBUG] Matched routing rule: CN (ID: 1764005946)
[DEBUG] Using upstream group: 国内DNS (group_1763970331409)
[DEBUG] Querying upstream: 114.114.114.114
[INFO] Response: 220.181.38.148 (latency: 15ms)
```

### 统计信息

建议记录以下统计信息：

- 每个规则的匹配次数
- 每个上游组的使用次数
- 每个上游服务器的响应时间
- 查询成功率

## 故障处理

### 1. 上游组不可用

```go
if upstreamGroup == nil || !upstreamGroup.Enabled {
    // 回退到默认上游 DNS
    return s.queryDefaultUpstreams(req)
}
```

### 2. 所有上游服务器失败

```go
for _, upstream := range upstreams {
    result := queryDNS(upstream, req)
    if result.Success {
        return result
    }
}

// 所有上游都失败，使用默认上游
return s.queryDefaultUpstreams(req)
```

### 3. 规则文件损坏

```go
if err := loadRuleFile(filePath); err != nil {
    log.Error("Failed to load rule file, disabling filter")
    filter.Enabled = false
}
```

## 总结

DNS 域名分流功能通过以下机制实现：

1. **规则匹配**：根据域名规则文件匹配查询域名
2. **组关联**：每个规则文件关联一个上游组
3. **智能路由**：根据匹配结果选择合适的 DNS 服务器
4. **性能优化**：通过缓存和索引提高匹配速度
5. **故障恢复**：提供多层回退机制确保服务可用

这样就实现了：
- 中国域名 → 国内 DNS（快速、准确）
- 国外域名 → 海外 DNS（无污染）
- 其他域名 → 默认 DNS（灵活配置）
