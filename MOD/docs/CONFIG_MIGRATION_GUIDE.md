# DNS配置迁移指南

## 概述

从 AdGuardHome v5 开始，DNS上游配置支持分组管理。旧的顶层配置字段仍然保留以确保向后兼容性。

---

## 配置字段说明

### 旧配置（仍然支持）

```yaml
dns:
  upstream_dns:
    - 8.8.8.8
    - 1.1.1.1
  bootstrap_dns:
    - 9.9.9.10
  fallback_dns: []
```

### 新配置（推荐）

```yaml
dns:
  # 旧字段保留作为后备配置
  upstream_dns:
    - 8.8.8.8
  bootstrap_dns:
    - 9.9.9.10
  fallback_dns: []
  
  # 新的分组配置
  upstream_groups:
    - id: "group-1"
      name: "国内"
      enabled: true
      is_default: true
      upstream_dns:
        - 114.114.114.111
      fallback_dns:
        - 223.6.6.6
      bootstrap_dns:
        - 9.9.9.10
    - id: "group-2"
      name: "海外"
      enabled: true
      is_default: false
      upstream_dns:
        - 8.8.8.8
      fallback_dns:
        - 1.1.1.1
      bootstrap_dns:
        - 9.9.9.10
```

---

## 配置优先级

### 1. 有默认分组时

```
默认分组配置 > 顶层配置
```

系统会使用默认分组的 `upstream_dns`、`fallback_dns` 和 `bootstrap_dns`。

### 2. 没有默认分组时

```
顶层配置
```

系统会使用顶层的 `upstream_dns`、`fallback_dns` 和 `bootstrap_dns`。

### 3. 所有分组都被禁用时

```
顶层配置（后备）
```

系统会回退到顶层配置。

---

## 迁移步骤

### 方案1：保留旧配置（推荐）

**适用场景**：希望保持向后兼容，或作为后备配置

1. 保留顶层的 `upstream_dns` 和 `bootstrap_dns`
2. 创建新的上游分组
3. 设置一个分组为默认

**优点**：
- ✅ 向后兼容
- ✅ 有后备配置
- ✅ 可以随时切换回旧配置

**配置示例**：
```yaml
dns:
  # 保留作为后备
  upstream_dns:
    - 8.8.8.8
  bootstrap_dns:
    - 9.9.9.10
  
  # 新的分组配置
  upstream_groups:
    - id: "default-group"
      name: "默认"
      enabled: true
      is_default: true
      upstream_dns:
        - 114.114.114.111
      fallback_dns:
        - 223.6.6.6
      bootstrap_dns:
        - 9.9.9.10
```

---

### 方案2：清空旧配置（不推荐）

**适用场景**：确定只使用分组功能，不需要后备配置

1. 创建上游分组
2. 设置默认分组
3. 清空顶层的 `upstream_dns`（保留空数组）

**缺点**：
- ⚠️ 如果所有分组被禁用，DNS将无法工作
- ⚠️ 失去后备配置

**配置示例**：
```yaml
dns:
  # 清空但保留字段
  upstream_dns: []
  bootstrap_dns: []
  fallback_dns: []
  
  # 必须有至少一个启用的默认分组
  upstream_groups:
    - id: "default-group"
      name: "默认"
      enabled: true
      is_default: true
      upstream_dns:
        - 114.114.114.111
      fallback_dns:
        - 223.6.6.6
      bootstrap_dns:
        - 9.9.9.10
```

---

## 常见问题

### Q1: 可以完全删除顶层的 upstream_dns 字段吗？

**A**: 不建议。原因：
1. 这些字段是配置结构的一部分，删除可能导致配置解析错误
2. 作为后备配置很有用
3. 保持向后兼容性

**建议**: 保留字段，可以设置为空数组 `[]`

---

### Q2: 如果顶层和分组都配置了，哪个生效？

**A**: 如果有启用的默认分组，使用分组配置；否则使用顶层配置。

---

### Q3: 可以不配置分组，只用顶层配置吗？

**A**: 可以。系统完全向后兼容，不配置分组时会使用顶层配置。

---

### Q4: 分组配置保存在哪里？

**A**: 保存在 `AdGuardHome.yaml` 配置文件的 `dns.upstream_groups` 字段中。

---

## 最佳实践

### 推荐配置

```yaml
dns:
  # 保留简单的后备配置
  upstream_dns:
    - 8.8.8.8
  bootstrap_dns:
    - 9.9.9.10
  fallback_dns: []
  
  # 使用分组管理复杂配置
  upstream_groups:
    - id: "domestic"
      name: "国内DNS"
      enabled: true
      is_default: true
      upstream_dns:
        - 114.114.114.111
        - 223.5.5.5
      fallback_dns:
        - 223.6.6.6
      bootstrap_dns:
        - 9.9.9.10
    
    - id: "international"
      name: "国际DNS"
      enabled: true
      is_default: false
      upstream_dns:
        - 8.8.8.8
        - 1.1.1.1
      fallback_dns:
        - 9.9.9.9
      bootstrap_dns:
        - 9.9.9.10
```

### 配置说明

1. **顶层配置**：简单的后备配置，使用可靠的公共DNS
2. **分组配置**：
   - 国内DNS分组：使用国内DNS服务器，设为默认
   - 国际DNS分组：使用国际DNS服务器，可手动切换

---

## 迁移检查清单

- [ ] 创建至少一个上游分组
- [ ] 设置一个分组为默认
- [ ] 测试分组功能是否正常
- [ ] 保留顶层配置作为后备
- [ ] 重启 AdGuardHome 验证配置
- [ ] 测试DNS解析是否正常

---

## 总结

### ✅ 推荐做法
- 保留顶层的 `upstream_dns` 和 `bootstrap_dns` 作为后备配置
- 使用分组功能管理多套DNS配置
- 设置一个分组为默认

### ⚠️ 不推荐做法
- 删除顶层配置字段
- 清空所有配置
- 禁用所有分组

### 🎯 最佳实践
- 顶层配置：简单可靠的后备配置
- 分组配置：灵活的多场景配置
- 定期测试：确保DNS解析正常
