# 分支保护设置指南

## 概述

本指南说明如何在GitHub上设置v5.1分支的保护规则，以防止意外的直接推送和确保代码质量。

## 为什么需要分支保护

- ✅ 防止意外的直接推送到主要分支
- ✅ 确保所有更改都经过代码审查
- ✅ 要求通过CI/CD测试后才能合并
- ✅ 保持代码库的稳定性
- ✅ 强制执行团队的开发流程

## 设置步骤

### 1. 访问仓库设置

1. 打开你的GitHub仓库：`https://github.com/lkxlzx/AdGuardHome`
2. 点击 **Settings**（设置）标签
3. 在左侧菜单中找到 **Branches**（分支）

### 2. 添加分支保护规则

1. 在 "Branch protection rules" 部分，点击 **Add rule**（添加规则）
2. 在 "Branch name pattern" 中输入：`v5.1`

### 3. 配置保护规则

#### 基础保护（推荐）

勾选以下选项：

**Require a pull request before merging**（合并前需要Pull Request）
- ✅ 勾选此项
- 设置 "Required approvals"（需要的审批数）：1（如果是个人项目可以设为0）
- ✅ Dismiss stale pull request approvals when new commits are pushed
  （新提交推送时取消过时的审批）

**Require status checks to pass before merging**（合并前需要通过状态检查）
- ✅ 勾选此项（如果有CI/CD配置）
- ✅ Require branches to be up to date before merging
  （合并前需要分支是最新的）

**Require conversation resolution before merging**（合并前需要解决所有对话）
- ✅ 勾选此项（确保所有评论都被处理）

**Require signed commits**（需要签名提交）
- ⚠️ 可选（如果团队使用GPG签名）

**Require linear history**（需要线性历史）
- ⚠️ 可选（防止合并提交）

**Include administrators**（包括管理员）
- ✅ 勾选此项（管理员也需要遵守规则）

**Restrict who can push to matching branches**（限制谁可以推送）
- ⚠️ 可选（如果是团队项目，可以指定特定用户）

**Allow force pushes**（允许强制推送）
- ❌ 不勾选（防止历史被覆盖）

**Allow deletions**（允许删除）
- ❌ 不勾选（防止分支被删除）

### 4. 保存规则

点击 **Create**（创建）或 **Save changes**（保存更改）

## 推荐配置

### 个人项目配置

```
✅ Require a pull request before merging
   - Required approvals: 0
✅ Require conversation resolution before merging
✅ Include administrators
❌ Allow force pushes
❌ Allow deletions
```

### 团队项目配置

```
✅ Require a pull request before merging
   - Required approvals: 1-2
   ✅ Dismiss stale pull request approvals when new commits are pushed
✅ Require status checks to pass before merging
   ✅ Require branches to be up to date before merging
✅ Require conversation resolution before merging
✅ Include administrators
✅ Restrict who can push to matching branches
❌ Allow force pushes
❌ Allow deletions
```

## 使用分支保护后的工作流程

### 1. 创建功能分支

```bash
# 从v5.1创建功能分支
git checkout v5.1
git pull origin v5.1
git checkout -b feature/your-feature-name
```

### 2. 开发和提交

```bash
# 进行开发
# ...

# 提交更改
git add .
git commit -m "feat: your feature description"
```

### 3. 推送功能分支

```bash
git push origin feature/your-feature-name
```

### 4. 创建Pull Request

1. 访问GitHub仓库
2. 点击 **Pull requests** 标签
3. 点击 **New pull request**
4. 选择：
   - base: `v5.1`
   - compare: `feature/your-feature-name`
5. 填写PR描述
6. 点击 **Create pull request**

### 5. 代码审查和合并

1. 等待代码审查（如果需要）
2. 解决所有评论
3. 确保所有检查通过
4. 点击 **Merge pull request**
5. 选择合并方式：
   - **Merge commit**（保留所有提交历史）
   - **Squash and merge**（压缩为单个提交）
   - **Rebase and merge**（变基合并）

### 6. 删除功能分支

```bash
# 本地删除
git branch -d feature/your-feature-name

# 远程删除
git push origin --delete feature/your-feature-name
```

## 绕过分支保护（紧急情况）

如果你是管理员且需要紧急修复：

### 方法1：临时禁用保护（不推荐）

1. 进入 Settings > Branches
2. 编辑v5.1的保护规则
3. 取消勾选 "Include administrators"
4. 进行紧急修复
5. 重新启用保护

### 方法2：使用Pull Request（推荐）

即使是紧急修复，也应该：
1. 创建hotfix分支
2. 快速修复
3. 创建PR并立即合并

```bash
git checkout v5.1
git checkout -b hotfix/critical-fix
# 进行修复
git add .
git commit -m "fix: critical issue"
git push origin hotfix/critical-fix
# 在GitHub上创建并立即合并PR
```

## 验证分支保护

### 测试1：尝试直接推送

```bash
git checkout v5.1
echo "test" >> test.txt
git add test.txt
git commit -m "test: direct push"
git push origin v5.1
```

**预期结果**：推送被拒绝
```
remote: error: GH006: Protected branch update failed for refs/heads/v5.1.
```

### 测试2：通过Pull Request

```bash
git checkout -b test/branch-protection
git push origin test/branch-protection
# 在GitHub上创建PR到v5.1
```

**预期结果**：可以创建PR并合并

## 常见问题

### Q: 我是管理员，为什么不能直接推送？

A: 如果勾选了 "Include administrators"，管理员也需要遵守分支保护规则。这是最佳实践。

### Q: 如何修改已保护的分支？

A: 通过Pull Request流程：
1. 创建功能分支
2. 进行修改
3. 推送功能分支
4. 创建PR到保护分支
5. 审查并合并

### Q: 可以强制推送到保护分支吗？

A: 不可以，除非在保护规则中启用了 "Allow force pushes"（不推荐）。

### Q: 如何删除保护分支？

A: 不可以直接删除，除非：
1. 在保护规则中启用 "Allow deletions"
2. 或者先删除保护规则，再删除分支

### Q: 分支保护会影响性能吗？

A: 不会。分支保护只是GitHub的策略检查，不影响Git操作性能。

## 最佳实践

1. **始终使用Pull Request**
   - 即使是小改动也通过PR
   - 保持代码审查的习惯

2. **保持分支同步**
   - 定期从保护分支拉取更新
   - 解决冲突后再创建PR

3. **编写清晰的PR描述**
   - 说明改动的目的
   - 列出主要变更
   - 添加测试说明

4. **及时响应审查意见**
   - 快速处理评论
   - 解释设计决策

5. **保持提交历史清晰**
   - 使用语义化提交信息
   - 考虑使用 Squash merge

## 相关文档

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- `V5.1_BRANCH_INFO.md` - v5.1分支信息
- `V5.2_BRANCH_INFO.md` - v5.2分支信息

## 总结

分支保护是维护代码质量和稳定性的重要工具。通过正确配置分支保护规则，可以：

- ✅ 防止意外的代码损坏
- ✅ 强制执行代码审查流程
- ✅ 确保所有更改都经过测试
- ✅ 保持清晰的开发历史

---

**更新时间**: 2025-11-28  
**适用分支**: v5.1  
**状态**: 待设置
