# DNS 上游分组 UI 组件

## 概述

这个目录包含了 DNS 上游分组管理功能的前端 UI 组件。

## 组件结构

```
UpstreamGroups/
├── index.tsx           # 主容器组件
├── GroupList.tsx       # 分组列表组件（使用 react-table）
├── GroupModal.tsx      # 分组创建/编辑对话框
├── UpstreamGroups.css  # 样式文件
└── README.md           # 本文档
```

## 功能特性

### 已实现的功能

1. **分组列表展示**
   - 显示所有 DNS 上游分组
   - 显示分组的启用状态、名称、上游服务器地址
   - 显示默认分组标识
   - 空状态提示

2. **创建分组**
   - 点击"添加DNS上游分组"按钮打开对话框
   - 输入分组名称（必填，最多50字符）
   - 输入上游服务器列表（必填，每行一个）
   - 输入 Bootstrap DNS 服务器列表（可选）
   - 设置启用状态
   - 设置是否为默认分组

3. **编辑分组**
   - 点击编辑按钮打开对话框
   - 预填充当前分组配置
   - 修改分组信息
   - 测试分组连通性

4. **删除分组**
   - 点击删除按钮
   - 显示确认提示
   - 阻止删除默认分组

5. **启用/禁用分组**
   - 通过复选框切换分组启用状态
   - 实时更新状态

6. **测试分组连通性**
   - 在编辑对话框中点击"检测"按钮
   - 显示每个上游服务器的测试结果
   - 显示响应时间或错误信息

7. **复制分组**
   - 点击复制按钮
   - 创建分组副本（名称添加 "Copy" 后缀）

## 数据流

```
用户操作 → Redux Action → API 调用 → 后端处理
                ↓
         Redux Reducer 更新状态
                ↓
         组件重新渲染
```

## Redux 状态

```typescript
upstreamGroups: {
    groups: UpstreamGroup[];       // 分组列表
    processing: boolean;           // 是否正在加载
    processingAdd: boolean;        // 是否正在添加
    processingUpdate: boolean;     // 是否正在更新
    processingDelete: boolean;     // 是否正在删除
    processingTest: boolean;       // 是否正在测试
    isModalOpen: boolean;          // 对话框是否打开
    modalType: 'add' | 'edit';     // 对话框类型
    currentGroup?: UpstreamGroup;  // 当前编辑的分组
    testResult?: TestResult;       // 测试结果
}
```

## API 端点

- `GET /control/dns/upstream_groups` - 获取所有分组
- `POST /control/dns/upstream_groups` - 创建新分组
- `PUT /control/dns/upstream_groups/:id` - 更新分组
- `DELETE /control/dns/upstream_groups/:id` - 删除分组
- `POST /control/dns/upstream_groups/:id/default` - 设置默认分组
- `POST /control/dns/upstream_groups/:id/test` - 测试分组连通性

## 国际化

所有文本内容都支持中文和英文，翻译键位于：
- `client/src/__locales/zh-cn.json`
- `client/src/__locales/en.json`

## 样式

组件样式完全遵循 AdGuard Home 现有的设计规范：
- 使用 react-table 实现表格（与 Rewrites 等页面一致）
- 使用 ReactModal 实现对话框
- 使用 form__group、form__label 等原有表单样式类
- 使用 checkbox、checkbox__input、checkbox__label 等原有复选框样式
- 使用 btn-list、btn-standard 等原有按钮样式
- 绿色主题色 (#67b279)
- 响应式设计
- 表格和卡片布局

## 待完成

前端 UI 部分已经完成，但需要后端 API 支持才能正常工作。

后端需要实现：
1. 数据模型和配置文件结构
2. HTTP API 端点
3. 分组管理器核心逻辑
4. DNS 解析器集成
5. 配置持久化

## 测试

前端组件可以通过以下方式测试：
1. 启动开发服务器：`npm start`
2. 访问 DNS 设置页面
3. 查看上游分组部分

注意：在后端 API 实现之前，所有操作都会失败并显示错误提示。

## 下一步

1. 实现后端 API（参考 tasks.md 中的任务 1-5）
2. 测试前后端集成
3. 编写单元测试和 E2E 测试
4. 性能优化
5. 文档完善
