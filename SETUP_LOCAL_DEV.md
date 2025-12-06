# 🚀 本地联合开发环境快速设置指南

按照以下步骤设置 AdGuardHome + dnsproxy 联合开发环境。

---

## 📋 前置要求

- Git
- Go 1.25.4+
- 文本编辑器或 IDE

---

## 🔧 设置步骤

### 步骤 1: 克隆 dnsproxy (在 PowerShell 中执行)

```powershell
# 进入父目录
cd E:\Kiro

# 克隆你的 dnsproxy fork
git clone https://github.com/lkxlzx/dnsproxy.git

# 进入 dnsproxy 目录
cd dnsproxy

# 切换到 v0.79.2 分支
git checkout v0.79.2

# 验证分支
git branch

# 返回 AdGuardHome 目录
cd ..\AdGuardHome
```

### 步骤 2: 切换到本地开发模式

```powershell
# 在 AdGuardHome 目录中执行
.\switch-dnsproxy.ps1 -Mode local
```

你应该看到：

```
=== DNSProxy Mode Switcher ===

Switching to LOCAL development mode...
  Using: ../dnsproxy

Updating dependencies...

SUCCESS! Switched to local mode

Current dnsproxy configuration:
github.com/AdguardTeam/dnsproxy v0.77.0 => ../dnsproxy
```

### 步骤 3: 验证设置

```powershell
# 检查依赖
go list -m github.com/AdguardTeam/dnsproxy

# 应该显示:
# github.com/AdguardTeam/dnsproxy v0.77.0 => ../dnsproxy

# 编译测试
go build -o AdGuardHome_dev.exe

# 运行测试
go test ./internal/dnsrouting -v
```

---

## 🎯 现在你可以开始联合开发了！

### 典型工作流程

1. **修改 dnsproxy 代码**
   ```powershell
   # 在另一个终端或编辑器中打开
   cd E:\Kiro\dnsproxy
   # 编辑文件...
   ```

2. **在 AdGuardHome 中测试修改**
   ```powershell
   cd E:\Kiro\AdGuardHome
   
   # 重新编译（会自动使用本地 dnsproxy 的修改）
   go build -o AdGuardHome_dev.exe
   
   # 测试
   .\AdGuardHome_dev.exe --check-config
   ```

3. **运行测试**
   ```powershell
   # 测试 dnsproxy
   cd E:\Kiro\dnsproxy
   go test ./...
   
   # 测试 AdGuardHome（包括 dnsproxy 集成）
   cd E:\Kiro\AdGuardHome
   go test ./...
   ```

---

## 🔄 切换模式

### 切换到本地开发模式

```powershell
.\switch-dnsproxy.ps1 -Mode local
```

### 切换回远程版本（提交前）

```powershell
.\switch-dnsproxy.ps1 -Mode remote -Version v0.79.2
```

---

## 📝 提交工作流

### 完成开发后

1. **提交 dnsproxy 更改**
   ```powershell
   cd E:\Kiro\dnsproxy
   
   git add .
   git commit -m "feat: your changes"
   git push origin v0.79.2
   
   # 如果需要创建新版本
   git tag -a v0.79.3 -m "Version 0.79.3"
   git push origin v0.79.3
   ```

2. **更新 AdGuardHome**
   ```powershell
   cd E:\Kiro\AdGuardHome
   
   # 切换回远程版本
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

## 🐛 调试技巧

### 1. 在 dnsproxy 中添加日志

```go
// dnsproxy/proxy/cache.go
log.Printf("DEBUG: Cache operation for %s", key)
```

重新编译 AdGuardHome 后就能看到这些日志。

### 2. 使用 VS Code 调试

在 `.vscode/launch.json` 中：

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug AdGuardHome",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["--no-check-update"]
    }
  ]
}
```

现在可以在 AdGuardHome 和 dnsproxy 的代码中设置断点！

### 3. 性能分析

```powershell
# 运行基准测试
go test -bench=. -benchmem -cpuprofile=cpu.prof ./internal/dnsrouting

# 分析结果（包括 dnsproxy 的调用）
go tool pprof cpu.prof
```

---

## ✅ 验证清单

设置完成后，确认：

- [ ] `E:\Kiro\dnsproxy` 目录存在
- [ ] dnsproxy 在 v0.79.2 分支
- [ ] `go list -m github.com/AdguardTeam/dnsproxy` 显示 `=> ../dnsproxy`
- [ ] AdGuardHome 编译成功
- [ ] 测试通过

---

## 🆘 遇到问题？

### 问题 1: 找不到 dnsproxy

```
ERROR: ../dnsproxy directory not found!
```

**解决**: 确保 dnsproxy 克隆在正确位置：
```powershell
cd E:\Kiro
ls  # 应该看到 AdGuardHome 和 dnsproxy 两个目录
```

### 问题 2: go mod tidy 失败

```
go: errors parsing go.mod
```

**解决**: 检查 go.mod 中的 replace 指令格式是否正确。

### 问题 3: 修改 dnsproxy 后没有效果

**解决**: 确保重新编译了 AdGuardHome：
```powershell
go clean -cache
go build -o AdGuardHome_dev.exe
```

---

## 📚 更多信息

详细文档请查看：
- `docs/LOCAL_DEVELOPMENT_SETUP.md` - 完整开发指南
- `docs/DNSPROXY_FORK_INTEGRATION.md` - Fork 集成说明

---

**准备好了吗？开始执行步骤 1！** 🚀

