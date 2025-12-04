# YAML配置格式说明

## 重要提示

DNS上游分组的配置文件格式必须与AdGuard Home现有的配置格式保持一致。

## 正确的YAML格式

### ✅ 正确格式（块格式）
```yaml
upstream_groups:
  - id: 550e8400-e29b-41d4-a716-446655440000
    name: 国内DNS
    enabled: true
    is_default: true
    upstream_dns:
      - 223.6.6.6
      - 119.29.29.29
    bootstrap_dns:
      - 223.5.5.5
```

### ❌ 错误格式（流式格式）
```yaml
upstream_groups:
  - id: "550e8400-e29b-41d4-a716-446655440000"
    name: "国内DNS"
    enabled: true
    is_default: true
    upstream_dns: ["223.6.6.6", "119.29.29.29"]
    bootstrap_dns: ["223.5.5.5"]
```

## Go结构体定义

```go
type UpstreamGroup struct {
    ID           string    `yaml:"id" json:"id"`
    Name         string    `yaml:"name" json:"name"`
    Enabled      bool      `yaml:"enabled" json:"enabled"`
    IsDefault    bool      `yaml:"is_default" json:"is_default"`
    UpstreamDNS  []string  `yaml:"upstream_dns" json:"upstream_dns"`
    FallbackDNS  []string  `yaml:"fallback_dns,omitempty" json:"fallback_dns,omitempty"`
    BootstrapDNS []string  `yaml:"bootstrap_dns,omitempty" json:"bootstrap_dns,omitempty"`
    CreatedAt    time.Time `yaml:"created_at" json:"created_at"`
    UpdatedAt    time.Time `yaml:"updated_at" json:"updated_at"`
}
```

## 关键点

1. **使用标准YAML标签**: 不需要特殊的 `flow` 标签
2. **参考现有实现**: 与 `bootstrap_dns` 字段使用相同的方式
3. **使用yaml.v3库**: AdGuard Home使用 `gopkg.in/yaml.v3`
4. **自动块格式**: `[]string` 类型会自动序列化为块格式

## 参考

- 现有配置示例: `MOD/app/AdGuardHome.yaml`
- 现有结构定义: `internal/dnsforward/config.go` 中的 `BootstrapDNS` 字段
- YAML库: `gopkg.in/yaml.v3`

## HTTP API响应格式

虽然配置文件使用YAML块格式，但HTTP API响应使用标准JSON格式：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "国内DNS",
  "enabled": true,
  "is_default": true,
  "upstream_dns": ["223.6.6.6", "119.29.29.29"],
  "bootstrap_dns": ["223.5.5.5"]
}
```

这是正常的，因为JSON和YAML的序列化方式不同。
