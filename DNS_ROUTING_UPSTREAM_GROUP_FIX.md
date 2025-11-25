# DNS路由规则上游组修复

## 问题描述

1. **添加域名规则时选择上游组无效，修改也无效**
   - 前端在提交表单时没有正确传递 `upstreamGroup` 参数
   - 后端API没有正确接收和保存 `upstream_group` 字段

2. **域名规则和白名单独立问题**
   - DNS路由规则使用的是 `whitelistFilters`，导致它们混在一起显示

## 修复内容

### 前端修复

#### 1. `client/src/components/Filters/DnsRouting.tsx`
- 修改 `handleSubmit` 方法，确保在添加和编辑时正确传递 `upstreamGroup` 参数
- 将 `upstreamGroup` 包含在 `filterData` 对象中

#### 2. `client/src/actions/filtering.ts`
- 修改 `addFilter` action，添加 `upstreamGroup` 可选参数
- 在调用API时将 `upstreamGroup` 转换为 `upstream_group` 字段
- 修改 `editFilter` action，确保更新时也能正确处理 `upstreamGroup`

### 后端修复

#### 1. `internal/filtering/http.go`
- 修改 `filterAddJSON` 结构体，添加 `UpstreamGroup` 字段
- 在 `handleFilteringAddURL` 中正确设置 `Filter.UpstreamGroup`
- `filterURLReqData` 已经有 `UpstreamGroup` 字段，确保在 `handleFilteringSetURL` 中正确使用

#### 2. `internal/filtering/filter.go`
- 在 `filterSetProperties` 函数中添加对 `UpstreamGroup` 的更新逻辑
- 当 `UpstreamGroup` 发生变化时，标记需要重启过滤引擎

#### 3. `internal/filtering/filtering.go`
- `Filter` 结构体已经有 `UpstreamGroup` 字段（带 yaml 标签）
- `filterToJSON` 函数已经正确地将 `UpstreamGroup` 包含在响应中

## 测试步骤

### 1. 测试添加DNS路由规则
1. 启动 AdGuard Home
2. 进入"DNS路由"页面
3. 点击"添加过滤器"
4. 填写名称和URL
5. **选择一个上游组**
6. 保存
7. 验证：
   - 规则成功添加
   - 在表格中显示正确的上游组名称
   - 刷新页面后上游组信息仍然保留

### 2. 测试编辑DNS路由规则
1. 点击已有规则的"编辑"按钮
2. 修改上游组选择
3. 保存
4. 验证：
   - 上游组更新成功
   - 表格中显示新的上游组名称
   - 刷新页面后更改仍然保留

### 3. 测试自定义域名规则
1. 在"自定义规则"部分添加新规则
2. 选择域名、匹配类型和上游组
3. 保存
4. 验证：
   - 规则成功添加
   - 上游组正确显示
   - 规则能够正常工作

### 4. 验证配置文件
检查 `AdGuardHome.yaml` 文件中的 `whitelist_filters` 部分：
```yaml
whitelist_filters:
  - enabled: true
    url: https://example.com/rules.txt
    name: Test Routing Rule
    id: 1234567890
    upstream_group: "group-id-here"
```

## 技术细节

### 数据流
1. 用户在前端选择上游组 → `upstreamGroup` (前端状态)
2. 提交表单 → `upstream_group` (API请求)
3. 后端接收 → `Filter.UpstreamGroup` (Go结构体)
4. 保存到配置 → `upstream_group` (YAML字段)
5. 返回给前端 → `upstream_group` (JSON响应)
6. 前端显示 → 查找对应的上游组名称

### 关键字段映射
- 前端表单：`upstreamGroup`
- API请求/响应：`upstream_group`
- Go结构体：`UpstreamGroup`
- YAML配置：`upstream_group`

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

## 注意事项

1. **DNS路由规则使用白名单过滤器**
   - DNS路由规则存储在 `whitelist_filters` 中
   - 通过 `upstream_group` 字段区分普通白名单和DNS路由规则
   - 有 `upstream_group` 的白名单过滤器会被用于DNS路由

2. **上游组必须存在且启用**
   - 前端只显示已启用的上游组
   - 如果选择的上游组被删除或禁用，规则可能无法正常工作

3. **配置文件兼容性**
   - 旧的配置文件（没有 `upstream_group` 字段）仍然兼容
   - 新添加的字段使用 `omitempty` 标签，不会影响现有配置

## 相关文件

- `client/src/components/Filters/DnsRouting.tsx`
- `client/src/components/Filters/Form.tsx`
- `client/src/components/Filters/Table.tsx`
- `client/src/actions/filtering.ts`
- `internal/filtering/http.go`
- `internal/filtering/filter.go`
- `internal/filtering/filtering.go`
