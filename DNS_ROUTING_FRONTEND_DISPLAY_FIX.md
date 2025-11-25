# DNS路由规则前端显示修复

## 问题描述

- 配置文件 `AdGuardHome.yaml` 中已经有 `dns_routing_filters` 字段和规则
- 后端API返回的数据中包含 `dns_routing_filters`
- 但前端页面不显示这些规则

## 问题原因

前端的 `normalizeFilteringStatus` 函数在处理API返回的过滤器状态时，没有提取和处理 `dns_routing_filters` 字段，导致这些数据被丢弃了。

## 数据流分析

### 正常流程
1. 前端调用 `getFilteringStatus()` action
2. API请求：`GET /control/filtering/status`
3. 后端返回：
```json
{
  "filters": [...],
  "whitelist_filters": [...],
  "dns_routing_filters": [
    {
      "id": 1764080553,
      "enabled": true,
      "url": "https://...",
      "name": "CN",
      "upstream_group": "group_1763970331409",
      "rules_count": 3728
    }
  ],
  "user_rules": [],
  "interval": 24,
  "enabled": true
}
```
4. 前端调用 `normalizeFilteringStatus(status)` 处理数据
5. **问题**：旧代码只提取了 `filters` 和 `whitelist_filters`，忽略了 `dns_routing_filters`
6. Redux store 中没有 `dnsRoutingFilters` 数据
7. 前端组件无法显示规则

## 解决方案

修改 `normalizeFilteringStatus` 函数，添加对 `dns_routing_filters` 的处理。

**文件**：`client/src/helpers/helpers.tsx`

### 修改前
```typescript
export const normalizeFilteringStatus = (filteringStatus: any) => {
    const { enabled, filters, user_rules: userRules, interval, whitelist_filters } = filteringStatus;
    const newUserRules = Array.isArray(userRules) ? userRules.join('\n') : '';

    return {
        enabled,
        userRules: newUserRules,
        filters: normalizeFilters(filters),
        whitelistFilters: normalizeFilters(whitelist_filters),
        interval,
    };
};
```

### 修改后
```typescript
export const normalizeFilteringStatus = (filteringStatus: any) => {
    const { enabled, filters, user_rules: userRules, interval, whitelist_filters, dns_routing_filters } = filteringStatus;
    const newUserRules = Array.isArray(userRules) ? userRules.join('\n') : '';

    return {
        enabled,
        userRules: newUserRules,
        filters: normalizeFilters(filters),
        whitelistFilters: normalizeFilters(whitelist_filters),
        dnsRoutingFilters: normalizeFilters(dns_routing_filters),  // 新增
        interval,
    };
};
```

## 修复后的数据流

1. 前端调用 `getFilteringStatus()` action
2. API请求：`GET /control/filtering/status`
3. 后端返回包含 `dns_routing_filters` 的数据
4. `normalizeFilteringStatus()` 提取 `dns_routing_filters`
5. 调用 `normalizeFilters(dns_routing_filters)` 格式化数据
6. 返回对象包含 `dnsRoutingFilters` 字段
7. Redux store 更新，包含 `filtering.dnsRoutingFilters`
8. `DnsRouting` 组件从 props 获取 `filtering.dnsRoutingFilters`
9. 规则正确显示在前端

## 相关代码

### 1. Action调用
**文件**：`client/src/components/Filters/DnsRouting.tsx`
```typescript
componentDidMount() {
    this.props.getFilteringStatus();  // 获取过滤器状态
    this.props.getDnsConfig();
}
```

### 2. Action定义
**文件**：`client/src/actions/filtering.ts`
```typescript
export const getFilteringStatus = () => async (dispatch: any) => {
    dispatch(getFilteringStatusRequest());
    try {
        const status = await apiClient.getFilteringStatus();
        dispatch(getFilteringStatusSuccess({ ...normalizeFilteringStatus(status) }));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(getFilteringStatusFailure());
    }
};
```

### 3. Reducer处理
**文件**：`client/src/reducers/filtering.ts`
```typescript
[actions.getFilteringStatusSuccess.toString()]: (state, { payload }: any) => ({
    ...state,
    ...payload,  // 包含 dnsRoutingFilters
    processingFilters: false,
}),
```

### 4. 组件使用
**文件**：`client/src/components/Filters/DnsRouting.tsx`
```typescript
render() {
    const {
        filtering: {
            dnsRoutingFilters,  // 从 Redux store 获取
            // ...
        },
    } = this.props;
    
    return (
        <Table
            filters={dnsRoutingFilters || []}  // 显示规则
            // ...
        />
    );
}
```

## 测试步骤

### 1. 测试规则显示
1. 确保配置文件中有 `dns_routing_filters`
2. 重启 AdGuard Home
3. 打开浏览器开发者工具 → Network 标签
4. 访问"DNS路由"页面
5. 查看 `/control/filtering/status` 请求的响应
6. 验证：响应中包含 `dns_routing_filters` 字段
7. 验证：前端页面正确显示规则

### 2. 测试数据格式化
1. 在浏览器控制台执行：
```javascript
// 查看 Redux store
window.__REDUX_DEVTOOLS_EXTENSION__ && console.log(store.getState().filtering)
```
2. 验证：`dnsRoutingFilters` 数组存在且包含规则
3. 验证：每个规则包含正确的字段（id, name, url, upstream_group等）

### 3. 测试规则操作
1. 点击规则的"编辑"按钮
2. 验证：弹出编辑对话框，显示正确的规则信息
3. 修改规则并保存
4. 验证：规则更新成功
5. 刷新页面
6. 验证：修改后的规则正确显示

## API响应示例

```json
{
  "filters": [
    {
      "id": 1,
      "enabled": true,
      "url": "https://adguardteam.github.io/HostlistsRegistry/assets/filter_1.txt",
      "name": "AdGuard DNS filter",
      "rules_count": 50000,
      "last_updated": "2024-01-01T00:00:00Z"
    }
  ],
  "whitelist_filters": [],
  "dns_routing_filters": [
    {
      "id": 1764080553,
      "enabled": true,
      "url": "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml",
      "name": "CN",
      "upstream_group": "group_1763970331409",
      "rules_count": 3728,
      "last_updated": "2024-11-24T23:55:17Z"
    }
  ],
  "user_rules": [],
  "interval": 24,
  "enabled": true
}
```

## Redux Store 结构

```javascript
{
  filtering: {
    enabled: true,
    interval: 24,
    filters: [
      { id: 1, name: "AdGuard DNS filter", ... }
    ],
    whitelistFilters: [],
    dnsRoutingFilters: [
      { id: 1764080553, name: "CN", upstream_group: "group_1763970331409", ... }
    ],
    userRules: "",
    // ... 其他状态
  }
}
```

## 注意事项

1. **数据格式化**：`normalizeFilters()` 函数会格式化过滤器数据，确保字段名称一致
2. **空数组处理**：如果 `dns_routing_filters` 为 `undefined` 或 `null`，`normalizeFilters()` 会返回空数组
3. **向后兼容**：旧版本的API响应中没有 `dns_routing_filters`，代码会正确处理这种情况

## 编译和部署

```bash
# 编译前端
cd client
npm run build-prod

# 重启 AdGuard Home
# 刷新浏览器页面
```

## 相关文件

- `client/src/helpers/helpers.tsx` - 数据格式化函数
- `client/src/actions/filtering.ts` - API调用和action定义
- `client/src/reducers/filtering.ts` - Redux reducer
- `client/src/components/Filters/DnsRouting.tsx` - DNS路由页面组件

## 调试技巧

### 1. 检查API响应
在浏览器开发者工具中：
1. Network 标签 → 找到 `/control/filtering/status` 请求
2. 查看 Response 标签
3. 确认 `dns_routing_filters` 字段存在

### 2. 检查Redux Store
在浏览器控制台中：
```javascript
// 如果安装了 Redux DevTools
// 查看当前状态
store.getState().filtering.dnsRoutingFilters

// 或者在组件中打印
console.log(this.props.filtering.dnsRoutingFilters)
```

### 3. 检查组件Props
在 `DnsRouting` 组件的 `render` 方法中添加：
```typescript
console.log('dnsRoutingFilters:', this.props.filtering.dnsRoutingFilters);
```

## 完整修复清单

- [x] 后端：添加 `DnsRoutingFilters` 字段到配置结构
- [x] 后端：配置保存时同步 `DnsRoutingFilters`
- [x] 后端：配置加载时同步 `DnsRoutingFilters`
- [x] 后端：过滤器初始化时加载 `DnsRoutingFilters`
- [x] 后端：API返回 `dns_routing_filters` 字段
- [x] 前端：`normalizeFilteringStatus` 处理 `dns_routing_filters`
- [x] 前端：Redux store 包含 `dnsRoutingFilters`
- [x] 前端：组件使用 `dnsRoutingFilters` 显示规则
