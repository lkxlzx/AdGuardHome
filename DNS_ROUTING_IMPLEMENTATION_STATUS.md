# DNS 域名分流规则实现状态

## 当前完成的功能

### ✅ 1. 数据结构扩展

#### Filter 结构体添加上游组字段
**文件**: `internal/filtering/filtering.go`

```go
type Filter struct {
    // ... 其他字段 ...
    
    // UpstreamGroup is the upstream group ID for DNS routing rules.
    // Only used for whitelist filters that serve as routing rules.
    UpstreamGroup string `yaml:"upstream_group,omitempty"`
}
```

**用途**: 
- 将白名单过滤器与上游组关联
- 标识哪些过滤器是域名分流规则

---

### ✅ 2. Clash 规则下载处理

#### 修改过滤器下载逻辑
**文件**: `internal/filtering/filter.go`

**功能**:
- 检测是否为 DNS 路由规则（白名单 + 有上游组）
- 检测是否为 Clash 规则 URL
- 只对 DNS 路由规则的 Clash 文件进行特殊处理
- 过滤掉 IP 规则，只保留域名规则

**代码逻辑**:
```go
// 检查是否为 DNS 路由规则
isDNSRoutingRule := flt.white && flt.UpstreamGroup != ""
isClashRule := IsClashRuleURL(flt.URL)

if isDNSRoutingRule && isClashRule {
    // 处理 Clash 规则：过滤 IP 规则，保留域名规则
    stats, err := ProcessClashRuleFile(r, tmpFile)
    // ...
}
```

**不影响**:
- ❌ 黑名单过滤器
- ❌ 普通白名单过滤器（没有上游组的）
- ❌ 非 Clash 规则的 URL

---

### ✅ 3. DNS 查询处理框架

#### 上游组查询函数
**文件**: `internal/dnsforward/upstream_groups.go`

**已实现的函数**:
```go
// 根据域名获取上游组 ID（待完善）
func (s *Server) GetUpstreamGroupForDomain(domain string) string

// 根据组 ID 获取上游服务器列表
func (s *Server) GetUpstreamStrings(groupID string) []string

// 根据组 ID 获取上游组对象
func (s *Server) GetUpstreamGroupByID(groupID string) *UpstreamGroup

// 根据组名获取上游组对象
func (s *Server) GetUpstreamGroupByName(name string) *UpstreamGroup
```

#### DNS 查询处理修改
**文件**: `internal/dnsforward/process.go`

**修改的函数**: `setCustomUpstream`

**新增逻辑**:
1. 首先检查域名分流规则
2. 如果匹配到规则，使用对应的上游组
3. 否则，使用客户端自定义上游（原有逻辑）

```go
func (s *Server) setCustomUpstream(ctx context.Context, pctx *proxy.DNSContext, clientID string) {
    // 1. 检查 DNS 路由规则
    if len(pctx.Req.Question) > 0 {
        domain := pctx.Req.Question[0].Name
        groupID := s.GetUpstreamGroupForDomain(domain)
        
        if groupID != "" {
            // 找到匹配的路由规则
            upstreams := s.GetUpstreamStrings(groupID)
            // TODO: 创建 CustomUpstreamConfig
        }
    }

    // 2. 回退到客户端自定义上游（原有逻辑）
    // ...
}
```

---

## 待完善的功能

### 🔄 1. 域名匹配逻辑

**需要实现**: `GetUpstreamGroupForDomain` 函数的完整逻辑

**要求**:
- 从白名单过滤器中加载域名规则
- 支持多种域名匹配模式：
  - `DOMAIN`: 精确匹配
  - `DOMAIN-SUFFIX`: 后缀匹配
  - `DOMAIN-KEYWORD`: 关键字匹配
- 按优先级匹配（精确 > 后缀 > 关键字）
- 缓存匹配结果以提高性能

**实现位置**: `internal/dnsforward/upstream_groups.go`

---

### 🔄 2. 上游配置创建

**需要实现**: 从上游服务器地址列表创建 `CustomUpstreamConfig`

**要求**:
- 将字符串地址转换为 `upstream.Upstream` 对象
- 创建 `proxy.CustomUpstreamConfig` 对象
- 处理上游服务器解析错误

**实现位置**: `internal/dnsforward/process.go`

**参考代码**:
```go
func createUpstreamConfig(upstreamStrings []string) *proxy.CustomUpstreamConfig {
    // TODO: 实现
    // 1. 解析上游服务器地址
    // 2. 创建 upstream.Upstream 对象
    // 3. 返回 CustomUpstreamConfig
}
```

---

### 🔄 3. 规则加载和缓存

**需要实现**: 
- 在服务器启动时加载所有 DNS 路由规则
- 构建域名匹配索引
- 实现规则更新机制

**数据结构建议**:
```go
type DNSRoutingRule struct {
    Domain        string
    MatchType     string // "DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD"
    UpstreamGroup string
    FilterID      int
}

type DNSRoutingCache struct {
    exactMatch    map[string]string // domain -> groupID
    suffixMatch   []DNSRoutingRule
    keywordMatch  []DNSRoutingRule
    mu            sync.RWMutex
}
```

---

### 🔄 4. HTTP API 扩展

**需要添加**: 过滤器 API 中的上游组字段

**文件**: `internal/dnsforward/http.go`

**修改的结构体**:
```go
type filterJSON struct {
    // ... 现有字段 ...
    UpstreamGroup string `json:"upstream_group,omitempty"`
}
```

**修改的 API**:
- `POST /control/filtering/add_url` - 添加过滤器时接收上游组
- `POST /control/filtering/set_url` - 更新过滤器时接收上游组
- `GET /control/filtering/status` - 返回过滤器时包含上游组

---

## 测试计划

### 单元测试

#### 1. Clash 规则处理测试
- ✅ 测试 Clash 规则解析
- ✅ 测试 IP 规则过滤
- ✅ 测试域名规则提取

#### 2. 域名匹配测试
- ⏳ 测试精确匹配
- ⏳ 测试后缀匹配
- ⏳ 测试关键字匹配
- ⏳ 测试优先级

#### 3. 上游选择测试
- ⏳ 测试规则匹配时的上游选择
- ⏳ 测试无匹配时的默认上游
- ⏳ 测试客户端自定义上游的优先级

### 集成测试

#### 1. 端到端测试
- ⏳ 添加 DNS 路由规则
- ⏳ 下载 Clash 规则文件
- ⏳ 验证规则已加载
- ⏳ 发送 DNS 查询
- ⏳ 验证使用了正确的上游

#### 2. 性能测试
- ⏳ 测试大量规则的加载时间
- ⏳ 测试域名匹配的性能
- ⏳ 测试并发查询的性能

---

## 实现优先级

### 高优先级 🔴

1. **域名匹配逻辑** - 核心功能
   - 实现 `GetUpstreamGroupForDomain`
   - 支持基本的域名匹配

2. **上游配置创建** - 核心功能
   - 实现 `createUpstreamConfig`
   - 将字符串转换为上游对象

3. **规则加载** - 核心功能
   - 在启动时加载规则
   - 构建匹配索引

### 中优先级 🟡

4. **HTTP API 扩展** - 用户界面
   - 添加上游组字段到 API
   - 更新前端集成

5. **规则更新机制** - 可靠性
   - 实现规则热更新
   - 避免服务中断

### 低优先级 🟢

6. **性能优化** - 优化
   - 实现匹配缓存
   - 优化查找算法

7. **监控和日志** - 可观测性
   - 添加详细日志
   - 添加性能指标

---

## 架构设计

### 数据流程

```
DNS 查询请求
    ↓
提取域名
    ↓
查找匹配的路由规则
    ↓
找到匹配？
    ├─ 是 → 使用规则指定的上游组
    └─ 否 → 使用默认上游或客户端自定义上游
    ↓
发送到上游服务器
    ↓
返回响应
```

### 规则匹配流程

```
域名: example.com
    ↓
1. 精确匹配查找
   - 查找 "example.com" → 找到？返回
    ↓
2. 后缀匹配查找
   - 查找 ".com" → 找到？返回
   - 查找 ".example.com" → 找到？返回
    ↓
3. 关键字匹配查找
   - 查找包含 "example" → 找到？返回
    ↓
4. 无匹配
   - 返回空字符串
```

---

## 配置示例

### AdGuardHome.yaml

```yaml
dns:
  upstream_groups:
    - id: "china"
      name: "China DNS"
      enabled: true
      is_default: false
      upstreams:
        - "223.5.5.5"
        - "119.29.29.29"
    
    - id: "global"
      name: "Global DNS"
      enabled: true
      is_default: true
      upstreams:
        - "8.8.8.8"
        - "1.1.1.1"

filters:
  - enabled: true
    url: "https://example.com/china_domains.yaml"
    name: "China Domains"
    id: 1
    upstream_group: "china"  # 关联到 China DNS 组

whitelist_filters:
  - enabled: true
    url: "https://example.com/global_domains.yaml"
    name: "Global Domains"
    id: 2
    upstream_group: "global"  # 关联到 Global DNS 组
```

---

## 总结

### 已完成 ✅
- 数据结构扩展（Filter 添加 UpstreamGroup 字段）
- Clash 规则下载处理（只处理 DNS 路由规则）
- DNS 查询处理框架（基本结构）
- 上游组查询函数（辅助函数）

### 进行中 🔄
- 域名匹配逻辑（框架已建立，待实现）
- 上游配置创建（待实现）

### 待开始 ⏳
- 规则加载和缓存
- HTTP API 扩展
- 完整的测试套件

### 不影响的部分 ✅
- 黑名单过滤器的下载和处理
- 普通白名单过滤器（无上游组）
- 非 Clash 规则的处理
- 现有的客户端自定义上游功能

---

**更新日期**: 2025年11月24日  
**状态**: 基础框架完成，核心功能待实现  
**编译状态**: ✅ 成功
