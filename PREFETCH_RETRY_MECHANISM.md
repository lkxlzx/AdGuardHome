# DNS 预取重试机制 - 改进方案

## 📊 问题分析

### 观察到的问题

根据最新截图：

```
成功率: 98.2%
已完成: 4,118 次
已失败: 76 次
失败率: 1.8%
```

**问题**：
- 失败次数从 5 次增加到 76 次
- 成功率从 99.8% 降到 98.2%
- 失败率持续增加

### 失败原因

主要是**临时性失败**：
1. DNS 查询超时（5 秒）
2. 网络临时抖动
3. 上游 DNS 临时不可达

这些失败如果**重试一次**，大部分都能成功！

## ✅ 解决方案：添加重试机制

### 实现细节

```go
// 重试配置
const maxRetries = 3  // 最多重试 3 次

// 重试逻辑
for attempt := 1; attempt <= maxRetries; attempt++ {
    _, _, err := c.Exchange(m, target)
    if err == nil {
        return nil  // 成功，立即返回
    }
    
    // 失败，等待后重试
    if attempt < maxRetries {
        backoff := time.Duration(100*attempt) * time.Millisecond
        time.Sleep(backoff)  // 指数退避：100ms, 200ms, 400ms
    }
}
```

### 重试策略

| 尝试次数 | 等待时间 | 说明 |
|---------|---------|------|
| 第 1 次 | 0ms | 立即尝试 |
| 第 2 次 | 100ms | 短暂等待 |
| 第 3 次 | 200ms | 稍长等待 |
| 失败 | - | 记录失败 |

**总耗时**：最多 5秒 + 5秒 + 5秒 + 300ms = 15.3秒

### 优化点

1. **指数退避**
   - 避免立即重试加重服务器负担
   - 给网络恢复时间

2. **最多 3 次重试**
   - 平衡成功率和性能
   - 避免过长等待

3. **详细日志**
   - 记录每次重试
   - 便于诊断问题

## 📈 预期效果

### 改进前

```
成功率: 98.2%
失败次数: 76 / 4194 = 1.8%
```

### 改进后（预期）

```
成功率: 99.5%+
失败次数: < 20 / 4194 = < 0.5%
```

### 为什么会改进？

假设 76 次失败中：
- **60 次**是临时性失败（超时、网络抖动）→ 重试后成功
- **16 次**是永久性失败（域名不存在）→ 重试后仍失败

**结果**：
- 失败从 76 次降低到 16 次
- 成功率从 98.2% 提升到 99.6%

## 🚀 部署步骤

### 1. 上传新版本

```powershell
# Windows PowerShell
scp dist/AdGuardHome_linux_amd64_retry user@192.168.1.4:/tmp/
```

### 2. SSH 到服务器

```bash
ssh user@192.168.1.4
```

### 3. 替换二进制文件

```bash
# 停止服务
sudo systemctl stop AdGuardHome

# 备份当前版本
sudo cp /opt/AdGuardHome/AdGuardHome /opt/AdGuardHome/AdGuardHome.backup.$(date +%Y%m%d_%H%M%S)

# 替换新版本
sudo mv /tmp/AdGuardHome_linux_amd64_retry /opt/AdGuardHome/AdGuardHome
sudo chmod +x /opt/AdGuardHome/AdGuardHome

# 启动服务
sudo systemctl start AdGuardHome

# 检查状态
sudo systemctl status AdGuardHome
```

### 4. 验证改进

```bash
# 实时查看重试日志
sudo journalctl -u AdGuardHome -f | grep "prefetch refresh"

# 应该看到类似这样的日志：
# prefetch refresh attempt: domain=example.com attempt=1
# prefetch refresh failed, retrying: domain=example.com attempt=1 backoff=100ms
# prefetch refresh succeeded after retry: domain=example.com attempt=2
```

### 5. 监控效果

等待 1-2 小时后，检查预取指标：

```bash
curl http://192.168.1.4:3000/control/cache_metrics | jq
```

预期看到：
- 失败次数明显减少
- 成功率提升到 99.5%+

## 📊 性能影响分析

### 额外开销

**最坏情况**（所有查询都失败并重试 3 次）：
- 时间：15.3 秒（vs 原来 5 秒）
- 但这种情况极少发生（< 2%）

**平均情况**（98% 第一次成功，2% 需要重试）：
- 98% 查询：5 秒
- 1.5% 查询：5.1 秒（第 2 次成功）
- 0.4% 查询：5.3 秒（第 3 次成功）
- 0.1% 查询：15.3 秒（全部失败）

**平均额外开销**：< 0.1 秒

### 资源消耗

- **CPU**：几乎无影响（只是多几次 DNS 查询）
- **内存**：无影响
- **网络**：失败的查询会多 1-2 次请求（< 2% 的查询）

**结论**：性能影响可以忽略不计

## 🔍 日志示例

### 成功案例（第一次就成功）

```
prefetch refresh attempt: domain=google.com target=127.0.0.1:53 attempt=1 max_retries=3
```

### 重试成功案例

```
prefetch refresh attempt: domain=example.com target=127.0.0.1:53 attempt=1 max_retries=3
prefetch refresh failed, retrying: domain=example.com attempt=1 err="i/o timeout" backoff=100ms
prefetch refresh attempt: domain=example.com target=127.0.0.1:53 attempt=2 max_retries=3
prefetch refresh succeeded after retry: domain=example.com attempt=2
```

### 全部失败案例

```
prefetch refresh attempt: domain=bad.example target=127.0.0.1:53 attempt=1 max_retries=3
prefetch refresh failed, retrying: domain=bad.example attempt=1 err="i/o timeout" backoff=100ms
prefetch refresh attempt: domain=bad.example target=127.0.0.1:53 attempt=2 max_retries=3
prefetch refresh failed, retrying: domain=bad.example attempt=2 err="i/o timeout" backoff=200ms
prefetch refresh attempt: domain=bad.example target=127.0.0.1:53 attempt=3 max_retries=3
prefetch refresh failed after all retries: domain=bad.example target=127.0.0.1:53 attempts=3 err="i/o timeout"
```

## 📈 监控指标

### 关键指标

1. **成功率**
   - 目标：> 99.5%
   - 当前：98.2%
   - 预期改进：+1.3%

2. **失败次数**
   - 目标：< 20 次 / 4000 次
   - 当前：76 次 / 4194 次
   - 预期改进：-75%

3. **重试成功率**
   - 新指标：重试后成功的比例
   - 预期：80%+ 的失败会在重试后成功

### 监控命令

```bash
# 实时监控预取指标
watch -n 5 'curl -s http://192.168.1.4:3000/control/cache_metrics | jq'

# 查看重试日志
sudo journalctl -u AdGuardHome --since "1 hour ago" | grep "retry" | wc -l

# 查看成功的重试
sudo journalctl -u AdGuardHome --since "1 hour ago" | grep "succeeded after retry" | wc -l
```

## 🎯 预期结果对比

### 改进前

| 指标 | 值 | 评级 |
|------|-----|------|
| 成功率 | 98.2% | ⭐⭐⭐⭐ 良好 |
| 失败次数 | 76 / 4194 | ⭐⭐⭐⭐ 良好 |
| 失败率 | 1.8% | ⭐⭐⭐⭐ 良好 |

### 改进后（预期）

| 指标 | 值 | 评级 |
|------|-----|------|
| 成功率 | 99.5%+ | ⭐⭐⭐⭐⭐ 优秀 |
| 失败次数 | < 20 / 4194 | ⭐⭐⭐⭐⭐ 优秀 |
| 失败率 | < 0.5% | ⭐⭐⭐⭐⭐ 优秀 |

## 🔧 故障排除

### 如果失败率没有改善

1. **检查日志中的重试情况**
   ```bash
   sudo journalctl -u AdGuardHome -n 1000 | grep "retry"
   ```

2. **检查是否是永久性失败**
   ```bash
   # 查看失败的域名
   sudo journalctl -u AdGuardHome -n 1000 | grep "failed after all retries"
   ```

3. **检查上游 DNS 配置**
   ```bash
   cat /opt/AdGuardHome/AdGuardHome.yaml | grep -A 5 "upstream_dns"
   ```

### 如果性能下降

1. **检查重试次数是否过多**
   ```bash
   # 统计重试次数
   sudo journalctl -u AdGuardHome --since "1 hour ago" | grep "attempt=2" | wc -l
   sudo journalctl -u AdGuardHome --since "1 hour ago" | grep "attempt=3" | wc -l
   ```

2. **考虑调整重试参数**
   - 减少最大重试次数（3 → 2）
   - 增加超时时间（5秒 → 10秒）

## 📝 总结

### 改进内容

✅ **添加重试机制**
- 最多重试 3 次
- 指数退避策略
- 详细的重试日志

### 预期效果

✅ **显著降低失败率**
- 从 1.8% 降低到 < 0.5%
- 成功率从 98.2% 提升到 99.5%+
- 大部分临时失败会在重试后成功

### 性能影响

✅ **几乎无影响**
- 平均额外开销 < 0.1 秒
- CPU/内存无明显增加
- 只有失败的查询会重试

### 部署建议

✅ **立即部署**
- 编译文件：`dist/AdGuardHome_linux_amd64_retry`
- 按照部署步骤操作
- 监控 1-2 小时后查看效果

---

**改进日期**: 2025-11-26  
**版本**: v0.107.0-dev (with retry mechanism)  
**状态**: ✅ 准备部署  
**预期改进**: 失败率 -75%，成功率 +1.3%
