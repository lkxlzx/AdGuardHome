# DNS路由规则上游组显示修复

## 问题描述

- 添加DNS路由规则时，上游组选择和保存正常
- 配置文件中 `upstream_group` 字段正确保存
- 但前端显示时：
  - 表格中的"DNS分组"栏显示为空或显示ID而不是名称
  - 编辑对话框中的上游组选择框没有预选中当前值

## 问题原因

前端有两个数据处理函数没有正确处理 `upstream_group` 字段：

### 1. `normalizeFilters` 函数
**位置**：`client/src/helpers/helpers.tsx`

**问题**：从API响应中提取过滤器数据时，没有提取 `upstream_group` 字段

**影响**：Redux store 中的过滤器对象缺少 `upstreamGroup` 属性，导致表格无法显示

### 2. `getCurrentFilter` 函数
**位置**：`client/src/helpers/helpers.tsx`

**问题**：获取当前编辑的过滤器数据时，没有包含 `upstreamGroup` 字段

**影响**：编辑对话框的 `initialValues` 中缺少 `upstreamGroup`，导致选择框无法预选中

## 数据流分析

### 正常流程（修复后）

#### 1. 数据获取
```
API响应
↓
{
  "dns_routing_filters": [
    {
      "id": 1764080553,
      "name": "CN",
      "url": "https://...",
      "upstream_group": "group_1763970331409",  ← 后端返回
      "enabled": true,
      "rules_count": 3728
    }
  ]
}
↓
normalizeFilters() 处理
↓
{
  id: 1764080553,
  name: "CN",
  url: "https://...",
  upstreamGroup: "group_1763970331409",  ← 提取并转换为驼峰命名
  enabled: true,
  rulesCount: 3728
}
↓
Redux Store
```

#### 2. 表格显示
```
Redux Store
↓
filtering.dnsRoutingFilters
↓
Table 组件
↓
upstreamGroup 字段
↓
查找对应的上游组名称
↓
显示在"DNS分组"栏
```

#### 3. 编辑对话框
```
点击编辑按钮
↓
getCurrentFilter(url, dnsRoutingFilters)
↓
返回包含 upstreamGroup 的对象
↓
作为 initialValues 传递给 Form
↓
上游组选择框预选中当前值
```

## 解决方案

### 修改1：`normalizeFilters` 函数

**文件**：`client/src/helpers/helpers.tsx`

#### 修改前
```typescript
export const normalizeFilters = (filters: any) =>
    filters
        ? filters.map((filter: any) => {
              const { id, url, enabled, last_updated, name = 'Default name', rules_count = 0 } = filter;

              return {
                  id,
                  url,
                  enabled,
                  lastUpdated: last_updated,
                  name,
                  rulesCount: rules_count,
              };
          })
        : [];
```

#### 修改后
```typescript
export const normalizeFilters = (filters: any) =>
    filters
        ? filters.map((filter: any) => {
              const { id, url, enabled, last_updated, name = 'Default name', rules_count = 0, upstream_group } = filter;

              return {
                  id,
                  url,
                  enabled,
                  lastUpdated: last_updated,
                  name,
                  rulesCount: rules_count,
                  upstreamGroup: upstream_group,  // 新增
              };
          })
        : [];
```

### 修改2：`getCurrentFilter` 函数

**文件**：`client/src/helpers/helpers.tsx`

#### 修改前
```typescript
export const getCurrentFilter = (url: any, filters: any) => {
    const filter = filters?.find((item: any) => url === item.url);

    if (filter) {
        const { enabled, name, url } = filter;
        return {
            enabled,
            name,
            url,
        };
    }

    return {
        name: '',
        url: '',
    };
};
```

#### 修改后
```typescript
export const getCurrentFilter = (url: any, filters: any) => {
    const filter = filters?.find((item: any) => url === item.url);

    if (filter) {
        const { enabled, name, url, upstreamGroup } = filter;
        return {
            enabled,
            name,
            url,
            upstreamGroup,  // 新增
        };
    }

    return {
        name: '',
        url: '',
        upstreamGroup: '',  // 新增
    };
};
```

## 字段命名转换

### API响应（后端）
- 使用下划线命名：`upstream_group`
- JSON格式：`"upstream_group": "group_1763970331409"`

### Redux Store（前端）
- 使用驼峰命名：`upstreamGroup`
- JavaScript对象：`upstreamGroup: "group_1763970331409"`

### 转换位置
- `normalizeFilters()` 函数负责转换
- 从 `upstream_group` → `upstreamGroup`

## 测试步骤

### 1. 测试表格显示
1. 重启 AdGuard Home
2. 访问"DNS路由"页面
3. 查看规则列表
4. 验证："DNS分组"栏显示上游组名称（如"国内DNS"）而不是ID

### 2. 测试编辑对话框
1. 点击某个规则的"编辑"按钮
2. 查看上游组选择框
3. 验证：当前规则的上游组已预选中
4. 修改上游组并保存
5. 验证：修改成功

### 3. 测试数据完整性
在浏览器控制台执行：
```javascript
// 查看Redux store中的过滤器数据
console.log(store.getState().filtering.dnsRoutingFilters);

// 应该看到每个过滤器都有 upstreamGroup 字段
// 例如：
// [
//   {
//     id: 1764080553,
//     name: "CN",
//     url: "https://...",
//     upstreamGroup: "group_1763970331409",
//     enabled: true,
//     rulesCount: 3728
//   }
// ]
```

### 4. 测试编辑初始值
在 `DnsRouting.tsx` 的 `render` 方法中添加调试：
```typescript
console.log('currentFilterData:', currentFilterData);

// 点击编辑按钮后，应该看到：
// {
//   enabled: true,
//   name: "CN",
//   url: "https://...",
//   upstreamGroup: "group_1763970331409"
// }
```

## 相关组件

### 1. Table 组件
**文件**：`client/src/components/Filters/Table.tsx`

显示"DNS分组"栏：
```typescript
{
    Header: <Trans>dns_group_table_header</Trans>,
    accessor: 'upstreamGroup',  // 使用 upstreamGroup 字段
    Cell: ({ value }: any) => {
        const group = upstreamGroups.find((g: any) => g.id === value);
        const displayName = group ? group.name : value;
        return <div className="logs__row">{displayName}</div>;
    },
}
```

### 2. Form 组件
**文件**：`client/src/components/Filters/Form.tsx`

上游组选择框：
```typescript
<Controller
    name="upstreamGroup"
    control={control}
    rules={{ required: t('form_error_required') }}
    render={({ field, fieldState }) => (
        <select
            {...field}  // field.value 来自 initialValues.upstreamGroup
            id="upstreamGroup"
            className="form-control">
            <option value="">{t('custom_rule_select_group')}</option>
            {upstreamGroups
                .filter((group: any) => group.enabled)
                .map((group: any) => (
                    <option key={group.id} value={group.id}>
                        {group.name}
                    </option>
                ))}
        </select>
    )}
/>
```

### 3. Modal 组件
**文件**：`client/src/components/Filters/Modal.tsx`

传递初始值：
```typescript
case MODAL_TYPE.EDIT_FILTERS:
    initialValues = currentFilterData;  // 包含 upstreamGroup
    break;
```

## API响应示例

```json
{
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
  ]
}
```

## Redux Store 结构

```javascript
{
  filtering: {
    dnsRoutingFilters: [
      {
        id: 1764080553,
        name: "CN",
        url: "https://...",
        upstreamGroup: "group_1763970331409",  // 驼峰命名
        enabled: true,
        rulesCount: 3728,
        lastUpdated: "2024-11-24T23:55:17Z"
      }
    ]
  }
}
```

## 编译和部署

```bash
# 编译前端
cd client
npm run build-prod

# 编译后端
cd ..
go build -o AdGuardHome.exe

# 重启 AdGuard Home
# 刷新浏览器页面（Ctrl+F5 强制刷新）
```

## 注意事项

1. **字段命名一致性**：
   - 后端使用 `upstream_group`（下划线）
   - 前端使用 `upstreamGroup`（驼峰）
   - 转换在 `normalizeFilters()` 中完成

2. **空值处理**：
   - 如果 `upstream_group` 为 `undefined`，`upstreamGroup` 也会是 `undefined`
   - 表格和表单都能正确处理空值

3. **向后兼容**：
   - 旧的过滤器（没有 `upstream_group`）不会报错
   - 只是 `upstreamGroup` 字段为 `undefined`

## 完整修复清单

- [x] 后端：正确返回 `upstream_group` 字段
- [x] 前端：`normalizeFilters` 提取 `upstream_group`
- [x] 前端：`getCurrentFilter` 包含 `upstreamGroup`
- [x] 前端：表格正确显示上游组名称
- [x] 前端：编辑对话框预选中当前上游组
- [x] 前端：保存时正确传递上游组

## 相关文件

- `client/src/helpers/helpers.tsx` - 数据格式化函数
- `client/src/components/Filters/Table.tsx` - 表格显示
- `client/src/components/Filters/Form.tsx` - 编辑表单
- `client/src/components/Filters/Modal.tsx` - 对话框
- `client/src/components/Filters/DnsRouting.tsx` - DNS路由页面
