# DNS 上游分组功能 - 构建和测试指南

## 一、前置准备

### 1.1 环境要求
- Go 1.25.4 或更高版本
- Node.js v24.10.0 或更高版本
- npm v10.8 或更高版本

### 1.2 检查环境
```bash
go version
node --version
npm --version
```

## 二、添加翻译文本

在编译前，需要手动添加翻译文本到以下文件：

### 2.1 中文翻译
**文件**: `client/src/__locales/zh-cn.json`

在文件末尾（最后一个条目后）添加逗号，然后添加以下内容：

```json
    "upstream_groups_title": "上游 DNS 服务器分组",
    "upstream_groups_desc": "创建上游 DNS 服务器分组，可在 DNS 分流规则中使用这些分组。未命中分流规则的查询将使用默认组",
    "upstream_groups_empty": "暂无分组，点击下方按钮添加第一个分组（将自动设为默认组）",
    "upstream_group_name": "分组名称",
    "upstream_group_name_placeholder": "例如：国内 DNS、国外 DNS",
    "upstream_group_servers": "DNS 服务器列表",
    "upstream_group_servers_placeholder": "每行一个 DNS 服务器地址",
    "upstream_group_add": "添加分组",
    "upstream_group_confirm_delete": "确定要删除此分组吗？",
    "upstream_group_cannot_delete_default": "无法删除默认组，请先设置其他组为默认组",
    "default_group": "默认组",
    "default_group_hint": "分流规则未命中时使用此组",
    "set_as_default": "设为默认"
```

### 2.2 英文翻译
**文件**: `client/src/__locales/en.json`

同样在文件末尾添加：

```json
    "upstream_groups_title": "Upstream DNS Server Groups",
    "upstream_groups_desc": "Create upstream DNS server groups that can be used in DNS routing rules. Queries that don't match routing rules will use the default group",
    "upstream_groups_empty": "No groups yet. Click the button below to add your first group (it will be set as default automatically)",
    "upstream_group_name": "Group Name",
    "upstream_group_name_placeholder": "e.g., Domestic DNS, Foreign DNS",
    "upstream_group_servers": "DNS Server List",
    "upstream_group_servers_placeholder": "One DNS server address per line",
    "upstream_group_add": "Add Group",
    "upstream_group_confirm_delete": "Are you sure you want to delete this group?",
    "upstream_group_cannot_delete_default": "Cannot delete the default group. Please set another group as default first",
    "default_group": "Default",
    "default_group_hint": "Used when routing rules don't match",
    "set_as_default": "Set as Default"
```

**注意**: 如果这些键（如 save、cancel、edit、delete、add）已存在，则无需重复添加。

## 三、前端开发模式测试

### 3.1 安装依赖
```bash
cd client
npm install
```

### 3.2 启动开发服务器
```bash
npm run watch
```

这将启动 Webpack 开发服务器，支持热重载。

### 3.3 在浏览器中测试
1. 打开浏览器访问 `http://localhost:3000`（或开发服务器指定的端口）
2. 登录 AdGuard Home
3. 进入 设置 → DNS 设置 → 上游 DNS 服务器
4. 滚动到页面底部，查看"上游 DNS 服务器分组"部分

## 四、前端生产构建

### 4.1 构建前端
```bash
# 在项目根目录
make js-build
```

或者直接使用 npm：
```bash
cd client
npm run build-prod
```

### 4.2 检查构建产物
构建完成后，检查 `build/` 目录是否包含编译后的文件。

## 五、完整构建（前端 + 后端）

**注意**: 由于后端功能尚未实现，完整构建可能会失败。建议先完成后端实现。

### 5.1 初始化构建环境
```bash
make init
```

### 5.2 构建
```bash
make
```

这将：
1. 安装前端依赖
2. 构建前端
3. 编译 Go 后端
4. 生成可执行文件

### 5.3 运行
```bash
./AdGuardHome
```

## 六、前端功能测试步骤

### 6.1 基础功能测试

#### 测试 1: 添加第一个分组
1. 点击"添加分组"按钮
2. 输入分组名称：`国内 DNS`
3. 输入 DNS 服务器：
   ```
   223.5.5.5
   119.29.29.29
   ```
4. 点击"保存"
5. **验证**: 分组显示，且自动标记为"默认组"（黄色边框 + 徽章）

#### 测试 2: 添加第二个分组
1. 点击"添加分组"按钮
2. 输入分组名称：`国外 DNS`
3. 输入 DNS 服务器：
   ```
   8.8.8.8
   1.1.1.1
   ```
4. 点击"保存"
5. **验证**: 新分组显示，但不是默认组

#### 测试 3: 切换默认组
1. 在"国外 DNS"分组上点击"设为默认"按钮
2. **验证**: 
   - "国外 DNS"变为默认组（黄色边框 + 徽章）
   - "国内 DNS"不再是默认组（灰色边框，无徽章）

#### 测试 4: 编辑分组
1. 点击"国内 DNS"的"编辑"按钮
2. 修改名称为：`中国大陆 DNS`
3. 添加一个服务器：`114.114.114.114`
4. 点击"保存"
5. **验证**: 分组名称和服务器列表已更新

#### 测试 5: 尝试删除默认组
1. 点击默认组（"国外 DNS"）的"删除"按钮
2. **验证**: 按钮为禁用状态，无法点击

#### 测试 6: 删除非默认组
1. 点击"中国大陆 DNS"的"删除"按钮
2. 确认删除
3. **验证**: 分组被删除

#### 测试 7: 删除后只剩一个分组
1. 继续删除分组，直到只剩一个
2. **验证**: 最后一个分组自动成为默认组，删除按钮禁用

### 6.2 UI 显示测试

#### 测试 8: 空状态
1. 删除所有分组
2. **验证**: 显示空状态提示"暂无分组，点击下方按钮添加第一个分组（将自动设为默认组）"

#### 测试 9: 默认组样式
1. 添加一个分组（自动成为默认组）
2. **验证**:
   - 黄色边框（2px）
   - 黄色背景（#fff3cd）
   - 蓝色"默认组"徽章
   - 提示文字"分流规则未命中时使用此组"

#### 测试 10: 按钮状态
1. 点击"编辑"按钮进入编辑模式
2. **验证**: 其他分组的所有按钮变为禁用状态
3. 点击"取消"退出编辑模式
4. **验证**: 所有按钮恢复可用状态

### 6.3 数据持久化测试

#### 测试 11: 刷新页面
1. 添加几个分组，设置默认组
2. 刷新浏览器页面
3. **验证**: 所有分组和默认组标识正确恢复

**注意**: 此测试需要后端支持才能完全通过。

## 七、已知限制

### 7.1 当前版本限制
- ✅ 前端 UI 完全实现
- ✅ 前端逻辑完全实现
- ❌ 后端 API 尚未实现
- ❌ 数据持久化尚未实现
- ❌ 与 DNS 分流模块集成尚未实现

### 7.2 测试范围
当前可以测试：
- ✅ UI 显示和交互
- ✅ 前端状态管理
- ✅ 表单验证
- ✅ 按钮状态控制
- ✅ 默认组逻辑

当前无法测试：
- ❌ 数据保存到后端
- ❌ 页面刷新后数据恢复
- ❌ DNS 查询分流功能

## 八、下一步工作

### 8.1 后端实现优先级
1. **高优先级**: 
   - 修改 `internal/dnsforward/dnsforward.go` 添加 `UpstreamGroups` 字段
   - 修改 `internal/dnsforward/http.go` 支持分组数据的读写
   - 实现配置文件持久化

2. **中优先级**:
   - 实现 `getDefaultUpstreamGroup()` 函数
   - 修改 DNS 查询处理逻辑使用默认组

3. **低优先级**:
   - 完善错误处理
   - 添加日志记录
   - 性能优化

### 8.2 集成测试
后端实现完成后，需要进行：
1. 端到端测试
2. DNS 查询分流测试
3. 默认组回退测试
4. 性能测试

## 九、故障排查

### 9.1 前端构建失败
```bash
# 清理并重新安装依赖
cd client
rm -rf node_modules package-lock.json
npm install
npm run build-prod
```

### 9.2 翻译文本不显示
1. 检查 JSON 文件格式是否正确（逗号、引号）
2. 检查翻译键名是否拼写正确
3. 清除浏览器缓存并刷新

### 9.3 组件不显示
1. 打开浏览器开发者工具查看控制台错误
2. 检查 React 组件是否正确导入
3. 检查 Redux 状态是否正确初始化

## 十、联系和支持

如有问题，请查看：
- `DNS_UPSTREAM_GROUPS_IMPLEMENTATION.md` - 完整实现文档
- `DNS_UPSTREAM_GROUPS_TEST_CHECKLIST.md` - 详细测试清单
- `TRANSLATION_ADDITIONS.md` - 翻译文本参考
