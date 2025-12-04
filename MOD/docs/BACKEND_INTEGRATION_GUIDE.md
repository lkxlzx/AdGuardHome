# DNS上游分组功能 - 后端对接指南

## 功能概述

前端已经实现了完整的DNS上游分组管理UI，包括：
- 分组列表展示（表格形式）
- 创建/编辑分组对话框
- 删除分组功能
- 设置默认分组
- 启用/禁用分组
- 测试分组连通性
- 完整的Redux状态管理

## 前端已实现的API调用

前端API客户端（`client/src/api/Api.ts`）已经定义了以下方法：

### 1. 获取分组列表
```typescript
getUpstreamGroups()
// GET /control/dns/upstream_groups
// 返回: UpstreamGroup[]
```

### 2. 添加分组
```typescript
addUpstreamGroup(data: Partial<UpstreamGroup>)
// POST /control/dns/upstream_groups
// 请求体: { name, upstream_dns, fallback_dns?, bootstrap_dns?, enabled, is_default }
// 返回: UpstreamGroup
```

### 3. 更新分组
```typescript
updateUpstreamGroup(id: string, data: Partial<UpstreamGroup>)
// PUT /control/dns/upstream_groups/:id
// 请求体: { name, upstream_dns, fallback_dns?, bootstrap_dns?, enabled, is_default }
// 返回: UpstreamGroup
```

### 4. 删除分组
```typescript
deleteUpstreamGroup(id: string)
// DELETE /control/dns/upstream_groups/:id
// 返回: void
```

### 5. 设置默认分组
```typescript
setDefaultUpstreamGroup(id: string)
// POST /control/dns/upstream_groups/:id/default
// 返回: void
```

### 6. 测试分组连通性
```typescript
testUpstreamGroup(id: string)
// POST /control/dns/upstream_groups/:id/test
// 返回: TestResult
```


## 数据模型

### UpstreamGroup 接口
```typescript
interface UpstreamGroup {
  id: string;                    // UUID格式的唯一标识符
  name: string;                  // 分组名称
  enabled: boolean;              // 是否启用
  is_default: boolean;           // 是否为默认分组
  upstream_dns: string[];        // 上游DNS服务器列表
  fallback_dns?: string[];       // 备用DNS服务器列表（可选）
  bootstrap_dns?: string[];      // Bootstrap DNS服务器列表（可选）
  created_at: string;            // ISO 8601格式的创建时间
  updated_at: string;            // ISO 8601格式的更新时间
}
```

### TestResult 接口
```typescript
interface TestResult {
  group_id: string;
  results: UpstreamTestResult[];
}

interface UpstreamTestResult {
  upstream: string;              // 上游服务器地址
  success: boolean;              // 是否测试成功
  error?: string;                // 错误信息（如果失败）
  rtt?: number;                  // 响应时间（毫秒）
}
```

## 需要实现的后端API端点

### 路由注册位置
建议在 `internal/home/control.go` 或创建新文件 `internal/home/dns_upstream_groups.go`

### 1. GET /control/dns/upstream_groups
**功能**: 获取所有DNS上游分组

**响应示例**:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "国内DNS",
    "enabled": true,
    "is_default": true,
    "upstream_dns": ["223.6.6.6", "119.29.29.29"],
    "bootstrap_dns": ["223.5.5.5"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

**注意**: HTTP API响应使用JSON格式（数组），但配置文件使用YAML格式（每行一个条目）


### 2. POST /control/dns/upstream_groups
**功能**: 创建新的DNS上游分组

**请求体**:
```json
{
  "name": "海外DNS",
  "enabled": true,
  "is_default": false,
  "upstream_dns": ["https://dns.google/dns-query", "https://cloudflare-dns.com/dns-query"],
  "fallback_dns": ["8.8.8.8", "1.1.1.1"],
  "bootstrap_dns": ["8.8.8.8"]
}
```

**验证规则**:
- `name`: 必填，1-50字符，唯一
- `upstream_dns`: 必填，至少包含一个有效的DNS服务器地址
- `enabled`: 默认true
- `is_default`: 如果设为true，需要将其他分组的is_default设为false

**响应**: 返回创建的UpstreamGroup对象（包含生成的id和时间戳）

### 3. PUT /control/dns/upstream_groups/:id
**功能**: 更新现有DNS上游分组

**请求体**: 同POST，但所有字段都是可选的

**验证规则**:
- 验证分组ID存在
- 如果修改is_default为true，需要将其他分组的is_default设为false
- 不能清空upstream_dns

**响应**: 返回更新后的UpstreamGroup对象

### 4. DELETE /control/dns/upstream_groups/:id
**功能**: 删除DNS上游分组

**验证规则**:
- 验证分组ID存在
- **不能删除默认分组**（返回409 Conflict）

**响应**: 204 No Content

### 5. POST /control/dns/upstream_groups/:id/default
**功能**: 设置指定分组为默认分组

**逻辑**:
1. 将所有分组的is_default设为false
2. 将指定分组的is_default设为true
3. 保存配置

**响应**: 200 OK

### 6. POST /control/dns/upstream_groups/:id/test
**功能**: 测试分组中所有上游服务器的连通性

**逻辑**:
- 对分组中的每个upstream_dns进行DNS查询测试
- 记录响应时间或错误信息
- 可以复用现有的 `testUpstream` 逻辑

**响应示例**:
```json
{
  "group_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "upstream": "223.6.6.6",
      "success": true,
      "rtt": 15
    },
    {
      "upstream": "119.29.29.29",
      "success": false,
      "error": "connection timeout"
    }
  ]
}
```


## 配置文件结构

需要在 `internal/home/config.go` 的 `dnsConfig` 结构中添加：

```go
type dnsConfig struct {
    // ... 现有字段 ...
    
    // UpstreamGroups 是DNS上游分组配置
    UpstreamGroups []UpstreamGroup `yaml:"upstream_groups"`
}

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

**重要**: 
- YAML标签在前，JSON标签在后（与AdGuard Home代码风格保持一致）
- 使用标准YAML标签，`[]string` 类型会自动序列化为块格式
- `omitempty` 标签确保空数组不会出现在配置文件中

## YAML配置示例

```yaml
dns:
  # ... 现有配置 ...
  
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
    
    - id: 550e8400-e29b-41d4-a716-446655440001
      name: 海外DNS
      enabled: true
      is_default: false
      upstream_dns:
        - https://dns.google/dns-query
        - https://cloudflare-dns.com/dns-query
      fallback_dns:
        - 8.8.8.8
        - 1.1.1.1
      bootstrap_dns:
        - 8.8.8.8
        - 1.1.1.1
      created_at: 2024-01-01T00:00:00Z
      updated_at: 2024-01-01T00:00:00Z
```

**注意**: YAML格式说明
- 使用YAML数组格式（每行一个条目，以 `-` 开头）
- 字符串值不需要引号（除非包含特殊字符）
- 遵循AdGuard Home现有配置文件的格式风格
- 与现有的 `bootstrap_dns` 配置格式保持一致


## 错误处理

### 400 Bad Request
- 请求体格式错误
- 必填字段缺失
- 字段值不符合要求

```json
{
  "error": "validation_error",
  "message": "分组名称不能为空"
}
```

### 404 Not Found
- 请求的分组ID不存在

```json
{
  "error": "not_found",
  "message": "分组不存在"
}
```

### 409 Conflict
- 尝试删除默认分组
- 分组名称重复

```json
{
  "error": "conflict",
  "message": "不能删除默认分组"
}
```

### 500 Internal Server Error
- 配置文件读写失败
- 未预期的运行时错误

```json
{
  "error": "internal_error",
  "message": "服务器内部错误，请稍后重试"
}
```

## 实现建议

### 1. 创建处理器文件
建议创建 `internal/home/dns_upstream_groups.go` 文件，包含：
- HTTP处理器函数
- 数据验证逻辑
- 配置读写逻辑

### 2. 注册路由
在 `internal/home/control.go` 的 `registerControlHandlers()` 函数中注册路由：

```go
func (web *webAPI) registerControlHandlers() {
    // ... 现有路由 ...
    
    // DNS上游分组管理
    web.httpReg.HandleFunc("GET", "/control/dns/upstream_groups", web.handleGetUpstreamGroups)
    web.httpReg.HandleFunc("POST", "/control/dns/upstream_groups", web.handleAddUpstreamGroup)
    web.httpReg.HandleFunc("PUT", "/control/dns/upstream_groups/{id}", web.handleUpdateUpstreamGroup)
    web.httpReg.HandleFunc("DELETE", "/control/dns/upstream_groups/{id}", web.handleDeleteUpstreamGroup)
    web.httpReg.HandleFunc("POST", "/control/dns/upstream_groups/{id}/default", web.handleSetDefaultGroup)
    web.httpReg.HandleFunc("POST", "/control/dns/upstream_groups/{id}/test", web.handleTestUpstreamGroup)
}
```

### 3. 配置持久化
- 每次修改分组后，调用 `config.write()` 保存配置
- 在 `initDNS()` 时加载分组配置
- 确保配置文件格式向后兼容

**YAML序列化说明**:

AdGuard Home使用 `gopkg.in/yaml.v3` 库，它会自动将 `[]string` 类型序列化为块格式（每行一个条目）。只需使用标准的YAML标签即可：

```go
type UpstreamGroup struct {
    UpstreamDNS  []string  `yaml:"upstream_dns" json:"upstream_dns"`
    FallbackDNS  []string  `yaml:"fallback_dns,omitempty" json:"fallback_dns,omitempty"`
    BootstrapDNS []string  `yaml:"bootstrap_dns,omitempty" json:"bootstrap_dns,omitempty"`
}
```

这与现有的 `bootstrap_dns` 字段使用完全相同的方式，无需额外配置。

**预期输出格式**:
```yaml
upstream_dns:
  - 223.6.6.6
  - 119.29.29.29
bootstrap_dns:
  - 9.9.9.10
  - 149.112.112.10
```

**注意**: 如果发现输出格式不正确（如 `[item1, item2]`），请检查：
1. 确保使用的是 `gopkg.in/yaml.v3` 而不是其他yaml库
2. 参考 `internal/dnsforward/config.go` 中 `BootstrapDNS` 的实现
3. 确保配置写入时使用的编码器设置正确


### 4. 默认分组逻辑
- 系统启动时检查是否存在默认分组
- 如果不存在，自动将第一个启用的分组设为默认
- 如果没有任何分组，可以创建一个默认分组

### 5. DNS解析集成（可选，未来扩展）
当前阶段只需要实现分组的CRUD操作，DNS解析集成可以在后续阶段实现：
- 在 `internal/dnsforward/` 中实现分组管理器
- 修改DNS解析流程，支持根据分组选择上游服务器
- 预留过滤器规则调用分组的接口

## 测试建议

### 单元测试
- 测试每个HTTP处理器的正常流程
- 测试各种错误情况（无效输入、资源不存在等）
- 测试默认分组逻辑
- 测试配置持久化

### 集成测试
- 测试完整的CRUD流程
- 测试并发访问
- 测试配置文件读写

### 手动测试
1. 启动AdGuard Home
2. 访问DNS设置页面
3. 测试创建、编辑、删除分组
4. 测试设置默认分组
5. 测试启用/禁用分组
6. 测试分组连通性检测
7. 重启服务，验证配置持久化

## 前端UI预览

前端已经实现了完整的UI，包括：

### 分组列表
- 表格展示所有分组
- 显示启用状态、分组名称、上游服务器地址
- 默认分组显示绿色"默认"标签
- 操作按钮：编辑、复制、删除
- 支持分页

### 分组编辑对话框
- 分组名称输入框（必填，最多50字符）
- 上游服务器列表（多行文本域，每行一个服务器）
- 备用DNS服务器列表（可选）
- Bootstrap DNS服务器列表（可选）
- 启用此分组复选框
- 设为默认分组复选框
- 检测按钮（测试连通性）
- 保存和取消按钮

### 交互反馈
- 加载状态指示器
- 成功/失败Toast通知
- 表单验证错误提示
- 删除确认对话框

## 国际化支持

前端已经添加了中文翻译键值（`client/src/__locales/zh-cn.json`），后端错误消息建议也支持国际化。

## 下一步

1. **实现后端API端点**（优先级最高）
   - 创建 `internal/home/dns_upstream_groups.go`
   - 实现所有6个API端点
   - 添加数据验证逻辑

2. **配置文件集成**
   - 修改 `internal/home/config.go`
   - 实现配置读写逻辑
   - 添加配置迁移（如果需要）

3. **测试**
   - 编写单元测试
   - 进行集成测试
   - 手动测试前后端对接

4. **DNS解析集成**（可选，后续阶段）
   - 实现分组管理器
   - 修改DNS解析流程
   - 支持过滤器规则调用分组

## 参考文件

- 前端Actions: `client/src/actions/upstreamGroups.ts`
- 前端Reducer: `client/src/reducers/upstreamGroups.ts`
- 前端API客户端: `client/src/api/Api.ts`
- 前端UI组件: `client/src/components/Settings/Dns/UpstreamGroups/`
- 设计文档: `.kiro/specs/dns-upstream-groups/design.md`
- 需求文档: `.kiro/specs/dns-upstream-groups/requirements.md`
- 任务列表: `.kiro/specs/dns-upstream-groups/tasks.md`
- 格式说明: `MOD/docs/YAML_FORMAT_NOTE.md`
- 快速参考: `MOD/docs/QUICK_REFERENCE.md`
