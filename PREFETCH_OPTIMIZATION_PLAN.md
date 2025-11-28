# Prefetch模块性能优化方案

## 当前性能问题

### Benchmark测试结果

```
BenchmarkPrefetch_Record-8              1,000,000    13,045 ns/op    33,094 B/op    8 allocs/op
BenchmarkPrefetch_RecordHotDomains-8    1,000,000    28,074 ns/op    62,461 B/op    8 allocs/op
```

**性能分析**:
- 吞吐量: 76,657 ops/s (普通记录)
- 延迟: 13.045 μs/op
- 内存分配: 33 KB/op (过高)
- 分配次数: 8 allocs/op

**对比域名匹配性能**:
- 域名匹配: 24,939,797 ops/s (48.85 ns/op, 0 B/op)
- Prefetch记录: 76,657 ops/s (13,045 ns/op, 33 KB/op)
- **性能差距**: 325倍

## 性能瓶颈分析

### 1. 内存分配过多 (33 KB/op)

**问题代码** (`record` 函数):

```go
// 每次调用都创建新的slice
timestamps := shard.hitTimestamps[domain]
timestamps = append(timestamps, now)  // 可能触发内存重新分配

validTimestamps := make([]time.Time, 0, len(timestamps))  // 新分配
for _, ts := range timestamps {
    if !ts.Before(cutoff) {
        validTimestamps = append(validTimestamps, ts)
    }
}
```

**问题**:
- 每次Record都创建新的slice
- 频繁的append操作触发内存重新分配
- 没有复用内存

### 2. 时间窗口过滤效率低

**问题代码**:

```go
// 每次都遍历所有timestamps
cutoff := now.Add(-pm.timeWindow)
validTimestamps := make([]time.Time, 0, len(timestamps))
for _, ts := range timestamps {
    if !ts.Before(cutoff) {
        validTimestamps = append(validTimestamps, ts)
    }
}
```

**问题**:
- O(n)复杂度，每次都遍历所有时间戳
- 对于热门域名，timestamps可能很长
- 没有利用时间戳已排序的特性

### 3. 锁竞争

**问题代码**:

```go
shard.mu.Lock()
defer shard.mu.Unlock()

// 长时间持有锁
// 1. 时间戳过滤
// 2. 计算hit count
// 3. 更新多个map
// 4. 检查cleanup条件
```

**问题**:
- 锁持有时间过长
- 在锁内进行复杂计算
- 热门域名会导致锁竞争

### 4. 不必要的字符串操作

**问题代码**:

```go
domain = dns.Fqdn(domain)  // 每次都调用
```

**问题**:
- 每次Record都调用Fqdn
- 可能涉及字符串拼接和内存分配

### 5. 过度的日志记录

**问题代码**:

```go
pm.logger.Debug("domain marked as hot (cache hit, bypassed threshold)",
    "domain", domain,
    "hit_count", hitCount,
    "expiry", expiry,
    "ttl", ttl)
```

**问题**:
- 即使日志级别是Error，Debug调用仍有开销
- 字符串格式化和参数打包

## 优化方案

### 优化1: 减少内存分配 (预期提升: 5-10倍)

#### 方案A: 使用环形缓冲区

```go
type domainShard struct {
    mu         sync.RWMutex
    domains    map[string]time.Time
    hits       map[string]int
    lastAccess map[string]time.Time
    
    // 使用环形缓冲区代替slice
    hitTimestamps map[string]*ringBuffer
}

type ringBuffer struct {
    timestamps [100]time.Time  // 固定大小数组
    head       int
    size       int
}

func (rb *ringBuffer) Add(t time.Time) {
    rb.timestamps[rb.head] = t
    rb.head = (rb.head + 1) % len(rb.timestamps)
    if rb.size < len(rb.timestamps) {
        rb.size++
    }
}

func (rb *ringBuffer) CountAfter(cutoff time.Time) int {
    count := 0
    for i := 0; i < rb.size; i++ {
        if !rb.timestamps[i].Before(cutoff) {
            count++
        }
    }
    return count
}
```

**优点**:
- 零内存分配
- 固定内存占用
- 快速访问

#### 方案B: 简化时间窗口跟踪

```go
type domainShard struct {
    mu         sync.RWMutex
    domains    map[string]time.Time
    
    // 简化：只记录最近N次访问的时间戳
    recentHits map[string]*hitCounter
    lastAccess map[string]time.Time
}

type hitCounter struct {
    count      int       // 当前计数
    windowStart time.Time // 窗口开始时间
}

func (pm *PrefetchManager) record(domain string, ttl uint32, isCacheHit bool) {
    shard := pm.getShard(domain)
    now := time.Now()
    
    shard.mu.Lock()
    
    counter := shard.recentHits[domain]
    if counter == nil {
        counter = &hitCounter{count: 1, windowStart: now}
        shard.recentHits[domain] = counter
    } else {
        // 检查是否需要重置窗口
        if now.Sub(counter.windowStart) > pm.timeWindow {
            counter.count = 1
            counter.windowStart = now
        } else {
            counter.count++
        }
    }
    
    shard.lastAccess[domain] = now
    
    // 快速判断是否为热门域名
    if isCacheHit || counter.count >= pm.threshold {
        expiry := now.Add(time.Duration(ttl) * time.Second)
        shard.domains[domain] = expiry
    }
    
    shard.mu.Unlock()
}
```

**优点**:
- 极少内存分配
- O(1)复杂度
- 简单高效

### 优化2: 减少锁持有时间 (预期提升: 2-3倍)

```go
func (pm *PrefetchManager) record(domain string, ttl uint32, isCacheHit bool) {
    if domain == "" || ttl == 0 {
        return
    }
    
    // 在锁外进行字符串处理
    domain = dns.Fqdn(domain)
    now := time.Now()
    expiry := now.Add(time.Duration(ttl) * time.Second)
    cutoff := now.Add(-pm.timeWindow)
    
    shard := pm.getShard(domain)
    
    // 最小化锁持有时间
    shard.mu.Lock()
    
    counter := shard.recentHits[domain]
    if counter == nil {
        counter = &hitCounter{count: 1, windowStart: now}
        shard.recentHits[domain] = counter
    } else {
        if now.Sub(counter.windowStart) > pm.timeWindow {
            counter.count = 1
            counter.windowStart = now
        } else {
            counter.count++
        }
    }
    
    hitCount := counter.count
    shard.lastAccess[domain] = now
    
    if isCacheHit || hitCount >= pm.threshold {
        shard.domains[domain] = expiry
    }
    
    shard.mu.Unlock()
    
    // 在锁外进行日志记录
    if pm.logger.Enabled(nil, slog.LevelDebug) && (isCacheHit || hitCount >= pm.threshold) {
        pm.logger.Debug("domain marked as hot",
            "domain", domain,
            "hit_count", hitCount)
    }
}
```

### 优化3: 使用对象池 (预期提升: 2倍)

```go
var hitCounterPool = sync.Pool{
    New: func() interface{} {
        return &hitCounter{}
    },
}

func (pm *PrefetchManager) record(domain string, ttl uint32, isCacheHit bool) {
    // ... 前置检查 ...
    
    shard := pm.getShard(domain)
    shard.mu.Lock()
    
    counter := shard.recentHits[domain]
    if counter == nil {
        counter = hitCounterPool.Get().(*hitCounter)
        counter.count = 1
        counter.windowStart = now
        shard.recentHits[domain] = counter
    } else {
        // ... 更新逻辑 ...
    }
    
    shard.mu.Unlock()
}

// 在cleanup时回收对象
func (pm *PrefetchManager) cleanupInternal(force bool) {
    // ... cleanup逻辑 ...
    
    for _, c := range candidates {
        c.shard.mu.Lock()
        
        // 回收hitCounter对象
        if counter := c.shard.recentHits[c.domain]; counter != nil {
            hitCounterPool.Put(counter)
        }
        
        delete(c.shard.recentHits, c.domain)
        // ... 其他删除操作 ...
        
        c.shard.mu.Unlock()
    }
}
```

### 优化4: 批量操作 (预期提升: 1.5倍)

```go
// 批量记录多个域名
func (pm *PrefetchManager) RecordBatch(records []DomainRecord) {
    // 按shard分组
    shardGroups := make(map[int][]DomainRecord)
    for _, record := range records {
        domain := dns.Fqdn(record.Domain)
        shardIdx := pm.getShardIndex(domain)
        shardGroups[shardIdx] = append(shardGroups[shardIdx], record)
    }
    
    // 每个shard只锁一次
    for shardIdx, group := range shardGroups {
        shard := pm.shards[shardIdx]
        shard.mu.Lock()
        
        for _, record := range group {
            // 批量更新
            pm.recordInternal(shard, record.Domain, record.TTL, record.IsCacheHit)
        }
        
        shard.mu.Unlock()
    }
}
```

### 优化5: 使用原子操作 (预期提升: 1.2倍)

```go
type hitCounter struct {
    count       atomic.Int32
    windowStart atomic.Int64 // Unix timestamp
}

func (pm *PrefetchManager) record(domain string, ttl uint32, isCacheHit bool) {
    // ... 前置处理 ...
    
    shard := pm.getShard(domain)
    
    // 先尝试读锁快速路径
    shard.mu.RLock()
    counter := shard.recentHits[domain]
    shard.mu.RUnlock()
    
    if counter != nil {
        // 使用原子操作更新计数
        nowUnix := now.Unix()
        windowStart := counter.windowStart.Load()
        
        if nowUnix - windowStart > int64(pm.timeWindow.Seconds()) {
            // 需要重置窗口，使用写锁
            shard.mu.Lock()
            counter.count.Store(1)
            counter.windowStart.Store(nowUnix)
            shard.mu.Unlock()
        } else {
            // 只增加计数，无需锁
            counter.count.Add(1)
        }
    } else {
        // 新域名，需要写锁
        shard.mu.Lock()
        // 双重检查
        if counter = shard.recentHits[domain]; counter == nil {
            counter = &hitCounter{}
            counter.count.Store(1)
            counter.windowStart.Store(now.Unix())
            shard.recentHits[domain] = counter
        }
        shard.mu.Unlock()
    }
}
```

## 综合优化方案

### 推荐实施顺序

1. **优化1B: 简化时间窗口跟踪** (最大收益)
   - 预期提升: 5-10倍
   - 实施难度: 低
   - 风险: 低

2. **优化2: 减少锁持有时间**
   - 预期提升: 2-3倍
   - 实施难度: 低
   - 风险: 低

3. **优化3: 使用对象池**
   - 预期提升: 2倍
   - 实施难度: 中
   - 风险: 低

4. **优化5: 使用原子操作**
   - 预期提升: 1.2倍
   - 实施难度: 中
   - 风险: 中

5. **优化4: 批量操作** (可选)
   - 预期提升: 1.5倍
   - 实施难度: 高
   - 风险: 中

### 预期总体提升

**保守估计**: 10-20倍性能提升
- 当前: 76,657 ops/s (13 μs/op, 33 KB/op)
- 优化后: 766,570 - 1,533,140 ops/s (0.65-1.3 μs/op, <1 KB/op)

**理想情况**: 20-50倍性能提升
- 优化后: 1,533,140 - 3,832,850 ops/s (0.26-0.65 μs/op, <100 B/op)

## 实施计划

### 阶段1: 核心优化 (1-2天)

1. 实施优化1B (简化时间窗口跟踪)
2. 实施优化2 (减少锁持有时间)
3. 运行benchmark验证

**目标**: 达到5-10倍性能提升

### 阶段2: 进阶优化 (2-3天)

1. 实施优化3 (对象池)
2. 实施优化5 (原子操作)
3. 运行benchmark验证

**目标**: 达到10-20倍性能提升

### 阶段3: 高级优化 (可选, 3-5天)

1. 实施优化4 (批量操作)
2. 性能调优
3. 压力测试

**目标**: 达到20-50倍性能提升

## 测试验证

### Benchmark测试

```bash
# 优化前
go test -bench=BenchmarkPrefetch_Record -benchmem -benchtime=10s

# 优化后
go test -bench=BenchmarkPrefetch_Record -benchmem -benchtime=10s

# 对比结果
```

### 性能指标

| 指标 | 优化前 | 目标 | 验收标准 |
|------|--------|------|---------|
| 吞吐量 | 76,657 ops/s | 766,570 ops/s | >500,000 ops/s |
| 延迟 | 13,045 ns/op | 1,300 ns/op | <2,000 ns/op |
| 内存分配 | 33,094 B/op | 1,000 B/op | <5,000 B/op |
| 分配次数 | 8 allocs/op | 2 allocs/op | <4 allocs/op |

## 风险评估

### 低风险

- 优化1B: 简化逻辑，降低复杂度
- 优化2: 代码重构，不改变语义

### 中风险

- 优化3: 需要仔细管理对象生命周期
- 优化5: 原子操作需要正确的内存顺序

### 高风险

- 优化4: 改变API接口，需要调用方配合

## 兼容性

### 向后兼容

- 所有优化保持API不变
- 配置参数保持兼容
- 行为语义保持一致

### 测试覆盖

- 单元测试: 覆盖所有优化路径
- 集成测试: 验证整体功能
- 性能测试: 验证性能提升
- 压力测试: 验证稳定性

## 总结

Prefetch模块当前性能瓶颈主要在于:
1. 过多的内存分配 (33 KB/op)
2. 低效的时间窗口过滤
3. 锁竞争

通过实施推荐的优化方案，预期可以达到:
- **10-20倍性能提升** (保守)
- **20-50倍性能提升** (理想)
- 内存分配降低到 <1 KB/op
- 延迟降低到 <2 μs/op

这将使Prefetch模块的性能接近域名匹配模块的水平。

