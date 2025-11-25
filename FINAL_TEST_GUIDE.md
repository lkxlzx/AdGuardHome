# 最终测试指南

## 测试前准备

1. **备份配置**
   ```bash
   copy AdGuardHome.yaml AdGuardHome.yaml.backup
   ```

2. **停止旧版本**
   - 停止 AdGuard Home 服务

3. **替换文件**
   ```bash
   copy AdGuardHome.exe AdGuardHome.exe.old
   # 使用新编译的 AdGuardHome.exe
   ```

4. **启动新版本**
   ```bash
   AdGuardHome.exe
   ```

---

## 测试 1：白名单不影响 DNS 路由规则

### 步骤

1. **添加 DNS 路由规则**
   - 进入 "过滤器" → "DNS 路由规则"
   - 添加规则：
     - 名称：`测试规则`
     - URL：`https://example.com/rules.txt`
     - 上游组：选择任意组

2. **记录配置**
   - 打开 `AdGuardHome.yaml`
   - 找到 `dns_routing_filters` 部分
   - 记录 `upstream_group` 的值

3. **操作白名单**
   - 进入 "过滤器" → "DNS 白名单"
   - 添加规则：
     - 名称：`白名单测试`
     - URL：`https://example.com/allowlist.txt`

4. **验证结果**
   - 再次打开 `AdGuardHome.yaml`
   - 检查 `dns_routing_filters` 部分
   - ✅ 确认 `upstream_group` 字段仍然存在

### 预期结果

```yaml
dns_routing_filters:
  - enabled: true
    url: https://example.com/rules.txt
    name: 测试规则
    id: 1234567890
    upstream_group: group_xxx  # ✅ 此字段不应该丢失
```

---

## 测试 2：添加 Clash 规则自动处理

### 步骤

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

3. **等待下载**
   - 观察规则列表
   - 等待规则数量显示（可能需要几秒钟）

4. **检查规则文件**
   - 打开 `data/filters/` 目录
   - 找到对应 ID 的 `.txt` 文件
   - 用文本编辑器打开

5. **验证内容**
   - ✅ 文件中只包含域名规则
   - ✅ 格式类似：`google.com`, `||baidu.com^`, `*taobao*`
   - ❌ 不应该包含 IP 规则（如 `192.168.0.0/16`）

6. **查看日志**
   - 在 AdGuard Home 日志中搜索 "clash rule processed"
   - 应该看到类似信息：
     ```
     [INFO] clash rule processed id=xxx total_rules=1000 valid_domains=800 filtered_ip_rules=200
     ```

### 预期结果

**规则文件内容示例**：
```
google.com
||baidu.com^
||qq.com^
*taobao*
||jd.com^
||weibo.com^
```

**不应该包含**：
```
192.168.0.0/16
10.0.0.0/8
IP-CIDR,172.16.0.0/12
```

---

## 测试 3：重启后启用规则正常处理（关键测试）

### 步骤

1. **确认规则已添加**
   - 确保测试 2 中的 Clash 规则已成功添加
   - 规则文件只包含域名规则

2. **记录规则文件内容**
   - 打开 `data/filters/[id].txt`
   - 记录文件大小和内容摘要
   - 确认只有域名规则

3. **重启程序**
   - 停止 AdGuard Home
   - 再次启动 AdGuard Home

4. **禁用规则**
   - 进入 "过滤器" → "DNS 路由规则"
   - 找到 CN域名 规则
   - 点击开关，禁用该规则

5. **启用规则**
   - 再次点击开关，启用该规则
   - 等待更新完成

6. **验证规则文件**
   - 再次打开 `data/filters/[id].txt`
   - ✅ 确认文件仍然只包含域名规则
   - ✅ 确认没有 IP 规则
   - ✅ 文件大小应该与之前相近

7. **查看日志**
   - 在日志中搜索 "clash rule processed"
   - ✅ 应该看到处理信息

### 预期结果

- ✅ 规则文件内容不变（仍然只有域名规则）
- ✅ 没有 IP 规则被添加
- ✅ 日志显示 Clash 规则处理信息
- ✅ 规则数量保持一致

### 如果失败

如果规则文件包含了 IP 规则，说明修复未生效：
1. 检查是否使用了新编译的版本
2. 检查 `internal/home/home.go` 是否包含修复
3. 重新编译并替换文件

---

## 测试 4：完整流程测试

### 步骤

1. **清理环境**
   - 删除所有测试规则
   - 重启程序

2. **添加 Clash 规则**
   - 添加 Clash 规则 URL
   - 选择上游组
   - 保存

3. **添加白名单规则**
   - 添加一个白名单规则
   - 保存

4. **检查配置文件**
   - ✅ `dns_routing_filters` 的 `upstream_group` 存在
   - ✅ `whitelist_filters` 正常

5. **重启程序**
   - 停止并重启

6. **操作规则**
   - 禁用 DNS 路由规则
   - 启用 DNS 路由规则
   - 编辑白名单规则

7. **最终验证**
   - ✅ DNS 路由规则文件只包含域名规则
   - ✅ 配置文件中 `upstream_group` 仍然存在
   - ✅ 所有功能正常

---

## 常见问题排查

### Q1: 规则文件仍然包含 IP 规则

**可能原因**：
- 使用的不是新编译的版本
- URL 不符合 Clash 规则检测条件

**解决方法**：
1. 确认使用新编译的 `AdGuardHome.exe`
2. 确认 URL 包含 `/Clash/` 或 `/clash/`，或以 `.yaml`/`.yml` 结尾
3. 查看日志确认是否有 "clash rule processed" 信息

### Q2: upstream_group 字段丢失

**可能原因**：
- 使用的不是新编译的版本

**解决方法**：
1. 确认使用新编译的 `AdGuardHome.exe`
2. 检查 `internal/filtering/filtering.go` 是否包含修复

### Q3: 重启后启用规则失败

**可能原因**：
- `MarkAsDnsRouting` 方法未生效

**解决方法**：
1. 确认使用新编译的 `AdGuardHome.exe`
2. 检查 `internal/home/home.go` 和 `internal/filtering/filter.go` 是否包含修复
3. 查看日志获取详细错误信息

---

## 成功标准

所有以下条件都满足才算测试通过：

- ✅ 测试 1 通过：白名单操作不影响 DNS 路由规则
- ✅ 测试 2 通过：添加 Clash 规则自动处理
- ✅ 测试 3 通过：重启后启用规则正常处理
- ✅ 测试 4 通过：完整流程正常
- ✅ 配置文件正确保存
- ✅ 日志显示处理信息
- ✅ DNS 查询正常工作

---

## 测试报告模板

```
测试日期：____________________
测试人员：____________________

测试 1：白名单不影响 DNS 路由规则
结果：[ ] 通过  [ ] 失败
备注：_______________________

测试 2：添加 Clash 规则自动处理
结果：[ ] 通过  [ ] 失败
备注：_______________________

测试 3：重启后启用规则正常处理
结果：[ ] 通过  [ ] 失败
备注：_______________________

测试 4：完整流程测试
结果：[ ] 通过  [ ] 失败
备注：_______________________

总体评价：[ ] 全部通过  [ ] 部分失败  [ ] 全部失败

问题描述：
_________________________________
_________________________________
```

---

## 总结

完成所有测试后，如果全部通过，说明修复成功！可以正式部署使用。

如果有任何测试失败，请查看日志获取详细信息，并根据常见问题排查部分进行处理。
