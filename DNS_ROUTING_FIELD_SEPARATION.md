# DNS路由字段分离说明

## 字段结构

为了避免与现有的上游规则组功能冲突，DNS路由规则使用独立的字段：

### Filter 结构体（不变）
```go
type Filter struct {
    FilePath string `yaml:"-"`
    Data []byte `yaml:"-"`
    ID rules.ListID `yaml:"id"`
    
    // UpstreamGroup - 用于原有的上游规则组功能
    // 保持不变，不用于DNS路由
    UpstreamGroup string `yaml:"upstream_group,omitempty"`
}
```

### FilterYAML 结构体（新增字段）
```go
type FilterYAML struct {
    Enabled     bool
    URL         string
    Name        string    `yaml:"name"`
    RulesCount  int       `yaml:"-"`
    LastUpdated time.Time `yaml:"-"`
    checksum    uint32
    white       bool

    // DnsRoutingUpstreamGroup - 专门用于DNS路由规则
    // 独立字段，不与 Filter.UpstreamGroup 冲突
    DnsRoutingUpstreamGroup string `yaml:"dns_routing_upstream_group,omitempty"`

    Filter `yaml:",inline"`
}
```

## 配置文件格式

### DNS路由规则
```yaml
dns_routing_filters:
  - enabled: true
    url: https://raw.githubusercontent.com/.../China_Classical.yaml
    name: CN
    id: 1764000553
    dns_routing_upstream_group: group_1763970331409  # 使用新字段名
```

### 普通过滤器（如果使用上游组功能）
```yaml
filters:
  - enabled: true
    url: https://example.com/filter.txt
    name: My Filter
    id: 123
    upstream_group: some_group_id  # 原有字段，用于其他功能
```

## 字段用途对比

| 字段名 | 所属结构 | 用途 | YAML字段名 |
|--------|---------|------|-----------|
| `Filter.UpstreamGroup` | Filter | 原有的上游规则组功能 | `upstream_group` |
| `FilterYAML.DnsRoutingUpstreamGroup` | FilterYAML | DNS路由规则的上游组 | `dns_routing_upstream_group` |

## 代码中的使用

### 判断是否为DNS路由规则
```go
// 使用 DnsRoutingUpstreamGroup 字段
isDNSRoutingRule := flt.DnsRoutingUpstreamGroup != ""
```

### 添加DNS路由规则
```go
filt := FilterYAML{
    Enabled:                 true,
    URL:                     url,
    Name:                    name,
    DnsRoutingUpstreamGroup: upstreamGroupID,  // 使用新字段
    Filter: Filter{
        ID: generateID(),
        // UpstreamGroup 保持为空，不使用
    },
}
```

### 更新DNS路由规则
```go
if flt.DnsRoutingUpstreamGroup != newList.DnsRoutingUpstreamGroup {
    flt.DnsRoutingUpstreamGroup = newList.DnsRoutingUpstreamGroup
    shouldRestart = true
}
```

### 返回给前端
```go
func filterToJSON(f FilterYAML) filterJSON {
    return filterJSON{
        ID:            f.ID,
        Name:          f.Name,
        URL:           f.URL,
        UpstreamGroup: f.DnsRoutingUpstreamGroup,  // 映射到前端的 upstream_group
        // ...
    }
}
```

## API接口

### 前端到后端
前端发送：
```json
{
  "name": "CN",
  "url": "https://...",
  "dns_routing": true,
  "upstream_group": "group_1763970331409"
}
```

后端接收并映射到 `DnsRoutingUpstreamGroup`

### 后端到前端
后端返回：
```json
{
  "id": 1764000553,
  "name": "CN",
  "url": "https://...",
  "upstream_group": "group_1763970331409",
  "enabled": true,
  "rules_count": 3728
}
```

前端接收的 `upstream_group` 实际来自 `DnsRoutingUpstreamGroup`

## Clash规则处理

当满足以下条件时，自动处理Clash规则：
1. `DnsRoutingUpstreamGroup != ""` （是DNS路由规则）
2. `IsClashRuleURL(url)` 返回 true（URL包含 `/Clash/` 或以 `.yaml` 结尾）

处理逻辑：
```go
if isDNSRoutingRule && isClashRule {
    // 过滤IP规则，只保留域名规则
    stats, err := ProcessClashRuleFile(r, tmpFile)
    // ...
}
```

## 重要说明

1. **不要混淆两个字段**：
   - `Filter.UpstreamGroup` - 原有功能，保持不变
   - `FilterYAML.DnsRoutingUpstreamGroup` - DNS路由专用

2. **配置文件中的字段名**：
   - 原有功能：`upstream_group`
   - DNS路由：`dns_routing_upstream_group`

3. **前端API保持简单**：
   - 前端统一使用 `upstream_group` 字段名
   - 后端根据 `dns_routing` 标志决定映射到哪个字段

4. **向后兼容**：
   - 旧的配置文件不受影响
   - 新字段使用 `omitempty`，不存在时不会保存

## 测试验证

### 1. 添加DNS路由规则
```bash
# 添加规则后，检查配置文件
cat AdGuardHome.yaml | grep -A 5 "dns_routing_filters"

# 应该看到 dns_routing_upstream_group 字段
```

### 2. 检查Clash规则处理
```bash
# 查看日志
# 应该看到：
# [info] processing clash rule for dns routing id=xxx
# [info] clash rule processed total_rules=5000 valid_domains=3728 filtered_ip_rules=1272
```

### 3. 检查规则文件
```bash
# 查看保存的规则文件
cat data/filters/1764000553.txt | head -20

# 应该只看到域名规则，格式如：
# baidu.com
# ||cn^
# *china*
# 不应该有 IP-CIDR 或原始的 YAML 格式
```

## 故障排查

### 问题：规则文件仍包含IP规则
**原因**：`DnsRoutingUpstreamGroup` 字段为空，未触发Clash处理

**解决**：
1. 检查配置文件中是否有 `dns_routing_upstream_group` 字段
2. 确保添加规则时选择了上游组
3. 删除旧的规则文件，让系统重新下载

### 问题：配置文件中没有 dns_routing_upstream_group
**原因**：添加规则时没有正确保存

**解决**：
1. 检查前端是否正确传递 `upstream_group`
2. 检查后端是否正确映射到 `DnsRoutingUpstreamGroup`
3. 重新添加规则并选择上游组

## 相关文件

- `internal/filtering/filter.go` - FilterYAML 结构体定义
- `internal/filtering/filtering.go` - Filter 结构体定义
- `internal/filtering/http.go` - API处理和字段映射
- `internal/filtering/clash_rules.go` - Clash规则处理
- `AdGuardHome.yaml` - 配置文件
