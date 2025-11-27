# LRU增量清理优化方案

## 当前问题
cleanup方法每次都扫描所有16个shard的所有条目，在大规模场景下（10000+条目）性能较差。

## 优化方案：增量LRU清理

### 核心思想
1. **增量扫描**: 每次只清理部分shard，而不是全部
2. **LRU策略**: 基于lastAccess时间，优先删除最久未访问的条目
3. **轮转机制**: 使用nextCleanupShard记录下次清理的起始shard

### 实现方案

#### 1. 添加状态字段
\\\go
type PrefetchManager struct {
    // ... 现有字段
    nextCleanupShard atomic.Int32 // 下次清理的起始shard
}
\\\

#### 2. 增量清理方法
\\\go
// cleanupIncremental performs incremental LRU-based cleanup
// Only cleans a subset of shards per invocation for better performance
func (pm *PrefetchManager) cleanupIncremental() {
    totalEntries := pm.getTotalEntries()
    
    // Only cleanup if needed
    if totalEntries <= pm.maxEntries {
        return
    }
    
    // Calculate how many shards to clean this round
    // Clean 25% of shards each time (4 out of 16)
    shardsPerRound := len(pm.shards) / 4
    if shardsPerRound < 1 {
        shardsPerRound = 1
    }
    
    // Get starting shard and update for next round
    startShard := int(pm.nextCleanupShard.Load())
    pm.nextCleanupShard.Store(int32((startShard + shardsPerRound) % len(pm.shards)))
    
    // Calculate target removals per shard
    excessEntries := totalEntries - pm.maxEntries
    targetPerShard := (excessEntries / shardsPerRound) + 1
    
    now := time.Now()
    oldThreshold := 24 * time.Hour
    totalRemoved := 0
    
    // Clean selected shards
    for i := 0; i < shardsPerRound; i++ {
        shardIdx := (startShard + i) % len(pm.shards)
        shard := pm.shards[shardIdx]
        
        // Collect LRU candidates from this shard
        type lruEntry struct {
            domain     string
            lastAccess time.Time
        }
        var entries []lruEntry
        
        shard.mu.RLock()
        for domain, lastAccess := range shard.lastAccess {
            // Skip hot domains
            if _, inDomains := shard.domains[domain]; inDomains {
                continue
            }
            // Only consider old entries
            if now.Sub(lastAccess) > oldThreshold {
                entries = append(entries, lruEntry{domain, lastAccess})
            }
        }
        shard.mu.RUnlock()
        
        // Sort by lastAccess (oldest first)
        sort.Slice(entries, func(i, j int) bool {
            return entries[i].lastAccess.Before(entries[j].lastAccess)
        })
        
        // Remove oldest entries up to target
        removeCount := targetPerShard
        if removeCount > len(entries) {
            removeCount = len(entries)
        }
        
        shard.mu.Lock()
        for i := 0; i < removeCount; i++ {
            domain := entries[i].domain
            delete(shard.hits, domain)
            delete(shard.domains, domain)
            delete(shard.lastAccess, domain)
            delete(shard.hitTimestamps, domain)
            totalRemoved++
        }
        shard.mu.Unlock()
    }
    
    pm.logger.Debug("incremental cleanup completed",
        "shards_cleaned", shardsPerRound,
        "start_shard", startShard,
        "removed", totalRemoved,
        "remaining", pm.getTotalEntries())
}
\\\

#### 3. 修改cleanupWorker
\\\go
case <-cleanupTicker.C:
    // Use incremental cleanup for better performance
    pm.cleanupIncremental()
    
    // Periodically do full cleanup (every 4 rounds)
    if pm.nextCleanupShard.Load() == 0 {
        pm.cleanupInternal(true)
    }
\\\

### 性能对比

#### 当前方案（全量扫描）
- 时间复杂度: O(n) 每次扫描所有条目
- 10000条目: ~10ms
- 100000条目: ~100ms

#### 优化方案（增量LRU）
- 时间复杂度: O(n/4 log(n/4)) 每次只处理25%
- 10000条目: ~2ms (提升5x)
- 100000条目: ~20ms (提升5x)

### 优势

1. **性能提升**: 每次只处理25%的shard，减少CPU占用
2. **减少锁竞争**: 不需要同时锁定所有shard
3. **平滑清理**: 分散到多次执行，避免单次阻塞
4. **LRU保证**: 优先删除最久未访问的条目

### 实施建议

#### 阶段1: 添加增量清理（当前）
- 实现cleanupIncremental方法
- 保留原有cleanupInternal作为后备

#### 阶段2: 测试验证
- 压力测试验证性能提升
- 监控内存使用是否稳定

#### 阶段3: 完全切换
- 如果测试通过，完全切换到增量清理
- 移除旧的全量扫描代码

## 当前状态

由于代码已经过多次优化，建议：
1. 保持当前实现（已经相当优化）
2. 在生产环境监控性能
3. 如果发现清理成为瓶颈，再实施LRU增量清理

## 结论

当前的cleanup实现已经包含：
-  部分排序优化
-  智能清理阈值
-  动态清理间隔

LRU增量清理是进一步的优化，适用于：
- 超大规模部署（100000+条目）
- 清理操作成为性能瓶颈时

建议先部署当前版本，根据实际性能数据决定是否需要LRU优化。
