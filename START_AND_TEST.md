# 启动和测试 Prefetch UI 版本

## 编译信息

✅ **前端编译完成**：包含最新的 Prefetch UI 界面  
✅ **后端编译完成**：包含 Prefetch 配置 API  
✅ **可执行文件**：`AdGuardHome_prefetch_ui.exe` (约 39 MB)

## 快速启动

### 方法 1：直接启动（推荐）

```bash
.\AdGuardHome_prefetch_ui.exe
```

### 方法 2：指定配置文件

```bash
.\AdGuardHome_prefetch_ui.exe -c AdGuardHome.yaml
```

### 方法 3：调试模式

```bash
.\AdGuardHome_prefetch_ui.exe -v
```

## 访问 Web 界面

启动后，在浏览器中访问：

```
http://localhost:3000
```

或者使用服务器 IP：

```
http://YOUR_SERVER_IP:3000
```

## 测试 Prefetch UI

### 步骤 1：登录

使用你的管理员账号登录（如果是首次启动，会引导你创建账号）

### 步骤 2：进入 DNS 设置

1. 点击左侧菜单 **"设置"**
2. 点击 **"DNS 设置"**
3. 向下滚动找到 **"DNS 缓存配置"** 卡片

### 步骤 3：查看 Prefetch 设置

在 DNS 缓存配置卡片中，你应该看到：

```
DNS 缓存配置
├── ☑ 启用缓存
├── 缓存大小
├── 最小 TTL
├── 最大 TTL  
├── ☑ 乐观缓存
├── ─────────────────────
└── DNS 预取设置
    └── ☐ 启用 DNS 预取  ← 新增功能！
```

### 步骤 4：启用并配置 Prefetch

1. **勾选** "启用 DNS 预取"
2. 你会看到展开的配置选项：
   - 访问阈值（默认：5）
   - 时间窗口（默认：3600 秒）
   - 最大跟踪域名数（默认：10000）
   - 清理间隔（默认：3600 秒）
   - 软并发限制（默认：50）
   - 硬并发限制（默认：150）

3. **保持默认值**或根据需要调整

4. 点击 **"保存"** 按钮

5. 等待成功提示（约 1-2 秒）

### 步骤 5：验证配置生效

#### 方法 1：查看日志

在控制台中查看日志输出：

```
[info] prefetch manager initialized threshold=5 time_window=1h0m0s max_entries=10000 ...
```

#### 方法 2：刷新页面

刷新浏览器页面，确认配置值正确显示

#### 方法 3：查看配置文件

打开 `AdGuardHome.yaml`，确认包含：

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

## 功能测试

### 测试 1：基本功能

1. 启用 Prefetch
2. 使用 `nslookup` 查询一个域名 5 次以上：
   ```bash
   nslookup google.com 127.0.0.1
   nslookup google.com 127.0.0.1
   nslookup google.com 127.0.0.1
   nslookup google.com 127.0.0.1
   nslookup google.com 127.0.0.1
   ```
3. 查看日志，确认域名被标记为热门

### 测试 2：配置修改

1. 修改访问阈值为 3
2. 修改时间窗口为 7200 秒
3. 保存配置
4. 刷新页面确认修改生效

### 测试 3：表单验证

1. 设置软并发限制为 200
2. 设置硬并发限制为 100
3. 确认显示错误：**"软限制必须小于或等于硬限制"**
4. 确认保存按钮被禁用

### 测试 4：API 测试

使用 curl 测试 API：

```bash
# 获取配置
curl http://localhost:3000/control/dns_info

# 设置配置
curl -X POST http://localhost:3000/control/dns_info \
  -H "Content-Type: application/json" \
  -d '{"prefetch_enabled": true, "prefetch_threshold": 10}'
```

## 监控 Prefetch 运行

### 查看实时日志

日志会每分钟输出一次 Prefetch 指标：

```
[info] prefetch metrics 
  active_tasks=5 
  urgent_queue=2 
  normal_queue=10 
  tasks_completed=150 
  tasks_failed=0
  tracked_hits=500
  hot_domains=50
```

### 指标说明

- **active_tasks**: 当前正在执行的刷新任务数
- **urgent_queue**: 紧急队列中的任务数
- **normal_queue**: 普通队列中的任务数
- **tasks_completed**: 已完成的任务总数
- **tasks_failed**: 失败的任务总数
- **tracked_hits**: 正在跟踪的域名数
- **hot_domains**: 热门域名数（会被预取）

## 性能建议

### 家庭网络（默认配置）
```
访问阈值: 5
时间窗口: 3600 秒
最大跟踪域名数: 10000
软并发限制: 50
硬并发限制: 150
```

### 小型企业网络
```
访问阈值: 3
时间窗口: 7200 秒
最大跟踪域名数: 20000
软并发限制: 100
硬并发限制: 300
```

### 大型企业网络
```
访问阈值: 10
时间窗口: 3600 秒
最大跟踪域名数: 50000
软并发限制: 200
硬并发限制: 500
```

## 故障排查

### 问题 1：UI 不显示 Prefetch 设置

**原因**：浏览器缓存  
**解决**：
1. 按 `Ctrl + Shift + R` 强制刷新
2. 或清除浏览器缓存

### 问题 2：保存后配置不生效

**原因**：DNS 服务未重启  
**解决**：
1. 查看日志确认是否有错误
2. 手动重启 AdGuard Home

### 问题 3：日志中没有 Prefetch 信息

**原因**：Prefetch 未启用  
**解决**：
1. 确认 UI 中已勾选"启用 DNS 预取"
2. 确认配置已保存
3. 重启服务

## 停止服务

按 `Ctrl + C` 停止 AdGuard Home

## 下一步

- 📖 阅读 [PREFETCH_UI_INTEGRATION.md](./PREFETCH_UI_INTEGRATION.md) 了解详细配置
- 📊 阅读 [PREFETCH_CONFIG_GUIDE.md](./PREFETCH_CONFIG_GUIDE.md) 了解性能优化
- 🧪 阅读 [PREFETCH_UI_QUICK_TEST.md](./PREFETCH_UI_QUICK_TEST.md) 进行完整测试

## 反馈

如果遇到问题或有建议，请查看日志文件或配置文件进行排查。

---

**版本信息**：
- 编译时间：2025-11-26
- 包含功能：完整的 Prefetch UI 支持
- 前端版本：最新（包含 Prefetch 界面）
- 后端版本：最新（包含 Prefetch API）

🎉 **现在可以开始测试了！**
