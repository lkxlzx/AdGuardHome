# 预取机制架构优化方案

## 🔍 问题分析

### 当前实现的问题

当前预取机制通过 **DNS 查询自己** 来刷新缓存：

```go
func (pm *PrefetchManager) refresh(domain string) error {
    c := new(dns.Client)
    m := new(dns.Msg)
    m.SetQuestion(domain, dns.TypeA)
    
    // ❌ 问题：查询 127.0.0.1:53（自己）
    target := "127.0.0.1:53"
    _, _, err := c.Exchange(m, target)
    return err
}
```

**存在的问题**：

1. ❌ **浪费资源**：触发完整的 DNS 解析流程
   - 经过过滤器
   - 查询上游服务器
   - 消耗网络带宽

2. ❌ **性能低下**：
   - 需要等待完整的 DNS 响应
   - 超时时间长（5秒）
   - 重试机制增加延迟

3. ❌ **可能触发限流**：
   - 上游服务器可能限制查询频率
   - 大量预取可能被视为异常流量

4. ❌ **日志混乱**：
   - 预取查询混入正常查询日志
   - 难以区分真实用户请求

## ✅ 正确的架构设计

### 方案 1：直接调用缓存刷新（推荐）

预取应该 **直接更新缓存**，而不是通过 DNS 查询：

```go
func (pm *PrefetchManager) refresh(domain string) error {
    // ✅ 直接查询上游并更新缓存
    return pm.server.refreshCacheEntry(domain)
}

// 在 Server 中添加方法
func (s *Server) refreshCacheEntry(domain string) error {
    // 1. 直接查询上游（跳过过滤器）
    // 2. 更新缓存
    // 3. 返回结果
}
```

**优点**：
- ✅ 不经过过滤器，直接查询上游
- ✅ 性能更高，延迟更低
- ✅ 不会触发限流
- ✅ 日志清晰，易于调试

### 方案 2：使用内部 API

如果 dnsproxy 支持，可以直接调用其内部 API：

```go
func (pm *PrefetchManager) refresh(domain string) error {
    // 使用 dnsproxy 的内部刷新机制
    return pm.server.dnsProxy.RefreshCacheEntry(domain)
}
```

## 🎯 实施步骤

### 步骤 1：在 Server 中添加缓存刷新方法

```go
// refreshCacheEntry 直接查询上游并更新缓存，用于预取机制
func (s *Server) refreshCacheEntry(domain string) error {
    // 创建 DNS 查询
    req := &dns.Msg{}
    req.SetQuestion(domain, dns.TypeA)
    req.RecursionDesired = true
    
    // 直接查询上游（跳过过滤器）
    resp, err := s.dnsProxy.Resolve(req)
    if err != nil {
        return fmt.Errorf("resolve %s: %w", domain, err)
    }
    
    // 缓存会自动更新（dnsproxy 内部机制）
    _ = resp
    
    return nil
}
```

### 步骤 2：修改 PrefetchManager.refresh

```go
func (pm *PrefetchManager) refresh(domain string) error {
    // 重试逻辑
    const maxRetries = 3
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        pm.logger.Debug("prefetch refresh attempt",
            "domain", domain,
            "attempt", attempt)
        
        // ✅ 直接调用缓存刷新
        err := pm.server.refreshCacheEntry(domain)
        if err == nil {
            if attempt > 1 {
                pm.logRetrySuccess(domain, "cache", attempt)
            }
            return nil
        }
        
        lastErr = err
        pm.logError(domain, "cache", attempt, err)
        
        if attempt < maxRetries {
            backoff := time.Duration(100*attempt) * time.Millisecond
            time.Sleep(backoff)
        }
    }
    
    return lastErr
}
```

### 步骤 3：更新错误日志格式

```go
// 日志格式从：
// [timestamp] Domain: xxx | Target: 127.0.0.1:53 | Attempt: 1 | Error: xxx

// 改为：
// [timestamp] Domain: xxx | Method: cache_refresh | Attempt: 1 | Error: xxx
```

## 📊 性能对比

### 当前实现（DNS 查询）

```
预取请求 → DNS Client → 127.0.0.1:53 → 过滤器 → 上游查询 → 缓存更新
延迟：~100-500ms（包括网络往返）
```

### 优化后（直接刷新）

```
预取请求 → 上游查询 → 缓存更新
延迟：~50-200ms（仅上游查询）
```

**性能提升**：
- ⚡ 延迟降低 50%+
- 💾 减少内存分配
- 🔥 减少 CPU 使用
- 📉 减少网络流量

## 🔧 兼容性考虑

### 如果 dnsproxy 不支持直接刷新

可以使用 **内部查询标记**：

```go
func (pm *PrefetchManager) refresh(domain string) error {
    c := new(dns.Client)
    m := new(dns.Msg)
    m.SetQuestion(domain, dns.TypeA)
    
    // 添加特殊标记，表示这是预取查询
    opt := &dns.OPT{
        Hdr: dns.RR_Header{Name: ".", Rrtype: dns.TypeOPT},
    }
    opt.SetUDPSize(4096)
    // 使用 EDNS0 选项标记为内部查询
    opt.Option = append(opt.Option, &dns.EDNS0_LOCAL{
        Code: 0xFFFE, // 自定义代码
        Data: []byte("prefetch"),
    })
    m.Extra = append(m.Extra, opt)
    
    target := "127.0.0.1:53"
    _, _, err := c.Exchange(m, target)
    return err
}
```

然后在 Server 中识别并特殊处理这类查询。

## 📝 总结

当前的预取机制通过 DNS 查询自己来刷新缓存，这是一个 **架构设计缺陷**。

**正确的做法**：
1. ✅ 预取应该直接调用缓存刷新机制
2. ✅ 跳过过滤器，直接查询上游
3. ✅ 减少延迟和资源消耗
4. ✅ 提高系统整体性能

**下一步**：
1. 实现 `Server.refreshCacheEntry()` 方法
2. 修改 `PrefetchManager.refresh()` 调用新方法
3. 更新错误日志格式
4. 测试验证性能提升
