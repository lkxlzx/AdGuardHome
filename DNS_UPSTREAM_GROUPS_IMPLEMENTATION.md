# DNS 上游服务器分组功能实现说明

## 一、功能概述

在 DNS 设置页面的"上游 DNS 服务器"部分添加分组管理功能，用户可以：
1. 创建多个上游 DNS 服务器分组
2. 为每个分组指定名称和服务器列表
3. 编辑、删除现有分组
4. 这些分组可在 DNS 分流模块中使用

## 二、已完成的文件

### 2.1 前端组件

#### 文件：`client/src/components/Settings/Dns/Upstream/UpstreamGroups.tsx`
**功能**：上游 DNS 服务器分组管理组件

**特性**：
- 显示所有已创建的分组
- 添加新分组（分组名称 + DNS 服务器列表）
- 编辑现有分组
- 删除分组（带确认提示，默认组不可删除）
- 设置默认组（分流规则未命中时使用）
- 默认组特殊标识（黄色边框 + 徽章）
- 内联编辑模式
- 响应式设计，符合现有 UI 风格

**数据结构**：
```typescript
interface UpstreamGroup {
    id: string;          // 唯一标识符
    name: string;        // 分组名称（用户自定义）
    upstreams: string;   // DNS 服务器列表（每行一个）
    isDefault?: boolean; // 是否为默认组（分流未命中时使用）
}
```

#### 文件：`client/src/components/Settings/Dns/Upstream/index.tsx`（已修改）
**修改内容**：
- 引入 `UpstreamGroups` 组件
- 添加本地状态管理分组数据
- 在表单提交时保存分组数据
- 分组变更时自动保存到后端

## 三、需要完成的工作

### 3.1 后端支持

#### 3.1.1 修改 DNS 配置结构
**文件**：`internal/dnsforward/dnsforward.go`

在 `Config` 结构体中添加：
```go
type Config struct {
    // ... 现有字段 ...
    
    // UpstreamGroups 上游 DNS 服务器分组
    UpstreamGroups []UpstreamGroup `yaml:"upstream_groups" json:"upstream_groups"`
}

// UpstreamGroup 上游服务器组
type UpstreamGroup struct {
    ID        string   `json:"id" yaml:"id"`
    Name      string   `json:"name" yaml:"name"`
    Upstreams []string `json:"upstreams" yaml:"upstreams"`
    IsDefault bool     `json:"is_default" yaml:"is_default"` // 是否为默认组
}
```

#### 3.1.2 修改 HTTP API
**文件**：`internal/dnsforward/http.go`

在 `jsonDNSConfig` 结构体中添加：
```go
type jsonDNSConfig struct {
    // ... 现有字段 ...
    
    // UpstreamGroups 上游服务器分组
    UpstreamGroups *[]UpstreamGroup `json:"upstream_groups"`
}
```

在 `handleGetConfig` 中返回分组数据：
```go
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
    // ... 现有代码 ...
    
    resp := jsonDNSConfig{
        // ... 现有字段 ...
        UpstreamGroups: &s.conf.UpstreamGroups,
    }
    
    // ...
}
```

在 `handleSetConfig` 中保存分组数据：
```go
func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
    // ... 现有代码 ...
    
    if req.UpstreamGroups != nil {
        s.conf.UpstreamGroups = *req.UpstreamGroups
        shouldRestart = true
    }
    
    // ...
}
```

#### 3.1.3 配置文件持久化
**文件**：`internal/home/config.go`

确保 `upstream_groups` 字段被正确序列化到 YAML 配置文件中。

### 3.2 前端 Redux 状态管理

#### 3.2.1 修改 DNS 配置 Reducer
**文件**：`client/src/reducers/dnsConfig.ts`

在初始状态中添加：
```typescript
const initialState = {
    // ... 现有字段 ...
    upstream_groups: [],
};
```

在成功获取配置时更新：
```typescript
[actions.getDnsConfigSuccess]: (state, { payload }) => ({
    ...state,
    // ... 现有字段 ...
    upstream_groups: payload.upstream_groups || [],
}),
```

#### 3.2.2 修改初始状态类型
**文件**：`client/src/initialState.ts`

添加类型定义：
```typescript
export interface UpstreamGroup {
    id: string;
    name: string;
    upstreams: string;
}

export interface DnsConfigState {
    // ... 现有字段 ...
    upstream_groups: UpstreamGroup[];
}
```

### 3.3 国际化翻译

需要在以下文件中添加翻译键（详见 `TRANSLATION_ADDITIONS.md`）：
- `client/src/__locales/zh-cn.json`
- `client/src/__locales/en.json`
- 其他语言文件（可选）

### 3.4 DNS 分流模块集成

#### 3.4.1 获取分组列表 API
**文件**：`internal/dnsforward/http_routing.go`

修改 `handleGetUpstreamGroups` 函数：
```go
func (s *Server) handleGetUpstreamGroups(w http.ResponseWriter, r *http.Request) {
    s.serverLock.RLock()
    defer s.serverLock.RUnlock()

    // 返回用户自定义的分组
    groups := make([]map[string]interface{}, 0, len(s.conf.UpstreamGroups))
    for _, group := range s.conf.UpstreamGroups {
        groups = append(groups, map[string]interface{}{
            "name":      group.Name,
            "upstreams": group.Upstreams,
        })
    }

    aghhttp.WriteJSONResponse(w, r, http.StatusOK, map[string]interface{}{
        "groups": groups,
    })
}
```

#### 3.4.2 DNS 查询处理
**文件**：`internal/dnsforward/process.go`

添加默认组支持和修改 `getUpstreamsByGroup` 函数：
```go
func (s *Server) getUpstreamsByGroup(groupName string) *proxy.UpstreamConfig {
    // 首先查找用户自定义分组
    for _, group := range s.conf.UpstreamGroups {
        if group.Name == groupName {
            if len(group.Upstreams) > 0 {
                return createUpstreamConfig(group.Upstreams)
            }
            return nil
        }
    }

    return nil
}

// getDefaultUpstreamGroup 获取默认上游分组
func (s *Server) getDefaultUpstreamGroup() *proxy.UpstreamConfig {
    // 查找标记为默认的分组
    for _, group := range s.conf.UpstreamGroups {
        if group.IsDefault && len(group.Upstreams) > 0 {
            return createUpstreamConfig(group.Upstreams)
        }
    }

    // 如果没有设置默认组，使用系统默认上游
    if len(s.conf.UpstreamDNS) > 0 {
        return createUpstreamConfig(s.conf.UpstreamDNS)
    }

    return nil
}
```

## 四、UI 设计说明

### 4.1 布局结构

```
┌─────────────────────────────────────────────────────┐
│ 上游 DNS 服务器                                      │
├─────────────────────────────────────────────────────┤
│ [现有的上游 DNS 配置表单]                            │
│                                                      │
│ ─────────────────────────────────────────────────── │
│                                                      │
│ 上游 DNS 服务器分组                                  │
│ 创建上游 DNS 服务器分组，可在 DNS 分流规则中使用。   │
│ 未命中分流规则的查询将使用默认组                     │
│                                                      │
│ ┌─────────────────────────────────────────────────┐ │
│ │ 国内 DNS [默认组]          [编辑][删除(禁用)]     │ │ ← 黄色边框
│ │ 分流规则未命中时使用此组                         │ │
│ │ 223.5.5.5                                        │ │
│ │ 119.29.29.29                                     │ │
│ └─────────────────────────────────────────────────┘ │
│                                                      │
│ ┌─────────────────────────────────────────────────┐ │
│ │ 国外 DNS                  [设为默认][编辑][删除]  │ │
│ │ 8.8.8.8                                          │ │
│ │ 1.1.1.1                                          │ │
│ └─────────────────────────────────────────────────┘ │
│                                                      │
│ [+ 添加分组]                                         │
└─────────────────────────────────────────────────────┘
```

### 4.2 添加/编辑分组表单

```
┌─────────────────────────────────────────────────────┐
│ 分组名称                                             │
│ ┌─────────────────────────────────────────────────┐ │
│ │ 国内 DNS                                         │ │
│ └─────────────────────────────────────────────────┘ │
│                                                      │
│ DNS 服务器列表                                       │
│ ┌─────────────────────────────────────────────────┐ │
│ │ 223.5.5.5                                        │ │
│ │ 119.29.29.29                                     │ │
│ │                                                  │ │
│ └─────────────────────────────────────────────────┘ │
│                                                      │
│ [保存] [取消]                                        │
└─────────────────────────────────────────────────────┘
```

### 4.3 样式特点

- 使用卡片式布局，与现有 DNS 黑名单页面风格一致
- 浅灰色背景区分不同分组
- **默认组使用黄色边框（2px）+ 黄色背景高亮**
- **默认组显示蓝色"默认组"徽章**
- **默认组显示提示文字"分流规则未命中时使用此组"**
- 内联编辑模式，点击编辑后原地展开表单
- 新增分组时使用蓝色边框高亮
- 默认组的删除按钮禁用状态
- 非默认组显示"设为默认"按钮
- 按钮使用 Bootstrap 样式类
- 响应式设计，移动端友好

## 五、测试要点

### 5.1 功能测试
- [ ] 添加新分组
- [ ] 编辑现有分组
- [ ] 删除分组（带确认）
- [ ] **设置默认组**
- [ ] **默认组不可删除（按钮禁用 + 提示）**
- [ ] **切换默认组（只能有一个默认组）**
- [ ] 分组名称验证（不能为空）
- [ ] DNS 服务器列表验证（不能为空）
- [ ] 配置保存和加载
- [ ] 页面刷新后数据持久化
- [ ] **默认组标识正确显示**

### 5.2 集成测试
- [ ] DNS 分流规则可以选择自定义分组
- [ ] **分流规则未命中时使用默认组**
- [ ] **默认组切换后，未命中查询使用新默认组**
- [ ] 分组删除后，使用该分组的分流规则如何处理
- [ ] 分组重命名后，分流规则是否需要更新
- [ ] **没有默认组时的回退机制（使用系统默认上游）**

### 5.3 UI 测试
- [ ] 多个分组的显示效果
- [ ] 编辑模式切换流畅
- [ ] 按钮禁用状态正确
- [ ] 移动端显示正常
- [ ] 长文本处理（分组名称、服务器列表）

## 六、后续优化建议

1. **分组排序**：允许用户拖拽调整分组顺序
2. **分组导入导出**：支持批量导入导出分组配置
3. **分组模板**：提供常用 DNS 服务器分组模板
4. **分组验证**：添加 DNS 服务器可用性测试
5. **分组统计**：显示每个分组的使用次数
6. **分组标签**：为分组添加标签或分类

## 七、注意事项

1. **向后兼容**：确保旧版本配置文件可以正常加载
2. **数据验证**：前后端都需要验证分组数据的有效性
3. **错误处理**：妥善处理分组不存在、服务器无效等异常情况
4. **性能考虑**：大量分组时的渲染性能
5. **并发安全**：后端分组数据的并发访问保护

## 八、开发进度

- [x] 前端组件开发（UpstreamGroups.tsx）
- [x] 前端页面集成（index.tsx）
- [x] 翻译文本准备
- [ ] 后端数据结构定义
- [ ] 后端 API 实现
- [ ] Redux 状态管理
- [ ] 配置文件持久化
- [ ] DNS 分流模块集成
- [ ] 测试和调试
- [ ] 文档完善
