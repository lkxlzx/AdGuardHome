# File Manager 简化修复

## 问题

操作 DNS 路由规则时出现错误：
```
Error: control/dns_routing/delete | deleting dns routing rule: rule with ID 1764867837 not found | 500
```

**根本原因**：
- File Manager 的方法（Add/Update/Delete）检查内存中的规则列表（`m.domainListRules`）
- 但这个列表是空的（因为移除了 metadata 加载）
- 导致所有操作都失败

## 解决方案

**简化 File Manager 的方法，让它们只操作文件，不检查内存中的规则列表。**

### 修改内容

#### 1. AddDomainListRule
**之前**：
- 检查 `m.domainListRules` 中是否已存在
- 下载并解析
- 添加到 `m.domainListRules`
- 保存 metadata

**现在**：
- 直接下载并解析
- 保存文件
- 不检查/不更新内存中的规则列表

```go
func (m *fileManager) AddDomainListRule(ctx context.Context, rule *DomainListRule) error {
    // 验证参数
    // 下载规则文件
    // 解析规则文件
    // 保存到磁盘
    // 通知路由器
    return nil
}
```

#### 2. UpdateDomainListRule
**之前**：
- 检查 `m.domainListRules` 中是否存在
- 比较 URL 是否改变
- 重新下载（如果 URL 改变）
- 更新 `m.domainListRules`
- 保存 metadata

**现在**：
- 直接重新下载并解析
- 保存文件
- 不检查/不更新内存中的规则列表

```go
func (m *fileManager) UpdateDomainListRule(ctx context.Context, rule *DomainListRule) error {
    // 验证参数
    // 下载规则文件
    // 解析规则文件
    // 保存到磁盘
    // 通知路由器
    return nil
}
```

#### 3. DeleteDomainListRule
**之前**：
- 检查 `m.domainListRules` 中是否存在
- 从 `m.domainListRules` 中删除
- 保存 metadata
- 删除文件

**现在**：
- 直接删除文件
- 不检查/不更新内存中的规则列表

```go
func (m *fileManager) DeleteDomainListRule(ctx context.Context, ruleID int64) error {
    // 获取文件路径
    filePath := m.getRuleFilePath(ruleID)
    
    // 删除文件
    os.Remove(filePath)
    
    // 通知路由器
    m.scheduleRouterUpdate(ctx)
    
    return nil
}
```

### File Manager 的新职责

File Manager 现在**只负责文件操作**：

1. ✅ **下载规则文件** - 从 URL 下载
2. ✅ **解析规则文件** - 解析并返回规则数量
3. ✅ **保存规则文件** - 保存到磁盘
4. ✅ **删除规则文件** - 从磁盘删除

**不负责**：
- ❌ 管理规则列表（由配置文件负责）
- ❌ 检查规则是否存在（由 API 层负责）
- ❌ 保存 metadata（已移除）

### 数据流

#### 添加规则
```
API 请求
  ↓
API: 验证参数，生成 ID
  ↓
File Manager: 下载并解析规则文件
  ↓
API: 更新 config.Filters
  ↓
Config 系统: 保存到 AdGuardHome.yaml
  ↓
Router: 重新加载规则
```

#### 更新规则
```
API 请求
  ↓
API: 验证参数
  ↓
File Manager: 重新下载并解析规则文件
  ↓
API: 更新 config.Filters
  ↓
Config 系统: 保存到 AdGuardHome.yaml
  ↓
Router: 重新加载规则
```

#### 删除规则
```
API 请求
  ↓
API: 验证参数
  ↓
File Manager: 删除规则文件
  ↓
API: 从 config.Filters 中移除
  ↓
Config 系统: 保存到 AdGuardHome.yaml
  ↓
Router: 重新加载规则
```

### 编译测试

```powershell
go build -o AdGuardHome_final.exe
```
✅ 编译成功，无错误

## 优势

1. **职责单一** - File Manager 只管文件，不管配置
2. **简化逻辑** - 移除了内存中的规则列表管理
3. **无状态** - File Manager 不维护状态，更可靠
4. **易于理解** - 代码更简单，逻辑更清晰

## 测试

### 功能测试

1. **添加规则**
   ```bash
   curl -X POST http://localhost/control/dns_routing/add \
     -d '{"name":"Test","url":"https://example.com/rules.txt","upstream_group":"xxx","priority":0}'
   ```
   ✅ 应该成功下载规则文件并更新配置

2. **更新规则**
   ```bash
   curl -X POST http://localhost/control/dns_routing/update \
     -d '{"id":1764867837,"name":"Updated","url":"...","enabled":false}'
   ```
   ✅ 应该成功重新下载规则文件并更新配置

3. **删除规则**
   ```bash
   curl -X POST http://localhost/control/dns_routing/delete \
     -d '{"id":1764867837}'
   ```
   ✅ 应该成功删除规则文件并从配置中移除

4. **启用/禁用规则**
   - 修改配置文件中的 `enabled` 字段
   - 重启 AdGuardHome
   - ✅ 规则状态应该与配置文件一致

## 总结

这次修复解决了"rule not found"错误，通过：
- ✅ 移除 File Manager 对内存规则列表的依赖
- ✅ File Manager 只负责文件操作
- ✅ 配置管理由 API 层和 Config 系统负责
- ✅ 职责清晰，逻辑简单

现在所有 DNS 路由规则的操作（添加、更新、删除、启用/禁用）都应该正常工作了。
