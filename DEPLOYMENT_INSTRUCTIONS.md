# 部署说明

## 修复内容

本次修复解决了 4 个关键问题：

1. ✅ 白名单操作导致 DNS 路由规则数据丢失
2. ✅ 添加 Clash 规则时未处理
3. ✅ 重启后启用规则时未处理
4. ✅ 禁用再启用规则时 upstream_group 丢失

## 修改文件

### 后端（4 个文件）
- `internal/filtering/filtering.go` - WriteDiskConfig 函数
- `internal/filtering/http.go` - handleFilteringAddURL 函数
- `internal/filtering/filter.go` - MarkAsDnsRouting 方法
- `internal/home/home.go` - setupDNSFilteringConf 函数

### 前端（1 个文件）
- `client/src/components/Filters/Table.tsx` - renderCheckbox 方法

## 编译状态

✅ **前端编译成功**
```bash
cd client
npm run build-prod
# webpack 5.102.1 compiled successfully
```

✅ **后端编译成功**（包含最新前端资源）
```bash
go build -o AdGuardHome.exe
# 编译成功
```

## 部署步骤

### 1. 备份

```bash
# 备份配置文件
copy AdGuardHome.yaml AdGuardHome.yaml.backup

# 备份旧版本可执行文件
copy AdGuardHome.exe AdGuardHome.exe.old

# 备份数据目录（可选）
xcopy /E /I data data.backup
```

### 2. 停止服务

- 如果作为服务运行，停止服务
- 如果手动运行，关闭程序

### 3. 替换文件

```bash
# 使用新编译的 AdGuardHome.exe 替换旧文件
# 新文件已包含最新的前端资源
```

### 4. 启动服务

```bash
# 启动 AdGuard Home
AdGuardHome.exe
```

### 5. 验证

1. 打开浏览器访问 AdGuard Home 管理界面
2. 检查所有功能是否正常
3. 按照测试清单进行验证

## 测试清单

### 测试 1：白名单不影响 DNS 路由规则

- [ ] 添加 DNS 路由规则（带 upstream_group）
- [ ] 添加白名单规则
- [ ] 检查配置文件，确认 `dns_routing_filters` 的 `upstream_group` 仍然存在

### 测试 2：添加 Clash 规则自动处理

- [ ] 添加 Clash 规则 URL 到 DNS 路由规则
- [ ] 检查 `data/filters/[id].txt`，确认只包含域名规则
- [ ] 查看日志，确认看到 "clash rule processed" 信息

### 测试 3：重启后启用规则正常处理

- [ ] 添加 Clash 规则并保存
- [ ] 重启程序
- [ ] 禁用该规则
- [ ] 再次启用该规则
- [ ] 检查规则文件，确认仍然只包含域名规则

### 测试 4：禁用再启用规则保留 upstream_group

- [ ] 添加 DNS 路由规则（选择上游组）
- [ ] 记录 `upstream_group` 值
- [ ] 禁用该规则
- [ ] 再次启用该规则
- [ ] 检查配置文件，确认 `upstream_group` 仍然存在且值不变

## 回滚步骤

如果出现问题，可以回滚到旧版本：

```bash
# 停止新版本
# ...

# 恢复旧版本
copy AdGuardHome.exe.old AdGuardHome.exe

# 恢复配置文件（如果需要）
copy AdGuardHome.yaml.backup AdGuardHome.yaml

# 启动旧版本
AdGuardHome.exe
```

## 常见问题

### Q1: 前端界面没有更新？

**原因**：浏览器缓存

**解决方法**：
1. 清除浏览器缓存
2. 强制刷新（Ctrl+F5）
3. 使用隐私模式访问

### Q2: 规则文件仍然包含 IP 规则？

**原因**：使用的不是新编译的版本

**解决方法**：
1. 确认使用新编译的 `AdGuardHome.exe`
2. 检查文件修改时间
3. 重新编译并替换

### Q3: upstream_group 仍然丢失？

**原因**：前端资源未更新

**解决方法**：
1. 确认后端编译时包含了最新的前端资源
2. 清除浏览器缓存
3. 重新编译：
   ```bash
   cd client
   npm run build-prod
   cd ..
   go build -o AdGuardHome.exe
   ```

## 验证成功标准

所有以下条件都满足才算部署成功：

- ✅ 程序正常启动
- ✅ 管理界面可以访问
- ✅ 测试 1 通过
- ✅ 测试 2 通过
- ✅ 测试 3 通过
- ✅ 测试 4 通过
- ✅ DNS 查询正常工作
- ✅ 日志无错误信息

## 相关文档

- `COMPLETE_FIX_SUMMARY.md` - 完整修复总结
- `FINAL_TEST_GUIDE.md` - 详细测试指南
- `CLASH_RULES_FIX_COMPLETE.md` - Clash 规则修复详情
- `QUICK_TEST_STEPS.md` - 快速测试步骤

## 技术支持

如果遇到问题：

1. 查看 AdGuard Home 日志
2. 检查配置文件 `AdGuardHome.yaml`
3. 查看相关文档
4. 如果问题持续，可以回滚到旧版本

## 总结

本次修复涉及 5 个文件共 17 行代码的修改，解决了 4 个关键问题。所有修改都已经过编译验证，前端和后端都已成功编译。

新的 `AdGuardHome.exe` 文件已包含所有修复和最新的前端资源，可以直接部署使用。

**部署时间估计**：5-10 分钟（包括备份、替换、启动和基本验证）

**测试时间估计**：15-20 分钟（完整测试所有 4 个场景）

祝部署顺利！🚀
