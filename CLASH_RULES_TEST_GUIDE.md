# Clash 规则处理测试指南

## 功能说明

当在域名分流规则页面添加 Clash 规则 URL 时，系统会自动：
1. 检测 URL 是否为 Clash 规则文件
2. 下载并解析 Clash 规则
3. 过滤掉 IP 规则（IP-CIDR, IP-CIDR6）
4. 只保留域名规则（DOMAIN, DOMAIN-SUFFIX, DOMAIN-KEYWORD）
5. 转换为 AdGuard Home 格式

## Clash 规则检测条件

URL 满足以下任一条件即被识别为 Clash 规则：
- 包含 `/Clash/` 或 `/clash/`
- 以 `.yaml` 结尾
- 以 `.yml` 结尾

## 规则转换示例

### Clash 格式 → AdGuard 格式

| Clash 规则类型 | Clash 示例 | AdGuard 格式 | 说明 |
|---------------|-----------|-------------|------|
| DOMAIN | `DOMAIN,google.com` | `google.com` | 精确匹配 |
| DOMAIN-SUFFIX | `DOMAIN-SUFFIX,google.com` | `\|\|google.com^` | 后缀匹配 |
| DOMAIN-KEYWORD | `DOMAIN-KEYWORD,google` | `*google*` | 关键词匹配 |
| IP-CIDR | `IP-CIDR,192.168.0.0/16` | ❌ 被过滤 | IP 规则不用于域名分流 |
| IP-CIDR6 | `IP-CIDR6,2001::/32` | ❌ 被过滤 | IPv6 规则不用于域名分流 |

## 测试步骤

### 1. 准备测试 URL

使用以下 Clash 规则 URL 进行测试：
```
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml
```

### 2. 添加 DNS 路由规则

1. 打开 AdGuard Home 管理界面
2. 进入 "过滤器" → "DNS 路由规则"
3. 点击 "添加过滤器"
4. 输入：
   - 名称：`CN域名`
   - URL：上述 Clash 规则 URL
   - 上游组：选择一个上游组（如 "国内DNS"）
5. 点击 "保存"

### 3. 查看日志验证

在 AdGuard Home 日志中应该看到类似信息：
```
[INFO] clash rule processed id=xxx total_rules=1000 valid_domains=800 filtered_ip_rules=200
```

这表示：
- 总共 1000 条规则
- 800 条有效域名规则被保留
- 200 条 IP 规则被过滤掉

### 4. 验证配置文件

检查 `AdGuardHome.yaml` 文件：
```yaml
dns_routing_filters:
  - enabled: true
    url: https://raw.githubusercontent.com/.../China_Classical.yaml
    name: CN域名
    id: 1764000553
    upstream_group: group_1763970331409  # ✅ upstream_group 字段存在
```

### 5. 验证规则文件

检查生成的规则文件（在 `data/filters/` 目录下）：
```bash
# 文件名类似：1764000553.txt
# 内容应该只包含域名规则，没有 IP 规则
```

示例内容：
```
google.com
||baidu.com^
*taobao*
```

### 6. 测试 DNS 查询

使用 `nslookup` 或 `dig` 测试：
```bash
# 测试匹配规则的域名
nslookup baidu.com 127.0.0.1

# 应该使用指定的上游组进行解析
```

## 常见问题

### Q1: Clash 规则没有被识别？
**检查**：
- URL 是否包含 `/Clash/` 或 `/clash/`
- URL 是否以 `.yaml` 或 `.yml` 结尾
- 查看日志是否有 "processing clash rule for domain routing" 信息

### Q2: 规则数量为 0？
**可能原因**：
- Clash 规则文件格式不正确
- 所有规则都是 IP 规则（被过滤掉了）
- 网络问题导致下载失败

**解决方法**：
- 检查 Clash 规则文件格式
- 查看 AdGuard Home 日志获取详细错误信息

### Q3: upstream_group 字段丢失？
**已修复**：
- 确保使用修复后的版本
- `WriteDiskConfig` 函数已添加 `DnsRoutingFilters` 复制

### Q4: 普通 URL 也被当作 Clash 规则处理？
**检查**：
- 确认 URL 不包含 `/Clash/` 或 `/clash/`
- 确认 URL 不以 `.yaml` 或 `.yml` 结尾
- 如果是普通规则列表，使用 `.txt` 扩展名

## 支持的 Clash 规则文件格式

### YAML 格式
```yaml
payload:
  - DOMAIN,google.com
  - DOMAIN-SUFFIX,baidu.com
  - DOMAIN-KEYWORD,taobao
  - IP-CIDR,192.168.0.0/16
```

### 纯文本格式（每行一条规则）
```
DOMAIN,google.com
DOMAIN-SUFFIX,baidu.com
DOMAIN-KEYWORD,taobao
IP-CIDR,192.168.0.0/16
```

## 性能说明

- Clash 规则处理在下载/更新过滤器时进行
- 处理后的规则以 AdGuard 格式存储在本地
- 后续 DNS 查询直接使用本地规则，无需重复处理
- IP 规则被过滤掉，减少内存占用和查询时间

## 日志级别

要查看详细的 Clash 规则处理日志，可以在配置中设置：
```yaml
log:
  verbose: true
```

然后重启 AdGuard Home，日志中会显示：
- 下载进度
- 规则解析详情
- 转换统计信息
- 错误信息（如果有）

## 总结

✅ Clash 规则自动检测和处理已完全实现
✅ IP 规则自动过滤
✅ 域名规则自动转换为 AdGuard 格式
✅ 与上游组正确关联
✅ 配置正确保存

只需在域名分流规则页面添加 Clash 规则 URL，系统会自动处理一切！
