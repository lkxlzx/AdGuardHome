# V5分支保护设置指南

## 目的

将V5分支设置为只读保护分支，防止意外修改或删除。

## 设置步骤

### 1. 访问仓库设置

1. 打开GitHub仓库：https://github.com/lkxlzx/AdGuardHome
2. 点击顶部的 **Settings** (设置)
3. 在左侧菜单中找到 **Branches** (分支)

### 2. 添加分支保护规则

1. 在 "Branch protection rules" 部分，点击 **Add rule** (添加规则)
2. 在 "Branch name pattern" 中输入：`v5`

### 3. 配置保护规则

建议启用以下选项：

#### 基本保护
- ✅ **Require a pull request before merging** (合并前需要Pull Request)
  - ✅ Require approvals (需要审批)
  - 设置 Required number of approvals: 1

#### 防止强制推送和删除
- ✅ **Do not allow bypassing the above settings** (不允许绕过上述设置)
- ✅ **Do not allow force pushes** (不允许强制推送)
- ✅ **Do not allow deletions** (不允许删除)

#### 可选保护（根据需要）
- ⬜ Require status checks to pass before merging (合并前需要通过状态检查)
- ⬜ Require conversation resolution before merging (合并前需要解决讨论)
- ⬜ Require signed commits (需要签名提交)
- ⬜ Require linear history (需要线性历史)

### 4. 保存设置

点击页面底部的 **Create** (创建) 或 **Save changes** (保存更改)

## 推荐配置（只读保护）

如果要将V5设置为完全只读（不允许任何直接推送）：

```
Branch name pattern: v5

✅ Require a pull request before merging
   ✅ Require approvals (1)
   ✅ Dismiss stale pull request approvals when new commits are pushed
   
✅ Require status checks to pass before merging
   ✅ Require branches to be up to date before merging

✅ Require conversation resolution before merging

✅ Do not allow bypassing the above settings

✅ Restrict who can push to matching branches
   - 不添加任何用户（完全禁止直接推送）

✅ Do not allow force pushes

✅ Do not allow deletions
```

## 效果

设置完成后：

1. ✅ 无法直接推送到V5分支
2. ✅ 无法强制推送（force push）
3. ✅ 无法删除V5分支
4. ✅ 所有更改必须通过Pull Request
5. ✅ Pull Request需要审批才能合并

## 如何在保护分支上工作

如果需要更新V5分支：

1. 创建新分支：
   ```bash
   git checkout -b v5-update
   ```

2. 进行修改并提交：
   ```bash
   git add .
   git commit -m "更新说明"
   git push origin v5-update
   ```

3. 在GitHub上创建Pull Request：
   - 从 `v5-update` 到 `v5`
   - 等待审批
   - 合并到V5分支

## 紧急情况

如果需要临时解除保护：

1. 进入 Settings → Branches
2. 找到V5的保护规则
3. 点击 **Edit** (编辑)
4. 临时禁用某些规则
5. 完成操作后立即恢复保护

## 当前V5分支状态

- **最新提交**: 1e9e5e92
- **提交消息**: "新增上游DNS分组测试功能并编译主流平台可执行文件"
- **提交时间**: 2025-11-28
- **包含功能**:
  - 上游DNS分组测试功能
  - 6个主流平台可执行文件
  - Hot domains修复
  - Prefetch UI改进
  - 所有V4优化

## 注意事项

1. 分支保护规则只能由仓库管理员设置
2. 保护规则不影响本地仓库操作
3. 建议为重要的稳定版本分支都设置保护
4. 可以为不同分支设置不同的保护级别

## 相关文档

- [GitHub分支保护文档](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [V5分支信息](V5_BRANCH_INFO.md)
- [构建完成报告](UPSTREAM_TEST_BUILD_COMPLETE.md)

---

**创建时间**: 2025-11-28  
**仓库**: https://github.com/lkxlzx/AdGuardHome  
**保护分支**: v5
