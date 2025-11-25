# AdGuard Home V2 - 发布说明

## 版本信息
- **版本**: v2.0.0-dev
- **发布日期**: 2024-11-25
- **基于**: AdGuard Home v0.107.x

## 🎯 主要新功能

### 1. DNS路由规则优先级支持 ⭐

为DNS路由过滤器添加了完整的优先级（Priority）功能，解决了规则冲突问题。

**功能特点**:
- ✅ 优先级值范围: 0-100（数字越小，优先级越高）
- ✅ 前端UI支持设置和编辑优先级
- ✅ 双重保护机制：加载时排序 + 匹配时选择
- ✅ 日志中显示匹配的规则和优先级
- ✅ 配置文件正确保存和加载

**使用场景**:
```yaml
dns_routing_filters:
  - name: 特定域名规则
    priority: 10  # 高优先级
    url: https://example.com/specific.yaml
    
  - name: 通用域名规则
    priority: 30  # 低优先级
    url: https://example.com/general.yaml
```

当`qq.com`同时匹配两个规则时，系统会使用priority=10的规则。

### 2. 自定义域名路由规则

支持在Web界面中直接添加自定义域名路由规则，无需编辑配置文件。

**功能特点**:
- ✅ 三种匹配类型：
  - **精确匹配** (DOMAIN): 完全匹配域名
  - **后缀匹配** (DOMAIN-SUFFIX): 匹配域名及其子域名
  - **关键字匹配** (DOMAIN-KEYWORD): 匹配包含关键字的域名
- ✅ 可视化管理界面
- ✅ 支持启用/禁用
- ✅ 支持编辑和删除

**示例**:
```yaml
custom_domain_rules:
  - domain: example.com
    match_type: DOMAIN-SUFFIX
    upstream_group: group_domestic
    enabled: true
```

### 3. 规则自动更新间隔

为DNS路由过滤器添加了自定义更新间隔功能。

**功能特点**:
- ✅ 支持按分钟设置更新间隔
- ✅ 0表示禁用自动更新
- ✅ 独立于全局过滤器更新间隔
- ✅ 前端UI支持设置

## 🐛 Bug修复

### 修复的问题

1. **Priority字段保存问题**
   - 修复了编辑过滤器时priority字段被清空的问题
   - 修复了前端editFilter action中的字段删除bug

2. **自定义规则enabled字段**
   - 修复了新增规则默认未启用的问题
   - 修复了编辑规则后enabled状态丢失的问题

3. **前端状态一致性**
   - 改进了自定义规则提交后的状态同步
   - 确保UI始终反映服务器实际状态

4. **翻译缺失**
   - 添加了自定义规则对话框按钮的中文翻译
   - 修复了显示翻译键而不是实际文本的问题

5. **配置保存**
   - 修复了WriteDiskConfig未保存UpstreamGroups等字段的问题
   - 确保所有DNS路由配置正确持久化

## 🔧 技术改进

### 架构优化

1. **双重优先级保护**
   - 加载时按priority排序过滤器
   - 匹配时从多个规则中选择最高优先级

2. **代码质量**
   - 修复了代码审查中发现的严重问题
   - 改进了错误处理和状态管理
   - 优化了类型安全

3. **性能优化**
   - 使用`slices.SortFunc`进行高效排序
   - 优化了规则匹配逻辑

## 📦 构建和部署

### 支持的平台

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Windows | AMD64 | AdGuardHome_windows_amd64.exe |
| Windows | ARM64 | AdGuardHome_windows_arm64.exe |
| Linux | AMD64 | AdGuardHome_linux_amd64 |
| Linux | ARM64 | AdGuardHome_linux_arm64 |
| Linux | ARMv7 | AdGuardHome_linux_armv7 |
| macOS | Intel | AdGuardHome_darwin_amd64 |
| macOS | Apple Silicon | AdGuardHome_darwin_arm64 |

### 构建方法

```bash
# Windows
.\build-all-platforms.bat

# 或手动构建单个平台
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o AdGuardHome_linux_amd64
```

## 📚 文档

### 新增文档

- `FEATURE_PRIORITY_COMPLETE.md` - Priority功能完整文档
- `FEATURE_AUTO_UPDATE_INTERVAL.md` - 自动更新间隔功能文档
- `BUG_FIX_DNS_ROUTING_CLASH.md` - DNS路由冲突修复文档
- `CODE_REVIEW_REPORT.md` - 代码审查报告
- `dist/README.md` - 构建文件说明

### 更新的文档

- `V2_RELEASE_SUMMARY.md` - V2版本总结
- `DEPLOYMENT_CHECKLIST_V2.md` - 部署检查清单
- `AdGuardHome_v2_README.md` - V2版本说明

## 🔄 升级指南

### 从V1升级到V2

1. **备份配置**
   ```bash
   cp AdGuardHome.yaml AdGuardHome.yaml.backup
   ```

2. **停止服务**
   ```bash
   # Linux/macOS
   sudo systemctl stop AdGuardHome
   
   # Windows
   net stop AdGuardHome
   ```

3. **替换可执行文件**
   - 下载对应平台的V2版本
   - 替换旧的可执行文件

4. **启动服务**
   ```bash
   # Linux/macOS
   sudo systemctl start AdGuardHome
   
   # Windows
   net start AdGuardHome
   ```

5. **验证功能**
   - 检查Web界面是否正常
   - 验证DNS路由规则是否工作
   - 查看日志确认无错误

### 配置迁移

V2版本完全兼容V1的配置文件，无需手动迁移。新功能的配置会自动添加默认值。

### 新增配置项

```yaml
dns_routing_filters:
  - priority: 0  # 新增：优先级字段
    update_interval: 0  # 新增：更新间隔（分钟）

custom_domain_rules:  # 新增：自定义域名规则
  - domain: example.com
    match_type: DOMAIN-SUFFIX
    upstream_group: group_id
    enabled: true
```

## ⚠️ 注意事项

### 重要提示

1. **Priority默认值**: 所有现有规则的priority默认为0
2. **更新间隔**: 现有规则的update_interval默认为0（禁用自动更新）
3. **配置备份**: 升级前务必备份配置文件
4. **端口权限**: Linux/macOS需要root权限绑定53端口

### 已知限制

1. Priority范围未在后端强制验证（仅前端限制0-100）
2. 相同priority的规则匹配顺序未定义
3. 自定义规则不支持批量导入

## 🚀 未来计划

### 短期计划（1-2周）

- [ ] 添加priority范围后端验证
- [ ] 实现相同priority的次级排序规则
- [ ] 优化域名匹配性能（使用Trie树）
- [ ] 添加规则冲突检测和警告

### 中期计划（1个月）

- [ ] 支持自定义规则批量导入/导出
- [ ] 添加规则测试工具
- [ ] 实现规则使用统计
- [ ] 优化Web界面性能

### 长期计划

- [ ] 支持规则分组管理
- [ ] 添加规则模板功能
- [ ] 实现智能规则推荐
- [ ] 支持规则版本控制

## 📊 性能指标

### 构建大小

- 二进制文件: ~30-32 MB（已优化）
- 前端资源: ~2 MB（已压缩）

### 运行性能

- 内存占用: ~50-100 MB
- CPU使用: <5%（空闲时）
- DNS查询延迟: <10ms（本地规则）

## 🤝 贡献

感谢所有贡献者的支持！

### 主要贡献

- Priority功能设计和实现
- 自定义规则UI开发
- Bug修复和代码审查
- 文档编写和维护

## 📄 许可证

AdGuard Home is licensed under the GNU General Public License v3.0.

## 🔗 相关链接

- [GitHub Repository](https://github.com/AdguardTeam/AdGuardHome)
- [Official Documentation](https://github.com/AdguardTeam/AdGuardHome/wiki)
- [Community Forum](https://forum.adguard.com/)

---

**发布日期**: 2024-11-25  
**版本**: v2.0.0-dev  
**分支**: v2
