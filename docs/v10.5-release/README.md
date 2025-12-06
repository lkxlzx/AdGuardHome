# AdGuard Home v10.5 开发计划

## 版本信息
- **版本号**: v10.5
- **基于**: v10.4
- **状态**: 开发中
- **创建日期**: 2025-12-06

## v10.4 完成的功能

### DNS 路由系统
✅ 完整的 DNS 路由规则支持
✅ 上游组管理
✅ 规则优先级控制
✅ 自动更新规则列表
✅ 路由缓存优化

### 性能优化
✅ Trie 树优化的域名匹配
✅ 批量查询优化
✅ 智能缓存系统
✅ 并发处理优化

### 日志改进
✅ 修复误导性的规则加载日志
✅ 准确显示实际加载的规则数量
✅ 降低不必要的日志噪音

## v10.5 开发计划

### 待规划功能
- [ ] 待定

### 技术债务
- [ ] 修复 dnsroutingfiles 测试用例
- [ ] 完善单元测试覆盖率
- [ ] 优化构建流程

### 文档改进
- [ ] 完善 API 文档
- [ ] 添加用户使用指南
- [ ] 更新部署文档

## 开发指南

### 分支策略
- `v10.5`: 主开发分支
- 功能开发: 从 v10.5 创建 feature 分支
- Bug 修复: 从 v10.5 创建 fix 分支

### 提交规范
```
feat: 新功能
fix: Bug 修复
docs: 文档更新
style: 代码格式
refactor: 重构
perf: 性能优化
test: 测试相关
chore: 构建/工具相关
```

### 测试要求
- 所有新功能必须包含单元测试
- 核心功能需要集成测试
- 性能敏感代码需要基准测试

## 构建说明

### 开发构建
```bash
go build -o AdGuardHome.exe
```

### 生产构建
```bash
# 多平台构建
.\build-multiplatform.ps1

# 单平台构建
go build -ldflags "-s -w" -o AdGuardHome.exe
```

### 前端构建
```bash
cd client
npm run build-prod
```

## 测试说明

### 运行测试
```bash
# 所有测试
go test ./...

# 特定包
go test ./internal/dnsrouting/...

# 带覆盖率
go test -cover ./...

# 基准测试
go test -bench=. ./internal/dnsrouting/...
```

## 发布流程

1. 完成所有功能开发和测试
2. 更新版本号和 CHANGELOG
3. 创建 release tag
4. 构建多平台二进制文件
5. 生成 release notes
6. 发布到 GitHub Releases

## 联系方式

如有问题或建议，请提交 Issue。
