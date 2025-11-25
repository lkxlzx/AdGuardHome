# AdGuard Home V2 部署清单

## ✅ 编译完成

- [x] 前端编译成功 (webpack 5.102.1)
- [x] 后端编译成功 (Go 1.25.4)
- [x] 生成 V2 可执行文件: `AdGuardHome_v2.exe` (30.8 MB)
- [x] 代码推送到 GitHub (v2 分支)

---

## 📦 文件清单

### 可执行文件
- ✅ `AdGuardHome_v2.exe` - V2 版本 (30.8 MB)
- ✅ `AdGuardHome.exe` - V1 备份版本 (39.2 MB)

### 文档文件
- ✅ `AdGuardHome_v2_README.md` - V2 版本说明
- ✅ `VERSION_COMPARISON.md` - 版本对比指南
- ✅ `CODE_REVIEW_REPORT.md` - 代码审查报告
- ✅ `V2_FIX_PROGRESS.md` - 修复进度报告
- ✅ `README_DNS_UPSTREAM_GROUPS.md` - 功能文档

### 配置文件
- ✅ `AdGuardHome.yaml` - 当前配置（兼容 V1 和 V2）

---

## 🔍 V2 改进总结

### 已修复问题 (7/16)

#### 🔴 严重问题 (2/3)
1. ✅ **并发安全** - 使用 ID 标识规则
2. ✅ **状态一致性** - API 成功后再更新状态
3. ⏳ 内存泄漏 - 待后续版本

#### 🟡 重要问题 (2/5)
4. ✅ **代码重复** - 提取公共函数
5. ⏳ 性能优化 - 待后续版本
6. ⏳ 错误处理 - 待后续版本
7. ⏳ 类型安全 - 待后续版本

#### 🟢 中等问题 (2/5)
8. ✅ **日志优化** - Debug 级别
9. ✅ **冗余清理** - 删除未使用函数
10. ⏳ 组件重构 - 待后续版本
11. ⏳ 输入验证 - 部分完成
12. ⏳ 加载状态 - 待后续版本

#### 🔵 低优先级 (1/3)
13. ⏳ 注释完善 - 待后续版本
14. ✅ **魔法数字** - 使用常量
15. ⏳ 翻译键 - 待后续版本

---

## 🚀 部署步骤

### 测试环境部署

#### 1. 准备工作
```bash
# 创建测试目录
mkdir AdGuardHome_v2_test
cd AdGuardHome_v2_test

# 复制文件
copy ..\AdGuardHome_v2.exe .
copy ..\AdGuardHome.yaml .
```

#### 2. 首次启动
```bash
# 启动 V2
AdGuardHome_v2.exe

# 访问管理界面
# http://localhost:3000
```

#### 3. 功能验证
- [ ] 登录管理界面
- [ ] 检查 DNS 路由规则列表
- [ ] 添加新的 DNS 路由规则
- [ ] 编辑现有规则
- [ ] 删除规则
- [ ] 测试自定义域名规则
- [ ] 验证上游分组功能
- [ ] 检查日志输出

#### 4. 性能测试
- [ ] 并发添加多个规则
- [ ] 同时编辑不同规则
- [ ] 大量 DNS 查询测试
- [ ] 内存占用监控
- [ ] CPU 使用率监控

---

### 生产环境部署

#### 前置条件
- [ ] 测试环境验证通过
- [ ] 备份当前配置和数据
- [ ] 准备回滚方案
- [ ] 通知用户维护窗口

#### 1. 备份
```bash
# 备份配置
copy AdGuardHome.yaml AdGuardHome.yaml.backup_$(Get-Date -Format "yyyyMMdd_HHmmss")

# 备份数据
xcopy /E /I data data_backup_$(Get-Date -Format "yyyyMMdd_HHmmss")

# 备份 V1 可执行文件
copy AdGuardHome.exe AdGuardHome_v1_backup.exe
```

#### 2. 停止服务
```bash
# 停止当前运行的服务
taskkill /F /IM AdGuardHome.exe

# 等待进程完全退出
timeout /t 5
```

#### 3. 部署 V2
```bash
# 方案 A: 替换主文件
ren AdGuardHome.exe AdGuardHome_v1.exe
copy AdGuardHome_v2.exe AdGuardHome.exe

# 方案 B: 直接使用 V2
# 直接运行 AdGuardHome_v2.exe
```

#### 4. 启动服务
```bash
# 启动 V2
AdGuardHome.exe
# 或
AdGuardHome_v2.exe
```

#### 5. 验证
- [ ] 服务正常启动
- [ ] 管理界面可访问
- [ ] DNS 解析正常工作
- [ ] 规则正常加载
- [ ] 日志正常输出
- [ ] 无错误信息

#### 6. 监控
```bash
# 监控日志
Get-Content -Path "data\querylog.json" -Wait -Tail 50

# 监控进程
Get-Process AdGuardHome* | Select-Object Name, CPU, WorkingSet
```

---

## 🔄 回滚方案

### 快速回滚
```bash
# 1. 停止 V2
taskkill /F /IM AdGuardHome_v2.exe

# 2. 启动 V1
AdGuardHome_v1.exe
```

### 完整回滚
```bash
# 1. 停止 V2
taskkill /F /IM AdGuardHome.exe

# 2. 恢复 V1 可执行文件
del AdGuardHome.exe
copy AdGuardHome_v1_backup.exe AdGuardHome.exe

# 3. 恢复配置（如果需要）
copy AdGuardHome.yaml.backup AdGuardHome.yaml

# 4. 恢复数据（如果需要）
rmdir /S /Q data
xcopy /E /I data_backup data

# 5. 启动 V1
AdGuardHome.exe
```

---

## 📊 监控指标

### 关键指标
- [ ] 服务可用性: 99.9%+
- [ ] DNS 查询响应时间: < 50ms
- [ ] 内存使用: < 200MB
- [ ] CPU 使用: < 10%
- [ ] 错误率: < 0.1%

### 日志检查
```bash
# 检查错误日志
Select-String -Path "data\*.log" -Pattern "error|ERROR|Error" -Context 2

# 检查警告日志
Select-String -Path "data\*.log" -Pattern "warn|WARN|Warning" -Context 2
```

---

## ⚠️ 注意事项

### 已知限制
1. V2 与 V1 不能同时运行（端口冲突）
2. 配置文件完全兼容，无需修改
3. 数据目录结构相同
4. API 向后兼容（支持 URL 和 ID）

### 风险评估
- **低风险**: 配置兼容性 ✅
- **低风险**: 数据迁移 ✅
- **低风险**: 功能回归 ✅
- **中风险**: 性能影响 ⚠️ (需监控)
- **低风险**: 回滚难度 ✅

---

## 📞 支持信息

### 问题排查
1. 检查日志文件
2. 对比 V1 和 V2 行为
3. 查看 GitHub Issues
4. 参考文档

### 文档链接
- [V2 版本说明](AdGuardHome_v2_README.md)
- [版本对比](VERSION_COMPARISON.md)
- [代码审查](CODE_REVIEW_REPORT.md)
- [修复进度](V2_FIX_PROGRESS.md)

---

## ✅ 部署确认

### 测试环境
- [ ] 编译完成
- [ ] 功能测试通过
- [ ] 性能测试通过
- [ ] 文档准备完成

### 生产环境
- [ ] 备份完成
- [ ] 部署完成
- [ ] 验证通过
- [ ] 监控正常
- [ ] 用户通知

---

## 📝 部署记录

| 环境 | 时间 | 版本 | 操作人 | 状态 | 备注 |
|------|------|------|--------|------|------|
| 开发 | 2024-11-25 12:12 | V2 | - | ✅ | 编译完成 |
| 测试 | - | - | - | ⏳ | 待部署 |
| 生产 | - | - | - | ⏳ | 待部署 |

---

## 🎯 下一步计划

### 短期 (1-2 周)
- [ ] 在测试环境运行 V2
- [ ] 收集用户反馈
- [ ] 修复发现的问题
- [ ] 准备 V2.1 版本

### 中期 (1 个月)
- [ ] 性能优化 (问题 #5)
- [ ] 完善错误处理 (问题 #6)
- [ ] 类型安全改进 (问题 #7)
- [ ] 发布 V2.2 版本

### 长期 (2-3 个月)
- [ ] 组件重构 (问题 #10)
- [ ] 完整的单元测试
- [ ] 性能基准测试
- [ ] 文档完善

---

**状态**: ✅ V2 编译完成，已推送到 GitHub，等待部署测试

**推荐**: 先在测试环境验证 1-2 天，确认无问题后再部署到生产环境
