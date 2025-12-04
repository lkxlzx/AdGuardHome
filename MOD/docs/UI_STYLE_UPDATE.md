# UI 风格更新说明

## 更新概述

根据 AdGuard Home 原有的 UI 设计风格，已对 DNS 上游分组功能的前端组件进行了全面重构，确保与现有页面（如 Rewrites、Clients 等）的风格完全一致。

## 主要变更

### 1. 表格组件重构 ✅

**之前**: 使用原生 HTML `<table>` 标签
**现在**: 使用 `react-table` 库（与 Rewrites 等页面一致）

**变更内容**:
- 删除了 `GroupRow.tsx` 组件
- 在 `GroupList.tsx` 中集成 react-table
- 实现了与 Rewrites 页面相同的表格样式和交互
- 支持分页、排序、加载状态等功能

**代码示例**:
```tsx
<ReactTable
    data={groups || []}
    columns={columns}
    loading={processing}
    className="-striped -highlight card-table-overflow"
    showPagination
    defaultPageSize={10}
    // ... 其他配置
/>
```

### 2. 对话框样式更新 ✅

**之前**: 使用自定义的 Modal 组件和 Bootstrap 样式类
**现在**: 使用 AdGuard Home 原有的表单样式类

**变更内容**:
- 使用 `form__group`、`form__label` 等原有表单样式类
- 使用 `checkbox`、`checkbox__input`、`checkbox__label` 等原有复选框样式
- 使用 `btn-list`、`btn-standard` 等原有按钮样式
- 使用 `form__desc` 显示帮助文本
- 使用 `form__message--error` 显示错误信息

**代码示例**:
```tsx
<div className="form__group">
    <label className="form__label" htmlFor="name">
        <Trans>group_name</Trans>
    </label>
    <input
        type="text"
        className="form-control"
        id="name"
    />
    <div className="form__desc">
        <Trans>group_name_help</Trans>
    </div>
</div>
```

### 3. 复选框样式更新 ✅

**之前**: 使用 Bootstrap 的 `custom-control-input` 样式
**现在**: 使用 AdGuard Home 原有的 `checkbox` 样式

**代码示例**:
```tsx
<label className="checkbox">
    <input
        type="checkbox"
        className="checkbox__input"
        checked={field.value}
    />
    <span className="checkbox__label">
        <Trans>enable_group</Trans>
    </span>
</label>
```

### 4. 按钮样式更新 ✅

**之前**: 使用 `btn btn-success mt-3` 等样式
**现在**: 使用 `btn btn-success btn-standard` 和 `card-actions` 容器

**代码示例**:
```tsx
<div className="card-actions">
    <button
        type="button"
        className="btn btn-success btn-standard"
        onClick={onAddGroup}
    >
        {t('add_upstream_group')}
    </button>
</div>
```

### 5. 表格操作按钮更新 ✅

**之前**: 使用自定义的 `btn-icon` 样式
**现在**: 使用 AdGuard Home 原有的 `btn-icon btn-outline-primary` 样式

**代码示例**:
```tsx
<button
    type="button"
    className="btn btn-icon btn-outline-primary btn-sm mr-2"
    onClick={handleEdit}
    title={t('edit')}
>
    <svg className="icons icon12">
        <use xlinkHref="#edit" />
    </svg>
</button>
```

### 6. 国际化更新 ✅

**变更内容**:
- 使用 `<Trans>` 组件替代 `{t()}` 函数（与现有页面一致）
- 添加了缺少的翻译键（如 `enabled_table_header`、`actions_table_header` 等）
- 更新了描述文本以支持 HTML 标签（如链接）

### 7. LocalStorage 支持 ✅

**变更内容**:
- 在 `localStorageHelper.ts` 中添加了 `UPSTREAM_GROUPS_PAGE_SIZE` 键
- 支持保存和恢复表格分页大小设置

### 8. CSS 样式简化 ✅

**变更内容**:
- 删除了大部分自定义样式
- 只保留了必要的样式（如 `badge-success`、`test-results-alert`）
- 完全依赖 AdGuard Home 原有的样式类

## 文件变更清单

### 修改的文件
1. `client/src/components/Settings/Dns/UpstreamGroups/GroupList.tsx` - 重构为使用 react-table
2. `client/src/components/Settings/Dns/UpstreamGroups/GroupModal.tsx` - 更新为使用原有样式类
3. `client/src/components/Settings/Dns/UpstreamGroups/index.tsx` - 更新导入和使用 Trans 组件
4. `client/src/components/Settings/Dns/UpstreamGroups/UpstreamGroups.css` - 简化样式
5. `client/src/helpers/localStorageHelper.ts` - 添加新的存储键
6. `client/src/__locales/zh-cn.json` - 添加缺少的翻译键
7. `client/src/__locales/en.json` - 添加缺少的翻译键

### 删除的文件
1. `client/src/components/Settings/Dns/UpstreamGroups/GroupRow.tsx` - 功能已整合到 GroupList 中

## 样式对比

### 表格样式
| 元素 | 之前 | 现在 |
|------|------|------|
| 表格容器 | `<table className="table table-hover">` | `<ReactTable className="-striped -highlight card-table-overflow">` |
| 复选框 | `<input type="checkbox">` | `<label className="checkbox"><input className="checkbox__input">` |
| 操作按钮 | `<button className="btn btn-icon">` | `<button className="btn btn-icon btn-outline-primary btn-sm">` |

### 对话框样式
| 元素 | 之前 | 现在 |
|------|------|------|
| 表单组 | `<div className="form-group">` | `<div className="form__group">` |
| 标签 | `<label>` | `<label className="form__label">` |
| 帮助文本 | `<small className="form-text text-muted">` | `<div className="form__desc">` |
| 错误信息 | `<div className="invalid-feedback">` | `<div className="form__message form__message--error">` |
| 按钮容器 | `<div className="modal-footer">` | `<div className="modal-footer"><div className="btn-list">` |

## 验证清单

- [x] 表格使用 react-table
- [x] 表格支持分页和排序
- [x] 表格支持加载状态
- [x] 复选框使用原有样式
- [x] 表单使用原有样式类
- [x] 按钮使用原有样式类
- [x] 对话框使用原有样式类
- [x] 国际化使用 Trans 组件
- [x] 添加了所有缺少的翻译键
- [x] LocalStorage 支持分页大小保存
- [x] 删除了不必要的自定义样式
- [x] 完全匹配 AdGuard Home 原有风格

## 测试建议

1. **视觉测试**: 对比 Rewrites 页面和上游分组页面，确保样式一致
2. **交互测试**: 测试表格的分页、排序、复选框等交互
3. **对话框测试**: 测试创建和编辑对话框的表单验证和提交
4. **响应式测试**: 测试不同屏幕尺寸下的显示效果
5. **国际化测试**: 切换语言，确保所有文本正确显示

## 后续工作

前端 UI 已完全匹配 AdGuard Home 原有风格，现在可以：
1. 开始后端 API 实现
2. 进行前后端集成测试
3. 编写单元测试和 E2E 测试
4. 进行性能优化

## 参考页面

在实现过程中参考了以下现有页面：
- `client/src/components/Filters/Rewrites/` - 表格和对话框样式
- `client/src/components/Settings/Clients/` - 表单样式
- `client/src/components/Settings/Dhcp/` - 复选框样式

所有样式类和组件使用方式都与这些页面保持一致。
