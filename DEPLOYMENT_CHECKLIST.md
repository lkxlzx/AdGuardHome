# 部署清单

## 编译完成 ✅

- [x] 前端编译完成（2025-11-24 19:07）
- [x] 后端编译完成（2025-11-24 19:07）
- [x] 前端资源已嵌入可执行文件
- [x] 可执行文件大小：32.2 MB

## 文件清单

### 必需文件

1. **AdGuardHome.exe** - 主程序（已更新）
2. **AdGuardHome.yaml** - 配置文件（已更新格式）

### 文档文件

3. **VERSION_UPDATE.md** - 版本更新说明
4. **BUILD_INSTRUCTIONS.md** - 编译说明
5. **UPSTREAM_GROUPS_FIX.md** - 详细修复说明
6. **QUICK_FIX_SUMMARY.md** - 快速总结

## 部署步骤

### 1. 备份当前配置

```bash
copy AdGuardHome.yaml AdGuardHome.yaml.backup
```

### 2. 停止旧版本

如果 AdGuard Home 正在运行，先停止它。

### 3. 替换可执行文件

```bash
# 备份旧版本
copy AdGuardHome.exe AdGuardHome.exe.old

# 使用新版本
# （新的 AdGuardHome.exe 已经在当前目录）
```

### 4. 更新配置文件

检查 `AdGuardHome.yaml` 中的 `upstream_groups` 格式：

```yaml
upstream_groups:
  - id: group_1763970331409
    name: 国内DNS
    upstreams:
      - 114.114.114.114
      - 223.5.5.5
    enabled: true
    is_default: true
```

### 5. 启动新版本

```bash
.\AdGuardHome.exe
```

### 6. 验证功能

- [ ] 访问管理界面：http://localhost:80
- [ ] 检查上游组是否正常显示
- [ ] 尝试添加新的上游组
- [ ] 尝试编辑现有上游组
- [ ] 验证 DNS 解析功能
- [ ] 检查日志是否有错误

## 验证清单

### 前端验证

- [ ] 管理界面可以正常访问
- [ ] DNS 设置页面可以打开
- [ ] 上游组列表正常显示
- [ ] 可以添加新的上游组
- [ ] 可以编辑现有上游组
- [ ] 可以删除上游组
- [ ] 可以设置默认组
- [ ] 保存配置不会卡死（30秒内完成）

### 后端验证

- [ ] 程序正常启动
- [ ] 配置文件正确加载
- [ ] DNS 解析功能正常
- [ ] 上游组切换正常
- [ ] 日志没有错误信息

### 配置验证

- [ ] `AdGuardHome.yaml` 格式正确
- [ ] `upstream_groups` 使用数组格式
- [ ] 所有上游组都能正常工作
- [ ] 默认组设置正确

## 性能测试

### DNS 解析测试

```bash
# 测试默认上游组
nslookup google.com 127.0.0.1

# 测试特定域名路由（如果配置了）
nslookup baidu.com 127.0.0.1
```

### 压力测试

```bash
# 使用 dnsperf 或类似工具测试
# 确保在高负载下也能正常工作
```

## 回滚计划

如果出现问题，按以下步骤回滚：

1. 停止新版本
2. 恢复旧的可执行文件：
   ```bash
   copy AdGuardHome.exe.old AdGuardHome.exe
   ```
3. 恢复旧的配置文件：
   ```bash
   copy AdGuardHome.yaml.backup AdGuardHome.yaml
   ```
4. 启动旧版本

## 常见问题

### Q: 前端显示空白？
A: 清除浏览器缓存（Ctrl+F5）

### Q: 上游组不显示？
A: 检查 `AdGuardHome.yaml` 格式是否正确

### Q: 保存配置时卡死？
A: 检查网络连接，等待最多30秒

### Q: DNS 解析失败？
A: 检查上游服务器地址是否正确

## 监控建议

部署后建议监控以下指标：

- DNS 查询响应时间
- 上游服务器可用性
- 错误日志数量
- 内存使用情况
- CPU 使用情况

## 支持

如有问题，请查看：
- 日志文件
- 浏览器控制台
- 网络请求（F12 开发者工具）

## 完成确认

- [ ] 所有验证项通过
- [ ] 性能测试正常
- [ ] 文档已更新
- [ ] 备份已创建
- [ ] 回滚计划已准备

---

**部署日期**：2025年11月24日  
**部署人员**：_____________  
**验证人员**：_____________  
**状态**：□ 成功  □ 失败  □ 需要回滚
