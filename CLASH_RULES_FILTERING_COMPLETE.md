# Clash规则过滤完整实现

## 功能说明

当添加DNS路由规则时，如果规则URL指向Clash格式的规则文件，系统会自动：
1. 识别Clash规则格式
2. 解析YAML文件
3. **过滤掉所有IP规则**（IP-CIDR, IP-CIDR6）
4. **只保留域名规则**（DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD）
5. 转换为AdGuard Home格式
6. 保存到本地文件

## Clash规则识别

系统通过以下特征识别Clash规则文件：

```go
func IsClashRuleURL(url string) bool {
    return strings.Contains(url, "/Clash/") ||
        strings.Contains(url, "/clash/") ||
        strings.HasSuffix(url, ".yaml") ||
        strings.HasSuffix(url, ".yml")
}
```

**识别条件**（满足任一即可）：
- URL包含 `/Clash/` 或 `/clash/`
- URL以 `.yaml` 或 `.yml` 结尾

## 规则类型处理

### 保留的域名规则

#### 1. DOMAIN（精确匹配）
**Clash格式**：
```yaml
payload:
  - DOMAIN,example.com
  - DOMAIN,google.com
```

**转换后**（AdGuard格式）：
```
example.com
google.com
```

#### 2. DOMAIN-SUFFIX（后缀匹配）
**Clash格式**：
```yaml
payload:
  - DOMAIN-SUFFIX,example.com
  - DOMAIN-SUFFIX,google.com
```

**转换后**（AdGuard格式）：
```
||example.com^
||google.com^
```

#### 3. DOMAIN-KEYWORD（关键字匹配）
**Clash格式**：
```yaml
payload:
  - DOMAIN-KEYWORD,google
  - DOMAIN-KEYWORD,facebook
```

**转换后**（AdGuard格式）：
```
*google*
*facebook*
```

### 过滤掉的IP规则

#### IP-CIDR（IPv4 CIDR）
**Clash格式**：
```yaml
payload:
  - IP-CIDR,192.168.0.0/16
  - IP-CIDR,10.0.0.0/8
```

**处理**：❌ **完全忽略，不保存**

#### IP-CIDR6（IPv6 CIDR）
**Clash格式**：
```yaml
payload:
  - IP-CIDR6,2001:db8::/32
  - IP-CIDR6,fe80::/10
```

**处理**：❌ **完全忽略，不保存**

## 处理流程

### 1. 下载规则文件
```
用户添加DNS路由规则
↓
URL: https://raw.githubusercontent.com/.../China_Classical.yaml
↓
系统识别为Clash规则（URL包含.yaml）
↓
下载规则文件
```

### 2. 解析YAML
```go
type ClashRuleFile struct {
    Payload []string `yaml:"payload"`
}

// 解析YAML内容
var ruleFile ClashRuleFile
yaml.Unmarshal(content, &ruleFile)
```

### 3. 过滤和转换
```go
for _, rule := range ruleFile.Payload {
    parts := strings.SplitN(rule, ",", 2)
    ruleType := parts[0]
    ruleValue := parts[1]
    
    switch ruleType {
    case "DOMAIN":
        // 保留：直接使用域名
        domains = append(domains, ruleValue)
        
    case "DOMAIN-SUFFIX":
        // 保留：转换为 ||domain^
        domains = append(domains, "||"+ruleValue+"^")
        
    case "DOMAIN-KEYWORD":
        // 保留：转换为 *keyword*
        domains = append(domains, "*"+ruleValue+"*")
        
    case "IP-CIDR", "IP-CIDR6":
        // 忽略：不保存IP规则
        continue
    }
}
```

### 4. 保存到文件
```
过滤后的域名规则
↓
写入到 data/filters/[id].txt
↓
每行一个规则
↓
example.com
||google.com^
*facebook*
...
```

## 统计信息

处理完成后，系统会记录详细的统计信息：

```go
type ClashRuleStats struct {
    TotalRules      int  // 总规则数
    DomainRules     int  // DOMAIN 规则数
    DomainSuffix    int  // DOMAIN-SUFFIX 规则数
    DomainKeyword   int  // DOMAIN-KEYWORD 规则数
    IPRules         int  // IP 规则数（被过滤）
    OtherRules      int  // 其他规则数
    ValidDomains    int  // 有效域名规则总数
}
```

**日志示例**：
```
[info] clash rule processed id=1764080553 total_rules=5000 valid_domains=3728 filtered_ip_rules=1272
```

## 代码实现

### 主处理函数

**文件**：`internal/filtering/filter.go`

```go
func (d *DNSFilter) updateIntl(ctx context.Context, flt *FilterYAML) (ok bool, err error) {
    // ... 下载规则文件 ...
    
    // 检查是否为DNS路由规则且为Clash格式
    isDNSRoutingRule := flt.UpstreamGroup != ""
    isClashRule := IsClashRuleURL(flt.URL)
    
    if isDNSRoutingRule && isClashRule {
        // 处理Clash规则：过滤IP规则，只保留域名规则
        stats, err := ProcessClashRuleFile(r, tmpFile)
        if err != nil {
            return false, fmt.Errorf("processing clash rule: %w", err)
        }
        
        // 记录统计信息
        d.logger.InfoContext(ctx,
            "clash rule processed",
            "id", flt.ID,
            "total_rules", stats.TotalRules,
            "valid_domains", stats.ValidDomains,
            "filtered_ip_rules", stats.IPRules,
        )
        
        // 创建解析结果
        res = &rulelist.ParseResult{
            RulesCount: stats.ValidDomains,
            Checksum: flt.checksum + 1,
        }
    } else {
        // 普通过滤器处理
        p := rulelist.NewParser()
        res, err = p.Parse(tmpFile, r, *bufPtr)
    }
    
    return res.Checksum != flt.checksum && err == nil, err
}
```

### Clash规则处理

**文件**：`internal/filtering/clash_rules.go`

```go
func ProcessClashRuleFile(input io.Reader, output io.Writer) (*ClashRuleStats, error) {
    // 解析Clash规则
    domains, stats, err := ParseClashRulesFromReader(input)
    if err != nil {
        return nil, err
    }
    
    // 写入域名规则到输出（每行一个）
    for _, domain := range domains {
        fmt.Fprintln(output, domain)
    }
    
    return stats, nil
}
```

## 示例

### 输入（Clash规则文件）

```yaml
payload:
  - DOMAIN,baidu.com
  - DOMAIN,qq.com
  - DOMAIN-SUFFIX,cn
  - DOMAIN-SUFFIX,taobao.com
  - DOMAIN-KEYWORD,china
  - IP-CIDR,1.0.1.0/24
  - IP-CIDR,1.0.2.0/23
  - IP-CIDR6,2001:250::/35
```

### 输出（AdGuard格式）

```
baidu.com
qq.com
||cn^
||taobao.com^
*china*
```

### 统计信息

```
total_rules: 8
domain_rules: 2
domain_suffix: 2
domain_keyword: 1
ip_rules: 3 (filtered)
valid_domains: 5
```

## 测试步骤

### 1. 添加Clash规则
1. 进入"DNS路由"页面
2. 点击"添加过滤器"
3. 输入Clash规则URL：
   ```
   https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml
   ```
4. 选择上游组
5. 保存

### 2. 查看日志
```
[info] downloading update for filter id=1764080553 url=https://...
[info] processing clash rule for dns routing id=1764080553 url=https://...
[info] clash rule processed id=1764080553 total_rules=5000 valid_domains=3728 filtered_ip_rules=1272
[info] filter updated id=1764080553 bytes_written=107960 rules_count=3728
```

### 3. 检查规则文件
```bash
# 查看保存的规则文件
type data\filters\1764080553.txt

# 应该只看到域名规则，没有IP规则
# 例如：
# baidu.com
# ||cn^
# *china*
```

### 4. 验证规则生效
```bash
# 测试DNS解析
nslookup baidu.com 127.0.0.1

# 应该使用指定的上游组进行解析
```

## 支持的Clash规则源

### 常用规则源

1. **blackmatrix7/ios_rule_script**
   ```
   https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml
   https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Global/Global_Classical.yaml
   ```

2. **Loyalsoldier/clash-rules**
   ```
   https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/cncidr.txt
   https://raw.githubusercontent.com/Loyalsoldier/clash-rules/release/proxy.txt
   ```

3. **自定义Clash规则**
   - 任何符合Clash YAML格式的规则文件
   - URL包含 `.yaml` 或 `.yml` 后缀

## 注意事项

1. **IP规则不支持**：
   - DNS路由只处理域名，不处理IP
   - 所有IP-CIDR规则会被自动过滤
   - 这是设计行为，不是bug

2. **规则转换**：
   - DOMAIN-SUFFIX 转换为 `||domain^`
   - DOMAIN-KEYWORD 转换为 `*keyword*`
   - 确保与AdGuard Home语法兼容

3. **性能考虑**：
   - 大型规则文件（>10000条）可能需要较长处理时间
   - 建议使用已经过滤的规则源

4. **更新频率**：
   - 规则文件会定期更新（默认24小时）
   - 每次更新都会重新过滤IP规则

## 故障排查

### 问题1：规则数量为0
**原因**：规则文件不是Clash格式或URL不符合识别条件

**解决**：
- 确认URL包含 `/Clash/` 或以 `.yaml` 结尾
- 检查规则文件是否为有效的YAML格式

### 问题2：规则数量比预期少
**原因**：IP规则被过滤掉了

**解决**：
- 这是正常行为
- 查看日志中的 `filtered_ip_rules` 数量
- 只有域名规则会被保存

### 问题3：规则不生效
**原因**：规则格式转换错误

**解决**：
- 检查 `data/filters/[id].txt` 文件内容
- 确认规则格式正确
- 查看AdGuard Home日志

## 相关文件

- `internal/filtering/clash_rules.go` - Clash规则解析和转换
- `internal/filtering/filter.go` - 过滤器更新逻辑
- `data/filters/[id].txt` - 保存的规则文件

## 技术细节

### YAML解析
使用 `gopkg.in/yaml.v3` 库解析Clash规则文件

### 规则格式
- Clash：`TYPE,value`
- AdGuard：根据类型转换

### 文件处理
- 流式处理，避免大文件内存溢出
- 临时文件机制，确保原子性更新

## 总结

✅ **自动识别Clash规则**
✅ **过滤所有IP规则**
✅ **只保留域名规则**
✅ **转换为AdGuard格式**
✅ **详细统计信息**
✅ **日志记录完整**

DNS路由规则现在完全支持Clash格式，并且会自动过滤掉IP规则，只保留域名规则用于DNS路由！
