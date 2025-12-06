# 🚀 快速开发指南 - AdGuardHome + DNSProxy 联合开发

**当前状态**: ✅ 已配置完成，可以立即开始开发！

---

## 📁 当前设置

```
E:\Kiro\AdGuardHome\
├── vendor-dev/
│   └── dnsproxy/          # 你的 dnsproxy fork (v0.79.2)
├── internal/
├── go.mod                 # 已配置使用 ./vendor-dev/dnsproxy
└── switch-dnsproxy.ps1    # 快速切换脚本
```

---

## ⚡ 快速开始

### 1. 修改 dnsproxy 代码

```powershell
# 在 VS Code 或你喜欢的编辑器中打开
code vendor-dev/dnsproxy

# 或直接编辑文件
notepad vendor-dev/dnsproxy/proxy/cache.go
```

### 2. 测试修改

```powershell
# 重新编译 AdGuardHome（会自动使用你修改的 dnsproxy）
go build -o AdGuardHome_dev.exe

# 运行测试
go test ./internal/dnsrouting -v

# 启动服务
.\AdGuardHome_dev.exe --no-check-update
```

### 3. 提交修改

```powershell
# 提交 dnsproxy 更改
cd vendor-dev/dnsproxy
git add .
git commit -m "feat: your changes"
git push origin v0.79.2

# 返回主目录
cd ..\..
```

---

## 🔄 模式切换

### 本地开发模式（当前）

```powershell
.\switch-dnsproxy.ps1 -Mode local
```

使用 `vendor-dev/dnsproxy` 中的代码

### 远程版本模式

```powershell
.\switch-dnsproxy.ps1 -Mode remote -Version v0.79.2
```

使用 GitHub 上的版本

---

## 🧪 测试工作流

### 快速测试

```powershell
# 测试 dnsproxy
cd vendor-dev/dnsproxy
go test ./proxy -v
cd ..\..

# 测试 AdGuardHome 集成
go test ./internal/dnsrouting -v
```

### 完整测试

```powershell
# 所有测试
go test ./...

# 基准测试
go test -bench=. -benchmem ./internal/dnsrouting
```

### 性能对比

```powershell
# 1. 测试本地版本
.\switch-dnsproxy.ps1 -Mode local
go test -bench=BenchmarkCache -benchmem ./internal/dnsrouting > bench-local.txt

# 2. 测试远程版本
.\switch-dnsproxy.ps1 -Mode remote
go test -bench=BenchmarkCache -benchmem ./internal/dnsrouting > bench-remote.txt

# 3. 对比
# 手动对比两个文件，或使用 benchstat
```

---

## 📝 常见开发场景

### 场景 1: 修改缓存逻辑

```powershell
# 1. 编辑缓存代码
code vendor-dev/dnsproxy/proxy/cache.go

# 2. 添加测试
code vendor-dev/dnsproxy/proxy/cache_test.go

# 3. 测试 dnsproxy
cd vendor-dev/dnsproxy
go test ./proxy -v -run TestCache
cd ..\..

# 4. 测试 AdGuardHome 集成
go build -o test.exe
.\test.exe --check-config
```

### 场景 2: 添加新功能

```powershell
# 1. 在 dnsproxy 中实现功能
code vendor-dev/dnsproxy/proxy/new_feature.go

# 2. 在 AdGuardHome 中使用
code internal/dnsforward/dnsforward.go

# 3. 编译测试
go build -o test.exe
go test ./...
```

### 场景 3: 调试问题

```powershell
# 1. 添加调试日志
# 在 vendor-dev/dnsproxy 中添加:
log.Printf("DEBUG: %v", someVariable)

# 2. 重新编译
go build -o debug.exe

# 3. 运行并查看日志
.\debug.exe --no-check-update
```

---

## 🐛 调试技巧

### 使用 VS Code 调试

1. 打开 AdGuardHome 项目
2. 设置断点（可以在 dnsproxy 代码中设置！）
3. 按 F5 开始调试

### 使用 Delve

```powershell
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试
dlv debug . -- --no-check-update

# 在 dnsproxy 代码中设置断点
(dlv) break vendor-dev/dnsproxy/proxy/cache.go:123
(dlv) continue
```

### 性能分析

```powershell
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/dnsrouting
go tool pprof cpu.prof

# 内存 profiling
go test -memprofile=mem.prof -bench=. ./internal/dnsrouting
go tool pprof mem.prof
```

---

## 📦 提交和发布

### 开发完成后

1. **测试所有功能**
   ```powershell
   go test ./...
   go build
   ```

2. **提交 dnsproxy**
   ```powershell
   cd vendor-dev/dnsproxy
   git add .
   git commit -m "feat: your feature"
   git push origin v0.79.2
   
   # 创建新版本（可选）
   git tag -a v0.79.3 -m "Version 0.79.3"
   git push origin v0.79.3
   cd ..\..
   ```

3. **更新 AdGuardHome**
   ```powershell
   # 切换到远程版本
   .\switch-dnsproxy.ps1 -Mode remote -Version v0.79.3
   
   # 测试
   go build
   go test ./...
   
   # 提交
   git add go.mod go.sum
   git commit -m "chore: update dnsproxy to v0.79.3"
   git push origin v10.5
   ```

---

## 🎯 最佳实践

### 1. 频繁测试

每次修改后都运行测试：
```powershell
go test ./internal/dnsrouting -v
```

### 2. 保持同步

定期从上游同步：
```powershell
cd vendor-dev/dnsproxy
git fetch upstream
git merge upstream/master
cd ..\..
```

### 3. 小步提交

频繁提交小的改动，而不是一次性大改动。

### 4. 性能监控

修改后运行基准测试，确保性能没有退化。

---

## ✅ 当前状态检查

运行以下命令确认设置正确：

```powershell
# 1. 检查 dnsproxy 位置
ls vendor-dev/dnsproxy

# 2. 检查 go.mod 配置
go list -m github.com/AdguardTeam/dnsproxy
# 应该显示: => ./vendor-dev/dnsproxy

# 3. 编译测试
go build -o test.exe

# 4. 运行测试
go test ./internal/dnsrouting -v
```

全部通过？**你已经准备好开始开发了！** 🎉

---

## 🆘 需要帮助？

- 查看详细文档: `docs/LOCAL_DEVELOPMENT_SETUP.md`
- 查看集成说明: `docs/DNSPROXY_FORK_INTEGRATION.md`
- 查看设置指南: `SETUP_LOCAL_DEV.md`

---

**Happy Coding!** 🚀

