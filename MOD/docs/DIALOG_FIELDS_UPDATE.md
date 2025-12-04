# 对话框字段更新说明

## 更新概述

在添加/编辑 DNS 分组对话框中新增了"后备 DNS 服务器"和"Bootstrap DNS 服务器"两个文本域输入框。

## 对话框字段列表

### 1. 分组名称 (必填)
- **类型**: 单行文本输入
- **验证**: 必填，最大50字符
- **占位符**: "例如：国内 DNS、国外 DNS"

### 2. 上游服务器列表 (必填)
- **类型**: 多行文本输入（textarea）
- **行数**: 6行
- **验证**: 必填，至少一个服务器地址
- **占位符**: "每行一个 DNS 服务器地址\n例如：\n8.8.8.8\n1.1.1.1"
- **帮助文本**: "支持普通 DNS、DNS-over-HTTPS (DoH)、DNS-over-TLS (DoT) 和 DNS-over-QUIC (DoQ) 服务器"

### 3. 后备 DNS 服务器 (可选) ✨ 新增
- **类型**: 多行文本输入（textarea）
- **行数**: 3行
- **验证**: 可选
- **占位符**: "每行输入一个后备 DNS 服务器"
- **说明**: "当上游 DNS 服务器没有响应时使用的后备 DNS 服务器列表。语法与上面的「主要上游」字段相同。"

### 4. Bootstrap DNS 服务器 (可选) ✨ 新增
- **类型**: 多行文本输入（textarea）
- **行数**: 3行
- **验证**: 可选
- **占位符**: "Bootstrap DNS 服务器"
- **说明**: "DNS 服务器的 IP 地址，用于解析指定为上游的 DoH/DoT 解析器的 IP 地址。不允许添加注释。"

### 5. 启用此分组
- **类型**: 复选框
- **默认值**: 选中（true）

### 6. 设为默认分组
- **类型**: 复选框
- **默认值**: 未选中（false）

## 按钮

### 取消按钮
- **样式**: `btn btn-secondary btn-standard`
- **功能**: 关闭对话框，不保存更改

### 检测按钮（仅编辑模式）
- **样式**: `btn btn-primary btn-standard`
- **功能**: 测试分组中所有上游服务器的连通性
- **显示条件**: 仅在编辑现有分组时显示

### 保存按钮
- **样式**: `btn btn-success btn-standard`
- **功能**: 保存分组配置
- **加载状态**: 显示"保存中"文本

## 数据结构

### 前端 FormData
```typescript
interface FormData {
    name: string;              // 分组名称
    upstream_dns: string;      // 上游服务器列表（换行分隔）
    fallback_dns?: string;     // 后备 DNS 服务器（换行分隔）
    bootstrap_dns?: string;    // Bootstrap DNS 服务器（换行分隔）
    enabled: boolean;          // 是否启用
    is_default: boolean;       // 是否为默认分组
}
```

### 后端 UpstreamGroup
```typescript
interface UpstreamGroup {
    id: string;
    name: string;
    enabled: boolean;
    is_default: boolean;
    upstream_dns: string[];        // 上游 DNS 服务器列表
    fallback_dns?: string[];       // 后备 DNS 服务器列表（可选）
    bootstrap_dns?: string[];      // Bootstrap DNS 服务器列表（可选）
    created_at: string;
    updated_at: string;
}
```

## 数据处理

### 提交时的数据转换
```typescript
const onSubmit = (data: FormData) => {
    // 上游服务器列表（必填）
    const upstreamDns = data.upstream_dns
        .split('\n')
        .map((s) => s.trim())
        .filter((s) => s.length > 0);

    // 后备 DNS 服务器（可选）
    const fallbackDns = data.fallback_dns
        ? data.fallback_dns
              .split('\n')
              .map((s) => s.trim())
              .filter((s) => s.length > 0)
        : [];

    // Bootstrap DNS 服务器（可选）
    const bootstrapDns = data.bootstrap_dns
        ? data.bootstrap_dns
              .split('\n')
              .map((s) => s.trim())
              .filter((s) => s.length > 0)
        : [];

    const groupData = {
        name: data.name,
        upstream_dns: upstreamDns,
        fallback_dns: fallbackDns.length > 0 ? fallbackDns : undefined,
        bootstrap_dns: bootstrapDns.length > 0 ? bootstrapDns : undefined,
        enabled: data.enabled,
        is_default: data.is_default,
    };
};
```

## 样式

### 文本域样式
```css
.form-control--textarea {
    font-family: monospace;
    font-size: 13px;
    line-height: 1.5;
    resize: vertical;
}
```

### 帮助文本样式
```css
.form__desc--top {
    margin-top: 0.5rem;
    color: #6c757d;
}
```

## 国际化

### 使用的翻译键
- `fallback_dns_title` - 后备 DNS 服务器
- `fallback_dns_desc` - 后备 DNS 服务器说明
- `fallback_dns_placeholder` - 后备 DNS 服务器占位符
- `bootstrap_dns` - Bootstrap DNS 服务器
- `bootstrap_dns_desc` - Bootstrap DNS 服务器说明

## 文件变更

### 修改的文件
1. **`client/src/components/Settings/Dns/UpstreamGroups/GroupModal.tsx`**
   - 添加 `fallback_dns` 字段到 FormData 接口
   - 添加后备 DNS 服务器文本域
   - 添加 Bootstrap DNS 服务器文本域（之前已有但被移除，现在恢复）
   - 更新数据提交逻辑

2. **`client/src/types/upstreamGroups.ts`**
   - 添加 `fallback_dns?: string[]` 字段到 UpstreamGroup 接口

## 测试建议

### 功能测试
1. 打开添加分组对话框
2. 填写所有字段（包括后备 DNS 和 Bootstrap DNS）
3. 保存并验证数据正确提交
4. 编辑分组，验证字段正确预填充
5. 测试可选字段为空时的保存

### 验证测试
1. 测试必填字段验证（分组名称、上游服务器列表）
2. 测试可选字段可以为空
3. 测试多行输入的正确解析

### 样式测试
1. 验证文本域样式一致
2. 验证帮助文本显示正确
3. 验证占位符文本显示正确

## 后端 API 要求

后端需要支持以下字段：
- `upstream_dns` (必填) - 上游 DNS 服务器列表
- `fallback_dns` (可选) - 后备 DNS 服务器列表
- `bootstrap_dns` (可选) - Bootstrap DNS 服务器列表

### API 请求示例
```json
{
  "name": "国内",
  "upstream_dns": ["223.6.6.6", "119.29.29.29"],
  "fallback_dns": ["8.8.8.8", "1.1.1.1"],
  "bootstrap_dns": ["223.5.5.5"],
  "enabled": true,
  "is_default": true
}
```

## 总结

✅ **已完成**
- 添加后备 DNS 服务器字段
- 添加 Bootstrap DNS 服务器字段
- 更新数据类型定义
- 更新数据处理逻辑
- 保持与原有 UI 风格一致

🎯 **效果**
- 对话框现在包含完整的 DNS 配置选项
- 用户可以为每个分组单独配置后备 DNS 和 Bootstrap DNS
- 所有字段都有适当的验证和帮助文本
