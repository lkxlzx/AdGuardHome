# 快速测试步骤

## 修复内容

本次修复了 2 个文件，共 2 行代码：

1. **internal/filtering/filtering.go** - 修复配置保存时 DNS 路由规则丢失
2. **internal/filtering/http.go** - 修复添加 DNS 路由规则时 Clash 规则不处理

## 测试步骤

### 测试 1：白名单不影响 DNS 路由规则

1. **添加 DNS 路由规则**
   - 进入 "过滤器" → "DNS 路由规则"
   - 添加一个规则，选择上游组
   - 记录配置文件中的 `dns_routing_filters` 内容

2. **操作白名单**
   - 进入 "过滤器" → "DNS 白名单"
   - 添加一个白名单规则
   - 保存

3. **验证**
   - 打开 `AdGuardHome.yaml`
   - 检查 `dns_routing_filters` 部分
   - ✅ 确认 `upstream_group` 字段仍然存在

### 测试 2：Clash 规则自动处理

1. **准备 Clash 规则 URL**
   ```
   https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/China/China_Classical.yaml
   ```

2. **添加 DNS 路由规则**
   - 进入 "过滤器" → "DNS 路由规则"
   - 点击 "添加过滤器"
   - 输入：
     - 名称：`CN域名`
     - URL：上述 Clash 规则 URL
     - 上游组：选择一个（如 "国内DNS"）
   - 点击 "保存"

3. **等待下载完成**
   - 观察规则列表，等待规则数量显示

4. **检查规则文件**
   - 打开 `data/filters/` 目录
   - 找到对应 ID 的 `.txt` 文件
   - 打开文件查看内容

5. **验证结果**
   - ✅ 文件中只包含域名规则（如 `google.com`, `||baidu.com^`, `*taobao*`）
   - ❌ 文件中不应该包含 IP 规则（如 `192.168.0.0/16`）

6. **查看日志**
   - 在 AdGuard Home 日志中搜索 "clash rule processed"
   - 应该看到类似信息：
     ```
     [INFO] clash rule processed id=xxx total_rules=1000 valid_domains=800 filtered_ip_rules=200
     ```

## 预期结果

### 测试 1 预期结果
配置文件中 `dns_routing_filters` 保持完整：
```yaml
dns_routing_filters:
  - enabled: true
    url: https://example.com/rules.yaml
    name: 测试规则
    id: 1234567890
    upstream_group: group_1763970331409  # ✅ 此字段不会丢失
```

### 测试 2 预期结果

**规则文件内容示例** (`data/filters/1234567890.txt`):
```
google.com
||baidu.com^
||qq.com^
*taobao*
||jd.com^
```

**不应该包含的内容**:
```
192.168.0.0/16
10.0.0.0/8
IP-CIDR,172.16.0.0/12
```

## 如果测试失败

### 测试 1 失败（upstream_group 丢失）
- 确认使用的是修复后的版本
- 检查 `internal/filtering/filtering.go` 是否包含修复
- 重新编译：`go build -o AdGuardHome.exe`

### 测试 2 失败（Clash 规则未处理）
- 确认 URL 包含 `/Clash/` 或 `/clash/`，或以 `.yaml`/`.yml` 结尾
- 检查 `internal/filtering/http.go` 是否包含修复
- 查看日志获取详细错误信息
- 重新编译：`go build -o AdGuardHome.exe`

## 常见问题

**Q: 规则文件在哪里？**
A: `data/filters/[规则ID].txt`，规则 ID 可以在配置文件中查看

**Q: 如何查看日志？**
A: 
- 在 AdGuard Home 管理界面查看
- 或者查看控制台输出
- 或者查看日志文件（如果配置了）

**Q: Clash 规则没有被识别？**
A: 确保 URL 满足以下条件之一：
- 包含 `/Clash/` 或 `/clash/`
- 以 `.yaml` 结尾
- 以 `.yml` 结尾

**Q: 需要重启 AdGuard Home 吗？**
A: 是的，修复后需要：
1. 停止旧版本
2. 替换可执行文件
3. 启动新版本

## 总结

✅ 两个问题都已修复
✅ 编译成功
✅ 可以开始测试

如果测试通过，说明修复成功！
