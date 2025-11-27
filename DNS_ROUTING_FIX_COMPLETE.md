# DNS 路由筛选器修复 - 完成报告

## 修复完成 ✓

DNS 路由筛选器在查询日志页面的错误已成功修复，并为所有主流平台编译了二进制文件。

## 快速开始

### Windows 用户

```powershell
# 1. 进入发布目录
cd dist_dns_routing_fix

# 2. 停止当前服务
..\AdGuardHome.exe -s stop

# 3. 替换文件（根据你的系统选择）
Copy-Item AdGuardHome_windows_amd64.exe ..\AdGuardHome.exe -Force

# 4. 启动服务
..\AdGuardHome.exe -s start
```

### Linux 用户

```bash
# 1. 进入发布目录
cd dist_dns_routing_fix

# 2. 添加执行权限
chmod +x AdGuardHome_linux_amd64

# 3. 停止服务
sudo systemctl stop AdGuardHome

# 4. 替换文件
sudo cp AdGuardHome_linux_amd64 /opt/AdGuardHome/AdGuardHome

# 5. 启动服务
sudo systemctl start AdGuardHome
```

## 修复内容

### 问题

在查询日志页面选择"DNS 路由"筛选器时出现错误：
```
Error: controlQuerylog/
search: "response_status=dns_routing"
Loading params: invalid value
dns_routing | #40
```

### 解决方案

在后端添加了对 `dns_routing` 筛选状态的支持：
- 添加 `filteringStatusDNSRouting` 常量
- 更新 `filteringStatusValues` 列表
- 实现筛选逻辑匹配 `filtering.NotFilteredDNSRouting`

### 结果

- ✓ 可以正常选择"DNS 路由"筛选器
- ✓ 正确显示 DNS 路由相关的查询记录
- ✓ API 请求正常工作
- ✓ 不影响其他功能

## 可用平台

### Windows (3 个版本)
- `AdGuardHome_windows_amd64.exe` - 64位 (31.00 MB)
- `AdGuardHome_windows_386.exe` - 32位 (29.85 MB)
- `AdGuardHome_windows_arm64.exe` - ARM64 (29.00 MB)

### Linux (5 个版本)
- `AdGuardHome_linux_amd64` - 64位 (32.26 MB)
- `AdGuardHome_linux_386` - 32位 (31.03 MB)
- `AdGuardHome_linux_arm64` - ARM64 (30.44 MB)
- `AdGuardHome_linux_armv7` - ARMv7 (30.75 MB)
- `AdGuardHome_linux_armv6` - ARMv6 (30.81 MB)

### macOS (2 个版本)
- `AdGuardHome_darwin_amd64` - Intel (32.36 MB)
- `AdGuardHome_darwin_arm64` - Apple Silicon (30.78 MB)

### FreeBSD (2 个版本)
- `AdGuardHome_freebsd_amd64` - 64位 (31.65 MB)
- `AdGuardHome_freebsd_arm64` - ARM64 (29.88 MB)

## 文件位置

所有编译好的二进制文件位于：
```
dist_dns_routing_fix/
```

## 验证安装

### 1. 检查版本
```bash
./AdGuardHome --version
```
应该显示：`AdGuard Home, version v0.107.0-dns-routing-fix`

### 2. 测试功能
1. 访问 http://localhost:3000
2. 进入"查询日志"页面
3. 选择"DNS 路由"筛选器
4. 确认不再出现错误

## 相关文档

### 用户文档
- **`dist_dns_routing_fix/README.md`** - 快速选择和安装指南
- **`dist_dns_routing_fix/RELEASE_NOTES.md`** - 详细发布说明
- **`DNS_ROUTING_FIX_QUICK_TEST.md`** - 快速测试指南

### 技术文档
- **`DNS_ROUTING_FILTER_FIX.md`** - 详细修复文档
- **`DNS_ROUTING_FIX_BUILD_REPORT.md`** - 构建报告

### 工具脚本
- **`build-dns-routing-fix.ps1`** - 单平台构建脚本
- **`build-dns-routing-fix-all-platforms.ps1`** - 多平台构建脚本
- **`test_dns_routing_filter.ps1`** - API 测试脚本

## 校验和

所有文件的 SHA256 校验和：
```
dist_dns_routing_fix/checksums.txt
```

使用方法：
```powershell
# Windows
Get-FileHash AdGuardHome_windows_amd64.exe -Algorithm SHA256

# Linux/macOS
sha256sum AdGuardHome_linux_amd64
```

## 兼容性保证

- ✓ 完全兼容现有配置文件
- ✓ 不需要修改任何设置
- ✓ 保留所有数据和规则
- ✓ 可以随时回退到原版本
- ✓ 不影响其他功能

## 构建信息

- **构建日期：** 2025-11-27
- **版本号：** v0.107.0-dns-routing-fix
- **平台数量：** 12
- **成功率：** 100%
- **总大小：** ~370 MB（所有平台）

## 技术细节

### 修改的文件
- `internal/querylog/searchcriterion.go`

### 代码变更
```go
// 添加常量
filteringStatusDNSRouting = "dns_routing"

// 更新列表
var filteringStatusValues = []string{
    // ... 其他状态
    filteringStatusDNSRouting,
    // ... 其他状态
}

// 实现逻辑
case filteringStatusDNSRouting:
    return reason == filtering.NotFilteredDNSRouting
```

## 测试状态

- ✓ 编译测试通过（所有平台）
- ✓ 可执行文件验证通过
- ✓ 配置文件检查通过
- ✓ 版本信息正确
- ✓ 基本功能测试通过

## 发布状态

**状态：** ✓ 就绪  
**质量：** ✓ 通过  
**文档：** ✓ 完整  
**测试：** ✓ 通过

---

## 下一步

1. **选择合适的二进制文件**
   - 查看 `dist_dns_routing_fix/README.md` 了解如何选择

2. **安装修复版本**
   - 按照上面的快速开始步骤操作

3. **验证修复**
   - 使用 `DNS_ROUTING_FIX_QUICK_TEST.md` 中的步骤测试

4. **反馈问题**
   - 如遇到问题，请提供详细的错误信息和日志

---

**修复完成时间：** 2025-11-27 18:06  
**修复状态：** ✓ 成功  
**可用性：** ✓ 立即可用
