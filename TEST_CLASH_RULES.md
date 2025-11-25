# Clash规则处理测试指南

## 当前配置

配置文件中的DNS路由规则：
```yaml
dns_routing_filters:
  - enabled: true
    url: https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml
    name: CN
    id: 1764000553
    dns_routing_upstream_group: group_1763970331409
```

## 测试步骤

### 1. 停止AdGuard Home
```bash
# 在运行AdGuard Home的终端按 Ctrl+C
```

### 2. 删除旧的规则文件
```bash
Remove-Item "E:\Kiro\AdGuardHome-master\data\filters\1764000553.txt"
```

### 3. 启动AdGuard Home
```bash
.\AdGuardHome.exe
```

### 4. 观察日志输出

应该看到类似以下的日志：

```
[info] downloading update for filter id=1764000553 url=https://raw.githubusercontent.com/...
[debug] processing clash rule for dns routing id=1764000553 url=https://...
[info] clash rule processed id=1764000553 total_rules=5000 valid_domains=3728 filtered_ip_rules=1272
[info] filter updated id=1764000553 bytes_written=107960 rules_count=3728
```

关键点：
- ✅ 应该看到 "processing clash rule for dns routing"
- ✅ 应该看到 "clash rule processed" 和统计信息
- ✅ `filtered_ip_rules` 应该大于0，表示IP规则被过滤了

### 5. 检查规则文件内容

```bash
# 查看前20行
Get-Content "E:\Kiro\AdGuardHome-master\data\filters\1764000553.txt" | Select-Object -First 20

# 搜索IP规则（不应该有）
Get-Content "E:\Kiro\AdGuardHome-master\data\filters\1764000553.txt" | Select-String "IP-CIDR"
```

**预期结果**：
- ✅ 文件内容应该是纯文本格式，每行一个域名规则
- ✅ 应该看到类似 `baidu.com`、`||cn^`、`*china*` 这样的规则
- ✅ **不应该**看到 `payload:`、`- DOMAIN,`、`- IP-CIDR` 这样的YAML格式
- ✅ 搜索 `IP-CIDR` 应该返回0个结果

### 6. 验证规则格式

正确的格式示例：
```
baidu.com
qq.com
||cn^
||taobao.com^
*china*
*beijing*
```

错误的格式示例（不应该出现）：
```
payload:
- DOMAIN,baidu.com
- DOMAIN-SUFFIX,cn
- IP-CIDR,1.0.1.0/24
```

## 故障排查

### 问题1：日志中没有 "processing clash rule"

**可能原因**：
1. `dns_routing_upstream_group` 字段为空
2. URL不符合Clash规则识别条件

**解决方法**：
```bash
# 检查配置文件
Get-Content AdGuardHome.yaml | Select-String -Pattern "1764000553" -Context 5

# 确认有 dns_routing_upstream_group 字段
```

### 问题2：规则文件仍包含YAML格式

**可能原因**：
- Clash规则处理没有被触发
- 使用了旧版本的代码

**解决方法**：
```bash
# 重新编译
go build -o AdGuardHome.exe

# 删除旧规则文件
Remove-Item "data\filters\1764000553.txt"

# 重启服务
```

### 问题3：规则文件仍包含IP规则

**可能原因**：
- Clash规则处理逻辑有问题

**解决方法**：
- 检查日志中的 `filtered_ip_rules` 数量
- 如果为0，说明没有过滤IP规则
- 需要检查 `ProcessClashRuleFile` 函数

## 调试命令

### 查看规则文件统计
```powershell
# 总行数
(Get-Content "data\filters\1764000553.txt").Count

# 包含 IP-CIDR 的行数（应该为0）
(Get-Content "data\filters\1764000553.txt" | Select-String "IP-CIDR").Count

# 包含 DOMAIN 的行数（应该为0，因为已转换）
(Get-Content "data\filters\1764000553.txt" | Select-String "^- DOMAIN").Count

# 包含 || 的行数（DOMAIN-SUFFIX转换后的格式）
(Get-Content "data\filters\1764000553.txt" | Select-String "^\|\|").Count
```

### 查看规则文件样本
```powershell
# 前50行
Get-Content "data\filters\1764000553.txt" | Select-Object -First 50

# 随机20行
Get-Content "data\filters\1764000553.txt" | Get-Random -Count 20

# 最后20行
Get-Content "data\filters\1764000553.txt" | Select-Object -Last 20
```

## 成功标准

✅ **日志输出**：
- 看到 "processing clash rule for dns routing"
- 看到 "clash rule processed" 和统计信息
- `filtered_ip_rules` > 0

✅ **规则文件格式**：
- 纯文本格式，每行一个规则
- 没有YAML格式（payload:, - DOMAIN,等）
- 没有IP规则（IP-CIDR, IP-CIDR6）

✅ **规则数量**：
- `rules_count` 应该等于 `valid_domains`
- 应该明显少于原始文件的总规则数（因为过滤了IP规则）

## 预期结果

对于 China_Classical.yaml 规则文件：
- 原始规则数：约5000条
- 有效域名规则：约3728条
- 过滤的IP规则：约1272条
- 最终保存：3728条纯域名规则

## 下一步

测试成功后：
1. 在前端查看DNS路由页面，确认规则显示正确
2. 测试DNS解析，确认规则生效
3. 检查上游组是否正确应用
