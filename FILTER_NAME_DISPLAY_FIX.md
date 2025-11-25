# 查询日志过滤器名称显示修复

## 问题描述

在查询日志的详情弹窗中，"规则"字段显示的是过滤器ID（例如：1764006015），而不是过滤器名称。这使得用户难以识别是哪个过滤器匹配了该规则。

## 修复方案

### 后端修改

1. **在 `internal/querylog/querylog.go` 中添加 GetFilterName 回调**
   - 在 `Config` 结构体中添加 `GetFilterName func(id int64) (name string)` 字段
   - 在 `queryLog` 结构体中添加 `getFilterName` 字段
   - 在 `newQueryLog` 函数中初始化 `getFilterName`

2. **在 `internal/filtering/filtering.go` 中添加 GetFilterName 方法**
   - 添加 `GetFilterName(filterID int64) (name string)` 方法
   - 该方法在所有过滤器列表（Filters、WhitelistFilters、DnsRoutingFilters）中查找匹配的过滤器ID
   - 返回对应的过滤器名称

3. **在 `internal/querylog/json.go` 中添加过滤器名称字段**
   - 在 `entryToJSON` 函数中，为每个规则添加 `filter_name` 字段
   - 添加新方法 `resultRulesToJSONRulesWithNames`，为 rules 数组中的每个规则添加 `filter_name`

4. **在 `internal/home/dns.go` 中连接回调**
   - 将 filters 初始化移到 querylog 之前
   - 在创建 querylog 配置时，设置 `GetFilterName: globalContext.filters.GetFilterName`

### 前端修改

1. **在 `client/src/helpers/helpers.tsx` 中修改类型定义**
   - 在 `Rule` 类型中添加可选字段 `filter_name?: string`

2. **修改 getFilterName 函数**
   - 添加可选参数 `filterName?: string`
   - 如果后端提供了 `filter_name`，优先使用它
   - 否则继续使用原有的查找逻辑

3. **更新相关函数**
   - 修改 `getFilterNames` 函数，传递 `filter_name` 参数
   - 修改 `getFilterNameToRulesMap` 函数，传递 `filter_name` 参数

## 实现细节

### 后端 API 响应格式

查询日志 API (`/control/querylog`) 现在返回的每个规则对象包含：

```json
{
  "filter_list_id": 1764006015,
  "text": "||example.com^",
  "filter_name": "我的自定义过滤器"
}
```

### 前端处理逻辑

前端在显示过滤器名称时：
1. 首先检查规则对象中是否有 `filter_name` 字段
2. 如果有，直接使用该名称
3. 如果没有，使用原有逻辑在本地过滤器列表中查找

## 优势

1. **性能优化**：后端直接提供过滤器名称，减少前端查找开销
2. **准确性提升**：避免前端过滤器列表与后端不同步导致的显示错误
3. **向后兼容**：保留原有的查找逻辑作为后备方案

## 测试步骤

1. 启动 AdGuardHome
2. 添加一个自定义过滤器（例如：Clash规则或DNS路由规则）
3. 进行DNS查询，触发该过滤器的规则
4. 打开查询日志，点击查询详情
5. 验证"规则"字段显示的是过滤器名称而不是ID

## 相关文件

### 后端
- `internal/querylog/querylog.go`
- `internal/querylog/qlog.go`
- `internal/querylog/json.go`
- `internal/filtering/filtering.go`
- `internal/home/dns.go`

### 前端
- `client/src/helpers/helpers.tsx`
- `client/src/components/Logs/Cells/ResponseCell.tsx`

## 完成状态

✅ 后端实现完成
✅ 前端实现完成
✅ 编译测试通过
⏳ 等待用户验证
