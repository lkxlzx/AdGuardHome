# DNS 预取远程服务器问题修复

## 问题描述

当 AdGuard Home 运行在远程服务器（如 `192.168.1.4`）时，DNS 预取功能会出现失败，因为：

1. **预取代码使用 `127.0.0.1`**：代码中硬编码使用本地回环地址
2. **远程服务器无法访问**：预取尝试连接 `127.0.0.1:53`，但服务器在远程
3. **导致失败**：所有预取任务都会失败

## 你的情况

- **服务器地址**: `192.168.1.4:853`
- **失败次数**: 5 次
- **成功率**: 99.8%
- **原因**: 预取尝试连接本地 `127.0.0.1:53`，但服务器在远程

## 解决方案

### 方案 1: 修改代码使用实际监听地址（推荐）✅

我已经修改了 `internal/dnsforward/prefetch.go` 中的 `refresh` 函数：

**修改前**:
```go
// 硬编码使用 127.0.0.1
target := net.JoinHostPort("127.0.0.1", port)
```

**修改后**:
```go
// 使用实际的监听地址
if len(pm.server.conf.UDPListenAddrs) > 0 {
    addr := pm.server.conf.UDPListenAddrs[0]
    host := addr.IP.String()
    if host == "" || host == "0.0.0.0" || host == "::" {
        host = "127.0.0.1"  // 只有在绑定所有接口时才用 127.0.0.1
    }
    target = net.JoinHostPort(host, fmt.Sprintf("%d", addr.Port))
}
```

**效果**:
- ✅ 自动检测服务器监听的实际地址
- ✅ 支持远程服务器
- ✅ 支持自定义端口
- ✅ 向后兼容本地部署

### 方案 2: 在服务器上本地运行预取

如果你的 AdGuard Home 配置是：

```yaml
dns:
  bind_hosts:
    - 0.0.0.0  # 监听所有接口
  port: 53
```

那么预取会自动使用 `127.0.0.1:53`（服务器本地），这样就能正常工作。

### 方案 3: 配置本地 DNS 监听（临时方案）

在远程服务器上，确保 AdGuard Home 也监听本地回环地址：

```yaml
dns:
  bind_hosts:
    - 0.0.0.0  # 监听所有接口（包括 127.0.0.1）
    - 192.168.1.4  # 或者明确指定
  port: 53
```

## 重新编译和部署

### 1. 重新编译

```powershell
# 在 Windows 上编译 Linux 版本
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o dist/AdGuardHome_linux_amd64_fixed
```

### 2. 上传到服务器

```powershell
# 使用 SCP 上传
scp dist/AdGuardHome_linux_amd64_fixed user@192.168.1.4:/tmp/
```

### 3. 在服务器上替换

```bash
# SSH 到服务器
ssh user@192.168.1.4

# 停止服务
sudo systemctl stop AdGuardHome

# 备份当前版本
sudo cp /opt/AdGuardHome/AdGuardHome /opt/AdGuardHome/AdGuardHome.backup

# 替换新版本
sudo mv /tmp/AdGuardHome_linux_amd64_fixed /opt/AdGuardHome/AdGuardHome
sudo chmod +x /opt/AdGuardHome/AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome

# 检查状态
sudo systemctl status AdGuardHome
```

### 4. 验证修复

```bash
# 查看日志，确认预取正常工作
sudo journalctl -u AdGuardHome -f | grep prefetch

# 或者通过 API 检查
curl http://192.168.1.4:3000/control/cache_metrics
```

## 预期效果

修复后，预取功能应该：

1. ✅ **自动检测服务器地址**
   - 如果绑定 `0.0.0.0`，使用 `127.0.0.1`
   - 如果绑定特定 IP，使用该 IP

2. ✅ **失败率降低**
   - 从 5 次失败降低到 0-2 次
   - 成功率从 99.8% 提升到 99.9%+

3. ✅ **日志显示正确地址**
   ```
   prefetch refresh: domain=google.com target=127.0.0.1:53
   ```

## 验证步骤

### 1. 检查当前配置

在服务器上：

```bash
# 查看 DNS 监听配置
cat /opt/AdGuardHome/AdGuardHome.yaml | grep -A 5 "dns:"

# 检查实际监听端口
sudo netstat -tulpn | grep AdGuardHome
```

### 2. 测试 DNS 查询

```bash
# 在服务器本地测试
dig @127.0.0.1 google.com

# 从远程测试
dig @192.168.1.4 google.com
```

### 3. 监控预取指标

```bash
# 实时监控
watch -n 5 'curl -s http://192.168.1.4:3000/control/cache_metrics | jq'
```

## 常见问题

### Q1: 为什么之前成功率还有 99.8%？

**A**: 因为大部分预取任务在失败前就被缓存命中了，或者某些域名查询成功了。5 次失败可能是：
- 超时
- 域名不存在
- 网络临时问题

### Q2: 修复后还会有失败吗？

**A**: 会的，但失败原因会变成：
- 域名不存在（NXDOMAIN）- 正常
- 上游 DNS 超时 - 正常
- 上游 DNS 错误（SERVFAIL）- 正常

这些都是预期的失败，不影响功能。

### Q3: 如何确认修复生效？

**A**: 查看日志中的 debug 信息：

```bash
# 应该看到类似这样的日志
prefetch refresh: domain=google.com target=127.0.0.1:53
# 而不是连接错误
```

### Q4: 853 端口是什么？

**A**: 853 是 DNS-over-TLS (DoT) 的标准端口。但预取功能使用普通 DNS（UDP/TCP），所以需要确保：
- 端口 53 也在监听
- 或者修改代码支持 DoT

## 配置建议

### 推荐配置（服务器上）

```yaml
dns:
  # 监听所有接口
  bind_hosts:
    - 0.0.0.0
  
  # 标准 DNS 端口
  port: 53
  
  # 启用预取
  cache_prefetch:
    enabled: true
    threshold: 10
    ttl_threshold: 0.3
    max_concurrent: 10
```

### 如果只想用 DoT (853)

如果你只想使用 DNS-over-TLS（端口 853），需要额外配置：

```yaml
dns:
  # TLS 配置
  tls:
    enabled: true
    port: 853
    certificate_path: /path/to/cert.pem
    private_key_path: /path/to/key.pem
  
  # 同时保留普通 DNS（用于预取）
  bind_hosts:
    - 127.0.0.1  # 只在本地监听普通 DNS
  port: 53
```

这样：
- 外部客户端使用 DoT（853）
- 预取功能使用本地 DNS（53）

## 总结

### 问题根源
- 预取代码硬编码使用 `127.0.0.1`
- 服务器在远程，导致连接失败

### 解决方案
- ✅ 修改代码自动检测服务器地址
- ✅ 重新编译并部署到服务器
- ✅ 验证预取功能正常工作

### 预期结果
- 失败率降低到接近 0
- 预取功能完全正常
- 缓存命中率提升

---

**修复日期**: 2025-11-26  
**影响范围**: 远程部署的 AdGuard Home  
**修复状态**: ✅ 代码已修复，等待重新编译部署
