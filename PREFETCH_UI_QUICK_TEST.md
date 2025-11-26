# Prefetch UI 快速测试指南

## 测试前准备

1. 确保已编译最新版本：
```bash
go build -o AdGuardHome.exe .
```

2. 启动 AdGuard Home：
```bash
./AdGuardHome.exe
```

3. 打开浏览器访问：`http://localhost:3000`

## 测试步骤

### 1. 验证配置读取

1. 登录 AdGuard Home
2. 进入 **设置 → DNS 设置**
3. 找到 **DNS 缓存配置** 卡片
4. 向下滚动到 **DNS 预取设置** 部分
5. 确认看到以下字段：
   - ✅ "启用 DNS 预取" 复选框
   - ✅ 所有输入框初始为空或显示默认值

### 2. 验证启用/禁用切换

1. 点击 **"启用 DNS 预取"** 复选框
2. 确认以下 6 个输入框显示出来：
   - ✅ 访问阈值
   - ✅ 时间窗口（秒）
   - ✅ 最大跟踪域名数
   - ✅ 清理间隔（秒）
   - ✅ 软并发限制
   - ✅ 硬并发限制
3. 再次点击复选框取消选中
4. 确认所有输入框隐藏

### 3. 验证默认值

1. 启用 DNS 预取
2. 确认各字段显示的占位符文本：
   - 访问阈值：`默认：5`
   - 时间窗口：`默认：3600（1 小时）`
   - 最大跟踪域名数：`默认：10000`
   - 清理间隔：`默认：3600（1 小时）`
   - 软并发限制：`默认：50`
   - 硬并发限制：`默认：150`

### 4. 验证配置保存

1. 启用 DNS 预取
2. 填写以下测试值：
   ```
   访问阈值：10
   时间窗口：7200
   最大跟踪域名数：20000
   清理间隔：1800
   软并发限制：100
   硬并发限制：300
   ```
3. 点击 **"保存"** 按钮
4. 等待成功提示
5. 刷新页面
6. 确认所有值正确保存并显示

### 5. 验证表单验证

#### 测试 1：软限制大于硬限制
1. 设置软并发限制：200
2. 设置硬并发限制：100
3. 确认显示错误提示：**"软限制必须小于或等于硬限制"**
4. 确认保存按钮被禁用

#### 测试 2：超出范围的值
1. 尝试在"访问阈值"输入 0（小于最小值 1）
2. 确认浏览器阻止输入或显示验证错误
3. 尝试输入 101（大于最大值 100）
4. 确认浏览器阻止输入或显示验证错误

### 6. 验证 API 端点

#### 测试 GET 端点
```bash
curl http://localhost:3000/control/dns_info
```

确认响应包含 Prefetch 字段：
```json
{
  "prefetch_enabled": true,
  "prefetch_threshold": 10,
  "prefetch_time_window": 7200,
  "prefetch_max_entries": 20000,
  "prefetch_cleanup_interval": 1800,
  "prefetch_soft_limit": 100,
  "prefetch_hard_limit": 300
}
```

#### 测试 POST 端点
```bash
curl -X POST http://localhost:3000/control/dns_config \
  -H "Content-Type: application/json" \
  -d '{
    "prefetch_enabled": true,
    "prefetch_threshold": 5,
    "prefetch_time_window": 3600,
    "prefetch_max_entries": 10000,
    "prefetch_cleanup_interval": 3600,
    "prefetch_soft_limit": 50,
    "prefetch_hard_limit": 150
  }'
```

### 7. 验证配置文件

1. 停止 AdGuard Home
2. 打开 `AdGuardHome.yaml`
3. 确认包含以下配置：
```yaml
dns:
  prefetch_enabled: true
  prefetch_threshold: 5
  prefetch_time_window: 1h
  prefetch_max_entries: 10000
  prefetch_cleanup_interval: 1h
  prefetch_soft_limit: 50
  prefetch_hard_limit: 150
```

### 8. 验证日志输出

1. 启动 AdGuard Home
2. 查看控制台日志
3. 确认看到初始化信息：
```
[info] prefetch manager initialized threshold=5 time_window=1h0m0s max_entries=10000 cleanup_interval=1h0m0s soft_limit=50 hard_limit=150
```

### 9. 验证国际化

#### 英文界面
1. 切换语言到 English
2. 确认所有标签和描述显示英文

#### 中文界面
1. 切换语言到简体中文
2. 确认所有标签和描述显示中文

### 10. 验证响应式设计

1. 调整浏览器窗口大小
2. 确认在不同屏幕尺寸下布局正常
3. 在移动设备上测试（如果可能）

## 预期结果

所有测试应该通过，具体表现为：

✅ 配置正确读取和显示  
✅ 启用/禁用切换正常工作  
✅ 默认值正确显示  
✅ 配置保存成功  
✅ 表单验证正常工作  
✅ API 端点返回正确数据  
✅ 配置文件正确更新  
✅ 日志输出正确  
✅ 国际化正常工作  
✅ 响应式设计正常  

## 故障排查

### 问题：配置不显示
- 检查浏览器控制台是否有 JavaScript 错误
- 确认 API 端点返回正确数据
- 清除浏览器缓存后重试

### 问题：保存失败
- 检查服务器日志
- 确认所有字段值在有效范围内
- 检查网络连接

### 问题：配置不生效
- 确认 DNS 服务已重启
- 检查配置文件是否正确更新
- 查看日志中的错误信息

## 性能测试（可选）

### 测试 Prefetch 功能
1. 启用 Prefetch，设置较低的阈值（如 3）
2. 使用 `nslookup` 或 `dig` 多次查询同一域名
3. 查看日志，确认域名被标记为热门
4. 等待缓存即将过期
5. 查看日志，确认自动刷新

### 监控指标
查看日志中的 Prefetch 指标（每分钟输出）：
```
[info] prefetch metrics active_tasks=X urgent_queue=Y normal_queue=Z ...
```

## 测试完成

完成所有测试后，确认：
- [ ] 所有 UI 功能正常
- [ ] 配置读写正确
- [ ] 验证逻辑有效
- [ ] API 端点工作正常
- [ ] 国际化支持完整
- [ ] 无明显 bug

如果所有测试通过，Prefetch UI 集成成功！🎉
