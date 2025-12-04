# DNS路由 UI测试版本说明

## 版本信息

- **版本**: v10.2-ui-test
- **分支**: v10.2
- **状态**: UI测试版（使用模拟数据）
- **创建日期**: 2025年12月4日

## 已完成的工作

### 1. 类型定义
- ✅ `client/src/types/dnsRouting.ts` - 完整的TypeScript类型定义

### 2. Redux状态管理
- ✅ `client/src/actions/dnsRouting.ts` - Actions（含模拟数据）
- ✅ `client/src/reducers/dnsRouting.ts` - Reducer
- ✅ 已集成到根reducer和initialState

### 3. UI组件
- ✅ `client/src/components/Filters/DnsRouting/index.tsx` - 主组件
- ✅ `client/src/components/Filters/DnsRouting/RuleListsTable.tsx` - 规则列表表格
- ✅ `client/src/components/Filters/DnsRouting/CustomRulesTable.tsx` - 自定义规则表格
- ✅ `client/src/components/Filters/DnsRouting/RuleListModal.tsx` - 规则列表对话框
- ✅ `client/src/components/Filters/DnsRouting/CustomRuleModal.tsx` - 自定义规则对话框
- ✅ `client/src/components/Filters/DnsRouting/DnsRouting.css` - 样式文件

### 4. 国际化
- ✅ `client/src/__locales/dns-routing-zh-cn.json` - 中文翻译
- ✅ `client/src/__locales/dns-routing-en.json` - 英文翻译

### 5. 文档
- ✅ `client/src/components/Filters/DnsRouting/README.md` - 组件文档
- ✅ `.kiro/specs/dns-routing/UI_DESIGN_PLAN.md` - UI设计规划
- ✅ `.kiro/specs/dns-routing/requirements.md` - 需求文档

## 如何启动测试

### 前置条件

确保你已经：
1. 切换到v10.2分支
2. 安装了所有依赖：`npm install`

### 启动步骤

#### 1. 合并翻译文件

由于翻译文件是独立的，需要手动合并或在i18n配置中加载：

**方法A：手动合并（推荐用于测试）**

```bash
# Windows PowerShell
cd client/src/__locales

# 合并中文翻译
$zhcn = Get-Content zh-cn.json | ConvertFrom-Json
$dnsrouting_zhcn = Get-Content dns-routing-zh-cn.json | ConvertFrom-Json
$merged = $zhcn.PSObject.Copy()
$dnsrouting_zhcn.PSObject.Properties | ForEach-Object { $merged | Add-Member -NotePropertyName $_.Name -NotePropertyValue $_.Value -Force }
$merged | ConvertTo-Json -Depth 10 | Set-Content zh-cn.json

# 合并英文翻译
$en = Get-Content en.json | ConvertFrom-Json
$dnsrouting_en = Get-Content dns-routing-en.json | ConvertFrom-Json
$merged_en = $en.PSObject.Copy()
$dnsrouting_en.PSObject.Properties | ForEach-Object { $merged_en | Add-Member -NotePropertyName $_.Name -NotePropertyValue $_.Value -Force }
$merged_en | ConvertTo-Json -Depth 10 | Set-Content en.json
```

**方法B：临时路由（快速测试）**

在 `client/src/components/App/index.tsx` 中添加临时路由：

```typescript
import DnsRouting from '../Filters/DnsRouting';

// 在路由配置中添加
<Route path="/dns-routing-test" component={DnsRouting} />
```

#### 2. 启动开发服务器

```bash
# 在项目根目录
npm start

# 或者如果有特定的客户端启动命令
cd client
npm start
```

#### 3. 访问测试页面

打开浏览器访问：
```
http://localhost:3000/dns-routing-test
```

或者根据你的路由配置访问相应的URL。

## 测试清单

### 基础功能测试

- [ ] **页面加载**
  - [ ] 页面标题显示正确
  - [ ] 副标题显示正确
  - [ ] 两个卡片都正常显示

- [ ] **规则列表表格**
  - [ ] 默认显示一条CN规则
  - [ ] 表格列显示正确（已启用、名称、URL、DNS分组、规则数、更新时间、操作）
  - [ ] 分页控件正常工作
  - [ ] 搜索框显示正常

- [ ] **规则列表操作**
  - [ ] 点击"添加规则"打开对话框
  - [ ] 填写表单并保存成功
  - [ ] 新规则出现在表格中
  - [ ] 点击编辑按钮打开对话框并显示现有数据
  - [ ] 修改数据并保存成功
  - [ ] 点击刷新按钮，规则数和更新时间变化
  - [ ] 点击删除按钮，弹出确认对话框
  - [ ] 确认删除后规则从表格中移除
  - [ ] 切换启用/禁用复选框，状态立即改变

- [ ] **自定义规则表格**
  - [ ] 初始状态显示"未添加自定义规则"
  - [ ] 表格列显示正确（已启用、域名、匹配类型、DNS分组、操作）

- [ ] **自定义规则操作**
  - [ ] 点击"自定义规则"打开对话框
  - [ ] 填写域名、选择匹配类型和上游组
  - [ ] 保存成功，新规则出现在表格中
  - [ ] 点击编辑按钮打开对话框并显示现有数据
  - [ ] 修改数据并保存成功
  - [ ] 点击删除按钮，弹出确认对话框
  - [ ] 确认删除后规则从表格中移除
  - [ ] 切换启用/禁用复选框，状态立即改变

### UI风格测试

- [ ] **按钮样式**
  - [ ] 图标按钮样式与上游分组页面一致
  - [ ] 主要操作按钮（绿色）样式正确
  - [ ] 次要操作按钮（蓝色）样式正确
  - [ ] 取消按钮（灰色）样式正确

- [ ] **表格样式**
  - [ ] 表格条纹样式正确
  - [ ] 鼠标悬停高亮效果正常
  - [ ] 复选框样式与其他页面一致
  - [ ] 文本溢出处理正确（显示省略号和title提示）

- [ ] **对话框样式**
  - [ ] 对话框宽度为600px
  - [ ] 表单字段间距正确
  - [ ] 说明文本颜色和字体大小正确
  - [ ] 按钮布局和间距正确

- [ ] **字体和间距**
  - [ ] 所有文本字体与应用一致
  - [ ] URL和域名使用等宽字体
  - [ ] 间距符合设计规范

### 国际化测试

- [ ] **中文界面**
  - [ ] 切换到中文，所有文本正确显示
  - [ ] 对话框标题和标签正确
  - [ ] 按钮文本正确
  - [ ] 提示信息正确

- [ ] **英文界面**
  - [ ] 切换到英文，所有文本正确显示
  - [ ] 对话框标题和标签正确
  - [ ] 按钮文本正确
  - [ ] 提示信息正确

### 交互测试

- [ ] **表单验证**
  - [ ] 必填字段为空时无法提交
  - [ ] 数字字段只接受数字输入
  - [ ] 下拉框必须选择选项

- [ ] **加载状态**
  - [ ] 操作进行中按钮显示禁用状态
  - [ ] 表格显示加载指示器

- [ ] **确认对话框**
  - [ ] 删除操作弹出确认对话框
  - [ ] 点击取消不执行删除
  - [ ] 点击确认执行删除

## 已知限制

1. **模拟数据**：所有数据都是前端模拟的，刷新页面会重置
2. **上游分组列表**：下拉框中的选项是硬编码的（"国内"、"国外"）
3. **搜索功能**：搜索框UI已实现，但搜索逻辑未实现
4. **检查更新**：按钮已实现，但功能未实现
5. **图标**：使用了`#refresh`图标，如果SVG sprite中不存在需要替换

## 反馈和问题

测试过程中如发现问题，请记录：

1. **问题描述**：详细描述问题
2. **重现步骤**：如何重现该问题
3. **预期行为**：应该是什么样的
4. **实际行为**：实际发生了什么
5. **截图**：如果可能，提供截图

## 下一步计划

UI验收通过后：

1. **后端API设计**
   - 设计RESTful API端点
   - 定义数据模型
   - 实现规则存储和管理

2. **前后端对接**
   - 替换模拟数据为真实API调用
   - 实现错误处理
   - 添加加载状态

3. **功能完善**
   - 实现搜索和筛选
   - 实现规则优先级排序
   - 实现自动更新
   - 实现规则测试

4. **测试**
   - 单元测试
   - 集成测试
   - E2E测试

5. **文档**
   - API文档
   - 用户手册
   - 开发者文档
