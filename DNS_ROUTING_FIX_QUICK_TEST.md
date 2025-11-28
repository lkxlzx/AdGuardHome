# DNS 路由筛选器修复 - 快速测试指南

## 问题描述

在查询日志页面选择"DNS 路由"筛选器时出现错误：
```
Error: controlQuerylog/
search: "response_status=dns_routing"
Loading params: invalid value
dns_routing | #40
```

## 修复版本

已编译的修复版本：`AdGuardHome_dns_routing_fix.exe`

## 快速测试步骤

### 1. 停止当前运行的 AdGuard Home

如果 AdGuard Home 正在运行，先停止它。

### 2. 使用修复版本启动

```powershell
.\AdGuardHome_dns_routing_fix.exe
```

### 3. 测试 Web 界面

1. 打开浏览器访问：http://localhost:3000
2. 登录管理界面
3. 点击左侧菜单的"查询日志"
4. 在页面右上角的筛选器下拉菜单中选择"DNS 路由"
5. **预期结果**：不再出现错误，页面正常显示 DNS 路由相关的查询记录

### 4. 使用 API 测试（可选）

运行测试脚本：

```powershell
.\test_dns_routing_filter.ps1
```

**预期输出**：
```
=== Testing DNS Routing Filter ===

Testing query log filter with response_status=dns_routing...
✓ DNS routing filter request successful!
  Found X DNS routing entries
```

## 验证修复

修复成功的标志：
- ✓ 筛选器下拉菜单中可以选择"DNS 路由"
- ✓ 选择后不再出现错误提示
- ✓ 页面正常显示筛选结果
- ✓ API 请求返回成功（200 状态码）

## 技术细节

修复内容：
- 在后端 `internal/querylog/searchcriterion.go` 中添加了 `dns_routing` 支持
- 添加了 `filteringStatusDNSRouting` 常量
- 实现了对 `filtering.NotFilteredDNSRouting` 的筛选逻辑

## 如果遇到问题

1. 确认使用的是修复版本：
   ```powershell
   .\AdGuardHome_dns_routing_fix.exe --version
   ```
   应该显示：`AdGuard Home, version v0.107.0-dns-routing-fix`

2. 检查配置文件是否正确：
   ```powershell
   .\AdGuardHome_dns_routing_fix.exe --check-config
   ```

3. 查看日志输出，确认没有其他错误

## 相关文件

- `AdGuardHome_dns_routing_fix.exe` - 修复后的可执行文件
- `build-dns-routing-fix.ps1` - 构建脚本
- `test_dns_routing_filter.ps1` - API 测试脚本
- `DNS_ROUTING_FILTER_FIX.md` - 详细修复文档
