# Metadata 移除总结

## 问题

DNS 路由规则的配置被保存在两个地方：
1. **AdGuardHome.yaml** - 配置文件
2. **metadata.json** - File Manager 的 metadata 文件

这导致两个文件不同步，用户修改配置文件后不生效。

## 解决方案

**移除 metadata.json，配置文件是唯一数据源。**

### 修改内容

#### 1. storage.go
- `saveMetadata()` - 改为空函数（不再保存 metadata）
- `loadMetadata()` - 改为空函数（不再加载 metadata）
- 移除 `loadFromConfig()` 和 `ConfigFilter`
- 移除未使用的 imports（json, yaml, time）

#### 2. manager.go
- `LoadAll()` - 保持接口不变，但内部不再加载 metadata
- `migrateFromOldImplementation()` - 保持迁移逻辑

#### 3. metadata.go
- 保持不变（为了向后兼容）

### File Manager 的新职责

File Manager 现在只负责：
1. ✅ **下载规则文件**
2. ✅ **解析规则文件**
3. ✅ **管理规则文件**（保存、删除）
4. ✅ **管理内存中的规则列表**（临时状态）

**不负责**：
- ❌ 配置持久化（由 AdGuardHome 的 config 系统负责）
- ❌ Metadata 文件管理

### 数据流

#### 添加规则
```
API 请求
  ↓
1. API: 验证参数，生成 ID
  ↓
2. File Manager: 下载并解析规则文件
  ↓
3. File Manager: 更新内存中的规则列表
  ↓
4. API: 更新 config.Filters
  ↓
5. Config 系统: 保存到 AdGuardHome.yaml
  ↓
6. Router: 重新加载规则
```

#### 启动加载
```
启动
  ↓
1. Config 系统: 从 AdGuardHome.yaml 加载配置
  ↓
2. File Manager: LoadAll()（现在只加载 custom rules）
  ↓
3. API: 从 config.Filters 读取规则配置
  ↓
4. File Manager: 加载规则文件
  ↓
5. Router: 加载所有规则
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
    rules_count: 1234
    last_updated: 2025-12-05T10:30:00Z
```

### 向后兼容

- ✅ 旧的 metadata.json 文件会被忽略
- ✅ 迁移逻辑保留（从旧的 data/filters 目录迁移文件）
- ✅ API 接口不变
- ✅ 配置文件格式不变

### 优势

1. **单一数据源** - AdGuardHome.yaml 是唯一的配置来源
2. **用户友好** - 修改配置文件后重启即可生效
3. **简化架构** - 移除了 metadata 文件的复杂性
4. **职责清晰** - File Manager 只管文件，Config 系统管配置

## 测试

### 编译测试
```powershell
go build -o AdGuardHome_check.exe
```
✅ 编译成功，无错误

### 诊断测试
```powershell
getDiagnostics(["internal/dnsroutingfiles/*.go", "internal/home/dns_routing.go"])
```
✅ 无诊断错误

### 功能测试

1. **修改配置文件测试**
   - 编辑 AdGuardHome.yaml
   - 修改 DNS 路由规则的 `enabled` 状态
   - 重启 AdGuardHome
   - 验证规则状态是否与配置文件一致

2. **API 测试**
   - 通过 API 添加规则
   - 检查配置文件是否更新
   - 重启后规则是否保留

3. **迁移测试**
   - 从旧版本升级
   - 验证规则文件是否正确迁移

## 后续工作

### 可选优化

1. **完全移除 metadata 相关代码**
   - 删除 metadata.go
   - 删除 metadata_test.go
   - 清理所有 metadata 相关的测试

2. **使用 SimpleManager**
   - 用 `manager_simple.go` 替换复杂的 manager.go
   - 进一步简化代码

3. **Custom Rules 也从配置文件读取**
   - 当前 custom rules 仍然存储在文件中
   - 可以考虑移到配置文件的 `custom_domain_rules` 部分

## 总结

这次修改解决了配置文件和 metadata 文件不同步的问题，通过：
- ✅ 移除 metadata 文件的保存/加载逻辑
- ✅ 配置文件成为唯一数据源
- ✅ 保持 API 接口不变
- ✅ 保持向后兼容

File Manager 现在只负责文件操作，配置管理由 AdGuardHome 的 config 系统负责，职责清晰，易于维护。
