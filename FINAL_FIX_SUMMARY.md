# DNS Routing 最终修复总结

## 问题历史

### 问题 1: 死锁和目录混乱
- **原因**: 更新操作使用旧的 filtering 系统
- **修复**: 完全分离 DNS routing 和 filtering 系统

### 问题 2: 前端错误 "must be updated through..."
- **原因**: 前端仍使用 filtering API
- **修复**: 前端使用专用的 DNS routing API

### 问题 3: 前端不显示规则
- **原因**: API 从错误的地方读取数据
- **修复**: API 从 File Manager 读取

### 问题 4: 迁移失败
- **原因**: 文件路径错误
- **修复**: 修正迁移路径

### 问题 5: 操作不生效
- **原因**: 规则在 metadata.json 但配置保存时写回 YAML
- **修复**: 配置保存时过滤掉 DNS routing 规则

## 最终架构

### 数据存储分离

```
DNS Routing 规则:
  ├─ 存储位置: data/dns_routing_rules/metadata.json
  ├─ 规则文件: data/dns_routing_rules/*.txt
  ├─ 管理器: dnsroutingfiles.Manager
  └─ API: /control/dns_routing/*

Filtering 规则:
  ├─ 存储位置: AdGuardHome.yaml (filters 数组)
  ├─ 规则文件: data/filters/*.txt
  ├─ 管理器: filtering.DNSFilter
  └─ API: /control/filtering/*
```

### 数据流

```
启动时:
  1. File Manager 检查 metadata.json 是否存在
  2. 如果不存在，从 AdGuardHome.yaml 迁移 DNS routing 规则
  3. 迁移后，规则存储在 metadata.json 中
  4. File Manager 加载 metadata.json 中的规则

运行时:
  1. 前端调用 /control/dns_routing/* API
  2. API 调用 File Manager 方法
  3. File Manager 更新 metadata.json
  4. File Manager 通知 DNS router 重新加载

配置保存时:
  1. 系统收集所有配置
  2. 过滤掉 dns_routing: true 的 filters
  3. 只保存普通 filters 到 AdGuardHome.yaml
  4. DNS routing 规则保持在 metadata.json 中
```

## 关键修复

### 1. API 端点 (`internal/home/dns_routing.go`)

```go
// handleGetDnsRoutingRules - 从 File Manager 读取
func (web *webAPI) handleGetDnsRoutingRules(w http.ResponseWriter, r *http.Request) {
    // 从 File Manager 获取规则
    domainListRules := globalContext.dnsRoutingFileManager.GetAllDomainListRules()
    
    // 转换为响应格式
    rules := make([]DnsRoutingRule, 0, len(domainListRules))
    for _, rule := range domainListRules {
        rules = append(rules, DnsRoutingRule{
            ID:            rule.ID,
            Enabled:       rule.Enabled,
            URL:           rule.URL,
            Name:          rule.Name,
            UpstreamGroup: rule.UpstreamGroup,
            Priority:      rule.Priority,
            RulesCount:    rule.RulesCount,
            LastUpdated:   rule.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
        })
    }
    
    aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, rules)
}
```

### 2. 迁移路径 (`internal/dnsroutingfiles/manager.go`)

```go
// 修复前: data/data/filters/xxx.txt (错误!)
oldFilePath := filepath.Join(m.config.DataDir, "data", "filters", ...)

// 修复后: data/filters/xxx.txt (正确!)
oldFilePath := filepath.Join("data", "filters", ...)
```

### 3. 配置保存 (`internal/home/config.go`)

```go
if globalContext.filters != nil {
    globalContext.filters.WriteDiskConfig(config.Filtering)
    
    // 过滤掉 DNS routing 规则
    regularFilters := make([]filtering.FilterYAML, 0, len(config.Filtering.Filters))
    for _, f := range config.Filtering.Filters {
        if !f.DnsRouting {
            regularFilters = append(regularFilters, f)
        }
    }
    config.Filters = regularFilters
    
    config.WhitelistFilters = config.Filtering.WhitelistFilters
    config.UserRules = config.Filtering.UserRules
}
```

## 测试

### 完整测试流程

```powershell
# 1. 运行完整测试脚本
.\test-complete-fix.ps1

# 2. 启动服务
.\AdGuardHome_config_fix.exe

# 3. 验证 API
.\test-api-response.ps1

# 4. 验证前端
# 打开 http://localhost:3000
# 进入 DNS 路由页面
# 测试所有操作

# 5. 验证配置文件
# 检查 AdGuardHome.yaml - 不应包含 dns_routing: true 的规则
# 检查 data/dns_routing_rules/metadata.json - 应包含所有 DNS routing 规则
```

### 预期结果

#### ✅ 迁移
- 首次启动时自动迁移
- 创建 `data/dns_routing_rules/metadata.json`
- 复制规则文件到 `data/dns_routing_rules/*.txt`
- 日志显示迁移信息

#### ✅ API
- `/control/dns_routing/rules` 返回规则列表
- 所有 CRUD 操作正常工作

#### ✅ 前端
- 显示规则列表
- 启用/禁用立即生效，不触发下载
- 编辑、删除、刷新正常工作

#### ✅ 配置文件
- `AdGuardHome.yaml` 中 `filters` 数组不包含 `dns_routing: true` 的规则
- DNS routing 规则只在 `metadata.json` 中
- 配置保存后规则仍然正常工作

#### ✅ 智能更新
- 启用/禁用: 不触发下载，`last_updated` 不变
- 更新 URL: 触发下载，`last_updated` 更新
- 添加规则: 触发下载
- 手动刷新: 触发下载

## 文件清单

### 修改的文件

1. **`internal/home/dns_routing.go`**
   - `handleGetDnsRoutingRules`: 从 File Manager 读取

2. **`internal/dnsroutingfiles/manager.go`**
   - `migrateFromOldImplementation`: 修复文件路径

3. **`internal/home/config.go`**
   - `write`: 过滤 DNS routing 规则

4. **`internal/filtering/http.go`**
   - 阻止通过 filtering API 操作 DNS routing

5. **`internal/filtering/filter.go`**
   - 内部方法跳过 DNS routing

6. **`client/src/api/Api.ts`**
   - 添加 DNS routing API 方法

7. **`client/src/actions/dnsRouting.ts`**
   - 使用新 API

8. **`client/src/reducers/dnsRouting.ts`**
   - 添加 modalFilter

9. **`client/src/components/Filters/DnsRouting.tsx`**
   - 更新组件逻辑

### 新增的文件

- `AdGuardHome_config_fix.exe` - 最终修复版本
- `test-complete-fix.ps1` - 完整测试脚本
- `test-api-response.ps1` - API 测试脚本
- `test-toggle-no-download.ps1` - 启用/禁用测试
- `FINAL_FIX_SUMMARY.md` - 本文档

## 总结

通过完全分离 DNS routing 和 filtering 系统，并确保配置保存时正确处理规则，现在：

1. ✅ **数据一致性**: DNS routing 规则只在 metadata.json 中
2. ✅ **操作生效**: 所有操作立即生效并持久化
3. ✅ **配置清晰**: YAML 和 metadata.json 职责明确
4. ✅ **性能优化**: 启用/禁用不触发下载
5. ✅ **用户体验**: 前端显示正常，操作流畅

所有功能现在都应该完美工作！🎉
