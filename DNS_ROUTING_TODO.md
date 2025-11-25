# DNS 路由功能待实现

## 当前状态

### 已完成
1. ✅ 规则文件正确加载（`data/filters/1764005946.txt`）
2. ✅ 规则匹配成功（`||baidu.com^` 匹配到了）
3. ✅ 过滤器ID正确识别（1764005946）
4. ✅ `filtering.Result` 结构已添加 `UpstreamGroup` 字段

### 问题
- ❌ DNS 查询使用的是默认DNS（223.6.6.6），而不是指定的国内DNS分组（114.114.114.114, 223.5.5.5）

## 需要实现的功能

### 步骤 1：在过滤匹配时设置 UpstreamGroup

**文件**：`internal/filtering/filtering.go`

**位置**：`matchHostProcessAllowList` 函数

**需要做的**：
1. 从匹配的规则中获取 `FilterListID`
2. 根据 `FilterListID` 查找对应的过滤器配置
3. 获取过滤器的 `UpstreamGroup`
4. 设置到 `Result.UpstreamGroup`

**代码示例**：
```go
func (d *DNSFilter) matchHostProcessAllowList(
    ctx context.Context,
    host string,
    dnsres *urlfilter.DNSResult,
) (res Result, err error) {
    // ... 现有代码 ...
    
    res = makeResult(matchedRules, NotFilteredAllowList)
    
    // 新增：获取 upstream group
    if len(matchedRules) > 0 {
        filterID := matchedRules[0].GetFilterListID()
        upstreamGroup := d.getUpstreamGroupByFilterID(filterID)
        res.UpstreamGroup = upstreamGroup
    }
    
    return res, nil
}

// 新增函数：根据 FilterID 获取 UpstreamGroup
func (d *DNSFilter) getUpstreamGroupByFilterID(filterID rules.ListID) string {
    d.conf.filtersMu.RLock()
    defer d.conf.filtersMu.RUnlock()
    
    // 在 DnsRoutingFilters 中查找
    for _, filter := range d.conf.DnsRoutingFilters {
        if filter.ID == filterID {
            return filter.UpstreamGroup
        }
    }
    
    return ""
}
```

### 步骤 2：在 DNS 查询处理中使用 UpstreamGroup

**文件**：`internal/dnsforward/process.go`

**位置**：`processUpstream` 函数，在 `setCustomUpstream` 之后

**需要做的**：
1. 检查 `dctx.result.UpstreamGroup` 是否有值
2. 如果有值，根据 `UpstreamGroup` ID 查找上游组配置
3. 创建自定义上游配置并设置到 `pctx.CustomUpstreamConfig`

**代码示例**：
```go
func (s *Server) processUpstream(ctx context.Context, dctx *dnsContext) (rc resultCode) {
    // ... 现有代码 ...
    
    // 现有的客户端自定义上游
    s.setCustomUpstream(ctx, pctx, dctx.clientID)
    
    // 新增：DNS 路由规则的上游组
    if dctx.result != nil && dctx.result.UpstreamGroup != "" {
        upstreamGroup := s.getUpstreamGroupByID(dctx.result.UpstreamGroup)
        if upstreamGroup != nil && upstreamGroup.Enabled {
            s.logger.DebugContext(
                ctx,
                "using dns routing upstream group",
                "group_id", dctx.result.UpstreamGroup,
                "group_name", upstreamGroup.Name,
            )
            
            // 创建自定义上游配置
            upsConf := s.createUpstreamConfig(upstreamGroup)
            if upsConf != nil {
                pctx.CustomUpstreamConfig = upsConf
            }
        }
    }
    
    // ... 其余代码 ...
}
```

### 步骤 3：实现辅助函数

**文件**：`internal/dnsforward/upstream_groups.go`

**需要实现的函数**：

```go
// getUpstreamGroupByID 根据组ID获取上游组配置
func (s *Server) getUpstreamGroupByID(groupID string) *UpstreamGroup {
    s.serverLock.RLock()
    defer s.serverLock.RUnlock()
    
    for _, group := range s.conf.UpstreamGroups {
        if group.ID == groupID {
            return &group
        }
    }
    
    return nil
}

// createUpstreamConfig 根据上游组创建上游配置
func (s *Server) createUpstreamConfig(group *UpstreamGroup) *proxy.CustomUpstreamConfig {
    if len(group.Upstreams) == 0 {
        return nil
    }
    
    // 解析上游服务器地址
    upstreams := make([]upstream.Upstream, 0, len(group.Upstreams))
    for _, addr := range group.Upstreams {
        u, err := upstream.AddressToUpstream(addr, s.conf.UpstreamConfig)
        if err != nil {
            s.logger.Error("failed to parse upstream", "addr", addr, "error", err)
            continue
        }
        upstreams = append(upstreams, u)
    }
    
    if len(upstreams) == 0 {
        return nil
    }
    
    return &proxy.CustomUpstreamConfig{
        Upstreams: upstreams,
    }
}
```

## 测试验证

### 测试步骤

1. **编译并启动**
   ```bash
   go build -o AdGuardHome.exe
   ./AdGuardHome.exe
   ```

2. **查询中国域名**
   ```bash
   nslookup baidu.com 127.0.0.1
   ```

3. **检查日志**
   应该看到：
   ```
   [DEBUG] using dns routing upstream group group_id=group_1763970331409 group_name=国内DNS
   ```

4. **检查响应详情**
   - DNS 服务器应该是：`114.114.114.114:53` 或 `223.5.5.5:53`
   - 不应该是：`223.6.6.6:53`（默认DNS）

### 预期结果

```
响应细节：
状态: 允许项
DNS 服务器: 114.114.114.114:53  ✅ 正确
耗时: 15 毫秒
响应代码: NOERROR
规则: ||baidu.com^
未知过滤器 1764005946
上游组: 国内DNS (group_1763970331409)  ✅ 新增显示
```

## 配置示例

### AdGuardHome.yaml

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
    
    - id: group_1763984268251
      name: 223.6.6.6
      upstreams:
        - 223.6.6.6
      enabled: true
      is_default: true  # 默认DNS

dns_routing_filters:
  - enabled: true
    url: https://.../China_Classical.yaml
    name: CN
    id: 1764005946
    upstream_group: group_1763970331409  # 关联到国内DNS
```

## 关键点

1. **优先级**：DNS 路由规则的上游组应该优先于默认上游
2. **回退机制**：如果上游组不可用，应该回退到默认上游
3. **日志记录**：应该记录使用了哪个上游组
4. **性能**：查找上游组应该高效（考虑使用 map 缓存）

## 文件清单

需要修改的文件：
1. ✅ `internal/filtering/result.go` - 添加 UpstreamGroup 字段（已完成）
2. `internal/filtering/filtering.go` - 在匹配时设置 UpstreamGroup
3. `internal/dnsforward/process.go` - 使用 UpstreamGroup 选择上游
4. `internal/dnsforward/upstream_groups.go` - 实现辅助函数

## 下一步

继续实现步骤 1、2、3，完成 DNS 路由功能。
