# DNS路由规则独立于保护开关的修复

## 问题描述

**发现日期**: 2024-11-25  
**严重程度**: 🔴 高  
**影响范围**: 所有DNS路由规则

### 问题现象

当用户取消勾选"使用过滤器和 Hosts 文件以拦截指定域名"（`ProtectionEnabled`）时，**所有DNS路由规则都失效了**。

这是不合理的，因为：
- DNS路由规则是"路由"功能，不是"过滤"功能
- 用户可能只想关闭广告过滤，但仍然需要DNS路由
- 这两个功能应该是独立的

### 根本原因

在 `internal/filtering/filtering.go` 的 `matchHost` 函数中：

```go
// 问题代码（修复前）
if setts.ProtectionEnabled && d.filteringEngineAllow != nil {
    dnsres, ok := d.filteringEngineAllow.MatchRequest(ufReq)
    if ok {
        return d.matchHostProcessAllowList(ctx, host, dnsres)
    }
}
```

**问题**：
- DNS路由规则存储在 `filteringEngineAllow`（允许列表引擎）中
- 但只有当 `ProtectionEnabled` 为 `true` 时才会检查
- 当用户关闭保护开关时，DNS路由规则也被跳过了

### 设计问题

DNS路由规则被错误地归类为"过滤"功能的一部分，但实际上它们是：
- **路由功能**：决定使用哪个上游DNS服务器
- **不是过滤功能**：不拦截或修改DNS响应

---

## 解决方案

### 修复逻辑

DNS路由规则应该**独立于** `ProtectionEnabled` 开关：

```go
// 修复后
// Check allow list (including DNS routing rules)
// DNS routing rules should work regardless of ProtectionEnabled
if d.filteringEngineAllow != nil {
    dnsres, ok := d.filteringEngineAllow.MatchRequest(ufReq)
    if ok {
        result, err := d.matchHostProcessAllowList(ctx, host, dnsres)
        if err != nil {
            return Result{}, err
        }
        // If this is a DNS routing rule (has UpstreamGroup), return it even if protection is disabled
        if result.UpstreamGroup != "" || setts.ProtectionEnabled {
            return result, nil
        }
        // Otherwise, only return if protection is enabled
    }
}
```

### 关键改进

1. **移除 `ProtectionEnabled` 检查**：允许列表引擎总是被检查
2. **区分路由和过滤**：
   - 如果结果包含 `UpstreamGroup`（DNS路由规则）→ 总是返回
   - 如果是普通的允许列表规则 → 只在保护开启时返回
3. **保持向后兼容**：不影响其他过滤功能

---

## 行为对比

### 修复前

| 场景 | ProtectionEnabled=true | ProtectionEnabled=false |
|------|----------------------|------------------------|
| DNS路由规则 | ✅ 工作 | ❌ 失效 |
| 广告过滤 | ✅ 工作 | ❌ 关闭 |
| 允许列表 | ✅ 工作 | ❌ 失效 |

### 修复后

| 场景 | ProtectionEnabled=true | ProtectionEnabled=false |
|------|----------------------|------------------------|
| DNS路由规则 | ✅ 工作 | ✅ 工作 |
| 广告过滤 | ✅ 工作 | ❌ 关闭 |
| 允许列表 | ✅ 工作 | ❌ 关闭 |

---

## 用户场景

### 场景1：只想关闭广告过滤，保留DNS路由

**用户需求**：
- 暂时关闭广告过滤（可能某些网站被误拦截）
- 但仍然需要DNS路由（国内外分流）

**修复前**：❌ 无法实现，关闭保护后DNS路由也失效  
**修复后**：✅ 可以实现，DNS路由独立工作

### 场景2：测试DNS路由规则

**用户需求**：
- 测试DNS路由配置是否正确
- 不想被广告过滤干扰

**修复前**：❌ 必须开启保护才能测试  
**修复后**：✅ 可以独立测试DNS路由

---

## 技术细节

### 判断逻辑

修复后的判断逻辑：

```go
if result.UpstreamGroup != "" {
    // 这是DNS路由规则，总是返回
    return result, nil
} else if setts.ProtectionEnabled {
    // 这是普通允许列表规则，只在保护开启时返回
    return result, nil
}
// 保护关闭且不是DNS路由规则，继续检查其他规则
```

### Result结构

```go
type Result struct {
    // ...
    UpstreamGroup string  // 如果非空，表示这是DNS路由规则
    // ...
}
```

---

## 测试验证

### 测试步骤

1. **配置DNS路由规则**
   ```yaml
   dns_routing_filters:
     - enabled: true
       url: https://example.com/rules.txt
       name: 测试规则
       upstream_group: group_test
   ```

2. **测试场景A：保护开启**
   - 勾选"使用过滤器和 Hosts 文件以拦截指定域名"
   - 查询匹配DNS路由规则的域名
   - 预期：✅ 使用指定的上游组

3. **测试场景B：保护关闭**
   - 取消勾选"使用过滤器和 Hosts 文件以拦截指定域名"
   - 查询匹配DNS路由规则的域名
   - 修复前：❌ 不使用指定的上游组
   - 修复后：✅ 使用指定的上游组

### 验证命令

```bash
# 查询测试域名
nslookup test.example.com 127.0.0.1

# 检查日志，应该显示使用了DNS路由规则
# 即使保护开关关闭
```

---

## 影响分析

### 正面影响

1. **功能独立性** ✅
   - DNS路由和广告过滤解耦
   - 用户可以独立控制

2. **用户体验** ✅
   - 更符合用户预期
   - 更灵活的配置选项

3. **功能完整性** ✅
   - DNS路由作为核心功能，不应该被保护开关影响

### 潜在风险

1. **行为变化** ⚠️
   - 风险：用户可能习惯了旧行为
   - 缓解：这是bug修复，新行为更合理
   - 评估：**低风险**

2. **性能影响** ⚠️
   - 风险：保护关闭时仍然检查允许列表
   - 影响：微小（只是一次额外的检查）
   - 评估：**可忽略**

### 风险等级：低 ✅

---

## 向后兼容性

✅ **完全兼容**
- 不影响现有配置
- 不改变API接口
- 保护开启时行为完全相同
- 保护关闭时行为更合理

---

## 相关问题

### 为什么DNS路由规则在允许列表引擎中？

历史原因：
1. DNS路由规则使用与允许列表相同的规则格式
2. 它们都是"不拦截"的规则
3. 但DNS路由规则有额外的 `UpstreamGroup` 属性

### 是否应该重构？

长期来看，可以考虑：
1. 将DNS路由规则移到独立的引擎
2. 创建专门的DNS路由检查器
3. 完全独立于过滤系统

但当前的修复已经足够解决问题。

---

## 修改的文件

### internal/filtering/filtering.go

**修改位置**：`matchHost` 函数（第930-943行）

**修改内容**：
- 移除 `ProtectionEnabled` 检查条件
- 添加 `UpstreamGroup` 判断逻辑
- 区分DNS路由规则和普通允许列表规则

---

## 建议

### 立即行动

1. ✅ 应用此修复（已完成）
2. ✅ 重新编译（已完成）
3. 📝 更新用户文档
4. 📝 在发布说明中说明此修复

### 用户通知

建议在发布说明中说明：

> **重要修复**：DNS路由规则现在独立于"使用过滤器和 Hosts 文件以拦截指定域名"开关。
> 
> 即使关闭保护开关，DNS路由规则仍然会生效。这使得用户可以在关闭广告过滤的同时，
> 继续使用DNS路由功能（如国内外分流）。

### 文档更新

1. 在用户手册中明确说明：
   - DNS路由规则独立于保护开关
   - 保护开关只影响广告过滤、安全浏览等功能
   - DNS路由规则总是生效（如果启用）

2. 在界面上添加说明：
   - 在保护开关旁边添加提示
   - 说明DNS路由规则不受此开关影响

---

## 总结

这是一个**重要的功能修复**：

✅ **修复内容**：
- DNS路由规则独立于保护开关
- 用户可以关闭广告过滤但保留DNS路由
- 更符合用户预期和功能设计

✅ **影响范围**：
- 所有使用DNS路由规则的用户
- 特别是需要灵活控制过滤和路由的用户

✅ **风险评估**：
- 低风险
- 完全向后兼容
- 行为更合理

**强烈建议所有使用DNS路由功能的用户升级到此版本！**

---

**修复日期**: 2024-11-25  
**修复版本**: V3 Latest  
**修复工程师**: Kiro AI  
**审核状态**: ✅ 已验证
