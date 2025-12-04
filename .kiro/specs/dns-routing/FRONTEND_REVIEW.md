# DNS路由前端代码审查报告

## 审查目的
为对接后端API做准备，全面审查DNS路由前端代码的数据结构、API调用和需要调整的部分。

## 1. 前端架构概览

### 1.1 核心文件结构
```
client/src/
├── actions/dnsRouting.ts          # Redux actions
├── reducers/dnsRouting.ts         # Redux reducer
├── components/Filters/
│   ├── DnsRouting.tsx            # 主组件
│   ├── DnsRoutingTable.tsx       # 规则表格
│   ├── DnsRoutingForm.tsx        # 规则表单
│   ├── DnsRoutingModal.tsx       # 规则对话框
│   ├── DnsRoutingActions.tsx     # 操作按钮
│   ├── CustomRuleModal.tsx       # 自定义规则对话框
│   └── CustomRulesTable.tsx      # 自定义规则表格
├── hooks/useDnsRoutingCustomRules.ts  # 自定义规则Hook
├── containers/DnsRouting.ts      # Redux容器
└── __locales/
    ├── zh-cn.json               # 中文翻译
    └── en.json                  # 英文翻译
```

### 1.2 独立性确认
✅ 完全独立的状态管理（不复用filtering）
✅ 独立的actions和reducer
✅ 独立的组件实现
✅ 独立的翻译键

## 2. 数据结构分析

### 2.1 前端期望的API数据结构

#### 规则列表数据（从 getFilteringStatus 获取）

```typescript
// 前端期望从 GET /control/filtering/status 返回
{
  dns_routing_filters: [
    {
      id: number,
      url: string,
      name: string,
      enabled: boolean,
      upstream_group: string,      // 新增字段
      update_interval: number,      // 新增字段（分钟）
      priority: number,             // 新增字段
      rules_count: number,
      last_updated: string,         // ISO时间戳
    }
  ]
}
```

#### 添加规则请求（POST /control/filtering/add_url）
```typescript
{
  url: string,
  name: string,
  whitelist: false,              // 固定为false
  dns_routing: true,             // 标识为DNS路由规则
  upstream_group: string,        // 新增字段
  update_interval: number,       // 新增字段
  priority: number,              // 新增字段
}
```

#### 编辑规则请求（POST /control/filtering/set_url）
```typescript
{
  url: string,                   // 原URL（作为标识）
  data: {
    name: string,
    url: string,
    whitelist: false,
    dns_routing: true,
    upstream_group: string,
    update_interval: number,
    priority: number,
  }
}
```

#### 删除规则请求（POST /control/filtering/remove_url）
```typescript
{
  url: string,
  whitelist: false,
  dns_routing: true,
}
```

#### 刷新规则请求（POST /control/filtering/refresh）
```typescript
{
  whitelist: false,
  dns_routing: true,
  url?: string,  // 可选，指定刷新单个规则
}
```

### 2.2 自定义域名规则数据结构

#### 存储在 DNS Config 中
```typescript
// GET /control/dns_info 返回
// POST /control/dns_config 提交
{
  custom_domain_rules: [
    {
      domain: string,              // 域名
      matchType: 'DOMAIN' | 'DOMAIN-SUFFIX' | 'DOMAIN-KEYWORD',
      upstreamGroup: string,       // 上游组ID
      enabled: boolean,
    }
  ]
}
```


## 3. 现有API客户端分析

### 3.1 当前使用的API方法

前端通过 `client/src/api/Api.ts` 调用以下方法：

1. **getFilteringStatus()** - GET /control/filtering/status
   - 获取所有过滤规则（包括DNS路由规则）
   - 前端从返回的 `dns_routing_filters` 字段获取DNS路由规则

2. **addFilter(config)** - POST /control/filtering/add_url
   - 添加新规则
   - 前端传递 `dns_routing: true` 标识

3. **setFilterUrl(config)** - POST /control/filtering/set_url
   - 编辑现有规则
   - 前端传递 `dns_routing: true` 标识

4. **removeFilter(config)** - POST /control/filtering/remove_url
   - 删除规则
   - 前端传递 `dns_routing: true` 标识

5. **refreshFilters(config)** - POST /control/filtering/refresh
   - 刷新规则（单个或全部）
   - 前端传递 `dns_routing: true` 标识

### 3.2 DNS Config API

1. **getDnsConfig()** - GET /control/dns_info
   - 获取DNS配置（包括自定义域名规则）

2. **setDnsConfig(config)** - POST /control/dns_config
   - 更新DNS配置（包括自定义域名规则）

## 4. 后端需要实现的功能

### 4.1 扩展现有 filtering API

#### 4.1.1 扩展 filterAddJSON 结构
```go
type filterAddJSON struct {
    Name           string `json:"name"`
    URL            string `json:"url"`
    Whitelist      bool   `json:"whitelist"`
    DnsRouting     bool   `json:"dns_routing"`      // 新增
    UpstreamGroup  string `json:"upstream_group"`   // 新增
    UpdateInterval int    `json:"update_interval"`  // 新增（分钟）
    Priority       int    `json:"priority"`         // 新增
}
```

#### 4.1.2 扩展 FilterYAML 结构
```go
type FilterYAML struct {
    Enabled        bool   `yaml:"enabled"`
    URL            string `yaml:"url"`
    Name           string `yaml:"name"`
    RulesCount     int    `yaml:"-"`
    LastUpdated    string `yaml:"-"`
    ID             int64  `yaml:"-"`
    white          bool
    DnsRouting     bool   `yaml:"dns_routing"`      // 新增
    UpstreamGroup  string `yaml:"upstream_group"`   // 新增
    UpdateInterval int    `yaml:"update_interval"`  // 新增
    Priority       int    `yaml:"priority"`         // 新增
    Filter
}
```


#### 4.1.3 修改 handleFilteringStatus
需要在返回的JSON中添加 `dns_routing_filters` 字段：
```go
// 从所有filters中筛选出 dns_routing=true 的规则
dnsRoutingFilters := []FilterJSON{}
for _, filter := range filters {
    if filter.DnsRouting {
        dnsRoutingFilters = append(dnsRoutingFilters, convertToFilterJSON(filter))
    }
}

response := map[string]interface{}{
    "filters": filters,
    "whitelist_filters": whitelistFilters,
    "dns_routing_filters": dnsRoutingFilters,  // 新增
    // ... 其他字段
}
```

#### 4.1.4 修改 handleFilteringAddURL
- 读取并保存 `dns_routing`, `upstream_group`, `update_interval`, `priority` 字段
- 如果 `dns_routing=true`，不加载到普通过滤引擎
- 将规则存储到独立的DNS路由规则列表

#### 4.1.5 修改 handleFilteringSetURL
- 支持更新DNS路由相关字段
- 保持 `dns_routing` 标识不变

#### 4.1.6 修改 handleFilteringRemoveURL
- 支持通过 `dns_routing=true` 删除DNS路由规则

#### 4.1.7 修改 handleFilteringRefresh
- 支持刷新DNS路由规则
- 可以刷新单个或全部DNS路由规则

### 4.2 扩展 DNS Config API

#### 4.2.1 在 dnsConfig 结构中添加字段
```go
type dnsConfig struct {
    // ... 现有字段
    CustomDomainRules []CustomDomainRule `yaml:"custom_domain_rules"`
}

type CustomDomainRule struct {
    Domain        string `yaml:"domain" json:"domain"`
    MatchType     string `yaml:"match_type" json:"matchType"`
    UpstreamGroup string `yaml:"upstream_group" json:"upstreamGroup"`
    Enabled       bool   `yaml:"enabled" json:"enabled"`
}
```

#### 4.2.2 修改 handleGetConfig (GET /control/dns_info)
返回 `custom_domain_rules` 字段

#### 4.2.3 修改 handleSetConfig (POST /control/dns_config)
接收并保存 `custom_domain_rules` 字段

## 5. 前端代码问题和需要调整的地方

### 5.1 ✅ 已解决的问题
1. 独立性 - 完全独立实现，不复用filtering组件
2. 样式一致性 - 使用相同的CSS类名
3. 翻译完整性 - 所有文本都有翻译键
4. Modal显示 - 使用ReactModal正确实现

### 5.2 ⚠️ 需要注意的地方

#### 5.2.1 API响应字段映射
前端期望的字段名与后端可能不一致：
- 前端: `upstream_group` → 后端需要确认字段名
- 前端: `update_interval` → 后端需要确认字段名
- 前端: `rules_count` → 后端: `RulesCount`
- 前端: `last_updated` → 后端: `LastUpdated`

#### 5.2.2 时间格式
前端期望 `last_updated` 为ISO格式字符串，后端需要确保格式正确。


#### 5.2.3 上游组验证
前端在表单中只显示 `enabled=true` 的上游组：
```typescript
upstreamGroups.filter((group: any) => group.enabled)
```
后端需要确保返回的上游组列表包含 `enabled` 字段。

#### 5.2.4 错误处理
前端使用 toast 显示错误，后端需要返回清晰的错误消息。

## 6. 对接清单

### 6.1 后端必须实现的功能

- [ ] 扩展 `filterAddJSON` 结构，添加DNS路由字段
- [ ] 扩展 `FilterYAML` 结构，添加DNS路由字段
- [ ] 修改 `handleFilteringStatus`，返回 `dns_routing_filters`
- [ ] 修改 `handleFilteringAddURL`，支持DNS路由规则
- [ ] 修改 `handleFilteringSetURL`，支持更新DNS路由字段
- [ ] 修改 `handleFilteringRemoveURL`，支持删除DNS路由规则
- [ ] 修改 `handleFilteringRefresh`，支持刷新DNS路由规则
- [ ] 扩展 `dnsConfig` 结构，添加 `custom_domain_rules`
- [ ] 修改 `handleGetConfig`，返回自定义域名规则
- [ ] 修改 `handleSetConfig`，保存自定义域名规则
- [ ] 实现DNS路由规则的存储和加载（YAML配置）
- [ ] 实现DNS路由规则的定时更新机制
- [ ] 实现DNS路由规则的优先级排序

### 6.2 前端可能需要调整的地方

- [ ] 确认后端返回的字段名是否与前端期望一致
- [ ] 如果字段名不一致，添加字段映射逻辑
- [ ] 测试所有CRUD操作
- [ ] 测试自定义规则的增删改查
- [ ] 测试规则刷新功能
- [ ] 测试优先级排序显示
- [ ] 测试时间格式显示

### 6.3 集成测试项

1. **规则列表加载**
   - 页面加载时正确获取DNS路由规则
   - 规则按优先级排序显示
   - 显示正确的上游组名称

2. **添加规则**
   - 填写表单并提交
   - 验证必填字段
   - 验证URL格式
   - 验证上游组选择
   - 成功后刷新列表

3. **编辑规则**
   - 点击编辑按钮
   - 表单预填充现有数据
   - 修改后提交
   - 成功后刷新列表

4. **删除规则**
   - 点击删除按钮
   - 显示确认对话框
   - 确认后删除
   - 成功后刷新列表

5. **启用/禁用规则**
   - 点击开关
   - 立即生效
   - 刷新列表

6. **刷新规则**
   - 刷新单个规则
   - 刷新所有规则
   - 显示加载状态
   - 更新规则数量和时间

7. **自定义域名规则**
   - 添加自定义规则
   - 编辑自定义规则
   - 删除自定义规则
   - 启用/禁用自定义规则
   - 规则保存到DNS配置

## 7. API调用流程图

### 7.1 页面初始化
```
用户访问 /dns_routing
  ↓
DnsRouting.useEffect()
  ↓
getDnsRoutingFilters() → GET /control/filtering/status
  ↓
getDnsConfig() → GET /control/dns_info
  ↓
渲染规则列表和自定义规则列表
```

### 7.2 添加规则
```
用户点击"添加规则"
  ↓
打开 DnsRoutingModal
  ↓
用户填写表单并提交
  ↓
addDnsRoutingFilter() → POST /control/filtering/add_url
  ↓
成功后: toggleDnsRoutingModal() + addSuccessToast() + getDnsRoutingFilters()
```

### 7.3 添加自定义规则
```
用户点击"添加自定义规则"
  ↓
打开 CustomRuleModal
  ↓
用户填写表单并提交
  ↓
handleCustomRuleSubmit() → setDnsConfig() → POST /control/dns_config
  ↓
成功后: getDnsConfig() + addSuccessToast()
```


## 8. 关键代码片段

### 8.1 前端发送的添加规则请求
```typescript
// client/src/actions/dnsRouting.ts
export const addDnsRoutingFilter = (
    url: string,
    name: string,
    upstreamGroup: string,
    updateInterval: number = 0,
    priority: number = 0
) => async (dispatch: any) => {
    dispatch(addDnsRoutingFilterRequest());
    try {
        await apiClient.addFilter({
            url,
            name,
            whitelist: false,
            dns_routing: true,           // 关键标识
            upstream_group: upstreamGroup,
            update_interval: updateInterval,
            priority,
        });
        // ... 成功处理
    } catch (error) {
        // ... 错误处理
    }
};
```

### 8.2 前端获取规则列表
```typescript
// client/src/actions/dnsRouting.ts
export const getDnsRoutingFilters = () => async (dispatch: any) => {
    dispatch(getDnsRoutingFiltersRequest());
    try {
        const status = await apiClient.getFilteringStatus();
        // 期望从 status.dns_routing_filters 获取
        dispatch(getDnsRoutingFiltersSuccess(status.dns_routing_filters || []));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(getDnsRoutingFiltersFailure());
    }
};
```

### 8.3 前端自定义规则管理
```typescript
// client/src/hooks/useDnsRoutingCustomRules.ts
const handleCustomRuleSubmit = useCallback(async (rule: CustomRule) => {
    try {
        let updatedRules: CustomRule[];
        
        if (editingRule) {
            updatedRules = customRules.map(r => 
                r === editingRule ? rule : r
            );
        } else {
            updatedRules = [...customRules, rule];
        }
        
        const newConfig: DnsConfig = {
            ...dnsConfig,
            custom_domain_rules: updatedRules,  // 关键字段
        };
        
        await setDnsConfig(newConfig);  // POST /control/dns_config
        await getDnsConfig();           // GET /control/dns_info
        
        addSuccessToast(t('custom_rule_saved'));
    } catch (error) {
        addErrorToast({ error });
    }
}, [customRules, editingRule, dnsConfig, setDnsConfig, getDnsConfig]);
```

## 9. 后端实现建议

### 9.1 数据存储建议

#### 9.1.1 YAML配置文件结构
```yaml
dns:
  upstream_groups:
    - id: "group1"
      name: "国内DNS"
      enabled: true
      # ...
  
  # DNS路由规则（从URL加载）
  dns_routing_filters:
    - id: 1
      enabled: true
      url: "https://example.com/china-domains.txt"
      name: "中国域名列表"
      upstream_group: "group1"
      update_interval: 1440  # 24小时
      priority: 10
      rules_count: 5000
      last_updated: "2024-12-05T10:30:00Z"
  
  # 自定义域名规则
  custom_domain_rules:
    - domain: "example.com"
      match_type: "DOMAIN"
      upstream_group: "group1"
      enabled: true
    - domain: "google.com"
      match_type: "DOMAIN-SUFFIX"
      upstream_group: "group2"
      enabled: true
```

### 9.2 规则加载和匹配逻辑

1. **启动时加载**
   - 从YAML加载所有DNS路由规则配置
   - 从URL下载规则文件内容
   - 解析规则文件（支持多种格式）
   - 构建域名匹配树

2. **定时更新**
   - 根据 `update_interval` 设置定时器
   - 重新下载规则文件
   - 更新 `rules_count` 和 `last_updated`
   - 重新构建匹配树

3. **优先级匹配**
   - 按 `priority` 从小到大排序
   - 自定义规则优先级最高
   - 依次匹配每个规则
   - 返回第一个匹配的上游组

### 9.3 与现有filtering系统的区别

| 特性 | Filtering (黑白名单) | DNS Routing |
|------|---------------------|-------------|
| 用途 | 过滤/阻止域名 | 路由到不同上游 |
| 加载引擎 | urlfilter引擎 | 独立匹配引擎 |
| 存储位置 | filters数组 | dns_routing_filters数组 |
| 标识字段 | whitelist: true/false | dns_routing: true |
| 额外字段 | 无 | upstream_group, priority, update_interval |
| 匹配结果 | 阻止/允许 | 返回上游组ID |

## 10. 总结

### 10.1 前端代码质量评估
✅ **优秀**
- 完全独立的实现
- 清晰的代码结构
- 完整的类型定义
- 良好的错误处理
- 完整的国际化支持

### 10.2 准备就绪程度
前端代码已经完全准备好对接后端，只需要后端实现相应的API即可。

### 10.3 下一步行动
1. 后端实现扩展的filtering API
2. 后端实现DNS Config扩展
3. 后端实现DNS路由匹配引擎
4. 集成测试
5. 性能测试
6. 文档更新

---

**审查完成时间**: 2024-12-05
**审查人**: Kiro AI
**状态**: ✅ 前端代码审查通过，可以开始后端对接
