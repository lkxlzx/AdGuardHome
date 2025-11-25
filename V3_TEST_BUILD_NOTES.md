# AdGuard Home V3 测试版本说明

## 版本信息
- **版本**: V3 Test Build
- **构建日期**: 2024-11-25
- **文件名**: AdGuardHome_v3_test.exe
- **文件大小**: 30.85 MB
- **平台**: Windows AMD64

## 本次更新内容

### 🐛 Bug修复

#### 问题#4: 域名匹配逻辑优化 ✅
- **修复内容**: 
  - 修复了 `matchDomainPattern` 函数的大小写敏感bug
  - 确保所有域名匹配逻辑使用统一的公共函数
  - 消除了代码重复

- **性能指标**:
  - `matchDomainPattern`: ~188 ns/op
  - `matchDomainWithType`: ~57 ns/op

- **测试覆盖**:
  - 新增16个单元测试用例，全部通过
  - 新增性能基准测试

- **影响范围**:
  - DNS路由规则匹配
  - 自定义域名规则匹配
  - Upstream组选择

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

## 测试建议

### 1. 基本功能测试
- [ ] 启动程序，访问Web界面
- [ ] 检查DNS路由规则页面是否正常显示
- [ ] 检查自定义域名规则页面是否正常显示

### 2. 域名匹配测试
- [ ] 测试DOMAIN类型匹配（精确匹配）
- [ ] 测试DOMAIN-SUFFIX类型匹配（后缀匹配）
- [ ] 测试DOMAIN-KEYWORD类型匹配（关键词匹配）
- [ ] 测试大小写不敏感匹配

### 3. 规则优先级测试
- [ ] 创建多个不同优先级的规则
- [ ] 验证低数字优先级规则优先匹配
- [ ] 测试自定义规则优先于过滤列表规则

### 4. 性能测试
- [ ] 测试大量规则下的DNS查询性能
- [ ] 监控内存使用情况
- [ ] 检查CPU使用率

## 已知问题

无新增已知问题。

## 升级说明

### 从V2升级
1. 停止当前运行的AdGuardHome
2. 备份配置文件 `AdGuardHome.yaml`
3. 替换可执行文件为 `AdGuardHome_v3_test.exe`
4. 启动新版本

### 配置兼容性
- ✅ 完全兼容V2配置文件
- ✅ 无需修改现有配置
- ✅ 自动迁移所有设置

## 回滚方案

如果遇到问题，可以回滚到V2版本：
1. 停止V3测试版
2. 恢复V2可执行文件
3. 使用备份的配置文件
4. 重新启动

## 技术细节

### 代码变更
- 修改: `internal/dnsforward/domain_match.go`
- 新增: `internal/dnsforward/domain_match_test.go`
- 更新: `V3_DEVELOPMENT_PLAN.md`

### 测试结果
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

## 反馈

如有问题或建议，请记录以下信息：
- 操作系统版本
- 问题描述
- 复现步骤
- 日志文件（如有）

---

**构建时间**: 2024-11-25 16:24:01  
**基于版本**: V2 (commit 2f1e0490)  
**开发分支**: v3
