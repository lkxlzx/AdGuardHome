# AdGuard Home V3 最新版本说明

## 版本信息
- **版本**: V3 Latest Build
- **构建日期**: 2024-11-25 17:19
- **文件名**: AdGuardHome_v3_latest.exe
- **文件大小**: 30.85 MB
- **平台**: Windows AMD64

## 本次更新内容

### 🐛 Bug修复

#### ✅ 问题#4: 域名匹配逻辑优化
- **修复内容**: 
  - 修复了 `matchDomainPattern` 函数的大小写敏感bug
  - 确保所有域名匹配逻辑使用统一的公共函数
  - 消除了代码重复

- **性能指标**:
  - `matchDomainPattern`: ~188 ns/op
  - `matchDomainWithType`: ~57 ns/op

- **测试覆盖**:
  - 新增16个单元测试用例，全部通过 ✅
  - 新增性能基准测试 ✅

- **影响范围**:
  - DNS路由规则匹配
  - 自定义域名规则匹配
  - Upstream组选择

#### ✅ 问题#7: 前端类型安全性提升
- **修复内容**:
  - 移除了所有 `as any` 类型断言（4处）
  - 添加了 `DnsConfig` 接口定义
  - 扩展了 `DnsRoutingProps` 接口

- **改进效果**:
  - ✅ 提升了TypeScript类型安全性
  - ✅ 改善了IDE自动补全和类型提示
  - ✅ 增强了代码可维护性
  - ✅ 减少了潜在的运行时错误

- **影响范围**:
  - DNS路由规则管理界面
  - 自定义域名规则管理界面

## V2功能保留

本版本保留了V2的所有功能：

### ✅ 核心功能
1. DNS路由规则优先级（Priority）支持
2. 自定义域名路由规则管理
3. 规则自动更新间隔设置
4. 双重优先级保护机制

### ✅ 已修复的V2问题
1. Priority字段保存问题
2. 自定义规则enabled字段问题
3. 前端状态一致性问题
4. 翻译缺失问题

## 与之前版本的区别

### 相比 V3_test.exe
- ✅ 新增：前端类型安全性修复（问题#7）
- ✅ 改进：更好的TypeScript类型支持
- ✅ 改进：更安全的前端代码

### 相比 V2
- ✅ 修复：域名匹配大小写敏感bug
- ✅ 修复：前端类型断言不安全问题
- ✅ 改进：代码质量和可维护性
- ✅ 新增：完整的单元测试和基准测试

## 测试建议

### 1. 基本功能测试
- [ ] 启动程序，访问Web界面
- [ ] 检查DNS路由规则页面是否正常显示
- [ ] 检查自定义域名规则页面是否正常显示
- [ ] 验证所有按钮和表单正常工作

### 2. 域名匹配测试（问题#4修复验证）
- [ ] 测试DOMAIN类型匹配（精确匹配）
- [ ] 测试DOMAIN-SUFFIX类型匹配（后缀匹配）
- [ ] 测试DOMAIN-KEYWORD类型匹配（关键词匹配）
- [ ] 测试大小写不敏感匹配（example.com vs Example.COM）

### 3. 前端功能测试（问题#7修复验证）
- [ ] 添加自定义域名规则
- [ ] 编辑自定义域名规则
- [ ] 删除自定义域名规则
- [ ] 启用/禁用自定义域名规则
- [ ] 验证所有操作后数据正确保存

### 4. 规则优先级测试
- [ ] 创建多个不同优先级的规则
- [ ] 验证低数字优先级规则优先匹配
- [ ] 测试自定义规则优先于过滤列表规则

### 5. 性能测试
- [ ] 测试大量规则下的DNS查询性能
- [ ] 监控内存使用情况
- [ ] 检查CPU使用率
- [ ] 长时间运行稳定性测试

## 技术细节

### 后端代码变更
- 修改: `internal/dnsforward/domain_match.go`
- 新增: `internal/dnsforward/domain_match_test.go`

### 前端代码变更
- 修改: `client/src/components/Filters/DnsRouting.tsx`
  - 添加 `DnsConfig` 接口
  - 扩展 `DnsRoutingProps` 接口
  - 移除所有 `as any` 类型断言

### 测试结果

#### 后端测试
```
=== RUN   TestMatchDomainPattern
--- PASS: TestMatchDomainPattern (0.00s)
=== RUN   TestMatchDomainWithType
--- PASS: TestMatchDomainWithType (0.00s)
PASS

Benchmark Results:
BenchmarkMatchDomainPattern-8      6398600    187.9 ns/op    32 B/op    1 allocs/op
BenchmarkMatchDomainWithType-8    19725129     56.59 ns/op    0 B/op    0 allocs/op
```

#### 前端测试
```
✅ TypeScript编译通过
✅ 无类型诊断错误
✅ Webpack构建成功
✅ 所有 as any 已移除
```

## 升级说明

### 从V2升级
1. 停止当前运行的AdGuardHome
2. 备份配置文件 `AdGuardHome.yaml`
3. 替换可执行文件为 `AdGuardHome_v3_latest.exe`
4. 启动新版本

### 从V3_test.exe升级
1. 停止当前运行的程序
2. 替换可执行文件为 `AdGuardHome_v3_latest.exe`
3. 启动新版本（无需修改配置）

### 配置兼容性
- ✅ 完全兼容V2配置文件
- ✅ 完全兼容V3_test配置文件
- ✅ 无需修改现有配置
- ✅ 自动迁移所有设置

## 回滚方案

如果遇到问题，可以回滚：

### 回滚到V3_test
1. 停止V3_latest
2. 使用 `AdGuardHome_v3_test.exe`
3. 重新启动

### 回滚到V2
1. 停止V3_latest
2. 恢复V2可执行文件
3. 使用备份的配置文件
4. 重新启动

## 已知问题

目前无新增已知问题。

## 性能指标

| 指标 | 目标 | 实际 |
|------|------|------|
| DNS查询延迟 | < 5ms | 待测试 |
| CPU使用率 | < 10% | 待测试 |
| 内存占用 | < 150MB | 待测试 |
| 域名匹配速度 | < 200ns | 188ns ✅ |
| 前端类型安全 | 100% | 100% ✅ |

## 开发进度

### 已完成（第1周）
- [x] 问题#4: 域名匹配逻辑优化 ✅
- [x] 问题#7: 前端类型安全性提升 ✅

### 待完成
- [ ] 问题#9: 删除未使用的函数
- [ ] 问题#11: 添加输入验证
- [ ] 问题#12: 添加加载状态
- [ ] 功能1: Priority范围验证

## 相关文档

- `V3_DEVELOPMENT_PLAN.md` - 完整开发计划
- `ISSUE_7_FIX_DETAILS.md` - 问题#7详细修复说明
- `V3_TEST_BUILD_NOTES.md` - V3测试版说明
- `CODE_REVIEW_REPORT.md` - 代码审查报告

## 快速开始

### 启动程序
```cmd
AdGuardHome_v3_latest.exe
```

### 访问Web界面
```
http://localhost:3000
```

### 查看详细日志
```cmd
AdGuardHome_v3_latest.exe -v
```

## 反馈

如有问题或建议，请记录以下信息：
- 操作系统版本
- 问题描述
- 复现步骤
- 日志文件（如有）
- 截图（如有）

## 测试清单

完成以下测试项目：

### 基本功能 ✓
- [ ] 程序正常启动
- [ ] Web界面可访问
- [ ] DNS查询正常工作
- [ ] 日志正常记录

### 域名匹配（问题#4）✓
- [ ] DOMAIN类型匹配正常
- [ ] DOMAIN-SUFFIX类型匹配正常
- [ ] DOMAIN-KEYWORD类型匹配正常
- [ ] 大小写不敏感匹配正常

### 前端功能（问题#7）✓
- [ ] 添加自定义规则正常
- [ ] 编辑规则正常
- [ ] 删除规则正常
- [ ] 启用/禁用规则正常
- [ ] 数据保存正确

### 规则管理 ✓
- [ ] 可以添加DNS路由规则
- [ ] 可以编辑规则
- [ ] 可以删除规则
- [ ] 可以启用/禁用规则
- [ ] 规则优先级正常工作

### 性能 ✓
- [ ] DNS查询响应快速（< 5ms）
- [ ] CPU使用率正常（< 10%）
- [ ] 内存使用正常（< 150MB）
- [ ] 长时间运行稳定

---

**构建时间**: 2024-11-25 17:19:52  
**基于版本**: V2 (commit 2f1e0490)  
**开发分支**: v3  
**包含修复**: 问题#4 + 问题#7

**祝使用愉快！** 🎉
