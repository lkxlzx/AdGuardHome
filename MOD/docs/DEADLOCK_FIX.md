# 死锁问题修复

## 🐛 问题描述

**症状**: 
- 添加DNS上游分组后，前端卡死
- 服务器重启后前端不响应
- 无法保存配置

**根本原因**: 
在handler中调用`config.write()`之前先锁定了config，而`config.write()`内部也会锁定config，导致**死锁**。

```go
// 错误的代码
func (web *webAPI) handleAddUpstreamGroup(...) {
    config.Lock()           // 外层锁定
    defer config.Unlock()
    
    // ... 操作 ...
    
    config.write(...)       // 内部也会Lock()，导致死锁！
}
```

---

## ✅ 解决方案

### 原则

**在调用`config.write()`之前必须释放锁**，因为`config.write()`会自己处理锁定。

### 修复模式

```go
// 正确的代码
func (web *webAPI) handleAddUpstreamGroup(...) {
    // 1. 读取时使用RLock（如果只读）
    config.RLock()
    // ... 读取操作 ...
    config.RUnlock()
    
    // 2. 修改时使用Lock，但在调用write前释放
    config.Lock()
    // ... 修改操作 ...
    config.Unlock()
    
    // 3. config.write自己处理锁定
    config.write(...)
}
```

---

## 📝 修改的文件

### internal/home/dns_upstream_groups.go

修复了所有handler中的锁定逻辑：

#### 1. handleAddUpstreamGroup

**修改前**:
```go
config.Lock()
defer config.Unlock()
// ... 所有操作 ...
config.write(...)  // 死锁！
```

**修改后**:
```go
// 检查重复名称（只读）
config.RLock()
// ... 检查 ...
config.RUnlock()

// 修改配置
config.Lock()
// ... 修改 ...
config.Unlock()

// 保存配置
config.write(...)  // 正常
```

#### 2. handleUpdateUpstreamGroup

**修改前**:
```go
config.Lock()
defer config.Unlock()
// ... 所有操作 ...
config.write(...)  // 死锁！
```

**修改后**:
```go
config.Lock()
// ... 查找和修改 ...
config.Unlock()

config.write(...)  // 正常
```

#### 3. handleDeleteUpstreamGroup

**修改前**:
```go
config.Lock()
defer config.Unlock()
// ... 所有操作 ...
config.write(...)  // 死锁！
```

**修改后**:
```go
config.Lock()
// ... 查找和删除 ...
config.Unlock()

config.write(...)  // 正常
```

#### 4. handleSetDefaultGroup

**修改前**:
```go
config.Lock()
defer config.Unlock()
// ... 所有操作 ...
config.write(...)  // 死锁！
```

**修改后**:
```go
config.Lock()
// ... 查找和设置 ...
config.Unlock()

config.write(...)  // 正常
```

---

## 🎯 关键点

### 1. 理解config.write()的行为

`config.write()` 内部会：
```go
func (c *configuration) write(...) error {
    c.Lock()           // 自己会锁定
    defer c.Unlock()
    
    // ... 写入文件 ...
}
```

### 2. 避免嵌套锁定

```go
// ❌ 错误：嵌套锁定
config.Lock()
config.write()  // 内部也Lock()，死锁！

// ✅ 正确：分开锁定
config.Lock()
// ... 修改 ...
config.Unlock()
config.write()  // 独立锁定
```

### 3. 使用RLock进行只读操作

```go
// 只读操作使用RLock
config.RLock()
value := config.DNS.SomeValue
config.RUnlock()

// 写操作使用Lock
config.Lock()
config.DNS.SomeValue = newValue
config.Unlock()
```

### 4. 提前返回时记得解锁

```go
config.Lock()

if someCondition {
    config.Unlock()  // 记得解锁！
    return
}

// ... 继续操作 ...
config.Unlock()
```

---

## 🧪 测试验证

### 测试步骤

1. **启动服务**:
   ```bash
   .\AdGuardHome.exe
   ```

2. **添加分组**:
   - 登录系统
   - 进入DNS设置
   - 添加新的上游分组
   - 点击保存

3. **验证结果**:
   - ✅ 保存成功，没有卡死
   - ✅ 前端正常响应
   - ✅ 可以继续操作

4. **重启服务**:
   - 停止服务
   - 重新启动
   - ✅ 服务正常启动
   - ✅ 配置正确加载

### 预期结果

- ✅ 添加分组成功
- ✅ 前端不会卡死
- ✅ 配置正确保存到文件
- ✅ 重启后配置正确加载
- ✅ 所有操作（添加、编辑、删除、设置默认）都正常工作

---

## 📚 参考其他代码

查看项目中其他正确使用`config.write()`的例子：

### internal/home/home.go
```go
// 正确：没有在外层锁定
err = config.write(ctx, slogLogger, nil, nil, workDir, confPath)
```

### internal/home/config.go
```go
// defaultConfigModifier.Apply
func (cm *defaultConfigModifier) Apply(ctx context.Context) {
    // 正确：直接调用write，不预先锁定
    err := cm.config.write(ctx, cm.logger, ...)
}
```

---

## ✨ 总结

通过修复所有handler中的锁定逻辑，确保：

1. **在调用`config.write()`之前释放所有锁**
2. **只读操作使用`RLock()`**
3. **写操作使用`Lock()`，但在`write()`前释放**
4. **提前返回时记得解锁**

现在DNS上游分组功能可以正常保存配置，不会出现死锁或卡死的问题了！🎉

---

**修复日期**: 2024-12-04  
**状态**: ✅ 已修复并验证
