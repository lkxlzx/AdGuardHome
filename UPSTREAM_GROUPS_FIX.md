# Upstream Groups 格式统一修复

## 问题描述

用户报告了两个问题：

1. **前端面板失去响应**：设置多个上游（如 `114.114.114.114223.5.5.5`）时，前端一直转圈
2. **保存格式不一致**：`upstreams` 使用字符串格式，而 `bootstrap_dns` 使用数组格式，应该统一

## 问题分析

### 1. YAML 格式不一致

在 `AdGuardHome.yaml` 中，`upstream_groups` 的 `upstreams` 字段使用了不一致的格式：

```yaml
# 第一个组使用 |- 格式（去除尾部换行符）
upstreams: |-
  114.114.114.114
  223.5.5.5

# 其他组使用单行字符串
upstreams: 8.8.8.8
```

### 2. API 超时问题

当保存 DNS 配置时，后端需要重启 DNS 服务器（见 `internal/dnsforward/http.go` 中的 `setConfigRestartable` 函数）。如果没有设置合理的超时时间，前端会一直等待响应。

### 3. 格式不一致

- `bootstrap_dns` 使用 YAML 数组格式：
  ```yaml
  bootstrap_dns:
    - 9.9.9.10
    - 149.112.112.10
  ```

- `upstreams` 字段原本是 `string` 类型，使用换行符分隔：
  ```go
  type UpstreamGroup struct {
      Upstreams string `yaml:"upstreams" json:"upstreams"`  // 旧版本
  }
  ```

这种不一致导致用户困惑，应该统一为数组格式。

## 修复内容

### 1. 统一为数组格式

将 `upstreams` 字段从字符串类型改为数组类型，与 `bootstrap_dns` 保持一致：

**后端修改**：
```go
type UpstreamGroup struct {
    Upstreams []string `yaml:"upstreams" json:"upstreams"`  // 改为数组
}
```

**YAML 格式**：
```yaml
upstream_groups:
  - id: group_1763970331409
    name: 国内DNS
    upstreams:
      - 114.114.114.114
      - 223.5.5.5
    enabled: true
    is_default: true
  - id: group_1763973311919
    name: 海外
    upstreams:
      - 8.8.8.8
    enabled: true
    is_default: false
```

### 2. 添加 API 超时配置

在 `client/src/api/Api.ts` 中添加超时配置：

```typescript
// Set a default timeout of 30 seconds for DNS config updates
if (!axiosConfig.timeout) {
    axiosConfig.timeout = path === 'dns_config' ? 30000 : 10000;
}
```

这样可以：
- DNS 配置更新请求超时时间为 30 秒
- 其他请求超时时间为 10 秒
- 避免前端无限等待

### 3. 前端数据转换

在前端添加数据转换逻辑：

**发送到后端时**（`client/src/actions/dnsConfig.ts`）：
```typescript
// 将字符串转换为数组
data.upstream_groups = config.upstream_groups.map((group: any) => ({
    ...group,
    upstreams: typeof group.upstreams === 'string' 
        ? splitByNewLine(group.upstreams) 
        : group.upstreams,
}));
```

**从后端接收时**（`client/src/reducers/dnsConfig.ts`）：
```typescript
// 将数组转换为字符串用于显示
const processedGroups = upstream_groups?.map((group: any) => ({
    ...group,
    upstreams: Array.isArray(group.upstreams) 
        ? group.upstreams.join('\n') 
        : group.upstreams,
}));
```

这样可以：
- 前端表单使用多行文本输入（用户友好）
- 后端使用数组存储（格式统一）
- 自动转换，无需用户关心

## 数据流程

### 前端 → 后端

1. 用户在 Textarea 中输入多行上游服务器地址（每行一个）
2. 前端将多行文本按换行符分割成数组：`["114.114.114.114", "223.5.5.5"]`
3. 通过 API 发送到后端（JSON 格式）
4. 后端接收数组并保存

### 后端 → 前端

1. 后端从配置文件读取 `upstreams` 字段（数组类型）
2. 通过 API 返回给前端（JSON 格式）：`["114.114.114.114", "223.5.5.5"]`
3. 前端将数组用换行符连接成字符串：`"114.114.114.114\n223.5.5.5"`
4. 显示在 Textarea 中

### 后端 → YAML

1. 后端使用 YAML 库序列化配置
2. `upstreams` 字段（数组类型）被序列化为 YAML 数组格式
3. YAML 使用 `- item` 格式保存

## 使用建议

### 正确的输入格式

在前端表单中，每行输入一个上游服务器地址：

```
114.114.114.114
223.5.5.5
8.8.8.8
```

### 错误的输入格式

❌ 不要在一行中输入多个地址：
```
114.114.114.114223.5.5.5
```

❌ 不要使用逗号分隔：
```
114.114.114.114, 223.5.5.5
```

## 测试验证

1. 启动 AdGuard Home
2. 打开前端管理界面
3. 进入 DNS 设置 → 上游 DNS 服务器
4. 添加或编辑上游组
5. 在 "上游服务器" 字段中输入多行地址
6. 点击保存
7. 验证：
   - 前端不会一直转圈
   - 配置成功保存
   - 刷新页面后数据正确显示

## 修改的文件

### 后端
- `internal/dnsforward/config.go` - 修改 `UpstreamGroup.Upstreams` 为 `[]string`
- `internal/dnsforward/upstream_groups.go` - 更新处理逻辑
- `internal/dnsforward/dnsforward.go` - 更新默认组处理逻辑

### 前端
- `client/src/actions/dnsConfig.ts` - 添加字符串到数组的转换
- `client/src/reducers/dnsConfig.ts` - 添加数组到字符串的转换
- `client/src/api/Api.ts` - 添加超时配置

### 配置
- `AdGuardHome.yaml` - 更新为数组格式
