# Clash 规则集成功能

## 功能概述

实现了 Clash 规则格式的解析和域名提取功能，用于域名分流规则管理。

---

## 功能特性

### 1. Clash 规则解析
- 支持从 URL 下载 Clash 规则文件
- 自动解析 YAML 格式
- 提取域名相关规则

### 2. 规则类型识别
支持的规则类型：
- `DOMAIN` - 完整域名匹配
- `DOMAIN-SUFFIX` - 域名后缀匹配
- `DOMAIN-KEYWORD` - 域名关键字匹配

过滤的规则类型：
- `IP-CIDR` - IP 地址段（自动过滤）
- `IP-CIDR6` - IPv6 地址段（自动过滤）

### 3. 格式转换
自动将 Clash 规则转换为 AdGuard Home 格式：
- `DOMAIN,example.com` → `example.com`
- `DOMAIN-SUFFIX,example.com` → `||example.com^`
- `DOMAIN-KEYWORD,keyword` → `*keyword*`

### 4. 统计信息
返回详细的规则统计：
- 总规则数
- 有效域名规则数
- DOMAIN 规则数
- DOMAIN-SUFFIX 规则数
- DOMAIN-KEYWORD 规则数
- IP 规则数（已过滤）

---

## 后端实现

### 1. 数据结构

#### DnsRoutingRule (`internal/dnsforward/config.go`)
```go
type DnsRoutingRule struct {
    ID          string   `yaml:"id" json:"id"`
    Name        string   `yaml:"name" json:"name"`
    URL         string   `yaml:"url" json:"url"`
    Domains     []string `yaml:"domains" json:"domains"`
    GroupID     string   `yaml:"group_id" json:"group_id"`
    Enabled     bool     `yaml:"enabled" json:"enabled"`
    RuleCount   int      `yaml:"rule_count" json:"rule_count"`
    LastUpdated int64    `yaml:"last_updated" json:"last_updated"`
}
```

#### ClashRuleStats (`internal/filtering/clash_rules.go`)
```go
type ClashRuleStats struct {
    TotalRules    int `json:"total_rules"`
    DomainRules   int `json:"domain_rules"`
    DomainSuffix  int `json:"domain_suffix"`
    DomainKeyword int `json:"domain_keyword"`
    IPRules       int `json:"ip_rules"`
    OtherRules    int `json:"other_rules"`
    ValidDomains  int `json:"valid_domains"`
}
```

### 2. 核心函数

#### ParseClashRules
```go
func ParseClashRules(url string) ([]string, *ClashRuleStats, error)
```
- 从 URL 下载并解析 Clash 规则
- 返回域名列表和统计信息

#### ValidateClashRuleURL
```go
func ValidateClashRuleURL(url string) (*ClashRuleStats, error)
```
- 验证 Clash 规则 URL
- 返回统计信息（不返回域名列表）

#### ConvertClashRuleToAdGuardFormat
```go
func ConvertClashRuleToAdGuardFormat(rule string) (string, bool)
```
- 转换单条 Clash 规则为 AdGuard Home 格式

### 3. HTTP API

#### POST /control/validate_clash_rule

验证 Clash 规则 URL 并返回统计信息。

**请求**：
```json
{
  "url": "https://example.com/rules.yaml"
}
```

**响应**：
```json
{
  "total_rules": 3727,
  "domain_rules": 20,
  "domain_suffix": 3677,
  "domain_keyword": 9,
  "ip_rules": 21,
  "other_rules": 0,
  "valid_domains": 3706
}
```

---

## 前端集成

### 1. UI 修改

#### 域名分流页面文案
- 按钮：**添加规则**
- 标题：**新增规则**
- 提示：**输入有效的规则 URL**

### 2. 翻译键

新增的翻译键（`client/src/__locales/zh-cn.json`）：
```json
{
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
}
```

### 3. 组件修改

修改的组件：
- `Actions.tsx` - 添加 `isRoutingRule` 参数
- `Modal.tsx` - 支持路由规则标题
- `Form.tsx` - 支持路由规则 placeholder
- `DnsRouting.tsx` - 传递 `isRoutingRule={true}`

---

## 使用示例

### 1. 验证 Clash 规则

```bash
curl -X POST http://localhost:80/control/validate_clash_rule \
  -H "Content-Type: application/json" \
  -d '{"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml"}'
```

**响应**：
```json
{
  "total_rules": 3727,
  "valid_domains": 3706,
  "domain_rules": 20,
  "domain_suffix": 3677,
  "domain_keyword": 9,
  "ip_rules": 21,
  "other_rules": 0
}
```

### 2. 解析规则（Go 代码）

```go
import "github.com/AdguardTeam/AdGuardHome/internal/filtering"

// 从 URL 解析
domains, stats, err := filtering.ParseClashRules(url)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("提取了 %d 条域名规则\n", len(domains))
fmt.Printf("总规则数: %d\n", stats.TotalRules)
fmt.Printf("有效域名: %d\n", stats.ValidDomains)
```

### 3. 配置文件格式

```yaml
dns:
  dns_routing_rules:
    - id: "rule_1701234567890"
      name: "中国域名"
      url: "https://example.com/china_domains.yaml"
      domains:
        - "||baidu.com^"
        - "||qq.com^"
        - "*taobao*"
      group_id: "group_1701234567890"
      enabled: true
      rule_count: 3706
      last_updated: 1701234567890
```

---

## 测试结果

### 测试规则文件
URL: `https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml`

### 解析结果
```
总规则数: 3727
有效域名规则: 3706
  - DOMAIN: 20
  - DOMAIN-SUFFIX: 3677
  - DOMAIN-KEYWORD: 9
IP 规则（已过滤）: 21
其他规则: 0
```

### 提取的域名示例
```
1. analytics.strava.com
2. blzddist1-a.akamaihd.net
3. cdn.angruo.com
4. cdn.lilyemby.com
5. client.amplifi.com
6. download.jetbrains.com
7. download.microsoft.com
8. images-cn.ssl-images-amazon.com
9. ip.istatmenus.app
10. kc.kexinshe.com
```

---

## 技术细节

### 1. YAML 解析

使用 `gopkg.in/yaml.v3` 解析 Clash 规则文件：

```go
type ClashRuleFile struct {
    Payload []string `yaml:"payload"`
}

var ruleFile ClashRuleFile
err := yaml.Unmarshal(content, &ruleFile)
```

### 2. 规则过滤

```go
for _, rule := range ruleFile.Payload {
    parts := strings.SplitN(rule, ",", 2)
    ruleType := strings.TrimSpace(parts[0])
    ruleValue := strings.TrimSpace(parts[1])
    
    switch ruleType {
    case "DOMAIN":
        domains = append(domains, ruleValue)
    case "DOMAIN-SUFFIX":
        domains = append(domains, "||"+ruleValue+"^")
    case "DOMAIN-KEYWORD":
        domains = append(domains, "*"+ruleValue+"*")
    case "IP-CIDR", "IP-CIDR6":
        // Skip IP rules
    }
}
```

### 3. HTTP 下载

```go
client := &http.Client{
    Timeout: 30 * time.Second,
}
resp, err := client.Get(url)
```

---

## 下一步开发

### 1. 前端功能
- [ ] 添加规则时调用验证 API
- [ ] 显示规则数量统计
- [ ] 规则预览功能
- [ ] 规则更新功能

### 2. 后端功能
- [ ] 规则缓存机制
- [ ] 定期更新规则
- [ ] 规则应用到 DNS 查询
- [ ] 规则优先级管理

### 3. 性能优化
- [ ] 规则编译和索引
- [ ] 快速域名匹配
- [ ] 内存优化

---

## 文件清单

### 新增文件
- `internal/filtering/clash_rules.go` - Clash 规则解析

### 修改文件
- `internal/dnsforward/config.go` - 添加 DnsRoutingRule 结构
- `internal/dnsforward/http.go` - 添加 API 端点
- `client/src/__locales/zh-cn.json` - 添加翻译
- `client/src/components/Filters/Actions.tsx` - UI 修改
- `client/src/components/Filters/Modal.tsx` - UI 修改
- `client/src/components/Filters/Form.tsx` - UI 修改
- `client/src/components/Filters/DnsRouting.tsx` - UI 修改

---

## 总结

✅ **Clash 规则集成完成！**

- 支持 Clash 规则格式解析
- 自动过滤 IP 规则，只保留域名规则
- 提供详细的统计信息
- HTTP API 可供前端调用
- UI 文案已更新

---

**完成日期**：2025年11月24日  
**开发人员**：Kiro AI  
**状态**：✅ 后端完成，前端待集成
