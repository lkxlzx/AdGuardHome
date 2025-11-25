# DNS 路由规则定时更新功能

## 📋 功能概述

为 DNS 路由规则添加了可自定义的定时更新间隔功能，支持分钟级精度的自动更新控制。

---

## ✨ 新增功能

### 1. 独立更新间隔

每个 DNS 路由规则可以设置独立的更新间隔，不再受全局更新间隔限制。

### 2. 分钟级精度

更新间隔以分钟为单位，提供更灵活的更新控制：
- 1 分钟 = 最快更新
- 60 分钟 = 每小时更新
- 1440 分钟 = 每天更新
- 0 分钟 = 禁用自动更新

### 3. 禁用自动更新

设置为 0 可以禁用特定规则的自动更新，适用于：
- 静态规则列表
- 手动管理的规则
- 不需要频繁更新的规则

---

## 🎯 使用场景

### 场景 1: 高频更新规则
```
规则名称: 实时广告拦截
更新间隔: 30 分钟
说明: 广告规则变化快，需要频繁更新
```

### 场景 2: 日常更新规则
```
规则名称: CN 域名列表
更新间隔: 1440 分钟（24 小时）
说明: 域名列表相对稳定，每天更新一次即可
```

### 场景 3: 静态规则
```
规则名称: 自定义规则
更新间隔: 0 分钟
说明: 手动维护的规则，不需要自动更新
```

### 场景 4: 按需更新
```
规则名称: 测试规则
更新间隔: 5 分钟
说明: 测试期间需要快速更新
```

---

## 🖥️ 界面说明

### 添加规则界面

```
┌─────────────────────────────────────┐
│ 编辑规则                             │
├─────────────────────────────────────┤
│ 名称: [CN域名列表              ]    │
│                                      │
│ URL:  [https://example.com/...  ]   │
│                                      │
│ 目标上游组: [国内 ▼]                │
│                                      │
│ 定时更新间隔（分钟）: [1440    ]    │
│ 设置规则自动更新的时间间隔，        │
│ 单位为分钟。设置为 0 表示不自动更新 │
│                                      │
│           [取消]  [保存]             │
└─────────────────────────────────────┘
```

### 字段说明

- **定时更新间隔（分钟）**: 
  - 类型: 数字输入框
  - 范围: ≥ 0
  - 默认值: 0（不自动更新）
  - 单位: 分钟

---

## 🔧 技术实现

### 前端实现

#### 1. 表单字段

```typescript
type FormValues = {
    enabled: boolean;
    name: string;
    url: string;
    upstreamGroup?: string;
    updateInterval?: number;  // 新增字段
};
```

#### 2. 输入验证

```typescript
rules={{
    validate: (value) => {
        const num = Number(value);
        if (isNaN(num) || num < 0) {
            return t('form_error_positive_number');
        }
        return true;
    }
}}
```

#### 3. UI 组件

```tsx
<input
    type="number"
    id="updateInterval"
    className="form-control"
    placeholder="0"
    min="0"
    step="1"
    onChange={(e) => field.onChange(Number(e.target.value))}
/>
```

### 后端实现

#### 1. 数据结构

```go
type FilterYAML struct {
    Enabled        bool
    URL            string
    Name           string
    UpdateInterval int       `yaml:"update_interval"` // 分钟
    // ... 其他字段
}
```

#### 2. API 接口

```go
type filterAddJSON struct {
    Name           string `json:"name"`
    URL            string `json:"url"`
    UpdateInterval int    `json:"update_interval"`
    // ... 其他字段
}
```

#### 3. 更新逻辑

```go
func (d *DNSFilter) listsToUpdate(filters *[]FilterYAML, force bool) (toUpd []FilterYAML) {
    for i := range *filters {
        flt := &(*filters)[i]
        
        if !flt.Enabled {
            continue
        }
        
        if !force {
            var updateInterval time.Duration
            if flt.UpdateInterval > 0 {
                // 使用自定义间隔（分钟）
                updateInterval = time.Duration(flt.UpdateInterval) * time.Minute
            } else if flt.UpdateInterval == 0 && flt.dnsRouting {
                // DNS 路由规则设置为 0 表示不自动更新
                continue
            } else {
                // 使用全局间隔（小时）
                updateInterval = time.Duration(d.conf.FiltersUpdateIntervalHours) * time.Hour
            }
            
            exp := flt.LastUpdated.Add(updateInterval)
            if now.Before(exp) {
                continue
            }
        }
        
        // 添加到更新列表
        toUpd = append(toUpd, flt)
    }
    
    return toUpd
}
```

---

## 📊 更新间隔对照表

| 间隔（分钟） | 等效时间 | 适用场景 |
|-------------|---------|---------|
| 0 | 不更新 | 静态规则、手动管理 |
| 5 | 5 分钟 | 测试、开发 |
| 15 | 15 分钟 | 高频更新 |
| 30 | 半小时 | 广告拦截 |
| 60 | 1 小时 | 常规更新 |
| 120 | 2 小时 | 中频更新 |
| 360 | 6 小时 | 低频更新 |
| 720 | 12 小时 | 半天更新 |
| 1440 | 24 小时 | 每日更新 |
| 10080 | 7 天 | 每周更新 |

---

## 🔄 更新流程

### 自动更新流程

```
定时检查（每分钟）
    ↓
遍历所有 DNS 路由规则
    ↓
检查规则是否启用
    ↓
检查 UpdateInterval
    ├─ = 0 → 跳过（不自动更新）
    └─ > 0 → 检查上次更新时间
        ↓
    计算下次更新时间
    LastUpdated + UpdateInterval
        ↓
    当前时间 >= 下次更新时间？
        ├─ 是 → 下载并更新规则
        └─ 否 → 跳过
```

### 手动更新流程

```
用户点击"刷新规则"
    ↓
强制更新所有启用的规则
    ↓
忽略 UpdateInterval 设置
    ↓
下载并更新所有规则
```

---

## 💡 使用建议

### 1. 合理设置更新间隔

- **广告拦截规则**: 30-60 分钟
- **域名分流规则**: 1440 分钟（每天）
- **静态规则**: 0（禁用）
- **测试规则**: 5-15 分钟

### 2. 避免过于频繁

- 不建议设置小于 5 分钟
- 过于频繁会增加网络负载
- 大多数规则列表不会频繁变化

### 3. 考虑服务器负载

- 多个规则同时更新会增加负载
- 可以错开不同规则的更新时间
- 例如: 规则 A 每 60 分钟，规则 B 每 90 分钟

### 4. 监控更新状态

- 查看"上次更新时间"
- 检查规则数量是否正常
- 注意更新失败的规则

---

## 🔍 配置示例

### 示例 1: 多规则配置

```yaml
dns_routing_filters:
  - name: "CN 域名"
    url: "https://example.com/cn-domains.txt"
    enabled: true
    update_interval: 1440  # 每天更新
    upstream_group: "domestic"
    
  - name: "广告拦截"
    url: "https://example.com/ad-rules.txt"
    enabled: true
    update_interval: 60    # 每小时更新
    upstream_group: "adblock"
    
  - name: "自定义规则"
    url: "file:///path/to/custom.txt"
    enabled: true
    update_interval: 0     # 不自动更新
    upstream_group: "custom"
```

### 示例 2: API 请求

#### 添加规则

```bash
POST /control/filtering/add_url
Content-Type: application/json

{
  "name": "CN域名列表",
  "url": "https://example.com/cn-domains.txt",
  "dns_routing": true,
  "upstream_group": "domestic",
  "update_interval": 1440
}
```

#### 更新规则

```bash
POST /control/filtering/set_url
Content-Type: application/json

{
  "url": "https://example.com/cn-domains.txt",
  "dns_routing": true,
  "data": {
    "name": "CN域名列表",
    "url": "https://example.com/cn-domains.txt",
    "enabled": true,
    "upstream_group": "domestic",
    "update_interval": 720
  }
}
```

---

## 🐛 故障排查

### 问题 1: 规则没有自动更新

**可能原因**:
- UpdateInterval 设置为 0
- 规则被禁用
- 上次更新时间未到

**解决方法**:
1. 检查 UpdateInterval 值
2. 确认规则已启用
3. 查看上次更新时间
4. 手动刷新规则测试

### 问题 2: 更新过于频繁

**可能原因**:
- UpdateInterval 设置过小
- 多个规则同时更新

**解决方法**:
1. 增加 UpdateInterval 值
2. 错开不同规则的更新时间
3. 监控系统负载

### 问题 3: 更新失败

**可能原因**:
- 网络连接问题
- URL 无效
- 服务器返回错误

**解决方法**:
1. 检查网络连接
2. 验证 URL 是否有效
3. 查看错误日志
4. 尝试手动刷新

---

## 📈 性能影响

### 内存使用
- 每个规则增加 4 字节（int 类型）
- 影响可忽略不计

### CPU 使用
- 定时检查逻辑轻量级
- 只在需要更新时下载规则
- 影响可忽略不计

### 网络使用
- 取决于更新频率和规则大小
- 建议合理设置更新间隔
- 避免不必要的频繁更新

---

## 🔄 向后兼容性

### 配置文件兼容
- ✅ 旧配置文件自动兼容
- ✅ 未设置 UpdateInterval 默认为 0
- ✅ 不影响现有规则

### API 兼容
- ✅ UpdateInterval 为可选字段
- ✅ 不传递时默认为 0
- ✅ 旧客户端仍可正常工作

---

## 📝 更新日志

### V2 (2024-11-25)
- ✅ 添加 UpdateInterval 字段
- ✅ 实现分钟级更新间隔
- ✅ 支持禁用自动更新（设置为 0）
- ✅ 前端表单添加输入字段
- ✅ 后端实现更新逻辑
- ✅ 添加中文翻译

---

## 🎯 未来计划

### V2.1
- [ ] 添加更新历史记录
- [ ] 显示下次更新时间
- [ ] 更新失败重试机制

### V2.2
- [ ] 更新成功/失败通知
- [ ] 更新统计信息
- [ ] 批量设置更新间隔

---

**功能状态**: ✅ 已实现
**版本**: V2
**提交**: a03979ef
