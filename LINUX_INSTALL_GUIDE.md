# Linux 安装指南

## 方法一：使用编译好的二进制文件（推荐）

### 1. 上传二进制文件到 Linux 服务器

从 Windows 编译完成后，将对应的二进制文件上传到 Linux 服务器：

```bash
# 使用 scp 上传（在 Windows PowerShell 中执行）
scp dist/AdGuardHome_linux_amd64 user@your-server:/tmp/

# 或使用 SFTP、FTP 等工具上传
```

### 2. 在 Linux 服务器上安装

```bash
# 创建安装目录
sudo mkdir -p /opt/AdGuardHome

# 移动并重命名二进制文件
sudo mv /tmp/AdGuardHome_linux_amd64 /opt/AdGuardHome/AdGuardHome

# 添加执行权限
sudo chmod +x /opt/AdGuardHome/AdGuardHome

# 安装为系统服务
cd /opt/AdGuardHome
sudo ./AdGuardHome -s install

# 启动服务
sudo ./AdGuardHome -s start
```

### 3. 验证安装

```bash
# 检查服务状态
sudo ./AdGuardHome -s status

# 或使用 systemctl（如果安装为 systemd 服务）
sudo systemctl status AdGuardHome
```

### 4. 访问 Web 界面

打开浏览器访问：`http://your-server-ip:3000`

## 方法二：使用官方安装脚本

官方脚本会自动下载最新的官方版本（不是你编译的版本）：

```bash
# 下载并运行官方安装脚本
curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

### 指定版本和架构

```bash
# 安装到自定义目录
curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v -o /opt

# 指定 CPU 架构
curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v -C arm64

# 指定发布渠道
curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v -c beta
```

## 方法三：手动安装（完全控制）

### 1. 准备文件

```bash
# 创建目录结构
sudo mkdir -p /opt/AdGuardHome
cd /opt/AdGuardHome

# 上传你编译的二进制文件
# 假设已经上传到 /tmp/AdGuardHome_linux_amd64

sudo mv /tmp/AdGuardHome_linux_amd64 ./AdGuardHome
sudo chmod +x ./AdGuardHome
```

### 2. 创建 systemd 服务文件

```bash
sudo nano /etc/systemd/system/AdGuardHome.service
```

添加以下内容：

```ini
[Unit]
Description=AdGuard Home: Network-level blocker
After=network.target
After=syslog.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/AdGuardHome
ExecStart=/opt/AdGuardHome/AdGuardHome -c /opt/AdGuardHome/AdGuardHome.yaml -w /opt/AdGuardHome --no-check-update
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 3. 启用并启动服务

```bash
# 重新加载 systemd
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome

# 查看状态
sudo systemctl status AdGuardHome
```

## 支持的 Linux 架构

我们已编译以下 Linux 版本：

| 架构 | 文件名 | 适用设备 |
|------|--------|----------|
| amd64 | `AdGuardHome_linux_amd64` | 64位 Intel/AMD 服务器 |
| 386 | `AdGuardHome_linux_386` | 32位 Intel/AMD 服务器 |
| arm64 | `AdGuardHome_linux_arm64` | 64位 ARM（树莓派4、云服务器） |
| armv7 | `AdGuardHome_linux_armv7` | ARMv7（树莓派3、旧版树莓派） |

## 服务管理命令

### 使用 AdGuardHome 内置命令

```bash
# 启动
sudo /opt/AdGuardHome/AdGuardHome -s start

# 停止
sudo /opt/AdGuardHome/AdGuardHome -s stop

# 重启
sudo /opt/AdGuardHome/AdGuardHome -s restart

# 状态
sudo /opt/AdGuardHome/AdGuardHome -s status

# 卸载服务
sudo /opt/AdGuardHome/AdGuardHome -s uninstall
```

### 使用 systemctl（如果安装为 systemd 服务）

```bash
# 启动
sudo systemctl start AdGuardHome

# 停止
sudo systemctl stop AdGuardHome

# 重启
sudo systemctl restart AdGuardHome

# 状态
sudo systemctl status AdGuardHome

# 查看日志
sudo journalctl -u AdGuardHome -f
```

## 配置文件位置

- 主配置文件：`/opt/AdGuardHome/AdGuardHome.yaml`
- 数据目录：`/opt/AdGuardHome/data/`
- 工作目录：`/opt/AdGuardHome/work/`

## 防火墙配置

如果启用了防火墙，需要开放以下端口：

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 3000/tcp  # Web 界面
sudo ufw allow 53/tcp    # DNS
sudo ufw allow 53/udp    # DNS
sudo ufw allow 80/tcp    # HTTP (可选)
sudo ufw allow 443/tcp   # HTTPS (可选)

# firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --permanent --add-port=53/tcp
sudo firewall-cmd --permanent --add-port=53/udp
sudo firewall-cmd --reload
```

## 更新

### 更新到新编译的版本

```bash
# 停止服务
sudo systemctl stop AdGuardHome

# 备份当前版本
sudo cp /opt/AdGuardHome/AdGuardHome /opt/AdGuardHome/AdGuardHome.backup

# 上传新版本并替换
sudo mv /tmp/AdGuardHome_linux_amd64 /opt/AdGuardHome/AdGuardHome
sudo chmod +x /opt/AdGuardHome/AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome
```

## 卸载

```bash
# 停止并卸载服务
sudo /opt/AdGuardHome/AdGuardHome -s uninstall

# 或使用 systemctl
sudo systemctl stop AdGuardHome
sudo systemctl disable AdGuardHome
sudo rm /etc/systemd/system/AdGuardHome.service
sudo systemctl daemon-reload

# 删除文件
sudo rm -rf /opt/AdGuardHome
```

## 故障排除

### 检查二进制文件是否可执行

```bash
file /opt/AdGuardHome/AdGuardHome
# 应该显示：ELF 64-bit LSB executable, x86-64...

# 测试运行
/opt/AdGuardHome/AdGuardHome --version
```

### 检查端口占用

```bash
# 检查 53 端口是否被占用
sudo netstat -tulpn | grep :53
sudo lsof -i :53

# 如果被 systemd-resolved 占用，需要禁用它
sudo systemctl disable systemd-resolved
sudo systemctl stop systemd-resolved
```

### 查看日志

```bash
# 查看 systemd 日志
sudo journalctl -u AdGuardHome -n 100 --no-pager

# 实时查看日志
sudo journalctl -u AdGuardHome -f

# 查看 AdGuardHome 自己的日志
sudo tail -f /opt/AdGuardHome/data/querylog.json
```

## 验证编译版本

我们编译的版本包含以下特性：
- ✅ DNS 缓存优化
- ✅ 预取功能
- ✅ 改进的仪表板
- ✅ 所有最新的修复和优化

验证版本信息：

```bash
/opt/AdGuardHome/AdGuardHome --version
# 应该显示：v0.107.0-dev-<commit-hash>
```

## 注意事项

1. **权限要求**：AdGuard Home 需要 root 权限才能绑定到 53 端口（DNS）
2. **SELinux**：如果启用了 SELinux，可能需要配置策略
3. **配置保留**：更新时会保留 `AdGuardHome.yaml` 配置文件
4. **数据备份**：建议定期备份 `/opt/AdGuardHome/data/` 目录

## 与官方版本的区别

| 特性 | 官方版本 | 我们的编译版本 |
|------|----------|----------------|
| 安装方式 | 官方脚本自动下载 | 手动上传二进制文件 |
| 版本号 | 官方发布版本 | v0.107.0-dev |
| 功能 | 稳定版功能 | 包含最新开发功能 |
| 更新 | 自动检查更新 | 手动更新 |

我们编译的版本包含了所有最新的代码修复和功能优化，适合测试和使用最新特性。
