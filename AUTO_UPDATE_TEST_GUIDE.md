# DNS 路由规则自动更新测试指南

## 📋 测试目的

验证 DNS 路由规则的分钟级自动更新功能是否正常工作。

---

## 🔧 测试准备

### 1. 确认版本
确保使用的是包含自动更新功能的 V2 版本：
```bash
# 检查可执行文件
AdGuardHome_v2.exe --version
```

### 2. 准备测试规则
使用一个会频繁更新的测试 URL，或者使用本地文件进行测试。

---

## 🧪 测试场景

### 场景 1: 1 分钟更新间隔

#### 步骤：
1. 添加 DNS 路由规则
   - 名称: 测试规则
   - URL: https://example.com/test-rules.yaml
   - 更新间隔: 1 分钟
   - 上游组: 选择任意组

2. 保存规则

3. 观察日志
   ```bash
   # 查看日志文件
   tail -f data/*.log
   
   # 或在 Windows 上
   Get-Content data\*.log -Wait -Tail 50
   ```

#### 期望结果：
- ✅ 规则立即下载并处理
- ✅ 1 分钟后自动触发更新
- ✅ 日志显示类似：
  ```
  [info] filtering: checking filters for updates
  [info] filtering: downloading update for filter id=xxx
  [info] filtering: clash rule processed id=xxx
  [info] filtering: filter updated id=xxx
  ```

---

### 场景 2: 5 分钟更新间隔

#### 步骤：
1. 编辑现有规则
2. 将更新间隔改为 5 分钟
3. 保存

#### 期望结果：
- ✅ 5 分钟后触发更新
- ✅ 不会在 1 分钟时更新

---

### 场景 3: 禁用自动更新（设置为 0）

#### 步骤：
1. 编辑规则
2. 将更新间隔设置为 0
3. 保存
4. 等待 5 分钟

#### 期望结果：
- ✅ 不会自动更新
- ✅ 只能通过"刷新规则"按钮手动更新

---

### 场景 4: 多个规则不同间隔

#### 步骤：
1. 添加规则 A，更新间隔 1 分钟
2. 添加规则 B，更新间隔 3 分钟
3. 添加规则 C，更新间隔 0（禁用）

#### 期望结果：
- ✅ 规则 A 每 1 分钟更新
- ✅ 规则 B 每 3 分钟更新
- ✅ 规则 C 不自动更新

---

## 📊 日志分析

### 正常更新日志示例

```
2025/11/25 13:10:00 [debug] filtering: starting update
2025/11/25 13:10:00 [debug] filtering: checking filters for updates
2025/11/25 13:10:00 [debug] filtering: downloading update for filter id=1764044678 url=https://...
2025/11/25 13:10:01 [debug] filtering: processing clash rule for domain routing id=1764044678
2025/11/25 13:10:01 [info] filtering: clash rule processed id=1764044678 total_rules=251 valid_domains=251 filtered_ip_rules=0
2025/11/25 13:10:01 [info] filtering: saving contents id=1764044678 path=data/filters/1764044678.txt
2025/11/25 13:10:01 [info] filtering: filter updated id=1764044678 rules_count=251
2025/11/25 13:10:01 [debug] filtering: finished update updated=1
```

### 跳过更新日志示例

```
2025/11/25 13:10:00 [debug] filtering: starting update
2025/11/25 13:10:00 [debug] filtering: checking filters for updates
2025/11/25 13:10:00 [debug] filtering: filter not due for update id=1764044678 last_updated=2025-11-25T13:09:00
2025/11/25 13:10:00 [debug] filtering: finished update updated=0
```

---

## 🔍 故障排查

### 问题 1: 规则不自动更新

**检查项**:
1. 确认更新间隔不是 0
   ```yaml
   # 检查 AdGuardHome.yaml
   dns_routing_filters:
     - name: 测试规则
       update_interval: 1  # 应该 > 0
   ```

2. 检查规则是否启用
   ```yaml
   dns_routing_filters:
     - enabled: true  # 必须是 true
   ```

3. 检查上次更新时间
   - 查看配置文件或日志
   - 计算是否已到更新时间

4. 检查日志级别
   ```yaml
   # 设置为 debug 查看详细日志
   log:
     verbose: true
   ```

---

### 问题 2: 更新太频繁

**可能原因**:
- 更新间隔设置过小
- 多个规则同时更新

**解决方法**:
1. 增加更新间隔
2. 错开不同规则的更新时间

---

### 问题 3: 更新失败

**检查项**:
1. 网络连接
   ```bash
   # 测试 URL 是否可访问
   curl -I https://example.com/rules.yaml
   ```

2. URL 有效性
   - 确认 URL 返回 200 状态码
   - 确认内容格式正确

3. 查看错误日志
   ```bash
   grep -i error data/*.log
   ```

---

## 📈 性能监控

### 监控指标

1. **更新频率**
   ```bash
   # 统计更新次数
   grep "filter updated" data/*.log | wc -l
   ```

2. **更新耗时**
   ```bash
   # 查看更新耗时
   grep "filter updated" data/*.log
   ```

3. **内存使用**
   ```bash
   # Windows
   Get-Process AdGuardHome_v2 | Select-Object WorkingSet
   
   # Linux
   ps aux | grep AdGuardHome_v2
   ```

4. **CPU 使用**
   - 更新时 CPU 会短暂升高（正常）
   - 平时应该很低

---

## ✅ 测试检查清单

### 功能测试
- [ ] 1 分钟间隔正常工作
- [ ] 5 分钟间隔正常工作
- [ ] 设置 0 禁用自动更新
- [ ] 多个规则独立更新
- [ ] 手动刷新仍然工作
- [ ] 编辑规则后间隔生效

### 边界测试
- [ ] 最小间隔（1 分钟）
- [ ] 大间隔（1440 分钟/24 小时）
- [ ] 禁用（0 分钟）
- [ ] 网络错误时的行为
- [ ] URL 无效时的行为

### 性能测试
- [ ] 内存使用正常
- [ ] CPU 使用正常
- [ ] 不影响 DNS 解析性能
- [ ] 多规则同时更新不卡顿

### 持久化测试
- [ ] 重启后间隔设置保留
- [ ] 配置文件正确保存
- [ ] 上次更新时间正确记录

---

## 🎯 预期行为总结

| 更新间隔 | 行为 | 日志频率 |
|---------|------|---------|
| 0 分钟 | 不自动更新 | 无自动更新日志 |
| 1 分钟 | 每分钟检查并更新 | 每分钟一次 |
| 5 分钟 | 每 5 分钟检查并更新 | 每 5 分钟一次 |
| 60 分钟 | 每小时检查并更新 | 每小时一次 |
| 1440 分钟 | 每天检查并更新 | 每天一次 |

---

## 📝 测试记录模板

```
测试日期: ____________________
测试人员: ____________________
版本号: V2 (Commit: a147ba22)

测试场景 1: 1 分钟更新
- 开始时间: __:__
- 第一次更新: __:__ (预期 +1 分钟)
- 第二次更新: __:__ (预期 +2 分钟)
- 结果: [ ] 通过 [ ] 失败
- 备注: ____________________

测试场景 2: 5 分钟更新
- 开始时间: __:__
- 第一次更新: __:__ (预期 +5 分钟)
- 结果: [ ] 通过 [ ] 失败
- 备注: ____________________

测试场景 3: 禁用自动更新
- 等待时间: 5 分钟
- 是否更新: [ ] 是 [ ] 否 (预期: 否)
- 结果: [ ] 通过 [ ] 失败
- 备注: ____________________

总体评价: ____________________
```

---

## 🔗 相关文档

- [功能说明](FEATURE_AUTO_UPDATE_INTERVAL.md)
- [V2 发布总结](V2_RELEASE_SUMMARY.md)
- [部署清单](DEPLOYMENT_CHECKLIST_V2.md)

---

**测试版本**: V2 (Commit: a147ba22)
**测试状态**: ⏳ 待测试
**最后更新**: 2024-11-25
