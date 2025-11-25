# DNS路由规则配置保存修复

## 问题描述

添加DNS路由规则时：
- 后端日志显示规则已成功下载和保存
- 但配置文件 `AdGuardHome.yaml` 中没有 `dns_routing_filters` 字段
- 前端刷新后看不到添加的规则

## 问题原因

虽然在 `internal/filtering/filtering.go` 中添加了 `DnsRoutingFilters` 字段，但配置文件的结构定义和同步逻辑中缺少这个字段：

1. **配置结构缺失**：`internal/home/config.go` 中的 `configuration` 结构体没有 `DnsRoutingFilters` 字段
2. **配置同步缺失**：配置保存和加载时没有同步 `DnsRoutingFilters`
3. **过滤器加载缺失**：`filtering.New()` 函数中没有加载 `DnsRoutingFilters`

## 解决方案

### 1. 添加配置结构字段

**文件**：`internal/home/config.go`

```go
type configuration struct {
    // ... 其他字段 ...
    
    Filters           []filtering.FilterYAML `yaml:"filters"`
    WhitelistFilters  []filtering.FilterYAML `yaml:"whitelist_filters"`
    DnsRoutingFilters []filtering.FilterYAML `yaml:"dns_routing_filters"`  // 新增
    UserRules         []string               `yaml:"user_rules"`
    
    // ... 其他字段 ...
}
```

### 2. 添加配置保存同步

**文件**：`internal/home/config.go`

```go
func (c *configuration) write() (err error) {
    // ... 其他代码 ...
    
    globalContext.filters.WriteDiskConfig(config.Filtering)
    config.Filters = config.Filtering.Filters
    config.WhitelistFilters = config.Filtering.WhitelistFilters
    config.DnsRoutingFilters = config.Filtering.DnsRoutingFilters  // 新增
    config.UserRules = config.Filtering.UserRules
    
    // ... 其他代码 ...
}
```

### 3. 添加配置加载同步

**文件**：`internal/home/home.go`

```go
func initFiltering(ctx context.Context, config *configuration, ...) (err error) {
    // ... 其他代码 ...
    
    conf.DataDir = filepath.Join(workDir, dataDir)
    conf.Filters = slices.Clone(config.Filters)
    conf.WhitelistFilters = slices.Clone(config.WhitelistFilters)
    conf.DnsRoutingFilters = slices.Clone(config.DnsRoutingFilters)  // 新增
    conf.UserRules = slices.Clone(config.UserRules)
    
    // ... 其他代码 ...
}
```

### 4. 添加过滤器加载

**文件**：`internal/filtering/filtering.go`

```go
func New(c *Config, blockFilters []Filter) (d *DNSFilter, err error) {
    // ... 其他代码 ...
    
    d.loadFilters(ctx, d.conf.Filters)
    d.loadFilters(ctx, d.conf.WhitelistFilters)
    d.loadFilters(ctx, d.conf.DnsRoutingFilters)  // 新增
    
    // ... 其他代码 ...
}
```

## 配置文件结构

修复后，`AdGuardHome.yaml` 将包含独立的 `dns_routing_filters` 字段：

```yaml
filters:
  - enabled: true
    url: https://example.com/blocklist.txt
    name: 黑名单规则
    id: 1

whitelist_filters:
  - enabled: true
    url: https://example.com/allowlist.txt
    name: 白名单规则
    id: 2

dns_routing_filters:
  - enabled: true
    url: https://example.com/routing.txt
    name: DNS路由规则
    id: 3
    upstream_group: "group-id"
```

## 数据流

### 添加DNS路由规则
1. 用户在前端添加规则
2. API调用：`POST /control/filtering/add_url` with `dns_routing=true`
3. 后端调用 `filterAddDnsRouting()`
4. 规则添加到 `Config.DnsRoutingFilters`
5. 下载并保存规则文件到 `data/filters/[id].txt`
6. 调用 `ConfModifier.Apply()` 触发配置保存
7. `config.write()` 将 `DnsRoutingFilters` 同步到 `configuration.DnsRoutingFilters`
8. 配置保存到 `AdGuardHome.yaml`

### 加载DNS路由规则
1. AdGuard Home 启动
2. 读取 `AdGuardHome.yaml`
3. 解析 `dns_routing_filters` 字段到 `configuration.DnsRoutingFilters`
4. 调用 `initFiltering()`
5. 将 `configuration.DnsRoutingFilters` 克隆到 `filtering.Config.DnsRoutingFilters`
6. 调用 `filtering.New()`
7. 调用 `d.loadFilters(ctx, d.conf.DnsRoutingFilters)`
8. 从 `data/filters/[id].txt` 加载规则内容
9. 规则生效

## 测试步骤

### 1. 测试配置保存
1. 停止 AdGuard Home
2. 删除或备份 `AdGuardHome.yaml`
3. 启动 AdGuard Home
4. 添加一个DNS路由规则
5. 停止 AdGuard Home
6. 检查 `AdGuardHome.yaml`
7. 验证：文件中包含 `dns_routing_filters` 字段和添加的规则

### 2. 测试配置加载
1. 确保 `AdGuardHome.yaml` 中有 `dns_routing_filters` 字段
2. 重启 AdGuard Home
3. 进入"DNS路由"页面
4. 验证：之前添加的规则正确显示

### 3. 测试规则持久化
1. 添加多个DNS路由规则
2. 重启 AdGuard Home
3. 验证：所有规则都保留并正确显示

### 4. 测试规则文件
1. 添加DNS路由规则
2. 检查 `data/filters/` 目录
3. 验证：规则文件已创建（如 `1763999702.txt`）
4. 重启后验证：规则文件仍然存在且被正确加载

## 配置同步流程

```
内存中的过滤器配置 (filtering.Config)
    ↓
    ↓ (保存时)
    ↓
配置文件结构 (configuration)
    ↓
    ↓ (序列化)
    ↓
YAML文件 (AdGuardHome.yaml)
    ↓
    ↓ (加载时)
    ↓
配置文件结构 (configuration)
    ↓
    ↓ (初始化)
    ↓
内存中的过滤器配置 (filtering.Config)
```

## 注意事项

1. **配置迁移**：如果之前有DNS路由规则保存在 `whitelist_filters` 中，需要手动迁移到 `dns_routing_filters`

2. **ID唯一性**：确保所有过滤器（黑名单、白名单、DNS路由）的ID全局唯一

3. **文件路径**：所有过滤器的规则文件都保存在 `data/filters/` 目录下，使用ID作为文件名

4. **配置备份**：修改前建议备份 `AdGuardHome.yaml`

## 编译和部署

```bash
# 编译后端
go build -o AdGuardHome.exe

# 停止旧版本
# 启动新版本
./AdGuardHome.exe
```

## 相关文件

- `internal/home/config.go` - 配置结构定义和保存
- `internal/home/home.go` - 配置加载和初始化
- `internal/filtering/filtering.go` - 过滤器初始化和加载

## 技术细节

### 配置同步时机

1. **保存时机**：
   - 添加/删除/修改过滤器后
   - 调用 `ConfModifier.Apply(ctx)`
   - 触发 `config.write()`

2. **加载时机**：
   - AdGuard Home 启动时
   - 读取 `AdGuardHome.yaml`
   - 调用 `initFiltering()`

### 字段映射

| 内存结构 | 配置文件 | YAML字段 |
|---------|---------|----------|
| `filtering.Config.Filters` | `configuration.Filters` | `filters` |
| `filtering.Config.WhitelistFilters` | `configuration.WhitelistFilters` | `whitelist_filters` |
| `filtering.Config.DnsRoutingFilters` | `configuration.DnsRoutingFilters` | `dns_routing_filters` |
| `filtering.Config.UserRules` | `configuration.UserRules` | `user_rules` |
