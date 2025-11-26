# 代码修复行动计划

## 修复优先级

根据代码审查报告，按优先级排列的修复任务。

---

## 🔴 P0 - 必须立即修复

### 任务 1: 实现真实的缓存统计数据
**预计时间**: 2-3 小时  
**难度**: 中等

**步骤**:
1. 在 `internal/dnsforward/dnsforward.go` 中添加缓存统计字段
2. 在 DNS 查询处理中记录缓存命中/未命中
3. 实现历史数据存储（可使用环形缓冲区）
4. 更新 `handleGetCacheMetrics` 使用真实数据

**代码示例**:
```go
// 在 Server 结构体中添加
type CacheStats struct {
    mu           sync.RWMutex
    totalQueries int64
    cacheHits    int64
    cacheMisses  int64
    history      [24]float64 // 24小时历史
    lastUpdate   time.Time
}

// 在查询处理中记录
func (s *Server) recordCacheHit(hit bool) {
    s.cacheStats.mu.Lock()
    defer s.cacheStats.mu.Unlock()
    
    s.cacheStats.totalQueries++
    if hit {
        s.cacheStats.cacheHits++
    } else {
        s.cacheStats.cacheMisses++
    }
}
```

---

### 任务 2: 修复 LastPrefetchTime
**预计时间**: 30 分钟  
**难度**: 简单

**步骤**:
1. 在 `PrefetchManager` 中添加 `lastPrefetchTime` 字段
2. 在 `refreshDomain` 成功后更新时间
3. 在 `GetMetrics` 中返回该时间
4. 更新 API 使用真实时间

**代码示例**:
```go
type PrefetchManager struct {
    // ... 现有字段
    lastPrefetchTime atomic.Value // stores time.Time
}

// 在刷新成功后
func (pm *PrefetchManager) refreshDomain(domain string) error {
    // ... 刷新逻辑
    if err == nil {
        pm.lastPrefetchTime.Store(time.Now())
    }
    return err
}

// 在 GetMetrics 中
func (pm *PrefetchManager) GetMetrics() map[string]int64 {
    metrics := make(map[string]int64)
    // ... 其他指标
    
    if t := pm.lastPrefetchTime.Load(); t != nil {
        if lastTime, ok := t.(time.Time); ok {
            metrics["last_prefetch_unix"] = lastTime.Unix()
        }
    }
    return metrics
}
```

---

## 🟡 P1 - 应该尽快修复

### 任务 3: 清理未使用的 Props 和导入
**预计时间**: 15 分钟  
**难度**: 简单

**文件**:
- `client/src/components/Dashboard/CacheMetrics.tsx`
- `client/src/components/Dashboard/PrefetchMetrics.tsx`

**修改**:
```typescript
// CacheMetrics.tsx
// 移除未使用的导入
// import { formatNumber } from '../../helpers/helpers';

// 移除未使用的 prop
interface CacheMetricsProps {}
const CacheMetrics = () => {
    // ... 组件代码
};

// PrefetchMetrics.tsx
// 移除未使用的 prop
interface PrefetchMetricsProps {}
const PrefetchMetrics = () => {
    // ... 组件代码
};
```

---

### 任务 4: 改进 API 错误处理
**预计时间**: 1 小时  
**难度**: 简单

**步骤**:
1. 添加错误状态
2. 检查 HTTP 响应状态
3. 显示用户友好的错误消息
4. 添加重试机制（可选）

**代码示例**:
```typescript
const CacheMetrics = () => {
    const { t } = useTranslation();
    const [data, setData] = useState<CacheMetricsData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchMetrics = async () => {
        try {
            const response = await fetch('/control/cache_metrics');
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            setData(result);
            setError(null);
        } catch (error) {
            console.error('Failed to fetch cache metrics:', error);
            setError(t('failed_to_load_cache_metrics'));
        } finally {
            setLoading(false);
        }
    };

    // ... useEffect

    if (error) {
        return (
            <Card type="card--full" bodyType="card-body">
                <div className="alert alert-danger" role="alert">
                    <svg className="icons icon--20 mr-2">
                        <use xlinkHref="#exclamation" />
                    </svg>
                    {error}
                    <button 
                        className="btn btn-sm btn-outline-primary ml-3"
                        onClick={() => {
                            setError(null);
                            setLoading(true);
                            fetchMetrics();
                        }}>
                        {t('retry')}
                    </button>
                </div>
            </Card>
        );
    }

    // ... 其余代码
};
```

**需要添加的翻译**:
```json
// en.json
"failed_to_load_cache_metrics": "Failed to load cache metrics",
"failed_to_load_prefetch_metrics": "Failed to load prefetch metrics",
"retry": "Retry",

// zh-cn.json
"failed_to_load_cache_metrics": "加载缓存指标失败",
"failed_to_load_prefetch_metrics": "加载预取指标失败",
"retry": "重试",
```

---

## 🟢 P2 - 建议修复

### 任务 5: 提取常量和配置
**预计时间**: 20 分钟  
**难度**: 简单

**创建配置文件**:
```typescript
// client/src/components/Dashboard/constants.ts
export const DASHBOARD_CONFIG = {
    REFRESH_INTERVAL: 30000, // 30 seconds
    CACHE_HISTORY_HOURS: 24,
    MAX_RETRY_ATTEMPTS: 3,
    RETRY_DELAY: 5000, // 5 seconds
} as const;
```

**使用配置**:
```typescript
import { DASHBOARD_CONFIG } from './constants';

const interval = setInterval(fetchMetrics, DASHBOARD_CONFIG.REFRESH_INTERVAL);
```

---

### 任务 6: 删除未使用的 CSS
**预计时间**: 5 分钟  
**难度**: 简单

**文件**: `client/src/components/Dashboard/MetricsCards.css`

**删除**:
```css
/* 删除这个未使用的类 */
.cache-hits-count {
    position: absolute;
    top: 15px;
    right: 15px;
    font-size: 0.875rem;
    font-weight: 600;
}
```

---

## 📋 实施时间表

### 第一天 (4-5 小时)
- ✅ 任务 3: 清理未使用代码 (15分钟)
- ✅ 任务 5: 提取常量 (20分钟)
- ✅ 任务 6: 删除未使用CSS (5分钟)
- ⏳ 任务 4: 改进错误处理 (1小时)
- ⏳ 任务 2: 修复 LastPrefetchTime (30分钟)
- ⏳ 任务 1: 实现真实缓存统计 (2-3小时)

### 测试和验证 (1-2 小时)
- 单元测试
- 集成测试
- 手动测试
- 性能测试

---

## 🧪 测试计划

### 单元测试
```go
// internal/dnsforward/cache_stats_test.go
func TestCacheStats(t *testing.T) {
    stats := &CacheStats{}
    
    // 测试记录命中
    stats.recordHit(true)
    assert.Equal(t, int64(1), stats.cacheHits)
    
    // 测试记录未命中
    stats.recordHit(false)
    assert.Equal(t, int64(1), stats.cacheMisses)
    
    // 测试命中率计算
    rate := stats.getHitRate()
    assert.Equal(t, 50.0, rate)
}
```

### 集成测试
```typescript
// client/src/components/Dashboard/__tests__/CacheMetrics.test.tsx
describe('CacheMetrics', () => {
    it('should display error message when API fails', async () => {
        // Mock failed API call
        global.fetch = jest.fn(() =>
            Promise.reject(new Error('Network error'))
        );
        
        render(<CacheMetrics />);
        
        await waitFor(() => {
            expect(screen.getByText(/failed to load/i)).toBeInTheDocument();
        });
    });
    
    it('should retry on error', async () => {
        // Test retry functionality
    });
});
```

---

## 📊 预期改进

### 修复前
- ❌ 显示假数据
- ❌ 错误无提示
- ❌ 代码冗余
- ⚠️ 用户体验差

### 修复后
- ✅ 显示真实数据
- ✅ 完善错误处理
- ✅ 代码整洁
- ✅ 用户体验好

---

## 🎯 成功标准

1. **功能性**
   - [ ] 缓存命中率显示真实数据
   - [ ] 最后预取时间准确
   - [ ] API 错误有友好提示
   - [ ] 所有功能正常工作

2. **代码质量**
   - [ ] 无 TypeScript 警告
   - [ ] 无未使用的代码
   - [ ] 代码通过 lint 检查
   - [ ] 测试覆盖率 > 80%

3. **用户体验**
   - [ ] 加载状态清晰
   - [ ] 错误提示友好
   - [ ] 可以重试失败的请求
   - [ ] 响应速度快

4. **性能**
   - [ ] 无内存泄漏
   - [ ] API 响应时间 < 100ms
   - [ ] 前端渲染流畅
   - [ ] 资源占用合理

---

## 📝 检查清单

在提交代码前，确保：

- [ ] 所有 P0 问题已修复
- [ ] 所有 P1 问题已修复
- [ ] 代码已通过测试
- [ ] 文档已更新
- [ ] 无编译警告
- [ ] 无 lint 错误
- [ ] 已手动测试所有功能
- [ ] 已更新 CHANGELOG
- [ ] 已更新版本号

---

**创建时间**: 2025-11-26  
**预计完成时间**: 1-2 天  
**负责人**: 开发团队
