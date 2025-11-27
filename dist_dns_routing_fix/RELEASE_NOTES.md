# AdGuard Home - DNS Routing Filter Fix Release

## Version: v0.107.0-dns-routing-fix

## 修复内容 (What's Fixed)

### DNS 路由筛选器错误修复

**问题描述：**
在查询日志页面中，当用户选择"DNS 路由"筛选器时，会出现以下错误：
```
Error: controlQuerylog/
search: "response_status=dns_routing"
Loading params: invalid value
dns_routing | #40
```

**修复方案：**
- 在后端 `internal/querylog/searchcriterion.go` 中添加了对 `dns_routing` 筛选状态的支持
- 添加了 `filteringStatusDNSRouting` 常量定义
- 实现了对 `filtering.NotFilteredDNSRouting` 原因的筛选逻辑
- 将 `dns_routing` 添加到支持的筛选状态列表中

**影响范围：**
- 修复了查询日志页面中"DNS 路由"筛选器的功能
- 用户现在可以正常筛选和查看通过 DNS 路由规则处理的查询记录
- 不影响其他筛选器和功能

## 支持的平台 (Supported Platforms)

本次发布包含以下平台的二进制文件：

### Windows
- `AdGuardHome_windows_amd64.exe` - Windows 64-bit (Intel/AMD)
- `AdGuardHome_windows_386.exe` - Windows 32-bit
- `AdGuardHome_windows_arm64.exe` - Windows ARM64

### Linux
- `AdGuardHome_linux_amd64` - Linux 64-bit (Intel/AMD)
- `AdGuardHome_linux_386` - Linux 32-bit
- `AdGuardHome_linux_arm64` - Linux ARM64
- `AdGuardHome_linux_armv7` - Linux ARMv7 (Raspberry Pi 2+)
- `AdGuardHome_linux_armv6` - Linux ARMv6 (Raspberry Pi 1)

### macOS
- `AdGuardHome_darwin_amd64` - macOS Intel
- `AdGuardHome_darwin_arm64` - macOS Apple Silicon (M1/M2/M3)

### FreeBSD
- `AdGuardHome_freebsd_amd64` - FreeBSD 64-bit
- `AdGuardHome_freebsd_arm64` - FreeBSD ARM64

## 安装说明 (Installation)

### Windows

1. 下载对应的 `.exe` 文件
2. 停止当前运行的 AdGuard Home
3. 替换原有的可执行文件
4. 重新启动 AdGuard Home

```powershell
# 停止服务（如果作为服务运行）
.\AdGuardHome.exe -s stop

# 替换文件
Copy-Item AdGuardHome_windows_amd64.exe AdGuardHome.exe -Force

# 启动服务
.\AdGuardHome.exe -s start
```

### Linux / macOS / FreeBSD

1. 下载对应平台的二进制文件
2. 添加执行权限
3. 停止当前运行的 AdGuard Home
4. 替换原有的可执行文件
5. 重新启动 AdGuard Home

```bash
# 添加执行权限
chmod +x AdGuardHome_linux_amd64

# 停止服务
sudo systemctl stop AdGuardHome

# 替换文件
sudo cp AdGuardHome_linux_amd64 /opt/AdGuardHome/AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome
```

## 验证修复 (Verification)

### 1. 检查版本

```bash
./AdGuardHome --version
```

应该显示：`AdGuard Home, version v0.107.0-dns-routing-fix`

### 2. 测试 Web 界面

1. 访问 AdGuard Home 管理界面
2. 进入"查询日志"页面
3. 在筛选器下拉菜单中选择"DNS 路由"
4. 确认不再出现错误，可以正常筛选 DNS 路由记录

### 3. API 测试（可选）

使用 curl 测试 API：

```bash
curl -u admin:password -X POST \
  http://localhost:3000/control/querylog \
  -H "Content-Type: application/json" \
  -d '{"response_status":"dns_routing","older_than":""}'
```

应该返回成功响应（HTTP 200）。

## 技术细节 (Technical Details)

### 修改的文件

- `internal/querylog/searchcriterion.go`

### 代码变更

1. 添加常量定义：
```go
filteringStatusDNSRouting = "dns_routing" // DNS routing
```

2. 更新支持的筛选状态列表：
```go
var filteringStatusValues = []string{
    // ... 其他状态
    filteringStatusDNSRouting,
    // ... 其他状态
}
```

3. 实现筛选逻辑：
```go
case filteringStatusDNSRouting:
    return reason == filtering.NotFilteredDNSRouting
```

## 兼容性 (Compatibility)

- 完全兼容现有的配置文件
- 不需要修改任何配置
- 可以直接替换原有版本
- 保留所有现有功能和数据

## 已知问题 (Known Issues)

无

## 反馈 (Feedback)

如果遇到任何问题，请提供以下信息：
- 使用的平台和架构
- AdGuard Home 版本信息
- 错误日志
- 重现步骤

## 构建信息 (Build Information)

- 构建日期：2025-11-27
- Go 版本：使用系统默认 Go 版本
- CGO：禁用（CGO_ENABLED=0）
- 编译标志：`-ldflags="-s -w"`

## 相关文档 (Related Documentation)

- `DNS_ROUTING_FILTER_FIX.md` - 详细修复文档
- `DNS_ROUTING_FIX_QUICK_TEST.md` - 快速测试指南
- `test_dns_routing_filter.ps1` - Windows 测试脚本
