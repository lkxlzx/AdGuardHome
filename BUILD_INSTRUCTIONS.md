# AdGuard Home 编译说明

## 快速开始

### 编译当前平台（Windows）
```powershell
go build -ldflags="-s -w" -o AdGuardHome.exe
```

### 编译主要平台（推荐）
```powershell
.\build-release-simple.ps1
```

这将编译以下 9 个主要平台：
- Windows (amd64, 386, arm64)
- Linux (amd64, 386, arm64, armv7)
- macOS (amd64, arm64)

### 编译所有平台（完整版）
```powershell
.\build-all-platforms-complete.ps1
```

这将编译所有 25 个支持的平台，包括：
- Darwin (macOS): amd64, arm64
- FreeBSD: 386, amd64, armv5, armv6, armv7, arm64
- Linux: 386, amd64, armv5, armv6, armv7, arm64, mips*, ppc64le, riscv64
- OpenBSD: amd64, arm64
- Windows: 386, amd64, arm64

## 构建脚本说明

### build-release-simple.ps1
- **用途**: 编译最常用的 9 个平台
- **速度**: 快速（约 2-3 分钟）
- **推荐**: 日常开发和测试

### build-all-platforms-complete.ps1
- **用途**: 编译所有 25 个支持的平台
- **速度**: 较慢（约 5-8 分钟）
- **推荐**: 正式发布

## 输出目录

所有编译的二进制文件都在 `dist/` 目录下：

```
dist/
├── AdGuardHome_darwin_amd64
├── AdGuardHome_darwin_arm64
├── AdGuardHome_linux_amd64
├── AdGuardHome_linux_arm64
├── AdGuardHome_windows_amd64.exe
└── ...
```

## 版本号

构建脚本会自动从 Git 获取版本号：
- 如果有 Git 标签：使用标签版本
- 如果没有标签：使用 `v0.107.0-dev-<commit-hash>`

## 前端构建

如果需要重新构建前端：

```powershell
cd client
npm ci
npm run build-prod
cd ..
```

## 构建标志说明

- `-s`: 去除符号表
- `-w`: 去除 DWARF 调试信息
- `-X`: 设置版本号变量

这些标志可以显著减小二进制文件大小（约 30-40%）。

## 故障排除

### 问题：编译失败
**解决方案**: 确保已安装 Go 1.25.4 或更高版本
```powershell
go version
```

### 问题：前端未更新
**解决方案**: 重新构建前端
```powershell
cd client
npm run build-prod
```

### 问题：跨平台编译失败
**解决方案**: 某些平台可能需要特定的工具链，Windows 平台编译其他平台通常没有问题

## 文件大小

典型的二进制文件大小：
- Windows: ~29-31 MB
- Linux: ~30-32 MB
- macOS: ~30-33 MB

## Linux 安装

编译完成后，如何在 Linux 服务器上安装？请查看详细指南：

📖 **[Linux 安装指南](LINUX_INSTALL_GUIDE.md)**

包含以下内容：
- 使用编译好的二进制文件安装
- 使用官方安装脚本
- 手动安装和配置 systemd 服务
- 服务管理命令
- 故障排除

## 注意事项

1. **代码修复**: 已修复以下编译问题：
   - `validatedHostname` 函数被注释导致的编译错误
   - `raCtx` 结构体缺失导致的编译错误
   - 配置结构体缺少字段的问题

2. **平台兼容性**: 所有修复都参考了 Windows 平台的实现，确保跨平台兼容性

3. **构建时间**: 完整构建所有平台大约需要 5-8 分钟

4. **Linux 兼容性**: 编译的 Linux 版本完全兼容官方安装方式，可以直接替换官方二进制文件使用

## 官方构建方法

如果需要使用官方的 Makefile 构建系统（需要 POSIX 环境）：

```bash
make build-release
```

注意：这需要在 Linux/macOS 或 WSL 环境中运行。
