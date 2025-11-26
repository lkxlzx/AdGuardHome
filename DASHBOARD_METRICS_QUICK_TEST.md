# Dashboard Metrics 快速测试指南

## 快速启动

### 1. 编译前端
```bash
cd client
npm run build-prod
```

### 2. 编译后端
```bash
cd ..
go build
```

### 3. 启动服务
```bash
./AdGuardHome
```

## 测试步骤

### 测试 1：查看 Dashboard 卡片

1. 打开浏览器访问 `http://localhost:3000`
2. 登录 AdGuard Home
3. 进入 Dashboard 页面
4. 滚动到页面底部
5. 应该看到两个新卡片：
   - **Cache Metrics（缓存指标）**
   - **Prefetch Metrics（预取指标）**

### 测试 2：缓存禁用状态

**预期结果：**
- Cache Metrics 卡片显示："DNS cache is disabled"（DNS 缓存已禁用）

**启用缓存：**
1. 进入 **设置 → DNS 设置**
2. 找到 **DNS 缓存配置** 卡片
3. 勾选 **启用缓存**
4. 点击 **保存**
5. 返回 Dashboard
6. Cache Metrics 卡片应显示：
   - Cache size: 4 MB
   - Min TTL: 0s
   - Max TTL: 0s
   - Optimistic cache: Disabled

### 测试 3：预取禁用状态

**预期结果：**
- Prefetch Metrics 卡片显示："DNS prefetch is disabled"（DNS 预取已禁用）

**启用预取：**
1. 进入 **设置 → DNS 设置**
2. 找到 **DNS 缓存配置** 卡片
3. 展开 **DNS 预取设置**
4. 勾选 **启用 DNS 预取**
5. 点击 **保存**
6. 返回 Dashboard
7. Prefetch Metrics 卡片应显示：
   - Hot domains: 0
   - Completed: 0
   - Failed: 0
   - Success rate: 100.0%

### 测试 4：API 端点测试

#### 测试 API 响应
```bash
curl http://localhost:3000/control/dashboard_metrics
```

**预期响应：**
```json
{
  "cache_enabled": true,
  "cache_size": 4194304,
  "cache_ttl_min": 0,
  "cache_ttl_max": 0,
  "cache_optimistic": false,
  "prefetch_enabled": true,
  "prefetch_hot_domains": 0,
  "prefetch_completed": 0,
  "prefetch_failed": 0,
  "prefetch_success_rate": 100
}
```

### 测试 5：刷新按钮

1. 在 Dashboard 页面
2. 点击 Cache Metrics 卡片右上角的刷新按钮
3. 点击 Prefetch Metrics 卡片右上角的刷新按钮
4. 数据应该重新加载（如果有变化会更新）

### 测试 6：工具提示

1. 将鼠标悬停在每个指标旁边的 ⓘ 图标上
2. 应该显示详细的说明文字
3. 测试所有 8 个工具提示：
   - Cache size hint
   - Min TTL hint
   - Max TTL hint
   - Optimistic cache hint
   - Hot domains hint
   - Completed hint
   - Failed hint
   - Success rate hint

### 测试 7：国际化

#### 测试英文界面
1. 在设置中切换语言为 English
2. 返回 Dashboard
3. 确认卡片标题和内容都是英文

#### 测试中文界面
1. 在设置中切换语言为简体中文
2. 返回 Dashboard
3. 确认卡片标题和内容都是中文

### 测试 8：预取功能运行

**生成一些 DNS 查询：**
```bash
# 多次查询同一个域名
for i in {1..10}; do
  nslookup google.com 127.0.0.1
  sleep 1
done
```

**等待几分钟后：**
1. 刷新 Dashboard
2. Prefetch Metrics 应该显示：
   - Hot domains > 0
   - Completed > 0
   - Success rate 接近 100%

### 测试 9：响应式布局

1. 调整浏览器窗口大小
2. 在桌面宽度：卡片应该并排显示（2 列）
3. 在移动宽度：卡片应该堆叠显示（1 列）

### 测试 10：错误处理

#### 模拟 API 错误
1. 停止 AdGuard Home 服务
2. 刷新 Dashboard 页面
3. 打开浏览器控制台
4. 应该看到错误信息："Failed to fetch cache metrics"
5. 页面不应该崩溃，其他卡片应该正常显示

## 验证清单

- [ ] Dashboard 页面显示两个新卡片
- [ ] 缓存禁用时显示正确提示
- [ ] 预取禁用时显示正确提示
- [ ] 启用缓存后显示正确指标
- [ ] 启用预取后显示正确指标
- [ ] API 端点返回正确数据
- [ ] 刷新按钮正常工作
- [ ] 所有工具提示正确显示
- [ ] 英文界面显示正确
- [ ] 中文界面显示正确
- [ ] 预取功能运行后指标更新
- [ ] 响应式布局正常
- [ ] 错误处理正常

## 常见问题

### Q1: 卡片不显示
**A:** 检查前端是否正确编译，确认 `build` 目录存在

### Q2: API 返回 404
**A:** 确认后端已重新编译，API 端点已注册

### Q3: 翻译不显示
**A:** 检查 `__locales/en.json` 和 `__locales/zh-cn.json` 是否正确更新

### Q4: 预取指标始终为 0
**A:** 
1. 确认预取功能已启用
2. 生成一些 DNS 查询
3. 等待几分钟让预取功能运行
4. 检查日志确认预取正在工作

### Q5: 成功率显示 NaN
**A:** 这是正常的，当还没有任何预取操作时，默认显示 100.0%

## 性能基准

### 预期性能指标

**正常运行状态：**
- Hot domains: 50-200（取决于访问模式）
- Completed: 持续增长
- Failed: < 5%
- Success rate: > 95%

**高负载状态：**
- Hot domains: 200-500
- Completed: 快速增长
- Failed: 5-10%
- Success rate: > 90%

## 调试技巧

### 1. 查看浏览器控制台
```javascript
// 打开控制台，查看 API 请求
// Network 标签 → 找到 dashboard_metrics 请求
// 查看响应数据
```

### 2. 查看后端日志
```bash
# 启动时带调试日志
./AdGuardHome -v
```

### 3. 手动测试 API
```bash
# 测试 dashboard_metrics
curl -v http://localhost:3000/control/dashboard_metrics

# 测试 prefetch_status（更详细的预取信息）
curl -v http://localhost:3000/control/prefetch_status
```

## 下一步

测试通过后，可以：
1. 提交代码到版本控制
2. 更新用户文档
3. 准备发布说明
4. 考虑添加更多高级功能（图表、实时更新等）

## 相关文档

- [DASHBOARD_METRICS_UI_DESIGN.md](./DASHBOARD_METRICS_UI_DESIGN.md) - 详细设计文档
- [PREFETCH_UI_INTEGRATION.md](./PREFETCH_UI_INTEGRATION.md) - Prefetch UI 集成指南
- [PREFETCH_CONFIG_GUIDE.md](./PREFETCH_CONFIG_GUIDE.md) - Prefetch 配置指南
