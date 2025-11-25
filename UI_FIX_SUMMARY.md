# UI 修复总结

## 修复内容

### 域名分流页面文案修改

**问题**：域名分流页面使用了与白名单相同的文案，导致混淆。

**修复**：
1. 添加了专门用于域名分流的翻译键
2. 修改了相关组件以支持区分显示

---

## 修改的文件

### 1. 翻译文件 (`client/src/__locales/zh-cn.json`)

添加了新的翻译键：

```json
"dns_routing_rules": "域名分流规则",
"add_routing_rule": "添加规则",
"new_routing_rule": "新增规则",
"edit_routing_rule": "编辑规则",
"no_routing_rule_added": "未添加规则",
"enter_valid_routing_rule_url": "输入有效的规则 URL",
"routing_rule_name": "规则名称",
"routing_rule_url": "规则 URL",
"routing_rule_group": "目标上游组",
"routing_rule_count": "规则数量"
```

### 2. Actions 组件 (`client/src/components/Filters/Actions.tsx`)

**修改**：
- 添加 `isRoutingRule` 参数
- 根据 `isRoutingRule` 显示不同的按钮文本

**修改前**：
```tsx
{whitelist ? <Trans>add_allowlist</Trans> : <Trans>add_blocklist</Trans>}
```

**修改后**：
```tsx
{isRoutingRule ? <Trans>add_routing_rule</Trans> : (whitelist ? <Trans>add_allowlist</Trans> : <Trans>add_blocklist</Trans>)}
```

### 3. Modal 组件 (`client/src/components/Filters/Modal.tsx`)

**修改**：
- 添加 `isRoutingRule` 参数
- 修改 `getTitle` 函数支持路由规则
- 将 `isRoutingRule` 传递给 Form 组件

**getTitle 函数修改**：
```tsx
const getTitle = (modalType: any, whitelist: any, isRoutingRule?: boolean) => {
    const titleType = MODAL_TYPE_TO_TITLE_TYPE_MAP[modalType];
    if (!titleType) {
        return null;
    }
    if (isRoutingRule) {
        return `${titleType}_routing_rule`;
    }
    return `${titleType}_${whitelist ? 'allowlist' : 'blocklist'}`;
};
```

### 4. Form 组件 (`client/src/components/Filters/Form.tsx`)

**修改**：
- 添加 `isRoutingRule` 参数
- 根据 `isRoutingRule` 显示不同的 placeholder 文本

**修改前**：
```tsx
{whitelist ? t('enter_valid_allowlist') : t('enter_valid_blocklist')}
```

**修改后**：
```tsx
{isRoutingRule ? t('enter_valid_routing_rule_url') : (whitelist ? t('enter_valid_allowlist') : t('enter_valid_blocklist'))}
```

### 5. DnsRouting 组件 (`client/src/components/Filters/DnsRouting.tsx`)

**修改**：
- 传递 `isRoutingRule={true}` 给 Actions 和 Modal 组件

```tsx
<Actions
    handleAdd={this.openAddFiltersModal}
    handleRefresh={this.handleRefresh}
    processingRefreshFilters={processingRefreshFilters}
    whitelist={whitelist}
    isRoutingRule={true}
/>

<Modal
    ...
    whitelist={whitelist}
    isRoutingRule={true}
/>
```

---

## 效果对比

### 修改前
- 按钮文本：**添加白名单**
- 弹窗标题：**新增白名单**
- Placeholder：**输入有效的白名单 URL**

### 修改后
- 按钮文本：**添加规则**
- 弹窗标题：**新增规则**
- Placeholder：**输入有效的规则 URL**

---

## 技术实现

### 参数传递链

```
DnsRouting (isRoutingRule=true)
    ↓
Actions (isRoutingRule=true)
    → 显示 "添加规则"
    
DnsRouting (isRoutingRule=true)
    ↓
Modal (isRoutingRule=true)
    ↓
    getTitle() → "new_routing_rule"
    ↓
Form (isRoutingRule=true)
    → 显示 "输入有效的规则 URL"
```

### 向后兼容

- 所有修改都是向后兼容的
- 原有的白名单和黑名单页面不受影响
- 通过可选参数 `isRoutingRule?` 实现

---

## 测试验证

### ✅ 编译测试
- [x] TypeScript 类型检查通过
- [x] 前端编译成功
- [x] 后端编译成功

### 待测试
- [ ] 域名分流页面显示正确的文案
- [ ] 白名单页面仍然显示原有文案
- [ ] 黑名单页面仍然显示原有文案

---

## 后续工作

### 1. 功能完善
- 添加 Clash 规则解析功能
- 显示规则数量统计
- 支持规则预览

### 2. UI 优化
- 添加规则类型选择（Clash / AdGuard）
- 显示规则来源信息
- 添加规则验证提示

### 3. 后端集成
- 实现 Clash 规则解析 API
- 保存规则到配置文件
- 应用规则到 DNS 查询

---

## 总结

✅ **UI 文案修复完成！**

- 域名分流页面现在使用专门的文案
- 与白名单/黑名单页面完全区分
- 代码结构清晰，易于维护
- 向后兼容，不影响现有功能

---

**修复日期**：2025年11月24日  
**修复人员**：Kiro AI  
**状态**：✅ 完成并编译成功
