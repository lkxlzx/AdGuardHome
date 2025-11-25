# DNS路由规则与白名单完全分离

## 问题描述

之前DNS路由规则和白名单规则共用 `whitelist_filters` 字段，导致：
1. DNS路由规则出现在白名单页面
2. 白名单规则出现在DNS路由页面
3. 两种规则混在一起，难以管理

## 解决方案

创建独立的 `dns_routing_filters` 字段来存储DNS路由规则，完全与白名单分离。

## 修改内容

### 后端修改

#### 1. `internal/filtering/filtering.go`
- 在 `Config` 结构体中添加 `DnsRoutingFilters []FilterYAML` 字段
- 独立存储DNS路由过滤器列表

#### 2. `internal/filtering/http.go`
- 修改 `filteringConfig` 结构体，添加 `DnsRoutingFilters` 字段
- 修改 `filterAddJSON` 结构体，添加 `DnsRouting` 标志
- 修改 `filterURLReq` 结构体，添加 `DnsRouting` 标志
- 修改 `handleFilteringStatus`，返回 `dns_routing_filters`
- 修改 `handleFilteringAddURL`，根据 `dns_routing` 标志添加到正确的列表
- 修改 `handleFilteringRemoveURL`，支持删除DNS路由过滤器
- 修改 `handleFilteringSetURL`，支持更新DNS路由过滤器

#### 3. `internal/filtering/filter.go`
- 添加 `filterAddDnsRouting` 函数，专门用于添加DNS路由过滤器
- 修改 `filterExistsLocked`，也检查DNS路由过滤器
- 修改 `filterSetProperties`，添加 `isDnsRouting` 参数
- 修改 `enableFiltersLocked`，加载DNS路由过滤器到 `allowFilters`

### 前端修改

#### 1. `client/src/reducers/filtering.ts`
- 在初始状态中添加 `dnsRoutingFilters: []`

#### 2. `client/src/components/Filters/DnsRouting.tsx`
- 修改 `DnsRoutingProps` 接口，添加 `dnsRoutingFilters` 字段
- 修改 `render` 方法，使用 `dnsRoutingFilters` 而不是 `whitelistFilters`
- 修改 `handleSubmit`，传递 `dnsRouting=true` 标志
- 修改 `handleDelete`，传递 `dnsRouting=true` 标志
- 修改 `toggleFilter`，传递 `dnsRouting=true` 标志
- 修改 `handleRefresh`，传递 `dns_routing=true` 标志

#### 3. `client/src/actions/filtering.ts`
- 修改 `addFilter`，添加 `dnsRouting` 参数
- 修改 `editFilter`，添加 `dnsRouting` 参数
- 修改 `removeFilter`，添加 `dnsRouting` 参数
- 修改 `toggleFilterStatus`，添加 `dnsRouting` 参数

## 配置文件结构

### 旧结构（问题）
```yaml
whitelist_filters:
  - enabled: true
    url: https://example.com/whitelist.txt
    name: 白名单规则
    id: 1
  - enabled: true
    url: https://example.com/routing.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"  # 混在一起
```

### 新结构（正确）
```yaml
whitelist_filters:
  - enabled: true
    url: https://example.com/whitelist.txt
    name: 白名单规则
    id: 1

dns_routing_filters:
  - enabled: true
    url: https://example.com/routing.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"
```

## API变化

### 添加过滤器
```json
POST /control/filtering/add_url
{
  "name": "规则名称",
  "url": "https://example.com/rules.txt",
  "whitelist": false,
  "dns_routing": true,  // 新增：标识为DNS路由规则
  "upstream_group": "group-id"
}
```

### 更新过滤器
```json
POST /control/filtering/set_url
{
  "url": "https://example.com/rules.txt",
  "whitelist": false,
  "dns_routing": true,  // 新增：标识为DNS路由规则
  "data": {
    "name": "新名称",
    "url": "https://example.com/rules.txt",
    "enabled": true,
    "upstream_group": "group-id"
  }
}
```

### 删除过滤器
```json
POST /control/filtering/remove_url
{
  "url": "https://example.com/rules.txt",
  "whitelist": false,
  "dns_routing": true  // 新增：标识为DNS路由规则
}
```

### 获取状态
```json
GET /control/filtering/status

Response:
{
  "filters": [...],
  "whitelist_filters": [...],
  "dns_routing_filters": [  // 新增：独立的DNS路由规则列表
    {
      "id": 1,
      "enabled": true,
      "url": "https://example.com/routing.txt",
      "name": "DNS路由规则",
      "upstream_group": "group-id",
      "rules_count": 1000,
      "last_updated": "2024-01-01T00:00:00Z"
    }
  ],
  "user_rules": [],
  "interval": 24,
  "enabled": true
}
```

## 数据流

### 添加DNS路由规则
1. 用户在"DNS路由"页面点击"添加过滤器"
2. 填写名称、URL、选择上游组
3. 前端调用 `addFilter(url, name, false, true, upstreamGroup)`
   - `whitelist=false`
   - `dnsRouting=true`
4. API请求：`POST /control/filtering/add_url` with `dns_routing=true`
5. 后端调用 `filterAddDnsRouting()`
6. 规则添加到 `Config.DnsRoutingFilters`
7. 保存到配置文件的 `dns_routing_filters` 字段

### 显示DNS路由规则
1. 前端调用 `getFilteringStatus()`
2. API请求：`GET /control/filtering/status`
3. 后端返回 `dns_routing_filters` 数组
4. 前端存储到 `filtering.dnsRoutingFilters`
5. DNS路由页面显示 `dnsRoutingFilters`
6. 白名单页面显示 `whitelistFilters`
7. 两者完全独立，不会混淆

## 测试步骤

### 1. 测试DNS路由规则独立性
1. 启动 AdGuard Home
2. 进入"白名单"页面，确认只显示白名单规则
3. 进入"DNS路由"页面，确认只显示DNS路由规则
4. 添加一个DNS路由规则
5. 验证：
   - 规则只出现在DNS路由页面
   - 不出现在白名单页面
   - 配置文件中规则在 `dns_routing_filters` 字段下

### 2. 测试白名单规则独立性
1. 进入"白名单"页面
2. 添加一个白名单规则
3. 验证：
   - 规则只出现在白名单页面
   - 不出现在DNS路由页面
   - 配置文件中规则在 `whitelist_filters` 字段下

### 3. 测试DNS路由规则功能
1. 添加DNS路由规则并选择上游组
2. 保存后刷新页面
3. 验证：
   - 规则正确显示
   - 上游组正确显示
   - 编辑功能正常
   - 删除功能正常
   - 启用/禁用功能正常

### 4. 测试配置文件
检查 `AdGuardHome.yaml`：
```yaml
whitelist_filters:
  - enabled: true
    url: https://example.com/whitelist.txt
    name: 白名单规则
    id: 1

dns_routing_filters:
  - enabled: true
    url: https://example.com/routing.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"
```

## 迁移指南

### 自动迁移（推荐）
如果你的配置文件中有带 `upstream_group` 的 `whitelist_filters`，需要手动迁移：

1. 停止 AdGuard Home
2. 编辑 `AdGuardHome.yaml`
3. 将带有 `upstream_group` 的规则从 `whitelist_filters` 移动到 `dns_routing_filters`
4. 启动 AdGuard Home

### 示例迁移
**迁移前：**
```yaml
whitelist_filters:
  - enabled: true
    url: https://example.com/whitelist.txt
    name: 白名单规则
    id: 1
  - enabled: true
    url: https://example.com/routing.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"
```

**迁移后：**
```yaml
whitelist_filters:
  - enabled: true
    url: https://example.com/whitelist.txt
    name: 白名单规则
    id: 1

dns_routing_filters:
  - enabled: true
    url: https://example.com/routing.txt
    name: DNS路由规则
    id: 2
    upstream_group: "group-id"
```

## 技术细节

### 过滤器类型标识
- `whitelist=false, dns_routing=false` → 黑名单过滤器 (`filters`)
- `whitelist=true, dns_routing=false` → 白名单过滤器 (`whitelist_filters`)
- `whitelist=false, dns_routing=true` → DNS路由过滤器 (`dns_routing_filters`)

### 过滤器加载
在 `enableFiltersLocked` 函数中：
1. 加载黑名单过滤器到 `filters`
2. 加载白名单过滤器到 `allowFilters`
3. 加载DNS路由过滤器到 `allowFilters`（带 `UpstreamGroup` 信息）

### 过滤器应用
DNS路由过滤器作为允许列表的一部分，但带有额外的 `UpstreamGroup` 信息，用于DNS查询路由决策。

## 编译和部署

```bash
# 编译前端
cd client
npm run build-prod

# 编译后端
cd ..
go build -o AdGuardHome.exe

# 运行
./AdGuardHome.exe
```

## 相关文件

### 后端
- `internal/filtering/filtering.go` - 配置结构
- `internal/filtering/http.go` - HTTP API处理
- `internal/filtering/filter.go` - 过滤器管理逻辑

### 前端
- `client/src/reducers/filtering.ts` - 状态管理
- `client/src/components/Filters/DnsRouting.tsx` - DNS路由页面
- `client/src/actions/filtering.ts` - API调用

## 注意事项

1. **向后兼容性**：旧的配置文件仍然可以工作，但建议迁移到新结构
2. **ID唯一性**：确保所有过滤器（黑名单、白名单、DNS路由）的ID全局唯一
3. **上游组验证**：DNS路由规则必须指定有效的上游组ID
4. **刷新功能**：DNS路由规则的刷新独立于白名单规则
