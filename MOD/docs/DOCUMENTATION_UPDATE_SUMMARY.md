# 文档更新总结

## 更新原因

确保所有文档中关于配置文件格式的描述保持一致，使用与AdGuard Home现有配置格式相同的YAML块格式。

## 更新的文档

### 1. BACKEND_INTEGRATION_GUIDE.md（后端对接指南）

**更新内容**:
- ✅ Go结构体定义：YAML标签在前，JSON标签在后
- ✅ 移除了不必要的 `flow` 标签
- ✅ YAML配置示例：使用块格式，移除引号
- ✅ 添加了详细的YAML序列化说明
- ✅ 强调与现有 `bootstrap_dns` 格式保持一致

**关键变更**:
```go
// 之前（错误）
UpstreamDNS  []string  `json:"upstream_dns" yaml:"upstream_dns,flow"`

// 现在（正确）
UpstreamDNS  []string  `yaml:"upstream_dns" json:"upstream_dns"`
```

```yaml
# 之前（错误）
upstream_dns: ["223.6.6.6", "119.29.29.29"]

# 现在（正确）
upstream_dns:
  - 223.6.6.6
  - 119.29.29.29
```

### 2. .kiro/specs/dns-upstream-groups/design.md（设计文档）

**更新内容**:
- ✅ 更新了两处Go结构体定义，YAML标签在前
- ✅ 更新了配置文件格式示例，使用块格式
- ✅ 添加了格式说明注释

**关键变更**:
```go
// 更新前
type UpstreamGroup struct {
    ID           string    `json:"id" yaml:"id"`
    UpstreamDNS  []string  `json:"upstream_dns" yaml:"upstream_dns"`
}

// 更新后
type UpstreamGroup struct {
    ID           string    `yaml:"id" json:"id"`
    UpstreamDNS  []string  `yaml:"upstream_dns" json:"upstream_dns"`
}
```

### 3. .kiro/specs/dns-upstream-groups/tasks.md（任务文档）

**更新内容**:
- ✅ 在概述部分添加了"重要说明"章节
- ✅ 说明了配置文件格式要求
- ✅ 提供了正确的Go结构体定义示例
- ✅ 更新了任务1.1和1.2的描述，强调格式验证

**新增内容**:
```markdown
## 重要说明

### 配置文件格式
配置文件必须使用YAML块格式（每行一个条目）...

### Go结构体定义
type UpstreamGroup struct {
    UpstreamDNS  []string  `yaml:"upstream_dns" json:"upstream_dns"`
    ...
}
```

### 4. YAML_FORMAT_NOTE.md（新建）

**内容**:
- ✅ 清晰展示正确和错误的格式对比
- ✅ 提供完整的Go结构体定义
- ✅ 说明HTTP API和配置文件格式的区别
- ✅ 提供参考文件路径

### 5. .kiro/specs/dns-upstream-groups/requirements.md（需求文档）

**状态**: 无需更新
- 已包含配置文件格式兼容性要求（需求10.5）
- 描述足够通用，不需要具体格式细节

## 统一的格式标准

### YAML配置文件格式（块格式）
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

### Go结构体定义
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

### HTTP API响应格式（JSON）
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

## 关键要点

1. **YAML标签顺序**: YAML在前，JSON在后（与AdGuard Home代码风格一致）
2. **无需特殊标签**: 使用标准标签，`gopkg.in/yaml.v3` 会自动处理
3. **块格式输出**: `[]string` 类型自动序列化为块格式（每行一个条目）
4. **omitempty标签**: 确保空数组不会出现在配置文件中
5. **参考现有实现**: 与 `bootstrap_dns` 字段使用完全相同的方式

## 验证清单

实现时需要验证：
- [ ] Go结构体标签顺序正确（YAML在前）
- [ ] 配置文件输出为块格式，不是流式格式
- [ ] 字符串值没有不必要的引号
- [ ] 空数组不会出现在配置文件中（omitempty生效）
- [ ] HTTP API响应使用标准JSON格式
- [ ] 与现有 `bootstrap_dns` 配置格式完全一致

## 参考文件

- 现有配置示例: `MOD/app/AdGuardHome.yaml`
- 现有结构定义: `internal/dnsforward/config.go` 中的 `BootstrapDNS` 字段
- YAML库: `gopkg.in/yaml.v3`
- 后端对接指南: `BACKEND_INTEGRATION_GUIDE.md`
- 格式说明: `YAML_FORMAT_NOTE.md`

## 更新日期

2024-12-04

## 更新人员

Kiro AI Assistant
