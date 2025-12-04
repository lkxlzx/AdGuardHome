# 需求文档

## 简介

本功能为 AdGuard Home 的 DNS 设置添加上游服务器分组管理能力。用户可以创建多个独立的 DNS 上游分组，每个分组包含不同的 DNS 服务器列表，并可以设置默认分组。分组可以被过滤器规则调用，实现更灵活的 DNS 解析策略。

## 术语表

- **System**: AdGuard Home 系统
- **User**: 使用 AdGuard Home 的管理员用户
- **DNS Upstream Group**: DNS 上游服务器分组，包含一组 DNS 服务器地址
- **Default Group**: 默认分组，当没有特定规则匹配时使用的 DNS 上游分组
- **Group Name**: 分组名称，用于标识和引用分组的唯一标识符
- **Upstream Server List**: 上游服务器列表，包含一个或多个 DNS 服务器地址
- **Filter Rule**: 过滤器规则，可以指定使用特定的 DNS 上游分组
- **Group Status**: 分组状态，包括"已启用"和"设为默认分组"两种状态
- **DNS Query**: DNS 查询请求
- **UI Component**: 用户界面组件

## 需求

### 需求 1

**用户故事:** 作为管理员，我想要创建多个 DNS 上游分组，以便为不同的域名或客户端配置不同的 DNS 解析策略。

#### 验收标准

1. WHEN 用户点击"添加DNS上游分组"按钮 THEN THE System SHALL 显示分组创建对话框
2. WHEN 用户在创建对话框中输入分组名称和上游服务器列表并点击保存 THEN THE System SHALL 创建新的 DNS 上游分组并添加到分组列表中
3. WHEN 用户创建分组时未输入分组名称 THEN THE System SHALL 显示验证错误提示并阻止创建
4. WHEN 用户创建分组时未输入任何上游服务器 THEN THE System SHALL 显示验证错误提示并阻止创建
5. WHEN 新分组创建成功 THEN THE System SHALL 在分组列表中显示该分组，包括分组名称、状态、上游服务器地址和操作按钮

### 需求 2

**用户故事:** 作为管理员，我想要编辑现有的 DNS 上游分组，以便更新分组配置。

#### 验收标准

1. WHEN 用户点击分组列表中的编辑按钮 THEN THE System SHALL 显示分组编辑对话框，并预填充当前分组的配置信息
2. WHEN 用户修改分组名称或上游服务器列表并点击保存 THEN THE System SHALL 更新该分组的配置
3. WHEN 用户编辑分组时清空分组名称 THEN THE System SHALL 显示验证错误提示并阻止保存
4. WHEN 用户编辑分组时清空所有上游服务器 THEN THE System SHALL 显示验证错误提示并阻止保存
5. WHEN 分组更新成功 THEN THE System SHALL 在分组列表中显示更新后的信息

### 需求 3

**用户故事:** 作为管理员，我想要删除不需要的 DNS 上游分组，以便保持配置的整洁。

#### 验收标准

1. WHEN 用户点击分组列表中的删除按钮 THEN THE System SHALL 显示确认对话框
2. WHEN 用户在确认对话框中确认删除 THEN THE System SHALL 从分组列表中移除该分组
3. WHEN 用户尝试删除默认分组 THEN THE System SHALL 显示错误提示并阻止删除
4. WHEN 分组删除成功 THEN THE System SHALL 更新分组列表显示
5. WHEN 用户在确认对话框中取消删除 THEN THE System SHALL 关闭对话框并保持分组不变

### 需求 4

**用户故事:** 作为管理员，我想要设置某个分组为默认上游分组，以便在没有特定规则匹配时使用该分组的 DNS 服务器。

#### 验收标准

1. WHEN 用户勾选分组的"设为默认分组"复选框 THEN THE System SHALL 将该分组标记为默认分组
2. WHEN 用户设置新的默认分组 THEN THE System SHALL 取消之前默认分组的默认状态
3. WHEN 分组被设置为默认分组 THEN THE System SHALL 在分组列表中显示"默认"标识
4. THE System SHALL 确保始终有且仅有一个分组被标记为默认分组
5. WHEN DNS 查询没有匹配任何过滤器规则 THEN THE System SHALL 使用默认分组的上游服务器进行解析

### 需求 5

**用户故事:** 作为管理员，我想要启用或禁用 DNS 上游分组，以便临时停用某些分组而不删除它们。

#### 验收标准

1. WHEN 用户勾选分组的"已启用"复选框 THEN THE System SHALL 启用该分组
2. WHEN 用户取消勾选分组的"已启用"复选框 THEN THE System SHALL 禁用该分组
3. WHEN 分组被禁用 THEN THE System SHALL 不使用该分组的上游服务器进行 DNS 解析
4. WHEN 用户尝试禁用默认分组 THEN THE System SHALL 显示警告提示但允许禁用
5. WHEN 分组状态改变 THEN THE System SHALL 在分组列表中更新状态显示

### 需求 6

**用户故事:** 作为管理员，我想要测试 DNS 上游分组的连通性，以便验证配置的正确性。

#### 验收标准

1. WHEN 用户在分组编辑对话框中点击"检测"按钮 THEN THE System SHALL 测试该分组中所有上游服务器的连通性
2. WHEN 上游服务器测试开始 THEN THE System SHALL 显示加载状态指示器
3. WHEN 上游服务器测试完成 THEN THE System SHALL 显示测试结果，包括每个服务器的响应时间或错误信息
4. WHEN 所有上游服务器测试成功 THEN THE System SHALL 显示成功提示
5. WHEN 任何上游服务器测试失败 THEN THE System SHALL 显示失败的服务器和错误原因

### 需求 7

**用户故事:** 作为管理员，我想要在分组列表中查看所有 DNS 上游分组的概览信息，以便快速了解当前配置。

#### 验收标准

1. WHEN 用户访问 DNS 设置页面 THEN THE System SHALL 显示所有 DNS 上游分组的列表
2. WHEN 显示分组列表 THEN THE System SHALL 为每个分组显示已启用状态、分组名称、检测按钮、上游服务器地址和操作按钮
3. WHEN 分组被标记为默认分组 THEN THE System SHALL 在分组名称旁显示"默认"标识
4. WHEN 分组列表为空 THEN THE System SHALL 显示提示信息引导用户创建第一个分组
5. WHEN 分组列表包含多个分组 THEN THE System SHALL 支持分页显示，每页显示可配置数量的分组

### 需求 8

**用户故事:** 作为开发者，我想要为 DNS 上游分组提供 API 接口，以便过滤器规则可以调用特定的分组。

#### 验收标准

1. THE System SHALL 提供 API 接口用于获取所有 DNS 上游分组列表
2. THE System SHALL 提供 API 接口用于根据分组名称获取特定分组的配置
3. THE System SHALL 提供 API 接口用于创建新的 DNS 上游分组
4. THE System SHALL 提供 API 接口用于更新现有 DNS 上游分组的配置
5. THE System SHALL 提供 API 接口用于删除 DNS 上游分组
6. THE System SHALL 提供 API 接口用于设置默认 DNS 上游分组
7. THE System SHALL 提供 API 接口用于启用或禁用 DNS 上游分组
8. THE System SHALL 提供 API 接口用于测试 DNS 上游分组的连通性

### 需求 9

**用户故事:** 作为管理员，我想要 UI 界面完全遵循现有的 AdGuard Home 设计风格，以便保持用户体验的一致性。

#### 验收标准

1. WHEN 显示分组列表 THEN THE System SHALL 使用与现有 DNS 设置页面相同的表格样式和布局
2. WHEN 显示分组编辑对话框 THEN THE System SHALL 使用与现有对话框相同的样式、字体和颜色方案
3. WHEN 显示按钮和表单控件 THEN THE System SHALL 使用与现有 UI 组件相同的样式和交互效果
4. WHEN 显示状态标识（如"默认"标签） THEN THE System SHALL 使用与现有标签样式一致的绿色背景标签
5. WHEN 显示操作按钮（编辑、复制、删除） THEN THE System SHALL 使用与现有操作按钮相同的图标和布局

### 需求 10

**用户故事:** 作为管理员，我想要分组配置能够持久化保存，以便系统重启后配置不丢失。

#### 验收标准

1. WHEN 用户创建或修改 DNS 上游分组 THEN THE System SHALL 将配置保存到持久化存储
2. WHEN 系统启动 THEN THE System SHALL 从持久化存储加载所有 DNS 上游分组配置
3. WHEN 配置保存失败 THEN THE System SHALL 显示错误提示并保持当前配置不变
4. WHEN 配置加载失败 THEN THE System SHALL 使用默认配置并记录错误日志
5. THE System SHALL 确保配置文件的格式与现有 AdGuard Home 配置文件格式兼容
