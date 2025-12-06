# 本地联合开发环境设置

**更新日期**: 2025-12-06  
**适用场景**: 同时开发 AdGuardHome 和 dnsproxy

---

## 📁 目录结构

推荐的本地开发目录结构：

```
E:\Kiro\
├── AdGuardHome\          # 本项目
└── dnsproxy\             # dnsproxy fork 项目
```

---

## 🚀 快速设置

### 1. 克隆 dnsproxy fork

```bash
# 进入父目录
cd E:\Kiro

# 克隆你的 dnsproxy fork
git clone https://github.com/lkxlzx/dnsproxy.git

# 切换到 v0.79.2 分支
cd dnsproxy
git checkout v0.79.2
```

### 2. 验证目录结构

```bash
# 应该看到两个目录
E:\Kiro\
├── AdGuardHome\
└── dnsproxy\
```

### 3. 配置 go.mod (已完成)

AdGuardHome 的 `go.mod` 已配置为使用本地路径：

```go
replace github.com/AdguardTeam/dnsproxy => ../dnsproxy
```

### 4. 更新依赖

```bash
cd E:\Kiro\AdGuardHome
go mod tidy
```

### 5. 验证设置

```bash
# 检查使用的是本地版本
go list -m github.com/AdguardTeam/dnsproxy

# 应该显示:
# github.com/AdguardTeam/dnsproxy v0.77.0 => ../dnsproxy
```

---

## 🔄 开发工作流

### 典型开发流程

1. **修改 dnsproxy 代码**
   ```bash
   cd E:\Kiro\dnsproxy
   # 编辑代码...
   ```

2. **在 AdGuardHome 中测试**
   ```bash
   cd E:\Kiro\AdGuardHome
   
   # 重新编译（会自动使用本地 dnsproxy）
   go build -o AdGuardHome_dev.exe
   
   # 运行测试
   go test ./...
   ```

3. **实时调试**
   ```bash
   # 启动 AdGuardHome
   .\AdGuardHome_dev.exe
   
   # 修改 dnsproxy 代码后，只需重新编译 AdGuardHome
   go build -o AdGuardHome_dev.exe
   ```

---

## 🐛 调试技巧

### 1. 添加调试日志

在 dnsproxy 中添加日志：

```go
// dnsproxy/proxy/cache.go
log.Printf("DEBUG: Cache hit for %s", domain)
```

在 AdGuardHome 中会立即看到这些日志。

### 2. 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试 AdGuardHome（包括 dnsproxy 代码）
dlv debug . -- --no-check-update

# 在 dnsproxy 代码中设置断点
(dlv) break github.com/AdguardTeam/dnsproxy/proxy.(*Proxy).Resolve
(dlv) continue
```

### 3. 性能分析

```bash
# 生成 CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/dnsrouting

# 分析 profile（包括 dnsproxy 的调用）
go tool pprof cpu.prof
```

---

## 📝 提交工作流

### 开发完成后的提交流程

#### 1. 提交 dnsproxy 更改

```bash
cd E:\Kiro\dnsproxy

# 查看更改
git status
git diff

# 提交更改
git add .
git commit -m "feat: improve cache performance"

# 推送到你的 fork
git push origin v0.79.2

# 创建新版本标签（如果需要）
git tag -a v0.79.3 -m "Version 0.79.3 with cache improvements"
git push origin v0.79.3
```

#### 2. 更新 AdGuardHome 依赖

如果创建了新版本标签，更新 go.mod：

```bash
cd E:\Kiro\AdGuardHome

# 编辑 go.mod，改回远程版本
# replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy v0.79.3

# 更新依赖
go mod tidy

# 测试
go build
go test ./...

# 提交
git add go.mod go.sum
git commit -m "chore: update dnsproxy to v0.79.3"
git push origin v10.5
```

---

## 🔀 切换开发模式

### 切换到本地开发模式

```bash
# 编辑 go.mod
replace github.com/AdguardTeam/dnsproxy => ../dnsproxy

# 更新
go mod tidy
```

### 切换到远程版本

```bash
# 编辑 go.mod
replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy v0.79.2

# 更新
go mod tidy
```

### 使用脚本快速切换

创建 `switch-dnsproxy.ps1`:

```powershell
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("local", "remote")]
    [string]$Mode,
    
    [string]$Version = "v0.79.2"
)

$goModPath = "go.mod"
$content = Get-Content $goModPath -Raw

if ($Mode -eq "local") {
    Write-Host "Switching to local dnsproxy..." -ForegroundColor Yellow
    $content = $content -replace 'replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy .*', 'replace github.com/AdguardTeam/dnsproxy => ../dnsproxy'
} else {
    Write-Host "Switching to remote dnsproxy $Version..." -ForegroundColor Yellow
    $content = $content -replace 'replace github.com/AdguardTeam/dnsproxy => \.\./dnsproxy', "replace github.com/AdguardTeam/dnsproxy => github.com/lkxlzx/dnsproxy $Version"
}

$content | Set-Content $goModPath -NoNewline
go mod tidy

Write-Host "Switched to $Mode mode!" -ForegroundColor Green
go list -m github.com/AdguardTeam/dnsproxy
```

使用方法：

```bash
# 切换到本地模式
.\switch-dnsproxy.ps1 -Mode local

# 切换到远程模式
.\switch-dnsproxy.ps1 -Mode remote -Version v0.79.3
```

---

## 🧪 测试策略

### 1. 单元测试

```bash
# 测试 dnsproxy
cd E:\Kiro\dnsproxy
go test ./...

# 测试 AdGuardHome（包括 dnsproxy 集成）
cd E:\Kiro\AdGuardHome
go test ./...
```

### 2. 集成测试

```bash
# 测试 DNS 路由（使用本地 dnsproxy）
go test ./internal/dnsrouting -v

# 测试 DNS 转发
go test ./internal/dnsforward -v
```

### 3. 性能测试

```bash
# 基准测试
go test -bench=. -benchmem ./internal/dnsrouting

# 对比测试（本地 vs 远程）
# 1. 使用本地版本测试
.\switch-dnsproxy.ps1 -Mode local
go test -bench=. -benchmem ./internal/dnsrouting > bench-local.txt

# 2. 使用远程版本测试
.\switch-dnsproxy.ps1 -Mode remote
go test -bench=. -benchmem ./internal/dnsrouting > bench-remote.txt

# 3. 对比结果
benchstat bench-remote.txt bench-local.txt
```

---

## 📊 性能监控

### 实时性能监控

```bash
# 启动 pprof 服务器
go run . --pprof-port 6060

# 在浏览器中查看
# http://localhost:6060/debug/pprof/

# 查看 CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 查看内存 profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 🔍 常见问题

### Q1: 修改 dnsproxy 后 AdGuardHome 没有变化？

**A**: 确保重新编译 AdGuardHome：

```bash
# 清理缓存
go clean -cache

# 重新编译
go build -o AdGuardHome_dev.exe
```

### Q2: go mod tidy 报错？

**A**: 检查目录结构：

```bash
# 确认 dnsproxy 在正确位置
ls ../dnsproxy

# 如果路径不对，调整 go.mod 中的路径
replace github.com/AdguardTeam/dnsproxy => /path/to/your/dnsproxy
```

### Q3: 如何确认使用的是本地版本？

**A**: 查看依赖信息：

```bash
go list -m -f '{{.Path}} {{.Version}} {{.Replace}}' github.com/AdguardTeam/dnsproxy

# 应该显示本地路径
```

### Q4: 提交时忘记切换回远程版本？

**A**: 不用担心，CI 会失败提醒你：

```bash
# 快速修复
.\switch-dnsproxy.ps1 -Mode remote
git add go.mod go.sum
git commit --amend --no-edit
git push --force-with-lease
```

---

## 📚 推荐工具

### VS Code 配置

`.vscode/settings.json`:

```json
{
  "go.toolsEnvVars": {
    "GOFLAGS": "-tags=debug"
  },
  "go.buildFlags": ["-v"],
  "go.testFlags": ["-v"],
  "go.lintOnSave": "workspace",
  "go.formatTool": "gofumpt"
}
```

### Git 配置

`.git/config`:

```ini
[remote "dnsproxy"]
    url = https://github.com/lkxlzx/dnsproxy.git
    fetch = +refs/heads/*:refs/remotes/dnsproxy/*
```

这样可以在 AdGuardHome 仓库中查看 dnsproxy 的更新：

```bash
git fetch dnsproxy
git log dnsproxy/v0.79.2
```

---

## 🎯 最佳实践

### 1. 保持同步

定期同步上游 dnsproxy 的更新：

```bash
cd E:\Kiro\dnsproxy
git remote add upstream https://github.com/AdguardTeam/dnsproxy.git
git fetch upstream
git merge upstream/master
```

### 2. 分支管理

为每个功能创建独立分支：

```bash
# 在 dnsproxy 中
git checkout -b feature/cache-optimization v0.79.2

# 开发完成后合并回 v0.79.2
git checkout v0.79.2
git merge feature/cache-optimization
```

### 3. 测试覆盖

确保修改有足够的测试覆盖：

```bash
# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## ✅ 检查清单

开发前：
- [ ] dnsproxy 已克隆到正确位置
- [ ] go.mod 配置为本地路径
- [ ] `go mod tidy` 成功
- [ ] 编译成功

开发中：
- [ ] 修改后重新编译 AdGuardHome
- [ ] 运行相关测试
- [ ] 检查性能影响

提交前：
- [ ] dnsproxy 测试通过
- [ ] AdGuardHome 测试通过
- [ ] 性能基准测试对比
- [ ] 切换回远程版本（如果需要）
- [ ] 更新文档

---

**维护者**: lkxlzx  
**最后更新**: 2025-12-06  
**状态**: ✅ 本地开发环境已配置

