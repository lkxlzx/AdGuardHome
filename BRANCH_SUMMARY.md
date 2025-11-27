# AdGuard Home 分支总结

## 分支概览

本项目目前维护多个开发分支，每个分支都有特定的功能和改进。

### 分支列表

| 分支 | 状态 | 说明 | 最后更新 |
|------|------|------|----------|
| master | 稳定 | 主分支，基于官方版本 | - |
| v1 | 归档 | 第一版改进 | - |
| v2 | 归档 | 第二版改进 | - |
| v3 | 归档 | 第三版改进 | 2025-11-26 |
| v3-backup-20251126 | 备份 | V3分支备份 | 2025-11-26 |
| v4 | ✓ 活跃 | DNS路由修复 + 性能优化 | 2025-11-27 |
| v5 | ✓ 活跃 | 基于V4的新开发分支 | 2025-11-27 |

## 分支详情

### Master 分支
- **用途：** 主分支，保持与官方同步
- **特点：** 稳定、官方版本
- **更新：** 跟随官方发布

### V1 分支
- **用途：** 第一版功能改进
- **状态：** 已归档
- **特点：** 早期实验性功能

### V2 分支
- **用途：** 第二版功能改进
- **状态：** 已归档
- **特点：** 功能迭代

### V3 分支
- **用途：** 第三版功能改进
- **状态：** 已归档
- **特点：** 
  - Dashboard V2 改进
  - 缓存功能优化
  - 预取功能初步实现
- **备份：** v3-backup-20251126

### V4 分支 ⭐
- **用途：** DNS路由修复 + 多项性能优化
- **状态：** ✓ 活跃开发
- **创建时间：** 2025-11-27
- **提交哈希：** 14e51f3b

#### 主要功能

1. **DNS路由筛选器修复** ✓
   - 修复查询日志页面的筛选器错误
   - 支持 `dns_routing` 筛选状态
   - 文件：`internal/querylog/searchcriterion.go`

2. **预取功能优化** ✓
   - 修复预取架构问题
   - 提高缓存命中率
   - 完整的测试工具
   - 文件：`internal/dnsforward/prefetch.go`

3. **LRU缓存优化** ✓
   - 智能LRU清理机制
   - 优化内存使用
   - 性能提升
   - 文件：`internal/dnsforward/dnsforward.go`

4. **统计功能改进** ✓
   - 准确的缓存统计
   - 优化数据收集
   - 文件：`internal/dnsforward/stats.go`

5. **多平台构建** ✓
   - 12个平台的DNS路由修复版本
   - 7个平台的LRU优化版本
   - 完整的构建脚本

#### 文件统计
- 修改文件：6个核心文件
- 新增测试脚本：20+
- 新增文档：25+
- 新增二进制文件：19个
- 总计：69个文件，+7,517行，-76行

#### 相关文档
- `V4_COMMIT_SUMMARY.md` - V4提交总结
- `DNS_ROUTING_FIX_COMPLETE.md` - DNS路由修复完整文档
- `LRU_OPTIMIZATION_FINAL_SUMMARY.md` - LRU优化总结
- `PREFETCH_OPTIMIZATION_COMPLETE.md` - 预取优化文档

### V5 分支 ⭐
- **用途：** 基于V4的新开发分支
- **状态：** ✓ 活跃开发
- **创建时间：** 2025-11-27
- **基于：** V4 分支（14e51f3b）
- **提交哈希：** 7d4ee414

#### 继承功能
V5 完全继承 V4 的所有功能和改进：
- ✓ DNS路由筛选器修复
- ✓ 预取功能优化
- ✓ LRU缓存优化
- ✓ 统计功能改进
- ✓ 多平台支持
- ✓ 完整文档

#### 开发目标
1. **性能优化**
   - 进一步优化DNS查询性能
   - 改进缓存命中率
   - 优化内存使用

2. **功能增强**
   - 新增功能特性
   - 改进用户体验
   - 增强稳定性

3. **代码质量**
   - 代码重构
   - 增加测试覆盖率
   - 改进错误处理

#### 相关文档
- `V5_BRANCH_INFO.md` - V5分支详细信息

## 分支关系图

```
master (官方版本)
  │
  ├─ v1 (归档)
  │
  ├─ v2 (归档)
  │
  ├─ v3 (归档)
  │   └─ v3-backup-20251126 (备份)
  │
  ├─ v4 (活跃) ⭐
  │   │
  │   └─ v5 (活跃) ⭐
  │
  └─ ...
```

## 版本对比

| 特性 | V3 | V4 | V5 |
|------|----|----|-----|
| Dashboard V2 | ✓ | ✓ | ✓ |
| 缓存优化 | 部分 | ✓ | ✓ |
| DNS路由筛选器 | ✗ | ✓ | ✓ |
| 预取优化 | 部分 | ✓ | ✓ |
| LRU优化 | ✗ | ✓ | ✓ |
| 统计改进 | 部分 | ✓ | ✓ |
| 多平台构建 | 部分 | ✓ | ✓ |
| 完整文档 | 部分 | ✓ | ✓ |
| 新功能开发 | - | - | 进行中 |

## 推荐使用

### 生产环境
- **推荐：** V4 分支
- **原因：** 稳定、经过测试、功能完整
- **版本：** v0.107.0-dns-routing-fix

### 开发测试
- **推荐：** V5 分支
- **原因：** 最新功能、持续开发
- **注意：** 可能包含未完全测试的功能

### 稳定保守
- **推荐：** Master 分支
- **原因：** 官方版本、最稳定
- **注意：** 缺少自定义优化

## 获取代码

### 克隆仓库
```bash
git clone https://github.com/lkxlzx/AdGuardHome.git
cd AdGuardHome
```

### 切换到特定分支
```bash
# V4 分支（推荐）
git checkout v4

# V5 分支（开发中）
git checkout v5

# Master 分支（官方）
git checkout master
```

### 拉取最新更新
```bash
git pull origin v4
# 或
git pull origin v5
```

## 构建说明

### V4 分支
```bash
# 切换到 V4
git checkout v4

# 构建 DNS 路由修复版本（单平台）
.\build-dns-routing-fix.ps1

# 构建所有平台
.\build-dns-routing-fix-all-platforms.ps1
```

### V5 分支
```bash
# 切换到 V5
git checkout v5

# 使用相同的构建脚本
.\build-dns-routing-fix.ps1
```

## 二进制文件

### V4 预编译版本
位置：`dist_dns_routing_fix/`

**Windows:**
- `AdGuardHome_windows_amd64.exe` (31 MB)
- `AdGuardHome_windows_386.exe` (29.85 MB)
- `AdGuardHome_windows_arm64.exe` (29 MB)

**Linux:**
- `AdGuardHome_linux_amd64` (32.26 MB)
- `AdGuardHome_linux_386` (31.03 MB)
- `AdGuardHome_linux_arm64` (30.44 MB)
- `AdGuardHome_linux_armv7` (30.75 MB)
- `AdGuardHome_linux_armv6` (30.81 MB)

**macOS:**
- `AdGuardHome_darwin_amd64` (32.36 MB)
- `AdGuardHome_darwin_arm64` (30.78 MB)

**FreeBSD:**
- `AdGuardHome_freebsd_amd64` (31.65 MB)
- `AdGuardHome_freebsd_arm64` (29.88 MB)

### LRU 优化版本
位置：`dist_lru_optimized/`
- 7个平台的LRU优化版本

## 文档资源

### 核心文档
- `README.md` - 项目说明
- `BRANCH_SUMMARY.md` - 本文档

### V4 文档
- `V4_COMMIT_SUMMARY.md` - V4提交总结
- `DNS_ROUTING_FIX_COMPLETE.md` - DNS路由修复
- `LRU_OPTIMIZATION_FINAL_SUMMARY.md` - LRU优化
- `PREFETCH_OPTIMIZATION_COMPLETE.md` - 预取优化
- `BUILD_REPORT_FINAL.md` - 构建报告

### V5 文档
- `V5_BRANCH_INFO.md` - V5分支信息

### 技术文档
- `DNS_ROUTING_FILTER_FIX.md` - DNS路由筛选器修复详解
- `LRU_CLEANUP_IMPLEMENTATION.md` - LRU清理实现
- `PREFETCH_ARCHITECTURE_FIX.md` - 预取架构修复

### 测试文档
- `DNS_ROUTING_FIX_QUICK_TEST.md` - 快速测试指南
- `LRU_PERFORMANCE_TEST_FINAL.md` - LRU性能测试
- `PREFETCH_TEST_GUIDE.md` - 预取测试指南

## 贡献指南

### 提交到 V4
```bash
git checkout v4
# 进行修改
git add .
git commit -m "修复：描述"
git push origin v4
```

### 提交到 V5
```bash
git checkout v5
# 进行修改
git add .
git commit -m "功能：描述"
git push origin v5
```

### 提交规范
- 使用中文提交信息
- 使用类型前缀：功能、修复、优化、文档、测试、构建
- 提供详细的变更说明
- 列出影响的文件

## 远程仓库

**GitHub：** https://github.com/lkxlzx/AdGuardHome

### 分支链接
- **V4：** https://github.com/lkxlzx/AdGuardHome/tree/v4
- **V5：** https://github.com/lkxlzx/AdGuardHome/tree/v5

### 创建 Pull Request
- **V4：** https://github.com/lkxlzx/AdGuardHome/pull/new/v4
- **V5：** https://github.com/lkxlzx/AdGuardHome/pull/new/v5

## 常见问题

### Q: 应该使用哪个分支？
**A:** 
- 生产环境：V4（稳定、功能完整）
- 开发测试：V5（最新功能）
- 保守使用：Master（官方版本）

### Q: V4 和 V5 有什么区别？
**A:** V5 基于 V4，继承所有功能，是未来新功能的开发分支。

### Q: 如何升级？
**A:** 
1. 备份配置文件
2. 下载新版本二进制文件
3. 替换旧文件
4. 重启服务

### Q: 是否兼容官方版本？
**A:** 是的，配置文件完全兼容，可以随时切换。

## 更新日志

### 2025-11-27
- ✓ 创建 V4 分支
- ✓ 修复 DNS 路由筛选器
- ✓ 优化预取功能
- ✓ 优化 LRU 缓存
- ✓ 编译 12 个平台版本
- ✓ 创建 V5 分支

## 联系方式

- **GitHub Issues：** https://github.com/lkxlzx/AdGuardHome/issues
- **Pull Requests：** https://github.com/lkxlzx/AdGuardHome/pulls

---

**最后更新：** 2025-11-27  
**维护分支：** V4, V5  
**推荐版本：** V4 (v0.107.0-dns-routing-fix)
