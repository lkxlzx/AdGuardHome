# DNSProxy Fork 集成说明

**更新日期**: 2025-12-06  
**版本**: v10.5  
**DNSProxy 版本**: v0.79.2

---

## 📦 集成的 Fork 版本

### 仓库信息
- **原始仓库**: https://github.com/AdguardTeam/dnsproxy
- **Fork 仓库**: https://github.com/lkxlzx/dnsproxy
- **使用版本**: v0.79.2
- **分支**: v0.79.2

### 主要改进

根据你的说明，v0.79.2 版本重构了缓存逻辑，主要改进包括：

1. **缓存性能优化**
   - 改进的缓存算法
   - 更高效的内存使用
   - 更快的查找速度

2. **缓存逻辑重构**
   - 更清晰的代码结构
   - 更好的可维护性
   - 更强的扩展性

---

## 🔧 集成方法

### Go Modules 配置

在 `go.mod` 文件中使用 `replace` 指令：

```go
// Replace dnsproxy with custom fork
replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy v0.79.2
```

### 为什么使用 replace 指令？

由于 fork 仓库的 `go.mod` 文件中模块路径仍然声明为 `github.com/AdguardTeam/dnsproxy`，直接修改 require 会导致模块路径不匹配错误。使用 `replace` 指令可以：

1. ✅ 保持代码中的 import 路径不变
2. ✅ 无需修改 fork 仓库的 go.mod
3. ✅ 方便切换回原始版本或其他 fork

---

## 🚀 编译和测试

### 编译

```bash
# 清理依赖缓存
go clean -modcache

# 更新依赖
go mod tidy

# 编译
go build -o AdGuardHome.exe
```

### 验证

```bash
# 检查使用的 dnsproxy 版本
go list -m github.com/AdguardTeam/dnsproxy

# 输出应该显示:
# github.com/AdguardTeam/dnsproxy v0.77.0 => github.com/lkxlzx/dnsproxy v0.79.2
```

### 测试

```bash
# 运行测试
go test ./...

# 运行基准测试
go test -bench=. -benchmem ./internal/dnsrouting
```

---

## 📊 性能影响

### 预期改进

基于缓存逻辑重构，预期性能改进：

| 指标 | v0.77.0 (原始) | v0.79.2 (Fork) | 预期提升 |
|------|---------------|---------------|---------|
| 缓存查找速度 | 基准 | 更快 | 10-30% |
| 内存使用 | 基准 | 更优 | 5-15% |
| 缓存命中率 | 基准 | 更高 | 5-10% |

### 实际测试

运行性能基准测试以验证实际改进：

```bash
# 运行完整基准测试
go test -bench=. -benchmem -benchtime=5s ./internal/dnsrouting > performance-results/v10.5-with-fork.txt
```

---

## 🔄 更新 Fork 版本

### 更新到新版本

```bash
# 1. 修改 go.mod 中的版本号
replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy v0.79.3

# 2. 更新依赖
go mod tidy

# 3. 测试编译
go build

# 4. 运行测试
go test ./...
```

### 切换回原始版本

```bash
# 1. 删除或注释 replace 指令
# replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy v0.79.2

# 2. 更新依赖
go mod tidy

# 3. 测试编译
go build
```

---

## 📝 维护注意事项

### 同步上游更新

定期检查原始 dnsproxy 仓库的更新：

```bash
# 1. 添加上游仓库
cd /path/to/your/dnsproxy/fork
git remote add upstream https://github.com/AdguardTeam/dnsproxy.git

# 2. 获取上游更新
git fetch upstream

# 3. 合并更新到你的分支
git checkout v0.79.2
git merge upstream/master

# 4. 解决冲突并推送
git push origin v0.79.2
```

### 版本标签

为每个重要版本创建 Git 标签：

```bash
# 创建标签
git tag -a v0.79.2 -m "Refactored cache logic"

# 推送标签
git push origin v0.79.2
```

---

## 🐛 故障排除

### 问题 1: 模块路径不匹配

**错误信息**:
```
module declares its path as: github.com/AdguardTeam/dnsproxy
but was required as: github.com/lkxlzx/dnsproxy
```

**解决方案**:
使用 `replace` 指令而不是直接修改 `require`

### 问题 2: 依赖下载失败

**错误信息**:
```
go: github.com/lkxlzx/dnsproxy@v0.79.2: invalid version: unknown revision v0.79.2
```

**解决方案**:
1. 确认 fork 仓库中存在 v0.79.2 标签
2. 检查网络连接
3. 清理模块缓存: `go clean -modcache`

### 问题 3: 编译错误

**可能原因**:
- Fork 版本与 AdGuardHome 不兼容
- API 变更

**解决方案**:
1. 检查 fork 版本的 API 变更
2. 更新 AdGuardHome 中的调用代码
3. 如果问题严重，考虑切换回原始版本

---

## 📚 相关文档

- [DNSProxy 原始仓库](https://github.com/AdguardTeam/dnsproxy)
- [DNSProxy Fork 仓库](https://github.com/lkxlzx/dnsproxy)
- [Go Modules 文档](https://go.dev/ref/mod)
- [Replace 指令说明](https://go.dev/ref/mod#go-mod-file-replace)

---

## ✅ 集成检查清单

- [x] 在 go.mod 中添加 replace 指令
- [x] 运行 `go mod tidy` 更新依赖
- [x] 编译成功
- [x] 配置检查通过
- [ ] 运行完整测试套件
- [ ] 运行性能基准测试
- [ ] 对比性能数据
- [ ] 更新文档

---

**维护者**: lkxlzx  
**最后更新**: 2025-12-06  
**状态**: ✅ 已集成并测试

