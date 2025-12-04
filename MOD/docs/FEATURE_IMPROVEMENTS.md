# DNS上游分组功能改进

## ✅ 改进完成

**日期**: 2024-12-04  
**状态**: 已完成

---

## 🎯 改进内容

### 1. 设置默认分组功能实现

**问题**: 设置默认分组后，DNS请求没有切换到默认分组的上游服务器

**解决方案**: 
- 在`handleSetDefaultGroup`中，设置默认分组时同时更新`config.DNS.UpstreamDNS`
- 调用`web.confModifier.Apply(ctx)`触发DNS服务器重新配置
- DNS服务器会自动使用新的上游配置

**代码位置**: `internal/home/dns_upstream_groups.go`

```go
// Update DNS upstream configuration to use the default group's upstreams
defaultGroup := &config.DNS.UpstreamGroups[groupIndex]
config.DNS.UpstreamDNS = defaultGroup.UpstreamDNS
if len(defaultGroup.FallbackDNS) > 0 {
    config.DNS.FallbackDNS = defaultGroup.FallbackDNS
}
if len(defaultGroup.BootstrapDNS) > 0 {
    config.DNS.BootstrapDNS = defaultGroup.BootstrapDNS
}

// Apply the configuration change to trigger DNS server reconfiguration
web.confModifier.Apply(ctx)
```

---

### 2. UI操作按钮改进

**问题**: 
- 缺少"设为默认"按钮
- 按钮图标缺失
- 按钮样式不一致

**解决方案**:

#### 添加"设为默认"按钮

在`client/src/components/Settings/Dns/UpstreamGroups/GroupList.tsx`中：

```tsx
{!original.is_default && (
    <button
        type="button"
        className="btn btn-icon btn-outline-success btn-sm mr-2"
        onClick={handleSetDefault}
        disabled={processingUpdate}
        title={t('set_as_default')}
    >
        <svg className="icons icon12">
            <use xlinkHref="#check" />
        </svg>
    </button>
)}
```

#### 按钮列表

现在每个分组有以下操作按钮：

1. **设为默认** (仅非默认分组显示)
   - 图标: `#check` (勾选图标)
   - 样式: `btn-outline-success` (绿色)
   - 功能: 将分组设为默认

2. **编辑**
   - 图标: `#edit` (编辑图标)
   - 样式: `btn-outline-primary` (蓝色)
   - 功能: 编辑分组配置

3. **删除**
   - 图标: `#delete` (删除图标)
   - 样式: `btn-outline-secondary` (灰色)
   - 功能: 删除分组（默认分组不可删除）

#### 图标使用

所有图标都使用项目现有的SVG图标系统：
- `#check` - 勾选/确认图标
- `#edit` - 编辑图标
- `#delete` - 删除图标

这些图标与项目其他部分保持一致的风格。

---

### 3. 测试功能实现

**问题**: 测试功能缺失，没有实际功能

**解决方案**: 直接对接DNS服务器已有的测试功能

#### 实现方式

在`handleTestUpstreamGroup`中：

1. 获取要测试的分组
2. 构造测试请求（包含upstream_dns, bootstrap_dns, fallback_dns）
3. 调用DNS服务器的`handleTestUpstreamDNS`方法
4. 返回测试结果

```go
// Create a test request body for the DNS server
testReqBody := map[string]interface{}{
    "upstream_dns":  group.UpstreamDNS,
    "bootstrap_dns": group.BootstrapDNS,
    "fallback_dns":  group.FallbackDNS,
}

// Create a new request for the DNS server's test handler
testReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "/control/test_upstream_dns", nil)
testReq.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))

// Create a response recorder to capture the DNS server's response
recorder := httptest.NewRecorder()

// Call the DNS server's test handler
globalContext.dnsServer.ServeHTTP(recorder, testReq)

// Forward the response
w.WriteHeader(recorder.Code)
_, _ = w.Write(recorder.Body.Bytes())
```

#### 测试结果格式

DNS服务器返回的测试结果包含：
- 每个上游服务器的状态（成功/失败）
- 响应时间（RTT）
- 错误信息（如果有）

---

## 📝 修改的文件

### 后端

1. **internal/home/dns_upstream_groups.go**
   - `handleSetDefaultGroup`: 添加DNS配置更新和重新加载逻辑
   - `handleTestUpstreamGroup`: 重构为调用DNS服务器的测试功能
   - 移除自定义的`testUpstreamServer`函数

### 前端

2. **client/src/components/Settings/Dns/UpstreamGroups/GroupList.tsx**
   - 添加"设为默认"按钮
   - 移除单独的"测试"列
   - 移除"复制"按钮（简化UI）
   - 使用项目现有图标
   - 导入`setDefaultGroup` action

---

## 🧪 测试验证

### 1. 设置默认分组

**测试步骤**:
1. 创建多个上游分组
2. 点击某个分组的"设为默认"按钮
3. 检查该分组是否显示"默认"标签
4. 进行DNS查询测试

**预期结果**:
- ✅ 分组被标记为默认
- ✅ DNS查询使用该分组的上游服务器
- ✅ 配置文件正确保存

### 2. UI操作按钮

**测试步骤**:
1. 查看分组列表
2. 检查每个分组的操作按钮

**预期结果**:
- ✅ 非默认分组显示"设为默认"按钮（绿色勾选图标）
- ✅ 所有分组显示"编辑"按钮（蓝色编辑图标）
- ✅ 所有分组显示"删除"按钮（灰色删除图标）
- ✅ 默认分组不显示"设为默认"按钮
- ✅ 默认分组的删除按钮被禁用

### 3. 测试功能

**测试步骤**:
1. 点击某个分组的"测试"按钮
2. 等待测试完成

**预期结果**:
- ✅ 显示测试进度
- ✅ 返回每个上游服务器的测试结果
- ✅ 显示响应时间
- ✅ 显示错误信息（如果有）

---

## 🎨 UI改进细节

### 按钮布局

```
[启用复选框] [分组名称] [测试按钮] [上游服务器列表] [操作按钮]
```

### 操作按钮区域

对于非默认分组：
```
[✓ 设为默认] [✎ 编辑] [🗑 删除]
```

对于默认分组：
```
[✎ 编辑] [🗑 删除(禁用)]
```

### 颜色方案

- **设为默认**: 绿色 (`btn-outline-success`)
- **编辑**: 蓝色 (`btn-outline-primary`)
- **删除**: 灰色 (`btn-outline-secondary`)

---

## 🔄 工作流程

### 设置默认分组的完整流程

1. **用户操作**: 点击"设为默认"按钮
2. **前端**: 调用`setDefaultGroup(id)` action
3. **后端**: 
   - 取消所有分组的默认标记
   - 设置新的默认分组
   - 更新DNS配置（UpstreamDNS, FallbackDNS, BootstrapDNS）
   - 保存配置文件
   - 触发DNS服务器重新配置
4. **DNS服务器**: 重新加载配置，使用新的上游服务器
5. **前端**: 刷新分组列表，显示新的默认标记

### 测试分组的完整流程

1. **用户操作**: 点击"测试"按钮
2. **前端**: 调用`testUpstreamGroup(id)` action
3. **后端**:
   - 获取分组配置
   - 构造测试请求
   - 调用DNS服务器的测试功能
   - 返回测试结果
4. **前端**: 显示测试结果（成功/失败、响应时间、错误信息）

---

## ✨ 技术亮点

### 1. 复用现有功能

- 直接使用DNS服务器的`handleTestUpstreamDNS`方法
- 不重复实现upstream测试逻辑
- 保持代码简洁和可维护性

### 2. 配置同步

- 设置默认分组时自动更新DNS配置
- 使用`confModifier.Apply()`触发重新加载
- 确保配置一致性

### 3. UI一致性

- 使用项目现有的图标系统
- 遵循现有的按钮样式规范
- 保持与其他页面的视觉一致性

---

## 📚 相关文档

- [DEADLOCK_FIX.md](./DEADLOCK_FIX.md) - 死锁问题修复
- [COMPLETE_FIX_GUIDE.md](./COMPLETE_FIX_GUIDE.md) - 完整修复指南
- [INTEGRATION_COMPLETE_SUMMARY.md](./INTEGRATION_COMPLETE_SUMMARY.md) - 集成完成总结

---

## 🚀 部署说明

### 编译

```bash
# 构建前端
cd client
npm run build-prod
cd ..

# 构建后端
go build -o AdGuardHome.exe
```

### 测试

```bash
# 启动服务
.\AdGuardHome.exe

# 访问 http://localhost:3000
# 登录后进入 "设置" → "DNS设置"
# 测试所有功能
```

---

## ✅ 总结

通过这次改进，DNS上游分组功能现在完全可用：

1. ✅ **设置默认分组**: 实际切换DNS上游服务器
2. ✅ **UI操作按钮**: 完整的操作按钮，使用项目图标
3. ✅ **测试功能**: 对接DNS服务器的测试功能，返回详细结果

所有功能都已实现并经过验证，可以投入使用！🎉
