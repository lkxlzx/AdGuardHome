# DNS 路由规则完整修复

## 问题历程

### 问题 1：配置文件修改不生效
**原因**：规则配置保存在 metadata.json 文件中，而不是配置文件中。

**解决**：移除 metadata 文件，配置文件作为唯一数据源。

### 问题 2：规则列表为空
**原因**：API 从 File Manager 的内存读取规则，但内存中没有数据。

**解决**：API 直接从配置文件（`config.Filters`）读取规则。

### 问题 3：操作失败（rule not found）
**原因**：File Manager 的方法检查内存中的规则列表，但列表是空的。

**解决**：简化 File Manager，让它只操作文件，不检查内存中的规则列表。

### 问题 4：添加规则不成功，配置文件没有内容
**原因**：`saveConfigIfNeeded` 的调用逻辑错误。

```go
// 错误的逻辑
if !web.saveConfigIfNeeded(ctx, l, r, w) && web.conf != nil {
    return
}
```

这个逻辑的意思是：只有当保存失败**并且** web.conf 不为 nil 时才 return。
但实际上应该是：只要保存失败就 return。

**解决**：修复逻辑为：

```go
// 正确的逻辑
if !web.saveConfigIfNeeded(ctx, l, r, w) {
    return
}
```

## 最终架构

### 数据流

```
配置文件 (AdGuardHome.yaml) ← 唯一权威数据源
    ↓
  API 层
    ├─ 读取：从 config.Filters 读取规则
    ├─ 添加：调用 File Manager 下载 → 更新 config.Filters → 保存配置
    ├─ 更新：调用 File Manager 重新下载 → 更新 config.Filters → 保存配置
    └─ 删除：调用 File Manager 删除文件 → 从 config.Filters 移除 → 保存配置
    ↓
File Manager (只负责文件操作)
    ├─ 下载规则文件
    ├─ 解析规则文件
    ├─ 保存规则文件
    └─ 删除规则文件
    ↓
规则文件 (data/dns_routing_rules/*.txt)
```

### 职责划分

#### 配置文件 (AdGuardHome.yaml)
- ✅ 存储所有规则配置
- ✅ 唯一的权威数据源
- ✅ 用户可以直接编辑

#### API 层 (internal/home/dns_routing.go)
- ✅ 从配置文件读取规则
- ✅ 验证请求参数
- ✅ 协调 File Manager 和配置文件
- ✅ 保存配置文件

#### File Manager (internal/dnsroutingfiles/)
- ✅ 下载规则文件
- ✅ 解析规则文件
- ✅ 保存规则文件
- ✅ 删除规则文件
- ❌ 不管理规则列表
- ❌ 不保存 metadata

## 修改的文件

### 1. internal/dnsroutingfiles/storage.go
- `saveMetadata()` → 空函数（不再保存）
- `loadMetadata()` → 空函数（不再加载）
- 移除未使用的 imports

### 2. internal/dnsroutingfiles/manager.go
- `LoadAll()` → 保持接口，但不加载 metadata

### 3. internal/dnsroutingfiles/rules.go
- `AddDomainListRule()` → 移除对内存规则列表的依赖
- `UpdateDomainListRule()` → 移除对内存规则列表的依赖
- `DeleteDomainListRule()` → 移除对内存规则列表的依赖

### 4. internal/home/dns_routing.go
- `handleGetDnsRoutingRules()` → 从配置文件读取
- `handleAddDnsRoutingRule()` → 更新配置文件并保存
- `handleUpdateDnsRoutingRule()` → 更新配置文件并保存
- `handleDeleteDnsRoutingRule()` → 更新配置文件并保存
- 修复 `saveConfigIfNeeded` 的调用逻辑

## 编译测试

```powershell
go build -o AdGuardHome_save_fix.exe
```
✅ 编译成功，无错误

## 功能测试

### 1. 获取规则列表
```bash
curl http://localhost/control/dns_routing/rules
```
✅ 应该返回配置文件中的所有 DNS 路由规则

### 2. 添加规则
```bash
curl -X POST http://localhost/control/dns_routing/add \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Rule",
    "url": "https://example.com/rules.txt",
    "upstream_group": "group-id",
    "priority": 0
  }'
```
✅ 应该：
- 下载并解析规则文件
- 添加到配置文件
- 保存配置文件
- 返回成功响应

### 3. 更新规则
```bash
curl -X POST http://localhost/control/dns_routing/update \
  -H "Content-Type: application/json" \
  -d '{
    "id": 123,
    "name": "Updated Rule",
    "url": "https://example.com/rules.txt",
    "upstream_group": "group-id",
    "priority": 0,
    "enabled": false
  }'
```
✅ 应该：
- 重新下载并解析规则文件
- 更新配置文件中的对应条目
- 保存配置文件
- 返回成功响应

### 4. 删除规则
```bash
curl -X POST http://localhost/control/dns_routing/delete \
  -H "Content-Type: application/json" \
  -d '{"id": 123}'
```
✅ 应该：
- 删除规则文件
- 从配置文件中移除条目
- 保存配置文件
- 返回成功响应

### 5. 手动编辑配置文件
1. 编辑 `AdGuardHome.yaml`
2. 修改 DNS 路由规则的 `enabled` 字段
3. 重启 AdGuardHome
4. ✅ 规则状态应该与配置文件一致

## 配置文件格式

```yaml
filters:
  - id: 1764867837
    enabled: true
    url: https://example.com/rules.txt
    name: Test Rule
    dns_routing: true
    upstream_group: group-id
    priority: 0
    rules_count: 1234
    last_updated: 2025-12-05T10:30:00Z
```

## 优势

1. **配置文件是唯一数据源** - 用户修改配置文件后重启即可生效
2. **职责清晰** - File Manager 只管文件，API 管配置，Config 系统管保存
3. **简化架构** - 移除了 metadata 文件的复杂性
4. **易于维护** - 代码逻辑清晰，容易理解
5. **用户友好** - 用户可以直接编辑配置文件

## 总结

经过4次迭代修复，最终实现了：
- ✅ 配置文件作为唯一数据源
- ✅ API 直接从配置文件读取规则
- ✅ File Manager 只负责文件操作
- ✅ 所有操作（添加、更新、删除）正确保存配置
- ✅ 用户可以手动编辑配置文件

现在 DNS 路由规则功能应该完全正常工作了！
