# DNS路由规则优先级功能 - 完整实现报告

## 功能概述

为DNS路由过滤器添加了优先级（Priority）支持，允许用户控制规则的匹配顺序。优先级值越小，优先级越高。

## 实现细节

### 1. 数据结构

#### 后端 (Go)
```go
// FilterYAML - 过滤器配置
type FilterYAML struct {
    // ... 其他字段
    UpdateInterval int       `yaml:"update_interval"`
    Priority       int       `yaml:"priority"`  // 新增字段
}
```

#### 前端 (TypeScript)
```typescript
type FormValues = {
    // ... 其他字段
    updateInterval?: number;
    priority?: number;  // 新增字段
};
```

### 2. 前端实现

#### 2.1 表单输入
- 文件: `client/src/components/Filters/Form.tsx`
- 添加了priority输入字段（0-100范围）
- 默认值为0
- 包含中文提示："数字越小优先级越高"

#### 2.2 数据传递
- 文件: `client/src/components/Filters/DnsRouting.tsx`
- `handleSubmit`函数提取priority值
- 传递给`addFilter`和`editFilter` actions

#### 2.3 Actions
- 文件: `client/src/actions/filtering.ts`
- `addFilter`: 接收priority参数，发送到API
- `editFilter`: 修复了字段删除bug，正确传递priority

#### 2.4 数据规范化
- 文件: `client/src/helpers/helpers.tsx`
- `normalizeFilters`: 从API响应中提取priority
- `getCurrentFilter`: 读取当前过滤器的priority值

### 3. 后端实现

#### 3.1 HTTP API
- 文件: `internal/filtering/http.go`
- `filterAddJSON`: 包含Priority字段
- `filterURLReqData`: 包含Priority字段
- `handleFilteringAddURL`: 创建FilterYAML时设置Priority
- `handleFilteringSetURL`: 更新时处理Priority
- `filterToJSON`: 返回Priority给前端

#### 3.2 过滤器属性更新
- 文件: `internal/filtering/filter.go`
- `filterSetProperties`: 更新Priority字段
- 逻辑与UpdateInterval一致

#### 3.3 配置保存
- 文件: `internal/dnsforward/dnsforward.go`
- `WriteDiskConfig`: 保存UpstreamGroups、DnsRoutingRules、CustomDomainRules

### 4. 匹配逻辑实现

#### 4.1 过滤器加载排序
- 文件: `internal/filtering/filter.go`
- `enableFiltersLocked`函数中：
  - 收集启用的DNS路由过滤器
  - 按Priority升序排序（使用`slices.SortFunc`）
  - 按排序后的顺序加载到filtering引擎

```go
// 排序逻辑
slices.SortFunc(sortedDnsRoutingFilters, func(a, b FilterYAML) int {
    return a.Priority - b.Priority
})
```

#### 4.2 规则匹配选择
- 文件: `internal/filtering/filtering.go`
- `getUpstreamGroupByPriority`函数：
  - 当多个规则匹配时
  - 遍历所有匹配的规则
  - 选择Priority值最小的规则
  - 返回对应的UpstreamGroup

```go
// 选择逻辑
for _, rule := range matchedRules {
    filterID := rule.GetFilterListID()
    priority := filterPriority[filterID]
    
    if !found || priority < minPriority {
        minPriority = priority
        selectedUpstream = filterUpstream[filterID]
        found = true
    }
}
```

### 5. 双重保护机制

系统在两个层面确保priority正确工作：

1. **加载时排序**: 过滤器按priority排序后加载，filtering引擎按顺序匹配
2. **匹配时选择**: 如果多个规则都匹配，显式选择priority最小的

这种双重机制确保了即使filtering引擎的行为发生变化，priority功能仍然能正确工作。

## 修复的问题

### 问题1: Priority值保存为0
**原因**: `editFilter` action中错误地删除了刚设置的priority字段

**修复**: 
```typescript
// 之前（错误）
const filterData: any = { ...data };
if (data.priority !== undefined) {
    filterData.priority = data.priority;
    delete filterData.priority;  // ❌ 删除了刚设置的值
}

// 之后（正确）
const filterData: any = {};
if (data.priority !== undefined) {
    filterData.priority = data.priority;  // ✅ 不再删除
}
```

### 问题2: 匹配逻辑不按priority工作
**原因**: Filtering引擎按加载顺序匹配，没有考虑priority

**修复**: 
1. 在加载时按priority排序过滤器
2. 添加`getUpstreamGroupByPriority`函数作为额外保护

## 测试场景

### 场景1: 精确规则优先于通用规则
```yaml
dns_routing_filters:
  - name: CN规则
    priority: 10
    # 包含 ||qq.com^
  - name: 腾讯规则
    priority: 20
    # 包含 ||qq.com^
```

查询`qq.com`时：
- 两个规则都匹配
- 系统选择priority=10的CN规则
- 使用CN规则对应的上游组

### 场景2: 多层级规则
```yaml
dns_routing_filters:
  - name: 特定域名
    priority: 5
  - name: 公司域名
    priority: 10
  - name: 国内域名
    priority: 20
  - name: 全球域名
    priority: 30
```

匹配顺序：特定域名 → 公司域名 → 国内域名 → 全球域名

## 配置示例

```yaml
dns_routing_filters:
  - enabled: true
    url: https://example.com/specific-rules.yaml
    name: 特定规则
    update_interval: 3600
    priority: 10  # 高优先级
    id: 1764049985
    upstream_group: group_specific
    
  - enabled: true
    url: https://example.com/general-rules.yaml
    name: 通用规则
    update_interval: 3600
    priority: 30  # 低优先级
    id: 1764049986
    upstream_group: group_general
```

## 日志示例

```
[debug] dns routing upstream group found filter_id=1764049985 upstream_group=group_specific priority=10
```

## 相关提交

1. `172c4d78` - Fix: Priority field being cleared when editing DNS routing filters
2. `1ae63722` - Implement priority-based DNS routing rule matching
3. `4fd1378d` - Fix: Sort DNS routing filters by priority when loading
4. `fc0b8ed5` - Fix code review issue #3: Improve state consistency in custom rules

## 技术债务和未来改进

### 已知限制
1. Priority范围限制在0-100，但后端没有强制验证
2. 相同priority的规则匹配顺序未定义（依赖加载顺序）

### 建议改进
1. 添加priority范围验证（后端）
2. 为相同priority的规则定义明确的次级排序规则（如按名称）
3. 在UI中显示规则的实际匹配顺序
4. 添加priority冲突检测和警告

## 总结

Priority功能现已完全实现并经过测试。系统能够正确地：
- 保存和加载priority值
- 按priority排序过滤器
- 在匹配时选择最高优先级的规则
- 在日志中显示priority信息

该功能为用户提供了精确控制DNS路由规则匹配顺序的能力，解决了规则冲突的问题。
