# AdGuard Home - DNS Routing Filter Fix

## 快速选择指南 (Quick Selection Guide)

### Windows 用户

| 系统类型 | 文件名 |
|---------|--------|
| Windows 10/11 64位 (最常见) | `AdGuardHome_windows_amd64.exe` |
| Windows 32位 | `AdGuardHome_windows_386.exe` |
| Windows ARM64 (Surface Pro X等) | `AdGuardHome_windows_arm64.exe` |

**如何确定系统类型：**
- 按 `Win + Pause` 键，查看"系统类型"
- 或在 PowerShell 中运行：`[Environment]::Is64BitOperatingSystem`

### Linux 用户

| 系统类型 | 文件名 |
|---------|--------|
| Ubuntu/Debian/CentOS 64位 | `AdGuardHome_linux_amd64` |
| Ubuntu/Debian 32位 | `AdGuardHome_linux_386` |
| Raspberry Pi 4/3/2 | `AdGuardHome_linux_armv7` |
| Raspberry Pi 1/Zero | `AdGuardHome_linux_armv6` |
| ARM64 服务器 | `AdGuardHome_linux_arm64` |

**如何确定系统类型：**
```bash
uname -m
# x86_64 -> 使用 amd64
# i686 -> 使用 386
# armv7l -> 使用 armv7
# armv6l -> 使用 armv6
# aarch64 -> 使用 arm64
```

### macOS 用户

| 系统类型 | 文件名 |
|---------|--------|
| Intel Mac | `AdGuardHome_darwin_amd64` |
| Apple Silicon (M1/M2/M3) | `AdGuardHome_darwin_arm64` |

**如何确定系统类型：**
- 点击左上角苹果图标 → "关于本机"
- 查看"处理器"或"芯片"信息
- 或在终端运行：`uname -m`

### FreeBSD 用户

| 系统类型 | 文件名 |
|---------|--------|
| FreeBSD 64位 | `AdGuardHome_freebsd_amd64` |
| FreeBSD ARM64 | `AdGuardHome_freebsd_arm64` |

## 文件列表 (File List)

```
dist_dns_routing_fix/
├── AdGuardHome_windows_amd64.exe    (31.00 MB)
├── AdGuardHome_windows_386.exe      (29.85 MB)
├── AdGuardHome_windows_arm64.exe    (29.00 MB)
├── AdGuardHome_linux_amd64          (32.26 MB)
├── AdGuardHome_linux_386            (31.03 MB)
├── AdGuardHome_linux_arm64          (30.44 MB)
├── AdGuardHome_linux_armv7          (30.75 MB)
├── AdGuardHome_linux_armv6          (30.81 MB)
├── AdGuardHome_darwin_amd64         (32.36 MB)
├── AdGuardHome_darwin_arm64         (30.78 MB)
├── AdGuardHome_freebsd_amd64        (31.65 MB)
├── AdGuardHome_freebsd_arm64        (29.88 MB)
├── RELEASE_NOTES.md                 (发布说明)
└── README.md                        (本文件)
```

## 快速安装 (Quick Installation)

### Windows

```powershell
# 1. 下载文件
# 2. 停止当前服务
.\AdGuardHome.exe -s stop

# 3. 备份原文件
Copy-Item AdGuardHome.exe AdGuardHome.exe.backup

# 4. 替换文件
Copy-Item AdGuardHome_windows_amd64.exe AdGuardHome.exe -Force

# 5. 启动服务
.\AdGuardHome.exe -s start
```

### Linux

```bash
# 1. 下载文件
# 2. 添加执行权限
chmod +x AdGuardHome_linux_amd64

# 3. 停止服务
sudo systemctl stop AdGuardHome

# 4. 备份原文件
sudo cp /opt/AdGuardHome/AdGuardHome /opt/AdGuardHome/AdGuardHome.backup

# 5. 替换文件
sudo cp AdGuardHome_linux_amd64 /opt/AdGuardHome/AdGuardHome

# 6. 启动服务
sudo systemctl start AdGuardHome
```

### macOS

```bash
# 1. 下载文件
# 2. 添加执行权限
chmod +x AdGuardHome_darwin_arm64

# 3. 停止服务
sudo launchctl unload /Library/LaunchDaemons/AdGuardHome.plist

# 4. 备份原文件
sudo cp /Applications/AdGuardHome/AdGuardHome /Applications/AdGuardHome/AdGuardHome.backup

# 5. 替换文件
sudo cp AdGuardHome_darwin_arm64 /Applications/AdGuardHome/AdGuardHome

# 6. 启动服务
sudo launchctl load /Library/LaunchDaemons/AdGuardHome.plist
```

## 验证安装 (Verify Installation)

### 1. 检查版本

```bash
./AdGuardHome --version
```

应该显示：`AdGuard Home, version v0.107.0-dns-routing-fix`

### 2. 检查配置

```bash
./AdGuardHome --check-config
```

应该显示：`configuration file is ok`

### 3. 测试功能

1. 访问 Web 管理界面（默认：http://localhost:3000）
2. 进入"查询日志"页面
3. 在筛选器下拉菜单中选择"DNS 路由"
4. 确认不再出现错误

## 修复说明 (What's Fixed)

本版本修复了查询日志页面中"DNS 路由"筛选器的错误。

**修复前：**
选择"DNS 路由"筛选器时出现错误：
```
Error: controlQuerylog/
search: "response_status=dns_routing"
Loading params: invalid value
```

**修复后：**
- ✓ 可以正常选择"DNS 路由"筛选器
- ✓ 正确显示 DNS 路由相关的查询记录
- ✓ API 请求正常工作

## 兼容性 (Compatibility)

- ✓ 完全兼容现有配置文件
- ✓ 不需要修改任何设置
- ✓ 保留所有数据和规则
- ✓ 可以随时回退到原版本

## 需要帮助？ (Need Help?)

1. 查看 `RELEASE_NOTES.md` 了解详细信息
2. 查看项目根目录的 `DNS_ROUTING_FILTER_FIX.md` 了解技术细节
3. 查看 `DNS_ROUTING_FIX_QUICK_TEST.md` 了解测试方法

## 校验和 (Checksums)

如需验证文件完整性，可以使用以下命令：

### Windows (PowerShell)
```powershell
Get-FileHash AdGuardHome_windows_amd64.exe -Algorithm SHA256
```

### Linux / macOS
```bash
sha256sum AdGuardHome_linux_amd64
# 或
shasum -a 256 AdGuardHome_darwin_arm64
```

## 许可证 (License)

本修复版本遵循 AdGuard Home 的原始许可证（GPL-3.0）。
