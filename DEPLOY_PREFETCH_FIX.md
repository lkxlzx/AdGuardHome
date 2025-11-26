# 部署预取修复版本 - 快速指南

## 📦 已编译文件

✅ **文件**: `dist/AdGuardHome_linux_amd64_prefetch_fixed`  
✅ **大小**: 32.24 MB  
✅ **修复**: 支持远程服务器预取  
✅ **日期**: 2025-11-26

## 🚀 快速部署步骤

### 步骤 1: 上传到服务器

在 Windows PowerShell 中：

```powershell
# 使用 SCP 上传（需要安装 OpenSSH 或使用 WinSCP）
scp dist/AdGuardHome_linux_amd64_prefetch_fixed user@192.168.1.4:/tmp/

# 或者使用 SFTP 工具（如 WinSCP、FileZilla）上传到 /tmp/
```

### 步骤 2: SSH 到服务器

```powershell
ssh user@192.168.1.4
```

### 步骤 3: 替换二进制文件

```bash
# 停止 AdGuard Home
sudo systemctl stop AdGuardHome

# 备份当前版本
sudo cp /opt/AdGuardHome/AdGuardHome /opt/AdGuardHome/AdGuardHome.backup.$(date +%Y%m%d)

# 替换新版本
sudo mv /tmp/AdGuardHome_linux_amd64_prefetch_fixed /opt/AdGuardHome/AdGuardHome
sudo chmod +x /opt/AdGuardHome/AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome

# 检查状态
sudo systemctl status AdGuardHome
```

### 步骤 4: 验证修复

```bash
# 查看预取日志（应该看到 debug 信息）
sudo journalctl -u AdGuardHome -f | grep prefetch

# 检查预取指标
curl http://192.168.1.4:3000/control/cache_metrics | jq
```

## 📊 预期效果

### 修复前
```json
{
  "prefetch_completed": 2435,
  "prefetch_failed": 5,
  "prefetch_success_rate": 99.8
}
```

### 修复后
```json
{
  "prefetch_completed": 2500+,
  "prefetch_failed": 0-2,
  "prefetch_success_rate": 99.9+
}
```

## 🔍 验证清单

- [ ] 服务正常启动
- [ ] Web 界面可访问（http://192.168.1.4:3000）
- [ ] DNS 查询正常工作
- [ ] 预取失败次数减少
- [ ] 日志中看到预取 debug 信息

## 🛠️ 故障排除

### 问题 1: 服务启动失败

```bash
# 查看详细错误
sudo journalctl -u AdGuardHome -n 50 --no-pager

# 检查二进制文件权限
ls -la /opt/AdGuardHome/AdGuardHome

# 恢复备份
sudo systemctl stop AdGuardHome
sudo cp /opt/AdGuardHome/AdGuardHome.backup.* /opt/AdGuardHome/AdGuardHome
sudo systemctl start AdGuardHome
```

### 问题 2: 预取仍然失败

```bash
# 检查 DNS 监听配置
cat /opt/AdGuardHome/AdGuardHome.yaml | grep -A 10 "dns:"

# 确认端口监听
sudo netstat -tulpn | grep AdGuardHome

# 测试本地 DNS
dig @127.0.0.1 google.com
```

### 问题 3: 权限问题

```bash
# 确保正确的所有者和权限
sudo chown root:root /opt/AdGuardHome/AdGuardHome
sudo chmod 755 /opt/AdGuardHome/AdGuardHome
```

## 📝 回滚步骤

如果需要回滚到之前的版本：

```bash
# 停止服务
sudo systemctl stop AdGuardHome

# 恢复备份
sudo cp /opt/AdGuardHome/AdGuardHome.backup.20251126 /opt/AdGuardHome/AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome
```

## 🎯 关键改进

### 代码修改

**文件**: `internal/dnsforward/prefetch.go`

**改进点**:
1. ✅ 自动检测服务器监听地址
2. ✅ 支持 UDP、TCP 多种协议
3. ✅ 添加详细的 debug 日志
4. ✅ 向后兼容本地部署

**修改内容**:
```go
// 修改前：硬编码 127.0.0.1
target := net.JoinHostPort("127.0.0.1", port)

// 修改后：使用实际监听地址
if len(pm.server.conf.UDPListenAddrs) > 0 {
    addr := pm.server.conf.UDPListenAddrs[0]
    host := addr.IP.String()
    if host == "" || host == "0.0.0.0" || host == "::" {
        host = "127.0.0.1"
    }
    target = net.JoinHostPort(host, fmt.Sprintf("%d", addr.Port))
}
```

## 📞 支持

如果遇到问题：

1. 查看日志：`sudo journalctl -u AdGuardHome -n 100`
2. 检查配置：`cat /opt/AdGuardHome/AdGuardHome.yaml`
3. 测试 DNS：`dig @192.168.1.4 google.com`
4. 查看指标：`curl http://192.168.1.4:3000/control/cache_metrics`

---

**编译日期**: 2025-11-26  
**版本**: v0.107.0-dev (with prefetch fix)  
**状态**: ✅ 准备部署
