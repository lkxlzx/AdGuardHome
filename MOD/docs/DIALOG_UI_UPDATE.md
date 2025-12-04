# 对话框 UI 更新说明

## 更新概述

根据提供的截图，已对 DNS 上游分组的添加/编辑对话框进行了样式调整，确保与设计稿完全一致。

## 对话框设计对比

### 截图中的设计元素

1. **标题**: "编辑分组" 或 "添加DNS上游分组"
2. **分组名称**: 
   - 标签: "分组名称"
   - 输入框: 单行文本输入
   - 占位符: "例如：国内 DNS、国外 DNS"

3. **上游服务器列表**:
   - 标签: "上游服务器列表"
   - 文本域: 多行输入，6行高度
   - 占位符: "每行一个 DNS 服务器地址\n例如：\n8.8.8.8\n1.1.1.1"
   - 错误提示: "必填字段" (红色)
   - 帮助文本: "支持普通 DNS、DNS-over-HTTPS (DoH)、DNS-over-TLS (DoT) 和 DNS-over-QUIC (DoQ) 服务器" (灰色)

4. **复选框**:
   - "启用此分组" (默认选中)
   - "设为默认分组" (默认未选中)

5. **按钮**:
   - "取消" (灰色按钮，左侧)
   - "保存" (绿色按钮，右侧)

### 实现的变更

#### 1. 移除 Bootstrap DNS 字段 ✅
根据截图，对话框中没有 Bootstrap DNS 字段，已从代码中移除。

#### 2. 更新占位符文本 ✅
```typescript
placeholder={t('upstream_dns_placeholder')}
// 翻译: "每行一个 DNS 服务器地址\n例如：\n8.8.8.8\n1.1.1.1"
```

#### 3. 更新错误提示 ✅
使用统一的 "必填字段" 提示：
```typescript
<div className="form__message form__message--error">
    <Trans>form_error_required</Trans>
</div>
```

#### 4. 调整帮助文本位置 ✅
帮助文本显示在文本域下方：
```typescript
<div className="form__desc form__desc--top">
    <Trans>upstream_servers_help</Trans>
</div>
```

#### 5. 添加对话框样式类 ✅
```typescript
className="Modal__Bootstrap modal-dialog modal-dialog-centered upstream-group-modal"
```

#### 6. 更新 CSS 样式 ✅
```css
/* 文本域样式 */
.upstream-group-modal .form-control--textarea {
    font-family: monospace;
    font-size: 13px;
    line-height: 1.5;
    resize: vertical;
}

.upstream-group-modal .form-control--textarea:focus {
    border-color: #5bc0de;
    box-shadow: 0 0 0 0.2rem rgba(91, 192, 222, 0.25);
}

/* 帮助文本样式 */
.upstream-group-modal .form__desc--top {
    margin-top: 0.5rem;
    color: #6c757d;
}
```

## 对话框字段说明

### 分组名称
- **类型**: 单行文本输入
- **验证**: 必填，最大50字符
- **占位符**: "例如：国内 DNS、国外 DNS"

### 上游服务器列表
- **类型**: 多行文本输入（textarea）
- **行数**: 6行
- **验证**: 必填，至少一个服务器地址
- **占位符**: 包含示例的多行文本
- **样式**: 等宽字体（monospace）
- **聚焦效果**: 蓝色边框

### 启用此分组
- **类型**: 复选框
- **默认值**: 选中（true）
- **说明**: 控制分组是否启用

### 设为默认分组
- **类型**: 复选框
- **默认值**: 未选中（false）
- **说明**: 设置为默认分组后，其他分组的默认状态会被取消

## 按钮说明

### 取消按钮
- **样式**: `btn btn-secondary btn-standard`
- **位置**: 左侧
- **功能**: 关闭对话框，不保存更改

### 检测按钮（仅编辑模式）
- **样式**: `btn btn-primary btn-standard`
- **位置**: 中间
- **功能**: 测试分组中所有上游服务器的连通性
- **显示条件**: 仅在编辑现有分组时显示

### 保存按钮
- **样式**: `btn btn-success btn-standard`
- **位置**: 右侧
- **功能**: 保存分组配置
- **加载状态**: 显示"保存中"文本

## 表单验证

### 分组名称验证
- 必填验证
- 长度验证（最大50字符）
- 错误提示：显示具体的错误信息

### 上游服务器列表验证
- 必填验证
- 至少包含一个有效的服务器地址
- 错误提示："必填字段"

## 样式细节

### 输入框样式
- 边框: `1px solid #ced4da`
- 圆角: `4px`
- 内边距: `8px 12px`
- 聚焦边框: `#5bc0de`（蓝色）

### 文本域样式
- 字体: `monospace`
- 字体大小: `13px`
- 行高: `1.5`
- 可调整大小: 垂直方向

### 错误提示样式
- 颜色: `#cd201f`（红色）
- 字体大小: 继承
- 显示位置: 输入框下方

### 帮助文本样式
- 颜色: `#6c757d`（灰色）
- 字体大小: 继承
- 显示位置: 输入框下方

### 复选框样式
- 使用 AdGuard Home 原有的 `checkbox` 样式
- 选中颜色: `#67b279`（绿色）

## 国际化支持

### 新增翻译键
```json
{
  "upstream_dns_placeholder": "每行一个 DNS 服务器地址\n例如：\n8.8.8.8\n1.1.1.1",
  "form_error_required": "必填字段"
}
```

### 使用的翻译键
- `edit_group` - 编辑分组
- `add_upstream_group` - 添加DNS上游分组
- `group_name` - 分组名称
- `group_name_placeholder` - 例如：国内 DNS、国外 DNS
- `upstream_servers` - 上游服务器列表
- `upstream_servers_help` - 支持普通 DNS、DNS-over-HTTPS...
- `enable_group` - 启用此分组
- `set_as_default` - 设为默认分组
- `cancel_btn` - 取消
- `save_btn` - 保存
- `test` - 检测
- `testing` - 检测中
- `saving` - 保存中

## 测试建议

### 视觉测试
1. 对比截图，确保布局完全一致
2. 检查输入框和文本域的样式
3. 检查复选框的样式和位置
4. 检查按钮的样式和顺序

### 功能测试
1. 测试表单验证（必填、长度限制）
2. 测试占位符文本显示
3. 测试错误提示显示
4. 测试帮助文本显示
5. 测试复选框切换
6. 测试保存和取消功能
7. 测试检测功能（编辑模式）

### 响应式测试
1. 测试不同屏幕尺寸下的显示
2. 测试对话框的居中显示
3. 测试文本域的自适应高度

## 文件变更清单

### 修改的文件
1. `client/src/components/Settings/Dns/UpstreamGroups/GroupModal.tsx`
   - 移除 Bootstrap DNS 字段
   - 更新占位符文本
   - 更新错误提示
   - 调整帮助文本位置
   - 添加对话框样式类

2. `client/src/components/Settings/Dns/UpstreamGroups/UpstreamGroups.css`
   - 添加对话框样式
   - 添加文本域样式
   - 添加帮助文本样式

3. `client/src/components/Settings/Dns/UpstreamGroups/index.tsx`
   - 导入 CSS 文件

4. `client/src/__locales/zh-cn.json`
   - 添加 `upstream_dns_placeholder` 翻译

5. `client/src/__locales/en.json`
   - 添加 `upstream_dns_placeholder` 翻译

## 完成状态

- [x] 移除 Bootstrap DNS 字段
- [x] 更新占位符文本
- [x] 更新错误提示为"必填字段"
- [x] 调整帮助文本位置
- [x] 添加对话框样式类
- [x] 更新 CSS 样式
- [x] 添加国际化翻译
- [x] 导入 CSS 文件

## 总结

对话框 UI 已完全按照截图进行调整，包括：
- 字段布局和样式
- 占位符文本
- 错误提示
- 帮助文本
- 复选框样式
- 按钮样式和顺序

所有样式都遵循 AdGuard Home 原有的设计规范，确保与现有页面的一致性。
