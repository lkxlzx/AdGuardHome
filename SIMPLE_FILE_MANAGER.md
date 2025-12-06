# 简化的 File Manager 架构

## 问题

之前的 File Manager 承担了太多职责：
- ❌ 配置管理（保存到 metadata.json）
- ❌ 规则的增删改查
- ❌ 状态管理
- ✅ 下载规则文件
- ✅ 解析规则文件

这导致配置文件和 metadata 文件不同步的问题。

## 新架构

### File Manager 的职责

**只负责文件操作**：
1. 下载规则文件（从 URL）
2. 解析规则文件（转换格式）
3. 保存规则文件到磁盘
4. 删除规则文件

**不负责**：
- 配置管理 → 由 AdGuardHome 的 config 系统负责
- 规则增删改查 → 由 API 层负责
- 状态管理 → 存储在配置文件中

### 接口设计

```go
type SimpleManager interface {
    // 下载并解析规则文件
    DownloadAndParseRule(ctx context.Context, ruleID int64, url string) (rulesCount int, err error)
    
    // 获取规则文件路径
    GetRuleFilePath(ruleID int64) string
    
    // 删除规则文件
    DeleteRuleFile(ctx context.Context, ruleID int64) error
    
    // 加载现有规则文件
    LoadRuleFile(ctx context.Context, ruleID int64) (rules []dnsrouting.ParsedRule, err error)
}
```

### 数据流

#### 添加规则

```
API 请求
  ↓
1. API 层：验证参数，生成 ID
  ↓
2. File Manager：下载并解析规则文件
  ↓
3. API 层：更新 config.Filters（包含 rules_count）
  ↓
4. Config 系统：保存到 AdGuardHome.yaml
  ↓
5. Router：重新加载规则
```

#### 更新规则

```
API 请求
  ↓
1. API 层：验证参数
  ↓
2. File Manager：重新下载并解析规则文件
  ↓
3. API 层：更新 config.Filters（包含 rules_count, last_updated）
  ↓
4. Config 系统：保存到 AdGuardHome.yaml
  ↓
5. Router：重新加载规则
```

#### 删除规则

```
API 请求
  ↓
1. API 层：验证参数
  ↓
2. File Manager：删除规则文件
  ↓
3. API 层：从 config.Filters 中移除
  ↓
4. Config 系统：保存到 AdGuardHome.yaml
  ↓
5. Router：重新加载规则
```

#### 启动加载

```
启动
  ↓
1. Config 系统：从 AdGuardHome.yaml 加载配置
  ↓
2. 遍历 config.Filters（dns_routing=true）
  ↓
3. File Manager：加载每个规则文件
  ↓
4. Router：加载所有规则
```

### 配置文件格式

所有配置都在 `AdGuardHome.yaml` 中：

```yaml
filters:
  - id: 1764867837
    enabled: true
    url: https://example.com/rules.txt
    name: CN Rules
    dns_routing: true
    upstream_group: 708dd52d-f6fb-4863-bfce-75a7f33c899d
    priority: 0
    rules_count: 1234        # 由 File Manager 返回
    last_updated: 2025-12-05T10:30:00Z  # 由 API 层设置
```

### 优势

1. **单一数据源**：配置文件是唯一的权威来源
2. **职责清晰**：File Manager 只管文件，不管配置
3. **简单直接**：没有 metadata 文件，没有同步问题
4. **易于理解**：数据流清晰，容易维护

## 实现步骤

### 1. 使用新的 SimpleManager

在 `internal/home/dns.go` 中：

```go
// 初始化
globalContext.dnsRoutingFileManager = dnsroutingfiles.NewSimpleManager(
    dnsroutingfiles.SimpleConfig{
        DataDir: workDir,
        Logger: baseLogger.With("component", "dns_routing_files"),
    },
)
```

### 2. 更新 API 层

在 `internal/home/dns_routing.go` 中：

```go
func (web *webAPI) handleAddDnsRoutingRule(w http.ResponseWriter, r *http.Request) {
    // 1. 验证参数
    // 2. 生成 ID
    // 3. 调用 File Manager 下载
    rulesCount, err := globalContext.dnsRoutingFileManager.DownloadAndParseRule(ctx, newID, req.URL)
    
    // 4. 更新配置
    config.Lock()
    config.Filters = append(config.Filters, filtering.FilterYAML{
        Filter: filtering.Filter{
            ID:            newID,
            Name:          req.Name,
            URL:           req.URL,
            DnsRouting:    true,
            UpstreamGroup: req.UpstreamGroup,
            Priority:      req.Priority,
            RulesCount:    rulesCount,
            LastUpdated:   time.Now(),
        },
        Enabled: true,
    })
    config.Unlock()
    
    // 5. 保存配置
    web.conf.Apply(ctx)
    
    // 6. 重新加载路由器
    reloadDnsRoutingRules(ctx)
}
```

### 3. 启动时加载

在 `internal/home/dns.go` 中：

```go
func reloadDnsRoutingRules(ctx context.Context, baseLogger *slog.Logger) {
    config.RLock()
    filters := config.Filters
    config.RUnlock()
    
    for _, filter := range filters {
        if !filter.DnsRouting || !filter.Enabled {
            continue
        }
        
        // 加载规则文件
        rules, err := globalContext.dnsRoutingFileManager.LoadRuleFile(ctx, filter.ID)
        if err != nil {
            baseLogger.ErrorContext(ctx, "failed to load rule file", "id", filter.ID, "error", err)
            continue
        }
        
        // 更新路由器
        globalContext.dnsServer.UpdateDnsRoutingRules(filter.ID, filter.UpstreamGroup, filter.Priority, rules)
    }
}
```

## 迁移

对于已有的 metadata.json 文件：
1. 读取 metadata.json
2. 将 rules_count 和 last_updated 更新到配置文件
3. 删除 metadata.json（可选）

## 总结

新架构的核心思想：
- **File Manager = 文件工具**（下载、解析、保存、删除）
- **Config = 配置管理**（AdGuardHome.yaml）
- **API = 业务逻辑**（协调 File Manager 和 Config）

这样职责清晰，没有同步问题，易于维护。
