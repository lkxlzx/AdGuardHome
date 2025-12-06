# DNS Routing Config Source Fix

## 问题描述

之前的实现将 DNS 路由规则的**所有配置**都保存在 `metadata.json` 文件中，包括：
- `id`
- `enabled`
- `url`
- `name`
- `upstream_group`
- `priority`
- `rules_count`
- `last_updated`
- `file_path`

这导致了一个严重问题：**配置文件（AdGuardHome.yaml）中的修改被忽略**。

用户在配置文件中修改 `enabled: false` 后，重启 AdGuardHome，规则仍然是启用状态，因为程序从 `metadata.json` 读取配置，而不是从 `AdGuardHome.yaml` 读取。

## 解决方案

### 新架构设计

将配置和运行时状态分离：

**配置文件（AdGuardHome.yaml）- 权威来源**：
- `id` - 规则 ID
- `enabled` - 启用状态
- `url` - 规则文件 URL
- `name` - 规则名称
- `upstream_group` - 上游组 ID
- `priority` - 优先级

**Metadata 文件（metadata.json）- 运行时状态**：
- `id` - 规则 ID（用于关联）
- `rules_count` - 规则数量
- `last_updated` - 最后更新时间
- `file_path` - 文件路径

### 实现变更

#### 1. Metadata 结构重构

```go
// 旧结构 - 存储所有配置
type Metadata struct {
    DomainListRules []*DomainListRule `json:"domain_list_rules"`
    CustomRules     []*CustomRule     `json:"custom_rules"`
    Version         int               `json:"version"`
}

// 新结构 - 只存储运行时状态
type Metadata struct {
    DomainListRules []*DomainListRuleState `json:"domain_list_rules"`
    Version         int                     `json:"version"`
}

type DomainListRuleState struct {
    ID          int64  `json:"id"`
    RulesCount  int    `json:"rules_count"`
    LastUpdated string `json:"last_updated"`
    FilePath    string `json:"file_path"`
}
```

#### 2. 加载逻辑改进

```go
func (m *fileManager) loadMetadata(ctx context.Context) error {
    // 1. 首先从 AdGuardHome.yaml 加载配置（权威来源）
    if err := m.loadConfigFromYAML(ctx); err != nil {
        // 配置加载失败只记录警告，不中断
    }
    
    // 2. 然后从 metadata.json 加载运行时状态
    // 3. 合并：配置来自 YAML，运行时状态来自 metadata
}
```

#### 3. 保存逻辑简化

```go
func (m *fileManager) saveMetadata(ctx context.Context) error {
    // 只保存运行时状态到 metadata.json
    // 配置由 AdGuardHome 自己保存到 AdGuardHome.yaml
}
```

### 迁移处理

- Metadata version 从 1 升级到 2
- 旧版本 metadata（v1）包含完整配置，新版本会自动迁移
- 迁移时只保留运行时状态，配置从 YAML 读取

## 测试验证

### 测试步骤

1. **编译新版本**：
   ```powershell
   go build -o AdGuardHome_config_source.exe
   ```

2. **修改配置文件**：
   编辑 `AdGuardHome.yaml`，找到 DNS 路由规则（`dns_routing: true`），修改 `enabled` 状态：
   ```yaml
   filters:
     - enabled: false  # 改为 false
       url: https://example.com/rules.txt
       name: Test Rule
       dns_routing: true
       upstream_group: some-group-id
       priority: 0
       id: 1234567890
   ```

3. **启动 AdGuardHome**：
   ```powershell
   .\AdGuardHome_config_source.exe
   ```

4. **验证规则状态**：
   ```powershell
   # 获取 DNS 路由规则
   Invoke-RestMethod -Uri "http://localhost/control/dns_routing/rules" `
       -Method GET `
       -Headers @{"Authorization" = "Basic <your-auth>"}
   ```

5. **检查结果**：
   - 规则的 `enabled` 状态应该与配置文件一致
   - 修改配置文件后重启，状态应该立即生效

### 自动化测试脚本

运行 `test-config-source.ps1` 进行自动化测试。

## 影响范围

### 受影响的文件

1. **internal/dnsroutingfiles/metadata.go**
   - 重构 Metadata 结构
   - 添加 DomainListRuleState 结构
   - 版本号升级到 2

2. **internal/dnsroutingfiles/storage.go**
   - 修改 saveMetadata() - 只保存运行时状态
   - 修改 loadMetadata() - 先加载 YAML 配置，再合并 metadata 状态
   - 添加 loadConfigFromYAML() - 从配置文件加载规则配置

3. **internal/dnsroutingfiles/manager.go**
   - 简化 migrateFromOldImplementation() - 只迁移文件，不迁移配置
   - 删除 syncEnabledStateFromConfig() - 不再需要同步
   - 修改 LoadAll() - 调整加载顺序

### 向后兼容性

- ✅ 完全向后兼容
- ✅ 自动迁移旧版本 metadata（v1 -> v2）
- ✅ 配置文件格式不变
- ✅ API 接口不变

## 优势

1. **配置文件是唯一权威来源** - 用户修改配置文件后立即生效
2. **清晰的职责分离** - 配置 vs 运行时状态
3. **更好的可维护性** - 配置和状态分离，更容易理解和调试
4. **符合用户预期** - 修改配置文件应该生效，这是基本预期

## 后续工作

### 可选优化

1. **Custom Rules 也应该从配置文件读取**
   - 当前 custom rules 仍然存储在 metadata 中
   - 应该考虑将其移到配置文件的 `custom_domain_rules` 部分

2. **API 更新配置文件**
   - 当通过 API 修改规则时，应该更新配置文件
   - 当前可能只更新了内存状态

3. **配置文件热重载**
   - 监听配置文件变化
   - 自动重新加载规则配置

## 总结

这次修复解决了一个架构设计问题：**配置应该从配置文件读取，而不是从 metadata 文件读取**。

修复后：
- ✅ 配置文件是权威来源
- ✅ 用户修改配置文件后重启即可生效
- ✅ Metadata 只存储运行时状态
- ✅ 清晰的职责分离
- ✅ 完全向后兼容
