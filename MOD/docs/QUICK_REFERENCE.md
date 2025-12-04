# DNS上游分组 - 快速参考

## Go结构体定义（复制即用）

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

## API端点清单

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/control/dns/upstream_groups` | 获取所有分组 |
| POST | `/control/dns/upstream_groups` | 创建分组 |
| PUT | `/control/dns/upstream_groups/:id` | 更新分组 |
| DELETE | `/control/dns/upstream_groups/:id` | 删除分组 |
| POST | `/control/dns/upstream_groups/:id/default` | 设置默认分组 |
| POST | `/control/dns/upstream_groups/:id/test` | 测试连通性 |

## 配置文件格式示例

```yaml
dns:
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
      created_at: 2024-01-01T00:00:00Z
      updated_at: 2024-01-01T00:00:00Z
```

## 验证规则

### 创建/更新分组
- ✅ 分组名称：必填，1-50字符，唯一
- ✅ upstream_dns：必填，至少一个服务器
- ✅ is_default：如果为true，取消其他分组的默认状态
- ✅ enabled：默认true

### 删除分组
- ❌ 不能删除默认分组（返回409 Conflict）

### 默认分组
- ✅ 系统中始终有且仅有一个默认分组
- ✅ 启动时自动检查并修复

## 错误响应格式

```json
{
  "error": "error_type",
  "message": "错误描述"
}
```

| HTTP状态码 | error类型 | 说明 |
|-----------|----------|------|
| 400 | validation_error | 请求数据验证失败 |
| 404 | not_found | 分组不存在 |
| 409 | conflict | 业务规则冲突 |
| 500 | internal_error | 服务器内部错误 |

## 实现文件位置

| 文件 | 说明 |
|------|------|
| `internal/home/dns_upstream_groups.go` | HTTP处理器（新建） |
| `internal/home/config.go` | 配置结构（修改） |
| `internal/home/control.go` | 路由注册（修改） |
| `internal/dnsforward/upstream_groups.go` | 分组管理器（可选） |

## 前端已实现

- ✅ Redux Actions: `client/src/actions/upstreamGroups.ts`
- ✅ Redux Reducer: `client/src/reducers/upstreamGroups.ts`
- ✅ API客户端: `client/src/api/Api.ts`
- ✅ UI组件: `client/src/components/Settings/Dns/UpstreamGroups/`
- ✅ 国际化: `client/src/__locales/zh-cn.json`

## 测试清单

- [ ] 创建分组
- [ ] 编辑分组
- [ ] 删除分组（非默认）
- [ ] 尝试删除默认分组（应失败）
- [ ] 设置默认分组
- [ ] 启用/禁用分组
- [ ] 测试连通性
- [ ] 配置持久化
- [ ] 重启后配置恢复
- [ ] 验证YAML格式正确

## 参考文档

- 详细对接指南: `BACKEND_INTEGRATION_GUIDE.md`
- 格式说明: `YAML_FORMAT_NOTE.md`
- 更新总结: `DOCUMENTATION_UPDATE_SUMMARY.md`
- 设计文档: `.kiro/specs/dns-upstream-groups/design.md`
- 任务列表: `.kiro/specs/dns-upstream-groups/tasks.md`
